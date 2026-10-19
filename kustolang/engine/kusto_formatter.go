package kustolang

import (
	"strings"
	"unicode"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoFormatter.cs at the commit
// recorded in third_party/kusto/UPSTREAM.json. See third_party/kusto/LICENSE.
//
// This implements the upstream two-pass rule/emission design, token and binary
// operator spacing, delimiter-specific spacing, statement and pipe placement,
// function call/declaration list layout, bracket alignment, comment handling,
// missing-token insertion, malformed-tree recovery, datatable schema/value
// layout, wide binary-chain and query-operator list wrapping, clause/sub-element
// alignment, semicolon placement, smart-pipe normalization, and caret mapping.

// FormatKustoCode formats an already parsed Kusto code value. Positions are
// byte offsets, matching the native Go syntax tree.
func FormatKustoCode(code *KustoCode, options FormattingOptions, caretPosition int) FormattedText {
	if code == nil || code.Syntax() == nil {
		return NewFormattedText("", caretPosition)
	}
	return GetKustoFormattedText(code.Syntax(), options, caretPosition)
}

// GetKustoFormattedText formats a syntax root and maps caretPosition into the
// returned text. It is the native counterpart of KustoFormatter.GetFormattedText.
func GetKustoFormattedText(node *SyntaxNode, options FormattingOptions, caretPosition int) FormattedText {
	formatter := newKustoFormatter(caretPosition, options)
	formatter.format(node)
	return NewFormattedText(formatter.builder.String(), formatter.newCaretPosition)
}

type kustoSpacingKind uint8

const (
	kustoSpacingAsIs kustoSpacingKind = iota
	kustoSpacingNoSpace
	kustoSpacingSingleSpace
	kustoSpacingMinimalSpace
	kustoSpacingNoSpaceIfSameLine
	kustoSpacingSingleSpaceIfSameLine
	kustoSpacingMinimalSpaceIfSameLine
	kustoSpacingAlignOnly
	kustoSpacingNewLine
)

type kustoSpacingRule struct {
	kind      kustoSpacingKind
	condition func() bool
	otherwise *kustoSpacingRule
}

func (r *kustoSpacingRule) getKind() kustoSpacingKind {
	for current := r; current != nil; current = current.otherwise {
		if current.condition == nil || current.condition() {
			return current.kind
		}
	}
	return kustoSpacingAsIs
}

type kustoAlignmentRule struct {
	relativeTo SyntaxElement
	delta      int
}

type kustoFormatter struct {
	options          FormattingOptions
	builder          strings.Builder
	caretPosition    int
	newCaretPosition int
	currentLineStart int
	spacingRules     map[*SyntaxToken]*kustoSpacingRule
	resolvingRules   map[*SyntaxToken]bool
	alignmentRules   map[SyntaxElement]kustoAlignmentRule
	lineOffsets      map[SyntaxElement]int
}

func newKustoFormatter(caretPosition int, options FormattingOptions) *kustoFormatter {
	if options.IndentationSize < 0 {
		options.IndentationSize = 0
	}
	return &kustoFormatter{
		options:          options,
		caretPosition:    caretPosition,
		newCaretPosition: -1,
		spacingRules:     make(map[*SyntaxToken]*kustoSpacingRule),
		resolvingRules:   make(map[*SyntaxToken]bool),
		alignmentRules:   make(map[SyntaxElement]kustoAlignmentRule),
		lineOffsets:      make(map[SyntaxElement]int),
	}
}

func (f *kustoFormatter) format(node *SyntaxNode) {
	if node == nil {
		f.newCaretPosition = f.caretPosition
		return
	}
	f.identifyRules(node)
	f.writeElement(node, 0)
	if f.newCaretPosition >= 0 {
		return
	}
	if f.caretPosition <= 0 {
		f.newCaretPosition = f.caretPosition
	} else if f.caretPosition <= node.End() {
		f.newCaretPosition = f.builder.Len()
	} else {
		// Go's SyntaxNode.Width sums child token widths and excludes interior
		// trivia; upstream Width is the contiguous formatted source span.
		f.newCaretPosition = f.caretPosition + f.builder.Len() - (node.End() - node.TextStart())
	}
}

func (f *kustoFormatter) identifyRules(element SyntaxElement) {
	if element == nil {
		return
	}
	for i := 0; i < element.ChildCount(); i++ {
		f.identifyRules(element.GetChild(i))
	}
	if token, ok := element.(*SyntaxToken); ok {
		f.addTokenRules(token)
		return
	}
	f.addElementRules(element)
}

func (f *kustoFormatter) addRule(token *SyntaxToken, rule *kustoSpacingRule) {
	if token == nil || rule == nil {
		return
	}
	if existing := f.spacingRules[token]; existing != nil {
		last := rule
		for last.otherwise != nil {
			last = last.otherwise
		}
		last.otherwise = existing
	}
	f.spacingRules[token] = rule
}

func unconditionalSpacing(kind kustoSpacingKind) *kustoSpacingRule {
	return &kustoSpacingRule{kind: kind}
}

func (f *kustoFormatter) addAlignment(element, relativeTo SyntaxElement, delta int) {
	if element != nil {
		f.alignmentRules[element] = kustoAlignmentRule{relativeTo: relativeTo, delta: delta}
	}
}

func (f *kustoFormatter) addTokenRules(token *SyntaxToken) {
	previous := token.GetPreviousToken()
	if previous == nil {
		return
	}
	if kustoIdentifierOrKeyword(token) && kustoIdentifierOrKeyword(previous) {
		f.addRule(token, unconditionalSpacing(kustoSpacingSingleSpaceIfSameLine))
	} else if kustoTokenIsNamedChild(token, "Operator") && kustoIsBinaryExpression(token.Parent()) {
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.InfixOperatorSpacing)))
	} else if kustoTokenIsNamedChild(previous, "Operator") && kustoIsBinaryExpression(previous.Parent()) {
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.InfixOperatorSpacing)))
	} else if kustoTokenIsNamedChild(previous, "Operator") && kustoIsPrefixUnaryExpression(previous.Parent()) {
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.options.PrefixOperatorSpacing)))
	} else if token.Kind() != SyntaxKindEndOfTextToken {
		f.addBeforeTokenRule(token, previous)
	}

	if token.Kind() != SyntaxKindEndOfTextToken {
		f.addAfterPreviousTokenRule(token, previous)
	}
}

