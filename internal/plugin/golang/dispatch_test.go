// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package golang

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestFileCallEdges_MethodCallOnVariable(t *testing.T) {
	// This tests that method calls like table.Dispatch() produce call edges.
	src := []byte(`
package main

type MyTable map[string]func()

func (t MyTable) Dispatch(key string) {
}

func WalkAll(table MyTable) {
	table.Dispatch("foo")
}
`)
	_, edges, err := parseSource(t, "test.go", src)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "WalkAll" && e.To == "Dispatch" {
			found = true
		}
	}
	if !found {
		t.Log("edges:")
		for _, e := range edges {
			t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
		}
		t.Fatalf("expected call edge WalkAll -> Dispatch")
	}
}
