package kustolang

import "testing"

func TestSyntaxConstructorsRejectChildWithExistingParent(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account"))
	first := NewSyntaxNode(SyntaxKindTokenName, []SyntaxElement{token})

	assertSyntaxOwnershipPanic(t, func() {
		NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{token})
	})
	if token.Parent() != first || first.GetChild(0) != token {
		t.Fatal("failed attachment changed the original parent")
	}
}

func TestSyntaxConstructorsRejectChildAlreadyInTree(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account"))
	root := NewSyntaxNode(SyntaxKindTokenName, []SyntaxElement{token})
	tree := NewSyntaxTree(root, nil, 0)

	assertSyntaxOwnershipPanic(t, func() {
		NewSyntaxList(token)
	})
	if token.Parent() != root || token.Tree() != tree {
		t.Fatal("failed attachment detached the child from its syntax tree")
	}
}

func TestGeneratedConstructorsAttachInitializedSubtreeWithoutReparentingDescendants(t *testing.T) {
	nameToken := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account"))
	name := MustNewTokenNameTyped(nameToken)
	reference := MustNewNameReferenceTyped(name, SymbolMatchColumn)

	if name.Parent() != reference.SyntaxNode {
		t.Fatalf("name parent = %T, want generated name-reference node", name.Parent())
	}
	if nameToken.Parent() != name.SyntaxNode {
		t.Fatalf("name token parent = %T, want original token-name node", nameToken.Parent())
	}
	if nameToken.Root() != reference.SyntaxNode {
		t.Fatalf("name token root = %T, want generated name-reference node", nameToken.Root())
	}
}

func TestSyntaxConstructorsRejectSameChildTwice(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account"))
	assertSyntaxOwnershipPanic(t, func() {
		NewSyntaxList(token, token)
	})
}

func assertSyntaxOwnershipPanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		err, ok := recovered.(InvalidOperationError)
		if !ok || err.Message != "The syntax element already has a parent." {
			t.Fatalf("panic = %#v, want InvalidOperationError", recovered)
		}
	}()
	action()
	t.Fatal("syntax ownership violation did not panic")
}
