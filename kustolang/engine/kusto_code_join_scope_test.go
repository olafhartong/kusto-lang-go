package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeJoinAndLookupScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeJoinScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, operator := range []string{"join", "lookup"} {
				t.Run(operator, func(t *testing.T) {
					text := "Left | " + operator + " Right on Common"
					code := profile.parse(text, kustoCodeJoinScopeGlobals())

					rightOperand := strings.Index(text, "Right") + 2
					if got := code.GetColumnsInScope(rightOperand); got != nil {
						t.Fatalf("right operand row scope = %#v, want nil", got)
					}
					if symbols := code.GetSymbolsInScope(rightOperand, SymbolMatchColumn, IncludeFunctionKindAll); len(symbols) != 0 {
						t.Fatalf("right operand column symbols = %#v, want none", symbols)
					}

					condition := strings.Index(text, "Common")
					if got := code.GetColumnsInScope(condition); got == nil || got.SymbolName() != "Left" {
						t.Fatalf("condition row scope = %#v, want Left", got)
					}
					symbols := code.GetSymbolsInScope(condition, SymbolMatchColumn, IncludeFunctionKindAll)
					assertJoinScopeSymbols(t, symbols)
					if symbol := code.GetSpeculativeReferencedSymbol(condition, "$right", SymbolMatchAny); symbol == nil || symbol.SymbolKind() != SymbolKindVariable {
						t.Fatalf("speculative $right = %#v, want variable", symbol)
					}
				})
			}
		})
	}
}

func TestKustoCompleterJoinAndLookupConditionScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeJoinScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, operator := range []string{"join", "lookup"} {
				text := "Left | " + operator + " Right on Co"
				code := profile.parse(text, kustoCodeJoinScopeGlobals())
				items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, items, "$left", "$right", "Common")
				assertCompletionExcludes(t, items, "LeftOnly", "RightOnly")

				leftPath := "Left | " + operator + " Right on $left."
				leftCode := profile.parse(leftPath, kustoCodeJoinScopeGlobals())
				if len(leftCode.GetDiagnostics()) == 0 {
					t.Fatalf("malformed %s $left path produced no diagnostics", operator)
				}
				leftItems := NewKustoCodeServiceForCode(leftCode).GetCompletionItems(len(leftPath), CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, leftItems, "Common", "LeftOnly")
				assertCompletionExcludes(t, leftItems, "RightOnly", "$left", "$right")

				rightPath := "Left | " + operator + " Right on $right."
				rightCode := profile.parse(rightPath, kustoCodeJoinScopeGlobals())
				if len(rightCode.GetDiagnostics()) == 0 {
					t.Fatalf("malformed %s $right path produced no diagnostics", operator)
				}
				rightItems := NewKustoCodeServiceForCode(rightCode).GetCompletionItems(len(rightPath), CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, rightItems, "Common", "RightOnly")
				assertCompletionExcludes(t, rightItems, "LeftOnly", "$left", "$right")
			}
		})
	}
}

func TestKustoCodeMalformedJoinAndLookupScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeJoinScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, operator := range []string{"join", "lookup"} {
				missingKey := "Left | " + operator + " Right on "
				code := profile.parse(missingKey, kustoCodeJoinScopeGlobals())
				if len(code.GetDiagnostics()) == 0 {
					t.Fatalf("malformed %s produced no diagnostics", operator)
				}
				if got := code.GetColumnsInScope(len(missingKey)); got == nil || got.SymbolName() != "Left" {
					t.Fatalf("malformed %s row scope = %#v, want Left", operator, got)
				}
				items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(missingKey), CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, items, "$left", "$right", "Common")
				assertCompletionExcludes(t, items, "LeftOnly", "RightOnly")
			}

			missingLookupClause := `Left | lookup Right `
			lookup := profile.parse(missingLookupClause, kustoCodeJoinScopeGlobals())
			if len(lookup.GetDiagnostics()) == 0 {
				t.Fatal("lookup without on produced no diagnostics")
			}
			if got := lookup.GetColumnsInScope(len(missingLookupClause)); got != nil {
				t.Fatalf("missing lookup clause row scope = %#v, want nil", got)
			}
			items := NewKustoCodeServiceForCode(lookup).GetCompletionItems(len(missingLookupClause), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionExcludes(t, items, "$left", "$right", "Common", "LeftOnly", "RightOnly")
		})
	}
}

type kustoCodeJoinScopeProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeJoinScopeProfiles() []kustoCodeJoinScopeProfile {
	return []kustoCodeJoinScopeProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeJoinScopeGlobals() *GlobalState {
	left := NewTableSymbol("Left",
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("LeftOnly", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right",
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("RightOnly", ScalarTypes.Long.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", left, right))
}

func assertJoinScopeSymbols(t *testing.T, symbols []Symbol) {
	t.Helper()
	byName := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		byName[symbol.SymbolName()] = symbol
	}
	for _, name := range []string{"$left", "$right", "Common"} {
		if byName[name] == nil {
			t.Fatalf("join scope missing %q: %#v", name, symbols)
		}
	}
	for _, name := range []string{"LeftOnly", "RightOnly"} {
		if byName[name] != nil {
			t.Fatalf("join scope unexpectedly includes %q: %#v", name, symbols)
		}
	}
	if byName["$left"].SymbolKind() != SymbolKindVariable || byName["$right"].SymbolKind() != SymbolKindVariable {
		t.Fatalf("join tuple symbols = %#v / %#v, want variables", byName["$left"], byName["$right"])
	}
}
