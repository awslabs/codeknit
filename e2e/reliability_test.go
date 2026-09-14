// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeknit/internal/emitter/parser"
)

func runBoundedCLI(t *testing.T, cwd string, args ...string) (output, diagnostic string, runErr error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath, args...) //nolint:gosec // controlled CLI regression test
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not terminate: %v", args)
	}
	return stdout.String(), stderr.String(), err
}

func TestE2E_SingleFileAllOutputModes(t *testing.T) {
	input := writeFixture(t, map[string]string{"app.go": "package app\nfunc Read() {}\n"})
	cwd, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(input, "app.go")
	relative, err := filepath.Rel(cwd, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{file, relative} {
		for _, mode := range []string{"inline", "directory-flat", "directory-tree"} {
			for _, format := range []string{"json", "skt"} {
				t.Run(fmt.Sprintf("%s/%s/absolute=%v", mode, format, filepath.IsAbs(path)), func(t *testing.T) {
					out := filepath.Join(t.TempDir(), "output")
					stdout, stderr, err := runBoundedCLI(t, cwd, "parse", path, out, "--output-mode", mode, "--format", format)
					if err != nil {
						t.Fatalf("single-file parse failed: %v\n%s", err, stderr)
					}
					content := stdout
					if mode != "inline" {
						name := "map_001.skt"
						if mode == "directory-tree" {
							name = "app.go.skt"
						}
						if format == "json" {
							name = "codeknit.json"
						}
						data, err := os.ReadFile(filepath.Join(out, name)) //nolint:gosec // generated test output
						if err != nil {
							t.Fatal(err)
						}
						content = string(data)
					}
					if !strings.Contains(content, "Read") {
						t.Fatalf("source declaration disappeared:\n%s\n%s", content, stderr)
					}
					if format == "json" && !json.Valid([]byte(content)) {
						t.Fatalf("invalid JSON output:\n%s", content)
					}
				})
			}
		}
	}
}

func TestE2E_TreeTinyChunksAndSameStemSources(t *testing.T) {
	input := writeFixture(t, map[string]string{
		"foo.c":       "int Implementation(void) { return 1; }\n",
		"foo.h":       "int Declaration(void);\n",
		"app.go":      "package app\nfunc First() { Second() }\nfunc Second() {}\n",
		"dict.go":     "package app\nconst Dictionary = 1\n",
		"warnings.go": "package app\nconst Warnings = 2\n",
	})
	out := filepath.Join(t.TempDir(), "out")
	_, stderr, err := runBoundedCLI(t, input, "parse", input, out, "--output-mode", "directory-tree", "--max-lines", "3", "--edges", "--minify")
	if err != nil {
		t.Fatalf("tiny chunk parse failed: %v\n%s", err, stderr)
	}
	var readers []io.Reader
	dict, err := os.ReadFile(filepath.Join(out, "dict.skt")) //nolint:gosec // generated test output
	if err != nil {
		t.Fatal(err)
	}
	readers = append(readers, bytes.NewReader(dict))
	err = filepath.WalkDir(out, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || entry.Name() == "dict.skt" {
			return walkErr
		}
		content, readErr := os.ReadFile(path) //nolint:gosec // generated test output
		if readErr != nil {
			return readErr
		}
		if strings.Count(string(content), "\n") > 3 {
			t.Errorf("source chunk exceeds line limit: %s", path)
		}
		readers = append(readers, bytes.NewReader(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	graph, err := parser.ParseOutput(readers, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Implementation", "Declaration", "First", "Second", "Dictionary", "Warnings"} {
		found := false
		for _, symbol := range graph.Symbols {
			if strings.Contains(symbol.Signature, name) {
				found = true
				if symbol.FilePath == "" {
					t.Errorf("%s lost its source path in chunk round trip", name)
				}
			}
		}
		if !found {
			t.Errorf("tree output lost %s", name)
		}
	}
	if len(graph.Edges) == 0 {
		t.Error("chunk round trip lost relationships")
	}
	for _, limit := range []string{"1", "2"} {
		_, stderr, err := runBoundedCLI(t, input, "parse", input, out, "--output-mode", "directory-tree", "--max-lines", limit, "--clean")
		if err == nil || !strings.Contains(stderr, "at least 3") {
			t.Fatalf("invalid chunk limit must fail promptly: %v\n%s", err, stderr)
		}
	}
}

func TestE2E_ScanControlsAndStrictFailures(t *testing.T) {
	input := writeFixture(t, map[string]string{
		".gitignore":           "ignored.go\n/src/generated/\n",
		"src/app.go":           "package app\nfunc Included() {}\n",
		"src/ignored.go":       "package app\nfunc Ignored() {}\n",
		"src/generated/gen.go": "package app\nfunc Generated() {}\n",
		"src/excluded.go":      "package app\nfunc Excluded() {}\n",
		"src/large.go":         strings.Repeat(" ", 200),
	})
	source := filepath.Join(input, "src")
	args := []string{"parse", source, "--output-mode", "inline", "--format", "json", "--include", "*.go", "--exclude", "excluded.go", "--max-file-size", "100"}
	for _, cwd := range []string{input, t.TempDir()} {
		stdout, stderr, err := runBoundedCLI(t, cwd, args...)
		if err != nil || !json.Valid([]byte(stdout)) {
			t.Fatalf("partial analysis failed: %v\n%s\n%s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "Included") || strings.Contains(stdout, "Generated") || strings.Contains(stdout, "Ignored") || strings.Contains(stdout, "Excluded") {
			t.Fatalf("scan selection changed at cwd %s:\n%s", cwd, stdout)
		}
		if !strings.Contains(stdout, "\"errors\"") || !strings.Contains(stdout, "max-file-size") {
			t.Fatalf("JSON lost scan diagnostic:\n%s", stdout)
		}
		stdout, stderr, err = runBoundedCLI(t, cwd, append(args, "--strict")...)
		if err == nil || stdout != "" || !strings.Contains(stderr, "strict") {
			t.Fatalf("strict mode emitted successful partial output: %v\n%s\n%s", err, stdout, stderr)
		}
	}
	out := filepath.Join(t.TempDir(), "out")
	if err := os.MkdirAll(out, 0o700); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(out, "old.skt")
	if err := os.WriteFile(old, []byte("previous result"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, stderr, err := runBoundedCLI(t, input, "parse", source, out, "--strict", "--max-file-size", "100", "--clean")
	if err == nil {
		t.Fatalf("strict failure returned success: %s", stderr)
	}
	data, err := os.ReadFile(old) //nolint:gosec // previous test output
	if err != nil || string(data) != "previous result" {
		t.Fatalf("strict failure changed existing output: %q, %v", data, err)
	}
}

func TestE2E_AllSourceReadsFail(t *testing.T) {
	input := writeFixture(t, map[string]string{"app.go": "package app\n"})
	file := filepath.Join(input, "app.go")
	if err := os.Chmod(file, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(file, 0o600) })
	if _, err := os.ReadFile(file); err == nil { //nolint:gosec // test source permission probe
		t.Skip("source permissions are not enforced for this user")
	}
	stdout, stderr, err := runBoundedCLI(t, input, "parse", file, "--output-mode", "inline", "--format", "json")
	if err == nil || stdout != "" || !strings.Contains(stderr, "no source files could be parsed") {
		t.Fatalf("total read failure must fail without empty successful JSON: %v\n%s\n%s", err, stdout, stderr)
	}
}

func TestE2E_ExactCloneThresholdAndDeterminism(t *testing.T) {
	var source strings.Builder
	source.WriteString("package clones\n")
	for i := range 500 {
		fmt.Fprintf(&source, "func F%03d(x int) int { if x > 0 { x += 2 }; for x < 10 { x++ }; return x }\n", i)
	}
	input := writeFixture(t, map[string]string{"clones.go": source.String()})
	out := filepath.Join(t.TempDir(), "fingerprints.skt")
	args := []string{"fingerprint", input, "-o", out, "--min-similarity", "100", "--max-similarity", "100"}
	_, stderr, err := runBoundedCLI(t, input, args...)
	if err != nil {
		t.Fatalf("fingerprint failed: %v\n%s", err, stderr)
	}
	first, err := os.ReadFile(out) //nolint:gosec // generated test output
	if err != nil {
		t.Fatal(err)
	}
	if got := len(parseDuplicates(string(first))); got != 124750 {
		t.Fatalf("500 exact clones: got %d pairs, want 124750", got)
	}
	if !strings.Contains(string(first), "approximate near-duplicate") {
		t.Fatal("report must disclose approximate candidate retrieval")
	}
	_, stderr, err = runBoundedCLI(t, input, args...)
	if err != nil {
		t.Fatalf("second fingerprint failed: %v\n%s", err, stderr)
	}
	second, err := os.ReadFile(out) //nolint:gosec // generated test output
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("unchanged inputs produced different reports: %v", err)
	}
}
