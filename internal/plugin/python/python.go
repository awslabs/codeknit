// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package python implements a LanguagePlugin for Python files
// using tree-sitter for parsing.
package python

import (
	"codeknit/internal/plugin"

	tspython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .py and .pyi files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"assignment", "expression_statement"},
		ObjectPairKinds: []string{"pair"},
		ReturnKinds:     []string{"return_statement"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"type"},
		CallTarget:      callTarget,
		RichCallTarget:  richCallTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".py", ".pyi"},
			TestConf: plugin.TestConfig{
				NamePrefixes: []string{"test_"},
				NameSuffixes: []string{"_test"},
			},
			TSLang:   tspython.Language(),
			Handlers: handlers,
			TokenMap: bodyTokenMap,
			CallableRef: plugin.CallableRefConfig{
				CallNodeKinds: []string{"call"},
				ArgListKinds:  []string{"argument_list"},
				IdentKinds:    []string{"identifier"},
			},
			Dataflow: &df,
		}),
	}
}
