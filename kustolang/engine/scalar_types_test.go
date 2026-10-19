package kustolang

import "testing"

func TestScalarTypeAliases(t *testing.T) {
	tests := map[string]string{
		"boolean": "bool",
		"int64":   "long",
		"double":  "real",
		"date":    "datetime",
		"uuid":    "guid",
	}
	for alias, canonical := range tests {
		if got := ScalarTypes.CanonicalName(alias); got != canonical {
			t.Fatalf("CanonicalName(%q) = %q, want %q", alias, got, canonical)
		}
	}
}

func TestScalarWidening(t *testing.T) {
	if !ScalarTypes.Long.IsWiderThan(ScalarTypes.Int) {
		t.Fatal("long should be wider than int")
	}
	if !ScalarTypes.Real.IsWiderThan(ScalarTypes.Long) {
		t.Fatal("real should be wider than long")
	}
	if ScalarTypes.Int.IsWiderThan(ScalarTypes.Real) {
		t.Fatal("int should not be wider than real")
	}
}

func TestDynamicScalarCatalog(t *testing.T) {
	if ScalarTypes.Dynamic.SymbolKind() != SymbolKindPrimitive {
		t.Fatalf("dynamic kind = %s, want primitive", ScalarTypes.Dynamic.SymbolKind())
	}
	if ScalarTypes.DynamicArray.SymbolKind() != SymbolKindArray {
		t.Fatalf("dynamic array kind = %s", ScalarTypes.DynamicArray.SymbolKind())
	}
	if ScalarTypes.DynamicBag.SymbolKind() != SymbolKindBag {
		t.Fatalf("dynamic bag kind = %s", ScalarTypes.DynamicBag.SymbolKind())
	}
	if ScalarTypes.DynamicArray.ElementType != ScalarTypes.Dynamic.Type() {
		t.Fatalf("dynamic array element type = %#v", ScalarTypes.DynamicArray.ElementType)
	}
	if ScalarTypes.DynamicArrayOfLong.ElementType != ScalarTypes.Long.Type() {
		t.Fatalf("dynamic long array element type = %#v", ScalarTypes.DynamicArrayOfLong.ElementType)
	}

	if got := ScalarTypes.GetDynamic(ScalarTypes.Int.Type()); got != ScalarTypes.DynamicLong.Type() {
		t.Fatalf("GetDynamic(int) = %#v, want dynamic-long", got)
	}
	if got := ScalarTypes.GetDynamic(ScalarTypes.String.Type()); got != ScalarTypes.DynamicString.Type() {
		t.Fatalf("GetDynamic(string) = %#v, want dynamic-string", got)
	}
	customPrimitive := NewDynamicPrimitiveSymbol(ScalarTypes.Decimal)
	if got, ok := DynamicPrimitiveForType(customPrimitive.Type()); !ok || got != customPrimitive {
		t.Fatalf("custom dynamic primitive back-reference = %#v ok=%t", got, ok)
	}
	if got := ScalarTypes.GetDynamic(customPrimitive.Type()); got != customPrimitive.Type() {
		t.Fatalf("GetDynamic(custom primitive) = %#v, want original type", got)
	}
	if got := ScalarTypes.GetDynamicArray(customPrimitive.Type()); got != ScalarTypes.DynamicArray {
		t.Fatalf("GetDynamicArray(custom dynamic decimal) = %#v, want generic array", got)
	}
	customDynamicLong := NewDynamicPrimitiveSymbol(ScalarTypes.Long)
	if got := ScalarTypes.GetDynamicArray(customDynamicLong.Type()); got != ScalarTypes.DynamicArrayOfLong {
		t.Fatalf("GetDynamicArray(custom dynamic long) = %#v, want long array", got)
	}
	if got := ScalarTypes.GetDynamicArray(ScalarTypes.Int.Type()); got != ScalarTypes.DynamicArrayOfLong {
		t.Fatalf("GetDynamicArray(int) = %#v, want dynamic array of long", got)
	}
	if got := ScalarTypes.GetDynamicArray(ScalarTypes.DynamicString.Type()); got != ScalarTypes.DynamicArrayOfString {
		t.Fatalf("GetDynamicArray(dynamic-string) = %#v, want dynamic array of string", got)
	}
	if got := ScalarTypes.GetDynamicArray(ScalarTypes.Decimal.Type()); got != ScalarTypes.DynamicArray {
		t.Fatalf("GetDynamicArray(decimal) = %#v, want generic dynamic array", got)
	}
	typedBag := ScalarTypes.GetDynamicBag(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	if got := ScalarTypes.GetDynamic(typedBag.Type()); got != typedBag.Type() {
		t.Fatalf("GetDynamic(typed bag) = %#v, want original type", got)
	}
	if got := ScalarTypes.GetDynamicArray(typedBag.Type()); got == ScalarTypes.DynamicArray || got.ElementType != typedBag.Type() {
		t.Fatalf("GetDynamicArray(typed bag) = %#v, want array preserving typed bag element", got)
	}
	typedArray := NewDynamicArraySymbol(typedBag.Type())
	if got := ScalarTypes.GetDynamic(typedArray.Type()); got != typedArray.Type() {
		t.Fatalf("GetDynamic(typed array) = %#v, want original type", got)
	}
	if got := ScalarTypes.GetDynamicArray(typedArray.Type()); got == ScalarTypes.DynamicArray || got.ElementType != typedArray.Type() {
		t.Fatalf("GetDynamicArray(typed array) = %#v, want array preserving typed array element", got)
	}
	if got := ScalarTypes.GetDynamic(ScalarTypes.GeoShape.Type()); got != ScalarTypes.GeoShape.Type() {
		t.Fatalf("GetDynamic(GeoShape) = %#v, want original type", got)
	}
	if got := ScalarTypes.GetDynamicBag(); got != ScalarTypes.DynamicBag {
		t.Fatalf("empty dynamic bag should reuse singleton")
	}
	if prop, ok := ScalarTypes.GeoShape.TryGetProperty("coordinates"); !ok || prop.Type != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("geo shape coordinates property = %#v ok=%t", prop, ok)
	}
}

