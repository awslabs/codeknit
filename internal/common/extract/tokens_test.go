// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"bytes"
	"testing"

	"codeknit/internal/common/types"
)

// ---------------------------------------------------------------------------
// fnv32
// ---------------------------------------------------------------------------

func TestFnv32_Deterministic(t *testing.T) {
	a := fnv32([]byte("hello"))
	b := fnv32([]byte("hello"))
	if a != b {
		t.Errorf("fnv32 should be deterministic: %x != %x", a, b)
	}
}

func TestFnv32_DistinctValues(t *testing.T) {
	inputs := []string{"hello", "world", "foo", "bar", "extractFunction", "extractImports"}
	seen := make(map[[4]byte]string)
	for _, s := range inputs {
		h := fnv32([]byte(s))
		if prev, ok := seen[h]; ok {
			t.Errorf("fnv32 collision: %q and %q both produce %x", prev, s, h)
		}
		seen[h] = s
	}
}

func TestFnv32_EmptyInput(t *testing.T) {
	h := fnv32([]byte{})
	_ = h
}

// ---------------------------------------------------------------------------
// walkBodyWithScope — variable ordinal normalization
// ---------------------------------------------------------------------------

func TestWalkBodyWithScope_SameStructureDifferentNamesMatch(t *testing.T) {
	scope1 := make(map[string]byte)
	scope2 := make(map[string]byte)

	tokens1 := make([]byte, 0, 2)
	tokens2 := make([]byte, 0, 2)

	name1 := "myVariable"
	name2 := "differentName"

	ord1a, seen := scope1[name1]
	if !seen {
		next := byte(len(scope1) + 1) //nolint:gosec // ordinal capped to 255 by caller
		scope1[name1] = next
		ord1a = next
	}
	ord1b, seen := scope1[name1]
	if !seen {
		t.Fatal("should have been seen")
	}

	ord2a, seen := scope2[name2]
	if !seen {
		next := byte(len(scope2) + 1) //nolint:gosec // ordinal capped to 255 by caller
		scope2[name2] = next
		ord2a = next
	}

	tokens1 = append(tokens1, ord1a, ord1b)
	tokens2 = append(tokens2, ord2a, ord2a)

	if !bytes.Equal(tokens1, tokens2) {
		t.Errorf("first-seen identifiers should get ordinal 1 regardless of name: %x != %x", tokens1, tokens2)
	}
}

func TestWalkBodyWithScope_DifferentReturnVariableProducesDifferentTokens(t *testing.T) {
	scope := make(map[string]byte)

	assign := func(name string) byte {
		if ord, ok := scope[name]; ok {
			return ord
		}
		next := byte(len(scope) + 1) //nolint:gosec // ordinal capped to 255 by caller
		scope[name] = next
		return next
	}

	ordA := assign("a")
	ordB := assign("b")

	returnA := []byte{ordA, ordB, ordA}
	returnB := []byte{ordA, ordB, ordB}

	if bytes.Equal(returnA, returnB) {
		t.Error("returning different variables should produce different token streams")
	}
}

// ---------------------------------------------------------------------------
// fnv32 — callee name hashing
// ---------------------------------------------------------------------------

func TestFnv32_DifferentCalleeNames(t *testing.T) {
	h1 := fnv32([]byte("extractFunction"))
	h2 := fnv32([]byte("extractImports"))
	if h1 == h2 {
		t.Error("extractFunction and extractImports should produce different hashes")
	}
}

func TestFnv32_SameCalleeName(t *testing.T) {
	h1 := fnv32([]byte("BuildDispatchTable"))
	h2 := fnv32([]byte("BuildDispatchTable"))
	if h1 != h2 {
		t.Error("same callee name should produce same hash")
	}
}

// ---------------------------------------------------------------------------
// walkInitRecursive — literal value hashing
// ---------------------------------------------------------------------------

