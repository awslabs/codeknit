// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Command docgen generates codeknit documentation pages using Amazon Bedrock.
// It reads prompt templates from docs/prompts/, generates English content, then
// translates into multiple languages — all via the Bedrock Converse API.
// Generation and translation use separate models configurable via --model and
// --translate-model flags.
//
// Pipeline:
//  1. Run codeknit parse to extract codebase structure
//  2. Generate all English pages (incremental — skips existing unless --force)
//  3. Translate all pages into target languages (incremental)
//
// No local GPU required. Uses AWS credentials from the default chain.
//
// This tool is internal to the docs pipeline and is NOT part of the codeknit binary.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Page defines a documentation page to generate.
type Page struct {
	Prompt string // prompt file name (without .txt)
	Output string // output path relative to content/docs/ (root locale)
}

var pages = []Page{
	{Prompt: "index", Output: "index.mdx"},
	{Prompt: "getting-started", Output: "getting-started.md"},
	{Prompt: "installation", Output: "installation.md"},
	{Prompt: "parse-command", Output: "guides/parse-command.md"},
	{Prompt: "graph-commands", Output: "guides/graph-commands.md"},
	{Prompt: "fingerprint-command", Output: "guides/fingerprint-command.md"},
	{Prompt: "output-modes", Output: "guides/output-modes.md"},
	{Prompt: "ai-assistants", Output: "guides/ai-assistants.md"},
	{Prompt: "output-format", Output: "reference/output-format.md"},
	{Prompt: "cli-flags", Output: "reference/cli-flags.md"},
}

type lang struct {
	Code string
	Name string
}

var languages = []lang{
	{Code: "es", Name: "Spanish"},
	{Code: "de", Name: "German"},
	{Code: "zh-cn", Name: "Simplified Chinese"},
	{Code: "ja", Name: "Japanese"},
	{Code: "ko", Name: "Korean"},
	{Code: "fr", Name: "French"},
	{Code: "vi", Name: "Vietnamese"},
	{Code: "it", Name: "Italian"},
}

// Glossary holds do-not-translate terms and fixed translations per language.
type Glossary struct {
	FixedTranslations map[string]map[string]string `json:"fixed_translations"`
	DoNotTranslate    []string                     `json:"do_not_translate"`
}

// formatGlossaryForLang formats the glossary as a text block for injection into
// the translation prompt for a specific target language.
func formatGlossaryForLang(g *Glossary, langCode string) string {
	var sb strings.Builder

	sb.WriteString("GLOSSARY:\n\n")
	sb.WriteString("DO-NOT-TRANSLATE (keep these terms verbatim in the output):\n")
	for _, term := range g.DoNotTranslate {
		fmt.Fprintf(&sb, "- %s\n", term)
	}

	sb.WriteString("\nFIXED TRANSLATIONS (use these exact translations):\n")
	for term, translations := range g.FixedTranslations {
		if trans, ok := translations[langCode]; ok {
			fmt.Fprintf(&sb, "- \"%s\" → \"%s\"\n", term, trans)
		}
	}

	return sb.String()
}

const (
	defaultGenModelID       = "us.anthropic.claude-sonnet-4-6-20250514-v1:0"
	defaultTranslateModelID = "mistral.mistral-large-3-675b-instruct"
	defaultRegion           = "us-west-2"
	// codeknitBin is hardcoded to avoid passing a user-controlled path to
	// exec.Command. The doc generator is always run from the repo root.
	codeknitBin = "./bin/codeknit"
)

