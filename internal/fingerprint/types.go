// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"bytes"
	"sort"
	"strings"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// Structural shape tokens for type fingerprinting.
// These encode the shape of a type (field count, method signatures, etc.)
// rather than method body logic.
const (
	fpSep        byte = 0xFE // separator between sections
	fpField      byte = 0xE0 // a field/property exists
	fpMethod     byte = 0xE1 // a method exists
	fpInherits   byte = 0xE2 // inherits from a parent
	fpImplements byte = 0xE3 // implements an interface
	fpNested     byte = 0xE4 // contains a nested type
	fpParamMark  byte = 0xE5 // method parameter count follows
	fpRetMark    byte = 0xE6 // method has a return type
	fpAbstract   byte = 0xE7 // method/class is abstract
	fpStatic     byte = 0xE8 // method/field is static
	fpAsync      byte = 0xE9 // method is async
	fpKindMark   byte = 0xEA // type kind discriminator follows (2 bytes)
	fpNameMark   byte = 0xEB // method name hash follows (2 bytes)
)

// Types computes structural shape fingerprints for CategoryType
// symbols (classes, structs, interfaces, enums). This must be called after
// the planner has built the SymbolGraph with edges.
//
// The structural shape encodes:
//   - Number and kind of children (fields vs methods vs nested types)
//   - Method signatures (parameter count, return type presence, modifiers)
//   - Parent structural fingerprints (so A extends Foo matches B extends Bar
//     only when Foo and Bar have the same shape)
//   - All sorted for order-independence
//
// Method body tokens (from the per-file pass) are NOT included — this
// fingerprint captures the type's API surface, not its implementation.
func Types(sg *ir.SymbolGraph) {
	// Build indexes.
	idToIdx := make(map[string]int, len(sg.Symbols))
	for i := range sg.Symbols {
		idToIdx[sg.Symbols[i].ID] = i
	}

	children := make(map[string][]int)            // parent ID → child indexes
	inheritParents := make(map[string][]string)   // type ID → parent IDs (inherits)
	implementParents := make(map[string][]string) // type ID → parent IDs (implements)

	for _, edge := range sg.Edges {
		switch edge.Kind {
		case plugin.EdgeContains:
			if ci, ok := idToIdx[edge.To]; ok {
				children[edge.From] = append(children[edge.From], ci)
			}
		case plugin.EdgeInherits:
			inheritParents[edge.From] = append(inheritParents[edge.From], edge.To)
		case plugin.EdgeImplements:
			implementParents[edge.From] = append(implementParents[edge.From], edge.To)
		}
	}

	// Two-pass approach: process types with no parents first so their
	// BodyTokens are available when child types reference them.
	// Pass 1: types with no inheritance.
	// Pass 2: types with inheritance (parents may now have tokens).
	for pass := range 2 {
		for i := range sg.Symbols {
			sym := &sg.Symbols[i]
			if sym.Category != plugin.CategoryType {
				continue
			}
			if len(sym.BodyTokens) > 0 {
				continue
			}
			hasParents := len(inheritParents[sym.ID]) > 0 || len(implementParents[sym.ID]) > 0
			if pass == 0 && hasParents {
				continue // defer to pass 1
			}

			tokens := buildTypeShape(sym.Kind, children[sym.ID], inheritParents[sym.ID], implementParents[sym.ID], sg, idToIdx)
			if len(tokens) > 0 {
				sym.BodyTokens = tokens
			}
		}
	}
}

