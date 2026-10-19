package kustolang_test

// Tests for the public facade ported from
// src/Kusto.Language/Parser/Combinators/Describer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDescriberSequenceAlternationTagsAndHiddenParsers(t *testing.T) {
	a := kustolang.MatchText("A").WithTag("A")
	b := kustolang.MatchText("B").WithTag("B")
	c := kustolang.MatchText("C").WithTag("C")

	tests := []struct {
		name   string
		parser kustolang.Parser[byte]
		want   string
	}{
		{"sequence containing alternation", kustolang.Sequence(a, kustolang.First(b, c)), "A (B | C)"},
		{"alternation containing sequence", kustolang.First(kustolang.Sequence(a, b), c), "(A B) | C"},
		{"same separator remains unparenthesized", kustolang.Sequence(a, kustolang.Sequence(b, c)), "A B C"},
		{"hidden parser is omitted", kustolang.Sequence(a, b.Hide(), c), "A C"},
		{"tag replaces nested grammar", kustolang.Sequence(a, kustolang.Sequence(b, c).WithTag("pair")), "A pair"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := kustolang.Describe(test.parser); got != test.want {
				t.Fatalf("Describe() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPublicDescriberOptionalRepetitionAndRequiredParsers(t *testing.T) {
	a := kustolang.MatchText("A").WithTag("A")
	b := kustolang.MatchText("B").WithTag("B")
	c := kustolang.MatchText("C").WithTag("C")

	grammar := kustolang.Sequence(
		kustolang.ZeroOrOne(a),
		kustolang.ZeroOrMore(b),
		kustolang.OneOrMore(c),
	)
	if got := kustolang.Describe(grammar); got != "[A] {B} {C}+" {
		t.Fatalf("repetition description = %q", got)
	}

	value := kustolang.Produce(kustolang.Sequence(a, b), func([]any) string { return "AB" })
	required := kustolang.RequiredValue(value, func() string { return "missing" })
	if got := kustolang.Describe(required); got != "(A B)!" {
		t.Fatalf("required description = %q", got)
	}
	if got := kustolang.Describe(required, false); got != "A B" {
		t.Fatalf("required description without marker = %q", got)
	}
	if got := kustolang.Describe(required, true); got != "(A B)!" {
		t.Fatalf("required description with explicit marker = %q", got)
	}
}

func TestPublicDescriberTermsWrappersAndApplyModes(t *testing.T) {
	a := kustolang.MatchText("A").WithTag("A")
	b := kustolang.MatchText("B").WithTag("B")
	left := kustolang.ConvertValue(a, "A")
	right := kustolang.RuleRight1(kustolang.ConvertValue(b, "B"), func(left, right string) string {
		return left + right
	})

	tests := []struct {
		name   string
		parser kustolang.Parser[byte]
		want   string
	}{
		{"not", kustolang.Not(a), "not(A)"},
		{"fails", kustolang.Fails(a), "fails(A)"},
		{"match", kustolang.MatchText("untagged"), "match()"},
		{"map", kustolang.MapValues([]kustolang.MapValueEntry[byte, string]{{Key: []byte("A"), Value: "A"}}), "map()"},
		{"forward", kustolang.ForwardValue(left), "forward()"},
		{"convert", left, "A"},
		{"produce", kustolang.Produce(kustolang.Sequence(a, b), func([]any) string { return "AB" }), "A B"},
		{"if", kustolang.If(kustolang.MatchText("test").WithTag("test"), b), "B"},
		{"limit", kustolang.Limit(kustolang.MatchText("limit").WithTag("limit"), left), "A"},
		{"apply once", kustolang.Apply(left, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] { return right }), "A B"},
		{"apply optional", kustolang.ApplyOptional(left, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] { return right }), "A [B]"},
		{"apply zero or more", kustolang.ApplyZeroOrMore(left, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] { return right }), "A {B}"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := kustolang.Describe(test.parser); got != test.want {
				t.Fatalf("Describe() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPublicDescriberNilAndCompatibilitySpelling(t *testing.T) {
	if got := kustolang.Describe[byte](nil); got != "" {
		t.Fatalf("Describe(nil) = %q", got)
	}

	parser := kustolang.Sequence(
		kustolang.MatchText("A").WithTag("A"),
		kustolang.MatchText("B").WithTag("B"),
	)
	if got, want := kustolang.DescribeParser(parser), kustolang.Describe(parser); got != want {
		t.Fatalf("DescribeParser() = %q, Describe() = %q", got, want)
	}
}
