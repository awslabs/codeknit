// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"fmt"
	"strings"
	"testing"

	"codeknit/internal/plugin"

	"pgregory.net/rapid"
)

// --- Generators ---

var categories = []plugin.SymbolCategory{
	plugin.CategoryCallable,
	plugin.CategoryType,
	plugin.CategoryValue,
	plugin.CategoryModule,
	plugin.CategoryMeta,
}

var kinds = []string{
	"function", "arrow_function", "method", "class", "interface",
	"type_alias", "enum", "variable", "exported_constant", "generic_type_param",
}

var edgeKinds = []plugin.EdgeKind{
	plugin.EdgeCalls, plugin.EdgeInherits, plugin.EdgeContains,
	plugin.EdgeReferences, plugin.EdgeImplements, plugin.EdgeOverrides,
	plugin.EdgeImports, plugin.EdgeDecorates,
}

func genName() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(2, 10).Draw(t, "len")
		chars := make([]byte, n)
		for i := range chars {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}

func genFilePath() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		depth := rapid.IntRange(1, 3).Draw(t, "depth")
		parts := make([]string, depth)
		for i := range parts {
			parts[i] = genName().Draw(t, "dir")
		}
		return strings.Join(parts, "/") + "/" + genName().Draw(t, "file") + ".ts"
	})
}

func genSymbol(filePath string) *rapid.Generator[plugin.Symbol] {
	return rapid.Custom(func(t *rapid.T) plugin.Symbol {
		name := genName().Draw(t, "symName")
		cat := categories[rapid.IntRange(0, len(categories)-1).Draw(t, "cat")]
		kind := kinds[rapid.IntRange(0, len(kinds)-1).Draw(t, "kind")]
		startLine := rapid.IntRange(1, 500).Draw(t, "start")
		endLine := startLine + rapid.IntRange(0, 50).Draw(t, "span")
		sig := name + "(" + genName().Draw(t, "param") + ")"
		props := map[string]string{
			"exported": rapid.StringMatching(`^(true|false)$`).Draw(t, "exp"),
		}
		return plugin.Symbol{
			Name:       name,
			FilePath:   filePath,
			Category:   cat,
			Kind:       kind,
			Signature:  sig,
			Properties: props,
			Span:       [2]int{startLine, endLine},
		}
	})
}

// genFileData generates a map of file paths to symbols and intra-file edges.
// Ensures unique symbol names per file and edges reference valid local names.
func genFileData() *rapid.Generator[struct {
	Symbols map[string][]plugin.Symbol
	Edges   map[string][]plugin.Edge
}] {
	return rapid.Custom(func(t *rapid.T) struct {
		Symbols map[string][]plugin.Symbol
		Edges   map[string][]plugin.Edge
	} {
		numFiles := rapid.IntRange(1, 4).Draw(t, "numFiles")
		fileSymbols := make(map[string][]plugin.Symbol)
		fileEdges := make(map[string][]plugin.Edge)

		usedPaths := make(map[string]bool)
		for i := 0; i < numFiles; i++ {
			var fp string
			for {
				fp = genFilePath().Draw(t, "fp")
				if !usedPaths[fp] {
					usedPaths[fp] = true
					break
				}
			}

			numSyms := rapid.IntRange(1, 5).Draw(t, "numSyms")
			usedNames := make(map[string]bool)
			var syms []plugin.Symbol
			for j := 0; j < numSyms; j++ {
				s := genSymbol(fp).Draw(t, "sym")
				// Ensure unique names within a file.
				for usedNames[s.Name] {
					s.Name += genName().Draw(t, "suffix")
				}
				usedNames[s.Name] = true
				syms = append(syms, s)
			}
			fileSymbols[fp] = syms

			// Generate intra-file edges between symbols in this file.
			if len(syms) >= 2 {
				numEdges := rapid.IntRange(0, len(syms)-1).Draw(t, "numEdges")
				var edges []plugin.Edge
				for j := 0; j < numEdges; j++ {
					fromIdx := rapid.IntRange(0, len(syms)-1).Draw(t, "fromIdx")
					toIdx := rapid.IntRange(0, len(syms)-1).Draw(t, "toIdx")
					if fromIdx == toIdx {
						continue
					}
					ek := edgeKinds[rapid.IntRange(0, len(edgeKinds)-1).Draw(t, "ek")]
					edges = append(edges, plugin.Edge{
						From: syms[fromIdx].Name,
						To:   syms[toIdx].Name,
						Kind: ek,
					})
				}
				fileEdges[fp] = edges
			}
		}

		return struct {
			Symbols map[string][]plugin.Symbol
			Edges   map[string][]plugin.Edge
		}{Symbols: fileSymbols, Edges: fileEdges}
	})
}

