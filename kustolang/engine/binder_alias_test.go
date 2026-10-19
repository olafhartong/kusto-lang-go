package kustolang

import "testing"

func aliasBinderProfiles() []struct {
	name  string
	parse func(string) *QueryBlock
} {
	return []struct {
		name  string
		parse func(string) *QueryBlock
	}{{name: "current", parse: ParseQuery}}
}

func TestNativeTreeBinderDatabaseAliasResolvesExactDatabaseAndTable(t *testing.T) {
	for _, profile := range aliasBinderProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			column := NewColumnSymbol("A", ScalarTypes.String.Type())
			table := NewTableSymbol("T", column)
			remoteDatabase := NewDatabaseSymbol("db", table)
			remoteCluster := NewClusterSymbol("c", remoteDatabase)
			currentDatabase := NewDatabaseSymbol("current")
			currentCluster := NewClusterSymbol("current-cluster", currentDatabase)
			globals := NewGlobalState(currentDatabase).WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).WithCluster(currentCluster)
			query := profile.parse(`alias database Sales = cluster("c").database("db"); database("Sales").T | project A`)

			binder := NewBinder(globals)
			binder.TryBind(query)
			if binder.aliasedDatabases["Sales"] != remoteDatabase {
				t.Fatalf("registered aliases = %#v, want Sales remote database", binder.aliasedDatabases)
			}

			if diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
				t.Fatalf("alias diagnostics = %#v", diagnostics)
			}
			databaseCall := lastNamedFunctionCall(t, query, "database")
			result, _ := databaseCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
			if result != remoteDatabase {
				t.Fatalf("database alias result = %#v, want remote database", result)
			}
			projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
			if projected.GetColumn("A") != column {
				t.Fatal("database alias did not preserve table column identity")
			}
		})
	}
}

func TestNativeTreeBinderDatabaseAliasRegistersRecoveredMissingEqual(t *testing.T) {
	for _, profile := range aliasBinderProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
			remoteDatabase := NewDatabaseSymbol("db", table)
			remoteCluster := NewClusterSymbol("c", remoteDatabase)
			currentDatabase := NewDatabaseSymbol("current")
			currentCluster := NewClusterSymbol("current-cluster", currentDatabase)
			globals := NewGlobalState(currentDatabase).WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).WithCluster(currentCluster)
			query := profile.parse(`alias database Sales cluster("c").database("db"); database("Sales").T | count`)

			NewBinder(globals).TryBind(query)

			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeDefault)
			if !diagnosticsContain(diagnostics, "KS005") || diagnosticsContain(diagnostics, "KS208") {
				t.Fatalf("recovered alias diagnostics = %#v, want KS005 without KS208", diagnostics)
			}
			result, _ := lastNamedFunctionCall(t, query, "database").SyntaxNode.Payload(functionCallResultSymbolPayload)
			if result != remoteDatabase {
				t.Fatalf("recovered database alias result = %#v, want remote database", result)
			}
		})
	}
}

func TestNativeTreeBinderDatabaseAliasRejectsNonDatabaseAndMatchesExactCase(t *testing.T) {
	remoteDatabase := NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())))
	remoteCluster := NewClusterSymbol("c", remoteDatabase)
	currentDatabase := NewDatabaseSymbol("current")
	currentCluster := NewClusterSymbol("current-cluster", currentDatabase)
	globals := NewGlobalState(currentDatabase).WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).WithCluster(currentCluster)

	nonDatabase := ParseQuery(`alias database Sales = 1; database("Sales").T`)
	NewBinder(globals).TryBind(nonDatabase)
	if diagnosticsContain(GetContainedDiagnostics(nonDatabase, DiagnosticsIncludeDefault), "KS114") {
		t.Fatal("non-database alias surfaced the upstream-discarded KS114")
	}
	assertSemanticDiagnostic(t, lastNamedFunctionCall(t, nonDatabase, "database"), "KS208")

	wrongCase := ParseQuery(`alias database Sales = cluster("c").database("db"); database("sales").T`)
	NewBinder(globals).TryBind(wrongCase)
	assertSemanticDiagnostic(t, lastNamedFunctionCall(t, wrongCase, "database"), "KS208")
}

func lastNamedFunctionCall(t *testing.T, root SyntaxElement, name string) *FunctionCallExpression {
	t.Helper()
	var found *FunctionCallExpression
	for _, call := range boundNodes[*FunctionCallExpression](root) {
		if NameReferenceSimpleName(call.Name()) == name {
			found = call
		}
	}
	if found == nil {
		t.Fatalf("function call %q not found", name)
	}
	return found
}
