package kustolang_test

// Tests for the public EditString and client-directive facades ported from
// Kusto.Language.Editor. See third_party/kusto/LICENSE.

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicEditStringTracksUTF16CollectiveChanges(t *testing.T) {
	original := kustolang.NewEditString("a😀bc")
	if original.OriginalText() != "a😀bc" || original.CurrentText() != "a😀bc" || original.Length() != 5 {
		t.Fatalf("original edit string = %q/%q length %d", original.OriginalText(), original.CurrentText(), original.Length())
	}
	if original.At(1) != 0xd83d || original.At(2) != 0xde00 || original.IndexOf("b") != 3 {
		t.Fatalf("UTF-16 indexing = %#x/%#x/%d", original.At(1), original.At(2), original.IndexOf("b"))
	}

	edited := original.ReplaceAt(1, 2, "X").Insert(2, "YZ")
	if edited.CurrentText() != "aXYZbc" || edited.OriginalText() != original.OriginalText() {
		t.Fatalf("edited text = %q from %q", edited.CurrentText(), edited.OriginalText())
	}
	changes := edited.GetChanges()
	if !reflect.DeepEqual(changes, []kustolang.TextEdit{{Start: 1, DeleteLength: 2, InsertText: "XYZ"}}) {
		t.Fatalf("collective changes = %#v", changes)
	}
	if replayed := original.ApplyAll(changes); replayed.CurrentText() != edited.CurrentText() {
		t.Fatalf("replayed text = %q, want %q", replayed.CurrentText(), edited.CurrentText())
	}
	if edited.GetCurrentPosition(3) != 4 || edited.GetOriginalPosition(4) != 3 {
		t.Fatalf("position mapping = current %d original %d", edited.GetCurrentPosition(3), edited.GetOriginalPosition(4))
	}
	if rng := edited.GetChangeRange(); rng.Start != 1 || rng.Length != 3 {
		t.Fatalf("change range = %#v", rng)
	}
}

func TestPublicEditStringApplyAllValidationAndSequentialChanges(t *testing.T) {
	value := kustolang.NewEditString("a😀bc")
	edits := []kustolang.TextEdit{
		kustolang.NewTextEdit(1, 2, "X"),
		kustolang.NewTextEdit(4, 1, "!"),
	}
	if !value.CanApplyAll(edits) {
		t.Fatal("valid non-overlapping UTF-16 edits were rejected")
	}
	if got := value.ApplyAll(edits).CurrentText(); got != "aXb!" {
		t.Fatalf("ApplyAll text = %q", got)
	}

	sequential := kustolang.ConvertToSequentialChanges(edits)
	if !reflect.DeepEqual(sequential, []kustolang.TextEdit{
		{Start: 1, DeleteLength: 2, InsertText: "X"},
		{Start: 3, DeleteLength: 1, InsertText: "!"},
	}) {
		t.Fatalf("sequential changes = %#v", sequential)
	}
	if value.CanApplyAll([]kustolang.TextEdit{{Start: 1, DeleteLength: 3}, {Start: 2, DeleteLength: 1}}) {
		t.Fatal("overlapping edits were accepted")
	}
	if value.CanApplyAll([]kustolang.TextEdit{{Start: 2, DeleteLength: 0, InsertText: "split"}}) {
		t.Fatal("edit splitting a surrogate pair was accepted")
	}
}

func TestPublicClientDirectiveParsingAndMalformedRecovery(t *testing.T) {
	text := kustolang.NewEditString("#set name=\"😀\", count=42 missing=\n#client 'unterminated\nT | where (\n")
	info := kustolang.ParseClientDirectiveInfo(text)
	directives := info.Directives()
	if len(directives) != 2 || directives[0].Name() != "set" || directives[1].Name() != "client" {
		t.Fatalf("directives = %#v", directives)
	}
	arguments := directives[0].Arguments()
	if len(arguments) != 3 || arguments[0].Name() != "name" || arguments[0].Text() != "\"😀\"" || arguments[0].Value() != "😀" {
		t.Fatalf("first directive arguments = %#v", arguments)
	}
	if arguments[1].Name() != "count" || arguments[1].Value() != int64(42) {
		t.Fatalf("numeric argument = %#v", arguments[1])
	}
	if arguments[2].Name() != "missing" || arguments[2].Value() != nil {
		t.Fatalf("missing argument = %#v", arguments[2])
	}
	arguments[0] = kustolang.NewClientDirectiveArgument("changed", "changed", "changed")
	directives[0] = kustolang.ClientDirective{}
	if info.Directives()[0].Arguments()[0].Name() != "name" {
		t.Fatal("public directive slices mutated the underlying directive info")
	}
	malformedArguments := directives[1].Arguments()
	if len(malformedArguments) != 1 || malformedArguments[0].Value() != "unterminated" {
		t.Fatalf("unclosed string recovery = %#v", malformedArguments)
	}
	if got := info.RemainingText().CurrentText(); got != "T | where (" {
		t.Fatalf("remaining malformed KQL = %q", got)
	}

	service := kustolang.NewCodeService(info.RemainingText().CurrentText())
	if diagnostics := service.GetDiagnostics(true, kustolang.CancellationToken{}); len(diagnostics) == 0 {
		t.Fatal("malformed KQL remaining after directives produced no diagnostics")
	}
	if _, ok := kustolang.TryParseClientDirective(kustolang.NewEditString("T | count")); ok {
		t.Fatal("ordinary KQL was accepted as a client directive")
	}
}
