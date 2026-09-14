// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"slices"
	"sort"

	"codeknit/internal/plugin"
)

// detectDeepInheritance finds the longest chain through all bases/interfaces.
// Cyclic hierarchies are reported by cycle analysis, not assigned a fake depth.
func detectDeepInheritance(parents map[string][]string, maxDepth int) []inheritanceChain {
	memo := make(map[string][]string)
	state := make(map[string]int)
	var chain func(string) []string
	chain = func(sid string) []string {
		if state[sid] == 1 {
			return nil
		}
		if state[sid] == 2 {
			return memo[sid]
		}
		state[sid] = 1
		var longest []string
		invalid := false
		for _, parent := range parents[sid] {
			path := chain(parent)
			if path == nil {
				invalid = true
			}
			if len(path) > len(longest) || len(path) == len(longest) && slices.Compare(path, longest) < 0 {
				longest = path
			}
		}
		state[sid] = 2
		if invalid {
			memo[sid] = nil
			return nil
		}
		memo[sid] = append(slices.Clone(longest), sid)
		return memo[sid]
	}
	var result []inheritanceChain
	for sid := range parents {
		path := chain(sid)
		if len(path)-1 >= maxDepth {
			result = append(result, inheritanceChain{Leaf: sid, Depth: len(path) - 1, Chain: path})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth > result[j].Depth
		}
		return result[i].Leaf < result[j].Leaf
	})
	return result
}

// A filename is not proof that every declaration in it is an entry point.
func isEntryPoint(sid string, g *graphCtx) bool {
	sym := g.sidInfo[sid]
	return sym != nil && sym.Category == plugin.CategoryCallable &&
		(sym.Name == "main" || sym.Name == "Main" || sym.Name == "init")
}

func findUnreachable(g *graphCtx) []symbolEntry {
	var queue []string
	reachable := make(map[string]bool)
	for _, sid := range g.allSIDs {
		if isEntryPoint(sid, g) {
			queue = append(queue, sid)
			reachable[sid] = true
		}
	}
	if len(queue) == 0 {
		return nil // library/framework roots cannot be inferred from filenames
	}
	for head := 0; head < len(queue); head++ {
		for _, next := range g.adj[queue[head]] {
			if !reachable[next] {
				reachable[next] = true
				queue = append(queue, next)
			}
		}
	}
	// Owners of reachable members are relevant, but this does not imply that
	// their other members execute. Keep ancestry separate from dependency BFS.
	var ancestors func(string)
	ancestors = func(sid string) {
		for _, parent := range g.containsRev[sid] {
			if !reachable[parent] {
				reachable[parent] = true
				ancestors(parent)
			}
		}
	}
	for _, sid := range queue {
		ancestors(sid)
	}
	var result []symbolEntry
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		if !reachable[sid] && sym.Category != plugin.CategoryModule {
			result = append(result, symbolEntry{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				Category:  string(sym.Category),
			})
		}
	}
	return result
}
