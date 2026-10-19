package kustolang

import "strings"

// Ported from:
//   src/Kusto.Language/Editor/FormattingOptions.cs
//   src/Kusto.Language/Editor/OutliningOptions.cs
//   src/Kusto.Language/Editor/QuickInfo.cs
//   src/Kusto.Language/Editor/RelatedInfo.cs
//   src/Kusto.Language/Editor/FindRelatedOptions.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type FormattingOptions struct {
	IndentationSize                int
	InsertMissingTokens            bool
	BrackettingStyle               BrackettingStyle
	SchemaStyle                    BrackettingStyle
	DataTableValueStyle            BrackettingStyle
	FunctionBodyStyle              BrackettingStyle
	FunctionParameterStyle         BrackettingStyle
	FunctionArgumentStyle          BrackettingStyle
	PipeOperatorStyle              PlacementStyle
	ExpressionStyle                PlacementStyle
	StatementStyle                 PlacementStyle
	SemicolonStyle                 PlacementStyle
	GeneralSpacing                 SpacingStyle
	PrefixOperatorSpacing          SpacingStyle
	InfixOperatorSpacing           DualSpacingStyle
	PipeOperatorSpacing            DualSpacingStyle
	CommaSpacing                   DualSpacingStyle
	ColonSpacing                   DualSpacingStyle
	AssignmentSpacing              DualSpacingStyle
	RangeOperatorSpacing           DualSpacingStyle
	SemicolonSpacing               DualSpacingStyle
	ParenthesizedExpressionSpacing SpacingStyle
	ArgumentListSpacing            SpacingStyle
	EmptyArgumentListSpacing       SpacingStyle
	ParameterListSpacing           SpacingStyle
	EmptyParameterListSpacing      SpacingStyle
	JsonArraySpacing               SpacingStyle
	EmptyJsonArraySpacing          SpacingStyle
	JsonObjectSpacing              SpacingStyle
	EmptyJsonObjectSpacing         SpacingStyle
	DataTableValueSpacing          SpacingStyle
	EmptyDataTableValueSpacing     SpacingStyle
	FunctionBodySpacing            SpacingStyle
	EmptyFunctionBodySpacing       SpacingStyle
	BeforeFunctionBodySpacing      SpacingStyle
	BeforeParameterListSpacing     SpacingStyle
	BeforeArgumentListSpacing      SpacingStyle
	BeforeDataTableValueSpacing    SpacingStyle
}

var FormattingOptionsDefault = FormattingOptions{
	IndentationSize:                4,
	InsertMissingTokens:            true,
	BrackettingStyle:               BrackettingStyleVertical,
	SchemaStyle:                    BrackettingStyleDefault,
	DataTableValueStyle:            BrackettingStyleDefault,
	FunctionBodyStyle:              BrackettingStyleDefault,
	FunctionParameterStyle:         BrackettingStyleDefault,
	FunctionArgumentStyle:          BrackettingStyleDiagonal,
	PipeOperatorStyle:              PlacementStyleSmart,
	ExpressionStyle:                PlacementStyleSmart,
	StatementStyle:                 PlacementStyleNewLine,
	SemicolonStyle:                 PlacementStyleNone,
	GeneralSpacing:                 SpacingStyleOne,
	PrefixOperatorSpacing:          SpacingStyleMinimal,
	InfixOperatorSpacing:           DualSpacingStyleBoth,
	PipeOperatorSpacing:            DualSpacingStyleBoth,
	CommaSpacing:                   DualSpacingStyleAfter,
	ColonSpacing:                   DualSpacingStyleAfter,
	AssignmentSpacing:              DualSpacingStyleBoth,
	RangeOperatorSpacing:           DualSpacingStyleBoth,
	SemicolonSpacing:               DualSpacingStyleAfter,
	ParenthesizedExpressionSpacing: SpacingStyleMinimal,
	ArgumentListSpacing:            SpacingStyleMinimal,
	EmptyArgumentListSpacing:       SpacingStyleMinimal,
	ParameterListSpacing:           SpacingStyleMinimal,
	EmptyParameterListSpacing:      SpacingStyleMinimal,
	JsonArraySpacing:               SpacingStyleMinimal,
	EmptyJsonArraySpacing:          SpacingStyleMinimal,
	JsonObjectSpacing:              SpacingStyleMinimal,
	EmptyJsonObjectSpacing:         SpacingStyleMinimal,
	DataTableValueSpacing:          SpacingStyleMinimal,
	EmptyDataTableValueSpacing:     SpacingStyleMinimal,
	FunctionBodySpacing:            SpacingStyleOne,
	EmptyFunctionBodySpacing:       SpacingStyleOne,
	BeforeFunctionBodySpacing:      SpacingStyleOne,
	BeforeParameterListSpacing:     SpacingStyleMinimal,
	BeforeArgumentListSpacing:      SpacingStyleMinimal,
	BeforeDataTableValueSpacing:    SpacingStyleOne,
}

