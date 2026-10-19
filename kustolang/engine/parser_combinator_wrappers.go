package kustolang

// Ported from:
//   src/Kusto.Language/Parser/Combinators/Parsers/FailsParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/NotParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ProduceParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/RequiredParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/OptionalParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/LimitParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ZeroOrMoreParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/OneOrMoreParser.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type FailsParser[TInput any] struct {
	ParserInfo[TInput]
	Pattern Parser[TInput]
}

func NewFailsParser[TInput any](pattern Parser[TInput]) *FailsParser[TInput] {
	EnsureArgumentNotNil(pattern, "pattern")
	return &FailsParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Negation: true}),
			WithParserChildren[TInput](pattern),
		),
		Pattern: pattern,
	}
}

func (p *FailsParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *FailsParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *FailsParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *FailsParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *FailsParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *FailsParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return p.Scan(source, inputStart)
}

func (p *FailsParser[TInput]) Scan(source Source[TInput], start int) int {
	if source.IsEnd(start) {
		return 0
	}
	if p.Pattern.Scan(source, start) >= 0 {
		return -1
	}
	return 0
}

func (p *FailsParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type NotParser[TInput any] struct {
	ParserInfo[TInput]
	Pattern Parser[TInput]
}

func NewNotParser[TInput any](pattern Parser[TInput]) *NotParser[TInput] {
	EnsureArgumentNotNil(pattern, "pattern")
	return &NotParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Negation: true}),
			WithParserChildren[TInput](pattern),
		),
		Pattern: pattern,
	}
}

func (p *NotParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *NotParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *NotParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *NotParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *NotParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *NotParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return p.Scan(source, inputStart)
}

func (p *NotParser[TInput]) Scan(source Source[TInput], start int) int {
	if source.IsEnd(start) {
		return -1
	}
	if p.Pattern.Scan(source, start) >= 0 {
		return -1
	}
	return 1
}

func (p *NotParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type ProduceParser[TInput any, TProducer any] struct {
	ParserInfo[TInput]
	Parser   Parser[TInput]
	Producer func([]any, int) TProducer
}

func NewProduceParser[TInput any, TProducer any](parser Parser[TInput], producer func([]any, int) TProducer) *ProduceParser[TInput, TProducer] {
	EnsureArgumentNotNil(parser, "parser")
	EnsureArgumentNotNil(producer, "producer")
	return &ProduceParser[TInput, TProducer]{
		ParserInfo: NewParserInfo(WithParserChildren[TInput](parser)),
		Parser:     parser,
		Producer:   producer,
	}
}

func (p *ProduceParser[TInput, TProducer]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *ProduceParser[TInput, TProducer]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *ProduceParser[TInput, TProducer]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *ProduceParser[TInput, TProducer]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *ProduceParser[TInput, TProducer]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ProduceParser[TInput, TProducer]) ParseValue(source Source[TInput], inputStart int) ParseResult[TProducer] {
	list := listPrimaryOutputListPool.AllocateFromPool()
	defer listPrimaryOutputListPool.ReturnToPool(list)

	n := p.ParseList(source, inputStart, list, 0)
	var value TProducer
	if n >= 0 && len(*list) > 0 {
		value = (*list)[0].(TProducer)
	}
	return NewParseResult(n, value)
}

func (p *ProduceParser[TInput, TProducer]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	originalOutputCount := outputCount(output)
	length := p.Parser.ParseList(source, inputStart, output, outputStart)
	if length >= 0 {
		value := p.Producer(outputValues(output), outputStart)
		setOutputCount(output, outputStart)
		appendOutput(output, value)
	} else {
		setOutputCount(output, originalOutputCount)
	}
	return length
}

func (p *ProduceParser[TInput, TProducer]) Scan(source Source[TInput], start int) int {
	return p.Parser.Scan(source, start)
}

func (p *ProduceParser[TInput, TProducer]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type RequiredParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Parser   ResultParser[TInput, TOutput]
	Producer func(Source[TInput], int) TOutput
}

func NewRequiredParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func(Source[TInput], int) TOutput) *RequiredParser[TInput, TOutput] {
	EnsureArgumentNotNil(parser, "parser")
	EnsureArgumentNotNil(producer, "producer")
	return &RequiredParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Required: true}),
			WithParserChildren[TInput](parser),
		),
		Parser:   parser,
		Producer: producer,
	}
}

