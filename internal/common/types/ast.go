// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types provides the core Symbol, Edge, and property types used
// throughout the codeknit plugin system. It is a leaf package with no
// internal imports, allowing both plugin and extract to depend on it
// without circular dependencies.
package types

// SymbolCategory is a language-agnostic classification for a Symbol.
type SymbolCategory string

// SymbolCategory constants.
const (
	CategoryCallable SymbolCategory = "callable"
	CategoryType     SymbolCategory = "type"
	CategoryValue    SymbolCategory = "value"
	CategoryModule   SymbolCategory = "module"
	CategoryMeta     SymbolCategory = "meta"
)

// Symbol represents a named element extracted from source code.
type Symbol struct {
	Properties map[string]string
	ID         string
	Name       string
	ScopedName string
	FilePath   string
	Category   SymbolCategory
	Kind       string
	Signature  string
	BodyTokens []byte
	Span       [2]int
}

// EdgeKind represents the type of relationship between two Symbols.
type EdgeKind string

// EdgeKind constants for supported relationship types.
const (
	EdgeCalls      EdgeKind = "calls"
	EdgeInherits   EdgeKind = "inherits"
	EdgeContains   EdgeKind = "contains"
	EdgeReferences EdgeKind = "references"
	EdgeImplements EdgeKind = "implements"
	EdgeOverrides  EdgeKind = "overrides"
	EdgeImports    EdgeKind = "imports"
	EdgeDecorates  EdgeKind = "decorates"
	EdgeAliases    EdgeKind = "aliases"
	EdgeReturns    EdgeKind = "returns"
)

// Edge represents a typed, directed relationship between two Symbols.
type Edge struct {
	From string
	To   string
	Kind EdgeKind
}

// EffectiveScopedName returns ScopedName if set, otherwise Name.
func (s *Symbol) EffectiveScopedName() string {
	if s.ScopedName != "" {
		return s.ScopedName
	}
	return s.Name
}

// MakeScopedName builds a scope-qualified name from a parent and child name.
func MakeScopedName(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// Props is a builder for symbol property maps.
type Props map[string]string

// NewProps creates an empty Props map.
func NewProps() Props { return Props{} }

// Set adds a key-value pair. Empty values are silently ignored.
func (p Props) Set(key, value string) Props {
	if value != "" {
		p[key] = value
	}
	return p
}

// SetBool adds a boolean property. Only "true" values are stored.
func (p Props) SetBool(key string, value bool) Props {
	if value {
		p[key] = "true"
	}
	return p
}

// Map returns the underlying map for use in Symbol.Properties.
func (p Props) Map() map[string]string { return p }
