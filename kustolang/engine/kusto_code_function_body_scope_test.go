package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeFunctionBodyScopeBoundariesProfiles(t *testing.T) {
	const text = `let Outer=1; let f=(Param:long){ let Inner=2; print Value=Param+Inner }; print Result=Outer`
	inside := strings.Index(text, "Param+Inner") + len("Param+")
	closeBrace := strings.Index(text, " }; print") + 1
	afterBody := closeBrace + 1
	afterDeclaration := strings.Index(text, "print Result")

	for _, profile := range kustoCodeStatementScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid function diagnostics = %#v", diagnostics)
			}

			for _, name := range []string{"Outer", "Param", "Inner"} {
				if symbolInScope(code, inside, name) == nil {
					t.Fatalf("%s missing inside function body", name)
				}
			}
			if symbolInScope(code, closeBrace, "Param") == nil {
				t.Fatal("function parameter missing at close-brace boundary")
			}
			if symbolInScope(code, closeBrace, "Inner") != nil {
				t.Fatal("function-body local remained visible at close-brace boundary")
			}
			for _, position := range []int{afterBody, afterDeclaration} {
				if symbolInScope(code, position, "Param") != nil || symbolInScope(code, position, "Inner") != nil {
					t.Fatalf("function scope leaked at position %d", position)
				}
			}
			if symbolInScope(code, afterDeclaration, "Outer") == nil || symbolInScope(code, afterDeclaration, "f") == nil {
				t.Fatal("completed outer declarations missing after function declaration")
			}
		})
	}
}

func TestKustoCodeMissingFunctionCloseRetainsEditorScopeProfiles(t *testing.T) {
	const text = `let Outer=1; let f=(Param:long){ let Inner=2; print Value=`

	for _, profile := range kustoCodeStatementScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing expression and close brace should retain syntax diagnostics")
			}
			for _, name := range []string{"Outer", "Param", "Inner"} {
				if symbolInScope(code, len(text), name) == nil {
					t.Fatalf("%s missing from recovered function-body scope", name)
				}
			}
			if got := code.GetSpeculativeReferencedSymbol(len(text), "Param", SymbolMatchLocal); got == nil {
				t.Fatal("function parameter missing from speculative recovered scope")
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, "Param", CompletionKindParameter)
			requireCompletionKind(t, items, "Inner", CompletionKindVariable)
		})
	}
}
