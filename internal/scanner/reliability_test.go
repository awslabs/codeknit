// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeknit/internal/scanpolicy"
)

func TestIgnoreRulesIndependentOfWorkingDirectory(t *testing.T) {
	project := setupTree(t, map[string]string{
		".gitignore":  "ignored.go\n/src/generated/\n",
		"src/kept.go": "", "src/ignored.go": "", "src/generated/model.go": "",
		"src/lib/generated/keep.go": "",
	})
	sc := &Scanner{Extensions: []string{".go"}}
	input := filepath.Join(project, "src")
	for _, cwd := range []string{project, t.TempDir()} {
		t.Chdir(cwd)
		got, err := sc.Scan(input)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"kept.go", "lib/generated/keep.go"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("cwd=%s: got %v, want %v", cwd, got, want)
		}
	}
}

func TestScanSelectionAndSizeDiagnostics(t *testing.T) {
	project := setupTree(t, map[string]string{
		"src/app.go": "short", "src/large.go": strings.Repeat("x", 30),
		"src/generated/model.go": "", "other/lib.go": "", "src/app.ts": "",
	})
	sc := &Scanner{Extensions: []string{".go", ".ts"}, Policy: scanpolicy.Options{
		Include: []string{"src/**/*.go"}, Exclude: []string{"**/generated/"},
		MaxFileSize: 10,
	}}
	result, err := sc.ScanDetailed(project)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Files, []string{"src/app.go"}) {
		t.Fatalf("selection mismatch: %v", result.Files)
	}
	if len(result.Diagnostics) != 1 || !strings.Contains(result.Diagnostics[0].Reason, "max-file-size") {
		t.Fatalf("expected a structured size diagnostic: %+v", result.Diagnostics)
	}
	sc.Policy.MaxFileSize = 0
	files, err := sc.Scan(project)
	if err != nil || len(files) != 2 {
		t.Fatalf("unlimited size selection: %v, %v", files, err)
	}
}

func TestIgnoreRepositoryBoundaryAndNegations(t *testing.T) {
	project := setupTree(t, map[string]string{
		".gitignore":       "*.go\n",
		"repo/.git":        "gitdir: ../worktree\n",
		"repo/.gitignore":  "*.go\n!src/keep.go\n/blocked/\n",
		"repo/src/keep.go": "", "repo/src/drop.go": "",
		"repo/blocked/.gitignore": "!keep.go\n", "repo/blocked/keep.go": "",
	})
	sc := &Scanner{Extensions: []string{".go"}}
	files, err := sc.Scan(filepath.Join(project, "repo", "src"))
	if err != nil || !reflect.DeepEqual(files, []string{"keep.go"}) {
		t.Fatalf("nested repo and negation: %v, %v", files, err)
	}
	files, err = sc.Scan(filepath.Join(project, "repo", "blocked"))
	if err != nil || len(files) != 0 {
		t.Fatalf("selected ignored directory must stay ignored: %v, %v", files, err)
	}
}

func TestScanReportsUnreadableDirectory(t *testing.T) {
	project := setupTree(t, map[string]string{"app.go": "", "locked/hidden.go": ""})
	locked := filepath.Join(project, "locked")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) }) //nolint:gosec // restore traversable test directory
	if _, err := os.ReadDir(locked); err == nil {
		t.Skip("filesystem/user does not enforce directory permissions")
	}
	sc := &Scanner{Extensions: []string{".go"}}
	result, err := sc.ScanDetailed(project)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Files, []string{"app.go"}) || len(result.Diagnostics) == 0 {
		t.Fatalf("lost partial scan or diagnostic: %+v", result)
	}
	if result.Diagnostics[0].FilePath != locked {
		t.Fatalf("missing failed-directory identity: %+v", result.Diagnostics)
	}
}

func TestScanReportsInvalidIgnoreFile(t *testing.T) {
	project := setupTree(t, map[string]string{"app.go": "", ".gitignore": strings.Repeat("x", 70_000)})
	sc := &Scanner{Extensions: []string{".go"}}
	result, err := sc.ScanDetailed(project)
	if err != nil || len(result.Files) != 1 || len(result.Diagnostics) != 1 {
		t.Fatalf("ignore read failure was not reported with partial files: %+v, %v", result, err)
	}
}

func TestScannerDoesNotFollowFileSymlinks(t *testing.T) {
	project := setupTree(t, map[string]string{"app.go": ""})
	outside := setupTree(t, map[string]string{"outside.go": ""})
	link := filepath.Join(project, "linked.go")
	if err := os.Symlink(filepath.Join(outside, "outside.go"), link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	sc := &Scanner{Extensions: []string{".go"}}
	files, _ := sc.Scan(project)
	if !reflect.DeepEqual(files, []string{"app.go"}) {
		t.Fatalf("scan included symlink target: %v", files)
	}
}

func TestIncludePatternsStillReportDirectorySymlinks(t *testing.T) {
	project := setupTree(t, map[string]string{"app.go": ""})
	outside := setupTree(t, map[string]string{"outside.go": ""})
	link := filepath.Join(project, "linked")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip(err)
	}
	sc := &Scanner{Extensions: []string{".go"}, Policy: scanpolicy.Options{Include: []string{"**/*.go"}}}
	result, err := sc.ScanDetailed(project)
	if err != nil || !reflect.DeepEqual(result.Files, []string{"app.go"}) {
		t.Fatalf("directory symlink escaped selection: %+v, %v", result, err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].FilePath != link {
		t.Fatalf("include filter concealed skipped directory symlink: %+v", result.Diagnostics)
	}
}
