package kustolang

import "testing"

func TestQueryGrammarPublicScanFullAST(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.scanOperator,
		`scan kind=full with_match_id=MatchId order by Timestamp asc nulls last, Sequence desc partition by Tenant, bin(Timestamp, 1h) declare (threshold:long=3, label:string="ready", state:dynamic) with (`+
			`step ["first step"] optional output=all: Value > threshold => ["state.total"] = coalesce(state.total, 0) + 1, ["last value"] = Value; `+
			`step ["final.output"] output=last: state.total >= threshold;`+
			`)`).(*ScanOperator)

	if operator.Kind() != SyntaxKindScanOperator || operator.ScanKeyword().Kind() != SyntaxKindScanKeyword {
		t.Fatalf("scan operator = kind:%v keyword:%v", operator.Kind(), operator.ScanKeyword().Kind())
	}
	if operator.Parameters().Count() != 2 {
		t.Fatalf("scan parameter count = %d, want 2", operator.Parameters().Count())
	}
	for i, want := range []string{"kind", "with_match_id"} {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want || parameter.EqualToken().Width() == 0 || parameter.Expression().Width() == 0 {
			t.Fatalf("scan parameter %d = %#v, want %q with value", i, parameter, want)
		}
	}

	order := operator.OrderByClause()
	if order == nil || order.Kind() != SyntaxKindScanOrderByClause || order.OrderKeyword().Kind() != SyntaxKindOrderKeyword ||
		order.ByKeyword().Kind() != SyntaxKindByKeyword || order.Expressions().Count() != 2 {
		t.Fatalf("scan order-by = %#v", order)
	}
	for i := 0; i < order.Expressions().Count(); i++ {
		if order.Expressions().At(i).Element().Kind() != SyntaxKindOrderedExpression {
			t.Fatalf("scan ordered expression %d kind = %v", i, order.Expressions().At(i).Element().Kind())
		}
	}
	partition := operator.PartitionByClause()
	if partition == nil || partition.Kind() != SyntaxKindScanPartitionByClause ||
		partition.PartitionKeyword().Kind() != SyntaxKindPartitionKeyword || partition.ByKeyword().Kind() != SyntaxKindByKeyword ||
		partition.Expressions().Count() != 2 || partition.Expressions().At(0).Element().Kind() != SyntaxKindNameReference ||
		partition.Expressions().At(1).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("scan partition-by = %#v", partition)
	}

	declare := operator.DeclareClause()
	if declare == nil || declare.Kind() != SyntaxKindScanDeclareClause || declare.DeclareKeyword().Kind() != SyntaxKindDeclareKeyword ||
		declare.OpenParen().Kind() != SyntaxKindOpenParenToken || declare.CloseParen().Kind() != SyntaxKindCloseParenToken ||
		declare.Declarations().Count() != 3 {
		t.Fatalf("scan declare = %#v", declare)
	}
	for i, wantDefault := range []bool{true, true, false} {
		parameter := declare.Declarations().At(i).Element()
		if (parameter.DefaultValue() != nil) != wantDefault {
			t.Fatalf("scan declaration %d default = %#v, want present:%t", i, parameter.DefaultValue(), wantDefault)
		}
	}
	if operator.WithKeyword().Kind() != SyntaxKindWithKeyword || operator.OpenParenToken().Kind() != SyntaxKindOpenParenToken ||
		operator.CloseParenToken().Kind() != SyntaxKindCloseParenToken || operator.Steps().Count() != 2 {
		t.Fatalf("scan body delimiters/steps = with:%#v open:%#v close:%#v count:%d", operator.WithKeyword(),
			operator.OpenParenToken(), operator.CloseParenToken(), operator.Steps().Count())
	}

	first := operator.Steps().At(0)
	if first.Kind() != SyntaxKindScanStep || first.StepKeyword().Kind() != SyntaxKindStepKeyword ||
		NameDeclarationSimpleName(first.Name()) != "first step" || first.OptionalKeyword() == nil ||
		first.OptionalKeyword().Kind() != SyntaxKindOptionalKeyword || first.ColonToken().Kind() != SyntaxKindColonToken ||
		first.Condition().Kind() != SyntaxKindGreaterThanExpression || first.SemicolonToken().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("first scan step = %#v", first)
	}
	output := first.ScanStepOutput()
	if output == nil || output.Kind() != SyntaxKindScanStepOutput || output.OutputKeyword().Kind() != SyntaxKindOutputKeyword ||
		output.EqualToken().Kind() != SyntaxKindEqualToken || output.OutputKind().RawText() != "all" {
		t.Fatalf("first scan output = %#v", output)
	}
	computation := first.ComputationClause()
	if computation == nil || computation.Kind() != SyntaxKindScanComputationClause ||
		computation.ArrowToken().Kind() != SyntaxKindFatArrowToken || computation.Assignments().Count() != 2 {
		t.Fatalf("first scan computation = %#v", computation)
	}
	for i, want := range []string{"state.total", "last value"} {
		assignment := computation.Assignments().At(i).Element()
		if assignment.Kind() != SyntaxKindScanAssignment || NameReferenceSimpleName(assignment.Name()) != want ||
			assignment.EqualToken().Kind() != SyntaxKindEqualToken || assignment.Expression().Width() == 0 {
			t.Fatalf("scan assignment %d = %#v, want name %q", i, assignment, want)
		}
	}
	if computation.Assignments().At(0).Separator() == nil || computation.Assignments().At(0).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("scan assignment separator = %#v", computation.Assignments().At(0).Separator())
	}

	second := operator.Steps().At(1)
	if NameDeclarationSimpleName(second.Name()) != "final.output" || second.OptionalKeyword() != nil ||
		second.ScanStepOutput() == nil || second.ScanStepOutput().OutputKind().RawText() != "last" ||
		second.Condition().Kind() != SyntaxKindGreaterThanOrEqualExpression || second.ComputationClause() != nil {
		t.Fatalf("second scan step = %#v", second)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid full scan diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicScanMinimalZeroAndMultipleSteps(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	zero := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with ()").(*ScanOperator)
	if zero.Parameters().Count() != 0 || zero.OrderByClause() != nil || zero.PartitionByClause() != nil ||
		zero.DeclareClause() != nil || zero.Steps().Count() != 0 || len(zero.ContainedDiagnostics()) != 0 {
		t.Fatalf("minimal scan = %#v diagnostics:%#v", zero, zero.ContainedDiagnostics())
	}

	multiple := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan declare () with (step a: true; step b optional output=none: false => x=1; step c: x == 1;)").(*ScanOperator)
	if multiple.DeclareClause() == nil || multiple.DeclareClause().Declarations().Count() != 0 || multiple.Steps().Count() != 3 {
		t.Fatalf("multi-step scan = %#v", multiple)
	}
	if multiple.Steps().At(0).ScanStepOutput() != nil || multiple.Steps().At(1).OptionalKeyword() == nil ||
		multiple.Steps().At(1).ScanStepOutput().OutputKind().RawText() != "none" ||
		multiple.Steps().At(1).ComputationClause().Assignments().Count() != 1 {
		t.Fatalf("multi-step optional fields = %#v", multiple.Steps())
	}
	if diagnostics := multiple.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid multi-step scan diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicScanPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "private", parser: grammar.scanOperator},
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe", parser: grammar.FollowingPipeElementExpression},
		{name: "fork pipe", parser: grammar.forkPipeOperator},
	} {
		t.Run(production.name, func(t *testing.T) {
			if parsed := parseQueryGrammarProduction(t, production.parser, "scan with ()"); parsed.Kind() != SyntaxKindScanOperator {
				t.Fatalf("scan placement kind = %v", parsed.Kind())
			}
		})
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "Events | scan with (step found: Value > 0;) | where found")
	wherePipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindFilterOperator)
	scanPipe, scan := queryGrammarPipeStage(t, wherePipe.Expression(), SyntaxKindScanOperator)
	if source, ok := scanPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("scan pipeline source = %#v", scanPipe.Expression())
	}
	if scan.(*ScanOperator).Steps().Count() != 1 || len(pipeline.ContainedDiagnostics()) != 0 {
		t.Fatalf("scan pipeline = %#v diagnostics:%#v", scan, pipeline.ContainedDiagnostics())
	}

	fork := parseQueryGrammarProduction(t, grammar.forkExpression, "(scan with (step s: true;))")
	if fork.Expression().Kind() != SyntaxKindScanOperator {
		t.Fatalf("fork scan kind = %v", fork.Expression().Kind())
	}
}

