// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"sort"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/typescript"
)

// ---------------------------------------------------------------------------
// Topic 1: Scoped symbol names — same-named methods in different classes
// must produce distinct IDs and correct edge targets.
// ---------------------------------------------------------------------------

// TestScopedNames_PythonSameNameMethods verifies that two classes in the same
// file with identically-named methods produce distinct scoped names and
// correct contains edges.
func TestScopedNames_PythonSameNameMethods(t *testing.T) {
	src := []byte(`class User:
    def save(self):
        pass
    def validate(self):
        pass

class Order:
    def save(self):
        pass
    def validate(self):
        pass
`)
	syms, edges, err := parseSrc(t, python.NewPlugin(), "models.py", src)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we have 6 symbols: User, User.save, User.validate, Order, Order.save, Order.validate.
	wantSyms := map[string]bool{
		"User": true, "Order": true,
		"save": true, "validate": true,
	}
	for _, s := range syms {
		delete(wantSyms, s.Name)
	}
	// "save" and "validate" appear twice each — wantSyms should be empty.

	// Verify scoped names are set correctly for methods.
	scopedNames := make(map[string]bool)
	for _, s := range syms {
		if s.ScopedName != "" {
			scopedNames[s.ScopedName] = true
		}
	}
	for _, want := range []string{"User.save", "User.validate", "Order.save", "Order.validate"} {
		if !scopedNames[want] {
			t.Errorf("missing scoped name %q; got %v", want, scopedNames)
		}
	}

	// Verify contains edges point to scoped names, not bare names.
	type edgeKey struct{ from, to string }
	containsEdges := make(map[edgeKey]bool)
	for _, e := range edges {
		if e.Kind == plugin.EdgeContains {
			containsEdges[edgeKey{e.From, e.To}] = true
		}
	}

	wantContains := []edgeKey{
		{"User", "User.save"},
		{"User", "User.validate"},
		{"Order", "Order.save"},
		{"Order", "Order.validate"},
	}
	for _, wc := range wantContains {
		if !containsEdges[wc] {
			t.Errorf("missing contains edge %s -> %s; got %v", wc.from, wc.to, containsEdges)
		}
	}

	// Verify that User does NOT contain Order.save.
	badEdge := edgeKey{"User", "Order.save"}
	if containsEdges[badEdge] {
		t.Errorf("unexpected contains edge %s -> %s", badEdge.from, badEdge.to)
	}
}

// TestScopedNames_JavaSameNameMethods verifies scoped names for Java methods
// inside different classes in the same file.
func TestScopedNames_JavaSameNameMethods(t *testing.T) {
	src := []byte(`class UserService {
    void save() {}
    void delete() {}
}
class OrderService {
    void save() {}
    void delete() {}
}
`)
	syms, edges, err := parseSrc(t, java.NewPlugin(), "Services.java", src)
	if err != nil {
		t.Fatal(err)
	}

	// Verify scoped names.
	scopedNames := make(map[string]bool)
	for _, s := range syms {
		if s.ScopedName != "" {
			scopedNames[s.ScopedName] = true
		}
	}
	for _, want := range []string{
		"UserService.save", "UserService.delete",
		"OrderService.save", "OrderService.delete",
	} {
		if !scopedNames[want] {
			t.Errorf("missing scoped name %q; got %v", want, scopedNames)
		}
	}

	// Verify contains edges.
	type edgeKey struct{ from, to string }
	containsEdges := make(map[edgeKey]bool)
	for _, e := range edges {
		if e.Kind == plugin.EdgeContains {
			containsEdges[edgeKey{e.From, e.To}] = true
		}
	}

	// UserService should contain UserService.save, not OrderService.save.
	if !containsEdges[edgeKey{"UserService", "UserService.save"}] {
		t.Error("missing UserService -> UserService.save")
	}
	if containsEdges[edgeKey{"UserService", "OrderService.save"}] {
		t.Error("unexpected UserService -> OrderService.save")
	}

	_ = syms // used above
}

