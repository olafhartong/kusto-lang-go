package kustolang

import (
	"reflect"
	"testing"
)

func TestSyntaxTraversalAncestorsDescendantsAndNames(t *testing.T) {
	query := ParseQuery("T | where X > 1 | project X")
	NewSyntaxTree(query.SyntaxNode, nil, 0)
	where := GetFirstDescendant[*FilterOperator](query)
	if where == nil {
		t.Fatal("missing generated filter descendant")
	}
	name := GetFirstDescendant[*NameReference](where, func(reference *NameReference) bool {
		return NameReferenceSimpleName(reference) == "X"
	})
	if name == nil {
		t.Fatal("missing filtered name descendant")
	}
	if got := GetFirstAncestor[*FilterOperator](name); got == nil || got.SyntaxNode != where.SyntaxNode {
		t.Fatalf("filter ancestor = %#v", got)
	}
	if got := GetFirstAncestorOrSelf[*NameReference](name); got == nil || got.SyntaxNode != name.SyntaxNode {
		t.Fatalf("self ancestor = %#v", got)
	}
	if !SyntaxIsAncestorOf(where, name) || SyntaxIsAncestorOf(name, where) {
		t.Fatal("ancestor relation is incorrect")
	}
	if index := SyntaxGetDescendantIndex(where, name); index < 0 {
		t.Fatalf("descendant index = %d", index)
	}
	if SyntaxNameInParent(where) == "" || SyntaxDepth(name) <= SyntaxDepth(where) {
		t.Fatalf("name/depth = %q %d <= %d", SyntaxNameInParent(where), SyntaxDepth(name), SyntaxDepth(where))
	}

	names := GetDescendants[*NameReference](query, func(reference *NameReference) bool {
		return NameReferenceSimpleName(reference) == "X"
	})
	if len(names) != 2 {
		t.Fatalf("X descendants = %d, want 2", len(names))
	}
	ancestors := GetAncestors[SyntaxElement](name)
	if len(ancestors) == 0 || ancestors[0] != name.Parent() || ancestors[len(ancestors)-1] != query.SyntaxNode {
		t.Fatalf("ancestor order = %#v", ancestors)
	}
}

func TestSyntaxTraversalWalkOrderTokensAndPruning(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "a"))
	missing := NewMissingToken(SyntaxKindCommaToken, NewDiagnostic("KS006", "Token expected: ,"))
	second := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "b"))
	left := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{first, missing})
	right := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{second})
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{left, right})
	NewSyntaxTree(root, nil, 0)

	if !SyntaxIsMissing(missing) || !SyntaxHasMissingChildren(root) || SyntaxGetChildIndex(root, right) != 1 {
		t.Fatal("missing-child or child-index behavior is incorrect")
	}
	if !missing.ContainsSyntaxDiagnostics() || !missing.HasSyntaxDiagnostics() || !missing.IsMissing() {
		t.Fatalf("missing token diagnostic state = contains:%t has:%t missing:%t", missing.ContainsSyntaxDiagnostics(), missing.HasSyntaxDiagnostics(), missing.IsMissing())
	}
	if !root.ContainsSyntaxDiagnostics() || root.HasSyntaxDiagnostics() || root.IsMissing() {
		t.Fatalf("root diagnostic state = contains:%t has:%t missing:%t", root.ContainsSyntaxDiagnostics(), root.HasSyntaxDiagnostics(), root.IsMissing())
	}
	if !root.HasMissingChildren() || root.GetChildIndex(right) != 1 || root.GetDescendantIndex(second) != 1 || !root.IsAncestorOf(second) {
		t.Fatal("promoted navigation methods are inconsistent")
	}
	if got := SyntaxGetCommonAncestor(first, missing); got != left {
		t.Fatalf("common ancestor = %#v, want left", got)
	}
	if tokens := GetSyntaxTokens(root); !reflect.DeepEqual(tokens, []*SyntaxToken{first, second}) {
		t.Fatalf("tokens = %#v", tokens)
	}
	if tokens := GetSyntaxTokens(root, true); !reflect.DeepEqual(tokens, []*SyntaxToken{first, missing, second}) {
		t.Fatalf("tokens including missing = %#v", tokens)
	}
	if tokens := root.GetTokens(true); !reflect.DeepEqual(tokens, []*SyntaxToken{first, missing, second}) {
		t.Fatalf("promoted token collection = %#v", tokens)
	}

	var before, after []SyntaxKind
	WalkSyntaxElements(root, func(element SyntaxElement) {
		before = append(before, element.Kind())
	}, func(element SyntaxElement) {
		after = append(after, element.Kind())
	}, func(element SyntaxElement) bool {
		return element != left
	})
	if !reflect.DeepEqual(before, []SyntaxKind{SyntaxKindQueryBlock, SyntaxKindNameReference, SyntaxKindNameReference, SyntaxKindIdentifierToken}) {
		t.Fatalf("preorder with pruning = %#v", before)
	}
	if !reflect.DeepEqual(after, []SyntaxKind{SyntaxKindNameReference, SyntaxKindIdentifierToken, SyntaxKindNameReference, SyntaxKindQueryBlock}) {
		t.Fatalf("postorder with pruning = %#v", after)
	}

	var ranged []*SyntaxToken
	WalkSyntaxTokens(root, second.TriviaStart(), second.End(), func(token *SyntaxToken) { ranged = append(ranged, token) })
	if !reflect.DeepEqual(ranged, []*SyntaxToken{second}) {
		t.Fatalf("ranged tokens = %#v", ranged)
	}
}

