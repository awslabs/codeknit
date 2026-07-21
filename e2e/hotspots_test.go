// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestE2E_HistoryHotspots(t *testing.T) {
	inputDir := writeFixture(t, map[string]string{
		"api.go": `package sample
func API() { Service() }
`,
		"service.go": `package sample
func Service() {}
`,
	})
	repo, err := git.PlainInit(inputDir, false)
	if err != nil {
		t.Fatal(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	commitAll(t, worktree, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))

	service := filepath.Join(inputDir, "service.go")
	if writeErr := os.WriteFile(service, []byte(`package sample
func Service() { println("changed") }
`), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	commitAll(t, worktree, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))

	output := filepath.Join(t.TempDir(), "hotspots.skt")
	//nolint:gosec // controlled test binary and arguments
	cmd := exec.Command(binPath, "graph", "hotspots", inputDir,
		"--since", "2y",
		"--min-cochanges", "1",
		"--output", output,
	)
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("hotspots failed: %v\nstderr: %s", runErr, stderr.String())
	}

	content, err := os.ReadFile(output) //nolint:gosec // controlled test output path
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{
		"[history_hotspots]",
		"[hotspots]",
		"service.go",
		"[temporal_coupling]",
		"api.go <-> service.go",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}

func commitAll(t *testing.T, worktree *git.Worktree, when time.Time) {
	t.Helper()
	if err := worktree.AddGlob("*.go"); err != nil {
		t.Fatal(err)
	}
	signature := &object.Signature{Name: "Test", Email: "test@example.com", When: when}
	if _, err := worktree.Commit("change", &git.CommitOptions{
		Author: signature, Committer: signature,
	}); err != nil {
		t.Fatal(err)
	}
}
