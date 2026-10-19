package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type publicTestScope struct {
	symbols []kustolang.Symbol
}

func (s publicTestScope) GetSymbols(name string, match kustolang.SymbolMatch, symbols *[]kustolang.Symbol) {
	for _, symbol := range s.symbols {
		if (name == "" || symbol.SymbolName() == name) && kustolang.SymbolMatches(symbol, match) {
			*symbols = append(*symbols, symbol)
		}
	}
}

func TestPublicScopeAndEmptyScope(t *testing.T) {
	table := kustolang.NewTableSymbol("Events")
	function := kustolang.NewScalarFunctionSymbol("normalize", kustolang.ScalarTypes.String.Type())
	scope := publicTestScope{symbols: []kustolang.Symbol{table, function}}

	if got := kustolang.GetScopeSymbols(scope, "Events", kustolang.SymbolMatchTable); len(got) != 1 || got[0] != table {
		t.Fatalf("named scope symbols = %#v", got)
	}
	if got := kustolang.GetAllScopeSymbols(scope, kustolang.SymbolMatchFunction); len(got) != 1 || got[0] != function {
		t.Fatalf("all function symbols = %#v", got)
	}
	if got := kustolang.GetAllScopeSymbols(kustolang.EmptyScopeInstance, kustolang.SymbolMatchAny); len(got) != 0 {
		t.Fatalf("empty scope symbols = %#v", got)
	}
}

func TestPublicFakeExpressionsCarrySemanticsAndMissingSyntax(t *testing.T) {
	expression := kustolang.CreateFakeExpression(kustolang.ScalarTypes.String.Type())
	literal, ok := expression.(*kustolang.LiteralExpression)
	if !ok {
		t.Fatalf("fake expression = %T, want *LiteralExpression", expression)
	}
	if kustolang.ExpressionResultType(literal) != kustolang.ScalarTypes.String.Type() || !literal.IsBound() {
		t.Fatalf("fake expression semantics = %#v", literal)
	}
	if token := literal.Token(); token == nil || !token.IsMissing() || token.Kind() != kustolang.SyntaxKindIdentifierToken {
		t.Fatalf("fake expression token = %#v", token)
	}

	named := kustolang.CreateNamedFakeExpression("Answer", kustolang.ScalarTypes.Long.Type())
	namedExpression, ok := named.(*kustolang.SimpleNamedExpression)
	if !ok || kustolang.ExpressionResultType(named) != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("named fake expression = %T, type %v", named, kustolang.ExpressionResultType(named))
	}
	if namedExpression.Expression() == nil || !namedExpression.Expression().GetFirstToken(true).IsMissing() {
		t.Fatalf("named fake expression did not retain recovered missing syntax")
	}
}

func TestPublicSubstringMapUsesLongestUTF16Match(t *testing.T) {
	matches := kustolang.NewSubstringMap([]kustolang.SubstringMapEntry[int]{
		{Key: "where", Value: 1},
		{Key: "whereby", Value: 2},
		{Key: "😀where", Value: 3},
	})

	if got := matches.GetLongestMatch("😀whereby x", 0); got.Key != "😀where" || got.Value != 3 {
		t.Fatalf("emoji-prefix match = %#v", got)
	}
	if got := matches.GetLongestMatch("😀whereby x", 2); got.Key != "whereby" || got.Value != 2 {
		t.Fatalf("UTF-16 offset match = %#v", got)
	}
	if got := matches.GetLongestMatch("😀whereby x", 1); got.Key != "" || got.Value != 0 {
		t.Fatalf("split-surrogate match = %#v, want no match", got)
	}
	if got := matches.GetLongestMatch("whereby", 20); got.Key != "" || got.Value != 0 {
		t.Fatalf("past-end match = %#v, want no match", got)
	}
}
