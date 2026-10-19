package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxParserFacadeCombinesAdjacentTokens(t *testing.T) {
	tokens := []kustolang.LexicalToken{
		kustolang.NewLexicalToken(kustolang.SyntaxKindBangToken, " ", "!"),
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "in"),
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, " ", "value"),
		kustolang.NewLexicalToken(kustolang.SyntaxKindEndOfTextToken, "", ""),
	}
	source := kustolang.NewArraySource(tokens)

	if got := kustolang.MatchesLexicalText(source, 0, "!in"); got != 2 {
		t.Fatalf("MatchesLexicalText() = %d, want 2", got)
	}
	if got := kustolang.MatchesLexicalText(source, 1, "invalue"); got >= 0 {
		t.Fatalf("MatchesLexicalText() crossed trivia with result %d", got)
	}

	produced := kustolang.ProduceSyntaxToken(source, 0, 2)
	if produced == nil || produced.Kind() != kustolang.SyntaxKindIdentifierToken || produced.Trivia() != " " || produced.RawText() != "!in" {
		t.Fatalf("ProduceSyntaxToken() = %#v", produced)
	}
	producedAsInput := kustolang.ProduceSyntaxTokenFromText(source, 0, 2, "!in", kustolang.SyntaxKindInputTextToken)
	if producedAsInput == nil || producedAsInput.Kind() != kustolang.SyntaxKindInputTextToken || producedAsInput.RawText() != "!in" {
		t.Fatalf("ProduceSyntaxTokenFromText() = %#v", producedAsInput)
	}

	if got := kustolang.CombinedLexicalTokenText(source, 0, 3); got != "!in value" {
		t.Fatalf("CombinedLexicalTokenText() = %q", got)
	}
	if got := kustolang.CombinedLexicalTokenText(source, 0, 3, false); got != "!invalue" {
		t.Fatalf("CombinedLexicalTokenText(..., false) = %q", got)
	}

	match := kustolang.SyntaxMatchTextParser("!in", kustolang.SyntaxKindInputTextToken).ParseValue(source, 0)
	if match.Length != 2 || match.Value == nil || match.Value.Kind() != kustolang.SyntaxKindInputTextToken || match.Value.RawText() != "!in" {
		t.Fatalf("SyntaxMatchTextParser() = %#v", match)
	}
	texts := kustolang.SyntaxTokenTextsParser([]string{"project", "!in"}).ParseValue(source, 0)
	if texts.Length != 2 || texts.Value == nil || texts.Value.RawText() != "!in" {
		t.Fatalf("SyntaxTokenTextsParser() = %#v", texts)
	}

	if result := kustolang.AnySyntaxTokenButEndParser().ParseValue(source, 0); result.Length != 1 || result.Value.RawText() != "!" {
		t.Fatalf("AnySyntaxTokenButEndParser() token result = %#v", result)
	}
	if result := kustolang.AnySyntaxTokenButEndParser().ParseValue(source, 3); result.Length >= 0 || result.Value != nil {
		t.Fatalf("AnySyntaxTokenButEndParser() end result = %#v", result)
	}

	identifiers := kustolang.ParseAllSyntaxTokens(kustolang.SyntaxTokenKindParser(kustolang.SyntaxKindIdentifierToken), tokens[1:])
	if len(identifiers) != 2 || identifiers[0].RawText() != "in" || identifiers[1].RawText() != "value" {
		t.Fatalf("ParseAllSyntaxTokens() = %#v", identifiers)
	}
	if got := kustolang.ScanFirstSyntax(kustolang.EndOfTextSyntaxParser(), ""); got >= 0 {
		t.Fatalf("ScanFirstSyntax() = %d, want failure without explicit EOF", got)
	}
	if got := kustolang.ScanFirstSyntax(kustolang.EndOfTextSyntaxParser(), "", true); got != 1 {
		t.Fatalf("ScanFirstSyntax(..., true) = %d, want 1", got)
	}
}