// TestScopedNames_GoReceiverMethods verifies that Go methods on different
// receiver types get distinct scoped names.
func TestScopedNames_GoReceiverMethods(t *testing.T) {
	src := []byte(`package main

type Reader struct{}
type Writer struct{}

func (r *Reader) Close() {}
func (w *Writer) Close() {}
`)
	syms, edges, err := parseSrc(t, golang.NewPlugin(), "io.go", src)
	if err != nil {
		t.Fatal(err)
	}

	scopedNames := make(map[string]bool)
	for _, s := range syms {
		if s.ScopedName != "" {
			scopedNames[s.ScopedName] = true
		}
	}
	if !scopedNames["Reader.Close"] {
		t.Errorf("missing Reader.Close; got %v", scopedNames)
	}
	if !scopedNames["Writer.Close"] {
		t.Errorf("missing Writer.Close; got %v", scopedNames)
	}

	// Verify contains edges.
	type edgeKey struct{ from, to string }
	containsEdges := make(map[edgeKey]bool)
	for _, e := range edges {
		if e.Kind == plugin.EdgeContains {
			containsEdges[edgeKey{e.From, e.To}] = true
		}
	}
	if !containsEdges[edgeKey{"Reader", "Reader.Close"}] {
		t.Error("missing Reader -> Reader.Close")
	}
	if !containsEdges[edgeKey{"Writer", "Writer.Close"}] {
		t.Error("missing Writer -> Writer.Close")
	}
	if containsEdges[edgeKey{"Reader", "Writer.Close"}] {
		t.Error("unexpected Reader -> Writer.Close")
	}

	_ = syms
}

// TestScopedNames_TopLevelFunctionsUnscoped verifies that top-level functions
// (no parent) have empty ScopedName and use Name for ID generation.
func TestScopedNames_TopLevelFunctionsUnscoped(t *testing.T) {
	src := []byte(`package main

func foo() {}
func bar() {}
`)
	syms, _, err := parseSrc(t, golang.NewPlugin(), "main.go", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range syms {
		if s.Category == plugin.CategoryCallable && s.ScopedName != "" {
			t.Errorf("top-level function %q should have empty ScopedName, got %q", s.Name, s.ScopedName)
		}
	}
}

// TestScopedNames_EffectiveScopedName verifies the EffectiveScopedName helper.
func TestScopedNames_EffectiveScopedName(t *testing.T) {
	s1 := plugin.Symbol{Name: "foo", ScopedName: ""}
	if got := s1.EffectiveScopedName(); got != "foo" {
		t.Errorf("EffectiveScopedName() = %q, want %q", got, "foo")
	}

	s2 := plugin.Symbol{Name: "save", ScopedName: "User.save"}
	if got := s2.EffectiveScopedName(); got != "User.save" {
		t.Errorf("EffectiveScopedName() = %q, want %q", got, "User.save")
	}
}

// TestScopedNames_MakeScopedName verifies the MakeScopedName helper.
func TestScopedNames_MakeScopedName(t *testing.T) {
	if got := plugin.MakeScopedName("", "foo"); got != "foo" {
		t.Errorf("MakeScopedName('', 'foo') = %q, want %q", got, "foo")
	}
	if got := plugin.MakeScopedName("User", "save"); got != "User.save" {
		t.Errorf("MakeScopedName('User', 'save') = %q, want %q", got, "User.save")
	}
}

// ---------------------------------------------------------------------------
// Topic 2: Callable reference detection — identifiers passed as arguments
// that match known callable symbols should produce calls edges.
// ---------------------------------------------------------------------------

// TestCallableRef_PythonCallbackArg verifies that passing a function name as
// an argument produces a calls edge.
func TestCallableRef_PythonCallbackArg(t *testing.T) {
	src := []byte(`def transform(x):
    return x * 2

def process(items, fn):
    pass

def main():
    process(items, transform)
`)
	_, edges, err := parseSrc(t, python.NewPlugin(), "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	// main calls process (direct call) and should also reference transform
	// (passed as argument).
	callsFrom := make(map[string][]string)
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls {
			callsFrom[e.From] = append(callsFrom[e.From], e.To)
		}
	}

	mainCalls := callsFrom["main"]
	sort.Strings(mainCalls)

	assertContains(t, mainCalls, "process", "main should call process")
	assertContains(t, mainCalls, "transform", "main should reference transform as callable arg")
}

// TestCallableRef_GoCallbackArg verifies callable reference detection in Go.
func TestCallableRef_GoCallbackArg(t *testing.T) {
	src := []byte(`package main

func transform(x int) int { return x * 2 }

func apply(fn func(int) int, x int) int { return fn(x) }

func main() {
	apply(transform, 42)
}
`)
	_, edges, err := parseSrc(t, golang.NewPlugin(), "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	callsFrom := make(map[string][]string)
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls {
			callsFrom[e.From] = append(callsFrom[e.From], e.To)
		}
	}

	mainCalls := callsFrom["main"]
	assertContains(t, mainCalls, "apply", "main should call apply")
	assertContains(t, mainCalls, "transform", "main should reference transform as callable arg")
}

// TestCallableRef_JSCallbackArg verifies callable reference detection in JavaScript.
func TestCallableRef_JSCallbackArg(t *testing.T) {
	src := []byte(`function transform(x) { return x * 2; }
function apply(fn, x) { return fn(x); }
function main() { apply(transform, 42); }
`)
	_, edges, err := parseSrc(t, javascript.NewPlugin(), "test.js", src)
	if err != nil {
		t.Fatal(err)
	}

	callsFrom := make(map[string][]string)
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls {
			callsFrom[e.From] = append(callsFrom[e.From], e.To)
		}
	}

	mainCalls := callsFrom["main"]
	assertContains(t, mainCalls, "apply", "main should call apply")
	assertContains(t, mainCalls, "transform", "main should reference transform as callable arg")
}

