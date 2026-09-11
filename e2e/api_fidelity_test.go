// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_APIFidelity(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"api.go": `package api
type Store interface { Read(key string) (string, error) }
type Item struct { A, B int }
func Pair(a, b int, rest ...string) (string, error) { return "", nil }
func Multi() (
    string,
    error,
) { return "", nil }
`,
		"card.tsx":   "export function Card() { return <div>Hello</div>; }",
		"module.mts": "export function typed(key: number): number { return key; }",
		"module.cts": "export function common(key: number): number { return key; }",
		"module.mjs": "export function esm(key) { return key; }",
		"module.cjs": "function cjs(key) { return key; }",
		"lib.rs": `mod left { mod inner { pub fn work() -> i32 { 1 } } }
mod right { mod inner { pub fn work() -> i32 { 2 } } }
struct Named { name: String }
impl Named {
    fn name(&self) -> String { self.name.clone() }
    fn run(&self) -> String { self.name() }
}`,
		"item.h": "#ifndef ITEM_H\n#define ITEM_H\nclass HeaderItem { public: int get(int key); };\n#endif\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := runcodeknitInline(t, root, "parse", root, "--output-mode", "inline", "--format", "json", "--edges", "--header-language", "cpp")
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Symbols []struct {
			ID, Signature string
		}
		Edges  []struct{ From, To, Kind string }
		Errors []json.RawMessage
	}
	if decodeErr := json.Unmarshal([]byte(raw), &output); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if len(output.Errors) != 0 {
		t.Fatalf("valid source produced parse errors: %s", raw)
	}
	signatures := make(map[string]string)
	for _, sym := range output.Symbols {
		_, name, _ := strings.Cut(sym.ID, "::")
		signatures[name] = sym.Signature
	}
	for name, signature := range map[string]string{
		"Pair":       "Pair(a: int, b: int, rest: ...string) -> (string, error)",
		"Store.Read": "Read(key: string) -> (string, error)", "Item.A": "A: int", "Item.B": "B: int",
		"Card": "Card()", "typed": "typed(key: number) -> number", "common": "common(key: number) -> number",
		"esm": "esm(key)", "cjs": "cjs(key)", "HeaderItem.get": "get(key: int) -> int",
		"left.inner": "inner", "left.inner.work": "work() -> i32",
		"right.inner": "inner", "right.inner.work": "work() -> i32",
		"Named.name#field": "name: String", "Named.name": "name(&self) -> String",
	} {
		if signatures[name] != signature {
			t.Errorf("%s: got %q, want %q", name, signatures[name], signature)
		}
	}
	parents := make(map[string][]string)
	namedMethodCalled := false
	for _, edge := range output.Edges {
		_, from, _ := strings.Cut(edge.From, "::")
		_, to, _ := strings.Cut(edge.To, "::")
		if edge.Kind == "contains" {
			parents[to] = append(parents[to], from)
		}
		if edge.Kind == "calls" && from == "Named.run" && to == "Named.name" {
			namedMethodCalled = true
		}
	}
	if !namedMethodCalled {
		t.Error("same-named Rust field must not shadow Named.name()")
	}
	for child, parent := range map[string]string{
		"Item.A": "Item", "Item.B": "Item", "Store.Read": "Store",
		"left.inner": "left", "left.inner.work": "left.inner",
		"right.inner": "right", "right.inner.work": "right.inner",
		"Named.name#field": "Named", "Named.name": "Named",
	} {
		if got := parents[child]; len(got) != 1 || got[0] != parent {
			t.Errorf("%s parents: got %v, want [%s]", child, got, parent)
		}
	}
	// SKT must expose the same restored contract, not only JSON.
	skt, err := runcodeknitInline(t, root, "parse", root, "--output-mode", "inline", "--header-language", "cpp")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(skt, "Pair(a: int, b: int, rest: ...string) -> (string, error)") {
		t.Fatalf("SKT lost the function signature: %s", skt)
	}
	if !strings.Contains(skt, `Multi() -> (\n    string,\n    error,\n)`) {
		t.Fatalf("SKT must keep multiline signatures in one record: %s", skt)
	}
	if signatures["Multi"] != "Multi() -> (\n    string,\n    error,\n)" {
		t.Errorf("JSON must retain source formatting: %q", signatures["Multi"])
	}
}
