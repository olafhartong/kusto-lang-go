package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicKustoCodeDefaultScopeEntryPointsMatchExplicitDefaults(t *testing.T) {
	table := kustolang.NewTableSymbol("Events", kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type()))
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))
	const text = `let Threshold=1; Events | where Account == Threshold`
	code := kustolang.ParseAndAnalyze(text, globals)
	position := len(text)
	token := kustolang.CancellationToken{}

	got := code.GetSymbolsInScopeDefault(position, token)
	want := code.GetSymbolsInScope(position, kustolang.SymbolMatchAny, kustolang.IncludeFunctionKindAll, token)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default scope symbols differ:\ngot:  %#v\nwant: %#v", got, want)
	}
	if got := code.GetSpeculativeReferencedSymbolDefault(position, "Account", token); got !=
		code.GetSpeculativeReferencedSymbol(position, "Account", kustolang.SymbolMatchAny, token) {
		t.Fatalf("default speculative symbol = %#v", got)
	}
}
