// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package java

import (
	"testing"

	"codeknit/internal/plugin"
)

func TestDataflow_ReturnTracking(t *testing.T) {
	// Java doesn't support bare function references (myFunc without ()).
	// Method references use :: syntax (App::myFunc) which is a different
	// AST node kind. The dataflow extractor won't catch these yet.
	// This test verifies the extractor doesn't produce false edges.
	src := []byte(`
public class App {
    static Runnable myFunc() { return null; }

    static Runnable getHandler() {
        return myFunc();
    }
}
`)
	_, edges, err := parseSource(t, "App.java", src)
	if err != nil {
		t.Fatal(err)
	}

	// myFunc() is a call, not a reference — no return edge expected.
	for _, e := range edges {
		if e.Kind == plugin.EdgeReturns {
			t.Fatalf("unexpected return edge: %s -> %s", e.From, e.To)
		}
	}
}
