// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"slices"
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
		{From: "file.ts::handler", To: "file.ts::myFunc", Kind: plugin.EdgeAliases},
		// main calls handler (AST-level: main() { handler() })
		{From: "file.ts::main", To: "file.ts::handler", Kind: plugin.EdgeCalls},
	}

	result := resolveDataflow(symbols, edges)

	// The alias call is replaced with its resolved callable target.
	calls := callEdgesFrom(result, "file.ts::main")
	sort.Strings(calls)
	want := []string{"file.ts::myFunc"}
	if !slices.Equal(calls, want) {
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
		{From: "file.ts::forOf", To: "file.ts::forOfStmt", Kind: plugin.EdgeAliases},
		// caller calls "forOf" (through obj.forOf())
		{From: "file.ts::caller", To: "file.ts::forOf", Kind: plugin.EdgeCalls},
	}

	result := resolveDataflow(symbols, edges)

	calls := callEdgesFrom(result, "file.ts::caller")
	sort.Strings(calls)
	// Should resolve forOf → forOfStmt
	want := []string{"file.ts::forOfStmt"}
	if !slices.Equal(calls, want) {
		t.Fatalf("expected calls %v, got %v", want, calls)
	}
}

func TestResolveDataflow_ReturnValueTracking(t *testing.T) {
	// Scenario: function getHandler() { return myFunc }
	// Calling getHandler does not invoke its returned myFunc.
	symbols := []plugin.Symbol{
		{ID: "file.ts::myFunc", Name: "myFunc", Category: plugin.CategoryCallable},
		{ID: "file.ts::getHandler", Name: "getHandler", Category: plugin.CategoryCallable},
		{ID: "file.ts::caller", Name: "caller", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "file.ts::getHandler", To: "file.ts::myFunc", Kind: plugin.EdgeReturns},
		{From: "file.ts::caller", To: "file.ts::getHandler", Kind: plugin.EdgeCalls},
	}

	result := resolveDataflow(symbols, edges)

	calls := callEdgesFrom(result, "file.ts::caller")
	sort.Strings(calls)
	want := []string{"file.ts::getHandler"}
	if !slices.Equal(calls, want) {
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
		{From: "f::x", To: "f::myFunc", Kind: plugin.EdgeAliases},
		{From: "f::y", To: "f::x", Kind: plugin.EdgeAliases},
		{From: "f::caller", To: "f::y", Kind: plugin.EdgeCalls},
	}

	result := resolveDataflow(symbols, edges)

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
		{From: "f::x", To: "f::y", Kind: plugin.EdgeAliases},
		{From: "f::y", To: "f::x", Kind: plugin.EdgeAliases},
		{From: "f::caller", To: "f::x", Kind: plugin.EdgeCalls},
	}

	// Should not hang
	result := resolveDataflow(symbols, edges)
	_ = result // just verify it terminates
}

func TestResolveDataflow_NoMetadataEdgesInOutput(t *testing.T) {
	// Metadata edges (aliases, returns) should be stripped from output
	symbols := []plugin.Symbol{
		{ID: "f::a", Name: "a", Category: plugin.CategoryCallable},
	}
	edges := []plugin.Edge{
		{From: "f::x", To: "f::a", Kind: plugin.EdgeAliases},
		{From: "f::a", To: "f::a", Kind: plugin.EdgeReturns},
		{From: "f::a", To: "f::a", Kind: plugin.EdgeCalls}, // self-call, just to have a structural edge
	}
	for _, e := range resolveDataflow(symbols, edges) {
		if e.Kind == plugin.EdgeAliases || e.Kind == plugin.EdgeReturns {
			t.Fatalf("unconsumed metadata: %+v", e)
		}
	}
}

func TestResolveDataflow_NoHintsPassthrough(t *testing.T) {
	// When there are no dataflow hints, edges pass through unchanged
	edges := []plugin.Edge{
		{From: "a", To: "b", Kind: plugin.EdgeCalls},
		{From: "c", To: "d", Kind: plugin.EdgeContains},
	}

	result := resolveDataflow(nil, edges)

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
	for edgeIndex := range edges {
		e := &edges[edgeIndex]
		if e.Kind == plugin.EdgeCalls && e.From == from {
			targets = append(targets, e.To)
		}
	}
	return targets
}
