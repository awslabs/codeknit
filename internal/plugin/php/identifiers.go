// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package php

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

// phpParamTypeKinds are the node kinds that represent type hints in PHP parameters.
var phpParamTypeKinds = []string{
	"named_type", "primitive_type", "optional_type",
	"union_type", "intersection_type", "nullable_type",
}

// phpReturnTypeKinds are the valid type node kinds after ":" in a PHP function.
var phpReturnTypeKinds = []string{
	"named_type", "primitive_type", "optional_type",
	"union_type", "intersection_type", "nullable_type",
}

// callTargetName extracts the function name from a function_call_expression.
func callTargetName(node *sitter.Node, src []byte) string {
	// function_call_expression: function (name or qualified_name), arguments
	if qn := childByKind(node, "qualified_name"); qn != nil {
		return qn.Utf8Text(src)
	}
	if n := childByKind(node, "name"); n != nil {
		return n.Utf8Text(src)
	}
	return ""
}

// memberCallName extracts "obj->method" from a member_call_expression.
func memberCallName(node *sitter.Node, src []byte) string {
	name := childText(node, "name", src)
	return name
}

// scopedCallName extracts "Class::method" from a scoped_call_expression.
func scopedCallName(node *sitter.Node, src []byte) string {
	name := childText(node, "name", src)
	return name
}

// stripDollar removes the leading "$" from a PHP variable name
// so that emitted symbols use a uniform naming convention across all languages.
func stripDollar(name string) string {
	if name != "" && name[0] == '$' {
		return name[1:]
	}
	return name
}

// extractParamStrings extracts parameter "name: type" pairs from a formal_parameters node.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	params := childByKind(node, "formal_parameters")
	if params == nil {
		return nil
	}
	var result []string
	for i := range params.ChildCount() {
		child := params.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "simple_parameter" || child.Kind() == "variadic_parameter" || child.Kind() == "property_promotion_parameter" {
			vn := childByKind(child, "variable_name")
			if vn == nil {
				continue
			}
			name := stripDollar(vn.Utf8Text(src))
			typeName := phpParamType(child, src)
			if typeName != "" {
				result = append(result, name+": "+typeName)
			} else {
				result = append(result, name)
			}
		}
	}
	return result
}

// phpParamType extracts the type hint from a PHP parameter node.
func phpParamType(node *sitter.Node, src []byte) string {
	return plugin.FirstChildTextByKinds(node, src, phpParamTypeKinds)
}

// extractReturnType extracts the return type hint from a PHP function/method declaration.
func extractReturnType(node *sitter.Node, src []byte) string {
	return plugin.ReturnTypeAfterToken(node, src, ":", phpReturnTypeKinds)
}

// extractQualifiedName extracts a type name from a node (base_clause etc.).
func extractQualifiedName(node *sitter.Node, src []byte) string {
	if qn := childByKind(node, "qualified_name"); qn != nil {
		return childText(qn, "name", src)
	}
	if n := childByKind(node, "name"); n != nil {
		return n.Utf8Text(src)
	}
	return ""
}

// extractNameList extracts a list of type names from a node
// (class_interface_clause, base_clause with multiple names).
func extractNameList(node *sitter.Node, src []byte) []string {
	var names []string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "qualified_name":
			name := childText(child, "name", src)
			if name != "" {
				names = append(names, name)
			}
		case "name":
			names = append(names, child.Utf8Text(src))
		}
	}
	return names
}

// copyProps creates a shallow copy of a properties map.
func copyProps(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// callTarget is the CallTargetFunc for PHP.
// It handles function_call_expression, member_call_expression, and scoped_call_expression.
func callTarget(node *sitter.Node, src []byte) string {
	switch node.Kind() {
	case "function_call_expression":
		return callTargetName(node, src)
	case "member_call_expression":
		return memberCallName(node, src)
	case "scoped_call_expression":
		return scopedCallName(node, src)
	}
	return ""
}

// extractModifiers collects modifier keywords from a declaration node.
// PHP modifiers include: public, private, protected, static, abstract, final, readonly.
func extractModifiers(node *sitter.Node, src []byte) map[string]string {
	p := plugin.NewProps()
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "visibility_modifier":
			p.Set("visibility", child.Utf8Text(src))
		case "static_modifier":
			p.SetBool("static", true)
		case "abstract_modifier":
			p.SetBool("abstract", true)
		case "final_modifier":
			p.SetBool("final", true)
		case "readonly_modifier":
			p.SetBool("readonly", true)
		case "var_modifier":
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				switch gc.Kind() {
				case "visibility_modifier":
					p.Set("visibility", gc.Utf8Text(src))
				case "static_modifier":
					p.SetBool("static", true)
				case "readonly_modifier":
					p.SetBool("readonly", true)
				}
			}
		}
	}
	return p.Map()
}

