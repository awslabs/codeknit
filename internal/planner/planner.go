// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package planner builds a SymbolGraph from per-file extraction results.
package planner

import (
	"fmt"
	"path/filepath"
	"sort"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// Planner builds a SymbolGraph from per-file Symbols and Edges.
type Planner struct{}

// Plan takes per-file Symbols and Edges and produces a unified SymbolGraph.
// Symbol IDs are assigned using the format "filePath::SymbolName".
// Intra-file Edge references are remapped to global IDs.
// Cross-file references are resolved using import context when available,
// preserving uncertainty when imports and lexical scope cannot identify a target.
func (p *Planner) Plan(
	fileSymbols map[string][]plugin.Symbol,
	fileEdges map[string][]plugin.Edge,
) *ir.SymbolGraph {
	sg := &ir.SymbolGraph{}

	// Process files in deterministic order.
	files := plugin.SortedStringKeys(fileSymbols)

	// Phase 0: Deduplicate module/package symbols.
	// Each file-level extractor runs independently (like a compiler's
	// per-translation-unit parse), so every Go file emits "package main",
	// every Java file emits "package com.example", etc. The planner acts
	// as the linker: it sees all translation units and merges duplicate
	// module symbols that share the same directory and name into one.
	// Structural containers (Ruby modules, C++ namespaces) that wrap code
	// are typically unique per file and pass through unaffected.
	type dirName struct{ dir, name, language string }
	moduleCanonFile := make(map[dirName]string) // (dir,name) → first file that declared it
	filePackages := make(map[string]string)

	for _, fp := range files {
		symbols := fileSymbols[fp]
		dir := filepath.Dir(fp)
		filtered := symbols[:0:0]
		for i := range symbols {
			sym := &symbols[i]
			// Scoped modules are lexical containers (for example Rust inline
			// modules), not repeated file-level package declarations.
			if sym.Category != plugin.CategoryModule || sym.ScopedName != "" {
				filtered = append(filtered, *sym)
				continue
			}
			key := dirName{dir: dir, name: sym.Name, language: languageFamily(fp)}
			filePackages[fp] = sym.Name
			if _, seen := moduleCanonFile[key]; seen {
				// Duplicate — drop it from this file's symbols.
				continue
			}
			moduleCanonFile[key] = fp
			filtered = append(filtered, *sym)
		}
		fileSymbols[fp] = filtered
	}

	// Phase 1: Assign globally-unique IDs and collect all symbols.
	localToGlobal := make(map[fileLocal]string)
	nameCount := make(map[fileLocal]int)

	for _, fp := range files {
		symbols := fileSymbols[fp]
		for i := range symbols {
			sym := &symbols[i]
			sym.FilePath = fp

			scopedName := sym.EffectiveScopedName()
			key := fileLocal{file: fp, name: scopedName}
			count := nameCount[key]
			nameCount[key] = count + 1

			globalID := fp + "::" + scopedName
			if count > 0 {
				globalID = fmt.Sprintf("%s::%s#%d", fp, scopedName, count)
			}
			sym.ID = globalID

			if count == 0 {
				localToGlobal[key] = globalID
			}

			sg.Symbols = append(sg.Symbols, *sym)
		}
	}

	// Phase 2: Build per-file import maps from EdgeImports edges.
	// importMap[filePath][localName] = modulePath
	// This tells us: "in file F, the name N was imported from module M."
	resolver := newResolver(sg.Symbols, fileEdges)
	resolver.packages = filePackages
	resolver.files = files

	// Phase 3: Remap edges and resolve cross-file references.
	for _, fp := range files {
		edges := fileEdges[fp]
		for i := range edges {
			e := &edges[i]
			// Skip import edges — they're metadata consumed by the planner,
			// not structural relationships to include in the output graph.
			if e.Kind == plugin.EdgeImports {
				continue
			}

			remapped := resolver.edge(fp, e)
			// Overloads share a declaration name, but all belong to the same
			// lexical owner. Containment can include each overload even when a
			// call to that name cannot select one of them.
			if e.Kind == plugin.EdgeContains && remapped.Resolution == "ambiguous" &&
				parentScope(e.To) == e.From && len(resolver.local[fileLocal{fp, e.From}]) == 1 {
				for _, target := range remapped.Candidates {
					child := remapped
					child.To, child.Resolution, child.Candidates = target, "", nil
					sg.Edges = append(sg.Edges, child)
				}
				continue
			}

			sg.Edges = append(sg.Edges, remapped)
		}
	}

	// Phase 3b: Add contains edges from module symbols to their top-level siblings.
	// A package/namespace symbol logically contains all top-level declarations
	// in the same directory AND same language. This connects otherwise-isolated
	// constants, variables, and types to their parent module in the graph.
	//
	// Iterate the module map in sorted key order so the resulting edge
	// sequence is deterministic across runs (Go randomizes map iteration).
	contained := make(map[string]bool)
	for i := range sg.Edges {
		edge := &sg.Edges[i]
		if edge.Kind == plugin.EdgeContains && edge.Resolution == "" {
			contained[edge.To] = true
		}
	}
	moduleKeys := make([]dirName, 0, len(moduleCanonFile))
	for k := range moduleCanonFile {
		moduleKeys = append(moduleKeys, k)
	}
	sort.Slice(moduleKeys, func(i, j int) bool {
		if moduleKeys[i].dir != moduleKeys[j].dir {
			return moduleKeys[i].dir < moduleKeys[j].dir
		}
		if moduleKeys[i].name != moduleKeys[j].name {
			return moduleKeys[i].name < moduleKeys[j].name
		}
		return moduleKeys[i].language < moduleKeys[j].language
	})

	for _, key := range moduleKeys {
		canonFile := moduleCanonFile[key]
		moduleGID, ok := localToGlobal[fileLocal{file: canonFile, name: key.name}]
		if !ok {
			continue
		}
		canonLanguage := languageFamily(canonFile)
		// Find all top-level symbols in same-language files sharing this directory.
		for _, fp := range files {
			if filepath.Dir(fp) != key.dir {
				continue
			}
			if languageFamily(fp) != canonLanguage {
				continue
			}
			for i := range fileSymbols[fp] {
				sym := &fileSymbols[fp][i]
				if sym.Category == plugin.CategoryModule || contained[sym.ID] || sym.IsFileMetadata() {
					continue
				}
				if sym.ScopedName != "" && sym.ScopedName != sym.Name {
					continue // missing lexical owners must not become package ownership
				}
				sg.Edges = append(sg.Edges, plugin.Edge{
					From: moduleGID,
					To:   sym.ID,
					Kind: plugin.EdgeContains,
				})
			}
		}
	}

	// Phase 3c: Resolve override targets through inheritance, then follow
	// scoped aliases for indirect calls.
	resolveOverrides(sg.Symbols, sg.Edges)
	sg.Edges = resolveDataflow(sg.Symbols, sg.Edges)
	sg.Edges = uniqueEdges(sg.Edges)

	// Phase 4: Build derived indexes for downstream consumers.
	sg.BuildIndexes()

	return sg
}

// fileLocal is a composite key for (filePath, localName) lookups.
type fileLocal struct {
	file string
	name string
}
