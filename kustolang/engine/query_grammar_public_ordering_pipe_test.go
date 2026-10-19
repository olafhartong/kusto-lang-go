package kustolang

import "testing"

func TestQueryGrammarPublicOrderingOperatorKindsAndAliases(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name string
		text string
		want SyntaxKind
	}{
		{name: "sort", text: "sort by Timestamp", want: SyntaxKindSortOperator},
		{name: "order alias", text: "order by Timestamp", want: SyntaxKindSortOperator},
		{name: "top", text: "top 5 by Score", want: SyntaxKindTopOperator},
		{name: "top hitters", text: "top-hitters 10 of User", want: SyntaxKindTopHittersOperator},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if operator.Kind() != test.want {
				t.Fatalf("operator %q kind = %v, want %v", test.text, operator.Kind(), test.want)
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("operator %q diagnostics = %#v", test.text, diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicSortOrderingASTHintsAndDefaults(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"order hint.strategy=splitBlock by Timestamp desc nulls last, Tenant asc nulls first, Sequence nulls last, DefaultOrder")
	sort, ok := operator.(*SortOperator)
	if !ok {
		t.Fatalf("order operator = %T, want *SortOperator", operator)
	}
	if sort.Keyword().Kind() != SyntaxKindOrderKeyword {
		t.Fatalf("order alias keyword kind = %v", sort.Keyword().Kind())
	}
	if sort.Parameters().Count() != 1 {
		t.Fatalf("sort parameters = %d, want 1", sort.Parameters().Count())
	}
	parameter := sort.Parameters().At(0)
	if NameDeclarationSimpleName(parameter.Name()) != "hint.strategy" || parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("sort hint = %#v, want hint.strategy token literal", parameter)
	}
	if sort.ByKeyword().Kind() != SyntaxKindByKeyword || sort.Expressions().Count() != 4 {
		t.Fatalf("sort by/list = %v/%d, want by with four keys", sort.ByKeyword().Kind(), sort.Expressions().Count())
	}

	assertQueryGrammarOrderedExpression(t, sort.Expressions().At(0).Element(), SyntaxKindDescKeyword, SyntaxKindLastKeyword)
	assertQueryGrammarOrderedExpression(t, sort.Expressions().At(1).Element(), SyntaxKindAscKeyword, SyntaxKindFirstKeyword)
	assertQueryGrammarOrderedExpression(t, sort.Expressions().At(2).Element(), SyntaxKindNone, SyntaxKindLastKeyword)
	defaultOrder, ok := sort.Expressions().At(3).Element().(*NameReference)
	if !ok || NameReferenceSimpleName(defaultOrder) != "DefaultOrder" {
		t.Fatalf("default ordered key = %#v, want unwrapped DefaultOrder name", sort.Expressions().At(3).Element())
	}
}

func TestQueryGrammarPublicTopAndTopHittersAST(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	top := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top hint.progressive_top=true 5 by Score nulls first").(*TopOperator)
	if top.Parameters().Count() != 1 || NameDeclarationSimpleName(top.Parameters().At(0).Name()) != "hint.progressive_top" {
		t.Fatalf("top parameters = %#v, want hint.progressive_top", top.Parameters())
	}
	if top.Parameters().At(0).Expression().Kind() != SyntaxKindBooleanLiteralExpression || top.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("top hint/count kinds = %v/%v", top.Parameters().At(0).Expression().Kind(), top.Expression().Kind())
	}
	if top.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("top by keyword = %v", top.ByKeyword().Kind())
	}
	assertQueryGrammarOrderedExpression(t, top.ByExpression(), SyntaxKindNone, SyntaxKindFirstKeyword)

	hitters := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-hitters 10 of Principal=User by Total=Score").(*TopHittersOperator)
	if hitters.Expression().Kind() != SyntaxKindLongLiteralExpression || hitters.OfKeyword().Kind() != SyntaxKindOfKeyword {
		t.Fatalf("top-hitters count/of kinds = %v/%v", hitters.Expression().Kind(), hitters.OfKeyword().Kind())
	}
	ofExpression, ok := hitters.OfExpression().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(ofExpression.Name()) != "Principal" {
		t.Fatalf("top-hitters of expression = %#v, want Principal assignment", hitters.OfExpression())
	}
	if hitters.ByClause() == nil || hitters.ByClause().ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("top-hitters by clause = %#v", hitters.ByClause())
	}
	byExpression, ok := hitters.ByClause().Expression().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(byExpression.Name()) != "Total" {
		t.Fatalf("top-hitters by expression = %#v, want Total assignment", hitters.ByClause().Expression())
	}
}