func main() {
	var (
		genModelID       = flag.String("model", defaultGenModelID, "Bedrock model ID for English generation (cross-region inference profile)")
		translateModelID = flag.String("translate-model", defaultTranslateModelID, "Bedrock model ID for translation")
		region           = flag.String("region", defaultRegion, "AWS region (defaults to AWS_REGION / config)")
		promptDir        = flag.String("prompts", "docs/prompts", "Directory containing prompt templates")
		contentDir       = flag.String("content", "docs/src/content/docs", "Starlight content directory")
		force            = flag.Bool("force", false, "Regenerate/retranslate all files even if they already exist")
		skipTranslate    = flag.Bool("skip-translate", false, "Skip translation step")
		skipGenerate     = flag.Bool("skip-generate", false, "Skip English generation (translate existing)")
		maxTokens        = flag.Int("max-tokens", 8192, "Maximum output tokens per request")
		concurrency      = flag.Int("concurrency", 5, "Max concurrent Bedrock requests")
		backtranslate    = flag.Bool("backtranslate", false, "Run backtranslation quality check (costs extra API calls)")
		btThreshold      = flag.Float64("bt-threshold", 0.5, "Backtranslation similarity threshold (0-1); below this triggers a warning")
	)
	flag.Parse()

	ctx := context.Background()

	// Load AWS config.
	var cfgOpts []func(*awsconfig.LoadOptions) error
	if *region != "" {
		cfgOpts = append(cfgOpts, awsconfig.WithRegion(*region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	client := bedrockruntime.NewFromConfig(awsCfg)

	systemPrompt, err := os.ReadFile(filepath.Join(*promptDir, "system.txt"))
	if err != nil {
		log.Fatalf("reading system prompt: %v", err)
	}

	translateSystem, err := os.ReadFile(filepath.Join(*promptDir, "translate-system.txt"))
	if err != nil {
		log.Fatalf("reading translate system prompt: %v", err)
	}

	// Load glossary for translation consistency.
	var glossary Glossary
	glossaryPath := filepath.Join(*promptDir, "glossary.json")
	glossaryData, err := os.ReadFile(filepath.Clean(glossaryPath))
	switch {
	case err != nil:
		log.Printf("  [WARN] no glossary found at %s: %v", glossaryPath, err)
	default:
		if err := json.Unmarshal(glossaryData, &glossary); err != nil {
			log.Fatalf("failed to parse glossary: %v", err)
		}
	}

	// Run codeknit on the project and load key docs as context for generation.
	sourceContext := loadSourceContext()

	// --- Phase 1: Generate English pages ---
	if !*skipGenerate {
		var needsGen int
		for _, p := range pages {
			if *force || !fileExists(filepath.Join(*contentDir, p.Output)) {
				needsGen++
			}
		}

		if needsGen == 0 {
			fmt.Println("==> Phase 1: All English pages cached (use --force to regenerate)")
		} else {
			fmt.Printf("==> Phase 1: Generating %d English pages with %s\n", needsGen, *genModelID)

			sem := make(chan struct{}, *concurrency)
			var wg sync.WaitGroup

			for _, p := range pages {
				outPath := filepath.Join(*contentDir, p.Output)
				if !*force && fileExists(outPath) {
					fmt.Printf("  [CACHED] %s\n", p.Output)
					continue
				}

				wg.Add(1)
				go func(p Page, outPath string) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					promptBytes, err := os.ReadFile(filepath.Join(*promptDir, p.Prompt+".txt"))
					if err != nil {
						log.Printf("  [SKIP] %s: %v", p.Prompt, err)
						return
					}

					fullPrompt := sourceContext + "\n\n---\n\n" + string(promptBytes)

					fmt.Printf("  [GEN] %s ...\n", p.Output)
					content, err := converse(ctx, client, *genModelID, string(systemPrompt), fullPrompt, *maxTokens, 0)
					if err != nil {
						log.Printf("  [FAIL] %s: %v", p.Prompt, err)
						return
					}

					if err := writeFile(outPath, content); err != nil {
						log.Printf("  [FAIL] write %s: %v", outPath, err)
						return
					}
					fmt.Printf("  [OK]  %s (%d bytes)\n", p.Output, len(content))
				}(p, outPath)
			}
			wg.Wait()
			fmt.Println("==> Phase 1 complete")
		}
	}

	// --- Phase 2: Translate into target languages ---
	if !*skipTranslate {
		manifestPath := filepath.Join(filepath.Dir(*contentDir), ".translation-hashes.json")
		manifest := loadManifest(manifestPath)

		var needsTrans int
		for _, p := range pages {
			enPath := filepath.Join(*contentDir, p.Output)
			currentHash := hashFile(enPath)
			for _, l := range languages {
				outPath := filepath.Join(*contentDir, l.Code, p.Output)
				key := l.Code + "/" + p.Output
				if *force || !fileExists(outPath) || manifest[key] != currentHash {
					needsTrans++
				}
			}
		}

		if needsTrans == 0 {
			fmt.Println("==> Phase 2: All translations up to date (use --force to retranslate)")
		} else {
			fmt.Printf("==> Phase 2: Translating %d pages with %s\n", needsTrans, *translateModelID)

			sem := make(chan struct{}, *concurrency)
			var wg sync.WaitGroup
			var mu sync.Mutex

			for _, p := range pages {
				enPath := filepath.Join(*contentDir, p.Output)
				enContent, err := os.ReadFile(filepath.Clean(enPath))
				if err != nil {
					log.Printf("  [SKIP] %s: English source not found: %v", p.Output, err)
					continue
				}
				currentHash := hashFile(enPath)

				for _, l := range languages {
					outPath := filepath.Join(*contentDir, l.Code, p.Output)
					key := l.Code + "/" + p.Output

					if !*force && fileExists(outPath) && manifest[key] == currentHash {
						fmt.Printf("  [UP-TO-DATE] %s/%s\n", l.Code, p.Output)
						continue
					}

					wg.Add(1)
					go func(p Page, l lang, enContent []byte, outPath, key, hash string) {
						defer wg.Done()
						sem <- struct{}{}
						defer func() { <-sem }()

						glossaryBlock := formatGlossaryForLang(&glossary, l.Code)
						prompt := fmt.Sprintf("%s\n\nTARGET LANGUAGE: %s\n\nTranslate the following documentation:\n\n%s",
							glossaryBlock, l.Name, string(enContent))

						fmt.Printf("  [TRANSLATE] %s -> %s ...\n", p.Output, l.Code)
						content, err := converse(ctx, client, *translateModelID, string(translateSystem), prompt, *maxTokens, 0)
						if err != nil {
							log.Printf("  [FAIL] %s/%s: %v", l.Code, p.Output, err)
							return
						}

						if err := writeFile(outPath, content); err != nil {
							log.Printf("  [FAIL] write %s: %v", outPath, err)
							return
						}

						// Post-translation validation.
						if warns := validateTranslation(string(enContent), content); len(warns) > 0 {
							for _, w := range warns {
								log.Printf("  [WARN] %s/%s: %s", l.Code, p.Output, w)
							}
						}

						// Optional backtranslation quality check.
						if *backtranslate {
							score, btErr := backtranslateCheck(ctx, client, *translateModelID, string(enContent), content, *maxTokens)
							switch {
							case btErr != nil:
								log.Printf("  [WARN] backtranslation failed for %s/%s: %v", l.Code, p.Output, btErr)
							case score < *btThreshold:
								log.Printf("  [WARN] %s/%s: low backtranslation similarity %.2f (threshold %.2f)", l.Code, p.Output, score, *btThreshold)
							default:
								fmt.Printf("  [QA]  %s/%s backtranslation score: %.2f\n", l.Code, p.Output, score)
							}
						}

						mu.Lock()
						manifest[key] = hash
						mu.Unlock()

						fmt.Printf("  [OK]  %s/%s (%d bytes)\n", l.Code, p.Output, len(content))
					}(p, l, enContent, outPath, key, currentHash)
				}
			}
			wg.Wait()

			if err := saveManifest(manifestPath, manifest); err != nil {
				log.Printf("  [WARN] failed to save translation manifest: %v", err)
			}
			fmt.Println("==> Phase 2 complete")
		}
	}

	fmt.Println("==> Done")
}

// converse calls the Bedrock Converse API with the given model, system prompt,
// and user message. Returns the text content of the assistant's response.
func converse(ctx context.Context, client *bedrockruntime.Client, modelID, system, prompt string, maxTokens int, temperature float32) (string, error) {
	input := &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelID),
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: system},
		},
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: prompt},
				},
			},
		},
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens:   aws.Int32(int32(min(maxTokens, math.MaxInt32))), //nolint:gosec // maxTokens is user-controlled CLI flag, clamped to safe range
			Temperature: aws.Float32(temperature),
		},
	}

	resp, err := client.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("bedrock converse: %w", err)
	}

	// Extract text from the response.
	msgOutput, ok := resp.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected response output type: %T", resp.Output)
	}

	var sb strings.Builder
	for _, block := range msgOutput.Value.Content {
		if textBlock, ok := block.(*types.ContentBlockMemberText); ok {
			sb.WriteString(textBlock.Value)
		}
	}

	content := strings.TrimSpace(sb.String())
	content = stripCodeFences(content)

	return content, nil
}

