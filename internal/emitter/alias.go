// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package emitter serializes a SymbolGraph to structured text files.
package emitter

import (
	"sort"

	"codeknit/internal/ir"
)

// Dictionary maps repeated tokens (category/kind, property entries, edge kinds)
// to short codes for compact output.
type Dictionary struct {
	Forward map[string]string // token → short code
	Reverse map[string]string // short code → token
}

// NewDictionary builds a deterministic dictionary from a SymbolGraph.
// It collects all category/kind combinations, property entries, and edge kinds,
// then assigns sequential short codes to any token whose code is shorter than
// the original text (i.e. the code saves at least one character per occurrence).
func NewDictionary(sg *ir.SymbolGraph) *Dictionary {
	// Count occurrences of each token.
	counts := make(map[string]int)
	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		catKind := string(sym.Category) + "/" + sym.Kind
		counts[catKind]++

		for k, v := range sym.Properties {
			if v == "true" {
				counts[k]++
			} else {
				counts[k+"="+v]++
			}
		}
	}

	for i := range sg.Edges {
		counts[string(sg.Edges[i].Kind)]++
	}

	// Collect tokens that save space when replaced by a short code.
	// A code "dN" is 2-3 chars; include a token if total bytes saved > 0.
	tokens := make([]string, 0, len(counts))
	for token, count := range counts {
		tokens = append(tokens, token)
		_ = count // all tokens are candidates; sorted below for deterministic codes
	}
	sort.Strings(tokens)

	// Assign codes and keep only those that actually save bytes.
	dict := &Dictionary{
		Forward: make(map[string]string, len(tokens)),
		Reverse: make(map[string]string, len(tokens)),
	}
	codeIdx := 0
	for _, token := range tokens {
		code := indexToCode(codeIdx)
		count := counts[token]
		saved := (len(token) - len(code)) * count
		if saved > 0 {
			dict.Forward[token] = code
			dict.Reverse[code] = token
			codeIdx++
		}
	}
	return dict
}

// Encode returns the short code for a token if it exists in the dictionary,
// otherwise returns the token unchanged.
func (d *Dictionary) Encode(token string) string {
	if code, ok := d.Forward[token]; ok {
		return code
	}
	return token
}

// indexToCode converts a zero-based index to a sequential short code:
// 0→"d0", 1→"d1", ..., 9→"d9", 10→"d10", ...
func indexToCode(i int) string {
	return "d" + itoa(i)
}

// itoa is a simple int-to-string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
