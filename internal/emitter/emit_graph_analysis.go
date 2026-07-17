// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeknit/internal/config"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// AnalysisOptions controls graph analysis behavior.
type AnalysisOptions struct {
	OutputPath string
	// FanThreshold is the minimum fan-in or fan-out to flag a hub symbol.
	FanThreshold int
	// GodThreshold is the minimum contains-edge count to flag a god class.
	GodThreshold int
	// MaxInheritanceDepth flags inheritance chains deeper than this.
	MaxInheritanceDepth int
	// TopN caps ranked output sections (betweenness, pagerank, transitive fan-in,
	// change propagation, dependency weight). 0 means no limit.
	TopN int
	// BetweennessThreshold is the minimum betweenness centrality value to report.
	BetweennessThreshold float64
	// PropagationCutoff is the minimum probability to continue change propagation.
	PropagationCutoff float64
}

// DefaultAnalysisOptions returns sensible defaults.
func DefaultAnalysisOptions() *AnalysisOptions {
	return &AnalysisOptions{
		OutputPath:           config.DefaultAnalyzeOutput,
		FanThreshold:         config.DefaultAnalyzeFanThreshold,
		GodThreshold:         config.DefaultAnalyzeGodThreshold,
		MaxInheritanceDepth:  config.DefaultAnalyzeMaxInheritanceDepth,
		TopN:                 config.DefaultAnalyzeTopN,
		BetweennessThreshold: config.DefaultAnalyzeBetweennessThreshold,
		PropagationCutoff:    config.DefaultAnalyzePropagationCutoff,
	}
}

// --- shared graph context ---

// graphCtx holds pre-computed adjacency lists and lookups shared by all algorithms.
type graphCtx struct {
	sg            *ir.SymbolGraph
	adj           map[string][]string       // outgoing (non-contains)
	radj          map[string][]string       // incoming (non-contains)
	containsAdj   map[string][]string       // parent → children (contains edges only)
	containsRev   map[string]string         // child → parent (contains edges only)
	inheritEdges  map[string]string         // child → parent (inherits/implements)
	sidInfo       map[string]*plugin.Symbol // SID → symbol
	sidToPkg      map[string]string         // SID → package directory
	fanOut        map[string]int
	fanIn         map[string]int
	containsCount map[string]int
	pkgEdges      map[[2]string]int // [fromPkg, toPkg] → edge count
	allSIDs       []string          // sorted list of all SIDs
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
		containsRev:   make(map[string]string),
		inheritEdges:  make(map[string]string),
		sidInfo:       make(map[string]*plugin.Symbol, len(sg.Symbols)),
		sidToPkg:      make(map[string]string, len(sg.Symbols)),
		pkgEdges:      make(map[[2]string]int),
	}

	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
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

	for _, edge := range sg.Edges {
		fromSID, fromOK := sg.ShortIDs[edge.From]
		toSID, toOK := sg.ShortIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}

		switch edge.Kind {
		case plugin.EdgeContains:
			g.containsCount[fromSID]++
			g.containsAdj[fromSID] = append(g.containsAdj[fromSID], toSID)
			g.containsRev[toSID] = fromSID
		case plugin.EdgeInherits, plugin.EdgeImplements:
			g.inheritEdges[fromSID] = toSID
			g.adj[fromSID] = append(g.adj[fromSID], toSID)
			g.radj[toSID] = append(g.radj[toSID], fromSID)
			g.fanOut[fromSID]++
			g.fanIn[toSID]++
		default:
			g.adj[fromSID] = append(g.adj[fromSID], toSID)
			g.radj[toSID] = append(g.radj[toSID], fromSID)
			g.fanOut[fromSID]++
			g.fanIn[toSID]++
		}

		// Package-level edge tracking (non-contains only).
		if edge.Kind != plugin.EdgeContains {
			fromPkg := g.sidToPkg[fromSID]
			toPkg := g.sidToPkg[toSID]
			if fromPkg != "" && toPkg != "" && fromPkg != toPkg {
				g.pkgEdges[[2]string{fromPkg, toPkg}]++
			}
		}
	}

	return g
}

// --- analysis result types ---

type cycle struct {
	Members []string
}

type hubSymbol struct {
	baseEntry
	FanIn  int
	FanOut int
}

// baseEntry is the common SID/Name/File triple shared by all ranked findings.
type baseEntry struct {
	SID  string
	Name string
	File string
}

// symbolEntry is a generic symbol reference used for orphan and unreachable findings.
type symbolEntry struct {
	baseEntry
	Category string
}

type godSymbol struct {
	baseEntry
	Children int
}

type instabilityEntry struct {
	baseEntry
	Ca          int
	Ce          int
	Instability float64
}

type inheritanceChain struct {
	Leaf  string
	Chain []string
	Depth int
}

type betweennessEntry struct {
	baseEntry
	Betweenness float64
}

type articulationEntry struct {
	baseEntry
	MinSplit int // size of the smaller component if this node is removed
	MaxSplit int // size of the larger component if this node is removed
}

type pageRankEntry struct {
	baseEntry
	Rank float64
}

type transitiveFanInEntry struct {
	baseEntry
	DirectFanIn     int
	TransitiveFanIn int
}

type changePropEntry struct {
	baseEntry
	BlastRadius int     // number of symbols reachable via weighted propagation
	Probability float64 // average propagation probability
}

type layerViolation struct {
	FromSID   string
	ToSID     string
	FromLayer string
	ToLayer   string
	FromFile  string
	EdgeKind  string
}

type connectedComponent struct {
	Members []string // SIDs
	Sample  []string // first few names for display
	Size    int
}

type depWeightEntry struct {
	FromPkg string
	ToPkg   string
	Weight  int // number of distinct symbol-level edges
}

type distanceEntry struct {
	Pkg          string
	Instability  float64
	Abstractness float64
	Distance     float64 // |A + I - 1|
}

type shotgunSurgeryGroup struct {
	Pkg     string   // package of the targets
	Callers []string // SIDs of the shared callers
	Targets []string // SIDs of the co-called symbols
	Names   []string // names of the targets for display
}

type featureEnvyEntry struct {
	baseEntry
	OwnPkg      string
	EnviedPkg   string
	OwnRefs     int // references to own package
	ForeignRefs int // references to the envied package
}

type stableDepViolation struct {
	FromPkg    string
	ToPkg      string
	FromInstab float64
	ToInstab   float64
}

type interfaceSegregationEntry struct {
	baseEntry
	Members int // total members in the interface
	MaxUsed int // max members used by any single implementor
	Impls   int // number of implementors
}

