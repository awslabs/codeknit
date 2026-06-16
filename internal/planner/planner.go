// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package planner builds a SymbolGraph from per-file extraction results.
package planner

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// Planner builds a SymbolGraph from per-file Symbols and Edges.
type Planner struct{}

// Plan takes per-file Symbols and Edges and produces a unified SymbolGraph.
// Symbol IDs are assigned using the format "filePath::SymbolName".
// Intra-file Edge references are remapped to global IDs.
// Cross-file references are resolved using import context when available,
// falling back to directory proximity for unqualified names.
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
	type dirName struct{ dir, name string }
	moduleCanonFile := make(map[dirName]string) // (dir,name) → first file that declared it

	for _, fp := range files {
		symbols := fileSymbols[fp]
		dir := filepath.Dir(fp)
		filtered := symbols[:0:0]
		for i := range symbols {
			sym := &symbols[i]
			if sym.Category != plugin.CategoryModule {
				filtered = append(filtered, *sym)
				continue
			}
			key := dirName{dir: dir, name: sym.Name}
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
	globalByName := make(map[string][]string)
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
			globalByName[scopedName] = append(globalByName[scopedName], globalID)

			// Also index by unscoped Name so that unqualified references
			// (e.g., "Dispatch" for "DispatchTable.Dispatch") can be resolved.
			if sym.Name != scopedName {
				globalByName[sym.Name] = append(globalByName[sym.Name], globalID)
			}

			sg.Symbols = append(sg.Symbols, *sym)
		}
	}

	// Phase 2: Build per-file import maps from EdgeImports edges.
	// importMap[filePath][localName] = modulePath
	// This tells us: "in file F, the name N was imported from module M."
	importMap := buildImportMap(files, fileEdges)

	// Phase 3: Remap edges and resolve cross-file references.
	for _, fp := range files {
		edges := fileEdges[fp]
		for _, e := range edges {
			// Skip import edges — they're metadata consumed by the planner,
			// not structural relationships to include in the output graph.
			if e.Kind == plugin.EdgeImports {
				continue
			}

			remapped := plugin.Edge{Kind: e.Kind}
			fileImports := importMap[fp]
			remapped.From = resolveID(fp, e.From, localToGlobal, globalByName, fileImports)
			remapped.To = resolveID(fp, e.To, localToGlobal, globalByName, fileImports)

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
	moduleKeys := make([]dirName, 0, len(moduleCanonFile))
	for k := range moduleCanonFile {
		moduleKeys = append(moduleKeys, k)
	}
	sort.Slice(moduleKeys, func(i, j int) bool {
		if moduleKeys[i].dir != moduleKeys[j].dir {
			return moduleKeys[i].dir < moduleKeys[j].dir
		}
		return moduleKeys[i].name < moduleKeys[j].name
	})

	for _, key := range moduleKeys {
		canonFile := moduleCanonFile[key]
		moduleGID, ok := localToGlobal[fileLocal{file: canonFile, name: key.name}]
		if !ok {
			continue
		}
		canonExt := filepath.Ext(canonFile)
		// Find all top-level symbols in same-language files sharing this directory.
		for _, fp := range files {
			if filepath.Dir(fp) != key.dir {
				continue
			}
			if filepath.Ext(fp) != canonExt {
				continue
			}
			for i := range fileSymbols[fp] {
				sym := &fileSymbols[fp][i]
				if sym.Category == plugin.CategoryModule {
					continue
				}
				sg.Edges = append(sg.Edges, plugin.Edge{
					From: moduleGID,
					To:   sym.ID,
					Kind: plugin.EdgeContains,
				})
			}
		}
	}

	// Phase 3c: Dataflow resolution — resolve indirect calls through aliases
	// and return values using fixed-point iteration.
	sg.Edges = resolveDataflow(sg.Symbols, sg.Edges, localToGlobal, globalByName)

	// Phase 4: Build derived indexes for downstream consumers.
	sg.BuildIndexes()

	return sg
}

// fileLocal is a composite key for (filePath, localName) lookups.
type fileLocal struct {
	file string
	name string
}

// buildImportMap extracts import information from EdgeImports edges.
// Returns a map: filePath → (localName → modulePath).
//
// Plugins emit EdgeImports edges where:
//   - From = the local name being imported (e.g., "ArrayList", "os", "User")
//   - To   = the module/package path (e.g., "java.util", "os", "models")
//
// The planner uses this to prefer candidates from imported modules when
// resolving cross-file references.
func buildImportMap(files []string, fileEdges map[string][]plugin.Edge) map[string]map[string]string {
	importMap := make(map[string]map[string]string)
	for _, fp := range files {
		for _, e := range fileEdges[fp] {
			if e.Kind != plugin.EdgeImports {
				continue
			}
			if importMap[fp] == nil {
				importMap[fp] = make(map[string]string)
			}
			importMap[fp][e.From] = e.To
		}
	}
	return importMap
}

// resolveID resolves a local symbol name to a global ID.
// Resolution order (first match wins):
//  1. Same-file: exact (filePath, localName) match in localToGlobal.
//  2. Import-aware: if the file has an import for this name, prefer
//     candidates whose file path contains the imported module path.
//  3. Single cross-file candidate: used directly.
//  4. Multiple cross-file candidates: pick the one with the smallest
//     PathDistance to the referencing file. Ties are broken by choosing
//     the lexicographically smallest global ID, guaranteeing a
//     deterministic and unique result for any given input.
//  5. No candidates: return "filePath::localName" as an unresolved marker.
func resolveID(
	filePath, localName string,
	localToGlobal map[fileLocal]string,
	globalByName map[string][]string,
	fileImports map[string]string,
) string {
	// 1. Same-file resolution.
	if gid, ok := localToGlobal[fileLocal{file: filePath, name: localName}]; ok {
		return gid
	}

	// 2–4. Cross-file resolution.
	candidates := globalByName[localName]
	if len(candidates) == 0 {
		// 5. Unresolved — return a synthetic ID so the edge is still recorded.
		return filePath + "::" + localName
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	// 2. Import-aware resolution: if we have import context for this name,
	// filter candidates to those whose file path matches the import source.
	if modulePath, ok := fileImports[localName]; ok && modulePath != "" {
		imported := filterByImport(candidates, modulePath, filePath)
		if len(imported) == 1 {
			return imported[0]
		}
		if len(imported) > 1 {
			// Multiple matches within the imported module — pick by proximity.
			candidates = imported
		}
		// If no import matches, fall through to proximity-based resolution.
	}

	// 3–4. Proximity-based resolution with deterministic tiebreaking.
	sourceDir := filepath.Dir(filePath)
	best := candidates[0]
	bestDist := ir.PathDistance(sourceDir, candidateDir(best))
	for _, gid := range candidates[1:] {
		d := ir.PathDistance(sourceDir, candidateDir(gid))
		if d < bestDist || (d == bestDist && gid < best) {
			bestDist = d
			best = gid
		}
	}
	return best
}

// filterByImport returns candidates whose file path matches the module path.
//
// Resolution strategy (first match wins):
//
//  1. Relative resolution: if the module path looks relative (starts with
//     "./", "../", or Python-style leading dots), resolve it against the
//     importing file's directory and match candidates whose path starts
//     with the resolved prefix.
//
//  2. Directory-scoped substring: try matching the normalized module path
//     only against candidates that share the importing file's directory
//     tree. This prevents `from models import X` in folder1/ from matching
//     folder2/models/.
//
//  3. Global substring fallback: match any candidate whose path contains
//     the normalized module path. This handles fully-qualified imports
//     (Java packages, Go module paths) where directory locality doesn't
//     apply.
func filterByImport(candidates []string, modulePath, importingFile string) []string {
	importDir := filepath.Dir(importingFile)

	// Step 1: Detect and resolve relative imports.
	if resolved, ok := resolveRelativeImport(modulePath, importDir); ok {
		var matched []string
		for _, gid := range candidates {
			candidatePath := filepath.ToSlash(candidateFilePath(gid))
			if strings.Contains(candidatePath, resolved) {
				matched = append(matched, gid)
			}
		}
		if len(matched) > 0 {
			return matched
		}
		// Relative resolution found no matches — fall through.
	}

	// Normalize non-relative module paths: convert dots/colons to slashes.
	// Java: com.example.models → com/example/models
	// Rust: crate::models → crate/models
	normalized := strings.NewReplacer(".", "/", "::", "/").Replace(modulePath)

	// Step 2: Prefer candidates under the same directory tree as the importer.
	// This handles bare imports like Python's `from models import X` where
	// folder1/app.py should prefer folder1/models/ over folder2/models/.
	importDirSlash := filepath.ToSlash(importDir)
	var localMatched []string
	for _, gid := range candidates {
		candidatePath := filepath.ToSlash(candidateFilePath(gid))
		if strings.Contains(candidatePath, normalized) && sharePathPrefix(importDirSlash, candidatePath) {
			localMatched = append(localMatched, gid)
		}
	}
	if len(localMatched) > 0 {
		return localMatched
	}

	// Step 3: Global substring fallback for fully-qualified imports.
	var matched []string
	for _, gid := range candidates {
		candidatePath := filepath.ToSlash(candidateFilePath(gid))
		if strings.Contains(candidatePath, normalized) {
			matched = append(matched, gid)
		}
	}
	return matched
}

// resolveRelativeImport detects relative import paths and resolves them
// against the importing file's directory. Returns the resolved path and
// true if the import was relative, or ("", false) otherwise.
//
// Supported patterns:
//   - JS/TS:  "./foo", "../foo"
//   - Python: ".foo" (current package), "..foo" (parent), "...foo" etc.
func resolveRelativeImport(modulePath, importDir string) (string, bool) {
	// JS/TS style: ./foo or ../foo
	if strings.HasPrefix(modulePath, "./") || strings.HasPrefix(modulePath, "../") {
		resolved := filepath.Join(importDir, modulePath)
		return filepath.ToSlash(filepath.Clean(resolved)), true
	}

	// Python style: leading dots without slash (.models, ..models, ...)
	// Count leading dots and strip them.
	if modulePath != "" && modulePath[0] == '.' {
		dots := 0
		for dots < len(modulePath) && modulePath[dots] == '.' {
			dots++
		}
		remainder := modulePath[dots:]

		// Each dot means "go up one level" (first dot = current dir).
		dir := importDir
		for i := 1; i < dots; i++ {
			dir = filepath.Dir(dir)
		}

		// Convert the remainder from dotted to slashed (Python: models.user → models/user).
		if remainder != "" {
			remainder = strings.ReplaceAll(remainder, ".", "/")
			dir = filepath.Join(dir, remainder)
		}
		return filepath.ToSlash(filepath.Clean(dir)), true
	}

	return "", false
}

// sharePathPrefix returns true if two slash-normalized paths share a common
// non-trivial directory prefix. This detects whether two files are in the
// same project subtree (e.g., both under "src/" or both under "folder1/").
func sharePathPrefix(dirA, pathB string) bool {
	dirB := pathB
	if idx := strings.LastIndex(pathB, "/"); idx >= 0 {
		dirB = pathB[:idx]
	}

	// Find the first path component of each.
	rootA := dirA
	if idx := strings.Index(dirA, "/"); idx >= 0 {
		rootA = dirA[:idx]
	}
	rootB := dirB
	if idx := strings.Index(dirB, "/"); idx >= 0 {
		rootB = dirB[:idx]
	}

	// Both must share the same top-level directory.
	return rootA != "" && rootA == rootB
}

// candidateFilePath extracts the file path from a global ID ("filePath::Name").
func candidateFilePath(globalID string) string {
	if before, _, ok := strings.Cut(globalID, "::"); ok {
		return before
	}
	return globalID
}

// candidateDir extracts the directory from a global ID ("filePath::Name").
func candidateDir(globalID string) string {
	return filepath.Dir(candidateFilePath(globalID))
}
