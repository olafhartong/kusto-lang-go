package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTypedSyntaxTraversalCompleteSurface(t *testing.T) {
	code := kustolang.Parse("T | project X = A + B")
	root := code.Syntax()
	binary := kustolang.GetFirstDescendant[*kustolang.BinaryExpression](root)
	if binary == nil {
		t.Fatal("binary expression was not found")
	}
	name := kustolang.GetFirstDescendant[*kustolang.NameReference](binary)
	if name == nil {
		t.Fatal("name reference was not found")
	}
	if got := kustolang.GetFirstAncestorOrSelf[*kustolang.NameReference](name); got != name {
		t.Fatalf("first ancestor-or-self = %#v", got)
	}
	if got := kustolang.GetFirstAncestorOrSelf[*kustolang.BinaryExpression](name); got != binary {
		t.Fatalf("binary ancestor = %#v, want %#v", got, binary)
	}
	if got := kustolang.GetFirstDescendantOrSelf[*kustolang.BinaryExpression](binary); got != binary {
		t.Fatalf("first descendant-or-self = %#v", got)
	}

	ancestors := kustolang.GetAncestors[kustolang.SyntaxElement](name)
	ancestorsOrSelf := kustolang.GetAncestorsOrSelf[kustolang.SyntaxElement](name)
	if len(ancestorsOrSelf) != len(ancestors)+1 || ancestorsOrSelf[0] != name {
		t.Fatalf("ancestors = %d, ancestors-or-self = %#v", len(ancestors), ancestorsOrSelf)
	}
	descendants := kustolang.GetDescendants[kustolang.SyntaxElement](binary)
	descendantsOrSelf := kustolang.GetDescendantsOrSelf[kustolang.SyntaxElement](binary)
	if len(descendantsOrSelf) != len(descendants)+1 || descendantsOrSelf[0] != binary {
		t.Fatalf("descendants = %d, descendants-or-self = %#v", len(descendants), descendantsOrSelf)
	}

	filtered := kustolang.GetAncestorsOrSelf[kustolang.SyntaxElement](name, func(element kustolang.SyntaxElement) bool {
		return element == binary
	})
	if len(filtered) != 1 || filtered[0] != binary {
		t.Fatalf("filtered ancestors = %#v", filtered)
	}
	if got := kustolang.GetAncestorsOrSelf[kustolang.SyntaxElement](nil); len(got) != 0 {
		t.Fatalf("nil ancestors = %#v", got)
	}
}

func TestPublicWalkSyntaxNodesExcludesTokensAndHonorsDescend(t *testing.T) {
	root := kustolang.Parse("T | project X = A + B").Syntax()
	visited := []*kustolang.SyntaxNode{}
	pruned := (*kustolang.SyntaxNode)(nil)

	kustolang.WalkSyntaxNodes(root, func(node *kustolang.SyntaxNode) {
		visited = append(visited, node)
		if pruned == nil && node != root {
			pruned = node
		}
	}, nil, func(node *kustolang.SyntaxNode) bool {
		return node != pruned
	})

	if len(visited) < 2 || visited[0] != root || pruned == nil {
		t.Fatalf("visited nodes = %#v", visited)
	}
	for _, node := range visited {
		if node == nil || node.IsToken() {
			t.Fatalf("WalkSyntaxNodes returned non-node %#v", node)
		}
		if node != pruned && kustolang.SyntaxIsAncestorOf(pruned, node) {
			t.Fatalf("walk descended below pruned node: %#v", node)
		}
	}
}

func TestPublicSyntaxIndexAndParentNameNavigation(t *testing.T) {
	root := kustolang.Parse("T | project X = A + B").Syntax()
	binary := kustolang.GetFirstDescendant[*kustolang.BinaryExpression](root)
	name := kustolang.GetFirstDescendant[*kustolang.NameReference](binary)
	if binary == nil || name == nil || name.Parent() == nil {
		t.Fatal("expected parsed descendants")
	}

	parent := name.Parent()
	index := kustolang.SyntaxGetChildIndex(parent, name)
	if index < 0 || parent.GetChild(index) != name {
		t.Fatalf("child index = %d", index)
	}
	if got := kustolang.SyntaxNameInParent(name); got != parent.GetName(index) {
		t.Fatalf("name in parent = %q, want %q", got, parent.GetName(index))
	}
	descendantIndex := kustolang.SyntaxGetDescendantIndex(binary, name)
	if descendantIndex < 0 || binary.GetChild(descendantIndex) == nil || !kustolang.SyntaxIsAncestorOf(binary.GetChild(descendantIndex), name) && binary.GetChild(descendantIndex) != name {
		t.Fatalf("descendant index = %d", descendantIndex)
	}
	if kustolang.SyntaxGetChildIndex(binary, root) != -1 || kustolang.SyntaxGetDescendantIndex(name, binary) != -1 {
		t.Fatal("unrelated navigation should return -1")
	}
}

func TestPublicSyntaxElementNavigationAndTraversalMethods(t *testing.T) {
	var root kustolang.SyntaxElement = kustolang.Parse("T | project X = A + B").Syntax()
	var binary kustolang.SyntaxElement = kustolang.GetFirstDescendant[*kustolang.BinaryExpression](root)
	var name kustolang.SyntaxElement = kustolang.GetFirstDescendant[*kustolang.NameReference](binary)
	if binary == nil || name == nil || name.Parent() == nil {
		t.Fatal("expected parsed descendants")
	}

	parent := name.Parent()
	childIndex := parent.GetChildIndex(name)
	if childIndex < 0 || parent.GetChild(childIndex) != name {
		t.Fatalf("child index = %d", childIndex)
	}
	if got := name.NameInParent(); got != parent.GetName(childIndex) {
		t.Fatalf("name in parent = %q, want %q", got, parent.GetName(childIndex))
	}
	if root.Depth() != 0 || name.Depth() <= binary.Depth() {
		t.Fatalf("depths: root=%d binary=%d name=%d", root.Depth(), binary.Depth(), name.Depth())
	}
	descendantIndex := binary.GetDescendantIndex(name)
	if descendantIndex < 0 || !binary.IsAncestorOf(name) {
		t.Fatalf("descendant index = %d, ancestor = %t", descendantIndex, binary.IsAncestorOf(name))
	}
	if binary.GetChild(descendantIndex) != name && !binary.GetChild(descendantIndex).IsAncestorOf(name) {
		t.Fatalf("child %d does not contain descendant", descendantIndex)
	}

	tokens := root.GetTokens()
	walkedTokens := []*kustolang.SyntaxToken{}
	root.WalkTokens(func(token *kustolang.SyntaxToken) {
		walkedTokens = append(walkedTokens, token)
	})
	if len(tokens) == 0 || len(walkedTokens) != len(tokens) {
		t.Fatalf("tokens=%d walked tokens=%d", len(tokens), len(walkedTokens))
	}
	for index := range tokens {
		if walkedTokens[index] != tokens[index] {
			t.Fatalf("walked token %d differs from GetTokens", index)
		}
	}

	walkedElements := []kustolang.SyntaxElement{}
	root.WalkElements(func(element kustolang.SyntaxElement) {
		walkedElements = append(walkedElements, element)
	})
	if len(walkedElements) == 0 || walkedElements[0] != root {
		t.Fatalf("walked elements = %#v", walkedElements)
	}

	var malformed kustolang.SyntaxElement = kustolang.ParseFunctionBody("{ T | where A == }")
	if root.HasMissingChildren() || !malformed.HasMissingChildren() {
		t.Fatalf("missing children: valid=%t malformed=%t", root.HasMissingChildren(), malformed.HasMissingChildren())
	}
}
