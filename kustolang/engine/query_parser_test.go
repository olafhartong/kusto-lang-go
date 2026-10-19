package kustolang

import (
	"reflect"
	"strings"
	"testing"
)

func parsedQueryExpression(t *testing.T, text string) Expression {
	t.Helper()
	return parsedQueryExpressionAt(t, text, 0, true)
}

func firstParsedQueryExpression(t *testing.T, text string) Expression {
	t.Helper()
	return parsedQueryExpressionAt(t, text, 0, false)
}

func parsedQueryExpressionAt(t *testing.T, text string, index int, requireSingleStatement bool) Expression {
	t.Helper()
	query := ParseQuery(text)
	return queryExpressionAt(t, query, index, requireSingleStatement)
}

func queryExpressionAt(t *testing.T, query *QueryBlock, index int, requireSingleStatement bool) Expression {
	t.Helper()
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || index < 0 || index >= statements.Count() || requireSingleStatement && statements.Count() != 1 {
		var got []string
		if statements != nil {
			for i := 0; i < statements.Count(); i++ {
				if separated, ok := statements.At(i).(*SeparatedElement); ok && separated.Element() != nil {
					got = append(got, separated.Element().Text(IncludeTriviaAll))
				}
			}
		}
		t.Fatalf("statements = %d %q, node = %#v", statements.Count(), got, query.GetChild(1))
	}
	separated, ok := statements.At(index).(*SeparatedElement)
	if !ok {
		t.Fatalf("statement element = %#v", statements.At(index))
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("statement = %#v", separated.Element())
	}
	return statement.Expression()
}

func namedParameterNameText(t *testing.T, parameter *NamedParameter) string {
	t.Helper()
	name, ok := parameter.Name().Name().(*TokenName)
	if !ok {
		t.Fatalf("parameter name = %#v", parameter.Name().Name())
	}
	return name.Name().RawText()
}

func diagnosticsContain(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func diagnosticsContainMessage(diagnostics []Diagnostic, code string, message string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Message == message {
			return true
		}
	}
	return false
}

func TestQueryRecoveryDoesNotDuplicateLeadingDirectives(t *testing.T) {
	// Matches the directive/statement separation in upstream QueryParser.ParseQuery.
	for _, query := range []string{"#\r&", "#\n&", "#\r\n&", "#one\n#two\n&", "#set x=1\nprint x=1 @"} {
		code := Parse(query)
		if got := code.Syntax().Text(IncludeTriviaAll); got != query {
			t.Fatalf("reconstructed %q as %q", query, got)
		}
		for _, diagnostic := range code.GetDiagnostics() {
			if diagnostic.Start < 0 || diagnostic.Start+diagnostic.Length > len(query) {
				t.Fatalf("%q has out-of-bounds diagnostic %#v", query, diagnostic)
			}
			if diagnostic.Code == "KS002" && strings.HasSuffix(query, "&") && (diagnostic.Start != len(query)-1 || diagnostic.Length != 1) {
				t.Fatalf("unexpected token location in %q: %#v", query, diagnostic)
			}
		}
	}
}

func tokenNameRawText(name Name) string {
	tokenName, ok := name.(*TokenName)
	if !ok || tokenName.Name() == nil {
		return ""
	}
	return tokenName.Name().RawText()
}

func nameDeclarationText(t *testing.T, declaration *NameDeclaration) string {
	t.Helper()
	name, ok := declaration.Name().(*TokenName)
	if !ok {
		t.Fatalf("name declaration = %#v", declaration.Name())
	}
	return name.Name().RawText()
}

func TestParseQuerySegmentsTopLevelPipes(t *testing.T) {
	query := ParseQuery("SigninLogs | where UserPrincipalName in ((_GetWatchlist('x') | project SPN)) | project UserPrincipalName")
	if query.Kind() != SyntaxKindQueryBlock {
		t.Fatalf("kind = %v", query.Kind())
	}
	if query.GeneratedNodeName() != "QueryBlock" {
		t.Fatalf("generated node name = %q", query.GeneratedNodeName())
	}
	if query.ChildCount() != 4 {
		t.Fatalf("child count = %d", query.ChildCount())
	}
	if query.GetName(0) != "Directives" || query.GetName(1) != "Statements" || query.GetName(2) != "SkippedTokens" || query.GetName(3) != "EndOfQuery" {
		t.Fatalf("query child names = %q, %q, %q, %q", query.GetName(0), query.GetName(1), query.GetName(2), query.GetName(3))
	}
	if query.IsOptional(0) || query.IsOptional(1) || !query.IsOptional(2) || !query.IsOptional(3) {
		t.Fatalf("query child optionality = %t, %t, %t, %t", query.IsOptional(0), query.IsOptional(1), query.IsOptional(2), query.IsOptional(3))
	}
	if query.GetChild(2) != nil {
		t.Fatalf("skipped tokens child = %#v", query.GetChild(2))
	}
	if len(query.Segments) != 3 {
		t.Fatalf("segments = %#v", query.Segments)
	}
	if query.Segments[0].Text != "SigninLogs" {
		t.Fatalf("segment 0 = %q", query.Segments[0].Text)
	}
	if query.Segments[1].Text != " where UserPrincipalName in ((_GetWatchlist('x') | project SPN))" {
		t.Fatalf("segment 1 = %q", query.Segments[1].Text)
	}
	if query.Segments[2].Text != " project UserPrincipalName" {
		t.Fatalf("segment 2 = %q", query.Segments[2].Text)
	}
}

func TestParseQueryConsumesLeadingDirectives(t *testing.T) {
	query := ParseQuery("#set query_now=datetime(2024-01-01)\nT | count")
	directives, ok := query.GetChild(0).(*SyntaxListOf[*Directive])
	if !ok || directives == nil || directives.Count() != 1 {
		t.Fatalf("directives = %#v", query.GetChild(0))
	}
	if directives.At(0).Token().RawText() != "#set query_now=datetime(2024-01-01)" {
		t.Fatalf("directive text = %q", directives.At(0).Token().RawText())
	}
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("leading directive diagnostics = %#v", diagnostics)
	}
}

func TestKustoCodeUsesQueryBlockRoot(t *testing.T) {
	code := Parse("T | take 1")
	if code.Query == nil {
		t.Fatal("missing query block")
	}
	if code.Syntax().Kind() != SyntaxKindQueryBlock {
		t.Fatalf("syntax kind = %v", code.Syntax().Kind())
	}
	if len(code.Query.Segments) != 2 {
		t.Fatalf("segments = %#v", code.Query.Segments)
	}
	if got := code.Syntax().Text(IncludeTriviaAll); got != "T | take 1" {
		t.Fatalf("syntax text = %q", got)
	}
}

func TestParseQueryBuildsLetMaterializeStatement(t *testing.T) {
	const text = "let Cached = materialize(SigninLogs | summarize Count=count() by UserPrincipalName); Cached | take 5"
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	letStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*LetStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if _, ok := letStatement.Expression().(*MaterializeExpression); !ok {
		t.Fatalf("let expression = %#v", letStatement.Expression())
	}
	materialized := letStatement.Expression().(*MaterializeExpression)
	if _, ok := materialized.Expression().(*PipeExpression); !ok {
		t.Fatalf("materialize inner expression = %#v", materialized.Expression())
	}
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("let separator = %#v", first.Separator())
	}

	second, ok := statements.At(1).(*SeparatedElement)
	if !ok {
		t.Fatalf("second statement element = %#v", statements.At(1))
	}
	expressionStatement, ok := WrapGeneratedSyntaxElement(second.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("second statement = %#v", second.Element())
	}
	if _, ok := expressionStatement.Expression().(*PipeExpression); !ok {
		t.Fatalf("query expression = %#v", expressionStatement.Expression())
	}
}

func TestParseQueryRejectsSourceMaterializeExpression(t *testing.T) {
	const text = "materialize(T | project A) | project A"
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("source materialize diagnostics = %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 1 {
		t.Fatalf("source materialize statements = %#v", query.GetChild(1))
	}
	separated, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("source materialize statement = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok || statement.Expression() != nil {
		t.Fatalf("source materialize expression = %#v", separated.Element())
	}
}

func TestParseQueryBuildsLetEntityGroupStatement(t *testing.T) {
	const text = "let Targets = entity_group[cluster('c1').database('db1'), database('db2')]; SigninLogs | take 1"
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	letStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*LetStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	entityGroup, ok := letStatement.Expression().(*EntityGroup)
	if !ok {
		t.Fatalf("let expression = %#v", letStatement.Expression())
	}
	if entityGroup.EntityGroupKeyword().Kind() != SyntaxKindEntityGroupKeyword {
		t.Fatalf("entity_group keyword = %v", entityGroup.EntityGroupKeyword().Kind())
	}
	if entityGroup.Entities().Count() != 2 {
		t.Fatalf("entity group entities = %d", entityGroup.Entities().Count())
	}
	if _, ok := entityGroup.Entities().At(0).Element().(*PathExpression); !ok {
		t.Fatalf("first entity = %#v", entityGroup.Entities().At(0).Element())
	}
}

func TestParseQueryBuildsLetFunctionDeclarationStatement(t *testing.T) {
	const text = `let FilterSignin = (User:string="admin") { declare query_parameters(App:string="Portal"); let Cached = materialize(SigninLogs | project UserPrincipalName, AppDisplayName); Cached | where UserPrincipalName == User and AppDisplayName == App | project UserPrincipalName }; SigninLogs | take 1`
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	letStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*LetStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	function, ok := letStatement.Expression().(*FunctionDeclaration)
	if !ok {
		t.Fatalf("let expression = %#v", letStatement.Expression())
	}
	if function.ViewKeyword() != nil {
		t.Fatalf("unexpected view keyword = %#v", function.ViewKeyword())
	}
	if function.Parameters().Parameters().Count() != 1 {
		t.Fatalf("function parameters = %d", function.Parameters().Parameters().Count())
	}
	parameter := function.Parameters().Parameters().At(0).Element()
	if got := nameDeclarationText(t, parameter.NameAndType().Name()); got != "User" {
		t.Fatalf("function parameter name = %q", got)
	}
	if parameter.DefaultValue() == nil || parameter.DefaultValue().Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("function parameter default = %#v", parameter.DefaultValue())
	}
	body := function.Body()
	if body.OpenBrace().Kind() != SyntaxKindOpenBraceToken || body.CloseBrace().Kind() != SyntaxKindCloseBraceToken {
		t.Fatalf("function body braces = %v/%v", body.OpenBrace().Kind(), body.CloseBrace().Kind())
	}
	if body.Statements().Count() != 2 {
		t.Fatalf("function body statements = %d", body.Statements().Count())
	}
	if _, ok := body.Statements().At(0).Element().(*QueryParametersStatement); !ok {
		t.Fatalf("function body statement 0 = %#v", body.Statements().At(0).Element())
	}
	innerLet, ok := body.Statements().At(1).Element().(*LetStatement)
	if !ok {
		t.Fatalf("function body statement 1 = %#v", body.Statements().At(1).Element())
	}
	if _, ok := innerLet.Expression().(*MaterializeExpression); !ok {
		t.Fatalf("function body materialize expression = %#v", innerLet.Expression())
	}
	if _, ok := body.Expression().(*PipeExpression); !ok {
		t.Fatalf("function body expression = %#v", body.Expression())
	}
	if body.Semicolon() != nil {
		t.Fatalf("unexpected function body semicolon = %#v", body.Semicolon())
	}
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("function let separator = %#v", first.Separator())
	}
}

func TestParseFunctionBodyPublicFacade(t *testing.T) {
	const text = `{ declare query_parameters(App:string="Portal"); let Cached = materialize(SigninLogs | project UserPrincipalName); Cached | where App == "Portal" }`
	body := ParseFunctionBody(text)
	if body == nil {
		t.Fatal("expected function body")
	}
	if body.OpenBrace().Kind() != SyntaxKindOpenBraceToken || body.CloseBrace().Kind() != SyntaxKindCloseBraceToken {
		t.Fatalf("function body braces = %v/%v", body.OpenBrace().Kind(), body.CloseBrace().Kind())
	}
	if body.Statements().Count() != 2 {
		t.Fatalf("function body statements = %d", body.Statements().Count())
	}
	if _, ok := body.Statements().At(0).Element().(*QueryParametersStatement); !ok {
		t.Fatalf("function body statement 0 = %#v", body.Statements().At(0).Element())
	}
	if _, ok := body.Statements().At(1).Element().(*LetStatement); !ok {
		t.Fatalf("function body statement 1 = %#v", body.Statements().At(1).Element())
	}
	if _, ok := body.Expression().(*PipeExpression); !ok {
		t.Fatalf("function body expression = %#v", body.Expression())
	}
	if body.Semicolon() != nil {
		t.Fatalf("unexpected function body semicolon = %#v", body.Semicolon())
	}
	if diagnostics := body.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("function body diagnostics = %#v", diagnostics)
	}
}

func TestParseFunctionBodyPublicFacadeSemicolonRecoveryAndNegativeInput(t *testing.T) {
	withSemicolon := ParseFunctionBody(`{ let x = 1; x; }`)
	if withSemicolon == nil {
		t.Fatal("expected function body with final semicolon")
	}
	if withSemicolon.Semicolon() == nil || withSemicolon.Semicolon().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("function body semicolon = %#v", withSemicolon.Semicolon())
	}
	if diagnostics := withSemicolon.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("function body with semicolon diagnostics = %#v", diagnostics)
	}

	recovered := ParseFunctionBody(`{ let x = ; x }`)
	if recovered == nil {
		t.Fatal("expected recovered function body")
	}
	if !diagnosticsContainMessage(recovered.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("expected missing-expression diagnostic, got %#v", recovered.ContainedDiagnostics())
	}

	if got := ParseFunctionBody(`let x = 1; x`); got != nil {
		t.Fatalf("function body without brace = %#v, want nil", got)
	}
}

func TestParseQueryBuildsMetadataStatements(t *testing.T) {
	const text = `alias database Sales = cluster('c').database('db'); set query_take_max_records=1000; declare query_parameters(User:string="admin", Threshold:long=5); SigninLogs | where UserPrincipalName == User | take Threshold`
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 4 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}

	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	aliasStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*AliasStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if aliasStatement.AliasKeyword().Kind() != SyntaxKindAliasKeyword || aliasStatement.DatabaseKeyword().Kind() != SyntaxKindDatabaseKeyword {
		t.Fatalf("alias keywords = %v/%v", aliasStatement.AliasKeyword().Kind(), aliasStatement.DatabaseKeyword().Kind())
	}
	if got := nameDeclarationText(t, aliasStatement.Name()); got != "Sales" {
		t.Fatalf("alias name = %q", got)
	}
	if _, ok := aliasStatement.Expression().(*PathExpression); !ok {
		t.Fatalf("alias expression = %#v", aliasStatement.Expression())
	}

	second, ok := statements.At(1).(*SeparatedElement)
	if !ok {
		t.Fatalf("second statement element = %#v", statements.At(1))
	}
	setStatement, ok := WrapGeneratedSyntaxElement(second.Element()).(*SetOptionStatement)
	if !ok {
		t.Fatalf("second statement = %#v", second.Element())
	}
	if got := nameDeclarationText(t, setStatement.Name()); got != "query_take_max_records" {
		t.Fatalf("set option name = %q", got)
	}
	if setStatement.ValueClause() == nil || setStatement.ValueClause().Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("set option value = %#v", setStatement.ValueClause())
	}

	third, ok := statements.At(2).(*SeparatedElement)
	if !ok {
		t.Fatalf("third statement element = %#v", statements.At(2))
	}
	parametersStatement, ok := WrapGeneratedSyntaxElement(third.Element()).(*QueryParametersStatement)
	if !ok {
		t.Fatalf("third statement = %#v", third.Element())
	}
	if parametersStatement.DeclareKeyword().Kind() != SyntaxKindDeclareKeyword || parametersStatement.QueryParametersKeyword().Kind() != SyntaxKindQueryParametersKeyword {
		t.Fatalf("query-parameters keywords = %v/%v", parametersStatement.DeclareKeyword().Kind(), parametersStatement.QueryParametersKeyword().Kind())
	}
	if parametersStatement.Parameters().Count() != 2 {
		t.Fatalf("query parameters = %d", parametersStatement.Parameters().Count())
	}
	firstParameter := parametersStatement.Parameters().At(0).Element()
	if got := nameDeclarationText(t, firstParameter.NameAndType().Name()); got != "User" {
		t.Fatalf("first query parameter name = %q", got)
	}
	if firstParameter.DefaultValue() == nil || firstParameter.DefaultValue().Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("first query parameter default = %#v", firstParameter.DefaultValue())
	}
	secondParameter := parametersStatement.Parameters().At(1).Element()
	if got := nameDeclarationText(t, secondParameter.NameAndType().Name()); got != "Threshold" {
		t.Fatalf("second query parameter name = %q", got)
	}
	if secondParameter.DefaultValue() == nil || secondParameter.DefaultValue().Value().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("second query parameter default = %#v", secondParameter.DefaultValue())
	}
}

func TestParseQueryParametersTrailingCommaReportsMissingParameter(t *testing.T) {
	query := ParseQuery(`declare query_parameters(x:string,); print x`)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing parameter") {
		t.Fatalf("expected missing parameter diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("trailing query parameter comma should not add missing-token diagnostics, got %#v", diagnostics)
	}
}

func TestParseQueryBuildsPatternStatement(t *testing.T) {
	const text = `declare pattern LoginPattern = (User:string, App:string) [Path:string] { ("admin", "portal").["signin"] = { let Label = User; print Value=Label }; }; SigninLogs | take 1`
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	patternStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*PatternStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if patternStatement.DeclareKeyword().Kind() != SyntaxKindDeclareKeyword || patternStatement.PatternKeyword().Kind() != SyntaxKindPatternKeyword {
		t.Fatalf("pattern statement keywords = %v/%v", patternStatement.DeclareKeyword().Kind(), patternStatement.PatternKeyword().Kind())
	}
	if got := nameDeclarationText(t, patternStatement.Name()); got != "LoginPattern" {
		t.Fatalf("pattern name = %q", got)
	}
	declaration := patternStatement.Pattern()
	if declaration == nil {
		t.Fatal("missing pattern declaration")
	}
	if declaration.Parameters().Count() != 2 {
		t.Fatalf("pattern parameters = %d", declaration.Parameters().Count())
	}
	if got := nameDeclarationText(t, declaration.Parameters().At(0).Element().Name()); got != "User" {
		t.Fatalf("first pattern parameter = %q", got)
	}
	if declaration.PathParameter() == nil || nameDeclarationText(t, declaration.PathParameter().Parameter().Name()) != "Path" {
		t.Fatalf("pattern path parameter = %#v", declaration.PathParameter())
	}
	if declaration.Patterns().Count() != 1 {
		t.Fatalf("pattern matches = %d", declaration.Patterns().Count())
	}
	match := declaration.Patterns().At(0)
	if match.ParameterValues().Expressions().Count() != 2 {
		t.Fatalf("match parameter values = %d", match.ParameterValues().Expressions().Count())
	}
	if match.PathValue() == nil || match.PathValue().Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("match path value = %#v", match.PathValue())
	}
	if match.Body().Statements().Count() != 1 {
		t.Fatalf("match body statements = %d", match.Body().Statements().Count())
	}
	if _, ok := match.Body().Statements().At(0).Element().(*LetStatement); !ok {
		t.Fatalf("match body statement = %#v", match.Body().Statements().At(0).Element())
	}
	if _, ok := match.Body().Expression().(*PrintOperator); !ok {
		t.Fatalf("match body expression = %#v", match.Body().Expression())
	}
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("pattern statement separator = %#v", first.Separator())
	}
}

func TestParsePatternStatementMissingMatchDiagnostic(t *testing.T) {
	query := ParseQuery(`declare pattern Empty = (User:string) { }; print x=1`)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing pattern match clause") {
		t.Fatalf("expected missing pattern match diagnostic, got %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first := statements.At(0).(*SeparatedElement)
	patternStatement, ok := WrapGeneratedSyntaxElement(first.Element()).(*PatternStatement)
	if !ok || patternStatement.Pattern() == nil || patternStatement.Pattern().Patterns().Count() != 1 {
		t.Fatalf("pattern statement = %#v", first.Element())
	}
}

func TestParseQueryBuildsRestrictStatement(t *testing.T) {
	const text = `restrict access to (cluster('c').database('db').Signin*, Device*) with (AllowManagedIdentityAuthentication=true); SigninLogs | take 1`
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	restrict, ok := WrapGeneratedSyntaxElement(first.Element()).(*RestrictStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if restrict.RestrictKeyword().Kind() != SyntaxKindRestrictKeyword || restrict.AccessKeyword().Kind() != SyntaxKindAccessKeyword || restrict.ToKeyword().Kind() != SyntaxKindToKeyword {
		t.Fatalf("restrict keywords = %v/%v/%v", restrict.RestrictKeyword().Kind(), restrict.AccessKeyword().Kind(), restrict.ToKeyword().Kind())
	}
	if restrict.Expressions().Count() != 2 {
		t.Fatalf("restrict expressions = %d", restrict.Expressions().Count())
	}
	if _, ok := restrict.Expressions().At(0).Element().(*PathExpression); !ok {
		t.Fatalf("restrict first expression = %#v", restrict.Expressions().At(0).Element())
	}
	wildcard, ok := restrict.Expressions().At(1).Element().(*NameReference)
	if !ok {
		t.Fatalf("restrict second expression = %#v", restrict.Expressions().At(1).Element())
	}
	if wildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("restrict wildcard kind = %v", wildcard.Name().Kind())
	}
	if restrict.WithClause() == nil || restrict.WithClause().Properties().Count() != 1 {
		t.Fatalf("restrict with clause = %#v", restrict.WithClause())
	}
	property := restrict.WithClause().Properties().At(0).Element()
	if got := namedParameterNameText(t, property); got != "AllowManagedIdentityAuthentication" {
		t.Fatalf("restrict property = %q", got)
	}
	if property.Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("restrict property value = %v", property.Expression().Kind())
	}
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("restrict separator = %#v", first.Separator())
	}
}

func TestParseRestrictStatementMalformedCommaListRecovers(t *testing.T) {
	query := ParseQuery(`restrict access to (,); T | count`)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing-expression diagnostic, got %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	restrict, ok := WrapGeneratedSyntaxElement(first.Element()).(*RestrictStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if restrict.Expressions().Count() != 2 {
		t.Fatalf("restrict expressions = %d", restrict.Expressions().Count())
	}
	for i, element := range restrict.Expressions().Elements() {
		if element.Element() == nil {
			t.Fatalf("restrict expression %d is nil", i)
		}
	}
}

func TestParseRestrictStatementMalformedListDiagnostics(t *testing.T) {
	tests := []struct {
		query string
		codes []string
	}{
		{query: `restrict access to (); T | count`, codes: []string{"KS006"}},
		{query: `restrict access to (,); T | count`, codes: []string{"KS006", "KS006"}},
		{query: `restrict access to (T,); T | count`, codes: []string{"KS006"}},
		{query: `restrict access to (T U); T | count`, codes: []string{"KS005"}},
		{query: `restrict access to (T,,U); T | count`, codes: []string{"KS006", "KS005", "KS198"}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			diagnostics := ParseQuery(tt.query).ContainedDiagnostics()
			if len(diagnostics) != len(tt.codes) {
				t.Fatalf("diagnostics = %#v, want codes %#v", diagnostics, tt.codes)
			}
			for i, code := range tt.codes {
				if diagnostics[i].Code != code {
					t.Fatalf("diagnostic %d = %#v, want %s", i, diagnostics[i], code)
				}
			}
		})
	}
}

func TestParseQueryBuildsMacroExpandOperator(t *testing.T) {
	const text = `macro-expand force_remote=true entity_group[cluster('c').database('db')] as Scope (SigninLogs | project UserPrincipalName)`
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected macro-expand diagnostics = %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, text)
	macro, ok := expr.(*MacroExpandOperator)
	if !ok {
		t.Fatalf("macro-expand expression = %#v", expr)
	}
	if macro.MacroExpandKeyword().Kind() != SyntaxKindMacroExpandKeyword {
		t.Fatalf("macro-expand keyword = %v", macro.MacroExpandKeyword().Kind())
	}
	if macro.Parameters().Count() != 1 {
		t.Fatalf("macro-expand parameters = %d", macro.Parameters().Count())
	}
	if got := namedParameterNameText(t, macro.Parameters().At(0)); got != "force_remote" {
		t.Fatalf("macro-expand parameter = %q", got)
	}
	if _, ok := macro.EntityGroup().(*EntityGroup); !ok {
		t.Fatalf("macro-expand entity group = %#v", macro.EntityGroup())
	}
	scope := macro.ScopeReferenceName()
	if scope == nil || scope.AsKeyword().Kind() != SyntaxKindAsKeyword {
		t.Fatalf("macro-expand scope = %#v", scope)
	}
	scopeName, ok := scope.EntityGroupReferenceName().Name().(*TokenName)
	if !ok || scopeName.Name().RawText() != "Scope" {
		t.Fatalf("macro-expand scope name = %#v", scope.EntityGroupReferenceName().Name())
	}
	if macro.StatementList().Count() != 1 {
		t.Fatalf("macro-expand statement list = %d", macro.StatementList().Count())
	}
	statement, ok := macro.StatementList().At(0).Element().(*ExpressionStatement)
	if !ok {
		t.Fatalf("macro-expand statement = %#v", macro.StatementList().At(0).Element())
	}
	if _, ok := statement.Expression().(*PipeExpression); !ok {
		t.Fatalf("macro-expand statement expression = %#v", statement.Expression())
	}
}

func TestParseQueryBuildsPipedMacroExpandOperator(t *testing.T) {
	const text = `SigninLogs | macro-expand entity_group("prod") (print UserPrincipalName="x")`
	expr := parsedQueryExpression(t, text)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %#v", expr)
	}
	macro, ok := pipe.Operator().(*MacroExpandOperator)
	if !ok {
		t.Fatalf("macro-expand operator = %#v", pipe.Operator())
	}
	if _, ok := macro.EntityGroup().(*FunctionCallExpression); !ok {
		t.Fatalf("macro-expand entity group reference = %#v", macro.EntityGroup())
	}
	if macro.ScopeReferenceName() != nil {
		t.Fatalf("unexpected macro-expand scope = %#v", macro.ScopeReferenceName())
	}
	if macro.StatementList().Count() != 1 {
		t.Fatalf("macro-expand statement list = %d", macro.StatementList().Count())
	}
	statement, ok := macro.StatementList().At(0).Element().(*ExpressionStatement)
	if !ok {
		t.Fatalf("macro-expand statement = %#v", macro.StatementList().At(0).Element())
	}
	if _, ok := statement.Expression().(*PrintOperator); !ok {
		t.Fatalf("macro-expand body expression = %#v", statement.Expression())
	}
}

func TestParseQueryBuildsMacroExpandQualifiedEntityPathTarget(t *testing.T) {
	const text = `macro-expand cluster("c").database("d").Targets as Scope (print x=1)`
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected macro-expand qualified path diagnostics = %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, text)
	macro, ok := expr.(*MacroExpandOperator)
	if !ok {
		t.Fatalf("macro-expand expression = %#v", expr)
	}
	path, ok := macro.EntityGroup().(*PathExpression)
	if !ok {
		t.Fatalf("macro-expand qualified path target = %#v", macro.EntityGroup())
	}
	target, ok := path.Selector().(*NameReference)
	if !ok || tokenNameRawText(target.Name()) != "Targets" {
		t.Fatalf("macro-expand qualified path selector = %#v", path.Selector())
	}
}

func TestParseMacroExpandRejectsLiteralEntityGroupReference(t *testing.T) {
	query := ParseQuery(`macro-expand 1 (print x=1)`)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("expected missing expression diagnostic, got %#v", query.ContainedDiagnostics())
	}
}

func TestParseMacroExpandCommaTargetTailReportsIncompleteFragment(t *testing.T) {
	query := ParseQuery(`macro-expand entity_group("prod"), Extra (print x=1)`)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected macro-expand target comma-tail diagnostic, got %#v", query.ContainedDiagnostics())
	}
	expr := parsedQueryExpression(t, `macro-expand entity_group("prod"), Extra (print x=1)`)
	macro, ok := expr.(*MacroExpandOperator)
	if !ok {
		t.Fatalf("macro-expand expression = %#v", expr)
	}
	if macro.StatementList() == nil || macro.StatementList().Count() != 0 {
		t.Fatalf("macro-expand statement list = %#v", macro.StatementList())
	}
}

func TestParseQueryBuildsPipedOperatorTree(t *testing.T) {
	const text = `SigninLogs | where UserPrincipalName has "admin" | project UserPrincipalName, Score = 1 + 2 | take 5 | count as Total`
	query := ParseQuery(text)
	if got := query.Text(IncludeTriviaAll); got != text {
		t.Fatalf("query text = %q", got)
	}
	expr := parsedQueryExpression(t, text)

	countPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("count pipe = %#v", expr)
	}
	count, ok := countPipe.Operator().(*CountOperator)
	if !ok {
		t.Fatalf("count operator = %#v", countPipe.Operator())
	}
	if count.AsIdentifier() == nil || count.AsIdentifier().Identifier().RawText() != "Total" {
		t.Fatalf("count as clause = %#v", count.AsIdentifier())
	}

	takePipe, ok := countPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("take pipe = %#v", countPipe.Expression())
	}
	take, ok := takePipe.Operator().(*TakeOperator)
	if !ok {
		t.Fatalf("take operator = %#v", takePipe.Operator())
	}
	if take.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("take expression kind = %v", take.Expression().Kind())
	}

	projectPipe, ok := takePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", takePipe.Expression())
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 2 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}
	if project.Expressions().At(1).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("project named expression kind = %v", project.Expressions().At(1).Element().Kind())
	}

	filterPipe, ok := projectPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("filter pipe = %#v", projectPipe.Expression())
	}
	filter, ok := filterPipe.Operator().(*FilterOperator)
	if !ok {
		t.Fatalf("filter operator = %#v", filterPipe.Operator())
	}
	if filter.Condition().Kind() != SyntaxKindHasExpression {
		t.Fatalf("filter condition kind = %v", filter.Condition().Kind())
	}
	if filterPipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("source expression kind = %v", filterPipe.Expression().Kind())
	}
}

func TestParseCountAsInvalidAliasUsesMissingIdentifierToken(t *testing.T) {
	const text = `T | count as 123`
	query := ParseQuery(text)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("expected missing identifier diagnostic, got %#v", query.ContainedDiagnostics())
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() < 1 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	separated, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("statement element = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("statement = %#v", separated.Element())
	}
	expr := statement.Expression()
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %#v", expr)
	}
	count, ok := pipe.Operator().(*CountOperator)
	if !ok {
		t.Fatalf("count operator = %#v", pipe.Operator())
	}
	if count.AsIdentifier() == nil {
		t.Fatalf("missing count as clause")
	}
	identifier := count.AsIdentifier().Identifier()
	if identifier == nil || identifier.Kind() != SyntaxKindIdentifierToken || identifier.RawText() != "" || identifier.ValueText() != "" {
		t.Fatalf("count alias identifier = kind:%v raw:%q value:%q", identifier.Kind(), identifier.RawText(), identifier.ValueText())
	}
}

func TestParseCountAsCommaTailKeepsAliasToken(t *testing.T) {
	for _, text := range []string{
		`T | count as Total Extra`,
		`T | count as Total, Extra`,
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			expr := parsedQueryExpressionAt(t, text, 0, false)
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("pipe expression = %#v", expr)
			}
			count, ok := pipe.Operator().(*CountOperator)
			if !ok {
				t.Fatalf("count operator = %#v", pipe.Operator())
			}
			if count.AsIdentifier() == nil || count.AsIdentifier().Identifier() == nil {
				t.Fatalf("missing count as clause")
			}
			if got := count.AsIdentifier().Identifier().RawText(); got != "Total" {
				t.Fatalf("count alias raw text = %q", got)
			}
			if strings.Contains(text, ",") {
				if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
					t.Fatalf("expected comma skipped-tail diagnostic, got %#v", query.ContainedDiagnostics())
				}
			} else if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: ;") || diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
				t.Fatalf("expected whitespace statement-tail diagnostic, got %#v", query.ContainedDiagnostics())
			}
		})
	}
}

func TestParseQueryBuildsExecuteAndCacheOperator(t *testing.T) {
	const text = "T | __executeAndCache | take 1"
	expr := parsedQueryExpression(t, text)

	takePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("take pipe = %#v", expr)
	}
	executePipe, ok := takePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("execute pipe = %#v", takePipe.Expression())
	}
	execute, ok := executePipe.Operator().(*ExecuteAndCacheOperator)
	if !ok {
		t.Fatalf("execute-and-cache operator = %#v", executePipe.Operator())
	}
	if execute.ExecuteAndCacheKeyword().Kind() != SyntaxKindExecuteAndCacheKeyword {
		t.Fatalf("execute-and-cache keyword = %v", execute.ExecuteAndCacheKeyword().Kind())
	}
}

func TestParseQueryBuildsAssertSchemaOperator(t *testing.T) {
	const text = "T | assert-schema (UserPrincipalName:string, Count:long)"
	expr := parsedQueryExpression(t, text)

	assertPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("assert-schema pipe = %#v", expr)
	}
	assertSchema, ok := assertPipe.Operator().(*AssertSchemaOperator)
	if !ok {
		t.Fatalf("assert-schema operator = %#v", assertPipe.Operator())
	}
	if assertSchema.AssertSchemaKeyword().Kind() != SyntaxKindAssertSchemaKeyword {
		t.Fatalf("assert-schema keyword = %v", assertSchema.AssertSchemaKeyword().Kind())
	}
	if assertSchema.Schema() == nil {
		t.Fatal("missing assert-schema row schema")
	}
	if assertSchema.Schema().Columns().Count() != 2 {
		t.Fatalf("assert-schema columns = %d", assertSchema.Schema().Columns().Count())
	}
	first := assertSchema.Schema().Columns().At(0).Element()
	if first.Name().Name().Kind() != SyntaxKindTokenName || first.Type().Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("assert-schema first column = %#v", first)
	}

	commaTail := ParseQuery("T | assert-schema (UserPrincipalName:string), Extra")
	if !diagnosticsContain(commaTail.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected assert-schema comma-tail diagnostic, got %#v", commaTail.ContainedDiagnostics())
	}
}