func (p *RequiredParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *RequiredParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *RequiredParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *RequiredParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *RequiredParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *RequiredParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	result := p.Parser.ParseValue(source, start)
	if result.Length < 0 {
		return NewParseResult(0, p.Producer(source, start))
	}
	return result
}

func (p *RequiredParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	originalOutputCount := outputCount(output)
	length := p.Parser.ParseList(source, inputStart, output, originalOutputCount)
	if length < 0 || outputCount(output) == originalOutputCount {
		setOutputCount(output, originalOutputCount)
		appendOutput(output, p.Producer(source, inputStart))
		return 0
	}
	return length
}

func (p *RequiredParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	n := p.Parser.Scan(source, start)
	if n < 0 {
		return 0
	}
	return n
}

func (p *RequiredParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type OptionalParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Parser   ResultParser[TInput, TOutput]
	Producer func() TOutput
}

func NewOptionalParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput], producer func() TOutput) *OptionalParser[TInput, TOutput] {
	EnsureArgumentNotNil(parser, "parser")
	EnsureArgumentNotNil(producer, "producer")
	return &OptionalParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Optional: true}),
			WithParserChildren[TInput](parser),
		),
		Parser:   parser,
		Producer: producer,
	}
}

func (p *OptionalParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *OptionalParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *OptionalParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *OptionalParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *OptionalParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *OptionalParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	result := p.Parser.ParseValue(source, start)
	if result.Length < 0 {
		return NewParseResult(0, p.Producer())
	}
	return result
}

func (p *OptionalParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	originalOutputCount := outputCount(output)
	length := p.Parser.ParseList(source, inputStart, output, originalOutputCount)
	if length < 0 || outputCount(output) == originalOutputCount {
		setOutputCount(output, originalOutputCount)
		appendOutput(output, p.Producer())
		return 0
	}
	return length
}

func (p *OptionalParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	n := p.Parser.Scan(source, start)
	if n < 0 {
		return 0
	}
	return n
}

func (p *OptionalParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type LimitParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Limiter Parser[TInput]
	Limited ResultParser[TInput, TOutput]
}

func NewLimitParser[TInput any, TOutput any](limiter Parser[TInput], limited ResultParser[TInput, TOutput]) *LimitParser[TInput, TOutput] {
	EnsureArgumentNotNil(limiter, "limiter")
	EnsureArgumentNotNil(limited, "limited")
	return &LimitParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Conditional: true}),
			WithParserChildren[TInput](limiter, limited),
		),
		Limiter: limiter,
		Limited: limited,
	}
}

func (p *LimitParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *LimitParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *LimitParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *LimitParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *LimitParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *LimitParser[TInput, TOutput]) ParseValue(source Source[TInput], inputStart int) ParseResult[TOutput] {
	length := p.Limiter.Scan(source, inputStart)
	if length >= 0 {
		return p.Limited.ParseValue(NewLimitSource(source, inputStart+length), inputStart)
	}
	var zero TOutput
	return NewParseResult(-1, zero)
}

func (p *LimitParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := p.Limiter.Scan(source, inputStart)
	if length >= 0 {
		return p.Limited.ParseList(NewLimitSource(source, inputStart+length), inputStart, output, outputStart)
	}
	return -1
}

func (p *LimitParser[TInput, TOutput]) Scan(source Source[TInput], inputStart int) int {
	length := p.Limiter.Scan(source, inputStart)
	if length >= 0 {
		return p.Limited.Scan(NewLimitSource(source, inputStart+length), inputStart)
	}
	return -1
}

