package kustolang

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestKustoQuickInfoDisplaysSymbolsSignaturesTypesAndDescriptions(t *testing.T) {
	account := NewColumnSymbolFull("Account", ScalarTypes.String.Type(), "User account", nil, nil, nil)
	table := NewTableSymbolWithDescription("Events", []*ColumnSymbol{account, NewColumnSymbol("Score", ScalarTypes.Long.Type())}, "Security events")
	database := NewDatabaseSymbol("db", table)
	service := NewKustoCodeService("Events | project Account, Length=strlen(Account)", NewGlobalState(database))

	tests := []struct {
		needle string
		kind   QuickInfoKind
		want   []string
	}{
		{needle: "Events", kind: QuickInfoKindTable, want: []string{"Events: table(Account, Score)", "Security events"}},
		{needle: "Account", kind: QuickInfoKindColumn, want: []string{"Account: string", "User account"}},
		{needle: "strlen", kind: QuickInfoKindBuiltInFunction, want: []string{"strlen(string): long"}},
	}
	for _, test := range tests {
		t.Run(test.needle, func(t *testing.T) {
			position := strings.Index(service.Text(), test.needle) + 1
			info := service.GetQuickInfo(position, QuickInfoOptionsDefault, CancellationToken{})
			if len(info.Items) == 0 || info.Items[0].Kind != test.kind {
				t.Fatalf("quick info = %#v, want kind %v", info, test.kind)
			}
			for _, want := range test.want {
				if !strings.Contains(info.Items[0].Text(), want) {
					t.Fatalf("quick info text = %q, want %q", info.Items[0].Text(), want)
				}
			}
		})
	}
}

