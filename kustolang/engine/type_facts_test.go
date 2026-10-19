package kustolang

import "testing"

func TestGetElementAndWidestScalarType(t *testing.T) {
	if got := GetElementType(ScalarTypes.DynamicArrayOfString.Type()); !sameType(got, ScalarTypes.DynamicString.Type()) {
		t.Fatalf("array element type = %s, want dynamic string", typeName(got))
	}
	if got := GetElementType(ScalarTypes.Dynamic.Type()); !sameType(got, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("dynamic element type = %s, want dynamic", typeName(got))
	}
	if got := GetElementType(ScalarTypes.String.Type()); got != nil {
		t.Fatalf("string element type = %s, want nil", typeName(got))
	}
	typedBag := ScalarTypes.GetDynamicBag(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	typedArray := NewDynamicArraySymbol(typedBag.Type())
	if got := GetElementType(typedArray.Type()); got != typedBag.Type() {
		t.Fatalf("custom array element type = %#v, want typed bag %#v", got, typedBag.Type())
	}
	nestedArray := NewDynamicArraySymbol(typedArray.Type())
	if got := GetElementType(nestedArray.Type()); got != typedArray.Type() {
		t.Fatalf("nested custom array element type = %#v, want typed array %#v", got, typedArray.Type())
	}
	if !IsDynamicArray(typedArray.Type()) || !IsDynamicBag(typedBag.Type()) || !IsDynamicArrayOrBag(nestedArray.Type()) {
		t.Fatal("custom dynamic TypeSymbol back-references should satisfy dynamic type facts")
	}
	if got := GetWidestScalarType([]*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Real.Type(), ScalarTypes.Long.Type()}); got != ScalarTypes.Real {
		t.Fatalf("widest scalar = %#v, want real", got)
	}
}

func TestPromoteToLongMirrorsUpstreamNullPassthrough(t *testing.T) {
	if got := PromoteToLong(ScalarTypes.Int.Type()); got != ScalarTypes.Long.Type() {
		t.Fatalf("promote int = %s, want long", typeName(got))
	}
	if got := PromoteToLong(ScalarTypes.String.Type()); got != ScalarTypes.String.Type() {
		t.Fatalf("promote string = %s, want original string", typeName(got))
	}
	if got := PromoteToLong(nil); got != nil {
		t.Fatalf("promote nil = %s, want nil", typeName(got))
	}
}

