// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"io"
	"slices"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"
	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

func TestSKTEdgeResolutionSyntax(t *testing.T) {
	sg := &ir.SymbolGraph{
		Symbols: []plugin.Symbol{
			{ID: "api.go::Run", Name: "Run", Signature: "Run()", FilePath: "api.go", Category: plugin.CategoryCallable, Kind: "function", Span: [2]int{1, 1}},
			{ID: "api.go::Helper", Name: "Helper", Signature: "Helper()", FilePath: "api.go", Category: plugin.CategoryCallable, Kind: "function", Span: [2]int{2, 2}},
			{ID: "api.go::Helper#1", Name: "Helper", Signature: "Helper(x: int)", FilePath: "api.go", Category: plugin.CategoryCallable, Kind: "function", Span: [2]int{3, 3}},
		},
		Edges: []plugin.Edge{
			{From: "api.go::Run", To: "api.go::Helper", Kind: plugin.EdgeReferences},
			{From: "api.go::Run", To: "api.go::string", Kind: plugin.EdgeReferences, Resolution: "unresolved"},
			{From: "api.go::Run", To: "api.go::bool", Kind: plugin.EdgeReferences, Resolution: "unresolved"},
			{From: "api.go::Run", To: "api.go::Helper", Kind: plugin.EdgeCalls, Resolution: "ambiguous", Candidates: []string{"api.go::Helper", "api.go::Helper#1"}},
			{From: "api.go::Helper", To: "api.go::S2", Kind: plugin.EdgeReferences, Resolution: "unresolved"},
		},
	}
	sg.BuildIndexes()
	for _, minify := range []bool{false, true} {
		t.Run(map[bool]string{false: "plain", true: "minified"}[minify], func(t *testing.T) {
			output := emitInlineString(t, sg, &EmitOptions{Minify: minify})
			references, calls := "references", "calls"
			if minify {
				dict := NewDictionary(sg)
				references, calls = dict.Encode(references), dict.Encode(calls)
			}
			for _, want := range []string{
				"S1 --" + references + "--> S2",
				"S1 --" + references + "[unresolved]--> string, bool",
				"S1 --" + calls + "[ambiguous]--> Helper [candidates=S2, S3]",
				`S2 --` + references + `[unresolved]--> "api.go::S2"`,
			} {
				if !strings.Contains(output, want) {
					t.Errorf("missing %q:\n%s", want, output)
				}
			}
			if strings.Contains(output, "? {") {
				t.Errorf("JSON record leaked into SKT:\n%s", output)
			}
			parsed, err := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, minify)
			if err != nil {
				t.Fatal(err)
			}
			if len(parsed.Edges) != len(sg.Edges) {
				t.Fatalf("lost grouped edges: %+v", parsed.Edges)
			}
			for i, want := range sg.Edges {
				got := parsed.Edges[i]
				if got.From != want.From || got.To != want.To || got.Kind != want.Kind ||
					got.Resolution != want.Resolution || !slices.Equal(got.Candidates, want.Candidates) {
					t.Errorf("edge changed on round trip: got %+v, want %+v", got, want)
				}
			}
		})
	}
}

func TestSKTUncertainEdgeEscaping(t *testing.T) {
	source := "folder with spaces/api.go::missing"
	target := "folder with spaces/api.go::call(\"a, b\")\n[candidates=S9]"
	sg := &ir.SymbolGraph{Edges: []plugin.Edge{
		{From: source, To: target, Kind: plugin.EdgeCalls, Resolution: "ambiguous", Candidates: []string{"elsewhere.go::Other"}},
	}}
	sg.BuildIndexes()
	output := emitInlineString(t, sg, &EmitOptions{})
	parsed, err := parser.ParseOutput([]io.Reader{strings.NewReader(output)}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Edges) != 1 || parsed.Edges[0].From != source || parsed.Edges[0].To != target ||
		!slices.Equal(parsed.Edges[0].Candidates, sg.Edges[0].Candidates) {
		t.Fatalf("quoted edge did not round trip: %+v\n%s", parsed.Edges, output)
	}
}