// buildTypeShape encodes the structural shape of a type as a byte sequence.
// Returns nil if the type has no meaningful structure to encode.
func buildTypeShape(
	kind string,
	childIdxs []int,
	inheritIDs []string,
	implementIDs []string,
	sg *ir.SymbolGraph,
	idToIdx map[string]int,
) []byte {
	// Collect per-child shape descriptors.
	var fieldShapes [][]byte
	var methodShapes [][]byte
	var nestedShapes [][]byte

	for _, ci := range childIdxs {
		child := &sg.Symbols[ci]
		switch child.Category {
		case plugin.CategoryValue:
			fieldShapes = append(fieldShapes, encodeFieldShape(child))
		case plugin.CategoryCallable:
			methodShapes = append(methodShapes, encodeMethodShape(child))
		case plugin.CategoryType:
			nestedShapes = append(nestedShapes, []byte{fpNested})
		}
	}

	// Build parent shape descriptors.
	// Each parent contributes [fpInherits/fpImplements] + first 4 bytes of
	// its own structural fingerprint (or its name hash if not yet computed).
	// This means A extends Foo matches B extends Bar only when Foo and Bar
	// have the same structural shape.
	inheritShapes := make([][]byte, 0, len(inheritIDs))
	for _, pid := range inheritIDs {
		inheritShapes = append(inheritShapes, encodeParentShape(fpInherits, pid, sg, idToIdx))
	}
	implementShapes := make([][]byte, 0, len(implementIDs))
	for _, pid := range implementIDs {
		implementShapes = append(implementShapes, encodeParentShape(fpImplements, pid, sg, idToIdx))
	}

	// Nothing to encode.
	if len(fieldShapes) == 0 && len(methodShapes) == 0 &&
		len(nestedShapes) == 0 && len(inheritShapes) == 0 && len(implementShapes) == 0 {
		return nil
	}

	// Sort each group for order-independence.
	sort.Slice(fieldShapes, func(a, b int) bool { return bytes.Compare(fieldShapes[a], fieldShapes[b]) < 0 })
	sort.Slice(methodShapes, func(a, b int) bool { return bytes.Compare(methodShapes[a], methodShapes[b]) < 0 })
	sort.Slice(nestedShapes, func(a, b int) bool { return bytes.Compare(nestedShapes[a], nestedShapes[b]) < 0 })
	sort.Slice(inheritShapes, func(a, b int) bool { return bytes.Compare(inheritShapes[a], inheritShapes[b]) < 0 })
	sort.Slice(implementShapes, func(a, b int) bool { return bytes.Compare(implementShapes[a], implementShapes[b]) < 0 })

	// Layout: [KIND_MARK k0 k1] [inherit shapes] [implement shapes] SEP [fields] SEP [methods] SEP? [nested]
	var tokens []byte

	k := kindHash(kind)
	tokens = append(tokens, fpKindMark, k[0], k[1])

	for _, s := range inheritShapes {
		tokens = append(tokens, s...)
	}
	for _, s := range implementShapes {
		tokens = append(tokens, s...)
	}

	tokens = append(tokens, fpSep)

	for _, fs := range fieldShapes {
		tokens = append(tokens, fs...)
	}

	tokens = append(tokens, fpSep)

	for _, ms := range methodShapes {
		tokens = append(tokens, ms...)
	}

	if len(nestedShapes) > 0 {
		tokens = append(tokens, fpSep)
		for _, ns := range nestedShapes {
			tokens = append(tokens, ns...)
		}
	}

	return tokens
}

// encodeParentShape encodes a parent type reference as a short byte sequence.
// Format: [marker] [h0..h7] where the 8 bytes are a hash of the parent's full
// BodyTokens (if already computed) or a hash of the parent ID string (fallback).
// Hashing the full token stream ensures parents with the same kind but different
// fields/methods are distinguished.
func encodeParentShape(marker byte, parentID string, sg *ir.SymbolGraph, idToIdx map[string]int) []byte {
	result := []byte{marker}
	if idx, ok := idToIdx[parentID]; ok {
		pt := sg.Symbols[idx].BodyTokens
		if len(pt) > 0 {
			h := hashBytes8(pt)
			result = append(result, h[:]...)
			return result
		}
	}
	// Fallback: hash the parent ID string.
	h := hashBytes8([]byte(parentID))
	result = append(result, h[:]...)
	return result
}

