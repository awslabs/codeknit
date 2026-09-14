// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeknit/internal/config"
	"codeknit/internal/fingerprint"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

func TestTreeOutputKeepsSameStemSources(t *testing.T) {
	source := t.TempDir()
	graph := &ir.SymbolGraph{}
	for _, name := range []string{"foo.c", "foo.h", "dict.go", "warnings.go"} {
		file := filepath.Join(source, name)
		if err := os.WriteFile(file, []byte("source"), 0o600); err != nil {
			t.Fatal(err)
		}
		graph.Symbols = append(graph.Symbols, plugin.Symbol{
			ID: file + "::Unique", FilePath: file, Name: name, Signature: name,
			Category: plugin.CategoryType, Kind: "struct", Span: [2]int{1, 1},
		})
	}
	graph.Errors = []ir.ParseError{{FilePath: "broken.go", Reason: "syntax warning"}}
	graph.BuildIndexes()
	out := t.TempDir()
	files, err := (&Emitter{}).Emit(graph, &EmitOptions{
		OutputMode: config.OutputDirectoryTree, OutputDir: out, InputPath: source,
		MaxLines: 20, Minify: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	unique := make(map[string]bool)
	for _, file := range files {
		if unique[file] {
			t.Errorf("multiple outputs collide at %s", file)
		}
		unique[file] = true
	}
	if len(unique) != len(graph.Symbols)+2 {
		t.Fatalf("lost output: got %d unique files for %d sources plus dictionary and warnings", len(unique), len(graph.Symbols))
	}
}

func TestFailedGenerationPreservesPreviousOutput(t *testing.T) {
	out := t.TempDir()
	old := filepath.Join(out, "old.skt")
	if err := os.WriteFile(old, []byte("previous complete result"), 0o600); err != nil {
		t.Fatal(err)
	}
	// This unrelated directory blocks the new output filename.
	if err := os.Mkdir(filepath.Join(out, "map_001.skt"), 0o700); err != nil {
		t.Fatal(err)
	}
	graph := &ir.SymbolGraph{Symbols: []plugin.Symbol{{
		ID: "a.go::Read", FilePath: "a.go", Name: "Read", Signature: "Read()",
		Category: plugin.CategoryCallable, Kind: "function", Span: [2]int{1, 1},
	}}}
	graph.BuildIndexes()
	_, err := (&Emitter{}).Emit(graph, &EmitOptions{OutputDir: out, Clean: true, MaxLines: 500})
	if err == nil {
		t.Fatal("expected the output path collision to fail")
	}
	content, readErr := os.ReadFile(old) //nolint:gosec // previous test output
	if readErr != nil || string(content) != "previous complete result" {
		t.Fatalf("failed generation destroyed previous output: %q, %v", content, readErr)
	}
}

func exactCloneEntries(n int) []fpEntry {
	tokens := []byte{
		plugin.FPIf, plugin.FPEq, plugin.FPReturn, plugin.FPAdd,
		plugin.FPFor, plugin.FPAssign, plugin.FPAdd, plugin.FPReturn,
	}
	entries := make([]fpEntry, n)
	for i := range entries {
		entries[i] = fpEntry{
			filePath: "clones.go", name: fmt.Sprintf("F%04d", i),
			category: plugin.CategoryCallable, tokens: tokens, fp: fingerprint.Hash(tokens),
		}
	}
	return entries
}

func TestExactDuplicatesAcrossCandidateThreshold(t *testing.T) {
	for _, n := range []int{499, 500, 501} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			matches := findDuplicates(exactCloneEntries(n), 100, 100)
			if want := n * (n - 1) / 2; len(matches) != want {
				t.Fatalf("got %d exact pairs, want all %d", len(matches), want)
			}
			seen := make(map[candidatePair]bool, len(matches))
			for _, match := range matches {
				key := newCandidatePair(match.i, match.j)
				if seen[key] {
					t.Fatalf("duplicate result: %+v", key)
				}
				seen[key] = true
			}
		})
	}
}

func TestDuplicateOrderingIsDeterministic(t *testing.T) {
	entries := exactCloneEntries(501)
	first := findDuplicates(entries, 100, 100)
	second := findDuplicates(entries, 100, 100)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("unchanged inputs produced different duplicate ordering")
	}
}

func TestTreeChunksRetainFileHeaders(t *testing.T) {
	block := fileBlock{
		filePath: "app.go",
		symbolLines: []string{
			"## app.go",
			"S1 callable/function L1-L1 First()",
			"S2 callable/function L2-L2 Second()",
			"S3 callable/function L3-L3 Third()",
		},
		edgeLines: []string{"S1 --calls--> S2", "S2 --calls--> S3"},
	}
	for _, chunk := range splitTreeContent(block, 4) {
		if strings.Contains(chunk, "callable/") && !strings.Contains(chunk, "## app.go\n") {
			t.Fatalf("symbol chunk lost its source identity:\n%s", chunk)
		}
		if lines := strings.Count(chunk, "\n"); lines > 4 {
			t.Fatalf("chunk exceeded its line budget: %d", lines)
		}
	}
}

func TestOutputCannotReplaceSourceDirectory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "src")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(source, "app.go")
	if err := os.WriteFile(file, []byte("package app"), 0o600); err != nil {
		t.Fatal(err)
	}
	graph := &ir.SymbolGraph{}
	graph.BuildIndexes()
	for _, output := range []string{source, root} {
		if _, err := (&Emitter{}).Emit(graph, &EmitOptions{
			InputPath: source, OutputDir: output, Clean: true,
		}); err == nil {
			t.Fatalf("accepted output overlapping source: %s", output)
		}
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Skip(err)
	}
	if _, err := (&Emitter{}).Emit(graph, &EmitOptions{
		InputPath: source, OutputDir: filepath.Join(alias, "src"), Clean: true,
	}); err == nil {
		t.Fatal("source alias bypassed overlap protection")
	}
	content, err := os.ReadFile(file) //nolint:gosec // verify source fixture preservation
	if err != nil || string(content) != "package app" {
		t.Fatalf("source was changed: %q, %v", content, err)
	}
}
