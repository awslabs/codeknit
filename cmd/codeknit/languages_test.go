// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"codeknit/internal/config"
	"codeknit/internal/plugin"
)

func TestHeaderLanguageSelection(t *testing.T) {
	for _, tc := range []struct {
		mode, source, want string
	}{
		{"cpp", "#ifndef API_H\n#define API_H\nclass Store { public: int get(int key); };\n#endif\n", "Store"},
		{"cpp", "class Store { public: int get(int key); };", "Store"},
		{"c", "int read_value(int key);", "read_value"},
		{"", "#define LIMIT 10\ntypedef unsigned long size_type;\nint read_value(int key);", "LIMIT"},
	} {
		t.Run(tc.mode+"/"+tc.want, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "api.h")
			if err := os.WriteFile(path, []byte(tc.source), 0o600); err != nil {
				t.Fatal(err)
			}
			registry, _ := newRegistry(tc.mode)
			parser, ok := registry.Lookup(".h")
			if !ok {
				t.Fatal("header parser missing")
			}
			if tc.mode == "c" {
				sourceParser, _ := registry.Lookup(".c")
				if parser != sourceParser {
					t.Fatal("explicit C mode must use the C source parser")
				}
			}
			for _, fingerprint := range []bool{false, true} {
				symbols, _, err := plugin.ParseFile(parser, path, fingerprint)
				if err != nil {
					t.Fatalf("fingerprint=%t: %v", fingerprint, err)
				}
				found := false
				for _, sym := range symbols {
					found = found || sym.Name == tc.want
				}
				if !found {
					t.Errorf("missing %s: %v", tc.want, symbols)
				}
			}
		})
	}
}

func TestHeaderLanguageValidation(t *testing.T) {
	for _, mode := range []string{"", "c", "cpp", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			cfg := config.Common{InputPath: t.TempDir(), HeaderLanguage: mode}
			err := cfg.Validate()
			if (err != nil) != (mode == "invalid") {
				t.Fatalf("mode=%q: %v", mode, err)
			}
			if mode == "" && cfg.HeaderLanguage != config.DefaultHeaderLanguage {
				t.Fatal("missing default header language")
			}
		})
	}
}
