// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package php implements a LanguagePlugin for PHP files
// using tree-sitter for parsing.
package php

import (
	"codeknit/internal/plugin"

	tsphp "github.com/tree-sitter/tree-sitter-php/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .php files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"assignment_expression", "simple_parameter"},
		ObjectPairKinds: []string{"pair"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier", "name"},
		NameChildKinds:  []string{"variable_name", "name"},
		ValueChildKinds: []string{"identifier", "name"},
		TypeRefKinds:    []string{"named_type"},
		CallTarget:      callTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".php"},
			TestConf:   plugin.TestConfig{NameSuffixes: []string{"Test", "Spec"}},
			TSLang:     tsphp.LanguagePHP(),
			Handlers:   handlers,
			TokenMap:   bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"function_call_expression", "member_call_expression", "scoped_call_expression"},
				ArgListKinds:  []string{"arguments"},
				IdentKinds:    []string{"identifier", "name"},
			},
			Dataflow: &df,
		}),
	}
}
