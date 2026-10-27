package kustolang

import (
	"sort"
	"strconv"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoActors.cs and the actor
// dispatch in KustoCodeService.cs. See third_party/kusto/LICENSE.

type kustoActorRegistration struct {
	actorName  string
	getActions func(*KustoCodeService, *KustoCode, int, int, int, CodeActionOptions, bool, CancellationToken) []CodeAction
	apply      func(*KustoCodeService, *KustoCode, ApplyAction, int, CodeActionOptions, CancellationToken) CodeActionResult
}

// KustoActor implements a code fix or refactoring provider.
type KustoActor interface {
	Name() string
	GetActions(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction
	ApplyAction(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
}

type kustoActorUTF16 interface {
	GetActionsUTF16(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction
	ApplyActionUTF16(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
}

type KustoActorCallbacks struct {
	Name       string
	GetActions func(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction
	Apply      func(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
}

type KustoActorBase struct {
	callbacks KustoActorCallbacks
}

func NewKustoActor(callbacks KustoActorCallbacks) *KustoActorBase {
	if callbacks.Name == "" {
		callbacks.Name = "KustoActor"
	}
	return &KustoActorBase{callbacks: callbacks}
}

func (a *KustoActorBase) Name() string {
	if a == nil || a.callbacks.Name == "" {
		return "KustoActor"
	}
	return a.callbacks.Name
}

func (a *KustoActorBase) GetActions(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction {
	if a != nil && a.callbacks.GetActions != nil {
		return a.callbacks.GetActions(service, code, position, selectionStart, selectionLength, options, actions, waitForAnalysis, cancellationToken)
	}
	return actions
}

func (a *KustoActorBase) ApplyAction(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	if a != nil && a.callbacks.Apply != nil {
		return a.callbacks.Apply(service, code, action, caretPosition, options, cancellationToken)
	}
	return NewCodeActionFailure()
}

func (a kustoActorRegistration) Name() string {
	return a.actorName
}

func (a kustoActorRegistration) GetActions(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction {
	if a.getActions == nil {
		return actions
	}
	return append(actions, a.getActions(service, code, position, selectionStart, selectionLength, options, waitForAnalysis, cancellationToken)...)
}

func (a kustoActorRegistration) GetActionsUTF16(service *KustoCodeService, code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, actions []CodeAction, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction {
	return a.GetActions(service, code, position, selectionStart, selectionLength, options, actions, waitForAnalysis, cancellationToken)
}

func (a kustoActorRegistration) ApplyAction(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	if a.apply == nil {
		return NewCodeActionFailure()
	}
	return a.apply(service, code, action, caretPosition, options, cancellationToken)
}

func (a kustoActorRegistration) ApplyActionUTF16(service *KustoCodeService, code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	return a.ApplyAction(service, code, action, caretPosition, options, cancellationToken)
}

func kustoActorRegistrations() []kustoActorRegistration {
	return []kustoActorRegistration{
		{
			actorName: kustoFunctionInlinerActorName,
			getActions: func(_ *KustoCodeService, code *KustoCode, position, start, length int, options CodeActionOptions, _ bool, token CancellationToken) []CodeAction {
				return NewInlineDatabaseFunctionActor().GetActionsUTF16(code, position, start, length, options, token).Actions
			},
			apply: func(_ *KustoCodeService, code *KustoCode, action ApplyAction, caret int, options CodeActionOptions, token CancellationToken) CodeActionResult {
				return NewInlineDatabaseFunctionActor().ApplyActionUTF16(code, action, caret, options, token)
			},
		},
		{
			actorName: kustoExtractExpressionActorName,
			getActions: func(_ *KustoCodeService, code *KustoCode, position, start, length int, options CodeActionOptions, _ bool, token CancellationToken) []CodeAction {
				return NewExtractExpressionActor().GetActionsUTF16(code, position, start, length, options, token).Actions
			},
			apply: func(_ *KustoCodeService, code *KustoCode, action ApplyAction, caret int, options CodeActionOptions, token CancellationToken) CodeActionResult {
				return NewExtractExpressionActor().ApplyActionUTF16(code, action, caret, options, token)
			},
		},
		{
			actorName: kustoConvertToMacroExpandActorName,
			getActions: func(_ *KustoCodeService, code *KustoCode, position, start, length int, options CodeActionOptions, _ bool, token CancellationToken) []CodeAction {
				return NewConvertToMacroExpandActor().GetActionsUTF16(code, position, start, length, options, token).Actions
			},
			apply: func(_ *KustoCodeService, code *KustoCode, action ApplyAction, caret int, options CodeActionOptions, token CancellationToken) CodeActionResult {
				return NewConvertToMacroExpandActor().ApplyActionUTF16(code, action, caret, options, token)
			},
		},
		{
			actorName: kustoRenameActorName,
			getActions: func(service *KustoCodeService, code *KustoCode, position, _, _ int, _ CodeActionOptions, _ bool, token CancellationToken) []CodeAction {
				bytePosition, ok := utf16RangeStartToByte(code.Text, position)
				if !ok {
					return nil
				}
				if action := kustoRenameActorGetAction(service, code, bytePosition, token); action != nil {
					data := action.ApplyData()
					if len(data) != 1 {
						return nil
					}
					byteStart, err := strconv.Atoi(data[0])
					if err != nil {
						return nil
					}
					return []CodeAction{action.WithDataStrings(strconv.Itoa(byteOffsetToUTF16Position(code.Text, byteStart)))}
				}
				return nil
			},
			apply: func(_ *KustoCodeService, _ *KustoCode, action ApplyAction, _ int, _ CodeActionOptions, _ CancellationToken) CodeActionResult {
				return kustoRenameActorApplyAction(action)
			},
		},
		{
			actorName: kustoDiagnosticFixActorName,
			getActions: func(service *KustoCodeService, code *KustoCode, position, start, length int, options CodeActionOptions, wait bool, token CancellationToken) []CodeAction {
				bytePosition, positionOK := utf16RangeStartToByte(code.Text, position)
				byteStart, byteLength, selectionOK := utf16RangeToBytes(code.Text, start, length)
				if !positionOK || !selectionOK {
					return nil
				}
				return service.kustoCodeServiceGetDiagnosticFixActions(code, bytePosition, byteStart, byteLength, options, wait, token)
			},
			apply: kustoActorApplyDiagnosticFix,
		},
	}
}

func KustoActorsAll() []KustoActor {
	registrations := kustoActorRegistrations()
	actors := make([]KustoActor, 0, len(registrations))
	for _, actor := range registrations {
		actors = append(actors, actor)
	}
	return ListToReadOnly(actors)
}

func kustoActorByName(actors []KustoActor, name string) (KustoActor, bool) {
	for _, actor := range actors {
		if actor != nil && actor.Name() == name {
			return actor, true
		}
	}
	return nil, false
}

func kustoActorGetActions(actor KustoActor, service *KustoCodeService, code *KustoCode, position, start, length int, options CodeActionOptions, wait bool, token CancellationToken) (actions []CodeAction) {
	defer func() {
		if recover() != nil {
			actions = nil
		}
	}()
	if actorUTF16, ok := actor.(kustoActorUTF16); ok {
		return actorUTF16.GetActionsUTF16(service, code, position, start, length, options, nil, wait, token)
	}
	bytePosition, positionOK := utf16RangeStartToByte(code.Text, position)
	byteStart, byteLength, selectionOK := utf16RangeToBytes(code.Text, start, length)
	if !positionOK || !selectionOK {
		return nil
	}
	return actor.GetActions(service, code, bytePosition, byteStart, byteLength, options, nil, wait, token)
}

func kustoActorApply(actor KustoActor, service *KustoCodeService, code *KustoCode, action ApplyAction, caret int, options CodeActionOptions, token CancellationToken) (result CodeActionResult) {
	result = NewCodeActionFailure()
	defer func() {
		if recover() != nil {
			result = NewCodeActionFailure()
		}
	}()
	if actorUTF16, ok := actor.(kustoActorUTF16); ok {
		return actorUTF16.ApplyActionUTF16(service, code, action, caret, options, token)
	}
	byteCaret, ok := utf16RangeStartToByte(code.Text, caret)
	if !ok {
		return NewCodeActionFailure("Invalid caret position")
	}
	return extractionResultToUTF16(code.Text, actor.ApplyAction(service, code, action, byteCaret, options, token))
}

func kustoActorApplyDiagnosticFix(service *KustoCodeService, code *KustoCode, action ApplyAction, caret int, options CodeActionOptions, token CancellationToken) CodeActionResult {
	data := action.ApplyData()
	if len(data) == 0 {
		return NewCodeActionFailure("Unknown fixer")
	}
	fixerName := data[len(data)-1]
	action = action.RemoveData(1)
	fixer := service.kustoCodeServiceFixersByName()[fixerName]
	if fixer == nil {
		return NewCodeActionFailure("Unknown fixer")
	}
	byteCaret, ok := utf16RangeStartToByte(code.Text, caret)
	if !ok {
		return NewCodeActionFailure("Invalid caret position")
	}
	return extractionResultToUTF16(code.Text, fixer.ApplyFixAction(code, action, byteCaret, options, token))
}

func kustoSortCodeActions(actions []CodeAction) {
	sort.SliceStable(actions, func(i, j int) bool {
		return ordinalCompare(actions[i].Base().Title, actions[j].Base().Title, true) < 0
	})
}
