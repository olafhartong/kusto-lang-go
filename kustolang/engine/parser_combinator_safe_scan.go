package kustolang

// Ported from src/Kusto.Language/Parser/Combinators/SafeScan.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type safeScanApplyParser[TInput any] interface {
	safeScanApplyParts() (Parser[TInput], Parser[TInput], ApplyKind)
}

type safeScanBestParser[TInput any] interface {
	safeScanBestParsers() []Parser[TInput]
}

type safeScanFirstParser[TInput any] interface {
	safeScanFirstParsers() []Parser[TInput]
}

type safeScanForwardParser[TInput any] interface {
	safeScanForwardParser() Parser[TInput]
}

type safeScanIfParser[TInput any] interface {
	safeScanIfParts() (Parser[TInput], Parser[TInput])
}

type safeScanInnerParser[TInput any] interface {
	safeScanInnerParser() Parser[TInput]
}

type safeScanLimitParser[TInput any] interface {
	safeScanLimitParts() (Parser[TInput], Parser[TInput])
}

type safeScanPatternParser[TInput any] interface {
	safeScanPatternParser() Parser[TInput]
}

type safeScanRepetitionParser[TInput any] interface {
	safeScanRepetitionParts() (Parser[TInput], bool)
}

type safeScanSequenceParser[TInput any] interface {
	safeScanSequenceParsers() []Parser[TInput]
}

func ScanSafe[TInput any](parser Parser[TInput], source Source[TInput], start int) int {
	scanner := newStackSafeScanner(source)
	return scanner.scan(parser, start)
}

type stackSafeScanner[TInput any] struct {
	source Source[TInput]
	stack  []safeScanState[TInput]
}

type safeScanState[TInput any] struct {
	parser            Parser[TInput]
	inputStart        int
	inputLength       int
	state             int
	lastResult        int
	bestFailedResult  int
	bestSuccessResult int
	previousSource    Source[TInput]
}

func newStackSafeScanner[TInput any](source Source[TInput]) *stackSafeScanner[TInput] {
	return &stackSafeScanner[TInput]{source: source}
}

func (s *stackSafeScanner[TInput]) scan(parser Parser[TInput], start int) int {
	if parser == nil {
		return -1
	}
	s.stack = s.stack[:0]
	s.push(parser, start)

	for {
		current := &s.stack[len(s.stack)-1]
		next := s.visit(current)
		if next != nil {
			s.push(next, current.inputStart+current.inputLength)
			continue
		}

		result := current.inputLength
		if len(s.stack) == 1 {
			return result
		}

		s.stack = s.stack[:len(s.stack)-1]
		parent := &s.stack[len(s.stack)-1]
		parent.lastResult = result
	}
}

func (s *stackSafeScanner[TInput]) push(parser Parser[TInput], inputStart int) {
	s.stack = append(s.stack, safeScanState[TInput]{
		parser:     parser,
		inputStart: inputStart,
	})
}

func (s *stackSafeScanner[TInput]) visit(state *safeScanState[TInput]) Parser[TInput] {
	if apply, ok := state.parser.(safeScanApplyParser[TInput]); ok {
		return s.visitApply(state, apply)
	}
	if best, ok := state.parser.(safeScanBestParser[TInput]); ok {
		return s.visitBest(state, best.safeScanBestParsers())
	}
	if first, ok := state.parser.(safeScanFirstParser[TInput]); ok {
		return s.visitFirst(state, first.safeScanFirstParsers())
	}
	if forward, ok := state.parser.(safeScanForwardParser[TInput]); ok {
		return s.visitForward(state, forward.safeScanForwardParser())
	}
	if conditional, ok := state.parser.(safeScanIfParser[TInput]); ok {
		test, parser := conditional.safeScanIfParts()
		return s.visitIf(state, test, parser)
	}
	if limit, ok := state.parser.(safeScanLimitParser[TInput]); ok {
		limiter, limited := limit.safeScanLimitParts()
		return s.visitLimit(state, limiter, limited)
	}
	if pattern, ok := state.parser.(*FailsParser[TInput]); ok {
		return s.visitFails(state, pattern.Pattern)
	}
	if pattern, ok := state.parser.(*NotParser[TInput]); ok {
		return s.visitNot(state, pattern.Pattern)
	}
	if repetition, ok := state.parser.(safeScanRepetitionParser[TInput]); ok {
		parser, zeroOrOne := repetition.safeScanRepetitionParts()
		return s.visitZeroOrMore(state, parser, zeroOrOne)
	}
	if oneOrMore, ok := state.parser.(*OneOrMoreParser[TInput]); ok {
		return s.visitOneOrMore(state, oneOrMore.Parser)
	}
	if sequence, ok := state.parser.(safeScanSequenceParser[TInput]); ok {
		return s.visitSequence(state, sequence.safeScanSequenceParsers())
	}
	if inner, ok := state.parser.(safeScanInnerParser[TInput]); ok {
		return s.visitInner(state, inner.safeScanInnerParser())
	}

	state.inputLength = state.parser.Scan(s.source, state.inputStart)
	return nil
}

