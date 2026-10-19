package kustolang

import "strings"

// Ported from the column-example completion path in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (c *kustoCompleter) getColumnExampleCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	if c == nil || c.code == nil || builder == nil {
		return completionModeCombined
	}
	if context, ok := c.binaryOperatorArgumentCompletionContext(position); ok {
		switch context.operator {
		case Operators.Equal, Operators.NotEqual, Operators.EqualTilde:
			if column := c.completionColumnForExpression(context.arguments[0], position); column != nil {
				c.addColumnExamples(builder, column)
			}
		}
	}
	if in := c.inExpressionCompletionContext(position); in != nil {
		if column := c.completionColumnForExpression(in.Left(), position); column != nil {
			c.addColumnExamples(builder, column)
		}
	}
	return completionModeCombined
}

func (c *kustoCompleter) inExpressionCompletionContext(position int) *InExpression {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return nil
	}
	var selected *InExpression
	selectedStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		in, ok := WrapGeneratedSyntaxElement(element).(*InExpression)
		if !ok || in == nil || in.Left() == nil || in.Right() == nil || in.Right().OpenParen() == nil {
			return
		}
		list := in.Right()
		open := list.OpenParen()
		if open.Width() == 0 || position < open.End() {
			return
		}
		if close := list.CloseParen(); close != nil && position > close.TextStart() {
			if close.Width() > 0 || close.TextStart() < 0 || close.TextStart() > len(c.code.Text) || strings.TrimSpace(c.code.Text[close.TextStart():position]) != "" {
				return
			}
		}
		if open.TextStart() >= selectedStart {
			selected = in
			selectedStart = open.TextStart()
		}
	})
	return selected
}

func (c *kustoCompleter) completionColumnForExpression(expression Expression, position int) *ColumnSymbol {
	if column, ok := expressionReferencedSymbol(expression).(*ColumnSymbol); ok {
		return column
	}
	switch typed := expression.(type) {
	case *SimpleNamedExpression:
		return c.completionColumnForExpression(typed.Expression(), position)
	case *ParenthesizedExpression:
		return c.completionColumnForExpression(typed.Expression(), position)
	case *BracketedExpression:
		return c.completionColumnForExpression(typed.Expression(), position)
	case *NameReference:
		name := NameReferenceSimpleName(typed)
		if c != nil && c.code != nil {
			if columns := c.code.GetColumnsInScope(position); columns != nil {
				if column := columns.GetColumn(name); column != nil {
					return column
				}
			}
			if rowScope := getQueryCompletionScope(c.code, position).rowScope; rowScope != nil {
				return rowScope.GetColumn(name)
			}
		}
	}
	return nil
}

func (c *kustoCompleter) addColumnExamples(builder *kustoCompletionBuilder, column *ColumnSymbol) {
	if builder == nil || column == nil {
		return
	}
	if len(column.Examples) > 0 {
		for _, example := range column.Examples {
			builder.add(createColumnExampleCompletion(example, column.Type))
		}
	} else if call, ok := column.Source.(*FunctionCallExpression); ok {
		budget := maxComputedColumnExamples
		getPossibleValueExamples(call, builder, &budget)
	}
}

// Bound traversal even when case has thousands of arguments.
const maxComputedColumnExamples = 100

// Skip oversized literals rather than truncating them into invalid KQL.
const maxComputedColumnExampleLength = 100

func getPossibleValueExamples(expression Expression, builder *kustoCompletionBuilder, budget *int) {
	if expression == nil || *budget <= 0 {
		return
	}
	*budget--
	if ExpressionIsLiteral(expression) {
		text := expression.Text(IncludeTriviaInterior)
		if literal, ok := expression.(*LiteralExpression); ok {
			text = literal.Token().RawText()
		}
		if length := byteOffsetToUTF16Position(text, len(text)); length > 0 && length <= maxComputedColumnExampleLength {
			builder.add(createComputedColumnExampleCompletion(expression, text))
		}
	} else if call, ok := expression.(*FunctionCallExpression); ok {
		if signature := call.ReferencedSignature(); signature != nil {
			arguments := functionCallArguments(call)
			parameters := signature.GetArgumentParameters(arguments)
			for i, argument := range arguments {
				if i >= len(parameters) || *budget <= 0 {
					break
				}
				if parameters[i] != nil && IsCommonArgumentTypeKind(parameters[i].TypeKind) {
					getPossibleValueExamples(argument, builder, budget)
				}
			}
		}
	}
}

