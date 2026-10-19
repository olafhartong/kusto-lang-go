package kustolang

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestConvertToMacroExpandActorConvertsWholeSelectionAndReferences(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "T | union T | where X > 0"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	position := strings.Index(text, "T")

	action := requireMacroExpandAction(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ConvertSelectionKind)
	if got, want := action.ApplyData(), []string{"0", strconv.Itoa(len(text)), "0"}; !equalStrings(got, want) {
		t.Fatalf("selection action data = %#v, want %#v", got, want)
	}
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	want := "macro-expand entity_group [T] as _scope\n(\n    _scope | union _scope | where X > 0\n)\n"
	assertMacroExpandResult(t, text, result, want)
}

func TestConvertToMacroExpandActorConvertsExplicitNestedSelection(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "let q = T\n| where X > 0;\nq | count"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	position := strings.Index(text, "T")
	start := position
	length := len("T\n| where X > 0")

	action := requireMacroExpandAction(t, actor.GetActions(code, position, start, length, CodeActionOptionsDefault, CancellationToken{}), ConvertSelectionKind)
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	want := "let q = macro-expand entity_group [T] as _scope\n(\n    _scope\n    | where X > 0\n)\n;\nq | count"
	assertMacroExpandResult(t, text, result, want)
}

func TestConvertToMacroExpandActorConvertsUnionChainAndDownstreamPipe(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "T\n| union U\n| union V\n| where X > 0"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	position := strings.Index(text, "union")
	info := actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{})
	action := requireMacroExpandAction(t, info, ConvertUnionKind)
	if got := action.ApplyData(); !equalStrings(got, []string{strconv.Itoa(position)}) {
		t.Fatalf("union action data = %#v", got)
	}
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	want := "macro-expand entity_group [T, U, V] as _scope\n(\n    _scope\n    | where X > 0\n)\n"
	assertMacroExpandResult(t, text, result, want)
}

func TestConvertToMacroExpandActorEntityDiscoverySupportsQualifiedPaths(t *testing.T) {
	remoteTable := NewTableSymbol("Remote", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	remote := NewDatabaseSymbol("other", remoteTable)
	current := NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type())))
	globals := NewGlobalState(current).WithCluster(NewClusterSymbol("cluster", current, remote)).WithDatabase(current)
	text := "database(\"other\").Remote | union T | count"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	position := strings.Index(text, "Remote")

	if actorGetEntityExpressionStart(code.Syntax().GetTokenAt(position), globals) == nil {
		t.Fatal("qualified entity expression was not discovered")
	}
	action := requireMacroExpandAction(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ConvertUnionKind)
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	want := "macro-expand entity_group [database(\"other\").Remote, T] as _scope\n(\n    _scope | count\n)\n"
	assertMacroExpandResult(t, text, result, want)
}

func TestConvertToMacroExpandActorDiagonalStyleCaretAndRename(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "T | take 1"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	options := CodeActionOptionsDefault
	options.FormattingOptions = options.FormattingOptions.WithBrackettingStyle(BrackettingStyleDiagonal).WithIndentationSize(2)
	action := requireMacroExpandAction(t, actor.GetActions(code, 0, 0, 0, options, CancellationToken{}), ConvertSelectionKind)
	result := actor.ApplyAction(code, action, 0, options, CancellationToken{})
	want := "macro-expand entity_group [T] as _scope (\n  _scope | take 1\n)\n"
	assertMacroExpandResult(t, text, result, want)
}

func TestConvertToMacroExpandActorRejectsMalformedUnsupportedAndNestedMacro(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	actor := NewConvertToMacroExpandActor()

	for _, text := range []string{
		"T | union (print X=1)",
		"T | union Missing*",
		"T | union",
	} {
		code := ParseAndAnalyze(text, globals)
		position := strings.Index(text, "union")
		info := actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{})
		for _, action := range info.Actions {
			if action.Base().Kind == ConvertUnionKind {
				t.Fatalf("unsupported query %q offered union conversion: %#v", text, info.Actions)
			}
		}
	}

	nested := "macro-expand entity_group [T] as S (S | take 1)"
	nestedCode := ParseAndAnalyze(nested, globals)
	position := strings.LastIndex(nested, "S")
	if actions := actor.GetActions(nestedCode, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("nested macro actions = %#v, want none", actions.Actions)
	}

	valid := ParseAndAnalyze("T | take 1", globals)
	badActions := []ApplyAction{
		NewCodeAction(ConvertSelectionKind, "", "", nil),
		NewCodeAction(ConvertSelectionKind, "", "", []string{"bad", "1", "0"}),
		NewCodeAction(ConvertSelectionKind, "", "", []string{"0", "999", "0"}),
		NewCodeAction(ConvertUnionKind, "", "", []string{"999"}),
		NewCodeAction("NoSuchKind", "", "", []string{"0"}),
	}
	for _, action := range badActions {
		result := actor.ApplyAction(valid, action, 0, CodeActionOptionsDefault, CancellationToken{})
		if result.FailureReason == "" || len(result.Actions) != 0 {
			t.Fatalf("malformed action result = %#v, want failure", result)
		}
	}
}

