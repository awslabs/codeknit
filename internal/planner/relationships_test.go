// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"slices"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/golang"
	"codeknit/internal/plugin/typescript"
)

func TestRelationshipsImportIsAuthoritative(t *testing.T) {
	for _, module := range []string{"external-parser", "./parser", "parserOld"} {
		t.Run(module, func(t *testing.T) {
			sg := (&Planner{}).Plan(map[string][]plugin.Symbol{
				"app.ts":       {{Name: "run", Category: plugin.CategoryCallable}},
				"parserOld.ts": {{Name: "parse", Category: plugin.CategoryCallable}},
			}, map[string][]plugin.Edge{
				"app.ts": {
					{From: "parse", To: module, Kind: plugin.EdgeImports},
					{From: "run", To: "parse", Kind: plugin.EdgeCalls},
				},
			})
			if got := sg.Edges[0].To; got != "app.ts::parse" {
				t.Fatalf("import from %q resolved to unrelated declaration %q", module, got)
			}
		})
	}
}

func TestRelationshipsAmbiguityIsExplicit(t *testing.T) {
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{
		"app.go": {{Name: "Run", Category: plugin.CategoryCallable}},
		"a.go":   {{Name: "Helper", Category: plugin.CategoryCallable}},
		"b.go":   {{Name: "Helper", Category: plugin.CategoryCallable}},
		"app.py": {{Name: "Helper", Category: plugin.CategoryCallable}},
	}, map[string][]plugin.Edge{
		"app.go": {{From: "Run", To: "Helper", Kind: plugin.EdgeCalls}},
	})
	e := sg.Edges[0]
	if e.Resolution != "ambiguous" || !slices.Equal(e.Candidates, []string{"a.go::Helper", "b.go::Helper"}) {
		t.Fatalf("expected same-language ambiguity, got %+v", e)
	}
}

func TestRelationshipsOverloadsKeepTheirOwner(t *testing.T) {
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{
		"A.java": {
			{Name: "app", Category: plugin.CategoryModule},
			{Name: "A", Category: plugin.CategoryType},
			{Name: "run", ScopedName: "A.run", Category: plugin.CategoryCallable},
			{Name: "run", ScopedName: "A.run", Category: plugin.CategoryCallable},
			{Name: "caller", Category: plugin.CategoryCallable},
		},
	}, map[string][]plugin.Edge{
		"A.java": {
			{From: "A", To: "A.run", Kind: plugin.EdgeContains},
			{From: "caller", To: "A.run", Kind: plugin.EdgeCalls},
		},
	})
	var children []string
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeContains && (e.To == "A.java::A.run" || e.To == "A.java::A.run#1") {
			if e.From != "A.java::A" || e.Resolution != "" {
				t.Errorf("overload has the wrong owner: %+v", e)
			}
			children = append(children, e.To)
		}
		if e.Kind == plugin.EdgeCalls && e.Resolution != "ambiguous" {
			t.Errorf("overload call was guessed: %+v", e)
		}
	}
	slices.Sort(children)
	if !slices.Equal(children, []string{"A.java::A.run", "A.java::A.run#1"}) {
		t.Fatalf("missing overload containment: %v", children)
	}
}

func TestRelationshipsRenamedAndNamespaceImports(t *testing.T) {
	p, dir := typescript.NewPlugin(), t.TempDir()
	targets, targetEdges := parseFile(t, p, dir, "parser.ts", "export function parse() {}")
	caller, calls := parseFile(t, p, dir, "app.ts", `
import { parse as read } from './parser';
import * as parser from './parser';
function one() { read(); }
function two() { parser.parse(); }
`)
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"parser.ts": targets, "app.ts": caller},
		map[string][]plugin.Edge{"parser.ts": targetEdges, "app.ts": calls})
	var callers []string
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls && e.To == "parser.ts::parse" && e.Resolution == "" {
			callers = append(callers, e.From)
		}
	}
	slices.Sort(callers)
	if !slices.Equal(callers, []string{"app.ts::one", "app.ts::two"}) {
		t.Fatalf("lost imported bindings: %+v", sg.Edges)
	}
}

