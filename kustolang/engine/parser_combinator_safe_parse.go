package kustolang

// Ported from src/Kusto.Language/Parser/Combinators/SafeParse.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ParseSafe[TInput any](parser Parser[TInput], source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	safeParser := newStackSafeParser(source, output)
	return safeParser.parse(parser, inputStart, outputStart)
}

type stackSafeParser[TInput any] struct {
	source Source[TInput]
	output *[]any
	stack  []safeParseState[TInput]
}

type safeParseState[TInput any] struct {
	parser              Parser[TInput]
	inputStart          int
	outputStart         int
	originalOutputCount int
	inputLength         int
	nextOutputStart     int
	state               int
	lastResult          int
	bestFailedResult    int
	previousSource      Source[TInput]
}

func newStackSafeParser[TInput any](source Source[TInput], output *[]any) *stackSafeParser[TInput] {
	return &stackSafeParser[TInput]{source: source, output: output}
}

func (p *stackSafeParser[TInput]) parse(parser Parser[TInput], inputStart int, outputStart int) int {
	if parser == nil {
		return -1
	}
	p.stack = p.stack[:0]
	p.push(parser, inputStart, outputStart)

	for {
		current := &p.stack[len(p.stack)-1]
		next := p.visit(current)
		if next != nil {
			p.push(next, current.inputStart+current.inputLength, current.nextOutputStart)
			continue
		}

		result := current.inputLength
		if len(p.stack) == 1 {
			return result
		}

		p.stack = p.stack[:len(p.stack)-1]
		parent := &p.stack[len(p.stack)-1]
		parent.lastResult = result
	}
}

func (p *stackSafeParser[TInput]) push(parser Parser[TInput], inputStart int, outputStart int) {
	p.stack = append(p.stack, safeParseState[TInput]{
		parser:              parser,
		inputStart:          inputStart,
		outputStart:         outputStart,
		originalOutputCount: outputCount(p.output),
		nextOutputStart:     outputCount(p.output),
	})
}

func (p *stackSafeParser[TInput]) visit(state *safeParseState[TInput]) Parser[TInput] {
	if apply, ok := state.parser.(safeScanApplyParser[TInput]); ok {
		left, right, kind := apply.safeScanApplyParts()
		return p.visitApply(state, left, right, kind)
	}
	if best, ok := state.parser.(safeScanBestParser[TInput]); ok {
		return p.visitBest(state, best.safeScanBestParsers())
	}
	if first, ok := state.parser.(safeScanFirstParser[TInput]); ok {
		return p.visitFirst(state, first.safeScanFirstParsers())
	}
	if forward, ok := state.parser.(safeScanForwardParser[TInput]); ok {
		return p.visitForward(state, forward.safeScanForwardParser())
	}
	if conditional, ok := state.parser.(safeScanIfParser[TInput]); ok {
		test, parser := conditional.safeScanIfParts()
		return p.visitIf(state, test, parser)
	}
	if limit, ok := state.parser.(safeScanLimitParser[TInput]); ok {
		limiter, limited := limit.safeScanLimitParts()
		return p.visitLimit(state, limiter, limited)
	}
	if oneOrMore, ok := state.parser.(*OneOrMoreParser[TInput]); ok {
		return p.visitOneOrMore(state, oneOrMore.Parser)
	}
	if optional, ok := state.parser.(safeParseOptionalParser[TInput]); ok {
		return p.visitOptional(state, optional.safeParseInnerParser(), optional.safeParseOptionalProducer())
	}
	if produce, ok := state.parser.(safeParseProduceParser[TInput]); ok {
		return p.visitProduce(state, produce.safeParseInnerParser(), produce.safeParseOutputProducer())
	}
	if required, ok := state.parser.(safeParseRequiredParser[TInput]); ok {
		return p.visitRequired(state, required.safeParseInnerParser(), required.safeParseRequiredProducer())
	}
	if sequence, ok := state.parser.(safeScanSequenceParser[TInput]); ok {
		return p.visitSequence(state, sequence.safeScanSequenceParsers())
	}
	if repetition, ok := state.parser.(safeScanRepetitionParser[TInput]); ok {
		parser, zeroOrOne := repetition.safeScanRepetitionParts()
		return p.visitZeroOrMore(state, parser, zeroOrOne)
	}

	state.inputLength = state.parser.ParseList(p.source, state.inputStart, p.output, outputCount(p.output))
	return nil
}

type safeParseOptionalParser[TInput any] interface {
	safeParseInnerParser() Parser[TInput]
	safeParseOptionalProducer() func() any
}

type safeParseProduceParser[TInput any] interface {
	safeParseInnerParser() Parser[TInput]
	safeParseOutputProducer() func([]any, int) any
}

type safeParseRequiredParser[TInput any] interface {
	safeParseInnerParser() Parser[TInput]
	safeParseRequiredProducer() func(Source[TInput], int) any
}

func (p *stackSafeParser[TInput]) visitApply(state *safeParseState[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = state.outputStart
		return left
	}
	if state.state == 1 {
		if state.lastResult < 0 {
			if kind == ApplyKindOne {
				state.inputLength = -state.inputLength + state.lastResult
			} else {
				state.inputLength = state.lastResult
			}
			return nil
		}
		state.inputLength = state.lastResult
		state.state = 2
		state.nextOutputStart = state.outputStart
		return right
	}
	if state.lastResult > 0 {
		state.inputLength += state.lastResult
		if kind == ApplyKindZeroOrMore {
			return right
		}
	}
	return nil
}