func (f *kustoFormatter) addBeforeTokenRule(token, previous *SyntaxToken) {
	switch token.Kind() {
	case SyntaxKindCommaToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.CommaSpacing)))
	case SyntaxKindColonToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.ColonSpacing)))
	case SyntaxKindBarToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.PipeOperatorSpacing)))
	case SyntaxKindEqualToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.AssignmentSpacing)))
	case SyntaxKindDotToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingNoSpaceIfSameLine))
	case SyntaxKindDotDotToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.RangeOperatorSpacing)))
	case SyntaxKindSemicolonToken:
		f.addRule(token, unconditionalSpacing(kustoBeforeSpacing(f.options.SemicolonSpacing)))
	case SyntaxKindOpenBraceToken:
		if token.Parent() != nil && token.Parent().Kind() == SyntaxKindFunctionBody {
			f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.options.BeforeFunctionBodySpacing)))
		}
	case SyntaxKindCloseBraceToken:
		style := SpacingStyleOne
		switch token.Parent().Kind() {
		case SyntaxKindFunctionBody:
			style = f.options.FunctionBodySpacing
			if previous.Kind() == SyntaxKindOpenBraceToken {
				style = f.options.EmptyFunctionBodySpacing
			}
		case SyntaxKindJsonObjectExpression:
			style = f.options.JsonObjectSpacing
			if previous.Kind() == SyntaxKindOpenBraceToken {
				style = f.options.EmptyJsonObjectSpacing
			}
		case SyntaxKindBracedName:
			style = SpacingStyleMinimal
		}
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(style)))
	case SyntaxKindOpenBracketToken:
		if token.Parent() != nil && token.Parent().Kind() == SyntaxKindDataTableExpression {
			f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.options.BeforeDataTableValueSpacing)))
		}
	case SyntaxKindCloseBracketToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.bracketInteriorStyle(token, previous.Kind() == SyntaxKindOpenBracketToken))))
	case SyntaxKindOpenParenToken:
		switch token.Parent().Kind() {
		case SyntaxKindExpressionList:
			f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.options.BeforeArgumentListSpacing)))
		case SyntaxKindFunctionParameters:
			f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.options.BeforeParameterListSpacing)))
		case SyntaxKindDynamicExpression:
			f.addRule(token, unconditionalSpacing(kustoSpacingMinimalSpace))
		}
	case SyntaxKindCloseParenToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.parenInteriorStyle(token, previous.Kind() == SyntaxKindOpenParenToken))))
	}
}

func (f *kustoFormatter) addAfterPreviousTokenRule(token, previous *SyntaxToken) {
	switch previous.Kind() {
	case SyntaxKindCommaToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.CommaSpacing)))
	case SyntaxKindColonToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.ColonSpacing)))
	case SyntaxKindBarToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.PipeOperatorSpacing)))
	case SyntaxKindEqualToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.AssignmentSpacing)))
	case SyntaxKindDotToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingNoSpaceIfSameLine))
	case SyntaxKindDotDotToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.RangeOperatorSpacing)))
	case SyntaxKindSemicolonToken:
		f.addRule(token, unconditionalSpacing(kustoAfterSpacing(f.options.SemicolonSpacing)))
	case SyntaxKindOpenBraceToken:
		style := SpacingStyleOne
		if previous.Parent() != nil {
			switch previous.Parent().Kind() {
			case SyntaxKindFunctionBody:
				style = f.options.FunctionBodySpacing
			case SyntaxKindJsonObjectExpression:
				style = f.options.JsonObjectSpacing
			case SyntaxKindBracedName:
				style = SpacingStyleMinimal
			}
		}
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(style)))
	case SyntaxKindOpenBracketToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.bracketInteriorStyle(previous, false))))
	case SyntaxKindOpenParenToken:
		f.addRule(token, unconditionalSpacing(kustoSpacingFromStyle(f.parenInteriorStyle(previous, false))))
	}
}

