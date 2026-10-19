package kustolang

// Ported from the coordinate semantics of Kusto.Language syntax elements.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import "testing"

func TestSyntaxExplicitByteAndUTF16Coordinates(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "😀 ", "A😀"))
	second := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "😀", "B"))
	missing := NewMissingToken(SyntaxKindEndOfTextToken, Diagnostic{})
	list := NewSyntaxList(first, second)
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{list, missing})
	NewSyntaxTree(root, nil, 0)

	assertSyntaxCoordinateMetrics(t, first, 5, 5, 10, 3, 3, 6)
	assertSyntaxCoordinateMetrics(t, second, 1, 4, 5, 1, 2, 3)
	assertSyntaxCoordinateMetrics(t, list, 10, 5, 15, 6, 3, 9)
	assertSyntaxCoordinateMetrics(t, root, 10, 5, 15, 6, 3, 9)
	assertSyntaxCoordinateMetrics(t, missing, 0, 0, 0, 0, 0, 0)

	if second.ByteOffsetInParent() != 10 || second.UTF16OffsetInParent() != 6 {
		t.Fatalf("second sibling offsets = byte:%d UTF-16:%d", second.ByteOffsetInParent(), second.UTF16OffsetInParent())
	}
	if second.ByteTriviaStart() != 10 || second.UTF16TriviaStart() != 6 ||
		second.ByteTextStart() != 14 || second.UTF16TextStart() != 8 ||
		second.ByteEnd() != 15 || second.UTF16End() != 9 {
		t.Fatalf("second bounds = byte:%d/%d/%d UTF-16:%d/%d/%d",
			second.ByteTriviaStart(), second.ByteTextStart(), second.ByteEnd(),
			second.UTF16TriviaStart(), second.UTF16TextStart(), second.UTF16End())
	}
	if missing.ByteOffsetInParent() != 15 || missing.UTF16OffsetInParent() != 9 ||
		missing.ByteTriviaStart() != 15 || missing.UTF16TriviaStart() != 9 {
		t.Fatalf("missing token offsets = byte:%d/%d UTF-16:%d/%d",
			missing.ByteOffsetInParent(), missing.ByteTriviaStart(),
			missing.UTF16OffsetInParent(), missing.UTF16TriviaStart())
	}

	if root.GetTokenAtByte(7) != first || root.GetTokenAtByte(10) != second || root.GetTokenAtByte(15) != missing {
		t.Fatal("byte token lookup did not preserve existing boundaries")
	}
	if root.GetTokenAtUTF16(4) != first || root.GetTokenAtUTF16(5) != first {
		t.Fatal("UTF-16 token lookup did not resolve both units of an astral rune")
	}
	if root.GetTokenAtUTF16(6) != second || root.GetTokenAtUTF16(7) != second || root.GetTokenAtUTF16(9) != missing {
		t.Fatal("UTF-16 token lookup did not resolve trivia or trailing missing token")
	}
	if root.GetTokenAtByte(-1) != nil || root.GetTokenAtUTF16(-1) != nil {
		t.Fatal("negative token positions should not resolve")
	}
}

func TestSyntaxFragmentProjectsByteAndUTF16Positions(t *testing.T) {
	prefix := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "😀"))
	fragmentToken := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "A😀"))
	fragmentSource := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{fragmentToken})
	originalRoot := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{prefix, fragmentSource})
	original := NewSyntaxTree(originalRoot, nil, 0)

	fragment := CopySyntaxElementAsFragment(fragmentSource)
	if fragment.Tree() == nil || fragment.Tree().Original != original || fragment.Tree().OffsetInOriginal != 4 {
		t.Fatalf("fragment origin = %#v", fragment.Tree())
	}
	if got := fragment.GetPositionInOriginalTreeByte(fragment.ByteTextStart()); got != 5 {
		t.Fatalf("byte fragment text start = %d, want 5", got)
	}
	if got := fragment.GetPositionInOriginalTreeUTF16(fragment.UTF16TextStart()); got != 3 {
		t.Fatalf("UTF-16 fragment text start = %d, want 3", got)
	}
	if got := fragment.GetPositionInOriginalTreeUTF16(fragment.UTF16End()); got != 6 {
		t.Fatalf("UTF-16 fragment end = %d, want 6", got)
	}

	nested := CopySyntaxElementAsFragment(fragment)
	if got := nested.GetPositionInOriginalTreeUTF16(nested.UTF16TextStart()); got != 3 {
		t.Fatalf("nested UTF-16 fragment text start = %d, want 3", got)
	}
}

func TestSyntaxUTF16NodeLookupAndTokenWalkingOnMalformedQuery(t *testing.T) {
	const text = `print prefix="😀", x=`
	root := Parse(text).Syntax()
	if root.ByteFullWidth() != len(text) || root.UTF16FullWidth() != 21 {
		t.Fatalf("malformed root widths = byte:%d UTF-16:%d", root.ByteFullWidth(), root.UTF16FullWidth())
	}

	firstUnit := root.GetTokenAtUTF16(14)
	secondUnit := root.GetTokenAtUTF16(15)
	if firstUnit == nil || secondUnit != firstUnit || firstUnit.Kind() != SyntaxKindStringLiteralToken {
		t.Fatalf("astral token lookup = first:%#v second:%#v", firstUnit, secondUnit)
	}
	if node := root.GetNodeAtUTF16(14, 2); node == nil || node.UTF16End() < 16 {
		t.Fatalf("UTF-16 node lookup = %#v", node)
	}
	if node := root.GetNodeAtByte(14, 4); node == nil || node.ByteEnd() < 18 {
		t.Fatalf("byte node lookup = %#v", node)
	}
	if missing := root.GetNodeAtUTF16(21, 0); missing == nil || missing.UTF16End() != 21 {
		t.Fatalf("missing node lookup = %#v", missing)
	}

	var utf16Tokens []*SyntaxToken
	WalkSyntaxTokensUTF16(root, 14, 16, func(token *SyntaxToken) { utf16Tokens = append(utf16Tokens, token) })
	if len(utf16Tokens) != 1 || utf16Tokens[0] != firstUnit {
		t.Fatalf("UTF-16 ranged tokens = %#v", utf16Tokens)
	}
	var byteTokens []*SyntaxToken
	WalkSyntaxTokensByte(root, 14, 18, func(token *SyntaxToken) { byteTokens = append(byteTokens, token) })
	if len(byteTokens) != 1 || byteTokens[0] != firstUnit {
		t.Fatalf("byte ranged tokens = %#v", byteTokens)
	}
}

func assertSyntaxCoordinateMetrics(t *testing.T, element SyntaxElement, byteWidth, byteTrivia, byteFull, utf16Width, utf16Trivia, utf16Full int) {
	t.Helper()
	if element.Width() != byteWidth || element.ByteWidth() != byteWidth ||
		element.TriviaWidth() != byteTrivia || element.ByteTriviaWidth() != byteTrivia ||
		element.FullWidth() != byteFull || element.ByteFullWidth() != byteFull {
		t.Fatalf("%T byte metrics = width:%d/%d trivia:%d/%d full:%d/%d",
			element, element.Width(), element.ByteWidth(), element.TriviaWidth(), element.ByteTriviaWidth(), element.FullWidth(), element.ByteFullWidth())
	}
	if element.UTF16Width() != utf16Width || element.UTF16TriviaWidth() != utf16Trivia || element.UTF16FullWidth() != utf16Full {
		t.Fatalf("%T UTF-16 metrics = width:%d trivia:%d full:%d",
			element, element.UTF16Width(), element.UTF16TriviaWidth(), element.UTF16FullWidth())
	}
}
