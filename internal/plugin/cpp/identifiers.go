// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cpp

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

// cppTypeKinds are the node kinds that represent types in C++.
var cppTypeKinds = []string{
	"primitive_type", "type_identifier", "sized_type_specifier",
	"template_type", "qualified_identifier", "auto",
	"struct_specifier", "enum_specifier",
}

// cppReturnStopKinds are node kinds that signal we've passed the return type in C++.
var cppReturnStopKinds = []string{"function_declarator", "pointer_declarator", "reference_declarator"}

// cppTypeParamConfig describes how C++ template parameter nodes are structured.
var cppTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "template_parameter_list",
	TypeParamKind:  "type_parameter_declaration",
	NameKind:       "type_identifier",
	ConstraintKind: "",
}

// callTarget is the CallTargetFunc for C++.
// Matches call_expression nodes. For simple identifiers returns the text directly.
// For qualified nodes (field_expression, qualified_identifier) extracts only the
// last identifier. For template_function, extracts the inner identifier name.
var callTarget plugin.CallTargetFunc = func(node *sitter.Node, src []byte) string {
	if node.Kind() != "call_expression" || node.ChildCount() == 0 {
		return ""
	}
	first := node.Child(0)
	if first == nil {
		return ""
	}
	switch first.Kind() {
	case "identifier":
		return first.Utf8Text(src)
	case "field_expression", "qualified_identifier":
		return plugin.LastNamedLeaf(first, src)
	case "template_function":
		return childText(first, "identifier", src)
	default:
		return ""
	}
}

// richCallTarget is the RichCallTargetFunc for C++.
var richCallTarget plugin.RichCallTargetFunc = func(node *sitter.Node, src []byte) plugin.CallTargetResult {
	if node.Kind() != "call_expression" || node.ChildCount() == 0 {
		return plugin.CallTargetResult{}
	}
	first := node.Child(0)
	if first == nil {
		return plugin.CallTargetResult{}
	}
	switch first.Kind() {
	case "identifier":
		return plugin.CallTargetResult{Name: first.Utf8Text(src)}
	case "field_expression", "qualified_identifier":
		return plugin.CallTargetResult{Name: plugin.LastNamedLeaf(first, src), Qualified: true}
	case "template_function":
		return plugin.CallTargetResult{Name: childText(first, "identifier", src)}
	default:
		return plugin.CallTargetResult{}
	}
}

// extractParamStrings extracts parameter "name: type" pairs from a function definition.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	decl := childByKind(node, "function_declarator")
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
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "parameter_declaration", "optional_parameter_declaration":
			name := childText(child, "identifier", src)
			if name == "" {
				continue
			}
			typeName := cppParamType(child, src)
			if typeName != "" {
				result = append(result, name+": "+typeName)
			} else {
				result = append(result, name)
			}
		}
	}
	return result
}

// cppParamType extracts the type from a C++ parameter_declaration node.
func cppParamType(node *sitter.Node, src []byte) string {
	return plugin.FirstChildTextByKinds(node, src, cppTypeKinds)
}

// extractReturnType extracts the return type from a C++ function_definition node.
func extractReturnType(node *sitter.Node, src []byte) string {
	return plugin.ReturnTypeByKinds(node, src, cppTypeKinds, cppReturnStopKinds)
}

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

// hasSpecifier checks if a node has a child with the given kind containing the given text.
func hasSpecifier(node *sitter.Node, src []byte, specKind, specText string) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == specKind && child.Utf8Text(src) == specText {
			return true
		}
	}
	return false
}

// hasVirtualSpecifier checks if a node has a "virtual" type qualifier or virtual function specifier.
func hasVirtualSpecifier(node *sitter.Node, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "virtual_function_specifier", "virtual":
			text := child.Utf8Text(src)
			if text == "virtual" {
				return true
			}
		case "type_qualifier":
			if child.Utf8Text(src) == "virtual" {
				return true
			}
		}
	}
	return false
}

