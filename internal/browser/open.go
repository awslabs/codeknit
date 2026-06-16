// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package browser provides utilities for opening files in the default browser.
package browser

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Open opens the given file path in the default browser.
// Works on macOS, Linux, and Windows.
func Open(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	url := "file://" + abs

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url) // #nosec G204 -- url is constructed from a local file path, not user input
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) // #nosec G204 -- url is constructed from a local file path, not user input
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url) // #nosec G204 -- url is constructed from a local file path, not user input
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
