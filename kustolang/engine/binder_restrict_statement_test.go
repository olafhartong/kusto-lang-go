package kustolang

import "testing"

func TestNativeRestrictStatementBindsWildcardTablesAndProperties(t *testing.T) {
	t1 := NewTableSymbol("SigninLogs", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	t2 := NewTableSymbol("SigninAudit", NewColumnSymbol("B", ScalarTypes.String.Type()))
	other := NewTableSymbol("Events", NewColumnSymbol("C", ScalarTypes.Bool.Type()))
	database := NewDatabaseSymbol("db", t1, t2, other)
	query := ParseQuery(`restrict access to (Signin*, Events) with (AllowManagedIdentityAuthentication=true); Events | count`)
	NewBinder(NewGlobalState(database)).TryBind(query)

	restrict := firstBoundNode[*RestrictStatement](t, query)
	if restrict.IsBound() {
		t.Fatal("RestrictStatement received semantic info; upstream VisitRestrictStatement returns null")
	}
	wildcard := restrict.Expressions().At(0).Element().(*NameReference)
	group, ok := wildcard.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 || group.Members()[0] != t1 || group.Members()[1] != t2 {
		t.Fatalf("restrict wildcard symbol = %#v, want ordered SigninLogs/SigninAudit group", wildcard.ReferencedSymbol())
	}
	exact := restrict.Expressions().At(1).Element().(*NameReference)
	if exact.ReferencedSymbol() != other {
		t.Fatalf("restrict exact symbol = %#v, want Events", exact.ReferencedSymbol())
	}
	property := restrict.WithClause().Properties().At(0).Element()
	if property.Expression() == nil || simpleBinderSyntaxNode(property.Expression()) == nil ||
		!simpleBinderSyntaxNode(property.Expression()).IsBound() || ExpressionResultType(property.Expression()) != ScalarTypes.Bool.Type() {
		t.Fatalf("restrict property expression = %#v", property.Expression())
	}
	if count := firstBoundNode[*CountOperator](t, query); len(count.SemanticDiagnostics()) != 0 {
		t.Fatalf("post-restrict count diagnostics = %#v", count.SemanticDiagnostics())
	}
}

func TestNativeRestrictStatementCurrentUnqualifiedWildcardIncludesLocalViews(t *testing.T) {
	table := NewTableSymbol("RowsTable", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	materialized := NewMaterializedViewSymbol("RowsMaterialized", []*ColumnSymbol{NewColumnSymbol("A", ScalarTypes.Long.Type())}, "")
	external := NewExternalTableSymbol("RowsExternal", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	stored := NewStoredQueryResultSymbol("RowsStored", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table, materialized, external, stored))
	code := ParseWithGlobals(`let RowsView=view(){print A=1}; restrict access to (Rows*); print x=1`, globals).Analyze()

	reference := firstNamedReference(t, code.Syntax(), "Rows*")
	group, ok := reference.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 || group.Members()[0] != table {
		t.Fatalf("current wildcard symbol = %#v, want ordered table/local-view group", reference.ReferencedSymbol())
	}
	view, ok := group.Members()[1].(*FunctionSymbol)
	if !ok || view.SymbolName() != "RowsView" || !view.IsView() || view.MinArgumentCount() != 0 {
		t.Fatalf("current wildcard local view = %#v", group.Members()[1])
	}
	if diagnostics := reference.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("current wildcard diagnostics = %#v", diagnostics)
	}
}

func TestNativeRestrictStatementQualifiedWildcardExcludesLocalViews(t *testing.T) {
	table := NewTableSymbol("RowsTable", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	code := ParseWithGlobals(`let RowsView=view(){print A=1}; restrict access to (database("db").Rows*); print x=1`, globals).Analyze()

	reference := firstNamedReference(t, code.Syntax(), "Rows*")
	if reference.ReferencedSymbol() != table || len(reference.SemanticDiagnostics()) != 0 {
		t.Fatalf("qualified current wildcard = %#v diagnostics=%#v, want table only", reference.ReferencedSymbol(), reference.SemanticDiagnostics())
	}
}

func TestNativeRestrictStatementExactFunctionAndQualifiedMissingRecovery(t *testing.T) {
	required := NewComputedFunctionSymbol("RowsRequired", "RowsTable", TabularityTabular, NewParameter("x", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", required))

	functionCode := ParseWithGlobals(`restrict access to (RowsRequired); print x=1`, globals).Analyze()
	functionReference := firstNamedReference(t, functionCode.Syntax(), "RowsRequired")
	if functionReference.ReferencedSymbol() != required {
		t.Fatalf("required function reference = %#v, want function", functionReference.ReferencedSymbol())
	}
	assertDiagnosticCodes(t, functionReference, "KS117")

	missingCode := ParseWithGlobals(`restrict access to (database("db").Missing); print x=1`, globals).Analyze()
	missingReference := firstNamedReference(t, missingCode.Syntax(), "Missing")
	recovery, ok := tableSymbolFromSymbol(missingReference.ReferencedSymbol())
	if !ok || !recovery.IsOpen() {
		t.Fatalf("qualified missing reference = %#v, want open-table recovery", missingReference.ReferencedSymbol())
	}
	assertDiagnosticCodes(t, missingReference, "KS204")
}

func TestNativeRestrictStatementDatabaseGroupsAndOpenClusterRecovery(t *testing.T) {
	currentDatabase := NewDatabaseSymbol("Current", NewTableSymbol("CurrentEvents"))
	currentCluster := NewClusterSymbol("current", currentDatabase)
	tenantAEvents := NewTableSymbol("Events", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	tenantBEvents := NewTableSymbol("Events", NewColumnSymbol("B", ScalarTypes.String.Type()))
	unique := NewTableSymbol("Unique", NewColumnSymbol("U", ScalarTypes.Bool.Type()))
	tenantA := NewDatabaseSymbol("TenantA", tenantAEvents, unique)
	tenantB := NewDatabaseSymbolWithAlternateName("TenantB", "TenantPretty", []Symbol{tenantBEvents}, false)
	prod := NewClusterSymbol("prod", tenantA, tenantB)
	globals := NewGlobalState(currentDatabase).
		WithClusterList([]*ClusterSymbol{currentCluster, prod}).
		WithCluster(currentCluster).
		WithDatabase(currentDatabase)

	wildcardCode := ParseWithGlobals(`restrict access to (cluster("prod").database("Tenant*").Event*); print x=1`, globals).Analyze()
	wildcard := firstNamedReference(t, wildcardCode.Syntax(), "Event*")
	wildcardGroup, ok := wildcard.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(wildcardGroup.Members()) != 2 || wildcardGroup.Members()[0] != tenantAEvents || wildcardGroup.Members()[1] != tenantBEvents {
		t.Fatalf("wildcard group = %#v", wildcard.ReferencedSymbol())
	}

	alternateCode := ParseWithGlobals(`restrict access to (cluster("prod").database("TenantPretty").Events); print x=1`, globals).Analyze()
	if reference := firstNamedReference(t, alternateCode.Syntax(), "Events"); reference.ReferencedSymbol() != tenantBEvents {
		t.Fatalf("alternate database reference = %#v, want TenantB.Events", reference.ReferencedSymbol())
	}

	uniqueCode := ParseWithGlobals(`restrict access to (cluster("prod").database("Tenant*").Unique); print x=1`, globals).Analyze()
	if reference := firstNamedReference(t, uniqueCode.Syntax(), "Unique"); reference.ReferencedSymbol() != unique || len(reference.SemanticDiagnostics()) != 0 {
		t.Fatalf("unique group pass-through = %#v diagnostics=%#v", reference.ReferencedSymbol(), reference.SemanticDiagnostics())
	}

	ambiguousCode := ParseWithGlobals(`restrict access to (cluster("prod").database("Tenant*").Events); print x=1`, globals).Analyze()
	ambiguous := firstNamedReference(t, ambiguousCode.Syntax(), "Events")
	if group, ok := ambiguous.ReferencedSymbol().(*GroupSymbol); !ok || len(group.Members()) != 2 {
		t.Fatalf("ambiguous group = %#v", ambiguous.ReferencedSymbol())
	}
	assertDiagnosticCodes(t, ambiguous, "KS146")

	unknownClusterCode := ParseWithGlobals(`restrict access to (cluster("missing").database("Any").Events); print x=1`, globals).Analyze()
	unknownDiagnostics := unknownClusterCode.GetDiagnostics()
	if len(unknownDiagnostics) != 1 || unknownDiagnostics[0].Code != "KS207" {
		t.Fatalf("unknown cluster diagnostics = %#v, want only KS207", unknownDiagnostics)
	}
	if table, ok := tableSymbolFromSymbol(firstNamedReference(t, unknownClusterCode.Syntax(), "Events").ReferencedSymbol()); !ok || !table.IsOpen() {
		t.Fatalf("unknown cluster member did not recover as open table")
	}

	malformedCode := ParseWithGlobals(`restrict access to (cluster("prod").database("TenantA").Unique,); print x=1`, globals).Analyze()
	if reference := firstNamedReference(t, malformedCode.Syntax(), "Unique"); reference.ReferencedSymbol() != unique {
		t.Fatalf("malformed prefix reference = %#v, want Unique", reference.ReferencedSymbol())
	}
	if diagnostics := malformedCode.GetDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS006" {
		t.Fatalf("malformed diagnostics = %#v, want only KS006", diagnostics)
	}
}
