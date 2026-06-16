// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"bytes"
	"testing"

	"codeknit/internal/ir"
	"codeknit/internal/plugin"
)

// --- helpers ---

// makeGraph builds a minimal SymbolGraph from the given symbols and edges.
// It assigns IDs of the form "file.go::Name" and calls BuildIndexes.
func makeGraph(syms []plugin.Symbol, edges []plugin.Edge) *ir.SymbolGraph {
	sg := &ir.SymbolGraph{
		Symbols: syms,
		Edges:   edges,
	}
	sg.BuildIndexes()
	return sg
}

// sym is a shorthand for building a plugin.Symbol.
func sym(id, name string, cat plugin.SymbolCategory, kind string, props map[string]string, sig string) plugin.Symbol {
	return plugin.Symbol{
		ID:         id,
		Name:       name,
		FilePath:   "file.go",
		Category:   cat,
		Kind:       kind,
		Properties: props,
		Signature:  sig,
	}
}

func edgeContains(from, to string) plugin.Edge {
	return plugin.Edge{From: from, To: to, Kind: plugin.EdgeContains}
}

func edgeInherits(from, to string) plugin.Edge {
	return plugin.Edge{From: from, To: to, Kind: plugin.EdgeInherits}
}

func edgeImplements(from, to string) plugin.Edge {
	return plugin.Edge{From: from, To: to, Kind: plugin.EdgeImplements}
}

// --- Types() happy path ---

func TestTypes_EmptyGraphDoesNothing(t *testing.T) {
	sg := makeGraph(nil, nil)
	Types(sg) // must not panic
}

func TestTypes_NonTypeSymbolsAreSkipped(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::fn", "fn", plugin.CategoryCallable, "function", nil, "fn()"),
		sym("f.go::v", "v", plugin.CategoryValue, "variable", nil, ""),
	}, nil)
	Types(sg)
	for _, s := range sg.Symbols {
		if len(s.BodyTokens) > 0 {
			t.Errorf("non-type symbol %q should not get BodyTokens", s.Name)
		}
	}
}

func TestTypes_TypeWithNoChildrenGetsNoTokens(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::User", "User", plugin.CategoryType, "class", nil, ""),
	}, nil)
	Types(sg)
	if len(sg.Symbols[0].BodyTokens) > 0 {
		t.Error("type with no children should not get BodyTokens")
	}
}

func TestTypes_TypeWithPreexistingTokensIsNotOverwritten(t *testing.T) {
	existing := []byte{0x01, 0x02, 0x03}
	sg := makeGraph([]plugin.Symbol{
		{
			ID: "f.go::User", Name: "User", FilePath: "f.go",
			Category: plugin.CategoryType, Kind: "class",
			BodyTokens: existing,
		},
		sym("f.go::User.save", "save", plugin.CategoryCallable, "method", nil, "save()"),
	}, []plugin.Edge{edgeContains("f.go::User", "f.go::User.save")})
	Types(sg)
	if !bytes.Equal(sg.Symbols[0].BodyTokens, existing) {
		t.Error("pre-existing BodyTokens should not be overwritten")
	}
}

func TestTypes_FieldsAreEncoded(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::User", "User", plugin.CategoryType, "class", nil, ""),
		sym("f.go::User.name", "name", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::User.email", "email", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::User", "f.go::User.name"),
		edgeContains("f.go::User", "f.go::User.email"),
	})
	Types(sg)
	tokens := sg.Symbols[0].BodyTokens
	if len(tokens) == 0 {
		t.Fatal("type with fields should get BodyTokens")
	}
	// Count fpField bytes.
	fieldCount := 0
	for _, b := range tokens {
		if b == fpField {
			fieldCount++
		}
	}
	if fieldCount != 2 {
		t.Errorf("expected 2 field tokens, got %d in %x", fieldCount, tokens)
	}
}