// extractVisibility determines the access specifier for a class member.
// Walks siblings backwards from the member to find the most recent access_specifier.
// Returns "" if no access specifier is found.
func extractVisibility(node *sitter.Node, src []byte) string {
	parent := node.Parent()
	if parent == nil {
		return ""
	}

	// Find the index of this node among its siblings by comparing byte offsets.
	nodeStart := node.StartByte()
	nodeIdx := -1
	for i := range parent.ChildCount() {
		child := parent.Child(i)
		if child != nil && child.StartByte() == nodeStart && child.Kind() == node.Kind() {
			nodeIdx = int(i)
			break
		}
	}
	if nodeIdx < 0 {
		return ""
	}

	// Walk backwards to find the nearest access_specifier.
	for i := nodeIdx - 1; i >= 0; i-- {
		sib := parent.Child(uint(i))
		if sib == nil {
			continue
		}
		if sib.Kind() == "access_specifier" {
			// The access_specifier contains a keyword child (public, private, protected).
			for j := range sib.ChildCount() {
				gc := sib.Child(j)
				if gc != nil {
					t := gc.Utf8Text(src)
					if t == "public" || t == "private" || t == "protected" {
						return t
					}
				}
			}
			// Fallback: use the node text itself.
			return sib.Utf8Text(src)
		}
	}
	return ""
}

// extractFunction extracts a function_definition node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := funcDeclName(node, src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	isStatic, _ := storageClassProps(node, src)
	p := plugin.NewProps().
		SetBool("static", isStatic).
		SetBool("virtual", hasVirtualSpecifier(node, src)).
		SetBool("inline", hasSpecifier(node, src, "storage_class_specifier", "inline"))

	if parentClass != "" {
		p.Set("visibility", extractVisibility(node, src))
		p.SetBool("const", isConstMethod(node, src))
	}

	kind := "function"
	if parentClass != "" {
		kind = "method"
	}

	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: p.Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, cppTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})

		if hasOverrideSpecifier(node, src) {
			c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
		}
	}
}

// funcDeclName extracts the function name from a function_definition.
func funcDeclName(node *sitter.Node, src []byte) string {
	decl := childByKind(node, "function_declarator")
	if decl == nil {
		return ""
	}
	// Check for qualified_identifier (e.g., ClassName::method).
	if qi := childByKind(decl, "qualified_identifier"); qi != nil {
		return qi.Utf8Text(src)
	}
	// Check for field_identifier (method names).
	if name := childText(decl, "field_identifier", src); name != "" {
		return name
	}
	// Check for destructor_name.
	if dn := childByKind(decl, "destructor_name"); dn != nil {
		return dn.Utf8Text(src)
	}
	return childText(decl, "identifier", src)
}

// isConstMethod checks if a function definition has a const qualifier after the parameter list.
func isConstMethod(node *sitter.Node, src []byte) bool {
	return hasSpecifier(node, src, "type_qualifier", "const")
}

// isConstMethodDecl checks if a function_declarator has a trailing const qualifier.
func isConstMethodDecl(funcDecl *sitter.Node, src []byte) bool {
	return hasSpecifier(funcDecl, src, "type_qualifier", "const")
}

// hasOverrideSpecifier checks if a function has an "override" virtual specifier.
func hasOverrideSpecifier(node *sitter.Node, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "virtual_specifier" && child.Utf8Text(src) == "override" {
			return true
		}
	}
	// Also check inside the function_declarator.
	if decl := childByKind(node, "function_declarator"); decl != nil {
		for i := range decl.ChildCount() {
			child := decl.Child(i)
			if child == nil {
				continue
			}
			if child.Kind() == "virtual_specifier" && child.Utf8Text(src) == "override" {
				return true
			}
		}
	}
	return false
}

// extractClass extracts a class_specifier node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	p := plugin.NewProps()
	isAbstract := false

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "class",
		Signature:  name,
		Properties: p.Map(),
		Span:       nodeSpan(node),
	}

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, cppTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	extractBaseClasses(node, src, c, scopedName)

	if body := childByKind(node, "field_declaration_list"); body != nil {
		if scanForPureVirtual(body, src) {
			isAbstract = true
		}
		extractClassBody(body, src, c, scopedName)
	}

	p.SetBool("abstract", isAbstract)
	c.AddSymbol(&sym)
}

// extractStruct extracts a struct_specifier node.
func extractStruct(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "struct",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, cppTypeParamConfig)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}

	extractBaseClasses(node, src, c, scopedName)

	if body := childByKind(node, "field_declaration_list"); body != nil {
		extractClassBody(body, src, c, scopedName)
	}
}

