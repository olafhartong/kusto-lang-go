package kustolang

// Ported from:
//   src/Kusto.Language/Parser/Combinators/Parsers/ForwardParser.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const forwardParserMaxCallDepth = 30

type forwardParserCallState struct {
	depth     int
	safeDepth int
}

type ForwardParser[TInput any, TOutput any] struct {
	ParserInfo[TInput]
	DeferredParser func() ResultParser[TInput, TOutput]
}

func NewForwardParser[TInput any, TOutput any](deferredParser func() ResultParser[TInput, TOutput]) *ForwardParser[TInput, TOutput] {
	EnsureArgumentNotNil(deferredParser, "deferredParser")
	return &ForwardParser[TInput, TOutput]{
		ParserInfo:     NewParserInfo(WithParserTraits[TInput](ParserTraits{Forward: true})),
		DeferredParser: deferredParser,
	}
}

func (p *ForwardParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *ForwardParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *ForwardParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *ForwardParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *ForwardParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ForwardParser[TInput, TOutput]) ChildParserCount() int {
	return 1
}

func (p *ForwardParser[TInput, TOutput]) ChildParser(index int) Parser[TInput] {
	if index == 0 {
		return p.DeferredParser()
	}
	return nil
}

func (p *ForwardParser[TInput, TOutput]) ChildParserIndex(child Parser[TInput]) int {
	if p.DeferredParser() == child {
		return 0
	}
	return -1
}

func (p *ForwardParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	return withForwardParserCall(source, func(state *forwardParserCallState) ParseResult[TOutput] {
		parser := p.DeferredParser()
		if state.depth > forwardParserMaxCallDepth && state.safeDepth == 0 {
			list := listPrimaryOutputListPool.AllocateFromPool()
			defer listPrimaryOutputListPool.ReturnToPool(list)

			length := parseForwardSafe(parser, source, start, list, 0)
			var value TOutput
			if length >= 0 && len(*list) > 0 {
				value = (*list)[0].(TOutput)
			}
			return NewParseResult(length, value)
		}
		return parser.ParseValue(source, start)
	})
}

func (p *ForwardParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return withForwardParserCall(source, func(state *forwardParserCallState) int {
		parser := p.DeferredParser()
		if state.depth > forwardParserMaxCallDepth && state.safeDepth == 0 {
			return parseForwardSafe(parser, source, inputStart, output, outputStart)
		}
		return parser.ParseList(source, inputStart, output, outputStart)
	})
}

func (p *ForwardParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	return withForwardParserCall(source, func(state *forwardParserCallState) int {
		parser := p.DeferredParser()
		if state.depth > forwardParserMaxCallDepth && state.safeDepth == 0 {
			return scanForwardSafe(parser, source, start)
		}
		return parser.Scan(source, start)
	})
}

func (p *ForwardParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

func withForwardParserCall[TInput any, TResult any](source Source[TInput], fn func(*forwardParserCallState) TResult) TResult {
	state := forwardParserState(source)
	state.depth++
	defer func() {
		state.depth--
	}()
	return fn(state)
}

func parseForwardSafe[TInput any](parser Parser[TInput], source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	return withForwardParserSafeMode(source, func() int {
		return ParseSafe(parser, source, inputStart, output, outputStart)
	})
}

func scanForwardSafe[TInput any](parser Parser[TInput], source Source[TInput], start int) int {
	return withForwardParserSafeMode(source, func() int {
		return ScanSafe(parser, source, start)
	})
}

func withForwardParserSafeMode[TInput any, TResult any](source Source[TInput], fn func() TResult) TResult {
	state := forwardParserState(source)
	state.safeDepth++
	defer func() {
		state.safeDepth--
	}()
	return fn()
}

func forwardParserState[TInput any](source Source[TInput]) *forwardParserCallState {
	if source == nil {
		return &forwardParserCallState{}
	}
	return SourceCacheGetOrCreate[*forwardParserCallState](source.Cache())
}