func TestQueryGrammarPublicScanCompletionAndAssignmentTag(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	items := queryGrammarScanCompletionItems(grammar.QueryOperator, "")
	scan, ok := completionItemByDisplayText(items, "scan")
	if !ok || scan.Kind != CompletionKindQueryPrefix {
		t.Fatalf("scan completion = %#v, visible:%#v", scan, completionDisplayTexts(items))
	}

	withItems := queryGrammarScanCompletionItems(grammar.scanOperator, "scan ")
	with, ok := completionItemByDisplayText(withItems, "with")
	if !ok || with.Kind != CompletionKindSyntax || with.BeforeText() != "with (" || with.AfterText() != ")" {
		t.Fatalf("scan with completion = %#v, visible:%#v", with, completionDisplayTexts(withItems))
	}
	if !queryGrammarParserTreeHasTag(grammar.scanOperator, "<assignment>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.QueryOperator, "<assignment>", 0) {
		t.Fatal("scan parser graph is missing the assignment tag")
	}
}

func TestQueryGrammarPublicScanRequiredRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingWith := parseQueryGrammarProduction(t, grammar.scanOperator, "scan ()").(*ScanOperator)
	assertQueryGrammarScanMissingToken(t, missingWith.WithKeyword(), "Expected: with")

	missingOpen := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with step s: true;)").(*ScanOperator)
	assertQueryGrammarScanMissingToken(t, missingOpen.OpenParenToken(), "Expected: (")
	if missingOpen.Steps().Count() != 1 {
		t.Fatalf("missing-open scan steps = %d, want 1", missingOpen.Steps().Count())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with (step s: true;").(*ScanOperator)
	assertQueryGrammarScanMissingToken(t, missingClose.CloseParenToken(), "Expected: )")

	missingName := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with (step : true;)").(*ScanOperator).Steps().At(0)
	if missingName.Name().Width() != 0 || !diagnosticsContainMessage(missingName.Name().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing scan step name = %#v diagnostics:%#v", missingName.Name(), missingName.ContainedDiagnostics())
	}

	missingColon := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with (step s true;)").(*ScanOperator).Steps().At(0)
	assertQueryGrammarScanMissingToken(t, missingColon.ColonToken(), "Expected: :")
	if missingColon.Condition().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("missing-colon predicate = %#v", missingColon.Condition())
	}

	missingPredicate := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with (step s: ;)").(*ScanOperator).Steps().At(0)
	assertQueryGrammarScanMissingExpression(t, missingPredicate.Condition())

	missingSemicolon := parseQueryGrammarProduction(t, grammar.scanOperator, "scan with (step s: true)").(*ScanOperator).Steps().At(0)
	assertQueryGrammarScanMissingToken(t, missingSemicolon.SemicolonToken(), "Expected: ;")
}

