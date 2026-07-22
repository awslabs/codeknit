// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/config"
	"codeknit/internal/fingerprint"
	"codeknit/internal/ir"
	"codeknit/internal/ollama"
	"codeknit/internal/plugin"
)

// FingerprintOptions controls fingerprint emission behavior.
type FingerprintOptions struct {
	// OutputPath is the file to write the fingerprint results to.
	OutputPath string
	// EmbedModel is the Ollama model used for semantic candidate retrieval and
	// weighted reranking. When non-empty, embedding neighbors are merged with
	// structural candidates, filtered by cosine agreement, and ranked by the
	// combined semantic and structural score.
	// Requires Ollama to be running locally.
	// Recommended: "qwen3-embedding:0.6b"
	EmbedModel string
	// MinSimilarity is the lower bound (inclusive) of the similarity range
	// to report, in percent (0-100).
	MinSimilarity int
	// MaxSimilarity is the upper bound (inclusive) of the similarity range
	// to report, in percent (0-100).
	MaxSimilarity int
	// ShowAll includes the [fingerprints] section with per-symbol fingerprints
	// and raw token data. When false (default), only [duplicates] is emitted.
	ShowAll bool
}

// DefaultFingerprintOptions returns sensible defaults.
func DefaultFingerprintOptions() *FingerprintOptions {
	return &FingerprintOptions{
		OutputPath:    config.DefaultFingerprintOutput,
		MinSimilarity: config.DefaultFingerprintMinSimilarity,
		MaxSimilarity: config.DefaultFingerprintMaxSimilarity,
		ShowAll:       config.DefaultFingerprintShowAll,
	}
}

// FingerprintResult holds the counts returned from EmitFingerprints.
type FingerprintResult struct {
	SymbolsFingerprinted int
	DuplicatePairs       int
}

// EmitFingerprints computes fuzzy fingerprints for all symbols with body
// tokens, finds near-duplicate pairs within the configured similarity range,
// and writes a .skt file containing the results.
//
// When opts.EmbedModel is set, CTPH candidates are reranked using cosine
// similarity of Ollama embeddings, filtering out structural false positives.
func (e *Emitter) EmitFingerprints(sg *ir.SymbolGraph, opts *FingerprintOptions) (*FingerprintResult, error) {
	entries := collectFingerprints(sg)
	matches := findDuplicates(entries, opts.MinSimilarity, opts.MaxSimilarity)

	if opts.EmbedModel != "" && len(entries) > 1 {
		var err error
		matches, err = rerank(entries, matches, opts)
		if err != nil {
			return nil, err
		}
	}

	content := renderFingerprints(entries, matches, opts)

	if dir := filepath.Dir(opts.OutputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories
			return nil, fmt.Errorf("create output directory: %w", err)
		}
	}
	if err := os.WriteFile(opts.OutputPath, []byte(content), 0o600); err != nil {
		return nil, fmt.Errorf("write fingerprints: %w", err)
	}

	return &FingerprintResult{
		SymbolsFingerprinted: len(entries),
		DuplicatePairs:       len(matches),
	}, nil
}

// fpEntry is a fingerprinted symbol ready for comparison and emission.
type fpEntry struct {
	filePath   string
	name       string
	fp         string
	sourceBody string // raw source code of the symbol body (for embedding)
	category   plugin.SymbolCategory
	tokens     []byte
	span       [2]int // line range [start, end] in the source file
}

// fpMatch is a duplicate pair with its similarity score.
type fpMatch struct {
	cosineSim  float64
	rrfScore   float64
	i, j       int
	similarity int // best of CTPH and token-edit similarity
	ctphSim    int // raw CTPH similarity
	tokenSim   int // raw token-edit similarity
}

// collectFingerprints walks the SymbolGraph and hashes every symbol with
// body tokens. Symbols whose tokens are too short for fuzzy hashing are skipped.
// For each entry, the raw source lines are read from disk (when available)
// so the embedding model can operate on real code rather than pseudo-code.
func collectFingerprints(sg *ir.SymbolGraph) []fpEntry {
	// Cache file contents to avoid re-reading the same file for every symbol.
	fileCache := make(map[string][]string)

	entries := make([]fpEntry, 0, len(sg.Symbols))
	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		if len(sym.BodyTokens) == 0 {
			continue
		}
		fp := fingerprint.Hash(sym.BodyTokens)
		if fp == "" {
			continue
		}

		src := extractSourceBody(sym.FilePath, sym.Span, fileCache)

		entries = append(entries, fpEntry{
			filePath:   sym.FilePath,
			name:       sym.EffectiveScopedName(),
			fp:         fp,
			tokens:     sym.BodyTokens,
			sourceBody: src,
			span:       sym.Span,
			category:   sym.Category,
		})
	}
	return entries
}

