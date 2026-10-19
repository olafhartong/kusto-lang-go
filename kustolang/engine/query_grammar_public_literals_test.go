package kustolang

import "testing"

func TestQueryGrammarLiteralParserAnnotationsAndCompletions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	tagged := []struct {
		parser Parser[LexicalToken]
		tag    string
	}{
		{grammar.booleanLiteral, "<bool-literal>"},
		{grammar.longLiteral, "<long-literal>"},
		{grammar.realLiteral, "<real-literal>"},
		{grammar.decimalLiteral, "<decimal-literal>"},
		{grammar.intLiteral, "<int-literal>"},
		{grammar.guidLiteral, "<guid-literal>"},
		{grammar.rawGuidLiteral, "<raw-guid-literal>"},
		{grammar.dateTimeLiteral, "<datetime-literal>"},
		{grammar.timespanLiteral, "<timespan-literal>"},
		{grammar.StringLiteral, "<string-literal>"},
		{grammar.literalCore, "<literal>"},
		{grammar.Literal, "<literal>"},
	}
	for _, test := range tagged {
		if got := test.parser.Tag(); got != test.tag {
			t.Errorf("parser tag = %q, want %q", got, test.tag)
		}
		if !queryGrammarParserTreeHasTag(grammar.Literal, test.tag, 0) && test.parser != grammar.StringLiteral {
			t.Errorf("literal parser graph does not contain tag %q", test.tag)
		}
	}

	assertLiteralConstructorCompletions(t, grammar.longLiteral, "long")
	assertLiteralConstructorCompletions(t, grammar.decimalLiteral, "decimal")
	assertLiteralConstructorCompletions(t, grammar.intLiteral, "int")
	assertLiteralConstructorCompletions(t, grammar.guidLiteral, "guid")
	assertLiteralConstructorCompletions(t, grammar.dateTimeLiteral, "datetime")
	assertLiteralConstructorCompletions(t, grammar.timespanLiteral, "timespan")
	assertLiteralConstructorCompletions(t, grammar.realLiteral, "real", "double")
	if items := syntaxCompletionItemsFromAnnotations(grammar.rawGuidLiteral.Annotations()); len(items) != 0 {
		t.Fatalf("raw-guid constructor completions = %#v, want none", items)
	}

	graphConstructors := []string{}
	seenConstructors := map[string]bool{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.Literal, 0) {
		for _, item := range syntaxCompletionItemsFromAnnotations(parser.Annotations()) {
			if item.Kind == CompletionKindScalarPrefix && !seenConstructors[item.DisplayText] {
				seenConstructors[item.DisplayText] = true
				graphConstructors = append(graphConstructors, item.DisplayText)
			}
		}
	}
	wantGraphConstructors := []string{"long()", "real()", "double()", "decimal()", "int()", "guid()", "datetime()", "timespan()"}
	if len(graphConstructors) != len(wantGraphConstructors) {
		t.Fatalf("literal graph constructor completions = %#v, want %#v", graphConstructors, wantGraphConstructors)
	}
	for i, want := range wantGraphConstructors {
		if graphConstructors[i] != want {
			t.Fatalf("literal graph constructor completion %d = %q, want %q; all:%#v", i, graphConstructors[i], want, graphConstructors)
		}
	}

	booleanItems := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.booleanLiteralWithCompletion, 0) {
		booleanItems = append(booleanItems, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	if got := completionDisplayTexts(booleanItems); len(got) != 2 || got[0] != "true" || got[1] != "false" {
		t.Fatalf("boolean completion order = %#v, want [true false]", got)
	}
	for i, want := range []string{"true", "false"} {
		item := booleanItems[i]
		if item.Kind != CompletionKindSyntax || item.DisplayText != want || item.MatchText != want || item.OrderText != want ||
			item.rank != CompletionRankDefault || item.priority != CompletionPriorityNormal || len(item.ApplyTexts) != 1 ||
			item.ApplyTexts[0] != (CompletionText{Text: want}) {
			t.Errorf("boolean completion %q = %#v", want, item)
		}
	}
}

func TestQueryGrammarLiteralPrivateProductionsHaveNoRecoveryAndRawGuidBoundary(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parsers := []ResultParser[LexicalToken, Expression]{
		grammar.booleanLiteral,
		grammar.longLiteral,
		grammar.realLiteral,
		grammar.decimalLiteral,
		grammar.intLiteral,
		grammar.guidLiteral,
		grammar.rawGuidLiteral,
		grammar.dateTimeLiteral,
		grammar.timespanLiteral,
		grammar.StringLiteral,
	}
	for _, parser := range parsers {
		assertQueryGrammarProductionRejects(t, parser, "")
	}

	const rawGuid = "00000000-0000-0000-0000-000000000001"
	private := parseQueryGrammarProduction(t, grammar.rawGuidLiteral, rawGuid)
	if diagnostics := private.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("private raw-guid diagnostics = %#v, want none", diagnostics)
	}
	public := parseQueryGrammarProduction(t, grammar.Literal, rawGuid)
	if diagnostics := public.ContainedDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS214" {
		t.Fatalf("public raw-guid diagnostics = %#v, want only KS214", diagnostics)
	}
}

