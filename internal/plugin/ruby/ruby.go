// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package ruby implements a LanguagePlugin for Ruby files
// using tree-sitter for parsing.
package ruby

import (
	"codeknit/internal/plugin"

	tsruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .rb files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	df := plugin.DataflowConfig{
		AssignmentKinds: []string{"assignment"},
		ObjectPairKinds: []string{"pair"},
		ReturnKinds:     []string{"return"},
		IdentKinds:      []string{"identifier"},
		NameChildKinds:  []string{"identifier"},
		ValueChildKinds: []string{"identifier"},
		TypeRefKinds:    []string{"constant"},
		CallTarget:      callTarget,
		RichCallTarget:  richCallTarget,
	}

	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions: []string{".rb"},
			TestConf: plugin.TestConfig{
				NamePrefixes: []string{"test_"},
				NameSuffixes: []string{"_test", "_spec"},
			},
			TSLang:   tsruby.Language(),
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