func TestTypes_MethodsAreEncoded(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::User", "User", plugin.CategoryType, "class", nil, ""),
		sym("f.go::User.save", "save", plugin.CategoryCallable, "method", nil, "save()"),
		sym("f.go::User.load", "load", plugin.CategoryCallable, "method", nil, "load() -> bool"),
	}, []plugin.Edge{
		edgeContains("f.go::User", "f.go::User.save"),
		edgeContains("f.go::User", "f.go::User.load"),
	})
	Types(sg)
	tokens := sg.Symbols[0].BodyTokens
	if len(tokens) == 0 {
		t.Fatal("type with methods should get BodyTokens")
	}
	methodCount := 0
	for _, b := range tokens {
		if b == fpMethod {
			methodCount++
		}
	}
	if methodCount != 2 {
		t.Errorf("expected 2 method tokens, got %d in %x", methodCount, tokens)
	}
}

func TestTypes_InheritanceIsEncoded(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::Base", "Base", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Child", "Child", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Child.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{
		edgeInherits("f.go::Child", "f.go::Base"),
		edgeContains("f.go::Child", "f.go::Child.m"),
	})
	Types(sg)
	// Find Child's tokens.
	var childTokens []byte
	for _, s := range sg.Symbols {
		if s.Name == "Child" {
			childTokens = s.BodyTokens
		}
	}
	if len(childTokens) == 0 {
		t.Fatal("Child should have BodyTokens")
	}
	found := false
	for _, b := range childTokens {
		if b == fpInherits {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fpInherits in Child tokens, got %x", childTokens)
	}
}

func TestTypes_ImplementsIsEncoded(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::Iface", "Iface", plugin.CategoryType, "interface", nil, ""),
		sym("f.go::Impl", "Impl", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Impl.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{
		edgeImplements("f.go::Impl", "f.go::Iface"),
		edgeContains("f.go::Impl", "f.go::Impl.m"),
	})
	Types(sg)
	var implTokens []byte
	for _, s := range sg.Symbols {
		if s.Name == "Impl" {
			implTokens = s.BodyTokens
		}
	}
	found := false
	for _, b := range implTokens {
		if b == fpImplements {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fpImplements in Impl tokens, got %x", implTokens)
	}
}

func TestTypes_NestedTypeIsEncoded(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::Outer", "Outer", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Outer.Inner", "Inner", plugin.CategoryType, "class", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::Outer", "f.go::Outer.Inner"),
	})
	Types(sg)
	var outerTokens []byte
	for _, s := range sg.Symbols {
		if s.Name == "Outer" {
			outerTokens = s.BodyTokens
		}
	}
	found := false
	for _, b := range outerTokens {
		if b == fpNested {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fpNested in Outer tokens, got %x", outerTokens)
	}
}

// --- Order independence ---

func TestTypes_MethodOrderDoesNotAffectFingerprint(t *testing.T) {
	// Two types with the same methods in different declaration order.
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.save", "save", plugin.CategoryCallable, "method", nil, "save()"),
		sym("f.go::A.load", "load", plugin.CategoryCallable, "method", nil, "load()"),
	}, []plugin.Edge{
		edgeContains("f.go::A", "f.go::A.save"),
		edgeContains("f.go::A", "f.go::A.load"),
	})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.load", "load", plugin.CategoryCallable, "method", nil, "load()"),
		sym("f.go::B.save", "save", plugin.CategoryCallable, "method", nil, "save()"),
	}, []plugin.Edge{
		edgeContains("f.go::B", "f.go::B.load"),
		edgeContains("f.go::B", "f.go::B.save"),
	})
	Types(sg1)
	Types(sg2)
	if !bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("method order should not affect fingerprint:\n  A: %x\n  B: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

func TestTypes_FieldOrderDoesNotAffectFingerprint(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "struct", nil, ""),
		sym("f.go::A.x", "x", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::A.y", "y", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::A", "f.go::A.x"),
		edgeContains("f.go::A", "f.go::A.y"),
	})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "struct", nil, ""),
		sym("f.go::B.y", "y", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::B.x", "x", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::B", "f.go::B.y"),
		edgeContains("f.go::B", "f.go::B.x"),
	})
	Types(sg1)
	Types(sg2)
	if !bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("field order should not affect fingerprint:\n  A: %x\n  B: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

// --- Structural distinction ---

func TestTypes_DifferentMethodCountProducesDifferentFingerprint(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m1", "m1", plugin.CategoryCallable, "method", nil, "m1()"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.m1")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m1", "m1", plugin.CategoryCallable, "method", nil, "m1()"),
		sym("f.go::B.m2", "m2", plugin.CategoryCallable, "method", nil, "m2()"),
	}, []plugin.Edge{
		edgeContains("f.go::B", "f.go::B.m1"),
		edgeContains("f.go::B", "f.go::B.m2"),
	})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("different method counts should produce different fingerprints")
	}
}

