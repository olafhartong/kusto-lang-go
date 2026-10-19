package kustolang

import "testing"

type testMemberScope struct {
	root Symbol
}

func (s testMemberScope) GetSymbols(name string, match SymbolMatch, symbols *[]Symbol) {
	if s.root == nil || symbols == nil {
		return
	}
	if name == "" {
		*symbols = append(*symbols, FindMembersByMatch(s.root, match)...)
		return
	}
	*symbols = append(*symbols, FindMembers(s.root, name, match, false)...)
}

func TestEmptyScope(t *testing.T) {
	var symbols []Symbol
	EmptyScopeInstance.GetSymbols("T", SymbolMatchAny, &symbols)
	if len(symbols) != 0 {
		t.Fatalf("empty scope symbols = %#v", symbols)
	}
	if got := GetScopeSymbols(nil, "T", SymbolMatchAny); len(got) != 0 {
		t.Fatalf("nil scope symbols = %#v", got)
	}
}

func TestGetScopeSymbols(t *testing.T) {
	table := NewTableSymbol("T")
	fn := NewScalarFunctionSymbol("F", ScalarTypes.Bool.Type())
	db := NewDatabaseSymbol("db", table, fn)
	scope := testMemberScope{root: db}

	if got := GetScopeSymbols(scope, "T", SymbolMatchTable); len(got) != 1 || got[0] != table {
		t.Fatalf("named table scope lookup = %#v", got)
	}
	if got := GetScopeSymbols(scope, "", SymbolMatchFunction); len(got) != 1 || got[0] != fn {
		t.Fatalf("function scope lookup = %#v", got)
	}
	if got := GetScopeSymbols(scope, "F", SymbolMatchTable); len(got) != 0 {
		t.Fatalf("mismatched scope lookup = %#v", got)
	}
}
