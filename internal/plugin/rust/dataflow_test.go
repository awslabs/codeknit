// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rust

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_AssignmentAlias(t *testing.T) {
	src := []byte(`
fn my_func() {}

fn setup() {
    let handler = my_func;
}
`)
	_, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasAliasTo(edges, "my_func") {
		logEdges(t, edges)
		t.Fatalf("expected alias to my_func")
	}
}

func TestDataflow_ReturnTracking(t *testing.T) {
	src := []byte(`
fn my_func() {}

fn get_handler() -> fn() {
    return my_func;
}
`)
	_, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasReturn(edges, "get_handler", "my_func") {
		logEdges(t, edges)
		// Rust uses return_expression not return_statement
		t.Fatalf("expected return edge get_handler->my_func")
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