func TestWalkInitRecursive_LiteralTokensIncludeValueHash(t *testing.T) {
	val1 := []byte(`"hello"`)
	val2 := []byte(`"world"`)

	h1 := fnv32(val1)
	h2 := fnv32(val2)

	tokens1 := []byte{types.FPLitStr, h1[0], h1[1], h1[2], h1[3]}
	tokens2 := []byte{types.FPLitStr, h2[0], h2[1], h2[2], h2[3]}

	if bytes.Equal(tokens1, tokens2) {
		t.Errorf("different string literals should produce different token streams: %x == %x", tokens1, tokens2)
	}
}

func TestWalkInitRecursive_SameLiteralProducesSameTokens(t *testing.T) {
	val := []byte(`"hello"`)
	h := fnv32(val)
	tokens1 := []byte{types.FPLitStr, h[0], h[1], h[2], h[3]}
	tokens2 := []byte{types.FPLitStr, h[0], h[1], h[2], h[3]}
	if !bytes.Equal(tokens1, tokens2) {
		t.Error("same literal should produce same token stream")
	}
}

func TestWalkInitRecursive_NumericLiteralsDistinguished(t *testing.T) {
	h42 := fnv32([]byte("42"))
	h100 := fnv32([]byte("100"))

	tok42 := []byte{types.FPLitNum, h42[0], h42[1], h42[2], h42[3]}
	tok100 := []byte{types.FPLitNum, h100[0], h100[1], h100[2], h100[3]}

	if bytes.Equal(tok42, tok100) {
		t.Error("numeric literals 42 and 100 should produce different token streams")
	}
}

// ---------------------------------------------------------------------------
// Container content hashing
// ---------------------------------------------------------------------------

func containerHash(tok byte, elems [][]byte) []byte {
	type h4 [4]byte
	hashes := make([]h4, 0, len(elems))
	for _, e := range elems {
		hashes = append(hashes, fnv32(e))
	}
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			if bytes.Compare(hashes[i][:], hashes[j][:]) > 0 {
				hashes[i], hashes[j] = hashes[j], hashes[i]
			}
		}
	}
	out := make([]byte, 0, 2+4*len(hashes))
	out = append(out, tok, byte(len(hashes))) //nolint:gosec // len capped to 255 by walkInitRecursive
	for _, h := range hashes {
		out = append(out, h[0], h[1], h[2], h[3])
	}
	return out
}

func TestContainerHash_DifferentElementsProduceDifferentHashes(t *testing.T) {
	elems1 := [][]byte{[]byte(`"function_definition"`), []byte(`"method_declaration"`)}
	elems2 := [][]byte{[]byte(`"primitive_type"`), []byte(`"type_identifier"`)}

	t1 := containerHash(types.FPArray, elems1)
	t2 := containerHash(types.FPArray, elems2)

	if bytes.Equal(t1, t2) {
		t.Errorf("arrays with different string elements should produce different tokens:\n  t1: %x\n  t2: %x", t1, t2)
	}
}

func TestContainerHash_SameElementsProduceSameHash(t *testing.T) {
	elems := [][]byte{[]byte(`"a"`), []byte(`"b"`), []byte(`"c"`)}
	t1 := containerHash(types.FPArray, elems)
	t2 := containerHash(types.FPArray, elems)
	if !bytes.Equal(t1, t2) {
		t.Error("same elements should produce same hash")
	}
}

func TestContainerHash_OrderIndependent(t *testing.T) {
	elems1 := [][]byte{[]byte(`"a": 1`), []byte(`"b": 2`)}
	elems2 := [][]byte{[]byte(`"b": 2`), []byte(`"a": 1`)}

	t1 := containerHash(types.FPDict, elems1)
	t2 := containerHash(types.FPDict, elems2)
	if !bytes.Equal(t1, t2) {
		t.Errorf("container hash should be order-independent:\n  t1: %x\n  t2: %x", t1, t2)
	}
}
