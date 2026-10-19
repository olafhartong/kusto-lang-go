package kustolang

import "testing"

func TestQueryGrammarPublicForkNameEqualsAndExpressionProductions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	nameEquals := parseQueryGrammarProduction(t, grammar.nameEqualsClause, `["Named Branch"]=`)
	if NameDeclarationSimpleName(nameEquals.Name()) != "Named Branch" ||
		nameEquals.EqualToken().Kind() != SyntaxKindEqualToken || nameEquals.EqualToken().Width() == 0 {
		t.Fatalf("name-equals clause = %#v", nameEquals)
	}

	unnamed := parseQueryGrammarProduction(t, grammar.forkExpression, "(where Active)")
	if unnamed.NameEquals() != nil || unnamed.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		unnamed.Expression().Kind() != SyntaxKindFilterOperator || unnamed.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("unnamed fork expression = %#v", unnamed)
	}

	named := parseQueryGrammarProduction(t, grammar.forkExpression, "Branch=(take 1)")
	if named.NameEquals() == nil || NameDeclarationSimpleName(named.NameEquals().Name()) != "Branch" ||
		named.Expression().Kind() != SyntaxKindTakeOperator {
		t.Fatalf("named fork expression = %#v", named)
	}
}

func TestQueryGrammarPublicForkAdjacentBranchesAndPipelines(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parsed := parseQueryGrammarProduction(t, grammar.forkOperator,
		`fork (where Active)ByValue=(take 5 | project Value)["Totals"]=(summarize Count=count())`)
	operator, ok := any(parsed).(*ForkOperator)
	if !ok {
		t.Fatalf("fork operator = %T, want *ForkOperator", parsed)
	}
	if operator.Expressions().Count() != 3 {
		t.Fatalf("adjacent fork branch count = %d, want 3", operator.Expressions().Count())
	}

	first := operator.Expressions().At(0)
	if first.NameEquals() != nil || first.Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("first fork branch = %#v", first)
	}
	second := operator.Expressions().At(1)
	if second.NameEquals() == nil || NameDeclarationSimpleName(second.NameEquals().Name()) != "ByValue" {
		t.Fatalf("second fork branch name = %#v", second.NameEquals())
	}
	projectPipe, project := queryGrammarPipeStage(t, second.Expression(), SyntaxKindProjectOperator)
	if projectPipe.Expression().Kind() != SyntaxKindTakeOperator {
		t.Fatalf("second fork pipeline source = %#v, want TakeOperator", projectPipe.Expression())
	}
	if _, ok := project.(*ProjectOperator); !ok {
		t.Fatalf("second fork pipeline operator = %T, want *ProjectOperator", project)
	}
	third := operator.Expressions().At(2)
	if third.NameEquals() == nil || NameDeclarationSimpleName(third.NameEquals().Name()) != "Totals" ||
		third.Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("third fork branch = %#v", third)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid fork diagnostics = %#v", diagnostics)
	}

	pipeline := parseQueryGrammarProduction(t, grammar.forkPipeExpression, "where Active | extend Value=1 | count")
	countPipe, count := queryGrammarPipeStage(t, pipeline, SyntaxKindCountOperator)
	extendPipe, extend := queryGrammarPipeStage(t, countPipe.Expression(), SyntaxKindExtendOperator)
	if extendPipe.Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("fork pipeline first stage = %#v, want FilterOperator", extendPipe.Expression())
	}
	if _, ok := count.(*CountOperator); !ok {
		t.Fatalf("fork pipeline final operator = %T, want *CountOperator", count)
	}
	if _, ok := extend.(*ExtendOperator); !ok {
		t.Fatalf("fork pipeline middle operator = %T, want *ExtendOperator", extend)
	}
}

func TestQueryGrammarPublicForkHiddenFallbacks(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	fallback := parseQueryGrammarProduction(t, grammar.forkPipeOperator, "join Right on Key")
	if _, ok := fallback.(*JoinOperator); !ok {
		t.Fatalf("fork query-operator fallback = %T, want *JoinOperator", fallback)
	}

	branch := parseQueryGrammarProduction(t, grammar.forkExpression, "(Events | where Active)")
	pipe, ok := branch.Expression().(*PipeExpression)
	if !ok || pipe.Operator().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("fork expression fallback = %#v, want ordinary pipe expression", branch.Expression())
	}
	if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("fork expression fallback source = %#v, want Events", pipe.Expression())
	}

	operatorBranch := parseQueryGrammarProduction(t, grammar.forkExpression, "(join Right on Key)")
	if operatorBranch.Expression().Kind() != SyntaxKindJoinOperator {
		t.Fatalf("fork branch hidden query-operator fallback kind = %v, want JoinOperator", operatorBranch.Expression().Kind())
	}
}

