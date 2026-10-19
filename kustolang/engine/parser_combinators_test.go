package kustolang

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type parserCombinatorCacheValue struct {
	Name string
}

func TestParseResultSucceededUsesNonNegativeLength(t *testing.T) {
	success := NewParseResult(2, "value")
	if !success.Succeeded() || success.Length != 2 || success.Value != "value" {
		t.Fatalf("success result = %#v", success)
	}

	failure := NewParseResult(-1, "partial")
	if failure.Succeeded() {
		t.Fatalf("failure result Succeeded = true")
	}
}

func TestOffsetValueStoresOffsetAndValue(t *testing.T) {
	value := NewOffsetValue(7, "name")
	if value.Offset != 7 || value.Value != "name" {
		t.Fatalf("offset value = %#v", value)
	}
}

func TestSourceCacheCreatesStoresAndRetrievesTypedValues(t *testing.T) {
	cache := NewSourceCache()
	value := SourceCacheGetOrCreate[*parserCombinatorCacheValue](cache)
	if value == nil {
		t.Fatal("default pointer cache value is nil")
	}
	value.Name = "cached"

	again := SourceCacheGetOrCreate[*parserCombinatorCacheValue](cache)
	if again != value || again.Name != "cached" {
		t.Fatalf("cached pointer = %#v, want original", again)
	}

	if got, ok := SourceCacheTryGetValue[*parserCombinatorCacheValue](cache); !ok || got != value {
		t.Fatalf("TryGetValue = %#v, %t", got, ok)
	}
}

func TestSourceCacheFactoryMirrorsUpstreamDoubleInvocationOnMiss(t *testing.T) {
	cache := NewSourceCache()
	calls := 0
	first := SourceCacheGetOrCreateFunc(cache, func() *parserCombinatorCacheValue {
		calls++
		return &parserCombinatorCacheValue{Name: "value"}
	})
	if calls != 2 {
		t.Fatalf("factory calls = %d, want 2", calls)
	}
	if first == nil || first.Name != "value" {
		t.Fatalf("factory value = %#v", first)
	}

	second := SourceCacheGetOrCreateFunc(cache, func() *parserCombinatorCacheValue {
		t.Fatal("factory should not run for cached value")
		return nil
	})
	if second != first {
		t.Fatalf("cached factory value = %#v, want original", second)
	}
}

func TestArraySourcePeekAndEndUseStartAndLength(t *testing.T) {
	source := NewArraySource([]int{10, 20, 30, 40}, 1, 2)
	if source.IsEnd() || source.Peek() != 20 || source.Peek(1) != 30 {
		t.Fatalf("array source current values wrong")
	}
	if !source.IsEnd(2) || source.Peek(2) != 0 {
		t.Fatalf("array source end = %t peek=%d, want end and zero", source.IsEnd(2), source.Peek(2))
	}
}

func TestTextSourcePeekTextMatchesAndPosition(t *testing.T) {
	source := NewTextSource("StormEvents", 5, 6)
	if source.Position() != 5 {
		t.Fatalf("position = %d, want 5", source.Position())
	}
	if got := source.Peek(); got != 'E' {
		t.Fatalf("peek = %q, want E", got)
	}
	if got := source.PeekText(0, 6); got != "Events" {
		t.Fatalf("peek text = %q, want Events", got)
	}
	if got := source.PeekText(1); got != "E" {
		t.Fatalf("peek text length = %q, want E", got)
	}
	if !source.Matches(0, "Events") || source.Matches(0, "") {
		t.Fatalf("exact Matches returned unexpected result")
	}
	if !source.Matches(0, "events", true) || source.Matches(0, "events", false) {
		t.Fatalf("case-aware Matches returned unexpected result")
	}
	if !source.IsEnd(6) || source.Peek(6) != 0 {
		t.Fatalf("text source end = %t peek=%q, want end and nul", source.IsEnd(6), source.Peek(6))
	}
}

func TestLimitSourceConstrainsUnderlyingSource(t *testing.T) {
	base := NewArraySource([]string{"a", "b", "c"})
	limited := NewLimitSource[string](base, 2)
	if limited.IsEnd(1) || limited.Peek(1) != "b" {
		t.Fatalf("limited source before limit wrong")
	}
	if !limited.IsEnd(2) || limited.Peek(2) != "c" {
		t.Fatalf("limited source should report end at limit but still delegate Peek")
	}
}

func TestSourceCacheAndDelegateAliases(t *testing.T) {
	source := NewArraySource([]int{4, 5, 6})
	if source.Cache() != source.Cache() {
		t.Fatal("source cache should be stable")
	}

	var consumer SourceConsumer[int] = func(source Source[int], start int) int {
		if source.Peek(start) == 5 {
			return 1
		}
		return -1
	}
	var producer SourceProducer[int, string] = func(source Source[int], start int, length int) string {
		return fmt.Sprint(NewParseResult(length, source.Peek(start)).Value)
	}

	if got := consumer(source, 1); got != 1 {
		t.Fatalf("consumer = %d, want 1", got)
	}
	if got := producer(source, 2, 1); got != "6" {
		t.Fatalf("producer = %q, want 6", got)
	}
}

func TestResultPrimaryParserAppendsOnlyOnSuccess(t *testing.T) {
	parser := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		if source.Peek(start) != 'A' {
			return NewParseResult(-1, "")
		}
		return NewParseResult(1, "letter-a")
	}, WithParserDescription[byte]("literal"))

	output := []any{"prefix"}
	n := parser.ParseList(NewTextSource("Alpha"), 0, &output, len(output))
	if n != 1 || len(output) != 2 || output[1] != "letter-a" {
		t.Fatalf("success parse length=%d output=%#v", n, output)
	}

	n = parser.ParseList(NewTextSource("Beta"), 0, &output, len(output))
	if n != -1 || len(output) != 2 {
		t.Fatalf("failure parse length=%d output=%#v, want no append", n, output)
	}
}

func TestListPrimaryParserResultUsesFirstOutputValue(t *testing.T) {
	parser := NewListPrimaryParser[byte, string](func(source Source[byte], start int, output *[]any, outputStart int) int {
		if source.Peek(start) != 'A' {
			return -1
		}
		*output = append(*output, "first", "second")
		return 1
	})

	result := parser.ParseValue(NewTextSource("Alpha"), 0)
	if !result.Succeeded() || result.Length != 1 || result.Value != "first" {
		t.Fatalf("result = %#v, want first output value", result)
	}

	failure := parser.ParseValue(NewTextSource("Beta"), 0)
	if failure.Succeeded() || failure.Value != "" {
		t.Fatalf("failure = %#v, want default string value", failure)
	}
}

func TestParserMetadataCloneHelpersDoNotMutateOriginal(t *testing.T) {
	child := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		return NewParseResult(0, "child")
	})
	parser := NewResultPrimaryParser(
		func(source Source[byte], start int) ParseResult[string] {
			return NewParseResult(0, "value")
		},
		WithParserTag[byte]("original"),
		WithParserAnnotations[byte]([]any{"a"}),
		WithParserChildren[byte](child),
		WithParserTraits[byte](ParserTraits{Match: true}),
	)

	if same := parser.WithTag("original"); same != parser {
		t.Fatal("WithTag with the same value should return the same parser")
	}
	tagged := parser.WithTag("updated")
	if tagged == parser || tagged.Tag() != "updated" || parser.Tag() != "original" {
		t.Fatalf("tag clone changed original=%q clone=%q", parser.Tag(), tagged.Tag())
	}

	hidden := tagged.Hide()
	if !hidden.IsHidden() || tagged.IsHidden() {
		t.Fatal("Hide should clone with IsHidden=true without mutating the source parser")
	}
	if !parser.IsMatch() || parser.ChildParserCount() != 1 || parser.ChildParser(0) != child || parser.ChildParserIndex(child) != 0 {
		t.Fatalf("traits/children not preserved")
	}

	annotations := parser.Annotations()
	annotations[0] = "mutated"
	if parser.Annotations()[0] != "a" {
		t.Fatalf("annotations should be exposed as a copy")
	}
}

func TestRightParserWrapsTypedMetadataClones(t *testing.T) {
	parser := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		return NewParseResult(1, string([]byte{source.Peek(start)}))
	})
	right := NewRightParser[byte, string](parser).WithTag("right").Hide().WithAnnotations([]any{"anno"})

	if right.Parser() == parser || right.Parser().Tag() != "right" || !right.Parser().IsHidden() {
		t.Fatalf("right parser metadata clone = tag:%q hidden:%t", right.Parser().Tag(), right.Parser().IsHidden())
	}
	if got := right.Parser().Annotations(); len(got) != 1 || got[0] != "anno" {
		t.Fatalf("right parser annotations = %#v", got)
	}
	if value := right.Parser().ParseValue(NewTextSource("Z"), 0); value.Value != "Z" || value.Length != 1 {
		t.Fatalf("right parser value = %#v", value)
	}
}

