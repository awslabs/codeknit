// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"codeknit/internal/config"
	"codeknit/internal/console"
	"codeknit/internal/emitter"
	"codeknit/internal/pipeline"

	"github.com/spf13/cobra"
)

// fingerprintOptions holds the flag-bound state for `fingerprint`.
// Fields are ordered for optimal struct alignment (strings → float → ints → bools).
type fingerprintOptions struct {
	model string
	commonOptions
	minSim  int
	maxSim  int
	rerank  bool
	showAll bool
}

func newFingerprintCmd(con *console.Console) *cobra.Command {
	opts := &fingerprintOptions{}

	cmd := &cobra.Command{
		Use:   "fingerprint <input-path>",
		Short: "Detect duplicate and near-duplicate code using fuzzy hashing",
		Long: `Analyze source files under <input-path> and generate fuzzy fingerprints
for each function, method, variable, and type. Similar code produces similar
fingerprints, enabling detection of duplicated logic across the codebase —
even across different programming languages.

The fingerprint is computed from a normalized intermediate representation
that captures the semantic operations (assignments, calls, comparisons,
control flow) while ignoring variable names, string literals, and type
annotations.

Only duplicates within the similarity range [--min-similarity, --max-similarity]
are reported. Use --show-all to also include the raw fingerprint listing.`,
		Example: `  # Find near-duplicates (65-95% similarity, default)
  codeknit fingerprint ./myproject

  # Find only exact duplicates
  codeknit fingerprint ./myproject --min-similarity 100

  # Semantic reranking — filters false positives via Ollama embeddings
  # requires: ollama serve && ollama pull qwen3-embedding:0.6b
  codeknit fingerprint ./myproject --rerank

  # Semantic reranking with a different model
  codeknit fingerprint ./myproject --rerank --model qwen3-embedding:4b

  # Include raw fingerprint listing
  codeknit fingerprint ./myproject --show-all

  # Custom output file
  codeknit fingerprint ./myproject -o duplicates.skt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.FingerprintConfig{
				Common: config.Common{
					InputPath:   args[0],
					Workers:     opts.workers,
					CollectTest: opts.collectTest,
					Verbose:     opts.verbose,
				},
				Output:     opts.output,
				EmbedModel: config.ResolveFingerprintEmbedModel(opts.rerank, opts.model),
				MinSim:     opts.minSim,
				MaxSim:     opts.maxSim,
				ShowAll:    opts.showAll,
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			return runFingerprint(cfg, con)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", config.DefaultFingerprintOutput,
		"output file path")
	cmd.Flags().IntVar(&opts.minSim, "min-similarity", config.DefaultFingerprintMinSimilarity,
		"minimum similarity percentage to report (0-100)")
	cmd.Flags().IntVar(&opts.maxSim, "max-similarity", config.DefaultFingerprintMaxSimilarity,
		"maximum similarity percentage to report (0-100)")
	cmd.Flags().BoolVar(&opts.showAll, "show-all", config.DefaultFingerprintShowAll,
		"include the [fingerprints] section with raw token data")
	cmd.Flags().BoolVar(&opts.rerank, "rerank", config.DefaultFingerprintRerank,
		"rerank CTPH candidates with semantic embeddings via Ollama to eliminate\n"+
			"false positives (requires: ollama serve && ollama pull qwen3-embedding:0.6b)")
	cmd.Flags().StringVar(&opts.model, "model", config.DefaultFingerprintModelOverride,
		"Ollama embedding model to use with --rerank (default: qwen3-embedding:0.6b)")
	cmd.Flags().BoolVar(&opts.collectTest, "collect-test", config.DefaultCollectTest,
		"include test files in analysis")
	cmd.Flags().IntVar(&opts.workers, "workers", config.DefaultWorkers,
		"max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", config.DefaultVerbose,
		"print progress information during processing")

	return cmd
}

// runFingerprint executes the fingerprint pipeline, shared between the
// cobra command and the TUI runner.
func runFingerprint(cfg *config.FingerprintConfig, con *console.Console) error {
	con.SetVerbose(cfg.Verbose)
	registry, initDur := newRegistry()

	onScan, onParse := progressCallbacks(con)
	gr, err := pipeline.BuildGraph(pipeline.BuildOptions{
		InputPath:   cfg.InputPath,
		Workers:     cfg.Workers,
		CollectTest: cfg.CollectTest,
		Fingerprint: true,
	}, registry, initDur, onScan, onParse)
	con.ProgressDone()
	if err != nil {
		return err
	}

	for i := range gr.Skipped {
		con.Warn(gr.Skipped[i].Error())
	}
	for i := range gr.ParseErrors {
		con.Warn(gr.ParseErrors[i].Error())
	}

	e := &emitter.Emitter{}
	res, err := e.EmitFingerprints(gr.Graph, &emitter.FingerprintOptions{
		OutputPath:    cfg.Output,
		MinSimilarity: cfg.MinSim,
		MaxSimilarity: cfg.MaxSim,
		EmbedModel:    cfg.EmbedModel,
		ShowAll:       cfg.ShowAll,
	})
	if err != nil {
		return err
	}

	con.Success(fmt.Sprintf("Fingerprints written to %s (%d symbols fingerprinted, %d duplicate pairs found)",
		cfg.Output, res.SymbolsFingerprinted, res.DuplicatePairs))

	return nil
}
