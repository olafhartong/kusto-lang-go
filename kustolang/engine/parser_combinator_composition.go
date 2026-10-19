package kustolang

// Ported from:
//   src/Kusto.Language/Parser/Combinators/Parsers/MatchParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ConvertParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/SequenceParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/FirstParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/BestParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/RuleParser.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type MatchParser[TInput any] struct {
	ParserInfo[TInput]
	Consumer SourceConsumer[TInput]
}

func NewMatchParser[TInput any](consumer SourceConsumer[TInput]) *MatchParser[TInput] {
	EnsureArgumentNotNil(consumer, "consumer")
	return &MatchParser[TInput]{
		ParserInfo: NewParserInfo(WithParserTraits[TInput](ParserTraits{Match: true})),
		Consumer:   consumer,
	}
}

func NewPredicateMatchParser[TInput any](predicate func(TInput) bool) *MatchParser[TInput] {
	EnsureArgumentNotNil(predicate, "predicate")
	return NewMatchParser(func(source Source[TInput], start int) int {
		if !source.IsEnd(start) && predicate(source.Peek(start)) {
			return 1
		}
		return -1
	})
}

func (p *MatchParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *MatchParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *MatchParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *MatchParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *MatchParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *MatchParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return p.Scan(source, inputStart)
}

func (p *MatchParser[TInput]) Scan(source Source[TInput], start int) int {
	return p.Consumer(source, start)
}

func (p *MatchParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type MatchValueParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Consumer SourceConsumer[TInput]
	Producer SourceProducer[TInput, TOutput]
}

func NewMatchValueParser[TInput any, TOutput any](consumer SourceConsumer[TInput], producer SourceProducer[TInput, TOutput]) *MatchValueParser[TInput, TOutput] {
	EnsureArgumentNotNil(consumer, "consumer")
	EnsureArgumentNotNil(producer, "producer")
	return &MatchValueParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(WithParserTraits[TInput](ParserTraits{Match: true})),
		Consumer:   consumer,
		Producer:   producer,
	}
}

func NewPredicateMatchValueParser[TInput any, TOutput any](predicate func(TInput) bool, producer func(TInput) TOutput) *MatchValueParser[TInput, TOutput] {
	EnsureArgumentNotNil(predicate, "predicate")
	EnsureArgumentNotNil(producer, "producer")
	return NewMatchValueParser(
		func(source Source[TInput], start int) int {
			if !source.IsEnd(start) && predicate(source.Peek(start)) {
				return 1
			}
			return -1
		},
		func(source Source[TInput], start int, length int) TOutput {
			return producer(source.Peek(start))
		},
	)
}

func (p *MatchValueParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *MatchValueParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *MatchValueParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *MatchValueParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *MatchValueParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *MatchValueParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	length := p.Consumer(source, start)
	if length >= 0 {
		return NewParseResult(length, p.Producer(source, start, length))
	}
	var zero TOutput
	return NewParseResult(length, zero)
}

func (p *MatchValueParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := p.Consumer(source, inputStart)
	if length >= 0 {
		appendOutput(output, p.Producer(source, inputStart, length))
	}
	return length
}

func (p *MatchValueParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	return p.Consumer(source, start)
}

func (p *MatchValueParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type ConvertParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Pattern        Parser[TInput]
	ListProducer   SourceProducer[TInput, TOutput]
	SingleProducer func(TInput) TOutput
}

func NewConvertParser[TInput any, TOutput any](pattern Parser[TInput], producer SourceProducer[TInput, TOutput]) *ConvertParser[TInput, TOutput] {
	return newConvertParser(pattern, producer, nil)
}

func NewConvertParserFromList[TInput any, TOutput any](pattern Parser[TInput], producer func([]TInput) TOutput) *ConvertParser[TInput, TOutput] {
	EnsureArgumentNotNil(producer, "producer")
	return newConvertParser(pattern, func(source Source[TInput], start int, length int) TOutput {
		values := make([]TInput, 0, length)
		for i := 0; i < length; i++ {
			values = append(values, source.Peek(start+i))
		}
		return producer(values)
	}, nil)
}

func NewConvertParserFromSingle[TInput any, TOutput any](pattern Parser[TInput], producer func(TInput) TOutput) *ConvertParser[TInput, TOutput] {
	return newConvertParser(pattern, nil, producer)
}

