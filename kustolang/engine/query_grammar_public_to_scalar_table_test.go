package kustolang

import "testing"

func TestQueryGrammarPublicToScalarAndToTableNodesAndFullExpressionBodies(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	scalar := parseQueryGrammarProduction(t, grammar.UnnamedExpression,
		"toscalar kind=nooptimization (SigninLogs | count)").(*ToScalarExpression)
	assertQueryGrammarToScalarTableChildren(t, scalar.SyntaxNode, SyntaxKindToScalarKeyword)
	assertQueryGrammarToScalarTableParameter(t, scalar.KindParameter())
	if scalar.ToScalar().Kind() != SyntaxKindToScalarKeyword || scalar.Expression().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("toscalar keyword/body = %v/%v, want ToScalarKeyword/PipeExpression", scalar.ToScalar().Kind(), scalar.Expression().Kind())
	}

	table := parseQueryGrammarProduction(t, grammar.UnnamedExpression,
		"totable kind=nooptimization (SigninLogs | count)").(*ToTableExpression)
	assertQueryGrammarToScalarTableChildren(t, table.SyntaxNode, SyntaxKindToTableKeyword)
	assertQueryGrammarToScalarTableParameter(t, table.KindParameter())
	if table.ToTable().Kind() != SyntaxKindToTableKeyword || table.Expression().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("totable keyword/body = %v/%v, want ToTableKeyword/PipeExpression", table.ToTable().Kind(), table.Expression().Kind())
	}

	if diagnostics := scalar.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid toscalar diagnostics = %#v", diagnostics)
	}
	if diagnostics := table.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid totable diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicToScalarAndToTableKindParameterRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, test := range []struct {
		name  string
		parse func(string) *NamedParameter
	}{
		{name: "toscalar", parse: func(text string) *NamedParameter {
			return parseQueryGrammarProduction(t, grammar.UnnamedExpression, text).(*ToScalarExpression).KindParameter()
		}},
		{name: "totable", parse: func(text string) *NamedParameter {
			return parseQueryGrammarProduction(t, grammar.UnnamedExpression, text).(*ToTableExpression).KindParameter()
		}},
	} {
		t.Run(test.name+" missing equal", func(t *testing.T) {
			parameter := test.parse(test.name + " kind nooptimization (1)")
			if parameter.EqualToken().Width() != 0 ||
				!diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
				t.Fatalf("missing equal token = %#v diagnostics:%#v", parameter.EqualToken(), parameter.ContainedDiagnostics())
			}
			if parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
				t.Fatalf("recovered parameter value = %v, want TokenLiteralExpression", parameter.Expression().Kind())
			}
		})

		t.Run(test.name+" missing value", func(t *testing.T) {
			parameter := test.parse(test.name + " kind= (1)")
			literal, ok := parameter.Expression().(*LiteralExpression)
			if !ok || literal.Token().Width() != 0 ||
				!diagnosticsContainMessage(parameter.Expression().Diagnostics(), "KS005", "Expected: nooptimization") {
				t.Fatalf("missing parameter value = %#v diagnostics:%#v", parameter.Expression(), parameter.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicToScalarAndToTableRequiredPartRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, test := range []struct {
		name  string
		text  string
		parts func(Expression) (*SyntaxToken, Expression, *SyntaxToken)
	}{
		{name: "toscalar missing all", text: "toscalar", parts: queryGrammarToScalarParts},
		{name: "totable missing all", text: "totable", parts: queryGrammarToTableParts},
		{name: "toscalar missing body", text: "toscalar()", parts: queryGrammarToScalarParts},
		{name: "totable missing body", text: "totable()", parts: queryGrammarToTableParts},
		{name: "toscalar missing close", text: "toscalar(1", parts: queryGrammarToScalarParts},
		{name: "totable missing close", text: "totable(1", parts: queryGrammarToTableParts},
	} {
		t.Run(test.name, func(t *testing.T) {
			expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, test.text)
			open, body, close := test.parts(expression)
			if test.text == "toscalar" || test.text == "totable" {
				if open.Width() != 0 || !diagnosticsContainMessage(open.Diagnostics(), "KS005", "Expected: (") {
					t.Fatalf("missing open paren = %#v", open)
				}
			}
			if test.text == "toscalar" || test.text == "totable" || test.text == "toscalar()" || test.text == "totable()" {
				if body.Width() != 0 || !diagnosticsContainMessage(body.Diagnostics(), "KS006", "Missing expression") {
					t.Fatalf("missing body = %#v diagnostics:%#v", body, body.Diagnostics())
				}
			}
			if test.text != "toscalar()" && test.text != "totable()" {
				if close.Width() != 0 || !diagnosticsContainMessage(close.Diagnostics(), "KS005", "Expected: )") {
					t.Fatalf("missing close paren = %#v", close)
				}
			}
		})
	}
}

func TestQueryGrammarPublicToScalarAndToTableRootPipelineAndSuffixBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	if _, ok := parseQueryGrammarProduction(t, grammar.Expression, "toscalar(1)").(*ToScalarExpression); !ok {
		t.Fatal("root toscalar did not produce ToScalarExpression")
	}
	if _, ok := parseQueryGrammarProduction(t, grammar.Expression, "totable(SigninLogs)").(*ToTableExpression); !ok {
		t.Fatal("root totable did not produce ToTableExpression")
	}
	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "totable(SigninLogs) | where true").(*PipeExpression)
	if _, ok := pipeline.Expression().(*ToTableExpression); !ok {
		t.Fatalf("pipeline source = %T, want *ToTableExpression", pipeline.Expression())
	}

	suffixed := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "toscalar(1).value[0]")
	element, ok := suffixed.(*ElementExpression)
	if !ok {
		t.Fatalf("toscalar suffix root = %T, want *ElementExpression", suffixed)
	}
	path, ok := element.Expression().(*PathExpression)
	if !ok {
		t.Fatalf("toscalar suffix receiver = %T, want *PathExpression", element.Expression())
	}
	if _, ok := path.Expression().(*ToScalarExpression); !ok {
		t.Fatalf("toscalar dotted receiver = %T, want *ToScalarExpression", path.Expression())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.UnnamedExpression, "totable(1).value", 4)
}

