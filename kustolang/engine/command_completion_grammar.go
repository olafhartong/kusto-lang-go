package kustolang

import "strings"

// Ported from the command-completion annotation flow used by
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs over CommandGrammar.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type commandCompletionRuleAnnotation string

func (g *CommandGrammar) CommandCompletionParser() Parser[LexicalToken] {
	if g == nil {
		return nil
	}
	g.commandCompletionOnce.Do(func() {
		g.commandCompletionParser = g.newCommandCompletionParser()
	})
	return g.commandCompletionParser
}

func (g *CommandGrammar) newCommandCompletionParser() Parser[LexicalToken] {
	if len(g.commandParsers) == 0 {
		return nil
	}

	parsers := make([]Parser[LexicalToken], 0, len(g.commandParsers))
	for _, info := range g.commandParsers {
		parser := g.commandDescriptorElementListParser(info.Elements, true)
		if parser == nil && len(info.Tokens) > 0 {
			parser = g.commandCompletionTextSequenceParser(info.Tokens)
		}
		if parser != nil {
			parsers = append(parsers, parser.WithTag("<"+info.Kind+">"))
		}
	}
	if len(parsers) == 0 {
		return nil
	}
	return Sequence(
		HiddenSyntaxTokenKindParser(SyntaxKindDotToken),
		Best(parsers...),
	).WithTag("<command-completion>")
}

func (g *CommandGrammar) partialCommandParsers() []Parser[LexicalToken] {
	if g == nil {
		return nil
	}
	g.commandPartialOnce.Do(func() {
		parsers := make([]Parser[LexicalToken], 0, len(g.commandParsers))
		for _, info := range g.commandParsers {
			parser := g.commandDescriptorElementListParser(info.Elements, false)
			if parser == nil && len(info.Tokens) > 0 {
				parser = g.commandCompletionTextSequenceParser(info.Tokens)
			}
			if parser != nil {
				parsers = append(parsers, Sequence(
					HiddenSyntaxTokenKindParser(SyntaxKindDotToken),
					parser,
				).WithTag(info.Kind))
			}
		}
		g.commandPartialParsers = parsers
	})
	return append([]Parser[LexicalToken]{}, g.commandPartialParsers...)
}

func (g *CommandGrammar) commandCompletionTextSequenceParser(texts []string) Parser[LexicalToken] {
	parsers := make([]Parser[LexicalToken], 0, len(texts))
	for _, text := range texts {
		parsers = append(parsers, commandCompletionTokenTextsParser([]string{text}))
	}
	return Sequence(parsers...)
}

func (g *CommandGrammar) commandDescriptorElementListParser(elements []commandTokenElement, completion bool) Parser[LexicalToken] {
	parsers := make([]Parser[LexicalToken], 0, len(elements))
	for i := 0; i < len(elements); {
		element := elements[i]
		if element.OptionalGroup != 0 {
			groupEnd := i + 1
			for groupEnd < len(elements) && elements[groupEnd].OptionalGroup == element.OptionalGroup {
				groupEnd++
			}
			groupParser := g.commandDescriptorRequiredElementListParser(elements[i:groupEnd], completion)
			if groupParser != nil {
				if commandCompletionElementsCanSkip(elements[i:groupEnd]) {
					groupParser = ZeroOrOne(groupParser)
				}
				parsers = append(parsers, groupParser)
			}
			i = groupEnd
			continue
		}

		parser := g.commandDescriptorElementParser(element, completion)
		if parser != nil {
			if commandElementCanSkip(element) {
				parser = ZeroOrOne(parser)
			}
			parsers = append(parsers, parser)
		}
		i++
	}
	if len(parsers) == 0 {
		return nil
	}
	return Sequence(parsers...)
}

func (g *CommandGrammar) commandDescriptorRequiredElementListParser(elements []commandTokenElement, completion bool) Parser[LexicalToken] {
	parsers := make([]Parser[LexicalToken], 0, len(elements))
	for _, element := range elements {
		parser := g.commandDescriptorElementParser(element, completion)
		if parser == nil {
			return nil
		}
		parsers = append(parsers, parser)
	}
	if len(parsers) == 0 {
		return nil
	}
	return Sequence(parsers...)
}

