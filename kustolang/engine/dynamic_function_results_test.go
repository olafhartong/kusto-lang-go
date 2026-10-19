package kustolang

import "testing"

func TestHasDynamicPrimitives(t *testing.T) {
	primitives := []*TypeSymbol{
		ScalarTypes.DynamicBool.Type(),
		ScalarTypes.DynamicLong.Type(),
		ScalarTypes.DynamicReal.Type(),
		ScalarTypes.DynamicDateTime.Type(),
		ScalarTypes.DynamicTimeSpan.Type(),
		ScalarTypes.DynamicGuid.Type(),
		ScalarTypes.DynamicString.Type(),
	}
	for _, primitive := range primitives {
		if !HasDynamicPrimitives([]*TypeSymbol{primitive}) {
			t.Fatalf("HasDynamicPrimitives(%p) = false, want true", primitive)
		}
	}

	if !HasDynamicPrimitives([]*TypeSymbol{ScalarTypes.Long.Type(), nil, ScalarTypes.DynamicString.Type(), ScalarTypes.Bool.Type()}) {
		t.Fatal("mixed ordinary and dynamic primitive arguments were not detected")
	}

	nonPrimitives := []*TypeSymbol{
		nil,
		&Error.TypeSymbol,
		ScalarTypes.Dynamic.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.DynamicArray.Type(),
		ScalarTypes.DynamicArrayOfLong.Type(),
		ScalarTypes.DynamicBag.Type(),
		ScalarTypes.GeoShape.Type(),
	}
	if HasDynamicPrimitives(nonPrimitives) {
		t.Fatal("nil, errors, ordinary scalars, arrays, or bags must not be treated as dynamic primitives")
	}
	if HasDynamicPrimitives(nil) {
		t.Fatal("nil argument types must not contain dynamic primitives")
	}
}

func TestGetUnwrappedDynamicPrimitives(t *testing.T) {
	typedBag := ScalarTypes.GetDynamicBag(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	nilPrimitive := NewDynamicPrimitiveSymbol(nil)
	types := []*TypeSymbol{
		ScalarTypes.DynamicBool.Type(),
		ScalarTypes.DynamicLong.Type(),
		ScalarTypes.DynamicReal.Type(),
		ScalarTypes.DynamicDateTime.Type(),
		ScalarTypes.DynamicTimeSpan.Type(),
		ScalarTypes.DynamicGuid.Type(),
		ScalarTypes.String.Type(),
		ScalarTypes.DynamicString.Type(),
		nilPrimitive.Type(),
		nil,
		&Error.TypeSymbol,
		ScalarTypes.DynamicArrayOfLong.Type(),
		typedBag.Type(),
	}
	want := []*TypeSymbol{
		ScalarTypes.Bool.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.Real.Type(),
		ScalarTypes.DateTime.Type(),
		ScalarTypes.TimeSpan.Type(),
		ScalarTypes.Guid.Type(),
		ScalarTypes.String.Type(),
		ScalarTypes.String.Type(),
		nil,
		nil,
		&Error.TypeSymbol,
		ScalarTypes.DynamicArrayOfLong.Type(),
		typedBag.Type(),
	}

	got := GetUnwrappedDynamicPrimitives(types)
	if len(got) != len(want) {
		t.Fatalf("unwrapped length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unwrapped[%d] = %p, want %p", i, got[i], want[i])
		}
	}
	if &got[0] == &types[0] {
		t.Fatal("unwrapping must return an independent slice")
	}
	if GetUnwrappedDynamicPrimitives(nil) != nil {
		t.Fatal("unwrapping nil must preserve a nil slice")
	}
}

func TestRefineDynamicFunctionResult(t *testing.T) {
	tests := []struct {
		name    string
		rebound *TypeSymbol
		want    *TypeSymbol
	}{
		{name: "bool", rebound: ScalarTypes.Bool.Type(), want: ScalarTypes.DynamicBool.Type()},
		{name: "int normalizes to long", rebound: ScalarTypes.Int.Type(), want: ScalarTypes.DynamicLong.Type()},
		{name: "long", rebound: ScalarTypes.Long.Type(), want: ScalarTypes.DynamicLong.Type()},
		{name: "real", rebound: ScalarTypes.Real.Type(), want: ScalarTypes.DynamicReal.Type()},
		{name: "datetime", rebound: ScalarTypes.DateTime.Type(), want: ScalarTypes.DynamicDateTime.Type()},
		{name: "timespan", rebound: ScalarTypes.TimeSpan.Type(), want: ScalarTypes.DynamicTimeSpan.Type()},
		{name: "guid", rebound: ScalarTypes.Guid.Type(), want: ScalarTypes.DynamicGuid.Type()},
		{name: "string", rebound: ScalarTypes.String.Type(), want: ScalarTypes.DynamicString.Type()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := RefineDynamicFunctionResult(ScalarTypes.Dynamic.Type(), test.rebound); got != test.want {
				t.Fatalf("RefineDynamicFunctionResult(dynamic, %p) = %p, want %p", test.rebound, got, test.want)
			}
		})
	}
}

func TestRefineDynamicFunctionResultNoRefinement(t *testing.T) {
	typedBag := ScalarTypes.GetDynamicBag(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	dynamic := ScalarTypes.Dynamic.Type()
	tests := []struct {
		name    string
		initial *TypeSymbol
		rebound *TypeSymbol
		want    *TypeSymbol
	}{
		{name: "nil initial", initial: nil, rebound: ScalarTypes.Long.Type(), want: nil},
		{name: "non-dynamic initial", initial: ScalarTypes.String.Type(), rebound: ScalarTypes.Long.Type(), want: ScalarTypes.String.Type()},
		{name: "nil rebound", initial: dynamic, rebound: nil, want: dynamic},
		{name: "error rebound", initial: dynamic, rebound: &Error.TypeSymbol, want: dynamic},
		{name: "unknown rebound", initial: dynamic, rebound: ScalarTypes.Unknown.Type(), want: dynamic},
		{name: "generic dynamic rebound", initial: dynamic, rebound: dynamic, want: dynamic},
		{name: "dynamic primitive rebound", initial: dynamic, rebound: ScalarTypes.DynamicLong.Type(), want: dynamic},
		{name: "dynamic array rebound", initial: dynamic, rebound: ScalarTypes.DynamicArrayOfLong.Type(), want: dynamic},
		{name: "dynamic bag rebound", initial: dynamic, rebound: typedBag.Type(), want: dynamic},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := RefineDynamicFunctionResult(test.initial, test.rebound); got != test.want {
				t.Fatalf("RefineDynamicFunctionResult(%p, %p) = %p, want %p", test.initial, test.rebound, got, test.want)
			}
		})
	}
}
