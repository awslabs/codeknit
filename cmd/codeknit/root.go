// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the codeknit CLI.
package main

import (
	"fmt"
	"os"

	"codeknit/internal/config"
	"codeknit/internal/console"
	"codeknit/internal/scanpolicy"

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
	output         string
	headerLanguage string
	scan           scanpolicy.Options
	workers        int
	collectTest    bool
	verbose        bool
}

func (o *commonOptions) commonConfig(inputPath string) config.Common {
	return config.Common{
		Scan:      o.scan,
		InputPath: inputPath, HeaderLanguage: o.headerLanguage,
		Workers: o.workers, CollectTest: o.collectTest, Verbose: o.verbose,
	}
}

func (o *commonOptions) addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.scan.Include, "include", nil, "include source paths matching a glob (repeatable, relative to input)")
	cmd.Flags().StringArrayVar(&o.scan.Exclude, "exclude", nil, "exclude source paths matching a glob (repeatable, relative to input)")
	cmd.Flags().Int64Var(&o.scan.MaxFileSize, "max-file-size", scanpolicy.DefaultMaxFileSize, "maximum source file size in bytes (0 disables the limit)")
	cmd.Flags().BoolVar(&o.scan.Strict, "strict", false, "fail before writing output if any scan or parse diagnostics occur")
	cmd.Flags().StringVar(&o.headerLanguage, "header-language", config.DefaultHeaderLanguage,
		"parser for .h files: c, cpp")
	cmd.Flags().BoolVar(&o.collectTest, "collect-test", config.DefaultCollectTest, "include test files in analysis")
	cmd.Flags().IntVar(&o.workers, "workers", config.DefaultWorkers, "max concurrent parsing goroutines (0 = NumCPU)")
	cmd.Flags().BoolVar(&o.verbose, "verbose", config.DefaultVerbose, "print progress information during processing")
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
