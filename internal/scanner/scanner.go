// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package scanner discovers source code files in a directory tree,
// applying .gitignore rules, extension filtering, and test-file exclusion.
package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"codeknit/internal/plugin"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Scanner discovers source code files in a directory tree.
type Scanner struct {
	OnProgress   func(visited, matched int)   // called periodically during scan, if non-nil
	TestPatterns map[string]plugin.TestConfig // extension → test file patterns
	Extensions   []string                     // registered file extensions from plugins
	CollectTest  bool
}

// Scan walks the input path and returns relative file paths that match
// the registered extensions, honoring .gitignore rules and test-file filtering.
// If inputPath is a regular file with a supported extension, it returns a
// single-element list containing the file's basename. If the extension is
// unsupported, it returns an empty list. If inputPath is a directory, it
// walks recursively using parallel directory reads.
func (s *Scanner) Scan(inputPath string) ([]string, error) {
	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, err
	}

	extSet := make(map[string]bool, len(s.Extensions))
	for _, ext := range s.Extensions {
		extSet[ext] = true
	}

	// If inputPath is a regular file, handle it directly.
	info, err := os.Stat(absInput)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		ext := filepath.Ext(info.Name())
		if extSet[ext] {
			return []string{filepath.Base(absInput)}, nil
		}
		return nil, nil
	}

	relPrefix := absInput + string(filepath.Separator)

	// Parallel directory walker.
	// Each directory is processed as a work item: read its entries, filter,
	// collect matched files, and enqueue subdirectories.
	type dirJob struct {
		absPath  string
		patterns []gitignore.Pattern // inherited gitignore patterns for this dir
	}

	// Pre-load .gitignore patterns from the CWD down to absInput.
	// This ensures that when the user scans a subdirectory (e.g.
	// "codeknit parse src/"), ancestor .gitignore files (like the one
	// in the project root) are still honored.
	rootPatterns := collectAncestorGitignorePatterns(absInput)

	var (
		mu      sync.Mutex
		results []string
		visited int64
	)

	// Use a WaitGroup to track in-flight directory jobs.
	var wg sync.WaitGroup
	jobs := make(chan dirJob, 256)

	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	}

	// processDir handles a single directory job: reads entries, filters,
	// collects files, and enqueues or recursively processes subdirectories.
	var processDir func(job dirJob)
	processDir = func(job dirJob) {
		defer wg.Done()

		entries, readErr := os.ReadDir(job.absPath)
		if readErr != nil {
			return
		}

		var localFiles []string
		var subDirs []dirJob

		// Build the matcher once for all entries in this directory.
		matcher := gitignore.NewMatcher(job.patterns)

		for _, entry := range entries {
			name := entry.Name()
			entryPath := filepath.Join(job.absPath, name)

			// Skip well-known non-source directories.
			if entry.IsDir() {
				switch name {
				case ".git", ".hg", ".svn", "node_modules", "__pycache__",
					".tox", ".mypy_cache", ".pytest_cache", ".venv", "venv":
					continue
				}
			}

			atomic.AddInt64(&visited, 1)

			// For files, check extension early.
			if !entry.IsDir() {
				ext := filepath.Ext(name)
				if !extSet[ext] {
					continue
				}
			}

			// Check gitignore using go-git's spec-compliant matcher.
			rel := entryPath[len(relPrefix):]
			pathSegments := strings.Split(filepath.ToSlash(rel), "/")
			if matcher.Match(pathSegments, entry.IsDir()) {
				continue
			}

			if entry.IsDir() {
				// Build patterns for this subdirectory by appending any
				// .gitignore found in the child directory.
				childPatterns := job.patterns
				gi := filepath.Join(entryPath, ".gitignore")
				if sub := parseGitignoreFile(gi, pathSegments); len(sub) > 0 {
					merged := make([]gitignore.Pattern, len(job.patterns)+len(sub))
					copy(merged, job.patterns)
					copy(merged[len(job.patterns):], sub)
					childPatterns = merged
				}
				subDirs = append(subDirs, dirJob{absPath: entryPath, patterns: childPatterns})
				continue
			}

			// File passed all filters — compute relative path.
			slashRel := filepath.ToSlash(rel)

			if !s.CollectTest && s.isTestFile(slashRel) {
				continue
			}

			localFiles = append(localFiles, slashRel)
		}

		// Batch-append results under lock.
		if len(localFiles) > 0 {
			mu.Lock()
			results = append(results, localFiles...)
			mu.Unlock()
		}

		// Report progress.
		if s.OnProgress != nil {
			mu.Lock()
			matched := len(results)
			mu.Unlock()
			s.OnProgress(int(atomic.LoadInt64(&visited)), matched)
		}

		// Enqueue subdirectories. Use non-blocking send to avoid
		// deadlock when the channel is full — process overflow inline.
		wg.Add(len(subDirs))
		for _, sub := range subDirs {
			select {
			case jobs <- sub:
			default:
				processDir(sub)
			}
		}
	}

	// Worker: process directories from the jobs channel.
	for range numWorkers {
		go func() {
			for job := range jobs {
				processDir(job)
			}
		}()
	}

	// Seed with root directory.
	wg.Add(1)
	jobs <- dirJob{absPath: absInput, patterns: rootPatterns}

	// Wait for all work to complete, then close the channel.
	wg.Wait()
	close(jobs)

	// Sort for deterministic output.
	sort.Strings(results)
	return results, nil
}

