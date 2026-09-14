// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/ir"
	"codeknit/internal/outputfs"
)

// auxPaths holds the file paths of auxiliary .skt files written alongside the
// main output (dictionary and/or errors).
type auxPaths struct {
	Dict   string // empty when minification is disabled
	Errors string // empty when there are no parse errors
}

// writeAuxFiles writes the optional dict.skt and warnings.skt files into the
// output directory. It returns the paths of any files written so callers can
// append them to their result slices.
func writeAuxFiles(opts *EmitOptions, dict *Dictionary, sg *ir.SymbolGraph) (auxPaths, error) {
	var aux auxPaths

	if dict != nil {
		aux.Dict = filepath.Join(opts.OutputDir, "dict.skt")
		if err := os.MkdirAll(opts.OutputDir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories (execute bit required for traversal)
			return aux, fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
		}
		if err := outputfs.WriteFile(aux.Dict, []byte(renderDictSection(dict))); err != nil {
			return aux, fmt.Errorf("writing %s: %w", aux.Dict, err)
		}
	}

	errSection := renderErrorsSection(sg, opts.InputPath)
	if errSection != "" {
		aux.Errors = filepath.Join(opts.OutputDir, "warnings.skt")
		if err := os.MkdirAll(opts.OutputDir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories (execute bit required for traversal)
			return aux, fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
		}
		if err := outputfs.WriteFile(aux.Errors, []byte(errSection)); err != nil {
			return aux, fmt.Errorf("writing %s: %w", aux.Errors, err)
		}
	}

	return aux, nil
}

// renderDictSection renders the [dict] section as a string.
func renderDictSection(dict *Dictionary) string {
	lines := make([]string, 0, 1+len(dict.Forward)+1)
	lines = append(lines, "[dict]")

	type dictEntry struct {
		token string
		code  string
	}
	entries := make([]dictEntry, 0, len(dict.Forward))
	for token, code := range dict.Forward {
		entries = append(entries, dictEntry{token: token, code: code})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].token < entries[j].token
	})

	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("- %s: %s", e.code, e.token))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// renderErrorsSection renders the [errors] section as a string.
// Returns an empty string if there are no errors.
func renderErrorsSection(sg *ir.SymbolGraph, inputPath string) string {
	if len(sg.Errors) == 0 {
		return ""
	}

	lines := make([]string, 0, 1+len(sg.Errors)+1)
	lines = append(lines, "[errors]")

	// Sort errors by file path for deterministic output.
	sorted := make([]ir.ParseError, len(sg.Errors))
	copy(sorted, sg.Errors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FilePath < sorted[j].FilePath
	})

	for _, e := range sorted {
		fp := e.FilePath
		if inputPath != "" {
			if rel, err := filepath.Rel(inputPath, fp); err == nil {
				fp = rel
			}
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", fp, e.Reason))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}
