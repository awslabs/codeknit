// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"time"

	"codeknit/internal/browser"
	"codeknit/internal/config"
	"codeknit/internal/console"
	"codeknit/internal/emitter"
	"codeknit/internal/pipeline"
	"codeknit/internal/plugin"
	"codeknit/internal/plugin/clang"
	"codeknit/internal/plugin/cpp"
	"codeknit/internal/plugin/csharp"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/php"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/ruby"
	"codeknit/internal/plugin/rust"
	"codeknit/internal/plugin/scala"
	"codeknit/internal/plugin/typescript"
	"codeknit/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		con := console.New()
		con.Error(err.Error())
		os.Exit(1)
	}
}

// commandRunner is implemented by anything the TUI can produce — each
// per-command config picks its own pipeline. This replaces the old
// `switch cfg.Command` dispatch with a typed choice.
type commandRunner func(con *console.Console) error

// runTUI launches the interactive TUI and returns a commandRunner bound to
// the command-specific config the user built.
func runTUI() (commandRunner, error) {
	m := tui.NewModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	fm, ok := finalModel.(tui.Model)
	if !ok {
		return nil, fmt.Errorf("unexpected TUI model type")
	}
	if !fm.Confirmed() {
		return nil, fmt.Errorf("configuration canceled")
	}

	switch fm.SelectedCommand() {
	case tui.CmdGraphShow:
		cfg := fm.ToGraphConfig()
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return func(con *console.Console) error { return runGraphShow(&cfg, con) }, nil

	case tui.CmdGraphAnalyze:
		cfg := fm.ToAnalyzeConfig()
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return func(con *console.Console) error { return runGraphAnalyze(&cfg, con) }, nil

	case tui.CmdFingerprint:
		cfg := fm.ToFingerprintConfig()
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return func(con *console.Console) error { return runFingerprint(&cfg, con) }, nil

	case tui.CmdParse:
		cfg := fm.ToParseConfig()
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return func(con *console.Console) error { return runParse(&cfg, con) }, nil

	default:
		// Unreachable: SelectedCommand() always returns a valid kind.
		return nil, fmt.Errorf("unknown command kind")
	}
}

// newRegistry builds the plugin registry with all supported languages.
func newRegistry() (*plugin.Registry, time.Duration) {
	t := time.Now()
	registry := plugin.NewRegistry()
	registry.Register(typescript.NewPlugin())
	registry.Register(javascript.NewPlugin())
	registry.Register(clang.NewPlugin())
	registry.Register(cpp.NewPlugin())
	registry.Register(csharp.NewPlugin())
	registry.Register(golang.NewPlugin())
	registry.Register(java.NewPlugin())
	registry.Register(php.NewPlugin())
	registry.Register(python.NewPlugin())
	registry.Register(ruby.NewPlugin())
	registry.Register(rust.NewPlugin())
	registry.Register(scala.NewPlugin())
	return registry, time.Since(t)
}

// progressCallbacks returns a pair of scan/parse progress printers for the console.
func progressCallbacks(con *console.Console) (onScan func(visited, matched int), onParse func(done, total int)) {
	return func(visited, matched int) {
			con.Progress(fmt.Sprintf("Scanning... %d entries visited, %d files matched", visited, matched))
		},
		func(done, total int) {
			con.Progress(fmt.Sprintf("Parsing... %d/%d", done, total))
		}
}

// buildGraphResult runs the shared scan→parse pipeline and prints any warnings.
// Returns the GraphResult or an error.
func buildGraphResult(cfg config.Common, con *console.Console) (*pipeline.GraphResult, error) {
	registry, initDur := newRegistry()
	onScan, onParse := progressCallbacks(con)
	gr, err := pipeline.BuildGraph(pipeline.BuildOptions{
		InputPath:   cfg.InputPath,
		Workers:     cfg.Workers,
		CollectTest: cfg.CollectTest,
	}, registry, initDur, onScan, onParse)
	con.ProgressDone()
	if err != nil {
		return nil, err
	}
	for i := range gr.Skipped {
		con.Warn(gr.Skipped[i].Error())
	}
	for i := range gr.ParseErrors {
		con.Warn(gr.ParseErrors[i].Error())
	}
	return gr, nil
}

