// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// DecoratorNameFunc extracts a decorator/annotation name from a decorator node.
type DecoratorNameFunc func(node *sitter.Node, src []byte) string

// MakeDecoratorNameFunc returns a DecoratorNameFunc that extracts a decorator
// name from a decorator node. It handles three forms:
//
//   - @Foo            → identifier child → "Foo"
//   - @Foo.bar        → qualifiedKind child → last named leaf → "bar"
//   - @Foo()/@Foo.bar() → call/call_expression child → recurse into its first child
func MakeDecoratorNameFunc(qualifiedKind, callKind string) DecoratorNameFunc {
	return func(node *sitter.Node, src []byte) string {
		for i := range node.ChildCount() {
			child := node.Child(i)
			if child == nil {
				continue
			}
			switch child.Kind() {
			case "identifier":
				return child.Utf8Text(src)
			case qualifiedKind:
				return LastNamedLeaf(child, src)
			case callKind:
				if child.ChildCount() > 0 {
					first := child.Child(0)
					if first != nil {
						switch first.Kind() {
						case "identifier":
							return first.Utf8Text(src)
						case qualifiedKind:
							return LastNamedLeaf(first, src)
						}
					}
				}
			}
		}
		return ""
	}
}

// DecoratorEdges emits EdgeDecorates edges from each decorator name to
// the decorated symbol.
func DecoratorEdges(decoratedParent *sitter.Node, src []byte, c *Collector, targetName, decoratorKind string, nameFunc DecoratorNameFunc) {
	for i := range decoratedParent.ChildCount() {
		child := decoratedParent.Child(i)
		if child == nil || child.Kind() != decoratorKind {
			continue
		}
		name := nameFunc(child, src)
		if name != "" {
			c.AddEdge(types.Edge{
				From: name,
				To:   targetName,
				Kind: types.EdgeDecorates,
			})
		}
	}
}
