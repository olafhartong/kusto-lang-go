package kustolang

import "testing"

func TestQueryGrammarPublicMacroExpandCompletionTagAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	if got := grammar.macroExpandOperator.Tag(); got != "<macro-expand>" {
		t.Fatalf("macro-expand parser tag = %q, want <macro-expand>", got)
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<macro-expand>", 0) {
		t.Fatal("QueryOperator parser tree has no <macro-expand> tag")
	}

	for _, parser := range []ResultParser[LexicalToken, QueryOperator]{grammar.macroExpandOperator, grammar.QueryOperator} {
		operator := parseQueryGrammarProduction(t, parser, "macro-expand Targets as Scope (print x=1)")
		if _, ok := operator.(*MacroExpandOperator); !ok {
			t.Fatalf("macro-expand production = %T, want *MacroExpandOperator", operator)
		}
	}
	if _, ok := parseQueryGrammarProduction(t, grammar.Expression, "macro-expand Targets as Scope (print x=1)").(*MacroExpandOperator); !ok {
		t.Fatal("root macro-expand did not produce MacroExpandOperator")
	}

	piped := parseQueryGrammarProduction(t, grammar.Expression, "Source | macro-expand Targets as Scope (print x=1)")
	pipe, operator := queryGrammarPipeStage(t, piped, SyntaxKindMacroExpandOperator)
	if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Source" {
		t.Fatalf("piped macro-expand source = %#v, want Source", pipe.Expression())
	}
	if _, ok := operator.(*MacroExpandOperator); !ok {
		t.Fatalf("piped macro-expand operator = %T", operator)
	}

	for name, parser := range map[string]Parser[LexicalToken]{
		"private":        grammar.macroExpandOperator,
		"query operator": grammar.QueryOperator,
	} {
		item, ok := completionItemByDisplayText(queryGrammarMacroExpandCompletionItems(parser), "macro-expand")
		if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityHigh {
			t.Fatalf("%s completion = %#v", name, item)
		}
	}
	if _, ok := completionItemByDisplayText(queryGrammarMacroExpandCompletionItems(grammar.FollowingPipeElementExpression), "macro-expand"); ok {
		t.Fatal("hidden piped macro-expand exposed a completion")
	}
	if _, ok := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, "macro-expand Targets as Scope (print x=1)").(*MacroExpandOperator); !ok {
		t.Fatal("hidden piped production did not parse macro-expand")
	}
}

func TestQueryGrammarPublicMacroExpandParametersDeclaredOrKnown(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).macroExpandOperator,
		"macro-expand force_remote=true isfuzzy=false best_effort=true hint.concurrency=4 Targets as Scope (print x=1)").(*MacroExpandOperator)

	want := []struct {
		name string
		kind SyntaxKind
	}{
		{name: "force_remote", kind: SyntaxKindBooleanLiteralExpression},
		{name: "isfuzzy", kind: SyntaxKindBooleanLiteralExpression},
		{name: "best_effort", kind: SyntaxKindBooleanLiteralExpression},
		{name: "hint.concurrency", kind: SyntaxKindLongLiteralExpression},
	}
	if operator.Parameters().Count() != len(want) {
		t.Fatalf("macro-expand parameters = %d, want %d", operator.Parameters().Count(), len(want))
	}
	for i, expected := range want {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != expected.name || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != expected.kind {
			t.Fatalf("parameter %d = %#v, want %q/%v", i, parameter, expected.name, expected.kind)
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid macro-expand diagnostics = %#v", diagnostics)
	}

	fallback := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).macroExpandOperator,
		"macro-expand force_remote=perhaps Targets as Scope (print x=1)").(*MacroExpandOperator).Parameters().At(0)
	if fallback.Expression().Kind() != SyntaxKindTokenLiteralExpression || len(fallback.ContainedDiagnostics()) != 0 {
		t.Fatalf("declared parameter fallback = %#v diagnostics:%#v", fallback.Expression(), fallback.ContainedDiagnostics())
	}

	// equalsNeeded=true does not recover a parameter when '=' is absent. Since
	// force_remote is tokenized as a parameter name rather than a target name,
	// the whole operator production rejects this incomplete prefix.
	assertQueryGrammarProductionRejects(t, QueryGrammarFrom(nil).macroExpandOperator,
		"macro-expand force_remote true")
}