func (f *kustoFormatter) bracketInteriorStyle(bracket *SyntaxToken, empty bool) SpacingStyle {
	if bracket.Parent() == nil {
		return SpacingStyleMinimal
	}
	switch bracket.Parent().Kind() {
	case SyntaxKindJsonArrayExpression:
		if empty {
			return f.options.EmptyJsonArraySpacing
		}
		return f.options.JsonArraySpacing
	case SyntaxKindDataTableExpression:
		if empty {
			return f.options.EmptyDataTableValueSpacing
		}
		return f.options.DataTableValueSpacing
	default:
		return SpacingStyleMinimal
	}
}

func (f *kustoFormatter) parenInteriorStyle(paren *SyntaxToken, empty bool) SpacingStyle {
	if paren.Parent() == nil {
		return SpacingStyleMinimal
	}
	switch paren.Parent().Kind() {
	case SyntaxKindExpressionList:
		if empty {
			return f.options.EmptyArgumentListSpacing
		}
		return f.options.ArgumentListSpacing
	case SyntaxKindFunctionParameters:
		if empty {
			return f.options.EmptyParameterListSpacing
		}
		return f.options.ParameterListSpacing
	case SyntaxKindParenthesizedExpression:
		return f.options.ParenthesizedExpressionSpacing
	default:
		return SpacingStyleMinimal
	}
}

func (f *kustoFormatter) addElementRules(element SyntaxElement) {
	if list, ok := element.(*SyntaxList); ok {
		f.addQueryOperatorCommaListRules(list)
	}
	if separated, ok := element.(*SeparatedElement); ok {
		if _, isLet := WrapGeneratedSyntaxElement(separated.Element()).(*LetStatement); isLet {
			f.addSemicolonRules(separated)
		}
	}
	f.addSubElementRules(element)
	wrapped := WrapGeneratedSyntaxElement(element)
	if statement, ok := wrapped.(Statement); ok {
		f.addStatementRules(syntaxNodeForBinderLocation(statement))
	}
	switch typed := wrapped.(type) {
	case *PipeExpression:
		f.addPipeRules(typed)
	case *FunctionCallExpression:
		f.addFunctionCallRules(typed)
	case *FunctionDeclaration:
		f.addFunctionDeclarationRules(typed)
	case *DataTableExpression:
		f.addDataTableExpressionRules(typed)
	case *BinaryExpression:
		f.addBinaryOperatorChainRules(typed)
	case *BracketedExpression:
		if typed.Parent() != nil && typed.Parent().Kind() == SyntaxKindElementExpression {
			f.addRule(typed.GetFirstToken(true), unconditionalSpacing(kustoSpacingNoSpaceIfSameLine))
		}
	}
}

func (f *kustoFormatter) addSubElementRules(element SyntaxElement) {
	if element == nil || element.IsToken() || element.Parent() == nil {
		return
	}
	parent := WrapGeneratedSyntaxElement(element.Parent())
	_, statementParent := parent.(Statement)
	_, expressionStatement := parent.(*ExpressionStatement)
	_, queryOperatorParent := parent.(QueryOperator)
	_, commandParent := parent.(Command)
	if statementParent && !expressionStatement || queryOperatorParent || commandParent {
		if element.IndexInParent() > 0 {
			f.addAlignment(element, nil, f.options.IndentationSize)
		}
		if _, clause := WrapGeneratedSyntaxElement(element).(Clause); clause {
			first := element.GetFirstToken(true)
			f.addRule(first, &kustoSpacingRule{
				kind:      kustoSpacingNewLine,
				condition: func() bool { return f.spansOrWillSpan(element.Parent(), nil) },
			})
		}
	}
}

const (
	kustoMaxBinaryOperatorChainWidth = 80
	kustoMaxCommaListWidth           = 80
	kustoMaxSingleLineDataTableWidth = 80
	kustoMaxSingleLineSchemaWidth    = 60
	kustoMaxSingleLineRowWidth       = 80
	kustoMaxDataTableValueWidth      = 40
)

