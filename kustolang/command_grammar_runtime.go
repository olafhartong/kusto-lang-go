package kustolang

// Ported from the runtime helpers in
// src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func CommandLimiterParser() Parser[LexicalToken] {
	return internal.CommandLimiterParser()
}

func BadCommandParser() ResultParser[LexicalToken, Command] {
	return internal.BadCommandParser()
}

func CreateMissingCommandStatement(source Source[LexicalToken], start int) Statement {
	return internal.CreateMissingCommandStatement(source, start)
}

func IsPossibleStartOfCommand(token LexicalToken, isFirstToken ...bool) bool {
	return internal.IsPossibleStartOfCommand(token, isFirstToken...)
}

func CommandParserExpectedTerms[TInput any](parser Parser[TInput]) []string {
	return internal.CommandParserExpectedTerms(parser)
}

func CommandParserOnFailure(parser Parser[LexicalToken], output *[]any) {
	internal.CommandParserOnFailure(parser, output)
}

func IsBetterCommandSyntax(command1 SyntaxElement, command2 SyntaxElement) bool {
	return internal.IsBetterCommandSyntax(command1, command2)
}

func CommandGrammarCommandLimiterParser() Parser[LexicalToken] {
	return internal.CommandGrammarCommandLimiterParser()
}

func CommandGrammarBadCommandParser() ResultParser[LexicalToken, Command] {
	return internal.CommandGrammarBadCommandParser()
}

func CommandGrammarCreateMissingCommandStatement(source Source[LexicalToken], start int) Statement {
	return internal.CommandGrammarCreateMissingCommandStatement(source, start)
}

func CommandGrammarIsPossibleStartOfCommand(token LexicalToken, isFirstToken ...bool) bool {
	return internal.CommandGrammarIsPossibleStartOfCommand(token, isFirstToken...)
}

func CommandGrammarExpectedTerms(parser Parser[LexicalToken]) []string {
	return internal.CommandGrammarExpectedTerms(parser)
}

func CommandGrammarOnFailure(parser Parser[LexicalToken], output *[]any) {
	internal.CommandGrammarOnFailure(parser, output)
}

func CommandGrammarIsBetterSyntax(command1 SyntaxElement, command2 SyntaxElement) bool {
	return internal.CommandGrammarIsBetterSyntax(command1, command2)
}
