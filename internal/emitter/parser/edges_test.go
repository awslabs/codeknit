// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"io"
	"slices"
	"strings"
	"testing"
)

func TestUncertainEdgesResolveCandidatesFromLaterChunks(t *testing.T) {
	first := `[symbols]
## app.go
S1 callable/function L1-L1 Run()
[edges]
S1 --calls[ambiguous]--> Helper [candidates=S2, S3]
S1 --references[unresolved]--> "app.go::S2"
S1 --references--> S2
S1 --references[unresolved]--> S2
`
	second := `[symbols]
## helpers.go
S2 callable/function L1-L1 Helper()
S3 callable/function L2-L2 Helper(x: int)
`
	sg, err := ParseOutput([]io.Reader{strings.NewReader(first), strings.NewReader(second)}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(sg.Edges) != 4 {
		t.Fatalf("lost edges: %+v", sg.Edges)
	}
	if !slices.Equal(sg.Edges[0].Candidates, []string{"helpers.go::Helper", "helpers.go::Helper#1"}) {
		t.Errorf("candidate identities lost: %+v", sg.Edges[0])
	}
	if sg.Edges[1].To != "app.go::S2" || sg.Edges[1].Resolution != "unresolved" {
		t.Errorf("unresolved name mistaken for a symbol ID: %+v", sg.Edges[1])
	}
	if sg.Edges[2].To != "helpers.go::Helper" || sg.Edges[2].Resolution != "" {
		t.Errorf("resolved forward reference lost: %+v", sg.Edges[2])
	}
	if sg.Edges[3].To != "helpers.go::Helper" || sg.Edges[3].Resolution != "unresolved" {
		t.Errorf("uncertain forward reference lost: %+v", sg.Edges[3])
	}
}

func TestRejectMalformedEdgeAnnotations(t *testing.T) {
	for _, line := range []string{
		`S1 --calls[ambiguous]--> Foo [candidates=]`,
		`S1 --calls--> S2 [candidates=S2]`,
		`S1 --calls[ambiguous]--> "unterminated`,
		`S1 --calls[unresolved]--> "bad\q"`,
		`S1 --calls[ambiguous]--> Foo,`,
		`S1 --calls[ambiguous]--> Foo [candidates=S2] trailing`,
	} {
		t.Run(line, func(t *testing.T) {
			if _, err := parseEdgeLine(line, nil, nil, false); err == nil {
				t.Fatal("malformed edge accepted")
			}
		})
	}
}
