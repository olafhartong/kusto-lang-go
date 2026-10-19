// Package guard_test enforces package boundaries and public API snapshots.
package guard_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

const modulePrefix = "github.com/olafhartong/kusto-lang-go/"

var repoRoot = filepath.Join("..", "..")

// boundaryRules lists permitted imports within this module.
var boundaryRules = map[string][]string{
	"kustolang/engine":      {},
	"kustolang/environment": {"kustolang/engine"},
	"kustolang":             {"kustolang/engine", "kustolang/environment"},
}

func TestImportBoundaries(t *testing.T) {
	dirs := make([]string, 0, len(boundaryRules))
	for dir := range boundaryRules {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		allowed := boundaryRules[dir]
		for _, imported := range projectImports(t, filepath.Join(repoRoot, dir)) {
			if !slices.Contains(allowed, imported) {
				t.Errorf("%s imports %s; allowed project-local imports are %v", dir, imported, allowed)
			}
		}
	}
}

// projectImports returns the sorted, module-relative import paths of all
// project-local packages imported by the non-test Go files in dir.
func projectImports(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(path, modulePrefix) {
				seen[strings.TrimPrefix(path, modulePrefix)] = true
			}
		}
	}
	imports := make([]string, 0, len(seen))
	for path := range seen {
		imports = append(imports, path)
	}
	sort.Strings(imports)
	return imports
}
