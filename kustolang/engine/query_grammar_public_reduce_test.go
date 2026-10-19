package kustolang

import "testing"

func TestQueryGrammarPublicReduceValidASTAndNamedExpression(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.QueryOperator,
		`reduce kind=source by Pattern=Message with threshold=0.5, characters='-_'`)
	reduce, ok := operator.(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %T, want *ReduceByOperator", operator)
	}
	if reduce.ReduceKeyword().Kind() != SyntaxKindReduceKeyword || reduce.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("reduce keywords = %v/%v", reduce.ReduceKeyword().Kind(), reduce.ByKeyword().Kind())
	}
	if reduce.Parameters().Count() != 1 {
		t.Fatalf("reduce parameters = %d, want 1", reduce.Parameters().Count())
	}
	kind := reduce.Parameters().At(0)
	if NameDeclarationSimpleName(kind.Name()) != "kind" || kind.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("reduce kind parameter = %#v", kind)
	}
	named, ok := reduce.Expression().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(named.Name()) != "Pattern" || named.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("reduce named expression = %#v", reduce.Expression())
	}
	if reduce.With() == nil || reduce.With().WithKeyword().Kind() != SyntaxKindWithKeyword || reduce.With().Parameters().Count() != 2 {
		t.Fatalf("reduce with clause = %#v", reduce.With())
	}
	threshold := reduce.With().Parameters().At(0).Element()
	characters := reduce.With().Parameters().At(1).Element()
	if NameDeclarationSimpleName(threshold.Name()) != "threshold" || threshold.Expression().Kind() != SyntaxKindRealLiteralExpression {
		t.Fatalf("reduce threshold parameter = %#v", threshold)
	}
	if NameDeclarationSimpleName(characters.Name()) != "characters" || characters.Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("reduce characters parameter = %#v", characters)
	}
	if diagnostics := reduce.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid reduce diagnostics = %#v", diagnostics)
	}

	withoutWith := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce by Message").(*ReduceByOperator)
	if withoutWith.With() != nil || withoutWith.Parameters().Count() != 0 || len(withoutWith.ContainedDiagnostics()) != 0 {
		t.Fatalf("minimal reduce = %#v diagnostics:%#v", withoutWith, withoutWith.ContainedDiagnostics())
	}
	emptyWith := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce by Message with").(*ReduceByOperator)
	if emptyWith.With() == nil || emptyWith.With().Parameters().Count() != 0 || len(emptyWith.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty reduce with = %#v diagnostics:%#v", emptyWith.With(), emptyWith.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicReduceParametersDoNotNeedEqualsForRecognition(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	reduce := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"reduce kind source hint.shufflekey Tenant by Message").(*ReduceByOperator)
	if reduce.Parameters().Count() != 2 {
		t.Fatalf("reduce parameters = %d, want 2", reduce.Parameters().Count())
	}
	for i, want := range []struct {
		name string
		kind SyntaxKind
	}{{"kind", SyntaxKindTokenLiteralExpression}, {"hint.shufflekey", SyntaxKindNameReference}} {
		parameter := reduce.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want.name || parameter.Expression().Kind() != want.kind {
			t.Fatalf("reduce parameter %d = %#v", i, parameter)
		}
		if parameter.EqualToken().Width() != 0 ||
			!diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
			t.Fatalf("reduce parameter %d missing equals = %#v diagnostics:%#v", i, parameter.EqualToken(), parameter.EqualToken().Diagnostics())
		}
	}
	if reduce.ByKeyword().Width() == 0 || reduce.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("reduce clause after recovered parameters = %#v", reduce)
	}

	emptyValue := parseQueryGrammarProduction(t, grammar.reduceByOperator, "reduce kind=").(*ReduceByOperator)
	if emptyValue.Parameters().Count() != 1 || emptyValue.ByKeyword().Width() != 0 ||
		emptyValue.Parameters().At(0).Expression().Width() != 0 ||
		!diagnosticsContainMessage(emptyValue.Parameters().At(0).Expression().ContainedDiagnostics(), "KS005", "Expected: source") {
		t.Fatalf("empty kind value recovery = %#v diagnostics:%#v", emptyValue, emptyValue.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicReduceWithCommaParametersDeclaredOrKnown(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	reduce := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"reduce by Message with threshold=1, hint.shufflekey=Tenant, characters='x'").(*ReduceByOperator)
	parameters := reduce.With().Parameters()
	if parameters.Count() != 3 || parameters.At(0).Separator() == nil || parameters.At(1).Separator() == nil ||
		parameters.At(2).Separator() != nil {
		t.Fatalf("reduce with parameter/separator shape = %#v, want three parameters with two separators", parameters)
	}
	for i, want := range []string{"threshold", "hint.shufflekey", "characters"} {
		if got := NameDeclarationSimpleName(parameters.At(i).Element().Name()); got != want {
			t.Fatalf("reduce with parameter %d = %q, want %q", i, got, want)
		}
	}
	if diagnostics := reduce.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("declared-or-known with diagnostics = %#v", diagnostics)
	}

	missingEqual := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"reduce by Message with threshold 1, characters 'x'").(*ReduceByOperator)
	for i := 0; i < missingEqual.With().Parameters().Count(); i++ {
		parameter := missingEqual.With().Parameters().At(i).Element()
		if parameter.EqualToken().Width() != 0 ||
			!diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
			t.Fatalf("with parameter %d missing equals = %#v diagnostics:%#v", i, parameter.EqualToken(), parameter.EqualToken().Diagnostics())
		}
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "reduce by Message with unknown=1", 4)
	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"reduce by Message with threshold=1 characters='x'").(*ReduceByOperator).With().Parameters()
	if missingComma.Count() != 2 || missingComma.At(0).Separator() == nil || missingComma.At(0).Separator().Width() != 0 ||
		!diagnosticsContainMessage(missingComma.At(0).Separator().Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("reduce missing comma recovery = %#v", missingComma)
	}
	trailingComma := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"reduce by Message with threshold=1,").(*ReduceByOperator).With().Parameters()
	if trailingComma.Count() != 2 || trailingComma.At(0).Separator() == nil || trailingComma.At(0).Separator().Width() == 0 ||
		trailingComma.At(1).Element().Width() != 0 {
		t.Fatalf("reduce trailing comma = %#v", trailingComma)
	}
}

