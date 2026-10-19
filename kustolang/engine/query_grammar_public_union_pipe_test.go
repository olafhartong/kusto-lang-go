package kustolang

import "testing"

func TestQueryGrammarPublicUnionRootAndPostPipePlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, parser := range []ResultParser[LexicalToken, QueryOperator]{
		grammar.QueryOperator,
		grammar.FollowingPipeElementExpression,
	} {
		operator := parseQueryGrammarProduction(t, parser, "union SigninLogs").(*UnionOperator)
		if operator.UnionKeyword().Kind() != SyntaxKindUnionKeyword || operator.Expressions().Count() != 1 {
			t.Fatalf("direct union = keyword:%v expressions:%d", operator.UnionKeyword().Kind(), operator.Expressions().Count())
		}
	}

	root := parseQueryGrammarProduction(t, grammar.Expression, "union SigninLogs")
	if _, ok := root.(*UnionOperator); !ok {
		t.Fatalf("root union = %T, want *UnionOperator", root)
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		"Source | union SigninLogs | take 1")
	takePipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindTakeOperator)
	unionPipe, unionNode := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindUnionOperator)
	if source, ok := unionPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Source" {
		t.Fatalf("post-pipe union source = %#v, want Source", unionPipe.Expression())
	}
	if _, ok := unionNode.(*UnionOperator); !ok {
		t.Fatalf("post-pipe union operator = %T, want *UnionOperator", unionNode)
	}
	if diagnostics := pipeline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("union pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicUnionParameters(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"union kind=outer withsource=SourceTable isfuzzy=true best_effort=false hint.concurrency=4 hint.spread=2 SigninLogs").(*UnionOperator)

	want := []struct {
		name string
		kind SyntaxKind
	}{
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
		{name: "withsource", kind: SyntaxKindNameDeclaration},
		{name: "isfuzzy", kind: SyntaxKindBooleanLiteralExpression},
		{name: "best_effort", kind: SyntaxKindBooleanLiteralExpression},
		{name: "hint.concurrency", kind: SyntaxKindLongLiteralExpression},
		{name: "hint.spread", kind: SyntaxKindLongLiteralExpression},
	}
	if operator.Parameters().Count() != len(want) {
		t.Fatalf("union parameters = %d, want %d", operator.Parameters().Count(), len(want))
	}
	for i, expected := range want {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != expected.name ||
			parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != expected.kind {
			t.Fatalf("union parameter %d = %#v, want %q/%v", i, parameter, expected.name, expected.kind)
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid union diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicUnionOperandAlternatives(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`union SigninLogs, ["Security Events"], Device*, [Archive*], MyFunc(), database("db").RemoteFunc(), database("db")["Audit Logs"], database("db")[Parity*]`).(*UnionOperator)

	if operator.Expressions().Count() != 8 {
		t.Fatalf("union expressions = %d, want 8", operator.Expressions().Count())
	}
	assertQueryGrammarUnionName(t, operator.Expressions().At(0).Element(), "SigninLogs", SyntaxKindTokenName)
	assertQueryGrammarUnionName(t, operator.Expressions().At(1).Element(), "Security Events", SyntaxKindBracketedName)
	assertQueryGrammarUnionName(t, operator.Expressions().At(2).Element(), "Device*", SyntaxKindWildcardedName)
	assertQueryGrammarUnionName(t, operator.Expressions().At(3).Element(), "Archive*", SyntaxKindBracketedWildcardedName)

	if expression := operator.Expressions().At(4).Element(); expression.Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("simple function-call operand = %T/%v", expression, expression.Kind())
	}
	qualified, ok := operator.Expressions().At(5).Element().(*PathExpression)
	if !ok || qualified.Expression().Kind() != SyntaxKindFunctionCallExpression ||
		qualified.Selector().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("qualified function-call operand = %#v", operator.Expressions().At(5).Element())
	}
	assertQueryGrammarUnionElementSelector(t, operator.Expressions().At(6).Element(), "Audit Logs", SyntaxKindBracketedName)
	assertQueryGrammarUnionElementSelector(t, operator.Expressions().At(7).Element(), "Parity*", SyntaxKindBracketedWildcardedName)

	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("union operand diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicUnionParenthesizedFullPipelines(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"union (SigninLogs | where ResultType == 0 | project User), (AuditLogs | take 5)").(*UnionOperator)

	if operator.Expressions().Count() != 2 {
		t.Fatalf("parenthesized union expressions = %d, want 2", operator.Expressions().Count())
	}
	for i := 0; i < operator.Expressions().Count(); i++ {
		parenthesized, ok := operator.Expressions().At(i).Element().(*ParenthesizedExpression)
		if !ok || parenthesized.OpenParen().Kind() != SyntaxKindOpenParenToken ||
			parenthesized.Expression().Kind() != SyntaxKindPipeExpression ||
			parenthesized.CloseParen().Kind() != SyntaxKindCloseParenToken {
			t.Fatalf("parenthesized pipeline %d = %#v", i, operator.Expressions().At(i).Element())
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("parenthesized union diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicUnionCompletionAndTag(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parsers := FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.QueryOperator, 0)
	items := []CompletionItem{}
	for _, parser := range parsers {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	item, ok := completionItemByDisplayText(items, "union")
	if !ok || item.Kind != CompletionKindQueryPrefix {
		t.Fatalf("union completion = %#v, visible items:%#v", item, completionDisplayTexts(items))
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<union>", 0) {
		t.Fatal("QueryOperator parser tree has no <union> tag")
	}
}

func TestQueryGrammarPublicUnionParameterCompletionShape(t *testing.T) {
	primary := queryGrammarParameterNameParser(QueryOperatorParameters.WithSource)
	items := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, primary, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	item, ok := completionItemByDisplayText(items, "withsource")
	if !ok || item.BeforeText() != "withsource=" || item.AfterText() != "" || !item.HasCaret() {
		t.Fatalf("withsource completion = %#v, before:%q after:%q", item, item.BeforeText(), item.AfterText())
	}

	hidden := queryGrammarParameterNameParser(QueryOperatorParameters.IsFuzzy.Hide())
	hiddenItems := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, hidden, 0) {
		hiddenItems = append(hiddenItems, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	if len(hiddenItems) != 0 {
		t.Fatalf("hidden parameter exposed completion items: %#v", hiddenItems)
	}
}

func TestQueryGrammarPublicUnionRequiresOneOperand(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "union").(*UnionOperator)
	if operator.Expressions().Count() != 1 {
		t.Fatalf("operand-less union expression count = %d, want 1", operator.Expressions().Count())
	}
	missing := operator.Expressions().At(0).Element()
	if missing.Width() != 0 || !diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing union operand = %#v diagnostics:%#v", missing, missing.Diagnostics())
	}
}

func TestQueryGrammarPublicUnionCommaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name         string
		text         string
		count        int
		missingIndex int
	}{
		{name: "leading", text: "union , SigninLogs", count: 2, missingIndex: 0},
		{name: "middle", text: "union SigninLogs,, AuditLogs", count: 3, missingIndex: 1},
		{name: "trailing", text: "union SigninLogs,", count: 2, missingIndex: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*UnionOperator)
			if operator.Expressions().Count() != test.count {
				t.Fatalf("union comma expression count = %d, want %d", operator.Expressions().Count(), test.count)
			}
			missing := operator.Expressions().At(test.missingIndex).Element()
			if missing.Width() != 0 || !diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing expression") {
				t.Fatalf("union comma recovery = %#v diagnostics:%#v", missing, missing.Diagnostics())
			}
		})
	}

	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator, "union SigninLogs AuditLogs").(*UnionOperator)
	if missingComma.Expressions().Count() != 2 {
		t.Fatalf("missing-comma union expressions = %d, want 2", missingComma.Expressions().Count())
	}
	separator := missingComma.Expressions().At(0).Separator()
	if separator == nil || separator.Width() != 0 ||
		!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing union comma = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}
}

