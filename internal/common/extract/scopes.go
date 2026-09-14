// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"strings"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// symbolScopes associates declarations with AST byte ranges. Line ranges alone
// cannot distinguish adjacent one-line functions or nested declarations.
type symbolScope struct {
	node *sitter.Node
	sym  *types.Symbol
}

type symbolScopes struct {
	byRange      map[[2]uint]*types.Symbol
	declarations []symbolScope
}

func newSymbolScopes(root *sitter.Node, src []byte, symbols []types.Symbol) symbolScopes {
	scopes := symbolScopes{byRange: make(map[[2]uint]*types.Symbol)}
	byName := make(map[string][]*types.Symbol)
	for i := range symbols {
		sym := &symbols[i]
		if sym.Category == types.CategoryCallable || sym.Category == types.CategoryType {
			byName[sym.Name] = append(byName[sym.Name], sym)
		}
	}
	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		name := declarationName(n)
		if name != nil {
			for _, sym := range byName[name.Utf8Text(src)] {
				if sym.ByteSpan != [2]uint{} && sym.ByteSpan != NodeByteSpan(n) {
					continue
				}
				if declarationScopeMatches(n, sym, src) &&
					int(n.StartPosition().Row)+1 >= sym.Span[0] && //nolint:gosec // source line number
					int(n.EndPosition().Row) < sym.Span[1] && //nolint:gosec // source line number
					(n.ChildByFieldName("body") != nil || int(n.StartPosition().Row)+1 == sym.Span[0] && int(n.EndPosition().Row)+1 == sym.Span[1]) { //nolint:gosec // source line numbers
					scopes.declarations = append(scopes.declarations, symbolScope{node: n, sym: sym})
					scopes.byRange[[2]uint{n.StartByte(), n.EndByte()}] = sym
				}
			}
		}
		for i := range n.NamedChildCount() {
			walk(n.NamedChild(i))
		}
	}
	walk(root)
	return scopes
}

func declarationScopeMatches(n *sitter.Node, sym *types.Symbol, src []byte) bool {
	i := strings.LastIndex(sym.EffectiveScopedName(), ".")
	if i < 0 {
		return true
	}
	parent := sym.EffectiveScopedName()[:i]
	if j := strings.LastIndex(parent, "."); j >= 0 {
		parent = parent[j+1:]
	}
	if receiver := n.ChildByFieldName("receiver"); receiver != nil {
		for i := range receiver.NamedChildCount() {
			if typ := receiver.NamedChild(i).ChildByFieldName("type"); typ != nil {
				return baseTypeName(typ.Utf8Text(src)) == parent
			}
		}
	}
	for ancestor := n.Parent(); ancestor != nil; ancestor = ancestor.Parent() {
		if ancestor.Kind() == "impl_item" {
			if typ := ancestor.ChildByFieldName("type"); typ != nil {
				return baseTypeName(typ.Utf8Text(src)) == parent
			}
		}
		if name := declarationName(ancestor); name != nil && ancestor.ChildByFieldName("body") != nil {
			return name.Utf8Text(src) == parent
		}
	}
	return true // Go receivers and Rust impl blocks are not named ancestors.
}

func baseTypeName(name string) string {
	name = strings.TrimLeft(name, "*&")
	if i := strings.IndexAny(name, "[<"); i >= 0 {
		name = name[:i]
	}
	return name
}

// Restrict matching to declarations: an identifier in a call expression must
// never masquerade as a declaration on the same source line.
func declarationName(n *sitter.Node) *sitter.Node {
	kind := n.Kind()
	if n.ChildByFieldName("body") == nil &&
		!strings.Contains(kind, "declaration") && !strings.Contains(kind, "definition") &&
		!strings.Contains(kind, "declarator") && !strings.Contains(kind, "item") &&
		!strings.Contains(kind, "signature") && kind != "method" && kind != "singleton_method" {
		return nil
	}
	if name := n.ChildByFieldName("name"); name != nil {
		return name
	}
	for decl := n.ChildByFieldName("declarator"); decl != nil; decl = decl.ChildByFieldName("declarator") {
		if decl.Kind() == "identifier" || decl.Kind() == "field_identifier" {
			return decl
		}
	}
	return ChildByKind(n, "identifier")
}

func (scopes symbolScopes) owner(n *sitter.Node, callableOnly bool) *types.Symbol {
	for ancestor := n; ancestor != nil; ancestor = ancestor.Parent() {
		if sym := scopes.byRange[[2]uint{ancestor.StartByte(), ancestor.EndByte()}]; sym != nil {
			if !callableOnly || sym.Category == types.CategoryCallable {
				return sym
			}
		}
	}
	return nil
}
