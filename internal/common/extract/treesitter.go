// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"fmt"
	"unsafe"

	"codeknit/internal/common/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Func is the language-specific function that walks the tree-sitter
// AST root and populates the Collector with symbols and edges.
type Func func(root *sitter.Node, src []byte, c *Collector)

// SyntaxError indicates that a file had syntax errors but partial results
// were still extracted.
type SyntaxError struct {
	FilePath string
	Message  string
}

func (w *SyntaxError) Error() string {
	return w.FilePath + ": " + w.Message
}

// AsError returns the SyntaxError as an error interface, returning nil if w is nil.
func (w *SyntaxError) AsError() error {
	if w == nil {
		return nil
	}
	return w
}

// ParseWithTreeSitter is the shared parse pipeline used by every language plugin.
func ParseWithTreeSitter(
	filePath string,
	src []byte,
	tsLang unsafe.Pointer,
	fn Func,
	fingerprint ...bool,
) ([]types.Symbol, []types.Edge, error) {
	parser := sitter.NewParser()
	defer parser.Close()

	lang := sitter.NewLanguage(tsLang)
	if err := parser.SetLanguage(lang); err != nil {
		return nil, nil, fmt.Errorf("set language: %w", err)
	}

	tree := parser.Parse(src, nil)
	defer tree.Close()

	root := tree.RootNode()
	var syntaxWarning *SyntaxError
	if root.HasError() {
		line, col := FindFirstError(root)
		syntaxWarning = &SyntaxError{
			FilePath: filePath,
			Message:  fmt.Sprintf("%s:%d:%d: syntax error", filePath, line+1, col+1),
		}
	}

	c := &Collector{FilePath: filePath}
	if len(fingerprint) > 0 && fingerprint[0] {
		c.Fingerprint = true
	}
	fn(root, src, c)
	if c.Symbols == nil {
		c.Symbols = []types.Symbol{}
	}
	if c.Edges == nil {
		c.Edges = []types.Edge{}
	}
	return c.Symbols, c.Edges, syntaxWarning.AsError()
}
