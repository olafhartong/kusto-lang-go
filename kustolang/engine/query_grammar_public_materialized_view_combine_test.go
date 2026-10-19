package kustolang

import "testing"

func TestQueryGrammarPublicMaterializedViewCombineCompleteAST(t *testing.T) {
	combine := parseQueryGrammarMaterializedViewCombine(t,
		`materialized-view-combine("SigninView") base(SigninLogs) delta(SigninDelta) aggregations(summarize Total=count() by UserPrincipalName)`)

	if combine.MaterializedViewCombineKeyword().Kind() != SyntaxKindMaterializedViewCombineKeyword ||
		combine.MaterializedViewCombineKeyword().RawText() != "materialized-view-combine" {
		t.Fatalf("materialized-view-combine keyword = %#v", combine.MaterializedViewCombineKeyword())
	}
	if combine.ViewName().OpenParen().Kind() != SyntaxKindOpenParenToken ||
		combine.ViewName().Value().Kind() != SyntaxKindStringLiteralExpression ||
		combine.ViewName().CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("view-name clause = %#v", combine.ViewName())
	}
	assertQueryGrammarMaterializedClause(t, combine.BaseClause(), "base", SyntaxKindNameReference)
	assertQueryGrammarMaterializedClause(t, combine.DeltaClause(), "delta", SyntaxKindNameReference)
	assertQueryGrammarMaterializedClause(t, combine.AggregationsClause(), "aggregations", SyntaxKindSummarizeOperator)

	summarize := combine.AggregationsClause().Expression().(*SummarizeOperator)
	if summarize.Aggregates().Count() != 1 || summarize.ByClause() == nil || summarize.ByClause().Expressions().Count() != 1 {
		t.Fatalf("aggregations summarize = %#v, want one aggregate and one by expression", summarize)
	}
	if diagnostics := combine.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("complete materialized-view-combine diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMaterializedViewCombinePipelinesAndNestedExpressions(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`materialized-view-combine(strcat("Signin", "View")) base(SigninLogs | where ResultType == 0 | project UserPrincipalName, Duration=1 + 2 * 3) delta(SigninDelta | extend Normalized=tolower(UserPrincipalName)) aggregations(summarize Total=sum(iif(ResultType == 0, 1, 0)) by Tenant=tolower(TenantId)) | project Tenant, Total`)

	projectPipe, project := queryGrammarPipeStage(t, expression, SyntaxKindProjectOperator)
	if project.(*ProjectOperator).Expressions().Count() != 2 {
		t.Fatalf("outer project = %#v, want two expressions", project)
	}
	combine, ok := projectPipe.Expression().(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("outer pipe source = %T, want *MaterializedViewCombineExpression", projectPipe.Expression())
	}
	if combine.ViewName().Value().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("nested view-name expression kind = %v", combine.ViewName().Value().Kind())
	}
	baseProjectPipe, _ := queryGrammarPipeStage(t, combine.BaseClause().Expression(), SyntaxKindProjectOperator)
	_, _ = queryGrammarPipeStage(t, baseProjectPipe.Expression(), SyntaxKindFilterOperator)
	_, _ = queryGrammarPipeStage(t, combine.DeltaClause().Expression(), SyntaxKindExtendOperator)
	summarize := combine.AggregationsClause().Expression().(*SummarizeOperator)
	if summarize.Aggregates().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression ||
		summarize.ByClause().Expressions().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("nested summarize expressions = %#v / %#v", summarize.Aggregates(), summarize.ByClause().Expressions())
	}
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("pipeline materialized-view-combine diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMaterializedViewCombineMissingEntireClauses(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name        string
		text        string
		clause      func(*MaterializedViewCombineExpression) *MaterializedViewCombineClause
		wantMessage string
	}{
		{
			name:        "base",
			text:        `materialized-view-combine("v") delta(D) aggregations(summarize count())`,
			clause:      (*MaterializedViewCombineExpression).BaseClause,
			wantMessage: "Missing base clause",
		},
		{
			name:        "delta",
			text:        `materialized-view-combine("v") base(B) aggregations(summarize count())`,
			clause:      (*MaterializedViewCombineExpression).DeltaClause,
			wantMessage: "Missing delta clause",
		},
		{
			name:        "aggregations uses pinned diagnostic name",
			text:        `materialized-view-combine("v") base(B) delta(D)`,
			clause:      (*MaterializedViewCombineExpression).AggregationsClause,
			wantMessage: "Missing aggregates clause",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			combine := parseQueryGrammarMaterializedViewCombineWith(t, grammar.Expression, test.text)
			clause := test.clause(combine)
			if clause.Keyword().Width() != 0 || clause.OpenParen().Width() != 0 || clause.CloseParen().Width() != 0 {
				t.Fatalf("missing clause delimiters = %d/%d/%d, want all omitted", clause.Keyword().Width(), clause.OpenParen().Width(), clause.CloseParen().Width())
			}
			if clause.Expression().Kind() != SyntaxKindNameReference {
				t.Fatalf("missing clause expression kind = %v, want recovered NameReference", clause.Expression().Kind())
			}
			if !diagnosticsContainMessage(clause.Diagnostics(), "KS006", test.wantMessage) {
				t.Fatalf("missing clause diagnostics = %#v, want KS006 %q", clause.Diagnostics(), test.wantMessage)
			}
		})
	}
}

