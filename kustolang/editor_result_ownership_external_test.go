package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCodeServiceSnapshotsCustomResultCollections(t *testing.T) {
	diagnostic := kustolang.NewDiagnostic("APP001", "original")
	diagnostics := []kustolang.Diagnostic{diagnostic}
	analyzers := []kustolang.AnalyzerInfo{{Name: "sample", Diagnostics: diagnostics}}
	quickParts := []kustolang.ClassifiedText{kustolang.NewClassifiedText("original")}
	quick := kustolang.QuickInfo{Items: []kustolang.QuickInfoItem{{Kind: kustolang.QuickInfoKindText, Parts: quickParts}}}
	completionTexts := []kustolang.CompletionText{kustolang.NewCompletionText("original", false, false)}
	completion := kustolang.CompletionInfo{Items: []kustolang.CompletionItem{{DisplayText: "original", ApplyTexts: completionTexts}}}
	classifications := []kustolang.ClassifiedRange{kustolang.NewClassifiedRange(kustolang.ClassificationKindKeyword, 0, 1)}
	outlines := []kustolang.OutlineRange{kustolang.NewOutlineRange(0, 1, "original")}
	index := 1
	parameters := []kustolang.ClientParameter{kustolang.NewClientParameterAt("p", &index, 0, 1)}
	sharedAction := kustolang.NewCodeAction("custom", "Original", "", []string{"data"})
	sharedResult := kustolang.NewCodeActionChange([]kustolang.TextEdit{kustolang.NewTextEdit(0, 1, "x")}, "x")

	service := kustolang.NewCustomCodeService("x", kustolang.CustomCodeServiceCallbacks{
		TryGetCachedDiagnostics: func() ([]kustolang.Diagnostic, bool) { return diagnostics, true },
		GetDiagnostics:          func(bool, kustolang.CancellationToken) []kustolang.Diagnostic { return diagnostics },
		GetAnalyzers:            func() []kustolang.AnalyzerInfo { return analyzers },
		GetQuickInfo: func(int, kustolang.QuickInfoOptions, kustolang.CancellationToken) kustolang.QuickInfo {
			return quick
		},
		GetCompletionItems: func(int, kustolang.CompletionOptions, kustolang.CancellationToken) kustolang.CompletionInfo {
			return completion
		},
		GetClassifications: func(int, int, bool, bool, kustolang.CancellationToken) kustolang.ClassificationInfo {
			return kustolang.ClassificationInfo{Classifications: classifications}
		},
		GetOutlines: func(kustolang.OutliningOptions, kustolang.CancellationToken) kustolang.OutlineInfo {
			return kustolang.OutlineInfo{Ranges: outlines}
		},
		GetClientParameters: func() []kustolang.ClientParameter { return parameters },
		GetCodeActions: func(int, int, int, kustolang.CodeActionOptions, bool, string, kustolang.CancellationToken) kustolang.CodeActionInfo {
			return kustolang.CodeActionInfo{Actions: []kustolang.CodeAction{sharedAction}}
		},
		ApplyCodeAction: func(kustolang.ApplyAction, int, kustolang.CodeActionOptions, kustolang.CancellationToken) kustolang.CodeActionResult {
			return sharedResult
		},
	})

	cached, ok := service.TryGetCachedDiagnostics()
	if !ok || len(cached) != 1 {
		t.Fatalf("cached diagnostics = %#v, %t", cached, ok)
	}
	cached[0] = kustolang.NewDiagnostic("APP999", "changed")
	service.GetDiagnostics(false, kustolang.CancellationToken{})[0] = kustolang.NewDiagnostic("APP998", "changed")
	service.GetAnalyzers()[0].Diagnostics[0] = kustolang.NewDiagnostic("APP997", "changed")
	service.GetQuickInfo(0, kustolang.QuickInfoOptionsDefault, kustolang.CancellationToken{}).Items[0].Parts[0] = kustolang.NewClassifiedText("changed")
	service.GetCompletionItems(0, kustolang.CompletionOptionsDefault, kustolang.CancellationToken{}).Items[0].ApplyTexts[0] = kustolang.NewCompletionText("changed", false, false)
	service.GetClassifications(0, 1, true, false, kustolang.CancellationToken{}).Classifications[0].Kind = kustolang.ClassificationKindComment
	service.GetOutlines(kustolang.OutliningOptionsDefault, kustolang.CancellationToken{}).Ranges[0].CollapsedText = "changed"
	returnedParameters := service.GetClientParameters()
	*returnedParameters[0].Index = 99
	returnedAction := service.GetCodeActions(0, 0, 0, kustolang.CodeActionOptionsDefault, false, "", kustolang.CancellationToken{}).Actions[0].(*kustolang.SingleAction)
	returnedAction.Data[0] = "changed"
	returnedResult := service.ApplyCodeAction(sharedAction, 0, kustolang.CodeActionOptionsDefault, kustolang.CancellationToken{})
	returnedResult.Actions[0].(kustolang.ChangeTextAction).Changes[0].InsertText = "changed"

	if diagnostics[0] != diagnostic || analyzers[0].Diagnostics[0] != diagnostic {
		t.Fatal("diagnostic or analyzer callback storage was mutated")
	}
	if quickParts[0].Text != "original" || completionTexts[0].Text != "original" || classifications[0].Kind != kustolang.ClassificationKindKeyword || outlines[0].CollapsedText != "original" {
		t.Fatal("editor callback collection storage was mutated")
	}
	if parameters[0].Index == nil || *parameters[0].Index != 1 {
		t.Fatal("client parameter pointer storage was mutated")
	}
	if sharedAction.Data[0] != "data" || sharedResult.Actions[0].(kustolang.ChangeTextAction).Changes[0].InsertText != "x" {
		t.Fatal("code action callback storage was mutated")
	}

	infoCopy := service.GetAnalyzers()[0].DiagnosticsList()
	infoCopy[0] = kustolang.NewDiagnostic("APP996", "changed")
	quickCopy := service.GetQuickInfo(0, kustolang.QuickInfoOptionsDefault, kustolang.CancellationToken{}).ItemsList()
	quickCopy[0].Parts[0] = kustolang.NewClassifiedText("changed accessor")
	completionCopy := service.GetCompletionItems(0, kustolang.CompletionOptionsDefault, kustolang.CancellationToken{}).ItemsList()
	completionCopy[0].ApplyTextsList()[0] = kustolang.NewCompletionText("changed accessor", false, false)
	if diagnostics[0] != diagnostic || quickParts[0].Text != "original" || completionTexts[0].Text != "original" {
		t.Fatal("defensive editor result accessors exposed callback storage")
	}
}
