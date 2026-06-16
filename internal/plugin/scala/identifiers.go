// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scala

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

// scalaParamConfig describes how Scala parameter nodes are structured.
var scalaParamConfig = plugin.ParamConfig{
	ParamListKind: "parameters",
	ParamKinds:    []string{"parameter"},
	NameKind:      "identifier",
	TypeExtractor: func(node *sitter.Node, src []byte) string {
		return plugin.ReturnTypeAfterToken(node, src, ":", nil)
	},
}

// scalaReturnTypeKinds are the valid type node kinds after ":" in a Scala function def.
var scalaReturnTypeKinds = []string{
	"type_identifier", "generic_type", "compound_type",
	"infix_type", "tuple_type", "function_type",
	"parameterized_type",
}

// scalaTypeParamConfig describes how Scala type parameter nodes are structured.
var scalaTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameters",
	TypeParamKind:  "identifier",
	NameKind:       "identifier",
	ConstraintKind: "",
}

// extractScalaModifiers collects modifiers from a Scala definition node.
// Scala modifiers include: abstract, sealed, final, case, override, private, protected, implicit, lazy.
// In the tree-sitter AST, keyword modifiers are anonymous children of a "modifiers" node,
// while "case" appears as a direct anonymous child of the definition node itself.
func extractScalaModifiers(node *sitter.Node, src []byte) map[string]string {
	p := plugin.NewProps()

	// Check for "case" as a direct anonymous child of the definition node.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() && child.Utf8Text(src) == "case" {
			p.SetBool("case", true)
		}
	}

	// Walk the "modifiers" named child for keyword modifiers.
	modsNode := childByKind(node, "modifiers")
	if modsNode == nil {
		return p.Map()
	}
	for i := range modsNode.ChildCount() {
		child := modsNode.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "access_modifier" {
			// access_modifier wraps private/protected as anonymous children.
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				text := gc.Utf8Text(src)
				if text == "private" || text == "protected" {
					p.Set("visibility", text)
				}
			}
			continue
		}
		// Anonymous keyword children: abstract, sealed, final, override, implicit, lazy.
		if !child.IsNamed() {
			text := child.Utf8Text(src)
			switch text {
			case "abstract", "sealed", "final", "override", "implicit", "lazy":
				p.SetBool(text, true)
			}
		}
	}
	return p.Map()
}

// extractParamStrings extracts parameter "name: type" pairs from a parameters node.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	return plugin.ExtractTypedParams(node, src, scalaParamConfig)
}

// extractReturnType extracts the return type from a Scala function definition.
func extractReturnType(node *sitter.Node, src []byte) string {
	return plugin.ReturnTypeAfterToken(node, src, ":", scalaReturnTypeKinds)
}

// callTarget is the CallTargetFunc for this language.
// It returns the call target name if the node is a call_expression, or "" otherwise.
func callTarget(node *sitter.Node, src []byte) string {
	if node.Kind() != "call_expression" {
		return ""
	}
	return callTargetName(node, src)
}

// callTargetName extracts the function/method name from a call_expression node.
// For qualified calls (field_expression), extracts only the method name.
func callTargetName(node *sitter.Node, src []byte) string {
	if node.ChildCount() == 0 {
		return ""
	}
	first := node.Child(0)
	if first == nil {
		return ""
	}
	switch first.Kind() {
	case "identifier":
		return first.Utf8Text(src)
	case "field_expression":
		return plugin.LastNamedLeaf(first, src)
	case "call_expression":
		// Chained call — use the inner target.
		return callTargetName(first, src)
	default:
		return ""
	}
}

// extendsTypes extracts the parent type names from an extends_clause node.
// The first type after "extends" is the superclass; types after "with" keywords
// are traits (implements). This uses the anonymous "with" keyword nodes in the
// AST to structurally distinguish inheritance from trait mixing.
func extendsTypes(node *sitter.Node, src []byte) (superclass string, traits []string) {
	ext := childByKind(node, "extends_clause")
	if ext == nil {
		return "", nil
	}
	afterWith := false
	for i := range ext.ChildCount() {
		child := ext.Child(i)
		if child == nil {
			continue
		}
		// The "with" keyword is an anonymous node that separates the
		// superclass from trait mixins.
		if !child.IsNamed() && child.Utf8Text(src) == "with" {
			afterWith = true
			continue
		}
		var name string
		switch child.Kind() {
		case "type_identifier":
			name = child.Utf8Text(src)
		case "generic_type":
			name = childText(child, "type_identifier", src)
		case "applied_constructor_type":
			name = childText(child, "type_identifier", src)
		}
		if name == "" {
			continue
		}
		if afterWith {
			traits = append(traits, name)
		} else {
			superclass = name
		}
	}
	return superclass, traits
}