func newConvertParser[TInput any, TOutput any](pattern Parser[TInput], listProducer SourceProducer[TInput, TOutput], singleProducer func(TInput) TOutput) *ConvertParser[TInput, TOutput] {
	EnsureArgumentNotNil(pattern, "pattern")
	EnsureIsTrue(listProducer != nil || singleProducer != nil)
	return &ConvertParser[TInput, TOutput]{
		ParserInfo:     NewParserInfo(WithParserChildren[TInput](pattern)),
		Pattern:        pattern,
		ListProducer:   listProducer,
		SingleProducer: singleProducer,
	}
}

func (p *ConvertParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *ConvertParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *ConvertParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *ConvertParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *ConvertParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ConvertParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	length := p.Pattern.Scan(source, start)
	if length < 0 {
		var zero TOutput
		return NewParseResult(length, zero)
	}
	return NewParseResult(length, p.produce(source, start, length))
}

func (p *ConvertParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	result := p.ParseValue(source, inputStart)
	if result.Length >= 0 {
		appendOutput(output, result.Value)
	}
	return result.Length
}

func (p *ConvertParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	return p.Pattern.Scan(source, start)
}

func (p *ConvertParser[TInput, TOutput]) produce(source Source[TInput], start int, length int) TOutput {
	if p.SingleProducer != nil {
		return p.SingleProducer(source.Peek(start))
	}
	return p.ListProducer(source, start, length)
}

func (p *ConvertParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type SequenceParser[TInput any] struct {
	ParserInfo[TInput]
	Parsers []Parser[TInput]
}

func NewSequenceParser[TInput any](parsers []Parser[TInput]) *SequenceParser[TInput] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	cloned := cloneParserSlice(parsers)
	return &SequenceParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Sequence: true}),
			WithParserChildren[TInput](cloned...),
		),
		Parsers: cloned,
	}
}

func (p *SequenceParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *SequenceParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *SequenceParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *SequenceParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *SequenceParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *SequenceParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := 0
	originalOutputCount := outputCount(output)
	for _, parser := range p.Parsers {
		n := parser.ParseList(source, inputStart+length, output, outputCount(output))
		if n < 0 {
			setOutputCount(output, originalOutputCount)
			return -length + n
		}
		length += n
	}
	return length
}

func (p *SequenceParser[TInput]) Scan(source Source[TInput], start int) int {
	length := 0
	for _, parser := range p.Parsers {
		n := parser.Scan(source, start+length)
		if n < 0 {
			return n - length
		}
		length += n
	}
	return length
}

func (p *SequenceParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneParserSlice(p.Parsers)
	return &clone
}

type FirstParser[TInput any] struct {
	ParserInfo[TInput]
	Parsers []Parser[TInput]
}

func NewFirstParser[TInput any](parsers []Parser[TInput]) *FirstParser[TInput] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	cloned := cloneParserSlice(parsers)
	return &FirstParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Alternation: true}),
			WithParserChildren[TInput](cloned...),
		),
		Parsers: cloned,
	}
}

func (p *FirstParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *FirstParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *FirstParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *FirstParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *FirstParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *FirstParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	minLength := -1
	for _, parser := range p.Parsers {
		length := parser.ParseList(source, inputStart, output, outputStart)
		if length >= 0 {
			return length
		}
		if length < minLength {
			minLength = length
		}
	}
	return minLength
}

func (p *FirstParser[TInput]) Scan(source Source[TInput], start int) int {
	minLength := -1
	for _, parser := range p.Parsers {
		n := parser.Scan(source, start)
		if n >= 0 {
			return n
		}
		if n < minLength {
			minLength = n
		}
	}
	return minLength
}

func (p *FirstParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneParserSlice(p.Parsers)
	return &clone
}

type FirstValueParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Parsers []ResultParser[TInput, TOutput]
}

func NewFirstValueParser[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput]) *FirstValueParser[TInput, TOutput] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	cloned := cloneResultParserSlice(parsers)
	return &FirstValueParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Alternation: true}),
			WithParserChildren[TInput](resultParsersAsParsers(cloned)...),
		),
		Parsers: cloned,
	}
}

func (p *FirstValueParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *FirstValueParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *FirstValueParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *FirstValueParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *FirstValueParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *FirstValueParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	minLength := -1
	for _, parser := range p.Parsers {
		result := parser.ParseValue(source, start)
		if result.Length >= 0 {
			return result
		}
		if result.Length < minLength {
			minLength = result.Length
		}
	}
	var zero TOutput
	return NewParseResult(minLength, zero)
}

