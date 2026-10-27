package kustolang

// Ported from src/Kusto.Language/Symbols/SymbolMatch.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type SymbolMatch uint32

const (
	SymbolMatchNone SymbolMatch = 0

	SymbolMatchColumn SymbolMatch = 1

	SymbolMatchTable SymbolMatch = SymbolMatchColumn << 1

	SymbolMatchExternalTable SymbolMatch = SymbolMatchTable << 1

	SymbolMatchFunction SymbolMatch = SymbolMatchExternalTable << 1

	SymbolMatchView SymbolMatch = SymbolMatchFunction << 1

	SymbolMatchLocal SymbolMatch = SymbolMatchView << 1

	SymbolMatchDatabase SymbolMatch = SymbolMatchLocal << 1

	SymbolMatchCluster SymbolMatch = SymbolMatchDatabase << 1

	SymbolMatchEntityGroup SymbolMatch = SymbolMatchCluster << 1

	SymbolMatchEntityGroupElement SymbolMatch = SymbolMatchEntityGroup << 1

	SymbolMatchScalar SymbolMatch = SymbolMatchEntityGroupElement << 1

	SymbolMatchTabular SymbolMatch = SymbolMatchScalar << 1

	SymbolMatchNonScalar SymbolMatch = SymbolMatchTabular << 1

	SymbolMatchMaterializedView SymbolMatch = SymbolMatchNonScalar << 1

	SymbolMatchOption SymbolMatch = SymbolMatchMaterializedView << 1

	SymbolMatchGraph SymbolMatch = SymbolMatchOption << 1

	SymbolMatchStoredQueryResult SymbolMatch = SymbolMatchGraph << 1

	SymbolMatchGraphModel SymbolMatch = SymbolMatchStoredQueryResult << 1

	SymbolMatchGraphSnapshot SymbolMatch = SymbolMatchGraphModel << 1

	SymbolMatchAny = SymbolMatchColumn |
		SymbolMatchTable |
		SymbolMatchFunction |
		SymbolMatchView |
		SymbolMatchLocal |
		SymbolMatchDatabase |
		SymbolMatchCluster |
		SymbolMatchMaterializedView |
		SymbolMatchEntityGroup |
		SymbolMatchEntityGroupElement |
		SymbolMatchGraph |
		SymbolMatchGraphModel |
		SymbolMatchGraphSnapshot

	SymbolMatchDefault = SymbolMatchColumn |
		SymbolMatchTable |
		SymbolMatchFunction |
		SymbolMatchView |
		SymbolMatchLocal |
		SymbolMatchMaterializedView |
		SymbolMatchEntityGroup |
		SymbolMatchEntityGroupElement |
		SymbolMatchGraph |
		SymbolMatchExternalTable
)

func (m SymbolMatch) Has(flag SymbolMatch) bool {
	return m&flag != 0
}

func SymbolMatches(symbol Symbol, match SymbolMatch) bool {
	return symbolMatchesName(symbol, nil, match, false)
}

func SymbolMatchesName(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) bool {
	return symbolMatchesName(symbol, &name, match, ignoreCase)
}

func FindMembers(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) []Symbol {
	if symbol == nil {
		return nil
	}
	if name != "" {
		if table, ok := tableSymbolFromSymbol(symbol); ok && len(table.Columns()) > 0 {
			if !ignoreCase {
				if column, found := table.TryGetColumn(name); found && SymbolMatchesName(column, name, match, false) {
					return []Symbol{column}
				}
				return nil
			}
			for _, column := range table.Columns() {
				if SymbolMatchesName(column, name, match, true) {
					return []Symbol{column}
				}
			}
			return nil
		}
	}
	var out []Symbol
	for _, member := range symbol.Members() {
		if SymbolMatchesName(member, name, match, ignoreCase) {
			out = append(out, member)
		}
	}
	return out
}

func FindMembersByMatch(symbol Symbol, match SymbolMatch) []Symbol {
	if symbol == nil {
		return nil
	}
	var out []Symbol
	for _, member := range symbol.Members() {
		if SymbolMatches(member, match) {
			out = append(out, member)
		}
	}
	return out
}

func FindFirstMemberMatch(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) Symbol {
	for _, member := range FindMembers(symbol, name, match, ignoreCase) {
		return member
	}
	return nil
}

