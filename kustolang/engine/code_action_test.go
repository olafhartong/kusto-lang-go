package kustolang

import "testing"

func TestCodeActionFactoriesAndCloneHelpers(t *testing.T) {
	action := NewCodeAction("kind", "title", "description", []string{"a"})
	if action.Kind != "kind" || action.Title != "title" || action.Description != "description" {
		t.Fatalf("single action base = %#v", action.CodeActionBase)
	}
	if len(action.Data) != 1 || action.Data[0] != "a" {
		t.Fatalf("single action data = %#v", action.Data)
	}

	titleOnly := NewCodeActionTitle("rename", "description", nil)
	if titleOnly.Kind != "rename" || titleOnly.Title != "rename" || len(titleOnly.Data) != 0 {
		t.Fatalf("title action = %#v data:%#v", titleOnly.CodeActionBase, titleOnly.Data)
	}

	withTitle, ok := action.WithTitle("new title").(*SingleAction)
	if !ok || withTitle.Kind != "kind" || withTitle.Title != "new title" || withTitle.Description != "description" {
		t.Fatalf("WithTitle result = %#v", withTitle)
	}
	withDescription, ok := action.WithDescription("new description").(*SingleAction)
	if !ok || withDescription.Title != "title" || withDescription.Description != "new description" {
		t.Fatalf("WithDescription result = %#v", withDescription)
	}
	if action.Title != "title" || action.Description != "description" {
		t.Fatalf("source action mutated = %#v", action.CodeActionBase)
	}
}

func TestApplyActionDataHelpersPreserveActionShape(t *testing.T) {
	action := NewCodeAction("kind", "title", "description", []string{"a", "b"})
	if same := action.WithData([]string{"a", "b"}); same != action {
		t.Fatalf("WithData with equal values should preserve identity")
	}

	added, ok := action.AddDataStrings("c", "d").(*SingleAction)
	if !ok || !readOnlyStringListEquals(added.Data, []string{"a", "b", "c", "d"}) {
		t.Fatalf("AddData result = %#v", added)
	}
	removed, ok := added.RemoveData(3).(*SingleAction)
	if !ok || !readOnlyStringListEquals(removed.Data, []string{"a"}) {
		t.Fatalf("RemoveData result = %#v", removed)
	}
	removedNone, ok := removed.RemoveData(-1).(*SingleAction)
	if !ok || !readOnlyStringListEquals(removedNone.Data, []string{"a"}) {
		t.Fatalf("negative RemoveData result = %#v", removedNone)
	}
	removedAll, ok := removed.RemoveData(10).(*SingleAction)
	if !ok || len(removedAll.Data) != 0 {
		t.Fatalf("large RemoveData result = %#v", removedAll)
	}
}

func TestMultiAndMenuActionsCopyNestedActions(t *testing.T) {
	first := NewCodeAction("one", "One", "", nil)
	second := NewCodeAction("two", "Two", "", nil)
	actions := []ApplyAction{first}
	fixAll := NewCodeActionFixAll("Fix all", "description", actions)
	actions[0] = second
	if fixAll.Kind != CodeActionFixAllKind || len(fixAll.Actions) != 1 || fixAll.Actions[0] != first {
		t.Fatalf("fix-all action = %#v", fixAll)
	}

	withData, ok := fixAll.WithDataStrings("batch").(*MultiAction)
	if !ok || len(withData.Actions) != 1 || withData.Actions[0] != first || len(withData.Data) != 1 || withData.Data[0] != "batch" {
		t.Fatalf("multi WithData result = %#v", withData)
	}

	menuActions := []CodeAction{first}
	menu := NewCodeActionMenu("Menu", "description", menuActions)
	menuActions[0] = second
	if menu.Kind != CodeActionMenuKind || len(menu.Actions) != 1 || menu.Actions[0] != first {
		t.Fatalf("menu action = %#v", menu)
	}
	replaced := menu.WithActionsList(second)
	if len(replaced.Actions) != 1 || replaced.Actions[0] != second || menu.Actions[0] != first {
		t.Fatalf("menu WithActions result = replaced:%#v original:%#v", replaced, menu)
	}
}

