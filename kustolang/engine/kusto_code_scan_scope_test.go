package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeScanDeclarationsAndAliasesScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeScanScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | scan declare (State:string, Total:long) with (step s1: State == ""; step s2: s1.State == State;) | project User`
			code := profile.parse(text, kustoCodeScanScopeGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid scan diagnostics = %#v", diagnostics)
			}

			withStart := strings.Index(text, "with")
			if symbols := code.GetSymbolsInScope(withStart, SymbolMatchAny, IncludeFunctionKindAll); containsScopeSymbolName(symbols, "State") || containsScopeSymbolName(symbols, "s1") {
				t.Fatalf("scan declarations crossed strict with boundary: %#v", symbols)
			}
			inside := strings.Index(text, `State == ""`)
			symbols := code.GetSymbolsInScope(inside, SymbolMatchAny, IncludeFunctionKindAll)
			assertScanScopeSymbols(t, symbols)

			first := code.GetSpeculativeReferencedSymbol(inside, "s1", SymbolMatchAny)
			variable, ok := first.(*VariableSymbol)
			if !ok {
				t.Fatalf("speculative s1 = %#v, want variable", first)
			}
			tuple, ok := variable.Type.(*TupleSymbol)
			if !ok || FindFirstMemberMatch(tuple, "User", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "State", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Total", SymbolMatchColumn, false) == nil {
				t.Fatalf("scan step tuple = %#v", variable.Type)
			}

			after := strings.LastIndex(text, "project User") + len("project ")
			if symbols := code.GetSymbolsInScope(after, SymbolMatchAny, IncludeFunctionKindAll); containsScopeSymbolName(symbols, "s1") || containsScopeSymbolName(symbols, "s2") {
				t.Fatalf("scan step aliases leaked after operator: %#v", symbols)
			}
		})
	}
}

func TestKustoCompleterScanDeclarationsAndAliasesProfiles(t *testing.T) {
	for _, profile := range kustoCodeScanScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | scan declare (State:string, Total:long) with (step s1: Sta`
			code := profile.parse(text, kustoCodeScanScopeGlobals())
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "State", "Total", "s1", "User", "Value")

			path := `T | scan declare (State:string, Total:long) with (step s1: true; step s2: s1.`
			pathCode := profile.parse(path, kustoCodeScanScopeGlobals())
			pathItems := NewKustoCodeServiceForCode(pathCode).GetCompletionItems(len(path), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, pathItems, "State", "Total", "User", "Value")
			assertCompletionExcludes(t, pathItems, "s1", "s2")
		})
	}
}

func TestKustoCodeMalformedScanDeclarationsRemainInScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeScanScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | scan declare (State:string) with (step s1: ; step s2: true;)`
			code := profile.parse(text, kustoCodeScanScopeGlobals())
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("missing scan predicate produced no diagnostics")
			}
			position := strings.Index(text, ";")
			symbols := code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll)
			for _, name := range []string{"State", "s1", "s2"} {
				if !containsScopeSymbolName(symbols, name) {
					t.Fatalf("malformed scan scope missing %q: %#v", name, symbols)
				}
			}
		})
	}
}

type kustoCodeScanScopeProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeScanScopeProfiles() []kustoCodeScanScopeProfile {
	return []kustoCodeScanScopeProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeScanScopeGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("User", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table))
}

func assertScanScopeSymbols(t *testing.T, symbols []Symbol) {
	t.Helper()
	byName := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		byName[symbol.SymbolName()] = symbol
	}
	for _, name := range []string{"User", "Value", "State", "Total", "s1", "s2"} {
		if byName[name] == nil {
			t.Fatalf("scan scope missing %q: %#v", name, symbols)
		}
	}
	if _, ok := byName["State"].(*ColumnSymbol); !ok {
		t.Fatalf("State = %T, want ColumnSymbol", byName["State"])
	}
	if _, ok := byName["s1"].(*VariableSymbol); !ok {
		t.Fatalf("s1 = %T, want VariableSymbol", byName["s1"])
	}
}
