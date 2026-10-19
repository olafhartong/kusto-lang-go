package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeAggregateScopeContextsProfiles(t *testing.T) {
	for _, profile := range kustoCodeScopeKindProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			tests := []struct {
				name       string
				text       string
				aggregate  string
				argument   string
				normalTail string
			}{
				{
					name:       "summarize",
					text:       `T | summarize Total=sum(Value) by User`,
					aggregate:  "sum(Value)",
					argument:   "Value",
					normalTail: "User",
				},
				{
					name:       "make-series",
					text:       `T | make-series Total=sum(Value) on Timestamp step 1h`,
					aggregate:  "sum(Value)",
					argument:   "Value",
					normalTail: "Timestamp",
				},
				{
					name:       "top-nested",
					text:       `T | top-nested 5 of User by sum(Value)`,
					aggregate:  "sum(Value)",
					argument:   "Value",
					normalTail: "",
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					code := profile.parse(tt.text, kustoCodeScopeKindGlobals())
					if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
						t.Fatalf("valid %s diagnostics = %#v", tt.name, diagnostics)
					}
					aggregatePosition := strings.Index(tt.text, tt.aggregate)
					assertAggregateOnlyScope(t, code.GetSymbolsInScope(aggregatePosition, SymbolMatchAny, IncludeFunctionKindAll))

					argumentPosition := strings.Index(tt.text[aggregatePosition:], tt.argument) + aggregatePosition
					assertNormalRowScope(t, code.GetSymbolsInScope(argumentPosition, SymbolMatchAny, IncludeFunctionKindAll))
					if tt.normalTail != "" {
						tailPosition := strings.LastIndex(tt.text, tt.normalTail)
						assertNormalRowScope(t, code.GetSymbolsInScope(tailPosition, SymbolMatchAny, IncludeFunctionKindAll))
					}
				})
			}
		})
	}
}

func TestKustoCodeNamedParameterRestoresNormalScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeScopeKindProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | summarize hint.strategy= `
			code := profile.parse(text, kustoCodeScopeKindGlobals())
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("missing summarize parameter value produced no diagnostics")
			}
			assertNormalRowScope(t, code.GetSymbolsInScope(len(text), SymbolMatchAny, IncludeFunctionKindAll))
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "User", "Value", "Timestamp", "strlen(string)")
			assertCompletionExcludes(t, items, "sum(expr)", "bag_unpack(column, [OutputSchema])")
		})
	}
}

func TestKustoCodeEvaluatePlugInScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeScopeKindProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | evaluate `
			code := profile.parse(text, kustoCodeScopeKindGlobals())
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("missing evaluate plug-in produced no diagnostics")
			}
			symbols := code.GetSymbolsInScope(len(text), SymbolMatchAny, IncludeFunctionKindAll)
			for _, name := range []string{"bag_unpack", "narrow"} {
				if !containsScopeSymbolName(symbols, name) {
					t.Fatalf("plug-in scope missing %q", name)
				}
			}
			for _, name := range []string{"User", "T", "strlen", "sum"} {
				if containsScopeSymbolName(symbols, name) {
					t.Fatalf("plug-in scope unexpectedly contains %q", name)
				}
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "bag_unpack(column, [column_prefix])", "narrow()")
			assertCompletionExcludes(t, items, "User", "Value", "T", "strlen(string)", "sum(expr)")

			argument := `T | evaluate bag_unpack(`
			argumentCode := profile.parse(argument, kustoCodeScopeKindGlobals())
			if len(argumentCode.GetDiagnostics()) == 0 {
				t.Fatal("missing plug-in argument produced no diagnostics")
			}
			assertNormalRowScope(t, argumentCode.GetSymbolsInScope(len(argument), SymbolMatchAny, IncludeFunctionKindAll))
			argumentItems := NewKustoCodeServiceForCode(argumentCode).GetCompletionItems(len(argument), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, argumentItems, "User", "Value", "Timestamp")
			assertCompletionExcludes(t, argumentItems, "bag_unpack(column, [column_prefix])", "sum(expr)")
		})
	}
}

func TestKustoCompleterMalformedAggregateContextsProfiles(t *testing.T) {
	for _, profile := range kustoCodeScopeKindProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, test := range []struct {
				text     string
				position int
			}{
				{`T | summarize ,`, len(`T | summarize `)},
				{`T | make-series ,`, len(`T | make-series `)},
				{`T | top-nested 5 of User by ,`, len(`T | top-nested 5 of User by `)},
			} {
				code := profile.parse(test.text, kustoCodeScopeKindGlobals())
				if len(code.GetDiagnostics()) == 0 {
					t.Fatalf("malformed aggregate context %q produced no diagnostics", test.text)
				}
				assertAggregateOnlyScope(t, code.GetSymbolsInScope(test.position, SymbolMatchAny, IncludeFunctionKindAll))
				items := NewKustoCodeServiceForCode(code).GetCompletionItems(test.position, CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, items, "sum(expr)", "count()", "strlen(string)")
				assertCompletionExcludes(t, items, "User", "Value", "Timestamp", "T", "bag_unpack(column, [column_prefix])")
			}
		})
	}
}

func TestKustoCodeSetOptionScopeProfiles(t *testing.T) {
	option := NewOptionSymbolWithType("custom_option", "", ScalarTypes.String, nil)
	for _, profile := range kustoCodeScopeKindProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals := kustoCodeScopeKindGlobals().WithOptions([]*OptionSymbol{option})
			text := `set `
			code := profile.parse(text, globals)
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("missing option name produced no diagnostics")
			}
			symbols := code.GetSymbolsInScope(len(text), SymbolMatchOption, IncludeFunctionKindAll)
			if !containsScopeSymbolName(symbols, "custom_option") {
				t.Fatalf("option scope = %#v", symbols)
			}
			for _, name := range []string{"T", "User", "strlen", "sum", "bag_unpack"} {
				if containsScopeSymbolName(symbols, name) {
					t.Fatalf("option scope unexpectedly contains %q", name)
				}
			}
		})
	}
}

type kustoCodeScopeKindProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeScopeKindProfiles() []kustoCodeScopeKindProfile {
	return []kustoCodeScopeKindProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeScopeKindGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("User", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()),
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table))
}

func assertAggregateOnlyScope(t *testing.T, symbols []Symbol) {
	t.Helper()
	for _, name := range []string{"sum", "count", "strlen"} {
		if !containsScopeSymbolName(symbols, name) {
			t.Fatalf("aggregate scope missing %q", name)
		}
	}
	for _, name := range []string{"User", "Value", "Timestamp", "T", "bag_unpack"} {
		if containsScopeSymbolName(symbols, name) {
			t.Fatalf("aggregate scope unexpectedly contains %q", name)
		}
	}
}

func assertNormalRowScope(t *testing.T, symbols []Symbol) {
	t.Helper()
	for _, name := range []string{"User", "Value", "Timestamp", "T", "strlen"} {
		if !containsScopeSymbolName(symbols, name) {
			t.Fatalf("normal row scope missing %q", name)
		}
	}
	for _, name := range []string{"sum", "bag_unpack"} {
		if containsScopeSymbolName(symbols, name) {
			t.Fatalf("normal row scope unexpectedly contains %q", name)
		}
	}
}
