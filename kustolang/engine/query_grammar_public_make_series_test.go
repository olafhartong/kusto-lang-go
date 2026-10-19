package kustolang

import "testing"

func TestQueryGrammarPublicMakeSeriesDirectFullFromToStepAST(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).makeSeriesOperator,
		"make-series hint.shufflekey=Tenant kind=nonempty Total=count() default=0, Average=avg(Value) on Timestamp from ago(1d) to now() step 1h by Region, Group=tostring(Code)").(*MakeSeriesOperator)

	if operator.Kind() != SyntaxKindMakeSeriesOperator || operator.MakeSeriesKeyword().Kind() != SyntaxKindMakeSeriesKeyword {
		t.Fatalf("make-series operator = kind:%v keyword:%v", operator.Kind(), operator.MakeSeriesKeyword().Kind())
	}
	if operator.Parameters().Count() != 2 {
		t.Fatalf("make-series parameters = %d, want 2", operator.Parameters().Count())
	}
	for i, want := range []struct {
		name string
		kind SyntaxKind
	}{
		{name: "hint.shufflekey", kind: SyntaxKindNameReference},
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
	} {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want.name || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != want.kind {
			t.Fatalf("parameter %d = %#v, want %q/%v", i, parameter, want.name, want.kind)
		}
	}

	if operator.Aggregates().Count() != 2 {
		t.Fatalf("make-series aggregates = %d, want 2", operator.Aggregates().Count())
	}
	first := operator.Aggregates().At(0).Element()
	if first.Kind() != SyntaxKindMakeSeriesExpression || first.Expression().Kind() != SyntaxKindSimpleNamedExpression ||
		first.DefaultExpression() == nil || first.DefaultExpression().Kind() != SyntaxKindDefaultExpressionClause ||
		first.DefaultExpression().DefaultKeyword().Kind() != SyntaxKindDefaultKeyword ||
		first.DefaultExpression().EqualToken().Width() == 0 ||
		first.DefaultExpression().Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("first aggregate/default AST = %#v", first)
	}
	second := operator.Aggregates().At(1).Element()
	if second.Kind() != SyntaxKindMakeSeriesExpression || second.Expression().Kind() != SyntaxKindSimpleNamedExpression ||
		second.DefaultExpression() != nil {
		t.Fatalf("second aggregate AST = %#v", second)
	}
	if separator := operator.Aggregates().At(0).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken || separator.Width() == 0 {
		t.Fatalf("aggregate separator = %#v", separator)
	}

	on := operator.OnClause()
	if on.Kind() != SyntaxKindMakeSeriesOnClause || on.OnKeyword().Kind() != SyntaxKindOnKeyword ||
		on.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("make-series on clause = %#v", on)
	}
	rangeClause, ok := operator.RangeClause().(*MakeSeriesFromToStepClause)
	if !ok || rangeClause.Kind() != SyntaxKindMakeSeriesFromToStepClause {
		t.Fatalf("make-series range = %#v, want *MakeSeriesFromToStepClause", operator.RangeClause())
	}
	if rangeClause.MakeSeriesFromClause().Kind() != SyntaxKindMakeSeriesFromClause ||
		rangeClause.MakeSeriesFromClause().Expression().Kind() != SyntaxKindFunctionCallExpression ||
		rangeClause.MakeSeriesToClause().Kind() != SyntaxKindMakeSeriesToClause ||
		rangeClause.MakeSeriesToClause().Expression().Kind() != SyntaxKindFunctionCallExpression ||
		rangeClause.MakeSeriesStepClause().Kind() != SyntaxKindMakeSeriesStepClause ||
		rangeClause.MakeSeriesStepClause().Expression().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("make-series from/to/step AST = %#v", rangeClause)
	}
	by := operator.ByClause()
	if by == nil || by.Kind() != SyntaxKindMakeSeriesByClause || by.ByKeyword().Kind() != SyntaxKindByKeyword ||
		by.Expressions().Count() != 2 || by.Expressions().At(0).Element().Kind() != SyntaxKindNameReference ||
		by.Expressions().At(1).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("make-series by clause = %#v", by)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid make-series diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMakeSeriesDeprecatedInRangeAST(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).makeSeriesOperator,
		"make-series count() on Timestamp in range(ago(1d), now(), 1h)").(*MakeSeriesOperator)
	rangeClause, ok := operator.RangeClause().(*MakeSeriesInRangeClause)
	if !ok || rangeClause.Kind() != SyntaxKindMakeSeriesInRangeClause ||
		rangeClause.InKeyword().Kind() != SyntaxKindInKeyword || rangeClause.RangeKeyword().Kind() != SyntaxKindRangeKeyword {
		t.Fatalf("deprecated make-series range = %#v", operator.RangeClause())
	}
	arguments := rangeClause.Arguments()
	if arguments.Kind() != SyntaxKindExpressionList || arguments.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		arguments.CloseParen().Kind() != SyntaxKindCloseParenToken || arguments.Expressions().Count() != 3 {
		t.Fatalf("deprecated make-series arguments = %#v", arguments)
	}
	wantKinds := []SyntaxKind{SyntaxKindFunctionCallExpression, SyntaxKindFunctionCallExpression, SyntaxKindTimespanLiteralExpression}
	for i, want := range wantKinds {
		if got := arguments.Expressions().At(i).Element().Kind(); got != want {
			t.Fatalf("range argument %d kind = %v, want %v", i, got, want)
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("deprecated range diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMakeSeriesProductionsAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := "make-series count() on Timestamp step 1h"
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private":        grammar.makeSeriesOperator,
		"query operator": grammar.QueryOperator,
		"following pipe": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, parser, text)
			if operator.Kind() != SyntaxKindMakeSeriesOperator {
				t.Fatalf("%s kind = %v, want %v", name, operator.Kind(), SyntaxKindMakeSeriesOperator)
			}
		})
	}

	root := parseQueryGrammarProduction(t, grammar.Expression, text)
	if root.Kind() != SyntaxKindMakeSeriesOperator {
		t.Fatalf("expression-root make-series kind = %v, want %v", root.Kind(), SyntaxKindMakeSeriesOperator)
	}
	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"Events | make-series count() on Timestamp step 1h by Region | where Region != ''")
	wherePipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindFilterOperator)
	seriesPipe, series := queryGrammarPipeStage(t, wherePipe.Expression(), SyntaxKindMakeSeriesOperator)
	if source, ok := seriesPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("make-series pipeline source = %#v, want Events", seriesPipe.Expression())
	}
	if _, ok := series.(*MakeSeriesOperator); !ok {
		t.Fatalf("make-series pipeline operator = %T", series)
	}
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("make-series pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMakeSeriesCompletionAndTag(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	items := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.QueryOperator, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	item, ok := completionItemByDisplayText(items, "make-series")
	if !ok || item.Kind != CompletionKindQueryPrefix {
		t.Fatalf("make-series completion = %#v, visible items:%#v", item, completionDisplayTexts(items))
	}
	if grammar.makeSeriesOperator.Tag() != "<make-series>" || !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<make-series>", 0) {
		t.Fatalf("make-series tags = private:%q query-tree:%v", grammar.makeSeriesOperator.Tag(),
			queryGrammarParserTreeHasTag(grammar.QueryOperator, "<make-series>", 0))
	}
}

