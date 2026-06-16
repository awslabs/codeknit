// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the codeknit CLI.
package main

import (
	"fmt"
	"os"

	"codeknit/internal/console"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// commonOptions holds the flags shared by all subcommands.
type commonOptions struct {
	output      string
	workers     int
	collectTest bool
	verbose     bool
}

func newRootCmd() *cobra.Command {
	con := console.New()

	root := &cobra.Command{
		Use:   "codeknit",
		Short: "codeknit — static code structure extractor",
		Long: `codeknit parses source code and extracts structural information
(functions, classes, methods, relationships) into a compact intermediate
representation suitable for LLM consumption.

Running "codeknit" with no arguments launches the interactive terminal UI.
Use "codeknit parse" or "codeknit graph" for direct CLI usage.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	root.SetVersionTemplate("codeknit {{.Version}}\n")

	// When invoked with no subcommand and no args, launch the TUI.
	// Only "codeknit --help" / "codeknit -h" shows the help menu.
	root.RunE = func(cmd *cobra.Command, args []string) error {
		runner, err := runTUI()
		if err != nil {
			return err
		}
		return runner(con)
	}

	root.AddCommand(newParseCmd(con))
	root.AddCommand(newGraphCmd(con))
	root.AddCommand(newFingerprintCmd(con))

	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	return root
}