func TestParseAssertSchemaUnparenthesizedTailReportsSkippedTokens(t *testing.T) {
	const text = "T | assert-schema A:string B:long"
	query := ParseQuery(text)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("expected missing schema diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected skipped-token diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if query.GetChild(2) == nil {
		t.Fatal("unparenthesized assert-schema tail should remain in QueryBlock skipped-token recovery")
	}
	expr := firstParsedQueryExpression(t, text)
	assertPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("assert-schema pipe = %#v", expr)
	}
	assertSchema, ok := assertPipe.Operator().(*AssertSchemaOperator)
	if !ok {
		t.Fatalf("assert-schema operator = %#v", assertPipe.Operator())
	}
	if assertSchema.Schema() == nil || !diagnosticsContainMessage(assertSchema.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("assert-schema schema diagnostics = %#v", assertSchema.Schema())
	}
}

func TestParseFindOperandAliasesMirrorUpstreamEntitySelectors(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantPipe bool
	}{
		{
			name:     "bare name alias",
			text:     `find in (SigninLogs | as Signins) where true project UserPrincipalName`,
			wantPipe: true,
		},
		{
			name:     "bracketed wildcard alias",
			text:     `find in ([Signin*] | as Signins) where true project UserPrincipalName`,
			wantPipe: true,
		},
		{
			name:     "wildcard without alias",
			text:     `find in (Signin*) where true project UserPrincipalName`,
			wantPipe: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}

			find, ok := parsedQueryExpression(t, tt.text).(*FindOperator)
			if !ok {
				t.Fatalf("find operator = %#v", parsedQueryExpression(t, tt.text))
			}
			first := find.InClause().Expressions().At(0).Element()
			if gotPipe := first.Kind() == SyntaxKindPipeExpression; gotPipe != tt.wantPipe {
				t.Fatalf("first operand pipe = %t, want %t (%#v)", gotPipe, tt.wantPipe, first)
			}
		})
	}
}

func TestParseFindRejectsAliasesAfterWildcardAndFullEntityOperands(t *testing.T) {
	tests := []string{
		`find in (Signin* | as Signins) where true project UserPrincipalName`,
		`find in (database("db").SigninLogs | as Signins) where true project UserPrincipalName`,
		`find in (SomeTabularFn() | as F) where true project UserPrincipalName`,
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			code := Parse(text)
			diagnostics := code.GetDiagnostics()
			if !diagnosticsContain(diagnostics, "KS198") {
				t.Fatalf("expected skipped alias-tail diagnostic, got %#v", diagnostics)
			}
		})
	}
}

func TestParseQueryBuildsPartitionOperator(t *testing.T) {
	const text = "SigninLogs | partition hint.strategy=native by UserPrincipalName (summarize Count=count() by AppDisplayName)"
	expr := parsedQueryExpression(t, text)

	partitionPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("partition pipe = %#v", expr)
	}
	partition, ok := partitionPipe.Operator().(*PartitionOperator)
	if !ok {
		t.Fatalf("partition operator = %#v", partitionPipe.Operator())
	}
	if partition.Parameters().Count() != 1 {
		t.Fatalf("partition parameters = %d", partition.Parameters().Count())
	}
	if partition.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("partition by keyword = %v", partition.ByKeyword().Kind())
	}
	if partition.ByExpression().Kind() != SyntaxKindNameReference {
		t.Fatalf("partition by expression kind = %v", partition.ByExpression().Kind())
	}
	subquery, ok := partition.Operand().(*PartitionSubquery)
	if !ok {
		t.Fatalf("partition operand = %#v", partition.Operand())
	}
	if subquery.Subquery().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("partition subquery kind = %v", subquery.Subquery().Kind())
	}
}

func TestParseQueryBuildsPartitionOperatorWithBracketedByPath(t *testing.T) {
	const text = `T | partition by Payload["tenant"] (summarize Count=count())`
	expr := parsedQueryExpression(t, text)

	partitionPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("partition pipe = %#v", expr)
	}
	partition, ok := partitionPipe.Operator().(*PartitionOperator)
	if !ok {
		t.Fatalf("partition operator = %#v", partitionPipe.Operator())
	}
	element, ok := partition.ByExpression().(*ElementExpression)
	if !ok {
		t.Fatalf("partition by expression = %#v", partition.ByExpression())
	}
	if element.Expression().Kind() != SyntaxKindNameReference || element.Selector().Kind() != SyntaxKindNameReference {
		t.Fatalf("partition by element = %v / %v", element.Expression().Kind(), element.Selector().Kind())
	}
	if diagnostics := partition.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("partition bracketed path diagnostics = %#v", diagnostics)
	}
}

func TestParseQueryBuildsPartitionOperatorWithDynamicScope(t *testing.T) {
	const text = "MyTable | partition by Account in dynamic([1,2]) (summarize Count=count() by Value)"
	expr := parsedQueryExpression(t, text)

	partitionPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("partition pipe = %#v", expr)
	}
	partition, ok := partitionPipe.Operator().(*PartitionOperator)
	if !ok {
		t.Fatalf("partition operator = %#v", partitionPipe.Operator())
	}
	if partition.Scope() == nil {
		t.Fatal("partition scope is nil")
	}
	if partition.Scope().Expression().Kind() != SyntaxKindDynamicExpression {
		t.Fatalf("partition scope expression kind = %v", partition.Scope().Expression().Kind())
	}
	if diagnostics := partition.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("partition dynamic scope diagnostics = %#v", diagnostics)
	}
}

func TestParsePartitionReportsUnexpectedByExpressionTail(t *testing.T) {
	for _, test := range []struct {
		query     string
		wantKS198 bool
	}{
		{`T | partition by X Z (project X) | project X`, false},
		{`T | partition by X, Z (project X) | project X`, true},
		{`T | partition by X in dynamic([1]) Z (project X) | project X`, false},
		{`T | partition by X, Z in dynamic([1]) (project X) | project X`, true},
		{`T | partition by X (project X), Z | project X`, true},
	} {
		code := Parse(test.query)
		if len(code.GetDiagnostics()) == 0 {
			t.Fatalf("malformed partition query %q produced no diagnostics", test.query)
		}
		if test.wantKS198 && !diagnosticsContain(code.GetDiagnostics(), "KS198") {
			t.Fatalf("expected QueryBlock skipped-token diagnostic for %q, got %#v", test.query, code.GetDiagnostics())
		}
	}
}

func TestParseQueryBuildsPartitionByOperator(t *testing.T) {
	const text = "SigninLogs | __partitionby hint.strategy=native UserPrincipalName id 'slice' (summarize Count=count() by AppDisplayName)"
	expr := parsedQueryExpression(t, text)

	partitionPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("partition-by pipe = %#v", expr)
	}
	partitionBy, ok := partitionPipe.Operator().(*PartitionByOperator)
	if !ok {
		t.Fatalf("partition-by operator = %#v", partitionPipe.Operator())
	}
	if partitionBy.Parameters().Count() != 1 {
		t.Fatalf("partition-by parameters = %d", partitionBy.Parameters().Count())
	}
	if got := namedParameterNameText(t, partitionBy.Parameters().At(0)); got != "hint.strategy" {
		t.Fatalf("partition-by parameter = %q", got)
	}
	if partitionBy.Entity().Kind() != SyntaxKindNameReference {
		t.Fatalf("partition-by entity kind = %v", partitionBy.Entity().Kind())
	}
	if partitionBy.IdClause() == nil || partitionBy.IdClause().Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("partition-by id clause = %#v", partitionBy.IdClause())
	}
	if partitionBy.Subquery().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("partition-by subquery kind = %v", partitionBy.Subquery().Kind())
	}
}

func TestParseQueryBuildsPartitionBySourcePipeSubquery(t *testing.T) {
	const text = "SigninLogs | __partitionby UserPrincipalName (Other | count)"
	expr := parsedQueryExpression(t, text)

	partitionPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("partition-by pipe = %#v", expr)
	}
	partitionBy, ok := partitionPipe.Operator().(*PartitionByOperator)
	if !ok {
		t.Fatalf("partition-by operator = %#v", partitionPipe.Operator())
	}
	if _, ok := partitionBy.Subquery().(*PipeExpression); !ok {
		t.Fatalf("partition-by source-pipe subquery = %#v", partitionBy.Subquery())
	}
	if diagnostics := partitionBy.ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS176") {
		t.Fatalf("partition-by source-pipe should require an operator root, got %#v", diagnostics)
	}
}

func TestParsePartitionByReportsUnexpectedEntityExpressionTail(t *testing.T) {
	for _, test := range []struct {
		query     string
		wantKS198 bool
	}{
		{`T | __partitionby X Z (project X) | project X`, false},
		{`T | __partitionby X, Z (project X) | project X`, true},
		{`T | __partitionby X id "slice" Z (project X) | project X`, false},
		{`T | __partitionby X id "slice", Z (project X) | project X`, true},
	} {
		code := Parse(test.query)
		if len(code.GetDiagnostics()) == 0 {
			t.Fatalf("malformed partition-by query %q produced no diagnostics", test.query)
		}
		if test.wantKS198 && !diagnosticsContain(code.GetDiagnostics(), "KS198") {
			t.Fatalf("expected QueryBlock skipped-token diagnostic for %q, got %#v", test.query, code.GetDiagnostics())
		}
	}
}

func TestParseQueryRejectsPartitionByBareSourceSubquery(t *testing.T) {
	code := Parse("SigninLogs | __partitionby UserPrincipalName (Other)")
	if !diagnosticsContain(code.GetDiagnostics(), "KS176") {
		t.Fatalf("expected partition-by bare source diagnostic, got %#v", code.GetDiagnostics())
	}
}

func TestParseQueryBuildsForkOperator(t *testing.T) {
	const text = "SigninLogs | fork (where UserPrincipalName == 'a') ByApp=(summarize Count=count() by AppDisplayName)"
	expr := parsedQueryExpression(t, text)

	forkPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("fork pipe = %#v", expr)
	}
	fork, ok := forkPipe.Operator().(*ForkOperator)
	if !ok {
		t.Fatalf("fork operator = %#v", forkPipe.Operator())
	}
	if fork.Expressions().Count() != 2 {
		t.Fatalf("fork expressions = %d", fork.Expressions().Count())
	}
	first := fork.Expressions().At(0)
	if first.NameEquals() != nil {
		t.Fatalf("first fork branch name = %#v", first.NameEquals())
	}
	if first.Expression().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("first fork branch expression kind = %v", first.Expression().Kind())
	}
	second := fork.Expressions().At(1)
	if second.NameEquals() == nil {
		t.Fatal("second fork branch missing name")
	}
	if second.NameEquals().Name().Text(IncludeTriviaMinimal) != "ByApp" {
		t.Fatalf("second fork branch name = %q", second.NameEquals().Name().Text(IncludeTriviaMinimal))
	}
	if second.Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("second fork branch expression kind = %v", second.Expression().Kind())
	}
}

func TestParseQueryBuildsFacetOperator(t *testing.T) {
	const text = "SigninLogs | facet by UserPrincipalName, AppDisplayName with (summarize Count=count() by ResourceDisplayName)"
	expr := parsedQueryExpression(t, text)

	facetPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("facet pipe = %#v", expr)
	}
	facet, ok := facetPipe.Operator().(*FacetOperator)
	if !ok {
		t.Fatalf("facet operator = %#v", facetPipe.Operator())
	}
	if facet.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("facet by keyword = %v", facet.ByKeyword().Kind())
	}
	if facet.Expressions().Count() != 2 {
		t.Fatalf("facet expressions = %d", facet.Expressions().Count())
	}
	if facet.Expressions().At(0).Element().Kind() != SyntaxKindNameReference {
		t.Fatalf("facet first expression = %v", facet.Expressions().At(0).Element().Kind())
	}
	withExpr, ok := facet.WithClause().(*FacetWithExpressionClause)
	if !ok {
		t.Fatalf("facet with clause = %#v", facet.WithClause())
	}
	if withExpr.Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("facet with expression kind = %v", withExpr.Expression().Kind())
	}
}

func TestParseFacetWithEmptyParenthesesReportsQueryOperatorExpected(t *testing.T) {
	query := ParseQuery("SigninLogs | facet by UserPrincipalName with ()")
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContain(diagnostics, "KS176") {
		t.Fatalf("expected facet with empty parentheses query-operator diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS006") {
		t.Fatalf("facet with empty parentheses should not report generic missing expression: %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("facet with empty parentheses should not add missing-token diagnostics, got %#v", diagnostics)
	}
}

func TestParseFacetWithClauseTailsUseQueryBlockRecovery(t *testing.T) {
	comma := Parse("T | facet by X with (where X == 1), Extra").GetDiagnostics()
	if !diagnosticsContainLocation(comma, "KS198", 34, 1) {
		t.Fatalf("expected comma skipped-token diagnostic, got %#v", comma)
	}
	if diagnosticsContain(comma, "KS001") {
		t.Fatalf("facet comma tail should not add generic syntax diagnostics, got %#v", comma)
	}

	expression := Parse("T | facet by X with (where X == 1) Extra").GetDiagnostics()
	if !diagnosticsContainLocation(expression, "KS005", 35, 5) {
		t.Fatalf("expected statement-separator diagnostic on expression tail, got %#v", expression)
	}
	if diagnosticsContain(expression, "KS198") || diagnosticsContain(expression, "KS001") {
		t.Fatalf("facet expression tail should remain a query-block statement, got %#v", expression)
	}
}

func diagnosticsContainLocation(diagnostics []Diagnostic, code string, start, length int) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Start == start && diagnostic.Length == length {
			return true
		}
	}
	return false
}

func TestParseFacetMissingKeyReportsMissingName(t *testing.T) {
	tests := []string{
		"SigninLogs | facet by with count",
		"SigninLogs | facet by UserPrincipalName, with count",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			diagnostics := ParseQuery(text).ContainedDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS006", "Missing name: If the keyword 'with' is intended be used as the name, it needs to be bracketted as ['with'].") {
				t.Fatalf("expected keyword-aware missing name diagnostic, got %#v", diagnostics)
			}
			if diagnosticsContainMessage(diagnostics, "KS006", "Missing expression: If the keyword 'with' is intended to be part of an expression it needs to be bracketted as ['with'].") {
				t.Fatalf("facet missing key should not report missing expression, got %#v", diagnostics)
			}
		})
	}
}

func TestParseQueryBuildsFacetOperatorWithBracketedPath(t *testing.T) {
	const text = `T | facet by Payload["user"] with count`
	expr := parsedQueryExpression(t, text)

	facetPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("facet pipe = %#v", expr)
	}
	facet, ok := facetPipe.Operator().(*FacetOperator)
	if !ok {
		t.Fatalf("facet operator = %#v", facetPipe.Operator())
	}
	if facet.Expressions().Count() != 1 {
		t.Fatalf("facet expressions = %d", facet.Expressions().Count())
	}
	element, ok := facet.Expressions().At(0).Element().(*ElementExpression)
	if !ok {
		t.Fatalf("facet bracketed expression = %#v", facet.Expressions().At(0).Element())
	}
	if element.Expression().Kind() != SyntaxKindNameReference || element.Selector().Kind() != SyntaxKindNameReference {
		t.Fatalf("facet element = %v / %v", element.Expression().Kind(), element.Selector().Kind())
	}
	if diagnostics := facet.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("facet bracketed path diagnostics = %#v", diagnostics)
	}
}

func TestParseQueryBuildsTopNestedOperator(t *testing.T) {
	const text = "T | top-nested 3 of UserPrincipalName with others='Other' by Count=count(), top-nested 2 of AppDisplayName by DurationMs desc"
	expr := parsedQueryExpression(t, text)

	topNestedPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-nested pipe = %#v", expr)
	}
	topNested, ok := topNestedPipe.Operator().(*TopNestedOperator)
	if !ok {
		t.Fatalf("top-nested operator = %#v", topNestedPipe.Operator())
	}
	if topNested.Clauses().Count() != 2 {
		t.Fatalf("top-nested clauses = %d", topNested.Clauses().Count())
	}

	first := topNested.Clauses().At(0).Element()
	if first.TopNestedKeyword().Kind() != SyntaxKindTopNestedKeyword {
		t.Fatalf("top-nested keyword = %v", first.TopNestedKeyword().Kind())
	}
	if first.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("top-nested expression kind = %v", first.Expression().Kind())
	}
	if first.OfExpression().Kind() != SyntaxKindNameReference {
		t.Fatalf("top-nested of expression kind = %v", first.OfExpression().Kind())
	}
	if first.WithOthersClause() == nil || first.WithOthersClause().Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("top-nested with-others clause = %#v", first.WithOthersClause())
	}
	if first.ByExpression().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("top-nested by expression kind = %v", first.ByExpression().Kind())
	}

	second := topNested.Clauses().At(1).Element()
	if second.WithOthersClause() != nil {
		t.Fatalf("second with-others clause = %#v", second.WithOthersClause())
	}
	if second.ByExpression().Kind() != SyntaxKindOrderedExpression {
		t.Fatalf("second by expression kind = %v", second.ByExpression().Kind())
	}
}

func TestParseQueryReportsMissingTopNestedClauseAfterTrailingComma(t *testing.T) {
	code := Parse(`MyTable | top-nested 5 of User by sum(Count),`)
	if !diagnosticsContainMessage(code.GetDiagnostics(), "KS006", "Missing clause") {
		t.Fatalf("expected missing top-nested clause diagnostic, got %#v", code.GetDiagnostics())
	}
	if len(code.GetDiagnostics()) != 1 {
		t.Fatalf("expected exactly one top-nested trailing-comma diagnostic, got %#v", code.GetDiagnostics())
	}
}

func TestParseTopNestedMissingOfKeywordRecoversClause(t *testing.T) {
	expr := parsedQueryExpression(t, "T | top-nested 3")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-nested pipe = %#v", expr)
	}
	topNested, ok := pipe.Operator().(*TopNestedOperator)
	if !ok {
		t.Fatalf("top-nested operator = %#v", pipe.Operator())
	}
	if topNested.Clauses().Count() != 1 {
		t.Fatalf("top-nested clauses = %d", topNested.Clauses().Count())
	}
	clause := topNested.Clauses().At(0).Element()
	if clause.OfKeyword().RawText() != "" {
		t.Fatalf("top-nested missing of keyword = %#v", clause.OfKeyword())
	}
	if clause.OfExpression() == nil {
		t.Fatal("top-nested missing of expression is nil")
	}
	diagnostics := clause.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: of") || !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing of/expression diagnostics, got %#v", diagnostics)
	}
}

func TestParseTopNestedMissingByExpressionRecoversClause(t *testing.T) {
	expr := parsedQueryExpression(t, "T | top-nested 3 of User by")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-nested pipe = %#v", expr)
	}
	topNested, ok := pipe.Operator().(*TopNestedOperator)
	if !ok {
		t.Fatalf("top-nested operator = %#v", pipe.Operator())
	}
	clause := topNested.Clauses().At(0).Element()
	if clause.ByExpression() == nil {
		t.Fatal("top-nested missing by expression recovery node is nil")
	}
	diagnostics := clause.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing by expression diagnostic, got %#v", diagnostics)
	}
}

func TestParseExpressionBuildsPipeExpression(t *testing.T) {
	expr := ParseExpression("T | project A | limit 10")
	limitPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("limit pipe = %#v", expr)
	}
	if _, ok := limitPipe.Operator().(*TakeOperator); !ok {
		t.Fatalf("limit operator = %#v", limitPipe.Operator())
	}
	projectPipe, ok := limitPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", limitPipe.Expression())
	}
	if _, ok := projectPipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
}

func TestParseQueryReportsUnknownPipeOperator(t *testing.T) {
	code := Parse(`MyTable | summarise count() by User`)
	if !diagnosticsContainMessage(code.GetDiagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("expected unknown pipe-operator diagnostic, got %#v", code.GetDiagnostics())
	}
}

func TestParseQueryBuildsUnpipedUnionOperator(t *testing.T) {
	const text = "union kind=outer withsource=SourceTable DeviceProcessEvents, Device* | where Timestamp > ago(1h) | project SourceTable, Timestamp"
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	if _, ok := projectPipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	wherePipe, ok := projectPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("where pipe = %#v", projectPipe.Expression())
	}
	if _, ok := wherePipe.Operator().(*FilterOperator); !ok {
		t.Fatalf("where operator = %#v", wherePipe.Operator())
	}
	union, ok := wherePipe.Expression().(*UnionOperator)
	if !ok {
		t.Fatalf("union expression = %#v", wherePipe.Expression())
	}
	if union.Parameters().Count() != 2 {
		t.Fatalf("union parameters = %d", union.Parameters().Count())
	}
	if got := namedParameterNameText(t, union.Parameters().At(0)); got != "kind" {
		t.Fatalf("union parameter 0 = %q", got)
	}
	if union.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("union kind expression = %v", union.Parameters().At(0).Expression().Kind())
	}
	if got := namedParameterNameText(t, union.Parameters().At(1)); got != "withsource" {
		t.Fatalf("union parameter 1 = %q", got)
	}
	if union.Expressions().Count() != 2 {
		t.Fatalf("union expressions = %d", union.Expressions().Count())
	}
	wildcard, ok := union.Expressions().At(1).Element().(*NameReference)
	if !ok {
		t.Fatalf("union wildcard expression = %#v", union.Expressions().At(1).Element())
	}
	if wildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("union wildcard kind = %v", wildcard.Name().Kind())
	}
}

func TestParseUnionWhitespaceBeforeParenthesizedArgument(t *testing.T) {
	const text = `union MyTable (RightTable)`
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("union parser diagnostics = %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	union, ok := expr.(*UnionOperator)
	if !ok {
		t.Fatalf("union expression = %#v", expr)
	}
	if union.Expressions().Count() != 1 {
		t.Fatalf("union expressions = %d", union.Expressions().Count())
	}
	first := union.Expressions().At(0)
	if first.Separator() != nil || first.Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("union first expression = %#v", first.Element())
	}

	functionExpr := parsedQueryExpression(t, `union MyFunc()`)
	functionUnion, ok := functionExpr.(*UnionOperator)
	if !ok {
		t.Fatalf("union function expression = %#v", functionExpr)
	}
	if functionUnion.Expressions().Count() != 1 || functionUnion.Expressions().At(0).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("union function operand = %#v", functionUnion.Expressions())
	}
}

func TestParseUnionAcceptsBracketedNameEntityPathSelector(t *testing.T) {
	expr := parsedQueryExpression(t, `union database("db")["SigninLogs"]`)
	union, ok := expr.(*UnionOperator)
	if !ok {
		t.Fatalf("union expression = %#v", expr)
	}
	if union.Expressions().Count() != 1 {
		t.Fatalf("union expressions = %d", union.Expressions().Count())
	}
	element, ok := union.Expressions().At(0).Element().(*ElementExpression)
	if !ok {
		t.Fatalf("union bracketed name path expression = %#v", union.Expressions().At(0).Element())
	}
	name, ok := element.Selector().(*NameReference)
	if !ok {
		t.Fatalf("union bracketed name path selector = %#v", element.Selector())
	}
	if name.Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("union bracketed name path selector kind = %v", name.Name().Kind())
	}
}

func TestParseUnionReportsMissingCloseBracketForWildcardedEntityPathSelector(t *testing.T) {
	query := ParseQuery(`union database("db")[ParityWild*`)
	expr := parsedQueryExpression(t, `union database("db")[ParityWild*`)
	union, ok := expr.(*UnionOperator)
	if !ok {
		t.Fatalf("union expression = %#v", expr)
	}
	if union.Expressions().Count() != 1 {
		t.Fatalf("union expressions = %d", union.Expressions().Count())
	}
	element, ok := union.Expressions().At(0).Element().(*ElementExpression)
	if !ok {
		t.Fatalf("union malformed bracketed wildcard path expression = %#v", union.Expressions().At(0).Element())
	}
	wildcard, ok := element.Selector().(*NameReference)
	if !ok {
		t.Fatalf("union malformed bracketed wildcard path selector = %#v", element.Selector())
	}
	if wildcard.Name().Kind() != SyntaxKindBracketedWildcardedName {
		t.Fatalf("union malformed bracketed wildcard path selector kind = %v", wildcard.Name().Kind())
	}
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing close bracket diagnostic, got %#v", query.ContainedDiagnostics())
	}
}

func TestParseQueryBuildsUnpipedPrintOperator(t *testing.T) {
	const text = "print a=true, b=1 + 2 | project a"
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	print, ok := projectPipe.Expression().(*PrintOperator)
	if !ok {
		t.Fatalf("print expression = %#v", projectPipe.Expression())
	}
	if print.Expressions().Count() != 2 {
		t.Fatalf("print expressions = %d", print.Expressions().Count())
	}
	if print.Expressions().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("print first expression kind = %v", print.Expressions().At(0).Element().Kind())
	}
}

func TestParseQueryRejectsToTableSelectorTail(t *testing.T) {
	const text = "print x = totable(T | project A).A"
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("totable selector diagnostics: %#v", diagnostics)
	}
	if query.GetChild(2) == nil {
		t.Fatal("totable selector tail was not retained as skipped tokens")
	}
	expr := parsedQueryExpression(t, text)
	print, ok := expr.(*PrintOperator)
	if !ok {
		t.Fatalf("print expression = %#v", expr)
	}
	if print.Expressions().Count() != 1 {
		t.Fatalf("print expressions = %d", print.Expressions().Count())
	}
	named, ok := print.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("print expression = %#v", print.Expressions().At(0).Element())
	}
	table, ok := named.Expression().(*ToTableExpression)
	if !ok {
		t.Fatalf("totable selector expression = %#v", named.Expression())
	}
	if _, ok := table.Expression().(*PipeExpression); !ok {
		t.Fatalf("totable prefix body = %T", table.Expression())
	}
}

func TestParseQueryToTableSelectorRootKeepsParsedSource(t *testing.T) {
	const text = "totable(print A=1).A"
	query := ParseQuery(text)
	if diagnostics := query.ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("totable root selector diagnostics: %#v", diagnostics)
	}
	if query.GetChild(2) == nil {
		t.Fatal("totable root selector tail was not retained as skipped tokens")
	}
	if _, ok := parsedQueryExpression(t, text).(*ToTableExpression); !ok {
		t.Fatalf("totable root expression = %T, want *ToTableExpression", parsedQueryExpression(t, text))
	}
}

func TestParseQueryBuildsCompoundNamedExpression(t *testing.T) {
	const text = "T | project (Alpha, Beta)=pack_array(Account, DeviceName)"
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 1 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}
	compound, ok := project.Expressions().At(0).Element().(*CompoundNamedExpression)
	if !ok {
		t.Fatalf("project expression = %#v", project.Expressions().At(0).Element())
	}
	if compound.Names().Names().Count() != 2 {
		t.Fatalf("compound names = %d", compound.Names().Names().Count())
	}
	if got := nameDeclarationText(t, compound.Names().Names().At(0).Element()); got != "Alpha" {
		t.Fatalf("compound name 0 = %q", got)
	}
	if got := nameDeclarationText(t, compound.Names().Names().At(1).Element()); got != "Beta" {
		t.Fatalf("compound name 1 = %q", got)
	}
	if compound.Expression().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("compound expression kind = %v", compound.Expression().Kind())
	}
}

func TestParseQueryBuildsCompoundStringBracketedNamedExpression(t *testing.T) {
	const text = "print ['a' 'b']=1"
	expr := parsedQueryExpression(t, text)

	print, ok := expr.(*PrintOperator)
	if !ok {
		t.Fatalf("print expression = %#v", expr)
	}
	if print.Expressions().Count() != 1 {
		t.Fatalf("print expressions = %d", print.Expressions().Count())
	}
	named, ok := print.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("print expression = %#v", print.Expressions().At(0).Element())
	}
	if got := NameDeclarationSimpleName(named.Name()); got != "ab" {
		t.Fatalf("compound bracketed name = %q", got)
	}
	bracketed, ok := named.Name().Name().(*BracketedName)
	if !ok {
		t.Fatalf("name kind = %#v", named.Name().Name())
	}
	if bracketed.Name().Kind() != SyntaxKindCompoundStringLiteralExpression {
		t.Fatalf("bracketed name expression kind = %v", bracketed.Name().Kind())
	}
}

func TestParseQueryCompoundStringBracketedNameMissingExpression(t *testing.T) {
	query := ParseQuery("print ['a' 'b']=")

	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContain(diagnostics, "KS006") {
		t.Fatalf("missing expression diagnostics = %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("compound bracketed name should scan before missing-expression recovery: %#v", diagnostics)
	}
}

func TestParseQueryDashedNamedExpressionRequiresBrackets(t *testing.T) {
	const text = "T | project foo-bar=Account"
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 1 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}
	named, ok := project.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("project expression = %#v", project.Expressions().At(0).Element())
	}
	if got := nameDeclarationText(t, named.Name()); got != "foo-bar" {
		t.Fatalf("dashed name = %q", got)
	}
	diagnostics := named.Name().Diagnostics()
	if !diagnosticsContain(diagnostics, "KS228") {
		t.Fatalf("expected KS228 on dashed project name, got %#v", diagnostics)
	}
	if diagnostics[0].Message != "The name 'foo-bar' needs to be bracketed as ['foo-bar'] to be used in this context." {
		t.Fatalf("dashed-name diagnostic message = %q", diagnostics[0].Message)
	}
}

func TestParseQueryInvalidKeywordNameReferenceRequiresBrackets(t *testing.T) {
	const text = "T | project project, Account"
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 2 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}
	keyword, ok := project.Expressions().At(0).Element().(*NameReference)
	if !ok {
		t.Fatalf("keyword expression = %#v", project.Expressions().At(0).Element())
	}
	if got := tokenNameRawText(keyword.Name()); got != "project" {
		t.Fatalf("keyword name = %q", got)
	}
	diagnostics := keyword.Diagnostics()
	if !diagnosticsContain(diagnostics, "KS228") {
		t.Fatalf("expected KS228 on keyword name, got %#v", diagnostics)
	}
	if diagnostics[0].Message != "The name 'project' needs to be bracketed as ['project'] to be used in this context." {
		t.Fatalf("keyword diagnostic message = %q", diagnostics[0].Message)
	}
}

func TestParseQuerySpecialKeywordsAsPathSelectors(t *testing.T) {
	const text = "T | project K=Bag.kind, WS=Bag.withsource, WSU=Bag.with_source, Root=kind"
	if diagnostics := Parse(text).GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 4 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}

	tests := []struct {
		index int
		name  string
	}{
		{0, "kind"},
		{1, "withsource"},
		{2, "with_source"},
	}
	for _, tt := range tests {
		named, ok := project.Expressions().At(tt.index).Element().(*SimpleNamedExpression)
		if !ok {
			t.Fatalf("expression %d = %#v", tt.index, project.Expressions().At(tt.index).Element())
		}
		path, ok := named.Expression().(*PathExpression)
		if !ok {
			t.Fatalf("expression %d path = %#v", tt.index, named.Expression())
		}
		selector, ok := path.Selector().(*NameReference)
		if !ok {
			t.Fatalf("expression %d selector = %#v", tt.index, path.Selector())
		}
		if got := NameReferenceSimpleName(selector); got != tt.name {
			t.Fatalf("expression %d selector name = %q, want %q", tt.index, got, tt.name)
		}
	}

	root, ok := project.Expressions().At(3).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("root expression = %#v", project.Expressions().At(3).Element())
	}
	rootReference, ok := root.Expression().(*NameReference)
	if !ok {
		t.Fatalf("root special keyword expression = %#v", root.Expression())
	}
	if got := NameReferenceSimpleName(rootReference); got != "kind" {
		t.Fatalf("root special keyword name = %q", got)
	}
}

func TestParseQueryFixedTextMatchingIsCaseSensitive(t *testing.T) {
	const text = "T | summarize First=minif(Count, Count > 0) by User | project Missing=Bag.Missing | render timechart with (title=Title)"
	if diagnostics := Parse(text).GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	if render.WithClause() == nil || render.WithClause().Properties().Count() != 1 {
		t.Fatalf("render with clause = %#v", render.WithClause())
	}
	title := render.WithClause().Properties().At(0).Element()
	if got := namedParameterNameText(t, title); got != "title" {
		t.Fatalf("render property name = %q", got)
	}
	titleValue, ok := title.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(titleValue) != "Title" {
		t.Fatalf("render title expression = %#v", title.Expression())
	}

	projectPipe, ok := renderPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", renderPipe.Expression())
	}
	project, ok := projectPipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	missing, ok := project.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(missing.Name()) != "Missing" {
		t.Fatalf("project expression = %#v", project.Expressions().At(0).Element())
	}
	missingPath, ok := missing.Expression().(*PathExpression)
	if !ok {
		t.Fatalf("missing path = %#v", missing.Expression())
	}
	missingSelector, ok := missingPath.Selector().(*NameReference)
	if !ok || NameReferenceSimpleName(missingSelector) != "Missing" {
		t.Fatalf("missing selector = %#v", missingPath.Selector())
	}

	summarizePipe, ok := projectPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("summarize pipe = %#v", projectPipe.Expression())
	}
	summarize, ok := summarizePipe.Operator().(*SummarizeOperator)
	if !ok {
		t.Fatalf("summarize operator = %#v", summarizePipe.Operator())
	}
	first, ok := summarize.Aggregates().At(0).Element().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(first.Name()) != "First" {
		t.Fatalf("summarize aggregate = %#v", summarize.Aggregates().At(0).Element())
	}
}

func TestParseQueryBuildsDataTableExpression(t *testing.T) {
	const text = `datatable (Account:string, Score:long) ["alpha", 1, "beta", 2] | project Account`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	if _, ok := projectPipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	dataTable, ok := projectPipe.Expression().(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", projectPipe.Expression())
	}
	if dataTable.Parameters().Count() != 0 {
		t.Fatalf("datatable parameters = %d", dataTable.Parameters().Count())
	}
	if dataTable.Schema().Columns().Count() != 2 {
		t.Fatalf("datatable schema columns = %d", dataTable.Schema().Columns().Count())
	}
	if dataTable.OpenBracket().Kind() != SyntaxKindOpenBracketToken || dataTable.CloseBracket().Kind() != SyntaxKindCloseBracketToken {
		t.Fatalf("datatable brackets = %v/%v", dataTable.OpenBracket().Kind(), dataTable.CloseBracket().Kind())
	}
	if dataTable.Values().Count() != 4 {
		t.Fatalf("datatable values = %d", dataTable.Values().Count())
	}
	if dataTable.Values().At(0).Element().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("datatable value 0 kind = %v", dataTable.Values().At(0).Element().Kind())
	}
	if dataTable.Values().At(1).Element().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("datatable value 1 kind = %v", dataTable.Values().At(1).Element().Kind())
	}
}

