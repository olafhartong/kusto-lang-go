package kustolang

import (
	"strings"
	"testing"
)

func TestColumnHasSameNameAsVariableAnalyzerReportsColumnHidingLocalVariable(t *testing.T) {
	text := "let Account = 'alice'; T | project Account"
	code := Parse(text)
	variable := NewVariableSymbol("Account", ScalarTypes.String.Type(), true, "alice", nil)
	parsedNameDeclarationBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(variable, ScalarTypes.String.Type()))
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	reference := lastParsedNameReferenceBySimpleName(t, code, "Account")
	reference.SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))

	diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeColumnHasSameNameAsVariable ||
		diagnostic.Category != DiagnosticCategoryCorrectness ||
		diagnostic.Severity != DiagnosticSeverityWarning ||
		diagnostic.Start != strings.LastIndex(text, "Account") ||
		diagnostic.Length != len("Account") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestColumnHasSameNameAsVariableAnalyzerReportsColumnHidingParameter(t *testing.T) {
	text := "declare query_parameters(Account:string); T | project Account"
	code := Parse(text)
	parameter := NewParameterSymbol("Account", ScalarTypes.String.Type(), "")
	parsedNameDeclarationBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(parameter, ScalarTypes.String.Type()))
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	reference := lastParsedNameReferenceBySimpleName(t, code, "Account")
	reference.SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))

	diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Start != strings.LastIndex(text, "Account") {
		t.Fatalf("diagnostic = %#v", diagnostics[0])
	}
}

func TestColumnHasSameNameAsVariableAnalyzerSkipsUnboundNonScalarAndPathSelectors(t *testing.T) {
	text := "let Account = T; T | project Account, T.Account"
	code := Parse(text)
	table := NewTableSymbol("T", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	parsedNameDeclarationBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(NewVariableSymbol("Account", table, false, nil, nil), &table.TypeSymbol))
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	references := parsedNameReferencesBySimpleName(code, "Account")
	if len(references) < 2 {
		t.Fatalf("Account references = %d, want at least 2", len(references))
	}
	references[len(references)-2].SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	references[len(references)-1].SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	if diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(code, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("non-scalar diagnostics = %#v, want empty", diagnostics)
	}

	path := Parse("let Account = 'alice'; T | project T.Account")
	parsedNameDeclarationBySimpleName(t, path, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(NewVariableSymbol("Account", ScalarTypes.String.Type(), false, nil, nil), ScalarTypes.String.Type()))
	lastParsedNameReferenceBySimpleName(t, path, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	if diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(path, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("path diagnostics = %#v, want empty", diagnostics)
	}

	unbound := Parse("let Account = 'alice'; T | project Account")
	parsedNameDeclarationBySimpleName(t, unbound, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(NewVariableSymbol("Account", ScalarTypes.String.Type(), false, nil, nil), ScalarTypes.String.Type()))
	if diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}
}

func TestColumnHasSameNameAsVariableAnalyzerSkipsFunctionScopedDeclarations(t *testing.T) {
	text := "let f = (Account:string) { T | project Account }; T | project Account"
	code := Parse(text)
	parameter := NewParameterSymbol("Account", ScalarTypes.String.Type(), "")
	parsedNameDeclarationBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(parameter, ScalarTypes.String.Type()))
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	lastParsedNameReferenceBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	if diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Analyze(code, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("function-scoped diagnostics = %#v, want empty", diagnostics)
	}
}

func TestColumnHasSameNameAsVariableAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewColumnHasSameNameAsVariableAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeColumnHasSameNameAsVariable ||
		diagnostics[0].Category != DiagnosticCategoryCorrectness ||
		diagnostics[0].Severity != DiagnosticSeverityWarning ||
		diagnostics[0].Description != "A variable or parameter is ignored in favor of a column with the same name." ||
		diagnostics[0].Message != "A variable or parameter is ignored in favor of a column with the same name." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func TestKustoCodeSpeculativeReferencedSymbolFindsAttachedLocalDeclarations(t *testing.T) {
	text := "let Account = 'alice'; T | project Account"
	code := Parse(text)
	variable := NewVariableSymbol("Account", ScalarTypes.String.Type(), true, "alice", nil)
	parsedNameDeclarationBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(variable, ScalarTypes.String.Type()))

	position := strings.LastIndex(text, "Account")
	if got := code.GetSpeculativeReferencedSymbol(position, "Account", SymbolMatchLocal); got != nil {
		t.Fatalf("public speculative symbol without analyzed semantics = %#v, want nil", got)
	}
	if got := code.getAttachedSpeculativeReferencedSymbol(position, "Account", SymbolMatchLocal); got != variable {
		t.Fatalf("attached speculative symbol = %#v, want variable", got)
	}
	if got := code.getAttachedSpeculativeReferencedSymbol(0, "Account", SymbolMatchLocal); got != nil {
		t.Fatalf("speculative symbol before declaration = %#v, want nil", got)
	}
	if got := code.getAttachedSpeculativeReferencedSymbol(position, "Account", SymbolMatchColumn); got != nil {
		t.Fatalf("speculative non-local symbol = %#v, want nil", got)
	}
}

func parsedNameDeclarationBySimpleName(t *testing.T, code *KustoCode, name string) *NameDeclaration {
	t.Helper()
	var declaration *NameDeclaration
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if declaration != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*NameDeclaration); ok && NameDeclarationSimpleName(found) == name {
			declaration = found
		}
	})
	if declaration == nil {
		t.Fatalf("name declaration %q not found in %q", name, code.Text)
	}
	return declaration
}

func lastParsedNameReferenceBySimpleName(t *testing.T, code *KustoCode, name string) *NameReference {
	t.Helper()
	references := parsedNameReferencesBySimpleName(code, name)
	if len(references) == 0 {
		t.Fatalf("name reference %q not found in %q", name, code.Text)
	}
	return references[len(references)-1]
}

func parsedNameReferencesBySimpleName(code *KustoCode, name string) []*NameReference {
	var references []*NameReference
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if found, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(found) == name {
			references = append(references, found)
		}
	})
	return references
}
