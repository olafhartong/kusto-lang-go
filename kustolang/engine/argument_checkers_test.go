package kustolang

import (
	"strings"
	"testing"
)

func TestCheckArgumentNullReturnsValue(t *testing.T) {
	value := &listExtensionItem{name: "value"}
	if got := CheckArgumentNull(value, "value"); got != value {
		t.Fatal("CheckArgumentNull should return the original value")
	}
	if got := CheckArgumentNull("text", "value"); got != "text" {
		t.Fatalf("non-pointer value = %q, want text", got)
	}
}

func TestCheckArgumentNullPanicsForNil(t *testing.T) {
	var value *listExtensionItem
	err := recoverPanic(t, func() {
		CheckArgumentNull(value, "value")
	})
	if _, ok := err.(ArgumentNullError); !ok {
		t.Fatalf("panic = %#v, want ArgumentNullError", err)
	}
	if !strings.Contains(err.(ArgumentNullError).Error(), "value") {
		t.Fatalf("panic message = %q, want parameter name", err.(ArgumentNullError).Error())
	}
}

func TestCheckArgumentNullOrEmpty(t *testing.T) {
	values := []int{1}
	if got := CheckArgumentNullOrEmpty(values, "values"); len(got) != 1 || got[0] != 1 {
		t.Fatalf("valid values = %#v, want [1]", got)
	}

	err := recoverPanic(t, func() {
		CheckArgumentNullOrEmpty[int](nil, "values")
	})
	if _, ok := err.(ArgumentNullError); !ok {
		t.Fatalf("nil panic = %#v, want ArgumentNullError", err)
	}

	err = recoverPanic(t, func() {
		CheckArgumentNullOrEmpty([]int{}, "values")
	})
	if argErr, ok := err.(ArgumentError); !ok || !strings.Contains(argErr.Error(), "at least one element") {
		t.Fatalf("empty panic = %#v, want element-count ArgumentError", err)
	}
}

func TestCheckArgumentNullOrElementNull(t *testing.T) {
	first := &listExtensionItem{name: "first"}
	second := &listExtensionItem{name: "second"}
	values := []*listExtensionItem{first, second}
	if got := CheckArgumentNullOrElementNull(values, "values"); len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("valid values = %#v, want original values", got)
	}

	err := recoverPanic(t, func() {
		CheckArgumentNullOrElementNull[*listExtensionItem](nil, "values")
	})
	if _, ok := err.(ArgumentNullError); !ok {
		t.Fatalf("nil list panic = %#v, want ArgumentNullError", err)
	}

	err = recoverPanic(t, func() {
		CheckArgumentNullOrElementNull([]*listExtensionItem{first, nil}, "values")
	})
	if argErr, ok := err.(ArgumentError); !ok || !strings.Contains(argErr.Error(), "Element 1") {
		t.Fatalf("nil element panic = %#v, want element ArgumentError", err)
	}
}

func TestCheckArgumentNullOrEmptyOrElementNull(t *testing.T) {
	value := &listExtensionItem{name: "value"}
	if got := CheckArgumentNullOrEmptyOrElementNull([]*listExtensionItem{value}, "values"); len(got) != 1 || got[0] != value {
		t.Fatalf("valid combined check = %#v, want original value", got)
	}

	err := recoverPanic(t, func() {
		CheckArgumentNullOrEmptyOrElementNull([]*listExtensionItem{}, "values")
	})
	if argErr, ok := err.(ArgumentError); !ok || !strings.Contains(argErr.Error(), "at least one element") {
		t.Fatalf("empty combined panic = %#v, want element-count ArgumentError", err)
	}

	err = recoverPanic(t, func() {
		CheckArgumentNullOrEmptyOrElementNull([]*listExtensionItem{nil}, "values")
	})
	if argErr, ok := err.(ArgumentError); !ok || !strings.Contains(argErr.Error(), "Element 0") {
		t.Fatalf("nil element combined panic = %#v, want element ArgumentError", err)
	}
}

func recoverPanic(t *testing.T, fn func()) any {
	t.Helper()
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		fn()
	}()
	if recovered == nil {
		t.Fatal("expected panic")
	}
	return recovered
}
