package kustolang

import (
	"strings"
	"unicode/utf16"
)

// Ported from the parser-combinator factory surface in
// src/Kusto.Language/Parser/Combinators/Parsers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type LeftValue[TLeft any] struct{}

type MapValueEntry[TInput comparable, TOutput any] struct {
	Key   []TInput
	Value TOutput
}

type ElementAndSeparator[TElement any, TSeparator any] struct {
	Element   TElement
	Separator TSeparator
}

func NewElementAndSeparator[TElement any, TSeparator any](element TElement, separator ...TSeparator) ElementAndSeparator[TElement, TSeparator] {
	EnsureArgumentNotNil(element, "element")
	var sep TSeparator
	if len(separator) > 0 {
		sep = separator[0]
	}
	return ElementAndSeparator[TElement, TSeparator]{Element: element, Separator: sep}
}

func And[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return NewSequenceParser(parsers)
}

func Sequence[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return NewSequenceParser(parsers)
}

func Any[TInput any]() Parser[TInput] {
	return Match(func(TInput) bool { return true }).WithTag("<any>")
}

func Apply[TInput any, TLeft any, TOutput any](leftParser ResultParser[TInput, TLeft], fnRightParser func(LeftValue[TLeft]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(fnRightParser, "fnRightParser")
	return NewApplyParser(ApplyKindOne, leftParser, fnRightParser(LeftValue[TLeft]{}))
}

func ApplyOptional[TInput any, TOutput any](leftParser ResultParser[TInput, TOutput], fnRightParser func(LeftValue[TOutput]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(fnRightParser, "fnRightParser")
	return NewApplyParser(ApplyKindZeroOrOne, leftParser, fnRightParser(LeftValue[TOutput]{}))
}

func ApplyZeroOrMore[TInput any, TOutput any](leftParser ResultParser[TInput, TOutput], fnRightParser func(LeftValue[TOutput]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(fnRightParser, "fnRightParser")
	return NewApplyParser(ApplyKindZeroOrMore, leftParser, fnRightParser(LeftValue[TOutput]{}))
}

func Best[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return NewBestParser(parsers)
}

func BestValue[TInput any, TOutput any](parsers ...ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewBestValueParser(parsers)
}

func BestValueBy[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput], better func(TOutput, TOutput) bool) ResultParser[TInput, TOutput] {
	return NewBestValueParser(parsers, better)
}

func BestRight[TInput any, TOutput any](parsers ...RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	resultParsers := make([]ResultParser[TInput, TOutput], len(parsers))
	for i, parser := range parsers {
		resultParsers[i] = parser.Parser()
	}
	return NewRightParser[TInput, TOutput](NewBestValueParser(resultParsers))
}

// ResultParserCast is the Go spelling of Parser<TInput, TOutput>.Cast<TNewOutput>().
func ResultParserCast[TInput any, TOutput any, TNewOutput any](parser ResultParser[TInput, TOutput]) ResultParser[TInput, TNewOutput] {
	EnsureArgumentNotNil(parser, "parser")
	cast := Rule1(parser, func(value TOutput) TNewOutput {
		return any(value).(TNewOutput)
	})
	return asResultParser[TInput, TNewOutput](cast.WithTag(parser.Tag()))
}

func Convert[TInput any, TOutput any](pattern Parser[TInput], producer SourceProducer[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewConvertParser(pattern, producer)
}

func ConvertList[TInput any, TOutput any](pattern Parser[TInput], producer func([]TInput) TOutput) ResultParser[TInput, TOutput] {
	return NewConvertParserFromList(pattern, producer)
}

// ConvertString is the Go spelling of the Parsers<char>.Convert overload whose producer accepts a string.
func ConvertString[TOutput any](pattern Parser[byte], producer func(string) TOutput) ResultParser[byte, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return Convert(pattern, func(source Source[byte], start int, length int) TOutput {
		if textSource, ok := source.(*TextSource); ok {
			return producer(textSource.PeekText(start, length))
		}

		var builder strings.Builder
		builder.Grow(length)
		for i := 0; i < length; i++ {
			builder.WriteByte(source.Peek(start + i))
		}
		return producer(builder.String())
	})
}

// ConvertUTF16String is the exact UTF-16 code-unit adaptation of the
// Parsers<char>.Convert overload.
func ConvertUTF16String[TOutput any](pattern Parser[uint16], producer func(string) TOutput) ResultParser[uint16, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return Convert(pattern, func(source Source[uint16], start int, length int) TOutput {
		units := make([]uint16, length)
		for i := range units {
			units[i] = source.Peek(start + i)
		}
		return producer(string(utf16.Decode(units)))
	})
}

func ConvertSingle[TInput any, TOutput any](pattern Parser[TInput], producer func(TInput) TOutput) ResultParser[TInput, TOutput] {
	return NewConvertParserFromSingle(pattern, producer)
}

func ConvertValue[TInput any, TOutput any](pattern Parser[TInput], value TOutput) ResultParser[TInput, TOutput] {
	return ConvertSingle(pattern, func(TInput) TOutput { return value })
}

func Count[TInput any](scanner Parser[TInput]) ResultParser[TInput, int] {
	return Convert(scanner, func(source Source[TInput], start int, length int) int { return length })
}

func Fails[TInput any](parser Parser[TInput]) Parser[TInput] {
	return NewFailsParser(parser)
}

func First[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return NewFirstParser(parsers)
}

func FirstValue[TInput any, TOutput any](parsers ...ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewFirstValueParser(parsers)
}

func FirstRight[TInput any, TOutput any](parsers ...RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	resultParsers := make([]ResultParser[TInput, TOutput], len(parsers))
	for i, parser := range parsers {
		resultParsers[i] = parser.Parser()
	}
	return NewRightParser[TInput, TOutput](NewFirstValueParser(resultParsers))
}

func Forward[TInput any, TOutput any](deferredParser func() ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewForwardParser(deferredParser)
}

func ForwardValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(parser, "parser")
	return NewForwardParser(func() ResultParser[TInput, TOutput] { return parser })
}

func If[TInput any](test Parser[TInput], parser Parser[TInput]) Parser[TInput] {
	return NewIfParser(test, parser)
}

func IfValue[TInput any, TOutput any](test Parser[TInput], parser ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewIfValueParser(test, parser)
}

func IfRight[TInput any, TOutput any](test Parser[TInput], parser RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return NewRightParser[TInput, TOutput](NewIfValueParser(test, parser.Parser()))
}

func Limit[TInput any, TOutput any](limiter Parser[TInput], limited ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewLimitParser(limiter, limited)
}

func ListElements[TInput any, TElement any, TProducer any](
	elementParser ResultParser[TInput, TElement],
	oneOrMore bool,
	producer func([]TElement) TProducer,
) ResultParser[TInput, TProducer] {
	return ListElementsWithMissing(elementParser, nil, oneOrMore, producer)
}

func ListElementsWithMissing[TInput any, TElement any, TProducer any](
	elementParser ResultParser[TInput, TElement],
	fnMissingElement func(Source[TInput], int) TElement,
	oneOrMore bool,
	producer func([]TElement) TProducer,
) ResultParser[TInput, TProducer] {
	EnsureArgumentNotNil(elementParser, "elementParser")
	EnsureArgumentNotNil(producer, "producer")
	if oneOrMore {
		if fnMissingElement != nil {
			requiredElement := Required(elementParser, fnMissingElement)
			return ProduceElements(
				Sequence(requiredElement, ZeroOrMore(elementParser)),
				producer,
			)
		}
		return OneOrMoreProduce(elementParser, producer)
	}
	return ZeroOrMoreProduce(elementParser, producer)
}

func OList[TInput any, TElement any, TSeparator any, TProducer any](
	elementParser ResultParser[TInput, TElement],
	separatorParser ResultParser[TInput, TSeparator],
	fnMissingElement func(Source[TInput], int) TElement,
	fnMissingSeparator func(Source[TInput], int) TSeparator,
	endOfList Parser[TInput],
	oneOrMore bool,
	allowTrailingSeparator bool,
	producer func([]any) TProducer,
) ResultParser[TInput, TProducer] {
	return OListFull(
		elementParser,
		separatorParser,
		elementParser,
		fnMissingElement,
		fnMissingSeparator,
		fnMissingElement,
		endOfList,
		oneOrMore,
		allowTrailingSeparator,
		producer,
	)
}

func OListFull[TInput any, TElement any, TSeparator any, TProducer any](
	primaryElementParser ResultParser[TInput, TElement],
	separatorParser ResultParser[TInput, TSeparator],
	secondaryElementParser ResultParser[TInput, TElement],
	fnMissingPrimaryElement func(Source[TInput], int) TElement,
	fnMissingSeparator func(Source[TInput], int) TSeparator,
	fnMissingSecondaryElement func(Source[TInput], int) TElement,
	endOfList Parser[TInput],
	oneOrMore bool,
	allowTrailingSeparator bool,
	producer func([]any) TProducer,
) ResultParser[TInput, TProducer] {
	EnsureArgumentNotNil(primaryElementParser, "primaryElementParser")
	EnsureArgumentNotNil(separatorParser, "separatorParser")
	EnsureArgumentNotNil(producer, "producer")

	if listItemIsNil(secondaryElementParser) {
		secondaryElementParser = primaryElementParser
	}

	requiredPrimaryElementParser := primaryElementParser
	if fnMissingPrimaryElement != nil {
		requiredPrimaryElementParser = Required(primaryElementParser, fnMissingPrimaryElement)
	}
	requiredSecondaryElementParser := secondaryElementParser
	if fnMissingSecondaryElement != nil {
		requiredSecondaryElementParser = Required(secondaryElementParser, fnMissingSecondaryElement)
	}
	requiredSeparatorParser := separatorParser
	if fnMissingSeparator != nil {
		requiredSeparatorParser = Required(separatorParser, fnMissingSeparator)
	}
	emptyList := func() TProducer { return producer([]any{}) }

	if oneOrMore {
		if allowTrailingSeparator {
			secondaryParser := Sequence(separatorParser, secondaryElementParser)
			if fnMissingSeparator != nil && endOfList != nil {
				secondaryParser = First(
					secondaryParser,
					If(Not(endOfList), Sequence(requiredSeparatorParser, secondaryElementParser)).Hide(),
				)
			}
			return Produce(
				Sequence(
					requiredPrimaryElementParser,
					ZeroOrMore(secondaryParser),
					ZeroOrOne(separatorParser),
				),
				producer,
			)
		}

		secondaryParser := Sequence(separatorParser, requiredSecondaryElementParser)
		if fnMissingSeparator != nil && endOfList != nil {
			secondaryParser = First(
				secondaryParser,
				If(Not(endOfList), Sequence(requiredSeparatorParser, secondaryElementParser)).Hide(),
			)
		}
		return Produce(
			Sequence(
				requiredPrimaryElementParser,
				ZeroOrMore(secondaryParser),
			),
			producer,
		)
	}

	if allowTrailingSeparator {
		secondaryParser := Sequence(separatorParser, secondaryElementParser)
		if fnMissingSeparator != nil && endOfList != nil {
			secondaryParser = First(
				secondaryParser,
				If(Not(endOfList), Sequence(requiredSeparatorParser, secondaryElementParser)).Hide(),
			)
		}
		return OptionalValue(
			Produce(
				Sequence(
					First(
						If(separatorParser, requiredPrimaryElementParser).Hide(),
						primaryElementParser,
					),
					ZeroOrMore(secondaryParser),
					ZeroOrOne(separatorParser),
				),
				producer,
			),
			emptyList,
		)
	}

	secondaryParser := Sequence(separatorParser, requiredSecondaryElementParser)
	if fnMissingSeparator != nil && endOfList != nil {
		secondaryParser = First(
			secondaryParser,
			If(Not(endOfList), Sequence(requiredSeparatorParser, secondaryElementParser)).Hide(),
		)
	}
	return OptionalValue(
		Produce(
			Sequence(
				First(
					If(separatorParser, requiredPrimaryElementParser).Hide(),
					primaryElementParser,
				),
				ZeroOrMore(secondaryParser),
			),
			producer,
		),
		emptyList,
	)
}

func ListSeparated[TInput any, TElement any, TSeparator any, TProducer any](
	elementParser ResultParser[TInput, TElement],
	separatorParser ResultParser[TInput, TSeparator],
	oneOrMore bool,
	producer func([]ElementAndSeparator[TElement, TSeparator]) TProducer,
) ResultParser[TInput, TProducer] {
	return ListSeparatedFull(
		elementParser,
		separatorParser,
		elementParser,
		nil,
		nil,
		nil,
		nil,
		oneOrMore,
		false,
		producer,
	)
}

func ListSeparatedFull[TInput any, TElement any, TSeparator any, TProducer any](
	elementParser ResultParser[TInput, TElement],
	separatorParser ResultParser[TInput, TSeparator],
	secondaryElementParser ResultParser[TInput, TElement],
	fnMissingElement func(Source[TInput], int) TElement,
	fnMissingSeparator func(Source[TInput], int) TSeparator,
	fnMissingSecondaryElement func(Source[TInput], int) TElement,
	endOfList Parser[TInput],
	oneOrMore bool,
	allowTrailingSeparator bool,
	producer func([]ElementAndSeparator[TElement, TSeparator]) TProducer,
) ResultParser[TInput, TProducer] {
	EnsureArgumentNotNil(producer, "producer")
	return OListFull(
		elementParser,
		separatorParser,
		secondaryElementParser,
		fnMissingElement,
		fnMissingSeparator,
		fnMissingSecondaryElement,
		endOfList,
		oneOrMore,
		allowTrailingSeparator,
		func(output []any) TProducer {
			return produceElementAndSeparator(output, producer)
		},
	)
}

func MapValues[TInput comparable, TOutput any](entries []MapValueEntry[TInput, TOutput]) ResultParser[TInput, TOutput] {
	mapped := make([]MapEntry[TInput, TOutput], len(entries))
	for i, entry := range entries {
		value := entry.Value
		mapped[i] = MapEntry[TInput, TOutput]{
			Key:   entry.Key,
			Value: func() TOutput { return value },
		}
	}
	return NewMapParser(mapped)
}

func MapFuncs[TInput comparable, TOutput any](entries []MapEntry[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewMapParser(entries)
}

func Match[TInput any](predicate func(TInput) bool) Parser[TInput] {
	return NewPredicateMatchParser(predicate)
}

func MatchConsumer[TInput any](consumer SourceConsumer[TInput]) Parser[TInput] {
	return NewMatchParser(consumer)
}

func MatchItem[TInput comparable](item TInput) Parser[TInput] {
	return Match(func(input TInput) bool { return input == item })
}

func MatchItemWithComparer[TInput any](item TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	if comparer == nil {
		panic("kustolang: nil equality comparer")
	}
	return Match(func(input TInput) bool { return comparer(input, item) })
}

func MatchAny[TInput comparable](items ...TInput) Parser[TInput] {
	set := map[TInput]struct{}{}
	for _, item := range items {
		set[item] = struct{}{}
	}
	return Match(func(input TInput) bool {
		_, ok := set[input]
		return ok
	})
}

func MatchAnyWithComparer[TInput any](items []TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	if comparer == nil {
		panic("kustolang: nil equality comparer")
	}
	items = append([]TInput{}, items...)
	return Match(func(input TInput) bool {
		for _, item := range items {
			if comparer(input, item) {
				return true
			}
		}
		return false
	})
}

func MatchSequence[TInput comparable](items []TInput) Parser[TInput] {
	return MatchConsumer(func(source Source[TInput], start int) int {
		for i, item := range items {
			if source.IsEnd(start + i) {
				return ^i
			}
			if source.Peek(start+i) != item {
				return ^i
			}
		}
		return len(items)
	})
}

func MatchSequenceWithComparer[TInput any](items []TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	if comparer == nil {
		panic("kustolang: nil equality comparer")
	}
	items = append([]TInput{}, items...)
	return MatchConsumer(func(source Source[TInput], start int) int {
		for i, item := range items {
			if source.IsEnd(start+i) || !comparer(source.Peek(start+i), item) {
				return ^i
			}
		}
		return len(items)
	})
}

func MatchByte(ch byte, ignoreCase ...bool) Parser[byte] {
	if len(ignoreCase) > 0 && ignoreCase[0] {
		return Match(func(input byte) bool {
			return ordinalCaseKey(rune(input)) == ordinalCaseKey(rune(ch))
		})
	}
	return MatchItem(ch)
}

func MatchText(text string, ignoreCase ...bool) Parser[byte] {
	if len(text) == 1 {
		return MatchByte(text[0], ignoreCase...)
	}
	ignore := len(ignoreCase) > 0 && ignoreCase[0]
	if ignore {
		return MatchConsumer(func(source Source[byte], start int) int { return scanOrdinalText(source, start, text) })
	}
	return MatchConsumer(func(source Source[byte], start int) int {
		if textSource, ok := source.(*TextSource); ok {
			if textSource.Matches(start, text, ignore) {
				return len(text)
			}
			return -1
		}

		for i := 0; i < len(text); i++ {
			if source.IsEnd(start + i) {
				return ^i
			}
			got := source.Peek(start + i)
			if got != text[i] {
				return ^i
			}
		}
		return len(text)
	})
}

func MatchValue[TInput any, TOutput any](predicate func(TInput) bool, producer func(TInput) TOutput) ResultParser[TInput, TOutput] {
	return NewPredicateMatchValueParser(predicate, producer)
}

func MatchConsumerValue[TInput any, TOutput any](consumer SourceConsumer[TInput], producer SourceProducer[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return NewMatchValueParser(consumer, producer)
}

func Not[TInput any](parser Parser[TInput]) Parser[TInput] {
	return NewNotParser(parser)
}

func OneOrMore[TInput any](parser Parser[TInput]) Parser[TInput] {
	return NewOneOrMoreParser(parser)
}

func OneOrMoreProduce[TInput any, TElement any, TOutput any](parser ResultParser[TInput, TElement], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return ProduceElements(OneOrMore(parser), producer)
}

func OptionalValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer ...func() TOutput) ResultParser[TInput, TOutput] {
	if len(producer) > 0 {
		return NewOptionalParser(parser, producer[0])
	}
	return NewOptionalParser(parser, func() TOutput {
		var zero TOutput
		return zero
	})
}

func Or[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return NewBestParser(parsers)
}

func Produce[TInput any, TOutput any](parser Parser[TInput], producer func([]any) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewProduceParser(parser, func(output []any, outputStart int) TOutput {
		return producer(output[outputStart:])
	})
}

func ProduceElements[TInput any, TElement any, TOutput any](parser Parser[TInput], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewProduceParser(parser, func(output []any, outputStart int) TOutput {
		elements := make([]TElement, 0, len(output)-outputStart)
		for i := outputStart; i < len(output); i++ {
			elements = append(elements, output[i].(TElement))
		}
		return producer(elements)
	})
}

func Required[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func(Source[TInput], int) TOutput) ResultParser[TInput, TOutput] {
	return NewRequiredParser(parser, producer)
}

func RequiredValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func() TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRequiredParser(parser, func(Source[TInput], int) TOutput { return producer() })
}

func ruleOutputValue[T any](output []any, index int) T {
	if output[index] == nil {
		var zero T
		return zero
	}
	return output[index].(T)
}

func Rule1[TInput any, TParser1 any, TOutput any](parser1 ResultParser[TInput, TParser1], producer func(TParser1) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1},
		func(output []any, start int) TOutput {
			EnsureAreEqual(1, len(output)-start)
			return producer(ruleOutputValue[TParser1](output, start))
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			return NewParseResult(result1.Length, producer(result1.Value))
		},
	)
}

