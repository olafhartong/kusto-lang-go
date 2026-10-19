package kustolang

import "testing"

func TestQueryGrammarPublicParseOperatorsFullAST(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`Logs | parse kind=regex flags=U Message with * "user=" User:string " score=" Score:long | parse-where Detail with "ip=" IP:string`)

	parseWherePipe, parseWhereNode := queryGrammarPipeStage(t, expression, SyntaxKindParseWhereOperator)
	parsePipe, parseNode := queryGrammarPipeStage(t, parseWherePipe.Expression(), SyntaxKindParseOperator)
	if source, ok := parsePipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Logs" {
		t.Fatalf("parse pipeline source = %#v, want Logs", parsePipe.Expression())
	}

	parse := parseNode.(*ParseOperator)
	if parse.ParseKeyword().Kind() != SyntaxKindParseKeyword || parse.Expression().Kind() != SyntaxKindNameReference ||
		parse.WithKeyword().Kind() != SyntaxKindWithKeyword {
		t.Fatalf("parse fixed children = keyword:%v expression:%v with:%v", parse.ParseKeyword().Kind(), parse.Expression().Kind(), parse.WithKeyword().Kind())
	}
	if parse.Parameters().Count() != 2 || namedParameterNameText(t, parse.Parameters().At(0)) != "kind" ||
		namedParameterNameText(t, parse.Parameters().At(1)) != "flags" {
		t.Fatalf("parse parameters = %#v", parse.Parameters())
	}
	assertQueryGrammarParsePatternKinds(t, parse.Patterns(),
		SyntaxKindStarExpression,
		SyntaxKindStringLiteralExpression,
		SyntaxKindNameAndTypeDeclaration,
		SyntaxKindStringLiteralExpression,
		SyntaxKindNameAndTypeDeclaration,
	)

	parseWhere := parseWhereNode.(*ParseWhereOperator)
	if parseWhere.ParseKeyword().Kind() != SyntaxKindParseWhereKeyword || parseWhere.Expression().Kind() != SyntaxKindNameReference ||
		parseWhere.WithKeyword().Kind() != SyntaxKindWithKeyword || parseWhere.Parameters().Count() != 0 {
		t.Fatalf("parse-where fixed children = keyword:%v expression:%v with:%v parameters:%d",
			parseWhere.ParseKeyword().Kind(), parseWhere.Expression().Kind(), parseWhere.WithKeyword().Kind(), parseWhere.Parameters().Count())
	}
	assertQueryGrammarParsePatternKinds(t, parseWhere.Patterns(), SyntaxKindStringLiteralExpression, SyntaxKindNameAndTypeDeclaration)
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid parse pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicParseAndParseWherePlacements(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "parse", text: `parse Message with "id=" Id:long`, kind: SyntaxKindParseOperator},
		{name: "parse-where", text: `parse-where Message with "id=" Id:long`, kind: SyntaxKindParseWhereOperator},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, parser := range []ResultParser[LexicalToken, QueryOperator]{grammar.QueryOperator, grammar.FollowingPipeElementExpression} {
				operator := parseQueryGrammarProduction(t, parser, test.text)
				if operator.Kind() != test.kind {
					t.Fatalf("direct %s kind = %v, want %v", test.name, operator.Kind(), test.kind)
				}
			}

			pipeline := parseQueryGrammarProduction(t, grammar.Expression, "T | "+test.text)
			pipe, operator := queryGrammarPipeStage(t, pipeline, test.kind)
			if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
				t.Fatalf("%s pipeline source = %#v, want T", test.name, pipe.Expression())
			}
			if operator.Kind() != test.kind {
				t.Fatalf("%s pipeline operator = %v, want %v", test.name, operator.Kind(), test.kind)
			}
		})
	}
}

func TestQueryGrammarPublicParseParametersCompletionsAndTags(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.QueryOperator, `parse kind=regex flags=U Message with *`).(*ParseOperator)
	if operator.Parameters().Count() != 2 {
		t.Fatalf("parse parameters = %d, want 2", operator.Parameters().Count())
	}
	for i, want := range []string{"kind", "flags"} {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want || parameter.EqualToken().Kind() != SyntaxKindEqualToken ||
			parameter.EqualToken().Width() == 0 || parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
			t.Fatalf("parse parameter %d = %#v, want %q token literal", i, parameter, want)
		}
	}

	items := queryGrammarParseCompletionItems(grammar.QueryOperator, "")
	for _, display := range []string{"parse", "parse-where"} {
		item, ok := completionItemByDisplayText(items, display)
		if !ok || item.Kind != CompletionKindQueryPrefix {
			t.Fatalf("%s completion = %#v, visible:%#v", display, item, completionDisplayTexts(items))
		}
	}
	parameterItems := queryGrammarParseCompletionItems(grammar.QueryOperator, "parse ")
	for _, display := range []string{"kind", "flags"} {
		item, ok := completionItemByDisplayText(parameterItems, display)
		if !ok || item.BeforeText() != display+"=" || item.AfterText() != "" || !item.HasCaret() {
			t.Fatalf("%s parameter completion = %#v, visible:%#v", display, item, completionDisplayTexts(parameterItems))
		}
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<parse>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.QueryOperator, "<parse-where>", 0) {
		t.Fatal("QueryOperator parser tree is missing parse operator tags")
	}
}

