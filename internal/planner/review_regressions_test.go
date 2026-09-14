// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
	"codeknit/internal/plugin/cpp"
	"codeknit/internal/plugin/csharp"
	"codeknit/internal/plugin/java"
	"codeknit/internal/plugin/javascript"
	"codeknit/internal/plugin/python"
	"codeknit/internal/plugin/typescript"
)

func planSources(t *testing.T, p plugin.LanguagePlugin, sources map[string]string) *ir.SymbolGraph {
	t.Helper()
	dir := t.TempDir()
	symbols, edges := make(map[string][]plugin.Symbol), make(map[string][]plugin.Edge)
	for _, name := range plugin.SortedStringKeys(sources) {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(sources[name]), 0o600); err != nil {
			t.Fatal(err)
		}
		var err error
		symbols[name], edges[name], err = p.Parse(path)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
	}
	return (&Planner{}).Plan(symbols, edges)
}

func TestReviewShadowedNamespaceImports(t *testing.T) {
	sg := planSources(t, javascript.NewPlugin(), map[string]string{
		"tools.js": "export function parse() {}",
		"app.js": `import * as tools from './tools';
function shadowed(tools) { tools.parse(); }
function local() { const tools = unknown; tools.parse(); }
function known() { tools.parse(); }
function object() { const tools = { parse: target }; tools.parse(); }
function target() {}
`,
	})
	want := map[string]string{"known": "tools.js::parse", "object": "app.js::target"}
	found := make(map[string]bool)
	for _, e := range sg.Edges {
		if e.Kind != plugin.EdgeCalls {
			continue
		}
		name := strings.TrimPrefix(e.From, "app.js::")
		found[name] = true
		if target := want[name]; target != "" {
			if e.Resolution != "" || e.To != target {
				t.Errorf("valid binding lost: %+v", e)
			}
		} else if e.Resolution != "unresolved" || e.To == "tools.js::parse" {
			t.Errorf("shadowed receiver resolved through import: %+v", e)
		}
	}
	if len(found) != 4 {
		t.Fatalf("missing calls: %+v", found)
	}
}

func TestReviewPythonPackageBoundaries(t *testing.T) {
	for _, module := range []string{".service", "pkg.service"} {
		for _, exported := range []bool{false, true} {
			t.Run(module+map[bool]string{false: "/external", true: "/local"}[exported], func(t *testing.T) {
				init := "from external_api import execute\n"
				if exported {
					init = "def execute():\n    pass\n"
				}
				sg := planSources(t, python.NewPlugin(), map[string]string{
					"pkg/__init__.py":          "",
					"pkg/app.py":               "from " + module + " import execute\ndef run():\n    execute()\n",
					"pkg/service/__init__.py":  init,
					"pkg/service/unrelated.py": "def execute():\n    pass\n",
				})
				var calls int
				for _, e := range sg.Edges {
					if e.Kind != plugin.EdgeCalls {
						continue
					}
					calls++
					if exported {
						if e.Resolution != "" || e.To != "pkg/service/__init__.py::execute" {
							t.Errorf("package export not resolved: %+v", e)
						}
					} else if e.Resolution != "unresolved" {
						t.Errorf("package import resolved to a descendant: %+v", e)
					}
				}
				if calls != 1 {
					t.Fatalf("expected one call, got %d", calls)
				}
			})
		}
	}
}

func TestReviewPythonMissingExportCannotSwitchPackages(t *testing.T) {
	sg := planSources(t, python.NewPlugin(), map[string]string{
		"project1/app.py":             "from models import execute\ndef run():\n    execute()\n",
		"project1/models/__init__.py": "from external_api import execute\n",
		"project2/models/__init__.py": "def execute():\n    pass\n",
	})
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			if e.Resolution != "unresolved" || e.To != "project1/app.py::execute" {
				t.Fatalf("missing export resolved through another package: %+v", e)
			}
			return
		}
	}
	t.Fatal("missing call")
}

func TestReviewOverloadedCallers(t *testing.T) {
	for _, tc := range []struct {
		parser plugin.LanguagePlugin
		file   string
		source string
	}{
		{java.NewPlugin(), "A.java", "class A { void run() { first(); } void run(int x) { second(); } void first() {} void second() {} }"},
		{cpp.NewPlugin(), "A.cpp", "class A { public: void run() { first(); } void run(int x) { second(); } void first() {} void second() {} };"},
		{csharp.NewPlugin(), "A.cs", "class A { void run() { first(); } void run(int x) { second(); } void first() {} void second() {} }"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			for _, src := range []string{tc.source, strings.ReplaceAll(tc.source, "} ", "}\n")} {
				sg := planSources(t, tc.parser, map[string]string{tc.file: src})
				want := map[string]string{
					tc.file + "::A.run":   tc.file + "::A.first",
					tc.file + "::A.run#1": tc.file + "::A.second",
				}
				for _, e := range sg.Edges {
					if e.Kind != plugin.EdgeCalls {
						continue
					}
					if want[e.From] != e.To || e.Resolution != "" {
						t.Errorf("overloaded body lost its declaration: %+v", e)
					}
					delete(want, e.From)
				}
				if len(want) != 0 {
					t.Errorf("missing calls from overloads: %v", want)
				}
			}
		})
	}
}

