// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package console provides colored output methods for CLI feedback.
package console

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Console wraps fatih/color for consistent colored CLI output.
// Color is auto-disabled when stdout/stderr is not a terminal.
type Console struct {
	errRed       *color.Color
	errYell      *color.Color
	green        *color.Color
	cyan         *color.Color
	lastProgress time.Time
	lastMsg      string
	verbose      bool
}

// New returns a Console with pre-configured color attributes.
func New() *Console {
	errRed := color.New(color.FgRed)
	errRed.SetWriter(os.Stderr)

	errYell := color.New(color.FgYellow)
	errYell.SetWriter(os.Stderr)

	return &Console{
		errRed:  errRed,
		errYell: errYell,
		green:   color.New(color.FgGreen),
		cyan:    color.New(color.FgCyan),
	}
}

// logStderr prints a message to stderr using the given color.
func (c *Console) logStderr(col *color.Color, msg string) {
	_, _ = col.Fprintln(os.Stderr, msg)
}

// Error prints a red error message to stderr.
func (c *Console) Error(msg string) { c.logStderr(c.errRed, msg) }

// Warn prints a yellow warning message to stderr.
func (c *Console) Warn(msg string) { c.logStderr(c.errYell, msg) }

// Success prints a green success message to stdout.
func (c *Console) Success(msg string) {
	_, _ = c.green.Println(msg)
}

// Summary prints a formatted summary of the pipeline run.
func (c *Console) Summary(processed, skipped, parseErrors, written int) {
	_, _ = c.cyan.Println(fmt.Sprintf("Files processed: %d", processed))
	if skipped > 0 {
		_, _ = c.errYell.Fprintln(os.Stderr, fmt.Sprintf("Files skipped:   %d", skipped))
	}
	if parseErrors > 0 {
		_, _ = c.errYell.Fprintln(os.Stderr, fmt.Sprintf("Parse warnings:  %d", parseErrors))
	}
	_, _ = c.green.Println(fmt.Sprintf("Output files:    %d", written))
}

// SetVerbose enables or disables verbose output.
func (c *Console) SetVerbose(v bool) { c.verbose = v }

// Verbose prints a cyan message to stderr only when verbose mode is enabled.
func (c *Console) Verbose(msg string) {
	if c.verbose {
		_, _ = c.cyan.Fprintln(os.Stderr, msg)
	}
}

// Progress prints a throttled progress update to stderr on a single line.
// Uses \r to overwrite the previous output. Only active in verbose mode.
func (c *Console) Progress(msg string) {
	if !c.verbose {
		return
	}
	c.lastMsg = msg
	now := time.Now()
	if now.Sub(c.lastProgress) < 100*time.Millisecond {
		return
	}
	c.lastProgress = now
	_, _ = os.Stderr.WriteString("\r\033[K" + msg)
}

// ProgressDone flushes the final progress message and moves to a new line.
func (c *Console) ProgressDone() {
	if !c.verbose {
		return
	}
	if c.lastMsg != "" {
		_, _ = os.Stderr.WriteString("\r\033[K" + c.lastMsg)
		c.lastMsg = ""
	}
	_, _ = os.Stderr.WriteString("\n")
	_ = os.Stderr.Sync()
}