func TestQueryGrammarPublicParsePatternAlternatives(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`parse Message with * "one" "two" Plain ['Display Name'] Typed:string ['Typed Name']:guid :long Bad:notatype`).(*ParseOperator)
	assertQueryGrammarParsePatternKinds(t, operator.Patterns(),
		SyntaxKindStarExpression,
		SyntaxKindCompoundStringLiteralExpression,
		SyntaxKindNameDeclaration,
		SyntaxKindNameDeclaration,
		SyntaxKindNameAndTypeDeclaration,
		SyntaxKindNameAndTypeDeclaration,
		SyntaxKindNameAndTypeDeclaration,
		SyntaxKindNameAndTypeDeclaration,
	)

	plain := operator.Patterns().At(2).(*NameDeclaration)
	bracketed := operator.Patterns().At(3).(*NameDeclaration)
	if NameDeclarationSimpleName(plain) != "Plain" || NameDeclarationSimpleName(bracketed) != "Display Name" {
		t.Fatalf("optional type declarations = plain:%#v bracketed:%#v", plain, bracketed)
	}
	missingName := operator.Patterns().At(6).(*NameAndTypeDeclaration)
	if missingName.Name().Width() != 0 || !diagnosticsContainMessage(missingName.Name().Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("colon-only declaration name = %#v diagnostics:%#v", missingName.Name(), missingName.Name().Diagnostics())
	}
	invalid := operator.Patterns().At(7).(*NameAndTypeDeclaration)
	if invalid.Type().Kind() != SyntaxKindPrimitiveTypeExpression || !diagnosticsContain(invalid.Type().Diagnostics(), "KS160") {
		t.Fatalf("invalid parse type = %#v diagnostics:%#v", invalid.Type(), invalid.Type().Diagnostics())
	}
	if diagnostics := operator.ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS160") {
		t.Fatalf("parse pattern diagnostics = %#v, want invalid type", diagnostics)
	}
}

func TestQueryGrammarPublicParseRequiredPartRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name string
		text string
	}{
		{name: "parse missing source and with", text: "parse"},
		{name: "parse-where missing source and with", text: "parse-where"},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			var source Expression
			var with *SyntaxToken
			var patterns *SyntaxListOf[SyntaxNodeElement]
			switch node := operator.(type) {
			case *ParseOperator:
				source, with, patterns = node.Expression(), node.WithKeyword(), node.Patterns()
			case *ParseWhereOperator:
				source, with, patterns = node.Expression(), node.WithKeyword(), node.Patterns()
			default:
				t.Fatalf("operator = %T, want parse operator", operator)
			}
			if source.Width() != 0 || !diagnosticsContainMessage(source.Diagnostics(), "KS006", "Missing expression") {
				t.Fatalf("missing parse source = %#v diagnostics:%#v", source, source.Diagnostics())
			}
			if with.Width() != 0 || !diagnosticsContainMessage(with.Diagnostics(), "KS005", "Expected: with") {
				t.Fatalf("missing with token = %#v diagnostics:%#v", with, with.Diagnostics())
			}
			if patterns.Count() != 0 {
				t.Fatalf("missing parse pattern list = %d, want empty", patterns.Count())
			}
		})
	}

	missingType := parseQueryGrammarProduction(t, grammar.QueryOperator, `parse Message with "id=" Id:`).(*ParseOperator)
	declaration := missingType.Patterns().At(1).(*NameAndTypeDeclaration)
	if declaration.Type().Width() != 0 || !diagnosticsContainMessage(declaration.Type().Diagnostics(), "KS006", "Missing type name") {
		t.Fatalf("missing parse type = %#v diagnostics:%#v", declaration.Type(), declaration.Type().Diagnostics())
	}
	if diagnosticsContain(declaration.Type().Diagnostics(), "KS001") {
		t.Fatalf("missing parse type has redundant missing-token diagnostic: %#v", declaration.Type().Diagnostics())
	}
}

func TestQueryGrammarPublicParseEmptyPatternsMalformedTailsAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{"parse Message with", "parse-where Message with"} {
		operator := parseQueryGrammarProduction(t, grammar.QueryOperator, text)
		var count int
		switch node := operator.(type) {
		case *ParseOperator:
			count = node.Patterns().Count()
		case *ParseWhereOperator:
			count = node.Patterns().Count()
		default:
			t.Fatalf("operator %q = %T, want parse operator", text, operator)
		}
		if count != 0 || len(operator.ContainedDiagnostics()) != 0 {
			t.Fatalf("empty patterns for %q = count:%d diagnostics:%#v", text, count, operator.ContainedDiagnostics())
		}
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, `parse Message with * , trailing`, 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, `parse-where Message with Name:string . tail`, 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, `parse Message with "ok" 42`, 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, `parse Message with Name = value`, 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, `parse unknown=1 Message with *`, 2)

	for _, text := range []string{"parsewhere Message with *", "parse_where Message with *"} {
		assertQueryGrammarProductionRejects(t, grammar.QueryOperator, text)
	}
}

func assertQueryGrammarParsePatternKinds(t *testing.T, patterns *SyntaxListOf[SyntaxNodeElement], want ...SyntaxKind) {
	t.Helper()
	if patterns.Count() != len(want) {
		t.Fatalf("parse patterns = %d, want %d", patterns.Count(), len(want))
	}
	for i, kind := range want {
		if patterns.At(i).Kind() != kind {
			t.Fatalf("parse pattern %d = %T/%v, want %v", i, patterns.At(i), patterns.At(i).Kind(), kind)
		}
	}
}

func queryGrammarParseCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
