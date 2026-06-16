// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ruby

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

// className extracts the name from a class or module node.
// Ruby class/module names are constants (e.g., "MyClass", "MyModule").
func className(node *sitter.Node, src []byte) string {
	// Try constant first (most common for class/module names).
	if name := childText(node, "constant", src); name != "" {
		return name
	}
	// Try scope_resolution for namespaced names like Foo::Bar.
	if sr := childByKind(node, "scope_resolution"); sr != nil {
		return sr.Utf8Text(src)
	}
	return ""
}

// methodName extracts the method name from a singleton_method node.
// singleton_method has the form: def <object>.<name>
func methodName(node *sitter.Node, src []byte) string {
	// The method name is the last identifier child.
	var name string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "identifier" {
			name = child.Utf8Text(src)
		}
	}
	return name
}

// superclassName extracts the superclass name from a superclass node.
func superclassName(node *sitter.Node, src []byte) string {
	// The superclass node wraps a constant or scope_resolution.
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "constant":
			return child.Utf8Text(src)
		case "scope_resolution":
			return child.Utf8Text(src)
		}
	}
	return ""
}

// visibilityCallName extracts the method name from a call node
// to detect visibility modifiers (private, protected, public).
func visibilityCallName(node *sitter.Node, src []byte) string {
	if node.ChildCount() == 0 {
		return ""
	}
	first := node.Child(0)
	if first == nil {
		return ""
	}
	if first.Kind() == "identifier" {
		return first.Utf8Text(src)
	}
	return ""
}

// callTarget is the CallTargetFunc for Ruby.
// Matches "call" nodes whose first child is an identifier or call (method chain),
// filtering out visibility keywords (private, protected, public).
var callTarget = plugin.FilterCallTarget(
	plugin.UnqualifiedCallTarget("call", []string{"identifier", "call"}, nil),
	"private", "protected", "public",
)

// richCallTarget is the RichCallTargetFunc for Ruby.
// Ruby's UnqualifiedCallTarget has no qualified kinds (nil), so all calls are
// treated as unqualified. No rich variant needed — use the plain callTarget
// wrapped as a RichCallTargetFunc with Qualified always false.
var richCallTarget plugin.RichCallTargetFunc = func(node *sitter.Node, src []byte) plugin.CallTargetResult {
	return plugin.CallTargetResult{Name: callTarget(node, src)}
}

// extractParamStrings extracts method parameter names.
func extractParamStrings(node *sitter.Node, src []byte) []string {
	params := childByKind(node, "method_parameters")
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
			result = append(result, child.Utf8Text(src))
		case "optional_parameter":
			name := childText(child, "identifier", src)
			if name != "" {
				result = append(result, name)
			}
		case "splat_parameter":
			name := childText(child, "identifier", src)
			if name != "" {
				result = append(result, "*"+name)
			}
		case "keyword_parameter":
			name := childText(child, "identifier", src)
			if name != "" {
				result = append(result, name+":")
			}
		case "block_parameter":
			name := childText(child, "identifier", src)
			if name != "" {
				result = append(result, "&"+name)
			}
		}
	}
	return result
}

// extractFunction extracts a top-level method node as a function symbol.
func extractFunction(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := childText(node, "identifier", src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "function",
		Signature:  buildFuncSignature(name, params, ""),
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// extractSingletonMethod extracts a singleton_method node (e.g., def self.foo).
func extractSingletonMethod(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := methodName(node, src)
	if name == "" {
		return
	}

	params := extractParamStrings(node, src)
	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryCallable,
		Kind:       "method",
		Signature:  buildFuncSignature(name, params, ""),
		Properties: plugin.NewProps().SetBool("static", true).Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)
}

// extractClass extracts a class node.
func extractClass(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := className(node, src)
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

	// Extract inheritance edge from superclass.
	if sup := childByKind(node, "superclass"); sup != nil {
		superName := superclassName(sup, src)
		if superName != "" {
			c.AddEdge(plugin.Edge{
				From: name,
				To:   superName,
				Kind: plugin.EdgeInherits,
			})
		}
	}

	// Extract class body members with visibility tracking.
	if body := childByKind(node, "body_statement"); body != nil {
		extractClassMembers(body, src, c, name)
	}
}

// extractModule extracts a module node.
func extractModule(node *sitter.Node, src []byte, c *plugin.Collector) {
	name := className(node, src)
	if name == "" {
		return
	}

	sym := plugin.Symbol{
		Name:       name,
		Category:   plugin.CategoryModule,
		Kind:       "module",
		Signature:  name,
		Properties: plugin.NewProps().Map(),
		Span:       nodeSpan(node),
	}
	c.AddSymbol(&sym)

	// Extract module body members with visibility tracking.
	if body := childByKind(node, "body_statement"); body != nil {
		extractClassMembers(body, src, c, name)
	}
}

// extractClassMembers extracts methods from a class/module body with visibility tracking.
// Ruby uses method-level visibility calls (public, private, protected) that affect
// subsequent method definitions until the next visibility call.
func extractClassMembers(body *sitter.Node, src []byte, c *plugin.Collector, parentName string) {
	visibility := "public" // default visibility
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "call":
			// Check for visibility modifier calls: private, protected, public.
			name := visibilityCallName(child, src)
			if name == "private" || name == "protected" || name == "public" {
				visibility = name
			}
		case "identifier":
			// Bare identifier visibility calls (e.g., just `private` on its own line).
			text := child.Utf8Text(src)
			if text == "private" || text == "protected" || text == "public" {
				visibility = text
			}
		case "method":
			mName := childText(child, "identifier", src)
			if mName == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentName, mName)
			params := extractParamStrings(child, src)
			p := plugin.NewProps()
			if visibility != "public" {
				p.Set("visibility", visibility)
			}
			sym := plugin.Symbol{
				Name:       mName,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(mName, params, ""),
				Properties: p.Map(),
				Span:       nodeSpan(child),
			}
			c.AddSymbol(&sym)
			c.AddEdge(plugin.Edge{
				From: parentName,
				To:   scopedName,
				Kind: plugin.EdgeContains,
			})
		case "singleton_method":
			mName := methodName(child, src)
			if mName == "" {
				continue
			}
			scopedName := plugin.MakeScopedName(parentName, mName)
			params := extractParamStrings(child, src)
			sym := plugin.Symbol{
				Name:       mName,
				ScopedName: scopedName,
				Category:   plugin.CategoryCallable,
				Kind:       "method",
				Signature:  buildFuncSignature(mName, params, ""),
				Properties: plugin.NewProps().SetBool("static", true).Map(),
				Span:       nodeSpan(child),
			}
			c.AddSymbol(&sym)
			c.AddEdge(plugin.Edge{
				From: parentName,
				To:   scopedName,
				Kind: plugin.EdgeContains,
			})
		case "class":
			// Nested class — extract recursively but don't track visibility.
			extractClass(child, src, c)
		case "module":
			extractModule(child, src, c)
		}
	}
}

