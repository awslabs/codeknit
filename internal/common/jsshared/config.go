// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package jsshared provides shared configuration and utilities for the
// JavaScript and TypeScript language plugins.
package jsshared

import (
	"codeknit/internal/common/extract"
	"codeknit/internal/common/types"
)

// JSBodyTokenMap is the shared body token map for JavaScript and TypeScript.
var JSBodyTokenMap = map[string]byte{
	"if_statement":                    types.FPIf,
	"else_clause":                     types.FPElse,
	"for_statement":                   types.FPFor,
	"for_in_statement":                types.FPFor,
	"while_statement":                 types.FPWhile,
	"do_statement":                    types.FPWhile,
	"return_statement":                types.FPReturn,
	"switch_statement":                types.FPSwitch,
	"switch_case":                     types.FPCase,
	"switch_default":                  types.FPCase,
	"break_statement":                 types.FPBreak,
	"continue_statement":              types.FPCont,
	"try_statement":                   types.FPTry,
	"catch_clause":                    types.FPCatch,
	"throw_statement":                 types.FPThrow,
	"yield_expression":                types.FPYield,
	"await_expression":                types.FPAwait,
	"finally_clause":                  types.FPDefer,
	"assignment_expression":           types.FPAssign,
	"variable_declarator":             types.FPAssign,
	"augmented_assignment_expression": types.FPAssign,
	"call_expression":                 types.FPCall,
	"member_expression":               types.FPMember,
	"subscript_expression":            types.FPIndex,
	"new_expression":                  types.FPNew,
	"arrow_function":                  types.FPLambda,
	"function_expression":             types.FPLambda,
	"delete_expression":               types.FPDelete,
}

// JSCallableRefConfig is the shared CallableRefConfig for JavaScript and TypeScript plugins.
var JSCallableRefConfig = extract.CallableRefConfig{
	CallNodeKinds: []string{"call_expression"},
	ArgListKinds:  []string{"arguments"},
	IdentKinds:    []string{"identifier"},
}

// JSDataflowConfig returns the shared DataflowConfig for JavaScript and TypeScript plugins.
func JSDataflowConfig(typeRefKinds []string, ct extract.CallTargetFunc, rct extract.RichCallTargetFunc) extract.DataflowConfig {
	return extract.DataflowConfig{
		AssignmentKinds: []string{"variable_declarator", "assignment_expression"},
		ObjectPairKinds: []string{"pair"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    typeRefKinds,
		CallTarget:      ct,
		RichCallTarget:  rct,
	}
}
