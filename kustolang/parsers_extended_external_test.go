package kustolang_test

// Tests for the public facade ported from Kusto.Language parser combinators.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicRule4ThroughRule9(t *testing.T) {
	a := kustolang.TextLiteral("A")
	b := kustolang.TextLiteral("B")
	c := kustolang.TextLiteral("C")
	d := kustolang.TextLiteral("D")
	e := kustolang.TextLiteral("E")
	f := kustolang.TextLiteral("F")
	g := kustolang.TextLiteral("G")
	h := kustolang.TextLiteral("H")
	i := kustolang.TextLiteral("I")

	tests := []struct {
		name   string
		parser kustolang.ResultParser[byte, string]
		input  string
	}{
		{"Rule4", kustolang.Rule4(a, b, c, d, func(v1, v2, v3, v4 string) string { return v1 + v2 + v3 + v4 }), "ABCD"},
		{"Rule5", kustolang.Rule5(a, b, c, d, e, func(v1, v2, v3, v4, v5 string) string { return v1 + v2 + v3 + v4 + v5 }), "ABCDE"},
		{"Rule6", kustolang.Rule6(a, b, c, d, e, f, func(v1, v2, v3, v4, v5, v6 string) string { return v1 + v2 + v3 + v4 + v5 + v6 }), "ABCDEF"},
		{"Rule7", kustolang.Rule7(a, b, c, d, e, f, g, func(v1, v2, v3, v4, v5, v6, v7 string) string { return v1 + v2 + v3 + v4 + v5 + v6 + v7 }), "ABCDEFG"},
		{"Rule8", kustolang.Rule8(a, b, c, d, e, f, g, h, func(v1, v2, v3, v4, v5, v6, v7, v8 string) string { return v1 + v2 + v3 + v4 + v5 + v6 + v7 + v8 }), "ABCDEFGH"},
		{"Rule9", kustolang.Rule9(a, b, c, d, e, f, g, h, i, func(v1, v2, v3, v4, v5, v6, v7, v8, v9 string) string {
			return v1 + v2 + v3 + v4 + v5 + v6 + v7 + v8 + v9
		}), "ABCDEFGHI"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := kustolang.ParseTextValue(test.parser, test.input)
			if !result.Succeeded() || result.Length != len(test.input) || result.Value != test.input {
				t.Fatalf("parse result = %#v", result)
			}
		})
	}

	rule9 := tests[len(tests)-1].parser
	if result := kustolang.ParseTextValue(rule9, "ABCDEFGX"); result.Succeeded() {
		t.Fatalf("malformed Rule9 input succeeded: %#v", result)
	}
}

func TestPublicListCombinatorsRecoverMissingElementsAndSeparators(t *testing.T) {
	element := publicByteValue('A', "a")
	separator := publicByteValue(',', ",")
	end := kustolang.MatchByte(';')

	elements := kustolang.ListElementsWithMissing(
		element,
		func(kustolang.Source[byte], int) string { return "<element>" },
		true,
		func(values []string) string { return strings.Join(values, "") },
	)
	if result := elements.ParseValue(kustolang.NewTextSource("x"), 0); result.Length != 0 || result.Value != "<element>" {
		t.Fatalf("missing first element result = %#v", result)
	}

	list := kustolang.ListSeparatedFull(
		element,
		separator,
		element,
		func(kustolang.Source[byte], int) string { return "<element>" },
		func(kustolang.Source[byte], int) string { return "<separator>" },
		func(kustolang.Source[byte], int) string { return "<element>" },
		end,
		true,
		false,
		joinPublicSeparated,
	)
	if result := list.ParseValue(kustolang.NewTextSource("A,,A;"), 0); result.Length != 4 || result.Value != "a,<element>,a" {
		t.Fatalf("missing list element recovery = %#v", result)
	}
	if result := list.ParseValue(kustolang.NewTextSource("AA;"), 0); result.Length != 2 || result.Value != "a<separator>a" {
		t.Fatalf("missing list separator recovery = %#v", result)
	}
}

func TestPublicOListAndOListFull(t *testing.T) {
	elementA := publicByteValue('A', "a")
	elementB := publicByteValue('B', "b")
	separator := publicByteValue(',', ",")

	list := kustolang.OList(elementA, separator, nil, nil, nil, true, false, joinPublicOutput)
	if result := list.ParseValue(kustolang.NewTextSource("A,A"), 0); result.Length != 3 || result.Value != "a,a" {
		t.Fatalf("OList result = %#v", result)
	}

	full := kustolang.OListFull(elementA, separator, elementB, nil, nil, nil, nil, true, false, joinPublicOutput)
	if result := full.ParseValue(kustolang.NewTextSource("A,B"), 0); result.Length != 3 || result.Value != "a,b" {
		t.Fatalf("OListFull result = %#v", result)
	}
}