func TestQueryGrammarPublicScanDeclareAndClauseRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingDeclareOpen := parseQueryGrammarProduction(t, grammar.scanOperator, "scan declare x:long) with ()").(*ScanOperator).DeclareClause()
	assertQueryGrammarScanMissingToken(t, missingDeclareOpen.OpenParen(), "Expected: (")
	if missingDeclareOpen.Declarations().Count() != 1 {
		t.Fatalf("missing declare open declarations = %d", missingDeclareOpen.Declarations().Count())
	}

	missingDeclareClose := parseQueryGrammarProduction(t, grammar.scanOperator, "scan declare (x:long with ()").(*ScanOperator).DeclareClause()
	assertQueryGrammarScanMissingToken(t, missingDeclareClose.CloseParen(), "Expected: )")

	for _, test := range []struct {
		name    string
		text    string
		message string
	}{
		{name: "declaration name", text: "scan declare (:long) with ()", message: "Missing name"},
		{name: "declaration colon", text: "scan declare (x long) with ()", message: "Expected: :"},
		{name: "declaration type", text: "scan declare (x:) with ()", message: "Missing type name"},
		{name: "declaration comma", text: "scan declare (x:long y:string) with ()", message: "Expected: ,"},
		{name: "declaration trailing comma", text: "scan declare (x:long,) with ()", message: "Missing parameter"},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.scanOperator, test.text).(*ScanOperator)
			if !diagnosticsContainMessage(operator.DeclareClause().ContainedDiagnostics(), "KS005", test.message) &&
				!diagnosticsContainMessage(operator.DeclareClause().ContainedDiagnostics(), "KS006", test.message) {
				t.Fatalf("scan declare diagnostics = %#v, want %q", operator.DeclareClause().ContainedDiagnostics(), test.message)
			}
		})
	}

	missingOrderBy := parseQueryGrammarProduction(t, grammar.scanOperator, "scan order Timestamp with ()").(*ScanOperator)
	assertQueryGrammarScanMissingToken(t, missingOrderBy.OrderByClause().ByKeyword(), "Expected: by")
	missingPartitionBy := parseQueryGrammarProduction(t, grammar.scanOperator, "scan partition Tenant with ()").(*ScanOperator)
	assertQueryGrammarScanMissingToken(t, missingPartitionBy.PartitionByClause().ByKeyword(), "Expected: by")

	for _, test := range []struct {
		name string
		text string
		get  func(*ScanOperator) *SeparatedSyntaxListOf[Expression]
	}{
		{name: "empty order", text: "scan order by", get: func(o *ScanOperator) *SeparatedSyntaxListOf[Expression] { return o.OrderByClause().Expressions() }},
		{name: "empty partition", text: "scan partition by", get: func(o *ScanOperator) *SeparatedSyntaxListOf[Expression] { return o.PartitionByClause().Expressions() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			list := test.get(parseQueryGrammarProduction(t, grammar.scanOperator, test.text).(*ScanOperator))
			if list.Count() < 1 {
				t.Fatalf("missing clause list count = %d, want at least 1", list.Count())
			}
			assertQueryGrammarScanMissingExpression(t, list.At(0).Element())
		})
	}
}

