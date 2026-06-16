// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clang

import (
	"codeknit/internal/plugin"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Aliases for shared helpers — keeps call sites unchanged.
var (
	childByKind        = plugin.ChildByKind
	childText          = plugin.ChildText
	nodeSpan           = plugin.NodeSpan
	buildFuncSignature = plugin.BuildFuncSignature
)

// cTypeKinds are the node kinds that represent types in C.
var cTypeKinds = []string{
	"primitive_type", "type_identifier", "sized_type_specifier",
	"struct_specifier", "union_specifier", "enum_specifier",
}

// cReturnStopKinds are node kinds that signal we've passed the return type in C.
var cReturnStopKinds = []string{"function_declarator", "pointer_declarator"}

// storageClassProps extracts static and extern storage class specifiers from a declaration node.
func storageClassProps(node *sitter.Node, src []byte) (isStatic, isExtern bool) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "storage_class_specifier" {
			text := child.Utf8Text(src)
			switch text {
			case "static":
				isStatic = true
			case "extern":
				isExtern = true
			}
		}
	}
	return
}

// callTarget is the CallTargetFunc for C.
// Matches call_expression nodes whose first child is an identifier or field_expression.
// For field_expression (e.g. ptr->func), extracts only the function name.
var callTarget = plugin.UnqualifiedCallTarget("call_expression",
	[]string{"identifier"},
	[]string{"field_expression"},
)

// richCallTarget is the RichCallTargetFunc for C.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call_expression",
	[]string{"identifier"},
	[]string{"field_expression"},
)

// extractParamStrings extracts function parameter "name: type" pairs from a function_definition.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	decl := childByKind(node, "function_declarator")
	if decl == nil {
		pd := childByKind(node, "pointer_declarator")
		if pd != nil {
			decl = childByKind(pd, "function_declarator")
		}
	}
	if decl == nil {
		return nil
	}
	paramList := childByKind(decl, "parameter_list")
	if paramList == nil {
		return nil
	}
	var result []string
	for i := range paramList.ChildCount() {
		child := paramList.Child(i)
		if child == nil || child.Kind() != "parameter_declaration" {
			continue
		}
		name := childText(child, "identifier", src)
		if name == "" {
			continue
		}
		typeName := cParamType(child, src)
		if typeName != "" {
			result = append(result, name+": "+typeName)
		} else {
			result = append(result, name)
		}
	}
	return result
}

// cParamType extracts the type from a C parameter_declaration node.
func cParamType(node *sitter.Node, src []byte) string {
	return plugin.FirstChildTextByKinds(node, src, cTypeKinds)
}

// extractReturnType extracts the return type from a C function_definition node.
func extractReturnType(node *sitter.Node, src []byte) string {
	return plugin.ReturnTypeByKinds(node, src, cTypeKinds, cReturnStopKinds)
}

// extractFunction extracts a function_definition node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := funcDeclName(node, src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)
	isStatic, isExtern := storageClassProps(node, src)

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().SetBool("static", isStatic).SetBool("extern", isExtern).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// funcDeclName extracts the function name from a function_definition.
// The name lives inside a function_declarator child.
func funcDeclName(node *sitter.Node, src []byte) string {
	decl := childByKind(node, "function_declarator")
	if decl == nil {
		// Try pointer_declarator wrapping a function_declarator.
		pd := childByKind(node, "pointer_declarator")
		if pd != nil {
			decl = childByKind(pd, "function_declarator")
		}
	}
	if decl == nil {
		return ""
	}
	return childText(decl, "identifier", src)
}

// extractDeclaration handles a top-level "declaration" node.
// This can be a function prototype or a global variable declaration.
func extractDeclaration(node *sitter.Node, src []byte, c *plugin.Collector) {
	// Check if this declaration contains a function_declarator (prototype).
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "function_declarator":
			extractPrototype(node, child, src, c)
			return
		case "init_declarator":
			// Check if init_declarator wraps a function_declarator.
			if fd := childByKind(child, "function_declarator"); fd != nil {
				extractPrototype(node, fd, src, c)
				return
			}
			// Otherwise it's a variable with initializer.
			extractGlobalVar(node, child, src, c)
			return
		}
	}

	// Plain variable declaration without init_declarator.
	name := childText(node, "identifier", src)
	if name != "" {
		isStatic, isExtern := storageClassProps(node, src)
		sym := plugin.Symbol{
			Name:       name,
			Category:   plugin.CategoryValue,
			Kind:       "variable",
			Signature:  name,
			Properties: plugin.NewProps().SetBool("static", isStatic).SetBool("extern", isExtern).Map(),
			Span:       nodeSpan(node),
		}
		c.AddSymbol(&sym)
	}
}

