// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"codeknit/internal/plugin"
)

const (
	edgeEndpointPattern = `(?:"(?:\\.|[^"\\])*"|[^\s,"\[\]]+)`
	edgeListPattern     = edgeEndpointPattern + `(?:\s*,\s*` + edgeEndpointPattern + `)*`
)

// Quoting protects delimiters in literal IDs. Known SIDs identify endpoints
// regardless of resolution; other unquoted uncertain targets are local names.
var (
	edgeLineRe = regexp.MustCompile(`^(` + edgeEndpointPattern + `)\s+--(\w+)(?:\[(unresolved|ambiguous)\])?-->\s+(` +
		edgeListPattern + `)(?:\s+\[candidates=(` + edgeListPattern + `)\])?$`)
	edgeEndpointRe = regexp.MustCompile(edgeEndpointPattern)
)

func decodeEdgeID(token string, shortToFull map[string]string) (string, error) {
	if strings.HasPrefix(token, `"`) {
		return strconv.Unquote(token)
	}
	if full, ok := shortToFull[token]; ok {
		return full, nil
	}
	return token, nil
}

func parseEdgeLine(line string, shortToFull, dict map[string]string, minified bool) ([]plugin.Edge, error) {
	m := edgeLineRe.FindStringSubmatch(line)
	if m == nil {
		return nil, fmt.Errorf("line does not match edge format")
	}
	kind, resolution := m[2], m[3]
	if minified {
		if decoded, ok := dict[kind]; ok {
			kind = decoded
		}
	}
	from, err := decodeEdgeID(m[1], shortToFull)
	if err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}
	sourceFile, _, _ := strings.Cut(from, "::")
	var candidates []string
	for _, token := range edgeEndpointRe.FindAllString(m[5], -1) {
		candidate, decodeErr := decodeEdgeID(token, shortToFull)
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid candidate: %w", decodeErr)
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) > 0 && resolution == "" {
		return nil, fmt.Errorf("candidate list requires an unresolved or ambiguous edge")
	}

	var edges []plugin.Edge
	for _, token := range edgeEndpointRe.FindAllString(m[4], -1) {
		var target string
		if resolution != "" && shortToFull[token] == "" && !strings.HasPrefix(token, `"`) && !strings.Contains(token, "::") {
			target = sourceFile + "::" + token
		} else {
			target, err = decodeEdgeID(token, shortToFull)
			if err != nil {
				return nil, fmt.Errorf("invalid target: %w", err)
			}
		}
		edges = append(edges, plugin.Edge{
			From: from, To: target, Kind: plugin.EdgeKind(kind),
			Resolution: resolution, Candidates: slices.Clone(candidates),
		})
	}
	return edges, nil
}
