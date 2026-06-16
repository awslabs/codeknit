// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"bytes"
	"sort"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// fpSpanKey matches symbols to AST nodes by line span.
type fpSpanKey struct{ start, end uint }

// fpSpanEntry tracks a symbol index and its category for span matching.
type fpSpanEntry struct {
	category types.SymbolCategory
	idx      int
}

// ExtractBodyTokens walks the AST to find symbol bodies and initializers,
// extracting normalized token streams for fingerprinting.
func (c *Collector) ExtractBodyTokens(root *sitter.Node, src []byte, tokenMap map[string]byte) {
	if !c.Fingerprint || tokenMap == nil {
		return
	}

	spans := make(map[fpSpanKey]fpSpanEntry, len(c.Symbols))
	for i := range c.Symbols {
		cat := c.Symbols[i].Category
		if cat == types.CategoryCallable || cat == types.CategoryValue {
			s := c.Symbols[i].Span
			//nolint:gosec // Span values are always non-negative line numbers.
			spans[fpSpanKey{uint(s[0]), uint(s[1])}] = fpSpanEntry{idx: i, category: cat}
		}
	}

	// Pass 1: Walk callable bodies and value initializers.
	findSymbolNodes(root, src, tokenMap, spans, c)

	// Pass 2: Walk top-level code not claimed by any symbol.
	c.extractTopLevelTokens(root, src, tokenMap)
}

// findSymbolNodes recursively searches for AST nodes whose span matches
// a tracked symbol, then extracts tokens from the appropriate child.
func findSymbolNodes(node *sitter.Node, src []byte, tokenMap map[string]byte, spans map[fpSpanKey]fpSpanEntry, c *Collector) {
	startLine := node.StartPosition().Row + 1
	endLine := node.EndPosition().Row + 1
	key := fpSpanKey{startLine, endLine}

	if entry, ok := spans[key]; ok {
		var tokens []byte
		switch entry.category {
		case types.CategoryCallable:
			if body := findBodyChild(node); body != nil {
				tokens = walkBodyForTokens(body, src, tokenMap)
			}
		case types.CategoryValue:
			tokens = extractInitializerTokens(node, src, tokenMap)
		}
		if len(tokens) > 0 {
			c.Symbols[entry.idx].BodyTokens = tokens
		}
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			findSymbolNodes(child, src, tokenMap, spans, c)
		}
	}
}

// extractInitializerTokens finds the initializer expression of a variable/
// constant declaration and walks it to produce tokens.
func extractInitializerTokens(node *sitter.Node, src []byte, tokenMap map[string]byte) []byte {
	foundEq := false
	var tokens []byte
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			text := string(src[child.StartByte():child.EndByte()])
			if text == "=" || text == ":=" {
				foundEq = true
				continue
			}
		}
		if foundEq {
			tokens = walkInitExpr(child, src, tokenMap)
			break
		}
	}

	if !foundEq {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil || !child.IsNamed() {
				continue
			}
			if t := extractInitializerTokens(child, src, tokenMap); len(t) > 0 {
				return t
			}
		}
	}

	return tokens
}

func walkInitExpr(node *sitter.Node, src []byte, tokenMap map[string]byte) []byte {
	tokens := make([]byte, 0, 16)
	walkInitRecursive(node, src, tokenMap, &tokens)
	return tokens
}

// literalTokens maps tree-sitter literal node kinds to fingerprint tokens.
var literalTokens = map[string]byte{
	// Numeric
	"int_literal":   types.FPLitNum,
	"float_literal": types.FPLitNum,
	"integer":       types.FPLitNum,
	"float":         types.FPLitNum,
	"number":        types.FPLitNum,
	"rune_literal":  types.FPLitNum,
	"char_literal":  types.FPLitNum,
	// String
	"string_literal":             types.FPLitStr,
	"interpreted_string_literal": types.FPLitStr,
	"raw_string_literal":         types.FPLitStr,
	"string":                     types.FPLitStr,
	"template_string":            types.FPLitStr,
	"regex":                      types.FPLitStr,
	// Boolean
	"true":    types.FPLitBool,
	"false":   types.FPLitBool,
	"boolean": types.FPLitBool,
	// Nil
	"nil":       types.FPLitNil,
	"null":      types.FPLitNil,
	"none":      types.FPLitNil,
	"undefined": types.FPLitNil,
}

