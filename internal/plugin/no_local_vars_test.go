// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/ruby"
	"codeknit/internal/plugin/typescript"
)

// TestNoLocalVars_JS verifies that destructured variables inside function
// bodies are NOT collected as symbols.
func TestNoLocalVars_JS(t *testing.T) {
	src := []byte(`
// Top-level destructuring — SHOULD be collected
const { topA, topB } = config;
const [topX, topY] = coords;

function process(data) {
    // Local destructuring — should NOT be collected
    const { localA, localB } = data;
    const [localX, localY] = data.items;
    let localVar = 42;
    const { nested: { deepLocal } } = data;
}

class MyClass {
    method() {
        // Local inside method — should NOT be collected
        const { methodLocal } = this.data;
    }
}
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)

	// Top-level destructured names SHOULD be present
	for _, want := range []string{"topA", "topB", "topX", "topY"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing top-level destructured symbol %q", want)
		}
	}

	// Local variables should NOT be present
	for _, bad := range []string{"localA", "localB", "localX", "localY", "localVar", "deepLocal", "methodLocal"} {
		if findSymbol(symbols, bad) != nil {
			t.Errorf("local variable %q should NOT be collected as a symbol", bad)
		}
	}

	// Function and class should be present
	if findSymbol(symbols, "process") == nil {
		t.Error("missing function 'process'")
	}
	if findSymbol(symbols, "MyClass") == nil {
		t.Error("missing class 'MyClass'")
	}
}

// TestNoLocalVars_TS verifies the same for TypeScript.
func TestNoLocalVars_TS(t *testing.T) {
	src := []byte(`
const { topA, topB } = config;

function process(data: any): void {
    const { localA, localB }: { localA: string; localB: number } = data;
    const [localX, localY] = data.items;
}
`)
	symbols, _ := parseWith(t, typescript.NewPlugin(), "test.ts", src)

	for _, want := range []string{"topA", "topB"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing top-level destructured symbol %q", want)
		}
	}

	for _, bad := range []string{"localA", "localB", "localX", "localY"} {
		if findSymbol(symbols, bad) != nil {
			t.Errorf("local variable %q should NOT be collected as a symbol", bad)
		}
	}
}

// TestNoLocalVars_Python verifies that tuple-unpacked variables inside
// function bodies are NOT collected.
func TestNoLocalVars_Python(t *testing.T) {
	src := []byte(`
# Top-level — SHOULD be collected
top_x, top_y = 1, 2
TOP_NAME = "hello"

def process(data):
    # Local — should NOT be collected
    local_a, local_b = data
    local_var = 42

class MyClass:
    # Class-level assignment — should NOT be collected (it's inside class body)
    class_var = "test"
    def method(self):
        method_local = 1
`)
	symbols, _ := parseWith(t, python.NewPlugin(), "test.py", src)

	for _, want := range []string{"top_x", "top_y", "TOP_NAME"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing top-level symbol %q", want)
		}
	}

	for _, bad := range []string{"local_a", "local_b", "local_var", "method_local"} {
		if findSymbol(symbols, bad) != nil {
			t.Errorf("local variable %q should NOT be collected as a symbol", bad)
		}
	}

	if findSymbol(symbols, "process") == nil {
		t.Error("missing function 'process'")
	}
}

// TestNoLocalVars_Ruby verifies that multi-assigned variables inside
// method bodies are NOT collected.
func TestNoLocalVars_Ruby(t *testing.T) {
	src := []byte(`
# Top-level — SHOULD be collected
top_a, top_b = 1, 2
TOP_CONST = "hello"

def process(data)
    # Local — should NOT be collected
    local_a, local_b = data
    local_var = 42
end
`)
	symbols, _ := parseWith(t, ruby.NewPlugin(), "test.rb", src)

	for _, want := range []string{"top_a", "top_b", "TOP_CONST"} {
		if findSymbol(symbols, want) == nil {
			t.Errorf("missing top-level symbol %q", want)
		}
	}

	for _, bad := range []string{"local_a", "local_b", "local_var"} {
		if s := findSymbol(symbols, bad); s != nil {
			t.Errorf("local variable %q should NOT be collected as a symbol (got %s/%s)", bad, s.Category, s.Kind)
		}
	}
}

// TestNoLocalVars_JS_ExportedDestructuring verifies that exported
// destructuring at top level works but non-exported inside functions doesn't.
func TestNoLocalVars_JS_ExportedDestructuring(t *testing.T) {
	src := []byte(`
export const { API_KEY, SECRET } = process.env;

function init() {
    const { LOCAL_KEY } = getConfig();
}
`)
	symbols, _ := parseWith(t, javascript.NewPlugin(), "test.js", src)

	if findSymbol(symbols, "API_KEY") == nil {
		t.Error("missing exported destructured 'API_KEY'")
	}
	if findSymbol(symbols, "SECRET") == nil {
		t.Error("missing exported destructured 'SECRET'")
	}

	apiKey := findSymbol(symbols, "API_KEY")
	if apiKey != nil && apiKey.Properties["exported"] != "true" {
		t.Error("API_KEY should be exported")
	}

	if findSymbol(symbols, "LOCAL_KEY") != nil {
		t.Error("local destructured 'LOCAL_KEY' should NOT be collected")
	}
}

// Ensure the test helpers are used so the compiler doesn't complain.
var _ = []plugin.LanguagePlugin{
	javascript.NewPlugin(),
	typescript.NewPlugin(),
	python.NewPlugin(),
	ruby.NewPlugin(),
}