func Rule2[TInput any, TParser1 any, TParser2 any, TOutput any](parser1 ResultParser[TInput, TParser1], parser2 ResultParser[TInput, TParser2], producer func(TParser1, TParser2) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2},
		func(output []any, start int) TOutput {
			EnsureAreEqual(2, len(output)-start)
			return producer(ruleOutputValue[TParser1](output, start), ruleOutputValue[TParser2](output, start+1))
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			return NewParseResult(length, producer(result1.Value, result2.Value))
		},
	)
}

func Rule3[TInput any, TParser1 any, TParser2 any, TParser3 any, TOutput any](parser1 ResultParser[TInput, TParser1], parser2 ResultParser[TInput, TParser2], parser3 ResultParser[TInput, TParser3], producer func(TParser1, TParser2, TParser3) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3},
		func(output []any, start int) TOutput {
			EnsureAreEqual(3, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value))
		},
	)
}

func Rule4[TInput any, TParser1 any, TParser2 any, TParser3 any, TParser4 any, TOutput any](parser1 ResultParser[TInput, TParser1], parser2 ResultParser[TInput, TParser2], parser3 ResultParser[TInput, TParser3], parser4 ResultParser[TInput, TParser4], producer func(TParser1, TParser2, TParser3, TParser4) TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4},
		func(output []any, start int) TOutput {
			EnsureAreEqual(4, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value))
		},
	)
}

