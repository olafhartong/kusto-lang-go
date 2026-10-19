package kustolang

import "testing"

func TestQueryGrammarPublicPathElementSelectorForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, test := range []struct {
		name string
		text string
		want func(*testing.T, Expression)
	}{
		{name: "identifier", text: "Account", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "Account", SyntaxKindTokenName)
		}},
		{name: "at", text: "@", want: func(t *testing.T, expression Expression) {
			at, ok := expression.(*AtExpression)
			if !ok || at.AtToken().Kind() != SyntaxKindAtToken {
				t.Fatalf("at selector = %#v, want *AtExpression", expression)
			}
		}},
		{name: "client parameter", text: "{tenant}", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "tenant", SyntaxKindBracedName)
		}},
		{name: "kind keyword", text: "kind", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "kind", SyntaxKindTokenName)
		}},
		{name: "withsource keyword", text: "withsource", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "withsource", SyntaxKindTokenName)
		}},
		{name: "with_source keyword", text: "with_source", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "with_source", SyntaxKindTokenName)
		}},
		{name: "bracketed string name", text: `["Column Name"]`, want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "Column Name", SyntaxKindBracketedName)
		}},
		{name: "bracketed wildcard", text: "[Sec*2024]", want: func(t *testing.T, expression Expression) {
			assertQueryGrammarPathSelectorName(t, expression, "Sec*2024", SyntaxKindBracketedWildcardedName)
		}},
		{name: "generic bracketed expression", text: "[0]", want: func(t *testing.T, expression Expression) {
			bracketed, ok := expression.(*BracketedExpression)
			if !ok || bracketed.Expression().Kind() != SyntaxKindLongLiteralExpression {
				t.Fatalf("generic bracket selector = %#v, want bracketed long literal", expression)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, test.text)
			test.want(t, expression)
			if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("selector diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicPrimaryPathSelectorDataScope(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, value := range []string{"all", "hotcache"} {
		t.Run(value, func(t *testing.T) {
			expression := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "SigninLogs datascope="+value)
			scoped, ok := expression.(*DataScopeExpression)
			if !ok {
				t.Fatalf("datascope selector = %T, want *DataScopeExpression", expression)
			}
			assertQueryGrammarPathSelectorName(t, scoped.Expression(), "SigninLogs", SyntaxKindTokenName)
			clause := scoped.DataScopeClause()
			if clause.EqualToken().RawText() != "=" || clause.Value().RawText() != value {
				t.Fatalf("datascope clause = equal:%q value:%q", clause.EqualToken().RawText(), clause.Value().RawText())
			}
			if diagnostics := scoped.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("datascope diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicPrimaryPathSelectorInRootPipelineAndDottedPath(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	root := parseQueryGrammarProduction(t, grammar.Expression, "SigninLogs datascope=hotcache")
	if _, ok := root.(*DataScopeExpression); !ok {
		t.Fatalf("root path selector = %T, want *DataScopeExpression", root)
	}

	pipe := parseQueryGrammarProduction(t, grammar.Expression, "SigninLogs datascope=all | where true")
	pipeline, ok := pipe.(*PipeExpression)
	if !ok {
		t.Fatalf("pipeline root = %T, want *PipeExpression", pipe)
	}
	if _, ok := pipeline.Expression().(*DataScopeExpression); !ok {
		t.Fatalf("pipeline source = %T, want *DataScopeExpression", pipeline.Expression())
	}
	if _, ok := pipeline.Operator().(*FilterOperator); !ok {
		t.Fatalf("pipeline operator = %T, want *FilterOperator", pipeline.Operator())
	}

	dotted := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "Payload.with_source")
	path, ok := dotted.(*PathExpression)
	if !ok {
		t.Fatalf("dotted path = %T, want *PathExpression", dotted)
	}
	assertQueryGrammarPathSelectorName(t, path.Expression(), "Payload", SyntaxKindTokenName)
	assertQueryGrammarPathSelectorName(t, path.Selector(), "with_source", SyntaxKindTokenName)
}

func TestQueryGrammarPublicPrimaryPathSelectorDataScopeRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingEqual := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "SigninLogs datascope hotcache").(*DataScopeExpression)
	equal := missingEqual.DataScopeClause().EqualToken()
	if equal.Width() != 0 || !diagnosticsContainMessage(equal.Diagnostics(), "KS005", "Expected: =") {
		t.Fatalf("missing datascope equal = %#v", equal)
	}
	if missingEqual.DataScopeClause().Value().RawText() != "hotcache" {
		t.Fatalf("value after missing equal = %q", missingEqual.DataScopeClause().Value().RawText())
	}

	missingValue := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "SigninLogs datascope=").(*DataScopeExpression)
	value := missingValue.DataScopeClause().Value()
	if value.Width() != 0 || !diagnosticsContainMessage(value.Diagnostics(), "KS005", "Expected one of: all, hotcache") {
		t.Fatalf("missing datascope value = %#v", value)
	}

	tokens := Lex("SigninLogs datascope=coldcache")
	result := grammar.UnnamedExpression.ParseValue(NewArraySource(tokens), 0)
	scoped, ok := result.Value.(*DataScopeExpression)
	if result.Length != 3 || !ok || result.Length == len(tokens)-1 {
		t.Fatalf("invalid datascope prefix parse = %#v, want three tokens with coldcache unconsumed", result)
	}
	invalidValue := scoped.DataScopeClause().Value()
	if invalidValue.Width() != 0 || !diagnosticsContainMessage(invalidValue.Diagnostics(), "KS005", "Expected one of: all, hotcache") {
		t.Fatalf("invalid datascope recovery value = %#v", invalidValue)
	}
}

