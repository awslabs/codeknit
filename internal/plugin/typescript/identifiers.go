// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package typescript

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

// tsDestructureConfig describes how TypeScript destructuring patterns are structured.
var tsDestructureConfig = extract.DestructureConfig{
	ObjectPatternKind:     "object_pattern",
	ArrayPatternKind:      "array_pattern",
	IdentKinds:            []string{"identifier", "shorthand_property_identifier_pattern"},
	RestKind:              "rest_pattern",
	AssignmentPatternKind: "assignment_pattern",
	PairPatternKind:       "pair_pattern",
}

// tsParamConfig describes how TypeScript parameter nodes are structured.
var tsParamConfig = plugin.ParamConfig{
	ParamListKind: "formal_parameters",
	ParamKinds:    []string{"required_parameter", "optional_parameter", "identifier"},
	NameKind:      "identifier",
	TypeExtractor: typeAnnotationText,
}

// tsTypeParamConfig describes how TypeScript type parameter nodes are structured.
var tsTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameters",
	TypeParamKind:  "type_parameter",
	NameKind:       "type_identifier",
	ConstraintKind: "constraint",
}

// typeAnnotationText extracts the type text from a node containing a type_annotation child.
func typeAnnotationText(node *sitter.Node, src []byte) string {
	ta := childByKind(node, "type_annotation")
	if ta == nil {
		return ""
	}
	return typeNodeText(ta, src)
}

// typeNodeText extracts the type text from a type_annotation node (skips the colon).
func typeNodeText(ta *sitter.Node, src []byte) string {
	for i := range ta.ChildCount() {
		child := ta.Child(i)
		if child != nil && child.Kind() != ":" {
			return child.Utf8Text(src)
		}
	}
	return ""
}

// extractParamStrings extracts function parameters as "name: type" string pairs.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	return plugin.ExtractTypedParams(node, src, tsParamConfig)
}

// extractReturnTypeString extracts the return type annotation as a string.
func extractReturnTypeString(node *sitter.Node, src []byte) string {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "type_annotation" {
			text := typeNodeText(child, src)
			if text != "" {
				return text
			}
		}
	}
	return ""
}

// extractHeritageNames extracts base type names from extends/implements clauses as strings.
func extractHeritageNames(node *sitter.Node, src []byte) []string {
	heritage := childByKind(node, "class_heritage")
	if heritage == nil {
		return extractHeritageNamesFromChildren(node, src)
	}
	return extractHeritageNamesFromChildren(heritage, src)
}

// extractHeritageNamesFromChildren extracts type names from extends/implements clause children.
func extractHeritageNamesFromChildren(node *sitter.Node, src []byte) []string {
	var result []string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		kind := child.Kind()
		switch kind {
		case "extends_clause", "implements_clause", "extends_type_clause":
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				switch gc.Kind() {
				case "identifier", "type_identifier":
					result = append(result, gc.Utf8Text(src))
				case "generic_type":
					name := childText(gc, "type_identifier", src)
					if name == "" {
						name = childText(gc, "identifier", src)
					}
					if name != "" {
						result = append(result, name)
					}
				}
			}
		case "identifier", "type_identifier":
			result = append(result, child.Utf8Text(src))
		case "generic_type":
			name := childText(child, "type_identifier", src)
			if name == "" {
				name = childText(child, "identifier", src)
			}
			if name != "" {
				result = append(result, name)
			}
		}
	}
	return result
}

// callTarget is the CallTargetFunc for TypeScript.
var callTarget = plugin.UnqualifiedCallTarget("call_expression",
	[]string{"identifier"},
	[]string{"member_expression"},
)

// richCallTarget is the RichCallTargetFunc for TypeScript.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call_expression",
	[]string{"identifier"},
	[]string{"member_expression"},
)

// buildVarSignature builds a human-readable variable signature string.
func buildVarSignature(name, typeName string) string {
	if typeName != "" {
		return name + ": " + typeName
	}
	return name
}

// extractFunction extracts a function_declaration node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	returnType := extractReturnTypeString(node, src)

	c.AddSymbol(&plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryCallable,
		Kind:      "function",
		Signature: buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			SetBool("async", hasChildKeyword(node, "async", src)).
			Map(),
		Span: nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, name, tsTypeParamConfig)
}