func TestParseQueryBuildsExternalDataExpression(t *testing.T) {
	const text = `externaldata(User:string, ['bad name']:long)["https://example.test/a.csv", "https://example.test/b.csv"] with(format="csv", ignoreFirstRecord=true, data=native) | project User`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	external, ok := projectPipe.Expression().(*ExternalDataExpression)
	if !ok {
		t.Fatalf("externaldata expression = %#v", projectPipe.Expression())
	}
	if external.ExternalDataKeyword().Kind() != SyntaxKindExternalDataKeyword {
		t.Fatalf("externaldata keyword = %v", external.ExternalDataKeyword().Kind())
	}
	if external.Schema().Columns().Count() != 2 {
		t.Fatalf("externaldata schema columns = %d", external.Schema().Columns().Count())
	}
	if external.URIs().Count() != 2 {
		t.Fatalf("externaldata uri count = %d", external.URIs().Count())
	}
	if external.WithClause() == nil || external.WithClause().Properties().Count() != 3 {
		t.Fatalf("externaldata with clause = %#v", external.WithClause())
	}
	if got := namedParameterNameText(t, external.WithClause().Properties().At(0).Element()); got != "format" {
		t.Fatalf("externaldata first property = %q", got)
	}
	dataProperty := external.WithClause().Properties().At(2).Element()
	if got := namedParameterNameText(t, dataProperty); got != "data" {
		t.Fatalf("externaldata third property = %q", got)
	}
	if dataProperty.Expression().Kind() != SyntaxKindNameDeclaration || NameDeclarationSimpleName(dataProperty.Expression().(*NameDeclaration)) != "native" {
		t.Fatalf("externaldata name property expression = %#v", dataProperty.Expression())
	}

	legacy := ParseExpression(`external_data(X:string)["https://example.test/a.csv"]`)
	legacyExpr, ok := legacy.(*ExternalDataExpression)
	if !ok {
		t.Fatalf("legacy external_data expression = %#v", legacy)
	}
	if legacyExpr.ExternalDataKeyword().Kind() != SyntaxKindExternal_DataKeyword {
		t.Fatalf("legacy external_data keyword = %v", legacyExpr.ExternalDataKeyword().Kind())
	}
}

func TestParseExternalDataPropertyValuesUseRestrictedSourceGrammar(t *testing.T) {
	for _, value := range []string{`dynamic([])`, `1.25d`, `1h`} {
		t.Run(value, func(t *testing.T) {
			expression := ParseExpression(`externaldata(Value:string)["https://example.test/a.csv"] with(option=` + value + `)`)
			external, ok := expression.(*ExternalDataExpression)
			if !ok || external.WithClause() == nil || external.WithClause().Properties().Count() != 1 {
				t.Fatalf("restricted externaldata property expression = %#v", expression)
			}
			property := external.WithClause().Properties().At(0).Element()
			if property.Expression().Width() != 0 ||
				!diagnosticsContainMessage(property.Expression().ContainedDiagnostics(), "KS006", "Missing value") {
				t.Fatalf("restricted externaldata property value %q = %#v diagnostics:%#v", value, property.Expression(), property.Expression().ContainedDiagnostics())
			}
		})
	}
}

func TestParseExternalDataWithClauseTrailingCommaCreatesMissingParameter(t *testing.T) {
	const text = `externaldata(Account:string)["https://example.test/signins.csv"] with(format="csv",) | project Account`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	external, ok := projectPipe.Expression().(*ExternalDataExpression)
	if !ok {
		t.Fatalf("externaldata expression = %#v", projectPipe.Expression())
	}
	if external.WithClause() == nil || external.WithClause().Properties().Count() != 2 {
		t.Fatalf("externaldata with clause = %#v", external.WithClause())
	}
	missing := external.WithClause().Properties().At(1).Element()
	if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("trailing comma diagnostics = %#v", missing.ContainedDiagnostics())
	}
}

func TestParseExternalDataCommaTailReportsIncompleteFragment(t *testing.T) {
	const text = `externaldata(x:long)["https://example.test/a.csv"], Tail | project x`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected externaldata comma-tail diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if diagnosticsContain(query.ContainedDiagnostics(), "KS001") {
		t.Fatalf("externaldata comma-tail should not cascade to missing-token diagnostics: %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	external, ok := expr.(*ExternalDataExpression)
	if !ok {
		t.Fatalf("externaldata expression = %#v", expr)
	}
	if external.Schema().Columns().Count() != 1 {
		t.Fatalf("externaldata schema columns = %d", external.Schema().Columns().Count())
	}
}

func TestParseQueryBuildsMaterializedViewCombineExpression(t *testing.T) {
	const text = `materialized-view-combine("SigninView") base(SigninLogs | project UserPrincipalName) delta(SigninLogs | where ResultType != 0) aggregations(summarize Count=count() by UserPrincipalName) | project UserPrincipalName, Count`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	combine, ok := projectPipe.Expression().(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("materialized-view-combine expression = %#v", projectPipe.Expression())
	}
	if combine.MaterializedViewCombineKeyword().Kind() != SyntaxKindMaterializedViewCombineKeyword {
		t.Fatalf("materialized-view-combine keyword = %v", combine.MaterializedViewCombineKeyword().Kind())
	}
	if combine.ViewName().Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("view name = %#v", combine.ViewName().Value())
	}
	if combine.BaseClause().Keyword().RawText() != "base" || combine.DeltaClause().Keyword().RawText() != "delta" {
		t.Fatalf("base/delta keywords = %q/%q", combine.BaseClause().Keyword().RawText(), combine.DeltaClause().Keyword().RawText())
	}
	if _, ok := combine.BaseClause().Expression().(*PipeExpression); !ok {
		t.Fatalf("base expression = %#v", combine.BaseClause().Expression())
	}
	if _, ok := combine.AggregationsClause().Expression().(*SummarizeOperator); !ok {
		t.Fatalf("aggregations expression = %#v", combine.AggregationsClause().Expression())
	}
}

func TestParseMaterializedViewCombineClauseCommaTailRecovers(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "base clause",
			text: `materialized-view-combine("v") base(T), Extra delta(T) aggregations(summarize count())`,
		},
		{
			name: "aggregations clause",
			text: `materialized-view-combine("v") base(T) delta(T) aggregations(summarize count()), Extra | project count_`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
				t.Fatalf("expected incomplete-fragment diagnostic, got %#v", query.ContainedDiagnostics())
			}
			if diagnosticsContain(query.ContainedDiagnostics(), "KS001") {
				t.Fatalf("clause comma-tail should not add missing-token cascades: %#v", query.ContainedDiagnostics())
			}
			expr := parsedQueryExpression(t, tt.text)
			if pipe, ok := expr.(*PipeExpression); ok {
				expr = pipe.Expression()
			}
			combine, ok := expr.(*MaterializedViewCombineExpression)
			if !ok {
				t.Fatalf("materialized-view-combine expression = %#v", expr)
			}
			if tt.name == "base clause" {
				if combine.DeltaClause().Keyword().RawText() != "" || combine.AggregationsClause().Keyword().RawText() != "" {
					t.Fatalf("clauses after base tail should be missing: delta=%#v aggregations=%#v", combine.DeltaClause(), combine.AggregationsClause())
				}
			} else if combine.AggregationsClause().Keyword().RawText() != "aggregations" {
				t.Fatalf("aggregations clause = %#v", combine.AggregationsClause())
			}
		})
	}
}

func TestParseMaterializedViewCombineMalformedAggregationsRecover(t *testing.T) {
	const text = `materialized-view-combine() base(T) delta(T) aggregations(count())`
	query := ParseQuery(text)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() == 0 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	separated, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("statement element = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("statement = %#v", separated.Element())
	}
	expr := statement.Expression()

	combine, ok := expr.(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("materialized-view-combine expression = %#v", expr)
	}
	if combine.AggregationsClause().Expression() == nil {
		t.Fatalf("malformed aggregations clause should keep a recovery expression")
	}
	diagnostics := expr.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing-expression diagnostic, got %#v", diagnostics)
	}
}

func TestParseMaterializedViewCombineNameClauseCommaTail(t *testing.T) {
	const text = `materialized-view-combine("v",) base(T) delta(T) aggregations(summarize count())`
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: )") || !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected missing close-paren and skipped-tail diagnostics, got %#v", diagnostics)
	}
	if query.GetChild(2) == nil {
		t.Fatalf("expected skipped comma tail in query block")
	}
}

func TestParseMaterializedViewCombineMissingAggregationsClause(t *testing.T) {
	const text = `materialized-view-combine("v") base(T) delta(T)`
	expr := parsedQueryExpression(t, text)

	combine, ok := expr.(*MaterializedViewCombineExpression)
	if !ok {
		t.Fatalf("materialized-view-combine expression = %#v", expr)
	}
	if combine.AggregationsClause() == nil || combine.AggregationsClause().Expression() == nil {
		t.Fatalf("missing aggregations clause should keep a recovery expression: %#v", combine.AggregationsClause())
	}
	diagnostics := expr.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing aggregations clause") || diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("expected missing aggregations diagnostics, got %#v", diagnostics)
	}
}

func TestParseMaterializedViewCombineMissingBaseAndDeltaClauses(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantMissing string
	}{
		{
			name:        "missing base",
			text:        `materialized-view-combine("v") delta(T) aggregations(summarize count())`,
			wantMissing: "base",
		},
		{
			name:        "missing delta",
			text:        `materialized-view-combine("v") base(T) aggregations(summarize count())`,
			wantMissing: "delta",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parsedQueryExpression(t, tt.text)

			combine, ok := expr.(*MaterializedViewCombineExpression)
			if !ok {
				t.Fatalf("materialized-view-combine expression = %#v", expr)
			}
			if combine.BaseClause() == nil || combine.BaseClause().Expression() == nil {
				t.Fatalf("base clause should keep a recovery expression: %#v", combine.BaseClause())
			}
			if combine.DeltaClause() == nil || combine.DeltaClause().Expression() == nil {
				t.Fatalf("delta clause should keep a recovery expression: %#v", combine.DeltaClause())
			}
			diagnostics := expr.ContainedDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS006", "Missing "+tt.wantMissing+" clause") || diagnosticsContain(diagnostics, "KS001") {
				t.Fatalf("expected missing %s diagnostics, got %#v", tt.wantMissing, diagnostics)
			}
		})
	}
}

func TestParseQueryBuildsInlineExternalTableExpression(t *testing.T) {
	const text = `inline_external_table(Account:string, Score:long) kind=storage dataformat=csv ("https://example.test/a.csv") with(format="csv", data=native) | project Account`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	inline, ok := projectPipe.Expression().(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", projectPipe.Expression())
	}
	if inline.Schema().Columns().Count() != 2 {
		t.Fatalf("inline external table schema columns = %d", inline.Schema().Columns().Count())
	}
	if inline.KindParameter().Value().ValueText() != "storage" {
		t.Fatalf("inline external table kind = %q", inline.KindParameter().Value().ValueText())
	}
	if inline.DataFormatParameter().Value().ValueText() != "csv" {
		t.Fatalf("inline external table dataformat = %q", inline.DataFormatParameter().Value().ValueText())
	}
	if inline.ConnectionStrings().ConnectionStrings().Count() != 1 {
		t.Fatalf("inline external table connection strings = %d", inline.ConnectionStrings().ConnectionStrings().Count())
	}
	if inline.WithClause() == nil || inline.WithClause().Properties().Count() != 2 {
		t.Fatalf("inline external table with clause = %#v", inline.WithClause())
	}
	dataProperty := inline.WithClause().Properties().At(1).Element()
	if got := namedParameterNameText(t, dataProperty); got != "data" {
		t.Fatalf("inline external table second property = %q", got)
	}
	if dataProperty.Expression().Kind() != SyntaxKindNameDeclaration || NameDeclarationSimpleName(dataProperty.Expression().(*NameDeclaration)) != "native" {
		t.Fatalf("inline external table name property expression = %#v", dataProperty.Expression())
	}
	if _, ok := projectPipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
}

func TestParseInlineExternalTableCommaTailReportsIncompleteFragment(t *testing.T) {
	const text = `inline_external_table(Account:string) kind=storage dataformat=csv ("https://example.test/a.csv"), Extra`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected inline external table comma-tail diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	inline, ok := expr.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", expr)
	}
	if inline.Schema().Columns().Count() != 1 {
		t.Fatalf("inline external table schema columns = %d", inline.Schema().Columns().Count())
	}
}

func TestParseInlineExternalTableMissingDataFormatValue(t *testing.T) {
	const text = `inline_external_table(Account:string) kind=storage dataformat= ("https://example.test/a.csv")`
	code := Parse(text)
	diagnostics := code.GetDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing name") {
		t.Fatalf("expected missing dataformat name diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS253") || diagnosticsContain(diagnostics, "KS256") {
		t.Fatalf("dataformat recovery should not cascade into unrelated diagnostics, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	inline, ok := expr.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", expr)
	}
	if inline.ConnectionStrings().ConnectionStrings().Count() != 1 {
		t.Fatalf("inline external table connection strings = %d", inline.ConnectionStrings().ConnectionStrings().Count())
	}
}

func TestParseInlineExternalTableMissingKindValue(t *testing.T) {
	const text = `inline_external_table(Account:string) kind=() | project Account`
	code := Parse(text)
	diagnostics := code.GetDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected one of: storage, delta") {
		t.Fatalf("expected missing kind value diagnostic, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	inline, ok := projectPipe.Expression().(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", projectPipe.Expression())
	}
	if inline.Schema().Columns().Count() != 1 {
		t.Fatalf("inline external table schema columns = %d", inline.Schema().Columns().Count())
	}
}

func TestParseInlineExternalTableKeywordFunctionCallFallback(t *testing.T) {
	const text = `let inline_external_table = () { 1 };
print y=inline_external_table()`
	query := ParseQuery(text)
	if len(query.ContainedDiagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", query.ContainedDiagnostics())
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	second, ok := statements.At(1).(*SeparatedElement)
	if !ok {
		t.Fatalf("second statement element = %#v", statements.At(1))
	}
	statement, ok := WrapGeneratedSyntaxElement(second.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("second statement = %#v", second.Element())
	}
	printOp, ok := statement.Expression().(*PrintOperator)
	if !ok {
		t.Fatalf("print expression = %#v", statement.Expression())
	}
	value, ok := printOp.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("print value = %#v", printOp.Expressions().At(0).Element())
	}
	if _, ok := value.Expression().(*FunctionCallExpression); !ok {
		t.Fatalf("inline_external_table call expression = %#v", value.Expression())
	}
}

func TestParseInlineExternalTableEmptyArgListFallsBackToFunctionCall(t *testing.T) {
	const text = `inline_external_table() kind=storage dataformat=csv ('https://example.test/a.csv')`
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") || !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected skipped-tail diagnostics, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS253") || diagnosticsContain(diagnostics, "KS256") {
		t.Fatalf("function-call fallback should not emit inline external table diagnostics: %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() == 0 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(first.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	if _, ok := statement.Expression().(*FunctionCallExpression); !ok {
		t.Fatalf("empty arglist expression = %#v", statement.Expression())
	}
}

func TestParseInlineExternalTableEmptyConnectionStringsArePresent(t *testing.T) {
	const text = `inline_external_table(Account:string) kind=storage dataformat=csv ()`
	code := Parse(text)
	diagnostics := code.GetDiagnostics()
	if diagnosticsContainMessage(diagnostics, "KS256", "Connection strings definition is expected.") {
		t.Fatalf("explicit empty connection clause reported missing, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	inline, ok := expr.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", expr)
	}
	if inline.ConnectionStrings().ConnectionStrings().Count() != 0 {
		t.Fatalf("inline external table connection strings = %d", inline.ConnectionStrings().ConnectionStrings().Count())
	}
}

func TestParseQueryBuildsPipedInlineExternalTableExpression(t *testing.T) {
	const text = `datatable(Account:string)["old"] | inline_external_table(Account:string, Score:long) kind=storage dataformat=csv ("https://example.test/a.csv")`
	expr := parsedQueryExpression(t, text)

	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %#v", expr)
	}
	if _, ok := pipe.Expression().(*DataTableExpression); !ok {
		t.Fatalf("pipe source = %#v", pipe.Expression())
	}
	inline, ok := pipe.Operator().(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("piped inline external table = %#v", pipe.Operator())
	}
	if inline.Schema().Columns().Count() != 2 {
		t.Fatalf("inline external table schema columns = %d", inline.Schema().Columns().Count())
	}
}

func TestParseQueryOperatorLookaheadFacetBy(t *testing.T) {
	expr := parsedQueryExpression(t, `facet by Account`)
	facet, ok := expr.(*FacetOperator)
	if !ok {
		t.Fatalf("facet expression = %#v", expr)
	}
	if facet.Expressions().Count() != 1 {
		t.Fatalf("facet expressions = %d", facet.Expressions().Count())
	}

	identifier := parsedQueryExpression(t, `facet`)
	if _, ok := identifier.(*FacetOperator); ok {
		t.Fatalf("bare facet parsed as operator: %#v", identifier)
	}
}

func TestParseQueryBuildsInlineExternalTablePartitionAndPathFormat(t *testing.T) {
	const text = `inline_external_table(Account:string, Timestamp:datetime) kind=storage partition by (Tenant:string, Day:datetime=startofday(Timestamp),) pathformat=("/" Tenant "/" datetime_pattern("yyyy/MM/dd", Day)) dataformat=csv ("https://example.test/a.csv")`
	expr := parsedQueryExpression(t, text)
	inline, ok := expr.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", expr)
	}
	if inline.PartitionClause() == nil || inline.PartitionClause().PartitionColumns().Count() != 2 {
		t.Fatalf("partition clause = %#v", inline.PartitionClause())
	}
	if inline.PathFormat() == nil || inline.PathFormat().PathExpressions().Count() != 2 {
		t.Fatalf("pathformat clause = %#v", inline.PathFormat())
	}
	if inline.PathFormat().OptionalSeparatorLiteral() == nil || inline.PathFormat().OptionalSeparatorLiteral().Token().ValueText() != "/" {
		t.Fatalf("pathformat separator = %#v", inline.PathFormat().OptionalSeparatorLiteral())
	}
	second := inline.PathFormat().PathExpressions().At(1)
	if _, ok := second.PartitionColumnExpression().(*DateTimePattern); !ok {
		t.Fatalf("second pathformat expression = %#v", second.PartitionColumnExpression())
	}
}

func TestParseQueryBuildsDeltaInlineExternalTableWithoutSchemaDiagnostics(t *testing.T) {
	const text = `inline_external_table kind=delta ("https://example.test/root")`
	code := Parse(text)
	expr := parsedQueryExpression(t, text)
	inline, ok := expr.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %#v", expr)
	}
	if inline.Schema() == nil || inline.Schema().Columns().Count() != 0 || inline.Schema().OpenParen().RawText() != "" {
		t.Fatalf("delta omitted schema = %#v", inline.Schema())
	}
	if inline.DataFormatParameter() == nil || inline.DataFormatParameter().DataFormatKeyword().RawText() != "" {
		t.Fatalf("delta omitted dataformat = %#v", inline.DataFormatParameter())
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("delta diagnostics = %#v", diagnostics)
	}
}

func TestParseQueryBuildsDataTableLeadingAndTrailingCommaValues(t *testing.T) {
	const text = `datatable (Account:string) [, "alpha",]`
	dataTable, ok := parsedQueryExpression(t, text).(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", parsedQueryExpression(t, text))
	}
	if dataTable.LeadingComma() == nil || dataTable.LeadingComma().Kind() != SyntaxKindCommaToken {
		t.Fatalf("datatable leading comma = %#v", dataTable.LeadingComma())
	}
	if dataTable.Values().Count() != 1 {
		t.Fatalf("datatable values = %d", dataTable.Values().Count())
	}
	if dataTable.Values().At(0).Separator() == nil || dataTable.Values().At(0).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("datatable trailing separator = %#v", dataTable.Values().At(0).Separator())
	}
}

func TestParseQueryDataTableInvalidLiteralValueStopsValueList(t *testing.T) {
	const text = `datatable (Account:string) [strcat("a", "b")]`
	code := Parse(text)
	diagnostics := code.GetDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ]") ||
		!diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") ||
		!diagnosticsContain(diagnostics, "KS198") ||
		diagnosticsContain(diagnostics, "KS006") {
		t.Fatalf("invalid datatable literal diagnostics = %#v", diagnostics)
	}
	dataTable, ok := firstParsedQueryExpression(t, text).(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", firstParsedQueryExpression(t, text))
	}
	if dataTable.Values().Count() != 0 {
		t.Fatalf("datatable values = %d", dataTable.Values().Count())
	}
}

func TestParseQueryDataTableDynamicLiteralDoesNotBecomeSkippedTokens(t *testing.T) {
	const text = `datatable (Payload:dynamic) [dynamic({"x":1})]`
	code := Parse(text)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics = %#v", diagnostics)
	}
	dataTable, ok := parsedQueryExpression(t, text).(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", parsedQueryExpression(t, text))
	}
	if dataTable.Values().Count() != 1 {
		t.Fatalf("datatable values = %d", dataTable.Values().Count())
	}
	if skipped := code.Query.GetChild(2); skipped != nil {
		t.Fatalf("unexpected skipped tokens = %#v", skipped)
	}
}

func TestParseQueryDataTableGooLiteralsDoNotBecomeMissingValues(t *testing.T) {
	const text = `datatable (Timestamp:datetime, Duration:timespan, Payload:dynamic) [datetime(1910-06-11), timespan(1d), dynamic({"x":1})]`
	code := Parse(text)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics = %#v", diagnostics)
	}
	dataTable, ok := parsedQueryExpression(t, text).(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", parsedQueryExpression(t, text))
	}
	if dataTable.Values().Count() != 3 {
		t.Fatalf("datatable values = %d", dataTable.Values().Count())
	}
	if dataTable.Values().At(0).Element().Kind() != SyntaxKindDateTimeLiteralExpression {
		t.Fatalf("value 0 kind = %v", dataTable.Values().At(0).Element().Kind())
	}
	if dataTable.Values().At(1).Element().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("value 1 kind = %v", dataTable.Values().At(1).Element().Kind())
	}
	if dataTable.Values().At(2).Element().Kind() != SyntaxKindDynamicExpression {
		t.Fatalf("value 2 kind = %v", dataTable.Values().At(2).Element().Kind())
	}
}

func TestParseExpressionBuildsContextualDataTableExpression(t *testing.T) {
	expr := ParseExpression(`__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (Account:string)`)
	contextual, ok := expr.(*ContextualDataTableExpression)
	if !ok {
		t.Fatalf("contextual datatable = %#v", expr)
	}
	if contextual.Id().Kind() != SyntaxKindGuidLiteralExpression {
		t.Fatalf("contextual id kind = %v", contextual.Id().Kind())
	}
	if contextual.Schema().Columns().Count() != 1 {
		t.Fatalf("contextual schema columns = %d", contextual.Schema().Columns().Count())
	}
}

func TestParseExpressionMissingRowSchemaRecovery(t *testing.T) {
	datatable := ParseExpression(`datatable`)
	dt, ok := datatable.(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %#v", datatable)
	}
	if dt.Schema() == nil || !diagnosticsContainMessage(dt.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("datatable missing schema diagnostics = %#v", dt.ContainedDiagnostics())
	}

	contextual := ParseExpression(`__contextual_datatable`)
	ct, ok := contextual.(*ContextualDataTableExpression)
	if !ok {
		t.Fatalf("contextual datatable expression = %#v", contextual)
	}
	if ct.Schema() == nil || !diagnosticsContainMessage(ct.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("contextual missing schema diagnostics = %#v", ct.ContainedDiagnostics())
	}
}

func TestParseRawGuidLiteralCarriesDiagnostic(t *testing.T) {
	expr := ParseExpression(`00000000-0000-0000-0000-000000000001`)
	if expr == nil || expr.Kind() != SyntaxKindGuidLiteralExpression {
		t.Fatalf("raw guid expression = %#v", expr)
	}
	if !diagnosticsContain(expr.ContainedDiagnostics(), "KS214") {
		t.Fatalf("expected raw guid diagnostic, got %#v", expr.ContainedDiagnostics())
	}
}

func TestParseQueryBuildsAggregationAndOrderingOperators(t *testing.T) {
	const text = "T | summarize hint.shufflekey = UserPrincipalName FailedCount=count() by UserPrincipalName bin=1h | sort hint.strategy=shuffle by FailedCount desc nulls last | top hint.progressive_top=true 5 by Timestamp desc"
	expr := parsedQueryExpression(t, text)

	topPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top pipe = %#v", expr)
	}
	top, ok := topPipe.Operator().(*TopOperator)
	if !ok {
		t.Fatalf("top operator = %#v", topPipe.Operator())
	}
	if top.Parameters().Count() != 1 {
		t.Fatalf("top parameters = %d", top.Parameters().Count())
	}
	if got := namedParameterNameText(t, top.Parameters().At(0)); got != "hint.progressive_top" {
		t.Fatalf("top parameter name = %q", got)
	}
	if top.Parameters().At(0).Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("top parameter expression kind = %v", top.Parameters().At(0).Expression().Kind())
	}
	if top.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("top expression kind = %v", top.Expression().Kind())
	}
	if top.ByExpression().Kind() != SyntaxKindOrderedExpression {
		t.Fatalf("top by expression kind = %v", top.ByExpression().Kind())
	}

	sortPipe, ok := topPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("sort pipe = %#v", topPipe.Expression())
	}
	sort, ok := sortPipe.Operator().(*SortOperator)
	if !ok {
		t.Fatalf("sort operator = %#v", sortPipe.Operator())
	}
	if sort.Parameters().Count() != 1 {
		t.Fatalf("sort parameters = %d", sort.Parameters().Count())
	}
	if got := namedParameterNameText(t, sort.Parameters().At(0)); got != "hint.strategy" {
		t.Fatalf("sort parameter name = %q", got)
	}
	if sort.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("sort parameter expression kind = %v", sort.Parameters().At(0).Expression().Kind())
	}
	if sort.Expressions().Count() != 1 {
		t.Fatalf("sort expressions = %d", sort.Expressions().Count())
	}
	ordered, ok := sort.Expressions().At(0).Element().(*OrderedExpression)
	if !ok {
		t.Fatalf("sort ordered expression = %#v", sort.Expressions().At(0).Element())
	}
	if ordered.Ordering().AscOrDescKeyword().Kind() != SyntaxKindDescKeyword {
		t.Fatalf("sort ordering kind = %v", ordered.Ordering().AscOrDescKeyword().Kind())
	}
	if ordered.Ordering().NullsClause() == nil || ordered.Ordering().NullsClause().FirstOrLastKeyword().Kind() != SyntaxKindLastKeyword {
		t.Fatalf("sort nulls clause = %#v", ordered.Ordering().NullsClause())
	}

	summarizePipe, ok := sortPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("summarize pipe = %#v", sortPipe.Expression())
	}
	summarize, ok := summarizePipe.Operator().(*SummarizeOperator)
	if !ok {
		t.Fatalf("summarize operator = %#v", summarizePipe.Operator())
	}
	if summarize.Parameters().Count() != 1 {
		t.Fatalf("summarize parameters = %d", summarize.Parameters().Count())
	}
	if got := namedParameterNameText(t, summarize.Parameters().At(0)); got != "hint.shufflekey" {
		t.Fatalf("summarize parameter name = %q", got)
	}
	if summarize.Parameters().At(0).ExpressionHint() != CompletionHintColumn {
		t.Fatalf("summarize parameter hint = %v", summarize.Parameters().At(0).ExpressionHint())
	}
	if summarize.Aggregates().Count() != 1 {
		t.Fatalf("summarize aggregate count = %d", summarize.Aggregates().Count())
	}
	if summarize.ByClause() == nil || summarize.ByClause().Expressions().Count() != 1 || summarize.ByClause().BinClause() == nil {
		t.Fatalf("summarize by clause = %#v", summarize.ByClause())
	}
}

func TestParseQueryBuildsDistinctAndProjectAwayOperators(t *testing.T) {
	const text = "T | project-away Account*, DeviceName | distinct hint.strategy=shuffle UserPrincipalName, *"
	expr := parsedQueryExpression(t, text)

	distinctPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("distinct pipe = %#v", expr)
	}
	distinct, ok := distinctPipe.Operator().(*DistinctOperator)
	if !ok {
		t.Fatalf("distinct operator = %#v", distinctPipe.Operator())
	}
	if distinct.Parameters().Count() != 1 {
		t.Fatalf("distinct parameters = %d", distinct.Parameters().Count())
	}
	if got := namedParameterNameText(t, distinct.Parameters().At(0)); got != "hint.strategy" {
		t.Fatalf("distinct parameter name = %q", got)
	}
	if distinct.Expressions().Count() != 2 {
		t.Fatalf("distinct expressions = %d", distinct.Expressions().Count())
	}
	if distinct.Expressions().At(1).Element().Kind() != SyntaxKindStarExpression {
		t.Fatalf("distinct star expression kind = %v", distinct.Expressions().At(1).Element().Kind())
	}

	projectAwayPipe, ok := distinctPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project-away pipe = %#v", distinctPipe.Expression())
	}
	projectAway, ok := projectAwayPipe.Operator().(*ProjectAwayOperator)
	if !ok {
		t.Fatalf("project-away operator = %#v", projectAwayPipe.Operator())
	}
	if projectAway.Expressions().Count() != 2 {
		t.Fatalf("project-away expressions = %d", projectAway.Expressions().Count())
	}
	wildcard, ok := projectAway.Expressions().At(0).Element().(*NameReference)
	if !ok {
		t.Fatalf("project-away wildcard = %#v", projectAway.Expressions().At(0).Element())
	}
	if wildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("wildcard name kind = %v", wildcard.Name().Kind())
	}
}

func TestParseQueryBuildsProjectVariantOperators(t *testing.T) {
	const text = `T | project-keep Account*, DeviceName | project-rename NewDevice=DeviceName | project-reorder NewDevice desc, Account* asc | project-by-names pack_array("NewDevice", "Account")`
	expr := parsedQueryExpression(t, text)

	byNamesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project-by-names pipe = %#v", expr)
	}
	byNames, ok := byNamesPipe.Operator().(*ProjectByNamesOperator)
	if !ok {
		t.Fatalf("project-by-names operator = %#v", byNamesPipe.Operator())
	}
	if byNames.Expressions().Count() != 1 {
		t.Fatalf("project-by-names expressions = %d", byNames.Expressions().Count())
	}
	if byNames.Expressions().At(0).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("project-by-names expression kind = %v", byNames.Expressions().At(0).Element().Kind())
	}

	reorderPipe, ok := byNamesPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project-reorder pipe = %#v", byNamesPipe.Expression())
	}
	reorder, ok := reorderPipe.Operator().(*ProjectReorderOperator)
	if !ok {
		t.Fatalf("project-reorder operator = %#v", reorderPipe.Operator())
	}
	if reorder.Expressions().Count() != 2 {
		t.Fatalf("project-reorder expressions = %d", reorder.Expressions().Count())
	}
	ordered, ok := reorder.Expressions().At(0).Element().(*OrderedExpression)
	if !ok {
		t.Fatalf("project-reorder ordered expression = %#v", reorder.Expressions().At(0).Element())
	}
	if ordered.Ordering().AscOrDescKeyword().Kind() != SyntaxKindDescKeyword {
		t.Fatalf("project-reorder ordering = %v", ordered.Ordering().AscOrDescKeyword().Kind())
	}
	if ordered.Ordering().NullsClause() != nil {
		t.Fatalf("project-reorder nulls clause = %#v", ordered.Ordering().NullsClause())
	}
	wildcardOrdered, ok := reorder.Expressions().At(1).Element().(*OrderedExpression)
	if !ok {
		t.Fatalf("project-reorder wildcard ordered expression = %#v", reorder.Expressions().At(1).Element())
	}
	wildcardName, ok := wildcardOrdered.Expression().(*NameReference)
	if !ok {
		t.Fatalf("project-reorder wildcard expression = %#v", wildcardOrdered.Expression())
	}
	if wildcardName.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("project-reorder wildcard name kind = %v", wildcardName.Name().Kind())
	}

	renamePipe, ok := reorderPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project-rename pipe = %#v", reorderPipe.Expression())
	}
	rename, ok := renamePipe.Operator().(*ProjectRenameOperator)
	if !ok {
		t.Fatalf("project-rename operator = %#v", renamePipe.Operator())
	}
	if rename.Expressions().Count() != 1 {
		t.Fatalf("project-rename expressions = %d", rename.Expressions().Count())
	}
	if rename.Expressions().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("project-rename expression kind = %v", rename.Expressions().At(0).Element().Kind())
	}

	keepPipe, ok := renamePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("project-keep pipe = %#v", renamePipe.Expression())
	}
	keep, ok := keepPipe.Operator().(*ProjectKeepOperator)
	if !ok {
		t.Fatalf("project-keep operator = %#v", keepPipe.Operator())
	}
	if keep.Expressions().Count() != 2 {
		t.Fatalf("project-keep expressions = %d", keep.Expressions().Count())
	}
	keepWildcard, ok := keep.Expressions().At(0).Element().(*NameReference)
	if !ok {
		t.Fatalf("project-keep wildcard = %#v", keep.Expressions().At(0).Element())
	}
	if keepWildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("project-keep wildcard kind = %v", keepWildcard.Name().Kind())
	}
}

func TestParseQueryProjectColumnListNoProgressRecovery(t *testing.T) {
	code := Parse("T | project-keep missing")
	if code == nil || code.Query == nil {
		t.Fatalf("missing query parse result")
	}
}

