package kustolang

import "testing"

func TestQueryGrammarPublicSummarizeCoreForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.QueryOperator, "summarize").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, empty, 0, 0, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, empty)

	aggregate := parseQueryGrammarProduction(t, grammar.QueryOperator, "summarize count()").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, aggregate, 1, 0, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, aggregate)
	if aggregate.Aggregates().At(0).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("aggregate = %#v, want function call", aggregate.Aggregates().At(0).Element())
	}

	byOnly := parseQueryGrammarProduction(t, grammar.QueryOperator, "summarize by State").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, byOnly, 0, 1, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, byOnly)
	assertQueryGrammarSummarizeNameReference(t, byOnly.ByClause().Expressions().At(0).Element(), "State")

	aggregateAndBy := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize Total=count() by State").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, aggregateAndBy, 1, 1, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, aggregateAndBy)
	assertQueryGrammarSummarizeSimpleNamed(t, aggregateAndBy.Aggregates().At(0).Element(), "Total", SyntaxKindFunctionCallExpression)
	assertQueryGrammarSummarizeNameReference(t, aggregateAndBy.ByClause().Expressions().At(0).Element(), "State")
}

func TestQueryGrammarPublicSummarizeNamedCompoundAndMultipleLists(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"summarize Total=count(), (Minimum, Maximum)=series_stats(Value), sum(Cost) by Region, Normalized=tolower(Category), (First, Last)=split_name(User)").(*SummarizeOperator)

	assertQueryGrammarSummarizeShape(t, operator, 3, 3, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, operator)
	assertQueryGrammarSummarizeSimpleNamed(t, operator.Aggregates().At(0).Element(), "Total", SyntaxKindFunctionCallExpression)
	assertQueryGrammarSummarizeCompoundNamed(t, operator.Aggregates().At(1).Element(), 2, SyntaxKindFunctionCallExpression)
	if operator.Aggregates().At(2).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("third aggregate = %#v, want function call", operator.Aggregates().At(2).Element())
	}
	assertQueryGrammarSummarizeNameReference(t, operator.ByClause().Expressions().At(0).Element(), "Region")
	assertQueryGrammarSummarizeSimpleNamed(t, operator.ByClause().Expressions().At(1).Element(), "Normalized", SyntaxKindFunctionCallExpression)
	assertQueryGrammarSummarizeCompoundNamed(t, operator.ByClause().Expressions().At(2).Element(), 2, SyntaxKindFunctionCallExpression)

	for i := 0; i < operator.Aggregates().Count()-1; i++ {
		if separator := operator.Aggregates().At(i).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken || separator.Width() == 0 {
			t.Fatalf("aggregate separator %d = %#v, want present comma", i, separator)
		}
	}
	for i := 0; i < operator.ByClause().Expressions().Count()-1; i++ {
		if separator := operator.ByClause().Expressions().At(i).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken || separator.Width() == 0 {
			t.Fatalf("by separator %d = %#v, want present comma", i, separator)
		}
	}
}

func TestQueryGrammarPublicSummarizeParameters(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"summarize hint.shufflekey=Tenant hint.strategy=shuffle hint.num_partitions=8 count() by Region").(*SummarizeOperator)

	assertQueryGrammarSummarizeShape(t, operator, 1, 1, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, operator)
	assertQueryGrammarSharedParameters(t, operator.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.shufflekey", kind: SyntaxKindNameReference},
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.num_partitions", kind: SyntaxKindLongLiteralExpression},
	})

	// Summarize uses equals-needed lookahead, so this is not a parameter.
	assertQueryGrammarProductionLeavesTail(t, QueryGrammarFrom(nil).QueryOperator,
		"summarize hint.strategy shuffle count()", 1)
}