// extractClass extracts a class_declaration node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		name = childText(node, "identifier", src)
	}
	if name == "" {
		return
	}

	c.AddSymbol(&plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryType,
		Kind:      "class",
		Signature: name,
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			SetBool("abstract", hasChildKeyword(node, "abstract", src)).
			Map(),
		Span: nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, name, tsTypeParamConfig)

	for _, baseName := range extractHeritageNames(node, src) {
		c.AddEdge(plugin.Edge{From: name, To: baseName, Kind: plugin.EdgeInherits})
	}

	if body := childByKind(node, "class_body"); body != nil {
		extractClassMembers(body, src, c, name)
	}

	// Emit decorator edges for the class.
	extractTSDecoratorEdges(node, src, c, name)
}

// extractClassMembers extracts methods and fields from a class body.
func extractClassMembers(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "method_definition":
			extractTSMethod(child, src, c, className)
		case "public_field_definition":
			extractTSField(child, src, c, className)
		}
	}
}

// extractTSMethod extracts a method_definition node from a class body.
func extractTSMethod(child *sitter.Node, src []byte, c *plugin.Collector, className string) {
	name := childText(child, "property_identifier", src)
	if name == "" || name == "constructor" {
		return
	}

	scopedName := plugin.MakeScopedName(className, name)
	params := extractParamStrings(child, src)
	returnType := extractReturnTypeString(child, src)

	isOverride := hasChildKeyword(child, "override", src)

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
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().
			SetBool("async", hasChildKeyword(child, "async", src)).
			Set("visibility", extractVisibility(child, src)).
			SetBool("static", hasChildKeyword(child, "static", src)).
			SetBool("abstract", hasChildKeyword(child, "abstract", src)).
			SetBool("override", isOverride).
			Map(),
		Span: nodeSpan(child),
	})

	c.AddEdge(plugin.Edge{From: className, To: scopedName, Kind: plugin.EdgeContains})

	if isOverride {
		c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
	}

	// Emit decorator edges for the method.
	extractTSDecoratorEdges(child, src, c, scopedName)
}

// extractTSField extracts a public_field_definition node from a class body.
func extractTSField(child *sitter.Node, src []byte, c *plugin.Collector, className string) {
	name := childText(child, "property_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(className, name)
	typeName := typeAnnotationText(child, src)

	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryValue,
		Kind:       "field",
		Signature:  buildVarSignature(name, typeName),
		Properties: plugin.NewProps().
			Set("visibility", extractVisibility(child, src)).
			SetBool("static", hasChildKeyword(child, "static", src)).
			SetBool("readonly", hasChildKeyword(child, "readonly", src)).
			Map(),
		Span: nodeSpan(child),
	})

	c.AddEdge(plugin.Edge{From: className, To: scopedName, Kind: plugin.EdgeContains})

	// Emit decorator edges for the field.
	extractTSDecoratorEdges(child, src, c, scopedName)
}

// extractVisibility returns the explicit visibility modifier, or "" if none.
func extractVisibility(node *sitter.Node, src []byte) string {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.Kind() == "accessibility_modifier" {
			return child.Utf8Text(src)
		}
	}
	return ""
}

// tsDecoratorName is the DecoratorNameFunc for TypeScript.
// TypeScript uses "member_expression" for qualified decorators (@Foo.bar) and
// "call_expression" for decorators with arguments (@Foo()).
var tsDecoratorName = plugin.MakeDecoratorNameFunc("member_expression", "call_expression")

// extractTSDecoratorEdges emits EdgeDecorates edges from each decorator name
// to the decorated symbol. Handles @Foo, @Foo(), @Foo.bar, @Foo.bar().
func extractTSDecoratorEdges(node *sitter.Node, src []byte, c *plugin.Collector, targetName string) {
	plugin.ExtractDecoratorEdges(node, src, c, targetName, "decorator", tsDecoratorName)
}

