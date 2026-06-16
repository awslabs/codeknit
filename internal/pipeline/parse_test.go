// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/typescript"

	"pgregory.net/rapid"
)

// Feature: code-concept-mapper, Property 16: Parallel Parsing Determinism
func TestProperty_ParallelParsingDeterminism(t *testing.T) {
	dir := t.TempDir()
	sources := []struct {
		name string
		src  string
	}{
		{"alpha", "export function alpha(x: number): number { return x; }\n"},
		{"beta", "export function beta(a: string): string { return a; }\n"},
		{"gamma", "export class Gamma { run(): void {} }\n"},
		{"delta", "export const delta = (n: number) => n * 2;\n"},
		{"epsilon", "interface Epsilon { value: string; }\nfunction epsilon(): void {}\n"},
	}

	files := make([]string, 0, len(sources))
	for _, s := range sources {
		fp := filepath.Join(dir, s.name+".ts")
		if err := os.WriteFile(fp, []byte(s.src), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, fp)
	}

	reg := plugin.NewRegistry()
	reg.Register(typescript.NewPlugin())

	rapid.Check(t, func(rt *rapid.T) {
		workers := rapid.IntRange(1, 8).Draw(rt, "workers")

		pr1 := ParseFiles(files, reg, 1, nil, false)
		if len(pr1.Skipped) > 0 || len(pr1.ParseErrors) > 0 {
			rt.Fatalf("sequential parse errors: skipped=%d, parseErrors=%d", len(pr1.Skipped), len(pr1.ParseErrors))
		}

		pr2 := ParseFiles(files, reg, workers, nil, false)
		if len(pr2.Skipped) > 0 || len(pr2.ParseErrors) > 0 {
			rt.Fatalf("parallel parse errors (workers=%d): skipped=%d, parseErrors=%d", workers, len(pr2.Skipped), len(pr2.ParseErrors))
		}

		if len(pr1.Symbols) != len(pr2.Symbols) {
			rt.Fatalf("symbol map sizes differ: %d vs %d", len(pr1.Symbols), len(pr2.Symbols))
		}

		for fp, s1 := range pr1.Symbols {
			s2, ok := pr2.Symbols[fp]
			if !ok {
				rt.Fatalf("file %q missing from parallel result", fp)
			}
			if !symbolSlicesEqual(s1, s2) {
				rt.Fatalf("symbols differ for %q", fp)
			}
			if !edgeSlicesEqual(pr1.Edges[fp], pr2.Edges[fp]) {
				rt.Fatalf("edges differ for %q", fp)
			}
		}
	})
}

func symbolSlicesEqual(a, b []plugin.Symbol) bool {
	if len(a) != len(b) {
		return false
	}
	sa := sortedSymbols(a)
	sb := sortedSymbols(b)
	for i := range sa {
		if sa[i].Name != sb[i].Name || sa[i].Category != sb[i].Category || sa[i].Kind != sb[i].Kind ||
			sa[i].Signature != sb[i].Signature || sa[i].Span != sb[i].Span {
			return false
		}
	}
	return true
}

func edgeSlicesEqual(a, b []plugin.Edge) bool {
	if len(a) != len(b) {
		return false
	}
	sa := sortedEdges(a)
	sb := sortedEdges(b)
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func sortedSymbols(s []plugin.Symbol) []plugin.Symbol {
	out := make([]plugin.Symbol, len(s))
	copy(out, s)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedEdges(e []plugin.Edge) []plugin.Edge {
	out := make([]plugin.Edge, len(e))
	copy(out, e)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

type mockPlugin struct {
	err     error
	exts    []string
	symbols []plugin.Symbol
	edges   []plugin.Edge
}

func (m *mockPlugin) Extensions() []string            { return m.exts }
func (m *mockPlugin) TestPatterns() plugin.TestConfig { return plugin.TestConfig{} }
func (m *mockPlugin) Parse(_ string) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	return m.symbols, m.edges, m.err
}

func TestParseFiles_NonFatalErrorsCollected(t *testing.T) {
	dir := t.TempDir()

	good1 := filepath.Join(dir, "a.mock")
	good2 := filepath.Join(dir, "b.mock")
	bad := filepath.Join(dir, "c.unknown")
	for _, f := range []string{good1, good2, bad} {
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	reg := plugin.NewRegistry()
	reg.Register(&mockPlugin{
		exts:    []string{".mock"},
		symbols: []plugin.Symbol{{Name: "sym", Category: plugin.CategoryCallable, Kind: "function"}},
	})

	pr := ParseFiles([]string{good1, good2, bad}, reg, 2, nil, false)

	if len(pr.Symbols) != 2 {
		t.Fatalf("expected 2 parsed files, got %d", len(pr.Symbols))
	}

	if len(pr.Skipped) != 1 {
		t.Fatalf("expected 1 skipped file, got %d", len(pr.Skipped))
	}
	if pr.Skipped[0].FilePath != bad {
		t.Fatalf("expected skip for %q, got %q", bad, pr.Skipped[0].FilePath)
	}
}

func TestParseFiles_SyntaxWarningsCollected(t *testing.T) {
	dir := t.TempDir()

	bad := filepath.Join(dir, "broken.ts")
	if err := os.WriteFile(bad, []byte("function broken( { }"), 0o600); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(dir, "ok.ts")
	if err := os.WriteFile(good, []byte("function ok(): void {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	reg := plugin.NewRegistry()
	reg.Register(typescript.NewPlugin())

	pr := ParseFiles([]string{bad, good}, reg, 1, nil, false)

	// Both files should be parsed (broken one with partial results).
	if len(pr.Symbols) != 2 {
		t.Fatalf("expected 2 parsed files, got %d", len(pr.Symbols))
	}

	// One parse error warning for the broken file.
	if len(pr.ParseErrors) == 0 {
		t.Fatal("expected a ParseError for the broken file")
	}
}
