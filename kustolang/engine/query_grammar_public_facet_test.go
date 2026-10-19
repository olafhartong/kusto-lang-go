package kustolang

import "testing"

func TestQueryGrammarPublicFacetPlacementTagAndRootAmbiguity(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	if got := grammar.facetOperator.Tag(); got != "<facet>" {
		t.Fatalf("facet parser tag = %q, want <facet>", got)
	}
	for _, parser := range []Parser[LexicalToken]{grammar.QueryOperator, grammar.FollowingPipeElementExpression} {
		if !queryGrammarParserTreeHasTag(parser, "<facet>", 0) {
			t.Fatalf("parser tag %q does not contain <facet>", parser.Tag())
		}
	}

	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, production.parser, "facet by Account")
			if _, ok := operator.(*FacetOperator); !ok {
				t.Fatalf("facet placement = %T, want *FacetOperator", operator)
			}
		})
	}

	rootTokens := Lex("facet by Account")
	rootResult := grammar.Expression.ParseValue(NewArraySource(rootTokens), 0)
	root, ok := rootResult.Value.(*NameReference)
	if rootResult.Length != 1 || !ok || NameReferenceSimpleName(root) != "facet" {
		t.Fatalf("ambiguous root facet-by parse = %#v, want name facet with by-tail", rootResult)
	}
	bare := parseQueryGrammarProduction(t, grammar.Expression, "facet")
	if _, ok := bare.(*NameReference); !ok {
		t.Fatalf("bare root facet = %T, want *NameReference", bare)
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "Events | facet by Account with count")
	pipe, operator := queryGrammarPipeStage(t, pipeline, SyntaxKindFacetOperator)
	if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("facet pipeline source = %#v, want Events", pipe.Expression())
	}
	if _, ok := operator.(*FacetOperator); !ok {
		t.Fatalf("facet pipeline operator = %T, want *FacetOperator", operator)
	}
}

func TestQueryGrammarPublicFacetSimplePathKeysAndOneOrMoreList(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.facetOperator,
		`facet by Account.Region, Payload["user"], {tenant}, kind, Payload.with_source, ["Display Name"]`).(*FacetOperator)

	if operator.FacetKeyword().Kind() != SyntaxKindFacetKeyword || operator.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("facet fixed tokens = %v/%v", operator.FacetKeyword().Kind(), operator.ByKeyword().Kind())
	}
	if operator.Expressions().Count() != 6 {
		t.Fatalf("facet key count = %d, want 6", operator.Expressions().Count())
	}
	wantKinds := []SyntaxKind{
		SyntaxKindPathExpression,
		SyntaxKindElementExpression,
		SyntaxKindNameReference,
		SyntaxKindNameReference,
		SyntaxKindPathExpression,
		SyntaxKindNameReference,
	}
	for i, want := range wantKinds {
		if got := operator.Expressions().At(i).Element().Kind(); got != want {
			t.Fatalf("facet key %d kind = %v, want %v", i, got, want)
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid facet diagnostics = %#v", diagnostics)
	}

	single := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account").(*FacetOperator)
	if single.Expressions().Count() != 1 || single.WithClause() != nil {
		t.Fatalf("single-key facet = keys:%d with:%#v", single.Expressions().Count(), single.WithClause())
	}
}

