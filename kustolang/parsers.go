package kustolang

// Ported from src/Kusto.Language/Parser/Combinators.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type Parser[TInput any] = internal.Parser[TInput]
type ResultParser[TInput any, TOutput any] = internal.ResultParser[TInput, TOutput]
type RightParser[TInput any, TOutput any] = internal.RightParser[TInput, TOutput]
type ParseResult[T any] = internal.ParseResult[T]
type OffsetValue[T any] = internal.OffsetValue[T]
type Source[T any] = internal.Source[T]
type SourceConsumer[TInput any] = internal.SourceConsumer[TInput]
type SourceProducer[TInput any, TOutput any] = internal.SourceProducer[TInput, TOutput]
type SourceCache = internal.SourceCache
type ArraySource[T any] = internal.ArraySource[T]
type LimitSource[T any] = internal.LimitSource[T]
type TextSource = internal.TextSource
type LeftValue[TLeft any] = internal.LeftValue[TLeft]
type MapEntry[TInput comparable, TOutput any] = internal.MapEntry[TInput, TOutput]
type MapValueEntry[TInput comparable, TOutput any] = internal.MapValueEntry[TInput, TOutput]
type ElementAndSeparator[TElement any, TSeparator any] = internal.ElementAndSeparator[TElement, TSeparator]
type PartialParserFailureHandler[TInput any] = internal.PartialParserFailureHandler[TInput]
type ApplyKind = internal.ApplyKind
type ParserTraits = internal.ParserTraits
type ParserOption[TInput any] = internal.ParserOption[TInput]
type ParserVisitorFuncs[TInput any] = internal.ParserVisitorFuncs[TInput]
type ParserVisitorResultFuncs[TInput any, TResult any] = internal.ParserVisitorResultFuncs[TInput, TResult]
type ParserVisitorArgResultFuncs[TInput any, TArg any, TResult any] = internal.ParserVisitorArgResultFuncs[TInput, TArg, TResult]
type IsParentVisitor = ParserVisitorArgResultFuncs[LexicalToken, Parser[LexicalToken], bool]

// IsParentVisitorInstance is the Go spelling of IsParentVisitor.Instance.
var IsParentVisitorInstance = newIsParentVisitor()

var (
	CharScannerLetter     Parser[byte] = internal.Letter
	CharScannerDigit      Parser[byte] = internal.Digit
	CharScannerHexDigit   Parser[byte] = internal.HexDigit
	CharScannerWhitespace Parser[byte] = internal.Whitespace
	CharScannerLineBreak  Parser[byte] = internal.LineBreak
)

