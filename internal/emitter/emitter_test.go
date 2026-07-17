// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/config"
	"codeknit/internal/emitter/parser"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"

	"pgregory.net/rapid"
)

// emitInlineString is a test helper that emits to a buffer and returns the full output.
func emitInlineString(t interface{ Fatalf(string, ...any) }, sg *ir.SymbolGraph, opts *EmitOptions) string {
	em := &Emitter{}
	var buf bytes.Buffer
	if err := em.EmitInline(&buf, sg, opts); err != nil {
		t.Fatalf("EmitInline: %v", err)
	}
	return buf.String()
}

func TestDefaultAnalysisOptionsMatchConfig(t *testing.T) {
	opts := DefaultAnalysisOptions()
	if opts.OutputPath != config.DefaultAnalyzeOutput {
		t.Errorf("OutputPath default: got %q, want %q", opts.OutputPath, config.DefaultAnalyzeOutput)
	}
	if opts.FanThreshold != config.DefaultAnalyzeFanThreshold {
		t.Errorf("FanThreshold default: got %d, want %d", opts.FanThreshold, config.DefaultAnalyzeFanThreshold)
	}
	if opts.GodThreshold != config.DefaultAnalyzeGodThreshold {
		t.Errorf("GodThreshold default: got %d, want %d", opts.GodThreshold, config.DefaultAnalyzeGodThreshold)
	}
	if opts.MaxInheritanceDepth != config.DefaultAnalyzeMaxInheritanceDepth {
		t.Errorf("MaxInheritanceDepth default: got %d, want %d", opts.MaxInheritanceDepth, config.DefaultAnalyzeMaxInheritanceDepth)
	}
	if opts.TopN != config.DefaultAnalyzeTopN {
		t.Errorf("TopN default: got %d, want %d", opts.TopN, config.DefaultAnalyzeTopN)
	}
	if opts.BetweennessThreshold != config.DefaultAnalyzeBetweennessThreshold {
		t.Errorf("BetweennessThreshold default: got %g, want %g", opts.BetweennessThreshold, config.DefaultAnalyzeBetweennessThreshold)
	}
	if opts.PropagationCutoff != config.DefaultAnalyzePropagationCutoff {
		t.Errorf("PropagationCutoff default: got %g, want %g", opts.PropagationCutoff, config.DefaultAnalyzePropagationCutoff)
	}
}

// emitFlatFiles is a test helper that emits in directory-flat mode to a temp dir
// and returns the contents of each map_NNN.skt file in order.
func emitFlatFiles(t interface {
	Fatalf(string, ...any)
}, sg *ir.SymbolGraph, opts *EmitOptions,
) []string {
	dir, err := os.MkdirTemp("", "emitter-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir) //nolint:errcheck // test cleanup

	optsCopy := *opts
	optsCopy.OutputDir = dir
	optsCopy.OutputMode = config.OutputDirectoryFlat

	em := &Emitter{}
	written, emitErr := em.Emit(sg, &optsCopy)
	if emitErr != nil {
		t.Fatalf("Emit: %v", emitErr)
	}

	// Read map files in order.
	var files []string
	for _, path := range written {
		if !strings.Contains(filepath.Base(path), "map_") {
			continue
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		files = append(files, string(data))
	}
	return files
}

// genSymbolCategory generates a random SymbolCategory.
func genSymbolCategory() *rapid.Generator[plugin.SymbolCategory] {
	return rapid.SampledFrom([]plugin.SymbolCategory{
		plugin.CategoryCallable,
		plugin.CategoryType,
		plugin.CategoryValue,
		plugin.CategoryModule,
		plugin.CategoryMeta,
	})
}

// genEdgeKind generates a random EdgeKind.
func genEdgeKind() *rapid.Generator[plugin.EdgeKind] {
	return rapid.SampledFrom([]plugin.EdgeKind{
		plugin.EdgeCalls,
		plugin.EdgeInherits,
		plugin.EdgeContains,
		plugin.EdgeReferences,
		plugin.EdgeImplements,
		plugin.EdgeOverrides,
		plugin.EdgeImports,
		plugin.EdgeDecorates,
	})
}

// genSymbol generates a random Symbol with a given file path and unique name.
func genSymbol(filePath, name string, t *rapid.T) plugin.Symbol {
	cat := genSymbolCategory().Draw(t, "category")
	kinds := map[plugin.SymbolCategory][]string{
		plugin.CategoryCallable: {"function", "arrow_function", "method"},
		plugin.CategoryType:     {"class", "interface", "type_alias", "enum"},
		plugin.CategoryValue:    {"variable", "exported_constant"},
		plugin.CategoryModule:   {"module"},
		plugin.CategoryMeta:     {"generic_type_param"},
	}
	kind := rapid.SampledFrom(kinds[cat]).Draw(t, "kind")
	startLine := rapid.IntRange(1, 500).Draw(t, "startLine")
	endLine := rapid.IntRange(startLine, startLine+100).Draw(t, "endLine")

	props := make(map[string]string)
	if rapid.Bool().Draw(t, "hasAsync") {
		props["async"] = "true"
	}
	if rapid.Bool().Draw(t, "hasExported") {
		props["exported"] = "true"
	}

	sig := rapid.StringMatching(`[a-zA-Z_][a-zA-Z0-9_]*\([a-z]*\)`).Draw(t, "signature")

	return plugin.Symbol{
		ID:         filePath + "::" + name,
		Name:       name,
		FilePath:   filePath,
		Category:   cat,
		Kind:       kind,
		Signature:  sig,
		Properties: props,
		Span:       [2]int{startLine, endLine},
	}
}

// genSymbolGraph generates a random SymbolGraph with 1-50 symbols across 1-5 files.
func genSymbolGraph(t *rapid.T) *ir.SymbolGraph {
	numFiles := rapid.IntRange(1, 5).Draw(t, "numFiles")
	sg := &ir.SymbolGraph{}

	var allIDs []string
	for f := 0; f < numFiles; f++ {
		filePath := rapid.StringMatching(`src/[a-z]{1,8}/[a-z]{1,8}\.ts`).Draw(t, "filePath")
		numSyms := rapid.IntRange(1, 10).Draw(t, "numSymbols")
		usedNames := make(map[string]bool)
		for s := 0; s < numSyms; s++ {
			var name string
			for {
				name = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,15}`).Draw(t, "name")
				if !usedNames[name] {
					usedNames[name] = true
					break
				}
			}
			sym := genSymbol(filePath, name, t)
			sg.Symbols = append(sg.Symbols, sym)
			allIDs = append(allIDs, sym.ID)
		}
	}

	// Generate some edges between existing symbols.
	if len(allIDs) > 1 {
		numEdges := rapid.IntRange(0, len(allIDs)*2).Draw(t, "numEdges")
		for e := 0; e < numEdges; e++ {
			fromIdx := rapid.IntRange(0, len(allIDs)-1).Draw(t, "fromIdx")
			toIdx := rapid.IntRange(0, len(allIDs)-1).Draw(t, "toIdx")
			if fromIdx == toIdx {
				continue
			}
			sg.Edges = append(sg.Edges, plugin.Edge{
				From: allIDs[fromIdx],
				To:   allIDs[toIdx],
				Kind: genEdgeKind().Draw(t, "edgeKind"),
			})
		}
	}

	sg.BuildIndexes()
	return sg
}

func TestEmitJSONInline(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{
				ID:         "src/app.go::User",
				Name:       "User",
				FilePath:   "src/app.go",
				Category:   plugin.CategoryType,
				Kind:       "struct",
				Signature:  "type User struct",
				Properties: map[string]string{"exported": "true"},
				Span:       [2]int{1, 3},
			},
			{
				ID:        "src/app.go::Save",
				Name:      "Save",
				FilePath:  "src/app.go",
				Category:  plugin.CategoryCallable,
				Kind:      "function",
				Signature: "Save(u: User)",
				Span:      [2]int{5, 8},
			},
		},
		Edges: []plugin.Edge{
			{From: "src/app.go::Save", To: "src/app.go::User", Kind: plugin.EdgeReferences},
		},
	}
	sg.BuildIndexes()

	var buf bytes.Buffer
	if err := (&Emitter{}).EmitJSON(&buf, sg); err != nil {
		t.Fatalf("EmitJSON: %v", err)
	}

	var out jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal json: %v\n%s", err, buf.String())
	}
	if len(out.Files) != 1 || out.Files[0] != "src/app.go" {
		t.Fatalf("files = %v, want [src/app.go]", out.Files)
	}
	if len(out.Symbols) != 2 {
		t.Fatalf("symbols len = %d, want 2", len(out.Symbols))
	}
	if out.Symbols[1].Signature != "Save(u: S1)" {
		t.Fatalf("signature = %q, want type ref resolved to short ID", out.Symbols[1].Signature)
	}
	if len(out.Edges) != 1 || out.Edges[0].FromShort != "S2" || out.Edges[0].ToShort != "S1" {
		t.Fatalf("edges = %+v, want resolved short IDs", out.Edges)
	}
}

func TestEmitJSONDirectoryWritesSingleFile(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{{
			ID:        "src/app.go::main",
			Name:      "main",
			FilePath:  "src/app.go",
			Category:  plugin.CategoryCallable,
			Kind:      "function",
			Signature: "main()",
			Span:      [2]int{1, 1},
		}},
	}
	sg.BuildIndexes()

	dir := t.TempDir()
	written, err := (&Emitter{}).Emit(sg, &EmitOptions{
		OutputDir:    dir,
		OutputMode:   config.OutputDirectoryFlat,
		OutputFormat: config.OutputFormatJSON,
		MaxLines:     500,
	})
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if len(written) != 1 || filepath.Base(written[0]) != "codeknit.json" {
		t.Fatalf("written = %v, want single codeknit.json", written)
	}
	if _, err := os.Stat(filepath.Join(dir, "codeknit.json")); err != nil {
		t.Fatalf("codeknit.json not written: %v", err)
	}
}

func TestEmitGraphIsSelfContained(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{
				ID:        "src/app.go::main",
				Name:      "main",
				FilePath:  "src/app.go",
				Category:  plugin.CategoryCallable,
				Kind:      "function",
				Signature: "main()",
				Span:      [2]int{1, 3},
			},
			{
				ID:        "src/app.go::Server",
				Name:      "Server",
				FilePath:  "src/app.go",
				Category:  plugin.CategoryType,
				Kind:      "struct",
				Signature: "type Server struct",
				Span:      [2]int{5, 8},
			},
		},
		Edges: []plugin.Edge{
			{From: "src/app.go::main", To: "src/app.go::Server", Kind: plugin.EdgeReferences},
		},
	}
	sg.BuildIndexes()

	outputPath := filepath.Join(t.TempDir(), "codeknit-graph.html")
	if err := (&Emitter{}).EmitGraph(sg, outputPath); err != nil {
		t.Fatalf("EmitGraph: %v", err)
	}

	htmlBytes, err := os.ReadFile(outputPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("read graph HTML: %v", err)
	}
	html := string(htmlBytes)
	for _, disallowed := range []string{`src="https://d3js.org`, `src="http://d3js.org`, "/*D3_JS*/", "/*GRAPH_DATA_JSON*/"} {
		if strings.Contains(html, disallowed) {
			t.Fatalf("graph HTML contains %q; output should be self-contained", disallowed)
		}
	}
	if !strings.Contains(html, "d3.zoomIdentity") {
		t.Fatal("graph HTML does not include embedded D3")
	}
	if !strings.Contains(html, `"shortId": "S1"`) || !strings.Contains(html, `"kind": "references"`) {
		t.Fatal("graph HTML does not include generated graph data")
	}
}

// Feature: code-concept-mapper, Property 11: Dictionary Uniqueness
func TestProperty11_DictionaryUniqueness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genSymbolGraph(t)
		dict := NewDictionary(sg)

		// All codes must be unique.
		codesSeen := make(map[string]string)
		for token, code := range dict.Forward {
			if prev, dup := codesSeen[code]; dup {
				t.Fatalf("Code %q assigned to both %q and %q", code, prev, token)
			}
			codesSeen[code] = token
		}

		// Forward and Reverse must be consistent.
		for token, code := range dict.Forward {
			if dict.Reverse[code] != token {
				t.Fatalf("Reverse[%q] = %q, want %q", code, dict.Reverse[code], token)
			}
		}
		for code, token := range dict.Reverse {
			if dict.Forward[token] != code {
				t.Fatalf("Forward[%q] = %q, want %q", token, dict.Forward[token], code)
			}
		}
	})
}