func Rule5[
	TInput any,
	TParser1 any,
	TParser2 any,
	TParser3 any,
	TParser4 any,
	TParser5 any,
	TOutput any,
](
	parser1 ResultParser[TInput, TParser1],
	parser2 ResultParser[TInput, TParser2],
	parser3 ResultParser[TInput, TParser3],
	parser4 ResultParser[TInput, TParser4],
	parser5 ResultParser[TInput, TParser5],
	producer func(TParser1, TParser2, TParser3, TParser4, TParser5) TOutput,
) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4, parser5},
		func(output []any, start int) TOutput {
			EnsureAreEqual(5, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
				ruleOutputValue[TParser5](output, start+4),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			result5 := parser5.ParseValue(source, start+length)
			if result5.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result5.Length, zero)
			}
			length += result5.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value, result5.Value))
		},
	)
}

func Rule6[
	TInput any,
	TParser1 any,
	TParser2 any,
	TParser3 any,
	TParser4 any,
	TParser5 any,
	TParser6 any,
	TOutput any,
](
	parser1 ResultParser[TInput, TParser1],
	parser2 ResultParser[TInput, TParser2],
	parser3 ResultParser[TInput, TParser3],
	parser4 ResultParser[TInput, TParser4],
	parser5 ResultParser[TInput, TParser5],
	parser6 ResultParser[TInput, TParser6],
	producer func(TParser1, TParser2, TParser3, TParser4, TParser5, TParser6) TOutput,
) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4, parser5, parser6},
		func(output []any, start int) TOutput {
			EnsureAreEqual(6, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
				ruleOutputValue[TParser5](output, start+4),
				ruleOutputValue[TParser6](output, start+5),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			result5 := parser5.ParseValue(source, start+length)
			if result5.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result5.Length, zero)
			}
			length += result5.Length

			result6 := parser6.ParseValue(source, start+length)
			if result6.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result6.Length, zero)
			}
			length += result6.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value, result5.Value, result6.Value))
		},
	)
}