func (f *kustoFormatter) addBinaryOperatorChainRules(binary *BinaryExpression) {
	if binary == nil || binary.SyntaxNode == nil || !kustoIsChainableBinaryOperator(binary.Kind()) ||
		!kustoIsDirectQueryOperatorPart(binary.SyntaxNode) || kustoSpansMultipleLines(binary.SyntaxNode, nil) ||
		kustoContiguousWidth(binary.SyntaxNode) <= kustoMaxBinaryOperatorChainWidth || kustoBinaryOperatorChainDepth(binary) <= 1 {
		return
	}
	for current := binary; current != nil && current.Kind() == binary.Kind(); {
		f.addRule(current.Operator(), unconditionalSpacing(kustoSpacingNewLine))
		current, _ = current.Left().(*BinaryExpression)
	}
}

func (f *kustoFormatter) addQueryOperatorCommaListRules(list *SyntaxList) {
	if list == nil || list.Count() <= 1 || !kustoIsDirectQueryOperatorPart(list) {
		return
	}
	first, ok := list.At(0).(*SeparatedElement)
	if !ok || first.Separator() == nil || first.Separator().Kind() != SyntaxKindCommaToken {
		return
	}
	if kustoContiguousWidth(list) > kustoMaxCommaListWidth && !kustoSpansMultipleLines(list, nil) {
		f.addListRules(list, nil, list, PlacementStyleNewLine)
	} else {
		f.addListRules(list, nil, list, f.options.ExpressionStyle)
	}
}

func (f *kustoFormatter) addDataTableExpressionRules(dataTable *DataTableExpression) {
	if dataTable == nil || dataTable.SyntaxNode == nil ||
		kustoContiguousWidth(dataTable.SyntaxNode) <= kustoMaxSingleLineDataTableWidth && !kustoSpansMultipleLines(dataTable.SyntaxNode, nil) {
		return
	}

	indentRelativeTo := SyntaxElement(dataTable.SyntaxNode)
	for ancestor := dataTable.Parent(); ancestor != nil; ancestor = ancestor.Parent() {
		if statement, ok := WrapGeneratedSyntaxElement(ancestor).(Statement); ok {
			indentRelativeTo = syntaxNodeForBinderLocation(statement)
			break
		}
	}

	schema := dataTable.Schema()
	schemaColumns := 0
	if schema != nil && schema.Columns() != nil {
		schemaColumns = schema.Columns().Count()
		if kustoContiguousWidth(schema.SyntaxNode) > kustoMaxSingleLineSchemaWidth || kustoSpansMultipleLines(schema.SyntaxNode, nil) {
			f.addBrackettingRules(schema.OpenParen(), schema.CloseParen(), indentRelativeTo, f.options.SchemaStyle, f.options.GeneralSpacing)
			f.addListRules(schema.Columns().SyntaxList, indentRelativeTo, schema.SyntaxNode, PlacementStyleNewLine)
		}
	}

	values := dataTable.Values()
	valueColumns := schemaColumns
	if valueColumns <= 1 {
		valueColumns = kustoBalancedDataTableColumnCount(values)
	}
	f.addBrackettingRules(dataTable.OpenBracket(), dataTable.CloseBracket(), indentRelativeTo, f.options.DataTableValueStyle, f.options.BeforeDataTableValueSpacing)
	if values == nil || valueColumns <= 0 {
		return
	}
	for i := 0; i < values.Count(); i++ {
		separated := values.At(i)
		if separated == nil || separated.Element() == nil {
			continue
		}
		first := separated.Element().GetFirstToken(true)
		if i%valueColumns == 0 {
			f.addAlignment(first, indentRelativeTo, f.options.IndentationSize)
			f.addRule(first, unconditionalSpacing(kustoSpacingNewLine))
		} else {
			f.addRule(first, unconditionalSpacing(kustoSpacingSingleSpace))
		}
	}
}

func kustoBalancedDataTableColumnCount(values *SeparatedSyntaxListOf[Expression]) int {
	if values == nil || values.Count() == 0 {
		return 1
	}
	if kustoMaximumDataTableValueWidth(values) >= kustoMaxDataTableValueWidth {
		return 1
	}
	for columns := 10; ; {
		if kustoMaximumDataTableRowWidth(values, columns) <= kustoMaxSingleLineRowWidth {
			return columns
		}
		if columns <= 4 {
			return 1
		}
		columns -= 2
	}
}

func kustoMaximumDataTableValueWidth(values *SeparatedSyntaxListOf[Expression]) int {
	maximum := 0
	for i := 0; i < values.Count(); i++ {
		if value := values.At(i); value != nil && value.Element() != nil {
			if width := kustoContiguousWidth(value.Element()); width > maximum {
				maximum = width
			}
		}
	}
	return maximum
}

func kustoMaximumDataTableRowWidth(values *SeparatedSyntaxListOf[Expression], columns int) int {
	maximum := 0
	for rowStart := 0; rowStart < values.Count(); rowStart += columns {
		rowEnd := rowStart + columns - 1
		if rowEnd >= values.Count() {
			rowEnd = values.Count() - 1
		}
		first, last := values.At(rowStart), values.At(rowEnd)
		if first == nil || last == nil || first.Element() == nil || last.Element() == nil {
			continue
		}
		if width := last.Element().End() - first.Element().TriviaStart(); width > maximum {
			maximum = width
		}
	}
	return maximum
}