// Feature: code-concept-mapper, Property 8: Planner Data Preservation
func TestProperty_PlannerDataPreservation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := genFileData().Draw(t, "data")
		p := &Planner{}
		sg := p.Plan(data.Symbols, data.Edges)

		// Build expected symbol count.
		expectedCount := 0
		for _, syms := range data.Symbols {
			expectedCount += len(syms)
		}
		if len(sg.Symbols) != expectedCount {
			t.Fatalf("expected %d symbols, got %d", expectedCount, len(sg.Symbols))
		}

		// Index output symbols by ID for lookup.
		outByID := make(map[string]plugin.Symbol, len(sg.Symbols))
		for _, s := range sg.Symbols {
			outByID[s.ID] = s
		}

		// Verify every input symbol is preserved with correct global ID and fields.
		// Track name counts per file for disambiguation.
		nameCount := make(map[string]int)
		for fp, syms := range data.Symbols {
			for _, in := range syms {
				key := fp + "::" + in.Name
				count := nameCount[key]
				nameCount[key] = count + 1

				expectedID := fp + "::" + in.Name
				if count > 0 {
					expectedID = fmt.Sprintf("%s::%s#%d", fp, in.Name, count)
				}
				out, ok := outByID[expectedID]
				if !ok {
					t.Fatalf("symbol %q not found in output", expectedID)
				}
				if out.Name != in.Name {
					t.Errorf("symbol %q: Name = %q, want %q", expectedID, out.Name, in.Name)
				}
				if out.FilePath != fp {
					t.Errorf("symbol %q: FilePath = %q, want %q", expectedID, out.FilePath, fp)
				}
				if out.Category != in.Category {
					t.Errorf("symbol %q: Category = %q, want %q", expectedID, out.Category, in.Category)
				}
				if out.Kind != in.Kind {
					t.Errorf("symbol %q: Kind = %q, want %q", expectedID, out.Kind, in.Kind)
				}
				if out.Signature != in.Signature {
					t.Errorf("symbol %q: Signature = %q, want %q", expectedID, out.Signature, in.Signature)
				}
				if out.Span != in.Span {
					t.Errorf("symbol %q: Span = %v, want %v", expectedID, out.Span, in.Span)
				}
				for k, v := range in.Properties {
					if out.Properties[k] != v {
						t.Errorf("symbol %q: Properties[%q] = %q, want %q", expectedID, k, out.Properties[k], v)
					}
				}
			}
		}

		// Verify all intra-file edges are present with remapped global IDs.
		// Build a set of output edges for lookup.
		type edgeKey struct {
			from, to string
			kind     plugin.EdgeKind
		}
		outEdges := make(map[edgeKey]bool, len(sg.Edges))
		for _, e := range sg.Edges {
			outEdges[edgeKey{from: e.From, to: e.To, kind: e.Kind}] = true
		}

		for fp, edges := range data.Edges {
			// Build local name → global ID map for this file (first occurrence).
			localMap := make(map[string]string)
			for _, s := range data.Symbols[fp] {
				if _, exists := localMap[s.Name]; !exists {
					localMap[s.Name] = fp + "::" + s.Name
				}
			}

			for _, e := range edges {
				// Import edges are consumed by the planner for resolution
				// context and are not passed through to the output graph.
				if e.Kind == plugin.EdgeImports {
					continue
				}
				expectedFrom := localMap[e.From]
				expectedTo := localMap[e.To]
				if expectedFrom == "" {
					expectedFrom = fp + "::" + e.From
				}
				if expectedTo == "" {
					expectedTo = fp + "::" + e.To
				}
				key := edgeKey{from: expectedFrom, to: expectedTo, kind: e.Kind}
				if !outEdges[key] {
					t.Errorf("edge %q --%s--> %q not found in output", expectedFrom, e.Kind, expectedTo)
				}
			}
		}

		// Verify all output symbol IDs follow the "filePath::SymbolName" or "filePath::SymbolName#N" format.
		for _, s := range sg.Symbols {
			if !strings.Contains(s.ID, "::") {
				t.Errorf("symbol ID %q does not contain '::'", s.ID)
			}
			parts := strings.SplitN(s.ID, "::", 2)
			if parts[0] != s.FilePath {
				t.Errorf("symbol ID %q: file prefix = %q, want %q", s.ID, parts[0], s.FilePath)
			}
			// The name suffix may have a #N disambiguator.
			namePart := parts[1]
			if idx := strings.LastIndex(namePart, "#"); idx >= 0 {
				namePart = namePart[:idx]
			}
			if namePart != s.Name {
				t.Errorf("symbol ID %q: name suffix = %q, want %q", s.ID, namePart, s.Name)
			}
		}
	})
}