func TestQueryGrammarPublicToScalarAndToTableCompletionVisibility(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parsers := FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.Expression, 0)
	items := []CompletionItem{}
	for _, parser := range parsers {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}

	scalar, ok := completionItemByDisplayText(items, "toscalar")
	if !ok || scalar.Kind != CompletionKindScalarPrefix {
		t.Fatalf("toscalar completion = %#v, visible items:%#v", scalar, completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "totable"); ok {
		t.Fatalf("hidden totable completion became visible: %#v", completionDisplayTexts(items))
	}
}

func assertQueryGrammarToScalarTableChildren(t *testing.T, node *SyntaxNode, keyword SyntaxKind) {
	t.Helper()
	want := []SyntaxKind{keyword, SyntaxKindNamedParameter, SyntaxKindOpenParenToken, SyntaxKindPipeExpression, SyntaxKindCloseParenToken}
	if node.ChildCount() != len(want) {
		t.Fatalf("child count = %d, want %d", node.ChildCount(), len(want))
	}
	for i, wantKind := range want {
		if child := node.GetChild(i); child == nil || child.Kind() != wantKind {
			t.Fatalf("child %d = %#v, want kind %v", i, child, wantKind)
		}
	}
}

func assertQueryGrammarToScalarTableParameter(t *testing.T, parameter *NamedParameter) {
	t.Helper()
	if parameter == nil || NameDeclarationSimpleName(parameter.Name()) != "kind" ||
		parameter.EqualToken().RawText() != "=" || parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("kind parameter = %#v", parameter)
	}
}

func queryGrammarToScalarParts(expression Expression) (*SyntaxToken, Expression, *SyntaxToken) {
	node := expression.(*ToScalarExpression)
	return node.OpenParen(), node.Expression(), node.CloseParen()
}

func queryGrammarToTableParts(expression Expression) (*SyntaxToken, Expression, *SyntaxToken) {
	node := expression.(*ToTableExpression)
	return node.OpenParen(), node.Expression(), node.CloseParen()
}
