// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"
	"codeknit/internal/ir"

	"pgregory.net/rapid"
)

// snapshotDir is the directory where golden snapshot files are stored (inside e2e/).
const snapshotDir = "snapshots"

// updateSnapshots returns true when the UPDATE_SNAPSHOTS env var is set,
// causing assertSnapshot to write/overwrite golden files instead of comparing.
func updateSnapshots() bool {
	return os.Getenv("UPDATE_SNAPSHOTS") == "1"
}

// normalizeFilePath strips the temp directory prefix from a file path,
// returning only the fixture-relative portion (e.g., "pkg/models/user.go").
func normalizeFilePath(p string) string {
	// The fixture temp dirs contain a pattern like "TestE2E_Go-*/" or similar.
	// We find the last component that looks like a temp dir and return everything after.
	parts := strings.Split(filepath.ToSlash(p), "/")
	for i, part := range parts {
		if strings.Contains(part, "fixtures") {
			// The next part is the temp dir name (e.g., "TestE2E_Go-12345"),
			// and everything after that is the relative path.
			if i+2 < len(parts) {
				return strings.Join(parts[i+2:], "/")
			}
		}
	}
	// Fallback: return the last 3 components.
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return filepath.ToSlash(p)
}

// assertSnapshot compares the raw CLI output files from outputDir against a
// golden file. Absolute fixture paths in the output are replaced with the
// relative portion so snapshots are deterministic across machines.
// If UPDATE_SNAPSHOTS=1, it writes/overwrites the golden file instead.
func assertSnapshot(t *testing.T, outputDir, inputDir string) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("assertSnapshot: glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("assertSnapshot: no map_*.skt output files found")
	}
	sort.Strings(matches)

	// If dict.skt exists, prepend it so the snapshot includes the dictionary.
	dictPath := filepath.Join(outputDir, "dict.skt")
	if _, statErr := os.Stat(dictPath); statErr == nil {
		matches = append([]string{dictPath}, matches...)
	}

	parts := make([]string, 0, len(matches))
	for _, path := range matches {
		data, readErr := os.ReadFile(path) //nolint:gosec // test output file
		if readErr != nil {
			t.Fatalf("assertSnapshot: read %s: %v", path, readErr)
		}
		parts = append(parts, string(data))
	}
	got := strings.Join(parts, "")
	// Normalize absolute fixture paths to relative.
	got = strings.ReplaceAll(got, inputDir+"/", "")

	snapFile := filepath.Join(snapshotDir, t.Name()+".snap")

	if updateSnapshots() {
		if mkdirErr := os.MkdirAll(filepath.Dir(snapFile), 0o750); mkdirErr != nil {
			t.Fatalf("assertSnapshot: mkdir: %v", mkdirErr)
		}
		if writeErr := os.WriteFile(snapFile, []byte(got), 0o600); writeErr != nil { //nolint:gosec // snapshot path is derived from test name
			t.Fatalf("assertSnapshot: write: %v", writeErr)
		}
		t.Logf("snapshot updated: %s", snapFile)
		return
	}

	want, err := os.ReadFile(snapFile) //nolint:gosec // snapshot path is derived from test name, not user input
	if err != nil {
		t.Fatalf("assertSnapshot: snapshot file not found: %s\nRun with UPDATE_SNAPSHOTS=1 to create it.\nGot:\n%s", snapFile, got)
	}

	if got != string(want) {
		t.Errorf("snapshot mismatch for %s\nRun with UPDATE_SNAPSHOTS=1 to update.\n\n--- want ---\n%s\n--- got ---\n%s", snapFile, string(want), got)
	}
}

