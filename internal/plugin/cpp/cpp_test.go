// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cpp

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/plugin"
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
#include <iostream>

namespace myns {

class Animal {
public:
    virtual void speak() = 0;
    virtual ~Animal() {}
};

struct Point {
    int x;
    int y;
};

enum Color { RED, GREEN, BLUE };

void greet() {
    std::cout << "hello";
}

int global_var = 42;

}
`)

	symbols, _, err := parseSource(t, "test.cpp", src)
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
		"myns":       {plugin.CategoryModule, "namespace"},
		"Animal":     {plugin.CategoryType, "class"},
		"Point":      {plugin.CategoryType, "struct"},
		"Color":      {plugin.CategoryType, "enum"},
		"greet":      {plugin.CategoryCallable, "function"},
		"global_var": {plugin.CategoryValue, "variable"},
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

func TestParseSource_VisibilityExtraction(t *testing.T) {
	src := []byte(`
class Foo {
public:
    void pubMethod();
    int pubField;
private:
    void privMethod();
    int privField;
protected:
    void protMethod();
};
`)
	symbols, _, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	wantVis := map[string]string{
		"pubMethod":  "public",
		"pubField":   "public",
		"privMethod": "private",
		"privField":  "private",
		"protMethod": "protected",
	}

	for name, wantV := range wantVis {
		p, ok := props[name]
		if !ok {
			t.Errorf("missing symbol %q", name)
			continue
		}
		if p["visibility"] != wantV {
			t.Errorf("symbol %q: visibility got %q, want %q", name, p["visibility"], wantV)
		}
	}
}

func TestParseSource_VirtualMethodExtraction(t *testing.T) {
	src := []byte(`
class Base {
public:
    virtual void doWork();
    void normalMethod();
};
`)
	symbols, _, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	if props["doWork"]["virtual"] != "true" {
		t.Errorf("doWork should have virtual=true, got %q", props["doWork"]["virtual"])
	}
	if props["normalMethod"]["virtual"] != "" {
		t.Errorf("normalMethod should have virtual=\"\", got %q", props["normalMethod"]["virtual"])
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
class Base {};
class Child : public Base {};
`)
	_, edges, err := parseSource(t, "test.cpp", src)
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
class MyClass {
public:
    void doSomething() {}
    int value;
};
`)
	_, edges, err := parseSource(t, "test.cpp", src)
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
	want := []string{"MyClass.doSomething", "MyClass.value"}
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
void helper() {}

void caller() {
    helper();
}
`)
	_, edges, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	foundCall := false
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "caller" && e.To == "helper" {
			foundCall = true
		}
	}
	if !foundCall {
		t.Error("expected call edge from caller to helper")
	}
}

func TestParseSource_OverrideDetection(t *testing.T) {
	src := []byte(`
class Base {
public:
    virtual void doWork() {}
};

class Child : public Base {
public:
    void doWork() override {}
};
`)
	_, edges, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	foundOverride := false
	for _, e := range edges {
		if e.Kind == plugin.EdgeOverrides && e.From == "Child.doWork" {
			foundOverride = true
		}
	}
	if !foundOverride {
		t.Error("expected overrides edge for override method doWork")
	}
}

func TestParseSource_NamespaceExtraction(t *testing.T) {
	src := []byte(`
namespace outer {
    void foo() {}
    namespace inner {
        void bar() {}
    }
}
`)
	symbols, _, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	foundOuter := false
	foundInner := false
	foundFoo := false
	foundBar := false
	for _, s := range symbols {
		switch s.Name {
		case "outer":
			foundOuter = true
			if s.Kind != "namespace" {
				t.Errorf("outer: got kind %q, want namespace", s.Kind)
			}
		case "inner":
			foundInner = true
		case "foo":
			foundFoo = true
		case "bar":
			foundBar = true
		}
	}
	if !foundOuter {
		t.Error("missing namespace 'outer'")
	}
	if !foundInner {
		t.Error("missing namespace 'inner'")
	}
	if !foundFoo {
		t.Error("missing function 'foo'")
	}
	if !foundBar {
		t.Error("missing function 'bar'")
	}
}

func TestParseSource_TemplateExtraction(t *testing.T) {
	src := []byte(`
template<typename T>
class Container {
public:
    void add(T item) {}
};

template<typename T>
T identity(T x) { return x; }
`)
	symbols, _, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	foundClass := false
	foundFunc := false
	for _, s := range symbols {
		if s.Name == "Container" && s.Kind == "class" {
			foundClass = true
		}
		if s.Name == "identity" && s.Kind == "function" {
			foundFunc = true
		}
	}
	if !foundClass {
		t.Error("missing template class 'Container'")
	}
	if !foundFunc {
		t.Error("missing template function 'identity'")
	}
}

func TestParseSource_IncludeReferences(t *testing.T) {
	src := []byte(`
#include <iostream>
#include "myheader.h"

void foo() {}
`)
	_, edges, err := parseSource(t, "test.cpp", src)
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
	want := []string{"\"myheader.h\"", "<iostream>"}
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
	_, _, err := parseSource(t, "bad.cpp", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.cpp") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`void broken( {`)
	symbols, _, err := parseSource(t, "bad.cpp", src)
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
	if len(exts) != 5 {
		t.Fatalf("expected 5 extensions, got %d", len(exts))
	}
	sort.Strings(exts)
	want := []string{".cc", ".cpp", ".cxx", ".hpp", ".hxx"}
	for i := range want {
		if exts[i] != want[i] {
			t.Errorf("extension[%d]: got %q, want %q", i, exts[i], want[i])
		}
	}
}

func TestParseSource_AbstractClass(t *testing.T) {
	src := []byte(`
class Shape {
public:
    virtual void draw() = 0;
    virtual double area() = 0;
};
`)
	symbols, _, err := parseSource(t, "test.cpp", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range symbols {
		if s.Name == "Shape" {
			if s.Properties["abstract"] != "true" {
				t.Errorf("Shape should have abstract=true, got %q", s.Properties["abstract"])
			}
			return
		}
	}
	t.Error("missing class symbol 'Shape'")
}
