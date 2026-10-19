package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// SyntaxFactsAPI mirrors the upstream static SyntaxFacts class while
// preserving the return shapes of the existing public Go functions.
type SyntaxFactsAPI struct{}

func (SyntaxFactsAPI) GetText(kind SyntaxKind) string {
	return SyntaxText(kind)
}

func (SyntaxFactsAPI) GetCategory(kind SyntaxKind) SyntaxCategory {
	return SyntaxCategoryOf(kind)
}

func (SyntaxFactsAPI) IsKeyword(kind SyntaxKind) bool {
	return IsKeywordKind(kind)
}

func (SyntaxFactsAPI) IsPunctuation(kind SyntaxKind) bool {
	return IsPunctuationKind(kind)
}

func (SyntaxFactsAPI) IsOperator(kind SyntaxKind) bool {
	return IsOperatorKind(kind)
}

func (SyntaxFactsAPI) IsLiteral(kind SyntaxKind) bool {
	return IsLiteralKind(kind)
}

func (SyntaxFactsAPI) IsType(kind SyntaxKind) bool {
	return IsTypeKeyword(kind)
}

func (SyntaxFactsAPI) GetOperatorKind(kind SyntaxKind) OperatorKind {
	return SyntaxOperatorKind(kind)
}

func (SyntaxFactsAPI) GetKind(text string) (SyntaxKind, bool) {
	return SyntaxKindFromText(text)
}

func (SyntaxFactsAPI) GetKindFromRange(text string, offset, length int) (SyntaxKind, bool) {
	return SyntaxKindFromTextRange(text, offset, length)
}

func (SyntaxFactsAPI) GetKinds(category SyntaxCategory) []SyntaxKind {
	return SyntaxKinds(category)
}

func (SyntaxFactsAPI) GetKindsWithFixedText() []SyntaxKind {
	return SyntaxKindsWithFixedText()
}

func (SyntaxFactsAPI) CanBeIdentifier(kind SyntaxKind) bool {
	return SyntaxKindCanBeIdentifier(kind)
}

// IsKeywordText is the Go spelling of the upstream IsKeyword(string)
// overload, which cannot share a name with IsKeyword(SyntaxKind) in Go.
func (SyntaxFactsAPI) IsKeywordText(text string) bool {
	return IsKeywordText(text)
}

func (SyntaxFactsAPI) IsKeywordThatCanBeIdentifier(text string) bool {
	return IsKeywordThatCanBeIdentifier(text)
}

func (SyntaxFactsAPI) Keywords() []string {
	return SyntaxKeywordTexts()
}

func (SyntaxFactsAPI) Punctuation() []string {
	return SyntaxPunctuationTexts()
}

func (SyntaxFactsAPI) Operators() []string {
	return SyntaxOperatorTexts()
}

// SyntaxFacts exposes source-shaped access to the upstream static API.
var SyntaxFacts SyntaxFactsAPI
