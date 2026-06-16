// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csharp

import "codeknit/internal/plugin"

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"namespace_declaration":             plugin.Adapt(extractNamespace),
		"class_declaration":                 plugin.AdaptParent(extractClass),
		"record_declaration":                plugin.AdaptParent(extractRecord),
		"record_struct_declaration":         plugin.AdaptParent(extractRecord),
		"interface_declaration":             plugin.AdaptParent(extractInterface),
		"struct_declaration":                plugin.AdaptParent(extractStruct),
		"enum_declaration":                  plugin.AdaptParent(extractEnum),
		"method_declaration":                plugin.AdaptParent(extractMethod),
		"constructor_declaration":           plugin.AdaptParent(extractConstructor),
		"field_declaration":                 plugin.AdaptParent(extractField),
		"using_directive":                   plugin.Adapt(extractUsingDirective),
		"property_declaration":              plugin.AdaptParent(extractProperty),
		"delegate_declaration":              plugin.AdaptParent(extractDelegate),
		"event_field_declaration":           plugin.AdaptParent(extractEvent),
		"operator_declaration":              plugin.AdaptParent(extractOperator),
		"file_scoped_namespace_declaration": plugin.Adapt(extractNamespace),
	}
}

var bodyTokenMap = map[string]byte{
	"if_statement":               plugin.FPIf,
	"else_clause":                plugin.FPElse,
	"for_statement":              plugin.FPFor,
	"for_each_statement":         plugin.FPFor,
	"while_statement":            plugin.FPWhile,
	"do_statement":               plugin.FPWhile,
	"return_statement":           plugin.FPReturn,
	"switch_statement":           plugin.FPSwitch,
	"switch_expression":          plugin.FPSwitch,
	"switch_section":             plugin.FPCase,
	"switch_expression_arm":      plugin.FPCase,
	"break_statement":            plugin.FPBreak,
	"continue_statement":         plugin.FPCont,
	"try_statement":              plugin.FPTry,
	"catch_clause":               plugin.FPCatch,
	"throw_statement":            plugin.FPThrow,
	"throw_expression":           plugin.FPThrow,
	"yield_statement":            plugin.FPYield,
	"await_expression":           plugin.FPAwait,
	"finally_clause":             plugin.FPDefer,
	"assignment_expression":      plugin.FPAssign,
	"variable_declaration":       plugin.FPAssign,
	"invocation_expression":      plugin.FPCall,
	"member_access_expression":   plugin.FPMember,
	"element_access_expression":  plugin.FPIndex,
	"object_creation_expression": plugin.FPNew,
	"cast_expression":            plugin.FPCast,
	"lambda_expression":          plugin.FPLambda,
}