func TestQueryGrammarPublicMaterializedViewCombineMissingParensAndExpressions(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		selectNode func(*MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken)
		wantOpen   bool
		wantClose  bool
		wantExpr   bool
	}{
		{
			name: "name expression",
			text: `materialized-view-combine() base(B) delta(D) aggregations(summarize count())`,
			selectNode: func(c *MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken) {
				return c.ViewName().Value(), c.ViewName().OpenParen(), c.ViewName().CloseParen()
			},
			wantExpr: true,
		},
		{
			name: "name close paren",
			text: `materialized-view-combine("v" base(B) delta(D) aggregations(summarize count())`,
			selectNode: func(c *MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken) {
				return c.ViewName().Value(), c.ViewName().OpenParen(), c.ViewName().CloseParen()
			},
			wantClose: true,
		},
		{
			name: "base open paren",
			text: `materialized-view-combine("v") base B) delta(D) aggregations(summarize count())`,
			selectNode: func(c *MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken) {
				return c.BaseClause().Expression(), c.BaseClause().OpenParen(), c.BaseClause().CloseParen()
			},
			wantOpen: true,
		},
		{
			name: "base expression",
			text: `materialized-view-combine("v") base() delta(D) aggregations(summarize count())`,
			selectNode: func(c *MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken) {
				return c.BaseClause().Expression(), c.BaseClause().OpenParen(), c.BaseClause().CloseParen()
			},
			wantExpr: true,
		},
		{
			name: "base close paren",
			text: `materialized-view-combine("v") base(B delta(D) aggregations(summarize count())`,
			selectNode: func(c *MaterializedViewCombineExpression) (Expression, *SyntaxToken, *SyntaxToken) {
				return c.BaseClause().Expression(), c.BaseClause().OpenParen(), c.BaseClause().CloseParen()
			},
			wantClose: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			combine := parseQueryGrammarMaterializedViewCombine(t, test.text)
			expression, open, close := test.selectNode(combine)
			if test.wantOpen && (open.Width() != 0 || !diagnosticsContainMessage(open.Diagnostics(), "KS005", "Expected: (")) {
				t.Fatalf("open paren recovery = width:%d diagnostics:%#v", open.Width(), open.Diagnostics())
			}
			if test.wantClose && (close.Width() != 0 || !diagnosticsContainMessage(close.Diagnostics(), "KS005", "Expected: )")) {
				t.Fatalf("close paren recovery = width:%d diagnostics:%#v", close.Width(), close.Diagnostics())
			}
			if test.wantExpr && (expression.Kind() != SyntaxKindNameReference || !diagnosticsContainMessage(expression.Diagnostics(), "KS006", "Missing expression")) {
				t.Fatalf("expression recovery = kind:%v diagnostics:%#v", expression.Kind(), expression.Diagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicMaterializedViewCombineAggregationsRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	empty := parseQueryGrammarMaterializedViewCombineWith(t, grammar.Expression,
		`materialized-view-combine("v") base(B) delta(D) aggregations()`)
	assertQueryGrammarMaterializedBadAggregation(t, empty.AggregationsClause())
	if empty.AggregationsClause().CloseParen().Width() == 0 {
		t.Fatalf("empty aggregations close paren = %#v, want source token", empty.AggregationsClause().CloseParen())
	}

	text := `materialized-view-combine("v") base(B) delta(D) aggregations(where Enabled)`
	tokens := Lex(text)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	if result.Length < 0 || result.Length >= len(tokens)-1 {
		t.Fatalf("non-summarize aggregations parse = %#v, want recovered prefix with unconsumed operator", result)
	}
	combine, ok := result.Value.(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("non-summarize aggregations value = %T", result.Value)
	}
	assertQueryGrammarMaterializedBadAggregation(t, combine.AggregationsClause())
	if combine.AggregationsClause().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(combine.AggregationsClause().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("non-summarize close recovery = %#v", combine.AggregationsClause().CloseParen())
	}
}

func TestQueryGrammarPublicMaterializedViewCombineGeneratedChildShapes(t *testing.T) {
	tests := []struct {
		name     string
		kind     SyntaxKind
		baseType string
		children []GeneratedSyntaxNodeChild
	}{
		{
			name: "MaterializedViewCombineExpression", kind: SyntaxKindMaterializedViewCombineExpression, baseType: "Expression",
			children: []GeneratedSyntaxNodeChild{
				{Name: "MaterializedViewCombineKeyword", Type: "SyntaxToken", Hint: CompletionHintSyntax},
				{Name: "ViewName", Type: "MaterializedViewCombineNameClause", Hint: CompletionHintSyntax},
				{Name: "BaseClause", Type: "MaterializedViewCombineClause", Hint: CompletionHintSyntax},
				{Name: "DeltaClause", Type: "MaterializedViewCombineClause", Hint: CompletionHintSyntax},
				{Name: "AggregationsClause", Type: "MaterializedViewCombineClause", Hint: CompletionHintSyntax},
			},
		},
		{
			name: "MaterializedViewCombineNameClause", kind: SyntaxKindMaterializedViewCombineExpression, baseType: "SyntaxNode",
			children: []GeneratedSyntaxNodeChild{
				{Name: "OpenParen", Type: "SyntaxToken", Hint: CompletionHintSyntax},
				{Name: "Value", Type: "Expression", Hint: CompletionHintLiteral},
				{Name: "CloseParen", Type: "SyntaxToken", Hint: CompletionHintSyntax},
			},
		},
		{
			name: "MaterializedViewCombineClause", kind: SyntaxKindMaterializedViewCombineClause, baseType: "SyntaxNode",
			children: []GeneratedSyntaxNodeChild{
				{Name: "Keyword", Type: "SyntaxToken", Hint: CompletionHintSyntax},
				{Name: "OpenParen", Type: "SyntaxToken", Hint: CompletionHintSyntax},
				{Name: "Expression", Type: "Expression", Hint: CompletionHintSyntax},
				{Name: "CloseParen", Type: "SyntaxToken", Hint: CompletionHintSyntax},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata, ok := GeneratedSyntaxNodeMetadataForName(test.name)
			if !ok {
				t.Fatalf("generated metadata %q not found", test.name)
			}
			if metadata.Kind != test.kind || metadata.BaseType != test.baseType || len(metadata.Children) != len(test.children) {
				t.Fatalf("generated metadata = %#v", metadata)
			}
			for i, want := range test.children {
				got := metadata.Children[i]
				if got.Name != want.Name || got.Type != want.Type || got.Optional || got.Hint != want.Hint {
					t.Fatalf("generated child %d = %#v, want %#v", i, got, want)
				}
			}
		})
	}
}

func TestQueryGrammarPublicMaterializedViewCombineRootOnlyPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := `materialized-view-combine("v") base(B) delta(D) aggregations(summarize count())`
	if _, ok := parseQueryGrammarProduction(t, grammar.Expression, text).(*MaterializedViewCombineExpression); !ok {
		t.Fatal("materialized-view-combine was not accepted as a root expression")
	}
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, text)
	assertQueryGrammarProductionRejects(t, grammar.FollowingPipeElementExpression, text)
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, `materialized_view_combine("v")`)
}

