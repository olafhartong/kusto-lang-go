package kustolang

import "testing"

func TestQueryGrammarPublicFindSearchPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	find := parseQueryGrammarProduction(t, grammar.QueryOperator, "find Error").(*FindOperator)
	if find.FindKeyword().Kind() != SyntaxKindFindKeyword || find.Condition().Kind() != SyntaxKindNameReference {
		t.Fatalf("direct find = %#v", find)
	}
	if root := parseQueryGrammarProduction(t, grammar.Expression, "find Error"); root.Kind() != SyntaxKindFindOperator {
		t.Fatalf("root find kind = %v, want FindOperator", root.Kind())
	}
	hiddenFind := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, "find Error")
	if _, ok := hiddenFind.(*FindOperator); !ok {
		t.Fatalf("hidden following-pipe find = %T, want *FindOperator", hiddenFind)
	}

	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			search := parseQueryGrammarProduction(t, production.parser, "search Error").(*SearchOperator)
			if search.SearchKeyword().Kind() != SyntaxKindSearchKeyword || search.Condition().Kind() != SyntaxKindNameReference {
				t.Fatalf("direct search = %#v", search)
			}
		})
	}
	if root := parseQueryGrammarProduction(t, grammar.Expression, "search Error"); root.Kind() != SyntaxKindSearchOperator {
		t.Fatalf("root search kind = %v, want SearchOperator", root.Kind())
	}
	pipe := parseQueryGrammarProduction(t, grammar.Expression, "Events | search Error | take 1")
	takePipe, _ := queryGrammarPipeStage(t, pipe, SyntaxKindTakeOperator)
	searchPipe, node := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindSearchOperator)
	if source, ok := searchPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("search pipe source = %#v, want Events", searchPipe.Expression())
	}
	if _, ok := node.(*SearchOperator); !ok {
		t.Fatalf("search pipe operator = %T, want *SearchOperator", node)
	}
}

