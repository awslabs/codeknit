// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csharp

import (
	"codeknit/internal/common/extract"
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

// csParamConfig describes how C# parameter nodes are structured.
var csParamConfig = plugin.ParamConfig{
	ParamListKind: "parameter_list",
	ParamKinds:    []string{"parameter"},
	NameKind:      "identifier",
	TypeExtractor: csParamTypeName,
}

// csTypeParamConfig describes how C# type parameter nodes are structured.
var csTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameter_list",
	TypeParamKind:  "type_parameter",
	NameKind:       "identifier",
	ConstraintKind: "where_clause",
}

// callTarget is the CallTargetFunc for C#.
// Matches invocation_expression nodes, extracting the unqualified method name.
// For member_access_expression (e.g. Console.WriteLine), extracts only "WriteLine".
var callTarget plugin.CallTargetFunc = func(node *sitter.Node, src []byte) string {
	if node.Kind() != "invocation_expression" {
		return ""
	}
	// Try member_access_expression first — extract just the method name.
	if mae := childByKind(node, "member_access_expression"); mae != nil {
		return plugin.LastNamedLeaf(mae, src)
	}
	// Simple call: identifier_name or identifier.
	if name := childText(node, "identifier_name", src); name != "" {
		return name
	}
	return childText(node, "identifier", src)
}

// extractParamStrings extracts parameter "name: type" pairs from a parameter_list node.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	return plugin.ExtractTypedParams(node, src, csParamConfig)
}

// csParamTypeName extracts the type name from a C# parameter node.
func csParamTypeName(node *sitter.Node, src []byte) string {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier_name", "identifier":
			// Skip — this is the parameter name, not the type.
			continue
		case "predefined_type", "generic_name", "qualified_name",
			"nullable_type", "array_type":
			return child.Utf8Text(src)
		}
	}
	return ""
}

// extractReturnType extracts the return type from a C# method_declaration node.
// C# has a special case: an "identifier" before parameter_list is the method name,
// not the return type. This requires custom logic beyond ReturnTypeByKinds.
func extractReturnType(node *sitter.Node, src []byte) string {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "predefined_type", "generic_name", "qualified_name",
			"nullable_type", "array_type", "identifier_name":
			return child.Utf8Text(src)
		case "identifier":
			if i+1 < node.ChildCount() {
				next := node.Child(i + 1)
				if next != nil && next.Kind() == "parameter_list" {
					return ""
				}
			}
			return child.Utf8Text(src)
		}
	}
	return ""
}

