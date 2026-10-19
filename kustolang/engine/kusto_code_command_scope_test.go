package kustolang

func symbolInScope(code *KustoCode, position int, name string) Symbol {
	if code == nil {
		return nil
	}
	for _, symbol := range code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll) {
		if symbol != nil && symbol.SymbolName() == name {
			return symbol
		}
	}
	return nil
}
