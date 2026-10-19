package kustolang

import (
	"strings"
	"testing"
)

func projectionScalarParityGlobals() (*GlobalState, *TableSymbol, *ColumnSymbol, *ColumnSymbol) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n)
	return NewGlobalState(NewDatabaseSymbol("db", table)), table, a, n
}

func TestNativeProjectionNamedGeneralScalarsAndDownstreamBinding(t *testing.T) {
	globals, _, a, n := projectionScalarParityGlobals()
	query := ParseQuery("T | project X=strlen(A), N | extend Y=N + 1, Z=-N | project X, Y, Z")
	NewBinder(globals).TryBind(query)

	projects := boundNodes[*ProjectOperator](query)
	if len(projects) != 2 || !projects[0].IsBound() || !projects[1].IsBound() {
		t.Fatalf("project binding = count:%d first:%t last:%t", len(projects), len(projects) > 0 && projects[0].IsBound(), len(projects) > 1 && projects[1].IsBound())
	}
	extend := firstBoundNode[*ExtendOperator](t, query)
	if !extend.IsBound() {
		t.Fatal("extend with arithmetic and unary expressions was not bound")
	}

	projected := semanticTable(t, projects[0])
	assertOrderedColumns(t, projected, "X", "N")
	if x := projected.GetColumn("X"); x == nil || !sameType(x.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("project X = %#v, want long", x)
	}
	extended := semanticTable(t, extend)
	assertOrderedColumns(t, extended, "X", "N", "Y", "Z")
	for _, name := range []string{"X", "Y", "Z"} {
		if column := extended.GetColumn(name); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
			t.Fatalf("extend %s = %#v, want long", name, column)
		}
	}
	assertOrderedColumns(t, semanticTable(t, projects[1]), "X", "Y", "Z")

	if firstNamedReference(t, query, "A").ReferencedSymbol() != a {
		t.Fatal("strlen argument did not preserve source-column identity")
	}
	if firstNamedReference(t, query, "N").ReferencedSymbol() != n {
		t.Fatal("projected N did not preserve source-column identity")
	}
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if !pipe.IsBound() {
			t.Fatal("downstream pipe was not bound")
		}
	}
}

func TestNativeProjectFunctionCallMalformedClosePreservesSyntaxDiagnostic(t *testing.T) {
	globals, _, a, _ := projectionScalarParityGlobals()
	query := ParseQuery("T | project X=strlen(A")
	NewBinder(globals).TryBind(query)

	project := firstBoundNode[*ProjectOperator](t, query)
	if !project.IsBound() {
		t.Fatal("project with recovered function close was not bound")
	}
	table := semanticTable(t, project)
	assertOrderedColumns(t, table, "X")
	if !sameType(table.GetColumn("X").Type, ScalarTypes.Long.Type()) {
		t.Fatalf("recovered X type = %v, want long", table.GetColumn("X").Type)
	}
	if firstNamedReference(t, query, "A").ReferencedSymbol() != a {
		t.Fatal("recovered function argument did not bind to A")
	}

	diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS005" || diagnostics[0].Message != "Expected: )" || diagnostics[0].Start != 22 || diagnostics[0].Length != 0 {
		t.Fatalf("malformed diagnostics = %#v, want KS005 Expected: ) at 22+0", diagnostics)
	}
}

