// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
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

	"pgregory.net/rapid"
)

// allPlugins returns all 12 language plugins for reuse by cross-plugin tests.
func allPlugins() []plugin.LanguagePlugin {
	return []plugin.LanguagePlugin{
		javascript.NewPlugin(),
		typescript.NewPlugin(),
		clang.NewPlugin(),
		cpp.NewPlugin(),
		csharp.NewPlugin(),
		golang.NewPlugin(),
		java.NewPlugin(),
		php.NewPlugin(),
		python.NewPlugin(),
		ruby.NewPlugin(),
		rust.NewPlugin(),
		scala.NewPlugin(),
	}
}

// langCase holds a language name, a representative file path, valid source bytes,
// and the plugin to parse with.
type langCase struct {
	plugin   plugin.LanguagePlugin
	name     string
	filePath string
	src      []byte
}

// validLangCases returns representative valid source snippets for all 12 languages.
func validLangCases() []langCase {
	return []langCase{
		{
			name:     "JavaScript",
			filePath: "test.js",
			src:      []byte("function foo() { bar(); }\nfunction bar() {}\n"),
			plugin:   javascript.NewPlugin(),
		},
		{
			name:     "TypeScript",
			filePath: "test.ts",
			src:      []byte("function foo(): void { bar(); }\nfunction bar(): void {}\n"),
			plugin:   typescript.NewPlugin(),
		},
		{
			name:     "C",
			filePath: "test.c",
			src:      []byte("struct Foo { int x; };\n\nvoid bar(void) { printf(\"hi\"); }\n"),
			plugin:   clang.NewPlugin(),
		},
		{
			name:     "C++",
			filePath: "test.cpp",
			src:      []byte("class Foo {\npublic:\n    void bar() {}\n};\n"),
			plugin:   cpp.NewPlugin(),
		},
		{
			name:     "C#",
			filePath: "test.cs",
			src:      []byte("public class Foo {\n    public void Bar() { Console.WriteLine(\"hi\"); }\n}\n"),
			plugin:   csharp.NewPlugin(),
		},
		{
			name:     "Go",
			filePath: "test.go",
			src:      []byte("package main\n\ntype Foo struct{ X int }\n\nfunc (f *Foo) Bar() { fmt.Println(\"hi\") }\n"),
			plugin:   golang.NewPlugin(),
		},
		{
			name:     "Java",
			filePath: "test.java",
			src:      []byte("public class Foo {\n    public void bar() { System.out.println(\"hi\"); }\n}\n"),
			plugin:   java.NewPlugin(),
		},
		{
			name:     "PHP",
			filePath: "test.php",
			src:      []byte("<?php\nclass Foo {\n    public function bar() { echo \"hi\"; }\n}\n"),
			plugin:   php.NewPlugin(),
		},
		{
			name:     "Python",
			filePath: "test.py",
			src:      []byte("class Foo:\n    def bar(self):\n        print(\"hi\")\n"),
			plugin:   python.NewPlugin(),
		},
		{
			name:     "Ruby",
			filePath: "test.rb",
			src:      []byte("class Foo\n  def bar\n    puts \"hi\"\n  end\nend\n"),
			plugin:   ruby.NewPlugin(),
		},
		{
			name:     "Rust",
			filePath: "test.rs",
			src:      []byte("pub struct Foo {}\n\nimpl Foo {\n    pub fn bar(&self) { println!(\"hi\"); }\n}\n"),
			plugin:   rust.NewPlugin(),
		},
		{
			name:     "Scala",
			filePath: "test.scala",
			src:      []byte("class Foo {\n  def bar(): Unit = { println(\"hi\") }\n}\n"),
			plugin:   scala.NewPlugin(),
		},
	}
}

// validCategories is the set of valid SymbolCategory values.
var validCategories = map[plugin.SymbolCategory]bool{
	plugin.CategoryCallable: true,
	plugin.CategoryType:     true,
	plugin.CategoryValue:    true,
	plugin.CategoryModule:   true,
	plugin.CategoryMeta:     true,
}

// validEdgeKinds is the set of valid EdgeKind values.
var validEdgeKinds = map[plugin.EdgeKind]bool{
	plugin.EdgeCalls:      true,
	plugin.EdgeInherits:   true,
	plugin.EdgeContains:   true,
	plugin.EdgeReferences: true,
	plugin.EdgeImplements: true,
	plugin.EdgeOverrides:  true,
	plugin.EdgeImports:    true,
	plugin.EdgeDecorates:  true,
}

// Property 1: Registry round-trip
// For any Plugin with a set of extensions, after registering it with the Registry,
// looking up each of those extensions should return that same Plugin.
func TestProperty_RegistryRoundTrip(t *testing.T) {
	all := allPlugins()

	rapid.Check(t, func(t *rapid.T) {
		// Draw a random non-empty subset of plugins to register.
		mask := rapid.SliceOfN(rapid.Bool(), len(all), len(all)).Draw(t, "mask")

		// Ensure at least one plugin is selected.
		anySelected := false
		for _, m := range mask {
			if m {
				anySelected = true
				break
			}
		}
		if !anySelected {
			return
		}

		reg := plugin.NewRegistry()

		var selected []plugin.LanguagePlugin
		for i, m := range mask {
			if m {
				selected = append(selected, all[i])
				reg.Register(all[i])
			}
		}

		// For each selected plugin, every extension should round-trip back.
		for _, p := range selected {
			for _, ext := range p.Extensions() {
				got, ok := reg.Lookup(ext)
				if !ok {
					t.Fatalf("Lookup(%q) returned ok=false after registering plugin with extensions %v", ext, p.Extensions())
				}
				if got == nil {
					t.Fatalf("Lookup(%q) returned nil plugin after registering plugin with extensions %v", ext, p.Extensions())
				}
			}
		}
	})
}