// extractSourceBody reads the source lines for a symbol from disk.
// Returns "" if the file cannot be read or the span is invalid.
func extractSourceBody(filePath string, span [2]int, cache map[string][]string) string {
	if span[0] <= 0 || span[1] <= 0 || span[0] > span[1] {
		return ""
	}

	lines, ok := cache[filePath]
	if !ok {
		data, err := os.ReadFile(filePath) //nolint:gosec // file paths come from the parsed symbol graph
		if err != nil {
			cache[filePath] = nil
			return ""
		}
		lines = strings.Split(string(data), "\n")
		cache[filePath] = lines
	}
	if lines == nil {
		return ""
	}

	// Span is 1-based [start, end] inclusive.
	start := span[0] - 1
	end := span[1]
	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

// annThreshold is the minimum number of entries above which the ANN
// pre-filter is used instead of brute-force O(N²) pairwise comparison.
// Below this threshold, brute-force is fast enough and avoids the overhead
// of building the ANN index.
const annThreshold = 500

// findDuplicates compares symbol pairs using both CTPH fuzzy hashing
// and direct token-edit similarity. The final similarity is the average of
// both signals — requiring agreement from both reduces false positives from
// short functions with similar control flow but different semantics.
//
// For large codebases (≥500 symbols), an ANN pre-filter narrows the search
// to the top-K nearest neighbors per symbol, reducing complexity from O(N²)
// to O(N·K).
func findDuplicates(entries []fpEntry, minSim, maxSim int) []fpMatch {
	if len(entries) >= annThreshold {
		return findDuplicatesANN(entries, minSim, maxSim)
	}
	return findDuplicatesBrute(entries, minSim, maxSim)
}

// findDuplicatesBrute is the O(N²) pairwise comparison for small codebases.
func findDuplicatesBrute(entries []fpEntry, minSim, maxSim int) []fpMatch {
	var matches []fpMatch
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			if m, ok := scorePair(entries, i, j, minSim, maxSim); ok {
				matches = append(matches, m)
			}
		}
	}
	sort.Slice(matches, func(a, b int) bool {
		return matches[a].similarity > matches[b].similarity
	})
	return matches
}

// findDuplicatesANN uses a lightweight ANN index built from structural token
// bigrams to narrow the candidate set to top-K neighbors per symbol, then
// scores only those candidates with the full CTPH + token-edit pipeline.
func findDuplicatesANN(entries []fpEntry, minSim, maxSim int) []fpMatch {
	groups := make(map[plugin.SymbolCategory][]int)
	for i := range entries {
		groups[entries[i].category] = append(groups[entries[i].category], i)
	}

	categories := make([]plugin.SymbolCategory, 0, len(groups))
	for category := range groups {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i] < categories[j]
	})

	var matches []fpMatch
	for _, category := range categories {
		entryIndexes := groups[category]
		if len(entryIndexes) < 2 {
			continue
		}

		streams := make([][]byte, len(entryIndexes))
		for i, entryIndex := range entryIndexes {
			streams[i] = entries[entryIndex].tokens
		}

		var idx *fingerprint.ANNIndex
		switch category {
		case plugin.CategoryType, plugin.CategoryValue:
			idx = fingerprint.BuildRawANNIndex(streams)
		default:
			idx = fingerprint.BuildANNIndex(streams)
		}

		for _, candidate := range idx.FindCandidates(fingerprint.DefaultANNK()) {
			i := entryIndexes[candidate.I]
			j := entryIndexes[candidate.J]
			if m, ok := scorePair(entries, i, j, minSim, maxSim); ok {
				matches = append(matches, m)
			}
		}
	}
	sort.Slice(matches, func(a, b int) bool {
		return matches[a].similarity > matches[b].similarity
	})
	return matches
}