// Feature: code-concept-mapper, Property 2: Emitter Determinism
func TestProperty2_EmitterDeterminism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genSymbolGraph(t)
		opts := &EmitOptions{MaxLines: 500}

		out1 := emitInlineString(t, sg, opts)
		out2 := emitInlineString(t, sg, opts)

		if out1 != out2 {
			t.Fatal("inline output differs between two identical runs")
		}
	})
}

// Feature: code-concept-mapper, Property 2: Emitter Determinism (minified)
func TestProperty2_EmitterDeterminism_Minified(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genSymbolGraph(t)
		opts := &EmitOptions{MaxLines: 500, Minify: true}

		out1 := emitInlineString(t, sg, opts)
		out2 := emitInlineString(t, sg, opts)

		if out1 != out2 {
			t.Fatal("minified inline output differs between two identical runs")
		}
	})
}

// Feature: code-concept-mapper, Property 12: Minified Output Dictionary Consistency
func TestProperty12_MinifiedOutputDictConsistency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genSymbolGraph(t)

		dir, err := os.MkdirTemp("", "emitter-dict-test-*")
		if err != nil {
			t.Fatalf("MkdirTemp: %v", err)
		}
		defer os.RemoveAll(dir) //nolint:errcheck // test cleanup

		opts := &EmitOptions{
			OutputDir:  dir,
			OutputMode: config.OutputDirectoryFlat,
			MaxLines:   500,
			Minify:     true,
		}

		em := &Emitter{}
		written, emitErr := em.Emit(sg, opts)
		if emitErr != nil {
			t.Fatalf("Emit: %v", emitErr)
		}
		if len(written) == 0 {
			t.Fatal("no output files")
		}

		// (a) dict.skt must exist and contain [dict] section.
		dictPath := filepath.Join(dir, "dict.skt")
		dictData, readErr := os.ReadFile(dictPath) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("dict.skt not found: %v", readErr)
		}
		dictContent := string(dictData)
		if !strings.Contains(dictContent, "[dict]") {
			t.Fatal("dict.skt missing [dict] section")
		}

		// (b) Parse dict mappings and verify uniqueness.
		dictMap := make(map[string]string) // code → token
		lines := strings.Split(dictContent, "\n")
		inDict := false
		for _, line := range lines {
			if line == "[dict]" {
				inDict = true
				continue
			}
			if strings.HasPrefix(line, "[") && line != "[dict]" {
				inDict = false
				continue
			}
			if inDict && strings.HasPrefix(line, "- ") {
				parts := strings.SplitN(line[2:], ": ", 2)
				if len(parts) == 2 {
					dictMap[parts[0]] = parts[1]
				}
			}
		}

		codesSeen := make(map[string]bool)
		for code := range dictMap {
			if codesSeen[code] {
				t.Fatalf("duplicate dict code: %s", code)
			}
			codesSeen[code] = true
		}

		// (c) Map files must NOT contain [dict] — it lives in dict.skt.
		for _, path := range written {
			base := filepath.Base(path)
			if !strings.HasPrefix(base, "map_") {
				continue
			}
			data, readErr := os.ReadFile(path) //nolint:gosec // test file
			if readErr != nil {
				t.Fatalf("read %s: %v", path, readErr)
			}
			if strings.Contains(string(data), "[dict]") {
				t.Fatalf("%s should not contain [dict] section", base)
			}
		}
	})
}

