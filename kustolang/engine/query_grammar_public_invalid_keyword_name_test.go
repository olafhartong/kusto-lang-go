package kustolang

import "testing"

func TestQueryGrammarPublicInvalidKeywordAsNameRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	t.Run("binary expression", func(t *testing.T) {
		result := grammar.UnnamedExpression.ParseValue(NewArraySource(Lex("and == 1")), 0)
		if result.Length != 3 {
			t.Fatalf("invalid-keyword binary parse = %#v, want length 3", result)
		}
		binary, ok := result.Value.(*BinaryExpression)
		if !ok {
			t.Fatalf("invalid-keyword binary expression = %T, want *BinaryExpression", result.Value)
		}
		assertQueryGrammarInvalidKeywordName(t, binary.Left(), "and")
	})

	t.Run("pipeline source", func(t *testing.T) {
		result := grammar.Expression.ParseValue(NewArraySource(Lex("and | take 1")), 0)
		if result.Length != 4 {
			t.Fatalf("invalid-keyword pipeline parse = %#v, want length 4", result)
		}
		pipe, ok := result.Value.(*PipeExpression)
		if !ok {
			t.Fatalf("invalid-keyword pipeline expression = %T, want *PipeExpression", result.Value)
		}
		assertQueryGrammarInvalidKeywordName(t, pipe.Expression(), "and")
	})

	t.Run("colon", func(t *testing.T) {
		result := grammar.UnnamedExpression.ParseValue(NewArraySource(Lex("where:string")), 0)
		if result.Length != 2 {
			t.Fatalf("invalid-keyword colon parse = %#v, want length 2", result)
		}
		binary, ok := result.Value.(*BinaryExpression)
		if !ok {
			t.Fatalf("invalid-keyword colon expression = %T, want *BinaryExpression", result.Value)
		}
		assertQueryGrammarInvalidKeywordName(t, binary.Left(), "where")
	})

	for _, test := range []struct {
		name string
		text string
		want string
	}{
		{name: "comma", text: "project, Account", want: "project"},
		{name: "close paren", text: "and) trailing", want: "and"},
	} {
		t.Run(test.name, func(t *testing.T) {
			tokens := Lex(test.text)
			result := grammar.UnnamedExpression.ParseValue(NewArraySource(tokens), 0)
			if result.Length != 1 {
				t.Fatalf("invalid-keyword prefix parse %q = %#v, want length 1", test.text, result)
			}
			if result.Length == len(tokens)-1 {
				t.Fatalf("invalid-keyword prefix unexpectedly consumed all of %q", test.text)
			}
			assertQueryGrammarInvalidKeywordName(t, result.Value, test.want)
		})
	}
}

func TestQueryGrammarPublicInvalidKeywordAsNameRejectsMalformedAndAmbiguousTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		"and",
		"and(Account)",
		"and[Account]",
		"and + 1",
		"and - 1",
	} {
		t.Run(text, func(t *testing.T) {
			result := grammar.UnnamedExpression.ParseValue(NewArraySource(Lex(text)), 0)
			if result.Length >= 0 || result.Value != nil {
				t.Fatalf("ambiguous invalid-keyword name %q parsed: %#v", text, result)
			}
		})
	}
}

func assertQueryGrammarInvalidKeywordName(t *testing.T, expression Expression, keyword string) {
	t.Helper()
	reference, ok := expression.(*NameReference)
	if !ok {
		t.Fatalf("invalid keyword expression = %T, want *NameReference", expression)
	}
	if got := NameReferenceSimpleName(reference); got != keyword {
		t.Fatalf("invalid keyword name = %q, want %q", got, keyword)
	}
	wantMessage := "The name '" + keyword + "' needs to be bracketed as ['" + keyword + "'] to be used in this context."
	if !diagnosticsContainMessage(reference.Diagnostics(), "KS228", wantMessage) {
		t.Fatalf("invalid keyword %q diagnostics = %#v, want KS228 %q", keyword, reference.Diagnostics(), wantMessage)
	}
}
