package kustolang_test

// External-package parity tests for the pinned ScriptFacts API.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicScriptFactsFindsUTF16KustoBlocks(t *testing.T) {
	text := "let face = \"😀\";\n\nlet s = ```one\n\ninside```;\n\nT\n| count"
	lineStarts := parserFactsTestLineStarts(text)
	letS := parserFactsTestUTF16Index(text, "let s")
	inside := parserFactsTestUTF16Index(text, "inside```")
	table := parserFactsTestUTF16Index(text, "T\n| count")

	separated := kustolang.ScriptFacts.GetLineSeparatedBlockStarts(text, lineStarts)
	if want := []int{0, letS, inside, table}; !reflect.DeepEqual(separated, want) {
		t.Fatalf("GetLineSeparatedBlockStarts() = %#v, want %#v", separated, want)
	}

	got := kustolang.ScriptFacts.GetKustoBlockStarts(text, lineStarts)
	if want := []int{0, letS, table}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetKustoBlockStarts() = %#v, want %#v", got, want)
	}
	if table == strings.Index(text, "T\n| count") {
		t.Fatal("test did not distinguish UTF-16 positions from UTF-8 byte offsets")
	}
}

func TestPublicScriptFactsReturnsUTF16InvalidRanges(t *testing.T) {
	text := "let face = \"😀\"; // note 😀\r\nT"
	ranges := kustolang.ScriptFacts.GetInvalidKustoBlockStartRanges(text)
	if len(ranges) != 2 {
		t.Fatalf("GetInvalidKustoBlockStartRanges() = %#v, want string and comment ranges", ranges)
	}

	stringStart := parserFactsTestUTF16Index(text, "\"😀\"")
	if got, want := ranges[0], kustolang.NewTextRange(stringStart, 4); got != want {
		t.Fatalf("string range = %#v, want %#v", got, want)
	}
	commentStart := parserFactsTestUTF16Index(text, "// note")
	commentEnd := parserFactsTestUTF16Index(text, "T")
	if got, want := ranges[1], kustolang.NewTextRange(commentStart, commentEnd-commentStart); got != want {
		t.Fatalf("comment range = %#v, want %#v", got, want)
	}
	blockStarts := []int{
		ranges[1].End(), ranges[1].Start + 1, ranges[1].Start,
		ranges[0].End(), ranges[0].Start + 1, ranges[0].Start,
	}
	wantStarts := []int{ranges[0].Start, ranges[0].End(), ranges[1].Start, ranges[1].End()}
	if got := kustolang.ScriptFacts.RemoveInvalidKustoBlockStarts(blockStarts, text); !reflect.DeepEqual(got, wantStarts) {
		t.Fatalf("RemoveInvalidKustoBlockStarts() = %#v, want %#v", got, wantStarts)
	}

	if got := kustolang.ScriptFacts.GetInvalidKustoBlockStartRanges("T | count"); got == nil || len(got) != 0 {
		t.Fatalf("range-free result = %#v, want non-nil empty slice", got)
	}
}

func TestPublicScriptFactsKeepsMalformedMultilineStringInOneBlock(t *testing.T) {
	text := "let s = ```😀\n\nstill string"
	lineStarts := parserFactsTestLineStarts(text)

	if got, want := kustolang.ScriptFacts.GetKustoBlockStarts(text, lineStarts), []int{0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unterminated multiline blocks = %#v, want %#v", got, want)
	}
	ranges := kustolang.ScriptFacts.GetInvalidKustoBlockStartRanges(text)
	if len(ranges) != 1 || ranges[0].Start != parserFactsTestUTF16Index(text, "```") || ranges[0].End() != parserFactsTestUTF16Length(text) {
		t.Fatalf("unterminated multiline range = %#v", ranges)
	}
}

func TestPublicScriptFactsSortsDeduplicatesAndUsesStrictRangeInteriors(t *testing.T) {
	input := []int{9, 3, 5, 0, 3, 6}
	original := append([]int(nil), input...)
	if got, want := kustolang.ScriptFacts.RemoveDuplicateBlockStarts(input), []int{0, 3, 5, 6, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("RemoveDuplicateBlockStarts() = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("RemoveDuplicateBlockStarts mutated input: %#v", input)
	}

	ranges := []kustolang.TextRange{kustolang.NewTextRange(3, 3)}
	if got, want := kustolang.ScriptFacts.RemoveInvalidBlockStarts(input, ranges), []int{0, 3, 3, 6, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("RemoveInvalidBlockStarts() = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("RemoveInvalidBlockStarts mutated input: %#v", input)
	}

	if got := kustolang.ScriptFacts.RemoveDuplicateBlockStarts(nil); got == nil || len(got) != 0 {
		t.Fatalf("empty duplicate result = %#v, want non-nil empty slice", got)
	}
}

func TestPublicScriptFactsDefendsAgainstInvalidUTF16LineStarts(t *testing.T) {
	text := "😀\n\nT"
	for _, starts := range [][]int{nil, {0, 1}, {0, -1}, {0, 100}} {
		if got, want := kustolang.ScriptFacts.GetLineSeparatedBlockStarts(text, starts), []int{0}; !reflect.DeepEqual(got, want) {
			t.Fatalf("GetLineSeparatedBlockStarts(%v) = %#v, want %#v", starts, got, want)
		}
	}
}

func parserFactsTestLineStarts(text string) []int {
	byteStarts := kustolang.GetLineStarts(text)
	starts := make([]int, len(byteStarts))
	for i, start := range byteStarts {
		starts[i] = parserFactsTestUTF16Length(text[:start])
	}
	return starts
}

func parserFactsTestUTF16Index(text, substring string) int {
	byteIndex := strings.Index(text, substring)
	if byteIndex < 0 {
		return -1
	}
	return parserFactsTestUTF16Length(text[:byteIndex])
}

func parserFactsTestUTF16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}
