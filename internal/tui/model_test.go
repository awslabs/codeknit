// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"strconv"
	"testing"

	"codeknit/internal/config"

	"pgregory.net/rapid"
)

// TestProperty14_TUIConfigConversion verifies that ToParseConfig() copies
// every user-editable parse field from the TUI model to the resulting
// ParseConfig without mutation.
//
// Feature: cli-output-modes, Property 14: TUI config conversion
func TestProperty14_TUIConfigConversion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		inputDir := rapid.StringMatching(`[a-zA-Z0-9_/]{1,50}`).Draw(t, "inputDir")
		outputDir := rapid.StringMatching(`[a-zA-Z0-9_/]{1,50}`).Draw(t, "outputDir")
		collectTest := rapid.Bool().Draw(t, "collectTest")
		minify := rapid.Bool().Draw(t, "minify")
		edges := rapid.Bool().Draw(t, "edges")
		clean := rapid.Bool().Draw(t, "clean")
		workers := rapid.IntRange(0, 128).Draw(t, "workers")
		outputMode := rapid.SampledFrom(config.ValidOutputModes()).Draw(t, "outputMode")
		outputFormat := rapid.SampledFrom(config.ValidOutputFormats()).Draw(t, "outputFormat")
		maxLines := rapid.IntRange(1, 10000).Draw(t, "maxLines")

		m := Model{
			InputPath:    inputDir,
			OutputDir:    outputDir,
			OutputMode:   outputMode,
			OutputFormat: outputFormat,
			MaxLines:     strconv.Itoa(maxLines),
			CollectTest:  collectTest,
			Minify:       minify,
			Edges:        edges,
			Clean:        clean,
			Workers:      strconv.Itoa(workers),
		}

		cfg := m.ToParseConfig()

		if cfg.InputPath != inputDir {
			t.Fatalf("InputPath: got %q, want %q", cfg.InputPath, inputDir)
		}
		if cfg.OutputDir != outputDir {
			t.Fatalf("OutputDir: got %q, want %q", cfg.OutputDir, outputDir)
		}
		if cfg.OutputMode != outputMode {
			t.Fatalf("OutputMode: got %q, want %q", cfg.OutputMode, outputMode)
		}
		if cfg.OutputFormat != outputFormat {
			t.Fatalf("OutputFormat: got %q, want %q", cfg.OutputFormat, outputFormat)
		}
		if cfg.MaxLines != maxLines {
			t.Fatalf("MaxLines: got %d, want %d", cfg.MaxLines, maxLines)
		}
		if cfg.CollectTest != collectTest {
			t.Fatalf("CollectTest: got %v, want %v", cfg.CollectTest, collectTest)
		}
		if cfg.Minify != minify {
			t.Fatalf("Minify: got %v, want %v", cfg.Minify, minify)
		}
		if cfg.Edges != edges {
			t.Fatalf("Edges: got %v, want %v", cfg.Edges, edges)
		}
		if cfg.Clean != clean {
			t.Fatalf("Clean: got %v, want %v", cfg.Clean, clean)
		}
		if cfg.Workers != workers {
			t.Fatalf("Workers: got %d, want %d", cfg.Workers, workers)
		}
	})
}

// TestInlineModeSkipsOutputDir verifies that validate() does not require
// OutputDir when OutputMode is inline, and accepts empty OutputDir for
// directory modes (ParseConfig.Validate will default it to ./skeleton).
func TestInlineModeSkipsOutputDir(t *testing.T) {
	tests := []struct {
		name      string
		mode      config.OutputMode
		outputDir string
		wantErr   bool
	}{
		{
			name:      "inline mode without output dir is valid",
			mode:      config.OutputInline,
			outputDir: "",
			wantErr:   false,
		},
		{
			name:      "directory-flat without output dir is valid",
			mode:      config.OutputDirectoryFlat,
			outputDir: "",
			wantErr:   false,
		},
		{
			name:      "directory-tree without output dir is valid",
			mode:      config.OutputDirectoryTree,
			outputDir: "",
			wantErr:   false,
		},
		{
			name:      "directory-flat with output dir is valid",
			mode:      config.OutputDirectoryFlat,
			outputDir: "/tmp/out",
			wantErr:   false,
		},
		{
			name:      "directory-tree with output dir is valid",
			mode:      config.OutputDirectoryTree,
			outputDir: "/tmp/out",
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{
				InputPath:  "/some/path",
				OutputMode: tc.mode,
				OutputDir:  tc.outputDir,
			}
			errMsg := m.validate()
			if tc.wantErr && errMsg == "" {
				t.Fatalf("expected validation error for mode=%s outputDir=%q, got none", tc.mode, tc.outputDir)
			}
			if !tc.wantErr && errMsg != "" {
				t.Fatalf("unexpected validation error for mode=%s outputDir=%q: %s", tc.mode, tc.outputDir, errMsg)
			}
		})
	}
}

