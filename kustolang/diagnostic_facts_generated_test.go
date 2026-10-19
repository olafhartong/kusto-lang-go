package kustolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func diagnosticFactoryNames(t *testing.T, path, prefix string) map[string]struct{} {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]struct{}{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && strings.HasPrefix(function.Name.Name, prefix) {
			names[function.Name.Name] = struct{}{}
		}
	}
	return names
}
