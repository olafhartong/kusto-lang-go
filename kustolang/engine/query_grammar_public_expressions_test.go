package kustolang

import "testing"

func TestQueryGrammarPublicUnnamedExpressionArithmeticPrecedence(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "1 + 2 * 3 - 4 / 2")

	subtract, ok := expression.(*BinaryExpression)
	if !ok || subtract.Kind() != SyntaxKindSubtractExpression {
		t.Fatalf("root = %#v, want SubtractExpression", expression)
	}
	add, ok := subtract.Left().(*BinaryExpression)
	if !ok || add.Kind() != SyntaxKindAddExpression {
		t.Fatalf("subtract left = %#v, want AddExpression", subtract.Left())
	}
	if multiply, ok := add.Right().(*BinaryExpression); !ok || multiply.Kind() != SyntaxKindMultiplyExpression {
		t.Fatalf("add right = %#v, want MultiplyExpression", add.Right())
	}
	if divide, ok := subtract.Right().(*BinaryExpression); !ok || divide.Kind() != SyntaxKindDivideExpression {
		t.Fatalf("subtract right = %#v, want DivideExpression", subtract.Right())
	}
}

func TestQueryGrammarPublicUnnamedExpressionLogicalPrecedence(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "a == b or c == d and e != f")

	orExpression, ok := expression.(*BinaryExpression)
	if !ok || orExpression.Kind() != SyntaxKindOrExpression {
		t.Fatalf("root = %#v, want OrExpression", expression)
	}
	if equal, ok := orExpression.Left().(*BinaryExpression); !ok || equal.Kind() != SyntaxKindEqualExpression {
		t.Fatalf("or left = %#v, want EqualExpression", orExpression.Left())
	}
	andExpression, ok := orExpression.Right().(*BinaryExpression)
	if !ok || andExpression.Kind() != SyntaxKindAndExpression {
		t.Fatalf("or right = %#v, want AndExpression", orExpression.Right())
	}
	if equal, ok := andExpression.Left().(*BinaryExpression); !ok || equal.Kind() != SyntaxKindEqualExpression {
		t.Fatalf("and left = %#v, want EqualExpression", andExpression.Left())
	}
	if notEqual, ok := andExpression.Right().(*BinaryExpression); !ok || notEqual.Kind() != SyntaxKindNotEqualExpression {
		t.Fatalf("and right = %#v, want NotEqualExpression", andExpression.Right())
	}
}

func TestQueryGrammarPublicUnnamedExpressionParenthesizedForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	name := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "(Account)")
	parenthesized, ok := name.(*ParenthesizedExpression)
	if !ok {
		t.Fatalf("parenthesized name = %#v", name)
	}
	if reference, ok := parenthesized.Expression().(*NameReference); !ok || NameReferenceSimpleName(reference) != "Account" {
		t.Fatalf("parenthesized value = %#v, want Account name reference", parenthesized.Expression())
	}

	precedence := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "(1 + 2) * 3")
	multiply, ok := precedence.(*BinaryExpression)
	if !ok || multiply.Kind() != SyntaxKindMultiplyExpression {
		t.Fatalf("parenthesized precedence root = %#v, want MultiplyExpression", precedence)
	}
	left, ok := multiply.Left().(*ParenthesizedExpression)
	if !ok || left.Expression().Kind() != SyntaxKindAddExpression {
		t.Fatalf("multiply left = %#v, want parenthesized AddExpression", multiply.Left())
	}
}

func TestQueryGrammarPublicUnnamedExpressionNamesAndCalls(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	name := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "Account")
	reference, ok := name.(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != "Account" {
		t.Fatalf("simple name = %#v", name)
	}

	expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, `series_stats(value=DurationMs, bins=5)`)
	call, ok := expression.(*FunctionCallExpression)
	if !ok || NameReferenceSimpleName(call.Name()) != "series_stats" {
		t.Fatalf("function call = %#v", expression)
	}
	arguments := call.ArgumentList().Expressions()
	if arguments.Count() != 2 {
		t.Fatalf("argument count = %d, want 2", arguments.Count())
	}
	for i, wantName := range []string{"value", "bins"} {
		named, ok := arguments.At(i).Element().(*SimpleNamedExpression)
		if !ok || NameDeclarationSimpleName(named.Name()) != wantName {
			t.Fatalf("argument %d = %#v, want named argument %q", i, arguments.At(i).Element(), wantName)
		}
	}
}

