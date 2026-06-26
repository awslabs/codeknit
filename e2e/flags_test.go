// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"codeknit/internal/emitter/parser"

	"pgregory.net/rapid"
)

type parseJSONOutput struct {
	Files   []string `json:"files"`
	Symbols []struct {
		Name     string `json:"name"`
		File     string `json:"file"`
		Category string `json:"category"`
	} `json:"symbols"`
	Edges []struct {
		Kind string `json:"kind"`
	} `json:"edges,omitempty"`
}

// TestE2E_TestFileExclusion verifies that test-patterned files are excluded by
// default and included when --collect-test is passed.
func TestE2E_TestFileExclusion(t *testing.T) {
	fixtures := map[string]string{
		// Regular source files.
		"src/app.ts": `export class App {
  start(): void {}
}
`,
		"src/utils.js": `function helper() {
  return 42;
}
`,
		"lib/core.py": `class Core:
    def run(self):
        pass
`,
		// Test-patterned files that should be excluded by default.
		"src/app.test.ts": `import { App } from './app';
describe('App', () => {
  it('starts', () => {
    const app = new App();
    app.start();
  });
});
`,
		"src/utils.spec.js": `const { helper } = require('./utils');
describe('helper', () => {
  it('returns 42', () => {
    expect(helper()).toBe(42);
  });
});
`,
		"__tests__/helper.py": `import unittest

class TestHelper(unittest.TestCase):
    def test_something(self):
        self.assertTrue(True)
`,
	}

	inputDir := writeFixture(t, fixtures)

	testPatterns := []string{".test.", ".spec.", "__tests__/"}

	t.Run("default", func(t *testing.T) {
		sgDefault, outputDir := runcodeknit(t, inputDir)
		assertSnapshot(t, outputDir, inputDir)

		for _, sym := range sgDefault.Symbols {
			norm := normalizeFilePath(sym.FilePath)
			for _, pat := range testPatterns {
				if strings.Contains(norm, pat) {
					t.Errorf("test file should be excluded by default, but found symbol in %q", norm)
				}
			}
		}

		// Verify regular source files ARE present.
		regularFiles := []string{"app.ts", "utils.js", "core.py"}
		for _, want := range regularFiles {
			found := false
			for _, sym := range sgDefault.Symbols {
				if strings.HasSuffix(normalizeFilePath(sym.FilePath), want) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("regular file %q should be present in default output", want)
			}
		}
	})

	t.Run("with_collect_test", func(t *testing.T) {
		sgWithTest, outputDir := runcodeknit(t, inputDir, "--collect-test")
		assertSnapshot(t, outputDir, inputDir)

		foundTestFiles := make(map[string]bool)
		for _, sym := range sgWithTest.Symbols {
			norm := normalizeFilePath(sym.FilePath)
			for _, pat := range testPatterns {
				if strings.Contains(norm, pat) {
					foundTestFiles[norm] = true
				}
			}
		}
		if len(foundTestFiles) == 0 {
			t.Error("with --collect-test, expected test-patterned files in output but found none")
		}
	})
}

