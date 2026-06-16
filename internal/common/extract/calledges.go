// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// CallTargetFunc extracts the call target name from a call-expression node.
// Return "" to skip the node.
type CallTargetFunc func(node *sitter.Node, src []byte) string

// CallTargetResult holds the extracted call target name and whether it came
// from a qualified expression (e.g. obj.method()).
type CallTargetResult struct {
	Name      string
	Receiver  string // receiver/object name for qualified calls (e.g. "emitter" in emitter.emitAll())
	Qualified bool
}

// RichCallTargetFunc extracts a CallTargetResult from a call-expression node.
type RichCallTargetFunc func(node *sitter.Node, src []byte) CallTargetResult

// UnqualifiedCallTarget builds a CallTargetFunc that extracts only the last
// identifier from qualified expressions.
func UnqualifiedCallTarget(triggerKind string, simpleKinds, qualifiedKinds []string) CallTargetFunc {
	rich := UnqualifiedCallTargetRich(triggerKind, simpleKinds, qualifiedKinds)
	return func(node *sitter.Node, src []byte) string {
		return rich(node, src).Name
	}
}

// UnqualifiedCallTargetRich is like UnqualifiedCallTarget but returns a
// CallTargetResult that indicates whether the call was qualified.
func UnqualifiedCallTargetRich(triggerKind string, simpleKinds, qualifiedKinds []string) RichCallTargetFunc {
	simpleSet := make(map[string]bool, len(simpleKinds))
	for _, k := range simpleKinds {
		simpleSet[k] = true
	}
	qualifiedSet := make(map[string]bool, len(qualifiedKinds))
	for _, k := range qualifiedKinds {
		qualifiedSet[k] = true
	}
	return func(node *sitter.Node, src []byte) CallTargetResult {
		if node.Kind() != triggerKind {
			return CallTargetResult{}
		}
		if node.ChildCount() == 0 {
			return CallTargetResult{}
		}
		first := node.Child(0)
		if first == nil {
			return CallTargetResult{}
		}
		if simpleSet[first.Kind()] {
			return CallTargetResult{Name: first.Utf8Text(src)}
		}
		if qualifiedSet[first.Kind()] {
			if first.ChildCount() > 0 {
				receiver := first.Child(0)
				if receiver != nil && receiver.Kind() == triggerKind {
					return CallTargetResult{}
				}
			}
			// Extract receiver name (first named child of the qualified expression).
			var receiverName string
			if first.ChildCount() > 0 {
				rc := first.Child(0)
				if rc != nil && rc.IsNamed() {
					receiverName = rc.Utf8Text(src)
				}
			}
			return CallTargetResult{Name: lastNamedLeaf(first, src), Receiver: receiverName, Qualified: true}
		}
		return CallTargetResult{}
	}
}

// LastNamedLeaf walks a node tree and returns the text of the last named leaf.
func LastNamedLeaf(node *sitter.Node, src []byte) string {
	return lastNamedLeaf(node, src)
}

// lastNamedLeaf walks a node tree and returns the text of the last named leaf.
func lastNamedLeaf(node *sitter.Node, src []byte) string {
	if node.ChildCount() == 0 {
		return node.Utf8Text(src)
	}
	count := node.ChildCount()
	for i := count; i > 0; i-- {
		child := node.Child(i - 1)
		if child == nil || !child.IsNamed() {
			continue
		}
		return lastNamedLeaf(child, src)
	}
	return node.Utf8Text(src)
}

// FilterCallTarget wraps an existing CallTargetFunc and drops results that
// match any of the given names.
func FilterCallTarget(inner CallTargetFunc, skip ...string) CallTargetFunc {
	set := make(map[string]bool, len(skip))
	for _, s := range skip {
		set[s] = true
	}
	return func(node *sitter.Node, src []byte) string {
		name := inner(node, src)
		if set[name] {
			return ""
		}
		return name
	}
}

// CallEdges walks a node tree and returns EdgeCalls edges with dedup.
func CallEdges(node *sitter.Node, src []byte, callerName string, targetFn CallTargetFunc) []types.Edge {
	seen := make(map[string]bool)
	var edges []types.Edge
	walkCallEdges(node, src, callerName, seen, &edges, targetFn)
	return edges
}

func walkCallEdges(node *sitter.Node, src []byte, callerName string, seen map[string]bool, edges *[]types.Edge, targetFn CallTargetFunc) {
	if name := targetFn(node, src); name != "" && !seen[name] {
		seen[name] = true
		*edges = append(*edges, types.Edge{
			From: callerName,
			To:   name,
			Kind: types.EdgeCalls,
		})
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			walkCallEdges(child, src, callerName, seen, edges, targetFn)
		}
	}
}

// CallableRefConfig describes the language-specific node kinds needed for
// callable reference detection.
type CallableRefConfig struct {
	// CallNodeKinds are the tree-sitter node kinds for call expressions.
	CallNodeKinds []string
	// ArgListKinds are the node kinds for argument lists.
	ArgListKinds []string
	// IdentKinds are the node kinds for simple identifiers.
	IdentKinds []string
}
