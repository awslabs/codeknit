// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package golang implements a LanguagePlugin for Go files
// using tree-sitter for parsing.
package golang

import (
	"codeknit/internal/plugin"

	tsgo "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .go files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"short_var_declaration", "assignment_statement", "var_spec"},
		ObjectPairKinds: []string{"keyed_element", "literal_element"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier", "expression_list"},
		ValueChildKinds: []string{"identifier", "expression_list"},
		TypeRefKinds:    []string{"type_identifier"},
		CallTarget:      callTarget,
		RichCallTarget:  richCallTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".go"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"_test"}},
			TSLang:     tsgo.Language(),
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
