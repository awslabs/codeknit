// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package golang

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_AssignmentAlias(t *testing.T) {
	src := []byte(`
package main

func myFunc() {}

func setup() {
	handler := myFunc
	_ = handler
}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasAlias(edges, "handler", "myFunc") {
		logEdges(t, edges)
		t.Fatalf("expected alias handler=myFunc")
	}
}

func TestDataflow_ReturnTracking(t *testing.T) {
	src := []byte(`
package main

func myFunc() {}

func getHandler() func() {
	return myFunc
}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasReturn(edges, "getHandler", "myFunc") {
		logEdges(t, edges)
		t.Fatalf("expected return edge getHandler->myFunc")
	}
}

func hasAlias(edges []plugin.Edge, from, to string) bool {
	for _, e := range edges {
		if e.Kind == plugin.EdgeAliases && e.From == from && e.To == to {
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
