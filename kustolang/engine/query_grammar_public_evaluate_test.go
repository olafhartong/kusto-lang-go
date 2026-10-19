package kustolang

import "testing"

func TestQueryGrammarPublicEvaluateCompleteASTParametersAndSchema(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"evaluate hint.distribution=per_node hint.remote=local plugin(1, Value) : (*, Id:long, ['Display Name']:string,)").(*EvaluateOperator)

	if operator.ChildCount() != 4 || operator.EvaluateKeyword().Kind() != SyntaxKindEvaluateKeyword {
		t.Fatalf("evaluate children/keyword = %d/%v, want 4/EvaluateKeyword", operator.ChildCount(), operator.EvaluateKeyword().Kind())
	}
	assertQueryGrammarSharedParameters(t, operator.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.distribution", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.remote", kind: SyntaxKindTokenLiteralExpression},
	})
	call := assertQueryGrammarEvaluateCall(t, operator.FunctionCall(), "plugin", 2)
	if call.ArgumentList().Expressions().At(0).Element().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("first evaluate argument kind = %v, want LongLiteralExpression", call.ArgumentList().Expressions().At(0).Element().Kind())
	}
	if value, ok := call.ArgumentList().Expressions().At(1).Element().(*NameReference); !ok || NameReferenceSimpleName(value) != "Value" {
		t.Fatalf("second evaluate argument = %#v, want Value name reference", call.ArgumentList().Expressions().At(1).Element())
	}

	clause := operator.Schema()
	if clause == nil || clause.ColonToken().Kind() != SyntaxKindColonToken || clause.ChildCount() != 2 {
		t.Fatalf("evaluate schema clause = %#v, want colon and schema", clause)
	}
	schema := clause.Schema()
	if schema.ChildCount() != 6 || schema.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		schema.LeadingComma() != nil || schema.AsteriskToken() == nil || schema.AsteriskToken().Kind() != SyntaxKindAsteriskToken ||
		schema.AsteriskTokenComma() == nil || schema.AsteriskTokenComma().Kind() != SyntaxKindCommaToken ||
		schema.CloseParen().Kind() != SyntaxKindCloseParenToken || schema.Columns().Count() != 2 {
		t.Fatalf("evaluate row schema shape = %#v", schema)
	}
	assertQueryGrammarEvaluateColumn(t, schema.Columns().At(0).Element(), "Id", SyntaxKindLongKeyword)
	assertQueryGrammarEvaluateColumn(t, schema.Columns().At(1).Element(), "Display Name", SyntaxKindStringKeyword)
	if separator := schema.Columns().At(1).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken {
		t.Fatalf("evaluate trailing column comma = %#v, want comma", separator)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("complete evaluate diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicEvaluateAllPlacements(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, production.parser, "evaluate plugin()").(*EvaluateOperator)
			assertQueryGrammarEvaluateCall(t, operator.FunctionCall(), "plugin", 0)
		})
	}

	root := parseQueryGrammarProduction(t, grammar.Expression, "evaluate plugin()")
	if _, ok := root.(*EvaluateOperator); !ok {
		t.Fatalf("root evaluate = %T, want *EvaluateOperator", root)
	}
	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "T | evaluate plugin()")
	pipe, node := queryGrammarPipeStage(t, pipeline, SyntaxKindEvaluateOperator)
	if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("evaluate pipeline source = %#v, want T", pipe.Expression())
	}
	assertQueryGrammarEvaluateCall(t, node.(*EvaluateOperator).FunctionCall(), "plugin", 0)

	nested := parseQueryGrammarProduction(t, grammar.Expression, "T | evaluate pluginOne() | evaluate pluginTwo()")
	secondPipe, secondNode := queryGrammarPipeStage(t, nested, SyntaxKindEvaluateOperator)
	_, firstNode := queryGrammarPipeStage(t, secondPipe.Expression(), SyntaxKindEvaluateOperator)
	assertQueryGrammarEvaluateCall(t, firstNode.(*EvaluateOperator).FunctionCall(), "pluginOne", 0)
	assertQueryGrammarEvaluateCall(t, secondNode.(*EvaluateOperator).FunctionCall(), "pluginTwo", 0)
}

