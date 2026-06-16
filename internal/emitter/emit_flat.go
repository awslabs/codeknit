// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"codeknit/internal/ir"
)

// emitDirectoryFlat writes numbered map_NNN.skt files into the output directory.
func (e *Emitter) emitDirectoryFlat(sg *ir.SymbolGraph, opts *EmitOptions) ([]string, error) {
	var dict *Dictionary
	if opts.Minify {
		dict = NewDictionary(sg)
	}

	blocks := buildBlocks(sg, dict)
	outputFiles := splitBlocks(blocks, opts)

	// Write auxiliary files (dict + warnings).
	aux, auxErr := writeAuxFiles(opts, dict, sg)
	if auxErr != nil {
		return nil, auxErr
	}

	written := make([]string, len(outputFiles))
	errs := make([]error, len(outputFiles))

	workers := runtime.NumCPU()
	if workers > len(outputFiles) {
		workers = len(outputFiles)
	}

	var wg sync.WaitGroup
	ch := make(chan int, len(outputFiles))
	for i := range outputFiles {
		ch <- i
	}
	close(ch)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				name := fmt.Sprintf("map_%03d.skt", i+1)
				path := filepath.Join(opts.OutputDir, name)
				if err := os.WriteFile(path, []byte(outputFiles[i]), 0o600); err != nil {
					errs[i] = fmt.Errorf("writing %s: %w", path, err)
					return
				}
				written[i] = path
			}
		}()
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			// Return whatever was written before the first error.
			var partial []string
			for _, p := range written {
				if p != "" {
					partial = append(partial, p)
				}
			}
			return partial, err
		}
	}

	if aux.Dict != "" {
		written = append(written, aux.Dict)
	}

	if aux.Errors != "" {
		written = append(written, aux.Errors)
	}

	return written, nil
}

// splitBlocks distributes file blocks across output files respecting MaxLines.
// The dictionary (if any) is written separately by the caller; splitBlocks only
// handles [symbols] and [edges] sections.
// No file's symbol block is split across two output files.
func splitBlocks(blocks []fileBlock, opts *EmitOptions) []string {
	type outputChunk struct {
		symbolLines []string
		edgeLines   []string
	}

	chunks := make([]outputChunk, 0, len(blocks))
	for _, b := range blocks {
		chunks = append(chunks, outputChunk{
			symbolLines: b.symbolLines,
			edgeLines:   b.edgeLines,
		})
	}

	var files []string
	var currentLines []string
	currentLineCount := 0

	currentLines = append(currentLines, "[symbols]")
	currentLineCount++

	var edgeBuffer []string

	for _, chunk := range chunks {
		chunkSize := len(chunk.symbolLines) + len(chunk.edgeLines)

		if currentLineCount > 1 && currentLineCount+chunkSize+2 > opts.MaxLines {
			if len(edgeBuffer) > 0 {
				currentLines = append(currentLines, "", "[edges]")
				currentLines = append(currentLines, edgeBuffer...)
			}
			files = append(files, strings.Join(currentLines, "\n")+"\n")

			currentLines = []string{"[symbols]"}
			currentLineCount = 1
			edgeBuffer = nil
		}

		currentLines = append(currentLines, chunk.symbolLines...)
		currentLineCount += len(chunk.symbolLines)
		edgeBuffer = append(edgeBuffer, chunk.edgeLines...)
		currentLineCount += len(chunk.edgeLines)
	}

	if len(edgeBuffer) > 0 {
		currentLines = append(currentLines, "", "[edges]")
		currentLines = append(currentLines, edgeBuffer...)
	}
	files = append(files, strings.Join(currentLines, "\n")+"\n")

	return files
}