func (f *kustoFormatter) addStatementRules(statement *SyntaxNode) {
	first := statement.GetFirstToken(true)
	if first == nil || first.GetPreviousToken(true) == nil {
		return
	}
	switch f.options.StatementStyle {
	case PlacementStyleNewLine:
		f.addRule(first, unconditionalSpacing(kustoSpacingNewLine))
	case PlacementStyleSmart:
		separated, ok := statement.Parent().(*SeparatedElement)
		if !ok {
			return
		}
		list, ok := separated.Parent().(*SyntaxList)
		if !ok {
			return
		}
		f.addRule(first, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: func() bool { return f.spansOrWillSpan(list, first) }})
	}
}

func (f *kustoFormatter) addSemicolonRules(separated *SeparatedElement) {
	if separated == nil || separated.Separator() == nil || separated.Separator().Kind() != SyntaxKindSemicolonToken {
		return
	}

	semicolon := separated.Separator()
	switch f.options.SemicolonStyle {
	case PlacementStyleSmart:
		f.addRule(semicolon, &kustoSpacingRule{
			kind:      kustoSpacingNewLine,
			condition: func() bool { return f.spansOrWillSpan(separated.Element(), nil) },
		})
		f.addAlignment(semicolon, separated, f.options.IndentationSize)
	case PlacementStyleNewLine:
		f.addRule(semicolon, unconditionalSpacing(kustoSpacingNewLine))
		f.addAlignment(semicolon, separated, f.options.IndentationSize)
	}
}

func (f *kustoFormatter) addPipeRules(pipe *PipeExpression) {
	if pipe == nil || pipe.SyntaxNode == nil || pipe.Parent() != nil && pipe.Parent().Kind() == SyntaxKindPipeExpression {
		return
	}
	style := f.options.PipeOperatorStyle
	if !kustoSpansMultipleLines(pipe.SyntaxNode, nil) && pipe.Width() > 80 {
		style = PlacementStyleNewLine
	}
	whole := pipe.SyntaxNode
	for current := pipe; current != nil; {
		bar := current.Bar()
		if bar == nil {
			break
		}
		switch style {
		case PlacementStyleNewLine:
			f.addRule(bar, unconditionalSpacing(kustoSpacingNewLine))
		case PlacementStyleSmart:
			spansQuery := func() bool { return f.spansOrWillSpan(whole, bar) }
			f.addRule(bar, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: spansQuery})

			if next := bar.GetNextToken(); next != nil && kustoWhitespaceOnly(next.Trivia()) && strings.ContainsAny(next.Trivia(), "\r\n") {
				f.addRule(next, &kustoSpacingRule{
					kind:      kustoForcedAfterSpacing(f.options.PipeOperatorSpacing),
					condition: spansQuery,
				})
			}

			if current == pipe {
				first := current.GetFirstToken()
				if first != nil {
					previous := first.GetPreviousToken()
					if previous != nil && (previous.Kind() == SyntaxKindOpenParenToken || previous.Kind() == SyntaxKindOpenBraceToken) {
						f.addRule(first, &kustoSpacingRule{
							kind:      kustoSpacingNewLine,
							condition: func() bool { return f.spansOrWillSpan(whole, first) },
						})
					}
				}
			}
		}
		next, _ := current.Expression().(*PipeExpression)
		current = next
	}
}

func (f *kustoFormatter) addFunctionCallRules(call *FunctionCallExpression) {
	if call == nil || call.ArgumentList() == nil {
		return
	}
	args := call.ArgumentList()
	f.addBrackettingRules(args.OpenParen(), args.CloseParen(), call.Name(), f.options.FunctionArgumentStyle, f.options.BeforeArgumentListSpacing)
	f.addListRules(args.Expressions().SyntaxList, call.Name(), args.SyntaxNode, PlacementStyleSmart)
}

func (f *kustoFormatter) addFunctionDeclarationRules(declaration *FunctionDeclaration) {
	if declaration == nil || declaration.Parameters() == nil || declaration.Body() == nil {
		return
	}
	parameters, body := declaration.Parameters(), declaration.Body()
	relative := declaration.Parent()
	f.addBrackettingRules(parameters.OpenParen(), parameters.CloseParen(), relative, f.options.FunctionParameterStyle, f.options.BeforeParameterListSpacing)
	f.addBrackettingRules(body.OpenBrace(), body.CloseBrace(), relative, f.options.FunctionBodyStyle, f.options.BeforeFunctionBodySpacing)
	f.addListRules(parameters.Parameters().SyntaxList, relative, parameters.SyntaxNode, PlacementStyleSmart)
	f.addListRules(body.Statements().SyntaxList, relative, body.SyntaxNode, PlacementStyleSmart)
	if expression := body.Expression(); expression != nil {
		f.addListElementRule(expression, relative, body.SyntaxNode, PlacementStyleSmart)
	}
}

