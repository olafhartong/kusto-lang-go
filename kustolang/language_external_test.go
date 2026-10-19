package kustolang_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func TestPublicLanguageTypedSyntaxAndSemantics(t *testing.T) {
	account := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
	timestamp := kustolang.NewColumnSymbol("Timestamp", kustolang.ScalarTypes.DateTime.Type())
	table := kustolang.NewTableSymbol("Events", account, timestamp)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))
	text := `Events | where Account == "alice" | project Account`

	code := kustolang.ParseAndAnalyze(text, globals)
	if code == nil || !code.HasSemantics() || code.Kind != kustolang.CodeKindQuery {
		t.Fatalf("code = %#v", code)
	}
	if code.Syntax().Kind() != kustolang.SyntaxKindQueryBlock || code.Query == nil {
		t.Fatalf("syntax root = %v, query = %#v", code.Syntax().Kind(), code.Query)
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}

	result, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok || result.GetColumn("Account") == nil || result.GetColumn("Timestamp") != nil {
		t.Fatalf("result table = %#v", result)
	}
	row := code.GetColumnsInScope(strings.Index(text, "Account ==") + 1)
	if row.GetColumn("Account") != account || row.GetColumn("Timestamp") != timestamp {
		t.Fatalf("row scope = %#v", row.Columns())
	}

	var sawWhere bool
	var tableReference, columnReference kustolang.Symbol
	code.Syntax().WalkElements(func(element kustolang.SyntaxElement) {
		wrapped := kustolang.WrapGeneratedSyntaxElement(element)
		if _, ok := wrapped.(*kustolang.FilterOperator); ok {
			sawWhere = true
		}
		reference, ok := wrapped.(*kustolang.NameReference)
		if !ok {
			return
		}
		switch strings.TrimSpace(reference.Text(kustolang.IncludeTriviaAll)) {
		case "Events":
			tableReference = reference.ReferencedSymbol()
		case "Account":
			if reference.ReferencedSymbol() != nil {
				columnReference = reference.ReferencedSymbol()
			}
		}
	})
	if !sawWhere || tableReference != table || columnReference != account {
		t.Fatalf("typed/semantic walk = where:%t table:%#v column:%#v", sawWhere, tableReference, columnReference)
	}
}

func TestPublicLanguageEnvironmentAndMalformedRecovery(t *testing.T) {
	local := environment.LocalData{Tables: map[string]environment.TableDetails{
		"Events": {"Account": "string", "Timestamp": "datetime"},
	}}
	text := `Events | where Account ==`
	code, err := kustolang.ParseAndAnalyzeForEnvironment(text, nil, local)
	if err != nil {
		t.Fatal(err)
	}
	if code == nil || !code.HasSemantics() {
		t.Fatalf("malformed code = %#v", code)
	}
	if diagnostics := code.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS006") {
		t.Fatalf("malformed diagnostics = %#v", diagnostics)
	}
	if code.Syntax().GetLastToken(true) == nil || code.Syntax().GetLastToken(true).Kind() != kustolang.SyntaxKindEndOfTextToken {
		t.Fatalf("malformed tree did not retain end token")
	}
	row := code.GetColumnsInScope(len(text))
	if row.GetColumn("Account") == nil || row.GetColumn("Timestamp") == nil {
		t.Fatalf("malformed row scope = %#v", row.Columns())
	}

	var missingExpression bool
	code.Syntax().WalkElements(func(element kustolang.SyntaxElement) {
		if element.IsMissing() && element.HasSyntaxDiagnostics() {
			missingExpression = true
		}
	})
	if !missingExpression {
		t.Fatal("malformed syntax tree has no missing diagnostic element")
	}
}

func TestPublicLanguageLexerKindsScopeAndCancellation(t *testing.T) {
	tokens := kustolang.Lex(`Events | where Account == "alice"`)
	var sawWhere bool
	for _, token := range tokens {
		if token.Kind == kustolang.SyntaxKindWhereKeyword {
			sawWhere = true
		}
	}
	if !sawWhere || kustolang.SyntaxText(kustolang.SyntaxKindWhereKeyword) != "where" || !kustolang.IsKeywordKind(kustolang.SyntaxKindWhereKeyword) {
		t.Fatalf("public lexical catalog = %#v", tokens)
	}

	column := kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.Long.Type())
	table := kustolang.NewTableSymbol("T", column)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))
	text := `let x=1; T | where A > x`
	code := kustolang.ParseAndAnalyze(text, globals)
	position := len(text)
	if code.GetSpeculativeReferencedSymbol(position, "A", kustolang.SymbolMatchColumn) != column {
		t.Fatal("column scope did not retain source symbol identity")
	}
	local := code.GetSpeculativeReferencedSymbol(position, "x", kustolang.SymbolMatchLocal)
	if local == nil || local.SymbolName() != "x" {
		t.Fatalf("local scope = %#v", local)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := kustolang.ParseAndAnalyzeForEnvironmentContext(ctx, text, nil, environment.LocalData{})
	if result != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled parse = %#v, %v", result, err)
	}
}

