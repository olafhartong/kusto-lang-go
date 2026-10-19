package kustolang

import "testing"

func TestNativeMvExpandStagesFunctionCallAndBindsDownstream(t *testing.T) {
	bag := NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type())
	keep := NewColumnSymbol("Keep", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", bag, keep)
	query := ParseQuery(`T | mv-expand Key=bag_keys(Bag) | project Key, Keep`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	expand := firstBoundNode[*MvExpandOperator](t, query)
	assertDiagnosticCodes(t, expand)
	call := firstBoundNode[*FunctionCallExpression](t, expand)
	if !call.IsBound() || call.ReferencedSymbol() != Functions.BagKeys {
		t.Fatalf("bag_keys binding = bound:%t symbol:%#v", call.IsBound(), call.ReferencedSymbol())
	}
	result := semanticTable(t, expand)
	assertOrderedColumns(t, result, "Bag", "Keep", "Key")
	if key := result.GetColumn("Key"); key == nil || key.Type != ScalarTypes.DynamicString.Type() {
		t.Fatalf("expanded Key = %#v, want dynamic(string)", key)
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	if !project.IsBound() {
		t.Fatal("project downstream of function-call mv-expand was not bound")
	}
	assertOrderedColumns(t, semanticTable(t, project), "Key", "Keep")
}

func TestNativeMvExpandGeneralScalarDiagnosticsAndTypes(t *testing.T) {
	text := NewColumnSymbol("Text", ScalarTypes.String.Type())
	items := NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type())
	table := NewTableSymbol("T", text, items)
	query := ParseQuery(`T | mv-expand X=tostring(Text), Y=Items[0], Z=Items + Items | count`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	expand := firstBoundNode[*MvExpandOperator](t, query)
	if !expand.IsBound() {
		t.Fatal("diagnosed general-scalar mv-expand was not bound")
	}
	diagnostics := expand.SemanticDiagnostics()
	if !diagnosticsContain(diagnostics, "KS141") {
		t.Fatalf("mv-expand diagnostics = %#v, want KS141", diagnostics)
	}
	result := semanticTable(t, expand)
	assertOrderedColumns(t, result, "Text", "Items", "X", "Y", "Z")
	if result.GetColumn("X").Type != ScalarTypes.String.Type() {
		t.Fatalf("non-dynamic X type = %v, want string", result.GetColumn("X").Type)
	}
	if result.GetColumn("Y").Type != ScalarTypes.Dynamic.Type() {
		t.Fatalf("re-expanded dynamic element Y type = %v, want dynamic", result.GetColumn("Y").Type)
	}
	if !firstBoundNode[*CountOperator](t, query).IsBound() {
		t.Fatal("count downstream of diagnosed mv-expand was not bound")
	}
}

func TestNativeMvExpandMalformedScalarRollsBack(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type())).WithIsOpen(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	recovered := ParseQuery(`T | mv-expand Good=bag_keys(Bag), Bad=bag_keys(Bag`)
	NewBinder(globals).TryBind(recovered)
	if expand := firstBoundNode[*MvExpandOperator](t, recovered); !expand.IsBound() ||
		!diagnosticsContainMessage(expand.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("recovered mv-expand call = bound:%t diagnostics:%#v", expand.IsBound(), expand.ContainedDiagnostics())
	}

	query := ParseQuery(`T | mv-expand Good=bag_keys(Bag), Bad=bag_keys(Bag)`)
	expand := firstBoundNode[*MvExpandOperator](t, query)
	calls := boundNodes[*FunctionCallExpression](expand)
	calls[len(calls)-1].ArgumentList().OpenParen().text = ""
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 29
	tree.bind(query)

	if expand.IsBound() {
		t.Fatal("mv-expand with injected invalid later function call was bound")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 29 {
		t.Fatal("malformed mv-expand scalar changed binder or tree state")
	}
	if binder.openTableInferredColumns != nil || binder.openTableInferredOrder != nil {
		t.Fatal("malformed mv-expand scalar leaked inferred columns")
	}
	for _, call := range boundNodes[*FunctionCallExpression](expand) {
		if call.IsBound() {
			t.Fatal("malformed mv-expand retained staged function semantics")
		}
	}
}

func TestNativeMvExpandListPredicatesUseOrdinaryScalarBinding(t *testing.T) {
	items := NewColumnSymbol("Items", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", items)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name       string
		expression string
	}{
		{name: "in", expression: "X=Items in (1, 2)"},
		{name: "has any", expression: "X=Items has_any (1, 2)"},
		{name: "has all", expression: "X=Items has_all (1, 2)"},
		{name: "between", expression: "X=Items between (1 .. 2)"},
	} {
		t.Run(test.name, func(t *testing.T) {
			text := "T | mv-expand " + test.expression + " | project X"
			query := ParseQuery(text)
			NewBinder(globals).TryBind(query)
			expand := firstBoundNode[*MvExpandOperator](t, query)
			if !expand.IsBound() {
				t.Fatalf("mv-expand was not bound: %#v", expand.ContainedDiagnostics())
			}
			assertDiagnosticLocations(t, expand.SemanticDiagnostics(),
				diagnosticLocation{"KS141", len("T | mv-expand "), len(test.expression)})
			if column := semanticTable(t, expand).GetColumn("X"); column == nil || column.Type != ScalarTypes.Bool.Type() {
				t.Fatalf("expanded X = %#v, want bool", column)
			}
			if !firstBoundNode[*ProjectOperator](t, query).IsBound() {
				t.Fatal("downstream project was not bound")
			}
		})
	}
}

func TestNativeMvExpandListPredicateMissingCloseStillBinds(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Items", ScalarTypes.Dynamic.Type()))
	text := "T | mv-expand X=Items in (1, 2 | project X"
	query := ParseQuery(text)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	expand := firstBoundNode[*MvExpandOperator](t, query)
	if !expand.IsBound() {
		t.Fatalf("mv-expand was not bound: %#v", expand.ContainedDiagnostics())
	}
	assertDiagnosticLocations(t, expand.SemanticDiagnostics(),
		diagnosticLocation{"KS141", len("T | mv-expand "), len("X=Items in (1, 2")})
	if !diagnosticsContainMessage(expand.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing-close diagnostics = %#v", expand.ContainedDiagnostics())
	}
}