type containmentDepthEntry struct {
	baseEntry
	Chain []string // names from root to leaf
	Depth int
}

// analysisResult holds all findings.
type analysisResult struct {
	Cycles               []cycle
	Hubs                 []hubSymbol
	Orphans              []symbolEntry
	GodSymbols           []godSymbol
	Instability          []instabilityEntry
	DeepInherit          []inheritanceChain
	Betweenness          []betweennessEntry
	ArticulationPoints   []articulationEntry
	PageRanks            []pageRankEntry
	TransitiveFanIn      []transitiveFanInEntry
	ChangePropagation    []changePropEntry
	PkgCycles            []cycle
	LayerViolations      []layerViolation
	Unreachable          []symbolEntry
	WeakComponents       []connectedComponent
	DependencyWeights    []depWeightEntry
	DistanceFromMain     []distanceEntry
	ShotgunSurgery       []shotgunSurgeryGroup
	FeatureEnvy          []featureEnvyEntry
	StableDepViolations  []stableDepViolation
	InterfaceSegregation []interfaceSegregationEntry
	ContainmentDepth     []containmentDepthEntry
	TotalSymbols         int
	TotalEdges           int
	TotalFiles           int
}

// EmitGraphAnalysis runs graph algorithms on the SymbolGraph and writes
// a graph_analysis.skt file for LLM consumption.
func (e *Emitter) EmitGraphAnalysis(sg *ir.SymbolGraph, opts *AnalysisOptions) error {
	res := runAnalysis(sg, opts)
	content := renderAnalysis(res, opts)

	if dir := filepath.Dir(opts.OutputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // 0o700 is the least-privilege permission for directories (execute bit required for traversal)
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := os.WriteFile(opts.OutputPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write graph analysis: %w", err)
	}

	return nil
}

func runAnalysis(sg *ir.SymbolGraph, opts *AnalysisOptions) *analysisResult {
	g := buildGraphCtx(sg)

	res := &analysisResult{
		TotalSymbols: len(sg.Symbols),
		TotalEdges:   len(sg.Edges),
		TotalFiles:   len(sg.FileOrder),
	}

	res.Cycles = detectCycles(g.adj)
	res.Hubs = detectHubs(g, opts.FanThreshold, opts)
	res.Orphans = detectOrphans(g)
	res.GodSymbols = detectGodSymbols(g, opts.GodThreshold)
	res.Instability = computeInstability(g)
	res.DeepInherit = detectDeepInheritance(g.inheritEdges, opts.MaxInheritanceDepth)
	res.Betweenness = computeBetweenness(g, opts)
	res.ArticulationPoints = findArticulationPoints(g)
	res.PageRanks = computePageRank(g, opts)
	res.TransitiveFanIn = computeTransitiveFanIn(g, opts)
	res.ChangePropagation = computeChangePropagation(g, opts)
	res.PkgCycles = detectPkgCycles(g)
	res.LayerViolations = detectLayerViolations(g)
	res.Unreachable = findUnreachable(g)
	res.WeakComponents = findWeakComponents(g)
	res.DependencyWeights = computeDependencyWeights(g, opts)
	res.DistanceFromMain = computeDistanceFromMainSequence(g)
	res.ShotgunSurgery = detectShotgunSurgery(g)
	res.FeatureEnvy = detectFeatureEnvy(g, opts)
	res.StableDepViolations = detectStableDepViolations(g)
	res.InterfaceSegregation = detectInterfaceSegregation(g)
	res.ContainmentDepth = computeContainmentDepth(g)

	return res
}

// ============================================================
// Original algorithms (refactored to use graphCtx)
// ============================================================

func detectHubs(g *graphCtx, threshold int, opts *AnalysisOptions) []hubSymbol {
	var hubs []hubSymbol
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		fi := g.fanIn[sid]
		fo := g.fanOut[sid]
		if fi >= threshold || fo >= threshold {
			hubs = append(hubs, hubSymbol{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				FanIn:     fi, FanOut: fo,
			})
		}
	}
	sort.Slice(hubs, func(i, j int) bool {
		ti := hubs[i].FanIn + hubs[i].FanOut
		tj := hubs[j].FanIn + hubs[j].FanOut
		return ti > tj
	})
	if opts.TopN > 0 && len(hubs) > opts.TopN {
		hubs = hubs[:opts.TopN]
	}
	return hubs
}

func detectOrphans(g *graphCtx) []symbolEntry {
	// Build set of contained targets for fast lookup.
	containedTargets := make(map[string]bool, len(g.containsRev))
	for child := range g.containsRev {
		containedTargets[child] = true
	}

	var orphans []symbolEntry
	for _, sid := range g.allSIDs {
		if g.fanIn[sid] > 0 || g.containsCount[sid] > 0 {
			continue
		}
		sym := g.sidInfo[sid]
		if sym.Category == plugin.CategoryModule {
			continue
		}
		if containedTargets[sid] {
			continue
		}
		orphans = append(orphans, symbolEntry{
			baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
			Category:  string(sym.Category),
		})
	}
	return orphans
}

func detectGodSymbols(g *graphCtx, threshold int) []godSymbol {
	var gods []godSymbol
	for sid, count := range g.containsCount {
		if count >= threshold {
			sym := g.sidInfo[sid]
			if sym == nil {
				continue
			}
			// Skip module/package symbols — they naturally contain all
			// top-level declarations and aren't actionable god classes.
			if sym.Category == plugin.CategoryModule {
				continue
			}
			gods = append(gods, godSymbol{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				Children:  count,
			})
		}
	}
	sort.Slice(gods, func(i, j int) bool {
		return gods[i].Children > gods[j].Children
	})
	return gods
}

func computeInstability(g *graphCtx) []instabilityEntry {
	var entries []instabilityEntry
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		if sym.Category != plugin.CategoryType && sym.Category != plugin.CategoryModule {
			continue
		}
		ca := g.fanIn[sid]
		ce := g.fanOut[sid]
		if ca+ce == 0 {
			continue
		}
		inst := float64(ce) / float64(ca+ce)
		// Only report symbols that have both incoming and outgoing coupling.
		// Pure sinks (I=0, Ce=0) and pure sources (I=1, Ca=0) with no
		// counterpart aren't actionable — filter to symbols with Ce > 0.
		if ce == 0 {
			continue
		}
		entries = append(entries, instabilityEntry{
			baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
			Ca:        ca, Ce: ce, Instability: inst,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Instability > entries[j].Instability
	})
	return entries
}

