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
	// was qualified (e.g. obj.method()), preserving its receiver for resolution.
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
func DataflowHints(node *sitter.Node, src []byte, cfg *DataflowConfig, symbols []types.Symbol) []types.Edge {
	if len(cfg.AssignmentKinds) == 0 && len(cfg.ReturnKinds) == 0 {
		return nil
	}

	scopes := newSymbolScopes(node, src, symbols)
	assignSet := makeSet(cfg.AssignmentKinds)
	pairSet := makeSet(cfg.ObjectPairKinds)
	returnSet := makeSet(cfg.ReturnKinds)
	identSet := makeSet(cfg.IdentKinds)
	nameSet := makeSet(cfg.NameChildKinds)
	valueSet := makeSet(cfg.ValueChildKinds)

	var edges []types.Edge
	type flowKey struct {
		key    string
		offset uint
	}
	seen := make(map[flowKey]bool)
	add := func(key string, offset uint, edge types.Edge) {
		identity := flowKey{key, offset}
		if !seen[identity] {
			seen[identity] = true
			edge.SourceOffset = offset
			edges = append(edges, edge)
		}
	}

	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		kind := n.Kind()
		currentScope := ""
		var sourceOffset uint
		if owner := scopes.owner(n, true); owner != nil {
			currentScope = owner.EffectiveScopedName()
			sourceOffset = owner.DeclarationOffset()
		}

		if assignSet[kind] {
			lhs, rhs := extractAssignmentParts(n, src, nameSet, valueSet, identSet)
			if lhs != "" && !declaresCallable(n) {
				if rhs == "" {
					rhs = "<unknown>"
				}
				key := currentScope + ":" + lhs + "=" + rhs
				add(key, sourceOffset, types.Edge{From: lhs, To: rhs, Kind: types.EdgeAliases, Scope: currentScope})
			}
		}

		if pairSet[kind] {
			key, val := extractPairParts(n, src, identSet)
			if object := objectBinding(n, src); object != "" {
				key = object + "." + key
			}
			identity := currentScope + ":" + key + "=" + val
			if key != "" && val != "" {
				add(identity, sourceOffset, types.Edge{From: key, To: val, Kind: types.EdgeAliases, Scope: currentScope})
			}
		}

		if returnSet[kind] && currentScope != "" {
			val := extractReturnValue(n, src, identSet)
			if val != "" {
				add("ret:"+currentScope+":"+val, sourceOffset, types.Edge{From: currentScope, To: val, Kind: types.EdgeReturns})
			}
		}

		for i := range n.ChildCount() {
			child := n.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}
	walk(node)
	return edges
}

func declaresCallable(n *sitter.Node) bool {
	value := n.ChildByFieldName("value")
	if value == nil {
		value = n.ChildByFieldName("right")
	}
	if value == nil {
		return false
	}
	switch value.Kind() {
	case "arrow_function", "function_expression", "generator_function", "lambda", "func_literal":
		return true
	default:
		return false
	}
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
	return node.Utf8Text(src)
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

// objectBinding keeps keys from different object literals in separate namespaces.
func objectBinding(node *sitter.Node, src []byte) string {
	suffix := ""
	for n := node.Parent(); n != nil; n = n.Parent() {
		switch n.Kind() {
		case "variable_declarator", "assignment", "short_var_declaration", "assignment_expression":
			name := n.ChildByFieldName("name")
			if name == nil {
				name = n.ChildByFieldName("left")
			}
			if name != nil {
				return name.Utf8Text(src) + suffix
			}
			return ""
		case "pair":
			if key := n.ChildByFieldName("key"); key != nil {
				suffix = "." + key.Utf8Text(src) + suffix
			}
		case "function_declaration", "function_definition", "method_definition":
			return ""
		}
	}
	return ""
}
