package kustolang_test

// Tests for the public ParserPath port from Kusto.Language.Editor.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"sync"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicParserPathTraversalAndConstruction(t *testing.T) {
	left := kustolang.MatchText("A").WithTag("left")
	right := kustolang.MatchText("B").WithTag("right")
	rootParser := kustolang.Sequence(left, right).WithTag("root")
	root := kustolang.NewParserPath(kustolang.NewParserAndIndex(rootParser, -1))

	if root.Parser() != rootParser || root.Parent() != nil || root.IndexInParent() != -1 || root.ChildCount() != 2 {
		t.Fatalf("root path = parser:%T parent:%p index:%d children:%d", root.Parser(), root.Parent(), root.IndexInParent(), root.ChildCount())
	}
	leftPath := root.GetChild(0)
	if leftPath == nil || leftPath.Parser() != left || leftPath.Parent() != root || leftPath.IndexInParent() != 0 {
		t.Fatalf("left path = %#v", leftPath)
	}
	if root.GetChild(0) != leftPath {
		t.Fatal("GetChild did not preserve cached path identity")
	}
	if root.GetChild(-1) != nil || root.GetChild(2) != nil {
		t.Fatal("out-of-range child lookup did not return nil")
	}

	leaf := kustolang.ParserPathFrom([]kustolang.ParserAndIndex[byte]{
		kustolang.NewParserAndIndex(rootParser, -1),
		kustolang.NewParserAndIndex(right, 1),
	})
	if leaf == nil || leaf.Parser() != right || leaf.IndexInParent() != 1 || leaf.Parent() == nil || leaf.Parent().Parser() != rootParser {
		t.Fatalf("constructed leaf = %#v", leaf)
	}
	if kustolang.ParserPathFrom[byte](nil) != nil {
		t.Fatal("empty parser path should be nil")
	}
}

func TestPublicParserPathConcurrentChildCaching(t *testing.T) {
	rootParser := kustolang.Sequence(kustolang.MatchText("A"), kustolang.MatchText("B"))
	root := kustolang.NewParserPath(kustolang.NewParserAndIndex(rootParser, -1))
	want := root.GetChild(1)

	var wait sync.WaitGroup
	for range 64 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				if got := root.GetChild(1); got != want || got.Parent() != root {
					t.Errorf("concurrent child = %p parent %p, want %p/%p", got, got.Parent(), want, root)
					return
				}
			}
		}()
	}
	wait.Wait()
}
