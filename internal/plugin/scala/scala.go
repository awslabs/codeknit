// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package scala implements a LanguagePlugin for Scala files
// using tree-sitter for parsing.
package scala

import (
	"codeknit/internal/plugin"

	tsscala "github.com/tree-sitter/tree-sitter-scala/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .scala and .sc files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"val_definition", "var_definition"},
		ReturnKinds:     []string{"return_expression"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"type_identifier"},
		CallTarget:      callTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".scala", ".sc"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"Test", "Tests", "Spec", "Suite"}},
			TSLang:     tsscala.Language(),
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
