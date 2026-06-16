// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package extract provides tree-sitter AST extraction utilities shared
// across all language plugins.
package extract

import "codeknit/internal/common/types"

// Collector accumulates Symbols and Edges during AST extraction.
type Collector struct {
	FilePath    string
	Symbols     []types.Symbol
	Edges       []types.Edge
	Fingerprint bool // when true, extract body tokens for fingerprinting
}

// AddSymbol stamps the collector's FilePath onto the symbol and appends it.
func (c *Collector) AddSymbol(s *types.Symbol) {
	s.FilePath = c.FilePath
	c.Symbols = append(c.Symbols, *s)
}

// AddEdge appends an edge to the collector.
func (c *Collector) AddEdge(e types.Edge) {
	c.Edges = append(c.Edges, e)
}

// KnownCallables returns a set of symbol names that have CategoryCallable
// in the collector so far.
func (c *Collector) KnownCallables() map[string]bool {
	m := make(map[string]bool)
	for i := range c.Symbols {
		if c.Symbols[i].Category == types.CategoryCallable {
			m[c.Symbols[i].Name] = true
		}
	}
	return m
}

// KnownCallablesAndImports returns a set of names that are either callable
// symbols or imported names. Imported names are included because they may
// refer to callables defined in other files — the per-file extractor can't
// know their category, but the planner will resolve them later.
func (c *Collector) KnownCallablesAndImports() map[string]bool {
	m := c.KnownCallables()
	for i := range c.Edges {
		if c.Edges[i].Kind == types.EdgeImports && c.Edges[i].From != "*" {
			m[c.Edges[i].From] = true
		}
	}
	return m
}
