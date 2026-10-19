package kustolang_test

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestKustoCodePublicWrapperPreservesIdentity(t *testing.T) {
	code := kustolang.ParseAndAnalyze("print value=1")
	if code.Analyze() != code {
		t.Fatal("Analyze did not preserve public wrapper identity")
	}
	if code.AnalyzeWithCancellation(nil, kustolang.CancellationToken{}) != code {
		t.Fatal("AnalyzeWithCancellation did not preserve public wrapper identity")
	}
	if code.WithGlobals(code.Globals) != code {
		t.Fatal("WithGlobals did not preserve public wrapper identity")
	}
	if code.WithGlobalsWithCancellation(code.Globals, kustolang.CancellationToken{}) != code {
		t.Fatal("WithGlobalsWithCancellation did not preserve public wrapper identity")
	}
	if code.WithText(code.Text) != code {
		t.Fatal("WithText did not preserve public wrapper identity")
	}
	if code.WithTextWithCancellation(code.Text, kustolang.CancellationToken{}) != code {
		t.Fatal("WithTextWithCancellation did not preserve public wrapper identity")
	}
	replaced := code.WithText("print other=2")
	if replaced == code || replaced.Text != "print other=2" || !replaced.HasSemantics() {
		t.Fatalf("WithText replacement = same:%t text:%q semantics:%t", replaced == code, replaced.Text, replaced.HasSemantics())
	}
}

func TestKustoCodeFacadeTextMutationDoesNotDesynchronizeMethods(t *testing.T) {
	const text = `datatable(A:long)[1] | where "😀" != "" and Missing > A`
	code := kustolang.ParseAndAnalyze(text)
	position := len(utf16.Encode([]rune(text[:strings.LastIndex(text, "A")])))

	diagnostics := code.GetDiagnostics()
	tokenIndex := code.GetTokenIndex(position)
	columns := code.GetColumnsInScope(position)
	if len(diagnostics) == 0 || columns == nil || columns.GetColumn("A") == nil {
		t.Fatalf("test setup failed: diagnostics=%#v columns=%#v", diagnostics, columns)
	}

	code.Text = "mutated facade text"
	if got := code.GetDiagnostics(); !reflect.DeepEqual(got, diagnostics) {
		t.Fatalf("diagnostics changed after facade Text mutation:\ngot:  %#v\nwant: %#v", got, diagnostics)
	}
	if got := code.GetTokenIndex(position); got != tokenIndex {
		t.Fatalf("token index after facade Text mutation = %d, want %d", got, tokenIndex)
	}
	if got := code.GetColumnsInScope(position); got == nil || got.GetColumn("A") == nil {
		t.Fatalf("columns after facade Text mutation = %#v", got)
	}
}

func TestKustoCodeTryGetLineAndOffsetPreservesUTF16SurrogatePositions(t *testing.T) {
	const text = "// 😀\nT | where ("
	code := kustolang.ParseAndAnalyze(text)

	tests := []struct {
		name       string
		position   int
		line       int
		lineOffset int
	}{
		{name: "before surrogate", position: 3, line: 1, lineOffset: 4},
		{name: "between surrogate code units", position: 4, line: 1, lineOffset: 5},
		{name: "after surrogate", position: 5, line: 1, lineOffset: 6},
		{name: "malformed second line end", position: len(utf16.Encode([]rune(text))), line: 2, lineOffset: len(utf16.Encode([]rune("T | where ("))) + 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			line, lineOffset, ok := code.TryGetLineAndOffset(test.position)
			if !ok || line != test.line || lineOffset != test.lineOffset {
				t.Fatalf("line/offset at UTF-16 position %d = %d/%d/%t, want %d/%d/true",
					test.position, line, lineOffset, ok, test.line, test.lineOffset)
			}
		})
	}

	for _, position := range []int{-1, len(utf16.Encode([]rune(text))) + 1} {
		if line, lineOffset, ok := code.TryGetLineAndOffset(position); ok {
			t.Fatalf("out-of-range UTF-16 position %d = %d/%d/true", position, line, lineOffset)
		}
	}
}