func (p *FirstValueParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	minLength := -1
	originalOutputCount := outputCount(output)
	for _, parser := range p.Parsers {
		setOutputCount(output, originalOutputCount)
		length := parser.ParseList(source, inputStart, output, outputStart)
		if length >= 0 {
			return length
		}
		if length < minLength {
			minLength = length
		}
	}
	return minLength
}

func (p *FirstValueParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	minLength := -1
	for _, parser := range p.Parsers {
		n := parser.Scan(source, start)
		if n >= 0 {
			return n
		}
		if n < minLength {
			minLength = n
		}
	}
	return minLength
}

func (p *FirstValueParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneResultParserSlice(p.Parsers)
	return &clone
}

type BestParser[TInput any] struct {
	ParserInfo[TInput]
	Parsers []Parser[TInput]
}

func NewBestParser[TInput any](parsers []Parser[TInput]) *BestParser[TInput] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	cloned := cloneParserSlice(parsers)
	return &BestParser[TInput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Alternation: true}),
			WithParserChildren[TInput](cloned...),
		),
		Parsers: cloned,
	}
}

func (p *BestParser[TInput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *BestParser[TInput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *BestParser[TInput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *BestParser[TInput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *BestParser[TInput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *BestParser[TInput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	EnsureAreEqual(outputCount(output), outputStart)
	minLength := -1
	maxLength := -1
	bestParser := -1
	for index, parser := range p.Parsers {
		length := parser.Scan(source, inputStart)
		if length > maxLength {
			maxLength = length
			bestParser = index
		} else if length < minLength {
			minLength = length
		}
	}
	if maxLength >= 0 {
		p.Parsers[bestParser].ParseList(source, inputStart, output, outputStart)
	}
	return maxLength
}

func (p *BestParser[TInput]) Scan(source Source[TInput], start int) int {
	maxLength := -1
	minLength := -1
	for _, parser := range p.Parsers {
		n := parser.Scan(source, start)
		if n > maxLength {
			maxLength = n
		} else if n < minLength {
			minLength = n
		}
	}
	if maxLength >= 0 {
		return maxLength
	}
	return minLength
}

func (p *BestParser[TInput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneParserSlice(p.Parsers)
	return &clone
}

type BestValueParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	Parsers  []ResultParser[TInput, TOutput]
	IsBetter func(TOutput, TOutput) bool
}

func NewBestValueParser[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput], isBetter ...func(TOutput, TOutput) bool) *BestValueParser[TInput, TOutput] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	cloned := cloneResultParserSlice(parsers)
	var better func(TOutput, TOutput) bool
	if len(isBetter) > 0 {
		better = isBetter[0]
	}
	return &BestValueParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Alternation: true}),
			WithParserChildren[TInput](resultParsersAsParsers(cloned)...),
		),
		Parsers:  cloned,
		IsBetter: better,
	}
}

func (p *BestValueParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *BestValueParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *BestValueParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *BestValueParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *BestValueParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *BestValueParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	minLength, maxLength, bestParser, candidates := p.findBest(source, start)
	if maxLength < 0 {
		var zero TOutput
		return NewParseResult(minLength, zero)
	}

	bestValue := p.Parsers[bestParser].ParseValue(source, start).Value
	for _, candidate := range candidates {
		otherValue := candidate.ParseValue(source, start).Value
		if p.IsBetter(otherValue, bestValue) {
			bestValue = otherValue
		}
	}
	return NewParseResult(maxLength, bestValue)
}

func (p *BestValueParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	_, maxLength, bestParser, candidates := p.findBest(source, inputStart)
	if maxLength >= 0 {
		if len(candidates) > 0 {
			p.Parsers[bestParser].ParseList(source, inputStart, output, outputStart)
			bestValue := (*output)[outputStart].(TOutput)
			setOutputCount(output, outputStart)

			for _, candidate := range candidates {
				candidate.ParseList(source, inputStart, output, outputStart)
				otherValue := (*output)[outputStart].(TOutput)
				setOutputCount(output, outputStart)
				if p.IsBetter(otherValue, bestValue) {
					bestValue = otherValue
				}
			}
			appendOutput(output, bestValue)
		} else {
			p.Parsers[bestParser].ParseList(source, inputStart, output, outputStart)
		}
	}
	return maxLength
}

