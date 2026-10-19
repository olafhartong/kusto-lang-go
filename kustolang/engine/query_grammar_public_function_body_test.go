package kustolang

import "testing"

func TestQueryGrammarPublicFunctionBodyForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	tests := []struct {
		name       string
		text       string
		statements int
		expression bool
		semicolon  bool
	}{
		{name: "empty", text: "{}"},
		{name: "final expression", text: "{ x }", expression: true},
		{name: "final expression and semicolon", text: "{ x; }", expression: true, semicolon: true},
		{name: "leading let", text: "{ let x = 1; x }", statements: 1, expression: true},
		{
			name:       "leading query parameters and let",
			text:       "{ declare query_parameters(x:long); let y = x; y; }",
			statements: 2,
			expression: true,
			semicolon:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := parseQueryGrammarProduction(t, grammar.FunctionBody, test.text)
			if body.OpenBrace().Kind() != SyntaxKindOpenBraceToken || body.CloseBrace().Kind() != SyntaxKindCloseBraceToken {
				t.Fatalf("braces = %v/%v", body.OpenBrace().Kind(), body.CloseBrace().Kind())
			}
			if body.Statements().Count() != test.statements {
				t.Fatalf("statement count = %d, want %d", body.Statements().Count(), test.statements)
			}
			if (body.Expression() != nil) != test.expression {
				t.Fatalf("expression = %#v, want present %t", body.Expression(), test.expression)
			}
			if (body.Semicolon() != nil) != test.semicolon {
				t.Fatalf("semicolon = %#v, want present %t", body.Semicolon(), test.semicolon)
			}
			if diagnostics := body.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want none", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicFunctionBodyLeadingStatementSemicolons(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	letBody := parseQueryGrammarProduction(t, grammar.FunctionBody, "{ let x = 1 x }")
	if letBody.Statements().Count() != 1 || letBody.Expression() == nil {
		t.Fatalf("recovered let body = statements:%d expression:%#v", letBody.Statements().Count(), letBody.Expression())
	}
	assertMissingFunctionBodySemicolon(t, letBody.Statements().At(0).Separator())

	parametersBody := parseQueryGrammarProduction(t, grammar.FunctionBody, "{ declare query_parameters(x:long) x }")
	if parametersBody.Statements().Count() != 1 || parametersBody.Expression() == nil {
		t.Fatalf("recovered query-parameters body = statements:%d expression:%#v", parametersBody.Statements().Count(), parametersBody.Expression())
	}
	if _, ok := parametersBody.Statements().At(0).Element().(*QueryParametersStatement); !ok {
		t.Fatalf("leading statement = %#v, want *QueryParametersStatement", parametersBody.Statements().At(0).Element())
	}
	assertMissingFunctionBodySemicolon(t, parametersBody.Statements().At(0).Separator())
}

func TestQueryGrammarPublicFunctionBodyBraceRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingOpen := parseQueryGrammarProduction(t, grammar.FunctionBody, "x }")
	if missingOpen.OpenBrace().Width() != 0 ||
		!diagnosticsContainMessage(missingOpen.OpenBrace().Diagnostics(), "KS005", "Expected: {") {
		t.Fatalf("missing open brace = %#v diagnostics:%#v", missingOpen.OpenBrace(), missingOpen.OpenBrace().Diagnostics())
	}
	if missingOpen.Expression() == nil || missingOpen.CloseBrace().Width() == 0 {
		t.Fatalf("missing-open recovery = expression:%#v close:%#v", missingOpen.Expression(), missingOpen.CloseBrace())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.FunctionBody, "{ x")
	if missingClose.CloseBrace().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.CloseBrace().Diagnostics(), "KS005", "Expected: }") {
		t.Fatalf("missing close brace = %#v diagnostics:%#v", missingClose.CloseBrace(), missingClose.CloseBrace().Diagnostics())
	}
}

func TestQueryGrammarPublicFunctionBodyShapeTagsAndCompletionHints(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	body := parseQueryGrammarProduction(t, grammar.FunctionBody, "{ let x = 1; x; }")

	if body.ChildCount() != 5 {
		t.Fatalf("FunctionBody child count = %d, want exactly 5", body.ChildCount())
	}
	wantHints := []CompletionHint{
		CompletionHintSyntax,
		CompletionHintNonScalar,
		CompletionHintExpression,
		CompletionHintSyntax,
		CompletionHintSyntax,
	}
	for index, want := range wantHints {
		if got := body.CompletionHint(index); got != want {
			t.Errorf("FunctionBody child %d completion hint = %v, want %v", index, got, want)
		}
	}
	if !queryGrammarParserTreeHasTag(grammar.FunctionBody, "<statement-list>", 0) {
		t.Fatal("FunctionBody parser graph is missing <statement-list> tag")
	}
	if grammar.functionDeclaration.Tag() != "<function-declaration>" {
		t.Fatalf("functionDeclaration tag = %q, want <function-declaration>", grammar.functionDeclaration.Tag())
	}
}

