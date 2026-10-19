package kustolang

// Ported from the authoring helpers in
// src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type CommandGrammarDescriptorOptions = internal.CommandGrammarDescriptorOptions

func CommandGrammarCD(name string, options ...CommandGrammarDescriptorOptions) CustomElementDescriptor {
	return internal.CommandGrammarCD(name, options...)
}

func CommandGrammarCDWithoutName(options ...CommandGrammarDescriptorOptions) CustomElementDescriptor {
	return internal.CommandGrammarCDWithoutName(options...)
}

func CommandGrammarCustom(parsers []Parser[LexicalToken], shape ...[]CustomElementDescriptor) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarCustom(parsers, shape...)
}

func CommandGrammarCustomElements(parsers ...Parser[LexicalToken]) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarCustomElements(parsers...)
}

func CommandGrammarCustomWithShape(shape []CustomElementDescriptor, parsers ...Parser[LexicalToken]) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarCustomWithShape(shape, parsers...)
}

func CommandGrammarCommand(commandName string, contentParser ResultParser[LexicalToken, SyntaxElement]) ResultParser[LexicalToken, Command] {
	return internal.CommandGrammarCommand(commandName, contentParser)
}

func CommandGrammarCreateMissingToken(texts ...string) SyntaxElement {
	return internal.CommandGrammarCreateMissingToken(texts...)
}

func CommandGrammarCreateMissingTokenFromList(texts []string) SyntaxElement {
	return internal.CommandGrammarCreateMissingTokenFromList(texts)
}

func CommandGrammarToken(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarToken(text, asKind...)
}

func CommandGrammarTokenWithCompletion(text string, completionKind CompletionKind, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarTokenWithCompletion(text, completionKind, asKind...)
}

func CommandGrammarTokens(texts ...string) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarTokens(texts...)
}

func CommandGrammarRequiredToken(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarRequiredToken(text, asKind...)
}

func CommandGrammarRequiredTokenWithCompletion(text string, completionKind CompletionKind, asKind ...SyntaxKind) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarRequiredTokenWithCompletion(text, completionKind, asKind...)
}

func CommandGrammarRequiredTokens(texts ...string) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarRequiredTokens(texts...)
}

func CommandGrammarZeroOrMoreList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	separatorParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarZeroOrMoreList(elementParser, separatorParser, fnMissingElement, allowTrailingSeparator)
}

func CommandGrammarZeroOrMoreCommaList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarZeroOrMoreCommaList(elementParser, fnMissingElement, allowTrailingSeparator)
}

func CommandGrammarOneOrMoreList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	separatorParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarOneOrMoreList(elementParser, separatorParser, fnMissingElement, allowTrailingSeparator)
}

func CommandGrammarOneOrMoreCommaList(
	elementParser ResultParser[LexicalToken, SyntaxElement],
	fnMissingElement func(Source[LexicalToken], int) SyntaxElement,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, SyntaxElement] {
	return internal.CommandGrammarOneOrMoreCommaList(elementParser, fnMissingElement, allowTrailingSeparator)
}