func TestQueryGrammarPublicMakeSeriesRequiredRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingAggregate := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series on Timestamp step 1h").(*MakeSeriesOperator)
	if missingAggregate.Aggregates().Count() != 1 {
		t.Fatalf("missing aggregate count = %d, want 1", missingAggregate.Aggregates().Count())
	}
	assertQueryGrammarMakeSeriesMissingExpression(t, missingAggregate.Aggregates().At(0).Element().Expression())

	missingOn := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count()").(*MakeSeriesOperator)
	if missingOn.OnClause().OnKeyword().Width() != 0 ||
		!diagnosticsContainMessage(missingOn.OnClause().OnKeyword().Diagnostics(), "KS005", "Expected: on") {
		t.Fatalf("missing on recovery = %#v diagnostics:%#v", missingOn.OnClause(), missingOn.OnClause().ContainedDiagnostics())
	}

	missingAxis := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on step 1h").(*MakeSeriesOperator)
	assertQueryGrammarMakeSeriesMissingExpression(t, missingAxis.OnClause().Expression())
	if _, ok := missingAxis.RangeClause().(*MakeSeriesFromToStepClause); !ok {
		t.Fatalf("missing axis range = %#v", missingAxis.RangeClause())
	}

	missingStepArgument := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp step").(*MakeSeriesOperator)
	stepRange, ok := missingStepArgument.RangeClause().(*MakeSeriesFromToStepClause)
	if !ok {
		t.Fatalf("missing step argument range = %#v", missingStepArgument.RangeClause())
	}
	assertQueryGrammarMakeSeriesMissingExpression(t, stepRange.MakeSeriesStepClause().Expression())

	missingRange := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp").(*MakeSeriesOperator)
	inRange, ok := missingRange.RangeClause().(*MakeSeriesInRangeClause)
	if !ok {
		t.Fatalf("missing range recovery = %#v", missingRange.RangeClause())
	}
	for name, token := range map[string]*SyntaxToken{
		"in":    inRange.InKeyword(),
		"range": inRange.RangeKeyword(),
		"open":  inRange.Arguments().OpenParen(),
		"close": inRange.Arguments().CloseParen(),
	} {
		if token.Width() != 0 || !diagnosticsContain(token.Diagnostics(), "KS005") {
			t.Fatalf("missing range %s token = %#v diagnostics:%#v", name, token, token.Diagnostics())
		}
	}
}