func TestParserTextExtensionsUseTextSource(t *testing.T) {
	parser := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		if source.Peek(start) != 'S' {
			return NewParseResult(-1, "")
		}
		return NewParseResult(5, source.(*TextSource).PeekText(start, 5))
	})

	output := []any{}
	if n := ParseText(parser, "StormEvents", &output); n != 5 || len(output) != 1 || output[0] != "Storm" {
		t.Fatalf("ParseText length/output = %d/%#v", n, output)
	}

	result := ParseTextValue(parser, "StormEvents")
	if result.Length != 5 || result.Value != "Storm" {
		t.Fatalf("ParseTextValue = %#v", result)
	}

	value, ok := TryParseTextValue(parser, "StormEvents")
	if !ok || value != "Storm" {
		t.Fatalf("TryParseTextValue success = %q, %t", value, ok)
	}
	value, ok = TryParseTextValue(parser, "Events")
	if ok || value != "" {
		t.Fatalf("TryParseTextValue failure = %q, %t", value, ok)
	}
}

func TestFailsAndNotParsersUseUpstreamNegationSemantics(t *testing.T) {
	pattern := newByteLiteralParser('A', "A")
	fails := NewFailsParser[byte](pattern)
	not := NewNotParser[byte](pattern)

	source := NewTextSource("AB")
	if got := fails.Scan(source, 0); got != -1 {
		t.Fatalf("FailsParser on matching pattern = %d, want -1", got)
	}
	if got := fails.Scan(source, 1); got != 0 {
		t.Fatalf("FailsParser on failing pattern = %d, want 0", got)
	}
	if got := fails.Scan(source, 2); got != 0 {
		t.Fatalf("FailsParser at EOF = %d, want 0", got)
	}

	if got := not.Scan(source, 0); got != -1 {
		t.Fatalf("NotParser on matching pattern = %d, want -1", got)
	}
	if got := not.Scan(source, 1); got != 1 {
		t.Fatalf("NotParser on failing pattern = %d, want 1", got)
	}
	if got := not.Scan(source, 2); got != -1 {
		t.Fatalf("NotParser at EOF = %d, want -1", got)
	}
	if !fails.IsNegation() || fails.ChildParser(0) != pattern || not.ChildParserIndex(pattern) != 0 {
		t.Fatalf("negation metadata not preserved")
	}
}

func TestOptionalAndRequiredParsersProduceDefaultsOnFailure(t *testing.T) {
	pattern := newByteLiteralParser('A', "matched")
	optional := NewOptionalParser[byte, string](pattern, func() string { return "optional" })
	required := NewRequiredParser[byte, string](pattern, func(source Source[byte], start int) string {
		return "required-" + string([]byte{source.Peek(start)})
	})

	success := optional.ParseValue(NewTextSource("Alpha"), 0)
	if success.Length != 1 || success.Value != "matched" {
		t.Fatalf("optional success = %#v", success)
	}
	failure := optional.ParseValue(NewTextSource("Beta"), 0)
	if failure.Length != 0 || failure.Value != "optional" {
		t.Fatalf("optional failure = %#v", failure)
	}

	output := []any{"prefix"}
	if n := required.ParseList(NewTextSource("Beta"), 0, &output, len(output)); n != 0 || len(output) != 2 || output[1] != "required-B" {
		t.Fatalf("required failure parse length/output = %d/%#v", n, output)
	}
	if n := required.Scan(NewTextSource("Beta"), 0); n != 0 {
		t.Fatalf("required scan failure = %d, want 0", n)
	}
	if !optional.IsOptional() || !required.IsRequired() {
		t.Fatalf("optional/required traits missing")
	}
}

func TestProduceParserReplacesChildOutputsAndRollsBackFailures(t *testing.T) {
	child := NewListPrimaryParser[byte, string](func(source Source[byte], start int, output *[]any, outputStart int) int {
		if source.Peek(start) != 'A' {
			return -1
		}
		*output = append(*output, "left", "right")
		return 1
	})
	produce := NewProduceParser[byte, string](child, func(output []any, outputStart int) string {
		return fmt.Sprintf("%s+%s", output[outputStart], output[outputStart+1])
	})

	output := []any{"prefix"}
	n := produce.ParseList(NewTextSource("Alpha"), 0, &output, len(output))
	if n != 1 || len(output) != 2 || output[1] != "left+right" {
		t.Fatalf("produce success length/output = %d/%#v", n, output)
	}

	n = produce.ParseList(NewTextSource("Beta"), 0, &output, len(output))
	if n != -1 || len(output) != 2 {
		t.Fatalf("produce failure length/output = %d/%#v, want rollback", n, output)
	}

	result := produce.ParseValue(NewTextSource("Alpha"), 0)
	if result.Length != 1 || result.Value != "left+right" {
		t.Fatalf("produce value = %#v", result)
	}
}

func TestLimitParserConstrainsLimitedParserToLimiterScanLength(t *testing.T) {
	limiter := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		if source.Peek(start) != 'A' || source.Peek(start+1) != 'B' {
			return NewParseResult(-1, "")
		}
		return NewParseResult(2, "limit")
	})
	limited := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		if !source.IsEnd(start+2) && source.Peek(start+2) == 'C' {
			return NewParseResult(3, "too-far")
		}
		if source.Peek(start) == 'A' && source.Peek(start+1) == 'B' && source.IsEnd(start+2) {
			return NewParseResult(2, "limited")
		}
		return NewParseResult(-1, "")
	})

	parser := NewLimitParser[byte, string](limiter, limited)
	result := parser.ParseValue(NewTextSource("ABC"), 0)
	if result.Length != 2 || result.Value != "limited" {
		t.Fatalf("limit result = %#v", result)
	}
	if n := parser.Scan(NewTextSource("ABC"), 0); n != 2 {
		t.Fatalf("limit scan = %d, want 2", n)
	}
	if n := parser.Scan(NewTextSource("XBC"), 0); n != -1 {
		t.Fatalf("limit scan with limiter failure = %d, want -1", n)
	}
	if !parser.IsConditional() || parser.ChildParserCount() != 2 {
		t.Fatalf("limit metadata not preserved")
	}
}

func TestRepetitionParsersStopOnNonPositiveMatches(t *testing.T) {
	child := newByteLiteralParser('A', "A")
	zeroOrMore := NewZeroOrMoreParser[byte](child)
	zeroOrOne := NewZeroOrMoreParser[byte](child, true)
	oneOrMore := NewOneOrMoreParser[byte](child)

	output := []any{}
	if n := zeroOrMore.ParseList(NewTextSource("AAB"), 0, &output, 0); n != 2 || len(output) != 2 {
		t.Fatalf("zero-or-more length/output = %d/%#v", n, output)
	}
	output = []any{}
	if n := zeroOrOne.ParseList(NewTextSource("AAB"), 0, &output, 0); n != 1 || len(output) != 1 {
		t.Fatalf("zero-or-one length/output = %d/%#v", n, output)
	}
	output = []any{}
	if n := zeroOrMore.ParseList(NewTextSource("BAA"), 0, &output, 0); n != 0 || len(output) != 0 {
		t.Fatalf("zero-or-more no-match length/output = %d/%#v", n, output)
	}
	output = []any{}
	if n := oneOrMore.ParseList(NewTextSource("BAA"), 0, &output, 0); n != -1 || len(output) != 0 {
		t.Fatalf("one-or-more failure length/output = %d/%#v", n, output)
	}
	if n := oneOrMore.Scan(NewTextSource("AA"), 0); n != 2 {
		t.Fatalf("one-or-more scan = %d, want 2", n)
	}
	if !zeroOrMore.IsOptional() || !zeroOrMore.IsRepetition() || !oneOrMore.IsRepetition() {
		t.Fatalf("repetition traits missing")
	}
}

