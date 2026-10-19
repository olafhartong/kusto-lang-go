package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeCommandNearestTableRowScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeCommandTableProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `  .alter table T column-docstrings (A:"alpha", B:"beta")`
			code := profile.parse(text, kustoCodeCommandTableGlobals())
			if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid command syntax diagnostics = %#v", diagnostics)
			}
			beforeTableEnd := strings.Index(text, "T column")
			if row := code.GetColumnsInScope(beforeTableEnd); row != nil {
				t.Fatalf("command row before table end = %#v, want nil", row)
			}
			assertExactRowColumnNames(t, code.GetColumnsInScope(strings.Index(text, `A:"alpha"`)), "A", "B")
			assertExactRowColumnNames(t, code.GetColumnsInScope(strings.Index(text, `B:"beta"`)), "A", "B")
		})
	}
}

func TestKustoCompleterMalformedCommandUsesNearestTableColumnsProfiles(t *testing.T) {
	for _, profile := range kustoCodeCommandTableProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `.alter table T column-docstrings (`
			code := profile.parse(text, kustoCodeCommandTableGlobals())
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("unclosed command column list produced no diagnostics")
			}
			assertExactRowColumnNames(t, code.GetColumnsInScope(len(text)), "A", "B")
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "A", "B")
			assertCompletionExcludes(t, items, "OtherColumn")
		})
	}
}

type kustoCodeCommandTableProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeCommandTableProfiles() []kustoCodeCommandTableProfile {
	return []kustoCodeCommandTableProfile{{
		name: "current",
		parse: func(text string, globals *GlobalState) *KustoCode {
			return ParseWithGlobals(text, globals).Analyze()
		},
	}}
}

func kustoCodeCommandTableGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("B", ScalarTypes.Long.Type()))
	other := NewTableSymbol("Other", NewColumnSymbol("OtherColumn", ScalarTypes.Bool.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table, other))
}
