// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package php

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
	src := []byte(`<?php
namespace App\Models;

class User extends BaseModel implements Serializable {
    public string $name;
    private int $age;

    public function __construct(string $name, int $age) {
        $this->name = $name;
        $this->age = $age;
    }

    public function getName(): string {
        return $this->name;
    }

    public static function create(string $name): self {
        return new self($name, 0);
    }
}

function helper(): void {
    echo "hello";
}
`)

	symbols, _, err := parseSource(t, "User.php", src)
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
		"App\\Models": {plugin.CategoryModule, "namespace"},
		"User":        {plugin.CategoryType, "class"},
		"getName":     {plugin.CategoryCallable, "method"},
		"create":      {plugin.CategoryCallable, "method"},
		"__construct": {plugin.CategoryCallable, "method"},
		"helper":      {plugin.CategoryCallable, "function"},
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

	// Verify property symbols exist.
	foundName := false
	foundAge := false
	for _, s := range symbols {
		if s.Kind == "property" && s.Name == "name" {
			foundName = true
		}
		if s.Kind == "property" && s.Name == "age" {
			foundAge = true
		}
	}
	if !foundName {
		t.Error("missing property symbol 'name'")
	}
	if !foundAge {
		t.Error("missing property symbol 'age'")
	}
}

func TestParseSource_ModifierExtraction(t *testing.T) {
	src := []byte(`<?php
abstract class Foo {
    public static function staticMethod(): void {}
    private int $count = 0;
    protected abstract function abstractMethod(): void;
    final public function finalMethod(): void {}
}
`)
	symbols, _, err := parseSource(t, "Foo.php", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	// Class modifiers.
	if props["Foo"]["abstract"] != "true" {
		t.Errorf("Foo should have abstract=true")
	}

	// Static method.
	if props["staticMethod"]["static"] != "true" {
		t.Errorf("staticMethod should have static=true")
	}
	if props["staticMethod"]["visibility"] != "public" {
		t.Errorf("staticMethod visibility: got %q, want %q", props["staticMethod"]["visibility"], "public")
	}

	// Private property.
	if props["count"]["visibility"] != "private" {
		t.Errorf("count visibility: got %q, want %q", props["count"]["visibility"], "private")
	}

	// Abstract method.
	if props["abstractMethod"]["abstract"] != "true" {
		t.Errorf("abstractMethod should have abstract=true")
	}
	if props["abstractMethod"]["visibility"] != "protected" {
		t.Errorf("abstractMethod visibility: got %q, want %q", props["abstractMethod"]["visibility"], "protected")
	}

	// Final method.
	if props["finalMethod"]["final"] != "true" {
		t.Errorf("finalMethod should have final=true")
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`<?php
class Child extends Parent implements Serializable, JsonSerializable {
}
`)
	_, edges, err := parseSource(t, "Child.php", src)
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
	wantImpl := []string{"JsonSerializable", "Serializable"}
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
	src := []byte(`<?php
class MyClass {
    private int $x;
    public function doSomething(): void {}
    public function __construct() {}
}
`)
	_, edges, err := parseSource(t, "MyClass.php", src)
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
	want := []string{"MyClass.__construct", "MyClass.doSomething", "MyClass.x"}
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
	src := []byte(`<?php
class App {
    public function run(): void {
        echo strtoupper("hello");
        $this->helper();
    }
    private function helper(): void {}
}
`)
	_, edges, err := parseSource(t, "App.php", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "App.run" {
			callTargets = append(callTargets, e.To)
		}
	}
	if len(callTargets) == 0 {
		t.Fatal("expected call edges from run")
	}
	// Should contain at least "strtoupper".
	foundStrtoupper := false
	for _, ct := range callTargets {
		if ct == "strtoupper" {
			foundStrtoupper = true
		}
	}
	if !foundStrtoupper {
		t.Errorf("expected call edge from run to strtoupper, got targets: %v", callTargets)
	}
}

func TestParseSource_TraitExtraction(t *testing.T) {
	src := []byte(`<?php
trait Loggable {
    public function log(string $msg): void {
        echo $msg;
    }
}
`)
	symbols, edges, err := parseSource(t, "Loggable.php", src)
	if err != nil {
		t.Fatal(err)
	}

	foundTrait := false
	for _, s := range symbols {
		if s.Name == "Loggable" && s.Kind == "trait" && s.Category == plugin.CategoryType {
			foundTrait = true
		}
	}
	if !foundTrait {
		t.Error("missing trait symbol 'Loggable'")
	}

	// Trait should contain the log method.
	foundContains := false
	for _, e := range edges {
		if e.From == "Loggable" && e.To == "Loggable.log" && e.Kind == plugin.EdgeContains {
			foundContains = true
		}
	}
	if !foundContains {
		t.Error("expected contains edge from Loggable to log")
	}
}

func TestParseSource_InterfaceDeclaration(t *testing.T) {
	src := []byte(`<?php
interface Repository {
    public function find(int $id): mixed;
    public function save(mixed $entity): void;
}
`)
	symbols, edges, err := parseSource(t, "Repository.php", src)
	if err != nil {
		t.Fatal(err)
	}

	foundInterface := false
	for _, s := range symbols {
		if s.Name == "Repository" && s.Kind == "interface" {
			foundInterface = true
		}
	}
	if !foundInterface {
		t.Error("missing interface symbol 'Repository'")
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Repository" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Repository.find", "Repository.save"}
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
	src := []byte(`<?php
enum Color {
    case Red;
    case Green;
    case Blue;
}
`)
	symbols, _, err := parseSource(t, "Color.php", src)
	if err != nil {
		t.Fatal(err)
	}

	foundEnum := false
	for _, s := range symbols {
		if s.Name == "Color" && s.Kind == "enum" && s.Category == plugin.CategoryType {
			foundEnum = true
		}
	}
	if !foundEnum {
		t.Error("missing enum symbol 'Color'")
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`<?php
class Broken {
    public function bad( {
    }
}`)
	_, _, err := parseSource(t, "Broken.php", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "Broken.php") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`<?php
class Broken {
    public function bad( {
    }
}`)
	symbols, _, err := parseSource(t, "Broken.php", src)
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
	if len(exts) != 1 || exts[0] != ".php" {
		t.Errorf("expected [.php], got %v", exts)
	}
}