// scorePair computes the combined similarity for a single (i, j) pair.
// Returns the match and true if it falls within [minSim, maxSim].
func scorePair(entries []fpEntry, i, j, minSim, maxSim int) (fpMatch, bool) {
	if entries[i].category != entries[j].category {
		return fpMatch{}, false
	}

	ctph := fingerprint.Distance(entries[i].fp, entries[j].fp)
	if ctph < 0 {
		ctph = 0
	}
	var tokSim int
	switch entries[i].category {
	case plugin.CategoryType, plugin.CategoryValue:
		tokSim = fingerprint.RawTokenEditSimilarity(entries[i].tokens, entries[j].tokens)
	default:
		tokSim = fingerprint.TokenEditSimilarity(entries[i].tokens, entries[j].tokens)
	}

	// Average both signals — a pair must score well on both to pass.
	sim := (ctph + tokSim) / 2

	if sim >= minSim && sim <= maxSim {
		return fpMatch{
			i:          i,
			j:          j,
			similarity: sim,
			ctphSim:    ctph,
			tokenSim:   tokSim,
		}, true
	}
	return fpMatch{}, false
}

// tokenText maps single-byte FP tokens to readable pseudo-code words.
var tokenText = map[byte]string{
	plugin.FPIf:      "if",
	plugin.FPElse:    "else",
	plugin.FPElseIf:  "elif",
	plugin.FPFor:     "for",
	plugin.FPWhile:   "while",
	plugin.FPReturn:  "return",
	plugin.FPSwitch:  "switch",
	plugin.FPCase:    "case",
	plugin.FPBreak:   "break",
	plugin.FPCont:    "continue",
	plugin.FPTry:     "try",
	plugin.FPCatch:   "catch",
	plugin.FPThrow:   "throw",
	plugin.FPYield:   "yield",
	plugin.FPAwait:   "await",
	plugin.FPGo:      "go",
	plugin.FPSelect:  "select",
	plugin.FPDefer:   "defer",
	plugin.FPAssign:  "assign",
	plugin.FPMember:  "member",
	plugin.FPIndex:   "index",
	plugin.FPNew:     "new",
	plugin.FPCast:    "cast",
	plugin.FPLambda:  "lambda",
	plugin.FPRange:   "range",
	plugin.FPMatch:   "match",
	plugin.FPDelete:  "delete",
	plugin.FPAdd:     "+",
	plugin.FPSub:     "-",
	plugin.FPMul:     "*",
	plugin.FPDiv:     "/",
	plugin.FPMod:     "%",
	plugin.FPEq:      "==",
	plugin.FPNeq:     "!=",
	plugin.FPLt:      "<",
	plugin.FPGt:      ">",
	plugin.FPLte:     "<=",
	plugin.FPGte:     ">=",
	plugin.FPAnd:     "&&",
	plugin.FPOr:      "||",
	plugin.FPNot:     "!",
	plugin.FPBitAnd:  "&",
	plugin.FPBitOr:   "|",
	plugin.FPBitXor:  "^",
	plugin.FPBitNot:  "~",
	plugin.FPShl:     "<<",
	plugin.FPShr:     ">>",
	plugin.FPLitNum:  "num",
	plugin.FPLitStr:  "str",
	plugin.FPLitBool: "bool",
	plugin.FPLitNil:  "nil",
	plugin.FPDict:    "dict",
	plugin.FPArray:   "array",
}

