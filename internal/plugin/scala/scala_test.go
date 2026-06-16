// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scala

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
package com.example

abstract class Foo extends Bar with Baz {
  def hello(x: Int): String = "hi"
  override def toString(): String = "Foo"
  val name: String = "foo"
  var count: Int = 0
}

case class Point(x: Int, y: Int)

trait Drawable {
  def draw(): Unit
}

object Main {
  def main(args: Array[String]): Unit = {
    println("hello")
  }
}

type MyAlias = List[Int]
`)

	symbols, _, err := parseSource(t, "App.scala", src)
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
		"Foo":         {plugin.CategoryType, "class"},
		"hello":       {plugin.CategoryCallable, "method"},
		"toString":    {plugin.CategoryCallable, "method"},
		"name":        {plugin.CategoryValue, "val"},
		"count":       {plugin.CategoryValue, "variable"},
		"Point":       {plugin.CategoryType, "class"},
		"Drawable":    {plugin.CategoryType, "trait"},
		"draw":        {plugin.CategoryCallable, "method"},
		"Main":        {plugin.CategoryType, "object"},
		"main":        {plugin.CategoryCallable, "method"},
		"MyAlias":     {plugin.CategoryType, "type_alias"},
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

func TestParseSource_CaseClassDetection(t *testing.T) {
	src := []byte(`
case class Point(x: Int, y: Int)
class Regular(name: String)
`)
	symbols, _, err := parseSource(t, "Point.scala", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	if props["Point"]["case"] != "true" {
		t.Errorf("Point should have case=true, got %q", props["Point"]["case"])
	}
	if props["Regular"]["case"] == "true" {
		t.Error("Regular should not have case=true")
	}
}

func TestParseSource_ModifierExtraction(t *testing.T) {
	src := []byte(`
abstract sealed class Foo {
  override final def bar(): Unit = {}
  private lazy val secret: String = "shh"
  protected def helper(): Unit = {}
}
`)
	symbols, _, err := parseSource(t, "Foo.scala", src)
	if err != nil {
		t.Fatal(err)
	}

	props := make(map[string]map[string]string)
	for _, s := range symbols {
		props[s.Name] = s.Properties
	}

	// Class modifiers.
	if props["Foo"]["abstract"] != "true" {
		t.Error("Foo should have abstract=true")
	}
	if props["Foo"]["sealed"] != "true" {
		t.Error("Foo should have sealed=true")
	}

	// Override final method.
	if props["bar"]["override"] != "true" {
		t.Error("bar should have override=true")
	}
	if props["bar"]["final"] != "true" {
		t.Error("bar should have final=true")
	}

	// Private lazy val.
	if props["secret"]["visibility"] != "private" {
		t.Errorf("secret visibility: got %q, want %q", props["secret"]["visibility"], "private")
	}
	if props["secret"]["lazy"] != "true" {
		t.Error("secret should have lazy=true")
	}

	// Protected method.
	if props["helper"]["visibility"] != "protected" {
		t.Errorf("helper visibility: got %q, want %q", props["helper"]["visibility"], "protected")
	}
}

func TestParseSource_InheritanceEdges(t *testing.T) {
	src := []byte(`
class Child extends Parent with Serializable with Comparable {
}
`)
	_, edges, err := parseSource(t, "Child.scala", src)
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
class MyClass {
  val x: Int = 1
  def doSomething(): Unit = {}
  var y: String = "hi"
}
`)
	_, edges, err := parseSource(t, "MyClass.scala", src)
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
	want := []string{"MyClass.doSomething", "MyClass.x", "MyClass.y"}
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
object App {
  def run(): Unit = {
    println("hello")
    helper()
  }
  def helper(): Unit = {}
}
`)
	_, edges, err := parseSource(t, "App.scala", src)
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
	foundHelper := false
	foundPrintln := false
	for _, ct := range callTargets {
		if ct == "helper" {
			foundHelper = true
		}
		if ct == "println" {
			foundPrintln = true
		}
	}
	if !foundHelper {
		t.Errorf("expected call edge from run to helper, got targets: %v", callTargets)
	}
	if !foundPrintln {
		t.Errorf("expected call edge from run to println, got targets: %v", callTargets)
	}
}

func TestParseSource_OverrideDetection(t *testing.T) {
	src := []byte(`
class Child extends Parent {
  override def doWork(): Unit = {}
}
`)
	_, edges, err := parseSource(t, "Child.scala", src)
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

func TestParseSource_TraitDeclaration(t *testing.T) {
	src := []byte(`
sealed trait Animal {
  def speak(): String
  def name: String
}
`)
	symbols, edges, err := parseSource(t, "Animal.scala", src)
	if err != nil {
		t.Fatal(err)
	}

	foundTrait := false
	for _, s := range symbols {
		if s.Name == "Animal" && s.Kind == "trait" {
			foundTrait = true
			if s.Properties["sealed"] != "true" {
				t.Error("Animal should have sealed=true")
			}
		}
	}
	if !foundTrait {
		t.Error("missing trait symbol 'Animal'")
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Animal" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Animal.name", "Animal.speak"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_ObjectExtraction(t *testing.T) {
	src := []byte(`
object Singleton {
  val instance: String = "one"
  def get(): String = instance
}
`)
	symbols, _, err := parseSource(t, "Singleton.scala", src)
	if err != nil {
		t.Fatal(err)
	}

	foundObject := false
	for _, s := range symbols {
		if s.Name == "Singleton" && s.Kind == "object" && s.Category == plugin.CategoryType {
			foundObject = true
		}
	}
	if !foundObject {
		t.Error("missing object symbol 'Singleton'")
	}
}

func TestParseSource_ValVarExtraction(t *testing.T) {
	src := []byte(`
val topVal: Int = 42
var topVar: String = "hello"
`)
	symbols, _, err := parseSource(t, "vals.scala", src)
	if err != nil {
		t.Fatal(err)
	}

	foundVal := false
	foundVar := false
	for _, s := range symbols {
		if s.Name == "topVal" && s.Kind == "val" && s.Category == plugin.CategoryValue {
			foundVal = true
		}
		if s.Name == "topVar" && s.Kind == "variable" && s.Category == plugin.CategoryValue {
			foundVar = true
		}
	}
	if !foundVal {
		t.Error("missing val symbol 'topVal'")
	}
	if !foundVar {
		t.Error("missing var symbol 'topVar'")
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`class Broken {
    def bad(: Unit = {
    }
}`)
	_, _, err := parseSource(t, "Broken.scala", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "Broken.scala") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`class Broken {
    def bad(: Unit = {
    }
}`)
	symbols, _, err := parseSource(t, "Broken.scala", src)
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
		t.Fatalf("expected 2 extensions, got %v", exts)
	}
	sort.Strings(exts)
	if exts[0] != ".sc" || exts[1] != ".scala" {
		t.Errorf("expected [.sc, .scala], got %v", exts)
	}
}
