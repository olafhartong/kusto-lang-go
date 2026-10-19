package kustolang

import "fmt"

// Ported from CommandGrammar.CommandParserInfo and the constructor-local
// parser composition in src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// CommandParserInfo associates a command kind with its complete parser.
// Parser is expected to consume the leading dot and produce a Command.
type CommandParserInfo struct {
	Kind   string
	Parser ResultParser[LexicalToken, Command]
}

// NewCommandParserInfo constructs the source-shaped command parser pair.
func NewCommandParserInfo(kind string, parser ResultParser[LexicalToken, Command]) CommandParserInfo {
	return CommandParserInfo{Kind: kind, Parser: parser}
}

// CommandParserInfoProvider is the Go extension point corresponding to a
// CommandGrammar subclass overriding CreateCommandParsers.
type CommandParserInfoProvider interface {
	CreateCommandParsers(rules *PredefinedRuleParsers) []CommandParserInfo
}

// CommandParserInfoFactory adapts a function to CommandParserInfoProvider.
type CommandParserInfoFactory func(rules *PredefinedRuleParsers) []CommandParserInfo

func (f CommandParserInfoFactory) CreateCommandParsers(rules *PredefinedRuleParsers) []CommandParserInfo {
	if f == nil {
		return nil
	}
	return f(rules)
}

// NewCommandGrammarWithCommandParsers constructs a base command grammar from
// a fixed parser set. The supplied parsers replace the base grammar's empty
// CreateCommandParsers result.
func NewCommandGrammarWithCommandParsers(globals *GlobalState, infos ...CommandParserInfo) *CommandGrammar {
	grammar := NewCommandGrammar(globals)
	grammar.installCommandParserInfos(infos)
	return grammar
}

// NewCommandGrammarWithCommandParserFactory constructs a command grammar after
// its predefined rules exist, allowing parsers to refer to those rules.
func NewCommandGrammarWithCommandParserFactory(globals *GlobalState, factory CommandParserInfoFactory) *CommandGrammar {
	return NewCommandGrammarWithCommandParserProvider(globals, factory)
}

// NewCommandGrammarWithCommandParserProvider is the interface-based spelling
// for packages that model a derived CommandGrammar as a Go type.
func NewCommandGrammarWithCommandParserProvider(globals *GlobalState, provider CommandParserInfoProvider) *CommandGrammar {
	grammar := NewCommandGrammar(globals)
	if provider == nil {
		grammar.installCommandParserInfos(nil)
		return grammar
	}
	grammar.installCommandParserInfos(provider.CreateCommandParsers(grammar.PredefinedRules))
	return grammar
}

func (g *CommandGrammar) installCommandParserInfos(infos []CommandParserInfo) {
	if g == nil {
		panic("cannot install command parsers on a nil CommandGrammar")
	}

	runtimeInfos := make([]commandParserInfo, 0, len(infos))
	parsers := make([]Parser[LexicalToken], 0, len(infos))
	for index, info := range infos {
		if info.Parser == nil {
			panic(fmt.Sprintf("command parser info %d (%q) has a nil parser", index, info.Kind))
		}

		parser := asResultParser[LexicalToken, Command](info.Parser.WithTag(info.Kind))
		runtimeInfos = append(runtimeInfos, commandParserInfo{
			Kind: info.Kind,
			Parse: func(state *queryParserState) Command {
				if state == nil {
					return nil
				}
				result := parser.ParseValue(NewArraySource(state.tokens), state.pos)
				if result.Length < 0 || isNilSyntaxElement(result.Value) {
					return nil
				}
				state.pos += result.Length
				return result.Value
			},
		})
		parsers = append(parsers, parser)
	}

	g.commandParsers = runtimeInfos
	g.allowPartialKnownCommand = len(parsers) > 0

	// The native generated grammars derive partial/completion parsers from
	// descriptors. Extension parsers are already the source parser trees, so
	// preserve them directly for partial failure recovery and annotations.
	g.commandPartialParsers = append([]Parser[LexicalToken]{}, parsers...)
	g.commandPartialOnce.Do(func() {})
	if len(parsers) == 1 {
		g.commandCompletionParser = parsers[0]
	} else if len(parsers) > 1 {
		g.commandCompletionParser = Best(parsers...)
	}
	g.commandCompletionOnce.Do(func() {})
}
