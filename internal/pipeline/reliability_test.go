// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/scanpolicy"
)

func TestBuildGraphSingleFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "app.go")
	if err := os.WriteFile(file, []byte("package app\nfunc Read() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry()
	registry.Register(golang.NewPlugin())
	result, err := BuildGraph(&BuildOptions{InputPath: file}, registry, 0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, symbol := range result.Graph.Symbols {
		if symbol.Name == "Read" {
			return
		}
	}
	t.Fatalf("single-file input lost its declarations: %+v", result.ParseErrors)
}

func TestBuildGraphStrictAndPartialResults(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{
		"good.go":  "package app\nfunc Read() {}\n",
		"large.go": strings.Repeat(" ", 200),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	registry := plugin.NewRegistry()
	registry.Register(golang.NewPlugin())
	opts := BuildOptions{InputPath: dir, Scan: scanpolicy.Options{MaxFileSize: 100}}
	partial, err := BuildGraph(&opts, registry, 0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(partial.Skipped) != 1 || len(partial.Graph.Errors) != 1 {
		t.Fatalf("partial results lost their diagnostics: %+v", partial)
	}
	opts.Scan.Strict = true
	if _, err := BuildGraph(&opts, registry, 0, nil, nil); err == nil || !strings.Contains(err.Error(), "strict") {
		t.Fatalf("strict analysis accepted incomplete output: %v", err)
	}
}

func TestSourceReadsEnforceBoundaryAndSizeAfterScan(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.go")
	if err := os.WriteFile(file, []byte("package app\nfunc Read() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := openSourceReader(ReadOptions{Root: dir, MaxFileSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.root.Close() }()
	if _, err := reader.read(file); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(strings.Repeat("x", 101)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.read(file); err == nil {
		t.Fatal("source growth bypassed read-time size check")
	}
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("package outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, file); err != nil {
		t.Skip(err)
	}
	if _, err := reader.read(file); err == nil {
		t.Fatal("replacement symlink bypassed scan boundary")
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Symlink(filepath.Dir(outside), sub); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.read(filepath.Join(sub, "outside.go")); err == nil {
		t.Fatal("intermediate symlink escaped scan root")
	}
}

func TestFingerprintUsesParsedSourceSnapshot(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.go")
	source := "package app\nfunc Read(x int) int { if x > 0 { return x + 1 }; return 0 }\n"
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry()
	registry.Register(golang.NewPlugin())
	result, err := BuildGraph(&BuildOptions{InputPath: dir, Fingerprint: true}, registry, 0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	for _, symbol := range result.Graph.Symbols {
		if symbol.Name == "Read" && strings.Contains(symbol.SourceBody, "return x + 1") {
			return
		}
	}
	t.Fatal("fingerprint graph lost the parsed source snapshot")
}

func TestBuildGraphAllReadsFail(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.mock"), []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry()
	registry.Register(&mockPlugin{exts: []string{".mock"}, err: errors.New("read failed")})
	if _, err := BuildGraph(&BuildOptions{InputPath: dir}, registry, 0, nil, nil); err == nil {
		t.Fatal("a completely failed analysis must not report success")
	}
}
