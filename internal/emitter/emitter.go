// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/config"
	"codeknit/internal/ir"
	"codeknit/internal/outputfs"
	"codeknit/internal/plugin"
)

// EmitOptions controls emitter behavior.
type EmitOptions struct {
	OutputDir    string
	OutputMode   config.OutputMode
	OutputFormat config.OutputFormat
	InputPath    string
	FileOrder    []string
	MaxLines     int
	Minify       bool
	Clean        bool
}

// Emitter serializes a SymbolGraph to structured text files.
type Emitter struct{}

// Emit writes output files and returns the list of written file paths.
// For inline mode it writes to stdout and returns an empty slice.
// The SymbolGraph must have its indexes built (via BuildIndexes) before calling Emit.
func (e *Emitter) Emit(sg *ir.SymbolGraph, opts *EmitOptions) ([]string, error) {
	effective := *opts
	if effective.MaxLines == 0 {
		effective.MaxLines = config.DefaultParseMaxLines
	}
	if effective.OutputMode == config.OutputInline {
		if effective.OutputFormat == config.OutputFormatJSON {
			return e.emitJSON(sg, &effective)
		}
		return nil, e.EmitInline(os.Stdout, sg, &effective)
	}
	if effective.OutputFormat != config.OutputFormatJSON &&
		effective.OutputMode == config.OutputDirectoryTree && effective.MaxLines < 3 {
		return nil, fmt.Errorf("max-lines must be at least 3 for directory-tree SKT output")
	}
	if err := validateOutputLocation(effective.InputPath, effective.OutputDir); err != nil {
		return nil, err
	}
	managed := func(path string) bool {
		if effective.OutputFormat == config.OutputFormatJSON {
			return path == "codeknit.json"
		}
		return filepath.Ext(path) == ".skt"
	}
	return outputfs.ReplaceDirectory(effective.OutputDir, effective.Clean, managed, func(stage string) ([]string, error) {
		effective.OutputDir = stage
		switch {
		case effective.OutputFormat == config.OutputFormatJSON:
			return e.emitJSON(sg, &effective)
		case effective.OutputMode == config.OutputDirectoryTree:
			return e.emitDirectoryTree(sg, &effective)
		default:
			return e.emitDirectoryFlat(sg, &effective)
		}
	})
}

func validateOutputLocation(input, output string) error {
	if input == "" {
		return nil
	}
	base, err := filepath.Abs(input)
	if err != nil {
		return err
	}
	if info, statErr := os.Stat(base); statErr == nil && !info.IsDir() {
		base = filepath.Dir(base)
	}
	out, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	base, err = physicalPath(base)
	if err != nil {
		return err
	}
	out, err = physicalPath(out)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(out, base); err == nil && (rel == "." || filepath.IsLocal(rel)) {
		return fmt.Errorf("output directory must not be the source directory or its ancestor")
	}
	return nil
}

