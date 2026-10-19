package kustolang

// Ported from the runtime helpers in
// src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// BadCommandParser parses a lone command dot and attaches the source-compatible
// missing-command diagnostic.
func BadCommandParser() ResultParser[LexicalToken, Command] {
	parser := Rule1(
		SyntaxTokenKindParser(SyntaxKindDotToken),
		func(dot *SyntaxToken) Command {
			return MustNewBadCommandTyped(dot, DiagnosticMissingCommand())
		},
	)
	return asResultParser[LexicalToken, Command](parser.WithTag("<bad-command>"))
}

// CommandParserOnFailure appends the missing token produced by the command
// grammar's partial-parser failure handler.
func CommandParserOnFailure(parser Parser[LexicalToken], output *[]any) {
	commandPartialOnFailure(parser, output)
}

// CommandGrammarCommandLimiterParser preserves the CommandGrammar member's
// owner-qualified spelling for callers building command script parsers.
func CommandGrammarCommandLimiterParser() Parser[LexicalToken] {
	return CommandLimiterParser()
}

// CommandGrammarBadCommandParser preserves the CommandGrammar member's
// owner-qualified spelling.
func CommandGrammarBadCommandParser() ResultParser[LexicalToken, Command] {
	return BadCommandParser()
}

// CommandGrammarCreateMissingCommandStatement preserves the owner-qualified
// spelling of the command-list recovery factory.
func CommandGrammarCreateMissingCommandStatement(source Source[LexicalToken], start int) Statement {
	return CreateMissingCommandStatement(source, start)
}

// CommandGrammarIsPossibleStartOfCommand preserves the owner-qualified
// spelling of the command-boundary predicate.
func CommandGrammarIsPossibleStartOfCommand(token LexicalToken, isFirstToken ...bool) bool {
	return IsPossibleStartOfCommand(token, isFirstToken...)
}

// CommandGrammarExpectedTerms returns the sorted, distinct terms selected by
// the source TagFinder traversal. The returned slice is newly allocated.
func CommandGrammarExpectedTerms(parser Parser[LexicalToken]) []string {
	return CommandParserExpectedTerms(parser)
}

// CommandGrammarOnFailure preserves the owner-qualified spelling of the
// partial-parser failure handler.
func CommandGrammarOnFailure(parser Parser[LexicalToken], output *[]any) {
	CommandParserOnFailure(parser, output)
}

// CommandGrammarIsBetterSyntax applies the source command-candidate tie break.
func CommandGrammarIsBetterSyntax(command1 SyntaxElement, command2 SyntaxElement) bool {
	return IsBetterCommandSyntax(command1, command2)
}
