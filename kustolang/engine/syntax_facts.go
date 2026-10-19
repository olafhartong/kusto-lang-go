package kustolang

import (
	"sort"
)

// Ported from src/Kusto.Language/Syntax/SyntaxFacts.cs. This table is the
// initial subset used by the incremental lexer/parser port.

type syntaxData struct {
	Kind            SyntaxKind
	Text            string
	Category        SyntaxCategory
	OperatorKind    OperatorKind
	CanBeIdentifier bool
	IsType          bool
}

var syntaxByKind map[SyntaxKind]syntaxData
var syntaxByText map[string]syntaxData
var syntaxTextsByLength []syntaxData
var syntaxTextsByFirstByte [256][]syntaxData

func init() {
	syntaxByKind = map[SyntaxKind]syntaxData{}
	syntaxByText = map[string]syntaxData{}
	for _, data := range syntaxFacts {
		syntaxByKind[data.Kind] = data
		if data.Text != "" {
			syntaxByText[data.Text] = data
			syntaxTextsByLength = append(syntaxTextsByLength, data)
		}
	}
	sort.SliceStable(syntaxTextsByLength, func(i, j int) bool {
		return len(syntaxTextsByLength[i].Text) > len(syntaxTextsByLength[j].Text)
	})
	// Preserve longest-token-first precedence while avoiding a scan of every
	// keyword and punctuation token at each source position.
	for _, data := range syntaxTextsByLength {
		syntaxTextsByFirstByte[data.Text[0]] = append(syntaxTextsByFirstByte[data.Text[0]], data)
	}
	initializeKustoFactsSyntaxKinds()
}

func SyntaxText(kind SyntaxKind) string {
	if kind == SyntaxKindGraphMergeKeyword {
		return "graph-merge"
	}
	return syntaxByKind[kind].Text
}

func SyntaxCategoryOf(kind SyntaxKind) SyntaxCategory {
	switch kind {
	case SyntaxKindGraphMergeKeyword:
		return SyntaxCategoryKeyword
	case SyntaxKindGraphMergeOperator:
		return SyntaxCategoryNode
	}
	if data, ok := syntaxByKind[kind]; ok {
		return data.Category
	}
	return SyntaxCategoryOther
}

func SyntaxOperatorKind(kind SyntaxKind) OperatorKind {
	if data, ok := syntaxByKind[kind]; ok {
		return data.OperatorKind
	}
	return OperatorKindNone
}

func SyntaxKindFromText(text string) (SyntaxKind, bool) {
	data, ok := syntaxByText[text]
	return data.Kind, ok
}

func SyntaxKindFromTextRange(text string, offset int, length int) (SyntaxKind, bool) {
	if offset < 0 || length < 0 || offset > len(text) || offset+length > len(text) {
		return SyntaxKindNone, false
	}
	return SyntaxKindFromText(text[offset : offset+length])
}

func SyntaxKinds(category SyntaxCategory) []SyntaxKind {
	kinds := make([]SyntaxKind, 0)
	for _, data := range syntaxFacts {
		if data.Category == category {
			kinds = append(kinds, data.Kind)
		}
	}
	return kinds
}

func SyntaxKindsWithFixedText() []SyntaxKind {
	kinds := make([]SyntaxKind, 0)
	for _, data := range syntaxFacts {
		if data.Text != "" {
			kinds = append(kinds, data.Kind)
		}
	}
	return kinds
}

func SyntaxKeywordTexts() []string {
	return syntaxTextsForCategory(SyntaxCategoryKeyword)
}

func SyntaxPunctuationTexts() []string {
	return syntaxTextsForCategory(SyntaxCategoryPunctuation)
}

func SyntaxOperatorTexts() []string {
	return syntaxTextsForCategory(SyntaxCategoryOperator)
}

func syntaxTextsForCategory(category SyntaxCategory) []string {
	texts := make([]string, 0)
	for _, kind := range SyntaxKinds(category) {
		if text := SyntaxText(kind); text != "" {
			texts = append(texts, text)
		}
	}
	return texts
}

func IsKeywordKind(kind SyntaxKind) bool {
	return SyntaxCategoryOf(kind) == SyntaxCategoryKeyword
}

func IsPunctuationKind(kind SyntaxKind) bool {
	return SyntaxCategoryOf(kind) == SyntaxCategoryPunctuation
}

func IsOperatorKind(kind SyntaxKind) bool {
	return SyntaxCategoryOf(kind) == SyntaxCategoryOperator
}

func IsLiteralKind(kind SyntaxKind) bool {
	return SyntaxCategoryOf(kind) == SyntaxCategoryLiteral
}

func IsTypeKeyword(kind SyntaxKind) bool {
	return syntaxByKind[kind].IsType
}

func SyntaxKindCanBeIdentifier(kind SyntaxKind) bool {
	return syntaxByKind[kind].CanBeIdentifier
}

func IsKeywordText(text string) bool {
	kind, ok := SyntaxKindFromText(text)
	return ok && IsKeywordKind(kind)
}

func IsKeywordThatCanBeIdentifier(text string) bool {
	kind, ok := SyntaxKindFromText(text)
	return ok && IsKeywordKind(kind) && SyntaxKindCanBeIdentifier(kind)
}

func IsEndOfText(kind SyntaxKind) bool {
	return kind == SyntaxKindEndOfTextToken
}
