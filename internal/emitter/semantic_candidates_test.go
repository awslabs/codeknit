// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"bytes"
	"testing"

	"codeknit/internal/fingerprint"
	"codeknit/internal/plugin"
)

func TestFindSemanticCandidates_CategoryAware(t *testing.T) {
	entries := []fpEntry{
		{category: plugin.CategoryCallable},
		{category: plugin.CategoryCallable},
		{category: plugin.CategoryCallable},
		{category: plugin.CategoryValue},
	}
	vectors := [][]float32{
		{1, 0},
		{0.99, 0.01},
		{0, 1},
		{1, 0},
	}

	candidates := findSemanticCandidates(entries, vectors, 1)
	if len(candidates) == 0 {
		t.Fatal("expected semantic candidates")
	}
	if candidates[0].i != 0 || candidates[0].j != 1 {
		t.Fatalf("top candidate = (%d,%d), want (0,1)", candidates[0].i, candidates[0].j)
	}
	for _, candidate := range candidates {
		if entries[candidate.i].category != entries[candidate.j].category {
			t.Fatalf("cross-category candidate = (%d,%d)", candidate.i, candidate.j)
		}
	}
}

func TestFindSemanticCandidates_InvalidInput(t *testing.T) {
	entries := []fpEntry{{category: plugin.CategoryCallable}, {category: plugin.CategoryCallable}}
	if got := findSemanticCandidates(entries, nil, 10); got != nil {
		t.Fatalf("mismatched vectors returned %d candidates", len(got))
	}
	if got := findSemanticCandidates(entries, make([][]float32, len(entries)), 0); got != nil {
		t.Fatalf("zero K returned %d candidates", len(got))
	}
}

func TestRerankWithVectors_AddsSemanticOnlyCandidate(t *testing.T) {
	tokens := bytes.Repeat([]byte{plugin.FPIf, plugin.FPReturn}, 32)
	fp := fingerprint.Hash(tokens)
	entries := []fpEntry{
		{category: plugin.CategoryCallable, fp: fp, tokens: tokens},
		{category: plugin.CategoryCallable, fp: fp, tokens: tokens},
	}
	vectors := [][]float32{{1, 0}, {1, 0}}
	opts := &FingerprintOptions{MinSimilarity: 0, MaxSimilarity: 100}

	matches := rerankWithVectors(entries, nil, vectors, opts)
	if len(matches) != 1 {
		t.Fatalf("semantic retrieval produced %d matches, want 1", len(matches))
	}
	if matches[0].i != 0 || matches[0].j != 1 {
		t.Fatalf("semantic match = (%d,%d), want (0,1)", matches[0].i, matches[0].j)
	}
	if matches[0].similarity != 100 || matches[0].cosineSim != 1 {
		t.Fatalf("semantic match scores = similarity:%d cosine:%f, want 100 and 1", matches[0].similarity, matches[0].cosineSim)
	}
}

func TestRerankWithVectors_AppliesRangeToWeightedSimilarity(t *testing.T) {
	tokens := bytes.Repeat([]byte{plugin.FPIf, plugin.FPReturn}, 32)
	fp := fingerprint.Hash(tokens)
	entries := []fpEntry{
		{category: plugin.CategoryCallable, fp: fp, tokens: tokens},
		{category: plugin.CategoryCallable, fp: fp, tokens: tokens},
	}
	vectors := [][]float32{{1, 0}, {1, 0}}
	opts := &FingerprintOptions{MinSimilarity: 65, MaxSimilarity: 95}

	if matches := rerankWithVectors(entries, nil, vectors, opts); len(matches) != 0 {
		t.Fatalf("weighted similarity outside range produced %d matches", len(matches))
	}
}
