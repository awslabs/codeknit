// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package golang

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

// callTarget is the CallTargetFunc for Go.
// Matches call_expression nodes whose first child is an identifier or selector_expression.
// For selector_expression (e.g. pkg.Func), extracts only the field name ("Func")
// so the planner can resolve it against known symbols.
var callTarget = plugin.UnqualifiedCallTarget("call_expression",
	[]string{"identifier"},
	[]string{"selector_expression"},
)

// richCallTarget is the RichCallTargetFunc for Go.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call_expression",
	[]string{"identifier"},
	[]string{"selector_expression"},
)

// goTypeParamConfig describes how Go type parameter nodes are structured.
var goTypeParamConfig = extract.TypeParamConfig{
	TypeParamsKind: "type_parameter_list",
	TypeParamKind:  "type_parameter_declaration",
	NameKind:       "identifier",
	ConstraintKind: "constraint",
}

// extractParamStrings extracts function parameter "name: type" pairs from a function_declaration.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	paramList := childByKind(node, "parameter_list")
	if paramList == nil {
		return nil
	}
	return extractParamPairs(paramList, src)
}

// extractMethodParamStrings extracts method parameter "name: type" pairs, skipping the receiver parameter_list.
// In Go, method_declaration has two parameter_lists: first is receiver, second is params.
func extractMethodParamStrings(node *sitter.Node, src []byte) []string {
	count := 0
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil && child.Kind() == "parameter_list" {
			count++
			if count == 2 {
				return extractParamPairs(child, src)
			}
		}
	}
	return nil
}

// extractParamPairs extracts "name: type" pairs from a parameter_list node.
func extractParamPairs(paramList *sitter.Node, src []byte) []string {
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
		typeNode := lastTypedChild(child)
		if typeNode != nil {
			result = append(result, name+": "+typeNode.Utf8Text(src))
		} else {
			result = append(result, name)
		}
	}
	return result
}

// extractReturnType extracts the return type string from a Go function/method declaration.
// Go return types appear in a result node after the parameter lists.
func extractReturnType(node *sitter.Node, src []byte) string {
	// Look for a result type: can be a simple type or a parameter_list (multiple returns).
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "type_identifier", "pointer_type", "slice_type", "array_type",
			"map_type", "channel_type", "interface_type", "struct_type",
			"qualified_type", "function_type":
			// Check this comes after the parameter lists (not the receiver type).
			// In Go AST, result types appear after the last parameter_list.
			return child.Utf8Text(src)
		}
	}
	return ""
}

// isExported returns true if the Go identifier starts with an uppercase letter.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// extractFunction extracts a function_declaration node.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	returnType := extractReturnType(node, src)

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().SetBool("exported", isExported(name)).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters
	extract.TypeParams(node, src, c, name, goTypeParamConfig)
}

// extractMethod extracts a method_declaration node.
func extractMethod(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "field_identifier", src)
	if name == "" {
		return
	}

	receiver := extractReceiver(node, src)
	recType := receiverTypeName(receiver)
	scopedName := plugin.MakeScopedName(recType, name)
	params := extractMethodParamStrings(node, src)
	returnType := extractReturnType(node, src)

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       "method",
		Signature:  buildFuncSignature(name, params, returnType),
		Properties: plugin.NewProps().SetBool("exported", isExported(name)).Set("receiver", receiver).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Add contains edge from receiver type to method.
	if recType != "" {
		c.AddEdge(plugin.Edge{
			From: recType,
			To:   scopedName,
			Kind: plugin.EdgeContains,
		})
	}

	// Extract type parameters
	extract.TypeParams(node, src, c, scopedName, goTypeParamConfig)
}

