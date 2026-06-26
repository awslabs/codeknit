// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"fmt"
	"path/filepath"
	"time"

	"codeknit/internal/config"
	"codeknit/internal/emitter"
	"codeknit/internal/fingerprint"
	"codeknit/internal/ir"
	"codeknit/internal/planner"
	"codeknit/internal/plugin"
	"codeknit/internal/scanner"
)

// Timings holds the duration of each pipeline stage.
type Timings struct {
	Init  time.Duration
	Scan  time.Duration
	Parse time.Duration
	Plan  time.Duration
	Emit  time.Duration
	Total time.Duration
}

// Counts holds the summary counts for the pipeline run.
type Counts struct {
	FilesProcessed int
	Skipped        int
	ParseErrors    int
	OutputFiles    int
}

// ProgressFunc is called by pipeline stages to report progress.
type ProgressFunc func(msg string)

// Result holds the complete output of a pipeline execution.
type Result struct {
	Skipped     []SkipError
	ParseErrors []ir.ParseError
	Written     []string
	Timings     Timings
	Counts      Counts
}

// GraphResult holds the output of the shared scan → parse → plan pipeline.
type GraphResult struct {
	Graph       *ir.SymbolGraph
	Files       []string // relative paths from scan
	Skipped     []SkipError
	ParseErrors []ir.ParseError
	Timings     Timings
}

// BuildOptions holds the inputs BuildGraph actually reads. It is intentionally
// narrower than any single command's config so callers can't accidentally
// depend on unrelated fields.
type BuildOptions struct {
	InputPath   string
	Workers     int
	CollectTest bool
	Fingerprint bool
	NoEdges     bool
}

// BuildGraph runs the shared pipeline stages (scan → parse → plan) and returns
// the SymbolGraph. Callers supply their own emitter logic afterward.
// This is the single source of truth for building a graph from source files.
func BuildGraph(
	opts BuildOptions,
	registry *plugin.Registry,
	initDur time.Duration,
	onScanProgress func(visited, matched int),
	onParseProgress func(done, total int),
) (*GraphResult, error) {
	runStart := time.Now().Add(-initDur)
	res := &GraphResult{}
	res.Timings.Init = initDur

	// 1. Scan for source files.
	t0 := time.Now()
	sc := &scanner.Scanner{
		Extensions:   registry.AllExtensions(),
		TestPatterns: registry.AllTestPatterns(),
		CollectTest:  opts.CollectTest,
		OnProgress:   onScanProgress,
	}
	files, err := sc.Scan(opts.InputPath)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}
	res.Timings.Scan = time.Since(t0)
	if len(files) == 0 {
		return nil, fmt.Errorf("no supported source files found in %s", opts.InputPath)
	}
	res.Files = files

	// Resolve file paths relative to input path for the parser.
	absPaths := make([]string, len(files))
	for i, f := range files {
		absPaths[i] = filepath.Join(opts.InputPath, f)
	}

	// 2. Parse files concurrently.
	t1 := time.Now()
	pr := ParseFiles(absPaths, registry, opts.Workers, onParseProgress, opts.Fingerprint)
	res.Timings.Parse = time.Since(t1)
	res.Skipped = pr.Skipped
	res.ParseErrors = pr.ParseErrors

	// 3. Plan: build SymbolGraph.
	t2 := time.Now()
	p := &planner.Planner{}
	sg := p.Plan(pr.Symbols, pr.Edges)
	res.Timings.Plan = time.Since(t2)

	// 3b. Post-plan fingerprinting: compute type fingerprints from children.
	// This must happen after Plan() because it depends on EdgeContains edges.
	if opts.Fingerprint {
		fingerprint.Types(sg)
	}

	if opts.NoEdges {
		sg.Edges = nil
	}

	sg.Errors = append(sg.Errors, pr.ParseErrors...)
	res.Graph = sg

	res.Timings.Total = time.Since(runStart)
	return res, nil
}

// Execute runs the full parse pipeline: scan → parse → plan → emit.
// It is free of any console/UI concerns — callers inspect the returned Result
// to decide what to display.
//
// onScanProgress and onParseProgress are optional callbacks for live progress
// reporting. Pass nil to suppress progress.
func Execute(
	cfg *config.ParseConfig,
	registry *plugin.Registry,
	initDur time.Duration,
	onScanProgress func(visited, matched int),
	onParseProgress func(done, total int),
) (*Result, error) {
	gr, err := BuildGraph(BuildOptions{
		InputPath:   cfg.InputPath,
		Workers:     cfg.Workers,
		CollectTest: cfg.CollectTest,
		NoEdges:     !cfg.Edges,
	}, registry, initDur, onScanProgress, onParseProgress)
	if err != nil {
		return nil, err
	}

	res := &Result{}
	res.Timings = gr.Timings
	res.Skipped = gr.Skipped
	res.ParseErrors = gr.ParseErrors

	// 4. Emit output.
	t3 := time.Now()
	e := &emitter.Emitter{}
	opts := &emitter.EmitOptions{
		OutputDir:    cfg.OutputDir,
		OutputMode:   cfg.OutputMode,
		OutputFormat: cfg.OutputFormat,
		InputPath:    cfg.InputPath,
		FileOrder:    gr.Files,
		MaxLines:     cfg.MaxLines,
		Minify:       cfg.Minify,
		Clean:        cfg.Clean,
	}
	written, err := e.Emit(gr.Graph, opts)
	if err != nil {
		return nil, fmt.Errorf("emit failed: %w", err)
	}
	res.Timings.Emit = time.Since(t3)
	res.Written = written

	// 5. Compute summary counts.
	res.Counts.FilesProcessed = len(gr.Skipped) + len(gr.ParseErrors) + len(gr.Graph.FileOrder)
	res.Counts.OutputFiles = len(written)
	res.Counts.Skipped = len(gr.Skipped)
	res.Counts.ParseErrors = len(gr.ParseErrors)

	res.Timings.Total = gr.Timings.Total
	return res, nil
}
