// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rust

import "codeknit/internal/plugin"

var handlers = plugin.DispatchTable{
	"function_item":    plugin.Adapt(extractFunction),
	"struct_item":      plugin.Adapt(extractStruct),
	"enum_item":        plugin.Adapt(extractEnum),
	"trait_item":       plugin.Adapt(extractTrait),
	"type_item":        plugin.Adapt(extractTypeAlias),
	"mod_item":         plugin.Adapt(extractMod),
	"const_item":       plugin.Adapt(extractConst),
	"use_declaration":  plugin.Adapt(extractUseDecl),
	"impl_item":        plugin.Adapt(extractImpl),
	"static_item":      plugin.Adapt(extractStatic),
	"macro_definition": plugin.Adapt(extractMacro),
}

var bodyTokenMap = map[string]byte{
	"if_expression":         plugin.FPIf,
	"else_clause":           plugin.FPElse,
	"for_expression":        plugin.FPFor,
	"while_expression":      plugin.FPWhile,
	"loop_expression":       plugin.FPWhile,
	"return_expression":     plugin.FPReturn,
	"match_expression":      plugin.FPMatch,
	"match_arm":             plugin.FPCase,
	"break_expression":      plugin.FPBreak,
	"continue_expression":   plugin.FPCont,
	"try_expression":        plugin.FPTry,
	"yield_expression":      plugin.FPYield,
	"await_expression":      plugin.FPAwait,
	"assignment_expression": plugin.FPAssign,
	"let_declaration":       plugin.FPAssign,
	"call_expression":       plugin.FPCall,
	"field_expression":      plugin.FPMember,
	"index_expression":      plugin.FPIndex,
	"struct_expression":     plugin.FPNew,
	"type_cast_expression":  plugin.FPCast,
	"closure_expression":    plugin.FPLambda,
	"range_expression":      plugin.FPRange,
}