func TestQueryGrammarPublicOrderingMultiStagePipe(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		"T | order hint.strategy=splitBlock by Timestamp desc nulls last, Sequence | top hint.progressive_top=true 5 by Score asc nulls first | top-hitters 3 of User by Score")

	hittersPipe, hittersNode := queryGrammarPipeStage(t, expression, SyntaxKindTopHittersOperator)
	topPipe, topNode := queryGrammarPipeStage(t, hittersPipe.Expression(), SyntaxKindTopOperator)
	sortPipe, sortNode := queryGrammarPipeStage(t, topPipe.Expression(), SyntaxKindSortOperator)
	if source, ok := sortPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("pipe source = %#v, want T", sortPipe.Expression())
	}
	if sortNode.(*SortOperator).Expressions().Count() != 2 || topNode.(*TopOperator).ByExpression().Kind() != SyntaxKindOrderedExpression {
		t.Fatalf("sort/top pipeline AST = %#v / %#v", sortNode, topNode)
	}
	if hittersNode.(*TopHittersOperator).ByClause() == nil {
		t.Fatal("final top-hitters stage is missing its by clause")
	}
}

func TestQueryGrammarPublicOrderingRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name  string
		text  string
		kind  SyntaxKind
		codes []string
	}{
		{name: "sort missing by and expression", text: "sort", kind: SyntaxKindSortOperator, codes: []string{"KS005", "KS006"}},
		{name: "order missing expression", text: "order by", kind: SyntaxKindSortOperator, codes: []string{"KS006"}},
		{name: "top missing count by and expression", text: "top", kind: SyntaxKindTopOperator, codes: []string{"KS005", "KS006"}},
		{name: "top missing count", text: "top by Score", kind: SyntaxKindTopOperator, codes: []string{"KS006"}},
		{name: "top missing by expression", text: "top 5 by", kind: SyntaxKindTopOperator, codes: []string{"KS006"}},
		{name: "top hitters missing count", text: "top-hitters of User", kind: SyntaxKindTopHittersOperator, codes: []string{"KS006"}},
		{name: "top hitters missing of and expression", text: "top-hitters 5", kind: SyntaxKindTopHittersOperator, codes: []string{"KS005", "KS006"}},
		{name: "top hitters missing of expression", text: "top-hitters 5 of", kind: SyntaxKindTopHittersOperator, codes: []string{"KS006"}},
		{name: "top hitters missing by expression", text: "top-hitters 5 of User by", kind: SyntaxKindTopHittersOperator, codes: []string{"KS006"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if operator.Kind() != test.kind {
				t.Fatalf("operator kind = %v, want %v", operator.Kind(), test.kind)
			}
			diagnostics := operator.ContainedDiagnostics()
			for _, code := range test.codes {
				if !diagnosticsContain(diagnostics, code) {
					t.Fatalf("operator %q missing %s diagnostic: %#v", test.text, code, diagnostics)
				}
			}
		})
	}
}

func TestQueryGrammarPublicOrderingNullRecoveryAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{"sort by Score desc nulls", "sort by Score nulls"} {
		t.Run(text, func(t *testing.T) {
			sort := parseQueryGrammarProduction(t, grammar.QueryOperator, text).(*SortOperator)
			ordered, ok := sort.Expressions().At(0).Element().(*OrderedExpression)
			if !ok || ordered.Ordering().NullsClause() == nil {
				t.Fatalf("sort expression = %#v, want ordered expression with nulls clause", sort.Expressions().At(0).Element())
			}
			firstOrLast := ordered.Ordering().NullsClause().FirstOrLastKeyword()
			if firstOrLast.RawText() != "" || !diagnosticsContainMessage(firstOrLast.Diagnostics(), "KS006", "Missing first or last") {
				t.Fatalf("incomplete nulls clause token/diagnostics = %#v/%#v", firstOrLast, firstOrLast.Diagnostics())
			}
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "sort by Score nulls first desc", 5)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "sort by Score nulls asc", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "top 5 by Score nulls last asc", 6)
}

func assertQueryGrammarOrderedExpression(t *testing.T, expression Expression, direction, nullPlacement SyntaxKind) {
	t.Helper()
	ordered, ok := expression.(*OrderedExpression)
	if !ok {
		t.Fatalf("ordered expression = %T, want *OrderedExpression", expression)
	}
	ordering := ordered.Ordering()
	if direction == SyntaxKindNone {
		if ordering.AscOrDescKeyword() != nil {
			t.Fatalf("ordering direction = %v, want omitted", ordering.AscOrDescKeyword().Kind())
		}
	} else if ordering.AscOrDescKeyword() == nil || ordering.AscOrDescKeyword().Kind() != direction {
		t.Fatalf("ordering direction = %#v, want %v", ordering.AscOrDescKeyword(), direction)
	}
	if ordering.NullsClause() == nil || ordering.NullsClause().FirstOrLastKeyword().Kind() != nullPlacement {
		t.Fatalf("ordering null placement = %#v, want %v", ordering.NullsClause(), nullPlacement)
	}
}
