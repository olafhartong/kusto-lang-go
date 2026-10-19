package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type KustoAnalyzer interface {
	Name() string
	Diagnostics() []Diagnostic
	Analyze(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic
	Fixer() KustoFixer
}

type KustoAnalyzerCallbacks struct {
	Name        string
	Diagnostics func() []Diagnostic
	Analyze     func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic
	FixAction   func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction
	FixEdits    func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits
}

type KustoAnalyzerBase struct {
	callbacks        KustoAnalyzerCallbacks
	diagnostics      []Diagnostic
	diagnosticsReady bool
	fixer            KustoFixer
}

func NewKustoAnalyzer(callbacks KustoAnalyzerCallbacks) *KustoAnalyzerBase {
	if callbacks.Name == "" {
		callbacks.Name = "KustoAnalyzer"
	}
	return &KustoAnalyzerBase{callbacks: callbacks}
}

func (a *KustoAnalyzerBase) Name() string {
	if a == nil || a.callbacks.Name == "" {
		return "KustoAnalyzer"
	}
	return a.callbacks.Name
}

func (a *KustoAnalyzerBase) Diagnostics() []Diagnostic {
	if a == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	if !a.diagnosticsReady {
		if a.callbacks.Diagnostics != nil {
			a.diagnostics = a.callbacks.Diagnostics()
		}
		a.diagnosticsReady = true
	}
	return ListToReadOnly(a.diagnostics)
}

func (a *KustoAnalyzerBase) Analyze(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if a != nil && a.callbacks.Analyze != nil {
		return a.callbacks.Analyze(code, diagnostics, cancellationToken)
	}
	return diagnostics
}

func (a *KustoAnalyzerBase) Fixer() KustoFixer {
	if a == nil {
		return NewKustoFixer(KustoFixerCallbacks{Name: "KustoAnalyzer"})
	}
	if a.fixer == nil {
		a.fixer = a.createFixer()
	}
	return a.fixer
}

func (a *KustoAnalyzerBase) createFixer() KustoFixer {
	return NewKustoFixer(KustoFixerCallbacks{
		Name:        a.Name(),
		Diagnostics: a.Diagnostics,
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			if a.callbacks.FixAction != nil {
				return a.callbacks.FixAction(code, diagnostic, options, actions, cancellationToken)
			}
			return actions
		},
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			if a.callbacks.FixEdits != nil {
				return a.callbacks.FixEdits(code, action, caretPosition, options, cancellationToken)
			}
			return NewFixEdits(caretPosition, nil)
		},
	})
}