// detectCycles uses Tarjan's SCC algorithm to find cycles.
func detectCycles(adj map[string][]string) []cycle {
	nodes := make(map[string]bool)
	for k, vs := range adj {
		nodes[k] = true
		for _, v := range vs {
			nodes[v] = true
		}
	}

	index := 0
	stack := []string{}
	onStack := make(map[string]bool)
	indices := make(map[string]int)
	lowlinks := make(map[string]int)
	var sccs [][]string

	var strongconnect func(v string)
	strongconnect = func(v string) {
		indices[v] = index
		lowlinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adj[v] {
			if _, visited := indices[w]; !visited {
				strongconnect(w)
				if lowlinks[w] < lowlinks[v] {
					lowlinks[v] = lowlinks[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlinks[v] {
					lowlinks[v] = indices[w]
				}
			}
		}

		if lowlinks[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			if len(scc) > 1 {
				sort.Strings(scc)
				sccs = append(sccs, scc)
			}
		}
	}

	sorted := make([]string, 0, len(nodes))
	for n := range nodes {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	for _, n := range sorted {
		if _, visited := indices[n]; !visited {
			strongconnect(n)
		}
	}

	cycles := make([]cycle, len(sccs))
	for i, scc := range sccs {
		cycles[i] = cycle{Members: scc}
	}
	return cycles
}

func detectDeepInheritance(inheritEdges map[string]string, maxDepth int) []inheritanceChain {
	memo := make(map[string][]string)

	var getChain func(sid string, visited map[string]bool) []string
	getChain = func(sid string, visited map[string]bool) []string {
		if chain, ok := memo[sid]; ok {
			return chain
		}
		if visited[sid] {
			return []string{sid}
		}
		visited[sid] = true

		parent, hasParent := inheritEdges[sid]
		if !hasParent {
			chain := []string{sid}
			memo[sid] = chain
			return chain
		}
		parentChain := getChain(parent, visited)
		chain := make([]string, len(parentChain)+1)
		copy(chain, parentChain)
		chain[len(parentChain)] = sid
		memo[sid] = chain
		return chain
	}

	var results []inheritanceChain
	for sid := range inheritEdges {
		visited := make(map[string]bool)
		chain := getChain(sid, visited)
		depth := len(chain) - 1
		if depth >= maxDepth {
			results = append(results, inheritanceChain{
				Leaf: sid, Depth: depth, Chain: chain,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Depth > results[j].Depth
	})
	return results
}

// ============================================================
// New algorithms
// ============================================================

// --- Betweenness centrality (Brandes' algorithm) ---

func computeBetweenness(g *graphCtx, opts *AnalysisOptions) []betweennessEntry {
	cb := make(map[string]float64, len(g.allSIDs))

	for _, s := range g.allSIDs {
		// BFS from s.
		stack := []string{}
		pred := make(map[string][]string)
		sigma := make(map[string]float64)
		dist := make(map[string]int)
		for _, v := range g.allSIDs {
			dist[v] = -1
		}
		sigma[s] = 1
		dist[s] = 0
		queue := []string{s}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)
			for _, w := range g.adj[v] {
				if dist[w] < 0 {
					dist[w] = dist[v] + 1
					queue = append(queue, w)
				}
				if dist[w] == dist[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}

		delta := make(map[string]float64)
		for len(stack) > 0 {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, v := range pred[w] {
				delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
			}
			if w != s {
				cb[w] += delta[w]
			}
		}
	}

	// Normalize by (n-1)*(n-2) for directed graphs.
	n := float64(len(g.allSIDs))
	norm := 1.0
	if n > 2 {
		norm = 1.0 / ((n - 1) * (n - 2))
	}

	var entries []betweennessEntry
	for _, sid := range g.allSIDs {
		val := cb[sid] * norm
		if val > opts.BetweennessThreshold {
			sym := g.sidInfo[sid]
			entries = append(entries, betweennessEntry{
				baseEntry:   baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				Betweenness: val,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Betweenness > entries[j].Betweenness
	})
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}
	return entries
}

// --- Articulation points (cut vertices) ---
// Uses undirected DFS on the non-contains edge graph.
// Only reports points where removal creates two components of size >= 5 each.

const minSplitSize = 5

func findArticulationPoints(g *graphCtx) []articulationEntry {
	// Build undirected adjacency.
	undirected := make(map[string]map[string]bool)
	addUndirected := func(a, b string) {
		if undirected[a] == nil {
			undirected[a] = make(map[string]bool)
		}
		if undirected[b] == nil {
			undirected[b] = make(map[string]bool)
		}
		undirected[a][b] = true
		undirected[b][a] = true
	}
	for from, tos := range g.adj {
		for _, to := range tos {
			addUndirected(from, to)
		}
	}

	disc := make(map[string]int)
	low := make(map[string]int)
	parent := make(map[string]string)
	isAP := make(map[string]bool)
	timer := 0

	var dfs func(u string)
	dfs = func(u string) {
		disc[u] = timer
		low[u] = timer
		timer++
		childCount := 0

		for nb := range undirected[u] {
			if _, visited := disc[nb]; !visited {
				childCount++
				parent[nb] = u
				dfs(nb)
				if low[nb] < low[u] {
					low[u] = low[nb]
				}
				if _, hasParent := parent[u]; !hasParent {
					if childCount > 1 {
						isAP[u] = true
					}
				} else if low[nb] >= disc[u] {
					isAP[u] = true
				}
			} else if nb != parent[u] {
				if disc[nb] < low[u] {
					low[u] = disc[nb]
				}
			}
		}
	}

	for _, sid := range g.allSIDs {
		if _, visited := disc[sid]; !visited {
			dfs(sid)
		}
	}

	// For each AP, simulate removal and measure resulting component sizes.
	var entries []articulationEntry
	for _, sid := range g.allSIDs {
		if !isAP[sid] {
			continue
		}
		// BFS on undirected graph excluding this node.
		visited := map[string]bool{sid: true} // mark removed node as visited
		var compSizes []int
		for _, start := range g.allSIDs {
			if visited[start] {
				continue
			}
			if undirected[start] == nil {
				visited[start] = true
				continue
			}
			size := 0
			queue := []string{start}
			visited[start] = true
			for len(queue) > 0 {
				v := queue[0]
				queue = queue[1:]
				size++
				for nb := range undirected[v] {
					if !visited[nb] {
						visited[nb] = true
						queue = append(queue, nb)
					}
				}
			}
			compSizes = append(compSizes, size)
		}

		if len(compSizes) < 2 {
			continue // not actually splitting into multiple components
		}

		sort.Sort(sort.Reverse(sort.IntSlice(compSizes)))
		maxSplit := compSizes[0]
		minSplit := compSizes[1]

		// Only report if both sides are non-trivial.
		if minSplit >= minSplitSize {
			sym := g.sidInfo[sid]
			entries = append(entries, articulationEntry{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				MinSplit:  minSplit, MaxSplit: maxSplit,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].MinSplit > entries[j].MinSplit
	})
	return entries
}

// --- PageRank ---

func computePageRank(g *graphCtx, opts *AnalysisOptions) []pageRankEntry {
	n := len(g.allSIDs)
	if n == 0 {
		return nil
	}

	damping := 0.85
	iterations := 30
	initial := 1.0 / float64(n)

	rank := make(map[string]float64, n)
	for _, sid := range g.allSIDs {
		rank[sid] = initial
	}

	for iter := 0; iter < iterations; iter++ {
		newRank := make(map[string]float64, n)
		for _, sid := range g.allSIDs {
			newRank[sid] = (1 - damping) / float64(n)
		}
		for _, sid := range g.allSIDs {
			outDeg := len(g.adj[sid])
			if outDeg == 0 {
				// Dangling node: distribute rank evenly.
				share := rank[sid] / float64(n)
				for _, other := range g.allSIDs {
					newRank[other] += damping * share
				}
			} else {
				share := rank[sid] / float64(outDeg)
				for _, target := range g.adj[sid] {
					newRank[target] += damping * share
				}
			}
		}
		rank = newRank
	}

	var entries []pageRankEntry
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		entries = append(entries, pageRankEntry{
			baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
			Rank:      rank[sid],
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rank > entries[j].Rank
	})
	// Top N.
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}
	return entries
}

// --- Transitive fan-in (afferent coupling depth) ---

func computeTransitiveFanIn(g *graphCtx, opts *AnalysisOptions) []transitiveFanInEntry {
	// For each symbol, BFS backwards through radj to count all transitive dependents.
	type result struct {
		sid        string
		direct     int
		transitive int
	}

	var results []result
	for _, sid := range g.allSIDs {
		direct := g.fanIn[sid]
		if direct == 0 {
			continue
		}
		// BFS on reverse adjacency.
		visited := map[string]bool{sid: true}
		queue := []string{sid}
		count := 0
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			for _, pred := range g.radj[v] {
				if !visited[pred] {
					visited[pred] = true
					count++
					queue = append(queue, pred)
				}
			}
		}
		if count > direct { // only interesting when transitive > direct
			results = append(results, result{sid: sid, direct: direct, transitive: count})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].transitive > results[j].transitive
	})
	if opts.TopN > 0 && len(results) > opts.TopN {
		results = results[:opts.TopN]
	}

	entries := make([]transitiveFanInEntry, len(results))
	for i, r := range results {
		sym := g.sidInfo[r.sid]
		entries[i] = transitiveFanInEntry{
			baseEntry:       baseEntry{SID: r.sid, Name: sym.Name, File: sym.FilePath},
			DirectFanIn:     r.direct,
			TransitiveFanIn: r.transitive,
		}
	}
	return entries
}

// --- Change propagation probability ---
// Simulates forward propagation from each symbol using edge-kind weights.

func computeChangePropagation(g *graphCtx, opts *AnalysisOptions) []changePropEntry {
	// Edge kind → propagation probability.
	kindWeight := map[plugin.EdgeKind]float64{
		plugin.EdgeInherits:   0.9,
		plugin.EdgeImplements: 0.85,
		plugin.EdgeOverrides:  0.8,
		plugin.EdgeContains:   0.7,
		plugin.EdgeCalls:      0.5,
		plugin.EdgeReferences: 0.3,
		plugin.EdgeDecorates:  0.2,
	}

	// Build weighted forward adjacency from the raw edges.
	type weightedEdge struct {
		to     string
		weight float64
	}
	wadj := make(map[string][]weightedEdge)
	for _, edge := range g.sg.Edges {
		fromSID, fromOK := g.sg.ShortIDs[edge.From]
		toSID, toOK := g.sg.ShortIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}
		w := kindWeight[edge.Kind]
		if w == 0 {
			w = 0.3
		}
		// Propagation goes forward: if A calls B, a change in B propagates to A.
		// So we reverse: from target to dependents.
		wadj[toSID] = append(wadj[toSID], weightedEdge{to: fromSID, weight: w})
	}

	var entries []changePropEntry
	for _, sid := range g.allSIDs {
		if len(wadj[sid]) == 0 && g.fanIn[sid] == 0 {
			continue
		}
		// BFS with probability decay.
		visited := map[string]bool{sid: true}
		type queueItem struct {
			node string
			prob float64
		}
		queue := []queueItem{{sid, 1.0}}
		blastRadius := 0
		totalProb := 0.0

		for len(queue) > 0 {
			item := queue[0]
			queue = queue[1:]
			for _, we := range wadj[item.node] {
				newProb := item.prob * we.weight
				if newProb < opts.PropagationCutoff || visited[we.to] { // cutoff
					continue
				}
				visited[we.to] = true
				blastRadius++
				totalProb += newProb
				queue = append(queue, queueItem{we.to, newProb})
			}
		}

		if blastRadius > 0 {
			sym := g.sidInfo[sid]
			entries = append(entries, changePropEntry{
				baseEntry:   baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				BlastRadius: blastRadius,
				Probability: totalProb / float64(blastRadius),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].BlastRadius > entries[j].BlastRadius
	})
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}
	return entries
}

// --- Circular package dependencies ---

func detectPkgCycles(g *graphCtx) []cycle {
	// Build package-level adjacency.
	pkgAdj := make(map[string][]string)
	seen := make(map[[2]string]bool)
	for pair := range g.pkgEdges {
		if !seen[pair] {
			seen[pair] = true
			pkgAdj[pair[0]] = append(pkgAdj[pair[0]], pair[1])
		}
	}

	sccs := detectCycles(pkgAdj)
	return sccs
}

// --- Layer violation detection ---
// Infers layers from directory naming conventions.

var layerOrder = map[string]int{
	"handler":        0,
	"handlers":       0,
	"controller":     0,
	"controllers":    0,
	"api":            0,
	"route":          0,
	"routes":         0,
	"cmd":            0,
	"service":        1,
	"services":       1,
	"usecase":        1,
	"usecases":       1,
	"domain":         1,
	"model":          2,
	"models":         2,
	"entity":         2,
	"entities":       2,
	"repo":           3,
	"repository":     3,
	"store":          3,
	"db":             3,
	"database":       3,
	"infra":          3,
	"infrastructure": 3,
}

func inferLayer(filePath string) (layer string, order int, found bool) {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		lower := strings.ToLower(parts[i])
		if order, ok := layerOrder[lower]; ok {
			return lower, order, true
		}
	}
	return "", -1, false
}

func detectLayerViolations(g *graphCtx) []layerViolation {
	var violations []layerViolation

	for _, edge := range g.sg.Edges {
		if edge.Kind == plugin.EdgeContains || edge.Kind == plugin.EdgeImports {
			continue
		}
		fromSID, fromOK := g.sg.ShortIDs[edge.From]
		toSID, toOK := g.sg.ShortIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}
		fromSym := g.sidInfo[fromSID]
		toSym := g.sidInfo[toSID]
		if fromSym == nil || toSym == nil {
			continue
		}

		fromLayer, fromOrder, fromHas := inferLayer(fromSym.FilePath)
		toLayer, toOrder, toHas := inferLayer(toSym.FilePath)
		if !fromHas || !toHas || fromLayer == toLayer {
			continue
		}

		// Violation: lower layer depends on higher layer (inverted dependency).
		if fromOrder > toOrder {
			violations = append(violations, layerViolation{
				FromSID:   fromSID,
				ToSID:     toSID,
				FromLayer: fromLayer,
				ToLayer:   toLayer,
				FromFile:  fromSym.FilePath,
				EdgeKind:  string(edge.Kind),
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].FromSID < violations[j].FromSID
	})
	return violations
}

// --- Reachability from entry points ---
// Entry points: main functions, exported symbols with no incoming edges,
// and symbols in files named main.*, app.*, index.*, etc.

func isEntryPoint(sid string, g *graphCtx) bool {
	sym := g.sidInfo[sid]
	if sym == nil {
		return false
	}
	name := strings.ToLower(sym.Name)
	if name == "main" || name == "app" || name == "init" {
		return true
	}
	base := strings.ToLower(filepath.Base(sym.FilePath))
	if strings.HasPrefix(base, "main.") || strings.HasPrefix(base, "app.") ||
		strings.HasPrefix(base, "index.") || strings.HasPrefix(base, "server.") {
		// Exported/top-level symbols in entry files.
		if sym.Category == plugin.CategoryCallable || sym.Category == plugin.CategoryType {
			return true
		}
	}
	return false
}

func findUnreachable(g *graphCtx) []symbolEntry {
	// Collect entry points.
	var entryPoints []string
	for _, sid := range g.allSIDs {
		if isEntryPoint(sid, g) {
			entryPoints = append(entryPoints, sid)
		}
	}

	if len(entryPoints) == 0 {
		return nil // can't determine reachability without entry points
	}

	// BFS forward from all entry points using both adj and containsAdj.
	reachable := make(map[string]bool)
	queue := make([]string, len(entryPoints))
	copy(queue, entryPoints)
	for _, ep := range entryPoints {
		reachable[ep] = true
	}

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		// Follow call/reference edges.
		for _, w := range g.adj[v] {
			if !reachable[w] {
				reachable[w] = true
				queue = append(queue, w)
			}
		}
		// Follow contains edges (if a type is reachable, its members are too).
		for _, w := range g.containsAdj[v] {
			if !reachable[w] {
				reachable[w] = true
				queue = append(queue, w)
			}
		}
		// If a member is reachable, its parent container is too.
		if parent, ok := g.containsRev[v]; ok && !reachable[parent] {
			reachable[parent] = true
			queue = append(queue, parent)
		}
	}

	var unreachables []symbolEntry
	for _, sid := range g.allSIDs {
		if reachable[sid] {
			continue
		}
		sym := g.sidInfo[sid]
		if sym.Category == plugin.CategoryModule {
			continue
		}
		unreachables = append(unreachables, symbolEntry{
			baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
			Category:  string(sym.Category),
		})
	}
	return unreachables
}

// --- Weakly connected components ---

func findWeakComponents(g *graphCtx) []connectedComponent {
	// Build undirected adjacency from all edge types.
	undirected := make(map[string][]string)
	addEdge := func(a, b string) {
		undirected[a] = append(undirected[a], b)
		undirected[b] = append(undirected[b], a)
	}
	for from, tos := range g.adj {
		for _, to := range tos {
			addEdge(from, to)
		}
	}
	for from, tos := range g.containsAdj {
		for _, to := range tos {
			addEdge(from, to)
		}
	}

	visited := make(map[string]bool)
	var components []connectedComponent

	for _, sid := range g.allSIDs {
		if visited[sid] {
			continue
		}
		// BFS to find component.
		var members []string
		queue := []string{sid}
		visited[sid] = true
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			members = append(members, v)
			for _, w := range undirected[v] {
				if !visited[w] {
					visited[w] = true
					queue = append(queue, w)
				}
			}
		}

		if len(members) > 0 {
			sort.Strings(members)
			sample := make([]string, 0, 5)
			for i := 0; i < len(members) && i < 5; i++ {
				if sym := g.sidInfo[members[i]]; sym != nil {
					sample = append(sample, sym.Name)
				}
			}
			components = append(components, connectedComponent{
				Size:    len(members),
				Members: members,
				Sample:  sample,
			})
		}
	}

	sort.Slice(components, func(i, j int) bool {
		return components[i].Size > components[j].Size
	})

	// Only report if there are isolated components (more than 1 component).
	if len(components) <= 1 {
		return nil
	}
	return components
}

