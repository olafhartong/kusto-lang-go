package kustolang

// Ported from src/Kusto.Language/Parser/Combinators/Describer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// Describe returns a textual representation of a parser's grammar.
//
// Required parsers are marked with "!" by default. Pass false to omit the
// marker. DescribeParser is retained as the parser-specific compatibility
// spelling.
func Describe[TInput any](parser Parser[TInput], showRequired ...bool) string {
	return DescribeParser(parser, showRequired...)
}
