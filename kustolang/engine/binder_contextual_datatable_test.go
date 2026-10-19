package kustolang

import "testing"

const contextualDataTableGuid = "00000000-0000-0000-0000-000000000001"

func TestNativeTreeBinderContextualDataTableSchemaAndDownstreamScope(t *testing.T) {
	query := ParseQuery(`__contextual_datatable guid(` + contextualDataTableGuid + `) ` +
		`(Account:string, Count:long, Account:bool, Nested:(Value:long), Recovered:) | project Count, Account`)
	binder := NewBinder(nil)
	binder.TryBind(query)

	contextual := firstBoundNode[*ContextualDataTableExpression](t, query)
	table := semanticTable(t, contextual)
	assertOrderedColumns(t, table, "Account", "Count", "Recovered")
	assertColumnTypes(t, table, map[string]*TypeSymbol{
		"Account":   ScalarTypes.String.Type(),
		"Count":     ScalarTypes.Long.Type(),
		"Recovered": ScalarTypes.Unknown.Type(),
	})
	if table.IsOpen() || table.IsSorted() || table.IsSerialized() {
		t.Fatalf("contextual datatable state = open:%t sorted:%t serialized:%t", table.IsOpen(), table.IsSorted(), table.IsSerialized())
	}
	assertDiagnosticCodes(t, contextual, "KS171", "KS170")

	declarations := contextual.Schema().Columns()
	if declarations.At(0).Element().Name().ReferencedSymbol() != table.Columns()[0] ||
		declarations.At(1).Element().Name().ReferencedSymbol() != table.Columns()[1] ||
		declarations.At(4).Element().Name().ReferencedSymbol() != table.Columns()[2] {
		t.Fatal("successful contextual declarations were not bound to their result columns")
	}
	if declarations.At(2).Element().Name().IsBound() || declarations.At(3).Element().Name().IsBound() {
		t.Fatal("omitted contextual declarations unexpectedly received column symbols")
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, project), "Count", "Account")
}

func TestNativeTreeBinderContextualDataTableIDValidation(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want []string
	}{
		{name: "guid literal", id: `guid(` + contextualDataTableGuid + `)`},
		{name: "non guid literal", id: `"context"`, want: []string{"KS141"}},
		{name: "non guid expression", id: `1 == 1`, want: []string{"KS141"}},
		{name: "parenthesized guid is not literal", id: `(guid(` + contextualDataTableGuid + `))`, want: []string{"KS131"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contextual := contextualDataTableWithID(t, test.id)
			NewBinder(nil).TryBind(contextual)
			if !contextual.IsBound() {
				t.Fatal("contextual datatable was not bound")
			}
			assertDiagnosticCodes(t, contextual, test.want...)
		})
	}

	nonLiteral := contextualDataTableWithID(t, `Id`)
	binder := NewBinder(nil)
	binder.LocalScope.AddSymbol(NewVariableSymbol("Id", ScalarTypes.Guid.Type(), false, nil, nil))
	binder.TryBind(nonLiteral)
	assertDiagnosticCodes(t, nonLiteral, "KS131")

	unknown := contextualDataTableWithID(t, `Id`)
	binder = NewBinder(nil)
	binder.LocalScope.AddSymbol(NewVariableSymbol("Id", ScalarTypes.Unknown.Type(), false, nil, nil))
	binder.TryBind(unknown)
	assertDiagnosticCodes(t, unknown, "KS131")
}

func TestNativeTreeBinderContextualDataTableIDRecoveryAndRawGuid(t *testing.T) {
	unresolved := contextualDataTableWithID(t, `Missing`)
	NewBinder(nil).TryBind(unresolved)
	unresolvedID := unresolved.Id().(*NameReference)
	if !unresolved.IsBound() || unresolvedID.ReferencedSymbol() != Error {
		t.Fatalf("unresolved id = bound:%t symbol:%#v", unresolved.IsBound(), unresolvedID.ReferencedSymbol())
	}
	assertDiagnosticCodes(t, unresolved)

	missing := parsedQueryExpression(t, `__contextual_datatable`).(*ContextualDataTableExpression)
	NewBinder(nil).TryBind(missing)
	missingID := missing.Id().(*NameReference)
	if !missing.IsBound() || missingID.ReferencedSymbol() != Error {
		t.Fatalf("missing id = bound:%t symbol:%#v", missing.IsBound(), missingID.ReferencedSymbol())
	}
	assertDiagnosticCodes(t, missing)

	raw := parsedQueryExpression(t, `__contextual_datatable `+contextualDataTableGuid+` (A:long)`).(*ContextualDataTableExpression)
	NewBinder(nil).TryBind(raw)
	assertDiagnosticCodes(t, raw)
	if !diagnosticsContain(raw.Id().ContainedDiagnostics(), "KS214") {
		t.Fatalf("raw guid syntax diagnostics = %#v, want KS214", raw.Id().ContainedDiagnostics())
	}
}

