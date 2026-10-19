package kustolang

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestExtractExpressionActorExtractsLiteralAtCaret(t *testing.T) {
	text := "print Value=42"
	code := ParseAndAnalyze(text)
	actor := NewExtractExpressionActor()
	position := strings.Index(text, "42") + 1

	action := requireExtractionAction(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ExtractValueKind)
	if got, want := action.ApplyData(), []string{strconv.Itoa(position), "0"}; !equalStrings(got, want) {
		t.Fatalf("literal action data = %#v, want %#v", got, want)
	}
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, result, "let varName = 42;\nprint Value=varName", 4)
	change := result.Actions[0].(ChangeTextAction)
	if replayed := NewEditString(text).ApplyAll(change.Changes).CurrentText; replayed != change.ChangedText {
		t.Fatalf("replayed edits = %q, want %q", replayed, change.ChangedText)
	}
}

func TestExtractExpressionActorSupportsUpstreamLiteralFormsAndCaretEdges(t *testing.T) {
	actor := NewExtractExpressionActor()
	for _, literal := range []string{`"text"`, "true", "1.5", "2h", "datetime(2024-01-02)"} {
		t.Run(literal, func(t *testing.T) {
			text := "print Value=" + literal
			code := ParseAndAnalyze(text)
			start := strings.Index(text, literal)
			for _, position := range []int{start, start + len(literal)} {
				info := actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{})
				if len(info.Actions) != 1 || info.Actions[0].Base().Kind != ExtractValueKind {
					t.Fatalf("literal %q actions at %d = %#v", literal, position, info.Actions)
				}
			}
		})
	}
}

func TestExtractExpressionActorExtractsSelectedFreeStandingExpression(t *testing.T) {
	text := "print Value=1 + 2"
	code := ParseAndAnalyze(text)
	actor := NewExtractExpressionActor()
	start := strings.Index(text, "1 + 2")
	length := len("1 + 2")

	action := requireExtractionAction(t, actor.GetActions(code, start, start, length, CodeActionOptionsDefault, CancellationToken{}), ExtractExpressionKind)
	result := actor.ApplyAction(code, action, start, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, result, "let varName = 1 + 2;\nprint Value=varName", 4)

	letText := "let Existing=1 + 2;\nprint Existing"
	letCode := ParseAndAnalyze(letText)
	letStart := strings.Index(letText, "1 + 2")
	if actions := actor.GetActions(letCode, letStart, letStart, length, CodeActionOptionsDefault, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("let initializer actions = %#v, want none", actions.Actions)
	}
}

func TestExtractExpressionActorTrimsSelectionTrivia(t *testing.T) {
	text := "print Value= 1 + 2 \n"
	code := ParseAndAnalyze(text)
	actor := NewExtractExpressionActor()
	start := strings.Index(text, " 1 + 2")
	length := len(" 1 + 2 ")
	action := requireExtractionAction(t, actor.GetActions(code, start, start, length, CodeActionOptionsDefault, CancellationToken{}), ExtractExpressionKind)
	result := actor.ApplyAction(code, action, start, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, result, "let varName = 1 + 2;\nprint Value= varName \n", 4)
}

func TestExtractExpressionActorUsesAvailableCollisionFreeName(t *testing.T) {
	text := "let varName=0;\nlet varName1=1;\nprint Value=42"
	code := ParseAndAnalyze(text)
	actor := NewExtractExpressionActor()
	position := strings.LastIndex(text, "42")
	action := requireExtractionAction(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ExtractValueKind)
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	want := "let varName=0;\nlet varName1=1;\nlet varName2 = 42;\nprint Value=varName2"
	assertExtractionResult(t, result, want, strings.Index(want, "varName2"))
}

func TestExtractExpressionActorExtractsColumnExpressionAsScalarFunction(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))))
	text := "T | where X > 1"
	code := ParseAndAnalyze(text, globals)
	actor := NewExtractExpressionActor()
	start := strings.Index(text, "X > 1")
	length := len("X > 1")
	info := actor.GetActions(code, start, start, length, CodeActionOptionsDefault, CancellationToken{})
	if len(info.Actions) != 1 || info.Actions[0].Base().Kind != ExtractFunctionKind {
		t.Fatalf("column expression actions = %#v, want only Extract Function", info.Actions)
	}
	action := info.Actions[0].(ApplyAction)
	result := actor.ApplyAction(code, action, start, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, result, "let functionName = (X: long) { X > 1 };\nT | where functionName(X)", 4)
}

func TestExtractExpressionActorExtractsSelectedOperatorAsTabularFunction(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))))
	text := "T\n| where X > 1\n| project X"
	code := ParseAndAnalyze(text, globals)
	actor := NewExtractExpressionActor()
	start := strings.Index(text, "where")
	length := len("where X > 1")
	action := requireExtractionAction(t, actor.GetActions(code, start, start, length, CodeActionOptionsDefault, CancellationToken{}), ExtractFunctionKind)
	result := actor.ApplyAction(code, action, start, CodeActionOptionsDefault, CancellationToken{})
	want := "let functionName = (inputTable: (X: long)) {\n    inputTable\n    | where X > 1\n};\nT\n| invoke functionName()\n| project X"
	assertExtractionResult(t, result, want, 4)
}

