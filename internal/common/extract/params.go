// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import sitter "github.com/tree-sitter/go-tree-sitter"

// ParamConfig describes how to locate and destructure parameter nodes
// within a given tree-sitter grammar.
type ParamConfig struct {
	TypeExtractor func(node *sitter.Node, src []byte) string
	ParamListKind string
	NameKind      string
	ParamKinds    []string
}

// TypedParams walks a parameter list and returns "name: type" pairs.
func TypedParams(node *sitter.Node, src []byte, cfg ParamConfig) []string {
	paramList := ChildByKind(node, cfg.ParamListKind)
	if paramList == nil {
		return nil
	}
	kindSet := make(map[string]bool, len(cfg.ParamKinds))
	for _, k := range cfg.ParamKinds {
		kindSet[k] = true
	}
	var result []string
	for i := range paramList.ChildCount() {
		child := paramList.Child(i)
		if child == nil || !kindSet[child.Kind()] {
			continue
		}
		var name string
		if child.Kind() == cfg.NameKind {
			name = child.Utf8Text(src)
		} else {
			name = ChildText(child, cfg.NameKind, src)
		}
		if name == "" {
			continue
		}
		if cfg.TypeExtractor != nil {
			if typeName := cfg.TypeExtractor(child, src); typeName != "" {
				result = append(result, name+": "+typeName)
				continue
			}
		}
		result = append(result, name)
	}
	return result
}

// ReturnTypeByKinds scans direct children of node and returns the text of
// the first child whose kind is in typeKinds.
func ReturnTypeByKinds(node *sitter.Node, src []byte, typeKinds, stopKinds []string) string {
	typeSet := make(map[string]bool, len(typeKinds))
	for _, k := range typeKinds {
		typeSet[k] = true
	}
	stopSet := make(map[string]bool, len(stopKinds))
	for _, k := range stopKinds {
		stopSet[k] = true
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		kind := child.Kind()
		if stopSet[kind] {
			return ""
		}
		if typeSet[kind] {
			return child.Utf8Text(src)
		}
	}
	return ""
}

// ReturnTypeAfterToken scans children of node for an anonymous child whose
// text equals token (e.g. "->" or ":").
func ReturnTypeAfterToken(node *sitter.Node, src []byte, token string, validKinds []string) string {
	kindSet := make(map[string]bool, len(validKinds))
	for _, k := range validKinds {
		kindSet[k] = true
	}
	foundToken := false
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !foundToken {
			if child.Utf8Text(src) == token {
				foundToken = true
			}
			continue
		}
		if len(kindSet) == 0 {
			return child.Utf8Text(src)
		}
		if kindSet[child.Kind()] {
			return child.Utf8Text(src)
		}
		return ""
	}
	return ""
}

// FirstChildTextByKinds returns the text of the first child whose kind is
// in the given set.
func FirstChildTextByKinds(node *sitter.Node, src []byte, kinds []string) string {
	kindSet := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		kindSet[k] = true
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if kindSet[child.Kind()] {
			return child.Utf8Text(src)
		}
	}
	return ""
}
