package kustolang

// Ported from:
//   src/Kusto.Language/Parser/Combinators/Parsers/IfParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ApplyParser.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type IfParser[TInput any] struct {
	ParserInfo[TInput]
	Test   Parser[TInput]
	Parser Parser[TInput]
}

func NewIfParser[TInput any](test Parser[TInput], parser Parser[TInput]) *IfParser[TInput] {
	EnsureArgumentNotNil(test, "test")
	EnsureArgumentNotNil(parser, "parser")
	return &IfParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Conditional: true}),
			WithParserChildren[TInput](test, parser),
		),
		Test:   test,
		Parser: parser,
	}
}

func (p *IfParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *IfParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *IfParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *IfParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *IfParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *IfParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := p.Test.Scan(source, inputStart)
	if length < 0 {
		return length
	}
	return p.Parser.ParseList(source, inputStart, output, outputStart)
}

func (p *IfParser[TInput]) Scan(source Source[TInput], start int) int {
	length := p.Test.Scan(source, start)
	if length < 0 {
		return length
	}
	return p.Parser.Scan(source, start)
}

func (p *IfParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type IfValueParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Test   Parser[TInput]
	Parser ResultParser[TInput, TOutput]
}

func NewIfValueParser[TInput any, TOutput any](test Parser[TInput], parser ResultParser[TInput, TOutput]) *IfValueParser[TInput, TOutput] {
	EnsureArgumentNotNil(test, "test")
	EnsureArgumentNotNil(parser, "parser")
	return &IfValueParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Conditional: true}),
			WithParserChildren[TInput](test, parser),
		),
		Test:   test,
		Parser: parser,
	}
}

func (p *IfValueParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *IfValueParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *IfValueParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *IfValueParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *IfValueParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *IfValueParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	length := p.Test.Scan(source, start)
	if length < 0 {
		var zero TOutput
		return NewParseResult(length, zero)
	}
	return p.Parser.ParseValue(source, start)
}

func (p *IfValueParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := p.Test.Scan(source, inputStart)
	if length < 0 {
		return length
	}
	return p.Parser.ParseList(source, inputStart, output, outputStart)
}

func (p *IfValueParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	length := p.Test.Scan(source, start)
	if length < 0 {
		return length
	}
	return p.Parser.Scan(source, start)
}

func (p *IfValueParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type ApplyKind int

const (
	ApplyKindOne ApplyKind = iota
	ApplyKindZeroOrOne
	ApplyKindZeroOrMore
)

type ApplyParser[TInput any, TLeft any, TOutput any] struct {
	ParserInfo[TInput]
	LeftParser  ResultParser[TInput, TLeft]
	RightParser ResultParser[TInput, TOutput]
	ApplyKind   ApplyKind
}

func NewApplyParser[TInput any, TLeft any, TOutput any](kind ApplyKind, leftParser ResultParser[TInput, TLeft], rightParser RightParser[TInput, TOutput]) *ApplyParser[TInput, TLeft, TOutput] {
	EnsureArgumentNotNil(leftParser, "leftParser")
	EnsureArgumentNotNil(rightParser.Parser(), "rightParser")
	return &ApplyParser[TInput, TLeft, TOutput]{
		ParserInfo:  NewParserInfo(WithParserChildren[TInput](leftParser, rightParser.Parser())),
		LeftParser:  leftParser,
		RightParser: rightParser.Parser(),
		ApplyKind:   kind,
	}
}

func (p *ApplyParser[TInput, TLeft, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) ParseValue(source Source[TInput], inputStart int) ParseResult[TOutput] {
	list := listPrimaryOutputListPool.AllocateFromPool()
	defer listPrimaryOutputListPool.ReturnToPool(list)

	n := p.ParseList(source, inputStart, list, 0)
	var value TOutput
	if n >= 0 && len(*list) > 0 {
		value = (*list)[0].(TOutput)
	}
	return NewParseResult(n, value)
}

func (p *ApplyParser[TInput, TLeft, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	leftOriginalOutputCount := outputCount(output)
	leftLength := p.LeftParser.ParseList(source, inputStart, output, outputStart)
	if leftLength < 0 {
		return leftLength
	}

	rightOriginalOutputCount := outputCount(output)
	for {
		rightLength := p.RightParser.ParseList(source, inputStart+leftLength, output, outputStart)
		if rightLength < 0 {
			if p.ApplyKind == ApplyKindOne {
				setOutputCount(output, leftOriginalOutputCount)
				return -leftLength + rightLength
			}
			setOutputCount(output, rightOriginalOutputCount)
			return leftLength
		}

		leftLength += rightLength
		if p.ApplyKind != ApplyKindZeroOrMore {
			return leftLength
		}
	}
}

func (p *ApplyParser[TInput, TLeft, TOutput]) Scan(source Source[TInput], start int) int {
	leftLength := p.LeftParser.Scan(source, start)
	if leftLength < 0 {
		return leftLength
	}

	for {
		rightLength := p.RightParser.Scan(source, start+leftLength)
		if rightLength < 0 {
			if p.ApplyKind == ApplyKindOne {
				return -leftLength + rightLength
			}
			return leftLength
		}

		leftLength += rightLength
		if p.ApplyKind != ApplyKindZeroOrMore {
			return leftLength
		}
	}
}

func (p *ApplyParser[TInput, TLeft, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}
