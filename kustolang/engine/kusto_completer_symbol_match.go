package kustolang

// Ported from KustoCompleter.GetSymbolMatch,
// GetSymbolMatchFromGrammar and GetSymbolMatchFromSyntaxTree in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (c *kustoCompleter) getSymbolMatch(position int) SymbolMatch {
	if match := c.getSymbolMatchFromGrammar(position); match != SymbolMatchNone {
		return match
	}
	return c.getSymbolMatchFromSyntaxTree(position)
}

func (c *kustoCompleter) getSymbolMatchFromGrammar(position int) SymbolMatch {
	if c == nil || c.code == nil {
		return SymbolMatchNone
	}
	root, startIndex, matchIndex, ok := c.extendedGrammarContext(position)
	if !ok || root == nil || startIndex < 0 || matchIndex < startIndex || matchIndex >= len(c.code.Tokens) {
		return SymbolMatchNone
	}

	parsers := FindAnnotatedParsers(
		NewArraySource(c.code.Tokens[:matchIndex+1]),
		matchIndex,
		root,
		startIndex,
	)
	hints := CompletionHintNone
	for _, parser := range parsers {
		hints |= syntaxCompletionHintFromAnnotations(parser.Annotations())
	}
	return symbolMatchFromCompletionHint(hints)
}

func (c *kustoCompleter) getSymbolMatchFromSyntaxTree(position int) SymbolMatch {
	context, ok := c.tryGetCompletionContext(position)
	if !ok {
		return SymbolMatchNone
	}

	match := symbolMatchFromCompletionHint(
		c.getCompletionHintForContext(position, context.node, context.childIndex, CompletionHintNone),
	)
	for {
		if _, ok := WrapGeneratedSyntaxElement(context.node).(*ParenthesizedExpression); !ok {
			break
		}
		parent := context.node.Parent()
		if parent == nil {
			break
		}
		context.childIndex = context.node.IndexInParent()
		context.node = parent
		match |= symbolMatchFromCompletionHint(
			c.getCompletionHintForContext(position, context.node, context.childIndex, CompletionHintNone),
		)
	}
	return match
}
