// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

// FingerprintParser is implemented by parsers that also extract body tokens.
type FingerprintParser interface {
	ParseWithOptions(filePath string, fingerprint bool) ([]Symbol, []Edge, error)
}

// ParseFile preserves optional fingerprint support when composing parsers.
func ParseFile(p LanguagePlugin, filePath string, fingerprint bool) (symbols []Symbol, edges []Edge, err error) {
	if fingerprint {
		if fp, ok := p.(FingerprintParser); ok {
			return fp.ParseWithOptions(filePath, true)
		}
	}
	return p.Parse(filePath)
}