func (options FormattingOptions) WithIndentationSize(size int) FormattingOptions {
	options.IndentationSize = size
	return options
}

func (options FormattingOptions) WithInsertMissingTokens(enable bool) FormattingOptions {
	options.InsertMissingTokens = enable
	return options
}

func (options FormattingOptions) WithBrackettingStyle(style BrackettingStyle) FormattingOptions {
	options.BrackettingStyle = style
	return options
}

func (options FormattingOptions) WithSchemaStyle(style BrackettingStyle) FormattingOptions {
	options.SchemaStyle = style
	return options
}

func (options FormattingOptions) WithDataTableValueStyle(style BrackettingStyle) FormattingOptions {
	options.DataTableValueStyle = style
	return options
}

func (options FormattingOptions) WithFunctionBodyStyle(style BrackettingStyle) FormattingOptions {
	options.FunctionBodyStyle = style
	return options
}

func (options FormattingOptions) WithFunctionParameterStyle(style BrackettingStyle) FormattingOptions {
	options.FunctionParameterStyle = style
	return options
}

func (options FormattingOptions) WithFunctionArgumentStyle(style BrackettingStyle) FormattingOptions {
	options.FunctionArgumentStyle = style
	return options
}

func (options FormattingOptions) WithPipeOperatorStyle(style PlacementStyle) FormattingOptions {
	options.PipeOperatorStyle = style
	return options
}

func (options FormattingOptions) WithExpressionStyle(style PlacementStyle) FormattingOptions {
	options.ExpressionStyle = style
	return options
}

func (options FormattingOptions) WithStatementStyle(style PlacementStyle) FormattingOptions {
	options.StatementStyle = style
	return options
}

func (options FormattingOptions) WithSemicolonStyle(style PlacementStyle) FormattingOptions {
	options.SemicolonStyle = style
	return options
}

func (options FormattingOptions) WithGeneralSpacing(style SpacingStyle) FormattingOptions {
	options.GeneralSpacing = style
	return options
}

func (options FormattingOptions) WithPrefixOperatorSpacing(style SpacingStyle) FormattingOptions {
	options.PrefixOperatorSpacing = style
	return options
}

func (options FormattingOptions) WithInfixOperatorSpacing(style DualSpacingStyle) FormattingOptions {
	options.InfixOperatorSpacing = style
	return options
}

func (options FormattingOptions) WithPipeOperatorSpacing(style DualSpacingStyle) FormattingOptions {
	options.PipeOperatorSpacing = style
	return options
}

func (options FormattingOptions) WithCommaSpacing(style DualSpacingStyle) FormattingOptions {
	options.CommaSpacing = style
	return options
}

func (options FormattingOptions) WithColonSpacing(style DualSpacingStyle) FormattingOptions {
	options.ColonSpacing = style
	return options
}

func (options FormattingOptions) WithAssignmentSpacing(style DualSpacingStyle) FormattingOptions {
	options.AssignmentSpacing = style
	return options
}

func (options FormattingOptions) WithRangeOperatorSpacing(style DualSpacingStyle) FormattingOptions {
	options.RangeOperatorSpacing = style
	return options
}

func (options FormattingOptions) WithSemicolonSpacing(style DualSpacingStyle) FormattingOptions {
	options.SemicolonSpacing = style
	return options
}

func (options FormattingOptions) WithExpressionParenSpacing(style SpacingStyle) FormattingOptions {
	options.ParenthesizedExpressionSpacing = style
	return options
}

func (options FormattingOptions) WithArgumentListSpacing(style SpacingStyle) FormattingOptions {
	options.ArgumentListSpacing = style
	return options
}

func (options FormattingOptions) WithEmptyArgumentListSpacing(style SpacingStyle) FormattingOptions {
	options.EmptyArgumentListSpacing = style
	return options
}

func (options FormattingOptions) WithParameterListSpacing(style SpacingStyle) FormattingOptions {
	options.ParameterListSpacing = style
	return options
}

func (options FormattingOptions) WithEmptyParameterListSpacing(style SpacingStyle) FormattingOptions {
	options.EmptyParameterListSpacing = style
	return options
}

func (options FormattingOptions) WithJsonArraySpacing(style SpacingStyle) FormattingOptions {
	options.JsonArraySpacing = style
	return options
}

func (options FormattingOptions) WithEmptyJsonArraySpacing(style SpacingStyle) FormattingOptions {
	options.EmptyJsonArraySpacing = style
	return options
}