// runGraphShow executes the graph-show pipeline.
func runGraphShow(cfg *config.GraphConfig, con *console.Console) error {
	con.SetVerbose(cfg.Verbose)
	gr, err := buildGraphResult(cfg.Common, con)
	if err != nil {
		return err
	}

	e := &emitter.Emitter{}
	if err := e.EmitGraph(gr.Graph, cfg.Output); err != nil {
		return err
	}

	con.Success(fmt.Sprintf("Graph written to %s (%d symbols, %d edges, %d files)",
		cfg.Output, len(gr.Graph.Symbols), len(gr.Graph.Edges), len(gr.Files)))

	if err := browser.Open(cfg.Output); err != nil {
		con.Warn(fmt.Sprintf("Could not open browser: %v", err))
	}

	return nil
}

// runGraphAnalyze executes the graph-analyze pipeline.
func runGraphAnalyze(cfg *config.AnalyzeConfig, con *console.Console) error {
	con.SetVerbose(cfg.Verbose)
	gr, err := buildGraphResult(cfg.Common, con)
	if err != nil {
		return err
	}

	e := &emitter.Emitter{}
	if err := e.EmitGraphAnalysis(gr.Graph, &emitter.AnalysisOptions{
		OutputPath:           cfg.Output,
		FanThreshold:         cfg.FanThreshold,
		GodThreshold:         cfg.GodThreshold,
		MaxInheritanceDepth:  cfg.MaxInheritanceDepth,
		TopN:                 cfg.TopN,
		BetweennessThreshold: cfg.BetweennessThreshold,
		PropagationCutoff:    cfg.PropagationCutoff,
	}); err != nil {
		return err
	}

	con.Success(fmt.Sprintf("Analysis written to %s (%d symbols, %d edges, %d files)",
		cfg.Output, len(gr.Graph.Symbols), len(gr.Graph.Edges), len(gr.Files)))

	return nil
}

// runParse executes the full parse pipeline and prints the summary.
func runParse(cfg *config.ParseConfig, con *console.Console) error {
	con.SetVerbose(cfg.Verbose)
	registry, initDur := newRegistry()

	onScan, onParse := progressCallbacks(con)
	res, err := pipeline.Execute(cfg, registry, initDur, onScan, onParse)
	con.ProgressDone()
	if err != nil {
		return err
	}

	for i := range res.Skipped {
		con.Warn(res.Skipped[i].Error())
	}
	for i := range res.ParseErrors {
		con.Warn(res.ParseErrors[i].Error())
	}

	if cfg.OutputMode == config.OutputInline {
		t := res.Timings
		con.Verbose(fmt.Sprintf("\nTiming: init %s, scan %s, parse %s, plan %s, emit %s (total %s)",
			formatDur(t.Init), formatDur(t.Scan), formatDur(t.Parse), formatDur(t.Plan), formatDur(t.Emit), formatDur(t.Total)))
		return nil
	}

	c := res.Counts
	con.Summary(c.FilesProcessed, c.Skipped, c.ParseErrors, c.OutputFiles)
	con.Success(fmt.Sprintf("Output written to %s", cfg.OutputDir))

	t := res.Timings
	con.Verbose(fmt.Sprintf("\nTiming: init %s, scan %s, parse %s, plan %s, emit %s (total %s)",
		formatDur(t.Init), formatDur(t.Scan), formatDur(t.Parse), formatDur(t.Plan), formatDur(t.Emit), formatDur(t.Total)))

	return nil
}

// formatDur formats a duration as a human-friendly string.
func formatDur(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