func TestNativeProjectionNamedSupportedPathAndElementScalars(t *testing.T) {
	d := NewColumnSymbol("D", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", d)
	query := ParseQuery(`T | project P=D.k, E=D[0] | project P, E`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	projects := boundNodes[*ProjectOperator](query)
	if len(projects) != 2 || !projects[0].IsBound() || !projects[1].IsBound() {
		t.Fatalf("path/element project binding = count:%d first:%t last:%t", len(projects), len(projects) > 0 && projects[0].IsBound(), len(projects) > 1 && projects[1].IsBound())
	}
	result := semanticTable(t, projects[0])
	assertOrderedColumns(t, result, "P", "E")
	for _, name := range []string{"P", "E"} {
		if column := result.GetColumn(name); column == nil || !IsScalar(column.Type) || column.Type == &Error.TypeSymbol {
			t.Fatalf("%s = %#v, want supported scalar", name, column)
		}
	}
	if firstNamedReference(t, query, "D").ReferencedSymbol() != d {
		t.Fatal("path source did not bind to D")
	}
}

func TestNativeProjectionNamedPreflightAndTabularRecovery(t *testing.T) {
	globals, table, _, _ := projectionScalarParityGlobals()

	unbound := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project X=strlen(A)").(*ProjectOperator)
	binder := NewBinder(globals)
	binder.rowScope = table
	if info := newSimpleNodeBinder(binder).bind(unbound.SyntaxNode); info != nil {
		t.Fatalf("unbound scalar preflight returned semantic info %#v", info)
	}
	if unbound.IsBound() || unbound.Expressions().At(0).Element().(*SimpleNamedExpression).Name().ReferencedSymbol() != nil {
		t.Fatal("unbound scalar preflight mutated projection semantics")
	}

	errorQuery := ParseQuery("T | project X=missing_function(A) | project X")
	NewBinder(globals).TryBind(errorQuery)
	projects := boundNodes[*ProjectOperator](errorQuery)
	if len(projects) != 2 || !projects[0].IsBound() || !projects[1].IsBound() {
		t.Fatalf("error-valued project binding = %#v", projects)
	}
	if column := semanticTable(t, projects[0]).GetColumn("X"); column == nil || column.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("error-valued X = %#v, want current unknown", column)
	}

	for _, test := range []struct {
		name        string
		text        string
		wantColumns []string
	}{
		{name: "project", text: "T | project X=datatable(V:long)[1] | project X", wantColumns: []string{"X"}},
		{name: "extend", text: "T | extend X=datatable(V:long)[1] | project X", wantColumns: []string{"A", "N", "X"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			assigned := firstBoundNode[*SimpleNamedExpression](t, query)
			projects := boundNodes[*ProjectOperator](query)
			var result *TableSymbol
			if test.name == "project" {
				if len(projects) != 2 || !projects[0].IsBound() {
					t.Fatalf("tabular project binding = %#v", projects)
				}
				result = semanticTable(t, projects[0])
			} else {
				extend := firstBoundNode[*ExtendOperator](t, query)
				if !extend.IsBound() {
					t.Fatal("tabular extend was not bound")
				}
				result = semanticTable(t, extend)
			}
			assertOrderedColumns(t, result, test.wantColumns...)

			x := result.GetColumn("X")
			if x == nil || x.Type.SymbolTabularity() != TabularityTabular || assigned.Name().ReferencedSymbol() != x {
				t.Fatalf("recovered X = %#v declaration=%#v, want a committed tabular column", x, assigned.Name().ReferencedSymbol())
			}
			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			start := strings.Index(test.text, "X=datatable")
			downstreamStart := strings.LastIndex(test.text, "X")
			if len(diagnostics) != 2 || diagnostics[0].Code != "KS108" || diagnostics[0].Start != start || diagnostics[0].Length != len("X=datatable(V:long)[1]") ||
				diagnostics[1].Code != "KS108" || diagnostics[1].Start != downstreamStart || diagnostics[1].Length != 1 {
				t.Fatalf("tabular assignment diagnostics = %#v", diagnostics)
			}

			downstream := projects[len(projects)-1]
			if !downstream.IsBound() {
				t.Fatal("downstream projection of recovered X was not bound")
			}
			if column := semanticTable(t, downstream).GetColumn("X"); column != x || firstNamedReference(t, downstream, "X").ReferencedSymbol() != x {
				t.Fatalf("downstream X = %#v, want recovered column %#v", column, x)
			}
		})
	}
}