func (f *kustoFormatter) addBrackettingRules(open, close *SyntaxToken, alignedTo SyntaxElement, style BrackettingStyle, beforeOpen SpacingStyle) {
	if open == nil || close == nil {
		return
	}
	if style == BrackettingStyleDefault {
		style = f.options.BrackettingStyle
	}
	conditionOpen := func() bool { return f.spansOrWillSpanBetween(open, close, open) }
	conditionClose := func() bool { return f.spansOrWillSpanBetween(open, close, close) }
	switch style {
	case BrackettingStyleVertical:
		f.addAlignment(open, alignedTo, 0)
		f.addAlignment(close, alignedTo, 0)
		f.addRule(open, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: conditionOpen})
		f.addRule(close, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: conditionClose})
	case BrackettingStyleDiagonal:
		f.addAlignment(close, alignedTo, 0)
		f.addRule(close, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: conditionClose})
		if kustoWhitespaceOnly(open.Trivia()) && strings.ContainsAny(open.Trivia(), "\r\n") {
			f.addRule(open, &kustoSpacingRule{kind: kustoForcedSpacing(beforeOpen), condition: conditionClose})
		}
	}
}

func (f *kustoFormatter) addListRules(list *SyntaxList, relative SyntaxElement, trigger SyntaxElement, style PlacementStyle) {
	if list == nil {
		return
	}
	for i := 0; i < list.Count(); i++ {
		f.addListElementRule(list.At(i), relative, trigger, style)
	}
}

func (f *kustoFormatter) addListElementRule(element, relative, trigger SyntaxElement, style PlacementStyle) {
	if element == nil {
		return
	}
	first := element.GetFirstToken(true)
	if relative != nil {
		f.addAlignment(first, relative, f.options.IndentationSize)
	}
	switch style {
	case PlacementStyleNewLine:
		f.addRule(first, unconditionalSpacing(kustoSpacingNewLine))
	case PlacementStyleSmart:
		f.addRule(first, &kustoSpacingRule{kind: kustoSpacingNewLine, condition: func() bool { return f.spansOrWillSpan(trigger, first) }})
	}
}

func (f *kustoFormatter) writeElement(element SyntaxElement, indentation int) {
	if element == nil {
		return
	}
	if rule, ok := f.alignmentRules[element]; ok {
		indentation = f.resolveIndentation(rule, indentation)
	}
	if token, ok := element.(*SyntaxToken); ok {
		spacing := kustoSpacingFromStyle(f.options.GeneralSpacing)
		if f.spacingRules[token] != nil {
			spacing = f.spacingKind(token)
		}
		offset := f.writeToken(token, indentation, spacing)
		f.lineOffsets[token] = offset
		return
	}
	for i := 0; i < element.ChildCount(); i++ {
		child := element.GetChild(i)
		f.writeElement(child, indentation)
	}
	if first := element.GetFirstToken(true); first != nil {
		if offset, ok := f.lineOffsets[first]; ok {
			f.lineOffsets[element] = offset
		}
	}
}

func (f *kustoFormatter) resolveIndentation(rule kustoAlignmentRule, fallback int) int {
	base := fallback
	if rule.relativeTo != nil {
		if offset, ok := f.lineOffsets[rule.relativeTo]; ok {
			base = offset
		} else if first := rule.relativeTo.GetFirstToken(true); first != nil {
			if offset, ok := f.lineOffsets[first]; ok {
				base = offset
			}
		}
	}
	if base+rule.delta < 0 {
		return 0
	}
	return base + rule.delta
}

func (f *kustoFormatter) writeToken(token *SyntaxToken, indentation int, spacing kustoSpacingKind) int {
	missing := token.RawText() == "" && token.Kind() != SyntaxKindEndOfTextToken
	if token.RawText() != "" || token.Trivia() != "" || missing && f.options.InsertMissingTokens {
		f.writeTrivia(token, indentation, spacing, token.Kind() != SyntaxKindEndOfTextToken)
	}
	if f.newCaretPosition < 0 {
		if f.caretPosition >= token.TextStart() && f.caretPosition <= token.End() {
			f.newCaretPosition = f.builder.Len() + f.caretPosition - token.TextStart()
		} else if f.caretPosition < token.TextStart() {
			f.newCaretPosition = f.builder.Len()
		}
	}
	start := f.builder.Len()
	if missing && f.options.InsertMissingTokens {
		f.builder.WriteString(SyntaxText(token.Kind()))
	} else {
		f.builder.WriteString(token.RawText())
	}
	return start - f.currentLineStart
}

