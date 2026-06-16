// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scala

import "codeknit/internal/plugin"

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"package_clause":       plugin.Adapt(extractPackage),
		"class_definition":     plugin.AdaptParent(extractClass),
		"trait_definition":     plugin.AdaptParent(extractTrait),
		"function_definition":  plugin.AdaptParent(extractFunction),
		"function_declaration": plugin.AdaptParent(extractFunction),
		"var_definition":       plugin.AdaptParent(extractVar),
		"var_declaration":      plugin.AdaptParent(extractVar),
		"val_definition":       plugin.AdaptParent(extractVal),
		"val_declaration":      plugin.AdaptParent(extractVal),
		"type_definition":      plugin.AdaptParent(extractTypeDef),
		"enum_definition":      plugin.AdaptParent(extractEnum),
		"import_declaration":   plugin.Adapt(extractImport),
		"object_definition":    plugin.AdaptParent(extractObject),
	}
}

var bodyTokenMap = map[string]byte{
	"if_expression":         plugin.FPIf,
	"else_clause":           plugin.FPElse,
	"for_expression":        plugin.FPFor,
	"while_expression":      plugin.FPWhile,
	"return_expression":     plugin.FPReturn,
	"match_expression":      plugin.FPMatch,
	"case_clause":           plugin.FPCase,
	"try_expression":        plugin.FPTry,
	"catch_clause":          plugin.FPCatch,
	"throw_expression":      plugin.FPThrow,
	"yield":                 plugin.FPYield,
	"finally_clause":        plugin.FPDefer,
	"assignment_expression": plugin.FPAssign,
	"val_definition":        plugin.FPAssign,
	"var_definition":        plugin.FPAssign,
	"call_expression":       plugin.FPCall,
	"field_expression":      plugin.FPMember,
	"new_expression":        plugin.FPNew,
	"lambda_expression":     plugin.FPLambda,
}
