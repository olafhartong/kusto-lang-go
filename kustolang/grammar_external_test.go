package kustolang_test

import (
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryGrammarValidAndMalformed(t *testing.T) {
	grammar := kustolang.QueryGrammarFrom(kustolang.DefaultGlobalState)
	if grammar == nil || grammar.QueryBlock == nil || grammar.Expression == nil || grammar.QueryOperator == nil {
		t.Fatal("query grammar did not expose its typed productions")
	}

	valid := kustolang.ParseFirstSyntax(grammar.QueryBlock, `StormEvents | where State == "TEXAS" | summarize count() by EventType`)
	if valid == nil || valid.Kind() != kustolang.SyntaxKindQueryBlock {
		t.Fatalf("valid query block = %#v", valid)
	}
	if diagnostics := valid.GetContainedSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid grammar diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseFirstSyntax(grammar.QueryBlock, `StormEvents | where State ==`)
	if malformed == nil || !kustolang.SyntaxHasMissingChildren(malformed) {
		t.Fatalf("malformed query did not recover with a missing node: %#v", malformed)
	}
	if diagnostics := malformed.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed query grammar parse produced no diagnostics")
	}
}

func TestPublicQueryGrammarMalformedUnicode(t *testing.T) {
	grammar := kustolang.QueryGrammarFrom(nil)
	malformed := kustolang.ParseFirstSyntax(grammar.QueryBlock, `print Emoji = "😀", Broken = strcat("x", )`)
	if malformed == nil {
		t.Fatal("Unicode query did not produce a recoverable syntax tree")
	}
	if diagnostics := malformed.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed Unicode query produced no syntax diagnostics")
	}

	code := kustolang.Parse(`print Emoji = "😀" | where`)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("KustoCode facade did not report malformed Unicode syntax")
	}
}

func TestPublicCommandGrammarValidAndMalformed(t *testing.T) {
	grammar := kustolang.CommandGrammarFrom(kustolang.DefaultGlobalState)
	if grammar == nil || grammar.CommandBlock == nil || grammar.PredefinedRules == nil {
		t.Fatal("command grammar was not initialized")
	}

	valid := kustolang.ParseFirstSyntax(grammar.CommandBlock, `.show version`)
	if valid == nil || valid.Kind() != kustolang.SyntaxKindCommandBlock {
		t.Fatalf("valid command block = %#v", valid)
	}
	if diagnostics := valid.GetContainedSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid command diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseFirstSyntax(grammar.CommandBlock, `.show table`)
	if malformed == nil {
		t.Fatal("malformed command did not produce a recoverable block")
	}
	if diagnostics := malformed.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed command produced no syntax diagnostics")
	}
}

func TestPublicParserCombinatorsAcceptAndRejectInput(t *testing.T) {
	identifier := kustolang.Text(kustolang.OneOrMore(kustolang.Match(func(ch byte) bool {
		return ch >= 'a' && ch <= 'z'
	})))
	assignment := kustolang.Rule3(
		identifier,
		kustolang.TextLiteral("="),
		identifier,
		func(name, _ string, value string) string { return name + ":" + value },
	)

	parsed := kustolang.ParseTextValue(assignment, "alpha=beta")
	if !parsed.Succeeded() || parsed.Length != len("alpha=beta") || parsed.Value != "alpha:beta" {
		t.Fatalf("custom parser result = %#v", parsed)
	}
	if _, ok := kustolang.TryParseTextValue(assignment, "alpha==beta"); ok {
		t.Fatal("custom parser accepted incorrect syntax")
	}
	if got := kustolang.DescribeParser(assignment); strings.TrimSpace(got) == "" {
		t.Fatal("custom parser has no description")
	}

	digit := kustolang.MatchValue(
		func(ch byte) bool { return ch >= '0' && ch <= '9' },
		func(ch byte) int { return int(ch - '0') },
	)
	sum := kustolang.ApplyZeroOrMore(digit, func(kustolang.LeftValue[int]) kustolang.RightParser[byte, int] {
		return kustolang.RuleRight2(
			kustolang.TextLiteral("+"),
			digit,
			func(left int, _ string, right int) int { return left + right },
		)
	})
	result := kustolang.ParseTextValue(sum, "1+2+3")
	if !result.Succeeded() || result.Length != 5 || result.Value != 6 {
		t.Fatalf("right parser result = %#v", result)
	}

	cache := kustolang.NewSourceCache()
	created := kustolang.SourceCacheGetOrCreateFunc(cache, func() *int { value := 7; return &value })
	cached, ok := kustolang.SourceCacheTryGetValue[*int](cache)
	if !ok || cached != created || *cached != 7 {
		t.Fatalf("source cache = created:%p cached:%p ok:%t", created, cached, ok)
	}
}