func TestQueryGrammarPublicEvaluateCompletionTagAndParameterValues(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			items := queryGrammarEvaluateCompletionItems(production.parser, "")
			item, ok := completionItemByDisplayText(items, "evaluate")
			if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityLow {
				t.Fatalf("evaluate completion = %#v, visible items:%#v", item, completionDisplayTexts(items))
			}
			if !queryGrammarParserTreeHasTag(production.parser, "<evaluate>", 0) {
				t.Fatal("evaluate production parser tree has no <evaluate> tag")
			}
		})
	}

	for _, test := range []struct {
		text string
		want []string
	}{
		{text: "evaluate hint.distribution=", want: []string{"single", "per_node", "per_shard", "default"}},
		{text: "evaluate hint.remote=", want: []string{"auto", "local"}},
	} {
		t.Run(test.text, func(t *testing.T) {
			items := queryGrammarEvaluateCompletionItems(grammar.QueryOperator, test.text)
			for _, value := range test.want {
				item, ok := completionItemByDisplayText(items, value)
				if !ok || item.Kind != CompletionKindSyntax {
					t.Fatalf("evaluate value completion %q = %#v, visible items:%#v", value, item, completionDisplayTexts(items))
				}
			}
		})
	}
}

func TestQueryGrammarPublicEvaluateFunctionCallForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	known := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate count()").(*EvaluateOperator)
	assertQueryGrammarEvaluateCall(t, known.FunctionCall(), "count", 0)

	bracketed := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate ['plugin name']()").(*EvaluateOperator)
	call := assertQueryGrammarEvaluateCall(t, bracketed.FunctionCall(), "plugin name", 0)
	if call.Name().Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("bracketed evaluate function name kind = %v, want BracketedName", call.Name().Name().Kind())
	}

	bare := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin").(*EvaluateOperator)
	bareCall := assertQueryGrammarEvaluateCall(t, bare.FunctionCall(), "plugin", 0)
	if bareCall.ArgumentList().OpenParen().Width() != 0 || bareCall.ArgumentList().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(bareCall.ArgumentList().OpenParen().Diagnostics(), "KS005", "Expected: (") ||
		!diagnosticsContainMessage(bareCall.ArgumentList().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("bare evaluate argument list = %#v diagnostics:%#v", bareCall.ArgumentList(), bareCall.ArgumentList().ContainedDiagnostics())
	}

	missing := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate").(*EvaluateOperator)
	missingCall := missing.FunctionCall()
	if missingCall.Width() != 0 || !diagnosticsContainMessage(missingCall.Diagnostics(), "KS006", "Missing function call") {
		t.Fatalf("missing evaluate call = %#v diagnostics:%#v", missingCall, missingCall.Diagnostics())
	}
}

func TestQueryGrammarPublicEvaluateBracketRecoveryDottedCallsAndTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tokens := Lex("evaluate ['plugin'()")
	result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
	operator, ok := result.Value.(*EvaluateOperator)
	if !ok || result.Length != len(tokens)-1 {
		t.Fatalf("incomplete bracketed evaluate = %#v", result)
	}
	call := assertQueryGrammarEvaluateCall(t, operator.FunctionCall(), "plugin", 0)
	name, ok := call.Name().Name().(*BracketedName)
	if !ok || name.CloseBracket().Width() != 0 || !diagnosticsContainMessage(name.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("incomplete bracketed evaluate name = %#v diagnostics:%#v", call.Name().Name(), call.Name().ContainedDiagnostics())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate database('db').plugin()", 5)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin().member", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin() extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin(), extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin() | where true", 4)
}

func TestQueryGrammarPublicEvaluateSchemaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingSchema := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin():").(*EvaluateOperator).Schema().Schema()
	if missingSchema.Width() != 0 || !diagnosticsContainMessage(missingSchema.Diagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing evaluate schema = %#v diagnostics:%#v", missingSchema, missingSchema.Diagnostics())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin():(A:string").(*EvaluateOperator).Schema().Schema()
	if missingClose.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing evaluate schema close = %#v diagnostics:%#v", missingClose.CloseParen(), missingClose.CloseParen().Diagnostics())
	}

	missingType := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin():(A:)").(*EvaluateOperator).Schema().Schema().Columns().At(0).Element()
	if missingType.Type().Width() != 0 || !diagnosticsContainMessage(missingType.Type().Diagnostics(), "KS006", "Missing type name") {
		t.Fatalf("missing evaluate column type = %#v diagnostics:%#v", missingType.Type(), missingType.Type().Diagnostics())
	}

	missingName := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin():(:string)").(*EvaluateOperator).Schema().Schema().Columns().At(0).Element()
	if missingName.Name().Width() != 0 || !diagnosticsContainMessage(missingName.Name().Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing evaluate column name = %#v diagnostics:%#v", missingName.Name(), missingName.Name().Diagnostics())
	}

	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator, "evaluate plugin():(A:string B:long)").(*EvaluateOperator).Schema().Schema()
	if missingComma.Columns().Count() != 2 || missingComma.Columns().At(0).Separator().Width() != 0 ||
		!diagnosticsContainMessage(missingComma.Columns().At(0).Separator().Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing evaluate column comma = %#v", missingComma.Columns())
	}
}

