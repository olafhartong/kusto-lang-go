package kustolang

// Ported from src/Kusto.Language/Binder/LocalScope.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// LocalScope models nested scoping for let variables and function parameters.
type LocalScope struct {
	outerScope    *LocalScope
	symbols       map[string]Symbol
	symbolOrder   []string
	sharedSymbols *LocalScope
}

func NewLocalScope(outerScope *LocalScope) *LocalScope {
	return newLocalScope(nil, nil, outerScope, nil)
}

func newLocalScope(symbols map[string]Symbol, symbolOrder []string, outerScope *LocalScope, sharedScope *LocalScope) *LocalScope {
	return &LocalScope{
		symbols:       symbols,
		symbolOrder:   symbolOrder,
		outerScope:    minimalOuterScope(outerScope),
		sharedSymbols: sharedScope,
	}
}

func minimalOuterScope(outerScope *LocalScope) *LocalScope {
	for outerScope != nil && len(outerScope.symbols) == 0 && outerScope.sharedSymbols == nil {
		outerScope = outerScope.outerScope
	}
	return outerScope
}

func (s *LocalScope) HasSymbol(name string) bool {
	if s == nil {
		return false
	}
	if s.symbols != nil {
		if _, ok := s.symbols[name]; ok {
			return true
		}
	}
	return s.sharedSymbols != nil && s.sharedSymbols.HasSymbol(name)
}

func (s *LocalScope) HasSymbolMatch(name string, match SymbolMatch) bool {
	if s == nil {
		return false
	}
	if s.symbols != nil {
		if symbol, ok := s.symbols[name]; ok && SymbolMatches(symbol, match) {
			return true
		}
	}
	return s.sharedSymbols != nil && s.sharedSymbols.HasSymbolMatch(name, match)
}

func (s *LocalScope) HasAnySymbol(match SymbolMatch) bool {
	if s == nil {
		return false
	}
	if s.symbols != nil {
		for _, name := range s.symbolOrder {
			symbol := s.symbols[name]
			if SymbolMatches(symbol, match) {
				return true
			}
		}
	}
	return s.sharedSymbols != nil && s.sharedSymbols.HasAnySymbol(match)
}

func (s *LocalScope) ContainsSymbol(name string) bool {
	for scope := s; scope != nil; scope = scope.outerScope {
		if scope.HasSymbol(name) {
			return true
		}
	}
	return false
}

func (s *LocalScope) ContainsSymbolMatch(name string, match SymbolMatch) bool {
	for scope := s; scope != nil; scope = scope.outerScope {
		if scope.HasSymbolMatch(name, match) {
			return true
		}
	}
	return false
}

func (s *LocalScope) ContainsAnySymbol(match SymbolMatch) bool {
	for scope := s; scope != nil; scope = scope.outerScope {
		if scope.HasAnySymbol(match) {
			return true
		}
	}
	return false
}

func (s *LocalScope) Copy() *LocalScope {
	if s == nil {
		return NewLocalScope(nil)
	}
	if len(s.symbols) > 0 {
		s.sharedSymbols = newLocalScope(s.symbols, s.symbolOrder, nil, s.sharedSymbols)
		s.symbols = nil
		s.symbolOrder = nil
	}
	return newLocalScope(nil, nil, s.outerScope, s.sharedSymbols)
}

func (s *LocalScope) AddSymbol(symbol Symbol) bool {
	if s == nil || symbol == nil {
		return false
	}
	if s.symbols == nil {
		s.symbols = map[string]Symbol{}
	}
	s.addSymbolName(symbol.SymbolName(), symbol)
	if alternate := symbol.AlternateName(); alternate != "" {
		s.addSymbolName(alternate, symbol)
	}
	return true
}

func (s *LocalScope) addSymbolName(name string, symbol Symbol) {
	if _, exists := s.symbols[name]; !exists {
		s.symbolOrder = append(s.symbolOrder, name)
	}
	s.symbols[name] = symbol
}

func (s *LocalScope) AddSymbols(symbols []Symbol) {
	for _, symbol := range symbols {
		s.AddSymbol(symbol)
	}
}

func (s *LocalScope) GetSymbols(name string, match SymbolMatch, symbols *[]Symbol) {
	if s == nil || symbols == nil {
		return
	}
	originalCount := len(*symbols)
	if s.symbols != nil {
		if name != "" {
			if symbol, ok := s.symbols[name]; ok && SymbolMatchesName(symbol, name, match, false) {
				*symbols = append(*symbols, symbol)
			}
		} else {
			for _, symbolName := range s.symbolOrder {
				symbol := s.symbols[symbolName]
				if SymbolMatches(symbol, match) {
					*symbols = append(*symbols, symbol)
				}
			}
		}
	}
	if s.sharedSymbols != nil && (name == "" || len(*symbols) == originalCount) {
		s.sharedSymbols.GetSymbols(name, match, symbols)
	}
	if s.outerScope != nil && (name == "" || len(*symbols) == originalCount) {
		s.outerScope.GetSymbols(name, match, symbols)
	}
}
