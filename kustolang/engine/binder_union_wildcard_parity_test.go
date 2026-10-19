package kustolang

import (
	"strings"
	"testing"
)

func TestNativeUnionWildcardBindsOrdinaryTablesInDatabaseOrder(t *testing.T) {
	second := NewTableSymbol("AppTwo", NewColumnSymbol("B", ScalarTypes.Bool.Type()))
	first := NewTableSymbol("AppOne", NewColumnSymbol("A", ScalarTypes.String.Type()))
	solo := NewTableSymbol("SoloTable", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	database := NewDatabaseSymbol("db",
		NewExternalTableSymbol("AppExternal"),
		second,
		NewMaterializedViewSymbol("AppView", nil, ""),
		first,
		NewStoredQueryResultSymbol("AppStored"),
		solo,
	)
	query := ParseQuery("union App*, [Solo*]")
	NewBinder(NewGlobalState(database)).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, union), "B", "A", "N")
	app := firstNamedReference(t, union, "App*")
	group, ok := app.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 || group.Members()[0] != second || group.Members()[1] != first {
		t.Fatalf("App* symbol = %#v, want ordinary tables in database order", app.ReferencedSymbol())
	}
	if got := firstNamedReference(t, union, "Solo*").ReferencedSymbol(); got != solo {
		t.Fatalf("single wildcard symbol = %#v, want SoloTable", got)
	}
}

func TestNativeUnionWildcardIsCaseSensitiveAndUsesOrdinaryKS142(t *testing.T) {
	table := NewTableSymbol("Logs", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("union log*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	reference := firstNamedReference(t, query, "log*")
	diagnostics := reference.SemanticDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS142" || !strings.Contains(diagnostics[0].Message, "known table") {
		t.Fatalf("zero-match diagnostics = %#v, want ordinary wildcard table KS142", diagnostics)
	}
	if reference.ReferencedSymbol() != Error || !firstBoundNode[*UnionOperator](t, query).IsBound() {
		t.Fatalf("zero-match binding = symbol:%#v union-bound:%t", reference.ReferencedSymbol(), firstBoundNode[*UnionOperator](t, query).IsBound())
	}
}

func TestNativeUnionWildcardQualifiedPathParity(t *testing.T) {
	one := NewTableSymbol("LogOne", NewColumnSymbol("A", ScalarTypes.String.Type()))
	two := NewTableSymbol("LogTwo", NewColumnSymbol("B", ScalarTypes.Long.Type()))
	remote := NewDatabaseSymbol("Remote", two, NewExternalTableSymbol("LogExternal"), one)
	current := NewDatabaseSymbol("Current", NewTableSymbol("Input", NewColumnSymbol("InputOnly", ScalarTypes.Bool.Type())))
	cluster := NewClusterSymbol("cluster", current, remote)
	globals := NewGlobalState(current).WithCluster(cluster).WithDatabase(current)
	query := ParseQuery(`Input | union withsource=Source database("Remote").Log*`)
	NewBinder(globals).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, union), "Source", "InputOnly", "B", "A")
	reference := firstNamedReference(t, union, "Log*")
	group, ok := reference.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 || group.Members()[0] != two || group.Members()[1] != one {
		t.Fatalf("qualified wildcard symbol = %#v, want remote table group", reference.ReferencedSymbol())
	}
}

func TestNativeUnionWildcardRecoveredBracketAndInjectedDelimiterTransaction(t *testing.T) {
	table := NewTableSymbol("LogOne", NewColumnSymbol("A", ScalarTypes.String.Type()))
	database := NewDatabaseSymbol("Remote", table)
	cluster := NewClusterSymbol("cluster", database)
	globals := NewGlobalState(database).WithCluster(cluster).WithDatabase(database)

	recovered := ParseQuery(`union database("Remote")[Log*`)
	NewBinder(globals).TryBind(recovered)
	recoveredUnion := firstBoundNode[*UnionOperator](t, recovered)
	if !recoveredUnion.IsBound() || firstNamedReference(t, recoveredUnion, "Log*").ReferencedSymbol() != table ||
		!diagnosticsContainMessage(recoveredUnion.ContainedDiagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("parser-recovered wildcard = bound:%t diagnostics:%#v", recoveredUnion.IsBound(), recoveredUnion.ContainedDiagnostics())
	}

	injected := ParseQuery(`union database("Remote")[Log*]`)
	wildcard := firstNamedReference(t, injected, "Log*").Name().(*BracketedWildcardedName)
	wildcard.SyntaxNode.children[2] = NewMissingToken(SyntaxKindCloseBracketToken, Diagnostic{})
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Outside", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 29
	tree.bind(injected)

	injectedUnion := boundNodes[*UnionOperator](injected)[0]
	if injectedUnion.IsBound() || firstNamedReference(t, injectedUnion, "Log*").IsBound() {
		t.Fatal("undiagnosed injected delimiter committed union bindings")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 29 {
		t.Fatal("rejected wildcard delimiter changed binder or tree state")
	}
}
