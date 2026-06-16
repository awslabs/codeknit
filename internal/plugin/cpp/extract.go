// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cpp

import "codeknit/internal/plugin"

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"function_definition":  plugin.AdaptParent(extractFunction),
		"class_specifier":      plugin.AdaptParent(extractClass),
		"struct_specifier":     plugin.AdaptParent(extractStruct),
		"enum_specifier":       plugin.AdaptParent(extractEnum),
		"namespace_definition": plugin.Adapt(extractNamespace),
		"preproc_include":      plugin.Adapt(extractInclude),
		"template_declaration": plugin.AdaptParent(extractTemplate),
		"declaration":          plugin.AdaptParent(extractDeclaration),
		"field_declaration":    plugin.AdaptParent(extractFieldDeclaration),
	}
}

var bodyTokenMap = map[string]byte{
	"if_statement":            plugin.FPIf,
	"else_clause":             plugin.FPElse,
	"for_statement":           plugin.FPFor,
	"for_range_loop":          plugin.FPFor,
	"while_statement":         plugin.FPWhile,
	"do_statement":            plugin.FPWhile,
	"return_statement":        plugin.FPReturn,
	"switch_statement":        plugin.FPSwitch,
	"case_statement":          plugin.FPCase,
	"break_statement":         plugin.FPBreak,
	"continue_statement":      plugin.FPCont,
	"try_statement":           plugin.FPTry,
	"catch_clause":            plugin.FPCatch,
	"throw_statement":         plugin.FPThrow,
	"co_yield_statement":      plugin.FPYield,
	"co_await_expression":     plugin.FPAwait,
	"assignment_expression":   plugin.FPAssign,
	"init_declarator":         plugin.FPAssign,
	"call_expression":         plugin.FPCall,
	"field_expression":        plugin.FPMember,
	"subscript_expression":    plugin.FPIndex,
	"new_expression":          plugin.FPNew,
	"cast_expression":         plugin.FPCast,
	"static_cast_expression":  plugin.FPCast,
	"dynamic_cast_expression": plugin.FPCast,
	"lambda_expression":       plugin.FPLambda,
	"delete_expression":       plugin.FPDelete,
}
