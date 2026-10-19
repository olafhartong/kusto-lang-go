package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxTreeRejectsAttachedNonRoot(t *testing.T) {
	token := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "Account"),
	)
	name := kustolang.MustNewTokenNameTyped(token)
	_ = kustolang.MustNewNameReferenceTyped(name, kustolang.SymbolMatchColumn)

	assertInvalidOperation(t, "Cannot assign tree. This node is not the root element.", func() {
		kustolang.NewSyntaxTree(name.SyntaxNode, nil, 0)
	})
}

func TestPublicSyntaxTreeRejectsSecondTree(t *testing.T) {
	token := kustolang.SyntaxTokenFromLexical(
		kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "Account"),
	)
	name := kustolang.MustNewTokenNameTyped(token)
	first := kustolang.NewSyntaxTree(name.SyntaxNode, nil, 0)
	if name.Tree() != first {
		t.Fatal("first syntax tree did not own its root")
	}

	assertInvalidOperation(t, "Cannot assign tree. This node already has a syntax tree.", func() {
		kustolang.NewSyntaxTree(name.SyntaxNode, nil, 0)
	})
}

func TestPublicSyntaxTreeAcceptsDetachedValidAndMalformedClones(t *testing.T) {
	validSource := kustolang.Parse(`print value=1`)
	validRoot := validSource.Syntax().Clone(true).(*kustolang.SyntaxNode)
	validTree := kustolang.NewSyntaxTree(validRoot, validSource.Syntax().Tree(), 7)
	if validRoot.Tree() != validTree || validTree.Original != validSource.Syntax().Tree() || validTree.OffsetInOriginal != 7 {
		t.Fatalf("valid clone tree = %#v", validTree)
	}

	malformedSource := kustolang.Parse(`print value=`)
	malformedRoot := malformedSource.Syntax().Clone(true).(*kustolang.SyntaxNode)
	malformedTree := kustolang.NewSyntaxTree(malformedRoot, malformedSource.Syntax().Tree(), 3)
	if diagnostics := malformedTree.Diagnostics(); len(diagnostics) == 0 || diagnostics[0].Code != "KS006" {
		t.Fatalf("malformed clone diagnostics = %#v", diagnostics)
	}
	if got := malformedRoot.GetPositionInOriginalTree(2); got != 5 {
		t.Fatalf("malformed clone original position = %d, want 5", got)
	}
}

func TestPublicCopyAsFragmentPreservesGeneratedTypeAndOriginalPosition(t *testing.T) {
	code := kustolang.Parse(`T | where Account ==`)
	filter := kustolang.GetFirstDescendant[*kustolang.FilterOperator](code.Syntax())
	if filter == nil {
		t.Fatal("missing filter operator")
	}
	originalTree := filter.Tree()
	originalStart := filter.TriviaStart()

	fragment := kustolang.CopyAsFragment(filter)
	if fragment == filter || fragment.Parent() != nil || fragment.Tree() == originalTree {
		t.Fatalf("fragment ownership = parent:%v tree:%p original:%p", fragment.Parent(), fragment.Tree(), originalTree)
	}
	if fragment.Tree().Original != originalTree || fragment.Tree().OffsetInOriginal != originalStart {
		t.Fatalf("fragment tree origin = %#v", fragment.Tree())
	}
	if diagnostics := fragment.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 || diagnostics[0].Code != "KS006" {
		t.Fatalf("fragment diagnostics = %#v", diagnostics)
	}
	if got := fragment.GetPositionInOriginalTree(fragment.TextStart()); got != originalStart+fragment.TextStart() {
		t.Fatalf("fragment original position = %d", got)
	}
}

func assertInvalidOperation(t *testing.T, message string, action func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		err, ok := recovered.(kustolang.InvalidOperationError)
		if !ok || err.Message != message {
			t.Fatalf("panic = %#v, want InvalidOperationError(%q)", recovered, message)
		}
	}()
	action()
	t.Fatal("operation did not panic")
}