func TestQueryGrammarPublicUnnamedExpressionSelectors(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	dotted := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "account.profile.name")
	outerPath, ok := dotted.(*PathExpression)
	if !ok {
		t.Fatalf("dotted selector = %#v", dotted)
	}
	if selector, ok := outerPath.Selector().(*NameReference); !ok || NameReferenceSimpleName(selector) != "name" {
		t.Fatalf("outer path selector = %#v, want name", outerPath.Selector())
	}
	innerPath, ok := outerPath.Expression().(*PathExpression)
	if !ok {
		t.Fatalf("outer path receiver = %#v", outerPath.Expression())
	}
	if selector, ok := innerPath.Selector().(*NameReference); !ok || NameReferenceSimpleName(selector) != "profile" {
		t.Fatalf("inner path selector = %#v, want profile", innerPath.Selector())
	}

	bracketed := parseQueryGrammarProduction(t, grammar.UnnamedExpression, `payload["items"][0]`)
	outerElement, ok := bracketed.(*ElementExpression)
	if !ok {
		t.Fatalf("bracket selector = %#v", bracketed)
	}
	if _, ok := outerElement.Selector().(*BracketedExpression); !ok {
		t.Fatalf("outer bracket selector = %#v", outerElement.Selector())
	}
	innerElement, ok := outerElement.Expression().(*ElementExpression)
	if !ok {
		t.Fatalf("outer bracket receiver = %#v", outerElement.Expression())
	}
	if _, ok := innerElement.Selector().(*BracketedExpression); !ok {
		t.Fatalf("inner bracket selector = %#v", innerElement.Selector())
	}
}

func TestQueryGrammarPublicUnnamedExpressionUnaryOperators(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		text string
		kind SyntaxKind
	}{
		{text: "+value", kind: SyntaxKindUnaryPlusExpression},
		{text: "-value", kind: SyntaxKindUnaryMinusExpression},
	} {
		t.Run(test.text, func(t *testing.T) {
			expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, test.text)
			unary, ok := expression.(*PrefixUnaryExpression)
			if !ok || unary.Kind() != test.kind {
				t.Fatalf("unary expression %q = %#v, want kind %v", test.text, expression, test.kind)
			}
			if reference, ok := unary.Expression().(*NameReference); !ok || NameReferenceSimpleName(reference) != "value" {
				t.Fatalf("unary operand %q = %#v, want value name reference", test.text, unary.Expression())
			}
		})
	}
}

func TestQueryGrammarPublicUnnamedExpressionMissingRHSKeywordDiagnostic(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	result := grammar.UnnamedExpression.ParseValue(NewArraySource(Lex("1 + where")), 0)
	if result.Length != 2 {
		t.Fatalf("missing-RHS parse length = %d, want keyword left unconsumed", result.Length)
	}
	expression := result.Value

	binary, ok := expression.(*BinaryExpression)
	if !ok || binary.Kind() != SyntaxKindAddExpression {
		t.Fatalf("missing-RHS expression = %#v, want AddExpression", expression)
	}
	right, ok := binary.Right().(*NameReference)
	if !ok {
		t.Fatalf("missing-RHS node = %#v, want NameReference", binary.Right())
	}
	const want = "Missing expression: If the keyword 'where' is intended to be part of an expression it needs to be bracketted as ['where']."
	if !diagnosticsContainMessage(right.Diagnostics(), "KS006", want) {
		t.Fatalf("missing-RHS diagnostics = %#v", right.Diagnostics())
	}
}

func TestQueryGrammarPublicNamedExpression(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.NamedExpression, "Total = Amount + Tax")

	named, ok := expression.(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("named expression = %#v, want SimpleNamedExpression", expression)
	}
	if got := NameDeclarationSimpleName(named.Name()); got != "Total" {
		t.Fatalf("named expression name = %q, want Total", got)
	}
	if named.EqualToken().Kind() != SyntaxKindEqualToken {
		t.Fatalf("named expression equals token = %v", named.EqualToken().Kind())
	}
	if named.Expression().Kind() != SyntaxKindAddExpression {
		t.Fatalf("named expression value = %#v, want AddExpression", named.Expression())
	}
	if diagnostics := named.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("named expression diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicUnnamedExpressionStringAndEqualityFamilies(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	has := parseQueryGrammarProduction(t, grammar.UnnamedExpression, `UserPrincipalName has "admin"`)
	if has.Kind() != SyntaxKindHasExpression {
		t.Fatalf("has expression = %#v", has)
	}
	in := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "EventID in (4624, 4625)")
	inExpression, ok := in.(*InExpression)
	if !ok || inExpression.Right().Expressions().Count() != 2 {
		t.Fatalf("in expression = %#v", in)
	}
	between := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "Score between (1 .. 10)")
	if couple, ok := between.(*BetweenExpression); !ok || couple.Right().First().Kind() != SyntaxKindLongLiteralExpression || couple.Right().Second().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("between expression = %#v", between)
	}
	hasAll := parseQueryGrammarProduction(t, grammar.UnnamedExpression, `Tags has_all ("a", "b")`)
	if expression, ok := hasAll.(*HasAllExpression); !ok || expression.Right().Expressions().Count() != 2 {
		t.Fatalf("has_all expression = %#v", hasAll)
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.UnnamedExpression, "Value in range(1, 2, 1)", 1)
}

