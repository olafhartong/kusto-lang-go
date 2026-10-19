package kustolang

import "testing"

func TestBinderTableAggregationCachesOrderedDatabaseTablesAcrossBinders(t *testing.T) {
	leftID := NewColumnSymbol("Id", ScalarTypes.Long.Type())
	rightID := NewColumnSymbol("Id", ScalarTypes.String.Type())
	leftCommon := NewColumnSymbol("Common", ScalarTypes.String.Type())
	rightCommon := NewColumnSymbol("Common", ScalarTypes.String.Type())
	left := NewTableSymbol("Left", leftID, leftCommon)
	right := NewTableSymbol("Right", rightID, rightCommon)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right)).WithCache()

	first := NewBinder(globals)
	byName := first.getTableOfColumnsUnifiedByName([]*TableSymbol{left, right})
	byNameAndType := first.getTableOfColumnsUnifiedByNameAndType([]*TableSymbol{left, right})
	common := first.getTableOfCommonColumns([]*TableSymbol{left, right})
	if column := byName.GetColumn("Id"); column == nil || column.Type != ScalarTypes.Dynamic.Type() {
		t.Fatalf("name-unified Id = %#v", column)
	}
	if byNameAndType.GetColumn("Id_long") == nil || byNameAndType.GetColumn("Id_string") == nil {
		t.Fatalf("name/type-unified columns = %#v", byNameAndType.Columns())
	}
	if columns := common.Columns(); len(columns) != 1 || columns[0].SymbolName() != "Common" || len(columns[0].OriginalColumns) != 2 {
		t.Fatalf("common columns = %#v", columns)
	}

	second := NewBinder(globals)
	if got := second.getTableOfColumnsUnifiedByName([]*TableSymbol{left, right}); got != byName {
		t.Fatal("name-unified database table was not shared through GlobalState cache")
	}
	if got := second.getTableOfColumnsUnifiedByNameAndType([]*TableSymbol{left, right}); got != byNameAndType {
		t.Fatal("name/type-unified database table was not shared through GlobalState cache")
	}
	if got := second.getTableOfCommonColumns([]*TableSymbol{left, right}); got != common {
		t.Fatal("common database table was not shared through GlobalState cache")
	}
	if reversed := second.getTableOfColumnsUnifiedByName([]*TableSymbol{right, left}); reversed == byName {
		t.Fatal("ordered table-list cache treated a reversed list as the same key")
	}
	if got := first.globalBindingCache.UnifiedNameColumnsMap.Count(); got != 2 {
		t.Fatalf("name-unified cache entries = %d, want 2", got)
	}
}

func TestBinderTableAggregationDoesNotGloballyCacheLocalOrInferredTables(t *testing.T) {
	databaseTable := NewTableSymbol("Database", NewColumnSymbol("Declared", ScalarTypes.Long.Type())).WithIsOpen(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", databaseTable)).WithCache()
	binder := NewBinder(globals)

	local := NewTableSymbol("Local", NewColumnSymbol("LocalValue", ScalarTypes.String.Type()))
	firstLocal := binder.getTableOfColumnsUnifiedByName([]*TableSymbol{local})
	secondLocal := NewBinder(globals).getTableOfColumnsUnifiedByName([]*TableSymbol{local})
	if firstLocal == secondLocal {
		t.Fatal("non-database aggregate leaked into the global cache")
	}
	if got := binder.globalBindingCache.UnifiedNameColumnsMap.Count(); got != 0 {
		t.Fatalf("cache contains local tables: %d entries", got)
	}

	inferred, ok := binder.TryGetDeclaredOrInferredColumn(databaseTable, "Future")
	if !ok {
		t.Fatal("failed to infer an open-table column")
	}
	withInference := binder.getTableOfColumnsUnifiedByName([]*TableSymbol{databaseTable})
	if withInference.GetColumn("Future") != inferred {
		t.Fatalf("aggregate omitted binder-inferred column: %#v", withInference.Columns())
	}
	if got := binder.globalBindingCache.UnifiedNameColumnsMap.Count(); got != 0 {
		t.Fatalf("transient inferred schema was globally cached: %d entries", got)
	}

	withoutInference := NewBinder(globals).getTableOfColumnsUnifiedByName([]*TableSymbol{databaseTable})
	if withoutInference.GetColumn("Future") != nil || withoutInference.GetColumn("Declared") == nil {
		t.Fatalf("fresh binder inherited transient schema: %#v", withoutInference.Columns())
	}
}

func TestBinderTableAggregationCommonOpenState(t *testing.T) {
	column := NewColumnSymbol("Id", ScalarTypes.Long.Type())
	openLeft := NewTableSymbol("OpenLeft", column).WithIsOpen(true)
	openRight := NewTableSymbol("OpenRight", NewColumnSymbol("Id", ScalarTypes.Long.Type())).WithIsOpen(true)
	closed := NewTableSymbol("Closed", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", openLeft, openRight, closed)))

	if result := binder.getTableOfCommonColumns([]*TableSymbol{openLeft, openRight}); !result.IsOpen() {
		t.Fatal("common table for all-open inputs is closed")
	}
	if result := binder.getTableOfCommonColumns([]*TableSymbol{openLeft, closed}); result.IsOpen() {
		t.Fatal("common table for mixed inputs is open")
	}
	if result := binder.getTableOfCommonColumns(nil); result.IsOpen() {
		t.Fatal("common table for no inputs is open")
	}
}
