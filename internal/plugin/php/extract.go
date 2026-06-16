// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package php

import "codeknit/internal/plugin"

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"namespace_definition":      plugin.Adapt(extractNamespace),
		"function_definition":       plugin.AdaptParent(extractFunction),
		"class_declaration":         plugin.AdaptParent(extractClass),
		"interface_declaration":     plugin.AdaptParent(extractInterface),
		"trait_declaration":         plugin.AdaptParent(extractTrait),
		"enum_declaration":          plugin.AdaptParent(extractEnum),
		"method_declaration":        plugin.AdaptParent(extractMethod),
		"property_declaration":      plugin.AdaptParent(extractProperty),
		"const_declaration":         plugin.AdaptParent(extractConst),
		"namespace_use_declaration": plugin.Adapt(extractUseDeclaration),
	}
}

var bodyTokenMap = map[string]byte{
	"if_statement":                           plugin.FPIf,
	"else_if_clause":                         plugin.FPIf,
	"else_clause":                            plugin.FPElse,
	"for_statement":                          plugin.FPFor,
	"foreach_statement":                      plugin.FPFor,
	"while_statement":                        plugin.FPWhile,
	"do_statement":                           plugin.FPWhile,
	"return_statement":                       plugin.FPReturn,
	"switch_statement":                       plugin.FPSwitch,
	"case_statement":                         plugin.FPCase,
	"default_statement":                      plugin.FPCase,
	"break_statement":                        plugin.FPBreak,
	"continue_statement":                     plugin.FPCont,
	"try_statement":                          plugin.FPTry,
	"catch_clause":                           plugin.FPCatch,
	"throw_expression":                       plugin.FPThrow,
	"yield_expression":                       plugin.FPYield,
	"finally_clause":                         plugin.FPDefer,
	"assignment_expression":                  plugin.FPAssign,
	"augmented_assignment_expression":        plugin.FPAssign,
	"function_call_expression":               plugin.FPCall,
	"member_call_expression":                 plugin.FPCall,
	"member_access_expression":               plugin.FPMember,
	"subscript_expression":                   plugin.FPIndex,
	"object_creation_expression":             plugin.FPNew,
	"cast_expression":                        plugin.FPCast,
	"anonymous_function_creation_expression": plugin.FPLambda,
	"arrow_function":                         plugin.FPLambda,
	"match_expression":                       plugin.FPMatch,
}
