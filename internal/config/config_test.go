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
	if cfg.OutputMode != DefaultParseOutputMode {
		t.Fatalf("expected OutputMode to default to %q, got %q", DefaultParseOutputMode, cfg.OutputMode)
	}
}

func TestValidate_OutputFormatDefaultsToSKT(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir()},
		OutputDir: t.TempDir(),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputFormat != DefaultParseOutputFormat {
		t.Fatalf("expected OutputFormat to default to %q, got %q", DefaultParseOutputFormat, cfg.OutputFormat)
	}
}

func TestValidate_MaxLinesUsesDefault(t *testing.T) {
	cfg := ParseConfig{
		Common:    Common{InputPath: t.TempDir()},
		OutputDir: t.TempDir(),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxLines != DefaultParseMaxLines {
		t.Fatalf("expected MaxLines to default to %d, got %d", DefaultParseMaxLines, cfg.MaxLines)
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
	if cfg.OutputDir != DefaultParseOutputDir {
		t.Fatalf("expected OutputDir to default to %q, got %q", DefaultParseOutputDir, cfg.OutputDir)
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

func TestValidate_InvalidOutputFormat(t *testing.T) {
	cfg := ParseConfig{
		Common:       Common{InputPath: t.TempDir()},
		OutputDir:    t.TempDir(),
		OutputFormat: "yaml",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid output format")
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

func TestFingerprintValidate_DefaultSimilarityRange(t *testing.T) {
	cfg := FingerprintConfig{Common: Common{InputPath: t.TempDir()}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinSim != DefaultFingerprintMinSimilarity {
		t.Fatalf("MinSim default: got %d, want %d", cfg.MinSim, DefaultFingerprintMinSimilarity)
	}
	if cfg.MaxSim != DefaultFingerprintMaxSimilarity {
		t.Fatalf("MaxSim default: got %d, want %d", cfg.MaxSim, DefaultFingerprintMaxSimilarity)
	}
	if cfg.Output != DefaultFingerprintOutput {
		t.Fatalf("Output default: got %q, want %q", cfg.Output, DefaultFingerprintOutput)
	}
}

func TestGraphValidate_UsesDefaults(t *testing.T) {
	cfg := GraphConfig{Common: Common{InputPath: t.TempDir()}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != DefaultGraphOutput {
		t.Fatalf("Output default: got %q, want %q", cfg.Output, DefaultGraphOutput)
	}
}

func TestAnalyzeValidate_UsesDefaults(t *testing.T) {
	cfg := AnalyzeConfig{Common: Common{InputPath: t.TempDir()}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != DefaultAnalyzeOutput {
		t.Errorf("Output default: got %q, want %q", cfg.Output, DefaultAnalyzeOutput)
	}
	if cfg.FanThreshold != DefaultAnalyzeFanThreshold {
		t.Errorf("FanThreshold default: got %d, want %d", cfg.FanThreshold, DefaultAnalyzeFanThreshold)
	}
	if cfg.GodThreshold != DefaultAnalyzeGodThreshold {
		t.Errorf("GodThreshold default: got %d, want %d", cfg.GodThreshold, DefaultAnalyzeGodThreshold)
	}
	if cfg.MaxInheritanceDepth != DefaultAnalyzeMaxInheritanceDepth {
		t.Errorf("MaxInheritanceDepth default: got %d, want %d", cfg.MaxInheritanceDepth, DefaultAnalyzeMaxInheritanceDepth)
	}
	if cfg.TopN != DefaultAnalyzeTopN {
		t.Errorf("TopN default: got %d, want %d", cfg.TopN, DefaultAnalyzeTopN)
	}
	if cfg.BetweennessThreshold != DefaultAnalyzeBetweennessThreshold {
		t.Errorf("BetweennessThreshold default: got %g, want %g", cfg.BetweennessThreshold, DefaultAnalyzeBetweennessThreshold)
	}
	if cfg.PropagationCutoff != DefaultAnalyzePropagationCutoff {
		t.Errorf("PropagationCutoff default: got %g, want %g", cfg.PropagationCutoff, DefaultAnalyzePropagationCutoff)
	}
}

func TestResolveFingerprintEmbedModel(t *testing.T) {
	tests := []struct {
		name          string
		modelOverride string
		want          string
		rerank        bool
	}{
		{name: "disabled", want: ""},
		{name: "rerank default", rerank: true, want: DefaultFingerprintEmbedModel},
		{name: "override implies rerank", modelOverride: "custom-model", want: "custom-model"},
		{name: "override wins", rerank: true, modelOverride: "custom-model", want: "custom-model"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveFingerprintEmbedModel(tc.rerank, tc.modelOverride)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
