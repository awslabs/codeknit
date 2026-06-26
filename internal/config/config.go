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
		c.OutputMode = OutputDirectoryFlat
	}
	if !c.OutputMode.IsValid() {
		return fmt.Errorf("invalid output mode %s: must be one of inline, directory-flat, directory-tree", c.OutputMode)
	}

	if c.OutputFormat == "" {
		c.OutputFormat = OutputFormatSKT
	}
	if !c.OutputFormat.IsValid() {
		return fmt.Errorf("invalid output format %s: must be one of skt, json", c.OutputFormat)
	}

	if c.OutputMode == OutputDirectoryFlat || c.OutputMode == OutputDirectoryTree {
		if c.OutputDir == "" {
			c.OutputDir = "./skeleton"
		}
		if err := os.MkdirAll(c.OutputDir, 0o700); err != nil { //nolint:gosec // 0o700 is least-privilege for directories (execute bit needed for traversal)
			return fmt.Errorf("failed to create output directory: %s: %w", c.OutputDir, err)
		}
	}

	if c.MaxLines == 0 {
		c.MaxLines = 500
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
		c.Output = "./skeleton/codeknit-graph.html"
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

// Validate checks AnalyzeConfig and fills defaults.
func (c *AnalyzeConfig) Validate() error {
	if err := c.Common.Validate(); err != nil {
		return err
	}
	if c.Output == "" {
		c.Output = "./skeleton/graph_analysis.skt"
	}
	if c.FanThreshold <= 0 {
		c.FanThreshold = 10
	}
	if c.GodThreshold <= 0 {
		c.GodThreshold = 15
	}
	if c.MaxInheritanceDepth <= 0 {
		c.MaxInheritanceDepth = 5
	}
	if c.TopN <= 0 {
		c.TopN = 30
	}
	if c.BetweennessThreshold <= 0 {
		c.BetweennessThreshold = 0.001
	}
	if c.PropagationCutoff <= 0 {
		c.PropagationCutoff = 0.05
	}
	return nil
}

// FingerprintConfig holds options for the `fingerprint` subcommand.
type FingerprintConfig struct {
	Output     string
	EmbedModel string // Ollama model for semantic reranking via RRF; "" disables
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
		c.Output = "./skeleton/fingerprints.skt"
	}
	if c.MinSim <= 0 {
		c.MinSim = 75
	}
	if c.MaxSim <= 0 {
		c.MaxSim = 100
	}
	return nil
}
