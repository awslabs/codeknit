// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package typescript

import (
	"codeknit/internal/common/jsshared"
	"codeknit/internal/plugin"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"function_declaration":           plugin.AdaptExported(extractFunction),
		"generator_function_declaration": plugin.AdaptExported(extractFunction),
		"class_declaration":              plugin.AdaptExported(extractClass),
		"abstract_class_declaration":     plugin.AdaptExported(extractClass),
		"interface_declaration":          plugin.AdaptExported(extractInterface),
		"type_alias_declaration":         plugin.AdaptExported(extractTypeAlias),
		"enum_declaration":               plugin.AdaptExported(extractEnum),
		"lexical_declaration":            plugin.AdaptExported(extractLexicalDeclaration),
		"import_statement":               plugin.Adapt(extractImportStatement),
		"module":                         plugin.Adapt(extractNamespace),
		"internal_module":                plugin.Adapt(extractNamespace),
		"ambient_declaration":            plugin.Adapt(extractAmbientDeclaration),
		"namespace":                      plugin.Adapt(extractNamespace),
		"export_statement": func(n *sitter.Node, src []byte, c *plugin.Collector, _ plugin.HandlerContext) {
			for i := range n.ChildCount() {
				child := n.Child(i)
				if child != nil && child.Kind() != "export" && child.Kind() != "default" {
					handlers.Dispatch(child, src, c, plugin.HandlerContext{Exported: true})
				}
			}
		},
		"variable_declaration": plugin.AdaptExported(extractVariableDeclaration),
	}
}

var bodyTokenMap = func() map[string]byte {
	m := make(map[string]byte, len(jsshared.JSBodyTokenMap)+2)
	for k, v := range jsshared.JSBodyTokenMap {
		m[k] = v
	}
	m["type_assertion"] = plugin.FPCast
	m["as_expression"] = plugin.FPCast
	return m
}()
