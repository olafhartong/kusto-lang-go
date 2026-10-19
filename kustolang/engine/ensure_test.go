package kustolang

import (
	"strings"
	"testing"
)

func TestEnsureSuccessfulChecksDoNotPanic(t *testing.T) {
	value := &listExtensionItem{name: "value"}
	EnsureIsTrue(true)
	EnsureAreEqual(3, 3)
	EnsureNotNil(value)
	EnsureIsNil((*listExtensionItem)(nil))
	EnsureArgumentNotNil(value, "value")
	EnsureElementsNotNil([]*listExtensionItem{value}, "values")
}

func TestEnsureIsTruePanicsWithDefaultAndCustomMessages(t *testing.T) {
	err := recoverPanic(t, func() {
		EnsureIsTrue(false)
	})
	if opErr, ok := err.(InvalidOperationError); !ok || opErr.Error() != "Expected true" {
		t.Fatalf("default IsTrue panic = %#v", err)
	}

	err = recoverPanic(t, func() {
		EnsureIsTrue(false, "custom")
	})
	if opErr, ok := err.(InvalidOperationError); !ok || opErr.Error() != "custom" {
		t.Fatalf("custom IsTrue panic = %#v", err)
	}
}

func TestEnsureAreEqualPanicsWithExpectedAndActual(t *testing.T) {
	err := recoverPanic(t, func() {
		EnsureAreEqual(1, 2)
	})
	opErr, ok := err.(InvalidOperationError)
	if !ok || !strings.Contains(opErr.Error(), "Expected: 1 actual: 2") {
		t.Fatalf("AreEqual panic = %#v", err)
	}
}

func TestEnsureNilChecksPanic(t *testing.T) {
	err := recoverPanic(t, func() {
		EnsureNotNil((*listExtensionItem)(nil))
	})
	if opErr, ok := err.(InvalidOperationError); !ok || opErr.Error() != "Expected not null" {
		t.Fatalf("NotNil panic = %#v", err)
	}

	err = recoverPanic(t, func() {
		EnsureIsNil(&listExtensionItem{name: "value"})
	})
	if opErr, ok := err.(InvalidOperationError); !ok || opErr.Error() != "Expected null" {
		t.Fatalf("IsNil panic = %#v", err)
	}
}

func TestEnsureArgumentNotNilPanicsWithArgumentNullError(t *testing.T) {
	err := recoverPanic(t, func() {
		EnsureArgumentNotNil((*listExtensionItem)(nil), "value")
	})
	if argErr, ok := err.(ArgumentNullError); !ok || !strings.Contains(argErr.Error(), "value") {
		t.Fatalf("ArgumentNotNil panic = %#v", err)
	}
}

func TestEnsureElementsNotNilPanicsWithIndexedMessage(t *testing.T) {
	err := recoverPanic(t, func() {
		EnsureElementsNotNil([]*listExtensionItem{{name: "first"}, nil}, "values")
	})
	if opErr, ok := err.(InvalidOperationError); !ok || opErr.Error() != "values[1] is nil" {
		t.Fatalf("ElementsNotNil panic = %#v", err)
	}
}
