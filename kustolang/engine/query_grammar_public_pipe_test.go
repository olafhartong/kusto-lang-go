package kustolang

import "testing"

func TestQueryGrammarPublicExpressionRootOperators(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	printed := parseQueryGrammarProduction(t, grammar.Expression, "print total=1 + 2, label='ok'")
	printOperator, ok := printed.(*PrintOperator)
	if !ok {
		t.Fatalf("print root = %T, want *PrintOperator", printed)
	}
	if printOperator.Expressions().Count() != 2 {
		t.Fatalf("print expressions = %d, want 2", printOperator.Expressions().Count())
	}
	total, ok := printOperator.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(total.Name()) != "total" || total.Expression().Kind() != SyntaxKindAddExpression {
		t.Fatalf("print first expression = %#v, want named add expression", printOperator.Expressions().At(0).Element())
	}

	ranged := parseQueryGrammarProduction(t, grammar.Expression, "range n from 1 to 5 step 2")
	rangeOperator, ok := ranged.(*RangeOperator)
	if !ok {
		t.Fatalf("range root = %T, want *RangeOperator", ranged)
	}
	if NameDeclarationSimpleName(rangeOperator.Name()) != "n" ||
		rangeOperator.From().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.To().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.Step().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range operator = %#v, want n with literal bounds and step", rangeOperator)
	}
	if diagnostics := rangeOperator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("range diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicPipeProductions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	pipe := parseQueryGrammarProduction(t, grammar.PipeExpression, "Events | where Severity > 2")
	wherePipe, ok := pipe.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %T, want *PipeExpression", pipe)
	}
	if source, ok := wherePipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("pipe source = %#v, want Events", wherePipe.Expression())
	}
	filter, ok := wherePipe.Operator().(*FilterOperator)
	if !ok || filter.Condition().Kind() != SyntaxKindGreaterThanExpression {
		t.Fatalf("pipe operator = %#v, want where with greater-than condition", wherePipe.Operator())
	}

	sub := parseQueryGrammarProduction(t, grammar.PipeSubExpression, "Events | count")
	subPipe, ok := sub.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe subexpression = %T, want *PipeExpression", sub)
	}
	if _, ok := subPipe.Operator().(*CountOperator); !ok {
		t.Fatalf("pipe subexpression operator = %T, want *CountOperator", subPipe.Operator())
	}

	following := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, "where Enabled")
	if filter, ok := following.(*FilterOperator); !ok || filter.Condition().Kind() != SyntaxKindNameReference {
		t.Fatalf("following pipe element = %#v, want where operator", following)
	}
}

