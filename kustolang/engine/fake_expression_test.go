package kustolang

import "testing"

func TestFakeExpressionCreateBuildsTypedTokenLiteral(t *testing.T) {
	expression := FakeExpressionCreate(ScalarTypes.String.Type())
	literal, ok := expression.(*LiteralExpression)
	if !ok {
		t.Fatalf("fake expression type = %T, want *LiteralExpression", expression)
	}
	if literal.Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("literal kind = %v, want TokenLiteralExpression", literal.Kind())
	}
	if token := literal.Token(); token == nil || token.Kind() != SyntaxKindIdentifierToken || token.RawText() != "" {
		t.Fatalf("literal token = %#v, want missing identifier token", token)
	} else if !token.IsMissing() || !token.Clone(true).(*SyntaxToken).IsMissing() {
		t.Fatal("diagnostic-free missing token or its clone was not marked missing")
	}
	if got := ExpressionResultType(expression); got != ScalarTypes.String.Type() {
		t.Fatalf("literal result type = %#v, want string", got)
	}
	if !literal.IsBound() {
		t.Fatal("fake literal should have semantic info attached")
	}
}

func TestFakeExpressionCreateNamedBuildsNamedExpressionWithSemantics(t *testing.T) {
	expression := FakeExpressionCreateNamed("Answer", ScalarTypes.Long.Type())
	named, ok := expression.(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("fake named expression type = %T, want *SimpleNamedExpression", expression)
	}
	if got := ExpressionResultType(named); got != ScalarTypes.Long.Type() {
		t.Fatalf("named result type = %#v, want long", got)
	}
	if named.EqualToken() == nil || named.EqualToken().Kind() != SyntaxKindEqualToken || named.EqualToken().RawText() != "=" {
		t.Fatalf("equal token = %#v, want '=' token", named.EqualToken())
	}

	nameDeclaration := named.Name()
	if nameDeclaration == nil {
		t.Fatal("named expression missing name declaration")
	}
	tokenName, ok := nameDeclaration.Name().(*TokenName)
	if !ok {
		t.Fatalf("name declaration child = %T, want *TokenName", nameDeclaration.Name())
	}
	if tokenName.Name() == nil || tokenName.Name().RawText() != "Answer" {
		t.Fatalf("token name = %#v, want Answer", tokenName.Name())
	}

	inner, ok := named.Expression().(*LiteralExpression)
	if !ok {
		t.Fatalf("named inner expression = %T, want *LiteralExpression", named.Expression())
	}
	if got := ExpressionResultType(inner); got != ScalarTypes.Long.Type() {
		t.Fatalf("inner result type = %#v, want long", got)
	}
}
