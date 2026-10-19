package kustolang

import (
	"context"
	"errors"
	"testing"
)

func TestCancellationTokenDefaultDoesNotCancel(t *testing.T) {
	var token CancellationToken
	if err := token.Err(); err != nil {
		t.Fatalf("default token Err = %v, want nil", err)
	}
	token.ThrowIfCancellationRequested()
}

func TestCancellationTokenUsesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	token := NewCancellationToken(ctx)
	if err := token.Err(); err != nil {
		t.Fatalf("uncanceled context Err = %v, want nil", err)
	}

	cancel()
	err := token.Err()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context Err = %v, want context.Canceled", err)
	}
	var canceled OperationCanceledError
	if !errors.As(err, &canceled) {
		t.Fatalf("canceled context Err type = %T, want OperationCanceledError", err)
	}
}

func TestCancellationTokenUsesCallback(t *testing.T) {
	canceled := false
	token := NewCancellationTokenFunc(func() bool {
		return canceled
	})
	if err := token.Err(); err != nil {
		t.Fatalf("uncanceled callback Err = %v, want nil", err)
	}

	canceled = true
	var canceledErr OperationCanceledError
	if err := token.Err(); !errors.As(err, &canceledErr) {
		t.Fatalf("canceled callback Err = %v, want OperationCanceledError", err)
	}
}

func TestCancellationTokenThrowsWhenCanceled(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("ThrowIfCancellationRequested did not panic")
		}
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("panic value = %T, want error", recovered)
		}
		var canceled OperationCanceledError
		if !errors.As(err, &canceled) {
			t.Fatalf("panic error = %v, want OperationCanceledError", err)
		}
	}()

	NewCancellationTokenFunc(func() bool { return true }).ThrowIfCancellationRequested()
}
