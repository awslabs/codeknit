// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package java

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

// javaParamTypeKinds are the node kinds that represent types in Java parameter nodes.
var javaParamTypeKinds = []string{
	"type_identifier", "generic_type", "array_type",
	"integral_type", "floating_point_type", "boolean_type", "void_type",
}

// javaParamConfig describes how Java parameter nodes are structured.
var javaParamConfig = plugin.ParamConfig{
	ParamListKind: "formal_parameters",
	ParamKinds:    []string{"formal_parameter", "spread_parameter"},
	NameKind:      "identifier",
	TypeExtractor: func(node *sitter.Node, src []byte) string {
		return plugin.FirstChildTextByKinds(node, src, javaParamTypeKinds)
	},
}

// javaReturnTypeKinds are the node kinds that represent return types in Java.
var javaReturnTypeKinds = []string{
	"type_identifier", "generic_type", "array_type",
	"integral_type", "floating_point_type", "boolean_type", "void_type",
}

// javaTypeParamConfig describes how Java type parameter nodes are structured.
var javaTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameters",
	TypeParamKind:  "type_parameter",
	NameKind:       "identifier",
	ConstraintKind: "type_bound",
}

// callTarget is the CallTargetFunc for this language.
// It returns the call target name if the node is a method_invocation, or "" otherwise.
func callTarget(node *sitter.Node, src []byte) string {
	if node.Kind() != "method_invocation" {
		return ""
	}
	return callTargetName(node, src)
}

// callTargetName extracts the method name from a method_invocation node.
// Returns only the unqualified method name (e.g. "println" not "System.out.println").
//
// Java method_invocation AST structure:
//
//	Qualified: field_access, ".", identifier, argument_list
//	Simple:    identifier, argument_list
func callTargetName(node *sitter.Node, src []byte) string {
	// Check for qualified call: field_access . identifier(args)
	// The identifier after the field_access is the method name.
	if childByKind(node, "field_access") != nil {
		foundFA := false
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if child.Kind() == "field_access" {
				foundFA = true
				continue
			}
			if foundFA && child.Kind() == "identifier" {
				return child.Utf8Text(src)
			}
		}
	}

	// Simple call: identifier(args)
	return childText(node, "identifier", src)
}

// extractParamStrings extracts parameter "name: type" pairs from a formal_parameters node.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	return plugin.ExtractTypedParams(node, src, javaParamConfig)
}

// extractReturnType extracts the return type from a method_declaration node.
func extractReturnType(node *sitter.Node, src []byte) string {
	return plugin.ReturnTypeByKinds(node, src, javaReturnTypeKinds, nil)
}

// scopedName extracts the package name from a package_declaration.
// Handles both simple identifiers and scoped_identifier (e.g., com.example).
func scopedName(node *sitter.Node, src []byte) string {
	if si := childByKind(node, "scoped_identifier"); si != nil {
		return si.Utf8Text(src)
	}
	return childText(node, "identifier", src)
}

// typeIdentText extracts a type name from a node that may contain
// type_identifier or generic_type children.
func typeIdentText(node *sitter.Node, src []byte) string {
	if ti := childByKind(node, "type_identifier"); ti != nil {
		return ti.Utf8Text(src)
	}
	if gt := childByKind(node, "generic_type"); gt != nil {
		return childText(gt, "type_identifier", src)
	}
	return ""
}

// extractModifiers collects modifier keywords from a declaration node.
// Java modifiers include: public, private, protected, static, abstract, final, synchronized.
func extractModifiers(node *sitter.Node, src []byte) map[string]string {
	p := plugin.NewProps()
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "modifiers" {
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				text := gc.Utf8Text(src)
				switch text {
				case "public", "private", "protected":
					p.Set("visibility", text)
				case "static", "abstract", "final", "synchronized":
					p.SetBool(text, true)
				}
			}
		}
	}
	return p.Map()
}

// hasOverrideAnnotation checks if a declaration has an @Override annotation.
func hasOverrideAnnotation(node *sitter.Node, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "modifiers" {
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				if gc.Kind() == "marker_annotation" {
					name := childText(gc, "identifier", src)
					if name == "Override" {
						return true
					}
				}
			}
		}
	}
	return false
}

// extractPackage extracts a package_declaration node.
func extractPackage(node *sitter.Node, src []byte, c *plugin.Collector) {
	// package_declaration has a scoped_identifier or identifier child.
	name := scopedName(node, src)
	if name == "" {
		return
	}
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryModule,
		Kind:       "package",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})
}