func TestGetOriginalSyntaxNodeAcrossFragment(t *testing.T) {
	query := ParseQuery("T | where X > 1")
	original := NewSyntaxTree(query.SyntaxNode, nil, 0)
	where := GetFirstDescendant[*FilterOperator](query)
	if where == nil {
		t.Fatal("missing original filter")
	}
	clone := where.SyntaxNode.Clone(true).(*SyntaxNode)
	NewSyntaxTree(clone, original, where.TriviaStart())
	if got := clone.GetOriginalNode(); got != where.SyntaxNode {
		t.Fatalf("original node = %#v, want %#v", got, where.SyntaxNode)
	}
}

func TestCopyAsFragmentPreservesGeneratedWrapperAndOriginalTree(t *testing.T) {
	query := ParseQuery("T | where X > 1")
	original := NewSyntaxTree(query.SyntaxNode, nil, 0)
	where := GetFirstDescendant[*FilterOperator](query)
	if where == nil {
		t.Fatal("missing original filter")
	}

	fragment := CopyAsFragment(where)
	if fragment == where || fragment.SyntaxNode == where.SyntaxNode {
		t.Fatal("fragment reused original generated node")
	}
	if fragment.Tree() == nil || fragment.Tree().Root != fragment.SyntaxNode {
		t.Fatalf("fragment tree root = %#v, want cloned filter node", fragment.Tree())
	}
	if fragment.Tree().Original != original || fragment.Tree().OffsetInOriginal != where.TriviaStart() {
		t.Fatalf("fragment original/offset = %#v/%d, want %#v/%d", fragment.Tree().Original, fragment.Tree().OffsetInOriginal, original, where.TriviaStart())
	}
	if got := fragment.GetOriginalNode(); got != where.SyntaxNode {
		t.Fatalf("fragment original node = %#v, want %#v", got, where.SyntaxNode)
	}
	if fragment.Condition() == where.Condition() {
		t.Fatal("fragment reused generated child")
	}
}

func TestCopyAsFragmentPreservesCustomNodeShape(t *testing.T) {
	custom, err := NewCustomNodeWithShape(
		[]CustomElementDescriptor{
			NewCustomElementDescriptor("Value", CompletionHintExpression, false),
			NewCustomElementDescriptor("Optional", CompletionHintSyntax, true),
		},
		[]SyntaxElement{SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x")), nil},
		DiagnosticMissingExpression(),
	)
	if err != nil {
		t.Fatalf("NewCustomNodeWithShape: %v", err)
	}
	original := NewSyntaxTree(custom.SyntaxNode, nil, 0)

	fragment := CopyAsFragment(custom)
	if fragment == custom || fragment.SyntaxNode == custom.SyntaxNode {
		t.Fatal("fragment reused original custom node")
	}
	if fragment.Tree().Original != original || fragment.GetName(0) != "Value" || !fragment.IsOptional(1) || fragment.CompletionHint(0) != CompletionHintExpression {
		t.Fatalf("custom fragment = %#v", fragment)
	}
	if got := fragment.GetOriginalNode(); got != custom.SyntaxNode {
		t.Fatalf("custom original node = %#v, want %#v", got, custom.SyntaxNode)
	}
	if diagnostics := fragment.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS006" {
		t.Fatalf("custom fragment diagnostics = %#v", diagnostics)
	}
}

