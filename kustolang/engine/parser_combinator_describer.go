package kustolang

import "strings"

// Ported from src/Kusto.Language/Parser/Combinators/Describer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type describerApplyParser[TInput any] interface {
	describeApplyParts() (Parser[TInput], Parser[TInput], ApplyKind)
}

type describerInnerParser[TInput any] interface {
	describeInnerParser() Parser[TInput]
}

type describerNegationParser interface {
	describeNegationTerm() string
}

type describerTermParser interface {
	describeTerm() string
}

type describerZeroOrMoreParser interface {
	describeZeroOrOne() bool
}

func DescribeParser[TInput any](parser Parser[TInput], showRequired ...bool) string {
	show := true
	if len(showRequired) > 0 {
		show = showRequired[0]
	}
	writer := newParserDescriber[TInput](show)
	writer.visit(parser)
	return writer.String()
}

type parserDescriber[TInput any] struct {
	builder      strings.Builder
	separator    string
	hasSeparator bool
	showRequired bool
}

func newParserDescriber[TInput any](showRequired bool) *parserDescriber[TInput] {
	return &parserDescriber[TInput]{showRequired: showRequired}
}

func (w *parserDescriber[TInput]) String() string {
	return w.builder.String()
}

func (w *parserDescriber[TInput]) visit(parser Parser[TInput]) {
	if parser == nil || parser.IsHidden() {
		return
	}

	if tag := parser.Tag(); tag != "" {
		w.writeTerm(tag)
		return
	}

	if apply, ok := parser.(describerApplyParser[TInput]); ok {
		w.visitApply(apply)
		return
	}

	if parser.IsForward() {
		w.writeTerm("forward()")
		return
	}

	if parser.IsNegation() {
		term := "not"
		if negation, ok := parser.(describerNegationParser); ok {
			term = negation.describeNegationTerm()
		}
		w.writeBracketed(term+"(", ")", parser.ChildParser(0))
		return
	}

	if parser.IsAlternation() {
		w.writeAlternation(childParsers(parser))
		return
	}

	if parser.IsSequence() {
		w.writeSequence(childParsers(parser))
		return
	}

	if parser.IsRequired() {
		if w.showRequired {
			w.writeRequired(parser.ChildParser(0))
		} else {
			w.visit(parser.ChildParser(0))
		}
		return
	}

	if parser.IsRepetition() {
		if parser.IsOptional() {
			if zeroOrMore, ok := parser.(describerZeroOrMoreParser); ok && zeroOrMore.describeZeroOrOne() {
				w.writeOptional(parser.ChildParser(0))
			} else {
				w.writeZeroOrMore(parser.ChildParser(0))
			}
		} else {
			w.writeOneOrMore(parser.ChildParser(0))
		}
		return
	}

	if parser.IsOptional() {
		w.writeOptional(parser.ChildParser(0))
		return
	}

	if inner, ok := parser.(describerInnerParser[TInput]); ok {
		w.visit(inner.describeInnerParser())
		return
	}

	if term, ok := parser.(describerTermParser); ok {
		w.writeTerm(term.describeTerm())
		return
	}

	if parser.IsMatch() {
		w.writeTerm("match()")
		return
	}
}

func (w *parserDescriber[TInput]) visitApply(parser describerApplyParser[TInput]) {
	left, right, kind := parser.describeApplyParts()
	rightWriter := newParserDescriber[TInput](w.showRequired)
	switch kind {
	case ApplyKindZeroOrMore:
		rightWriter.writeZeroOrMore(right)
	case ApplyKindZeroOrOne:
		rightWriter.writeOptional(right)
	default:
		rightWriter.visit(right)
	}

	leftWriter := newParserDescriber[TInput](w.showRequired)
	leftWriter.visit(left)
	w.writeSeparatedWriters(" ", []*parserDescriber[TInput]{leftWriter, rightWriter})
}

func (w *parserDescriber[TInput]) writeTerm(term string) {
	if term != "" {
		w.builder.WriteString(term)
	}
}

func (w *parserDescriber[TInput]) writeAlternation(parsers []Parser[TInput]) {
	w.writeSeparated(" | ", parsers)
}

