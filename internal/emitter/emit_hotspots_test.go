// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeknit/internal/config"
	"codeknit/internal/hotspot"
)

func TestEmitHotspotsSKTAndJSON(t *testing.T) {
	now := time.Date(2026, time.July, 21, 0, 0, 0, 0, time.UTC)
	result := &hotspot.Result{
		GeneratedAt:     now,
		Since:           now.AddDate(-1, 0, 0),
		Confidence:      "medium",
		CommitsAnalyzed: 42,
		Hotspots: []hotspot.Entry{{
			File: "internal/service.go", Score: 0.91, HistoryScore: 0.9,
			StructureScore: 0.8, Commits: 12, Churn: 300, LastChanged: now,
		}},
		TemporalCoupling: []hotspot.Coupling{{
			Left: "a.go", Right: "b.go", Strength: 0.75, CoChanges: 3,
		}},
	}
	emitter := &Emitter{}

	sktPath := filepath.Join(t.TempDir(), "hotspots.skt")
	if err := emitter.EmitHotspots(result, &HotspotOptions{
		OutputPath: sktPath, Format: config.OutputFormatSKT,
	}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(sktPath) //nolint:gosec // controlled test output path
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "[hotspots]") ||
		!strings.Contains(string(content), "internal/service.go") ||
		!strings.Contains(string(content), "a.go <-> b.go") {
		t.Fatalf("unexpected SKT output:\n%s", content)
	}

	jsonPath := filepath.Join(t.TempDir(), "hotspots.json")
	if emitErr := emitter.EmitHotspots(result, &HotspotOptions{
		OutputPath: jsonPath, Format: config.OutputFormatJSON,
	}); emitErr != nil {
		t.Fatal(emitErr)
	}
	data, err := os.ReadFile(jsonPath) //nolint:gosec // controlled test output path
	if err != nil {
		t.Fatal(err)
	}
	var decoded hotspot.Result
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(decoded.Hotspots) != 1 || decoded.Hotspots[0].File != "internal/service.go" {
		t.Fatalf("unexpected JSON result: %#v", decoded)
	}
}
