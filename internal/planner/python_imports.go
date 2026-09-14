// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"path/filepath"
	"strings"
)

// A Python package exports declarations in its __init__ module. Descendant
// modules do not implicitly export their declarations through the package.
// Re-exports requiring another import traversal remain unresolved.
func (r *symbolResolver) filterPythonImport(candidates []string, module, importingFile string) []string {
	dir := filepath.Dir(importingFile)
	key := fileLocal{dir, module}
	selected, cached := r.pythonImportFiles[key]
	if !cached {
		selected = pythonImportFile(r.files, module, dir)
		r.pythonImportFiles[key] = selected
	}
	if selected == "" {
		return nil
	}
	var result []string
	for _, id := range candidates {
		if candidateFilePath(id) == selected {
			result = append(result, id)
		}
	}
	return result
}

func pythonImportFile(files []string, module, dir string) string {
	var modules []string
	if resolved, ok := resolveRelativeImport(module, dir); ok {
		modules = pythonModuleFiles(files, resolved, false)
	} else {
		module = strings.ReplaceAll(module, ".", "/")
		modules = pythonModuleFiles(files, filepath.Join(dir, module), false)
		if len(modules) == 0 {
			modules = pythonModuleFiles(files, module, true)
		}
	}
	// Locate the module before looking for its exported name. An existing
	// module without that name must not fall back to another project's module.
	if len(modules) != 1 {
		return ""
	}
	return modules[0]
}

func pythonModuleFiles(files []string, module string, suffix bool) []string {
	module = filepath.ToSlash(filepath.Clean(module))
	var result []string
	for _, fp := range files {
		if languageFamily(fp) != ".py" {
			continue
		}
		stem := strings.TrimSuffix(fp, filepath.Ext(fp))
		if filepath.Base(stem) == "__init__" {
			stem = filepath.Dir(stem)
		}
		stem = filepath.ToSlash(filepath.Clean(stem))
		if stem == module || suffix && strings.HasSuffix(stem, "/"+module) {
			result = append(result, fp)
		}
	}
	return result
}
