package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodePipeRowScopeStartsAtOperatorTriviaProfiles(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	const text = `Events | project Account`
	bar := strings.Index(text, "|")
	operatorTrivia := bar + 1
	operatorText := strings.Index(text, "project")

	for _, profile := range kustoCodePipeScopeProfiles(globals) {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid pipe diagnostics = %#v", diagnostics)
			}
			if row := code.GetColumnsInScope(bar); row != nil {
				t.Fatalf("row scope entered at pipe token: %#v", row)
			}
			for _, position := range []int{operatorTrivia, operatorText} {
				if row := code.GetColumnsInScope(position); row != table {
					t.Fatalf("row scope at position %d = %#v, want input table", position, row)
				}
			}
		})
	}
}

func TestKustoCodeMissingPipeOperatorRetainsDiagnosticsAndRowScopeProfiles(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	const text = `Events |`

	for _, profile := range kustoCodePipeScopeProfiles(globals) {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing pipe operator should retain syntax diagnostics")
			}
			if row := code.GetColumnsInScope(len(text)); row != table {
				t.Fatalf("recovered pipe row scope = %#v, want input table", row)
			}
			assertScopeNames(t,
				code.GetSymbolsInScope(len(text), SymbolMatchColumn, IncludeFunctionKindNone),
				[]string{"Account"},
			)
		})
	}
}

type kustoCodePipeScopeProfile struct {
	name  string
	parse func(string) *KustoCode
}

func kustoCodePipeScopeProfiles(globals *GlobalState) []kustoCodePipeScopeProfile {
	return []kustoCodePipeScopeProfile{{
		name:  "current",
		parse: func(text string) *KustoCode { return ParseAndAnalyze(text, globals) },
	}}
}