func TestQueryGrammarPublicInFamiliesAcceptSingleCompleteQuery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		"Account in (Watchlist | project Account)",
		"Account in~ (Watchlist | project Account)",
		"Account !in (Watchlist | project Account)",
		"Account !in~ (Watchlist | project Account)",
		"Account has_any (Watchlist | project Account)",
		"Account has_all (Watchlist | project Account)",
	} {
		expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, text)
		var list *ExpressionList
		switch expression := expression.(type) {
		case *InExpression:
			list = expression.Right()
		case *HasAnyExpression:
			list = expression.Right()
		case *HasAllExpression:
			list = expression.Right()
		default:
			t.Fatalf("%q expression = %T", text, expression)
		}
		if list.Expressions().Count() != 1 {
			t.Fatalf("%q list count = %d, want 1", text, list.Expressions().Count())
		}
		if _, ok := list.Expressions().At(0).Element().(*PipeExpression); !ok {
			t.Fatalf("%q list element = %T, want PipeExpression", text, list.Expressions().At(0).Element())
		}
		if list.CloseParen().Width() == 0 || len(expression.ContainedDiagnostics()) != 0 {
			t.Fatalf("%q close/diagnostics = %d/%#v", text, list.CloseParen().Width(), expression.ContainedDiagnostics())
		}
	}
}

func TestQueryGrammarPublicInFamilyNestedQueryOwnsMalformedRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "Account has_any (Watchlist | where)").(*HasAnyExpression)
	list := expression.Right()
	pipe, ok := list.Expressions().At(0).Element().(*PipeExpression)
	if !ok {
		t.Fatalf("nested malformed list element = %T, want PipeExpression", list.Expressions().At(0).Element())
	}
	where, ok := pipe.Operator().(*FilterOperator)
	if !ok || !diagnosticsContain(where.Condition().Diagnostics(), "KS006") {
		t.Fatalf("nested malformed operator = %#v diagnostics:%#v", pipe.Operator(), expression.ContainedDiagnostics())
	}
	if list.CloseParen().Width() == 0 || diagnosticsContain(list.CloseParen().Diagnostics(), "KS005") {
		t.Fatalf("nested malformed close = %#v, want source close without KS005", list.CloseParen())
	}
}

func TestQueryGrammarPublicNamedExpressionRecoveryForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	dashed := parseQueryGrammarProduction(t, grammar.NamedExpression, "foo-bar=1")
	if expression, ok := dashed.(*SimpleNamedExpression); !ok || !diagnosticsContain(expression.Name().Diagnostics(), "KS228") {
		t.Fatalf("dashed named expression = %#v diagnostics:%#v", dashed, dashed.ContainedDiagnostics())
	}
	compound := parseQueryGrammarProduction(t, grammar.NamedExpression, "(left,right)=series_stats(value)")
	if expression, ok := compound.(*CompoundNamedExpression); !ok || expression.Names().Names().Count() != 2 {
		t.Fatalf("compound named expression = %#v", compound)
	}
	missing := parseQueryGrammarProduction(t, grammar.NamedExpression, "alias=")
	if expression, ok := missing.(*SimpleNamedExpression); !ok || !diagnosticsContainMessage(expression.Expression().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing named RHS = %#v diagnostics:%#v", missing, missing.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicFunctionArgumentRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	call := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "coalesce(1,)").(*FunctionCallExpression)
	arguments := call.ArgumentList().Expressions()
	if arguments.Count() != 2 || !diagnosticsContainMessage(arguments.At(1).Element().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("trailing argument recovery = count:%d diagnostics:%#v", arguments.Count(), call.ContainedDiagnostics())
	}
}
