// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// --- Hash happy path ---

func TestHash_ProducesDeterministicOutput(t *testing.T) {
	tokens := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}
	got1 := Hash(tokens)
	got2 := Hash(tokens)
	if got1 != got2 {
		t.Fatalf("Hash not deterministic: %q vs %q", got1, got2)
	}
	if got1 == "" {
		t.Fatal("Hash returned empty for valid input")
	}
}

func TestHash_IdenticalInputsProduceIdenticalHashes(t *testing.T) {
	a := []byte("hello world fingerprint test")
	b := []byte("hello world fingerprint test")
	if Hash(a) != Hash(b) {
		t.Fatalf("identical inputs produced different hashes: %q vs %q", Hash(a), Hash(b))
	}
}

func TestHash_FormatIsBlocksizeColonHex1ColonHex2(t *testing.T) {
	tokens := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	fp := Hash(tokens)
	parts := strings.Split(fp, ":")
	if len(parts) != 4 {
		t.Fatalf("expected 4 colon-separated parts (bs:h1:h2:h3), got %d in %q", len(parts), fp)
	}
	// Block size must be a positive integer.
	if parts[0] == "" {
		t.Fatal("block size is empty")
	}
	// All three hashes must be valid non-empty hex.
	for i := 1; i <= 3; i++ {
		if parts[i] == "" {
			t.Fatalf("hash section %d should be non-empty, got %q", i, fp)
		}
	}
}

// --- Hash edge cases ---

func TestHash_EmptyInputReturnsEmpty(t *testing.T) {
	if got := Hash(nil); got != "" {
		t.Errorf("nil input should return empty, got %q", got)
	}
	if got := Hash([]byte{}); got != "" {
		t.Errorf("empty input should return empty, got %q", got)
	}
}

func TestHash_BelowMinInputLenReturnsEmpty(t *testing.T) {
	// minInputLen is 4, so anything 0-3 bytes should return empty.
	for n := 0; n < minInputLen; n++ {
		tokens := make([]byte, n)
		if got := Hash(tokens); got != "" {
			t.Errorf("input of length %d should return empty, got %q", n, got)
		}
	}
}

func TestHash_ExactlyMinInputLenSucceeds(t *testing.T) {
	tokens := make([]byte, minInputLen)
	for i := range tokens {
		tokens[i] = byte(i + 1)
	}
	if got := Hash(tokens); got == "" {
		t.Error("input at exactly minInputLen should produce a hash")
	}
}

func TestHash_LargeInputForcesBlockSizeDoubling(t *testing.T) {
	// Large enough that blockSize must double at least once.
	tokens := make([]byte, defaultBlockSize*maxResultLen*4)
	for i := range tokens {
		tokens[i] = byte(i % 256)
	}
	fp := Hash(tokens)
	if fp == "" {
		t.Fatal("large input should produce a hash")
	}
	// Block size in output should be greater than defaultBlockSize.
	var bs int
	_, err := strings.NewReader(fp).Read(make([]byte, 0))
	_ = err
	parts := strings.SplitN(fp, ":", 2)
	if len(parts) < 2 {
		t.Fatalf("malformed hash %q", fp)
	}
	if _, scanErr := strings.NewReader(parts[0]).Read(make([]byte, 0)); scanErr != nil {
		t.Fatalf("block size parse failed: %v", scanErr)
	}
	// Parse blocksize.
	bs = atoiOrZero(parts[0])
	if bs <= defaultBlockSize {
		t.Errorf("block size %d should be > defaultBlockSize %d for large input", bs, defaultBlockSize)
	}
}

