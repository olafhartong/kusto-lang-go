package kustolang

import (
	"reflect"
	"testing"
)

func TestKustoFactsPublicStringLiteralParity(t *testing.T) {
	for _, test := range []struct {
		name string
		got  string
		want string
	}{
		{name: "single quoted", got: GetStringLiteral("alpha"), want: `'alpha'`},
		{name: "quote selects double", got: GetStringLiteral("can't"), want: `"can't"`},
		{name: "single escapes", got: GetSingleQuotedStringLiteral("a\n'b\\c"), want: `'a\n\'b\\c'`},
		{name: "vertical tab is not encoded", got: GetSingleQuotedStringLiteral("\v"), want: "'\v'"},
		{name: "multiline", got: GetMultiLineStringLiteral("a\nb"), want: "```a\nb```"},
		{name: "hidden prefix", got: GetStringLiteralValue(`h'a\n'`), want: "a\n"},
		{name: "verbatim doubled quote", got: GetStringLiteralValue(`@'can''t'`), want: "can't"},
		{name: "bracketed", got: GetStringLiteralValue(`['column name']`), want: `'column name'`},
		{name: "unicode", got: GetStringLiteralValue(`'\u0041\U0001F642'`), want: "A🙂"},
		{name: "short hex", got: GetStringLiteralValue(`'\x4'`), want: string(rune(4))},
		{name: "source octal digit handling", got: GetStringLiteralValue(`'\128'`), want: "X"},
		{name: "incomplete multiline", got: GetStringLiteralValue("```unterminated"), want: "unterminated"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("value = %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestKustoFactsPublicHostParity(t *testing.T) {
	if !IsHostName("HELP.KUSTO.WINDOWS.NET", "help.kusto.windows.net") {
		t.Fatal("host comparison is not ordinal-ignore-case")
	}
	if got := GetHostName("https://help.kusto.windows.net:443/Samples?x=1"); got != "help.kusto.windows.net" {
		t.Fatalf("host = %q", got)
	}
	if host, path := GetHostAndPath("https://help.kusto.windows.net:443/Samples?x=1"); host != "help.kusto.windows.net" || path != "Samples" {
		t.Fatalf("host/path = %q/%q", host, path)
	}
	if got := GetHostName("help_cluster"); got != "help_cluster" {
		t.Fatalf("plain host = %q", got)
	}
	if got := GetHostName("https://münchen.example/Samples"); got != "münchen.example" {
		t.Fatalf("unicode host = %q", got)
	}
	if got := GetFullHostName("help.west", KustoWindowsNet); got != "help.west"+KustoWindowsNet {
		t.Fatalf("full host = %q", got)
	}
	if GetFullHostName("help.west.europe", KustoWindowsNet) != "help.west.europe" {
		t.Fatal("multi-dot host was treated as short")
	}
	full := "help.west" + KustoWindowsNet
	if !HasShortHostName(full, KustoWindowsNet) || !IsShortHostName("HELP.WEST", full, KustoWindowsNet) || GetShortHostName(full, KustoWindowsNet) != "help.west" {
		t.Fatalf("short-host helpers rejected %q", full)
	}
	if HasShortHostName("help.example.com", KustoWindowsNet) || GetShortHostName("help.example.com", KustoWindowsNet) != "" {
		t.Fatal("unrelated domain has a short host")
	}
}

func TestKustoFactsHasPipedInputParity(t *testing.T) {
	piped := ParseQuery("T | project A")
	project := firstBoundNode[*ProjectOperator](t, piped)
	if !HasPipedInput(project) {
		t.Fatal("ordinary pipe operator has no piped input")
	}

	root := ParseQuery("project A")
	if HasPipedInput(firstBoundNode[*ProjectOperator](t, root)) {
		t.Fatal("root operator has piped input")
	}

	fork := ParseQuery("T | fork (where A > 0)")
	if !HasPipedInput(firstBoundNode[*FilterOperator](t, fork)) {
		t.Fatal("fork branch root has no implicit piped input")
	}
}

func TestKustoFactsSyntaxTabularityParity(t *testing.T) {
	for _, test := range []struct {
		text string
		want Tabularity
	}{
		{text: "1 + 2", want: TabularityScalar},
		{text: "datatable(A:long)[1]", want: TabularityTabular},
		{text: `strlen("x")`, want: TabularityScalar},
		{text: "MissingFunction()", want: TabularityUnknown},
	} {
		expression := ParseExpression(test.text)
		if got := GetSyntaxTabularity(expression); got != test.want {
			t.Fatalf("GetSyntaxTabularity(%q) = %s, want %s", test.text, got, test.want)
		}
	}

	query := ParseQuery("let x=1; print x")
	if got := GetSyntaxTabularity(query.GetChild(1)); got != TabularityTabular {
		t.Fatalf("statement-list tabularity = %s", got)
	}
	if got := GetSyntaxTabularity(firstBoundNode[*LetStatement](t, query)); got != TabularityNone {
		t.Fatalf("declaration tabularity = %s", got)
	}
}

func TestKustoFactsLiteralTypesAndExpressionNamesParity(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		want *ScalarSymbol
	}{
		{SyntaxKindBooleanLiteralToken, ScalarTypes.Bool},
		{SyntaxKindStringLiteralExpression, ScalarTypes.String},
		{SyntaxKindIntLiteralToken, ScalarTypes.Int},
		{SyntaxKindLongLiteralExpression, ScalarTypes.Long},
		{SyntaxKindRealLiteralToken, ScalarTypes.Real},
		{SyntaxKindDecimalLiteralExpression, ScalarTypes.Decimal},
		{SyntaxKindTimespanLiteralToken, ScalarTypes.TimeSpan},
		{SyntaxKindDateTimeLiteralExpression, ScalarTypes.DateTime},
		{SyntaxKindRawGuidLiteralToken, ScalarTypes.Guid},
		{SyntaxKindTypeOfLiteralExpression, ScalarTypes.TypeLit},
		{SyntaxKindDynamicExpression, ScalarTypes.Dynamic},
		{SyntaxKindList, ScalarTypes.Unknown},
	}
	for _, test := range tests {
		if got := GetLiteralType(test.kind); got != test.want {
			t.Fatalf("GetLiteralType(%v) = %#v, want %#v", test.kind, got, test.want)
		}
	}

	if got := GetExpressionResultName(ParseExpression("Account"), "fallback", nil); got != "Account" {
		t.Fatalf("name result = %q", got)
	}
	if got := GetExpressionResultName(ParseExpression("1 + 2"), "fallback", nil); got != "fallback" {
		t.Fatalf("default result = %q", got)
	}
	if got, want := []string{MultiLineStringQuote, AlternateMultiLineStringQuote}, []string{"```", "~~~"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("multiline quotes = %#v", got)
	}
}
