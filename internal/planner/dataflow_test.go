// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"sort"
	"testing"

	"codeknit/internal/plugin"
)

func TestResolveDataflow_AliasResolution(t *testing.T) {
	// Scenario: const handler = myFunc; handler() should resolve to myFunc
	symbols := []plugin.Symbol{
		{ID: "file.ts::myFunc", Name: "myFunc", Category: plugin.CategoryCallable},
		{ID: "file.ts::handler", Name: "handler", Category: plugin.CategoryValue},
		{ID: "file.ts::main", Name: "main", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		// handler aliases myFunc (from dataflow extraction)
		{From: "handler", To: "myFunc", Kind: plugin.EdgeAliases},
		// main calls handler (AST-level: main() { handler() })
		{From: "file.ts::main", To: "file.ts::handler", Kind: plugin.EdgeCalls},
	}
	localToGlobal := map[fileLocal]string{
		{file: "file.ts", name: "myFunc"}:  "file.ts::myFunc",
		{file: "file.ts", name: "handler"}: "file.ts::handler",
		{file: "file.ts", name: "main"}:    "file.ts::main",
	}
	globalByName := map[string][]string{
		"myFunc":  {"file.ts::myFunc"},
		"handler": {"file.ts::handler"},
		"main":    {"file.ts::main"},
	}

	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)

	// Should have: main→handler (original) + main→myFunc (resolved through alias)
	calls := callEdgesFrom(result, "file.ts::main")
	sort.Strings(calls)
	want := []string{"file.ts::handler", "file.ts::myFunc"}
	if !strSliceEqual(calls, want) {
		t.Fatalf("expected calls %v, got %v", want, calls)
	}
}

func TestResolveDataflow_ObjectPropertyAlias(t *testing.T) {
	// Scenario: const handlers = { forOf: forOfStmt }
	// forOf aliases forOfStmt
	symbols := []plugin.Symbol{
		{ID: "file.ts::forOfStmt", Name: "forOfStmt", Category: plugin.CategoryCallable},
		{ID: "file.ts::caller", Name: "caller", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "forOf", To: "forOfStmt", Kind: plugin.EdgeAliases},
		// caller calls "forOf" (through obj.forOf())
		{From: "file.ts::caller", To: "file.ts::forOf", Kind: plugin.EdgeCalls},
	}
	localToGlobal := map[fileLocal]string{
		{file: "file.ts", name: "forOfStmt"}: "file.ts::forOfStmt",
		{file: "file.ts", name: "caller"}:    "file.ts::caller",
	}
	globalByName := map[string][]string{
		"forOfStmt": {"file.ts::forOfStmt"},
		"caller":    {"file.ts::caller"},
	}

	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)

	calls := callEdgesFrom(result, "file.ts::caller")
	sort.Strings(calls)
	// Should resolve forOf → forOfStmt
	want := []string{"file.ts::forOf", "file.ts::forOfStmt"}
	if !strSliceEqual(calls, want) {
		t.Fatalf("expected calls %v, got %v", want, calls)
	}
}

func TestResolveDataflow_ReturnValueTracking(t *testing.T) {
	// Scenario: function getHandler() { return myFunc }
	// caller calls getHandler, should transitively depend on myFunc
	symbols := []plugin.Symbol{
		{ID: "file.ts::myFunc", Name: "myFunc", Category: plugin.CategoryCallable},
		{ID: "file.ts::getHandler", Name: "getHandler", Category: plugin.CategoryCallable},
		{ID: "file.ts::caller", Name: "caller", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "file.ts::getHandler", To: "myFunc", Kind: plugin.EdgeReturns},
		{From: "file.ts::caller", To: "file.ts::getHandler", Kind: plugin.EdgeCalls},
	}
	localToGlobal := map[fileLocal]string{
		{file: "file.ts", name: "myFunc"}:     "file.ts::myFunc",
		{file: "file.ts", name: "getHandler"}: "file.ts::getHandler",
		{file: "file.ts", name: "caller"}:     "file.ts::caller",
	}
	globalByName := map[string][]string{
		"myFunc":     {"file.ts::myFunc"},
		"getHandler": {"file.ts::getHandler"},
		"caller":     {"file.ts::caller"},
	}

	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)

	calls := callEdgesFrom(result, "file.ts::caller")
	sort.Strings(calls)
	want := []string{"file.ts::getHandler", "file.ts::myFunc"}
	if !strSliceEqual(calls, want) {
		t.Fatalf("expected calls %v, got %v", want, calls)
	}
}