// importPath extracts the import path from an import_declaration node.
func importPath(node *sitter.Node, src []byte) string {
	var parts []string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "identifier" {
			parts = append(parts, child.Utf8Text(src))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += "." + p
	}
	return result
}

// extractPackage extracts a package_clause node.
func extractPackage(node *sitter.Node, src []byte, c *plugin.Collector) {
	pi := childByKind(node, "package_identifier")
	if pi == nil {
		return
	}
	name := pi.Utf8Text(src)
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

// extractScalaTypeDecl is the shared implementation for class, trait, and object
// definitions. All three have identical structure: name, modifiers, optional
// extends/with edges, and a template_body to recurse into.
func extractScalaTypeDecl(node *sitter.Node, src []byte, c *plugin.Collector, parentName, kind string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)
	mods := extractScalaModifiers(node, src)
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
	extract.TypeParams(node, src, c, scopedName, scalaTypeParamConfig)

	if parentName != "" {
		c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
	}

	superclass, traits := extendsTypes(node, src)
	if superclass != "" {
		c.AddEdge(plugin.Edge{From: scopedName, To: superclass, Kind: plugin.EdgeInherits})
	}
	for _, t := range traits {
		c.AddEdge(plugin.Edge{From: scopedName, To: t, Kind: plugin.EdgeImplements})
	}

	if body := childByKind(node, "template_body"); body != nil {
		extractBody(body, src, c, scopedName)
	}
}

// extractClass extracts a class_definition node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractScalaTypeDecl(node, src, c, parentName, "class")
}

// extractTrait extracts a trait_definition node.
func extractTrait(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractScalaTypeDecl(node, src, c, parentName, "trait")
}

// extractObject extracts an object_definition node.
func extractObject(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractScalaTypeDecl(node, src, c, parentName, "object")
}

// extractFunction extracts a function_definition or function_declaration node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)
	mods := extractScalaModifiers(node, src)
	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)
	kind := "function"
	if parentName != "" {
		kind = "method"
	}
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: mods,
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, scalaTypeParamConfig)

	if parentName != "" {
		c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
		// Check for override modifier.
		if mods["override"] == "true" {
			c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
		}
	}
}

// extractSimpleScalaSymbol is the shared implementation for simple Scala symbols
// that have a name, modifiers, a category/kind, and an optional contains edge.
// Used by extractVal, extractVar, extractTypeDef, and extractEnum.
func extractSimpleScalaSymbol(node *sitter.Node, src []byte, c *plugin.Collector, parentName, nameKind string, cat plugin.SymbolCategory, kind string) {
	name := childText(node, nameKind, src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)
	mods := extractScalaModifiers(node, src)
	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   cat,
		Kind:       kind,
		Signature:  name,
		Properties: mods,
		Span:       nodeSpan(node),
	})

	if parentName != "" {
		c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractVal extracts a val_definition or val_declaration node.
func extractVal(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractSimpleScalaSymbol(node, src, c, parentName, "identifier", plugin.CategoryValue, "val")
}

// extractVar extracts a var_definition or var_declaration node.
func extractVar(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractSimpleScalaSymbol(node, src, c, parentName, "identifier", plugin.CategoryValue, "variable")
}

// extractTypeDef extracts a type_definition node.
func extractTypeDef(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractSimpleScalaSymbol(node, src, c, parentName, "type_identifier", plugin.CategoryType, "type_alias")
}

// extractEnum extracts an enum_definition node (Scala 3).
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	extractSimpleScalaSymbol(node, src, c, parentName, "identifier", plugin.CategoryType, "enum")
}

// extractImport extracts an import_declaration node and emits EdgeImports edges.
// For "import scala.collection.mutable.HashMap", emits From="HashMap", To="scala.collection.mutable".
func extractImport(node *sitter.Node, src []byte, c *plugin.Collector) {
	path := importPath(node, src)
	if path == "" {
		return
	}
	// Split into module path and local name at the last dot.
	if dot := plugin.LastSepIndex(path, "."); dot >= 0 {
		localName := path[dot+1:]
		modulePath := path[:dot]
		if localName == "_" {
			// Wildcard import: import scala.collection.mutable._
			c.AddEdge(plugin.Edge{From: "*", To: modulePath, Kind: plugin.EdgeImports})
		} else {
			c.AddEdge(plugin.Edge{From: localName, To: modulePath, Kind: plugin.EdgeImports})
		}
	} else {
		// Simple import with no dots.
		c.AddEdge(plugin.Edge{From: path, To: path, Kind: plugin.EdgeImports})
	}
}

// extractBody extracts members from a template_body node.
func extractBody(body *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{ParentName: parentName})
}