func (f *kustoFormatter) writeTrivia(token *SyntaxToken, indentation int, spacing kustoSpacingKind, following bool) {
	previous := token.GetPreviousToken()
	trivia := token.Trivia()
	// Forced spacing is only allowed to replace whitespace. Comments are syntax
	// trivia and must survive layout rules such as datatable row balancing.
	if kustoWhitespaceOnly(trivia) {
		switch spacing {
		case kustoSpacingNoSpace:
			return
		case kustoSpacingSingleSpace:
			if previous != nil {
				f.builder.WriteByte(' ')
			}
			return
		case kustoSpacingMinimalSpace:
			if previous != nil && kustoRequiresSpacing(previous, token) {
				f.builder.WriteByte(' ')
			}
			return
		}
	}
	sameLine := true
	if !kustoWhitespaceOnly(trivia) {
		f.writeCommentTrivia(trivia, indentation, following)
		sameLine = !strings.ContainsAny(trivia, "\r\n")
	} else if strings.ContainsAny(trivia, "\r\n") || spacing == kustoSpacingNewLine {
		f.builder.WriteByte('\n')
		f.currentLineStart = f.builder.Len()
		f.writeIndentation(indentation)
		return
	} else if spacing == kustoSpacingAsIs || spacing == kustoSpacingAlignOnly {
		f.builder.WriteString(trivia)
		return
	}
	if previous != nil && sameLine && token.Kind() != SyntaxKindEndOfTextToken {
		switch spacing {
		case kustoSpacingSingleSpaceIfSameLine:
			f.builder.WriteByte(' ')
		case kustoSpacingMinimalSpaceIfSameLine:
			if kustoRequiresSpacing(previous, token) {
				f.builder.WriteByte(' ')
			}
		}
	}
}

func (f *kustoFormatter) writeCommentTrivia(trivia string, indentation int, following bool) {
	for start := 0; start < len(trivia); {
		lineEnd := start
		for lineEnd < len(trivia) && trivia[lineEnd] != '\r' && trivia[lineEnd] != '\n' {
			lineEnd++
		}
		contentStart := start
		for contentStart < lineEnd && (trivia[contentStart] == ' ' || trivia[contentStart] == '\t') {
			contentStart++
		}
		if start == 0 {
			f.builder.WriteString(trivia[start:contentStart])
		} else {
			f.writeIndentation(indentation)
		}
		f.builder.WriteString(trivia[contentStart:lineEnd])
		if lineEnd < len(trivia) {
			if trivia[lineEnd] == '\r' && lineEnd+1 < len(trivia) && trivia[lineEnd+1] == '\n' {
				lineEnd++
			}
			f.builder.WriteByte('\n')
			f.currentLineStart = f.builder.Len()
			lineEnd++
		}
		start = lineEnd
	}
	if following && strings.ContainsAny(trivia, "\r\n") && (strings.HasSuffix(trivia, "\n") || strings.HasSuffix(trivia, "\r")) {
		f.writeIndentation(indentation)
	}
}

func (f *kustoFormatter) writeIndentation(indentation int) {
	for i := 0; i < indentation; i++ {
		f.builder.WriteByte(' ')
	}
}

func (f *kustoFormatter) spansOrWillSpan(element SyntaxElement, excluded *SyntaxToken) bool {
	if kustoSpansMultipleLines(element, excluded) {
		return true
	}
	if element == nil {
		return false
	}
	for token := element.GetFirstToken(true); token != nil; token = token.GetNextToken(true) {
		if token != excluded {
			if f.spacingRules[token] != nil && f.spacingKind(token) == kustoSpacingNewLine {
				return true
			}
		}
		if token == element.GetLastToken(true) {
			break
		}
	}
	return false
}

func (f *kustoFormatter) spansOrWillSpanBetween(first, last, excluded *SyntaxToken) bool {
	for token := first; token != nil; token = token.GetNextToken(true) {
		if token != excluded && (strings.ContainsAny(token.Trivia(), "\r\n") || f.spacingRules[token] != nil && f.spacingKind(token) == kustoSpacingNewLine) {
			return true
		}
		if token == last {
			break
		}
	}
	return false
}

// spacingKind mirrors upstream's nullable WillSpanMultipleLines evaluation:
// a rule encountered recursively has an unknown outcome and therefore does
// not itself prove that a line break will be introduced.
func (f *kustoFormatter) spacingKind(token *SyntaxToken) kustoSpacingKind {
	rule := f.spacingRules[token]
	if rule == nil || f.resolvingRules[token] {
		return kustoSpacingAsIs
	}
	f.resolvingRules[token] = true
	kind := rule.getKind()
	delete(f.resolvingRules, token)
	return kind
}

func kustoSpansMultipleLines(element SyntaxElement, excluded *SyntaxToken) bool {
	if element == nil {
		return false
	}
	last := element.GetLastToken(true)
	for token := element.GetFirstToken(true); token != nil; token = token.GetNextToken(true) {
		if token != excluded && strings.ContainsAny(token.Trivia(), "\r\n") {
			return true
		}
		if token == last {
			break
		}
	}
	return false
}

