// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package golang

import (
	"errors"
	"hash/fnv"
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

func TestParseWithOptions_FingerprintCallMetadata(t *testing.T) {
	src := []byte(`package main

var DefaultValue = makeValue(1, 2)

func run() {
	processValue(1, 2, 3)
}
`)
	path := filepath.Join(t.TempDir(), "calls.go")
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	symbols, _, err := NewPlugin().ParseWithOptions(path, true)
	if err != nil {
		t.Fatal(err)
	}

	assertCallPayload := func(symbolName, callee string, argCount byte) {
		t.Helper()
		for _, symbol := range symbols {
			if symbol.Name != symbolName {
				continue
			}

			hasher := fnv.New32a()
			_, _ = hasher.Write([]byte(callee))
			hash := hasher.Sum(nil)
			want := []byte{plugin.FPCall, hash[0], hash[1], argCount}
			if len(symbol.BodyTokens) < len(want) {
				t.Fatalf("%s tokens %v are too short for call payload %v", symbolName, symbol.BodyTokens, want)
			}
			for i := range want {
				if symbol.BodyTokens[i] != want[i] {
					t.Fatalf("%s call payload: got %v, want prefix %v", symbolName, symbol.BodyTokens, want)
				}
			}
			return
		}
		t.Fatalf("symbol %q not found", symbolName)
	}

	assertCallPayload("run", "processvalue", 3)
	assertCallPayload("DefaultValue", "makevalue", 2)
}

func TestParseSource_Symbols(t *testing.T) {
	src := []byte(`
package main

type Greeter struct {
	Name string
}

func (g *Greeter) Greet() string {
	return g.Name
}

func main() {
}

var Version = "1.0"

const MaxRetries = 3
`)

	symbols, _, err := parseSource(t, "test.go", src)
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
		"main":       {plugin.CategoryCallable, "function"},
		"Greeter":    {plugin.CategoryType, "struct"},
		"Greet":      {plugin.CategoryCallable, "method"},
		"Version":    {plugin.CategoryValue, "variable"},
		"MaxRetries": {plugin.CategoryValue, "constant"},
	}

	// Check package symbol separately since "main" is also a function.
	foundPkg := false
	for _, s := range symbols {
		if s.Name == "main" && s.Category == plugin.CategoryModule && s.Kind == "package" {
			foundPkg = true
		}
	}
	if !foundPkg {
		t.Error("missing package symbol 'main'")
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

func TestParseSource_ExportedDetection(t *testing.T) {
	src := []byte(`
package main

func ExportedFunc() {}
func unexportedFunc() {}

type ExportedType struct{}
type unexportedType struct{}

var ExportedVar = 1
var unexportedVar = 2

const ExportedConst = 1
const unexportedConst = 2
`)
	symbols, _, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		if sym.Kind == "package" {
			continue
		}
		exported := isExported(sym.Name)
		got := sym.Properties["exported"]
		if exported && got != "true" {
			t.Errorf("symbol %q: exported got %q, want %q", sym.Name, got, "true")
		}
		if !exported && got != "" {
			t.Errorf("symbol %q: exported got %q, want %q", sym.Name, got, "")
		}
	}
}

func TestParseSource_MethodReceiver(t *testing.T) {
	src := []byte(`
package main

type Foo struct{}

func (f *Foo) PtrMethod() {}
func (f Foo) ValMethod() {}
`)
	symbols, _, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		switch sym.Name {
		case "PtrMethod":
			if sym.Properties["receiver"] != "*Foo" {
				t.Errorf("PtrMethod receiver: got %q, want %q", sym.Properties["receiver"], "*Foo")
			}
		case "ValMethod":
			if sym.Properties["receiver"] != "Foo" {
				t.Errorf("ValMethod receiver: got %q, want %q", sym.Properties["receiver"], "Foo")
			}
		}
	}
}