// hashBytes8 computes a stable 8-byte hash of the input using two FNV-1a
// passes with different seeds for better distribution.
func hashBytes8(data []byte) [8]byte {
	var h1 uint32 = 2166136261
	var h2 uint32 = 2166136261 ^ 0xdeadbeef
	for _, b := range data {
		h1 ^= uint32(b)
		h1 *= 16777619
		h2 ^= uint32(b)
		h2 *= 16777619 + 1
	}
	return [8]byte{
		byte(h1 >> 24), byte(h1 >> 16), byte(h1 >> 8), byte(h1), //nolint:gosec // FNV-1a hash bytes, values are always 0-255
		byte(h2 >> 24), byte(h2 >> 16), byte(h2 >> 8), byte(h2), //nolint:gosec // FNV-1a hash bytes, values are always 0-255
	}
}

// kindHash returns a stable 2-byte hash of a type kind string.
// This is intentionally simple — we only need enough entropy to distinguish
// "struct" from "class" from "interface" from "type_alias" etc.
func kindHash(kind string) [2]byte {
	var h uint32 = 2166136261 // FNV-1a offset basis
	for i := range len(kind) {
		h ^= uint32(kind[i])
		h *= 16777619
	}
	return [2]byte{byte(h >> 8), byte(h)} //nolint:gosec // FNV-1a hash bytes, values are always 0-255
}

// encodeFieldShape encodes a field/property as a short byte sequence.
// Format: [FIELD] [NAME_MARK n0 n1] [modifiers...]
// The name hash ensures InputPath ≠ Output even when both are plain fields.
func encodeFieldShape(sym *plugin.Symbol) []byte {
	tokens := []byte{fpField}
	n := kindHash(sym.Name)
	tokens = append(tokens, fpNameMark, n[0], n[1])
	if sym.Properties["static"] == "true" {
		tokens = append(tokens, fpStatic)
	}
	return tokens
}

// encodeMethodShape encodes a method's signature shape as a byte sequence.
// Format: [METHOD] [NAME_MARK n0 n1] [modifiers...] [PARAM_MARK paramCount] [RET_MARK]?
// The name hash distinguishes methods with the same shape but different names
// (e.g. IsValid vs Validate both having 0 params and a return type).
func encodeMethodShape(sym *plugin.Symbol) []byte {
	tokens := []byte{fpMethod}

	// Stable 2-byte hash of the method name — prevents IsValid() matching Validate().
	n := kindHash(sym.Name)
	tokens = append(tokens, fpNameMark, n[0], n[1])

	// Modifiers.
	if sym.Properties["abstract"] == "true" {
		tokens = append(tokens, fpAbstract)
	}
	if sym.Properties["static"] == "true" {
		tokens = append(tokens, fpStatic)
	}
	if sym.Properties["async"] == "true" {
		tokens = append(tokens, fpAsync)
	}

	// Parameter count from signature (clamped to uint8 range).
	paramCount := countParams(sym.Signature)
	if paramCount > 255 {
		paramCount = 255
	}
	tokens = append(tokens, fpParamMark, byte(paramCount)) //nolint:gosec // clamped above

	// Return type presence.
	if hasReturnType(sym.Signature) {
		tokens = append(tokens, fpRetMark)
	}

	return tokens
}

// countParams counts parameters in a signature string like "foo(a: int, b: string) -> bool".
// It counts commas between the first '(' and its matching ')' and adds 1 (unless empty).
func countParams(sig string) int {
	start := strings.IndexByte(sig, '(')
	if start < 0 {
		return 0
	}
	end := strings.LastIndexByte(sig, ')')
	if end <= start+1 {
		return 0 // empty parens "()"
	}
	inner := sig[start+1 : end]
	if strings.TrimSpace(inner) == "" {
		return 0
	}
	// Count commas at depth 0 (skip nested parens/generics).
	count := 1
	depth := 0
	for _, ch := range inner {
		switch ch {
		case '(', '<', '[':
			depth++
		case ')', '>', ']':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}

// hasReturnType checks if a signature has a return type (contains "->").
func hasReturnType(sig string) bool {
	return strings.Contains(sig, "->")
}
