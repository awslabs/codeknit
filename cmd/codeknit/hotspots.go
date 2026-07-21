// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"codeknit/internal/config"
	"codeknit/internal/console"

	"github.com/spf13/cobra"
)

type graphHotspotOptions struct {
	output            string
	format            string
	since             string
	workers           int
	maxCommits        int
	maxFilesPerCommit int
	minCoChanges      int
	topN              int
	collectTest       bool
	verbose           bool
	includeMerges     bool
}

func newGraphHotspotsCmd(con *console.Console) *cobra.Command {
	opts := &graphHotspotOptions{}

	cmd := &cobra.Command{
		Use:   "hotspots <input-path>",
		Short: "Rank change hotspots using Git history and graph structure",
		Long: `Analyze Git history and the current source graph to rank files that are
both frequently changed and structurally important. The report also identifies
temporal coupling between files that repeatedly change in the same commits.`,
		Example: `  # Analyze the last 12 months
  codeknit graph hotspots ./myproject

  # Analyze two years and emit JSON
  codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

  # Include larger commits and show more results
  codeknit graph hotspots . --max-files-per-commit 100 --top-n 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.HotspotConfig{
				Common: config.Common{
					InputPath:   args[0],
					Workers:     opts.workers,
					CollectTest: opts.collectTest,
					Verbose:     opts.verbose,
				},
				Output:            opts.output,
				Format:            config.OutputFormat(opts.format),
				Since:             opts.since,
				MaxCommits:        opts.maxCommits,
				MaxFilesPerCommit: opts.maxFilesPerCommit,
				MinCoChanges:      opts.minCoChanges,
				TopN:              opts.topN,
				IncludeMerges:     opts.includeMerges,
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			return runGraphHotspots(cfg, con)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", config.DefaultHotspotOutput,
		"output file path")
	cmd.Flags().StringVar(&opts.format, "format", string(config.DefaultHotspotFormat),
		"output format: skt, json")
	cmd.Flags().StringVar(&opts.since, "since", config.DefaultHotspotSince,
		"history window: positive duration such as 180d, 12mo, or 2y")
	cmd.Flags().IntVar(&opts.maxCommits, "max-commits", config.DefaultHotspotMaxCommits,
		"maximum commits to inspect")
	cmd.Flags().IntVar(&opts.maxFilesPerCommit, "max-files-per-commit", config.DefaultHotspotMaxFilesPerCommit,
		"exclude bulk commits changing more than this many files")
	cmd.Flags().IntVar(&opts.minCoChanges, "min-cochanges", config.DefaultHotspotMinCoChanges,
		"minimum shared commits required for temporal coupling")
	cmd.Flags().IntVar(&opts.topN, "top-n", config.DefaultHotspotTopN,
		"maximum hotspots and coupling pairs to report")
	cmd.Flags().BoolVar(&opts.includeMerges, "include-merges", config.DefaultHotspotIncludeMerges,
		"include merge commits in history metrics")
	cmd.Flags().BoolVar(&opts.collectTest, "collect-test", config.DefaultCollectTest,
		"include test files in analysis")
	cmd.Flags().IntVar(&opts.workers, "workers", config.DefaultWorkers,
		"max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", config.DefaultVerbose,
		"print progress information during processing")

	return cmd
}
