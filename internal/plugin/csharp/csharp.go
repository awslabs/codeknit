// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package csharp implements a LanguagePlugin for C# files
// using tree-sitter for parsing.
package csharp

import (
	"codeknit/internal/plugin"

	tscsharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .cs files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"variable_declaration", "assignment_expression"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier", "variable_declarator"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"identifier_name"},
		CallTarget:      callTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".cs"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"Test", "Tests", "Spec"}},
			TSLang:     tscsharp.Language(),
			Handlers:   handlers,
			TokenMap:   bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"invocation_expression"},
				ArgListKinds:  []string{"argument_list"},
				IdentKinds:    []string{"identifier"},
			},
			Dataflow: &df,
		}),
	}
}
