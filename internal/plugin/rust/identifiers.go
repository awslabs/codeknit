// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rust

import (
	"slices"
	"strings"

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

// rustTypeParamConfig describes how Rust type parameter nodes are structured.
var rustTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameters",
	TypeParamKind:  "type_parameter",
	NameKind:       "identifier",
	ConstraintKind: "trait_bounds",
}

// isPublic checks if a Rust item has a visibility_modifier child (pub, pub(crate), etc.).
func isPublic(node *sitter.Node) bool {
	return childByKind(node, "visibility_modifier") != nil
}

// hasChildKind returns true if the node has a direct child with the given kind.
func hasChildKind(node *sitter.Node, kind string) bool {
	return childByKind(node, kind) != nil
}

// hasModifier checks if a function_item has a specific modifier (async, unsafe, etc.)
// inside its function_modifiers child.
func hasModifier(node *sitter.Node, modifier string) bool {
	mods := childByKind(node, "function_modifiers")
	if mods == nil {
		return false
	}
	return childByKind(mods, modifier) != nil
}

// callTarget is the CallTargetFunc for Rust.
// Matches call_expression nodes whose first child is an identifier, field_expression, or scoped_identifier.
// For qualified nodes (e.g. Vec::new, self.method), extracts only the function name.
var callTarget = plugin.UnqualifiedCallTarget("call_expression",
	[]string{"identifier"},
	[]string{"field_expression", "scoped_identifier"},
)

// richCallTarget is the RichCallTargetFunc for Rust.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call_expression",
	[]string{"identifier"},
	[]string{"field_expression", "scoped_identifier"},
)

// extractParamStrings extracts function parameter "name: type" pairs from a parameters node.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	params := node.ChildByFieldName("parameters")
	if params == nil {
		return nil
	}
	var result []string
	for i := range params.NamedChildCount() {
		param := params.NamedChild(i)
		switch param.Kind() {
		case "self_parameter", "variadic_parameter":
			result = append(result, param.Utf8Text(src))
		case "parameter":
			pattern := extract.FieldText(param, "pattern", src)
			if hasChildKind(param, "mutable_specifier") {
				pattern = "mut " + pattern
			}
			result = append(result, pattern+": "+extract.FieldText(param, "type", src))
		}
	}
	return result
}

// extractReturnType extracts the return type from a Rust function_item or function_signature_item.
func extractReturnType(node *sitter.Node, src []byte) string {
	return extract.FieldText(node, "return_type", src)
}

// moduleScope derives an item's lexical module path from its ancestors.
// No mutable scope is shared between files or recursive extraction calls.
func moduleScope(node *sitter.Node, src []byte) string {
	var names []string
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Kind() == "mod_item" {
			names = append(names, extract.FieldText(parent, "name", src))
		}
	}
	slices.Reverse(names)
	return strings.Join(names, ".")
}

// addItem records an item and its immediate module owner, returning its name
// for use as the scope of fields, methods and type parameters.
func addItem(node *sitter.Node, src []byte, c *plugin.Collector, sym *plugin.Symbol) string {
	parent := moduleScope(node, src)
	sym.ScopedName = plugin.MakeScopedName(parent, sym.Name)
	c.AddSymbol(sym)
	if parent != "" {
		c.AddEdge(&plugin.Edge{From: parent, To: sym.ScopedName, Kind: plugin.EdgeContains})
	}
	return sym.ScopedName
}

// ---------------------------------------------------------------------------
// Shared symbol builders
// ---------------------------------------------------------------------------

// extractSimpleRustSymbol emits a symbol that has a name, pub flag, category,
// and kind but no body to recurse into (enum, type alias, const).
// Mirrors extractSimpleScalaSymbol in the Scala plugin.
func extractSimpleRustSymbol(node *sitter.Node, src []byte, c *plugin.Collector, nameKind string, cat plugin.SymbolCategory, kind string) {
	name := childText(node, nameKind, src)
	if name == "" {
		return
	}
	addItem(node, src, c, &plugin.Symbol{
		Name:       name,
		Category:   cat,
		Kind:       kind,
		Signature:  name,
		Properties: plugin.NewProps().SetBool("pub", isPublic(node)).Map(),
		Span:       nodeSpan(node),
	})
}

