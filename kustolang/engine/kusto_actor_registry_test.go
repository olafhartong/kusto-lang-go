package kustolang

import (
	"strings"
	"testing"
)

type kustoActorOpaqueResultAction struct {
	ResultActionBase
	Data []string
}

type explicitUTF16KustoActor struct {
	*KustoActorBase
	getPosition int
	getStart    int
	getLength   int
	applyCaret  int
	applyData   []string
	result      CodeActionResult
}

func (a *explicitUTF16KustoActor) GetActionsUTF16(_ *KustoCodeService, _ *KustoCode, position, selectionStart, selectionLength int, _ CodeActionOptions, actions []CodeAction, _ bool, _ CancellationToken) []CodeAction {
	a.getPosition = position
	a.getStart = selectionStart
	a.getLength = selectionLength
	return append(actions, NewCodeAction("Explicit", "Explicit UTF-16 action", "", []string{"opaque", "12"}))
}

func (a *explicitUTF16KustoActor) ApplyActionUTF16(_ *KustoCodeService, _ *KustoCode, action ApplyAction, caretPosition int, _ CodeActionOptions, _ CancellationToken) CodeActionResult {
	a.applyCaret = caretPosition
	a.applyData = action.ApplyData()
	return a.result
}

func TestKustoActorRegistryMatchesPublicActorOrder(t *testing.T) {
	registrations := kustoActorRegistrations()
	names := KustoActorNamesAll()
	if len(registrations) != len(names) {
		t.Fatalf("registry count = %d, names = %d", len(registrations), len(names))
	}
	for index, actor := range registrations {
		if actor.Name() != names[index] || actor.getActions == nil || actor.apply == nil {
			t.Fatalf("registry[%d] = %#v, want %q with get/apply", index, actor, names[index])
		}
		if _, ok := any(actor).(kustoActorUTF16); !ok {
			t.Fatalf("registry[%d] = %#v, want explicit UTF-16 dispatch", index, actor)
		}
	}
}

func TestKustoCodeServiceWithActorsFiltersAndAppliesCustomActor(t *testing.T) {
	const actorName = "CustomTestActor"
	text := "T | where (X >"
	code := ParseAndAnalyze(text, convertMacroExpandTestGlobals())
	getCalled := false
	applyCalled := false
	actor := NewKustoActor(KustoActorCallbacks{
		Name: actorName,
		GetActions: func(_ *KustoCodeService, received *KustoCode, _, _, _ int, _ CodeActionOptions, actions []CodeAction, _ bool, _ CancellationToken) []CodeAction {
			getCalled = true
			if received != code || !received.HasSemantics() || len(received.GetSyntaxDiagnostics()) == 0 {
				t.Fatalf("custom actor received code = %#v", received)
			}
			return append(actions, NewCodeAction("Custom", "Custom action", "", []string{"payload"}))
		},
		Apply: func(_ *KustoCodeService, received *KustoCode, action ApplyAction, caret int, _ CodeActionOptions, _ CancellationToken) CodeActionResult {
			applyCalled = true
			if received != code {
				return NewCodeActionFailure("code instance was not preserved")
			}
			data := action.ApplyData()
			if len(data) != 1 || data[0] != "payload" {
				return NewCodeActionFailure("invalid custom payload")
			}
			edit := NewEditString(received.Text).Prepend("// custom\n")
			return NewCodeActionChangeAndMoveEditString(edit, caret+len("// custom\n"))
		},
	})
	service := NewKustoCodeServiceForCode(code).WithActors([]KustoActor{nil, actor})

	info := service.GetCodeActions(len(text), 0, len(text), CodeActionOptionsDefault, true, actorName, CancellationToken{})
	if !getCalled || len(info.Actions) != 1 {
		t.Fatalf("custom actions = %#v, getCalled = %t", info.Actions, getCalled)
	}
	action, ok := info.Actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("custom action type = %T", info.Actions[0])
	}
	data := action.ApplyData()
	if len(data) != 2 || data[0] != "payload" || data[1] != actorName {
		t.Fatalf("service custom action data = %#v", data)
	}

	result := service.ApplyCodeAction(action, len(text), CodeActionOptionsDefault, CancellationToken{})
	if !applyCalled || result.FailureReason != "" || len(result.Actions) != 2 {
		t.Fatalf("custom apply result = %#v, applyCalled = %t", result, applyCalled)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "// custom\n"+text {
		t.Fatalf("custom text change = %#v", result.Actions[0])
	}

	malformed := NewCodeAction("Custom", "", "", []string{"bad", actorName})
	failed := service.ApplyCodeAction(malformed, 0, CodeActionOptionsDefault, CancellationToken{})
	if failed.FailureReason != "invalid custom payload" || len(failed.Actions) != 0 {
		t.Fatalf("malformed custom action result = %#v", failed)
	}
}

