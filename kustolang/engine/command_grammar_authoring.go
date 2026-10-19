package kustolang

// Ported from the authoring helpers in
// src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// CommandGrammarDescriptorOptions models the optional hint and optionality
// arguments of CommandGrammar.CD.
type CommandGrammarDescriptorOptions struct {
	Hint       CompletionHint
	IsOptional bool
}

// CommandGrammarCD constructs a CustomElementDescriptor. With no options it
// uses the upstream defaults: Syntax completion and a required child.
func CommandGrammarCD(name string, options ...CommandGrammarDescriptorOptions) CustomElementDescriptor {
	if len(options) > 1 {
		panic("CommandGrammarCD accepts at most one options value")
	}
	hint := CompletionHintSyntax
	isOptional := false
	if len(options) == 1 {
		hint = options[0].Hint
		isOptional = options[0].IsOptional
	}
	return NewCustomElementDescriptor(name, hint, isOptional)
}

// CommandGrammarCDWithoutName preserves the source overload that specifies
// only completion hint and optionality.
func CommandGrammarCDWithoutName(options ...CommandGrammarDescriptorOptions) CustomElementDescriptor {
	return CommandGrammarCD("", options...)
}

// CommandGrammarCustom constructs a CustomNode parser from an untyped parser
// sequence, matching the source helper's parser-list overload.
func CommandGrammarCustom(parsers []Parser[LexicalToken], shape ...[]CustomElementDescriptor) ResultParser[LexicalToken, SyntaxElement] {
	if len(shape) > 1 {
		panic("CommandGrammarCustom accepts at most one shape")
	}
	customShape := []CustomElementDescriptor(nil)
	if len(shape) == 1 {
		customShape = shape[0]
	}
	if customShape == nil {
		customShape = CustomElementDescriptorDefaultShape(len(parsers))
	}
	return Produce(Sequence(parsers...), func(items []any) SyntaxElement {
		elements := make([]SyntaxElement, len(items))
		for i, item := range items {
			elements[i] = item.(SyntaxElement)
		}
		node, err := NewCustomNodeWithShape(customShape, elements)
		if err != nil {
			panic(err)
		}
		return node
	})
}

// CommandGrammarCustomElements is the variadic spelling of the source
// overloads that take one through nine parsers and use the default shape.
func CommandGrammarCustomElements(parsers ...Parser[LexicalToken]) ResultParser[LexicalToken, SyntaxElement] {
	return CommandGrammarCustom(parsers)
}

// CommandGrammarCustomWithShape is the variadic spelling of the source
// overloads that take one through nine parsers and an explicit shape.
func CommandGrammarCustomWithShape(shape []CustomElementDescriptor, parsers ...Parser[LexicalToken]) ResultParser[LexicalToken, SyntaxElement] {
	return CommandGrammarCustom(parsers, shape)
}

// CommandGrammarCommand wraps custom command content with the leading dot and
// tags the parser with the command name used by completion and recovery.
func CommandGrammarCommand(commandName string, contentParser ResultParser[LexicalToken, SyntaxElement]) ResultParser[LexicalToken, Command] {
	parser := Rule2(
		SyntaxTokenKindParser(SyntaxKindDotToken),
		contentParser,
		func(dot *SyntaxToken, custom SyntaxElement) Command {
			return MustNewCustomCommandTyped(commandName, dot, custom)
		},
	)
	return asResultParser[LexicalToken, Command](parser.WithTag("<" + commandName + ">"))
}

// CommandGrammarCreateMissingToken constructs the token used by RequiredToken.
// A single text preserves its known syntax kind; alternatives use IdentifierToken.
func CommandGrammarCreateMissingToken(texts ...string) SyntaxElement {
	if len(texts) == 1 {
		return CreateMissingSyntaxTokenForText(texts[0])
	}
	return CreateMissingSyntaxTokenForTexts(texts)
}

// CommandGrammarCreateMissingTokenFromList preserves the IReadOnlyList
// overload without requiring callers to expand a slice.
func CommandGrammarCreateMissingTokenFromList(texts []string) SyntaxElement {
	return CommandGrammarCreateMissingToken(texts...)
}