// TestNewModelDefaults verifies that NewModel() returns sensible defaults.
func TestNewModelDefaults(t *testing.T) {
	m := NewModel()

	if m.OutputMode != config.DefaultParseOutputMode {
		t.Fatalf("OutputMode default: got %q, want %q", m.OutputMode, config.DefaultParseOutputMode)
	}
	if m.OutputFormat != config.DefaultParseOutputFormat {
		t.Fatalf("OutputFormat default: got %q, want %q", m.OutputFormat, config.DefaultParseOutputFormat)
	}
	if m.OutputDir != config.DefaultParseOutputDir {
		t.Fatalf("OutputDir default: got %q, want %q", m.OutputDir, config.DefaultParseOutputDir)
	}
	if m.MaxLines != strconv.Itoa(config.DefaultParseMaxLines) {
		t.Fatalf("MaxLines default: got %q, want %d", m.MaxLines, config.DefaultParseMaxLines)
	}
	if m.Workers != strconv.Itoa(config.DefaultWorkers) {
		t.Fatalf("Workers default: got %q, want %d", m.Workers, config.DefaultWorkers)
	}
	if m.GraphOutput != config.DefaultGraphOutput {
		t.Fatalf("GraphOutput default: got %q, want %q", m.GraphOutput, config.DefaultGraphOutput)
	}
	if m.AnalysisOutput != config.DefaultAnalyzeOutput {
		t.Fatalf("AnalysisOutput default: got %q, want %q", m.AnalysisOutput, config.DefaultAnalyzeOutput)
	}
	if m.FanThreshold != strconv.Itoa(config.DefaultAnalyzeFanThreshold) {
		t.Fatalf("FanThreshold default: got %q, want %d", m.FanThreshold, config.DefaultAnalyzeFanThreshold)
	}
	if m.GodThreshold != strconv.Itoa(config.DefaultAnalyzeGodThreshold) {
		t.Fatalf("GodThreshold default: got %q, want %d", m.GodThreshold, config.DefaultAnalyzeGodThreshold)
	}
	if m.MaxInheritanceDepth != strconv.Itoa(config.DefaultAnalyzeMaxInheritanceDepth) {
		t.Fatalf("MaxInheritanceDepth default: got %q, want %d", m.MaxInheritanceDepth, config.DefaultAnalyzeMaxInheritanceDepth)
	}
	if m.TopN != strconv.Itoa(config.DefaultAnalyzeTopN) {
		t.Fatalf("TopN default: got %q, want %d", m.TopN, config.DefaultAnalyzeTopN)
	}
	if m.BetweennessThreshold != strconv.FormatFloat(config.DefaultAnalyzeBetweennessThreshold, 'g', -1, 64) {
		t.Fatalf("BetweennessThreshold default: got %q, want %g", m.BetweennessThreshold, config.DefaultAnalyzeBetweennessThreshold)
	}
	if m.PropagationCutoff != strconv.FormatFloat(config.DefaultAnalyzePropagationCutoff, 'g', -1, 64) {
		t.Fatalf("PropagationCutoff default: got %q, want %g", m.PropagationCutoff, config.DefaultAnalyzePropagationCutoff)
	}
	if m.FingerprintOutput != config.DefaultFingerprintOutput {
		t.Fatalf("FingerprintOutput default: got %q, want %q", m.FingerprintOutput, config.DefaultFingerprintOutput)
	}
	if m.FingerprintMinSim != strconv.Itoa(config.DefaultFingerprintMinSimilarity) {
		t.Fatalf("FingerprintMinSim default: got %q, want %d", m.FingerprintMinSim, config.DefaultFingerprintMinSimilarity)
	}
	if m.FingerprintMaxSim != strconv.Itoa(config.DefaultFingerprintMaxSimilarity) {
		t.Fatalf("FingerprintMaxSim default: got %q, want %d", m.FingerprintMaxSim, config.DefaultFingerprintMaxSimilarity)
	}
	if m.FingerprintModel != config.DefaultFingerprintModelOverride {
		t.Fatalf("FingerprintModel default: got %q, want %q", m.FingerprintModel, config.DefaultFingerprintModelOverride)
	}
	if m.FingerprintShowAll != config.DefaultFingerprintShowAll {
		t.Fatalf("FingerprintShowAll default: got %t, want %t", m.FingerprintShowAll, config.DefaultFingerprintShowAll)
	}
	if m.FingerprintRerank != config.DefaultFingerprintRerank {
		t.Fatalf("FingerprintRerank default: got %t, want %t", m.FingerprintRerank, config.DefaultFingerprintRerank)
	}
	if m.CollectTest != config.DefaultCollectTest {
		t.Fatalf("CollectTest default: got %t, want %t", m.CollectTest, config.DefaultCollectTest)
	}
	if m.Minify != config.DefaultParseMinify || m.Edges != config.DefaultParseEdges || m.Clean != config.DefaultParseClean {
		t.Fatal("parse boolean defaults do not match config defaults")
	}
}
