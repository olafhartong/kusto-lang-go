package kustolang

import "testing"

func TestSyntaxTokenCloneRespectsDiagnostics(t *testing.T) {
	diagnostic := DiagnosticMissingExpression()
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "a", diagnostic))

	withDiagnostics := token.Clone(true).(*SyntaxToken)
	if withDiagnostics == token {
		t.Fatal("token clone returned original pointer")
	}
	if withDiagnostics.RawText() != "a" || withDiagnostics.Trivia() != " " {
		t.Fatalf("cloned token text/trivia = %q/%q", withDiagnostics.RawText(), withDiagnostics.Trivia())
	}
	if len(withDiagnostics.Diagnostics()) != 1 {
		t.Fatalf("diagnostics with include = %#v", withDiagnostics.Diagnostics())
	}

	withoutDiagnostics := token.Clone(false).(*SyntaxToken)
	if len(withoutDiagnostics.Diagnostics()) != 0 {
		t.Fatalf("diagnostics without include = %#v", withoutDiagnostics.Diagnostics())
	}
}

func TestSyntaxListCloneRecursesChildren(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a", DiagnosticMissingExpression()))
	list := NewTypedSyntaxListWithDiagnostics("SyntaxToken", []SyntaxElement{token}, DiagnosticMissingText(")"))

	clone := list.Clone(true).(*SyntaxList)
	if clone == list {
		t.Fatal("list clone returned original pointer")
	}
	if clone.ElementType() != "SyntaxToken" || clone.Count() != 1 {
		t.Fatalf("clone metadata = type:%q count:%d", clone.ElementType(), clone.Count())
	}
	if clone.At(0) == token {
		t.Fatal("list child was not cloned")
	}
	if len(clone.ContainedDiagnostics()) != 2 {
		t.Fatalf("clone diagnostics = %#v", clone.ContainedDiagnostics())
	}

	withoutDiagnostics := list.Clone(false).(*SyntaxList)
	if len(withoutDiagnostics.ContainedDiagnostics()) != 0 {
		t.Fatalf("diagnostics without include = %#v", withoutDiagnostics.ContainedDiagnostics())
	}
}

func TestSeparatedElementClonePreservesSeparator(t *testing.T) {
	element := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a"))
	separator := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCommaToken, "", ","))
	separated := NewSeparatedElementWithDiagnostics(element, separator, DiagnosticMissingExpression())

	clone := separated.Clone(true).(*SeparatedElement)
	if clone == separated {
		t.Fatal("separated clone returned original pointer")
	}
	if clone.Element() == element || clone.Separator() == separator {
		t.Fatal("separated clone reused child pointers")
	}
	if got := clone.Text(IncludeTriviaAll); got != "a," {
		t.Fatalf("clone text = %q", got)
	}
	if len(clone.Diagnostics()) != 1 {
		t.Fatalf("clone diagnostics = %#v", clone.Diagnostics())
	}
}

func TestCustomNodeClonePreservesShape(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"))
	custom, err := NewCustomNodeWithShape(
		[]CustomElementDescriptor{
			NewCustomElementDescriptor("Value", CompletionHintExpression, false),
			NewCustomElementDescriptor("Optional", CompletionHintSyntax, true),
		},
		[]SyntaxElement{token, nil},
		DiagnosticMissingText("custom"),
	)
	if err != nil {
		t.Fatalf("NewCustomNodeWithShape: %v", err)
	}

	clone := custom.Clone(true).(*CustomNode)
	if clone == custom {
		t.Fatal("custom clone returned original pointer")
	}
	if clone.GetName(0) != "Value" || !clone.IsOptional(1) || clone.CompletionHint(0) != CompletionHintExpression {
		t.Fatalf("custom shape not preserved")
	}
	if clone.GetChild(0) == token {
		t.Fatal("custom child was not cloned")
	}
	if len(clone.Diagnostics()) != 1 {
		t.Fatalf("custom clone diagnostics = %#v", clone.Diagnostics())
	}
}