func TestMatchAndConvertParsersProduceFromConsumedInput(t *testing.T) {
	digits := NewMatchParser(func(source Source[byte], start int) int {
		length := 0
		for !source.IsEnd(start+length) && source.Peek(start+length) >= '0' && source.Peek(start+length) <= '9' {
			length++
		}
		if length == 0 {
			return -1
		}
		return length
	})
	if !digits.IsMatch() || digits.Scan(NewTextSource("123ms"), 0) != 3 {
		t.Fatalf("match parser scan/traits failed")
	}

	valueParser := NewMatchValueParser(
		func(source Source[byte], start int) int {
			if source.Peek(start) == 'A' {
				return 1
			}
			return -1
		},
		func(source Source[byte], start int, length int) string {
			return string([]byte{source.Peek(start)})
		},
	)
	if result := valueParser.ParseValue(NewTextSource("Alpha"), 0); result.Length != 1 || result.Value != "A" {
		t.Fatalf("match value result = %#v", result)
	}
	if result := valueParser.ParseValue(NewTextSource("Beta"), 0); result.Length != -1 || result.Value != "" {
		t.Fatalf("match value failure = %#v", result)
	}
	emptyValue := NewMatchValueParser(
		func(Source[byte], int) int { return 0 },
		func(Source[byte], int, int) string { return "empty" },
	)
	if result := emptyValue.ParseValue(NewTextSource("Alpha"), 0); result.Length != 0 || result.Value != "empty" {
		t.Fatalf("zero-width match value result = %#v", result)
	}

	converted := NewConvertParser(digits, func(source Source[byte], start int, length int) string {
		return source.(*TextSource).PeekText(start, length)
	})
	if result := converted.ParseValue(NewTextSource("123ms"), 0); result.Length != 3 || result.Value != "123" {
		t.Fatalf("convert source-producer result = %#v", result)
	}

	listConverted := NewConvertParserFromList(digits, func(values []byte) string {
		return string(values)
	})
	if result := listConverted.ParseValue(NewTextSource("456ms"), 0); result.Length != 3 || result.Value != "456" {
		t.Fatalf("convert list result = %#v", result)
	}

	singleConverted := NewConvertParserFromSingle(newByteLiteralParser('Z', "z"), func(value byte) string {
		return string([]byte{value + 1})
	})
	if result := singleConverted.ParseValue(NewTextSource("Zed"), 0); result.Length != 1 || result.Value != "[" {
		t.Fatalf("convert single result = %#v", result)
	}
}

func TestSequenceParserRollsBackOutputsAndAdjustsFailureLength(t *testing.T) {
	seq := NewSequenceParser([]Parser[byte]{
		newByteLiteralParser('A', "a"),
		NewMatchParser(func(source Source[byte], start int) int {
			return -2
		}),
	})
	output := []any{"prefix"}
	if n := seq.ParseList(NewTextSource("AB"), 0, &output, len(output)); n != -3 || len(output) != 1 {
		t.Fatalf("sequence failure length/output = %d/%#v, want adjusted rollback", n, output)
	}
	if n := seq.Scan(NewTextSource("AB"), 0); n != -3 {
		t.Fatalf("sequence scan failure = %d, want -3", n)
	}

	success := NewSequenceParser([]Parser[byte]{
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
	})
	output = []any{}
	if n := success.ParseList(NewTextSource("AB"), 0, &output, 0); n != 2 || len(output) != 2 || output[0] != "a" || output[1] != "b" {
		t.Fatalf("sequence success length/output = %d/%#v", n, output)
	}
	if !success.IsSequence() || success.ChildParserCount() != 2 {
		t.Fatalf("sequence metadata missing")
	}
}

func TestFirstParsersChooseFirstSuccessAndTrackFarthestFailure(t *testing.T) {
	first := NewFirstParser([]Parser[byte]{
		NewMatchParser(func(source Source[byte], start int) int { return -3 }),
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('A', "a"),
	})
	output := []any{}
	if n := first.ParseList(NewTextSource("Alpha"), 0, &output, 0); n != 1 || len(output) != 1 || output[0] != "a" {
		t.Fatalf("first success length/output = %d/%#v", n, output)
	}
	if n := first.Scan(NewTextSource("Zulu"), 0); n != -3 {
		t.Fatalf("first farthest failure = %d, want -3", n)
	}
	if !first.IsAlternation() || first.ChildParserCount() != 3 {
		t.Fatalf("first metadata missing")
	}

	noisyFailure := NewListPrimaryParser[byte, string](func(source Source[byte], start int, output *[]any, outputStart int) int {
		*output = append(*output, "noise")
		return -1
	})
	typed := NewFirstValueParser([]ResultParser[byte, string]{
		noisyFailure,
		newByteLiteralParser('A', "a"),
	})
	output = []any{"prefix"}
	if n := typed.ParseList(NewTextSource("Alpha"), 0, &output, len(output)); n != 1 || len(output) != 2 || output[1] != "a" {
		t.Fatalf("first typed output reset length/output = %d/%#v", n, output)
	}
}

func TestBestParsersChooseLongestAndResolveTiesWithComparer(t *testing.T) {
	short := NewMatchValueParser(
		func(source Source[byte], start int) int {
			if source.Peek(start) == 'A' {
				return 1
			}
			return -1
		},
		func(source Source[byte], start int, length int) string { return "short" },
	)
	long := NewMatchValueParser(
		func(source Source[byte], start int) int {
			if source.Peek(start) == 'A' && source.Peek(start+1) == 'B' {
				return 2
			}
			return -2
		},
		func(source Source[byte], start int, length int) string { return "long" },
	)
	best := NewBestValueParser([]ResultParser[byte, string]{short, long})
	if result := best.ParseValue(NewTextSource("ABC"), 0); result.Length != 2 || result.Value != "long" {
		t.Fatalf("best longest result = %#v", result)
	}
	if result := best.ParseValue(NewTextSource("ZZZ"), 0); result.Length != -2 || result.Value != "" {
		t.Fatalf("best failure result = %#v", result)
	}

	tieA := NewMatchValueParser(
		func(source Source[byte], start int) int { return 1 },
		func(source Source[byte], start int, length int) string { return "a" },
	)
	tieB := NewMatchValueParser(
		func(source Source[byte], start int) int { return 1 },
		func(source Source[byte], start int, length int) string { return "b" },
	)
	tie := NewBestValueParser([]ResultParser[byte, string]{tieA, tieB}, func(candidate string, current string) bool {
		return candidate > current
	})
	if result := tie.ParseValue(NewTextSource("X"), 0); result.Length != 1 || result.Value != "b" {
		t.Fatalf("best tie result = %#v", result)
	}

	listBest := NewBestParser([]Parser[byte]{newByteLiteralParser('A', "a"), long})
	output := []any{}
	if n := listBest.ParseList(NewTextSource("ABC"), 0, &output, 0); n != 2 || len(output) != 1 || output[0] != "long" {
		t.Fatalf("best list length/output = %d/%#v", n, output)
	}
	if !listBest.IsAlternation() || listBest.ChildParserCount() != 2 {
		t.Fatalf("best metadata missing")
	}
}

func TestRuleParserProducesFromSequenceOutputsAndUsesResultProducer(t *testing.T) {
	rule := NewRuleParser([]Parser[byte]{
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
	}, func(output []any, outputStart int) string {
		return output[outputStart].(string) + output[outputStart+1].(string)
	})
	output := []any{"prefix"}
	if n := rule.ParseList(NewTextSource("AB"), 0, &output, len(output)); n != 2 || len(output) != 2 || output[1] != "ab" {
		t.Fatalf("rule success length/output = %d/%#v", n, output)
	}
	if result := rule.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "ab" {
		t.Fatalf("rule value result = %#v", result)
	}
	if n := rule.ParseList(NewTextSource("AX"), 0, &output, len(output)); n != -2 || len(output) != 2 {
		t.Fatalf("rule failure length/output = %d/%#v, want rollback", n, output)
	}

	resultRule := NewRuleParser([]Parser[byte]{newByteLiteralParser('A', "a")}, func(output []any, outputStart int) string {
		t.Fatal("list producer should not be used by ParseValue when result producer exists")
		return ""
	}, func(source Source[byte], start int) ParseResult[string] {
		return NewParseResult(1, "result")
	})
	if result := resultRule.ParseValue(NewTextSource("A"), 0); result.Length != 1 || result.Value != "result" {
		t.Fatalf("rule explicit result = %#v", result)
	}
	if !rule.IsSequence() || rule.ChildParserCount() != 2 {
		t.Fatalf("rule metadata missing")
	}
}

