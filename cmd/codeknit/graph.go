// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"codeknit/internal/config"
	"codeknit/internal/console"

	"github.com/spf13/cobra"
)

func newGraphCmd(con *console.Console) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Graph commands for codebase structure visualization and analysis",
		Long: `Analyze source files and either visualize the codebase as an interactive
HTML graph or run graph algorithms to detect code quality issues.

Use "graph show" to generate an interactive HTML visualization.
Use "graph analyze" to run structural analysis algorithms.`,
	}

	cmd.AddCommand(newGraphShowCmd(con))
	cmd.AddCommand(newGraphAnalyzeCmd(con))

	return cmd
}

// graphShowOptions holds the flag-bound state for `graph show`.
type graphShowOptions struct {
	commonOptions
}

func newGraphShowCmd(con *console.Console) *cobra.Command {
	opts := &graphShowOptions{}

	cmd := &cobra.Command{
		Use:   "show <input-path>",
		Short: "Generate an interactive HTML graph of the codebase structure",
		Long: `Analyze source files under <input-path> and generate a self-contained
HTML file with an interactive graph visualization. The graph shows symbols
(functions, classes, types) as nodes and their relationships (calls, contains,
implements) as edges. Open the HTML file in any browser to explore.`,
		Example: `  # Generate graph with default output
  codeknit graph show ./myproject

  # Custom output file
  codeknit graph show ./myproject -o graph.html

  # Include test files
  codeknit graph show ./src --collect-test`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.GraphConfig{
				Common: config.Common{
					InputPath:   args[0],
					Workers:     opts.workers,
					CollectTest: opts.collectTest,
					Verbose:     opts.verbose,
				},
				Output: opts.output,
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			return runGraphShow(cfg, con)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", config.DefaultGraphOutput,
		"output HTML file path")
	cmd.Flags().BoolVar(&opts.collectTest, "collect-test", config.DefaultCollectTest,
		"include test files in analysis")
	cmd.Flags().IntVar(&opts.workers, "workers", config.DefaultWorkers,
		"max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", config.DefaultVerbose,
		"print progress information during processing")

	return cmd
}

// graphAnalyzeOptions holds the flag-bound state for `graph analyze`.
type graphAnalyzeOptions struct {
	commonOptions
	betweennessThreshold float64
	propagationCutoff    float64
	fanThreshold         int
	godThreshold         int
	maxInheritanceDepth  int
	topN                 int
}

func newGraphAnalyzeCmd(con *console.Console) *cobra.Command {
	opts := &graphAnalyzeOptions{}

	cmd := &cobra.Command{
		Use:   "analyze <input-path>",
		Short: "Run graph analysis algorithms and emit an LLM-readable report",
		Long: `Analyze source files under <input-path> and run structural graph
algorithms to detect code quality issues. The output is a .skt file
designed for LLM consumption.

Algorithms:
  - Cyclic dependencies (Tarjan's SCC)
  - Hub detection (high fan-in/fan-out coupling)
  - Orphan detection (dead code candidates)
  - God class/function detection (excessive children)
  - Instability metric (Robert C. Martin's Ce/(Ca+Ce))
  - Deep inheritance chains
  - Betweenness centrality (bottleneck detection)
  - Articulation points (single points of failure)
  - PageRank (recursive importance)
  - Transitive fan-in (blast radius)
  - Change propagation simulation
  - Circular package dependencies
  - Layer violation detection
  - Reachability from entry points
  - Weakly connected components
  - Dependency weight (package coupling strength)
  - Distance from Main Sequence (A+I balance)`,
		Example: `  # Analyze with defaults
  codeknit graph analyze ./myproject

  # Custom output and thresholds
  codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 8

  # Show more results per section
  codeknit graph analyze ./myproject --top-n 50

  # Include test files
  codeknit graph analyze ./src --collect-test`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.AnalyzeConfig{
				Common: config.Common{
					InputPath:   args[0],
					Workers:     opts.workers,
					CollectTest: opts.collectTest,
					Verbose:     opts.verbose,
				},
				Output:               opts.output,
				FanThreshold:         opts.fanThreshold,
				GodThreshold:         opts.godThreshold,
				MaxInheritanceDepth:  opts.maxInheritanceDepth,
				TopN:                 opts.topN,
				BetweennessThreshold: opts.betweennessThreshold,
				PropagationCutoff:    opts.propagationCutoff,
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			return runGraphAnalyze(cfg, con)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", config.DefaultAnalyzeOutput,
		"output .skt file path")
	cmd.Flags().BoolVar(&opts.collectTest, "collect-test", config.DefaultCollectTest,
		"include test files in analysis")
	cmd.Flags().IntVar(&opts.workers, "workers", config.DefaultWorkers,
		"max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", config.DefaultVerbose,
		"print progress information during processing")
	cmd.Flags().IntVar(&opts.fanThreshold, "fan-threshold", config.DefaultAnalyzeFanThreshold,
		"minimum fan-in or fan-out to flag a hub symbol")
	cmd.Flags().IntVar(&opts.godThreshold, "god-threshold", config.DefaultAnalyzeGodThreshold,
		"minimum contains-edge count to flag a god class/function")
	cmd.Flags().IntVar(&opts.maxInheritanceDepth, "max-inheritance-depth", config.DefaultAnalyzeMaxInheritanceDepth,
		"flag inheritance chains deeper than this")
	cmd.Flags().IntVar(&opts.topN, "top-n", config.DefaultAnalyzeTopN,
		"cap ranked output sections (betweenness, pagerank, etc.); 0 = no limit")
	cmd.Flags().Float64Var(&opts.betweennessThreshold, "betweenness-threshold", config.DefaultAnalyzeBetweennessThreshold,
		"minimum betweenness centrality value to report")
	cmd.Flags().Float64Var(&opts.propagationCutoff, "propagation-cutoff", config.DefaultAnalyzePropagationCutoff,
		"minimum probability to continue change propagation simulation")

	return cmd
}
