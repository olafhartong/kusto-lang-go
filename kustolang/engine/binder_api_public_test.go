package kustolang

import (
	"context"
	"sync"
	"testing"
)

func TestBinderAPIPublicTreeBindingAndScope(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", column)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	text := `let x=1; T | where A > x | project A`
	query := ParseQuery(text)
	tree := NewSyntaxTree(query.SyntaxNode, nil, 0)
	if !TryBindSyntaxTree(tree, globals) {
		t.Fatal("TryBindSyntaxTree failed")
	}

	position := len(`let x=1; T | where A > `)
	row := GetRowScope(tree, position, globals)
	if row.GetColumn("A") != column {
		t.Fatalf("row scope = %#v, want source column identity", row)
	}
	if got := GetReferencedSymbol(tree, position, "A", globals, SymbolMatchColumn); got != column {
		t.Fatalf("referenced A = %#v", got)
	}
	local := GetReferencedSymbol(tree, position, "x", globals, SymbolMatchLocal)
	if local == nil || local.SymbolName() != "x" {
		t.Fatalf("referenced x = %#v", local)
	}
	assertScopeContainsNames(t, GetSymbolsInScope(tree, len(text), globals, SymbolMatchColumn|SymbolMatchLocal, IncludeFunctionKindNone), "A", "x")
}

func TestBinderAPIPublicMalformedScopeAndCancellation(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", column)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	text := `let x=1; T | where (`
	query := ParseQuery(text)
	tree := NewSyntaxTree(query.SyntaxNode, nil, 0)
	if !TryBindSyntaxTree(tree, globals) {
		t.Fatal("malformed tree did not bind")
	}
	if diagnostics := GetContainedDiagnostics(tree.Root, DiagnosticsIncludeDefault); len(diagnostics) == 0 {
		t.Fatal("malformed tree produced no diagnostics")
	}
	row := GetRowScope(tree, len(text), globals)
	if row.GetColumn("A") != column {
		t.Fatalf("malformed row scope = %#v", row)
	}
	if got := GetReferencedSymbol(tree, len(text), "x", globals, SymbolMatchLocal); got == nil {
		t.Fatal("malformed recovery lost prior local")
	}
	assertScopeContainsNames(t, GetSymbolsInScope(tree, len(text), globals, SymbolMatchColumn|SymbolMatchLocal, IncludeFunctionKindNone), "A", "x")

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("canceled Binder API call did not panic")
		}
	}()
	GetSymbolsInScope(tree, len(text), globals, SymbolMatchAny, IncludeFunctionKindAll, NewCancellationToken(canceled))
}

func TestBinderAPIComputedReturnAndCalledBody(t *testing.T) {
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", value)
	function := NewComputedFunctionSymbol("Rows", `T | project Value`, TabularityTabular)
	database := NewDatabaseSymbol("db", table, function)
	globals := NewGlobalState(database)
	signature := function.Signatures()[0]
	result := GetComputedReturnType(signature, globals, nil)
	resultTable, _ := tableSymbolFromSymbol(result)
	if resultTable == nil || resultTable.GetColumn("Value") == nil || resultTable.GetColumn("Value").Type != ScalarTypes.Long.Type() {
		t.Fatalf("computed return type = %#v", result)
	}

	parameter := NewParameterKind("Input", ParameterTypeTabular)
	typed := NewComputedFunctionSymbol("Typed", `Input | project Value`, TabularityTabular, parameter)
	typedResult := GetComputedReturnType(typed.Signatures()[0], globals, []*TypeSymbol{&table.TypeSymbol})
	typedTable, _ := TableSymbolForType(typedResult)
	if typedTable == nil || typedTable.GetColumn("Value") == nil || typedTable.GetColumn("Value").Type != ScalarTypes.Long.Type() {
		t.Fatalf("type-only tabular computed return = %#v", typedResult)
	}

	scalar := NewComputedFunctionSymbol("Increment", `value + 1`, TabularityScalar, NewParameter("value", ScalarTypes.Long.Type()))
	if got := GetComputedReturnType(scalar.Signatures()[0], globals, []*TypeSymbol{ScalarTypes.Long.Type()}); got != ScalarTypes.Long.Type() {
		t.Fatalf("type-only scalar computed return = %#v, want long", got)
	}
	declared := NewScalarFunctionSymbol("Declared", ScalarTypes.Long.Type())
	if got := GetComputedReturnType(declared.Signatures()[0], globals, nil); got != ScalarTypes.Unknown.Type() {
		t.Fatalf("declared signature computed return = %#v, want pinned unknown", got)
	}

	body := ParseFunctionBody(`{ local + 1 }`)
	tree := NewSyntaxTree(body.SyntaxNode, nil, 0)
	outer := NewBinder(globals)
	local := NewVariableSymbol("local", ScalarTypes.Long.Type(), false, nil, nil)
	if !TryBindCalledFunctionBody(tree, outer, nil, nil, function, nil, []Symbol{local}) {
		t.Fatal("TryBindCalledFunctionBody failed")
	}
	if got := ExpressionResultType(body.Expression()); got != ScalarTypes.Long.Type() {
		t.Fatalf("called body result = %#v", got)
	}
}

