// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// Every generated page must declare its Diátaxis type and brief so system.txt
// can apply the matching principles.
func TestPromptsDeclareDiataxisType(t *testing.T) {
	t.Chdir("../..")
	valid := map[string]bool{"Tutorial": true, "How-to guide": true, "Reference": true, "Explanation": true, "Landing page": true}
	for _, p := range pages {
		if p.Prompt == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join("docs/prompts", p.Prompt+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.SplitN(string(data), "\n", 5)
		if len(lines) < 5 {
			t.Fatalf("%s: prompt too short for the Diátaxis header", p.Prompt)
		}
		for i, key := range []string{"DOCUMENT TYPE: ", "AUDIENCE: ", "USER GOAL: ", "SCOPE: "} {
			value, ok := strings.CutPrefix(lines[i], key)
			if !ok || strings.TrimSpace(value) == "" {
				t.Fatalf("%s: line %d must start with %q and have a value", p.Prompt, i+1, key)
			}
			if i == 0 && !valid[value] {
				t.Fatalf("%s: unknown DOCUMENT TYPE %q", p.Prompt, value)
			}
		}
	}
}

func TestValidateTranslationLinks(t *testing.T) {
	src := "---\ntitle: x\n---\n\nSee [a](../a/) and [b](../b/).\n"
	cases := map[string]bool{
		"---\ntitle: y\n---\n\nVoir [b](../b/) et [a](../a/).\n":           true,  // reordered is fine
		"---\ntitle: y\n---\n\nVoir [a](../fr/a/) et [b](../b/).\n":        false, // rewritten target
		"---\ntitle: y\n---\n\nVoir [a](../a/).\n":                         false, // dropped link
		"---\ntitle: y\n---\n\nVoir [a](../a/) [b](../b/) [c](../c/).\n":   false, // added link
		"---\ntitle: y\n---\n\nVoir [a](/codeknit/fr/a/) et [b](../b/).\n": false, // relative rewritten to absolute
		"---\ntitle: y\n---\n\nVoir [a](../a/) et [b](../b/#ancre).\n":     false, // anchor added
	}
	anchored := "---\ntitle: x\n---\n\nSee [a](../a/#some-heading).\n"
	if w := validateTranslation(anchored, "---\ntitle: y\n---\n\nVoir [a](../a/#un-titre).\n", "fr"); len(w) > 0 {
		t.Errorf("translated cross-page anchor rejected: %v", w)
	}
	for translated, ok := range cases {
		if got := len(validateTranslation(src, translated, "fr")) == 0; got != ok {
			t.Errorf("validateTranslation(%q) valid=%t, want %t", translated, got, ok)
		}
	}
}

func TestBrokenLinks(t *testing.T) {
	content := "[ok](../../reference/output-format/) [anchor](../../reference/cli-flags/#parse) [self](#x) " +
		"[ext](https://example.com) [bad1](/reference/cli/) [bad2](../../reference/cli/) [abs](/codeknit/installation/)"
	got := brokenLinks("guides/parse-command.md", content)
	want := []string{"/reference/cli/", "../../reference/cli/"}
	if !slices.Equal(got, want) {
		t.Fatalf("brokenLinks = %v, want %v", got, want)
	}
	if bad := brokenLinks("index.mdx", "hero:\n  actions:\n    - link: getting-started/\n<LinkCard href=\"/codeknit/how-to/\" />"); !slices.Equal(bad, []string{"/codeknit/how-to/"}) {
		t.Fatalf("brokenLinks(index) = %v", bad)
	}
	if got := localizeLinks("[a](/codeknit/x/) [b](../y/) href=\"/codeknit/\"\n  link: /codeknit/z/", "de"); got != "[a](/codeknit/de/x/) [b](../y/) href=\"/codeknit/de/\"\n  link: /codeknit/de/z/" {
		t.Fatalf("localizeLinks = %q", got)
	}
	if !strings.Contains(sitemapFor("getting-started.md"), "- guides/parse-command.md: ../guides/parse-command/") {
		t.Fatalf("sitemap missing relative link:\n%s", sitemapFor("getting-started.md"))
	}
}

func TestWithFrontmatter(t *testing.T) {
	fm, ok := promptFrontmatter("DOCUMENT TYPE: Reference\n\nFrontmatter:\n---\ntitle: T\n---\n\nCover:\n---\n")
	if !ok || fm != "---\ntitle: T\n---" {
		t.Fatalf("promptFrontmatter = %q, %t", fm, ok)
	}
	want := "---\ntitle: T\n---\n\nBody\n"
	for _, in := range []string{"Body\n", "\n\nBody\n", "---\ntitle: Wrong\n---\n\nBody\n"} {
		if got := withFrontmatter(in, fm); got != want {
			t.Errorf("withFrontmatter(%q) = %q, want %q", in, got, want)
		}
	}
}
