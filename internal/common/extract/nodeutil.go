// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"sort"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// ChildByKind returns the first child with the given kind, or nil.
func ChildByKind(node *sitter.Node, kind string) *sitter.Node {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.Kind() == kind {
			return child
		}
	}
	return nil
}

// ChildText returns the text of the first child with the given kind, or "".
func ChildText(node *sitter.Node, kind string, src []byte) string {
	child := ChildByKind(node, kind)
	if child == nil {
		return ""
	}
	return child.Utf8Text(src)
}

// FindFirstError finds the first ERROR node in the tree and returns its row and column.
func FindFirstError(node *sitter.Node) (row, col uint) {
	if node.Kind() == "ERROR" || node.IsError() {
		pos := node.StartPosition()
		return pos.Row, pos.Column
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.HasError() {
			return FindFirstError(child)
		}
	}
	pos := node.StartPosition()
	return pos.Row, pos.Column
}

// NodeSpan returns the 1-indexed [startLine, endLine] span for a tree-sitter node.
func NodeSpan(node *sitter.Node) [2]int {
	return [2]int{
		int(node.StartPosition().Row) + 1, //nolint:gosec // row values are small
		int(node.EndPosition().Row) + 1,   //nolint:gosec // row values are small
	}
}

// BoolStr returns "true" or "false" as a string.
func BoolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// BuildFuncSignature builds a human-readable function signature string.
func BuildFuncSignature(name string, params []string, returnType string) string {
	sig := name + "(" + strings.Join(params, ", ") + ")"
	if returnType != "" {
		sig += " -> " + returnType
	}
	return sig
}

// HasChildKeyword checks if a node has a direct child whose text matches keyword.
func HasChildKeyword(node *sitter.Node, keyword string, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.Utf8Text(src) == keyword {
			return true
		}
	}
	return false
}

// SortedStringKeys returns the keys of any map[string]V sorted lexicographically.
func SortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// LastSepIndex returns the index of the last occurrence of sep in s, or -1.
func LastSepIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}

// makeSet builds a set from a slice of strings.
func makeSet(kinds []string) map[string]bool {
	s := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		s[k] = true
	}
	return s
}

// deepFindIdentifier searches for the first "identifier" node within maxDepth
// levels of a node tree.
func deepFindIdentifier(node *sitter.Node, src []byte, maxDepth int) string {
	if maxDepth <= 0 {
		return ""
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		ck := child.Kind()
		if ck == "compound_statement" || ck == "block" || ck == "statement_block" || ck == "body" {
			continue
		}
		if ck == "identifier" {
			return child.Utf8Text(src)
		}
		if child.IsNamed() {
			if id := deepFindIdentifier(child, src, maxDepth-1); id != "" {
				return id
			}
		}
	}
	return ""
}