func TestQueryGrammarPublicPathElementSelectorBracketRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingBody := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "[]")
	bracketed, ok := missingBody.(*BracketedExpression)
	if !ok {
		t.Fatalf("missing bracket body = %T, want *BracketedExpression", missingBody)
	}
	if bracketed.Expression().Width() != 0 || !diagnosticsContainMessage(bracketed.Expression().Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing bracket body expression = %#v", bracketed.Expression())
	}
	if bracketed.CloseBracket().RawText() != "]" {
		t.Fatalf("missing bracket body close = %q", bracketed.CloseBracket().RawText())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.UnnamedExpression, "[0")
	open, ok := missingClose.(*BracketedExpression)
	if !ok {
		t.Fatalf("missing bracket close = %T, want *BracketedExpression", missingClose)
	}
	if open.Expression().Kind() != SyntaxKindLongLiteralExpression || open.CloseBracket().Width() != 0 ||
		!diagnosticsContainMessage(open.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("missing bracket close selector = %#v", open)
	}

	incompleteName := parseQueryGrammarProduction(t, grammar.UnnamedExpression, `["name"`)
	incompleteBracket, ok := incompleteName.(*BracketedExpression)
	if !ok || incompleteBracket.Expression().Kind() != SyntaxKindStringLiteralExpression ||
		incompleteBracket.CloseBracket().Width() != 0 ||
		!diagnosticsContainMessage(incompleteBracket.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("incomplete bracketed name recovery = %#v, want generic bracketed expression", incompleteName)
	}
}

func TestQueryGrammarPublicPathElementSelectorWildcardAdjacencyProfiles(t *testing.T) {
	const text = "[Sec * 2024]"

	strict := QueryGrammarFrom(NewGlobalState(nil))
	strictExpression := parseQueryGrammarProduction(t, strict.UnnamedExpression, text)
	strictBracketed, ok := strictExpression.(*BracketedExpression)
	if !ok || strictBracketed.Expression().Kind() != SyntaxKindMultiplyExpression {
		t.Fatalf("strict non-adjacent wildcard = %#v, want generic bracketed multiplication", strictExpression)
	}

	options := DefaultParseOptions.WithAllowNonAdjacentWildcardParts(true)
	permissive := QueryGrammarFrom(NewGlobalState(nil).WithParseOptions(options))
	expression := parseQueryGrammarProduction(t, permissive.UnnamedExpression, text)
	reference, ok := expression.(*NameReference)
	if !ok || reference.Name().Kind() != SyntaxKindBracketedWildcardedName || NameReferenceSimpleName(reference) != "Sec*2024" {
		t.Fatalf("permissive non-adjacent wildcard = %#v", expression)
	}
	pattern := reference.Name().(*BracketedWildcardedName).Pattern()
	if pattern.RawText() != "Sec * 2024" || pattern.ValueText() != "Sec*2024" {
		t.Fatalf("permissive wildcard pattern = raw:%q value:%q", pattern.RawText(), pattern.ValueText())
	}
}

func assertQueryGrammarPathSelectorName(t *testing.T, expression Expression, want string, wantKind SyntaxKind) {
	t.Helper()
	reference, ok := expression.(*NameReference)
	if !ok {
		t.Fatalf("path selector = %T, want *NameReference", expression)
	}
	if reference.Name().Kind() != wantKind || NameReferenceSimpleName(reference) != want {
		t.Fatalf("path selector name = kind:%v simple:%q, want kind:%v simple:%q",
			reference.Name().Kind(), NameReferenceSimpleName(reference), wantKind, want)
	}
}
