package kustolang

import "testing"

func TestOptionalZeroValueAndValue(t *testing.T) {
	var empty Optional[int]
	if empty.HasValue {
		t.Fatal("zero-value Optional should not have a value")
	}
	if empty.Value != 0 {
		t.Fatalf("zero-value Optional value = %d, want 0", empty.Value)
	}

	withValue := NewOptional("value")
	if !withValue.HasValue {
		t.Fatal("NewOptional should have a value")
	}
	if withValue.Value != "value" {
		t.Fatalf("optional value = %q, want value", withValue.Value)
	}
}
