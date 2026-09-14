// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

func accuracyGraph(names []string, edges []plugin.Edge) *ir.SymbolGraph {
	sg := &ir.SymbolGraph{Edges: edges}
	for _, name := range names {
		sg.Symbols = append(sg.Symbols, plugin.Symbol{
			ID: name, Name: name, FilePath: "app.go", Category: plugin.CategoryCallable,
		})
	}
	sg.BuildIndexes()
	return sg
}

func TestGraphAccuracyReportIsDeterministic(t *testing.T) {
	sg := &ir.SymbolGraph{}
	for i := range 18 {
		name := fmt.Sprintf("node%02d", i)
		category := plugin.CategoryCallable
		if i%3 == 0 {
			category = plugin.CategoryType
		}
		sg.Symbols = append(sg.Symbols, plugin.Symbol{
			ID: name, Name: name, FilePath: fmt.Sprintf("pkg%d/file.go", i%3), Category: category, Kind: "class",
		})
		for j := range 18 {
			if i != j && (i+j)%4 == 0 {
				sg.Edges = append(sg.Edges, plugin.Edge{From: name, To: fmt.Sprintf("node%02d", j), Kind: plugin.EdgeCalls})
			}
		}
	}
	sg.BuildIndexes()
	opts := DefaultAnalysisOptions()
	opts.TopN = 0
	want := renderAnalysis(runAnalysis(sg, opts), opts)
	for range 15 {
		if got := renderAnalysis(runAnalysis(sg, opts), opts); got != want {
			wantLines, gotLines := strings.Split(want, "\n"), strings.Split(got, "\n")
			for i := range min(len(wantLines), len(gotLines)) {
				if wantLines[i] != gotLines[i] {
					t.Fatalf("non-deterministic report line %d: %q vs %q", i+1, wantLines[i], gotLines[i])
				}
			}
			t.Fatal("identical graph produced reports of different lengths")
		}
	}
}

func TestGraphAccuracyUncertaintyCannotBecomeAnEdge(t *testing.T) {
	sg := accuracyGraph([]string{"a", "b", "c"}, []plugin.Edge{
		{From: "a", To: "b", Kind: plugin.EdgeCalls, Resolution: "ambiguous", Candidates: []string{"b", "c"}},
	})
	g := buildGraphCtx(sg)
	if g.fanOut["S1"] != 0 || g.ambiguousEdges != 1 || g.resolvedEdges != 0 {
		t.Fatalf("ambiguous endpoint was treated as a dependency: %+v", g)
	}
	edges := buildJSONEdges(sg)
	if len(edges) != 1 || edges[0].ToShort != "" || !slices.Equal(edges[0].Candidates, []string{"b", "c"}) {
		t.Fatalf("JSON lost ambiguity: %+v", edges)
	}
}

func TestGraphAccuracyUnresolvedSKTRoundTrip(t *testing.T) {
	sg := accuracyGraph([]string{"a"}, []plugin.Edge{{From: "a", To: "external::method", Kind: plugin.EdgeCalls}})
	sg.Symbols[0].Kind, sg.Symbols[0].Signature, sg.Symbols[0].Span = "function", "a()", [2]int{1, 1}
	opts := &EmitOptions{}
	text := emitInlineString(t, sg, opts)
	got, err := parser.ParseOutput([]io.Reader{strings.NewReader(text)}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Edges) != 1 || got.Edges[0].To != "external::method" || got.Edges[0].Resolution != "unresolved" {
		t.Fatalf("uncertainty lost on SKT round trip: %+v", got.Edges)
	}
}

func TestGraphAccuracyFileWithoutSymbolsKeepsRelationships(t *testing.T) {
	sg := &ir.SymbolGraph{Edges: []plugin.Edge{
		{From: "header.h::header.h", To: "header.h::external.h", Kind: plugin.EdgeReferences},
	}}
	sg.BuildIndexes()
	text := emitInlineString(t, sg, &EmitOptions{})
	if !strings.Contains(text, `[unresolved]--> external.h`) {
		t.Fatalf("file without symbols lost its relationship: %s", text)
	}
}