// sourceContextFiles are key project documentation files included as-is in the
// generation context alongside the codeknit structural output.
var sourceContextFiles = []string{
	"README.md",
	"skills/codeknit-parse/SKILL.md",
	"skills/codeknit-parse/OUTPUT-FORMAT.md",
}

// loadSourceContext runs "codeknit parse" on the project root to get a compact
// structural representation of the entire codebase, then appends key markdown
// files verbatim. This gives the LLM full visibility into the project.
func loadSourceContext() string {
	var sb strings.Builder

	sb.WriteString("You are writing documentation for the codeknit project. " +
		"Below is the complete codebase context: a structural skeleton from codeknit itself, " +
		"plus key documentation files. Use these as the authoritative source of truth for all " +
		"facts, flags, commands, algorithms, and examples. Do not invent features or flags " +
		"that are not present in this context.\n\n")

	sktOutput, err := runcodeknit()
	if err != nil {
		log.Printf("  [WARN] codeknit parse failed: %v — falling back to docs files only", err)
	} else {
		sb.WriteString("=== CODEBASE STRUCTURE (codeknit parse output) ===\n")
		sb.WriteString(sktOutput)
		sb.WriteString("\n\n")
	}

	for _, path := range sourceContextFiles {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			log.Printf("  [WARN] could not read %s: %v", path, err)
			continue
		}
		fmt.Fprintf(&sb, "=== FILE: %s ===\n%s\n\n", path, string(content))
	}

	return sb.String()
}

