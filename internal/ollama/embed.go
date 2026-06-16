// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package ollama provides a minimal client for Ollama's embedding API.
package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

const (
	// DefaultModel is the recommended fast local embedding model as of 2026.
	// qwen3-embedding:0.6b offers 32k context, 1024 dims, and explicit code
	// retrieval training — significantly better than nomic-embed-text for
	// distinguishing semantically different functions with similar structure.
	DefaultModel = "qwen3-embedding:0.6b"

	defaultBaseURL = "http://localhost:11434"
	embedPath      = "/api/embed"
	httpTimeout    = 120 * time.Second
)

// Client is a minimal Ollama HTTP client for embedding requests.
type Client struct {
	http    *http.Client
	baseURL string
	model   string
}

// NewClient returns a Client targeting the given Ollama base URL and model.
// Pass "" for baseURL to use the default (http://localhost:11434).
func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{Timeout: httpTimeout},
	}
}

// embedRequest is the JSON body sent to /api/embed.
type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embedResponse is the JSON body returned by /api/embed.
type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed sends a batch of texts to Ollama and returns one embedding vector
// per input. The returned slice has the same length as texts.
//
// Returns a descriptive error if Ollama is not reachable, suggesting the
// user install and start it.
func (c *Client) Embed(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embedRequest{Model: c.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.http.Post(c.baseURL+embedPath, "application/json", bytes.NewReader(body)) //nolint:noctx // no context needed for a simple CLI embedding call
	if err != nil {
		return nil, fmt.Errorf(
			"ollama not reachable at %s: %w\n\nTo enable semantic reranking:\n  1. Install Ollama: https://ollama.com\n  2. Run: ollama serve\n  3. Pull the model: ollama pull %s",
			c.baseURL, err, c.model,
		)
	}
	defer resp.Body.Close() //nolint:errcheck // response body close error is not actionable in a CLI tool

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ollama returned HTTP %d — is the model pulled?\n  Run: ollama pull %s",
			resp.StatusCode, c.model,
		)
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(result.Embeddings), len(texts))
	}
	return result.Embeddings, nil
}

// CosineSimilarity returns the cosine similarity between two vectors,
// in the range [-1, 1]. Returns 0 if either vector has zero magnitude.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		magA += float64(a[i]) * float64(a[i])
		magB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(magA) * math.Sqrt(magB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