func TestQueryGrammarPublicForkRecoveryAndCommaBoundary(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingEqual := parseQueryGrammarProduction(t, grammar.nameEqualsClause, "Branch")
	if missingEqual.EqualToken().Width() != 0 ||
		!diagnosticsContainMessage(missingEqual.EqualToken().Diagnostics(), "KS005", "Expected: =") {
		t.Fatalf("missing fork branch equals = %#v diagnostics:%#v", missingEqual.EqualToken(), missingEqual.ContainedDiagnostics())
	}
	recoveredName := parseQueryGrammarProduction(t, grammar.forkExpression, "Branch(where Active)")
	if recoveredName.NameEquals() == nil || recoveredName.NameEquals().EqualToken().Width() != 0 ||
		recoveredName.Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("fork expression missing-equals recovery = %#v", recoveredName)
	}

	assertQueryGrammarProductionRejects(t, grammar.forkExpression, "Branch=where Active)")
	missingBranchParsed := parseQueryGrammarProduction(t, grammar.forkOperator, "fork")
	missingBranch := any(missingBranchParsed).(*ForkOperator).Expressions().At(0)
	if missingBranch.OpenParen().Width() != 0 || missingBranch.Expression().Width() != 0 || missingBranch.CloseParen().Width() != 0 {
		t.Fatalf("missing fork branch = %#v", missingBranch)
	}
	if !diagnosticsContainMessage(missingBranch.OpenParen().Diagnostics(), "KS005", "Expected: (") ||
		!diagnosticsContainMessage(missingBranch.Expression().Diagnostics(), "KS006", "Missing expression") ||
		!diagnosticsContainMessage(missingBranch.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing fork branch diagnostics = %#v", missingBranch.ContainedDiagnostics())
	}

	missingExpression := parseQueryGrammarProduction(t, grammar.forkExpression, "()")
	if missingExpression.Expression().Width() != 0 ||
		!diagnosticsContainMessage(missingExpression.Expression().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing fork expression = %#v diagnostics:%#v", missingExpression.Expression(), missingExpression.ContainedDiagnostics())
	}
	missingClose := parseQueryGrammarProduction(t, grammar.forkExpression, "(count")
	if missingClose.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing fork close = %#v diagnostics:%#v", missingClose.CloseParen(), missingClose.ContainedDiagnostics())
	}

	dangling := parseQueryGrammarProduction(t, grammar.forkPipeExpression, "count |")
	danglingPipe, ok := dangling.(*PipeExpression)
	if !ok {
		t.Fatalf("dangling fork pipe = %T, want *PipeExpression", dangling)
	}
	bad, ok := danglingPipe.Operator().(*BadQueryOperator)
	if !ok || !diagnosticsContainMessage(bad.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("dangling fork pipe operator = %#v diagnostics:%#v", danglingPipe.Operator(), dangling.ContainedDiagnostics())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.forkOperator, "fork (count), (take 1)", 4)
}

func TestQueryGrammarPublicForkPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			parsed := parseQueryGrammarProduction(t, production.parser, "fork (count)")
			if _, ok := parsed.(*ForkOperator); !ok {
				t.Fatalf("fork placement = %T, want *ForkOperator", parsed)
			}
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.Expression, "fork (count)", 1)
	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "Events | fork (where Active) Named=(count)")
	forkPipe, forkNode := queryGrammarPipeStage(t, pipeline, SyntaxKindForkOperator)
	if source, ok := forkPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("fork pipeline source = %#v, want Events", forkPipe.Expression())
	}
	if forkNode.(*ForkOperator).Expressions().Count() != 2 {
		t.Fatalf("fork pipeline branch count = %d, want 2", forkNode.(*ForkOperator).Expressions().Count())
	}
}

func TestQueryGrammarPublicForkTagsAndCompletionVisibility(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name   string
		parser Parser[LexicalToken]
		want   string
	}{
		{name: "name equals", parser: grammar.nameEqualsClause, want: ""},
		{name: "fork expression", parser: grammar.forkExpression, want: ""},
		{name: "fork operator", parser: grammar.forkOperator, want: "<fork>"},
		{name: "fork pipe operator", parser: grammar.forkPipeOperator, want: "<fork-pipe-operator>"},
		{name: "fork pipe expression", parser: grammar.forkPipeExpression, want: "<fork-pipe-expression>"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.parser.Tag(); got != test.want {
				t.Fatalf("parser tag = %q, want %q", got, test.want)
			}
		})
	}

	for _, parser := range []Parser[LexicalToken]{grammar.forkOperator, grammar.QueryOperator, grammar.FollowingPipeElementExpression} {
		items := queryGrammarForkCompletionItems(parser, "")
		if _, ok := completionItemByDisplayText(items, "fork"); ok {
			t.Fatalf("hidden fork completion exposed by parser tag %q: %#v", parser.Tag(), completionDisplayTexts(items))
		}
	}

	items := queryGrammarForkCompletionItems(grammar.forkPipeOperator, "")
	for _, display := range []string{"where", "count", "extend", "project", "summarize"} {
		item, ok := completionItemByDisplayText(items, display)
		if !ok || item.Kind != CompletionKindQueryPrefix {
			t.Fatalf("legal fork branch completion %q = %#v, visible:%#v", display, item, completionDisplayTexts(items))
		}
	}
	for _, display := range []string{"fork", "join", "lookup", "search", "union"} {
		if _, ok := completionItemByDisplayText(items, display); ok {
			t.Fatalf("hidden fork fallback exposed completion %q: %#v", display, completionDisplayTexts(items))
		}
	}
}

func queryGrammarForkCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
