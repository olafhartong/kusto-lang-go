package kustolang

// Ported from the table aggregation and caching region of
// src/Kusto.Language/Binder/Binder_Misc.cs at the pinned upstream revision.

func (b *Binder) canGlobalCacheTables(tables []*TableSymbol) bool {
	if b == nil || b.Globals == nil {
		return false
	}
	for _, table := range tables {
		if table == nil || !b.Globals.IsDatabaseTable(table) || b.hasInferredColumns(table) {
			return false
		}
	}
	return true
}

func (b *Binder) hasInferredColumns(table *TableSymbol) bool {
	return b != nil && table != nil && len(b.openTableInferredOrder[table]) > 0
}

func (b *Binder) aggregateTableColumnSets(tables []*TableSymbol) ([][]*ColumnSymbol, bool) {
	sets := make([][]*ColumnSymbol, 0, len(tables))
	open := false
	for _, table := range tables {
		if table == nil {
			continue
		}
		sets = append(sets, b.declaredAndInferredColumns(table))
		open = open || table.IsOpen()
	}
	return sets, open
}

func (b *Binder) getTableOfColumnsUnifiedByName(tables []*TableSymbol) *TableSymbol {
	cache := b.globalBindingCache
	if cache != nil && !hasBinderInferredColumns(b, tables) {
		if table, ok := cache.UnifiedNameColumnsMap.TryGetValue(tables); ok {
			return table
		}
	}
	sets, open := b.aggregateTableColumnSets(tables)
	result := NewTableSymbol("", CombineColumns(CombineKindUnifySameName, sets...)...).WithIsOpen(open)
	if cache != nil && b.canGlobalCacheTables(tables) {
		cache.UnifiedNameColumnsMap.AddOrUpdate(tables, result)
	}
	return result
}

func (b *Binder) getTableOfColumnsUnifiedByNameAndType(tables []*TableSymbol) *TableSymbol {
	cache := b.globalBindingCache
	if cache != nil && !hasBinderInferredColumns(b, tables) {
		if table, ok := cache.UnifiedNameAndTypeColumnsMap.TryGetValue(tables); ok {
			return table
		}
	}
	sets, open := b.aggregateTableColumnSets(tables)
	result := NewTableSymbol("", CombineColumns(CombineKindUnifySameNameAndType, sets...)...).WithIsOpen(open)
	if cache != nil && b.canGlobalCacheTables(tables) {
		cache.UnifiedNameAndTypeColumnsMap.AddOrUpdate(tables, result)
	}
	return result
}

func (b *Binder) getTableOfCommonColumns(tables []*TableSymbol) *TableSymbol {
	cache := b.globalBindingCache
	if cache != nil && !hasBinderInferredColumns(b, tables) {
		if table, ok := cache.CommonColumnsMap.TryGetValue(tables); ok {
			return table
		}
	}
	result := NewTableSymbol("", findCommonColumnsWithBinder(b, tables)...)
	if len(tables) > 0 && allTablesOpen(tables) {
		result = result.WithIsOpen(true)
	}
	if cache != nil && b.canGlobalCacheTables(tables) {
		cache.CommonColumnsMap.AddOrUpdate(tables, result)
	}
	return result
}

func hasBinderInferredColumns(b *Binder, tables []*TableSymbol) bool {
	for _, table := range tables {
		if b.hasInferredColumns(table) {
			return true
		}
	}
	return false
}

func allTablesOpen(tables []*TableSymbol) bool {
	for _, table := range tables {
		if table == nil || !table.IsOpen() {
			return false
		}
	}
	return true
}

func findCommonColumnsWithBinder(b *Binder, tables []*TableSymbol) []*ColumnSymbol {
	if !hasBinderInferredColumns(b, tables) {
		return findCommonColumns(tables)
	}
	if len(tables) == 0 {
		return nil
	}
	sets := make([]*TableSymbol, 0, len(tables))
	for _, table := range tables {
		sets = append(sets, NewTableSymbol("", b.declaredAndInferredColumns(table)...).WithIsOpen(table.IsOpen()))
	}
	return findCommonColumns(sets)
}
