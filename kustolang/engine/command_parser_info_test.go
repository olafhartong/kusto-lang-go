package kustolang

import (
	"strings"
	"testing"
)

func TestCommandParserInfoFactoryParsesAndRecoversCommands(t *testing.T) {
	var receivedRules *PredefinedRuleParsers
	grammar := NewCommandGrammarWithCommandParserFactory(DefaultGlobalState, func(rules *PredefinedRuleParsers) []CommandParserInfo {
		receivedRules = rules
		content := CommandGrammarCustomElements(
			CommandGrammarToken("echo"),
			CommandGrammarRequiredToken("value"),
		)
		return []CommandParserInfo{
			NewCommandParserInfo("EchoValue", CommandGrammarCommand("EchoValue", content)),
		}
	})
	if receivedRules == nil || receivedRules != grammar.PredefinedRules {
		t.Fatal("factory did not receive the grammar's predefined rules")
	}

	command := firstCommandExpression(t, grammar.ParseCommandBlock(".echo value"))
	custom, ok := command.(*CustomCommand)
	if !ok || custom.CommandKind() != "EchoValue" {
		t.Fatalf("valid command = %#v", command)
	}
	if diagnostics := custom.GetContainedSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid command diagnostics = %#v", diagnostics)
	}

	recovered := firstCommandExpression(t, grammar.ParseCommandBlock(".echo"))
	custom, ok = recovered.(*CustomCommand)
	if !ok || custom.CommandKind() != "EchoValue" {
		t.Fatalf("recovered command = %#v", recovered)
	}
	if diagnostics := custom.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("recovered command has no missing-token diagnostic")
	}
}

func TestCommandParserInfoUsesSuppliedParserForPartialRecovery(t *testing.T) {
	content := CommandGrammarCustomElements(
		CommandGrammarToken("deploy"),
		CommandGrammarToken("database"),
	)
	grammar := NewCommandGrammarWithCommandParsers(
		DefaultGlobalState,
		NewCommandParserInfo("DeployDatabase", CommandGrammarCommand("DeployDatabase", content)),
	)

	command := firstCommandExpression(t, grammar.ParseCommandBlock(".deploy"))
	partial, ok := command.(*PartialCommand)
	if !ok {
		t.Fatalf("partial command = %T, want *PartialCommand", command)
	}
	if kinds := partial.CommandKinds(); len(kinds) != 1 || kinds[0] != "DeployDatabase" {
		t.Fatalf("partial command kinds = %#v", kinds)
	}
	if diagnostics := partial.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("partial command has no expected-term diagnostic")
	}
}

func TestCommandParserInfoGrammarRetainsFallbacks(t *testing.T) {
	grammar := NewCommandGrammarWithCommandParserProvider(DefaultGlobalState, nil)
	if command := firstCommandExpression(t, grammar.ParseCommandBlock(".unknown value")); command.Kind() != SyntaxKindUnknownCommand {
		t.Fatalf("unknown fallback kind = %v", command.Kind())
	}
	if command := firstCommandExpression(t, grammar.ParseCommandBlock(".")); command.Kind() != SyntaxKindBadCommand {
		t.Fatalf("bad fallback kind = %v", command.Kind())
	}
}

func TestCommandParserInfoParserTreesDriveTiedPartialRecovery(t *testing.T) {
	database := CommandGrammarCustomElements(
		CommandGrammarToken("deploy"),
		CommandGrammarToken("database"),
	)
	cluster := CommandGrammarCustomElements(
		CommandGrammarToken("deploy"),
		CommandGrammarToken("cluster"),
	)
	grammar := NewCommandGrammarWithCommandParsers(
		DefaultGlobalState,
		NewCommandParserInfo("DeployDatabase", CommandGrammarCommand("DeployDatabase", database)),
		NewCommandParserInfo("DeployCluster", CommandGrammarCommand("DeployCluster", cluster)),
	)
	for _, info := range grammar.commandParsers {
		if len(info.Elements) != 0 || len(info.Tokens) != 0 {
			t.Fatalf("custom parser unexpectedly depended on generated descriptors: %#v", info)
		}
	}

	command := firstCommandExpression(t, grammar.ParseCommandBlock(".deploy"))
	partial, ok := command.(*PartialCommand)
	if !ok {
		t.Fatalf("partial command = %T, want *PartialCommand", command)
	}
	wantKinds := []string{"DeployDatabase", "DeployCluster"}
	if kinds := partial.CommandKinds(); len(kinds) != len(wantKinds) || kinds[0] != wantKinds[0] || kinds[1] != wantKinds[1] {
		t.Fatalf("partial command kinds = %#v, want %#v", kinds, wantKinds)
	}
	diagnostics := partial.GetContainedSyntaxDiagnostics()
	if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Message, "cluster") || !strings.Contains(diagnostics[0].Message, "database") {
		t.Fatalf("partial expected-term diagnostics = %#v", diagnostics)
	}
}

func TestCommandParserInfoParserTreesDriveCompletionAnnotations(t *testing.T) {
	database := CommandGrammarCustomElements(
		CommandGrammarToken("deploy"),
		CommandGrammarToken("database"),
	)
	cluster := CommandGrammarCustomElements(
		CommandGrammarToken("deploy"),
		CommandGrammarToken("cluster"),
	)
	grammar := NewCommandGrammarWithCommandParsers(
		DefaultGlobalState,
		NewCommandParserInfo("DeployDatabase", CommandGrammarCommand("DeployDatabase", database)),
		NewCommandParserInfo("DeployCluster", CommandGrammarCommand("DeployCluster", cluster)),
	)

	code := Parse(".deploy ")
	parsers := FindAnnotatedParsers(
		NewArraySource(code.Tokens),
		code.GetTokenIndex(len(".deploy ")),
		grammar.CommandCompletionParser(),
		0,
	)
	items := commandCompletionItemsFromParsers(parsers)
	for _, display := range []string{"cluster", "database"} {
		if _, ok := completionItemByDisplayText(items, display); !ok {
			t.Fatalf("custom parser annotations missing %q: %#v", display, completionDisplayTexts(items))
		}
	}
}
