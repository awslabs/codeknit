// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"strings"
	"testing"

	"codeknit/internal/config"
	"codeknit/internal/plugin"
)

// ---------------------------------------------------------------------------
// tokensToText
// ---------------------------------------------------------------------------

func TestTokensToText_Empty(t *testing.T) {
	if got := tokensToText(nil); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if got := tokensToText([]byte{}); got != "" {
		t.Errorf("expected empty string for empty slice, got %q", got)
	}
}

func TestTokensToText_SimpleControlFlow(t *testing.T) {
	// if return
	tokens := []byte{plugin.FPIf, plugin.FPReturn}
	got := tokensToText(tokens)
	if got != "if return" {
		t.Errorf("got %q, want %q", got, "if return")
	}
}

func TestTokensToText_AllControlFlowTokens(t *testing.T) {
	cases := []struct {
		want string
		tok  byte
	}{
		{"if", plugin.FPIf},
		{"else", plugin.FPElse},
		{"elif", plugin.FPElseIf},
		{"for", plugin.FPFor},
		{"while", plugin.FPWhile},
		{"return", plugin.FPReturn},
		{"switch", plugin.FPSwitch},
		{"case", plugin.FPCase},
		{"break", plugin.FPBreak},
		{"continue", plugin.FPCont},
		{"try", plugin.FPTry},
		{"catch", plugin.FPCatch},
		{"throw", plugin.FPThrow},
		{"yield", plugin.FPYield},
		{"await", plugin.FPAwait},
		{"go", plugin.FPGo},
		{"select", plugin.FPSelect},
		{"defer", plugin.FPDefer},
	}
	for _, tc := range cases {
		got := tokensToText([]byte{tc.tok})
		if got != tc.want {
			t.Errorf("token 0x%02x: got %q, want %q", tc.tok, got, tc.want)
		}
	}
}

func TestTokensToText_Operators(t *testing.T) {
	tokens := []byte{plugin.FPAdd, plugin.FPEq, plugin.FPAnd, plugin.FPLt}
	got := tokensToText(tokens)
	if got != "+ == && <" {
		t.Errorf("got %q, want %q", got, "+ == && <")
	}
}

func TestTokensToText_Call_WithArgCount(t *testing.T) {
	// FPCall + 2-byte callee hash + 1-byte arg count (3)
	tokens := []byte{plugin.FPCall, 0xAB, 0xCD, 3}
	got := tokensToText(tokens)
	if got != "call(3)" {
		t.Errorf("got %q, want %q", got, "call(3)")
	}
}

func TestTokensToText_Call_ZeroArgs(t *testing.T) {
	tokens := []byte{plugin.FPCall, 0x00, 0x00, 0}
	got := tokensToText(tokens)
	if got != "call(0)" {
		t.Errorf("got %q, want %q", got, "call(0)")
	}
}

func TestTokensToText_Call_Truncated(t *testing.T) {
	// Stream ends before arg count byte — should not panic.
	tokens := []byte{plugin.FPCall, 0xAB} // only 1 of 2 hash bytes
	got := tokensToText(tokens)
	if got != "call" {
		t.Errorf("got %q, want %q", got, "call")
	}
}

func TestTokensToText_Literal_Num(t *testing.T) {
	// FPLitNum + 4-byte hash
	tokens := []byte{plugin.FPLitNum, 0x01, 0x02, 0x03, 0x04}
	got := tokensToText(tokens)
	if got != "num" {
		t.Errorf("got %q, want %q", got, "num")
	}
}

func TestTokensToText_Literal_Str(t *testing.T) {
	tokens := []byte{plugin.FPLitStr, 0xAA, 0xBB, 0xCC, 0xDD}
	got := tokensToText(tokens)
	if got != "str" {
		t.Errorf("got %q, want %q", got, "str")
	}
}

func TestTokensToText_Literal_Bool(t *testing.T) {
	tokens := []byte{plugin.FPLitBool, 0x00, 0x00, 0x00, 0x01}
	got := tokensToText(tokens)
	if got != "bool" {
		t.Errorf("got %q, want %q", got, "bool")
	}
}

func TestTokensToText_Literal_Nil(t *testing.T) {
	tokens := []byte{plugin.FPLitNil, 0x00, 0x00, 0x00, 0x00}
	got := tokensToText(tokens)
	if got != "nil" {
		t.Errorf("got %q, want %q", got, "nil")
	}
}