// TestE2E_CompressedOutput verifies that --minify produces dictionary-based output
// that is parseable and preserves the same symbol count as unminified output.
func TestE2E_CompressedOutput(t *testing.T) {
	fixtures := map[string]string{
		"src/models.go": `package src

type User struct {
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}
`,
		"src/service.go": `package src

type Service struct {
	Name string
}

func NewService(name string) *Service {
	return &Service{Name: name}
}

func (s *Service) Run() string {
	return s.Name
}
`,
		"src/handler.py": `class Handler:
    def __init__(self, name):
        self.name = name

    def handle(self, request):
        return process(request)

def process(request):
    return "ok"
`,
	}

	inputDir := writeFixture(t, fixtures)

	// Run with --minify and get raw output directory.
	minifiedDir := runcodeknitRaw(t, inputDir, "--minify")

	// Verify dict.skt exists and contains [dict] section.
	dictPath := filepath.Join(minifiedDir, "dict.skt")
	dictData, err := os.ReadFile(dictPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("dict.skt not found: %v", err)
	}
	if !strings.Contains(string(dictData), "[dict]") {
		t.Error("dict.skt missing [dict] section")
	}

	// Verify map files exist and do NOT contain [dict].
	matches, err := filepath.Glob(filepath.Join(minifiedDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("glob minified output: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no output files from --minify run")
	}
	sort.Strings(matches)

	for _, path := range matches {
		data, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		if strings.Contains(string(data), "[dict]") {
			t.Errorf("%s should not contain [dict] (dict lives in dict.skt)", filepath.Base(path))
		}
	}

	// Snapshot the raw minified output (preserves [dict] section).
	assertSnapshot(t, minifiedDir, inputDir)

	// Verify minified output is parseable by parser.ParseOutput.
	readers, closers := collectOutputReaders(t, minifiedDir)
	sgMinified, parseErr := parser.ParseOutput(readers, true)
	for _, c := range closers {
		_ = c.Close()
	}
	if parseErr != nil {
		t.Fatalf("ParseOutput failed on minified output: %v", parseErr)
	}

	// Run without --minify and compare symbol counts.
	sgNormal, _ := runcodeknit(t, inputDir)

	if len(sgMinified.Symbols) != len(sgNormal.Symbols) {
		t.Errorf("symbol count mismatch: minified=%d, normal=%d",
			len(sgMinified.Symbols), len(sgNormal.Symbols))
	}
}

// TestE2E_FlatSplitOutput verifies that directory-flat mode with a low --max-lines
// limit splits output across multiple map_NNN.skt files. The fixture is sized to
// produce exactly 2 output files at --max-lines 40.
func TestE2E_FlatSplitOutput(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models/user.go": `package models

type User struct {
	ID    int
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}

func ValidateUser(u *User) bool {
	return u.Name != "" && u.Email != ""
}
`,
		"pkg/models/product.go": `package models

type Product struct {
	ID    int
	Name  string
	Price float64
}

func NewProduct(name string, price float64) *Product {
	return &Product{Name: name, Price: price}
}
`,
		"pkg/store/user_store.go": `package store

import "myapp/pkg/models"

type UserStore struct {
	users []*models.User
}

func NewUserStore() *UserStore {
	return &UserStore{}
}

func (s *UserStore) Add(u *models.User) {
	s.users = append(s.users, u)
}

func (s *UserStore) FindByID(id int) *models.User {
	for _, u := range s.users {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (s *UserStore) Count() int {
	return len(s.users)
}
`,
		"pkg/store/product_store.go": `package store

import "myapp/pkg/models"

type ProductStore struct {
	products []*models.Product
}

func NewProductStore() *ProductStore {
	return &ProductStore{}
}

func (s *ProductStore) Add(p *models.Product) {
	s.products = append(s.products, p)
}

func (s *ProductStore) FindByID(id int) *models.Product {
	for _, p := range s.products {
		if p.ID == id {
			return p
		}
	}
	return nil
}
`,
		"pkg/service/user_service.go": `package service

import "myapp/pkg/models"

type UserService struct {
	store *UserStore
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Register(name, email string) *models.User {
	u := models.NewUser(name, email)
	return u
}

func (s *UserService) Lookup(id int) *models.User {
	return nil
}

func (s *UserService) Delete(id int) bool {
	return true
}
`,
		"pkg/service/product_service.go": `package service

import "myapp/pkg/models"

type ProductService struct {
	store *ProductStore
}

func NewProductService() *ProductService {
	return &ProductService{}
}

func (s *ProductService) Create(name string, price float64) *models.Product {
	p := models.NewProduct(name, price)
	return p
}

func (s *ProductService) Get(id int) *models.Product {
	return nil
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir, "--max-lines", "40")

	// Verify exactly 2 output files were produced.
	matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(matches))
	}

	assertTreeSnapshot(t, outputDir, inputDir)
}

// TestE2E_ErrorsInFlatOutput verifies that parse errors from broken source files
// are written to a separate warnings.skt file in flat mode.
func TestE2E_ErrorsInFlatOutput(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models/user.go": `package models

type User struct {
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}
`,
		// Broken Go file — unclosed struct.
		"pkg/models/broken.go": `package models

type Broken struct {
	Name string

func Oops() {
`,
		// Valid Python file.
		"pkg/service/handler.py": `class Handler:
    def __init__(self):
        self.ready = True

    def handle(self, request):
        return "ok"
`,
		// Broken Python file — bad indentation / syntax.
		"pkg/service/broken.py": `def broken(
    class Bad(
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)

	// Verify warnings.skt exists in the output.
	warningsPath := filepath.Join(outputDir, "warnings.skt")
	warningsContent, err := os.ReadFile(warningsPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("warnings.skt not found: %v", err)
	}
	if !strings.Contains(string(warningsContent), "[errors]") {
		t.Error("warnings.skt missing [errors] section")
	}

	assertTreeSnapshot(t, outputDir, inputDir)
}

// TestE2E_ErrorsInTreeOutput verifies that parse errors from broken source files
// are written to a separate warnings.skt file in directory-tree mode.
func TestE2E_ErrorsInTreeOutput(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models/user.go": `package models

type User struct {
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}
`,
		// Broken Go file — unclosed struct.
		"pkg/models/broken.go": `package models

type Broken struct {
	Name string

func Oops() {
`,
		// Valid Python file.
		"pkg/service/handler.py": `class Handler:
    def __init__(self):
        self.ready = True

    def handle(self, request):
        return "ok"
`,
		// Broken Python file — bad syntax.
		"pkg/service/broken.py": `def broken(
    class Bad(
`,
	}

	inputDir := writeFixture(t, fixtures)
	outputDir := runcodeknitRaw(t, inputDir, "--output-mode", "directory-tree")

	// Verify warnings.skt exists in the output root.
	errorsPath := filepath.Join(outputDir, "warnings.skt")
	errContent, err := os.ReadFile(errorsPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("warnings.skt not found: %v", err)
	}
	if !strings.Contains(string(errContent), "[errors]") {
		t.Error("warnings.skt missing [errors] section")
	}

	// Per-source .skt files should NOT contain [errors].
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		if filepath.Base(path) == "warnings.skt" || filepath.Ext(path) != ".skt" {
			return nil
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), "[errors]") {
			rel, _ := filepath.Rel(outputDir, path)
			t.Errorf("tree file %s should not contain [errors] section", rel)
		}
		return nil
	})

	assertTreeSnapshot(t, outputDir, inputDir)
}

// TestE2E_CompressedFlatOutput verifies that directory-flat mode with --minify embeds
// the [dict] section in the first map file and that minified codes are used throughout.
// Uses a multi-language fixture with enough symbols to exercise the dictionary.
func TestE2E_CompressedFlatOutput(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models/user.go": `package models

type User struct {
	ID    int
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}

func ValidateUser(u *User) bool {
	return u.Name != "" && u.Email != ""
}
`,
		"pkg/models/order.go": `package models

type Order struct {
	ID     int
	UserID int
	Total  float64
}

func NewOrder(userID int, total float64) *Order {
	return &Order{UserID: userID, Total: total}
}
`,
		"pkg/service/user_service.py": `class UserService:
    def __init__(self):
        self.users = []

    def create(self, name, email):
        return {"name": name, "email": email}

    def find_by_id(self, user_id):
        for u in self.users:
            if u["id"] == user_id:
                return u
        return None

    def delete(self, user_id):
        self.users = [u for u in self.users if u["id"] != user_id]
`,
		"pkg/service/order_service.py": `class OrderService:
    def __init__(self):
        self.orders = []

    def place(self, user_id, total):
        order = {"user_id": user_id, "total": total}
        self.orders.append(order)
        return order

    def get(self, order_id):
        for o in self.orders:
            if o["id"] == order_id:
                return o
        return None
`,
		"pkg/api/router.ts": `import { UserService } from '../service/user_service';
import { OrderService } from '../service/order_service';

export class Router {
  private userService: UserService;
  private orderService: OrderService;

  constructor() {
    this.userService = new UserService();
    this.orderService = new OrderService();
  }

  handle(path: string, body: unknown): unknown {
    if (path === '/users') {
      return this.userService.create(body);
    }
    if (path === '/orders') {
      return this.orderService.place(body);
    }
    return null;
  }
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	minifiedDir := runcodeknitRaw(t, inputDir, "--minify", "--max-lines", "30")

	// Verify dict.skt exists and contains [dict] section.
	dictPath := filepath.Join(minifiedDir, "dict.skt")
	dictData, err := os.ReadFile(dictPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("dict.skt not found: %v", err)
	}
	if !strings.Contains(string(dictData), "[dict]") {
		t.Error("dict.skt missing [dict] section")
	}

	matches, err := filepath.Glob(filepath.Join(minifiedDir, "map_*.skt"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no map_*.skt output files")
	}
	sort.Strings(matches)

	// No map file should contain [dict] — it lives in dict.skt.
	for i := 0; i < len(matches); i++ {
		data, readErr := os.ReadFile(matches[i]) //nolint:gosec // test file
		if readErr != nil {
			t.Fatalf("read %s: %v", matches[i], readErr)
		}
		if strings.Contains(string(data), "[dict]") {
			t.Errorf("map_%03d.skt should not contain [dict] (dict lives in dict.skt)", i+1)
		}
	}

	assertTreeSnapshot(t, minifiedDir, inputDir)

	// Verify parseable and symbol count matches non-minified.
	readers, closers := collectOutputReaders(t, minifiedDir)
	sgMinified, parseErr := parser.ParseOutput(readers, true)
	for _, c := range closers {
		_ = c.Close()
	}
	if parseErr != nil {
		t.Fatalf("ParseOutput failed: %v", parseErr)
	}

	sgNormal, _ := runcodeknit(t, inputDir)
	if len(sgMinified.Symbols) != len(sgNormal.Symbols) {
		t.Errorf("symbol count mismatch: minified=%d, normal=%d",
			len(sgMinified.Symbols), len(sgNormal.Symbols))
	}
}

// TestE2E_CompressedTreeOutput verifies that --minify with --output-mode directory-tree
// writes the dictionary as a separate dict.skt file in the output root, and that the
// per-source .skt files do NOT contain [dict].
func TestE2E_CompressedTreeOutput(t *testing.T) {
	fixtures := map[string]string{
		// Domain models — referenced by service and handler layers.
		"pkg/models/user.go": `package models

type User struct {
	ID    int
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}
`,
		"pkg/models/order.go": `package models

type Order struct {
	ID     int
	UserID int
	Total  float64
}

func NewOrder(userID int, total float64) *Order {
	return &Order{UserID: userID, Total: total}
}
`,
		// Service layer — uses models, called by handlers.
		"pkg/service/user_service.go": `package service

import "myapp/pkg/models"

type UserService struct {
	users []*models.User
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Create(name, email string) *models.User {
	u := models.NewUser(name, email)
	s.users = append(s.users, u)
	return u
}

func (s *UserService) FindByID(id int) *models.User {
	for _, u := range s.users {
		if u.ID == id {
			return u
		}
	}
	return nil
}
`,
		"pkg/service/order_service.go": `package service

import "myapp/pkg/models"

type OrderService struct {
	orders []*models.Order
}

func NewOrderService() *OrderService {
	return &OrderService{}
}

func (s *OrderService) Place(userID int, total float64) *models.Order {
	o := models.NewOrder(userID, total)
	s.orders = append(s.orders, o)
	return o
}
`,
		// HTTP handler layer — uses services, defines request/response types.
		"pkg/api/handler/user_handler.py": `from service.user_service import UserService

class UserRequest:
    def __init__(self, name, email):
        self.name = name
        self.email = email

class UserResponse:
    def __init__(self, user_id, name):
        self.user_id = user_id
        self.name = name

class UserHandler:
    def __init__(self):
        self.service = UserService()

    def create(self, request):
        user = self.service.create(request.name, request.email)
        return UserResponse(user.id, user.name)

    def get(self, user_id):
        user = self.service.find_by_id(user_id)
        if user is None:
            return None
        return UserResponse(user.id, user.name)
`,
		"pkg/api/handler/order_handler.py": `from service.order_service import OrderService

class OrderRequest:
    def __init__(self, user_id, total):
        self.user_id = user_id
        self.total = total

class OrderResponse:
    def __init__(self, order_id, user_id, total):
        self.order_id = order_id
        self.user_id = user_id
        self.total = total

class OrderHandler:
    def __init__(self):
        self.service = OrderService()

    def place(self, request):
        order = self.service.place(request.user_id, request.total)
        return OrderResponse(order.id, order.user_id, order.total)
`,
		// Router wiring — imports handlers.
		"pkg/api/router.ts": `import { UserHandler } from './handler/user_handler';
import { OrderHandler } from './handler/order_handler';

export class Router {
  private userHandler: UserHandler;
  private orderHandler: OrderHandler;

  constructor() {
    this.userHandler = new UserHandler();
    this.orderHandler = new OrderHandler();
  }

  route(path: string, body: unknown): unknown {
    if (path === '/users') {
      return this.userHandler.create(body);
    }
    if (path === '/orders') {
      return this.orderHandler.place(body);
    }
    return null;
  }
}
`,
	}

	inputDir := writeFixture(t, fixtures)

	// Run with --minify --output-mode directory-tree.
	minifiedDir := runcodeknitRaw(t, inputDir, "--minify", "--output-mode", "directory-tree")

	// Verify dict.skt exists in the output root and contains [dict].
	dictPath := filepath.Join(minifiedDir, "dict.skt")
	dictContent, err := os.ReadFile(dictPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("dict.skt not found: %v", err)
	}
	if !strings.Contains(string(dictContent), "[dict]") {
		t.Error("dict.skt missing [dict] section")
	}

	// Snapshot the full tree output (dict.skt + per-source .skt files).
	assertTreeSnapshot(t, minifiedDir, inputDir)

	// Verify per-source .skt files do NOT contain [dict].
	err = filepath.Walk(minifiedDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Base(path) == "dict.skt" {
			return nil
		}
		if filepath.Ext(path) != ".skt" {
			return nil
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // test file
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		if strings.Contains(string(data), "[dict]") {
			rel, _ := filepath.Rel(minifiedDir, path)
			t.Errorf("tree file %s should not contain [dict] section", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk output dir: %v", err)
	}

	// Verify the output is parseable when dict.skt is included first.
	readers, closers := collectOutputReaders(t, minifiedDir)
	sgMinified, parseErr := parser.ParseOutput(readers, true)
	for _, c := range closers {
		_ = c.Close()
	}
	if parseErr != nil {
		t.Fatalf("ParseOutput failed on tree minified output: %v", parseErr)
	}

	// Run without --minify and compare symbol counts.
	sgNormal, _ := runcodeknit(t, inputDir)

	if len(sgMinified.Symbols) != len(sgNormal.Symbols) {
		t.Errorf("symbol count mismatch: tree minified=%d, normal=%d",
			len(sgMinified.Symbols), len(sgNormal.Symbols))
	}
}

// testFilePattern holds a test-patterned filename template and its language extension.
type testFilePattern struct {
	path    string
	content string
}

// allTestFilePatterns returns test-patterned file variants across languages.
func allTestFilePatterns() []testFilePattern {
	return []testFilePattern{
		{
			path:    "src/app.test.ts",
			content: "export function testHelper(): void {}\n",
		},
		{
			path:    "src/utils.spec.js",
			content: "function specHelper() { return 1; }\n",
		},
		{
			path:    "__tests__/helper.py",
			content: "def test_helper():\n    return True\n",
		},
		{
			path:    "src/service.test.ts",
			content: "export class TestService { run(): void {} }\n",
		},
		{
			path:    "tests/math.spec.js",
			content: "function mathSpec() { return 2; }\n",
		},
	}
}

// regularFileFixtures returns non-test source files that always produce symbols.
func regularFileFixtures() map[string]string {
	return map[string]string{
		"src/main.ts":  "export function main(): void {}\n",
		"src/utils.js": "function helper() { return 42; }\n",
		"lib/core.py":  "def process(data):\n    return data\n",
	}
}

// Feature: e2e-cli-tests, Property 3: Test file flag controls inclusion
// For any fixture directory containing both regular source files and
// test-patterned files, running the CLI without --collect-test should exclude
// test-patterned files from the output, and running with --collect-test should
// include them.
func TestProperty_TestFileFlagControlsInclusion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Start with regular files.
		files := regularFileFixtures()

		// Draw a random non-empty subset of test-patterned files.
		allPatterns := allTestFilePatterns()
		mask := rapid.SliceOfN(rapid.Bool(), len(allPatterns), len(allPatterns)).Draw(t, "testMask")
		var selectedTestFiles []string
		for i, m := range mask {
			if m {
				files[allPatterns[i].path] = allPatterns[i].content
				selectedTestFiles = append(selectedTestFiles, allPatterns[i].path)
			}
		}
		// Ensure at least one test file is selected.
		if len(selectedTestFiles) == 0 {
			idx := rapid.IntRange(0, len(allPatterns)-1).Draw(t, "fallbackTest")
			files[allPatterns[idx].path] = allPatterns[idx].content
		}

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		testPatterns := []string{".test.", ".spec.", "__tests__/"}

		// Run without --collect-test: test files should be excluded.
		sgDefault := runcodeknitForProperty(t, inputDir)
		for _, sym := range sgDefault.Symbols {
			norm := normalizeFilePath(sym.FilePath)
			for _, pat := range testPatterns {
				if strings.Contains(norm, pat) {
					t.Fatalf("test file should be excluded by default, but found symbol in %q", norm)
				}
			}
		}

		// Run with --collect-test: test files should be included.
		sgWithTest := runcodeknitForProperty(t, inputDir, "--collect-test")
		foundTestFiles := make(map[string]bool)
		for _, sym := range sgWithTest.Symbols {
			norm := normalizeFilePath(sym.FilePath)
			for _, pat := range testPatterns {
				if strings.Contains(norm, pat) {
					foundTestFiles[norm] = true
				}
			}
		}
		if len(foundTestFiles) == 0 {
			t.Fatal("with --collect-test, expected test-patterned files in output but found none")
		}
	})
}

// Feature: e2e-cli-tests, Property 5: Minification preserves symbol count
// For any fixture directory, running the CLI with --minify and without
// --minify should produce the same number of symbols in the parsed output.
func TestProperty_MinificationPreservesSymbolCount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		langs := drawLangSubset(t)
		files := buildFixtureFromSubset(langs)

		inputDir := writeFixtureForProperty(t, files)
		defer func() { _ = os.RemoveAll(inputDir) }()

		// Run without --minify.
		sgNormal := runcodeknitForProperty(t, inputDir)

		// Run with --minify — need to parse with minified=true.
		outputDir := runcodeknitRawForProperty(t, inputDir, "--minify")
		defer func() { _ = os.RemoveAll(outputDir) }()

		matches, err := filepath.Glob(filepath.Join(outputDir, "map_*.skt"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("no output files from --minify run")
		}
		sort.Strings(matches)

		readers := make([]io.Reader, 0, len(matches)+1)
		closers := make([]io.Closer, 0, len(matches)+1)

		// Include dict.skt first so the parser can resolve minified codes.
		dictPath := filepath.Join(outputDir, "dict.skt")
		if df, openErr := os.Open(dictPath); openErr == nil { //nolint:gosec // test file
			closers = append(closers, df)
			readers = append(readers, df)
		}

		for _, path := range matches {
			f, openErr := os.Open(path) //nolint:gosec // test file
			if openErr != nil {
				t.Fatalf("open %s: %v", path, openErr)
			}
			closers = append(closers, f)
			readers = append(readers, f)
		}
		sgMinified, parseErr := parser.ParseOutput(readers, true)
		for _, c := range closers {
			_ = c.Close()
		}
		if parseErr != nil {
			t.Fatalf("ParseOutput failed on minified output: %v", parseErr)
		}

		if len(sgMinified.Symbols) != len(sgNormal.Symbols) {
			t.Fatalf("symbol count mismatch: minified=%d, normal=%d",
				len(sgMinified.Symbols), len(sgNormal.Symbols))
		}
	})
}

// TestE2E_EdgesFlag verifies that by default the [edges] section is omitted
// and that --edges includes it. The e2e helpers inject --edges automatically
// for the rest of the suite; this test bypasses that to check the real
// default behavior.
func TestE2E_EdgesFlag(t *testing.T) {
	// Use a Go fixture that produces both symbols and edges.
	fixtures := map[string]string{
		"pkg/models.go": `package pkg

type User struct {
	Name string
	Age  int
}

func NewUser(name string, age int) *User {
	return &User{Name: name, Age: age}
}
`,
		"pkg/service.go": `package pkg

type Service struct {
	Name string
}

func NewService(name string) *Service {
	return &Service{Name: name}
}

func (s *Service) Run() string {
	return s.Name
}
`,
	}
	inputDir := writeFixture(t, fixtures)

	// Run without --edges: should have zero edges (default behavior).
	sgDefault := runParseNoDefaults(t, inputDir)
	if len(sgDefault.Edges) != 0 {
		t.Fatalf("expected zero edges by default, got %d", len(sgDefault.Edges))
	}
	if len(sgDefault.Symbols) == 0 {
		t.Fatal("expected symbols in default output, got none")
	}

	// Run with --edges: should have edges but the same symbol count.
	sgWithEdges := runParseNoDefaults(t, inputDir, "--edges")
	if len(sgWithEdges.Edges) == 0 {
		t.Fatal("expected edges when --edges is passed, got none")
	}
	if len(sgWithEdges.Symbols) != len(sgDefault.Symbols) {
		t.Fatalf("symbol count mismatch: default=%d, with --edges=%d",
			len(sgDefault.Symbols), len(sgWithEdges.Symbols))
	}
}

// TestE2E_EdgesInlineMode verifies that inline output hides [edges] by
// default and includes it when --edges is passed.
func TestE2E_EdgesInlineMode(t *testing.T) {
	fixtures := map[string]string{
		"app.py": `class App:
    def __init__(self):
        self.name = "app"

    def run(self):
        self.start()

    def start(self):
        pass
`,
	}
	inputDir := writeFixture(t, fixtures)

	// Default inline: no [edges] section.
	outDefault, err := runcodeknitInline(t, inputDir, "parse", "--output-mode", "inline", inputDir)
	if err != nil {
		t.Fatalf("codeknit inline (default) failed: %v", err)
	}
	if strings.Contains(outDefault, "[edges]") {
		t.Fatal("expected no [edges] section in default inline output")
	}
	if !strings.Contains(outDefault, "[symbols]") {
		t.Fatal("expected [symbols] section in inline output")
	}

	// Inline with --edges: [edges] section present.
	outEdges, err := runcodeknitInline(t, inputDir, "parse", "--output-mode", "inline", "--edges", inputDir)
	if err != nil {
		t.Fatalf("codeknit inline (--edges) failed: %v", err)
	}
	if !strings.Contains(outEdges, "[edges]") {
		t.Fatal("expected [edges] section when --edges is passed")
	}
}

func TestE2E_JSONOutput(t *testing.T) {
	inputDir := writeFixture(t, map[string]string{
		"app.go": `package main

type User struct{}

func Save(u User) {}
`,
	})

	out, err := runcodeknitInline(t, inputDir, "parse", "--output-mode", "inline", "--format", "json", "--edges", inputDir)
	if err != nil {
		t.Fatalf("codeknit json output failed: %v", err)
	}

	var parsed parseJSONOutput
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("json output is not valid JSON: %v\n%s", err, out)
	}
	if len(parsed.Files) != 1 || parsed.Files[0] != "app.go" {
		t.Fatalf("files = %v, want [app.go]", parsed.Files)
	}
	if len(parsed.Symbols) == 0 {
		t.Fatalf("expected symbols in JSON output, got none:\n%s", out)
	}
	if len(parsed.Edges) == 0 {
		t.Fatalf("expected edges in JSON output with --edges, got none:\n%s", out)
	}
}