func assertLiteralConstructorCompletions(t *testing.T, parser Parser[LexicalToken], names ...string) {
	t.Helper()
	items := syntaxCompletionItemsFromAnnotations(parser.Annotations())
	if len(items) != len(names) {
		t.Fatalf("constructor completion count = %d, want %d: %#v", len(items), len(names), items)
	}
	for i, name := range names {
		item := items[i]
		wantApply := []CompletionText{{Text: name + "("}, {Text: ")", Caret: true}}
		if item.Kind != CompletionKindScalarPrefix || item.DisplayText != name+"()" || item.MatchText != name || item.OrderText != name ||
			item.rank != CompletionRankFunction || item.priority != CompletionPriorityNormal || len(item.ApplyTexts) != len(wantApply) ||
			item.ApplyTexts[0] != wantApply[0] || item.ApplyTexts[1] != wantApply[1] {
			t.Errorf("constructor completion %q = %#v, want apply %#v", name, item, wantApply)
		}
	}
}

func TestQueryGrammarPublicParamTypeExtended(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{"bool", "dynamic", "float", "int16", "uint64", "ulong"} {
		t.Run(text, func(t *testing.T) {
			typ := parseQueryGrammarProduction(t, grammar.ParamTypeExtended, text)
			primitive, ok := typ.(*PrimitiveTypeExpression)
			if !ok || primitive.Type().RawText() != text {
				t.Fatalf("extended parameter type %q = %#v", text, typ)
			}
		})
	}
	for _, text := range []string{"CustomType", "project", "'string'"} {
		t.Run("reject "+text, func(t *testing.T) {
			assertQueryGrammarProductionRejects(t, grammar.ParamTypeExtended, text)
		})
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.ParamTypeExtended, "long trailing", 1)
}

func TestQueryGrammarPublicSchemaTypeValidAndRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	asterisk := parseQueryGrammarProduction(t, grammar.SchemaType, "(*)")
	if asterisk.Columns().Count() != 1 || asterisk.Columns().At(0).Element().Kind() != SyntaxKindStarExpression {
		t.Fatalf("asterisk schema columns = %#v", asterisk.Columns())
	}

	nested := parseQueryGrammarProduction(t, grammar.SchemaType, "(Account:string, Nested:(Value:long, Tags:dynamic),)")
	if nested.Columns().Count() != 2 {
		t.Fatalf("nested schema column count = %d", nested.Columns().Count())
	}
	nestedColumn, ok := nested.Columns().At(1).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("nested schema column = %#v", nested.Columns().At(1).Element())
	}
	nestedType, ok := nestedColumn.Type().(*SchemaTypeExpression)
	if !ok || nestedType.Columns().Count() != 2 {
		t.Fatalf("nested schema type = %#v", nestedColumn.Type())
	}
	if diagnostics := nested.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("nested schema diagnostics = %#v", diagnostics)
	}

	missingClose := parseQueryGrammarProduction(t, grammar.SchemaType, "(Account:string")
	if missingClose.CloseParen().Width() != 0 || !diagnosticsContainMessage(missingClose.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing-close schema = close:%#v diagnostics:%#v", missingClose.CloseParen(), missingClose.ContainedDiagnostics())
	}
	missingType := parseQueryGrammarProduction(t, grammar.SchemaType, "(Account:)")
	if !diagnosticsContain(missingType.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing-type schema diagnostics = %#v", missingType.ContainedDiagnostics())
	}
	missingElement := parseQueryGrammarProduction(t, grammar.SchemaType, "(,)")
	if missingElement.Columns().Count() != 1 || !diagnosticsContainMessage(missingElement.ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("missing schema element = count:%d diagnostics:%#v", missingElement.Columns().Count(), missingElement.ContainedDiagnostics())
	}
	missingName := parseQueryGrammarProduction(t, grammar.SchemaType, "(:string)")
	if missingName.Columns().Count() != 1 || !diagnosticsContainMessage(missingName.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing schema name diagnostics = %#v", missingName.ContainedDiagnostics())
	}
	missingComma := parseQueryGrammarProduction(t, grammar.SchemaType, "(A:string B:long)")
	if missingComma.Columns().Count() != 2 || !diagnosticsContainMessage(missingComma.ContainedDiagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing schema comma = count:%d diagnostics:%#v", missingComma.Columns().Count(), missingComma.ContainedDiagnostics())
	}
	starMissingClose := grammar.SchemaType.ParseValue(NewArraySource(Lex("(*")), 0)
	if starMissingClose.Length != 1 || starMissingClose.Value == nil || starMissingClose.Value.Columns().Count() != 0 {
		t.Fatalf("incomplete star schema fallback = %#v", starMissingClose)
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.SchemaType, "(A:string) trailing", 5)
}

func TestQueryGrammarPublicStringLiteralForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name      string
		text      string
		valueText string
	}{
		{name: "quoted", text: `'a\nb'`, valueText: "a\nb"},
		{name: "verbatim", text: `@"a ""quoted"" value"`, valueText: `a "quoted" value`},
		{name: "hidden", text: `h'secret'`, valueText: "secret"},
		{name: "multiline", text: "```line1\nline2```", valueText: "line1\nline2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			literal := parseQueryGrammarStringLiteral(t, grammar.StringLiteral, test.text)
			if literal.Kind() != SyntaxKindStringLiteralExpression {
				t.Fatalf("string literal kind = %v", literal.Kind())
			}
			if literal.Token().RawText() != test.text || literal.Token().ValueText() != test.valueText {
				t.Fatalf("string token = raw:%q value:%q, want raw:%q value:%q", literal.Token().RawText(), literal.Token().ValueText(), test.text, test.valueText)
			}
			if diagnostics := literal.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid string diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicStringLiteralPreservesMalformedTokenDiagnostics(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		name    string
		text    string
		message string
	}{
		{name: "single quote", text: `'unterminated`, message: "Missing: '"},
		{name: "multiline", text: "```unterminated", message: "Missing: ```"},
	} {
		t.Run(test.name, func(t *testing.T) {
			literal := parseQueryGrammarStringLiteral(t, grammar.StringLiteral, test.text)
			diagnostics := literal.ContainedDiagnostics()
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS001" || diagnostics[0].Message != test.message || diagnostics[0].LocationKind != DiagnosticLocationRelativeEnd {
				t.Fatalf("malformed string diagnostics = %#v", diagnostics)
			}
			if literal.Token().RawText() != test.text {
				t.Fatalf("malformed string raw text = %q, want %q", literal.Token().RawText(), test.text)
			}
		})
	}
}

func TestQueryGrammarPublicStringLiteralOptionsAndFullConsumption(t *testing.T) {
	defaultGrammar := QueryGrammarFrom(NewGlobalState(nil))
	lineBreakOptions := DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	lineBreakGrammar := QueryGrammarFrom(NewGlobalState(nil).WithParseOptions(lineBreakOptions))
	if lineBreakGrammar == defaultGrammar || !lineBreakGrammar.Options.AllowLiteralsWithLineBreaks {
		t.Fatalf("line-break options grammar = %#v, default = %#v", lineBreakGrammar.Options, defaultGrammar.Options)
	}
	parseQueryGrammarStringLiteral(t, lineBreakGrammar.StringLiteral, "~~~line1\nline2~~~")

	for _, text := range []string{"", "42", "true", "identifier", "`not a string`"} {
		t.Run("reject "+text, func(t *testing.T) {
			result := defaultGrammar.StringLiteral.ParseValue(NewArraySource(Lex(text)), 0)
			if result.Length >= 0 || result.Value != nil {
				t.Fatalf("non-string %q parsed: %#v", text, result)
			}
		})
	}

	for _, text := range []string{"'first' 'second'", "'value' trailing"} {
		t.Run("prefix "+text, func(t *testing.T) {
			tokens := Lex(text)
			result := defaultGrammar.StringLiteral.ParseValue(NewArraySource(tokens), 0)
			if result.Length != 1 || result.Value == nil {
				t.Fatalf("string prefix parse = %#v", result)
			}
			if result.Length == len(tokens)-1 {
				t.Fatalf("string parser unexpectedly consumed all of %q", text)
			}
		})
	}
}

