// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fingerprint

import (
	"bytes"
	"testing"

	"codeknit/internal/plugin"
)

func TestStructuralTokensSkipsTaggedVariableOrdinal(t *testing.T) {
	tokens := []byte{
		plugin.FPVar, plugin.FPIf,
		plugin.FPReturn,
		plugin.FPVar, plugin.FPCall,
	}

	got := structuralTokens(tokens)
	want := []byte{plugin.FPReturn}
	if !bytes.Equal(got, want) {
		t.Fatalf("structuralTokens() = %x, want %x", got, want)
	}
}

func TestRawTokenEditSimilarityUsesShapePayloads(t *testing.T) {
	a := []byte{0xE0, 0xEB, 0x01, 0x02, 0xE1}
	b := []byte{0xE0, 0xEB, 0x01, 0x02, 0xE1}
	c := []byte{0xE0, 0xEB, 0x09, 0x09, 0xE1}

	if got := RawTokenEditSimilarity(a, b); got != 100 {
		t.Fatalf("identical shape score = %d, want 100", got)
	}
	if got := RawTokenEditSimilarity(a, c); got >= 100 {
		t.Fatalf("different shape score = %d, want less than 100", got)
	}
}