func Rule7[
	TInput any,
	TParser1 any,
	TParser2 any,
	TParser3 any,
	TParser4 any,
	TParser5 any,
	TParser6 any,
	TParser7 any,
	TOutput any,
](
	parser1 ResultParser[TInput, TParser1],
	parser2 ResultParser[TInput, TParser2],
	parser3 ResultParser[TInput, TParser3],
	parser4 ResultParser[TInput, TParser4],
	parser5 ResultParser[TInput, TParser5],
	parser6 ResultParser[TInput, TParser6],
	parser7 ResultParser[TInput, TParser7],
	producer func(TParser1, TParser2, TParser3, TParser4, TParser5, TParser6, TParser7) TOutput,
) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4, parser5, parser6, parser7},
		func(output []any, start int) TOutput {
			EnsureAreEqual(7, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
				ruleOutputValue[TParser5](output, start+4),
				ruleOutputValue[TParser6](output, start+5),
				ruleOutputValue[TParser7](output, start+6),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			result5 := parser5.ParseValue(source, start+length)
			if result5.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result5.Length, zero)
			}
			length += result5.Length

			result6 := parser6.ParseValue(source, start+length)
			if result6.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result6.Length, zero)
			}
			length += result6.Length

			result7 := parser7.ParseValue(source, start+length)
			if result7.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result7.Length, zero)
			}
			length += result7.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value, result5.Value, result6.Value, result7.Value))
		},
	)
}