func createComputedColumnExampleCompletion(expression Expression, text string) CompletionItem {
	if compound, ok := expression.(*CompoundStringLiteralExpression); ok {
		value := ExpressionLiteralValueInfo(compound).ValueText
		editText := GetStringLiteral(value)
		for _, token := range compound.Tokens().Elements() {
			if raw := token.RawText(); len(raw) > 0 && (raw[0] == 'h' || raw[0] == 'H') {
				editText = "h" + editText
				break
			}
		}
		return NewCompletionItem(CompletionKindExample, value, WithCompletionBeforeText(editText))
	}
	// Keep the literal's source syntax when its inferred type is promoted.
	item := createColumnExampleCompletion(text, ExpressionResultType(expression))
	item.ApplyTexts = []CompletionText{NewCompletionText(text, false, false)}
	return item
}

func createColumnExampleCompletion(example string, typ *TypeSymbol) CompletionItem {
	displayText := example
	editText := example
	matchText := example

	switch {
	case sameType(typ, ScalarTypes.String.Type()):
		if token, ok := completionExactLiteral(example, SyntaxKindStringLiteralToken); ok {
			value := SyntaxTokenFromLexical(token).ValueText()
			displayText = value
			matchText = value
		} else {
			editText = GetStringLiteral(example)
		}
	case sameType(typ, ScalarTypes.Bool.Type()):
		if strings.HasPrefix(example, "bool(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
			if _, ok := completionExactLiteral(displayText, SyntaxKindBooleanLiteralToken); ok {
				editText = displayText
			}
		} else if _, ok := completionExactLiteral(example, SyntaxKindBooleanLiteralToken); !ok {
			editText = "bool(" + example + ")"
		}
	case sameType(typ, ScalarTypes.Int.Type()):
		if strings.HasPrefix(example, "int(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
		} else {
			editText = "int(" + example + ")"
		}
	case sameType(typ, ScalarTypes.Long.Type()):
		if strings.HasPrefix(example, "long(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
			if _, ok := completionExactLiteral(displayText, SyntaxKindLongLiteralToken); ok {
				editText = displayText
			}
		} else if _, ok := completionExactLiteral(example, SyntaxKindLongLiteralToken); !ok {
			editText = "long(" + example + ")"
		}
	case sameType(typ, ScalarTypes.Real.Type()):
		if strings.HasPrefix(example, "real(") || strings.HasPrefix(example, "double(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
			if _, ok := completionExactLiteral(displayText, SyntaxKindRealLiteralToken); ok {
				editText = displayText
			}
		} else if _, ok := completionExactLiteral(example, SyntaxKindRealLiteralToken); !ok {
			editText = "real(" + example + ")"
		}
	case sameType(typ, ScalarTypes.Decimal.Type()):
		if strings.HasPrefix(example, "decimal(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
		} else {
			editText = "decimal(" + example + ")"
		}
	case sameType(typ, ScalarTypes.TimeSpan.Type()):
		if strings.HasPrefix(example, "timespan(") || strings.HasPrefix(example, "time(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
			if _, ok := completionExactLiteral(displayText, SyntaxKindTimespanLiteralToken); ok {
				editText = displayText
			}
		} else if _, ok := completionExactLiteral(example, SyntaxKindTimespanLiteralToken); !ok {
			editText = "timespan(" + example + ")"
		}
	case sameType(typ, ScalarTypes.DateTime.Type()):
		if strings.HasPrefix(example, "datetime(") || strings.HasPrefix(example, "date(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
		} else {
			editText = "datetime(" + example + ")"
		}
	case sameType(typ, ScalarTypes.Guid.Type()):
		if strings.HasPrefix(example, "guid(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
		} else {
			editText = "guid(" + example + ")"
		}
	case isDynamicType(typ):
		if strings.HasPrefix(example, "dynamic(") {
			displayText = completionGooLiteralContent(example)
			matchText = displayText
		} else {
			editText = "dynamic(" + example + ")"
		}
	}

	return NewCompletionItem(CompletionKindExample, displayText,
		WithCompletionBeforeText(editText),
		WithCompletionMatchText(matchText))
}

func completionExactLiteral(text string, kinds ...SyntaxKind) (LexicalToken, bool) {
	tokens := Lex(text)
	if len(tokens) != 2 || tokens[0].Trivia != "" || tokens[0].Text != text || tokens[1].Kind != SyntaxKindEndOfTextToken || tokens[1].Trivia != "" {
		return LexicalToken{}, false
	}
	for _, kind := range kinds {
		if tokens[0].Kind == kind {
			return tokens[0], true
		}
	}
	return LexicalToken{}, false
}

func completionGooLiteralContent(literal string) string {
	open := strings.IndexByte(literal, '(')
	close := strings.LastIndexByte(literal, ')')
	if open >= 0 && close >= 0 && close >= open {
		return literal[open+1 : close]
	}
	return literal
}
