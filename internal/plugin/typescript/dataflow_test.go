// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package typescript

import (
	"sort"
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_ObjectPropertyAlias(t *testing.T) {
	src := []byte(`
function forOfStmt() {}
function emptyStmt() {}

const handlers = {
  forOf: forOfStmt,
  blank: emptyStmt,
};
`)
	_, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}

	var aliases []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeAliases {
			aliases = append(aliases, e.From+"="+e.To)
		}
	}
	sort.Strings(aliases)

	// Should detect: forOf=forOfStmt, blank=emptyStmt
	want := []string{"blank=emptyStmt", "forOf=forOfStmt"}
	if !strEqual(aliases, want) {
		t.Fatalf("expected aliases %v, got %v", want, aliases)
	}
}

func TestDataflow_AssignmentAlias(t *testing.T) {
	src := []byte(`
function myFunc() {}
const handler = myFunc;
`)
	_, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}

	var aliases []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeAliases {
			aliases = append(aliases, e.From+"="+e.To)
		}
	}

	found := false
	for _, a := range aliases {
		if a == "handler=myFunc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected alias handler=myFunc, got aliases: %v", aliases)
	}
}

func TestDataflow_ReturnTracking(t *testing.T) {
	src := []byte(`
function myFunc() {}
function getHandler() {
  return myFunc;
}
`)
	_, edges, err := parseSource(t, "test.ts", src)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range edges {
		if e.Kind == plugin.EdgeReturns && e.From == "getHandler" && e.To == "myFunc" {
			found = true
		}
	}
	if !found {
		t.Logf("edges:")
		for _, e := range edges {
			t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
		}
		t.Fatalf("expected return edge getHandler->myFunc")
	}
}

func strEqual(a, b []string) bool {
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
