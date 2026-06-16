// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package python

import (
	"codeknit/internal/plugin"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

var handlers = plugin.DispatchTable{
	"function_definition": plugin.Adapt(func(n *sitter.Node, src []byte, c *plugin.Collector) {
		extractFunction(n, src, c, false, "")
	}),
	"class_definition": plugin.Adapt(func(n *sitter.Node, src []byte, c *plugin.Collector) {
		extractClass(n, src, c)
	}),
	"expression_statement": plugin.Adapt(extractAssignment),
	"import_statement":     plugin.Adapt(extractImportStmt),
	"decorated_definition": func(n *sitter.Node, src []byte, c *plugin.Collector, _ plugin.HandlerContext) {
		for i := range n.ChildCount() {
			child := n.Child(i)
			if child == nil {
				continue
			}
			switch child.Kind() {
			case "function_definition":
				extractFunction(child, src, c, false, "", n)
			case "class_definition":
				extractClass(child, src, c, n)
			}
		}
	},
	"import_from_statement": plugin.Adapt(extractImportFromStmt),
}

var bodyTokenMap = map[string]byte{
	"if_statement":         plugin.FPIf,
	"elif_clause":          plugin.FPElseIf,
	"else_clause":          plugin.FPElse,
	"for_statement":        plugin.FPFor,
	"while_statement":      plugin.FPWhile,
	"return_statement":     plugin.FPReturn,
	"match_statement":      plugin.FPMatch,
	"case_clause":          plugin.FPCase,
	"break_statement":      plugin.FPBreak,
	"continue_statement":   plugin.FPCont,
	"try_statement":        plugin.FPTry,
	"except_clause":        plugin.FPCatch,
	"raise_statement":      plugin.FPThrow,
	"yield":                plugin.FPYield,
	"await":                plugin.FPAwait,
	"finally_clause":       plugin.FPDefer,
	"assignment":           plugin.FPAssign,
	"augmented_assignment": plugin.FPAssign,
	"call":                 plugin.FPCall,
	"attribute":            plugin.FPMember,
	"subscript":            plugin.FPIndex,
	"lambda":               plugin.FPLambda,
	"for_in_clause":        plugin.FPRange,
	"delete_statement":     plugin.FPDelete,
}
