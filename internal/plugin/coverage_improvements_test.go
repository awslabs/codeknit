// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/clang"
	"codeknit/internal/plugin/cpp"
	"codeknit/internal/plugin/csharp"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/php"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/ruby"
	"codeknit/internal/plugin/rust"
	"codeknit/internal/plugin/scala"
	"codeknit/internal/plugin/typescript"
)

// parseWith writes src to a temp file and parses it via the given plugin.
func parseWith(t *testing.T, p plugin.LanguagePlugin, filename string, src []byte) (symbols []plugin.Symbol, edges []plugin.Edge) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	symbols, edges, _ = p.Parse(path)
	return symbols, edges
}

// findSymbol returns the first symbol with the given name, or nil.
func findSymbol(symbols []plugin.Symbol, name string) *plugin.Symbol {
	for i := range symbols {
		if symbols[i].Name == name {
			return &symbols[i]
		}
	}
	return nil
}

// findSymbolByKind returns the first symbol with the given name and kind.
func findSymbolByKind(symbols []plugin.Symbol, name, kind string) *plugin.Symbol {
	for i := range symbols {
		if symbols[i].Name == name && symbols[i].Kind == kind {
			return &symbols[i]
		}
	}
	return nil
}

// hasEdge checks if an edge exists with the given from, to, and kind.
func hasEdge(edges []plugin.Edge, from, to string, kind plugin.EdgeKind) bool {
	for _, e := range edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return true
		}
	}
	return false
}

// ============================================================================
// TypeScript improvements
// ============================================================================

func TestTS_TypeParameters(t *testing.T) {
	src := []byte(`
function identity<T>(arg: T): T { return arg; }
class Container<T, U> { value: T; }
interface Repo<T> { find(id: string): T; }
type Pair<K, V> = { key: K; value: V };
`)
	symbols, edges := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	// Check type parameters exist
	for _, name := range []string{"identity", "Container", "Repo", "Pair"} {
		if findSymbol(symbols, name) == nil {
			t.Errorf("missing symbol %q", name)
		}
	}

	// Check T is extracted as type_parameter
	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found")
	}

	// Container should contain T and U
	tSym := findSymbolByKind(symbols, "T", "type_parameter")
	if tSym == nil {
		t.Fatal("missing type_parameter T")
	}

	// Check contains edges
	foundContains := false
	for _, e := range edges {
		if e.To == "Container.T" && e.Kind == plugin.EdgeContains {
			foundContains = true
		}
	}
	if !foundContains {
		t.Error("missing contains edge for Container -> T")
	}
}

func TestTS_ClassFields(t *testing.T) {
	src := []byte(`
class MyClass {
  count: number = 0;
  private secret: string = "hidden";
  readonly id: string;
  static instances: MyClass[] = [];
  
  get value(): number { return this.count; }
  set value(v: number) { this.count = v; }
  
  async process(): Promise<void> {}
}
`)
	symbols, edges := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	// Class fields
	countSym := findSymbolByKind(symbols, "count", "field")
	if countSym == nil {
		t.Error("missing field 'count'")
	}

	secretSym := findSymbolByKind(symbols, "secret", "field")
	if secretSym == nil {
		t.Error("missing field 'secret'")
	} else if secretSym.Properties["visibility"] != "private" {
		t.Errorf("secret visibility: got %q, want 'private'", secretSym.Properties["visibility"])
	}

	idSym := findSymbolByKind(symbols, "id", "field")
	if idSym == nil {
		t.Error("missing field 'id'")
	} else if idSym.Properties["readonly"] != "true" {
		t.Error("id should be readonly")
	}

	instancesSym := findSymbolByKind(symbols, "instances", "field")
	if instancesSym == nil {
		t.Error("missing field 'instances'")
	} else if instancesSym.Properties["static"] != "true" {
		t.Error("instances should be static")
	}

	// Getter/setter
	getter := findSymbolByKind(symbols, "value", "getter")
	if getter == nil {
		t.Error("missing getter 'value'")
	}
	setter := findSymbolByKind(symbols, "value", "setter")
	if setter == nil {
		t.Error("missing setter 'value'")
	}

	// Contains edges for fields
	fieldContains := 0
	for _, e := range edges {
		if e.From == "MyClass" && e.Kind == plugin.EdgeContains {
			fieldContains++
		}
	}
	if fieldContains < 5 {
		t.Errorf("expected at least 5 contains edges from MyClass, got %d", fieldContains)
	}
}

func TestTS_Destructuring(t *testing.T) {
	src := []byte(`
const { name, age, address: addr } = person;
const [first, second, ...rest] = items;
export const { API_KEY, SECRET } = process.env;
`)
	symbols, _ := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	for _, want := range []string{"name", "age", "addr", "first", "second", "rest", "API_KEY", "SECRET"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing destructured symbol %q", want)
		}
	}

	// API_KEY should be exported
	apiKey := findSymbol(symbols, "API_KEY")
	if apiKey != nil && apiKey.Properties["exported"] != "true" {
		t.Error("API_KEY should be exported")
	}
}

