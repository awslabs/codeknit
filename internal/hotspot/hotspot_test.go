// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package hotspot

import (
	"path/filepath"
	"testing"
	"time"

	"codeknit/internal/history"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

func TestAnalyzeRanksVolatileStructuralDependency(t *testing.T) {
	root := t.TempDir()
	files := []string{"api.go", "service.go", "util.go"}
	symbols := make([]plugin.Symbol, 0, len(files))
	for _, file := range files {
		symbols = append(symbols, plugin.Symbol{
			ID: file, Name: file, FilePath: filepath.Join(root, file),
			Category: plugin.CategoryCallable,
		})
	}
	graph := &ir.SymbolGraph{
		Symbols: symbols,
		Edges: []plugin.Edge{
			{From: "api.go", To: "service.go", Kind: plugin.EdgeCalls},
			{From: "util.go", To: "service.go", Kind: plugin.EdgeCalls},
		},
	}
	graph.BuildIndexes()

	now := time.Date(2026, time.July, 21, 0, 0, 0, 0, time.UTC)
	hist := &history.Result{
		RepositoryRoot:  root,
		CommitsAnalyzed: 40,
		Files: map[string]*history.FileMetrics{
			"api.go": {
				Path: "api.go", Commits: 5, Additions: 20,
				RecencyScore: 2, LastChanged: now,
			},
			"service.go": {
				Path: "service.go", Commits: 12, Additions: 100, Deletions: 40,
				RecencyScore: 8, LastChanged: now,
			},
			"util.go": {
				Path: "util.go", Commits: 2, Additions: 5,
				RecencyScore: 1, LastChanged: now,
			},
		},
		CoChanges: map[history.Pair]int{
			history.NewPair("api.go", "service.go"): 4,
		},
	}

	result := Analyze(graph, hist, now.AddDate(-1, 0, 0), now, Options{
		TopN: 10, MinCoChanges: 3,
	})
	if result.Confidence != "medium" {
		t.Fatalf("confidence = %q, want medium", result.Confidence)
	}
	if len(result.Hotspots) != 3 {
		t.Fatalf("hotspots = %d, want 3", len(result.Hotspots))
	}
	if result.Hotspots[0].File != "service.go" {
		t.Fatalf("top hotspot = %q, want service.go", result.Hotspots[0].File)
	}
	if result.Hotspots[0].StructureScore <= result.Hotspots[1].StructureScore {
		t.Fatal("expected service.go to receive structural amplification")
	}
	if len(result.TemporalCoupling) != 1 {
		t.Fatalf("couplings = %d, want 1", len(result.TemporalCoupling))
	}
	if got := result.TemporalCoupling[0].Strength; got != 0.8 {
		t.Fatalf("coupling strength = %v, want 0.8", got)
	}
}

func TestAnalyzeHonorsLimitsAndMinimumCoChanges(t *testing.T) {
	root := t.TempDir()
	graph := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "a", Name: "a", FilePath: filepath.Join(root, "a.go")},
			{ID: "b", Name: "b", FilePath: filepath.Join(root, "b.go")},
		},
	}
	graph.BuildIndexes()
	hist := &history.Result{
		RepositoryRoot: root,
		Files: map[string]*history.FileMetrics{
			"a.go": {Path: "a.go", Commits: 2},
			"b.go": {Path: "b.go", Commits: 2},
		},
		CoChanges: map[history.Pair]int{
			history.NewPair("a.go", "b.go"): 1,
		},
	}

	result := Analyze(graph, hist, time.Time{}, time.Now(), Options{
		TopN: 1, MinCoChanges: 2,
	})
	if len(result.Hotspots) != 1 {
		t.Fatalf("hotspots = %d, want 1", len(result.Hotspots))
	}
	if len(result.TemporalCoupling) != 0 {
		t.Fatalf("couplings = %d, want 0", len(result.TemporalCoupling))
	}
}