// --- Dependency weight (package-level coupling strength) ---

func computeDependencyWeights(g *graphCtx, opts *AnalysisOptions) []depWeightEntry {
	var entries []depWeightEntry
	for pair, weight := range g.pkgEdges {
		entries = append(entries, depWeightEntry{
			FromPkg: pair[0], ToPkg: pair[1], Weight: weight,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Weight > entries[j].Weight
	})
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}
	return entries
}

// --- Distance from Main Sequence (Robert C. Martin) ---
// Per package: D = |A + I - 1| where A = abstractness, I = instability.

func computeDistanceFromMainSequence(g *graphCtx) []distanceEntry {
	type pkgStats struct {
		totalTypes    int
		abstractTypes int
		ca            int // afferent: edges coming into this package
		ce            int // efferent: edges going out of this package
	}

	stats := make(map[string]*pkgStats)
	ensurePkg := func(pkg string) *pkgStats {
		if stats[pkg] == nil {
			stats[pkg] = &pkgStats{}
		}
		return stats[pkg]
	}

	// Count types and abstract types per package.
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		pkg := g.sidToPkg[sid]
		if pkg == "" {
			continue
		}
		s := ensurePkg(pkg)
		if sym.Category == plugin.CategoryType {
			s.totalTypes++
			kind := strings.ToLower(sym.Kind)
			if kind == "interface" || kind == "abstract_class" || kind == "trait" ||
				sym.Properties["abstract"] == "true" {
				s.abstractTypes++
			}
		}
	}

	// Count package-level coupling.
	for pair, weight := range g.pkgEdges {
		ensurePkg(pair[0]).ce += weight
		ensurePkg(pair[1]).ca += weight
	}

	var entries []distanceEntry
	for pkg, s := range stats {
		if s.totalTypes == 0 {
			continue // no types → abstractness is meaningless
		}
		abstractness := float64(s.abstractTypes) / float64(s.totalTypes)
		instability := 0.0
		if s.ca+s.ce > 0 {
			instability = float64(s.ce) / float64(s.ca+s.ce)
		}
		dist := math.Abs(abstractness + instability - 1.0)
		if dist < 0.1 {
			continue // near the main sequence — not interesting
		}
		entries = append(entries, distanceEntry{
			Pkg:          pkg,
			Instability:  instability,
			Abstractness: abstractness,
			Distance:     dist,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Distance > entries[j].Distance
	})
	return entries
}

