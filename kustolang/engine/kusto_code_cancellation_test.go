package kustolang

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestKustoCodeScopeTraversalCancellation(t *testing.T) {
	var text strings.Builder
	for i := 0; i < 80; i++ {
		fmt.Fprintf(&text, "let value%d = %d;\n", i, i)
	}
	text.WriteString("print result=value79")
	code := ParseAndAnalyze(text.String())
	position := len(text.String())

	tests := []struct {
		name string
		call func(CancellationToken)
	}{
		{
			name: "symbols",
			call: func(token CancellationToken) {
				code.GetSymbolsInScopeWithCancellation(position, SymbolMatchAny, IncludeFunctionKindAll, token)
			},
		},
		{
			name: "speculative symbol",
			call: func(token CancellationToken) {
				code.GetSpeculativeReferencedSymbolWithCancellation(position, "not_present", SymbolMatchAny, token)
			},
		},
		{
			name: "columns",
			call: func(token CancellationToken) {
				code.GetColumnsInScopeWithCancellation(position, token)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checks := 0
			test.call(NewCancellationTokenFunc(func() bool {
				checks++
				return false
			}))
			if checks < 10 {
				t.Fatalf("only %d cancellation checks; traversal checks were not propagated", checks)
			}

			cancelAt := checks / 2
			checks = 0
			assertKustoCodeCanceled(t, func() {
				test.call(NewCancellationTokenFunc(func() bool {
					checks++
					return checks >= cancelAt
				}))
			})
			if checks != cancelAt {
				t.Fatalf("canceled after %d checks, want deterministic midpoint %d", checks, cancelAt)
			}
		})
	}
}

func assertKustoCodeCanceled(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			t.Fatal("operation did not cancel")
		}
		err, ok := panicValue.(error)
		if !ok {
			t.Fatalf("panic = %#v, want OperationCanceledError", panicValue)
		}
		var canceled OperationCanceledError
		if !errors.As(err, &canceled) {
			t.Fatalf("panic = %T: %v, want OperationCanceledError", err, err)
		}
	}()
	call()
}
