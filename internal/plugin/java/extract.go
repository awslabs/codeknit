// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package java

import "codeknit/internal/plugin"

var handlers plugin.DispatchTable

func init() {
	handlers = plugin.DispatchTable{
		"package_declaration":                 plugin.Adapt(extractPackage),
		"class_declaration":                   plugin.AdaptParent(extractClass),
		"record_declaration":                  plugin.AdaptParent(extractRecord),
		"interface_declaration":               plugin.AdaptParent(extractInterface),
		"enum_declaration":                    plugin.AdaptParent(extractEnum),
		"method_declaration":                  plugin.AdaptParent(extractMethod),
		"constructor_declaration":             plugin.AdaptParent(extractConstructor),
		"field_declaration":                   plugin.AdaptParent(extractField),
		"import_declaration":                  plugin.Adapt(extractImport),
		"annotation_type_declaration":         plugin.AdaptParent(extractAnnotationType),
		"annotation_type_element_declaration": plugin.AdaptParent(extractMethod),
	}
}

var bodyTokenMap = map[string]byte{
	"if_statement":                 plugin.FPIf,
	"else_clause":                  plugin.FPElse,
	"for_statement":                plugin.FPFor,
	"enhanced_for_statement":       plugin.FPFor,
	"while_statement":              plugin.FPWhile,
	"do_statement":                 plugin.FPWhile,
	"return_statement":             plugin.FPReturn,
	"switch_expression":            plugin.FPSwitch,
	"switch_block_statement_group": plugin.FPCase,
	"switch_rule":                  plugin.FPCase,
	"break_statement":              plugin.FPBreak,
	"continue_statement":           plugin.FPCont,
	"try_statement":                plugin.FPTry,
	"catch_clause":                 plugin.FPCatch,
	"throw_statement":              plugin.FPThrow,
	"yield_statement":              plugin.FPYield,
	"finally_clause":               plugin.FPDefer,
	"assignment_expression":        plugin.FPAssign,
	"variable_declarator":          plugin.FPAssign,
	"method_invocation":            plugin.FPCall,
	"field_access":                 plugin.FPMember,
	"array_access":                 plugin.FPIndex,
	"object_creation_expression":   plugin.FPNew,
	"cast_expression":              plugin.FPCast,
	"lambda_expression":            plugin.FPLambda,
}