func TestQueryGrammarPublicQueryOperatorSubset(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name string
		text string
		want SyntaxKind
	}{
		{name: "where", text: "where Value > 0", want: SyntaxKindFilterOperator},
		{name: "project", text: "project Value, Doubled=Value * 2", want: SyntaxKindProjectOperator},
		{name: "extend", text: "extend Added=Value + 1", want: SyntaxKindExtendOperator},
		{name: "take", text: "take 5", want: SyntaxKindTakeOperator},
		{name: "limit", text: "limit 6", want: SyntaxKindTakeOperator},
		{name: "count", text: "count", want: SyntaxKindCountOperator},
		{name: "distinct", text: "distinct Value, Category", want: SyntaxKindDistinctOperator},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if operator.Kind() != test.want {
				t.Fatalf("query operator %q kind = %v, want %v", test.text, operator.Kind(), test.want)
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("query operator %q diagnostics = %#v", test.text, diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicCountAndDistinctParameters(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	count := parseQueryGrammarProduction(t, grammar.QueryOperator, "count as Total").(*CountOperator)
	if count.AsIdentifier() == nil || count.AsIdentifier().Identifier().RawText() != "Total" {
		t.Fatalf("count as clause = %#v", count.AsIdentifier())
	}
	distinct := parseQueryGrammarProduction(t, grammar.QueryOperator, "distinct hint.strategy=shuffle hint.num_partitions=4 hint.shufflekey=Tenant Tenant, Account").(*DistinctOperator)
	if distinct.Parameters().Count() != 3 || distinct.Expressions().Count() != 2 || len(distinct.ContainedDiagnostics()) != 0 {
		t.Fatalf("distinct parameters/expressions = %d/%d diagnostics:%#v", distinct.Parameters().Count(), distinct.Expressions().Count(), distinct.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicRangeFunctionLookaheadAndRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "range(1, 2, 1)")
	missing := parseQueryGrammarProduction(t, grammar.QueryOperator, "range n from to 5 step").(*RangeOperator)
	if !diagnosticsContainMessage(missing.From().Diagnostics(), "KS006", "Missing expression: If the keyword 'to' is intended to be part of an expression it needs to be bracketted as ['to'].") ||
		!diagnosticsContainMessage(missing.Step().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("range recovery diagnostics = %#v", missing.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicMultiStagePipeAndScalarPrecedence(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"print x=1 | extend y=x + 2 * 3 | where y > 5 and x == 1 or y == 7 | project y | take 10")

	takePipe, take := queryGrammarPipeStage(t, expression, SyntaxKindTakeOperator)
	projectPipe, project := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindProjectOperator)
	filterPipe, filterNode := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	extendPipe, extendNode := queryGrammarPipeStage(t, filterPipe.Expression(), SyntaxKindExtendOperator)
	if _, ok := extendPipe.Expression().(*PrintOperator); !ok {
		t.Fatalf("initial pipe expression = %T, want *PrintOperator", extendPipe.Expression())
	}
	if _, ok := take.(*TakeOperator); !ok {
		t.Fatalf("last operator = %T, want *TakeOperator", take)
	}
	if _, ok := project.(*ProjectOperator); !ok {
		t.Fatalf("project operator = %T, want *ProjectOperator", project)
	}

	extend := extendNode.(*ExtendOperator)
	named, ok := extend.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("extend expression = %#v, want named expression", extend.Expressions().At(0).Element())
	}
	add, ok := named.Expression().(*BinaryExpression)
	if !ok || add.Kind() != SyntaxKindAddExpression || add.Right().Kind() != SyntaxKindMultiplyExpression {
		t.Fatalf("extend value = %#v, want add with multiply RHS", named.Expression())
	}

	filter := filterNode.(*FilterOperator)
	orExpression, ok := filter.Condition().(*BinaryExpression)
	if !ok || orExpression.Kind() != SyntaxKindOrExpression {
		t.Fatalf("where condition = %#v, want OrExpression", filter.Condition())
	}
	if orExpression.Left().Kind() != SyntaxKindAndExpression || orExpression.Right().Kind() != SyntaxKindEqualExpression {
		t.Fatalf("where precedence left/right = %v/%v, want AndExpression/EqualExpression", orExpression.Left().Kind(), orExpression.Right().Kind())
	}
}

func TestQueryGrammarPublicPipeRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingCondition := parseQueryGrammarProduction(t, grammar.QueryOperator, "where").(*FilterOperator)
	if !diagnosticsContainMessage(missingCondition.Condition().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing where condition diagnostics = %#v", missingCondition.ContainedDiagnostics())
	}

	missingExpression := parseQueryGrammarProduction(t, grammar.QueryOperator, "take").(*TakeOperator)
	if !diagnosticsContainMessage(missingExpression.Expression().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing take expression diagnostics = %#v", missingExpression.ContainedDiagnostics())
	}

	missingList := parseQueryGrammarProduction(t, grammar.QueryOperator, "extend").(*ExtendOperator)
	if missingList.Expressions().Count() != 1 ||
		!diagnosticsContainMessage(missingList.Expressions().At(0).Element().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing extend list = count:%d diagnostics:%#v", missingList.Expressions().Count(), missingList.ContainedDiagnostics())
	}

	missingOperator := parseQueryGrammarProduction(t, grammar.Expression, "print x=1 |").(*PipeExpression)
	bad, ok := missingOperator.Operator().(*BadQueryOperator)
	if !ok || !diagnosticsContain(bad.Diagnostics(), "KS176") {
		t.Fatalf("missing pipe operator = %#v diagnostics:%#v", missingOperator.Operator(), missingOperator.ContainedDiagnostics())
	}
}

func queryGrammarPipeStage(t *testing.T, expression Expression, want SyntaxKind) (*PipeExpression, QueryOperator) {
	t.Helper()
	pipe, ok := expression.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe stage = %T, want *PipeExpression with operator kind %v", expression, want)
	}
	if pipe.Operator().Kind() != want {
		t.Fatalf("pipe stage operator kind = %v, want %v", pipe.Operator().Kind(), want)
	}
	return pipe, pipe.Operator()
}
