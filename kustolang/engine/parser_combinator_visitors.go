package kustolang

// Ported from the concrete IsParentVisitor behavior in
// src/Kusto.Language/Parser/Combinators/ParserVisitors.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type parserParentPredicate[TInput any] interface {
	isParserParentOf(child Parser[TInput]) bool
}

type parserVisitorConvertParser[TInput any] interface {
	parserVisitorConvertPattern() Parser[TInput]
}

type ParserVisitorFuncs[TInput any] struct {
	Apply       func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind)
	Best        func(parser Parser[TInput], parsers []Parser[TInput])
	Convert     func(parser Parser[TInput], pattern Parser[TInput])
	Fails       func(parser Parser[TInput], pattern Parser[TInput])
	First       func(parser Parser[TInput], parsers []Parser[TInput])
	Forward     func(parser Parser[TInput], deferred Parser[TInput])
	If          func(parser Parser[TInput], test Parser[TInput], body Parser[TInput])
	Limit       func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput])
	Map         func(parser Parser[TInput])
	Match       func(parser Parser[TInput])
	Not         func(parser Parser[TInput], pattern Parser[TInput])
	OneOrMore   func(parser Parser[TInput], repeated Parser[TInput])
	Optional    func(parser Parser[TInput], optional Parser[TInput])
	Produce     func(parser Parser[TInput], inner Parser[TInput])
	Required    func(parser Parser[TInput], required Parser[TInput])
	Rule        func(parser Parser[TInput], parsers []Parser[TInput])
	Sequence    func(parser Parser[TInput], parsers []Parser[TInput])
	ZeroOrMore  func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool)
	Unsupported func(parser Parser[TInput])
}

type ParserVisitorResultFuncs[TInput any, TResult any] struct {
	Apply       func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) TResult
	Best        func(parser Parser[TInput], parsers []Parser[TInput]) TResult
	Convert     func(parser Parser[TInput], pattern Parser[TInput]) TResult
	Fails       func(parser Parser[TInput], pattern Parser[TInput]) TResult
	First       func(parser Parser[TInput], parsers []Parser[TInput]) TResult
	Forward     func(parser Parser[TInput], deferred Parser[TInput]) TResult
	If          func(parser Parser[TInput], test Parser[TInput], body Parser[TInput]) TResult
	Limit       func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput]) TResult
	Map         func(parser Parser[TInput]) TResult
	Match       func(parser Parser[TInput]) TResult
	Not         func(parser Parser[TInput], pattern Parser[TInput]) TResult
	OneOrMore   func(parser Parser[TInput], repeated Parser[TInput]) TResult
	Optional    func(parser Parser[TInput], optional Parser[TInput]) TResult
	Produce     func(parser Parser[TInput], inner Parser[TInput]) TResult
	Required    func(parser Parser[TInput], required Parser[TInput]) TResult
	Rule        func(parser Parser[TInput], parsers []Parser[TInput]) TResult
	Sequence    func(parser Parser[TInput], parsers []Parser[TInput]) TResult
	ZeroOrMore  func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool) TResult
	Unsupported func(parser Parser[TInput]) TResult
}

type ParserVisitorArgResultFuncs[TInput any, TArg any, TResult any] struct {
	Apply       func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind, arg TArg) TResult
	Best        func(parser Parser[TInput], parsers []Parser[TInput], arg TArg) TResult
	Convert     func(parser Parser[TInput], pattern Parser[TInput], arg TArg) TResult
	Fails       func(parser Parser[TInput], pattern Parser[TInput], arg TArg) TResult
	First       func(parser Parser[TInput], parsers []Parser[TInput], arg TArg) TResult
	Forward     func(parser Parser[TInput], deferred Parser[TInput], arg TArg) TResult
	If          func(parser Parser[TInput], test Parser[TInput], body Parser[TInput], arg TArg) TResult
	Limit       func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput], arg TArg) TResult
	Map         func(parser Parser[TInput], arg TArg) TResult
	Match       func(parser Parser[TInput], arg TArg) TResult
	Not         func(parser Parser[TInput], pattern Parser[TInput], arg TArg) TResult
	OneOrMore   func(parser Parser[TInput], repeated Parser[TInput], arg TArg) TResult
	Optional    func(parser Parser[TInput], optional Parser[TInput], arg TArg) TResult
	Produce     func(parser Parser[TInput], inner Parser[TInput], arg TArg) TResult
	Required    func(parser Parser[TInput], required Parser[TInput], arg TArg) TResult
	Rule        func(parser Parser[TInput], parsers []Parser[TInput], arg TArg) TResult
	Sequence    func(parser Parser[TInput], parsers []Parser[TInput], arg TArg) TResult
	ZeroOrMore  func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool, arg TArg) TResult
	Unsupported func(parser Parser[TInput], arg TArg) TResult
}

