package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCombineColumnsAllModes(t *testing.T) {
	leftValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.String.Type())
	rightValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.Long.Type())
	rightValueAgain := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.String.Type())
	reservedName := kustolang.NewColumnSymbol("Value1", kustolang.ScalarTypes.Real.Type())

	byNameAndType := kustolang.CombineColumns(
		kustolang.CombineKindUnifySameNameAndType,
		[]*kustolang.ColumnSymbol{leftValue},
		[]*kustolang.ColumnSymbol{rightValue, rightValueAgain},
	)
	requirePublicColumn(t, byNameAndType, 0, "Value_string", kustolang.ScalarTypes.String.Type(), leftValue, rightValueAgain)
	requirePublicColumn(t, byNameAndType, 1, "Value_long", kustolang.ScalarTypes.Long.Type(), rightValue)

	byName := kustolang.CombineColumns(
		kustolang.CombineKindUnifySameName,
		[]*kustolang.ColumnSymbol{leftValue},
		[]*kustolang.ColumnSymbol{rightValue},
	)
	requirePublicColumn(t, byName, 0, "Value", kustolang.ScalarTypes.Dynamic.Type(), leftValue, rightValue)

	unique := kustolang.CombineColumns(
		kustolang.CombineKindUniqueNames,
		[]*kustolang.ColumnSymbol{leftValue, rightValue, reservedName},
	)
	requirePublicColumn(t, unique, 0, "Value", kustolang.ScalarTypes.String.Type())
	requirePublicColumn(t, unique, 1, "Value2", kustolang.ScalarTypes.Long.Type(), rightValue)
	requirePublicColumn(t, unique, 2, "Value1", kustolang.ScalarTypes.Real.Type())
}

func TestPublicCombineTablesAndResultNameMetadata(t *testing.T) {
	leftValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.String.Type())
	rightValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.Long.Type())
	combined := kustolang.CombineTableSymbols(
		kustolang.CombineKindUnifySameNameAndType,
		kustolang.NewTableSymbol("Left", leftValue),
		kustolang.NewTableSymbol("Right", rightValue),
	)
	columns := combined.Columns()
	requirePublicColumn(t, columns, 0, "Value_string", kustolang.ScalarTypes.String.Type(), leftValue)
	requirePublicColumn(t, columns, 1, "Value_long", kustolang.ScalarTypes.Long.Type(), rightValue)

	function := kustolang.NewScalarFunctionSymbol("normalize", kustolang.ScalarTypes.String.Type()).
		WithResultNameKind(kustolang.ResultNamePrefixAndFirstArgument).
		WithResultNamePrefix("normalized")
	if function.ResultNameKind() != kustolang.ResultNamePrefixAndFirstArgument || function.ResultNamePrefix() != "normalized" {
		t.Fatalf("result-name metadata = %q %q", function.ResultNameKind(), function.ResultNamePrefix())
	}
}

func TestPublicMalformedUnionRetainsCombinedSchema(t *testing.T) {
	leftValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.String.Type())
	rightValue := kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.Long.Type())
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol(
		"db",
		kustolang.NewTableSymbol("Left", leftValue),
		kustolang.NewTableSymbol("Right", rightValue),
	))

	code := kustolang.ParseAndAnalyze("union Left, Right | where", globals)
	if diagnostics := code.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS006") {
		t.Fatalf("malformed union diagnostics = %#v", diagnostics)
	}
	result, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("malformed union result type = %#v", code.ResultType)
	}
	columns := result.Columns()
	requirePublicColumn(t, columns, 0, "Value_string", kustolang.ScalarTypes.String.Type(), leftValue)
	requirePublicColumn(t, columns, 1, "Value_long", kustolang.ScalarTypes.Long.Type(), rightValue)
}

func requirePublicColumn(t *testing.T, columns []*kustolang.ColumnSymbol, index int, name string, typ *kustolang.TypeSymbol, originals ...*kustolang.ColumnSymbol) {
	t.Helper()
	if index >= len(columns) {
		t.Fatalf("missing column %d in %#v", index, columns)
	}
	column := columns[index]
	if column.SymbolName() != name || column.Type != typ {
		t.Fatalf("column[%d] = %s:%s, want %s:%s", index, column.SymbolName(), column.Type.SymbolName(), name, typ.SymbolName())
	}
	if len(column.OriginalColumns) != len(originals) {
		t.Fatalf("column[%d] originals = %#v, want %#v", index, column.OriginalColumns, originals)
	}
	for i := range originals {
		if column.OriginalColumns[i] != originals[i] {
			t.Fatalf("column[%d] original[%d] = %#v, want %#v", index, i, column.OriginalColumns[i], originals[i])
		}
	}
}
