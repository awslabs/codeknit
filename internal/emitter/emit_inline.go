// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"io"

	"codeknit/internal/ir"
)

// EmitInline writes all output as a single continuous stream to the provided
// io.Writer. No file splitting is applied. It creates no files on disk.
// The SymbolGraph must have its indexes built (via BuildIndexes) before calling.
func (e *Emitter) EmitInline(w io.Writer, sg *ir.SymbolGraph, opts *EmitOptions) error {
	var dict *Dictionary
	if opts.Minify {
		dict = NewDictionary(sg)
	}

	blocks := buildBlocks(sg, dict)

	// Write [dict] section if minified.
	if dict != nil {
		if _, err := fmt.Fprint(w, renderDictSection(dict)+"\n"); err != nil {
			return err
		}
	}

	// Write [symbols] section.
	if _, err := fmt.Fprintln(w, "[symbols]"); err != nil {
		return err
	}
	for _, b := range blocks {
		for _, line := range b.symbolLines {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}

	// Write [edges] section.
	var allEdges []string
	for _, b := range blocks {
		allEdges = append(allEdges, b.edgeLines...)
	}
	if len(allEdges) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "[edges]"); err != nil {
			return err
		}
		for _, line := range allEdges {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}

	// Write [errors] section if there are parse errors.
	errSection := renderErrorsSection(sg, opts.InputPath)
	if errSection != "" {
		if _, err := fmt.Fprint(w, "\n"+errSection); err != nil {
			return err
		}
	}

	return nil
}