func TestTypes_DifferentParamCountProducesDifferentFingerprint(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m", "m", plugin.CategoryCallable, "method", nil, "m(x: int)"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.m")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m", "m", plugin.CategoryCallable, "method", nil, "m(x: int, y: int)"),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.m")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("different param counts should produce different fingerprints")
	}
}

func TestTypes_ReturnTypePresenceDistinguishesFingerprint(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.m")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m", "m", plugin.CategoryCallable, "method", nil, "m() -> bool"),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.m")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("presence of return type should produce different fingerprints")
	}
}

func TestTypes_InheritanceCountDistinguishesFingerprint(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::Base", "Base", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{
		edgeInherits("f.go::A", "f.go::Base"),
		edgeContains("f.go::A", "f.go::A.m"),
	})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{
		edgeContains("f.go::B", "f.go::B.m"),
	})
	Types(sg1)
	Types(sg2)
	var aTokens, bTokens []byte
	for _, s := range sg1.Symbols {
		if s.Name == "A" {
			aTokens = s.BodyTokens
		}
	}
	for _, s := range sg2.Symbols {
		if s.Name == "B" {
			bTokens = s.BodyTokens
		}
	}
	if bytes.Equal(aTokens, bTokens) {
		t.Error("presence of inheritance should produce different fingerprints")
	}
}

// --- Modifier encoding ---

func TestTypes_StaticFieldIsDistinctFromRegularField(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.x", "x", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.x")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.x", "x", plugin.CategoryValue, "field", map[string]string{"static": "true"}, ""),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.x")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("static field should produce different fingerprint from regular field")
	}
}

func TestTypes_AbstractMethodIsDistinctFromRegularMethod(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.m")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m", "m", plugin.CategoryCallable, "method", map[string]string{"abstract": "true"}, "m()"),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.m")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("abstract method should produce different fingerprint from regular method")
	}
}

func TestTypes_AsyncMethodIsDistinctFromRegularMethod(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.m", "m", plugin.CategoryCallable, "method", nil, "m()"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.m")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.m", "m", plugin.CategoryCallable, "method", map[string]string{"async": "true"}, "m()"),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.m")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Error("async method should produce different fingerprint from regular method")
	}
}

// --- Identical types produce identical fingerprints ---

func TestTypes_IdenticalTypesProduceIdenticalFingerprints(t *testing.T) {
	build := func(prefix string) *ir.SymbolGraph {
		return makeGraph([]plugin.Symbol{
			sym(prefix+"::T", "T", plugin.CategoryType, "class", nil, ""),
			sym(prefix+"::T.name", "name", plugin.CategoryValue, "field", nil, ""),
			sym(prefix+"::T.save", "save", plugin.CategoryCallable, "method", nil, "save()"),
			sym(prefix+"::T.load", "load", plugin.CategoryCallable, "method", nil, "load() -> bool"),
		}, []plugin.Edge{
			edgeContains(prefix+"::T", prefix+"::T.name"),
			edgeContains(prefix+"::T", prefix+"::T.save"),
			edgeContains(prefix+"::T", prefix+"::T.load"),
		})
	}
	sg1 := build("a.go")
	sg2 := build("b.go")
	Types(sg1)
	Types(sg2)
	if !bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("identical types should produce identical fingerprints:\n  a: %x\n  b: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

// --- Non-contains edges are ignored ---

func TestTypes_NonContainsEdgesAreIgnored(t *testing.T) {
	// A type with only calls/references/imports edges (no contains) should
	// get no BodyTokens because it has no children to encode.
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
	}, []plugin.Edge{
		{From: "f.go::A", To: "f.go::B", Kind: plugin.EdgeCalls},
		{From: "f.go::A", To: "f.go::B", Kind: plugin.EdgeReferences},
		{From: "f.go::A", To: "f.go::B", Kind: plugin.EdgeImports},
	})
	Types(sg)
	// Neither A nor B has children, so neither should get BodyTokens.
	for _, s := range sg.Symbols {
		if len(s.BodyTokens) > 0 {
			t.Errorf("type %q with no children should not get BodyTokens (got %x)", s.Name, s.BodyTokens)
		}
	}
}

