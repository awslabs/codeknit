// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// DestructureConfig describes the language-specific node kinds for
// destructuring pattern extraction.
type DestructureConfig struct {
	ObjectPatternKind     string
	ArrayPatternKind      string
	RestKind              string
	AssignmentPatternKind string
	PairPatternKind       string
	IdentKinds            []string
}

// DestructuredNames extracts all identifier names from a destructuring
// pattern node (object_pattern or array_pattern) and emits them as value symbols.
func DestructuredNames(node *sitter.Node, src []byte, c *Collector, cfg *DestructureConfig, exported bool) {
	names := collectDestructuredNames(node, src, cfg)
	for _, name := range names {
		c.AddSymbol(&types.Symbol{
			Name:       name,
			Category:   types.CategoryValue,
			Kind:       "variable",
			Signature:  name,
			Properties: types.NewProps().SetBool("exported", exported).Map(),
			Span:       NodeSpan(node),
		})
	}
}

// collectDestructuredNames recursively collects all identifier names from
// a destructuring pattern.
func collectDestructuredNames(node *sitter.Node, src []byte, cfg *DestructureConfig) []string {
	if node == nil {
		return nil
	}

	identSet := make(map[string]bool, len(cfg.IdentKinds))
	for _, k := range cfg.IdentKinds {
		identSet[k] = true
	}

	var names []string
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}
		kind := n.Kind()

		// Direct identifier in pattern
		if identSet[kind] {
			name := n.Utf8Text(src)
			if name != "" {
				names = append(names, name)
			}
			return
		}

		// Object pattern: { a, b, c: d }
		if kind == cfg.ObjectPatternKind {
			for i := range n.ChildCount() {
				child := n.Child(i)
				if child != nil {
					walk(child)
				}
			}
			return
		}

		// Array pattern: [a, b, c]
		if kind == cfg.ArrayPatternKind {
			for i := range n.ChildCount() {
				child := n.Child(i)
				if child != nil {
					walk(child)
				}
			}
			return
		}

		// Pair pattern: { key: value } — extract the value side
		if kind == cfg.PairPatternKind {
			// The value is the last named child
			if n.ChildCount() >= 2 {
				for i := n.ChildCount(); i > 0; i-- {
					child := n.Child(i - 1)
					if child != nil && child.IsNamed() {
						walk(child)
						return
					}
				}
			}
			return
		}

		// Assignment pattern: { a = defaultVal } — extract the name side
		if kind == cfg.AssignmentPatternKind {
			if n.ChildCount() > 0 {
				first := n.Child(0)
				if first != nil {
					walk(first)
				}
			}
			return
		}

		// Rest pattern: { ...rest } or [...rest] — extract the identifier
		if kind == cfg.RestKind {
			for i := range n.ChildCount() {
				child := n.Child(i)
				if child != nil && identSet[child.Kind()] {
					walk(child)
					return
				}
			}
			return
		}

		// For any other node, recurse into children
		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}

	walk(node)
	return names
}
