// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ruby

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
class Animal
  def initialize(name)
    @name = name
  end

  def speak
    @name
  end
end

class Dog < Animal
  def speak
    @name + " barks"
  end
end

module Utils
  def self.helper
    puts "help"
  end
end

def greet(animal)
  puts animal.speak
end

MAX_RETRIES = 3
counter = 0
`)

	symbols, _, err := parseSource(t, "test.rb", src)
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
		"Utils":       {plugin.CategoryModule, "module"},
		"greet":       {plugin.CategoryCallable, "function"},
		"MAX_RETRIES": {plugin.CategoryValue, "constant"},
		"counter":     {plugin.CategoryValue, "variable"},
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

func TestParseSource_VisibilityTracking(t *testing.T) {
	src := []byte(`
class MyClass
  def public_method
  end

  private

  def private_method
  end

  def another_private
  end

  protected

  def protected_method
  end

  public

  def back_to_public
  end
end
`)
	symbols, _, err := parseSource(t, "test.rb", src)
	if err != nil {
		t.Fatal(err)
	}

	wantVis := map[string]string{
		"public_method":    "",
		"private_method":   "private",
		"another_private":  "private",
		"protected_method": "protected",
		"back_to_public":   "",
	}

	for _, sym := range symbols {
		want, ok := wantVis[sym.Name]
		if !ok {
			continue
		}
		got := sym.Properties["visibility"]
		if got != want {
			t.Errorf("method %q: visibility got %q, want %q", sym.Name, got, want)
		}
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
class Base
end

class Child < Base
end
`)
	_, edges, err := parseSource(t, "test.rb", src)
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
class MyClass
  def method_a
  end

  def method_b
  end
end
`)
	_, edges, err := parseSource(t, "test.rb", src)
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
def foo
  bar
  baz(1)
end
`)
	_, edges, err := parseSource(t, "test.rb", src)
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
	// In Ruby, `bar` without parens is an identifier (not a call), but `baz(1)` is a call.
	found := false
	for _, target := range callTargets {
		if target == "baz" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected call edge to 'baz', got targets: %v", callTargets)
	}
}

func TestParseSource_SingletonMethod(t *testing.T) {
	src := []byte(`
class MyClass
  def self.class_method
    puts "hello"
  end
end
`)
	symbols, edges, err := parseSource(t, "test.rb", src)
	if err != nil {
		t.Fatal(err)
	}

	foundSingleton := false
	for _, sym := range symbols {
		if sym.Name == "class_method" && sym.Kind == "method" {
			foundSingleton = true
			if sym.Properties["static"] != "true" {
				t.Errorf("singleton method should have static=true, got %q", sym.Properties["static"])
			}
		}
	}
	if !foundSingleton {
		t.Error("missing singleton method 'class_method'")
	}

	// Check contains edge.
	foundContains := false
	for _, e := range edges {
		if e.From == "MyClass" && e.To == "MyClass.class_method" && e.Kind == plugin.EdgeContains {
			foundContains = true
		}
	}
	if !foundContains {
		t.Error("expected contains edge from MyClass to class_method")
	}
}

func TestParseSource_ModuleExtraction(t *testing.T) {
	src := []byte(`
module Helpers
  def helper_method
  end
end
`)
	symbols, edges, err := parseSource(t, "test.rb", src)
	if err != nil {
		t.Fatal(err)
	}

	foundModule := false
	for _, sym := range symbols {
		if sym.Name == "Helpers" && sym.Kind == "module" && sym.Category == plugin.CategoryModule {
			foundModule = true
		}
	}
	if !foundModule {
		t.Error("missing module symbol 'Helpers'")
	}

	// Check contains edge for module method.
	foundContains := false
	for _, e := range edges {
		if e.From == "Helpers" && e.To == "Helpers.helper_method" && e.Kind == plugin.EdgeContains {
			foundContains = true
		}
	}
	if !foundContains {
		t.Error("expected contains edge from Helpers to helper_method")
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`def broken(`)
	_, _, err := parseSource(t, "bad.rb", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.rb") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`def broken(`)
	symbols, _, err := parseSource(t, "bad.rb", src)
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
	if len(exts) != 1 || exts[0] != ".rb" {
		t.Errorf("expected [.rb], got %v", exts)
	}
}

// Property 14: Ruby visibility tracking
// For any Ruby class with visibility modifier calls (public, private, protected),
// methods defined after a visibility call should have that visibility,
// and methods before any visibility call should default to "public".
// **Validates: Requirement 5.4**
func TestProperty_RubyVisibilityTracking(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		// Generate random method names.
		pubName := genIdent().Draw(t, "pubName")
		privName := genIdent().Draw(t, "privName")
		protName := genIdent().Draw(t, "protName")

		// Ensure all names are unique.
		if pubName == privName || pubName == protName || privName == protName {
			return
		}

		var b strings.Builder
		b.WriteString("class TestClass\n")
		b.WriteString("  def " + pubName + "\n  end\n\n")
		b.WriteString("  private\n\n")
		b.WriteString("  def " + privName + "\n  end\n\n")
		b.WriteString("  protected\n\n")
		b.WriteString("  def " + protName + "\n  end\n")
		b.WriteString("end\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.rb", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}

		foundPub := false
		foundPriv := false
		foundProt := false
		for _, sym := range symbols {
			if sym.Name == pubName {
				foundPub = true
				if sym.Properties["visibility"] != "" {
					t.Errorf("method %q before any visibility call should have visibility=\"\", got %q",
						pubName, sym.Properties["visibility"])
				}
			}
			if sym.Name == privName {
				foundPriv = true
				if sym.Properties["visibility"] != "private" {
					t.Errorf("method %q after private call should have visibility=private, got %q",
						privName, sym.Properties["visibility"])
				}
			}
			if sym.Name == protName {
				foundProt = true
				if sym.Properties["visibility"] != "protected" {
					t.Errorf("method %q after protected call should have visibility=protected, got %q",
						protName, sym.Properties["visibility"])
				}
			}
		}
		if !foundPub {
			t.Errorf("missing public method %q in symbols", pubName)
		}
		if !foundPriv {
			t.Errorf("missing private method %q in symbols", privName)
		}
		if !foundProt {
			t.Errorf("missing protected method %q in symbols", protName)
		}
	})
}

// genIdent generates a valid lowercase Ruby identifier (a-z, 3-8 chars).
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
