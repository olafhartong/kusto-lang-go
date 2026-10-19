package kustolang

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestKustoCodeCancellationAwareAPIs(t *testing.T) {
	canceled := NewCancellationTokenFunc(func() bool { return true })
	assertOperationCanceled(t, func() {
		ParseAndAnalyzeWithCancellation("print x=1", nil, canceled)
	})

	unbound := Parse("print x=1")
	assertOperationCanceled(t, func() {
		unbound.AnalyzeWithCancellation(nil, canceled)
	})
	assertOperationCanceled(t, func() {
		unbound.GetDiagnosticsWithCancellation(canceled)
	})
}

func TestBinderCancellationDuringNodeBinding(t *testing.T) {
	text := cancellationStressQuery("", 80)
	query := ParseQuery(text)
	canceled := NewCancellationTokenFunc(func() bool {
		return syntaxHasBoundDescendant(query)
	})
	assertOperationCanceled(t, func() {
		NewBinder(cancellationStressGlobals()).TryBindWithCancellation(query, canceled)
	})
	if !syntaxHasBoundDescendant(query) {
		t.Fatal("binding canceled before any descendant was bound")
	}

	retry := ParseAndAnalyze(text, cancellationStressGlobals())
	if retry == nil || !retry.HasSemantics() || !diagnosticCodesContain(retry.GetDiagnostics(), "KS006") {
		t.Fatal("fresh uncanceled bind did not complete")
	}
}

func TestContainedDiagnosticsCancellationDuringTraversal(t *testing.T) {
	code := ParseAndAnalyze(cancellationStressQuery("", 120), cancellationStressGlobals())
	var checks atomic.Int32
	canceled := NewCancellationTokenFunc(func() bool { return checks.Add(1) >= 40 })
	assertOperationCanceled(t, func() {
		GetContainedDiagnosticsWithCancellation(code.Syntax(), DiagnosticsIncludeDefault, canceled)
	})
	if checks.Load() < 40 {
		t.Fatalf("diagnostic traversal checks = %d, want at least 40", checks.Load())
	}
	if diagnostics := GetContainedDiagnostics(code.Syntax(), DiagnosticsIncludeDefault); len(diagnostics) == 0 || !diagnosticCodesContain(diagnostics, "KS006") {
		t.Fatalf("uncanceled diagnostic retry = %#v", diagnostics)
	}
}

func TestKustoCodeServiceCanceledBindingCanRetryMalformedQuery(t *testing.T) {
	text := cancellationStressQuery("// 😀\n", 160)
	service := NewKustoCodeService(text, cancellationStressGlobals())
	var checks atomic.Int32
	canceled := NewCancellationTokenFunc(func() bool { return checks.Add(1) >= 32 })

	if diagnostics := service.GetDiagnostics(true, canceled); len(diagnostics) != 0 {
		t.Fatalf("canceled diagnostics = %#v, want empty", diagnostics)
	}
	if checks.Load() < 32 {
		t.Fatalf("cancellation checks = %d, want at least 32", checks.Load())
	}
	if service.boundCodeState.Load() != nil || service.codeFailureState.Load() != nil || service.diagnosticsState.Load() != nil {
		t.Fatalf("canceled binding published state: bound=%#v failure=%#v diagnostics=%#v",
			service.boundCodeState.Load(), service.codeFailureState.Load(), service.diagnosticsState.Load())
	}

	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) == 0 || !diagnosticCodesContain(diagnostics, "KS006") {
		t.Fatalf("retried malformed diagnostics = %#v", diagnostics)
	}
	if bound := service.boundCodeState.Load(); bound == nil || bound.code == nil || !bound.code.HasSemantics() {
		t.Fatalf("retried bound state = %#v", bound)
	}
	if service.codeFailureState.Load() != nil || service.diagnosticsState.Load() == nil {
		t.Fatalf("retried publication: failure=%#v diagnostics=%#v", service.codeFailureState.Load(), service.diagnosticsState.Load())
	}

	missing := firstDiagnosticWithCode(diagnostics, "KS006")
	if missing == nil || missing.Start != utf16TextLength(text) || missing.Length != 0 {
		t.Fatalf("UTF-16 missing-expression diagnostic = %#v, want start %d length 0", missing, utf16TextLength(text))
	}
}

