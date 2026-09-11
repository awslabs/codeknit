// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMaintainedReferenceSurvivesGeneration(t *testing.T) {
	t.Chdir("../..")
	const output = "reference/language-support.md"
	source, err := os.ReadFile("docs/reference/language-support.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, initial := range []string{"", "stale generated content"} {
		t.Run(initial, func(t *testing.T) {
			contentDir := t.TempDir()
			outPath := filepath.Join(contentDir, output)
			if initial != "" {
				if err := writeFile(outPath, initial); err != nil {
					t.Fatal(err)
				}
			}
			// Keep an ordinary generated page cached to exercise both modes.
			if err := writeFile(filepath.Join(contentDir, "reference", "cli-flags.md"), "cached"); err != nil {
				t.Fatal(err)
			}
			for _, force := range []bool{false, true} {
				if err := syncSourcePages(contentDir); err != nil {
					t.Fatal(err)
				}
				cliPending := false
				for _, p := range pagesToGenerate(contentDir, force) {
					if p.Output == output {
						t.Fatalf("force=%t: maintained reference would be rewritten by the model", force)
					}
					cliPending = cliPending || p.Output == "reference/cli-flags.md"
				}
				if cliPending != force {
					t.Fatalf("force=%t: generated page cache behavior changed", force)
				}
				got, err := os.ReadFile(filepath.Clean(outPath))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, source) {
					t.Fatalf("force=%t: maintained reference was not restored verbatim", force)
				}
			}
		})
	}
}

func TestMaintainedReferencesMatchPublishedCopies(t *testing.T) {
	t.Chdir("../..")
	for _, p := range pages {
		if p.Source == "" {
			continue
		}
		source, err := os.ReadFile(filepath.Clean(p.Source))
		if err != nil {
			t.Fatal(err)
		}
		published, err := os.ReadFile(filepath.Join("docs", "src", "content", "docs", p.Output))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(source, published) {
			t.Errorf("%s differs from maintained source %s; synchronize the output copy", p.Output, p.Source)
		}
	}
}
