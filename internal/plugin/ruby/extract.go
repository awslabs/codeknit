// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ruby

import "codeknit/internal/plugin"

var handlers = plugin.DispatchTable{
	"method":              plugin.Adapt(extractFunction),
	"class":               plugin.Adapt(extractClass),
	"module":              plugin.Adapt(extractModule),
	"assignment":          plugin.Adapt(extractAssignment),
	"operator_assignment": plugin.Adapt(extractAssignment),
	"singleton_method":    plugin.Adapt(extractSingletonMethod),
	"call":                plugin.Adapt(extractRequire),
}

var bodyTokenMap = map[string]byte{
	"if":                  plugin.FPIf,
	"elsif":               plugin.FPIf,
	"unless":              plugin.FPIf,
	"else":                plugin.FPElse,
	"for":                 plugin.FPFor,
	"while":               plugin.FPWhile,
	"until":               plugin.FPWhile,
	"return":              plugin.FPReturn,
	"case":                plugin.FPSwitch,
	"when":                plugin.FPCase,
	"break":               plugin.FPBreak,
	"next":                plugin.FPCont,
	"begin":               plugin.FPTry,
	"rescue":              plugin.FPCatch,
	"raise":               plugin.FPThrow,
	"yield":               plugin.FPYield,
	"ensure":              plugin.FPDefer,
	"assignment":          plugin.FPAssign,
	"operator_assignment": plugin.FPAssign,
	"call":                plugin.FPCall,
	"method_call":         plugin.FPCall,
	"element_reference":   plugin.FPIndex,
	"lambda":              plugin.FPLambda,
	"block":               plugin.FPLambda,
	"do_block":            plugin.FPLambda,
}
