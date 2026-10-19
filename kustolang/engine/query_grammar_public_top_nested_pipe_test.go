package kustolang

import "testing"

func TestQueryGrammarPublicTopNestedASTAndOptionalFields(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	operator := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested Limit=5 of Group=Category with others='Other' by Total=sum(Value) desc nulls last")
	topNested, ok := operator.(*TopNestedOperator)
	if !ok {
		t.Fatalf("top-nested operator = %T, want *TopNestedOperator", operator)
	}
	if topNested.Clauses().Count() != 1 {
		t.Fatalf("top-nested clauses = %d, want 1", topNested.Clauses().Count())
	}
	clause := topNested.Clauses().At(0).Element()
	if clause.TopNestedKeyword().Kind() != SyntaxKindTopNestedKeyword ||
		clause.OfKeyword().Kind() != SyntaxKindOfKeyword || clause.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("top-nested clause keywords = %v/%v/%v", clause.TopNestedKeyword().Kind(), clause.OfKeyword().Kind(), clause.ByKeyword().Kind())
	}
	assertQueryGrammarSimpleNamedExpression(t, clause.Expression(), "Limit", SyntaxKindLongLiteralExpression)
	assertQueryGrammarSimpleNamedExpression(t, clause.OfExpression(), "Group", SyntaxKindNameReference)

	withOthers := clause.WithOthersClause()
	if withOthers == nil || withOthers.WithKeyword().Kind() != SyntaxKindWithKeyword ||
		withOthers.OthersKeyword().Kind() != SyntaxKindOthersKeyword || withOthers.Equal().Kind() != SyntaxKindEqualToken ||
		withOthers.Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("with-others clause = %#v, want string literal clause", withOthers)
	}

	ordered, ok := clause.ByExpression().(*OrderedExpression)
	if !ok {
		t.Fatalf("top-nested by expression = %T, want *OrderedExpression", clause.ByExpression())
	}
	assertQueryGrammarSimpleNamedExpression(t, ordered.Expression(), "Total", SyntaxKindFunctionCallExpression)
	if ordering := ordered.Ordering(); ordering.AscOrDescKeyword() == nil ||
		ordering.AscOrDescKeyword().Kind() != SyntaxKindDescKeyword || ordering.NullsClause() == nil ||
		ordering.NullsClause().FirstOrLastKeyword().Kind() != SyntaxKindLastKeyword {
		t.Fatalf("top-nested ordering = %#v, want desc nulls last", ordered.Ordering())
	}
	if diagnostics := topNested.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("top-nested diagnostics = %#v", diagnostics)
	}

	withoutCount := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested of Category by count() asc nulls first").(*TopNestedOperator).Clauses().At(0).Element()
	if withoutCount.Expression() != nil || withoutCount.WithOthersClause() != nil {
		t.Fatalf("optional count/with-others = %#v/%#v, want nil/nil", withoutCount.Expression(), withoutCount.WithOthersClause())
	}
	assertQueryGrammarOrderedExpression(t, withoutCount.ByExpression(), SyntaxKindAscKeyword, SyntaxKindFirstKeyword)
	if diagnostics := withoutCount.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("countless top-nested diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicTopNestedMultipleClausesAndPipeline(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested 3 of Region by count(), top-nested of City with others=0 by Total=sum(Value) nulls first")
	topNested := operator.(*TopNestedOperator)
	if topNested.Clauses().Count() != 2 {
		t.Fatalf("top-nested clauses = %d, want 2", topNested.Clauses().Count())
	}
	first := topNested.Clauses().At(0)
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("first top-nested separator = %#v, want comma", first.Separator())
	}
	if first.Element().Expression().Kind() != SyntaxKindLongLiteralExpression ||
		NameReferenceSimpleName(first.Element().OfExpression().(*NameReference)) != "Region" {
		t.Fatalf("first top-nested clause = %#v", first.Element())
	}
	second := topNested.Clauses().At(1).Element()
	if second.Expression() != nil || second.WithOthersClause() == nil ||
		second.WithOthersClause().Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("second top-nested optional fields = %#v/%#v", second.Expression(), second.WithOthersClause())
	}
	assertQueryGrammarOrderedExpression(t, second.ByExpression(), SyntaxKindNone, SyntaxKindFirstKeyword)
	if diagnostics := topNested.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("multi-clause top-nested diagnostics = %#v", diagnostics)
	}

	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"Events | top-nested 4 of Category with others='Other' by Total=count() desc | project Category, Total | take 10")
	takePipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindTakeOperator)
	projectPipe, _ := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindProjectOperator)
	topNestedPipe, topNestedNode := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindTopNestedOperator)
	if source, ok := topNestedPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("top-nested pipeline source = %#v, want Events", topNestedPipe.Expression())
	}
	pipedClause := topNestedNode.(*TopNestedOperator).Clauses().At(0).Element()
	if pipedClause.WithOthersClause() == nil || pipedClause.ByExpression().Kind() != SyntaxKindOrderedExpression {
		t.Fatalf("piped top-nested clause = %#v", pipedClause)
	}
}