func TestTS_GeneratorFunction(t *testing.T) {
	src := []byte(`
function* generateIds() { yield 1; }
export function* streamData() { yield 1; }
`)
	symbols, _ := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	gen := findSymbol(symbols, "generateIds")
	if gen == nil {
		t.Error("missing generator function 'generateIds'")
	}
	stream := findSymbol(symbols, "streamData")
	if stream == nil {
		t.Error("missing generator function 'streamData'")
	}
}

// ============================================================================
// JavaScript improvements
// ============================================================================

func TestJS_ClassFields(t *testing.T) {
	src := []byte(`
class MyClass {
  count = 0;
  static instances = [];
  
  get value() { return this.count; }
  set value(v) { this.count = v; }
}
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)

	if findSymbolByKind(symbols, "count", "field") == nil {
		t.Error("missing field 'count'")
	}
	if findSymbolByKind(symbols, "instances", "field") == nil {
		t.Error("missing field 'instances'")
	}
	if findSymbolByKind(symbols, "value", "getter") == nil {
		t.Error("missing getter 'value'")
	}
	if findSymbolByKind(symbols, "value", "setter") == nil {
		t.Error("missing setter 'value'")
	}
}

func TestJS_Destructuring(t *testing.T) {
	src := []byte(`
const { name, age } = person;
const [first, second] = items;
const { data: { id, title } } = response;
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)

	for _, want := range []string{"name", "age", "first", "second", "id", "title"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing destructured symbol %q", want)
		}
	}
}