// genLargeSymbolGraph generates a SymbolGraph large enough to produce >500 lines of output.
func genLargeSymbolGraph(t *rapid.T) *ir.SymbolGraph {
	sg := &ir.SymbolGraph{}
	numFiles := rapid.IntRange(10, 30).Draw(t, "numFiles")
	var allIDs []string

	for f := 0; f < numFiles; f++ {
		filePath := fmt.Sprintf("src/pkg%d/file%d.ts", f/5, f)
		numSyms := rapid.IntRange(5, 20).Draw(t, "numSymbols")
		for s := 0; s < numSyms; s++ {
			name := fmt.Sprintf("sym_%d_%d", f, s)
			sym := plugin.Symbol{
				ID:         filePath + "::" + name,
				Name:       name,
				FilePath:   filePath,
				Category:   plugin.CategoryCallable,
				Kind:       "function",
				Signature:  fmt.Sprintf("%s(x: number) -> void", name),
				Properties: map[string]string{"exported": "true"},
				Span:       [2]int{s*10 + 1, s*10 + 9},
			}
			sg.Symbols = append(sg.Symbols, sym)
			allIDs = append(allIDs, sym.ID)
		}
	}

	// Add edges.
	for i := 0; i+1 < len(allIDs); i += 2 {
		sg.Edges = append(sg.Edges, plugin.Edge{
			From: allIDs[i],
			To:   allIDs[i+1],
			Kind: plugin.EdgeCalls,
		})
	}

	sg.BuildIndexes()
	return sg
}

// Feature: code-concept-mapper, Property 13: Output File Splitting Constraints
func TestProperty13_OutputFileSplittingConstraints(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genLargeSymbolGraph(t)
		opts := &EmitOptions{MaxLines: 500}

		outputs := emitFlatFiles(t, sg, opts)

		// Verify each file respects the ~500 line limit (allow some slack for edge sections).
		for i, content := range outputs {
			lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			// Allow 10% slack over MaxLines since we don't split file blocks.
			if len(lines) > opts.MaxLines+opts.MaxLines/5 {
				t.Fatalf("output file %d has %d lines (max ~%d)", i+1, len(lines), opts.MaxLines)
			}
		}

		// Verify no file's symbol block is split across two output files.
		// A file's symbol block starts with "## filepath" and includes all subsequent
		// symbol lines until the next "##" or section header.
		for _, content := range outputs {
			lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			fileHeaders := make(map[string]bool)
			for _, line := range lines {
				if strings.HasPrefix(line, "## ") {
					fp := line[3:]
					if fileHeaders[fp] {
						t.Fatalf("file %q appears twice in the same output file", fp)
					}
					fileHeaders[fp] = true
				}
			}
		}

		// Verify a file path doesn't appear in multiple output files.
		globalFileHeaders := make(map[string]int)
		for i, content := range outputs {
			lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "## ") {
					fp := line[3:]
					if prev, exists := globalFileHeaders[fp]; exists {
						t.Fatalf("file %q appears in output files %d and %d", fp, prev+1, i+1)
					}
					globalFileHeaders[fp] = i
				}
			}
		}

		// Verify sequential naming would work (just check count > 0).
		if len(outputs) == 0 {
			t.Fatal("no output files generated")
		}
	})
}

// genRoundTripSymbolGraph generates a SymbolGraph suitable for round-trip testing.
// All symbol IDs follow the "filePath::Name" format and names are unique within each file.
// Signatures and names contain only safe characters for parsing.
func genRoundTripSymbolGraph(t *rapid.T) *ir.SymbolGraph {
	numFiles := rapid.IntRange(1, 5).Draw(t, "numFiles")
	sg := &ir.SymbolGraph{}
	var allIDs []string

	for f := 0; f < numFiles; f++ {
		filePath := fmt.Sprintf("src/pkg%d/file%d.ts", f/3, f)
		numSyms := rapid.IntRange(1, 8).Draw(t, "numSymbols")
		usedNames := make(map[string]bool)

		for s := 0; s < numSyms; s++ {
			var name string
			for {
				name = rapid.StringMatching(`[A-Z][a-zA-Z]{1,10}`).Draw(t, "name")
				if !usedNames[name] {
					usedNames[name] = true
					break
				}
			}

			cat := genSymbolCategory().Draw(t, "cat")
			kinds := map[plugin.SymbolCategory][]string{
				plugin.CategoryCallable: {"function", "method"},
				plugin.CategoryType:     {"class", "interface"},
				plugin.CategoryValue:    {"variable"},
				plugin.CategoryModule:   {"module"},
				plugin.CategoryMeta:     {"generic_type_param"},
			}
			kind := rapid.SampledFrom(kinds[cat]).Draw(t, "kind")
			startLine := rapid.IntRange(1, 200).Draw(t, "start")
			endLine := rapid.IntRange(startLine, startLine+50).Draw(t, "end")

			// Use simple signatures without quotes or braces.
			sig := fmt.Sprintf("%s(x) -> void", name)

			props := make(map[string]string)
			if rapid.Bool().Draw(t, "async") {
				props["async"] = "true"
			}
			if rapid.Bool().Draw(t, "exported") {
				props["exported"] = "true"
			}

			sym := plugin.Symbol{
				ID:         filePath + "::" + name,
				Name:       name,
				FilePath:   filePath,
				Category:   cat,
				Kind:       kind,
				Signature:  sig,
				Properties: props,
				Span:       [2]int{startLine, endLine},
			}
			sg.Symbols = append(sg.Symbols, sym)
			allIDs = append(allIDs, sym.ID)
		}
	}

	// Generate edges between existing symbols.
	if len(allIDs) > 1 {
		numEdges := rapid.IntRange(0, len(allIDs)).Draw(t, "numEdges")
		for e := 0; e < numEdges; e++ {
			fi := rapid.IntRange(0, len(allIDs)-1).Draw(t, "fi")
			ti := rapid.IntRange(0, len(allIDs)-1).Draw(t, "ti")
			if fi == ti {
				continue
			}
			sg.Edges = append(sg.Edges, plugin.Edge{
				From: allIDs[fi],
				To:   allIDs[ti],
				Kind: genEdgeKind().Draw(t, "ek"),
			})
		}
	}

	sg.BuildIndexes()
	return sg
}

// Feature: code-concept-mapper, Property 1: SymbolGraph Round-Trip
func TestProperty1_SymbolGraphRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sg := genRoundTripSymbolGraph(t)
		opts := &EmitOptions{MaxLines: 500}

		output := emitInlineString(t, sg, opts)

		// Parse back.
		parsed, err := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, false)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		// Compare symbols.
		if len(parsed.Symbols) != len(sg.Symbols) {
			t.Fatalf("symbol count: got %d, want %d",
				len(parsed.Symbols), len(sg.Symbols))
		}

		// Build maps for comparison (order may differ due to file grouping).
		origSyms := make(map[string]plugin.Symbol)
		for _, s := range sg.Symbols {
			origSyms[s.ID] = s
		}
		for _, ps := range parsed.Symbols {
			orig, ok := origSyms[ps.ID]
			if !ok {
				t.Fatalf("parsed symbol ID %q not in original", ps.ID)
			}
			compareSymbols(t, &orig, &ps)
		}

		// Compare edges.
		if len(parsed.Edges) != len(sg.Edges) {
			t.Fatalf("edge count: got %d, want %d",
				len(parsed.Edges), len(sg.Edges))
		}
	})
}

