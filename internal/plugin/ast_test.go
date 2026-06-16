// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import "testing"

func TestSymbolCategoryValues(t *testing.T) {
	tests := []struct {
		cat  SymbolCategory
		want string
	}{
		{CategoryCallable, "callable"},
		{CategoryType, "type"},
		{CategoryValue, "value"},
		{CategoryModule, "module"},
		{CategoryMeta, "meta"},
	}
	for _, tt := range tests {
		if string(tt.cat) != tt.want {
			t.Errorf("SymbolCategory = %q, want %q", tt.cat, tt.want)
		}
	}
}

func TestEdgeKindValues(t *testing.T) {
	tests := []struct {
		kind EdgeKind
		want string
	}{
		{EdgeCalls, "calls"},
		{EdgeInherits, "inherits"},
		{EdgeContains, "contains"},
		{EdgeReferences, "references"},
		{EdgeImplements, "implements"},
		{EdgeOverrides, "overrides"},
		{EdgeImports, "imports"},
		{EdgeDecorates, "decorates"},
	}
	for _, tt := range tests {
		if string(tt.kind) != tt.want {
			t.Errorf("EdgeKind = %q, want %q", tt.kind, tt.want)
		}
	}
}

func TestSymbolStruct(t *testing.T) {
	s := Symbol{
		ID:         "file.ts::MyFunc",
		Name:       "MyFunc",
		FilePath:   "file.ts",
		Category:   CategoryCallable,
		Kind:       "function",
		Signature:  "MyFunc(x: number) -> string",
		Properties: map[string]string{"exported": "true"},
		Span:       [2]int{10, 20},
	}
	if s.ID != "file.ts::MyFunc" {
		t.Errorf("Symbol.ID = %q, want %q", s.ID, "file.ts::MyFunc")
	}
	if s.Name != "MyFunc" {
		t.Errorf("Symbol.Name = %q, want %q", s.Name, "MyFunc")
	}
	if s.FilePath != "file.ts" {
		t.Errorf("Symbol.FilePath = %q, want %q", s.FilePath, "file.ts")
	}
	if s.Category != CategoryCallable {
		t.Errorf("Symbol.Category = %q, want %q", s.Category, CategoryCallable)
	}
	if s.Kind != "function" {
		t.Errorf("Symbol.Kind = %q, want %q", s.Kind, "function")
	}
	if s.Signature != "MyFunc(x: number) -> string" {
		t.Errorf("Symbol.Signature = %q, want %q", s.Signature, "MyFunc(x: number) -> string")
	}
	if s.Properties["exported"] != "true" {
		t.Errorf("Symbol.Properties[exported] = %q, want %q", s.Properties["exported"], "true")
	}
	if s.Span != [2]int{10, 20} {
		t.Errorf("Symbol.Span = %v, want %v", s.Span, [2]int{10, 20})
	}
}

func TestEdgeStruct(t *testing.T) {
	e := Edge{From: "A", To: "B", Kind: EdgeCalls}
	if e.From != "A" || e.To != "B" || e.Kind != EdgeCalls {
		t.Errorf("Edge = %+v, unexpected values", e)
	}
}
