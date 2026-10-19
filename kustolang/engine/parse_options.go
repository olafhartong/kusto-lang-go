package kustolang

// Ported from src/Kusto.Language/Parser/ParseOptions.cs and the ParserKind
// enum in src/Kusto.Language/KustoCode.cs. The current parser always follows
// the native QueryParser path, but preserving these immutable options keeps the
// public model aligned with upstream GlobalState/KustoCode behavior.

type ParserKind int

const (
	ParserKindGrammar ParserKind = iota
	ParserKindDefault
)

type ParseOptions struct {
	AlwaysProduceEndToken         bool
	AllowLiteralsWithLineBreaks   bool
	AllowNonAdjacentWildcardParts bool
	ParserKind                    ParserKind
}

var DefaultParseOptions = ParseOptions{
	AlwaysProduceEndToken:         true,
	AllowLiteralsWithLineBreaks:   false,
	AllowNonAdjacentWildcardParts: false,
	ParserKind:                    ParserKindDefault,
}

func (o ParseOptions) WithAlwaysProduceEndTokens(alwaysProduceEndTokens bool) ParseOptions {
	o.AlwaysProduceEndToken = alwaysProduceEndTokens
	return o
}

func (o ParseOptions) WithAllowLiteralsWithLineBreaks(allow bool) ParseOptions {
	o.AllowLiteralsWithLineBreaks = allow
	return o
}

func (o ParseOptions) WithAllowNonAdjacentWildcardParts(allow bool) ParseOptions {
	o.AllowNonAdjacentWildcardParts = allow
	return o
}

func (o ParseOptions) WithParserKind(kind ParserKind) ParseOptions {
	o.ParserKind = kind
	return o
}

func (o ParseOptions) Equal(other ParseOptions) bool {
	return o.EqualExceptForParseKind(other) && o.ParserKind == other.ParserKind
}

func (o ParseOptions) EqualExceptForParseKind(other ParseOptions) bool {
	return o.AlwaysProduceEndToken == other.AlwaysProduceEndToken &&
		o.AllowLiteralsWithLineBreaks == other.AllowLiteralsWithLineBreaks &&
		o.AllowNonAdjacentWildcardParts == other.AllowNonAdjacentWildcardParts
}