func TestQueryGrammarPublicLiteralSignedNumericAndRawGuid(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		text string
		kind SyntaxKind
	}{
		{text: "-42", kind: SyntaxKindLongLiteralExpression},
		{text: "+1.5", kind: SyntaxKindRealLiteralExpression},
		{text: "-2h", kind: SyntaxKindTimespanLiteralExpression},
	} {
		t.Run(test.text, func(t *testing.T) {
			expression := parseQueryGrammarProduction(t, grammar.Literal, test.text)
			literal, ok := expression.(*LiteralExpression)
			if !ok || literal.Kind() != test.kind || literal.Token().RawText() != test.text {
				t.Fatalf("signed literal %q = %#v", test.text, expression)
			}
		})
	}
	for _, text := range []string{"- 42", "+ long(1)"} {
		t.Run("reject separated "+text, func(t *testing.T) {
			assertQueryGrammarProductionRejects(t, grammar.Literal, text)
		})
	}

	raw := parseQueryGrammarProduction(t, grammar.Literal, "00000000-0000-0000-0000-000000000001")
	if raw.Kind() != SyntaxKindGuidLiteralExpression || !diagnosticsContain(raw.ContainedDiagnostics(), "KS214") {
		t.Fatalf("raw guid literal = kind:%v diagnostics:%#v", raw.Kind(), raw.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicTypeOfLiteralUsesNativeDeclarationGrammar(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	scalar := parseQueryGrammarProduction(t, grammar.Literal, "typeof(long)")
	if scalar.Kind() != SyntaxKindTypeOfLiteralExpression || len(scalar.ContainedDiagnostics()) != 0 {
		t.Fatalf("scalar typeof = kind:%v diagnostics:%#v", scalar.Kind(), scalar.ContainedDiagnostics())
	}
	tabular := parseQueryGrammarProduction(t, grammar.Literal, "typeof(Account:string, Nested:(Value:long), *)")
	typeof, ok := tabular.(*TypeOfLiteralExpression)
	if !ok || typeof.Types().Count() != 3 || len(typeof.ContainedDiagnostics()) != 0 {
		t.Fatalf("tabular typeof = %#v diagnostics:%#v", tabular, tabular.ContainedDiagnostics())
	}
	empty := parseQueryGrammarProduction(t, grammar.Literal, "typeof()")
	if !diagnosticsContainMessage(empty.ContainedDiagnostics(), "KS006", "Missing type name") {
		t.Fatalf("empty typeof diagnostics = %#v", empty.ContainedDiagnostics())
	}
	assertQueryGrammarProductionRejects(t, grammar.Literal, "typeof")
}

func TestQueryGrammarPublicJsonValueAndDynamicRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	expression := parseQueryGrammarProduction(t, grammar.Literal, `dynamic({"items":[1,-2,{"ok":true}],"nested":dynamic(null)})`)
	dynamic, ok := expression.(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic literal = %#v", expression)
	}
	object, ok := dynamic.Expression().(*JsonObjectExpression)
	if !ok || object.Pairs().Count() != 2 {
		t.Fatalf("dynamic JSON object = %#v", dynamic.Expression())
	}
	items, ok := object.Pairs().At(0).Element().Value().(*JsonArrayExpression)
	if !ok || items.Values().Count() != 3 {
		t.Fatalf("dynamic JSON items = %#v", object.Pairs().At(0).Element().Value())
	}
	if diagnostics := dynamic.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid dynamic diagnostics = %#v", diagnostics)
	}

	recovered := parseQueryGrammarProduction(t, grammar.JsonValue, `{"items":[1,2}`)
	recoveredObject, ok := recovered.(*JsonObjectExpression)
	if !ok || !diagnosticsContainMessage(recoveredObject.ContainedDiagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("missing JSON array close = %#v diagnostics:%#v", recovered, recovered.ContainedDiagnostics())
	}
	missingValue := parseQueryGrammarProduction(t, grammar.JsonValue, `{"missing":}`)
	if !diagnosticsContain(missingValue.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing JSON value diagnostics = %#v", missingValue.ContainedDiagnostics())
	}
	missingDynamicClose := parseQueryGrammarProduction(t, grammar.Literal, `dynamic([1,{"x":false}]`)
	if !diagnosticsContainMessage(missingDynamicClose.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing dynamic close diagnostics = %#v", missingDynamicClose.ContainedDiagnostics())
	}
	assertQueryGrammarProductionRejects(t, grammar.JsonValue, "int(1)")
}

func TestQueryGrammarPublicLiteralListRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	trailing := parseQueryGrammarProduction(t, grammar.LiteralList, "1, 'two',")
	if trailing.Count() != 2 || trailing.At(1).Separator() == nil || trailing.At(1).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("trailing-comma literal list = %#v", trailing)
	}
	if diagnostics := trailing.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("allowed trailing comma diagnostics = %#v", diagnostics)
	}
	leading := parseQueryGrammarProduction(t, grammar.LiteralList, ",1")
	if leading.Count() != 2 || !diagnosticsContainMessage(leading.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("leading missing literal = count:%d diagnostics:%#v", leading.Count(), leading.ContainedDiagnostics())
	}

	tokens := Lex("1,,3")
	doubleComma := grammar.LiteralList.ParseValue(NewArraySource(tokens), 0)
	if doubleComma.Length != 2 || doubleComma.Value == nil || doubleComma.Value.Count() != 1 {
		t.Fatalf("double-comma literal list = %#v", doubleComma)
	}
	if doubleComma.Length == len(tokens)-1 {
		t.Fatal("double-comma literal list unexpectedly consumed the trailing value")
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.LiteralList, "1, 2 trailing", 3)
}

func TestQueryGrammarPublicSkippedTokensFirstDiagnosticOnly(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	skipped := parseQueryGrammarProduction(t, grammar.SkippedTokens, "bad tail +")
	if skipped.Tokens().Count() != 3 {
		t.Fatalf("skipped token count = %d", skipped.Tokens().Count())
	}
	for i := 0; i < skipped.Tokens().Count(); i++ {
		diagnostics := skipped.Tokens().At(i).Diagnostics()
		if i == 0 {
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS198" {
				t.Fatalf("first skipped token diagnostics = %#v", diagnostics)
			}
		} else if diagnosticsContain(diagnostics, "KS198") {
			t.Fatalf("skipped token %d also has KS198: %#v", i, diagnostics)
		}
	}
	assertQueryGrammarProductionRejects(t, grammar.SkippedTokens, "")
}

func parseQueryGrammarStringLiteral(t *testing.T, parser ResultParser[LexicalToken, Expression], text string) *LiteralExpression {
	t.Helper()
	tokens := Lex(text)
	result := parser.ParseValue(NewArraySource(tokens), 0)
	literal, ok := result.Value.(*LiteralExpression)
	if result.Length < 0 || !ok {
		t.Fatalf("parse string literal %q failed: %#v", text, result)
	}
	if result.Length != len(tokens)-1 {
		t.Fatalf("parse string literal %q consumed %d of %d non-end tokens", text, result.Length, len(tokens)-1)
	}
	return literal
}

func parseQueryGrammarProduction[T any](t *testing.T, parser ResultParser[LexicalToken, T], text string) T {
	t.Helper()
	tokens := Lex(text)
	result := parser.ParseValue(NewArraySource(tokens), 0)
	if result.Length < 0 {
		t.Fatalf("parse production %q failed: %#v", text, result)
	}
	if result.Length != len(tokens)-1 {
		t.Fatalf("parse production %q consumed %d of %d non-end tokens", text, result.Length, len(tokens)-1)
	}
	return result.Value
}

func assertQueryGrammarProductionRejects[T any](t *testing.T, parser ResultParser[LexicalToken, T], text string) {
	t.Helper()
	result := parser.ParseValue(NewArraySource(Lex(text)), 0)
	if result.Length >= 0 {
		t.Fatalf("production unexpectedly parsed %q: %#v", text, result)
	}
}

func assertQueryGrammarProductionLeavesTail[T any](t *testing.T, parser ResultParser[LexicalToken, T], text string, wantLength int) {
	t.Helper()
	tokens := Lex(text)
	result := parser.ParseValue(NewArraySource(tokens), 0)
	if result.Length != wantLength {
		t.Fatalf("production prefix parse %q = %#v, want length %d", text, result, wantLength)
	}
	if result.Length == len(tokens)-1 {
		t.Fatalf("production unexpectedly consumed all of %q", text)
	}
}
