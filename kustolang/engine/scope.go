package kustolang

// Ported from src/Kusto.Language/Symbols/Scope.cs and EmptyScope.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type Scope interface {
	GetSymbols(name string, match SymbolMatch, symbols *[]Symbol)
}

func GetScopeSymbols(scope Scope, name string, match SymbolMatch) []Symbol {
	var symbols []Symbol
	if scope != nil {
		scope.GetSymbols(name, match, &symbols)
	}
	return symbols
}

type EmptyScope struct{}

func (EmptyScope) GetSymbols(name string, match SymbolMatch, symbols *[]Symbol) {
}

var EmptyScopeInstance Scope = EmptyScope{}
