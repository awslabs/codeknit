// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeknit/internal/config"
	"codeknit/internal/hotspot"
)

// HotspotOptions controls hotspot result serialization.
type HotspotOptions struct {
	OutputPath string
	Format     config.OutputFormat
}

// EmitHotspots writes hotspot analysis as SKT or JSON.
func (e *Emitter) EmitHotspots(result *hotspot.Result, opts *HotspotOptions) error {
	if dir := filepath.Dir(opts.OutputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	var content []byte
	var err error
	if opts.Format == config.OutputFormatJSON {
		content, err = json.MarshalIndent(result, "", "  ")
		if err == nil {
			content = append(content, '\n')
		}
	} else {
		content = []byte(renderHotspots(result))
	}
	if err != nil {
		return fmt.Errorf("encode hotspots: %w", err)
	}
	if err := os.WriteFile(opts.OutputPath, content, 0o600); err != nil {
		return fmt.Errorf("write hotspots: %w", err)
	}
	return nil
}

func renderHotspots(result *hotspot.Result) string {
	var b strings.Builder
	b.WriteString("[history_hotspots]\n")
	fmt.Fprintf(&b, "generated_at: %s\n", result.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(&b, "since: %s\n", result.Since.UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "confidence: %s\n", result.Confidence)
	fmt.Fprintf(&b, "commits_visited: %d | commits_analyzed: %d | skipped_merges: %d | skipped_bulk: %d\n\n",
		result.CommitsVisited, result.CommitsAnalyzed, result.SkippedMerges, result.SkippedBulkCommits)

	b.WriteString("[hotspots]\n")
	if len(result.Hotspots) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, entry := range result.Hotspots {
			fmt.Fprintf(&b,
				"score=%.3f history=%.3f structure=%.3f commits=%d churn=%d transitive_fan_in=%d pagerank=%.6f betweenness=%.6f last_changed=%s | %s\n",
				entry.Score, entry.HistoryScore, entry.StructureScore, entry.Commits, entry.Churn,
				entry.TransitiveIn, entry.PageRank, entry.Betweenness,
				entry.LastChanged.UTC().Format("2006-01-02"), entry.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[temporal_coupling]\n")
	if len(result.TemporalCoupling) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, coupling := range result.TemporalCoupling {
			fmt.Fprintf(&b, "strength=%.3f cochanges=%d %s <-> %s\n",
				coupling.Strength, coupling.CoChanges, coupling.Left, coupling.Right)
		}
	}
	return b.String()
}
