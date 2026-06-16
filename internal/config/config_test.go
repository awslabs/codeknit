// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_InputPathNotExist(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: "/nonexistent/path/that/does/not/exist"},
		OutputDir: t.TempDir(),
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for nonexistent input path, got nil")
	}
	if got := err.Error(); got == "" {
		t.Fatal("expected descriptive error message, got empty string")
	}
}

func TestValidate_InputPathIsFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "afile.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := ParseConfig{
		Common:    Common{InputPath: f},
		OutputDir: t.TempDir(),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error when input path is a file: %v", err)
	}
}

func TestValidate_CreatesOutputDir(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "nested", "output")

	cfg := ParseConfig{
		Common:    Common{InputPath: inputDir},
		OutputDir: outputDir,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(outputDir)
	if err != nil {
		t.Fatalf("output directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("output path is not a directory")
	}
}

func TestValidate_ExistingOutputDir(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	cfg := ParseConfig{
		Common:    Common{InputPath: inputDir},
		OutputDir: outputDir,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error with existing output directory: %v", err)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir(), CollectTest: true},
		OutputDir: t.TempDir(),
		Minify:    true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error for valid config: %v", err)
	}
}

func TestValidate_OutputModeDefaultsToDirectoryFlat(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir()},
		OutputDir: t.TempDir(),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputMode != OutputDirectoryFlat {
		t.Fatalf("expected OutputMode to default to %q, got %q", OutputDirectoryFlat, cfg.OutputMode)
	}
}

func TestValidate_MaxLinesDefaultsTo500(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir()},
		OutputDir: t.TempDir(),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxLines != 500 {
		t.Fatalf("expected MaxLines to default to 500, got %d", cfg.MaxLines)
	}
}

func TestValidate_InlineModeNoOutputDir(t *testing.T) {
	cfg := ParseConfig{
		Common:     Common{InputPath: t.TempDir()},
		OutputMode: OutputInline,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error for inline mode without output dir: %v", err)
	}
}

func TestValidate_DirectoryModeDefaultsOutputDir(t *testing.T) {
	cfg := ParseConfig{
		Common:     Common{InputPath: t.TempDir()},
		OutputMode: OutputDirectoryFlat,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != "./skeleton" {
		t.Fatalf("expected OutputDir to default to %q, got %q", "./skeleton", cfg.OutputDir)
	}
}

func TestValidate_InvalidOutputMode(t *testing.T) {
	cfg := ParseConfig{
		Common:     Common{InputPath: t.TempDir()},
		OutputDir:  t.TempDir(),
		OutputMode: "bogus",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid output mode")
	}
}

func TestValidate_MaxLinesNegative(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir()},
		OutputDir: t.TempDir(),
		MaxLines:  -5,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxLines")
	}
}