func (p *BestValueParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	minLength, maxLength, _, _ := p.findBest(source, start)
	if maxLength >= 0 {
		return maxLength
	}
	return minLength
}

func (p *BestValueParser[TInput, TOutput]) findBest(source Source[TInput], start int) (minLength int, maxLength int, bestParser int, candidates []ResultParser[TInput, TOutput]) {
	minLength = -1
	maxLength = -1
	bestParser = -1
	for index, parser := range p.Parsers {
		length := parser.Scan(source, start)
		if length > maxLength {
			maxLength = length
			bestParser = index
			candidates = candidates[:0]
		} else if length == maxLength && bestParser >= 0 && p.IsBetter != nil {
			candidates = append(candidates, parser)
		} else if length < minLength {
			minLength = length
		}
	}
	return minLength, maxLength, bestParser, candidates
}

func (p *BestValueParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneResultParserSlice(p.Parsers)
	return &clone
}

type RuleParser[TInput any, TProducer any] struct {
	ParserInfo[TInput]
	Parsers        []Parser[TInput]
	ListProducer   func([]any, int) TProducer
	ResultProducer func(Source[TInput], int) ParseResult[TProducer]
}

func NewRuleParser[TInput any, TProducer any](parsers []Parser[TInput], listProducer func([]any, int) TProducer, resultProducer ...func(Source[TInput], int) ParseResult[TProducer]) *RuleParser[TInput, TProducer] {
	EnsureArgumentNotNil(parsers, "parsers")
	EnsureElementsNotNil(parsers, "parsers")
	EnsureArgumentNotNil(listProducer, "listProducer")
	cloned := cloneParserSlice(parsers)
	var result func(Source[TInput], int) ParseResult[TProducer]
	if len(resultProducer) > 0 {
		result = resultProducer[0]
	}
	return &RuleParser[TInput, TProducer]{
		ParserInfo: NewParserInfo(
			WithParserTraits[TInput](ParserTraits{Sequence: true}),
			WithParserChildren[TInput](cloned...),
		),
		Parsers:        cloned,
		ListProducer:   listProducer,
		ResultProducer: result,
	}
}

func (p *RuleParser[TInput, TProducer]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *RuleParser[TInput, TProducer]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *RuleParser[TInput, TProducer]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *RuleParser[TInput, TProducer]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *RuleParser[TInput, TProducer]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *RuleParser[TInput, TProducer]) ParseValue(source Source[TInput], start int) ParseResult[TProducer] {
	if p.ResultProducer != nil {
		return p.ResultProducer(source, start)
	}

	list := listPrimaryOutputListPool.AllocateFromPool()
	defer listPrimaryOutputListPool.ReturnToPool(list)
	n := p.ParseList(source, start, list, 0)
	var value TProducer
	if n >= 0 && len(*list) > 0 {
		value = (*list)[0].(TProducer)
	}
	return NewParseResult(n, value)
}

func (p *RuleParser[TInput, TProducer]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	length := 0
	originalOutputCount := outputCount(output)
	for _, parser := range p.Parsers {
		n := parser.ParseList(source, inputStart+length, output, outputCount(output))
		if n < 0 {
			setOutputCount(output, originalOutputCount)
			return n - length
		}
		length += n
	}

	value := p.ListProducer(outputValues(output), outputStart)
	setOutputCount(output, outputStart)
	appendOutput(output, value)
	return length
}

func (p *RuleParser[TInput, TProducer]) Scan(source Source[TInput], start int) int {
	length := 0
	for _, parser := range p.Parsers {
		n := parser.Scan(source, start+length)
		if n < 0 {
			return n - length
		}
		length += n
	}
	return length
}

func (p *RuleParser[TInput, TProducer]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	clone.Parsers = cloneParserSlice(p.Parsers)
	return &clone
}

func cloneParserSlice[TInput any](parsers []Parser[TInput]) []Parser[TInput] {
	return append([]Parser[TInput]{}, parsers...)
}

func cloneResultParserSlice[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput]) []ResultParser[TInput, TOutput] {
	return append([]ResultParser[TInput, TOutput]{}, parsers...)
}

func resultParsersAsParsers[TInput any, TOutput any](parsers []ResultParser[TInput, TOutput]) []Parser[TInput] {
	result := make([]Parser[TInput], len(parsers))
	for i, parser := range parsers {
		result[i] = parser
	}
	return result
}