func TestWithDiagnosticsClonesAndPreservesGeneratedWrappers(t *testing.T) {
	query := ParseQuery("T | where X > 1")
	where := GetFirstDescendant[*FilterOperator](query)
	if where == nil {
		t.Fatal("missing filter operator")
	}
	first := NewDiagnostic("KS900", "first")
	second := NewDiagnostic("KS901", "second")

	withFirst, ok := WithDiagnostics(where, first).(*FilterOperator)
	if !ok {
		t.Fatalf("WithDiagnostics type = %T, want *FilterOperator", WithDiagnostics(where, first))
	}
	if withFirst == where || withFirst.SyntaxNode == where.SyntaxNode || withFirst.Condition() == where.Condition() {
		t.Fatal("WithDiagnostics reused part of the original generated node")
	}
	if got := withFirst.Diagnostics(); !reflect.DeepEqual(got, []Diagnostic{first}) {
		t.Fatalf("first diagnostics = %#v", got)
	}
	if got := where.Diagnostics(); len(got) != 0 {
		t.Fatalf("original diagnostics = %#v, want none", got)
	}

	replaced, ok := WithDiagnostics(withFirst, second).(*FilterOperator)
	if !ok || !reflect.DeepEqual(replaced.Diagnostics(), []Diagnostic{second}) {
		t.Fatalf("replaced diagnostics = %T %#v", replaced, replaced.Diagnostics())
	}
	additional, ok := WithAdditionalDiagnostics(withFirst, second).(*FilterOperator)
	if !ok || !reflect.DeepEqual(additional.Diagnostics(), []Diagnostic{first, second}) {
		t.Fatalf("additional diagnostics = %T %#v", additional, additional.Diagnostics())
	}
	without, ok := WithDiagnostics(withFirst).(*FilterOperator)
	if !ok || !reflect.DeepEqual(without.Diagnostics(), []Diagnostic{first}) {
		t.Fatalf("empty replacement diagnostics = %T %#v", without, without.Diagnostics())
	}
}

func TestWithDiagnosticsSupportsTokensListsSeparatedElementsAndCustomNodes(t *testing.T) {
	first := NewDiagnostic("KS900", "first")
	second := NewDiagnostic("KS901", "second")
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"), first)
	list := NewSyntaxListWithDiagnostics([]SyntaxElement{token.Clone(true)}, first)
	separated := NewSeparatedElementWithDiagnostics(token.Clone(true), nil, first)
	custom, err := NewCustomNodeWithShape(
		[]CustomElementDescriptor{NewCustomElementDescriptor("Value", CompletionHintExpression, false)},
		[]SyntaxElement{token.Clone(true)},
		first,
	)
	if err != nil {
		t.Fatalf("NewCustomNodeWithShape: %v", err)
	}

	for _, test := range []struct {
		name    string
		element SyntaxElement
	}{
		{name: "token", element: token},
		{name: "list", element: list},
		{name: "separated", element: separated},
		{name: "custom", element: custom},
	} {
		t.Run(test.name, func(t *testing.T) {
			clone := WithDiagnostics(test.element, second)
			if clone == test.element || !reflect.DeepEqual(clone.Diagnostics(), []Diagnostic{second}) {
				t.Fatalf("clone diagnostics = %#v", clone.Diagnostics())
			}
			if !reflect.DeepEqual(test.element.Diagnostics(), []Diagnostic{first}) {
				t.Fatalf("original diagnostics = %#v", test.element.Diagnostics())
			}
			if test.name == "custom" && clone.(*CustomNode).GetName(0) != "Value" {
				t.Fatalf("custom shape = %#v", clone.(*CustomNode).Shape())
			}
		})
	}
}

func TestSyntaxDiagnosticStatePropertiesCoverListsAndGeneratedNodes(t *testing.T) {
	diagnostic := NewDiagnostic("KS900", "test diagnostic")
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"), diagnostic)
	list := NewSyntaxList(token.Clone(true))
	separated := NewSeparatedElement(token.Clone(true), nil)
	query := ParseQuery("T | where X > 1")
	where := GetFirstDescendant[*FilterOperator](query)
	if where == nil {
		t.Fatal("missing filter operator")
	}
	withDiagnostic, ok := WithDiagnostics(where, diagnostic).(*FilterOperator)
	if !ok {
		t.Fatalf("WithDiagnostics type = %T, want *FilterOperator", WithDiagnostics(where, diagnostic))
	}

	for _, test := range []struct {
		name     string
		element  SyntaxElement
		contains bool
		has      bool
		missing  bool
	}{
		{name: "token", element: token, contains: true, has: true},
		{name: "list descendant", element: list, contains: true},
		{name: "separated descendant", element: separated, contains: true},
		{name: "generated root", element: withDiagnostic, contains: true, has: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.element.ContainsSyntaxDiagnostics() != test.contains || test.element.HasSyntaxDiagnostics() != test.has || test.element.IsMissing() != test.missing {
				t.Fatalf("diagnostic state = contains:%t has:%t missing:%t", test.element.ContainsSyntaxDiagnostics(), test.element.HasSyntaxDiagnostics(), test.element.IsMissing())
			}
		})
	}
}
