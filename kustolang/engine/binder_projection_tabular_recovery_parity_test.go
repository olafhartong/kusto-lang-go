package kustolang

// Ported from Binder_NodeBinder.VisitProjectOperator/VisitExtendOperator and
// Binder_Projection.CreateProjectionColumns in the pinned Apache-2.0 source.

import (
	"strings"
	"testing"
)

func TestNativeProjectionCompoundNamedTabularRecoveryParity(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		name        string
		text        string
		expression  string
		wantColumns []string
		wantCodes   []string
	}{
		{name: "project one", text: "T | project (X)=T", expression: "(X)=T", wantColumns: []string{"X"}, wantCodes: []string{"KS108"}},
		{name: "extend one", text: "T | extend (X)=T", expression: "(X)=T", wantColumns: []string{"A", "X"}, wantCodes: []string{"KS108"}},
		{name: "project many", text: "T | project (X,Y)=T", expression: "(X,Y)=T", wantCodes: []string{"KS108", "KS149"}},
		{name: "extend many", text: "T | extend (X,Y)=T", expression: "(X,Y)=T", wantColumns: []string{"A"}, wantCodes: []string{"KS108", "KS149"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			var result *TableSymbol
			if strings.Contains(test.name, "project") {
				operator := firstBoundNode[*ProjectOperator](t, query)
				if !operator.IsBound() {
					t.Fatal("project was not bound")
				}
				result = semanticTable(t, operator)
			} else {
				operator := firstBoundNode[*ExtendOperator](t, query)
				if !operator.IsBound() {
					t.Fatal("extend was not bound")
				}
				result = semanticTable(t, operator)
			}
			assertOrderedColumns(t, result, test.wantColumns...)

			compound := firstBoundNode[*CompoundNamedExpression](t, query)
			if len(test.wantColumns) > 0 && test.wantColumns[len(test.wantColumns)-1] == "X" {
				x := result.GetColumn("X")
				declaration := compound.Names().Names().At(0).Element()
				if x == nil || x.Type.SymbolTabularity() != TabularityTabular || declaration.ReferencedSymbol() != x {
					t.Fatalf("recovered X = %#v declaration=%#v", x, declaration.ReferencedSymbol())
				}
			}

			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			if len(diagnostics) != len(test.wantCodes) {
				t.Fatalf("diagnostics = %#v, want %v", diagnostics, test.wantCodes)
			}
			start := strings.Index(test.text, test.expression)
			for i, code := range test.wantCodes {
				wantLength := len(test.expression)
				if code == "KS149" {
					wantLength = len("(X,Y)")
				}
				if diagnostics[i].Code != code || diagnostics[i].Start != start || diagnostics[i].Length != wantLength {
					t.Fatalf("diagnostic[%d] = %#v, want %s at %d+%d", i, diagnostics[i], code, start, wantLength)
				}
			}
		})
	}
}

func TestNativeProjectionMissingNamedRHSStillRecoversAndDiagnoses(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name        string
		text        string
		wantColumns []string
		wantCodes   []string
	}{
		{name: "project simple", text: "T | project X=", wantColumns: []string{"X"}, wantCodes: []string{"KS006"}},
		{name: "extend simple", text: "T | extend X=", wantColumns: []string{"A", "X"}, wantCodes: []string{"KS006"}},
		{name: "project compound one", text: "T | project (X)=", wantColumns: []string{"X"}, wantCodes: []string{"KS006"}},
		{name: "extend compound one", text: "T | extend (X)=", wantColumns: []string{"A", "X"}, wantCodes: []string{"KS006"}},
		{name: "project compound many", text: "T | project (X,Y)=", wantCodes: []string{"KS149", "KS006"}},
		{name: "extend compound many", text: "T | extend (X,Y)=", wantColumns: []string{"A"}, wantCodes: []string{"KS149", "KS006"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)

			var result *TableSymbol
			if strings.HasPrefix(test.name, "project") {
				result = semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
			} else {
				result = semanticTable(t, firstBoundNode[*ExtendOperator](t, query))
			}
			assertOrderedColumns(t, result, test.wantColumns...)
			if strings.Contains(test.name, "simple") || strings.Contains(test.name, "one") {
				if x := result.GetColumn("X"); x == nil || x.Type != ScalarTypes.Unknown.Type() {
					t.Fatalf("recovered X = %#v, want unknown", x)
				}
			}
			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			if len(diagnostics) != len(test.wantCodes) {
				t.Fatalf("missing RHS diagnostics = %#v", diagnostics)
			}
			for i, code := range test.wantCodes {
				if diagnostics[i].Code != code {
					t.Fatalf("diagnostic[%d] = %#v, want %s", i, diagnostics[i], code)
				}
			}
			last := diagnostics[len(diagnostics)-1]
			if last.Code != "KS006" || last.Start != len(test.text) || last.Length != 0 {
				t.Fatalf("missing expression diagnostic = %#v", last)
			}
		})
	}
}

func TestNativeProjectionDirectBinderUsesBoundNamedRHSForScalarCheck(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	project := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project X=T").(*ProjectOperator)
	named := project.Expressions().At(0).Element().(*SimpleNamedExpression)
	named.Expression().(*NameReference).SetSemanticInfo(NewSemanticInfoForSymbol(table, &table.TypeSymbol))

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.rowScope = table
	nodeBinder := newSimpleNodeBinder(binder)
	nodeBinder.inPipe = true
	info := nodeBinder.bind(project.SyntaxNode)
	if info == nil || len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS108" {
		t.Fatalf("direct project semantic info = %#v, want KS108", info)
	}
	if named.Name().ReferencedSymbol() == nil {
		t.Fatal("direct project did not commit the recovered tabular column")
	}
}

func TestNativeProjectionScalarCheckStopsBeforeDuplicateDiagnostics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name        string
		text        string
		wantColumns []string
	}{
		{name: "project", text: "T | project X=T, X=T", wantColumns: []string{"X"}},
		{name: "extend", text: "T | extend X=T, X=T", wantColumns: []string{"A", "X"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)
			var result *TableSymbol
			if test.name == "project" {
				result = semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
			} else {
				result = semanticTable(t, firstBoundNode[*ExtendOperator](t, query))
			}
			assertOrderedColumns(t, result, test.wantColumns...)

			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			first := strings.Index(test.text, "X=T")
			second := strings.LastIndex(test.text, "X=T")
			if len(diagnostics) != 2 || diagnostics[0].Code != "KS108" || diagnostics[0].Start != first || diagnostics[0].Length != 3 ||
				diagnostics[1].Code != "KS171" || diagnostics[1].Start != second || diagnostics[1].Length != 1 {
				t.Fatalf("ordered diagnostics = %#v, want KS108 then KS171", diagnostics)
			}
		})
	}
}
