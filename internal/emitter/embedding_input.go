// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxEmbedSourceRunes    = 2000
	maxEmbedStructureRunes = 1000
	truncationMarker       = "\n... truncated ...\n"
)

// embedInput builds a model input without repository-path or symbol-name
// leakage. It combines a normalized structural view with cleaned source so
// embeddings can use both cross-language shape and meaningful API names.
func embedInput(e *fpEntry) string {
	category := string(e.category)
	if category == "" {
		category = "unknown"
	}
	language := embeddingLanguage(e.filePath)

	var b strings.Builder
	b.WriteString("category:")
	b.WriteString(category)
	b.WriteString(" language:")
	b.WriteString(language)

	if structure := truncateEmbeddingText(tokensToText(e.tokens), maxEmbedStructureRunes); structure != "" {
		b.WriteString("\nstructure:")
		b.WriteString(structure)
	}

	source := stripSourceComments(e.sourceBody, language)
	source = maskSymbolIdentifier(source, e.name)
	source = truncateEmbeddingText(strings.TrimSpace(source), maxEmbedSourceRunes)
	if source != "" {
		b.WriteString("\nsource:\n")
		b.WriteString(source)
	}

	return b.String()
}

func embeddingLanguage(filePath string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	switch ext {
	case "cc", "cxx", "hpp", "hxx":
		return "cpp"
	case "cs":
		return "csharp"
	case "js", "jsx":
		return "javascript"
	case "py":
		return "python"
	case "rb":
		return "ruby"
	case "rs":
		return "rust"
	case "ts", "tsx":
		return "typescript"
	case "":
		return "unknown"
	default:
		return ext
	}
}

func stripSourceComments(source, language string) string {
	if source == "" {
		return ""
	}

	slashComments := language != "python" && language != "ruby"
	hashComments := language == "python" || language == "ruby" || language == "php"
	backtickEscapes := language == "javascript" || language == "typescript"

	const (
		stateCode = iota
		stateLineComment
		stateBlockComment
		stateSingleQuote
		stateDoubleQuote
		stateBacktick
	)

	var b strings.Builder
	b.Grow(len(source))
	state := stateCode
	escaped := false

	for i := 0; i < len(source); i++ {
		ch := source[i]
		var next byte
		if i+1 < len(source) {
			next = source[i+1]
		}

		switch state {
		case stateLineComment:
			if ch == '\n' {
				b.WriteByte(ch)
				state = stateCode
			}
		case stateBlockComment:
			if ch == '\n' {
				b.WriteByte(ch)
			} else if ch == '*' && next == '/' {
				i++
				state = stateCode
			}
		case stateSingleQuote, stateDoubleQuote, stateBacktick:
			b.WriteByte(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && (state != stateBacktick || backtickEscapes) {
				escaped = true
				continue
			}
			if (state == stateSingleQuote && ch == '\'') ||
				(state == stateDoubleQuote && ch == '"') ||
				(state == stateBacktick && ch == '`') {
				state = stateCode
			}
		default:
			switch {
			case slashComments && ch == '/' && next == '/':
				b.WriteByte(' ')
				i++
				state = stateLineComment
			case slashComments && ch == '/' && next == '*':
				b.WriteByte(' ')
				i++
				state = stateBlockComment
			case hashComments && ch == '#':
				b.WriteByte(' ')
				state = stateLineComment
			case ch == '\'':
				b.WriteByte(ch)
				state = stateSingleQuote
			case ch == '"':
				b.WriteByte(ch)
				state = stateDoubleQuote
			case ch == '`':
				b.WriteByte(ch)
				state = stateBacktick
			default:
				b.WriteByte(ch)
			}
		}
	}

	return b.String()
}

func maskSymbolIdentifier(source, scopedName string) string {
	name := scopedName
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		name = name[dot+1:]
	}
	if source == "" || name == "" {
		return source
	}

	var b strings.Builder
	b.Grow(len(source))
	for source != "" {
		r, size := utf8.DecodeRuneInString(source)
		if !isIdentifierRune(r, true) {
			b.WriteString(source[:size])
			source = source[size:]
			continue
		}

		end := size
		for end < len(source) {
			next, nextSize := utf8.DecodeRuneInString(source[end:])
			if !isIdentifierRune(next, false) {
				break
			}
			end += nextSize
		}
		identifier := source[:end]
		if identifier == name {
			b.WriteString("<symbol>")
		} else {
			b.WriteString(identifier)
		}
		source = source[end:]
	}
	return b.String()
}

func isIdentifierRune(r rune, first bool) bool {
	if r == '_' || unicode.IsLetter(r) {
		return true
	}
	return !first && unicode.IsDigit(r)
}

func truncateEmbeddingText(text string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}

	head := maxRunes / 2
	tail := maxRunes - head
	return string(runes[:head]) + truncationMarker + string(runes[len(runes)-tail:])
}