func symbolMatchesName(symbol Symbol, name *string, match SymbolMatch, ignoreCase bool) bool {
	if symbol == nil {
		return false
	}

	if name != nil && !matchesSymbolPrimaryName(symbol, *name, ignoreCase) {
		return false
	}

	if match.Has(SymbolMatchColumn) {
		if _, ok := symbol.(*ColumnSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchTable) {
		if table, ok := tableSymbolFromSymbol(symbol); ok && !table.IsExternal() && !table.IsMaterializedView() && !table.IsStoredQueryResult() {
			return true
		}
	}

	if match.Has(SymbolMatchExternalTable) {
		if table, ok := tableSymbolFromSymbol(symbol); ok && table.IsExternal() {
			return true
		}
	}

	if match.Has(SymbolMatchMaterializedView) {
		if table, ok := tableSymbolFromSymbol(symbol); ok && table.IsMaterializedView() {
			return true
		}
	}

	if match.Has(SymbolMatchDatabase) {
		if _, ok := symbol.(*DatabaseSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchCluster) {
		if _, ok := symbol.(*ClusterSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchEntityGroup) && isSymbolTypeOrVariable(symbol, func(candidate Symbol) bool {
		_, ok := candidate.(*EntityGroupSymbol)
		return ok
	}) {
		return true
	}

	if match.Has(SymbolMatchEntityGroupElement) && isSymbolTypeOrVariable(symbol, func(candidate Symbol) bool {
		_, ok := candidate.(*EntityGroupElementSymbol)
		return ok
	}) {
		return true
	}

	if match.Has(SymbolMatchOption) {
		if _, ok := symbol.(*OptionSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchGraph) && isSymbolTypeOrVariable(symbol, func(candidate Symbol) bool {
		_, ok := candidate.(*GraphSymbol)
		return ok
	}) {
		return true
	}

	if match.Has(SymbolMatchGraphModel) {
		if _, ok := symbol.(*GraphModelSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchGraphSnapshot) {
		if _, ok := symbol.(*GraphSnapshotSymbol); ok {
			return true
		}
	}

	if match.Has(SymbolMatchStoredQueryResult) && isSymbolTypeOrVariable(symbol, func(candidate Symbol) bool {
		switch s := candidate.(type) {
		case *StoredQueryResultSymbol:
			return true
		case *TableSymbol:
			return s.IsStoredQueryResult()
		default:
			return false
		}
	}) {
		return true
	}

	if match.Has(SymbolMatchScalar) && !match.Has(SymbolMatchNonScalar) && !match.Has(SymbolMatchTabular) && !IsScalar(symbol) {
		return false
	}

	if match.Has(SymbolMatchNonScalar) && !match.Has(SymbolMatchScalar) && symbol.SymbolTabularity() == TabularityScalar {
		return false
	}

	if match.Has(SymbolMatchTabular) && !match.Has(SymbolMatchScalar) && !match.Has(SymbolMatchNonScalar) && !IsTabular(symbol) {
		return false
	}

	if match.Has(SymbolMatchFunction) {
		switch symbol.(type) {
		case *FunctionSymbol, *PatternSymbol:
			return true
		}
	}

	if match.Has(SymbolMatchView) {
		if fn, ok := symbol.(*FunctionSymbol); ok && fn.IsView() {
			return true
		}
	}

	if match.Has(SymbolMatchLocal) {
		switch symbol.(type) {
		case *VariableSymbol, *ParameterSymbol:
			return true
		}
	}

	return false
}

func isSymbolTypeOrVariable(symbol Symbol, predicate func(Symbol) bool) bool {
	if symbol == nil || predicate == nil {
		return false
	}
	if predicate(symbol) {
		return true
	}
	if variable, ok := symbol.(*VariableSymbol); ok {
		return predicate(variable.Type)
	}
	return false
}

func matchesSymbolPrimaryName(symbol Symbol, name string, ignoreCase bool) bool {
	symbolName := symbol.SymbolName()
	if ignoreCase {
		return ordinalEqualFold(symbolName, name)
	}
	return name != "" && symbolName != "" && name[0] == symbolName[0] && symbolName == name
}