func TestQueryGrammarPrivateFunctionDeclarationFormsAndRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	plain := parseQueryGrammarProduction(t, grammar.functionDeclaration, "(x:long) { x }")
	if plain.ViewKeyword() != nil || plain.Parameters().Parameters().Count() != 1 || plain.Body().Expression() == nil {
		t.Fatalf("plain declaration = view:%#v parameters:%d body:%#v", plain.ViewKeyword(), plain.Parameters().Parameters().Count(), plain.Body())
	}

	view := parseQueryGrammarProduction(t, grammar.functionDeclaration, "view () {}")
	if view.ViewKeyword() == nil || view.ViewKeyword().Kind() != SyntaxKindViewKeyword {
		t.Fatalf("view declaration keyword = %#v", view.ViewKeyword())
	}
	if view.Parameters().Parameters().Count() != 0 || view.Body().Expression() != nil {
		t.Fatalf("view declaration = parameters:%d expression:%#v", view.Parameters().Parameters().Count(), view.Body().Expression())
	}

	missingBodyOpen := parseQueryGrammarProduction(t, grammar.functionDeclaration, "() x }")
	if missingBodyOpen.Body().OpenBrace().Width() != 0 || missingBodyOpen.Body().Expression() == nil {
		t.Fatalf("declaration missing-open recovery = %#v", missingBodyOpen.Body())
	}
}

func TestQueryGrammarFunctionBodyPredefinedRulesReusePublicBehavior(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	rules := NewPredefinedRuleParsers()
	const bodyText = "{ declare query_parameters(x:long); let y = x; y; }"

	publicBody := parseQueryGrammarProduction(t, grammar.FunctionBody, bodyText)
	predefinedBody := parsePredefined(t, rules.FunctionBody, bodyText).(*FunctionBody)
	if predefinedBody.Text(IncludeTriviaMinimal) != publicBody.Text(IncludeTriviaMinimal) ||
		predefinedBody.Statements().Count() != publicBody.Statements().Count() ||
		predefinedBody.Semicolon() == nil {
		t.Fatalf("predefined FunctionBody differs = public:%#v predefined:%#v", publicBody, predefinedBody)
	}

	const declarationText = "view (x:long) { x }"
	publicDeclaration := parseQueryGrammarProduction(t, grammar.functionDeclaration, declarationText)
	if publicDeclaration.ViewKeyword() == nil {
		t.Fatalf("query FunctionDeclaration rejected view = %#v", publicDeclaration)
	}
	declarationTokens := Lex(declarationText)
	predefinedResult := rules.FunctionDeclaration.ParseValue(NewArraySource(declarationTokens), 0)
	predefinedDeclaration, ok := predefinedResult.Value.(*FunctionDeclaration)
	if predefinedResult.Length < 0 || !ok {
		t.Fatalf("predefined FunctionDeclaration view rejection = %#v", predefinedResult)
	}
	if predefinedDeclaration.ViewKeyword() != nil {
		t.Fatalf("predefined FunctionDeclaration accepted view = %#v", predefinedDeclaration.ViewKeyword())
	}
	if predefinedResult.Length >= len(declarationTokens)-1 || len(predefinedDeclaration.ContainedDiagnostics()) == 0 {
		t.Fatalf("predefined FunctionDeclaration did not reject and recover from view = length:%d declaration:%#v", predefinedResult.Length, predefinedDeclaration)
	}

	predefinedMissingOpen := parsePredefined(t, rules.FunctionBody, "x }").(*FunctionBody)
	if predefinedMissingOpen.OpenBrace().Width() != 0 || predefinedMissingOpen.Expression() == nil {
		t.Fatalf("predefined FunctionBody did not preserve public recovery = %#v", predefinedMissingOpen)
	}
}

func TestQueryGrammarPublicFunctionBodyMissingOpenDiffersFromImperativeParser(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	const text = "x }"

	result := grammar.FunctionBody.ParseValue(NewArraySource(Lex(text)), 0)
	if result.Length < 0 || result.Value == nil || result.Value.OpenBrace().Width() != 0 {
		t.Fatalf("public FunctionBody missing-open result = %#v", result)
	}
	if imperative := ParseFunctionBody(text); imperative != nil {
		t.Fatalf("imperative ParseFunctionBody missing-open result = %#v, want nil", imperative)
	}
}

func assertMissingFunctionBodySemicolon(t *testing.T, semicolon *SyntaxToken) {
	t.Helper()
	if semicolon == nil {
		t.Fatal("recovered statement semicolon = nil")
	}
	if semicolon.Width() != 0 ||
		!diagnosticsContainMessage(semicolon.Diagnostics(), "KS005", "Expected: ;") {
		t.Fatalf("recovered statement semicolon = %#v diagnostics:%#v", semicolon, semicolon.Diagnostics())
	}
}
