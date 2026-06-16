// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package java implements a LanguagePlugin for Java files
// using tree-sitter for parsing.
package java

import (
	"codeknit/internal/plugin"

	tsjava "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .java files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"local_variable_declaration", "assignment_expression"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier", "variable_declarator"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"type_identifier"},
		CallTarget:      callTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".java"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"Test", "Tests", "Spec", "Suite"}},
			TSLang:     tsjava.Language(),
			Handlers:   handlers,
			TokenMap:   bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"method_invocation"},
				ArgListKinds:  []string{"argument_list"},
				IdentKinds:    []string{"identifier"},
			},
			Dataflow: &df,
		}),
	}
}