// --- Shotgun surgery detection ---
// Finds groups of symbols that share the same set of callers, suggesting
// they change together and should be colocated or merged.

func detectShotgunSurgery(g *graphCtx) []shotgunSurgeryGroup {
	// For each symbol, compute its caller set (via radj).
	// Group symbols by their caller set signature.
	type callerSig struct {
		pkg string
		sig string // sorted, joined caller SIDs
	}

	sigToTargets := make(map[callerSig][]string)
	for _, sid := range g.allSIDs {
		callers := g.radj[sid]
		if len(callers) < 2 {
			continue // need at least 2 callers to form a pattern
		}
		sym := g.sidInfo[sid]
		if sym.Category == plugin.CategoryModule {
			continue
		}
		sorted := make([]string, len(callers))
		copy(sorted, callers)
		sort.Strings(sorted)
		key := callerSig{
			pkg: g.sidToPkg[sid],
			sig: strings.Join(sorted, ","),
		}
		sigToTargets[key] = append(sigToTargets[key], sid)
	}

	var groups []shotgunSurgeryGroup
	for key, targets := range sigToTargets {
		if len(targets) < 3 {
			continue // need at least 3 co-called symbols to be interesting
		}
		callers := strings.Split(key.sig, ",")
		names := make([]string, len(targets))
		for i, sid := range targets {
			names[i] = g.sidInfo[sid].Name
		}
		sort.Strings(names)
		groups = append(groups, shotgunSurgeryGroup{
			Callers: callers,
			Targets: targets,
			Names:   names,
			Pkg:     key.pkg,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Targets) > len(groups[j].Targets)
	})
	if len(groups) > 20 {
		groups = groups[:20]
	}
	return groups
}