func TestTryGetCommonTypeScalarConversions(t *testing.T) {
	common, ok := TryGetCommonType(ScalarTypes.Int.Type(), ScalarTypes.Real.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.Real.Type()) {
		t.Fatalf("int/real common = %s, %t; want real,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.String.Type(), ScalarTypes.Long.Type(), ConversionPromotable)
	if ok || common != nil {
		t.Fatalf("string/long promotable common = %s, %t; want nil,false", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.String.Type(), ScalarTypes.Long.Type(), ConversionDynamic)
	if !ok || !sameType(common, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("string/long dynamic common = %s, %t; want dynamic,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.Unknown.Type(), ScalarTypes.String.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.Unknown.Type()) {
		t.Fatalf("unknown/string common = %s, %t; want unknown,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.Null.Type(), ScalarTypes.String.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.String.Type()) {
		t.Fatalf("null/string common = %s, %t; want string,true", typeName(common), ok)
	}
}

func TestCommonTypeDynamicSymbols(t *testing.T) {
	common, ok := TryGetCommonType(ScalarTypes.DynamicLong.Type(), ScalarTypes.DynamicReal.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.DynamicReal.Type()) {
		t.Fatalf("dynamic long/real common = %s, %t; want dynamic real,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.DynamicLong.Type(), ScalarTypes.Real.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.DynamicReal.Type()) {
		t.Fatalf("dynamic long/real scalar common = %s, %t; want dynamic real,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.DynamicArrayOfString.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.DynamicArrayOfString.Type()) {
		t.Fatalf("same dynamic array common = %s, %t; want dynamic array of string,true", typeName(common), ok)
	}
	common, ok = TryGetCommonType(ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.DynamicArrayOfLong.Type(), ConversionPromotable)
	if !ok || !sameType(common, ScalarTypes.DynamicArray.Type()) {
		t.Fatalf("different dynamic array common = %s, %t; want dynamic array,true", typeName(common), ok)
	}
	customStringArrayA := NewDynamicArraySymbol(ScalarTypes.String.Type())
	customStringArrayB := NewDynamicArraySymbol(ScalarTypes.String.Type())
	common, ok = TryGetCommonType(customStringArrayA.Type(), customStringArrayB.Type(), ConversionPromotable)
	if !ok || common != ScalarTypes.String.Type() {
		t.Fatalf("distinct same-element custom array common = %s, %t; want upstream string element,true", typeName(common), ok)
	}
}

func TestTryGetCommonSymbolTypeMirrorsTableBranch(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	count := NewColumnSymbol("Count", ScalarTypes.Int.Type())
	payload := NewColumnSymbol("Payload", ScalarTypes.String.Type())
	left := NewTableSymbol("Left",
		account,
		count,
		NewColumnSymbol("OnlyLeft", ScalarTypes.Long.Type()),
		payload,
	)
	right := NewTableSymbol("Right",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Count", ScalarTypes.Long.Type()),
		NewColumnSymbol("OnlyRight", ScalarTypes.String.Type()),
		NewColumnSymbol("Payload", ScalarTypes.DateTime.Type()),
	)

	same, ok := TryGetCommonSymbolType(left, left, ConversionPromotable)
	if !ok || same != left {
		t.Fatalf("same table common = %#v,%t; want original table,true", same, ok)
	}

	commonSymbol, ok := TryGetCommonSymbolType(left, right, ConversionPromotable)
	if !ok {
		t.Fatal("table/table common returned false")
	}
	commonTable, ok := commonSymbol.(*TableSymbol)
	if !ok {
		t.Fatalf("table/table common = %T, want *TableSymbol", commonSymbol)
	}
	if commonTable.SymbolName() != "" {
		t.Fatalf("common table name = %q, want anonymous", commonTable.SymbolName())
	}
	columns := commonTable.Columns()
	if len(columns) != 3 {
		t.Fatalf("common table column count = %d, want 3", len(columns))
	}
	if columns[0] != account || !sameType(columns[0].Type, ScalarTypes.String.Type()) {
		t.Fatalf("common Account column = %#v, want original string column", columns[0])
	}
	if columns[1] == count || columns[1].SymbolName() != "Count" || !sameType(columns[1].Type, ScalarTypes.Long.Type()) {
		t.Fatalf("common Count column = %#v, want widened long clone", columns[1])
	}
	if columns[2] == payload || columns[2].SymbolName() != "Payload" || !sameType(columns[2].Type, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("common Payload column = %#v, want dynamic clone", columns[2])
	}

	scalar, ok := TryGetCommonSymbolType(ScalarTypes.Int.Type(), ScalarTypes.Long.Type(), ConversionPromotable)
	if !ok || !sameType(returnTypeFromSymbol(scalar), ScalarTypes.Long.Type()) {
		t.Fatalf("scalar common = %s,%t; want long,true", typeName(returnTypeFromSymbol(scalar)), ok)
	}
	if got := returnTypeFromSymbol(NewColumnSymbol("Flag", ScalarTypes.Bool.Type())); !sameType(got, ScalarTypes.Bool.Type()) {
		t.Fatalf("column return type = %s; want bool", typeName(got))
	}
	mixed, ok := TryGetCommonSymbolType(left, ScalarTypes.String.Type(), ConversionDynamic)
	if ok || mixed != nil {
		t.Fatalf("table/scalar common = %#v,%t; want nil,false", mixed, ok)
	}
}

func TestGetCommonTypeHelpers(t *testing.T) {
	common := GetCommonScalarType([]*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Long.Type()})
	if !sameType(common, ScalarTypes.Long.Type()) {
		t.Fatalf("common scalar = %s, want long", typeName(common))
	}
	common = GetCommonScalarType([]*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()})
	if common != nil {
		t.Fatalf("common scalar without dynamic conversion = %s, want nil", typeName(common))
	}
	common = GetCommonScalarType([]*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()}, ConversionDynamic)
	if !sameType(common, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("common scalar with dynamic conversion = %s, want dynamic", typeName(common))
	}

	columns := []*ColumnSymbol{
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("A", ScalarTypes.Long.Type()),
	}
	if common = GetCommonColumnType(columns); common != nil {
		t.Fatalf("common column type = %s, want nil", typeName(common))
	}
	common = GetCommonColumnType(columns, ConversionDynamic)
	if !sameType(common, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("common column type with dynamic conversion = %s, want dynamic", typeName(common))
	}
}

func TestTypeFactsPredicatesMirrorUpstreamDynamicSemantics(t *testing.T) {
	if !IsInteger(ScalarTypes.Int.Type()) || !IsNumeric(ScalarTypes.Real.Type()) || !IsSummable(ScalarTypes.DateTime.Type()) || !IsOrderable(ScalarTypes.String.Type()) {
		t.Fatal("primitive scalar flag predicates should use upstream scalar flags")
	}
	if IsInteger(ScalarTypes.DynamicLong) || IsNumeric(ScalarTypes.Dynamic.Type()) || IsSummable(ScalarTypes.DynamicArrayOfLong) {
		t.Fatal("dynamic symbols do not inherit primitive scalar flags in upstream TypeFacts")
	}
	if !IsIntegerOrDynamic(ScalarTypes.DynamicLong) || !IsIntegerOrArray(ScalarTypes.DynamicArrayOfLong) {
		t.Fatal("dynamic integer helper predicates should include dynamic long/array variants")
	}
	if !IsAnyScalarExceptDynamic(ScalarTypes.String.Type()) || IsAnyScalarExceptDynamic(ScalarTypes.DynamicString) {
		t.Fatal("IsAnyScalarExceptDynamic should accept primitive scalars and reject dynamic symbols")
	}
	if IsAnyScalarExceptBool(ScalarTypes.Bool.Type()) || IsAnyScalarExceptBool(ScalarTypes.DynamicBool) || !IsAnyScalarExceptBool(ScalarTypes.DynamicLong) {
		t.Fatal("IsAnyScalarExceptBool should reject bool/dynamic bool only")
	}
	if IsAnyScalarExceptRealOrBool(ScalarTypes.Real.Type()) || IsAnyScalarExceptRealOrBool(ScalarTypes.DynamicReal) || !IsAnyScalarExceptRealOrBool(ScalarTypes.DynamicString) {
		t.Fatal("IsAnyScalarExceptRealOrBool should reject real/bool dynamic counterparts only")
	}
	if !IsStringOrDynamic(ScalarTypes.DynamicBag) || !IsStringOrArray(ScalarTypes.DynamicArrayOfString) || IsStringOrArray(ScalarTypes.Long.Type()) {
		t.Fatal("string/dynamic/array helper predicates diverged")
	}
	if !IsDynamicArray(ScalarTypes.DynamicArrayOfString) || !IsDynamicBag(ScalarTypes.DynamicBag) || !IsDynamicArrayOrBag(ScalarTypes.GeoShape) || IsDynamicArray(ScalarTypes.String.Type()) {
		t.Fatal("dynamic array/bag helper predicates diverged")
	}
	if !IsPromotableTo(ScalarTypes.Int.Type(), ScalarTypes.Long.Type()) || !IsPromotableTo(ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.Dynamic.Type()) || IsPromotableTo(ScalarTypes.String.Type(), ScalarTypes.Long.Type()) {
		t.Fatal("IsPromotableTo should mirror scalar widening and dynamic array/bag promotion")
	}
}

func TestGetCommonArgumentTypeFiltersCommonParameters(t *testing.T) {
	predicate := NewParameter("predicate", ScalarTypes.Bool.Type())
	left := NewParameterKind("left", ParameterTypeCommonNumber)
	right := NewParameterKind("right", ParameterTypeCommonNumber)
	common := GetCommonArgumentType(
		[]*Parameter{predicate, left, right},
		[]*TypeSymbol{ScalarTypes.Bool.Type(), ScalarTypes.Int.Type(), ScalarTypes.Long.Type()},
	)
	if !sameType(common, ScalarTypes.Long.Type()) {
		t.Fatalf("common argument type = %s, want long", typeName(common))
	}

	common = GetCommonArgumentType(
		[]*Parameter{left, right},
		[]*TypeSymbol{ScalarTypes.Dynamic.Type(), ScalarTypes.Long.Type()},
	)
	if !sameType(common, ScalarTypes.Long.Type()) {
		t.Fatalf("common number with dynamic+long = %s, want long", typeName(common))
	}

	common = GetCommonArgumentType([]*Parameter{left}, []*TypeSymbol{ScalarTypes.Dynamic.Type()})
	if common != nil {
		t.Fatalf("common number with only dynamic = %s, want nil", typeName(common))
	}
}

func TestGetCommonArgumentTypeWithDefaultMirrorsUpstreamOptions(t *testing.T) {
	left := NewParameterKind("left", ParameterTypeCommonScalar)
	right := NewParameterKind("right", ParameterTypeCommonScalar)
	fallback := ScalarTypes.Unknown.Type()

	common := GetCommonArgumentTypeWithDefault(
		[]*Parameter{left, right},
		[]*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()},
		fallback,
		false,
	)
	if common != fallback {
		t.Fatalf("incompatible common scalar = %s, want default unknown", typeName(common))
	}

	number := NewParameterKind("number", ParameterTypeCommonNumber)
	common = GetCommonArgumentTypeWithDefault(
		[]*Parameter{number},
		[]*TypeSymbol{ScalarTypes.Dynamic.Type()},
		fallback,
		true,
	)
	if common != fallback {
		t.Fatalf("ignore-dynamic fallback common number = %s, want default unknown", typeName(common))
	}

	common = GetCommonArgumentTypeWithDefault(
		[]*Parameter{number, number},
		[]*TypeSymbol{ScalarTypes.Dynamic.Type(), ScalarTypes.Long.Type()},
		fallback,
		true,
	)
	if !sameType(common, ScalarTypes.Long.Type()) {
		t.Fatalf("ignore-dynamic common number = %s, want long", typeName(common))
	}
}

func TestGetCommonColumnTypeWithDefault(t *testing.T) {
	fallback := ScalarTypes.Unknown.Type()
	if got := GetCommonColumnTypeWithDefault(nil, ConversionPromotable, fallback); got != fallback {
		t.Fatalf("empty common column type = %s, want default unknown", typeName(got))
	}
	columns := []*ColumnSymbol{
		NewColumnSymbol("Value", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()),
	}
	if got := GetCommonColumnTypeWithDefault(columns, ConversionPromotable, fallback); got != fallback {
		t.Fatalf("incompatible common column type = %s, want default unknown", typeName(got))
	}
	if got := GetCommonColumnTypeWithDefault(columns, ConversionDynamic, fallback); !sameType(got, ScalarTypes.Dynamic.Type()) {
		t.Fatalf("dynamic common column type = %s, want dynamic", typeName(got))
	}
}

func TestTypeFactsUnionColumnsMirrorsSourceOrderingAndIdentity(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Int.Type())
	tenant := NewColumnSymbol("Tenant", ScalarTypes.String.Type())
	columns := UnionColumns(
		[]*ColumnSymbol{account, score},
		[]*ColumnSymbol{NewColumnSymbol("Score", ScalarTypes.Long.Type()), tenant},
	)
	if len(columns) != 3 || columns[0] != account || columns[1] != score || columns[2] != tenant {
		t.Fatalf("union columns = %#v, want original A columns then unique B column", columns)
	}
	if columns[1].Type != ScalarTypes.Int.Type() {
		t.Fatalf("same-name successful common-type probe changed first-list type to %s", columns[1].Type.SymbolName())
	}

	firstB := NewColumnSymbol("OnlyB", ScalarTypes.String.Type())
	lastB := NewColumnSymbol("OnlyB", ScalarTypes.Long.Type())
	duplicateB := UnionColumns(nil, []*ColumnSymbol{firstB, lastB})
	if len(duplicateB) != 1 || duplicateB[0] != firstB || duplicateB[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("duplicate second-list union = %#v, want first second-list column identity", duplicateB)
	}
}

func TestTypeFactsIntersectColumnsUsesLastSecondTypeAndFirstOrder(t *testing.T) {
	first := NewColumnSymbol("Value", ScalarTypes.Int.Type())
	second := NewColumnSymbol("Value", ScalarTypes.String.Type())
	columns := IntersectColumns(
		[]*ColumnSymbol{first, second, NewColumnSymbol("OnlyA", ScalarTypes.Bool.Type())},
		[]*ColumnSymbol{
			NewColumnSymbol("Value", ScalarTypes.String.Type()),
			NewColumnSymbol("Value", ScalarTypes.Long.Type()),
		},
	)
	if len(columns) != 2 {
		t.Fatalf("intersect columns = %#v, want both first-list duplicates", columns)
	}
	if columns[0].SymbolName() != "Value" || columns[0].Type != ScalarTypes.Long.Type() {
		t.Fatalf("first intersect column = %#v, want promoted long from last second-list type", columns[0])
	}
	if columns[1].SymbolName() != "Value" || columns[1].Type != ScalarTypes.Dynamic.Type() {
		t.Fatalf("second intersect column = %#v, want dynamic string/long common type", columns[1])
	}
}

func typeName(typ *TypeSymbol) string {
	if typ == nil {
		return "<nil>"
	}
	return typ.SymbolName()
}
