// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"os"
	"path/filepath"
	"testing"

	"codeknit/internal/plugin"
	"codeknit/internal/plugin/typescript"
)

func parseFile(t *testing.T, p plugin.LanguagePlugin, dir, name, content string) (symbols []plugin.Symbol, edges []plugin.Edge) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	symbols, edges, _ = p.Parse(path)
	return symbols, edges
}

func TestDataflow_CrossFile_ObjectDispatch(t *testing.T) {
	dir := t.TempDir()
	plug := typescript.NewPlugin()

	// File 1: defines functions
	sym1, edge1 := parseFile(t, plug, dir, "handlers.ts", `
export function handleCreate(): void {}
export function handleUpdate(): void {}
export function handleDelete(): void {}
`)

	// File 2: builds dispatch table and calls through it
	sym2, edge2 := parseFile(t, plug, dir, "router.ts", `
import { handleCreate, handleUpdate, handleDelete } from './handlers';

const routes = {
    create: handleCreate,
    update: handleUpdate,
    delete: handleDelete,
};

export function dispatch(action: string): void {
    routes.create();
    routes.update();
}
`)

	t.Log("=== Per-file edges for router.ts ===")
	for _, e := range edge2 {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	// Build the symbol graph through the planner
	fileSymbols := map[string][]plugin.Symbol{
		"handlers.ts": sym1,
		"router.ts":   sym2,
	}
	fileEdges := map[string][]plugin.Edge{
		"handlers.ts": edge1,
		"router.ts":   edge2,
	}

	planner := &Planner{}
	sg := planner.Plan(fileSymbols, fileEdges)

	t.Log("=== Symbols ===")
	for _, s := range sg.Symbols {
		t.Logf("  [%s] %s/%s %q", s.ID, s.Category, s.Kind, s.Name)
	}
	t.Log("=== Edges ===")
	for _, e := range sg.Edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	// Verify alias edges were created: create -> handleCreate, etc.
	// These are consumed by resolveDataflow and turned into call edges.
	// After resolution, we should see dispatch calling handleCreate/handleUpdate.

	// Find the dispatch function's global ID
	var dispatchID string
	var handleCreateID, handleUpdateID string
	for _, s := range sg.Symbols {
		switch s.Name {
		case "dispatch":
			dispatchID = s.ID
		case "handleCreate":
			handleCreateID = s.ID
		case "handleUpdate":
			handleUpdateID = s.ID
		}
	}

	if dispatchID == "" {
		t.Fatal("missing dispatch function")
	}
	if handleCreateID == "" {
		t.Fatal("missing handleCreate function")
	}

	// Check that dispatch has a calls edge to handleCreate (resolved through alias)
	foundCreate := false
	foundUpdate := false
	for _, e := range sg.Edges {
		if e.From == dispatchID && e.Kind == plugin.EdgeCalls {
			t.Logf("  dispatch calls: %s", e.To)
			if e.To == handleCreateID {
				foundCreate = true
			}
			if e.To == handleUpdateID {
				foundUpdate = true
			}
		}
	}

	if !foundCreate {
		t.Error("dispatch should have a resolved call edge to handleCreate (through routes.create alias)")
	}
	if !foundUpdate {
		t.Error("dispatch should have a resolved call edge to handleUpdate (through routes.update alias)")
	}
}

func TestDataflow_CrossFile_VariableAlias(t *testing.T) {
	dir := t.TempDir()
	plug := typescript.NewPlugin()

	sym1, edge1 := parseFile(t, plug, dir, "core.ts", `
export function processData(): void {}
`)

	sym2, edge2 := parseFile(t, plug, dir, "app.ts", `
import { processData } from './core';

const handler = processData;

export function main(): void {
    handler();
}
`)

	fileSymbols := map[string][]plugin.Symbol{
		"core.ts": sym1,
		"app.ts":  sym2,
	}
	fileEdges := map[string][]plugin.Edge{
		"core.ts": edge1,
		"app.ts":  edge2,
	}

	planner := &Planner{}
	sg := planner.Plan(fileSymbols, fileEdges)

	t.Log("=== Edges ===")
	for _, e := range sg.Edges {
		t.Logf("  %s --%s--> %s", e.From, e.Kind, e.To)
	}

	var mainID, processDataID string
	for _, s := range sg.Symbols {
		switch s.Name {
		case "main":
			mainID = s.ID
		case "processData":
			processDataID = s.ID
		}
	}

	if mainID == "" || processDataID == "" {
		t.Fatalf("missing symbols: main=%q processData=%q", mainID, processDataID)
	}

	// main() calls handler, handler aliases processData
	// After dataflow resolution, main should have a call edge to processData
	found := false
	for _, e := range sg.Edges {
		if e.From == mainID && e.To == processDataID && e.Kind == plugin.EdgeCalls {
			found = true
		}
	}

	if !found {
		t.Error("main should have a resolved call edge to processData (through handler alias)")
	}
}
