package kustolang

import "testing"

func TestSyntaxParsersCreateMissingTokensAndRequiredRecovery(t *testing.T) {
	missingName := CreateMissingSyntaxToken(SyntaxKindIdentifierToken)
	if missingName.Kind() != SyntaxKindIdentifierToken || !diagnosticsContain(missingName.Diagnostics(), "KS006") {
		t.Fatalf("missing name token = kind:%v diagnostics:%#v", missingName.Kind(), missingName.Diagnostics())
	}

	parser := RequiredSyntaxTokenTextParser("where")
	result := parser.ParseValue(NewArraySource(Lex("project")), 0)
	if result.Length != 0 || result.Value.Kind() != SyntaxKindWhereKeyword {
		t.Fatalf("required text recovery = length:%d token:%#v", result.Length, result.Value)
	}
	if !diagnosticsContain(result.Value.Diagnostics(), "KS005") {
		t.Fatalf("required text diagnostics = %#v", result.Value.Diagnostics())
	}

	missingClose := CreateMissingSyntaxToken(SyntaxKindCloseParenToken)
	if !diagnosticsContainMessage(missingClose.Diagnostics(), "KS005", "Expected: )") || diagnosticsContain(missingClose.Diagnostics(), "KS001") {
		t.Fatalf("missing close-token diagnostics = %#v", missingClose.Diagnostics())
	}

	kindParser := RequiredSyntaxTokenKindsParser([]SyntaxKind{SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken})
	kindResult := kindParser.ParseValue(NewArraySource(Lex("project")), 0)
	if kindResult.Length != 0 || kindResult.Value.Kind() != SyntaxKindOpenParenToken {
		t.Fatalf("required kind recovery = length:%d token:%#v", kindResult.Length, kindResult.Value)
	}
}

func TestSyntaxParsersTokenKindCompletionAnnotations(t *testing.T) {
	parser := SyntaxTokenKindParser(SyntaxKindProjectKeyword)
	result := parser.ParseValue(NewArraySource(Lex("project")), 0)
	if result.Length != 1 || result.Value.Kind() != SyntaxKindProjectKeyword {
		t.Fatalf("project token result = %#v", result)
	}

	items := completionItemsFromAnnotations(parser.Annotations())
	if len(items) != 1 || items[0].Kind != CompletionKindKeyword || items[0].DisplayText != "project" {
		t.Fatalf("completion items = %#v", items)
	}

	custom := WithSyntaxCompletion(parser, NewCompletionItem(CompletionKindExample, "custom"))
	items = completionItemsFromAnnotations(custom.Annotations())
	if len(items) != 1 || items[0].Kind != CompletionKindExample || items[0].DisplayText != "custom" {
		t.Fatalf("custom completion items = %#v", items)
	}

	hinted := WithSyntaxCompletionHint(custom, CompletionHintColumn)
	items = completionItemsFromAnnotations(hinted.Annotations())
	if len(items) != 1 || items[0].DisplayText != "custom" || completionHintFromAnnotations(hinted.Annotations()) != CompletionHintColumn {
		t.Fatalf("hinted annotations = %#v", hinted.Annotations())
	}

	hidden := HiddenSyntaxTokenKindParser(SyntaxKindProjectKeyword)
	if items := completionItemsFromAnnotations(hidden.Annotations()); len(items) != 0 {
		t.Fatalf("hidden token completions = %#v", items)
	}
}

func TestSyntaxParsersMatchAndProduceAdjacentTokenText(t *testing.T) {
	source := NewArraySource([]LexicalToken{
		NewLexicalToken(SyntaxKindBangToken, " ", "!"),
		NewLexicalToken(SyntaxKindIdentifierToken, "", "in"),
		NewLexicalToken(SyntaxKindIdentifierToken, " ", "x"),
	})

	if got := MatchesLexicalText(source, 0, "!in"); got != 2 {
		t.Fatalf("MatchesLexicalText adjacent = %d, want 2", got)
	}
	if got := MatchesLexicalText(source, 1, "inx"); got >= 0 {
		t.Fatalf("MatchesLexicalText should stop at inner trivia, got %d", got)
	}

	token := ProduceSyntaxToken(source, 0, 2)
	if token == nil || token.Kind() != SyntaxKindIdentifierToken || token.Trivia() != " " || token.RawText() != "!in" {
		t.Fatalf("produced identifier token = %#v", token)
	}

	other := ProduceSyntaxToken(source, 0, 2, SyntaxKindInputTextToken)
	if other == nil || other.Kind() != SyntaxKindInputTextToken || other.RawText() != "!in" {
		t.Fatalf("produced other token = %#v", other)
	}

	parser := SyntaxMatchTextParser("!in", SyntaxKindInputTextToken)
	result := parser.ParseValue(source, 0)
	if result.Length != 2 || result.Value.Kind() != SyntaxKindInputTextToken || result.Value.RawText() != "!in" {
		t.Fatalf("match text parser result = %#v", result)
	}

	if got := CombinedLexicalTokenText(source, 0, 3); got != "!in x" {
		t.Fatalf("combined token text = %q", got)
	}
	if got := CombinedLexicalTokenText(source, 0, 3, false); got != "!inx" {
		t.Fatalf("combined token text without trivia = %q", got)
	}
}

