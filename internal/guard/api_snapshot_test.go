package guard_test

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var updateSnapshots = flag.Bool("update", false, "rewrite the public API snapshot golden files")

// TestPublicAPISnapshot pins the exported surface of the public packages so
// refactors (package moves, alias re-pointing) cannot change signatures
// unnoticed. Intentional changes are reviewed by regenerating the golden
// files with: go test ./internal/guard -run TestPublicAPISnapshot -update
func TestPublicAPISnapshot(t *testing.T) {
	for _, pkg := range []string{"kustolang", "kustolang/environment", "kustolang/engine"} {
		t.Run(pkg, func(t *testing.T) {
			got := strings.Join(publicAPI(t, filepath.Join(repoRoot, pkg)), "\n") + "\n"
			golden := filepath.Join("testdata", strings.ReplaceAll(pkg, "/", "_")+"_api.golden")
			if *updateSnapshots {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read snapshot (run with -update to create it): %v", err)
			}
			if got != string(want) {
				t.Fatalf("public API of %s drifted from %s:\n%s\nIf intentional, regenerate with -update.",
					pkg, golden, snapshotDiff(string(want), got))
			}
		})
	}
}

// publicAPI renders every exported declaration of the package in dir, with
// project-local package qualifiers expanded to full import paths so alias
// renames do not mask (or fake) signature changes.
func publicAPI(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var lines []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		qualify := qualifier(file)
		for _, decl := range file.Decls {
			lines = append(lines, exportedDeclarations(t, fset, decl, qualify)...)
		}
	}
	sort.Strings(lines)
	return lines
}

func exportedDeclarations(t *testing.T, fset *token.FileSet, decl ast.Decl, qualify func(string) string) []string {
	t.Helper()
	var lines []string
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if !d.Name.IsExported() || (d.Recv != nil && !exportedReceiver(d.Recv)) {
			return nil
		}
		signature := *d
		signature.Body = nil
		signature.Doc = nil
		lines = append(lines, qualify(render(t, fset, &signature)))
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if !s.Name.IsExported() {
					continue
				}
				bare := *s
				bare.Doc = nil
				bare.Comment = nil
				lines = append(lines, qualify("type "+render(t, fset, &bare)))
			case *ast.ValueSpec:
				keyword := "var"
				if d.Tok == token.CONST {
					keyword = "const"
				}
				typeText := ""
				if s.Type != nil {
					typeText = " " + qualify(render(t, fset, s.Type))
				}
				for _, ident := range s.Names {
					if ident.IsExported() {
						lines = append(lines, keyword+" "+ident.Name+typeText)
					}
				}
			}
		}
	}
	return lines
}

func exportedReceiver(recv *ast.FieldList) bool {
	if len(recv.List) != 1 {
		return false
	}
	expr := recv.List[0].Type
	for {
		switch typed := expr.(type) {
		case *ast.StarExpr:
			expr = typed.X
		case *ast.IndexExpr:
			expr = typed.X
		case *ast.Ident:
			return typed.IsExported()
		default:
			return false
		}
	}
}

// qualifier expands project-local import aliases (e.g. api.LocalData) to full
// import paths in rendered declaration text.
func qualifier(file *ast.File) func(string) string {
	replacements := map[string]string{}
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !strings.HasPrefix(path, modulePrefix) {
			continue
		}
		alias := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		replacements[alias] = path
	}
	if len(replacements) == 0 {
		return func(text string) string { return text }
	}
	patterns := make(map[*regexp.Regexp]string, len(replacements))
	for alias, path := range replacements {
		patterns[regexp.MustCompile(`\b`+regexp.QuoteMeta(alias)+`\.`)] = path + "."
	}
	return func(text string) string {
		for pattern, path := range patterns {
			text = pattern.ReplaceAllString(text, path)
		}
		return text
	}
}

func render(t *testing.T, fset *token.FileSet, node any) string {
	t.Helper()
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func snapshotDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var diff []string
	for i := 0; i < len(wantLines) || i < len(gotLines); i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			diff = append(diff, "- "+w, "+ "+g)
			if len(diff) >= 20 {
				diff = append(diff, "… (truncated)")
				break
			}
		}
	}
	return strings.Join(diff, "\n")
}