func TestParseInvokeBuildsDotCompositeFunctionCall(t *testing.T) {
	const text = "T | invoke database('db').Transform(1)"
	expr := parsedQueryExpression(t, text)

	invokePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("invoke pipe = %#v", expr)
	}
	invoke, ok := invokePipe.Operator().(*InvokeOperator)
	if !ok {
		t.Fatalf("invoke operator = %#v", invokePipe.Operator())
	}
	path, ok := invoke.Function().(*PathExpression)
	if !ok {
		t.Fatalf("invoke function = %#v", invoke.Function())
	}
	database, ok := path.Expression().(*FunctionCallExpression)
	if !ok {
		t.Fatalf("invoke path receiver = %#v", path.Expression())
	}
	databaseName, ok := database.Name().Name().(*TokenName)
	if !ok || databaseName.Name().RawText() != "database" {
		t.Fatalf("database call name = %#v", database.Name().Name())
	}
	transform, ok := path.Selector().(*FunctionCallExpression)
	if !ok {
		t.Fatalf("invoke path selector = %#v", path.Selector())
	}
	transformName, ok := transform.Name().Name().(*TokenName)
	if !ok || transformName.Name().RawText() != "Transform" {
		t.Fatalf("transform call name = %#v", transform.Name().Name())
	}
}

func TestParseQueryBuildsEvaluateOperator(t *testing.T) {
	const text = "T | evaluate hint.distribution=single bag_unpack(Properties) : (*, Account:string, Score:long)"
	expr := parsedQueryExpression(t, text)

	evaluatePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("evaluate pipe = %#v", expr)
	}
	evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
	if !ok {
		t.Fatalf("evaluate operator = %#v", evaluatePipe.Operator())
	}
	if evaluate.Parameters().Count() != 1 {
		t.Fatalf("evaluate parameters = %d", evaluate.Parameters().Count())
	}
	if got := namedParameterNameText(t, evaluate.Parameters().At(0)); got != "hint.distribution" {
		t.Fatalf("evaluate parameter name = %q", got)
	}
	if evaluate.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("evaluate parameter expression kind = %v", evaluate.Parameters().At(0).Expression().Kind())
	}
	callName, ok := evaluate.FunctionCall().Name().Name().(*TokenName)
	if !ok {
		t.Fatalf("evaluate function name = %#v", evaluate.FunctionCall().Name().Name())
	}
	if callName.Name().RawText() != "bag_unpack" {
		t.Fatalf("evaluate function name text = %q", callName.Name().RawText())
	}
	if evaluate.FunctionCall().ArgumentList().Expressions().Count() != 1 {
		t.Fatalf("evaluate function arguments = %d", evaluate.FunctionCall().ArgumentList().Expressions().Count())
	}
	if evaluate.Schema() == nil {
		t.Fatal("missing evaluate schema")
	}
	if evaluate.Schema().ColonToken().Kind() != SyntaxKindColonToken {
		t.Fatalf("evaluate schema colon = %v", evaluate.Schema().ColonToken().Kind())
	}
	row := evaluate.Schema().Schema()
	if row.AsteriskToken() == nil || row.AsteriskToken().Kind() != SyntaxKindAsteriskToken {
		t.Fatalf("evaluate row asterisk = %#v", row.AsteriskToken())
	}
	if row.AsteriskTokenComma() == nil || row.AsteriskTokenComma().Kind() != SyntaxKindCommaToken {
		t.Fatalf("evaluate row asterisk comma = %#v", row.AsteriskTokenComma())
	}
	if row.Columns().Count() != 2 {
		t.Fatalf("evaluate row columns = %d", row.Columns().Count())
	}
	if row.Columns().At(0).Element().Name().Name().Kind() != SyntaxKindTokenName {
		t.Fatalf("evaluate first column name kind = %v", row.Columns().At(0).Element().Name().Name().Kind())
	}
	if row.Columns().At(1).Element().Type().Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("evaluate second column type kind = %v", row.Columns().At(1).Element().Type().Kind())
	}
	if evaluatePipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("evaluate source kind = %v", evaluatePipe.Expression().Kind())
	}
}

func TestParseQueryBuildsJoinAndLookupOperators(t *testing.T) {
	const text = "T | join kind=leftouter hint.shufflekey=UserPrincipalName (SigninLogs) on UserPrincipalName, AppId | lookup kind=inner (IdentityInfo) on AccountObjectId"
	expr := parsedQueryExpression(t, text)

	lookupPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("lookup pipe = %#v", expr)
	}
	lookup, ok := lookupPipe.Operator().(*LookupOperator)
	if !ok {
		t.Fatalf("lookup operator = %#v", lookupPipe.Operator())
	}
	if lookup.Parameters().Count() != 1 {
		t.Fatalf("lookup parameters = %d", lookup.Parameters().Count())
	}
	if got := namedParameterNameText(t, lookup.Parameters().At(0)); got != "kind" {
		t.Fatalf("lookup parameter name = %q", got)
	}
	if lookup.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("lookup kind expression = %v", lookup.Parameters().At(0).Expression().Kind())
	}
	if lookup.Expression().Kind() != SyntaxKindParenthesizedExpression {
		t.Fatalf("lookup expression kind = %v", lookup.Expression().Kind())
	}
	lookupOn, ok := lookup.LookupClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("lookup on clause = %#v", lookup.LookupClause())
	}
	if lookupOn.Expressions().Count() != 1 {
		t.Fatalf("lookup on expressions = %d", lookupOn.Expressions().Count())
	}
	if lookupOn.Expressions().At(0).Element().Kind() != SyntaxKindNameReference {
		t.Fatalf("lookup on expression kind = %v", lookupOn.Expressions().At(0).Element().Kind())
	}

	joinPipe, ok := lookupPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("join pipe = %#v", lookupPipe.Expression())
	}
	join, ok := joinPipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("join operator = %#v", joinPipe.Operator())
	}
	if join.Parameters().Count() != 2 {
		t.Fatalf("join parameters = %d", join.Parameters().Count())
	}
	if got := namedParameterNameText(t, join.Parameters().At(0)); got != "kind" {
		t.Fatalf("join parameter 0 = %q", got)
	}
	if got := namedParameterNameText(t, join.Parameters().At(1)); got != "hint.shufflekey" {
		t.Fatalf("join parameter 1 = %q", got)
	}
	if join.Parameters().At(1).ExpressionHint() != CompletionHintColumn {
		t.Fatalf("join shufflekey hint = %v", join.Parameters().At(1).ExpressionHint())
	}
	if join.Expression().Kind() != SyntaxKindParenthesizedExpression {
		t.Fatalf("join expression kind = %v", join.Expression().Kind())
	}
	joinOn, ok := join.ConditionClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("join on clause = %#v", join.ConditionClause())
	}
	if joinOn.Expressions().Count() != 2 {
		t.Fatalf("join on expressions = %d", joinOn.Expressions().Count())
	}
	if joinOn.Expressions().At(1).Element().Kind() != SyntaxKindNameReference {
		t.Fatalf("join on expression 1 kind = %v", joinOn.Expressions().At(1).Element().Kind())
	}
	if joinPipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("join source kind = %v", joinPipe.Expression().Kind())
	}
}

func TestParseQueryBuildsParseKvOperator(t *testing.T) {
	const text = `T | parse-kv Message as (Account:string, Score:long) with (pair_delimiter=' ', kv_delimiter='=', quote='"', greedy=true)`
	expr := parsedQueryExpression(t, text)

	parsePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse-kv pipe = %#v", expr)
	}
	parseKv, ok := parsePipe.Operator().(*ParseKvOperator)
	if !ok {
		t.Fatalf("parse-kv operator = %#v", parsePipe.Operator())
	}
	if parseKv.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("parse-kv expression kind = %v", parseKv.Expression().Kind())
	}
	if parseKv.AsKeyword().Kind() != SyntaxKindAsKeyword {
		t.Fatalf("parse-kv as keyword = %v", parseKv.AsKeyword().Kind())
	}
	if parseKv.Keys().Columns().Count() != 2 {
		t.Fatalf("parse-kv key columns = %d", parseKv.Keys().Columns().Count())
	}
	if parseKv.WithClause() == nil {
		t.Fatal("missing parse-kv with clause")
	}
	properties := parseKv.WithClause().Properties()
	if properties.Count() != 4 {
		t.Fatalf("parse-kv properties = %d", properties.Count())
	}
	if got := namedParameterNameText(t, properties.At(0).Element()); got != "pair_delimiter" {
		t.Fatalf("parse-kv property 0 = %q", got)
	}
	if properties.At(0).Element().Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("pair_delimiter expression kind = %v", properties.At(0).Element().Expression().Kind())
	}
	if got := namedParameterNameText(t, properties.At(2).Element()); got != "quote" {
		t.Fatalf("parse-kv property 2 = %q", got)
	}
	if got := namedParameterNameText(t, properties.At(3).Element()); got != "greedy" {
		t.Fatalf("parse-kv property 3 = %q", got)
	}
	if properties.At(3).Element().Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("greedy expression kind = %v", properties.At(3).Element().Expression().Kind())
	}
	if parsePipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("parse-kv source kind = %v", parsePipe.Expression().Kind())
	}
}

func TestParseQueryBuildsParseKvMissingSchema(t *testing.T) {
	const text = "T | parse-kv Message as"
	expr := parsedQueryExpression(t, text)

	parsePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse-kv pipe = %#v", expr)
	}
	parseKv, ok := parsePipe.Operator().(*ParseKvOperator)
	if !ok {
		t.Fatalf("parse-kv operator = %#v", parsePipe.Operator())
	}
	if parseKv.Keys() == nil {
		t.Fatal("missing parse-kv keys")
	}
	if len(parseKv.Keys().Diagnostics()) == 0 {
		t.Fatal("parse-kv missing keys should carry a missing-schema diagnostic")
	}
}

func TestParseQueryBuildsParseOperator(t *testing.T) {
	const text = `T | parse kind=regex flags=U Message with * "User=" Account:string " Score=" Score:long`
	expr := parsedQueryExpression(t, text)

	parsePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse pipe = %#v", expr)
	}
	parseOp, ok := parsePipe.Operator().(*ParseOperator)
	if !ok {
		t.Fatalf("parse operator = %#v", parsePipe.Operator())
	}
	if parseOp.Parameters().Count() != 2 {
		t.Fatalf("parse parameters = %d", parseOp.Parameters().Count())
	}
	if got := namedParameterNameText(t, parseOp.Parameters().At(0)); got != "kind" {
		t.Fatalf("parse parameter 0 = %q", got)
	}
	if parseOp.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("parse kind expression = %v", parseOp.Parameters().At(0).Expression().Kind())
	}
	if got := namedParameterNameText(t, parseOp.Parameters().At(1)); got != "flags" {
		t.Fatalf("parse parameter 1 = %q", got)
	}
	if parseOp.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("parse expression kind = %v", parseOp.Expression().Kind())
	}
	if parseOp.WithKeyword().Kind() != SyntaxKindWithKeyword {
		t.Fatalf("parse with keyword = %v", parseOp.WithKeyword().Kind())
	}
	if parseOp.Patterns().Count() != 5 {
		t.Fatalf("parse patterns = %d", parseOp.Patterns().Count())
	}
	if parseOp.Patterns().At(0).Kind() != SyntaxKindStarExpression {
		t.Fatalf("parse pattern 0 = %v", parseOp.Patterns().At(0).Kind())
	}
	if parseOp.Patterns().At(1).Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("parse pattern 1 = %v", parseOp.Patterns().At(1).Kind())
	}
	if parseOp.Patterns().At(2).Kind() != SyntaxKindNameAndTypeDeclaration {
		t.Fatalf("parse pattern 2 = %v", parseOp.Patterns().At(2).Kind())
	}
	if parseOp.Patterns().At(4).Kind() != SyntaxKindNameAndTypeDeclaration {
		t.Fatalf("parse pattern 4 = %v", parseOp.Patterns().At(4).Kind())
	}
	if parsePipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("parse source kind = %v", parsePipe.Expression().Kind())
	}
}

func TestParseQueryBuildsParseWhereOperator(t *testing.T) {
	const text = `T | parse-where Message with "ip=" IP:string`
	expr := parsedQueryExpression(t, text)

	parsePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse-where pipe = %#v", expr)
	}
	parseWhere, ok := parsePipe.Operator().(*ParseWhereOperator)
	if !ok {
		t.Fatalf("parse-where operator = %#v", parsePipe.Operator())
	}
	if parseWhere.Parameters().Count() != 0 {
		t.Fatalf("parse-where parameters = %d", parseWhere.Parameters().Count())
	}
	if parseWhere.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("parse-where expression kind = %v", parseWhere.Expression().Kind())
	}
	if parseWhere.WithKeyword().Kind() != SyntaxKindWithKeyword {
		t.Fatalf("parse-where with keyword = %v", parseWhere.WithKeyword().Kind())
	}
	if parseWhere.Patterns().Count() != 2 {
		t.Fatalf("parse-where patterns = %d", parseWhere.Patterns().Count())
	}
	if parseWhere.Patterns().At(0).Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("parse-where pattern 0 = %v", parseWhere.Patterns().At(0).Kind())
	}
	if parseWhere.Patterns().At(1).Kind() != SyntaxKindNameAndTypeDeclaration {
		t.Fatalf("parse-where pattern 1 = %v", parseWhere.Patterns().At(1).Kind())
	}
}

func TestParseQueryBuildsMakeSeriesInRangeOperator(t *testing.T) {
	const text = "T | make-series Count=count() on Timestamp in range(ago(1d), now(), 1h)"
	expr := parsedQueryExpression(t, text)

	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	if makeSeries.Aggregates().Count() != 1 {
		t.Fatalf("make-series aggregates = %d", makeSeries.Aggregates().Count())
	}
	if makeSeries.Aggregates().At(0).Element().DefaultExpression() != nil {
		t.Fatal("unexpected make-series default clause")
	}
	rangeClause, ok := makeSeries.RangeClause().(*MakeSeriesInRangeClause)
	if !ok {
		t.Fatalf("make-series range clause = %#v", makeSeries.RangeClause())
	}
	if rangeClause.Arguments().Expressions().Count() != 3 {
		t.Fatalf("make-series range arguments = %d", rangeClause.Arguments().Expressions().Count())
	}
	if makeSeries.ByClause() != nil {
		t.Fatalf("unexpected make-series by clause = %#v", makeSeries.ByClause())
	}
}

func TestParseQueryBuildsTopHittersByOperator(t *testing.T) {
	const text = "T | top-hitters 10 of UserPrincipalName by DurationMs"
	expr := parsedQueryExpression(t, text)

	topHittersPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-hitters pipe = %#v", expr)
	}
	topHitters, ok := topHittersPipe.Operator().(*TopHittersOperator)
	if !ok {
		t.Fatalf("top-hitters operator = %#v", topHittersPipe.Operator())
	}
	if topHitters.ByClause() == nil {
		t.Fatal("missing top-hitters by clause")
	}
	if topHitters.ByClause().Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("top-hitters by expression = %v", topHitters.ByClause().Expression().Kind())
	}
}

func TestParseQueryBuildsReduceByOperator(t *testing.T) {
	const text = "T | reduce kind=source by Message with threshold=0.5, characters='abc'"
	expr := parsedQueryExpression(t, text)

	reducePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", expr)
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	if reduce.ReduceKeyword().Kind() != SyntaxKindReduceKeyword {
		t.Fatalf("reduce keyword = %v", reduce.ReduceKeyword().Kind())
	}
	if reduce.Parameters().Count() != 1 {
		t.Fatalf("reduce parameters = %d", reduce.Parameters().Count())
	}
	if got := namedParameterNameText(t, reduce.Parameters().At(0)); got != "kind" {
		t.Fatalf("reduce parameter 0 = %q", got)
	}
	if reduce.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("reduce by keyword = %v", reduce.ByKeyword().Kind())
	}
	if reduce.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("reduce expression = %v", reduce.Expression().Kind())
	}
	if reduce.With() == nil {
		t.Fatal("missing reduce with clause")
	}
	if reduce.With().Parameters().Count() != 2 {
		t.Fatalf("reduce with parameters = %d", reduce.With().Parameters().Count())
	}
	if got := namedParameterNameText(t, reduce.With().Parameters().At(0).Element()); got != "threshold" {
		t.Fatalf("reduce with parameter 0 = %q", got)
	}
	if got := namedParameterNameText(t, reduce.With().Parameters().At(1).Element()); got != "characters" {
		t.Fatalf("reduce with parameter 1 = %q", got)
	}
}

func TestParseDataScopeInvalidValueCarriesDiagnostic(t *testing.T) {
	expr := parsedQueryExpression(t, `search datascope=coldcache in (SigninLogs) UserPrincipalName has "admin"`)
	search, ok := expr.(*SearchOperator)
	if !ok {
		t.Fatalf("search expression = %#v", expr)
	}
	if search.DataScope() == nil {
		t.Fatal("missing search datascope")
	}
	if got := search.DataScope().Value().RawText(); got != "coldcache" {
		t.Fatalf("datascope value = %q", got)
	}
	if !diagnosticsContain(search.DataScope().ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected invalid datascope diagnostic, got %#v", search.DataScope().ContainedDiagnostics())
	}
}

func TestParseDistinctTrailingCommaReportsOnlyMissingExpression(t *testing.T) {
	query := ParseQuery("datatable(x:long)[1] | distinct x,")
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected trailing distinct comma missing-expression diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("trailing distinct comma should not add missing-token diagnostics, got %#v", diagnostics)
	}
}

func TestParseProjectByNamesTrailingCommaReportsMissingExpression(t *testing.T) {
	expr := parsedQueryExpression(t, `T | project-by-names dynamic(["x"]),`)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project-by-names pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectByNamesOperator)
	if !ok {
		t.Fatalf("project-by-names operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 2 {
		t.Fatalf("project-by-names expressions = %d", project.Expressions().Count())
	}
	missing := project.Expressions().At(1).Element()
	if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("project-by-names trailing comma diagnostics = %#v", missing.ContainedDiagnostics())
	}
	if diagnosticsContain(expr.ContainedDiagnostics(), "KS001") {
		t.Fatalf("project-by-names trailing comma should not add missing-token diagnostics, got %#v", expr.ContainedDiagnostics())
	}
}

func TestParseProjectByNamesAllowsEmptyExpressionList(t *testing.T) {
	expr := parsedQueryExpression(t, `T | project-by-names`)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project-by-names pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectByNamesOperator)
	if !ok {
		t.Fatalf("project-by-names operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 0 {
		t.Fatalf("project-by-names expressions = %d", project.Expressions().Count())
	}
	if diagnostics := expr.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("empty project-by-names diagnostics = %#v", diagnostics)
	}
}

func TestParseProjectByNamesLeadingCommaRetainsMissingExpressions(t *testing.T) {
	expr := parsedQueryExpression(t, `T | project-by-names,`)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project-by-names pipe = %#v", expr)
	}
	project, ok := projectPipe.Operator().(*ProjectByNamesOperator)
	if !ok {
		t.Fatalf("project-by-names operator = %#v", projectPipe.Operator())
	}
	if project.Expressions().Count() != 2 {
		t.Fatalf("project-by-names expressions = %d", project.Expressions().Count())
	}
	for i := 0; i < project.Expressions().Count(); i++ {
		missing := project.Expressions().At(i).Element()
		if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing expression") {
			t.Fatalf("project-by-names expression %d diagnostics = %#v", i, missing.ContainedDiagnostics())
		}
	}
}

func TestParseBareProjectOperatorsAllowOptionalLists(t *testing.T) {
	tests := []struct {
		text string
		want SyntaxKind
	}{
		{text: "T | project", want: SyntaxKindProjectOperator},
		{text: "T | project-away", want: SyntaxKindProjectAwayOperator},
		{text: "T | project-keep", want: SyntaxKindProjectKeepOperator},
		{text: "T | project-rename", want: SyntaxKindProjectRenameOperator},
		{text: "T | project-reorder", want: SyntaxKindProjectReorderOperator},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			query := ParseQuery(tt.text)
			diagnostics := query.ContainedDiagnostics()
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics = %#v", diagnostics)
			}
			expr := parsedQueryExpression(t, tt.text)
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("pipe expression = %#v", expr)
			}
			if pipe.Operator().Kind() != tt.want {
				t.Fatalf("operator kind = %v, want %v", pipe.Operator().Kind(), tt.want)
			}
			switch op := pipe.Operator().(type) {
			case *ProjectOperator:
				if op.Expressions().Count() != 0 {
					t.Fatalf("project expressions = %d", op.Expressions().Count())
				}
			case *ProjectAwayOperator:
				if op.Expressions().Count() != 0 {
					t.Fatalf("project-away expressions = %d", op.Expressions().Count())
				}
			case *ProjectKeepOperator:
				if op.Expressions().Count() != 0 {
					t.Fatalf("project-keep expressions = %d", op.Expressions().Count())
				}
			case *ProjectRenameOperator:
				if op.Expressions().Count() != 0 {
					t.Fatalf("project-rename expressions = %d", op.Expressions().Count())
				}
			case *ProjectReorderOperator:
				if op.Expressions().Count() != 0 {
					t.Fatalf("project-reorder expressions = %d", op.Expressions().Count())
				}
			default:
				t.Fatalf("operator = %#v", op)
			}
		})
	}
}

func TestParseProjectPackStarExpression(t *testing.T) {
	query := ParseQuery("T | project p = pack(*)")
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected project pack diagnostics: %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, "T | project p = pack(*)")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	project, ok := pipe.Operator().(*ProjectOperator)
	if !ok {
		t.Fatalf("project operator = %#v", pipe.Operator())
	}
	if project.Expressions().Count() != 1 {
		t.Fatalf("project expressions = %d", project.Expressions().Count())
	}
	named, ok := project.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("project expression = %#v", project.Expressions().At(0).Element())
	}
	if named.Expression().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("project pack expression kind = %v", named.Expression().Kind())
	}
}

func TestParseQueryProjectRecoveryReportsMissingAndSkippedFragments(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCodes []string
	}{
		{
			name:      "dangling project comma",
			text:      "T | project Account,",
			wantCodes: []string{"KS006"},
		},
		{
			name:      "dangling project-rename comma",
			text:      "T | project-rename Account=User,",
			wantCodes: []string{"KS006"},
		},
		{
			name:      "dangling project-reorder comma",
			text:      "T | project-reorder Account,",
			wantCodes: []string{"KS006"},
		},
		{
			name:      "unparsed project fragment",
			text:      "T | project =",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "unparsed less-than expression fragment",
			text:      "print x=<",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "unparsed less-than-or-equal expression fragment",
			text:      "print x=<=",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "unparsed greater-than expression fragment",
			text:      "T | where >",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "unparsed greater-than-or-equal expression fragment",
			text:      "T | where >=",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "unparsed trailing question fragment",
			text:      "print x=1 ?",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "unparsed render deprecated property comma tail",
			text:      `T | render timechart , title="a"`,
			wantCodes: []string{"KS198"},
		},
		{
			name:      "unparsed render with-clause comma tail",
			text:      `T | render timechart with (xcolumn=Timestamp), Extra`,
			wantCodes: []string{"KS198"},
		},
		{
			name:      "malformed braced client parameter",
			text:      "print x={1}",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed dot named expression",
			text:      "print x=.",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed slash named expression",
			text:      "print x=/",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed percent named expression",
			text:      "print x=%",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed star named expression",
			text:      "print x=*",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed and keyword named expression",
			text:      "print x=and",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed contains keyword named expression",
			text:      "print x=contains",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed leading between predicate",
			text:      "T | where between",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed leading in predicate",
			text:      "T | where in",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "malformed between comma couple",
			text:      "T | where Count between (1, 2)",
			wantCodes: []string{"KS005", "KS006", "KS198"},
		},
		{
			name:      "malformed between trailing comma couple",
			text:      "T | where Count between (1..2,)",
			wantCodes: []string{"KS005", "KS198"},
		},
		{
			name:      "malformed not-between comma couple",
			text:      "T | where Count !between (1, 2)",
			wantCodes: []string{"KS005", "KS006", "KS198"},
		},
		{
			name:      "malformed not-between trailing comma couple",
			text:      "T | where Count !between (1..2,)",
			wantCodes: []string{"KS005", "KS198"},
		},
		{
			name:      "unparsed parse pattern tail",
			text:      `T | parse Raw with "x" Domain:string,`,
			wantCodes: []string{"KS198"},
		},
		{
			name:      "unparsed parse comma pattern tail",
			text:      `T | parse Raw with "x" , Domain:string`,
			wantCodes: []string{"KS198"},
		},
		{
			name:      "unparsed parse-where comma pattern tail",
			text:      `T | parse-where Raw with "x" , Domain:string`,
			wantCodes: []string{"KS198"},
		},
		{
			name:      "take unexpected by tail",
			text:      "T | take 10 by User",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "take extra comma expression",
			text:      "T | take 1, 2",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "take extra expression",
			text:      "T | take 1 x",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "limit extra comma expression",
			text:      "T | limit 1, 2",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "limit extra expression",
			text:      "T | limit 1 x",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "take unexpected close paren tail",
			text:      "T | take 10 )",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "take unexpected close bracket tail",
			text:      "T | take 10 ]",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "take unexpected close brace tail",
			text:      "T | take 10 }",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "where unexpected close paren tail",
			text:      "T | where Score == 1 )",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "where extra comma predicate tail",
			text:      "T | where Score == 1, Other",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "join where extra comma predicate tail",
			text:      "T | join (R) where Score == 1, Other",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "source extra comma tail before pipe",
			text:      "Source, Missing | project x",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "filter extra comma predicate tail",
			text:      "T | filter Score == 1, Other",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "where unexpected bang tail",
			text:      "T | where !Score",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "misplaced directive tail",
			text:      "T | count\n#set query_now=datetime(2024-01-01)",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "count unexpected by tail",
			text:      "T | count by User",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "count extra comma tail",
			text:      "T | count, Extra",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "assert-schema extra expression tail",
			text:      "T | assert-schema (A:long) Extra",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "consume extra expression tail",
			text:      "T | consume Extra",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "execute-and-cache extra expression tail",
			text:      "T | __executeAndCache Extra",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "getschema extra expression tail",
			text:      "T | getschema Extra",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "project unexpected by tail",
			text:      "T | project User by App",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "distinct hint strategy missing equals",
			text:      "T | distinct hint.strategy shuffle Account",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "sample unexpected of tail",
			text:      "T | sample 5 of User",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "sample extra expression",
			text:      "T | sample 5 User",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "sample-distinct unsupported with tail",
			text:      "T | sample-distinct 5 of x with others=y",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "sort unexpected leading expression before by",
			text:      "T | sort User by App",
			wantCodes: []string{"KS005", "KS198"},
		},
		{
			name:      "sort ordering modifier without expression",
			text:      "T | sort by asc",
			wantCodes: []string{"KS006", "KS198"},
		},
		{
			name:      "top extra comma expression",
			text:      "T | top 5 by x, y",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "top extra ordered expression",
			text:      "T | top 5 by x desc y asc",
			wantCodes: []string{"KS005", "KS198"},
		},
		{
			name:      "top unsupported with tail",
			text:      "T | top 5 by x with others=y",
			wantCodes: []string{"KS005", "KS198"},
		},
		{
			name:      "top-hitters extra comma by expression",
			text:      "T | top-hitters 5 of x by y, z",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "top-hitters extra by expression",
			text:      "T | top-hitters 5 of x by y z",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "top-hitters extra of expression",
			text:      "T | top-hitters 5 of x y",
			wantCodes: []string{"KS005"},
		},
		{
			name:      "top-hitters extra of expression before by",
			text:      "T | top-hitters 5 of x y by z",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "top-hitters extra comma of expression before by",
			text:      "T | top-hitters 5 of x, y by z",
			wantCodes: []string{"KS198"},
		},
		{
			name:      "fork comma before branch",
			text:      `T | fork (where x == 1), ()`,
			wantCodes: []string{"KS198"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.text)
			if code == nil || code.Query == nil {
				t.Fatalf("missing query parse result")
			}
			diagnostics := code.GetDiagnostics()
			for _, code := range tt.wantCodes {
				if !diagnosticsContain(diagnostics, code) {
					t.Fatalf("expected %s for %q, got %#v", code, tt.text, diagnostics)
				}
			}
		})
	}
}

func TestParseQueryBuildsCompactPipedOperators(t *testing.T) {
	const text = "range i from 1 to 3 step 1 | serialize doubled=i + i | as hint.materialized=true Rows | getschema kind=csl"
	expr := parsedQueryExpression(t, text)

	getSchemaPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("getschema pipe = %#v", expr)
	}
	getSchema, ok := getSchemaPipe.Operator().(*GetSchemaOperator)
	if !ok {
		t.Fatalf("getschema operator = %#v", getSchemaPipe.Operator())
	}
	if getSchema.KindParameter() == nil {
		t.Fatal("missing getschema kind parameter")
	}
	if got := namedParameterNameText(t, getSchema.KindParameter()); got != "kind" {
		t.Fatalf("getschema kind parameter name = %q", got)
	}
	if getSchema.KindParameter().Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("getschema kind expression = %v", getSchema.KindParameter().Expression().Kind())
	}

	missingEquals := ParseQuery("range i from 1 to 3 step 1 | getschema kind csl")
	missingEqualsCount := 0
	for _, diagnostic := range missingEquals.ContainedDiagnostics() {
		if diagnostic.Code == "KS005" {
			missingEqualsCount++
		}
	}
	if missingEqualsCount != 2 {
		t.Fatalf("getschema missing-equals diagnostics = %#v", missingEquals.ContainedDiagnostics())
	}
	missingEqualsPipe, ok := firstParsedQueryExpression(t, "range i from 1 to 3 step 1 | getschema kind csl").(*PipeExpression)
	if !ok {
		t.Fatalf("getschema missing-equals pipe = %#v", missingEqualsPipe)
	}
	missingEqualsGetSchema, ok := missingEqualsPipe.Operator().(*GetSchemaOperator)
	if !ok {
		t.Fatalf("getschema missing-equals operator = %#v", missingEqualsPipe.Operator())
	}
	if missingEqualsGetSchema.KindParameter() != nil {
		t.Fatalf("getschema missing-equals kind parameter = %#v, want nil", missingEqualsGetSchema.KindParameter())
	}

	trailing := ParseQuery("range i from 1 to 3 step 1 | getschema kind=csl, Extra")
	if !diagnosticsContain(trailing.ContainedDiagnostics(), "KS198") {
		t.Fatalf("getschema comma-tail diagnostics = %#v", trailing.ContainedDiagnostics())
	}
	trailingPipe, ok := parsedQueryExpression(t, "range i from 1 to 3 step 1 | getschema kind=csl, Extra").(*PipeExpression)
	if !ok {
		t.Fatalf("getschema comma-tail pipe = %#v", trailingPipe)
	}
	trailingGetSchema, ok := trailingPipe.Operator().(*GetSchemaOperator)
	if !ok {
		t.Fatalf("getschema comma-tail operator = %#v", trailingPipe.Operator())
	}
	if trailingGetSchema.KindParameter() == nil || namedParameterNameText(t, trailingGetSchema.KindParameter()) != "kind" {
		t.Fatalf("getschema comma-tail kind parameter = %#v", trailingGetSchema.KindParameter())
	}

	asPipe, ok := getSchemaPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("as pipe = %#v", getSchemaPipe.Expression())
	}
	asOperator, ok := asPipe.Operator().(*AsOperator)
	if !ok {
		t.Fatalf("as operator = %#v", asPipe.Operator())
	}
	if asOperator.Parameters().Count() != 1 {
		t.Fatalf("as parameters = %d", asOperator.Parameters().Count())
	}
	if got := namedParameterNameText(t, asOperator.Parameters().At(0)); got != "hint.materialized" {
		t.Fatalf("as parameter name = %q", got)
	}
	asName, ok := asOperator.Name().Name().(*TokenName)
	if !ok {
		t.Fatalf("as name = %#v", asOperator.Name().Name())
	}
	if asName.Name().RawText() != "Rows" {
		t.Fatalf("as name text = %q", asName.Name().RawText())
	}

	serializePipe, ok := asPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("serialize pipe = %#v", asPipe.Expression())
	}
	serializeOperator, ok := serializePipe.Operator().(*SerializeOperator)
	if !ok {
		t.Fatalf("serialize operator = %#v", serializePipe.Operator())
	}
	if serializeOperator.Expressions().Count() != 1 {
		t.Fatalf("serialize expressions = %d", serializeOperator.Expressions().Count())
	}
	if serializeOperator.Expressions().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("serialize expression kind = %v", serializeOperator.Expressions().At(0).Element().Kind())
	}

	rangeOperator, ok := serializePipe.Expression().(*RangeOperator)
	if !ok {
		t.Fatalf("range operator = %#v", serializePipe.Expression())
	}
	rangeName, ok := rangeOperator.Name().Name().(*TokenName)
	if !ok {
		t.Fatalf("range name = %#v", rangeOperator.Name().Name())
	}
	if rangeName.Name().RawText() != "i" {
		t.Fatalf("range name text = %q", rangeName.Name().RawText())
	}
	if rangeOperator.From().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range from kind = %v", rangeOperator.From().Kind())
	}
	if rangeOperator.To().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range to kind = %v", rangeOperator.To().Kind())
	}
	if rangeOperator.Step().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range step kind = %v", rangeOperator.Step().Kind())
	}
}

func TestParseRangeMissingNameBeforeFrom(t *testing.T) {
	const text = "range from 1 to 3 step 1"
	parser := newQueryParserState(Lex(text))
	rangeOperator := parser.parseRangeOperator()
	if rangeOperator == nil {
		t.Fatal("missing range operator")
	}
	if !parser.atEndOfQuery() {
		t.Fatalf("range parser did not consume full operator; next token = %v", parser.peekToken().Kind)
	}
	if !diagnosticsContainMessage(rangeOperator.Name().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("expected range missing-name diagnostic, got %#v", rangeOperator.Name().ContainedDiagnostics())
	}
	if rangeOperator.FromToken().Kind() != SyntaxKindFromKeyword {
		t.Fatalf("range from token = %v", rangeOperator.FromToken().Kind())
	}
	if rangeOperator.From().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.To().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.Step().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range expressions = %v, %v, %v", rangeOperator.From().Kind(), rangeOperator.To().Kind(), rangeOperator.Step().Kind())
	}
}

