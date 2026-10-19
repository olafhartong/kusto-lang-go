package kustolang

// Ported from CommandGrammar.CommandParserInfo and the constructor-local
// parser composition in src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type CommandParserInfo = internal.CommandParserInfo
type CommandParserInfoProvider = internal.CommandParserInfoProvider
type CommandParserInfoFactory = internal.CommandParserInfoFactory

func NewCommandParserInfo(kind string, parser ResultParser[LexicalToken, Command]) CommandParserInfo {
	return internal.NewCommandParserInfo(kind, parser)
}

func NewCommandGrammarWithCommandParsers(globals *GlobalState, infos ...CommandParserInfo) *CommandGrammar {
	return internal.NewCommandGrammarWithCommandParsers(globals, infos...)
}

func NewCommandGrammarWithCommandParserFactory(globals *GlobalState, factory CommandParserInfoFactory) *CommandGrammar {
	return internal.NewCommandGrammarWithCommandParserFactory(globals, factory)
}

func NewCommandGrammarWithCommandParserProvider(globals *GlobalState, provider CommandParserInfoProvider) *CommandGrammar {
	return internal.NewCommandGrammarWithCommandParserProvider(globals, provider)
}
