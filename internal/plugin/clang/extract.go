// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clang

import "codeknit/internal/plugin"

var handlers = plugin.DispatchTable{
	"function_definition":  plugin.Adapt(extractFunction),
	"struct_specifier":     plugin.Adapt(extractStruct),
	"enum_specifier":       plugin.Adapt(extractEnum),
	"preproc_include":      plugin.Adapt(extractInclude),
	"declaration":          plugin.Adapt(extractDeclaration),
	"union_specifier":      plugin.Adapt(extractUnion),
	"type_definition":      plugin.Adapt(extractTypedef),
	"preproc_def":          plugin.Adapt(extractPreprocDef),
	"preproc_function_def": plugin.Adapt(extractPreprocFunctionDef),
}

var bodyTokenMap = map[string]byte{
	"if_statement":          plugin.FPIf,
	"else_clause":           plugin.FPElse,
	"for_statement":         plugin.FPFor,
	"while_statement":       plugin.FPWhile,
	"do_statement":          plugin.FPWhile,
	"return_statement":      plugin.FPReturn,
	"switch_statement":      plugin.FPSwitch,
	"case_statement":        plugin.FPCase,
	"break_statement":       plugin.FPBreak,
	"continue_statement":    plugin.FPCont,
	"assignment_expression": plugin.FPAssign,
	"init_declarator":       plugin.FPAssign,
	"call_expression":       plugin.FPCall,
	"field_expression":      plugin.FPMember,
	"subscript_expression":  plugin.FPIndex,
	"cast_expression":       plugin.FPCast,
}
