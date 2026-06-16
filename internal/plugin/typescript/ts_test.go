// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package typescript

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

func TestParseSource_AllSymbolKinds(t *testing.T) {
	src := []byte(`
export function greet(name: string, age: number): string {
  return helper(name);
}

export class MyClass extends BaseClass implements IService {
  getValue(): number { return 0; }
}

interface IService {
  getValue(): number;
}

type Config = {
  host: string;
  port: number;
};

enum Status {
  Active,
  Inactive,
}

export const DEFAULT_OPTIONS = { timeout: 30 };

let count = 0;

function identity<T>(arg: T): T { return arg; }

export const helper = (x: number) => x * 2;
`)
	symbols, _, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected non-empty symbols")
	}

	// Build a map of name -> {Category, Kind} for easy lookup.
	type catKind struct {
		Category plugin.SymbolCategory
		Kind     string
	}
	found := make(map[string]catKind)
	for _, s := range symbols {
		found[s.Name] = catKind{s.Category, s.Kind}
	}

	expect := map[string]catKind{
		"greet":           {plugin.CategoryCallable, "function"},
		"MyClass":         {plugin.CategoryType, "class"},
		"getValue":        {plugin.CategoryCallable, "method"},
		"IService":        {plugin.CategoryType, "interface"},
		"Config":          {plugin.CategoryType, "type_alias"},
		"Status":          {plugin.CategoryType, "enum"},
		"DEFAULT_OPTIONS": {plugin.CategoryValue, "exported_constant"},
		"count":           {plugin.CategoryValue, "variable"},
		"identity":        {plugin.CategoryCallable, "function"},
		"helper":          {plugin.CategoryCallable, "arrow_function"},
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

func TestParseSource_FunctionSignature(t *testing.T) {
	src := []byte(`function greet(name: string, age: number): boolean { return true; }`)
	symbols, _, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}
	fn := symbols[0]
	wantSig := "greet(name: string, age: number) -> boolean"
	if fn.Signature != wantSig {
		t.Errorf("expected signature %q, got %q", wantSig, fn.Signature)
	}
}

func TestParseSource_CallEdges(t *testing.T) {
	src := []byte(`function foo() { bar(); baz(); utils.format(x); }`)
	_, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	// Look for edges with Kind == EdgeCalls and From == "foo".
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

func TestParseSource_HeritageEdges(t *testing.T) {
	src := []byte(`class Dog extends Animal implements Pet { }`)
	_, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	var heritageTargets []string
	for _, e := range edges {
		if e.From == "Dog" && (e.Kind == plugin.EdgeInherits || e.Kind == plugin.EdgeImplements) {
			heritageTargets = append(heritageTargets, e.To)
		}
	}
	sort.Strings(heritageTargets)
	if len(heritageTargets) != 2 || heritageTargets[0] != "Animal" || heritageTargets[1] != "Pet" {
		t.Errorf("expected heritage targets [Animal, Pet], got %v", heritageTargets)
	}
}

func TestParseSource_ExportedProperty(t *testing.T) {
	src := []byte(`
export function pub() {}
function priv() {}
`)
	symbols, _, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range symbols {
		if sym.Name == "pub" && sym.Properties["exported"] != "true" {
			t.Error("pub should have exported=true")
		}
		if sym.Name == "priv" && sym.Properties["exported"] != "" {
			t.Error("priv should have exported=\"\"")
		}
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`function broken( { }`) // missing closing paren
	_, _, err := parseSource(t, "bad.ts", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.ts") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`function broken( { }`)
	symbols, _, err := parseSource(t, "bad.ts", src)
	var sw *plugin.SyntaxError
	if err != nil && !errors.As(err, &sw) {
		t.Fatalf("expected nil or SyntaxWarning, got: %v", err)
	}
	if symbols == nil {
		t.Fatal("expected non-nil symbols for partial-error file")
	}
}

// Feature: code-concept-mapper, Property 10: TypeScript Symbol Extraction Completeness
func TestProperty_TypeScriptIdentifierExtraction(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		// Generate random identifiers of each kind and build a TS source.
		funcName := genIdent().Draw(t, "funcName")
		className := genUpperIdent().Draw(t, "className")
		ifaceName := genUpperIdent().Draw(t, "ifaceName")
		typeName := genUpperIdent().Draw(t, "typeName")
		enumName := genUpperIdent().Draw(t, "enumName")
		varName := genIdent().Draw(t, "varName")
		constName := genUpperIdent().Draw(t, "constName")
		genericName := genUpperSingleChar().Draw(t, "genericName")
		arrowName := genIdent().Draw(t, "arrowName")

		// Ensure all names are unique.
		names := []string{funcName, className, ifaceName, typeName, enumName, varName, constName, arrowName}
		seen := make(map[string]bool)
		for _, n := range names {
			if seen[n] {
				return // skip this iteration if collision
			}
			seen[n] = true
		}

		var b strings.Builder
		b.WriteString("function " + funcName + "<" + genericName + ">(x: " + genericName + "): " + genericName + " { return x; }\n")
		b.WriteString("class " + className + " { }\n")
		b.WriteString("interface " + ifaceName + " { }\n")
		b.WriteString("type " + typeName + " = { value: string };\n")
		b.WriteString("enum " + enumName + " { A, B }\n")
		b.WriteString("let " + varName + " = 0;\n")
		b.WriteString("export const " + constName + " = { key: 1 };\n")
		b.WriteString("const " + arrowName + " = (a: number) => a;\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.ts", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}
		if len(symbols) == 0 {
			t.Fatalf("no symbols for valid source:\n%s", src)
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
			funcName:  {plugin.CategoryCallable, "function"},
			className: {plugin.CategoryType, "class"},
			ifaceName: {plugin.CategoryType, "interface"},
			typeName:  {plugin.CategoryType, "type_alias"},
			enumName:  {plugin.CategoryType, "enum"},
			varName:   {plugin.CategoryValue, "variable"},
			constName: {plugin.CategoryValue, "exported_constant"},
			arrowName: {plugin.CategoryCallable, "arrow_function"},
		}

		for name, want := range expect {
			got, ok := found[name]
			if !ok {
				t.Errorf("missing symbol %q (expected %s/%s)\nsource:\n%s", name, want.Category, want.Kind, src)
				continue
			}
			if got.Category != want.Category || got.Kind != want.Kind {
				t.Errorf("symbol %q: got %s/%s, want %s/%s", name, got.Category, got.Kind, want.Category, want.Kind)
			}
		}
	})
}

