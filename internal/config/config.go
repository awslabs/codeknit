// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config provides per-command CLI configuration structs and validation.
//
// Each subcommand has its own config type that embeds [Common] for the
// flags shared by every command. This mirrors the per-command "options struct"
// pattern used by docker, kubectl, gh, and hugo.
package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// OutputMode controls how the parse emitter writes results.
type OutputMode string

// OutputFormat controls the serialization format for parse output.
type OutputFormat string

// Valid output mode constants.
const (
	OutputInline        OutputMode = "inline"
	OutputDirectoryFlat OutputMode = "directory-flat"
	OutputDirectoryTree OutputMode = "directory-tree"
)

// Valid output format constants.
const (
	OutputFormatSKT  OutputFormat = "skt"
	OutputFormatJSON OutputFormat = "json"
)

// Shared command defaults. CLI flags, the TUI, validators, and emitters must
// all consume these values rather than defining their own copies.
const (
	DefaultWorkers     = 0
	DefaultCollectTest = false
	DefaultVerbose     = false

	DefaultParseOutputMode   = OutputDirectoryFlat
	DefaultParseOutputFormat = OutputFormatSKT
	DefaultParseOutputDir    = "./skeleton"
	DefaultParseMaxLines     = 500
	DefaultParseMinify       = false
	DefaultParseEdges        = false
	DefaultParseClean        = false

	DefaultGraphOutput = "./skeleton/codeknit-graph.html"

	DefaultAnalyzeOutput               = "./skeleton/graph_analysis.skt"
	DefaultAnalyzeFanThreshold         = 10
	DefaultAnalyzeGodThreshold         = 15
	DefaultAnalyzeMaxInheritanceDepth  = 5
	DefaultAnalyzeTopN                 = 30
	DefaultAnalyzeBetweennessThreshold = 0.001
	DefaultAnalyzePropagationCutoff    = 0.05

	DefaultHotspotOutput            = "./skeleton/hotspots.skt"
	DefaultHotspotFormat            = OutputFormatSKT
	DefaultHotspotSince             = "12mo"
	DefaultHotspotMaxCommits        = 2000
	DefaultHotspotMaxFilesPerCommit = 50
	DefaultHotspotMinCoChanges      = 3
	DefaultHotspotTopN              = 30
	DefaultHotspotIncludeMerges     = false

	DefaultFingerprintOutput        = "./skeleton/fingerprints.skt"
	DefaultFingerprintMinSimilarity = 65
	DefaultFingerprintMaxSimilarity = 95
	DefaultFingerprintShowAll       = false
	DefaultFingerprintRerank        = false
	DefaultFingerprintModelOverride = ""
	DefaultFingerprintEmbedModel    = "qwen3-embedding:0.6b"
)

// ValidOutputModes returns the set of valid output mode values.
func ValidOutputModes() []OutputMode {
	return []OutputMode{OutputInline, OutputDirectoryFlat, OutputDirectoryTree}
}

// ValidOutputFormats returns the set of valid output format values.
func ValidOutputFormats() []OutputFormat {
	return []OutputFormat{OutputFormatSKT, OutputFormatJSON}
}

// IsValid reports whether m is a recognized output mode.
func (m OutputMode) IsValid() bool {
	switch m {
	case OutputInline, OutputDirectoryFlat, OutputDirectoryTree:
		return true
	}
	return false
}

// IsValid reports whether f is a recognized output format.
func (f OutputFormat) IsValid() bool {
	switch f {
	case OutputFormatSKT, OutputFormatJSON:
		return true
	}
	return false
}

// Common holds the flags that every codeknit subcommand shares.
//
// It is embedded in each per-command config so that shared validation
// (input-path existence, worker defaulting) lives in one place.
type Common struct {
	InputPath   string
	Workers     int
	CollectTest bool
	Verbose     bool
}

// Validate checks that the shared fields are valid and fills defaults.
func (c *Common) Validate() error {
	if _, err := os.Stat(c.InputPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input path does not exist: %s", c.InputPath)
		}
		return fmt.Errorf("cannot access input path: %s: %w", c.InputPath, err)
	}
	if c.Workers <= 0 {
		c.Workers = runtime.NumCPU()
	}
	return nil
}

// ParseConfig holds options for the `parse` subcommand.
type ParseConfig struct {
	OutputDir    string
	OutputMode   OutputMode
	OutputFormat OutputFormat
	Common
	MaxLines int
	Minify   bool
	Edges    bool
	Clean    bool
}

// Validate checks ParseConfig and fills defaults.
func (c *ParseConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}

	if c.OutputMode == "" {
		c.OutputMode = DefaultParseOutputMode
	}
	if !c.OutputMode.IsValid() {
		return fmt.Errorf("invalid output mode %s: must be one of inline, directory-flat, directory-tree", c.OutputMode)
	}

	if c.OutputFormat == "" {
		c.OutputFormat = DefaultParseOutputFormat
	}
	if !c.OutputFormat.IsValid() {
		return fmt.Errorf("invalid output format %s: must be one of skt, json", c.OutputFormat)
	}

	if c.OutputMode == OutputDirectoryFlat || c.OutputMode == OutputDirectoryTree {
		if c.OutputDir == "" {
			c.OutputDir = DefaultParseOutputDir
		}
		if err := os.MkdirAll(c.OutputDir, 0o700); err != nil { //nolint:gosec // 0o700 is least-privilege for directories (execute bit needed for traversal)
			return fmt.Errorf("failed to create output directory: %s: %w", c.OutputDir, err)
		}
	}

	if c.MaxLines == 0 {
		c.MaxLines = DefaultParseMaxLines
	}
	if c.MaxLines < 1 {
		return fmt.Errorf("max-lines must be at least 1")
	}

	return nil
}