func TestQueryGrammarPublicTopNestedRequiredPartRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name   string
		text   string
		assert func(*testing.T, *TopNestedClause)
	}{
		{
			name: "missing of keyword",
			text: "top-nested 5 Category by count()",
			assert: func(t *testing.T, clause *TopNestedClause) {
				assertQueryGrammarMissingTopNestedToken(t, clause.OfKeyword(), SyntaxKindOfKeyword)
			},
		},
		{
			name: "missing by keyword",
			text: "top-nested 5 of Category count()",
			assert: func(t *testing.T, clause *TopNestedClause) {
				assertQueryGrammarMissingTopNestedToken(t, clause.ByKeyword(), SyntaxKindByKeyword)
			},
		},
		{
			name: "missing of operand",
			text: "top-nested 5 of by count()",
			assert: func(t *testing.T, clause *TopNestedClause) {
				assertQueryGrammarMissingTopNestedExpression(t, clause.OfExpression())
			},
		},
		{
			name: "missing by operand",
			text: "top-nested 5 of Category by",
			assert: func(t *testing.T, clause *TopNestedClause) {
				assertQueryGrammarMissingTopNestedExpression(t, clause.ByExpression())
			},
		},
		{
			name: "with missing others keyword",
			text: "top-nested 5 of Category with ='Other' by count()",
			assert: func(t *testing.T, clause *TopNestedClause) {
				if clause.WithOthersClause() == nil {
					t.Fatal("with-others clause = nil")
				}
				assertQueryGrammarMissingTopNestedToken(t, clause.WithOthersClause().OthersKeyword(), SyntaxKindOthersKeyword)
			},
		},
		{
			name: "with others missing equal",
			text: "top-nested 5 of Category with others 'Other' by count()",
			assert: func(t *testing.T, clause *TopNestedClause) {
				if clause.WithOthersClause() == nil {
					t.Fatal("with-others clause = nil")
				}
				assertQueryGrammarMissingTopNestedToken(t, clause.WithOthersClause().Equal(), SyntaxKindEqualToken)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*TopNestedOperator)
			if operator.Clauses().Count() != 1 {
				t.Fatalf("top-nested clauses = %d, want 1", operator.Clauses().Count())
			}
			test.assert(t, operator.Clauses().At(0).Element())
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"top-nested 5 of Category with others=Fallback by count()", 5)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"top-nested 5 of Category with others=", 5)
}

func TestQueryGrammarPublicTopNestedListRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingElement := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested 3 of Region by count(),, top-nested 2 of City by count()").(*TopNestedOperator)
	if missingElement.Clauses().Count() != 3 {
		t.Fatalf("double-comma clauses = %d, want 3", missingElement.Clauses().Count())
	}
	middle := missingElement.Clauses().At(1).Element()
	if middle.TopNestedKeyword().RawText() != "" || !diagnosticsContain(middle.Diagnostics(), "KS006") {
		t.Fatalf("missing middle clause = %#v diagnostics:%#v", middle, middle.Diagnostics())
	}

	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested 3 of Region by count() top-nested 2 of City by count()").(*TopNestedOperator)
	if missingComma.Clauses().Count() != 2 {
		t.Fatalf("missing-comma clauses = %d, want 2", missingComma.Clauses().Count())
	}
	separator := missingComma.Clauses().At(0).Separator()
	if separator == nil || separator.RawText() != "" || !diagnosticsContain(separator.Diagnostics(), "KS005") {
		t.Fatalf("missing comma = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}

	trailingComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top-nested 3 of Region by count(),").(*TopNestedOperator)
	if trailingComma.Clauses().Count() != 2 {
		t.Fatalf("trailing-comma clauses = %d, want 2", trailingComma.Clauses().Count())
	}
	missing := trailingComma.Clauses().At(1).Element()
	if missing.TopNestedKeyword().RawText() != "" || !diagnosticsContain(missing.Diagnostics(), "KS006") {
		t.Fatalf("trailing-comma missing clause = %#v diagnostics:%#v", missing, missing.Diagnostics())
	}
}

func assertQueryGrammarSimpleNamedExpression(t *testing.T, expression Expression, name string, valueKind SyntaxKind) {
	t.Helper()
	named, ok := expression.(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(named.Name()) != name || named.Expression().Kind() != valueKind {
		t.Fatalf("named expression = %#v, want %s with value kind %v", expression, name, valueKind)
	}
}

func assertQueryGrammarMissingTopNestedToken(t *testing.T, token *SyntaxToken, kind SyntaxKind) {
	t.Helper()
	if token == nil {
		t.Fatalf("missing %v token = nil", kind)
	}
	if token.Kind() != kind || token.RawText() != "" || !diagnosticsContain(token.Diagnostics(), "KS005") {
		t.Fatalf("missing %v token = %#v diagnostics:%#v", kind, token, token.Diagnostics())
	}
}

func assertQueryGrammarMissingTopNestedExpression(t *testing.T, expression Expression) {
	t.Helper()
	if expression == nil {
		t.Fatal("missing expression = nil")
	}
	if !diagnosticsContain(expression.Diagnostics(), "KS006") {
		t.Fatalf("missing expression = %#v diagnostics:%#v", expression, expression.Diagnostics())
	}
}
