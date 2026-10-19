package kustolang

import "testing"

func TestCreateServerCommandGrammarPreservesConcreteTypes(t *testing.T) {
	tests := []struct {
		name       string
		serverKind string
		isType     func(ServerCommandGrammar) bool
		symbol     *CommandSymbol
	}{
		{name: "engine", serverKind: ServerKindEngine, isType: func(grammar ServerCommandGrammar) bool { _, ok := grammar.(*EngineCommandGrammar); return ok }, symbol: EngineCommands.ShowVersion},
		{name: "data-manager", serverKind: ServerKindDataManager, isType: func(grammar ServerCommandGrammar) bool { _, ok := grammar.(*DataManagerCommandGrammar); return ok }, symbol: DataManagerCommands.ShowVersion},
		{name: "cluster-manager", serverKind: ServerKindClusterManager, isType: func(grammar ServerCommandGrammar) bool { _, ok := grammar.(*ClusterManagerCommandGrammar); return ok }, symbol: ClusterManagerCommands.ShowVersion},
		{name: "aria-bridge", serverKind: ServerKindAriaBridge, isType: func(grammar ServerCommandGrammar) bool { _, ok := grammar.(*AriaBridgeCommandGrammar); return ok }, symbol: AriaBridgeCommands.ShowVersion},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals := NewGlobalState(nil).WithServerKind(test.serverKind)
			grammar := CreateServerCommandGrammar(globals)
			if !test.isType(grammar) || grammar.BaseCommandGrammar().Globals != globals {
				t.Fatalf("grammar = %T, globals = %#v", grammar, grammar.BaseCommandGrammar().Globals)
			}
			valid := grammar.BaseCommandGrammar().ParseCommandBlock(".show version")
			command, ok := firstCommandExpression(t, valid).(*CustomCommand)
			if !ok || command.CommandKind() != "ShowVersion" || command.ReferencedSymbol() != test.symbol {
				t.Fatalf("show-version command = %#v", command)
			}
			if diagnostics := valid.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid diagnostics = %#v", diagnostics)
			}
			result := test.symbol.ResultType()
			if result == nil || len(result.Columns()) != 4 || command.RawResultType() != &result.TypeSymbol {
				t.Fatalf("show-version result = command:%#v catalog:%#v", command.RawResultType(), result)
			}
			malformed := grammar.BaseCommandGrammar().ParseCommandBlock(".show")
			diagnostics := malformed.ContainedDiagnostics()
			if len(diagnostics) == 0 {
				t.Fatal("incomplete command produced no diagnostics")
			}
			if test.serverKind != ServerKindEngine && (len(diagnostics) != 1 || diagnostics[0].Code != "KS005" || diagnostics[0].Message != "Expected: version") {
				t.Fatalf("incomplete show-version diagnostics = %#v", diagnostics)
			}
			bad := grammar.BaseCommandGrammar().ParseCommandBlock(".")
			badDiagnostics := bad.ContainedDiagnostics()
			if len(badDiagnostics) != 1 || badDiagnostics[0].Code != "KS300" || badDiagnostics[0].Message != "Missing command." {
				t.Fatalf("bad-command diagnostics = %#v", badDiagnostics)
			}
			fresh := CreateServerCommandGrammar(globals)
			if fresh == grammar || fresh.BaseCommandGrammar() == grammar.BaseCommandGrammar() {
				t.Fatal("create factory reused grammar identity")
			}
		})
	}
}

func TestConcreteCommandGrammarConstructorsSelectParsersByDynamicType(t *testing.T) {
	dataGlobals := NewGlobalState(nil).WithServerKind(ServerKindDataManager)
	engine := NewEngineCommandGrammar(dataGlobals)
	if command, ok := firstCommandExpression(t, engine.ParseCommandBlock(".show queries")).(*CustomCommand); !ok || command.CommandKind() != "ShowQueries" {
		t.Fatalf("engine constructor command = %#v", command)
	}

	engineGlobals := NewGlobalState(nil).WithServerKind(ServerKindEngine)
	data := NewDataManagerCommandGrammar(engineGlobals)
	if command := firstCommandExpression(t, data.ParseCommandBlock(".show queries")); command.Kind() == SyntaxKindCustomCommand {
		if custom, ok := command.(*CustomCommand); ok && custom.CommandKind() == "ShowQueries" {
			t.Fatalf("data-manager constructor accepted engine command: %#v", command)
		}
	}
	if command, ok := firstCommandExpression(t, data.ParseCommandBlock(".show version")).(*CustomCommand); !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("data-manager constructor command = %#v", command)
	}
}

func TestServerCommandGrammarFromCachesConcreteViewOverBase(t *testing.T) {
	globals := NewGlobalState(nil).WithServerKind(ServerKindClusterManager).WithCache()
	first := ServerCommandGrammarFrom(globals)
	second := ServerCommandGrammarFrom(globals)
	if first != second {
		t.Fatalf("concrete grammar view was not cached: %p != %p", first, second)
	}
	if _, ok := first.(*ClusterManagerCommandGrammar); !ok {
		t.Fatalf("cached grammar type = %T", first)
	}
	if first.BaseCommandGrammar() != CommandGrammarFrom(globals) {
		t.Fatal("concrete view does not share the cached base grammar")
	}
	if defaultGrammar := ServerCommandGrammarFrom(nil); defaultGrammar.BaseCommandGrammar() != CommandGrammarFrom(nil) {
		t.Fatal("default concrete view does not share the default grammar")
	} else if _, ok := defaultGrammar.(*EngineCommandGrammar); !ok {
		t.Fatalf("default concrete grammar type = %T", defaultGrammar)
	}
}

func TestCreateServerCommandGrammarUnknownKindReturnsBase(t *testing.T) {
	globals := NewGlobalState(nil).WithServerKind("CustomServer")
	grammar := CreateServerCommandGrammar(globals)
	if _, ok := grammar.(*CommandGrammar); !ok {
		t.Fatalf("unknown server grammar type = %T", grammar)
	}
	if command := firstCommandExpression(t, grammar.BaseCommandGrammar().ParseCommandBlock(".show version")); command.Kind() == SyntaxKindCustomCommand {
		t.Fatalf("unknown server accepted known command: %#v", command)
	}
}