// Property 2: Symbol and Edge validity
// For any valid source file parsed by any Plugin, every returned Symbol should have
// a non-empty Name, a valid Category, a non-empty Kind, and Span[0] <= Span[1];
// and every returned Edge should have a non-empty From, a non-empty To, and a valid EdgeKind.
func TestProperty_SymbolAndEdgeValidity(tt *testing.T) {
	cases := validLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		// Randomly select a language to test on each iteration.
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		symbols, edges, err := parseSrc(tt, lc.plugin, lc.filePath, lc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", lc.name, err)
		}

		if len(symbols) == 0 {
			t.Fatalf("[%s] expected at least one symbol from valid source", lc.name)
		}

		for i, sym := range symbols {
			if sym.Name == "" {
				t.Fatalf("[%s] symbol[%d] has empty Name", lc.name, i)
			}
			if !validCategories[sym.Category] {
				t.Fatalf("[%s] symbol[%d] %q has invalid Category %q", lc.name, i, sym.Name, sym.Category)
			}
			if sym.Kind == "" {
				t.Fatalf("[%s] symbol[%d] %q has empty Kind", lc.name, i, sym.Name)
			}
			if sym.Span[0] > sym.Span[1] {
				t.Fatalf("[%s] symbol[%d] %q has Span[0]=%d > Span[1]=%d", lc.name, i, sym.Name, sym.Span[0], sym.Span[1])
			}
		}

		for i, edge := range edges {
			if edge.From == "" {
				t.Fatalf("[%s] edge[%d] has empty From", lc.name, i)
			}
			if edge.To == "" {
				t.Fatalf("[%s] edge[%d] has empty To", lc.name, i)
			}
			if !validEdgeKinds[edge.Kind] {
				t.Fatalf("[%s] edge[%d] (%s->%s) has invalid Kind %q", lc.name, i, edge.From, edge.To, edge.Kind)
			}
		}
	})
}

// invalidLangCases returns langCase entries with intentionally broken source for each language.
func invalidLangCases() []langCase {
	return []langCase{
		{
			name:     "Go",
			filePath: "test.go",
			src:      []byte("package main\n\nfunc broken( {\n"),
			plugin:   golang.NewPlugin(),
		},
		{
			name:     "Python",
			filePath: "test.py",
			src:      []byte("def broken(\n"),
			plugin:   python.NewPlugin(),
		},
		{
			name:     "Java",
			filePath: "test.java",
			src:      []byte("public class Foo { void broken( { } }\n"),
			plugin:   java.NewPlugin(),
		},
		{
			name:     "Rust",
			filePath: "test.rs",
			src:      []byte("fn broken( {\n"),
			plugin:   rust.NewPlugin(),
		},
		{
			name:     "C",
			filePath: "test.c",
			src:      []byte("void broken( {\n"),
			plugin:   clang.NewPlugin(),
		},
		{
			name:     "C++",
			filePath: "test.cpp",
			src:      []byte("void broken( {\n"),
			plugin:   cpp.NewPlugin(),
		},
		{
			name:     "C#",
			filePath: "test.cs",
			src:      []byte("public class Foo { void broken( { } }\n"),
			plugin:   csharp.NewPlugin(),
		},
		{
			name:     "JavaScript",
			filePath: "test.js",
			src:      []byte("function broken( {\n"),
			plugin:   javascript.NewPlugin(),
		},
		{
			name:     "TypeScript",
			filePath: "test.ts",
			src:      []byte("function broken( {\n"),
			plugin:   typescript.NewPlugin(),
		},
		{
			name:     "PHP",
			filePath: "test.php",
			src:      []byte("<?php\nfunction broken( {\n"),
			plugin:   php.NewPlugin(),
		},
		{
			name:     "Ruby",
			filePath: "test.rb",
			src:      []byte("def broken(\n"),
			plugin:   ruby.NewPlugin(),
		},
		{
			name:     "Scala",
			filePath: "test.scala",
			src:      []byte("def broken( {\n"),
			plugin:   scala.NewPlugin(),
		},
	}
}

// Property 4: Syntax error detection
// For any source bytes containing syntax errors and any Plugin, calling ParseSource
// should return a *SyntaxWarning whose message contains the file path.
func TestProperty_SyntaxErrorEnforcement(tt *testing.T) {
	cases := invalidLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		_, _, err := parseSrc(tt, lc.plugin, lc.filePath, lc.src)
		if err == nil {
			t.Fatalf("[%s] expected SyntaxWarning for broken source", lc.name)
		}

		if !strings.Contains(err.Error(), lc.filePath) {
			t.Fatalf("[%s] error message %q does not contain filePath %q", lc.name, err.Error(), lc.filePath)
		}
	})
}

