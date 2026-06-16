// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"codeknit/internal/config"
	"codeknit/internal/console"

	"github.com/spf13/cobra"
)

// parseOptions holds the flag-bound state for the `parse` subcommand.
// It mirrors the docker CLI pattern: one lowercase options struct per command.
type parseOptions struct {
	outputMode  string
	maxLines    int
	workers     int
	collectTest bool
	verbose     bool
	minify      bool
	edges       bool
	clean       bool
}

func newParseCmd(con *console.Console) *cobra.Command {
	opts := &parseOptions{}

	cmd := &cobra.Command{
		Use:   "parse <input-path> [output-dir]",
		Short: "Parse source code and extract structural information",
		Long: `Parse source files under <input-path> and emit a compact structural
representation. The output-dir defaults to ./skeleton for directory-flat
and directory-tree modes. For inline mode output is written to stdout.`,
		Example: `  # Flat directory output (default, writes to ./skeleton)
  codeknit parse ./myproject

  # Custom output directory
  codeknit parse ./myproject ./output

  # Tree-mirroring directory output
  codeknit parse ./myproject ./output --output-mode directory-tree

  # Inline output to stdout
  codeknit parse ./myproject --output-mode inline

  # Include test files and minify
  codeknit parse ./src --collect-test --minify

  # Limit output file size and parallelism
  codeknit parse ./src --max-lines 500 --workers 4`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var outputDir string
			if len(args) >= 2 {
				outputDir = args[1]
			}

			cfg := &config.ParseConfig{
				Common: config.Common{
					InputPath:   args[0],
					Workers:     opts.workers,
					CollectTest: opts.collectTest,
					Verbose:     opts.verbose,
				},
				OutputDir:  outputDir,
				OutputMode: config.OutputMode(opts.outputMode),
				MaxLines:   opts.maxLines,
				Minify:     opts.minify,
				Edges:      opts.edges,
				Clean:      opts.clean,
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			return runParse(cfg, con)
		},
	}

	cmd.Flags().StringVar(&opts.outputMode, "output-mode", "directory-flat",
		"output mode: inline, directory-flat, directory-tree")
	cmd.Flags().IntVar(&opts.maxLines, "max-lines", 500,
		"maximum lines per output file")
	cmd.Flags().BoolVar(&opts.collectTest, "collect-test", false,
		"include test files in analysis")
	cmd.Flags().BoolVar(&opts.minify, "minify", false,
		"enable dictionary-based output minification")
	cmd.Flags().BoolVar(&opts.edges, "edges", false,
		"include the [edges] section in output (off by default)")
	cmd.Flags().BoolVar(&opts.clean, "clean", false,
		"remove stale .skt files from the output directory before writing")
	cmd.Flags().IntVar(&opts.workers, "workers", 0,
		"max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false,
		"print progress information during processing")

	return cmd
}