func newIsParentVisitor() IsParentVisitor {
	isParent := func(parser Parser[LexicalToken], child Parser[LexicalToken]) bool {
		return internal.IsParserParent(parser, child)
	}
	return IsParentVisitor{
		Apply: func(parser, _, _ Parser[LexicalToken], _ ApplyKind, child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Best: func(parser Parser[LexicalToken], _ []Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Convert: func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Fails:   func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		First: func(parser Parser[LexicalToken], _ []Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Forward: func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		If: func(parser, _, _ Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Limit: func(parser, _, _ Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Map:       func(parser Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Match:     func(parser Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Not:       func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		OneOrMore: func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Optional:  func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Produce:   func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Required:  func(parser, _ Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
		Rule: func(parser Parser[LexicalToken], _ []Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Sequence: func(parser Parser[LexicalToken], _ []Parser[LexicalToken], child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		ZeroOrMore: func(parser, _ Parser[LexicalToken], _ bool, child Parser[LexicalToken]) bool {
			return isParent(parser, child)
		},
		Unsupported: func(parser Parser[LexicalToken], child Parser[LexicalToken]) bool { return isParent(parser, child) },
	}
}

// Concrete parser aliases preserve the source hierarchy while using distinct
// names for C# generic-arity overloads, which Go does not support.
type MatchParser[TInput any] = internal.MatchParser[TInput]
type MatchValueParser[TInput any, TOutput any] = internal.MatchValueParser[TInput, TOutput]
type ConvertParser[TInput any, TOutput any] = internal.ConvertParser[TInput, TOutput]
type SequenceParser[TInput any] = internal.SequenceParser[TInput]
type FirstParser[TInput any] = internal.FirstParser[TInput]
type FirstValueParser[TInput any, TOutput any] = internal.FirstValueParser[TInput, TOutput]
type BestParser[TInput any] = internal.BestParser[TInput]
type BestValueParser[TInput any, TOutput any] = internal.BestValueParser[TInput, TOutput]
type RuleParser[TInput any, TOutput any] = internal.RuleParser[TInput, TOutput]
type MapParser[TInput comparable, TOutput any] = internal.MapParser[TInput, TOutput]
type FailsParser[TInput any] = internal.FailsParser[TInput]
type NotParser[TInput any] = internal.NotParser[TInput]
type ProduceParser[TInput any, TOutput any] = internal.ProduceParser[TInput, TOutput]
type RequiredParser[TInput any, TOutput any] = internal.RequiredParser[TInput, TOutput]
type OptionalParser[TInput any, TOutput any] = internal.OptionalParser[TInput, TOutput]
type LimitParser[TInput any, TOutput any] = internal.LimitParser[TInput, TOutput]
type ZeroOrMoreParser[TInput any] = internal.ZeroOrMoreParser[TInput]
type OneOrMoreParser[TInput any] = internal.OneOrMoreParser[TInput]
type IfParser[TInput any] = internal.IfParser[TInput]
type IfValueParser[TInput any, TOutput any] = internal.IfValueParser[TInput, TOutput]
type ApplyParser[TInput any, TLeft any, TOutput any] = internal.ApplyParser[TInput, TLeft, TOutput]
type ForwardParser[TInput any, TOutput any] = internal.ForwardParser[TInput, TOutput]
type ResultPrimaryParser[TInput any, TOutput any] = internal.ResultPrimaryParser[TInput, TOutput]
type ListPrimaryParser[TInput any, TOutput any] = internal.ListPrimaryParser[TInput, TOutput]

const (
	ApplyKindOne        = internal.ApplyKindOne
	ApplyKindZeroOrOne  = internal.ApplyKindZeroOrOne
	ApplyKindZeroOrMore = internal.ApplyKindZeroOrMore
)

func NewParseResult[T any](length int, value T) ParseResult[T] {
	return internal.NewParseResult(length, value)
}

func NewOffsetValue[T any](offset int, value T) OffsetValue[T] {
	return internal.NewOffsetValue(offset, value)
}

func NewArraySource[T any](input []T, startAndLength ...int) *ArraySource[T] {
	return internal.NewArraySource(input, startAndLength...)
}

func NewLimitSource[T any](source Source[T], limit int) *LimitSource[T] {
	return internal.NewLimitSource(source, limit)
}

func NewSourceCache() *SourceCache { return internal.NewSourceCache() }

func SourceCacheGetOrCreate[T any](cache *SourceCache) T {
	return internal.SourceCacheGetOrCreate[T](cache)
}

func SourceCacheGetOrCreateFunc[T any](cache *SourceCache, creator func() T) T {
	return internal.SourceCacheGetOrCreateFunc(cache, creator)
}

func SourceCacheTryGetValue[T any](cache *SourceCache) (T, bool) {
	return internal.SourceCacheTryGetValue[T](cache)
}

func NewTextSource(source string, offsetAndLength ...int) *TextSource {
	return internal.NewTextSource(source, offsetAndLength...)
}

func CharScannerChars(text string, ignoreCase ...bool) Parser[byte] {
	return internal.Chars(text, ignoreCase...)
}

func CharScannerChar(ch rune, ignoreCase ...bool) Parser[byte] {
	return internal.Char(ch, ignoreCase...)
}

func ScannerMatches(scanner Parser[byte], text string, offsetAndLength ...int) bool {
	return internal.ScannerMatches(scanner, text, offsetAndLength...)
}

func NewElementAndSeparator[TElement any, TSeparator any](element TElement, separator ...TSeparator) ElementAndSeparator[TElement, TSeparator] {
	return internal.NewElementAndSeparator(element, separator...)
}

func NewRightParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return internal.NewRightParser(parser)
}

func WithParserTag[TInput any](tag string) ParserOption[TInput] {
	return internal.WithParserTag[TInput](tag)
}

func WithParserAnnotations[TInput any](annotations []any) ParserOption[TInput] {
	return internal.WithParserAnnotations[TInput](annotations)
}

func WithParserDescription[TInput any](description string) ParserOption[TInput] {
	return internal.WithParserDescription[TInput](description)
}

func WithParserHidden[TInput any](hidden bool) ParserOption[TInput] {
	return internal.WithParserHidden[TInput](hidden)
}

func WithParserTraits[TInput any](traits ParserTraits) ParserOption[TInput] {
	return internal.WithParserTraits[TInput](traits)
}

func WithParserChildren[TInput any](children ...Parser[TInput]) ParserOption[TInput] {
	return internal.WithParserChildren(children...)
}

func WithParserScan[TInput any](scan func(Source[TInput], int) int) ParserOption[TInput] {
	return internal.WithParserScan(scan)
}

func NewResultPrimaryParser[TInput any, TOutput any](parse func(Source[TInput], int) ParseResult[TOutput], options ...ParserOption[TInput]) *ResultPrimaryParser[TInput, TOutput] {
	return internal.NewResultPrimaryParser(parse, options...)
}

func NewListPrimaryParser[TInput any, TOutput any](parse func(Source[TInput], int, *[]any, int) int, options ...ParserOption[TInput]) *ListPrimaryParser[TInput, TOutput] {
	return internal.NewListPrimaryParser[TInput, TOutput](parse, options...)
}

func NewMatchParser[TInput any](consumer SourceConsumer[TInput]) *MatchParser[TInput] {
	return internal.NewMatchParser(consumer)
}

func NewPredicateMatchParser[TInput any](predicate func(TInput) bool) *MatchParser[TInput] {
	return internal.NewPredicateMatchParser(predicate)
}

func NewMatchValueParser[TInput any, TOutput any](consumer SourceConsumer[TInput], producer SourceProducer[TInput, TOutput]) *MatchValueParser[TInput, TOutput] {
	return internal.NewMatchValueParser(consumer, producer)
}

func NewPredicateMatchValueParser[TInput any, TOutput any](predicate func(TInput) bool, producer func(TInput) TOutput) *MatchValueParser[TInput, TOutput] {
	return internal.NewPredicateMatchValueParser(predicate, producer)
}

func NewConvertParser[TInput any, TOutput any](pattern Parser[TInput], producer SourceProducer[TInput, TOutput]) *ConvertParser[TInput, TOutput] {
	return internal.NewConvertParser(pattern, producer)
}

func NewConvertParserFromList[TInput any, TOutput any](pattern Parser[TInput], producer func([]TInput) TOutput) *ConvertParser[TInput, TOutput] {
	return internal.NewConvertParserFromList(pattern, producer)
}

func NewConvertParserFromSingle[TInput any, TOutput any](pattern Parser[TInput], producer func(TInput) TOutput) *ConvertParser[TInput, TOutput] {
	return internal.NewConvertParserFromSingle(pattern, producer)
}

func NewSequenceParser[TInput any](parsers []Parser[TInput]) *SequenceParser[TInput] {
	return internal.NewSequenceParser(parsers)
}

func NewFirstParser[TInput any](parsers []Parser[TInput]) *FirstParser[TInput] {
	return internal.NewFirstParser(parsers)
}

func NewFirstValueParser[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput]) *FirstValueParser[TInput, TOutput] {
	return internal.NewFirstValueParser(parsers)
}

func NewBestParser[TInput any](parsers []Parser[TInput]) *BestParser[TInput] {
	return internal.NewBestParser(parsers)
}

func NewBestValueParser[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput], isBetter ...func(TOutput, TOutput) bool) *BestValueParser[TInput, TOutput] {
	return internal.NewBestValueParser(parsers, isBetter...)
}

func NewRuleParser[TInput any, TOutput any](parsers []Parser[TInput], producer func([]any, int) TOutput, resultProducer ...func(Source[TInput], int) ParseResult[TOutput]) *RuleParser[TInput, TOutput] {
	return internal.NewRuleParser(parsers, producer, resultProducer...)
}

func NewMapParser[TInput comparable, TOutput any](entries []MapEntry[TInput, TOutput]) *MapParser[TInput, TOutput] {
	return internal.NewMapParser(entries)
}

func NewFailsParser[TInput any](pattern Parser[TInput]) *FailsParser[TInput] {
	return internal.NewFailsParser(pattern)
}

func NewNotParser[TInput any](pattern Parser[TInput]) *NotParser[TInput] {
	return internal.NewNotParser(pattern)
}

func NewProduceParser[TInput any, TOutput any](parser Parser[TInput], producer func([]any, int) TOutput) *ProduceParser[TInput, TOutput] {
	return internal.NewProduceParser(parser, producer)
}

func NewRequiredParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func(Source[TInput], int) TOutput) *RequiredParser[TInput, TOutput] {
	return internal.NewRequiredParser(parser, producer)
}

func NewOptionalParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func() TOutput) *OptionalParser[TInput, TOutput] {
	return internal.NewOptionalParser(parser, producer)
}

func NewLimitParser[TInput any, TOutput any](limiter Parser[TInput], limited ResultParser[TInput, TOutput]) *LimitParser[TInput, TOutput] {
	return internal.NewLimitParser(limiter, limited)
}

func NewZeroOrMoreParser[TInput any](parser Parser[TInput], zeroOrOne ...bool) *ZeroOrMoreParser[TInput] {
	return internal.NewZeroOrMoreParser(parser, zeroOrOne...)
}

func NewOneOrMoreParser[TInput any](parser Parser[TInput]) *OneOrMoreParser[TInput] {
	return internal.NewOneOrMoreParser(parser)
}

func NewIfParser[TInput any](test Parser[TInput], parser Parser[TInput]) *IfParser[TInput] {
	return internal.NewIfParser(test, parser)
}

func NewIfValueParser[TInput any, TOutput any](test Parser[TInput], parser ResultParser[TInput, TOutput]) *IfValueParser[TInput, TOutput] {
	return internal.NewIfValueParser(test, parser)
}

func NewApplyParser[TInput any, TLeft any, TOutput any](kind ApplyKind, left ResultParser[TInput, TLeft], right RightParser[TInput, TOutput]) *ApplyParser[TInput, TLeft, TOutput] {
	return internal.NewApplyParser(kind, left, right)
}

func NewForwardParser[TInput any, TOutput any](deferred func() ResultParser[TInput, TOutput]) *ForwardParser[TInput, TOutput] {
	return internal.NewForwardParser(deferred)
}

func And[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return internal.And(parsers...)
}

func Sequence[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return internal.Sequence(parsers...)
}

func Any[TInput any]() Parser[TInput] { return internal.Any[TInput]() }

func Apply[TInput any, TLeft any, TOutput any](left ResultParser[TInput, TLeft], right func(LeftValue[TLeft]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.Apply(left, right)
}

func ApplyOptional[TInput any, TOutput any](left ResultParser[TInput, TOutput], right func(LeftValue[TOutput]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.ApplyOptional(left, right)
}

func ApplyZeroOrMore[TInput any, TOutput any](left ResultParser[TInput, TOutput], right func(LeftValue[TOutput]) RightParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.ApplyZeroOrMore(left, right)
}

func Best[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return internal.Best(parsers...)
}

func BestValue[TInput any, TOutput any](parsers ...ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.BestValue(parsers...)
}

func BestValueBy[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput], better func(TOutput, TOutput) bool) ResultParser[TInput, TOutput] {
	return internal.BestValueBy(parsers, better)
}

func BestRight[TInput any, TOutput any](parsers ...RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return internal.BestRight(parsers...)
}

// ResultParserCast is the Go spelling of Parser<TInput, TOutput>.Cast<TNewOutput>().
func ResultParserCast[TInput any, TOutput any, TNewOutput any](parser ResultParser[TInput, TOutput]) ResultParser[TInput, TNewOutput] {
	return internal.ResultParserCast[TInput, TOutput, TNewOutput](parser)
}

func Convert[TInput any, TOutput any](pattern Parser[TInput], producer SourceProducer[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.Convert(pattern, producer)
}

func ConvertList[TInput any, TOutput any](pattern Parser[TInput], producer func([]TInput) TOutput) ResultParser[TInput, TOutput] {
	return internal.ConvertList(pattern, producer)
}

// ConvertString is the Go spelling of the Parsers<char>.Convert overload whose producer accepts a string.
func ConvertString[TOutput any](pattern Parser[byte], producer func(string) TOutput) ResultParser[byte, TOutput] {
	return internal.ConvertString(pattern, producer)
}

func ConvertUTF16String[TOutput any](pattern Parser[uint16], producer func(string) TOutput) ResultParser[uint16, TOutput] {
	return internal.ConvertUTF16String(pattern, producer)
}

func ConvertSingle[TInput any, TOutput any](pattern Parser[TInput], producer func(TInput) TOutput) ResultParser[TInput, TOutput] {
	return internal.ConvertSingle(pattern, producer)
}

func ConvertValue[TInput any, TOutput any](pattern Parser[TInput], value TOutput) ResultParser[TInput, TOutput] {
	return internal.ConvertValue(pattern, value)
}

func Count[TInput any](scanner Parser[TInput]) ResultParser[TInput, int] {
	return internal.Count(scanner)
}

func Fails[TInput any](parser Parser[TInput]) Parser[TInput] { return internal.Fails(parser) }

func First[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return internal.First(parsers...)
}

func FirstValue[TInput any, TOutput any](parsers ...ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.FirstValue(parsers...)
}

func FirstRight[TInput any, TOutput any](parsers ...RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return internal.FirstRight(parsers...)
}

func Forward[TInput any, TOutput any](deferred func() ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.Forward(deferred)
}

func ForwardValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.ForwardValue(parser)
}

func If[TInput any](test Parser[TInput], parser Parser[TInput]) Parser[TInput] {
	return internal.If(test, parser)
}

func IfValue[TInput any, TOutput any](test Parser[TInput], parser ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.IfValue(test, parser)
}

func IfRight[TInput any, TOutput any](test Parser[TInput], parser RightParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return internal.IfRight(test, parser)
}

func Limit[TInput any, TOutput any](limiter Parser[TInput], limited ResultParser[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.Limit(limiter, limited)
}

func Match[TInput any](predicate func(TInput) bool) Parser[TInput] {
	return internal.Match(predicate)
}

func MatchConsumer[TInput any](consumer SourceConsumer[TInput]) Parser[TInput] {
	return internal.MatchConsumer(consumer)
}

func MatchItem[TInput comparable](item TInput) Parser[TInput] {
	return internal.MatchItem(item)
}

func MatchItemWithComparer[TInput any](item TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	return internal.MatchItemWithComparer(item, comparer)
}

func MatchAny[TInput comparable](items ...TInput) Parser[TInput] {
	return internal.MatchAny(items...)
}

func MatchAnyWithComparer[TInput any](items []TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	return internal.MatchAnyWithComparer(items, comparer)
}

func MatchSequence[TInput comparable](items []TInput) Parser[TInput] {
	return internal.MatchSequence(items)
}

func MatchSequenceWithComparer[TInput any](items []TInput, comparer func(TInput, TInput) bool) Parser[TInput] {
	return internal.MatchSequenceWithComparer(items, comparer)
}

func MatchByte(ch byte, ignoreCase ...bool) Parser[byte] {
	return internal.MatchByte(ch, ignoreCase...)
}

func MatchText(text string, ignoreCase ...bool) Parser[byte] {
	return internal.MatchText(text, ignoreCase...)
}

func MatchValue[TInput any, TOutput any](predicate func(TInput) bool, producer func(TInput) TOutput) ResultParser[TInput, TOutput] {
	return internal.MatchValue(predicate, producer)
}

func MatchConsumerValue[TInput any, TOutput any](consumer SourceConsumer[TInput], producer SourceProducer[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.MatchConsumerValue(consumer, producer)
}

func MapValues[TInput comparable, TOutput any](entries []MapValueEntry[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.MapValues(entries)
}

func MapFuncs[TInput comparable, TOutput any](entries []MapEntry[TInput, TOutput]) ResultParser[TInput, TOutput] {
	return internal.MapFuncs(entries)
}

func Not[TInput any](parser Parser[TInput]) Parser[TInput] { return internal.Not(parser) }

func OneOrMore[TInput any](parser Parser[TInput]) Parser[TInput] {
	return internal.OneOrMore(parser)
}

func OneOrMoreProduce[TInput any, TElement any, TOutput any](parser ResultParser[TInput, TElement], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return internal.OneOrMoreProduce(parser, producer)
}

func OneOrMoreList[TInput any, TElement any](parser ResultParser[TInput, TElement]) ResultParser[TInput, []TElement] {
	return internal.OneOrMoreProduce(parser, func(elements []TElement) []TElement {
		return internal.ListToReadOnly(elements)
	})
}

func OptionalValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer ...func() TOutput) ResultParser[TInput, TOutput] {
	return internal.OptionalValue(parser, producer...)
}

func Or[TInput any](parsers ...Parser[TInput]) Parser[TInput] {
	return internal.Or(parsers...)
}

func Produce[TInput any, TOutput any](parser Parser[TInput], producer func([]any) TOutput) ResultParser[TInput, TOutput] {
	return internal.Produce(parser, producer)
}

func ProduceElements[TInput any, TElement any, TOutput any](parser Parser[TInput], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return internal.ProduceElements(parser, producer)
}

func ListElements[TInput any, TElement any, TOutput any](element ResultParser[TInput, TElement], oneOrMore bool, producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return internal.ListElements(element, oneOrMore, producer)
}

func ListElementsWithMissing[TInput any, TElement any, TOutput any](element ResultParser[TInput, TElement], missingElement func(Source[TInput], int) TElement, oneOrMore bool, producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return internal.ListElementsWithMissing(element, missingElement, oneOrMore, producer)
}

func OList[TInput any, TElement any, TSeparator any, TOutput any](element ResultParser[TInput, TElement], separator ResultParser[TInput, TSeparator], missingElement func(Source[TInput], int) TElement, missingSeparator func(Source[TInput], int) TSeparator, endOfList Parser[TInput], oneOrMore bool, allowTrailingSeparator bool, producer func([]any) TOutput) ResultParser[TInput, TOutput] {
	return internal.OList(element, separator, missingElement, missingSeparator, endOfList, oneOrMore, allowTrailingSeparator, producer)
}

func OListFull[TInput any, TElement any, TSeparator any, TOutput any](primaryElement ResultParser[TInput, TElement], separator ResultParser[TInput, TSeparator], secondaryElement ResultParser[TInput, TElement], missingPrimaryElement func(Source[TInput], int) TElement, missingSeparator func(Source[TInput], int) TSeparator, missingSecondaryElement func(Source[TInput], int) TElement, endOfList Parser[TInput], oneOrMore bool, allowTrailingSeparator bool, producer func([]any) TOutput) ResultParser[TInput, TOutput] {
	return internal.OListFull(primaryElement, separator, secondaryElement, missingPrimaryElement, missingSeparator, missingSecondaryElement, endOfList, oneOrMore, allowTrailingSeparator, producer)
}

func ListSeparated[TInput any, TElement any, TSeparator any, TOutput any](element ResultParser[TInput, TElement], separator ResultParser[TInput, TSeparator], oneOrMore bool, producer func([]ElementAndSeparator[TElement, TSeparator]) TOutput) ResultParser[TInput, TOutput] {
	return internal.ListSeparated(element, separator, oneOrMore, producer)
}

func ListSeparatedFull[TInput any, TElement any, TSeparator any, TOutput any](element ResultParser[TInput, TElement], separator ResultParser[TInput, TSeparator], secondaryElement ResultParser[TInput, TElement], missingElement func(Source[TInput], int) TElement, missingSeparator func(Source[TInput], int) TSeparator, missingSecondaryElement func(Source[TInput], int) TElement, endOfList Parser[TInput], oneOrMore bool, allowTrailingSeparator bool, producer func([]ElementAndSeparator[TElement, TSeparator]) TOutput) ResultParser[TInput, TOutput] {
	return internal.ListSeparatedFull(element, separator, secondaryElement, missingElement, missingSeparator, missingSecondaryElement, endOfList, oneOrMore, allowTrailingSeparator, producer)
}

func Required[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func(Source[TInput], int) TOutput) ResultParser[TInput, TOutput] {
	return internal.Required(parser, producer)
}

func RequiredValue[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func() TOutput) ResultParser[TInput, TOutput] {
	return internal.RequiredValue(parser, producer)
}

func Rule1[TInput any, T1 any, TOutput any](p1 ResultParser[TInput, T1], producer func(T1) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule1(p1, producer)
}

func Rule2[TInput any, T1 any, T2 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], producer func(T1, T2) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule2(p1, p2, producer)
}

func Rule3[TInput any, T1 any, T2 any, T3 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], producer func(T1, T2, T3) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule3(p1, p2, p3, producer)
}

func Rule4[TInput any, T1 any, T2 any, T3 any, T4 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], producer func(T1, T2, T3, T4) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule4(p1, p2, p3, p4, producer)
}

func Rule5[TInput any, T1 any, T2 any, T3 any, T4 any, T5 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], p5 ResultParser[TInput, T5], producer func(T1, T2, T3, T4, T5) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule5(p1, p2, p3, p4, p5, producer)
}

func Rule6[TInput any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], p5 ResultParser[TInput, T5], p6 ResultParser[TInput, T6], producer func(T1, T2, T3, T4, T5, T6) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule6(p1, p2, p3, p4, p5, p6, producer)
}

func Rule7[TInput any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], p5 ResultParser[TInput, T5], p6 ResultParser[TInput, T6], p7 ResultParser[TInput, T7], producer func(T1, T2, T3, T4, T5, T6, T7) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule7(p1, p2, p3, p4, p5, p6, p7, producer)
}

func Rule8[TInput any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], p5 ResultParser[TInput, T5], p6 ResultParser[TInput, T6], p7 ResultParser[TInput, T7], p8 ResultParser[TInput, T8], producer func(T1, T2, T3, T4, T5, T6, T7, T8) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule8(p1, p2, p3, p4, p5, p6, p7, p8, producer)
}

func Rule9[TInput any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], p4 ResultParser[TInput, T4], p5 ResultParser[TInput, T5], p6 ResultParser[TInput, T6], p7 ResultParser[TInput, T7], p8 ResultParser[TInput, T8], p9 ResultParser[TInput, T9], producer func(T1, T2, T3, T4, T5, T6, T7, T8, T9) TOutput) ResultParser[TInput, TOutput] {
	return internal.Rule9(p1, p2, p3, p4, p5, p6, p7, p8, p9, producer)
}

