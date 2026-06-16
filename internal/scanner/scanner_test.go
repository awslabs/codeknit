// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/plugin"

	"pgregory.net/rapid"
)

// allTestPatterns returns the full set of test patterns matching what the
// real plugins provide. Used by tests that need test file filtering.
func allTestPatterns() map[string]plugin.TestConfig {
	return map[string]plugin.TestConfig{
		".ts":    {ContainsDot: []string{".test.", ".spec."}},
		".tsx":   {ContainsDot: []string{".test.", ".spec."}},
		".js":    {ContainsDot: []string{".test.", ".spec."}},
		".jsx":   {ContainsDot: []string{".test.", ".spec."}},
		".go":    {NameSuffixes: []string{"_test"}},
		".py":    {NamePrefixes: []string{"test_"}, NameSuffixes: []string{"_test"}},
		".pyi":   {NamePrefixes: []string{"test_"}, NameSuffixes: []string{"_test"}},
		".rb":    {NamePrefixes: []string{"test_"}, NameSuffixes: []string{"_test", "_spec"}},
		".java":  {NameSuffixes: []string{"Test", "Tests", "Spec", "Suite"}},
		".scala": {NameSuffixes: []string{"Test", "Tests", "Spec", "Suite"}},
		".sc":    {NameSuffixes: []string{"Test", "Tests", "Spec", "Suite"}},
		".cs":    {NameSuffixes: []string{"Test", "Tests", "Spec"}},
		".c":     {NameSuffixes: []string{"_test"}},
		".h":     {NameSuffixes: []string{"_test"}},
		".cpp":   {NameSuffixes: []string{"_test"}},
		".hpp":   {NameSuffixes: []string{"_test"}},
		".cc":    {NameSuffixes: []string{"_test"}},
		".cxx":   {NameSuffixes: []string{"_test"}},
		".hxx":   {NameSuffixes: []string{"_test"}},
		".php":   {NameSuffixes: []string{"Test", "Spec"}},
		".rs":    {NameSuffixes: []string{"_test"}},
	}
}

// setupTree creates a temp directory with the given file tree.
// Keys are slash-separated relative paths; values are file contents.
func setupTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	populateTree(t, dir, files)
	return dir
}

// tHelper is satisfied by both *testing.T and *rapid.T.
type tHelper interface {
	Helper()
	Fatal(args ...any)
}

