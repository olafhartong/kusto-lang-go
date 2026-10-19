package kustolang

import "testing"

func TestQueryGrammarPublicMvExpandAliasesAndCompletionVisibility(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		text string
		kind SyntaxKind
	}{
		{text: "mv-expand Items", kind: SyntaxKindMvDashExpandKeyword},
		{text: "mvexpand Items", kind: SyntaxKindMvExpandKeyword},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*MvExpandOperator)
			if operator.MvExpandKeyword().Kind() != test.kind || operator.Expressions().Count() != 1 {
				t.Fatalf("mv-expand alias AST = keyword:%v expressions:%d", operator.MvExpandKeyword().Kind(), operator.Expressions().Count())
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("mv-expand alias diagnostics = %#v", diagnostics)
			}
		})
	}

	parsers := FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.QueryOperator, 0)
	items := []CompletionItem{}
	for _, parser := range parsers {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	canonical, ok := completionItemByDisplayText(items, "mv-expand")
	if !ok || canonical.Kind != CompletionKindQueryPrefix {
		t.Fatalf("mv-expand completion = %#v, visible items:%#v", canonical, completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "mvexpand"); ok {
		t.Fatalf("legacy mvexpand completion became visible: %#v", completionDisplayTexts(items))
	}
}

func TestQueryGrammarPublicMvExpandParametersAssignmentsTypeOfAndLimit(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"mv-expand kind=array bagexpansion=bag with_itemindex=ItemIndex Expanded=Payload to typeof(string), Count=Values to typeof(long), Other limit 25").(*MvExpandOperator)

	if operator.Parameters().Count() != 3 {
		t.Fatalf("mv-expand parameters = %d, want 3", operator.Parameters().Count())
	}
	wantParameters := []struct {
		name string
		kind SyntaxKind
	}{
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
		{name: "bagexpansion", kind: SyntaxKindTokenLiteralExpression},
		{name: "with_itemindex", kind: SyntaxKindNameDeclaration},
	}
	for i, want := range wantParameters {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want.name ||
			parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != want.kind {
			t.Fatalf("mv-expand parameter %d = %#v, want %q/%v", i, parameter, want.name, want.kind)
		}
	}

	if operator.Expressions().Count() != 3 {
		t.Fatalf("mv-expand expressions = %d, want 3", operator.Expressions().Count())
	}
	assertQueryGrammarMvExpandAssignment(t, operator.Expressions().At(0).Element(), "Expanded", "Payload", SyntaxKindStringKeyword)
	assertQueryGrammarMvExpandAssignment(t, operator.Expressions().At(1).Element(), "Count", "Values", SyntaxKindLongKeyword)
	third, ok := operator.Expressions().At(2).Element().Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(third) != "Other" || operator.Expressions().At(2).Element().ToTypeOf() != nil {
		t.Fatalf("third mv-expand expression = %#v", operator.Expressions().At(2).Element())
	}
	limit := operator.RowLimitClause()
	if limit == nil || limit.LimitKeyword().Kind() != SyntaxKindLimitKeyword || limit.RowLimit().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("mv-expand row limit = %#v", limit)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid mv-expand diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMvExpandProductionsAndMultiStagePlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	if _, ok := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand Items").(*MvExpandOperator); !ok {
		t.Fatal("QueryOperator did not produce MvExpandOperator")
	}
	if _, ok := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, "mv-expand Items").(*MvExpandOperator); !ok {
		t.Fatal("FollowingPipeElementExpression did not produce MvExpandOperator")
	}

	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"T | mv-expand kind=array Item=Items to typeof(string), Value limit 10 | where Item != '' | project Item")
	projectPipe, project := queryGrammarPipeStage(t, expression, SyntaxKindProjectOperator)
	wherePipe, where := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	mvPipe, mv := queryGrammarPipeStage(t, wherePipe.Expression(), SyntaxKindMvExpandOperator)
	if source, ok := mvPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("mv-expand pipeline source = %#v, want T", mvPipe.Expression())
	}
	if _, ok := mv.(*MvExpandOperator); !ok {
		t.Fatalf("first pipeline operator = %T, want *MvExpandOperator", mv)
	}
	if _, ok := where.(*FilterOperator); !ok {
		t.Fatalf("second pipeline operator = %T, want *FilterOperator", where)
	}
	if _, ok := project.(*ProjectOperator); !ok {
		t.Fatalf("third pipeline operator = %T, want *ProjectOperator", project)
	}
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("mv-expand pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMvExpandRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	leadingTo := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand to typeof(string)").(*MvExpandOperator)
	leadingExpression := leadingTo.Expressions().At(0).Element()
	if leadingExpression.Expression() == nil || leadingExpression.Expression().Width() != 0 ||
		!diagnosticsContain(leadingExpression.Expression().ContainedDiagnostics(), "KS006") ||
		leadingExpression.ToTypeOf() == nil || queryGrammarMvExpandTypeKind(leadingExpression) != SyntaxKindStringKeyword {
		t.Fatalf("leading-to recovery = %#v diagnostics:%#v", leadingExpression, leadingExpression.ContainedDiagnostics())
	}

	missingTypeOf := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand Items to").(*MvExpandOperator)
	clause := missingTypeOf.Expressions().At(0).Element().ToTypeOf()
	if clause == nil || clause.TypeOf() == nil || clause.TypeOf().Width() != 0 ||
		!diagnosticsContainMessage(clause.TypeOf().ContainedDiagnostics(), "KS006", "Missing typeof") {
		t.Fatalf("missing typeof recovery = %#v diagnostics:%#v", clause, clause.ContainedDiagnostics())
	}

	missingLimit := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand Items limit").(*MvExpandOperator)
	if missingLimit.RowLimitClause() == nil || missingLimit.RowLimitClause().RowLimit().Width() != 0 ||
		!diagnosticsContainMessage(missingLimit.RowLimitClause().RowLimit().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing limit value recovery = %#v diagnostics:%#v", missingLimit.RowLimitClause(), missingLimit.ContainedDiagnostics())
	}

	tokens := Lex("mv-expand Items, limit 5")
	result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
	if result.Length != 3 {
		t.Fatalf("mv-expand invalid comma prefix length = %d, want 3: %#v", result.Length, result)
	}
	missingBeforeLimit, ok := result.Value.(*MvExpandOperator)
	if !ok || missingBeforeLimit.Expressions().Count() != 2 || missingBeforeLimit.RowLimitClause() != nil {
		t.Fatalf("missing list element before limit = operator:%#v expressions:%#v limit:%#v",
			result.Value, missingBeforeLimit.Expressions(), missingBeforeLimit.RowLimitClause())
	}
	missing := missingBeforeLimit.Expressions().At(1).Element().Expression()
	if missing.Width() != 0 || !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing expression: If the keyword 'limit' is intended to be part of an expression it needs to be bracketted as ['limit'].") {
		t.Fatalf("missing list element = %#v diagnostics:%#v", missing, missing.ContainedDiagnostics())
	}

	validLimit := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand Items limit 5").(*MvExpandOperator)
	if validLimit.Expressions().Count() != 1 || validLimit.RowLimitClause() == nil ||
		validLimit.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("valid mv-expand limit = expressions:%#v limit:%#v", validLimit.Expressions(), validLimit.RowLimitClause())
	}
}

