// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rust

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/plugin"

	"pgregory.net/rapid"
)

// parseSource writes src to a temp file and parses it via the plugin.
func parseSource(t *testing.T, filename string, src []byte) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return NewPlugin().Parse(path)
}

func TestParseSource_Symbols(t *testing.T) {
	src := []byte(`
pub fn greet(name: &str) -> String {
    format!("Hello, {}", name)
}

fn helper() {}

pub struct Config {
    pub name: String,
    value: i32,
}

pub enum Color {
    Red,
    Green,
    Blue,
}

pub trait Drawable {
    fn draw(&self);
}

type Alias = i32;

pub mod utils {}

pub const MAX: i32 = 100;

static mut COUNTER: i32 = 0;

macro_rules! my_macro {
    () => {};
}
`)

	symbols, _, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected non-empty symbols")
	}

	type catKind struct {
		Category plugin.SymbolCategory
		Kind     string
	}
	found := make(map[string]catKind)
	for _, s := range symbols {
		found[s.Name] = catKind{s.Category, s.Kind}
	}

	expect := map[string]catKind{
		"greet":    {plugin.CategoryCallable, "function"},
		"helper":   {plugin.CategoryCallable, "function"},
		"Config":   {plugin.CategoryType, "struct"},
		"Color":    {plugin.CategoryType, "enum"},
		"Drawable": {plugin.CategoryType, "trait"},
		"Alias":    {plugin.CategoryType, "type_alias"},
		"utils":    {plugin.CategoryModule, "module"},
		"MAX":      {plugin.CategoryValue, "constant"},
		"COUNTER":  {plugin.CategoryValue, "variable"},
		"my_macro": {plugin.CategoryCallable, "macro"},
	}

	for name, want := range expect {
		got, ok := found[name]
		if !ok {
			t.Errorf("missing symbol %q (expected %s/%s)", name, want.Category, want.Kind)
			continue
		}
		if got.Category != want.Category || got.Kind != want.Kind {
			t.Errorf("symbol %q: got %s/%s, want %s/%s", name, got.Category, got.Kind, want.Category, want.Kind)
		}
	}
}

func TestParseSource_PubDetection(t *testing.T) {
	src := []byte(`
pub fn public_fn() {}
fn private_fn() {}

pub struct PubStruct {}
struct PrivStruct {}

pub const PUB_CONST: i32 = 1;
const PRIV_CONST: i32 = 2;

pub static PUB_STATIC: i32 = 1;
static PRIV_STATIC: i32 = 2;
`)
	symbols, _, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	wantPub := map[string]string{
		"public_fn":   "true",
		"private_fn":  "",
		"PubStruct":   "true",
		"PrivStruct":  "",
		"PUB_CONST":   "true",
		"PRIV_CONST":  "",
		"PUB_STATIC":  "true",
		"PRIV_STATIC": "",
	}

	for _, sym := range symbols {
		want, ok := wantPub[sym.Name]
		if !ok {
			continue
		}
		got := sym.Properties["pub"]
		if got != want {
			t.Errorf("symbol %q: pub got %q, want %q", sym.Name, got, want)
		}
	}
}

func TestParseSource_AsyncUnsafe(t *testing.T) {
	src := []byte(`
async fn async_fn() {}
unsafe fn unsafe_fn() {}
fn normal_fn() {}
`)
	symbols, _, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		switch sym.Name {
		case "async_fn":
			if sym.Properties["async"] != "true" {
				t.Errorf("async_fn should have async=true, got %q", sym.Properties["async"])
			}
			if sym.Properties["unsafe"] != "" {
				t.Errorf("async_fn should have unsafe=\"\", got %q", sym.Properties["unsafe"])
			}
		case "unsafe_fn":
			if sym.Properties["unsafe"] != "true" {
				t.Errorf("unsafe_fn should have unsafe=true, got %q", sym.Properties["unsafe"])
			}
			if sym.Properties["async"] != "" {
				t.Errorf("unsafe_fn should have async=\"\", got %q", sym.Properties["async"])
			}
		case "normal_fn":
			if sym.Properties["async"] != "" {
				t.Errorf("normal_fn should have async=\"\", got %q", sym.Properties["async"])
			}
			if sym.Properties["unsafe"] != "" {
				t.Errorf("normal_fn should have unsafe=\"\", got %q", sym.Properties["unsafe"])
			}
		}
	}
}