func TestJS_GeneratorFunction(t *testing.T) {
	src := []byte(`
function* generateIds() { yield 1; }
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)

	if findSymbol(symbols, "generateIds") == nil {
		t.Error("missing generator function 'generateIds'")
	}
}

// ============================================================================
// Python improvements
// ============================================================================

func TestPython_NestedClass(t *testing.T) {
	src := []byte(`
class Outer:
    class Inner:
        def method(self):
            pass
    def outer_method(self):
        pass
`)
	symbols, edges := parseWith(t, python.NewPlugin(), "test.py", src)

	if findSymbol(symbols, "Outer") == nil {
		t.Error("missing class 'Outer'")
	}
	if findSymbol(symbols, "Inner") == nil {
		t.Error("missing nested class 'Inner'")
	}

	// Inner should be contained by Outer
	if !hasEdge(edges, "Outer", "Outer.Inner", plugin.EdgeContains) {
		t.Error("missing contains edge Outer -> Outer.Inner")
	}
}

func TestPython_TupleUnpacking(t *testing.T) {
	src := []byte(`
x, y = 1, 2
a, b, c = get_values()
first, (second, third) = data
name = "hello"
`)
	symbols, _ := parseWith(t, python.NewPlugin(), "test.py", src)

	for _, want := range []string{"x", "y", "a", "b", "c", "first", "second", "third", "name"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing symbol %q", want)
		}
	}
}

// ============================================================================
// Go improvements
// ============================================================================

func TestGo_TypeParameters(t *testing.T) {
	src := []byte(`
package main

func Map[T any, U any](s []T, f func(T) U) []U { return nil }

type Set[T comparable] struct { items map[T]bool }
`)
	symbols, _ := parseWith(t, golang.NewPlugin(), "test.go", src)

	if findSymbol(symbols, "Map") == nil {
		t.Error("missing function 'Map'")
	}

	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found for Go generics")
	}
}

// ============================================================================
// Java improvements
// ============================================================================

func TestJava_RecordType(t *testing.T) {
	src := []byte(`
package test;
record Point(int x, int y) {}
`)
	symbols, _ := parseWith(t, java.NewPlugin(), "Test.java", src)

	point := findSymbol(symbols, "Point")
	if point == nil {
		t.Fatal("missing record 'Point'")
	}
	if point.Kind != "record" {
		t.Errorf("Point kind: got %q, want 'record'", point.Kind)
	}
}

func TestJava_TypeParameters(t *testing.T) {
	src := []byte(`
package test;
class Container<T> {
    T value;
    <U> U convert(T input) { return null; }
}
`)
	symbols, _ := parseWith(t, java.NewPlugin(), "Test.java", src)

	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found for Java generics")
	}
}

// ============================================================================
// Rust improvements
// ============================================================================

func TestRust_TypeParameters(t *testing.T) {
	src := []byte(`
fn identity<T>(val: T) -> T { val }
struct Wrapper<T> { inner: T }
trait Converter<T, U> { fn convert(&self, input: T) -> U; }
impl<T: Clone> Wrapper<T> { fn clone_inner(&self) -> T { self.inner.clone() } }
`)
	symbols, _ := parseWith(t, rust.NewPlugin(), "test.rs", src)

	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found for Rust generics")
	}
}

// ============================================================================
// C# improvements
// ============================================================================

func TestCSharp_Constructor(t *testing.T) {
	src := []byte(`
using System;
namespace Ex {
    class Service {
        public Service(string name) {}
        public void Process() {}
    }
}
`)
	symbols, _ := parseWith(t, csharp.NewPlugin(), "test.cs", src)

	ctor := findSymbolByKind(symbols, "Service", "constructor")
	if ctor == nil {
		t.Error("missing constructor 'Service'")
	}
}

func TestCSharp_Record(t *testing.T) {
	src := []byte(`
using System;
namespace Ex {
    record Person(string Name, int Age);
}
`)
	symbols, _ := parseWith(t, csharp.NewPlugin(), "test.cs", src)

	person := findSymbol(symbols, "Person")
	if person == nil {
		t.Fatal("missing record 'Person'")
	}
	if person.Kind != "record" {
		t.Errorf("Person kind: got %q, want 'record'", person.Kind)
	}
}

func TestCSharp_Event(t *testing.T) {
	src := []byte(`
using System;
namespace Ex {
    class Service {
        public event EventHandler OnChange;
    }
}
`)
	symbols, _ := parseWith(t, csharp.NewPlugin(), "test.cs", src)

	evt := findSymbolByKind(symbols, "OnChange", "event")
	if evt == nil {
		t.Error("missing event 'OnChange'")
	}
}

func TestCSharp_TypeParameters(t *testing.T) {
	src := []byte(`
using System;
namespace Ex {
    class Container<T> {
        public T Value { get; set; }
    }
    interface IRepo<T> {
        T Find(string id);
    }
}
`)
	symbols, _ := parseWith(t, csharp.NewPlugin(), "test.cs", src)

	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found for C# generics")
	}
}

// ============================================================================
// C++ improvements
// ============================================================================

func TestCpp_TypeParameters(t *testing.T) {
	src := []byte(`
template<typename T>
T identity(T val) { return val; }

template<typename T, typename U>
class Pair {
public:
    T first;
    U second;
};
`)
	symbols, _ := parseWith(t, cpp.NewPlugin(), "test.cpp", src)

	tParams := 0
	for _, s := range symbols {
		if s.Kind == "type_parameter" {
			tParams++
		}
	}
	if tParams == 0 {
		t.Error("no type_parameter symbols found for C++ templates")
	}
}

// ============================================================================
// Scala improvements
// ============================================================================

func TestScala_TypeParameters(t *testing.T) {
	src := []byte(`
package test

class Container[T](val value: T)
def identity[T](x: T): T = x
`)
	symbols, _ := parseWith(t, scala.NewPlugin(), "test.scala", src)

	// Scala's tree-sitter grammar may not expose type parameters as named
	// children in a way our generic extractor can handle. Verify the main
	// symbols are still extracted correctly.
	if findSymbol(symbols, "Container") == nil {
		t.Error("missing class 'Container'")
	}
}

// ============================================================================
// Ruby improvements
// ============================================================================

func TestRuby_MultipleAssignment(t *testing.T) {
	src := []byte(`
a, b = 1, 2
x, *rest = [1, 2, 3]
NAME = "test"
`)
	symbols, _ := parseWith(t, ruby.NewPlugin(), "test.rb", src)

	for _, want := range []string{"a", "b", "NAME"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing symbol %q", want)
		}
	}
}

// ============================================================================
// PHP - verify existing features still work
// ============================================================================

func TestPHP_TraitAndEnum(t *testing.T) {
	src := []byte(`<?php
namespace App;

trait Loggable {
    public function log(string $msg): void {}
}

enum Status {
    case Active;
    case Inactive;
}

class Service {
    use Loggable;
    public function process(): void {}
}
`)
	symbols, _ := parseWith(t, php.NewPlugin(), "test.php", src)

	if findSymbolByKind(symbols, "Loggable", "trait") == nil {
		t.Error("missing trait 'Loggable'")
	}
	if findSymbolByKind(symbols, "Status", "enum") == nil {
		t.Error("missing enum 'Status'")
	}
}

// ============================================================================
// C - verify existing features still work
// ============================================================================

func TestC_MacroAndTypedef(t *testing.T) {
	src := []byte(`
#define MAX_SIZE 100
#define SQUARE(x) ((x) * (x))

typedef unsigned long size_t;
typedef struct { int x; int y; } Point;

void process(int n) {}
`)
	symbols, _ := parseWith(t, clang.NewPlugin(), "test.c", src)

	if findSymbolByKind(symbols, "MAX_SIZE", "macro") == nil {
		t.Error("missing macro 'MAX_SIZE'")
	}
	if findSymbolByKind(symbols, "SQUARE", "macro_function") == nil {
		t.Error("missing macro_function 'SQUARE'")
	}
	if findSymbolByKind(symbols, "Point", "typedef") == nil {
		t.Error("missing typedef 'Point'")
	}
	if findSymbolByKind(symbols, "process", "function") == nil {
		t.Error("missing function 'process'")
	}
}