// --- countParams ---

func TestCountParams_KnownSignatures(t *testing.T) {
	cases := []struct {
		sig  string
		want int
	}{
		{"foo()", 0},
		{"foo(x: int)", 1},
		{"foo(x: int, y: string)", 2},
		{"foo(x: int, y: string, z: bool)", 3},
		{"foo(  )", 0},
		{"noparens", 0},
		{"foo(Map<K, V>)", 1},                    // nested generic — comma inside <> not counted
		{"foo(Map<K, V>, x: int)", 2},            // generic + plain param
		{"foo(fn: (a: int) -> bool)", 1},         // nested parens — comma inside () not counted
		{"foo(fn: (a: int, b: int) -> bool)", 1}, // nested parens with comma
		{"foo(a: int, fn: (x: int) -> bool)", 2}, // mixed
	}
	for _, tc := range cases {
		t.Run(tc.sig, func(t *testing.T) {
			if got := countParams(tc.sig); got != tc.want {
				t.Errorf("countParams(%q) = %d, want %d", tc.sig, got, tc.want)
			}
		})
	}
}

// --- hasReturnType ---

func TestHasReturnType(t *testing.T) {
	cases := []struct {
		sig  string
		want bool
	}{
		{"foo()", false},
		{"foo() -> bool", true},
		{"foo(x: int) -> string", true},
		{"foo(x: int)", false},
		{"", false},
		{"-> bool", true},
		// Note: hasReturnType uses a simple string search for "->".
		// A "->" inside a generic type parameter is a known false positive.
		// This is acceptable because such signatures are rare in practice.
		{"foo(x: Map<K -> V>)", true}, // known limitation: -> inside generic
	}
	for _, tc := range cases {
		t.Run(tc.sig, func(t *testing.T) {
			if got := hasReturnType(tc.sig); got != tc.want {
				t.Errorf("hasReturnType(%q) = %v, want %v", tc.sig, got, tc.want)
			}
		})
	}
}

// --- encodeFieldShape ---

func TestEncodeFieldShape_RegularField(t *testing.T) {
	s := sym("f.go::x", "x", plugin.CategoryValue, "field", nil, "")
	tokens := encodeFieldShape(&s)
	if len(tokens) == 0 || tokens[0] != fpField {
		t.Errorf("expected fpField as first byte, got %x", tokens)
	}
	// Now includes fpField + fpNameMark + 2 hash bytes = 4 bytes minimum.
	if len(tokens) < 4 {
		t.Errorf("field should produce at least 4 tokens (fpField+fpNameMark+h0+h1), got %d: %x", len(tokens), tokens)
	}
}

func TestEncodeFieldShape_StaticField(t *testing.T) {
	s := sym("f.go::x", "x", plugin.CategoryValue, "field", map[string]string{"static": "true"}, "")
	tokens := encodeFieldShape(&s)
	if len(tokens) < 2 {
		t.Fatalf("static field should produce at least 2 tokens, got %d: %x", len(tokens), tokens)
	}
	if tokens[0] != fpField {
		t.Errorf("first token should be fpField, got %x", tokens[0])
	}
	// fpNameMark is now at index 1, fpStatic comes after the name hash bytes.
	hasFpStatic := false
	for _, b := range tokens {
		if b == fpStatic {
			hasFpStatic = true
			break
		}
	}
	if !hasFpStatic {
		t.Errorf("static field tokens should contain fpStatic, got %x", tokens)
	}
}

// --- encodeMethodShape ---

