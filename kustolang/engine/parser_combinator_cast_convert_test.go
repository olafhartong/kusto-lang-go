package kustolang

// Tests for Parser<TInput, TOutput>.Cast and the character-to-string Convert overload.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"strings"
	"testing"
)

type parserCastValue string

func (v parserCastValue) String() string { return string(v) }

func TestResultParserCastPreservesTagAndCastsAtRuntime(t *testing.T) {
	source := MatchValue(
		func(input byte) bool { return input == 'A' },
		func(byte) parserCastValue { return "alpha" },
	)
	tagged := asResultParser[byte, parserCastValue](source.WithTag("letter"))

	cast := ResultParserCast[byte, parserCastValue, interface{ String() string }](tagged)
	if cast.Tag() != "letter" || tagged.Tag() != "letter" {
		t.Fatalf("cast/source tags = %q/%q, want letter/letter", cast.Tag(), tagged.Tag())
	}
	if _, ok := cast.(*RuleParser[byte, interface{ String() string }]); !ok {
		t.Fatalf("cast parser type = %T, want RuleParser", cast)
	}

	result := cast.ParseValue(NewTextSource("A"), 0)
	if !result.Succeeded() || result.Length != 1 || result.Value.String() != "alpha" {
		t.Fatalf("cast result = %#v", result)
	}
}

func TestResultParserCastPanicsForIncompatibleOutput(t *testing.T) {
	parser := ConvertValue(MatchByte('A'), "alpha")
	cast := ResultParserCast[byte, string, int](parser)

	defer func() {
		if recover() == nil {
			t.Fatal("incompatible result cast did not panic")
		}
	}()
	_ = cast.ParseValue(NewTextSource("A"), 0)
}

func TestConvertStringUsesMatchedTextFromAllByteSources(t *testing.T) {
	parser := ConvertString(MatchText("Storm"), strings.ToUpper)
	if _, ok := parser.(*ConvertParser[byte, string]); !ok {
		t.Fatalf("ConvertString parser type = %T, want ConvertParser", parser)
	}

	tests := []struct {
		name   string
		source Source[byte]
		start  int
	}{
		{name: "text source", source: NewTextSource("xxStormyy"), start: 2},
		{name: "array source", source: NewArraySource([]byte("xxStormyy")), start: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parser.ParseValue(test.source, test.start)
			if !result.Succeeded() || result.Length != 5 || result.Value != "STORM" {
				t.Fatalf("ConvertString result = %#v", result)
			}
		})
	}
}
