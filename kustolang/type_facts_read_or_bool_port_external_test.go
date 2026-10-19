package kustolang_test

// External-package tests for the pinned TypeFacts public API spelling.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicIsAnyScalarExceptReadOrBool(t *testing.T) {
	tests := []struct {
		name   string
		symbol kustolang.Symbol
		want   bool
	}{
		{"long", kustolang.ScalarTypes.Long.Type(), true},
		{"string", kustolang.ScalarTypes.String.Type(), true},
		{"real", kustolang.ScalarTypes.Real.Type(), false},
		{"bool", kustolang.ScalarTypes.Bool.Type(), false},
		{"dynamic real", kustolang.ScalarTypes.DynamicReal.Type(), false},
		{"dynamic bool", kustolang.ScalarTypes.DynamicBool.Type(), false},
		{"table", kustolang.NewTableSymbol("T"), false},
		{"nil", nil, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := kustolang.IsAnyScalarExceptReadOrBool(test.symbol)
			if got != test.want {
				t.Fatalf("IsAnyScalarExceptReadOrBool() = %t, want %t", got, test.want)
			}
			if corrected := kustolang.IsAnyScalarExceptRealOrBool(test.symbol); got != corrected {
				t.Fatalf("pinned spelling = %t, corrected spelling = %t", got, corrected)
			}
		})
	}
}
