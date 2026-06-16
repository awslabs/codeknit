// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package python

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
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        return self.name

class Dog(Animal):
    def speak(self):
        return self.name + " barks"

def greet(animal):
    print(animal.speak())

async def fetch_data(url):
    pass

MAX_RETRIES = 3
`)

	symbols, _, err := parseSource(t, "test.py", src)
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
		"Animal":      {plugin.CategoryType, "class"},
		"Dog":         {plugin.CategoryType, "class"},
		"speak":       {plugin.CategoryCallable, "method"},
		"greet":       {plugin.CategoryCallable, "function"},
		"fetch_data":  {plugin.CategoryCallable, "function"},
		"MAX_RETRIES": {plugin.CategoryValue, "variable"},
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

func TestParseSource_AsyncDetection(t *testing.T) {
	src := []byte(`
def sync_func():
    pass

async def async_func():
    pass
`)
	symbols, _, err := parseSource(t, "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		switch sym.Name {
		case "sync_func":
			if sym.Properties["async"] != "" {
				t.Errorf("sync_func should have async=\"\", got %q", sym.Properties["async"])
			}
		case "async_func":
			if sym.Properties["async"] != "true" {
				t.Errorf("async_func should have async=true, got %q", sym.Properties["async"])
			}
		}
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
class Base:
    pass

class Child(Base):
    pass
`)
	_, edges, err := parseSource(t, "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	foundInherits := false
	for _, e := range edges {
		if e.From == "Child" && e.To == "Base" && e.Kind == plugin.EdgeInherits {
			foundInherits = true
		}
	}
	if !foundInherits {
		t.Error("expected inherits edge from Child to Base")
	}
}

func TestParseSource_ContainsEdges(t *testing.T) {
	src := []byte(`
class MyClass:
    def method_a(self):
        pass

    def method_b(self):
        pass
`)
	_, edges, err := parseSource(t, "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "MyClass" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"MyClass.method_a", "MyClass.method_b"}
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
def foo():
    bar()
    baz()
    obj.method()
`)
	_, edges, err := parseSource(t, "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "foo" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	want := []string{"bar", "baz"}
	sort.Strings(want)
	if len(callTargets) != len(want) {
		t.Fatalf("expected call targets %v, got %v", want, callTargets)
	}
	for i := range want {
		if callTargets[i] != want[i] {
			t.Errorf("call target[%d]: got %q, want %q", i, callTargets[i], want[i])
		}
	}
}

func TestParseSource_DecoratedDefinition(t *testing.T) {
	src := []byte(`
class MyClass:
    @staticmethod
    def static_method():
        pass

    @classmethod
    def class_method(cls):
        pass

    def regular_method(self):
        pass
`)
	symbols, edges, err := parseSource(t, "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		switch sym.Name {
		case "static_method":
			if sym.Properties["static"] != "true" {
				t.Errorf("static_method should have static=true")
			}
		case "class_method":
			if sym.Properties["classmethod"] != "true" {
				t.Errorf("class_method should have classmethod=true")
			}
		case "regular_method":
			if sym.Properties["static"] == "true" {
				t.Errorf("regular_method should not have static=true")
			}
		}
	}

	// Verify all methods are contained in MyClass.
	var containsTargets []string
	for _, e := range edges {
		if e.From == "MyClass" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"MyClass.class_method", "MyClass.regular_method", "MyClass.static_method"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`def broken(`)
	_, _, err := parseSource(t, "bad.py", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.py") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`def broken(`)
	symbols, _, err := parseSource(t, "bad.py", src)
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
	if len(exts) != 2 || exts[0] != ".py" || exts[1] != ".pyi" {
		t.Errorf("expected [.py .pyi], got %v", exts)
	}
}

// Property 13: Python async detection
// For any Python source file containing functions, a function Symbol should have
// Properties["async"] == "true" if and only if the function declaration uses the async keyword.
func TestProperty_PythonAsyncDetection(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		// Generate random function names.
		syncName := genIdent().Draw(t, "syncName")
		asyncName := genIdent().Draw(t, "asyncName")

		// Ensure names are unique.
		if syncName == asyncName {
			return
		}

		var b strings.Builder
		b.WriteString("def " + syncName + "():\n    pass\n\n")
		b.WriteString("async def " + asyncName + "():\n    pass\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.py", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}

		foundSync := false
		foundAsync := false
		for _, sym := range symbols {
			if sym.Name == syncName {
				foundSync = true
				if sym.Properties["async"] != "" {
					t.Errorf("sync function %q should have async=\"\", got %q", syncName, sym.Properties["async"])
				}
			}
			if sym.Name == asyncName {
				foundAsync = true
				if sym.Properties["async"] != "true" {
					t.Errorf("async function %q should have async=true, got %q", asyncName, sym.Properties["async"])
				}
			}
		}
		if !foundSync {
			t.Errorf("missing sync function %q in symbols", syncName)
		}
		if !foundAsync {
			t.Errorf("missing async function %q in symbols", asyncName)
		}
	})
}

// genIdent generates a valid lowercase Python identifier (a-z, 3-8 chars).
func genIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		for i := range chars {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}