// genIdent generates a valid lowercase identifier (a-z, 3-8 chars).
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

// genUpperIdent generates a valid PascalCase identifier.
func genUpperIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		chars[0] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[rapid.IntRange(0, 25).Draw(t, "ch")]
		for i := 1; i < n; i++ {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}

// genUpperSingleChar generates a single uppercase letter.
func genUpperSingleChar() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		return string("ABCDEFGHIJKLMNOPQRSTUVWXYZ"[rapid.IntRange(0, 25).Draw(t, "ch")])
	})
}

func TestParseSource_AbstractClass(t *testing.T) {
	src := []byte(`
export abstract class CloudResource {
  public readonly metadata: ResourceMetadata;

  public constructor(metadata: ResourceMetadata) {
    this.metadata = metadata;
  }

  abstract getType(): string;
}

export class ConcreteResource extends CloudResource {
  getType(): string { return "concrete"; }
}
`)
	symbols, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}

	// Check that CloudResource is extracted as an abstract class.
	type catKind struct {
		Category plugin.SymbolCategory
		Kind     string
	}
	found := make(map[string]catKind)
	props := make(map[string]map[string]string)
	for _, s := range symbols {
		found[s.Name] = catKind{s.Category, s.Kind}
		props[s.Name] = s.Properties
	}

	// Abstract class must be present.
	if got, ok := found["CloudResource"]; !ok {
		t.Error("missing symbol CloudResource")
	} else if got.Category != plugin.CategoryType || got.Kind != "class" {
		t.Errorf("CloudResource: got %s/%s, want type/class", got.Category, got.Kind)
	}

	// Abstract property should be set.
	if props["CloudResource"]["abstract"] != "true" {
		t.Errorf("CloudResource should have abstract=true, got %v", props["CloudResource"])
	}

	// Exported property should be set.
	if props["CloudResource"]["exported"] != "true" {
		t.Errorf("CloudResource should have exported=true, got %v", props["CloudResource"])
	}

	// Concrete class must also be present.
	if got, ok := found["ConcreteResource"]; !ok {
		t.Error("missing symbol ConcreteResource")
	} else if got.Category != plugin.CategoryType || got.Kind != "class" {
		t.Errorf("ConcreteResource: got %s/%s, want type/class", got.Category, got.Kind)
	}

	// ConcreteResource should inherit from CloudResource.
	foundInherits := false
	for _, e := range edges {
		if e.From == "ConcreteResource" && e.To == "CloudResource" && e.Kind == plugin.EdgeInherits {
			foundInherits = true
		}
	}
	if !foundInherits {
		t.Error("expected ConcreteResource --inherits--> CloudResource edge")
	}

	// Abstract method getType should be extracted from CloudResource.
	if got, ok := found["getType"]; !ok {
		t.Error("missing symbol getType")
	} else if got.Category != plugin.CategoryCallable || got.Kind != "method" {
		t.Errorf("getType: got %s/%s, want callable/method", got.Category, got.Kind)
	}
}
