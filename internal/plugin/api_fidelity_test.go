// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/clang"
	"codeknit/internal/plugin/cpp"
	"codeknit/internal/plugin/csharp"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/php"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/ruby"
	"codeknit/internal/plugin/rust"
	"codeknit/internal/plugin/scala"
	"codeknit/internal/plugin/typescript"
)

// TestAPISignatureBaseline is the minimum signature contract for every
// supported language. More complex constructs have separate regressions below.
func TestAPISignatureBaseline(t *testing.T) {
	for _, tc := range []struct {
		parser                            plugin.LanguagePlugin
		filename, source, name, signature string
	}{
		{clang.NewPlugin(), "api.c", "int read_value(int key) { return key; }", "read_value", "read_value(key: int) -> int"},
		{cpp.NewPlugin(), "api.cpp", "int read_value(int key) { return key; }", "read_value", "read_value(key: int) -> int"},
		{csharp.NewPlugin(), "api.cs", "class Store { public int Read(int key) { return key; } }", "Store.Read", "Read(key: int) -> int"},
		{golang.NewPlugin(), "api.go", "package api\nfunc Read(key int) int { return key }", "Read", "Read(key: int) -> int"},
		{java.NewPlugin(), "Api.java", "class Store { int read(int key) { return key; } }", "Store.read", "read(key: int) -> int"},
		{javascript.NewPlugin(), "api.js", "function read(key) { return key; }", "read", "read(key)"},
		{php.NewPlugin(), "api.php", "<?php function read_value(int $key): int { return $key; }", "read_value", "read_value(key: int) -> int"},
		{python.NewPlugin(), "api.py", "def read(key: int) -> int:\n    return key\n", "read", "read(key: int) -> int"},
		{ruby.NewPlugin(), "api.rb", "def read(key)\n  key\nend\n", "read", "read(key)"},
		{rust.NewPlugin(), "api.rs", "fn read(key: i32) -> i32 { key }", "read", "read(key: i32) -> i32"},
		{scala.NewPlugin(), "api.scala", "def read(key: Int): Int = key", "read", "read(key: Int) -> Int"},
		{typescript.NewPlugin(), "api.ts", "function read(key: number): number { return key; }", "read", "read(key: number) -> number"},
	} {
		t.Run(tc.filename, func(t *testing.T) {
			symbols, _, err := parseSrc(t, tc.parser, tc.filename, []byte(tc.source))
			if err != nil {
				t.Fatal(err)
			}
			for _, sym := range symbols {
				if sym.EffectiveScopedName() == tc.name && sym.Signature == tc.signature {
					return
				}
			}
			t.Fatalf("missing %s with signature %q: %v", tc.name, tc.signature, symbols)
		})
	}
}

