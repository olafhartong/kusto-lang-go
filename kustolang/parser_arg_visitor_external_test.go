package kustolang_test

import (
	"fmt"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicParserArgumentVisitorDispatch(t *testing.T) {
	a := kustolang.NewPredicateMatchParser(func(input byte) bool { return input == 'A' })
	b := kustolang.NewPredicateMatchParser(func(input byte) bool { return input == 'B' })
	sequence := kustolang.NewSequenceParser([]kustolang.Parser[byte]{a, b})
	visitor := kustolang.ParserVisitorArgResultFuncs[byte, string, string]{
		Sequence: func(_ kustolang.Parser[byte], children []kustolang.Parser[byte], prefix string) string {
			return fmt.Sprintf("%s:%d", prefix, len(children))
		},
		Unsupported: func(parser kustolang.Parser[byte], prefix string) string {
			return fmt.Sprintf("%s:%v", prefix, parser == nil)
		},
	}

	if got := kustolang.AcceptParserVisitorArgResult(sequence, visitor, "children"); got != "children:2" {
		t.Fatalf("sequence argument visitor = %q", got)
	}
	if got := kustolang.AcceptParserVisitorArgResult[byte](nil, visitor, "nil"); got != "nil:true" {
		t.Fatalf("nil argument visitor = %q", got)
	}
}

func TestPublicParserArgumentVisitorApplyAndMalformedInput(t *testing.T) {
	a := publicByteValue('A', "A")
	b := publicByteValue('B', "B")
	apply := kustolang.Apply(a, func(kustolang.LeftValue[string]) kustolang.RightParser[byte, string] {
		return kustolang.RuleRight1(b, func(left, right string) string { return left + right })
	})
	visitor := kustolang.ParserVisitorArgResultFuncs[byte, int, bool]{
		Apply: func(_ kustolang.Parser[byte], left, right kustolang.Parser[byte], kind kustolang.ApplyKind, expected int) bool {
			return left != nil && right != nil && kind == kustolang.ApplyKindOne && expected == 2
		},
	}
	if !kustolang.AcceptParserVisitorArgResult(apply, visitor, 2) {
		t.Fatal("apply argument visitor did not receive parser parts and argument")
	}

	output := []any{}
	if length := apply.ParseList(kustolang.NewTextSource("A"), 0, &output, 0); length >= 0 {
		t.Fatalf("malformed apply length = %d, want failure", length)
	}
	if !kustolang.AcceptParserVisitorArgResult(apply, visitor, 2) {
		t.Fatal("visitor dispatch changed after malformed parser input")
	}
}
