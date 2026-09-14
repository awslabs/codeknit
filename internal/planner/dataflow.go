// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"slices"

	"codeknit/internal/plugin"
)

// resolveDataflow follows aliases by their full file/scope identity. Multiple
// assignments remain ambiguous; returning a callable does not invoke it.
func resolveDataflow(symbols []plugin.Symbol, edges []plugin.Edge) []plugin.Edge {
	callable := make(map[string]bool)
	for i := range symbols {
		s := &symbols[i]
		callable[s.ID] = s.Category == plugin.CategoryCallable
	}
	aliases := make(map[string][]*plugin.Edge)
	for i := range edges {
		e := &edges[i]
		if e.Kind == plugin.EdgeAliases {
			aliases[e.From] = append(aliases[e.From], e)
		}
	}
	var follow func(string, map[string]bool) ([]string, bool)
	follow = func(id string, visiting map[string]bool) ([]string, bool) {
		if visiting[id] {
			return nil, false
		}
		if len(aliases[id]) == 0 {
			if callable[id] {
				return []string{id}, true
			}
			return nil, false
		}
		visiting[id] = true
		defer delete(visiting, id)
		var targets []string
		complete := true
		for _, alias := range aliases[id] {
			if alias.Resolution == "ambiguous" {
				targets = append(targets, alias.Candidates...)
				complete = false
				continue
			}
			ids, ok := follow(alias.To, visiting)
			complete = complete && ok
			targets = append(targets, ids...)
		}
		slices.Sort(targets)
		return slices.Compact(targets), complete
	}
	result := make([]plugin.Edge, 0, len(edges))
	for i := range edges {
		e := edges[i] // transform a copy; alias lookups retain the original edges
		switch e.Kind {
		case plugin.EdgeAliases:
			continue
		case plugin.EdgeReturns:
			// This is a dependency on a returned value, not proof of its invocation.
			e.Kind = plugin.EdgeReferences
		case plugin.EdgeCalls:
			if len(aliases[e.To]) > 0 {
				targets, complete := follow(e.To, make(map[string]bool))
				switch {
				case len(targets) == 1 && complete:
					e.To, e.Resolution, e.Candidates = targets[0], "", nil
				case len(targets) > 0:
					e.Resolution, e.Candidates = "ambiguous", targets
				default:
					e.Resolution = "unresolved"
				}
			}
		}
		result = append(result, e)
	}
	return result
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