func TestPublicSyntaxParserFacadeCompletionHelpers(t *testing.T) {
	item, ok := kustolang.CreateSyntaxCompletionItemForKind(
		kustolang.SyntaxKindProjectKeyword,
		kustolang.CompletionKindKeyword,
		kustolang.CompletionPriorityNormal,
	)
	if !ok || item.DisplayText != "project" || item.Kind != kustolang.CompletionKindKeyword {
		t.Fatalf("CreateSyntaxCompletionItemForKind() = %#v, %t", item, ok)
	}
	if _, ok := kustolang.CreateSyntaxCompletionItemForText("_hidden", kustolang.CompletionKindKeyword, kustolang.CompletionPriorityNormal); ok {
		t.Fatal("CreateSyntaxCompletionItemForText() exposed underscore-prefixed syntax")
	}
	if got := kustolang.SyntaxCompletionKindForKind(kustolang.SyntaxKindCommaToken); got != kustolang.CompletionKindPunctuation {
		t.Fatalf("SyntaxCompletionKindForKind() = %v", got)
	}
	if got := kustolang.SyntaxCompletionKindForText("project"); got != kustolang.CompletionKindKeyword {
		t.Fatalf("SyntaxCompletionKindForText() = %v", got)
	}

	parser := kustolang.WithSyntaxCompletion(kustolang.HiddenSyntaxTokenKindParser(kustolang.SyntaxKindProjectKeyword), item)
	parser = kustolang.WithSyntaxCompletionHint(parser, kustolang.CompletionHint(1))
	parser = kustolang.WithSyntaxExamples(parser, "project Column")
	items := syntaxCompletionItems(parser.Annotations())
	if len(items) != 2 || items[0].DisplayText != "project" || items[1].Kind != kustolang.CompletionKindExample {
		t.Fatalf("completion annotations = %#v", parser.Annotations())
	}

	if result := kustolang.HiddenSyntaxTokenTextParser("project").ParseValue(kustolang.NewArraySource(kustolang.Lex("project")), 0); result.Length != 1 {
		t.Fatalf("HiddenSyntaxTokenTextParser() = %#v", result)
	}
	if result := kustolang.HiddenSyntaxTokenTextsParser([]string{"project", "where"}).ParseValue(kustolang.NewArraySource(kustolang.Lex("where")), 0); result.Length != 1 {
		t.Fatalf("HiddenSyntaxTokenTextsParser() = %#v", result)
	}
}