func TestCodeActionInfoAndOptionsDefaults(t *testing.T) {
	action := NewCodeAction("kind", "title", "", nil)
	actions := []CodeAction{action}
	info := NewCodeActionInfo(actions)
	actions[0] = NewCodeAction("other", "other", "", nil)
	if len(info.Actions) != 1 || info.Actions[0] != action {
		t.Fatalf("code action info should copy actions, got %#v", info.Actions)
	}
	if CodeActionInfoNoActions.Actions == nil || len(CodeActionInfoNoActions.Actions) != 0 {
		t.Fatalf("no-actions info = %#v", CodeActionInfoNoActions)
	}

	options := CodeActionOptionsDefault
	if options.DiagnosticFilter.String() != "" || options.FormattingOptions != FormattingOptionsDefault {
		t.Fatalf("code action option defaults = %#v", options)
	}
	filter := ParseDisabledDiagnostics("sev:Warning")
	formatting := FormattingOptionsDefault.WithIndentationSize(2)
	changed := options.WithDiagnosticFilter(&filter).WithFormattingOptions(&formatting)
	if changed.DiagnosticFilter.String() != "sev:Warning" || changed.FormattingOptions.IndentationSize != 2 {
		t.Fatalf("changed code action options = %#v", changed)
	}
	reset := changed.WithDiagnosticFilter(nil).WithFormattingOptions(nil)
	if reset.DiagnosticFilter.String() != "" || reset.FormattingOptions != FormattingOptionsDefault {
		t.Fatalf("reset code action options = %#v", reset)
	}
}

func TestCodeActionResultConstructorsAndPositionAdjustment(t *testing.T) {
	editString := NewEditString("abc").ReplaceAt(1, 1, "XYZ")
	change := NewCodeActionChangeEditString(editString)
	if len(change.Actions) != 1 || change.FailureReason != "" {
		t.Fatalf("change result = %#v", change)
	}
	changeAction, ok := change.Actions[0].(ChangeTextAction)
	if !ok || changeAction.ChangedText != "aXYZc" || len(changeAction.Changes) != 1 || changeAction.Changes[0].InsertText != "XYZ" {
		t.Fatalf("change action = %#v", change.Actions[0])
	}

	changeAndMove := NewCodeActionChangeAndMove([]TextEdit{NewTextInsertion(1, "x")}, "axbc", 2)
	adjusted := changeAndMove.WithAdjustedPosition(5)
	if len(adjusted.Actions) != 2 {
		t.Fatalf("adjusted action count = %d", len(adjusted.Actions))
	}
	if move, ok := adjusted.Actions[1].(MoveCaretAction); !ok || move.NewCaretPosition != 7 {
		t.Fatalf("adjusted move action = %#v", adjusted.Actions[1])
	}
	if originalMove, ok := changeAndMove.Actions[1].(MoveCaretAction); !ok || originalMove.NewCaretPosition != 2 {
		t.Fatalf("source move action mutated = %#v", changeAndMove.Actions[1])
	}

	withRename := adjusted.WithAction(RenameActionInstance)
	if len(withRename.Actions) != 3 {
		t.Fatalf("WithAction action count = %d", len(withRename.Actions))
	}
	if len(CodeActionResultNothing.Actions) != 0 || CodeActionResultNothing.FailureReason != "" || CodeActionResultNothing.HasFailureReason() {
		t.Fatalf("nothing result = %#v", CodeActionResultNothing)
	}
}

func TestCodeActionResultFailure(t *testing.T) {
	failed := NewCodeActionFailure()
	if failed.FailureReason != "Action failed" || !failed.HasFailureReason() || len(failed.Actions) != 0 {
		t.Fatalf("default failure = %#v", failed)
	}
	custom := NewCodeActionFailure("custom")
	if custom.FailureReason != "custom" || !custom.HasFailureReason() || len(custom.Actions) != 0 {
		t.Fatalf("custom failure = %#v", custom)
	}
	empty := NewCodeActionFailure("")
	if empty.FailureReason != "" || !empty.HasFailureReason() || len(empty.Actions) != 0 {
		t.Fatalf("empty failure = %#v", empty)
	}
}
