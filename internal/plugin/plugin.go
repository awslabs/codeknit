// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"codeknit/internal/common/extract"
)

// Collector accumulates Symbols and Edges during AST extraction.
type Collector = extract.Collector

// ExtractFunc is the language-specific function that walks the tree-sitter AST.
type ExtractFunc = extract.Func

// SyntaxError indicates that a file had syntax errors but partial results were extracted.
type SyntaxError = extract.SyntaxError

// CallTargetFunc extracts the call target name from a call-expression node.
type CallTargetFunc = extract.CallTargetFunc

// RichCallTargetFunc extracts a CallTargetResult from a call-expression node.
type RichCallTargetFunc = extract.RichCallTargetFunc

// CallTargetResult holds the extracted call target name and whether it was qualified.
type CallTargetResult = extract.CallTargetResult

// CallableRefConfig describes the language-specific node kinds for callable reference detection.
type CallableRefConfig = extract.CallableRefConfig

// DataflowConfig describes the language-specific node kinds for dataflow hint extraction.
type DataflowConfig = extract.DataflowConfig

// DispatchTable maps tree-sitter node kinds to their Handler functions.
type DispatchTable = extract.DispatchTable

// Handler is the unified function signature for dispatch table entries.
type Handler = extract.Handler

// HandlerContext carries contextual information passed through the dispatch table.
type HandlerContext = extract.HandlerContext

// ParamConfig describes how to locate and destructure parameter nodes.
type ParamConfig = extract.ParamConfig

// DecoratorNameFunc extracts a decorator/annotation name from a decorator node.
type DecoratorNameFunc = extract.DecoratorNameFunc

// Re-export extract functions so language plugins can keep using plugin.Adapt etc.
var (
	Adapt                           = extract.Adapt
	AdaptParent                     = extract.AdaptParent
	AdaptExported                   = extract.AdaptExported
	WalkTopLevel                    = extract.WalkTopLevel
	WalkChildren                    = extract.WalkChildren
	MakeExtractFuncWithCallableRefs = extract.MakeExtractFuncWithCallableRefs
	MakeExtractFuncWithFingerprint  = extract.MakeExtractFuncWithFingerprint
	ParseWithTreeSitter             = extract.ParseWithTreeSitter
	UnqualifiedCallTarget           = extract.UnqualifiedCallTarget
	UnqualifiedCallTargetRich       = extract.UnqualifiedCallTargetRich
	FilterCallTarget                = extract.FilterCallTarget
	LastNamedLeaf                   = extract.LastNamedLeaf
	ExtractCallEdges                = extract.CallEdges
	ExtractCallableRefEdges         = extract.CallableRefEdges
	ExtractDataflowHints            = extract.DataflowHints
	ExtractFileTypeRefs             = extract.FileTypeRefs
	ExtractFileCallEdges            = extract.FileCallEdges
	ExtractTypeRefEdges             = extract.TypeRefEdges
	ExtractInclude                  = extract.Include
	ExtractDecoratorEdges           = extract.DecoratorEdges
	MakeDecoratorNameFunc           = extract.MakeDecoratorNameFunc
	ExtractTypedParams              = extract.TypedParams
	ReturnTypeByKinds               = extract.ReturnTypeByKinds
	ReturnTypeAfterToken            = extract.ReturnTypeAfterToken
	FirstChildTextByKinds           = extract.FirstChildTextByKinds
	ChildByKind                     = extract.ChildByKind
	ChildText                       = extract.ChildText
	FindFirstError                  = extract.FindFirstError
	NodeSpan                        = extract.NodeSpan
	NodeByteSpan                    = extract.NodeByteSpan
	BoolStr                         = extract.BoolStr
	BuildFuncSignature              = extract.BuildFuncSignature
	HasChildKeyword                 = extract.HasChildKeyword
	LastSepIndex                    = extract.LastSepIndex
)

// SortedStringKeys returns the keys of any map[string]V sorted lexicographically.
func SortedStringKeys[V any](m map[string]V) []string {
	return extract.SortedStringKeys(m)
}

// LanguagePlugin defines the interface each language parser must implement.
type LanguagePlugin interface {
	Extensions() []string
	TestPatterns() TestConfig
	Parse(filePath string) ([]Symbol, []Edge, error)
}

// SourceParser accepts bytes already read within the pipeline's scan boundary.
// Plugins can implement this alongside LanguagePlugin to avoid reopening files.
type SourceParser interface {
	ParseSource(filePath string, source []byte, fingerprint bool) ([]Symbol, []Edge, error)
}

// TestConfig describes how to detect test files for a language.
type TestConfig struct {
	NameSuffixes []string
	NamePrefixes []string
	ContainsDot  []string
}