// extractEnum extracts an enum_specifier node.
func extractEnum(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "enum",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractNamespace extracts a namespace_definition node.
func extractNamespace(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "namespace_identifier", src)
	if name == "" {
		name = childText(node, "identifier", src)
	}
	if name == "" {
		// Anonymous namespace — skip.
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryModule,
		Kind:       "namespace",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract namespace body.
	if body := childByKind(node, "declaration_list"); body != nil {
		for i := range body.ChildCount() {
			child := body.Child(i)
			if child == nil {
				continue
			}
			handlers.Dispatch(child, src, c, plugin.HandlerContext{})
		}
	}
}

// extractTemplate extracts a template_declaration node.
func extractTemplate(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	// The template_declaration wraps an inner declaration.
	// Extract the inner declaration, then extract type parameters from the
	// template_declaration and associate them with the inner symbol.
	var innerName string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			extractFunction(child, src, c, parentClass)
			innerName = funcDeclName(child, src)
			if parentClass != "" {
				innerName = plugin.MakeScopedName(parentClass, innerName)
			}
		case "class_specifier":
			extractClass(child, src, c, parentClass)
			innerName = childText(child, "type_identifier", src)
			if parentClass != "" {
				innerName = plugin.MakeScopedName(parentClass, innerName)
			}
		case "struct_specifier":
			extractStruct(child, src, c, parentClass)
			innerName = childText(child, "type_identifier", src)
			if parentClass != "" {
				innerName = plugin.MakeScopedName(parentClass, innerName)
			}
		case "declaration":
			extractDeclaration(child, src, c, parentClass)
		}
	}

	// Extract type parameters from the template_declaration and associate
	// them with the inner symbol.
	if innerName != "" {
		extract.TypeParams(node, src, c, innerName, cppTypeParamConfig)
	}
}

// extractDeclaration handles a top-level "declaration" node.
// This can be a function prototype or a global variable.
func extractDeclaration(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	// Check for function_declarator (prototype).
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "function_declarator":
			extractPrototype(node, child, src, c, parentClass)
			return
		case "init_declarator":
			if fd := childByKind(child, "function_declarator"); fd != nil {
				extractPrototype(node, fd, src, c, parentClass)
				return
			}
			extractGlobalVar(node, child, src, c, parentClass)
			return
		}
	}

	// Plain variable declaration.
	name := childText(node, "identifier", src)
	if name != "" {
		scopedName := plugin.MakeScopedName(parentClass, name)
		isStatic, isExtern := storageClassProps(node, src)
		sym := plugin.Symbol{
			Name:       name,
			ScopedName: scopedName,
			Category:   plugin.CategoryValue,
			Kind:       "variable",
			Signature:  name,
			Properties: plugin.NewProps().SetBool("static", isStatic).SetBool("extern", isExtern).Map(),
			Span:       nodeSpan(node),
		}
		c.AddSymbol(&sym)

		if parentClass != "" {
			c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
		}
	}
}

// extractPrototype extracts a function prototype from a declaration.
func extractPrototype(declNode, funcDecl *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(funcDecl, "identifier", src)
	if name == "" {
		name = childText(funcDecl, "field_identifier", src)
	}
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	isStatic, isExtern := storageClassProps(declNode, src)
	p := plugin.NewProps().
		SetBool("static", isStatic).
		SetBool("extern", isExtern).
		SetBool("virtual", hasVirtualSpecifier(declNode, src))

	kind := "function_prototype"
	if parentClass != "" {
		kind = "method"
		p.Set("visibility", extractVisibility(declNode, src))
	}

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  name,
		Properties: p.Map(),
		Span:       nodeSpan(declNode),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractGlobalVar extracts a global variable from an init_declarator.
func extractGlobalVar(declNode, initDecl *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(initDecl, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	isStatic, isExtern := storageClassProps(declNode, src)

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryValue,
		Kind:       "variable",
		Signature:  name,
		Properties: plugin.NewProps().SetBool("static", isStatic).SetBool("extern", isExtern).Map(),
		Span:       nodeSpan(declNode),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractFieldDeclaration handles a field_declaration inside a class/struct body.
func extractFieldDeclaration(node *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "function_declarator" {
			extractMethodDecl(node, child, src, c, parentClass)
			return
		}
	}

	name := childText(node, "field_identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	isStatic, _ := storageClassProps(node, src)
	p := plugin.NewProps().SetBool("static", isStatic)
	if parentClass != "" {
		p.Set("visibility", extractVisibility(node, src))
	}

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryValue,
		Kind:       "field",
		Signature:  name,
		Properties: p.Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})
	}
}