// assertTreeSnapshot compares every .skt file in outputDir against individual
// golden files stored under snapshots/<testName>/. Each output file is matched
// one-to-one by its relative path. If UPDATE_SNAPSHOTS=1, golden files are
// written/overwritten instead of compared.
func assertTreeSnapshot(t *testing.T, outputDir, inputDir string) {
	t.Helper()

	snapBase := filepath.Join(snapshotDir, t.Name())

	// Collect all .skt files from the output directory.
	var outputFiles []string
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		if filepath.Ext(path) == ".skt" {
			outputFiles = append(outputFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("assertTreeSnapshot: walk: %v", err)
	}
	if len(outputFiles) == 0 {
		t.Fatal("assertTreeSnapshot: no .skt output files found")
	}
	sort.Strings(outputFiles)

	if updateSnapshots() {
		// Wipe old snapshot dir and recreate from scratch.
		_ = os.RemoveAll(snapBase)
		for _, path := range outputFiles {
			rel, _ := filepath.Rel(outputDir, path)
			data, readErr := os.ReadFile(path) //nolint:gosec // test output file
			if readErr != nil {
				t.Fatalf("assertTreeSnapshot: read %s: %v", path, readErr)
			}
			content := strings.ReplaceAll(string(data), inputDir+"/", "")
			snapPath := filepath.Join(snapBase, rel)
			if mkdirErr := os.MkdirAll(filepath.Dir(snapPath), 0o750); mkdirErr != nil {
				t.Fatalf("assertTreeSnapshot: mkdir: %v", mkdirErr)
			}
			if writeErr := os.WriteFile(snapPath, []byte(content), 0o600); writeErr != nil { //nolint:gosec // snapshot path
				t.Fatalf("assertTreeSnapshot: write: %v", writeErr)
			}
		}
		t.Logf("tree snapshot updated: %s (%d files)", snapBase, len(outputFiles))
		return
	}

	// Compare each output file against its golden counterpart.
	for _, path := range outputFiles {
		rel, _ := filepath.Rel(outputDir, path)
		snapPath := filepath.Join(snapBase, rel)

		got, readErr := os.ReadFile(path) //nolint:gosec // test output file
		if readErr != nil {
			t.Fatalf("assertTreeSnapshot: read %s: %v", path, readErr)
		}
		gotStr := strings.ReplaceAll(string(got), inputDir+"/", "")

		want, wantErr := os.ReadFile(snapPath) //nolint:gosec // snapshot path
		if wantErr != nil {
			t.Fatalf("assertTreeSnapshot: golden file not found: %s\nRun with UPDATE_SNAPSHOTS=1 to create it.\nGot:\n%s", snapPath, gotStr)
		}

		if gotStr != string(want) {
			t.Errorf("snapshot mismatch for %s\nRun with UPDATE_SNAPSHOTS=1 to update.\n\n--- want ---\n%s\n--- got ---\n%s", rel, string(want), gotStr)
		}
	}

	// Check for stale golden files that no longer have a corresponding output file.
	outputSet := make(map[string]struct{}, len(outputFiles))
	for _, path := range outputFiles {
		rel, _ := filepath.Rel(outputDir, path)
		outputSet[rel] = struct{}{}
	}
	_ = filepath.Walk(snapBase, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		rel, _ := filepath.Rel(snapBase, path)
		if _, ok := outputSet[rel]; !ok {
			t.Errorf("stale snapshot file %s has no corresponding output", rel)
		}
		return nil
	})
}

// binPath is the path to the compiled codeknit binary, set by TestMain.
var binPath string

func TestMain(m *testing.M) {
	// Ensure fixtures directory exists.
	fixturesBase, err := filepath.Abs("fixtures")
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: fixtures abs: %v\n", err)
		os.Exit(1)
	}
	if mkdirErr := os.MkdirAll(fixturesBase, 0o750); mkdirErr != nil {
		fmt.Fprintf(os.Stderr, "e2e: mkdir fixtures: %v\n", mkdirErr)
		os.Exit(1)
	}

	tmp, err := os.MkdirTemp(fixturesBase, "bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binPath = filepath.Join(tmp, "codeknit")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/codeknit") //nolint:gosec // fixed build args
	cmd.Dir = ".."                                                      // e2e/ is one level below project root
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: build failed: %v\n%s\n", err, out)
		_ = os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

// sanitizeName replaces path separators in test names so they are safe for
// use in os.MkdirTemp patterns (subtests contain '/').
func sanitizeName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

// fixturesDir returns the absolute path to e2e/fixtures and ensures it exists.
func fixturesDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("fixtures")
	if err != nil {
		t.Fatalf("fixturesDir: abs: %v", err)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("fixturesDir: mkdir: %v", err)
	}
	return dir
}

// writeFixture creates a uniquely-named subdirectory under e2e/fixtures/ and
// writes all files from the map. The directory is removed when the test ends.
// Keys are slash-separated relative paths; values are file contents.
func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	base := fixturesDir(t)
	dir, err := os.MkdirTemp(base, sanitizeName(t.Name())+"-*")
	if err != nil {
		t.Fatalf("writeFixture: mkdirtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	for rel, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
			t.Fatalf("writeFixture: mkdir %s: %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
			t.Fatalf("writeFixture: write %s: %v", abs, err)
		}
	}
	return dir
}