// extractDeclSymbol is the shared implementation for extractPrototype and
// extractGlobalVar. Both read storage class props from declNode, extract a
// name from innerDecl, and emit a symbol differing only in category and kind.
func extractDeclSymbol(declNode, innerDecl *sitter.Node, src []byte, c *plugin.Collector, cat plugin.SymbolCategory, kind string) {
	name := childText(innerDecl, "identifier", src)
	if name == "" {
		return
	}

	isStatic, isExtern := storageClassProps(declNode, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   cat,
		Kind:       kind,
		Signature:  name,
		Properties: plugin.NewProps().SetBool("static", isStatic).SetBool("extern", isExtern).Map(),
		Span:       nodeSpan(declNode),
	})
}

// extractPrototype extracts a function prototype from a declaration.
func extractPrototype(declNode, funcDecl *sitter.Node, src []byte, c *plugin.Collector) {
	extractDeclSymbol(declNode, funcDecl, src, c, plugin.CategoryCallable, "function_prototype")
}

// extractGlobalVar extracts a global variable from an init_declarator.
func extractGlobalVar(declNode, initDecl *sitter.Node, src []byte, c *plugin.Collector) {
	extractDeclSymbol(declNode, initDecl, src, c, plugin.CategoryValue, "variable")
}

// extractAggregateType is the shared implementation for extractStruct and
// extractUnion. Both extract a type_identifier, emit a CategoryType symbol,
// and recurse into a field_declaration_list.
func extractAggregateType(node *sitter.Node, src []byte, c *plugin.Collector, kind string) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	addSimpleCSymbol(node, src, c, "type_identifier", plugin.CategoryType, kind)

	if body := childByKind(node, "field_declaration_list"); body != nil {
		extractStructFields(body, src, c, name)
	}
}

// extractStruct extracts a struct_specifier node.
func extractStruct(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractAggregateType(node, src, c, "struct")
}

// extractUnion extracts a union_specifier node.
func extractUnion(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractAggregateType(node, src, c, "union")
}

// extractStructFields extracts field_declaration children from a field_declaration_list.
func extractStructFields(body *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil || child.Kind() != "field_declaration" {
			continue
		}
		fieldName := childText(child, "field_identifier", src)
		if fieldName != "" {
			scopedName := plugin.MakeScopedName(parentName, fieldName)
			c.AddEdge(plugin.Edge{
				From: parentName,
				To:   scopedName,
				Kind: plugin.EdgeContains,
			})
		}
	}
}

// addSimpleCSymbol emits a symbol with a name extracted from a child of the
// given kind. Used by extractEnum, extractTypedef, and extractMacroSymbol to
// avoid repeating the same name→Symbol→AddSymbol scaffold.
func addSimpleCSymbol(node *sitter.Node, src []byte, c *plugin.Collector, nameKind string, cat plugin.SymbolCategory, kind string) {
	name := childText(node, nameKind, src)
	if name == "" {
		return
	}
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   cat,
		Kind:       kind,
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})
}

// extractEnum extracts an enum_specifier node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector) {
	addSimpleCSymbol(node, src, c, "type_identifier", plugin.CategoryType, "enum")
}

// extractTypedef extracts a type_definition node.
func extractTypedef(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := typedefName(node, src)
	if name == "" {
		return
	}
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       "typedef",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})
}

// typedefName extracts the typedef alias name from a type_definition node.
// The name is typically a type_identifier, but tree-sitter-c may also represent
// well-known names (e.g. size_t) as primitive_type. We look for the last
// type_identifier or primitive_type that isn't the source type.
func typedefName(node *sitter.Node, src []byte) string {
	// Prefer type_identifier (most common case).
	if name := childText(node, "type_identifier", src); name != "" {
		return name
	}
	// Fallback: find the last primitive_type child that serves as the alias name.
	// In "typedef unsigned long size_t;", there are multiple primitive_type nodes;
	// the last one before ";" is the alias.
	var lastName string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "primitive_type" {
			lastName = child.Utf8Text(src)
		}
	}
	return lastName
}

// extractMacroSymbol extracts a preprocessor macro symbol. Delegates to
// addSimpleCSymbol since macros follow the same name→Symbol pattern.
func extractMacroSymbol(node *sitter.Node, src []byte, c *plugin.Collector, cat plugin.SymbolCategory, kind string) {
	addSimpleCSymbol(node, src, c, "identifier", cat, kind)
}

// extractPreprocDef extracts a preproc_def (#define NAME value) node.
func extractPreprocDef(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractMacroSymbol(node, src, c, plugin.CategoryValue, "macro")
}

// extractPreprocFunctionDef extracts a preproc_function_def (#define NAME(...) ...) node.
func extractPreprocFunctionDef(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractMacroSymbol(node, src, c, plugin.CategoryCallable, "macro_function")
}

// extractInclude extracts a preproc_include (#include) node as a references edge.
func extractInclude(node *sitter.Node, src []byte, c *plugin.Collector) {
	// The include path is in a system_lib_string or string_literal child.
	var target string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "system_lib_string", "string_literal":
			target = child.Utf8Text(src)
		}
	}
	if target == "" {
		return
	}

	c.AddEdge(plugin.Edge{
		From: c.FilePath,
		To:   target,
		Kind: plugin.EdgeReferences,
	})
}
