package kustolang

import "testing"

func TestNativeSearchPredicateBinding(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Message", ScalarTypes.String.Type()),
		NewColumnSymbol("Level", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, text := range []string{
		`T | search "error"`,
		`T | search *`,
		`T | search ("error" or "warning") and Level > 1`,
		`T | search Message has "error"`,
	} {
		query := ParseQuery(text)
		NewBinder(globals).TryBind(query)
		search := firstBoundNode[*SearchOperator](t, query)
		if !sameType(ExpressionResultType(search.Condition()), ScalarTypes.Bool.Type()) {
			t.Fatalf("%q condition type = %v, want bool", text, ExpressionResultType(search.Condition()))
		}
		assertDiagnosticCodes(t, search)
	}
}

func TestNativeSearchPipelineAndExplicitInputSchemas(t *testing.T) {
	t1 := NewTableSymbol("T1", NewColumnSymbol("X", ScalarTypes.Long.Type()), NewColumnSymbol("A", ScalarTypes.String.Type()))
	t2 := NewTableSymbol("T2", NewColumnSymbol("X", ScalarTypes.String.Type()), NewColumnSymbol("B", ScalarTypes.Bool.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", t1, t2))

	piped := ParseQuery(`T1 | search "x"`)
	NewBinder(globals).TryBind(piped)
	pipedTable := expressionTableResult(firstBoundNode[*SearchOperator](t, piped))
	assertSearchColumnNames(t, pipedTable, "$table", "X", "A")

	root := ParseQuery(`search in (T1, T2) "x"`)
	NewBinder(globals).TryBind(root)
	search := firstBoundNode[*SearchOperator](t, root)
	assertDiagnosticCodes(t, search)
	result := expressionTableResult(search)
	assertSearchColumnNames(t, result, "$table", "X_long", "X_string", "A", "B")
	for _, reference := range boundNodes[*NameReference](root) {
		name := NameReferenceSimpleName(reference)
		if (name == "T1" && reference.ReferencedSymbol() != t1) || (name == "T2" && reference.ReferencedSymbol() != t2) {
			t.Fatalf("%s reference = %#v", name, reference.ReferencedSymbol())
		}
	}
}

func TestNativeSearchIncludesZeroArgumentLocalViewsInImpliedTables(t *testing.T) {
	code := ParseAndAnalyze(`let First = view () { datatable(A:string)["x"] };
let Second = view () { datatable(B:long)[1] };
search A == "x" or B == 1`, NewGlobalState(NewDatabaseSymbol("db")))
	search := firstBoundNode[*SearchOperator](t, code.Syntax())
	assertDiagnosticCodes(t, search)
	// Binding Second snapshots First into the shared-scope chain, so upstream
	// enumeration observes the current scope before the earlier shared scope.
	assertSearchColumnNames(t, expressionTableResult(search), "$table", "B", "A")
}

func TestNativeSearchExcludesOrdinaryAndRequiredArgumentFunctionsFromImpliedTables(t *testing.T) {
	for _, text := range []string{
		`let F = () { datatable(FOnly:string)["x"] }; search FOnly == "x"`,
		`let V = view (x:string) { datatable(VOnly:string)[x] }; search VOnly == "x"`,
	} {
		code := ParseAndAnalyze(text, NewGlobalState(NewDatabaseSymbol("db")))
		if !diagnosticsContain(code.GetDiagnostics(), "KS142") {
			t.Fatalf("%q diagnostics = %#v, want KS142", text, code.GetDiagnostics())
		}
	}
}

func TestNativeSearchIncludesOptionalArgumentViewAndSuppressesItForExplicitSources(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("DatabaseOnly", ScalarTypes.Long.Type()))))
	included := ParseAndAnalyze(`let V = view (x:string="x") { print OptionalOnly=x }; search OptionalOnly == "x"`, globals)
	if diagnostics := included.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("optional view diagnostics = %#v", diagnostics)
	}
	assertSearchColumnNames(t, expressionTableResult(firstBoundNode[*SearchOperator](t, included.Syntax())), "$table", "DatabaseOnly", "OptionalOnly")

	for _, text := range []string{
		`let V = view () { print ViewOnly="x" }; search in (T) ViewOnly == "x"`,
		`let V = view () { print ViewOnly="x" }; T | search ViewOnly == "x"`,
	} {
		code := ParseAndAnalyze(text, globals)
		if !diagnosticsContain(code.GetDiagnostics(), "KS142") {
			t.Fatalf("%q diagnostics = %#v, want KS142", text, code.GetDiagnostics())
		}
	}
}

func TestNativeSearchImpliedViewUsesEarlierStaticScope(t *testing.T) {
	code := ParseAndAnalyze(`let Source = datatable(Value:string)["x"];
let V = view () { Source };
search Value == "x"`, NewGlobalState(NewDatabaseSymbol("db")))
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("view with captured source diagnostics = %#v", diagnostics)
	}
	search := firstBoundNode[*SearchOperator](t, code.Tree.Root)
	assertSearchColumnNames(t, expressionTableResult(search), "$table", "Value")
}

func TestNativeSearchReportsMalformedPredicateAfterLocalView(t *testing.T) {
	code := ParseAndAnalyze(`let V = view () { datatable(Value:string)["x"] }; search * and`, NewGlobalState(NewDatabaseSymbol("db")))
	if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed search predicate produced no diagnostics")
	}
}

func TestNativeSearchDiagnosticsStateRestorationAndIdempotence(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery(`T | search in (T) 1`)
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)
	search := firstBoundNode[*SearchOperator](t, query)
	assertDiagnosticCodes(t, search, "KS141", "KS175")
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatalf("search leaked binder state: row=%#v kind=%v local=%p", binder.rowScope, binder.scopeKind, binder.LocalScope)
	}
	before, _ := search.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := search.GetSemanticInfo()
	if len(after.Diagnostics) != len(before.Diagnostics) || expressionTableResult(search) == nil {
		t.Fatalf("search rebinding changed semantic info: before=%#v after=%#v", before, after)
	}
}

func TestNativeSearchInClausePlacementDiagnostics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		query string
		codes []string
	}{
		{query: `search in (T) "x"`},
		{query: `T | search "x"`},
		{query: `T | search in (T) "x"`, codes: []string{"KS175"}},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			binder := NewBinder(globals)
			newTreeBinder(binder).visitElement(query)
			search := firstBoundNode[*SearchOperator](t, query)
			assertDiagnosticCodes(t, search, test.codes...)
			if len(test.codes) > 0 {
				info, _ := search.GetSemanticInfo()
				if len(info.Diagnostics) != 1 || info.Diagnostics[0].Message != "The operator must be the first operator in the query." {
					t.Fatalf("placement diagnostics = %#v", info.Diagnostics)
				}
			}
		})
	}
}

func assertSearchColumnNames(t *testing.T, table *TableSymbol, want ...string) {
	t.Helper()
	if table == nil {
		t.Fatal("search result is not tabular")
	}
	columns := table.Columns()
	if len(columns) != len(want) {
		t.Fatalf("columns = %#v, want %v", columns, want)
	}
	for i, name := range want {
		if columns[i].SymbolName() != name {
			t.Fatalf("column %d = %q, want %q", i, columns[i].SymbolName(), name)
		}
	}
}