func TestIfParsersGateOnTestScanWithoutConsumingIt(t *testing.T) {
	test := NewMatchParser(func(source Source[byte], start int) int {
		if source.Peek(start) == 'A' {
			return 1
		}
		return -2
	})
	parser := NewIfParser[byte](test, newByteLiteralParser('A', "a"))
	output := []any{}
	if n := parser.ParseList(NewTextSource("Alpha"), 0, &output, 0); n != 1 || len(output) != 1 || output[0] != "a" {
		t.Fatalf("if parser success length/output = %d/%#v", n, output)
	}
	if n := parser.ParseList(NewTextSource("Beta"), 0, &output, len(output)); n != -2 || len(output) != 1 {
		t.Fatalf("if parser test failure length/output = %d/%#v", n, output)
	}
	if !parser.IsConditional() || parser.ChildParserCount() != 2 || parser.ChildParser(0) != test {
		t.Fatalf("if parser metadata missing")
	}

	typed := NewIfValueParser[byte, string](test, newByteLiteralParser('A', "typed"))
	if result := typed.ParseValue(NewTextSource("Alpha"), 0); result.Length != 1 || result.Value != "typed" {
		t.Fatalf("typed if success = %#v", result)
	}
	if result := typed.ParseValue(NewTextSource("Beta"), 0); result.Length != -2 || result.Value != "" {
		t.Fatalf("typed if failure = %#v", result)
	}
}

func TestApplyParserPreservesUpstreamApplicationAndRollback(t *testing.T) {
	left := newByteLiteralParser('A', "left")
	right := NewRightParser[byte, string](newByteLiteralParser('B', "right"))

	one := NewApplyParser[byte, string, string](ApplyKindOne, left, right)
	output := []any{}
	if n := one.ParseList(NewTextSource("AB"), 0, &output, 0); n != 2 || len(output) != 2 || output[0] != "left" || output[1] != "right" {
		t.Fatalf("apply one success length/output = %d/%#v", n, output)
	}
	if result := one.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "left" {
		t.Fatalf("apply one value = %#v", result)
	}

	output = []any{"prefix"}
	if n := one.ParseList(NewTextSource("AX"), 0, &output, len(output)); n != -2 || len(output) != 1 {
		t.Fatalf("apply one right failure length/output = %d/%#v, want rollback", n, output)
	}
	if n := one.Scan(NewTextSource("AX"), 0); n != -2 {
		t.Fatalf("apply one scan right failure = %d, want -2", n)
	}

	zeroOrOne := NewApplyParser[byte, string, string](ApplyKindZeroOrOne, left, right)
	output = []any{}
	if n := zeroOrOne.ParseList(NewTextSource("AX"), 0, &output, 0); n != 1 || len(output) != 1 || output[0] != "left" {
		t.Fatalf("apply zero-or-one failure length/output = %d/%#v", n, output)
	}

	zeroOrMore := NewApplyParser[byte, string, string](ApplyKindZeroOrMore, left, right)
	output = []any{}
	if n := zeroOrMore.ParseList(NewTextSource("ABBBX"), 0, &output, 0); n != 4 || len(output) != 1 || output[0] != "left" {
		t.Fatalf("apply zero-or-more length/output = %d/%#v", n, output)
	}
	if n := zeroOrMore.Scan(NewTextSource("ABBBX"), 0); n != 4 {
		t.Fatalf("apply zero-or-more scan = %d, want 4", n)
	}
	if zeroOrMore.ChildParserCount() != 2 || zeroOrMore.ChildParser(0) != left || zeroOrMore.ChildParser(1) != right.Parser() {
		t.Fatalf("apply child metadata missing")
	}
}

func TestMapParserChoosesLongestMappedPrefix(t *testing.T) {
	parser := NewMapParser([]MapEntry[byte, string]{
		{Key: []byte("a"), Value: func() string { return "short" }},
		{Key: []byte("abc"), Value: func() string { return "long" }},
		{Key: []byte("ab"), Value: func() string { return "middle" }},
	})

	result := parser.ParseValue(NewTextSource("abcd"), 0)
	if result.Length != 3 || result.Value != "long" {
		t.Fatalf("map parser result = %#v, want longest abc", result)
	}
	if n := parser.Scan(NewTextSource("abcd"), 0); n != 3 {
		t.Fatalf("map parser scan = %d, want 3", n)
	}

	output := []any{"prefix"}
	if n := parser.ParseList(NewTextSource("abcd"), 0, &output, len(output)); n != 3 || len(output) != 2 || output[1] != "long" {
		t.Fatalf("map parser list length/output = %d/%#v", n, output)
	}
	if !parser.IsMatch() || parser.ChildParserCount() != 0 {
		t.Fatalf("map parser metadata missing")
	}
}

func TestMapParserFailureDuplicateAndEmptyKeyBehavior(t *testing.T) {
	parser := NewMapParser([]MapEntry[byte, string]{
		{Key: []byte(""), Value: func() string { return "empty" }},
		{Key: []byte("x"), Value: func() string { return "first" }},
		{Key: []byte("x"), Value: func() string { return "second" }},
	})

	duplicate := parser.ParseValue(NewTextSource("xyz"), 0)
	if duplicate.Length != 1 || duplicate.Value != "second" {
		t.Fatalf("map duplicate result = %#v, want last value", duplicate)
	}

	failure := parser.ParseValue(NewTextSource("abc"), 0)
	if failure.Length != -1 || failure.Value != "" {
		t.Fatalf("map failure = %#v, want default string and -1", failure)
	}
	if n := parser.Scan(NewTextSource(""), 0); n != -1 {
		t.Fatalf("map empty input scan = %d, want -1", n)
	}

	tagged := parser.WithTag("map")
	if tagged == parser || tagged.Tag() != "map" || parser.Tag() != "" {
		t.Fatalf("map WithTag clone failed: original=%q clone=%q", parser.Tag(), tagged.Tag())
	}
}

func TestParserFactoriesMatchConvertAndTextHelpers(t *testing.T) {
	anyParser := Any[byte]()
	if anyParser.Tag() != "<any>" || anyParser.Scan(NewTextSource("x"), 0) != 1 {
		t.Fatalf("Any parser tag/scan = %q/%d", anyParser.Tag(), anyParser.Scan(NewTextSource("x"), 0))
	}

	sequence := MatchSequence([]byte("ab"))
	if n := sequence.Scan(NewTextSource("ax"), 0); n != -2 {
		t.Fatalf("MatchSequence mismatch = %d, want -2", n)
	}
	if n := MatchText("Storm", true).Scan(NewTextSource("storm"), 0); n != 5 {
		t.Fatalf("MatchText ignore-case = %d, want 5", n)
	}

	text := Text(MatchText("Storm"))
	if result := text.ParseValue(NewTextSource("StormEvents"), 0); result.Length != 5 || result.Value != "Storm" {
		t.Fatalf("Text result = %#v", result)
	}
	if result := TextLiteral("Storm").ParseValue(NewTextSource("StormEvents"), 0); result.Length != 5 || result.Value != "Storm" {
		t.Fatalf("TextLiteral result = %#v", result)
	}
	if result := TextAndOffset(MatchText("Events")).ParseValue(NewTextSource("StormEvents"), 5); result.Length != 6 || result.Value.Offset != 5 || result.Value.Value != "Events" {
		t.Fatalf("TextAndOffset result = %#v", result)
	}

	count := Count(MatchSequence([]byte("abc")))
	if result := count.ParseValue(NewTextSource("abcdef"), 0); result.Length != 3 || result.Value != 3 {
		t.Fatalf("Count result = %#v", result)
	}
}

func TestParserFactoriesComposeMapsRulesAndApply(t *testing.T) {
	mapped := MapValues([]MapValueEntry[byte, string]{
		{Key: []byte("a"), Value: "short"},
		{Key: []byte("ab"), Value: "long"},
	})
	if result := mapped.ParseValue(NewTextSource("abc"), 0); result.Length != 2 || result.Value != "long" {
		t.Fatalf("MapValues result = %#v", result)
	}

	first := First(newByteLiteralParser('X', "x"), newByteLiteralParser('A', "a"))
	output := []any{}
	if n := first.ParseList(NewTextSource("Alpha"), 0, &output, 0); n != 1 || output[0] != "a" {
		t.Fatalf("First factory length/output = %d/%#v", n, output)
	}

	best := BestValue(newByteLiteralParser('A', "short"), NewMatchValueParser(
		func(source Source[byte], start int) int {
			if source.Peek(start) == 'A' && source.Peek(start+1) == 'B' {
				return 2
			}
			return -1
		},
		func(source Source[byte], start int, length int) string { return "long" },
	))
	if result := best.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "long" {
		t.Fatalf("BestValue result = %#v", result)
	}

	rule := Rule2(newByteLiteralParser('A', "a"), newByteLiteralParser('B', "b"), func(a string, b string) string {
		return a + b
	})
	if result := rule.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "ab" {
		t.Fatalf("Rule2 result = %#v", result)
	}

	rightRule := RuleRight1[byte, string](newByteLiteralParser('B', "b"), func(left string, right string) string {
		return left + right
	})
	apply := Apply(newByteLiteralParser('A', "a"), func(LeftValue[string]) RightParser[byte, string] {
		return rightRule
	})
	if result := apply.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "ab" {
		t.Fatalf("Apply/RuleRight result = %#v", result)
	}
}