// extractBaseListNames extracts type names from a base_list node.
// C# uses base_list for both inheritance and interface implementation.
func extractBaseListNames(node *sitter.Node, src []byte) []string {
	bl := childByKind(node, "base_list")
	if bl == nil {
		return nil
	}
	var names []string
	for i := range bl.ChildCount() {
		child := bl.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier_name", "identifier":
			names = append(names, child.Utf8Text(src))
		case "generic_name":
			name := childText(child, "identifier_name", src)
			if name == "" {
				name = childText(child, "identifier", src)
			}
			if name != "" {
				names = append(names, name)
			}
		case "qualified_name":
			names = append(names, child.Utf8Text(src))
		case "predefined_type":
			names = append(names, child.Utf8Text(src))
		case "simple_base_type":
			// simple_base_type wraps the actual type name
			name := extractTypeName(child, src)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// extractTypeName extracts a type name from a type node.
func extractTypeName(node *sitter.Node, src []byte) string {
	// Try identifier_name first
	if name := childText(node, "identifier_name", src); name != "" {
		return name
	}
	// Try identifier
	if name := childText(node, "identifier", src); name != "" {
		return name
	}
	// Try generic_name
	if gn := childByKind(node, "generic_name"); gn != nil {
		name := childText(gn, "identifier_name", src)
		if name == "" {
			name = childText(gn, "identifier", src)
		}
		return name
	}
	// Try qualified_name
	if qn := childByKind(node, "qualified_name"); qn != nil {
		return qn.Utf8Text(src)
	}
	return ""
}

// extractModifiers collects modifier keywords from a declaration node.
// C# modifiers include: public, private, protected, internal, static, abstract,
// sealed, virtual, override, async, readonly, new, partial, extern, unsafe, volatile.
func extractModifiers(node *sitter.Node, src []byte) map[string]string {
	p := plugin.NewProps()
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		kind := child.Kind()
		// Handle both "modifiers" wrapper and individual "modifier" nodes.
		if kind == "modifiers" || kind == "modifier" {
			// If it has children, iterate them.
			if child.ChildCount() > 0 {
				for j := range child.ChildCount() {
					gc := child.Child(j)
					if gc != nil {
						applyModifier(p, gc.Utf8Text(src))
					}
				}
			} else {
				// Individual modifier node with no children — use its own text.
				applyModifier(p, child.Utf8Text(src))
			}
		}
	}
	return p.Map()
}

// applyModifier maps a modifier keyword to the appropriate property.
func applyModifier(p plugin.Props, text string) {
	switch text {
	case "public", "private", "protected", "internal":
		p.Set("visibility", text)
	case "static", "abstract", "sealed", "virtual", "override", "async",
		"readonly", "new", "partial", "extern", "unsafe", "volatile":
		p.SetBool(text, true)
	}
}

// extractNamespace extracts a namespace_declaration node.
func extractNamespace(node *sitter.Node, src []byte, c *plugin.Collector) {
	// namespace name can be identifier or qualified_name.
	name := childText(node, "qualified_name", src)
	if name == "" {
		name = childText(node, "identifier", src)
	}
	if name == "" {
		name = childText(node, "identifier_name", src)
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
	if body := childByKind(node, "declaration_list"); body != nil {
		for i := range body.ChildCount() {
			child := body.Child(i)
			if child != nil {
				handlers.Dispatch(child, src, c, plugin.HandlerContext{})
			}
		}
	}
}

// extractTypeDecl is the shared implementation for class, interface, and struct
// declarations. It creates a symbol with the given kind, emits a contains edge
// if nested, emits base-type edges using edgeKind, and recurses into the body.
func extractTypeDecl(node *sitter.Node, src []byte, c *plugin.Collector, parentClass, kind string, edgeKind plugin.EdgeKind) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       kind,
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, csTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	baseNames := extractBaseListNames(node, src)
	for _, base := range baseNames {
		c.AddEdge(plugin.Edge{From: scopedName, To: base, Kind: edgeKind})
	}

	if body := childByKind(node, "declaration_list"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractClass extracts a class_declaration node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractTypeDecl(node, src, c, parentClass, "class", plugin.EdgeInherits)
}

// extractInterface extracts an interface_declaration node.
func extractInterface(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractTypeDecl(node, src, c, parentClass, "interface", plugin.EdgeInherits)
}

// extractStruct extracts a struct_declaration node.
func extractStruct(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractTypeDecl(node, src, c, parentClass, "struct", plugin.EdgeImplements)
}

// extractSimpleCSharpDecl is the shared implementation for C# declarations that
// have a name, modifiers, a category/kind, and an optional contains edge.
// Used by extractEnum, extractProperty, and extractDelegate.
func extractSimpleCSharpDecl(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string, cat plugin.SymbolCategory, kind string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   cat,
		Kind:       kind,
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	})

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractEnum extracts an enum_declaration node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractSimpleCSharpDecl(node, src, c, parentClass, plugin.CategoryType, "enum")
}

// extractProperty extracts a property_declaration node.
func extractProperty(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractSimpleCSharpDecl(node, src, c, parentClass, plugin.CategoryValue, "property")
}

// extractDelegate extracts a delegate_declaration node.
func extractDelegate(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractSimpleCSharpDecl(node, src, c, parentClass, plugin.CategoryType, "delegate")
}

// extractMethod extracts a method_declaration node.
func extractMethod(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "method",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: mods,
		Span:       nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, csTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	if mods["override"] == "true" && parentClass != "" {
		c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
	}
}