func (options FormattingOptions) WithJsonObjectSpacing(style SpacingStyle) FormattingOptions {
	options.JsonObjectSpacing = style
	return options
}

func (options FormattingOptions) WithEmptyJsonObjectSpacing(style SpacingStyle) FormattingOptions {
	options.EmptyJsonObjectSpacing = style
	return options
}

func (options FormattingOptions) WithDataTableValueSpacing(style SpacingStyle) FormattingOptions {
	options.DataTableValueSpacing = style
	return options
}

func (options FormattingOptions) WithEmptyDataTableValueSpacing(style SpacingStyle) FormattingOptions {
	options.EmptyDataTableValueSpacing = style
	return options
}

func (options FormattingOptions) WithFunctionBodySpacing(style SpacingStyle) FormattingOptions {
	options.FunctionBodySpacing = style
	return options
}

func (options FormattingOptions) WithEmptyFunctionBodySpacing(style SpacingStyle) FormattingOptions {
	options.EmptyFunctionBodySpacing = style
	return options
}

func (options FormattingOptions) WithBeforeFunctionBodySpacing(style SpacingStyle) FormattingOptions {
	options.BeforeFunctionBodySpacing = style
	return options
}

func (options FormattingOptions) WithBeforeParameterListSpacing(style SpacingStyle) FormattingOptions {
	options.BeforeParameterListSpacing = style
	return options
}

func (options FormattingOptions) WithBeforeArgumentListSpacing(style SpacingStyle) FormattingOptions {
	options.BeforeArgumentListSpacing = style
	return options
}

func (options FormattingOptions) WithBeforeDataTableValueSpacing(style SpacingStyle) FormattingOptions {
	options.BeforeDataTableValueSpacing = style
	return options
}

type OutliningOptions struct {
	MaxCollapsedTextLength int
	Queries                bool
	Statements             bool
}

var OutliningOptionsDefault = OutliningOptions{
	MaxCollapsedTextLength: 120,
	Queries:                true,
	Statements:             false,
}

func (options OutliningOptions) WithMaxCollapesedTextlength(length int) OutliningOptions {
	options.MaxCollapsedTextLength = length
	return options
}

func (options OutliningOptions) WithMaxCollapsedTextLength(length int) OutliningOptions {
	return options.WithMaxCollapesedTextlength(length)
}

func (options OutliningOptions) WithQueries(enable bool) OutliningOptions {
	options.Queries = enable
	return options
}

func (options OutliningOptions) WithStatements(enable bool) OutliningOptions {
	options.Statements = enable
	return options
}

type QuickInfo struct {
	Items []QuickInfoItem
	text  *string
}

func NewQuickInfoText(text string) QuickInfo {
	return QuickInfo{
		Items: []QuickInfoItem{NewQuickInfoItem(QuickInfoKindText, NewClassifiedText(text))},
		text:  &text,
	}
}

func NewQuickInfo(items ...QuickInfoItem) QuickInfo {
	return NewQuickInfoItems(items)
}

func NewQuickInfoItems(items []QuickInfoItem) QuickInfo {
	cloned := make([]QuickInfoItem, len(items))
	for i, item := range items {
		cloned[i] = cloneQuickInfoItem(item)
	}
	return QuickInfo{Items: ListToReadOnly(cloned)}
}

func (info QuickInfo) ItemsList() []QuickInfoItem {
	return NewQuickInfoItems(info.Items).Items
}

func (info QuickInfo) Text() string {
	if info.text != nil {
		return *info.text
	}
	var builder strings.Builder
	for i, item := range info.Items {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		if kindText := quickInfoKindText(item.Kind); kindText != "" {
			builder.WriteString("(")
			builder.WriteString(kindText)
			builder.WriteString(") ")
		}
		builder.WriteString(item.Text())
	}
	return builder.String()
}

var QuickInfoEmpty = NewQuickInfoItems(nil)

type QuickInfoItem struct {
	Kind  QuickInfoKind
	Parts []ClassifiedText
}

func NewQuickInfoItem(kind QuickInfoKind, parts ...ClassifiedText) QuickInfoItem {
	return QuickInfoItem{Kind: kind, Parts: ListToReadOnly(parts)}
}

func cloneQuickInfoItem(item QuickInfoItem) QuickInfoItem {
	return NewQuickInfoItem(item.Kind, item.Parts...)
}

func (item QuickInfoItem) PartsList() []ClassifiedText {
	return ListToReadOnly(item.Parts)
}

func NewQuickInfoItemText(kind QuickInfoKind, text string) QuickInfoItem {
	return NewQuickInfoItem(kind, NewClassifiedText(text))
}

