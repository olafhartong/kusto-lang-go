package kustolang

import "testing"

func TestParameterMatchKindValuesMatchUpstream(t *testing.T) {
	tests := []struct {
		name string
		got  ParameterMatchKind
		want ParameterMatchKind
	}{
		{"None", ParameterMatchNone, 0},
		{"Unknown", ParameterMatchUnknown, 1},
		{"Dynamic", ParameterMatchDynamic, 2},
		{"Compatible", ParameterMatchCompatible, 3},
		{"Promoted", ParameterMatchPromoted, 4},
		{"NotType", ParameterMatchNotType, 5},
		{"OneOfMany", ParameterMatchOneOfMany, 6},
		{"Scalar", ParameterMatchScalar, 7},
		{"Summable", ParameterMatchSummable, 8},
		{"Orderable", ParameterMatchOrderable, 9},
		{"Number", ParameterMatchNumber, 10},
		{"Integer", ParameterMatchInteger, 11},
		{"Tabular", ParameterMatchTabular, 12},
		{"Table", ParameterMatchTable, 13},
		{"Database", ParameterMatchDatabase, 14},
		{"Cluster", ParameterMatchCluster, 15},
		{"Exact", ParameterMatchExact, 16},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
	if !(ParameterMatchExact > ParameterMatchTable && ParameterMatchPromoted > ParameterMatchCompatible && ParameterMatchCompatible > ParameterMatchDynamic) {
		t.Fatal("parameter match ordering should preserve upstream specificity")
	}
}

func TestSignatureParameterMatchKindClassifications(t *testing.T) {
	signature := NewSignature(
		ScalarTypes.Bool.Type(),
		NewParameter("long", ScalarTypes.Long.Type()),
		NewParameter("int", ScalarTypes.Int.Type()),
		NewParameter("dynamic", ScalarTypes.Dynamic.Type()),
	)
	longParam := signature.Parameters[0]
	intParam := signature.Parameters[1]
	dynamicParam := signature.Parameters[2]

	if got := signature.ParameterMatchKind(nil, nil, longParam, nil, ScalarTypes.Long.Type(), false); got != ParameterMatchExact {
		t.Fatalf("exact long match = %d, want %d", got, ParameterMatchExact)
	}
	if got := signature.ParameterMatchKind(nil, nil, longParam, nil, ScalarTypes.Int.Type(), false); got != ParameterMatchPromoted {
		t.Fatalf("promoted int->long match = %d, want %d", got, ParameterMatchPromoted)
	}
	if got := signature.ParameterMatchKind(nil, nil, dynamicParam, nil, ScalarTypes.String.Type(), true); got != ParameterMatchDynamic {
		t.Fatalf("dynamic string->dynamic match = %d, want %d", got, ParameterMatchDynamic)
	}
	if got := signature.ParameterMatchKind(nil, nil, intParam, nil, ScalarTypes.Long.Type(), true); got != ParameterMatchCompatible {
		t.Fatalf("compatible long->int match = %d, want %d", got, ParameterMatchCompatible)
	}
	if got := signature.ParameterMatchKind(nil, nil, intParam, nil, ScalarTypes.String.Type(), true); got != ParameterMatchNone {
		t.Fatalf("non-matching string->int match = %d, want %d", got, ParameterMatchNone)
	}
}

func TestSignatureParameterMatchKindDoesNotTreatDynamicAsNumeric(t *testing.T) {
	signature := NewReturnKindSignature(
		ReturnTypeWidest,
		NewParameterKind("array", ParameterTypeDynamicArray),
		NewParameterKind("number", ParameterTypeNumber),
	)
	parameters := signature.Parameters
	argumentTypes := []Symbol{ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type()}

	if got := signature.ParameterMatchKind(parameters, argumentTypes, parameters[0], nil, argumentTypes[0], false); got != ParameterMatchOneOfMany {
		t.Fatalf("dynamic array match = %d, want %d", got, ParameterMatchOneOfMany)
	}
	if got := signature.ParameterMatchKind(parameters, argumentTypes, parameters[1], nil, argumentTypes[1], false); got != ParameterMatchNone {
		t.Fatalf("generic dynamic numeric match = %d, want %d", got, ParameterMatchNone)
	}
}

func TestSignatureParameterMatchKindDynamicPrimitiveMatrix(t *testing.T) {
	dynamicLong := ScalarTypes.DynamicLong.Type()
	tests := []struct {
		name      string
		parameter *Parameter
		want      ParameterMatchKind
	}{
		{name: "array", parameter: NewParameterKind("value", ParameterTypeDynamicArray), want: ParameterMatchNone},
		{name: "bag", parameter: NewParameterKind("value", ParameterTypeDynamicBag), want: ParameterMatchNone},
		{name: "number", parameter: NewParameterKind("value", ParameterTypeNumber), want: ParameterMatchNone},
		{name: "scalar", parameter: NewParameterKind("value", ParameterTypeScalar), want: ParameterMatchScalar},
		{name: "declared dynamic", parameter: NewParameter("value", ScalarTypes.Dynamic.Type()), want: ParameterMatchExact},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := NewSignature(ScalarTypes.Bool.Type(), tt.parameter)
			if got := signature.ParameterMatchKind(signature.Parameters, []Symbol{dynamicLong}, tt.parameter, nil, dynamicLong, false); got != tt.want {
				t.Fatalf("dynamic primitive match = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSignatureParameterMatchKindNotDynamicRejectsTypedDynamic(t *testing.T) {
	parameter := NewParameterKind("value", ParameterTypeNotDynamic)
	signature := NewSignature(ScalarTypes.Bool.Type(), parameter)
	tests := []struct {
		name string
		typ  Symbol
		want ParameterMatchKind
	}{
		{name: "string", typ: ScalarTypes.String.Type(), want: ParameterMatchNotType},
		{name: "generic", typ: ScalarTypes.Dynamic.Type(), want: ParameterMatchNone},
		{name: "primitive", typ: ScalarTypes.DynamicLong.Type(), want: ParameterMatchNone},
		{name: "array", typ: ScalarTypes.DynamicArrayOfLong.Type(), want: ParameterMatchNone},
		{name: "bag", typ: ScalarTypes.GetDynamicBag(NewColumnSymbol("Value", ScalarTypes.Long.Type())).Type(), want: ParameterMatchNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signature.ParameterMatchKind(signature.Parameters, []Symbol{tt.typ}, parameter, nil, tt.typ, false); got != tt.want {
				t.Fatalf("not-dynamic match = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSignatureParameterMatchKindConcreteTableUsesTabularRank(t *testing.T) {
	parameter := NewParameterKind("value", ParameterTypeTabular)
	signature := NewSignature(ScalarTypes.Bool.Type(), parameter)
	table := NewTableSymbol("T", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	if got := signature.ParameterMatchKind(signature.Parameters, []Symbol{table}, parameter, nil, table, false); got != ParameterMatchTabular {
		t.Fatalf("concrete table match = %d, want %d", got, ParameterMatchTabular)
	}
}
