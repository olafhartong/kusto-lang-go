package kustolang

import (
	"math"
	"testing"
)

func TestValueAreEquivalentNilAndSameTypeValues(t *testing.T) {
	if !ValueAreEquivalent(nil, nil) {
		t.Fatal("nil values should be equivalent")
	}
	if ValueAreEquivalent(nil, 1) {
		t.Fatal("nil and non-nil values should not be equivalent")
	}
	if !ValueAreEquivalent(true, true) || ValueAreEquivalent(true, false) {
		t.Fatal("same-type bool equivalence mismatch")
	}
}

func TestValueAreEquivalentStringCaseSensitivity(t *testing.T) {
	if !ValueAreEquivalent("Alpha", "alpha") {
		t.Fatal("strings should compare case-insensitively by default")
	}
	if ValueAreEquivalent("Alpha", "alpha", true) {
		t.Fatal("case-sensitive string comparison should reject different case")
	}
	if !ValueAreEquivalent("True", true) {
		t.Fatal("string/bool comparison should use reference implementation-style bool text")
	}
	if !ValueAreEquivalent("42", 42) {
		t.Fatal("string/numeric comparison should compare converted text")
	}
}

func TestValueAreEquivalentNumericValues(t *testing.T) {
	if !ValueAreEquivalent(int32(5), int64(5)) {
		t.Fatal("integer numeric values should compare through int64")
	}
	if !ValueAreEquivalent(uint64(5), uint32(5)) {
		t.Fatal("uint64 numeric values should compare through uint64")
	}
	if !ValueAreEquivalent(float32(5), int32(5)) {
		t.Fatal("float32 comparison should compare after single-precision conversion")
	}
	if ValueAreEquivalent(float64(5.1), int64(5)) {
		t.Fatal("different float/int values should not compare equivalent")
	}
}

func TestValueAreEquivalentDifferentNonComparableTypes(t *testing.T) {
	if ValueAreEquivalent(struct{ A int }{A: 1}, struct{ B int }{B: 1}) {
		t.Fatal("different non-string/non-numeric types should not compare equivalent")
	}
}

func TestValueAreEquivalentReferenceIdentityTypedNilAndNaN(t *testing.T) {
	left := []int{1, 2}
	right := []int{1, 2}
	if ValueAreEquivalent(left, right) {
		t.Fatal("distinct equal-content slices should use reference identity")
	}
	if !ValueAreEquivalent(left, left) {
		t.Fatal("same slice should be equivalent")
	}
	var nilLeft *int
	var nilRight *int
	if !ValueAreEquivalent(nilLeft, nilRight) || ValueAreEquivalent(nilLeft, new(int)) {
		t.Fatal("typed nil reference semantics differ from object null")
	}
	if !ValueAreEquivalent(math.NaN(), math.NaN()) {
		t.Fatal("same-type NaN values should be equivalent")
	}
}

func TestValueAreEquivalentNegativeToUInt64Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative Int64 to UInt64 comparison did not panic")
		}
	}()
	ValueAreEquivalent(int64(-1), uint64(^uint64(0)))
}
