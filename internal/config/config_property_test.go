// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"pgregory.net/rapid"
)

// Feature: cli-output-modes, Property 2: Non-existent path validation
// Validates: Requirement 1.4
func TestProperty_NonExistentPathValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		base := rapid.StringMatching(`[a-zA-Z0-9_]{3,20}`).Draw(t, "segment")
		path := filepath.Join(os.TempDir(), "codeknit_nonexistent", base, "does_not_exist")

		if _, err := os.Stat(path); err == nil {
			t.Skip("generated path unexpectedly exists")
		}

		cfg := ParseConfig{
			Common:    Common{InputPath: path},
			OutputDir: os.TempDir(),
		}
		err := cfg.Validate()
		if err == nil {
			t.Fatalf("expected error for non-existent path %q, got nil", path)
		}
	})
}

// Feature: cli-output-modes, Property 3: OutputMode validation
func TestProperty_OutputModeValidation(t *testing.T) {
	validModes := map[OutputMode]bool{
		OutputInline:        true,
		OutputDirectoryFlat: true,
		OutputDirectoryTree: true,
	}

	rapid.Check(t, func(t *rapid.T) {
		s := rapid.StringMatching(`[a-zA-Z0-9_-]{0,30}`).Draw(t, "mode")
		mode := OutputMode(s)
		got := mode.IsValid()
		want := validModes[mode]
		if got != want {
			t.Fatalf("IsValid(%q) = %v, want %v", s, got, want)
		}
	})
}

// Feature: cli-output-modes, Property 4: Directory modes default output directory
// Validates: Requirement 2.6
func TestProperty_DirectoryModesDefaultOutputDir(t *testing.T) {
	dirModes := []OutputMode{OutputDirectoryFlat, OutputDirectoryTree}

	tmpDir := t.TempDir()

	rapid.Check(t, func(t *rapid.T) {
		mode := dirModes[rapid.IntRange(0, len(dirModes)-1).Draw(t, "modeIdx")]

		cfg := ParseConfig{
			Common:     Common{InputPath: tmpDir},
			OutputDir:  "", // intentionally empty — should default to ./skeleton
			OutputMode: mode,
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error for empty OutputDir with mode %q: %v", mode, err)
		}
		if cfg.OutputDir != "./skeleton" {
			t.Fatalf("expected OutputDir to default to %q, got %q", "./skeleton", cfg.OutputDir)
		}
	})
}

// Feature: cli-output-modes, Property 10: MaxLines validation
// Validates: Requirement 7.4
func TestProperty_MaxLinesValidation(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := t.TempDir()

	rapid.Check(t, func(t *rapid.T) {
		// Generate integers < 1 (negative values). Zero is excluded because
		// Validate() defaults zero to 500.
		maxLines := rapid.IntRange(-1000, -1).Draw(t, "maxLines")

		cfg := ParseConfig{
			Common:    Common{InputPath: tmpDir},
			OutputDir: outDir,
			MaxLines:  maxLines,
		}
		if err := cfg.Validate(); err == nil {
			t.Fatalf("expected error for MaxLines=%d, got nil", maxLines)
		}
	})
}