func TestKustoQuickInfoHonorsDescriptionAndDiagnosticOptions(t *testing.T) {
	account := NewColumnSymbolFull("Account", ScalarTypes.String.Type(), "User account", nil, nil, nil)
	service := NewKustoCodeService("Events | project Account, Missing", NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", account))))

	accountPosition := strings.Index(service.Text(), "Account") + 1
	withoutDescription := service.GetQuickInfo(accountPosition, QuickInfoOptionsDefault.WithShowDescriptions(false), CancellationToken{})
	if len(withoutDescription.Items) != 1 || strings.Contains(withoutDescription.Items[0].Text(), "User account") {
		t.Fatalf("description-disabled quick info = %#v", withoutDescription)
	}

	missingPosition := strings.Index(service.Text(), "Missing") + 1
	withDiagnostic := service.GetQuickInfo(missingPosition, QuickInfoOptionsDefault, CancellationToken{})
	if len(withDiagnostic.Items) == 0 || withDiagnostic.Items[len(withDiagnostic.Items)-1].Kind != QuickInfoKindError ||
		!strings.Contains(withDiagnostic.Items[len(withDiagnostic.Items)-1].Text(), "does not refer to any known") {
		t.Fatalf("diagnostic quick info = %#v", withDiagnostic)
	}
	withoutDiagnostic := service.GetQuickInfo(missingPosition, QuickInfoOptionsDefault.WithShowDiagnostics(false), CancellationToken{})
	for _, item := range withoutDiagnostic.Items {
		if item.Kind == QuickInfoKindError {
			t.Fatalf("diagnostic-disabled quick info = %#v", withoutDiagnostic)
		}
	}
}

func TestKustoQuickInfoDisplaysLiteralTypeAndIgnoresTrivia(t *testing.T) {
	service := NewKustoCodeService("print Value=123 // comment")
	literal := service.GetQuickInfo(strings.Index(service.Text(), "123")+1, QuickInfoOptionsDefault, CancellationToken{})
	if len(literal.Items) != 1 || literal.Items[0].Kind != QuickInfoKindLiteral || literal.Items[0].Text() != "long" {
		t.Fatalf("literal quick info = %#v", literal)
	}
	if trivia := service.GetQuickInfo(strings.Index(service.Text(), "comment")+1, QuickInfoOptionsDefault, CancellationToken{}); len(trivia.Items) != 0 {
		t.Fatalf("comment quick info = %#v", trivia)
	}
}

func TestKustoQuickInfoPromotesBracketedColumnNameToEnclosingReference(t *testing.T) {
	query := "Events | project ['Account Name']"
	column := NewColumnSymbol("Account Name", ScalarTypes.String.Type())
	service := NewKustoCodeService(query, NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", column))))

	info := service.GetQuickInfo(strings.Index(query, "Account Name")+1, QuickInfoOptionsDefault.WithShowDiagnostics(false), CancellationToken{})
	if len(info.Items) != 1 || info.Items[0].Kind != QuickInfoKindColumn || info.Items[0].Text() != "Account Name: string" {
		t.Fatalf("bracketed-column quick info = %#v, want one column item %q", info, "Account Name: string")
	}
}

func TestKustoQuickInfoBracketedNamePromotionMalformedControls(t *testing.T) {
	column := NewColumnSymbol("Account Name", ScalarTypes.String.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", column)))

	literalQuery := "print Value='Account Name'"
	literalService := NewKustoCodeService(literalQuery, globals)
	literalInfo := literalService.GetQuickInfo(strings.Index(literalQuery, "Account Name")+1, QuickInfoOptionsDefault.WithShowDiagnostics(false), CancellationToken{})
	if len(literalInfo.Items) != 1 || literalInfo.Items[0].Kind != QuickInfoKindLiteral || literalInfo.Items[0].Text() != "string" {
		t.Fatalf("plain-string quick info = %#v, want one literal item %q", literalInfo, "string")
	}

	malformed := []struct {
		name  string
		query string
	}{
		{name: "missing close bracket", query: "Events | project ['Account Name'"},
		{name: "unterminated quoted name", query: "Events | project ['Account Name]"},
	}

	for _, test := range malformed {
		t.Run(test.name, func(t *testing.T) {
			service := NewKustoCodeService(test.query, globals)
			info := service.GetQuickInfo(strings.Index(test.query, "Account Name")+1, QuickInfoOptionsDefault.WithShowDiagnostics(false), CancellationToken{})
			if len(info.Items) != 0 {
				t.Fatalf("malformed-control quick info = %#v, want no promoted item", info)
			}
			if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
				t.Fatal("malformed bracketed name produced no diagnostics")
			}
		})
	}
}

func TestKustoQuickInfoToleratesMalformedQuery(t *testing.T) {
	service := NewKustoCodeService("Events | where (")
	info := service.GetQuickInfo(len(service.Text()), QuickInfoOptionsDefault, CancellationToken{})
	for _, item := range info.Items {
		if item.Text() == "" {
			t.Fatalf("malformed-query quick info contains empty item: %#v", info)
		}
	}
}

func TestKustoQuickInfoVariableItemKindMatchesUpstreamTypeWhitelist(t *testing.T) {
	builder := newKustoQuickInfoBuilder(nil, ParseAndAnalyze("print x=1"), QuickInfoOptionsDefault)
	table := NewTableSymbol("Rows")
	tuple := NewTupleSymbolOf(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	tests := []struct {
		name string
		typ  Symbol
		want QuickInfoKind
	}{
		{name: "scalar", typ: ScalarTypes.Long.Type(), want: QuickInfoKindType},
		{name: "tuple", typ: tuple, want: QuickInfoKindType},
		{name: "table", typ: table, want: QuickInfoKindTable},
		{name: "graph", typ: NewGraphSymbol("Graph", table, table), want: QuickInfoKindGraph},
		{name: "entity-group", typ: NewEntityGroupSymbol("Targets", table), want: QuickInfoKindVariable},
		{name: "group", typ: NewGroupSymbol(table), want: QuickInfoKindVariable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			variable := NewVariableSymbol("value", test.typ, false, nil, nil)
			if got := builder.itemKind(variable); got != test.want {
				t.Fatalf("item kind = %v, want %v", got, test.want)
			}
		})
	}
	if got := builder.itemKind(tuple); got != QuickInfoKindType {
		t.Fatalf("tuple item kind = %v, want %v", got, QuickInfoKindType)
	}
}

func TestKustoQuickInfoLiteralTruncationUsesUTF16Budget(t *testing.T) {
	builder := newKustoQuickInfoBuilder(nil, ParseAndAnalyze("print x=1"), QuickInfoOptionsDefault)
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "exact emoji budget", raw: strings.Repeat("\U0001F600", 20), want: strings.Repeat("\U0001F600", 20)},
		{name: "emoji over budget", raw: strings.Repeat("\U0001F600", 21), want: strings.Repeat("\U0001F600", 20) + " ..."},
		{name: "astral rune at boundary", raw: strings.Repeat("a", 39) + "\U0001F600tail", want: strings.Repeat("a", 39) + " ..."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item, ok := builder.symbolInfo(
				NewVariableSymbol("value", ScalarTypes.String.Type(), true, nil, nil),
				nil,
				ScalarTypes.String.Type(),
				NewValueInfo(test.raw, test.raw, test.raw),
				QuickInfoKindLiteral,
			)
			if !ok {
				t.Fatal("literal quick info was not produced")
			}
			got := item.Text()
			if want := "value: string = " + test.want; got != want || !utf8.ValidString(got) {
				t.Fatalf("literal quick info = %q, want valid UTF-8 %q", got, want)
			}
		})
	}
}
