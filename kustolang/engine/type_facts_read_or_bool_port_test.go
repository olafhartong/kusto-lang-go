package kustolang

// Tests for the pinned TypeFacts.IsAnyScalarExceptReadOrBool API spelling.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import "testing"

func TestIsAnyScalarExceptReadOrBoolMatchesPinnedPredicate(t *testing.T) {
	tests := []struct {
		name   string
		symbol Symbol
		want   bool
	}{
		{"int", ScalarTypes.Int.Type(), true},
		{"string", ScalarTypes.String.Type(), true},
		{"real", ScalarTypes.Real.Type(), false},
		{"bool", ScalarTypes.Bool.Type(), false},
		{"dynamic real", ScalarTypes.DynamicReal.Type(), false},
		{"dynamic bool", ScalarTypes.DynamicBool.Type(), false},
		{"table", NewTableSymbol("T"), false},
		{"nil", nil, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsAnyScalarExceptReadOrBool(test.symbol); got != test.want {
				t.Fatalf("IsAnyScalarExceptReadOrBool() = %t, want %t", got, test.want)
			}
			if got, corrected := IsAnyScalarExceptReadOrBool(test.symbol), IsAnyScalarExceptRealOrBool(test.symbol); got != corrected {
				t.Fatalf("pinned spelling = %t, corrected spelling = %t", got, corrected)
			}
		})
	}
}