// populateTree writes files into an existing directory.
func populateTree(t tHelper, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestScan_ExtensionFiltering(t *testing.T) {
	dir := setupTree(t, map[string]string{
		"main.ts":      "",
		"app.tsx":      "",
		"readme.md":    "",
		"style.css":    "",
		"src/index.ts": "",
		"src/utils.js": "",
	})
	s := &Scanner{Extensions: []string{".ts", ".tsx"}, TestPatterns: allTestPatterns(), CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"app.tsx", "main.ts", "src/index.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_TestFileExclusion(t *testing.T) {
	dir := setupTree(t, map[string]string{
		// JS/TS conventions
		"src/app.ts":              "",
		"src/app.test.ts":         "",
		"src/app.spec.ts":         "",
		"src/__tests__/helper.ts": "",
		"src/utils.ts":            "",
		// Go convention
		"pkg/main.go":      "",
		"pkg/main_test.go": "",
		// Python conventions
		"lib/helper.py":      "",
		"lib/test_helper.py": "",
		"lib/helper_test.py": "",
		// Ruby conventions
		"lib/models/user.rb":      "",
		"lib/models/user_spec.rb": "",
		"spec/user_spec.rb":       "",
		"test/test_animal.rb":     "",
		// Java conventions
		"src/main/App.java":          "",
		"src/main/AppTest.java":      "",
		"test/IntegrationTest.java":  "",
		"tests/IntegrationTest.java": "",
		// C# conventions
		"src/Models/User.cs":     "",
		"src/Models/UserTest.cs": "",
		// C/C++ conventions
		"src/config.c":       "",
		"src/config_test.c":  "",
		"src/utils.cpp":      "",
		"src/utils_test.cpp": "",
		// PHP conventions
		"src/App.php":     "",
		"src/AppTest.php": "",
		// Rust conventions
		"src/lib.rs":      "",
		"src/lib_test.rs": "",
		// Scala conventions
		"src/Main.scala":     "",
		"src/MainSpec.scala": "",
	})
	s := &Scanner{
		Extensions:   []string{".ts", ".go", ".py", ".rb", ".java", ".cs", ".c", ".cpp", ".php", ".rs", ".scala"},
		TestPatterns: allTestPatterns(),
		CollectTest:  false,
	}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{
		"lib/helper.py",
		"lib/models/user.rb",
		"pkg/main.go",
		"src/App.php",
		"src/app.ts",
		"src/config.c",
		"src/lib.rs",
		"src/Main.scala",
		"src/Models/User.cs",
		"src/main/App.java",
		"src/utils.cpp",
		"src/utils.ts",
	}
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_TestFileInclusion(t *testing.T) {
	dir := setupTree(t, map[string]string{
		"src/app.ts":              "",
		"src/app.test.ts":         "",
		"src/app.spec.ts":         "",
		"src/__tests__/helper.ts": "",
		"pkg/main.go":             "",
		"pkg/main_test.go":        "",
		"lib/helper.py":           "",
		"lib/test_helper.py":      "",
		"lib/helper_test.py":      "",
		"test/test_animal.rb":     "",
		"spec/user_spec.rb":       "",
		"tests/AppTest.java":      "",
	})
	s := &Scanner{
		Extensions:   []string{".ts", ".go", ".py", ".rb", ".java"},
		TestPatterns: allTestPatterns(),
		CollectTest:  true,
	}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{
		"lib/helper.py", "lib/helper_test.py", "lib/test_helper.py",
		"pkg/main.go", "pkg/main_test.go",
		"spec/user_spec.rb",
		"src/__tests__/helper.ts", "src/app.spec.ts", "src/app.test.ts", "src/app.ts",
		"test/test_animal.rb",
		"tests/AppTest.java",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_GitignoreRoot(t *testing.T) {
	dir := setupTree(t, map[string]string{
		".gitignore":   "*.log\nbuild/\n",
		"main.ts":      "",
		"debug.log":    "",
		"build/out.ts": "",
		"src/index.ts": "",
	})
	s := &Scanner{Extensions: []string{".ts", ".log"}, CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"main.ts", "src/index.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_GitignoreSubdirectory(t *testing.T) {
	dir := setupTree(t, map[string]string{
		"src/.gitignore":    "vendor/\n",
		"src/app.ts":        "",
		"src/vendor/lib.ts": "",
		"lib/vendor/ok.ts":  "",
	})
	s := &Scanner{Extensions: []string{".ts"}, CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	// src/vendor/ is ignored by src/.gitignore, but lib/vendor/ is not.
	want := []string{"lib/vendor/ok.ts", "src/app.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_GitignoreAnchoredPattern(t *testing.T) {
	dir := setupTree(t, map[string]string{
		// Root .gitignore with anchored patterns (contain /).
		".gitignore":                  "src/generated/\nbuild/output\n",
		"src/app.ts":                  "",
		"src/generated/models.ts":     "",
		"src/generated/api/client.ts": "",
		"build/output/bundle.ts":      "",
		"build/config.ts":             "",
		// "generated" dir elsewhere should NOT be ignored (pattern is anchored to root).
		"lib/generated/utils.ts": "",
	})
	s := &Scanner{Extensions: []string{".ts"}, CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"build/config.ts", "lib/generated/utils.ts", "src/app.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_GitignoreAnchoredPatternSubdir(t *testing.T) {
	dir := setupTree(t, map[string]string{
		// Subdirectory .gitignore with anchored pattern.
		"src/.gitignore":          "api/generated/\n",
		"src/app.ts":              "",
		"src/api/generated/v1.ts": "",
		"src/api/handler.ts":      "",
		// Same path under lib/ should NOT be ignored.
		"lib/api/generated/v1.ts": "",
	})
	s := &Scanner{Extensions: []string{".ts"}, CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"lib/api/generated/v1.ts", "src/api/handler.ts", "src/app.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_GitignoreAncestorFromCwd(t *testing.T) {
	// Simulate: CWD has a .gitignore ignoring *.log and dist/,
	// and we scan a child directory. The ancestor .gitignore must apply.
	dir := setupTree(t, map[string]string{
		".gitignore":         "*.log\ndist/\n",
		"src/app.ts":         "",
		"src/debug.log":      "",
		"src/dist/bundle.ts": "",
		"src/lib/util.ts":    "",
	})

	// Resolve symlinks so CWD and absInput agree (macOS /var → /private/var).
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	s := &Scanner{Extensions: []string{".ts", ".log"}, CollectTest: false}
	got, err := s.Scan(filepath.Join(dir, "src"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"app.ts", "lib/util.ts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScan_RelativePaths(t *testing.T) {
	dir := setupTree(t, map[string]string{
		"a/b/c.ts": "",
		"d.ts":     "",
	})
	s := &Scanner{Extensions: []string{".ts"}, CollectTest: false}
	got, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range got {
		if filepath.IsAbs(p) {
			t.Errorf("expected relative path, got absolute: %s", p)
		}
	}
}

// --- Property-based tests using pgregory.net/rapid ---

// safeName generates a short alphanumeric filename component (no special glob chars).
func safeName() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(1, 8).Draw(t, "len")
		chars := make([]byte, n)
		for i := range chars {
			chars[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[rapid.IntRange(0, 35).Draw(t, "ch")]
		}
		return string(chars)
	})
}

// safeExt generates a file extension like ".ts", ".go", etc.
func safeExt() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		return "." + safeName().Draw(t, "ext")
	})
}

// rapidTempDir creates a temp directory for use inside rapid.Check callbacks.
func rapidTempDir(t *rapid.T) string {
	dir, err := os.MkdirTemp("", "scanner-pbt-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// Feature: code-concept-mapper, Property 3: Scanner Extension Filtering
func TestProperty_ScannerExtensionFiltering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		numRegistered := rapid.IntRange(1, 4).Draw(t, "numRegistered")
		numUnregistered := rapid.IntRange(1, 4).Draw(t, "numUnregistered")

		registered := make([]string, numRegistered)
		allExts := make(map[string]bool)
		for i := range registered {
			for {
				ext := safeExt().Draw(t, "regExt")
				if !allExts[ext] {
					registered[i] = ext
					allExts[ext] = true
					break
				}
			}
		}
		unregistered := make([]string, numUnregistered)
		for i := range unregistered {
			for {
				ext := safeExt().Draw(t, "unregExt")
				if !allExts[ext] {
					unregistered[i] = ext
					allExts[ext] = true
					break
				}
			}
		}

		files := make(map[string]string)
		numFiles := rapid.IntRange(1, 10).Draw(t, "numFiles")
		for i := 0; i < numFiles; i++ {
			name := safeName().Draw(t, "fname")
			allExtSlice := make([]string, 0, len(allExts))
			for e := range allExts {
				allExtSlice = append(allExtSlice, e)
			}
			sort.Strings(allExtSlice)
			ext := allExtSlice[rapid.IntRange(0, len(allExtSlice)-1).Draw(t, "extIdx")]
			files[name+ext] = ""
		}

		dir := rapidTempDir(t)
		populateTree(t, dir, files)
		s := &Scanner{Extensions: registered, CollectTest: true}
		got, err := s.Scan(dir)
		if err != nil {
			t.Fatal(err)
		}

		regSet := make(map[string]bool, len(registered))
		for _, e := range registered {
			regSet[e] = true
		}

		for _, p := range got {
			ext := filepath.Ext(p)
			if !regSet[ext] {
				t.Errorf("returned file %q with unregistered extension %q", p, ext)
			}
		}
	})
}

// Feature: code-concept-mapper, Property 4: Scanner Test File Filtering
func TestProperty_ScannerTestFileFiltering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ext := ".ts"
		// Generate a mix of normal and test files.
		numNormal := rapid.IntRange(1, 5).Draw(t, "numNormal")
		numTest := rapid.IntRange(1, 5).Draw(t, "numTest")

		files := make(map[string]string)
		var normalPaths []string
		var testPaths []string

		for i := 0; i < numNormal; i++ {
			name := safeName().Draw(t, "normalName")
			p := "src/" + name + ext
			files[p] = ""
			normalPaths = append(normalPaths, p)
		}

		testPatterns := []string{".test.", ".spec."}
		for i := 0; i < numTest; i++ {
			name := safeName().Draw(t, "testName")
			patIdx := rapid.IntRange(0, len(testPatterns)-1).Draw(t, "patIdx")
			var p string
			if rapid.Bool().Draw(t, "useTestDir") {
				p = "src/__tests__/" + name + ext
			} else {
				p = "src/" + name + testPatterns[patIdx] + ext[1:]
			}
			files[p] = ""
			testPaths = append(testPaths, p)
		}

		dir := rapidTempDir(t)
		populateTree(t, dir, files)

		// With CollectTest=false, test files must be excluded.
		s := &Scanner{Extensions: []string{ext}, TestPatterns: allTestPatterns(), CollectTest: false}
		got, err := s.Scan(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range got {
			if s.isTestFile(p) {
				t.Errorf("CollectTest=false but got test file: %q", p)
			}
		}

		// With CollectTest=true, test files must be included.
		s2 := &Scanner{Extensions: []string{ext}, TestPatterns: allTestPatterns(), CollectTest: true}
		got2, err := s2.Scan(dir)
		if err != nil {
			t.Fatal(err)
		}
		gotSet := make(map[string]bool, len(got2))
		for _, p := range got2 {
			gotSet[p] = true
		}
		for _, tp := range testPaths {
			if !gotSet[tp] {
				t.Errorf("CollectTest=true but missing test file: %q", tp)
			}
		}
		_ = normalPaths
	})
}

// Feature: code-concept-mapper, Property 5: Scanner Relative Paths
func TestProperty_ScannerRelativePaths(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		numFiles := rapid.IntRange(1, 10).Draw(t, "numFiles")
		files := make(map[string]string)
		for i := 0; i < numFiles; i++ {
			depth := rapid.IntRange(0, 3).Draw(t, "depth")
			parts := make([]string, depth+1)
			for j := 0; j < depth; j++ {
				parts[j] = safeName().Draw(t, "dir")
			}
			parts[depth] = safeName().Draw(t, "file") + ".ts"
			files[strings.Join(parts, "/")] = ""
		}

		dir := rapidTempDir(t)
		populateTree(t, dir, files)
		s := &Scanner{Extensions: []string{".ts"}, CollectTest: true}
		got, err := s.Scan(dir)
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range got {
			if filepath.IsAbs(p) {
				t.Errorf("got absolute path: %s", p)
			}
			if strings.HasPrefix(p, "..") {
				t.Errorf("path escapes input directory: %s", p)
			}
			if strings.HasPrefix(p, "/") {
				t.Errorf("path starts with /: %s", p)
			}
		}
	})
}

// Feature: code-concept-mapper, Property 6: Scanner Gitignore Application
func TestProperty_ScannerGitignoreApplication(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a set of filenames to ignore via gitignore glob patterns.
		numIgnored := rapid.IntRange(1, 3).Draw(t, "numIgnored")
		numKept := rapid.IntRange(1, 3).Draw(t, "numKept")

		ext := ".ts"
		files := make(map[string]string)
		var ignoredNames []string
		var keptNames []string

		for i := 0; i < numIgnored; i++ {
			name := safeName().Draw(t, "ignoredName")
			ignoredNames = append(ignoredNames, name)
			files["src/"+name+ext] = ""
		}
		for i := 0; i < numKept; i++ {
			name := safeName().Draw(t, "keptName")
			keptNames = append(keptNames, name)
			files["src/"+name+ext] = ""
		}

		// Build a .gitignore at root that ignores the specific files.
		gitignoreLines := make([]string, 0, len(ignoredNames))
		for _, name := range ignoredNames {
			gitignoreLines = append(gitignoreLines, name+ext)
		}
		files[".gitignore"] = strings.Join(gitignoreLines, "\n") + "\n"

		dir := rapidTempDir(t)
		populateTree(t, dir, files)
		s := &Scanner{Extensions: []string{ext}, CollectTest: true}
		got, err := s.Scan(dir)
		if err != nil {
			t.Fatal(err)
		}

		gotSet := make(map[string]bool, len(got))
		for _, p := range got {
			gotSet[p] = true
		}

		// Verify ignored files are excluded.
		for _, name := range ignoredNames {
			p := "src/" + name + ext
			if gotSet[p] {
				t.Errorf("gitignored file should be excluded: %q", p)
			}
		}

		// Verify kept files are included (unless they collide with an ignored name).
		ignoredSet := make(map[string]bool, len(ignoredNames))
		for _, n := range ignoredNames {
			ignoredSet[n] = true
		}
		for _, name := range keptNames {
			if ignoredSet[name] {
				continue // name collision, skip check
			}
			p := "src/" + name + ext
			if !gotSet[p] {
				t.Errorf("non-ignored file should be included: %q", p)
			}
		}
	})
}

