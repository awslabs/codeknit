// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"errors"
	"os"
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

// partialErrorCase holds a language name, file path, a valid declaration,
// a broken declaration, and the plugin to parse with.
type partialErrorCase struct {
	plugin     plugin.LanguagePlugin
	name       string
	filePath   string
	brokenDecl string
	validDecls []string
}

// parseSrc writes src to a temp file and parses it through the plugin.
func parseSrc(t *testing.T, p plugin.LanguagePlugin, filename string, src []byte) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return p.Parse(path)
}

// partialErrorCases returns test cases for all 12 language plugins.
// Each case has one valid declaration + one broken declaration.
func partialErrorCases() []partialErrorCase {
	return []partialErrorCase{
		{
			name:     "Go",
			filePath: "test.go",
			plugin:   golang.NewPlugin(),
			validDecls: []string{
				"package main\n\nfunc main() {}\n",
				"package main\n\nfunc hello() {}\n",
				"package main\n\nvar x int = 42\n",
			},
			brokenDecl: "\ntype X struct {\n",
		},
		{
			name:     "Python",
			filePath: "test.py",
			plugin:   python.NewPlugin(),
			validDecls: []string{
				"class Foo:\n    pass\n",
				"def hello():\n    pass\n",
				"x = 42\n",
			},
			brokenDecl: "\ndef bar(\n",
		},
		{
			name:     "TypeScript",
			filePath: "test.ts",
			plugin:   typescript.NewPlugin(),
			validDecls: []string{
				"function greet(): void {}\n",
				"const x: number = 42;\n",
				"class Foo {}\n",
			},
			brokenDecl: "\nconst y: = ;\n",
		},
		{
			name:     "JavaScript",
			filePath: "test.js",
			plugin:   javascript.NewPlugin(),
			validDecls: []string{
				"function greet() {}\n",
				"const x = 42;\n",
				"class Foo {}\n",
			},
			brokenDecl: "\nfunction broken( {\n",
		},
		{
			name:     "Java",
			filePath: "test.java",
			plugin:   java.NewPlugin(),
			validDecls: []string{
				"public class Foo { public void bar() {} }\n",
				"interface Baz { void run(); }\n",
			},
			brokenDecl: "\npublic class Broken { void bad( { } }\n",
		},
		{
			name:     "Rust",
			filePath: "test.rs",
			plugin:   rust.NewPlugin(),
			validDecls: []string{
				"fn main() {}\n",
				"pub struct Foo {}\n",
				"pub fn hello() {}\n",
			},
			brokenDecl: "\nfn broken( {\n",
		},
		{
			name:     "Ruby",
			filePath: "test.rb",
			plugin:   ruby.NewPlugin(),
			validDecls: []string{
				"class Foo\n  def bar\n    puts \"hi\"\n  end\nend\n",
				"def hello\n  puts \"hi\"\nend\n",
			},
			brokenDecl: "\ndef broken(\n",
		},
		{
			name:     "PHP",
			filePath: "test.php",
			plugin:   php.NewPlugin(),
			validDecls: []string{
				"<?php\nclass Foo { public function bar() {} }\n",
				"<?php\nfunction hello() {}\n",
			},
			brokenDecl: "\nfunction broken( {\n",
		},
		{
			name:     "C",
			filePath: "test.c",
			plugin:   clang.NewPlugin(),
			validDecls: []string{
				"void hello(void) {}\n",
				"struct Foo { int x; };\n",
				"int main(void) { return 0; }\n",
			},
			brokenDecl: "\nvoid broken( {\n",
		},
		{
			name:     "C++",
			filePath: "test.cpp",
			plugin:   cpp.NewPlugin(),
			validDecls: []string{
				"class Foo { public: void bar() {} };\n",
				"void hello() {}\n",
			},
			brokenDecl: "\nvoid broken( {\n",
		},
		{
			name:     "C#",
			filePath: "test.cs",
			plugin:   csharp.NewPlugin(),
			validDecls: []string{
				"public class Foo { public void Bar() {} }\n",
				"public interface IBaz { void Run(); }\n",
			},
			brokenDecl: "\npublic class Broken { void bad( { } }\n",
		},
		{
			name:     "Scala",
			filePath: "test.scala",
			plugin:   scala.NewPlugin(),
			validDecls: []string{
				"class Foo { def bar(): Unit = {} }\n",
				"object Main { def main(): Unit = {} }\n",
			},
			brokenDecl: "\ndef broken( {\n",
		},
	}
}

