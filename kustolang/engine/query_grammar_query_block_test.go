package kustolang

import "testing"

func TestQueryGrammarPublicDirectiveAndStatementParsers(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	directive := parseQueryGrammarProduction(t, grammar.Directive, "#set query_now=datetime(2024-01-01)")
	if directive.Token().Kind() != SyntaxKindDirectiveToken || directive.Token().RawText() != "#set query_now=datetime(2024-01-01)" {
		t.Fatalf("directive = %#v", directive)
	}

	letStatement := parseQueryGrammarProduction(t, grammar.Statement, "let n = 1")
	if _, ok := letStatement.(*LetStatement); !ok {
		t.Fatalf("let statement = %T, want *LetStatement", letStatement)
	}
	queryStatement := parseQueryGrammarProduction(t, grammar.Statement, "T | take 1")
	if _, ok := queryStatement.(*ExpressionStatement); !ok {
		t.Fatalf("query statement = %T, want *ExpressionStatement", queryStatement)
	}
	assertQueryGrammarProductionRejects(t, grammar.Statement, ";")
}

func TestQueryGrammarPublicStatementListSemicolons(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	statements := parseQueryGrammarProduction(t, grammar.StatementList, "let n = 1; print n")
	if statements.Count() != 2 {
		t.Fatalf("statement count = %d, want 2", statements.Count())
	}
	if _, ok := statements.At(0).Element().(*LetStatement); !ok {
		t.Fatalf("first statement = %T, want *LetStatement", statements.At(0).Element())
	}
	if separator := statements.At(0).Separator(); separator == nil || separator.Kind() != SyntaxKindSemicolonToken || separator.Width() == 0 {
		t.Fatalf("first separator = %#v, want present semicolon", separator)
	}
	if _, ok := statements.At(1).Element().(*ExpressionStatement); !ok || statements.At(1).Separator() != nil {
		t.Fatalf("second statement = %T, separator %#v", statements.At(1).Element(), statements.At(1).Separator())
	}

	trailing := parseQueryGrammarProduction(t, grammar.StatementList, "let n = 1;")
	if trailing.Count() != 1 || trailing.At(0).Separator() == nil || trailing.At(0).Separator().Width() == 0 {
		t.Fatalf("trailing-semicolon list = %#v", trailing)
	}
	if diagnostics := trailing.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("trailing-semicolon diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMacroExpandSubQueryForwardsStatementList(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	statements := parseQueryGrammarProduction(t, grammar.MacroExpandSubQuery, "let n = 1; print n")
	if statements.Count() != 2 || grammar.MacroExpandSubQuery.Tag() != "<macro-expand-subquery>" {
		t.Fatalf("macro-expand subquery = count:%d tag:%q", statements.Count(), grammar.MacroExpandSubQuery.Tag())
	}
	if _, ok := statements.At(0).Element().(*LetStatement); !ok {
		t.Fatalf("macro-expand first statement = %T, want *LetStatement", statements.At(0).Element())
	}
	if _, ok := statements.At(1).Element().(*ExpressionStatement); !ok {
		t.Fatalf("macro-expand second statement = %T, want *ExpressionStatement", statements.At(1).Element())
	}
}

func TestQueryGrammarPublicStatementListRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingSemicolon := parseQueryGrammarProduction(t, grammar.StatementList, "let n = 1 print n")
	if missingSemicolon.Count() != 2 {
		t.Fatalf("missing-semicolon statement count = %d, want 2", missingSemicolon.Count())
	}
	separator := missingSemicolon.At(0).Separator()
	if separator == nil || separator.Kind() != SyntaxKindSemicolonToken || separator.Width() != 0 ||
		!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ;") {
		t.Fatalf("recovered separator = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}

	leadingSemicolon := parseQueryGrammarProduction(t, grammar.StatementList, "; print 1")
	if leadingSemicolon.Count() != 2 {
		t.Fatalf("leading-semicolon statement count = %d, want 2", leadingSemicolon.Count())
	}
	missing := leadingSemicolon.At(0)
	if missing.Separator() == nil || missing.Separator().Width() == 0 ||
		!diagnosticsContainMessage(missing.Element().ContainedDiagnostics(), "KS006", "Missing statement") {
		t.Fatalf("leading-semicolon recovery = element:%#v separator:%#v diagnostics:%#v", missing.Element(), missing.Separator(), missing.Element().ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicQueryBlockDirectivesAndSkippedTail(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	block := parseQueryGrammarQueryBlock(t, grammar.QueryBlock, "#set first=1\n#set second=2\nlet n = 1; print n ) trailing")

	if block.Directives().Count() != 2 || block.Directives().At(0).Token().RawText() != "#set first=1" || block.Directives().At(1).Token().RawText() != "#set second=2" {
		t.Fatalf("directives = %#v", block.Directives())
	}
	if block.Statements().Count() != 2 {
		t.Fatalf("statement count = %d, want 2", block.Statements().Count())
	}
	skipped := block.SkippedTokens()
	if skipped == nil || skipped.Tokens().Count() != 2 {
		t.Fatalf("skipped tail = %#v", skipped)
	}
	for i := 0; i < skipped.Tokens().Count(); i++ {
		diagnostics := skipped.Tokens().At(i).Diagnostics()
		if i == 0 {
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS198" {
				t.Fatalf("first skipped token diagnostics = %#v", diagnostics)
			}
		} else if diagnosticsContain(diagnostics, "KS198") {
			t.Fatalf("skipped token %d also has KS198: %#v", i, diagnostics)
		}
	}
	if end := block.EndOfQuery(); end == nil || end.Kind() != SyntaxKindEndOfTextToken {
		t.Fatalf("end of query = %#v", end)
	}
}

func TestQueryGrammarPublicQueryBlockOptionalEndToken(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tokens := Lex("print 1")
	result := grammar.QueryBlock.ParseValue(NewArraySource(tokens[:len(tokens)-1]), 0)
	if result.Length != len(tokens)-1 || result.Value == nil {
		t.Fatalf("query block without end token = %#v", result)
	}
	if result.Value.EndOfQuery() != nil || result.Value.Statements().Count() != 1 {
		t.Fatalf("query block without end token = end:%#v statements:%#v", result.Value.EndOfQuery(), result.Value.Statements())
	}
}

func TestQueryBlockTypedConstructorAccessorsAndClone(t *testing.T) {
	parsed := parseQueryGrammarQueryBlock(t, QueryGrammarFrom(nil).QueryBlock, "#set option=1\nprint 1 )")
	constructed := MustNewQueryBlockTyped(
		AsSyntaxListOf[*Directive](parsed.Directives().SyntaxList.Clone(true).(*SyntaxList)),
		AsSeparatedSyntaxListOf[Statement](parsed.Statements().SyntaxList.Clone(true).(*SyntaxList)),
		cloneSyntaxElementPreservingGeneratedWrapper(parsed.SkippedTokens(), true).(*SkippedTokens),
		parsed.EndOfQuery().Clone(true).(*SyntaxToken),
		DiagnosticMissingStatement(),
	)

	if constructed.Directives().Count() != 1 || constructed.Statements().Count() != 1 || constructed.SkippedTokens() == nil || constructed.EndOfQuery() == nil {
		t.Fatalf("constructed accessors = directives:%#v statements:%#v skipped:%#v end:%#v", constructed.Directives(), constructed.Statements(), constructed.SkippedTokens(), constructed.EndOfQuery())
	}
	clone := constructed.CloneTyped(true)
	if clone == constructed || clone.SyntaxNode == constructed.SyntaxNode {
		t.Fatal("typed clone reused original query block")
	}
	if clone.Directives().Count() != 1 || clone.Statements().Count() != 1 || clone.SkippedTokens() == nil || clone.EndOfQuery() == nil {
		t.Fatalf("clone accessors = directives:%#v statements:%#v skipped:%#v end:%#v", clone.Directives(), clone.Statements(), clone.SkippedTokens(), clone.EndOfQuery())
	}
	if !diagnosticsContainMessage(clone.Diagnostics(), "KS006", "Missing statement") {
		t.Fatalf("clone diagnostics = %#v", clone.Diagnostics())
	}
	withoutDiagnostics := constructed.CloneTyped(false)
	if diagnostics := withoutDiagnostics.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("clone without diagnostics = %#v", diagnostics)
	}
}

func parseQueryGrammarQueryBlock(t *testing.T, parser ResultParser[LexicalToken, *QueryBlock], text string) *QueryBlock {
	t.Helper()
	tokens := Lex(text)
	result := parser.ParseValue(NewArraySource(tokens), 0)
	if result.Length != len(tokens) || result.Value == nil {
		t.Fatalf("parse query block %q = %#v, want %d consumed tokens", text, result, len(tokens))
	}
	return result.Value
}
