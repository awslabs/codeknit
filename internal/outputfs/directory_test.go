// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package outputfs

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, path, content string) {
	t.Helper()
	if err := WriteFile(path, []byte(content)); err != nil {
		t.Fatal(err)
	}
}

func wantContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path) //nolint:gosec // test-created fixture path
	if err != nil || string(got) != want {
		t.Fatalf("%s: got %q (%v), want %q", path, got, err, want)
	}
}

func managedSKT(path string) bool { return strings.HasSuffix(path, ".skt") }

func generateOne(stage string) ([]string, error) {
	file := filepath.Join(stage, "new.skt")
	return []string{file}, WriteFile(file, []byte("new"))
}

func TestReplaceDirectoryPreservesUnmanagedEntries(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	put(t, filepath.Join(out, "nested", "old.skt"), "old")
	put(t, filepath.Join(out, "nested", "notes.txt"), "keep me")
	written, err := ReplaceDirectory(out, true, managedSKT, generateOne)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 || written[0] != filepath.Join(out, "new.skt") {
		t.Fatalf("wrong published paths: %v", written)
	}
	wantContent(t, written[0], "new")
	wantContent(t, filepath.Join(out, "nested", "notes.txt"), "keep me")
	if _, err := os.Stat(filepath.Join(out, "nested", "old.skt")); !os.IsNotExist(err) {
		t.Fatalf("obsolete generated file remains: %v", err)
	}
}

func TestReplaceDirectoryFailurePreservesPreviousOutput(t *testing.T) {
	for _, failure := range []string{"generation", "collision", "without-clean"} {
		t.Run(failure, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "out")
			put(t, filepath.Join(out, "old.skt"), "old")
			generate := generateOne
			clean := failure != "without-clean"
			if failure == "generation" {
				generate = func(stage string) ([]string, error) {
					_, _ = generateOne(stage)
					return nil, errors.New("injected write failure")
				}
			}
			if failure == "collision" {
				if err := os.Mkdir(filepath.Join(out, "new.skt"), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := ReplaceDirectory(out, clean, managedSKT, generate); err == nil {
				t.Fatal("expected generation to fail")
			}
			wantContent(t, filepath.Join(out, "old.skt"), "old")
		})
	}
}

func TestPublicationRollback(t *testing.T) {
	parent := t.TempDir()
	out, stage, backup := filepath.Join(parent, "out"), filepath.Join(parent, "stage"), filepath.Join(parent, "backup")
	put(t, filepath.Join(out, "old.skt"), "old")
	put(t, filepath.Join(stage, "new.skt"), "new")
	injected := errors.New("injected rename failure")
	err := publishDirectory(out, stage, backup, true, func(from, to string) error {
		if from == stage {
			return injected
		}
		return os.Rename(from, to) //nolint:gosec // controlled test publication paths
	})
	if !errors.Is(err, injected) {
		t.Fatalf("expected publication failure: %v", err)
	}
	wantContent(t, filepath.Join(out, "old.skt"), "old")
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatalf("backup not restored: %v", err)
	}
}

// Terminate a real publisher between renames, then verify recovery on the next
// invocation. Exit bypasses deferred cleanup, as a killed process would.
func TestPublicationInterruption(t *testing.T) {
	if parent := os.Getenv("CODEKNIT_PUBLICATION_TEST"); parent != "" {
		out := filepath.Join(parent, "out")
		stage := filepath.Join(parent, "stage")
		backup := filepath.Join(parent, ".out.codeknit-backup")
		_ = publishDirectory(out, stage, backup, true, func(from, to string) error {
			if err := os.Rename(from, to); err != nil { //nolint:gosec // controlled test publication paths
				return err
			}
			if from == out {
				os.Exit(23)
			}
			return nil
		})
		os.Exit(24)
	}
	parent := t.TempDir()
	out := filepath.Join(parent, "out")
	put(t, filepath.Join(out, "old.skt"), "old")
	put(t, filepath.Join(parent, "stage", "new.skt"), "new")
	cmd := exec.Command(os.Args[0], "-test.run=^TestPublicationInterruption$") //nolint:gosec // re-execute this test binary
	cmd.Env = append(os.Environ(), "CODEKNIT_PUBLICATION_TEST="+parent)
	if err := cmd.Run(); err == nil {
		t.Fatal("helper did not terminate mid-publication")
	}
	_, err := ReplaceDirectory(out, true, managedSKT, func(string) ([]string, error) {
		return nil, errors.New("new generation failed after recovery")
	})
	if err == nil {
		t.Fatal("expected generation failure")
	}
	wantContent(t, filepath.Join(out, "old.skt"), "old")
}

func TestPublicationLock(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	lock := filepath.Join(filepath.Dir(out), ".out.codeknit.lock")
	unlock, err := lockDirectory(lock)
	if err != nil {
		t.Fatal(err)
	}
	_, blocked := ReplaceDirectory(out, true, managedSKT, generateOne)
	unlock()
	if blocked == nil {
		t.Fatal("concurrent publisher acquired an already-held lock")
	}
	if _, err := ReplaceDirectory(out, true, managedSKT, generateOne); err != nil {
		t.Fatalf("publisher could not acquire released lock: %v", err)
	}
}

func TestRecoveryKeepsBackupWhenDestinationIsUnsafe(t *testing.T) {
	parent := t.TempDir()
	out, backup := filepath.Join(parent, "out"), filepath.Join(parent, "backup")
	put(t, filepath.Join(backup, "old.skt"), "old")
	put(t, out, "unrelated")
	if err := recoverDirectory(out, backup); err == nil {
		t.Fatal("expected unsafe recovery to fail")
	}
	wantContent(t, filepath.Join(backup, "old.skt"), "old")
	wantContent(t, out, "unrelated")
}

func TestWriteFileDoesNotFollowSymlinks(t *testing.T) {
	parent := t.TempDir()
	target, link := filepath.Join(parent, "target"), filepath.Join(parent, "link")
	put(t, target, "original")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if err := WriteFile(link, []byte("replacement")); err == nil {
		t.Fatal("expected symlink output rejection")
	}
	wantContent(t, target, "original")
}