// Feature: cli-output-modes, Property 1: Single-file scanner correctness
func TestProperty_SingleFileScannerCorrectness(t *testing.T) {
	supportedExts := []string{".ts", ".go", ".js", ".py", ".rb", ".rs"}

	rapid.Check(t, func(t *rapid.T) {
		// Pick whether this iteration uses a supported or unsupported extension.
		useSupported := rapid.Bool().Draw(t, "useSupported")

		var ext string
		if useSupported {
			ext = supportedExts[rapid.IntRange(0, len(supportedExts)-1).Draw(t, "extIdx")]
		} else {
			// Generate a random extension that is NOT in the supported set.
			for {
				ext = safeExt().Draw(t, "unsupExt")
				found := false
				for _, se := range supportedExts {
					if ext == se {
						found = true
						break
					}
				}
				if !found {
					break
				}
			}
		}

		// Create a temp directory and a single file inside it.
		dir := rapidTempDir(t)
		fileName := safeName().Draw(t, "fileName") + ext
		filePath := filepath.Join(dir, fileName)
		if err := os.WriteFile(filePath, []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}

		s := &Scanner{Extensions: supportedExts, CollectTest: true}
		got, err := s.Scan(filePath)
		if err != nil {
			t.Fatalf("Scan(%q) returned error: %v", filePath, err)
		}

		if useSupported {
			// Supported extension → exactly one element equal to the basename.
			if len(got) != 1 {
				t.Fatalf("supported ext %q: expected 1 result, got %d: %v", ext, len(got), got)
			}
			if got[0] != fileName {
				t.Errorf("supported ext %q: expected %q, got %q", ext, fileName, got[0])
			}
		} else if len(got) != 0 {
			// Unsupported extension → empty list.
			t.Errorf("unsupported ext %q: expected empty list, got %v", ext, got)
		}
	})
}
