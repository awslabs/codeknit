// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package golang

import "codeknit/internal/plugin"

// handlers maps tree-sitter node kinds directly to extraction functions.
// No Extractor interface, no wrapper methods — just the wiring.
var handlers = plugin.DispatchTable{
	"function_declaration": plugin.Adapt(extractFunction),
	"method_declaration":   plugin.Adapt(extractMethod),
	"type_declaration":     plugin.Adapt(extractTypeDecl),
	"package_clause":       plugin.Adapt(extractPackage),
	"var_declaration":      plugin.Adapt(extractVarDecl),
	"const_declaration":    plugin.Adapt(extractConstDecl),
	"import_declaration":   plugin.Adapt(extractImports),
}

// bodyTokenMap maps Go tree-sitter node kinds to universal fingerprint
// tokens for body-level semantic extraction.
var bodyTokenMap = map[string]byte{
	// Control flow
	"if_statement":       plugin.FPIf,
	"else_clause":        plugin.FPElse,
	"for_statement":      plugin.FPFor,
	"return_statement":   plugin.FPReturn,
	"switch_statement":   plugin.FPSwitch,
	"expression_case":    plugin.FPCase,
	"default_case":       plugin.FPCase,
	"break_statement":    plugin.FPBreak,
	"continue_statement": plugin.FPCont,
	// Concurrency
	"go_statement":     plugin.FPGo,
	"select_statement": plugin.FPSelect,
	"defer_statement":  plugin.FPDefer,
	// Operations
	"short_var_declaration": plugin.FPAssign,
	"assignment_statement":  plugin.FPAssign,
	"var_spec":              plugin.FPAssign,
	"call_expression":       plugin.FPCall,
	"selector_expression":   plugin.FPMember,
	"index_expression":      plugin.FPIndex,
	"composite_literal":     plugin.FPNew,
	"type_assertion":        plugin.FPCast,
	"func_literal":          plugin.FPLambda,
	"range_clause":          plugin.FPRange,
}