func TestTokensToText_Literal_Truncated(t *testing.T) {
	// Only 2 of 4 hash bytes present — should not panic.
	tokens := []byte{plugin.FPLitNum, 0x01, 0x02}
	got := tokensToText(tokens)
	if got != "num" {
		t.Errorf("got %q, want %q", got, "num")
	}
}

func TestTokensToText_Array(t *testing.T) {
	// FPArray + count=2 + 2*4 hash bytes
	tokens := []byte{plugin.FPArray, 2, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	got := tokensToText(tokens)
	if got != "array" {
		t.Errorf("got %q, want %q", got, "array")
	}
}

func TestTokensToText_Dict(t *testing.T) {
	// FPDict + count=1 + 1*4 hash bytes
	tokens := []byte{plugin.FPDict, 1, 0xAA, 0xBB, 0xCC, 0xDD}
	got := tokensToText(tokens)
	if got != "dict" {
		t.Errorf("got %q, want %q", got, "dict")
	}
}

func TestTokensToText_Dict_Truncated(t *testing.T) {
	// count says 3 elements but stream ends early — should not panic.
	tokens := []byte{plugin.FPDict, 3, 0x01, 0x02}
	got := tokensToText(tokens)
	if got != "dict" {
		t.Errorf("got %q, want %q", got, "dict")
	}
}

func TestTokensToText_VariableOrdinals(t *testing.T) {
	// Use ordinal values that don't collide with any defined FP token.
	// 0x50, 0x51, 0x52 are not assigned to any token constant.
	tokens := []byte{0x50, 0x51, 0x52}
	got := tokensToText(tokens)
	if got != "var80 var81 var82" {
		t.Errorf("got %q, want %q", got, "var80 var81 var82")
	}
}

func TestTokensToText_MixedStream(t *testing.T) {
	// Simulates: if var80 == num { return call(1 arg) }
	// Uses 0x50 as a variable ordinal (not a defined token).
	tokens := []byte{
		plugin.FPIf,
		0x50, // var80 (not a defined token)
		plugin.FPEq,
		plugin.FPLitNum, 0x00, 0x00, 0x00, 0x2A, // num (value=42)
		plugin.FPReturn,
		plugin.FPCall, 0xAB, 0xCD, 1, // call(1)
	}
	got := tokensToText(tokens)
	want := "if var80 == num return call(1)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTokensToText_NoTrailingSpace(t *testing.T) {
	tokens := []byte{plugin.FPIf, plugin.FPReturn}
	got := tokensToText(tokens)
	if strings.HasSuffix(got, " ") {
		t.Errorf("result has trailing space: %q", got)
	}
}

// ---------------------------------------------------------------------------
// embedInput
// ---------------------------------------------------------------------------

func TestEmbedInput_WithTokens(t *testing.T) {
	e := fpEntry{
		filePath:   "pkg/foo.go",
		name:       "myFunc",
		tokens:     []byte{plugin.FPIf, plugin.FPReturn},
		sourceBody: "if x > 0 {\n\treturn x\n}",
	}
	got := embedInput(&e)
	if !strings.HasPrefix(got, "symbol:myFunc file:pkg/foo.go\n") {
		t.Errorf("unexpected format: %q", got)
	}
	if !strings.Contains(got, "if x > 0") || !strings.Contains(got, "return x") {
		t.Errorf("body missing expected source code: %q", got)
	}
}

func TestEmbedInput_EmptyTokens(t *testing.T) {
	e := fpEntry{
		filePath: "pkg/foo.go",
		name:     "myFunc",
		tokens:   []byte{},
	}
	got := embedInput(&e)
	want := "symbol:myFunc file:pkg/foo.go"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmbedInput_NilTokens(t *testing.T) {
	e := fpEntry{filePath: "a.go", name: "fn", tokens: nil}
	got := embedInput(&e)
	if got != "symbol:fn file:a.go" {
		t.Errorf("got %q", got)
	}
}

// ---------------------------------------------------------------------------
// findDuplicates
// ---------------------------------------------------------------------------

func makeEntry(name, fp string) fpEntry {
	return fpEntry{filePath: "f.go", name: name, fp: fp, tokens: []byte{plugin.FPIf}}
}

func TestFindDuplicates_Empty(t *testing.T) {
	matches := findDuplicates(nil, 50, 100)
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}

func TestFindDuplicates_SingleEntry(t *testing.T) {
	entries := []fpEntry{makeEntry("a", "2:aabb:ccdd")}
	matches := findDuplicates(entries, 50, 100)
	if len(matches) != 0 {
		t.Errorf("single entry should produce no matches, got %d", len(matches))
	}
}

func TestFindDuplicates_IdenticalFingerprints(t *testing.T) {
	fp := "2:aabbccdd:eeff0011"
	entries := []fpEntry{
		makeEntry("a", fp),
		makeEntry("b", fp),
	}
	matches := findDuplicates(entries, 0, 100)
	if len(matches) == 0 {
		t.Error("identical fingerprints should produce at least one match")
	}
	if matches[0].similarity != 100 {
		t.Errorf("identical fingerprints should score 100, got %d", matches[0].similarity)
	}
}

func TestFindDuplicates_SortedDescending(t *testing.T) {
	fp := "2:aabbccdd:eeff0011"
	entries := []fpEntry{
		makeEntry("a", fp),
		makeEntry("b", fp),
		makeEntry("c", fp),
	}
	matches := findDuplicates(entries, 0, 100)
	for k := 1; k < len(matches); k++ {
		if matches[k].similarity > matches[k-1].similarity {
			t.Errorf("matches not sorted descending at index %d", k)
		}
	}
}

func TestFindDuplicates_RangeFilter(t *testing.T) {
	fp := "2:aabbccdd:eeff0011"
	entries := []fpEntry{
		makeEntry("a", fp),
		makeEntry("b", fp),
	}
	// Exact duplicates score 100 — should be excluded when maxSim=99.
	matches := findDuplicates(entries, 50, 99)
	if len(matches) != 0 {
		t.Errorf("score 100 should be excluded by maxSim=99, got %d matches", len(matches))
	}
}

// ---------------------------------------------------------------------------
// renderFingerprints
// ---------------------------------------------------------------------------

func TestDefaultFingerprintOptions_SimilarityRange(t *testing.T) {
	opts := DefaultFingerprintOptions()
	if opts.OutputPath != config.DefaultFingerprintOutput {
		t.Fatalf("OutputPath default: got %q, want %q", opts.OutputPath, config.DefaultFingerprintOutput)
	}
	if opts.MinSimilarity != config.DefaultFingerprintMinSimilarity {
		t.Fatalf("MinSimilarity default: got %d, want %d", opts.MinSimilarity, config.DefaultFingerprintMinSimilarity)
	}
	if opts.MaxSimilarity != config.DefaultFingerprintMaxSimilarity {
		t.Fatalf("MaxSimilarity default: got %d, want %d", opts.MaxSimilarity, config.DefaultFingerprintMaxSimilarity)
	}
	if opts.ShowAll != config.DefaultFingerprintShowAll {
		t.Fatalf("ShowAll default: got %t, want %t", opts.ShowAll, config.DefaultFingerprintShowAll)
	}
}

func TestRenderFingerprints_NoDuplicates(t *testing.T) {
	opts := &FingerprintOptions{MinSimilarity: 75, MaxSimilarity: 100}
	out := renderFingerprints(nil, nil, opts)
	if !strings.Contains(out, "[duplicates]") {
		t.Error("output must contain [duplicates] section")
	}
	if !strings.Contains(out, "no duplicates found") {
		t.Error("output must say no duplicates found")
	}
}

func TestRenderFingerprints_WithMatch(t *testing.T) {
	entries := []fpEntry{
		{filePath: "a/foo.go", name: "Foo", fp: "x", tokens: []byte{plugin.FPIf}},
		{filePath: "b/bar.go", name: "Bar", fp: "x", tokens: []byte{plugin.FPIf}},
	}
	matches := []fpMatch{{i: 0, j: 1, similarity: 90}}
	opts := &FingerprintOptions{MinSimilarity: 75, MaxSimilarity: 100}
	out := renderFingerprints(entries, matches, opts)

	if !strings.Contains(out, "similarity:90%") {
		t.Errorf("missing similarity line in output:\n%s", out)
	}
	if !strings.Contains(out, "a/foo.go::Foo") {
		t.Errorf("missing first symbol in output:\n%s", out)
	}
	if !strings.Contains(out, "b/bar.go::Bar") {
		t.Errorf("missing second symbol in output:\n%s", out)
	}
}

func TestRenderFingerprints_WithCosine(t *testing.T) {
	entries := []fpEntry{
		{filePath: "a.go", name: "A", fp: "x", tokens: []byte{plugin.FPIf}},
		{filePath: "b.go", name: "B", fp: "x", tokens: []byte{plugin.FPIf}},
	}
	matches := []fpMatch{{i: 0, j: 1, similarity: 85, cosineSim: 0.92, rrfScore: 0.0312}}
	opts := &FingerprintOptions{
		MinSimilarity: 75, MaxSimilarity: 100,
		EmbedModel: "qwen3-embedding:0.6b",
	}
	out := renderFingerprints(entries, matches, opts)

	if !strings.Contains(out, "cosine:0.92") {
		t.Errorf("missing cosine score in output:\n%s", out)
	}
	if !strings.Contains(out, "score:0.0312") {
		t.Errorf("missing weighted score in output:\n%s", out)
	}
	if !strings.Contains(out, "semantic model: qwen3-embedding:0.6b") {
		t.Errorf("missing model header in output:\n%s", out)
	}
	if !strings.Contains(out, "weighted(") {
		t.Errorf("missing weighted method in header:\n%s", out)
	}
}

func TestRenderFingerprints_ShowAll(t *testing.T) {
	entries := []fpEntry{
		{filePath: "a.go", name: "Fn", fp: "2:aa:bb", tokens: []byte{plugin.FPReturn}},
	}
	opts := &FingerprintOptions{MinSimilarity: 75, MaxSimilarity: 100, ShowAll: true}
	out := renderFingerprints(entries, nil, opts)

	if !strings.Contains(out, "[fingerprints]") {
		t.Error("ShowAll=true must include [fingerprints] section")
	}
	if !strings.Contains(out, "## a.go") {
		t.Error("ShowAll must group by file")
	}
	if !strings.Contains(out, "Fn") {
		t.Error("ShowAll must include symbol name")
	}
}

func TestRenderFingerprints_SimilarityRangeHeader(t *testing.T) {
	opts := &FingerprintOptions{MinSimilarity: 65, MaxSimilarity: 95}
	out := renderFingerprints(nil, nil, opts)
	if !strings.Contains(out, "65%-95%") {
		t.Errorf("missing similarity range header in:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Weighted scoring
// ---------------------------------------------------------------------------

func TestWeightedScoring_HighOnAllSignalsRanksFirst(t *testing.T) {
	// Pair A: perfect on all signals → highest score
	// Pair B: good CTPH, poor cosine → lower score
	scoreA := weightCosine*1.0 + weightTokenSim*1.0 + weightCTPH*1.0
	scoreB := weightCosine*0.3 + weightTokenSim*0.8 + weightCTPH*0.9

	if scoreA <= scoreB {
		t.Errorf("pair scoring high on all signals should outscore partial match: %.6f vs %.6f", scoreA, scoreB)
	}
}

func TestWeightedScoring_FalsePositiveSinksToBottom(t *testing.T) {
	// A false positive: high CTPH but low cosine and low token similarity.
	// A true duplicate: high cosine and high token similarity but mid CTPH.
	falsePositive := weightCosine*0.2 + weightTokenSim*0.3 + weightCTPH*1.0
	trueDuplicate := weightCosine*0.95 + weightTokenSim*0.9 + weightCTPH*0.5

	if falsePositive >= trueDuplicate {
		t.Errorf("false positive (high CTPH, low cosine) should score lower than true duplicate: %.6f vs %.6f",
			falsePositive, trueDuplicate)
	}
}

func TestWeightedScoring_WeightsSumToOne(t *testing.T) {
	sum := weightCosine + weightTokenSim + weightCTPH
	if sum < 0.999 || sum > 1.001 {
		t.Errorf("weights should sum to 1.0, got %f", sum)
	}
}

func TestWeightedScoring_CosineHasHighestWeight(t *testing.T) {
	if weightCosine <= weightTokenSim || weightCosine <= weightCTPH {
		t.Errorf("cosine should have the highest weight: cosine=%f token=%f ctph=%f",
			weightCosine, weightTokenSim, weightCTPH)
	}
}