func TestCustomNodeAllowsOmittedTrailingShapeSlots(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"))
	custom, err := NewCustomNodeWithShape(
		[]CustomElementDescriptor{
			NewCustomElementDescriptor("Value", CompletionHintExpression, false),
			NewCustomElementDescriptor("Recovered", CompletionHintSyntax, false),
		},
		[]SyntaxElement{token},
	)
	if err != nil {
		t.Fatalf("NewCustomNodeWithShape: %v", err)
	}
	if custom.ChildCount() != 2 || custom.GetChild(1) != nil || custom.GetName(1) != "Recovered" || custom.IsOptional(1) {
		t.Fatalf("custom omitted shape slot = child:%#v name:%q optional:%t", custom.GetChild(1), custom.GetName(1), custom.IsOptional(1))
	}
	clone := custom.Clone(true).(*CustomNode)
	if clone.ChildCount() != 2 || clone.GetChild(1) != nil || clone.GetName(1) != "Recovered" {
		t.Fatalf("custom clone omitted shape slot = child:%#v name:%q", clone.GetChild(1), clone.GetName(1))
	}

	if _, err := NewCustomNodeWithShape(custom.Shape(), []SyntaxElement{nil}); err == nil {
		t.Fatal("supplied missing required custom child did not fail")
	}
}

func TestGeneratedWrapperCloneTyped(t *testing.T) {
	projectKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeyword, "", SyntaxText(SyntaxKindProjectKeyword)))
	expression := MustNewLiteralExpression(SyntaxKindStringLiteralExpression, []SyntaxElement{
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'x'")),
	})
	expressions := NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(expression, nil))
	project := MustNewProjectOperator([]SyntaxElement{projectKeyword, expressions}, DiagnosticMissingText("project"))

	clone := project.CloneTyped(true)
	if clone == project {
		t.Fatal("generated clone returned original pointer")
	}
	if clone.GeneratedNodeName() != "ProjectOperator" || clone.GetName(1) != "Expressions" {
		t.Fatalf("generated clone metadata = name:%q child:%q", clone.GeneratedNodeName(), clone.GetName(1))
	}
	if clone.ProjectKeyword() == projectKeyword {
		t.Fatal("generated clone reused token child")
	}
	if clone.Expressions().SyntaxList == expressions {
		t.Fatal("generated clone reused list child")
	}
	if got := clone.Expressions().At(0).Element(); got == nil || got.Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("generated clone typed list element = %#v", got)
	}
	if len(clone.Diagnostics()) != 1 {
		t.Fatalf("generated clone diagnostics = %#v", clone.Diagnostics())
	}

	withoutDiagnostics := project.CloneTyped(false)
	if len(withoutDiagnostics.ContainedDiagnostics()) != 0 {
		t.Fatalf("generated clone diagnostics without include = %#v", withoutDiagnostics.ContainedDiagnostics())
	}
}

func TestGeneratedWrapperClonePreservesPayloads(t *testing.T) {
	name := MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account")))
	reference := MustNewNameReferenceTyped(name, SymbolMatchColumn, DiagnosticMissingExpression())

	clone := reference.CloneTyped(false)
	if clone == reference || clone.SyntaxNode == reference.SyntaxNode {
		t.Fatal("payload clone reused original node")
	}
	if clone.Match() != SymbolMatchColumn {
		t.Fatalf("cloned match = %d", clone.Match())
	}
	if len(clone.Diagnostics()) != 0 {
		t.Fatalf("clone diagnostics without include = %#v", clone.Diagnostics())
	}

	dot := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindDotToken, "", "."))
	partial := MustNewPartialCommandTyped([]string{"show", "set"}, dot, AsSyntaxListOf[SyntaxElement](NewTypedSyntaxList("SyntaxElement", reference)))
	partialClone := partial.CloneTyped(true)
	kinds := partialClone.CommandKinds()
	kinds[0] = "mutated"
	if partialClone.CommandKinds()[0] != "show" {
		t.Fatalf("cloned command kinds were not copied")
	}
}

func TestQueryBlockClonePreservesSegments(t *testing.T) {
	query := ParseQuery("T | take 1")
	clone := query.CloneTyped(true)
	if clone == query || clone.SyntaxNode == query.SyntaxNode {
		t.Fatal("query clone reused original node")
	}
	if clone.GeneratedNodeName() != "QueryBlock" || clone.GetName(1) != "Statements" {
		t.Fatalf("query clone metadata = %q/%q", clone.GeneratedNodeName(), clone.GetName(1))
	}
	if len(clone.Segments) != len(query.Segments) || clone.Segments[1].Text != " take 1" {
		t.Fatalf("query clone segments = %#v", clone.Segments)
	}
	if got := clone.Text(IncludeTriviaAll); got != "T | take 1" {
		t.Fatalf("query clone text = %q", got)
	}
}