// tHelper is satisfied by both *testing.T and *rapid.T for use in shared helpers.
type tHelper interface {
	Fatalf(string, ...any)
}

// collectOutputReaders returns readers for all output files in the given directory,
// with dict.skt (if present) as the first reader so the parser can resolve minified codes.
// It supports both flat mode (map_*.skt) and tree mode (recursive *.skt) output.
func collectOutputReaders(t tHelper, outputDir string) (readers []io.Reader, closers []io.Closer) {
	// If dict.skt exists, include it first.
	dictPath := filepath.Join(outputDir, "dict.skt")
	if _, err := os.Stat(dictPath); err == nil {
		f, openErr := os.Open(dictPath) //nolint:gosec // test output file
		if openErr != nil {
			t.Fatalf("open dict file %s: %v", dictPath, openErr)
		}
		closers = append(closers, f)
		readers = append(readers, f)
	}

	// Try flat mode first (map_*.skt).
	matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("glob output files: %v", err)
	}

	// Fall back to tree mode: walk for all *.skt files (excluding dict.skt).
	if len(matches) == 0 {
		_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return walkErr
			}
			if filepath.Ext(path) == ".skt" && filepath.Base(path) != "dict.skt" {
				matches = append(matches, path)
			}
			return nil
		})
	}

	if len(matches) == 0 {
		t.Fatalf("no .skt output files found in %s", outputDir)
	}
	sort.Strings(matches)

	for _, path := range matches {
		f, openErr := os.Open(path) //nolint:gosec // test output file
		if openErr != nil {
			t.Fatalf("open output file %s: %v", path, openErr)
		}
		closers = append(closers, f)
		readers = append(readers, f)
	}
	return readers, closers
}

// defaultParseFlags are prepended to every helper invocation so existing
// tests keep the old behavior (edges included in output). Tests that need
// to validate the default (no-edges) behavior call the binary directly.
var defaultParseFlags = []string{"--edges"}

// runcodeknit executes the CLI binary against inputDir and returns the parsed
// SymbolGraph and the output directory path.
func runcodeknit(t *testing.T, inputDir string, flags ...string) (sg *ir.SymbolGraph, outputDir string) {
	t.Helper()
	base := fixturesDir(t)
	outputDir, err := os.MkdirTemp(base, sanitizeName(t.Name())+"-out-*")
	if err != nil {
		t.Fatalf("runcodeknit: mkdirtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(outputDir) })
	args := make([]string, 0, len(flags)+len(defaultParseFlags)+3)
	args = append(args, "parse")
	args = append(args, defaultParseFlags...)
	args = append(args, flags...)
	args = append(args, inputDir, outputDir)
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper with controlled args
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("codeknit failed: %v\nstderr: %s", runErr, stderr.String())
	}

	readers, closers := collectOutputReaders(t, outputDir)
	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()

	sg, err = parser.ParseOutput(readers, false)
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}
	return sg, outputDir
}

// runcodeknitRaw executes the CLI binary and returns the output directory path.
func runcodeknitRaw(t *testing.T, inputDir string, flags ...string) string {
	t.Helper()
	allFlags := make([]string, 0, len(defaultParseFlags)+len(flags))
	allFlags = append(allFlags, defaultParseFlags...)
	allFlags = append(allFlags, flags...)
	return runBinaryRaw(t, "parse", inputDir, allFlags...)
}

