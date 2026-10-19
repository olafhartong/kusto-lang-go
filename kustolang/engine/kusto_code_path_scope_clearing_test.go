package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeFunctionArgumentsClearOuterPathScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `let TableName = "T"; database("db").table(TableName) | project A`
			code := profile.parse(text, kustoCodePathClearingGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid qualified table call diagnostics = %#v", diagnostics)
			}
			position := strings.LastIndex(text, "TableName")
			if symbol := code.GetSpeculativeReferencedSymbol(position, "TableName", SymbolMatchLocal); symbol == nil {
				t.Fatal("function argument did not clear the database path scope")
			}
			if symbols := code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll); !containsScopeSymbolName(symbols, "TableName") {
				t.Fatalf("function argument symbols omit local TableName: %#v", symbols)
			}

			malformedText := `let TableName = "T"; database("db").table(`
			malformed := profile.parse(malformedText, kustoCodePathClearingGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("missing qualified table argument produced no diagnostics")
			}
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "TableName")
		})
	}
}

func TestKustoCodeParenthesizedExpressionsClearOuterPathScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `datatable(Key:string, Bag:dynamic)["x", dynamic({"x":1})] | project Value=Bag[(Key)]`
			code := profile.parse(text, kustoCodePathClearingGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid parenthesized element selector diagnostics = %#v", diagnostics)
			}
			position := strings.LastIndex(text, "Key")
			assertExactRowColumnNames(t, code.GetColumnsInScope(position), "Key", "Bag")
			if symbol := code.GetSpeculativeReferencedSymbol(position, "Key", SymbolMatchColumn); symbol == nil {
				t.Fatal("parenthesized selector did not restore row-column lookup")
			}

			malformedText := `datatable(Key:string, Bag:dynamic)["x", dynamic({"x":1})] | project Value=Bag[(`
			malformed := profile.parse(malformedText, kustoCodePathClearingGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("missing parenthesized selector produced no diagnostics")
			}
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "Key", "Bag")
		})
	}
}

type kustoCodePathClearingProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodePathClearingProfiles() []kustoCodePathClearingProfile {
	return []kustoCodePathClearingProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodePathClearingGlobals() *GlobalState {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table))
}
