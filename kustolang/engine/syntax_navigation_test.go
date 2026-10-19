package kustolang

import "testing"

func TestSyntaxOwnershipBoundsAndNavigationAcrossLists(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "  ", "a"))
	comma := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCommaToken, "", ","))
	second := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "b"))
	eof := NewMissingToken(SyntaxKindEndOfTextToken, Diagnostic{})
	firstElement := NewSeparatedElement(first, comma)
	secondElement := NewSeparatedElement(second, nil)
	list := NewSyntaxList(firstElement, secondElement)
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{list, eof})
	tree := NewSyntaxTree(root, nil, 0)

	checks := []struct {
		element SyntaxElement
		parent  SyntaxElement
		index   int
		offset  int
		start   int
		text    int
		end     int
	}{
		{root, nil, 0, 0, 0, 2, 6},
		{list, root, 0, 0, 0, 2, 6},
		{firstElement, list, 0, 0, 0, 2, 4},
		{first, firstElement, 0, 0, 0, 2, 3},
		{comma, firstElement, 1, 3, 3, 3, 4},
		{secondElement, list, 1, 4, 4, 5, 6},
		{second, secondElement, 0, 0, 4, 5, 6},
		{eof, root, 1, 6, 6, 6, 6},
	}
	for _, check := range checks {
		if check.element.Parent() != check.parent || check.element.Root() != root || check.element.Tree() != tree {
			t.Fatalf("ownership for %T = parent:%T root:%T tree:%p", check.element, check.element.Parent(), check.element.Root(), check.element.Tree())
		}
		if check.element.IndexInParent() != check.index || check.element.OffsetInParent() != check.offset {
			t.Fatalf("location for %T = index:%d offset:%d", check.element, check.element.IndexInParent(), check.element.OffsetInParent())
		}
		if check.element.TriviaStart() != check.start || check.element.TextStart() != check.text || check.element.End() != check.end {
			t.Fatalf("bounds for %T = trivia:%d text:%d end:%d", check.element, check.element.TriviaStart(), check.element.TextStart(), check.element.End())
		}
	}

	if firstElement.GetNextSibling() != secondElement || secondElement.GetPreviousSibling() != firstElement {
		t.Fatal("list sibling navigation did not preserve lexical order")
	}
	if list.GetNextSibling() != nil || list.GetNextSibling(true) != eof {
		t.Fatal("zero-width sibling filtering is incorrect")
	}
	if root.GetFirstToken() != first || root.GetLastToken() != second || root.GetLastToken(true) != eof {
		t.Fatal("first/last token navigation is incorrect")
	}
	if first.GetNextToken() != comma || comma.GetNextToken() != second || second.GetNextToken() != nil || second.GetNextToken(true) != eof {
		t.Fatal("next token navigation is incorrect")
	}
	if eof.GetPreviousToken() != second || second.GetPreviousToken() != comma || first.GetPreviousToken() != nil {
		t.Fatal("previous token navigation is incorrect")
	}

	if root.GetTokenAt(0) != first || root.GetTokenAt(2) != first || root.GetTokenAt(3) != comma || root.GetTokenAt(4) != second {
		t.Fatal("GetTokenAt did not resolve token trivia and text")
	}
	if root.GetTokenAt(6) != eof || second.GetTokenAt(100) != second || first.GetTokenAt(-1) != nil || root.GetTokenAt(-1) != nil {
		t.Fatal("GetTokenAt boundary behavior is incorrect")
	}
	if root.GetNodeAt(4, 2) != root || root.GetNodeAt(-1, 1) != nil {
		t.Fatal("GetNodeAt did not return the nearest spanning syntax node")
	}
}

func TestSyntaxCloneDetachmentAndFragmentOriginalTree(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "name"))
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{NewSyntaxList(token)})
	original := NewSyntaxTree(root, nil, 0)

	clone := root.Clone(true).(*SyntaxNode)
	cloneList := clone.GetChild(0).(*SyntaxList)
	cloneToken := cloneList.GetChild(0).(*SyntaxToken)
	if clone.Parent() != nil || clone.Tree() != nil || cloneToken.Tree() != nil {
		t.Fatal("clone retained ownership from the source tree")
	}
	if cloneToken.Parent() != cloneList || cloneToken.Root() != clone || cloneToken.TriviaStart() != 0 || cloneToken.TextStart() != 1 {
		t.Fatal("detached clone did not retain internal relative ownership")
	}
	if token.Parent() == cloneList || token.Tree() != original {
		t.Fatal("cloning changed source ownership")
	}

	fragment := NewSyntaxTree(clone, original, 12)
	if fragment.Original != original || cloneToken.Tree() != fragment {
		t.Fatal("fragment did not preserve its original tree or own cloned descendants")
	}
	if got := cloneToken.GetPositionInOriginalTree(cloneToken.TextStart()); got != 13 {
		t.Fatalf("fragment original text start = %d, want 13", got)
	}

	detachedToken := token.Clone(true).(*SyntaxToken)
	if detachedToken.Parent() != nil || detachedToken.Tree() != nil || detachedToken.TriviaStart() != 0 {
		t.Fatal("token clone was not detached")
	}
}

func TestSyntaxTreeDetachesSubtreeAndRebuildsReplacedChildOwnership(t *testing.T) {
	originalChild := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "old"))
	subtree := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{originalChild})
	outer := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{subtree})
	replacement := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "new"))
	replacementOwner := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{replacement})

	// Some existing binder tests replace immutable children directly before
	// binding. Tree construction is the ownership integration boundary.
	subtree.children[0] = replacement
	tree := NewSyntaxTree(subtree, nil, 0)

	if subtree.Parent() != nil || subtree.Root() != subtree || subtree.Tree() != tree {
		t.Fatal("subtree was not detached as the new syntax tree root")
	}
	if replacement.Parent() != subtree || replacement.Root() != subtree || replacement.Tree() != tree {
		t.Fatal("replacement child ownership was not rebuilt")
	}
	if replacement.TriviaStart() != 0 || replacement.TextStart() != 1 {
		t.Fatal("replacement child positions were not rebuilt")
	}
	if outer.GetChild(0) != subtree || replacementOwner.GetChild(0) != replacement {
		t.Fatal("tree integration unexpectedly rewrote syntax child storage")
	}
}