func TestParseSource_TypeRouting(t *testing.T) {
	src := []byte(`
package main

type MyStruct struct {
	Field1 string
}

type MyInterface interface {
	DoSomething()
}

type MyAlias = int
`)
	symbols, _, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]string)
	for _, s := range symbols {
		if s.Category == plugin.CategoryType {
			found[s.Name] = s.Kind
		}
	}

	expect := map[string]string{
		"MyStruct":    "struct",
		"MyInterface": "interface",
		"MyAlias":     "type_alias",
	}
	for name, wantKind := range expect {
		gotKind, ok := found[name]
		if !ok {
			t.Errorf("missing type symbol %q", name)
			continue
		}
		if gotKind != wantKind {
			t.Errorf("type %q: got kind %q, want %q", name, gotKind, wantKind)
		}
	}
}

func TestParseSource_ContainsEdges(t *testing.T) {
	src := []byte(`
package main

type MyStruct struct {
	Name string
	Age  int
}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "MyStruct" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"MyStruct.Age", "MyStruct.Name"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_MethodContainsEdge(t *testing.T) {
	src := []byte(`
package main

type Greeter struct{}

func (g *Greeter) Greet() {}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range edges {
		if e.From == "Greeter" && e.To == "Greeter.Greet" && e.Kind == plugin.EdgeContains {
			found = true
		}
	}
	if !found {
		t.Error("expected contains edge from Greeter to Greet")
	}
}

func TestParseSource_CallEdges(t *testing.T) {
	src := []byte(`
package main

import "fmt"

func hello() {
	fmt.Println("hello")
	helper()
}

func helper() {}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "hello" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	want := []string{"helper"}
	if len(callTargets) != len(want) {
		t.Fatalf("expected call targets %v, got %v", want, callTargets)
	}
	for i := range want {
		if callTargets[i] != want[i] {
			t.Errorf("call target[%d]: got %q, want %q", i, callTargets[i], want[i])
		}
	}
}

func TestParseSource_InterfaceContainsEdge(t *testing.T) {
	src := []byte(`
package main

type Reader interface {
	Read(p []byte) (int, error)
	Close() error
}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Reader" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Reader.Close", "Reader.Read"}
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
	src := []byte(`package main

func broken( {
`)
	_, _, err := parseSource(t, "bad.go", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.go") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`package main

func broken( {
`)
	symbols, _, err := parseSource(t, "bad.go", src)
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
	if len(exts) != 1 || exts[0] != ".go" {
		t.Errorf("expected [.go], got %v", exts)
	}
}

// Property 11: Go exported detection
// For any Go source file, a Symbol should have Properties["exported"] == "true"
// if and only if its name starts with an uppercase letter.
// **Validates: Requirement 5.1**
func TestProperty_GoExportedDetection(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		// Generate a random exported identifier (starts with uppercase).
		exportedName := genExportedIdent().Draw(t, "exportedName")
		// Generate a random unexported identifier (starts with lowercase).
		unexportedName := genUnexportedIdent().Draw(t, "unexportedName")

		// Ensure names are unique.
		if exportedName == unexportedName {
			return
		}

		var b strings.Builder
		b.WriteString("package main\n\n")
		b.WriteString("func " + exportedName + "() {}\n\n")
		b.WriteString("func " + unexportedName + "() {}\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.go", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}

		for _, sym := range symbols {
			if sym.Kind == "package" {
				continue
			}
			if sym.Name == exportedName {
				if sym.Properties["exported"] != "true" {
					t.Errorf("exported function %q should have exported=true, got %q", exportedName, sym.Properties["exported"])
				}
			}
			if sym.Name == unexportedName {
				if sym.Properties["exported"] != "" {
					t.Errorf("unexported function %q should have exported=\"\", got %q", unexportedName, sym.Properties["exported"])
				}
			}
		}
	})
}

// genExportedIdent generates a valid Go identifier starting with an uppercase letter.
func genExportedIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		chars[0] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[rapid.IntRange(0, 25).Draw(t, "first")]
		for i := 1; i < n; i++ {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}

// genUnexportedIdent generates a valid Go identifier starting with a lowercase letter.
func genUnexportedIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		chars[0] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "first")]
		for i := 1; i < n; i++ {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}
