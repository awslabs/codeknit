// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"math"
	"sort"
)

// annDims is the dimensionality of the lightweight feature vector used for
// ANN pre-filtering. Each dimension corresponds to a bucket in a token
// bigram hash. 128 dims keeps memory small (~512 bytes per symbol) while
// providing enough resolution to separate structurally different functions.
const annDims = 128

// annK is the default number of nearest neighbors to retrieve per symbol.
// Only these candidates proceed to the expensive CTPH + token-edit comparison.
const annK = 50

// DefaultANNK returns the default K value for ANN candidate retrieval.
func DefaultANNK() int { return annK }

// ANNIndex is a brute-force nearest-neighbor candidate index built from
// lightweight token-bigram feature vectors. Vector retrieval is O(N²), but it
// limits the more expensive CTPH and edit-distance scoring to O(N·K) pairs.
type ANNIndex struct {
	vecs []annVec
}

// annVec is a single entry in the index.
type annVec struct {
	idx int // index into the original entries slice
	vec [annDims]float32
}

// Candidate is a pair of entry indexes with their ANN cosine similarity.
type Candidate struct {
	I, J      int
	CosineSim float32
}

// BuildANNIndex constructs an index from raw body-token streams.
// Each token stream is converted to a fixed-size feature vector using
// token bigram frequency hashing (a lightweight bag-of-bigrams).
func BuildANNIndex(tokenStreams [][]byte) *ANNIndex {
	return buildANNIndex(tokenStreams, true)
}

// BuildRawANNIndex constructs an index from full token streams, retaining
// payload bytes. This is appropriate for type shapes and value initializers,
// where names and literal values are meaningful.
func BuildRawANNIndex(tokenStreams [][]byte) *ANNIndex {
	return buildANNIndex(tokenStreams, false)
}

// BuildVectorANNIndex constructs a candidate index from dense embedding
// vectors. Signed feature hashing projects arbitrary model dimensions into
// the index's fixed 128 dimensions while approximately preserving cosine
// similarity.
func BuildVectorANNIndex(vectors [][]float32) *ANNIndex {
	idx := &ANNIndex{
		vecs: make([]annVec, len(vectors)),
	}
	for i, vector := range vectors {
		idx.vecs[i] = annVec{
			idx: i,
			vec: projectVector(vector),
		}
	}
	return idx
}

func buildANNIndex(tokenStreams [][]byte, structural bool) *ANNIndex {
	idx := &ANNIndex{
		vecs: make([]annVec, len(tokenStreams)),
	}
	for i, tokens := range tokenStreams {
		if structural {
			tokens = structuralTokens(tokens)
		}
		idx.vecs[i] = annVec{
			idx: i,
			vec: tokenBigramVec(tokens),
		}
	}
	return idx
}

func projectVector(vector []float32) [annDims]float32 {
	var projected [annDims]float32
	for i, value := range vector {
		hash := uint32(i+1) * 2654435761
		bucket := hash % annDims
		if hash&0x80000000 != 0 {
			projected[bucket] -= value
		} else {
			projected[bucket] += value
		}
	}
	normalizeANNVec(&projected)
	return projected
}

// FindCandidates returns deduplicated candidate pairs where at least one
// symbol considers the other among its top-K nearest neighbors. The result
// is sorted by cosine similarity descending.
func (idx *ANNIndex) FindCandidates(k int) []Candidate {
	if k <= 0 {
		k = annK
	}
	n := len(idx.vecs)
	if n < 2 {
		return nil
	}

	// For each vector, find its top-K neighbors by cosine similarity.
	// Use a map to deduplicate (i,j) pairs where i < j.
	type pairKey struct{ i, j int }
	seen := make(map[pairKey]float32, n*k)

	for a := 0; a < n; a++ {
		neighbors := idx.topK(a, k)
		for _, nb := range neighbors {
			i, j := a, nb.idx
			if i > j {
				i, j = j, i
			}
			if existing, ok := seen[pairKey{i, j}]; !ok || nb.sim > existing {
				seen[pairKey{i, j}] = nb.sim
			}
		}
	}

	candidates := make([]Candidate, 0, len(seen))
	for pk, sim := range seen {
		candidates = append(candidates, Candidate{I: pk.i, J: pk.j, CosineSim: sim})
	}
	sort.Slice(candidates, func(a, b int) bool {
		return candidates[a].CosineSim > candidates[b].CosineSim
	})
	return candidates
}

// neighbor is a (index, similarity) pair for top-K selection.
type neighbor struct {
	idx int
	sim float32
}

// topK finds the K nearest neighbors of vecs[a] by cosine similarity,
// excluding self. Uses a simple linear scan — fast enough for the vector
// sizes we deal with (128 dims, <100K entries).
func (idx *ANNIndex) topK(a, k int) []neighbor {
	va := &idx.vecs[a].vec
	heap := make([]neighbor, 0, k+1)

	for b := 0; b < len(idx.vecs); b++ {
		if b == a {
			continue
		}
		sim := cosSim32(va, &idx.vecs[b].vec)
		if len(heap) < k {
			heap = append(heap, neighbor{idx: b, sim: sim})
			if len(heap) == k {
				sort.Slice(heap, func(i, j int) bool { return heap[i].sim < heap[j].sim })
			}
		} else if sim > heap[0].sim {
			heap[0] = neighbor{idx: b, sim: sim}
			sort.Slice(heap, func(i, j int) bool { return heap[i].sim < heap[j].sim })
		}
	}
	return heap
}

// tokenBigramVec builds a fixed-size feature vector from a token stream
// by hashing consecutive token bigrams into buckets and L2-normalizing.
// This captures local structural patterns (e.g., "if → return", "for → call")
// in a compact, comparison-friendly form.
func tokenBigramVec(tokens []byte) [annDims]float32 {
	var vec [annDims]float32

	if len(tokens) < 2 {
		if len(tokens) == 1 {
			vec[int(tokens[0])%annDims]++
		}
		return vec
	}

	// Hash each bigram into a bucket.
	for i := 0; i < len(tokens)-1; i++ {
		// Simple hash: combine two consecutive tokens.
		h := uint(tokens[i])*31 + uint(tokens[i+1])
		vec[h%annDims]++
	}

	// Also add unigrams at half weight for short-function robustness.
	for _, tok := range tokens {
		vec[int(tok)%annDims] += 0.5
	}

	normalizeANNVec(&vec)
	return vec
}

func normalizeANNVec(vec *[annDims]float32) {
	var mag float64
	for i := range vec {
		mag += float64(vec[i]) * float64(vec[i])
	}
	if mag > 0 {
		inv := float32(1.0 / math.Sqrt(mag))
		for i := range vec {
			vec[i] *= inv
		}
	}
}

// cosSim32 computes cosine similarity between two fixed-size float32 vectors.
func cosSim32(a, b *[annDims]float32) float32 {
	var dot, magA, magB float64
	for i := 0; i < annDims; i++ {
		dot += float64(a[i]) * float64(b[i])
		magA += float64(a[i]) * float64(a[i])
		magB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(magA) * math.Sqrt(magB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}