func TestQueryGrammarPublicMvExpandDanglingCommasAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name         string
		text         string
		missingIndex int
		count        int
	}{
		{name: "leading", text: "mv-expand , Items", missingIndex: 0, count: 2},
		{name: "middle", text: "mv-expand Items,, Other", missingIndex: 1, count: 3},
		{name: "trailing", text: "mv-expand Items,", missingIndex: 1, count: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*MvExpandOperator)
			if operator.Expressions().Count() != test.count {
				t.Fatalf("mv-expand comma expression count = %d, want %d", operator.Expressions().Count(), test.count)
			}
			missing := operator.Expressions().At(test.missingIndex).Element().Expression()
			if missing.Width() != 0 || !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing expression") {
				t.Fatalf("mv-expand comma recovery = %#v diagnostics:%#v", missing, missing.ContainedDiagnostics())
			}
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "mv-expand Items | where Enabled", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "mv-expand Items limit 5 | where Enabled", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "mv-expand Items limit 5 Extra", 4)
	withoutEqual := parseQueryGrammarProduction(t, grammar.QueryOperator, "mv-expand kind array Items").(*MvExpandOperator)
	if withoutEqual.Parameters().Count() != 0 || withoutEqual.Expressions().Count() != 3 {
		t.Fatalf("mv-expand parameter without equal = parameters:%d expressions:%d, want 0/3",
			withoutEqual.Parameters().Count(), withoutEqual.Expressions().Count())
	}
	for i, want := range []string{"kind", "array", "Items"} {
		element := withoutEqual.Expressions().At(i)
		reference, ok := element.Element().Expression().(*NameReference)
		if !ok || NameReferenceSimpleName(reference) != want {
			t.Fatalf("mv-expand recovered expression %d = %#v, want %s", i, element.Element().Expression(), want)
		}
		if i < 2 && (element.Separator() == nil || element.Separator().Width() != 0 ||
			!diagnosticsContainMessage(element.Separator().Diagnostics(), "KS005", "Expected: ,")) {
			t.Fatalf("mv-expand recovered separator %d = %#v diagnostics:%#v", i, element.Separator(), element.Separator().Diagnostics())
		}
	}
}

func assertQueryGrammarMvExpandAssignment(t *testing.T, expression *MvExpandExpression, name, value string, typeKind SyntaxKind) {
	t.Helper()
	assignment, ok := expression.Expression().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(assignment.Name()) != name {
		t.Fatalf("mv-expand assignment = %#v, want %s", expression.Expression(), name)
	}
	reference, ok := assignment.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != value {
		t.Fatalf("mv-expand assignment value = %#v, want %s", assignment.Expression(), value)
	}
	clause := expression.ToTypeOf()
	if clause == nil || clause.ToKeyword().Kind() != SyntaxKindToKeyword || clause.TypeOf().Types().Count() != 1 ||
		queryGrammarMvExpandTypeKind(expression) != typeKind {
		t.Fatalf("mv-expand typeof clause = %#v, want %v", clause, typeKind)
	}
}

func queryGrammarMvExpandTypeKind(expression *MvExpandExpression) SyntaxKind {
	typeExpression, ok := expression.ToTypeOf().TypeOf().Types().At(0).Element().(*PrimitiveTypeExpression)
	if !ok || typeExpression.Type() == nil {
		return SyntaxKindNone
	}
	return typeExpression.Type().Kind()
}