func TestParserFactoriesHigherArityRules(t *testing.T) {
	rule3 := Rule3(
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		func(a string, b string, c string) string {
			return a + b + c
		},
	)
	if result := rule3.ParseValue(NewTextSource("ABC"), 0); result.Length != 3 || result.Value != "abc" {
		t.Fatalf("Rule3 result = %#v", result)
	}
	if result := rule3.ParseValue(NewTextSource("ABX"), 0); result.Length != -3 || result.Value != "" {
		t.Fatalf("Rule3 failure = %#v, want adjusted -3", result)
	}

	rule4 := Rule4(
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		newByteLiteralParser('D', "d"),
		func(a string, b string, c string, d string) string {
			return a + b + c + d
		},
	)
	output := []any{}
	if n := rule4.ParseList(NewTextSource("ABCD"), 0, &output, 0); n != 4 || len(output) != 1 || output[0] != "abcd" {
		t.Fatalf("Rule4 list length/output = %d/%#v", n, output)
	}

	rule5 := Rule5(
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		newByteLiteralParser('D', "d"),
		newByteLiteralParser('E', "e"),
		func(a string, b string, c string, d string, e string) string {
			return a + b + c + d + e
		},
	)
	if result := rule5.ParseValue(NewTextSource("ABCDE"), 0); result.Length != 5 || result.Value != "abcde" {
		t.Fatalf("Rule5 result = %#v", result)
	}
	if result := rule5.ParseValue(NewTextSource("ABCDx"), 0); result.Length != -5 || result.Value != "" {
		t.Fatalf("Rule5 failure = %#v, want adjusted -5", result)
	}

	rule9 := Rule9(
		newByteLiteralParser('A', "a"),
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		newByteLiteralParser('D', "d"),
		newByteLiteralParser('E', "e"),
		newByteLiteralParser('F', "f"),
		newByteLiteralParser('G', "g"),
		newByteLiteralParser('H', "h"),
		newByteLiteralParser('I', "i"),
		func(a string, b string, c string, d string, e string, f string, g string, h string, i string) string {
			return a + b + c + d + e + f + g + h + i
		},
	)
	output = []any{}
	if n := rule9.ParseList(NewTextSource("ABCDEFGHI"), 0, &output, 0); n != 9 || len(output) != 1 || output[0] != "abcdefghi" {
		t.Fatalf("Rule9 list length/output = %d/%#v", n, output)
	}
}

type optionalRuleInterface interface {
	optionalRuleMarker()
}

type optionalRuleValue struct{}

func (optionalRuleValue) optionalRuleMarker() {}

func TestRuleParseListPreservesNilOptionalInterfaceValues(t *testing.T) {
	optionalInterface := Rule1(
		newByteLiteralParser('B', "b"),
		func(string) optionalRuleInterface { return optionalRuleValue{} },
	)
	rule := Rule2(
		newByteLiteralParser('A', "a"),
		OptionalValue(optionalInterface),
		func(required string, optional optionalRuleInterface) string {
			if required == "a" && optional == nil {
				return "missing optional"
			}
			return "unexpected"
		},
	)

	output := []any{}
	if length := rule.ParseList(NewTextSource("A"), 0, &output, 0); length != 1 || len(output) != 1 || output[0] != "missing optional" {
		t.Fatalf("Rule2 optional-interface ParseList = length:%d output:%#v", length, output)
	}
}

func TestParserFactoriesRightRulesUseExistingLeftOutput(t *testing.T) {
	left := newByteLiteralParser('A', "a")
	right0 := RuleRight0[byte, string](func(left string) string {
		return left + "!"
	})
	apply0 := ApplyOptional(left, func(LeftValue[string]) RightParser[byte, string] {
		return right0
	})
	if result := apply0.ParseValue(NewTextSource("A"), 0); result.Length != 1 || result.Value != "a!" {
		t.Fatalf("RuleRight0 apply result = %#v", result)
	}

	right2 := RuleRight2(
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		func(left string, b string, c string) string {
			return left + b + c
		},
	)
	apply2 := Apply(left, func(LeftValue[string]) RightParser[byte, string] {
		return right2
	})
	if result := apply2.ParseValue(NewTextSource("ABC"), 0); result.Length != 3 || result.Value != "abc" {
		t.Fatalf("RuleRight2 apply result = %#v", result)
	}

	right3 := RuleRight3(
		newByteLiteralParser('B', "b"),
		newByteLiteralParser('C', "c"),
		newByteLiteralParser('D', "d"),
		func(left string, b string, c string, d string) string {
			return left + b + c + d
		},
	)
	apply3 := Apply(left, func(LeftValue[string]) RightParser[byte, string] {
		return right3
	})
	if result := apply3.ParseValue(NewTextSource("ABCD"), 0); result.Length != 4 || result.Value != "abcd" {
		t.Fatalf("RuleRight3 apply result = %#v", result)
	}
}

func TestForwardParserDelegatesAndExposesDeferredChild(t *testing.T) {
	literal := newByteLiteralParser('A', "a")
	deferredCalls := 0
	forward := Forward(func() ResultParser[byte, string] {
		deferredCalls++
		return literal
	})

	if !forward.IsForward() {
		t.Fatalf("Forward IsForward = false")
	}
	if forward.ChildParserCount() != 1 || forward.ChildParser(0) != literal || forward.ChildParser(1) != nil {
		t.Fatalf("Forward child parser metadata mismatch")
	}
	if forward.ChildParserIndex(literal) != 0 || forward.ChildParserIndex(newByteLiteralParser('B', "b")) != -1 {
		t.Fatalf("Forward child parser index mismatch")
	}
	if result := forward.ParseValue(NewTextSource("A"), 0); result.Length != 1 || result.Value != "a" {
		t.Fatalf("Forward ParseValue result = %#v", result)
	}
	if deferredCalls == 0 {
		t.Fatalf("Forward did not call deferred parser")
	}

	fixed := ForwardValue(literal)
	tagged := asResultParser[byte, string](fixed.WithTag("fixed"))
	if fixed.Tag() != "" || tagged.Tag() != "fixed" {
		t.Fatalf("ForwardValue clone tags = original %q tagged %q", fixed.Tag(), tagged.Tag())
	}
}

func TestForwardParserParsesNestedGrammarPastDepthThreshold(t *testing.T) {
	atom := newByteLiteralParser('A', "a")
	var expr ResultParser[byte, string]
	expr = Forward(func() ResultParser[byte, string] {
		return FirstValue(
			Rule3(
				newByteLiteralParser('(', "("),
				expr,
				newByteLiteralParser(')', ")"),
				func(left string, inner string, right string) string {
					return left + inner + right
				},
			),
			atom,
		)
	})

	depth := forwardParserMaxCallDepth + 5
	input := strings.Repeat("(", depth) + "A" + strings.Repeat(")", depth)
	expected := strings.Repeat("(", depth) + "a" + strings.Repeat(")", depth)
	if result := expr.ParseValue(NewTextSource(input), 0); result.Length != len(input) || result.Value != expected {
		t.Fatalf("deep Forward ParseValue result = %#v, want length %d value %q", result, len(input), expected)
	}
	output := []any{}
	if n := expr.ParseList(NewTextSource(input), 0, &output, 0); n != len(input) || len(output) != 1 || output[0] != expected {
		t.Fatalf("deep Forward ParseList length/output = %d/%#v", n, output)
	}
	if n := expr.Scan(NewTextSource(input), 0); n != len(input) {
		t.Fatalf("deep Forward Scan length = %d, want %d", n, len(input))
	}
}

func TestScanSafeMatchesDirectScanForPortedCombinators(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	ab := MatchText("AB").WithTag("AB")
	apply := Apply(a, func(LeftValue[string]) RightParser[byte, string] {
		return RuleRight1(b, func(left string, right string) string { return left + right })
	})
	limited := Limit(ab, a)
	sequence := Sequence(a, b)

	cases := []struct {
		name   string
		parser Parser[byte]
		input  string
	}{
		{"apply", apply, "AB"},
		{"best", Best(MatchText("A"), MatchText("AB")), "AB"},
		{"convert", Text(MatchText("AB")), "AB"},
		{"fails", Fails(MatchText("Z")), "AB"},
		{"first", First(MatchText("Z"), MatchText("AB")), "AB"},
		{"forward", ForwardValue(a), "AB"},
		{"if", If(MatchText("A"), ab), "AB"},
		{"limit", limited, "AB"},
		{"map", MapValues([]MapValueEntry[byte, string]{{Key: []byte("AB"), Value: "ab"}}), "AB"},
		{"match", MatchText("AB"), "AB"},
		{"not", Not(MatchText("Z")), "AB"},
		{"one-or-more", OneOrMore(a), "AAZ"},
		{"optional", OptionalValue(a), "Z"},
		{"produce", Produce(sequence, func([]any) string { return "ab" }), "AB"},
		{"required", RequiredValue(a, func() string { return "missing" }), "Z"},
		{"rule", Rule2(a, b, func(left string, right string) string { return left + right }), "AB"},
		{"sequence", sequence, "AB"},
		{"zero-or-more", ZeroOrMore(a), "AAZ"},
	}

	for _, tc := range cases {
		source := NewTextSource(tc.input)
		if got, want := ScanSafe(tc.parser, source, 0), tc.parser.Scan(source, 0); got != want {
			t.Fatalf("%s ScanSafe = %d, want direct scan %d", tc.name, got, want)
		}
	}
}