// extractField extracts a field_declaration node.
func extractField(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	vd := childByKind(node, "variable_declaration")
	if vd == nil {
		return
	}
	for i := range vd.ChildCount() {
		child := vd.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "variable_declarator" {
			name := childText(child, "identifier", src)
			if name == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentClass, name)
			mods := extractModifiers(node, src)
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				ScopedName: scopedName,
				Category:   plugin.CategoryValue,
				Kind:       "field",
				Signature:  name,
				Properties: mods,
				Span:       nodeSpan(node),
			})

			if parentClass != "" {
				c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
			}
		}
	}
}

// extractClassBody extracts members from a declaration_list node.
func extractClassBody(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{ParentName: className})
}

// extractConstructor extracts a constructor_declaration node.
func extractConstructor(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	params := extractParamStrings(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "constructor",
		Signature:  buildFuncSignature(name, params, ""),
		Properties: mods,
		Span:       nodeSpan(node),
	})

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractRecord extracts a record_declaration or record_struct_declaration node (C# 9+).
func extractRecord(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "record",
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, csTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	baseNames := extractBaseListNames(node, src)
	for _, base := range baseNames {
		c.AddEdge(plugin.Edge{From: scopedName, To: base, Kind: plugin.EdgeInherits})
	}

	if body := childByKind(node, "declaration_list"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractEvent extracts an event_field_declaration node.
func extractEvent(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	// event_field_declaration contains a variable_declaration with the event name.
	vd := childByKind(node, "variable_declaration")
	if vd == nil {
		return
	}
	for i := range vd.ChildCount() {
		child := vd.Child(i)
		if child == nil || child.Kind() != "variable_declarator" {
			continue
		}
		name := childText(child, "identifier", src)
		if name == "" {
			continue
		}
		scopedName := plugin.MakeScopedName(parentClass, name)
		mods := extractModifiers(node, src)
		c.AddSymbol(&plugin.Symbol{
			Name:       name,
			ScopedName: scopedName,
			Category:   plugin.CategoryValue,
			Kind:       "event",
			Signature:  name,
			Properties: mods,
			Span:       nodeSpan(node),
		})
		if parentClass != "" {
			c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
		}
	}
}

// extractOperator extracts an operator_declaration node.
func extractOperator(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	// The operator symbol is typically an anonymous child like "+", "-", "==", etc.
	var opName string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		text := child.Utf8Text(src)
		switch text {
		case "+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=",
			"&", "|", "^", "~", "<<", ">>", "!", "++", "--":
			opName = "operator" + text
		}
	}
	if opName == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, opName)
	mods := extractModifiers(node, src)
	params := extractParamStrings(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       opName,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "operator",
		Signature:  buildFuncSignature(opName, params, ""),
		Properties: mods,
		Span:       nodeSpan(node),
	})

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractUsingDirective extracts a using_directive node and emits EdgeImports edges.
// Handles: using System.Collections.Generic;
//
//	using Alias = System.Collections.Generic.List;
func extractUsingDirective(node *sitter.Node, src []byte, c *plugin.Collector) {
	// Check for alias: using Alias = Namespace.Type;
	if eq := childByKind(node, "name_equals"); eq != nil {
		alias := childText(eq, "identifier", src)
		if alias == "" {
			alias = childText(eq, "identifier_name", src)
		}
		// The qualified name after "=" is the target.
		qn := childText(node, "qualified_name", src)
		if qn == "" {
			qn = childText(node, "identifier_name", src)
		}
		if alias != "" && qn != "" {
			c.AddEdge(plugin.Edge{From: alias, To: qn, Kind: plugin.EdgeImports})
		}
		return
	}

	// Regular using: using System.Collections.Generic;
	// C# using directives import namespaces, not individual names.
	// We emit a wildcard import so the planner can match any name from that namespace.
	ns := childText(node, "qualified_name", src)
	if ns == "" {
		ns = childText(node, "identifier_name", src)
	}
	if ns == "" {
		ns = childText(node, "identifier", src)
	}
	if ns != "" {
		c.AddEdge(plugin.Edge{From: "*", To: ns, Kind: plugin.EdgeImports})
	}
}
