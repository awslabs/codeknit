// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cpp implements a LanguagePlugin for C++ files
// using tree-sitter for parsing.
package cpp

import (
	"codeknit/internal/plugin"

	tscpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for C++ files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"declaration", "init_declarator", "assignment_expression"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier", "declarator"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"type_identifier"},
		CallTarget:      callTarget,
		RichCallTarget:  richCallTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".cpp", ".hpp", ".cc", ".cxx", ".hxx"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"_test"}},
			TSLang:     tscpp.Language(),
			Handlers:   handlers,
			TokenMap:   bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"call_expression"},
				ArgListKinds:  []string{"argument_list"},
				IdentKinds:    []string{"identifier"},
			},
			Dataflow: &df,
		}),
	}
}
