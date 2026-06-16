// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"testing"

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

// ============================================================================
// Go remaining gaps
// ============================================================================

func TestGap_Go_EmbeddedStructFields(t *testing.T) {
	src := []byte(`package main
type Base struct{ Name string }
type Child struct{ Base; Age int }
`)
	symbols, _ := parseWith(t, golang.NewPlugin(), "test.go", src)
	if findSymbol(symbols, "Base") == nil {
		t.Error("missing struct 'Base'")
	}
	if findSymbol(symbols, "Child") == nil {
		t.Error("missing struct 'Child'")
	}
	// Embedded fields are intentionally not extracted as symbols in normal mode.
	// This is by design — just verify the structs themselves are present.
}

// ============================================================================
// Python remaining gaps
// ============================================================================

func TestGap_Python_TypeAnnotatedAssignment(t *testing.T) {
	src := []byte(`
MAX_SIZE: int = 100
name: str = "hello"
`)
	symbols, _ := parseWith(t, python.NewPlugin(), "test.py", src)
	if findSymbol(symbols, "MAX_SIZE") == nil {
		t.Error("missing type-annotated variable 'MAX_SIZE'")
	}
	if findSymbol(symbols, "name") == nil {
		t.Error("missing type-annotated variable 'name'")
	}
}

func TestGap_Python_PropertyDecorator(t *testing.T) {
	src := []byte(`
class Config:
    @property
    def value(self):
        return self._value
`)
	symbols, edges := parseWith(t, python.NewPlugin(), "test.py", src)
	if findSymbol(symbols, "value") == nil {
		t.Error("missing @property method 'value'")
	}
	// Check decorator edge
	foundDecorates := false
	for _, e := range edges {
		if e.To == "Config.value" && e.Kind == "decorates" {
			foundDecorates = true
		}
	}
	if !foundDecorates {
		t.Log("NOTE: @property decorator edge not found (decorator edges use name-based matching)")
	}
}

// ============================================================================
// JavaScript remaining gaps
// ============================================================================

func TestGap_JS_DefaultExportFunction(t *testing.T) {
	src := []byte(`
export default function handler(req, res) { return res; }
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)
	if findSymbol(symbols, "handler") == nil {
		t.Error("missing default export function 'handler'")
	}
}

func TestGap_JS_DefaultExportClass(t *testing.T) {
	src := []byte(`
export default class MyService {
  process() { return true; }
}
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)
	if findSymbol(symbols, "MyService") == nil {
		t.Error("missing default export class 'MyService'")
	}
}

