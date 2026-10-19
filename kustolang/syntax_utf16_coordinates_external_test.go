package kustolang_test

// Ported from the coordinate semantics of Kusto.Language syntax elements.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxExplicitByteAndUTF16Coordinates(t *testing.T) {
	first := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "😀 ", "A😀"))
	second := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "😀", "B"))
	missing := kustolang.NewMissingToken(kustolang.SyntaxKindEndOfTextToken, kustolang.Diagnostic{})
	list := kustolang.NewSyntaxList(first, second, missing)

	if list.Width() != 10 || list.ByteWidth() != 10 || list.UTF16Width() != 6 ||
		list.FullWidth() != 15 || list.ByteFullWidth() != 15 || list.UTF16FullWidth() != 9 ||
		list.TriviaWidth() != 5 || list.ByteTriviaWidth() != 5 || list.UTF16TriviaWidth() != 3 {
		t.Fatalf("public list metrics = byte:%d/%d/%d UTF-16:%d/%d/%d",
			list.Width(), list.TriviaWidth(), list.FullWidth(),
			list.UTF16Width(), list.UTF16TriviaWidth(), list.UTF16FullWidth())
	}
	if second.ByteOffsetInParent() != 10 || second.UTF16OffsetInParent() != 6 ||
		second.ByteTextStart() != 14 || second.UTF16TextStart() != 8 {
		t.Fatalf("public second coordinates = byte:%d/%d UTF-16:%d/%d",
			second.ByteOffsetInParent(), second.ByteTextStart(), second.UTF16OffsetInParent(), second.UTF16TextStart())
	}
	if list.GetTokenAtUTF16(4) != first || list.GetTokenAtUTF16(5) != first ||
		list.GetTokenAtUTF16(6) != second || list.GetTokenAtUTF16(9) != missing {
		t.Fatal("public UTF-16 token lookup failed at surrogate or missing-token boundary")
	}
	if list.GetTokenAtByte(15) != missing || list.GetTokenAt(15) != missing {
		t.Fatal("public byte token lookup changed unsuffixed behavior")
	}
}

func TestPublicSyntaxFragmentProjectsUTF16Position(t *testing.T) {
	code := kustolang.Parse(`print face="😀" | project face`)
	project := kustolang.GetFirstDescendant[*kustolang.ProjectOperator](code.Syntax())
	if project == nil {
		t.Fatal("missing project operator")
	}
	fragment := kustolang.CopyAsFragment(project)

	if got := fragment.GetPositionInOriginalTreeByte(fragment.ByteTextStart()); got != project.ByteTextStart() {
		t.Fatalf("public byte fragment text start = %d, want %d", got, project.ByteTextStart())
	}
	if got := fragment.GetPositionInOriginalTreeUTF16(fragment.UTF16TextStart()); got != project.UTF16TextStart() {
		t.Fatalf("public UTF-16 fragment text start = %d, want %d", got, project.UTF16TextStart())
	}
}

func TestPublicSyntaxUTF16NodeAndRangeNavigation(t *testing.T) {
	const text = `print prefix="😀", x=`
	root := kustolang.Parse(text).Syntax()
	token := root.GetTokenAtUTF16(14)
	if token == nil || root.GetTokenAtUTF16(15) != token || token.Kind() != kustolang.SyntaxKindStringLiteralToken {
		t.Fatalf("public astral token lookup = %#v", token)
	}
	if root.GetNodeAtUTF16(14, 2) == nil || root.GetNodeAtByte(14, 4) == nil || root.GetNodeAtUTF16(21, 0) == nil {
		t.Fatal("public byte/UTF-16 node lookup failed")
	}
	var walked []*kustolang.SyntaxToken
	kustolang.WalkSyntaxTokensUTF16(root, 14, 16, func(current *kustolang.SyntaxToken) { walked = append(walked, current) })
	if len(walked) != 1 || walked[0] != token {
		t.Fatalf("public UTF-16 ranged tokens = %#v", walked)
	}
}