func (w *parserDescriber[TInput]) writeSequence(parsers []Parser[TInput]) {
	w.writeSeparated(" ", parsers)
}

func (w *parserDescriber[TInput]) writeOptional(parser Parser[TInput]) {
	w.writeBracketed("[", "]", parser)
}

func (w *parserDescriber[TInput]) writeRequired(parser Parser[TInput]) {
	nested := newParserDescriber[TInput](w.showRequired)
	nested.visit(parser)
	if nested.hasSeparator {
		w.builder.WriteString("(")
		w.builder.WriteString(nested.String())
		w.builder.WriteString(")")
	} else {
		w.builder.WriteString(nested.String())
	}
	w.builder.WriteString("!")
}

func (w *parserDescriber[TInput]) writeZeroOrMore(parser Parser[TInput]) {
	w.writeBracketed("{", "}", parser)
}

func (w *parserDescriber[TInput]) writeOneOrMore(parser Parser[TInput]) {
	w.writeBracketed("{", "}+", parser)
}

func (w *parserDescriber[TInput]) writeSeparated(separator string, parsers []Parser[TInput]) {
	builders := make([]*parserDescriber[TInput], 0, len(parsers))
	for _, parser := range parsers {
		if parser == nil || parser.IsHidden() {
			continue
		}
		nested := newParserDescriber[TInput](w.showRequired)
		nested.visit(parser)
		if nested.builder.Len() > 0 {
			builders = append(builders, nested)
		}
	}
	w.writeSeparatedWriters(separator, builders)
}

func (w *parserDescriber[TInput]) writeSeparatedWriters(separator string, builders []*parserDescriber[TInput]) {
	if len(builders) > 1 {
		w.separator = separator
		w.hasSeparator = true
		for i, nested := range builders {
			if i > 0 {
				w.builder.WriteString(separator)
			}
			if nested.hasSeparator && nested.separator != separator {
				w.builder.WriteString("(")
				w.builder.WriteString(nested.String())
				w.builder.WriteString(")")
			} else {
				w.builder.WriteString(nested.String())
			}
		}
	} else if len(builders) == 1 {
		w.builder.Reset()
		w.builder.WriteString(builders[0].String())
		w.separator = builders[0].separator
		w.hasSeparator = builders[0].hasSeparator
	}
}

func (w *parserDescriber[TInput]) writeBracketed(startBracket string, endBracket string, parser Parser[TInput]) {
	nested := newParserDescriber[TInput](w.showRequired)
	nested.visit(parser)
	if nested.builder.Len() > 0 {
		w.builder.WriteString(startBracket)
		w.builder.WriteString(nested.String())
		w.builder.WriteString(endBracket)
	}
}

func childParsers[TInput any](parser Parser[TInput]) []Parser[TInput] {
	count := parser.ChildParserCount()
	children := make([]Parser[TInput], 0, count)
	for i := 0; i < count; i++ {
		children = append(children, parser.ChildParser(i))
	}
	return children
}

func (p *ApplyParser[TInput, TLeft, TOutput]) describeApplyParts() (Parser[TInput], Parser[TInput], ApplyKind) {
	return p.LeftParser, p.RightParser, p.ApplyKind
}

func (p *ConvertParser[TInput, TOutput]) describeInnerParser() Parser[TInput] {
	return p.Pattern
}

func (p *FailsParser[TInput]) describeNegationTerm() string {
	return "fails"
}

func (p *IfParser[TInput]) describeInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *IfValueParser[TInput, TOutput]) describeInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *LimitParser[TInput, TOutput]) describeInnerParser() Parser[TInput] {
	return p.Limited
}

func (p *MapParser[TInput, TOutput]) describeTerm() string {
	return "map()"
}

func (p *NotParser[TInput]) describeNegationTerm() string {
	return "not"
}

func (p *ProduceParser[TInput, TProducer]) describeInnerParser() Parser[TInput] {
	return p.Parser
}

func (p *ZeroOrMoreParser[TInput]) describeZeroOrOne() bool {
	return p.ZeroOrOne
}
