// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package javascript

import (
	"codeknit/internal/common/extract"
	"codeknit/internal/common/jsshared"
	"codeknit/internal/plugin"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Aliases for shared helpers — keeps call sites unchanged.
var (
	childByKind        = plugin.ChildByKind
	childText          = plugin.ChildText
	nodeSpan           = plugin.NodeSpan
	buildFuncSignature = plugin.BuildFuncSignature
	hasChildKeyword    = plugin.HasChildKeyword
)

// jsDestructureConfig describes how JavaScript destructuring patterns are structured.
var jsDestructureConfig = extract.DestructureConfig{
	ObjectPatternKind:     "object_pattern",
	ArrayPatternKind:      "array_pattern",
	IdentKinds:            []string{"identifier", "shorthand_property_identifier_pattern"},
	RestKind:              "rest_pattern",
	AssignmentPatternKind: "assignment_pattern",
	PairPatternKind:       "pair_pattern",
}

// jsParamConfig describes how JavaScript parameter nodes are structured.
// JavaScript has no type annotations, so TypeExtractor is nil.
var jsParamConfig = plugin.ParamConfig{
	ParamListKind: "formal_parameters",
	ParamKinds:    []string{"identifier"},
	NameKind:      "identifier",
	TypeExtractor: nil,
}

// extractParamStrings extracts function parameter names as strings.
// JavaScript has no type annotations, so we just return the names.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	return plugin.ExtractTypedParams(node, src, jsParamConfig)
}

// extractHeritageNames extracts base class names from extends clauses as strings.
func extractHeritageNames(node *sitter.Node, src []byte) []string {
	heritage := childByKind(node, "class_heritage")
	if heritage == nil {
		return nil
	}
	var result []string
	for i := range heritage.ChildCount() {
		child := heritage.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "identifier" {
			result = append(result, child.Utf8Text(src))
		}
	}
	return result
}

// callTarget is the CallTargetFunc for JavaScript.
// Matches call_expression nodes whose first child is an identifier or member_expression.
// For member_expression (e.g. console.log), extracts only the method name ("log").
var callTarget = plugin.UnqualifiedCallTarget("call_expression",
	[]string{"identifier"},
	[]string{"member_expression"},
)

// richCallTarget is the RichCallTargetFunc for JavaScript.
// Like callTarget but also indicates whether the call was qualified.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call_expression",
	[]string{"identifier"},
	[]string{"member_expression"},
)

// extractFunction extracts a function_declaration node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	isAsync := hasChildKeyword(node, "async", src)

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  buildFuncSignature(name, params, ""),
		Properties: plugin.NewProps().SetBool("exported", exported).SetBool("async", isAsync).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// extractClass extracts a class_declaration node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       "class",
		Signature:  name,
		Properties: plugin.NewProps().SetBool("exported", exported).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract heritage edges.
	heritageNames := extractHeritageNames(node, src)
	for _, baseName := range heritageNames {
		c.AddEdge(plugin.Edge{
			From: name,
			To:   baseName,
			Kind: plugin.EdgeInherits,
		})
	}

	// Extract methods from class body.
	if body := childByKind(node, "class_body"); body != nil {
		extractClassMembers(body, src, c, name)
	}
}

// extractClassMembers extracts methods from a class body.
func extractClassMembers(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "method_definition":
			extractJSMethod(child, src, c, className)
		case "field_definition":
			extractJSField(child, src, c, className)
		}
	}
}

// extractJSMethod extracts a method_definition node from a class body.
func extractJSMethod(child *sitter.Node, src []byte, c *plugin.Collector, className string) {
	name := childText(child, "property_identifier", src)
	if name == "" || name == "constructor" {
		return
	}

	scopedName := plugin.MakeScopedName(className, name)
	params := extractParamStrings(child, src)
	isAsync := hasChildKeyword(child, "async", src)
	isStatic := hasChildKeyword(child, "static", src)

	// Detect getter/setter
	kind := "method"
	if hasChildKeyword(child, "get", src) {
		kind = "getter"
	} else if hasChildKeyword(child, "set", src) {
		kind = "setter"
	}

	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  buildFuncSignature(name, params, ""),
		Properties: plugin.NewProps().SetBool("async", isAsync).SetBool("static", isStatic).Map(),
		Span:       nodeSpan(child),
	})

	c.AddEdge(plugin.Edge{From: className, To: scopedName, Kind: plugin.EdgeContains})
}

// extractJSField extracts a field_definition node from a class body.
func extractJSField(child *sitter.Node, src []byte, c *plugin.Collector, className string) {
	name := childText(child, "property_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(className, name)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryValue,
		Kind:       "field",
		Signature:  name,
		Properties: plugin.NewProps().
			SetBool("static", hasChildKeyword(child, "static", src)).
			Map(),
		Span: nodeSpan(child),
	})

	c.AddEdge(plugin.Edge{From: className, To: scopedName, Kind: plugin.EdgeContains})
}

