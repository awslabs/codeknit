// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pipeline provides concurrent file parsing for the codeknit pipeline.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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
// ParseResult.Skipped and ParseResult.ParseErrors.
func ParseFiles(
	files []string,
	registry *plugin.Registry,
	workers int,
	onProgress func(done, total int),
	fingerprint bool,
	readOptions ...ReadOptions,
) *ParseResult {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Per-file result slot — each goroutine writes to its own index, no lock needed.
	results := make([]fileResult, len(files))
	total := len(files)
	var completed int64
	var reader *sourceReader
	var openErr error
	if len(readOptions) > 0 {
		reader, openErr = openSourceReader(readOptions[0])
		if reader != nil {
			defer func() { _ = reader.root.Close() }()
		}
	}

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

			var syms []plugin.Symbol
			var edgs []plugin.Edge
			err := openErr
			if err == nil && reader != nil {
				var source []byte
				source, err = reader.read(file)
				if err == nil {
					if parser, ok := p.(plugin.SourceParser); ok {
						syms, edgs, err = parser.ParseSource(file, source, fingerprint)
					} else {
						err = errors.New("plugin does not support bounded source parsing")
					}
				}
			} else if err == nil {
				syms, edgs, err = plugin.ParseFile(p, file, fingerprint)
			}
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

// ReadOptions binds source reads to the scan root and enforces the size limit
// again when reading, including files that grow after discovery.
type ReadOptions struct {
	Root        string
	MaxFileSize int64
}

type sourceReader struct {
	root *os.Root
	opts ReadOptions
}

func openSourceReader(opts ReadOptions) (*sourceReader, error) {
	root, err := os.OpenRoot(opts.Root)
	if err != nil {
		return nil, err
	}
	return &sourceReader{root: root, opts: opts}, nil
}

func (r *sourceReader) read(file string) ([]byte, error) {
	rel, err := filepath.Rel(r.opts.Root, file)
	if err != nil || !filepath.IsLocal(rel) {
		return nil, fmt.Errorf("source %s is outside scan root", file)
	}
	info, err := r.root.Lstat(rel)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("source is no longer a regular file; symlinks are not followed")
	}
	f, err := r.root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err = f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("source is no longer a regular file")
	}
	limit := r.opts.MaxFileSize
	if limit > 0 && info.Size() > limit {
		return nil, fmt.Errorf("file size %d exceeds max-file-size %d", info.Size(), limit)
	}
	var source io.Reader = f
	if limit > 0 {
		// Bound the allocation even if the source grows while being read.
		source = io.LimitReader(f, limit)
	}
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, err
	}
	if limit > 0 && int64(len(data)) == limit {
		var extra [1]byte
		n, err := f.Read(extra[:])
		if n != 0 {
			return nil, fmt.Errorf("source exceeds max-file-size %d while reading", limit)
		}
		if err != nil && err != io.EOF {
			return nil, err
		}
	}
	return data, nil
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
