// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package typescript implements a LanguagePlugin for TypeScript and TSX files
// using tree-sitter for parsing.
package typescript

import (
	"codeknit/internal/common/jsshared"
	"codeknit/internal/plugin"

	tstypescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .ts and .tsx files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions:  []string{".ts", ".tsx"},
			TestConf:    plugin.TestConfig{ContainsDot: []string{".test.", ".spec."}},
			TSLang:      tstypescript.LanguageTypescript(),
			Handlers:    handlers,
			TokenMap:    bodyTokenMap,
			CallableRef: jsshared.JSCallableRefConfig,
			Dataflow:    ptrTo(jsshared.JSDataflowConfig([]string{"type_identifier"}, callTarget, richCallTarget)),
		}),
	}
}

func ptrTo(d plugin.DataflowConfig) *plugin.DataflowConfig { return &d } //nolint:gocritic // intentional copy to take address
