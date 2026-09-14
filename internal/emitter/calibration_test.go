// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"testing"

	"codeknit/internal/fingerprint"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
	"codeknit/internal/plugin/golang"
)

// This small labeled fixture is a regression calibration, not a claim about
// real-world precision/recall. All six pairs are labeled: three clone pairs
// (one exact, two near) and three unrelated pairs.
func TestCloneCalibration(t *testing.T) {
	parser := golang.NewPlugin()
	symbols, _, err := plugin.ParseFile(parser, "testdata/clone_calibration.go", true)
	if err != nil {
		t.Fatal(err)
	}
	graph := &ir.SymbolGraph{Symbols: symbols}
	graph.BuildIndexes()
	labeled := collectFingerprints(graph)
	if len(labeled) != 4 {
		t.Fatalf("expected four labeled declarations, got %d", len(labeled))
	}
	for _, size := range []int{4, 499, 500, 501} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			entries := append([]fpEntry(nil), labeled...)
			for len(entries) < size {
				n := len(entries)
				tokens := []byte{byte((n >> 8) & 0xff), byte(n & 0xff), 1, 2, 3, 4, 5, 6}
				entries = append(entries, fpEntry{
					category: plugin.CategoryValue, tokens: tokens, fp: fingerprint.Hash(tokens),
				})
			}
			matches := findDuplicates(entries, 70, 100)
			predicted := make(map[candidatePair]bool)
			for _, match := range matches {
				if match.i < len(labeled) && match.j < len(labeled) {
					predicted[newCandidatePair(match.i, match.j)] = true
				}
			}
			var tp, fp, fn, tn int
			for i := range labeled {
				for j := i + 1; j < len(labeled); j++ {
					positive := labeled[i].name != "lookup" && labeled[j].name != "lookup"
					found := predicted[newCandidatePair(i, j)]
					switch {
					case positive && found:
						tp++
					case positive:
						fn++
					case found:
						fp++
					default:
						tn++
					}
					if found != positive {
						t.Errorf("%s / %s: clone label=%v, predicted=%v", labeled[i].name, labeled[j].name, positive, found)
					}
				}
			}
			precision := float64(tp) / float64(max(1, tp+fp))
			recall := float64(tp) / float64(max(1, tp+fn))
			t.Logf("labeled fixture at 70%%: TP=%d FP=%d FN=%d TN=%d; precision=%.2f recall=%.2f", tp, fp, fn, tn, precision, recall)
		})
	}
}

func TestExactGroupsUseTokensAndCategory(t *testing.T) {
	entries := exactCloneEntries(4)
	entries[2].tokens = []byte("different token stream")
	entries[3].category = plugin.CategoryType
	groups, ids := groupExactFingerprints(entries)
	if len(groups) != 3 || ids[0] != ids[1] || ids[0] == ids[2] || ids[0] == ids[3] {
		t.Fatalf("fuzzy-hash or category collision created a false exact group: %v", groups)
	}
}

func TestAddingUnrelatedSymbolPreservesExactPairs(t *testing.T) {
	entries := exactCloneEntries(499)
	before := findDuplicates(entries, 100, 100)
	tokens := []byte("unrelated structural shape")
	entries = append(entries, fpEntry{category: plugin.CategoryValue, tokens: tokens, fp: fingerprint.Hash(tokens)})
	after := findDuplicates(entries, 100, 100)
	if len(before) != len(after) {
		t.Fatalf("adding unrelated symbol changed exact pair count from %d to %d", len(before), len(after))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("existing exact pair changed: %+v -> %+v", before[i], after[i])
		}
	}
}