// runBinaryRaw executes the CLI binary with the given subcommand against
// inputDir, creates a temp output dir, appends inputDir and outputDir as
// positional args, and returns the output directory path.
func runBinaryRaw(t *testing.T, subcommand, inputDir string, flags ...string) string {
	t.Helper()
	base := fixturesDir(t)
	outputDir, err := os.MkdirTemp(base, sanitizeName(t.Name())+"-out-*")
	if err != nil {
		t.Fatalf("runBinaryRaw: mkdirtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(outputDir) })
	args := make([]string, 0, len(flags)+3)
	args = append(args, subcommand)
	args = append(args, flags...)
	args = append(args, inputDir, outputDir)
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper with controlled args
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("codeknit %s failed: %v\nstderr: %s", subcommand, runErr, stderr.String())
	}
	return outputDir
}

// runParseNoDefaults is like runcodeknit but does NOT inject defaultParseFlags.
// Use this in tests that must exercise the real CLI defaults (e.g., verifying
// that the [edges] section is hidden by default).
func runParseNoDefaults(t *testing.T, inputDir string, flags ...string) *ir.SymbolGraph {
	t.Helper()
	base := fixturesDir(t)
	outputDir, err := os.MkdirTemp(base, sanitizeName(t.Name())+"-out-*")
	if err != nil {
		t.Fatalf("runParseNoDefaults: mkdirtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(outputDir) })

	args := make([]string, 0, len(flags)+3)
	args = append(args, "parse")
	args = append(args, flags...)
	args = append(args, inputDir, outputDir)
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper with controlled args
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("codeknit failed: %v\nstderr: %s", runErr, stderr.String())
	}

	readers, closers := collectOutputReaders(t, outputDir)
	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()

	sg, err := parser.ParseOutput(readers, false)
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}
	return sg
}

// Ensure helpers are referenced so the linter doesn't flag them as unused
// before the per-language test files are added.
var (
	_ = writeFixture
	_ = runcodeknit
	_ = runcodeknitRaw
	_ = runBinaryRaw
	_ = runParseNoDefaults
	_ = assertSnapshot
	_ = assertTreeSnapshot
)

// writeFixtureForProperty creates a fixture directory for use in property-based
// tests (rapid). It uses os.MkdirTemp directly since rapid.T doesn't support
// t.Cleanup. Callers must defer os.RemoveAll on the returned path.
func writeFixtureForProperty(t *rapid.T, files map[string]string) string {
	base, err := filepath.Abs("fixtures")
	if err != nil {
		t.Fatalf("writeFixtureForProperty: abs: %v", err)
	}
	if mkdirErr := os.MkdirAll(base, 0o750); mkdirErr != nil {
		t.Fatalf("writeFixtureForProperty: mkdir: %v", mkdirErr)
	}
	dir, err := os.MkdirTemp(base, "prop-*")
	if err != nil {
		t.Fatalf("writeFixtureForProperty: mkdirtemp: %v", err)
	}
	for rel, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
			t.Fatalf("writeFixtureForProperty: mkdir %s: %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
			t.Fatalf("writeFixtureForProperty: write %s: %v", abs, err)
		}
	}
	return dir
}

// runcodeknitForProperty executes the CLI binary and returns the parsed
// SymbolGraph. For use in property-based tests with rapid.T.
func runcodeknitForProperty(t *rapid.T, inputDir string, flags ...string) *ir.SymbolGraph {
	base, err := filepath.Abs("fixtures")
	if err != nil {
		t.Fatalf("runcodeknitForProperty: abs: %v", err)
	}
	outputDir, err := os.MkdirTemp(base, "prop-out-*")
	if err != nil {
		t.Fatalf("runcodeknitForProperty: mkdirtemp: %v", err)
	}
	defer os.RemoveAll(outputDir) //nolint:errcheck // cleanup

	args := make([]string, 0, len(flags)+len(defaultParseFlags)+3)
	args = append(args, "parse")
	args = append(args, defaultParseFlags...)
	args = append(args, flags...)
	args = append(args, inputDir, outputDir)
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("codeknit failed: %v\nstderr: %s", runErr, stderr.String())
	}

	readers, closers := collectOutputReaders(t, outputDir)
	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()

	sg, parseErr := parser.ParseOutput(readers, false)
	if parseErr != nil {
		t.Fatalf("ParseOutput failed: %v", parseErr)
	}
	return sg
}

// runcodeknitRawForProperty executes the CLI binary and returns the output
// directory path. For use in property-based tests with rapid.T.
// Callers must defer os.RemoveAll on the returned path.
func runcodeknitRawForProperty(t *rapid.T, inputDir string, flags ...string) string {
	base, err := filepath.Abs("fixtures")
	if err != nil {
		t.Fatalf("runcodeknitRawForProperty: abs: %v", err)
	}
	outputDir, err := os.MkdirTemp(base, "prop-out-*")
	if err != nil {
		t.Fatalf("runcodeknitRawForProperty: mkdirtemp: %v", err)
	}

	args := make([]string, 0, len(flags)+len(defaultParseFlags)+3)
	args = append(args, "parse")
	args = append(args, defaultParseFlags...)
	args = append(args, flags...)
	args = append(args, inputDir, outputDir)
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		_ = os.RemoveAll(outputDir)
		t.Fatalf("codeknit failed: %v\nstderr: %s", runErr, stderr.String())
	}
	return outputDir
}

// runcodeknitInline executes the CLI binary with the given args and returns stdout.
func runcodeknitInline(t *testing.T, workDir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper with controlled args
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("codeknit failed: %v\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
