package kustolang

import "testing"

func TestConversionOrderingMatchesUpstream(t *testing.T) {
	tests := []struct {
		name string
		got  Conversion
		want Conversion
	}{
		{"None", ConversionNone, 0},
		{"Promotable", ConversionPromotable, 1},
		{"Dynamic", ConversionDynamic, 2},
		{"Compatible", ConversionCompatible, 3},
		{"Any", ConversionAny, 4},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
	if !IsConversionAllowed(ConversionPromotable, ConversionCompatible) {
		t.Fatal("promotable should be allowed by compatible conversion")
	}
	if IsConversionAllowed(ConversionCompatible, ConversionPromotable) {
		t.Fatal("compatible should not be allowed by promotable conversion")
	}
}

func TestSymbolAssignableScalarConversions(t *testing.T) {
	if !SymbolAssignableTo(ScalarTypes.Int.Type(), ScalarTypes.Int.Type()) {
		t.Fatal("same scalar type should be assignable")
	}
	if SymbolAssignableTo(ScalarTypes.Int.Type(), ScalarTypes.Long.Type()) {
		t.Fatal("int should not assign to long without conversion")
	}
	if !SymbolAssignableTo(ScalarTypes.Int.Type(), ScalarTypes.Long.Type(), ConversionPromotable) {
		t.Fatal("int should promote to long")
	}
	if SymbolAssignableTo(ScalarTypes.Long.Type(), ScalarTypes.Int.Type(), ConversionPromotable) {
		t.Fatal("long should not promote to int")
	}
	if !SymbolAssignableTo(ScalarTypes.Long.Type(), ScalarTypes.Int.Type(), ConversionCompatible) {
		t.Fatal("long should be compatible with int")
	}
	if !SymbolAssignableTo(ScalarTypes.Unknown.Type(), ScalarTypes.String.Type()) {
		t.Fatal("unknown scalar should assign to scalar target")
	}
	if !SymbolAssignableTo(ScalarTypes.String.Type(), ScalarTypes.Unknown.Type()) {
		t.Fatal("scalar source should assign to unknown scalar target")
	}
	if !SymbolAssignableToAny(ScalarTypes.Int.Type(), []Symbol{ScalarTypes.String.Type(), ScalarTypes.Real.Type()}, ConversionPromotable) {
		t.Fatal("int should assign to any target containing real with promotable conversion")
	}
}

func TestSymbolAssignableDynamicConversions(t *testing.T) {
	if SymbolAssignableTo(ScalarTypes.String.Type(), ScalarTypes.Dynamic.Type()) {
		t.Fatal("scalar should not assign to dynamic without dynamic conversion")
	}
	if !SymbolAssignableTo(ScalarTypes.String.Type(), ScalarTypes.Dynamic.Type(), ConversionDynamic) {
		t.Fatal("scalar should assign to dynamic with dynamic conversion")
	}
	if !SymbolAssignableTo(ScalarTypes.DynamicArray, ScalarTypes.Dynamic.Type()) {
		t.Fatal("dynamic array should assign to dynamic")
	}
	if SymbolAssignableTo(ScalarTypes.DynamicArrayOfLong, ScalarTypes.DynamicArray.Type()) {
		t.Fatal("typed dynamic array should not assign to untyped dynamic array without conversion")
	}
	if !SymbolAssignableTo(ScalarTypes.DynamicArrayOfLong, ScalarTypes.DynamicArray.Type(), ConversionDynamic) {
		t.Fatal("typed dynamic array should assign to untyped dynamic array with conversion")
	}
	if !SymbolAssignableTo(ScalarTypes.DynamicLong, ScalarTypes.DynamicReal, ConversionPromotable) {
		t.Fatal("dynamic long should promote to dynamic real through underlying types")
	}
	if SymbolAssignableTo(ScalarTypes.DynamicReal, ScalarTypes.DynamicLong, ConversionPromotable) {
		t.Fatal("dynamic real should not promote to dynamic long")
	}

	customInt := NewDynamicPrimitiveSymbol(ScalarTypes.Int)
	customLong := NewDynamicPrimitiveSymbol(ScalarTypes.Long)
	if !SymbolAssignableTo(customInt.Type(), customLong.Type(), ConversionPromotable) {
		t.Fatal("wrapped custom dynamic int should promote to wrapped custom dynamic long")
	}
	if !SymbolAssignableTo(customInt, customLong.Type(), ConversionPromotable) || !SymbolAssignableTo(customInt.Type(), customLong, ConversionPromotable) {
		t.Fatal("mixed direct/wrapped custom dynamic primitives should preserve promotion")
	}

	longArrayA := NewDynamicArraySymbol(ScalarTypes.Long.Type())
	longArrayB := NewDynamicArraySymbol(ScalarTypes.Long.Type())
	realArray := NewDynamicArraySymbol(ScalarTypes.Real.Type())
	if !SymbolAssignableTo(longArrayA.Type(), longArrayB.Type()) {
		t.Fatal("distinct wrapped custom arrays with equal elements should be assignable")
	}
	if SymbolAssignableTo(longArrayA.Type(), realArray.Type(), ConversionPromotable) {
		t.Fatal("pinned upstream array direction rejects long source to real target")
	}
	if !SymbolAssignableTo(realArray.Type(), longArrayA.Type(), ConversionPromotable) {
		t.Fatal("pinned upstream array direction accepts real source to long target")
	}
	if !SymbolAssignableTo(realArray, longArrayA.Type(), ConversionPromotable) || !SymbolAssignableTo(realArray.Type(), longArrayA, ConversionPromotable) {
		t.Fatal("mixed direct/wrapped custom arrays should follow upstream element direction")
	}
	if SymbolAssignableTo(longArrayA.Type(), ScalarTypes.DynamicArray.Type()) {
		t.Fatal("wrapped typed array should not assign to generic array without conversion")
	}
	if !SymbolAssignableTo(longArrayA.Type(), ScalarTypes.DynamicArray.Type(), ConversionDynamic) {
		t.Fatal("wrapped typed array should assign to generic array with dynamic conversion")
	}
}

func TestSymbolAssignableTupleColumnTableAndBagShapes(t *testing.T) {
	tuple := NewTupleSymbolOf(NewColumnSymbol("value", ScalarTypes.Int.Type()))
	if SymbolAssignableTo(tuple, ScalarTypes.Long.Type(), ConversionPromotable) {
		t.Fatal("single-column tuple reduction should use upstream Conversion.None")
	}
	if !SymbolAssignableTo(tuple, ScalarTypes.Int.Type(), ConversionPromotable) {
		t.Fatal("single-column tuple should reduce to an identical scalar type")
	}

	targetColumn := NewColumnSymbol("Account", ScalarTypes.String.Type())
	sourceColumn := NewColumnSymbol("Account", ScalarTypes.String.Type())
	if !SymbolAssignableTo(sourceColumn, targetColumn) {
		t.Fatal("same-name same-type column should be assignable")
	}
	if SymbolAssignableTo(NewColumnSymbol("Other", ScalarTypes.String.Type()), targetColumn) {
		t.Fatal("different-name column should not be assignable")
	}

	sourceTable := NewTableSymbol("Source",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	)
	targetTable := NewTableSymbol("Target", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	if !SymbolAssignableTo(sourceTable, targetTable) {
		t.Fatal("source table with target columns should be assignable")
	}
	if SymbolAssignableTo(targetTable, sourceTable) {
		t.Fatal("source table missing target columns should not be assignable")
	}
	if SymbolAssignableTo(
		NewTableSymbol("Source", NewColumnSymbol("Account", ScalarTypes.Long.Type())),
		targetTable,
	) {
		t.Fatal("source table with incompatible column type should not be assignable")
	}

	sourceBag := NewDynamicBagSymbol(
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	)
	targetBag := NewDynamicBagSymbol(
		NewColumnSymbol("Score", ScalarTypes.Real.Type()),
	)
	if !SymbolAssignableTo(sourceBag, targetBag) {
		t.Fatal("bag source should satisfy target properties using any conversion")
	}
	if SymbolAssignableTo(sourceBag, NewDynamicBagSymbol(NewColumnSymbol("Missing", ScalarTypes.String.Type()))) {
		t.Fatal("bag source missing target property should not be assignable")
	}
	if !SymbolAssignableTo(sourceBag.Type(), targetBag.Type()) {
		t.Fatal("wrapped custom bag source should satisfy wrapped target properties")
	}
	if !SymbolAssignableTo(sourceBag, targetBag.Type()) || !SymbolAssignableTo(sourceBag.Type(), targetBag) {
		t.Fatal("mixed direct/wrapped custom bags should preserve structural assignment")
	}
}
