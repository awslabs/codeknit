//go:build !windows

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package outputfs

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func lockDirectory(path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // sibling lock path derived from the output directory
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil { //nolint:gosec // Unix file descriptors fit in int
		_ = file.Close()
		return nil, fmt.Errorf("another process is publishing this output: %w", err)
	}
	// The persistent lock file must not be unlinked: waiters must share its inode.
	return func() { _ = file.Close() }, nil
}