func Rule8[
	TInput any,
	TParser1 any,
	TParser2 any,
	TParser3 any,
	TParser4 any,
	TParser5 any,
	TParser6 any,
	TParser7 any,
	TParser8 any,
	TOutput any,
](
	parser1 ResultParser[TInput, TParser1],
	parser2 ResultParser[TInput, TParser2],
	parser3 ResultParser[TInput, TParser3],
	parser4 ResultParser[TInput, TParser4],
	parser5 ResultParser[TInput, TParser5],
	parser6 ResultParser[TInput, TParser6],
	parser7 ResultParser[TInput, TParser7],
	parser8 ResultParser[TInput, TParser8],
	producer func(TParser1, TParser2, TParser3, TParser4, TParser5, TParser6, TParser7, TParser8) TOutput,
) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4, parser5, parser6, parser7, parser8},
		func(output []any, start int) TOutput {
			EnsureAreEqual(8, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
				ruleOutputValue[TParser5](output, start+4),
				ruleOutputValue[TParser6](output, start+5),
				ruleOutputValue[TParser7](output, start+6),
				ruleOutputValue[TParser8](output, start+7),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			result5 := parser5.ParseValue(source, start+length)
			if result5.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result5.Length, zero)
			}
			length += result5.Length

			result6 := parser6.ParseValue(source, start+length)
			if result6.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result6.Length, zero)
			}
			length += result6.Length

			result7 := parser7.ParseValue(source, start+length)
			if result7.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result7.Length, zero)
			}
			length += result7.Length

			result8 := parser8.ParseValue(source, start+length)
			if result8.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result8.Length, zero)
			}
			length += result8.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value, result5.Value, result6.Value, result7.Value, result8.Value))
		},
	)
}

