// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// FileTypeRefs walks the entire AST and emits EdgeReferences edges for
// type identifiers, attributed to the enclosing type or function via scope
// tracking.
func FileTypeRefs(root *sitter.Node, src []byte, typeRefKinds []string, symbols []types.Symbol) []types.Edge {
	if len(typeRefKinds) == 0 {
		return nil
	}

	typeRefSet := makeSet(typeRefKinds)

	knownNames := make(map[string]bool, len(symbols))
	for i := range symbols {
		knownNames[symbols[i].Name] = true
	}

	type scopeSpan struct {
		name     string
		startRow int
		endRow   int
	}
	var scopes []scopeSpan
	for i := range symbols {
		sym := &symbols[i]
		if sym.Category == types.CategoryCallable || sym.Category == types.CategoryType {
			scopes = append(scopes, scopeSpan{
				name:     sym.EffectiveScopedName(),
				startRow: sym.Span[0],
				endRow:   sym.Span[1],
			})
		}
	}

	var edges []types.Edge
	seen := make(map[string]bool)

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if typeRefSet[n.Kind()] {
			typeName := n.Utf8Text(src)
			if typeName == "" {
				goto recurse
			}

			{
				row := int(n.StartPosition().Row) + 1 //nolint:gosec // Row is a small line number, no overflow risk
				var ownerName string
				for _, s := range scopes {
					if row >= s.startRow && row <= s.endRow {
						ownerName = s.name
					}
				}
				if ownerName == "" || ownerName == typeName {
					goto recurse
				}

				key := ownerName + ":" + typeName
				if !seen[key] {
					seen[key] = true
					edges = append(edges, types.Edge{
						From: ownerName,
						To:   typeName,
						Kind: types.EdgeReferences,
					})
				}
			}
		}
	recurse:
		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}
	walk(root)
	return edges
}

// FileCallEdges walks the entire AST and emits EdgeCalls edges for
// call expressions, attributed to the enclosing function via span-based scope
// tracking. For qualified calls (e.g. obj.method()), edges are only emitted
// when the target name matches a known symbol in the file, filtering out
// language built-in methods like .filter, .map, .forEach, etc.
func FileCallEdges(root *sitter.Node, src []byte, targetFn RichCallTargetFunc, symbols []types.Symbol) []types.Edge {
	if targetFn == nil {
		return nil
	}

	// Build a set of all known symbol names (both Name and ScopedName) so we
	// can gate qualified calls against it.
	knownNames := make(map[string]bool, len(symbols))
	for i := range symbols {
		knownNames[symbols[i].Name] = true
		if sn := symbols[i].ScopedName; sn != "" {
			knownNames[sn] = true
		}
	}

	type scopeSpan struct {
		name     string
		startRow int
		endRow   int
	}
	var scopes []scopeSpan
	for i := range symbols {
		sym := &symbols[i]
		if sym.Category == types.CategoryCallable {
			scopes = append(scopes, scopeSpan{
				name:     sym.EffectiveScopedName(),
				startRow: sym.Span[0],
				endRow:   sym.Span[1],
			})
		}
	}

	var edges []types.Edge
	seen := make(map[string]bool)

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if result := targetFn(n, src); result.Name != "" {
			// For qualified calls (obj.method()), only emit the edge if the
			// method name matches a symbol defined in this file.
			if result.Qualified && !knownNames[result.Name] {
				goto recurse
			}
			row := int(n.StartPosition().Row) + 1 //nolint:gosec // Row is a small line number, no overflow risk
			var ownerName string
			for _, s := range scopes {
				if row >= s.startRow && row <= s.endRow {
					ownerName = s.name
				}
			}
			if ownerName != "" {
				key := ownerName + ":" + result.Name
				if !seen[key] {
					seen[key] = true
					edges = append(edges, types.Edge{
						From: ownerName,
						To:   result.Name,
						Kind: types.EdgeCalls,
					})
				}
			}
		}
	recurse:
		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}
	walk(root)
	return edges
}

// FileCallEdgesWithAliases is like FileCallEdges but also considers alias
// source names (from EdgeAliases edges) as known names. This allows qualified
// calls like obj.prop() to produce edges when prop is an alias key that maps
// to a callable (e.g., const routes = { create: handleCreate }).
func FileCallEdgesWithAliases(root *sitter.Node, src []byte, targetFn RichCallTargetFunc, symbols []types.Symbol, edges []types.Edge) []types.Edge {
	if targetFn == nil {
		return nil
	}

	knownNames := make(map[string]bool, len(symbols))
	for i := range symbols {
		knownNames[symbols[i].Name] = true
		if sn := symbols[i].ScopedName; sn != "" {
			knownNames[sn] = true
		}
	}

	// Add alias source names as known names so that qualified calls like
	// obj.prop() produce edges when prop is an alias key pointing to a callable.
	for _, e := range edges {
		if e.Kind == types.EdgeAliases {
			knownNames[e.From] = true
		}
	}

	type scopeSpan struct {
		name     string
		startRow int
		endRow   int
	}
	var scopes []scopeSpan
	for i := range symbols {
		sym := &symbols[i]
		if sym.Category == types.CategoryCallable {
			scopes = append(scopes, scopeSpan{
				name:     sym.EffectiveScopedName(),
				startRow: sym.Span[0],
				endRow:   sym.Span[1],
			})
		}
	}

	var result []types.Edge
	seen := make(map[string]bool)

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if r := targetFn(n, src); r.Name != "" {
			// For qualified calls (obj.method()), only emit the edge if the
			// method name matches a known symbol or alias key in this file.
			// This filters out language built-ins like .filter/.map/.forEach.
			if r.Qualified && !knownNames[r.Name] {
				goto recurse
			}
			row := int(n.StartPosition().Row) + 1 //nolint:gosec // Row is a small line number, no overflow risk
			var ownerName string
			for _, s := range scopes {
				if row >= s.startRow && row <= s.endRow {
					ownerName = s.name
				}
			}
			if ownerName != "" {
				key := ownerName + ":" + r.Name
				if !seen[key] {
					seen[key] = true
					result = append(result, types.Edge{
						From: ownerName,
						To:   r.Name,
						Kind: types.EdgeCalls,
					})
				}
			}
		}
	recurse:
		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}
	walk(root)
	return result
}

// TypeRefEdges walks a node tree and returns EdgeReferences edges for
// type identifiers found in function signatures and bodies.
func TypeRefEdges(node *sitter.Node, src []byte, callerName string, typeNodeKinds []string) []types.Edge {
	kindSet := make(map[string]bool, len(typeNodeKinds))
	for _, k := range typeNodeKinds {
		kindSet[k] = true
	}
	seen := make(map[string]bool)
	var edges []types.Edge
	walkTypeRefs(node, src, callerName, kindSet, seen, &edges)
	return edges
}

func walkTypeRefs(node *sitter.Node, src []byte, callerName string, kindSet, seen map[string]bool, edges *[]types.Edge) {
	if kindSet[node.Kind()] {
		name := node.Utf8Text(src)
		if name != "" && name != callerName && !seen[name] {
			seen[name] = true
			*edges = append(*edges, types.Edge{
				From: callerName,
				To:   name,
				Kind: types.EdgeReferences,
			})
		}
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			walkTypeRefs(child, src, callerName, kindSet, seen, edges)
		}
	}
}
