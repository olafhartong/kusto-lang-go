package kustolang

import "testing"

func TestNativeTreeBinderDataTableValidEmptyDynamicAndDownstream(t *testing.T) {
	query := ParseQuery(`datatable (Account:string, Count:long, Payload:dynamic) ["a", 1, dynamic({"x":[1,-2]})] | project Payload, Account`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	dataTable := firstBoundNode[*DataTableExpression](t, query)
	result := semanticTable(t, dataTable)
	assertOrderedColumns(t, result, "Account", "Count", "Payload")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Account": ScalarTypes.String.Type(),
		"Count":   ScalarTypes.Long.Type(),
		"Payload": ScalarTypes.Dynamic.Type(),
	})
	if result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("datatable result state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	dynamic := dataTable.Values().At(2).Element().(*DynamicExpression)
	if !dynamic.IsBound() || !sameType(ExpressionResultType(dynamic), ScalarTypes.Dynamic.Type()) {
		info, _ := dynamic.GetSemanticInfo()
		t.Fatalf("dynamic value semantic info = %#v", info)
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, project), "Payload", "Account")

	empty := parsedQueryExpression(t, `datatable () []`).(*DataTableExpression)
	NewBinder(nil).TryBind(empty)
	if !empty.IsBound() || len(semanticTable(t, empty).Columns()) != 0 || len(empty.SemanticDiagnostics()) != 0 {
		t.Fatalf("empty datatable = bound:%t table:%#v diagnostics:%#v", empty.IsBound(), semanticTable(t, empty), empty.SemanticDiagnostics())
	}
}

func TestNativeTreeBinderDataTableColumnsDuplicatesCaseAndInvalidTypes(t *testing.T) {
	dataTable := parsedQueryExpression(t, `datatable (A:long, A:string, a:bool, Nested:(X:long)) [1, true]`).(*DataTableExpression)
	NewBinder(nil).TryBind(dataTable)

	result := semanticTable(t, dataTable)
	assertOrderedColumns(t, result, "A", "a")
	assertColumnTypes(t, result, map[string]*TypeSymbol{"A": ScalarTypes.Long.Type(), "a": ScalarTypes.Bool.Type()})
	assertDiagnosticCodes(t, dataTable, "KS171", "KS170")
	declarations := dataTable.Schema().Columns()
	if declarations.At(0).Element().Name().ReferencedSymbol() != result.Columns()[0] ||
		declarations.At(2).Element().Name().ReferencedSymbol() != result.Columns()[1] {
		t.Fatal("successful datatable declarations were not bound to result columns")
	}
	if declarations.At(1).Element().Name().IsBound() || declarations.At(3).Element().Name().IsBound() {
		t.Fatal("omitted datatable declarations unexpectedly received column symbols")
	}
}

func TestNativeTreeBinderDataTableCountScalarAndNoCellConversion(t *testing.T) {
	count := parsedQueryExpression(t, `datatable (A:long, B:string) [1, "x", 2]`).(*DataTableExpression)
	NewBinder(nil).TryBind(count)
	assertDiagnosticCodes(t, count, "KS173")

	runtimeMismatch := parsedQueryExpression(t, `datatable (A:long, B:bool) ["not a long", 99]`).(*DataTableExpression)
	NewBinder(nil).TryBind(runtimeMismatch)
	if diagnostics := runtimeMismatch.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("datatable checked declared cell conversions: %#v", diagnostics)
	}

	tabular := parsedQueryExpression(t, `datatable (A:long) [1]`).(*DataTableExpression)
	rangeValue := cloneSyntaxElementPreservingGeneratedWrapper(
		parsedQueryExpression(t, `range X from 1 to 1 step 1`), true,
	).(*RangeOperator)
	tabular.Values().SyntaxList.elements[0] = NewSeparatedElement(rangeValue, nil)
	NewBinder(nil).TryBind(tabular)
	if !rangeValue.IsBound() {
		t.Fatal("tabular datatable cell was not bound")
	}
	assertDiagnosticCodes(t, tabular, "KS108")
}