func TestRelationshipsAliasesStayInTheirScope(t *testing.T) {
	symbols, edges := parseFile(t, typescript.NewPlugin(), t.TempDir(), "app.ts", `
function first() {}
function second() {}
function a() {
    const handler = first;
    const routes = { run: first };
    handler();
    routes.run();
}
function b() {
    const handler = second;
    const routes = { run: second };
    handler();
    routes.run();
}
function c() { handler(); unknown.run(); }
function factory() { return first; }
function create() { factory(); }
`)
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"app.ts": symbols},
		map[string][]plugin.Edge{"app.ts": edges})
	want := map[string]string{"app.ts::a": "app.ts::first", "app.ts::b": "app.ts::second", "app.ts::create": "app.ts::factory"}
	found := make(map[string]bool)
	for _, e := range sg.Edges {
		if e.Kind != plugin.EdgeCalls || e.Resolution != "" {
			continue
		}
		if e.To != want[e.From] {
			t.Errorf("invented call or alias crossed scope: %+v", e)
		}
		found[e.From] = true
	}
	for caller := range want {
		if !found[caller] {
			t.Errorf("missing resolved call from %s: %+v", caller, sg.Edges)
		}
	}
}

func TestRelationshipsQualifiedImport(t *testing.T) {
	dir := t.TempDir()
	p := golang.NewPlugin()
	caller, calls := parseFile(t, p, dir, "main.go", `package main
import service "./service"
func Run() { service.Work() }
`)
	target, _ := parseFile(t, p, dir, "work.go", "package service\nfunc Work() {}")
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{
		"main.go": caller, "service/work.go": target,
	}, map[string][]plugin.Edge{"main.go": calls})
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls && e.From == "main.go::Run" && e.To == "service/work.go::Work" {
			return
		}
	}
	t.Fatalf("qualified cross-file call missing: %+v", sg.Edges)
}

func TestRelationshipsSameLineOwners(t *testing.T) {
	symbols, edges := parseFile(t, typescript.NewPlugin(), t.TempDir(), "app.ts",
		"function a() { first(); } function b() { second(); }")
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"app.ts": symbols},
		map[string][]plugin.Edge{"app.ts": edges})
	want := map[[2]string]bool{
		{"app.ts::a", "app.ts::first"}:  true,
		{"app.ts::b", "app.ts::second"}: true,
	}
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			if !want[[2]string{e.From, e.To}] {
				t.Errorf("call attributed to wrong owner: %+v", e)
			}
			delete(want, [2]string{e.From, e.To})
		}
	}
	if len(want) != 0 {
		t.Errorf("missing calls: %v", want)
	}
}

func TestRelationshipsSameLineClassScopes(t *testing.T) {
	symbols, edges := parseFile(t, typescript.NewPlugin(), t.TempDir(), "app.ts",
		"class A { helper() {} run() { this.helper(); } } class B { helper() {} run() { this.helper(); } }")
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"app.ts": symbols},
		map[string][]plugin.Edge{"app.ts": edges})
	want := map[string]string{"app.ts::A.run": "app.ts::A.helper", "app.ts::B.run": "app.ts::B.helper"}
	for _, e := range sg.Edges {
		if e.Kind == plugin.EdgeCalls {
			if e.To != want[e.From] || e.Resolution != "" {
				t.Errorf("wrong class scope: %+v", e)
			}
			delete(want, e.From)
		}
	}
	if len(want) > 0 {
		t.Fatalf("missing method calls: %+v", want)
	}
}

func TestRelationshipsParameterShadowing(t *testing.T) {
	symbols, edges := parseFile(t, typescript.NewPlugin(), t.TempDir(), "app.ts", `
function helper() {}
function run(helper: () => void) { helper(); }
const arrow = () => {};
function useArrow() { arrow(); }
`)
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"app.ts": symbols},
		map[string][]plugin.Edge{"app.ts": edges})
	foundArrow := false
	for _, e := range sg.Edges {
		if e.Kind != plugin.EdgeCalls {
			continue
		}
		if e.From == "app.ts::run" && e.Resolution == "" {
			t.Errorf("parameter resolved to unrelated function: %+v", e)
		}
		if e.From == "app.ts::useArrow" && e.To == "app.ts::arrow" && e.Resolution == "" {
			foundArrow = true
		}
	}
	if !foundArrow {
		t.Fatalf("arrow declaration was mistaken for an unknown alias: %+v", sg.Edges)
	}
}

func TestRelationshipsGroupedParametersShadowFunctions(t *testing.T) {
	symbols, edges := parseFile(t, golang.NewPlugin(), t.TempDir(), "app.go", `
package app
func second() {}
func Run(first, second func()) { second() }
`)
	sg := (&Planner{}).Plan(map[string][]plugin.Symbol{"app.go": symbols},
		map[string][]plugin.Edge{"app.go": edges})
	for _, edge := range sg.Edges {
		if edge.Kind == plugin.EdgeCalls && edge.From == "app.go::Run" {
			if edge.Resolution != "unresolved" || edge.To != "app.go::Run.second" {
				t.Fatalf("grouped parameter resolved to global function: %+v", edge)
			}
			return
		}
	}
	t.Fatal("missing call through grouped parameter")
}
