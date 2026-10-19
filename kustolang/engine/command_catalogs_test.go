package kustolang

import "testing"

func TestGeneratedSmallCommandCatalogsExposeShowVersion(t *testing.T) {
	for name, catalog := range map[string]*commandSet{
		"aria":    AriaBridgeCommands,
		"cluster": ClusterManagerCommands,
		"data":    DataManagerCommands,
	} {
		if catalog == nil || catalog.ShowVersion == nil || len(catalog.All) != 1 || catalog.All[0] != catalog.ShowVersion {
			t.Fatalf("%s catalog = %#v", name, catalog)
		}
		if catalog.ShowVersion.SymbolName() != "ShowVersion" || catalog.ShowVersion.SymbolKind() != SymbolKindCommand {
			t.Fatalf("%s ShowVersion = %#v", name, catalog.ShowVersion)
		}
		table := catalog.ShowVersion.ResultType()
		if table == nil || len(table.Columns()) != 4 {
			t.Fatalf("%s ShowVersion result table = %#v", name, table)
		}
		if table.Columns()[0].SymbolName() != "BuildVersion" || table.Columns()[0].Type != ScalarTypes.String.Type() {
			t.Fatalf("%s first result column = %#v", name, table.Columns()[0])
		}
		if table.Columns()[1].SymbolName() != "BuildTime" || table.Columns()[1].Type != ScalarTypes.DateTime.Type() {
			t.Fatalf("%s second result column = %#v", name, table.Columns()[1])
		}
	}
}

func TestGeneratedEngineCommandCatalogExposesUpstreamSymbols(t *testing.T) {
	if EngineCommands == nil || len(EngineCommands.All) != 637 {
		t.Fatalf("engine command catalog size = %#v/%d", EngineCommands, len(EngineCommands.All))
	}
	if EngineCommands.ShowExternalTablePolicyRowLevelSecurity == nil || EngineCommands.AlterExternalTablePolicyRowLevelSecurity == nil {
		t.Fatal("12.4.1 external-table row-level-security commands are missing")
	}
	if EngineCommands.ShowVersion == nil || EngineCommands.ShowVersion.SymbolName() != "ShowVersion" {
		t.Fatalf("engine ShowVersion = %#v", EngineCommands.ShowVersion)
	}
	if EngineCommands.ShowClockDiagnostics == nil || EngineCommands.ShowClockDiagnostics.SymbolName() != "ShowClockDiagnostics" ||
		EngineCommands.ShowClockDiagnostics.ResultSchema() != "()" {
		t.Fatalf("engine ShowClockDiagnostics = %#v", EngineCommands.ShowClockDiagnostics)
	}
	if EngineCommands.ShowTables == nil || EngineCommands.ShowTables.ResultSchema() != "(TableName: string, DatabaseName: string, Folder: string, DocString: string)" {
		t.Fatalf("engine ShowTables = %#v schema=%q", EngineCommands.ShowTables, EngineCommands.ShowTables.ResultSchema())
	}
	for name, command := range map[string]*CommandSymbol{
		"ShowTablesDetails": EngineCommands.ShowTablesDetails,
		"ShowTableDetails":  EngineCommands.ShowTableDetails,
	} {
		table := command.ResultType()
		if table == nil || len(table.Columns()) != 27 {
			t.Fatalf("engine %s result table = %#v", name, table)
		}
		for _, columnName := range []string{"RowLevelSecurityPolicy", "AutoDeletePolicy", "UpdatePolicy", "UpdatePolicyStates"} {
			column := table.GetColumn(columnName)
			if column == nil || column.Type != ScalarTypes.String.Type() {
				t.Fatalf("engine %s column %s = %#v, want string", name, columnName, column)
			}
		}
	}
	table := EngineCommands.ShowQueries.ResultType()
	if table == nil || len(table.Columns()) != 21 {
		t.Fatalf("engine ShowQueries result table = %#v", table)
	}
	if table.Columns()[0].SymbolName() != "ClientActivityId" || table.Columns()[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("engine ShowQueries first column = %#v", table.Columns()[0])
	}
	shardGroups := EngineCommands.TableShardGroupsStatisticsShow.ResultType()
	if shardGroups == nil {
		t.Fatal("engine TableShardGroupsStatisticsShow result table was nil")
	}
	wantShardColumns := []struct {
		name string
		typ  *TypeSymbol
	}{
		{"TableId", ScalarTypes.Guid.Type()},
	}
	if len(shardGroups.Columns()) != len(wantShardColumns) {
		t.Fatalf("engine shard-group statistics columns = %#v", shardGroups.Columns())
	}
	for i, want := range wantShardColumns {
		got := shardGroups.Columns()[i]
		if got.SymbolName() != want.name || got.Type != want.typ {
			t.Fatalf("engine shard-group statistics column %d = %#v, want %s:%s", i, got, want.name, want.typ.SymbolName())
		}
	}
	diagnostics := EngineCommands.ShowDiagnostics.ResultType()
	if diagnostics == nil {
		t.Fatal("engine ShowDiagnostics result table was nil")
	}
	for _, want := range []struct {
		name string
		typ  *TypeSymbol
	}{
		{"IsHealthy", ScalarTypes.Bool.Type()},
		{"DataPartitioningOperationsInProgress", ScalarTypes.Long.Type()},
		{"IngestionCapacityUtilization", ScalarTypes.Real.Type()},
		{"GraphSnapshotsInProgress", ScalarTypes.Long.Type()},
	} {
		got := diagnostics.GetColumn(want.name)
		if got == nil || got.Type != want.typ {
			t.Fatalf("engine ShowDiagnostics column %s = %#v, want %s", want.name, got, want.typ.SymbolName())
		}
	}
	if EngineCommands.All[0] != EngineCommands.ShowDatabase || EngineCommands.All[len(EngineCommands.All)-1] != EngineCommands.ReplaceDatabaseKeyVaultSecrets {
		t.Fatalf("engine All ordering endpoints = %#v/%#v", EngineCommands.All[0], EngineCommands.All[len(EngineCommands.All)-1])
	}
}

