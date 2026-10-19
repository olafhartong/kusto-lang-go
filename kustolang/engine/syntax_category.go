package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxCategory.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type SyntaxCategory int

const (
	SyntaxCategoryNone SyntaxCategory = iota
	SyntaxCategoryKeyword
	SyntaxCategoryIdentifier
	SyntaxCategoryPunctuation
	SyntaxCategoryOperator
	SyntaxCategoryLiteral
	SyntaxCategoryList
	SyntaxCategoryNode
	SyntaxCategoryOther
)

func (c SyntaxCategory) String() string {
	switch c {
	case SyntaxCategoryNone:
		return "None"
	case SyntaxCategoryKeyword:
		return "Keyword"
	case SyntaxCategoryIdentifier:
		return "Identifier"
	case SyntaxCategoryPunctuation:
		return "Punctuation"
	case SyntaxCategoryOperator:
		return "Operator"
	case SyntaxCategoryLiteral:
		return "Literal"
	case SyntaxCategoryList:
		return "List"
	case SyntaxCategoryNode:
		return "Node"
	case SyntaxCategoryOther:
		return "Other"
	default:
		return "Unknown"
	}
}