func TestQueryGrammarPublicReduceRequiredPartsAndMalformedSyntax(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingBoth := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce").(*ReduceByOperator)
	if missingBoth.ByKeyword().Width() != 0 ||
		!diagnosticsContainMessage(missingBoth.ByKeyword().Diagnostics(), "KS005", "Expected: by") ||
		missingBoth.Expression().Width() != 0 ||
		!diagnosticsContain(missingBoth.Expression().ContainedDiagnostics(), "KS006") {
		t.Fatalf("reduce missing required parts = %#v diagnostics:%#v", missingBoth, missingBoth.ContainedDiagnostics())
	}

	missingBy := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce Message").(*ReduceByOperator)
	if missingBy.ByKeyword().Width() != 0 || missingBy.Expression().Kind() != SyntaxKindNameReference ||
		!diagnosticsContainMessage(missingBy.ByKeyword().Diagnostics(), "KS005", "Expected: by") {
		t.Fatalf("reduce missing by recovery = %#v diagnostics:%#v", missingBy, missingBy.ContainedDiagnostics())
	}
	missingExpression := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce by").(*ReduceByOperator)
	if missingExpression.Expression().Width() != 0 ||
		!diagnosticsContain(missingExpression.Expression().ContainedDiagnostics(), "KS006") {
		t.Fatalf("reduce missing expression = %#v diagnostics:%#v", missingExpression.Expression(), missingExpression.ContainedDiagnostics())
	}
	withBoundary := parseQueryGrammarProduction(t, grammar.QueryOperator, "reduce by with threshold=1").(*ReduceByOperator)
	if withBoundary.Expression().Width() != 0 || withBoundary.With() == nil || withBoundary.With().Parameters().Count() != 1 {
		t.Fatalf("reduce missing expression before with = %#v", withBoundary)
	}

	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "by Message")
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "reduce by Message Extra", 3)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "reduce by Message, Extra", 3)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "reduce by Message with threshold=1 Extra", 7)
}