// extractReceiver extracts the receiver string from a method_declaration.
// Returns e.g. "*Greeter" or "Greeter".
func extractReceiver(node *sitter.Node, src []byte) string {
	paramList := childByKind(node, "parameter_list")
	if paramList == nil {
		return ""
	}
	// The first parameter_declaration in the parameter_list is the receiver.
	for i := range paramList.ChildCount() {
		child := paramList.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "parameter_declaration" {
			// The type is the last named child (could be pointer_type or type_identifier).
			typeNode := lastTypedChild(child)
			if typeNode != nil {
				return typeNode.Utf8Text(src)
			}
		}
	}
	return ""
}

// lastTypedChild returns the last child that represents a type in a parameter_declaration.
func lastTypedChild(node *sitter.Node) *sitter.Node {
	var last *sitter.Node
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "pointer_type", "type_identifier", "qualified_type", "slice_type",
			"array_type", "map_type", "channel_type", "interface_type", "struct_type":
			last = child
		}
	}
	return last
}

// receiverTypeName strips the pointer prefix from a receiver type string.
func receiverTypeName(receiver string) string {
	if receiver == "" {
		return ""
	}
	if receiver[0] == '*' {
		return receiver[1:]
	}
	return receiver
}

// extractTypeDecl routes a type_declaration to struct/interface/type_alias.
func extractTypeDecl(node *sitter.Node, src []byte, c *plugin.Collector) {
	// type_declaration contains type_spec or type_alias children.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "type_spec":
			extractTypeSpec(child, src, c)
		case "type_alias":
			extractTypeAlias(child, src, c)
		}
	}
}

// extractTypeSpec extracts a single type_spec node.
func extractTypeSpec(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	// Determine the inner type to decide kind.
	var kind string
	var innerNode *sitter.Node
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "struct_type":
			kind = "struct"
			innerNode = child
		case "interface_type":
			kind = "interface"
			innerNode = child
		case "type_identifier", "pointer_type", "slice_type", "array_type",
			"map_type", "channel_type", "function_type", "qualified_type":
			if kind == "" {
				kind = "type_alias"
			}
		}
	}
	if kind == "" {
		kind = "type_alias"
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       kind,
		Signature:  name,
		Properties: plugin.NewProps().SetBool("exported", isExported(name)).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract type parameters from the type_spec node.
	extract.TypeParams(node, src, c, name, goTypeParamConfig)

	// Extract contains edges for struct fields or interface method specs.
	if innerNode != nil {
		switch kind {
		case "struct":
			extractStructFields(innerNode, src, c, name)
		case "interface":
			extractInterfaceSpecs(innerNode, src, c, name)
		}
	}
}

// extractTypeAlias extracts a type_alias node (e.g., `type MyAlias = int`).
func extractTypeAlias(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "type_identifier", src)
	if name == "" {
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       "type_alias",
		Signature:  name,
		Properties: plugin.NewProps().SetBool("exported", isExported(name)).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// extractStructFields extracts field_declaration children from a struct_type.
func extractStructFields(structNode *sitter.Node, src []byte, c *plugin.Collector, structName string) {
	fieldList := childByKind(structNode, "field_declaration_list")
	if fieldList == nil {
		return
	}
	for i := range fieldList.ChildCount() {
		child := fieldList.Child(i)
		if child == nil || child.Kind() != "field_declaration" {
			continue
		}
		fieldName := childText(child, "field_identifier", src)
		if fieldName != "" {
			scopedName := plugin.MakeScopedName(structName, fieldName)
			// Emit field as a symbol only during fingerprinting so Types() can
			// encode field names into the structural fingerprint. During normal
			// parse runs the field symbol is not emitted — only the edge is.
			if c.Fingerprint {
				c.AddSymbol(&plugin.Symbol{
					Name:       fieldName,
					ScopedName: scopedName,
					Category:   plugin.CategoryValue,
					Kind:       "field",
					Signature:  fieldName,
					Properties: plugin.NewProps().Map(),
					Span:       nodeSpan(child),
				})
			}
			c.AddEdge(plugin.Edge{
				From: structName,
				To:   scopedName,
				Kind: plugin.EdgeContains,
			})
		}
	}
}

// extractInterfaceSpecs extracts method_elem children from an interface_type.
func extractInterfaceSpecs(ifaceNode *sitter.Node, src []byte, c *plugin.Collector, ifaceName string) {
	// interface_type has method_elem or type_identifier (embedding) children.
	for i := range ifaceNode.ChildCount() {
		child := ifaceNode.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "method_elem":
			specName := childText(child, "field_identifier", src)
			if specName != "" {
				scopedName := plugin.MakeScopedName(ifaceName, specName)
				c.AddEdge(plugin.Edge{
					From: ifaceName,
					To:   scopedName,
					Kind: plugin.EdgeContains,
				})
			}
		case "type_identifier":
			// Interface embedding → implements edge.
			embeddedName := child.Utf8Text(src)
			if embeddedName != "" {
				c.AddEdge(plugin.Edge{
					From: ifaceName,
					To:   embeddedName,
					Kind: plugin.EdgeImplements,
				})
			}
		}
	}
}

