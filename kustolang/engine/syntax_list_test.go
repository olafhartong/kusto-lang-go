package kustolang

import "testing"

func TestSyntaxListPreservesChildrenAndText(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a"))
	comma := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCommaToken, "", ","))
	second := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "b"))
	list := NewTypedSyntaxList("SyntaxToken", first, comma, second)

	if list.Kind() != SyntaxKindList {
		t.Fatalf("kind = %s", SyntaxText(list.Kind()))
	}
	if list.ElementType() != "SyntaxToken" {
		t.Fatalf("element type = %q", list.ElementType())
	}
	if list.Count() != 3 || list.ChildCount() != 3 {
		t.Fatalf("count=%d childCount=%d", list.Count(), list.ChildCount())
	}
	if list.At(2) != second || list.Child(2) != second {
		t.Fatal("list did not preserve child order")
	}
	if list.GetChild(1) != comma || list.GetName(1) != "" || list.IsOptional(1) {
		t.Fatalf("list child metadata = child:%#v name:%q optional:%t", list.GetChild(1), list.GetName(1), list.IsOptional(1))
	}
	if first.CompletionHint(0) != CompletionHintInherit || list.CompletionHint(1) != CompletionHintInherit {
		t.Fatalf("token/list completion hints = %d/%d", first.CompletionHint(0), list.CompletionHint(1))
	}
	if got := list.Text(IncludeTriviaAll); got != "a, b" {
		t.Fatalf("text = %q", got)
	}
	if got := list.Width(); got != 4 {
		t.Fatalf("width = %d", got)
	}
	if got := list.FullWidth(); got != 4 {
		t.Fatalf("full width = %d", got)
	}
}

func TestSeparatedElementPreservesOptionalSeparator(t *testing.T) {
	element := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a"))
	separator := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCommaToken, "", ","))
	separated := NewSeparatedElement(element, separator)

	if separated.Kind() != SyntaxKindSeparatedElement {
		t.Fatalf("kind = %s", SyntaxText(separated.Kind()))
	}
	if separated.Element() != element || separated.Separator() != separator {
		t.Fatal("separated element did not preserve element/separator")
	}
	if separated.ChildCount() != 2 || separated.Child(0) != element || separated.Child(1) != separator {
		t.Fatalf("unexpected children = %#v", separated.Children())
	}
	if separated.GetName(0) != "Element" || separated.GetName(1) != "Separator" || separated.IsOptional(0) || !separated.IsOptional(1) {
		t.Fatalf("separated metadata = %q/%t %q/%t", separated.GetName(0), separated.IsOptional(0), separated.GetName(1), separated.IsOptional(1))
	}
	if separated.CompletionHint(0) != CompletionHintInherit || separated.CompletionHint(1) != CompletionHintSyntax {
		t.Fatalf("separated completion hints = %d/%d", separated.CompletionHint(0), separated.CompletionHint(1))
	}
	if got := separated.Text(IncludeTriviaAll); got != "a," {
		t.Fatalf("text = %q", got)
	}

	emptySeparator := NewSeparatedElement(element.Clone(true), nil)
	if emptySeparator.ChildCount() != 2 || emptySeparator.Child(1) != nil {
		t.Fatalf("missing separator child = %#v", emptySeparator.Children())
	}
	if got := emptySeparator.Text(IncludeTriviaAll); got != "a" {
		t.Fatalf("text without separator = %q", got)
	}
}

func TestSyntaxListContainedDiagnostics(t *testing.T) {
	diagnostic := DiagnosticMissingExpression()
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a", diagnostic))
	list := NewSyntaxListWithDiagnostics([]SyntaxElement{token}, DiagnosticMissingText(")"))

	diagnostics := list.ContainedDiagnostics()
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Code == "" || diagnostics[1].Code == "" {
		t.Fatalf("diagnostic codes missing = %#v", diagnostics)
	}

	publicDiagnostics := list.GetContainedSyntaxDiagnostics()
	if len(publicDiagnostics) != len(diagnostics) || publicDiagnostics[0].Code != diagnostics[0].Code || publicDiagnostics[1].Code != diagnostics[1].Code {
		t.Fatalf("public contained diagnostics = %#v, want codes from %#v", publicDiagnostics, diagnostics)
	}
	for _, diagnostic := range publicDiagnostics {
		if diagnostic.LocationKind != DiagnosticLocationAbsolute || diagnostic.Start != 0 || diagnostic.Length != 1 {
			t.Fatalf("public diagnostic location = %#v", diagnostic)
		}
	}
}
