package kustolang

// Ported from src/Kusto.Language/Syntax/FakeExpression.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func FakeExpressionCreate(resultType *TypeSymbol) Expression {
	expression := MustNewLiteralExpressionTyped(
		SyntaxKindTokenLiteralExpression,
		NewMissingToken(SyntaxKindIdentifierToken, Diagnostic{}),
	)
	expression.SetSemanticInfo(NewSemanticInfoForType(resultType))
	return expression
}

func FakeExpressionCreateNamed(name string, resultType *TypeSymbol) Expression {
	named := MustNewSimpleNamedExpressionTyped(
		MustNewNameDeclarationTyped(
			MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", name))),
		),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", SyntaxText(SyntaxKindEqualToken))),
		FakeExpressionCreate(resultType),
	)
	named.SetSemanticInfo(NewSemanticInfoForType(resultType))
	return named
}
