// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// minInputLen is the minimum token stream length required to produce
// a meaningful fingerprint. Shorter streams are too small for fuzzy matching.
const minInputLen = 4

// defaultBlockSize is the starting block size for CTPH chunking.
// The algorithm doubles this when the output gets too long.
const defaultBlockSize = 2

// maxResultLen caps the number of hash characters per block size.
const maxResultLen = 64

// rollingPrime is the prime used in the rolling hash (Adler-like).
const rollingPrime uint32 = 37

// Hash computes a multi-resolution CTPH fuzzy hash of a normalized token
// stream. The result encodes three block-size levels (bs, bs*2, bs*4) so
// that Distance can compare at more resolutions, reducing sensitivity to
// single-token insertions that shift chunk boundaries.
//
// Format: "blockSize:hex1:hex2:hex3"
//
// Returns "" if the input is too short for meaningful fingerprinting.
func Hash(tokens []byte) string {
	if len(tokens) < minInputLen {
		return ""
	}

	// Find appropriate block size: start small, double until the output
	// fits within maxResultLen characters.
	blockSize := defaultBlockSize
	for blockSize*maxResultLen < len(tokens) {
		blockSize *= 2
	}

	h1 := computeCTPH(tokens, blockSize)
	h2 := computeCTPH(tokens, blockSize*2)
	h3 := computeCTPH(tokens, blockSize*4)

	return fmt.Sprintf("%d:%s:%s:%s",
		blockSize,
		hex.EncodeToString(h1),
		hex.EncodeToString(h2),
		hex.EncodeToString(h3))
}

// computeCTPH runs the core CTPH algorithm at a given block size.
// It uses a rolling hash to determine chunk boundaries, then produces
// a hash byte for each chunk using FNV-like mixing.
func computeCTPH(tokens []byte, blockSize int) []byte {
	if len(tokens) == 0 {
		return nil
	}

	var (
		rolling uint32 // rolling hash for boundary detection
		chunk   uint32 // FNV-like hash for current chunk content
		result  []byte
	)

	chunk = 2166136261 // FNV offset basis

	// #nosec G115 -- blockSize is always a small positive int (≤ 2^k * defaultBlockSize)
	// bounded by the doubling loop in Hash, safely fits in uint32.
	bs := uint32(blockSize)
	for i, tok := range tokens {
		// Update rolling hash (sliding window).
		rolling = rolling*rollingPrime + uint32(tok)

		// Update chunk hash (FNV-1a style).
		chunk ^= uint32(tok)
		chunk *= 16777619 // FNV prime

		// Check if we hit a chunk boundary.
		if rolling%bs == bs-1 || i == len(tokens)-1 {
			// Emit one byte from the chunk hash (truncation is intentional).
			result = append(result, byte(chunk&0xFF))
			chunk = 2166136261 // reset for next chunk
			rolling = 0

			if len(result) >= maxResultLen {
				break
			}
		}
	}

	return result
}

// Distance computes the edit distance between two CTPH fingerprints,
// returning a similarity score from 0 (completely different) to 100 (identical).
// Returns -1 if the fingerprints cannot be compared (different block sizes
// with no overlap).
//
// With multi-resolution hashes (3 levels), Distance tries all compatible
// block-size pairs and returns the highest score, reducing sensitivity to
// chunk-boundary shifts caused by small insertions/deletions.
func Distance(fp1, fp2 string) int {
	bs1, hashes1 := parseCTPH(fp1)
	bs2, hashes2 := parseCTPH(fp2)

	if bs1 == 0 || bs2 == 0 {
		return -1
	}

	// Try all compatible block-size pairs and return the best score.
	// hashes1[k] is at block size bs1 * 2^k, hashes2[k] at bs2 * 2^k.
	best := -1
	for i, h1 := range hashes1 {
		for j, h2 := range hashes2 {
			// Block sizes for these levels.
			bsi := bs1 << uint(i) //nolint:gosec // i is 0..2, safe shift
			bsj := bs2 << uint(j) //nolint:gosec // j is 0..2, safe shift
			if bsi != bsj {
				continue
			}
			if len(h1) == 0 || len(h2) == 0 {
				continue
			}
			s := compareHashes(h1, h2)
			if s > best {
				best = s
			}
		}
	}

	return best
}

// parseCTPH splits a CTPH string into its block size and hash components.
// Supports both legacy 2-hash format "N:hex1:hex2" and the new 3-hash
// format "N:hex1:hex2:hex3". Returns the block size and a slice of decoded
// hash byte slices (length 2 or 3).
func parseCTPH(fp string) (blockSize int, hashes [][]byte) {
	// Parse leading "N:"
	var bs int
	n, _ := fmt.Sscanf(fp, "%d:", &bs)
	if n != 1 || bs == 0 {
		return 0, nil
	}

	// Split on colons to extract hex segments after the block size.
	parts := strings.SplitN(fp, ":", 5) // at most "bs:h1:h2:h3" = 4 parts
	if len(parts) < 3 {
		return 0, nil
	}

	hexParts := parts[1:] // skip the block-size part
	hashes = make([][]byte, 0, len(hexParts))
	for _, h := range hexParts {
		if h == "" {
			return 0, nil
		}
		decoded, err := hex.DecodeString(h)
		if err != nil {
			return 0, nil
		}
		hashes = append(hashes, decoded)
	}

	if len(hashes) < 2 {
		return 0, nil
	}

	return bs, hashes
}

// compareHashes computes a similarity score (0-100) between two hash byte slices
// using a weighted edit distance.
func compareHashes(a, b []byte) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	dist := editDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	if maxLen == 0 {
		return 100
	}

	score := 100 - (dist*100)/maxLen
	if score < 0 {
		score = 0
	}
	return score
}

// editDistance computes the Levenshtein distance between two byte slices.
func editDistance(a, b []byte) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows for space efficiency.
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := prev[j] + 1
			del := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(ins, del, sub)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