func Rule9[
	TInput any,
	TParser1 any,
	TParser2 any,
	TParser3 any,
	TParser4 any,
	TParser5 any,
	TParser6 any,
	TParser7 any,
	TParser8 any,
	TParser9 any,
	TOutput any,
](
	parser1 ResultParser[TInput, TParser1],
	parser2 ResultParser[TInput, TParser2],
	parser3 ResultParser[TInput, TParser3],
	parser4 ResultParser[TInput, TParser4],
	parser5 ResultParser[TInput, TParser5],
	parser6 ResultParser[TInput, TParser6],
	parser7 ResultParser[TInput, TParser7],
	parser8 ResultParser[TInput, TParser8],
	parser9 ResultParser[TInput, TParser9],
	producer func(TParser1, TParser2, TParser3, TParser4, TParser5, TParser6, TParser7, TParser8, TParser9) TOutput,
) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRuleParser(
		[]Parser[TInput]{parser1, parser2, parser3, parser4, parser5, parser6, parser7, parser8, parser9},
		func(output []any, start int) TOutput {
			EnsureAreEqual(9, len(output)-start)
			return producer(
				ruleOutputValue[TParser1](output, start),
				ruleOutputValue[TParser2](output, start+1),
				ruleOutputValue[TParser3](output, start+2),
				ruleOutputValue[TParser4](output, start+3),
				ruleOutputValue[TParser5](output, start+4),
				ruleOutputValue[TParser6](output, start+5),
				ruleOutputValue[TParser7](output, start+6),
				ruleOutputValue[TParser8](output, start+7),
				ruleOutputValue[TParser9](output, start+8),
			)
		},
		func(source Source[TInput], start int) ParseResult[TOutput] {
			result1 := parser1.ParseValue(source, start)
			if result1.Length < 0 {
				var zero TOutput
				return NewParseResult(result1.Length, zero)
			}
			length := result1.Length

			result2 := parser2.ParseValue(source, start+length)
			if result2.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result2.Length, zero)
			}
			length += result2.Length

			result3 := parser3.ParseValue(source, start+length)
			if result3.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result3.Length, zero)
			}
			length += result3.Length

			result4 := parser4.ParseValue(source, start+length)
			if result4.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result4.Length, zero)
			}
			length += result4.Length

			result5 := parser5.ParseValue(source, start+length)
			if result5.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result5.Length, zero)
			}
			length += result5.Length

			result6 := parser6.ParseValue(source, start+length)
			if result6.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result6.Length, zero)
			}
			length += result6.Length

			result7 := parser7.ParseValue(source, start+length)
			if result7.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result7.Length, zero)
			}
			length += result7.Length

			result8 := parser8.ParseValue(source, start+length)
			if result8.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result8.Length, zero)
			}
			length += result8.Length

			result9 := parser9.ParseValue(source, start+length)
			if result9.Length < 0 {
				var zero TOutput
				return NewParseResult(-length+result9.Length, zero)
			}
			length += result9.Length

			return NewParseResult(length, producer(result1.Value, result2.Value, result3.Value, result4.Value, result5.Value, result6.Value, result7.Value, result8.Value, result9.Value))
		},
	)
}