func TestQueryGrammarPublicSummarizeLegacyBinClauseAndLiteralGuard(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	legacy := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize count() by Timestamp bin=1h").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, legacy, 1, 1, true)
	assertQueryGrammarSummarizeNoDiagnostics(t, legacy)
	bin := legacy.ByClause().BinClause().(*SimpleNamedExpression)
	if NameDeclarationSimpleName(bin.Name()) != "bin" || bin.EqualToken().Kind() != SyntaxKindEqualToken ||
		bin.EqualToken().Width() == 0 || bin.Expression().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("legacy bin clause = %#v", bin)
	}

	// The pinned grammar only reserves bin=<literal> for the legacy clause.
	// Non-literal values remain regular named expressions in the by list.
	guarded := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize by Region, bin=Bucket").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, guarded, 0, 2, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, guarded)
	assertQueryGrammarSummarizeSimpleNamed(t, guarded.ByClause().Expressions().At(1).Element(), "bin", SyntaxKindNameReference)

	call := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize by bin=make_value(1h)").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, call, 0, 1, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, call)
	assertQueryGrammarSummarizeSimpleNamed(t, call.ByClause().Expressions().At(0).Element(), "bin", SyntaxKindFunctionCallExpression)
}

func TestQueryGrammarPublicSummarizeRootAndPipelinePlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	// Summarize is a post-pipe operator. At an expression root, the keyword is
	// accepted only as the initial name reference and the operator tail remains.
	assertQueryGrammarProductionLeavesTail(t, grammar.Expression, "summarize count() by Region", 1)

	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"Events | where Active | summarize Total=count() by Region | take 5")
	takePipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindTakeOperator)
	summarizePipe, summarizeNode := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindSummarizeOperator)
	wherePipe, _ := queryGrammarPipeStage(t, summarizePipe.Expression(), SyntaxKindFilterOperator)
	if source, ok := wherePipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("summarize pipeline source = %#v, want Events", wherePipe.Expression())
	}
	assertQueryGrammarSummarizeShape(t, summarizeNode.(*SummarizeOperator), 1, 1, false)
	assertQueryGrammarSummarizeNoDiagnostics(t, summarizeNode.(*SummarizeOperator))

	following := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression,
		"summarize count() by Region")
	if following.Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("following-pipe summarize kind = %v", following.Kind())
	}
}

func TestQueryGrammarPublicSummarizeRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingAggregate := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize Total=").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, missingAggregate, 1, 0, false)
	missingAggregateValue := missingAggregate.Aggregates().At(0).Element().(*SimpleNamedExpression).Expression()
	assertQueryGrammarSummarizeMissingExpression(t, missingAggregateValue)

	missingBy := parseQueryGrammarProduction(t, grammar.QueryOperator, "summarize by").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, missingBy, 0, 1, false)
	assertQueryGrammarSummarizeMissingExpression(t, missingBy.ByClause().Expressions().At(0).Element())
	if !diagnosticsContainMessage(missingBy.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing summarize by diagnostics = %#v", missingBy.ContainedDiagnostics())
	}

	missingAggregateComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize count() sum(Value)").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, missingAggregateComma, 2, 0, false)
	assertQueryGrammarSummarizeMissingComma(t, missingAggregateComma.Aggregates().At(0).Separator())

	missingByComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize by Region Category").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, missingByComma, 0, 2, false)
	assertQueryGrammarSummarizeMissingComma(t, missingByComma.ByClause().Expressions().At(0).Separator())

	// Without a literal after '=', the pinned lookahead does not select the
	// legacy bin clause; NamedExpression performs its usual RHS recovery.
	missingBinValue := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize by Region, bin=").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, missingBinValue, 0, 2, false)
	binExpression := missingBinValue.ByClause().Expressions().At(1).Element().(*SimpleNamedExpression)
	if NameDeclarationSimpleName(binExpression.Name()) != "bin" {
		t.Fatalf("missing bin expression name = %q, want bin", NameDeclarationSimpleName(binExpression.Name()))
	}
	assertQueryGrammarSummarizeMissingExpression(t, binExpression.Expression())

	trailingAggregateComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"summarize count(), by Region").(*SummarizeOperator)
	assertQueryGrammarSummarizeShape(t, trailingAggregateComma, 2, 1, false)
	assertQueryGrammarSummarizeMissingExpression(t, trailingAggregateComma.Aggregates().At(1).Element())
}

