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
	known := c.KnownCallablesAndImports()
	args, ids, calls := makeSet(cfg.ArgListKinds), makeSet(cfg.IdentKinds), makeSet(cfg.CallNodeKinds)
	// Passing a callable as a value establishes a reference, not an invocation.
	c.Edges = append(c.Edges, fileRelationships(root, src, c.Symbols, true, types.EdgeReferences,
		func(n *sitter.Node, _ *types.Symbol) string {
			parent := n.Parent()
			if !ids[n.Kind()] || parent == nil || !args[parent.Kind()] {
				return ""
			}
			call := parent.Parent()
			name := n.Utf8Text(src)
			if call != nil && calls[call.Kind()] && known[name] {
				return name
			}
			return ""
		})...)
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