// allErrorCases returns test cases where every top-level node has errors.
func allErrorCases() []partialErrorCase {
	return []partialErrorCase{
		{name: "Go", filePath: "test.go", plugin: golang.NewPlugin(), validDecls: nil, brokenDecl: "package main\n\nfunc broken( {\n\ntype X struct {\n"},
		{name: "Python", filePath: "test.py", plugin: python.NewPlugin(), validDecls: nil, brokenDecl: "def broken(\n\nclass Bad(\n"},
		{name: "TypeScript", filePath: "test.ts", plugin: typescript.NewPlugin(), validDecls: nil, brokenDecl: "function broken( {\n\nconst x: = ;\n"},
		{name: "JavaScript", filePath: "test.js", plugin: javascript.NewPlugin(), validDecls: nil, brokenDecl: "function broken( {\n\nconst x = = ;\n"},
		{name: "Java", filePath: "test.java", plugin: java.NewPlugin(), validDecls: nil, brokenDecl: "public class Broken { void bad( { } }\n"},
		{name: "Rust", filePath: "test.rs", plugin: rust.NewPlugin(), validDecls: nil, brokenDecl: "fn broken( {\n\nfn also_broken( {\n"},
		{name: "Ruby", filePath: "test.rb", plugin: ruby.NewPlugin(), validDecls: nil, brokenDecl: "def broken(\n\ndef also_broken(\n"},
		{name: "PHP", filePath: "test.php", plugin: php.NewPlugin(), validDecls: nil, brokenDecl: "<?php\nfunction broken( {\n"},
		{name: "C", filePath: "test.c", plugin: clang.NewPlugin(), validDecls: nil, brokenDecl: "void broken( {\n\nvoid also( {\n"},
		{name: "C++", filePath: "test.cpp", plugin: cpp.NewPlugin(), validDecls: nil, brokenDecl: "void broken( {\n\nvoid also( {\n"},
		{name: "C#", filePath: "test.cs", plugin: csharp.NewPlugin(), validDecls: nil, brokenDecl: "public class Broken { void bad( { } }\n"},
		{name: "Scala", filePath: "test.scala", plugin: scala.NewPlugin(), validDecls: nil, brokenDecl: "def broken( {\n\ndef also( {\n"},
	}
}

// TestProperty_BugCondition_PartialErrorReturnsSymbols is a property-based test
// that demonstrates the bug: Parse returns nil instead of valid symbols
// when a file has partial syntax errors.
//
// EXPECTED: This test FAILS on unfixed code, confirming the bug exists.
// After the fix, this test should PASS.
func TestProperty_BugCondition_PartialErrorReturnsSymbols(tt *testing.T) {
	for _, tc := range partialErrorCases() {
		tt.Run(tc.name, func(t *testing.T) {
			rapid.Check(tt, func(t *rapid.T) {
				// Pick a random valid declaration from the available set
				idx := rapid.IntRange(0, len(tc.validDecls)-1).Draw(t, "validDeclIdx")
				validDecl := tc.validDecls[idx]

				// Construct source: valid declaration + broken declaration
				src := []byte(validDecl + tc.brokenDecl)

				symbols, edges, err := parseSrc(tt, tc.plugin, tc.filePath, src)
				// May return a *SyntaxWarning alongside partial results
				var sw *plugin.SyntaxError
				if err != nil && !errors.As(err, &sw) {
					t.Fatalf("Parse returned unexpected error: %v", err)
				}

				// Bug condition: symbols must be non-nil (currently returns nil)
				if symbols == nil {
					t.Fatalf("symbols is nil; expected non-nil slice with extracted symbols from valid nodes")
				}

				// Bug condition: edges must be non-nil (currently returns nil)
				if edges == nil {
					t.Fatalf("edges is nil; expected non-nil slice")
				}

				// There is at least one valid declaration, so we expect symbols
				if len(symbols) == 0 {
					t.Fatalf("symbols is empty; expected at least one symbol from the valid declaration")
				}
			})
		})
	}
}