func TestDynamicTypeBackReferenceIdentity(t *testing.T) {
	primitive := NewDynamicPrimitiveSymbol(ScalarTypes.Decimal)
	array := NewDynamicArraySymbol(primitive.Type())
	bag := NewDynamicBagSymbol(NewColumnSymbol("Values", array.Type()))

	if got, ok := DynamicPrimitiveForType(primitive.Type()); !ok || got != primitive {
		t.Fatalf("primitive owner = %#v ok=%t, want original", got, ok)
	}
	if got, ok := DynamicArrayForType(array.Type()); !ok || got != array {
		t.Fatalf("array owner = %#v ok=%t, want original", got, ok)
	}
	if got, ok := DynamicBagForType(bag.Type()); !ok || got != bag {
		t.Fatalf("bag owner = %#v ok=%t, want original", got, ok)
	}
	if got, ok := DynamicBagForType(ScalarTypes.GeoShape.Type()); !ok || got != ScalarTypes.GeoShape {
		t.Fatalf("GeoShape owner = %#v ok=%t, want catalog singleton", got, ok)
	}

	copiedPrimitiveType := *primitive.Type()
	if got, ok := DynamicPrimitiveForType(&copiedPrimitiveType); ok || got != nil {
		t.Fatalf("copied primitive type owner = %#v ok=%t, want nil,false", got, ok)
	}
	copiedArrayType := *array.Type()
	if got, ok := DynamicArrayForType(&copiedArrayType); ok || got != nil {
		t.Fatalf("copied array type owner = %#v ok=%t, want nil,false", got, ok)
	}
	copiedBagType := *bag.Type()
	if got, ok := DynamicBagForType(&copiedBagType); ok || got != nil {
		t.Fatalf("copied bag type owner = %#v ok=%t, want nil,false", got, ok)
	}

	copiedPrimitive := *primitive
	if got, ok := DynamicPrimitiveForType(copiedPrimitive.Type()); ok || got != nil {
		t.Fatalf("copied primitive owner = %#v ok=%t, want nil,false", got, ok)
	}
	copiedArray := *array
	if got, ok := DynamicArrayForType(copiedArray.Type()); ok || got != nil {
		t.Fatalf("copied array owner = %#v ok=%t, want nil,false", got, ok)
	}
	copiedBag := *bag
	if got, ok := DynamicBagForType(copiedBag.Type()); ok || got != nil {
		t.Fatalf("copied bag owner = %#v ok=%t, want nil,false", got, ok)
	}
}

