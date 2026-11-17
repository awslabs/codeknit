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
//  2. Copy maintained reference pages and generate English pages (incremental)
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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Page defines a documentation page to generate or copy from a maintained source.
type Page struct {
	Prompt string // prompt file name (without .txt); empty for maintained pages
	Source string // maintained source path relative to the repo root; copied verbatim
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
	{Source: "docs/reference/language-support.md", Output: "reference/language-support.md"},
	{Source: "docs/reference/analysis-reliability.md", Output: "reference/analysis-reliability.md"},
}

// syncSourcePages restores maintained pages on every run, including --force and
// --skip-generate. They participate in translation but never go through an LLM
// for English generation.
func syncSourcePages(contentDir string) error {
	for _, p := range pages {
		if p.Source == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Clean(p.Source))
		if err != nil {
			return fmt.Errorf("read maintained page %s: %w", p.Source, err)
		}
		if len(bytes.TrimSpace(content)) == 0 {
			return fmt.Errorf("maintained page %s is empty", p.Source)
		}
		if err := writeFile(filepath.Join(contentDir, p.Output), string(content)); err != nil {
			return fmt.Errorf("copy maintained page %s: %w", p.Output, err)
		}
	}
	return nil
}

func pagesToGenerate(contentDir string, force bool) []Page {
	var pending []Page
	for _, p := range pages {
		if p.Source != "" {
			continue
		}
		if !force && fileExists(filepath.Join(contentDir, p.Output)) {
			fmt.Printf("  [CACHED] %s\n", p.Output)
			continue
		}
		pending = append(pending, p)
	}
	return pending
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
	{Code: "th", Name: "Thai"},
	{Code: "id", Name: "Indonesian"},
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
	defaultGenModelID       = "global.anthropic.claude-sonnet-5-5"
	defaultTranslateModelID = "mistral.mistral-large-3-675b-instruct"
	defaultRegion           = "us-west-2"

	// translateAttempts and generateAttempts bound retries for output that fails validation.
	translateAttempts = 3
	generateAttempts  = 3
	// docsBase must match `base` in docs/astro.config.mjs.
	docsBase = "/codeknit"
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
		effort           = flag.String("effort", "", "Claude effort for English generation: low, medium, high, xhigh, max (empty = model default)")
		maxTokens        = flag.Int("max-tokens", 16384, "Maximum output tokens per request")
		concurrency      = flag.Int("concurrency", 5, "Max concurrent Bedrock requests")
		backtranslate    = flag.Bool("backtranslate", false, "Run backtranslation quality check (costs extra API calls)")
		btThreshold      = flag.Float64("bt-threshold", 0.5, "Backtranslation similarity threshold (0-1); below this triggers a warning")
	)
	flag.Parse()

	if err := syncSourcePages(*contentDir); err != nil {
		log.Fatalf("syncing maintained documentation: %v", err)
	}
	if *skipGenerate && *skipTranslate {
		fmt.Println("==> Maintained documentation synchronized")
		return
	}

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
		pending := pagesToGenerate(*contentDir, *force)
		if len(pending) == 0 {
			fmt.Println("==> Phase 1: All English pages cached (use --force to regenerate)")
		} else {
			fmt.Printf("==> Phase 1: Generating %d English pages with %s\n", len(pending), *genModelID)

			sem := make(chan struct{}, *concurrency)
			var wg sync.WaitGroup

			for _, p := range pending {
				outPath := filepath.Join(*contentDir, p.Output)
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

					frontmatter, ok := promptFrontmatter(string(promptBytes))
					if !ok {
						log.Printf("  [FAIL] %s: prompt has no ---frontmatter--- block", p.Prompt)
						return
					}
					fullPrompt := sourceContext + "\n\n---\n\n" + sitemapFor(p.Output) + "\n" + string(promptBytes)

					fmt.Printf("  [GEN] %s ...\n", p.Output)
					// Pages with links to non-existent pages are retried; a page that never passes is not written.
					var content string
					var problems []string
					for attempt := 1; attempt <= generateAttempts; attempt++ {
						var err error
						content, err = converse(ctx, client, *genModelID, string(systemPrompt), fullPrompt, *maxTokens, nil, *effort)
						if err != nil {
							problems = []string{err.Error()}
						} else {
							content = withFrontmatter(content, frontmatter)
							problems = nil
							if bad := brokenLinks(p.Output, content); len(bad) > 0 {
								problems = []string{fmt.Sprintf("links to unknown pages: %v", bad)}
							}
						}
						if len(problems) == 0 {
							break
						}
						log.Printf("  [RETRY %d/%d] %s: %s", attempt, generateAttempts, p.Output, strings.Join(problems, "; "))
					}
					if len(problems) > 0 {
						log.Printf("  [FAIL] %s: still invalid after %d attempts, existing file left unchanged", p.Output, generateAttempts)
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
						prompt := fmt.Sprintf("%s\n\nTARGET LANGUAGE: %s\nTARGET LOCALE: %s\n\nTranslate the following documentation:\n\n%s",
							glossaryBlock, l.Name, l.Code, string(enContent))

						fmt.Printf("  [TRANSLATE] %s -> %s ...\n", p.Output, l.Code)
						// Translations that fail structural validation are retried; a page that never
						// passes keeps its previous file and stays out of the manifest, so the next run retries it.
						var content string
						var problems []string
						for attempt := 1; attempt <= translateAttempts; attempt++ {
							var err error
							// Temperature 0 repeats the same mistake, so retries sample a little.
							temperature := float32(0)
							if attempt > 1 {
								temperature = 0.3
							}
							content, err = converse(ctx, client, *translateModelID, string(translateSystem), prompt, *maxTokens, aws.Float32(temperature), "")
							if err != nil {
								problems = []string{err.Error()}
							} else {
								problems = validateTranslation(string(enContent), content, l.Code)
							}
							if len(problems) == 0 {
								break
							}
							log.Printf("  [RETRY %d/%d] %s/%s: %s", attempt, translateAttempts, l.Code, p.Output, strings.Join(problems, "; "))
						}
						if len(problems) > 0 {
							rejected := filepath.Join(os.TempDir(), "codeknit-rejected-"+strings.ReplaceAll(l.Code+"/"+p.Output, "/", "_"))
							_ = os.WriteFile(rejected, []byte(content), 0o600)
							log.Printf("  [FAIL] %s/%s: still invalid after %d attempts, existing file left unchanged (last output: %s)", l.Code, p.Output, translateAttempts, rejected)
							return
						}

						content = localizeLinks(content, l.Code)
						if err := writeFile(outPath, content); err != nil {
							log.Printf("  [FAIL] write %s: %v", outPath, err)
							return
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
func converse(ctx context.Context, client *bedrockruntime.Client, modelID, system, prompt string, maxTokens int, temperature *float32, effort string) (string, error) {
	// temperature nil omits the field: newer Claude models reject it as deprecated.
	// effort is Claude-only (output_config.effort); leave it empty for other providers.
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
			Temperature: temperature,
		},
	}

	if effort != "" {
		input.AdditionalModelRequestFields = document.NewLazyDocument(map[string]any{
			"output_config": map[string]any{"effort": effort},
		})
	}

	resp, err := client.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("bedrock converse: %w", err)
	}

	if resp.StopReason == types.StopReasonMaxTokens {
		return "", fmt.Errorf("response truncated at max tokens (%d)", maxTokens)
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
	"docs/reference/language-support.md",
	"docs/reference/analysis-reliability.md",
	"skills/codeknit/SKILL.md",
	"skills/codeknit/references/output-format.md",
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
func validateTranslation(source, translated, langCode string) []string {
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

	// Check that link targets are preserved exactly (translators must not add, drop, or rewrite URLs).
	srcLinks := comparableLinks(source, "")
	transLinks := comparableLinks(translated, langCode)
	if !slices.Equal(srcLinks, transLinks) {
		warnings = append(warnings, fmt.Sprintf("link targets differ: source %v, translation %v", srcLinks, transLinks))
	}

	// Check heading count matches.
	srcHeadings := countHeadings(source)
	transHeadings := countHeadings(translated)
	if srcHeadings != transHeadings {
		warnings = append(warnings, fmt.Sprintf("heading count mismatch: source has %d, translation has %d", srcHeadings, transHeadings))
	}

	return warnings
}

// linkTargetRes match markdown links, MDX/HTML href attributes, and frontmatter hero `link:` values.
var linkTargetRes = []*regexp.Regexp{
	regexp.MustCompile(`\]\(([^)\s]*)`),
	regexp.MustCompile(`href="([^"]*)"`),
	regexp.MustCompile(`(?m)^\s*link:\s*(\S+)`),
}

// linkTargets returns the link targets in text.
func linkTargets(text string) []string {
	var targets []string
	for _, re := range linkTargetRes {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			targets = append(targets, m[1])
		}
	}
	return targets
}

// localizeLinks adds the locale segment to absolute site links. Done in code because
// asking the translator to do it made it rewrite relative links too.
func localizeLinks(text, langCode string) string {
	for _, prefix := range []string{"](", `href="`, "link: "} {
		text = strings.ReplaceAll(text, prefix+docsBase+"/", prefix+docsBase+"/"+langCode+"/")
	}
	return text
}

// comparableLinks returns sorted link targets with the differences a correct
// translation may introduce removed: anchors follow translated headings (so only
// their presence is compared), and absolute /codeknit/ links gain the locale segment.
func comparableLinks(text, langCode string) []string {
	var out []string
	for _, t := range linkTargets(text) {
		if i := strings.IndexByte(t, '#'); i >= 0 {
			t = t[:i+1]
		}
		if langCode != "" {
			if rest, ok := strings.CutPrefix(t, docsBase+"/"+langCode+"/"); ok {
				t = docsBase + "/" + rest
			}
		}
		out = append(out, t)
	}
	slices.Sort(out)
	return out
}

// promptFrontmatter returns the first ---...--- block of a page prompt, delimiters included.
func promptFrontmatter(prompt string) (string, bool) {
	lines := strings.Split(prompt, "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) != "---" {
			continue
		}
		if start < 0 {
			start = i
			continue
		}
		return strings.Join(lines[start:i+1], "\n"), true
	}
	return "", false
}

// withFrontmatter replaces any frontmatter the model wrote with the prompt's exact block,
// so pages always build and titles never drift (models regularly omit or alter it).
func withFrontmatter(content, frontmatter string) string {
	body := strings.TrimLeft(content, "\n")
	if rest, ok := strings.CutPrefix(body, "---\n"); ok {
		if end := strings.Index(rest, "\n---"); end >= 0 {
			body = rest[end+len("\n---"):]
		}
	}
	return frontmatter + "\n\n" + strings.TrimLeft(body, "\n")
}

// pageURL is the site URL of a page given its output path under content/docs.
func pageURL(output string) string {
	slug := strings.TrimSuffix(strings.TrimSuffix(output, ".mdx"), ".md")
	if slug == "index" {
		return docsBase + "/"
	}
	return docsBase + "/" + slug + "/"
}

// sitemapFor tells the model the exact relative URL of every other page, as seen from output.
func sitemapFor(output string) string {
	var sb strings.Builder
	sb.WriteString("VALID LINKS: link to other documentation pages only with these exact relative URLs (add #anchor when needed). Do not invent other page URLs and do not use absolute site paths.\n")
	for _, p := range pages {
		if p.Output == output {
			continue
		}
		rel, err := filepath.Rel(pageURL(output), pageURL(p.Output))
		if err != nil {
			continue
		}
		fmt.Fprintf(&sb, "- %s: %s/\n", p.Output, filepath.ToSlash(rel))
	}
	return sb.String()
}

// brokenLinks returns internal link targets in content that do not resolve to a known page.
func brokenLinks(output, content string) []string {
	valid := map[string]bool{}
	for _, p := range pages {
		valid[pageURL(p.Output)] = true
	}
	base := &url.URL{Path: pageURL(output)}
	var bad []string
	for _, t := range linkTargets(content) {
		if t == "" || strings.HasPrefix(t, "#") || strings.Contains(t, "://") || strings.HasPrefix(t, "mailto:") {
			continue
		}
		u, err := url.Parse(t)
		if err != nil {
			bad = append(bad, t)
			continue
		}
		resolved := base.ResolveReference(u).Path
		if !strings.HasSuffix(resolved, "/") {
			resolved += "/"
		}
		if !valid[resolved] {
			bad = append(bad, t)
		}
	}
	return bad
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
	backTranslated, err := converse(ctx, client, modelID, "", prompt, maxTokens, aws.Float32(0), "")
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
	// The local --content flag intentionally selects the output directory;
	// page filenames and locale directories come from the fixed lists above.
	return os.WriteFile(filepath.Clean(path), []byte(content), 0o600) //#nosec G703 -- user-selected local output directory
}
