// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestE2E_ScopedSymbolResolution verifies that same-named symbols in different
// scopes within the same file get distinct IDs and correct edge resolution
// through the full pipeline (parse → plan → emit).
//
// This is the core regression test for the scoped symbol fix: before the fix,
// two classes with a method named "save" in the same file would have their
// contains edges both point to the first "save", because the planner's
// localToGlobal map only stored the first occurrence of each bare name.
//
// We verify correctness by inspecting the raw .skt output directly (not the
// round-trip parser) since the output format uses short IDs (S1, S2...) that
// unambiguously identify each symbol.
func TestE2E_ScopedSymbolResolution(t *testing.T) {
	fixtures := map[string]string{
		// Python: two classes with identical method names in the same file.
		"models.py": `class User:
    def save(self):
        return validate_email(self.email)

    def validate(self):
        return len(self.name) > 0

    def delete(self):
        pass

class Order:
    def save(self):
        return validate_total(self.total)

    def validate(self):
        return self.total > 0

    def delete(self):
        pass

def validate_email(email):
    return "@" in email

def validate_total(total):
    return total >= 0
`,
		// Java: two classes with identical method names.
		"Services.java": `class UserService {
    void save(String name) {
        validate(name);
        log("saved user");
    }

    void validate(String input) {}

    void log(String msg) {}
}

class OrderService {
    void save(int orderId) {
        validate(orderId);
        log("saved order");
    }

    void validate(int id) {}

    void log(String msg) {}
}
`,
		// Go: two structs with identical method names via receivers.
		"handlers.go": `package handlers

type UserHandler struct{}
type OrderHandler struct{}

func (h *UserHandler) Handle() {
	h.Validate()
}

func (h *UserHandler) Validate() bool {
	return true
}

func (h *OrderHandler) Handle() {
	h.Validate()
}

func (h *OrderHandler) Validate() bool {
	return false
}
`,
		// TypeScript: two classes with identical method names.
		"controllers.ts": `class UserController {
    handle(): void {
        this.validate();
    }

    validate(): boolean {
        return true;
    }
}

class OrderController {
    handle(): void {
        this.validate();
    }

    validate(): boolean {
        return false;
    }
}
`,
		// Python: callable reference detection — function passed as argument.
		"pipeline.py": `def transform(data):
    return data.upper()

def filter_items(data):
    return [x for x in data if x]

def process(data, fn):
    return fn(data)

def run():
    process("hello", transform)
    process(items, filter_items)
`,
	}

	inputDir := writeFixture(t, fixtures)
	outputDir := runcodeknitRaw(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)

	// Read the raw output and parse short IDs for structural assertions.
	output := readAllSkt(t, outputDir, inputDir)

	// Build a short ID → symbol name map and a short ID → line number map
	// from the [symbols] section.
	symbolName := make(map[string]string) // S1 → "User", S2 → "save", etc.
	symbolLine := make(map[string]int)    // S1 → 1, S2 → 2, etc.
	symbolFile := make(map[string]string) // S1 → "models.py", etc.

	symRe := regexp.MustCompile(`^(S\d+)\s+\S+\s+L(\d+)-L\d+\s+(\S+)`)
	var currentFile string
	inSymbols := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[symbols]" {
			inSymbols = true
			continue
		}
		if trimmed == "[edges]" {
			inSymbols = false
			continue
		}
		if inSymbols && strings.HasPrefix(trimmed, "## ") {
			currentFile = trimmed[3:]
			continue
		}
		if inSymbols {
			if m := symRe.FindStringSubmatch(trimmed); m != nil {
				sid := m[1]
				startLine := 0
				for _, c := range m[2] {
					startLine = startLine*10 + int(c-'0')
				}
				// Name is the part of the signature before '(' or the whole thing.
				sig := m[3]
				name := sig
				if idx := strings.Index(sig, "("); idx >= 0 {
					name = sig[:idx]
				}
				symbolName[sid] = name
				symbolLine[sid] = startLine
				symbolFile[sid] = currentFile
			}
		}
	}

	// Parse contains edges from the [edges] section.
	// Format: "S1 --contains--> S2, S3, S4"
	containsChildren := make(map[string][]string) // parent SID → child SIDs
	edgeRe := regexp.MustCompile(`^(S\d+)\s+--(\w+)-->\s+(.+)`)
	for _, line := range strings.Split(output, "\n") {
		m := edgeRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		from := m[1]
		kind := m[2]
		targets := strings.Split(m[3], ", ")
		if kind == "contains" {
			for _, to := range targets {
				containsChildren[from] = append(containsChildren[from], strings.TrimSpace(to))
			}
		}
		if kind == "calls" {
			for _, to := range targets {
				containsChildren["calls:"+from] = append(containsChildren["calls:"+from], strings.TrimSpace(to))
			}
		}
	}

	// Helper: find the short ID for a symbol by name, file, and start line.
	findSID := func(name, file string, startLine int) string {
		for sid, n := range symbolName {
			if n == name && symbolFile[sid] == file && symbolLine[sid] == startLine {
				return sid
			}
		}
		return ""
	}

	// Helper: find the short ID for a unique symbol by name and file.
	findUniqueSID := func(name, file string) string {
		var match string
		for sid, n := range symbolName {
			if n == name && symbolFile[sid] == file {
				if match != "" {
					return "" // ambiguous
				}
				match = sid
			}
		}
		return match
	}

	// Helper: check if a parent's contains edges include a specific child.
	hasChild := func(parentSID, childSID string) bool {
		for _, c := range containsChildren[parentSID] {
			if c == childSID {
				return true
			}
		}
		return false
	}

	// ---- Python: User.save and Order.save must be distinct ----
	pyUserSID := findUniqueSID("User", "models.py")
	pyOrderSID := findUniqueSID("Order", "models.py")
	pyUserSaveSID := findSID("save", "models.py", 2)
	pyOrderSaveSID := findSID("save", "models.py", 12)

	if pyUserSID == "" || pyOrderSID == "" {
		t.Fatal("Python: User or Order class SID not found")
	}
	if pyUserSaveSID == "" || pyOrderSaveSID == "" {
		t.Fatal("Python: save method SIDs not found (expected at lines 2 and 12)")
	}
	if pyUserSaveSID == pyOrderSaveSID {
		t.Errorf("Python: User.save and Order.save have the same SID %s — scoping failed", pyUserSaveSID)
	}

	if !hasChild(pyUserSID, pyUserSaveSID) {
		t.Errorf("Python: %s (User) does not contain %s (User.save)", pyUserSID, pyUserSaveSID)
	}
	if hasChild(pyUserSID, pyOrderSaveSID) {
		t.Errorf("Python: %s (User) incorrectly contains %s (Order.save)", pyUserSID, pyOrderSaveSID)
	}
	if !hasChild(pyOrderSID, pyOrderSaveSID) {
		t.Errorf("Python: %s (Order) does not contain %s (Order.save)", pyOrderSID, pyOrderSaveSID)
	}
	if hasChild(pyOrderSID, pyUserSaveSID) {
		t.Errorf("Python: %s (Order) incorrectly contains %s (User.save)", pyOrderSID, pyUserSaveSID)
	}

	// ---- Java: UserService.save and OrderService.save must be distinct ----
	jUserSvcSID := findUniqueSID("UserService", "Services.java")
	jOrderSvcSID := findUniqueSID("OrderService", "Services.java")
	jUserSaveSID := findSID("save", "Services.java", 2)
	jOrderSaveSID := findSID("save", "Services.java", 13)

	if jUserSvcSID == "" || jOrderSvcSID == "" {
		t.Fatal("Java: UserService or OrderService SID not found")
	}
	if jUserSaveSID == "" || jOrderSaveSID == "" {
		t.Fatal("Java: save method SIDs not found")
	}

	if !hasChild(jUserSvcSID, jUserSaveSID) {
		t.Error("Java: UserService does not contain its save method")
	}
	if hasChild(jUserSvcSID, jOrderSaveSID) {
		t.Error("Java: UserService incorrectly contains OrderService.save")
	}
	if !hasChild(jOrderSvcSID, jOrderSaveSID) {
		t.Error("Java: OrderService does not contain its save method")
	}
	if hasChild(jOrderSvcSID, jUserSaveSID) {
		t.Error("Java: OrderService incorrectly contains UserService.save")
	}

	// ---- Go: UserHandler.Handle and OrderHandler.Handle must be distinct ----
	goUserHSID := findUniqueSID("UserHandler", "handlers.go")
	goOrderHSID := findUniqueSID("OrderHandler", "handlers.go")
	goUserHandleSID := findSID("Handle", "handlers.go", 6)
	goOrderHandleSID := findSID("Handle", "handlers.go", 14)

	if goUserHSID == "" || goOrderHSID == "" {
		t.Fatal("Go: UserHandler or OrderHandler SID not found")
	}
	if goUserHandleSID == "" || goOrderHandleSID == "" {
		t.Fatal("Go: Handle method SIDs not found")
	}

	if !hasChild(goUserHSID, goUserHandleSID) {
		t.Error("Go: UserHandler does not contain its Handle method")
	}
	if hasChild(goUserHSID, goOrderHandleSID) {
		t.Error("Go: UserHandler incorrectly contains OrderHandler.Handle")
	}
	if !hasChild(goOrderHSID, goOrderHandleSID) {
		t.Error("Go: OrderHandler does not contain its Handle method")
	}
	if hasChild(goOrderHSID, goUserHandleSID) {
		t.Error("Go: OrderHandler incorrectly contains UserHandler.Handle")
	}

	// ---- TypeScript: UserController.handle and OrderController.handle ----
	tsUserCSID := findUniqueSID("UserController", "controllers.ts")
	tsOrderCSID := findUniqueSID("OrderController", "controllers.ts")
	tsUserHandleSID := findSID("handle", "controllers.ts", 2)
	tsOrderHandleSID := findSID("handle", "controllers.ts", 12)

	if tsUserCSID == "" || tsOrderCSID == "" {
		t.Fatal("TS: UserController or OrderController SID not found")
	}
	if tsUserHandleSID == "" || tsOrderHandleSID == "" {
		t.Fatal("TS: handle method SIDs not found")
	}

	if !hasChild(tsUserCSID, tsUserHandleSID) {
		t.Error("TS: UserController does not contain its handle method")
	}
	if hasChild(tsUserCSID, tsOrderHandleSID) {
		t.Error("TS: UserController incorrectly contains OrderController.handle")
	}
	if !hasChild(tsOrderCSID, tsOrderHandleSID) {
		t.Error("TS: OrderController does not contain its handle method")
	}
	if hasChild(tsOrderCSID, tsUserHandleSID) {
		t.Error("TS: OrderController incorrectly contains UserController.handle")
	}

	// ---- Python callable refs: run() should call transform and filter_items ----
	runSID := findUniqueSID("run", "pipeline.py")
	transformSID := findUniqueSID("transform", "pipeline.py")
	filterSID := findUniqueSID("filter_items", "pipeline.py")

	if runSID == "" || transformSID == "" || filterSID == "" {
		t.Fatal("Python pipeline: missing run, transform, or filter_items SID")
	}

	runCalls := containsChildren["calls:"+runSID]
	if !containsStr(runCalls, transformSID) {
		t.Errorf("Python: run (%s) does not call transform (%s) — callable ref not detected", runSID, transformSID)
	}
	if !containsStr(runCalls, filterSID) {
		t.Errorf("Python: run (%s) does not call filter_items (%s) — callable ref not detected", runSID, filterSID)
	}
}

// readAllSkt reads all .skt files from outputDir, normalizes fixture paths,
// and returns the concatenated content.
func readAllSkt(t *testing.T, outputDir, inputDir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(outputDir, "*.skt"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	sort.Strings(matches)

	parts := make([]string, 0, len(matches))
	for _, path := range matches {
		data, readErr := os.ReadFile(path) //nolint:gosec // path is constructed from test output dir, not user input
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		content := strings.ReplaceAll(string(data), inputDir+"/", "")
		parts = append(parts, content)
	}
	return strings.Join(parts, "")
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
