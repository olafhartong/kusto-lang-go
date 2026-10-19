package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeMvApplyIntermediateScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeMvApplyScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | mv-apply with_itemindex=Idx Expanded=Items to typeof(string) on (project Expanded, Idx, Keep)`
			code := profile.parse(text, kustoCodeMvApplyScopeGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid mv-apply diagnostics = %#v", diagnostics)
			}

			beforeOn := strings.Index(text, "Expanded=Items") + len("Expanded=")
			before := code.GetColumnsInScope(beforeOn)
			assertMvApplyScopeColumns(t, before, map[string]*TypeSymbol{
				"Items": ScalarTypes.Dynamic.Type(),
				"Keep":  ScalarTypes.String.Type(),
			})
			if before.GetColumn("Expanded") != nil || before.GetColumn("Idx") != nil {
				t.Fatalf("pre-on scope contains expansion columns: %#v", before.Columns())
			}

			inside := strings.Index(text, "project Expanded") + len("project ")
			intermediate := code.GetColumnsInScope(inside)
			assertMvApplyScopeColumns(t, intermediate, map[string]*TypeSymbol{
				"Items":    ScalarTypes.Dynamic.Type(),
				"Keep":     ScalarTypes.String.Type(),
				"Expanded": ScalarTypes.String.Type(),
				"Idx":      ScalarTypes.Long.Type(),
			})
		})
	}
}

func TestKustoCompleterMvApplyIntermediateScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeMvApplyScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | mv-apply with_itemindex=Idx Expanded=Items to typeof(string) on (project Ex`
			code := profile.parse(text, kustoCodeMvApplyScopeGlobals())
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "Expanded", "Idx", "Keep", "Items")
		})
	}
}

func TestKustoCodeMalformedMvApplyIntermediateScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeMvApplyScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `T | mv-apply Item=Items to typeof(long) on ()`
			code := profile.parse(text, kustoCodeMvApplyScopeGlobals())
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("empty mv-apply subquery produced no diagnostics")
			}
			position := strings.LastIndex(text, ")")
			intermediate := code.GetColumnsInScope(position)
			assertMvApplyScopeColumns(t, intermediate, map[string]*TypeSymbol{
				"Items": ScalarTypes.Dynamic.Type(),
				"Keep":  ScalarTypes.String.Type(),
				"Item":  ScalarTypes.Long.Type(),
			})
			symbols := code.GetSymbolsInScope(position, SymbolMatchColumn, IncludeFunctionKindAll)
			for _, name := range []string{"Items", "Keep", "Item"} {
				if !containsScopeSymbolName(symbols, name) {
					t.Fatalf("malformed mv-apply scope missing %q: %#v", name, symbols)
				}
			}
		})
	}
}

type kustoCodeMvApplyScopeProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeMvApplyScopeProfiles() []kustoCodeMvApplyScopeProfile {
	return []kustoCodeMvApplyScopeProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeMvApplyScopeGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("Items", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("Keep", ScalarTypes.String.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table))
}

func assertMvApplyScopeColumns(t *testing.T, table *TableSymbol, expected map[string]*TypeSymbol) {
	t.Helper()
	if table == nil {
		t.Fatal("mv-apply scope is nil")
	}
	for name, typ := range expected {
		column := table.GetColumn(name)
		if column == nil || column.Type != typ {
			t.Fatalf("mv-apply scope column %s = %#v, want type %v; all=%#v", name, column, typ, table.Columns())
		}
	}
}

func containsScopeSymbolName(symbols []Symbol, name string) bool {
	for _, symbol := range symbols {
		if symbol != nil && symbol.SymbolName() == name {
			return true
		}
	}
	return false
}
