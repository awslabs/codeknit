// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package scanner discovers source files within an explicit input boundary.
package scanner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"codeknit/internal/plugin"
	"codeknit/internal/scanpolicy"
	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Scanner discovers regular source files; symlinks are never followed.
type Scanner struct {
	OnProgress   func(visited, matched int)
	TestPatterns map[string]plugin.TestConfig
	Extensions   []string
	Policy       scanpolicy.Options
	CollectTest  bool
}

// ScanError describes a source or directory omitted from the scan.
type ScanError struct{ FilePath, Reason string }

func (d ScanError) Error() string { return d.FilePath + ": " + d.Reason }

// Result retains the input base and diagnostics alongside relative file paths.
type Result struct {
	Root        string
	Files       []string
	Diagnostics []ScanError
}

// Scan returns discovered files and reports any incomplete traversal as an error.
// Call ScanDetailed to explicitly handle partial results and their diagnostics.
func (s *Scanner) Scan(inputPath string) ([]string, error) {
	result, err := s.ScanDetailed(inputPath)
	if err != nil {
		return nil, err
	}
	var issues []error
	for _, d := range result.Diagnostics {
		issues = append(issues, d)
	}
	return result.Files, errors.Join(issues...)
}

// ScanDetailed uses project-relative ignore matching independently of the CWD.
// Explicit file inputs retain the same size and filesystem-boundary policy.
func (s *Scanner) ScanDetailed(inputPath string) (*Result, error) {
	if err := s.Policy.Validate(); err != nil {
		return nil, err
	}
	input, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(input)
	if err != nil {
		return nil, err
	}
	result := &Result{Root: input}
	single := !info.IsDir()
	if single {
		result.Root = filepath.Dir(input)
	}

	extensions := make(map[string]bool, len(s.Extensions))
	for _, ext := range s.Extensions {
		extensions[ext] = true
	}
	include, exclude := selectionMatcher(s.Policy.Include), selectionMatcher(s.Policy.Exclude)
	selected := func(rel string, directory bool) bool {
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if exclude.Match(parts, directory) {
			return false
		}
		return directory || len(s.Policy.Include) == 0 || include.Match(parts, false)
	}
	checkFile := func(path string, info os.FileInfo) *ScanError {
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			return &ScanError{path, "scan: symlink skipped; symlinks are not followed"}
		case !info.Mode().IsRegular():
			return &ScanError{path, "scan: non-regular file skipped"}
		case s.Policy.MaxFileSize > 0 && info.Size() > s.Policy.MaxFileSize:
			return &ScanError{path, fmt.Sprintf("scan: file size %d exceeds max-file-size %d", info.Size(), s.Policy.MaxFileSize)}
		}
		return nil
	}
	if single {
		if issue := checkFile(input, info); issue != nil {
			result.Diagnostics = append(result.Diagnostics, *issue)
		} else if extensions[filepath.Ext(input)] && selected(filepath.Base(input), false) {
			result.Files = []string{filepath.Base(input)}
		}
		return result, nil
	}
	root, err := os.OpenRoot(input)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	ignoreRoot, patterns, issues := ancestorIgnorePatterns(input)
	result.Diagnostics = append(result.Diagnostics, issues...)
	// Git does not descend into excluded ancestors, even when the requested
	// input starts beneath one or a nested ignore file tries to re-include it.
	ancestorMatcher := gitignore.NewMatcher(patterns)
	inputRel, _ := filepath.Rel(ignoreRoot, input)
	if inputRel != "." {
		parts := strings.Split(filepath.ToSlash(inputRel), "/")
		for i := range parts {
			if ancestorMatcher.Match(parts[:i+1], true) {
				return result, nil
			}
		}
	}
	type dirJob struct {
		path     string
		patterns []gitignore.Pattern
	}
	var mu sync.Mutex
	var visited atomic.Int64
	report := func(d ScanError) { mu.Lock(); result.Diagnostics = append(result.Diagnostics, d); mu.Unlock() }
	jobs := make(chan dirJob, 256)
	var pending sync.WaitGroup
	var process func(dirJob)
	process = func(job dirJob) {
		defer pending.Done()
		dirRel, _ := filepath.Rel(input, job.path)
		directory, readErr := root.Open(dirRel)
		if readErr != nil {
			report(ScanError{job.path, "scan: " + readErr.Error()})
			return
		}
		entries, readErr := directory.ReadDir(-1)
		_ = directory.Close()
		if readErr != nil {
			report(ScanError{job.path, "scan: " + readErr.Error()})
			return
		}
		matcher := gitignore.NewMatcher(job.patterns)
		var children []dirJob
		var files []string
		for _, entry := range entries {
			name := entry.Name()
			file := filepath.Join(job.path, name)
			if entry.IsDir() && excludedDirectory(name) {
				continue
			}
			visited.Add(1)
			rel, _ := filepath.Rel(input, file)
			ignoredRel, _ := filepath.Rel(ignoreRoot, file)
			if matcher.Match(strings.Split(filepath.ToSlash(ignoredRel), "/"), entry.IsDir()) ||
				!selected(rel, entry.IsDir() || entry.Type()&os.ModeSymlink != 0) {
				continue
			}
			if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 && !extensions[filepath.Ext(name)] {
				continue
			}
			fi, statErr := root.Lstat(rel)
			if statErr != nil {
				report(ScanError{file, "scan: " + statErr.Error()})
				continue
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				report(ScanError{file, "scan: symlink skipped; symlinks are not followed"})
				continue
			}
			if entry.IsDir() {
				sub, issue := readIgnoreFile(filepath.Join(file, ".gitignore"), strings.Split(filepath.ToSlash(ignoredRel), "/"), root)
				if issue != nil {
					report(*issue)
				}
				inherited := job.patterns
				if len(sub) > 0 {
					inherited = append(append([]gitignore.Pattern(nil), job.patterns...), sub...)
				}
				children = append(children, dirJob{file, inherited})
				continue
			}
			if !s.CollectTest && s.isTestFile(filepath.ToSlash(rel)) {
				continue
			}
			if issue := checkFile(file, fi); issue != nil {
				report(*issue)
				continue
			}
			files = append(files, filepath.ToSlash(rel))
		}
		mu.Lock()
		result.Files = append(result.Files, files...)
		matched := len(result.Files)
		mu.Unlock()
		if s.OnProgress != nil {
			s.OnProgress(int(visited.Load()), matched)
		}
		pending.Add(len(children))
		for _, child := range children {
			select {
			case jobs <- child:
			default:
				process(child)
			}
		}
	}
	var workers sync.WaitGroup
	for range max(4, runtime.NumCPU()) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				process(job)
			}
		}()
	}
	pending.Add(1)
	jobs <- dirJob{input, patterns}
	pending.Wait()
	close(jobs)
	workers.Wait()
	sort.Strings(result.Files)
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		a, b := result.Diagnostics[i], result.Diagnostics[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		return a.Reason < b.Reason
	})
	return result, nil
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "__pycache__", ".tox", ".mypy_cache", ".pytest_cache", ".venv", "venv":
		return true
	}
	return false
}

func selectionMatcher(patterns []string) gitignore.Matcher {
	compiled := make([]gitignore.Pattern, 0, len(patterns))
	for _, pattern := range patterns {
		compiled = append(compiled, gitignore.ParsePattern(pattern, nil))
	}
	return gitignore.NewMatcher(compiled)
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