// extractInterface extracts an interface_declaration node.
func extractInterface(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	c.AddSymbol(&plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryType,
		Kind:      "interface",
		Signature: name,
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			Map(),
		Span: nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, name, tsTypeParamConfig)

	for _, baseName := range extractHeritageNames(node, src) {
		c.AddEdge(plugin.Edge{From: name, To: baseName, Kind: plugin.EdgeImplements})
	}
}

// extractTypeAlias extracts a type_alias_declaration node.
func extractTypeAlias(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	c.AddSymbol(&plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryType,
		Kind:      "type_alias",
		Signature: name,
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			Map(),
		Span: nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, name, tsTypeParamConfig)
}

// extractEnum extracts an enum_declaration node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, exported bool) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	c.AddSymbol(&plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryType,
		Kind:      "enum",
		Signature: name,
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			Map(),
		Span: nodeSpan(node),
	})
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
		extract.DestructuredNames(pat, src, c, &tsDestructureConfig, exported)
		return
	}
	if pat := childByKind(node, "array_pattern"); pat != nil {
		extract.DestructuredNames(pat, src, c, &tsDestructureConfig, exported)
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
		kind := "function"
		if isArrow {
			kind = "arrow_function"
		}

		c.AddSymbol(&plugin.Symbol{
			Name:      name,
			Category:  plugin.CategoryCallable,
			Kind:      kind,
			Signature: buildFuncSignature(name, extractParamStrings(init, src), extractReturnTypeString(init, src)),
			Properties: plugin.NewProps().
				SetBool("exported", exported).
				SetBool("async", hasChildKeyword(init, "async", src)).
				Map(),
			Span: nodeSpan(node),
		})

		return
	}

	// Exported const with object literal.
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
		Name:      name,
		Category:  plugin.CategoryValue,
		Kind:      "variable",
		Signature: buildVarSignature(name, typeAnnotationText(node, src)),
		Properties: plugin.NewProps().
			SetBool("exported", exported).
			Map(),
		Span: nodeSpan(node),
	})
}

// extractObjectMethods extracts method shorthand definitions from an object
// literal and emits them as contained callable symbols. This handles patterns
// like: export const pass: CompilerPass = { execute(ctx) { ... } };
// where "execute" is a method shorthand inside the object.
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
			returnType := extractReturnTypeString(child, src)
			c.AddSymbol(&plugin.Symbol{
				Name:       methodName,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(methodName, params, returnType),
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
			returnType := extractReturnTypeString(valFn, src)
			c.AddSymbol(&plugin.Symbol{
				Name:       key,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(key, params, returnType),
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
func extractImportStatement(node *sitter.Node, src []byte, c *plugin.Collector) {
	jsshared.ExtractJSImportStatement(node, src, c)
}

// extractNamespace extracts a TypeScript namespace/module declaration.
// In tree-sitter-typescript, `namespace Foo { ... }` is parsed as a "module" node.
func extractNamespace(node *sitter.Node, src []byte, c *plugin.Collector) {
	// The name can be an identifier or a string (for ambient module declarations).
	name := childText(node, "identifier", src)
	if name == "" {
		// Ambient module: declare module "express" { ... }
		if str := childByKind(node, "string"); str != nil {
			raw := str.Utf8Text(src)
			if len(raw) >= 2 {
				name = raw[1 : len(raw)-1]
			}
		}
	}
	if name == "" {
		return
	}

	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryModule,
		Kind:       "namespace",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})

	// Extract members inside the namespace body.
	if body := childByKind(node, "statement_block"); body != nil {
		for i := range body.ChildCount() {
			child := body.Child(i)
			if child != nil {
				handlers.Dispatch(child, src, c, plugin.HandlerContext{})
			}
		}
	}
}

// extractAmbientDeclaration extracts a `declare` statement.
// This handles: declare module "express" { ... }, declare namespace Foo { ... }, etc.
func extractAmbientDeclaration(node *sitter.Node, src []byte, c *plugin.Collector) {
	// Ambient declarations wrap inner declarations. Dispatch each child.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "module", "internal_module":
			extractNamespace(child, src, c)
		default:
			handlers.Dispatch(child, src, c, plugin.HandlerContext{})
		}
	}
}