// containerTokens maps collection literal node kinds to fingerprint tokens.
var containerTokens = map[string]byte{
	"dictionary":                types.FPDict,
	"object":                    types.FPDict,
	"literal_value":             types.FPDict,
	"hash":                      types.FPDict,
	"map_literal":               types.FPDict,
	"list":                      types.FPArray,
	"array":                     types.FPArray,
	"tuple":                     types.FPArray,
	"slice_literal":             types.FPArray,
	"array_creation_expression": types.FPArray,
}

func walkInitRecursive(node *sitter.Node, src []byte, tokenMap map[string]byte, tokens *[]byte) {
	kind := node.Kind()

	if kind == "identifier" || kind == "type_identifier" || kind == "field_identifier" ||
		kind == "property_identifier" || kind == "package_identifier" || kind == "name" ||
		kind == "type_annotation" || kind == "type_arguments" || kind == "generic_type" ||
		kind == "predefined_type" || kind == "primitive_type" ||
		kind == "comment" || kind == "line_comment" || kind == "block_comment" ||
		kind == "string_content" || kind == "escape_sequence" || kind == "string_start" ||
		kind == "string_end" || kind == "string_fragment" ||
		kind == "interpreted_string_literal_content" {
		return
	}

	if litTok, ok := literalTokens[kind]; ok {
		*tokens = append(*tokens, litTok)
		val := src[node.StartByte():node.EndByte()]
		h := fnv32(val)
		*tokens = append(*tokens, h[0], h[1], h[2], h[3])
		return
	}

	if contTok, ok := containerTokens[kind]; ok {
		*tokens = append(*tokens, contTok)
		var elemHashes [][4]byte
		for i := range node.ChildCount() {
			ch := node.Child(i)
			if ch == nil || !ch.IsNamed() {
				continue
			}
			h := fnv32(src[ch.StartByte():ch.EndByte()])
			elemHashes = append(elemHashes, h)
		}
		count := len(elemHashes)
		if count > 255 {
			count = 255
		}
		*tokens = append(*tokens, byte(count))
		sort.Slice(elemHashes, func(a, b int) bool {
			return bytes.Compare(elemHashes[a][:], elemHashes[b][:]) < 0
		})
		for _, h := range elemHashes {
			*tokens = append(*tokens, h[0], h[1], h[2], h[3])
		}
		return
	}

	tok, mapped := tokenMap[kind]
	if mapped {
		*tokens = append(*tokens, tok)
	}

	if operatorKinds[kind] {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child != nil && child.ChildCount() == 0 && !child.IsNamed() {
				if opTok, ok := operatorTokenMap[string(src[child.StartByte():child.EndByte()])]; ok {
					*tokens = append(*tokens, opTok)
					break
				}
			}
		}
	}

	if mapped && tok == types.FPCall {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child != nil && callArgKinds[child.Kind()] {
				count := 0
				for j := range child.ChildCount() {
					arg := child.Child(j)
					if arg != nil && arg.IsNamed() {
						count++
					}
				}
				if count < 255 {
					*tokens = append(*tokens, byte(count))
				}
				break
			}
		}
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			walkInitRecursive(child, src, tokenMap, tokens)
		}
	}
}

