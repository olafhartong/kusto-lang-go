package kustolang_test

// Tests for the public Parser<TInput, TOutput>.Cast and string Convert facade.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"strconv"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type publicParserCastValue int

func (v publicParserCastValue) Value() int { return int(v) }

func TestPublicResultParserCastPreservesTag(t *testing.T) {
	source := kustolang.NewResultPrimaryParser(
		func(kustolang.Source[byte], int) kustolang.ParseResult[publicParserCastValue] {
			return kustolang.NewParseResult(0, publicParserCastValue(42))
		},
		kustolang.WithParserTag[byte]("answer"),
	)

	cast := kustolang.ResultParserCast[byte, publicParserCastValue, interface{ Value() int }](source)
	result := cast.ParseValue(kustolang.NewTextSource(""), 0)
	if cast.Tag() != "answer" || !result.Succeeded() || result.Value.Value() != 42 {
		t.Fatalf("public cast tag/result = %q/%#v", cast.Tag(), result)
	}
}

func TestPublicConvertStringFacade(t *testing.T) {
	parser := kustolang.ConvertString(kustolang.MatchText("123"), func(text string) int {
		value, err := strconv.Atoi(text)
		if err != nil {
			t.Fatalf("strconv.Atoi(%q): %v", text, err)
		}
		return value
	})

	result := parser.ParseValue(kustolang.NewArraySource([]byte("x123y")), 1)
	if !result.Succeeded() || result.Length != 3 || result.Value != 123 {
		t.Fatalf("public ConvertString result = %#v", result)
	}
}

func TestPublicConvertUTF16StringFacade(t *testing.T) {
	parser := kustolang.ConvertUTF16String(kustolang.UTF16CharScannerChars("😀x"), func(text string) string {
		return text
	})
	units := kustolang.UTF16CodeUnits("_😀x_")
	result := parser.ParseValue(kustolang.NewArraySource(units), 1)
	if !result.Succeeded() || result.Length != 3 || result.Value != "😀x" {
		t.Fatalf("public ConvertUTF16String result = %#v", result)
	}
}
