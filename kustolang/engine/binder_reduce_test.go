package kustolang

import "testing"

func TestNativeReduceRequiresPipeAndChecksExpressionType(t *testing.T) {
	first := ParseQuery("reduce by 'x'")
	NewBinder(nil).TryBind(first)
	assertDiagnosticCodes(t, firstBoundNode[*ReduceByOperator](t, first), "KS174")

	table := NewTableSymbol("T",
		NewColumnSymbol("Text", ScalarTypes.String.Type()),
		NewColumnSymbol("Payload", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("Number", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery("T | reduce by Text")
	NewBinder(globals).TryBind(query)
	assertDiagnosticCodes(t, firstBoundNode[*ReduceByOperator](t, query))

	for _, column := range []string{"Payload", "Number"} {
		query := ParseQuery("T | reduce by " + column)
		NewBinder(globals).TryBind(query)
		assertDiagnosticCodes(t, firstBoundNode[*ReduceByOperator](t, query), "KS141")
	}
}

func TestNativeReduceValidatesLeadingAndWithParameters(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Text", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery("T | reduce kind=SOURCE by Text with threshold='high', characters=1")
	NewBinder(globals).TryBind(query)
	reduce := firstBoundNode[*ReduceByOperator](t, query)
	assertDiagnosticCodes(t, reduce, "KS005", "KS224", "KS222")

	if _, ok := reduce.Parameters().At(0).Name().ReferencedSymbol().(*QueryOperatorParameter); !ok {
		t.Fatal("leading parameter was not bound to its declaration")
	}
	for _, item := range reduce.With().Parameters().Elements() {
		if _, ok := item.Element().Name().ReferencedSymbol().(*QueryOperatorParameter); !ok {
			t.Fatal("with parameter was not bound to its declaration")
		}
	}
}

func TestNativeReduceProducesFreshClosedSourceShapeWithoutState(t *testing.T) {
	declared := NewColumnSymbol("Declared", ScalarTypes.String.Type())
	table := NewTableSymbol("T", declared).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery("T | project Declared, Future | reduce kind=source by Declared")
	binder := NewBinder(globals)
	binder.TryBind(query)

	result := semanticTable(t, firstBoundNode[*ReduceByOperator](t, query))
	assertOrderedColumns(t, result, "Declared", "Future", "Pattern", "Count", "Representative")
	if result.IsOpen() || result.IsSerialized() || result.IsSorted() {
		t.Fatalf("reduce result state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}

	other := ParseQuery("T | reduce by Declared")
	binder.TryBind(other)
	otherResult := semanticTable(t, firstBoundNode[*ReduceByOperator](t, other))
	if otherResult == result {
		t.Fatal("separate reduce operators shared a result table")
	}
	assertOrderedColumns(t, otherResult, "Pattern", "Count", "Representative")
}

func TestNativeReduceAppendsLiteralColumnsIncludingDuplicateNames(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Pattern", ScalarTypes.Long.Type()),
		NewColumnSymbol("Count", ScalarTypes.String.Type()),
		NewColumnSymbol("Representative", ScalarTypes.Dynamic.Type()))
	query := ParseQuery("T | reduce kind=source by tostring(Pattern)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	columns := semanticTable(t, firstBoundNode[*ReduceByOperator](t, query)).Columns()
	if len(columns) != 6 {
		t.Fatalf("reduce columns = %#v", columns)
	}
	wantNames := []string{"Pattern", "Count", "Representative", "Pattern", "Count", "Representative"}
	wantTypes := []*TypeSymbol{ScalarTypes.Long.Type(), ScalarTypes.String.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.String.Type(), ScalarTypes.Long.Type(), ScalarTypes.String.Type()}
	for i := range columns {
		if columns[i].SymbolName() != wantNames[i] || !sameType(columns[i].Type, wantTypes[i]) {
			t.Fatalf("column %d = %s:%s, want %s:%s", i, columns[i].SymbolName(), columns[i].Type.SymbolName(), wantNames[i], wantTypes[i].SymbolName())
		}
	}
}

func TestNativeReduceResultColumnsRetainExpressionSourceAcrossRecovery(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Text", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name      string
		text      string
		wantCodes []string
	}{
		{name: "valid", text: "T | reduce by Text"},
		{name: "dynamic", text: "T | reduce by dynamic({})", wantCodes: []string{"KS141"}},
		{name: "missing expression", text: "T | reduce by", wantCodes: []string{"KS006"}},
		{name: "unknown name", text: "T | reduce by Missing", wantCodes: []string{"KS142"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)
			reduce := firstBoundNode[*ReduceByOperator](t, query)
			if !reduce.IsBound() {
				t.Fatal("reduce operator was not bound")
			}
			result := semanticTable(t, reduce)
			assertOrderedColumns(t, result, "Pattern", "Count", "Representative")
			for _, column := range result.Columns() {
				if column.Source != reduce.Expression() {
					t.Fatalf("%s source = %#v, want reduce expression %#v", column.SymbolName(), column.Source, reduce.Expression())
				}
			}
			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			if len(diagnostics) != len(test.wantCodes) {
				t.Fatalf("diagnostics = %#v, want %v", diagnostics, test.wantCodes)
			}
			for i, code := range test.wantCodes {
				if diagnostics[i].Code != code {
					t.Fatalf("diagnostic[%d] = %#v, want %s", i, diagnostics[i], code)
				}
			}
		})
	}
}
