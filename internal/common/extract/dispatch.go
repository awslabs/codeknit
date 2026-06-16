// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"sort"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// HandlerContext carries contextual information passed through the dispatch
// table during AST extraction.
type HandlerContext struct {
	ParentName string
	Exported   bool
}

// Handler is the unified function signature for dispatch table entries.
type Handler func(node *sitter.Node, src []byte, c *Collector, ctx HandlerContext)

// Adapt wraps a simple extraction function (without HandlerContext) into a
// Handler.
func Adapt(fn func(node *sitter.Node, src []byte, c *Collector)) Handler {
	return func(node *sitter.Node, src []byte, c *Collector, _ HandlerContext) {
		fn(node, src, c)
	}
}

// AdaptParent wraps an extraction function that takes a parent name from
// HandlerContext.
func AdaptParent(fn func(node *sitter.Node, src []byte, c *Collector, parentName string)) Handler {
	return func(node *sitter.Node, src []byte, c *Collector, ctx HandlerContext) {
		fn(node, src, c, ctx.ParentName)
	}
}

// AdaptExported wraps an extraction function that takes an exported flag
// from HandlerContext.
func AdaptExported(fn func(node *sitter.Node, src []byte, c *Collector, exported bool)) Handler {
	return func(node *sitter.Node, src []byte, c *Collector, ctx HandlerContext) {
		fn(node, src, c, ctx.Exported)
	}
}

// DispatchTable maps tree-sitter node kinds to their Handler functions.
type DispatchTable map[string]Handler

// DispatchMeta holds metadata about the dispatch table entries.
type DispatchMeta struct {
	HandlerNames map[string]string
}

// MakeExtractFuncWithCallableRefs builds a Func from a DispatchTable
// that also runs callable reference detection and dataflow hint extraction
// after the main extraction.
func MakeExtractFuncWithCallableRefs(table DispatchTable, meta *DispatchMeta, cfg CallableRefConfig, dfCfgs ...DataflowConfig) Func {
	return makeExtractFunc(table, meta, cfg, nil, dfCfgs...)
}

// MakeExtractFuncWithFingerprint is like MakeExtractFuncWithCallableRefs but
// also accepts a body token map for fingerprint extraction.
func MakeExtractFuncWithFingerprint(table DispatchTable, meta *DispatchMeta, cfg CallableRefConfig, bodyTokenMap map[string]byte, dfCfgs ...DataflowConfig) Func {
	return makeExtractFunc(table, meta, cfg, bodyTokenMap, dfCfgs...)
}

func makeExtractFunc(table DispatchTable, meta *DispatchMeta, cfg CallableRefConfig, bodyTokenMap map[string]byte, dfCfgs ...DataflowConfig) Func {
	var dfCfg DataflowConfig
	if len(dfCfgs) > 0 {
		dfCfg = dfCfgs[0]
	}
	return func(root *sitter.Node, src []byte, c *Collector) {
		WalkTopLevel(root, src, c, table)
		c.ExtractCallableRefs(root, src, cfg)
		if meta != nil {
			seen := make(map[string]bool)
			var names []string
			for _, name := range meta.HandlerNames {
				if !seen[name] {
					seen[name] = true
					names = append(names, name)
				}
			}
			sort.Strings(names)
			for _, name := range names {
				c.AddEdge(types.Edge{
					From: "Dispatch",
					To:   name,
					Kind: types.EdgeCalls,
				})
			}
		}

		// Run dataflow hints BEFORE FileCallEdges so that alias keys
		// (object property names mapped to callables) are available as
		// known names for qualified call resolution.
		if len(dfCfg.AssignmentKinds) > 0 || len(dfCfg.ReturnKinds) > 0 || len(dfCfg.ObjectPairKinds) > 0 {
			knownCallables := c.KnownCallablesAndImports()
			for _, edge := range DataflowHints(root, src, "", &dfCfg, knownCallables) {
				c.AddEdge(edge)
			}
		}

		if dfCfg.CallTarget != nil {
			existingCalls := make(map[[2]string]bool)
			for _, e := range c.Edges {
				if e.Kind == types.EdgeCalls {
					existingCalls[[2]string{e.From, e.To}] = true
				}
			}
			// Use RichCallTarget when available so FileCallEdges can gate
			// qualified calls against known symbols. Fall back to wrapping
			// CallTarget (all calls treated as unqualified, no filtering).
			richFn := dfCfg.RichCallTarget
			if richFn == nil {
				plainFn := dfCfg.CallTarget
				richFn = func(node *sitter.Node, src []byte) CallTargetResult {
					return CallTargetResult{Name: plainFn(node, src)}
				}
			}
			// Include alias source names as known names so that qualified
			// calls like obj.prop() are not filtered out when prop is an
			// alias key pointing to a callable.
			for _, edge := range FileCallEdgesWithAliases(root, src, richFn, c.Symbols, c.Edges) {
				if !existingCalls[[2]string{edge.From, edge.To}] {
					c.AddEdge(edge)
				}
			}
		}
		if len(dfCfg.TypeRefKinds) > 0 {
			for _, edge := range FileTypeRefs(root, src, dfCfg.TypeRefKinds, c.Symbols) {
				c.AddEdge(edge)
			}
		}
		if c.Fingerprint && bodyTokenMap != nil {
			c.ExtractBodyTokens(root, src, bodyTokenMap)
		}
	}
}

// WalkTopLevel walks top-level children of the AST root and dispatches
// each to the appropriate handler.
func WalkTopLevel(root *sitter.Node, src []byte, c *Collector, table DispatchTable) {
	ctx := HandlerContext{}
	for i := range root.ChildCount() {
		child := root.Child(i)
		if child == nil || child.HasError() {
			continue
		}
		table.Dispatch(child, src, c, ctx)
	}
}

// Dispatch looks up a node kind in the table and calls the handler if found.
func (dt DispatchTable) Dispatch(node *sitter.Node, src []byte, c *Collector, ctx HandlerContext) {
	if handler, ok := dt[node.Kind()]; ok {
		handler(node, src, c, ctx)
	}
}

// WalkChildren iterates over the children of a body node and dispatches each
// one through the given table.
func WalkChildren(body *sitter.Node, src []byte, c *Collector, table DispatchTable, ctx HandlerContext) {
	for i := range body.ChildCount() {
		child := body.Child(i)
		if child == nil {
			continue
		}
		table.Dispatch(child, src, c, ctx)
	}
}