// BasePlugin provides the shared Parse, Extensions, and TestPatterns
// implementations used by every language plugin.
type BasePlugin struct {
	LanguageByExtension map[string]unsafe.Pointer
	Exts                []string
	Extract             ExtractFunc
	TSLang              unsafe.Pointer
	TestConf            TestConfig
}

// Config holds everything needed to create a BasePlugin.
type Config struct {
	LanguageByExtension map[string]unsafe.Pointer
	Handlers            DispatchTable
	TokenMap            map[string]byte
	TSLang              unsafe.Pointer
	Dataflow            *DataflowConfig
	Extensions          []string
	CallableRef         CallableRefConfig
	TestConf            TestConfig
}

// NewBasePluginFromConfig creates a BasePlugin from a Config.
func NewBasePluginFromConfig(cfg *Config) BasePlugin {
	var dfCfgs []DataflowConfig
	if cfg.Dataflow != nil {
		dfCfgs = append(dfCfgs, *cfg.Dataflow)
	}

	var extractFn ExtractFunc
	if cfg.TokenMap != nil {
		extractFn = MakeExtractFuncWithFingerprint(cfg.Handlers, cfg.CallableRef, cfg.TokenMap, dfCfgs...)
	} else {
		extractFn = MakeExtractFuncWithCallableRefs(cfg.Handlers, cfg.CallableRef, dfCfgs...)
	}

	return BasePlugin{
		LanguageByExtension: cfg.LanguageByExtension,
		Exts:                cfg.Extensions,
		TestConf:            cfg.TestConf,
		TSLang:              cfg.TSLang,
		Extract:             extractFn,
	}
}

// Extensions returns the file extensions this plugin handles.
func (b *BasePlugin) Extensions() []string { return b.Exts }

// TestPatterns returns the test file detection rules for this language.
func (b *BasePlugin) TestPatterns() TestConfig { return b.TestConf }

// Parse reads the file and delegates to ParseWithTreeSitter.
func (b *BasePlugin) Parse(filePath string) (symbols []Symbol, edges []Edge, err error) {
	return b.ParseWithOptions(filePath, false)
}

// ParseWithOptions reads the file and delegates to ParseWithTreeSitter with optional fingerprint extraction.
func (b *BasePlugin) ParseWithOptions(filePath string, fingerprint bool) (symbols []Symbol, edges []Edge, err error) {
	src, err := os.ReadFile(filePath) //nolint:gosec // filePath is from the scanner
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", filePath, err)
	}
	return b.ParseSource(filePath, src, fingerprint)
}

// ParseSource extracts declarations from a source snapshot without filesystem access.
func (b *BasePlugin) ParseSource(filePath string, src []byte, fingerprint bool) (symbols []Symbol, edges []Edge, err error) {
	lang := b.TSLang
	if variant, ok := b.LanguageByExtension[filepath.Ext(filePath)]; ok {
		lang = variant
	}
	symbols, edges, err = ParseWithTreeSitter(filePath, src, lang, b.Extract, fingerprint)
	if fingerprint {
		lines := strings.Split(string(src), "\n")
		for i := range symbols {
			start, end := symbols[i].Span[0]-1, symbols[i].Span[1]
			if start >= 0 && start < len(lines) && end > start {
				symbols[i].SourceBody = strings.Join(lines[start:min(end, len(lines))], "\n")
			}
		}
	}
	return symbols, edges, err
}

// Registry holds registered language plugins keyed by extension.
type Registry struct {
	plugins map[string]LanguagePlugin
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]LanguagePlugin)}
}

// Register adds a language plugin to the registry for each of its extensions.
func (r *Registry) Register(p LanguagePlugin) {
	r.RegisterExtensions(p, p.Extensions()...)
}

// RegisterExtensions selects a parser for specific extensions without changing
// that parser's configuration or its registrations for other extensions.
func (r *Registry) RegisterExtensions(p LanguagePlugin, extensions ...string) {
	for _, ext := range extensions {
		r.plugins[ext] = p
	}
}

// Lookup returns the plugin registered for the given extension.
func (r *Registry) Lookup(ext string) (LanguagePlugin, bool) {
	p, ok := r.plugins[ext]
	return p, ok
}

// AllExtensions returns all registered file extensions in sorted order.
func (r *Registry) AllExtensions() []string {
	return extract.SortedStringKeys(r.plugins)
}

// AllTestPatterns returns a map from file extension to its TestConfig.
func (r *Registry) AllTestPatterns() map[string]TestConfig {
	result := make(map[string]TestConfig, len(r.plugins))
	for ext, p := range r.plugins {
		result[ext] = p.TestPatterns()
	}
	return result
}
