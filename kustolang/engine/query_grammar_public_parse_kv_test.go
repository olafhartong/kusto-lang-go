package kustolang

import (
	"strings"
	"testing"
)

func TestQueryGrammarPublicParseKvFullASTAndProperties(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).parseKvOperator,
		`parse-kv Message as (Key:string, Value:long) with (pair_delimiter=";", kv_delimiter="=", regex=".*", quote="q1", quote="q2", escape="e1", escape="e2", greedy=true)`).(*ParseKvOperator)

	if operator.ParseKvKeyword().Kind() != SyntaxKindParseKvKeyword || operator.AsKeyword().Kind() != SyntaxKindAsKeyword {
		t.Fatalf("parse-kv fixed tokens = %v/%v", operator.ParseKvKeyword().Kind(), operator.AsKeyword().Kind())
	}
	if source, ok := operator.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Message" {
		t.Fatalf("parse-kv source = %#v, want Message", operator.Expression())
	}
	if operator.Keys().OpenParen().Kind() != SyntaxKindOpenParenToken || operator.Keys().CloseParen().Kind() != SyntaxKindCloseParenToken || operator.Keys().Columns().Count() != 2 {
		t.Fatalf("parse-kv schema = %#v", operator.Keys())
	}
	assertQueryGrammarRowSchemaColumn(t, operator.Keys().Columns().At(0).Element(), "Key", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, operator.Keys().Columns().At(1).Element(), "Value", SyntaxKindLongKeyword)

	with := operator.WithClause()
	if with == nil || with.WithKeyword().Kind() != SyntaxKindWithKeyword || with.OpenParen().Kind() != SyntaxKindOpenParenToken || with.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("parse-kv with clause = %#v", with)
	}
	wantNames := []string{"pair_delimiter", "kv_delimiter", "regex", "quote", "quote", "escape", "escape", "greedy"}
	if with.Properties().Count() != len(wantNames) {
		t.Fatalf("parse-kv property count = %d, want %d", with.Properties().Count(), len(wantNames))
	}
	for i, want := range wantNames {
		property := with.Properties().At(i).Element()
		if NameDeclarationSimpleName(property.Name()) != want || property.EqualToken().Kind() != SyntaxKindEqualToken || property.EqualToken().Width() == 0 {
			t.Fatalf("parse-kv property %d = %#v, want %q with equals", i, property, want)
		}
		wantKind := SyntaxKindStringLiteralExpression
		if want == "greedy" {
			wantKind = SyntaxKindBooleanLiteralExpression
		}
		if property.Expression().Kind() != wantKind {
			t.Fatalf("parse-kv property %q value kind = %v, want %v", want, property.Expression().Kind(), wantKind)
		}
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid parse-kv diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicParseKvEmptyWithAndKnownProperties(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	empty := parseQueryGrammarProduction(t, grammar.parseKvOperator, `parse-kv Message as (Key:string) with ()`).(*ParseKvOperator)
	if empty.WithClause() == nil || empty.WithClause().Properties().Count() != 0 || len(empty.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty parse-kv with clause = %#v diagnostics:%#v", empty.WithClause(), empty.ContainedDiagnostics())
	}

	known := parseQueryGrammarProduction(t, grammar.parseKvOperator,
		`parse-kv Message as (Key:string) with (hint.materialized=true)`).(*ParseKvOperator)
	property := known.WithClause().Properties().At(0).Element()
	if NameDeclarationSimpleName(property.Name()) != "hint.materialized" || property.Expression().Kind() != SyntaxKindBooleanLiteralExpression || len(known.ContainedDiagnostics()) != 0 {
		t.Fatalf("known parse-kv property = %#v diagnostics:%#v", property, known.ContainedDiagnostics())
	}

	text := `parse-kv Message as (Key:string) with (not_declared=true)`
	result := grammar.parseKvOperator.ParseValue(NewArraySource(Lex(text)), 0)
	operator, ok := result.Value.(*ParseKvOperator)
	if result.Length < 0 || !ok || operator.WithClause() == nil || operator.WithClause().Properties().Count() != 0 || result.Length == len(Lex(text))-1 {
		t.Fatalf("unknown parse-kv property boundary = %#v", result)
	}
}

func TestQueryGrammarPublicParseKvCompletionApplyTextAndTag(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	items := queryGrammarParseKvCompletionItems(grammar.QueryOperator, "")
	item, ok := completionItemByDisplayText(items, "parse-kv")
	if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityLow || item.BeforeText() != "parse-kv" {
		t.Fatalf("parse-kv completion = %#v, visible:%#v", item, completionDisplayTexts(items))
	}

	items = queryGrammarParseKvCompletionItems(grammar.parseKvOperator, `parse-kv Message as (Key:string) with (`)
	for _, name := range []string{"pair_delimiter", "kv_delimiter", "regex", "quote", "escape"} {
		item, ok := completionItemByDisplayText(items, name)
		if !ok || item.BeforeText() != name+`="` || item.AfterText() != `"` || !item.HasCaret() {
			t.Fatalf("parse-kv property completion %q = %#v, visible:%#v", name, item, completionDisplayTexts(items))
		}
	}
	greedy, ok := completionItemByDisplayText(items, "greedy")
	if !ok || greedy.BeforeText() != "greedy=" || greedy.AfterText() != "" || !greedy.HasCaret() {
		t.Fatalf("parse-kv greedy completion = %#v, visible:%#v", greedy, completionDisplayTexts(items))
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<parse-kv>", 0) {
		t.Fatal("QueryOperator parser tree is missing <parse-kv> tag")
	}
}

func TestQueryGrammarPublicParseKvPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := `parse-kv Message as (Key:string)`
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private":        grammar.parseKvOperator,
		"query operator": grammar.QueryOperator,
		"following pipe": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, text).(*ParseKvOperator); !ok {
				t.Fatal("parse-kv placement produced unexpected node")
			}
		})
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression, `T | parse-kv Message as (Key:string)`)
	pipe, node := queryGrammarPipeStage(t, pipeline, SyntaxKindParseKvOperator)
	if source, ok := pipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("parse-kv pipeline source = %#v, want T", pipe.Expression())
	}
	if _, ok := node.(*ParseKvOperator); !ok {
		t.Fatalf("parse-kv pipeline node = %T", node)
	}

	fork := parseQueryGrammarProduction(t, grammar.QueryOperator,
		`fork (parse-kv Message as (Key:string))`).(*ForkOperator)
	if fork.Expressions().Count() != 1 || fork.Expressions().At(0).Expression().Kind() != SyntaxKindParseKvOperator || len(fork.ContainedDiagnostics()) != 0 {
		t.Fatalf("fork parse-kv expression = %#v diagnostics:%#v", fork.Expressions(), fork.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicParseKvRequiredRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	missingSource := parseQueryGrammarProduction(t, grammar.parseKvOperator, `parse-kv as (Key:string)`).(*ParseKvOperator)
	assertQueryGrammarParseKvMissing(t, missingSource.Expression(), "Missing expression")

	missingAs := parseQueryGrammarProduction(t, grammar.parseKvOperator, `parse-kv Message (Key:string)`).(*ParseKvOperator)
	assertQueryGrammarParseKvMissing(t, missingAs.AsKeyword(), "Expected: as")

	missingSchema := parseQueryGrammarProduction(t, grammar.parseKvOperator, `parse-kv Message as`).(*ParseKvOperator).Keys()
	if missingSchema.Width() != 0 || missingSchema.Columns().Count() != 0 || !diagnosticsContain(missingSchema.ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing parse-kv schema = %#v diagnostics:%#v", missingSchema, missingSchema.ContainedDiagnostics())
	}

	missingOpen := parseQueryGrammarProduction(t, grammar.parseKvOperator,
		`parse-kv Message as (Key:string) with pair_delimiter=";", greedy=true)`).(*ParseKvOperator).WithClause()
	assertQueryGrammarParseKvMissing(t, missingOpen.OpenParen(), "Expected: (")

	missingValue := parseQueryGrammarProduction(t, grammar.parseKvOperator,
		`parse-kv Message as (Key:string) with (pair_delimiter=)`).(*ParseKvOperator).WithClause().Properties().At(0).Element().Expression()
	assertQueryGrammarParseKvMissing(t, missingValue, "Missing string")

	missingComma := parseQueryGrammarProduction(t, grammar.parseKvOperator,
		`parse-kv Message as (Key:string) with (pair_delimiter=";" greedy=true)`).(*ParseKvOperator).WithClause()
	if missingComma.Properties().Count() != 2 {
		t.Fatalf("missing-comma parse-kv property count = %d, want 2", missingComma.Properties().Count())
	}
	separator := missingComma.Properties().At(0).Separator()
	assertQueryGrammarParseKvMissing(t, separator, "Expected: ,")

	missingClose := parseQueryGrammarProduction(t, grammar.parseKvOperator,
		`parse-kv Message as (Key:string) with (pair_delimiter=";"`).(*ParseKvOperator).WithClause()
	assertQueryGrammarParseKvMissing(t, missingClose.CloseParen(), "Expected: )")
}

func TestQueryGrammarPublicParseKvTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		`parse-kv Message as (Key:string) Extra`,
		`parse-kv Message as (Key:string) with (pair_delimiter=";") Extra`,
		`parse-kv Message as (Key:string) | where true`,
	} {
		tokens := Lex(text)
		result := grammar.parseKvOperator.ParseValue(NewArraySource(tokens), 0)
		if result.Length < 0 || result.Length == len(tokens)-1 {
			t.Fatalf("parse-kv tail boundary %q = %#v", text, result)
		}
	}
}

func queryGrammarParseKvCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}

func assertQueryGrammarParseKvMissing(t *testing.T, element SyntaxElement, message string) {
	t.Helper()
	found := false
	if element != nil {
		for _, diagnostic := range element.ContainedDiagnostics() {
			if strings.Contains(diagnostic.Message, message) {
				found = true
				break
			}
		}
	}
	if element == nil || element.Width() != 0 || !found {
		t.Fatalf("missing parse-kv element = %#v diagnostics:%#v, want %q", element, element.ContainedDiagnostics(), message)
	}
}
