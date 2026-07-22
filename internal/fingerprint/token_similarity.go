// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import "codeknit/internal/plugin"

// TokenEditSimilarity computes a normalized edit-distance similarity (0–100)
// directly on raw body-token streams after stripping variable ordinals and
// value hashes. This captures structural similarity that CTPH may miss due
// to chunk-boundary shifts from small insertions/deletions.
//
// The stripping produces a "structural skeleton" containing only semantic
// operation tokens (control flow, operators, calls) — the same tokens that
// define the function's shape — making the comparison insensitive to
// variable naming patterns and literal values.
func TokenEditSimilarity(a, b []byte) int {
	return normalizedEditSimilarity(structuralTokens(a), structuralTokens(b))
}

// RawTokenEditSimilarity computes normalized edit similarity without
// discarding payload bytes. It is used for type shapes and value initializers,
// where names and literal values are meaningful parts of the fingerprint.
func RawTokenEditSimilarity(a, b []byte) int {
	return normalizedEditSimilarity(a, b)
}

func normalizedEditSimilarity(a, b []byte) int {
	if len(a) == 0 && len(b) == 0 {
		return 100
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	dist := editDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	score := 100 - (dist*100)/maxLen
	if score < 0 {
		return 0
	}
	return score
}

// structuralTokens extracts only the semantic operation tokens from a
// body-token stream, stripping variable ordinals (per-scope byte values
// not in the token alphabet), literal value hashes (4 bytes after each
// literal token), callee name hashes (2 bytes after FPCall), and
// container element hashes.
//
// The result is a compact byte slice representing the function's control
// flow and operation structure — ideal for edit-distance comparison.
func structuralTokens(tokens []byte) []byte {
	out := make([]byte, 0, len(tokens))
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		i++

		// FPVar: skip the tagged variable ordinal. The tag prevents ordinal
		// bytes from being mistaken for semantic operation tokens.
		if tok == plugin.FPVar {
			if i < len(tokens) {
				i++
			}
			continue
		}

		// FPCall: keep the call token, skip 2-byte callee hash + 1-byte arg count.
		if tok == plugin.FPCall {
			out = append(out, tok)
			i += min(2, len(tokens)-i) // skip callee hash
			if i < len(tokens) {
				i++ // skip arg count
			}
			continue
		}

		// Literals: keep the type token, skip 4-byte value hash.
		if tok == plugin.FPLitNum || tok == plugin.FPLitStr ||
			tok == plugin.FPLitBool || tok == plugin.FPLitNil {
			out = append(out, tok)
			i += min(4, len(tokens)-i)
			continue
		}

		// Dict/Array: keep the type token, skip count + N*4-byte hashes.
		if tok == plugin.FPDict || tok == plugin.FPArray {
			out = append(out, tok)
			if i < len(tokens) {
				count := int(tokens[i])
				i += 1 + count*4
				if i > len(tokens) {
					i = len(tokens)
				}
			}
			continue
		}

		// Known semantic token — keep it.
		if isSemanticToken(tok) {
			out = append(out, tok)
			continue
		}

		// Everything else (variable ordinals, unknown bytes) — skip.
	}
	return out
}

// isSemanticToken returns true if the byte is a known FPToken constant
// (control flow, operators, operations) rather than a variable ordinal.
func isSemanticToken(tok byte) bool {
	// Control flow: 0x01–0x1C
	if tok >= 0x01 && tok <= 0x1C {
		return true
	}
	// Arithmetic: 0x20–0x24
	if tok >= 0x20 && tok <= 0x24 {
		return true
	}
	// Comparison: 0x28–0x2D
	if tok >= 0x28 && tok <= 0x2D {
		return true
	}
	// Logical: 0x30–0x32
	if tok >= 0x30 && tok <= 0x32 {
		return true
	}
	// Bitwise: 0x38–0x3D
	if tok >= 0x38 && tok <= 0x3D {
		return true
	}
	// Literals/containers: 0x40–0x45
	if tok >= 0x40 && tok <= 0x45 {
		return true
	}
	return false
}