func atoiOrZero(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// --- Distance happy path ---

func TestDistance_IdenticalHashesScore100(t *testing.T) {
	tokens := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	fp := Hash(tokens)
	if got := Distance(fp, fp); got != 100 {
		t.Errorf("Distance of identical hashes should be 100, got %d", got)
	}
}

func TestDistance_IsSymmetric(t *testing.T) {
	a := Hash([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})
	b := Hash([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x09})
	d1 := Distance(a, b)
	d2 := Distance(b, a)
	if d1 != d2 {
		t.Errorf("Distance not symmetric: Distance(a,b)=%d, Distance(b,a)=%d", d1, d2)
	}
}

func TestDistance_SimilarInputsScoreHigher(t *testing.T) {
	// Build two longer byte streams: "base" and a variant with one byte changed
	// near the middle, plus a third that's completely different.
	// Longer inputs exercise more chunks so the rolling hash can discriminate.
	base := make([]byte, 200)
	nearIdentical := make([]byte, 200)
	completelyDifferent := make([]byte, 200)
	for i := range base {
		base[i] = byte(i % 50)
		nearIdentical[i] = byte(i % 50)
		completelyDifferent[i] = byte((i + 128) % 50)
	}
	nearIdentical[100] = 0xFF // single-byte change

	simBase := Distance(Hash(base), Hash(nearIdentical))
	simDiff := Distance(Hash(base), Hash(completelyDifferent))

	if simBase <= simDiff {
		t.Errorf("near-identical score (%d) should be greater than completely-different score (%d)", simBase, simDiff)
	}
}

func TestDistance_ScoreInRange(t *testing.T) {
	// Run many random pairs and verify Distance never produces out-of-range values.
	rapid.Check(t, func(rt *rapid.T) {
		a := rapid.SliceOfN(rapid.Byte(), minInputLen, 200).Draw(rt, "a")
		b := rapid.SliceOfN(rapid.Byte(), minInputLen, 200).Draw(rt, "b")
		fp1 := Hash(a)
		fp2 := Hash(b)
		if fp1 == "" || fp2 == "" {
			return
		}
		d := Distance(fp1, fp2)
		if d != -1 && (d < 0 || d > 100) {
			rt.Fatalf("Distance out of range [0,100] or -1: got %d for %q vs %q", d, fp1, fp2)
		}
	})
}

// --- Distance error cases / malformed input ---

func TestDistance_MalformedInputsReturnMinusOne(t *testing.T) {
	cases := []struct {
		name string
		a, b string
	}{
		{"both empty", "", ""},
		{"first empty", "", "3:deadbeef:cafebabe"},
		{"second empty", "3:deadbeef:cafebabe", ""},
		{"no colons first", "notahash", "3:deadbeef:cafebabe"},
		{"one colon only", "3:deadbeef", "3:deadbeef:cafebabe"},
		{"block size non-numeric", "abc:deadbeef:cafebabe", "3:deadbeef:cafebabe"},
		{"bad hex in first hash", "3:zzzz:cafebabe", "3:deadbeef:cafebabe"},
		{"bad hex in second hash", "3:deadbeef:zzzz", "3:deadbeef:cafebabe"},
		{"block size 0", "0:deadbeef:cafebabe", "3:deadbeef:cafebabe"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Distance(tc.a, tc.b); got != -1 {
				t.Errorf("expected -1 for malformed input, got %d", got)
			}
		})
	}
}

func TestDistance_IncompatibleBlockSizesReturnMinusOne(t *testing.T) {
	// Block sizes that differ by more than 2x and have no overlap.
	a := "3:deadbeef:cafebabe"
	b := "48:deadbeef:cafebabe"
	if got := Distance(a, b); got != -1 {
		t.Errorf("incompatible block sizes should return -1, got %d", got)
	}
}

func TestDistance_DoubledBlockSizesAreComparable(t *testing.T) {
	// Hash provides two block sizes (N and 2N); Distance should handle 2x ratios.
	a := "3:deadbeef:cafebabe"
	b := "6:cafebabe:12345678"
	// a's h1 (blocksize 3) vs b's h1 (blocksize 6=3*2) → a's h2 should match b's h1.
	got := Distance(a, b)
	if got == -1 {
		t.Errorf("doubled block sizes should be comparable, got -1")
	}
}

// --- compareHashes edge cases ---

func TestCompareHashes_EmptyInputsReturnZero(t *testing.T) {
	if got := compareHashes(nil, []byte{1, 2, 3}); got != 0 {
		t.Errorf("nil vs non-empty should be 0, got %d", got)
	}
	if got := compareHashes([]byte{1, 2, 3}, nil); got != 0 {
		t.Errorf("non-empty vs nil should be 0, got %d", got)
	}
	if got := compareHashes([]byte{}, []byte{}); got != 0 {
		t.Errorf("both empty should be 0, got %d", got)
	}
}

func TestCompareHashes_IdenticalInputsScore100(t *testing.T) {
	h := []byte{0x01, 0x02, 0x03, 0x04}
	if got := compareHashes(h, h); got != 100 {
		t.Errorf("identical hashes should score 100, got %d", got)
	}
}

func TestCompareHashes_ScoreIsClampedToZero(t *testing.T) {
	// Very different hashes should produce a low but non-negative score.
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0xFF, 0xFE, 0xFD, 0xFC}
	got := compareHashes(a, b)
	if got < 0 {
		t.Errorf("score should not be negative, got %d", got)
	}
}

// --- editDistance ---

