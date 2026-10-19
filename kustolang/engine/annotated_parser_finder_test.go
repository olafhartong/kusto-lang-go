package kustolang

import "testing"

func TestFindAnnotatedParsersFindsParserAtOffset(t *testing.T) {
	keyword := MatchText("where").WithTag("where").WithAnnotations([]any{"keyword"})
	root := Sequence(MatchText("T"), MatchText(" "), keyword)

	results := FindAnnotatedParsers(NewTextSource("T "), 2, root, 0)
	if len(results) != 1 || results[0] != keyword {
		t.Fatalf("results = %#v, want annotated where parser", results)
	}
}

func TestFindAnnotatedParserPathsIncludeParentChain(t *testing.T) {
	left := MatchText("A").WithTag("left")
	right := MatchText("B").WithTag("right").WithAnnotations([]any{"right"})
	root := Sequence(left, right).WithTag("root")

	paths := FindAnnotatedParserPaths(NewTextSource("A"), 1, root, 0)
	if len(paths) != 1 {
		t.Fatalf("path count = %d, want 1", len(paths))
	}
	leaf := paths[0]
	if leaf.Parser() != right || leaf.IndexInParent() != 1 {
		t.Fatalf("leaf = parser:%#v index:%d", leaf.Parser(), leaf.IndexInParent())
	}
	if leaf.Parent() == nil || leaf.Parent().Parser() != root || leaf.Parent().IndexInParent() != 0 {
		t.Fatalf("parent chain = %#v", leaf.Parent())
	}
}

func TestFindAnnotatedParsersIgnoresHiddenAnnotatedParser(t *testing.T) {
	hidden := MatchText("A").WithAnnotations([]any{"hidden"}).Hide()
	results := FindAnnotatedParsers(NewTextSource("A"), 0, hidden, 0)
	if len(results) != 0 {
		t.Fatalf("hidden annotated results = %#v, want none", results)
	}
}

func TestFindAnnotatedParsersPrefersFewerMissingRequiredParsers(t *testing.T) {
	afterMissing := zeroLengthByteParser().WithTag("after-missing").WithAnnotations([]any{"missing"})
	direct := zeroLengthByteParser().WithTag("direct").WithAnnotations([]any{"direct"})
	withMissing := Sequence(
		RequiredValue(Text(MatchText("A")), func() string { return "" }),
		zeroLengthByteParser(),
		afterMissing,
	)
	root := First(withMissing, direct)

	results := FindAnnotatedParsers(NewTextSource(""), 0, root, 0)
	if len(results) != 1 || results[0] != direct {
		t.Fatalf("results = %#v, want only fewer-missing direct parser", results)
	}
}

func TestFindAnnotatedParsersTraversesLimitAndRepetition(t *testing.T) {
	item := MatchText("A").WithAnnotations([]any{"item"})
	repeated := ZeroOrMore(item)
	root := Limit(MatchText("AA"), Text(repeated))

	results := FindAnnotatedParsers(NewTextSource("AA"), 1, root, 0)
	if len(results) != 1 || results[0] != item {
		t.Fatalf("repetition results = %#v, want repeated item", results)
	}
}

func zeroLengthByteParser() Parser[byte] {
	return MatchConsumer(func(Source[byte], int) int { return 0 })
}
