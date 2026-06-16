// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeknit/internal/ir"
)

// graphSymbol is the JSON structure for a symbol in the graph HTML.
type graphSymbol struct {
	Props     map[string]string `json:"props"`
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	ShortID   string            `json:"shortId"`
	File      string            `json:"file"`
	Category  string            `json:"category"`
	Kind      string            `json:"kind"`
	Signature string            `json:"signature"`
	Span      [2]int            `json:"span"`
}

// graphEdge is the JSON structure for an edge in the graph HTML.
type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// graphData is the top-level JSON structure injected into the HTML template.
type graphData struct {
	Symbols []graphSymbol `json:"symbols"`
	Edges   []graphEdge   `json:"edges"`
}

// EmitGraph generates a self-contained HTML file with an interactive graph
// visualization of the SymbolGraph. The output file is written to outputPath.
func (e *Emitter) EmitGraph(sg *ir.SymbolGraph, outputPath string) error {
	data := buildGraphData(sg)

	jsonBytes, err := json.MarshalIndent(data, "    ", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph data: %w", err)
	}

	// Place the JSON inside a <script type="application/json"> tag.
	// Escape "</script>" sequences that could prematurely close the tag.
	safeJSON := strings.ReplaceAll(string(jsonBytes), "</script>", `<\/script>`)
	html := strings.Replace(graphTemplateHTML, "/*GRAPH_DATA_JSON*/", safeJSON, 1)

	if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories (execute bit required for traversal)
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, []byte(html), 0o600); err != nil {
		return fmt.Errorf("write graph HTML: %w", err)
	}

	return nil
}

func buildGraphData(sg *ir.SymbolGraph) graphData {
	symbols := make([]graphSymbol, 0, len(sg.Symbols))
	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		sid := sg.ShortIDs[sym.ID]
		props := sym.Properties
		if props == nil {
			props = map[string]string{}
		}
		symbols = append(symbols, graphSymbol{
			ID:        sym.ID,
			ShortID:   sid,
			Name:      sym.Name,
			File:      sym.FilePath,
			Category:  string(sym.Category),
			Kind:      sym.Kind,
			Signature: sym.Signature,
			Span:      sym.Span,
			Props:     props,
		})
	}

	edges := make([]graphEdge, 0, len(sg.Edges))
	for _, edge := range sg.Edges {
		fromSID, fromOK := sg.ShortIDs[edge.From]
		toSID, toOK := sg.ShortIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}
		edges = append(edges, graphEdge{
			From: fromSID,
			To:   toSID,
			Kind: string(edge.Kind),
		})
	}

	return graphData{
		Symbols: symbols,
		Edges:   edges,
	}
}
