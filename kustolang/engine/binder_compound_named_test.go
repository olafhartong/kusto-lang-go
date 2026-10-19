package kustolang

import (
	"strings"
	"testing"
)

func TestBinderCompoundNamedPrintBindsAppliedDeconstruction(t *testing.T) {
	globals := compoundNamedTestGlobals()
	code := ParseAndAnalyze("print (First, Second)=pair()", globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("applied deconstruction diagnostics = %#v", diagnostics)
	}
	table := compoundNamedResultTable(t, code)
	assertCompoundNamedColumns(t, table, []string{"First", "Second"}, []*TypeSymbol{ScalarTypes.Long.Type(), ScalarTypes.String.Type()})

	compound := firstBoundNode[*CompoundNamedExpression](t, code.Syntax())
	for i, column := range table.Columns() {
		declaration := compound.Names().Names().At(i).Element()
		if declaration.ReferencedSymbol() != column {
			t.Fatalf("compound declaration %d symbol = %#v, want %#v", i, declaration.ReferencedSymbol(), column)
		}
	}
}

func TestBinderCompoundNamedProjectionStyles(t *testing.T) {
	globals := compoundNamedTestGlobals()
	tests := []struct {
		name    string
		query   string
		columns []string
		types   []*TypeSymbol
	}{
		{
			name:    "print omits unnamed tuple remainder",
			query:   "print (Renamed)=pair()",
			columns: []string{"Renamed"},
			types:   []*TypeSymbol{ScalarTypes.Long.Type()},
		},
		{
			name:    "project retains unnamed tuple remainder",
			query:   "T | project (Renamed)=pair()",
			columns: []string{"Renamed", "Second"},
			types:   []*TypeSymbol{ScalarTypes.Long.Type(), ScalarTypes.String.Type()},
		},
		{
			name:    "extend replaces matching input and preserves row",
			query:   "T | extend (First, Renamed)=pair()",
			columns: []string{"Existing", "First", "Second", "Renamed"},
			types:   []*TypeSymbol{ScalarTypes.Bool.Type(), ScalarTypes.Long.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.String.Type()},
		},
		{
			name:    "extend uniquifies unclaimed direct function tuple remainder",
			query:   "T | extend (Renamed)=pair()",
			columns: []string{"Existing", "First", "Second", "Renamed", "Second1"},
			types:   []*TypeSymbol{ScalarTypes.Bool.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Long.Type(), ScalarTypes.String.Type()},
		},
		{
			name:    "serialize preserves row and appends declarations",
			query:   "T | serialize (X, Y)=pair()",
			columns: []string{"Existing", "First", "Second", "X", "Y"},
			types:   []*TypeSymbol{ScalarTypes.Bool.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Long.Type(), ScalarTypes.String.Type()},
		},
		{
			name:    "summarize deconstructs tuple aggregate",
			query:   "T | summarize (X, Y)=pair()",
			columns: []string{"X", "Y"},
			types:   []*TypeSymbol{ScalarTypes.Long.Type(), ScalarTypes.String.Type()},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query, globals)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("compound projection diagnostics = %#v", diagnostics)
			}
			assertCompoundNamedColumns(t, compoundNamedResultTable(t, code), test.columns, test.types)
		})
	}
}

func TestBinderCompoundNamedDiagnostics(t *testing.T) {
	globals := compoundNamedTestGlobals()
	tests := []struct {
		name        string
		query       string
		code        string
		columns     []string
		requireDiag bool
	}{
		{name: "extra tuple name", query: "print (A, B, C)=pair()", code: "KS150", columns: []string{"A", "B"}},
		{name: "non tuple multiple names", query: "print (A, B)=1", code: "KS149", columns: []string{}},
		{name: "non tuple one name", query: "print (A)=1", columns: []string{"A"}},
		{name: "duplicate tuple name", query: "print (A, A)=pair()", code: "KS171", columns: []string{"A"}},
		{name: "missing rhs", query: "print (A, B)=", columns: []string{}, requireDiag: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query, globals)
			diagnostics := code.GetDiagnostics()
			if test.code != "" && !diagnosticsContain(diagnostics, test.code) {
				t.Fatalf("compound diagnostics = %#v, want %s", diagnostics, test.code)
			}
			if test.code == "KS150" {
				start := strings.Index(test.query, "C")
				for _, diagnostic := range diagnostics {
					if diagnostic.Code == "KS150" && (diagnostic.LocationKind != DiagnosticLocationAbsolute || diagnostic.Start != start || diagnostic.Length != 1) {
						t.Fatalf("KS150 location = %s %d/%d, want Absolute %d/1", diagnostic.LocationKind, diagnostic.Start, diagnostic.Length, start)
					}
				}
			}
			if test.requireDiag && len(diagnostics) == 0 {
				t.Fatalf("malformed compound expression %q produced no diagnostics", test.query)
			}
			table := compoundNamedResultTable(t, code)
			got := make([]string, 0, len(table.Columns()))
			for _, column := range table.Columns() {
				got = append(got, column.SymbolName())
			}
			if len(got) != len(test.columns) {
				t.Fatalf("compound columns = %#v, want %#v", got, test.columns)
			}
			for i := range got {
				if got[i] != test.columns[i] {
					t.Fatalf("compound columns = %#v, want %#v", got, test.columns)
				}
			}
		})
	}
}

func compoundNamedTestGlobals() *GlobalState {
	pair := NewFunctionSymbol("pair", NewSymbolSignature(NewTupleSymbolOf(
		NewColumnSymbol("First", ScalarTypes.Long.Type()),
		NewColumnSymbol("Second", ScalarTypes.String.Type()),
	)))
	table := NewTableSymbol("T",
		NewColumnSymbol("Existing", ScalarTypes.Bool.Type()),
		NewColumnSymbol("First", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("Second", ScalarTypes.Dynamic.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", table)).
		WithFunctions([]*FunctionSymbol{pair}).
		WithAggregates([]*FunctionSymbol{pair})
}

func compoundNamedResultTable(t *testing.T, code *KustoCode) *TableSymbol {
	t.Helper()
	if code == nil || code.ResultType == nil {
		t.Fatal("compound query has no result type")
	}
	table, ok := TableSymbolForType(code.ResultType)
	if !ok || table == nil {
		t.Fatalf("compound result type = %#v, want table", code.ResultType)
	}
	return table
}

func assertCompoundNamedColumns(t *testing.T, table *TableSymbol, names []string, types []*TypeSymbol) {
	t.Helper()
	columns := table.Columns()
	if len(columns) != len(names) || len(columns) != len(types) {
		t.Fatalf("compound columns = %#v, want names:%#v types:%#v", columns, names, types)
	}
	for i, column := range columns {
		if column.SymbolName() != names[i] || !sameType(column.Type, types[i]) {
			t.Fatalf("compound column %d = %s:%s, want %s:%s", i, column.SymbolName(), typeName(column.Type), names[i], typeName(types[i]))
		}
	}
}
