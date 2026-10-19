package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCommandGrammarAuthoringFacade(t *testing.T) {
	shape := []kustolang.CustomElementDescriptor{
		kustolang.CommandGrammarCD("Verb"),
		kustolang.CommandGrammarCD("Target", kustolang.CommandGrammarDescriptorOptions{
			Hint:       kustolang.CompletionHintTable,
			IsOptional: true,
		}),
	}
	content := kustolang.CommandGrammarCustomWithShape(
		shape,
		kustolang.CommandGrammarToken("show"),
		kustolang.CommandGrammarRequiredToken("widgets"),
	)
	parser := kustolang.CommandGrammarCommand("ShowWidgets", content)
	result := parser.ParseValue(kustolang.NewArraySource(kustolang.Lex(".show")), 0)
	command, ok := result.Value.(*kustolang.CustomCommand)
	if result.Length != 2 || !ok || command.CommandKind() != "ShowWidgets" {
		t.Fatalf("command = %#v, length = %d", result.Value, result.Length)
	}
	custom, ok := command.Custom().(*kustolang.CustomNode)
	if !ok || custom.GetName(1) != "Target" || custom.CompletionHint(1) != kustolang.CompletionHintTable {
		t.Fatalf("custom = %#v", command.Custom())
	}
	missing, ok := custom.GetChild(1).(*kustolang.SyntaxToken)
	if !ok || !missing.IsMissing() || missing.Kind() != kustolang.SyntaxKindIdentifierToken {
		t.Fatalf("missing target = %#v", custom.GetChild(1))
	}
}

func TestPublicCommandGrammarOverloadSpellings(t *testing.T) {
	descriptor := kustolang.CommandGrammarCDWithoutName(kustolang.CommandGrammarDescriptorOptions{
		Hint:       kustolang.CompletionHintColumn,
		IsOptional: true,
	})
	if descriptor.Name != "" || descriptor.Hint != kustolang.CompletionHintColumn || !descriptor.Optional {
		t.Fatalf("unnamed descriptor = %#v", descriptor)
	}
	missing := kustolang.CommandGrammarCreateMissingTokenFromList([]string{"alpha", "beta"})
	if missing == nil || missing.Width() != 0 || len(missing.Diagnostics()) == 0 {
		t.Fatalf("missing token from list = %#v", missing)
	}
}

func TestPublicCommandGrammarListFacade(t *testing.T) {
	element := kustolang.CommandGrammarTokens("a", "b")
	missing := func(kustolang.Source[kustolang.LexicalToken], int) kustolang.SyntaxElement {
		return kustolang.CreateMissingSyntaxToken(kustolang.SyntaxKindIdentifierToken)
	}
	result := kustolang.CommandGrammarOneOrMoreCommaList(element, missing, false).
		ParseValue(kustolang.NewArraySource(kustolang.Lex("a,,b")), 0)
	list, ok := result.Value.(*kustolang.SyntaxList)
	if result.Length != 4 || !ok || list.Count() != 3 {
		t.Fatalf("list = %#v, length = %d", result.Value, result.Length)
	}
	middle, ok := list.At(1).(*kustolang.SeparatedElement)
	if !ok || middle.Element() == nil || !middle.Element().IsMissing() {
		t.Fatalf("middle = %#v", list.At(1))
	}
}
