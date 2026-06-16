// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package java

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
package com.example;

public class App extends BaseApp implements Runnable {
    private final String name;

    public App(String name) {
        this.name = name;
    }

    @Override
    public void run() {
        System.out.println(name);
    }

    public static void main(String[] args) {
        new App("test").run();
    }
}
`)

	symbols, _, err := parseSource(t, "App.java", src)
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
		"com.example": {plugin.CategoryModule, "package"},
		"name":        {plugin.CategoryValue, "field"},
		"run":         {plugin.CategoryCallable, "method"},
		"main":        {plugin.CategoryCallable, "method"},
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

	// Verify both class and constructor named "App" are present.
	foundClass := false
	foundCtor := false
	for _, s := range symbols {
		if s.Name == "App" && s.Category == plugin.CategoryType && s.Kind == "class" {
			foundClass = true
		}
		if s.Name == "App" && s.Category == plugin.CategoryCallable && s.Kind == "constructor" {
			foundCtor = true
		}
	}
	if !foundClass {
		t.Error("missing class symbol 'App'")
	}
	if !foundCtor {
		t.Error("missing constructor symbol 'App'")
	}
}

func TestParseSource_ModifierExtraction(t *testing.T) {
	src := []byte(`
public class Foo {
    public static void staticMethod() {}
    private final int count = 0;
    protected abstract void abstractMethod();
    public synchronized void syncMethod() {}
}
`)
	symbols, _, err := parseSource(t, "Foo.java", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	// Class modifiers.
	if props["Foo"]["visibility"] != "public" {
		t.Errorf("Foo visibility: got %q, want %q", props["Foo"]["visibility"], "public")
	}

	// Static method.
	if props["staticMethod"]["static"] != "true" {
		t.Errorf("staticMethod should have static=true")
	}
	if props["staticMethod"]["visibility"] != "public" {
		t.Errorf("staticMethod visibility: got %q, want %q", props["staticMethod"]["visibility"], "public")
	}

	// Private final field.
	if props["count"]["visibility"] != "private" {
		t.Errorf("count visibility: got %q, want %q", props["count"]["visibility"], "private")
	}
	if props["count"]["final"] != "true" {
		t.Errorf("count should have final=true")
	}

	// Abstract method.
	if props["abstractMethod"]["abstract"] != "true" {
		t.Errorf("abstractMethod should have abstract=true")
	}
	if props["abstractMethod"]["visibility"] != "protected" {
		t.Errorf("abstractMethod visibility: got %q, want %q", props["abstractMethod"]["visibility"], "protected")
	}

	// Synchronized method.
	if props["syncMethod"]["synchronized"] != "true" {
		t.Errorf("syncMethod should have synchronized=true")
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
public class Child extends Parent implements Serializable, Comparable {
}
`)
	_, edges, err := parseSource(t, "Child.java", src)
	if err != nil {
		t.Fatal(err)
	}

	foundInherits := false
	var implTargets []string
	for _, e := range edges {
		if e.From == "Child" && e.Kind == plugin.EdgeInherits && e.To == "Parent" {
			foundInherits = true
		}
		if e.From == "Child" && e.Kind == plugin.EdgeImplements {
			implTargets = append(implTargets, e.To)
		}
	}
	if !foundInherits {
		t.Error("expected inherits edge from Child to Parent")
	}
	sort.Strings(implTargets)
	wantImpl := []string{"Comparable", "Serializable"}
	if len(implTargets) != len(wantImpl) {
		t.Fatalf("expected implements targets %v, got %v", wantImpl, implTargets)
	}
	for i := range wantImpl {
		if implTargets[i] != wantImpl[i] {
			t.Errorf("implements target[%d]: got %q, want %q", i, implTargets[i], wantImpl[i])
		}
	}
}

func TestParseSource_ContainsEdges(t *testing.T) {
	src := []byte(`
public class MyClass {
    private int x;
    public void doSomething() {}
    public MyClass() {}
}
`)
	_, edges, err := parseSource(t, "MyClass.java", src)
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
	want := []string{"MyClass.MyClass", "MyClass.doSomething", "MyClass.x"}
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
public class App {
    public void run() {
        System.out.println("hello");
        helper();
    }
    private void helper() {}
}
`)
	_, edges, err := parseSource(t, "App.java", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "App.run" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	if len(callTargets) == 0 {
		t.Fatal("expected call edges from run")
	}
	// Should contain at least "helper".
	foundHelper := false
	for _, ct := range callTargets {
		if ct == "helper" {
			foundHelper = true
		}
	}
	if !foundHelper {
		t.Errorf("expected call edge from run to helper, got targets: %v", callTargets)
	}
}

func TestParseSource_OverrideDetection(t *testing.T) {
	src := []byte(`
public class Child extends Parent {
    @Override
    public void doWork() {}
}
`)
	_, edges, err := parseSource(t, "Child.java", src)
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
		t.Error("expected overrides edge for @Override method doWork")
	}
}

func TestParseSource_InterfaceDeclaration(t *testing.T) {
	src := []byte(`
public interface Service {
    void start();
    void stop();
}
`)
	symbols, edges, err := parseSource(t, "Service.java", src)
	if err != nil {
		t.Fatal(err)
	}

	foundInterface := false
	for _, s := range symbols {
		if s.Name == "Service" && s.Kind == "interface" {
			foundInterface = true
		}
	}
	if !foundInterface {
		t.Error("missing interface symbol 'Service'")
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Service" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Service.start", "Service.stop"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_EnumDeclaration(t *testing.T) {
	src := []byte(`
public enum Color {
    RED, GREEN, BLUE
}
`)
	symbols, _, err := parseSource(t, "Color.java", src)
	if err != nil {
		t.Fatal(err)
	}

	foundEnum := false
	for _, s := range symbols {
		if s.Name == "Color" && s.Kind == "enum" && s.Category == plugin.CategoryType {
			foundEnum = true
			if s.Properties["visibility"] != "public" {
				t.Errorf("Color visibility: got %q, want %q", s.Properties["visibility"], "public")
			}
		}
	}
	if !foundEnum {
		t.Error("missing enum symbol 'Color'")
	}
}

func TestParseSource_AnnotationType(t *testing.T) {
	src := []byte(`
public @interface MyAnnotation {
}
`)
	symbols, _, err := parseSource(t, "MyAnnotation.java", src)
	if err != nil {
		t.Fatal(err)
	}

	foundAnnotation := false
	for _, s := range symbols {
		if s.Name == "MyAnnotation" && s.Kind == "annotation" && s.Category == plugin.CategoryType {
			foundAnnotation = true
		}
	}
	if !foundAnnotation {
		t.Error("missing annotation type symbol 'MyAnnotation'")
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`public class Broken {
    public void bad( {
    }
}`)
	_, _, err := parseSource(t, "Broken.java", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "Broken.java") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`public class Broken {
    public void bad( {
    }
}`)
	symbols, _, err := parseSource(t, "Broken.java", src)
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
	if len(exts) != 1 || exts[0] != ".java" {
		t.Errorf("expected [.java], got %v", exts)
	}
}