func TestBinderAPIComputedReturnTypeIsSafeForConcurrentTypeOnlyCalls(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	parameter := NewParameterKind("Input", ParameterTypeTabular)
	function := NewComputedFunctionSymbol("Typed", `Input | project Value`, TabularityTabular, parameter)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, function)).WithCache()

	const workers = 24
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			result := GetComputedReturnType(function.Signatures()[0], globals, []*TypeSymbol{&table.TypeSymbol})
			resultTable, _ := TableSymbolForType(result)
			if resultTable == nil || resultTable.GetColumn("Value") == nil || resultTable.GetColumn("Value").Type != ScalarTypes.Long.Type() {
				t.Errorf("concurrent computed return = %#v", result)
			}
		}()
	}
	group.Wait()
}

func TestBinderAPIComputedReturnTypeStopsRecursiveExpansion(t *testing.T) {
	loop := NewComputedFunctionSymbol("Loop", `Loop()`, TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", loop)).WithCache()

	// The public API must use the same bounded expansion behavior as a query.
	if got := GetComputedReturnType(loop.Signatures()[0], globals, nil); got != ScalarTypes.Unknown.Type() {
		t.Fatalf("recursive computed return = %#v, want unknown", got)
	}
}

func TestComputedSignatureReturnTypeDelegatesToBinderAPI(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	parameter := NewParameterKind("Input", ParameterTypeTabular)
	static := NewComputedFunctionSymbol("Static", `T | project Value`, TabularityTabular)
	typed := NewComputedFunctionSymbol("Typed", `Input | project Value`, TabularityTabular, parameter)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, static, typed)).WithCache()

	staticTable, _ := TableSymbolForType(static.Signatures()[0].GetReturnType(globals))
	if staticTable == nil || staticTable.GetColumn("Value") == nil || staticTable.GetColumn("Value").Type != ScalarTypes.Long.Type() {
		t.Fatalf("computed signature static return = %#v", staticTable)
	}
	typedTable, _ := TableSymbolForType(typed.Signatures()[0].GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{&table.TypeSymbol}))
	if typedTable == nil || typedTable.GetColumn("Value") == nil || typedTable.GetColumn("Value").Type != ScalarTypes.Long.Type() {
		t.Fatalf("computed signature typed return = %#v", typedTable)
	}
}

func assertScopeContainsNames(t *testing.T, symbols []Symbol, names ...string) {
	t.Helper()
	for _, name := range names {
		found := false
		for _, symbol := range symbols {
			if symbol != nil && symbol.SymbolName() == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("scope %#v does not contain %q", symbols, name)
		}
	}
}
