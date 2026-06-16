// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// DataflowConfig describes the language-specific node kinds needed for
// dataflow hint extraction and file-level type reference extraction.
type DataflowConfig struct {
	// CallTarget is the language-specific function that extracts call target
	// names from call expression nodes.
	CallTarget CallTargetFunc
	// RichCallTarget is like CallTarget but also indicates whether the call
	// was qualified (e.g. obj.method()). When set, FileCallEdges uses it to
	// gate qualified calls against known symbols, filtering out language
	// built-in methods.
	RichCallTarget RichCallTargetFunc
	// AssignmentKinds are node kinds for assignment/declaration statements.
	AssignmentKinds []string
	// ObjectPairKinds are node kinds for key-value pairs in object/map literals.
	ObjectPairKinds []string
	// ReturnKinds are node kinds for return statements.
	ReturnKinds []string
	// IdentKinds are node kinds for simple identifiers.
	IdentKinds []string
	// NameChildKinds are node kinds for the "name" side of an assignment.
	NameChildKinds []string
	// ValueChildKinds are node kinds for the "value" side of an assignment.
	ValueChildKinds []string
	// TypeRefKinds are tree-sitter node kinds for type references.
	TypeRefKinds []string
}

// DataflowHints walks a function body and emits EdgeAliases and
// EdgeReturns edges for dataflow tracking.
func DataflowHints(node *sitter.Node, src []byte, scopeName string, cfg *DataflowConfig, knownCallables map[string]bool) []types.Edge {
	if len(cfg.AssignmentKinds) == 0 && len(cfg.ReturnKinds) == 0 {
		return nil
	}

	assignSet := makeSet(cfg.AssignmentKinds)
	pairSet := makeSet(cfg.ObjectPairKinds)
	returnSet := makeSet(cfg.ReturnKinds)
	identSet := makeSet(cfg.IdentKinds)
	nameSet := makeSet(cfg.NameChildKinds)
	valueSet := makeSet(cfg.ValueChildKinds)

	funcKinds := map[string]bool{
		"function_declaration":    true,
		"method_declaration":      true,
		"method_definition":       true,
		"function_definition":     true,
		"constructor_declaration": true,
		"function_item":           true,
		"function_signature_item": true,
		"method":                  true,
		"singleton_method":        true,
		"function_def":            true,
	}

	var edges []types.Edge
	seen := make(map[string]bool)

	var walk func(n *sitter.Node, currentScope string)
	walk = func(n *sitter.Node, currentScope string) {
		kind := n.Kind()

		if funcKinds[kind] {
			if nameNode := ChildByKind(n, "identifier"); nameNode != nil {
				currentScope = nameNode.Utf8Text(src)
			} else if nameNode := ChildByKind(n, "name"); nameNode != nil {
				currentScope = nameNode.Utf8Text(src)
			} else {
				if id := deepFindIdentifier(n, src, 4); id != "" {
					currentScope = id
				}
			}
		}

		if assignSet[kind] {
			lhs, rhs := extractAssignmentParts(n, src, nameSet, valueSet, identSet)
			if lhs != "" && rhs != "" && knownCallables[rhs] && !seen[lhs+"="+rhs] {
				seen[lhs+"="+rhs] = true
				edges = append(edges, types.Edge{From: lhs, To: rhs, Kind: types.EdgeAliases})
			}
		}

		if pairSet[kind] {
			key, val := extractPairParts(n, src, identSet)
			if key != "" && val != "" && knownCallables[val] && !seen[key+"="+val] {
				seen[key+"="+val] = true
				edges = append(edges, types.Edge{From: key, To: val, Kind: types.EdgeAliases})
			}
		}

		if returnSet[kind] && currentScope != "" {
			val := extractReturnValue(n, src, identSet)
			if val != "" && knownCallables[val] && !seen["ret:"+currentScope+":"+val] {
				seen["ret:"+currentScope+":"+val] = true
				edges = append(edges, types.Edge{From: currentScope, To: val, Kind: types.EdgeReturns})
			}
		}

		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child, currentScope)
			}
		}
	}
	walk(node, scopeName)
	return edges
}

// qualifiedExprKinds are tree-sitter node kinds that represent qualified
// expressions across all languages.
var qualifiedExprKinds = map[string]bool{
	"selector_expression":      true,
	"member_expression":        true,
	"attribute":                true,
	"field_access":             true,
	"member_access_expression": true,
	"field_expression":         true,
	"scoped_identifier":        true,
	"qualified_identifier":     true,
}

func extractQualifiedName(node *sitter.Node, src []byte) string {
	return lastNamedLeaf(node, src)
}

func extractValueName(node *sitter.Node, src []byte, identKinds map[string]bool) string {
	if identKinds[node.Kind()] {
		return node.Utf8Text(src)
	}
	if qualifiedExprKinds[node.Kind()] {
		return extractQualifiedName(node, src)
	}
	return ""
}

func extractAssignmentParts(node *sitter.Node, src []byte, nameKinds, valueKinds, identKinds map[string]bool) (lhs, rhs string) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		ck := child.Kind()
		switch {
		case lhs == "" && nameKinds[ck]:
			lhs = child.Utf8Text(src)
		case lhs != "" && rhs == "":
			if name := extractValueName(child, src, identKinds); name != "" {
				rhs = name
			} else if valueKinds[ck] {
				if child.ChildCount() == 1 {
					inner := child.Child(0)
					if inner != nil {
						if name := extractValueName(inner, src, identKinds); name != "" {
							rhs = name
						}
					}
				}
			}
		}
	}
	return lhs, rhs
}

func extractPairParts(node *sitter.Node, src []byte, identKinds map[string]bool) (key, val string) {
	if node.ChildCount() < 2 {
		return "", ""
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		if key == "" {
			key = child.Utf8Text(src)
		} else if val == "" {
			if name := extractValueName(child, src, identKinds); name != "" {
				val = name
			}
			break
		}
	}
	return key, val
}

func extractReturnValue(node *sitter.Node, src []byte, identKinds map[string]bool) string {
	callKinds := map[string]bool{
		"call_expression":          true,
		"method_invocation":        true,
		"invocation_expression":    true,
		"call":                     true,
		"function_call_expression": true,
		"member_call_expression":   true,
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if callKinds[child.Kind()] {
			continue
		}
		if name := extractValueName(child, src, identKinds); name != "" {
			return name
		}
		if child.IsNamed() {
			for j := range child.ChildCount() {
				grandchild := child.Child(j)
				if grandchild == nil || callKinds[grandchild.Kind()] {
					continue
				}
				if name := extractValueName(grandchild, src, identKinds); name != "" {
					return name
				}
			}
		}
	}
	return ""
}
