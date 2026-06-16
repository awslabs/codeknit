// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package javascript implements a LanguagePlugin for JavaScript and JSX files
// using tree-sitter for parsing.
package javascript

import (
	"codeknit/internal/common/jsshared"
	"codeknit/internal/plugin"

	tsjavascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
)

// Plugin implements plugin.LanguagePlugin for .js and .jsx files.
type Plugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a Plugin with the shared base configuration.
func NewPlugin() *Plugin {
	return &Plugin{
		BasePlugin: plugin.NewBasePluginFromConfig(&plugin.Config{
			Extensions:  []string{".js", ".jsx"},
			TestConf:    plugin.TestConfig{ContainsDot: []string{".test.", ".spec."}},
			TSLang:      tsjavascript.Language(),
			Handlers:    handlers,
			TokenMap:    bodyTokenMap,
			CallableRef: jsshared.JSCallableRefConfig,
			Dataflow:    ptrTo(jsshared.JSDataflowConfig(nil, callTarget, richCallTarget)),
		}),
	}
}

func ptrTo(d plugin.DataflowConfig) *plugin.DataflowConfig { return &d } //nolint:gocritic // intentional copy to take address
