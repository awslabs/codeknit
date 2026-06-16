// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// runFingerprint executes `codeknit fingerprint` against inputDir using the
// compiled binary and returns the output file content.
func runFingerprint(t *testing.T, inputDir string, minSim, maxSim int, showAll bool) string {
	t.Helper()
	base := fixturesDir(t)
	outputDir, err := os.MkdirTemp(base, sanitizeName(t.Name())+"-fp-out-*")
	if err != nil {
		t.Fatalf("runFingerprint: mkdirtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(outputDir) })

	outputFile := filepath.Join(outputDir, "fingerprints.skt")

	args := []string{
		"fingerprint",
		inputDir,
		"-o", outputFile,
		"--min-similarity", strconv.Itoa(minSim),
		"--max-similarity", strconv.Itoa(maxSim),
	}
	if showAll {
		args = append(args, "--show-all")
	}

	cmd := exec.Command(binPath, args...) //nolint:gosec // test helper with controlled args
	cmd.Dir = inputDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		t.Fatalf("codeknit fingerprint failed: %v\nstderr: %s", runErr, stderr.String())
	}

	data, readErr := os.ReadFile(outputFile) //nolint:gosec // test output file
	if readErr != nil {
		t.Fatalf("runFingerprint: read output: %v", readErr)
	}
	return string(data)
}

// parseDuplicates extracts all "similarity:N%  path::name <-> path::name" lines.
func parseDuplicates(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "similarity:") {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

// hasDuplicate returns true if any duplicate line contains both symbols.
func hasDuplicate(duplicates []string, sym1, sym2 string) bool {
	for _, d := range duplicates {
		if !strings.Contains(d, sym1) || !strings.Contains(d, sym2) {
			continue
		}
		return true
	}
	return false
}

// similarityOf returns the similarity percentage for the first duplicate line
// containing both symbols, or -1 if not found.
func similarityOf(duplicates []string, sym1, sym2 string) int {
	for _, d := range duplicates {
		if !strings.Contains(d, sym1) || !strings.Contains(d, sym2) {
			continue
		}
		after := strings.TrimPrefix(d, "similarity:")
		pct := strings.SplitN(after, "%", 2)[0]
		n, err := strconv.Atoi(pct)
		if err != nil {
			return -1
		}
		return n
	}
	return -1
}

// ============================================================
// Test: Exact duplicate functions in Go
// ============================================================

func TestFingerprint_ExactDuplicateFunctions_Go(t *testing.T) {
	fixtures := map[string]string{
		"order.go": `package service

func processOrder(price int, qty int) int {
	total := price * qty
	if total > 100 {
		total = applyDiscount(total)
	}
	tax := calculateTax(total)
	return total + tax
}

func processRefund(amount int, qty int) int {
	total := amount * qty
	if total > 100 {
		total = applyDiscount(total)
	}
	tax := calculateTax(total)
	return total + tax
}

func applyDiscount(x int) int { return x }
func calculateTax(x int) int  { return x }
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 100, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "processOrder", "processRefund") {
		t.Errorf("expected processOrder and processRefund to be exact duplicates\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Near-duplicate functions (one extra branch)
// ============================================================

func TestFingerprint_NearDuplicateFunctions_Go(t *testing.T) {
	fixtures := map[string]string{
		"calc.go": `package calc

// computeA has an extra bounds-clamp branch that computeB lacks.
func computeA(x int, y int) int {
	result := x * y
	if result > 1000 {
		result = result / 2
	}
	if result > 500 {
		result = result - 100
	}
	if result > 200 {
		result = result - 50
	}
	if result < 0 {
		result = 0
	}
	tax := computeTaxA(result)
	result = result + tax
	if result > 900 {
		result = applyCapA(result)
	}
	return result + 1
}

// computeB is the same pipeline without the extra clamp branch.
func computeB(x int, y int) int {
	result := x * y
	if result > 1000 {
		result = result / 2
	}
	if result > 500 {
		result = result - 100
	}
	if result > 200 {
		result = result - 50
	}
	tax := computeTaxA(result)
	result = result + tax
	if result > 900 {
		result = applyCapA(result)
	}
	return result + 1
}

func computeC(a int, b int) int {
	total := a + b
	if total > 50 {
		return total * 2
	}
	return total
}

func computeTaxA(x int) int { return x / 10 }
func applyCapA(x int) int   { return x }
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 70, 99, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "computeA", "computeB") {
		t.Errorf("expected computeA and computeB to be near-duplicates\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
	if hasDuplicate(dups, "computeA", "computeC") {
		t.Errorf("computeA and computeC should not be near-duplicates")
	}
}

// ============================================================
// Test: Cross-language exact duplicates (Go + Python)
// ============================================================

func TestFingerprint_CrossLanguage_GoAndPython(t *testing.T) {
	fixtures := map[string]string{
		"service.go": `package service

func processOrder(price int, qty int) int {
	total := price * qty
	if total > 100 {
		total = applyDiscount(total)
	}
	tax := calculateTax(total)
	return total + tax
}

func applyDiscount(x int) int { return x }
func calculateTax(x int) int  { return x }
`,
		"service.py": `def process_order(price, qty):
    total = price * qty
    if total > 100:
        total = apply_discount(total)
    tax = calculate_tax(total)
    return total + tax

def apply_discount(x):
    return x

def calculate_tax(x):
    return x
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 90, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "processOrder", "process_order") {
		t.Errorf("expected Go processOrder and Python process_order to match cross-language\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Cross-language exact duplicates (Go + JavaScript)
// ============================================================

func TestFingerprint_CrossLanguage_GoAndJavaScript(t *testing.T) {
	fixtures := map[string]string{
		"handler.go": `package handler

func validateInput(value int, limit int) bool {
	if value < 0 {
		return false
	}
	if value > limit {
		return false
	}
	return true
}
`,
		"handler.js": `function validateInput(value, limit) {
    if (value < 0) {
        return false;
    }
    if (value > limit) {
        return false;
    }
    return true;
}
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 85, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "validateInput", "validateInput") {
		t.Errorf("expected Go and JS validateInput to match cross-language\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Identifier-blind — different variable names, same logic
// ============================================================

func TestFingerprint_IdentifierBlind_SameLogicDifferentNames(t *testing.T) {
	fixtures := map[string]string{
		"funcs.go": `package funcs

func computeDiscount(originalPrice int, discountRate int) int {
	savings := originalPrice * discountRate
	if savings > originalPrice {
		savings = originalPrice
	}
	return originalPrice - savings
}

func calculateRebate(baseAmount int, rebatePercent int) int {
	reduction := baseAmount * rebatePercent
	if reduction > baseAmount {
		reduction = baseAmount
	}
	return baseAmount - reduction
}
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 90, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "computeDiscount", "calculateRebate") {
		t.Errorf("expected computeDiscount and calculateRebate to match (same logic, different names)\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Structurally different functions do NOT match
// ============================================================

func TestFingerprint_StructurallyDifferentFunctions_NoMatch(t *testing.T) {
	fixtures := map[string]string{
		"funcs.go": `package funcs

func simpleAdd(a int, b int) int {
	return a + b
}

func complexProcess(x int, threshold int) int {
	result := x * 2
	if result > threshold {
		result = applyMax(result)
	}
	if result < 0 {
		result = 0
	}
	return result + threshold
}

func applyMax(x int) int { return x }
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 70, 100, false)
	dups := parseDuplicates(output)

	if hasDuplicate(dups, "simpleAdd", "complexProcess") {
		t.Errorf("simpleAdd and complexProcess should NOT match — they are structurally different")
	}
}

// ============================================================
// Test: Duplicate class shapes (Go structs)
// ============================================================

func TestFingerprint_DuplicateTypeShapes_GoStructs(t *testing.T) {
	fixtures := map[string]string{
		"models.go": `package models

type User struct {
	Name  string
	Email string
	Age   int
}

func (u *User) Validate() bool {
	if u.Name == "" {
		return false
	}
	return true
}

func (u *User) Save() error {
	return nil
}

type Account struct {
	Name  string
	Email string
	Age   int
}

func (a *Account) Validate() bool {
	if a.Name == "" {
		return false
	}
	return true
}

func (a *Account) Save() error {
	return nil
}
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 90, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "User", "Account") {
		t.Errorf("expected User and Account structs to match (same shape)\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Duplicate method bodies across Python classes
// ============================================================

func TestFingerprint_DuplicateMethodBodies_Python(t *testing.T) {
	fixtures := map[string]string{
		"models.py": `class User:
    def validate(self):
        if not self.name:
            raise ValueError("name required")
        return True

    def save(self):
        result = db.insert(self)
        return result

class Account:
    def check(self):
        if not self.name:
            raise ValueError("name required")
        return True

    def persist(self):
        result = db.insert(self)
        return result
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 90, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "validate", "check") {
		t.Errorf("expected validate and check methods to match\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
	if !hasDuplicate(dups, "save", "persist") {
		t.Errorf("expected save and persist methods to match\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Similarity range filtering
// ============================================================

func TestFingerprint_SimilarityRangeFiltering(t *testing.T) {
	fixtures := map[string]string{
		"funcs.go": `package funcs

func processA(x int, y int) int {
	total := x * y
	if total > 100 {
		total = reduce(total)
	}
	tax := computeTax(total)
	return total + tax
}

func processB(x int, y int) int {
	total := x * y
	if total > 100 {
		total = reduce(total)
	}
	tax := computeTax(total)
	return total + tax
}

func reduce(x int) int     { return x }
func computeTax(x int) int { return x }
`,
	}
	inputDir := writeFixture(t, fixtures)

	// Range 100-100: exact duplicates appear.
	outputExact := runFingerprint(t, inputDir, 100, 100, false)
	exactDups := parseDuplicates(outputExact)
	if !hasDuplicate(exactDups, "processA", "processB") {
		t.Errorf("100%% range should include exact duplicates\nDuplicates found:\n%s",
			strings.Join(exactDups, "\n"))
	}

	// Range 0-99: exact duplicates excluded.
	outputBelow := runFingerprint(t, inputDir, 0, 99, false)
	belowDups := parseDuplicates(outputBelow)
	if hasDuplicate(belowDups, "processA", "processB") {
		t.Errorf("0-99%% range should exclude 100%% matches")
	}
}

// ============================================================
// Test: show-all flag includes [fingerprints] section
// ============================================================

func TestFingerprint_ShowAllFlag(t *testing.T) {
	fixtures := map[string]string{
		"fn.go": `package fn

func doWork(x int) int {
	if x > 0 {
		return x * 2
	}
	return x
}
`,
	}
	inputDir := writeFixture(t, fixtures)

	// Without show-all: no [fingerprints] section.
	outputDefault := runFingerprint(t, inputDir, 75, 100, false)
	if strings.Contains(outputDefault, "[fingerprints]") {
		t.Error("default output should not contain [fingerprints] section")
	}
	if !strings.Contains(outputDefault, "[duplicates]") {
		t.Error("default output should always contain [duplicates] section")
	}

	// With show-all: [fingerprints] section present.
	outputShowAll := runFingerprint(t, inputDir, 75, 100, true)
	if !strings.Contains(outputShowAll, "[fingerprints]") {
		t.Error("show-all output should contain [fingerprints] section")
	}
}

// ============================================================
// Test: Top-level code fingerprinting (Python scripts)
// ============================================================

func TestFingerprint_TopLevelCode_Python(t *testing.T) {
	fixtures := map[string]string{
		"setup.py": `import os
import sys

DB_URL = os.getenv("DATABASE_URL")

if not DB_URL:
    print("No database URL configured")
    sys.exit(1)

for item in os.listdir("/tmp"):
    if item.endswith(".log"):
        os.remove(item)
`,
		"init.py": `import os
import sys

API_KEY = os.getenv("API_KEY")

if not API_KEY:
    print("No API key configured")
    sys.exit(1)

for f in os.listdir("/var/log"):
    if f.endswith(".log"):
        os.remove(f)
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 70, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "top-level", "top-level") {
		t.Errorf("expected top-level code in setup.py and init.py to match\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Variable/constant fingerprinting
// ============================================================

func TestFingerprint_DuplicateVariableInitializers(t *testing.T) {
	fixtures := map[string]string{
		"errors.go": `package errors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrMissing  = errors.New("missing")
var ErrTimeout  = errors.New("timeout")

var ErrNotFound2 = errors.New("not found")
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 90, 100, false)
	dups := parseDuplicates(output)

	// Different string values should NOT match — literal content is now fingerprinted.
	if hasDuplicate(dups, "ErrNotFound", "ErrMissing") {
		t.Errorf("ErrNotFound and ErrMissing have different values and should NOT match\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}

	// Same string value should match exactly.
	if !hasDuplicate(dups, "ErrNotFound", "ErrNotFound2") {
		t.Errorf("expected ErrNotFound and ErrNotFound2 to match (identical initializer)\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Similarity score accuracy
// ============================================================

func TestFingerprint_SimilarityScoreAccuracy(t *testing.T) {
	fixtures := map[string]string{
		"funcs.go": `package funcs

func fnA(x int, y int) int {
	total := x * y
	if total > 100 {
		total = cap(total)
	}
	return total + 1
}

func fnB(x int, y int) int {
	total := x * y
	if total > 100 {
		total = cap(total)
	}
	return total + 1
}

func fnC(items []int) int {
	sum := 0
	for _, v := range items {
		sum = sum + v
	}
	return sum
}

func cap(x int) int { return x }
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 0, 100, false)
	dups := parseDuplicates(output)

	// fnA and fnB are identical — must score 100%.
	sim := similarityOf(dups, "fnA", "fnB")
	if sim != 100 {
		t.Errorf("fnA and fnB are identical, expected similarity 100, got %d\nDuplicates:\n%s",
			sim, strings.Join(dups, "\n"))
	}

	// fnA and fnC are structurally different — should score below 70%.
	simAC := similarityOf(dups, "fnA", "fnC")
	if simAC >= 70 {
		t.Errorf("fnA and fnC are structurally different, expected similarity < 70, got %d", simAC)
	}
}

// ============================================================
// Test: No false positives for completely different code
// ============================================================

func TestFingerprint_NoFalsePositives_CompletelyDifferentCode(t *testing.T) {
	fixtures := map[string]string{
		"funcs.go": `package funcs

func binarySearch(arr []int, target int) int {
	lo, hi := 0, len(arr)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return -1
}

func greet(name string) string {
	return "Hello, " + name
}
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 70, 100, false)
	dups := parseDuplicates(output)

	if hasDuplicate(dups, "binarySearch", "greet") {
		t.Error("binarySearch and greet are completely different — should not be flagged as duplicates")
	}
}

// ============================================================
// Test: Polyglot codebase — multiple languages together
// ============================================================

func TestFingerprint_Polyglot_MultipleLanguages(t *testing.T) {
	fixtures := map[string]string{
		"auth.go": `package auth

func authenticate(token string, secret string) bool {
	valid := isSet(token) && isSet(secret)
	if !valid {
		return false
	}
	return verify(token, secret)
}

func isSet(s string) bool { return s != "" }
func verify(t string, s string) bool { return true }
`,
		"auth.py": `def authenticate(token, secret):
    valid = bool(token) and bool(secret)
    if not valid:
        return False
    return verify(token, secret)

def verify(t, s):
    return True
`,
		"auth.js": `function authenticate(token, secret) {
    const valid = Boolean(token) && Boolean(secret);
    if (!valid) {
        return false;
    }
    return verify(token, secret);
}

function verify(t, s) { return true; }
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 70, 100, false)
	dups := parseDuplicates(output)

	if !hasDuplicate(dups, "auth.go::authenticate", "auth.js::authenticate") {
		t.Errorf("Go and JS authenticate should match\nDuplicates found:\n%s",
			strings.Join(dups, "\n"))
	}
}

// ============================================================
// Test: Output format correctness
// ============================================================

func TestFingerprint_OutputFormat(t *testing.T) {
	fixtures := map[string]string{
		"fn.go": `package fn

func fnA(x int) int {
	if x > 0 {
		return x * 2
	}
	return x
}

func fnB(x int) int {
	if x > 0 {
		return x * 2
	}
	return x
}
`,
	}
	inputDir := writeFixture(t, fixtures)
	output := runFingerprint(t, inputDir, 100, 100, false)

	if !strings.Contains(output, "[duplicates]") {
		t.Error("output must contain [duplicates] section")
	}
	if !strings.Contains(output, "similarity range: 100%-100%") {
		t.Errorf("output must contain similarity range header, got:\n%s", output)
	}

	dups := parseDuplicates(output)
	if len(dups) == 0 {
		t.Fatal("expected at least one duplicate line")
	}
	for _, d := range dups {
		if !strings.Contains(d, "<->") {
			t.Errorf("duplicate line missing '<->': %q", d)
		}
		if !strings.Contains(d, "::") {
			t.Errorf("duplicate line missing '::' path separator: %q", d)
		}
		// Must NOT contain short symbol IDs like "S123".
		for _, p := range strings.Fields(d) {
			if len(p) > 1 && p[0] == 'S' {
				allDigits := true
				for _, c := range p[1:] {
					if c < '0' || c > '9' {
						allDigits = false
						break
					}
				}
				if allDigits {
					t.Errorf("duplicate line should not contain short symbol IDs like %q: %q", p, d)
				}
			}
		}
	}
}
