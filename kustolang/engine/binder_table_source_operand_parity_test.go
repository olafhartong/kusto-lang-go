package kustolang

import "testing"

func TestNativeTableSourceOperandsUnionFunctionCallsAndQualifiedPath(t *testing.T) {
	u1 := NewTableSymbol("U1", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	u2 := NewTableSymbol("U2", NewColumnSymbol("B", ScalarTypes.String.Type()))
	u3 := NewTableSymbol("U3", NewColumnSymbol("C", ScalarTypes.Bool.Type()))
	database := NewDatabaseSymbol("db", u1, u2, u3)
	query := ParseQuery(`union table("U1"), table("U2"), database("db").U3, database("db").table("U3") | project A, B, C`)
	NewBinder(NewGlobalState(database)).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertDiagnosticCodes(t, union)
	assertOrderedColumns(t, semanticTable(t, union), "A", "B", "C")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "A", "B", "C")
	for _, separated := range union.Expressions().Elements() {
		if operand := separated.Element(); !expressionIsBound(operand) || unionExpressionTable(operand) == nil {
			t.Fatalf("union operand = %T, bound:%t table:%#v", operand, expressionIsBound(operand), unionExpressionTable(operand))
		}
	}
}

func TestNativeTableSourceOperandsJoinAndLookupCalls(t *testing.T) {
	left := NewTableSymbol("Left",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("LeftOnly", ScalarTypes.String.Type()),
	)
	right := NewTableSymbol("Right",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	joinQuery := ParseQuery(`Left | join table("Right") on Key | project LeftOnly, RightOnly`)
	NewBinder(globals).TryBind(joinQuery)
	join := firstBoundNode[*JoinOperator](t, joinQuery)
	assertDiagnosticCodes(t, join)
	assertOrderedColumns(t, semanticTable(t, join), "Key", "LeftOnly", "Key1", "RightOnly")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, joinQuery)), "LeftOnly", "RightOnly")

	lookupQuery := ParseQuery(`Left | lookup database("db").table("Right") on Key | project Key, LeftOnly, RightOnly`)
	NewBinder(globals).TryBind(lookupQuery)
	lookup := firstBoundNode[*LookupOperator](t, lookupQuery)
	assertDiagnosticCodes(t, lookup)
	assertOrderedColumns(t, semanticTable(t, lookup), "Key", "LeftOnly", "RightOnly")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, lookupQuery)), "Key", "LeftOnly", "RightOnly")
}

func TestNativeTableSourceOperandsBindWithoutIncomingRowScopeAndKeepChildDiagnostics(t *testing.T) {
	leftColumn := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	left := NewTableSymbol("Left", leftColumn)
	right := NewTableSymbol("Rows", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	rows := NewFunctionSymbol("Rows", NewSymbolSignature(right, NewParameter("value", ScalarTypes.Long.Type())))
	baseGlobals := NewGlobalState(NewDatabaseSymbol("db", left))
	functions := append(append([]*FunctionSymbol(nil), baseGlobals.Functions...), rows)
	globals := baseGlobals.WithFunctions(functions)

	query := ParseQuery(`Left | join Rows(Key) on Key | count`)
	NewBinder(globals).TryBind(query)
	argument := namedFunctionCall(t, query, "Rows").ArgumentList().Expressions().At(0).Element()
	if expressionReferencedSymbol(argument) == leftColumn {
		t.Fatal("join RHS function argument resolved against the incoming left row scope")
	}
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(firstBoundNode[*JoinOperator](t, query), DiagnosticsIncludeDefault); !diagnosticsContain(diagnostics, "KS142") {
		t.Fatalf("join diagnostics = %#v, want child KS142", diagnostics)
	}
	if !firstBoundNode[*CountOperator](t, query).IsBound() {
		t.Fatal("downstream count was not bound after a diagnosed table-valued call")
	}

	unknown := ParseQuery(`union Rows(MissingRows()) | count`)
	NewBinder(globals).TryBind(unknown)
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(firstBoundNode[*UnionOperator](t, unknown), DiagnosticsIncludeDefault); !diagnosticsContain(diagnostics, "KS211") {
		t.Fatalf("union diagnostics = %#v, want child KS211", diagnostics)
	}
	if !firstBoundNode[*CountOperator](t, unknown).IsBound() {
		t.Fatal("downstream count was not bound after an unresolved function call")
	}
}

func TestNativeTableSourceOperandsDiagnoseBoundScalars(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", left))
	for _, text := range []string{
		`union strlen("x") | count`,
		`Left | join strlen("x") on Key | count`,
		`Left | lookup strlen("x") on Key | count`,
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(globals).TryBind(query)
			diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault)
			if !diagnosticsContain(diagnostics, "KS111") {
				t.Fatalf("diagnostics = %#v, want KS111", diagnostics)
			}
			if !firstBoundNode[*CountOperator](t, query).IsBound() {
				t.Fatal("downstream count was not bound after scalar operand diagnostic")
			}
		})
	}
}

func TestNativeTableSourceOperandsMalformedCallsAndPathsRollback(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	for _, text := range []string{
		`union table(`,
		`T | join table( on Key | count`,
		`T | lookup database("db").table( on Key | count`,
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
			outer := NewTableSymbol("Outer", NewColumnSymbol("Outer", ScalarTypes.Bool.Type()))
			binder.rowScope = outer
			binder.scopeKind = ScopeKindAggregate
			local := binder.LocalScope
			tree := newTreeBinder(binder)
			tree.inPipe = true
			tree.inWildcard = true
			tree.bind(query)

			if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
				!tree.inPipe || !tree.inWildcard {
				t.Fatal("malformed table source changed binder state")
			}
			for _, call := range boundNodes[*FunctionCallExpression](query) {
				if call.IsBound() {
					t.Fatal("malformed table source committed function-call semantics")
				}
			}
			for _, count := range boundNodes[*CountOperator](query) {
				if count.IsBound() {
					t.Fatal("malformed table source bound a downstream operator")
				}
			}
		})
	}
}
