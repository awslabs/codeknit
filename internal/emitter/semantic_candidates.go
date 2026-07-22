// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"sort"

	"codeknit/internal/fingerprint"
	"codeknit/internal/ollama"
	"codeknit/internal/plugin"
)

const (
	semanticNeighborK            = 5
	semanticCandidateCosineFloor = 0.80
)

type candidatePair struct {
	i int
	j int
}

func newCandidatePair(i, j int) candidatePair {
	if i > j {
		i, j = j, i
	}
	return candidatePair{i: i, j: j}
}

type semanticCandidate struct {
	i          int
	j          int
	similarity float64
}

// findSemanticCandidates returns the union of each symbol's top-K embedding
// neighbors. Comparisons are category-aware, and dense vectors are projected
// into the existing fixed-size candidate index before exact cosine scoring.
func findSemanticCandidates(entries []fpEntry, vectors [][]float32, k int) []semanticCandidate {
	if len(entries) < 2 || len(entries) != len(vectors) || k <= 0 {
		return nil
	}

	groups := make(map[plugin.SymbolCategory][]int)
	for i := range entries {
		groups[entries[i].category] = append(groups[entries[i].category], i)
	}

	var candidates []semanticCandidate
	for _, indexes := range groups {
		if len(indexes) < 2 {
			continue
		}

		groupVectors := make([][]float32, len(indexes))
		for i, entryIndex := range indexes {
			groupVectors[i] = vectors[entryIndex]
		}

		index := fingerprint.BuildVectorANNIndex(groupVectors)
		for _, candidate := range index.FindCandidates(k) {
			i := indexes[candidate.I]
			j := indexes[candidate.J]
			candidates = append(candidates, semanticCandidate{
				i:          i,
				j:          j,
				similarity: ollama.CosineSimilarity(vectors[i], vectors[j]),
			})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].similarity != candidates[j].similarity {
			return candidates[i].similarity > candidates[j].similarity
		}
		if candidates[i].i != candidates[j].i {
			return candidates[i].i < candidates[j].i
		}
		return candidates[i].j < candidates[j].j
	})
	return candidates
}