func RuleRight0[TInput any, TLeft any, TOutput any](producer func(TLeft) TOutput) RightParser[TInput, TOutput] {
	return internal.RuleRight0[TInput](producer)
}

func RuleRight1[TInput any, TLeft any, T1 any, TOutput any](p1 ResultParser[TInput, T1], producer func(TLeft, T1) TOutput) RightParser[TInput, TOutput] {
	return internal.RuleRight1[TInput](p1, producer)
}

func RuleRight2[TInput any, TLeft any, T1 any, T2 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], producer func(TLeft, T1, T2) TOutput) RightParser[TInput, TOutput] {
	return internal.RuleRight2[TInput](p1, p2, producer)
}

func RuleRight3[TInput any, TLeft any, T1 any, T2 any, T3 any, TOutput any](p1 ResultParser[TInput, T1], p2 ResultParser[TInput, T2], p3 ResultParser[TInput, T3], producer func(TLeft, T1, T2, T3) TOutput) RightParser[TInput, TOutput] {
	return internal.RuleRight3[TInput](p1, p2, p3, producer)
}

func Text(pattern Parser[byte]) ResultParser[byte, string] { return internal.Text(pattern) }

func TextLiteral(value string, ignoreCase ...bool) ResultParser[byte, string] {
	return internal.TextLiteral(value, ignoreCase...)
}

