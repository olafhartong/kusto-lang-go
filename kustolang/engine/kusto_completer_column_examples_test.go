package kustolang

import "testing"

func TestKustoCompleterOffersColumnExamplesForEqualityAndIn(t *testing.T) {
	state := NewColumnSymbol("State", ScalarTypes.String.Type()).WithExamples([]string{"'TEXAS'", "California", "'TEXAS'"})
	other := NewColumnSymbol("Other", ScalarTypes.String.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", state, other)))
	tests := []struct {
		text string
	}{
		{text: "Events | where State == "},
		{text: "Events | where State != "},
		{text: "Events | where State =~ "},
		{text: "Events | where State in ("},
		{text: "Events | where State in ('TEXAS', "},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			service := NewKustoCodeService(test.text, globals)
			if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
				t.Fatalf("incomplete column-example query %q produced no diagnostics", test.text)
			}
			items := service.GetCompletionItems(len(test.text), CompletionOptionsDefault, CancellationToken{}).Items
			texas := requireCompletionKind(t, items, "TEXAS", CompletionKindExample)
			if texas.BeforeText() != "'TEXAS'" || texas.MatchText != "TEXAS" {
				t.Fatalf("quoted column example = before:%q match:%q", texas.BeforeText(), texas.MatchText)
			}
			california := requireCompletionKind(t, items, "California", CompletionKindExample)
			if california.BeforeText() != "'California'" || california.MatchText != "California" {
				t.Fatalf("unquoted column example = before:%q match:%q", california.BeforeText(), california.MatchText)
			}
			if count := completionDisplayCount(items, "TEXAS"); count != 1 {
				t.Fatalf("duplicate column example count = %d, want 1", count)
			}
			requireCompletionKind(t, items, "Other", CompletionKindColumn)
		})
	}
}

func TestKustoCompleterDoesNotLeakColumnExamples(t *testing.T) {
	state := NewColumnSymbol("State", ScalarTypes.String.Type()).WithExamples([]string{"'TEXAS'"})
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", state)))
	tests := []string{
		"Events | where State > ",
		"Events | where State in ('TEXAS') ",
		"Events | where State == ; print Result=",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireNoCompletion(t, items, "TEXAS")
		})
	}
}

func TestCreateColumnExampleCompletionMatchesUpstreamLiteralShapes(t *testing.T) {
	tests := []struct {
		name    string
		example string
		typ     *TypeSymbol
		display string
		before  string
		match   string
	}{
		{name: "quoted string", example: "'Texas'", typ: ScalarTypes.String.Type(), display: "Texas", before: "'Texas'", match: "Texas"},
		{name: "escaped string", example: "can't", typ: ScalarTypes.String.Type(), display: "can't", before: `"can't"`, match: "can't"},
		{name: "bool goo", example: "bool(true)", typ: ScalarTypes.Bool.Type(), display: "true", before: "true", match: "true"},
		{name: "bool invalid", example: "maybe", typ: ScalarTypes.Bool.Type(), display: "maybe", before: "bool(maybe)", match: "maybe"},
		{name: "int bare", example: "5", typ: ScalarTypes.Int.Type(), display: "5", before: "int(5)", match: "5"},
		{name: "int goo", example: "int(5)", typ: ScalarTypes.Int.Type(), display: "5", before: "int(5)", match: "5"},
		{name: "long bare", example: "5", typ: ScalarTypes.Long.Type(), display: "5", before: "5", match: "5"},
		{name: "long goo", example: "long(5)", typ: ScalarTypes.Long.Type(), display: "5", before: "5", match: "5"},
		{name: "real bare", example: "1.5", typ: ScalarTypes.Real.Type(), display: "1.5", before: "1.5", match: "1.5"},
		{name: "real integer", example: "5", typ: ScalarTypes.Real.Type(), display: "5", before: "real(5)", match: "5"},
		{name: "real goo", example: "double(1.5)", typ: ScalarTypes.Real.Type(), display: "1.5", before: "1.5", match: "1.5"},
		{name: "decimal bare", example: "1.5", typ: ScalarTypes.Decimal.Type(), display: "1.5", before: "decimal(1.5)", match: "1.5"},
		{name: "timespan bare", example: "1h", typ: ScalarTypes.TimeSpan.Type(), display: "1h", before: "1h", match: "1h"},
		{name: "timespan goo", example: "timespan(1h)", typ: ScalarTypes.TimeSpan.Type(), display: "1h", before: "1h", match: "1h"},
		{name: "datetime bare", example: "2024-01-01", typ: ScalarTypes.DateTime.Type(), display: "2024-01-01", before: "datetime(2024-01-01)", match: "2024-01-01"},
		{name: "datetime goo", example: "datetime(2024-01-01)", typ: ScalarTypes.DateTime.Type(), display: "2024-01-01", before: "datetime(2024-01-01)", match: "2024-01-01"},
		{name: "guid bare", example: "00000000-0000-0000-0000-000000000000", typ: ScalarTypes.Guid.Type(), display: "00000000-0000-0000-0000-000000000000", before: "guid(00000000-0000-0000-0000-000000000000)", match: "00000000-0000-0000-0000-000000000000"},
		{name: "dynamic bare", example: "[1,2]", typ: ScalarTypes.Dynamic.Type(), display: "[1,2]", before: "dynamic([1,2])", match: "[1,2]"},
		{name: "dynamic goo", example: "dynamic([1,2])", typ: ScalarTypes.Dynamic.Type(), display: "[1,2]", before: "dynamic([1,2])", match: "[1,2]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := createColumnExampleCompletion(test.example, test.typ)
			if item.Kind != CompletionKindExample || item.DisplayText != test.display || item.BeforeText() != test.before || item.MatchText != test.match {
				t.Fatalf("completion = kind:%v display:%q before:%q match:%q", item.Kind, item.DisplayText, item.BeforeText(), item.MatchText)
			}
		})
	}
}

func completionDisplayCount(items []CompletionItem, displayText string) int {
	count := 0
	for _, item := range items {
		if item.DisplayText == displayText {
			count++
		}
	}
	return count
}