func TestQueryGrammarPublicScanOutputAndAssignmentRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingOutputEqual := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan with (step s output all: true;)").(*ScanOperator).Steps().At(0).ScanStepOutput()
	assertQueryGrammarScanMissingToken(t, missingOutputEqual.EqualToken(), "Expected: =")

	missingOutputKind := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan with (step s output=: true;)").(*ScanOperator).Steps().At(0).ScanStepOutput()
	if missingOutputKind.OutputKind().Width() != 0 || !diagnosticsContain(missingOutputKind.OutputKind().Diagnostics(), "KS005") {
		t.Fatalf("missing scan output kind = %#v diagnostics:%#v", missingOutputKind.OutputKind(), missingOutputKind.ContainedDiagnostics())
	}

	missingAssignments := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan with (step s: true => ;)").(*ScanOperator).Steps().At(0).ComputationClause().Assignments()
	if missingAssignments.Count() != 1 {
		t.Fatalf("missing assignment count = %d, want 1", missingAssignments.Count())
	}
	assertQueryGrammarScanMissingAssignment(t, missingAssignments.At(0).Element())

	missingNameResult := grammar.scanOperator.ParseValue(NewArraySource(Lex("scan with (step s: true => =1;)")), 0)
	if missingNameResult.Length != 8 {
		t.Fatalf("missing assignment name recovery = %#v, want prefix length 8", missingNameResult)
	}
	missingAssignmentName := missingNameResult.Value.(*ScanOperator).Steps().At(0).ComputationClause().Assignments().At(0).Element()
	if !diagnosticsContainMessage(missingAssignmentName.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing assignment name = %#v diagnostics:%#v", missingAssignmentName, missingAssignmentName.ContainedDiagnostics())
	}

	missingEqualResult := grammar.scanOperator.ParseValue(NewArraySource(Lex("scan with (step s: true => total 1;)")), 0)
	if missingEqualResult.Length != 8 {
		t.Fatalf("missing assignment equals recovery = %#v, want prefix length 8", missingEqualResult)
	}
	assertQueryGrammarScanMissingAssignment(t,
		missingEqualResult.Value.(*ScanOperator).Steps().At(0).ComputationClause().Assignments().At(0).Element())

	missingAssignmentExpression := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan with (step s: true => total=;)").(*ScanOperator).Steps().At(0).ComputationClause().Assignments().At(0).Element()
	assertQueryGrammarScanMissingExpression(t, missingAssignmentExpression.Expression())
}

