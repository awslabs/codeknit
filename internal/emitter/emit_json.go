// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"codeknit/internal/config"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// jsonOutput is the stable machine-readable representation emitted by
// `codeknit parse --format json`.
type jsonOutput struct {
	Files   []string     `json:"files"`
	Symbols []jsonSymbol `json:"symbols"`
	Edges   []jsonEdge   `json:"edges,omitempty"`
	Errors  []jsonError  `json:"errors,omitempty"`
}

type jsonSymbol struct {
	Properties map[string]string `json:"properties,omitempty"`
	ID         string            `json:"id"`
	ShortID    string            `json:"short_id"`
	Name       string            `json:"name"`
	ScopedName string            `json:"scoped_name,omitempty"`
	File       string            `json:"file"`
	Category   string            `json:"category"`
	Kind       string            `json:"kind"`
	Signature  string            `json:"signature"`
	Span       [2]int            `json:"span"`
}

type jsonEdge struct {
	From      string `json:"from"`
	FromShort string `json:"from_short"`
	To        string `json:"to"`
	ToShort   string `json:"to_short"`
	Kind      string `json:"kind"`
}

type jsonError struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

func (e *Emitter) emitJSON(sg *ir.SymbolGraph, opts *EmitOptions) ([]string, error) {
	if opts.OutputMode == config.OutputInline {
		return nil, e.emitJSONWithInput(os.Stdout, sg, opts.InputPath)
	}

	if err := prepareJSONOutput(opts); err != nil {
		return nil, err
	}

	path := filepath.Join(opts.OutputDir, "codeknit.json")
	f, err := os.Create(path) //nolint:gosec // output path is derived from user-selected output dir
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}

	if err := e.emitJSONWithInput(f, sg, opts.InputPath); err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close %s: %w", path, err)
	}

	return []string{path}, nil
}

// EmitJSON writes the complete SymbolGraph as a single JSON document.
// The SymbolGraph must have its indexes built before calling.
func (e *Emitter) EmitJSON(w io.Writer, sg *ir.SymbolGraph) error {
	return e.emitJSONWithInput(w, sg, "")
}

func (e *Emitter) emitJSONWithInput(w io.Writer, sg *ir.SymbolGraph, inputPath string) error {
	out := buildJSONOutput(sg, inputPath)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func buildJSONOutput(sg *ir.SymbolGraph, inputPath string) jsonOutput {
	out := jsonOutput{
		Files: make([]string, len(sg.FileOrder)),
	}
	for i, fp := range sg.FileOrder {
		out.Files[i] = displayFilePath(fp, inputPath)
	}

	out.Symbols = make([]jsonSymbol, 0, len(sg.Symbols))
	for _, fp := range sg.FileOrder {
		for _, idx := range sg.ByFile[fp] {
			sym := sg.Symbols[idx]
			sym.Signature = resolveTypeRefs(sym.Signature, sym.Name, fp, sg.ResolveTypeSID)
			out.Symbols = append(out.Symbols, jsonSymbol{
				ID:         sym.ID,
				ShortID:    sg.ShortIDs[sym.ID],
				Name:       sym.Name,
				ScopedName: sym.ScopedName,
				File:       displayFilePath(sym.FilePath, inputPath),
				Category:   string(sym.Category),
				Kind:       sym.Kind,
				Signature:  sym.Signature,
				Span:       sym.Span,
				Properties: sortedProperties(sym.Properties),
			})
		}
	}

	out.Edges = buildJSONEdges(sg)

	if len(sg.Errors) > 0 {
		out.Errors = make([]jsonError, 0, len(sg.Errors))
		for _, err := range sg.Errors {
			out.Errors = append(out.Errors, jsonError{File: displayFilePath(err.FilePath, inputPath), Reason: err.Reason})
		}
	}

	return out
}

func displayFilePath(filePath, inputPath string) string {
	if inputPath == "" {
		return filepath.ToSlash(filePath)
	}
	rel, err := filepath.Rel(inputPath, filePath)
	if err != nil || rel == "." || rel == "" || !filepath.IsLocal(rel) {
		return filepath.ToSlash(filePath)
	}
	return filepath.ToSlash(rel)
}

func buildJSONEdges(sg *ir.SymbolGraph) []jsonEdge {
	type edgeWithFile struct {
		fromFile string
		edge     plugin.Edge
	}
	edges := make([]edgeWithFile, 0, len(sg.Edges))
	for _, edge := range sg.Edges {
		if _, ok := sg.ShortIDs[edge.From]; !ok {
			continue
		}
		if _, ok := sg.ShortIDs[edge.To]; !ok {
			continue
		}
		edges = append(edges, edgeWithFile{fromFile: sg.SymbolFile[edge.From], edge: edge})
	}

	sort.SliceStable(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		if a.fromFile != b.fromFile {
			return a.fromFile < b.fromFile
		}
		if a.edge.From != b.edge.From {
			return a.edge.From < b.edge.From
		}
		if a.edge.Kind != b.edge.Kind {
			return a.edge.Kind < b.edge.Kind
		}
		return a.edge.To < b.edge.To
	})

	out := make([]jsonEdge, 0, len(edges))
	for _, item := range edges {
		edge := item.edge
		out = append(out, jsonEdge{
			From:      edge.From,
			FromShort: sg.ShortIDs[edge.From],
			To:        edge.To,
			ToShort:   sg.ShortIDs[edge.To],
			Kind:      string(edge.Kind),
		})
	}
	return out
}

func sortedProperties(props map[string]string) map[string]string {
	if len(props) == 0 {
		return nil
	}
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(props))
	for _, key := range keys {
		out[key] = props[key]
	}
	return out
}

func prepareJSONOutput(opts *EmitOptions) error {
	if err := os.MkdirAll(opts.OutputDir, 0o700); err != nil { //nolint:gosec // 0o700 is least-privilege for directories
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	path := filepath.Join(opts.OutputDir, "codeknit.json")
	if _, err := os.Stat(path); err == nil {
		if !opts.Clean {
			return fmt.Errorf("output file %s already exists; pass --clean to remove it or choose a different directory", path)
		}
		if removeErr := os.Remove(path); removeErr != nil { //nolint:gosec // output path is controlled by codeknit
			return fmt.Errorf("removing stale output %s: %w", path, removeErr)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking output file %s: %w", path, err)
	}

	return nil
}
