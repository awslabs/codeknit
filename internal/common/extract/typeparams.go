// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TypeParamConfig describes how to extract type parameters from generic
// declarations. Different languages have different node structures for
// type parameters.
type TypeParamConfig struct {
	// TypeParamsKind is the node kind for the type parameter list
	// (e.g., "type_parameters" in TypeScript, "type_params" in Rust).
	TypeParamsKind string
	// TypeParamKind is the node kind for individual type parameters
	// (e.g., "type_parameter" in TypeScript, "type_identifier" in Go).
	TypeParamKind string
	// NameKind is the node kind for the type parameter name
	// (usually "type_identifier" or "identifier").
	NameKind string
	// ConstraintKind is an optional node kind for type constraints
	// (e.g., "constraint" in TypeScript, "where_clause" in C#).
	ConstraintKind string
}

// TypeParams extracts type parameters from a declaration node and
// emits them as CategoryType symbols with kind "type_parameter".
// It also emits EdgeContains edges from the parent symbol to each type parameter.
func TypeParams(node *sitter.Node, src []byte, c *Collector, parentName string, cfg TypeParamConfig) {
	tpList := ChildByKind(node, cfg.TypeParamsKind)
	if tpList == nil {
		return
	}

	for i := range tpList.ChildCount() {
		child := tpList.Child(i)
		if child == nil {
			continue
		}

		// Match the type parameter wrapper node kind.
		if child.Kind() != cfg.TypeParamKind {
			// Also accept direct identifier/type_identifier children of the
			// list node — some grammars (e.g., Scala) don't wrap type params
			// in a dedicated node kind.
			if child.Kind() != cfg.NameKind {
				continue
			}
			// The child IS the name node itself.
			name := child.Utf8Text(src)
			if name == "" {
				continue
			}
			emitTypeParam(child, src, c, parentName, name, cfg)
			continue
		}

		// Look for the name inside the type parameter wrapper node.
		name := ChildText(child, cfg.NameKind, src)
		if name == "" {
			// Fallback: if the wrapper node has no child with NameKind,
			// try using the first named child's text (handles grammars
			// where the identifier is the only named child).
			for j := range child.ChildCount() {
				gc := child.Child(j)
				if gc != nil && gc.IsNamed() {
					name = gc.Utf8Text(src)
					break
				}
			}
		}
		if name == "" {
			continue
		}

		emitTypeParam(child, src, c, parentName, name, cfg)
	}
}

// emitTypeParam creates a type_parameter symbol and optional contains edge.
func emitTypeParam(node *sitter.Node, src []byte, c *Collector, parentName, name string, cfg TypeParamConfig) {
	scopedName := types.MakeScopedName(parentName, name)

	props := types.NewProps()
	if cfg.ConstraintKind != "" {
		if constraint := ChildByKind(node, cfg.ConstraintKind); constraint != nil {
			props.Set("constraint", constraint.Utf8Text(src))
		}
	}

	c.AddSymbol(&types.Symbol{
		Name:       name,
		ScopedName: scopedName,
		Category:   types.CategoryType,
		Kind:       "type_parameter",
		Signature:  name,
		Properties: props.Map(),
		Span:       NodeSpan(node),
	})

	if parentName != "" {
		c.AddEdge(types.Edge{
			From: parentName,
			To:   scopedName,
			Kind: types.EdgeContains,
		})
	}
}

// TypeParamsFromConfig is a convenience function that extracts type
// parameters using a TypeParamConfig and adds them to the collector.
func TypeParamsFromConfig(node *sitter.Node, src []byte, c *Collector, parentName string, cfg *TypeParamConfig) {
	if cfg == nil {
		return
	}
	TypeParams(node, src, c, parentName, *cfg)
}

// CollectTypeParamNames extracts just the names of type parameters from a node.
// This is useful for building a set of known type parameters to filter out
// from type reference edges.
func CollectTypeParamNames(node *sitter.Node, src []byte, cfg TypeParamConfig) []string {
	tpList := ChildByKind(node, cfg.TypeParamsKind)
	if tpList == nil {
		return nil
	}

	var names []string
	for i := range tpList.ChildCount() {
		child := tpList.Child(i)
		if child == nil || child.Kind() != cfg.TypeParamKind {
			continue
		}
		name := ChildText(child, cfg.NameKind, src)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}
