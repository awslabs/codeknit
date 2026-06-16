// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clang

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/plugin"

	"pgregory.net/rapid"
)

// parseSource writes src to a temp file and parses it via the plugin.
func parseSource(t *testing.T, filename string, src []byte) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return NewPlugin().Parse(path)
}

func TestParseSource_Symbols(t *testing.T) {
	src := []byte(`
void greet(const char *name) {
    printf("Hello, %s\n", name);
}

static int helper(void) { return 0; }

struct Config {
    char *name;
    int value;
};

union Data {
    int i;
    float f;
};

enum Color { RED, GREEN, BLUE };

typedef unsigned long size_t;

int global_var = 42;

#define MAX_SIZE 100

#define SQUARE(x) ((x) * (x))
`)

	symbols, _, err := parseSource(t, "test.c", src)
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
		"greet":      {plugin.CategoryCallable, "function"},
		"helper":     {plugin.CategoryCallable, "function"},
		"Config":     {plugin.CategoryType, "struct"},
		"Data":       {plugin.CategoryType, "union"},
		"Color":      {plugin.CategoryType, "enum"},
		"size_t":     {plugin.CategoryType, "typedef"},
		"global_var": {plugin.CategoryValue, "variable"},
		"MAX_SIZE":   {plugin.CategoryValue, "macro"},
		"SQUARE":     {plugin.CategoryCallable, "macro_function"},
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

func TestParseSource_StaticExternDetection(t *testing.T) {
	src := []byte(`
static int static_func(void) { return 0; }
extern void extern_func(void);
void normal_func(void) {}

static int static_var = 1;
extern int extern_var;
int normal_var = 2;
`)
	symbols, _, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	wantStatic := map[string]string{
		"static_func": "true",
		"extern_func": "",
		"normal_func": "",
		"static_var":  "true",
		"extern_var":  "",
		"normal_var":  "",
	}
	wantExtern := map[string]string{
		"static_func": "",
		"extern_func": "true",
		"normal_func": "",
		"static_var":  "",
		"extern_var":  "true",
		"normal_var":  "",
	}

	for _, sym := range symbols {
		if ws, ok := wantStatic[sym.Name]; ok {
			if sym.Properties["static"] != ws {
				t.Errorf("symbol %q: static got %q, want %q", sym.Name, sym.Properties["static"], ws)
			}
		}
		if we, ok := wantExtern[sym.Name]; ok {
			if sym.Properties["extern"] != we {
				t.Errorf("symbol %q: extern got %q, want %q", sym.Name, sym.Properties["extern"], we)
			}
		}
	}
}

func TestParseSource_StructContainsFields(t *testing.T) {
	src := []byte(`
struct Point {
    int x;
    int y;
};
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Point" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Point.x", "Point.y"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_UnionContainsFields(t *testing.T) {
	src := []byte(`
union Value {
    int i;
    float f;
    char c;
};
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Value" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Value.c", "Value.f", "Value.i"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_CallEdges(t *testing.T) {
	src := []byte(`
void helper(void) {}

void caller(void) {
    helper();
    printf("hello");
}
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "caller" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	want := []string{"helper", "printf"}
	if len(callTargets) != len(want) {
		t.Fatalf("expected call targets %v, got %v", want, callTargets)
	}
	for i := range want {
		if callTargets[i] != want[i] {
			t.Errorf("call target[%d]: got %q, want %q", i, callTargets[i], want[i])
		}
	}
}

func TestParseSource_IncludeReferences(t *testing.T) {
	src := []byte(`
#include <stdio.h>
#include "myheader.h"

void foo(void) {}
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	var refs []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeReferences {
			refs = append(refs, e.To)
		}
	}
	sort.Strings(refs)
	want := []string{"\"myheader.h\"", "<stdio.h>"}
	if len(refs) != len(want) {
		t.Fatalf("expected references %v, got %v", want, refs)
	}
	for i := range want {
		if refs[i] != want[i] {
			t.Errorf("reference[%d]: got %q, want %q", i, refs[i], want[i])
		}
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`void broken( {`)
	_, _, err := parseSource(t, "bad.c", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.c") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`void broken( {`)
	symbols, _, err := parseSource(t, "bad.c", src)
	var sw *plugin.SyntaxError
	if err != nil && !errors.As(err, &sw) {
		t.Fatalf("expected nil or SyntaxWarning, got: %v", err)
	}
	if symbols == nil {
		t.Fatal("expected non-nil symbols for partial-error file")
	}
}

func TestExtensions(t *testing.T) {
	p := NewPlugin()
	exts := p.Extensions()
	if len(exts) != 2 {
		t.Fatalf("expected 2 extensions, got %d", len(exts))
	}
	sort.Strings(exts)
	if exts[0] != ".c" || exts[1] != ".h" {
		t.Errorf("expected [.c .h], got %v", exts)
	}
}

// Property 15: C static/extern detection
// For any C source file, a function Symbol should have Properties["static"] == "true"
// if and only if the declaration has a static storage class specifier.
// **Validates: Requirement 5.10**
func TestProperty_CStaticExternDetection(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		staticName := genCIdent().Draw(t, "staticName")
		normalName := genCIdent().Draw(t, "normalName")

		if staticName == normalName {
			return
		}

		var b strings.Builder
		b.WriteString("static void " + staticName + "(void) {}\n\n")
		b.WriteString("void " + normalName + "(void) {}\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.c", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}

		foundStatic := false
		foundNormal := false
		for _, sym := range symbols {
			if sym.Name == staticName {
				foundStatic = true
				if sym.Properties["static"] != "true" {
					t.Errorf("static function %q should have static=true, got %q", staticName, sym.Properties["static"])
				}
			}
			if sym.Name == normalName {
				foundNormal = true
				if sym.Properties["static"] != "" {
					t.Errorf("normal function %q should have static=\"\", got %q", normalName, sym.Properties["static"])
				}
			}
		}
		if !foundStatic {
			t.Errorf("missing static function %q in symbols", staticName)
		}
		if !foundNormal {
			t.Errorf("missing normal function %q in symbols", normalName)
		}
	})
}

// genCIdent generates a valid lowercase C identifier (a-z, 3-8 chars).
func genCIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		for i := range chars {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}
