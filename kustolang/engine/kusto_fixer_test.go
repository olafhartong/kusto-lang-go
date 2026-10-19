package kustolang

import "testing"

func TestKustoFixerDiagnosticsAreCachedAndCopied(t *testing.T) {
	calls := 0
	fixer := NewKustoFixer(KustoFixerCallbacks{
		Name: "ExampleFixer",
		Diagnostics: func() []Diagnostic {
			calls++
			return []Diagnostic{NewDiagnostic("KS999", "example")}
		},
	})

	first := fixer.Diagnostics()
	first[0] = NewDiagnostic("KS000", "mutated")
	second := fixer.Diagnostics()
	if calls != 1 || len(second) != 1 || second[0].Code != "KS999" || fixer.Name() != "ExampleFixer" {
		t.Fatalf("diagnostics cache = calls:%d second:%#v name:%q", calls, second, fixer.Name())
	}
}

func TestKustoFixerPromotesSingleActionToFixAllMenu(t *testing.T) {
	fixer := NewKustoFixer(KustoFixerCallbacks{
		Name: "MenuFixer",
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			action := NewCodeAction("replace", "Replace "+diagnostic.Code, "", []string{diagnostic.Code})
			return append(actions, action)
		},
	})

	selection := []Diagnostic{
		NewDiagnostic("KS001", "one"),
		NewDiagnostic("KS002", "two"),
	}
	actions := fixer.GetFixActions(Parse("T | count"), selection[0], selection, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one menu", actions)
	}
	menu, ok := actions[0].(*MenuAction)
	if !ok || menu.Title != "Replace KS001" || len(menu.Actions) != 2 {
		t.Fatalf("menu action = %#v", actions[0])
	}
	if menu.Actions[0].Base().Title != "Apply" {
		t.Fatalf("primary menu action title = %q", menu.Actions[0].Base().Title)
	}
	fixAll, ok := menu.Actions[1].(*MultiAction)
	if !ok || fixAll.Kind != CodeActionFixAllKind || len(fixAll.Actions) != 2 {
		t.Fatalf("fix-all action = %#v", menu.Actions[1])
	}
}

func TestKustoFixerApplyFixActionChangesTextAndMovesCaret(t *testing.T) {
	fixer := NewKustoFixer(KustoFixerCallbacks{
		Name: "EditFixer",
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			return NewFixEditsWithBiasList(1, PositionBiasRight, NewTextReplacement(1, 1, "XX"))
		},
	})

	result := fixer.ApplyFixAction(Parse("abc"), NewCodeAction("replace", "Replace", "", nil), 1, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) != 2 {
		t.Fatalf("result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "aXXc" || len(change.Changes) != 1 || change.Changes[0] != NewTextReplacement(1, 1, "XX") {
		t.Fatalf("change action = %#v", result.Actions[0])
	}
	move, ok := result.Actions[1].(MoveCaretAction)
	if !ok || move.NewCaretPosition != 3 {
		t.Fatalf("move action = %#v", result.Actions[1])
	}
}

func TestKustoFixerApplyMultiActionAggregatesEdits(t *testing.T) {
	fixer := NewKustoFixer(KustoFixerCallbacks{
		Name: "MultiFixer",
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			data := action.ApplyData()
			switch data[0] {
			case "left":
				return NewFixEditsList(caretPosition, NewTextReplacement(0, 1, "A"))
			case "right":
				return NewFixEditsList(caretPosition, NewTextReplacement(2, 1, "C"))
			default:
				return NewFixEdits(caretPosition, nil)
			}
		},
	})
	multi := NewCodeActionFixAll("Fix All", "", []ApplyAction{
		NewCodeAction("replace", "left", "", []string{"left"}),
		NewCodeAction("replace", "right", "", []string{"right"}),
	})

	result := fixer.ApplyFixAction(Parse("abc"), multi, 1, CodeActionOptionsDefault, CancellationToken{})
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "AbC" || len(change.Changes) != 2 {
		t.Fatalf("multi action change = %#v", result)
	}
}

func TestKustoFixerApplyInvalidFixChangesFails(t *testing.T) {
	fixer := NewKustoFixer(KustoFixerCallbacks{
		Name: "BadFixer",
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			return NewFixEditsList(caretPosition,
				NewTextReplacement(0, 2, "x"),
				NewTextReplacement(1, 1, "y"),
			)
		},
	})

	result := fixer.ApplyFixAction(Parse("abc"), NewCodeAction("replace", "bad", "", nil), 0, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "Invalid fix changes" || len(result.Actions) != 0 {
		t.Fatalf("invalid result = %#v", result)
	}
}