func TestQueryGrammarPublicEvaluateSchemaCommaAndStarForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name         string
		text         string
		leadingComma bool
		star         bool
		starComma    bool
		columns      int
	}{
		{name: "empty", text: "evaluate plugin():()"},
		{name: "star", text: "evaluate plugin():(*)", star: true},
		{name: "star trailing comma", text: "evaluate plugin():(*,)", star: true, starComma: true},
		{name: "star columns", text: "evaluate plugin():(*, A:string)", star: true, starComma: true, columns: 1},
		{name: "leading comma", text: "evaluate plugin():(, A:string)", leadingComma: true, columns: 1},
		{name: "leading comma star", text: "evaluate plugin():(,*)", leadingComma: true, star: true},
		{name: "all prefixes", text: "evaluate plugin():(,*, A:string)", leadingComma: true, star: true, starComma: true, columns: 1},
		{name: "star comma without star", text: "evaluate plugin():(,, A:string)", leadingComma: true, starComma: true, columns: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*EvaluateOperator)
			schema := operator.Schema().Schema()
			if (schema.LeadingComma() != nil) != test.leadingComma || (schema.AsteriskToken() != nil) != test.star ||
				(schema.AsteriskTokenComma() != nil) != test.starComma || schema.Columns().Count() != test.columns {
				t.Fatalf("evaluate schema slots = leading:%#v star:%#v star-comma:%#v columns:%d",
					schema.LeadingComma(), schema.AsteriskToken(), schema.AsteriskTokenComma(), schema.Columns().Count())
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("evaluate schema form diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicEvaluateParametersMissingEqualsGlobalAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	missingEqual := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"evaluate hint.remote local plugin()").(*EvaluateOperator)
	parameter := missingEqual.Parameters().At(0)
	if parameter.EqualToken().Width() != 0 || !diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") ||
		parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("evaluate missing-equals parameter = %#v", parameter)
	}

	global := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"evaluate hint.strategy=shuffle plugin()").(*EvaluateOperator)
	assertQueryGrammarSharedParameters(t, global.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
	})

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate unknown=1 plugin()", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin() : (*) trailing", 8)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin() + 1", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "evaluate plugin() )", 4)
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "plugin()")
}

func assertQueryGrammarEvaluateCall(t *testing.T, call *FunctionCallExpression, name string, argumentCount int) *FunctionCallExpression {
	t.Helper()
	if call == nil || call.ChildCount() != 2 || NameReferenceSimpleName(call.Name()) != name {
		t.Fatalf("evaluate call = %#v, want two-child %q call", call, name)
	}
	arguments := call.ArgumentList()
	if arguments == nil || arguments.ChildCount() != 3 || arguments.Expressions().Count() != argumentCount {
		t.Fatalf("evaluate call %q arguments = %#v, want %d", name, arguments, argumentCount)
	}
	return call
}

func assertQueryGrammarEvaluateColumn(t *testing.T, column *NameAndTypeDeclaration, name string, kind SyntaxKind) {
	t.Helper()
	primitive, ok := column.Type().(*PrimitiveTypeExpression)
	if NameDeclarationSimpleName(column.Name()) != name || column.Colon().Kind() != SyntaxKindColonToken ||
		!ok || primitive.Type().Kind() != kind {
		t.Fatalf("evaluate column = name:%q colon:%v type:%#v, want %q/ColonToken/primitive %v",
			NameDeclarationSimpleName(column.Name()), column.Colon().Kind(), column.Type(), name, kind)
	}
}

func queryGrammarEvaluateCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