// GraphConfig holds options for the `graph show` subcommand.
type GraphConfig struct {
	Output string
	Common
}

// Validate checks GraphConfig and fills defaults.
func (c *GraphConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}
	if c.Output == "" {
		c.Output = DefaultGraphOutput
	}
	return nil
}

// AnalyzeConfig holds options for the `graph analyze` subcommand.
type AnalyzeConfig struct {
	Output string
	Common
	BetweennessThreshold float64
	PropagationCutoff    float64
	FanThreshold         int
	GodThreshold         int
	MaxInheritanceDepth  int
	TopN                 int
}

// HotspotConfig holds options for the `graph hotspots` subcommand.
type HotspotConfig struct {
	Output string
	Format OutputFormat
	Since  string
	Common
	MaxCommits        int
	MaxFilesPerCommit int
	MinCoChanges      int
	TopN              int
	IncludeMerges     bool
}

// Validate checks HotspotConfig and fills defaults.
func (c *HotspotConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}
	if c.Output == "" {
		c.Output = DefaultHotspotOutput
	}
	if c.Format == "" {
		c.Format = DefaultHotspotFormat
	}
	if !c.Format.IsValid() {
		return fmt.Errorf("invalid output format %s: must be one of skt, json", c.Format)
	}
	if c.Since == "" {
		c.Since = DefaultHotspotSince
	}
	if _, err := ParseLookback(c.Since); err != nil {
		return err
	}
	if c.MaxCommits == 0 {
		c.MaxCommits = DefaultHotspotMaxCommits
	}
	if c.MaxCommits < 1 {
		return fmt.Errorf("max-commits must be at least 1")
	}
	if c.MaxFilesPerCommit == 0 {
		c.MaxFilesPerCommit = DefaultHotspotMaxFilesPerCommit
	}
	if c.MaxFilesPerCommit < 1 {
		return fmt.Errorf("max-files-per-commit must be at least 1")
	}
	if c.MinCoChanges == 0 {
		c.MinCoChanges = DefaultHotspotMinCoChanges
	}
	if c.MinCoChanges < 1 {
		return fmt.Errorf("min-cochanges must be at least 1")
	}
	if c.TopN == 0 {
		c.TopN = DefaultHotspotTopN
	}
	if c.TopN < 1 {
		return fmt.Errorf("top-n must be at least 1")
	}
	return nil
}

// ParseLookback parses history windows such as 180d, 12mo, and 2y.
func ParseLookback(value string) (time.Duration, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	units := []struct {
		suffix string
		days   int
	}{
		{suffix: "mo", days: 30},
		{suffix: "y", days: 365},
		{suffix: "w", days: 7},
		{suffix: "d", days: 1},
	}
	for _, unit := range units {
		if !strings.HasSuffix(value, unit.suffix) {
			continue
		}
		number := strings.TrimSuffix(value, unit.suffix)
		count, err := strconv.Atoi(number)
		if err != nil || count <= 0 {
			break
		}
		return time.Duration(count*unit.days) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("invalid since value %q: use a positive duration such as 180d, 12mo, or 2y", value)
}

// Validate checks AnalyzeConfig and fills defaults.
func (c *AnalyzeConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}
	if c.Output == "" {
		c.Output = DefaultAnalyzeOutput
	}
	if c.FanThreshold <= 0 {
		c.FanThreshold = DefaultAnalyzeFanThreshold
	}
	if c.GodThreshold <= 0 {
		c.GodThreshold = DefaultAnalyzeGodThreshold
	}
	if c.MaxInheritanceDepth <= 0 {
		c.MaxInheritanceDepth = DefaultAnalyzeMaxInheritanceDepth
	}
	if c.TopN <= 0 {
		c.TopN = DefaultAnalyzeTopN
	}
	if c.BetweennessThreshold <= 0 {
		c.BetweennessThreshold = DefaultAnalyzeBetweennessThreshold
	}
	if c.PropagationCutoff <= 0 {
		c.PropagationCutoff = DefaultAnalyzePropagationCutoff
	}
	return nil
}

// FingerprintConfig holds options for the `fingerprint` subcommand.
type FingerprintConfig struct {
	Output     string
	EmbedModel string // Ollama model for semantic retrieval and reranking; "" disables
	Common
	MinSim  int
	MaxSim  int
	ShowAll bool
}

// Validate checks FingerprintConfig and fills defaults.
func (c *FingerprintConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}
	if c.Output == "" {
		c.Output = DefaultFingerprintOutput
	}
	if c.MinSim <= 0 {
		c.MinSim = DefaultFingerprintMinSimilarity
	}
	if c.MaxSim <= 0 {
		c.MaxSim = DefaultFingerprintMaxSimilarity
	}
	return nil
}

// ResolveFingerprintEmbedModel returns the effective embedding model.
// A model override enables reranking even when rerank is false.
func ResolveFingerprintEmbedModel(rerank bool, modelOverride string) string {
	if modelOverride != "" {
		return modelOverride
	}
	if rerank {
		return DefaultFingerprintEmbedModel
	}
	return ""
}