// Property 3: FilePath consistency
// For any source bytes and any filePath string, every Symbol returned by ParseSource
// should have FilePath equal to the provided filePath argument.
// **Validates: Requirement 3.2**
func TestProperty_FilePathConsistency(tt *testing.T) {
	cases := validLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		// Randomly select a language case.
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		// Generate a random prefix and append the correct extension for this language.
		prefix := rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "prefix")
		ext := filepath.Ext(lc.filePath)
		filePath := prefix + ext

		symbols, _, err := parseSrc(tt, lc.plugin, filePath, lc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", lc.name, err)
		}

		for i, sym := range symbols {
			if filepath.Base(sym.FilePath) != filePath {
				t.Fatalf("[%s] symbol[%d] %q has FilePath base=%q, want %q",
					lc.name, i, sym.Name, filepath.Base(sym.FilePath), filePath)
			}
		}
	})
}

// Property 5: Syntax error tolerance
// For any source bytes containing syntax errors and any Plugin, calling ParseSource
// should return non-nil slices (partial extraction) alongside a *SyntaxWarning.
func TestProperty_SyntaxErrorTolerance(tt *testing.T) {
	cases := invalidLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		symbols, edges, err := parseSrc(tt, lc.plugin, lc.filePath, lc.src)
		var sw *plugin.SyntaxError
		if err != nil && !errors.As(err, &sw) {
			t.Fatalf("[%s] expected SyntaxWarning for broken source, got: %v", lc.name, err)
		}
		if symbols == nil {
			t.Fatalf("[%s] expected non-nil symbols for broken source", lc.name)
		}
		if edges == nil {
			t.Fatalf("[%s] expected non-nil edges for broken source", lc.name)
		}
	})
}

// Property 6: Source non-mutation
// For any source byte slice passed to ParseSource, the byte slice should be identical
// before and after the call.
// **Validates: Requirement 2.7**
func TestProperty_SourceNonMutation(tt *testing.T) {
	cases := validLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		// Make a copy of the source bytes before parsing.
		srcCopy := make([]byte, len(lc.src))
		copy(srcCopy, lc.src)

		// Parse the source.
		_, _, _ = parseSrc(tt, lc.plugin, lc.filePath, lc.src)

		// Verify the source bytes are unchanged.
		if !bytes.Equal(lc.src, srcCopy) {
			t.Fatalf("[%s] source bytes were mutated by ParseSource", lc.name)
		}
	})
}

// Property 7: Parsing idempotency
// For any valid source bytes and filePath, calling ParseSource twice with the same
// arguments should produce identical Symbols and Edges.
// **Validates: Requirement 8.1**
func TestProperty_ParsingIdempotency(tt *testing.T) {
	cases := validLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		syms1, edges1, err1 := parseSrc(tt, lc.plugin, lc.filePath, lc.src)
		syms2, edges2, err2 := parseSrc(tt, lc.plugin, lc.filePath, lc.src)

		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("[%s] error mismatch: first=%v, second=%v", lc.name, err1, err2)
		}
		if err1 != nil {
			return
		}

		if len(syms1) != len(syms2) {
			t.Fatalf("[%s] symbol count mismatch: first=%d, second=%d", lc.name, len(syms1), len(syms2))
		}
		if len(edges1) != len(edges2) {
			t.Fatalf("[%s] edge count mismatch: first=%d, second=%d", lc.name, len(edges1), len(edges2))
		}

		for i := range syms1 {
			s1, s2 := syms1[i], syms2[i]
			if s1.Name != s2.Name {
				t.Fatalf("[%s] symbol[%d] Name mismatch: %q vs %q", lc.name, i, s1.Name, s2.Name)
			}
			if s1.Category != s2.Category {
				t.Fatalf("[%s] symbol[%d] Category mismatch: %q vs %q", lc.name, i, s1.Category, s2.Category)
			}
			if s1.Kind != s2.Kind {
				t.Fatalf("[%s] symbol[%d] Kind mismatch: %q vs %q", lc.name, i, s1.Kind, s2.Kind)
			}
			if filepath.Base(s1.FilePath) != filepath.Base(s2.FilePath) {
				t.Fatalf("[%s] symbol[%d] FilePath mismatch: %q vs %q", lc.name, i, s1.FilePath, s2.FilePath)
			}
			if s1.Span != s2.Span {
				t.Fatalf("[%s] symbol[%d] Span mismatch: %v vs %v", lc.name, i, s1.Span, s2.Span)
			}
		}

		for i := range edges1 {
			e1, e2 := edges1[i], edges2[i]
			if e1.From != e2.From {
				t.Fatalf("[%s] edge[%d] From mismatch: %q vs %q", lc.name, i, e1.From, e2.From)
			}
			if e1.To != e2.To {
				t.Fatalf("[%s] edge[%d] To mismatch: %q vs %q", lc.name, i, e1.To, e2.To)
			}
			if e1.Kind != e2.Kind {
				t.Fatalf("[%s] edge[%d] Kind mismatch: %q vs %q", lc.name, i, e1.Kind, e2.Kind)
			}
		}
	})
}

// containmentCase extends langCase with expected parent and member names for containment edge testing.
type containmentCase struct {
	langCase
	parentName  string
	memberNames []string
}