func TestKustoActorDispatchAdaptsLegacyActorUTF16Coordinates(t *testing.T) {
	const (
		actorName = "LegacyUTF16DispatchTestActor"
		text      = "// 😀\nprint x=1"
	)
	opaque := &kustoActorOpaqueResultAction{Data: []string{"opaque", "12"}}
	getPosition, getStart, getLength, applyCaret := -1, -1, -1, -1
	var applyData []string
	actor := NewKustoActor(KustoActorCallbacks{
		Name: actorName,
		GetActions: func(_ *KustoCodeService, _ *KustoCode, position, selectionStart, selectionLength int, _ CodeActionOptions, actions []CodeAction, _ bool, _ CancellationToken) []CodeAction {
			getPosition, getStart, getLength = position, selectionStart, selectionLength
			return append(actions, NewCodeAction("Legacy", "Legacy action", "", []string{"opaque", "12"}))
		},
		Apply: func(_ *KustoCodeService, _ *KustoCode, action ApplyAction, caretPosition int, _ CodeActionOptions, _ CancellationToken) CodeActionResult {
			applyCaret = caretPosition
			applyData = action.ApplyData()
			return NewCodeActionResult(
				NewChangeTextAction([]TextEdit{NewTextReplacement(14, 1, "y")}, "// 😀\nprint y=1"),
				NewMoveCaretAction(15),
				opaque,
			)
		},
	})
	service := NewKustoCodeService(text, convertMacroExpandTestGlobals()).WithActors([]KustoActor{actor})

	info := service.GetCodeActions(12, 12, 1, CodeActionOptionsDefault, true, actorName, CancellationToken{})
	if getPosition != 14 || getStart != 14 || getLength != 1 {
		t.Fatalf("legacy GetActions coordinates = (%d, %d, %d), want (14, 14, 1)", getPosition, getStart, getLength)
	}
	if len(info.Actions) != 1 {
		t.Fatalf("legacy actions = %#v", info.Actions)
	}
	action, ok := info.Actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("legacy action type = %T", info.Actions[0])
	}
	if data := action.ApplyData(); len(data) != 3 || data[0] != "opaque" || data[1] != "12" || data[2] != actorName {
		t.Fatalf("legacy action data = %#v", data)
	}

	result := service.ApplyCodeAction(action, 12, CodeActionOptionsDefault, CancellationToken{})
	if applyCaret != 14 {
		t.Fatalf("legacy ApplyAction caret = %d, want byte offset 14", applyCaret)
	}
	if len(applyData) != 2 || applyData[0] != "opaque" || applyData[1] != "12" {
		t.Fatalf("legacy ApplyAction data = %#v", applyData)
	}
	if result.FailureReason != "" || len(result.Actions) != 3 {
		t.Fatalf("legacy result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || len(change.Changes) != 1 || change.Changes[0].Start != 12 || change.Changes[0].DeleteLength != 1 {
		t.Fatalf("legacy UTF-16 change = %#v", result.Actions[0])
	}
	move, ok := result.Actions[1].(MoveCaretAction)
	if !ok || move.NewCaretPosition != 13 {
		t.Fatalf("legacy UTF-16 caret = %#v", result.Actions[1])
	}
	if result.Actions[2] != opaque || len(opaque.Data) != 2 || opaque.Data[1] != "12" {
		t.Fatalf("legacy opaque result action = %#v", result.Actions[2])
	}
}

