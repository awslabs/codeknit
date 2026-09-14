// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import sitter "github.com/tree-sitter/go-tree-sitter"

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

// RegisterContainers makes syntax-only wrappers transparent to extraction.
// Only explicitly registered wrappers are traversed; function bodies remain
// opaque so local variables cannot become top-level declarations.
func (dt DispatchTable) RegisterContainers(kinds ...string) {
	for _, kind := range kinds {
		dt[kind] = func(node *sitter.Node, src []byte, c *Collector, ctx HandlerContext) {
			WalkChildren(node, src, c, dt, ctx)
		}
	}
}

// MakeExtractFuncWithCallableRefs builds a Func from a DispatchTable
// that also runs callable reference detection and dataflow hint extraction
// after the main extraction.
func MakeExtractFuncWithCallableRefs(table DispatchTable, cfg CallableRefConfig, dfCfgs ...DataflowConfig) Func {
	return makeExtractFunc(table, cfg, nil, dfCfgs...)
}

// MakeExtractFuncWithFingerprint is like MakeExtractFuncWithCallableRefs but
// also accepts a body token map for fingerprint extraction.
func MakeExtractFuncWithFingerprint(table DispatchTable, cfg CallableRefConfig, bodyTokenMap map[string]byte, dfCfgs ...DataflowConfig) Func {
	return makeExtractFunc(table, cfg, bodyTokenMap, dfCfgs...)
}

func makeExtractFunc(table DispatchTable, cfg CallableRefConfig, bodyTokenMap map[string]byte, dfCfgs ...DataflowConfig) Func {
	var dfCfg DataflowConfig
	if len(dfCfgs) > 0 {
		dfCfg = dfCfgs[0]
	}
	return func(root *sitter.Node, src []byte, c *Collector) {
		WalkTopLevel(root, src, c, table)
		c.ExtractCallableRefs(root, src, cfg)
		c.Edges = append(c.Edges, parameterEdges(root, src, c.Symbols)...)

		if len(dfCfg.AssignmentKinds) > 0 || len(dfCfg.ReturnKinds) > 0 || len(dfCfg.ObjectPairKinds) > 0 {
			c.Edges = append(c.Edges, DataflowHints(root, src, &dfCfg, c.Symbols)...)
		}
		richFn := dfCfg.RichCallTarget
		if richFn == nil && dfCfg.CallTarget != nil {
			richFn = func(node *sitter.Node, src []byte) CallTargetResult {
				return CallTargetResult{Name: dfCfg.CallTarget(node, src)}
			}
		}
		c.Edges = append(c.Edges, FileCallEdges(root, src, richFn, c.Symbols)...)
		if len(dfCfg.TypeRefKinds) > 0 {
			c.Edges = append(c.Edges, FileTypeRefs(root, src, dfCfg.TypeRefKinds, c.Symbols)...)
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
