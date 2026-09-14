//go:build !windows

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package outputfs

import (
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestNonRegularOutputEntryFailsWithoutBlocking(t *testing.T) {
	out := t.TempDir()
	put(t, filepath.Join(out, "old.skt"), "old")
	if err := unix.Mkfifo(filepath.Join(out, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := ReplaceDirectory(out, true, managedSKT, generateOne)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("non-regular output entry should reject publication")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("publication blocked reading an unrelated FIFO")
	}
	wantContent(t, filepath.Join(out, "old.skt"), "old")
}