// extractTopLevelTokens walks the root AST node's direct children and
// collects tokens from any that are NOT claimed by an existing symbol.
func (c *Collector) extractTopLevelTokens(root *sitter.Node, src []byte, tokenMap map[string]byte) {
	type lineRange struct{ start, end uint }
	claimed := make([]lineRange, 0, len(c.Symbols))
	for i := range c.Symbols {
		s := c.Symbols[i].Span
		if s[0] > 0 && s[1] > 0 {
			//nolint:gosec // Span values guarded as > 0 above.
			claimed = append(claimed, lineRange{uint(s[0]), uint(s[1])})
		}
	}

	skipKinds := map[string]bool{
		// Imports
		"package_clause":            true,
		"import_declaration":        true,
		"import_statement":          true,
		"import_from_statement":     true,
		"using_directive":           true,
		"use_declaration":           true,
		"preproc_include":           true,
		"namespace_use_declaration": true,
		"export_statement":          true,
		"package_declaration":       true,
		// Functions / methods
		"function_declaration": true,
		"method_declaration":   true,
		"function_definition":  true,
		"function_item":        true,
		"method":               true,
		"singleton_method":     true,
		"decorated_definition": true,
		// Types
		"type_declaration":            true,
		"class_declaration":           true,
		"class_definition":            true,
		"interface_declaration":       true,
		"struct_declaration":          true,
		"enum_declaration":            true,
		"trait_declaration":           true,
		"trait_definition":            true,
		"struct_specifier":            true,
		"enum_specifier":              true,
		"class_specifier":             true,
		"struct_item":                 true,
		"enum_item":                   true,
		"trait_item":                  true,
		"type_item":                   true,
		"type_definition":             true,
		"enum_definition":             true,
		"object_definition":           true,
		"annotation_type_declaration": true,
		// Variables / constants
		"var_declaration":      true,
		"const_declaration":    true,
		"lexical_declaration":  true,
		"variable_declaration": true,
		"val_definition":       true,
		"var_definition":       true,
		"const_item":           true,
		"static_item":          true,
		"field_declaration":    true,
		"property_declaration": true,
		// Containers
		"namespace_declaration": true,
		"namespace_definition":  true,
		"impl_item":             true,
		"mod_item":              true,
		"module":                true,
		"class":                 true,
		// Preprocessor / macros
		"preproc_def":          true,
		"preproc_function_def": true,
		"template_declaration": true,
		"union_specifier":      true,
		"macro_definition":     true,
	}

	var tokens []byte
	for i := range root.ChildCount() {
		child := root.Child(i)
		if child == nil {
			continue
		}

		if skipKinds[child.Kind()] {
			continue
		}

		childStart := child.StartPosition().Row + 1
		childEnd := child.EndPosition().Row + 1
		overlaps := false
		for _, cr := range claimed {
			if childStart <= cr.end && childEnd >= cr.start {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}

		chunk := walkBodyForTokens(child, src, tokenMap)
		tokens = append(tokens, chunk...)
	}

	if len(tokens) == 0 {
		return
	}

	for i := range c.Symbols {
		if c.Symbols[i].Category == types.CategoryModule {
			c.Symbols[i].BodyTokens = tokens
			return
		}
	}

	c.AddSymbol(&types.Symbol{
		Name:     "<top-level:" + c.FilePath + ">",
		Category: types.CategoryModule,
		Kind:     "script",
		//nolint:gosec // EndPosition().Row is a tree-sitter line count that fits in int.
		Span:       [2]int{1, int(root.EndPosition().Row + 1)},
		BodyTokens: tokens,
	})
}

// bodyChildKinds are tree-sitter node kinds that represent function bodies.
var bodyChildKinds = map[string]bool{
	"block":              true,
	"statement_block":    true,
	"compound_statement": true,
	"method_body":        true,
	"class_body":         true,
	"function_body":      true,
	"body":               true,
}

func findBodyChild(node *sitter.Node) *sitter.Node {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if bodyChildKinds[child.Kind()] {
			return child
		}
	}
	return nil
}

func walkBodyForTokens(node *sitter.Node, src []byte, tokenMap map[string]byte) []byte {
	tokens := make([]byte, 0, 64)
	walkBodyRecursive(node, src, tokenMap, &tokens)
	return tokens
}

// identKindsTokens are tree-sitter node kinds to skip entirely.
var identKindsTokens = map[string]bool{
	"identifier":                    true,
	"type_identifier":               true,
	"field_identifier":              true,
	"property_identifier":           true,
	"shorthand_property_identifier": true,
	"package_identifier":            true,
	"label_name":                    true,
	"name":                          true,
	"type_annotation":               true,
	"type_arguments":                true,
	"generic_type":                  true,
	"predefined_type":               true,
	"primitive_type":                true,
	"sized_type_specifier":          true,
	"comment":                       true,
	"line_comment":                  true,
	"block_comment":                 true,
	"string_content":                true,
	"escape_sequence":               true,
}

var operatorKinds = map[string]bool{
	"binary_expression":    true,
	"binary_operator":      true,
	"comparison_operator":  true,
	"boolean_operator":     true,
	"unary_expression":     true,
	"not_operator":         true,
	"augmented_assignment": true,
	"update_expression":    true,
}

var operatorTokenMap = map[string]byte{
	// Arithmetic
	"+":  types.FPAdd,
	"+=": types.FPAdd,
	"-":  types.FPSub,
	"-=": types.FPSub,
	"*":  types.FPMul,
	"*=": types.FPMul,
	"/":  types.FPDiv,
	"/=": types.FPDiv,
	"%":  types.FPMod,
	"%=": types.FPMod,
	// Comparison
	"==":  types.FPEq,
	"===": types.FPEq,
	"!=":  types.FPNeq,
	"!==": types.FPNeq,
	"<":   types.FPLt,
	">":   types.FPGt,
	"<=":  types.FPLte,
	">=":  types.FPGte,
	// Logical
	"&&":  types.FPAnd,
	"and": types.FPAnd,
	"||":  types.FPOr,
	"or":  types.FPOr,
	"!":   types.FPNot,
	"not": types.FPNot,
	// Bitwise
	"&":   types.FPBitAnd,
	"&=":  types.FPBitAnd,
	"|":   types.FPBitOr,
	"|=":  types.FPBitOr,
	"^":   types.FPBitXor,
	"^=":  types.FPBitXor,
	"~":   types.FPBitNot,
	"<<":  types.FPShl,
	"<<=": types.FPShl,
	">>":  types.FPShr,
	">>=": types.FPShr,
}