func AcceptParserVisitor[TInput any](parser Parser[TInput], visitor ParserVisitorFuncs[TInput]) {
	_ = AcceptParserVisitorResult(parser, ParserVisitorResultFuncs[TInput, struct{}]{
		Apply: func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) struct{} {
			if visitor.Apply != nil {
				visitor.Apply(parser, left, right, kind)
			}
			return struct{}{}
		},
		Best: func(parser Parser[TInput], parsers []Parser[TInput]) struct{} {
			if visitor.Best != nil {
				visitor.Best(parser, parsers)
			}
			return struct{}{}
		},
		Convert: func(parser Parser[TInput], pattern Parser[TInput]) struct{} {
			if visitor.Convert != nil {
				visitor.Convert(parser, pattern)
			}
			return struct{}{}
		},
		Fails: func(parser Parser[TInput], pattern Parser[TInput]) struct{} {
			if visitor.Fails != nil {
				visitor.Fails(parser, pattern)
			}
			return struct{}{}
		},
		First: func(parser Parser[TInput], parsers []Parser[TInput]) struct{} {
			if visitor.First != nil {
				visitor.First(parser, parsers)
			}
			return struct{}{}
		},
		Forward: func(parser Parser[TInput], deferred Parser[TInput]) struct{} {
			if visitor.Forward != nil {
				visitor.Forward(parser, deferred)
			}
			return struct{}{}
		},
		If: func(parser Parser[TInput], test Parser[TInput], body Parser[TInput]) struct{} {
			if visitor.If != nil {
				visitor.If(parser, test, body)
			}
			return struct{}{}
		},
		Limit: func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput]) struct{} {
			if visitor.Limit != nil {
				visitor.Limit(parser, limiter, limited)
			}
			return struct{}{}
		},
		Map: func(parser Parser[TInput]) struct{} {
			if visitor.Map != nil {
				visitor.Map(parser)
			}
			return struct{}{}
		},
		Match: func(parser Parser[TInput]) struct{} {
			if visitor.Match != nil {
				visitor.Match(parser)
			}
			return struct{}{}
		},
		Not: func(parser Parser[TInput], pattern Parser[TInput]) struct{} {
			if visitor.Not != nil {
				visitor.Not(parser, pattern)
			}
			return struct{}{}
		},
		OneOrMore: func(parser Parser[TInput], repeated Parser[TInput]) struct{} {
			if visitor.OneOrMore != nil {
				visitor.OneOrMore(parser, repeated)
			}
			return struct{}{}
		},
		Optional: func(parser Parser[TInput], optional Parser[TInput]) struct{} {
			if visitor.Optional != nil {
				visitor.Optional(parser, optional)
			}
			return struct{}{}
		},
		Produce: func(parser Parser[TInput], inner Parser[TInput]) struct{} {
			if visitor.Produce != nil {
				visitor.Produce(parser, inner)
			}
			return struct{}{}
		},
		Required: func(parser Parser[TInput], required Parser[TInput]) struct{} {
			if visitor.Required != nil {
				visitor.Required(parser, required)
			}
			return struct{}{}
		},
		Rule: func(parser Parser[TInput], parsers []Parser[TInput]) struct{} {
			if visitor.Rule != nil {
				visitor.Rule(parser, parsers)
			}
			return struct{}{}
		},
		Sequence: func(parser Parser[TInput], parsers []Parser[TInput]) struct{} {
			if visitor.Sequence != nil {
				visitor.Sequence(parser, parsers)
			}
			return struct{}{}
		},
		ZeroOrMore: func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool) struct{} {
			if visitor.ZeroOrMore != nil {
				visitor.ZeroOrMore(parser, repeated, zeroOrOne)
			}
			return struct{}{}
		},
		Unsupported: func(parser Parser[TInput]) struct{} {
			if visitor.Unsupported != nil {
				visitor.Unsupported(parser)
			}
			return struct{}{}
		},
	})
}