func TestScanSafeHandlesDeepForwardRecursion(t *testing.T) {
	atom := newByteLiteralParser('A', "a")
	var expr ResultParser[byte, string]
	expr = Forward(func() ResultParser[byte, string] {
		return FirstValue(
			Rule3(
				newByteLiteralParser('(', "("),
				expr,
				newByteLiteralParser(')', ")"),
				func(left string, inner string, right string) string {
					return left + inner + right
				},
			),
			atom,
		)
	})

	depth := forwardParserMaxCallDepth * 10
	input := strings.Repeat("(", depth) + "A" + strings.Repeat(")", depth)
	if got := ScanSafe(expr, NewTextSource(input), 0); got != len(input) {
		t.Fatalf("ScanSafe deep forward length = %d, want %d", got, len(input))
	}
	if got := expr.Scan(NewTextSource(input), 0); got != len(input) {
		t.Fatalf("Forward Scan deep length = %d, want %d", got, len(input))
	}
}

func TestParseSafeMatchesDirectParseForPortedCombinators(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	ab := MatchText("AB").WithTag("AB")
	apply := Apply(a, func(LeftValue[string]) RightParser[byte, string] {
		return RuleRight1(b, func(left string, right string) string { return left + right })
	})
	sequence := Sequence(a, b)

	cases := []struct {
		name   string
		parser Parser[byte]
		input  string
	}{
		{"apply", apply, "AB"},
		{"best", Best(MatchText("A"), Text(MatchText("AB"))), "AB"},
		{"convert", Text(MatchText("AB")), "AB"},
		{"fails", Fails(MatchText("Z")), "AB"},
		{"first", First(Text(MatchText("Z")), Text(MatchText("AB"))), "AB"},
		{"forward", ForwardValue(a), "AB"},
		{"if", If(MatchText("A"), Text(ab)), "AB"},
		{"limit", Limit(ab, Text(MatchText("A"))), "AB"},
		{"map", MapValues([]MapValueEntry[byte, string]{{Key: []byte("AB"), Value: "ab"}}), "AB"},
		{"match", MatchText("AB"), "AB"},
		{"not", Not(MatchText("Z")), "AB"},
		{"one-or-more", OneOrMore(a), "AAZ"},
		{"optional", OptionalValue(a), "Z"},
		{"produce", Produce(sequence, func(values []any) string { return values[0].(string) + values[1].(string) }), "AB"},
		{"required", RequiredValue(a, func() string { return "missing" }), "Z"},
		{"rule", Rule2(a, b, func(left string, right string) string { return left + right }), "AB"},
		{"sequence", sequence, "AB"},
		{"zero-or-more", ZeroOrMore(a), "AAZ"},
	}

	for _, tc := range cases {
		source := NewTextSource(tc.input)
		directOutput := []any{"prefix"}
		safeOutput := []any{"prefix"}
		directLength := tc.parser.ParseList(source, 0, &directOutput, len(directOutput))
		safeLength := ParseSafe(tc.parser, source, 0, &safeOutput, len(safeOutput))
		if directLength != safeLength || !reflect.DeepEqual(directOutput, safeOutput) {
			t.Fatalf("%s ParseSafe length/output = %d/%#v, want %d/%#v", tc.name, safeLength, safeOutput, directLength, directOutput)
		}
	}
}

func TestParseSafeHandlesDeepForwardRecursion(t *testing.T) {
	atom := newByteLiteralParser('A', "a")
	var expr ResultParser[byte, string]
	expr = Forward(func() ResultParser[byte, string] {
		return FirstValue(
			Rule3(
				newByteLiteralParser('(', "("),
				expr,
				newByteLiteralParser(')', ")"),
				func(left string, inner string, right string) string {
					return left + inner + right
				},
			),
			atom,
		)
	})

	depth := forwardParserMaxCallDepth * 10
	input := strings.Repeat("(", depth) + "A" + strings.Repeat(")", depth)
	expected := strings.Repeat("(", depth) + "a" + strings.Repeat(")", depth)
	output := []any{}
	if got := ParseSafe(expr, NewTextSource(input), 0, &output, 0); got != len(input) || len(output) != 1 || output[0] != expected {
		t.Fatalf("ParseSafe deep forward length/output = %d/%#v", got, output)
	}
	if result := expr.ParseValue(NewTextSource(input), 0); result.Length != len(input) || result.Value != expected {
		t.Fatalf("Forward ParseValue deep result = %#v", result)
	}
}

func TestParserDescriberFormatsGrammarLikeUpstream(t *testing.T) {
	a := MatchText("A").WithTag("A")
	b := MatchText("B").WithTag("B")
	c := MatchText("C").WithTag("C")

	if got := DescribeParser(Sequence(a, First(b, c))); got != "A (B | C)" {
		t.Fatalf("sequence/alternation description = %q", got)
	}
	if got := DescribeParser(Sequence(a, b.Hide(), c)); got != "A C" {
		t.Fatalf("hidden parser description = %q", got)
	}
	if got := DescribeParser(First(Sequence(a, b), c)); got != "(A B) | C" {
		t.Fatalf("same-separator description = %q", got)
	}
}

func TestParserDescriberRequiredWrappersAndDescription(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	required := RequiredValue(a, func() string { return "missing" })
	if got := DescribeParser(required); got != "A!" {
		t.Fatalf("required description = %q", got)
	}
	if got := DescribeParser(required, false); got != "A" {
		t.Fatalf("required hidden-marker description = %q", got)
	}
	if got := required.WithTag("required").Description(); got != "required: A!" {
		t.Fatalf("tagged required Description = %q", got)
	}

	literal := NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		return NewParseResult(-1, "")
	}, WithParserDescription[byte]("literal")).WithTag("named")
	if got := literal.Description(); got != "named: literal" {
		t.Fatalf("explicit parser Description = %q", got)
	}
}

func TestParserDescriberTermsNegationAndApply(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")

	if got := DescribeParser(Not(a)); got != "not(A)" {
		t.Fatalf("not description = %q", got)
	}
	if got := DescribeParser(Fails(a)); got != "fails(A)" {
		t.Fatalf("fails description = %q", got)
	}
	if got := DescribeParser(MatchText("Z")); got != "match()" {
		t.Fatalf("match description = %q", got)
	}
	if got := DescribeParser(MapValues([]MapValueEntry[byte, string]{{Key: []byte("A"), Value: "a"}})); got != "map()" {
		t.Fatalf("map description = %q", got)
	}
	if got := DescribeParser(ForwardValue(a)); got != "forward()" {
		t.Fatalf("forward description = %q", got)
	}

	right := RuleRight1(b, func(left string, right string) string { return left + right })
	if got := DescribeParser(ApplyZeroOrMore(a, func(LeftValue[string]) RightParser[byte, string] { return right })); got != "A {B}" {
		t.Fatalf("apply zero-or-more description = %q", got)
	}
	if got := DescribeParser(ApplyOptional(a, func(LeftValue[string]) RightParser[byte, string] { return right })); got != "A [B]" {
		t.Fatalf("apply optional description = %q", got)
	}
	if got := DescribeParser(Produce(Sequence(a, b), func([]any) string { return "" })); got != "A B" {
		t.Fatalf("produce description = %q", got)
	}
	if got := DescribeParser(If(MatchText("test").WithTag("test"), b)); got != "B" {
		t.Fatalf("if description = %q", got)
	}
	if got := DescribeParser(Limit(MatchText("limit").WithTag("limit"), b)); got != "B" {
		t.Fatalf("limit description = %q", got)
	}
}

