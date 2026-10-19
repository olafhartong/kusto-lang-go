package kustolang

import "testing"

func TestAsContextBuilderAddsPrecedingAsOperatorDeclaration(t *testing.T) {
	binder := NewBinder(nil)
	asOperator := testAsOperator("Scoped", ScalarTypes.String.Type())

	binder.AddAsContext(asOperator, asOperator.FullWidth())
	if binder.LocalScope.ContainsSymbol("Scoped") {
		t.Fatal("position at the as-operator end should not add declaration")
	}

	binder.AddAsContext(asOperator, asOperator.FullWidth()+1)
	symbols := []Symbol{}
	binder.LocalScope.GetSymbols("Scoped", SymbolMatchLocal, &symbols)
	if len(symbols) != 1 {
		t.Fatalf("scoped symbols = %#v", symbols)
	}
	variable, ok := symbols[0].(*VariableSymbol)
	if !ok || variable.SymbolName() != "Scoped" || variable.Type != ScalarTypes.String.Type() {
		t.Fatalf("as declaration symbol = %#v", symbols[0])
	}
}

func TestAsContextBuilderRequiresNameAndResultType(t *testing.T) {
	binder := NewBinder(nil)
	withoutType := testAsOperator("Untyped", nil)
	binder.AddAsContext(withoutType, withoutType.FullWidth()+1)
	if binder.LocalScope.ContainsSymbol("Untyped") {
		t.Fatal("as operator without semantic result type should not add declaration")
	}

	missingName := MustNewNameDeclarationTyped(
		MustNewTokenNameTyped(NewMissingToken(SyntaxKindIdentifierToken, DiagnosticMissingName())),
	)
	withoutName := MustNewAsOperatorTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindAsKeyword, "", SyntaxText(SyntaxKindAsKeyword))),
		AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter")),
		missingName,
	)
	withoutName.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()))
	binder.AddAsContext(withoutName, withoutName.FullWidth()+1)
	if binder.LocalScope.ContainsSymbol("") {
		t.Fatal("as operator without a simple name should not add declaration")
	}
}

func TestAsContextBuilderSkipsFunctionBodiesOutsidePosition(t *testing.T) {
	inside := NewBinder(nil)
	body := testFunctionBody(testAsOperator("Inside", ScalarTypes.Long.Type()))
	inside.AddAsContext(body, body.FullWidth()-1)
	if !inside.LocalScope.ContainsSymbol("Inside") {
		t.Fatal("position inside function body after as operator should add declaration")
	}

	outside := NewBinder(nil)
	outside.AddAsContext(body, body.FullWidth()+1)
	if outside.LocalScope.ContainsSymbol("Inside") {
		t.Fatal("position outside function body should skip nested as operators")
	}
}

func testAsOperator(name string, typ *TypeSymbol) *AsOperator {
	nameDeclaration := MustNewNameDeclarationTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", name))),
	)
	asOperator := MustNewAsOperatorTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindAsKeyword, "", SyntaxText(SyntaxKindAsKeyword))),
		AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter")),
		nameDeclaration,
	)
	if typ != nil {
		asOperator.SetSemanticInfo(NewSemanticInfoForType(typ))
	}
	return asOperator
}

func testFunctionBody(expression Expression) *FunctionBody {
	return MustNewFunctionBodyTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenBraceToken, "", "{")),
		AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>")),
		expression,
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindSemicolonToken, "", ";")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseBraceToken, "", "}")),
	)
}
