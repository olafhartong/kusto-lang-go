package kustolang

// Ported from GetImpliedTables and GetViewResultTablesInScope in
// src/Kusto.Language/Binder/Binder_TablesAndColumns.cs.

func (b *Binder) impliedTables() []*TableSymbol {
	var tables []*TableSymbol
	if b != nil && b.Globals != nil && b.Globals.Database != nil {
		tables = append(tables, b.Globals.Database.Tables()...)
	}
	return append(tables, b.viewResultTablesInScope()...)
}

func (b *Binder) viewResultTablesInScope() []*TableSymbol {
	if b == nil || b.LocalScope == nil {
		return nil
	}
	var symbols []Symbol
	b.LocalScope.GetSymbols("", SymbolMatchView, &symbols)
	var tables []*TableSymbol
	for _, symbol := range symbols {
		function, ok := symbol.(*FunctionSymbol)
		if !ok || !function.IsView() || function.MinArgumentCount() != 0 {
			continue
		}
		for _, signature := range function.Signatures() {
			if signature == nil || signature.MinArgumentCount != 0 {
				continue
			}
			var result Symbol
			if signature.Declaration != nil {
				result = functionBodyResultSymbol(signature.Declaration)
			} else if signature.DeclaredReturnSymbol != nil {
				result = signature.DeclaredReturnSymbol
			} else if signature.ReturnKind == ReturnTypeComputed {
				result, _ = newSimpleNodeBinder(b).bindComputedFunction(signature, nil, nil, nil)
			}
			if table, ok := tableSymbolFromSymbol(result); ok {
				tables = append(tables, table)
			}
			break
		}
	}
	return tables
}
