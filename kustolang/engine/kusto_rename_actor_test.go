package kustolang

import (
	"strconv"
	"strings"
	"testing"
)

func TestKustoRenameActorFiltersAndFindsDeclarationsAndReferences(t *testing.T) {
	text := `datatable(Account:string)["a"] | project Account`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	declaration := strings.Index(text, "Account")
	reference := strings.LastIndex(text, "Account")

	for _, test := range []struct {
		position  int
		nameStart int
	}{
		{position: declaration, nameStart: declaration},
		{position: reference + 2, nameStart: reference},
	} {
		position := test.position
		actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{})
		if len(actions.Actions) != 1 {
			t.Fatalf("rename actions at %d = %#v, want one", position, actions.Actions)
		}
		action, ok := actions.Actions[0].(ApplyAction)
		if !ok || action.Base().Kind != "Rename" || action.Base().Title != "Rename" || action.Base().Description != "Rename selected item" {
			t.Fatalf("rename action at %d = %#v", position, actions.Actions[0])
		}
		wantData := []string{strconv.Itoa(test.nameStart), kustoRenameActorName}
		if got := action.ApplyData(); len(got) != 2 || got[0] != wantData[0] || got[1] != wantData[1] {
			t.Fatalf("rename data at %d = %#v, want %#v", position, got, wantData)
		}
	}

	if actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("diagnostic actor returned rename action: %#v", actions.Actions)
	}
	if actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, true, "NoSuchActor", CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("unknown actor actions = %#v", actions.Actions)
	}
	if actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, true, "", CancellationToken{}); len(actions.Actions) != 1 {
		t.Fatalf("default actor actions = %#v, want rename", actions.Actions)
	}
}

func TestKustoRenameActorRejectsUnresolvedNamesAndWhitespace(t *testing.T) {
	text := `let Known=1; print a=Known, b=Missing`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	for _, position := range []int{
		strings.Index(text, "Missing"),
		strings.Index(text, " print") + 1,
	} {
		if actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 0 {
			t.Fatalf("rename actions at invalid position %d = %#v", position, actions.Actions)
		}
	}
}

func TestKustoRenameActorApplyMovesCaretAndRequestsRename(t *testing.T) {
	text := `let Threshold=10; print Value=Threshold`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	reference := strings.LastIndex(text, "Threshold")
	actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{})
	action, ok := actions.Actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("rename action = %#v", actions.Actions)
	}

	result := service.ApplyCodeAction(action, reference, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) != 2 {
		t.Fatalf("rename result = %#v", result)
	}
	move, ok := result.Actions[0].(MoveCaretAction)
	if !ok || move.NewCaretPosition != reference {
		t.Fatalf("move action = %#v, want %d", result.Actions[0], reference)
	}
	if _, ok := result.Actions[1].(RenameAction); !ok {
		t.Fatalf("rename request = %#v", result.Actions[1])
	}
}

func TestKustoRenameActorRejectsMalformedPayload(t *testing.T) {
	service := NewKustoCodeService(`let x=1; print x`).WithAnalyzers([]KustoAnalyzer{})
	for _, action := range []ApplyAction{
		NewCodeAction("Rename", "Rename", "", []string{kustoRenameActorName}),
		NewCodeAction("Rename", "Rename", "", []string{"not-a-position", kustoRenameActorName}),
		NewCodeAction("Rename", "Rename", "", []string{"2147483648", kustoRenameActorName}),
	} {
		result := service.ApplyCodeAction(action, 0, CodeActionOptionsDefault, CancellationToken{})
		if result.FailureReason == "" || len(result.Actions) != 0 {
			t.Fatalf("malformed rename result = %#v", result)
		}
	}
}

func TestKustoRenameActorPreservesPositionsAfterNonBMPText(t *testing.T) {
	text := "// \U0001F600\n" + `let Threshold=10; print Value=Threshold`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	reference := strings.LastIndex(text, "Threshold")
	utf16Reference := byteOffsetToUTF16Position(text, reference)
	if reference == utf16Reference {
		t.Fatalf("test setup did not produce distinct byte and UTF-16 positions: %d", reference)
	}
	actions := service.GetCodeActions(utf16Reference, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("non-BMP rename actions = %#v", actions.Actions)
	}
	result := service.ApplyCodeAction(actions.Actions[0].(ApplyAction), utf16Reference, CodeActionOptionsDefault, CancellationToken{})
	move, ok := result.Actions[0].(MoveCaretAction)
	if !ok || move.NewCaretPosition != utf16Reference {
		t.Fatalf("non-BMP move = %#v, want UTF-16 position %d (byte position %d)", result.Actions, utf16Reference, reference)
	}
}

func TestKustoRenameActorHonorsCachedAnalysisAndOffsetService(t *testing.T) {
	text := `let Threshold=10; print Value=Threshold`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	reference := strings.LastIndex(text, "Threshold")
	if actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, false, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("uncached rename actions = %#v, want none", actions.Actions)
	}
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics while priming cache: %#v", diagnostics)
	}
	if actions := service.GetCodeActions(reference, 0, 0, CodeActionOptionsDefault, false, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 1 {
		t.Fatalf("cached rename actions = %#v, want one", actions.Actions)
	}

	const offset = 17
	offsetService := NewOffsetCodeService(service, offset)
	actions := offsetService.GetCodeActions(reference+offset, 0, 0, CodeActionOptionsDefault, false, kustoRenameActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("offset rename actions = %#v", actions.Actions)
	}
	result := offsetService.ApplyCodeAction(actions.Actions[0].(ApplyAction), reference+offset, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) != 2 {
		t.Fatalf("offset rename result = %#v", result)
	}
	move, ok := result.Actions[0].(MoveCaretAction)
	if !ok || move.NewCaretPosition != reference+offset {
		t.Fatalf("offset move = %#v, want %d", result.Actions[0], reference+offset)
	}
}

func TestKustoRenameActorSupportsFunctionParameters(t *testing.T) {
	text := `let F=(Param:string) { print Value=Param }; F("x")`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	for _, position := range []int{strings.Index(text, "Param"), strings.LastIndex(text, "Param")} {
		if actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 1 {
			t.Fatalf("parameter rename actions at %d = %#v", position, actions.Actions)
		}
	}
}

func TestKustoRenameActorRejectsDeclarationOnlySymbols(t *testing.T) {
	text := `let Unused=1; print 1`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	position := strings.Index(text, "Unused")
	if actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("declaration-only rename actions = %#v, want none", actions.Actions)
	}
}

func TestKustoRenameActorUsesLeadingTriviaAffinity(t *testing.T) {
	text := `let X=1; print    X`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	spaceBeforeReference := strings.Index(text, "    X") + 1
	actions := service.GetCodeActions(spaceBeforeReference, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("leading-trivia rename actions = %#v, want one", actions.Actions)
	}
	action := actions.Actions[0].(ApplyAction)
	wantStart := strings.LastIndex(text, "X")
	if data := action.ApplyData(); len(data) != 2 || data[0] != strconv.Itoa(wantStart) {
		t.Fatalf("leading-trivia rename data = %#v, want start %d", data, wantStart)
	}
}

func TestKustoRenameActorSupportsQueryParameters(t *testing.T) {
	text := `declare query_parameters(p:string); print p`
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{})
	for _, position := range []int{strings.Index(text, "p:string"), strings.LastIndex(text, "p")} {
		if actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, true, kustoRenameActorName, CancellationToken{}); len(actions.Actions) != 1 {
			t.Fatalf("query-parameter rename actions at %d = %#v", position, actions.Actions)
		}
	}
}
