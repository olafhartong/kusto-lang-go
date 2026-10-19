package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type externalCodeAction struct {
	kustolang.CodeActionBase
}

func (a externalCodeAction) Base() kustolang.CodeActionBase { return a.CodeActionBase }

func (a externalCodeAction) WithTitle(title string) kustolang.CodeAction {
	a.Title = title
	return a
}

func (a externalCodeAction) WithDescription(description string) kustolang.CodeAction {
	a.Description = description
	return a
}

type externalApplyAction struct {
	kustolang.ApplyActionBase
	data []string
}

func (a externalApplyAction) Base() kustolang.CodeActionBase { return a.CodeActionBase }

func (a externalApplyAction) WithTitle(title string) kustolang.CodeAction {
	a.Title = title
	return a
}

func (a externalApplyAction) WithDescription(description string) kustolang.CodeAction {
	a.Description = description
	return a
}

func (a externalApplyAction) ApplyData() []string { return append([]string(nil), a.data...) }

func (a externalApplyAction) WithData(data []string) kustolang.ApplyAction {
	a.data = append([]string(nil), data...)
	return a
}

func (a externalApplyAction) WithDataStrings(data ...string) kustolang.ApplyAction {
	return a.WithData(data)
}

func (a externalApplyAction) AddData(data []string) kustolang.ApplyAction {
	return a.WithData(append(a.ApplyData(), data...))
}

func (a externalApplyAction) AddDataStrings(data ...string) kustolang.ApplyAction {
	return a.AddData(data)
}

func (a externalApplyAction) RemoveData(count int) kustolang.ApplyAction {
	keep := len(a.data) - count
	if keep < 0 {
		keep = 0
	}
	return a.WithData(a.data[:keep])
}

type externalResultAction struct {
	kustolang.ResultActionBase
	value string
}

func TestPublicCodeActionAuthoring(t *testing.T) {
	data := []string{"payload"}
	action := kustolang.NewCodeAction("replace", "Replace", "Replace text", data)
	data[0] = "mutated"
	if got := action.ApplyData(); !reflect.DeepEqual(got, []string{"payload"}) {
		t.Fatalf("action data = %#v", got)
	}
	returned := action.ApplyData()
	returned[0] = "also mutated"
	if got := action.ApplyData(); !reflect.DeepEqual(got, []string{"payload"}) {
		t.Fatalf("action data after returned-slice mutation = %#v", got)
	}

	titleAction := kustolang.NewCodeActionTitle("Rename", "Rename symbol", nil)
	if base := titleAction.Base(); base.Kind != "Rename" || base.Title != "Rename" || base.Description != "Rename symbol" {
		t.Fatalf("title action = %#v", base)
	}

	applyActions := []kustolang.ApplyAction{action}
	fixAll := kustolang.NewCodeActionFixAll("Fix all", "Apply every fix", applyActions)
	applyActions[0] = titleAction
	if fixAll.Kind != kustolang.CodeActionFixAllKind || len(fixAll.Actions) != 1 || fixAll.Actions[0] != action {
		t.Fatalf("fix-all action = %#v", fixAll)
	}

	menuActions := []kustolang.CodeAction{action}
	menu := kustolang.NewCodeActionMenu("Choose", "Choose a fix", menuActions)
	menuActions[0] = titleAction
	if menu.Kind != kustolang.CodeActionMenuKind || len(menu.Actions) != 1 || menu.Actions[0] != action {
		t.Fatalf("menu action = %#v", menu)
	}

	infoActions := []kustolang.CodeAction{menu}
	info := kustolang.NewCodeActionInfo(infoActions)
	infoActions[0] = action
	if len(info.Actions) != 1 || info.Actions[0] != menu {
		t.Fatalf("code action info = %#v", info)
	}
	if kustolang.CodeActionInfoNoActions.Actions == nil || len(kustolang.CodeActionInfoNoActions.Actions) != 0 {
		t.Fatalf("no-actions info = %#v", kustolang.CodeActionInfoNoActions)
	}
}

