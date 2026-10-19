package kustolang

import "testing"

func TestCommandGrammarCustomDescriptorsAndCommand(t *testing.T) {
	defaultResult := CommandGrammarCustomElements(CommandGrammarToken("show")).
		ParseValue(NewArraySource(Lex("show")), 0)
	defaultCustom, ok := defaultResult.Value.(*CustomNode)
	if defaultResult.Length != 1 || !ok || len(defaultCustom.Shape()) != 1 || !defaultCustom.Shape()[0].IsDefault() {
		t.Fatalf("default custom = %#v", defaultResult.Value)
	}

	shape := []CustomElementDescriptor{
		CommandGrammarCD("Verb"),
		CommandGrammarCD("Target", CommandGrammarDescriptorOptions{Hint: CompletionHintTable, IsOptional: true}),
	}
	content := CommandGrammarCustomWithShape(
		shape,
		CommandGrammarToken("show"),
		CommandGrammarToken("widgets"),
	)
	result := content.ParseValue(NewArraySource(Lex("show widgets")), 0)
	if result.Length != 2 {
		t.Fatalf("custom length = %d", result.Length)
	}
	custom, ok := result.Value.(*CustomNode)
	if !ok || custom.ChildCount() != 2 {
		t.Fatalf("custom = %#v", result.Value)
	}
	if custom.GetName(0) != "Verb" || custom.CompletionHint(1) != CompletionHintTable || !custom.IsOptional(1) {
		t.Fatalf("shape = %#v", custom.Shape())
	}

	commandParser := CommandGrammarCommand("ShowWidgets", content)
	commandResult := commandParser.ParseValue(NewArraySource(Lex(".show widgets")), 0)
	command, ok := commandResult.Value.(*CustomCommand)
	if commandResult.Length != 3 || !ok || command.CommandKind() != "ShowWidgets" || command.DotToken().Kind() != SyntaxKindDotToken {
		t.Fatalf("command = %#v, length = %d", commandResult.Value, commandResult.Length)
	}
	if commandParser.Tag() != "<ShowWidgets>" || command.Custom().Text(IncludeTriviaAll) != "show widgets" {
		t.Fatalf("tag = %q, custom = %#v", commandParser.Tag(), command.Custom())
	}
}

func TestCommandGrammarTokenCompletionAndRecovery(t *testing.T) {
	keyword := CommandGrammarToken("widgets")
	if !parserHasCompletionKind(keyword, "widgets", CompletionKindKeyword) {
		t.Fatalf("keyword annotations = %#v", keyword.Annotations())
	}
	example := CommandGrammarTokenWithCompletion("widgets", CompletionKindExample)
	if !parserHasCompletionKind(example, "widgets", CompletionKindExample) {
		t.Fatalf("example annotations = %#v", example.Annotations())
	}
	comma := CommandGrammarToken(",")
	if !parserHasCompletionKind(comma, ",", CompletionKindPunctuation) {
		t.Fatalf("comma annotations = %#v", comma.Annotations())
	}

	combined := CommandGrammarToken("project-away", SyntaxKindIdentifierToken).
		ParseValue(NewArraySource(Lex("project-away")), 0)
	combinedToken, ok := combined.Value.(*SyntaxToken)
	if combined.Length < 1 || !ok || combinedToken.RawText() != "project-away" || combinedToken.Kind() != SyntaxKindIdentifierToken {
		t.Fatalf("combined token = %#v", combined)
	}

	missing := CommandGrammarRequiredToken(")").ParseValue(NewArraySource(Lex("other")), 0)
	missingToken, ok := missing.Value.(*SyntaxToken)
	if missing.Length != 0 || !ok || !missingToken.IsMissing() || missingToken.Kind() != SyntaxKindCloseParenToken || !hasDiagnosticCode(missingToken.Diagnostics(), "KS005") {
		t.Fatalf("missing token = %#v, diagnostics = %#v", missing.Value, missing.Value.Diagnostics())
	}
	alternatives := CommandGrammarRequiredTokens("hot", "cold").ParseValue(NewArraySource(Lex("other")), 0)
	alternativeToken := alternatives.Value.(*SyntaxToken)
	if alternativeToken.Kind() != SyntaxKindIdentifierToken || !alternativeToken.IsMissing() {
		t.Fatalf("alternative token = %#v", alternativeToken)
	}
}