func compareSymbols(t *rapid.T, orig, parsed *plugin.Symbol) {
	t.Helper()
	if parsed.Name != orig.Name {
		t.Fatalf("symbol %q name: got %q, want %q", orig.ID, parsed.Name, orig.Name)
	}
	if parsed.FilePath != orig.FilePath {
		t.Fatalf("symbol %q filePath: got %q, want %q", orig.ID, parsed.FilePath, orig.FilePath)
	}
	if parsed.Category != orig.Category {
		t.Fatalf("symbol %q category: got %q, want %q", orig.ID, parsed.Category, orig.Category)
	}
	if parsed.Kind != orig.Kind {
		t.Fatalf("symbol %q kind: got %q, want %q", orig.ID, parsed.Kind, orig.Kind)
	}
	if parsed.Signature != orig.Signature {
		t.Fatalf("symbol %q signature: got %q, want %q", orig.ID, parsed.Signature, orig.Signature)
	}
	if parsed.Span != orig.Span {
		t.Fatalf("symbol %q span: got %v, want %v", orig.ID, parsed.Span, orig.Span)
	}
	if len(parsed.Properties) != len(orig.Properties) {
		t.Fatalf("symbol %q properties count: got %d, want %d",
			orig.ID, len(parsed.Properties), len(orig.Properties))
	}
	for k, v := range orig.Properties {
		if parsed.Properties[k] != v {
			t.Fatalf("symbol %q property %q: got %q, want %q",
				orig.ID, k, parsed.Properties[k], v)
		}
	}
}

// Feature: code-concept-mapper, Property 15: Edge Type Coverage
func TestProperty15_EdgeTypeCoverage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Build a SymbolGraph with edges of all supported EdgeKind values.
		allKinds := []plugin.EdgeKind{
			plugin.EdgeCalls,
			plugin.EdgeInherits,
			plugin.EdgeContains,
			plugin.EdgeReferences,
			plugin.EdgeImplements,
			plugin.EdgeOverrides,
			plugin.EdgeImports,
			plugin.EdgeDecorates,
		}

		sg := &ir.SymbolGraph{}
		filePath := "src/test/edges.ts"

		// Create enough symbols for all edge kinds.
		for i := 0; i < len(allKinds)+1; i++ {
			name := fmt.Sprintf("Sym%d", i)
			sg.Symbols = append(sg.Symbols, plugin.Symbol{
				ID:         filePath + "::" + name,
				Name:       name,
				FilePath:   filePath,
				Category:   plugin.CategoryCallable,
				Kind:       "function",
				Signature:  fmt.Sprintf("%s() -> void", name),
				Properties: map[string]string{},
				Span:       [2]int{i*10 + 1, i*10 + 9},
			})
		}

		// Create one edge of each kind.
		for i, kind := range allKinds {
			sg.Edges = append(sg.Edges, plugin.Edge{
				From: sg.Symbols[i].ID,
				To:   sg.Symbols[i+1].ID,
				Kind: kind,
			})
		}

		sg.BuildIndexes()
		opts := &EmitOptions{MaxLines: 500}

		output := emitInlineString(t, sg, opts)

		// Parse back.
		parsed, err := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, false)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		// Verify all edge kinds are preserved.
		if len(parsed.Edges) != len(allKinds) {
			t.Fatalf("edge count: got %d, want %d",
				len(parsed.Edges), len(allKinds))
		}

		parsedKinds := make(map[plugin.EdgeKind]bool)
		for _, e := range parsed.Edges {
			parsedKinds[e.Kind] = true
		}
		for _, kind := range allKinds {
			if !parsedKinds[kind] {
				t.Fatalf("edge kind %q not found in parsed output", kind)
			}
		}
	})
}

// TestDirectoryTreeSplit_ExceedsMaxLines verifies that emitDirectoryTree splits
// a single source file's output into multiple _partN.skt files when it exceeds MaxLines.
func TestDirectoryTreeSplit_ExceedsMaxLines(t *testing.T) {
	// Create a symbol graph with enough symbols to exceed a small MaxLines.
	sg := &ir.SymbolGraph{}
	filePath := "src/pkg/bigfile.ts"
	numSyms := 50
	var allIDs []string
	for i := 0; i < numSyms; i++ {
		name := fmt.Sprintf("Symbol%d", i)
		sym := plugin.Symbol{
			ID:         filePath + "::" + name,
			Name:       name,
			FilePath:   filePath,
			Category:   plugin.CategoryCallable,
			Kind:       "function",
			Signature:  fmt.Sprintf("%s(x: number) -> void", name),
			Properties: map[string]string{"exported": "true"},
			Span:       [2]int{i*10 + 1, i*10 + 9},
		}
		sg.Symbols = append(sg.Symbols, sym)
		allIDs = append(allIDs, sym.ID)
	}
	// Add edges.
	for i := 0; i+1 < len(allIDs); i += 2 {
		sg.Edges = append(sg.Edges, plugin.Edge{
			From: allIDs[i],
			To:   allIDs[i+1],
			Kind: plugin.EdgeCalls,
		})
	}

	sg.BuildIndexes()

	// Use a temp directory as output.
	outDir := t.TempDir()

	// Create a temp file to act as InputPath (single file).
	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "bigfile.ts")
	if err := os.WriteFile(inputFile, []byte("// placeholder"), 0o600); err != nil {
		t.Fatal(err)
	}

	em := &Emitter{}
	opts := &EmitOptions{
		OutputDir:  outDir,
		OutputMode: "directory-tree",
		MaxLines:   20, // Very small to force splitting.
		InputPath:  inputFile,
		FileOrder:  []string{filePath},
	}

	written, err := em.Emit(sg, opts)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	// Should have produced multiple part files.
	if len(written) < 2 {
		t.Fatalf("expected multiple part files, got %d files: %v", len(written), written)
	}

	// Verify naming: all files should match _partN.skt pattern.
	for _, path := range written {
		base := filepath.Base(path)
		if !strings.Contains(base, "_part") || !strings.HasSuffix(base, ".skt") {
			t.Errorf("unexpected file name: %s", base)
		}
	}

	// Verify no file exceeds MaxLines.
	for _, path := range written {
		data, err := os.ReadFile(path) //nolint:gosec // test file path
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if len(lines) > opts.MaxLines {
			t.Errorf("file %s has %d lines, exceeds MaxLines %d", filepath.Base(path), len(lines), opts.MaxLines)
		}
	}

	// Verify each part has a [symbols] section.
	for _, path := range written {
		data, err := os.ReadFile(path) //nolint:gosec // test file path
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if !strings.Contains(string(data), "[symbols]") {
			t.Errorf("file %s missing [symbols] section", filepath.Base(path))
		}
	}
}

// TestDirectoryTreeSplit_SingleFileInput verifies that single-file input with splitting
// writes part files in the root of the output directory.
func TestDirectoryTreeSplit_SingleFileInput(t *testing.T) {
	sg := &ir.SymbolGraph{}
	filePath := "myfile.ts"
	numSyms := 30
	for i := 0; i < numSyms; i++ {
		name := fmt.Sprintf("Func%d", i)
		sg.Symbols = append(sg.Symbols, plugin.Symbol{
			ID:         filePath + "::" + name,
			Name:       name,
			FilePath:   filePath,
			Category:   plugin.CategoryCallable,
			Kind:       "function",
			Signature:  fmt.Sprintf("%s() -> void", name),
			Properties: map[string]string{},
			Span:       [2]int{i*5 + 1, i*5 + 4},
		})
	}

	sg.BuildIndexes()

	outDir := t.TempDir()
	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "myfile.ts")
	if err := os.WriteFile(inputFile, []byte("// placeholder"), 0o600); err != nil {
		t.Fatal(err)
	}

	em := &Emitter{}
	opts := &EmitOptions{
		OutputDir:  outDir,
		OutputMode: "directory-tree",
		MaxLines:   10,
		InputPath:  inputFile,
		FileOrder:  []string{filePath},
	}

	written, err := em.Emit(sg, opts)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	if len(written) < 2 {
		t.Fatalf("expected multiple part files, got %d", len(written))
	}

	// All files should be in the output directory root (no subdirectories).
	for _, path := range written {
		dir := filepath.Dir(path)
		if dir != outDir {
			t.Errorf("expected file in %s, got %s", outDir, dir)
		}
		base := filepath.Base(path)
		if !strings.HasPrefix(base, "myfile_part") || !strings.HasSuffix(base, ".skt") {
			t.Errorf("unexpected file name: %s", base)
		}
	}
}