func TestParseRangeMissingToAndStepTokensRecoverExpressions(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		missing string
	}{
		{
			name:    "missing from",
			text:    "range x 1 to 3 step 1",
			missing: "Expected: from",
		},
		{
			name:    "missing to",
			text:    "range x from 1 3 step 1",
			missing: "Expected: to",
		},
		{
			name:    "missing step",
			text:    "range x from 1 to 3 1",
			missing: "Expected: step",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newQueryParserState(Lex(tt.text))
			rangeOperator := parser.parseRangeOperator()
			if rangeOperator == nil {
				t.Fatal("missing range operator")
			}
			if !parser.atEndOfQuery() {
				t.Fatalf("range parser did not consume full operator; next token = %v", parser.peekToken().Kind)
			}
			if !diagnosticsContainMessage(rangeOperator.ContainedDiagnostics(), "KS005", tt.missing) {
				t.Fatalf("expected %q diagnostic, got %#v", tt.missing, rangeOperator.ContainedDiagnostics())
			}
			if rangeOperator.From().Kind() != SyntaxKindLongLiteralExpression ||
				rangeOperator.To().Kind() != SyntaxKindLongLiteralExpression ||
				rangeOperator.Step().Kind() != SyntaxKindLongLiteralExpression {
				t.Fatalf("range expressions = %v, %v, %v", rangeOperator.From().Kind(), rangeOperator.To().Kind(), rangeOperator.Step().Kind())
			}
		})
	}
}

func TestParseRangeExtraFromExpressionReportsIncompleteFragment(t *testing.T) {
	code := Parse("range x from 1 2 to 10 step 1")
	want := []Diagnostic{
		DiagnosticTermsExpected("to").WithLocation(15, 1),
		DiagnosticTermsExpected("step").WithLocation(17, 2),
		DiagnosticMissingExpressionWithKeyword("to").WithLocation(16, 0),
		DiagnosticIncompleteFragment().WithLocation(17, 2),
	}
	got := code.GetDiagnostics()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range got {
		if got[i].Code != want[i].Code || got[i].Message != want[i].Message || got[i].Start != want[i].Start || got[i].Length != want[i].Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestParseRangeStepCommaTailReportsIncompleteFragment(t *testing.T) {
	const text = "range x from 1 to 10 step 1, Extra"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected range comma-tail diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("range comma-tail should not add missing-token diagnostics, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	rangeOperator, ok := expr.(*RangeOperator)
	if !ok {
		t.Fatalf("range operator = %#v", expr)
	}
	rangeName, ok := rangeOperator.Name().Name().(*TokenName)
	if !ok {
		t.Fatalf("range name = %#v", rangeOperator.Name().Name())
	}
	if rangeName.Name().RawText() != "x" {
		t.Fatalf("range name text = %q", rangeName.Name().RawText())
	}
	if rangeOperator.From().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.To().Kind() != SyntaxKindLongLiteralExpression ||
		rangeOperator.Step().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("range expressions = %v, %v, %v", rangeOperator.From().Kind(), rangeOperator.To().Kind(), rangeOperator.Step().Kind())
	}
}

func TestParseCurrentRangeElementRecoveryReferenceParity(t *testing.T) {
	const text = `range ['print' from 1 to 3 step 1`
	code := Parse(text).Analyze()
	wantDiagnostics := []Diagnostic{
		DiagnosticFunctionRequiresArgumentList("range").WithLocation(0, 5),
		DiagnosticTermsExpected("]").WithLocation(15, 4),
		DiagnosticTermsExpected(";").WithLocation(15, 4),
		NewDiagnostic("KS142", "The name 'from' does not refer to any known column, table, variable or function.").WithLocation(15, 4),
		DiagnosticTermsExpected(";").WithLocation(20, 1),
		DiagnosticIncompleteFragment().WithLocation(22, 2),
	}
	gotDiagnostics := code.GetDiagnostics()
	if len(gotDiagnostics) != len(wantDiagnostics) {
		t.Fatalf("diagnostics = %#v, want %#v", gotDiagnostics, wantDiagnostics)
	}
	for i := range wantDiagnostics {
		got, want := gotDiagnostics[i], wantDiagnostics[i]
		if got.Code != want.Code || got.Message != want.Message || got.LocationKind != want.LocationKind ||
			got.Start != want.Start || got.Length != want.Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got, want)
		}
	}

	statements := code.Query.Statements()
	if statements.Count() != 3 {
		t.Fatalf("statement count = %d, want 3", statements.Count())
	}
	first, ok := statements.At(0).Element().(*ExpressionStatement)
	if !ok {
		t.Fatalf("first statement = %T, want *ExpressionStatement", statements.At(0).Element())
	}
	element, ok := first.Expression().(*ElementExpression)
	if !ok || element.Kind() == SyntaxKindNone {
		t.Fatalf("first expression = %#v, want recovered *ElementExpression", first.Expression())
	}
	rangeName, ok := element.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(rangeName) != "range" {
		t.Fatalf("element base = %#v, want range name", element.Expression())
	}
	selector, ok := element.Selector().(*BracketedExpression)
	if !ok || selector.CloseBracket() == nil || selector.CloseBracket().RawText() != "" {
		t.Fatalf("element selector = %#v, want missing close bracket", element.Selector())
	}
	if got := statements.At(0).Separator(); got == nil || got.Kind() != SyntaxKindSemicolonToken || got.RawText() != "" {
		t.Fatalf("first separator = %#v, want missing semicolon", got)
	}

	second, ok := statements.At(1).Element().(*ExpressionStatement)
	if !ok {
		t.Fatalf("second statement = %T, want *ExpressionStatement", statements.At(1).Element())
	}
	fromName, ok := second.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(fromName) != "from" {
		t.Fatalf("second expression = %#v, want from name", second.Expression())
	}
	if got := statements.At(1).Separator(); got == nil || got.Kind() != SyntaxKindSemicolonToken || got.RawText() != "" {
		t.Fatalf("second separator = %#v, want missing semicolon", got)
	}
	third, ok := statements.At(2).Element().(*ExpressionStatement)
	if !ok || third.Expression().Kind() != SyntaxKindLongLiteralExpression || statements.At(2).Separator() != nil {
		t.Fatalf("third statement = %#v separator=%#v, want terminal long literal", statements.At(2).Element(), statements.At(2).Separator())
	}

	skipped := code.Query.SkippedTokens()
	if skipped == nil || skipped.Tokens().Count() != 4 || skipped.Tokens().At(0).Kind() != SyntaxKindToKeyword {
		t.Fatalf("skipped tokens = %#v, want to/3/step/1", skipped)
	}
	if ranges := boundNodes[*RangeOperator](code.Query); len(ranges) != 0 {
		t.Fatalf("recovery created range operators = %#v", ranges)
	}
	if table := terminalResultTable(code); table != nil && len(table.Columns()) != 0 {
		t.Fatalf("terminal result = %#v, want empty tabular output", table.Columns())
	}
}

