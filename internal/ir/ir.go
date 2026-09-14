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

	Symbols []plugin.Symbol
	Edges   []plugin.Edge
	Errors  []ParseError

	// FileOrder lists source file paths in deterministic (sorted) order.
	FileOrder []string
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
	// Include files with relationships but no extracted declarations (for
	// example, a header containing only includes). SKT must retain their edges.
	for i := range sg.Edges {
		if fp, _, ok := strings.Cut(sg.Edges[i].From, "::"); ok && fp != "" {
			if _, exists := groups[fp]; !exists {
				groups[fp] = nil
			}
		}
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

	// Index unique local types and confirmed cross-file type references only.
	// A competing or unresolved reference suppresses substitution of that name.
	sg.typeIndex = make(map[string]map[string]string)
	types := make(map[string]bool)
	add := func(fp, name, sid string) {
		if sg.typeIndex[fp] == nil {
			sg.typeIndex[fp] = make(map[string]string)
		}
		old, exists := sg.typeIndex[fp][name]
		if exists && old != sid {
			sg.typeIndex[fp][name] = ""
		} else {
			sg.typeIndex[fp][name] = sid
		}
	}
	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		if sym.Category == plugin.CategoryType {
			types[sym.ID] = true
			add(sym.FilePath, sym.Name, sg.ShortIDs[sym.ID])
		}
	}
	for i := range sg.Edges {
		e := &sg.Edges[i]
		if e.Kind != plugin.EdgeReferences && e.Kind != plugin.EdgeInherits && e.Kind != plugin.EdgeImplements {
			continue
		}
		fp := sg.SymbolFile[e.From]
		name := e.TargetName
		if name == "" {
			_, name, _ = strings.Cut(e.To, "::")
		}
		if fp == "" || name == "" {
			continue
		}
		if sg.EdgeResolution(e) != "" {
			add(fp, name, "")
			continue
		}
		if types[e.To] {
			add(fp, name, sg.ShortIDs[e.To])
		}
	}
}

// ResolveTypeSID only substitutes a name when the available relationships
// identify one type. It never selects a type by directory proximity.
func (sg *SymbolGraph) ResolveTypeSID(filePath, typeName string) (string, bool) {
	sid := sg.typeIndex[filePath][typeName]
	return sid, sid != ""
}

// EdgeResolution also handles graphs built directly by callers, where missing
// endpoints might not yet have an explicit resolution status.
func (sg *SymbolGraph) EdgeResolution(e *plugin.Edge) string {
	if e.Resolution != "" {
		return e.Resolution
	}
	if sg.ShortIDs[e.From] == "" || sg.ShortIDs[e.To] == "" {
		return "unresolved"
	}
	return ""
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
