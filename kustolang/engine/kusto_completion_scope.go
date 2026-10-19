package kustolang

import "strings"

// queryCompletionScope is the bounded semantic context used by the first
// native query-symbol completion tranche.
type queryCompletionScope struct {
	root          bool
	bounded       bool
	catalogScope  bool
	rowScope      *TableSymbol
	rightRowScope *TableSymbol
	inputRowScope *TableSymbol
	invoke        bool
	tableSource   bool
	scopeKind     ScopeKind
	globals       *GlobalState
	match         SymbolMatch
	hint          CompletionHint
}

func getQueryCompletionScope(code *KustoCode, position int) queryCompletionScope {
	if code == nil || !kustoCodeIsQueryCompletionPosition(code, position) || code.Globals == nil {
		return queryCompletionScope{}
	}

	globals := code.globalsInScope(position)
	barIndex := queryCompletionTopLevelBarBefore(code, position)
	if barIndex < 0 {
		if queryCompletionIsAtStatementStart(code, position) {
			return queryCompletionScope{root: true, globals: globals}
		}
		context := code.getRowScopeContext(position)
		if context.bounded {
			return queryCompletionScope{
				bounded:       true,
				catalogScope:  true,
				rowScope:      context.rowScope,
				rightRowScope: context.rightRowScope,
				scopeKind:     code.getScopeKindContext(position),
				globals:       globals,
			}
		}
		return queryCompletionScope{}
	}
	operatorKind, ok := queryCompletionOperatorAfterBar(code, barIndex, position)
	if !ok {
		return queryCompletionScope{}
	}
	context := code.getRowScopeContext(position)
	if context.bounded {
		invoke := operatorKind == SyntaxKindInvokeKeyword
		return queryCompletionScope{
			bounded:       true,
			rowScope:      context.rowScope,
			rightRowScope: context.rightRowScope,
			inputRowScope: queryCompletionPipeInputRowScope(code, position),
			invoke:        invoke,
			tableSource:   operatorKind == SyntaxKindUnionKeyword,
			scopeKind:     code.getScopeKindContext(position),
			globals:       globals,
		}
	}

	sourceName, ok := queryCompletionSimpleSourceName(code, barIndex)
	if !ok || globals == nil || globals.Database == nil {
		return queryCompletionScope{}
	}
	return queryCompletionScope{
		bounded:       true,
		rowScope:      globals.Database.GetTable(sourceName),
		inputRowScope: globals.Database.GetTable(sourceName),
		invoke:        operatorKind == SyntaxKindInvokeKeyword,
		tableSource:   operatorKind == SyntaxKindUnionKeyword,
		scopeKind:     code.getScopeKindContext(position),
		globals:       globals,
	}
}

func queryCompletionPipeInputRowScope(code *KustoCode, position int) *TableSymbol {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	var row *TableSymbol
	bestWidth := int(^uint(0) >> 1)
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		pipe, ok := WrapGeneratedSyntaxElement(element).(*PipeExpression)
		if !ok || pipe.Operator() == nil || pipe.Expression() == nil {
			return
		}
		operatorStart, found := syntaxElementTriviaStartOf(pipe, pipe.Operator(), triviaStart)
		if !found {
			return
		}
		operatorEnd := operatorStart + pipe.Operator().FullWidth()
		inTrailingTrivia := position > operatorEnd && position <= len(code.Text) && strings.TrimSpace(code.Text[operatorEnd:position]) == ""
		if position < operatorStart || (position > operatorEnd && !inTrailingTrivia) {
			return
		}
		width := operatorEnd - operatorStart
		if width < bestWidth {
			row = code.expressionTableResult(pipe.Expression())
			bestWidth = width
		}
	})
	return row
}

func queryCompletionTopLevelBarBefore(code *KustoCode, position int) int {
	depth := 0
	barIndex := -1
	for i, token := range code.Tokens {
		if i >= len(code.TokenStarts) || queryCompletionTokenTextStart(code, i) >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		switch token.Kind {
		case SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken, SyntaxKindOpenBraceToken:
			depth++
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			if depth > 0 {
				depth--
			}
		case SyntaxKindBarToken:
			if depth == 0 {
				barIndex = i
			}
		case SyntaxKindSemicolonToken:
			if depth == 0 {
				barIndex = -1
			}
		case SyntaxKindDirectiveToken:
			if depth == 0 {
				barIndex = -1
			}
		}
	}
	return barIndex
}

func queryCompletionIsAtStatementStart(code *KustoCode, position int) bool {
	depth := 0
	statementStart := 0
	for i, token := range code.Tokens {
		if i >= len(code.TokenStarts) || queryCompletionTokenTextStart(code, i) >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		switch token.Kind {
		case SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken, SyntaxKindOpenBraceToken:
			depth++
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			if depth > 0 {
				depth--
			}
		case SyntaxKindSemicolonToken:
			if depth == 0 {
				statementStart = i + 1
			}
		case SyntaxKindDirectiveToken:
			if depth == 0 {
				statementStart = i + 1
			}
		}
	}
	for i := statementStart; i < len(code.Tokens) && i < len(code.TokenStarts); i++ {
		token := code.Tokens[i]
		if queryCompletionTokenTextStart(code, i) >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		if token.Text != "" {
			return false
		}
	}
	return true
}

func queryCompletionOperatorAfterBar(code *KustoCode, barIndex int, position int) (SyntaxKind, bool) {
	for i := barIndex + 1; i < len(code.Tokens) && i < len(code.TokenStarts); i++ {
		token := code.Tokens[i]
		if queryCompletionTokenTextStart(code, i) >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		if token.Text != "" {
			return token.Kind, true
		}
	}
	return SyntaxKindNone, false
}

func queryCompletionSimpleSourceName(code *KustoCode, barIndex int) (string, bool) {
	name := ""
	statementStart := queryCompletionStatementStartBefore(code, barIndex)
	for i := statementStart; i < barIndex && i < len(code.Tokens); i++ {
		token := code.Tokens[i]
		if token.Kind == SyntaxKindEndOfTextToken || token.Text == "" {
			continue
		}
		if name != "" || !kustoCompleterIsKeywordOrIdentifierOrDash(token.Kind) || token.Kind == SyntaxKindMinusToken {
			return "", false
		}
		name = token.Text
	}
	return name, name != ""
}

func queryCompletionStatementStartBefore(code *KustoCode, beforeIndex int) int {
	depth := 0
	statementStart := 0
	for i := 0; i < beforeIndex && i < len(code.Tokens); i++ {
		switch code.Tokens[i].Kind {
		case SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken, SyntaxKindOpenBraceToken:
			depth++
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			if depth > 0 {
				depth--
			}
		case SyntaxKindSemicolonToken:
			if depth == 0 {
				statementStart = i + 1
			}
		case SyntaxKindDirectiveToken:
			if depth == 0 {
				statementStart = i + 1
			}
		}
	}
	return statementStart
}

func queryCompletionTokenTextStart(code *KustoCode, index int) int {
	return kustoCompleterLexicalTextStart(code.TokenStarts[index], code.Tokens[index])
}
