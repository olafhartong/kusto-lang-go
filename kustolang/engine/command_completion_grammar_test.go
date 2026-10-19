package kustolang

import "testing"

func TestCommandCompletionParserFindsCommandPrefixAnnotations(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	parser := grammar.CommandCompletionParser()
	if parser == nil {
		t.Fatal("command completion parser is nil")
	}

	code := Parse(".show ")
	parsers := FindAnnotatedParsers(NewArraySource(code.Tokens), code.GetTokenIndex(len(".show ")), parser, 0)
	items := commandCompletionItemsFromParsers(parsers)
	if _, ok := completionItemByDisplayText(items, "database"); !ok {
		t.Fatalf("command parser annotations missing database: %#v", completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "version"); !ok {
		t.Fatalf("command parser annotations missing version: %#v", completionDisplayTexts(items))
	}
}

func TestCommandCompletionParserFindsPartialFirstTokenAnnotations(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	code := Parse(".sh")
	parsers := FindAnnotatedParsers(NewArraySource(code.Tokens), code.GetTokenIndex(len(".")), grammar.CommandCompletionParser(), 0)
	items := commandCompletionItemsFromParsers(parsers)
	if _, ok := completionItemByDisplayText(items, "show"); !ok {
		t.Fatalf("command parser annotations missing show: %#v", completionDisplayTexts(items))
	}
}

func TestCommandCompletionParserPreservesRuleHints(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	code := Parse(".show table ")
	parsers := FindAnnotatedParsers(NewArraySource(code.Tokens), code.GetTokenIndex(len(".show table ")), grammar.CommandCompletionParser(), 0)
	for _, parser := range parsers {
		if syntaxCompletionHintFromAnnotations(parser.Annotations())&CompletionHintTable != 0 {
			return
		}
	}
	t.Fatalf("command parser annotations missing table hint at .show table")
}

func TestCommandCompletionParserPreservesPredefinedRuleIdentity(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	code := Parse(".create table ")
	parsers := FindAnnotatedParsers(NewArraySource(code.Tokens), code.GetTokenIndex(len(".create table ")), grammar.CommandCompletionParser(), 0)
	for _, parser := range parsers {
		if commandCompletionRuleFromAnnotations(parser.Annotations()) == "NameDeclaration" {
			return
		}
	}
	t.Fatal("command parser annotations missing NameDeclaration rule identity")
}

func TestCommandCompletionParserSuppressesHiddenCommandItems(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	code := Parse(".create external table Ext (Raw:string) kind = ")
	parsers := FindAnnotatedParsers(NewArraySource(code.Tokens), code.GetTokenIndex(len(".create external table Ext (Raw:string) kind = ")), grammar.CommandCompletionParser(), 0)
	items := commandCompletionItemsFromParsers(parsers)
	if _, ok := completionItemByDisplayText(items, "storage"); !ok {
		t.Fatalf("command parser annotations missing storage: %#v", completionDisplayTexts(items))
	}
	for _, hidden := range []string{"blob", "adl"} {
		if _, ok := completionItemByDisplayText(items, hidden); ok {
			t.Fatalf("command parser annotations unexpectedly included hidden %q: %#v", hidden, completionDisplayTexts(items))
		}
	}
}

func commandCompletionItemsFromParsers(parsers []Parser[LexicalToken]) []CompletionItem {
	items := []CompletionItem{}
	for _, parser := range parsers {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	return items
}