// runcodeknit executes "codeknit parse . --output-mode inline" to get a compact
// structural overview of the entire codebase on stdout.
func runcodeknit() (string, error) {
	cmd := exec.Command(codeknitBin, "parse", ".", "--output-mode", "inline")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println("==> Running codeknit parse to extract codebase structure...")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	fmt.Printf("    Extracted %d bytes of structural context\n", stdout.Len())

	return stdout.String(), nil
}

// stripCodeFences removes wrapping ```markdown ... ``` or ``` ... ``` fences
// that LLMs sometimes add around their output. It trims leading/trailing blank
// lines first, then checks for nested fences (e.g. ```yaml wrapping frontmatter).
func stripCodeFences(s string) string {
	lines := strings.Split(s, "\n")

	// Trim leading blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Trim trailing blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) < 2 {
		return strings.Join(lines, "\n")
	}

	first := strings.TrimSpace(lines[0])
	last := strings.TrimSpace(lines[len(lines)-1])

	// Strip outer wrapping fence (```markdown, ```md, ```yaml, etc.).
	if strings.HasPrefix(first, "```") && last == "```" {
		lines = lines[1 : len(lines)-1]
	}

	// Strip ```yaml / ```yml fence around frontmatter.
	// Pattern: ```yaml\n---\n...\n---\n```
	if len(lines) >= 4 {
		tag := strings.TrimSpace(lines[0])
		if tag == "```yaml" || tag == "```yml" {
			// Find the closing ``` that ends the fenced frontmatter.
			for i := 1; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "```" {
					// Verify there's a --- inside (valid frontmatter).
					if strings.TrimSpace(lines[1]) == "---" {
						lines = append(lines[1:i], lines[i+1:]...)
						break
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// fileExists returns true if the path exists and is a non-empty file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Size() > 0
}

// hashFile returns the hex-encoded SHA-256 hash of a file's contents.
// Returns an empty string if the file cannot be read.
func hashFile(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// loadManifest reads the translation hash manifest from disk.
// Returns an empty map if the file doesn't exist or can't be parsed.
func loadManifest(path string) map[string]string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return make(map[string]string)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]string)
	}
	return m
}