// extractCallable emits a callable symbol (function or method) and, when
// parentName is non-empty, a contains edge from parent → scopedName.
// Shared by extractFunction and extractMethodFromBlock.
func extractCallable(node *sitter.Node, src []byte, c *plugin.Collector, kind, parentName string) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)
	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)

	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().
			SetBool("pub", isPublic(node)).
			SetBool("async", hasModifier(node, "async")).
			SetBool("unsafe", hasModifier(node, "unsafe")).
			Map(),
		Span: nodeSpan(node),
	})

	if parentName != "" {
		c.AddEdge(&plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, rustTypeParamConfig)
}

// extractTypeDecl emits a type symbol (struct or trait), recurses into its
// body via bodyKind, and optionally records an "unsafe" property.
// Shared by extractStruct and extractTrait.
func extractTypeDecl(node *sitter.Node, src []byte, c *plugin.Collector, kind, bodyKind string) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	props := plugin.NewProps().SetBool("pub", isPublic(node))
	if kind == "trait" {
		props.SetBool("unsafe", hasChildKind(node, "unsafe"))
	}

	scopedName := addItem(node, src, c, &plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       kind,
		Signature:  name,
		Properties: props.Map(),
		Span:       nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, rustTypeParamConfig)

	if body := childByKind(node, bodyKind); body != nil {
		if kind == "struct" {
			extractStructFields(body, src, c, scopedName)
		} else {
			extractBlockMethods(body, src, c, scopedName)
		}
	}
}

// ---------------------------------------------------------------------------
// Top-level extractors
// ---------------------------------------------------------------------------

// extractFunction extracts a function_item node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractCallable(node, src, c, "function", moduleScope(node, src))
}

// extractStruct extracts a struct_item node.
func extractStruct(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractTypeDecl(node, src, c, "struct", "field_declaration_list")
}

// extractEnum extracts an enum_item node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractSimpleRustSymbol(node, src, c, "type_identifier", plugin.CategoryType, "enum")
}

// extractTrait extracts a trait_item node.
func extractTrait(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractTypeDecl(node, src, c, "trait", "declaration_list")
}

// extractTypeAlias extracts a type_item node.
func extractTypeAlias(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractSimpleRustSymbol(node, src, c, "type_identifier", plugin.CategoryType, "type_alias")
}

// extractMod extracts a mod_item node.
func extractMod(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractSimpleRustSymbol(node, src, c, "identifier", plugin.CategoryModule, "module")
	if body := node.ChildByFieldName("body"); body != nil {
		plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{})
	}
}

// extractConst extracts a const_item node.
func extractConst(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractSimpleRustSymbol(node, src, c, "identifier", plugin.CategoryValue, "constant")
}

// extractStatic extracts a static_item node.
// Unlike extractConst, statics carry a "mutable" flag and no pub via extractSimpleRustSymbol
// because they use a different props shape.
func extractStatic(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}
	addItem(node, src, c, &plugin.Symbol{
		Name:      name,
		Category:  plugin.CategoryValue,
		Kind:      "variable",
		Signature: name,
		Properties: plugin.NewProps().
			SetBool("pub", isPublic(node)).
			SetBool("mutable", hasChildKind(node, "mutable_specifier")).
			Map(),
		Span: nodeSpan(node),
	})
}

// extractMacro extracts a macro_definition node.
// Macros have no pub visibility in tree-sitter-rust, so they don't use extractSimpleRustSymbol.
func extractMacro(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}
	addItem(node, src, c, &plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "macro",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})
}

// ---------------------------------------------------------------------------
// impl extraction
// ---------------------------------------------------------------------------

// extractImpl extracts an impl_item node.
func extractImpl(node *sitter.Node, src []byte, c *plugin.Collector) {
	typeName, traitName := implNames(node, src)
	if typeName == "" {
		return
	}

	scopedName := addItem(node, src, c, &plugin.Symbol{
		Name:       typeName,
		Category:   plugin.CategoryType,
		Kind:       "impl",
		Signature:  typeName,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	})

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, rustTypeParamConfig)

	if traitName != "" {
		c.AddEdge(&plugin.Edge{From: scopedName, To: traitName, Kind: plugin.EdgeImplements})
	}

	if body := childByKind(node, "declaration_list"); body != nil {
		extractBlockMethods(body, src, c, scopedName)
	}
}

// implNames extracts both the type name and the trait name from an impl_item
// in a single pass. For "impl Foo", returns ("Foo", ""). For "impl Trait for
// Foo", returns ("Foo", "Trait").
func implNames(node *sitter.Node, src []byte) (typeName, traitName string) {
	hasFor := false
	var firstType, lastType string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "for" {
			hasFor = true
		}
		if child.Kind() == "type_identifier" || child.Kind() == "generic_type" || child.Kind() == "scoped_type_identifier" {
			text := child.Utf8Text(src)
			if firstType == "" {
				firstType = text
			}
			lastType = text
		}
	}
	if hasFor {
		return lastType, firstType
	}
	return firstType, ""
}

