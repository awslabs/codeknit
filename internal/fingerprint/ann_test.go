// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"testing"
)

func TestBuildANNIndex_EmptyInput(t *testing.T) {
	idx := BuildANNIndex(nil)
	candidates := idx.FindCandidates(10)
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates for empty input, got %d", len(candidates))
	}
}

func TestBuildANNIndex_SingleEntry(t *testing.T) {
	idx := BuildANNIndex([][]byte{{0x01, 0x02, 0x03, 0x04}})
	candidates := idx.FindCandidates(10)
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates for single entry, got %d", len(candidates))
	}
}

func TestBuildANNIndex_IdenticalStreamsAreTopCandidate(t *testing.T) {
	stream := []byte{0x01, 0x03, 0x05, 0x12, 0x01, 0x03, 0x05, 0x12}
	different := []byte{0x0A, 0x0B, 0x0C, 0x38, 0x39, 0x3A, 0x3B, 0x3C}

	idx := BuildANNIndex([][]byte{stream, stream, different})
	candidates := idx.FindCandidates(10)

	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}

	// The top candidate should be the identical pair (0, 1).
	top := candidates[0]
	if (top.I != 0 || top.J != 1) && (top.I != 1 || top.J != 0) {
		t.Errorf("expected top candidate to be (0,1), got (%d,%d)", top.I, top.J)
	}
	if top.CosineSim < 0.99 {
		t.Errorf("identical streams should have cosine ~1.0, got %.4f", top.CosineSim)
	}
}

func TestBuildANNIndex_CandidatesAreDeduplicated(t *testing.T) {
	streams := make([][]byte, 5)
	for i := range streams {
		streams[i] = []byte{0x01, 0x03, 0x05, byte(i), 0x12, 0x13}
	}

	idx := BuildANNIndex(streams)
	candidates := idx.FindCandidates(10)

	// Check no duplicate (i,j) pairs.
	type pair struct{ i, j int }
	seen := make(map[pair]bool)
	for _, c := range candidates {
		p := pair{c.I, c.J}
		if c.I > c.J {
			p = pair{c.J, c.I}
		}
		if seen[p] {
			t.Errorf("duplicate candidate pair (%d, %d)", p.i, p.j)
		}
		seen[p] = true
	}
}

func TestDefaultANNK(t *testing.T) {
	if k := DefaultANNK(); k != annK {
		t.Errorf("DefaultANNK() = %d, want %d", k, annK)
	}
}