// extractNamespace extracts a namespace_definition node.
func extractNamespace(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "namespace_name", src)
	if name == "" {
		name = childText(node, "name", src)
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
	if body := childByKind(node, "compound_statement"); body != nil {
		extractBody(body, src, c, "")
	}
	if body := childByKind(node, "declaration_list"); body != nil {
		extractBody(body, src, c, "")
	}
}

// extractFunction extracts a function_definition node (top-level function).
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "name", src)
	if name == "" {
		return
	}
	scopedName := plugin.MakeScopedName(parentClass, name)
	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// phpTypeDeclConfig describes the optional heritage and body extraction
// for a PHP type declaration (class, interface, trait, enum).
type phpTypeDeclConfig struct {
	kind       string // symbol kind: "class", "interface", "trait", "enum"
	extends    string // child kind for single-extends (e.g. "base_clause"), "" to skip
	implements string // child kind for implements list (e.g. "class_interface_clause"), "" to skip
	body       string // child kind for body (e.g. "declaration_list"), "" to skip
	useMods    bool   // whether to extract modifier keywords
	extendsAll bool   // true = emit all names from extends as EdgeInherits (interface), false = first only (class)
}

// extractPHPTypeDecl is the shared implementation for extractClass, extractInterface,
// extractTrait, and extractEnum. All follow the same scaffold: name → scoped →
// AddSymbol → contains edge → optional heritage edges → optional body walk.
func extractPHPTypeDecl(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string, cfg *phpTypeDeclConfig) {
	name := childText(node, "name", src)
	if name == "" {
		return
	}
	scopedName := plugin.MakeScopedName(parentClass, name)

	var props map[string]string
	if cfg.useMods {
		props = extractModifiers(node, src)
	} else {
		props = plugin.NewProps().Map()
	}

	c.AddSymbol(&plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       cfg.kind,
		Signature:  name,
		Properties: props,
		Span:       nodeSpan(node),
	})
	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	// Heritage: extends (single base or list).
	if cfg.extends != "" {
		if bc := childByKind(node, cfg.extends); bc != nil {
			if cfg.extendsAll {
				for _, base := range extractNameList(bc, src) {
					c.AddEdge(plugin.Edge{From: scopedName, To: base, Kind: plugin.EdgeInherits})
				}
			} else {
				if baseName := extractQualifiedName(bc, src); baseName != "" {
					c.AddEdge(plugin.Edge{From: scopedName, To: baseName, Kind: plugin.EdgeInherits})
				}
			}
		}
	}

	// Heritage: implements.
	if cfg.implements != "" {
		if ci := childByKind(node, cfg.implements); ci != nil {
			for _, iface := range extractNameList(ci, src) {
				c.AddEdge(plugin.Edge{From: scopedName, To: iface, Kind: plugin.EdgeImplements})
			}
		}
	}

	// Body members.
	if cfg.body != "" {
		if body := childByKind(node, cfg.body); body != nil {
			extractBody(body, src, c, scopedName)
		}
	}
}

// extractClass extracts a class_declaration node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractPHPTypeDecl(node, src, c, parentClass, &phpTypeDeclConfig{
		kind:       "class",
		useMods:    true,
		extends:    "base_clause",
		implements: "class_interface_clause",
		body:       "declaration_list",
	})
}

// extractInterface extracts an interface_declaration node.
func extractInterface(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractPHPTypeDecl(node, src, c, parentClass, &phpTypeDeclConfig{
		kind:       "interface",
		useMods:    true,
		extends:    "base_clause",
		extendsAll: true,
		body:       "declaration_list",
	})
}

// extractTrait extracts a trait_declaration node.
func extractTrait(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractPHPTypeDecl(node, src, c, parentClass, &phpTypeDeclConfig{
		kind: "trait",
		body: "declaration_list",
	})
}