func TestPublicSafeAndPartialParsing(t *testing.T) {
	a := kustolang.TextLiteral("A")
	b := kustolang.TextLiteral("B")
	c := kustolang.TextLiteral("C")
	partial := kustolang.Sequence(a, b, c)

	validSource := kustolang.NewTextSource("ABC")
	if got := kustolang.ScanSafe(partial, validSource, 0); got != 3 {
		t.Fatalf("ScanSafe length = %d, want 3", got)
	}
	safeOutput := []any{"prefix"}
	if got := kustolang.ParseSafe(partial, validSource, 0, &safeOutput, 1); got != 3 || !reflect.DeepEqual(safeOutput, []any{"prefix", "A", "B", "C"}) {
		t.Fatalf("ParseSafe result = %d/%#v", got, safeOutput)
	}

	malformedSource := kustolang.NewTextSource("ABx")
	if got := kustolang.ScanPartial(partial, malformedSource, 0); got != 2 {
		t.Fatalf("ScanPartial length = %d, want 2", got)
	}
	failures := 0
	partialOutput := []any{}
	if got := kustolang.ParsePartial(partial, malformedSource, 0, &partialOutput, 0, func(_ kustolang.Parser[byte], output *[]any) {
		failures++
		*output = append(*output, "<missing>")
	}); got != 2 || failures != 1 || !reflect.DeepEqual(partialOutput, []any{"A", "B", "<missing>"}) {
		t.Fatalf("ParsePartial result = %d/%#v, failures=%d", got, partialOutput, failures)
	}

	complete := kustolang.Sequence(a, b)
	parsers := []kustolang.Parser[byte]{partial, complete}
	if got := kustolang.ScanPartialBest(parsers, malformedSource, 0); got != 2 {
		t.Fatalf("ScanPartialBest length = %d, want 2", got)
	}
	bestOutput := []any{}
	bestParsers := []kustolang.Parser[byte]{}
	if got := kustolang.ParsePartialBest(parsers, malformedSource, 0, &bestOutput, 0, &bestParsers); got != 2 || !reflect.DeepEqual(bestOutput, []any{"A", "B"}) || len(bestParsers) != 2 {
		t.Fatalf("ParsePartialBest result = %d/%#v, best=%d", got, bestOutput, len(bestParsers))
	}
}

func TestPublicParserVisitors(t *testing.T) {
	a := kustolang.TextLiteral("A")
	b := kustolang.TextLiteral("B")
	rule := kustolang.Rule4(a, b, a, b, func(v1, v2, v3, v4 string) string { return v1 + v2 + v3 + v4 })

	children := kustolang.AcceptParserVisitorResult(rule, kustolang.ParserVisitorResultFuncs[byte, int]{
		Rule: func(_ kustolang.Parser[byte], parsers []kustolang.Parser[byte]) int { return len(parsers) },
	})
	if children != 4 || !kustolang.IsParserParent(rule, a) {
		t.Fatalf("rule visitor/parent result = %d/%t", children, kustolang.IsParserParent(rule, a))
	}

	apply := kustolang.Apply(a, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] {
		return kustolang.RuleRight1(b, func(left, right string) string { return left + right })
	})
	visited := false
	kustolang.AcceptParserVisitor(apply, kustolang.ParserVisitorFuncs[byte]{
		Apply: func(_ kustolang.Parser[byte], _ kustolang.Parser[byte], _ kustolang.Parser[byte], kind kustolang.ApplyKind) {
			visited = kind == kustolang.ApplyKindOne
		},
	})
	if !visited {
		t.Fatal("apply visitor did not expose ApplyKindOne")
	}
}

func publicByteValue(literal byte, value string) kustolang.ResultParser[byte, string] {
	return kustolang.MatchValue(
		func(input byte) bool { return input == literal },
		func(byte) string { return value },
	)
}

func joinPublicOutput(values []any) string {
	var output strings.Builder
	for _, value := range values {
		output.WriteString(value.(string))
	}
	return output.String()
}

func joinPublicSeparated(values []kustolang.ElementAndSeparator[string, string]) string {
	var output strings.Builder
	for _, value := range values {
		output.WriteString(value.Element)
		output.WriteString(value.Separator)
	}
	return output.String()
}