func RuleRight0[TInput any, TLeft any, TOutput any](producer func(TLeft) TOutput) RightParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRightParser[TInput, TOutput](
		NewRuleParser(
			[]Parser[TInput]{},
			func(output []any, start int) TOutput {
				EnsureAreEqual(1, len(output)-start)
				return producer(ruleOutputValue[TLeft](output, start))
			},
		),
	)
}

func RuleRight1[TInput any, TLeft any, TParser1 any, TOutput any](parser1 ResultParser[TInput, TParser1], producer func(TLeft, TParser1) TOutput) RightParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRightParser[TInput, TOutput](
		NewRuleParser(
			[]Parser[TInput]{parser1},
			func(output []any, start int) TOutput {
				EnsureAreEqual(2, len(output)-start)
				return producer(ruleOutputValue[TLeft](output, start), ruleOutputValue[TParser1](output, start+1))
			},
		),
	)
}

func RuleRight2[TInput any, TLeft any, TParser1 any, TParser2 any, TOutput any](parser1 ResultParser[TInput, TParser1], parser2 ResultParser[TInput, TParser2], producer func(TLeft, TParser1, TParser2) TOutput) RightParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRightParser[TInput, TOutput](
		NewRuleParser(
			[]Parser[TInput]{parser1, parser2},
			func(output []any, start int) TOutput {
				EnsureAreEqual(3, len(output)-start)
				return producer(
					ruleOutputValue[TLeft](output, start),
					ruleOutputValue[TParser1](output, start+1),
					ruleOutputValue[TParser2](output, start+2),
				)
			},
		),
	)
}