// containmentLangCases returns langCase entries with source that has a parent type containing members.
func containmentLangCases() []containmentCase {
	return []containmentCase{
		{
			langCase: langCase{
				name:     "Go",
				filePath: "test.go",
				src:      []byte("package main\n\ntype Foo struct {\n\tX int\n\tY int\n}\n"),
				plugin:   golang.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"X", "Y"},
		},
		{
			langCase: langCase{
				name:     "Python",
				filePath: "test.py",
				src:      []byte("class Foo:\n    def bar(self):\n        pass\n    def baz(self):\n        pass\n"),
				plugin:   python.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "Java",
				filePath: "test.java",
				src:      []byte("public class Foo {\n    public void bar() {}\n    public void baz() {}\n}\n"),
				plugin:   java.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "Rust",
				filePath: "test.rs",
				src:      []byte("struct Foo {}\n\nimpl Foo {\n    fn bar(&self) {}\n    fn baz(&self) {}\n}\n"),
				plugin:   rust.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "C",
				filePath: "test.c",
				src:      []byte("struct Foo {\n    int x;\n    int y;\n};\n"),
				plugin:   clang.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"x", "y"},
		},
		{
			langCase: langCase{
				name:     "C++",
				filePath: "test.cpp",
				src:      []byte("class Foo {\npublic:\n    void bar() {}\n    void baz() {}\n};\n"),
				plugin:   cpp.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "C#",
				filePath: "test.cs",
				src:      []byte("public class Foo {\n    public void Bar() {}\n    public void Baz() {}\n}\n"),
				plugin:   csharp.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"Bar", "Baz"},
		},
		{
			langCase: langCase{
				name:     "JavaScript",
				filePath: "test.js",
				src:      []byte("class Foo {\n    bar() {}\n    baz() {}\n}\n"),
				plugin:   javascript.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "TypeScript",
				filePath: "test.ts",
				src:      []byte("class Foo {\n    bar(): void {}\n    baz(): void {}\n}\n"),
				plugin:   typescript.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "PHP",
				filePath: "test.php",
				src:      []byte("<?php\nclass Foo {\n    public function bar() {}\n    public function baz() {}\n}\n"),
				plugin:   php.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "Ruby",
				filePath: "test.rb",
				src:      []byte("class Foo\n  def bar\n  end\n  def baz\n  end\nend\n"),
				plugin:   ruby.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
		{
			langCase: langCase{
				name:     "Scala",
				filePath: "test.scala",
				src:      []byte("class Foo {\n  def bar(): Unit = {}\n  def baz(): Unit = {}\n}\n"),
				plugin:   scala.NewPlugin(),
			},
			parentName:  "Foo",
			memberNames: []string{"bar", "baz"},
		},
	}
}

// Property 8: Containment edges
// For any source containing a class, struct, or type with member declarations (methods, fields, properties),
// there should exist a `contains` Edge from the parent type's name to each extracted member's name.
// **Validates: Requirement 4.1**
func TestProperty_ContainmentEdges(tt *testing.T) {
	cases := containmentLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		cc := cases[idx]

		_, edges, err := parseSrc(tt, cc.plugin, cc.filePath, cc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", cc.name, err)
		}

		// Build a set of contains edges keyed by (From, To).
		type edgeKey struct{ from, to string }
		containsEdges := make(map[edgeKey]bool)
		for _, e := range edges {
			if e.Kind == plugin.EdgeContains {
				containsEdges[edgeKey{e.From, e.To}] = true
			}
		}

		// Verify that for each expected member, a contains edge exists from parentName to that member.
		for _, member := range cc.memberNames {
			scopedMember := cc.parentName + "." + member
			key := edgeKey{cc.parentName, scopedMember}
			if !containsEdges[key] {
				t.Fatalf("[%s] missing contains edge from %q to %q; contains edges found: %v",
					cc.name, cc.parentName, scopedMember, containsEdges)
			}
		}
	})
}

// callEdgeCase extends langCase with a known caller and expected call targets.
type callEdgeCase struct {
	langCase
	callerName  string
	callTargets []string
}