func TestQueryGrammarPublicReducePostPipeAndForkPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe", parser: grammar.FollowingPipeElementExpression},
		{name: "fork pipe", parser: grammar.forkPipeOperator},
	} {
		t.Run(production.name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, production.parser, "reduce by Message").(*ReduceByOperator); !ok {
				t.Fatalf("reduce placement through %s did not return *ReduceByOperator", production.name)
			}
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.Expression, "reduce by Message", 1)
	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		"Events | reduce by Message with threshold=0.5 | project Pattern, Count")
	projectPipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindProjectOperator)
	reducePipe, node := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindReduceByOperator)
	if source, ok := reducePipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("reduce pipeline source = %#v, want Events", reducePipe.Expression())
	}
	if node.(*ReduceByOperator).With() == nil {
		t.Fatal("reduce pipeline lost with clause")
	}

	fork := parseQueryGrammarProduction(t, grammar.forkOperator,
		"fork (reduce by Message | project Pattern, Count) Named=(reduce by Other)").(*ForkOperator)
	if fork.Expressions().Count() != 2 {
		t.Fatalf("fork reduce branches = %d, want 2", fork.Expressions().Count())
	}
	firstPipe, _ := queryGrammarPipeStage(t, fork.Expressions().At(0).Expression(), SyntaxKindProjectOperator)
	if firstPipe.Expression().Kind() != SyntaxKindReduceByOperator || fork.Expressions().At(1).Expression().Kind() != SyntaxKindReduceByOperator {
		t.Fatalf("fork reduce branch kinds = %v/%v", firstPipe.Expression().Kind(), fork.Expressions().At(1).Expression().Kind())
	}
}

func TestQueryGrammarPublicReduceTagsAndCompletion(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	if grammar.reduceByOperator.Tag() != "<reduce-by>" ||
		!queryGrammarParserTreeHasTag(grammar.QueryOperator, "<reduce-by>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.FollowingPipeElementExpression, "<reduce-by>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.forkPipeOperator, "<reduce-by>", 0) {
		t.Fatalf("reduce tags = direct:%q query:%v following:%v fork:%v", grammar.reduceByOperator.Tag(),
			queryGrammarParserTreeHasTag(grammar.QueryOperator, "<reduce-by>", 0),
			queryGrammarParserTreeHasTag(grammar.FollowingPipeElementExpression, "<reduce-by>", 0),
			queryGrammarParserTreeHasTag(grammar.forkPipeOperator, "<reduce-by>", 0))
	}

	for _, parser := range []Parser[LexicalToken]{grammar.reduceByOperator, grammar.QueryOperator, grammar.FollowingPipeElementExpression, grammar.forkPipeOperator} {
		items := queryGrammarReduceCompletionItems(parser, "")
		item, ok := completionItemByDisplayText(items, "reduce")
		if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityLow {
			t.Fatalf("reduce completion through tag %q = %#v, visible:%#v", parser.Tag(), item, completionDisplayTexts(items))
		}
	}
	for _, value := range []string{"source"} {
		item, ok := completionItemByDisplayText(queryGrammarReduceCompletionItems(grammar.reduceByOperator, "reduce kind="), value)
		if !ok || item.Kind != CompletionKindSyntax {
			t.Fatalf("reduce kind completion %q = %#v", value, item)
		}
	}
	for _, parameter := range []string{"threshold", "characters"} {
		if _, ok := completionItemByDisplayText(queryGrammarReduceCompletionItems(grammar.reduceByOperator, "reduce by Message with "), parameter); !ok {
			t.Fatalf("reduce with completion %q missing", parameter)
		}
	}
}

func queryGrammarReduceCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
