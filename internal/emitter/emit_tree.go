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

// writeTask represents a single file write operation for parallel execution.
type writeTask struct {
	path    string
	content string
}

// emitDirectoryTree writes .skt files that mirror the input directory structure.
func (e *Emitter) emitDirectoryTree(sg *ir.SymbolGraph, opts *EmitOptions) ([]string, error) {
	var dict *Dictionary
	if opts.Minify {
		dict = NewDictionary(sg)
	}

	blocks := buildBlocks(sg, dict)

	// Build a map from source file path to its block for quick lookup.
	// Block file paths are absolute; normalize to relative (matching FileOrder)
	// by stripping the InputPath prefix when present.
	blockByFile := make(map[string]fileBlock, len(blocks))
	for _, b := range blocks {
		key := b.filePath
		if opts.InputPath != "" {
			rel, err := filepath.Rel(opts.InputPath, b.filePath)
			if err == nil {
				key = rel
			}
		}
		blockByFile[key] = b
	}

	// Determine whether InputPath is a single file (not a directory).
	singleFile := false
	if info, err := os.Stat(opts.InputPath); err == nil && !info.IsDir() {
		singleFile = true
	}

	// Write auxiliary files (dict + errors as separate .skt files in tree mode).
	aux, auxErr := writeAuxFiles(opts, dict, sg)
	if auxErr != nil {
		return nil, auxErr
	}

	// Use FileOrder to determine output structure; fall back to sg.FileOrder from blocks.
	order := opts.FileOrder
	if len(order) == 0 {
		order = sg.FileOrder
	}

	// Phase 1: prepare all write tasks sequentially (content rendering + path computation).
	// Collect unique directories to create them before parallel writes.
	var tasks []writeTask
	dirs := make(map[string]struct{})

	for _, srcRel := range order {
		block, ok := blockByFile[srcRel]
		if !ok {
			continue
		}

		ext := filepath.Ext(srcRel)
		outRel := strings.TrimSuffix(srcRel, ext) + ".skt"

		var outBase string
		if singleFile {
			outBase = filepath.Join(opts.OutputDir, filepath.Base(outRel))
		} else {
			outBase = filepath.Join(opts.OutputDir, outRel)
		}

		content := renderTreeFileContent(block)

		contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
		if len(contentLines) == 1 && contentLines[0] == "" {
			contentLines = nil
		}

		if len(contentLines) <= opts.MaxLines {
			if dir := filepath.Dir(outBase); dir != "." {
				dirs[dir] = struct{}{}
			}
			tasks = append(tasks, writeTask{path: outBase, content: content})
		} else {
			parts := splitTreeContent(block, opts.MaxLines)
			baseNoExt := strings.TrimSuffix(outBase, ".skt")
			for i, partContent := range parts {
				partPath := fmt.Sprintf("%s_part%d.skt", baseNoExt, i+1)
				if dir := filepath.Dir(partPath); dir != "." {
					dirs[dir] = struct{}{}
				}
				tasks = append(tasks, writeTask{path: partPath, content: partContent})
			}
		}
	}

	// Phase 2: create all directories up front (sequential — cheap and avoids races).
	for dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories (execute bit required for traversal)
			return nil, fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	// Phase 3: write files in parallel.
	written := make([]string, len(tasks))
	errs := make([]error, len(tasks))

	workers := runtime.NumCPU()
	if workers > len(tasks) {
		workers = len(tasks)
	}

	var wg sync.WaitGroup
	ch := make(chan int, len(tasks))
	for i := range tasks {
		ch <- i
	}
	close(ch)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				t := tasks[i]
				if err := os.WriteFile(t.path, []byte(t.content), 0o600); err != nil {
					errs[i] = fmt.Errorf("writing %s: %w", t.path, err)
					return
				}
				written[i] = t.path
			}
		}()
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
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

// renderTreeFileContent renders the output content for a single source file in tree mode.
func renderTreeFileContent(block fileBlock) string {
	var lines []string

	// [symbols] section with this file's symbols.
	lines = append(lines, "[symbols]")
	lines = append(lines, block.symbolLines...)

	// [edges] section with this file's edges.
	if len(block.edgeLines) > 0 {
		lines = append(lines, "", "[edges]")
		lines = append(lines, block.edgeLines...)
	}

	return strings.Join(lines, "\n") + "\n"
}

// splitTreeContent splits a single source file's tree-mode output into multiple parts,
// each not exceeding maxLines. Each part is a self-contained output chunk with proper
// section headers ([symbols], [edges]).
func splitTreeContent(block fileBlock, maxLines int) []string {
	symbolLines := block.symbolLines
	edgeLines := block.edgeLines

	var parts []string

	symIdx := 0
	edgeIdx := 0

	for symIdx < len(symbolLines) || edgeIdx < len(edgeLines) {
		var lines []string
		lineCount := 0

		// Add [symbols] header.
		lines = append(lines, "[symbols]")
		lineCount++

		// Fill with symbol lines up to maxLines.
		for symIdx < len(symbolLines) && lineCount < maxLines {
			lines = append(lines, symbolLines[symIdx])
			lineCount++
			symIdx++
		}

		// If we still have room and there are edge lines, add them.
		if edgeIdx < len(edgeLines) && lineCount+2 < maxLines {
			lines = append(lines, "", "[edges]")
			lineCount += 2
			for edgeIdx < len(edgeLines) && lineCount < maxLines {
				lines = append(lines, edgeLines[edgeIdx])
				lineCount++
				edgeIdx++
			}
		}

		parts = append(parts, strings.Join(lines, "\n")+"\n")
	}

	// If there are remaining edge lines that didn't fit, create additional parts for them.
	for edgeIdx < len(edgeLines) {
		var lines []string
		lineCount := 0

		lines = append(lines, "[symbols]")
		lineCount++

		lines = append(lines, "", "[edges]")
		lineCount += 2

		for edgeIdx < len(edgeLines) && lineCount < maxLines {
			lines = append(lines, edgeLines[edgeIdx])
			lineCount++
			edgeIdx++
		}

		parts = append(parts, strings.Join(lines, "\n")+"\n")
	}

	// Edge case: if no parts were created (empty block), return one empty part.
	if len(parts) == 0 {
		parts = append(parts, "[symbols]\n")
	}

	return parts
}
