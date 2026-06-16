// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Go(t *testing.T) {
	fixtures := map[string]string{
		"pkg/models/user.go": `package models

type User struct {
	Name  string
	Email string
	Age   int
}

func NewUser(name, email string, age int) *User {
	u := &User{Name: name, Email: email, Age: age}
	if err := validateEmail(email); err != nil {
		return nil
	}
	return u
}

func validateEmail(email string) error {
	if len(email) == 0 {
		return nil
	}
	return nil
}

var DefaultUser = &User{Name: "guest", Email: "guest@example.com"}
`,
		"pkg/models/role.go": `package models

type Role interface {
	Permissions() []string
	HasPermission(perm string) bool
}

type AdminRole struct {
	Level int
}

func (a *AdminRole) Permissions() []string {
	return []string{"read", "write", "delete"}
}

func (a *AdminRole) HasPermission(perm string) bool {
	for _, p := range a.Permissions() {
		if p == perm {
			return true
		}
	}
	return false
}

type ReadOnlyRole struct{}

func (r *ReadOnlyRole) Permissions() []string {
	return []string{"read"}
}

func (r *ReadOnlyRole) HasPermission(perm string) bool {
	return perm == "read"
}
`,
		"pkg/services/auth.go": `package services

import "fmt"

type AuthService struct {
	secret string
	ttl    int
}

func NewAuthService(secret string, ttl int) *AuthService {
	return &AuthService{secret: secret, ttl: ttl}
}

func (a *AuthService) Authenticate(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty token")
	}
	return verifyToken(token, a.secret), nil
}

func (a *AuthService) Refresh(token string) (string, error) {
	ok, err := a.Authenticate(token)
	if err != nil || !ok {
		return "", fmt.Errorf("invalid token")
	}
	return generateToken(a.secret, a.ttl), nil
}

func verifyToken(token, secret string) bool {
	return len(token) > 0 && len(secret) > 0
}

func generateToken(secret string, ttl int) string {
	return fmt.Sprintf("%s:%d", secret, ttl)
}

const MaxTokenLength = 4096
`,
		"pkg/services/math.go": `package services

const MaxValue = 1000

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

func Compute(x, y int) int {
	sum := Add(Multiply(x, y), x)
	if sum > MaxValue {
		return MaxValue
	}
	return sum
}

var DefaultPrecision = 2
`,
		// Test files that should be excluded by default.
		"pkg/models/user_test.go": `package models

import "testing"

func TestNewUser(t *testing.T) {
	u := NewUser("alice", "alice@example.com", 30)
	if u == nil {
		t.Fatal("expected non-nil user")
	}
}
`,
		"pkg/services/math.test.go": `package services

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("expected 3")
	}
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
