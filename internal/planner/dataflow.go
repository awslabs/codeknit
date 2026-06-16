// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"codeknit/internal/plugin"
)

// maxDataflowIterations caps the fixed-point iteration to prevent pathological cases.
const maxDataflowIterations = 10

// resolveDataflow performs inter-procedural dataflow analysis on the symbol graph.
// It consumes EdgeAliases and EdgeReturns metadata edges to discover additional
// call edges that the AST-level extraction missed, then strips the metadata edges.
//
// The algorithm:
//  1. Build alias chains: follow EdgeAliases transitively (x=y, y=z → x resolves to z)
//  2. Resolve indirect calls: for EdgeCalls where the target is a non-callable
//     (variable/property), follow alias chains to find the actual callable
//  3. Follow returns: when a function's return value is used, follow EdgeReturns
//     to determine what it returns
//  4. Fixed-point iteration: repeat until no new edges are discovered
func resolveDataflow(
	symbols []plugin.Symbol,
	edges []plugin.Edge,
	_ map[fileLocal]string,
	globalByName map[string][]string,
) []plugin.Edge {
	// Index: which global IDs are callables?
	callableIDs := make(map[string]bool, len(symbols))
	for i := range symbols {
		if symbols[i].Category == plugin.CategoryCallable {
			callableIDs[symbols[i].ID] = true
		}
	}

	// Separate metadata edges from structural edges.
	var structural []plugin.Edge
	aliasEdges := make(map[string][]string)   // from (global ID) → [to] (alias targets)
	aliasByLocal := make(map[string][]string) // from (local name) → [to] (alias targets)
	returnEdges := make(map[string][]string)  // funcID → [returned symbol IDs]

	for _, e := range edges {
		switch e.Kind {
		case plugin.EdgeAliases:
			aliasEdges[e.From] = append(aliasEdges[e.From], e.To)
			localName := extractLocalName(e.From)
			if localName != "" {
				aliasByLocal[localName] = append(aliasByLocal[localName], e.To)
			}
		case plugin.EdgeReturns:
			returnEdges[e.From] = append(returnEdges[e.From], e.To)
		default:
			structural = append(structural, e)
		}
	}

	// If no dataflow hints, return edges unchanged (minus metadata).
	if len(aliasEdges) == 0 && len(returnEdges) == 0 {
		return structural
	}

	// Step 1: Build transitive alias resolution with cycle detection.
	// Looks up by global ID first, then falls back to local name.
	resolveCache := make(map[string]string)
	var resolveAlias func(name string, visited map[string]bool) string
	resolveAlias = func(name string, visited map[string]bool) string {
		if cached, ok := resolveCache[name]; ok {
			return cached
		}
		if visited[name] {
			return name // cycle
		}
		visited[name] = true

		// Try global ID first, then local name.
		targets := aliasEdges[name]
		if len(targets) == 0 {
			targets = aliasByLocal[name]
		}
		if len(targets) == 0 {
			resolveCache[name] = name
			return name
		}
		// Follow the first alias (most assignments are single-target).
		// Extract local name from the target for recursive resolution.
		targetLocal := extractLocalName(targets[0])
		if targetLocal == "" {
			targetLocal = targets[0]
		}
		resolved := resolveAlias(targetLocal, visited)
		resolveCache[name] = resolved
		return resolved
	}

	// Step 2-4: Fixed-point iteration.
	// Track existing call edges to avoid duplicates.
	existingCalls := make(map[[2]string]bool)
	for _, e := range structural {
		if e.Kind == plugin.EdgeCalls {
			existingCalls[[2]string{e.From, e.To}] = true
		}
	}

	for iter := 0; iter < maxDataflowIterations; iter++ {
		newEdges := 0

		// Scan all call edges. For each one where the target is not a callable,
		// try to resolve it through aliases.
		var additions []plugin.Edge
		for _, e := range structural {
			if e.Kind != plugin.EdgeCalls {
				continue
			}
			// If target is already a callable, nothing to resolve.
			if callableIDs[e.To] {
				continue
			}

			// Try alias resolution: the target name might be an alias for a callable.
			// Extract the local name from the global ID for alias lookup.
			targetLocal := extractLocalName(e.To)
			if targetLocal == "" {
				continue
			}

			visited := make(map[string]bool)
			resolved := resolveAlias(targetLocal, visited)
			if resolved == targetLocal {
				continue // no alias found
			}

			// Find the global ID for the resolved name.
			candidates := globalByName[resolved]
			for _, candidate := range candidates {
				if callableIDs[candidate] {
					key := [2]string{e.From, candidate}
					if !existingCalls[key] {
						existingCalls[key] = true
						additions = append(additions, plugin.Edge{
							From: e.From,
							To:   candidate,
							Kind: plugin.EdgeCalls,
						})
						newEdges++
					}
				}
			}
		}

		// Also resolve return edges: if function F returns callable G,
		// and someone calls F and then uses the result, we can infer
		// that the caller transitively depends on G.
		for funcID, returnTargets := range returnEdges {
			// Find all callers of funcID.
			for _, e := range structural {
				if e.Kind != plugin.EdgeCalls || e.To != funcID {
					continue
				}
				caller := e.From
				for _, retTarget := range returnTargets {
					// Resolve the return target through aliases.
					visited := make(map[string]bool)
					resolved := resolveAlias(retTarget, visited)
					candidates := globalByName[resolved]
					for _, candidate := range candidates {
						if callableIDs[candidate] {
							key := [2]string{caller, candidate}
							if !existingCalls[key] {
								existingCalls[key] = true
								additions = append(additions, plugin.Edge{
									From: caller,
									To:   candidate,
									Kind: plugin.EdgeCalls,
								})
								newEdges++
							}
						}
					}
				}
			}
		}

		structural = append(structural, additions...)

		if newEdges == 0 {
			break // fixed point reached
		}
	}

	return structural
}

// extractLocalName extracts the local symbol name from a global ID.
// Global IDs have the format "filePath::ScopedName" or "filePath::ScopedName#N".
func extractLocalName(globalID string) string {
	for i := len(globalID) - 1; i >= 1; i-- {
		if globalID[i] == ':' && globalID[i-1] == ':' {
			name := globalID[i+1:]
			// Strip disambiguation suffix (#N).
			for j := range name {
				if name[j] == '#' {
					return name[:j]
				}
			}
			return name
		}
	}
	return ""
}