func TextAndOffset(pattern Parser[byte]) ResultParser[byte, OffsetValue[string]] {
	return internal.TextAndOffset(pattern)
}

func Value[TInput any, TOutput any](producer func() TOutput) ResultParser[TInput, TOutput] {
	return internal.Value[TInput](producer)
}

func ZeroOrMore[TInput any](parser Parser[TInput]) Parser[TInput] {
	return internal.ZeroOrMore(parser)
}

func ZeroOrMoreProduce[TInput any, TElement any, TOutput any](parser ResultParser[TInput, TElement], producer func([]TElement) TOutput) ResultParser[TInput, TOutput] {
	return internal.ZeroOrMoreProduce(parser, producer)
}

func ZeroOrMoreList[TInput any, TElement any](parser ResultParser[TInput, TElement]) ResultParser[TInput, []TElement] {
	return internal.ZeroOrMoreProduce(parser, func(elements []TElement) []TElement {
		return internal.ListToReadOnly(elements)
	})
}

func ZeroOrOne[TInput any](parser Parser[TInput]) Parser[TInput] {
	return internal.ZeroOrOne(parser)
}

func ParseText(parser Parser[byte], text string, output *[]any) int {
	return internal.ParseText(parser, text, output)
}

func ParseTextValue[TOutput any](parser ResultParser[byte, TOutput], text string) ParseResult[TOutput] {
	return internal.ParseTextValue(parser, text)
}

