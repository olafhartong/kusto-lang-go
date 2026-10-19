package kustolang

import "testing"

func TestNativeRenderRequiresPipeAndValidatesExactChartCatalog(t *testing.T) {
	first := ParseQuery("render timechart")
	NewBinder(NewGlobalState(nil)).TryBind(first)
	assertDiagnosticCodes(t, firstBoundNode[*RenderOperator](t, first), "KS174")

	table := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, chartType := range KustoFacts.ChartTypes {
		query := ParseQuery("T | render " + chartType)
		NewBinder(globals).TryBind(query)
		assertDiagnosticCodes(t, firstBoundNode[*RenderOperator](t, query))
	}

	unknown := ParseQuery("T | render custom_visual")
	NewBinder(globals).TryBind(unknown)
	assertDiagnosticCodes(t, firstBoundNode[*RenderOperator](t, unknown), "KS005")
}

func TestNativeRenderValidatesLegacyAndWithParameterCatalogs(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	valid := ParseQuery("T | render timechart kind=stacked title='legacy' accumulate=true with (xcolumn=Timestamp, ycolumns=Value, yaxis=log)")
	NewBinder(globals).TryBind(valid)
	render := firstBoundNode[*RenderOperator](t, valid)
	assertDiagnosticCodes(t, render)
	for _, parameter := range render.Parameters().Elements() {
		if _, ok := parameter.Name().ReferencedSymbol().(*QueryOperatorParameter); !ok {
			t.Fatalf("legacy parameter %q was not bound to its declaration", NameDeclarationSimpleName(parameter.Name()))
		}
	}
	for _, separated := range render.WithClause().Properties().Elements() {
		parameter := separated.Element()
		if _, ok := parameter.Name().ReferencedSymbol().(*QueryOperatorParameter); !ok {
			t.Fatalf("with parameter %q was not bound to its declaration", NameDeclarationSimpleName(parameter.Name()))
		}
	}

	invalid := ParseQuery("T | render timechart with (kind=invalid)")
	NewBinder(globals).TryBind(invalid)
	assertDiagnosticCodes(t, firstBoundNode[*RenderOperator](t, invalid), "KS005")
}

func TestNativeRenderPreservesRowsStateSuppressesMissingCascadeAndIsIdempotent(t *testing.T) {
	declared := NewColumnSymbol("Declared", ScalarTypes.String.Type())
	table := NewTableSymbol("T", declared).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery("T | project Declared, Future | render")
	binder := NewBinder(globals)
	binder.TryBind(query)

	render := firstBoundNode[*RenderOperator](t, query)
	result := semanticTable(t, render)
	assertOrderedColumns(t, result, "Declared", "Future")
	if result.GetColumn("Declared") != declared || !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("render result/state = %#v open:%t serialized:%t sorted:%t", result.Columns(), result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}
	if diagnosticsContainCode(render.ContainedDiagnostics(), "KS140") {
		t.Fatalf("missing chart token caused a binder cascade: %#v", render.ContainedDiagnostics())
	}
	if !diagnosticsContainCode(render.ContainedDiagnostics(), "KS005") {
		t.Fatalf("missing chart parser diagnostic was lost: %#v", render.ContainedDiagnostics())
	}

	binder.TryBind(query)
	if rebound := semanticTable(t, render); rebound != result {
		t.Fatal("rebinding replaced the render result table")
	}
}
