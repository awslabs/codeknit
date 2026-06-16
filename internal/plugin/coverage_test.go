// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

// TestLanguageFeatureCoverage verifies that each language plugin extracts
// symbols for all major language constructs. If a construct is silently
// dropped (e.g., abstract classes in TypeScript), this test catches it.
//
// Each test case provides source code exercising key constructs and a list
// of symbol names that MUST appear in the output. Missing symbols indicate
// a gap in the dispatch table or extraction logic.

import (
	"errors"
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

type coverageCase struct {
	plugin   plugin.LanguagePlugin
	expected map[string][2]string
	name     string
	filePath string
	src      []byte
}

func coverageCases() []coverageCase {
	return []coverageCase{
		{
			name:     "TypeScript",
			filePath: "test.ts",
			plugin:   typescript.NewPlugin(),
			src: []byte(`
export function regularFunc(): void {}
export const arrowFunc = (x: number) => x;
export class RegularClass {
  method(): void {}
}
export abstract class AbstractClass {
  abstract abstractMethod(): void;
  concreteMethod(): void {}
}
export interface MyInterface {
  field: string;
}
export type MyTypeAlias = string | number;
export enum MyEnum { A, B }
export const MY_CONST = { key: "value" };
let myVar = 0;
`),
			expected: map[string][2]string{
				"regularFunc":    {"callable", "function"},
				"arrowFunc":      {"callable", "arrow_function"},
				"RegularClass":   {"type", "class"},
				"method":         {"callable", "method"},
				"AbstractClass":  {"type", "class"},
				"concreteMethod": {"callable", "method"},
				"MyInterface":    {"type", "interface"},
				"MyTypeAlias":    {"type", "type_alias"},
				"MyEnum":         {"type", "enum"},
				"MY_CONST":       {"value", "exported_constant"},
				"myVar":          {"value", "variable"},
			},
		},
		{
			name:     "JavaScript",
			filePath: "test.js",
			plugin:   javascript.NewPlugin(),
			src: []byte(`
export function regularFunc() {}
export const arrowFunc = (x) => x;
export class RegularClass {
  method() {}
}
export const MY_CONST = { key: "value" };
let myVar = 0;
`),
			expected: map[string][2]string{
				"regularFunc":  {"callable", "function"},
				"arrowFunc":    {"callable", "arrow_function"},
				"RegularClass": {"type", "class"},
				"method":       {"callable", "method"},
				"MY_CONST":     {"value", "exported_constant"},
				"myVar":        {"value", "variable"},
			},
		},
		{
			name:     "Go",
			filePath: "test.go",
			plugin:   golang.NewPlugin(),
			src: []byte(`
package main

import "fmt"

func regularFunc() {}

type MyStruct struct {
	Field string
}

func (s *MyStruct) Method() {}

type MyInterface interface {
	DoSomething()
}

type MyAlias = int

var myVar = 0
const myConst = 42
`),
			expected: map[string][2]string{
				"main":        {"module", "package"},
				"regularFunc": {"callable", "function"},
				"MyStruct":    {"type", "struct"},
				"Method":      {"callable", "method"},
				"MyInterface": {"type", "interface"},
				"MyAlias":     {"type", "type_alias"},
				"myVar":       {"value", "variable"},
				"myConst":     {"value", "constant"},
			},
		},
		{
			name:     "Python",
			filePath: "test.py",
			plugin:   python.NewPlugin(),
			src: []byte(`
def regular_func():
    pass

async def async_func():
    pass

class MyClass:
    def method(self):
        pass

    @staticmethod
    def static_method():
        pass

class ChildClass(MyClass):
    pass

my_var = 42
`),
			expected: map[string][2]string{
				"regular_func":  {"callable", "function"},
				"async_func":    {"callable", "function"},
				"MyClass":       {"type", "class"},
				"method":        {"callable", "method"},
				"static_method": {"callable", "method"},
				"ChildClass":    {"type", "class"},
				"my_var":        {"value", "variable"},
			},
		},
		{
			name:     "Rust",
			filePath: "test.rs",
			plugin:   rust.NewPlugin(),
			src: []byte(`
fn regular_func() {}

pub struct MyStruct {
    pub field: String,
}

impl MyStruct {
    pub fn method(&self) {}
}

pub trait MyTrait {
    fn trait_method(&self);
}

impl MyTrait for MyStruct {
    fn trait_method(&self) {}
}

pub enum MyEnum {
    A,
    B,
}

type MyAlias = Vec<String>;

pub const MY_CONST: i32 = 42;
static MY_STATIC: i32 = 0;

mod my_module {}

macro_rules! my_macro {
    () => {};
}
`),
			expected: map[string][2]string{
				"regular_func": {"callable", "function"},
				"MyStruct":     {"type", "struct"},
				"method":       {"callable", "method"},
				"MyTrait":      {"type", "trait"},
				"trait_method": {"callable", "method"},
				"MyEnum":       {"type", "enum"},
				"MyAlias":      {"type", "type_alias"},
				"MY_CONST":     {"value", "constant"},
				"MY_STATIC":    {"value", "variable"},
				"my_module":    {"module", "module"},
				"my_macro":     {"callable", "macro"},
			},
		},
		{
			name:     "Java",
			filePath: "Test.java",
			plugin:   java.NewPlugin(),
			src: []byte(`
package com.example;

import java.util.List;

public class MyClass {
    private int field;

    public void method() {}

    public MyClass(int field) {
        this.field = field;
    }
}

interface MyInterface {
    void doSomething();
}

enum MyEnum {
    A, B
}

abstract class AbstractClass {
    abstract void abstractMethod();
    void concreteMethod() {}
}
`),
			expected: map[string][2]string{
				"com.example":    {"module", "package"},
				"MyClass":        {"type", "class"},
				"field":          {"value", "field"},
				"method":         {"callable", "method"},
				"MyInterface":    {"type", "interface"},
				"doSomething":    {"callable", "method"},
				"MyEnum":         {"type", "enum"},
				"AbstractClass":  {"type", "class"},
				"abstractMethod": {"callable", "method"},
				"concreteMethod": {"callable", "method"},
			},
		},
		{
			name:     "C#",
			filePath: "Test.cs",
			plugin:   csharp.NewPlugin(),
			src: []byte(`
namespace MyApp {
    public class MyClass {
        private int field;
        public void Method() {}
    }

    public interface IService {
        void DoWork();
    }

    public struct MyStruct {
        public int X;
    }

    public enum MyEnum {
        A, B
    }

    public abstract class AbstractClass {
        public abstract void AbstractMethod();
        public void ConcreteMethod() {}
    }

    public delegate void MyDelegate(int x);
}
`),
			expected: map[string][2]string{
				"MyApp":          {"module", "namespace"},
				"MyClass":        {"type", "class"},
				"field":          {"value", "field"},
				"Method":         {"callable", "method"},
				"IService":       {"type", "interface"},
				"DoWork":         {"callable", "method"},
				"MyStruct":       {"type", "struct"},
				"MyEnum":         {"type", "enum"},
				"AbstractClass":  {"type", "class"},
				"AbstractMethod": {"callable", "method"},
				"ConcreteMethod": {"callable", "method"},
				"MyDelegate":     {"type", "delegate"},
			},
		},
		{
			name:     "C",
			filePath: "test.c",
			plugin:   clang.NewPlugin(),
			src: []byte(`
#include <stdio.h>

#define MAX_SIZE 100
#define SQUARE(x) ((x) * (x))

void regular_func(void) {}

struct MyStruct {
    int x;
    int y;
};

union MyUnion {
    int i;
    float f;
};

enum MyEnum { A, B, C };

typedef unsigned long size_type;

static int static_var = 0;
extern int extern_var;
`),
			expected: map[string][2]string{
				"regular_func": {"callable", "function"},
				"MyStruct":     {"type", "struct"},
				"MyUnion":      {"type", "union"},
				"MyEnum":       {"type", "enum"},
				"size_type":    {"type", "typedef"},
				"MAX_SIZE":     {"value", "macro"},
				"SQUARE":       {"callable", "macro_function"},
				"static_var":   {"value", "variable"},
				"extern_var":   {"value", "variable"},
			},
		},
		{
			name:     "C++",
			filePath: "test.cpp",
			plugin:   cpp.NewPlugin(),
			src: []byte(`
#include <string>

namespace MyNamespace {

class MyClass {
public:
    virtual void method() {}
    virtual ~MyClass() {}
};

class ChildClass : public MyClass {
public:
    void method() override {}
};

struct MyStruct {
    int x;
};

enum MyEnum { A, B };

template<typename T>
T identity(T x) { return x; }

void regular_func() {}

}
`),
			expected: map[string][2]string{
				"MyNamespace":  {"module", "namespace"},
				"MyClass":      {"type", "class"},
				"method":       {"callable", "method"},
				"ChildClass":   {"type", "class"},
				"MyStruct":     {"type", "struct"},
				"MyEnum":       {"type", "enum"},
				"identity":     {"callable", "function"},
				"regular_func": {"callable", "function"},
			},
		},
		{
			name:     "PHP",
			filePath: "test.php",
			plugin:   php.NewPlugin(),
			src: []byte(`<?php
namespace App\Models;

function regular_func(): void {}

class MyClass {
    private int $field;
    public function method(): void {}
}

interface MyInterface {
    public function doWork(): void;
}

trait MyTrait {
    public function traitMethod(): void {}
}

enum MyEnum {
    case A;
    case B;
}

abstract class AbstractClass {
    abstract public function abstractMethod(): void;
    public function concreteMethod(): void {}
}
`),
			expected: map[string][2]string{
				"App\\Models":    {"module", "namespace"},
				"regular_func":   {"callable", "function"},
				"MyClass":        {"type", "class"},
				"method":         {"callable", "method"},
				"MyInterface":    {"type", "interface"},
				"doWork":         {"callable", "method"},
				"MyTrait":        {"type", "trait"},
				"traitMethod":    {"callable", "method"},
				"MyEnum":         {"type", "enum"},
				"AbstractClass":  {"type", "class"},
				"abstractMethod": {"callable", "method"},
				"concreteMethod": {"callable", "method"},
			},
		},
		{
			name:     "Ruby",
			filePath: "test.rb",
			plugin:   ruby.NewPlugin(),
			src: []byte(`
def regular_func
end

class MyClass
  def method
  end

  def self.static_method
  end

  private

  def private_method
  end
end

class ChildClass < MyClass
  def child_method
  end
end

module MyModule
  def module_method
  end
end

MY_CONST = 42
my_var = 0
`),
			expected: map[string][2]string{
				"regular_func":   {"callable", "function"},
				"MyClass":        {"type", "class"},
				"method":         {"callable", "method"},
				"static_method":  {"callable", "method"},
				"private_method": {"callable", "method"},
				"ChildClass":     {"type", "class"},
				"child_method":   {"callable", "method"},
				"MyModule":       {"module", "module"},
				"module_method":  {"callable", "method"},
				"MY_CONST":       {"value", "constant"},
				"my_var":         {"value", "variable"},
			},
		},
		{
			name:     "Scala",
			filePath: "test.scala",
			plugin:   scala.NewPlugin(),
			src: []byte(`
package com.example

class MyClass {
  def method(): Unit = {}
}

abstract class AbstractClass {
  def abstractMethod(): Unit
  def concreteMethod(): Unit = {}
}

case class CaseClass(name: String, age: Int)

trait MyTrait {
  def traitMethod(): Unit
}

object MyObject {
  def objectMethod(): Unit = {}
  val myVal: Int = 42
  var myVar: Int = 0
}

sealed trait SealedTrait

enum MyEnum {
  case A, B
}

type MyAlias = String
`),
			expected: map[string][2]string{
				"com.example":    {"module", "package"},
				"MyClass":        {"type", "class"},
				"method":         {"callable", "method"},
				"AbstractClass":  {"type", "class"},
				"abstractMethod": {"callable", "method"},
				"concreteMethod": {"callable", "method"},
				"CaseClass":      {"type", "class"},
				"MyTrait":        {"type", "trait"},
				"traitMethod":    {"callable", "method"},
				"MyObject":       {"type", "object"},
				"objectMethod":   {"callable", "method"},
				"myVal":          {"value", "val"},
				"myVar":          {"value", "variable"},
				"SealedTrait":    {"type", "trait"},
				"MyEnum":         {"type", "enum"},
				"MyAlias":        {"type", "type_alias"},
			},
		},
	}
}

func TestLanguageFeatureCoverage(t *testing.T) {
	for _, tc := range coverageCases() {
		t.Run(tc.name, func(t *testing.T) {
			symbols, _, err := parseSrc(t, tc.plugin, tc.filePath, tc.src)
			if err != nil {
				// Allow syntax warnings (partial extraction) but not hard errors.
				var sw *plugin.SyntaxError
				if !errors.As(err, &sw) {
					t.Fatalf("parse error: %v", err)
				}
			}

			// Build a multi-map: name → list of (category, kind) pairs.
			// This handles cases where multiple symbols share a name
			// (e.g., Java class + constructor both named "MyClass").
			type catKind struct{ cat, kind string }
			found := make(map[string][]catKind)
			for _, s := range symbols {
				found[s.Name] = append(found[s.Name], catKind{string(s.Category), s.Kind})
			}

			for name, want := range tc.expected {
				entries, ok := found[name]
				if !ok {
					t.Errorf("MISSING symbol %q (expected %s/%s)", name, want[0], want[1])
					continue
				}
				matched := false
				for _, e := range entries {
					if e.cat == want[0] && e.kind == want[1] {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("symbol %q: got %v, want %s/%s", name, entries, want[0], want[1])
				}
			}
		})
	}
}
