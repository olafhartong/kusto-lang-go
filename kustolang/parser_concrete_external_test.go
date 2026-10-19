package kustolang_test

// Tests for concrete parser facades ported from Kusto.Language combinators.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicConcreteParserFactoryTypes(t *testing.T) {
	matchA := kustolang.MatchByte('A')
	if _, ok := matchA.(*kustolang.MatchParser[byte]); !ok {
		t.Fatalf("MatchByte type = %T", matchA)
	}

	valueA := kustolang.MatchValue(func(input byte) bool { return input == 'A' }, func(byte) string { return "A" })
	if _, ok := valueA.(*kustolang.MatchValueParser[byte, string]); !ok {
		t.Fatalf("MatchValue type = %T", valueA)
	}

	tests := []struct {
		name string
		ok   bool
	}{
		{"SequenceParser", isSequenceParser(kustolang.Sequence(matchA, matchA))},
		{"FirstParser", isFirstParser(kustolang.First(matchA, matchA))},
		{"BestParser", isBestParser(kustolang.Best(matchA, matchA))},
		{"FailsParser", isFailsParser(kustolang.Fails(matchA))},
		{"NotParser", isNotParser(kustolang.Not(matchA))},
		{"ZeroOrMoreParser", isZeroOrMoreParser(kustolang.ZeroOrMore(matchA))},
		{"OneOrMoreParser", isOneOrMoreParser(kustolang.OneOrMore(matchA))},
		{"IfParser", isIfParser(kustolang.If(matchA, matchA))},
		{"ConvertParser", isConvertParser(kustolang.ConvertValue(matchA, "A"))},
		{"FirstValueParser", isFirstValueParser(kustolang.FirstValue(valueA, valueA))},
		{"BestValueParser", isBestValueParser(kustolang.BestValue(valueA, valueA))},
		{"OptionalParser", isOptionalParser(kustolang.OptionalValue(valueA))},
		{"RequiredParser", isRequiredParser(kustolang.RequiredValue(valueA, func() string { return "missing" }))},
		{"LimitParser", isLimitParser(kustolang.Limit(matchA, valueA))},
		{"ForwardParser", isForwardParser(kustolang.ForwardValue(valueA))},
		{"MapParser", isMapParser(kustolang.MapValues([]kustolang.MapValueEntry[byte, string]{{Key: []byte{'A'}, Value: "A"}}))},
		{"ProduceParser", isProduceParser(kustolang.Produce(valueA, func(values []any) string { return values[0].(string) }))},
		{"RuleParser", isRuleParser(kustolang.Rule1(valueA, func(value string) string { return value }))},
		{"ApplyParser", isApplyParser(kustolang.Apply(valueA, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] {
			return kustolang.NewRightParser(valueA)
		}))},
		{"IfValueParser", isIfValueParser(kustolang.IfValue(matchA, valueA))},
	}

	for _, test := range tests {
		if !test.ok {
			t.Errorf("%s factory did not return its concrete parser type", test.name)
		}
	}
}

func TestPublicConcreteParserTraversalCloneAndFailure(t *testing.T) {
	matchA := kustolang.NewPredicateMatchParser(func(input byte) bool { return input == 'A' })
	matchB := kustolang.NewPredicateMatchParser(func(input byte) bool { return input == 'B' })
	sequence := kustolang.NewSequenceParser([]kustolang.Parser[byte]{matchA, matchB})

	if sequence.ChildParserCount() != 2 || sequence.ChildParser(0) != matchA || sequence.ChildParserIndex(matchB) != 1 {
		t.Fatalf("sequence children = %d/%T/%d", sequence.ChildParserCount(), sequence.ChildParser(0), sequence.ChildParserIndex(matchB))
	}
	visited := false
	kustolang.AcceptParserVisitor(sequence, kustolang.ParserVisitorFuncs[byte]{
		Sequence: func(parser kustolang.Parser[byte], children []kustolang.Parser[byte]) {
			_, visited = parser.(*kustolang.SequenceParser[byte])
			visited = visited && len(children) == 2 && children[1] == matchB
		},
	})
	if !visited {
		t.Fatal("visitor did not expose the concrete sequence and its children")
	}

	tagged := sequence.WithTag("pair")
	if _, ok := tagged.(*kustolang.SequenceParser[byte]); !ok || tagged.Tag() != "pair" || sequence.Tag() != "" {
		t.Fatalf("metadata clone = %T tag %q, original tag %q", tagged, tagged.Tag(), sequence.Tag())
	}

	output := []any{"prefix"}
	if length := sequence.ParseList(kustolang.NewTextSource("AX"), 0, &output, 1); length != -2 {
		t.Fatalf("malformed sequence length = %d, want -2", length)
	}
	if !reflect.DeepEqual(output, []any{"prefix"}) {
		t.Fatalf("malformed sequence leaked output: %#v", output)
	}
}

