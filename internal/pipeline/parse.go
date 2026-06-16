// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pipeline provides concurrent file parsing for the codeknit pipeline.
package pipeline

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"

	"golang.org/x/sync/errgroup"
)

// ParseResult holds the aggregated output of parsing all files.
type ParseResult struct {
	Symbols     map[string][]plugin.Symbol
	Edges       map[string][]plugin.Edge
	Skipped     []SkipError
	ParseErrors []ir.ParseError
}

// ParseFiles parses all files concurrently using a bounded worker pool.
// workers sets the max concurrency (0 or negative = runtime.NumCPU()).
// onProgress, if non-nil, is called after each file completes with the
// count of files parsed so far.
// When fingerprint is true, body token extraction is enabled for callable symbols.
// Non-fatal errors (skipped files, syntax warnings) are collected in
// ParseResult.Warnings.
func ParseFiles(
	files []string,
	registry *plugin.Registry,
	workers int,
	onProgress func(done, total int),
	fingerprint bool,
) *ParseResult {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Per-file result slot — each goroutine writes to its own index, no lock needed.
	results := make([]fileResult, len(files))
	total := len(files)
	var completed int64

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(workers)

	for i, f := range files {
		idx, file := i, f
		g.Go(func() error {
			defer func() {
				if onProgress != nil {
					onProgress(int(atomic.AddInt64(&completed, 1)), total)
				}
			}()
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			ext := filepath.Ext(file)
			p, ok := registry.Lookup(ext)
			if !ok {
				results[idx].skip = &SkipError{FilePath: file, Reason: "unsupported extension: " + ext}
				return nil
			}

			syms, edgs, err := parseFile(p, file, fingerprint)
			if err != nil {
				var sw *plugin.SyntaxError
				if errors.As(err, &sw) {
					results[idx].syms = syms
					results[idx].edges = edgs
					results[idx].parseError = &ir.ParseError{FilePath: file, Reason: sw.Message}
					return nil
				}
				results[idx].parseError = &ir.ParseError{FilePath: file, Reason: err.Error()}
				return nil
			}

			results[idx].syms = syms
			results[idx].edges = edgs
			return nil
		})
	}

	// Wait is called for its side-effect of blocking until all goroutines
	// complete; the error (context cancellation) is non-fatal because
	// partial results are still useful.
	_ = g.Wait()
	return mergeResults(files, results)
}

// FingerprintParser is an optional interface for plugins that support
// fingerprint-aware parsing.
type FingerprintParser interface {
	ParseWithOptions(filePath string, fingerprint bool) ([]plugin.Symbol, []plugin.Edge, error)
}

// parseFile dispatches to ParseWithOptions when fingerprinting is enabled
// and the plugin supports it, otherwise falls back to Parse.
func parseFile(p plugin.LanguagePlugin, file string, fingerprint bool) (symbols []plugin.Symbol, edges []plugin.Edge, err error) {
	if fingerprint {
		if fp, ok := p.(FingerprintParser); ok {
			return fp.ParseWithOptions(file, true)
		}
	}
	return p.Parse(file)
}

// fileResult holds the output of a single file parse for lock-free collection.
type fileResult struct {
	skip       *SkipError
	parseError *ir.ParseError
	syms       []plugin.Symbol
	edges      []plugin.Edge
}

// mergeResults collects per-file slots into the final ParseResult.
func mergeResults(files []string, results []fileResult) *ParseResult {
	pr := &ParseResult{
		Symbols: make(map[string][]plugin.Symbol, len(files)),
		Edges:   make(map[string][]plugin.Edge, len(files)),
	}
	for i, file := range files {
		r := &results[i]
		if r.syms != nil {
			pr.Symbols[file] = r.syms
			pr.Edges[file] = r.edges
		}
		if r.skip != nil {
			pr.Skipped = append(pr.Skipped, *r.skip)
		}
		if r.parseError != nil {
			pr.ParseErrors = append(pr.ParseErrors, *r.parseError)
		}
	}
	return pr
}

// SkipError represents a non-fatal file skip during parsing.
type SkipError struct {
	FilePath string
	Reason   string
}

func (e *SkipError) Error() string {
	return e.FilePath + ": " + e.Reason
}
