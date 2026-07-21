// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strconv"
	"testing"

	"codeknit/internal/config"
	"codeknit/internal/console"

	"github.com/spf13/cobra"
)

func TestCommandFlagDefaultsMatchConfig(t *testing.T) {
	con := console.New()
	tests := []struct {
		newCmd func(*console.Console) *cobra.Command
		flags  map[string]string
		name   string
	}{
		{
			name:   "parse",
			newCmd: newParseCmd,
			flags: map[string]string{
				"output-mode":  string(config.DefaultParseOutputMode),
				"format":       string(config.DefaultParseOutputFormat),
				"max-lines":    strconv.Itoa(config.DefaultParseMaxLines),
				"collect-test": strconv.FormatBool(config.DefaultCollectTest),
				"minify":       strconv.FormatBool(config.DefaultParseMinify),
				"edges":        strconv.FormatBool(config.DefaultParseEdges),
				"clean":        strconv.FormatBool(config.DefaultParseClean),
				"workers":      strconv.Itoa(config.DefaultWorkers),
				"verbose":      strconv.FormatBool(config.DefaultVerbose),
			},
		},
		{
			name:   "graph show",
			newCmd: newGraphShowCmd,
			flags: map[string]string{
				"output":       config.DefaultGraphOutput,
				"collect-test": strconv.FormatBool(config.DefaultCollectTest),
				"workers":      strconv.Itoa(config.DefaultWorkers),
				"verbose":      strconv.FormatBool(config.DefaultVerbose),
			},
		},
		{
			name:   "graph analyze",
			newCmd: newGraphAnalyzeCmd,
			flags: map[string]string{
				"output":                config.DefaultAnalyzeOutput,
				"collect-test":          strconv.FormatBool(config.DefaultCollectTest),
				"workers":               strconv.Itoa(config.DefaultWorkers),
				"verbose":               strconv.FormatBool(config.DefaultVerbose),
				"fan-threshold":         strconv.Itoa(config.DefaultAnalyzeFanThreshold),
				"god-threshold":         strconv.Itoa(config.DefaultAnalyzeGodThreshold),
				"max-inheritance-depth": strconv.Itoa(config.DefaultAnalyzeMaxInheritanceDepth),
				"top-n":                 strconv.Itoa(config.DefaultAnalyzeTopN),
				"betweenness-threshold": fmt.Sprint(config.DefaultAnalyzeBetweennessThreshold),
				"propagation-cutoff":    fmt.Sprint(config.DefaultAnalyzePropagationCutoff),
			},
		},
		{
			name:   "graph hotspots",
			newCmd: newGraphHotspotsCmd,
			flags: map[string]string{
				"output":               config.DefaultHotspotOutput,
				"format":               string(config.DefaultHotspotFormat),
				"since":                config.DefaultHotspotSince,
				"max-commits":          strconv.Itoa(config.DefaultHotspotMaxCommits),
				"max-files-per-commit": strconv.Itoa(config.DefaultHotspotMaxFilesPerCommit),
				"min-cochanges":        strconv.Itoa(config.DefaultHotspotMinCoChanges),
				"top-n":                strconv.Itoa(config.DefaultHotspotTopN),
				"include-merges":       strconv.FormatBool(config.DefaultHotspotIncludeMerges),
				"collect-test":         strconv.FormatBool(config.DefaultCollectTest),
				"workers":              strconv.Itoa(config.DefaultWorkers),
				"verbose":              strconv.FormatBool(config.DefaultVerbose),
			},
		},
		{
			name:   "fingerprint",
			newCmd: newFingerprintCmd,
			flags: map[string]string{
				"output":         config.DefaultFingerprintOutput,
				"min-similarity": strconv.Itoa(config.DefaultFingerprintMinSimilarity),
				"max-similarity": strconv.Itoa(config.DefaultFingerprintMaxSimilarity),
				"show-all":       strconv.FormatBool(config.DefaultFingerprintShowAll),
				"rerank":         strconv.FormatBool(config.DefaultFingerprintRerank),
				"model":          config.DefaultFingerprintModelOverride,
				"collect-test":   strconv.FormatBool(config.DefaultCollectTest),
				"workers":        strconv.Itoa(config.DefaultWorkers),
				"verbose":        strconv.FormatBool(config.DefaultVerbose),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.newCmd(con)
			for name, want := range tc.flags {
				flag := cmd.Flags().Lookup(name)
				if flag == nil {
					t.Fatalf("flag %q not found", name)
				}
				if flag.DefValue != want {
					t.Errorf("%s default: got %q, want %q", name, flag.DefValue, want)
				}
			}
		})
	}
}
