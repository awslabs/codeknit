// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package hotspot combines Git history with structural graph metrics.
package hotspot

import (
	"math"
	"sort"
	"time"

	"codeknit/internal/history"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// Options controls hotspot scoring and result limits.
type Options struct {
	TopN         int
	MinCoChanges int
}

// Entry is one ranked file hotspot.
type Entry struct {
	LastChanged    time.Time `json:"last_changed"`
	File           string    `json:"file"`
	Score          float64   `json:"score"`
	HistoryScore   float64   `json:"history_score"`
	StructureScore float64   `json:"structure_score"`
	RecencyScore   float64   `json:"recency_score"`
	PageRank       float64   `json:"pagerank"`
	Betweenness    float64   `json:"betweenness"`
	Commits        int       `json:"commits"`
	Churn          int       `json:"churn"`
	TransitiveIn   int       `json:"transitive_fan_in"`
}

// Coupling is a pair of files that frequently change together.
type Coupling struct {
	Left      string  `json:"left"`
	Right     string  `json:"right"`
	Strength  float64 `json:"strength"`
	CoChanges int     `json:"co_changes"`
}

// Result is the complete hotspot analysis.
type Result struct {
	GeneratedAt        time.Time  `json:"generated_at"`
	Since              time.Time  `json:"since"`
	Confidence         string     `json:"confidence"`
	Hotspots           []Entry    `json:"hotspots"`
	TemporalCoupling   []Coupling `json:"temporal_coupling"`
	CommitsVisited     int        `json:"commits_visited"`
	CommitsAnalyzed    int        `json:"commits_analyzed"`
	SkippedMerges      int        `json:"skipped_merges"`
	SkippedBulkCommits int        `json:"skipped_bulk_commits"`
}

type structuralMetrics struct {
	pageRank     float64
	betweenness  float64
	transitiveIn int
}

// Analyze ranks current source files using historical and structural metrics.
func Analyze(sg *ir.SymbolGraph, hist *history.Result, since, generatedAt time.Time, opts Options) *Result {
	structure := computeStructure(sg, hist.RepositoryRoot)
	maxima := findMaxima(hist.Files, structure)

	entries := make([]Entry, 0, len(hist.Files))
	for file, metrics := range hist.Files {
		sm := structure[file]
		commitScore := logNormalize(float64(metrics.Commits), maxima.commits)
		churnScore := logNormalize(float64(metrics.Churn()), maxima.churn)
		recencyScore := linearNormalize(metrics.RecencyScore, maxima.recency)
		historyScore := 0.45*commitScore + 0.35*churnScore + 0.20*recencyScore

		pageRankScore := linearNormalize(sm.pageRank, maxima.pageRank)
		transitiveScore := logNormalize(float64(sm.transitiveIn), maxima.transitiveIn)
		betweennessScore := linearNormalize(sm.betweenness, maxima.betweenness)
		structureScore := 0.40*pageRankScore + 0.35*transitiveScore + 0.25*betweennessScore

		entries = append(entries, Entry{
			File:           file,
			Score:          historyScore * (0.5 + 0.5*structureScore),
			HistoryScore:   historyScore,
			StructureScore: structureScore,
			RecencyScore:   recencyScore,
			Commits:        metrics.Commits,
			Churn:          metrics.Churn(),
			LastChanged:    metrics.LastChanged,
			PageRank:       sm.pageRank,
			Betweenness:    sm.betweenness,
			TransitiveIn:   sm.transitiveIn,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score != entries[j].Score {
			return entries[i].Score > entries[j].Score
		}
		return entries[i].File < entries[j].File
	})
	if opts.TopN > 0 && len(entries) > opts.TopN {
		entries = entries[:opts.TopN]
	}

	couplings := rankCouplings(hist, opts.MinCoChanges)
	if opts.TopN > 0 && len(couplings) > opts.TopN {
		couplings = couplings[:opts.TopN]
	}

	return &Result{
		GeneratedAt:        generatedAt,
		Since:              since,
		Confidence:         confidence(hist.CommitsAnalyzed),
		Hotspots:           entries,
		TemporalCoupling:   couplings,
		CommitsVisited:     hist.CommitsVisited,
		CommitsAnalyzed:    hist.CommitsAnalyzed,
		SkippedMerges:      hist.SkippedMerges,
		SkippedBulkCommits: hist.SkippedBulkCommits,
	}
}

func rankCouplings(hist *history.Result, minCoChanges int) []Coupling {
	result := make([]Coupling, 0)
	for pair, count := range hist.CoChanges {
		if count < minCoChanges {
			continue
		}
		left := hist.Files[pair.Left]
		right := hist.Files[pair.Right]
		if left == nil || right == nil {
			continue
		}
		denominator := min(left.Commits, right.Commits)
		if denominator == 0 {
			continue
		}
		result = append(result, Coupling{
			Left: pair.Left, Right: pair.Right, CoChanges: count,
			Strength: float64(count) / float64(denominator),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Strength != result[j].Strength {
			return result[i].Strength > result[j].Strength
		}
		if result[i].CoChanges != result[j].CoChanges {
			return result[i].CoChanges > result[j].CoChanges
		}
		if result[i].Left != result[j].Left {
			return result[i].Left < result[j].Left
		}
		return result[i].Right < result[j].Right
	})
	return result
}

func confidence(commits int) string {
	switch {
	case commits >= 100:
		return "high"
	case commits >= 30:
		return "medium"
	default:
		return "low"
	}
}

type maxima struct {
	commits      float64
	churn        float64
	recency      float64
	pageRank     float64
	betweenness  float64
	transitiveIn float64
}

func findMaxima(files map[string]*history.FileMetrics, structure map[string]structuralMetrics) maxima {
	var result maxima
	for file, metrics := range files {
		result.commits = max(result.commits, float64(metrics.Commits))
		result.churn = max(result.churn, float64(metrics.Churn()))
		result.recency = max(result.recency, metrics.RecencyScore)
		sm := structure[file]
		result.pageRank = max(result.pageRank, sm.pageRank)
		result.betweenness = max(result.betweenness, sm.betweenness)
		result.transitiveIn = max(result.transitiveIn, float64(sm.transitiveIn))
	}
	return result
}

func logNormalize(value, maximum float64) float64 {
	if value <= 0 || maximum <= 0 {
		return 0
	}
	return math.Log1p(value) / math.Log1p(maximum)
}

func linearNormalize(value, maximum float64) float64 {
	if value <= 0 || maximum <= 0 {
		return 0
	}
	return value / maximum
}

type fileGraph struct {
	adj   map[string][]string
	radj  map[string][]string
	nodes []string
}

func computeStructure(sg *ir.SymbolGraph, root string) map[string]structuralMetrics {
	graph := buildFileGraph(sg, root)
	pageRanks := pageRank(graph)
	betweenness := brandes(graph)
	result := make(map[string]structuralMetrics, len(graph.nodes))
	for _, file := range graph.nodes {
		result[file] = structuralMetrics{
			pageRank:     pageRanks[file],
			betweenness:  betweenness[file],
			transitiveIn: transitiveFanIn(file, graph.radj),
		}
	}
	return result
}

func buildFileGraph(sg *ir.SymbolGraph, root string) fileGraph {
	nodeSet := make(map[string]bool)
	idFile := make(map[string]string, len(sg.Symbols))
	for i := range sg.Symbols {
		sym := &sg.Symbols[i]
		file, err := history.RelativePath(root, sym.FilePath)
		if err != nil {
			continue
		}
		nodeSet[file] = true
		idFile[sym.ID] = file
	}

	adjSet := make(map[string]map[string]bool)
	radjSet := make(map[string]map[string]bool)
	for _, edge := range sg.Edges {
		if edge.Kind == plugin.EdgeContains || edge.Kind == plugin.EdgeImports {
			continue
		}
		from, fromOK := idFile[edge.From]
		to, toOK := idFile[edge.To]
		if !fromOK || !toOK || from == to {
			continue
		}
		if adjSet[from] == nil {
			adjSet[from] = make(map[string]bool)
		}
		if radjSet[to] == nil {
			radjSet[to] = make(map[string]bool)
		}
		adjSet[from][to] = true
		radjSet[to][from] = true
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	adj := make(map[string][]string, len(nodes))
	radj := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		adj[node] = sortedSet(adjSet[node])
		radj[node] = sortedSet(radjSet[node])
	}
	return fileGraph{nodes: nodes, adj: adj, radj: radj}
}

func sortedSet(set map[string]bool) []string {
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func pageRank(graph fileGraph) map[string]float64 {
	const (
		damping    = 0.85
		iterations = 30
	)
	ranks := make(map[string]float64, len(graph.nodes))
	if len(graph.nodes) == 0 {
		return ranks
	}
	initial := 1.0 / float64(len(graph.nodes))
	for _, node := range graph.nodes {
		ranks[node] = initial
	}
	for range iterations {
		next := make(map[string]float64, len(graph.nodes))
		dangling := 0.0
		for _, node := range graph.nodes {
			if len(graph.adj[node]) == 0 {
				dangling += ranks[node]
				continue
			}
			share := ranks[node] / float64(len(graph.adj[node]))
			for _, target := range graph.adj[node] {
				next[target] += damping * share
			}
		}
		base := (1-damping)/float64(len(graph.nodes)) + damping*dangling/float64(len(graph.nodes))
		for _, node := range graph.nodes {
			next[node] += base
		}
		ranks = next
	}
	return ranks
}

func transitiveFanIn(start string, reverse map[string][]string) int {
	seen := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range reverse[current] {
			if seen[dependent] {
				continue
			}
			seen[dependent] = true
			queue = append(queue, dependent)
		}
	}
	return len(seen) - 1
}

// brandes computes normalized directed betweenness centrality.
func brandes(graph fileGraph) map[string]float64 {
	result := make(map[string]float64, len(graph.nodes))
	for _, source := range graph.nodes {
		var stack []string
		pred := make(map[string][]string, len(graph.nodes))
		sigma := make(map[string]float64, len(graph.nodes))
		distance := make(map[string]int, len(graph.nodes))
		for _, node := range graph.nodes {
			distance[node] = -1
		}
		sigma[source] = 1
		distance[source] = 0
		queue := []string{source}
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)
			for _, w := range graph.adj[v] {
				if distance[w] < 0 {
					distance[w] = distance[v] + 1
					queue = append(queue, w)
				}
				if distance[w] == distance[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}
		delta := make(map[string]float64, len(graph.nodes))
		for len(stack) > 0 {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, v := range pred[w] {
				if sigma[w] != 0 {
					delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
				}
			}
			if w != source {
				result[w] += delta[w]
			}
		}
	}
	if len(graph.nodes) > 2 {
		scale := 1 / float64((len(graph.nodes)-1)*(len(graph.nodes)-2))
		for node := range result {
			result[node] *= scale
		}
	}
	return result
}
