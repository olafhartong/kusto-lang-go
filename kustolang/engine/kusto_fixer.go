package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoFixer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type KustoFixer interface {
	Name() string
	Diagnostics() []Diagnostic
	GetFixActions(code *KustoCode, cursorDiagnostic Diagnostic, selectionDiagnostics []Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction
	ApplyFixAction(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
}

type KustoFixerCallbacks struct {
	Name        string
	Diagnostics func() []Diagnostic
	FixAction   func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction
	FixEdits    func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits
}

type KustoFixerBase struct {
	callbacks        KustoFixerCallbacks
	diagnostics      []Diagnostic
	diagnosticsReady bool
}

func NewKustoFixer(callbacks KustoFixerCallbacks) *KustoFixerBase {
	if callbacks.Name == "" {
		callbacks.Name = "KustoFixer"
	}
	return &KustoFixerBase{callbacks: callbacks}
}

func (f *KustoFixerBase) Name() string {
	if f == nil || f.callbacks.Name == "" {
		return "KustoFixer"
	}
	return f.callbacks.Name
}

func (f *KustoFixerBase) Diagnostics() []Diagnostic {
	if f == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	if !f.diagnosticsReady {
		if f.callbacks.Diagnostics != nil {
			f.diagnostics = f.callbacks.Diagnostics()
		}
		f.diagnosticsReady = true
	}
	return ListToReadOnly(f.diagnostics)
}

func (f *KustoFixerBase) GetFixActions(code *KustoCode, cursorDiagnostic Diagnostic, selectionDiagnostics []Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	actions = ListToReadOnly(actions)
	originalCount := len(actions)
	actions = f.getFixAction(code, cursorDiagnostic, options, actions, cancellationToken)

	if len(actions) == originalCount+1 && len(selectionDiagnostics) > 1 {
		primaryAction := actions[originalCount]
		primaryKind := primaryAction.Base().Kind
		relatedActions := []ApplyAction{}
		tmpActions := []CodeAction{}
		for _, diagnostic := range selectionDiagnostics {
			tmpActions = tmpActions[:0]
			tmpActions = f.getFixAction(code, diagnostic, options, tmpActions, cancellationToken)
			for _, action := range tmpActions {
				if apply, ok := action.(ApplyAction); ok && action.Base().Kind == primaryKind {
					relatedActions = append(relatedActions, apply)
				}
			}
		}
		if len(relatedActions) > 1 {
			menuActions := []CodeAction{
				primaryAction.WithTitle("Apply"),
				NewCodeActionFixAll("Fix All", "Apply to all occurences in query or selection.", relatedActions),
			}
			actions[originalCount] = NewCodeActionMenu(primaryAction.Base().Title, primaryAction.Base().Description, menuActions)
		}
	}

	return ListToReadOnly(actions)
}

func (f *KustoFixerBase) getFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	if f != nil && f.callbacks.FixAction != nil {
		return f.callbacks.FixAction(code, diagnostic, options, actions, cancellationToken)
	}
	return actions
}

func (f *KustoFixerBase) ApplyFixAction(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	if multi, ok := action.(*MultiAction); ok && multi != nil {
		return f.applyFixActions(code, multi.Actions, caretPosition, options, cancellationToken)
	}
	return f.applyFixActions(code, []ApplyAction{action}, caretPosition, options, cancellationToken)
}

func (f *KustoFixerBase) applyFixActions(code *KustoCode, actions []ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	suggestedCaretPosition := caretPosition
	suggestedCaretBias := PositionBiasLeft
	edits := []TextEdit{}

	for _, action := range actions {
		result := f.getFixEdits(code, action, caretPosition, options, cancellationToken)
		edits = append(edits, result.Edits...)
		if editContainsPosition(result.Edits, caretPosition) {
			suggestedCaretPosition = result.NewCaretPosition
			suggestedCaretBias = result.Bias
		}
	}

	text := ""
	if code != nil {
		text = code.Text
	}
	startingText := NewEditString(text)
	if len(edits) > 0 {
		if !startingText.CanApplyAll(edits) {
			return NewCodeActionFailure("Invalid fix changes")
		}
		newText := startingText.ApplyAll(edits)
		newCaretPosition := newText.GetCurrentPosition(suggestedCaretPosition, suggestedCaretBias)
		return NewCodeActionChangeAndMoveEditString(newText, newCaretPosition)
	}

	return CodeActionResultNothing
}

func (f *KustoFixerBase) getFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	if f != nil && f.callbacks.FixEdits != nil {
		return f.callbacks.FixEdits(code, action, caretPosition, options, cancellationToken)
	}
	return NewFixEdits(caretPosition, nil)
}

func editContainsPosition(edits []TextEdit, position int) bool {
	if len(edits) == 0 {
		return false
	}
	minStart := edits[0].Start
	maxEnd := edits[0].Start + edits[0].DeleteLength
	for _, edit := range edits[1:] {
		if edit.Start < minStart {
			minStart = edit.Start
		}
		if end := edit.Start + edit.DeleteLength; end > maxEnd {
			maxEnd = end
		}
	}
	return position >= minStart && position < maxEnd
}

type FixEdits struct {
	Edits            []TextEdit
	NewCaretPosition int
	Bias             PositionBias
}

func NewFixEdits(newCaretPosition int, edits []TextEdit) FixEdits {
	return NewFixEditsWithBias(newCaretPosition, PositionBiasLeft, edits)
}

func NewFixEditsWithBias(newCaretPosition int, bias PositionBias, edits []TextEdit) FixEdits {
	return FixEdits{
		Edits:            ListToReadOnly(edits),
		NewCaretPosition: newCaretPosition,
		Bias:             bias,
	}
}

func NewFixEditsList(newCaretPosition int, edits ...TextEdit) FixEdits {
	return NewFixEdits(newCaretPosition, edits)
}

func NewFixEditsWithBiasList(newCaretPosition int, bias PositionBias, edits ...TextEdit) FixEdits {
	return NewFixEditsWithBias(newCaretPosition, bias, edits)
}
