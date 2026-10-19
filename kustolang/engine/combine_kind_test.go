package kustolang

import "testing"

func TestCombineKindValuesMatchUpstream(t *testing.T) {
	tests := []struct {
		name string
		got  CombineKind
		want CombineKind
	}{
		{"UnifySameNameAndType", CombineKindUnifySameNameAndType, 0},
		{"UnifySameName", CombineKindUnifySameName, 1},
		{"UniqueNames", CombineKindUniqueNames, 2},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestCombineColumnsUnifySameName(t *testing.T) {
	left := []*ColumnSymbol{
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Int.Type()),
	}
	right := []*ColumnSymbol{
		NewColumnSymbol("Score", ScalarTypes.Real.Type()),
		NewColumnSymbol("TenantId", ScalarTypes.String.Type()),
	}
	combined := CombineColumns(CombineKindUnifySameName, left, right)
	requireCombinedColumn(t, combined, 0, "Account", ScalarTypes.String.Type(), 0)
	requireCombinedColumn(t, combined, 1, "Score", ScalarTypes.Real.Type(), 2)
	requireCombinedColumn(t, combined, 2, "TenantId", ScalarTypes.String.Type(), 0)

	incompatible := CombineColumns(
		CombineKindUnifySameName,
		[]*ColumnSymbol{NewColumnSymbol("Value", ScalarTypes.String.Type())},
		[]*ColumnSymbol{NewColumnSymbol("Value", ScalarTypes.Long.Type())},
	)
	requireCombinedColumn(t, incompatible, 0, "Value", ScalarTypes.Dynamic.Type(), 2)
}

func TestCombineColumnsUnifySameNameAndType(t *testing.T) {
	combined := CombineColumns(
		CombineKindUnifySameNameAndType,
		[]*ColumnSymbol{
			NewColumnSymbol("Value", ScalarTypes.String.Type()),
			NewColumnSymbol("Value", ScalarTypes.Long.Type()),
			NewColumnSymbol("Value", ScalarTypes.String.Type()),
		},
	)
	requireCombinedColumn(t, combined, 0, "Value_string", ScalarTypes.String.Type(), 2)
	requireCombinedColumn(t, combined, 1, "Value_long", ScalarTypes.Long.Type(), 1)
	if len(combined) != 2 {
		t.Fatalf("combined columns = %#v", combined)
	}
}

func TestCombineColumnsUnifySameNameAndTypeUsesColumnMapDynamicBucket(t *testing.T) {
	combined := CombineColumns(
		CombineKindUnifySameNameAndType,
		[]*ColumnSymbol{
			NewColumnSymbol("Value", ScalarTypes.DynamicLong.Type()),
			NewColumnSymbol("Value", ScalarTypes.DynamicReal.Type()),
		},
	)
	requireCombinedColumn(t, combined, 0, "Value", ScalarTypes.Dynamic.Type(), 2)
	if len(combined) != 1 {
		t.Fatalf("combined dynamic columns = %#v", combined)
	}
}

func TestCombineColumnsUnifySameNameAndTypePreservesPinnedDuplicateListEdge(t *testing.T) {
	combined := CombineColumns(
		CombineKindUnifySameNameAndType,
		[]*ColumnSymbol{
			NewColumnSymbol("Value", ScalarTypes.Long.Type()),
			NewColumnSymbol("Value", ScalarTypes.Long.Type()),
			NewColumnSymbol("Value", ScalarTypes.Dynamic.Type()),
		},
	)
	requireCombinedColumn(t, combined, 0, "Value", ScalarTypes.Long.Type(), 2)
	if len(combined) != 1 {
		t.Fatalf("pinned duplicate-list edge columns = %#v", combined)
	}
}

func TestCombineColumnsUniqueNames(t *testing.T) {
	columns := CombineColumns(
		CombineKindUniqueNames,
		[]*ColumnSymbol{
			NewColumnSymbol("c", ScalarTypes.String.Type()),
			NewColumnSymbol("c", ScalarTypes.Long.Type()),
			NewColumnSymbol("c1", ScalarTypes.Real.Type()),
		},
	)
	requireCombinedColumn(t, columns, 0, "c", ScalarTypes.String.Type(), 0)
	requireCombinedColumn(t, columns, 1, "c2", ScalarTypes.Long.Type(), 1)
	requireCombinedColumn(t, columns, 2, "c1", ScalarTypes.Real.Type(), 0)
}

func TestCombineTableSymbolsWithKind(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Id", ScalarTypes.String.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	combined := CombineTableSymbolsWithKind(CombineKindUnifySameNameAndType, left, right)
	requireTableColumn(t, combined, "Id_string", ScalarTypes.String.Type())
	requireTableColumn(t, combined, "Id_long", ScalarTypes.Long.Type())
	if len(combined.Columns()) != 2 {
		t.Fatalf("combined table columns = %#v", combined.Columns())
	}
}

func requireCombinedColumn(t *testing.T, columns []*ColumnSymbol, index int, name string, typ *TypeSymbol, originalCount int) {
	t.Helper()
	if len(columns) <= index {
		t.Fatalf("missing column %d in %#v", index, columns)
	}
	column := columns[index]
	if column.SymbolName() != name || !sameType(column.Type, typ) {
		t.Fatalf("column[%d] = %s:%s, want %s:%s", index, column.SymbolName(), column.Type.SymbolName(), name, typ.SymbolName())
	}
	if len(column.OriginalColumns) != originalCount {
		t.Fatalf("column[%d] original count = %d, want %d", index, len(column.OriginalColumns), originalCount)
	}
}
