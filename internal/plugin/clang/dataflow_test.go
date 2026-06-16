// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clang

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_AssignmentAlias(t *testing.T) {
	// Use typedef to simplify the function pointer declaration.
	// Raw function pointer syntax (void (*handler)() = myFunc) has a deeply
	// nested AST that the generic extractor can't handle without C-specific logic.
	src := []byte(`
void myFunc() {}

typedef void (*func_t)();

void setup() {
    func_t handler = myFunc;
}
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasAliasTo(edges, "myFunc") {
		logEdges(t, edges)
		t.Fatalf("expected alias to myFunc")
	}
}

func TestDataflow_ReturnTracking(t *testing.T) {
	src := []byte(`
void myFunc() {}

void* getHandler() {
    return myFunc;
}
`)
	_, edges, err := parseSource(t, "test.c", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasReturn(edges, "getHandler", "myFunc") {
		logEdges(t, edges)
		t.Fatalf("expected return edge getHandler->myFunc")
	}
}

func hasAliasTo(edges []plugin.Edge, to string) bool {
	for _, e := range edges {
		if e.Kind == plugin.EdgeAliases && e.To == to {
			return true
		}
	}
	return false
}

func hasReturn(edges []plugin.Edge, from, to string) bool {
	for _, e := range edges {
		if e.Kind == plugin.EdgeReturns && e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func logEdges(t *testing.T, edges []plugin.Edge) {
	t.Helper()
	for _, e := range edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}
}