// extractEnum extracts an enum_declaration node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	extractPHPTypeDecl(node, src, c, parentClass, &phpTypeDeclConfig{
		kind:       "enum",
		useMods:    true,
		implements: "class_interface_clause",
	})
}

// extractMethod extracts a method_declaration node.
func extractMethod(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "name", src)
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
	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractProperty extracts a property_declaration node.
func extractProperty(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	mods := extractModifiers(node, src)
	// property_declaration contains property_element children with variable_name.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "property_element" {
			vn := childByKind(child, "variable_name")
			if vn == nil {
				continue
			}
			name := stripDollar(vn.Utf8Text(src))
			if name == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentClass, name)
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				ScopedName: scopedName,
				Category:   plugin.CategoryValue,
				Kind:       "property",
				Signature:  name,
				Properties: copyProps(mods),
				Span:       nodeSpan(node),
			})
			if parentClass != "" {
				c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
			}
		}
	}
}

// extractConst extracts a const_declaration node.
func extractConst(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	mods := extractModifiers(node, src)
	// const_declaration contains const_element children with name.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "const_element" {
			name := childText(child, "name", src)
			if name == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentClass, name)
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				ScopedName: scopedName,
				Category:   plugin.CategoryValue,
				Kind:       "constant",
				Signature:  name,
				Properties: copyProps(mods),
				Span:       nodeSpan(node),
			})
			if parentClass != "" {
				c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
			}
		}
	}
}

// extractBody extracts members from a declaration_list or compound_statement.
func extractBody(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{ParentName: className})
}

// extractUseDeclaration extracts a namespace_use_declaration node and emits EdgeImports edges.
// Handles: use App\Models\User;
//
//	use App\Models\User as UserModel;
//	use App\Models\{User, Order};
func extractUseDeclaration(node *sitter.Node, src []byte, c *plugin.Collector) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "namespace_use_clause":
			extractUseClause(child, src, c)
		case "namespace_use_group":
			extractUseGroup(child, src, c)
		}
	}
}

// extractUseClause handles a single use clause like "App\Models\User" or "App\Models\User as UserModel".
func extractUseClause(node *sitter.Node, src []byte, c *plugin.Collector) {
	// qualified_name holds the full path, optional namespace_aliasing_clause holds the alias.
	var fullPath, alias string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "qualified_name":
			fullPath = child.Utf8Text(src)
		case "name":
			if fullPath == "" {
				fullPath = child.Utf8Text(src)
			}
		case "namespace_aliasing_clause":
			alias = childText(child, "name", src)
		}
	}
	if fullPath == "" {
		return
	}

	// Split into module path and local name at the last backslash.
	localName := alias
	modulePath := fullPath
	if last := plugin.LastSepIndex(fullPath, "\\"); last >= 0 {
		if localName == "" {
			localName = fullPath[last+1:]
		}
		modulePath = fullPath[:last]
	} else if localName == "" {
		localName = fullPath
	}

	c.AddEdge(plugin.Edge{From: localName, To: modulePath, Kind: plugin.EdgeImports})
}

// extractUseGroup handles grouped use like "App\Models\{User, Order}".
func extractUseGroup(node *sitter.Node, src []byte, c *plugin.Collector) {
	// Find the prefix (qualified_name before the group).
	var prefix string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "qualified_name" || child.Kind() == "name" {
			prefix = child.Utf8Text(src)
		}
		if child.Kind() == "namespace_use_group_clause" {
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				if gc.Kind() == "namespace_use_clause" {
					// Each clause in the group is relative to the prefix.
					extractUseClauseWithPrefix(gc, src, c, prefix)
				}
			}
		}
	}
}

// extractUseClauseWithPrefix handles a use clause within a group, prepending the prefix.
func extractUseClauseWithPrefix(node *sitter.Node, src []byte, c *plugin.Collector, prefix string) {
	var name, alias string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "qualified_name", "name":
			name = child.Utf8Text(src)
		case "namespace_aliasing_clause":
			alias = childText(child, "name", src)
		}
	}
	if name == "" {
		return
	}

	localName := alias
	if localName == "" {
		localName = name
	}

	c.AddEdge(plugin.Edge{From: localName, To: prefix, Kind: plugin.EdgeImports})
}
