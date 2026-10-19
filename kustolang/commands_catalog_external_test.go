package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicEngineCommandCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.EngineCommands.All()
	if len(all) != 637 {
		t.Fatalf("engine command count = %d, want 637", len(all))
	}
	if all[0] != kustolang.EngineCommands.ShowDatabase ||
		all[len(all)-1] != kustolang.EngineCommands.ReplaceDatabaseKeyVaultSecrets {
		t.Fatalf("engine command order = first:%v last:%v", all[0], all[len(all)-1])
	}
	if kustolang.EngineCommands.ShowVersion.SymbolName() != "ShowVersion" ||
		kustolang.EngineCommands.ShowVersion.SymbolKind() != kustolang.SymbolKindCommand {
		t.Fatalf("ShowVersion metadata = %#v", kustolang.EngineCommands.ShowVersion)
	}
	if got := kustolang.DefaultGlobalState.GetCommand("ShowVersion"); got != kustolang.EngineCommands.ShowVersion {
		t.Fatalf("default ShowVersion = %p, catalog = %p", got, kustolang.EngineCommands.ShowVersion)
	}

	showTables := kustolang.EngineCommands.ShowTables
	if showTables.ResultSchema() != "(TableName: string, DatabaseName: string, Folder: string, DocString: string)" {
		t.Fatalf("ShowTables schema = %q", showTables.ResultSchema())
	}
	if table := showTables.ResultType(); table == nil || len(table.Columns()) != 4 || table.Columns()[0].SymbolName() != "TableName" {
		t.Fatalf("ShowTables result type = %#v", table)
	}

	all[0] = nil
	if kustolang.EngineCommands.All()[0] == nil {
		t.Fatal("EngineCommands.All exposed mutable catalog storage")
	}
}

func TestPublicDatabaseRowIDPolicyCommands(t *testing.T) {
	const schema = "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)"
	for _, command := range []*kustolang.CommandSymbol{
		kustolang.EngineCommands.ShowDatabasePolicyRowId,
		kustolang.EngineCommands.AlterDatabasePolicyRowId,
		kustolang.EngineCommands.DeleteDatabasePolicyRowId,
	} {
		if command == nil || command.ResultSchema() != schema {
			t.Fatalf("row-ID policy command = %#v", command)
		}
		if got := kustolang.DefaultGlobalState.GetCommand(command.SymbolName()); got != command {
			t.Fatalf("%s is not registered in the default command catalog", command.SymbolName())
		}
	}
	for _, query := range []string{
		".show database policy rowid",
		".show database * policy rowid",
		".alter database policy rowid true",
		".alter database policy rowid false",
		".delete database policy rowid",
	} {
		t.Run(query, func(t *testing.T) {
			code := kustolang.ParseAndAnalyze(query)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			table, ok := kustolang.TableSymbolForType(code.ResultType)
			if !ok || table.IsOpen() || len(table.Columns()) != 5 {
				t.Fatalf("result type = %#v", code.ResultType)
			}
			for i, name := range []string{"PolicyName", "EntityName", "Policy", "ChildEntities", "EntityType"} {
				column := table.Columns()[i]
				if column.SymbolName() != name || column.Type.SymbolName() != "string" {
					t.Fatalf("column %d = %#v", i, column)
				}
			}
		})
	}
}

func TestPublicServerCommandCatalogsMatchPinnedUpstream(t *testing.T) {
	catalogs := map[string]*kustolang.ServerCommandCatalog{
		"AriaBridge":     kustolang.AriaBridgeCommands,
		"ClusterManager": kustolang.ClusterManagerCommands,
		"DataManager":    kustolang.DataManagerCommands,
	}
	for name, catalog := range catalogs {
		all := catalog.All()
		if len(all) != 1 || all[0] != catalog.ShowVersion || catalog.ShowVersion.SymbolName() != "ShowVersion" {
			t.Fatalf("%s commands = %#v", name, all)
		}
		all[0] = nil
		if catalog.All()[0] == nil {
			t.Fatalf("%s All exposed mutable storage", name)
		}
	}
}

func TestPublicCommandCatalogDrivesValidAndMalformedParsing(t *testing.T) {
	valid := kustolang.ParseAndAnalyze(`.show version`)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid command diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze(`.show table`)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed command produced no diagnostic")
	}
}