func TestPublicSyntaxTokenConstruction(t *testing.T) {
	lexical := kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, " ", "StormEvents")
	token := kustolang.SyntaxTokenFromLexical(lexical)
	if token.Kind() != kustolang.SyntaxKindIdentifierToken || token.RawText() != "StormEvents" || token.Trivia() != " " {
		t.Fatalf("constructed syntax token = %#v", token)
	}

	missing := kustolang.CreateMissingSyntaxToken(
		kustolang.SyntaxKindCloseParenToken,
		kustolang.NewDiagnostic("TEST001", "missing close parenthesis"),
	)
	if !missing.IsMissing() || !missing.HasSyntaxDiagnostics() {
		t.Fatalf("missing token = %#v", missing)
	}

	list := kustolang.NewTypedSyntaxList("SyntaxToken", token, missing)
	if list.ChildCount() != 2 || len(list.GetContainedSyntaxDiagnostics()) == 0 {
		t.Fatalf("constructed list = %#v", list)
	}
}

func TestPublicGeneratedTypedSyntaxConstruction(t *testing.T) {
	literalToken := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindLongLiteralToken, " ", "1"),
	)
	literal := kustolang.MustNewLiteralExpressionTyped(kustolang.SyntaxKindLongLiteralExpression, literalToken)
	separated := kustolang.NewSeparatedElement(literal, nil)
	expressions := kustolang.AsSeparatedSyntaxListOf[kustolang.Expression](
		kustolang.NewTypedSyntaxList("SeparatedElement", separated),
	)
	projectToken := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindProjectKeyword, " ", "project"),
	)
	project := kustolang.MustNewProjectOperatorTyped(projectToken, expressions)
	if project.Kind() != kustolang.SyntaxKindProjectOperator || project.Expressions().Count() != 1 {
		t.Fatalf("typed project = %#v", project)
	}

	leftToken := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindLongLiteralToken, "", "1"))
	rightToken := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindLongLiteralToken, "", "2"))
	left := kustolang.MustNewLiteralExpressionTyped(kustolang.SyntaxKindLongLiteralExpression, leftToken)
	right := kustolang.MustNewLiteralExpressionTyped(kustolang.SyntaxKindLongLiteralExpression, rightToken)
	operator := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindPlusToken, " ", "+"))
	binary := kustolang.MustNewBinaryExpressionTyped(kustolang.SyntaxKindAddExpression, left, operator, right)
	if binary.Kind() != kustolang.SyntaxKindAddExpression || binary.Left() != left || binary.Right() != right {
		t.Fatalf("typed binary = %#v", binary)
	}

	nameToken := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "Account"))
	name := kustolang.MustNewTokenNameTyped(nameToken)
	reference := kustolang.MustNewNameReferenceTyped(name, kustolang.SymbolMatchColumn)
	if reference.Match() != kustolang.SymbolMatchColumn {
		t.Fatalf("name reference payload = %v", reference.Match())
	}
}

func TestPublicPinnedDiagnosticFactories(t *testing.T) {
	tests := []struct {
		diagnostic kustolang.Diagnostic
		code       string
		message    string
	}{
		{kustolang.DiagnosticMissingFunctionDeclaration(), "KS006", "Missing function declaration"},
		{kustolang.DiagnosticMissingAllLastOrNone(), "KS006", "Missing all, last or none"},
		{kustolang.DiagnosticSingleColumnTableExpected(), "KS113", "A tabular value with only one column expected."},
		{kustolang.DiagnosticScalarFunctionNotDefined("f"), "KS122", "The scalar function 'f' is not defined."},
		{kustolang.DiagnosticAggregateFunctionNotDefined("f"), "KS123", "The aggregate function 'f' is not defined."},
		{kustolang.DiagnosticNameIsNotAFunction("f"), "KS128", "The name 'f' does not refer to a function."},
		{kustolang.DiagnosticFunctionNotDefined("f"), "KS143", "The function 'f' is not defined."},
		{kustolang.DiagnosticTheExpressionHasNoName(), "KS148", "A column name cannot be inferred for this expression."},
		{kustolang.DiagnosticTheExpressionRefersToMoreThanOneColumn(), "KS184", "The expression refers to more than one column."},
		{kustolang.DiagnosticDefaultValueExpected(), "KS189", "Default value expected."},
		{kustolang.DiagnosticMaterializedViewNameMustBeStringLiteral(), "KS202", "Materialized view name must be a string literal"},
		{kustolang.DiagnosticRealLiteralExpected(), "KS225", "Real literal expected."},
		{kustolang.DiagnosticLatestSnapshotUnknown("g"), "KS262", "The lastest snapshot of graph model 'g' is unknown."},
		{kustolang.DiagnosticVolatileGraphUnknown("g"), "KS263", "The volatile graph of model 'g' is unknown."},
	}
	for _, test := range tests {
		if test.diagnostic.Code != test.code || test.diagnostic.Message != test.message {
			t.Errorf("diagnostic = (%q, %q), want (%q, %q)", test.diagnostic.Code, test.diagnostic.Message, test.code, test.message)
		}
	}
}