func TestQueryGrammarPublicUnionSelectorAndBracketRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingSelector := parseQueryGrammarProduction(t, grammar.QueryOperator,
		`union database("db").`).(*UnionOperator)
	path, ok := missingSelector.Expressions().At(0).Element().(*PathExpression)
	if !ok {
		t.Fatalf("missing-selector operand = %T, want *PathExpression", missingSelector.Expressions().At(0).Element())
	}
	selector := path.Selector()
	if selector.Width() != 0 || !diagnosticsContainMessage(selector.Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing union selector = %#v diagnostics:%#v", selector, selector.Diagnostics())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.QueryOperator,
		`union database("db")[Parity*`).(*UnionOperator)
	element := missingClose.Expressions().At(0).Element().(*ElementExpression)
	reference := element.Selector().(*NameReference)
	wildcard := reference.Name().(*BracketedWildcardedName)
	if wildcard.CloseBracket().Width() != 0 ||
		!diagnosticsContainMessage(wildcard.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("missing union close bracket = %#v diagnostics:%#v", wildcard.CloseBracket(), wildcard.CloseBracket().Diagnostics())
	}
}

func TestQueryGrammarPublicUnionTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "union SigninLogs | where true", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "union SigninLogs, AuditLogs | where true", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "union SigninLogs + 1", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "union SigninLogs )", 2)
}

func assertQueryGrammarUnionName(t *testing.T, expression Expression, text string, kind SyntaxKind) {
	t.Helper()
	reference, ok := expression.(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != text || reference.Name().Kind() != kind {
		t.Fatalf("union name operand = %#v, want %q/%v", expression, text, kind)
	}
}

func assertQueryGrammarUnionElementSelector(t *testing.T, expression Expression, text string, kind SyntaxKind) {
	t.Helper()
	element, ok := expression.(*ElementExpression)
	if !ok {
		t.Fatalf("union element operand = %T, want *ElementExpression", expression)
	}
	assertQueryGrammarUnionName(t, element.Selector(), text, kind)
}

func queryGrammarParserTreeHasTag(parser Parser[LexicalToken], tag string, depth int) bool {
	return queryGrammarParserGraphHasTag(parser, tag, depth, map[Parser[LexicalToken]]struct{}{})
}

func queryGrammarParserGraphHasTag(parser Parser[LexicalToken], tag string, depth int, seen map[Parser[LexicalToken]]struct{}) bool {
	if parser == nil || depth > 100 {
		return false
	}
	if _, ok := seen[parser]; ok {
		return false
	}
	seen[parser] = struct{}{}
	if parser.Tag() == tag {
		return true
	}
	for i := 0; i < parser.ChildParserCount(); i++ {
		if queryGrammarParserGraphHasTag(parser.ChildParser(i), tag, depth+1, seen) {
			return true
		}
	}
	return false
}
