package kustolang

import "testing"

func TestParserPathExposesParserParentIndexAndChildren(t *testing.T) {
	left := MatchText("A").WithTag("left")
	right := MatchText("B").WithTag("right")
	rootParser := Sequence(left, right).WithTag("root")
	root := NewParserPath(NewParserAndIndex(rootParser, -1))

	if root.Parser() != rootParser {
		t.Fatalf("root parser = %#v, want original", root.Parser())
	}
	if root.Parent() != nil {
		t.Fatalf("root parent = %#v, want nil", root.Parent())
	}
	if root.IndexInParent() != -1 {
		t.Fatalf("root index = %d, want -1", root.IndexInParent())
	}
	if root.ChildCount() != 2 {
		t.Fatalf("child count = %d, want 2", root.ChildCount())
	}
	first := root.GetChild(0)
	if first == nil || first.Parser() != left || first.Parent() != root || first.IndexInParent() != 0 {
		t.Fatalf("first child = %#v", first)
	}
	if again := root.GetChild(0); again != first {
		t.Fatalf("child paths should be cached: first=%p again=%p", first, again)
	}
	second := root.GetChild(1)
	if second == nil || second.Parser() != right || second.Parent() != root || second.IndexInParent() != 1 {
		t.Fatalf("second child = %#v", second)
	}
	if root.GetChild(-1) != nil || root.GetChild(2) != nil {
		t.Fatal("out-of-range children should return nil")
	}
}

func TestParserPathFromBuildsParentChain(t *testing.T) {
	rootParser := Sequence(MatchText("A"), MatchText("B")).WithTag("root")
	childParser := rootParser.ChildParser(1).WithTag("child")
	path := ParserPathFrom([]ParserAndIndex[byte]{
		NewParserAndIndex(rootParser, -1),
		NewParserAndIndex(childParser, 1),
	})

	if path == nil || path.Parser() != childParser || path.IndexInParent() != 1 {
		t.Fatalf("path leaf = %#v", path)
	}
	if path.Parent() == nil || path.Parent().Parser() != rootParser || path.Parent().Parent() != nil {
		t.Fatalf("path parent chain = %#v", path.Parent())
	}
	if got := ParserPathFrom[byte](nil); got != nil {
		t.Fatalf("empty path = %#v, want nil", got)
	}
}

func TestParserPathDebugTextCombinesDescriptions(t *testing.T) {
	rootParser := MatchText("A").WithTag("root")
	childParser := MatchText("B").WithTag("child")
	path := ParserPathFrom([]ParserAndIndex[byte]{
		NewParserAndIndex(rootParser, -1),
		NewParserAndIndex(childParser, 0),
	})
	if got, want := path.DebugText(), "root: match()/child: match()"; got != want {
		t.Fatalf("debug text = %q, want %q", got, want)
	}
	var nilPath *ParserPath[byte]
	if nilPath.DebugText() != "" || nilPath.Parser() != nil || nilPath.Parent() != nil || nilPath.ChildCount() != 0 || nilPath.IndexInParent() != -1 {
		t.Fatal("nil parser path helpers should return zero values")
	}
}
