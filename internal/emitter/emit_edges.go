// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"codeknit/internal/ir"
)

// Simple unresolved names are relative to the source file. Other identities
// are quoted in full so spaces, commas, and syntax cannot alter an edge record.
var (
	localEdgeName = regexp.MustCompile(`^[\pL_$][\pL\pN_$#.]*$`)
	shortEdgeID   = regexp.MustCompile(`^S\d+$`)
)

func edgeIDToken(id string, shortIDs map[string]string) string {
	if sid := shortIDs[id]; sid != "" {
		return sid
	}
	return strconv.Quote(id)
}

func uncertainTargetToken(id, sourceFile string) string {
	if name, ok := strings.CutPrefix(id, sourceFile+"::"); ok && localEdgeName.MatchString(name) && !shortEdgeID.MatchString(name) {
		return name
	}
	return strconv.Quote(id)
}

// buildEdgeLines uses one SKT grammar for every relationship. Grouping includes
// resolution and candidate identities so uncertain targets never merge with
// resolved edges or with a different ambiguity.
func buildEdgeLines(sg *ir.SymbolGraph, dict *Dictionary) map[string][]string {
	type groupKey struct {
		from, kind, resolution, candidates string
	}
	targets := make(map[groupKey][]string)
	order := make(map[string][]groupKey)
	for i := range sg.Edges {
		edge := &sg.Edges[i]
		fp := sg.SymbolFile[edge.From]
		if fp == "" {
			fp, _, _ = strings.Cut(edge.From, "::")
		}
		key := groupKey{
			from: edgeIDToken(edge.From, sg.ShortIDs), kind: string(edge.Kind),
			resolution: sg.EdgeResolution(edge),
		}
		candidates := make([]string, 0, len(edge.Candidates))
		for _, candidate := range edge.Candidates {
			candidates = append(candidates, edgeIDToken(candidate, sg.ShortIDs))
		}
		key.candidates = strings.Join(candidates, ", ")
		target := edgeIDToken(edge.To, sg.ShortIDs)
		// An ambiguous name can coincide with the first overload's ID.
		// Keep that name instead of presenting one candidate as the target.
		if key.resolution != "" && (sg.ShortIDs[edge.To] == "" ||
			key.resolution == "ambiguous" && slices.Contains(edge.Candidates, edge.To)) {
			target = uncertainTargetToken(edge.To, fp)
		}
		if _, exists := targets[key]; !exists {
			order[fp] = append(order[fp], key)
		}
		targets[key] = append(targets[key], target)
	}

	lines := make(map[string][]string, len(order))
	for fp, groups := range order {
		for _, key := range groups {
			kind := key.kind
			if dict != nil {
				kind = dict.Encode(kind)
			}
			if key.resolution != "" {
				kind += "[" + key.resolution + "]"
			}
			line := fmt.Sprintf("%s --%s--> %s", key.from, kind, strings.Join(targets[key], ", "))
			if key.candidates != "" {
				line += " [candidates=" + key.candidates + "]"
			}
			lines[fp] = append(lines[fp], line)
		}
	}
	return lines
}
