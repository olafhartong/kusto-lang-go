package kustolang

import "testing"

func TestKnownCommandEqualLengthPrefersBetterSyntax(t *testing.T) {
	bad := commandParserInfo{Parse: func(p *queryParserState) Command {
		dot := p.parseToken(SyntaxKindDotToken)
		name := p.parseToken(SyntaxKindIdentifierToken)
		if dot == nil || name == nil {
			return nil
		}
		return MustNewBadCommandTyped(dot, DiagnosticMissingCommand())
	}}
	good := commandParserInfo{Parse: func(p *queryParserState) Command {
		dot := p.parseToken(SyntaxKindDotToken)
		name := p.parseToken(SyntaxKindIdentifierToken)
		if dot == nil || name == nil {
			return nil
		}
		parts := AsSyntaxListOf[SyntaxElement](NewTypedSyntaxList("SyntaxElement", name))
		return MustNewUnknownCommandTyped(dot, parts)
	}}
	grammar := NewCommandGrammar(DefaultGlobalState)
	grammar.commandParsers = []commandParserInfo{bad, good}

	block := grammar.ParseCommandBlock(`.x`)
	if diagnostics := block.GetContainedSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("selected equal-length command diagnostics = %#v", diagnostics)
	}
	separated := block.Statements().At(0)
	statement, _ := separated.Element().(*ExpressionStatement)
	if statement == nil || statement.Expression() == nil || statement.Expression().Kind() != SyntaxKindUnknownCommand {
		t.Fatalf("selected command = %#v", separated)
	}
}