func TestQueryGrammarPublicMakeSeriesRangeListsAndTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingArguments := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp in range(,now(),)").(*MakeSeriesOperator).
		RangeClause().(*MakeSeriesInRangeClause).Arguments().Expressions()
	if missingArguments.Count() != 3 {
		t.Fatalf("range missing argument count = %d, want 3", missingArguments.Count())
	}
	assertQueryGrammarMakeSeriesMissingExpression(t, missingArguments.At(0).Element())
	assertQueryGrammarMakeSeriesMissingExpression(t, missingArguments.At(2).Element())

	missingCommas := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp in range(ago(1d) now() 1h)").(*MakeSeriesOperator).
		RangeClause().(*MakeSeriesInRangeClause).Arguments().Expressions()
	if missingCommas.Count() != 3 {
		t.Fatalf("range missing comma argument count = %d, want 3", missingCommas.Count())
	}
	for i := 0; i < 2; i++ {
		separator := missingCommas.At(i).Separator()
		if separator == nil || separator.Width() != 0 ||
			!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ,") {
			t.Fatalf("range missing separator %d = %#v diagnostics:%#v", i, separator, separator.Diagnostics())
		}
	}

	missingClose := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp in range(ago(1d), now(), 1h").(*MakeSeriesOperator).
		RangeClause().(*MakeSeriesInRangeClause).Arguments().CloseParen()
	if missingClose.Width() != 0 || !diagnosticsContainMessage(missingClose.Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("range missing close = %#v diagnostics:%#v", missingClose, missingClose.Diagnostics())
	}

	missingBy := parseQueryGrammarProduction(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp step 1h by").(*MakeSeriesOperator).ByClause()
	if missingBy.Expressions().Count() != 1 {
		t.Fatalf("missing by expression count = %d, want 1", missingBy.Expressions().Count())
	}
	assertQueryGrammarMakeSeriesMissingExpression(t, missingBy.Expressions().At(0).Element())

	assertQueryGrammarProductionLeavesTail(t, grammar.makeSeriesOperator,
		"make-series count() on Timestamp step 1h | where true", 8)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"make-series count() on Timestamp in range(0, 10, 1) Extra", 15)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"make-series count() on Timestamp step 1h, Extra", 8)
}

func assertQueryGrammarMakeSeriesMissingExpression(t *testing.T, expression Expression) {
	t.Helper()
	if expression == nil {
		t.Fatal("missing make-series expression = nil")
	}
	if expression.Width() != 0 || !diagnosticsContain(expression.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing make-series expression = %#v diagnostics:%#v", expression, expression.ContainedDiagnostics())
	}
}
