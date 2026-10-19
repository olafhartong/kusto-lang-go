package kustolang

import "context"

// Ported from src/Kusto.Language/Utils/Cancellation.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type OperationCanceledError struct {
	Cause error
}

func (e OperationCanceledError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "operation canceled"
}

func (e OperationCanceledError) Unwrap() error {
	return e.Cause
}

type CancellationToken struct {
	context  context.Context
	canceled func() bool
}

func NewCancellationToken(ctx context.Context) CancellationToken {
	return CancellationToken{context: ctx}
}

func NewCancellationTokenFunc(canceled func() bool) CancellationToken {
	return CancellationToken{canceled: canceled}
}

func (t CancellationToken) Err() error {
	if t.canceled != nil && t.canceled() {
		return OperationCanceledError{}
	}
	if t.context != nil {
		if err := t.context.Err(); err != nil {
			return OperationCanceledError{Cause: err}
		}
	}
	return nil
}

func (t CancellationToken) ThrowIfCancellationRequested() {
	if err := t.Err(); err != nil {
		panic(err)
	}
}