func TestGraphAccuracyPropagationUsesStrongestPath(t *testing.T) {
	sg := accuracyGraph([]string{"a", "b", "c", "d"}, []plugin.Edge{
		{From: "b", To: "a", Kind: plugin.EdgeReferences}, // .3
		{From: "c", To: "a", Kind: plugin.EdgeInherits},   // .9
		{From: "b", To: "c", Kind: plugin.EdgeInherits},   // .81 reaches b
		{From: "d", To: "b", Kind: plugin.EdgeInherits},   // .729 reaches d
	})
	opts := DefaultAnalysisOptions()
	opts.PropagationCutoff = .5
	for range 2 {
		entries := computeChangePropagation(buildGraphCtx(sg), opts)
		var found bool
		for _, e := range entries {
			if e.SID == sg.ShortIDs["a"] {
				found = true
				if e.BlastRadius != 3 || math.Abs(e.Score-(.9+.81+.729)/3) > 1e-9 {
					t.Fatalf("wrong strongest-path score: %+v", e)
				}
			}
		}
		if !found {
			t.Fatal("missing propagation result")
		}
		slices.Reverse(sg.Edges)
	}
}

func TestGraphAccuracyInheritanceCyclesHaveNoDepth(t *testing.T) {
	if got := detectDeepInheritance(map[string][]string{"a": {"b"}, "b": {"a"}}, 1); len(got) != 0 {
		t.Fatalf("invented depth for a cycle: %+v", got)
	}
}

func TestGraphAccuracyMultipleInheritance(t *testing.T) {
	sg := accuracyGraph([]string{"Root", "Base", "Child", "Other"}, []plugin.Edge{
		{From: "Base", To: "Root", Kind: plugin.EdgeInherits},
		{From: "Child", To: "Base", Kind: plugin.EdgeInherits},
		{From: "Child", To: "Other", Kind: plugin.EdgeImplements},
	})
	chains := detectDeepInheritance(buildGraphCtx(sg).inheritEdges, 2)
	if len(chains) != 1 || chains[0].Leaf != sg.ShortIDs["Child"] || chains[0].Depth != 2 {
		t.Fatalf("lost longest inheritance path: %+v", chains)
	}
}

func TestGraphAccuracyReachableMemberDoesNotReachSiblings(t *testing.T) {
	sg := accuracyGraph([]string{"main", "Container", "used", "unused"}, []plugin.Edge{
		{From: "main", To: "used", Kind: plugin.EdgeCalls},
		{From: "Container", To: "used", Kind: plugin.EdgeContains},
		{From: "Container", To: "unused", Kind: plugin.EdgeContains},
	})
	unreachable := findUnreachable(buildGraphCtx(sg))
	if len(unreachable) != 1 || unreachable[0].Name != "unused" {
		t.Fatalf("expected unused sibling to remain unreachable: %+v", unreachable)
	}
}

func TestGraphAccuracyUniqueNeighbors(t *testing.T) {
	sg := accuracyGraph([]string{"a", "b"}, []plugin.Edge{
		{From: "a", To: "b", Kind: plugin.EdgeCalls},
		{From: "a", To: "b", Kind: plugin.EdgeCalls},
		{From: "a", To: "b", Kind: plugin.EdgeReferences},
	})
	g := buildGraphCtx(sg)
	if g.fanOut["S1"] != 1 || g.fanIn["S2"] != 1 {
		t.Fatalf("duplicate relations inflate degree: out=%v in=%v", g.fanOut, g.fanIn)
	}
}

func TestGraphAccuracyJSONPreservesUnresolved(t *testing.T) {
	sg := accuracyGraph([]string{"a"}, []plugin.Edge{
		{From: "a", To: "external", Kind: plugin.EdgeCalls},
	})
	if got := buildJSONEdges(sg); len(got) != 1 || got[0].To != "external" {
		t.Fatalf("unresolved relationship disappeared: %+v", got)
	}
}