func TestReviewOverrides(t *testing.T) {
	for _, tc := range []struct {
		name, source, target, resolution string
		candidates                       []string
	}{
		{"direct", "class Base { run() {} } class Child extends Base { override run() {} }", "Base.run", "", nil},
		{"inherited", "class Base { run() {} } class Middle extends Base {} class Child extends Middle { override run() {} }", "Base.run", "", nil},
		{"nearest", "class Base { run() {} } class Middle extends Base { run() {} } class Child extends Middle { override run() {} }", "Middle.run", "", nil},
		{"missing", "class Child extends External { override run() {} }", "super.run", "unresolved", nil},
		{"cycle", "class Base extends Child {} class Child extends Base { override run() {} }", "super.run", "unresolved", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sg := planSources(t, typescript.NewPlugin(), map[string]string{"app.ts": tc.source})
			var found bool
			for _, e := range sg.Edges {
				if e.Kind != plugin.EdgeOverrides {
					continue
				}
				found = true
				if e.From != "app.ts::Child.run" || e.To != "app.ts::"+tc.target || e.Resolution != tc.resolution ||
					!slices.Equal(e.Candidates, tc.candidates) || e.From == e.To {
					t.Errorf("incorrect override: %+v", e)
				}
			}
			if !found {
				t.Fatal("missing override relationship")
			}
		})
	}
}

func TestReviewImportedOverride(t *testing.T) {
	sg := planSources(t, typescript.NewPlugin(), map[string]string{
		"base.ts": "export class Base { run() {} }",
		"app.ts":  "import { Base as Parent } from './base'; class Child extends Parent { override run() {} }",
	})
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeOverrides {
			if e.To != "base.ts::Base.run" || e.Resolution != "" {
				t.Fatalf("imported override target lost: %+v", e)
			}
			return
		}
	}
	t.Fatal("missing override")
}

func TestReviewAmbiguousOverride(t *testing.T) {
	sg := planSources(t, java.NewPlugin(), map[string]string{
		"A.java": "class Base { void run() {} void run(int x) {} } class Child extends Base { @Override void run() {} }",
	})
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeOverrides {
			if e.Resolution != "ambiguous" || e.To != "A.java::super.run" ||
				!slices.Equal(e.Candidates, []string{"A.java::Base.run", "A.java::Base.run#1"}) {
				t.Fatalf("overload target guessed: %+v", e)
			}
			return
		}
	}
	t.Fatal("missing override")
}

func TestReviewOverloadsKeepSharedCalls(t *testing.T) {
	sg := planSources(t, java.NewPlugin(), map[string]string{
		"A.java": "class A { void run() { shared(); } void run(int x) { shared(); } void shared() {} }",
	})
	var sources []string
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			if e.To != "A.java::A.shared" || e.Resolution != "" {
				t.Errorf("incorrect shared call: %+v", e)
			}
			sources = append(sources, e.From)
		}
	}
	slices.Sort(sources)
	if !slices.Equal(sources, []string{"A.java::A.run", "A.java::A.run#1"}) {
		t.Fatalf("calls from distinct declarations were merged: %v", sources)
	}
}

func TestReviewOverloadParameterScope(t *testing.T) {
	sg := planSources(t, csharp.NewPlugin(), map[string]string{
		"A.cs": "class A { void run(Action callback) { callback(); } void run(int x) { callback(); } void callback() {} }",
	})
	want := map[string]string{"A.cs::A.run": "unresolved", "A.cs::A.run#1": ""}
	for _, e := range sg.Edges {
		if e.Kind != plugin.EdgeCalls {
			continue
		}
		if status, exists := want[e.From]; !exists || status != e.Resolution {
			t.Errorf("parameter binding crossed overload scope: %+v", e)
		}
		if e.From == "A.cs::A.run#1" && e.To != "A.cs::A.callback" {
			t.Errorf("unshadowed method did not resolve: %+v", e)
		}
		delete(want, e.From)
	}
	if len(want) != 0 {
		t.Fatalf("missing overload calls: %v", want)
	}
}