// TestDirectoryTreeNoSplit_UnderMaxLines verifies that files under MaxLines are not split.
func TestDirectoryTreeNoSplit_UnderMaxLines(t *testing.T) {
	sg := &ir.SymbolGraph{}
	filePath := "src/small.ts"
	sg.Symbols = append(sg.Symbols, plugin.Symbol{
		ID:         filePath + "::Foo",
		Name:       "Foo",
		FilePath:   filePath,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  "Foo() -> void",
		Properties: map[string]string{},
		Span:       [2]int{1, 5},
	})

	sg.BuildIndexes()

	outDir := t.TempDir()
	// Create a temp directory as InputPath.
	inputDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(inputDir, "src"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "src", "small.ts"), []byte("//"), 0o600); err != nil {
		t.Fatal(err)
	}

	em := &Emitter{}
	opts := &EmitOptions{
		OutputDir:  outDir,
		OutputMode: "directory-tree",
		MaxLines:   500,
		InputPath:  inputDir,
		FileOrder:  []string{filePath},
	}

	written, err := em.Emit(sg, opts)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	// Should produce exactly one file, not split.
	if len(written) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(written), written)
	}

	// Should NOT have _part in the name.
	base := filepath.Base(written[0])
	if strings.Contains(base, "_part") {
		t.Errorf("file should not be split: %s", base)
	}
	if base != "small.skt" {
		t.Errorf("expected small.skt, got %s", base)
	}
}

// Feature: cli-output-modes, Property 5: Inline mode produces no files
func TestProperty_InlineModeProducesNoFiles(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genSymbolGraph(rt)

		// Create a fresh temp directory per iteration to check for disk writes.
		tmpDir, err := os.MkdirTemp("", "inline-no-files-*")
		if err != nil {
			rt.Fatalf("creating temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		var buf bytes.Buffer
		em := &Emitter{}
		opts := &EmitOptions{
			MaxLines:   500,
			OutputDir:  tmpDir,
			OutputMode: "inline",
		}

		emitErr := em.EmitInline(&buf, sg, opts)
		if emitErr != nil {
			rt.Fatalf("EmitInline: %v", emitErr)
		}

		// Non-empty SymbolGraphs should produce non-empty output.
		if len(sg.Symbols) > 0 && buf.Len() == 0 {
			rt.Fatalf("expected non-empty output for SymbolGraph with %d symbols, got empty buffer", len(sg.Symbols))
		}

		// Verify no files were created in the temp directory.
		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			rt.Fatalf("reading temp dir: %v", err)
		}
		if len(entries) != 0 {
			names := make([]string, len(entries))
			for i, e := range entries {
				names[i] = e.Name()
			}
			rt.Fatalf("expected no files on disk, found %d: %v", len(entries), names)
		}
	})
}

// Feature: cli-output-modes, Property 6: Inline concatenation
func TestProperty_InlineConcatenation(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genLargeSymbolGraph(rt)

		// Use a large MaxLines so all output fits in a single flat chunk.
		// The property (inline == joined flat chunks) only holds when there
		// is no splitting, because each chunk gets its own [symbols]/[edges]
		// headers in flat mode.
		opts := &EmitOptions{MaxLines: 10000}

		// Get individual file contents via flat mode.
		chunks := emitFlatFiles(rt, sg, opts)

		// Get inline output via EmitInline.
		em := &Emitter{}
		var buf bytes.Buffer
		if err := em.EmitInline(&buf, sg, opts); err != nil {
			rt.Fatalf("EmitInline: %v", err)
		}

		// The inline output should equal the chunks joined by "\n" (blank line separator).
		expected := strings.Join(chunks, "\n")
		if buf.String() != expected {
			rt.Fatalf("inline output does not match joined chunks:\ngot length=%d\nwant length=%d", buf.Len(), len(expected))
		}
	})
}

// Feature: cli-output-modes, Property 7: Directory-flat naming and structure
func TestProperty_DirectoryFlatNaming(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genSymbolGraph(rt)

		outDir, err := os.MkdirTemp("", "flat-naming-*")
		if err != nil {
			rt.Fatalf("creating temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(outDir) }()

		em := &Emitter{}
		opts := &EmitOptions{
			OutputDir:  outDir,
			OutputMode: config.OutputDirectoryFlat,
			MaxLines:   500,
		}

		written, err := em.Emit(sg, opts)
		if err != nil {
			rt.Fatalf("Emit: %v", err)
		}

		// Pattern: map_NNN.skt where NNN is zero-padded digits.
		namePattern := regexp.MustCompile(`^map_\d{3}\.skt$`)

		for i, path := range written {
			base := filepath.Base(path)
			dir := filepath.Dir(path)

			// (a) All files reside directly in the output directory (no subdirectories).
			if dir != outDir {
				rt.Fatalf("file %q not in output dir %q, got dir %q", base, outDir, dir)
			}

			// (b) Named map_NNN.skt with sequential zero-padded numbers starting at 001.
			expectedName := fmt.Sprintf("map_%03d.skt", i+1)
			if base != expectedName {
				rt.Fatalf("file %d: expected name %q, got %q", i, expectedName, base)
			}

			// (c) Matches the naming pattern and has .skt extension.
			if !namePattern.MatchString(base) {
				rt.Fatalf("file %q does not match pattern map_NNN.skt", base)
			}
		}

		// Verify no subdirectories were created in the output directory.
		entries, err := os.ReadDir(outDir)
		if err != nil {
			rt.Fatalf("reading output dir: %v", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				rt.Fatalf("unexpected subdirectory %q in output dir", entry.Name())
			}
			// Every file in the directory should have .skt extension.
			if filepath.Ext(entry.Name()) != ".skt" {
				rt.Fatalf("file %q does not have .skt extension", entry.Name())
			}
		}

		// The number of written files should match the number of entries in the directory.
		if len(written) != len(entries) {
			rt.Fatalf("written %d files but found %d entries in output dir", len(written), len(entries))
		}
	})
}