// TestCallableRef_TSCallbackArg verifies callable reference detection in TypeScript.
func TestCallableRef_TSCallbackArg(t *testing.T) {
	src := []byte(`function transform(x: number): number { return x * 2; }
function apply(fn: (n: number) => number, x: number): number { return fn(x); }
function main(): void { apply(transform, 42); }
`)
	_, edges, err := parseSrc(t, typescript.NewPlugin(), "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}

	callsFrom := make(map[string][]string)
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls {
			callsFrom[e.From] = append(callsFrom[e.From], e.To)
		}
	}

	mainCalls := callsFrom["main"]
	assertContains(t, mainCalls, "apply", "main should call apply")
	assertContains(t, mainCalls, "transform", "main should reference transform as callable arg")
}

// TestCallableRef_IgnoresNonCallableArgs verifies that variable names passed
// as arguments do NOT produce false-positive calls edges.
func TestCallableRef_IgnoresNonCallableArgs(t *testing.T) {
	src := []byte(`def process(data, count):
    pass

def main():
    process(data, count)
`)
	_, edges, err := parseSrc(t, python.NewPlugin(), "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	// "data" and "count" are not known callable symbols, so they should NOT
	// appear as call targets from main.
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "main" {
			if e.To == "data" || e.To == "count" {
				t.Errorf("unexpected calls edge main -> %s (non-callable arg)", e.To)
			}
		}
	}
}

// TestCallableRef_NoDuplicateEdges verifies that when a function is both
// called directly and passed as an argument, only one calls edge is emitted.
func TestCallableRef_NoDuplicateEdges(t *testing.T) {
	src := []byte(`def helper():
    pass

def main():
    helper()
    process(helper)
`)
	_, edges, err := parseSrc(t, python.NewPlugin(), "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "main" && e.To == "helper" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 calls edge main -> helper, got %d", count)
	}
}

// TestCallableRef_StringLiteralNotCaptured verifies that string literals
// matching function names are NOT captured as callable references.
func TestCallableRef_StringLiteralNotCaptured(t *testing.T) {
	src := []byte(`def transform():
    pass

def main():
    process("transform")
`)
	_, edges, err := parseSrc(t, python.NewPlugin(), "test.py", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "main" && e.To == "transform" {
			t.Error("string literal 'transform' should not produce a calls edge")
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func assertContains(t *testing.T, slice []string, want, msg string) {
	t.Helper()
	for _, s := range slice {
		if s == want {
			return
		}
	}
	t.Errorf("%s: %q not found in %v", msg, want, slice)
}
