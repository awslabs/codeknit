// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// ancestorIgnorePatterns uses one coordinate system for patterns and paths.
// Stop at the nearest repository boundary; outside a repository, honor
// ancestor ignore files without depending on the process working directory.
func ancestorIgnorePatterns(input string) (string, []gitignore.Pattern, []ScanError) {
	var ancestors []string
	root := input
	for dir := input; ; dir = filepath.Dir(dir) {
		ancestors = append(ancestors, dir)
		if _, err := os.Lstat(filepath.Join(dir, ".gitignore")); err == nil {
			root = dir
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			root = dir
			break
		}
		if parent := filepath.Dir(dir); parent == dir {
			break
		}
	}
	slices.Reverse(ancestors)
	var patterns []gitignore.Pattern
	var diagnostics []ScanError
	for _, dir := range ancestors {
		rel, err := filepath.Rel(root, dir)
		if err != nil || rel != "." && !filepath.IsLocal(rel) {
			continue
		}
		var domain []string
		if rel != "." {
			domain = strings.Split(filepath.ToSlash(rel), "/")
		}
		sub, issue := readIgnoreFile(filepath.Join(dir, ".gitignore"), domain, nil)
		if issue != nil {
			diagnostics = append(diagnostics, *issue)
		}
		patterns = append(patterns, sub...)
	}
	return root, patterns, diagnostics
}

func readIgnoreFile(path string, domain []string, root *os.Root) ([]gitignore.Pattern, *ScanError) {
	lstat, open := os.Lstat, os.Open
	local := filepath.Clean(path)
	if root != nil {
		var err error
		local, err = filepath.Rel(root.Name(), path)
		if err != nil {
			return nil, &ScanError{path, "scan: " + err.Error()}
		}
		lstat, open = root.Lstat, root.Open
	}
	info, err := lstat(local)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, &ScanError{path, "scan: " + err.Error()}
	}
	if !info.Mode().IsRegular() {
		return nil, &ScanError{path, "scan: non-regular .gitignore skipped"}
	}
	f, err := open(local)
	if err != nil {
		return nil, &ScanError{path, "scan: " + err.Error()}
	}
	defer func() { _ = f.Close() }()
	var patterns []gitignore.Pattern
	reader := bufio.NewScanner(f)
	for reader.Scan() {
		line := reader.Text()
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			patterns = append(patterns, gitignore.ParsePattern(line, domain))
		}
	}
	if err := reader.Err(); err != nil {
		return patterns, &ScanError{path, "scan: " + err.Error()}
	}
	return patterns, nil
}
