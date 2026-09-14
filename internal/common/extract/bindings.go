// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"strings"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// parameterNames keeps bindings that can shadow module declarations. An
// unknown runtime value must not resolve to a same-named global callable.
func parameterNames(declaration *sitter.Node, src []byte) map[string]string {
	result := make(map[string]string)
	params := declaration.ChildByFieldName("parameters")
	if params == nil {
		return result
	}
	for i := range params.NamedChildCount() {
		param := params.NamedChild(i)
		name := param.ChildByFieldName("name")
		if name == nil {
			name = param.ChildByFieldName("pattern")
		}
		if name == nil && param.Kind() == "identifier" {
			name = param
		}
		if name == nil {
			continue
		}
		typeName := ""
		if typ := param.ChildByFieldName("type"); typ != nil {
			typeName = strings.TrimSpace(strings.TrimPrefix(typ.Utf8Text(src), ":"))
			typeName = strings.TrimLeft(typeName, "*&")
		}
		result[name.Utf8Text(src)] = typeName
		// Go groups names under one parameter declaration: a, b func().
		// Each name shadows an outer binding, not only the first name field.
		if param.Kind() == "parameter_declaration" {
			for j := range param.NamedChildCount() {
				child := param.NamedChild(j)
				if child.Kind() == "identifier" {
					result[child.Utf8Text(src)] = typeName
				}
			}
		}
	}
	return result
}

func parameterEdges(root *sitter.Node, src []byte, symbols []types.Symbol) []types.Edge {
	scopes := newSymbolScopes(root, src, symbols)
	var result []types.Edge
	type bindingKey struct {
		scope, name string
		offset      uint
	}
	seen := make(map[bindingKey]bool)
	for _, scope := range scopes.declarations {
		if scope.sym.Category != types.CategoryCallable {
			continue
		}
		for name := range parameterNames(scope.node, src) {
			key := bindingKey{scope.sym.EffectiveScopedName(), name, scope.sym.DeclarationOffset()}
			if !seen[key] {
				seen[key] = true
				result = append(result, types.Edge{
					From: name, To: "<unknown>", Scope: key.scope, Kind: types.EdgeAliases, SourceOffset: key.offset,
				})
			}
		}
	}
	return result
}

func declaredReceiverType(call *sitter.Node, src []byte, receiver string) string {
	for n := call.Parent(); n != nil; n = n.Parent() {
		if n.ChildByFieldName("parameters") != nil {
			return parameterNames(n, src)[receiver]
		}
	}
	return ""
}
