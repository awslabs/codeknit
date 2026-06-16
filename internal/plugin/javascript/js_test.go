// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"codeknit/internal/plugin"
)

// parseSource writes src to a temp file and parses it via the plugin.
func parseSource(t *testing.T, src []byte) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.js")
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return NewPlugin().Parse(path)
}

func TestParseSource_JSSymbols(t *testing.T) {
	src := []byte(`
export function greet(name) {
  return helper(name);
}

export class MyClass extends BaseClass {
  getValue() { return 0; }
}

const DEFAULT_OPTIONS = { timeout: 30 };

let count = 0;

const helper = (x) => x * 2;
`)

	symbols, _, err := parseSource(t, src)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected non-empty symbols")
	}

	type catKind struct {
		Category plugin.SymbolCategory
		Kind     string
	}
	found := make(map[string]catKind)
	for _, s := range symbols {
		found[s.Name] = catKind{s.Category, s.Kind}
	}

	expect := map[string]catKind{
		"greet":   {plugin.CategoryCallable, "function"},
		"MyClass": {plugin.CategoryType, "class"},
		"count":   {plugin.CategoryValue, "variable"},
		"helper":  {plugin.CategoryCallable, "arrow_function"},
	}

	for name, want := range expect {
		got, ok := found[name]
		if !ok {
			t.Errorf("missing symbol %q (expected %s/%s)", name, want.Category, want.Kind)
			continue
		}
		if got.Category != want.Category || got.Kind != want.Kind {
			t.Errorf("symbol %q: got %s/%s, want %s/%s", name, got.Category, got.Kind, want.Category, want.Kind)
		}
	}
}

func TestExtensions(t *testing.T) {
	p := NewPlugin()
	exts := p.Extensions()
	if len(exts) != 2 || exts[0] != ".js" || exts[1] != ".jsx" {
		t.Errorf("expected [.js .jsx], got %v", exts)
	}
}