func TestNativeTreeBinderDataTableParametersAndRecovery(t *testing.T) {
	parameterized := parsedQueryExpression(t, `datatable hint.materialized=true (A:long) [1]`).(*DataTableExpression)
	NewBinder(nil).TryBind(parameterized)
	if !simpleBinderSyntaxNode(parameterized.Parameters().At(0).Expression()).IsBound() {
		t.Fatal("datatable parameter value was not bound")
	}
	assertDiagnosticCodes(t, parameterized, "KS179")

	for _, text := range []string{
		`datatable (A:long) [,1,]`,
		`datatable (A:long) [,,1]`,
		`datatable (A:long [1`,
	} {
		t.Run(text, func(t *testing.T) {
			dataTable := firstParsedQueryExpression(t, text).(*DataTableExpression)
			NewBinder(nil).TryBind(dataTable)
			if !dataTable.IsBound() {
				t.Fatal("recovered datatable was not bound")
			}
		})
	}
}

func TestNativeTreeBinderDataTableFreshAndIdempotent(t *testing.T) {
	first := parsedQueryExpression(t, `datatable (A:long) [1]`).(*DataTableExpression)
	second := parsedQueryExpression(t, `datatable (A:long) [1]`).(*DataTableExpression)
	binder := NewBinder(nil)
	binder.TryBind(first)
	firstTable := semanticTable(t, first)
	binder.TryBind(first)
	if semanticTable(t, first) != firstTable {
		t.Fatal("repeated datatable binding replaced its committed table")
	}
	binder.TryBind(second)
	if semanticTable(t, second) == firstTable {
		t.Fatal("separate datatable expressions shared a result table")
	}
}

func TestNativeTreeBinderDataTableCorruptionAndRollback(t *testing.T) {
	for _, mutate := range []func(*DataTableExpression){
		func(node *DataTableExpression) { node.SyntaxNode.children[0] = nil },
		func(node *DataTableExpression) {
			node.Values().SyntaxList.elementType = "SeparatedElement<SyntaxElement>"
		},
		func(node *DataTableExpression) { node.Values().SyntaxList.elements[0] = node.Values().At(0).Element() },
		func(node *DataTableExpression) {
			dynamic := node.Values().At(0).Element().(*DynamicExpression)
			dynamic.Expression().(*JsonObjectExpression).SyntaxNode.children[1] = nil
		},
	} {
		dataTable := parsedQueryExpression(t, `datatable (Payload:dynamic) [dynamic({"x":1})]`).(*DataTableExpression)
		mutate(dataTable)
		binder := NewBinder(nil)
		outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type())).WithIsOpen(true)
		binder.rowScope = outer
		binder.scopeKind = ScopeKindAggregate
		local := binder.LocalScope
		existing := binder.getOpenTableInferredColumn("Existing", outer)
		tree := newTreeBinder(binder)
		tree.inPipe = true
		tree.inWildcard = true
		tree.nodeBinder.inPipe = true
		tree.nodeBinder.schemaCannotBeEmpty = true
		tree.nodeBinder.defaultColumnNameSuffix = 41
		tree.nodeBinder.pipedAsRowScope = outer
		tree.nodeBinder.joinRightRowScope = outer
		tree.nodeBinder.lookupRightRowScope = outer
		tree.bind(dataTable)

		if dataTable.IsBound() {
			t.Fatal("structurally corrupt datatable was bound")
		}
		if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
			!tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe || !tree.nodeBinder.schemaCannotBeEmpty ||
			tree.nodeBinder.defaultColumnNameSuffix != 41 || tree.nodeBinder.pipedAsRowScope != outer ||
			tree.nodeBinder.joinRightRowScope != outer || tree.nodeBinder.lookupRightRowScope != outer {
			t.Fatal("datatable preflight changed binder or tree state")
		}
		if binder.openTableInferredColumns[outer]["Existing"] != existing ||
			len(binder.openTableInferredOrder[outer]) != 1 || binder.openTableInferredOrder[outer][0] != existing {
			t.Fatal("datatable preflight changed open-table inference state")
		}
		for _, syntax := range boundNodes[*SyntaxNode](dataTable) {
			if syntax.IsBound() {
				t.Fatalf("datatable preflight committed child semantics: %v", syntax.Kind())
			}
		}
	}
}