func TestResolveDataflow_TransitiveAlias(t *testing.T) {
	// Scenario: x = myFunc; y = x; caller calls y → should resolve to myFunc
	symbols := []plugin.Symbol{
		{ID: "f::myFunc", Name: "myFunc", Category: plugin.CategoryCallable},
		{ID: "f::caller", Name: "caller", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "x", To: "myFunc", Kind: plugin.EdgeAliases},
		{From: "y", To: "x", Kind: plugin.EdgeAliases},
		{From: "f::caller", To: "f::y", Kind: plugin.EdgeCalls},
	}
	localToGlobal := map[fileLocal]string{
		{file: "f", name: "myFunc"}: "f::myFunc",
		{file: "f", name: "caller"}: "f::caller",
	}
	globalByName := map[string][]string{
		"myFunc": {"f::myFunc"},
		"caller": {"f::caller"},
	}

	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)

	calls := callEdgesFrom(result, "f::caller")
	// Should resolve y → x → myFunc
	found := false
	for _, c := range calls {
		if c == "f::myFunc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected transitive alias resolution to myFunc, got calls: %v", calls)
	}
}

func TestResolveDataflow_CycleDetection(t *testing.T) {
	// Scenario: x = y; y = x (cycle) — should not infinite loop
	symbols := []plugin.Symbol{
		{ID: "f::caller", Name: "caller", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "x", To: "y", Kind: plugin.EdgeAliases},
		{From: "y", To: "x", Kind: plugin.EdgeAliases},
		{From: "f::caller", To: "f::x", Kind: plugin.EdgeCalls},
	}
	localToGlobal := map[fileLocal]string{
		{file: "f", name: "caller"}: "f::caller",
	}
	globalByName := map[string][]string{
		"caller": {"f::caller"},
	}

	// Should not hang
	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)
	_ = result // just verify it terminates
}

func TestResolveDataflow_NoMetadataEdgesInOutput(t *testing.T) {
	// Metadata edges (aliases, returns) should be stripped from output
	symbols := []plugin.Symbol{
		{ID: "f::a", Name: "a", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "x", To: "a", Kind: plugin.EdgeAliases},
		{From: "f::a", To: "a", Kind: plugin.EdgeReturns},
		{From: "f::a", To: "f::a", Kind: plugin.EdgeCalls}, // self-call, just to have a structural edge
	}
	localToGlobal := map[fileLocal]string{{file: "f", name: "a"}: "f::a"}
	globalByName := map[string][]string{"a": {"f::a"}}

	result := resolveDataflow(symbols, edges, localToGlobal, globalByName)

	for _, e := range result {
		if e.Kind == plugin.EdgeAliases || e.Kind == plugin.EdgeReturns {
			t.Fatalf("metadata edge %s should have been stripped, got: %+v", e.Kind, e)
		}
	}
}

func TestResolveDataflow_NoHintsPassthrough(t *testing.T) {
	// When there are no dataflow hints, edges pass through unchanged
	edges := []plugin.Edge{
		{From: "a", To: "b", Kind: plugin.EdgeCalls},
		{From: "c", To: "d", Kind: plugin.EdgeContains},
	}

	result := resolveDataflow(nil, edges, nil, nil)

	if len(result) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(result))
	}
}

func TestExtractLocalName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"file.ts::myFunc", "myFunc"},
		{"path/to/file.go::User.Save", "User.Save"},
		{"file.ts::myFunc#1", "myFunc"},
		{"nodelimiter", ""},
		{"a::b::c", "c"},
	}
	for _, tc := range tests {
		got := extractLocalName(tc.input)
		if got != tc.want {
			t.Errorf("extractLocalName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- helpers ---

func callEdgesFrom(edges []plugin.Edge, from string) []string {
	var targets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == from {
			targets = append(targets, e.To)
		}
	}
	return targets
}

func strSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
