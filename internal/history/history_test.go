// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestCollectMetricsAndSkipsBulkCommits(t *testing.T) {
	root := t.TempDir()
	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.July, 21, 0, 0, 0, 0, time.UTC)
	commitFiles(t, root, worktree, now.Add(-100*24*time.Hour), map[string]string{
		"a.go": "package demo\nfunc A() {}\n",
		"b.go": "package demo\nfunc B() {}\n",
	})
	commitFiles(t, root, worktree, now.Add(-10*24*time.Hour), map[string]string{
		"a.go": "package demo\nfunc A() { B() }\n",
		"b.go": "package demo\nfunc B() { println(\"b\") }\n",
	})
	commitFiles(t, root, worktree, now.Add(-time.Hour), map[string]string{
		"a.go": "package demo\nfunc A() { B(); B() }\n",
		"x.go": "package demo\n",
		"y.go": "package demo\n",
	})

	result, err := Collect(root, []string{
		filepath.Join(root, "a.go"),
		filepath.Join(root, "b.go"),
	}, Options{
		Since:             now.Add(-365 * 24 * time.Hour),
		Now:               now,
		MaxCommits:        100,
		MaxFilesPerCommit: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.CommitsVisited != 3 {
		t.Fatalf("CommitsVisited = %d, want 3", result.CommitsVisited)
	}
	if result.CommitsAnalyzed != 2 {
		t.Fatalf("CommitsAnalyzed = %d, want 2", result.CommitsAnalyzed)
	}
	if result.SkippedBulkCommits != 1 {
		t.Fatalf("SkippedBulkCommits = %d, want 1", result.SkippedBulkCommits)
	}
	if result.Files["a.go"].Commits != 2 || result.Files["b.go"].Commits != 2 {
		t.Fatalf("unexpected file metrics: %#v", result.Files)
	}
	if result.Files["a.go"].Churn() == 0 {
		t.Fatal("expected non-zero churn for a.go")
	}
	if got := result.CoChanges[NewPair("a.go", "b.go")]; got != 2 {
		t.Fatalf("a.go/b.go cochanges = %d, want 2", got)
	}
	if !result.Files["a.go"].LastChanged.Equal(now.Add(-10 * 24 * time.Hour)) {
		t.Fatalf("last changed = %s", result.Files["a.go"].LastChanged)
	}
}

func TestRelativePathRejectsOutsideRepository(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.go")
	if _, err := RelativePath(root, outside); err == nil {
		t.Fatal("expected outside path to be rejected")
	}
}

func commitFiles(
	t *testing.T,
	root string,
	worktree *git.Worktree,
	when time.Time,
	files map[string]string,
) {
	t.Helper()
	for path, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := worktree.Add(filepath.FromSlash(path)); err != nil {
			t.Fatal(err)
		}
	}
	signature := &object.Signature{
		Name: "Test", Email: "test@example.com", When: when,
	}
	if _, err := worktree.Commit("change files", &git.CommitOptions{
		Author: signature, Committer: signature,
	}); err != nil {
		t.Fatal(err)
	}
}
