package kustolang

import "testing"

func TestQueryGrammarPublicScanSimpleName(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name string
		text string
		want int
	}{
		{name: "identifier", text: "Events", want: 1},
		{name: "complete bracketed", text: "['Event Name']", want: 3},
		{name: "incomplete bracketed", text: "['Event Name'", want: 2},
		{name: "legal keyword", text: "access", want: 1},
		{name: "braced client parameter", text: "{tenant}", want: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := NewArraySource(Lex(test.text))
			if got := grammar.scanSimpleName.Scan(source, 0); got != test.want {
				t.Fatalf("ScanSimpleName(%q) = %d, want %d", test.text, got, test.want)
			}
		})
	}
}

func TestQueryGrammarPublicDispatchTags(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name   string
		parser Parser[LexicalToken]
		want   string
	}{
		{name: "expression", parser: grammar.Expression, want: "<expression>"},
		{name: "pipe expression", parser: grammar.PipeExpression, want: "<pipe-expression>"},
		{name: "pipe subexpression", parser: grammar.PipeSubExpression, want: "<pipe-sub-expression>"},
		{name: "query operator", parser: grammar.QueryOperator, want: ""},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression, want: ""},
		{name: "fork", parser: grammar.forkOperator, want: "<fork>"},
		{name: "fork pipe operator", parser: grammar.forkPipeOperator, want: "<fork-pipe-operator>"},
		{name: "fork pipe expression", parser: grammar.forkPipeExpression, want: "<fork-pipe-expression>"},
		{name: "contextual subexpression", parser: grammar.contextualSubExpression, want: "<contextual-sub-expression>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.parser.Tag(); got != test.want {
				t.Fatalf("parser tag = %q, want %q", got, test.want)
			}
		})
	}
	if grammar.Expression == grammar.PipeExpression {
		t.Fatal("Expression aliases PipeExpression, want a separately tagged forward")
	}
	if _, ok := any(grammar.Expression).(*ForwardParser[LexicalToken, Expression]); !ok {
		t.Fatalf("Expression parser = %T, want *ForwardParser", grammar.Expression)
	}
}

func TestQueryGrammarPublicDispatchCompletionVisibilityAndOrder(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	root := queryGrammarDispatchCompletionItems(grammar.Expression, "")
	for _, visible := range []string{"evaluate", "find", "search", "union", "range", "print"} {
		if _, ok := completionItemByDisplayText(root, visible); !ok {
			t.Fatalf("root completion missing %q: %#v", visible, completionDisplayTexts(root))
		}
	}
	for _, hidden := range []string{"where", "project", "top", "top-hitters", "top-nested", "graph-where-nodes", "graph-where-edges"} {
		if _, ok := completionItemByDisplayText(root, hidden); ok {
			t.Fatalf("root completion exposed %q: %#v", hidden, completionDisplayTexts(root))
		}
	}

	following := queryGrammarDispatchCompletionItems(grammar.FollowingPipeElementExpression, "")
	for _, visible := range []string{"where", "project", "top-hitters", "top", "top-nested"} {
		if _, ok := completionItemByDisplayText(following, visible); !ok {
			t.Fatalf("post-pipe completion missing %q: %#v", visible, completionDisplayTexts(following))
		}
	}
	for _, hidden := range []string{"find", "range", "print", "graph-where-nodes", "graph-where-edges"} {
		if _, ok := completionItemByDisplayText(following, hidden); ok {
			t.Fatalf("post-pipe completion exposed %q: %#v", hidden, completionDisplayTexts(following))
		}
	}

	displays := completionDisplayTexts(following)
	hitters := queryGrammarCompletionIndex(displays, "top-hitters")
	top := queryGrammarCompletionIndex(displays, "top")
	nested := queryGrammarCompletionIndex(displays, "top-nested")
	if hitters < 0 || top < 0 || nested < 0 || !(hitters < top && top < nested) {
		t.Fatalf("post-pipe top completion order = %#v, want top-hitters before top before top-nested", displays)
	}
}

func TestQueryGrammarPublicDispatchRecoveryAndPipeAssociativity(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parsed := parseQueryGrammarProduction(t, grammar.Expression, "Events | unknown | take 1")
	takePipe, take := queryGrammarPipeStage(t, parsed, SyntaxKindTakeOperator)
	badPipe, bad := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindBadQueryOperator)
	if _, ok := take.(*TakeOperator); !ok {
		t.Fatalf("final operator = %T, want *TakeOperator", take)
	}
	badOperator, ok := bad.(*BadQueryOperator)
	if !ok || !diagnosticsContainMessage(badOperator.Diagnostics(), "KS176", "Query operator expected.") {
		t.Fatalf("bad operator = %#v diagnostics:%#v", bad, bad.ContainedDiagnostics())
	}
	if source, ok := badPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Events" {
		t.Fatalf("left-associative pipe source = %#v, want Events", badPipe.Expression())
	}

	braced := parseQueryGrammarProduction(t, grammar.Expression, "{tenant} | count")
	countPipe, _ := queryGrammarPipeStage(t, braced, SyntaxKindCountOperator)
	name, ok := countPipe.Expression().(*NameReference)
	if !ok || name.Name().Kind() != SyntaxKindBracedName {
		t.Fatalf("braced pipeline source = %#v, want braced NameReference", countPipe.Expression())
	}
}

func queryGrammarDispatchCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}

func queryGrammarCompletionIndex(items []string, want string) int {
	for i, item := range items {
		if item == want {
			return i
		}
	}
	return -1
}
