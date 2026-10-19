package kustolang

import "testing"

func TestNativeMvExpandRowLimitBindsGeneralScalarsAndDownstream(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	for _, test := range []struct {
		name string
		text string
	}{
		{name: "arithmetic", text: `T | mv-expand Items limit 1 + 2 * 3 | count`},
		{name: "nested functions", text: `T | mv-expand Items limit tolong(toreal(2) + 1.0) | count`},
		{name: "toscalar", text: `T | mv-expand Items limit toscalar(1 + 2) | count`},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			expand := firstBoundNode[*MvExpandOperator](t, query)
			if !expand.IsBound() {
				t.Fatalf("mv-expand was not bound: %#v", expand.ContainedDiagnostics())
			}
			if !firstBoundNode[*CountOperator](t, query).IsBound() {
				t.Fatal("downstream count was not bound")
			}
		})
	}
}

func TestNativeMvExpandRowLimitReportsNonIntegerScalar(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type()),
		NewColumnSymbol("RealLimit", ScalarTypes.Real.Type()))
	query := ParseQuery(`T | mv-expand Items limit RealLimit + 0.5 | count`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	expand := firstBoundNode[*MvExpandOperator](t, query)
	if !expand.IsBound() || !diagnosticsContain(expand.SemanticDiagnostics(), "KS134") {
		t.Fatalf("mv-expand binding = bound:%t diagnostics:%#v, want KS134", expand.IsBound(), expand.SemanticDiagnostics())
	}
	if !firstBoundNode[*CountOperator](t, query).IsBound() {
		t.Fatal("downstream count was not bound")
	}
}

func TestNativeMvExpandRowLimitBindsParserRecovery(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	for _, test := range []struct {
		name string
		text string
	}{
		{name: "missing operand", text: `T | mv-expand Items limit 1 + | count`},
		{name: "missing function delimiter", text: `T | mv-expand Items limit tolong(1 + 2 | count`},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			expand := firstBoundNode[*MvExpandOperator](t, query)
			if !expand.IsBound() {
				t.Fatalf("parser-recovered mv-expand was not bound: %#v", expand.ContainedDiagnostics())
			}
			if !firstBoundNode[*CountOperator](t, query).IsBound() {
				t.Fatal("downstream count was not bound")
			}
		})
	}
}

func TestNativeMvExpandRowLimitRejectsInjectedInvalidDelimiterTransactionally(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type())).WithIsOpen(true)
	query := ParseQuery(`T | mv-expand Items limit tolong(1 + 2) | count`)
	expand := firstBoundNode[*MvExpandOperator](t, query)
	firstBoundNode[*FunctionCallExpression](t, expand).ArgumentList().CloseParen().text = ""

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 41
	tree.bind(query)

	if expand.IsBound() {
		t.Fatal("mv-expand with an injected invalid row-limit delimiter was bound")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 41 {
		t.Fatal("invalid row limit changed binder or tree state")
	}
	if binder.openTableInferredColumns != nil || binder.openTableInferredOrder != nil {
		t.Fatal("invalid row limit leaked inferred columns")
	}
	for _, bound := range boundNodes[*NameReference](query) {
		if bound.IsBound() {
			t.Fatal("invalid row limit retained staged name-reference semantics")
		}
	}
}

func TestNativeMvApplyRemainsCompatibleWithSharedScalarCompleteness(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type()))
	query := ParseQuery(`T | mv-apply Value=Items limit 1 + 2 on (summarize Total=sum(Value)) | count`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	if !apply.IsBound() || !firstBoundNode[*CountOperator](t, query).IsBound() {
		t.Fatalf("mv-apply compatibility = bound:%t diagnostics:%#v", apply.IsBound(), apply.ContainedDiagnostics())
	}
}