// Feature: cli-output-modes, Property 8: Directory-tree mirrors input structure
func TestProperty_DirectoryTreeMirroring(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genSymbolGraph(rt)

		// Collect unique file paths from the SymbolGraph.
		fileSet := make(map[string]bool)
		for _, sym := range sg.Symbols {
			fileSet[sym.FilePath] = true
		}
		fileOrder := make([]string, 0, len(fileSet))
		for fp := range fileSet {
			fileOrder = append(fileOrder, fp)
		}
		sort.Strings(fileOrder)

		// Create a temp input directory with actual files matching the SymbolGraph paths.
		// emitDirectoryTree calls os.Stat(opts.InputPath) to check if it's a directory.
		inputDir, err := os.MkdirTemp("", "tree-mirror-input-*")
		if err != nil {
			rt.Fatalf("creating input temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(inputDir) }()

		for _, fp := range fileOrder {
			fullPath := filepath.Join(inputDir, fp)
			fpDir := filepath.Dir(fullPath)
			if mkErr := os.MkdirAll(fpDir, 0o750); mkErr != nil {
				rt.Fatalf("creating input dir %s: %v", fpDir, mkErr)
			}
			if wErr := os.WriteFile(fullPath, []byte("// placeholder"), 0o600); wErr != nil {
				rt.Fatalf("creating input file %s: %v", fullPath, wErr)
			}
		}

		// Create a temp output directory.
		outDir, err := os.MkdirTemp("", "tree-mirror-output-*")
		if err != nil {
			rt.Fatalf("creating output temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(outDir) }()

		em := &Emitter{}
		opts := &EmitOptions{
			OutputDir:  outDir,
			OutputMode: config.OutputDirectoryTree,
			MaxLines:   500,
			InputPath:  inputDir,
			FileOrder:  fileOrder,
		}

		written, err := em.Emit(sg, opts)
		if err != nil {
			rt.Fatalf("Emit: %v", err)
		}

		// Build a set of written paths relative to the output directory.
		writtenRel := make(map[string]bool, len(written))
		for _, w := range written {
			rel, err := filepath.Rel(outDir, w)
			if err != nil {
				rt.Fatalf("computing relative path for %s: %v", w, err)
			}
			writtenRel[rel] = true
		}

		// For each source file, verify the output mirrors the directory structure
		// and replaces the extension with .skt.
		for _, srcRel := range fileOrder {
			ext := filepath.Ext(srcRel)
			expectedRel := strings.TrimSuffix(srcRel, ext) + ".skt"

			// Check either the exact file exists, or split parts exist (_partN.skt).
			if writtenRel[expectedRel] {
				// (a) Output directory structure mirrors input.
				srcDir := filepath.Dir(srcRel)
				outFileDir := filepath.Dir(expectedRel)
				if srcDir != outFileDir {
					rt.Fatalf("directory mismatch for %s: source dir %q, output dir %q", srcRel, srcDir, outFileDir)
				}

				// (b) Extension is .skt.
				if filepath.Ext(expectedRel) != ".skt" {
					rt.Fatalf("expected .skt extension for %s, got %s", expectedRel, filepath.Ext(expectedRel))
				}
			} else {
				// Could be split into parts — look for _partN.skt files.
				baseNoExt := strings.TrimSuffix(expectedRel, ".skt")
				foundPart := false
				for rel := range writtenRel {
					if !strings.HasPrefix(rel, baseNoExt+"_part") || !strings.HasSuffix(rel, ".skt") {
						continue
					}
					foundPart = true
					// Verify directory structure still mirrors input.
					srcDir := filepath.Dir(srcRel)
					outFileDir := filepath.Dir(rel)
					if srcDir != outFileDir {
						rt.Fatalf("directory mismatch for split file %s: source dir %q, output dir %q", rel, srcDir, outFileDir)
					}
					// Verify .skt extension.
					if filepath.Ext(rel) != ".skt" {
						rt.Fatalf("expected .skt extension for %s, got %s", rel, filepath.Ext(rel))
					}
				}
				if !foundPart {
					rt.Fatalf("no output file found for source %s: expected %s or split parts", srcRel, expectedRel)
				}
			}
		}

		// Verify ALL output files have .skt extension.
		for rel := range writtenRel {
			if filepath.Ext(rel) != ".skt" {
				rt.Fatalf("output file %s does not have .skt extension", rel)
			}
		}
	})
}

// Feature: cli-output-modes, Property 9: Directory-tree split naming
func TestProperty_DirectoryTreeSplitNaming(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate parameters: enough symbols to force splitting with a small MaxLines.
		numSyms := rapid.IntRange(20, 50).Draw(rt, "numSyms")
		maxLines := rapid.IntRange(10, 20).Draw(rt, "maxLines")
		filePath := "src/pkg/bigfile.ts"

		// Build a single-file SymbolGraph with numSyms symbols.
		sg := &ir.SymbolGraph{}
		var allIDs []string
		for i := 0; i < numSyms; i++ {
			name := fmt.Sprintf("Func%d", i)
			sym := plugin.Symbol{
				ID:         filePath + "::" + name,
				Name:       name,
				FilePath:   filePath,
				Category:   plugin.CategoryCallable,
				Kind:       "function",
				Signature:  fmt.Sprintf("%s(x: number) -> void", name),
				Properties: map[string]string{"exported": "true"},
				Span:       [2]int{i*10 + 1, i*10 + 9},
			}
			sg.Symbols = append(sg.Symbols, sym)
			allIDs = append(allIDs, sym.ID)
		}
		// Add edges between consecutive symbols.
		for i := 0; i+1 < len(allIDs); i += 2 {
			sg.Edges = append(sg.Edges, plugin.Edge{
				From: allIDs[i],
				To:   allIDs[i+1],
				Kind: plugin.EdgeCalls,
			})
		}

		sg.BuildIndexes()

		// Create actual input directory structure (emitDirectoryTree calls os.Stat).
		inputDir, err := os.MkdirTemp("", "split-naming-input-*")
		if err != nil {
			rt.Fatalf("creating input dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(inputDir) }()

		srcDir := filepath.Join(inputDir, "src", "pkg")
		if mkErr := os.MkdirAll(srcDir, 0o750); mkErr != nil {
			rt.Fatalf("creating src dir: %v", mkErr)
		}
		if wErr := os.WriteFile(filepath.Join(srcDir, "bigfile.ts"), []byte("// placeholder"), 0o600); wErr != nil {
			rt.Fatalf("creating input file: %v", wErr)
		}

		// Create output directory.
		outDir, err := os.MkdirTemp("", "split-naming-output-*")
		if err != nil {
			rt.Fatalf("creating output dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(outDir) }()

		em := &Emitter{}
		opts := &EmitOptions{
			OutputDir:  outDir,
			OutputMode: config.OutputDirectoryTree,
			MaxLines:   maxLines,
			InputPath:  inputDir,
			FileOrder:  []string{filePath},
		}

		written, err := em.Emit(sg, opts)
		if err != nil {
			rt.Fatalf("Emit: %v", err)
		}

		// With enough symbols and a small MaxLines, splitting must occur.
		if len(written) < 2 {
			rt.Fatalf("expected multiple part files with %d symbols and maxLines=%d, got %d files",
				numSyms, maxLines, len(written))
		}

		// (a) Verify files are named <basename>_partN.skt with sequential part numbers starting at 1.
		partPattern := regexp.MustCompile(`^bigfile_part(\d+)\.skt$`)
		for i, path := range written {
			base := filepath.Base(path)
			matches := partPattern.FindStringSubmatch(base)
			if matches == nil {
				rt.Fatalf("file %q does not match _partN.skt pattern", base)
			}
			// Part numbers should be sequential starting at 1.
			expectedPartNum := fmt.Sprintf("%d", i+1)
			if matches[1] != expectedPartNum {
				rt.Fatalf("file %d: expected part number %s, got %s (file: %s)",
					i, expectedPartNum, matches[1], base)
			}
		}

		// (b) Verify each part file does not exceed MaxLines.
		for _, path := range written {
			data, readErr := os.ReadFile(path) //nolint:gosec // test file path
			if readErr != nil {
				rt.Fatalf("reading %s: %v", path, readErr)
			}
			lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			if len(lines) > maxLines {
				rt.Fatalf("file %s has %d lines, exceeds MaxLines %d",
					filepath.Base(path), len(lines), maxLines)
			}
		}

		// (c) Part numbers are sequential starting at 1 (already verified above via loop index).
		// Additionally verify no gaps by checking the count matches the last part number.
		lastBase := filepath.Base(written[len(written)-1])
		lastMatches := partPattern.FindStringSubmatch(lastBase)
		if lastMatches == nil {
			rt.Fatalf("last file %q does not match pattern", lastBase)
		}
		var lastPartNum int
		if _, scanErr := fmt.Sscanf(lastMatches[1], "%d", &lastPartNum); scanErr != nil {
			rt.Fatalf("parsing part number %q: %v", lastMatches[1], scanErr)
		}
		if lastPartNum != len(written) {
			rt.Fatalf("last part number %d does not match total file count %d (gap in numbering)",
				lastPartNum, len(written))
		}
	})
}

// Feature: cli-output-modes, Property 11: MaxLines respected in output
func TestProperty_MaxLinesRespected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genLargeSymbolGraph(rt)
		maxLines := rapid.IntRange(50, 500).Draw(rt, "maxLines")

		// --- directory-flat mode ---
		flatOutDir, err := os.MkdirTemp("", "maxlines-flat-*")
		if err != nil {
			rt.Fatalf("creating flat output dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(flatOutDir) }()

		em := &Emitter{}
		flatOpts := &EmitOptions{
			OutputDir:  flatOutDir,
			OutputMode: config.OutputDirectoryFlat,
			MaxLines:   maxLines,
		}

		flatWritten, err := em.Emit(sg, flatOpts)
		if err != nil {
			rt.Fatalf("Emit directory-flat: %v", err)
		}

		// In flat mode, file blocks are not split, so allow 20% slack.
		flatSlack := maxLines + maxLines/5
		for _, path := range flatWritten {
			data, readErr := os.ReadFile(path) //nolint:gosec // test file path
			if readErr != nil {
				rt.Fatalf("reading flat file %s: %v", path, readErr)
			}
			lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			if len(lines) > flatSlack {
				rt.Fatalf("directory-flat file %s has %d lines, exceeds MaxLines %d + 20%% slack (%d)",
					filepath.Base(path), len(lines), maxLines, flatSlack)
			}
		}

		// --- directory-tree mode ---
		// Collect unique file paths from the SymbolGraph.
		fileSet := make(map[string]bool)
		for _, sym := range sg.Symbols {
			fileSet[sym.FilePath] = true
		}
		fileOrder := make([]string, 0, len(fileSet))
		for fp := range fileSet {
			fileOrder = append(fileOrder, fp)
		}
		sort.Strings(fileOrder)

		// Create actual input directory with files matching the SymbolGraph paths.
		inputDir, err := os.MkdirTemp("", "maxlines-tree-input-*")
		if err != nil {
			rt.Fatalf("creating tree input dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(inputDir) }()

		for _, fp := range fileOrder {
			fullPath := filepath.Join(inputDir, fp)
			fpDir := filepath.Dir(fullPath)
			if mkErr := os.MkdirAll(fpDir, 0o750); mkErr != nil {
				rt.Fatalf("creating input dir %s: %v", fpDir, mkErr)
			}
			if wErr := os.WriteFile(fullPath, []byte("// placeholder"), 0o600); wErr != nil {
				rt.Fatalf("creating input file %s: %v", fullPath, wErr)
			}
		}

		treeOutDir, err := os.MkdirTemp("", "maxlines-tree-output-*")
		if err != nil {
			rt.Fatalf("creating tree output dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(treeOutDir) }()

		treeOpts := &EmitOptions{
			OutputDir:  treeOutDir,
			OutputMode: config.OutputDirectoryTree,
			MaxLines:   maxLines,
			InputPath:  inputDir,
			FileOrder:  fileOrder,
		}

		treeWritten, err := em.Emit(sg, treeOpts)
		if err != nil {
			rt.Fatalf("Emit directory-tree: %v", err)
		}

		// In tree mode, files are split per source file, so MaxLines should be respected exactly.
		for _, path := range treeWritten {
			data, readErr := os.ReadFile(path) //nolint:gosec // test file path
			if readErr != nil {
				rt.Fatalf("reading tree file %s: %v", path, readErr)
			}
			lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			if len(lines) > maxLines {
				rt.Fatalf("directory-tree file %s has %d lines, exceeds MaxLines %d",
					filepath.Base(path), len(lines), maxLines)
			}
		}
	})
}

// Feature: cli-output-modes, Property 12: Emitter determinism across modes
func TestProperty_EmitterDeterminismAcrossModes(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		sg := genSymbolGraph(rt)
		em := &Emitter{}

		// --- Inline mode: emit twice to separate buffers, compare ---
		inlineOpts := &EmitOptions{
			MaxLines:   500,
			OutputMode: config.OutputInline,
		}

		var buf1, buf2 bytes.Buffer
		if err := em.EmitInline(&buf1, sg, inlineOpts); err != nil {
			rt.Fatalf("EmitInline (1st): %v", err)
		}
		if err := em.EmitInline(&buf2, sg, inlineOpts); err != nil {
			rt.Fatalf("EmitInline (2nd): %v", err)
		}
		if buf1.String() != buf2.String() {
			rt.Fatalf("inline mode: outputs differ between two calls\nlen1=%d len2=%d", buf1.Len(), buf2.Len())
		}

		// --- Directory-flat mode: emit twice to different temp dirs, compare ---
		flatDir1, err := os.MkdirTemp("", "determ-flat1-*")
		if err != nil {
			rt.Fatalf("creating flat dir 1: %v", err)
		}
		defer func() { _ = os.RemoveAll(flatDir1) }()

		flatDir2, err := os.MkdirTemp("", "determ-flat2-*")
		if err != nil {
			rt.Fatalf("creating flat dir 2: %v", err)
		}
		defer func() { _ = os.RemoveAll(flatDir2) }()

		flatOpts1 := &EmitOptions{
			OutputDir:  flatDir1,
			OutputMode: config.OutputDirectoryFlat,
			MaxLines:   500,
		}
		flatOpts2 := &EmitOptions{
			OutputDir:  flatDir2,
			OutputMode: config.OutputDirectoryFlat,
			MaxLines:   500,
		}

		flatWritten1, err := em.Emit(sg, flatOpts1)
		if err != nil {
			rt.Fatalf("Emit flat (1st): %v", err)
		}
		flatWritten2, err := em.Emit(sg, flatOpts2)
		if err != nil {
			rt.Fatalf("Emit flat (2nd): %v", err)
		}

		if len(flatWritten1) != len(flatWritten2) {
			rt.Fatalf("directory-flat: different file counts: %d vs %d", len(flatWritten1), len(flatWritten2))
		}
		for i := range flatWritten1 {
			data1, readErr1 := os.ReadFile(flatWritten1[i]) //nolint:gosec // test file path
			if readErr1 != nil {
				rt.Fatalf("reading flat file 1[%d]: %v", i, readErr1)
			}
			data2, readErr2 := os.ReadFile(flatWritten2[i]) //nolint:gosec // test file path
			if readErr2 != nil {
				rt.Fatalf("reading flat file 2[%d]: %v", i, readErr2)
			}
			if !bytes.Equal(data1, data2) {
				rt.Fatalf("directory-flat: file %d contents differ\nfile1=%s\nfile2=%s",
					i, filepath.Base(flatWritten1[i]), filepath.Base(flatWritten2[i]))
			}
		}

		// --- Directory-tree mode: emit twice to different temp dirs, compare ---
		// Collect unique file paths and create input directory structure.
		fileSet := make(map[string]bool)
		for _, sym := range sg.Symbols {
			fileSet[sym.FilePath] = true
		}
		fileOrder := make([]string, 0, len(fileSet))
		for fp := range fileSet {
			fileOrder = append(fileOrder, fp)
		}
		sort.Strings(fileOrder)

		inputDir, err := os.MkdirTemp("", "determ-tree-input-*")
		if err != nil {
			rt.Fatalf("creating tree input dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(inputDir) }()

		for _, fp := range fileOrder {
			fullPath := filepath.Join(inputDir, fp)
			fpDir := filepath.Dir(fullPath)
			if mkErr := os.MkdirAll(fpDir, 0o750); mkErr != nil {
				rt.Fatalf("creating input dir %s: %v", fpDir, mkErr)
			}
			if wErr := os.WriteFile(fullPath, []byte("// placeholder"), 0o600); wErr != nil {
				rt.Fatalf("creating input file %s: %v", fullPath, wErr)
			}
		}

		treeDir1, err := os.MkdirTemp("", "determ-tree1-*")
		if err != nil {
			rt.Fatalf("creating tree dir 1: %v", err)
		}
		defer func() { _ = os.RemoveAll(treeDir1) }()

		treeDir2, err := os.MkdirTemp("", "determ-tree2-*")
		if err != nil {
			rt.Fatalf("creating tree dir 2: %v", err)
		}
		defer func() { _ = os.RemoveAll(treeDir2) }()

		treeOpts1 := &EmitOptions{
			OutputDir:  treeDir1,
			OutputMode: config.OutputDirectoryTree,
			MaxLines:   500,
			InputPath:  inputDir,
			FileOrder:  fileOrder,
		}
		treeOpts2 := &EmitOptions{
			OutputDir:  treeDir2,
			OutputMode: config.OutputDirectoryTree,
			MaxLines:   500,
			InputPath:  inputDir,
			FileOrder:  fileOrder,
		}

		treeWritten1, err := em.Emit(sg, treeOpts1)
		if err != nil {
			rt.Fatalf("Emit tree (1st): %v", err)
		}
		treeWritten2, err := em.Emit(sg, treeOpts2)
		if err != nil {
			rt.Fatalf("Emit tree (2nd): %v", err)
		}

		if len(treeWritten1) != len(treeWritten2) {
			rt.Fatalf("directory-tree: different file counts: %d vs %d", len(treeWritten1), len(treeWritten2))
		}
		for i := range treeWritten1 {
			data1, readErr1 := os.ReadFile(treeWritten1[i]) //nolint:gosec // test file path
			if readErr1 != nil {
				rt.Fatalf("reading tree file 1[%d]: %v", i, readErr1)
			}
			data2, readErr2 := os.ReadFile(treeWritten2[i]) //nolint:gosec // test file path
			if readErr2 != nil {
				rt.Fatalf("reading tree file 2[%d]: %v", i, readErr2)
			}
			if !bytes.Equal(data1, data2) {
				rt.Fatalf("directory-tree: file %d contents differ\nfile1=%s\nfile2=%s",
					i, filepath.Base(treeWritten1[i]), filepath.Base(treeWritten2[i]))
			}
		}
	})
}

// TestErrorsSection_EmptyWhenNoErrors verifies that no [errors] section appears
// when the SymbolGraph has no parse errors.
func TestErrorsSection_EmptyWhenNoErrors(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "main.go::main", Name: "main", FilePath: "main.go", Category: plugin.CategoryCallable, Kind: "function", Signature: "main()", Properties: map[string]string{}, Span: [2]int{1, 3}},
		},
	}
	sg.BuildIndexes()
	opts := &EmitOptions{MaxLines: 500}

	output := emitInlineString(t, sg, opts)
	if strings.Contains(output, "[errors]") {
		t.Error("output should not contain [errors] section when there are no errors")
	}
}

// TestErrorsSection_AppearsInFlatOutput verifies that [errors] appears at the end
// of the last flat output file when the SymbolGraph has parse errors.
func TestErrorsSection_AppearsInFlatOutput(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "main.go::main", Name: "main", FilePath: "main.go", Category: plugin.CategoryCallable, Kind: "function", Signature: "main()", Properties: map[string]string{}, Span: [2]int{1, 3}},
		},
		Errors: []ir.ParseError{
			{FilePath: "broken.go", Reason: "syntax error at line 5"},
			{FilePath: "bad.py", Reason: "unexpected indent"},
		},
	}
	sg.BuildIndexes()
	opts := &EmitOptions{MaxLines: 500}

	output := emitInlineString(t, sg, opts)

	if !strings.Contains(output, "[errors]") {
		t.Fatal("output should contain [errors] section")
	}

	// [errors] must appear after [symbols].
	errIdx := strings.Index(output, "[errors]")
	symIdx := strings.Index(output, "[symbols]")
	if errIdx <= symIdx {
		t.Error("[errors] should appear after [symbols]")
	}

	// Verify both errors are present, sorted by file path.
	if !strings.Contains(output, "- bad.py: unexpected indent") {
		t.Error("missing error for bad.py")
	}
	if !strings.Contains(output, "- broken.go: syntax error at line 5") {
		t.Error("missing error for broken.go")
	}

	// Verify order: bad.py before broken.go (alphabetical).
	badIdx := strings.Index(output, "- bad.py:")
	brokenIdx := strings.Index(output, "- broken.go:")
	if badIdx >= brokenIdx {
		t.Error("errors should be sorted alphabetically by file path")
	}
}

// TestErrorsSection_RoundTrip verifies that [errors] survives a round-trip
// through the emitter and parser.
func TestErrorsSection_RoundTrip(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "main.go::main", Name: "main", FilePath: "main.go", Category: plugin.CategoryCallable, Kind: "function", Signature: "main()", Properties: map[string]string{}, Span: [2]int{1, 3}},
		},
		Errors: []ir.ParseError{
			{FilePath: "broken.go", Reason: "syntax error at line 5"},
			{FilePath: "bad.py", Reason: "unexpected indent"},
		},
	}
	sg.BuildIndexes()
	opts := &EmitOptions{MaxLines: 500}

	output := emitInlineString(t, sg, opts)

	parsed, parseErr := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, false)
	if parseErr != nil {
		t.Fatalf("ParseOutput: %v", parseErr)
	}

	if len(parsed.Errors) != 2 {
		t.Fatalf("expected 2 parse errors, got %d", len(parsed.Errors))
	}

	// Verify round-tripped errors match (sorted by filepath).
	sort.Slice(parsed.Errors, func(i, j int) bool {
		return parsed.Errors[i].FilePath < parsed.Errors[j].FilePath
	})
	if parsed.Errors[0].FilePath != "bad.py" || parsed.Errors[0].Reason != "unexpected indent" {
		t.Errorf("error[0] mismatch: got %+v", parsed.Errors[0])
	}
	if parsed.Errors[1].FilePath != "broken.go" || parsed.Errors[1].Reason != "syntax error at line 5" {
		t.Errorf("error[1] mismatch: got %+v", parsed.Errors[1])
	}
}

