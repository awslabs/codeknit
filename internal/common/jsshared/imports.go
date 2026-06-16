// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package jsshared

import (
	"codeknit/internal/common/extract"
	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// ExtractJSImportStatement extracts an import_statement node and emits
// EdgeImports edges. Shared by the JavaScript and TypeScript plugins.
func ExtractJSImportStatement(node *sitter.Node, src []byte, c *extract.Collector) {
	var modulePath string
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "string" {
			raw := child.Utf8Text(src)
			if len(raw) >= 2 {
				modulePath = raw[1 : len(raw)-1]
			}
		}
	}
	if modulePath == "" {
		return
	}

	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "import_clause":
			jsExtractImportClause(child, src, c, modulePath)
		case "named_imports":
			jsExtractNamedImports(child, src, c, modulePath)
		}
	}
}

func jsExtractImportClause(node *sitter.Node, src []byte, c *extract.Collector, modulePath string) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier":
			name := child.Utf8Text(src)
			if name != "" {
				c.AddEdge(types.Edge{From: name, To: modulePath, Kind: types.EdgeImports})
			}
		case "named_imports":
			jsExtractNamedImports(child, src, c, modulePath)
		case "namespace_import":
			name := extract.ChildText(child, "identifier", src)
			if name != "" {
				c.AddEdge(types.Edge{From: name, To: modulePath, Kind: types.EdgeImports})
			}
		}
	}
}

func jsExtractNamedImports(node *sitter.Node, src []byte, c *extract.Collector, modulePath string) {
	for i := range node.ChildCount() {
		child := node.Child(i)
		if child == nil || child.Kind() != "import_specifier" {
			continue
		}
		var names []string
		for j := range child.ChildCount() {
			gc := child.Child(j)
			if gc != nil && gc.Kind() == "identifier" {
				names = append(names, gc.Utf8Text(src))
			}
		}
		if len(names) > 0 {
			c.AddEdge(types.Edge{From: names[len(names)-1], To: modulePath, Kind: types.EdgeImports})
		}
	}
}
