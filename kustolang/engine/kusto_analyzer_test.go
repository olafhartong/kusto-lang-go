package kustolang

import "testing"

func TestKustoAnalyzerDiagnosticsAnalyzeAndFixerDelegation(t *testing.T) {
	diagnosticCalls := 0
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "ExampleAnalyzer",
		Diagnostics: func() []Diagnostic {
			diagnosticCalls++
			return []Diagnostic{NewDiagnostic("KS777", "example")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			return append(diagnostics, NewDiagnostic("KS778", code.Text))
		},
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			return append(actions, NewCodeAction("fix", "Fix "+diagnostic.Code, "", []string{diagnostic.Code}))
		},
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			return NewFixEditsList(caretPosition, NewTextReplacement(0, 1, "X"))
		},
	})

	if analyzer.Name() != "ExampleAnalyzer" {
		t.Fatalf("name = %q", analyzer.Name())
	}
	diagnostics := analyzer.Diagnostics()
	diagnostics[0] = NewDiagnostic("KS000", "mutated")
	if got := analyzer.Diagnostics(); diagnosticCalls != 1 || got[0].Code != "KS777" {
		t.Fatalf("diagnostics = calls:%d %#v", diagnosticCalls, got)
	}

	analyzed := analyzer.Analyze(Parse("abc"), nil, CancellationToken{})
	if len(analyzed) != 1 || analyzed[0].Code != "KS778" || analyzed[0].Message != "abc" {
		t.Fatalf("analyzed diagnostics = %#v", analyzed)
	}

	fixer := analyzer.Fixer()
	if fixer != analyzer.Fixer() || fixer.Name() != "ExampleAnalyzer" {
		t.Fatalf("fixer identity/name = %#v %q", fixer, fixer.Name())
	}
	if got := fixer.Diagnostics(); len(got) != 1 || got[0].Code != "KS777" || diagnosticCalls != 1 {
		t.Fatalf("fixer diagnostics = calls:%d %#v", diagnosticCalls, got)
	}
	actions := fixer.GetFixActions(Parse("abc"), NewDiagnostic("KS777", "example"), nil, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 || actions[0].Base().Title != "Fix KS777" {
		t.Fatalf("fixer actions = %#v", actions)
	}
	result := fixer.ApplyFixAction(Parse("abc"), actions[0].(ApplyAction), 0, CodeActionOptionsDefault, CancellationToken{})
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "Xbc" {
		t.Fatalf("fixer result = %#v", result)
	}
}

func TestKustoAnalyzerDefaultsAreEmptyAndNoop(t *testing.T) {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{})
	if analyzer.Name() != "KustoAnalyzer" {
		t.Fatalf("default name = %q", analyzer.Name())
	}
	if diagnostics := analyzer.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("default diagnostics = %#v", diagnostics)
	}
	existing := []Diagnostic{NewDiagnostic("KS001", "existing")}
	if got := analyzer.Analyze(Parse("T | count"), existing, CancellationToken{}); len(got) != 1 || got[0].Code != "KS001" {
		t.Fatalf("default analyze = %#v", got)
	}
	if result := analyzer.Fixer().ApplyFixAction(Parse("abc"), NewCodeAction("noop", "Noop", "", nil), 0, CodeActionOptionsDefault, CancellationToken{}); len(result.Actions) != 0 || result.FailureReason != "" {
		t.Fatalf("default fixer result = %#v", result)
	}
}
