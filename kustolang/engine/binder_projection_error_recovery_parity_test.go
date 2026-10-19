package kustolang

import "testing"

func next192ProjectionErrorGlobals() (*GlobalState, *TableSymbol) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table)), table
}

func TestNext192NativeNamedErrorProjectionStructuralRollback(t *testing.T) {
	globals, table := next192ProjectionErrorGlobals()

	t.Run("missing equal", func(t *testing.T) {
		project := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project Bad=Missing").(*ProjectOperator)
		named := project.Expressions().At(0).Element().(*SimpleNamedExpression)
		named.Expression().(*NameReference).SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
		named.SyntaxNode.children[1] = NewMissingToken(SyntaxKindEqualToken, Diagnostic{})

		binder := NewBinder(globals)
		binder.rowScope = table
		if info := newSimpleNodeBinder(binder).bind(project.SyntaxNode); info != nil {
			t.Fatalf("missing-equal project returned semantic info %#v", info)
		}
		if project.IsBound() || named.Name().ReferencedSymbol() != nil {
			t.Fatal("missing-equal project committed projection semantics")
		}
	})

	t.Run("missing alias delimiter", func(t *testing.T) {
		project := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project ['Bad']=Missing").(*ProjectOperator)
		named := project.Expressions().At(0).Element().(*SimpleNamedExpression)
		named.Expression().(*NameReference).SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
		bracketed := named.Name().Name().(*BracketedName)
		bracketed.SyntaxNode.children[2] = NewMissingToken(SyntaxKindCloseBracketToken, Diagnostic{})

		binder := NewBinder(globals)
		binder.rowScope = table
		if info := newSimpleNodeBinder(binder).bind(project.SyntaxNode); info != nil {
			t.Fatalf("incomplete-alias project returned semantic info %#v", info)
		}
		if project.IsBound() || named.Name().ReferencedSymbol() != nil {
			t.Fatal("incomplete-alias project committed projection semantics")
		}
	})

	t.Run("unbound expression", func(t *testing.T) {
		extend := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "extend Bad=Missing").(*ExtendOperator)
		named := extend.Expressions().At(0).Element().(*SimpleNamedExpression)
		binder := NewBinder(globals)
		binder.rowScope = table

		if info := newSimpleNodeBinder(binder).bind(extend.SyntaxNode); info != nil {
			t.Fatalf("unbound extend returned semantic info %#v", info)
		}
		if extend.IsBound() || named.Name().ReferencedSymbol() != nil {
			t.Fatal("unbound extend committed projection semantics")
		}
	})

	t.Run("tabular expression binds and diagnoses", func(t *testing.T) {
		serialize := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "serialize Bad=T").(*SerializeOperator)
		named := serialize.Expressions().At(0).Element().(*SimpleNamedExpression)
		named.Expression().(*NameReference).SetSemanticInfo(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
		named.SetSemanticInfo(NewSemanticInfoForType(&table.TypeSymbol))
		binder := NewBinder(globals)
		binder.rowScope = table

		info := newSimpleNodeBinder(binder).bind(serialize.SyntaxNode)
		if info == nil || !hasDiagnosticCode(info.Diagnostics, "KS174") || !hasDiagnosticCode(info.Diagnostics, "KS108") {
			t.Fatalf("tabular serialize semantic info = %#v", info)
		}
		if named.Name().ReferencedSymbol() == nil {
			t.Fatal("tabular serialize did not commit projection semantics")
		}
	})
}
