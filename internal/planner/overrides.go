// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"slices"
	"strings"

	"codeknit/internal/plugin"
)

type overrideIndex struct {
	members    map[string][]*plugin.Symbol
	parents    map[string][]string
	incomplete map[string]bool
}

// resolveOverrides connects overriding methods to the nearest inherited
// declarations. Unknown bases, cycles, and competing declarations retain
// uncertainty; an override can never point back to itself.
func resolveOverrides(symbols []plugin.Symbol, edges []plugin.Edge) {
	pending := false
	for i := range edges {
		if pendingOverride(&edges[i]) {
			pending = true
			break
		}
	}
	if !pending {
		return
	}
	byID := make(map[string]*plugin.Symbol, len(symbols))
	owners := make(map[string][]string)
	index := overrideIndex{
		members: make(map[string][]*plugin.Symbol), parents: make(map[string][]string),
		incomplete: make(map[string]bool),
	}
	for i := range symbols {
		byID[symbols[i].ID] = &symbols[i]
	}
	for i := range edges {
		e := &edges[i]
		switch e.Kind {
		case plugin.EdgeContains:
			member := byID[e.To]
			if e.Resolution == "" && member != nil && member.Category == plugin.CategoryCallable {
				if !slices.Contains(owners[e.To], e.From) {
					owners[e.To] = append(owners[e.To], e.From)
					index.members[e.From] = append(index.members[e.From], member)
				}
			}
		case plugin.EdgeInherits, plugin.EdgeImplements:
			if e.Resolution != "" || byID[e.To] == nil {
				index.incomplete[e.From] = true
			} else {
				index.parents[e.From] = append(index.parents[e.From], e.To)
			}
		}
	}
	for i := range edges {
		e := &edges[i]
		if !pendingOverride(e) {
			continue
		}
		method := byID[e.From]
		if method == nil || e.Resolution == "ambiguous" {
			continue
		}
		name := "super." + method.Name
		e.To, e.Resolution, e.Candidates = method.FilePath+"::"+name, "unresolved", nil
		if len(owners[e.From]) != 1 {
			continue
		}
		owner := owners[e.From][0]
		visiting := map[string]bool{owner: true}
		complete := !index.incomplete[owner]
		var candidates []string
		for _, parent := range index.parents[owner] {
			found, ok := index.targets(parent, method, visiting)
			candidates = append(candidates, found...)
			complete = complete && ok
		}
		e.To, e.Resolution, e.Candidates = chooseTarget(method.FilePath, name, candidates)
		if !complete && len(candidates) > 0 {
			slices.Sort(candidates)
			e.To, e.Resolution, e.Candidates = method.FilePath+"::"+name, "ambiguous", slices.Compact(candidates)
		}
	}
}

func pendingOverride(e *plugin.Edge) bool {
	return e.Kind == plugin.EdgeOverrides && e.Resolution == "unresolved" &&
		strings.HasPrefix(e.TargetName, "super.")
}

func (index *overrideIndex) targets(owner string, method *plugin.Symbol, visiting map[string]bool) ([]string, bool) {
	if visiting[owner] {
		return nil, false
	}
	visiting[owner] = true
	defer delete(visiting, owner)
	var candidates []string
	for _, member := range index.members[owner] {
		if member.ID == method.ID || member.Name != method.Name || member.Properties["visibility"] == "private" {
			continue
		}
		if (member.Properties["static"] == "true") != (method.Properties["static"] == "true") {
			continue
		}
		if (method.Kind == "getter" || method.Kind == "setter") && member.Kind != method.Kind {
			continue
		}
		candidates = append(candidates, member.ID)
	}
	if len(candidates) > 0 {
		return candidates, true
	}
	complete := !index.incomplete[owner]
	for _, parent := range index.parents[owner] {
		found, ok := index.targets(parent, method, visiting)
		candidates = append(candidates, found...)
		complete = complete && ok
	}
	return candidates, complete
}
