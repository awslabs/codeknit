// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"strings"

	"codeknit/internal/common/types"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// FileTypeRefs records type references under their actual enclosing declaration.
func FileTypeRefs(root *sitter.Node, src []byte, typeRefKinds []string, symbols []types.Symbol) []types.Edge {
	kinds := makeSet(typeRefKinds)
	return fileRelationships(root, src, symbols, false, types.EdgeReferences, func(n *sitter.Node, _ *types.Symbol) string {
		if kinds[n.Kind()] {
			return n.Utf8Text(src)
		}
		return ""
	})
}

// FileCallEdges preserves qualified targets even when the declaration is in
// another file or unavailable. Resolution belongs to the planner.
func FileCallEdges(root *sitter.Node, src []byte, targetFn RichCallTargetFunc, symbols []types.Symbol) []types.Edge {
	if targetFn == nil {
		return nil
	}
	return fileRelationships(root, src, symbols, true, types.EdgeCalls, func(n *sitter.Node, owner *types.Symbol) string {
		result := targetFn(n, src)
		if result.Name == "" {
			return ""
		}
		if result.Qualified && result.Receiver != "" {
			receiver := result.Receiver
			if typeName := declaredReceiverType(n, src, receiver); typeName != "" {
				receiver = typeName
			}
			// self/this and explicit Go method receivers have an enclosing type.
			if receiver == "self" || receiver == "this" || receiver == "$this" || receiver == "Self" || isMethodReceiver(n, src, receiver) {
				if scope, _, ok := strings.Cut(owner.EffectiveScopedName(), "."); ok {
					if index := strings.LastIndex(owner.EffectiveScopedName(), "."); index >= 0 {
						scope = owner.EffectiveScopedName()[:index]
					}
					receiver = scope
				}
			}
			return receiver + "." + result.Name
		}
		return result.Name
	})
}

func isMethodReceiver(node *sitter.Node, src []byte, receiver string) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if params := n.ChildByFieldName("receiver"); params != nil {
			for i := range params.NamedChildCount() {
				if name := params.NamedChild(i).ChildByFieldName("name"); name != nil && name.Utf8Text(src) == receiver {
					return true
				}
			}
		}
	}
	return false
}

func fileRelationships(root *sitter.Node, src []byte, symbols []types.Symbol, callableOnly bool, kind types.EdgeKind, target func(*sitter.Node, *types.Symbol) string) []types.Edge {
	scopes := newSymbolScopes(root, src, symbols)
	type referenceKey struct {
		from, to string
		offset   uint
	}
	seen := make(map[referenceKey]bool)
	var edges []types.Edge
	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		if owner := scopes.owner(n, callableOnly); owner != nil {
			name := target(n, owner)
			key := referenceKey{owner.EffectiveScopedName(), name, owner.DeclarationOffset()}
			if name != "" && (kind != types.EdgeReferences || name != owner.Name) && !seen[key] {
				seen[key] = true
				edges = append(edges, types.Edge{From: key.from, To: name, Kind: kind, SourceOffset: owner.DeclarationOffset()})
			}
		}
		for i := range n.NamedChildCount() {
			walk(n.NamedChild(i))
		}
	}
	walk(root)
	return edges
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