// --- Feature envy detection ---
// A callable that references more symbols from another package than its own.

func detectFeatureEnvy(g *graphCtx, opts *AnalysisOptions) []featureEnvyEntry {
	var entries []featureEnvyEntry

	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		if sym.Category != plugin.CategoryCallable {
			continue
		}
		ownPkg := g.sidToPkg[sid]
		if ownPkg == "" {
			continue
		}

		// Count outgoing references by target package.
		pkgRefs := make(map[string]int)
		for _, target := range g.adj[sid] {
			targetPkg := g.sidToPkg[target]
			if targetPkg != "" {
				pkgRefs[targetPkg]++
			}
		}

		ownRefs := pkgRefs[ownPkg]
		// Find the most-referenced foreign package.
		bestPkg := ""
		bestCount := 0
		for pkg, count := range pkgRefs {
			if pkg == ownPkg {
				continue
			}
			if count > bestCount {
				bestCount = count
				bestPkg = pkg
			}
		}

		// Feature envy: more foreign refs than own refs, and at least 3 foreign refs.
		if bestCount > ownRefs && bestCount >= 3 {
			entries = append(entries, featureEnvyEntry{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				OwnPkg:    ownPkg, EnviedPkg: bestPkg,
				OwnRefs: ownRefs, ForeignRefs: bestCount,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ForeignRefs-entries[i].OwnRefs > entries[j].ForeignRefs-entries[j].OwnRefs
	})
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}
	return entries
}

// --- Stable dependencies principle violation ---
// A stable package (low I) depending on an unstable package (high I).