func TestPublicConcreteRequiredAndOptionalRecovery(t *testing.T) {
	valueA := kustolang.NewPredicateMatchValueParser(func(input byte) bool { return input == 'A' }, func(byte) string { return "A" })
	source := kustolang.NewTextSource("X")

	required := kustolang.NewRequiredParser(valueA, func(kustolang.Source[byte], int) string { return "<missing>" })
	if result := required.ParseValue(source, 0); result.Length != 0 || result.Value != "<missing>" {
		t.Fatalf("required malformed recovery = %#v", result)
	}

	optional := kustolang.NewOptionalParser(valueA, func() string { return "<absent>" })
	if result := optional.ParseValue(source, 0); result.Length != 0 || result.Value != "<absent>" {
		t.Fatalf("optional malformed recovery = %#v", result)
	}
}

func TestPublicPrimaryParserExtensionPoints(t *testing.T) {
	matchA := func(source kustolang.Source[byte], start int) int {
		if source.IsEnd(start) || source.Peek(start) != 'A' {
			return -1
		}
		return 1
	}
	resultParser := kustolang.NewResultPrimaryParser(
		func(source kustolang.Source[byte], start int) kustolang.ParseResult[string] {
			if length := matchA(source, start); length >= 0 {
				return kustolang.NewParseResult(length, "A")
			}
			return kustolang.NewParseResult(-1, "")
		},
		kustolang.WithParserTag[byte]("letter"),
		kustolang.WithParserTraits[byte](kustolang.ParserTraits{Match: true}),
		kustolang.WithParserScan[byte](matchA),
	)
	if resultParser.Tag() != "letter" || !resultParser.IsMatch() {
		t.Fatalf("result primary metadata = %q/%t", resultParser.Tag(), resultParser.IsMatch())
	}
	if result := kustolang.ParseTextValue[string](resultParser, "A"); result.Length != 1 || result.Value != "A" {
		t.Fatalf("result primary success = %#v", result)
	}
	if result := kustolang.ParseTextValue[string](resultParser, "X"); result.Succeeded() {
		t.Fatalf("result primary malformed input succeeded: %#v", result)
	}

	listParser := kustolang.NewListPrimaryParser[byte, string](
		func(source kustolang.Source[byte], start int, output *[]any, _ int) int {
			if length := matchA(source, start); length >= 0 {
				*output = append(*output, "A")
				return length
			}
			return -1
		},
		kustolang.WithParserScan[byte](matchA),
	)
	if result := kustolang.ParseTextValue[string](listParser, "A"); result.Length != 1 || result.Value != "A" {
		t.Fatalf("list primary success = %#v", result)
	}
	if result := kustolang.ParseTextValue[string](listParser, "X"); result.Succeeded() {
		t.Fatalf("list primary malformed input succeeded: %#v", result)
	}
}

func isSequenceParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.SequenceParser[byte])
	return ok
}
func isFirstParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.FirstParser[byte])
	return ok
}
func isBestParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.BestParser[byte])
	return ok
}
func isFailsParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.FailsParser[byte])
	return ok
}
func isNotParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.NotParser[byte])
	return ok
}
func isZeroOrMoreParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.ZeroOrMoreParser[byte])
	return ok
}
func isOneOrMoreParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.OneOrMoreParser[byte])
	return ok
}
func isIfParser(parser kustolang.Parser[byte]) bool {
	_, ok := parser.(*kustolang.IfParser[byte])
	return ok
}
func isConvertParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.ConvertParser[byte, string])
	return ok
}
func isFirstValueParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.FirstValueParser[byte, string])
	return ok
}
func isBestValueParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.BestValueParser[byte, string])
	return ok
}
func isOptionalParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.OptionalParser[byte, string])
	return ok
}
func isRequiredParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.RequiredParser[byte, string])
	return ok
}
func isLimitParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.LimitParser[byte, string])
	return ok
}
func isForwardParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.ForwardParser[byte, string])
	return ok
}
func isMapParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.MapParser[byte, string])
	return ok
}
func isProduceParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.ProduceParser[byte, string])
	return ok
}
func isRuleParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.RuleParser[byte, string])
	return ok
}
func isApplyParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.ApplyParser[byte, string, string])
	return ok
}
func isIfValueParser(parser kustolang.ResultParser[byte, string]) bool {
	_, ok := parser.(*kustolang.IfValueParser[byte, string])
	return ok
}