// saveManifest writes the translation hash manifest to disk.
func saveManifest(path string, m map[string]string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600) //#nosec G306 — manifest is non-sensitive project metadata
}

// validateTranslation checks structural integrity of a translated markdown file
// against its English source. Returns a list of warnings (empty = all good).
func validateTranslation(source, translated string) []string {
	var warnings []string

	// Check frontmatter is present.
	if !strings.HasPrefix(strings.TrimSpace(translated), "---") {
		warnings = append(warnings, "missing frontmatter (should start with ---)")
	}

	// Check code fence balance.
	srcFences := strings.Count(source, "```")
	transFences := strings.Count(translated, "```")
	if srcFences != transFences {
		warnings = append(warnings, fmt.Sprintf("code fence mismatch: source has %d, translation has %d", srcFences, transFences))
	}

	// Check code fence language tags are preserved.
	for _, tag := range []string{"```skt", "```bash", "```fish", "```powershell"} {
		srcCount := strings.Count(source, tag)
		transCount := strings.Count(translated, tag)
		if srcCount != transCount {
			warnings = append(warnings, fmt.Sprintf("code fence tag %q: source has %d, translation has %d", tag, srcCount, transCount))
		}
	}

	// Check that links are preserved (count markdown links).
	srcLinks := strings.Count(source, "](")
	transLinks := strings.Count(translated, "](")
	if srcLinks != transLinks {
		warnings = append(warnings, fmt.Sprintf("link count mismatch: source has %d, translation has %d", srcLinks, transLinks))
	}

	// Check heading count matches.
	srcHeadings := countHeadings(source)
	transHeadings := countHeadings(translated)
	if srcHeadings != transHeadings {
		warnings = append(warnings, fmt.Sprintf("heading count mismatch: source has %d, translation has %d", srcHeadings, transHeadings))
	}

	return warnings
}

// countHeadings counts markdown headings (## and above) in text.
func countHeadings(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") ||
			strings.HasPrefix(trimmed, "#### ") {
			count++
		}
	}
	return count
}

// backtranslateCheck translates the content back to English and compares similarity
// with the original. Returns a score between 0 and 1.
func backtranslateCheck(ctx context.Context, client *bedrockruntime.Client, modelID string,
	original, translated string, maxTokens int,
) (float64, error) {
	prompt := "Translate the following text back to English. Output ONLY the English translation, nothing else:\n\n" + translated
	backTranslated, err := converse(ctx, client, modelID, "", prompt, maxTokens, 0)
	if err != nil {
		return 0, fmt.Errorf("backtranslation: %w", err)
	}

	return jaccardSimilarity(original, backTranslated), nil
}

// jaccardSimilarity computes word-level Jaccard similarity between two texts.
func jaccardSimilarity(a, b string) float64 {
	wordsA := wordSet(strings.ToLower(a))
	wordsB := wordSet(strings.ToLower(b))

	intersection := 0
	for w := range wordsA {
		if wordsB[w] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// wordSet splits text into a set of unique words.
func wordSet(text string) map[string]bool {
	set := make(map[string]bool)
	for _, w := range strings.Fields(text) {
		w = strings.Trim(w, ".,;:!?\"'()[]{}#*`~")
		if w != "" {
			set[w] = true
		}
	}
	return set
}

// writeFile creates parent directories and writes content to the given path.
func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return os.WriteFile(path, []byte(content), 0o600)
}
