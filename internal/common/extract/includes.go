// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"bytes"
	"path/filepath"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Include records C/C++ file and include endpoints explicitly. An include
// symbol describes the written directive, not a resolved header declaration:
// resolving headers requires the compiler's include search configuration.
func Include(node *sitter.Node, src []byte, c *Collector) {
	var target string
	span := NodeSpan(node)
	for i := range node.NamedChildCount() {
		child := node.NamedChild(i)
		if child.Kind() == "system_lib_string" || child.Kind() == "string_literal" {
			target = child.Utf8Text(src)
			// preproc_include includes its trailing newline in the AST.
			span[1] = NodeSpan(child)[1]
			break
		}
	}
	if target == "" {
		return
	}

	fileName := filepath.Base(c.FilePath)
	if c.includeSymbols == nil {
		c.includeSymbols = make(map[string]bool)
		c.AddSymbol(&types.Symbol{
			Name: fileName, Signature: fileName, Category: types.CategoryMeta,
			Kind: "file", Span: [2]int{1, bytes.Count(src, []byte("\n")) + 1},
		})
	}
	if c.includeSymbols[target] {
		return
	}
	c.includeSymbols[target] = true
	c.AddSymbol(&types.Symbol{
		Name: target, Signature: target, Category: types.CategoryMeta,
		Kind: "include", Span: span,
		Properties: types.NewProps().Set("resolution", "unresolved").Map(),
	})
	c.AddEdge(&types.Edge{
		From: fileName, To: target, Kind: types.EdgeReferences, Resolution: "unresolved",
	})
}