func TestQueryGrammarPublicMacroExpandTargetOrderingAndShapes(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	literal := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		`macro-expand entity_group[cluster("c").database("d").T, database("d").U] (print x=1)`).(*MacroExpandOperator)
	group, ok := literal.EntityGroup().(*EntityGroup)
	if !ok || group.EntityGroupKeyword().Kind() != SyntaxKindEntityGroupKeyword || group.Entities().Count() != 2 ||
		group.OpenBracket().Width() == 0 || group.CloseBracket().Width() == 0 {
		t.Fatalf("entity_group literal target = %#v", literal.EntityGroup())
	}
	for i := 0; i < group.Entities().Count(); i++ {
		if group.Entities().At(i).Element().Kind() != SyntaxKindPathExpression {
			t.Fatalf("entity_group entity %d = %T/%v, want path", i, group.Entities().At(i).Element(), group.Entities().At(i).Element().Kind())
		}
	}

	for _, text := range []string{
		`macro-expand cluster("c").database("d").Targets as Scope (print x=1)`,
		`macro-expand database("d").Targets as Scope (print x=1)`,
	} {
		operator := parseQueryGrammarProduction(t, grammar.macroExpandOperator, text).(*MacroExpandOperator)
		path, ok := operator.EntityGroup().(*PathExpression)
		if !ok || path.Kind() != SyntaxKindPathExpression {
			t.Fatalf("qualified target %q = %#v", text, operator.EntityGroup())
		}
	}

	call := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand ResolveTargets() (print x=1)").(*MacroExpandOperator).EntityGroup()
	if function, ok := call.(*FunctionCallExpression); !ok || NameReferenceSimpleName(function.Name()) != "ResolveTargets" {
		t.Fatalf("direct function target = %#v", call)
	}
	name := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1)").(*MacroExpandOperator).EntityGroup()
	if reference, ok := name.(*NameReference); !ok || NameReferenceSimpleName(reference) != "Targets" {
		t.Fatalf("simple-name target = %#v", name)
	}
}

func TestQueryGrammarPublicMacroExpandRejectsAndBoundsArbitraryTargets(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		"macro-expand",
		"macro-expand 1 (print x=1)",
		`macro-expand "Targets" (print x=1)`,
	} {
		assertQueryGrammarProductionRejects(t, grammar.macroExpandOperator, text)
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator, "macro-expand Targets.Other (print x=1)", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator, "macro-expand ResolveTargets().database(\"d\") (print x=1)", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator, "macro-expand Targets, Other (print x=1)", 2)
}

func TestQueryGrammarPublicMacroExpandScopeIdentifierOnlyAndTail(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1)").(*MacroExpandOperator)
	scope := operator.ScopeReferenceName()
	if scope == nil || scope.AsKeyword().Kind() != SyntaxKindAsKeyword || scope.AsKeyword().Width() == 0 ||
		NameDeclarationSimpleName(scope.EntityGroupReferenceName()) != "Scope" ||
		scope.EntityGroupReferenceName().Name().Kind() != SyntaxKindTokenName {
		t.Fatalf("macro-expand scope = %#v", scope)
	}

	for _, test := range []struct {
		text       string
		wantLength int
	}{
		{text: "macro-expand Targets as (print x=1)", wantLength: 9},
		{text: "macro-expand Targets as [Scope] (print x=1)", wantLength: 5},
		{text: "macro-expand Targets as where (print x=1)", wantLength: 10},
	} {
		result := grammar.macroExpandOperator.ParseValue(NewArraySource(Lex(test.text)), 0)
		macro, ok := result.Value.(*MacroExpandOperator)
		if !ok || result.Length != test.wantLength || macro.ScopeReferenceName() != nil {
			t.Fatalf("identifier-only scope tail %q = %#v, want length %d", test.text, result, test.wantLength)
		}
		if macro.OpenParen().Width() != 0 ||
			!diagnosticsContainMessage(macro.OpenParen().Diagnostics(), "KS005", "Expected: (") {
			t.Fatalf("rolled-back scope tail %q = open:%#v diagnostics:%#v", test.text, macro.OpenParen(), macro.ContainedDiagnostics())
		}
	}
}