func TestNativeTreeBinderContextualDataTableIDDoesNotSeeIncomingRow(t *testing.T) {
	contextual := contextualDataTableWithID(t, `Id`)
	binder := NewBinder(nil)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Id", ScalarTypes.Guid.Type())).WithIsOpen(true)
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	newTreeBinder(binder).bind(contextual)

	id := contextual.Id().(*NameReference)
	if !contextual.IsBound() || id.ReferencedSymbol() != Error {
		t.Fatalf("contextual id resolved against incoming row: %#v", id.ReferencedSymbol())
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate {
		t.Fatal("contextual binding did not restore incoming scope")
	}
	assertDiagnosticCodes(t, contextual)
}

func contextualDataTableWithID(t *testing.T, id string) *ContextualDataTableExpression {
	t.Helper()
	node := parsedQueryExpression(t, `__contextual_datatable guid(`+contextualDataTableGuid+`) (A:long)`).(*ContextualDataTableExpression)
	expression := ParseExpression(id)
	if expression == nil {
		t.Fatalf("contextual id %q did not parse", id)
	}
	node.SyntaxNode.children[1] = syntaxElementFromTyped(expression)
	return node
}

func TestNativeTreeBinderContextualDataTableFreshAndIdempotent(t *testing.T) {
	text := `__contextual_datatable guid(` + contextualDataTableGuid + `) (A:long)`
	first := parsedQueryExpression(t, text).(*ContextualDataTableExpression)
	second := parsedQueryExpression(t, text).(*ContextualDataTableExpression)
	binder := NewBinder(nil)
	binder.TryBind(first)
	firstTable := semanticTable(t, first)
	binder.TryBind(first)
	if semanticTable(t, first) != firstTable {
		t.Fatal("repeated contextual binding replaced its committed table")
	}
	binder.TryBind(second)
	if semanticTable(t, second) == firstTable || second.Schema().Columns().At(0).Element().Name().ReferencedSymbol() == firstTable.Columns()[0] {
		t.Fatal("separate contextual expressions shared table or column symbols")
	}
}

func TestNativeTreeBinderContextualDataTableCorruptionPreflightRollback(t *testing.T) {
	for _, mutate := range []func(*ContextualDataTableExpression){
		func(node *ContextualDataTableExpression) { node.SyntaxNode.children[0] = nil },
		func(node *ContextualDataTableExpression) {
			node.Schema().Columns().SyntaxList.elementType = "SeparatedElement<SyntaxElement>"
		},
		func(node *ContextualDataTableExpression) {
			node.Schema().Columns().SyntaxList.elements[0] = node.Schema().Columns().At(0).Element()
		},
		func(node *ContextualDataTableExpression) {
			node.Schema().Columns().At(0).Element().SyntaxNode.children[1] = nil
		},
	} {
		contextual := parsedQueryExpression(t, `__contextual_datatable guid(`+contextualDataTableGuid+`) (A:long)`).(*ContextualDataTableExpression)
		mutate(contextual)
		binder := NewBinder(nil)
		outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type())).WithIsOpen(true)
		binder.rowScope = outer
		binder.scopeKind = ScopeKindAggregate
		local := binder.LocalScope
		existing := binder.getOpenTableInferredColumn("Existing", outer)
		tree := newTreeBinder(binder)
		tree.inPipe = true
		tree.inWildcard = true
		tree.nodeBinder.inPipe = true
		tree.nodeBinder.schemaCannotBeEmpty = true
		tree.nodeBinder.defaultColumnNameSuffix = 41
		tree.nodeBinder.pipedAsRowScope = outer
		tree.nodeBinder.joinRightRowScope = outer
		tree.nodeBinder.lookupRightRowScope = outer
		tree.bind(contextual)

		if contextual.IsBound() {
			t.Fatal("structurally corrupt contextual datatable was bound")
		}
		if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
			!tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe || !tree.nodeBinder.schemaCannotBeEmpty ||
			tree.nodeBinder.defaultColumnNameSuffix != 41 || tree.nodeBinder.pipedAsRowScope != outer ||
			tree.nodeBinder.joinRightRowScope != outer || tree.nodeBinder.lookupRightRowScope != outer {
			t.Fatal("contextual preflight changed binder or tree state")
		}
		if binder.openTableInferredColumns[outer]["Existing"] != existing ||
			len(binder.openTableInferredOrder[outer]) != 1 || binder.openTableInferredOrder[outer][0] != existing {
			t.Fatal("contextual preflight changed open-table inference state")
		}
		for _, syntax := range boundNodes[*SyntaxNode](contextual) {
			if syntax.IsBound() {
				t.Fatalf("contextual preflight committed child semantics: %v", syntax.Kind())
			}
		}
	}
}
