// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package planner

import (
	"path/filepath"
	"slices"
	"strings"

	"codeknit/internal/plugin"
)

type symbolResolver struct {
	pythonImportFiles map[fileLocal]string
	declarations      map[declarationKey]string
	local             map[fileLocal][]string
	named             map[string][]string
	imports           map[fileLocal][]*plugin.Edge
	aliases           map[fileLocal]bool
	packages          map[string]string
	files             []string
}

type declarationKey struct {
	file, name string
	offset     uint
}

func newResolver(symbols []plugin.Symbol, edges map[string][]plugin.Edge) *symbolResolver {
	r := &symbolResolver{
		pythonImportFiles: make(map[fileLocal]string),
		declarations:      make(map[declarationKey]string),
		local:             make(map[fileLocal][]string), named: make(map[string][]string),
		imports: make(map[fileLocal][]*plugin.Edge), aliases: make(map[fileLocal]bool),
	}
	for i := range symbols {
		s := &symbols[i]
		name := s.EffectiveScopedName()
		key := fileLocal{s.FilePath, name}
		r.local[key] = append(r.local[key], s.ID)
		r.named[name] = append(r.named[name], s.ID)
		if s.ByteSpan != [2]uint{} {
			r.declarations[declarationKey{s.FilePath, name, s.DeclarationOffset()}] = s.ID
		}
	}
	for fp, es := range edges {
		for i := range es {
			e := &es[i]
			switch e.Kind {
			case plugin.EdgeImports:
				key := fileLocal{fp, e.From}
				r.imports[key] = append(r.imports[key], e)
			case plugin.EdgeAliases:
				scope := r.scopeName(fp, e.Scope, e.SourceOffset)
				r.aliases[fileLocal{fp, plugin.MakeScopedName(scope, e.From)}] = true
			}
		}
	}
	return r
}

func (r *symbolResolver) edge(fp string, e *plugin.Edge) plugin.Edge {
	out := plugin.Edge{Kind: e.Kind, TargetName: e.To}
	source := e.From
	scope := r.scopeName(fp, e.From, e.SourceOffset)
	sourceResolution := ""
	if e.Kind == plugin.EdgeAliases {
		scope = r.scopeName(fp, e.Scope, e.SourceOffset)
		source = plugin.MakeScopedName(scope, source)
		out.From = fp + "::" + source
	} else {
		candidates := r.local[fileLocal{fp, source}]
		if id := r.declarations[declarationKey{fp, source, e.SourceOffset}]; id != "" {
			candidates = []string{id}
		}
		out.From, sourceResolution, _ = chooseTarget(fp, source, candidates)
	}
	if e.Resolution != "" {
		// Explicit metadata endpoints (such as #include directives) have
		// symbols, but their existence does not establish a dependency.
		out.To, _, _ = chooseTarget(fp, e.To, r.local[fileLocal{fp, e.To}])
		out.Resolution = e.Resolution
	} else {
		out.To, out.Resolution, out.Candidates = r.target(fp, scope, e.To)
	}
	if e.Kind != plugin.EdgeAliases && sourceResolution != "" {
		// Package declarations can be canonicalized into another file.
		if candidates := r.packageCandidates(fp, source); len(candidates) == 1 {
			out.From = candidates[0]
		} else if len(r.local[fileLocal{fp, source}]) > 1 {
			out.Resolution = "ambiguous"
		} else {
			out.Resolution = "unresolved"
		}
	}
	return out
}

func (r *symbolResolver) scopeName(fp, name string, offset uint) string {
	if id := r.declarations[declarationKey{fp, name, offset}]; id != "" {
		return strings.TrimPrefix(id, fp+"::")
	}
	return name
}

func (r *symbolResolver) target(fp, owner, name string) (id, resolution string, alternatives []string) {
	name = strings.ReplaceAll(name, "::", ".")
	head, member, qualified := strings.Cut(name, ".")
	// Resolve in the enclosing lexical scopes before file-level imports.
	// Qualified names are preserved: obj.run cannot become an unrelated run.
	for scope := owner; scope != ""; scope = parentScope(scope) {
		scoped := plugin.MakeScopedName(scope, name)
		if r.aliases[fileLocal{fp, scoped}] {
			return fp + "::" + scoped, "unresolved", nil
		}
		if candidates := r.local[fileLocal{fp, scoped}]; len(candidates) > 0 {
			return chooseTarget(fp, name, candidates)
		}
		if qualified && r.hasBinding(fp, plugin.MakeScopedName(scope, head)) {
			return fp + "::" + scoped, "unresolved", nil
		}
	}

	if qualified && r.hasBinding(fp, head) {
		if r.aliases[fileLocal{fp, name}] {
			return fp + "::" + name, "unresolved", nil
		}
		return chooseTarget(fp, name, r.local[fileLocal{fp, name}])
	}
	imports := r.imports[fileLocal{fp, head}]
	if len(imports) > 0 {
		var candidates []string
		for _, imp := range imports {
			targetName := name
			if qualified {
				targetName = member
				if imp.TargetName != "" && imp.TargetName != "*" {
					targetName = imp.TargetName + "." + member
				}
			} else if imp.TargetName != "" {
				targetName = imp.TargetName
			}
			candidates = append(candidates, r.filterByImport(r.compatible(fp, r.named[targetName]), imp.To, fp)...)
		}
		// An external/unavailable import must never fall back to a local namesake.
		return chooseTarget(fp, name, candidates)
	}
	if r.aliases[fileLocal{fp, name}] {
		return fp + "::" + name, "unresolved", nil
	}
	if candidates := r.local[fileLocal{fp, name}]; len(candidates) > 0 {
		return chooseTarget(fp, name, candidates)
	}
	// Go and Java declarations share a package across files in one directory.
	// Other module systems require explicit imports or same-file declarations.
	if candidates := r.packageCandidates(fp, name); len(candidates) > 0 {
		return chooseTarget(fp, name, candidates)
	}
	return chooseTarget(fp, name, nil)
}