func TestQueryGrammarPublicMacroExpandRequiredParentheses(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingOpen := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope print x=1)").(*MacroExpandOperator)
	if missingOpen.OpenParen().Width() != 0 ||
		!diagnosticsContainMessage(missingOpen.OpenParen().Diagnostics(), "KS005", "Expected: (") ||
		missingOpen.StatementList().Count() != 1 || missingOpen.CloseParen().Width() == 0 {
		t.Fatalf("missing open recovery = %#v diagnostics:%#v", missingOpen, missingOpen.ContainedDiagnostics())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1").(*MacroExpandOperator)
	if missingClose.OpenParen().Width() == 0 || missingClose.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.CloseParen().Diagnostics(), "KS005", "Expected: )") ||
		missingClose.StatementList().Count() != 1 {
		t.Fatalf("missing close recovery = %#v diagnostics:%#v", missingClose, missingClose.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicMacroExpandSubQueryStatementLists(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (declare query_parameters (p:long); let n = p; print n; T | take 1)").(*MacroExpandOperator)
	statements := operator.StatementList()
	want := []SyntaxKind{
		SyntaxKindQueryParametersStatement,
		SyntaxKindLetStatement,
		SyntaxKindExpressionStatement,
		SyntaxKindExpressionStatement,
	}
	if statements.Count() != len(want) {
		t.Fatalf("macro-expand statement count = %d, want %d", statements.Count(), len(want))
	}
	for i, kind := range want {
		if statements.At(i).Element().Kind() != kind {
			t.Fatalf("statement %d kind = %v, want %v", i, statements.At(i).Element().Kind(), kind)
		}
		if i < len(want)-1 && (statements.At(i).Separator() == nil || statements.At(i).Separator().Width() == 0) {
			t.Fatalf("statement %d separator = %#v, want present semicolon", i, statements.At(i).Separator())
		}
	}
	last := statements.At(3).Element().(*ExpressionStatement)
	if last.Expression().Kind() != SyntaxKindPipeExpression || len(operator.ContainedDiagnostics()) != 0 {
		t.Fatalf("last statement/diagnostics = %v/%#v", last.Expression().Kind(), operator.ContainedDiagnostics())
	}

	empty := parseQueryGrammarProduction(t, grammar.macroExpandOperator, "macro-expand Targets as Scope ()").(*MacroExpandOperator)
	if empty.StatementList().Count() != 0 || empty.OpenParen().Width() == 0 || empty.CloseParen().Width() == 0 {
		t.Fatalf("empty macro-expand body = %#v", empty)
	}
}

func TestQueryGrammarPublicMacroExpandSubQueryRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	leading := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (; print x=1)").(*MacroExpandOperator)
	if leading.StatementList().Count() != 2 || leading.StatementList().At(0).Separator() == nil ||
		leading.StatementList().At(0).Separator().Width() == 0 ||
		!diagnosticsContainMessage(leading.StatementList().At(0).Element().ContainedDiagnostics(), "KS006", "Missing statement") {
		t.Fatalf("leading-semicolon body recovery = %#v diagnostics:%#v", leading.StatementList(), leading.ContainedDiagnostics())
	}

	missingSeparator := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (let n = 1 print n)").(*MacroExpandOperator)
	if missingSeparator.StatementList().Count() != 2 {
		t.Fatalf("missing-semicolon body count = %d, want 2", missingSeparator.StatementList().Count())
	}
	separator := missingSeparator.StatementList().At(0).Separator()
	if separator == nil || separator.Width() != 0 ||
		!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ;") {
		t.Fatalf("missing body separator = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}

	trailing := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1;)").(*MacroExpandOperator)
	if trailing.StatementList().Count() != 1 || trailing.StatementList().At(0).Separator() == nil ||
		trailing.StatementList().At(0).Separator().Width() == 0 || len(trailing.ContainedDiagnostics()) != 0 {
		t.Fatalf("trailing-semicolon body = %#v diagnostics:%#v", trailing.StatementList(), trailing.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicMacroExpandMalformedTargetBodyAndTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, test := range []struct {
		name         string
		text         string
		count        int
		missingIndex int
	}{
		{name: "leading comma", text: "macro-expand entity_group[, T] (print x=1)", count: 2, missingIndex: 0},
		{name: "middle comma", text: "macro-expand entity_group[T,, U] (print x=1)", count: 3, missingIndex: 1},
		{name: "trailing comma", text: "macro-expand entity_group[T,] (print x=1)", count: 2, missingIndex: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			group := parseQueryGrammarProduction(t, grammar.macroExpandOperator, test.text).(*MacroExpandOperator).EntityGroup().(*EntityGroup)
			if group.Entities().Count() != test.count {
				t.Fatalf("entity count = %d, want %d", group.Entities().Count(), test.count)
			}
			missing := group.Entities().At(test.missingIndex).Element()
			if missing.Width() != 0 || !diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing expression") {
				t.Fatalf("missing entity = %#v diagnostics:%#v", missing, missing.Diagnostics())
			}
		})
	}

	missingBracket := parseQueryGrammarProduction(t, grammar.macroExpandOperator,
		"macro-expand entity_group[T (print x=1)").(*MacroExpandOperator)
	group := missingBracket.EntityGroup().(*EntityGroup)
	if group.CloseBracket().Width() != 0 || !diagnosticsContainMessage(group.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("missing entity_group close bracket = %#v diagnostics:%#v", group.CloseBracket(), group.CloseBracket().Diagnostics())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1) | take 1", 10)
	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator,
		"macro-expand Targets as Scope (print x=1) trailing", 10)
	assertQueryGrammarProductionLeavesTail(t, grammar.macroExpandOperator,
		"macro-expand Targets, Other (print x=1)", 2)
}

func queryGrammarMacroExpandCompletionItems(parser Parser[LexicalToken]) []CompletionItem {
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
