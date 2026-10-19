package kustolang_test

import (
	"encoding/json"
	"testing"
	"unicode/utf16"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxDiagnosticMutationAndUTF16Collection(t *testing.T) {
	expression := kustolang.ParseExpression(`"😀" + value`)
	first := kustolang.NewDiagnostic("APP100", "first")
	second := kustolang.NewDiagnostic("APP101", "second").WithLocationKind(kustolang.DiagnosticLocationRelativeEnd)
	absolute := kustolang.NewDiagnostic("APP102", "absolute").WithElementLocation(expression)
	direct := kustolang.NewDiagnostic("APP103", "direct").WithLocation(7, 5)
	withFirst := kustolang.WithDiagnostics(expression, first)
	withBoth := kustolang.WithAdditionalDiagnostics(withFirst, second, absolute, direct)
	if len(expression.Diagnostics()) != 0 {
		t.Fatal("diagnostic mutation changed the original syntax")
	}
	diagnostics := kustolang.GetContainedSyntaxDiagnostics(withBoth)
	if len(diagnostics) != 4 || diagnostics[0].Code != "APP100" || diagnostics[1].Code != "APP101" || diagnostics[2].Code != "APP102" || diagnostics[3].Code != "APP103" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	expectedLength := len(utf16.Encode([]rune(withBoth.Text(kustolang.IncludeTriviaAll))))
	if diagnostics[0].LocationKind != kustolang.DiagnosticLocationAbsolute || diagnostics[0].Start != 0 || diagnostics[0].Length != expectedLength {
		t.Fatalf("relative UTF-16 location = %#v, rendered text %q has %d UTF-16 units and width %d", diagnostics[0], withBoth.Text(kustolang.IncludeTriviaAll), expectedLength, withBoth.Width())
	}
	if diagnostics[1].LocationKind != kustolang.DiagnosticLocationAbsolute || diagnostics[1].Start != expectedLength || diagnostics[1].Length != 0 {
		t.Fatalf("relative-end UTF-16 location = %#v", diagnostics[1])
	}
	if diagnostics[2].Start != absolute.Start || diagnostics[2].Length != absolute.Length || diagnostics[2].Length != expectedLength {
		t.Fatalf("absolute UTF-16 location was converted again: before=%#v after=%#v", absolute, diagnostics[2])
	}
	if diagnostics[3].Start != direct.Start || diagnostics[3].Length != direct.Length {
		t.Fatalf("direct UTF-16 location was converted again: before=%#v after=%#v", direct, diagnostics[3])
	}
	encoded, err := json.Marshal(direct)
	if err != nil {
		t.Fatal(err)
	}
	var decoded kustolang.Diagnostic
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	decodedResult := kustolang.GetContainedSyntaxDiagnostics(kustolang.WithDiagnostics(expression, decoded))
	if len(decodedResult) != 1 || decodedResult[0].Start != direct.Start || decodedResult[0].Length != direct.Length {
		t.Fatalf("JSON diagnostic was converted again: before=%#v after=%#v", decoded, decodedResult)
	}

	preserved := kustolang.WithDiagnostics(withBoth)
	if len(preserved.Diagnostics()) != 4 {
		t.Fatalf("empty replacement did not preserve diagnostics: %#v", preserved.Diagnostics())
	}
}