func TestDynamicBagAndTupleSymbols(t *testing.T) {
	source := SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindIdentifierToken, Text: "source"})
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	bag := ScalarTypes.GetDynamicBag(account)
	if len(bag.Properties()) != 1 || len(bag.Members()) != 1 {
		t.Fatalf("dynamic bag members/properties not preserved")
	}
	if got, ok := bag.TryGetProperty("Account"); !ok || got != account {
		t.Fatalf("dynamic bag property lookup = %#v ok=%t", got, ok)
	}
	nested := ScalarTypes.GetDynamicBag(NewColumnSymbol("Family", ScalarTypes.String.Type()))
	nestedParent := ScalarTypes.GetDynamicBag(NewColumnSymbol("Browser", nested.Type()))
	nestedProperty, ok := nestedParent.TryGetProperty("Browser")
	if !ok {
		t.Fatal("nested dynamic bag property was not found")
	}
	if got, ok := DynamicBagForType(nestedProperty.Type); !ok || got != nested {
		t.Fatalf("dynamic bag type back-reference = %#v ok=%t, want nested bag", got, ok)
	}
	if got := dynamicBagSymbolFromSymbol(nestedProperty.Type); got != nested {
		t.Fatalf("dynamicBagSymbolFromSymbol returned %#v, want nested bag", got)
	}
	updated := bag.AddOrUpdateProperty(NewColumnSymbol("Account", ScalarTypes.Long.Type()))
	if got, ok := updated.TryGetProperty("Account"); !ok || got.Type != ScalarTypes.Long.Type() {
		t.Fatalf("dynamic bag property update = %#v ok=%t", got, ok)
	}
	added := updated.AddOrUpdateProperty(NewColumnSymbol("TenantId", ScalarTypes.String.Type()))
	if len(added.Properties()) != 2 {
		t.Fatalf("dynamic bag add property count = %d", len(added.Properties()))
	}
	sourced := added.WithSource(source)
	if sourced.Properties()[0].Source != source || added.Properties()[0].Source == source {
		t.Fatalf("dynamic bag WithSource should copy properties with source")
	}

	tuple := ScalarTypes.GetTuple(account)
	if tuple.SymbolKind() != SymbolKindTuple || !tuple.IsReducibleToScalar() {
		t.Fatalf("tuple metadata = kind:%s reducible:%t", tuple.SymbolKind(), tuple.IsReducibleToScalar())
	}
	if owner, ok := TupleForType(tuple.Type()); !ok || owner != tuple {
		t.Fatalf("tuple type owner = %#v ok=%t, want original tuple", owner, ok)
	}
	copiedTupleType := *tuple.Type()
	if owner, ok := TupleForType(&copiedTupleType); ok || owner != nil {
		t.Fatalf("copied tuple type owner = %#v ok=%t, want nil,false", owner, ok)
	}
	if tuple.Columns()[0] != account || len(tuple.Members()) != 1 {
		t.Fatalf("tuple columns/members not preserved")
	}
}

func TestDynamicBagTryGetPropertyReturnsLastDuplicate(t *testing.T) {
	first := NewColumnSymbol("Value", ScalarTypes.String.Type())
	last := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	bag := NewDynamicBagSymbol(first, last)

	got, ok := bag.TryGetProperty("Value")
	if !ok || got != last {
		t.Fatalf("duplicate property lookup = %#v ok=%t, want last property %#v", got, ok, last)
	}
}

func TestDynamicBagAddOrUpdatePropertyPreservesUnrelatedProperties(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.String.Type())
	c := NewColumnSymbol("C", ScalarTypes.String.Type())
	b2 := NewColumnSymbol("B", ScalarTypes.Long.Type())
	bag := NewDynamicBagSymbol(a, b, c)

	got := bag.AddOrUpdateProperty(b2).Properties()
	want := []*ColumnSymbol{a, b2, c}
	if !sameColumnSlice(got, want) {
		t.Fatalf("updated properties = %#v, want [A, B2, C]", got)
	}
}

func TestErrorAndVoidSymbols(t *testing.T) {
	if Error.SymbolKind() != SymbolKindError || !Error.IsError() || Error.SymbolTabularity() != TabularityNone {
		t.Fatalf("error symbol metadata = kind:%s error:%t tabularity:%s", Error.SymbolKind(), Error.IsError(), Error.SymbolTabularity())
	}
	if Void.SymbolKind() != SymbolKindVoid || Void.IsError() || Void.SymbolTabularity() != TabularityNone {
		t.Fatalf("void symbol metadata = kind:%s error:%t tabularity:%s", Void.SymbolKind(), Void.IsError(), Void.SymbolTabularity())
	}
}