func TestQueryGrammarPublicScanCommaAndClauseBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	order := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan order by Timestamp Sequence partition by Tenant declare (x:long) with ()").(*ScanOperator)
	if order.OrderByClause().Expressions().Count() != 2 || order.PartitionByClause() == nil || order.DeclareClause() == nil {
		t.Fatalf("order missing-comma clause boundaries = %#v", order)
	}
	assertQueryGrammarScanMissingToken(t, order.OrderByClause().Expressions().At(0).Separator(), "Expected: ,")

	partition := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan partition by Tenant Region declare (x:long) with ()").(*ScanOperator)
	if partition.PartitionByClause().Expressions().Count() != 2 || partition.DeclareClause() == nil {
		t.Fatalf("partition missing-comma clause boundaries = %#v", partition)
	}
	assertQueryGrammarScanMissingToken(t, partition.PartitionByClause().Expressions().At(0).Separator(), "Expected: ,")

	assignments := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan with (step s: true => a=1 b=2;)").(*ScanOperator).Steps().At(0).ComputationClause().Assignments()
	if assignments.Count() != 2 {
		t.Fatalf("assignment missing-comma count = %d, want 2", assignments.Count())
	}
	assertQueryGrammarScanMissingToken(t, assignments.At(0).Separator(), "Expected: ,")

	assertQueryGrammarProductionLeavesTail(t, grammar.scanOperator,
		"scan order by Timestamp, partition by Tenant with ()", 6)
	committedPartitionComma := parseQueryGrammarProduction(t, grammar.scanOperator,
		"scan partition by Tenant, declare (x:long) with ()").(*ScanOperator)
	if committedPartitionComma.PartitionByClause().Expressions().Count() < 2 ||
		committedPartitionComma.DeclareClause() != nil || len(committedPartitionComma.ContainedDiagnostics()) == 0 {
		t.Fatalf("committed partition comma boundary = %#v diagnostics:%#v", committedPartitionComma,
			committedPartitionComma.ContainedDiagnostics())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.scanOperator, "scan with (), Extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.scanOperator, "scan with () | where true", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.scanOperator, "scan with () trailing", 4)
}

func TestQueryGrammarPublicScanIntentionallyIncorrectSyntax(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		"with ()",
		"step s: true;",
		"scan",
		"scan with",
		"scan with (s: true;)",
		"scan with (step s true;)",
		"scan with (step s: true)",
		"scan with (step s: true =>;)",
		"scan declare (x) with ()",
		"scan order Timestamp with ()",
		"scan partition Tenant with ()",
	} {
		t.Run(text, func(t *testing.T) {
			result := grammar.scanOperator.ParseValue(NewArraySource(Lex(text)), 0)
			if result.Length < 0 {
				if text == "with ()" || text == "step s: true;" {
					return
				}
				t.Fatalf("recoverable incorrect scan rejected: %#v", result)
			}
			if text == "with ()" || text == "step s: true;" {
				t.Fatalf("non-scan syntax unexpectedly parsed: %#v", result)
			}
			if result.Value == nil || len(result.Value.ContainedDiagnostics()) == 0 {
				t.Fatalf("incorrect scan %q has no diagnostics: %#v", text, result.Value)
			}
		})
	}
}

func queryGrammarScanCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}

func assertQueryGrammarScanMissingToken(t *testing.T, token *SyntaxToken, message string) {
	t.Helper()
	if token == nil || token.Width() != 0 || !diagnosticsContainMessage(token.Diagnostics(), "KS005", message) {
		t.Fatalf("missing scan token = %#v diagnostics:%#v, want %q", token, tokenDiagnostics(token), message)
	}
}

func assertQueryGrammarScanMissingExpression(t *testing.T, expression Expression) {
	t.Helper()
	if expression == nil || expression.Width() != 0 || !diagnosticsContainMessage(expression.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing scan expression = %#v diagnostics:%#v", expression, expressionDiagnostics(expression))
	}
}

func assertQueryGrammarScanMissingAssignment(t *testing.T, assignment *ScanAssignment) {
	t.Helper()
	if assignment == nil || assignment.Width() != 0 || assignment.Name().Width() != 0 || assignment.EqualToken().Width() != 0 ||
		assignment.EqualToken().Kind() != SyntaxKindEqualEqualToken || assignment.Expression().Width() != 0 ||
		!diagnosticsContainMessage(assignment.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing scan assignment = %#v diagnostics:%#v", assignment, assignment.ContainedDiagnostics())
	}
}

func tokenDiagnostics(token *SyntaxToken) []Diagnostic {
	if token == nil {
		return nil
	}
	return token.Diagnostics()
}

func expressionDiagnostics(expression Expression) []Diagnostic {
	if expression == nil {
		return nil
	}
	return expression.ContainedDiagnostics()
}