// tokensToText reconstructs a human-readable pseudo-code summary from a
// BodyTokens byte stream so the embedding model can reason about what the
// code does rather than seeing opaque bytes.
//
// Stream format (from walkBodyWithScope in langutil.go):
//   - Most tokens: single byte mapped via tokenText
//   - Variable ordinals: single byte 0x01–0xFF not in tokenText → "varN"
//   - FPCall: FPCall + 2-byte callee hash + 1-byte arg count
//   - Literals: token byte + 4-byte FNV hash of the value
//   - Dict/Array: token byte + 1-byte element count + N*4-byte hashes
func tokensToText(tokens []byte) string {
	var b strings.Builder
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		i++

		// Tagged variable ordinal.
		if tok == plugin.FPVar {
			if i < len(tokens) {
				fmt.Fprintf(&b, "var%d ", tokens[i])
				i++
			}
			continue
		}

		// FPCall: skip 2-byte callee hash, read arg count.
		if tok == plugin.FPCall {
			skip := min(2, len(tokens)-i)
			i += skip
			if i < len(tokens) {
				fmt.Fprintf(&b, "call(%d) ", tokens[i])
				i++
			} else {
				b.WriteString("call ")
			}
			continue
		}

		// Literals: emit type word, skip 4-byte value hash.
		if tok == plugin.FPLitNum || tok == plugin.FPLitStr ||
			tok == plugin.FPLitBool || tok == plugin.FPLitNil {
			b.WriteString(tokenText[tok])
			b.WriteByte(' ')
			i += min(4, len(tokens)-i)
			continue
		}

		// Dict/array: emit type word, skip 1-byte count + N*4-byte hashes.
		if tok == plugin.FPDict || tok == plugin.FPArray {
			b.WriteString(tokenText[tok])
			b.WriteByte(' ')
			if i < len(tokens) {
				count := int(tokens[i])
				i += 1 + count*4
				if i > len(tokens) {
					i = len(tokens)
				}
			}
			continue
		}

		// Known semantic token.
		if word, ok := tokenText[tok]; ok {
			b.WriteString(word)
			b.WriteByte(' ')
			continue
		}

		// Variable ordinal (not in tokenText): emit "varN".
		if tok > 0 {
			fmt.Fprintf(&b, "var%d ", tok)
		}
	}
	return strings.TrimSpace(b.String())
}

// Reranking weights for the weighted linear combination of signals.
// Cosine similarity from a code-trained embedding model is the strongest
// signal for semantic clones (Type 3/4), so it gets the highest weight.
// Token-edit similarity captures structural shape missed by CTPH boundary
// shifts. CTPH is fast and good for exact/near matches but noisy for
// diverged clones.
const (
	weightCosine   = 0.50
	weightTokenSim = 0.30
	weightCTPH     = 0.20
)

// cosineFloorBase is the minimum cosine similarity any pair must meet.
// The actual threshold is dynamic: max(cosineFloorBase, structuralSim/100 - cosineFloorMargin).
// High structural similarity demands higher cosine agreement — this catches
// "same shape, different semantics" false positives (e.g. two unrelated
// constants that both look like `const X = 0xNN`).
const (
	cosineFloorBase   = 0.65
	cosineFloorMargin = 0.15
)

// cosineFloor computes the dynamic cosine threshold for a given structural
// similarity score. Higher structural similarity requires higher cosine
// agreement to survive, because structurally-similar-but-semantically-different
// pairs are the primary source of false positives.
func cosineFloor(structuralSim int) float64 {
	dynamic := float64(structuralSim)/100.0 - cosineFloorMargin
	if dynamic > cosineFloorBase {
		return dynamic
	}
	return cosineFloorBase
}

const embeddingBatchSize = 64

// rerank embeds every symbol so semantic nearest neighbors can contribute
// candidates that structural retrieval did not find. It unions both candidate
// sets, applies the dynamic cosine floor, and filters on the final weighted
// similarity range.
func rerank(entries []fpEntry, matches []fpMatch, opts *FingerprintOptions) ([]fpMatch, error) {
	texts := make([]string, len(entries))
	for i := range entries {
		texts[i] = embedInput(&entries[i])
	}

	client := ollama.NewClient("", opts.EmbedModel)
	vecs := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += embeddingBatchSize {
		end := min(start+embeddingBatchSize, len(texts))
		batch, err := client.Embed(texts[start:end])
		if err != nil {
			return nil, fmt.Errorf("embed symbols %d-%d: %w", start+1, end, err)
		}
		vecs = append(vecs, batch...)
	}

	return rerankWithVectors(entries, matches, vecs, opts), nil
}