func detectStableDepViolations(g *graphCtx) []stableDepViolation {
	// Compute per-package instability.
	type pkgCoupling struct {
		ca int
		ce int
	}
	pkgStats := make(map[string]*pkgCoupling)
	ensure := func(pkg string) *pkgCoupling {
		if pkgStats[pkg] == nil {
			pkgStats[pkg] = &pkgCoupling{}
		}
		return pkgStats[pkg]
	}
	for pair, weight := range g.pkgEdges {
		ensure(pair[0]).ce += weight
		ensure(pair[1]).ca += weight
	}

	instab := func(pkg string) float64 {
		s := pkgStats[pkg]
		if s == nil || s.ca+s.ce == 0 {
			return 0
		}
		return float64(s.ce) / float64(s.ca+s.ce)
	}

	var violations []stableDepViolation
	seen := make(map[[2]string]bool)
	for pair := range g.pkgEdges {
		if seen[pair] {
			continue
		}
		seen[pair] = true
		fromI := instab(pair[0])
		toI := instab(pair[1])
		// Violation: a more stable package depends on a less stable one.
		// Only flag when the difference is significant (> 0.3).
		if fromI < toI && (toI-fromI) > 0.3 {
			violations = append(violations, stableDepViolation{
				FromPkg: pair[0], ToPkg: pair[1],
				FromInstab: fromI, ToInstab: toI,
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		di := violations[i].ToInstab - violations[i].FromInstab
		dj := violations[j].ToInstab - violations[j].FromInstab
		return di > dj
	})
	return violations
}

// --- Interface segregation smell ---
// Interfaces with many members where implementors only use a subset.

func detectInterfaceSegregation(g *graphCtx) []interfaceSegregationEntry {
	// Find interface/trait symbols.
	var interfaces []string
	for _, sid := range g.allSIDs {
		sym := g.sidInfo[sid]
		if sym.Category != plugin.CategoryType {
			continue
		}
		kind := strings.ToLower(sym.Kind)
		if kind == "interface" || kind == "trait" || kind == "protocol" {
			interfaces = append(interfaces, sid)
		}
	}

	// For each interface, count its members (via contains edges)
	// and find implementors (via implements/inherits edges pointing to it).
	var entries []interfaceSegregationEntry
	for _, ifaceSID := range interfaces {
		members := g.containsAdj[ifaceSID]
		if len(members) < 4 {
			continue // small interfaces aren't a smell
		}

		// Find implementors: symbols that have an inherits/implements edge TO this interface.
		var implementors []string
		for child, parent := range g.inheritEdges {
			if parent == ifaceSID {
				implementors = append(implementors, child)
			}
		}
		if len(implementors) == 0 {
			continue
		}

		// For each implementor, count how many of the interface's members
		// it overrides (has a contains-child with the same name).
		memberNames := make(map[string]bool, len(members))
		for _, mSID := range members {
			if sym := g.sidInfo[mSID]; sym != nil {
				memberNames[sym.Name] = true
			}
		}

		maxUsed := 0
		for _, implSID := range implementors {
			used := 0
			for _, childSID := range g.containsAdj[implSID] {
				if sym := g.sidInfo[childSID]; sym != nil {
					if memberNames[sym.Name] {
						used++
					}
				}
			}
			if used > maxUsed {
				maxUsed = used
			}
		}

		// Smell: if the best implementor uses less than 60% of the interface.
		ratio := float64(maxUsed) / float64(len(members))
		if ratio < 0.6 {
			sym := g.sidInfo[ifaceSID]
			entries = append(entries, interfaceSegregationEntry{
				baseEntry: baseEntry{SID: ifaceSID, Name: sym.Name, File: sym.FilePath},
				Members:   len(members), MaxUsed: maxUsed, Impls: len(implementors),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Members > entries[j].Members
	})
	return entries
}

// --- Containment depth ---
// Detects deeply nested containment hierarchies (class → inner class → ...).

func computeContainmentDepth(g *graphCtx) []containmentDepthEntry {
	// Build depth for each symbol in the containment tree.
	depth := make(map[string]int)
	chain := make(map[string][]string) // sid → chain of names from root

	// Find roots (symbols that are not contained by anything).
	roots := make([]string, 0)
	for _, sid := range g.allSIDs {
		if _, hasParent := g.containsRev[sid]; !hasParent {
			if len(g.containsAdj[sid]) > 0 {
				roots = append(roots, sid)
			}
		}
	}

	// BFS from roots through containment edges.
	for _, root := range roots {
		rootSym := g.sidInfo[root]
		if rootSym == nil {
			continue
		}
		queue := []string{root}
		depth[root] = 0
		chain[root] = []string{rootSym.Name}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			for _, child := range g.containsAdj[v] {
				if _, visited := depth[child]; visited {
					continue
				}
				childSym := g.sidInfo[child]
				if childSym == nil {
					continue
				}
				// Only count type/callable nesting, not fields/values.
				if childSym.Category != plugin.CategoryType && childSym.Category != plugin.CategoryCallable {
					continue
				}
				depth[child] = depth[v] + 1
				newChain := make([]string, len(chain[v])+1)
				copy(newChain, chain[v])
				newChain[len(chain[v])] = childSym.Name
				chain[child] = newChain
				queue = append(queue, child)
			}
		}
	}

	// Report symbols with depth >= 3 (3 levels of nesting).
	var entries []containmentDepthEntry
	for sid, d := range depth {
		if d >= 3 {
			sym := g.sidInfo[sid]
			entries = append(entries, containmentDepthEntry{
				baseEntry: baseEntry{SID: sid, Name: sym.Name, File: sym.FilePath},
				Depth:     d, Chain: chain[sid],
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Depth > entries[j].Depth
	})
	if len(entries) > 30 {
		entries = entries[:30]
	}
	return entries
}

// ============================================================
// Renderer
// ============================================================

func renderAnalysis(res *analysisResult, opts *AnalysisOptions) string {
	var b strings.Builder

	b.WriteString("[graph_analysis]\n")
	fmt.Fprintf(&b, "symbols: %d | edges: %d | files: %d\n\n", res.TotalSymbols, res.TotalEdges, res.TotalFiles)

	// --- Original sections ---

	b.WriteString("[cyclic_dependencies]\n")
	if len(res.Cycles) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d cycle(s) found\n", len(res.Cycles))
		for i, c := range res.Cycles {
			fmt.Fprintf(&b, "  cycle_%d: %s\n", i+1, strings.Join(c.Members, " <-> "))
		}
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "[hub_symbols] (fan-in or fan-out >= %d)\n", opts.FanThreshold)
	if len(res.Hubs) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, h := range res.Hubs {
			fmt.Fprintf(&b, "  %s %s fan_in=%d fan_out=%d | %s\n", h.SID, h.Name, h.FanIn, h.FanOut, h.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[orphan_symbols] (zero incoming edges — potential dead code)\n")
	if len(res.Orphans) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d orphan(s)\n", len(res.Orphans))
		for _, o := range res.Orphans {
			fmt.Fprintf(&b, "  %s %s [%s] | %s\n", o.SID, o.Name, o.Category, o.File)
		}
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "[god_symbols] (contains >= %d children)\n", opts.GodThreshold)
	if len(res.GodSymbols) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, g := range res.GodSymbols {
			fmt.Fprintf(&b, "  %s %s children=%d | %s\n", g.SID, g.Name, g.Children, g.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[instability] (Ce/(Ca+Ce) — symbols with both incoming and outgoing coupling)\n")
	if len(res.Instability) == 0 {
		b.WriteString("no type/module symbols with bidirectional coupling\n")
	} else {
		for _, e := range res.Instability {
			fmt.Fprintf(&b, "  %s %s I=%.2f Ca=%d Ce=%d | %s\n", e.SID, e.Name, e.Instability, e.Ca, e.Ce, e.File)
		}
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "[deep_inheritance] (depth >= %d)\n", opts.MaxInheritanceDepth)
	if len(res.DeepInherit) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, d := range res.DeepInherit {
			fmt.Fprintf(&b, "  depth=%d: %s\n", d.Depth, strings.Join(d.Chain, " -> "))
		}
	}
	b.WriteByte('\n')

	// --- New sections ---

	b.WriteString("[betweenness_centrality] (bottleneck symbols — bridges between clusters)\n")
	if len(res.Betweenness) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, e := range res.Betweenness {
			fmt.Fprintf(&b, "  %s %s BC=%.4f | %s\n", e.SID, e.Name, e.Betweenness, e.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[articulation_points] (removal splits graph into components >= 5 each)\n")
	if len(res.ArticulationPoints) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d critical bridge(s)\n", len(res.ArticulationPoints))
		for _, a := range res.ArticulationPoints {
			fmt.Fprintf(&b, "  %s %s splits=[%d, %d] | %s\n", a.SID, a.Name, a.MaxSplit, a.MinSplit, a.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[pagerank] (importance by recursive incoming references)\n")
	if len(res.PageRanks) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, p := range res.PageRanks {
			fmt.Fprintf(&b, "  %s %s PR=%.6f | %s\n", p.SID, p.Name, p.Rank, p.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[transitive_fan_in] (blast radius — transitive dependents > direct)\n")
	if len(res.TransitiveFanIn) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, t := range res.TransitiveFanIn {
			fmt.Fprintf(&b, "  %s %s direct=%d transitive=%d | %s\n",
				t.SID, t.Name, t.DirectFanIn, t.TransitiveFanIn, t.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[change_propagation] (simulated change ripple effect)\n")
	if len(res.ChangePropagation) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, c := range res.ChangePropagation {
			fmt.Fprintf(&b, "  %s %s blast_radius=%d avg_prob=%.2f | %s\n",
				c.SID, c.Name, c.BlastRadius, c.Probability, c.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[circular_package_dependencies]\n")
	if len(res.PkgCycles) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d package cycle(s) found\n", len(res.PkgCycles))
		for i, c := range res.PkgCycles {
			fmt.Fprintf(&b, "  pkg_cycle_%d: %s\n", i+1, strings.Join(c.Members, " <-> "))
		}
	}
	b.WriteByte('\n')

	b.WriteString("[layer_violations] (lower layer depends on higher layer)\n")
	if len(res.LayerViolations) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d violation(s)\n", len(res.LayerViolations))
		for _, v := range res.LayerViolations {
			fmt.Fprintf(&b, "  %s --%s--> %s | %s(%s) -> %s\n",
				v.FromSID, v.EdgeKind, v.ToSID, v.FromLayer, v.FromFile, v.ToLayer)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[unreachable_symbols] (not reachable from entry points)\n")
	if len(res.Unreachable) == 0 {
		b.WriteString("none detected (or no entry points identified)\n")
	} else {
		fmt.Fprintf(&b, "%d unreachable symbol(s)\n", len(res.Unreachable))
		for _, u := range res.Unreachable {
			fmt.Fprintf(&b, "  %s %s [%s] | %s\n", u.SID, u.Name, u.Category, u.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[weakly_connected_components] (isolated subgraphs)\n")
	if len(res.WeakComponents) == 0 {
		b.WriteString("fully connected (single component)\n")
	} else {
		fmt.Fprintf(&b, "%d component(s)\n", len(res.WeakComponents))
		for i, c := range res.WeakComponents {
			fmt.Fprintf(&b, "  component_%d: size=%d sample=[%s]\n",
				i+1, c.Size, strings.Join(c.Sample, ", "))
		}
	}
	b.WriteByte('\n')

	b.WriteString("[dependency_weight] (package coupling strength — top cross-package edges)\n")
	if len(res.DependencyWeights) == 0 {
		b.WriteString("no cross-package dependencies\n")
	} else {
		for _, d := range res.DependencyWeights {
			fmt.Fprintf(&b, "  %s -> %s weight=%d\n", d.FromPkg, d.ToPkg, d.Weight)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[distance_from_main_sequence] (|A+I-1| — packages far from ideal, D >= 0.1)\n")
	if len(res.DistanceFromMain) == 0 {
		b.WriteString("all packages near the main sequence\n")
	} else {
		for _, d := range res.DistanceFromMain {
			fmt.Fprintf(&b, "  %s D=%.2f A=%.2f I=%.2f\n", d.Pkg, d.Distance, d.Abstractness, d.Instability)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[shotgun_surgery] (symbols that share the same callers — likely co-change)\n")
	if len(res.ShotgunSurgery) == 0 {
		b.WriteString("none detected\n")
	} else {
		fmt.Fprintf(&b, "%d group(s)\n", len(res.ShotgunSurgery))
		for i, g := range res.ShotgunSurgery {
			fmt.Fprintf(&b, "  group_%d: %d symbols in %s shared by %d callers: [%s]\n",
				i+1, len(g.Targets), g.Pkg, len(g.Callers), strings.Join(g.Names, ", "))
		}
	}
	b.WriteByte('\n')

	b.WriteString("[feature_envy] (callables referencing another package more than their own)\n")
	if len(res.FeatureEnvy) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, f := range res.FeatureEnvy {
			fmt.Fprintf(&b, "  %s %s own=%d foreign=%d envies=%s | %s\n",
				f.SID, f.Name, f.OwnRefs, f.ForeignRefs, f.EnviedPkg, f.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[stable_dependency_violations] (stable package depends on unstable package)\n")
	if len(res.StableDepViolations) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, v := range res.StableDepViolations {
			fmt.Fprintf(&b, "  %s (I=%.2f) -> %s (I=%.2f)\n",
				v.FromPkg, v.FromInstab, v.ToPkg, v.ToInstab)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[interface_segregation] (large interfaces where implementors use < 60%% of members)\n")
	if len(res.InterfaceSegregation) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, e := range res.InterfaceSegregation {
			fmt.Fprintf(&b, "  %s %s members=%d max_used=%d impls=%d | %s\n",
				e.SID, e.Name, e.Members, e.MaxUsed, e.Impls, e.File)
		}
	}
	b.WriteByte('\n')

	b.WriteString("[containment_depth] (deeply nested type/callable hierarchies, depth >= 3)\n")
	if len(res.ContainmentDepth) == 0 {
		b.WriteString("none detected\n")
	} else {
		for _, c := range res.ContainmentDepth {
			fmt.Fprintf(&b, "  %s %s depth=%d chain=[%s] | %s\n",
				c.SID, c.Name, c.Depth, strings.Join(c.Chain, " > "), c.File)
		}
	}
	b.WriteByte('\n')

	return b.String()
}