// TestErrorsSection_InlineMode verifies that [errors] appears in inline output.
func TestErrorsSection_InlineMode(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "main.go::main", Name: "main", FilePath: "main.go", Category: plugin.CategoryCallable, Kind: "function", Signature: "main()", Properties: map[string]string{}, Span: [2]int{1, 3}},
		},
		Errors: []ir.ParseError{
			{FilePath: "broken.go", Reason: "syntax error"},
		},
	}
	sg.BuildIndexes()
	em := &Emitter{}
	opts := &EmitOptions{MaxLines: 500}

	var buf bytes.Buffer
	if err := em.EmitInline(&buf, sg, opts); err != nil {
		t.Fatalf("EmitInline: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[errors]") {
		t.Fatal("inline output should contain [errors] section")
	}
	if !strings.Contains(output, "- broken.go: syntax error") {
		t.Fatal("inline output missing error entry")
	}
}

// TestErrorsSection_TreeMode verifies that warnings.skt is written in tree mode.
func TestErrorsSection_TreeMode(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "/tmp/src/main.go::main", Name: "main", FilePath: "/tmp/src/main.go", Category: plugin.CategoryCallable, Kind: "function", Signature: "main()", Properties: map[string]string{"exported": "true"}, Span: [2]int{1, 3}},
		},
		Errors: []ir.ParseError{
			{FilePath: "/tmp/src/broken.go", Reason: "syntax error"},
		},
	}
	sg.BuildIndexes()

	outDir := t.TempDir()
	// Create a minimal input file so Stat doesn't fail.
	inputDir := t.TempDir()

	em := &Emitter{}
	opts := &EmitOptions{
		OutputDir:  outDir,
		OutputMode: config.OutputDirectoryTree,
		InputPath:  inputDir,
		FileOrder:  []string{"main.go"},
		MaxLines:   500,
	}

	written, err := em.Emit(sg, opts)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	// warnings.skt should be among the written files.
	var foundErrors bool
	for _, path := range written {
		if filepath.Base(path) != "warnings.skt" {
			continue
		}
		foundErrors = true
		data, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("read warnings.skt: %v", readErr)
		}
		content := string(data)
		if !strings.Contains(content, "[errors]") {
			t.Error("warnings.skt missing [errors] section")
		}
		if !strings.Contains(content, "broken.go: syntax error") {
			t.Error("warnings.skt missing error entry")
		}
	}
	if !foundErrors {
		t.Error("warnings.skt not found in written files")
	}
}