func (g *CommandGrammar) commandDescriptorElementParser(element commandTokenElement, completion bool) Parser[LexicalToken] {
	var parser Parser[LexicalToken]
	switch {
	case len(element.Choices) > 0:
		choices := make([]Parser[LexicalToken], 0, len(element.Choices))
		for _, choice := range element.Choices {
			if choiceParser := g.commandDescriptorElementParser(choice, completion); choiceParser != nil {
				choices = append(choices, choiceParser)
			}
		}
		if len(choices) > 0 {
			parser = Best(choices...)
		}
	case len(element.Children) > 0:
		parser = g.commandDescriptorElementListParser(element.Children, completion)
	case len(element.List) > 0:
		parser = g.commandDescriptorListElementParser(element, completion)
	case len(element.Alternatives) > 0:
		if completion {
			parser = commandCompletionTokenTextsParser(element.Alternatives)
		} else {
			parser = commandPartialTokenTextsParser(element.Alternatives)
		}
	case element.Rule != "":
		parser = g.commandCompletionRuleParser(element.Rule)
	}

	if parser == nil {
		return nil
	}
	if len(element.Forbidden) > 0 {
		guard := Parser[LexicalToken](commandCompletionTokenTextsParser(element.Forbidden))
		if !completion {
			guard = commandPartialTokenTextsParser(element.Forbidden)
		}
		// CommandGenerator emits If(Not(Token(...)), nameLikeParser). Not consumes
		// one item on success, which is also how PartialParser recognizes that the
		// conditional body is eligible while leaving the body at the same start.
		if result, ok := parser.(ResultParser[LexicalToken, SyntaxElement]); ok {
			parser = IfValue(Not(guard), result)
		} else {
			parser = If(Not(guard), parser)
		}
	}
	if element.Hidden {
		if completion {
			return commandCompletionHideAnnotations(parser)
		}
		return parser.Hide()
	}
	if element.HasHint {
		if result, ok := parser.(ResultParser[LexicalToken, SyntaxElement]); ok {
			parser = WithSyntaxCompletionHint(result, element.Hint)
		} else {
			annotations := stripSyntaxCompletionHintAnnotations(parser.Annotations())
			annotations = append(annotations, element.Hint)
			parser = parser.WithAnnotations(annotations)
		}
	}
	return parser
}

func commandCompletionHideAnnotations(parser Parser[LexicalToken]) Parser[LexicalToken] {
	annotations := stripSyntaxCompletionItemAnnotations(parser.Annotations())
	annotations = stripSyntaxCompletionHintAnnotations(annotations)
	return parser.WithAnnotations(annotations).Hide()
}

func (g *CommandGrammar) commandDescriptorListElementParser(element commandTokenElement, completion bool) Parser[LexicalToken] {
	if len(element.List) != 1 {
		return nil
	}
	item := g.commandDescriptorElementParser(element.List[0], completion)
	if item == nil {
		return nil
	}

	var repeated Parser[LexicalToken]
	if element.ListSeparator == "" {
		repeated = item
		if element.ListAllowEmpty {
			return ZeroOrMore(repeated)
		}
		return OneOrMore(repeated)
	} else {
		separator := Parser[LexicalToken](commandCompletionTokenTextsParser([]string{element.ListSeparator}))
		if !completion {
			separator = commandPartialTokenTextsParser([]string{element.ListSeparator})
		}
		repeated = Sequence(separator, item)
	}
	list := Sequence(item, ZeroOrMore(repeated))
	if element.ListAllowEmpty {
		return ZeroOrOne(list)
	}
	return list
}

func (g *CommandGrammar) commandCompletionRuleParser(rule string) Parser[LexicalToken] {
	state := newCommandParserState(nil, g)
	parser := state.commandRuleParser(rule)
	if parser == nil {
		return nil
	}
	annotations := parser.Annotations()
	annotations = append(annotations, commandCompletionRuleAnnotation(rule))
	return parser.WithAnnotations(annotations)
}

func commandCompletionRuleFromAnnotations(annotations []any) string {
	for _, annotation := range annotations {
		if rule, ok := annotation.(commandCompletionRuleAnnotation); ok {
			return string(rule)
		}
	}
	return ""
}

func commandCompletionTokenTextsParser(texts []string) ResultParser[LexicalToken, *SyntaxToken] {
	parsers := make([]ResultParser[LexicalToken, *SyntaxToken], 0, len(texts))
	items := make([]CompletionItem, 0, len(texts))
	for _, text := range texts {
		parsers = append(parsers, commandCompletionTokenTextParser(text))
		if item, ok := commandSyntaxCompletionItem(text); ok {
			items = append(items, item)
		}
	}
	parser := asResultParser[LexicalToken, *SyntaxToken](
		FirstValue(parsers...).WithTag(strings.Join(syntaxTextTags(texts), " | ")),
	)
	if len(items) > 0 {
		return WithSyntaxCompletion(parser, items...)
	}
	return parser
}

func commandCompletionTokenTextParser(text string) ResultParser[LexicalToken, *SyntaxToken] {
	return MatchConsumerValue(
		func(source Source[LexicalToken], start int) int {
			return matchesLexicalTextIgnoreCase(source, start, text)
		},
		func(source Source[LexicalToken], start int, length int) *SyntaxToken {
			return ProduceSyntaxTokenFromText(source, start, length, text)
		},
	)
}

func commandPartialTokenTextsParser(texts []string) Parser[LexicalToken] {
	parser := commandCompletionTokenTextsParser(texts)
	converted := Rule1(parser, func(token *SyntaxToken) SyntaxElement { return token })
	return converted.WithTag(parser.Tag())
}

func commandCompletionElementsCanSkip(elements []commandTokenElement) bool {
	for _, element := range elements {
		if !commandElementCanSkip(element) {
			return false
		}
	}
	return true
}