// extractLexicalDeclaration extracts const/let/var declarations.
func extractLexicalDeclaration(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	isConst := false
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.Kind() == "const" {
			isConst = true
			break
		}
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || child.Kind() != "variable_declarator" {
			continue
		}
		extractVariableDeclarator(child, src, c, exported, isConst)
	}
}

// extractVariableDeclaration handles var declarations.
func extractVariableDeclaration(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || child.Kind() != "variable_declarator" {
			continue
		}
		extractVariableDeclarator(child, src, c, exported, false)
	}
}

// extractVariableDeclarator processes a single variable_declarator.
func extractVariableDeclarator(node *sitter.Node, src []byte, c *plugin.Collector, exported, isConst bool) {
	// Check for destructuring patterns first: const { a, b } = obj; or const [x, y] = arr;
	// Must check before identifier because the RHS identifier (e.g., "obj") would match.
	if pat := childByKind(node, "object_pattern"); pat != nil {
		extract.DestructuredNames(pat, src, c, &jsDestructureConfig, exported)
		return
	}
	if pat := childByKind(node, "array_pattern"); pat != nil {
		extract.DestructuredNames(pat, src, c, &jsDestructureConfig, exported)
		return
	}

	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	// Check if the initializer is an arrow function or function expression.
	init := childByKind(node, "arrow_function")
	isArrow := init != nil
	if init == nil {
		init = childByKind(node, "function")
	}

	if init != nil {
		params := extractParamStrings(init, src)
		isAsync := hasChildKeyword(init, "async", src)

		kind := "function"
		if isArrow {
			kind = "arrow_function"
		}

		sym := plugin.Symbol{
			Name:       name,
			Category:   plugin.CategoryCallable,
			Kind:       kind,
			Signature:  buildFuncSignature(name, params, ""),
			Properties: plugin.NewProps().SetBool("exported", exported).SetBool("async", isAsync).Map(),
			Span:       nodeSpan(node),
		}
		c.AddSymbol(&sym)
		return
	}

	// Check if it's an exported const with an object literal (exported constant).
	if exported && isConst {
		if obj := childByKind(node, "object"); obj != nil {
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				Category:   plugin.CategoryValue,
				Kind:       "exported_constant",
				Signature:  name,
				Properties: plugin.NewProps().SetBool("exported", true).Map(),
				Span:       nodeSpan(node),
			})
			extractObjectMethods(obj, src, c, name)
			return
		}
	}

	// Regular variable.
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryValue,
		Kind:       "variable",
		Signature:  name,
		Properties: plugin.NewProps().SetBool("exported", exported).Map(),
		Span:       nodeSpan(node),
	})
}

// extractObjectMethods extracts method shorthand definitions from an object
// literal and emits them as contained callable symbols. This handles patterns
// like: export const pass = { execute(ctx) { ... } };
func extractObjectMethods(obj *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	for i := range obj.ChildCount() {
		child := obj.Child(i)
		if child == nil {
			continue
		}
		// method_definition covers shorthand methods: { execute(ctx) { ... } }
		if child.Kind() == "method_definition" {
			methodName := childText(child, "property_identifier", src)
			if methodName == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentName, methodName)
			params := extractParamStrings(child, src)
			c.AddSymbol(&plugin.Symbol{
				Name:       methodName,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(methodName, params, ""),
				Properties: plugin.NewProps().
					SetBool("async", hasChildKeyword(child, "async", src)).
					Map(),
				Span: nodeSpan(child),
			})
			c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
		}
		// pair with arrow/function value: { handler: (req) => { ... } }
		if child.Kind() == "pair" {
			key := childText(child, "property_identifier", src)
			if key == "" {
				continue
			}
			valFn := childByKind(child, "arrow_function")
			if valFn == nil {
				valFn = childByKind(child, "function")
			}
			if valFn == nil {
				continue
			}
			scopedName := plugin.MakeScopedName(parentName, key)
			params := extractParamStrings(valFn, src)
			c.AddSymbol(&plugin.Symbol{
				Name:       key,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(key, params, ""),
				Properties: plugin.NewProps().
					SetBool("async", hasChildKeyword(valFn, "async", src)).
					Map(),
				Span: nodeSpan(child),
			})
			c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
		}
	}
}

// extractImportStatement extracts an import_statement node and emits EdgeImports edges.
// Handles: import { Foo, Bar } from './module'
//
//	import Foo from './module'
//	import * as Foo from './module'
func extractImportStatement(node *sitter.Node, src []byte, c *plugin.Collector) {
	jsshared.ExtractJSImportStatement(node, src, c)
}
