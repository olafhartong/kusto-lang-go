package kustolang

import "testing"

func TestQueryGrammarPublicMvApplyAliasesParametersAndExpressions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		text string
		kind SyntaxKind
	}{
		{text: "mv-apply Items on (where true)", kind: SyntaxKindMvDashApplyKeyword},
		{text: "mvapply Items on (where true)", kind: SyntaxKindMvApplyKeyword},
	} {
		t.Run(test.text, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*MvApplyOperator)
			if operator.MvApplyKeyword().Kind() != test.kind || operator.Expressions().Count() != 1 {
				t.Fatalf("mv-apply alias AST = keyword:%v expressions:%d", operator.MvApplyKeyword().Kind(), operator.Expressions().Count())
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("mv-apply alias diagnostics = %#v", diagnostics)
			}
		})
	}

	operator := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply with_itemindex=ItemIndex Expanded=Payload to typeof(string), Count=Values to typeof(long), Other limit 25 id ContextId on (where Expanded != '')").(*MvApplyOperator)
	if operator.Parameters().Count() != 1 {
		t.Fatalf("mv-apply parameters = %d, want 1", operator.Parameters().Count())
	}
	parameter := operator.Parameters().At(0)
	if NameDeclarationSimpleName(parameter.Name()) != "with_itemindex" || parameter.EqualToken().Width() == 0 ||
		parameter.Expression().Kind() != SyntaxKindNameDeclaration {
		t.Fatalf("mv-apply with_itemindex parameter = %#v", parameter)
	}
	if operator.Expressions().Count() != 3 {
		t.Fatalf("mv-apply expressions = %d, want 3", operator.Expressions().Count())
	}
	assertQueryGrammarMvApplyAssignment(t, operator.Expressions().At(0).Element(), "Expanded", "Payload", SyntaxKindStringKeyword)
	assertQueryGrammarMvApplyAssignment(t, operator.Expressions().At(1).Element(), "Count", "Values", SyntaxKindLongKeyword)
	third, ok := operator.Expressions().At(2).Element().Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(third) != "Other" || operator.Expressions().At(2).Element().ToTypeOf() != nil {
		t.Fatalf("third mv-apply expression = %#v", operator.Expressions().At(2).Element())
	}
	if operator.RowLimitClause() == nil || operator.RowLimitClause().LimitKeyword().Kind() != SyntaxKindLimitKeyword ||
		operator.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("mv-apply limit clause = %#v", operator.RowLimitClause())
	}
	if operator.ContextIdClause() == nil || operator.ContextIdClause().IdKeyword().Kind() != SyntaxKindIdKeyword ||
		operator.ContextIdClause().Id().Kind() != SyntaxKindNameReference {
		t.Fatalf("mv-apply id clause = %#v", operator.ContextIdClause())
	}
	if operator.OnKeyword().Kind() != SyntaxKindOnKeyword || operator.Subquery().Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("mv-apply on/subquery = %#v/%#v", operator.OnKeyword(), operator.Subquery())
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid mv-apply diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMvApplySingletonBareToTypeOf(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"mv-apply to typeof(dynamic) on (take 1)").(*MvApplyOperator)
	if operator.Expressions().Count() != 1 {
		t.Fatalf("bare-to expression count = %d, want 1", operator.Expressions().Count())
	}
	expression := operator.Expressions().At(0).Element()
	if expression.Expression() != nil || expression.ToTypeOf() == nil ||
		queryGrammarMvApplyTypeKind(expression) != SyntaxKindDynamicKeyword {
		t.Fatalf("bare-to mv-apply expression = %#v", expression)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("bare-to mv-apply diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMvApplyContextualSubqueries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operatorLeading := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Item=Items on (where Item != '' | project Item)").(*MvApplyOperator)
	projectPipe, _ := queryGrammarPipeStage(t, operatorLeading.Subquery().Expression(), SyntaxKindProjectOperator)
	if projectPipe.Expression().Kind() != SyntaxKindFilterOperator || len(operatorLeading.ContainedDiagnostics()) != 0 {
		t.Fatalf("operator-leading contextual subquery = %#v diagnostics:%#v",
			operatorLeading.Subquery().Expression(), operatorLeading.ContainedDiagnostics())
	}

	contextual := parseQueryGrammarProduction(t, grammar.QueryOperator,
		`mv-apply Item=Items on (__contextual_datatable "ContextId" (Item:string) | where Item != '')`).(*MvApplyOperator)
	wherePipe, _ := queryGrammarPipeStage(t, contextual.Subquery().Expression(), SyntaxKindFilterOperator)
	dataTable, ok := wherePipe.Expression().(*ContextualDataTableExpression)
	if !ok || dataTable.Schema().Columns().Count() != 1 || len(contextual.ContainedDiagnostics()) != 0 {
		t.Fatalf("contextual-datatable mv-apply subquery = %#v diagnostics:%#v",
			contextual.Subquery().Expression(), contextual.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicMvApplyPlacementCompletionAndTag(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := "mv-apply Item=Items on (take 1)"
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"query operator": grammar.QueryOperator,
		"following pipe": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, text).(*MvApplyOperator); !ok {
				t.Fatalf("mv-apply placement produced unexpected node")
			}
		})
	}
	root := parseQueryGrammarProduction(t, grammar.Expression, text)
	if _, ok := root.(*MvApplyOperator); !ok {
		t.Fatalf("root mv-apply placement = %T, want *MvApplyOperator", root)
	}

	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"T | mv-apply Item=Items on (take 1) | project Item")
	projectPipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindProjectOperator)
	applyPipe, apply := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindMvApplyOperator)
	if source, ok := applyPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("mv-apply pipeline source = %#v, want T", applyPipe.Expression())
	}
	if _, ok := apply.(*MvApplyOperator); !ok || len(expression.ContainedDiagnostics()) != 0 {
		t.Fatalf("mv-apply pipeline operator = %T diagnostics:%#v", apply, expression.ContainedDiagnostics())
	}

	items := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.QueryOperator, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	canonical, ok := completionItemByDisplayText(items, "mv-apply")
	if !ok || canonical.Kind != CompletionKindQueryPrefix {
		t.Fatalf("mv-apply completion = %#v, visible items:%#v", canonical, completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "mvapply"); ok {
		t.Fatalf("legacy mvapply completion became visible: %#v", completionDisplayTexts(items))
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<mvapply>", 0) {
		t.Fatal("QueryOperator parser tree is missing <mvapply> tag")
	}
}

