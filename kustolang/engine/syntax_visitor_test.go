package kustolang

import "testing"

func TestSyntaxVisitorDispatchesGeneratedNodes(t *testing.T) {
	if generatedSyntaxVisitorMethodCount != 213 {
		t.Fatalf("visitor method count = %d, want 213", generatedSyntaxVisitorMethodCount)
	}

	projectKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeyword, "", SyntaxText(SyntaxKindProjectKeyword)))
	expressions := NewTypedSyntaxList("SeparatedElement<Expression>")
	project := MustNewProjectOperator([]SyntaxElement{projectKeyword, expressions})

	var visited []string
	visitor := SyntaxVisitorFuncs{
		DefaultVisit: func(element SyntaxElement) {
			visited = append(visited, "default")
		},
		VisitProjectOperatorFunc: func(node *ProjectOperator) {
			visited = append(visited, "project:"+node.GetName(1))
		},
	}
	AcceptSyntaxVisitor(project, visitor)
	AcceptSyntaxVisitor(project.SyntaxNode, visitor)

	if len(visited) != 2 || visited[0] != "project:Expressions" || visited[1] != "project:Expressions" {
		t.Fatalf("visited = %#v", visited)
	}
}

func TestSyntaxVisitorDispatchesBaseElements(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"))
	list := NewTypedSyntaxList("SyntaxToken", token.Clone(true))
	separated := NewSeparatedElement(token.Clone(true), nil)
	custom, err := NewCustomNodeWithShape(
		[]CustomElementDescriptor{
			NewCustomElementDescriptor("First", CompletionHintSyntax, false),
			NewCustomElementDescriptor("Second", CompletionHintExpression, true),
		},
		[]SyntaxElement{token.Clone(true), nil},
	)
	if err != nil {
		t.Fatalf("NewCustomNodeWithShape: %v", err)
	}

	var visited []string
	visitor := SyntaxVisitorFuncs{
		VisitListFunc: func(node *SyntaxList) {
			visited = append(visited, "list")
		},
		VisitSeparatedElementFunc: func(node *SeparatedElement) {
			visited = append(visited, "separated:"+node.GetName(1))
		},
		VisitCustomFunc: func(node *CustomNode) {
			visited = append(visited, "custom:"+node.GetName(0))
		},
	}
	AcceptSyntaxVisitor(list, visitor)
	AcceptSyntaxVisitor(separated, visitor)
	AcceptSyntaxVisitor(custom, visitor)

	if len(visited) != 3 || visited[0] != "list" || visited[1] != "separated:Separator" || visited[2] != "custom:First" {
		t.Fatalf("visited = %#v", visited)
	}
	if !custom.IsOptional(1) || custom.CompletionHint(1) != CompletionHintExpression {
		t.Fatalf("custom metadata optional=%t hint=%d", custom.IsOptional(1), custom.CompletionHint(1))
	}
}

func TestSyntaxVisitorResultDispatchesDynamicAndQueryNodes(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'x'"))
	literal := MustNewLiteralExpression(SyntaxKindStringLiteralExpression, []SyntaxElement{token})
	query := ParseQuery("T | take 1")

	visitor := SyntaxVisitorResultFuncs[string]{
		DefaultVisit: func(element SyntaxElement) string {
			return "default"
		},
		VisitLiteralExpressionFunc: func(node *LiteralExpression) string {
			return "literal:" + node.GetName(0)
		},
		VisitQueryBlockFunc: func(node *QueryBlock) string {
			return "query:" + node.GetName(1)
		},
	}

	if got := AcceptSyntaxVisitorResult[string](literal.SyntaxNode, visitor); got != "literal:Token" {
		t.Fatalf("literal result = %q", got)
	}
	if got := AcceptSyntaxVisitorResult[string](query.SyntaxNode, visitor); got != "query:Statements" {
		t.Fatalf("query result = %q", got)
	}
	if got := AcceptSyntaxVisitorResult[string](NewSyntaxNode(SyntaxKindNone, nil), visitor); got != "" {
		t.Fatalf("unknown node result = %q", got)
	}
}
