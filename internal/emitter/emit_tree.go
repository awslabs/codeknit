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
	"codeknit/internal/outputfs"
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

	// Write auxiliary output into the unpublished staging directory.
	aux, auxErr := writeAuxFiles(opts, dict, sg)
	if auxErr != nil {
		return nil, auxErr
	}

	// Phase 1: prepare all write tasks sequentially (content rendering + path computation).
	// Collect unique directories to create them before parallel writes.
	var tasks []writeTask
	dirs := make(map[string]struct{})

	for _, block := range blocks {
		srcRel, err := treeSourcePath(block.filePath, opts.InputPath)
		if err != nil {
			return nil, err
		}
		// Keep the source extension to distinguish foo.c, foo.h and foo.go,
		// and to reserve dict.skt and warnings.skt for auxiliary output.
		outBase := filepath.Join(opts.OutputDir, srcRel+".skt")

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

	if err := validateTreePaths(tasks, aux); err != nil {
		return nil, err
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
				if err := outputfs.WriteFile(t.path, []byte(t.content)); err != nil {
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

// Validate the entire namespace before parallel writes, including directories
// and auxiliary files. Reject case-only collisions for portable output trees.
func validateTreePaths(tasks []writeTask, aux auxPaths) error {
	paths := make(map[string]string, len(tasks)+2)
	register := func(path string) error {
		if path == "" {
			return nil
		}
		key := strings.ToLower(filepath.Clean(path))
		if previous, exists := paths[key]; exists {
			return fmt.Errorf("tree output collision between %s and %s", previous, path)
		}
		paths[key] = path
		return nil
	}
	for _, task := range tasks {
		if err := register(task.path); err != nil {
			return err
		}
	}
	for _, path := range []string{aux.Dict, aux.Errors} {
		if err := register(path); err != nil {
			return err
		}
	}
	for key, path := range paths {
		for dir := filepath.Dir(key); dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			if file, exists := paths[dir]; exists {
				return fmt.Errorf("tree output %s conflicts with directory needed by %s", file, path)
			}
		}
	}
	return nil
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
	// Public emitters reject smaller budgets. Keep this helper total as well.
	if maxLines < 3 {
		return nil
	}
	header := "## " + block.filePath
	symbols := block.symbolLines
	if len(symbols) > 0 && strings.HasPrefix(symbols[0], "## ") {
		header, symbols = symbols[0], symbols[1:]
	}
	edges := block.edgeLines
	var parts []string
	for len(symbols) > 0 {
		n := min(len(symbols), maxLines-2)
		lines := append([]string{"[symbols]", header}, symbols[:n]...)
		symbols = symbols[n:]
		if len(symbols) == 0 && maxLines-len(lines) >= 3 && len(edges) > 0 {
			count := min(len(edges), maxLines-len(lines)-2)
			lines = append(lines, "", "[edges]")
			lines = append(lines, edges[:count]...)
			edges = edges[count:]
		}
		parts = append(parts, strings.Join(lines, "\n")+"\n")
	}
	for len(edges) > 0 {
		n := min(len(edges), maxLines-1)
		lines := append([]string{"[edges]"}, edges[:n]...)
		parts = append(parts, strings.Join(lines, "\n")+"\n")
		edges = edges[n:]
	}
	if len(parts) == 0 {
		parts = append(parts, "[symbols]\n"+header+"\n")
	}
	return parts
}

func treeSourcePath(file, input string) (string, error) {
	base := input
	if info, err := os.Stat(base); err == nil && !info.IsDir() {
		base = filepath.Dir(base)
	}
	rel := file
	if base != "" && filepath.IsAbs(file) {
		var err error
		rel, err = filepath.Rel(base, file)
		if err != nil {
			return "", err
		}
	} else if filepath.IsAbs(file) {
		rel = filepath.Base(file)
	}
	if !filepath.IsLocal(rel) || rel == "." {
		return "", fmt.Errorf("source %q is outside input %q", file, input)
	}
	return rel, nil
}
