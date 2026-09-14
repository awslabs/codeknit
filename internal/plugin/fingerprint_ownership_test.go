// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"bytes"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
)

func TestFingerprintCallableOwnership(t *testing.T) {
	cases := []struct {
		parser plugin.SourceParser
		file   string
		source string
	}{
		{golang.NewPlugin(), "app.go", `package app
func Compact(x int) int { if x > 0 { x += 2 }; for x < 10 { x++ }; return x }
func Expanded(x int) int {
	if x > 0 { x += 2 }
	for x < 10 { x++ }
	return x
}
func Other(x int) int { return x * 9 }
`},
		{javascript.NewPlugin(), "app.js", `function Compact(x) { if (x > 0) { x += 2; } for (; x < 10;) { x++; } return x; } function Other(x) { return x * 9; }
function Expanded(x) {
	if (x > 0) { x += 2; }
	for (; x < 10;) { x++; }
	return x;
}`},
		{java.NewPlugin(), "App.java", `class App {
int Compact(int x) { if (x > 0) { x += 2; } for (; x < 10;) { x++; } return x; } int Other(int x) { return x * 9; }
int Expanded(int x) {
	if (x > 0) { x += 2; }
	for (; x < 10;) { x++; }
	return x;
}
}`},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			symbols, _, err := tc.parser.ParseSource(tc.file, []byte(tc.source), true)
			if err != nil {
				t.Fatal(err)
			}
			tokens := make(map[string][]byte)
			for _, symbol := range symbols {
				if symbol.Category == plugin.CategoryCallable {
					tokens[symbol.Name] = symbol.BodyTokens
				}
			}
			if len(tokens["Compact"]) < 4 || !bytes.Equal(tokens["Compact"], tokens["Expanded"]) {
				t.Fatalf("formatting changed callable body: compact=%x expanded=%x", tokens["Compact"], tokens["Expanded"])
			}
			if len(tokens["Other"]) == 0 || bytes.Equal(tokens["Other"], tokens["Compact"]) {
				t.Fatalf("same-line neighbor has wrong body: %x", tokens["Other"])
			}
		})
	}
}