func TryParseTextValue[TOutput any](parser ResultParser[byte, TOutput], text string) (TOutput, bool) {
	return internal.TryParseTextValue(parser, text)
}

func ParseSafe[TInput any](parser Parser[TInput], source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return internal.ParseSafe(parser, source, inputStart, output, outputStart)
}

func ScanSafe[TInput any](parser Parser[TInput], source Source[TInput], inputStart int) int {
	return internal.ScanSafe(parser, source, inputStart)
}

func ScanPartial[TInput any](parser Parser[TInput], source Source[TInput], inputStart int) int {
	return internal.ScanPartial(parser, source, inputStart)
}

func ParsePartial[TInput any](parser Parser[TInput], source Source[TInput], inputStart int, output *[]any, outputStart int, onFailure ...PartialParserFailureHandler[TInput]) int {
	return internal.ParsePartial(parser, source, inputStart, output, outputStart, onFailure...)
}

func ScanPartialBest[TInput any](parsers []Parser[TInput], source Source[TInput], inputStart int) int {
	return internal.ScanPartialBest(parsers, source, inputStart)
}

func ParsePartialBest[TInput any](parsers []Parser[TInput], source Source[TInput], inputStart int, output *[]any, outputStart int, bestParsers *[]Parser[TInput], onFailure ...PartialParserFailureHandler[TInput]) int {
	return internal.ParsePartialBest(parsers, source, inputStart, output, outputStart, bestParsers, onFailure...)
}

func AcceptParserVisitor[TInput any](parser Parser[TInput], visitor ParserVisitorFuncs[TInput]) {
	internal.AcceptParserVisitor(parser, visitor)
}

func AcceptParserVisitorResult[TInput any, TResult any](parser Parser[TInput], visitor ParserVisitorResultFuncs[TInput, TResult]) TResult {
	return internal.AcceptParserVisitorResult(parser, visitor)
}

// AcceptParserVisitorArgResult is the Go spelling of
// Parser.Accept<TArg, TResult>(ParserVisitor<TInput, TArg, TResult>, TArg).
func AcceptParserVisitorArgResult[TInput any, TArg any, TResult any](parser Parser[TInput], visitor ParserVisitorArgResultFuncs[TInput, TArg, TResult], arg TArg) TResult {
	return internal.AcceptParserVisitorArgResult(parser, visitor, arg)
}

func IsParserParent[TInput any](parent Parser[TInput], child Parser[TInput]) bool {
	return internal.IsParserParent(parent, child)
}

func DescribeParser[TInput any](parser Parser[TInput], showRequired ...bool) string {
	return internal.DescribeParser(parser, showRequired...)
}