func AcceptParserVisitorResult[TInput any, TResult any](parser Parser[TInput], visitor ParserVisitorResultFuncs[TInput, TResult]) TResult {
	var zero TResult
	if parser == nil {
		if visitor.Unsupported != nil {
			return visitor.Unsupported(parser)
		}
		return zero
	}

	if apply, ok := parser.(safeScanApplyParser[TInput]); ok {
		left, right, kind := apply.safeScanApplyParts()
		if visitor.Apply != nil {
			return visitor.Apply(parser, left, right, kind)
		}
		return zero
	}
	if best, ok := parser.(safeScanBestParser[TInput]); ok {
		if visitor.Best != nil {
			return visitor.Best(parser, best.safeScanBestParsers())
		}
		return zero
	}
	if convert, ok := parser.(parserVisitorConvertParser[TInput]); ok {
		if visitor.Convert != nil {
			return visitor.Convert(parser, convert.parserVisitorConvertPattern())
		}
		return zero
	}
	if pattern, ok := parser.(*FailsParser[TInput]); ok {
		if visitor.Fails != nil {
			return visitor.Fails(parser, pattern.Pattern)
		}
		return zero
	}
	if first, ok := parser.(safeScanFirstParser[TInput]); ok {
		if visitor.First != nil {
			return visitor.First(parser, first.safeScanFirstParsers())
		}
		return zero
	}
	if forward, ok := parser.(safeScanForwardParser[TInput]); ok {
		if visitor.Forward != nil {
			return visitor.Forward(parser, forward.safeScanForwardParser())
		}
		return zero
	}
	if conditional, ok := parser.(safeScanIfParser[TInput]); ok {
		test, body := conditional.safeScanIfParts()
		if visitor.If != nil {
			return visitor.If(parser, test, body)
		}
		return zero
	}
	if limit, ok := parser.(safeScanLimitParser[TInput]); ok {
		limiter, limited := limit.safeScanLimitParts()
		if visitor.Limit != nil {
			return visitor.Limit(parser, limiter, limited)
		}
		return zero
	}
	if term, ok := parser.(describerTermParser); ok && term.describeTerm() == "map()" {
		if visitor.Map != nil {
			return visitor.Map(parser)
		}
		return zero
	}
	if pattern, ok := parser.(*NotParser[TInput]); ok {
		if visitor.Not != nil {
			return visitor.Not(parser, pattern.Pattern)
		}
		return zero
	}
	if oneOrMore, ok := parser.(*OneOrMoreParser[TInput]); ok {
		if visitor.OneOrMore != nil {
			return visitor.OneOrMore(parser, oneOrMore.Parser)
		}
		return zero
	}
	if optional, ok := parser.(safeParseOptionalParser[TInput]); ok {
		if visitor.Optional != nil {
			return visitor.Optional(parser, optional.safeParseInnerParser())
		}
		return zero
	}
	if produce, ok := parser.(safeParseProduceParser[TInput]); ok {
		if visitor.Produce != nil {
			return visitor.Produce(parser, produce.safeParseInnerParser())
		}
		return zero
	}
	if required, ok := parser.(safeParseRequiredParser[TInput]); ok {
		if visitor.Required != nil {
			return visitor.Required(parser, required.safeParseInnerParser())
		}
		return zero
	}
	if rule, ok := parser.(safeParseRuleProducer); ok {
		_ = rule
		if sequence, ok := parser.(safeScanSequenceParser[TInput]); ok {
			if visitor.Rule != nil {
				return visitor.Rule(parser, sequence.safeScanSequenceParsers())
			}
			return zero
		}
	}
	if sequence, ok := parser.(safeScanSequenceParser[TInput]); ok {
		if visitor.Sequence != nil {
			return visitor.Sequence(parser, sequence.safeScanSequenceParsers())
		}
		return zero
	}
	if repetition, ok := parser.(safeScanRepetitionParser[TInput]); ok {
		repeated, zeroOrOne := repetition.safeScanRepetitionParts()
		if visitor.ZeroOrMore != nil {
			return visitor.ZeroOrMore(parser, repeated, zeroOrOne)
		}
		return zero
	}
	if parser.IsMatch() {
		if visitor.Match != nil {
			return visitor.Match(parser)
		}
		return zero
	}
	if visitor.Unsupported != nil {
		return visitor.Unsupported(parser)
	}
	return zero
}