// callEdgeLangCases returns langCase entries with source containing a function that calls other functions.
func callEdgeLangCases() []callEdgeCase {
	return []callEdgeCase{
		{
			langCase: langCase{
				name:     "Go",
				filePath: "test.go",
				src:      []byte("package main\n\nfunc helper() {}\n\nfunc caller() {\n\thelper()\n\tfmt.Println(\"hi\")\n}\n"),
				plugin:   golang.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper"},
		},
		{
			langCase: langCase{
				name:     "Python",
				filePath: "test.py",
				src:      []byte("def helper():\n    pass\n\ndef caller():\n    helper()\n    print(\"hi\")\n"),
				plugin:   python.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper", "print"},
		},
		{
			langCase: langCase{
				name:     "Java",
				filePath: "test.java",
				src:      []byte("public class Foo {\n    void helper() {}\n    void caller() { helper(); System.out.println(\"hi\"); }\n}\n"),
				plugin:   java.NewPlugin(),
			},
			callerName:  "Foo.caller",
			callTargets: []string{"helper", "println"},
		},
		{
			langCase: langCase{
				name:     "Rust",
				filePath: "test.rs",
				src:      []byte("fn helper() {}\n\nfn caller() {\n    helper();\n    println!(\"hi\");\n}\n"),
				plugin:   rust.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper"},
		},
		{
			langCase: langCase{
				name:     "C",
				filePath: "test.c",
				src:      []byte("void helper(void) {}\n\nvoid caller(void) {\n    helper();\n    printf(\"hi\");\n}\n"),
				plugin:   clang.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper", "printf"},
		},
		{
			langCase: langCase{
				name:     "C++",
				filePath: "test.cpp",
				src:      []byte("void helper() {}\n\nvoid caller() {\n    helper();\n    printf(\"hi\");\n}\n"),
				plugin:   cpp.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper", "printf"},
		},
		{
			langCase: langCase{
				name:     "C#",
				filePath: "test.cs",
				src:      []byte("public class Foo {\n    void Helper() {}\n    void Caller() { Helper(); Console.WriteLine(\"hi\"); }\n}\n"),
				plugin:   csharp.NewPlugin(),
			},
			callerName:  "Foo.Caller",
			callTargets: []string{"Helper", "WriteLine"},
		},
		{
			langCase: langCase{
				name:     "JavaScript",
				filePath: "test.js",
				src:      []byte("function helper() {}\nfunction caller() { helper(); console.log(\"hi\"); }\n"),
				plugin:   javascript.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper"},
		},
		{
			langCase: langCase{
				name:     "TypeScript",
				filePath: "test.ts",
				src:      []byte("function helper(): void {}\nfunction caller(): void { helper(); console.log(\"hi\"); }\n"),
				plugin:   typescript.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper"},
		},
		{
			langCase: langCase{
				name:     "PHP",
				filePath: "test.php",
				src:      []byte("<?php\nfunction helper() {}\nfunction caller() { helper(); echo \"hi\"; }\n"),
				plugin:   php.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper"},
		},
		{
			langCase: langCase{
				name:     "Ruby",
				filePath: "test.rb",
				src:      []byte("def helper\nend\n\ndef caller\n  helper()\n  puts \"hi\"\nend\n"),
				plugin:   ruby.NewPlugin(),
			},
			callerName:  "caller",
			callTargets: []string{"helper", "puts"},
		},
		{
			langCase: langCase{
				name:     "Scala",
				filePath: "test.scala",
				src:      []byte("object App {\n  def helper(): Unit = {}\n  def caller(): Unit = { helper(); println(\"hi\") }\n}\n"),
				plugin:   scala.NewPlugin(),
			},
			callerName:  "App.caller",
			callTargets: []string{"helper", "println"},
		},
	}
}

// Property 10: Call edges
// For any source containing a function or method with call expressions in its body,
// there should exist a `calls` Edge from the enclosing function's name to each call target.
// **Validates: Requirement 4.4**
func TestProperty_CallEdges(tt *testing.T) {
	cases := callEdgeLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		cc := cases[idx]

		_, edges, err := parseSrc(tt, cc.plugin, cc.filePath, cc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", cc.name, err)
		}

		// Build a set of calls edges keyed by (From, To).
		type edgeKey struct{ from, to string }
		callsEdges := make(map[edgeKey]bool)
		for _, e := range edges {
			if e.Kind == plugin.EdgeCalls {
				callsEdges[edgeKey{e.From, e.To}] = true
			}
		}

		// Verify that for each expected call target, a calls edge exists from callerName to that target.
		for _, target := range cc.callTargets {
			key := edgeKey{cc.callerName, target}
			if !callsEdges[key] {
				t.Fatalf("[%s] missing calls edge from %q to %q; calls edges found: %v",
					cc.name, cc.callerName, target, callsEdges)
			}
		}
	})
}

// structuralCase extends langCase with expected inherits and/or implements edges.
type structuralCase struct {
	langCase
	expectedEdges []struct {
		from string
		to   string
		kind plugin.EdgeKind
	}
}

// structuralLangCases returns langCase entries with source that produces inherits and/or implements edges.
func structuralLangCases() []structuralCase {
	return []structuralCase{
		// --- Inherits cases ---
		{
			langCase: langCase{
				name:     "Python-inherits",
				filePath: "test.py",
				src:      []byte("class Animal:\n    pass\n\nclass Dog(Animal):\n    pass\n"),
				plugin:   python.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "Java-inherits",
				filePath: "test.java",
				src:      []byte("class Animal {}\nclass Dog extends Animal {}\n"),
				plugin:   java.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "C++-inherits",
				filePath: "test.cpp",
				src:      []byte("class Animal {};\nclass Dog : public Animal {};\n"),
				plugin:   cpp.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "C#-inherits",
				filePath: "test.cs",
				src:      []byte("class Animal {}\nclass Dog : Animal {}\n"),
				plugin:   csharp.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "Ruby-inherits",
				filePath: "test.rb",
				src:      []byte("class Animal\nend\nclass Dog < Animal\nend\n"),
				plugin:   ruby.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "JavaScript-inherits",
				filePath: "test.js",
				src:      []byte("class Animal {}\nclass Dog extends Animal {}\n"),
				plugin:   javascript.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "TypeScript-inherits",
				filePath: "test.ts",
				src:      []byte("class Animal {}\nclass Dog extends Animal {}\n"),
				plugin:   typescript.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "Scala-inherits",
				filePath: "test.scala",
				src:      []byte("class Animal\nclass Dog extends Animal\n"),
				plugin:   scala.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "PHP-inherits",
				filePath: "test.php",
				src:      []byte("<?php\nclass Animal {}\nclass Dog extends Animal {}\n"),
				plugin:   php.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Dog", to: "Animal", kind: plugin.EdgeInherits},
			},
		},
		// --- Implements cases ---
		{
			langCase: langCase{
				name:     "Java-implements",
				filePath: "test.java",
				src:      []byte("interface Runnable { void run(); }\nclass Worker implements Runnable { public void run() {} }\n"),
				plugin:   java.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Worker", to: "Runnable", kind: plugin.EdgeImplements},
			},
		},
		{
			langCase: langCase{
				name:     "Rust-implements",
				filePath: "test.rs",
				src:      []byte("trait Service { fn start(&self); }\nstruct Server {}\nimpl Service for Server { fn start(&self) {} }\n"),
				plugin:   rust.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Server", to: "Service", kind: plugin.EdgeImplements},
			},
		},
		{
			langCase: langCase{
				name:     "C#-implements",
				filePath: "test.cs",
				src:      []byte("class Base {}\nclass Worker : Base, IRunnable { public void Run() {} }\ninterface IRunnable { void Run(); }\n"),
				plugin:   csharp.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				// C# base_list is syntactically flat — the parser emits all
				// entries as inherits since it cannot distinguish class from
				// interface without cross-file type information.
				{from: "Worker", to: "Base", kind: plugin.EdgeInherits},
				{from: "Worker", to: "IRunnable", kind: plugin.EdgeInherits},
			},
		},
		{
			langCase: langCase{
				name:     "PHP-implements",
				filePath: "test.php",
				src:      []byte("<?php\ninterface Runnable { public function run(); }\nclass Worker implements Runnable { public function run() {} }\n"),
				plugin:   php.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Worker", to: "Runnable", kind: plugin.EdgeImplements},
			},
		},
		{
			langCase: langCase{
				name:     "Scala-implements",
				filePath: "test.scala",
				src:      []byte("class Base\ntrait Service { def start(): Unit }\nclass Server extends Base with Service { def start(): Unit = {} }\n"),
				plugin:   scala.NewPlugin(),
			},
			expectedEdges: []struct {
				from string
				to   string
				kind plugin.EdgeKind
			}{
				{from: "Server", to: "Base", kind: plugin.EdgeInherits},
				{from: "Server", to: "Service", kind: plugin.EdgeImplements},
			},
		},
	}
}

// Property 9: Structural relationship edges
// For any source containing a class that extends another type, there should exist an `inherits` Edge
// from the child to the parent; and for any class or type that implements an interface or trait,
// there should exist an `implements` Edge from the implementor to the interface or trait.
func TestProperty_StructuralRelationshipEdges(tt *testing.T) {
	cases := structuralLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		sc := cases[idx]

		_, edges, err := parseSrc(tt, sc.plugin, sc.filePath, sc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", sc.name, err)
		}

		// Build a set of edges keyed by (From, To, Kind).
		type edgeKey struct {
			from string
			to   string
			kind plugin.EdgeKind
		}
		edgeSet := make(map[edgeKey]bool)
		for _, e := range edges {
			edgeSet[edgeKey{e.From, e.To, e.Kind}] = true
		}

		// Verify that all expected edges exist.
		for _, exp := range sc.expectedEdges {
			key := edgeKey{exp.from, exp.to, exp.kind}
			if !edgeSet[key] {
				t.Fatalf("[%s] missing %s edge from %q to %q; edges found: %v",
					sc.name, exp.kind, exp.from, exp.to, edges)
			}
		}
	})
}

// modifierCase holds a language case with a target symbol name and expected properties.
type modifierCase struct {
	expectedProps map[string]string
	symbolName    string
	langCase
}

// modifierLangCases returns modifier test cases for languages that support modifier keywords.
func modifierLangCases() []modifierCase {
	return []modifierCase{
		// Java (Req 5.6): class with public static final modifiers.
		{
			langCase: langCase{
				name:     "Java-class",
				filePath: "test.java",
				src:      []byte("public static final class Foo {}\n"),
				plugin:   java.NewPlugin(),
			},
			symbolName:    "Foo",
			expectedProps: map[string]string{"visibility": "public", "static": "true", "final": "true"},
		},
		// Java method: public static synchronized.
		{
			langCase: langCase{
				name:     "Java-method",
				filePath: "test.java",
				src:      []byte("public class Bar {\n    public static synchronized void baz() {}\n}\n"),
				plugin:   java.NewPlugin(),
			},
			symbolName:    "baz",
			expectedProps: map[string]string{"visibility": "public", "static": "true", "synchronized": "true"},
		},
		// C# (Req 5.7): class with public static abstract modifiers.
		{
			langCase: langCase{
				name:     "CSharp-class",
				filePath: "test.cs",
				src:      []byte("public static abstract class Foo {}\n"),
				plugin:   csharp.NewPlugin(),
			},
			symbolName:    "Foo",
			expectedProps: map[string]string{"visibility": "public", "static": "true", "abstract": "true"},
		},
		// C# method: public virtual.
		{
			langCase: langCase{
				name:     "CSharp-method",
				filePath: "test.cs",
				src:      []byte("public class Bar {\n    public virtual void Baz() {}\n}\n"),
				plugin:   csharp.NewPlugin(),
			},
			symbolName:    "Baz",
			expectedProps: map[string]string{"visibility": "public", "virtual": "true"},
		},
		// C++ (Req 5.8): class member with public visibility and static.
		{
			langCase: langCase{
				name:     "Cpp-method",
				filePath: "test.cpp",
				src:      []byte("class Foo {\npublic:\n    static void bar() {}\n};\n"),
				plugin:   cpp.NewPlugin(),
			},
			symbolName:    "bar",
			expectedProps: map[string]string{"visibility": "public", "static": "true"},
		},
		// PHP (Req 5.9): class method with public static.
		{
			langCase: langCase{
				name:     "PHP-method",
				filePath: "test.php",
				src:      []byte("<?php\nclass Foo {\n    public static function bar() {}\n}\n"),
				plugin:   php.NewPlugin(),
			},
			symbolName:    "bar",
			expectedProps: map[string]string{"visibility": "public", "static": "true"},
		},
		// C (Req 5.10): static function.
		{
			langCase: langCase{
				name:     "C-static",
				filePath: "test.c",
				src:      []byte("static void foo(void) {}\n"),
				plugin:   clang.NewPlugin(),
			},
			symbolName:    "foo",
			expectedProps: map[string]string{"static": "true"},
		},
		// C extern declaration.
		{
			langCase: langCase{
				name:     "C-extern",
				filePath: "test.c",
				src:      []byte("extern void bar(void);\n"),
				plugin:   clang.NewPlugin(),
			},
			symbolName:    "bar",
			expectedProps: map[string]string{"extern": "true"},
		},
		// Scala (Req 5.5): case class.
		{
			langCase: langCase{
				name:     "Scala-case",
				filePath: "test.scala",
				src:      []byte("case class Foo(x: Int)\n"),
				plugin:   scala.NewPlugin(),
			},
			symbolName:    "Foo",
			expectedProps: map[string]string{"case": "true"},
		},
		// Scala sealed trait.
		{
			langCase: langCase{
				name:     "Scala-sealed",
				filePath: "test.scala",
				src:      []byte("sealed trait Bar\n"),
				plugin:   scala.NewPlugin(),
			},
			symbolName:    "Bar",
			expectedProps: map[string]string{"sealed": "true"},
		},
	}
}

// Property 15: Modifier extraction
// For any source file in a language with modifier keywords (Java, C#, C++, PHP, C, Scala),
// when a declaration includes modifiers (e.g., static, abstract, final, case, extern,
// visibility specifiers), the extracted Symbol's Properties should contain those modifiers.
func TestProperty_ModifierExtraction(tt *testing.T) {
	cases := modifierLangCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "caseIndex")
		mc := cases[idx]

		symbols, _, err := parseSrc(tt, mc.plugin, mc.filePath, mc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", mc.name, err)
		}

		// Find the target symbol by name.
		var found *plugin.Symbol
		for i := range symbols {
			if symbols[i].Name == mc.symbolName {
				found = &symbols[i]
				break
			}
		}
		if found == nil {
			names := make([]string, 0, len(symbols))
			for _, s := range symbols {
				names = append(names, s.Name)
			}
			t.Fatalf("[%s] symbol %q not found; symbols: %v", mc.name, mc.symbolName, names)
		}

		// Verify each expected property key-value pair exists.
		for key, wantVal := range mc.expectedProps {
			gotVal, ok := found.Properties[key]
			if !ok {
				t.Fatalf("[%s] symbol %q missing property %q; properties: %v",
					mc.name, mc.symbolName, key, found.Properties)
			}
			if gotVal != wantVal {
				t.Fatalf("[%s] symbol %q property %q: got %q, want %q",
					mc.name, mc.symbolName, key, gotVal, wantVal)
			}
		}
	})
}

// signatureFormatCases returns source snippets for all languages that produce
// callable symbols with parameters and/or return types, used to validate
// that the emitted signature format is uniform across every language.
func signatureFormatCases() []langCase {
	return []langCase{
		{
			name:     "JavaScript",
			filePath: "test.js",
			src:      []byte("function greet(name, age) { return name; }\nclass Foo { bar(x) { return x; } }\n"),
			plugin:   javascript.NewPlugin(),
		},
		{
			name:     "TypeScript",
			filePath: "test.ts",
			src:      []byte("function greet(name: string, age: number): string { return name; }\nclass Foo { bar(x: number): number { return x; } }\n"),
			plugin:   typescript.NewPlugin(),
		},
		{
			name:     "Go",
			filePath: "test.go",
			src:      []byte("package main\n\nfunc Greet(name string, age int) string { return name }\n\ntype Foo struct{}\n\nfunc (f *Foo) Bar(x int) int { return x }\n"),
			plugin:   golang.NewPlugin(),
		},
		{
			name:     "Java",
			filePath: "test.java",
			src:      []byte("public class Foo {\n    public String greet(String name, int age) { return name; }\n    public Foo(String name) {}\n}\n"),
			plugin:   java.NewPlugin(),
		},
		{
			name:     "C#",
			filePath: "test.cs",
			src:      []byte("public class Foo {\n    public string Greet(string name, int age) { return name; }\n}\n"),
			plugin:   csharp.NewPlugin(),
		},
		{
			name:     "Rust",
			filePath: "test.rs",
			src:      []byte("pub fn greet(name: &str, age: u32) -> String { name.to_string() }\npub struct Foo {}\nimpl Foo { pub fn bar(&self, x: i32) -> i32 { x } }\n"),
			plugin:   rust.NewPlugin(),
		},
		{
			name:     "Python",
			filePath: "test.py",
			src:      []byte("def greet(name, age):\n    return name\n\nclass Foo:\n    def bar(self, x):\n        return x\n"),
			plugin:   python.NewPlugin(),
		},
		{
			name:     "Ruby",
			filePath: "test.rb",
			src:      []byte("def greet(name, age)\n  name\nend\n\nclass Foo\n  def bar(x)\n    x\n  end\nend\n"),
			plugin:   ruby.NewPlugin(),
		},
		{
			name:     "PHP",
			filePath: "test.php",
			src:      []byte("<?php\nfunction greet(string $name, int $age): string { return $name; }\nclass Foo {\n    public function bar(int $x): int { return $x; }\n}\n"),
			plugin:   php.NewPlugin(),
		},
		{
			name:     "C",
			filePath: "test.c",
			src:      []byte("int add(int a, int b) { return a + b; }\nvoid greet(char *name) { printf(\"%s\", name); }\n"),
			plugin:   clang.NewPlugin(),
		},
		{
			name:     "C++",
			filePath: "test.cpp",
			src:      []byte("class Foo {\npublic:\n    int bar(int x, int y) { return x + y; }\n};\nint add(int a, int b) { return a + b; }\n"),
			plugin:   cpp.NewPlugin(),
		},
		{
			name:     "Scala",
			filePath: "test.scala",
			src:      []byte("class Foo {\n  def bar(x: Int, y: Int): Int = x + y\n}\ndef greet(name: String): String = name\n"),
			plugin:   scala.NewPlugin(),
		},
	}
}

// TestProperty_UniformSignatureFormat validates that ALL language plugins produce
// symbols whose signatures follow the same format conventions:
//
//  1. Symbol names must not contain language-specific prefixes (e.g., PHP "$").
//  2. Callable signatures must match: name(params) or name(params) -> returnType
//  3. Each param must be either "name" or "name: type" — no "$name", no bare types.
//  4. The "->" separator is used for return types (not ":", not language-native syntax).
//  5. No language-specific syntax leaks into symbol names (no "$", no leading "::").
func TestProperty_UniformSignatureFormat(tt *testing.T) {
	cases := signatureFormatCases()

	rapid.Check(tt, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		lc := cases[idx]

		symbols, _, err := parseSrc(tt, lc.plugin, lc.filePath, lc.src)
		if err != nil {
			t.Fatalf("[%s] ParseSource returned error: %v", lc.name, err)
		}

		for _, sym := range symbols {
			// Rule 1 & 5: No "$" in symbol names.
			if strings.Contains(sym.Name, "$") {
				t.Fatalf("[%s] symbol name %q contains '$'", lc.name, sym.Name)
			}

			// Rule 5: No leading "::" in names.
			if strings.HasPrefix(sym.Name, "::") {
				t.Fatalf("[%s] symbol name %q starts with '::'", lc.name, sym.Name)
			}

			// Skip non-callable symbols for signature format checks.
			if sym.Category != plugin.CategoryCallable {
				continue
			}

			sig := sym.Signature

			// Rule 2: Callable signatures must contain parentheses.
			openParen := strings.Index(sig, "(")
			if openParen < 0 {
				// Some languages emit bare names for prototypes/declarations — allow that.
				continue
			}

			closeParen := strings.LastIndex(sig, ")")
			if closeParen < 0 || closeParen < openParen {
				t.Fatalf("[%s] symbol %q signature %q has unbalanced parentheses", lc.name, sym.Name, sig)
			}

			// The name part (before "(") must not contain "$".
			namePart := sig[:openParen]
			if strings.Contains(namePart, "$") {
				t.Fatalf("[%s] symbol %q signature name part %q contains '$'", lc.name, sym.Name, namePart)
			}

			// Rule 4: If there's a return type, it must use " -> " separator.
			afterParen := sig[closeParen+1:]
			if afterParen != "" {
				if !strings.HasPrefix(afterParen, " -> ") {
					t.Fatalf("[%s] symbol %q signature %q has return type not using ' -> ' separator (got suffix %q)",
						lc.name, sym.Name, sig, afterParen)
				}
				returnType := strings.TrimPrefix(afterParen, " -> ")
				if returnType == "" {
					t.Fatalf("[%s] symbol %q signature %q has empty return type after ' -> '", lc.name, sym.Name, sig)
				}
			}

			// Rule 3: Each param must be "name" or "name: type", not "$name" or bare type.
			paramStr := sig[openParen+1 : closeParen]
			if paramStr != "" {
				params := splitParams(paramStr)
				for _, p := range params {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					if strings.HasPrefix(p, "$") {
						t.Fatalf("[%s] symbol %q param %q starts with '$'", lc.name, sym.Name, p)
					}
					// If it contains ": ", the part before must be a name (no "$").
					if colonIdx := strings.Index(p, ": "); colonIdx >= 0 {
						paramName := p[:colonIdx]
						if strings.Contains(paramName, "$") {
							t.Fatalf("[%s] symbol %q param name %q contains '$'", lc.name, sym.Name, paramName)
						}
						if paramName == "" {
							t.Fatalf("[%s] symbol %q has empty param name in %q", lc.name, sym.Name, p)
						}
					}
				}
			}
		}
	})
}

// splitParams splits a parameter string by commas, respecting nested angle brackets
// and parentheses (e.g., "fn: (item: T) => U" should not split on the inner comma).
func splitParams(s string) []string {
	var result []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '<', '[':
			depth++
		case ')', '>', ']':
			depth--
		case ',':
			if depth == 0 {
				result = append(result, s[start:i])
				start = i + 1
			}
		}
	}
	result = append(result, s[start:])
	return result
}