func kustoSpacingFromStyle(style SpacingStyle) kustoSpacingKind {
	switch style {
	case SpacingStyleMinimal:
		return kustoSpacingMinimalSpaceIfSameLine
	case SpacingStyleOne:
		return kustoSpacingSingleSpaceIfSameLine
	default:
		return kustoSpacingAsIs
	}
}

func kustoForcedSpacing(style SpacingStyle) kustoSpacingKind {
	switch style {
	case SpacingStyleMinimal:
		return kustoSpacingMinimalSpace
	case SpacingStyleOne:
		return kustoSpacingSingleSpace
	default:
		return kustoSpacingAsIs
	}
}

func kustoBeforeSpacing(style DualSpacingStyle) kustoSpacingKind {
	switch style {
	case DualSpacingStyleNeither, DualSpacingStyleAfter:
		return kustoSpacingMinimalSpaceIfSameLine
	case DualSpacingStyleBefore, DualSpacingStyleBoth:
		return kustoSpacingSingleSpaceIfSameLine
	default:
		return kustoSpacingAsIs
	}
}

func kustoAfterSpacing(style DualSpacingStyle) kustoSpacingKind {
	switch style {
	case DualSpacingStyleNeither, DualSpacingStyleBefore:
		return kustoSpacingMinimalSpaceIfSameLine
	case DualSpacingStyleAfter, DualSpacingStyleBoth:
		return kustoSpacingSingleSpaceIfSameLine
	default:
		return kustoSpacingAsIs
	}
}

func kustoForcedAfterSpacing(style DualSpacingStyle) kustoSpacingKind {
	switch style {
	case DualSpacingStyleNeither, DualSpacingStyleBefore:
		return kustoSpacingMinimalSpace
	case DualSpacingStyleAfter, DualSpacingStyleBoth:
		return kustoSpacingSingleSpace
	default:
		return kustoSpacingAsIs
	}
}

func kustoIdentifierOrKeyword(token *SyntaxToken) bool {
	if token == nil {
		return false
	}
	category := SyntaxCategoryOf(token.Kind())
	return category == SyntaxCategoryIdentifier || category == SyntaxCategoryKeyword
}

func kustoTokenIsNamedChild(token *SyntaxToken, name string) bool {
	if token == nil || token.Parent() == nil {
		return false
	}
	return token.Parent().GetName(token.IndexInParent()) == name
}

func kustoIsBinaryExpression(element SyntaxElement) bool {
	_, ok := WrapGeneratedSyntaxElement(element).(*BinaryExpression)
	return ok
}

func kustoIsPrefixUnaryExpression(element SyntaxElement) bool {
	_, ok := WrapGeneratedSyntaxElement(element).(*PrefixUnaryExpression)
	return ok
}

func kustoIsDirectQueryOperatorPart(element SyntaxElement) bool {
	if element == nil || element.Parent() == nil {
		return false
	}
	parent := WrapGeneratedSyntaxElement(element.Parent())
	if _, ok := parent.(QueryOperator); ok {
		return true
	}
	_, ok := parent.(Clause)
	return ok
}

func kustoIsChainableBinaryOperator(kind SyntaxKind) bool {
	return kind == SyntaxKindAndExpression || kind == SyntaxKindOrExpression
}

func kustoBinaryOperatorChainDepth(binary *BinaryExpression) int {
	depth := 1
	for binary != nil {
		left, ok := binary.Left().(*BinaryExpression)
		if !ok || left.Operator() == nil || binary.Operator() == nil || left.Operator().Kind() != binary.Operator().Kind() {
			break
		}
		depth++
		binary = left
	}
	return depth
}

// Width on the native aggregate nodes sums token text and excludes interior
// trivia. Upstream formatter width decisions use the contiguous source span.
func kustoContiguousWidth(element SyntaxElement) int {
	if element == nil {
		return 0
	}
	width := element.End() - element.TriviaStart()
	if width < 0 {
		return element.Width()
	}
	return width
}

func kustoWhitespaceOnly(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func kustoRequiresSpacing(left, right *SyntaxToken) bool {
	if kustoIdentifierOrKeyword(left) && kustoIdentifierOrKeyword(right) {
		return true
	}
	rightText := SyntaxText(right.Kind())
	switch left.Kind() {
	case SyntaxKindEqualToken:
		return strings.HasPrefix(rightText, "=") || strings.HasPrefix(rightText, ">") || strings.HasPrefix(rightText, "~")
	case SyntaxKindBangToken:
		return strings.HasPrefix(rightText, "=") || strings.HasPrefix(rightText, "~")
	case SyntaxKindLessThanToken:
		return strings.HasPrefix(rightText, ">") || strings.HasPrefix(rightText, "=") || strings.HasPrefix(rightText, "|")
	case SyntaxKindGreaterThanToken:
		return strings.HasPrefix(rightText, "=")
	case SyntaxKindDotToken:
		return strings.HasPrefix(rightText, ".")
	default:
		return false
	}
}