// extractMethodDecl extracts a method declaration from a field_declaration with a function_declarator.
func extractMethodDecl(node, funcDecl *sitter.Node, src []byte, c *plugin.Collector, parentClass string) {
	name := childText(funcDecl, "field_identifier", src)
	if name == "" {
		name = childText(funcDecl, "identifier", src)
	}
	if name == "" {
		if dn := childByKind(funcDecl, "destructor_name"); dn != nil {
			name = dn.Utf8Text(src)
		}
	}
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentClass, name)
	isStatic, _ := storageClassProps(node, src)
	p := plugin.NewProps().
		SetBool("static", isStatic).
		SetBool("virtual", hasVirtualSpecifier(node, src)).
		SetBool("const", isConstMethodDecl(funcDecl, src))

	if parentClass != "" {
		p.Set("visibility", extractVisibility(node, src))
	}

	p.SetBool("abstract", isPureVirtualDecl(node, src))

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "method",
		Signature:  name,
		Properties: p.Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	if parentClass != "" {
		c.AddEdge(plugin.Edge{From: parentClass, To: scopedName, Kind: plugin.EdgeContains})

		if hasOverrideInDecl(node, funcDecl, src) {
			c.AddEdge(plugin.Edge{From: scopedName, To: scopedName, Kind: plugin.EdgeOverrides})
		}
	}
}

// isPureVirtualDecl checks if a field_declaration is a pure virtual declaration (= 0).
func isPureVirtualDecl(node *sitter.Node, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "number_literal" && child.Utf8Text(src) == "0" {
			// Check that there's a preceding "=" sign — this is "= 0" pattern.
			if i > 0 {
				prev := node.Child(i - 1)
				if prev != nil && prev.Utf8Text(src) == "=" {
					return true
				}
			}
		}
	}
	return false
}

// hasOverrideInDecl checks for override specifier in a field_declaration or its function_declarator.
func hasOverrideInDecl(node, funcDecl *sitter.Node, src []byte) bool {
	// Check in the field_declaration node.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "virtual_specifier" && child.Utf8Text(src) == "override" {
			return true
		}
	}
	// Check in the function_declarator.
	for i := range funcDecl.ChildCount() {
		child := funcDecl.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "virtual_specifier" && child.Utf8Text(src) == "override" {
			return true
		}
	}
	return false
}

// extractBaseClasses extracts base class specifiers and produces inherits edges.
func extractBaseClasses(node *sitter.Node, src []byte, c *plugin.Collector, className string) {
	bcl := childByKind(node, "base_class_clause")
	if bcl == nil {
		return
	}
	for i := range bcl.ChildCount() {
		child := bcl.Child(i)
		if child == nil {
			continue
		}
		// base_class_clause contains type_identifier children for each base class.
		switch child.Kind() {
		case "type_identifier":
			c.AddEdge(plugin.Edge{From: className, To: child.Utf8Text(src), Kind: plugin.EdgeInherits})
		case "qualified_identifier":
			c.AddEdge(plugin.Edge{From: className, To: child.Utf8Text(src), Kind: plugin.EdgeInherits})
		}
	}
}

// extractClassBody extracts members from a field_declaration_list (class/struct body).
func extractClassBody(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	plugin.WalkChildren(body, src, c, handlers, plugin.HandlerContext{ParentName: className})
}

// scanForPureVirtual checks if a class body contains any pure virtual method declarations.
func scanForPureVirtual(body *sitter.Node, src []byte) bool {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil || child.Kind() != "field_declaration" {
			continue
		}
		if isPureVirtualDecl(child, src) {
			return true
		}
	}
	return false
}

// extractInclude extracts a preproc_include (#include) node as a references edge.
func extractInclude(node *sitter.Node, src []byte, c *plugin.Collector) {
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