func (r *symbolResolver) hasBinding(fp, name string) bool {
	key := fileLocal{fp, name}
	return r.aliases[key] || len(r.local[key]) != 0
}

func parentScope(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[:i]
	}
	return ""
}

func chooseTarget(fp, name string, candidates []string) (id, resolution string, alternatives []string) {
	candidates = slices.Clone(candidates)
	slices.Sort(candidates)
	candidates = slices.Compact(candidates)
	switch len(candidates) {
	case 0:
		return fp + "::" + name, "unresolved", nil
	case 1:
		return candidates[0], "", nil
	default:
		return fp + "::" + name, "ambiguous", candidates
	}
}

func (r *symbolResolver) compatible(fp string, candidates []string) []string {
	var result []string
	for _, id := range candidates {
		if languageFamily(fp) == languageFamily(candidateFilePath(id)) {
			result = append(result, id)
		}
	}
	return result
}

func (r *symbolResolver) packageCandidates(fp, name string) []string {
	family := languageFamily(fp)
	if family != ".go" && family != ".java" && family != ".cs" {
		return nil
	}
	var result []string
	for _, id := range r.compatible(fp, r.named[name]) {
		candidateFile := candidateFilePath(id)
		if pkg := r.packages[fp]; pkg != "" && r.packages[candidateFile] != pkg {
			continue
		}
		if filepath.Dir(candidateFile) == filepath.Dir(fp) {
			result = append(result, id)
		}
	}
	return result
}

func languageFamily(fp string) string {
	switch ext := strings.ToLower(filepath.Ext(fp)); ext {
	case ".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".py", ".pyi":
		return ".py"
	case ".scala", ".sc":
		return ".scala"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx":
		return "cpp"
	default:
		return ext
	}
}

// filterByImport matches complete path components, never substrings such as
// "models" in "old_models". Relative imports cannot fall back to another tree.
func (r *symbolResolver) filterByImport(candidates []string, modulePath, importingFile string) []string {
	if languageFamily(importingFile) == ".py" {
		return r.filterPythonImport(candidates, modulePath, importingFile)
	}
	dir := filepath.Dir(importingFile)
	if resolved, ok := resolveRelativeImport(modulePath, dir); ok {
		matches := matchingModule(candidates, resolved, false)
		if languageFamily(importingFile) == "javascript" {
			// A JS module is a file or an index module, not every descendant
			// declaration in the directory.
			var modules []string
			moduleStem := strings.TrimSuffix(resolved, filepath.Ext(resolved))
			if filepath.Ext(resolved) == "" {
				moduleStem = resolved
			}
			for _, id := range matches {
				fp := candidateFilePath(id)
				stem := strings.TrimSuffix(fp, filepath.Ext(fp))
				if stem == moduleStem || stem == filepath.Join(resolved, "index") {
					modules = append(modules, id)
				}
			}
			return modules
		}
		return matches
	}
	if languageFamily(importingFile) == "javascript" {
		// Bare specifiers depend on package exports or path configuration,
		// neither of which can be inferred from a matching repository filename.
		return nil
	}
	normalized := strings.ReplaceAll(modulePath, "::", "/")
	if !strings.Contains(normalized, "/") {
		normalized = strings.ReplaceAll(normalized, ".", "/")
	}
	// Python/Ruby bare imports commonly identify a sibling module.
	if family := languageFamily(importingFile); family == ".py" || family == ".rb" {
		if local := matchingModule(candidates, filepath.ToSlash(filepath.Join(dir, normalized)), false); len(local) > 0 {
			return local
		}
	}
	return matchingModule(candidates, normalized, true)
}

func matchingModule(candidates []string, module string, suffix bool) []string {
	module = filepath.ToSlash(filepath.Clean(module))
	var result []string
	for _, id := range candidates {
		fp := filepath.ToSlash(filepath.Clean(candidateFilePath(id)))
		stem := strings.TrimSuffix(fp, filepath.Ext(fp))
		match := fp == module || stem == module || strings.HasPrefix(fp, module+"/")
		if suffix {
			match = match || strings.HasSuffix(stem, "/"+module) || strings.Contains("/"+fp, "/"+module+"/")
		}
		if match {
			result = append(result, id)
		}
	}
	return result
}

func resolveRelativeImport(modulePath, importDir string) (string, bool) {
	if strings.HasPrefix(modulePath, "./") || strings.HasPrefix(modulePath, "../") {
		return filepath.ToSlash(filepath.Join(importDir, modulePath)), true
	}
	if strings.HasPrefix(modulePath, ".") {
		dots := len(modulePath) - len(strings.TrimLeft(modulePath, "."))
		for i := 1; i < dots; i++ {
			importDir = filepath.Dir(importDir)
		}
		return filepath.ToSlash(filepath.Join(importDir, strings.ReplaceAll(modulePath[dots:], ".", "/"))), true
	}
	return "", false
}

func candidateFilePath(id string) string {
	fp, _, _ := strings.Cut(id, "::")
	return fp
}

func uniqueEdges(edges []plugin.Edge) []plugin.Edge {
	type key struct{ from, to, kind, resolution string }
	seen := make(map[key]bool)
	result := make([]plugin.Edge, 0, len(edges))
	for i := range edges {
		e := &edges[i]
		k := key{e.From, e.To, string(e.Kind), e.Resolution}
		if !seen[k] {
			seen[k] = true
			result = append(result, *e)
		}
	}
	return result
}