func (p *LimitParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type ZeroOrMoreParser[TInput any] struct {
	ParserInfo[TInput]
	Parser    Parser[TInput]
	ZeroOrOne bool
}

func NewZeroOrMoreParser[TInput any](parser Parser[TInput], zeroOrOne ...bool) *ZeroOrMoreParser[TInput] {
	EnsureArgumentNotNil(parser, "parser")
	limitOne := len(zeroOrOne) > 0 && zeroOrOne[0]
	return &ZeroOrMoreParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Optional: true, Repetition: true}),
			WithParserChildren[TInput](parser),
		),
		Parser:    parser,
		ZeroOrOne: limitOne,
	}
}

func (p *ZeroOrMoreParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *ZeroOrMoreParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *ZeroOrMoreParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *ZeroOrMoreParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *ZeroOrMoreParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ZeroOrMoreParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := 0
	for {
		n := p.Parser.ParseList(source, inputStart+length, output, outputCount(output))
		if n > 0 {
			length += n
		}
		if n <= 0 || p.ZeroOrOne {
			return length
		}
	}
}

func (p *ZeroOrMoreParser[TInput]) Scan(source Source[TInput], start int) int {
	length := 0
	for {
		n := p.Parser.Scan(source, start+length)
		if n > 0 {
			length += n
		}
		if n <= 0 || p.ZeroOrOne {
			return length
		}
	}
}

func (p *ZeroOrMoreParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type OneOrMoreParser[TInput any] struct {
	ParserInfo[TInput]
	Parser Parser[TInput]
}

func NewOneOrMoreParser[TInput any](parser Parser[TInput]) *OneOrMoreParser[TInput] {
	EnsureArgumentNotNil(parser, "parser")
	return &OneOrMoreParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Repetition: true}),
			WithParserChildren[TInput](parser),
		),
		Parser: parser,
	}
}

func (p *OneOrMoreParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *OneOrMoreParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *OneOrMoreParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *OneOrMoreParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *OneOrMoreParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *OneOrMoreParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	firstLength := p.Parser.ParseList(source, inputStart, output, outputCount(output))
	if firstLength < 0 {
		return firstLength
	}

	length := firstLength
	for {
		n := p.Parser.ParseList(source, inputStart+length, output, outputCount(output))
		if n <= 0 {
			return length
		}
		length += n
	}
}

func (p *OneOrMoreParser[TInput]) Scan(source Source[TInput], start int) int {
	n := p.Parser.Scan(source, start)
	if n < 0 {
		return n
	}

	length := n
	for {
		n = p.Parser.Scan(source, start+length)
		if n <= 0 {
			return length
		}
		length += n
	}
}

func (p *OneOrMoreParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

func parserWithTag[TInput any](info ParserInfo[TInput], tag string, self Parser[TInput], clone func(ParserInfo[TInput]) Parser[TInput]) Parser[TInput] {
	updated, changed := info.WithTag(tag)
	if !changed {
		return self
	}
	return clone(updated)
}

func parserWithAnnotations[TInput any](info ParserInfo[TInput], annotations []any, self Parser[TInput], clone func(ParserInfo[TInput]) Parser[TInput]) Parser[TInput] {
	updated, changed := info.WithAnnotations(annotations)
	if !changed {
		return self
	}
	return clone(updated)
}

func parserWithIsHidden[TInput any](info ParserInfo[TInput], hidden bool, self Parser[TInput], clone func(ParserInfo[TInput]) Parser[TInput]) Parser[TInput] {
	updated, changed := info.WithIsHidden(hidden)
	if !changed {
		return self
	}
	return clone(updated)
}

func outputCount(output *[]any) int {
	if output == nil {
		return 0
	}
	return len(*output)
}

func outputValues(output *[]any) []any {
	if output == nil {
		return nil
	}
	return *output
}

func appendOutput(output *[]any, value any) {
	if output == nil {
		return
	}
	*output = append(*output, value)
}

func setOutputCount(output *[]any, count int) {
	if output == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	if count <= len(*output) {
		clear((*output)[count:])
		*output = (*output)[:count]
		return
	}
	for len(*output) < count {
		*output = append(*output, nil)
	}
}