func (s *stackSafeScanner[TInput]) visitApply(state *safeScanState[TInput], parser safeScanApplyParser[TInput]) Parser[TInput] {
	left, right, kind := parser.safeScanApplyParts()
	if state.state == 0 {
		state.state = 1
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

func (s *stackSafeScanner[TInput]) visitBest(state *safeScanState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if len(parsers) == 0 {
		state.inputLength = -1
		return nil
	}
	if state.state == 0 {
		state.bestFailedResult = -1
		state.bestSuccessResult = -1
		state.state = 1
		return parsers[0]
	}

	if state.lastResult > state.bestSuccessResult {
		state.bestSuccessResult = state.lastResult
	} else if state.lastResult < state.bestFailedResult {
		state.bestFailedResult = state.lastResult
	}
	if state.state >= len(parsers) {
		if state.bestSuccessResult >= 0 {
			state.inputLength = state.bestSuccessResult
		} else {
			state.inputLength = state.bestFailedResult
		}
		return nil
	}
	state.inputLength = 0
	next := parsers[state.state]
	state.state++
	return next
}

func (s *stackSafeScanner[TInput]) visitFirst(state *safeScanState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if len(parsers) == 0 {
		state.inputLength = -1
		return nil
	}
	if state.state == 0 {
		state.state = 1
		return parsers[0]
	}
	if state.lastResult < 0 {
		if state.lastResult < state.bestFailedResult {
			state.bestFailedResult = state.lastResult
		}
		if state.state < len(parsers) {
			next := parsers[state.state]
			state.state++
			return next
		}
		state.inputLength = state.bestFailedResult
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (s *stackSafeScanner[TInput]) visitForward(state *safeScanState[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		return parser
	}
	state.inputLength = state.lastResult
	return nil
}

func (s *stackSafeScanner[TInput]) visitIf(state *safeScanState[TInput], test Parser[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		return test
	}
	if state.state == 1 {
		if state.lastResult < 0 {
			state.inputLength = state.lastResult
			return nil
		}
		state.state = 2
		return parser
	}
	state.inputLength = state.lastResult
	return nil
}

func (s *stackSafeScanner[TInput]) visitLimit(state *safeScanState[TInput], limiter Parser[TInput], limited Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		length := limiter.Scan(s.source, state.inputStart)
		if length > 0 {
			state.previousSource = s.source
			s.source = NewLimitSource(s.source, state.inputStart+length)
			state.state = 1
			return limited
		}
		state.inputLength = -1
		return nil
	}
	s.source = state.previousSource
	state.inputLength = state.lastResult
	return nil
}

func (s *stackSafeScanner[TInput]) visitFails(state *safeScanState[TInput], pattern Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		if s.source.IsEnd(state.inputStart) {
			state.inputLength = 0
			return nil
		}
		state.state++
		return pattern
	}
	if state.lastResult >= 0 {
		state.inputLength = -1
	} else {
		state.inputLength = 0
	}
	return nil
}

func (s *stackSafeScanner[TInput]) visitNot(state *safeScanState[TInput], pattern Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		if s.source.IsEnd(state.inputStart) {
			state.inputLength = -1
			return nil
		}
		state.state++
		return pattern
	}
	if state.lastResult >= 0 {
		state.inputLength = -1
	} else {
		state.inputLength = 1
	}
	return nil
}

func (s *stackSafeScanner[TInput]) visitZeroOrMore(state *safeScanState[TInput], parser Parser[TInput], zeroOrOne bool) Parser[TInput] {
	if state.state == 0 || (state.lastResult > 0 && !zeroOrOne) {
		state.state++
		state.inputLength += state.lastResult
		return parser
	}
	return nil
}

func (s *stackSafeScanner[TInput]) visitOneOrMore(state *safeScanState[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		return parser
	}
	if state.state == 1 && state.lastResult < 0 {
		state.inputLength = state.lastResult
		return nil
	}
	if state.lastResult > 0 {
		state.state++
		state.inputLength += state.lastResult
		return parser
	}
	return nil
}

func (s *stackSafeScanner[TInput]) visitSequence(state *safeScanState[TInput], parsers []Parser[TInput]) Parser[TInput] {
	if state.lastResult < 0 {
		state.inputLength = -state.inputLength + state.lastResult
		return nil
	}
	state.inputLength += state.lastResult
	if state.state >= len(parsers) {
		return nil
	}
	next := parsers[state.state]
	state.state++
	return next
}

func (s *stackSafeScanner[TInput]) visitInner(state *safeScanState[TInput], parser Parser[TInput]) Parser[TInput] {
	if state.state == 0 {
		state.state = 1
		return parser
	}
	if state.lastResult < 0 && (state.parser.IsOptional() || state.parser.IsRequired()) {
		state.inputLength = 0
		return nil
	}
	state.inputLength = state.lastResult
	return nil
}

func (p *ApplyParser[TInput, TLeft, TOutput]) safeScanApplyParts() (Parser[TInput], Parser[TInput], ApplyKind) {
	return p.LeftParser, p.RightParser, p.ApplyKind
}

func (p *BestParser[TInput]) safeScanBestParsers() []Parser[TInput] {
	return p.Parsers
}

func (p *BestValueParser[TInput, TOutput]) safeScanBestParsers() []Parser[TInput] {
	return resultParsersAsParsers(p.Parsers)
}

func (p *ConvertParser[TInput, TOutput]) safeScanInnerParser() Parser[TInput] {
	return p.Pattern
}

func (p *FirstParser[TInput]) safeScanFirstParsers() []Parser[TInput] {
	return p.Parsers
}

func (p *FirstValueParser[TInput, TOutput]) safeScanFirstParsers() []Parser[TInput] {
	return resultParsersAsParsers(p.Parsers)
}

func (p *ForwardParser[TInput, TOutput]) safeScanForwardParser() Parser[TInput] {
	return p.DeferredParser()
}

func (p *IfParser[TInput]) safeScanIfParts() (Parser[TInput], Parser[TInput]) {
	return p.Test, p.Parser
}

func (p *IfValueParser[TInput, TOutput]) safeScanIfParts() (Parser[TInput], Parser[TInput]) {
	return p.Test, p.Parser
}

func (p *LimitParser[TInput, TOutput]) safeScanLimitParts() (Parser[TInput], Parser[TInput]) {
	return p.Limiter, p.Limited
}

func (p *OptionalParser[TInput, TOutput]) safeScanInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *ProduceParser[TInput, TProducer]) safeScanInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *RequiredParser[TInput, TOutput]) safeScanInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *RuleParser[TInput, TProducer]) safeScanSequenceParsers() []Parser[TInput] {
	return p.Parsers
}

func (p *SequenceParser[TInput]) safeScanSequenceParsers() []Parser[TInput] {
	return p.Parsers
}

func (p *ZeroOrMoreParser[TInput]) safeScanRepetitionParts() (Parser[TInput], bool) {
	return p.Parser, p.ZeroOrOne
}
