// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package ir defines the intermediate representation types for codeknit.
package ir

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/plugin"
)

// SymbolGraph is the intermediate representation of the entire analyzed codebase.
// After construction by the planner, it carries pre-built indexes so that
// downstream phases (emitter) can serialize without re-computing lookups.
type SymbolGraph struct {
	// ByFile maps file path → slice of indexes into Symbols.
	ByFile map[string][]int

	// ShortIDs maps global symbol ID → short display ID (e.g. "S1", "S2").
	ShortIDs map[string]string

	// SymbolFile maps global symbol ID → file path.
	SymbolFile map[string]string

	// typeIndex maps (filePath, typeName) → shortID for same-file type resolution.
	typeIndex map[string]map[string]string
	// globalTypeIndex maps typeName → candidates for cross-file type resolution.
	globalTypeIndex map[string][]typeCandidate

	Symbols []plugin.Symbol
	Edges   []plugin.Edge
	Errors  []ParseError

	// FileOrder lists source file paths in deterministic (sorted) order.
	FileOrder []string
}

// typeCandidate holds a short ID and its source file for cross-file type resolution.
type typeCandidate struct {
	ShortID string
	File    string
}

// BuildIndexes populates all derived indexes from the Symbols and Edges slices.
// This must be called after Symbols are fully populated (by the planner or tests).
// It computes FileOrder, ByFile, ShortIDs, SymbolFile, and type resolution indexes.
func (sg *SymbolGraph) BuildIndexes() {
	// Group symbols by file.
	groups := make(map[string][]int)
	for i := range sg.Symbols {
		fp := sg.Symbols[i].FilePath
		groups[fp] = append(groups[fp], i)
	}

	// Sorted file order.
	sg.FileOrder = make([]string, 0, len(groups))
	for fp := range groups {
		sg.FileOrder = append(sg.FileOrder, fp)
	}
	sort.Strings(sg.FileOrder)

	sg.ByFile = groups

	// Assign short IDs in deterministic order: iterate files sorted, symbols in slice order.
	sg.ShortIDs = make(map[string]string, len(sg.Symbols))
	sg.SymbolFile = make(map[string]string, len(sg.Symbols))
	counter := 0
	for _, fp := range sg.FileOrder {
		for _, idx := range sg.ByFile[fp] {
			counter++
			id := sg.Symbols[idx].ID
			sg.ShortIDs[id] = fmt.Sprintf("S%d", counter)
			sg.SymbolFile[id] = fp
		}
	}

	// Build type resolution indexes (only type-category symbols).
	sg.typeIndex = make(map[string]map[string]string)
	sg.globalTypeIndex = make(map[string][]typeCandidate)
	for _, fp := range sg.FileOrder {
		sg.typeIndex[fp] = make(map[string]string)
		for _, idx := range sg.ByFile[fp] {
			sym := &sg.Symbols[idx]
			if sym.Category == plugin.CategoryType {
				sid := sg.ShortIDs[sym.ID]
				sg.typeIndex[fp][sym.Name] = sid
				sg.globalTypeIndex[sym.Name] = append(sg.globalTypeIndex[sym.Name], typeCandidate{
					ShortID: sid,
					File:    fp,
				})
			}
		}
	}
}

// ResolveTypeSID resolves a type name to a short ID for a given file context.
// It checks same-file first, then falls back to cross-file resolution using
// directory proximity with deterministic lexicographic tiebreaking.
// Returns the short ID and true if found, or ("", false) if unresolved.
func (sg *SymbolGraph) ResolveTypeSID(filePath, typeName string) (string, bool) {
	// Same-file first.
	if m, ok := sg.typeIndex[filePath]; ok {
		if sid, ok := m[typeName]; ok {
			return sid, true
		}
	}

	// Cross-file fallback.
	candidates := sg.globalTypeIndex[typeName]
	if len(candidates) == 0 {
		return "", false
	}
	if len(candidates) == 1 {
		return candidates[0].ShortID, true
	}

	// Pick closest by directory proximity; break ties lexicographically
	// on file path to guarantee deterministic, unique resolution.
	sourceDir := filepath.Dir(filePath)
	best := candidates[0]
	bestDist := PathDistance(sourceDir, filepath.Dir(best.File))
	for _, c := range candidates[1:] {
		d := PathDistance(sourceDir, filepath.Dir(c.File))
		if d < bestDist || (d == bestDist && c.File < best.File) {
			bestDist = d
			best = c
		}
	}
	return best.ShortID, true
}

// ParseError represents a file that had syntax/parse errors during analysis.
type ParseError struct {
	FilePath string
	Reason   string
}

func (e *ParseError) Error() string {
	return e.FilePath + ": " + e.Reason
}

// PathDistance returns a rough measure of how far apart two directory paths are.
// 0 means same directory, higher means further apart.
func PathDistance(a, b string) int {
	if a == b {
		return 0
	}
	partsA := strings.Split(filepath.ToSlash(a), "/")
	partsB := strings.Split(filepath.ToSlash(b), "/")
	common := 0
	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		if partsA[i] != partsB[i] {
			break
		}
		common++
	}
	return (len(partsA) - common) + (len(partsB) - common)
}
