// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_RelationshipAccuracy(t *testing.T) {
	root := t.TempDir()
	for name, source := range map[string]string{
		"main.go":         "package app\nimport worker \"./service\"\nfunc main() { worker.Work(); Helper() }\n",
		"a.go":            "package app\nfunc Helper() {}\n",
		"b.go":            "package app\nfunc Helper() {}\n",
		"service/work.go": "package service\nfunc Work() {}\n",
		"parser.ts":       "export function parse() {}\n",
		"unrelated.ts":    "export function format() {}\n",
		"app.ts": "import { parse as read } from './parser';\n" +
			"import { format } from 'external-format';\n" +
			"function run() { read(); format(); }\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := runcodeknitInline(t, root, "parse", root, "--output-mode", "inline", "--format", "json", "--edges")
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Edges []struct {
			From, To, Kind, Resolution string
			ToShort                    string `json:"to_short"`
			Candidates                 []string
		}
	}
	if decodeErr := json.Unmarshal([]byte(raw), &output); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	found := make(map[string]bool)
	for _, e := range output.Edges {
		if e.Kind != "calls" {
			continue
		}
		switch {
		case strings.HasSuffix(e.To, "service/work.go::Work"):
			found["qualified"] = e.Resolution == "" && e.ToShort != ""
		case strings.HasSuffix(e.To, "parser.ts::parse"):
			found["renamed"] = e.Resolution == "" && e.ToShort != ""
		case strings.HasSuffix(e.To, "main.go::Helper"):
			found["ambiguous"] = e.Resolution == "ambiguous" && len(e.Candidates) == 2 && e.ToShort == ""
		case strings.HasSuffix(e.To, "app.ts::format"):
			found["external"] = e.Resolution == "unresolved" && e.ToShort == ""
		default:
			t.Errorf("unexpected call: %+v", e)
		}
	}
	for _, name := range []string{"qualified", "renamed", "ambiguous", "external"} {
		if !found[name] {
			t.Errorf("missing %s relationship: %s", name, raw)
		}
	}
	skt, sktErr := runcodeknitInline(t, root, "parse", root, "--output-mode", "inline", "--edges")
	if sktErr != nil {
		t.Fatal(sktErr)
	}
	for _, notation := range []string{
		"--calls[unresolved]--> format",
		"--calls[ambiguous]--> Helper [candidates=S",
	} {
		if !strings.Contains(skt, notation) {
			t.Errorf("missing native SKT notation %q:\n%s", notation, skt)
		}
	}
	if strings.Contains(skt, "? {") {
		t.Errorf("mixed JSON record in SKT:\n%s", skt)
	}
	reportPath := filepath.Join(t.TempDir(), "analysis.skt")
	if _, runErr := runcodeknitInline(t, root, "graph", "analyze", root, "--output", reportPath, "--top-n", "0"); runErr != nil {
		t.Fatal(runErr)
	}
	report, err := os.ReadFile(reportPath) //nolint:gosec // reportPath is a file created by this test in t.TempDir.
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(report), "unresolved=1 | ambiguous=1") {
		t.Fatalf("analysis did not account for uncertainty:\n%s", report)
	}
}