func TestQueryGrammarPublicSummarizeMalformedTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"summarize count() by Region by Category", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"summarize count() by Region bin=1h Extra", 9)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"summarize count() by Region | where Active", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"summarize count() by Region bin=1h, Extra", 9)
}

func assertQueryGrammarSummarizeShape(t *testing.T, operator *SummarizeOperator, aggregates, byExpressions int, hasBin bool) {
	t.Helper()
	if operator.SummarizeKeyword().Kind() != SyntaxKindSummarizeKeyword {
		t.Fatalf("summarize keyword kind = %v", operator.SummarizeKeyword().Kind())
	}
	if operator.Aggregates().Count() != aggregates {
		t.Fatalf("summarize aggregates = %d, want %d", operator.Aggregates().Count(), aggregates)
	}
	if byExpressions == 0 && operator.ByClause() == nil {
		if hasBin {
			t.Fatal("summarize has no by clause for expected bin clause")
		}
	} else {
		if operator.ByClause() == nil {
			t.Fatalf("summarize by clause = nil, want %d expressions", byExpressions)
		}
		if operator.ByClause().ByKeyword().Kind() != SyntaxKindByKeyword || operator.ByClause().Expressions().Count() != byExpressions {
			t.Fatalf("summarize by clause = %#v, want %d expressions", operator.ByClause(), byExpressions)
		}
	}
	if got := operator.ByClause() != nil && operator.ByClause().BinClause() != nil; got != hasBin {
		t.Fatalf("summarize has bin clause = %t, want %t", got, hasBin)
	}
}

func assertQueryGrammarSummarizeNoDiagnostics(t *testing.T, operator *SummarizeOperator) {
	t.Helper()
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("summarize diagnostics = %#v", diagnostics)
	}
}

func assertQueryGrammarSummarizeNameReference(t *testing.T, expression Expression, want string) {
	t.Helper()
	reference, ok := expression.(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != want {
		t.Fatalf("summarize expression = %#v, want name reference %q", expression, want)
	}
}

func assertQueryGrammarSummarizeSimpleNamed(t *testing.T, expression Expression, name string, valueKind SyntaxKind) {
	t.Helper()
	named, ok := expression.(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(named.Name()) != name || named.EqualToken().Kind() != SyntaxKindEqualToken ||
		named.EqualToken().Width() == 0 || named.Expression().Kind() != valueKind {
		t.Fatalf("summarize named expression = %#v, want %s=%v", expression, name, valueKind)
	}
}

func assertQueryGrammarSummarizeCompoundNamed(t *testing.T, expression Expression, names int, valueKind SyntaxKind) {
	t.Helper()
	compound, ok := expression.(*CompoundNamedExpression)
	if !ok || compound.Names().Names().Count() != names || compound.EqualToken().Kind() != SyntaxKindEqualToken ||
		compound.EqualToken().Width() == 0 || compound.Expression().Kind() != valueKind {
		t.Fatalf("summarize compound expression = %#v, want %d names and %v", expression, names, valueKind)
	}
}

func assertQueryGrammarSummarizeMissingExpression(t *testing.T, expression Expression) {
	t.Helper()
	if expression == nil {
		t.Fatal("missing summarize expression = nil")
	}
	if expression.Width() != 0 || !diagnosticsContain(expression.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing summarize expression = %#v diagnostics:%#v", expression, expression.ContainedDiagnostics())
	}
}

func assertQueryGrammarSummarizeMissingComma(t *testing.T, comma *SyntaxToken) {
	t.Helper()
	if comma == nil || comma.Kind() != SyntaxKindCommaToken || comma.Width() != 0 ||
		!diagnosticsContainMessage(comma.Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing summarize comma = %#v diagnostics:%#v", comma, comma.Diagnostics())
	}
}