func TestEncodeMethodShape_SimpleMethod(t *testing.T) {
	s := sym("f.go::m", "m", plugin.CategoryCallable, "method", nil, "m()")
	tokens := encodeMethodShape(&s)
	if tokens[0] != fpMethod {
		t.Errorf("first token should be fpMethod, got %x", tokens[0])
	}
	// Should contain fpParamMark followed by 0 (no params).
	found := false
	for i, b := range tokens {
		if b == fpParamMark && i+1 < len(tokens) && tokens[i+1] == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fpParamMark+0 in tokens, got %x", tokens)
	}
	// No return type → no fpRetMark.
	for _, b := range tokens {
		if b == fpRetMark {
			t.Errorf("unexpected fpRetMark for method without return type: %x", tokens)
		}
	}
}

func TestEncodeMethodShape_MethodWithParamsAndReturn(t *testing.T) {
	s := sym("f.go::m", "m", plugin.CategoryCallable, "method", nil, "m(a: int, b: string) -> bool")
	tokens := encodeMethodShape(&s)
	// Should have fpParamMark + 2.
	foundParams := false
	for i, b := range tokens {
		if b == fpParamMark && i+1 < len(tokens) && tokens[i+1] == 2 {
			foundParams = true
			break
		}
	}
	if !foundParams {
		t.Errorf("expected fpParamMark+2 in tokens, got %x", tokens)
	}
	// Should have fpRetMark.
	foundRet := false
	for _, b := range tokens {
		if b == fpRetMark {
			foundRet = true
			break
		}
	}
	if !foundRet {
		t.Errorf("expected fpRetMark in tokens, got %x", tokens)
	}
}

func TestEncodeMethodShape_AllModifiers(t *testing.T) {
	s := sym("f.go::m", "m", plugin.CategoryCallable, "method",
		map[string]string{"abstract": "true", "static": "true", "async": "true"}, "m()")
	tokens := encodeMethodShape(&s)
	has := func(b byte) bool {
		for _, t := range tokens {
			if t == b {
				return true
			}
		}
		return false
	}
	if !has(fpAbstract) {
		t.Errorf("expected fpAbstract in tokens: %x", tokens)
	}
	if !has(fpStatic) {
		t.Errorf("expected fpStatic in tokens: %x", tokens)
	}
	if !has(fpAsync) {
		t.Errorf("expected fpAsync in tokens: %x", tokens)
	}
}

// ---------------------------------------------------------------------------
// kindHash
// ---------------------------------------------------------------------------

func TestKindHash_Deterministic(t *testing.T) {
	a := kindHash("struct")
	b := kindHash("struct")
	if a != b {
		t.Errorf("kindHash should be deterministic: %x != %x", a, b)
	}
}

func TestKindHash_DistinctKinds(t *testing.T) {
	kinds := []string{"struct", "class", "interface", "type_alias", "enum"}
	seen := make(map[[2]byte]string)
	for _, k := range kinds {
		h := kindHash(k)
		if prev, ok := seen[h]; ok {
			t.Errorf("kindHash collision: %q and %q both produce %x", prev, k, h)
		}
		seen[h] = k
	}
}

// ---------------------------------------------------------------------------
// encodeFieldShape — field name included
// ---------------------------------------------------------------------------

func TestEncodeFieldShape_IncludesNameHash(t *testing.T) {
	a := sym("f.go::a", "inputPath", plugin.CategoryValue, "field", nil, "")
	b := sym("f.go::b", "output", plugin.CategoryValue, "field", nil, "")
	ta := encodeFieldShape(&a)
	tb := encodeFieldShape(&b)
	if bytes.Equal(ta, tb) {
		t.Errorf("fields with different names should produce different tokens: %x == %x", ta, tb)
	}
}

func TestEncodeFieldShape_SameNameProducesSameTokens(t *testing.T) {
	a := sym("f.go::a", "workers", plugin.CategoryValue, "field", nil, "")
	b := sym("g.go::b", "workers", plugin.CategoryValue, "field", nil, "")
	ta := encodeFieldShape(&a)
	tb := encodeFieldShape(&b)
	if !bytes.Equal(ta, tb) {
		t.Errorf("fields with same name should produce same tokens: %x != %x", ta, tb)
	}
}