// Feature: code-concept-mapper, Property 9: Calls Edge Inverse Traversal
func TestProperty_CallsEdgeInverseTraversal(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate file data that guarantees some calls edges.
		numFiles := rapid.IntRange(1, 3).Draw(t, "numFiles")
		fileSymbols := make(map[string][]plugin.Symbol)
		fileEdges := make(map[string][]plugin.Edge)

		usedPaths := make(map[string]bool)
		for i := 0; i < numFiles; i++ {
			var fp string
			for {
				fp = genFilePath().Draw(t, "fp")
				if !usedPaths[fp] {
					usedPaths[fp] = true
					break
				}
			}

			numSyms := rapid.IntRange(2, 5).Draw(t, "numSyms")
			usedNames := make(map[string]bool)
			var syms []plugin.Symbol
			for j := 0; j < numSyms; j++ {
				s := genSymbol(fp).Draw(t, "sym")
				for usedNames[s.Name] {
					s.Name += genName().Draw(t, "suffix")
				}
				usedNames[s.Name] = true
				syms = append(syms, s)
			}
			fileSymbols[fp] = syms

			// Generate calls edges between symbols in this file.
			numEdges := rapid.IntRange(1, len(syms)-1).Draw(t, "numEdges")
			var edges []plugin.Edge
			for j := 0; j < numEdges; j++ {
				fromIdx := rapid.IntRange(0, len(syms)-1).Draw(t, "fromIdx")
				toIdx := rapid.IntRange(0, len(syms)-1).Draw(t, "toIdx")
				if fromIdx == toIdx {
					continue
				}
				edges = append(edges, plugin.Edge{
					From: syms[fromIdx].Name,
					To:   syms[toIdx].Name,
					Kind: plugin.EdgeCalls,
				})
			}
			fileEdges[fp] = edges
		}

		p := &Planner{}
		sg := p.Plan(fileSymbols, fileEdges)

		// For every calls edge A→B in the output, verify that filtering
		// edges where To==B and Kind==EdgeCalls includes A.
		for _, e := range sg.Edges {
			if e.Kind != plugin.EdgeCalls {
				continue
			}
			found := false
			for _, candidate := range sg.Edges {
				if candidate.Kind == plugin.EdgeCalls && candidate.To == e.To && candidate.From == e.From {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("calls edge %q --> %q: inverse traversal for target %q does not include caller %q",
					e.From, e.To, e.To, e.From)
			}
		}
	})
}

