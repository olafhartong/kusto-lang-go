package kustolang

import "testing"

func TestLocalScopeHasAndContainsSymbols(t *testing.T) {
	outer := NewLocalScope(nil)
	outer.AddSymbol(NewVariableSymbol("Outer", ScalarTypes.String.Type(), false, nil, nil))
	inner := NewLocalScope(outer)
	inner.AddSymbol(NewParameterSymbol("Param", ScalarTypes.Long.Type(), ""))

	if !inner.HasSymbol("Param") {
		t.Fatalf("inner scope should have Param")
	}
	if inner.HasSymbol("Outer") {
		t.Fatalf("HasSymbol should not search outer scopes")
	}
	if !inner.ContainsSymbol("Outer") {
		t.Fatalf("ContainsSymbol should search outer scopes")
	}
	if inner.HasSymbol("param") || inner.ContainsSymbol("outer") {
		t.Fatalf("local scope lookups should be case-sensitive")
	}
	if !inner.HasSymbolMatch("Param", SymbolMatchLocal|SymbolMatchScalar) {
		t.Fatalf("Param should match local scalar")
	}
	if inner.HasSymbolMatch("Param", SymbolMatchTable) {
		t.Fatalf("Param should not match table")
	}
	if !inner.ContainsAnySymbol(SymbolMatchLocal | SymbolMatchScalar) {
		t.Fatalf("inner scope should contain scalar local symbols")
	}
}

func TestLocalScopeGetSymbolsNamedLookupShadowsOuterScope(t *testing.T) {
	outer := NewLocalScope(nil)
	outerSymbol := NewVariableSymbol("Name", ScalarTypes.String.Type(), false, nil, nil)
	outer.AddSymbol(outerSymbol)

	inner := NewLocalScope(outer)
	innerSymbol := NewParameterSymbol("Name", ScalarTypes.Long.Type(), "")
	inner.AddSymbol(innerSymbol)

	var symbols []Symbol
	inner.GetSymbols("Name", SymbolMatchLocal, &symbols)
	if len(symbols) != 1 || symbols[0] != innerSymbol {
		t.Fatalf("named lookup = %#v, want inner symbol only", symbols)
	}

	symbols = nil
	inner.GetSymbols("", SymbolMatchLocal, &symbols)
	if !containsSymbol(symbols, innerSymbol) || !containsSymbol(symbols, outerSymbol) {
		t.Fatalf("unnamed lookup should include inner and outer symbols: %#v", symbols)
	}
}

func TestLocalScopeCopySharesExistingSymbolsButNotFutureWrites(t *testing.T) {
	scope := NewLocalScope(nil)
	scope.AddSymbol(NewVariableSymbol("Before", ScalarTypes.String.Type(), false, nil, nil))

	copied := scope.Copy()
	scope.AddSymbol(NewVariableSymbol("After", ScalarTypes.String.Type(), false, nil, nil))

	if !scope.ContainsSymbol("Before") || !scope.ContainsSymbol("After") {
		t.Fatalf("original scope lost symbols after copy")
	}
	if !copied.ContainsSymbol("Before") {
		t.Fatalf("copied scope should contain pre-copy symbols")
	}
	if copied.ContainsSymbol("After") {
		t.Fatalf("copied scope should not observe post-copy writes")
	}
}

func TestLocalScopeAddSymbolsAndAlternateNames(t *testing.T) {
	scope := NewLocalScope(nil)
	database := NewDatabaseSymbolWithAlternateName("Primary", "Pretty", nil, false)
	scope.AddSymbols([]Symbol{database, nil})

	if !scope.HasSymbol("Primary") || !scope.HasSymbol("Pretty") {
		t.Fatalf("scope should index primary and alternate names")
	}
	if !scope.HasSymbolMatch("Pretty", SymbolMatchDatabase) {
		t.Fatalf("alternate key should preserve symbol-match checks")
	}
	var symbols []Symbol
	scope.GetSymbols("Pretty", SymbolMatchDatabase, &symbols)
	if len(symbols) != 0 {
		t.Fatalf("upstream named GetSymbols filters against primary name, got %#v", symbols)
	}
}

func TestLocalScopeUnnamedEnumerationPreservesDictionaryOrder(t *testing.T) {
	scope := NewLocalScope(nil)
	first := NewVariableSymbol("First", ScalarTypes.String.Type(), false, nil, nil)
	second := NewVariableSymbol("Second", ScalarTypes.Long.Type(), false, nil, nil)
	replacement := NewVariableSymbol("First", ScalarTypes.Bool.Type(), false, nil, nil)
	scope.AddSymbols([]Symbol{first, second, replacement})

	var symbols []Symbol
	scope.GetSymbols("", SymbolMatchLocal, &symbols)
	if len(symbols) != 2 || symbols[0] != replacement || symbols[1] != second {
		t.Fatalf("ordered symbols = %#v, want replacement at First's original position then Second", symbols)
	}

	copied := scope.Copy()
	var copiedSymbols []Symbol
	copied.GetSymbols("", SymbolMatchLocal, &copiedSymbols)
	if len(copiedSymbols) != 2 || copiedSymbols[0] != replacement || copiedSymbols[1] != second {
		t.Fatalf("copied ordered symbols = %#v", copiedSymbols)
	}
}

func TestLocalScopeUnnamedEnumerationIncludesAlternateDictionaryEntry(t *testing.T) {
	scope := NewLocalScope(nil)
	database := NewDatabaseSymbolWithAlternateName("Primary", "Pretty", nil, false)
	scope.AddSymbol(database)

	var symbols []Symbol
	scope.GetSymbols("", SymbolMatchDatabase, &symbols)
	if len(symbols) != 2 || symbols[0] != database || symbols[1] != database {
		t.Fatalf("alternate-name enumeration = %#v, want both dictionary entries", symbols)
	}
}

func containsSymbol(symbols []Symbol, want Symbol) bool {
	for _, symbol := range symbols {
		if symbol == want {
			return true
		}
	}
	return false
}
