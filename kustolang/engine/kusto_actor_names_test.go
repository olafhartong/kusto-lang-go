package kustolang

import "testing"

func TestKustoActorNamesMirrorOfficialActorTypes(t *testing.T) {
	want := []string{
		"InlineDatabaseFunctionActor",
		"ExtractExpressionActor",
		"ConvertToMacroExpandActor",
		"RenameActor",
		"DiagnosticFixActor",
	}
	got := KustoActorNamesAll()
	if len(got) != len(want) {
		t.Fatalf("actor names = %#v", got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("actor name %d = %q, want %q", index, got[index], want[index])
		}
	}
	got[0] = "mutated"
	if KustoActorNamesAll()[0] != KustoActorNameFunctionInliner {
		t.Fatal("actor names list is not defensive")
	}
}