func TestParseCurrentRangeElementRecoveryControls(t *testing.T) {
	t.Run("valid bracketed range", func(t *testing.T) {
		code := Parse(`range ['print'] from 1 to 3 step 1`).Analyze()
		assertDiagnosticLocations(t, code.GetDiagnostics())
		operator := firstBoundNode[*RangeOperator](t, code.Query)
		column := semanticTable(t, operator).GetColumn("print")
		if column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
			t.Fatalf("range column = %#v, want print:long", column)
		}
	})

	tests := []struct {
		name string
		text string
	}{
		{name: "closed selector missing from", text: `range ['print'] 1 to 3 step 1`},
		{name: "different skipped boundary", text: `range ['print' from 1 step 1`},
		{name: "neighboring root", text: `print ['range' from 1 to 3 step 1`},
		{name: "explicit separators", text: `range ['print'; from; 1 to 3 step 1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newQueryParserState(Lex(tt.text))
			statements := parser.parseQueryBlockStatementList()
			position := parser.pos
			if parser.shouldPreserveCurrentRangeElementRecovery(statements) {
				t.Fatalf("neighbor %q matched range element recovery", tt.text)
			}
			if parser.pos != position {
				t.Fatalf("recovery probe moved parser from %d to %d", position, parser.pos)
			}
		})
	}
}

func TestParseQueryMissingNameWithKeywordDiagnostic(t *testing.T) {
	query := ParseQuery("T | as project")
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() == 0 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(first.Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	expr := statement.Expression()
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %#v", expr)
	}
	asOperator, ok := pipe.Operator().(*AsOperator)
	if !ok {
		t.Fatalf("as operator = %#v", pipe.Operator())
	}
	const want = "Missing name: If the keyword 'project' is intended be used as the name, it needs to be bracketted as ['project']."
	if !diagnosticsContainMessage(asOperator.Name().Diagnostics(), "KS006", want) {
		t.Fatalf("expected missing-name keyword diagnostic, got %#v", asOperator.Name().Diagnostics())
	}
}

func TestParseAsOperatorRequiresNonExtendedName(t *testing.T) {
	t.Run("plain identifier", func(t *testing.T) {
		expr := parsedQueryExpression(t, `T | as Alias`)
		pipe, ok := expr.(*PipeExpression)
		if !ok {
			t.Fatalf("pipe expression = %#v", expr)
		}
		asOperator, ok := pipe.Operator().(*AsOperator)
		if !ok {
			t.Fatalf("as operator = %#v", pipe.Operator())
		}
		if got := NameDeclarationSimpleName(asOperator.Name()); got != "Alias" {
			t.Fatalf("as name = %q", got)
		}
		if len(asOperator.Name().ContainedDiagnostics()) != 0 {
			t.Fatalf("unexpected plain as-name diagnostics: %#v", asOperator.Name().ContainedDiagnostics())
		}
	})

	t.Run("bracketed keyword", func(t *testing.T) {
		expr := parsedQueryExpression(t, `T | as ['where']`)
		pipe, ok := expr.(*PipeExpression)
		if !ok {
			t.Fatalf("pipe expression = %#v", expr)
		}
		asOperator, ok := pipe.Operator().(*AsOperator)
		if !ok {
			t.Fatalf("as operator = %#v", pipe.Operator())
		}
		if got := NameDeclarationSimpleName(asOperator.Name()); got != "where" {
			t.Fatalf("as bracketed name = %q", got)
		}
		if len(asOperator.Name().ContainedDiagnostics()) != 0 {
			t.Fatalf("unexpected bracketed as-name diagnostics: %#v", asOperator.Name().ContainedDiagnostics())
		}
	})

	for _, keyword := range []string{"where", "by"} {
		t.Run("unbracketed "+keyword, func(t *testing.T) {
			query := ParseQuery(`T | as ` + keyword)
			statements, ok := query.GetChild(1).(*SyntaxList)
			if !ok || statements.Count() == 0 {
				t.Fatalf("statements = %#v", query.GetChild(1))
			}
			first, ok := statements.At(0).(*SeparatedElement)
			if !ok {
				t.Fatalf("first statement element = %#v", statements.At(0))
			}
			statement, ok := WrapGeneratedSyntaxElement(first.Element()).(*ExpressionStatement)
			if !ok {
				t.Fatalf("first statement = %#v", first.Element())
			}
			expr := statement.Expression()
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("pipe expression = %#v", expr)
			}
			asOperator, ok := pipe.Operator().(*AsOperator)
			if !ok {
				t.Fatalf("as operator = %#v", pipe.Operator())
			}
			want := "Missing name: If the keyword '" + keyword + "' is intended be used as the name, it needs to be bracketted as ['" + keyword + "']."
			if !diagnosticsContainMessage(asOperator.Name().Diagnostics(), "KS006", want) {
				t.Fatalf("expected missing-name keyword diagnostic for %s, got %#v", keyword, asOperator.Name().Diagnostics())
			}
		})
	}
}

func TestParseQueryMissingAsNameOmitsTokenDiagnostic(t *testing.T) {
	expr := parsedQueryExpression(t, "T | as")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("pipe expression = %#v", expr)
	}
	asOperator, ok := pipe.Operator().(*AsOperator)
	if !ok {
		t.Fatalf("as operator = %#v", pipe.Operator())
	}
	diagnostics := asOperator.Name().ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing name") {
		t.Fatalf("expected missing-name diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(expr.ContainedDiagnostics(), "KS001") {
		t.Fatalf("missing as name should not add missing-token diagnostics, got %#v", expr.ContainedDiagnostics())
	}
}

func TestParseAsOperatorReportsUnexpectedAliasTail(t *testing.T) {
	tests := []struct {
		text string
		code string
	}{
		{text: `T | as Alias Extra | project X`, code: "KS005"},
		{text: `T | as ['Alias'] Extra | project X`, code: "KS005"},
		{text: `T | as Alias, Extra | project X`, code: "KS198"},
		{text: `T | as ['Alias'], Extra | project X`, code: "KS198"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			query := ParseQuery(tt.text)
			if !diagnosticsContain(query.ContainedDiagnostics(), tt.code) {
				t.Fatalf("expected %s tail diagnostic, got %#v", tt.code, query.ContainedDiagnostics())
			}
			if tt.code == "KS005" && diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
				t.Fatalf("whitespace alias tail should become a statement, got %#v", query.ContainedDiagnostics())
			}
		})
	}
}

func TestParseAsMissingParameterEqualsStopsBeforeParameter(t *testing.T) {
	const text = `T | as hint.materialized maybe Rows | project A`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS006") || !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected missing-name and skipped-tail diagnostics, got %#v", query.ContainedDiagnostics())
	}
	asPipe, ok := firstParsedQueryExpression(t, text).(*PipeExpression)
	if !ok {
		t.Fatalf("as pipe = %#v", firstParsedQueryExpression(t, text))
	}
	asOperator, ok := asPipe.Operator().(*AsOperator)
	if !ok {
		t.Fatalf("as operator = %#v", asPipe.Operator())
	}
	if asOperator.Parameters().Count() != 0 {
		t.Fatalf("as parameters = %d", asOperator.Parameters().Count())
	}
	if !diagnosticsContain(asOperator.Name().ContainedDiagnostics(), "KS006") {
		t.Fatalf("as alias diagnostics = %#v", asOperator.Name().ContainedDiagnostics())
	}
	if query.GetChild(2) == nil {
		t.Fatalf("missing-equals tail should remain QueryBlock skipped tokens")
	}
}

func TestParseQueryBuildsInvokeAndConsumeOperators(t *testing.T) {
	const text = "T | invoke Transform(1) | consume decodeblocks=true"
	expr := parsedQueryExpression(t, text)

	consumePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("consume pipe = %#v", expr)
	}
	consume, ok := consumePipe.Operator().(*ConsumeOperator)
	if !ok {
		t.Fatalf("consume operator = %#v", consumePipe.Operator())
	}
	if consume.Parameters().Count() != 1 {
		t.Fatalf("consume parameters = %d", consume.Parameters().Count())
	}
	if got := namedParameterNameText(t, consume.Parameters().At(0)); got != "decodeblocks" {
		t.Fatalf("consume parameter name = %q", got)
	}
	if consume.Parameters().At(0).Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("consume parameter expression = %v", consume.Parameters().At(0).Expression().Kind())
	}

	commaTail := ParseQuery("T | consume, Extra")
	if !diagnosticsContain(commaTail.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected consume comma-tail diagnostic, got %#v", commaTail.ContainedDiagnostics())
	}

	invokePipe, ok := consumePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("invoke pipe = %#v", consumePipe.Expression())
	}
	invoke, ok := invokePipe.Operator().(*InvokeOperator)
	if !ok {
		t.Fatalf("invoke operator = %#v", invokePipe.Operator())
	}
	if invoke.Function().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("invoke function kind = %v", invoke.Function().Kind())
	}
	if invokePipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("invoke source kind = %v", invokePipe.Expression().Kind())
	}
}

func TestParseInvokeReportsUnexpectedCallTail(t *testing.T) {
	code := Parse(`T | invoke Keep() Extra`)
	if !diagnosticsContainMessage(code.GetDiagnostics(), "KS005", "Expected: ;") || diagnosticsContain(code.GetDiagnostics(), "KS198") {
		t.Fatalf("expected separate-statement diagnostic, got %#v", code.GetDiagnostics())
	}
	if diagnosticsContain(code.GetDiagnostics(), "KS001") {
		t.Fatalf("invoke call tail should not cascade to missing-token diagnostics, got %#v", code.GetDiagnostics())
	}

	commaTail := Parse(`T | invoke Keep(), Extra`)
	if !diagnosticsContain(commaTail.GetDiagnostics(), "KS198") {
		t.Fatalf("expected comma-tail skipped-token diagnostic, got %#v", commaTail.GetDiagnostics())
	}
	if diagnosticsContain(commaTail.GetDiagnostics(), "KS001") {
		t.Fatalf("invoke comma tail should not cascade to missing-token diagnostics, got %#v", commaTail.GetDiagnostics())
	}
}

func TestParseQueryBuildsSampleOperators(t *testing.T) {
	const text = "T | sample 10 | sample-distinct 3 of UserPrincipalName"
	expr := parsedQueryExpression(t, text)

	distinctPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("sample-distinct pipe = %#v", expr)
	}
	sampleDistinct, ok := distinctPipe.Operator().(*SampleDistinctOperator)
	if !ok {
		t.Fatalf("sample-distinct operator = %#v", distinctPipe.Operator())
	}
	if sampleDistinct.Parameters().Count() != 0 {
		t.Fatalf("sample-distinct parameters = %d", sampleDistinct.Parameters().Count())
	}
	if sampleDistinct.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("sample-distinct expression kind = %v", sampleDistinct.Expression().Kind())
	}
	if sampleDistinct.OfKeyword().Kind() != SyntaxKindOfKeyword {
		t.Fatalf("sample-distinct of keyword = %v", sampleDistinct.OfKeyword().Kind())
	}
	if sampleDistinct.OfExpression().Kind() != SyntaxKindNameReference {
		t.Fatalf("sample-distinct of expression kind = %v", sampleDistinct.OfExpression().Kind())
	}

	samplePipe, ok := distinctPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("sample pipe = %#v", distinctPipe.Expression())
	}
	sample, ok := samplePipe.Operator().(*SampleOperator)
	if !ok {
		t.Fatalf("sample operator = %#v", samplePipe.Operator())
	}
	if sample.Parameters().Count() != 0 {
		t.Fatalf("sample parameters = %d", sample.Parameters().Count())
	}
	if sample.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("sample expression kind = %v", sample.Expression().Kind())
	}
	if samplePipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("sample source kind = %v", samplePipe.Expression().Kind())
	}
}

func TestParseSampleOperatorsReportUnexpectedTailSkippedTokens(t *testing.T) {
	tests := []struct {
		text string
		code string
	}{
		{"T | sample 5, Other", "KS198"},
		{"T | sample-distinct 5 of User, Other", "KS198"},
		{"T | sample-distinct 5 of User Other", "KS005"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			query := ParseQuery(tt.text)
			if !diagnosticsContain(query.ContainedDiagnostics(), tt.code) {
				t.Fatalf("expected %s diagnostic, got %#v", tt.code, query.ContainedDiagnostics())
			}
		})
	}
}

func TestParseConsumeKeepsMissingEqualsParameter(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		parameterName  string
		parameterValue string
		ofOperandText  string
		operatorAssert func(*testing.T, QueryOperator) *SyntaxListOf[*NamedParameter]
	}{
		{
			name:           "consume",
			text:           "T | consume decodeblocks maybe",
			parameterName:  "decodeblocks",
			parameterValue: "maybe",
			operatorAssert: func(t *testing.T, op QueryOperator) *SyntaxListOf[*NamedParameter] {
				t.Helper()
				consume, ok := op.(*ConsumeOperator)
				if !ok {
					t.Fatalf("consume operator = %#v", op)
				}
				return consume.Parameters()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
				t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
			}
			expr := parsedQueryExpression(t, tt.text)
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("pipe expression = %#v", expr)
			}
			params := tt.operatorAssert(t, pipe.Operator())
			if params.Count() != 1 {
				t.Fatalf("parameters = %d", params.Count())
			}
			param := params.At(0)
			if got := namedParameterNameText(t, param); got != tt.parameterName {
				t.Fatalf("parameter = %q", got)
			}
			if param.EqualToken() == nil || !diagnosticsContain(param.EqualToken().Diagnostics(), "KS005") {
				t.Fatalf("missing-equals diagnostics = %#v", param.EqualToken())
			}
			if param.Expression() == nil || strings.TrimSpace(param.Expression().Text(IncludeTriviaAll)) != tt.parameterValue {
				t.Fatalf("parameter expression = %#v", param.Expression())
			}
		})
	}
}

func TestParseSampleDistinctDoesNotConsumeUnsupportedParameter(t *testing.T) {
	const text = "T | sample-distinct hint.strategy shuffle 5 of User"
	expr := parsedQueryExpression(t, text)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("sample-distinct pipe = %#v", expr)
	}
	op, ok := pipe.Operator().(*SampleDistinctOperator)
	if !ok {
		t.Fatalf("sample-distinct operator = %#v", pipe.Operator())
	}
	if op.Parameters().Count() != 0 {
		t.Fatalf("sample-distinct parameters = %d, want none", op.Parameters().Count())
	}
	if op.Expression().Width() != 0 || op.OfExpression().Width() != 0 {
		t.Fatalf("sample-distinct recovered expressions = %q / %q", op.Expression().Text(IncludeTriviaAll), op.OfExpression().Text(IncludeTriviaAll))
	}
	if diagnostics := ParseQuery(text).ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS006") || !diagnosticsContain(diagnostics, "KS005") || !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("sample-distinct diagnostics = %#v", diagnostics)
	}
}

func TestParseSampleDistinctMissingOfKeywordRecoversExpression(t *testing.T) {
	expr := parsedQueryExpression(t, "T | sample-distinct 1")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("sample-distinct pipe = %#v", expr)
	}
	sampleDistinct, ok := pipe.Operator().(*SampleDistinctOperator)
	if !ok {
		t.Fatalf("sample-distinct operator = %#v", pipe.Operator())
	}
	if sampleDistinct.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("sample-distinct count expression kind = %v", sampleDistinct.Expression().Kind())
	}
	if sampleDistinct.OfKeyword().RawText() != "" {
		t.Fatalf("sample-distinct missing of keyword = %#v", sampleDistinct.OfKeyword())
	}
	if sampleDistinct.OfExpression() == nil {
		t.Fatal("sample-distinct missing of expression is nil")
	}
	diagnostics := sampleDistinct.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: of") || !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing of/expression diagnostics, got %#v", diagnostics)
	}

	const commaBeforeOfText = "T | sample-distinct 1, Other of User"
	commaExpr := parsedQueryExpression(t, commaBeforeOfText)
	commaPipe, ok := commaExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("sample-distinct comma pipe = %#v", commaExpr)
	}
	commaSampleDistinct, ok := commaPipe.Operator().(*SampleDistinctOperator)
	if !ok {
		t.Fatalf("sample-distinct comma operator = %#v", commaPipe.Operator())
	}
	if commaSampleDistinct.OfKeyword().RawText() != "" {
		t.Fatalf("sample-distinct comma-before-of should keep missing of keyword, got %#v", commaSampleDistinct.OfKeyword())
	}
	if text := strings.TrimSpace(commaSampleDistinct.OfExpression().Text(IncludeTriviaAll)); text != "" {
		t.Fatalf("sample-distinct comma-before-of should keep missing of expression, got %q", text)
	}
	commaDiagnostics := commaSampleDistinct.ContainedDiagnostics()
	for _, code := range []string{"KS005", "KS006"} {
		if !diagnosticsContain(commaDiagnostics, code) {
			t.Fatalf("expected %s in sample-distinct comma-before-of diagnostics, got %#v", code, commaDiagnostics)
		}
	}
	queryDiagnostics := ParseQuery(commaBeforeOfText).ContainedDiagnostics()
	if !diagnosticsContain(queryDiagnostics, "KS198") {
		t.Fatalf("expected KS198 in sample-distinct comma-before-of query diagnostics, got %#v", queryDiagnostics)
	}
}

func TestParseEvaluateUnexpectedCommaTailProducesDiagnostic(t *testing.T) {
	const text = "T | evaluate bag_unpack(Properties), Extra"
	query := ParseQuery(text)
	expr := parsedQueryExpression(t, text)

	evaluatePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("evaluate pipe = %#v", expr)
	}
	evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
	if !ok {
		t.Fatalf("evaluate operator = %#v", evaluatePipe.Operator())
	}
	if !diagnosticsContainCode(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected evaluate comma-tail diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if diagnosticsContainCode(evaluate.ContainedDiagnostics(), "KS198") {
		t.Fatalf("evaluate operator should not own QueryBlock skipped-tail diagnostics, got %#v", evaluate.ContainedDiagnostics())
	}
	if evaluate.FunctionCall() == nil || evaluate.FunctionCall().ArgumentList().Expressions().Count() != 1 {
		t.Fatalf("evaluate function call should preserve original argument list, got %#v", evaluate.FunctionCall())
	}
}

func TestParseEvaluateMissingArgumentListPreservesFunctionName(t *testing.T) {
	const text = "T | evaluate bag_unpack : (Account:string)"
	expr := parsedQueryExpression(t, text)

	evaluatePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("evaluate pipe = %#v", expr)
	}
	evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
	if !ok {
		t.Fatalf("evaluate operator = %#v", evaluatePipe.Operator())
	}
	callName, ok := evaluate.FunctionCall().Name().Name().(*TokenName)
	if !ok || callName.Name().RawText() != "bag_unpack" {
		t.Fatalf("evaluate function name = %#v", evaluate.FunctionCall().Name().Name())
	}
	if evaluate.FunctionCall().ArgumentList().Expressions().Count() != 0 {
		t.Fatalf("missing argument list expressions = %d", evaluate.FunctionCall().ArgumentList().Expressions().Count())
	}
	diagnostics := evaluate.FunctionCall().ArgumentList().ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: (") || !diagnosticsContainMessage(diagnostics, "KS005", "Expected: )") {
		t.Fatalf("expected missing argument-list expected-token diagnostics, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("missing argument-list should not use missing-token diagnostics, got %#v", diagnostics)
	}
	if evaluate.Schema() == nil || evaluate.Schema().ColonToken().Kind() != SyntaxKindColonToken {
		t.Fatalf("evaluate schema was not preserved: %#v", evaluate.Schema())
	}
}

func TestParseEvaluateMissingFunctionCallReportsSingleDiagnostic(t *testing.T) {
	query := ParseQuery("T | evaluate")
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing function call") {
		t.Fatalf("expected missing function-call diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("bare evaluate should not add missing-token diagnostics, got %#v", diagnostics)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("expected exactly one bare evaluate diagnostic, got %#v", diagnostics)
	}
}

func TestParseEvaluateSchemaRecoveryDiagnostics(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantKS198 bool
	}{
		{
			name: "missing schema after colon",
			text: "T | evaluate bag_unpack(Properties) :",
		},
		{
			name:      "unparenthesized star schema",
			text:      "T | evaluate bag_unpack(Properties) : *",
			wantKS198: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.text)
			diagnostics := code.GetDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS006", "Missing schema declaration") {
				t.Fatalf("expected missing evaluate schema diagnostic, got %#v", diagnostics)
			}
			if gotKS198 := diagnosticsContain(diagnostics, "KS198"); gotKS198 != tt.wantKS198 {
				t.Fatalf("KS198 presence = %t, want %t; diagnostics %#v", gotKS198, tt.wantKS198, diagnostics)
			}
			if tt.name == "unparenthesized star schema" && !diagnosticsContainLocation(diagnostics, "KS198", 38, 1) {
				t.Fatalf("unparenthesized star KS198 location = %#v, want start 38 length 1", diagnostics)
			}
			expr := parsedQueryExpression(t, tt.text)
			evaluatePipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("evaluate pipe = %#v", expr)
			}
			evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
			if !ok {
				t.Fatalf("evaluate operator = %#v", evaluatePipe.Operator())
			}
			if evaluate.Schema() == nil || evaluate.Schema().Schema() == nil {
				t.Fatalf("evaluate schema was not preserved: %#v", evaluate.Schema())
			}
			if !diagnosticsContainMessage(evaluate.Schema().Schema().Diagnostics(), "KS006", "Missing schema declaration") {
				t.Fatalf("missing schema diagnostics = %#v", evaluate.Schema().Schema().Diagnostics())
			}
		})
	}

	const doubleCommaText = "T | evaluate bag_unpack(Properties) : (Unpacked:string,, Other:long)"
	doubleComma := Parse(doubleCommaText)
	if diagnostics := doubleComma.GetDiagnostics(); !diagnosticsContainMessage(diagnostics, "KS005", "Expected: )") || !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected double-comma evaluate schema close/skipped diagnostics, got %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, doubleCommaText)
	evaluatePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("double-comma evaluate pipe = %#v", expr)
	}
	evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
	if !ok || evaluate.Schema() == nil || evaluate.Schema().Schema() == nil {
		t.Fatalf("double-comma evaluate operator/schema = %#v", evaluatePipe.Operator())
	}
	if evaluate.Schema().Schema().Columns().Count() != 1 {
		t.Fatalf("double-comma evaluate schema column count = %d", evaluate.Schema().Schema().Columns().Count())
	}

	for _, tt := range []struct {
		name        string
		text        string
		wantColumns int
	}{
		{
			name:        "leading comma",
			text:        "T | evaluate bag_unpack(Properties) : (, Name:string)",
			wantColumns: 1,
		},
		{
			name:        "trailing comma",
			text:        "T | evaluate bag_unpack(Properties) : (Name:string,)",
			wantColumns: 1,
		},
		{
			name:        "star trailing comma",
			text:        "T | evaluate bag_unpack(Properties) : (*,)",
			wantColumns: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("unexpected edge-comma evaluate schema diagnostics: %#v", diagnostics)
			}
			expr := parsedQueryExpression(t, tt.text)
			evaluatePipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("edge-comma evaluate pipe = %#v", expr)
			}
			evaluate, ok := evaluatePipe.Operator().(*EvaluateOperator)
			if !ok || evaluate.Schema() == nil || evaluate.Schema().Schema() == nil {
				t.Fatalf("edge-comma evaluate operator/schema = %#v", evaluatePipe.Operator())
			}
			if evaluate.Schema().Schema().Columns().Count() != tt.wantColumns {
				t.Fatalf("edge-comma evaluate schema column count = %d, want %d", evaluate.Schema().Schema().Columns().Count(), tt.wantColumns)
			}
		})
	}
}

func TestParseQueryBuildsJoinShuffleKeyUnknownColumnParameter(t *testing.T) {
	const text = "T | join hint.shufflekey=MissingColumn (SigninLogs) on UserPrincipalName"
	expr := parsedQueryExpression(t, text)

	joinPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("join pipe = %#v", expr)
	}
	join, ok := joinPipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("join operator = %#v", joinPipe.Operator())
	}
	if join.Parameters().Count() != 1 {
		t.Fatalf("join parameters = %d", join.Parameters().Count())
	}
	parameter := join.Parameters().At(0)
	if got := namedParameterNameText(t, parameter); got != "hint.shufflekey" {
		t.Fatalf("join parameter = %q", got)
	}
	if parameter.Expression() == nil {
		t.Fatal("join shufflekey parameter expression is nil")
	}
	if parameter.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("join shufflekey expression kind = %v", parameter.Expression().Kind())
	}
	if parameter.ExpressionHint() != CompletionHintColumn {
		t.Fatalf("join shufflekey hint = %v", parameter.ExpressionHint())
	}
}

func TestParseQueryJoinShuffleKeyMissingKeywordDoesNotPanic(t *testing.T) {
	code := Parse("T | join hint.shufflekey=Missing (SigninLogs) on UserPrincipalName")
	if code.Syntax() == nil {
		t.Fatal("missing syntax root")
	}
}

func TestParseQueryBuildsJoinWhereAndLookupMissingOnClause(t *testing.T) {
	const joinText = "T | join (SigninLogs) where AppId == ResourceAppId"
	joinExpr := parsedQueryExpression(t, joinText)
	joinPipe, ok := joinExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("join pipe = %#v", joinExpr)
	}
	join, ok := joinPipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("join operator = %#v", joinPipe.Operator())
	}
	where, ok := join.ConditionClause().(*JoinWhereClause)
	if !ok {
		t.Fatalf("join where clause = %#v", join.ConditionClause())
	}
	if where.Expression().Kind() != SyntaxKindEqualExpression {
		t.Fatalf("join where expression kind = %v", where.Expression().Kind())
	}

	const lookupText = "T | lookup (IdentityInfo)"
	lookupExpr := parsedQueryExpression(t, lookupText)
	lookupPipe, ok := lookupExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("lookup pipe = %#v", lookupExpr)
	}
	lookup, ok := lookupPipe.Operator().(*LookupOperator)
	if !ok {
		t.Fatalf("lookup operator = %#v", lookupPipe.Operator())
	}
	missing, ok := lookup.LookupClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("lookup missing on clause = %#v", lookup.LookupClause())
	}
	if missing.OnKeyword().RawText() != "" {
		t.Fatalf("lookup missing on keyword = %#v", missing.OnKeyword())
	}
	diagnostics := missing.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing join on condition clause") {
		t.Fatalf("lookup missing on clause diagnostics = %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("lookup missing on clause should not add missing-token diagnostics, got %#v", diagnostics)
	}
}

func TestParseJoinMissingOnKeywordStartsNewStatement(t *testing.T) {
	const text = "T | join (SigninLogs) UserPrincipalName"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") {
		t.Fatalf("expected statement separator diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContainMessage(diagnostics, "KS006", "Missing join on condition clause") || diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("join missing on keyword should not synthesize a missing clause/token diagnostic, got %#v", diagnostics)
	}

	expr := firstParsedQueryExpression(t, text)
	joinPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("join pipe = %#v", expr)
	}
	join, ok := joinPipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("join operator = %#v", joinPipe.Operator())
	}
	if join.ConditionClause() != nil {
		t.Fatalf("join condition clause = %#v, want nil", join.ConditionClause())
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("query statements = %#v, want two", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok || first.Separator() == nil || first.Separator().RawText() != "" {
		t.Fatalf("join statement separator = %#v", statements.At(0))
	}
	if !diagnosticsContainMessage(first.Separator().ContainedDiagnostics(), "KS005", "Expected: ;") {
		t.Fatalf("join statement separator diagnostics = %#v", first.Separator().ContainedDiagnostics())
	}
	if second := parsedQueryExpressionAt(t, text, 1, false); second == nil || second.Kind() != SyntaxKindNameReference {
		t.Fatalf("second statement expression = %#v", second)
	}
}

func TestParseJoinRightWhitespaceTailStartsStatements(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "join on", text: "T | join U Extra on Account"},
		{name: "join where", text: "T | join U Extra where Account == Account"},
		{name: "lookup on", text: "T | lookup U Extra on Account"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			diagnostics := query.ContainedDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") || diagnosticsContain(diagnostics, "KS198") {
				t.Fatalf("expected statement-boundary diagnostics, got %#v", diagnostics)
			}
			statements, ok := query.GetChild(1).(*SyntaxList)
			if !ok || statements.Count() < 2 {
				t.Fatalf("query statements = %#v, want recovered tail statements", statements)
			}
		})
	}
}

func TestParseLookupMissingOnKeywordDoesNotRecoverConditionExpression(t *testing.T) {
	const text = "T | lookup (IdentityInfo) AccountObjectId"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing join on condition clause") {
		t.Fatalf("expected missing lookup-on clause diagnostic, got %#v", diagnostics)
	}
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") {
		t.Fatalf("lookup missing on keyword tail should be skipped, got %#v", diagnostics)
	}

	expr := firstParsedQueryExpression(t, text)
	lookupPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("lookup pipe = %#v", expr)
	}
	lookup, ok := lookupPipe.Operator().(*LookupOperator)
	if !ok {
		t.Fatalf("lookup operator = %#v", lookupPipe.Operator())
	}
	on, ok := lookup.LookupClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("lookup condition clause = %#v", lookup.LookupClause())
	}
	if on.OnKeyword().RawText() != "" {
		t.Fatalf("recovered lookup on keyword = %#v", on.OnKeyword())
	}
	if on.Expressions().Count() != 0 {
		t.Fatalf("recovered lookup on expressions = %d", on.Expressions().Count())
	}
}

func TestParseLookupWhereTailDoesNotRecoverConditionExpression(t *testing.T) {
	const text = "T | lookup (IdentityInfo) where AccountObjectId == AccountObjectId"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing join on condition clause") {
		t.Fatalf("expected missing lookup-on clause diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("lookup where tail should not be parsed as recovered expression, got %#v", diagnostics)
	}
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") || diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("lookup where tail should start a new statement with expected-semicolon recovery, got %#v", diagnostics)
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("lookup where query statements = %#v, want two", query.GetChild(1))
	}

	expr := firstParsedQueryExpression(t, text)
	lookupPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("lookup pipe = %#v", expr)
	}
	lookup, ok := lookupPipe.Operator().(*LookupOperator)
	if !ok {
		t.Fatalf("lookup operator = %#v", lookupPipe.Operator())
	}
	on, ok := lookup.LookupClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("lookup condition clause = %#v", lookup.LookupClause())
	}
	if on.Expressions().Count() != 0 {
		t.Fatalf("lookup where tail should not recover on expressions, got %d", on.Expressions().Count())
	}
}

func TestParseJoinOnMissingExpressionDoesNotReportMissingClause(t *testing.T) {
	expr := parsedQueryExpression(t, "T | join (SigninLogs) on")
	joinPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("join pipe = %#v", expr)
	}
	join, ok := joinPipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("join operator = %#v", joinPipe.Operator())
	}
	on, ok := join.ConditionClause().(*JoinOnClause)
	if !ok {
		t.Fatalf("join on clause = %#v", join.ConditionClause())
	}
	diagnostics := on.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing expression diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContainMessage(diagnostics, "KS006", "Missing join on condition clause") {
		t.Fatalf("join on missing expression should not report missing clause, got %#v", diagnostics)
	}
}

func TestParseQueryBuildsMvExpandOperator(t *testing.T) {
	const text = "T | mv-expand kind=array bagexpansion=bag Items to typeof(string), Count = Values to typeof(long) limit 10"
	expr := parsedQueryExpression(t, text)

	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-expand pipe = %#v", expr)
	}
	mvExpand, ok := mvPipe.Operator().(*MvExpandOperator)
	if !ok {
		t.Fatalf("mv-expand operator = %#v", mvPipe.Operator())
	}
	if mvExpand.MvExpandKeyword().Kind() != SyntaxKindMvDashExpandKeyword {
		t.Fatalf("mv-expand keyword = %v", mvExpand.MvExpandKeyword().Kind())
	}
	if mvExpand.Parameters().Count() != 2 {
		t.Fatalf("mv-expand parameters = %d", mvExpand.Parameters().Count())
	}
	if got := namedParameterNameText(t, mvExpand.Parameters().At(0)); got != "kind" {
		t.Fatalf("mv-expand parameter 0 = %q", got)
	}
	if got := namedParameterNameText(t, mvExpand.Parameters().At(1)); got != "bagexpansion" {
		t.Fatalf("mv-expand parameter 1 = %q", got)
	}
	if mvExpand.Expressions().Count() != 2 {
		t.Fatalf("mv-expand expressions = %d", mvExpand.Expressions().Count())
	}
	first := mvExpand.Expressions().At(0).Element()
	if first.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("mv-expand first expression kind = %v", first.Expression().Kind())
	}
	if first.ToTypeOf() == nil || first.ToTypeOf().TypeOf().Types().Count() != 1 {
		t.Fatalf("mv-expand first typeof = %#v", first.ToTypeOf())
	}
	if first.ToTypeOf().TypeOf().Types().At(0).Element().Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("mv-expand first type kind = %v", first.ToTypeOf().TypeOf().Types().At(0).Element().Kind())
	}
	second := mvExpand.Expressions().At(1).Element()
	if second.Expression().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("mv-expand second expression kind = %v", second.Expression().Kind())
	}
	if second.ToTypeOf() == nil || second.ToTypeOf().TypeOf().Types().At(0).Element().Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("mv-expand second typeof = %#v", second.ToTypeOf())
	}
	if mvExpand.RowLimitClause() == nil {
		t.Fatal("missing mv-expand row limit")
	}
	if mvExpand.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("mv-expand row limit kind = %v", mvExpand.RowLimitClause().RowLimit().Kind())
	}
	if mvPipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("mv-expand source kind = %v", mvPipe.Expression().Kind())
	}
}

func TestParseQueryBuildsMvExpandLeadingTypeOfExpression(t *testing.T) {
	const text = "T | mvexpand to typeof(string)"
	expr := parsedQueryExpression(t, text)

	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mvexpand pipe = %#v", expr)
	}
	mvExpand, ok := mvPipe.Operator().(*MvExpandOperator)
	if !ok {
		t.Fatalf("mvexpand operator = %#v", mvPipe.Operator())
	}
	if mvExpand.MvExpandKeyword().Kind() != SyntaxKindMvExpandKeyword {
		t.Fatalf("mvexpand keyword = %v", mvExpand.MvExpandKeyword().Kind())
	}
	if mvExpand.Expressions().Count() != 1 {
		t.Fatalf("mvexpand expressions = %d", mvExpand.Expressions().Count())
	}
	item := mvExpand.Expressions().At(0).Element()
	if item.ToTypeOf() == nil {
		t.Fatal("missing mvexpand typeof clause")
	}
	if !diagnosticsContain(item.Expression().Diagnostics(), "KS006") {
		t.Fatalf("leading typeof mvexpand expression should carry KS006, got %#v", item.Expression().Diagnostics())
	}
}

func TestParseMvExpandRecoveryDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		check func(*testing.T, *MvExpandOperator)
	}{
		{
			name: "missing typeof after to",
			text: "T | mv-expand Items to",
			check: func(t *testing.T, op *MvExpandOperator) {
				if op.Expressions().Count() != 1 {
					t.Fatalf("expressions = %d", op.Expressions().Count())
				}
				item := op.Expressions().At(0).Element()
				if item.Expression().Kind() != SyntaxKindNameReference {
					t.Fatalf("expression kind = %v", item.Expression().Kind())
				}
				if item.ToTypeOf() == nil || !diagnosticsContain(item.ToTypeOf().ContainedDiagnostics(), "KS006") {
					t.Fatalf("expected missing typeof diagnostic, got %#v", item.ToTypeOf())
				}
			},
		},
		{
			name: "trailing comma before limit",
			text: "T | mv-expand Items, limit 5",
			check: func(t *testing.T, op *MvExpandOperator) {
				if op.Expressions().Count() != 2 {
					t.Fatalf("expressions = %d", op.Expressions().Count())
				}
				item := op.Expressions().At(1).Element()
				if item.Expression().Kind() != SyntaxKindNameReference {
					t.Fatalf("missing expression kind = %v", item.Expression().Kind())
				}
				if !diagnosticsContain(item.Expression().Diagnostics(), "KS006") {
					t.Fatalf("expected missing expression diagnostic, got %#v", item.Expression().Diagnostics())
				}
				if op.RowLimitClause() == nil || op.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
					t.Fatalf("expected row-limit clause to survive recovery, got %#v", op.RowLimitClause())
				}
			},
		},
		{
			name: "missing row limit expression",
			text: "T | mv-expand Items limit",
			check: func(t *testing.T, op *MvExpandOperator) {
				if op.RowLimitClause() == nil || !diagnosticsContain(op.RowLimitClause().ContainedDiagnostics(), "KS006") {
					t.Fatalf("expected missing row-limit diagnostic, got %#v", op.RowLimitClause())
				}
			},
		},
		{
			name: "missing expand expression before limit",
			text: "T | mv-expand limit 5",
			check: func(t *testing.T, op *MvExpandOperator) {
				if op.Expressions().Count() != 1 {
					t.Fatalf("expressions = %d", op.Expressions().Count())
				}
				item := op.Expressions().At(0).Element()
				if !diagnosticsContain(item.Expression().Diagnostics(), "KS006") {
					t.Fatalf("expected missing expression diagnostic, got %#v", item.Expression().Diagnostics())
				}
				if op.RowLimitClause() == nil || op.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
					t.Fatalf("expected preserved row limit, got %#v", op.RowLimitClause())
				}
			},
		},
		{
			name: "missing parameter equals before expression",
			text: "T | mv-expand kind array Items",
			check: func(t *testing.T, op *MvExpandOperator) {
				if op.Parameters().Count() != 0 {
					t.Fatalf("parameters = %d", op.Parameters().Count())
				}
				if op.Expressions().Count() != 3 {
					t.Fatalf("expressions = %d", op.Expressions().Count())
				}
				for i, want := range []string{"kind", "array", "Items"} {
					if got := strings.TrimSpace(op.Expressions().At(i).Element().Text(IncludeTriviaAll)); got != want {
						t.Fatalf("expression %d = %q, want %q", i, got, want)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parsedQueryExpression(t, tt.text)
			mvPipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("mv-expand pipe = %#v", expr)
			}
			mvExpand, ok := mvPipe.Operator().(*MvExpandOperator)
			if !ok {
				t.Fatalf("mv-expand operator = %#v", mvPipe.Operator())
			}
			tt.check(t, mvExpand)
		})
	}
}

func TestParseMvExpandRowLimitTailStartsNewStatement(t *testing.T) {
	query := ParseQuery("T | mv-expand Items limit 5 Extra")
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %#v, want parsed mv-expand plus recovered Extra statement", statements)
	}
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ;") || diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("QueryBlock diagnostics = %#v, want missing semicolon without skipped-tail KS198", diagnostics)
	}
}

func TestParseMvExpandRejectsBareTypeAfterTo(t *testing.T) {
	const text = "T | mv-expand Payload to int"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing typeof") {
		t.Fatalf("expected missing typeof diagnostic, got %#v", diagnostics)
	}
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected skipped bare type diagnostic, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-expand pipe = %#v", expr)
	}
	mvExpand, ok := mvPipe.Operator().(*MvExpandOperator)
	if !ok {
		t.Fatalf("mv-expand operator = %#v", mvPipe.Operator())
	}
	if mvExpand.Expressions().Count() != 1 {
		t.Fatalf("expressions = %d, want 1", mvExpand.Expressions().Count())
	}
	item := mvExpand.Expressions().At(0).Element()
	if item.ToTypeOf() == nil || !diagnosticsContainMessage(item.ToTypeOf().ContainedDiagnostics(), "KS006", "Missing typeof") {
		t.Fatalf("expected missing typeof clause diagnostic, got %#v", item.ToTypeOf())
	}
}

func TestParseQueryBuildsMvApplyOperator(t *testing.T) {
	const text = "T | mv-apply with_itemindex=Idx Item=Items to typeof(string) limit 5 id TraceId on (summarize Count=count() by Item)"
	expr := parsedQueryExpression(t, text)

	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-apply pipe = %#v", expr)
	}
	mvApply, ok := mvPipe.Operator().(*MvApplyOperator)
	if !ok {
		t.Fatalf("mv-apply operator = %#v", mvPipe.Operator())
	}
	if mvApply.MvApplyKeyword().Kind() != SyntaxKindMvDashApplyKeyword {
		t.Fatalf("mv-apply keyword = %v", mvApply.MvApplyKeyword().Kind())
	}
	if mvApply.Parameters().Count() != 1 {
		t.Fatalf("mv-apply parameters = %d", mvApply.Parameters().Count())
	}
	if got := namedParameterNameText(t, mvApply.Parameters().At(0)); got != "with_itemindex" {
		t.Fatalf("mv-apply parameter = %q", got)
	}
	if mvApply.Expressions().Count() != 1 {
		t.Fatalf("mv-apply expressions = %d", mvApply.Expressions().Count())
	}
	item := mvApply.Expressions().At(0).Element()
	if item.Expression().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("mv-apply expression kind = %v", item.Expression().Kind())
	}
	if item.ToTypeOf() == nil || item.ToTypeOf().TypeOf().Types().At(0).Element().Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("mv-apply typeof = %#v", item.ToTypeOf())
	}
	if mvApply.RowLimitClause() == nil || mvApply.RowLimitClause().RowLimit().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("mv-apply row limit = %#v", mvApply.RowLimitClause())
	}
	if mvApply.ContextIdClause() == nil || mvApply.ContextIdClause().Id().Kind() != SyntaxKindNameReference {
		t.Fatalf("mv-apply id clause = %#v", mvApply.ContextIdClause())
	}
	if mvApply.OnKeyword().Kind() != SyntaxKindOnKeyword {
		t.Fatalf("mv-apply on keyword = %v", mvApply.OnKeyword().Kind())
	}
	if mvApply.Subquery() == nil || mvApply.Subquery().Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("mv-apply subquery = %#v", mvApply.Subquery())
	}
}

func TestParseMvApplyMissingParameterEqualsStopsBeforeParameter(t *testing.T) {
	const text = "T | mv-apply with_itemindex Idx Item=Items on (project Item, Idx)"
	query := ParseQuery(text)

	var mvApply *MvApplyOperator
	WalkSyntax(query, func(element SyntaxElement) {
		if operator, ok := WrapGeneratedSyntaxElement(element).(*MvApplyOperator); ok && mvApply == nil {
			mvApply = operator
		}
	})
	if mvApply == nil {
		t.Fatalf("mv-apply operator missing from recovered syntax")
	}
	if mvApply.Parameters().Count() != 0 {
		t.Fatalf("parameters = %d", mvApply.Parameters().Count())
	}
	if mvApply.Expressions().Count() != 1 || !diagnosticsContain(mvApply.Expressions().At(0).Element().Expression().ContainedDiagnostics(), "KS006") {
		t.Fatalf("expressions = %d", mvApply.Expressions().Count())
	}
	if query.GetChild(2) == nil {
		t.Fatalf("missing-equals tail should remain QueryBlock skipped tokens")
	}
}

func TestParseMvApplyTrailingCommaBeforeOnCreatesMissingExpression(t *testing.T) {
	const text = "T | mv-apply Items, on (summarize Count=count())"
	expr := parsedQueryExpression(t, text)

	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-apply pipe = %#v", expr)
	}
	mvApply, ok := mvPipe.Operator().(*MvApplyOperator)
	if !ok {
		t.Fatalf("mv-apply operator = %#v", mvPipe.Operator())
	}
	if mvApply.Expressions().Count() != 2 {
		t.Fatalf("mv-apply expressions = %d", mvApply.Expressions().Count())
	}
	missing := mvApply.Expressions().At(1).Element()
	if !diagnosticsContainMessage(missing.Expression().Diagnostics(), "KS006", "Missing expression: If the keyword 'on' is intended to be part of an expression it needs to be bracketted as ['on'].") {
		t.Fatalf("trailing comma diagnostics = %#v", missing.Expression().Diagnostics())
	}
	if mvApply.OnKeyword().Kind() != SyntaxKindOnKeyword || mvApply.Subquery() == nil || mvApply.Subquery().Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("mv-apply on/subquery recovery = %v %#v", mvApply.OnKeyword().Kind(), mvApply.Subquery())
	}
}

func TestParseMvApplyTrailingCommaAfterSubqueryCreatesIncompleteFragment(t *testing.T) {
	query := ParseQuery(`T | mv-apply Item=Items to typeof(string) on (summarize ItemCount=count() by Item), Extra`)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected incomplete-fragment diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, `T | mv-apply Item=Items to typeof(string) on (summarize ItemCount=count() by Item), Extra`)
	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-apply pipe = %#v", expr)
	}
	mvApply, ok := mvPipe.Operator().(*MvApplyOperator)
	if !ok {
		t.Fatalf("mv-apply operator = %#v", mvPipe.Operator())
	}
	if mvApply.Subquery() == nil || mvApply.Subquery().Expression().Kind() != SyntaxKindSummarizeOperator {
		t.Fatalf("mv-apply subquery recovery = %#v", mvApply.Subquery())
	}
	if diagnosticsContain(mvApply.Diagnostics(), "KS198") || query.GetChild(2) == nil {
		t.Fatalf("KS198 should belong to query skipped tokens: operator=%#v skipped=%#v", mvApply.Diagnostics(), query.GetChild(2))
	}
}

func TestParseMvApplySubqueryRequiresContextualOperator(t *testing.T) {
	query := ParseQuery(`T | mv-apply Items on (Items)`)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS176") {
		t.Fatalf("expected query-operator diagnostic, got %#v", query.ContainedDiagnostics())
	}
	expr := parsedQueryExpression(t, `T | mv-apply Items on (Items)`)
	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-apply pipe = %#v", expr)
	}
	mvApply, ok := mvPipe.Operator().(*MvApplyOperator)
	if !ok {
		t.Fatalf("mv-apply operator = %#v", mvPipe.Operator())
	}
	if mvApply.Subquery() == nil || mvApply.Subquery().Expression().Kind() != SyntaxKindBadQueryOperator {
		t.Fatalf("mv-apply bad subquery = %#v", mvApply.Subquery())
	}
}

func TestParseMvApplySubqueryAcceptsContextualOperatorAndDatatable(t *testing.T) {
	operator := ParseQuery(`T | mv-apply Items on (where isnotempty(Items))`)
	if diagnosticsContain(operator.ContainedDiagnostics(), "KS176") {
		t.Fatalf("operator subquery diagnostics = %#v", operator.ContainedDiagnostics())
	}

	sourcePipe := ParseQuery(`T | mv-apply Items on (Other | count)`)
	if !diagnosticsContain(sourcePipe.ContainedDiagnostics(), "KS176") {
		t.Fatalf("source pipe subquery should require an operator root, got %#v", sourcePipe.ContainedDiagnostics())
	}
	expr := parsedQueryExpression(t, `T | mv-apply Items on (Other | count)`)
	mvPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("mv-apply source pipe = %#v", expr)
	}
	mvApply, ok := mvPipe.Operator().(*MvApplyOperator)
	if !ok {
		t.Fatalf("mv-apply source pipe operator = %#v", mvPipe.Operator())
	}
	if mvApply.Subquery() == nil || mvApply.Subquery().Expression().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("mv-apply source pipe subquery = %#v", mvApply.Subquery())
	}

	datatable := ParseQuery(`T | mv-apply Items on (__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (Items:string) | count)`)
	if diagnosticsContain(datatable.ContainedDiagnostics(), "KS176") {
		t.Fatalf("contextual datatable subquery diagnostics = %#v", datatable.ContainedDiagnostics())
	}
}

func TestParseQueryBuildsScanOperator(t *testing.T) {
	const text = `T | scan kind=partial with_match_id=MatchId order by Timestamp asc partition by User declare (State:string, Count:long=0) with (step s output=all: EventType == "Start" => State = EventType, Count = Count + 1;)`
	expr := parsedQueryExpression(t, text)

	scanPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("scan pipe = %#v", expr)
	}
	scan, ok := scanPipe.Operator().(*ScanOperator)
	if !ok {
		t.Fatalf("scan operator = %#v", scanPipe.Operator())
	}
	if scan.Parameters().Count() != 2 {
		t.Fatalf("scan parameters = %d", scan.Parameters().Count())
	}
	if got := namedParameterNameText(t, scan.Parameters().At(0)); got != "kind" {
		t.Fatalf("scan first parameter = %q", got)
	}
	if got := namedParameterNameText(t, scan.Parameters().At(1)); got != "with_match_id" {
		t.Fatalf("scan second parameter = %q", got)
	}
	if scan.OrderByClause() == nil || scan.OrderByClause().Expressions().Count() != 1 {
		t.Fatalf("scan order clause = %#v", scan.OrderByClause())
	}
	if scan.PartitionByClause() == nil || scan.PartitionByClause().Expressions().Count() != 1 {
		t.Fatalf("scan partition clause = %#v", scan.PartitionByClause())
	}
	if scan.DeclareClause() == nil || scan.DeclareClause().Declarations().Count() != 2 {
		t.Fatalf("scan declare clause = %#v", scan.DeclareClause())
	}
	secondDecl := scan.DeclareClause().Declarations().At(1).Element()
	if secondDecl.DefaultValue() == nil || secondDecl.DefaultValue().Value() == nil || secondDecl.DefaultValue().Value().Text(IncludeTriviaAll) != "0" {
		t.Fatalf("scan declaration default = %#v", secondDecl.DefaultValue())
	}
	if scan.Steps().Count() != 1 {
		t.Fatalf("scan steps = %d", scan.Steps().Count())
	}
	step := scan.Steps().At(0)
	if step.ScanStepOutput() == nil || step.ScanStepOutput().OutputKind().Text(IncludeTriviaAll) != "all" {
		t.Fatalf("scan step output = %#v", step.ScanStepOutput())
	}
	if step.ComputationClause() == nil || step.ComputationClause().Assignments().Count() != 2 {
		t.Fatalf("scan computation = %#v", step.ComputationClause())
	}
}

func TestParseQueryScanLateParameterRecoveryMatchesCurrentReference(t *testing.T) {
	const query = `T | scan order by Timestamp User declare (State:string) with_match_id=Seq with (step s: true;)`
	block := ParseQuery(query)
	got := GetContainedDiagnostics(block, DiagnosticsIncludeSyntactic)
	want := []struct {
		code    string
		message string
		start   int
		length  int
	}{
		{"KS005", "Expected: ,", 28, 4},
		{"KS005", "Expected: with", 56, 13},
		{"KS005", "Expected: (", 56, 13},
		{"KS005", "Expected: )", 56, 13},
		{"KS198", "The incomplete fragment is unexpected.", 56, 13},
	}
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Code != want[i].code || got[i].Message != want[i].message || got[i].LocationKind != DiagnosticLocationAbsolute || got[i].Start != want[i].start || got[i].Length != want[i].length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestParseQueryScanStepOutputKindIsCaseSensitive(t *testing.T) {
	valid := ParseQuery(`T | scan with (step s output=all: true;)`)
	if diagnostics := valid.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid lowercase output diagnostics = %#v", diagnostics)
	}

	invalid := ParseQuery(`T | scan with (step s output=ALL: true;)`)
	diagnostics := invalid.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected one of: all, last, none") {
		t.Fatalf("expected scan output kind diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") || diagnosticsContain(diagnostics, "KS142") {
		t.Fatalf("invalid scan output kind should not cascade to missing-colon/name diagnostics, got %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, `T | scan with (step s output=ALL: true;)`)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("scan pipe = %#v", expr)
	}
	scan, ok := pipe.Operator().(*ScanOperator)
	if !ok {
		t.Fatalf("scan operator = %#v", pipe.Operator())
	}
	step := scan.Steps().At(0)
	if step.ScanStepOutput() == nil || step.ScanStepOutput().OutputKind().RawText() != "ALL" {
		t.Fatalf("invalid scan output kind should be consumed, got %#v", step.ScanStepOutput())
	}
}

func TestParseScanEmptyStepListIsAllowed(t *testing.T) {
	query := ParseQuery(`T | scan with ()`)
	if diagnostics := query.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("empty scan step-list diagnostics = %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, `T | scan with ()`)
	scanPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("scan pipe = %#v", expr)
	}
	scan, ok := scanPipe.Operator().(*ScanOperator)
	if !ok {
		t.Fatalf("scan operator = %#v", scanPipe.Operator())
	}
	if scan.Steps().Count() != 0 {
		t.Fatalf("scan steps = %d", scan.Steps().Count())
	}
}

func TestParseScanMissingAssignmentReportsSingleMissingName(t *testing.T) {
	query := ParseQuery(`T | scan with (step s: true => ;)`)
	diagnostics := query.ContainedDiagnostics()
	if len(diagnostics) != 1 || !diagnosticsContainMessage(diagnostics, "KS006", "Missing name") {
		t.Fatalf("expected single missing-name diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") || diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("missing scan assignment should not add token/expression recovery diagnostics, got %#v", diagnostics)
	}
}

func TestParseScanTrailingCommaReportsMissingAssignment(t *testing.T) {
	query := ParseQuery(`T | scan with (step s: true => x=1,;)`)
	diagnostics := query.ContainedDiagnostics()
	if len(diagnostics) != 1 || !diagnosticsContainMessage(diagnostics, "KS006", "Missing name") {
		t.Fatalf("expected single missing-name diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") || diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("missing scan assignment should not add token/expression recovery diagnostics, got %#v", diagnostics)
	}
}

func TestParseScanStepTrailingCommaReportsSkippedFragment(t *testing.T) {
	query := ParseQuery(`T | scan with (step s: true;,)`)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: )") {
		t.Fatalf("expected missing close-paren diagnostic, got %#v", diagnostics)
	}
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected skipped-token diagnostic, got %#v", diagnostics)
	}
}

func TestParseQueryBuildsParseKvEmptyAndCommaSchemas(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		columns int
	}{
		{
			name:    "empty",
			text:    "T | parse-kv Message as ()",
			columns: 0,
		},
		{
			name:    "leading comma",
			text:    "T | parse-kv Message as (, Account:string)",
			columns: 1,
		},
		{
			name:    "trailing comma",
			text:    "T | parse-kv Message as (Account:string,)",
			columns: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parsedQueryExpression(t, tt.text)
			parsePipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("parse-kv pipe = %#v", expr)
			}
			parseKv, ok := parsePipe.Operator().(*ParseKvOperator)
			if !ok {
				t.Fatalf("parse-kv operator = %#v", parsePipe.Operator())
			}
			if diagnostics := parseKv.Keys().ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("parse-kv %s schema diagnostics = %#v", tt.name, diagnostics)
			}
			if got := parseKv.Keys().Columns().Count(); got != tt.columns {
				t.Fatalf("parse-kv %s columns = %d, want %d", tt.name, got, tt.columns)
			}
		})
	}
}

func TestParseKvExpressionTailsUseStatementAndSkippedBoundaries(t *testing.T) {
	query := ParseQuery(`T | parse-kv Message as (Account:string) Extra | project Account`)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") || diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected missing statement separator without skipped-token diagnostics, got %#v", query.ContainedDiagnostics())
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("parse-kv whitespace tail statements = %T %#v", query.GetChild(1), query.GetChild(1))
	}
	if got := firstParsedQueryExpression(t, `T | parse-kv Message as (Account:string) Extra | project Account`); got.Kind() != SyntaxKindPipeExpression {
		t.Fatalf("parse-kv first statement expression = %#v", got)
	}

	commaTail := ParseQuery(`T | parse-kv Message as (Account:string), Extra | project Account`)
	if !diagnosticsContain(commaTail.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected comma-tail skipped-token diagnostic, got %#v", commaTail.ContainedDiagnostics())
	}
	skipped, ok := commaTail.GetChild(2).(*SkippedTokens)
	if !ok || skipped.Text(IncludeTriviaAll) != ", Extra | project Account" {
		t.Fatalf("parse-kv comma tail = %T %q, want QueryBlock skipped tokens", commaTail.GetChild(2), skipped.Text(IncludeTriviaAll))
	}
	expr := firstParsedQueryExpression(t, `T | parse-kv Message as (Account:string), Extra | project Account`)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse-kv comma-tail expression = %#v", expr)
	}
	if _, ok := pipe.Operator().(*ParseKvOperator); !ok {
		t.Fatalf("parse-kv comma-tail terminal operator = %#v", pipe.Operator())
	}
}

func TestParseQueryCurrentPreservesParseKvSourceTailRecovery(t *testing.T) {
	const query = `T | parse-kv Message as Account:string`
	block := ParseQuery(query)
	diagnostics := GetContainedDiagnostics(block, DiagnosticsIncludeSyntactic)
	wantDiagnostics := []struct {
		code   string
		start  int
		length int
	}{
		{"KS006", 23, 0},
		{"KS005", 24, 7},
		{"KS006", 32, 0},
		{"KS198", 32, 6},
	}
	if len(diagnostics) != len(wantDiagnostics) {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for i, want := range wantDiagnostics {
		got := diagnostics[i]
		if got.Code != want.code || got.LocationKind != DiagnosticLocationAbsolute || got.Start != want.start || got.Length != want.length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got, want)
		}
	}

	statements, ok := block.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("statements = %T %#v", block.GetChild(1), block.GetChild(1))
	}
	var parseKv *ParseKvOperator
	WalkSyntax(statements, func(element SyntaxElement) {
		if parseKv == nil {
			parseKv, _ = WrapGeneratedSyntaxElement(element).(*ParseKvOperator)
		}
	})
	if parseKv == nil || parseKv.Keys().Columns().Count() != 0 {
		t.Fatalf("parse-kv operator = %#v", parseKv)
	}
	skipped, ok := block.GetChild(2).(*SkippedTokens)
	if !ok || strings.TrimSpace(skipped.Text(IncludeTriviaAll)) != "string" {
		t.Fatalf("skipped tail = %T %q", block.GetChild(2), block.GetChild(2).Text(IncludeTriviaAll))
	}
}

func TestParseQueryBuildsParsePatternMissingTypeDiagnostic(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "parse",
			text: `T | parse Message with "user=" Account:`,
		},
		{
			name: "parse-where",
			text: `T | parse-where Message with "user=" Account:`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parsedQueryExpression(t, tt.text)
			parsePipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("%s pipe = %#v", tt.name, expr)
			}
			var patterns *SyntaxListOf[SyntaxNodeElement]
			switch op := parsePipe.Operator().(type) {
			case *ParseOperator:
				patterns = op.Patterns()
			case *ParseWhereOperator:
				patterns = op.Patterns()
			default:
				t.Fatalf("%s operator = %#v", tt.name, parsePipe.Operator())
			}
			if patterns.Count() != 2 {
				t.Fatalf("%s patterns = %d", tt.name, patterns.Count())
			}
			decl, ok := patterns.At(1).(*NameAndTypeDeclaration)
			if !ok {
				t.Fatalf("%s pattern 1 = %#v", tt.name, patterns.At(1))
			}
			diagnostics := decl.ContainedDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS006", "Missing type name") {
				t.Fatalf("%s missing type diagnostics = %#v", tt.name, diagnostics)
			}
			if diagnosticsContain(diagnostics, "KS001") {
				t.Fatalf("%s missing type should not also synthesize a missing-token diagnostic: %#v", tt.name, diagnostics)
			}
		})
	}
}

func TestParseQueryBuildsMakeSeriesFromToStepOperator(t *testing.T) {
	const text = "T | make-series kind=nonempty Count=count() default=0 on Timestamp from ago(1d) to now() step 1h by UserPrincipalName"
	expr := parsedQueryExpression(t, text)

	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	if makeSeries.MakeSeriesKeyword().Kind() != SyntaxKindMakeSeriesKeyword {
		t.Fatalf("make-series keyword = %v", makeSeries.MakeSeriesKeyword().Kind())
	}
	if makeSeries.Parameters().Count() != 1 {
		t.Fatalf("make-series parameters = %d", makeSeries.Parameters().Count())
	}
	if got := namedParameterNameText(t, makeSeries.Parameters().At(0)); got != "kind" {
		t.Fatalf("make-series parameter 0 = %q", got)
	}
	if makeSeries.Aggregates().Count() != 1 {
		t.Fatalf("make-series aggregates = %d", makeSeries.Aggregates().Count())
	}
	aggregate := makeSeries.Aggregates().At(0).Element()
	if aggregate.Expression().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("make-series aggregate expression = %v", aggregate.Expression().Kind())
	}
	if aggregate.DefaultExpression() == nil {
		t.Fatal("missing make-series default clause")
	}
	if aggregate.DefaultExpression().Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("make-series default expression = %v", aggregate.DefaultExpression().Expression().Kind())
	}
	if makeSeries.OnClause().Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("make-series on expression = %v", makeSeries.OnClause().Expression().Kind())
	}
	rangeClause, ok := makeSeries.RangeClause().(*MakeSeriesFromToStepClause)
	if !ok {
		t.Fatalf("make-series range clause = %#v", makeSeries.RangeClause())
	}
	if rangeClause.MakeSeriesFromClause() == nil || rangeClause.MakeSeriesToClause() == nil || rangeClause.MakeSeriesStepClause() == nil {
		t.Fatalf("make-series from/to/step clause = %#v", rangeClause)
	}
	if rangeClause.MakeSeriesStepClause().Expression().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("make-series step expression = %v", rangeClause.MakeSeriesStepClause().Expression().Kind())
	}
	if makeSeries.ByClause() == nil || makeSeries.ByClause().Expressions().Count() != 1 {
		t.Fatalf("make-series by clause = %#v", makeSeries.ByClause())
	}
}

func TestParseMakeSeriesEmptyDefaultConsumesOnAsSourceParserDoes(t *testing.T) {
	const text = "T | make-series Count=count() default= on Timestamp from datetime(2024-01-01) to datetime(2024-01-02) step 1h"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	for _, want := range []string{"Expected: ,", "Expected: on", "Expected: in", "Expected: range", "Expected: (", "Expected: )"} {
		if !diagnosticsContainMessage(diagnostics, "KS005", want) {
			t.Fatalf("missing %q diagnostic, got %#v", want, diagnostics)
		}
	}
	if !diagnosticsContain(diagnostics, "KS198") || diagnosticsContain(diagnostics, "KS006") {
		t.Fatalf("empty default source recovery diagnostics = %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	aggregate := makeSeries.Aggregates().At(0).Element()
	if aggregate.DefaultExpression() == nil {
		t.Fatal("missing recovered default clause")
	}
	if got := strings.TrimSpace(aggregate.DefaultExpression().Expression().Text(IncludeTriviaAll)); got != "on" {
		t.Fatalf("default expression = %q, want on", got)
	}
	if strings.TrimSpace(makeSeries.OnClause().Expression().Text(IncludeTriviaAll)) != "from" {
		t.Fatalf("make-series on expression = %q", makeSeries.OnClause().Expression().Text(IncludeTriviaAll))
	}
	if makeSeries.Aggregates().Count() != 2 || strings.TrimSpace(makeSeries.Aggregates().At(1).Element().Expression().Text(IncludeTriviaAll)) != "Timestamp" {
		t.Fatalf("make-series recovered aggregates = %#v", makeSeries.Aggregates())
	}
	if _, ok := makeSeries.RangeClause().(*MakeSeriesInRangeClause); !ok {
		t.Fatalf("make-series range recovery = %#v", makeSeries.RangeClause())
	}
}

func TestParseMakeSeriesMissingOnFromTailUsesInRangeRecovery(t *testing.T) {
	const text = "T | make-series Count=count() from ago(1d) to now() step 1h by Account"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	for _, want := range []string{"Expected: on", "Expected: in", "Expected: range", "Expected: (", "Expected: )"} {
		if !diagnosticsContainMessage(diagnostics, "KS005", want) {
			t.Fatalf("missing %q diagnostic, got %#v", want, diagnostics)
		}
	}
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("missing skipped-tail diagnostic, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	onName, ok := makeSeries.OnClause().Expression().(*NameReference)
	if !ok {
		t.Fatalf("make-series recovered on expression = %#v", makeSeries.OnClause().Expression())
	}
	if got := tokenNameRawText(onName.Name()); got != "from" {
		t.Fatalf("recovered on expression = %q, want from", got)
	}
	if _, ok := makeSeries.RangeClause().(*MakeSeriesInRangeClause); !ok {
		t.Fatalf("make-series range recovery = %#v", makeSeries.RangeClause())
	}
}

func TestParseMakeSeriesStepCommaTailReportsIncompleteFragment(t *testing.T) {
	const text = "T | make-series Hits=count() on TimeGenerated from ago(1d) to now() step 1h, Extra"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected make-series comma-tail diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("make-series comma-tail should not add missing-token diagnostics, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	rangeClause, ok := makeSeries.RangeClause().(*MakeSeriesFromToStepClause)
	if !ok {
		t.Fatalf("make-series range clause = %#v", makeSeries.RangeClause())
	}
	if rangeClause.MakeSeriesStepClause().Expression().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("make-series step expression = %v", rangeClause.MakeSeriesStepClause().Expression().Kind())
	}
}

func TestParseMakeSeriesOnExpressionTailRecoversRequiredInRange(t *testing.T) {
	const text = "T | make-series Hits=count() on TimeGenerated Extra from ago(1d) to now() step 1h"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: in") ||
		!diagnosticsContainMessage(diagnostics, "KS005", "Expected: range") ||
		!diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected required in-range and skipped-tail diagnostics, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	onName, ok := makeSeries.OnClause().Expression().(*NameReference)
	if !ok {
		t.Fatalf("make-series on expression = %#v", makeSeries.OnClause().Expression())
	}
	if got := tokenNameRawText(onName.Name()); got != "TimeGenerated" {
		t.Fatalf("make-series on expression = %q", got)
	}
	rangeClause, ok := makeSeries.RangeClause().(*MakeSeriesInRangeClause)
	if !ok {
		t.Fatalf("make-series range clause = %#v", makeSeries.RangeClause())
	}
	if rangeClause.Arguments() == nil || rangeClause.Arguments().Expressions().Count() != 3 {
		t.Fatalf("make-series recovered range arguments = %#v", rangeClause.Arguments())
	}
	skipped, _ := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
	if skipped == nil || skipped.Tokens().Count() == 0 || skipped.Tokens().At(0).RawText() != "to" {
		t.Fatalf("make-series skipped tokens = %#v, want to first", skipped)
	}
}

func TestParseMakeSeriesLateToClauseReportsIncompleteFragment(t *testing.T) {
	const text = "T | make-series Hits=count() on TimeGenerated from ago(1d) step 1h to now() by User"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContain(diagnostics, "KS198") {
		t.Fatalf("expected make-series late-to diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("make-series late-to should not add missing-token diagnostics, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	makeSeriesPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-series pipe = %#v", expr)
	}
	makeSeries, ok := makeSeriesPipe.Operator().(*MakeSeriesOperator)
	if !ok {
		t.Fatalf("make-series operator = %#v", makeSeriesPipe.Operator())
	}
	rangeClause, ok := makeSeries.RangeClause().(*MakeSeriesFromToStepClause)
	if !ok {
		t.Fatalf("make-series range clause = %#v", makeSeries.RangeClause())
	}
	if rangeClause.MakeSeriesToClause() != nil {
		t.Fatalf("late to clause should be skipped, got %#v", rangeClause.MakeSeriesToClause())
	}
	if rangeClause.MakeSeriesStepClause().Expression().Kind() != SyntaxKindTimespanLiteralExpression {
		t.Fatalf("make-series step expression = %v", rangeClause.MakeSeriesStepClause().Expression().Kind())
	}
	if makeSeries.ByClause() != nil {
		t.Fatalf("late by clause should be skipped, got %#v", makeSeries.ByClause())
	}
}

func TestParseQueryBuildsTopHittersOperator(t *testing.T) {
	const text = "T | top-hitters 5 of UserPrincipalName"
	expr := parsedQueryExpression(t, text)

	topHittersPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-hitters pipe = %#v", expr)
	}
	topHitters, ok := topHittersPipe.Operator().(*TopHittersOperator)
	if !ok {
		t.Fatalf("top-hitters operator = %#v", topHittersPipe.Operator())
	}
	if topHitters.TopHittersKeyword().Kind() != SyntaxKindTopHittersKeyword {
		t.Fatalf("top-hitters keyword = %v", topHitters.TopHittersKeyword().Kind())
	}
	if topHitters.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("top-hitters expression = %v", topHitters.Expression().Kind())
	}
	if topHitters.OfKeyword().Kind() != SyntaxKindOfKeyword {
		t.Fatalf("top-hitters of keyword = %v", topHitters.OfKeyword().Kind())
	}
	if topHitters.OfExpression().Kind() != SyntaxKindNameReference {
		t.Fatalf("top-hitters of expression = %v", topHitters.OfExpression().Kind())
	}
	if topHitters.ByClause() != nil {
		t.Fatalf("unexpected top-hitters by clause = %#v", topHitters.ByClause())
	}
}

func TestParseTopHittersMissingOfKeywordRecoversExpression(t *testing.T) {
	expr := parsedQueryExpression(t, "T | top-hitters 5")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-hitters pipe = %#v", expr)
	}
	topHitters, ok := pipe.Operator().(*TopHittersOperator)
	if !ok {
		t.Fatalf("top-hitters operator = %#v", pipe.Operator())
	}
	if topHitters.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("top-hitters count expression = %v", topHitters.Expression().Kind())
	}
	if topHitters.OfKeyword().RawText() != "" {
		t.Fatalf("top-hitters missing of keyword = %#v", topHitters.OfKeyword())
	}
	if topHitters.OfExpression() == nil {
		t.Fatal("top-hitters missing of expression is nil")
	}
	diagnostics := topHitters.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: of") || !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing of/expression diagnostics, got %#v", diagnostics)
	}
}

func TestParseTopHittersMissingByExpressionRecoversClause(t *testing.T) {
	expr := parsedQueryExpression(t, "T | top-hitters 5 of User by")
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("top-hitters pipe = %#v", expr)
	}
	topHitters, ok := pipe.Operator().(*TopHittersOperator)
	if !ok {
		t.Fatalf("top-hitters operator = %#v", pipe.Operator())
	}
	if topHitters.ByClause() == nil {
		t.Fatal("missing top-hitters by clause")
	}
	if topHitters.ByClause().Expression() == nil {
		t.Fatal("missing top-hitters by expression recovery node")
	}
	diagnostics := topHitters.ByClause().ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("expected missing by expression diagnostic, got %#v", diagnostics)
	}
}

func TestParseReduceEmptyKindValueDoesNotConsumeByClause(t *testing.T) {
	const text = "T | reduce kind= by Message"
	query := ParseQuery(text)
	diagnostics := query.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing value") {
		t.Fatalf("expected missing reduce kind value diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS005") {
		t.Fatalf("empty reduce kind should not be parsed as invalid kind value, got %#v", diagnostics)
	}

	expr := parsedQueryExpression(t, text)
	reducePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", expr)
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	if reduce.Parameters().Count() != 1 {
		t.Fatalf("reduce parameters = %d", reduce.Parameters().Count())
	}
	if got := namedParameterNameText(t, reduce.Parameters().At(0)); got != "kind" {
		t.Fatalf("reduce parameter 0 = %q", got)
	}
	if reduce.ByKeyword().Kind() != SyntaxKindByKeyword || reduce.ByKeyword().RawText() == "" {
		t.Fatalf("reduce by keyword = %#v", reduce.ByKeyword())
	}
	if strings.TrimSpace(reduce.Expression().Text(IncludeTriviaAll)) != "Message" {
		t.Fatalf("reduce expression = %q", reduce.Expression().Text(IncludeTriviaAll))
	}
}

func TestParseQueryReduceAcceptsLeadingWithParameters(t *testing.T) {
	const text = `T | reduce threshold=0.5 characters='abc' by Message`
	block := ParseQuery(text)
	if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	expr := parsedQueryExpression(t, text)

	reducePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", expr)
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	if reduce.Parameters().Count() != 2 {
		t.Fatalf("reduce parameters = %d", reduce.Parameters().Count())
	}
	if got := namedParameterNameText(t, reduce.Parameters().At(0)); got != "threshold" {
		t.Fatalf("reduce parameter 0 = %q", got)
	}
	if got := namedParameterNameText(t, reduce.Parameters().At(1)); got != "characters" {
		t.Fatalf("reduce parameter 1 = %q", got)
	}
}

func TestParseReduceLeadingParameterMissingEqualsKeepsValue(t *testing.T) {
	const text = `T | reduce threshold 0.5 by Message`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	reducePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", expr)
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	if reduce.Parameters().Count() != 1 {
		t.Fatalf("reduce parameters = %d", reduce.Parameters().Count())
	}
	if got := namedParameterNameText(t, reduce.Parameters().At(0)); got != "threshold" {
		t.Fatalf("reduce parameter = %q", got)
	}
	if got := strings.TrimSpace(reduce.Parameters().At(0).Expression().Text(IncludeTriviaAll)); got != "0.5" {
		t.Fatalf("reduce parameter value = %q", got)
	}
	if reduce.ByKeyword().Kind() != SyntaxKindByKeyword {
		t.Fatalf("reduce by keyword = %v", reduce.ByKeyword().Kind())
	}
	if got := strings.TrimSpace(reduce.Expression().Text(IncludeTriviaAll)); got != "Message" {
		t.Fatalf("reduce expression = %q", got)
	}
}

func TestParseCurrentReduceAndRenderWithClauseLeaveExpressionTailForStatementRecovery(t *testing.T) {
	for _, text := range []string{
		`T | reduce by User with threshold=0.1 Value | project Pattern, Count`,
		`T | render timechart with (xcolumn=Timestamp) Extra`,
	} {
		query := ParseQuery(text)
		if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: ;") || diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
			t.Fatalf("current tail diagnostics for %q = %#v", text, query.ContainedDiagnostics())
		}
	}
}

func TestParseReduceByWithClauseAllowsEmptyParameters(t *testing.T) {
	query := ParseQuery(`T | reduce by X with | project Pattern, Count`)
	if len(query.ContainedDiagnostics()) != 0 {
		t.Fatalf("unexpected empty reduce with diagnostics, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, `T | reduce by X with | project Pattern, Count`)
	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	reducePipe, ok := projectPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", projectPipe.Expression())
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	if reduce.With() == nil || reduce.With().Parameters().Count() != 0 {
		t.Fatalf("reduce with parameters = %#v", reduce.With())
	}
}

func TestParseQueryBuildsSearchOperator(t *testing.T) {
	const text = `search kind=case_sensitive datascope=hotcache in (SigninLogs, Device*) * and UserPrincipalName has "admin" | project UserPrincipalName`
	expr := parsedQueryExpression(t, text)

	projectPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("project pipe = %#v", expr)
	}
	if _, ok := projectPipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("project operator = %#v", projectPipe.Operator())
	}
	search, ok := projectPipe.Expression().(*SearchOperator)
	if !ok {
		t.Fatalf("search expression = %#v", projectPipe.Expression())
	}
	if search.Parameters().Count() != 1 {
		t.Fatalf("search parameters = %d", search.Parameters().Count())
	}
	if got := namedParameterNameText(t, search.Parameters().At(0)); got != "kind" {
		t.Fatalf("search parameter name = %q", got)
	}
	if search.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("search kind expression = %v", search.Parameters().At(0).Expression().Kind())
	}
	if search.DataScope() == nil {
		t.Fatal("missing search datascope")
	}
	if search.DataScope().Value().RawText() != "hotcache" {
		t.Fatalf("datascope value = %q", search.DataScope().Value().RawText())
	}
	if search.InClause() == nil {
		t.Fatal("missing search in clause")
	}
	if search.InClause().Expressions().Count() != 2 {
		t.Fatalf("search in expressions = %d", search.InClause().Expressions().Count())
	}
	wildcard, ok := search.InClause().Expressions().At(1).Element().(*NameReference)
	if !ok {
		t.Fatalf("search in wildcard = %#v", search.InClause().Expressions().At(1).Element())
	}
	if wildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("search wildcard kind = %v", wildcard.Name().Kind())
	}
	if search.Condition().Kind() != SyntaxKindAndExpression {
		t.Fatalf("search condition kind = %v", search.Condition().Kind())
	}
	condition, ok := search.Condition().(*BinaryExpression)
	if !ok {
		t.Fatalf("search condition = %#v", search.Condition())
	}
	if condition.Left().Kind() != SyntaxKindStarExpression {
		t.Fatalf("search condition left = %v", condition.Left().Kind())
	}
	if condition.Right().Kind() != SyntaxKindHasExpression {
		t.Fatalf("search condition right = %v", condition.Right().Kind())
	}

	recovered := ParseQuery(`search in (SigninLogs) UserPrincipalName has "admin", Extra | project UserPrincipalName`)
	if !diagnosticsContain(recovered.ContainedDiagnostics(), "KS198") {
		t.Fatalf("search predicate comma-tail diagnostics = %#v", recovered.ContainedDiagnostics())
	}
	recoveredExpr := parsedQueryExpression(t, `search in (SigninLogs) UserPrincipalName has "admin", Extra | project UserPrincipalName`)
	recoveredSearch, ok := recoveredExpr.(*SearchOperator)
	if !ok {
		t.Fatalf("recovered search expression = %#v", recoveredExpr)
	}
	if recoveredSearch.Condition().Kind() != SyntaxKindHasExpression {
		t.Fatalf("recovered search condition kind = %v", recoveredSearch.Condition().Kind())
	}
	skipped, ok := WrapGeneratedSyntaxElement(recovered.GetChild(2)).(*SkippedTokens)
	if !ok || skipped.Tokens().Count() == 0 || skipped.Tokens().At(0).RawText() != "," {
		t.Fatalf("recovered search skipped tokens = %#v", recovered.GetChild(2))
	}
}

func TestParseSearchBareInClauseRecoversMissingOperand(t *testing.T) {
	query := ParseQuery(`search in`)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: (") ||
		!diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing expression") ||
		!diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("expected bare search-in recovery diagnostics, got %#v", query.ContainedDiagnostics())
	}

	search, ok := parsedQueryExpression(t, `search in`).(*SearchOperator)
	if !ok {
		t.Fatalf("search expression = %#v", parsedQueryExpression(t, `search in`))
	}
	if search.InClause() == nil || search.InClause().Expressions().Count() != 1 {
		t.Fatalf("search in clause = %#v", search.InClause())
	}
	if !diagnosticsContainMessage(search.InClause().Expressions().At(0).Element().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("expected missing source operand diagnostic, got %#v", search.InClause().Expressions().At(0).Element().ContainedDiagnostics())
	}
}

func TestParseQueryBuildsRenderOperator(t *testing.T) {
	const text = "T | render timechart kind=stacked title='Title' with (, xcolumn=Timestamp, ycolumns=Count, OtherCount, series=UserPrincipalName, yaxis=log, accumulate=true, ymin=0)"
	expr := parsedQueryExpression(t, text)

	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	if render.ChartType().RawText() != "timechart" {
		t.Fatalf("render chart type = %q", render.ChartType().RawText())
	}
	if render.Parameters().Count() != 2 {
		t.Fatalf("deprecated render parameters = %d", render.Parameters().Count())
	}
	if got := namedParameterNameText(t, render.Parameters().At(0)); got != "kind" {
		t.Fatalf("render parameter 0 = %q", got)
	}
	if render.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("render kind expression = %v", render.Parameters().At(0).Expression().Kind())
	}
	if got := namedParameterNameText(t, render.Parameters().At(1)); got != "title" {
		t.Fatalf("render parameter 1 = %q", got)
	}
	if render.WithClause() == nil {
		t.Fatal("missing render with clause")
	}
	if render.WithClause().LeadingComma() == nil || render.WithClause().LeadingComma().Kind() != SyntaxKindCommaToken {
		t.Fatalf("render leading comma = %#v", render.WithClause().LeadingComma())
	}

	properties := render.WithClause().Properties()
	if properties.Count() != 6 {
		t.Fatalf("render with properties = %d", properties.Count())
	}
	if got := namedParameterNameText(t, properties.At(0).Element()); got != "xcolumn" {
		t.Fatalf("render property 0 = %q", got)
	}
	if properties.At(0).Element().ExpressionHint() != CompletionHintColumn {
		t.Fatalf("xcolumn expression hint = %v", properties.At(0).Element().ExpressionHint())
	}
	if got := namedParameterNameText(t, properties.At(1).Element()); got != "ycolumns" {
		t.Fatalf("render property 1 = %q", got)
	}
	ycolumns, ok := properties.At(1).Element().Expression().(*NameReferenceList)
	if !ok {
		t.Fatalf("ycolumns expression = %#v", properties.At(1).Element().Expression())
	}
	if ycolumns.Names().Count() != 2 {
		t.Fatalf("ycolumns names = %d", ycolumns.Names().Count())
	}
	if got := namedParameterNameText(t, properties.At(2).Element()); got != "series" {
		t.Fatalf("render property 2 = %q", got)
	}
	if properties.At(2).Element().Expression().Kind() != SyntaxKindNameReferenceList {
		t.Fatalf("series expression kind = %v", properties.At(2).Element().Expression().Kind())
	}
	if got := namedParameterNameText(t, properties.At(3).Element()); got != "yaxis" {
		t.Fatalf("render property 3 = %q", got)
	}
	if properties.At(3).Element().Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("yaxis expression kind = %v", properties.At(3).Element().Expression().Kind())
	}
	if got := namedParameterNameText(t, properties.At(4).Element()); got != "accumulate" {
		t.Fatalf("render property 4 = %q", got)
	}
	if properties.At(4).Element().Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("accumulate expression kind = %v", properties.At(4).Element().Expression().Kind())
	}
	if got := namedParameterNameText(t, properties.At(5).Element()); got != "ymin" {
		t.Fatalf("render property 5 = %q", got)
	}
	if properties.At(5).Element().Expression().Kind() != SyntaxKindRealLiteralExpression {
		t.Fatalf("ymin expression kind = %v", properties.At(5).Element().Expression().Kind())
	}
	if renderPipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("render source kind = %v", renderPipe.Expression().Kind())
	}
}

func TestParseRenderWithMissingPropertyEqualsKeepsValue(t *testing.T) {
	const text = `T | render timechart with (accumulate maybe, xcolumn Timestamp)`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	properties := render.WithClause().Properties()
	if properties.Count() != 2 {
		t.Fatalf("render with properties = %d", properties.Count())
	}
	if got := namedParameterNameText(t, properties.At(0).Element()); got != "accumulate" {
		t.Fatalf("render property 0 = %q", got)
	}
	if got := strings.TrimSpace(properties.At(0).Element().Expression().Text(IncludeTriviaAll)); got != "maybe" {
		t.Fatalf("render property 0 value = %q", got)
	}
	if got := namedParameterNameText(t, properties.At(1).Element()); got != "xcolumn" {
		t.Fatalf("render property 1 = %q", got)
	}
	if got := strings.TrimSpace(properties.At(1).Element().Expression().Text(IncludeTriviaAll)); got != "Timestamp" {
		t.Fatalf("render property 1 value = %q", got)
	}
}

func TestParseRenderMissingChartTypeExpectedValues(t *testing.T) {
	expr := parsedQueryExpression(t, `T | render`)
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	if render.ChartType() == nil {
		t.Fatal("missing render chart token")
	}
	diagnostics := render.ChartType().Diagnostics()
	if !diagnosticsContain(diagnostics, "KS005") {
		t.Fatalf("expected chart-type KS005 diagnostic, got %#v", diagnostics)
	}
	if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("missing chart type should not emit generic KS001, got %#v", diagnostics)
	}
}

func TestParseRenderWithTrailingCommaReportsMissingParameter(t *testing.T) {
	expr := parsedQueryExpression(t, "T | render timechart with (kind=stacked,)")
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	properties := render.WithClause().Properties()
	if properties.Count() != 2 {
		t.Fatalf("render with properties = %d", properties.Count())
	}
	if got := namedParameterNameText(t, properties.At(0).Element()); got != "kind" {
		t.Fatalf("render property 0 = %q", got)
	}
	missing := properties.At(1).Element()
	if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("trailing comma diagnostics = %#v", missing.ContainedDiagnostics())
	}
}

func TestParseRenderWithEmptyClauseKeepsEmptyPropertyList(t *testing.T) {
	expr := parsedQueryExpression(t, "T | render timechart with ()")
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	properties := render.WithClause().Properties()
	if properties.Count() != 0 {
		t.Fatalf("render with properties = %d", properties.Count())
	}
	if diagnostics := expr.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("empty render with-clause diagnostics = %#v", diagnostics)
	}
}

func TestParseRenderColumnListDoubleCommaStopsAfterMissingName(t *testing.T) {
	query := ParseQuery(`T | render timechart with (ycolumns=A,,B)`)
	var list *NameReferenceList
	WalkSyntax(query, func(element SyntaxElement) {
		if names, ok := element.(*NameReferenceList); ok {
			list = names
		}
	})
	if list == nil || list.Names().Count() != 2 {
		t.Fatalf("render double-comma name list = %#v", list)
	}
	missing := list.Names().At(1).Element()
	if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("render double-comma missing name diagnostics = %#v", missing.ContainedDiagnostics())
	}
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("render double-comma skipped-tail diagnostics = %#v", query.ContainedDiagnostics())
	}
}

func TestParseRenderColumnListTrailingCommaReportsMissingParameter(t *testing.T) {
	expr := parsedQueryExpression(t, "T | render timechart with (ycolumns=x,)")
	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	properties := render.WithClause().Properties()
	if properties.Count() != 2 {
		t.Fatalf("render with properties = %d", properties.Count())
	}
	ycolumns, ok := properties.At(0).Element().Expression().(*NameReferenceList)
	if !ok {
		t.Fatalf("render ycolumns expression = %#v", properties.At(0).Element().Expression())
	}
	if ycolumns.Names().Count() != 1 {
		t.Fatalf("render ycolumns names = %d", ycolumns.Names().Count())
	}
	missing := properties.At(1).Element()
	if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("render trailing ycolumns diagnostics = %#v", missing.ContainedDiagnostics())
	}
	if diagnosticsContain(expr.ContainedDiagnostics(), "KS001") {
		t.Fatalf("render trailing ycolumns should not add missing-token diagnostics, got %#v", expr.ContainedDiagnostics())
	}
}

func TestParseRenderColumnListEmptyValueReportsMissingName(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantList    bool
		property    string
		description string
	}{
		{
			name:        "column-list",
			query:       "T | render timechart with (ycolumns=)",
			wantList:    true,
			property:    "ycolumns",
			description: "empty ycolumns",
		},
		{
			name:        "single-column",
			query:       "T | render timechart with (xcolumn=)",
			property:    "xcolumn",
			description: "empty xcolumn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parsedQueryExpression(t, tt.query)
			renderPipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("render pipe = %#v", expr)
			}
			render, ok := renderPipe.Operator().(*RenderOperator)
			if !ok {
				t.Fatalf("render operator = %#v", renderPipe.Operator())
			}
			properties := render.WithClause().Properties()
			if properties.Count() != 1 {
				t.Fatalf("render with properties = %d", properties.Count())
			}
			property := properties.At(0).Element()
			if got := namedParameterNameText(t, property); got != tt.property {
				t.Fatalf("render property = %q", got)
			}
			if tt.wantList {
				names, ok := property.Expression().(*NameReferenceList)
				if !ok {
					t.Fatalf("render %s expression = %#v", tt.property, property.Expression())
				}
				if names.Names().Count() != 1 {
					t.Fatalf("render %s names = %d", tt.property, names.Names().Count())
				}
			}
			if !diagnosticsContainMessage(property.Expression().ContainedDiagnostics(), "KS006", "Missing name") {
				t.Fatalf("%s diagnostics = %#v", tt.description, property.Expression().ContainedDiagnostics())
			}
			if diagnosticsContain(expr.ContainedDiagnostics(), "KS001") {
				t.Fatalf("%s should not add missing-token diagnostics, got %#v", tt.description, expr.ContainedDiagnostics())
			}
		})
	}
}

func TestParseQueryOperatorParameterMissingValuesUseDeclaredKinds(t *testing.T) {
	const text = "T | render timechart with (accumulate=, ymin=, yaxis=)"
	expr := parsedQueryExpression(t, text)

	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	properties := render.WithClause().Properties()
	if properties.Count() != 3 {
		t.Fatalf("render with properties = %d", properties.Count())
	}

	cases := []struct {
		index int
		name  string
		kind  SyntaxKind
		code  string
	}{
		{0, "accumulate", SyntaxKindBooleanLiteralExpression, "KS006"},
		{1, "ymin", SyntaxKindRealLiteralExpression, "KS006"},
		{2, "yaxis", SyntaxKindTokenLiteralExpression, "KS005"},
	}
	for _, tc := range cases {
		property := properties.At(tc.index).Element()
		if got := namedParameterNameText(t, property); got != tc.name {
			t.Fatalf("property %d name = %q", tc.index, got)
		}
		if property.Expression().Kind() != tc.kind {
			t.Fatalf("%s expression kind = %v", tc.name, property.Expression().Kind())
		}
		if !diagnosticsContain(property.Expression().ContainedDiagnostics(), tc.code) {
			t.Fatalf("%s diagnostics = %#v", tc.name, property.Expression().ContainedDiagnostics())
		}
	}

	parseKvExpr := parsedQueryExpression(t, "T | parse-kv Message as (Account:string) with (pair_delimiter=)")
	parseKvPipe, ok := parseKvExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("parse-kv pipe = %#v", parseKvExpr)
	}
	parseKv, ok := parseKvPipe.Operator().(*ParseKvOperator)
	if !ok {
		t.Fatalf("parse-kv operator = %#v", parseKvPipe.Operator())
	}
	parseKvProperties := parseKv.WithClause().Properties()
	if parseKvProperties.Count() != 1 {
		t.Fatalf("parse-kv properties = %d", parseKvProperties.Count())
	}
	property := parseKvProperties.At(0).Element()
	if got := namedParameterNameText(t, property); got != "pair_delimiter" {
		t.Fatalf("parse-kv property name = %q", got)
	}
	if property.Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("parse-kv string-literal expression kind = %v", property.Expression().Kind())
	}
	if !diagnosticsContain(property.Expression().ContainedDiagnostics(), "KS006") {
		t.Fatalf("parse-kv string-literal diagnostics = %#v", property.Expression().ContainedDiagnostics())
	}
}

func TestParseQueryOperatorParameterCommaListUsesAllowedNamesAndMissingEqualsRecovery(t *testing.T) {
	renderExpr := parsedQueryExpression(t, "T | render timechart with (kind)")
	renderPipe, ok := renderExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", renderExpr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	renderProperties := render.WithClause().Properties()
	if renderProperties.Count() != 1 {
		t.Fatalf("render with properties = %d", renderProperties.Count())
	}
	renderProperty := renderProperties.At(0).Element()
	if got := namedParameterNameText(t, renderProperty); got != "kind" {
		t.Fatalf("render property name = %q", got)
	}
	if !diagnosticsContain(renderProperty.EqualToken().Diagnostics(), "KS005") {
		t.Fatalf("render missing equal diagnostics = %#v", renderProperty.EqualToken().Diagnostics())
	}
	if renderProperty.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("render missing value expression kind = %v", renderProperty.Expression().Kind())
	}
	if !diagnosticsContain(renderProperty.Expression().ContainedDiagnostics(), "KS005") {
		t.Fatalf("render missing value diagnostics = %#v", renderProperty.Expression().ContainedDiagnostics())
	}

	reduceExpr := parsedQueryExpression(t, "T | reduce by Message with hint.shufflekey=Account, threshold=0")
	reducePipe, ok := reduceExpr.(*PipeExpression)
	if !ok {
		t.Fatalf("reduce pipe = %#v", reduceExpr)
	}
	reduce, ok := reducePipe.Operator().(*ReduceByOperator)
	if !ok {
		t.Fatalf("reduce operator = %#v", reducePipe.Operator())
	}
	reduceParameters := reduce.With().Parameters()
	if reduceParameters.Count() != 2 {
		t.Fatalf("reduce with parameters = %d", reduceParameters.Count())
	}
	first := reduceParameters.At(0).Element()
	if got := namedParameterNameText(t, first); got != "hint.shufflekey" {
		t.Fatalf("reduce parameter 0 = %q", got)
	}
	if first.ExpressionHint() != CompletionHintColumn {
		t.Fatalf("hint.shufflekey expression hint = %v", first.ExpressionHint())
	}
	if first.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("hint.shufflekey expression kind = %v", first.Expression().Kind())
	}
	second := reduceParameters.At(1).Element()
	if got := namedParameterNameText(t, second); got != "threshold" {
		t.Fatalf("reduce parameter 1 = %q", got)
	}
	if second.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("threshold expression kind = %v", second.Expression().Kind())
	}
}

func TestParseQueryOperatorParameterAliasesUseExactCase(t *testing.T) {
	expr := parsedQueryExpression(t, "union __isFuzzy=true with_source=Src SigninLogs")
	union, ok := expr.(*UnionOperator)
	if !ok {
		t.Fatalf("union expression = %#v", expr)
	}
	if union.Parameters().Count() != 2 {
		t.Fatalf("union parameters = %d", union.Parameters().Count())
	}
	if got := namedParameterNameText(t, union.Parameters().At(0)); got != "__isFuzzy" {
		t.Fatalf("union parameter 0 = %q", got)
	}
	if union.Parameters().At(0).Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("__isFuzzy expression kind = %v", union.Parameters().At(0).Expression().Kind())
	}
	if got := namedParameterNameText(t, union.Parameters().At(1)); got != "with_source" {
		t.Fatalf("union parameter 1 = %q", got)
	}
	if union.Parameters().At(1).Expression().Kind() != SyntaxKindNameDeclaration {
		t.Fatalf("with_source expression kind = %v", union.Parameters().At(1).Expression().Kind())
	}
}

func TestParseQueryOperatorNameDeclarationParameterUsesRegularNameGrammar(t *testing.T) {
	valid := newQueryParserState(Lex("access"))
	validExpr := valid.parseQueryOperatorParameterValue(QueryOperatorParameters.WithSource)
	validName, ok := validExpr.(*NameDeclaration)
	if !ok {
		t.Fatalf("valid name-declaration parameter = %#v", validExpr)
	}
	if got := nameDeclarationText(t, validName); got != "access" {
		t.Fatalf("valid name-declaration text = %q", got)
	}
	if valid.pos != 1 {
		t.Fatalf("valid parser position = %d", valid.pos)
	}

	invalid := newQueryParserState(Lex("project"))
	invalidExpr := invalid.parseQueryOperatorParameterValue(QueryOperatorParameters.WithSource)
	invalidName, ok := invalidExpr.(*NameDeclaration)
	if !ok {
		t.Fatalf("invalid name-declaration parameter = %#v", invalidExpr)
	}
	if invalid.pos != 0 {
		t.Fatalf("reserved keyword should not be consumed, pos = %d", invalid.pos)
	}
	if !diagnosticsContain(invalidName.ContainedDiagnostics(), "KS006") {
		t.Fatalf("reserved keyword diagnostics = %#v", invalidName.ContainedDiagnostics())
	}

	bracketed := newQueryParserState(Lex("['project']"))
	bracketedExpr := bracketed.parseQueryOperatorParameterValue(QueryOperatorParameters.WithSource)
	bracketedName, ok := bracketedExpr.(*NameDeclaration)
	if !ok {
		t.Fatalf("bracketed name-declaration parameter = %#v", bracketedExpr)
	}
	if bracketedName.Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("bracketed name kind = %v", bracketedName.Name().Kind())
	}
}

func TestParseQueryBuildsDeprecatedRenderNoEqualsParameters(t *testing.T) {
	const text = "T | render timechart with 'Overview' by UserPrincipalName, AppDisplayName kind=unstacked"
	expr := parsedQueryExpression(t, text)

	renderPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("render pipe = %#v", expr)
	}
	render, ok := renderPipe.Operator().(*RenderOperator)
	if !ok {
		t.Fatalf("render operator = %#v", renderPipe.Operator())
	}
	if render.WithClause() != nil {
		t.Fatalf("deprecated no-equals with should not build modern with clause: %#v", render.WithClause())
	}
	if render.Parameters().Count() != 3 {
		t.Fatalf("deprecated render parameters = %d", render.Parameters().Count())
	}
	if got := namedParameterNameText(t, render.Parameters().At(0)); got != "with" {
		t.Fatalf("render parameter 0 = %q", got)
	}
	if render.Parameters().At(0).Expression().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("render with expression kind = %v", render.Parameters().At(0).Expression().Kind())
	}
	if got := namedParameterNameText(t, render.Parameters().At(1)); got != "by" {
		t.Fatalf("render parameter 1 = %q", got)
	}
	by, ok := render.Parameters().At(1).Expression().(*NameReferenceList)
	if !ok {
		t.Fatalf("render by expression = %#v", render.Parameters().At(1).Expression())
	}
	if by.Names().Count() != 2 {
		t.Fatalf("render by names = %d", by.Names().Count())
	}
	if got := namedParameterNameText(t, render.Parameters().At(2)); got != "kind" {
		t.Fatalf("render parameter 2 = %q", got)
	}
	if render.Parameters().At(2).Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("render kind expression = %v", render.Parameters().At(2).Expression().Kind())
	}
}

func TestParseQueryBuildsGraphOperatorPipeline(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id=NodeId | graph-mark-components with_component_id=Component | graph-to-table nodes with_node_id=NodeKey`
	expr := parsedQueryExpression(t, text)

	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	graphToTable, ok := graphToTablePipe.Operator().(*GraphToTableOperator)
	if !ok {
		t.Fatalf("graph-to-table operator = %#v", graphToTablePipe.Operator())
	}
	if graphToTable.OutputClause().Count() != 1 {
		t.Fatalf("graph-to-table output clauses = %d", graphToTable.OutputClause().Count())
	}
	output := graphToTable.OutputClause().At(0).Element()
	if output.EntityKeyword().Kind() != SyntaxKindNodesKeyword {
		t.Fatalf("graph-to-table entity keyword = %v", output.EntityKeyword().Kind())
	}
	if output.Parameters().Count() != 1 {
		t.Fatalf("graph-to-table parameters = %d", output.Parameters().Count())
	}
	if got := namedParameterNameText(t, output.Parameters().At(0)); got != "with_node_id" {
		t.Fatalf("graph-to-table parameter = %q", got)
	}

	markPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("graph-mark-components pipe = %#v", graphToTablePipe.Expression())
	}
	mark, ok := markPipe.Operator().(*GraphMarkComponentsOperator)
	if !ok {
		t.Fatalf("graph-mark-components operator = %#v", markPipe.Operator())
	}
	if mark.Parameters().Count() != 1 {
		t.Fatalf("graph-mark-components parameters = %d", mark.Parameters().Count())
	}
	if got := namedParameterNameText(t, mark.Parameters().At(0)); got != "with_component_id" {
		t.Fatalf("graph-mark-components parameter = %q", got)
	}

	makeGraphPipe, ok := markPipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", markPipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	if makeGraph.DirectionToken().Kind() != SyntaxKindDashDashGreaterThanToken {
		t.Fatalf("make-graph direction = %v", makeGraph.DirectionToken().Kind())
	}
	if _, ok := makeGraph.WithClause().(*MakeGraphWithImplicitIdClause); !ok {
		t.Fatalf("make-graph with clause = %#v", makeGraph.WithClause())
	}
	if makeGraphPipe.Expression().Kind() != SyntaxKindNameReference {
		t.Fatalf("make-graph source kind = %v", makeGraphPipe.Expression().Kind())
	}
}