func TestQueryGrammarPublicMvApplyRequiredRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingExpression := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply on (where true)").(*MvApplyOperator).Expressions().At(0).Element().Expression()
	assertQueryGrammarMvApplyMissingExpression(t, missingExpression)

	missingType := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Items to typeof() on (where true)").(*MvApplyOperator)
	typeOf := missingType.Expressions().At(0).Element().ToTypeOf().TypeOf()
	if typeOf.Types().Count() != 1 ||
		!diagnosticsContainMessage(typeOf.Types().At(0).Element().ContainedDiagnostics(), "KS006", "Missing type name") {
		t.Fatalf("missing mv-apply type = %#v diagnostics:%#v", typeOf, typeOf.ContainedDiagnostics())
	}

	missingOn := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Items limit 1 (where true)").(*MvApplyOperator)
	if missingOn.OnKeyword().Width() != 0 ||
		!diagnosticsContainMessage(missingOn.OnKeyword().Diagnostics(), "KS005", "Expected: on") {
		t.Fatalf("missing mv-apply on = %#v diagnostics:%#v", missingOn.OnKeyword(), missingOn.ContainedDiagnostics())
	}

	missingSubquery := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Items on ()").(*MvApplyOperator)
	assertQueryGrammarMvApplyMissingExpression(t, missingSubquery.Subquery().Expression())

	missingClose := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Items on (where true").(*MvApplyOperator)
	if missingClose.Subquery().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.Subquery().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing mv-apply close = %#v diagnostics:%#v", missingClose.Subquery().CloseParen(), missingClose.ContainedDiagnostics())
	}

	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"mv-apply Items Other on (where true)").(*MvApplyOperator)
	if missingComma.Expressions().Count() != 2 || missingComma.Expressions().At(0).Separator() == nil ||
		missingComma.Expressions().At(0).Separator().Width() != 0 ||
		!diagnosticsContainMessage(missingComma.Expressions().At(0).Separator().Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing mv-apply comma = %#v diagnostics:%#v", missingComma.Expressions(), missingComma.ContainedDiagnostics())
	}

	for _, test := range []struct {
		name   string
		text   string
		clause func(*MvApplyOperator) Expression
	}{
		{name: "limit expression", text: "mv-apply Items limit", clause: func(op *MvApplyOperator) Expression { return op.RowLimitClause().RowLimit() }},
		{name: "id expression", text: "mv-apply Items id", clause: func(op *MvApplyOperator) Expression { return op.ContextIdClause().Id() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*MvApplyOperator)
			assertQueryGrammarMvApplyMissingExpression(t, test.clause(operator))
		})
	}

	tokens := Lex("mv-apply Items on where true)")
	result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
	missingOpen, ok := result.Value.(*MvApplyOperator)
	if result.Length != 3 || !ok || missingOpen.Subquery().OpenParen().Width() != 0 ||
		!diagnosticsContainMessage(missingOpen.Subquery().OpenParen().Diagnostics(), "KS005", "Expected: (") {
		t.Fatalf("missing mv-apply open = %#v diagnostics:%#v", result, missingOpen.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicMvApplyBoundariesAndTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"mv-apply Items on (where true) | project Items", 7)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"mv-apply Items on (where true) Extra", 7)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator,
		"mv-apply Items limit 5 id ContextId on (where true) Extra", 11)

	for _, test := range []struct {
		name         string
		text         string
		missingIndex int
		count        int
	}{
		{name: "leading", text: "mv-apply , Items on (where true)", missingIndex: 0, count: 2},
		{name: "middle", text: "mv-apply Items,, Other on (where true)", missingIndex: 1, count: 3},
		{name: "trailing", text: "mv-apply Items, on (where true)", missingIndex: 1, count: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*MvApplyOperator)
			if operator.Expressions().Count() != test.count {
				t.Fatalf("mv-apply comma expression count = %d, want %d", operator.Expressions().Count(), test.count)
			}
			assertQueryGrammarMvApplyMissingExpression(t, operator.Expressions().At(test.missingIndex).Element().Expression())
		})
	}
}