func TestKustoCodeServiceCanceledDiagnosticTraversalCanRetry(t *testing.T) {
	text := cancellationStressQuery("", 160)
	service := NewKustoCodeService(text, cancellationStressGlobals())
	if code, ok := service.tryGetBoundCode(CancellationToken{}, true); !ok || code == nil {
		t.Fatal("failed to initialize bound code")
	}
	var checks atomic.Int32
	canceled := NewCancellationTokenFunc(func() bool { return checks.Add(1) >= 24 })

	if diagnostics := service.GetDiagnostics(true, canceled); len(diagnostics) != 0 {
		t.Fatalf("canceled diagnostic traversal = %#v, want empty", diagnostics)
	}
	if checks.Load() < 24 {
		t.Fatalf("diagnostic cancellation checks = %d, want at least 24", checks.Load())
	}
	if service.diagnosticsState.Load() != nil || service.codeFailureState.Load() != nil {
		t.Fatalf("canceled traversal published state: diagnostics=%#v failure=%#v", service.diagnosticsState.Load(), service.codeFailureState.Load())
	}

	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) == 0 || !diagnosticCodesContain(diagnostics, "KS006") || service.diagnosticsState.Load() == nil {
		t.Fatalf("retried diagnostic traversal = %#v, state=%#v", diagnostics, service.diagnosticsState.Load())
	}
}

func TestKustoCodeServiceCanceledInitializerDoesNotBlockLiveCallers(t *testing.T) {
	text := cancellationStressQuery("", 200)
	service := NewKustoCodeService(text, cancellationStressGlobals())
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	canceled := NewCancellationTokenFunc(func() bool {
		once.Do(func() {
			close(started)
			<-release
		})
		return true
	})

	canceledResult := make(chan []Diagnostic, 1)
	go func() {
		canceledResult <- service.GetDiagnostics(true, canceled)
	}()
	<-started

	const callers = 24
	results := make(chan []Diagnostic, callers)
	for i := 0; i < callers; i++ {
		go func() {
			results <- service.GetDiagnostics(true, CancellationToken{})
		}()
	}
	close(release)

	if diagnostics := <-canceledResult; len(diagnostics) != 0 {
		t.Fatalf("canceled initializer diagnostics = %#v, want empty", diagnostics)
	}
	var first []Diagnostic
	for i := 0; i < callers; i++ {
		diagnostics := <-results
		if len(diagnostics) == 0 || !diagnosticCodesContain(diagnostics, "KS006") {
			t.Fatalf("live caller %d diagnostics = %#v", i, diagnostics)
		}
		if first == nil {
			first = diagnostics
		} else if !reflect.DeepEqual(diagnostics, first) {
			t.Fatalf("live caller %d diagnostics differ: %#v vs %#v", i, diagnostics, first)
		}
	}
	if bound := service.boundCodeState.Load(); bound == nil || bound.code == nil || !bound.code.HasSemantics() {
		t.Fatalf("live callers did not publish bound code: %#v", bound)
	}
	if service.codeFailureState.Load() != nil || service.diagnosticsState.Load() == nil {
		t.Fatalf("live publication: failure=%#v diagnostics=%#v", service.codeFailureState.Load(), service.diagnosticsState.Load())
	}
}

func cancellationStressQuery(prefix string, columns int) string {
	var text strings.Builder
	text.WriteString(prefix)
	text.WriteString("T | extend ")
	for i := 0; i < columns; i++ {
		if i > 0 {
			text.WriteString(", ")
		}
		fmt.Fprintf(&text, "C%d=X+%d", i, i)
	}
	text.WriteString(" | where (")
	return text.String()
}

func cancellationStressGlobals() *GlobalState {
	table := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table))
}

func diagnosticCodesContain(diagnostics []Diagnostic, code string) bool {
	return firstDiagnosticWithCode(diagnostics, code) != nil
}

func firstDiagnosticWithCode(diagnostics []Diagnostic, code string) *Diagnostic {
	for i := range diagnostics {
		if diagnostics[i].Code == code {
			return &diagnostics[i]
		}
	}
	return nil
}

func syntaxHasBoundDescendant(root SyntaxElement) bool {
	if root == nil {
		return false
	}
	for i := 0; i < root.ChildCount(); i++ {
		child := root.GetChild(i)
		if node := syntaxNodeForBinderLocation(child); node != nil && node.IsBound() {
			return true
		}
		if syntaxHasBoundDescendant(child) {
			return true
		}
	}
	return false
}

func assertOperationCanceled(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("panic = %#v, want OperationCanceledError", recovered)
		}
		var canceled OperationCanceledError
		if !errors.As(err, &canceled) {
			t.Fatalf("panic = %v, want OperationCanceledError", err)
		}
	}()
	action()
	t.Fatal("operation did not cancel")
}
