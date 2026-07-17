// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/hex"
	"fmt"
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
	// EmbedModel is the Ollama model name to use for semantic reranking of
	// CTPH candidates via weighted scoring. When non-empty, all CTPH
	// candidates are re-sorted by a weighted combination of structural
	// similarity and cosine similarity — none are dropped.
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

	if opts.EmbedModel != "" && len(matches) > 0 {
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

	// Cap at ~2000 chars to avoid blowing up embedding context windows.
	body := strings.Join(lines[start:end], "\n")
	const maxEmbedLen = 2000
	if len(body) > maxEmbedLen {
		body = body[:maxEmbedLen]
	}
	return body
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
	// Build token streams slice for the ANN index.
	streams := make([][]byte, len(entries))
	for i := range entries {
		streams[i] = entries[i].tokens
	}

	idx := fingerprint.BuildANNIndex(streams)
	candidates := idx.FindCandidates(fingerprint.DefaultANNK())

	var matches []fpMatch
	for _, c := range candidates {
		if m, ok := scorePair(entries, c.I, c.J, minSim, maxSim); ok {
			matches = append(matches, m)
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
	ctph := fingerprint.Distance(entries[i].fp, entries[j].fp)
	if ctph < 0 {
		ctph = 0
	}
	tokSim := fingerprint.TokenEditSimilarity(entries[i].tokens, entries[j].tokens)

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

// embedInput builds the text sent to the embedding model for a symbol.
// It uses the raw source code read from disk — embedding models are trained
// on real code and produce much better semantic vectors than pseudo-code.
func embedInput(e *fpEntry) string {
	if e.sourceBody != "" {
		return fmt.Sprintf("symbol:%s file:%s\n%s", e.name, e.filePath, e.sourceBody)
	}
	return fmt.Sprintf("symbol:%s file:%s", e.name, e.filePath)
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

// rerank scores each CTPH candidate with cosine similarity via Ollama
// embeddings, applies a dynamic cosine floor to drop false positives, then
// re-sorts survivors by weighted score. Pairs where the embedding model
// disagrees with the structural signal are eliminated — this catches
// "same shape, different semantics" noise like unrelated constants or
// boilerplate methods that happen to share control flow patterns.
func rerank(entries []fpEntry, matches []fpMatch, opts *FingerprintOptions) ([]fpMatch, error) {
	// Collect unique entry indexes that appear in at least one match.
	needed := make(map[int]struct{}, len(matches)*2)
	for _, m := range matches {
		needed[m.i] = struct{}{}
		needed[m.j] = struct{}{}
	}

	idxOrder := make([]int, 0, len(needed))
	for idx := range needed {
		idxOrder = append(idxOrder, idx)
	}
	sort.Ints(idxOrder)

	pos := make(map[int]int, len(idxOrder))
	texts := make([]string, len(idxOrder))
	for i, idx := range idxOrder {
		pos[idx] = i
		texts[i] = embedInput(&entries[idx])
	}

	client := ollama.NewClient("", opts.EmbedModel)
	vecs, err := client.Embed(texts)
	if err != nil {
		return nil, err
	}

	// Score every match and apply the dynamic cosine floor.
	filtered := matches[:0] // reuse backing array
	for k := range matches {
		va := vecs[pos[matches[k].i]]
		vb := vecs[pos[matches[k].j]]
		matches[k].cosineSim = ollama.CosineSimilarity(va, vb)

		// Dynamic cosine floor: higher structural similarity demands
		// higher cosine agreement. Drop pairs the embedding model rejects.
		floor := cosineFloor(matches[k].similarity)
		if matches[k].cosineSim < floor {
			continue
		}

		// Normalize all signals to [0, 1].
		normCTPH := float64(matches[k].ctphSim) / 100.0
		normToken := float64(matches[k].tokenSim) / 100.0
		normCosine := matches[k].cosineSim
		if normCosine < 0 {
			normCosine = 0
		}

		matches[k].rrfScore = weightCosine*normCosine + weightTokenSim*normToken + weightCTPH*normCTPH
		filtered = append(filtered, matches[k])
	}

	// Re-sort survivors by weighted score descending.
	sort.Slice(filtered, func(a, b int) bool {
		return filtered[a].rrfScore > filtered[b].rrfScore
	})
	return filtered, nil
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
		fmt.Fprintf(&b, "# semantic model: %s  ranking: weighted(%.0f%% cosine, %.0f%% token, %.0f%% ctph)  cosine floor: %.2f+\n",
			opts.EmbedModel, weightCosine*100, weightTokenSim*100, weightCTPH*100, cosineFloorBase)
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