func TestSyntaxParsersParseAllAndScanFirst(t *testing.T) {
	parser := SyntaxTokenKindParser(SyntaxKindIdentifierToken)
	values := ParseAllSyntax(parser, "alpha beta | project")
	if len(values) != 2 || values[0].RawText() != "alpha" || values[1].RawText() != "beta" {
		t.Fatalf("ParseAllSyntax values = %#v", values)
	}

	first := ParseFirstSyntax(parser, "alpha")
	if first == nil || first.RawText() != "alpha" {
		t.Fatalf("ParseFirstSyntax = %#v", first)
	}

	if got := ScanFirstSyntax(EndOfTextSyntaxParser(), ""); got >= 0 {
		t.Fatalf("ScanFirstSyntax default end = %d, want failure", got)
	}
	if got := ScanFirstSyntax(EndOfTextSyntaxParser(), "", true); got != 1 {
		t.Fatalf("ScanFirstSyntax explicit end = %d, want 1", got)
	}
}

func TestSyntaxParsersListHelpersBuildSyntaxLists(t *testing.T) {
	identifier := SyntaxTokenKindParser(SyntaxKindIdentifierToken)
	listParser := SyntaxListParser(identifier, nil, false)
	list := listParser.ParseValue(NewArraySource(Lex("alpha beta")), 0)
	if list.Length != 2 || list.Value.Count() != 2 || list.Value.At(0).RawText() != "alpha" || list.Value.At(1).RawText() != "beta" {
		t.Fatalf("syntax list result = length:%d list:%#v", list.Length, list.Value)
	}
	if list.Value.ElementType() != "SyntaxToken" {
		t.Fatalf("syntax list element type = %q", list.Value.ElementType())
	}

	parserWithExamples := WithSyntaxExamples(identifier, "alpha", "beta")
	if items := completionItemsFromAnnotations(parserWithExamples.Annotations()); len(items) != 2 || items[0].Kind != CompletionKindExample || items[1].DisplayText != "beta" {
		t.Fatalf("example annotations = %#v", items)
	}
}

func TestSyntaxParsersSeparatedListHelpersRecoverMissingSeparators(t *testing.T) {
	identifier := SyntaxTokenKindParser(SyntaxKindIdentifierToken)
	missingIdentifier := func(Source[LexicalToken], int) *SyntaxToken {
		return CreateMissingSyntaxToken(SyntaxKindIdentifierToken)
	}

	parser := SyntaxCommaListParser(identifier, missingIdentifier, true, false)
	result := parser.ParseValue(NewArraySource(Lex("a,b)")), 0)
	if result.Length != 3 || result.Value.Count() != 2 {
		t.Fatalf("comma list result = length:%d list:%#v", result.Length, result.Value)
	}
	first := result.Value.At(0)
	second := result.Value.At(1)
	if first.Element().RawText() != "a" || first.Separator().Kind() != SyntaxKindCommaToken || second.Element().RawText() != "b" || second.Separator() != nil {
		t.Fatalf("comma list elements = %#v / %#v", first, second)
	}

	recovered := parser.ParseValue(NewArraySource(Lex("a b)")), 0)
	if recovered.Length != 2 || recovered.Value.Count() != 2 {
		t.Fatalf("recovered comma list = length:%d list:%#v", recovered.Length, recovered.Value)
	}
	recoveredFirst := recovered.Value.At(0)
	if recoveredFirst.Separator() == nil || recoveredFirst.Separator().Kind() != SyntaxKindCommaToken || !diagnosticsContain(recoveredFirst.Separator().Diagnostics(), "KS005") {
		t.Fatalf("recovered separator = %#v diagnostics:%#v", recoveredFirst.Separator(), recoveredFirst.Separator().Diagnostics())
	}

	trailing := SyntaxCommaListParser(identifier, missingIdentifier, true, true).ParseValue(NewArraySource(Lex("a,)")), 0)
	if trailing.Length != 2 || trailing.Value.Count() != 1 || trailing.Value.At(0).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("trailing comma list = length:%d list:%#v", trailing.Length, trailing.Value)
	}
}

func completionItemsFromAnnotations(annotations []any) []CompletionItem {
	items := []CompletionItem{}
	for _, annotation := range annotations {
		if item, ok := annotation.(CompletionItem); ok {
			items = append(items, item)
		}
	}
	return items
}

func completionHintFromAnnotations(annotations []any) CompletionHint {
	for _, annotation := range annotations {
		if hint, ok := annotation.(CompletionHint); ok {
			return hint
		}
	}
	return CompletionHintNone
}
