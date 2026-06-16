// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package python

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

// callTarget is the CallTargetFunc for Python.
// Matches call nodes whose first child is an identifier or attribute.
// For attribute (e.g. os.path.join), extracts only the method name ("join").
var callTarget = plugin.UnqualifiedCallTarget("call",
	[]string{"identifier"},
	[]string{"attribute"},
)

// richCallTarget is the RichCallTargetFunc for Python.
var richCallTarget = plugin.UnqualifiedCallTargetRich("call",
	[]string{"identifier"},
	[]string{"attribute"},
)

// extractParamStrings extracts function parameter names as strings.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	params := childByKind(node, "parameters")
	if params == nil {
		return nil
	}
	var result []string
	for i := range params.ChildCount() {
		child := params.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier":
			name := child.Utf8Text(src)
			if name != "self" && name != "cls" {
				result = append(result, name)
			}
		case "typed_parameter":
			name := childText(child, "identifier", src)
			if name != "" && name != "self" && name != "cls" {
				result = append(result, name)
			}
		case "default_parameter":
			name := childText(child, "identifier", src)
			if name != "" && name != "self" && name != "cls" {
				result = append(result, name)
			}
		case "typed_default_parameter":
			name := childText(child, "identifier", src)
			if name != "" && name != "self" && name != "cls" {
				result = append(result, name)
			}
		}
	}
	return result
}

// hasDecorator checks if a decorated_definition or function has a specific decorator.
func hasDecorator(node *sitter.Node, decoratorName string, src []byte) bool {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "decorator" {
			// The decorator node contains an identifier or attribute child.
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc != nil && gc.Kind() == "identifier" && gc.Utf8Text(src) == decoratorName {
					return true
				}
			}
		}
	}
	return false
}

// pyDecoratorName is the DecoratorNameFunc for Python.
// Python uses "attribute" for qualified decorators (@app.route) and "call" for
// decorators with arguments (@app.route("/path")).
var pyDecoratorName = plugin.MakeDecoratorNameFunc("attribute", "call")

// extractDecoratorEdges emits EdgeDecorates edges from each decorator name to
// the decorated symbol. The decoratedParent is the decorated_definition node.
func extractDecoratorEdges(decoratedParent *sitter.Node, src []byte, c *plugin.Collector, targetName string) {
	plugin.ExtractDecoratorEdges(decoratedParent, src, c, targetName, "decorator", pyDecoratorName)
}

// extractFunction extracts a function_definition node.
// decoratedParent is optional; if provided, decorators are inspected from it.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector, isMethod bool, parentName string, decoratedParent ...*sitter.Node) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)
	params := extractParamStrings(node, src)
	isAsync := childText(node, "async", src) != ""

	p := plugin.NewProps().SetBool("async", isAsync)

	kind := "function"
	if isMethod {
		kind = "method"
		// Check for @staticmethod or @classmethod decorators.
		for _, dp := range decoratedParent {
			if hasDecorator(dp, "staticmethod", src) {
				p.SetBool("static", true)
			}
			if hasDecorator(dp, "classmethod", src) {
				p.SetBool("classmethod", true)
			}
		}
	}

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryCallable,
		Kind:       kind,
		Signature:  buildFuncSignature(name, params, ""),
		Properties: p.Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Emit decorator edges from the decorated_definition parent.
	for _, dp := range decoratedParent {
		extractDecoratorEdges(dp, src, c, scopedName)
	}
}

// extractClass extracts a class_definition node.
// decoratedParent is optional; if provided, decorator edges are emitted.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector, decoratedParent ...*sitter.Node) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryType,
		Kind:       "class",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Emit decorator edges.
	for _, dp := range decoratedParent {
		extractDecoratorEdges(dp, src, c, name)
	}

	// Extract inheritance edges from argument_list (superclasses).
	if argList := childByKind(node, "argument_list"); argList != nil {
		for i := range argList.ChildCount() {
			child := argList.Child(i)
			if child == nil {
				continue
			}
			if child.Kind() == "identifier" {
				c.AddEdge(plugin.Edge{
					From: name,
					To:   child.Utf8Text(src),
					Kind: plugin.EdgeInherits,
				})
			}
		}
	}

	// Extract class body members.
	if body := childByKind(node, "block"); body != nil {
		extractClassMembers(body, src, c, name)
	}
}

// extractClassMembers extracts methods and nested definitions from a class body.
func extractClassMembers(body *sitter.Node, src []byte, c *plugin.Collector, className string) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			extractFunction(child, src, c, true, className)
			methodName := childText(child, "identifier", src)
			if methodName != "" {
				scopedName := plugin.MakeScopedName(className, methodName)
				c.AddEdge(plugin.Edge{
					From: className,
					To:   scopedName,
					Kind: plugin.EdgeContains,
				})
			}
		case "decorated_definition":
			// Unwrap decorated methods or classes in class body.
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc == nil {
					continue
				}
				switch gc.Kind() {
				case "function_definition":
					extractFunction(gc, src, c, true, className, child)
					methodName := childText(gc, "identifier", src)
					if methodName != "" {
						scopedName := plugin.MakeScopedName(className, methodName)
						c.AddEdge(plugin.Edge{
							From: className,
							To:   scopedName,
							Kind: plugin.EdgeContains,
						})
					}
				case "class_definition":
					extractNestedClass(gc, src, c, className, child)
				}
			}
		case "class_definition":
			// Nested class without decorators.
			extractNestedClass(child, src, c, className)
		}
	}
}

