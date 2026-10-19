package kustolang

// Ported from the generated server grammar classes at upstream commit
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856:
//
//   src/Kusto.Language/Parser/CodeGen/EngineCommandGrammar.cs
//   src/Kusto.Language/Parser/CodeGen/DataManagerCommandGrammar.cs
//   src/Kusto.Language/Parser/CodeGen/ClusterManagerCommandGrammar.cs
//   src/Kusto.Language/Parser/CodeGen/AriaBridgeCommandGrammar.cs
//
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// ServerCommandGrammar is the Go adaptation of the CommandGrammar inheritance
// hierarchy. BaseCommandGrammar exposes the common parser without erasing the
// dynamic concrete grammar value returned by the typed factories.
type ServerCommandGrammar interface {
	BaseCommandGrammar() *CommandGrammar
}

func (g *CommandGrammar) BaseCommandGrammar() *CommandGrammar { return g }

type EngineCommandGrammar struct{ *CommandGrammar }
type DataManagerCommandGrammar struct{ *CommandGrammar }
type ClusterManagerCommandGrammar struct{ *CommandGrammar }
type AriaBridgeCommandGrammar struct{ *CommandGrammar }

func NewEngineCommandGrammar(globals *GlobalState) *EngineCommandGrammar {
	grammar := NewCommandGrammar(globals)
	grammar.commandParsers = engineCommandParserInfos()
	grammar.allowPartialKnownCommand = true
	return &EngineCommandGrammar{CommandGrammar: grammar}
}

func NewDataManagerCommandGrammar(globals *GlobalState) *DataManagerCommandGrammar {
	return &DataManagerCommandGrammar{CommandGrammar: newShowVersionCommandGrammar(globals)}
}

func NewClusterManagerCommandGrammar(globals *GlobalState) *ClusterManagerCommandGrammar {
	return &ClusterManagerCommandGrammar{CommandGrammar: newShowVersionCommandGrammar(globals)}
}

func NewAriaBridgeCommandGrammar(globals *GlobalState) *AriaBridgeCommandGrammar {
	return &AriaBridgeCommandGrammar{CommandGrammar: newShowVersionCommandGrammar(globals)}
}

func newShowVersionCommandGrammar(globals *GlobalState) *CommandGrammar {
	grammar := NewCommandGrammar(globals)
	grammar.commandParsers = []commandParserInfo{showVersionCommandParserInfo()}
	grammar.allowPartialKnownCommand = true
	return grammar
}

// CreateServerCommandGrammar preserves the concrete server grammar type. The
// existing CreateCommandGrammar function remains the base-returning API for
// backward compatibility with callers that access CommandGrammar fields.
func CreateServerCommandGrammar(globals *GlobalState) ServerCommandGrammar {
	if globals == nil {
		globals = DefaultGlobalState
	}
	switch globals.ServerKind {
	case ServerKindEngine:
		return NewEngineCommandGrammar(globals)
	case ServerKindDataManager:
		return NewDataManagerCommandGrammar(globals)
	case ServerKindClusterManager:
		return NewClusterManagerCommandGrammar(globals)
	case ServerKindAriaBridge:
		return NewAriaBridgeCommandGrammar(globals)
	default:
		return NewCommandGrammar(globals)
	}
}

// ServerCommandGrammarFrom returns a cached concrete view over the same base
// grammar and parser state returned by CommandGrammarFrom.
func ServerCommandGrammarFrom(globals *GlobalState) ServerCommandGrammar {
	base := CommandGrammarFrom(globals)
	if base == nil || base.Globals == nil {
		return base
	}
	if base.Globals.ServerKind != ServerKindEngine &&
		base.Globals.ServerKind != ServerKindDataManager &&
		base.Globals.ServerKind != ServerKindClusterManager &&
		base.Globals.ServerKind != ServerKindAriaBridge {
		return base
	}
	base.serverViewOnce.Do(func() {
		switch base.Globals.ServerKind {
		case ServerKindEngine:
			base.serverView = &EngineCommandGrammar{CommandGrammar: base}
		case ServerKindDataManager:
			base.serverView = &DataManagerCommandGrammar{CommandGrammar: base}
		case ServerKindClusterManager:
			base.serverView = &ClusterManagerCommandGrammar{CommandGrammar: base}
		case ServerKindAriaBridge:
			base.serverView = &AriaBridgeCommandGrammar{CommandGrammar: base}
		}
	})
	return base.serverView
}
