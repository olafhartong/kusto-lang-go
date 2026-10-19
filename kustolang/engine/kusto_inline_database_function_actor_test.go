package kustolang

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestInlineDatabaseFunctionActorMenusAndInlineOnce(t *testing.T) {
	globals, _, function, _ := inlineActorGlobals()
	text := "print A=F(), B=G()"
	code := ParseAndAnalyze(text, globals)
	actor := NewInlineDatabaseFunctionActor()
	position := strings.Index(text, "F")

	menu := inlineRequireMenu(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}))
	if len(menu.Actions) != 4 {
		t.Fatalf("menu actions = %#v, want inline once/recursive/all/all-recursive", menu.Actions)
	}
	for index, kind := range []string{InlineOnceKind, InlineRecursiveKind, InlineAllOnceKind, InlineAllRecursiveKind} {
		if menu.Actions[index].Base().Kind != kind {
			t.Fatalf("menu action %d kind = %q, want %q", index, menu.Actions[index].Base().Kind, kind)
		}
	}

	action := menu.Actions[0].(ApplyAction)
	result := actor.ApplyAction(code, action, position, CodeActionOptionsDefault, CancellationToken{})
	changed := inlineRequireChange(t, result)
	want := "let F = () {1};\nprint A=F(), B=G()"
	if changed.ChangedText != want {
		t.Fatalf("inline once text = %q, want %q", changed.ChangedText, want)
	}
	if replayed := NewEditString(text).ApplyAll(changed.Changes).CurrentText; replayed != want {
		t.Fatalf("replayed edits = %q, want %q", replayed, want)
	}
	if function.SymbolName() != "F" {
		t.Fatal("test setup function identity changed")
	}
}

func TestInlineDatabaseFunctionActorRequalifiesBodyAndDequalifiesCalls(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	function := NewComputedFunctionSymbol("RemoteF", "T | project X", TabularityTabular)
	remoteDB := NewDatabaseSymbol("Remote", table, function)
	localDB := NewDatabaseSymbol("Local")
	localCluster := NewClusterSymbol("local.kusto.windows.net", localDB)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", remoteDB)
	globals := NewGlobalState(localDB).
		WithClusterList([]*ClusterSymbol{localCluster, remoteCluster}).
		WithCluster(localCluster).
		WithDatabase(localDB)
	text := "cluster('remote.kusto.windows.net').database('Remote').RemoteF()"
	code := ParseAndAnalyze(text, globals)
	actor := NewInlineDatabaseFunctionActor()
	position := strings.LastIndex(text, "RemoteF")
	menu := inlineRequireMenu(t, actor.GetActions(code, position, 0, 0, CodeActionOptionsDefault, CancellationToken{}))
	result := actor.ApplyAction(code, menu.Actions[0].(ApplyAction), position, CodeActionOptionsDefault, CancellationToken{})
	changed := inlineRequireChange(t, result)
	wantBody := "cluster('remote.kusto.windows.net').database('Remote').T | project X"
	if !strings.Contains(changed.ChangedText, wantBody) {
		t.Fatalf("qualified body missing from %q", changed.ChangedText)
	}
	if !strings.HasSuffix(changed.ChangedText, "\nRemoteF()") {
		t.Fatalf("qualified call was not dequalified: %q", changed.ChangedText)
	}
}

func TestInlineDatabaseFunctionActorUnavailableMalformedCancellationAndUTF16(t *testing.T) {
	unavailable := NewScalarFunctionSymbol("Unavailable", ScalarTypes.Long.Type())
	database := NewDatabaseSymbol("db", unavailable)
	cluster := NewClusterSymbol("current.kusto.windows.net", database)
	globals := NewGlobalState(database).WithClusterList([]*ClusterSymbol{cluster}).WithCluster(cluster).WithDatabase(database)
	text := "// \U0001f600\nprint Unavailable()"
	code := ParseAndAnalyze(text, globals)
	actor := NewInlineDatabaseFunctionActor()
	bytePosition := strings.Index(text, "Unavailable")
	utfPosition := len(utf16.Encode([]rune(text[:bytePosition])))
	menu := inlineRequireMenu(t, actor.GetActionsUTF16(code, utfPosition, 0, 0, CodeActionOptionsDefault, CancellationToken{}))
	action := menu.Actions[0].(ApplyAction)
	if action.ApplyData()[0] != strconv.Itoa(utfPosition) {
		t.Fatalf("UTF-16 action data = %#v, want %d", action.ApplyData(), utfPosition)
	}
	result := actor.ApplyActionUTF16(code, action, utfPosition, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "Could not access definition of referenced function" {
		t.Fatalf("unavailable body failure = %q", result.FailureReason)
	}

	for _, malformed := range []ApplyAction{
		NewCodeAction(InlineOnceKind, "", "", nil),
		NewCodeAction(InlineOnceKind, "", "", []string{"bad"}),
		NewCodeAction(InlineAllOnceKind, "", "", []string{"0"}),
		NewCodeAction(InlineAllOnceKind, "", "", []string{"0", "999"}),
		NewCodeAction("Unknown", "", "", []string{"0"}),
	} {
		got := actor.ApplyAction(code, malformed, 0, CodeActionOptionsDefault, CancellationToken{})
		if got.FailureReason == "" || len(got.Actions) != 0 {
			t.Fatalf("malformed payload result = %#v", got)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("canceled actor call did not panic")
		}
	}()
	actor.GetActions(code, bytePosition, 0, 0, CodeActionOptionsDefault, NewCancellationToken(ctx))
}

func inlineActorGlobals() (*GlobalState, *DatabaseSymbol, *FunctionSymbol, *FunctionSymbol) {
	f := NewComputedFunctionSymbol("F", "1", TabularityScalar)
	g := NewComputedFunctionSymbol("G", "2", TabularityScalar)
	database := NewDatabaseSymbol("db", f, g)
	cluster := NewClusterSymbol("current.kusto.windows.net", database)
	globals := NewGlobalState(database).WithClusterList([]*ClusterSymbol{cluster}).WithCluster(cluster).WithDatabase(database)
	return globals, database, f, g
}

func inlineRequireMenu(t *testing.T, info CodeActionInfo) *MenuAction {
	t.Helper()
	if len(info.Actions) != 1 {
		t.Fatalf("actions = %#v, want one menu", info.Actions)
	}
	menu, ok := info.Actions[0].(*MenuAction)
	if !ok {
		t.Fatalf("action = %T, want *MenuAction", info.Actions[0])
	}
	return menu
}

func inlineRequireChange(t *testing.T, result CodeActionResult) ChangeTextAction {
	t.Helper()
	if result.FailureReason != "" {
		t.Fatalf("action failed: %s", result.FailureReason)
	}
	for _, action := range result.Actions {
		if change, ok := action.(ChangeTextAction); ok {
			return change
		}
	}
	t.Fatalf("result = %#v, want change action", result)
	return ChangeTextAction{}
}