// ---------------------------------------------------------------------------
// Block method extraction (impl / trait bodies)
// ---------------------------------------------------------------------------

// extractBlockMethods extracts function_item and function_signature_item children
// from a declaration_list (impl/trait body).
func extractBlockMethods(body *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "function_item", "function_signature_item":
			extractCallable(child, src, c, "method", parentName)
		}
	}
}

// extractStructFields extracts field_declaration children from a field_declaration_list.
func extractStructFields(body *sitter.Node, src []byte, c *plugin.Collector, structName string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil || child.Kind() != "field_declaration" {
			continue
		}
		fieldName := childText(child, "field_identifier", src)
		if fieldName != "" {
			// Rust permits a field and method with the same name. Keep their
			// internal names distinct so containment cannot resolve to the
			// method (or vice versa). Display names and signatures stay intact.
			scopedName := plugin.MakeScopedName(structName, fieldName) + "#field"
			c.AddSymbol(&plugin.Symbol{
				Name:       fieldName,
				ScopedName: scopedName,
				Category:   plugin.CategoryValue,
				Kind:       "field",
				Signature:  fieldName + ": " + extract.FieldText(child, "type", src),
				Properties: plugin.NewProps().SetBool("pub", isPublic(child)).Map(),
				Span:       nodeSpan(child),
			})
			c.AddEdge(&plugin.Edge{From: structName, To: scopedName, Kind: plugin.EdgeContains})
		}
	}
}

// ---------------------------------------------------------------------------
// Use / import extraction
// ---------------------------------------------------------------------------

// extractUseDecl extracts a use_declaration node and emits EdgeImports edges.
// Handles: use std::collections::HashMap;
//
//	use crate::models::User;
//	use super::helpers::format;
func extractUseDecl(node *sitter.Node, src []byte, c *plugin.Collector) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "scoped_identifier":
			// use std::collections::HashMap → From="HashMap", To="std::collections"
			text := child.Utf8Text(src)
			if last := plugin.LastSepIndex(text, "::"); last >= 0 {
				c.AddEdge(&plugin.Edge{From: text[last+2:], To: text[:last], Kind: plugin.EdgeImports})
			}
		case "use_as_clause":
			// use std::io::Result as IoResult → From="IoResult", To="std::io"
			path := childText(child, "scoped_identifier", src)
			alias := childText(child, "identifier", src)
			if path != "" && alias != "" {
				modulePath := path
				if last := plugin.LastSepIndex(path, "::"); last >= 0 {
					modulePath = path[:last]
				}
				c.AddEdge(&plugin.Edge{From: alias, To: modulePath, Kind: plugin.EdgeImports})
			}
		case "scoped_use_list":
			// use std::collections::{HashMap, BTreeMap}
			extractScopedUseList(child, src, c)
		case "use_wildcard":
			// use std::prelude::*
			text := child.Utf8Text(src)
			if last := plugin.LastSepIndex(text, "::"); last >= 0 {
				c.AddEdge(&plugin.Edge{From: "*", To: text[:last], Kind: plugin.EdgeImports})
			}
		case "identifier":
			// use HashMap; (unlikely but handle gracefully)
			name := child.Utf8Text(src)
			if name != "" {
				c.AddEdge(&plugin.Edge{From: name, To: name, Kind: plugin.EdgeImports})
			}
		}
	}
}

// extractScopedUseList handles use std::collections::{HashMap, BTreeMap}.
func extractScopedUseList(node *sitter.Node, src []byte, c *plugin.Collector) {
	var prefix string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "scoped_identifier" || child.Kind() == "identifier" {
			prefix = child.Utf8Text(src)
		}
		if child.Kind() == "use_list" {
			for j := range child.ChildCount() {
				item := child.Child(j)
				if item == nil {
					continue
				}
				if item.Kind() == "identifier" {
					if name := item.Utf8Text(src); name != "" {
						c.AddEdge(&plugin.Edge{From: name, To: prefix, Kind: plugin.EdgeImports})
					}
				} else if item.Kind() == "use_as_clause" {
					if alias := childText(item, "identifier", src); alias != "" {
						c.AddEdge(&plugin.Edge{From: alias, To: prefix, Kind: plugin.EdgeImports})
					}
				}
			}
		}
	}
}
