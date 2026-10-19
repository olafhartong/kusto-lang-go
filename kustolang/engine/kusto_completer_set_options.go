package kustolang

import "strings"

// Ported from GetSetOptionCompletions in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (c *kustoCompleter) addSetOptionNameCompletions(position int, builder *kustoCompletionBuilder) bool {
	if !c.isSetOptionNameCompletionContext(position) || c.code.Globals == nil {
		return false
	}
	for _, option := range c.code.Globals.Options {
		if option == nil || option.SymbolName() == "" || option.IsHidden() {
			continue
		}
		name := option.SymbolName()
		builder.add(NewCompletionItem(CompletionKindOption, name,
			WithCompletionBeforeText(BracketNameIfNecessary(name))))
	}
	return true
}

func (c *kustoCompleter) isSetOptionNameCompletionContext(position int) bool {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return false
	}
	found := false
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		statement, ok := WrapGeneratedSyntaxElement(element).(*SetOptionStatement)
		if !ok || statement == nil || statement.SetKeyword() == nil || statement.Name() == nil {
			return
		}
		keyword := statement.SetKeyword()
		name := statement.Name()
		if keyword.Width() == 0 || position < keyword.End() {
			return
		}
		if clause := statement.ValueClause(); clause != nil && clause.Equal() != nil && clause.Equal().Width() > 0 && position >= clause.Equal().TextStart() {
			return
		}
		if name.Width() > 0 {
			found = position >= name.TextStart() && position <= name.End()
			return
		}
		found = strings.TrimSpace(c.code.Text[keyword.End():position]) == ""
	})
	return found
}

func (c *kustoCompleter) getSetOptionCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	option, ok := c.setOptionCompletionContext(position)
	if !ok {
		return completionModeCombined
	}
	for _, example := range option.Examples() {
		matchText := example
		if token, exact := completionExactLiteral(example, SyntaxKindStringLiteralToken); exact {
			matchText = SyntaxTokenFromLexical(token).ValueText()
		}
		builder.add(NewCompletionItem(CompletionKindExample, example,
			WithCompletionBeforeText(example),
			WithCompletionMatchText(matchText)))
	}
	return completionModeIsolated
}

func (c *kustoCompleter) setOptionCompletionContext(position int) (*OptionSymbol, bool) {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return nil, false
	}
	var selected *OptionSymbol
	selectedStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		clause, ok := WrapGeneratedSyntaxElement(element).(*OptionValueClause)
		if !ok || clause == nil || clause.Equal() == nil || clause.Expression() == nil {
			return
		}
		equal := clause.Equal()
		expression := clause.Expression()
		if equal.Width() == 0 || position < equal.End() {
			return
		}
		if expression.Width() > 0 && position > expression.TextStart() {
			return
		}
		if equal.End() < position && strings.TrimSpace(c.code.Text[equal.End():position]) != "" {
			return
		}
		statement, ok := WrapGeneratedSyntaxElement(clause.Parent()).(*SetOptionStatement)
		if !ok || statement == nil || statement.Name() == nil {
			return
		}
		option, _ := statement.Name().ReferencedSymbol().(*OptionSymbol)
		if option == nil || equal.TextStart() < selectedStart {
			return
		}
		selected = option
		selectedStart = equal.TextStart()
	})
	return selected, selected != nil
}
