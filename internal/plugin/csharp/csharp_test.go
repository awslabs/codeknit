// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csharp

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
namespace MyApp {
    public class App : BaseApp, IRunnable {
        private string name;

        public void Run() {
            Console.WriteLine(name);
        }

        public static void Main(string[] args) {
            new App().Run();
        }
    }

    public interface IService {
        void Start();
    }

    public struct Point {
        public int X;
    }

    public enum Color { Red, Green, Blue }

    public delegate void MyHandler(string msg);
}
`)

	symbols, _, err := parseSource(t, "App.cs", src)
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
		"MyApp":     {plugin.CategoryModule, "namespace"},
		"App":       {plugin.CategoryType, "class"},
		"name":      {plugin.CategoryValue, "field"},
		"Run":       {plugin.CategoryCallable, "method"},
		"Main":      {plugin.CategoryCallable, "method"},
		"IService":  {plugin.CategoryType, "interface"},
		"Start":     {plugin.CategoryCallable, "method"},
		"Point":     {plugin.CategoryType, "struct"},
		"X":         {plugin.CategoryValue, "field"},
		"Color":     {plugin.CategoryType, "enum"},
		"MyHandler": {plugin.CategoryType, "delegate"},
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

func TestParseSource_ModifierExtraction(t *testing.T) {
	src := []byte(`
public class Foo {
    public static void StaticMethod() {}
    private int count;
    protected abstract void AbstractMethod();
    public sealed class Inner {}
    public virtual void VirtualMethod() {}
    public async void AsyncMethod() {}
}
`)
	symbols, _, err := parseSource(t, "Foo.cs", src)
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
	if props["StaticMethod"]["static"] != "true" {
		t.Errorf("StaticMethod should have static=true")
	}
	if props["StaticMethod"]["visibility"] != "public" {
		t.Errorf("StaticMethod visibility: got %q, want %q", props["StaticMethod"]["visibility"], "public")
	}

	// Private field.
	if props["count"]["visibility"] != "private" {
		t.Errorf("count visibility: got %q, want %q", props["count"]["visibility"], "private")
	}

	// Abstract method.
	if props["AbstractMethod"]["abstract"] != "true" {
		t.Errorf("AbstractMethod should have abstract=true")
	}
	if props["AbstractMethod"]["visibility"] != "protected" {
		t.Errorf("AbstractMethod visibility: got %q, want %q", props["AbstractMethod"]["visibility"], "protected")
	}

	// Sealed class.
	if props["Inner"]["sealed"] != "true" {
		t.Errorf("Inner should have sealed=true")
	}

	// Virtual method.
	if props["VirtualMethod"]["virtual"] != "true" {
		t.Errorf("VirtualMethod should have virtual=true")
	}

	// Async method.
	if props["AsyncMethod"]["async"] != "true" {
		t.Errorf("AsyncMethod should have async=true")
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
public class Child : Parent, ISerializable, IComparable {
}
`)
	_, edges, err := parseSource(t, "Child.cs", src)
	if err != nil {
		t.Fatal(err)
	}

	// C# base_list is syntactically flat — the parser cannot distinguish a base
	// class from an implemented interface without cross-file type information.
	// All base list entries are emitted as inherits edges.
	var inheritsTargets []string
	for _, e := range edges {
		if e.From == "Child" && e.Kind == plugin.EdgeInherits {
			inheritsTargets = append(inheritsTargets, e.To)
		}
	}
	sort.Strings(inheritsTargets)
	want := []string{"IComparable", "ISerializable", "Parent"}
	if len(inheritsTargets) != len(want) {
		t.Fatalf("expected inherits targets %v, got %v", want, inheritsTargets)
	}
	for i := range want {
		if inheritsTargets[i] != want[i] {
			t.Errorf("inherits target[%d]: got %q, want %q", i, inheritsTargets[i], want[i])
		}
	}
}

func TestParseSource_ContainsEdges(t *testing.T) {
	src := []byte(`
public class MyClass {
    private int x;
    public string Name { get; set; }
    public void DoSomething() {}
}
`)
	_, edges, err := parseSource(t, "MyClass.cs", src)
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
	want := []string{"MyClass.DoSomething", "MyClass.Name", "MyClass.x"}
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
    public void Run() {
        Console.WriteLine("hello");
        Helper();
    }
    private void Helper() {}
}
`)
	_, edges, err := parseSource(t, "App.cs", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "App.Run" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	if len(callTargets) == 0 {
		t.Fatal("expected call edges from Run")
	}
	foundHelper := false
	for _, ct := range callTargets {
		if ct == "Helper" {
			foundHelper = true
		}
	}
	if !foundHelper {
		t.Errorf("expected call edge from Run to Helper, got targets: %v", callTargets)
	}
}

func TestParseSource_OverrideDetection(t *testing.T) {
	src := []byte(`
public class Child : Parent {
    public override void DoWork() {}
}
`)
	_, edges, err := parseSource(t, "Child.cs", src)
	if err != nil {
		t.Fatal(err)
	}

	foundOverride := false
	for _, e := range edges {
		if e.Kind == plugin.EdgeOverrides && e.From == "Child.DoWork" {
			foundOverride = true
		}
	}
	if !foundOverride {
		t.Error("expected overrides edge for override method DoWork")
	}
}

func TestParseSource_PropertyAndDelegate(t *testing.T) {
	src := []byte(`
public class Config {
    public string Name { get; set; }
    public static int Count { get; }
    public delegate void Handler(int value);
}
`)
	symbols, edges, err := parseSource(t, "Config.cs", src)
	if err != nil {
		t.Fatal(err)
	}

	// Check property symbols.
	foundName := false
	foundCount := false
	foundDelegate := false
	for _, s := range symbols {
		if s.Name == "Name" && s.Kind == "property" && s.Category == plugin.CategoryValue {
			foundName = true
			if s.Properties["visibility"] != "public" {
				t.Errorf("Name visibility: got %q, want %q", s.Properties["visibility"], "public")
			}
		}
		if s.Name == "Count" && s.Kind == "property" {
			foundCount = true
			if s.Properties["static"] != "true" {
				t.Errorf("Count should have static=true")
			}
		}
		if s.Name == "Handler" && s.Kind == "delegate" && s.Category == plugin.CategoryType {
			foundDelegate = true
			if s.Properties["visibility"] != "public" {
				t.Errorf("Handler visibility: got %q, want %q", s.Properties["visibility"], "public")
			}
		}
	}
	if !foundName {
		t.Error("missing property symbol 'Name'")
	}
	if !foundCount {
		t.Error("missing property symbol 'Count'")
	}
	if !foundDelegate {
		t.Error("missing delegate symbol 'Handler'")
	}

	// Check contains edges.
	var containsTargets []string
	for _, e := range edges {
		if e.From == "Config" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Config.Count", "Config.Handler", "Config.Name"}
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
	src := []byte(`public class Broken {
    public void Bad( {
    }
}`)
	_, _, err := parseSource(t, "Broken.cs", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "Broken.cs") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`public class Broken {
    public void Bad( {
    }
}`)
	symbols, _, err := parseSource(t, "Broken.cs", src)
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
	if len(exts) != 1 || exts[0] != ".cs" {
		t.Errorf("expected [.cs], got %v", exts)
	}
}