func TestParseMakeGraphMissingImplicitIDEqualsKeepsName(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id NodeId | graph-to-table nodes`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	implicitID, ok := makeGraph.WithClause().(*MakeGraphWithImplicitIdClause)
	if !ok {
		t.Fatalf("make-graph with clause = %#v", makeGraph.WithClause())
	}
	if got := nameDeclarationText(t, implicitID.Name()); got != "NodeId" {
		t.Fatalf("make-graph implicit node id = %q", got)
	}
}

func TestParseGraphToTableMissingParameterEqualsStartsStatements(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id=NodeId | graph-to-table nodes with_node_id NodeKey`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
	}

	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 3 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	expr := firstParsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	graphToTable, ok := graphToTablePipe.Operator().(*GraphToTableOperator)
	if !ok {
		t.Fatalf("graph-to-table operator = %#v", graphToTablePipe.Operator())
	}
	output := graphToTable.OutputClause().At(0).Element()
	if output.Parameters().Count() != 0 {
		t.Fatalf("graph-to-table parameters = %d", output.Parameters().Count())
	}
}

func TestParseGraphMarkComponentsExtraTailStartsNewStatement(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id=NodeId | graph-mark-components with_component_id=Component Extra | graph-to-table nodes with_node_id=NodeKey`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") || diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("expected graph-mark-components statement-boundary diagnostic, got %#v", query.ContainedDiagnostics())
	}
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 2 {
		t.Fatalf("graph-mark-components tail statements = %#v", query.GetChild(1))
	}

	expr := firstParsedQueryExpression(t, text)
	markPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-mark-components pipe = %#v", expr)
	}
	mark, ok := markPipe.Operator().(*GraphMarkComponentsOperator)
	if !ok {
		t.Fatalf("graph-mark-components operator = %#v", markPipe.Operator())
	}
	if diagnosticsContain(mark.Diagnostics(), "KS198") {
		t.Fatalf("graph-mark-components should not own tail diagnostics: %#v", mark.Diagnostics())
	}
	if mark.Parameters().Count() != 1 {
		t.Fatalf("graph-mark-components parameters = %d", mark.Parameters().Count())
	}
}

func TestParseGraphMarkComponentsMissingParameterEqualsStartsStatements(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-mark-components with_component_id Component | graph-to-table nodes`
	query := ParseQuery(text)
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected missing-equals diagnostic, got %#v", query.ContainedDiagnostics())
	}

	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() != 3 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	markPipe, ok := firstParsedQueryExpression(t, text).(*PipeExpression)
	if !ok {
		t.Fatalf("graph-mark-components pipe = %#v", firstParsedQueryExpression(t, text))
	}
	mark, ok := markPipe.Operator().(*GraphMarkComponentsOperator)
	if !ok {
		t.Fatalf("graph-mark-components operator = %#v", markPipe.Operator())
	}
	if mark.Parameters().Count() != 0 {
		t.Fatalf("parameters = %d", mark.Parameters().Count())
	}
}

