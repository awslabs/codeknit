// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package history derives file change metrics and temporal coupling from Git.
package history

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

const recencyHalfLife = 90 * 24 * time.Hour

// Options controls Git history collection.
type Options struct {
	Since             time.Time
	Now               time.Time
	MaxCommits        int
	MaxFilesPerCommit int
	IncludeMerges     bool
}

// FileMetrics contains historical change metrics for one current source file.
type FileMetrics struct {
	LastChanged  time.Time `json:"last_changed"`
	Path         string    `json:"path"`
	RecencyScore float64   `json:"recency_score"`
	Commits      int       `json:"commits"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
}

// Churn returns total lines added and deleted.
func (m FileMetrics) Churn() int {
	return m.Additions + m.Deletions
}

// Pair identifies two files that changed together.
type Pair struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

// Result is the historical data collected from a repository.
type Result struct {
	Files              map[string]*FileMetrics `json:"files"`
	CoChanges          map[Pair]int            `json:"co_changes"`
	RepositoryRoot     string                  `json:"-"`
	CommitsVisited     int                     `json:"commits_visited"`
	CommitsAnalyzed    int                     `json:"commits_analyzed"`
	SkippedMerges      int                     `json:"skipped_merges"`
	SkippedBulkCommits int                     `json:"skipped_bulk_commits"`
}

// Collect traverses repository history and computes metrics for currentFiles.
// Current file paths may be absolute or relative to inputPath.
func Collect(inputPath string, currentFiles []string, opts Options) (*Result, error) {
	root, err := findRepositoryRoot(inputPath)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(root)
	if err != nil {
		return nil, fmt.Errorf("open Git repository at %s: %w", root, err)
	}

	current := make(map[string]bool, len(currentFiles))
	for _, path := range currentFiles {
		rel, relErr := RelativePath(root, path)
		if relErr == nil {
			current[rel] = true
		}
	}

	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}

	iter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		Since: nonZeroTime(opts.Since),
	})
	if err != nil {
		return nil, fmt.Errorf("read Git history: %w", err)
	}
	defer iter.Close()

	result := &Result{
		Files:          make(map[string]*FileMetrics, len(current)),
		CoChanges:      make(map[Pair]int),
		RepositoryRoot: root,
	}

	err = iter.ForEach(func(commit *object.Commit) error {
		if opts.MaxCommits > 0 && result.CommitsVisited >= opts.MaxCommits {
			return storer.ErrStop
		}
		result.CommitsVisited++

		if !opts.IncludeMerges && commit.NumParents() > 1 {
			result.SkippedMerges++
			return nil
		}

		stats, statsErr := commit.Stats()
		if statsErr != nil {
			return fmt.Errorf("read stats for commit %s: %w", commit.Hash.String(), statsErr)
		}
		if opts.MaxFilesPerCommit > 0 && len(stats) > opts.MaxFilesPerCommit {
			result.SkippedBulkCommits++
			return nil
		}

		touched := collectTouched(stats, current)
		if len(touched) == 0 {
			return nil
		}
		result.CommitsAnalyzed++

		when := commit.Committer.When
		if when.IsZero() {
			when = commit.Author.When
		}
		weight := recencyWeight(opts.Now, when)
		for _, stat := range touched {
			metrics := result.Files[stat.Name]
			if metrics == nil {
				metrics = &FileMetrics{Path: stat.Name}
				result.Files[stat.Name] = metrics
			}
			metrics.Commits++
			metrics.Additions += stat.Addition
			metrics.Deletions += stat.Deletion
			metrics.RecencyScore += weight
			if when.After(metrics.LastChanged) {
				metrics.LastChanged = when
			}
		}

		for i := 0; i < len(touched); i++ {
			for j := i + 1; j < len(touched); j++ {
				result.CoChanges[NewPair(touched[i].Name, touched[j].Name)]++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// NewPair returns a deterministic pair key.
func NewPair(a, b string) Pair {
	if a > b {
		a, b = b, a
	}
	return Pair{Left: a, Right: b}
}

// RelativePath normalizes path to a slash-separated path relative to root.
func RelativePath(root, path string) (string, error) {
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = abs
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	if rel == "." || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("path %s is outside repository %s", path, root)
	}
	return filepath.ToSlash(rel), nil
}

func findRepositoryRoot(inputPath string) (string, error) {
	abs, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("resolve input path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat input path: %w", err)
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	for dir := abs; ; dir = filepath.Dir(dir) {
		if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("no Git repository found from %s", inputPath)
}

func collectTouched(stats object.FileStats, current map[string]bool) object.FileStats {
	touched := make(object.FileStats, 0, len(stats))
	for _, stat := range stats {
		name := filepath.ToSlash(stat.Name)
		if current[name] {
			stat.Name = name
			touched = append(touched, stat)
		}
	}
	sort.Slice(touched, func(i, j int) bool {
		return touched[i].Name < touched[j].Name
	})
	return touched
}

func recencyWeight(now, changed time.Time) float64 {
	age := now.Sub(changed)
	if age < 0 {
		age = 0
	}
	return math.Exp(-math.Ln2 * float64(age) / float64(recencyHalfLife))
}

func nonZeroTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