func TestConvertToMacroExpandActorUTF16BoundaryAndApplication(t *testing.T) {
	globals := convertMacroExpandTestGlobals()
	text := "// \U0001f600\nT | take 1"
	code := ParseAndAnalyze(text, globals)
	actor := NewConvertToMacroExpandActor()
	bytePosition := strings.Index(text, "T")
	utf16Position := len(utf16.Encode([]rune(text[:bytePosition])))
	action := requireMacroExpandAction(t, actor.GetActionsUTF16(code, utf16Position, 0, 0, CodeActionOptionsDefault, CancellationToken{}), ConvertSelectionKind)
	data := action.ApplyData()
	if len(data) != 3 || data[2] != strconv.Itoa(utf16Position) {
		t.Fatalf("UTF-16 action data = %#v", data)
	}
	result := actor.ApplyActionUTF16(code, action, utf16Position, CodeActionOptionsDefault, CancellationToken{})
	want := "// \U0001f600\nmacro-expand entity_group [T] as _scope\n(\n    _scope | take 1\n)\n"
	assertMacroExpandUTF16Result(t, result, want)
	move := result.Actions[1].(MoveCaretAction)
	wantCaret := len(utf16.Encode([]rune("// \U0001f600\nmacro-expand entity_group [T] as ")))
	if move.NewCaretPosition != wantCaret {
		t.Fatalf("UTF-16 caret = %d", move.NewCaretPosition)
	}
}

func TestConvertToMacroExpandActorHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := NewCancellationToken(ctx)
	actor := NewConvertToMacroExpandActor()
	code := ParseAndAnalyze("T | union U", convertMacroExpandTestGlobals())
	action := NewCodeAction(ConvertUnionKind, "", "", []string{"4"})

	for name, invoke := range map[string]func(){
		"get":   func() { actor.GetActions(code, 4, 0, 0, CodeActionOptionsDefault, token) },
		"apply": func() { actor.ApplyAction(code, action, 4, CodeActionOptionsDefault, token) },
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

func convertMacroExpandTestGlobals() *GlobalState {
	column := NewColumnSymbol("X", ScalarTypes.Long.Type())
	return NewGlobalState(NewDatabaseSymbol("db",
		NewTableSymbol("T", column),
		NewTableSymbol("U", NewColumnSymbol("X", ScalarTypes.Long.Type())),
		NewTableSymbol("V", NewColumnSymbol("X", ScalarTypes.Long.Type())),
	))
}

func requireMacroExpandAction(t *testing.T, info CodeActionInfo, kind string) ApplyAction {
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

func assertMacroExpandResult(t *testing.T, original string, result CodeActionResult, want string) {
	t.Helper()
	if result.FailureReason != "" || len(result.Actions) != 3 {
		t.Fatalf("macro-expand result = %#v, want change, move, rename", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != want {
		t.Fatalf("change = %#v, want text %q", result.Actions[0], want)
	}
	if replayed := NewEditString(original).ApplyAll(change.Changes).CurrentText; replayed != want {
		t.Fatalf("replayed edits = %q, want %q; edits=%#v", replayed, want, change.Changes)
	}
	move, ok := result.Actions[1].(MoveCaretAction)
	if !ok || move.NewCaretPosition != strings.Index(want, macroExpandInitialEntityName) {
		t.Fatalf("move = %#v, want caret %d", result.Actions[1], strings.Index(want, macroExpandInitialEntityName))
	}
	if _, ok := result.Actions[2].(RenameAction); !ok {
		t.Fatalf("third action = %#v, want RenameAction", result.Actions[2])
	}
}

func assertMacroExpandUTF16Result(t *testing.T, result CodeActionResult, want string) {
	t.Helper()
	if result.FailureReason != "" || len(result.Actions) != 3 {
		t.Fatalf("macro-expand result = %#v, want change, move, rename", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != want {
		t.Fatalf("change = %#v, want text %q", result.Actions[0], want)
	}
	if _, ok := result.Actions[1].(MoveCaretAction); !ok {
		t.Fatalf("second action = %#v, want MoveCaretAction", result.Actions[1])
	}
	if _, ok := result.Actions[2].(RenameAction); !ok {
		t.Fatalf("third action = %#v, want RenameAction", result.Actions[2])
	}
}