func TestAPIFidelity(t *testing.T) {
	tests := []struct {
		parser     plugin.LanguagePlugin
		signatures map[string]string
		name       string
		filename   string
		source     string
		contains   [][2]string
	}{
		{
			name: "Go parameters and results", filename: "api.go", parser: golang.NewPlugin(),
			source: `package api
func Pair(a, b int, rest ...string) (string, error) { return "", nil }
func Named() (value string, err error) { return "", nil }
func Apply(fn func(int) error, items []map[string]*Item) chan<- Item { return nil }
func Unnamed(func(int) error, ...string) {}
func Generic[T any](value T) Box[T] { panic(0) }
func Grouped[A, B any](a A, b B) {}
type Item struct{}
type Box[T any] struct{}
func (*Box[T]) Read(a, b int, rest ...string) (T, error) { panic(0) }
`,
			signatures: map[string]string{
				"Pair":      "Pair(a: int, b: int, rest: ...string) -> (string, error)",
				"Named":     "Named() -> (value string, err error)",
				"Apply":     "Apply(fn: func(int) error, items: []map[string]*Item) -> chan<- Item",
				"Unnamed":   "Unnamed(func(int) error, ...string)",
				"Generic":   "Generic(value: T) -> Box[T]",
				"Box.Read":  "Read(a: int, b: int, rest: ...string) -> (T, error)",
				"Grouped.A": "A", "Grouped.B": "B",
			},
			contains: [][2]string{{"Box", "Box.Read"}, {"Grouped", "Grouped.A"}, {"Grouped", "Grouped.B"}},
		},
		{
			name: "Go interface contracts and fields", filename: "api.go", parser: golang.NewPlugin(),
			source: "package api\n" +
				"type Base struct{}\n" +
				"type Item struct { A, B int; *Base; Name string `json:\"name\"` }\n" +
				"type Store interface { Read(key string) (string, error); Write(...string) error }\n",
			signatures: map[string]string{
				"Item.A": "A: int", "Item.B": "B: int", "Item.Base": "Base: *Base", "Item.Name": "Name: string",
				"Store.Read": "Read(key: string) -> (string, error)", "Store.Write": "Write(...string) -> error",
			},
			contains: [][2]string{{"Item", "Item.A"}, {"Item", "Item.B"}, {"Item", "Item.Base"}, {"Store", "Store.Read"}},
		},
		{
			name: "Rust inline modules", filename: "api.rs", parser: rust.NewPlugin(),
			source: `pub mod first {
    pub fn work() -> i32 { 1 }
    pub mod inner {
        pub struct Item { pub name: String }
        impl Item { pub fn read(&mut self, mut n: usize, (x,y): (i32,i32)) -> &str { "" } }
        pub trait Store { fn get(&self, key: &str) -> String; }
        pub const LIMIT: usize = 1;
    }
}
pub mod second { pub fn work() -> i32 { 2 } }
mod external;
`,
			signatures: map[string]string{
				"first": "first", "first.work": "work() -> i32", "first.inner": "inner",
				"first.inner.Item.name#field": "name: String",
				"first.inner.Item.read":       "read(&mut self, mut n: usize, (x,y): (i32,i32)) -> &str",
				"first.inner.Store.get":       "get(&self, key: &str) -> String",
				"first.inner.LIMIT":           "LIMIT", "second.work": "work() -> i32", "external": "external",
			},
			contains: [][2]string{
				{"first", "first.work"},
				{"first", "first.inner"},
				{"first.inner", "first.inner.Item"},
				{"first.inner.Item", "first.inner.Item.name#field"},
				{"first.inner.Item", "first.inner.Item.read"},
				{"first.inner.Store", "first.inner.Store.get"},
				{"second", "second.work"},
			},
		},
		{
			name: "TSX", filename: "card.tsx", parser: typescript.NewPlugin(),
			source: `export function Card(name: string): JSX.Element { return <div>{name}</div>; }
export const Badge = (label: string) => <span>{label}</span>;`,
			signatures: map[string]string{
				"Card": "Card(name: string) -> JSX.Element", "Badge": "Badge(label: string)",
			},
		},
		{
			name: "TypeScript namespace wrapper", filename: "api.ts", parser: typescript.NewPlugin(),
			source: `namespace Utils {
    export function read(key: number): number { return key; }
}`,
			signatures: map[string]string{"Utils": "Utils", "read": "read(key: number) -> number"},
		},
		{
			name: "Python annotations and calling convention", filename: "api.py", parser: python.NewPlugin(),
			source: `def typed(x: str, /, y: int = 0, *args: int, **kwargs: bool) -> str:
    return x
def keyword(*, flag: bool = False) -> bool:
    return flag
def receiver_name(self: str) -> str:
    return self
class Store:
    def get(self, key: str) -> str:
        return key
    @staticmethod
    def identity(self: str) -> str:
        return self
`,
			signatures: map[string]string{
				"typed":         "typed(x: str, /, y: int = 0, *args: int, **kwargs: bool) -> str",
				"keyword":       "keyword(*, flag: bool = False) -> bool",
				"receiver_name": "receiver_name(self: str) -> str", "Store.get": "get(key: str) -> str",
				"Store.identity": "identity(self: str) -> str",
			},
		},
		{
			name: "C header guards", filename: "api.h", parser: clang.NewPlugin(),
			source: `#ifndef API_H
#define API_H
int read_value(int key);
#if FEATURE
int enabled(void);
#else
int disabled(void);
#endif
#endif
`,
			signatures: map[string]string{
				"read_value": "read_value(key: int) -> int", "enabled": "enabled() -> int", "disabled": "disabled() -> int",
			},
		},
		{
			name: "C++ header guards and linkage", filename: "api.hpp", parser: cpp.NewPlugin(),
			source: `#ifndef API_HPP
#define API_HPP
class Store { public: int get(int key); };
extern "C" { int read_value(int key); }
#endif
`,
			signatures: map[string]string{
				"Store": "Store", "Store.get": "get(key: int) -> int", "read_value": "read_value(key: int) -> int",
			},
			contains: [][2]string{{"Store", "Store.get"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, fingerprint := range []bool{false, true} {
				t.Run(fmt.Sprintf("fingerprint=%t", fingerprint), func(t *testing.T) {
					path := filepath.Join(t.TempDir(), tc.filename)
					if err := os.WriteFile(path, []byte(tc.source), 0o600); err != nil {
						t.Fatal(err)
					}
					symbols, edges, err := plugin.ParseFile(tc.parser, path, fingerprint)
					if err != nil {
						t.Fatalf("valid source must parse without warnings: %v", err)
					}
					for name, signature := range tc.signatures {
						found := false
						for _, sym := range symbols {
							if sym.EffectiveScopedName() == name && sym.Signature == signature {
								found = true
								if sym.Span[0] < 1 || sym.Span[1] < sym.Span[0] {
									t.Errorf("%s has invalid source span %v", name, sym.Span)
								}
							}
						}
						if !found {
							t.Errorf("missing %s with signature %q; got %v", name, signature, symbols)
						}
					}
					for _, pair := range tc.contains {
						want := plugin.Edge{From: pair[0], To: pair[1], Kind: plugin.EdgeContains}
						if !slices.ContainsFunc(edges, func(e plugin.Edge) bool {
							return e.From == want.From && e.To == want.To && e.Kind == want.Kind
						}) {
							t.Errorf("missing containment %v", want)
						}
					}
				})
			}
		})
	}
}

func TestTypeScriptGrammarVariantsConcurrent(t *testing.T) {
	parser := typescript.NewPlugin()
	for i := range 12 {
		for _, tc := range []struct{ ext, source string }{
			{".ts", "export function cast(x: unknown) { return <number>x; }"},
			{".mts", "export function cast(x: unknown) { return <number>x; }"},
			{".cts", "export function cast(x: unknown) { return <number>x; }"},
			{".tsx", "export function Card() { return <div/>; }"},
		} {
			t.Run(fmt.Sprintf("%d%s", i, tc.ext), func(t *testing.T) {
				t.Parallel()
				symbols, _, err := parseSrc(t, parser, "api"+tc.ext, []byte(tc.source))
				if err != nil || len(symbols) != 1 {
					t.Fatalf("grammar selection failed: %v, symbols: %v", err, symbols)
				}
			})
		}
	}
}