func TestQueryGrammarPublicFindFullAST(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`find datascope=hotcache withsource=Source in (Signin*, database("db").Audit, ["Security Events"] | as Match) where Severity > 2 project pack(*), *, User:string, ["Event Id"]:long __projectAway Secret`).(*FindOperator)

	if operator.DataScope() == nil || operator.DataScope().Value().RawText() != "hotcache" {
		t.Fatalf("find datascope = %#v", operator.DataScope())
	}
	assertQueryGrammarFindSearchParameters(t, operator.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "withsource", kind: SyntaxKindNameDeclaration},
	})
	if operator.WhereKeyword() == nil || operator.WhereKeyword().Kind() != SyntaxKindWhereKeyword || operator.WhereKeyword().Width() == 0 {
		t.Fatalf("find where keyword = %#v", operator.WhereKeyword())
	}
	if operator.Condition().Kind() != SyntaxKindGreaterThanExpression {
		t.Fatalf("find condition kind = %v, want GreaterThanExpression", operator.Condition().Kind())
	}

	inClause := operator.InClause()
	if inClause == nil || inClause.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		inClause.CloseParen().Kind() != SyntaxKindCloseParenToken || inClause.Expressions().Count() != 3 {
		t.Fatalf("find in clause = %#v", inClause)
	}
	if inClause.Expressions().At(0).Element().Kind() != SyntaxKindNameReference {
		t.Fatalf("wildcard find operand = %#v", inClause.Expressions().At(0).Element())
	}
	qualified, ok := inClause.Expressions().At(1).Element().(*PathExpression)
	if !ok || qualified.Expression().Kind() != SyntaxKindFunctionCallExpression || qualified.Selector().Kind() != SyntaxKindNameReference {
		t.Fatalf("function/path find operand = %#v", inClause.Expressions().At(1).Element())
	}
	aliased, ok := inClause.Expressions().At(2).Element().(*PipeExpression)
	if !ok || aliased.Expression().Kind() != SyntaxKindNameReference || aliased.Operator().Kind() != SyntaxKindAsOperator {
		t.Fatalf("entity/as find operand = %#v", inClause.Expressions().At(2).Element())
	}

	project := operator.Project()
	if project == nil || project.ProjectKeyword().Kind() != SyntaxKindProjectKeyword || project.Columns().Count() != 4 {
		t.Fatalf("find project = %#v", project)
	}
	wantKinds := []SyntaxKind{SyntaxKindPackExpression, SyntaxKindStarExpression, SyntaxKindTypedColumnReference, SyntaxKindTypedColumnReference}
	for i, want := range wantKinds {
		if got := project.Columns().At(i).Element().Kind(); got != want {
			t.Fatalf("find project column %d kind = %v, want %v", i, got, want)
		}
	}
	away := operator.ProjectAway()
	if away == nil || away.ProjectKeyword().Kind() != SyntaxKind_ProjectAwayKeyword || away.Columns().Count() != 1 {
		t.Fatalf("find project-away = %#v", away)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("complete find diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicFindProjectSmart(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"find Error project-smart").(*FindOperator)
	project := operator.Project()
	if project == nil || project.ProjectKeyword().Kind() != SyntaxKindProjectSmartKeyword || project.Columns().Count() != 0 {
		t.Fatalf("find project-smart = %#v", project)
	}
	if operator.ProjectAway() != nil || len(operator.ContainedDiagnostics()) != 0 {
		t.Fatalf("find project-smart tail/diagnostics = away:%#v diagnostics:%#v", operator.ProjectAway(), operator.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicSearchFullASTAndWildcardPredicate(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`search kind=case_sensitive datascope=all in (SigninLogs, Device*, database("db").Audit) * and ResultType != 0`).(*SearchOperator)

	assertQueryGrammarFindSearchParameters(t, operator.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
	})
	if operator.DataScope() == nil || operator.DataScope().Value().RawText() != "all" {
		t.Fatalf("search datascope = %#v", operator.DataScope())
	}
	if operator.InClause() == nil || operator.InClause().Expressions().Count() != 3 {
		t.Fatalf("search in clause = %#v", operator.InClause())
	}
	predicate, ok := operator.Condition().(*BinaryExpression)
	if !ok || predicate.Kind() != SyntaxKindAndExpression || predicate.Left().Kind() != SyntaxKindStarExpression ||
		predicate.Right().Kind() != SyntaxKindNotEqualExpression {
		t.Fatalf("search wildcard predicate = %#v", operator.Condition())
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("complete search diagnostics = %#v", diagnostics)
	}

	star := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "search *").(*SearchOperator)
	if star.Condition().Kind() != SyntaxKindStarExpression || len(star.ContainedDiagnostics()) != 0 {
		t.Fatalf("bare search star = %#v diagnostics:%#v", star.Condition(), star.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicFindSearchCompletionAndTags(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name       string
		parser     ResultParser[LexicalToken, QueryOperator]
		completion string
		tag        string
	}{
		{name: "find query operator", parser: grammar.QueryOperator, completion: "find", tag: "<find>"},
		{name: "search query operator", parser: grammar.QueryOperator, completion: "search", tag: "<search>"},
		{name: "search following pipe", parser: grammar.FollowingPipeElementExpression, completion: "search", tag: "<search>"},
	} {
		t.Run(production.name, func(t *testing.T) {
			items := queryGrammarFindSearchCompletionItems(production.parser, "")
			item, ok := completionItemByDisplayText(items, production.completion)
			if !ok || item.Kind != CompletionKindQueryPrefix {
				t.Fatalf("%s completion = %#v, visible items:%#v", production.completion, item, completionDisplayTexts(items))
			}
			if !queryGrammarParserTreeHasTag(production.parser, production.tag, 0) {
				t.Fatalf("parser tree has no %s tag", production.tag)
			}
		})
	}
	if _, ok := completionItemByDisplayText(queryGrammarFindSearchCompletionItems(grammar.FollowingPipeElementExpression, ""), "find"); ok {
		t.Fatal("post-pipe completions unexpectedly include find")
	}

	for _, value := range []string{"default", "case_insensitive", "case_sensitive"} {
		item, ok := completionItemByDisplayText(queryGrammarFindSearchCompletionItems(grammar.QueryOperator, "search kind="), value)
		if !ok || item.Kind != CompletionKindSyntax {
			t.Fatalf("search kind completion %q = %#v", value, item)
		}
	}
	for _, value := range []string{"all", "hotcache"} {
		item, ok := completionItemByDisplayText(queryGrammarFindSearchCompletionItems(grammar.QueryOperator, "find datascope="), value)
		if !ok || (item.Kind != CompletionKindSyntax && item.Kind != CompletionKindKeyword) {
			t.Fatalf("find datascope completion %q = %#v", value, item)
		}
	}
}

func TestQueryGrammarPublicFindSearchRequiredClauseRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	find := parseQueryGrammarProduction(t, grammar.QueryOperator, "find withsource=Source in (Logs) Error").(*FindOperator)
	if find.WhereKeyword() == nil || find.WhereKeyword().Width() != 0 ||
		!diagnosticsContainMessage(find.WhereKeyword().Diagnostics(), "KS005", "Expected: where") {
		t.Fatalf("implicit required find where = %#v diagnostics:%#v", find.WhereKeyword(), find.ContainedDiagnostics())
	}
	if find.Condition().Kind() != SyntaxKindNameReference {
		t.Fatalf("recovered find condition = %#v", find.Condition())
	}

	missingFind := parseQueryGrammarProduction(t, grammar.QueryOperator, "find where").(*FindOperator)
	if missingFind.Condition().Width() != 0 || !diagnosticsContainMessage(missingFind.Condition().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing find condition = %#v diagnostics:%#v", missingFind.Condition(), missingFind.ContainedDiagnostics())
	}
	missingSearch := parseQueryGrammarProduction(t, grammar.QueryOperator, "search").(*SearchOperator)
	if missingSearch.Condition().Width() != 0 || !diagnosticsContainMessage(missingSearch.Condition().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing search condition = %#v diagnostics:%#v", missingSearch.Condition(), missingSearch.ContainedDiagnostics())
	}
	starAnd := parseQueryGrammarProduction(t, grammar.QueryOperator, "search * and").(*SearchOperator)
	binary, ok := starAnd.Condition().(*BinaryExpression)
	if !ok || binary.Right().Width() != 0 || !diagnosticsContainMessage(binary.Right().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing search star-and RHS = %#v diagnostics:%#v", starAnd.Condition(), starAnd.ContainedDiagnostics())
	}

	missingEqual := parseQueryGrammarProduction(t, grammar.QueryOperator, "search kind= Error").(*SearchOperator)
	if missingEqual.Parameters().Count() != 1 || missingEqual.Parameters().At(0).Expression().Width() == 0 {
		t.Fatalf("search parameter fallback = %#v", missingEqual.Parameters())
	}
	missingValue := parseQueryGrammarProduction(t, grammar.QueryOperator, "search kind=").(*SearchOperator)
	if missingValue.Parameters().Count() != 1 {
		t.Fatalf("missing search parameter count = %d, want 1", missingValue.Parameters().Count())
	}
	value := missingValue.Parameters().At(0).Expression()
	if value.Width() != 0 || !diagnosticsContainMessage(value.Diagnostics(), "KS005", "Expected one of: default, case_insensitive, case_sensitive") {
		t.Fatalf("missing search parameter value = %#v diagnostics:%#v", missingValue.Parameters(), missingValue.ContainedDiagnostics())
	}
	if missingValue.Condition().Width() != 0 || !diagnosticsContainMessage(missingValue.Condition().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing search condition after parameter = %#v diagnostics:%#v", missingValue.Condition(), missingValue.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicFindSearchInClauseRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, prefix := range []string{"find in ", "search in "} {
		t.Run(prefix, func(t *testing.T) {
			var clause *FindInClause
			if prefix[0] == 'f' {
				clause = parseQueryGrammarProduction(t, grammar.QueryOperator, prefix+"(,Logs,,) where Error").(*FindOperator).InClause()
			} else {
				clause = parseQueryGrammarProduction(t, grammar.QueryOperator, prefix+"(,Logs,,) Error").(*SearchOperator).InClause()
			}
			if clause == nil || clause.Expressions().Count() != 4 {
				t.Fatalf("comma-recovered in clause = %#v", clause)
			}
			for _, index := range []int{0, 2, 3} {
				missing := clause.Expressions().At(index).Element()
				if missing.Width() != 0 || !diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing expression") {
					t.Fatalf("in-clause missing operand %d = %#v diagnostics:%#v", index, missing, missing.Diagnostics())
				}
			}
		})
	}

	find := parseQueryGrammarProduction(t, grammar.QueryOperator, "find in (Logs Audit) where Error").(*FindOperator)
	assertQueryGrammarFindSearchMissingComma(t, find.InClause().Expressions(), 0)
	search := parseQueryGrammarProduction(t, grammar.QueryOperator, "search in (Logs Audit) Error").(*SearchOperator)
	assertQueryGrammarFindSearchMissingComma(t, search.InClause().Expressions(), 0)

	missingClose := parseQueryGrammarProduction(t, grammar.QueryOperator, "search in (Logs Error").(*SearchOperator)
	if missingClose.InClause().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.InClause().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing search in close = %#v diagnostics:%#v", missingClose.InClause().CloseParen(), missingClose.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicFindProjectionRecoveryAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	find := parseQueryGrammarProduction(t, grammar.QueryOperator, "find Error project ,User:string,,").(*FindOperator)
	if find.Project().Columns().Count() != 4 {
		t.Fatalf("comma-recovered find project columns = %d, want 4", find.Project().Columns().Count())
	}
	for _, index := range []int{0, 2, 3} {
		missing := find.Project().Columns().At(index).Element()
		if missing.Width() != 0 || !diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing expression") {
			t.Fatalf("find project missing column %d = %#v diagnostics:%#v", index, missing, missing.Diagnostics())
		}
	}
	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator, "find Error project User:string Account:long").(*FindOperator)
	assertQueryGrammarFindSearchMissingComma(t, missingComma.Project().Columns(), 0)

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "find Error extra", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "find Error | take 1", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "search Error project Value", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "search Error, Other", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "find withsource Source", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "search datascope=all Error kind=default", 5)
}

func assertQueryGrammarFindSearchParameters(t *testing.T, parameters *SyntaxListOf[*NamedParameter], want []queryGrammarSharedParameterExpectation) {
	t.Helper()
	assertQueryGrammarSharedParameters(t, parameters, want)
}

func assertQueryGrammarFindSearchMissingComma(t *testing.T, expressions *SeparatedSyntaxListOf[Expression], index int) {
	t.Helper()
	if expressions.Count() != 2 {
		t.Fatalf("missing-comma expression count = %d, want 2", expressions.Count())
	}
	separator := expressions.At(index).Separator()
	if separator == nil || separator.Width() != 0 || !diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing comma = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}
}

func queryGrammarFindSearchCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
