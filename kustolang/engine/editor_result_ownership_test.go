package kustolang

import "testing"

type sharedEditorResultService struct {
	*CommonCodeService
	diagnostics []Diagnostic
	completion  CompletionInfo
	quick       QuickInfo
	actions     CodeActionInfo
	result      CodeActionResult
	parameters  []ClientParameter
}

func (s *sharedEditorResultService) GetDiagnostics(bool, CancellationToken) []Diagnostic {
	return s.diagnostics
}

func (s *sharedEditorResultService) GetCompletionItems(int, CompletionOptions, CancellationToken) CompletionInfo {
	return s.completion
}

func (s *sharedEditorResultService) GetQuickInfo(int, QuickInfoOptions, CancellationToken) QuickInfo {
	return s.quick
}

func (s *sharedEditorResultService) GetCodeActions(int, int, int, CodeActionOptions, bool, string, CancellationToken) CodeActionInfo {
	return s.actions
}

func (s *sharedEditorResultService) ApplyCodeAction(ApplyAction, int, CodeActionOptions, CancellationToken) CodeActionResult {
	return s.result
}

func (s *sharedEditorResultService) GetClientParameters() []ClientParameter {
	return s.parameters
}

func TestOffsetCodeServiceSnapshotsUnadjustedNestedResults(t *testing.T) {
	diagnostic := NewDiagnostic("APP001", "original")
	completionTexts := []CompletionText{NewCompletionText("original", false, false)}
	quickParts := []ClassifiedText{NewClassifiedText("original")}
	action := NewCodeAction("custom", "Original", "", []string{"data"})
	change := NewCodeActionChange([]TextEdit{NewTextEdit(0, 1, "x")}, "x")
	index := 1
	service := &sharedEditorResultService{
		CommonCodeService: NewCommonCodeService("x"),
		diagnostics:       []Diagnostic{diagnostic},
		completion:        CompletionInfo{Items: []CompletionItem{{ApplyTexts: completionTexts}}},
		quick:             QuickInfo{Items: []QuickInfoItem{{Parts: quickParts}}},
		actions:           CodeActionInfo{Actions: []CodeAction{action}},
		result:            change,
		parameters:        []ClientParameter{NewClientParameterAt("p", &index, 0, 1)},
	}
	offset := NewOffsetCodeService(service, 0)

	offset.GetDiagnostics(false, CancellationToken{})[0] = NewDiagnostic("APP999", "changed")
	offset.GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{}).Items[0].ApplyTexts[0] = NewCompletionText("changed", false, false)
	offset.GetQuickInfo(0, QuickInfoOptionsDefault, CancellationToken{}).Items[0].Parts[0] = NewClassifiedText("changed")
	offset.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, false, "", CancellationToken{}).Actions[0].(*SingleAction).Data[0] = "changed"
	offset.ApplyCodeAction(action, 0, CodeActionOptionsDefault, CancellationToken{}).Actions[0].(ChangeTextAction).Changes[0].InsertText = "changed"
	parameters := offset.GetClientParameters()
	*parameters[0].Index = 99

	if service.diagnostics[0] != diagnostic || completionTexts[0].Text != "original" || quickParts[0].Text != "original" {
		t.Fatal("offset service exposed shared diagnostic or editor result storage")
	}
	if action.Data[0] != "data" || change.Actions[0].(ChangeTextAction).Changes[0].InsertText != "x" {
		t.Fatal("offset service exposed shared action storage")
	}
	if service.parameters[0].Index == nil || *service.parameters[0].Index != 1 {
		t.Fatal("offset service exposed client parameter pointer storage")
	}
}

func TestCustomCodeServiceSnapshotsCallbackActionGraphs(t *testing.T) {
	action := NewCodeAction("custom", "Original", "", []string{"data"})
	result := NewCodeActionChange([]TextEdit{NewTextEdit(0, 1, "x")}, "x")
	service := NewCustomCodeService("x", CustomCodeServiceCallbacks{
		GetCodeActions: func(int, int, int, CodeActionOptions, bool, string, CancellationToken) CodeActionInfo {
			return CodeActionInfo{Actions: []CodeAction{action}}
		},
		ApplyCodeAction: func(ApplyAction, int, CodeActionOptions, CancellationToken) CodeActionResult {
			return result
		},
	})

	service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, false, "", CancellationToken{}).Actions[0].(*SingleAction).Data[0] = "changed"
	service.ApplyCodeAction(action, 0, CodeActionOptionsDefault, CancellationToken{}).Actions[0].(ChangeTextAction).Changes[0].InsertText = "changed"
	if action.Data[0] != "data" || result.Actions[0].(ChangeTextAction).Changes[0].InsertText != "x" {
		t.Fatal("custom service exposed callback action storage")
	}
}