func TestParseGraphToTableMissingOutputClauseDiagnostic(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-to-table`
	query := ParseQuery(text)
	const want = "The graph-to-table operator requires 1-2 arguments (nodes, edges)."
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS230", want) {
		t.Fatalf("expected graph-to-table KS230 diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if diagnosticsContain(query.ContainedDiagnostics(), "KS001") {
		t.Fatalf("missing graph-to-table output clause should not add generic missing-token diagnostics: %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	graphToTable, ok := graphToTablePipe.Operator().(*GraphToTableOperator)
	if !ok {
		t.Fatalf("graph-to-table operator = %#v", graphToTablePipe.Operator())
	}
	if graphToTable.OutputClause().Count() != 1 {
		t.Fatalf("graph-to-table output clauses = %d", graphToTable.OutputClause().Count())
	}
	output := graphToTable.OutputClause().At(0).Element()
	if !diagnosticsContainMessage(output.Diagnostics(), "KS230", want) {
		t.Fatalf("expected recovered output clause KS230, got %#v", output.Diagnostics())
	}
	if diagnosticsContain(output.ContainedDiagnostics(), "KS001") {
		t.Fatalf("recovered output clause should only carry upstream graph-to-table diagnostics, got %#v", output.ContainedDiagnostics())
	}
}

func TestParseGraphToTableTrailingCommaReportsMissingOutputClause(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-to-table nodes,`
	query := ParseQuery(text)
	const want = "The graph-to-table operator requires 1-2 arguments (nodes, edges)."
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS230", want) {
		t.Fatalf("expected graph-to-table trailing-comma KS230 diagnostic, got %#v", query.ContainedDiagnostics())
	}
	if diagnosticsContain(query.ContainedDiagnostics(), "KS001") {
		t.Fatalf("graph-to-table trailing comma should not add generic missing-token diagnostics: %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	graphToTable, ok := graphToTablePipe.Operator().(*GraphToTableOperator)
	if !ok {
		t.Fatalf("graph-to-table operator = %#v", graphToTablePipe.Operator())
	}
	if graphToTable.OutputClause().Count() != 2 {
		t.Fatalf("graph-to-table output clauses = %d", graphToTable.OutputClause().Count())
	}
	output := graphToTable.OutputClause().At(1).Element()
	if !diagnosticsContainMessage(output.Diagnostics(), "KS230", want) {
		t.Fatalf("expected recovered trailing output clause KS230, got %#v", output.Diagnostics())
	}
	if diagnosticsContain(output.ContainedDiagnostics(), "KS001") {
		t.Fatalf("recovered trailing output clause should only carry upstream graph-to-table diagnostics, got %#v", output.ContainedDiagnostics())
	}
}

func TestParseMakeGraphEndpointTailIsNotConsumedIntoOperatorDiagnostic(t *testing.T) {
	const text = `Edges | make-graph Source Extra --> Target | graph-to-table edges`
	expr := firstParsedQueryExpression(t, text)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", expr)
	}
	makeGraph, ok := pipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", pipe.Operator())
	}
	if diagnosticsContain(makeGraph.Diagnostics(), "KS198") {
		t.Fatalf("make-graph should leave endpoint tails to QueryBlock recovery: %#v", makeGraph.Diagnostics())
	}
	if got := strings.TrimSpace(makeGraph.TargetColumn().Text(IncludeTriviaAll)); got != "Extra" {
		t.Fatalf("make-graph recovered target = %q, want Extra", got)
	}
}

func TestParseQueryBuildsMakeGraphWithTablesAndEdgesOutput(t *testing.T) {
	const text = `Edges | make-graph Source -- Target with NodesTable on NodeId | graph-to-table edges with_source_id=SrcId with_target_id=DstId`
	expr := parsedQueryExpression(t, text)

	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	graphToTable, ok := graphToTablePipe.Operator().(*GraphToTableOperator)
	if !ok {
		t.Fatalf("graph-to-table operator = %#v", graphToTablePipe.Operator())
	}
	output := graphToTable.OutputClause().At(0).Element()
	if output.EntityKeyword().Kind() != SyntaxKindGraphEdgesKeyword {
		t.Fatalf("graph-to-table entity keyword = %v", output.EntityKeyword().Kind())
	}
	if output.Parameters().Count() != 2 {
		t.Fatalf("graph-to-table edge parameters = %d", output.Parameters().Count())
	}
	if got := namedParameterNameText(t, output.Parameters().At(0)); got != "with_source_id" {
		t.Fatalf("graph-to-table source parameter = %q", got)
	}
	if got := namedParameterNameText(t, output.Parameters().At(1)); got != "with_target_id" {
		t.Fatalf("graph-to-table target parameter = %q", got)
	}

	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	if makeGraph.DirectionToken().Kind() != SyntaxKindDashDashToken {
		t.Fatalf("make-graph direction = %v", makeGraph.DirectionToken().Kind())
	}
	withTables, ok := makeGraph.WithClause().(*MakeGraphWithTablesAndKeysClause)
	if !ok {
		t.Fatalf("make-graph with tables clause = %#v", makeGraph.WithClause())
	}
	if withTables.TablesAndKeys().Count() != 1 {
		t.Fatalf("make-graph with tables count = %d", withTables.TablesAndKeys().Count())
	}
	if withTables.TablesAndKeys().At(0).Element().OnKeyword().Kind() != SyntaxKindOnKeyword {
		t.Fatalf("make-graph with tables on keyword = %v", withTables.TablesAndKeys().At(0).Element().OnKeyword().Kind())
	}
}

func TestParseMakeGraphTrailingTableCommaBeforePartitionedBy(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with Nodes on NodeId, partitioned-by Tenant (Edges | project Source, Target, Tenant) | graph-to-table edges`
	query := ParseQuery(text)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("expected missing table/key clause diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	withTables, ok := makeGraph.WithClause().(*MakeGraphWithTablesAndKeysClause)
	if !ok {
		t.Fatalf("make-graph with tables clause = %#v", makeGraph.WithClause())
	}
	if withTables.TablesAndKeys().Count() != 2 {
		t.Fatalf("make-graph with tables count = %d", withTables.TablesAndKeys().Count())
	}
	if !diagnosticsContainMessage(withTables.TablesAndKeys().At(1).Element().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("recovered table/key diagnostics = %#v", withTables.TablesAndKeys().At(1).Element().Diagnostics())
	}
}

func TestParseMakeGraphPartitionedBySourcePipeSubquery(t *testing.T) {
	const text = `Edges | make-graph Source --> Target partitioned-by Tenant (OtherEdges | count) | graph-to-table edges`
	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	if makeGraph.PartitionedByClause() == nil {
		t.Fatal("make-graph partitioned-by clause is nil")
	}
	if _, ok := makeGraph.PartitionedByClause().Subquery().(*PipeExpression); !ok {
		t.Fatalf("make-graph partitioned-by subquery = %#v", makeGraph.PartitionedByClause().Subquery())
	}
	if diagnostics := makeGraph.PartitionedByClause().ContainedDiagnostics(); !diagnosticsContain(diagnostics, "KS176") {
		t.Fatalf("make-graph partitioned-by source root diagnostics = %#v", diagnostics)
	}
}

func TestParseMakeGraphPartitionedByLeavesUnexpectedEntityTailForStatementRecovery(t *testing.T) {
	const query = `Edges | make-graph Source --> Target partitioned-by Tenant Extra (Edges | project Source, Target, Tenant) | graph-to-table edges`
	diagnostics := Parse(query).GetDiagnostics()
	want := []Diagnostic{
		DiagnosticTermsExpected("(").WithLocation(59, 5),
		DiagnosticQueryOperatorExpected().WithLocation(59, 5),
		DiagnosticTermsExpected(")").WithLocation(65, 1),
		DiagnosticTermsExpected(";").WithLocation(65, 1),
	}
	if !reflect.DeepEqual(diagnostics, want) {
		t.Fatalf("diagnostics = %#v, want %#v", diagnostics, want)
	}
}

func TestParseMakeGraphPartitionedByRejectsBareSourceSubquery(t *testing.T) {
	const text = `Edges | make-graph Source --> Target partitioned-by Tenant (OtherEdges) | graph-to-table edges`
	code := Parse(text)
	if !diagnosticsContain(code.GetDiagnostics(), "KS176") {
		t.Fatalf("expected make-graph partitioned-by bare-source diagnostic, got %#v", code.GetDiagnostics())
	}
}

func TestParseMakeGraphWithTableFunctionOperand(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with NodesFn("active", 10) on NodeId | graph-to-table nodes`
	expr := parsedQueryExpression(t, text)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	withTables, ok := makeGraph.WithClause().(*MakeGraphWithTablesAndKeysClause)
	if !ok {
		t.Fatalf("make-graph with tables clause = %#v", makeGraph.WithClause())
	}
	clause := withTables.TablesAndKeys().At(0).Element()
	if _, ok := clause.Table().(*FunctionCallExpression); !ok {
		t.Fatalf("make-graph table operand = %#v", clause.Table())
	}
	if clause.OnKeyword().Kind() != SyntaxKindOnKeyword {
		t.Fatalf("make-graph with tables on keyword = %v", clause.OnKeyword().Kind())
	}
}

func TestParseMakeGraphInvalidFunctionOperandStillFlagsSyntax(t *testing.T) {
	query := ParseQuery(`Edges | make-graph Source --> Target with NodesFn() NodeId | graph-to-table nodes`)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: on") {
		t.Fatalf("expected missing on diagnostic, got %#v", query.ContainedDiagnostics())
	}

	expr := parsedQueryExpression(t, `Edges | make-graph Source --> Target with NodesFn() NodeId | graph-to-table nodes`)
	graphToTablePipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-to-table pipe = %#v", expr)
	}
	makeGraphPipe, ok := graphToTablePipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("make-graph pipe = %#v", graphToTablePipe.Expression())
	}
	makeGraph, ok := makeGraphPipe.Operator().(*MakeGraphOperator)
	if !ok {
		t.Fatalf("make-graph operator = %#v", makeGraphPipe.Operator())
	}
	withTables, ok := makeGraph.WithClause().(*MakeGraphWithTablesAndKeysClause)
	if !ok {
		t.Fatalf("make-graph with tables clause = %#v", makeGraph.WithClause())
	}
	if _, ok := withTables.TablesAndKeys().At(0).Element().Table().(*FunctionCallExpression); !ok {
		t.Fatalf("make-graph table operand = %#v", withTables.TablesAndKeys().At(0).Element().Table())
	}
}

func TestParseMakeGraphMissingDirectionListsBothDirections(t *testing.T) {
	query := ParseQuery(`Edges | make-graph Source Target with Nodes on NodeId`)
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected one of: -->, --") {
		t.Fatalf("expected make-graph direction diagnostic, got %#v", query.ContainedDiagnostics())
	}
}

func TestParseQueryBuildsGraphMatchOperator(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id=NodeId | graph-match cycles=all (a)-[e*1..3]->(b) where e.Weight > 10 project Src=a.NodeId, Dst=b.NodeId, Path=e.Weight`
	expr := parsedQueryExpression(t, text)

	matchPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-match pipe = %#v", expr)
	}
	match, ok := matchPipe.Operator().(*GraphMatchOperator)
	if !ok {
		t.Fatalf("graph-match operator = %#v", matchPipe.Operator())
	}
	if match.Parameters().Count() != 1 {
		t.Fatalf("graph-match parameters = %d", match.Parameters().Count())
	}
	if got := namedParameterNameText(t, match.Parameters().At(0)); got != "cycles" {
		t.Fatalf("graph-match parameter = %q", got)
	}
	if match.Patterns().Count() != 1 {
		t.Fatalf("graph-match patterns = %d", match.Patterns().Count())
	}
	pattern := match.Patterns().At(0).Element()
	if pattern.PatternElements().Count() != 3 {
		t.Fatalf("graph-match pattern elements = %d", pattern.PatternElements().Count())
	}
	if _, ok := pattern.PatternElements().At(0).(*GraphMatchPatternNode); !ok {
		t.Fatalf("graph-match first element = %#v", pattern.PatternElements().At(0))
	}
	edge, ok := pattern.PatternElements().At(1).(*GraphMatchPatternEdge)
	if !ok {
		t.Fatalf("graph-match edge element = %#v", pattern.PatternElements().At(1))
	}
	if edge.FirstToken().Kind() != SyntaxKindDashBracketToken || edge.LastToken().Kind() != SyntaxKindBracketDashGreaterThanToken {
		t.Fatalf("graph-match edge tokens = %v/%v", edge.FirstToken().Kind(), edge.LastToken().Kind())
	}
	if edge.Range() == nil || edge.Range().Asterisk().Kind() != SyntaxKindAsteriskToken {
		t.Fatalf("graph-match edge range = %#v", edge.Range())
	}
	if match.WhereClause() == nil || match.WhereClause().Condition().Kind() != SyntaxKindGreaterThanExpression {
		t.Fatalf("graph-match where = %#v", match.WhereClause())
	}
	if match.ProjectClause() == nil || match.ProjectClause().Expressions().Count() != 3 {
		t.Fatalf("graph-match project = %#v", match.ProjectClause())
	}
}

func TestParseGraphMatchEdgeRangeAcceptsSignedInvocationBounds(t *testing.T) {
	cases := []struct {
		name          string
		text          string
		wantStartKind SyntaxKind
		wantEndKind   SyntaxKind
	}{
		{
			name:          "graph-match signed start",
			text:          `Edges | make-graph Source --> Target | graph-match (a)-[e*+1..3]->(b) project Src=a.Source`,
			wantStartKind: SyntaxKindUnaryPlusExpression,
			wantEndKind:   SyntaxKindLongLiteralExpression,
		},
		{
			name:          "graph-shortest signed start",
			text:          `Edges | make-graph Source --> Target | graph-shortest-paths (a)-[e*-1..3]->(b) project Src=a.Source`,
			wantStartKind: SyntaxKindUnaryMinusExpression,
			wantEndKind:   SyntaxKindLongLiteralExpression,
		},
		{
			name:          "signed end",
			text:          `Edges | make-graph Source --> Target | graph-match (a)-[e*1..-3]->(b) project Src=a.Source`,
			wantStartKind: SyntaxKindLongLiteralExpression,
			wantEndKind:   SyntaxKindUnaryMinusExpression,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.text)
			if len(query.ContainedDiagnostics()) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", query.ContainedDiagnostics())
			}
			rangeNode := parsedGraphEdgeRange(t, tc.text)
			if rangeNode.RangeStart().Kind() != tc.wantStartKind {
				t.Fatalf("range start kind = %v, want %v", rangeNode.RangeStart().Kind(), tc.wantStartKind)
			}
			if rangeNode.RangeEnd().Kind() != tc.wantEndKind {
				t.Fatalf("range end kind = %v, want %v", rangeNode.RangeEnd().Kind(), tc.wantEndKind)
			}
		})
	}
}

func TestParseGraphMatchEdgeRangeMissingBounds(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "missing start",
			text: `Edges | make-graph Source --> Target | graph-match (a)-[e*..3]->(b) project Src=a.Source`,
		},
		{
			name: "missing end",
			text: `Edges | make-graph Source --> Target | graph-match (a)-[e*1..]->(b) project Src=a.Source`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.text)
			if !diagnosticsContain(query.ContainedDiagnostics(), "KS006") {
				t.Fatalf("expected missing-bound diagnostic, got %#v", query.ContainedDiagnostics())
			}
		})
	}
}

func parsedGraphEdgeRange(t *testing.T, text string) *GraphMatchPatternEdgeRange {
	t.Helper()
	expr := parsedQueryExpression(t, text)
	matchPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph operator pipe = %#v", expr)
	}
	var patterns *SeparatedSyntaxListOf[*GraphMatchPattern]
	switch operator := matchPipe.Operator().(type) {
	case *GraphMatchOperator:
		patterns = operator.Patterns()
	case *GraphShortestPathsOperator:
		patterns = operator.Patterns()
	default:
		t.Fatalf("graph operator = %#v", matchPipe.Operator())
	}
	if patterns == nil || patterns.Count() != 1 {
		t.Fatalf("patterns = %#v", patterns)
	}
	pattern := patterns.At(0).Element()
	if pattern.PatternElements().Count() != 3 {
		t.Fatalf("pattern elements = %d", pattern.PatternElements().Count())
	}
	edge, ok := pattern.PatternElements().At(1).(*GraphMatchPatternEdge)
	if !ok {
		t.Fatalf("edge element = %#v", pattern.PatternElements().At(1))
	}
	if edge.Range() == nil {
		t.Fatalf("edge range = nil")
	}
	return edge.Range()
}

func TestParseGraphMatchWhereClauseAcceptsQueryExpression(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-match (a) where print x=1 project x`
	expr := parsedQueryExpression(t, text)

	matchPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-match pipe = %#v", expr)
	}
	match, ok := matchPipe.Operator().(*GraphMatchOperator)
	if !ok {
		t.Fatalf("graph-match operator = %#v", matchPipe.Operator())
	}
	if _, ok := match.WhereClause().Condition().(*PrintOperator); !ok {
		t.Fatalf("graph-match where condition = %#v", match.WhereClause().Condition())
	}
	if diagnosticsContainCode(match.WhereClause().Condition().ContainedDiagnostics(), "KS006") {
		t.Fatalf("graph-match query where condition has missing-expression diagnostics: %#v", match.WhereClause().Condition().ContainedDiagnostics())
	}
	if match.ProjectClause() == nil || match.ProjectClause().Expressions().Count() != 1 {
		t.Fatalf("graph-match project = %#v", match.ProjectClause())
	}
}

func TestParseQueryBuildsGraphShortestPathsOperator(t *testing.T) {
	const text = `Edges | make-graph Source --> Target with_node_id=NodeId | graph-shortest-paths output=all (start)-->(finish) project Src=start.NodeId, Dst=finish.NodeId`
	expr := parsedQueryExpression(t, text)

	shortestPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-shortest-paths pipe = %#v", expr)
	}
	shortest, ok := shortestPipe.Operator().(*GraphShortestPathsOperator)
	if !ok {
		t.Fatalf("graph-shortest-paths operator = %#v", shortestPipe.Operator())
	}
	if shortest.Parameters().Count() != 1 {
		t.Fatalf("graph-shortest-paths parameters = %d", shortest.Parameters().Count())
	}
	if got := namedParameterNameText(t, shortest.Parameters().At(0)); got != "output" {
		t.Fatalf("graph-shortest-paths parameter = %q", got)
	}
	if shortest.Patterns().Count() != 1 {
		t.Fatalf("graph-shortest-paths patterns = %d", shortest.Patterns().Count())
	}
	pattern := shortest.Patterns().At(0).Element()
	if pattern.PatternElements().Count() != 3 {
		t.Fatalf("graph-shortest-paths pattern elements = %d", pattern.PatternElements().Count())
	}
	edge, ok := pattern.PatternElements().At(1).(*GraphMatchPatternEdge)
	if !ok || edge.FirstToken().Kind() != SyntaxKindDashDashGreaterThanToken {
		t.Fatalf("graph-shortest-paths edge = %#v", pattern.PatternElements().At(1))
	}
	if shortest.ProjectClause() == nil || shortest.ProjectClause().Expressions().Count() != 2 {
		t.Fatalf("graph-shortest-paths project = %#v", shortest.ProjectClause())
	}
}

func TestParseGraphShortestPathsWhereClauseAcceptsPipedQueryExpression(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-shortest-paths (start)-->(finish) where range x from 1 to 2 step 1 | count project Count`
	expr := parsedQueryExpression(t, text)

	shortestPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-shortest-paths pipe = %#v", expr)
	}
	shortest, ok := shortestPipe.Operator().(*GraphShortestPathsOperator)
	if !ok {
		t.Fatalf("graph-shortest-paths operator = %#v", shortestPipe.Operator())
	}
	wherePipe, ok := shortest.WhereClause().Condition().(*PipeExpression)
	if !ok {
		t.Fatalf("graph-shortest-paths where condition = %#v", shortest.WhereClause().Condition())
	}
	if _, ok := wherePipe.Operator().(*CountOperator); !ok {
		t.Fatalf("graph-shortest-paths where pipe operator = %#v", wherePipe.Operator())
	}
	if shortest.ProjectClause() == nil || shortest.ProjectClause().Expressions().Count() != 1 {
		t.Fatalf("graph-shortest-paths project = %#v", shortest.ProjectClause())
	}
}

func TestParseGraphWherePipedProjectBelongsToWhereExpression(t *testing.T) {
	const text = `Edges | make-graph Source --> Target | graph-match (a)-[e]->(b) where e.Weight > 10 | project Source=a.Source`
	expr := parsedQueryExpression(t, text)

	matchPipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("graph-match pipe = %#v", expr)
	}
	match, ok := matchPipe.Operator().(*GraphMatchOperator)
	if !ok {
		t.Fatalf("graph-match operator = %#v", matchPipe.Operator())
	}
	wherePipe, ok := match.WhereClause().Condition().(*PipeExpression)
	if !ok {
		t.Fatalf("graph-match where condition = %#v", match.WhereClause().Condition())
	}
	if _, ok := wherePipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("graph-match where piped operator = %#v", wherePipe.Operator())
	}
	if match.ProjectClause() != nil {
		t.Fatalf("bare graph project clause should be nil when project is piped inside where expression: %#v", match.ProjectClause())
	}
}

func TestParseGraphMatchWhereCommaTailStopsBeforeProject(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		shortest  bool
		wantWhere string
	}{
		{
			name:      "graph-match",
			query:     `Edges | make-graph Source --> Target | graph-match (a)-[e]->(b) where a.Source == "x", Extra project Src=a.Source`,
			wantWhere: ` a.Source == "x"`,
		},
		{
			name:      "graph-shortest-paths",
			query:     `Edges | make-graph Source --> Target | graph-shortest-paths (a)-->(b) where a.Source == "x", Extra project Src=a.Source`,
			shortest:  true,
			wantWhere: ` a.Source == "x"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.query)
			if !diagnosticsContain(code.GetDiagnostics(), "KS198") {
				t.Fatalf("expected incomplete-fragment diagnostic, got %#v", code.GetDiagnostics())
			}
			if diagnosticsContain(code.GetDiagnostics(), "KS001") {
				t.Fatalf("comma-tail recovery should not add generic syntax diagnostics: %#v", code.GetDiagnostics())
			}
			expr := parsedQueryExpression(t, tt.query)
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("graph pipe = %#v", expr)
			}
			var where *WhereClause
			var project *ProjectClause
			if tt.shortest {
				shortest, ok := pipe.Operator().(*GraphShortestPathsOperator)
				if !ok {
					t.Fatalf("graph-shortest-paths operator = %#v", pipe.Operator())
				}
				where = shortest.WhereClause()
				project = shortest.ProjectClause()
			} else {
				match, ok := pipe.Operator().(*GraphMatchOperator)
				if !ok {
					t.Fatalf("graph-match operator = %#v", pipe.Operator())
				}
				where = match.WhereClause()
				project = match.ProjectClause()
			}
			if where == nil || where.Condition() == nil {
				t.Fatalf("where clause = %#v", where)
			}
			if got := where.Condition().Text(IncludeTriviaAll); got != tt.wantWhere {
				t.Fatalf("where condition text = %q, want %q", got, tt.wantWhere)
			}
			if project != nil {
				t.Fatalf("project clause should remain in skipped tail: %#v", project)
			}
		})
	}
}

func TestParseGraphWhereReportsUnexpectedPredicateTail(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantText string
		wantCode string
	}{
		{
			name:     "edges comma",
			query:    `Edges | graph-where-edges Weight > 1, Extra`,
			wantText: ` Weight > 1`,
			wantCode: "KS198",
		},
		{
			name:     "nodes whitespace",
			query:    `Edges | graph-where-nodes Label == "a" Extra`,
			wantText: ` Label == "a"`,
			wantCode: "KS005",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.query)
			if !diagnosticsContain(code.GetDiagnostics(), tt.wantCode) {
				t.Fatalf("expected %s tail diagnostic, got %#v", tt.wantCode, code.GetDiagnostics())
			}
			expr := firstParsedQueryExpression(t, tt.query)
			pipe, ok := expr.(*PipeExpression)
			if !ok {
				t.Fatalf("graph-where pipe = %#v", expr)
			}
			var condition Expression
			switch op := pipe.Operator().(type) {
			case *GraphWhereEdgesOperator:
				condition = op.Condition()
			case *GraphWhereNodesOperator:
				condition = op.Condition()
			default:
				t.Fatalf("graph-where operator = %#v", pipe.Operator())
			}
			if got := condition.Text(IncludeTriviaAll); got != tt.wantText {
				t.Fatalf("condition text = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestParseGraphPatternTrailingCommaReportsMissingPattern(t *testing.T) {
	t.Run("graph-match before project", func(t *testing.T) {
		const text = `Edges | make-graph Source --> Target | graph-match (a), project Src=a.Source`
		expr := parsedQueryExpression(t, text)
		matchPipe, ok := expr.(*PipeExpression)
		if !ok {
			t.Fatalf("graph-match pipe = %#v", expr)
		}
		match, ok := matchPipe.Operator().(*GraphMatchOperator)
		if !ok {
			t.Fatalf("graph-match operator = %#v", matchPipe.Operator())
		}
		if match.Patterns().Count() != 2 {
			t.Fatalf("graph-match patterns = %d", match.Patterns().Count())
		}
		missing := match.Patterns().At(1).Element()
		if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS217", "Missing graph-match pattern.") {
			t.Fatalf("missing graph-match pattern diagnostics = %#v", missing.ContainedDiagnostics())
		}
		if match.ProjectClause() == nil || match.ProjectClause().Expressions().Count() != 1 {
			t.Fatalf("graph-match project = %#v", match.ProjectClause())
		}
	})

	t.Run("graph-shortest-paths before project", func(t *testing.T) {
		const text = `Edges | make-graph Source --> Target | graph-shortest-paths (start), project Src=start.Source`
		expr := parsedQueryExpression(t, text)
		shortestPipe, ok := expr.(*PipeExpression)
		if !ok {
			t.Fatalf("graph-shortest-paths pipe = %#v", expr)
		}
		shortest, ok := shortestPipe.Operator().(*GraphShortestPathsOperator)
		if !ok {
			t.Fatalf("graph-shortest-paths operator = %#v", shortestPipe.Operator())
		}
		if shortest.Patterns().Count() != 2 {
			t.Fatalf("graph-shortest-paths patterns = %d", shortest.Patterns().Count())
		}
		missing := shortest.Patterns().At(1).Element()
		if !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS217", "Missing graph-match pattern.") {
			t.Fatalf("missing graph-shortest-paths pattern diagnostics = %#v", missing.ContainedDiagnostics())
		}
		if shortest.ProjectClause() == nil || shortest.ProjectClause().Expressions().Count() != 1 {
			t.Fatalf("graph-shortest-paths project = %#v", shortest.ProjectClause())
		}
	})

	for _, tt := range []struct {
		name string
		text string
	}{
		{
			name: "graph-match missing pattern",
			text: `Edges | make-graph Source --> Target | graph-match`,
		},
		{
			name: "graph-shortest-paths missing pattern",
			text: `Edges | make-graph Source --> Target | graph-shortest-paths`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			diagnostics := query.ContainedDiagnostics()
			if !diagnosticsContainMessage(diagnostics, "KS217", "Missing graph-match pattern.") {
				t.Fatalf("missing graph pattern diagnostics = %#v", diagnostics)
			}
			for _, code := range []string{"KS001", "KS006", "KS238"} {
				if diagnosticsContain(diagnostics, code) {
					t.Fatalf("unexpected %s in missing graph pattern diagnostics: %#v", code, diagnostics)
				}
			}
		})
	}
}