func (p *stackSafeParser[TInput]) visitBest(state *safeParseState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		bestParser := p.bestParserByScan(parsers, state.inputStart)
		if bestParser >= 0 {
			state.state = 1
			state.nextOutputStart = state.outputStart
			return parsers[bestParser]
		}
		state.inputLength = -1
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) bestParserByScan(parsers []Parser[TInput], inputStart int) int {
	maxLength := -1
	bestParser := -1
	for i, parser := range parsers {
		length := ScanSafe(parser, p.source, inputStart)
		if length > maxLength {
			maxLength = length
			bestParser = i
		}
	}
	return bestParser
}

func (p *stackSafeParser[TInput]) visitFirst(state *safeParseState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if len(parsers) == 0 {
		state.inputLength = -1
		return nil
	}
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = state.outputStart
		return parsers[0]
	}
	if state.lastResult < 0 {
		if state.lastResult < state.bestFailedResult {
			state.bestFailedResult = state.lastResult
		}
		setOutputCount(p.output, state.originalOutputCount)
		if state.state < len(parsers) {
			next := parsers[state.state]
			state.state++
			state.nextOutputStart = state.outputStart
			return next
		}
		state.inputLength = state.bestFailedResult
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitForward(state *safeParseState[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = state.outputStart
		return parser
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitIf(state *safeParseState[TInput], test Parser[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		length := ScanSafe(test, p.source, state.inputStart)
		if length < 0 {
			state.inputLength = length
			return nil
		}
		state.state = 1
		state.nextOutputStart = state.outputStart
		return parser
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitLimit(state *safeParseState[TInput], limiter Parser[TInput], limited Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		length := limiter.Scan(p.source, state.inputStart)
		if length > 0 {
			state.previousSource = p.source
			p.source = NewLimitSource(p.source, state.inputStart+length)
			state.state = 1
			state.nextOutputStart = outputCount(p.output)
			return limited
		}
		state.inputLength = -1
		return nil
	}
	p.source = state.previousSource
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitOneOrMore(state *safeParseState[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = outputCount(p.output)
		return parser
	}
	if state.state == 1 && state.lastResult < 0 {
		state.inputLength = state.lastResult
		return nil
	}
	if state.lastResult > 0 {
		state.state++
		state.inputLength += state.lastResult
		state.nextOutputStart = outputCount(p.output)
		return parser
	}
	return nil
}

func (p *stackSafeParser[TInput]) visitOptional(state *safeParseState[TInput], parser Parser[TInput], producer func() any) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = outputCount(p.output)
		return parser
	}
	if state.lastResult < 0 {
		setOutputCount(p.output, state.originalOutputCount)
		appendOutput(p.output, producer())
		state.inputLength = 0
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitProduce(state *safeParseState[TInput], parser Parser[TInput], producer func([]any, int) any) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = state.outputStart
		return parser
	}
	if state.lastResult >= 0 {
		value := producer(outputValues(p.output), state.outputStart)
		setOutputCount(p.output, state.outputStart)
		appendOutput(p.output, value)
	} else {
		setOutputCount(p.output, state.originalOutputCount)
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitRequired(state *safeParseState[TInput], parser Parser[TInput], producer func(Source[TInput], int) any) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		state.nextOutputStart = outputCount(p.output)
		return parser
	}
	if state.lastResult < 0 || outputCount(p.output) == state.originalOutputCount {
		setOutputCount(p.output, state.originalOutputCount)
		appendOutput(p.output, producer(p.source, state.inputStart))
		state.inputLength = 0
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *stackSafeParser[TInput]) visitSequence(state *safeParseState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if state.lastResult < 0 {
		state.inputLength = -state.inputLength + state.lastResult
		setOutputCount(p.output, state.originalOutputCount)
		return nil
	}
	state.inputLength += state.lastResult
	if state.state >= len(parsers) {
		if producer, ok := state.parser.(safeParseRuleProducer); ok {
			value := producer.safeParseProduceRule(outputValues(p.output), state.outputStart)
			setOutputCount(p.output, state.outputStart)
			appendOutput(p.output, value)
		}
		return nil
	}
	next := parsers[state.state]
	state.state++
	state.nextOutputStart = outputCount(p.output)
	return next
}

func (p *stackSafeParser[TInput]) visitZeroOrMore(state *safeParseState[TInput], parser Parser[TInput], zeroOrOne bool) Parser[TInput] {
	if state.state == 0 || (state.lastResult > 0 && !zeroOrOne) {
		state.state++
		state.inputLength += state.lastResult
		state.nextOutputStart = outputCount(p.output)
		return parser
	}
	return nil
}

type safeParseRuleProducer interface {
	safeParseProduceRule(output []any, outputStart int) any
}

func (p *OptionalParser[TInput, TOutput]) safeParseInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *OptionalParser[TInput, TOutput]) safeParseOptionalProducer() func() any {
	return func() any { return p.Producer() }
}

func (p *ProduceParser[TInput, TProducer]) safeParseInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *ProduceParser[TInput, TProducer]) safeParseOutputProducer() func([]any, int) any {
	return func(output []any, outputStart int) any {
		return p.Producer(output, outputStart)
	}
}

func (p *RequiredParser[TInput, TOutput]) safeParseInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *RequiredParser[TInput, TOutput]) safeParseRequiredProducer() func(Source[TInput], int) any {
	return func(source Source[TInput], start int) any {
		return p.Producer(source, start)
	}
}

func (p *RuleParser[TInput, TProducer]) safeParseProduceRule(output []any, outputStart int) any {
	return p.ListProducer(output, outputStart)
}