func TestPublicLexicalTokenByteAndUTF16Lengths(t *testing.T) {
	token := kustolang.NewLexicalToken(kustolang.SyntaxKindStringLiteralToken, "😀 ", `"é"`)
	if token.Length() != token.ByteLength() {
		t.Fatalf("legacy length = %d, byte length = %d", token.Length(), token.ByteLength())
	}
	if token.UTF16Length() != 6 {
		t.Fatalf("UTF-16 length = %d, want 6", token.UTF16Length())
	}
	if token.ByteLength() == token.UTF16Length() {
		t.Fatal("Unicode fixture does not distinguish byte and UTF-16 lengths")
	}
}

func TestPublicLanguageParserFamiliesTraversalAndVisitor(t *testing.T) {
	expression := kustolang.ParseExpression(`1 + 2`)
	var visitedBinary bool
	kustolang.AcceptSyntaxVisitor(expression, kustolang.SyntaxVisitorFuncs{
		VisitBinaryExpressionFunc: func(node *kustolang.BinaryExpression) {
			visitedBinary = node.Kind() == kustolang.SyntaxKindAddExpression
		},
	})
	if !visitedBinary {
		t.Fatalf("typed visitor did not dispatch expression %T", expression)
	}

	tokens := kustolang.Lex(`ignored (value:long) trailing`)
	parameters := kustolang.ParseFunctionParametersTokensFromWithOptions(tokens, 1, kustolang.DefaultParseOptions)
	if parameters == nil || parameters.Kind() != kustolang.SyntaxKindFunctionParameters || parameters.HasSyntaxDiagnostics() {
		t.Fatalf("function parameters = %#v", parameters)
	}
	malformedBody := kustolang.ParseFunctionBody(`{ T | where A == }`)
	if malformedBody == nil || !kustolang.SyntaxHasMissingChildren(malformedBody) || len(malformedBody.GetContainedSyntaxDiagnostics()) == 0 {
		t.Fatalf("malformed function body = %#v", malformedBody)
	}

	query := kustolang.ParseQuery(`T | project A`)
	projects := kustolang.GetDescendants[*kustolang.ProjectOperator](query)
	if len(projects) != 1 || kustolang.SyntaxDepth(projects[0]) <= 0 {
		t.Fatalf("project descendants = %#v", projects)
	}
}

func TestPublicLanguageKustoCodeUsesUTF16Positions(t *testing.T) {
	column := kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.Long.Type())
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", kustolang.NewTableSymbol("T", column)))
	text := "// 😀\nT | where A >"
	code := kustolang.ParseAndAnalyze(text, globals)
	position := utf16Length(text)

	diagnostics := code.GetDiagnostics()
	if !hasDiagnostic(diagnostics, "KS006") || diagnostics[0].Start != position {
		t.Fatalf("UTF-16 diagnostics = %#v, want start %d", diagnostics, position)
	}
	if row := code.GetColumnsInScope(position); row == nil || row.GetColumn("A") != column {
		t.Fatalf("UTF-16 row scope = %#v", row)
	}
	if symbol := code.GetSpeculativeReferencedSymbol(position, "A", kustolang.SymbolMatchColumn); symbol != column {
		t.Fatalf("UTF-16 speculative symbol = %#v", symbol)
	}
	line, offset, ok := code.TryGetLineAndOffset(position)
	if !ok || line != 2 || offset != utf16Length("T | where A >")+1 {
		t.Fatalf("UTF-16 line/offset = %d/%d/%t", line, offset, ok)
	}
	tokenIndex := code.GetTokenIndex(utf16Length("// 😀\nT | whe"))
	tokens := code.GetLexicalTokens()
	if tokenIndex < 0 || tokenIndex >= len(tokens) || tokens[tokenIndex].Text != "where" {
		t.Fatalf("UTF-16 token index = %d, tokens %#v", tokenIndex, tokens)
	}

	if row := kustolang.GetRowScope(code.Tree, position, globals); row == nil || row.GetColumn("A") != column {
		t.Fatalf("UTF-16 binder facade row scope = %#v", row)
	}
}

func TestPublicLanguageOverloadedFunctionAndDiagnosticFactories(t *testing.T) {
	stringSignature := kustolang.NewSignature(
		kustolang.ScalarTypes.String.Type(),
		kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type()),
	)
	longSignature := kustolang.NewSignature(
		kustolang.ScalarTypes.Long.Type(),
		kustolang.NewParameter("value", kustolang.ScalarTypes.Long.Type()),
	)
	function := kustolang.NewFunctionSymbol("Pick", stringSignature, longSignature)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", function))
	code := kustolang.ParseAndAnalyze(`print result=Pick(1)`, globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("overload diagnostics = %#v", diagnostics)
	}
	calls := kustolang.GetDescendants[*kustolang.FunctionCallExpression](code.Syntax())
	if len(calls) != 1 || calls[0].ReferencedSignature() != longSignature || kustolang.ExpressionResultType(calls[0]) != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("resolved call = %#v", calls)
	}

	missing := kustolang.DiagnosticMissingExpression().WithLocation(4, 0)
	if missing.Code != "KS006" || missing.Message != "Missing expression" || missing.Start != 4 {
		t.Fatalf("diagnostic factory = %#v", missing)
	}
	custom := kustolang.NewDiagnosticDetailed("APP001", kustolang.DiagnosticCategoryCorrectness, kustolang.DiagnosticSeverityWarning, "description", "message")
	if custom.Code != "APP001" || custom.Category != kustolang.DiagnosticCategoryCorrectness || custom.Severity != kustolang.DiagnosticSeverityWarning {
		t.Fatalf("custom diagnostic = %#v", custom)
	}
}