func RuleRight3[TInput any, TLeft any, TParser1 any, TParser2 any, TParser3 any, TOutput any](parser1 ResultParser[TInput, TParser1], parser2 ResultParser[TInput, TParser2], parser3 ResultParser[TInput, TParser3], producer func(TLeft, TParser1, TParser2, TParser3) TOutput) RightParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return NewRightParser[TInput, TOutput](
		NewRuleParser(
			[]Parser[TInput]{parser1, parser2, parser3},
			func(output []any, start int) TOutput {
				EnsureAreEqual(4, len(output)-start)
				return producer(
					ruleOutputValue[TLeft](output, start),
					ruleOutputValue[TParser1](output, start+1),
					ruleOutputValue[TParser2](output, start+2),
					ruleOutputValue[TParser3](output, start+3),
				)
			},
		),
	)
}

func Text(pattern Parser[byte]) ResultParser[byte, string] {
	return Convert(pattern, func(source Source[byte], start int, length int) string {
		if textSource, ok := source.(*TextSource); ok {
			return textSource.PeekText(start, length)
		}
		values := make([]byte, length)
		for i := 0; i < length; i++ {
			values[i] = source.Peek(start + i)
		}
		return string(values)
	})
}

func TextLiteral(text string, ignoreCase ...bool) ResultParser[byte, string] {
	return ConvertValue(MatchText(text, ignoreCase...), text)
}

func TextAndOffset(pattern Parser[byte]) ResultParser[byte, OffsetValue[string]] {
	return Convert(pattern, func(source Source[byte], start int, length int) OffsetValue[string] {
		if textSource, ok := source.(*TextSource); ok {
			return NewOffsetValue(start, textSource.PeekText(start, length))
		}
		values := make([]byte, length)
		for i := 0; i < length; i++ {
			values[i] = source.Peek(start + i)
		}
		return NewOffsetValue(start, string(values))
	})
}

func Value[TInput any, TOutput any](fnValue func() TOutput) ResultParser[TInput, TOutput] {
	EnsureArgumentNotNil(fnValue, "fnValue")
	return MatchConsumerValue(
		func(Source[TInput], int) int { return 0 },
		func(Source[TInput], int, int) TOutput { return fnValue() },
	)
}

func ZeroOrMore[TInput any](parser Parser[TInput]) Parser[TInput] {
	return NewZeroOrMoreParser(parser)
}

func ZeroOrMoreProduce[TInput any, TElement any, TOutput any](parser ResultParser[TInput, TElement], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return ProduceElements(ZeroOrMore(parser), producer)
}

func ZeroOrOne[TInput any](parser Parser[TInput]) Parser[TInput] {
	return NewZeroOrMoreParser(parser, true)
}

func produceElementAndSeparator[TElement any, TSeparator any, TProducer any](
	output []any,
	producer func([]ElementAndSeparator[TElement, TSeparator]) TProducer,
) TProducer {
	list := make([]ElementAndSeparator[TElement, TSeparator], 0, (len(output)+1)/2)
	for i := 0; i < len(output); i += 2 {
		element := output[i].(TElement)
		var separator TSeparator
		if i < len(output)-1 {
			separator = output[i+1].(TSeparator)
		}
		list = append(list, NewElementAndSeparator(element, separator))
	}
	return producer(list)
}