var callArgKinds = map[string]bool{
	"argument_list": true,
	"arguments":     true,
}

func fnv32(data []byte) [4]byte {
	var h uint32 = 2166136261
	for _, b := range data {
		h ^= uint32(b)
		h *= 16777619
	}
	return [4]byte{byte(h >> 24), byte(h >> 16), byte(h >> 8), byte(h)} //nolint:gosec // FNV-1a hash bytes, values are always 0-255
}

func normalizeCalleName(name []byte) []byte {
	out := make([]byte, 0, len(name))
	for _, b := range name {
		if b == '_' {
			continue
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		out = append(out, b)
	}
	return out
}

func walkBodyRecursive(node *sitter.Node, src []byte, tokenMap map[string]byte, tokens *[]byte) {
	walkBodyWithScope(node, src, tokenMap, tokens, make(map[string]byte))
}

func walkBodyWithScope(node *sitter.Node, src []byte, tokenMap map[string]byte, tokens *[]byte, varScope map[string]byte) {
	kind := node.Kind()

	if litTok, ok := literalTokens[kind]; ok {
		*tokens = append(*tokens, litTok)
		val := src[node.StartByte():node.EndByte()]
		h := fnv32(val)
		*tokens = append(*tokens, h[0], h[1], h[2], h[3])
		return
	}

	if kind == "identifier" || kind == "name" {
		name := string(src[node.StartByte():node.EndByte()])
		ord, seen := varScope[name]
		if !seen {
			next := byte(len(varScope) + 1) //nolint:gosec // capped to 255 below
			if next == 0 {
				next = 255
			}
			varScope[name] = next
			ord = next
		}
		*tokens = append(*tokens, ord)
		return
	}

	if identKindsTokens[kind] {
		return
	}

	if kind == "else_clause" || kind == "else" {
		if isElseIf(node) {
			*tokens = append(*tokens, types.FPElseIf)
			for i := range node.ChildCount() {
				child := node.Child(i)
				if child != nil && (child.Kind() == "if_statement" || child.Kind() == "if_expression") {
					for j := range child.ChildCount() {
						grandchild := child.Child(j)
						if grandchild != nil {
							walkBodyWithScope(grandchild, src, tokenMap, tokens, varScope)
						}
					}
					return
				}
			}
			return
		}
	}

	tok, mapped := tokenMap[kind]
	if mapped {
		*tokens = append(*tokens, tok)
	}

	if operatorKinds[kind] {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if child.ChildCount() == 0 {
				if opTok, ok := operatorTokenMap[string(src[child.StartByte():child.EndByte()])]; ok {
					*tokens = append(*tokens, opTok)
					break
				}
			}
		}
	}

	if mapped && tok == types.FPCall {
		if node.ChildCount() > 0 {
			callee := node.Child(0)
			if callee != nil {
				calleeName := normalizeCalleName(src[callee.StartByte():callee.EndByte()])
				h := fnv32(calleeName)
				*tokens = append(*tokens, h[0], h[1])
			}
		}
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if callArgKinds[child.Kind()] {
				count := 0
				for j := range child.ChildCount() {
					arg := child.Child(j)
					if arg != nil && arg.IsNamed() {
						count++
					}
				}
				if count < 255 {
					*tokens = append(*tokens, byte(count))
				} else {
					*tokens = append(*tokens, 0xFF)
				}
				for j := range child.ChildCount() {
					arg := child.Child(j)
					if arg != nil {
						walkBodyWithScope(arg, src, tokenMap, tokens, varScope)
					}
				}
				break
			}
		}
		return
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child != nil {
			walkBodyWithScope(child, src, tokenMap, tokens, varScope)
		}
	}
}

func isElseIf(node *sitter.Node) bool {
	namedCount := 0
	hasIf := false
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		namedCount++
		if child.Kind() == "if_statement" || child.Kind() == "if_expression" {
			hasIf = true
		}
	}
	return hasIf && namedCount == 1
}