func TestIsParserParentMirrorsVisitorChildRelationships(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	c := taggedByteLiteralParser('C', "c", "C")
	sequenceForProduce := Sequence(a, b)

	cases := []struct {
		name   string
		parent Parser[byte]
		child  Parser[byte]
	}{
		{"apply-left", Apply(a, func(LeftValue[string]) RightParser[byte, string] {
			return RuleRight1(b, func(left string, right string) string { return left + right })
		}), a},
		{"sequence", Sequence(a, b), b},
		{"first", First(a, b), b},
		{"best", Best(a, b), a},
		{"convert", Text(a), a},
		{"forward", ForwardValue(a), a},
		{"if-test", If(a, b), a},
		{"if-parser", If(a, b), b},
		{"limit-limiter", Limit(a, b), a},
		{"limit-limited", Limit(a, b), b},
		{"not", Not(a), a},
		{"one-or-more", OneOrMore(a), a},
		{"optional", OptionalValue(a), a},
		{"produce", Produce(sequenceForProduce, func([]any) string { return "" }), sequenceForProduce},
		{"required", RequiredValue(a, func() string { return "missing" }), a},
		{"rule", Rule2(a, b, func(left string, right string) string { return left + right }), b},
		{"zero-or-more", ZeroOrMore(a), a},
	}

	for _, tc := range cases {
		if !IsParserParent(tc.parent, tc.child) {
			t.Fatalf("%s parent relationship was false", tc.name)
		}
		if IsParserParent(tc.parent, c) && tc.child != c {
			t.Fatalf("%s unrelated child relationship was true", tc.name)
		}
	}

	if IsParserParent(MapValues([]MapValueEntry[byte, string]{{Key: []byte("A"), Value: "a"}}), a) {
		t.Fatalf("map parser reported a child")
	}
	if IsParserParent(MatchText("A"), a) {
		t.Fatalf("match parser reported a child")
	}
}

func TestIsParserParentPreservesFailsVisitorQuirk(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	fails := Fails(a)

	if IsParserParent(fails, a) {
		t.Fatalf("FailsParser IsParentVisitor should not report its pattern as parented")
	}
	if !IsParserParent(fails, b) {
		t.Fatalf("FailsParser IsParentVisitor should preserve upstream non-pattern result")
	}
	if fails.ChildParserIndex(a) != 0 {
		t.Fatalf("FailsParser child metadata should still expose the pattern")
	}
}

