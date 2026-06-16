// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package rust implements a LanguagePlugin for Rust files
// using tree-sitter for parsing.
package rust

import (
	"codeknit/internal/plugin"

	tsrust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .rs files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"let_declaration", "assignment_expression"},
		ReturnKinds:     []string{"return_expression"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"type_identifier"},
		CallTarget:      callTarget,
		RichCallTarget:  richCallTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".rs"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"_test"}},
			TSLang:     tsrust.Language(),
			Handlers:   handlers,
			TokenMap:   bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"call_expression"},
				ArgListKinds:  []string{"arguments"},
				IdentKinds:    []string{"identifier"},
			},
			Dataflow: &df,
		}),
	}
}
