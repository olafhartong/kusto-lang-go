package kustolang

import "testing"

func TestNativeBinderFoundationalDeclarationAndBadOperatorSemantics(t *testing.T) {
	t.Run("typed declaration is void", func(t *testing.T) {
		query := ParseQuery(`declare query_parameters(x:long); print x`)
		NewBinder(NewGlobalState(nil)).TryBind(query)

		declaration := firstBoundNode[*NameAndTypeDeclaration](t, query)
		info, ok := declaration.GetSemanticInfo()
		if !ok || info.ReferencedSymbol() != Void || info.ResultType != &Void.TypeSymbol || len(info.Diagnostics) != 0 {
			t.Fatalf("name-and-type semantic info = %#v, want Void", info)
		}
	})

	t.Run("bad query operator is error", func(t *testing.T) {
		table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
		query := ParseQuery(`T |`)
		NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

		bad := firstBoundNode[*BadQueryOperator](t, query)
		info, ok := bad.GetSemanticInfo()
		if !ok || info.ReferencedSymbol() != Error || info.ResultType != &Error.TypeSymbol || len(info.Diagnostics) != 0 {
			t.Fatalf("bad-query semantic info = %#v, want Error", info)
		}
		pipe := firstBoundNode[*PipeExpression](t, query)
		pipeInfo, ok := pipe.GetSemanticInfo()
		if !ok || pipeInfo.ReferencedSymbol() != Error || pipeInfo.ResultType != &Error.TypeSymbol {
			t.Fatalf("bad-query pipe semantic info = %#v, want Error", pipeInfo)
		}
		assertFoundationalSyntaxDiagnosticCodes(t, query, "KS176")
	})
}

func TestNativeBinderMacroExpandScopeDeclarationSemantics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	for _, test := range []struct {
		name        string
		text        string
		diagnostics []string
	}{
		{name: "valid", text: `macro-expand entity_group[T] as Scope (Scope | project A)`},
		{name: "malformed body", text: `macro-expand entity_group[T] as Scope (Scope |`, diagnostics: []string{"KS176", "KS005"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			operator := firstBoundNode[*MacroExpandOperator](t, query)
			declaration := operator.ScopeReferenceName().EntityGroupReferenceName()
			info, ok := declaration.GetSemanticInfo()
			symbol, symbolOK := info.ReferencedSymbol().(*EntityGroupElementSymbol)
			if !ok || !symbolOK || symbol.SymbolName() != "Scope" || symbol.EntityGroup == nil || info.ResultType != &symbol.TypeSymbol {
				t.Fatalf("scope declaration semantic info = %#v symbol=%#v", info, symbol)
			}
			assertFoundationalSyntaxDiagnosticCodes(t, query, test.diagnostics...)
		})
	}
}

func assertFoundationalSyntaxDiagnosticCodes(t *testing.T, element SyntaxElement, want ...string) {
	t.Helper()
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(element, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != len(want) {
		t.Fatalf("syntax diagnostics = %#v, want codes %v", diagnostics, want)
	}
	for i, code := range want {
		if diagnostics[i].Code != code {
			t.Fatalf("syntax diagnostics = %#v, want codes %v", diagnostics, want)
		}
	}
}