// extractNestedClass extracts a class_definition nested inside another class.
func extractNestedClass(node *sitter.Node, src []byte, c *plugin.Collector, parentName string, decoratedParent ...*sitter.Node) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	scopedName := plugin.MakeScopedName(parentName, name)

	sym := plugin.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   plugin.CategoryType,
		Kind:       "class",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Emit decorator edges.
	for _, dp := range decoratedParent {
		extractDecoratorEdges(dp, src, c, scopedName)
	}

	// Contains edge from parent class.
	c.AddEdge(plugin.Edge{From: parentName, To: scopedName, Kind: plugin.EdgeContains})

	// Extract inheritance edges.
	if argList := childByKind(node, "argument_list"); argList != nil {
		for i := range argList.ChildCount() {
			child := argList.Child(i)
			if child != nil && child.Kind() == "identifier" {
				c.AddEdge(plugin.Edge{From: scopedName, To: child.Utf8Text(src), Kind: plugin.EdgeInherits})
			}
		}
	}

	// Extract nested class body members.
	if body := childByKind(node, "block"); body != nil {
		extractClassMembers(body, src, c, scopedName)
	}
}

// extractAssignment extracts top-level assignment expressions as variable symbols.
func extractAssignment(node *sitter.Node, src []byte, c *plugin.Collector) {
	// expression_statement wraps an assignment node.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "assignment" {
			// Check for tuple/list unpacking first: a, b = 1, 2
			// Must check before identifier because pattern_list contains identifiers.
			if pat := childByKind(child, "pattern_list"); pat != nil {
				extractPythonPatternNames(pat, src, c, child)
				continue
			}
			if tup := childByKind(child, "tuple_pattern"); tup != nil {
				extractPythonPatternNames(tup, src, c, child)
				continue
			}
			if lst := childByKind(child, "list_pattern"); lst != nil {
				extractPythonPatternNames(lst, src, c, child)
				continue
			}
			// Simple identifier assignment: name = "hello"
			name := childText(child, "identifier", src)
			if name != "" {
				c.AddSymbol(&plugin.Symbol{
					Name:       name,
					Category:   plugin.CategoryValue,
					Kind:       "variable",
					Signature:  name,
					Properties: plugin.NewProps().Map(),
					Span:       nodeSpan(child),
				})
				continue
			}
		}
	}
}

// extractPythonPatternNames extracts identifier names from a pattern_list,
// tuple_pattern, or list_pattern node and emits them as variable symbols.
func extractPythonPatternNames(node *sitter.Node, src []byte, c *plugin.Collector, spanNode *sitter.Node) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "identifier" {
			name := child.Utf8Text(src)
			if name != "" {
				c.AddSymbol(&plugin.Symbol{
					Name:       name,
					Category:   plugin.CategoryValue,
					Kind:       "variable",
					Signature:  name,
					Properties: plugin.NewProps().Map(),
					Span:       nodeSpan(spanNode),
				})
			}
		}
		// Handle nested patterns: (a, (b, c)) = ...
		if child.Kind() == "pattern_list" || child.Kind() == "tuple_pattern" || child.Kind() == "list_pattern" {
			extractPythonPatternNames(child, src, c, spanNode)
		}
	}
}

// extractImportStmt extracts an import_statement node (e.g., "import os").
// Emits EdgeImports with From=localName, To=modulePath.
func extractImportStmt(node *sitter.Node, src []byte, c *plugin.Collector) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "dotted_name":
			// import foo.bar → local name is "foo" (top-level module)
			modulePath := child.Utf8Text(src)
			// The local name in Python is the first segment.
			localName := modulePath
			if dot := indexByte(modulePath, '.'); dot >= 0 {
				localName = modulePath[:dot]
			}
			c.AddEdge(plugin.Edge{From: localName, To: modulePath, Kind: plugin.EdgeImports})
		case "aliased_import":
			// import foo as bar → local name is "bar"
			modulePath := childText(child, "dotted_name", src)
			alias := childText(child, "identifier", src)
			if modulePath != "" && alias != "" {
				c.AddEdge(plugin.Edge{From: alias, To: modulePath, Kind: plugin.EdgeImports})
			}
		}
	}
}

// extractImportFromStmt extracts an import_from_statement node
// (e.g., "from models import User, Order").
// Emits EdgeImports with From=localName, To=modulePath for each imported name.
func extractImportFromStmt(node *sitter.Node, src []byte, c *plugin.Collector) {
	// Find the module path (dotted_name or relative_import).
	var modulePath string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "dotted_name" || child.Kind() == "relative_import" {
			modulePath = child.Utf8Text(src)
			break
		}
	}
	if modulePath == "" {
		return
	}

	// Find imported names.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "dotted_name":
			// Skip the module path itself (already captured above).
			if child.Utf8Text(src) == modulePath {
				continue
			}
			name := child.Utf8Text(src)
			c.AddEdge(plugin.Edge{From: name, To: modulePath, Kind: plugin.EdgeImports})
		case "aliased_import":
			alias := childText(child, "identifier", src)
			if alias != "" {
				c.AddEdge(plugin.Edge{From: alias, To: modulePath, Kind: plugin.EdgeImports})
			}
		case "wildcard_import":
			c.AddEdge(plugin.Edge{From: "*", To: modulePath, Kind: plugin.EdgeImports})
		}
	}
}

// indexByte returns the index of the first occurrence of b in s, or -1.
func indexByte(s string, b byte) int {
	for i := range len(s) {
		if s[i] == b {
			return i
		}
	}
	return -1
}