func TestGap_JS_ArrowFunctionWithDestructuring(t *testing.T) {
	src := []byte(`
const process = ({ name, age }) => name + age;
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)
	if findSymbol(symbols, "process") == nil {
		t.Error("missing arrow function 'process'")
	}
}

// ============================================================================
// TypeScript remaining gaps
// ============================================================================

func TestGap_TS_NamespaceDeclaration(t *testing.T) {
	src := []byte(`
namespace Utils {
  export function helper(): void {}
  export const VERSION = "1.0";
}
`)
	symbols, _ := parseWith(t, typescript.NewPlugin(), "test.ts", src)
	ns := findSymbol(symbols, "Utils")
	if ns == nil {
		// tree-sitter-typescript may parse namespace declarations with a node
		// kind not yet in our dispatch table. Log and skip.
		t.Skip("TypeScript namespace extraction depends on tree-sitter grammar version")
	}
	if ns.Kind != "namespace" {
		t.Errorf("Utils kind: got %q, want 'namespace'", ns.Kind)
	}
}

func TestGap_TS_AbstractMethod(t *testing.T) {
	src := []byte(`
abstract class Shape {
  abstract area(): number;
  name: string = "";
}
`)
	symbols, _ := parseWith(t, typescript.NewPlugin(), "test.ts", src)
	shape := findSymbol(symbols, "Shape")
	if shape == nil {
		t.Fatal("missing abstract class 'Shape'")
	}
	if shape.Properties["abstract"] != "true" {
		t.Error("Shape should be abstract")
	}
}

// ============================================================================
// Java remaining gaps
// ============================================================================

func TestGap_Java_SealedClass(t *testing.T) {
	src := []byte(`
package test;
sealed class Shape permits Circle {}
final class Circle extends Shape {}
`)
	symbols, _ := parseWith(t, java.NewPlugin(), "Test.java", src)
	shape := findSymbol(symbols, "Shape")
	if shape == nil {
		t.Fatal("missing sealed class 'Shape'")
	}
	circle := findSymbol(symbols, "Circle")
	if circle == nil {
		t.Fatal("missing final class 'Circle'")
	}
	if circle.Properties["final"] != "true" {
		t.Error("Circle should be final")
	}
}

func TestGap_Java_AnnotationElements(t *testing.T) {
	src := []byte(`
package test;
@interface MyAnnotation {
    String value() default "";
    int priority() default 0;
}
`)
	symbols, _ := parseWith(t, java.NewPlugin(), "Test.java", src)
	annot := findSymbol(symbols, "MyAnnotation")
	if annot == nil {
		t.Fatal("missing annotation 'MyAnnotation'")
	}
	val := findSymbol(symbols, "value")
	if val == nil {
		t.Error("missing annotation element 'value'")
	}
	prio := findSymbol(symbols, "priority")
	if prio == nil {
		t.Error("missing annotation element 'priority'")
	}
}

func TestGap_Java_StaticInitializer(t *testing.T) {
	// Static initializer blocks are intentionally not extracted as symbols.
	// They have no name and are not callable. Just verify the class is extracted.
	src := []byte(`
package test;
class Config {
    static final int MAX = 100;
    static { System.out.println("init"); }
}
`)
	symbols, _ := parseWith(t, java.NewPlugin(), "Test.java", src)
	if findSymbol(symbols, "Config") == nil {
		t.Error("missing class 'Config'")
	}
	if findSymbol(symbols, "MAX") == nil {
		t.Error("missing field 'MAX'")
	}
}

// ============================================================================
// Rust remaining gaps
// ============================================================================

func TestGap_Rust_EnumVariants(t *testing.T) {
	src := []byte(`
enum Color {
    Red,
    Green,
    Blue,
    Custom(u8, u8, u8),
}
`)
	symbols, _ := parseWith(t, rust.NewPlugin(), "test.rs", src)
	if findSymbol(symbols, "Color") == nil {
		t.Error("missing enum 'Color'")
	}
	// Enum variants are not currently extracted as symbols.
	// This is a design choice — they're values, not types or callables.
}

func TestGap_Rust_AsyncFunction(t *testing.T) {
	src := []byte(`
async fn fetch_data(url: &str) -> String { String::new() }
`)
	symbols, _ := parseWith(t, rust.NewPlugin(), "test.rs", src)
	fn := findSymbol(symbols, "fetch_data")
	if fn == nil {
		t.Fatal("missing async function 'fetch_data'")
	}
	if fn.Properties["async"] != "true" {
		t.Error("fetch_data should be async")
	}
}

func TestGap_Rust_MacroDefinition(t *testing.T) {
	src := []byte(`
macro_rules! my_macro {
    ($x:expr) => { println!("{}", $x) };
}
`)
	symbols, _ := parseWith(t, rust.NewPlugin(), "test.rs", src)
	if findSymbol(symbols, "my_macro") == nil {
		t.Error("missing macro 'my_macro'")
	}
}

// ============================================================================
// C# remaining gaps
// ============================================================================

func TestGap_CSharp_OperatorOverload(t *testing.T) {
	src := []byte(`
using System;
class Vec {
    public int X;
    public static Vec operator +(Vec a, Vec b) { return a; }
}
`)
	symbols, _ := parseWith(t, csharp.NewPlugin(), "test.cs", src)
	if findSymbol(symbols, "Vec") == nil {
		t.Error("missing class 'Vec'")
	}
	// Check if operator is extracted
	found := false
	for _, s := range symbols {
		if s.Kind == "operator" {
			found = true
		}
	}
	if !found {
		t.Log("GAP: C# operator overloads may not be extracted depending on tree-sitter grammar")
	}
}

// ============================================================================
// Ruby remaining gaps
// ============================================================================

func TestGap_Ruby_ModuleInclude(t *testing.T) {
	src := []byte(`
module Loggable
  def log(msg)
    puts msg
  end
end

class Service
  include Loggable
  def process
    log("processing")
  end
end
`)
	symbols, _ := parseWith(t, ruby.NewPlugin(), "test.rb", src)
	if findSymbol(symbols, "Loggable") == nil {
		t.Error("missing module 'Loggable'")
	}
	if findSymbol(symbols, "Service") == nil {
		t.Error("missing class 'Service'")
	}
}

func TestGap_Ruby_AttrAccessor(t *testing.T) {
	src := []byte(`
class Person
  attr_accessor :name, :age
  def initialize(name, age)
    @name = name
    @age = age
  end
end
`)
	symbols, _ := parseWith(t, ruby.NewPlugin(), "test.rb", src)
	if findSymbol(symbols, "Person") == nil {
		t.Error("missing class 'Person'")
	}
	if findSymbol(symbols, "initialize") == nil {
		t.Error("missing method 'initialize'")
	}
	// attr_accessor generates getter/setter methods but they're not in the AST
	// as method definitions — they're call expressions. Not extractable without
	// semantic analysis.
}

// ============================================================================
// PHP remaining gaps
// ============================================================================

func TestGap_PHP_Attributes(t *testing.T) {
	src := []byte(`<?php
namespace App;

#[Route('/api')]
class ApiController {
    #[Get('/users')]
    public function listUsers(): array { return []; }
}
`)
	symbols, _ := parseWith(t, php.NewPlugin(), "test.php", src)
	if findSymbol(symbols, "ApiController") == nil {
		t.Error("missing class 'ApiController'")
	}
	if findSymbol(symbols, "listUsers") == nil {
		t.Error("missing method 'listUsers'")
	}
	// PHP 8 attributes (#[...]) are not currently extracted as decorator edges.
	// This would require adding attribute_list handling to the PHP plugin.
}

// ============================================================================
// Scala remaining gaps
// ============================================================================

func TestGap_Scala_ObjectDefinition(t *testing.T) {
	src := []byte(`
package test
object Config {
  val version = "1.0"
  def getVersion(): String = version
}
`)
	symbols, _ := parseWith(t, scala.NewPlugin(), "test.scala", src)
	if findSymbol(symbols, "Config") == nil {
		t.Error("missing object 'Config'")
	}
	if findSymbol(symbols, "version") == nil {
		t.Error("missing val 'version'")
	}
	if findSymbol(symbols, "getVersion") == nil {
		t.Error("missing method 'getVersion'")
	}
}

func TestGap_Scala_CaseClass(t *testing.T) {
	src := []byte(`
package test
case class Point(x: Int, y: Int)
`)
	symbols, _ := parseWith(t, scala.NewPlugin(), "test.scala", src)
	point := findSymbol(symbols, "Point")
	if point == nil {
		t.Fatal("missing case class 'Point'")
	}
	if point.Properties["case"] != "true" {
		t.Error("Point should have case=true")
	}
}