func TestCommandGrammarSeparatedListRecovery(t *testing.T) {
	element := CommandGrammarTokens("a", "b")
	missingCalls := 0
	missing := func(Source[LexicalToken], int) SyntaxElement {
		missingCalls++
		return CreateMissingSyntaxToken(SyntaxKindIdentifierToken)
	}
	parser := CommandGrammarOneOrMoreCommaList(element, missing, false)
	result := parser.ParseValue(NewArraySource(Lex("a,,b")), 0)
	list, ok := result.Value.(*SyntaxList)
	if result.Length != 4 || !ok || list.Count() != 3 || missingCalls != 1 {
		t.Fatalf("recovered list = %#v, length = %d, calls = %d", result.Value, result.Length, missingCalls)
	}
	middle := list.At(1).(*SeparatedElement)
	if middle.Element() == nil || !middle.Element().IsMissing() || middle.Separator() == nil {
		t.Fatalf("middle = %#v", middle)
	}

	leading := CommandGrammarZeroOrMoreCommaList(element, missing, false).
		ParseValue(NewArraySource(Lex(",a")), 0)
	if leading.Length != 0 || leading.Value.(*SyntaxList).Count() != 0 || missingCalls != 1 {
		t.Fatalf("leading separator recovery = %#v, calls = %d", leading, missingCalls)
	}

	recoveredTrailing := CommandGrammarOneOrMoreCommaList(element, missing, false).
		ParseValue(NewArraySource(Lex("a,")), 0)
	recoveredTrailingList := recoveredTrailing.Value.(*SyntaxList)
	if recoveredTrailing.Length != 2 || recoveredTrailingList.Count() != 2 || !recoveredTrailingList.At(1).(*SeparatedElement).Element().IsMissing() || missingCalls != 2 {
		t.Fatalf("recovered trailing list = %#v, calls = %d", recoveredTrailing, missingCalls)
	}

	trailing := CommandGrammarOneOrMoreCommaList(element, missing, true).
		ParseValue(NewArraySource(Lex("a,")), 0)
	trailingList := trailing.Value.(*SyntaxList)
	if trailing.Length != 2 || trailingList.Count() != 1 || trailingList.At(0).(*SeparatedElement).Separator() == nil || missingCalls != 2 {
		t.Fatalf("trailing list = %#v, calls = %d", trailing, missingCalls)
	}
}

func TestCommandGrammarUnseparatedListIgnoresMissingFactory(t *testing.T) {
	element := CommandGrammarToken("a")
	missingCalls := 0
	missing := func(Source[LexicalToken], int) SyntaxElement {
		missingCalls++
		return CreateMissingSyntaxToken(SyntaxKindIdentifierToken)
	}

	one := CommandGrammarOneOrMoreList(element, nil, missing, false)
	empty := one.ParseValue(NewArraySource(Lex("b")), 0)
	if empty.Length >= 0 || missingCalls != 0 {
		t.Fatalf("one-or-more empty = %#v, calls = %d", empty, missingCalls)
	}
	items := one.ParseValue(NewArraySource(Lex("a a b")), 0)
	if items.Length != 2 || items.Value.(*SyntaxList).Count() != 2 || missingCalls != 0 {
		t.Fatalf("unseparated = %#v, calls = %d", items, missingCalls)
	}

	zero := CommandGrammarZeroOrMoreList(element, nil, missing, false).
		ParseValue(NewArraySource(Lex("b")), 0)
	if zero.Length != 0 || zero.Value.(*SyntaxList).Count() != 0 || missingCalls != 0 {
		t.Fatalf("zero-or-more empty = %#v, calls = %d", zero, missingCalls)
	}
}

func parserHasCompletionKind(parser Parser[LexicalToken], text string, kind CompletionKind) bool {
	for _, annotation := range parser.Annotations() {
		if item, ok := annotation.(CompletionItem); ok && item.DisplayText == text && item.Kind == kind {
			return true
		}
	}
	for i := 0; i < parser.ChildParserCount(); i++ {
		if parserHasCompletionKind(parser.ChildParser(i), text, kind) {
			return true
		}
	}
	return false
}