func TestQueryGrammarPublicFacetListRecoveryAndNoEndGuard(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingBy := parseQueryGrammarProduction(t, grammar.facetOperator, "facet Account").(*FacetOperator)
	if missingBy.ByKeyword().Width() != 0 ||
		!diagnosticsContainMessage(missingBy.ByKeyword().Diagnostics(), "KS005", "Expected: by") ||
		missingBy.Expressions().Count() != 1 {
		t.Fatalf("missing facet by recovery = %#v diagnostics:%#v", missingBy, missingBy.ContainedDiagnostics())
	}

	missingName := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by").(*FacetOperator)
	if missingName.Expressions().Count() != 1 || missingName.Expressions().At(0).Element().Width() != 0 ||
		!diagnosticsContainMessage(missingName.Expressions().At(0).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing facet name = %#v diagnostics:%#v", missingName.Expressions(), missingName.ContainedDiagnostics())
	}

	comma := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account,").(*FacetOperator)
	if comma.Expressions().Count() != 2 || comma.Expressions().At(1).Element().Width() != 0 {
		t.Fatalf("terminal comma facet list = %#v, want explicit missing second key", comma.Expressions())
	}
	if !diagnosticsContainMessage(comma.Expressions().At(1).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("terminal comma diagnostics = %#v", comma.ContainedDiagnostics())
	}

	// The source grammar intentionally supplies no list end guard. A legal simple-path
	// keyword after by or comma is therefore a key, not the start of a with clause.
	for _, test := range []struct {
		text       string
		wantLength int
		wantKeys   int
	}{
		{text: "facet by with count", wantLength: 3, wantKeys: 1},
		{text: "facet by Account, with count", wantLength: 5, wantKeys: 2},
	} {
		t.Run(test.text, func(t *testing.T) {
			tokens := Lex(test.text)
			result := grammar.facetOperator.ParseValue(NewArraySource(tokens), 0)
			operator := result.Value.(*FacetOperator)
			if result.Length != test.wantLength || result.Length == len(tokens)-1 ||
				operator.Expressions().Count() != test.wantKeys || operator.WithClause() != nil {
				t.Fatalf("no-end-guard facet parse = %#v keys:%d with:%#v", result, operator.Expressions().Count(), operator.WithClause())
			}
			last := operator.Expressions().At(operator.Expressions().Count() - 1).Element()
			if name, ok := last.(*NameReference); !ok || NameReferenceSimpleName(name) != "with" {
				t.Fatalf("no-end-guard final key = %#v, want keyword name with", last)
			}
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account 123", 3)
	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account, 123", 4)
}

func TestQueryGrammarPublicFacetWithClauseFormsAndPipelines(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	parenthesized := parseQueryGrammarProduction(t, grammar.facetOperator,
		"facet by Account with (where Active | extend Value=1 | count)").(*FacetOperator)
	withExpression, ok := parenthesized.WithClause().(*FacetWithExpressionClause)
	if !ok {
		t.Fatalf("parenthesized facet with = %T, want *FacetWithExpressionClause", parenthesized.WithClause())
	}
	if withExpression.OpenParen().Width() == 0 || withExpression.CloseParen().Width() == 0 {
		t.Fatalf("parenthesized facet delimiters = %#v/%#v", withExpression.OpenParen(), withExpression.CloseParen())
	}
	countPipe, count := queryGrammarPipeStage(t, withExpression.Expression(), SyntaxKindCountOperator)
	extendPipe, extend := queryGrammarPipeStage(t, countPipe.Expression(), SyntaxKindExtendOperator)
	if extendPipe.Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("facet fork pipeline source = %v, want FilterOperator", extendPipe.Expression().Kind())
	}
	if _, ok := extend.(*ExtendOperator); !ok {
		t.Fatalf("facet fork middle stage = %T, want *ExtendOperator", extend)
	}
	if _, ok := count.(*CountOperator); !ok {
		t.Fatalf("facet fork final stage = %T, want *CountOperator", count)
	}

	unparenthesized := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account with summarize Count=count()").(*FacetOperator)
	withOperator, ok := unparenthesized.WithClause().(*FacetWithOperatorClause)
	if !ok || withOperator.Operator().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("unparenthesized facet with = %#v, want summarize operator clause", unparenthesized.WithClause())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account with count | take 1", 5)
	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account with (Events | count)", 5)
}

func TestQueryGrammarPublicFacetWithClauseRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account with ()").(*FacetOperator)
	emptyClause := empty.WithClause().(*FacetWithExpressionClause)
	badExpression, ok := emptyClause.Expression().(*BadQueryOperator)
	if !ok || badExpression.Width() != 0 || !diagnosticsContainMessage(badExpression.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("missing facet expression = %#v diagnostics:%#v", emptyClause.Expression(), empty.ContainedDiagnostics())
	}

	missingOperator := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account with").(*FacetOperator)
	operatorClause := missingOperator.WithClause().(*FacetWithOperatorClause)
	badOperator, ok := operatorClause.Operator().(*BadQueryOperator)
	if !ok || badOperator.Width() != 0 || !diagnosticsContainMessage(badOperator.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("missing facet operator = %#v diagnostics:%#v", operatorClause.Operator(), missingOperator.ContainedDiagnostics())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account with (count").(*FacetOperator)
	closeClause := missingClose.WithClause().(*FacetWithExpressionClause)
	if closeClause.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(closeClause.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing facet close = %#v diagnostics:%#v", closeClause.CloseParen(), missingClose.ContainedDiagnostics())
	}

	dangling := parseQueryGrammarProduction(t, grammar.facetOperator, "facet by Account with (count |)").(*FacetOperator)
	danglingClause := dangling.WithClause().(*FacetWithExpressionClause)
	pipe, ok := danglingClause.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("facet internal dangling pipe = %T, want *PipeExpression", danglingClause.Expression())
	}
	badInternal, ok := pipe.Operator().(*BadQueryOperator)
	if !ok || !diagnosticsContainMessage(badInternal.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("facet internal missing operator = %#v diagnostics:%#v", pipe.Operator(), dangling.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicFacetTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account ;", 3)
	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account with (count), Extra", 7)
	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account with (count) Extra", 7)
	assertQueryGrammarProductionLeavesTail(t, grammar.facetOperator, "facet by Account with count Extra", 5)
}

func TestQueryGrammarPublicFacetCompletionVisibility(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, parser := range []Parser[LexicalToken]{grammar.facetOperator, grammar.QueryOperator, grammar.FollowingPipeElementExpression} {
		items := queryGrammarFacetCompletionItems(parser, "")
		if _, ok := completionItemByDisplayText(items, "facet"); ok {
			t.Fatalf("hidden facet completion exposed by parser tag %q: %#v", parser.Tag(), completionDisplayTexts(items))
		}
	}

	items := queryGrammarFacetCompletionItems(grammar.facetOperator, "facet ")
	if _, ok := completionItemByDisplayText(items, "by"); !ok {
		t.Fatalf("facet by completion missing: %#v", completionDisplayTexts(items))
	}

	items = queryGrammarFacetCompletionItems(grammar.facetOperator, "facet by Account with ")
	for _, display := range []string{"count", "extend", "project", "summarize", "where"} {
		item, ok := completionItemByDisplayText(items, display)
		if !ok || item.Kind != CompletionKindQueryPrefix {
			t.Fatalf("facet with completion %q = %#v, visible:%#v", display, item, completionDisplayTexts(items))
		}
	}
	for _, display := range []string{"fork", "join", "lookup", "search", "union"} {
		if _, ok := completionItemByDisplayText(items, display); ok {
			t.Fatalf("hidden fork fallback exposed facet completion %q: %#v", display, completionDisplayTexts(items))
		}
	}
}

func queryGrammarFacetCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