// TestBugCondition_AllErrorFileReturnsNonNilSlices tests that when every
// top-level node has errors, Parse returns non-nil empty slices.
func TestBugCondition_AllErrorFileReturnsNonNilSlices(t *testing.T) {
	for _, tc := range allErrorCases() {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.brokenDecl)

			symbols, edges, err := parseSrc(t, tc.plugin, tc.filePath, src)
			// May return a *SyntaxWarning — that's fine.
			var sw *plugin.SyntaxError
			if err != nil && !errors.As(err, &sw) {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			if symbols == nil {
				t.Fatalf("symbols is nil; expected non-nil empty slice for all-error file")
			}

			if edges == nil {
				t.Fatalf("edges is nil; expected non-nil empty slice for all-error file")
			}
		})
	}
}

// TestProperty_Preservation_ErrorFreeFilesProduceSymbols is a property-based test
// that verifies error-free files produce non-nil symbols.
// This captures the baseline behavior that MUST be preserved after the fix.
//
// EXPECTED: This test PASSES on unfixed code — error-free files are correctly parsed today.
func TestProperty_Preservation_ErrorFreeFilesProduceSymbols(tt *testing.T) {
	cases := partialErrorCases()

	rapid.Check(tt, func(t *rapid.T) {
		// Pick a random language
		langIdx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		tc := cases[langIdx]

		// Pick a random valid declaration (error-free source)
		declIdx := rapid.IntRange(0, len(tc.validDecls)-1).Draw(t, "validDeclIdx")
		validSrc := []byte(tc.validDecls[declIdx])

		symbols, edges, err := parseSrc(tt, tc.plugin, tc.filePath, validSrc)
		// Error-free files must not return an error
		if err != nil {
			t.Fatalf("[%s] Parse returned error for valid source: %v", tc.name, err)
		}

		// Symbols must be non-nil for error-free files
		if symbols == nil {
			t.Fatalf("[%s] symbols is nil for valid source; expected non-nil slice", tc.name)
		}

		// Error-free source with a declaration should produce at least one symbol
		if len(symbols) == 0 {
			t.Fatalf("[%s] symbols is empty for valid source %q; expected at least one symbol", tc.name, string(validSrc))
		}

		// Edges may be nil when there are no relationships to extract — that's valid baseline behavior.
		_ = edges
	})
}

// TestProperty_Preservation_StrictModeReturnsError is a property-based test
// that verifies parsing broken source returns a *SyntaxWarning with
// "syntax error" in the message.
func TestProperty_Preservation_StrictModeReturnsError(tt *testing.T) {
	cases := partialErrorCases()

	rapid.Check(tt, func(t *rapid.T) {
		langIdx := rapid.IntRange(0, len(cases)-1).Draw(t, "langIndex")
		tc := cases[langIdx]

		declIdx := rapid.IntRange(0, len(tc.validDecls)-1).Draw(t, "validDeclIdx")
		src := []byte(tc.validDecls[declIdx] + tc.brokenDecl)

		_, _, err := parseSrc(tt, tc.plugin, tc.filePath, src)

		if err == nil {
			t.Fatalf("[%s] expected SyntaxWarning for broken source", tc.name)
		}

		// The error message must contain "syntax error"
		errMsg := err.Error()
		if !strings.Contains(errMsg, "syntax error") {
			t.Fatalf("[%s] error message %q does not contain 'syntax error'", tc.name, errMsg)
		}
	})
}
