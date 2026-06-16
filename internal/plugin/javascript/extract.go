// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package javascript

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
		"lexical_declaration":            plugin.AdaptExported(extractLexicalDeclaration),
		"import_statement":               plugin.Adapt(extractImportStatement),
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

var bodyTokenMap = jsshared.JSBodyTokenMap
