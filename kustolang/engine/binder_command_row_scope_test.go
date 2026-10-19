package kustolang

import "testing"

func TestCommandColumnRowScopeSurvivesMalformedLists(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	b := NewColumnSymbol("B", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a, b)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	profiles := []func(string, *GlobalState) *KustoCode{ParseWithGlobals, ParseWithGlobals}
	queries := []string{
		`.drop table T columns (A, )`,
		`.alter table T column-docstrings (A:, B:"label")`,
		`.alter table T column-docstrings (A:"id",`,
		`.alter table T policy roworder (A asc, B)`,
	}

	for _, parse := range profiles {
		for _, text := range queries {
			code := parse(text, globals)
			if len(code.GetSyntaxDiagnostics()) == 0 {
				t.Fatalf("malformed command %q has no syntax diagnostics", text)
			}
			NewBinder(globals).TryBind(code.Syntax())
			for name, want := range map[string]*ColumnSymbol{"A": a, "B": b} {
				for _, reference := range commandNamedReferences(code.Syntax())[name] {
					if reference.ReferencedSymbol() != want {
						t.Fatalf("%q column %s = %#v, want source column", text, name, reference.ReferencedSymbol())
					}
				}
			}
		}
	}
}

func TestCommandRowScopeRejectsUnknownTableAndColumn(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	profiles := []func(string, *GlobalState) *KustoCode{ParseWithGlobals, ParseWithGlobals}

	for _, parse := range profiles {
		unknownColumn := parse(`.drop table T columns (Missing)`, globals)
		NewBinder(globals).TryBind(unknownColumn.Syntax())
		missing := commandNamedReferences(unknownColumn.Syntax())["Missing"][0]
		if !diagnosticsContain(missing.SemanticDiagnostics(), "KS142") {
			t.Fatalf("unknown column diagnostics = %#v", missing.SemanticDiagnostics())
		}

		unknownTable := parse(`.drop table Unknown columns (A)`, globals)
		NewBinder(globals).TryBind(unknownTable.Syntax())
		refs := commandNamedReferences(unknownTable.Syntax())
		if !diagnosticsContain(refs["Unknown"][0].SemanticDiagnostics(), "KS204") {
			t.Fatalf("unknown table diagnostics = %#v", refs["Unknown"][0].SemanticDiagnostics())
		}
		if refs["A"][0].ReferencedSymbol() == a || !diagnosticsContain(refs["A"][0].SemanticDiagnostics(), "KS142") {
			t.Fatalf("unknown table leaked row scope: symbol=%#v diagnostics=%#v", refs["A"][0].ReferencedSymbol(), refs["A"][0].SemanticDiagnostics())
		}
	}
}

func commandNamedReferences(root SyntaxElement) map[string][]*NameReference {
	refs := map[string][]*NameReference{}
	WalkSyntax(root, func(element SyntaxElement) {
		if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && !syntaxElementIsMissing(reference) {
			name := NameReferenceSimpleName(reference)
			refs[name] = append(refs[name], reference)
		}
	})
	return refs
}