func rerankWithVectors(entries []fpEntry, matches []fpMatch, vecs [][]float32, opts *FingerprintOptions) []fpMatch {
	if len(entries) != len(vecs) {
		return nil
	}

	semantic := findSemanticCandidates(entries, vecs, semanticNeighborK)
	byPair := make(map[candidatePair]fpMatch, len(matches)+len(semantic))
	structuralPairs := make(map[candidatePair]struct{}, len(matches))
	for _, match := range matches {
		key := newCandidatePair(match.i, match.j)
		byPair[key] = match
		structuralPairs[key] = struct{}{}
	}
	for _, candidate := range semantic {
		key := newCandidatePair(candidate.i, candidate.j)
		if _, exists := byPair[key]; exists {
			continue
		}
		if match, ok := scorePair(entries, key.i, key.j, 0, 100); ok {
			byPair[key] = match
		}
	}

	filtered := make([]fpMatch, 0, len(byPair))
	for key, match := range byPair {
		match.cosineSim = ollama.CosineSimilarity(vecs[match.i], vecs[match.j])

		// Dynamic cosine floor: higher structural similarity demands
		// higher cosine agreement. Drop pairs the embedding model rejects.
		floor := cosineFloor(match.similarity)
		if _, structurallyRetrieved := structuralPairs[key]; !structurallyRetrieved {
			floor = max(floor, semanticCandidateCosineFloor)
		}
		if match.cosineSim < floor {
			continue
		}

		// Normalize all signals to [0, 1].
		normCTPH := float64(match.ctphSim) / 100.0
		normToken := float64(match.tokenSim) / 100.0
		normCosine := match.cosineSim
		if normCosine < 0 {
			normCosine = 0
		}

		match.rrfScore = weightCosine*normCosine + weightTokenSim*normToken + weightCTPH*normCTPH
		weightedPercent := match.rrfScore * 100
		if weightedPercent < float64(opts.MinSimilarity) || weightedPercent > float64(opts.MaxSimilarity) {
			continue
		}
		match.similarity = int(math.Round(weightedPercent))
		filtered = append(filtered, match)
	}

	// Re-sort survivors by weighted score descending.
	sort.Slice(filtered, func(a, b int) bool {
		if filtered[a].rrfScore != filtered[b].rrfScore {
			return filtered[a].rrfScore > filtered[b].rrfScore
		}
		if filtered[a].i != filtered[b].i {
			return filtered[a].i < filtered[b].i
		}
		return filtered[a].j < filtered[b].j
	})
	return filtered
}

// renderFingerprints formats the output .skt content.
// The [fingerprints] section is included only when opts.ShowAll is true.
// The [duplicates] section is always present.
func renderFingerprints(entries []fpEntry, matches []fpMatch, opts *FingerprintOptions) string {
	var b strings.Builder

	if opts.ShowAll {
		b.WriteString("[fingerprints]\n")
		byFile := make(map[string][]fpEntry)
		for _, e := range entries {
			byFile[e.filePath] = append(byFile[e.filePath], e)
		}
		fileOrder := make([]string, 0, len(byFile))
		for fp := range byFile {
			fileOrder = append(fileOrder, fp)
		}
		sort.Strings(fileOrder)
		for _, fp := range fileOrder {
			fmt.Fprintf(&b, "## %s\n", fp)
			for _, e := range byFile[fp] {
				fmt.Fprintf(&b, "%s  FP:%s  tokens:%s\n",
					e.name, e.fp, hex.EncodeToString(e.tokens))
			}
		}
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "[duplicates]\n# similarity range: %d%%-%d%%\n",
		opts.MinSimilarity, opts.MaxSimilarity)
	if opts.EmbedModel != "" {
		fmt.Fprintf(&b, "# semantic model: %s  ranking: weighted(%.0f%% cosine, %.0f%% token, %.0f%% ctph)  cosine floor: %.2f+  semantic top-k: %d floor: %.2f\n",
			opts.EmbedModel, weightCosine*100, weightTokenSim*100, weightCTPH*100,
			cosineFloorBase, semanticNeighborK, semanticCandidateCosineFloor)
	}

	if len(matches) > 0 {
		for _, m := range matches {
			ei, ej := entries[m.i], entries[m.j]
			if m.rrfScore > 0 {
				fmt.Fprintf(&b, "similarity:%d%%  cosine:%.2f  score:%.4f  %s::%s <-> %s::%s\n",
					m.similarity, m.cosineSim, m.rrfScore, ei.filePath, ei.name, ej.filePath, ej.name)
			} else {
				fmt.Fprintf(&b, "similarity:%d%%  %s::%s <-> %s::%s\n",
					m.similarity, ei.filePath, ei.name, ej.filePath, ej.name)
			}
		}
	} else {
		b.WriteString("# no duplicates found\n")
	}

	return b.String()
}