func TestGlobalStateGetCommandUsesServerKindCatalogs(t *testing.T) {
	data := NewGlobalState(nil).WithServerKind(ServerKindDataManager)
	if got := data.GetCommand("ShowVersion"); got != DataManagerCommands.ShowVersion {
		t.Fatalf("data-manager command = %#v", got)
	}
	if got := data.GetCommand("Missing"); got != nil {
		t.Fatalf("missing data-manager command = %#v", got)
	}

	engine := data.WithServerKind(ServerKindEngine)
	if got := engine.GetCommand("ShowVersion"); got != EngineCommands.ShowVersion {
		t.Fatalf("engine ShowVersion = %#v", got)
	}
	if got := engine.GetCommand("ShowTables"); got != EngineCommands.ShowTables {
		t.Fatalf("engine ShowTables = %#v", got)
	}
	if got := engine.GetCommand("Missing"); got != nil {
		t.Fatalf("missing engine command = %#v", got)
	}

	cluster := engine.WithServerKind(ServerKindClusterManager)
	if got := cluster.GetCommand("ShowVersion"); got != ClusterManagerCommands.ShowVersion {
		t.Fatalf("cluster command = %#v", got)
	}

	aria := cluster.WithServerKind(ServerKindAriaBridge)
	if got := aria.GetCommand("ShowVersion"); got != AriaBridgeCommands.ShowVersion {
		t.Fatalf("aria command = %#v", got)
	}
	if got := aria.GetCommand(""); got != nil {
		t.Fatalf("empty command = %#v", got)
	}

	unspecified := aria.WithServerKind("")
	if unspecified.ServerKind != "" {
		t.Fatalf("unspecified server kind = %q", unspecified.ServerKind)
	}
	if got := unspecified.GetCommand("ShowVersion"); got != nil {
		t.Fatalf("unspecified server command = %#v", got)
	}
}
