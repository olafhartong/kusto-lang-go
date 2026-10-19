package kustolang

import "testing"

func TestBinderMiscPublicDeclaredAndInferredColumns(t *testing.T) {
	declared := NewColumnSymbol("Declared", ScalarTypes.String.Type())
	open := NewTableSymbol("Open", declared).WithIsOpen(true)
	closed := NewTableSymbol("Closed", declared)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", open, closed)))

	if got := binder.GetDeclaredAndInferredColumns(nil); len(got) != 0 {
		t.Fatalf("nil table columns = %#v", got)
	}
	if column, ok := binder.TryGetDeclaredOrInferredColumn(open, "Declared"); !ok || column != declared {
		t.Fatalf("declared column = %#v/%t", column, ok)
	}
	first, ok := binder.TryGetDeclaredOrInferredColumn(open, "Future")
	if !ok || first == nil || first.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("inferred column = %#v/%t", first, ok)
	}
	second, ok := binder.TryGetDeclaredOrInferredColumn(open, "Future")
	if !ok || second != first {
		t.Fatal("inferred column identity was not stable")
	}
	caseVariant, ok := binder.TryGetDeclaredOrInferredColumn(open, "future")
	if !ok || caseVariant == first {
		t.Fatal("inferred column lookup must retain source case sensitivity")
	}
	if column, ok := binder.TryGetDeclaredOrInferredColumn(closed, "Missing"); ok || column != nil {
		t.Fatalf("closed missing column = %#v/%t", column, ok)
	}
	columns := binder.GetDeclaredAndInferredColumns(open)
	if len(columns) != 3 || columns[0] != declared || columns[1] != first || columns[2] != caseVariant {
		t.Fatalf("declared/inferred order = %#v", columns)
	}
	if tableColumns := open.Columns(); len(tableColumns) != 1 || tableColumns[0] != declared {
		t.Fatalf("inference mutated declared table columns = %#v", tableColumns)
	}
}

func TestBinderMiscPublicOpenTableValidAndMalformedInference(t *testing.T) {
	for _, test := range []struct {
		name       string
		query      string
		columnName string
		malformed  bool
	}{
		{name: "valid", query: `Open | where Future > 0 | project Future`, columnName: "Future"},
		{name: "malformed", query: `Open | where Broken >`, columnName: "Broken", malformed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			open := NewTableSymbol("Open").WithIsOpen(true)
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", open)))
			query := ParseQuery(test.query)
			binder.TryBind(query)
			if diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault); (len(diagnostics) > 0) != test.malformed {
				t.Fatalf("diagnostics = %#v, malformed=%t", diagnostics, test.malformed)
			}
			column, ok := binder.TryGetDeclaredOrInferredColumn(open, test.columnName)
			if !ok || column == nil {
				t.Fatalf("bound inferred column = %#v/%t", column, ok)
			}
			columns := binder.GetDeclaredAndInferredColumns(open)
			if len(columns) != 1 || columns[0] != column {
				t.Fatalf("bound inferred columns = %#v", columns)
			}
		})
	}
}

func TestBinderMiscPublicMacroExpandScope(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	empty := NewTableSymbol("Empty")
	group := NewEntityGroupSymbol("G", empty, table)
	scope := GetMacroExpandScopeForGroup("X", group)
	if scope == nil || scope.SymbolName() != "X" || scope.EntityGroup != group || scope.UnderlyingSymbol != empty {
		t.Fatalf("group scope = %#v", scope)
	}

	path, _ := ParseExpression(`X.A`).(*PathExpression)
	left, _ := path.Expression().(*NameReference)
	left.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(NewVariableSymbol("X", scope, false, nil, nil)))
	NewSyntaxTree(path.SyntaxNode, nil, 0)
	if got := GetMacroExpandScope(path.SyntaxNode); got != scope {
		t.Fatalf("path scope = %#v", got)
	}
	selector, _ := path.Selector().(*NameReference)
	if got := GetMacroExpandScope(selector.SyntaxNode); got != scope {
		t.Fatalf("selector scope = %#v", got)
	}

	callPath, _ := ParseExpression(`X.database("db")`).(*PathExpression)
	callLeft, _ := callPath.Expression().(*NameReference)
	callLeft.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(scope))
	NewSyntaxTree(callPath.SyntaxNode, nil, 0)
	call, _ := callPath.Selector().(*FunctionCallExpression)
	if got := GetMacroExpandScope(call.Name().SyntaxNode); got != scope {
		t.Fatalf("function path scope = %#v", got)
	}
	if got := GetMacroExpandScope(simpleBinderSyntaxNode(ParseExpression(`A`))); got != nil {
		t.Fatalf("ordinary name macro scope = %#v", got)
	}
}
