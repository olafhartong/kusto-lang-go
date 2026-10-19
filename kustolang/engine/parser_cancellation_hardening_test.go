package kustolang

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

func TestLexerCancellationStopsBetweenTokens(t *testing.T) {
	text := strings.Repeat("T | where X == 1 | project X; ", 1_000)
	checks := 0
	token := NewCancellationTokenFunc(func() bool {
		checks++
		return checks >= 128
	})

	assertOperationCanceledPanic(t, func() {
		LexWithOptionsAndCancellation(text, DefaultParseOptions, token)
	})
	if checks < 128 {
		t.Fatalf("cancellation checks = %d, want at least 128", checks)
	}
}

func TestLexerCancellationStopsDuringHugeSingleToken(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "unterminated-string", text: `"` + strings.Repeat("x", 1<<20)},
		{name: "unterminated-multiline-string", text: "```" + strings.Repeat("x", 1<<20)},
		{name: "unterminated-comment", text: "//" + strings.Repeat("x", 1<<20)},
		{name: "identifier", text: strings.Repeat("x", 1<<20)},
		{name: "number", text: strings.Repeat("7", 1<<20)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checks := 0
			token := NewCancellationTokenFunc(func() bool {
				checks++
				return checks == 8
			})
			assertOperationCanceledPanic(t, func() {
				LexWithOptionsAndCancellation(test.text, DefaultParseOptions, token)
			})
			if checks != 8 {
				t.Fatalf("cancellation checks = %d, want 8", checks)
			}
		})
	}
}

func TestKustoCodePreflightAllowsOrdinaryNestedQueries(t *testing.T) {
	text := "print x=" + strings.Repeat("f(", 16) + "1" + strings.Repeat(")", 16)
	for _, parse := range []func(string, *GlobalState) *KustoCode{ParseWithGlobals, ParseWithGlobals} {
		code := parse(text, DefaultGlobalState)
		if code == nil || code.ParseLimitExceeded() {
			t.Fatal("ordinary nested query exceeded the parser limit")
		}
	}
}

func TestKustoCodePreflightIgnoresDelimitersInStringsAndComments(t *testing.T) {
	globals := GlobalStateWithProperty(DefaultGlobalState, Properties.MaxAnalysisDepth, 8)
	delimiters := strings.Repeat("([{", 10_000)
	for _, text := range []string{
		`print x="` + delimiters + `"`,
		"//" + delimiters + "\nprint x=1",
	} {
		for _, parse := range []func(string, *GlobalState) *KustoCode{ParseWithGlobals, ParseWithGlobals} {
			code := parse(text, globals)
			if code == nil || code.ParseLimitExceeded() {
				t.Fatalf("ParseLimitExceeded(%q...) = true", text[:16])
			}
		}
	}
}

func TestCancellationPollingDoesNotLeakGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()
	text := strings.Repeat("x", 1<<16)
	for i := 0; i < 100; i++ {
		checks := 0
		token := NewCancellationTokenFunc(func() bool {
			checks++
			return checks == 2
		})
		assertOperationCanceledPanic(t, func() {
			LexWithOptionsAndCancellation(text, DefaultParseOptions, token)
		})
	}
	runtime.GC()
	runtime.Gosched()
	if after := runtime.NumGoroutine(); after > before+1 {
		t.Fatalf("goroutines after cancellation = %d, before = %d", after, before)
	}
}

func hardeningHasDiagnosticCode(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestKustoCodeParseEntryPointsObserveCancellation(t *testing.T) {
	canceled := NewCancellationTokenFunc(func() bool { return true })
	for _, parse := range []func(){
		func() { ParseWithGlobalsAndCancellation("print x=1", nil, canceled) },
		func() { ParseWithGlobalsAndCancellation("print x=1", nil, canceled) },
		func() { ParseAndAnalyzeWithCancellation("print x=1", nil, canceled) },
	} {
		assertOperationCanceledPanic(t, parse)
	}
}

func assertOperationCanceledPanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("operation did not panic for cancellation")
		}
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("cancellation panic = %#v, want error", recovered)
		}
		var canceled OperationCanceledError
		if !errors.As(err, &canceled) {
			t.Fatalf("cancellation panic = %v, want OperationCanceledError", err)
		}
	}()
	action()
}