func parseQueryGrammarMaterializedViewCombine(t *testing.T, text string) *MaterializedViewCombineExpression {
	t.Helper()
	return parseQueryGrammarMaterializedViewCombineWith(t, QueryGrammarFrom(nil).Expression, text)
}

func parseQueryGrammarMaterializedViewCombineWith(t *testing.T, parser ResultParser[LexicalToken, Expression], text string) *MaterializedViewCombineExpression {
	t.Helper()
	expression := parseQueryGrammarProduction(t, parser, text)
	combine, ok := expression.(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("materialized-view-combine expression = %T, want *MaterializedViewCombineExpression", expression)
	}
	return combine
}

func assertQueryGrammarMaterializedClause(t *testing.T, clause *MaterializedViewCombineClause, keyword string, expressionKind SyntaxKind) {
	t.Helper()
	if clause.Keyword().RawText() != keyword || clause.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		clause.Expression().Kind() != expressionKind || clause.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("%s clause = %#v, expression kind %v", keyword, clause, clause.Expression().Kind())
	}
}

func assertQueryGrammarMaterializedBadAggregation(t *testing.T, clause *MaterializedViewCombineClause) {
	t.Helper()
	bad, ok := clause.Expression().(*BadQueryOperator)
	if !ok {
		t.Fatalf("aggregations recovery expression = %T, want *BadQueryOperator", clause.Expression())
	}
	if bad.Keyword().Width() != 0 || !diagnosticsContainMessage(bad.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("bad aggregations operator = %#v diagnostics:%#v", bad.Keyword(), bad.Diagnostics())
	}
}
