package kustolang

import "testing"

// Ported from KustoCompleter.GetCompletionHint, GetChildHint and
// TryGetCompletionContext in src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TestKustoCompleterCompletionHintsAcrossMissingSyntax(t *testing.T) {
	tests := []struct {
		name string
		text string
		want CompletionHint
	}{
		{name: "missing print expression", text: "print x=", want: CompletionHintScalar},
		{name: "missing filter predicate", text: "T | where ", want: CompletionHintBoolean},
		{name: "missing project expression", text: "T | project x=", want: CompletionHintScalar},
		{name: "missing function number argument", text: "print x=abs(", want: CompletionHintNumber},
		{name: "missing pipe operator", text: "T | ", want: CompletionHintQuery},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.text, DefaultGlobalState)
			if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) == 0 {
				t.Fatalf("malformed query %q produced no syntax diagnostics", test.text)
			}
			completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
			if hint := completer.getCompletionHint(len(test.text)); hint&test.want == 0 {
				t.Fatalf("completion hint for %q = %v, want %v: %#v", test.text, hint, test.want, code.GetSyntaxDiagnostics())
			}
		})
	}
}

func TestKustoCompleterCompletionContextRejectsAffinityAfterMissingToken(t *testing.T) {
	text := "T | where == true"
	code := ParseAndAnalyze(text, DefaultGlobalState)
	if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatalf("malformed query %q produced no syntax diagnostics", text)
	}

	position := len("T | where ==")
	completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
	if _, ok := completer.tryGetCompletionContext(position); ok {
		t.Fatalf("completion context unexpectedly accepted affinity after a missing token at %d", position)
	}
}