func TestExternalCodeActionImplementations(t *testing.T) {
	var code kustolang.CodeAction = externalCodeAction{CodeActionBase: kustolang.CodeActionBase{Kind: "external", Title: "External"}}
	if got := code.WithDescription("custom").Base(); got.Kind != "external" || got.Description != "custom" {
		t.Fatalf("external code action = %#v", got)
	}

	var apply kustolang.ApplyAction = externalApplyAction{
		ApplyActionBase: kustolang.ApplyActionBase{CodeActionBase: kustolang.CodeActionBase{Kind: "apply", Title: "Apply"}},
		data:            []string{"one"},
	}
	if got := apply.AddDataStrings("two").ApplyData(); !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatalf("external apply action data = %#v", got)
	}

	result := kustolang.NewCodeActionResult(externalResultAction{value: "done"})
	if got, ok := result.Actions[0].(externalResultAction); !ok || got.value != "done" {
		t.Fatalf("external result action = %#v", result.Actions)
	}
}

func TestPublicCodeActionResultAuthoring(t *testing.T) {
	edits := []kustolang.TextEdit{kustolang.NewTextReplacement(1, 2, "x")}
	result := kustolang.NewCodeActionChangeAndMove(edits, "axd", 2)
	edits[0] = kustolang.NewTextDeletion(0, 1)
	if len(result.Actions) != 2 {
		t.Fatalf("result actions = %#v", result.Actions)
	}
	change, ok := result.Actions[0].(kustolang.ChangeTextAction)
	if !ok || !reflect.DeepEqual(change.Changes, []kustolang.TextEdit{kustolang.NewTextReplacement(1, 2, "x")}) || change.ChangedText != "axd" {
		t.Fatalf("change action = %#v", result.Actions[0])
	}
	move, ok := result.Actions[1].(kustolang.MoveCaretAction)
	if !ok || move.NewCaretPosition != 2 {
		t.Fatalf("move action = %#v", result.Actions[1])
	}

	adjusted := result.WithAdjustedPosition(5).WithAction(kustolang.RenameActionInstance)
	adjustedMove, ok := adjusted.Actions[1].(kustolang.MoveCaretAction)
	if !ok || adjustedMove.NewCaretPosition != 7 || len(adjusted.Actions) != 3 {
		t.Fatalf("adjusted result = %#v", adjusted)
	}
	if _, ok := adjusted.Actions[2].(kustolang.RenameAction); !ok {
		t.Fatalf("rename action = %#v", adjusted.Actions[2])
	}
	if originalMove := result.Actions[1].(kustolang.MoveCaretAction); originalMove.NewCaretPosition != 2 {
		t.Fatalf("original move changed = %#v", originalMove)
	}

	if failed := kustolang.NewCodeActionFailure(); failed.FailureReason != "Action failed" || !failed.HasFailureReason() || len(failed.Actions) != 0 {
		t.Fatalf("default failure = %#v", failed)
	}
	if empty := kustolang.NewCodeActionFailure(""); empty.FailureReason != "" || !empty.HasFailureReason() || len(empty.Actions) != 0 {
		t.Fatalf("explicit empty failure = %#v", empty)
	}
	if len(kustolang.CodeActionResultNothing.Actions) != 0 || kustolang.CodeActionResultNothing.FailureReason != "" || kustolang.CodeActionResultNothing.HasFailureReason() {
		t.Fatalf("nothing result = %#v", kustolang.CodeActionResultNothing)
	}
}

func TestPublicCodeActionEditStringOverloadsUseUTF16Locations(t *testing.T) {
	changed := kustolang.NewEditString("😀x").ReplaceAt(2, 1, "y")
	result := kustolang.NewCodeActionChangeAndMoveEditString(changed, 3)
	change, ok := result.Actions[0].(kustolang.ChangeTextAction)
	if !ok || !reflect.DeepEqual(change.Changes, []kustolang.TextEdit{kustolang.NewTextReplacement(2, 1, "y")}) || change.ChangedText != "😀y" {
		t.Fatalf("UTF-16 change action = %#v", result.Actions[0])
	}
	move, ok := result.Actions[1].(kustolang.MoveCaretAction)
	if !ok || move.NewCaretPosition != 3 {
		t.Fatalf("UTF-16 move action = %#v", result.Actions[1])
	}
}