func TestKustoActorDispatchPrefersExplicitUTF16Actor(t *testing.T) {
	const (
		actorName = "ExplicitUTF16DispatchTestActor"
		text      = "// 😀\nprint x=1"
	)
	legacyGetCalled, legacyApplyCalled := false, false
	actor := &explicitUTF16KustoActor{
		KustoActorBase: NewKustoActor(KustoActorCallbacks{
			Name: actorName,
			GetActions: func(_ *KustoCodeService, _ *KustoCode, _, _, _ int, _ CodeActionOptions, actions []CodeAction, _ bool, _ CancellationToken) []CodeAction {
				legacyGetCalled = true
				return actions
			},
			Apply: func(_ *KustoCodeService, _ *KustoCode, _ ApplyAction, _ int, _ CodeActionOptions, _ CancellationToken) CodeActionResult {
				legacyApplyCalled = true
				return NewCodeActionFailure()
			},
		}),
		result: NewCodeActionChangeAndMove(
			[]TextEdit{NewTextReplacement(12, 1, "y")},
			"// 😀\nprint y=1",
			13,
		),
	}
	service := NewKustoCodeService(text, convertMacroExpandTestGlobals()).WithActors([]KustoActor{actor})

	info := service.GetCodeActions(12, 12, 1, CodeActionOptionsDefault, true, actorName, CancellationToken{})
	if legacyGetCalled || actor.getPosition != 12 || actor.getStart != 12 || actor.getLength != 1 {
		t.Fatalf("explicit GetActions coordinates = (%d, %d, %d), legacy called = %t", actor.getPosition, actor.getStart, actor.getLength, legacyGetCalled)
	}
	if len(info.Actions) != 1 {
		t.Fatalf("explicit actions = %#v", info.Actions)
	}
	action := info.Actions[0].(ApplyAction)
	result := service.ApplyCodeAction(action, 12, CodeActionOptionsDefault, CancellationToken{})
	if legacyApplyCalled || actor.applyCaret != 12 {
		t.Fatalf("explicit ApplyAction caret = %d, legacy called = %t", actor.applyCaret, legacyApplyCalled)
	}
	if len(actor.applyData) != 2 || actor.applyData[0] != "opaque" || actor.applyData[1] != "12" {
		t.Fatalf("explicit ApplyAction data = %#v", actor.applyData)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.Changes[0].Start != 12 {
		t.Fatalf("explicit UTF-16 change was converted = %#v", result.Actions[0])
	}
	move, ok := result.Actions[1].(MoveCaretAction)
	if !ok || move.NewCaretPosition != 13 {
		t.Fatalf("explicit UTF-16 caret was converted = %#v", result.Actions[1])
	}
}

func TestKustoCodeServiceWithActorsEmptySuppressesDefaultsAndNilRestoresThem(t *testing.T) {
	service := NewKustoCodeService("T | count", convertMacroExpandTestGlobals())
	empty := service.WithActors([]KustoActor{})
	if len(empty.actors) != 0 {
		t.Fatalf("empty actor set = %#v", empty.actors)
	}
	info := empty.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{})
	if len(info.Actions) != 0 {
		t.Fatalf("disabled default actor actions = %#v", info.Actions)
	}
	unknown := empty.ApplyCodeAction(NewCodeAction("Rename", "", "", []string{kustoRenameActorName}), 0, CodeActionOptionsDefault, CancellationToken{})
	if unknown.FailureReason != "Unknown actor: "+kustoRenameActorName {
		t.Fatalf("disabled actor apply result = %#v", unknown)
	}

	restored := empty.WithActors(nil)
	names := KustoActorNamesAll()
	if len(restored.actors) != len(names) {
		t.Fatalf("restored actors = %d, want %d", len(restored.actors), len(names))
	}
	for index, actor := range restored.actors {
		if actor == nil || actor.Name() != names[index] {
			t.Fatalf("restored actor[%d] = %#v, want %q", index, actor, names[index])
		}
	}
}

func TestKustoCodeServiceRoutesConvertToMacroExpandActor(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "T | union U | count"
	service := NewKustoCodeService(text, globals)
	position := strings.Index(text, "union")
	info := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoConvertToMacroExpandActorName, CancellationToken{})
	action := requireMacroExpandAction(t, info, ConvertUnionKind)
	data := action.ApplyData()
	if len(data) < 2 || data[len(data)-1] != kustoConvertToMacroExpandActorName {
		t.Fatalf("service action data = %#v", data)
	}
	result := service.ApplyCodeAction(action, position, CodeActionOptionsDefault, CancellationToken{})
	assertMacroExpandResult(t, text, result, "macro-expand entity_group [T, U] as _scope\n(\n    _scope | count\n)\n")
}

func TestKustoCodeServiceMalformedSyntaxAndActorPayloadsFailClosed(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "T | where (X >"
	service := NewKustoCodeService(text, globals)
	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) == 0 {
		t.Fatalf("malformed KQL %q produced no syntax diagnostics", text)
	}

	for _, actorName := range KustoActorNamesAll() {
		_ = service.GetCodeActions(len(text), 0, len(text), CodeActionOptionsDefault, true, actorName, CancellationToken{})
	}

	badActions := []ApplyAction{
		NewCodeAction(InlineOnceKind, "", "", []string{"not-a-position", kustoFunctionInlinerActorName}),
		NewCodeAction(ConvertSelectionKind, "", "", []string{"bad", "range", "data", kustoConvertToMacroExpandActorName}),
	}
	for _, action := range badActions {
		result := service.ApplyCodeAction(action, 0, CodeActionOptionsDefault, CancellationToken{})
		if result.FailureReason == "" || len(result.Actions) != 0 {
			t.Fatalf("malformed service action result = %#v", result)
		}
	}
}