func commandGrammarToken(text string, completionKind CompletionKind, explicitCompletion bool, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	if len(asKind) > 1 {
		panic("command grammar token accepts at most one syntax kind")
	}
	rule := SyntaxMatchTextParser(text, asKind...)
	tagged := asResultParser[LexicalToken, *SyntaxToken](rule.WithTag(defaultSyntaxParserTagForText(text)))
	kind := completionKind
	if !explicitCompletion {
		kind = syntaxCompletionKindForText(text, CompletionKindKeyword)
	}
	if item, ok := CreateSyntaxCompletionItemForText(text, kind, CompletionPriorityNormal); ok {
		tagged = WithSyntaxCompletion(tagged, item)
	}
	return Rule1(tagged, func(token *SyntaxToken) SyntaxElement { return token })
}

// CommandGrammarToken parses one command token. Unknown command words default
// to Keyword completion, unlike the general syntax-token helper.
func CommandGrammarToken(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return commandGrammarToken(text, CompletionKindUnknown, false, asKind...)
}

// CommandGrammarTokenWithCompletion is the Go spelling of the source overload
// that takes an explicit CompletionKind.
func CommandGrammarTokenWithCompletion(text string, completionKind CompletionKind, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return commandGrammarToken(text, completionKind, true, asKind...)
}

// CommandGrammarTokens parses one of several texts with Keyword as the default
// completion kind for words not represented in SyntaxFacts.
func CommandGrammarTokens(texts ...string) ResultParser[LexicalToken, SyntaxElement] {
	parser := SyntaxTokenTextsParser(texts, CompletionKindKeyword)
	return Rule1(parser, func(token *SyntaxToken) SyntaxElement { return token })
}

// CommandGrammarRequiredToken parses text or synthesizes the equivalent
// missing token without consuming input.
func CommandGrammarRequiredToken(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return Required(CommandGrammarToken(text, asKind...), func(Source[LexicalToken], int) SyntaxElement {
		return CommandGrammarCreateMissingToken(text)
	})
}

// CommandGrammarRequiredTokenWithCompletion combines explicit completion with
// the source RequiredToken recovery behavior.
func CommandGrammarRequiredTokenWithCompletion(text string, completionKind CompletionKind, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return Required(CommandGrammarTokenWithCompletion(text, completionKind, asKind...), func(Source[LexicalToken], int) SyntaxElement {
		return CommandGrammarCreateMissingToken(text)
	})
}

// CommandGrammarRequiredTokens parses one of several texts or synthesizes one
// IdentifierToken carrying the alternatives diagnostic.
func CommandGrammarRequiredTokens(texts ...string) ResultParser[LexicalToken, SyntaxElement] {
	return Required(CommandGrammarTokens(texts...), func(Source[LexicalToken], int) SyntaxElement {
		return CommandGrammarCreateMissingToken(texts...)
	})
}

func commandGrammarList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	separatorParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	oneOrMore bool,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	if separatorParser != nil {
		return OListFull(
			elementParser,
			separatorParser,
			elementParser,
			nil,
			nil,
			fnMissingElement,
			nil,
			oneOrMore,
			allowTrailingSeparator,
			func(items []any) SyntaxElement {
				return MakeSyntaxSeparatedList[SyntaxElement](items).SyntaxList
			},
		)
	}

	// CommandGrammar intentionally does not use fnMissingElement for an
	// unseparated list, including the required first element.
	return ListElements(elementParser, oneOrMore, func(elements []SyntaxElement) SyntaxElement {
		return NewTypedSyntaxList("SyntaxElement", elements...)
	})
}

func CommandGrammarZeroOrMoreList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	separatorParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return commandGrammarList(elementParser, separatorParser, fnMissingElement, false, allowTrailingSeparator)
}

func CommandGrammarZeroOrMoreCommaList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return CommandGrammarZeroOrMoreList(elementParser, CommandGrammarToken(","), fnMissingElement, allowTrailingSeparator)
}

func CommandGrammarOneOrMoreList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	separatorParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return commandGrammarList(elementParser, separatorParser, fnMissingElement, true, allowTrailingSeparator)
}

func CommandGrammarOneOrMoreCommaList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return CommandGrammarOneOrMoreList(elementParser, CommandGrammarToken(","), fnMissingElement, allowTrailingSeparator)
}