func TestParseSource_StaticMutable(t *testing.T) {
	src := []byte(`
static mut MUTABLE_COUNTER: i32 = 0;
static IMMUTABLE_COUNTER: i32 = 0;
`)
	symbols, _, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	for _, sym := range symbols {
		switch sym.Name {
		case "MUTABLE_COUNTER":
			if sym.Properties["mutable"] != "true" {
				t.Errorf("MUTABLE_COUNTER should have mutable=true, got %q", sym.Properties["mutable"])
			}
		case "IMMUTABLE_COUNTER":
			if sym.Properties["mutable"] != "" {
				t.Errorf("IMMUTABLE_COUNTER should have mutable=\"\", got %q", sym.Properties["mutable"])
			}
		}
	}
}

func TestParseSource_ImplMethods(t *testing.T) {
	src := []byte(`
struct Foo {}

impl Foo {
    pub fn new() -> Self {
        Foo {}
    }

    fn helper(&self) {}
}
`)
	symbols, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	// Check method symbols exist.
	foundNew := false
	foundHelper := false
	for _, sym := range symbols {
		if sym.Name == "new" && sym.Kind == "method" {
			foundNew = true
			if sym.Properties["pub"] != "true" {
				t.Errorf("new should have pub=true, got %q", sym.Properties["pub"])
			}
		}
		if sym.Name == "helper" && sym.Kind == "method" {
			foundHelper = true
			if sym.Properties["pub"] != "" {
				t.Errorf("helper should have pub=\"\", got %q", sym.Properties["pub"])
			}
		}
	}
	if !foundNew {
		t.Error("missing method symbol 'new'")
	}
	if !foundHelper {
		t.Error("missing method symbol 'helper'")
	}

	// Check contains edges.
	var containsTargets []string
	for _, e := range edges {
		if e.From == "Foo" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Foo.helper", "Foo.new"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_TraitMethods(t *testing.T) {
	src := []byte(`
trait Animal {
    fn speak(&self) -> String;
    fn name(&self) -> &str;
}
`)
	symbols, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	// Check trait method symbols.
	foundSpeak := false
	foundName := false
	for _, sym := range symbols {
		if sym.Name == "speak" && sym.Kind == "method" {
			foundSpeak = true
		}
		if sym.Name == "name" && sym.Kind == "method" {
			foundName = true
		}
	}
	if !foundSpeak {
		t.Error("missing trait method 'speak'")
	}
	if !foundName {
		t.Error("missing trait method 'name'")
	}

	// Check contains edges from trait.
	var containsTargets []string
	for _, e := range edges {
		if e.From == "Animal" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Animal.name", "Animal.speak"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_ImplementsEdge(t *testing.T) {
	src := []byte(`
trait Drawable {
    fn draw(&self);
}

struct Circle {}

impl Drawable for Circle {
    fn draw(&self) {}
}
`)
	_, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	foundImpl := false
	for _, e := range edges {
		if e.From == "Circle" && e.To == "Drawable" && e.Kind == plugin.EdgeImplements {
			foundImpl = true
		}
	}
	if !foundImpl {
		t.Error("expected implements edge from Circle to Drawable")
	}
}

func TestParseSource_CallEdges(t *testing.T) {
	src := []byte(`
fn caller() {
    helper();
    std::io::read();
}

fn helper() {}
`)
	_, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	var callTargets []string
	for _, e := range edges {
		if e.Kind == plugin.EdgeCalls && e.From == "caller" {
			callTargets = append(callTargets, e.To)
		}
	}
	sort.Strings(callTargets)
	want := []string{"helper"}
	if len(callTargets) != len(want) {
		t.Fatalf("expected call targets %v, got %v", want, callTargets)
	}
	for i := range want {
		if callTargets[i] != want[i] {
			t.Errorf("call target[%d]: got %q, want %q", i, callTargets[i], want[i])
		}
	}
}

func TestParseSource_StructContainsFields(t *testing.T) {
	src := []byte(`
pub struct Point {
    pub x: f64,
    pub y: f64,
}
`)
	_, edges, err := parseSource(t, "test.rs", src)
	if err != nil {
		t.Fatal(err)
	}

	var containsTargets []string
	for _, e := range edges {
		if e.From == "Point" && e.Kind == plugin.EdgeContains {
			containsTargets = append(containsTargets, e.To)
		}
	}
	sort.Strings(containsTargets)
	want := []string{"Point.x", "Point.y"}
	if len(containsTargets) != len(want) {
		t.Fatalf("expected contains targets %v, got %v", want, containsTargets)
	}
	for i := range want {
		if containsTargets[i] != want[i] {
			t.Errorf("contains target[%d]: got %q, want %q", i, containsTargets[i], want[i])
		}
	}
}

func TestParseSource_SyntaxError_ReturnsSyntaxWarning(t *testing.T) {
	src := []byte(`fn broken( {`)
	_, _, err := parseSource(t, "bad.rs", src)
	if err == nil {
		t.Fatal("expected SyntaxWarning for syntax error")
	}
	if !strings.Contains(err.Error(), "bad.rs") {
		t.Errorf("error should contain file path, got: %s", err)
	}
}

func TestParseSource_SyntaxError_NoEnforce(t *testing.T) {
	src := []byte(`fn broken( {`)
	symbols, _, err := parseSource(t, "bad.rs", src)
	var sw *plugin.SyntaxError
	if err != nil && !errors.As(err, &sw) {
		t.Fatalf("expected nil or SyntaxWarning, got: %v", err)
	}
	if symbols == nil {
		t.Fatal("expected non-nil symbols for partial-error file")
	}
}

func TestExtensions(t *testing.T) {
	p := NewPlugin()
	exts := p.Extensions()
	if len(exts) != 1 || exts[0] != ".rs" {
		t.Errorf("expected [.rs], got %v", exts)
	}
}

// Property 12: Rust pub detection
// For any Rust source file, a Symbol should have Properties["pub"] == "true"
// if and only if the declaration has a visibility_modifier (pub keyword).
// **Validates: Requirement 5.2**
func TestProperty_RustPubDetection(tt *testing.T) {
	rapid.Check(tt, func(t *rapid.T) {
		// Generate random function names.
		pubName := genIdent().Draw(t, "pubName")
		privName := genIdent().Draw(t, "privName")

		// Ensure names are unique.
		if pubName == privName {
			return
		}

		var b strings.Builder
		b.WriteString("pub fn " + pubName + "() {}\n\n")
		b.WriteString("fn " + privName + "() {}\n")

		src := []byte(b.String())
		symbols, _, err := parseSource(tt, "gen.rs", src)
		if err != nil {
			t.Fatalf("parse error: %s\nsource:\n%s", err, src)
		}

		foundPub := false
		foundPriv := false
		for _, sym := range symbols {
			if sym.Name == pubName {
				foundPub = true
				if sym.Properties["pub"] != "true" {
					t.Errorf("pub function %q should have pub=true, got %q", pubName, sym.Properties["pub"])
				}
			}
			if sym.Name == privName {
				foundPriv = true
				if sym.Properties["pub"] != "" {
					t.Errorf("private function %q should have pub=\"\", got %q", privName, sym.Properties["pub"])
				}
			}
		}
		if !foundPub {
			t.Errorf("missing pub function %q in symbols", pubName)
		}
		if !foundPriv {
			t.Errorf("missing private function %q in symbols", privName)
		}
	})
}

// genIdent generates a valid lowercase Rust identifier (a-z, 3-8 chars).
func genIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(3, 8).Draw(t, "len")
		chars := make([]byte, n)
		for i := range chars {
			chars[i] = "abcdefghijklmnopqrstuvwxyz"[rapid.IntRange(0, 25).Draw(t, "ch")]
		}
		return string(chars)
	})
}
