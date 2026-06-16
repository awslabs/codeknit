// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// ExtractCallableRefs runs a second pass over the AST to detect identifiers
// passed as arguments that match known callable symbols.
func (c *Collector) ExtractCallableRefs(root *sitter.Node, src []byte, cfg CallableRefConfig) {
	knownCallables := c.KnownCallables()
	if len(knownCallables) == 0 {
		return
	}

	// Build a set of already-known call targets per caller to avoid duplicates.
	existingCalls := make(map[string]map[string]bool)
	for _, e := range c.Edges {
		if e.Kind == types.EdgeCalls {
			if existingCalls[e.From] == nil {
				existingCalls[e.From] = make(map[string]bool)
			}
			existingCalls[e.From][e.To] = true
		}
	}

	// Build span-to-caller index.
	type callerSpan struct {
		name     string
		startRow int
		endRow   int
	}
	var callers []callerSpan
	for i := range c.Symbols {
		sym := &c.Symbols[i]
		if sym.Category != types.CategoryCallable {
			continue
		}
		callers = append(callers, callerSpan{
			name:     sym.EffectiveScopedName(),
			startRow: sym.Span[0],
			endRow:   sym.Span[1],
		})
	}

	argSet := make(map[string]bool, len(cfg.ArgListKinds))
	for _, k := range cfg.ArgListKinds {
		argSet[k] = true
	}
	identSet := make(map[string]bool, len(cfg.IdentKinds))
	for _, k := range cfg.IdentKinds {
		identSet[k] = true
	}

	callNodeSet := make(map[string]bool, len(cfg.CallNodeKinds))
	for _, k := range cfg.CallNodeKinds {
		callNodeSet[k] = true
	}

	// Single-pass walk: find call expressions, scan their argument lists.
	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if callNodeSet[node.Kind()] {
			// Determine which caller owns this call by span matching.
			callRow := int(node.StartPosition().Row) + 1 //nolint:gosec // Row is a small line number, no overflow risk
			var ownerName string
			for _, cs := range callers {
				if callRow >= cs.startRow && callRow <= cs.endRow {
					ownerName = cs.name
				}
			}
			if ownerName != "" {
				for i := range node.ChildCount() {
					child := node.Child(i)
					if child == nil || !argSet[child.Kind()] {
						continue
					}
					for j := range child.ChildCount() {
						arg := child.Child(j)
						if arg == nil || !identSet[arg.Kind()] {
							continue
						}
						name := arg.Utf8Text(src)
						if name == "" || !knownCallables[name] {
							continue
						}
						if existingCalls[ownerName] != nil && existingCalls[ownerName][name] {
							continue
						}
						if name == ownerName {
							continue
						}
						c.Edges = append(c.Edges, types.Edge{
							From: ownerName,
							To:   name,
							Kind: types.EdgeCalls,
						})
						if existingCalls[ownerName] == nil {
							existingCalls[ownerName] = make(map[string]bool)
						}
						existingCalls[ownerName][name] = true
					}
				}
			}
		}
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}
	walk(root)
}

// CallableRefEdges detects identifiers passed as arguments to function
// calls (callback/function-reference patterns) and returns EdgeCalls edges.
func CallableRefEdges(
	node *sitter.Node,
	src []byte,
	callerName string,
	callNodeKinds []string,
	argListKinds []string,
	identKinds []string,
	knownCallables map[string]bool,
	alreadySeen map[string]bool,
) []types.Edge {
	argSet := make(map[string]bool, len(argListKinds))
	for _, k := range argListKinds {
		argSet[k] = true
	}
	identSet := make(map[string]bool, len(identKinds))
	for _, k := range identKinds {
		identSet[k] = true
	}
	callSet := make(map[string]bool, len(callNodeKinds))
	for _, k := range callNodeKinds {
		callSet[k] = true
	}

	var edges []types.Edge
	walkCallableRefs(node, src, callerName, callSet, argSet, identSet, knownCallables, alreadySeen, &edges)
	return edges
}

func walkCallableRefs(
	node *sitter.Node,
	src []byte,
	callerName string,
	callNodeKinds map[string]bool,
	argListKinds map[string]bool,
	identKinds map[string]bool,
	knownCallables map[string]bool,
	seen map[string]bool,
	edges *[]types.Edge,
) {
	if callNodeKinds[node.Kind()] {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if argListKinds[child.Kind()] {
				for j := range child.ChildCount() {
					arg := child.Child(j)
					if arg == nil {
						continue
					}
					if identKinds[arg.Kind()] {
						name := arg.Utf8Text(src)
						if name != "" && knownCallables[name] && !seen[name] {
							seen[name] = true
							*edges = append(*edges, types.Edge{
								From: callerName,
								To:   name,
								Kind: types.EdgeCalls,
							})
						}
					}
				}
			}
		}
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			walkCallableRefs(child, src, callerName, callNodeKinds, argListKinds, identKinds, knownCallables, seen, edges)
		}
	}
}