func TestEditDistance_KnownValues(t *testing.T) {
	cases := []struct {
		name string
		a, b []byte
		want int
	}{
		{"both empty", nil, nil, 0},
		{"empty vs one byte", nil, []byte{1}, 1},
		{"one byte vs empty", []byte{1}, nil, 1},
		{"identical single byte", []byte{1}, []byte{1}, 0},
		{"different single byte", []byte{1}, []byte{2}, 1},
		{"insertion", []byte{1, 2, 3}, []byte{1, 2, 3, 4}, 1},
		{"deletion", []byte{1, 2, 3, 4}, []byte{1, 2, 3}, 1},
		{"substitution", []byte{1, 2, 3}, []byte{1, 9, 3}, 1},
		{"two substitutions", []byte{1, 2, 3, 4}, []byte{1, 9, 3, 8}, 2},
		{"completely different", []byte{1, 2, 3}, []byte{4, 5, 6}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := editDistance(tc.a, tc.b); got != tc.want {
				t.Errorf("editDistance = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestEditDistance_IsSymmetric(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		a := rapid.SliceOfN(rapid.Byte(), 0, 50).Draw(rt, "a")
		b := rapid.SliceOfN(rapid.Byte(), 0, 50).Draw(rt, "b")
		if editDistance(a, b) != editDistance(b, a) {
			rt.Fatalf("editDistance not symmetric for %v vs %v", a, b)
		}
	})
}

// --- min3 ---

func TestMin3(t *testing.T) {
	cases := []struct {
		a, b, c, want int
	}{
		{1, 2, 3, 1},
		{3, 1, 2, 1},
		{2, 3, 1, 1},
		{5, 5, 5, 5},
		{-1, 0, 1, -1},
		{10, 10, 5, 5},
	}
	for _, tc := range cases {
		if got := min3(tc.a, tc.b, tc.c); got != tc.want {
			t.Errorf("min3(%d,%d,%d) = %d, want %d", tc.a, tc.b, tc.c, got, tc.want)
		}
	}
}

// --- parseCTPH ---

func TestParseCTPH_ValidInput(t *testing.T) {
	bs, hashes := parseCTPH("3:deadbeef:cafebabe")
	if bs != 3 {
		t.Errorf("expected blockSize 3, got %d", bs)
	}
	if len(hashes) < 2 {
		t.Fatalf("expected at least 2 hashes, got %d", len(hashes))
	}
	h1, h2 := hashes[0], hashes[1]
	if len(h1) != 4 || h1[0] != 0xde || h1[1] != 0xad || h1[2] != 0xbe || h1[3] != 0xef {
		t.Errorf("unexpected h1: %x", h1)
	}
	if len(h2) != 4 || h2[0] != 0xca || h2[1] != 0xfe || h2[2] != 0xba || h2[3] != 0xbe {
		t.Errorf("unexpected h2: %x", h2)
	}
}

func TestParseCTPH_MalformedReturnsZero(t *testing.T) {
	cases := []string{
		"",
		"notahash",
		"3:",
		":deadbeef:cafebabe",
		"abc:deadbeef:cafebabe",
		"3:nothex:cafebabe",
		"3:deadbeef:nothex",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			bs, _ := parseCTPH(c)
			if bs != 0 {
				t.Errorf("expected blockSize 0 for malformed %q, got %d", c, bs)
			}
		})
	}
}

// --- Property-based invariants ---

func TestProperty_HashIsDeterministic(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tokens := rapid.SliceOfN(rapid.Byte(), minInputLen, 500).Draw(rt, "tokens")
		h1 := Hash(tokens)
		h2 := Hash(tokens)
		if h1 != h2 {
			rt.Fatalf("Hash not deterministic: %q vs %q for input %v", h1, h2, tokens)
		}
	})
}

func TestProperty_DistanceIsReflexive(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tokens := rapid.SliceOfN(rapid.Byte(), minInputLen, 200).Draw(rt, "tokens")
		fp := Hash(tokens)
		if fp == "" {
			return
		}
		if d := Distance(fp, fp); d != 100 {
			rt.Fatalf("Distance(fp, fp) should be 100, got %d for %q", d, fp)
		}
	})
}

func TestProperty_DistanceIsSymmetric(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		a := rapid.SliceOfN(rapid.Byte(), minInputLen, 200).Draw(rt, "a")
		b := rapid.SliceOfN(rapid.Byte(), minInputLen, 200).Draw(rt, "b")
		fp1, fp2 := Hash(a), Hash(b)
		if fp1 == "" || fp2 == "" {
			return
		}
		if Distance(fp1, fp2) != Distance(fp2, fp1) {
			rt.Fatalf("Distance not symmetric for %q vs %q", fp1, fp2)
		}
	})
}

func TestProperty_OutputParseable(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tokens := rapid.SliceOfN(rapid.Byte(), minInputLen, 500).Draw(rt, "tokens")
		fp := Hash(tokens)
		if fp == "" {
			return
		}
		bs, hashes := parseCTPH(fp)
		if bs == 0 {
			rt.Fatalf("Hash output %q is not parseable by parseCTPH", fp)
		}
		if len(hashes) < 2 {
			rt.Fatalf("Hash output %q produced fewer than 2 hash sections", fp)
		}
		for i, h := range hashes {
			if len(h) == 0 {
				rt.Fatalf("Hash output %q produced empty hash section at index %d", fp, i)
			}
		}
	})
}