// TestCrossFileResolution_Proximity verifies that when multiple files define
// a symbol with the same name, the planner resolves cross-file edges to the
// nearest candidate by directory proximity.
func TestCrossFileResolution_Proximity(t *testing.T) {
	// Three files define "Helper". The edge in handler.go calls "Helper".
	// The closest match is src/api/utils.go (same directory), not lib/models/helper.go.
	fileSymbols := map[string][]plugin.Symbol{
		"src/api/handler.go": {
			{
				Name: "Handle", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Handle()", Span: [2]int{1, 5},
			},
		},
		"src/api/utils.go": {
			{
				Name: "Helper", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Helper()", Span: [2]int{1, 3},
			},
		},
		"lib/models/helper.go": {
			{
				Name: "Helper", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Helper()", Span: [2]int{1, 3},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/api/handler.go": {
			{From: "Handle", To: "Helper", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	if edge.To != "src/api/utils.go::Helper" {
		t.Errorf("expected edge to resolve to src/api/utils.go::Helper, got %q", edge.To)
	}
}

// TestCrossFileResolution_SingleCandidate verifies that when only one file
// defines a symbol, cross-file resolution picks it regardless of distance.
func TestCrossFileResolution_SingleCandidate(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"src/api/handler.go": {
			{
				Name: "Handle", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Handle()", Span: [2]int{1, 5},
			},
		},
		"lib/deep/nested/utils.go": {
			{
				Name: "Format", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Format()", Span: [2]int{1, 3},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/api/handler.go": {
			{From: "Handle", To: "Format", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	if edge.To != "lib/deep/nested/utils.go::Format" {
		t.Errorf("expected edge to resolve to lib/deep/nested/utils.go::Format, got %q", edge.To)
	}
}

// TestCrossFileResolution_DeterministicTiebreak verifies that when multiple
// candidates have equal directory distance, the lexicographically smallest
// global ID wins — making the result deterministic and unique.
func TestCrossFileResolution_DeterministicTiebreak(t *testing.T) {
	// Two files at equal distance both define "Logger".
	// pkg/alpha/logger.go and pkg/beta/logger.go are equidistant from src/main.go.
	// The lexicographically smaller global ID (pkg/alpha/...) must win.
	fileSymbols := map[string][]plugin.Symbol{
		"src/main.go": {
			{
				Name: "Run", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Run()", Span: [2]int{1, 10},
			},
		},
		"pkg/beta/logger.go": {
			{
				Name: "Logger", Category: plugin.CategoryType, Kind: "struct",
				Signature: "Logger", Span: [2]int{1, 5},
			},
		},
		"pkg/alpha/logger.go": {
			{
				Name: "Logger", Category: plugin.CategoryType, Kind: "struct",
				Signature: "Logger", Span: [2]int{1, 5},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/main.go": {
			{From: "Run", To: "Logger", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	if edge.To != "pkg/alpha/logger.go::Logger" {
		t.Errorf("expected deterministic tiebreak to pkg/alpha/logger.go::Logger, got %q", edge.To)
	}

	// Run again to confirm determinism.
	sg2 := p.Plan(fileSymbols, fileEdges)
	edge2 := sg2.Edges[0]
	if edge.To != edge2.To {
		t.Errorf("non-deterministic: first run resolved to %q, second to %q", edge.To, edge2.To)
	}
}

// TestCrossFileResolution_Unresolved verifies that references to symbols
// that don't exist anywhere produce a synthetic "filePath::name" ID
// rather than panicking or silently dropping the edge.
func TestCrossFileResolution_Unresolved(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"src/app.go": {
			{
				Name: "Start", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Start()", Span: [2]int{1, 5},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/app.go": {
			{From: "Start", To: "NonExistent", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	if edge.From != "src/app.go::Start" {
		t.Errorf("expected From = src/app.go::Start, got %q", edge.From)
	}
	if edge.To != "src/app.go::NonExistent" {
		t.Errorf("expected unresolved To = src/app.go::NonExistent, got %q", edge.To)
	}
}

// TestProperty_ResolveIDDeterminism is a property test that verifies resolveID
// always returns the same result for the same inputs, regardless of call order.
func TestProperty_ResolveIDDeterminism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate 2-4 files, each with 1-3 symbols. Some symbols share names.
		numFiles := rapid.IntRange(2, 4).Draw(t, "numFiles")
		sharedName := genName().Draw(t, "sharedName")

		fileSymbols := make(map[string][]plugin.Symbol)
		usedPaths := make(map[string]bool)

		for range numFiles {
			var fp string
			for {
				fp = genFilePath().Draw(t, "fp")
				if !usedPaths[fp] {
					usedPaths[fp] = true
					break
				}
			}
			// Every file gets the shared name plus a unique symbol.
			fileSymbols[fp] = []plugin.Symbol{
				{
					Name: sharedName, Category: plugin.CategoryCallable, Kind: "function",
					Signature: sharedName + "()", Span: [2]int{1, 5},
				},
				{
					Name: genName().Draw(t, "unique"), Category: plugin.CategoryCallable, Kind: "function",
					Signature: "unique()", Span: [2]int{10, 15},
				},
			}
		}

		// Pick a random referencing file.
		refFile := genFilePath().Draw(t, "refFile")
		fileSymbols[refFile] = []plugin.Symbol{
			{
				Name: "Caller", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Caller()", Span: [2]int{1, 3},
			},
		}
		fileEdges := map[string][]plugin.Edge{
			refFile: {
				{From: "Caller", To: sharedName, Kind: plugin.EdgeCalls},
			},
		}

		// Run Plan twice and verify identical results.
		p := &Planner{}
		sg1 := p.Plan(fileSymbols, fileEdges)
		sg2 := p.Plan(fileSymbols, fileEdges)

		if len(sg1.Edges) != len(sg2.Edges) {
			t.Fatalf("edge count differs: %d vs %d", len(sg1.Edges), len(sg2.Edges))
		}
		for i := range sg1.Edges {
			if sg1.Edges[i].From != sg2.Edges[i].From {
				t.Errorf("edge[%d].From differs: %q vs %q", i, sg1.Edges[i].From, sg2.Edges[i].From)
			}
			if sg1.Edges[i].To != sg2.Edges[i].To {
				t.Errorf("edge[%d].To differs: %q vs %q", i, sg1.Edges[i].To, sg2.Edges[i].To)
			}
		}

		// Verify the resolved To is a valid global ID (contains "::").
		for _, e := range sg1.Edges {
			if !strings.Contains(e.To, "::") {
				t.Errorf("resolved edge To %q does not contain '::'", e.To)
			}
			if !strings.Contains(e.From, "::") {
				t.Errorf("resolved edge From %q does not contain '::'", e.From)
			}
		}
	})
}

// TestCrossFileResolution_ImportAware verifies that when a file has an import
// edge for a name, the planner prefers candidates from the imported module
// over proximity-based guessing.
func TestCrossFileResolution_ImportAware(t *testing.T) {
	// Two files define "User": one in models/ (far) and one in utils/ (near).
	// The handler file imports "User" from "models", so it should resolve
	// to models/user.go::User despite utils/user.go being closer.
	fileSymbols := map[string][]plugin.Symbol{
		"src/api/handler.go": {
			{
				Name: "Handle", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Handle()", Span: [2]int{1, 10},
			},
		},
		"src/api/utils/user.go": {
			{
				Name: "User", Category: plugin.CategoryType, Kind: "struct",
				Signature: "User", Span: [2]int{1, 5},
			},
		},
		"src/models/user.go": {
			{
				Name: "User", Category: plugin.CategoryType, Kind: "struct",
				Signature: "User", Span: [2]int{1, 5},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/api/handler.go": {
			// Import edge: "User" comes from "models"
			{From: "User", To: "models", Kind: plugin.EdgeImports},
			// Call edge using the imported name
			{From: "Handle", To: "User", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	// Should have exactly 1 edge (the calls edge; import edge is consumed).
	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	// Should resolve to models/user.go::User (imported), not utils/user.go::User (closer).
	if edge.To != "src/models/user.go::User" {
		t.Errorf("expected import-aware resolution to src/models/user.go::User, got %q", edge.To)
	}
}

// TestCrossFileResolution_ImportFallback verifies that when an import doesn't
// match any candidate, resolution falls back to proximity.
func TestCrossFileResolution_ImportFallback(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"src/app.go": {
			{
				Name: "Start", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Start()", Span: [2]int{1, 5},
			},
		},
		"src/helpers/format.go": {
			{
				Name: "Format", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Format()", Span: [2]int{1, 3},
			},
		},
		"lib/format.go": {
			{
				Name: "Format", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Format()", Span: [2]int{1, 3},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"src/app.go": {
			// Import points to a module that doesn't match any candidate path.
			{From: "Format", To: "nonexistent.module", Kind: plugin.EdgeImports},
			{From: "Start", To: "Format", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	if len(sg.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(sg.Edges))
	}
	edge := sg.Edges[0]
	// Should fall back to proximity: src/helpers/format.go is closer to src/app.go.
	if edge.To != "src/helpers/format.go::Format" {
		t.Errorf("expected proximity fallback to src/helpers/format.go::Format, got %q", edge.To)
	}
}

// TestCrossFileResolution_RelativeImportDisambiguation verifies that when two
// different folders contain a module with the same name, relative import paths
// are resolved against the importing file's directory so that each file links
// to the correct ModuleA in its own folder.
//
// Scenario:
//
//	folder1/ModuleA.ts  → exports Helper
//	folder1/ModuleB.ts  → import { Helper } from './ModuleA'
//	folder2/ModuleA.ts  → exports Helper
//	folder2/ModuleC.ts  → import { Helper } from './ModuleA'
//
// ModuleB should link to folder1/ModuleA::Helper, NOT folder2/ModuleA::Helper.
// ModuleC should link to folder2/ModuleA::Helper, NOT folder1/ModuleA::Helper.
func TestCrossFileResolution_RelativeImportDisambiguation(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"folder1/ModuleA.ts": {
			{
				Name: "Helper", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Helper()", Span: [2]int{1, 3},
			},
		},
		"folder1/ModuleB.ts": {
			{
				Name: "main", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "main()", Span: [2]int{1, 5},
			},
		},
		"folder2/ModuleA.ts": {
			{
				Name: "Helper", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "Helper()", Span: [2]int{1, 3},
			},
		},
		"folder2/ModuleC.ts": {
			{
				Name: "run", Category: plugin.CategoryCallable, Kind: "function",
				Signature: "run()", Span: [2]int{1, 5},
			},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"folder1/ModuleB.ts": {
			{From: "Helper", To: "./ModuleA", Kind: plugin.EdgeImports},
			{From: "main", To: "Helper", Kind: plugin.EdgeCalls},
		},
		"folder2/ModuleC.ts": {
			{From: "Helper", To: "./ModuleA", Kind: plugin.EdgeImports},
			{From: "run", To: "Helper", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	// Collect calls edges.
	type callEdge struct{ from, to string }
	var calls []callEdge
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			calls = append(calls, callEdge{e.From, e.To})
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls edges, got %d: %v", len(calls), calls)
	}

	// ModuleB's main() should call folder1/ModuleA::Helper
	found := false
	for _, c := range calls {
		if c.from == "folder1/ModuleB.ts::main" && c.to == "folder1/ModuleA.ts::Helper" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected folder1/ModuleB.ts::main -> folder1/ModuleA.ts::Helper, got %v", calls)
	}

	// ModuleC's run() should call folder2/ModuleA::Helper
	found = false
	for _, c := range calls {
		if c.from == "folder2/ModuleC.ts::run" && c.to == "folder2/ModuleA.ts::Helper" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected folder2/ModuleC.ts::run -> folder2/ModuleA.ts::Helper, got %v", calls)
	}
}

// TestCrossFileResolution_PythonRelativeImport verifies that Python-style
// relative imports (leading dots) resolve correctly when two packages share
// the same module name.
//
// Scenario:
//
//	pkg1/models.py  → exports User
//	pkg1/app.py     → from .models import User
//	pkg2/models.py  → exports User
//	pkg2/app.py     → from .models import User
func TestCrossFileResolution_PythonRelativeImport(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"pkg1/models.py": {
			{Name: "User", Category: plugin.CategoryType, Kind: "class", Signature: "User", Span: [2]int{1, 5}},
		},
		"pkg1/app.py": {
			{Name: "handler", Category: plugin.CategoryCallable, Kind: "function", Signature: "handler()", Span: [2]int{1, 5}},
		},
		"pkg2/models.py": {
			{Name: "User", Category: plugin.CategoryType, Kind: "class", Signature: "User", Span: [2]int{1, 5}},
		},
		"pkg2/app.py": {
			{Name: "handler", Category: plugin.CategoryCallable, Kind: "function", Signature: "handler()", Span: [2]int{1, 5}},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"pkg1/app.py": {
			{From: "User", To: ".models", Kind: plugin.EdgeImports},
			{From: "handler", To: "User", Kind: plugin.EdgeCalls},
		},
		"pkg2/app.py": {
			{From: "User", To: ".models", Kind: plugin.EdgeImports},
			{From: "handler", To: "User", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	type callEdge struct{ from, to string }
	var calls []callEdge
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			calls = append(calls, callEdge{e.From, e.To})
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls edges, got %d: %v", len(calls), calls)
	}

	for _, c := range calls {
		if c.from == "pkg1/app.py::handler" && c.to != "pkg1/models.py::User" {
			t.Errorf("pkg1/app.py::handler should call pkg1/models.py::User, got %q", c.to)
		}
		if c.from == "pkg2/app.py::handler" && c.to != "pkg2/models.py::User" {
			t.Errorf("pkg2/app.py::handler should call pkg2/models.py::User, got %q", c.to)
		}
	}
}

// TestCrossFileResolution_BareImportDirectoryScoped verifies that bare
// (non-relative) imports like Python's `from models import X` prefer
// candidates in the same directory tree over distant ones.
//
// Scenario:
//
//	project1/models/user.py  → exports User
//	project1/app.py          → from models import User
//	project2/models/user.py  → exports User
//	project2/app.py          → from models import User
func TestCrossFileResolution_BareImportDirectoryScoped(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"project1/models/user.py": {
			{Name: "User", Category: plugin.CategoryType, Kind: "class", Signature: "User", Span: [2]int{1, 5}},
		},
		"project1/app.py": {
			{Name: "handler", Category: plugin.CategoryCallable, Kind: "function", Signature: "handler()", Span: [2]int{1, 5}},
		},
		"project2/models/user.py": {
			{Name: "User", Category: plugin.CategoryType, Kind: "class", Signature: "User", Span: [2]int{1, 5}},
		},
		"project2/app.py": {
			{Name: "handler", Category: plugin.CategoryCallable, Kind: "function", Signature: "handler()", Span: [2]int{1, 5}},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"project1/app.py": {
			{From: "User", To: "models", Kind: plugin.EdgeImports},
			{From: "handler", To: "User", Kind: plugin.EdgeCalls},
		},
		"project2/app.py": {
			{From: "User", To: "models", Kind: plugin.EdgeImports},
			{From: "handler", To: "User", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	type callEdge struct{ from, to string }
	var calls []callEdge
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			calls = append(calls, callEdge{e.From, e.To})
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls edges, got %d: %v", len(calls), calls)
	}

	for _, c := range calls {
		if c.from == "project1/app.py::handler" && c.to != "project1/models/user.py::User" {
			t.Errorf("project1 handler should call project1 User, got %q", c.to)
		}
		if c.from == "project2/app.py::handler" && c.to != "project2/models/user.py::User" {
			t.Errorf("project2 handler should call project2 User, got %q", c.to)
		}
	}
}

// TestCrossFileResolution_RubyRequireRelative verifies that Ruby's
// require_relative paths (implicitly relative, no ./ prefix) resolve
// correctly when two directories have identically-named files.
func TestCrossFileResolution_RubyRequireRelative(t *testing.T) {
	fileSymbols := map[string][]plugin.Symbol{
		"lib/a/helper.rb": {
			{Name: "run", Category: plugin.CategoryCallable, Kind: "function", Signature: "run()", Span: [2]int{1, 3}},
		},
		"lib/a/main.rb": {
			{Name: "start", Category: plugin.CategoryCallable, Kind: "function", Signature: "start()", Span: [2]int{1, 5}},
		},
		"lib/b/helper.rb": {
			{Name: "run", Category: plugin.CategoryCallable, Kind: "function", Signature: "run()", Span: [2]int{1, 3}},
		},
		"lib/b/main.rb": {
			{Name: "start", Category: plugin.CategoryCallable, Kind: "function", Signature: "start()", Span: [2]int{1, 5}},
		},
	}
	fileEdges := map[string][]plugin.Edge{
		"lib/a/main.rb": {
			// require_relative "helper" → To is "helper" (bare, no ./)
			{From: "run", To: "helper", Kind: plugin.EdgeImports},
			{From: "start", To: "run", Kind: plugin.EdgeCalls},
		},
		"lib/b/main.rb": {
			{From: "run", To: "helper", Kind: plugin.EdgeImports},
			{From: "start", To: "run", Kind: plugin.EdgeCalls},
		},
	}

	p := &Planner{}
	sg := p.Plan(fileSymbols, fileEdges)

	type callEdge struct{ from, to string }
	var calls []callEdge
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			calls = append(calls, callEdge{e.From, e.To})
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls edges, got %d: %v", len(calls), calls)
	}

	for _, c := range calls {
		if c.from == "lib/a/main.rb::start" && c.to != "lib/a/helper.rb::run" {
			t.Errorf("lib/a start should call lib/a run, got %q", c.to)
		}
		if c.from == "lib/b/main.rb::start" && c.to != "lib/b/helper.rb::run" {
			t.Errorf("lib/b start should call lib/b run, got %q", c.to)
		}
	}
}