// extractPackage extracts a package_clause node.
func extractPackage(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "package_identifier", src)
	if name == "" {
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryModule,
		Kind:       "package",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// extractDeclGroup is the shared implementation for var_declaration and
// const_declaration, which both iterate spec children and emit value symbols.
func extractDeclGroup(node *sitter.Node, src []byte, c *plugin.Collector, specKind, symKind string) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || child.Kind() != specKind {
			continue
		}
		name := childText(child, "identifier", src)
		if name == "" {
			continue
		}
		c.AddSymbol(&plugin.Symbol{
			Name:       name,
			Category:   plugin.CategoryValue,
			Kind:       symKind,
			Signature:  name,
			Properties: plugin.NewProps().SetBool("exported", isExported(name)).Map(),
			Span:       nodeSpan(child),
		})
	}
}

// extractVarDecl extracts a var_declaration node (may contain multiple var_spec).
func extractVarDecl(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractDeclGroup(node, src, c, "var_spec", "variable")
}

// extractConstDecl extracts a const_declaration node (may contain multiple const_spec).
func extractConstDecl(node *sitter.Node, src []byte, c *plugin.Collector) {
	extractDeclGroup(node, src, c, "const_spec", "constant")
}

// extractImports extracts import_declaration nodes and emits EdgeImports edges.
// For Go, import paths like "fmt" or "github.com/user/pkg" are emitted with
// the last path segment as the local name (the package alias in Go).
// Aliased imports use the alias as the local name.
func extractImports(node *sitter.Node, src []byte, c *plugin.Collector) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "import_spec" {
			extractImportSpec(child, src, c)
		} else if child.Kind() == "import_spec_list" {
			for j := range child.ChildCount() {
				spec := child.Child(j)
				if spec != nil && spec.Kind() == "import_spec" {
					extractImportSpec(spec, src, c)
				}
			}
		}
	}
}

// extractImportSpec extracts a single import_spec node.
func extractImportSpec(node *sitter.Node, src []byte, c *plugin.Collector) {
	// import_spec children: optional package_identifier (alias), interpreted_string_literal (path)
	var alias, path string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "package_identifier":
			alias = child.Utf8Text(src)
		case "interpreted_string_literal":
			// Strip quotes.
			raw := child.Utf8Text(src)
			if len(raw) >= 2 {
				path = raw[1 : len(raw)-1]
			}
		case "dot":
			alias = "."
		case "blank_identifier":
			alias = "_"
		}
	}

	if path == "" {
		return
	}

	// Skip blank imports (side-effect only).
	if alias == "_" {
		return
	}

	// Determine the local name: alias if present, otherwise last path segment.
	localName := alias
	if localName == "" || localName == "." {
		// Use last segment of the import path as the package name.
		if idx := plugin.LastSepIndex(path, "/"); idx >= 0 {
			localName = path[idx+1:]
		} else {
			localName = path
		}
	}

	c.AddEdge(plugin.Edge{
		From: localName,
		To:   path,
		Kind: plugin.EdgeImports,
	})
}
