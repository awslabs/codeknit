// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package parser parses emitter text output back into a SymbolGraph.
package parser

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// ParseOutput reads emitter text output and reconstructs a SymbolGraph.
// readers contains the content of each output file in order.
// minified indicates whether the output uses dictionary-based minification.
func ParseOutput(readers []io.Reader, minified bool) (*ir.SymbolGraph, error) {
	sg := &ir.SymbolGraph{}

	// dict code → token mapping (populated from [dict] section).
	dict := make(map[string]string)

	// shortID → full symbol ID mapping (built during symbol parsing).
	shortToFull := make(map[string]string)

	for _, r := range readers {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("reading input: %w", err)
		}
		content := string(data)
		lines := strings.Split(strings.TrimRight(content, "\n"), "\n")

		var section string // "dict", "symbols", "edges"
		var currentFile string

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Detect section headers.
			switch trimmed {
			case "[dict]":
				section = "dict"
				continue
			case "[symbols]":
				section = "symbols"
				continue
			case "[edges]":
				section = "edges"
				continue
			case "[errors]":
				section = "errors"
				continue
			}

			if trimmed == "" {
				continue
			}

			switch section {
			case "dict":
				// Format: "- code: token"
				if strings.HasPrefix(trimmed, "- ") {
					rest := trimmed[2:]
					idx := strings.Index(rest, ": ")
					if idx < 0 {
						return nil, fmt.Errorf("invalid dict line: %s", line)
					}
					code := rest[:idx]
					token := rest[idx+2:]
					dict[code] = token
				}

			case "symbols":
				// File header: "## filepath"
				if strings.HasPrefix(trimmed, "## ") {
					currentFile = trimmed[3:]
					continue
				}

				sym, shortID, err := parseSymbolLine(trimmed, currentFile, dict, minified)
				if err != nil {
					return nil, fmt.Errorf("parsing symbol line %q: %w", line, err)
				}
				shortToFull[shortID] = sym.ID
				sg.Symbols = append(sg.Symbols, sym)

			case "edges":
				edges, err := parseEdgeLine(trimmed, shortToFull, dict, minified)
				if err != nil {
					return nil, fmt.Errorf("parsing edge line %q: %w", line, err)
				}
				sg.Edges = append(sg.Edges, edges...)

			case "errors":
				// Format: "- filepath: reason"
				if strings.HasPrefix(trimmed, "- ") {
					rest := trimmed[2:]
					idx := strings.Index(rest, ": ")
					if idx >= 0 {
						sg.Errors = append(sg.Errors, ir.ParseError{
							FilePath: rest[:idx],
							Reason:   rest[idx+2:],
						})
					}
				}
			}
		}
	}

	return sg, nil
}

// symbolLineRe matches a symbol line:
// ShortID category/kind L{start}-L{end} signature [{props}]
// The {props} suffix is optional — lines without properties omit it.
var symbolLineRe = regexp.MustCompile(
	`^(\S+)\s+(\S+)\s+L(\d+)-L(\d+)\s+(.*?)(?:\s+\{([^}]*)\})?$`,
)

// parseSymbolLine parses a single symbol line and returns the Symbol and its short ID.
func parseSymbolLine(
	line, filePath string,
	dict map[string]string,
	minified bool,
) (sym plugin.Symbol, shortID string, err error) {
	m := symbolLineRe.FindStringSubmatch(line)
	if m == nil {
		return plugin.Symbol{}, "", fmt.Errorf("line does not match symbol format")
	}

	shortID = m[1]
	catKindStr := m[2]
	startLine, _ := strconv.Atoi(m[3])
	endLine, _ := strconv.Atoi(m[4])
	signature := m[5]
	propsStr := m[6]

	// Decode category/kind if it's a dict code.
	if minified {
		if decoded, ok := dict[catKindStr]; ok {
			catKindStr = decoded
		}
	}

	// Split category/kind.
	parts := strings.SplitN(catKindStr, "/", 2)
	if len(parts) != 2 {
		return plugin.Symbol{}, "", fmt.Errorf("invalid category/kind: %s", catKindStr)
	}
	category := plugin.SymbolCategory(parts[0])
	kind := parts[1]

	// Extract name from signature (the part before the first '(' or the whole signature).
	name := signature
	if idx := strings.Index(signature, "("); idx >= 0 {
		name = signature[:idx]
	}

	// Build the full symbol ID.
	fullID := filePath + "::" + name

	props := parseProperties(propsStr, dict, minified)

	return plugin.Symbol{
		ID:         fullID,
		Name:       name,
		FilePath:   filePath,
		Category:   category,
		Kind:       kind,
		Signature:  signature,
		Properties: props,
		Span:       [2]int{startLine, endLine},
	}, shortID, nil
}

// parseProperties parses a properties string like "async, exported" or "visibility=public, async".
func parseProperties(s string, dict map[string]string, minified bool) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[string]string{}
	}
	props := make(map[string]string)
	entries := strings.Split(s, ", ")
	for _, p := range entries {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Decode dict code if minified.
		if minified {
			if decoded, ok := dict[p]; ok {
				p = decoded
			}
		}

		if idx := strings.Index(p, "="); idx >= 0 {
			props[p[:idx]] = p[idx+1:]
		} else {
			props[p] = "true"
		}
	}
	return props
}

// edgeLineRe matches an edge line: "FROM --kind--> TO1, TO2, ..."
var edgeLineRe = regexp.MustCompile(`^(\S+)\s+--(\w+)-->\s+(.+)$`)

// parseEdgeLine parses a single edge line which may contain multiple targets.
// When minified is true, the edge kind is decoded through the dictionary.
func parseEdgeLine(line string, shortToFull, dict map[string]string, minified bool) ([]plugin.Edge, error) {
	m := edgeLineRe.FindStringSubmatch(line)
	if m == nil {
		return nil, fmt.Errorf("line does not match edge format")
	}

	fromShort := m[1]
	kindStr := m[2]
	targetsStr := m[3]

	// Decode edge kind if minified.
	if minified {
		if decoded, ok := dict[kindStr]; ok {
			kindStr = decoded
		}
	}
	kind := plugin.EdgeKind(kindStr)

	fromFull, ok := shortToFull[fromShort]
	if !ok {
		fromFull = fromShort
	}

	targets := strings.Split(targetsStr, ", ")
	edges := make([]plugin.Edge, 0, len(targets))
	for _, toShort := range targets {
		toShort = strings.TrimSpace(toShort)
		if toShort == "" {
			continue
		}
		toFull, ok := shortToFull[toShort]
		if !ok {
			toFull = toShort
		}
		edges = append(edges, plugin.Edge{
			From: fromFull,
			To:   toFull,
			Kind: kind,
		})
	}

	return edges, nil
}