func TestEncodeFieldShape_ContainsFpNameMark(t *testing.T) {
	s := sym("f.go::x", "x", plugin.CategoryValue, "field", nil, "")
	tokens := encodeFieldShape(&s)
	found := false
	for _, b := range tokens {
		if b == fpNameMark {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("encodeFieldShape should contain fpNameMark, got %x", tokens)
	}
}

// ---------------------------------------------------------------------------
// encodeMethodShape — method name included
// ---------------------------------------------------------------------------

func TestEncodeMethodShape_DifferentNamesProduceDifferentTokens(t *testing.T) {
	validate := sym("f.go::Validate", "Validate", plugin.CategoryCallable, "method", nil, "Validate()")
	isValid := sym("f.go::IsValid", "IsValid", plugin.CategoryCallable, "method", nil, "IsValid()")
	tv := encodeMethodShape(&validate)
	ti := encodeMethodShape(&isValid)
	if bytes.Equal(tv, ti) {
		t.Errorf("methods with different names should produce different tokens: %x == %x", tv, ti)
	}
}

func TestEncodeMethodShape_SameNameProducesSameTokens(t *testing.T) {
	a := sym("f.go::Save", "Save", plugin.CategoryCallable, "method", nil, "Save()")
	b := sym("g.go::Save", "Save", plugin.CategoryCallable, "method", nil, "Save()")
	ta := encodeMethodShape(&a)
	tb := encodeMethodShape(&b)
	if !bytes.Equal(ta, tb) {
		t.Errorf("methods with same name/sig should produce same tokens: %x != %x", ta, tb)
	}
}

func TestEncodeMethodShape_ContainsFpNameMark(t *testing.T) {
	s := sym("f.go::m", "m", plugin.CategoryCallable, "method", nil, "m()")
	tokens := encodeMethodShape(&s)
	found := false
	for _, b := range tokens {
		if b == fpNameMark {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("encodeMethodShape should contain fpNameMark, got %x", tokens)
	}
}

// ---------------------------------------------------------------------------
// Types() — field and method name discrimination
// ---------------------------------------------------------------------------

func TestTypes_DifferentFieldNamesProduceDifferentFingerprints(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "struct", nil, ""),
		sym("f.go::A.inputPath", "inputPath", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::A.workers", "workers", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::A", "f.go::A.inputPath"),
		edgeContains("f.go::A", "f.go::A.workers"),
	})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "struct", nil, ""),
		sym("f.go::B.output", "output", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::B.minSim", "minSim", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{
		edgeContains("f.go::B", "f.go::B.output"),
		edgeContains("f.go::B", "f.go::B.minSim"),
	})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("structs with different field names should produce different fingerprints:\n  A: %x\n  B: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

func TestTypes_SameFieldNamesProduceSameFingerprints(t *testing.T) {
	build := func(prefix string) *ir.SymbolGraph {
		return makeGraph([]plugin.Symbol{
			sym(prefix+"::S", "S", plugin.CategoryType, "struct", nil, ""),
			sym(prefix+"::S.name", "name", plugin.CategoryValue, "field", nil, ""),
			sym(prefix+"::S.age", "age", plugin.CategoryValue, "field", nil, ""),
		}, []plugin.Edge{
			edgeContains(prefix+"::S", prefix+"::S.name"),
			edgeContains(prefix+"::S", prefix+"::S.age"),
		})
	}
	sg1 := build("a.go")
	sg2 := build("b.go")
	Types(sg1)
	Types(sg2)
	if !bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("structs with same field names should produce same fingerprints:\n  a: %x\n  b: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

func TestTypes_DifferentMethodNamesProduceDifferentFingerprints(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.Validate", "Validate", plugin.CategoryCallable, "method", nil, "Validate()"),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.Validate")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.IsValid", "IsValid", plugin.CategoryCallable, "method", nil, "IsValid()"),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.IsValid")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("types with different method names should produce different fingerprints:\n  A: %x\n  B: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}

// ---------------------------------------------------------------------------
// Types() — parent structural fingerprint
// ---------------------------------------------------------------------------

func TestTypes_DifferentParentShapesProduceDifferentFingerprints(t *testing.T) {
	sg := makeGraph([]plugin.Symbol{
		sym("f.go::Base1", "Base1", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Base1.x", "x", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::Base2", "Base2", plugin.CategoryType, "class", nil, ""),
		sym("f.go::Base2.x", "x", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::Base2.y", "y", plugin.CategoryValue, "field", nil, ""),
		sym("f.go::A", "A", plugin.CategoryType, "class", nil, ""),
		sym("f.go::A.run", "run", plugin.CategoryCallable, "method", nil, "run()"),
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.run", "run", plugin.CategoryCallable, "method", nil, "run()"),
	}, []plugin.Edge{
		edgeContains("f.go::Base1", "f.go::Base1.x"),
		edgeContains("f.go::Base2", "f.go::Base2.x"),
		edgeContains("f.go::Base2", "f.go::Base2.y"),
		edgeInherits("f.go::A", "f.go::Base1"),
		edgeContains("f.go::A", "f.go::A.run"),
		edgeInherits("f.go::B", "f.go::Base2"),
		edgeContains("f.go::B", "f.go::B.run"),
	})
	Types(sg)
	var aTokens, bTokens []byte
	for _, s := range sg.Symbols {
		switch s.Name {
		case "A":
			aTokens = s.BodyTokens
		case "B":
			bTokens = s.BodyTokens
		}
	}
	if bytes.Equal(aTokens, bTokens) {
		t.Errorf("classes with different parent shapes should produce different fingerprints:\n  A: %x\n  B: %x",
			aTokens, bTokens)
	}
}

func TestTypes_SameParentShapeProducesSameFingerprints(t *testing.T) {
	build := func(prefix, childName, parentName string) *ir.SymbolGraph {
		return makeGraph([]plugin.Symbol{
			sym(prefix+"::"+parentName, parentName, plugin.CategoryType, "class", nil, ""),
			sym(prefix+"::"+parentName+".x", "x", plugin.CategoryValue, "field", nil, ""),
			sym(prefix+"::"+childName, childName, plugin.CategoryType, "class", nil, ""),
			sym(prefix+"::"+childName+".run", "run", plugin.CategoryCallable, "method", nil, "run()"),
		}, []plugin.Edge{
			edgeContains(prefix+"::"+parentName, prefix+"::"+parentName+".x"),
			edgeInherits(prefix+"::"+childName, prefix+"::"+parentName),
			edgeContains(prefix+"::"+childName, prefix+"::"+childName+".run"),
		})
	}
	sg1 := build("a.go", "A", "Foo")
	sg2 := build("b.go", "B", "Bar")
	Types(sg1)
	Types(sg2)
	var aTokens, bTokens []byte
	for _, s := range sg1.Symbols {
		if s.Name == "A" {
			aTokens = s.BodyTokens
		}
	}
	for _, s := range sg2.Symbols {
		if s.Name == "B" {
			bTokens = s.BodyTokens
		}
	}
	if !bytes.Equal(aTokens, bTokens) {
		t.Errorf("classes with same parent shape should produce same fingerprints:\n  A: %x\n  B: %x",
			aTokens, bTokens)
	}
}

func TestTypes_KindDiscriminatesStructFromClass(t *testing.T) {
	sg1 := makeGraph([]plugin.Symbol{
		sym("f.go::A", "A", plugin.CategoryType, "struct", nil, ""),
		sym("f.go::A.x", "x", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{edgeContains("f.go::A", "f.go::A.x")})
	sg2 := makeGraph([]plugin.Symbol{
		sym("f.go::B", "B", plugin.CategoryType, "class", nil, ""),
		sym("f.go::B.x", "x", plugin.CategoryValue, "field", nil, ""),
	}, []plugin.Edge{edgeContains("f.go::B", "f.go::B.x")})
	Types(sg1)
	Types(sg2)
	if bytes.Equal(sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens) {
		t.Errorf("struct and class with same fields should produce different fingerprints:\n  struct: %x\n  class: %x",
			sg1.Symbols[0].BodyTokens, sg2.Symbols[0].BodyTokens)
	}
}
