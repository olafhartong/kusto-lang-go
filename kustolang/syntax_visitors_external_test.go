package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDefaultSyntaxVisitorForwardsAndOverrides(t *testing.T) {
	query := kustolang.ParseQuery(`datatable(A:long)[1] | where A > 0`)
	var defaults []kustolang.SyntaxKind
	var filters int
	visitor := kustolang.NewDefaultSyntaxVisitor(func(element kustolang.SyntaxElement) {
		defaults = append(defaults, element.Kind())
	})
	visitor.VisitFilterOperatorFunc = func(*kustolang.FilterOperator) {
		filters++
	}

	kustolang.AcceptSyntaxVisitor(query, visitor)
	var filter *kustolang.FilterOperator
	query.WalkElements(func(element kustolang.SyntaxElement) {
		if typed, ok := kustolang.WrapGeneratedSyntaxElement(element).(*kustolang.FilterOperator); ok {
			filter = typed
		}
	})
	if filter == nil {
		t.Fatal("valid query has no filter operator")
	}
	kustolang.AcceptSyntaxVisitor(filter, visitor)
	if len(defaults) != 1 || defaults[0] != kustolang.SyntaxKindQueryBlock || filters != 1 {
		t.Fatalf("visitor dispatch = defaults:%v filters:%d", defaults, filters)
	}
}

func TestPublicDefaultSyntaxVisitorResultHandlesMalformedTree(t *testing.T) {
	malformed := kustolang.ParseQuery(`print value=1 +`)
	if malformed == nil || !malformed.ContainsSyntaxDiagnostics() {
		t.Fatalf("malformed query = %#v", malformed)
	}

	visitor := kustolang.NewDefaultSyntaxVisitorResult(func(element kustolang.SyntaxElement) kustolang.SyntaxKind {
		return element.Kind()
	})
	visitor.VisitQueryBlockFunc = func(*kustolang.QueryBlock) kustolang.SyntaxKind {
		return kustolang.SyntaxKindQueryBlock
	}
	if got := kustolang.AcceptSyntaxVisitorResult(malformed, visitor); got != kustolang.SyntaxKindQueryBlock {
		t.Fatalf("malformed visitor result = %v", got)
	}

	var missingNode kustolang.SyntaxElement
	malformed.WalkElements(func(element kustolang.SyntaxElement) {
		if missingNode == nil && element.IsMissing() {
			missingNode = element
		}
	})
	if missingNode == nil {
		t.Fatal("malformed query has no missing recovery element")
	}
}