func (item QuickInfoItem) Text() string {
	var builder strings.Builder
	for _, part := range item.Parts {
		builder.WriteString(part.Text)
	}
	return builder.String()
}

type QuickInfoKind int

const (
	QuickInfoKindText QuickInfoKind = iota
	QuickInfoKindError
	QuickInfoKindWarning
	QuickInfoKindSuggestion
	QuickInfoKindColumn
	QuickInfoKindTable
	QuickInfoKindDatabase
	QuickInfoKindCluster
	QuickInfoKindLiteral
	QuickInfoKindType
	QuickInfoKindPattern
	QuickInfoKindParameter
	QuickInfoKindScalar
	QuickInfoKindVariable
	QuickInfoKindLocalFunction
	QuickInfoKindBuiltInFunction
	QuickInfoKindDatabaseFunction
	QuickInfoKindOperator
	QuickInfoKindCommand
	QuickInfoKindOption
	QuickInfoKindGraph
)

func quickInfoKindText(kind QuickInfoKind) string {
	switch kind {
	case QuickInfoKindTable:
		return "table"
	case QuickInfoKindDatabase:
		return "database"
	case QuickInfoKindCluster:
		return "cluster"
	case QuickInfoKindLiteral:
		return "literal"
	case QuickInfoKindType:
		return "type"
	case QuickInfoKindPattern:
		return "pattern"
	case QuickInfoKindParameter:
		return "parameter"
	case QuickInfoKindScalar:
		return "scalar"
	case QuickInfoKindVariable:
		return "variable"
	case QuickInfoKindLocalFunction, QuickInfoKindBuiltInFunction, QuickInfoKindDatabaseFunction:
		return "function"
	case QuickInfoKindOperator:
		return "operator"
	case QuickInfoKindGraph:
		return "graph"
	default:
		return ""
	}
}

type ClassifiedText struct {
	Kind ClassificationKind
	Text string
}

func NewClassifiedText(text string) ClassifiedText {
	return NewClassifiedTextKind(ClassificationKindPlainText, text)
}

func NewClassifiedTextKind(kind ClassificationKind, text string) ClassifiedText {
	return ClassifiedText{Kind: kind, Text: text}
}

type RelatedInfo struct {
	Elements     []RelatedElement
	CurrentIndex int
}

func NewRelatedInfo(elements []RelatedElement, currentIndex int) RelatedInfo {
	return RelatedInfo{Elements: ListToReadOnly(elements), CurrentIndex: currentIndex}
}

func (info RelatedInfo) ElementsList() []RelatedElement {
	return ListToReadOnly(info.Elements)
}

func (info RelatedInfo) GetNextIndex(index int) int {
	if index < len(info.Elements)-1 {
		return index + 1
	}
	return 0
}

func (info RelatedInfo) GetPreviousIndex(index int) int {
	if index > 0 {
		return index - 1
	}
	return len(info.Elements) - 1
}

var RelatedInfoEmpty = NewRelatedInfo(nil, 0)

type RelatedElementKind int

const (
	RelatedElementKindSyntax RelatedElementKind = iota
	RelatedElementKindReference
	RelatedElementKindDeclaration
	RelatedElementKindOther
)

type RelatedElement struct {
	Range       TextRange
	Kind        RelatedElementKind
	CursorLeft  int
	CursorRight int
}

func NewRelatedElementRangeWithCursors(rng TextRange, kind RelatedElementKind, cursorLeft int, cursorRight int) RelatedElement {
	return RelatedElement{Range: rng, Kind: kind, CursorLeft: cursorLeft, CursorRight: cursorRight}
}

func NewRelatedElementRange(rng TextRange, kind RelatedElementKind) RelatedElement {
	return NewRelatedElementRangeWithCursors(rng, kind, rng.Start, rng.End())
}

func NewRelatedElementWithCursors(start int, length int, kind RelatedElementKind, cursorLeft int, cursorRight int) RelatedElement {
	return NewRelatedElementRangeWithCursors(NewTextRange(start, length), kind, cursorLeft, cursorRight)
}

func NewRelatedElement(start int, length int, kind RelatedElementKind) RelatedElement {
	return NewRelatedElementWithCursors(start, length, kind, start, start+length)
}

func (element RelatedElement) Start() int {
	return element.Range.Start
}

func (element RelatedElement) Length() int {
	return element.Range.Length
}

func (element RelatedElement) End() int {
	return element.Range.End()
}

type FindRelatedOptions uint32

const (
	FindRelatedOptionsNone                FindRelatedOptions = 0
	FindRelatedOptionsSeeThroughVariables FindRelatedOptions = 1
)