// extractClass extracts a class_declaration node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "class",
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, javaTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Extract superclass (extends).
	if sc := childByKind(node, "superclass"); sc != nil {
		superName := typeIdentText(sc, src)
		if superName != "" {
			c.AddEdge(plugin.Edge{From: scopedName, To: superName, Kind: plugin.EdgeInherits})
		}
	}

	// Extract interfaces (implements).
	if si := childByKind(node, "super_interfaces"); si != nil {
		extractInterfaceList(si, src, c, scopedName, plugin.EdgeImplements)
	}

	// Extract class body members.
	if body := childByKind(node, "class_body"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractInterface extracts an interface_declaration node.
func extractInterface(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	mods := extractModifiers(node, src)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "interface",
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, javaTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Extract extends_interfaces.
	if ei := childByKind(node, "extends_interfaces"); ei != nil {
		extractInterfaceList(ei, src, c, scopedName, plugin.EdgeInherits)
	}

	// Extract interface body members.
	if body := childByKind(node, "interface_body"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractSimpleType is the shared implementation for type declarations that
// have no body or base types — just a name, kind, modifiers, and optional
// contains edge. Used by extractEnum and extractAnnotationType.
func extractSimpleType(node *sitter.Node, src []byte, c *plugin.Collector, parentClass, kind string) {
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

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractEnum extracts an enum_declaration node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractSimpleType(node, src, c, parentClass, "enum")
}

// extractAnnotationType extracts an annotation_type_declaration node.
func extractAnnotationType(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
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
		Kind:       "annotation",
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	})

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Extract annotation body members (annotation_type_element_declaration).
	if body := childByKind(node, "annotation_type_body"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractRecord extracts a record_declaration node (Java 16+).
// Records are like classes but with auto-generated constructors, accessors, etc.
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
	extract.TypeParams(node, src, c, scopedName, javaTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Extract interfaces (implements).
	if si := childByKind(node, "super_interfaces"); si != nil {
		extractInterfaceList(si, src, c, scopedName, plugin.EdgeImplements)
	}

	// Extract record body members.
	if body := childByKind(node, "class_body"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
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
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "method",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: mods,
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, javaTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Check for @Override annotation.
	if hasOverrideAnnotation(node, src) && parentClass != "" {
		c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
	}
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
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "constructor",
		Signature:  buildFuncSignature(name, params, ""),
		Properties: mods,
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractField extracts a field_declaration node.
func extractField(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	// field_declaration contains a variable_declarator with the field name.
	for i := range node.ChildCount() {
		child := node.Child(i)
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
			sym := plugin.Symbol{
				Name:       name,
				ScopedName: scopedName,
				Category:   plugin.CategoryValue,
				Kind:       "field",
				Signature:  name,
				Properties: mods,
				Span:       nodeSpan(node),
			}
			c.AddSymbol(&sym)

			if parentClass != "" {
				c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
			}
		}
	}
}

// extractClassBody extracts members from a class_body or interface_body node.
func extractClassBody(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{ParentName: className})
}

// extractInterfaceList extracts type names from a type_list node and creates edges.
func extractInterfaceList(node *sitter.Node, src []byte, c *plugin.Collector, className string, kind plugin.EdgeKind) {
	if tl := childByKind(node, "type_list"); tl != nil {
		node = tl
	}
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		name := ""
		switch child.Kind() {
		case "type_identifier":
			name = child.Utf8Text(src)
		case "generic_type":
			name = childText(child, "type_identifier", src)
		}
		if name != "" {
			c.AddEdge(plugin.Edge{From: className, To: name, Kind: kind})
		}
	}
}

// extractImport extracts an import_declaration node and emits an EdgeImports edge.
// For "import java.util.ArrayList", emits From="ArrayList", To="java.util".
// For "import static java.lang.Math.PI", emits From="PI", To="java.lang.Math".
// Wildcard imports (import java.util.*) emit From="*", To="java.util".
func extractImport(node *sitter.Node, src []byte, c *plugin.Collector) {
	// import_declaration contains a scoped_identifier (or identifier).
	fullPath := scopedName(node, src)
	if fullPath == "" {
		return
	}

	// Check for wildcard import: ends with ".*" or has asterisk child.
	if len(fullPath) > 2 && fullPath[len(fullPath)-2:] == ".*" {
		modulePath := fullPath[:len(fullPath)-2]
		c.AddEdge(plugin.Edge{
			From: "*",
			To:   modulePath,
			Kind: plugin.EdgeImports,
		})
		return
	}

	// Split into module path and local name at the last dot.
	lastDot := -1
	for i := len(fullPath) - 1; i >= 0; i-- {
		if fullPath[i] == '.' {
			lastDot = i
			break
		}
	}
	if lastDot < 0 {
		// Simple import with no dots — unlikely in Java but handle gracefully.
		c.AddEdge(plugin.Edge{
			From: fullPath,
			To:   fullPath,
			Kind: plugin.EdgeImports,
		})
		return
	}

	localName := fullPath[lastDot+1:]
	modulePath := fullPath[:lastDot]
	c.AddEdge(plugin.Edge{
		From: localName,
		To:   modulePath,
		Kind: plugin.EdgeImports,
	})
}
