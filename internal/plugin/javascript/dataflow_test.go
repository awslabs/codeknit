// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package javascript

import (
	"sort"
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_ObjectPropertyAlias(t *testing.T) {
	src := []byte(`
function processOrder() {}
function processRefund() {}

const handlers = {
  order: processOrder,
  refund: processRefund,
};
`)
	_, edges, err := parseSource(t, src)
	if err != nil {
		t.Fatal(err)
	}

	aliases := collectAliases(edges)
	sort.Strings(aliases)
	want := []string{"handlers.order=processOrder", "handlers.refund=processRefund", "handlers=<unknown>"}
	if !sliceEqual(aliases, want) {
		t.Fatalf("expected aliases %v, got %v", want, aliases)
	}
}

func TestDataflow_AssignmentAlias(t *testing.T) {
	src := []byte(`
function myFunc() {}
const handler = myFunc;
`)
	_, edges, err := parseSource(t, src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasAlias(edges, "handler", "myFunc") {
		t.Fatalf("expected alias handler=myFunc, got: %v", collectAliases(edges))
	}
}

func TestDataflow_ReturnTracking(t *testing.T) {
	src := []byte(`
function myFunc() {}
function getHandler() {
  return myFunc;
}
`)
	_, edges, err := parseSource(t, src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasReturn(edges, "getHandler", "myFunc") {
		logEdges(t, edges)
		t.Fatalf("expected return edge getHandler->myFunc")
	}
}

func collectAliases(edges []plugin.Edge) []string {
	var out []string
	for edgeIndex := range edges {
		e := &edges[edgeIndex]
		if e.Kind == plugin.EdgeAliases {
			out = append(out, e.From+"="+e.To)
		}
	}
	return out
}

func hasAlias(edges []plugin.Edge, from, to string) bool {
	for edgeIndex := range edges {
		e := &edges[edgeIndex]
		if e.Kind == plugin.EdgeAliases && e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func hasReturn(edges []plugin.Edge, from, to string) bool {
	for edgeIndex := range edges {
		e := &edges[edgeIndex]
		if e.Kind == plugin.EdgeReturns && e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func logEdges(t *testing.T, edges []plugin.Edge) {
	t.Helper()
	for edgeIndex := range edges {
		e := &edges[edgeIndex]
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}
}

func sliceEqual(a, b []string) bool {
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