// isTestFile returns true if the relative path matches test file patterns.
// It checks common test directory names and universal filename patterns first,
// then uses language-specific patterns from the plugin's TestPatterns configuration.
func (s *Scanner) isTestFile(rel string) bool {
	parts := strings.Split(rel, "/")

	// Check for common test directory components (shared across all languages).
	for _, p := range parts {
		switch p {
		case "__tests__", "test", "tests", "spec", "specs":
			return true
		}
	}

	base := filepath.Base(rel)
	ext := filepath.Ext(base)
	nameNoExt := strings.TrimSuffix(base, ext)

	// Universal filename patterns: .test. and .spec. in the basename.
	if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
		return true
	}

	// Look up language-specific patterns by extension.
	tc, ok := s.TestPatterns[ext]
	if !ok {
		return false
	}

	for _, dot := range tc.ContainsDot {
		if strings.Contains(base, dot) {
			return true
		}
	}
	for _, suffix := range tc.NameSuffixes {
		if strings.HasSuffix(nameNoExt, suffix) {
			return true
		}
	}
	for _, prefix := range tc.NamePrefixes {
		if strings.HasPrefix(nameNoExt, prefix) {
			return true
		}
	}

	return false
}

// parseGitignoreFile reads a .gitignore file and returns parsed patterns using
// go-git's spec-compliant gitignore implementation. The domain represents the
// path segments of the directory containing the .gitignore relative to the
// repository root (nil for the root .gitignore).
func parseGitignoreFile(path string, domain []string) []gitignore.Pattern {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var patterns []gitignore.Pattern
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			patterns = append(patterns, gitignore.ParsePattern(line, domain))
		}
	}
	return patterns
}

// collectAncestorGitignorePatterns walks from the CWD down to absInput,
// collecting .gitignore patterns from each directory along the straight-line
// path (inclusive on both ends). Ancestor patterns use a nil domain so they
// apply globally within the scanned tree.
func collectAncestorGitignorePatterns(absInput string) []gitignore.Pattern {
	cwd, err := os.Getwd()
	if err != nil {
		return parseGitignoreFile(filepath.Join(absInput, ".gitignore"), nil)
	}
	absCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return parseGitignoreFile(filepath.Join(absInput, ".gitignore"), nil)
	}

	// Resolve symlinks on absInput too so both paths are comparable
	// (e.g. macOS /var → /private/var).
	resolvedInput, err := filepath.EvalSymlinks(absInput)
	if err != nil {
		resolvedInput = absInput
	}

	rel, err := filepath.Rel(absCwd, resolvedInput)
	if err != nil || strings.HasPrefix(rel, "..") {
		// absInput is not under CWD — fall back to input dir only.
		return parseGitignoreFile(filepath.Join(absInput, ".gitignore"), nil)
	}

	// Build directory list from CWD down to absInput (inclusive).
	dirs := []string{absCwd}
	if rel != "." {
		segments := strings.Split(filepath.ToSlash(rel), "/")
		cur := absCwd
		for _, seg := range segments {
			cur = filepath.Join(cur, seg)
			dirs = append(dirs, cur)
		}
	}

	var patterns []gitignore.Pattern
	for _, dir := range dirs {
		gi := filepath.Join(dir, ".gitignore")
		// Ancestors (above absInput) get nil domain — their patterns
		// apply to all paths in the scan. The absInput dir itself also
		// gets nil domain, matching the original behavior.
		var domain []string
		if dirRel, err := filepath.Rel(absInput, dir); err == nil && dirRel != "." && !strings.HasPrefix(dirRel, "..") {
			domain = strings.Split(filepath.ToSlash(dirRel), "/")
		}
		if sub := parseGitignoreFile(gi, domain); len(sub) > 0 {
			patterns = append(patterns, sub...)
		}
	}
	return patterns
}