func TestPublicSyntaxParserFacadeRecoversMissingTokensAndLists(t *testing.T) {
	missingKinds := kustolang.CreateMissingSyntaxTokenForKinds([]kustolang.SyntaxKind{
		kustolang.SyntaxKindOpenParenToken,
		kustolang.SyntaxKindOpenBracketToken,
	})
	missingText := kustolang.CreateMissingSyntaxTokenForText(")")
	missingTexts := kustolang.CreateMissingSyntaxTokenForTexts([]string{"project", "where"})
	for name, token := range map[string]*kustolang.SyntaxToken{
		"kinds": missingKinds,
		"text":  missingText,
		"texts": missingTexts,
	} {
		if token == nil || !token.IsMissing() || !hasSyntaxDiagnosticCode(token.Diagnostics(), "KS005") {
			t.Fatalf("missing %s token = %#v, diagnostics = %#v", name, token, token.Diagnostics())
		}
	}

	projectSource := kustolang.NewArraySource(kustolang.Lex("project"))
	requiredKind := kustolang.RequiredSyntaxTokenKindsParser([]kustolang.SyntaxKind{
		kustolang.SyntaxKindOpenParenToken,
		kustolang.SyntaxKindOpenBracketToken,
	}).ParseValue(projectSource, 0)
	if requiredKind.Length != 0 || requiredKind.Value.Kind() != kustolang.SyntaxKindOpenParenToken || !requiredKind.Value.IsMissing() {
		t.Fatalf("RequiredSyntaxTokenKindsParser() = %#v", requiredKind)
	}
	requiredText := kustolang.RequiredSyntaxTokenTextsParser([]string{"where", "project-away"}).ParseValue(projectSource, 0)
	if requiredText.Length != 0 || !requiredText.Value.IsMissing() || !hasSyntaxDiagnosticCode(requiredText.Value.Diagnostics(), "KS005") {
		t.Fatalf("RequiredSyntaxTokenTextsParser() = %#v", requiredText)
	}

	identifier := kustolang.SyntaxTokenKindParser(kustolang.SyntaxKindIdentifierToken)
	missingIdentifier := func(kustolang.Source[kustolang.LexicalToken], int) *kustolang.SyntaxToken {
		return kustolang.CreateMissingSyntaxToken(kustolang.SyntaxKindIdentifierToken)
	}

	listResult := kustolang.SyntaxListParser(identifier, missingIdentifier, true).
		ParseValue(kustolang.NewArraySource(kustolang.Lex(")")), 0)
	if listResult.Length != 0 || listResult.Value.Count() != 1 || !listResult.Value.At(0).IsMissing() {
		t.Fatalf("SyntaxListParser() recovery = %#v", listResult)
	}

	commaParser := kustolang.SyntaxCommaListParser(identifier, missingIdentifier, true, false)
	missingComma := commaParser.ParseValue(kustolang.NewArraySource(kustolang.Lex("a b)")), 0)
	if missingComma.Length != 2 || missingComma.Value.Count() != 2 {
		t.Fatalf("missing-comma list = %#v", missingComma)
	}
	separator := missingComma.Value.At(0).Separator()
	if separator == nil || !separator.IsMissing() || separator.Kind() != kustolang.SyntaxKindCommaToken || !hasSyntaxDiagnosticCode(separator.Diagnostics(), "KS005") {
		t.Fatalf("recovered comma = %#v, diagnostics = %#v", separator, separator.Diagnostics())
	}

	missingElement := commaParser.ParseValue(kustolang.NewArraySource(kustolang.Lex("a,,b)")), 0)
	if missingElement.Length != 4 || missingElement.Value.Count() != 3 {
		t.Fatalf("missing-element list = %#v", missingElement)
	}
	recoveredElement := missingElement.Value.At(1).Element()
	if recoveredElement == nil || !recoveredElement.IsMissing() || !hasSyntaxDiagnosticCode(recoveredElement.Diagnostics(), "KS006") {
		t.Fatalf("recovered element = %#v, diagnostics = %#v", recoveredElement, recoveredElement.Diagnostics())
	}

	fullParser := kustolang.SyntaxSeparatedListFullParser(
		identifier,
		kustolang.SyntaxKindCommaToken,
		identifier,
		missingIdentifier,
		kustolang.EndOfCommaListParser(),
		true,
		false,
	)
	if result := fullParser.ParseValue(kustolang.NewArraySource(kustolang.Lex("x,y)")), 0); result.Length != 3 || result.Value.Count() != 2 {
		t.Fatalf("SyntaxSeparatedListFullParser() = %#v", result)
	}
	if result := kustolang.SyntaxSeparatedListParser(identifier, kustolang.SyntaxKindCommaToken, missingIdentifier, kustolang.EndOfCommaListParser(), true, false).
		ParseValue(kustolang.NewArraySource(kustolang.Lex("x,y)")), 0); result.Length != 3 || result.Value.Count() != 2 {
		t.Fatalf("SyntaxSeparatedListParser() = %#v", result)
	}
	if result := kustolang.SyntaxOneOrMoreCommaListParser(identifier, missingIdentifier).
		ParseValue(kustolang.NewArraySource(kustolang.Lex("x,y)")), 0); result.Length != 3 || result.Value.Count() != 2 {
		t.Fatalf("SyntaxOneOrMoreCommaListParser() = %#v", result)
	}

	first := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "left"))
	comma := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindCommaToken, "", ","))
	second := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "right"))
	made := kustolang.MakeSyntaxSeparatedList[*kustolang.SyntaxToken]([]any{first, comma, second})
	if made.Count() != 2 || made.At(0).Element().RawText() != "left" || made.At(0).Separator().Kind() != kustolang.SyntaxKindCommaToken || made.At(1).Element().RawText() != "right" {
		t.Fatalf("MakeSyntaxSeparatedList() = %#v", made)
	}
}

func syntaxCompletionItems(annotations []any) []kustolang.CompletionItem {
	items := make([]kustolang.CompletionItem, 0)
	for _, annotation := range annotations {
		if item, ok := annotation.(kustolang.CompletionItem); ok {
			items = append(items, item)
		}
	}
	return items
}

func hasSyntaxDiagnosticCode(diagnostics []kustolang.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