// extractAssignment extracts an assignment or operator_assignment as a variable/constant symbol.
func extractAssignment(node *sitter.Node, src []byte, c *plugin.Collector) {
	if node.ChildCount() == 0 {
		return
	}
	left := node.Child(0)
	if left == nil {
		return
	}

	switch left.Kind() {
	case "identifier":
		name := left.Utf8Text(src)
		if name != "" {
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				Category:   plugin.CategoryValue,
				Kind:       "variable",
				Signature:  name,
				Properties: plugin.NewProps().Map(),
				Span:       nodeSpan(node),
			})
		}
	case "constant":
		name := left.Utf8Text(src)
		if name != "" {
			c.AddSymbol(&plugin.Symbol{
				Name:       name,
				Category:   plugin.CategoryValue,
				Kind:       "constant",
				Signature:  name,
				Properties: plugin.NewProps().Map(),
				Span:       nodeSpan(node),
			})
		}
	case "left_assignment_list":
		// Multiple assignment: a, b = 1, 2
		extractRubyMultiAssignNames(left, src, c, node)
	}
}

// extractRubyMultiAssignNames extracts identifier names from a left_assignment_list
// node (Ruby multiple assignment) and emits them as variable symbols.
func extractRubyMultiAssignNames(node *sitter.Node, src []byte, c *plugin.Collector, spanNode *sitter.Node) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier":
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
		case "constant":
			name := child.Utf8Text(src)
			if name != "" {
				c.AddSymbol(&plugin.Symbol{
					Name:       name,
					Category:   plugin.CategoryValue,
					Kind:       "constant",
					Signature:  name,
					Properties: plugin.NewProps().Map(),
					Span:       nodeSpan(spanNode),
				})
			}
		case "rest_assignment":
			// *rest in multiple assignment
			idNode := childByKind(child, "identifier")
			if idNode != nil {
				name := idNode.Utf8Text(src)
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
		case "destructured_left_assignment":
			// Nested destructuring: (a, b), c = [1, 2], 3
			extractRubyMultiAssignNames(child, src, c, spanNode)
		}
	}
}

// extractRequire extracts require/require_relative calls and emits EdgeImports edges.
// Handles: require "json"
//
//	require_relative "models/user"
func extractRequire(node *sitter.Node, src []byte, c *plugin.Collector) {
	if node.ChildCount() == 0 {
		return
	}
	first := node.Child(0)
	if first == nil || first.Kind() != "identifier" {
		return
	}
	name := first.Utf8Text(src)
	if name != "require" && name != "require_relative" {
		return
	}

	// Find the string argument.
	argList := childByKind(node, "argument_list")
	if argList == nil {
		return
	}
	for i := range argList.ChildCount() {
		child := argList.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "string" {
			// Extract the string content (strip quotes).
			raw := child.Utf8Text(src)
			if len(raw) >= 2 {
				modulePath := raw[1 : len(raw)-1]
				// Local name is the last path segment.
				localName := modulePath
				if idx := plugin.LastSepIndex(modulePath, "/"); idx >= 0 {
					localName = modulePath[idx+1:]
				}
				c.AddEdge(plugin.Edge{From: localName, To: modulePath, Kind: plugin.EdgeImports})
			}
		}
	}
}