// Resolve existing ancestors so symlink aliases cannot hide source/output
// overlap. The final output directory need not exist yet.
func physicalPath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) || filepath.Dir(path) == path {
		return "", err
	}
	parent, err := physicalPath(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

// fileBlock holds the rendered lines for one source file's symbols and edges.
type fileBlock struct {
	filePath    string
	symbolLines []string // includes the "## filepath" header
	edgeLines   []string
}

// buildBlocks renders file blocks for all source files using the pre-built
// indexes in the SymbolGraph. No index-building or resolution happens here;
// the IR carries everything needed.
func buildBlocks(sg *ir.SymbolGraph, dict *Dictionary) []fileBlock {
	edgesByFile := buildEdgeLines(sg, dict)

	blocks := make([]fileBlock, 0, len(sg.FileOrder))
	for _, fp := range sg.FileOrder {
		idxs := sg.ByFile[fp]
		lines := make([]string, 0, 1+len(idxs))
		lines = append(lines, fmt.Sprintf("## %s", fp))

		for _, idx := range idxs {
			sym := sg.Symbols[idx]
			sid := sg.ShortIDs[sym.ID]
			if !sym.IsFileMetadata() {
				sym.Signature = resolveTypeRefs(sym.Signature, sym.Name, fp, sg.ResolveTypeSID)
			}
			lines = append(lines, formatSymbolLine(sid, &sym, dict))
		}

		blocks = append(blocks, fileBlock{
			filePath:    fp,
			symbolLines: lines,
			edgeLines:   edgesByFile[fp],
		})
	}

	return blocks
}

// formatSymbolLine formats a single symbol line for the output.
// Format: "ShortID category/kind L{start}-L{end} signature {props}"
// When there are no properties, the trailing {} is omitted to save tokens.
func formatSymbolLine(shortID string, sym *plugin.Symbol, dict *Dictionary) string {
	catKind := string(sym.Category) + "/" + sym.Kind
	if dict != nil {
		catKind = dict.Encode(catKind)
	}

	propsStr := formatProperties(sym.Properties, dict)
	// Declarations may span source lines (tuple results, defaults, inline
	// types). SKT records stay on one physical line; JSON retains the source.
	signature := escapeLineBreaks(sym.Signature)
	if propsStr == "" {
		return fmt.Sprintf("%s %s L%d-L%d %s",
			shortID,
			catKind,
			sym.Span[0], sym.Span[1],
			signature,
		)
	}
	return fmt.Sprintf("%s %s L%d-L%d %s %s",
		shortID,
		catKind,
		sym.Span[0], sym.Span[1],
		signature,
		propsStr,
	)
}

func escapeLineBreaks(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r", `\r`), "\n", `\n`)
}

// formatProperties formats the properties map as {key1, key2, ...} listing keys
// whose values are "true", sorted alphabetically. Properties with value "false"
// are omitted entirely to keep the output clean and consistent across languages.
// Returns an empty string when there are no properties.
func formatProperties(props map[string]string, dict *Dictionary) string {
	if len(props) == 0 {
		return ""
	}
	var keys []string
	for k, v := range props {
		entry := k
		if v != "true" {
			entry = k + "=" + escapeLineBreaks(v)
		}
		if dict != nil {
			entry = dict.Encode(entry)
		}
		keys = append(keys, entry)
	}
	sort.Strings(keys)
	return "{" + strings.Join(keys, ", ") + "}"
}

// resolveTypeRefs replaces type-category symbol names in a signature with their short IDs.
// It only replaces identifiers that appear in type positions:
//   - After ": " (parameter/variable type annotations)
//   - After "-> " (return types)
//   - Inside "<>" (generic type arguments)
//
// selfName is the symbol's own name, which is never replaced.
func resolveTypeRefs(sig, selfName, filePath string, resolve func(string, string) (string, bool)) string {
	var b strings.Builder
	b.Grow(len(sig))

	i := 0
	inTypePos := false // true when scanning a type position
	angleBrackets := 0 // nesting depth of <>

	for i < len(sig) {
		ch := sig[i]

		// Track <> nesting — inside generics is always a type position.
		if ch == '<' {
			angleBrackets++
			b.WriteByte(ch)
			i++
			inTypePos = true
			continue
		}
		if ch == '>' && angleBrackets > 0 {
			angleBrackets--
			b.WriteByte(ch)
			i++
			if angleBrackets == 0 {
				inTypePos = false
			}
			continue
		}

		// Detect "-> " prefix for return type position.
		if i+3 <= len(sig) && sig[i:i+3] == "-> " {
			b.WriteString("-> ")
			i += 3
			inTypePos = true
			continue
		}

		// Detect ": " prefix for type annotation position.
		if i+2 <= len(sig) && sig[i:i+2] == ": " {
			b.WriteString(": ")
			i += 2
			inTypePos = true
			continue
		}

		// If we hit a non-identifier, non-type character, exit type position
		// (unless inside angle brackets).
		if inTypePos && angleBrackets == 0 && (ch == ',' || ch == ')' || ch == '(') {
			inTypePos = false
		}

		if isIdentStart(ch) {
			j := i + 1
			for j < len(sig) && isIdentPart(sig[j]) {
				j++
			}
			word := sig[i:j]
			if inTypePos && word != selfName {
				if sid, ok := resolve(filePath, word); ok {
					b.WriteString(sid)
					i = j
					continue
				}
			}
			b.WriteString(word)
			i = j
		} else {
			b.WriteByte(ch)
			i++
		}
	}
	return b.String()
}

func isIdentStart(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}
