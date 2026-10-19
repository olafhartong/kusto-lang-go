package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryParserFacadeCoversTextAndTokenEntryPoints(t *testing.T) {
	options := kustolang.DefaultParseOptions

	query := kustolang.QueryParser.ParseQueryWithOptions("print x=1", options)
	if query == nil || query.Text(kustolang.IncludeTriviaAll) != "print x=1" {
		t.Fatalf("query facade result = %#v", query)
	}

	expressionTokens := kustolang.Lex("skip 1 + 2")
	expression := kustolang.QueryParser.ParseExpressionTokensFromWithOptions(expressionTokens, 1, options)
	if expression == nil || expression.Text(kustolang.IncludeTriviaAll) != " 1 + 2" {
		t.Fatalf("expression facade result = %#v", expression)
	}

	literal := kustolang.QueryParser.ParseLiteral("dynamic(null)")
	if literal == nil || !literal.IsLiteral() || literal.LiteralValueInfo() == nil {
		t.Fatalf("literal facade result = %#v", literal)
	}

	if path := kustolang.QueryParser.ParseEntityPath(`cluster("help").database("Samples")`); path == nil {
		t.Fatal("entity-path facade returned nil")
	}
	if group := kustolang.QueryParser.ParseEntityGroup("entity_group [T]"); group == nil {
		t.Fatal("entity-group facade returned nil")
	}
	if schema := kustolang.QueryParser.ParseRowSchema("(A:string, B:long)"); schema == nil {
		t.Fatal("row-schema facade returned nil")
	}
	if parameters := kustolang.QueryParser.ParseFunctionParameters("(x:string, y:long)"); parameters == nil {
		t.Fatal("function-parameters facade returned nil")
	}
	if parameter := kustolang.QueryParser.ParseFunctionParameter("x:string"); parameter == nil {
		t.Fatal("function-parameter facade returned nil")
	}
	if body := kustolang.QueryParser.ParseFunctionBody("{ print x=1 }"); body == nil {
		t.Fatal("function-body facade returned nil")
	}
}

func TestPublicQueryParserFacadeRecoversMalformedInput(t *testing.T) {
	query := kustolang.QueryParser.ParseQuery("T | where (")
	if query == nil || len(query.GetContainedSyntaxDiagnostics()) == 0 {
		t.Fatal("malformed query was not recovered with syntax diagnostics")
	}
}