func assertQueryGrammarMvApplyAssignment(t *testing.T, expression *MvApplyExpression, name, value string, typeKind SyntaxKind) {
	t.Helper()
	assignment, ok := expression.Expression().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(assignment.Name()) != name {
		t.Fatalf("mv-apply assignment = %#v, want %s", expression.Expression(), name)
	}
	reference, ok := assignment.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != value {
		t.Fatalf("mv-apply assignment value = %#v, want %s", assignment.Expression(), value)
	}
	clause := expression.ToTypeOf()
	if clause == nil || clause.ToKeyword().Kind() != SyntaxKindToKeyword || clause.TypeOf().Types().Count() != 1 ||
		queryGrammarMvApplyTypeKind(expression) != typeKind {
		t.Fatalf("mv-apply typeof clause = %#v, want %v", clause, typeKind)
	}
}

func queryGrammarMvApplyTypeKind(expression *MvApplyExpression) SyntaxKind {
	typeExpression, ok := expression.ToTypeOf().TypeOf().Types().At(0).Element().(*PrimitiveTypeExpression)
	if !ok || typeExpression.Type() == nil {
		return SyntaxKindNone
	}
	return typeExpression.Type().Kind()
}

func assertQueryGrammarMvApplyMissingExpression(t *testing.T, expression Expression) {
	t.Helper()
	if expression == nil {
		t.Fatal("missing mv-apply expression is nil")
	}
	if expression.Width() != 0 || !diagnosticsContain(expression.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing mv-apply expression = %#v diagnostics:%#v", expression, expression.ContainedDiagnostics())
	}
}
