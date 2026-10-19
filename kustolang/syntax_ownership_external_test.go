package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTypedConstructorsRejectReusedSyntaxChild(t *testing.T) {
	token := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "Account"),
	)
	name := kustolang.MustNewTokenNameTyped(token)

	defer func() {
		recovered := recover()
		err, ok := recovered.(kustolang.InvalidOperationError)
		if !ok || err.Message != "The syntax element already has a parent." {
			t.Fatalf("panic = %#v, want InvalidOperationError", recovered)
		}
		if token.Parent() != name.SyntaxNode {
			t.Fatal("failed public construction changed the original parent")
		}
	}()

	kustolang.MustNewTokenNameTyped(token)
	t.Fatal("public constructor silently reparented a reused token")
}

func TestPublicCloneCanBeAttachedIndependently(t *testing.T) {
	token := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "Account"),
	)
	original := kustolang.MustNewTokenNameTyped(token)
	clonedToken := token.Clone(true).(*kustolang.SyntaxToken)
	cloneParent := kustolang.MustNewTokenNameTyped(clonedToken)

	if token.Parent() != original.SyntaxNode || clonedToken.Parent() != cloneParent.SyntaxNode {
		t.Fatal("cloned token did not retain independent ownership")
	}
	if original.Name() == cloneParent.Name() {
		t.Fatal("independent syntax parents unexpectedly share their token")
	}
}
