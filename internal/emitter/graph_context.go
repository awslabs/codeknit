// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"path/filepath"
	"slices"
	"sort"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// --- shared graph context ---

// graphCtx holds pre-computed adjacency lists and lookups shared by all algorithms.
type graphCtx struct {
	sg              *ir.SymbolGraph
	adj             map[string][]string       // outgoing (non-contains)
	radj            map[string][]string       // incoming (non-contains)
	containsAdj     map[string][]string       // parent → children (contains edges only)
	containsRev     map[string][]string       // child → parent (contains edges only)
	inheritEdges    map[string][]string       // child → parent (inherits/implements)
	sidInfo         map[string]*plugin.Symbol // SID → symbol
	sidToPkg        map[string]string         // SID → package directory
	fanOut          map[string]int
	fanIn           map[string]int
	containsCount   map[string]int
	pkgEdges        map[[2]string]int // [fromPkg, toPkg] → edge count
	allSIDs         []string          // sorted list of all SIDs
	resolvedEdges   int
	unresolvedEdges int
	ambiguousEdges  int
}

func buildGraphCtx(sg *ir.SymbolGraph) *graphCtx {
	g := &graphCtx{
		sg:            sg,
		adj:           make(map[string][]string),
		radj:          make(map[string][]string),
		fanOut:        make(map[string]int),
		fanIn:         make(map[string]int),
		containsCount: make(map[string]int),
		containsAdj:   make(map[string][]string),
		containsRev:   make(map[string][]string),
		inheritEdges:  make(map[string][]string),
		sidInfo:       make(map[string]*plugin.Symbol, len(sg.Symbols)),
		sidToPkg:      make(map[string]string, len(sg.Symbols)),
		pkgEdges:      make(map[[2]string]int),
	}

	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		// Include descriptors and their file anchors are not declarations.
		// They must not appear as dead code or affect symbol-based rankings.
		if sym.IsFileMetadata() {
			continue
		}
		if sid, ok := sg.ShortIDs[sym.ID]; ok {
			g.sidInfo[sid] = sym
			g.sidToPkg[sid] = filepath.Dir(sym.FilePath)
		}
	}

	g.allSIDs = make([]string, 0, len(g.sidInfo))
	for sid := range g.sidInfo {
		g.allSIDs = append(g.allSIDs, sid)
	}
	sort.Strings(g.allSIDs)

	// A dependency graph has one neighbor per symbol pair. Repeated edges or
	// different relationship kinds must not inflate fan-in, PageRank or paths.
	seenEdges := make(map[[3]string]bool)
	addNeighbor := func(adj map[string][]string, from, to string) bool {
		if slices.Contains(adj[from], to) {
			return false
		}
		adj[from] = append(adj[from], to)
		return true
	}
	for i := range sg.Edges {
		edge := &sg.Edges[i]
		key := [3]string{edge.From, edge.To, string(edge.Kind)}
		resolution := sg.EdgeResolution(edge)
		if resolution != "" {
			if resolution == "ambiguous" {
				g.ambiguousEdges++
			} else {
				g.unresolvedEdges++
			}
			continue
		}
		if seenEdges[key] {
			continue
		}
		fromSID, toSID := sg.ShortIDs[edge.From], sg.ShortIDs[edge.To]
		if g.sidInfo[fromSID] == nil || g.sidInfo[toSID] == nil {
			continue
		}
		seenEdges[key] = true
		g.resolvedEdges++
		if edge.Kind == plugin.EdgeContains {
			if addNeighbor(g.containsAdj, fromSID, toSID) {
				g.containsCount[fromSID]++
			}
			addNeighbor(g.containsRev, toSID, fromSID)
			continue
		}
		if edge.Kind == plugin.EdgeInherits || edge.Kind == plugin.EdgeImplements {
			addNeighbor(g.inheritEdges, fromSID, toSID)
		}
		if addNeighbor(g.adj, fromSID, toSID) {
			addNeighbor(g.radj, toSID, fromSID)
			g.fanOut[fromSID]++
			g.fanIn[toSID]++
			fromPkg, toPkg := g.sidToPkg[fromSID], g.sidToPkg[toSID]
			if fromPkg != toPkg {
				g.pkgEdges[[2]string{fromPkg, toPkg}]++
			}
		}
	}
	for _, adj := range []map[string][]string{g.adj, g.radj, g.containsAdj, g.containsRev, g.inheritEdges} {
		for _, neighbors := range adj {
			sort.Strings(neighbors)
		}
	}

	return g
}