func TestExtractExpressionActorRejectsMalformedSyntaxAndPayloads(t *testing.T) {
	actor := NewExtractExpressionActor()
	malformed := "print Value=(1 + )"
	code := ParseAndAnalyze(malformed)
	start := strings.Index(malformed, "1 +")
	if actions := actor.GetActions(code, start, start, len("1 + "), CodeActionOptionsDefault, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("malformed syntax actions = %#v, want none", actions.Actions)
	}

	valid := ParseAndAnalyze("print Value=42")
	for _, action := range []ApplyAction{
		NewCodeAction(ExtractValueKind, ExtractValueKind, "", nil),
		NewCodeAction(ExtractValueKind, ExtractValueKind, "", []string{"bad", "0"}),
		NewCodeAction(ExtractValueKind, ExtractValueKind, "", []string{"999", "0"}),
		NewCodeAction("NoSuchKind", "NoSuchKind", "", []string{"0", "0"}),
	} {
		result := actor.ApplyAction(valid, action, 0, CodeActionOptionsDefault, CancellationToken{})
		if result.FailureReason == "" || len(result.Actions) != 0 {
			t.Fatalf("malformed action result = %#v, want failure", result)
		}
	}
}

func TestExtractExpressionActorUTF16BoundaryAndApplication(t *testing.T) {
	text := "// \U0001f600\nprint Value=42"
	code := ParseAndAnalyze(text)
	actor := NewExtractExpressionActor()
	bytePosition := strings.Index(text, "42")
	utf16Position := len(utf16.Encode([]rune(text[:bytePosition])))
	if bytePosition == utf16Position {
		t.Fatal("test setup requires distinct byte and UTF-16 positions")
	}
	action := requireExtractionAction(t, actor.GetActionsUTF16(code, utf16Position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ExtractValueKind)
	if got := action.ApplyData(); len(got) != 2 || got[0] != strconv.Itoa(utf16Position) {
		t.Fatalf("UTF-16 action data = %#v, want start %d", got, utf16Position)
	}
	result := actor.ApplyActionUTF16(code, action, utf16Position, CodeActionOptionsDefault, CancellationToken{})
	want := "// \U0001f600\nlet varName = 42;\nprint Value=varName"
	assertExtractionResult(t, result, want, len(utf16.Encode([]rune("// \U0001f600\nlet "))))
	change := result.Actions[0].(ChangeTextAction)
	for _, edit := range change.Changes {
		if edit.Start == bytePosition {
			t.Fatalf("UTF-16 edit retained byte position: %#v", change.Changes)
		}
	}

	service := NewKustoCodeServiceForCode(code).WithAnalyzers([]KustoAnalyzer{})
	serviceAction := requireExtractionAction(t, service.GetCodeActions(
		utf16Position, 0, 0, CodeActionOptionsDefault, true, kustoExtractExpressionActorName, CancellationToken{},
	), ExtractValueKind)
	serviceResult := service.ApplyCodeAction(serviceAction, utf16Position, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, serviceResult, want, len(utf16.Encode([]rune("// \U0001f600\nlet "))))
}

func TestExtractExpressionActorHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := NewCancellationToken(ctx)
	actor := NewExtractExpressionActor()
	code := ParseAndAnalyze("print Value=42")
	action := NewCodeAction(ExtractValueKind, ExtractValueKind, "", []string{"12", "0"})

	for name, invoke := range map[string]func(){
		"get":   func() { actor.GetActions(code, 12, 0, 0, CodeActionOptionsDefault, token) },
		"apply": func() { actor.ApplyAction(code, action, 12, CodeActionOptionsDefault, token) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				err, ok := recovered.(error)
				if !ok || !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation panic = %#v, want context.Canceled", recovered)
				}
			}()
			invoke()
		})
	}
}

func TestKustoCodeServiceRoutesExtractExpressionActor(t *testing.T) {
	text := "print Value=42"
	position := strings.Index(text, "42")
	service := NewKustoCodeService(text)

	if actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, false, kustoExtractExpressionActorName, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("uncached wait-false actions = %#v", actions.Actions)
	}
	service.GetDiagnostics(true, CancellationToken{})
	actions := service.GetCodeActions(position, 0, 0, CodeActionOptionsDefault, false, kustoExtractExpressionActorName, CancellationToken{})
	action := requireExtractionAction(t, actions, ExtractValueKind)
	data := action.ApplyData()
	if len(data) != 3 || data[len(data)-1] != kustoExtractExpressionActorName {
		t.Fatalf("service action data = %#v", data)
	}

	result := service.ApplyCodeAction(action, position, CodeActionOptionsDefault, CancellationToken{})
	assertExtractionResult(t, result, "let varName = 42;\nprint Value=varName", 4)
}

func requireExtractionAction(t *testing.T, info CodeActionInfo, kind string) ApplyAction {
	t.Helper()
	for _, action := range info.Actions {
		if action.Base().Kind == kind {
			apply, ok := action.(ApplyAction)
			if !ok {
				t.Fatalf("action %q is not applicable: %#v", kind, action)
			}
			return apply
		}
	}
	t.Fatalf("actions = %#v, want kind %q", info.Actions, kind)
	return nil
}

func assertExtractionResult(t *testing.T, result CodeActionResult, changedText string, caret int) {
	t.Helper()
	if result.FailureReason != "" || len(result.Actions) != 3 {
		t.Fatalf("extraction result = %#v, want change, move and rename", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != changedText || len(change.Changes) != 2 {
		t.Fatalf("change = %#v, want text %q and two edits", result.Actions[0], changedText)
	}
	move, ok := result.Actions[1].(MoveCaretAction)
	if !ok || move.NewCaretPosition != caret {
		t.Fatalf("move = %#v, want caret %d", result.Actions[1], caret)
	}
	if _, ok := result.Actions[2].(RenameAction); !ok {
		t.Fatalf("rename = %#v, want RenameAction", result.Actions[2])
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
