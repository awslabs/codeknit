// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package scanpolicy defines source selection and incomplete-scan behavior.
package scanpolicy

import (
	"fmt"
	"path"
	"strings"
)

// DefaultMaxFileSize bounds source reads made by CLI commands, in bytes.
const DefaultMaxFileSize int64 = 10 * 1024 * 1024

// Options applies to every source-processing command. Patterns use
// slash-separated, gitignore-style globs relative to the selected input root.
// Symlinks and non-regular files are always skipped and reported.
type Options struct {
	Include     []string
	Exclude     []string
	MaxFileSize int64 // zero explicitly disables the size limit
	Strict      bool  // reject incomplete results before writing output
}

// Validate rejects malformed or contradictory selection settings.
func (o *Options) Validate() error {
	if o.MaxFileSize < 0 {
		return fmt.Errorf("max-file-size must be nonnegative (0 disables the limit)")
	}
	for _, patterns := range [][]string{o.Include, o.Exclude} {
		for _, pattern := range patterns {
			if strings.TrimSpace(pattern) == "" || strings.ContainsAny(pattern, "\r\n") ||
				strings.HasPrefix(pattern, "!") {
				return fmt.Errorf("invalid selection pattern %q: use a nonempty glob without negation", pattern)
			}
			if _, err := path.Match(pattern, ""); err != nil {
				return fmt.Errorf("invalid selection pattern %q: %w", pattern, err)
			}
		}
	}
	return nil
}
