package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicServerCommandGrammarConcreteFactory(t *testing.T) {
	tests := []struct {
		name       string
		serverKind string
		isType     func(kustolang.ServerCommandGrammar) bool
	}{
		{name: "engine", serverKind: kustolang.ServerKindEngine, isType: func(grammar kustolang.ServerCommandGrammar) bool {
			_, ok := grammar.(*kustolang.EngineCommandGrammar)
			return ok
		}},
		{name: "data-manager", serverKind: kustolang.ServerKindDataManager, isType: func(grammar kustolang.ServerCommandGrammar) bool {
			_, ok := grammar.(*kustolang.DataManagerCommandGrammar)
			return ok
		}},
		{name: "cluster-manager", serverKind: kustolang.ServerKindClusterManager, isType: func(grammar kustolang.ServerCommandGrammar) bool {
			_, ok := grammar.(*kustolang.ClusterManagerCommandGrammar)
			return ok
		}},
		{name: "aria-bridge", serverKind: kustolang.ServerKindAriaBridge, isType: func(grammar kustolang.ServerCommandGrammar) bool {
			_, ok := grammar.(*kustolang.AriaBridgeCommandGrammar)
			return ok
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals := kustolang.NewGlobalState(nil).WithServerKind(test.serverKind)
			grammar := kustolang.CreateServerCommandGrammar(globals)
			if !test.isType(grammar) || grammar.BaseCommandGrammar().Globals != globals {
				t.Fatalf("grammar = %T, globals = %#v", grammar, grammar.BaseCommandGrammar().Globals)
			}
			valid := grammar.BaseCommandGrammar().ParseCommandBlock(".show version")
			command := publicFirstCommand(t, valid)
			custom, ok := command.(*kustolang.CustomCommand)
			if !ok || custom.CommandKind() != "ShowVersion" {
				t.Fatalf("show-version command = %#v", command)
			}
			if diagnostics := valid.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid diagnostics = %#v", diagnostics)
			}
			if diagnostics := grammar.BaseCommandGrammar().ParseCommandBlock(".show").ContainedDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("incomplete command produced no diagnostics")
			} else if test.serverKind != kustolang.ServerKindEngine && (len(diagnostics) != 1 || diagnostics[0].Code != "KS005" || diagnostics[0].Message != "Expected: version") {
				t.Fatalf("incomplete show-version diagnostics = %#v", diagnostics)
			}
			if diagnostics := grammar.BaseCommandGrammar().ParseCommandBlock(".").ContainedDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS300" || diagnostics[0].Message != "Missing command." {
				t.Fatalf("bad-command diagnostics = %#v", diagnostics)
			}
			fresh := kustolang.CreateServerCommandGrammar(globals)
			if fresh == grammar || fresh.BaseCommandGrammar() == grammar.BaseCommandGrammar() {
				t.Fatal("create factory reused grammar identity")
			}
		})
	}
}

func TestPublicServerCommandGrammarConstructorsAndCache(t *testing.T) {
	engine := kustolang.NewEngineCommandGrammar(kustolang.DefaultGlobalState.WithServerKind(kustolang.ServerKindDataManager))
	if command, ok := publicFirstCommand(t, engine.ParseCommandBlock(".show queries")).(*kustolang.CustomCommand); !ok || command.CommandKind() != "ShowQueries" {
		t.Fatalf("engine constructor command = %#v", command)
	}

	data := kustolang.NewDataManagerCommandGrammar(kustolang.DefaultGlobalState)
	if command := publicFirstCommand(t, data.ParseCommandBlock(".show queries")); command.Kind() == kustolang.SyntaxKindCustomCommand {
		if custom, ok := command.(*kustolang.CustomCommand); ok && custom.CommandKind() == "ShowQueries" {
			t.Fatalf("data-manager constructor accepted engine command: %#v", command)
		}
	}

	globals := kustolang.NewGlobalState(nil).WithServerKind(kustolang.ServerKindAriaBridge).WithCache()
	first := kustolang.ServerCommandGrammarFrom(globals)
	second := kustolang.ServerCommandGrammarFrom(globals)
	if first != second || first.BaseCommandGrammar() != kustolang.CommandGrammarFrom(globals) {
		t.Fatal("concrete grammar cache did not preserve identity")
	}
	if _, ok := first.(*kustolang.AriaBridgeCommandGrammar); !ok {
		t.Fatalf("cached grammar type = %T", first)
	}
}

func TestPublicUnknownServerCommandGrammarRecoversUnknownCommand(t *testing.T) {
	globals := kustolang.NewGlobalState(nil).WithServerKind("CustomServer")
	grammar := kustolang.CreateServerCommandGrammar(globals)
	if _, ok := grammar.(*kustolang.CommandGrammar); !ok {
		t.Fatalf("unknown server grammar type = %T", grammar)
	}
	if command := publicFirstCommand(t, grammar.BaseCommandGrammar().ParseCommandBlock(".show version")); command.Kind() == kustolang.SyntaxKindCustomCommand {
		t.Fatalf("unknown server accepted known command: %#v", command)
	}
}

func publicFirstCommand(t *testing.T, block *kustolang.CommandBlock) kustolang.Expression {
	t.Helper()
	if block == nil || block.Statements() == nil || block.Statements().Count() == 0 {
		t.Fatalf("command block = %#v", block)
	}
	statement, ok := block.Statements().At(0).Element().(*kustolang.ExpressionStatement)
	if !ok {
		t.Fatalf("command statement = %#v", block.Statements().At(0).Element())
	}
	return statement.Expression()
}
