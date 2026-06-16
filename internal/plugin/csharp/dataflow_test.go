// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csharp

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_ReturnTracking(t *testing.T) {
	// C# supports delegate references: return MyFunc; (without ())
	// This should produce a return edge.
	src := []byte(`
class App {
    static void MyFunc() {}

    static System.Action GetHandler() {
        return MyFunc;
    }
}
`)
	_, edges, err := parseSource(t, "test.cs", src)
	if err != nil {
		t.Fatal(err)
	}

	if !hasReturn(edges, "GetHandler", "MyFunc") {
		logEdges(t, edges)
		t.Fatalf("expected return edge GetHandler->MyFunc")
	}
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
