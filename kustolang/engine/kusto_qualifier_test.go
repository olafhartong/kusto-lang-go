package kustolang

import "testing"

func TestKustoQualifierFullyMinimallyAndCrossDatabase(t *testing.T) {
	currentTable := NewTableSymbol("CurrentT", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	otherTable := NewTableSymbol("OtherT", NewColumnSymbol("Y", ScalarTypes.Long.Type()))
	currentDB := NewDatabaseSymbol("Current", currentTable)
	otherDB := NewDatabaseSymbolWithAlternateName("Other", "Pretty", []Symbol{otherTable}, false)
	cluster := NewClusterSymbol("current.kusto.windows.net", currentDB, otherDB)
	globals := NewGlobalState(currentDB).WithClusterList([]*ClusterSymbol{cluster}).WithCluster(cluster).WithDatabase(currentDB)

	if got, want := KustoQualifierFullyQualify("CurrentT", globals), "cluster('current.kusto.windows.net').database('Current').CurrentT"; got != want {
		t.Fatalf("fully qualified = %q, want %q", got, want)
	}
	qualified := "cluster('current.kusto.windows.net').database('Current').CurrentT"
	if got, want := KustoQualifierMinimallyQualify(qualified, globals), "CurrentT"; got != want {
		t.Fatalf("minimally qualified = %q, want %q", got, want)
	}
	if got, want := KustoQualifierMinimallyQualify("database('Other').OtherT", globals), "database('Other').OtherT"; got != want {
		t.Fatalf("cross-database qualification = %q, want %q", got, want)
	}
}

func TestKustoQualifierRequalifiesAcrossClustersAndLeavesCommands(t *testing.T) {
	localDB := NewDatabaseSymbol("Local")
	remoteTable := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	remoteDB := NewDatabaseSymbol("Remote", remoteTable)
	localCluster := NewClusterSymbol("local.kusto.windows.net", localDB)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", remoteDB)
	globals := NewGlobalState(localDB).
		WithClusterList([]*ClusterSymbol{localCluster, remoteCluster}).
		WithCluster(localCluster).
		WithDatabase(localDB)

	if got, want := KustoQualifierMinimallyQualify("cluster('remote.kusto.windows.net').database('Remote').T", globals), "cluster('remote.kusto.windows.net').database('Remote').T"; got != want {
		t.Fatalf("remote qualification = %q, want %q", got, want)
	}
	command := ".show tables"
	if got := KustoQualifierFullyQualify(command, globals); got != command {
		t.Fatalf("command changed to %q", got)
	}
}