func TestParserVisitorResultDispatchesPortedKinds(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	sequence := Sequence(a, b)

	cases := []struct {
		name   string
		parser Parser[byte]
		want   string
	}{
		{"apply", Apply(a, func(LeftValue[string]) RightParser[byte, string] {
			return RuleRight1(b, func(left string, right string) string { return left + right })
		}), "apply:1"},
		{"best", Best(a, b), "best:2"},
		{"best-value", BestValue(a, b), "best:2"},
		{"convert", Text(MatchText("AB")), "convert"},
		{"fails", Fails(a), "fails"},
		{"first", First(a, b), "first:2"},
		{"first-value", FirstValue(a, b), "first:2"},
		{"forward", ForwardValue(a), "forward"},
		{"if", If(a, b), "if"},
		{"limit", Limit(MatchText("AB"), a), "limit"},
		{"map", MapValues([]MapValueEntry[byte, string]{{Key: []byte("AB"), Value: "ab"}}), "map"},
		{"match", MatchText("AB"), "match"},
		{"match-value", MatchValue(func(input byte) bool { return input == 'A' }, func(byte) string { return "a" }), "match"},
		{"not", Not(a), "not"},
		{"one-or-more", OneOrMore(a), "one-or-more"},
		{"optional", OptionalValue(a), "optional"},
		{"produce", Produce(sequence, func([]any) string { return "ab" }), "produce"},
		{"required", RequiredValue(a, func() string { return "missing" }), "required"},
		{"rule", Rule2(a, b, func(left string, right string) string { return left + right }), "rule:2"},
		{"sequence", sequence, "sequence:2"},
		{"zero-or-more", ZeroOrMore(a), "zero-or-more:false"},
		{"zero-or-one", ZeroOrOne(a), "zero-or-more:true"},
	}

	visitor := ParserVisitorResultFuncs[byte, string]{
		Apply:     func(Parser[byte], Parser[byte], Parser[byte], ApplyKind) string { return "apply:1" },
		Best:      func(_ Parser[byte], parsers []Parser[byte]) string { return fmt.Sprintf("best:%d", len(parsers)) },
		Convert:   func(Parser[byte], Parser[byte]) string { return "convert" },
		Fails:     func(Parser[byte], Parser[byte]) string { return "fails" },
		First:     func(_ Parser[byte], parsers []Parser[byte]) string { return fmt.Sprintf("first:%d", len(parsers)) },
		Forward:   func(Parser[byte], Parser[byte]) string { return "forward" },
		If:        func(Parser[byte], Parser[byte], Parser[byte]) string { return "if" },
		Limit:     func(Parser[byte], Parser[byte], Parser[byte]) string { return "limit" },
		Map:       func(Parser[byte]) string { return "map" },
		Match:     func(Parser[byte]) string { return "match" },
		Not:       func(Parser[byte], Parser[byte]) string { return "not" },
		OneOrMore: func(Parser[byte], Parser[byte]) string { return "one-or-more" },
		Optional:  func(Parser[byte], Parser[byte]) string { return "optional" },
		Produce:   func(Parser[byte], Parser[byte]) string { return "produce" },
		Required:  func(Parser[byte], Parser[byte]) string { return "required" },
		Rule:      func(_ Parser[byte], parsers []Parser[byte]) string { return fmt.Sprintf("rule:%d", len(parsers)) },
		Sequence:  func(_ Parser[byte], parsers []Parser[byte]) string { return fmt.Sprintf("sequence:%d", len(parsers)) },
		ZeroOrMore: func(_ Parser[byte], _ Parser[byte], zeroOrOne bool) string {
			return fmt.Sprintf("zero-or-more:%t", zeroOrOne)
		},
	}

	for _, tc := range cases {
		if got := AcceptParserVisitorResult(tc.parser, visitor); got != tc.want {
			t.Fatalf("%s visitor result = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestParserVisitorCallbackDispatchesUnsupported(t *testing.T) {
	called := ""
	AcceptParserVisitor[byte](nil, ParserVisitorFuncs[byte]{
		Unsupported: func(Parser[byte]) {
			called = "unsupported"
		},
	})
	if called != "unsupported" {
		t.Fatalf("unsupported visitor callback = %q", called)
	}
}

func TestPartialParserScansAndParsesFullAndPartialSequence(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	c := taggedByteLiteralParser('C', "c", "C")
	parser := Sequence(a, b, c)

	if got := ScanPartial(parser, NewTextSource("ABC"), 0); got != 3 {
		t.Fatalf("full ScanPartial = %d, want 3", got)
	}
	fullOutput := []any{}
	if got := ParsePartial(parser, NewTextSource("ABC"), 0, &fullOutput, 0); got != 3 || !reflect.DeepEqual(fullOutput, []any{"a", "b", "c"}) {
		t.Fatalf("full ParsePartial = %d/%#v, want 3/[a b c]", got, fullOutput)
	}

	failures := []string{}
	partialOutput := []any{"prefix"}
	handler := func(parser Parser[byte], output *[]any) {
		failures = append(failures, parser.Tag())
		appendOutput(output, "missing:"+parser.Tag())
	}
	if got := ParsePartial(parser, NewTextSource("ABx"), 0, &partialOutput, 1, handler); got != 2 {
		t.Fatalf("partial ParsePartial length = %d, want 2", got)
	}
	if !reflect.DeepEqual(partialOutput, []any{"prefix", "a", "b", "missing:C"}) {
		t.Fatalf("partial ParsePartial output = %#v", partialOutput)
	}
	if !reflect.DeepEqual(failures, []string{"C"}) {
		t.Fatalf("partial failures = %#v", failures)
	}
}

func TestPartialParserChoosesBestSuccessfulPathOnTie(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")
	c := taggedByteLiteralParser('C', "c", "C")
	partial := Sequence(a, b, c)
	complete := Sequence(a, b)
	parsers := []Parser[byte]{partial, complete}
	source := NewTextSource("ABx")

	if got := ScanPartialBest(parsers, source, 0); got != 2 {
		t.Fatalf("ScanPartialBest = %d, want 2", got)
	}

	var best []Parser[byte]
	failures := 0
	output := []any{}
	if got := ParsePartialBest(parsers, source, 0, &output, 0, &best, func(Parser[byte], *[]any) {
		failures++
	}); got != 2 {
		t.Fatalf("ParsePartialBest length = %d, want 2", got)
	}
	if !reflect.DeepEqual(output, []any{"a", "b"}) {
		t.Fatalf("ParsePartialBest output = %#v", output)
	}
	if failures != 0 {
		t.Fatalf("ParsePartialBest failure callbacks = %d, want 0", failures)
	}
	if len(best) != 2 || best[0] != partial || best[1] != complete {
		t.Fatalf("best parsers = %#v, want both max-length candidates", best)
	}

	if got := ScanPartialBest(parsers, source, 0); got != 2 {
		t.Fatalf("cached ScanPartialBest = %d, want 2", got)
	}
}

func TestPartialParserHandlesOptionalRepetitionAndApply(t *testing.T) {
	a := taggedByteLiteralParser('A', "a", "A")
	b := taggedByteLiteralParser('B', "b", "B")

	optional := OptionalValue(a, func() string { return "missing" })
	optionalOutput := []any{}
	if got := ParsePartial(optional, NewTextSource("Z"), 0, &optionalOutput, 0); got != 0 || !reflect.DeepEqual(optionalOutput, []any{"missing"}) {
		t.Fatalf("optional ParsePartial = %d/%#v, want 0/[missing]", got, optionalOutput)
	}

	repetition := OneOrMore(a)
	repetitionOutput := []any{}
	if got := ParsePartial(repetition, NewTextSource("AAZ"), 0, &repetitionOutput, 0); got != 2 || !reflect.DeepEqual(repetitionOutput, []any{"a", "a"}) {
		t.Fatalf("repetition ParsePartial = %d/%#v, want 2/[a a]", got, repetitionOutput)
	}

	apply := Apply(a, func(LeftValue[string]) RightParser[byte, string] {
		return RuleRight1(b, func(left string, right string) string { return left + right })
	})
	applyOutput := []any{}
	failures := 0
	if got := ParsePartial(apply, NewTextSource("AZ"), 0, &applyOutput, 0, func(Parser[byte], *[]any) {
		failures++
	}); got != 1 {
		t.Fatalf("apply ParsePartial length = %d, want 1", got)
	}
	if !reflect.DeepEqual(applyOutput, []any{"a"}) || failures != 1 {
		t.Fatalf("apply ParsePartial output/failures = %#v/%d, want [a]/1", applyOutput, failures)
	}
}

func TestParserFactoriesProduceOptionalRequiredAndRepetition(t *testing.T) {
	produced := ProduceElements(Sequence(newByteLiteralParser('A', "a"), newByteLiteralParser('B', "b")), func(values []string) string {
		return values[0] + values[1]
	})
	if result := produced.ParseValue(NewTextSource("AB"), 0); result.Length != 2 || result.Value != "ab" {
		t.Fatalf("ProduceElements result = %#v", result)
	}

	zeroOrMore := ZeroOrMoreProduce(newByteLiteralParser('A', "a"), func(values []string) int {
		return len(values)
	})
	if result := zeroOrMore.ParseValue(NewTextSource("AAx"), 0); result.Length != 2 || result.Value != 2 {
		t.Fatalf("ZeroOrMoreProduce result = %#v", result)
	}

	oneOrMore := OneOrMoreProduce(newByteLiteralParser('A', "a"), func(values []string) int {
		return len(values)
	})
	if result := oneOrMore.ParseValue(NewTextSource("xAA"), 0); result.Length != -1 || result.Value != 0 {
		t.Fatalf("OneOrMoreProduce failure = %#v", result)
	}

	optional := OptionalValue(newByteLiteralParser('A', "a"), func() string { return "missing" })
	if result := optional.ParseValue(NewTextSource("x"), 0); result.Length != 0 || result.Value != "missing" {
		t.Fatalf("OptionalValue result = %#v", result)
	}

	required := RequiredValue(newByteLiteralParser('A', "a"), func() string { return "required" })
	output := []any{}
	if n := required.ParseList(NewTextSource("x"), 0, &output, 0); n != 0 || len(output) != 1 || output[0] != "required" {
		t.Fatalf("RequiredValue length/output = %d/%#v", n, output)
	}

	value := Value[byte](func() string { return "constant" })
	output = []any{}
	if n := value.ParseList(NewTextSource("ignored"), 0, &output, 0); n != 0 || len(output) != 1 || output[0] != "constant" {
		t.Fatalf("Value list length/output = %d/%#v", n, output)
	}
}

func TestParserFactoriesListElements(t *testing.T) {
	list := ListElements(newByteLiteralParser('A', "a"), false, func(values []string) string {
		return fmt.Sprintf("%d:%s", len(values), strings.Join(values, ""))
	})
	if result := list.ParseValue(NewTextSource("AAx"), 0); result.Length != 2 || result.Value != "2:aa" {
		t.Fatalf("ListElements result = %#v", result)
	}
	if result := list.ParseValue(NewTextSource("x"), 0); result.Length != 0 || result.Value != "0:" {
		t.Fatalf("ListElements empty result = %#v", result)
	}

	required := ListElementsWithMissing(newByteLiteralParser('A', "a"), func(Source[byte], int) string {
		return "missing"
	}, true, func(values []string) string {
		return strings.Join(values, "|")
	})
	if result := required.ParseValue(NewTextSource("x"), 0); result.Length != 0 || result.Value != "missing" {
		t.Fatalf("ListElementsWithMissing result = %#v", result)
	}
}

func TestParserFactoriesSeparatedLists(t *testing.T) {
	element := newByteLiteralParser('A', "a")
	separator := newByteLiteralParser(',', ",")
	separated := ListSeparated(element, separator, true, func(values []ElementAndSeparator[string, string]) string {
		parts := make([]string, 0, len(values))
		for _, item := range values {
			parts = append(parts, item.Element+item.Separator)
		}
		return strings.Join(parts, "")
	})
	if result := separated.ParseValue(NewTextSource("A,A"), 0); result.Length != 3 || result.Value != "a,a" {
		t.Fatalf("ListSeparated result = %#v", result)
	}
	if result := separated.ParseValue(NewTextSource("x"), 0); result.Length != -1 || result.Value != "" {
		t.Fatalf("ListSeparated one-or-more failure = %#v", result)
	}

	trailing := ListSeparatedFull(
		element,
		separator,
		element,
		nil,
		nil,
		nil,
		nil,
		true,
		true,
		func(values []ElementAndSeparator[string, string]) string {
			parts := make([]string, 0, len(values))
			for _, item := range values {
				parts = append(parts, item.Element+item.Separator)
			}
			return strings.Join(parts, "")
		},
	)
	if result := trailing.ParseValue(NewTextSource("A,"), 0); result.Length != 2 || result.Value != "a," {
		t.Fatalf("ListSeparated trailing result = %#v", result)
	}
}

func TestParserFactoriesSeparatedListMissingSeparatorRecovery(t *testing.T) {
	element := newByteLiteralParser('A', "a")
	separator := newByteLiteralParser(',', ",")
	end := MatchText(";")
	list := ListSeparatedFull(
		element,
		separator,
		element,
		nil,
		func(Source[byte], int) string { return "<missing>" },
		nil,
		end,
		true,
		false,
		func(values []ElementAndSeparator[string, string]) string {
			parts := make([]string, 0, len(values))
			for _, item := range values {
				parts = append(parts, item.Element+item.Separator)
			}
			return strings.Join(parts, "")
		},
	)
	if result := list.ParseValue(NewTextSource("AA;"), 0); result.Length != 2 || result.Value != "a<missing>a" {
		t.Fatalf("missing separator list result = %#v", result)
	}
}

func newByteLiteralParser(literal byte, value string) *ResultPrimaryParser[byte, string] {
	return NewResultPrimaryParser(func(source Source[byte], start int) ParseResult[string] {
		if source.IsEnd(start) || source.Peek(start) != literal {
			return NewParseResult(-1, "")
		}
		return NewParseResult(1, value)
	}, WithParserTraits[byte](ParserTraits{Match: true}))
}

func taggedByteLiteralParser(literal byte, value string, tag string) ResultParser[byte, string] {
	return asResultParser[byte, string](newByteLiteralParser(literal, value).WithTag(tag))
}

func TestFirstParsersAllowGuardedZeroWidthRecoveryBeforeLaterAlternatives(t *testing.T) {
	required := Required(newByteLiteralParser('A', "a"), func(Source[byte], int) string { return "<missing>" })
	typed := FirstValue(required, newByteLiteralParser('B', "b"))
	if result := typed.ParseValue(NewTextSource("B"), 0); result.Length != 0 || result.Value != "<missing>" {
		t.Fatalf("typed zero-width first recovery = %#v", result)
	}

	var output []any
	untyped := First[byte](required, newByteLiteralParser('B', "b"))
	if length := untyped.ParseList(NewTextSource("B"), 0, &output, 0); length != 0 || len(output) != 1 || output[0] != "<missing>" {
		t.Fatalf("untyped zero-width first recovery = length:%d output:%#v", length, output)
	}
}