func AcceptParserVisitorArgResult[TInput any, TArg any, TResult any](parser Parser[TInput], visitor ParserVisitorArgResultFuncs[TInput, TArg, TResult], arg TArg) TResult {
	return AcceptParserVisitorResult(parser, ParserVisitorResultFuncs[TInput, TResult]{
		Apply: func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) TResult {
			if visitor.Apply != nil {
				return visitor.Apply(parser, left, right, kind, arg)
			}
			var zero TResult
			return zero
		},
		Best: func(parser Parser[TInput], parsers []Parser[TInput]) TResult {
			if visitor.Best != nil {
				return visitor.Best(parser, parsers, arg)
			}
			var zero TResult
			return zero
		},
		Convert: func(parser Parser[TInput], pattern Parser[TInput]) TResult {
			if visitor.Convert != nil {
				return visitor.Convert(parser, pattern, arg)
			}
			var zero TResult
			return zero
		},
		Fails: func(parser Parser[TInput], pattern Parser[TInput]) TResult {
			if visitor.Fails != nil {
				return visitor.Fails(parser, pattern, arg)
			}
			var zero TResult
			return zero
		},
		First: func(parser Parser[TInput], parsers []Parser[TInput]) TResult {
			if visitor.First != nil {
				return visitor.First(parser, parsers, arg)
			}
			var zero TResult
			return zero
		},
		Forward: func(parser Parser[TInput], deferred Parser[TInput]) TResult {
			if visitor.Forward != nil {
				return visitor.Forward(parser, deferred, arg)
			}
			var zero TResult
			return zero
		},
		If: func(parser Parser[TInput], test Parser[TInput], body Parser[TInput]) TResult {
			if visitor.If != nil {
				return visitor.If(parser, test, body, arg)
			}
			var zero TResult
			return zero
		},
		Limit: func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput]) TResult {
			if visitor.Limit != nil {
				return visitor.Limit(parser, limiter, limited, arg)
			}
			var zero TResult
			return zero
		},
		Map: func(parser Parser[TInput]) TResult {
			if visitor.Map != nil {
				return visitor.Map(parser, arg)
			}
			var zero TResult
			return zero
		},
		Match: func(parser Parser[TInput]) TResult {
			if visitor.Match != nil {
				return visitor.Match(parser, arg)
			}
			var zero TResult
			return zero
		},
		Not: func(parser Parser[TInput], pattern Parser[TInput]) TResult {
			if visitor.Not != nil {
				return visitor.Not(parser, pattern, arg)
			}
			var zero TResult
			return zero
		},
		OneOrMore: func(parser Parser[TInput], repeated Parser[TInput]) TResult {
			if visitor.OneOrMore != nil {
				return visitor.OneOrMore(parser, repeated, arg)
			}
			var zero TResult
			return zero
		},
		Optional: func(parser Parser[TInput], optional Parser[TInput]) TResult {
			if visitor.Optional != nil {
				return visitor.Optional(parser, optional, arg)
			}
			var zero TResult
			return zero
		},
		Produce: func(parser Parser[TInput], inner Parser[TInput]) TResult {
			if visitor.Produce != nil {
				return visitor.Produce(parser, inner, arg)
			}
			var zero TResult
			return zero
		},
		Required: func(parser Parser[TInput], required Parser[TInput]) TResult {
			if visitor.Required != nil {
				return visitor.Required(parser, required, arg)
			}
			var zero TResult
			return zero
		},
		Rule: func(parser Parser[TInput], parsers []Parser[TInput]) TResult {
			if visitor.Rule != nil {
				return visitor.Rule(parser, parsers, arg)
			}
			var zero TResult
			return zero
		},
		Sequence: func(parser Parser[TInput], parsers []Parser[TInput]) TResult {
			if visitor.Sequence != nil {
				return visitor.Sequence(parser, parsers, arg)
			}
			var zero TResult
			return zero
		},
		ZeroOrMore: func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool) TResult {
			if visitor.ZeroOrMore != nil {
				return visitor.ZeroOrMore(parser, repeated, zeroOrOne, arg)
			}
			var zero TResult
			return zero
		},
		Unsupported: func(parser Parser[TInput]) TResult {
			if visitor.Unsupported != nil {
				return visitor.Unsupported(parser, arg)
			}
			var zero TResult
			return zero
		},
	})
}

func IsParserParent[TInput any](parent Parser[TInput], child Parser[TInput]) bool {
	if parent == nil {
		return false
	}
	if predicate, ok := parent.(parserParentPredicate[TInput]); ok {
		return predicate.isParserParentOf(child)
	}
	return parent.ChildParserIndex(child) >= 0
}

func (p *FailsParser[TInput]) isParserParentOf(child Parser[TInput]) bool {
	return p.Pattern != child
}

func (p *ConvertParser[TInput, TOutput]) parserVisitorConvertPattern() Parser[TInput] {
	return p.Pattern
}
