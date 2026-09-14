// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"
	"codeknit/internal/planner"
	"codeknit/internal/plugin"
	"codeknit/internal/plugin/clang"
	"codeknit/internal/plugin/cpp"
)

func TestIncludeRelationshipsUseSymbols(t *testing.T) {
	for _, tc := range []struct {
		parser plugin.LanguagePlugin
		name   string
	}{
		{clang.NewPlugin(), "main.c"},
		{cpp.NewPlugin(), "main.cpp"},
		{clang.NewPlugin(), "only-includes.h"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "#if ENABLED\n" +
				"#include \"config.h\"\n#include \"dir/with space, (test).h\"\n" +
				"#include <stdio.h>\n#include \"config.h\"\n#endif\n"
			path := filepath.Join(t.TempDir(), tc.name)
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}
			symbols, edges, err := tc.parser.Parse(path)
			if err != nil {
				t.Fatal(err)
			}
			fp := "src/" + tc.name
			sg := (&planner.Planner{}).Plan(
				map[string][]plugin.Symbol{fp: symbols},
				map[string][]plugin.Edge{fp: edges},
			)
			var fileCount, includeCount int
			for _, sym := range sg.Symbols {
				switch sym.Kind {
				case "file":
					fileCount++
					if sym.Name != tc.name {
						t.Errorf("file endpoint repeats its path: %+v", sym)
					}
				case "include":
					includeCount++
					if sym.Properties["resolution"] != "unresolved" {
						t.Errorf("include appears to be a resolved header: %+v", sym)
					}
					if sym.Span != [2]int{includeCount + 1, includeCount + 1} {
						t.Errorf("include span extends into the next line: %+v", sym)
					}
				}
			}
			if fileCount != 1 || includeCount != 3 || len(sg.Edges) != 3 {
				t.Fatalf("include endpoints missing or duplicated: files=%d includes=%d edges=%+v", fileCount, includeCount, sg.Edges)
			}
			for _, edge := range sg.Edges {
				if sg.ShortIDs[edge.From] == "" || sg.ShortIDs[edge.To] == "" || edge.Resolution != "unresolved" {
					t.Fatalf("include endpoints must have IDs without claiming resolution: %+v", edge)
				}
			}
			g := buildGraphCtx(sg)
			if len(g.allSIDs) != 0 || g.unresolvedEdges != 3 || g.resolvedEdges != 0 {
				t.Fatalf("file/include metadata affected dependency metrics: %+v", g)
			}
			for _, minify := range []bool{false, true} {
				output := emitInlineString(t, sg, &EmitOptions{Minify: minify})
				kind := "references"
				if minify {
					kind = NewDictionary(sg).Encode(kind)
				}
				want := sg.ShortIDs[sg.Edges[0].From] + " --" + kind + "[unresolved]--> " +
					sg.ShortIDs[sg.Edges[0].To] + ", " + sg.ShortIDs[sg.Edges[1].To] + ", " + sg.ShortIDs[sg.Edges[2].To]
				if !strings.Contains(output, want) || strings.Contains(output, "::") {
					t.Fatalf("include edges must use short IDs:\n%s", output)
				}
				parsed, parseErr := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, minify)
				if parseErr != nil {
					t.Fatal(parseErr)
				}
				if len(parsed.Edges) != len(sg.Edges) {
					t.Fatalf("include edges lost on round trip: %+v", parsed.Edges)
				}
				for i, edge := range parsed.Edges {
					wantEdge := sg.Edges[i]
					if edge.From != wantEdge.From || edge.To != wantEdge.To || edge.Resolution != wantEdge.Resolution {
						t.Errorf("include endpoint changed on round trip: got %+v, want %+v", edge, wantEdge)
					}
				}
			}
		})
	}
}
