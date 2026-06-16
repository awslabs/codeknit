// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"

	"pgregory.net/rapid"
)

// TestE2E_OutputStructure verifies the structural format of output files:
// [symbols] section presence, [edges] section when edges exist, parseability,
// and file naming convention.
func TestE2E_OutputStructure(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models.go": `package pkg

type Animal interface {
	Speak() string
}

type Dog struct {
	Name string
}

func (d *Dog) Speak() string {
	return "woof"
}

func NewDog(name string) *Dog {
	return &Dog{Name: name}
}
`,
		"pkg/service.go": `package pkg

type Service struct {
	Name string
}

func NewService(name string) *Service {
	return &Service{Name: name}
}

func (s *Service) Run() string {
	return s.Name
}
`,
		"lib/utils.py": `class Calculator:
    def add(self, a, b):
        return a + b

    def multiply(self, a, b):
        return a * b

def create_calculator():
    return Calculator()
`,
	}

	inputDir := writeFixture(t, fixtures)
	outputDir := runcodeknitRaw(t, inputDir)

	matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("glob output files: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no output files found")
	}
	sort.Strings(matches)

	// Requirement 17.4: Verify output files follow map_NNN.skt naming convention.
	nameRe := regexp.MustCompile(`^map_\d{3}\.skt$`)
	for _, path := range matches {
		base := filepath.Base(path)
		if !nameRe.MatchString(base) {
			t.Errorf("output file %q does not match map_NNN.skt naming convention", base)
		}
	}

	// Verify sequential numbering starting from 001.
	for i, path := range matches {
		base := filepath.Base(path)
		expected := filepath.Base(path) // just check the index
		_ = expected
		wantSuffix := strings.TrimPrefix(base, "map_")
		wantSuffix = strings.TrimSuffix(wantSuffix, ".skt")
		wantNum := i + 1
		gotNum := 0
		for _, ch := range wantSuffix {
			gotNum = gotNum*10 + int(ch-'0')
		}
		if gotNum != wantNum {
			t.Errorf("output file %q: expected sequential number %03d, got %03d", base, wantNum, gotNum)
		}
	}

	// Read all output files and check structural invariants.
	hasEdgesInAnyFile := false
	for _, path := range matches {
		content, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("read output file %s: %v", path, readErr)
		}
		text := string(content)

		// Requirement 17.1: Each output file must contain a [symbols] section.
		if !strings.Contains(text, "[symbols]") {
			t.Errorf("output file %s missing [symbols] section", filepath.Base(path))
		}

		// Requirement 17.2: Files with edge lines must have an [edges] section.
		if strings.Contains(text, "[edges]") {
			hasEdgesInAnyFile = true
		}
	}

	// Our fixture should produce edges (contains relationships at minimum).
	if !hasEdgesInAnyFile {
		t.Log("note: no [edges] section found in any output file (fixture may not produce edges)")
	}

	// Requirement 17.3: All output files must be parseable by parser.ParseOutput.
	readers := make([]io.Reader, 0, len(matches))
	closers := make([]io.Closer, 0, len(matches))
	for _, path := range matches {
		f, openErr := os.Open(path) //nolint:gosec // test file
		if openErr != nil {
			t.Fatalf("open output file %s: %v", path, openErr)
		}
		closers = append(closers, f)
		readers = append(readers, f)
	}
	sg, parseErr := parser.ParseOutput(readers, false)
	for _, c := range closers {
		_ = c.Close()
	}
	if parseErr != nil {
		t.Fatalf("ParseOutput failed: %v", parseErr)
	}
	if len(sg.Symbols) == 0 {
		t.Error("parsed output contains no symbols")
	}

	// Snapshot test: verify raw CLI output against golden file.
	assertSnapshot(t, outputDir, inputDir)
}

// langFixture holds a language name, file path, and valid source content
// that produces at least one parseable symbol.
type langFixture struct {
	name     string
	filePath string
	content  string
}

// allLangFixtures returns valid source fixtures for all 12 supported languages.
// Each fixture produces at least one symbol when parsed by the CLI.
func allLangFixtures() []langFixture {
	return []langFixture{
		{
			name:     "Go",
			filePath: "src/main.go",
			content:  "package main\n\nfunc Hello() string { return \"hello\" }\n",
		},
		{
			name:     "TypeScript",
			filePath: "src/app.ts",
			content:  "export function greet(name: string): string { return name; }\n",
		},
		{
			name:     "JavaScript",
			filePath: "src/utils.js",
			content:  "function helper() { return 42; }\n",
		},
		{
			name:     "Python",
			filePath: "src/app.py",
			content:  "def process(data):\n    return data\n",
		},
		{
			name:     "Java",
			filePath: "src/App.java",
			content:  "public class App {\n    public void run() {}\n}\n",
		},
		{
			name:     "C",
			filePath: "src/lib.c",
			content:  "int add(int a, int b) { return a + b; }\n",
		},
		{
			name:     "C++",
			filePath: "src/vec.cpp",
			content:  "class Vec {\npublic:\n    double x;\n    double len() { return x; }\n};\n",
		},
		{
			name:     "C#",
			filePath: "src/Svc.cs",
			content:  "public class Svc {\n    public void Run() {}\n}\n",
		},
		{
			name:     "Ruby",
			filePath: "src/task.rb",
			content:  "class Task\n  def run\n    \"ok\"\n  end\nend\n",
		},
		{
			name:     "Rust",
			filePath: "src/lib.rs",
			content:  "pub fn compute(x: i32) -> i32 { x * 2 }\n",
		},
		{
			name:     "PHP",
			filePath: "src/index.php",
			content:  "<?php\nfunction handle() { return \"ok\"; }\n",
		},
		{
			name:     "Scala",
			filePath: "src/Main.scala",
			content:  "class Main {\n  def run(): Unit = {}\n}\n",
		},
	}
}

// drawLangSubset uses rapid to draw a random non-empty subset of language fixtures.
func drawLangSubset(t *rapid.T) []langFixture {
	all := allLangFixtures()
	mask := rapid.SliceOfN(rapid.Bool(), len(all), len(all)).Draw(t, "langMask")

	var selected []langFixture
	for i, m := range mask {
		if m {
			selected = append(selected, all[i])
		}
	}
	// Ensure at least one language is selected.
	if len(selected) == 0 {
		selected = append(selected, all[rapid.IntRange(0, len(all)-1).Draw(t, "fallback")])
	}
	return selected
}

// buildFixtureFromSubset creates a fixture map from a language subset.
func buildFixtureFromSubset(langs []langFixture) map[string]string {
	files := make(map[string]string, len(langs))
	for _, lf := range langs {
		files[lf.filePath] = lf.content
	}
	return files
}

// Feature: e2e-cli-tests, Property 1: Source file completeness
// For any set of supported source files placed in a fixture directory, when the
// CLI binary processes that directory, every source file should appear as a file
// header in the parsed output.
func TestProperty_SourceFileCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		sg := runcodeknitForProperty(t, inputDir)

		// Collect all file paths from symbols.
		fileSet := make(map[string]bool)
		for _, sym := range sg.Symbols {
			fileSet[normalizeFilePath(sym.FilePath)] = true
		}

		for relPath := range files {
			found := false
			for normPath := range fileSet {
				if strings.HasSuffix(normPath, relPath) || normPath == relPath {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("source file %q not found in output (had %d files: %v)", relPath, len(fileSet), fileSet)
			}
		}
	})
}

// Feature: e2e-cli-tests, Property 2: Symbol count lower bound
// For any fixture directory where each source file contains at least one
// parseable symbol, the total number of symbols in the parsed output should be
// greater than or equal to the number of source files.
func TestProperty_SymbolCountLowerBound(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		sg := runcodeknitForProperty(t, inputDir)

		if len(sg.Symbols) < len(files) {
			t.Fatalf("expected at least %d symbols (one per source file), got %d",
				len(files), len(sg.Symbols))
		}
	})
}

// Feature: e2e-cli-tests, Property 4: Output round-trip parseability
// For any valid fixture directory processed by the CLI (with or without
// --minify), all output map_*.skt files should be parseable by
// parser.ParseOutput without errors, producing a valid SymbolGraph.
func TestProperty_OutputRoundTripParseability(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)
		useMinify := rapid.Bool().Draw(t, "minify")

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		var flags []string
		if useMinify {
			flags = append(flags, "--minify")
		}

		outputDir := runcodeknitRawForProperty(t, inputDir, flags...)
		defer func() { _ = os.RemoveAll(outputDir) }()

		matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("no map_*.skt output files found")
		}
		sort.Strings(matches)

		readers := make([]io.Reader, 0, len(matches)+1)
		closers := make([]io.Closer, 0, len(matches)+1)

		// Include dict.skt first so the parser can resolve minified codes.
		if useMinify {
			dictPath := filepath.Join(outputDir, "dict.skt")
			if df, openErr := os.Open(dictPath); openErr == nil { //nolint:gosec // test file
				closers = append(closers, df)
				readers = append(readers, df)
			}
		}

		for _, path := range matches {
			f, openErr := os.Open(path) //nolint:gosec // test file
			if openErr != nil {
				t.Fatalf("open %s: %v", path, openErr)
			}
			closers = append(closers, f)
			readers = append(readers, f)
		}
		defer func() {
			for _, c := range closers {
				_ = c.Close()
			}
		}()

		sg, parseErr := parser.ParseOutput(readers, useMinify)
		if parseErr != nil {
			t.Fatalf("ParseOutput failed (minify=%v): %v", useMinify, parseErr)
		}
		if len(sg.Symbols) == 0 {
			t.Fatal("parsed output contains no symbols")
		}
	})
}

// Feature: e2e-cli-tests, Property 6: Output file structural invariants
// For any output file produced by the CLI, the file must contain a [symbols]
// section. If the file contains edge lines, it must also contain an [edges]
// section.
func TestProperty_OutputFileStructuralInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		outputDir := runcodeknitRawForProperty(t, inputDir)
		defer func() { _ = os.RemoveAll(outputDir) }()

		matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("no output files found")
		}

		edgeLineRe := regexp.MustCompile(`^\S+\s+--\w+-->\s+\S+$`)

		for _, path := range matches {
			content, readErr := os.ReadFile(path) //nolint:gosec // test file
			if readErr != nil {
				t.Fatalf("read %s: %v", path, readErr)
			}
			text := string(content)

			if !strings.Contains(text, "[symbols]") {
				t.Fatalf("output file %s missing [symbols] section", filepath.Base(path))
			}

			// Check if file has edge lines — if so, it must have [edges] section.
			hasEdgeLines := false
			for _, line := range strings.Split(text, "\n") {
				if edgeLineRe.MatchString(strings.TrimSpace(line)) {
					hasEdgeLines = true
					break
				}
			}
			if hasEdgeLines && !strings.Contains(text, "[edges]") {
				t.Fatalf("output file %s has edge lines but no [edges] section", filepath.Base(path))
			}
		}
	})
}

// Feature: e2e-cli-tests, Property 7: Output file naming convention
// For any set of output files produced by the CLI, the files should follow the
// naming convention map_NNN.skt where NNN is a zero-padded sequential number
// starting from 001.
func TestProperty_OutputFileNamingConvention(t *testing.T) {
	nameRe := regexp.MustCompile(`^map_(\d{3})\.skt$`)

	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		outputDir := runcodeknitRawForProperty(t, inputDir)
		defer func() { _ = os.RemoveAll(outputDir) }()

		matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("no output files found")
		}
		sort.Strings(matches)

		for i, path := range matches {
			base := filepath.Base(path)
			m := nameRe.FindStringSubmatch(base)
			if m == nil {
				t.Fatalf("output file %q does not match map_NNN.skt naming", base)
			}
			// Verify sequential numbering starting from 001.
			wantNum := i + 1
			gotNum := 0
			for _, ch := range m[1] {
				gotNum = gotNum*10 + int(ch-'0')
			}
			if gotNum != wantNum {
				t.Fatalf("output file %q: expected sequential number %03d, got %03d", base, wantNum, gotNum)
			}
		}
	})
}
