// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/typescript"
)

// collectEdges returns all edges matching the given kind.
func collectEdges(edges []plugin.Edge, kind plugin.EdgeKind) []plugin.Edge {
	var result []plugin.Edge
	for _, e := range edges {
		if e.Kind == kind {
			result = append(result, e)
		}
	}
	return result
}

// edgeExists checks if a specific from->to edge of the given kind exists.
func edgeExists(edges []plugin.Edge, from, to string, kind plugin.EdgeKind) bool {
	for _, e := range edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return true
		}
	}
	return false
}

// TestDataflow_TS_ObjectDispatchTable verifies that when a function is assigned
// to an object property and then called via obj.prop or obj['prop'], the
// dataflow system links the property to the original function.
func TestDataflow_TS_ObjectDispatchTable(t *testing.T) {
	src := []byte(`
function funcA(): void { }
function funcB(): void { }
function funcC(): void { }

const handlers = {
    propa: funcA,
    propb: funcB,
    propc: funcC,
};

function dispatch() {
    handlers.propa();
    handlers['propb']();
}
`)
	symbols, edges := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	t.Log("=== Symbols ===")
	for _, s := range symbols {
		t.Logf("  %s/%s %q scoped=%q", s.Category, s.Kind, s.Name, s.ScopedName)
	}
	t.Log("=== Edges ===")
	for _, e := range edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	// Check that funcA, funcB, funcC are extracted
	for _, name := range []string{"funcA", "funcB", "funcC"} {
		if findSymbol(symbols, name) == nil {
			t.Errorf("missing function %q", name)
		}
	}

	// Check for aliases edges: propa -> funcA, propb -> funcB, propc -> funcC
	aliases := collectEdges(edges, plugin.EdgeAliases)
	t.Logf("=== Alias edges (%d) ===", len(aliases))
	for _, e := range aliases {
		t.Logf("  %s -> %s", e.From, e.To)
	}

	if !edgeExists(edges, "propa", "funcA", plugin.EdgeAliases) {
		t.Error("missing alias edge: propa -> funcA")
	}
	if !edgeExists(edges, "propb", "funcB", plugin.EdgeAliases) {
		t.Error("missing alias edge: propb -> funcB")
	}
	if !edgeExists(edges, "propc", "funcC", plugin.EdgeAliases) {
		t.Error("missing alias edge: propc -> funcC")
	}
}

// TestDataflow_JS_ObjectDispatchTable same test for JavaScript.
func TestDataflow_JS_ObjectDispatchTable(t *testing.T) {
	src := []byte(`
function funcA() { }
function funcB() { }

const handlers = {
    propa: funcA,
    propb: funcB,
};

function dispatch() {
    handlers.propa();
    handlers['propb']();
}
`)
	symbols, edges := parseWith(t, javascript.NewPlugin(), "test.js", src)

	t.Log("=== Symbols ===")
	for _, s := range symbols {
		t.Logf("  %s/%s %q scoped=%q", s.Category, s.Kind, s.Name, s.ScopedName)
	}
	t.Log("=== Edges ===")
	for _, e := range edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	aliases := collectEdges(edges, plugin.EdgeAliases)
	t.Logf("=== Alias edges (%d) ===", len(aliases))
	for _, e := range aliases {
		t.Logf("  %s -> %s", e.From, e.To)
	}

	if !edgeExists(edges, "propa", "funcA", plugin.EdgeAliases) {
		t.Error("missing alias edge: propa -> funcA")
	}
	if !edgeExists(edges, "propb", "funcB", plugin.EdgeAliases) {
		t.Error("missing alias edge: propb -> funcB")
	}
}

// TestDataflow_TS_VariableAlias verifies that simple variable aliasing works:
// const myFunc = originalFunc; myFunc() should link to originalFunc.
func TestDataflow_TS_VariableAlias(t *testing.T) {
	src := []byte(`
function originalFunc(): void { }

const myFunc = originalFunc;

function caller() {
    myFunc();
}
`)
	symbols, edges := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	t.Log("=== Symbols ===")
	for _, s := range symbols {
		t.Logf("  %s/%s %q", s.Category, s.Kind, s.Name)
	}
	t.Log("=== Edges ===")
	for _, e := range edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	if findSymbol(symbols, "originalFunc") == nil {
		t.Error("missing function 'originalFunc'")
	}

	// Check alias: myFunc -> originalFunc
	if !edgeExists(edges, "myFunc", "originalFunc", plugin.EdgeAliases) {
		t.Error("missing alias edge: myFunc -> originalFunc")
	}
}

// TestDataflow_TS_CallThroughAlias verifies that calling through an alias
// produces a calls edge.
func TestDataflow_TS_CallThroughAlias(t *testing.T) {
	src := []byte(`
function target(): void { }

const alias = target;

function caller() {
    alias();
}
`)
	_, edges := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	t.Log("=== All edges ===")
	for _, e := range edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	// Should have alias edge
	if !edgeExists(edges, "alias", "target", plugin.EdgeAliases) {
		t.Error("missing alias edge: alias -> target")
	}

	// caller should call alias (direct call edge)
	if !edgeExists(edges, "caller", "alias", plugin.EdgeCalls) {
		t.Error("missing call edge: caller -> alias")
	}
}
