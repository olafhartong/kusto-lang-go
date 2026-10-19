package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type externalCommandParserProvider struct{}

func (externalCommandParserProvider) CreateCommandParsers(rules *kustolang.PredefinedRuleParsers) []kustolang.CommandParserInfo {
	content := kustolang.CommandGrammarCustomElements(
		kustolang.CommandGrammarToken("inspect"),
		kustolang.CommandGrammarRequiredToken("widget"),
		rules.TableNameReference,
	)
	return []kustolang.CommandParserInfo{
		kustolang.NewCommandParserInfo("InspectWidget", kustolang.CommandGrammarCommand("InspectWidget", content)),
	}
}

func TestPublicCommandParserInfoProviderIsExternallyImplementable(t *testing.T) {
	var _ kustolang.CommandParserInfoProvider = externalCommandParserProvider{}

	grammar := kustolang.NewCommandGrammarWithCommandParserProvider(nil, externalCommandParserProvider{})
	valid := firstPublicCommandExpression(t, grammar.ParseCommandBlock(".inspect widget W"))
	command, ok := valid.(*kustolang.CustomCommand)
	if !ok || command.CommandKind() != "InspectWidget" {
		t.Fatalf("valid command = %#v", valid)
	}

	recovered := firstPublicCommandExpression(t, grammar.ParseCommandBlock(".inspect W"))
	command, ok = recovered.(*kustolang.CustomCommand)
	if !ok || command.CommandKind() != "InspectWidget" {
		t.Fatalf("recovered command = %#v", recovered)
	}
	if diagnostics := command.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("recovered command has no missing-token diagnostic")
	}
}

func firstPublicCommandExpression(t *testing.T, block *kustolang.CommandBlock) kustolang.Expression {
	t.Helper()
	if block == nil || block.Statements() == nil || block.Statements().Count() == 0 {
		t.Fatalf("command block has no statements: %#v", block)
	}
	statement, ok := block.Statements().At(0).Element().(*kustolang.ExpressionStatement)
	if !ok || statement.Expression() == nil {
		t.Fatalf("first statement = %#v", block.Statements().At(0).Element())
	}
	return statement.Expression()
}
