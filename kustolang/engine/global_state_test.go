package kustolang

import (
	"sync"
	"testing"
)

func TestGlobalStateDoesNotClassifyDatabaseFunctionsAsBuiltIn(t *testing.T) {
	databaseFunction := NewScalarFunctionSymbol("database_only", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", databaseFunction))
	if globals.GetFunction(databaseFunction.SymbolName()) != nil {
		t.Fatal("GetFunction should only search the ordinary built-in catalog")
	}
	if globals.GetAnyFunction(databaseFunction.SymbolName()) != databaseFunction {
		t.Fatal("database function should remain resolvable through the compatibility lookup")
	}
	if globals.IsBuiltInFunction(databaseFunction) {
		t.Fatal("database function was classified as built-in")
	}
	if !globals.IsBuiltInFunction(Functions.ToString) || !globals.IsBuiltInFunction(Aggregates.Count) || !globals.IsBuiltInFunction(PlugIns.BagUnpack) {
		t.Fatal("known function, aggregate, or plug-in was not classified as built-in")
	}
}

func TestDefaultGlobalStateUsesUnknownEnvironment(t *testing.T) {
	if DefaultGlobalState.Database != UnknownDatabaseSymbol || DefaultGlobalState.Cluster != UnknownClusterSymbol {
		t.Fatalf("default environment = cluster:%#v database:%#v", DefaultGlobalState.Cluster, DefaultGlobalState.Database)
	}
	if len(DefaultGlobalState.Clusters) != 0 {
		t.Fatalf("default clusters = %#v, want empty", DefaultGlobalState.Clusters)
	}
	if !DefaultGlobalState.Database.IsOpen() || !DefaultGlobalState.Cluster.IsOpen() {
		t.Fatal("default unknown database and cluster should be open")
	}
	compatibility := NewGlobalState(nil)
	if compatibility.Database == UnknownDatabaseSymbol || compatibility.Database.SymbolName() != "db" ||
		compatibility.Cluster == UnknownClusterSymbol || compatibility.Cluster.SymbolName() != "db:cluster" {
		t.Fatalf("nil constructor compatibility environment = cluster:%#v database:%#v", compatibility.Cluster, compatibility.Database)
	}
}

func TestGlobalStateFunctionLookupsMirrorCatalogBoundaries(t *testing.T) {
	first := NewScalarFunctionSymbol("duplicate", ScalarTypes.Long.Type())
	last := NewScalarFunctionSymbol("duplicate", ScalarTypes.String.Type())
	aggregate := NewScalarFunctionSymbol("aggregate_only", ScalarTypes.Long.Type())
	plugin := NewScalarFunctionSymbol("plugin_only", ScalarTypes.Dynamic.Type())
	database := NewScalarFunctionSymbol("database_only", ScalarTypes.Bool.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", database)).
		WithFunctions([]*FunctionSymbol{first, last}).
		WithAggregates([]*FunctionSymbol{aggregate}).
		WithPlugIns([]*FunctionSymbol{plugin})

	if globals.GetFunction("duplicate") != last {
		t.Fatal("GetFunction should select the last exact ordinary function")
	}
	if globals.GetFunction("Duplicate") != nil {
		t.Fatal("GetFunction should use source-compatible case-sensitive lookup")
	}
	if globals.IsBuiltInFunction(first) || !globals.IsBuiltInFunction(last) {
		t.Fatal("built-in identity should follow the selected duplicate")
	}
	if globals.GetFunction("aggregate_only") != nil || globals.GetFunction("plugin_only") != nil || globals.GetFunction("database_only") != nil {
		t.Fatal("GetFunction crossed a catalog boundary")
	}
	if globals.GetAggregate("aggregate_only") != aggregate || globals.GetPlugIn("plugin_only") != plugin {
		t.Fatal("scope-specific catalog lookup failed")
	}
	if globals.GetAnyFunction("database_only") != database || globals.GetAnyFunction("aggregate_only") != aggregate || globals.GetAnyFunction("plugin_only") != plugin {
		t.Fatal("compatibility lookup did not search every function source")
	}
}

func TestDefaultGlobalStateValidAndMalformedFunctionCalls(t *testing.T) {
	valid := ParseAndAnalyze(`print Value=strcat("a", "b")`, DefaultGlobalState)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid built-in call diagnostics = %#v", diagnostics)
	}

	malformed := ParseAndAnalyze(`print Value=strcat("a",`, DefaultGlobalState)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed built-in call should report syntax diagnostics")
	}
}

func TestGlobalStateCopyOperatorsAndParameterCompatibility(t *testing.T) {
	ambientTable := NewTableSymbol("Ambient")
	originalParameter := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	customAdd := NewOperatorSymbol(OperatorKindAdd,
		NewSignature(ScalarTypes.String.Type(),
			NewParameter("left", ScalarTypes.Long.Type()),
			NewParameter("right", ScalarTypes.Long.Type())))
	globals := NewGlobalState(NewDatabaseSymbol("db", NewScalarFunctionSymbol("database_only", ScalarTypes.Long.Type()))).
		WithAmbientSymbols([]Symbol{ambientTable, originalParameter}).
		WithOperators([]*OperatorSymbol{Operators.Add, customAdd}).
		WithCache()

	copy := globals.Copy()
	if copy == nil || copy == globals {
		t.Fatal("Copy should return a distinct GlobalState")
	}
	if copy.Cache == nil || copy.Cache == globals.Cache || copy.Cache.Globals != copy {
		t.Fatalf("Copy should rebind its cache, original:%#v copy:%#v", globals.Cache, copy.Cache)
	}
	if copy.GetOperator(OperatorKindAdd) != customAdd || globals.GetOperator(OperatorKindAdd) != customAdd {
		t.Fatal("operator lookup should retain the last catalog entry")
	}
	if got := copy.Parameters(); len(got) != 1 || got[0] != originalParameter {
		t.Fatalf("Parameters = %#v, want original ambient parameter", got)
	}

	replacementParameter := NewParameterSymbol("workspace", ScalarTypes.Long.Type(), "")
	withParameters := copy.WithParameters([]*ParameterSymbol{replacementParameter})
	if got := withParameters.Parameters(); len(got) != 1 || got[0] != replacementParameter {
		t.Fatalf("WithParameters = %#v, want replacement parameter", got)
	}
	if withParameters.GetAmbientSymbol("Ambient") != ambientTable || copy.GetAmbientSymbol("tenant") != originalParameter {
		t.Fatal("parameter compatibility helpers should not remove non-parameter ambient symbols or mutate the source")
	}
	additionalParameter := NewParameterSymbol("region", ScalarTypes.String.Type(), "")
	withAdditional := withParameters.AddParameters(additionalParameter)
	if got := withAdditional.GetAmbientSymbol("region"); got != additionalParameter {
		t.Fatalf("AddParameters region = %#v", got)
	}
	if !copy.IsBuiltInFunctionName("tolower") || copy.IsBuiltInFunctionName("database_only") || copy.IsBuiltInFunctionName("") {
		t.Fatal("IsBuiltInFunctionName should only recognize exact built-in catalog names")
	}
}

func TestGlobalStateOperatorsDriveBindingAndCompletion(t *testing.T) {
	customAdd := NewOperatorSymbol(OperatorKindAdd,
		NewSignature(ScalarTypes.String.Type(),
			NewParameter("left", ScalarTypes.Long.Type()),
			NewParameter("right", ScalarTypes.Long.Type())))
	globals := NewGlobalState(nil).WithOperators([]*OperatorSymbol{customAdd})
	code := ParseAndAnalyze("print Result=1 + 2", globals)
	binary := firstBoundNode[*BinaryExpression](t, code.Syntax())
	if result := ExpressionResultType(binary); result != ScalarTypes.String.Type() {
		t.Fatalf("custom Add result = %#v, want string", result)
	}
	if completionHasMatchingInfixOperator(globals, OperatorKindEqual, ScalarTypes.Long.Type(), CompletionHintBoolean) {
		t.Fatal("completion should not use operators excluded from GlobalState")
	}
}

func TestGlobalStateClusterAndDatabaseTransitionsMirrorUpstream(t *testing.T) {
	db := NewDatabaseSymbol("db")
	globals := NewGlobalState(db)
	if globals.Domain != KustoWindowsNet || globals.Cluster == nil || globals.Cluster.SymbolName() != "db:cluster" ||
		len(globals.Clusters) != 1 || globals.Clusters[0] != globals.Cluster || globals.Database != db {
		t.Fatalf("synthetic current cluster = domain:%q clusters:%#v current:%#v database:%#v", globals.Domain, globals.Clusters, globals.Cluster, globals.Database)
	}
	if globals.GetCluster("DB:CLUSTER") != nil {
		t.Fatal("cluster names containing a colon should be parsed as host and port")
	}

	samples := NewDatabaseSymbol("Samples")
	help := NewClusterSymbol("help.kusto.windows.net", samples)
	configured := globals.WithClusterList([]*ClusterSymbol{globals.Cluster, help})
	for _, name := range []string{"help.kusto.windows.net", "HELP.KUSTO.WINDOWS.NET", "help", "https://help.kusto.windows.net/Samples"} {
		if got := configured.GetCluster(name); got != help {
			t.Fatalf("GetCluster(%q) = %#v, want help", name, got)
		}
	}
	if configured.GetCluster("missing") != nil {
		t.Fatal("unknown cluster should not resolve")
	}

	selected := configured.WithCluster(help)
	if selected.Cluster != help || selected.Database != UnknownDatabaseSymbol {
		t.Fatalf("WithCluster should preserve a same-named database only, got cluster:%#v database:%#v", selected.Cluster, selected.Database)
	}
	selected = selected.WithDatabase(samples)
	if selected.Cluster != help || selected.Database != samples || selected.GetClusterForDatabase(samples) != help {
		t.Fatalf("WithDatabase known remote transition = cluster:%#v database:%#v", selected.Cluster, selected.Database)
	}

	other := NewDatabaseSymbol("other")
	withOther := configured.WithDatabase(other)
	if withOther.Database != other || withOther.Cluster == nil || withOther.Cluster.SymbolName() != "other:cluster" ||
		withOther.GetClusterForDatabase(other) != withOther.Cluster {
		t.Fatalf("WithDatabase synthetic transition = clusters:%#v current:%#v database:%#v", withOther.Clusters, withOther.Cluster, withOther.Database)
	}

	unknown := configured.WithCluster(nil)
	if unknown.Cluster != UnknownClusterSymbol || unknown.Database != UnknownDatabaseSymbol {
		t.Fatalf("unknown cluster transition = cluster:%#v database:%#v", unknown.Cluster, unknown.Database)
	}
}

func TestGlobalStateGetClusterURIParity(t *testing.T) {
	help := NewClusterSymbol("help.kusto.windows.net", NewDatabaseSymbol("Samples"))
	globals := NewGlobalState(nil).WithClusterList([]*ClusterSymbol{help})

	for _, name := range []string{
		"help.kusto.windows.net/Samples",
		"help.kusto.windows.net:443/Samples?x=1",
		"https://help.kusto.windows.net/Samples",
		"https://help.kusto.windows.net:443/Samples?x=1",
	} {
		t.Run(name, func(t *testing.T) {
			if got := globals.GetCluster(name); got != help {
				t.Fatalf("GetCluster(%q) = %#v, want help", name, got)
			}
		})
	}
}

func TestGlobalStateClusterListReplacementAndDomainAreImmutable(t *testing.T) {
	db := NewDatabaseSymbol("db")
	original := NewGlobalState(db).WithCache()
	replacementDB := NewDatabaseSymbol("db")
	replacement := NewClusterSymbol("db:cluster", replacementDB)
	updated := original.WithClusterList([]*ClusterSymbol{replacement})
	if updated == original || updated.Cluster != replacement || updated.Database != replacementDB {
		t.Fatalf("cluster replacement = %#v", updated)
	}
	if original.Cluster == replacement || original.Database != db {
		t.Fatal("cluster replacement mutated original state")
	}
	if updated.Cache == nil || updated.Cache.Globals != updated {
		t.Fatal("cluster replacement should rebind the cache")
	}

	domain := updated.WithDomain("example.test")
	if domain.Domain != ".example.test" || updated.Domain != KustoWindowsNet {
		t.Fatalf("domain transition = old:%q new:%q", updated.Domain, domain.Domain)
	}
	cluster := NewClusterSymbol("tenant.example.test", NewDatabaseSymbol("db"))
	domain = domain.WithClusterList([]*ClusterSymbol{cluster}).WithCluster(cluster)
	if domain.GetCluster("tenant") != cluster {
		t.Fatal("custom-domain short cluster lookup failed")
	}
}

func TestGlobalStateReverseOwnershipIncludesRemoteDatabases(t *testing.T) {
	currentTable := NewTableSymbol("Current", NewColumnSymbol("Local", ScalarTypes.String.Type()))
	current := NewDatabaseSymbol("db", currentTable)
	remoteColumn := NewColumnSymbol("RemoteValue", ScalarTypes.Long.Type())
	remoteTable := NewTableSymbol("Remote", remoteColumn)
	remoteFunction := NewFunctionSymbol("RemoteRows", NewSymbolSignature(remoteTable))
	remote := NewDatabaseSymbol("Samples", remoteTable, remoteFunction)
	currentCluster := NewClusterSymbol("current.kusto.windows.net", current)
	remoteCluster := NewClusterSymbol("help.kusto.windows.net", remote)
	globals := NewGlobalState(current).WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster})

	if globals.GetDatabase(remoteTable) != remote || globals.GetDatabase(remoteFunction) != remote {
		t.Fatal("remote database ownership was not resolved")
	}
	if !globals.IsDatabaseTable(remoteTable) || !globals.IsDatabaseFunction(remoteFunction) || !globals.IsDatabaseSymbol(remoteTable) {
		t.Fatal("remote database membership was not recognized")
	}
	if globals.GetTable(remoteColumn) != remoteTable {
		t.Fatal("remote column ownership was not resolved")
	}
	if globals.GetDatabase(NewTableSymbol("Unknown")) != nil || globals.GetTable(NewColumnSymbol("Unknown", ScalarTypes.String.Type())) != nil {
		t.Fatal("unconfigured symbols should not acquire database ownership")
	}
}

func TestGlobalStateRecognizesDerivedTableOwnership(t *testing.T) {
	external := NewExternalTableSymbol("Archive", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	view := NewMaterializedViewSymbol("Daily", []*ColumnSymbol{NewColumnSymbol("Day", ScalarTypes.DateTime.Type())}, "T | summarize by bin(Timestamp, 1d)")
	database := NewDatabaseSymbol("db", external, view)
	globals := NewGlobalState(database)

	for _, symbol := range []Symbol{external, &external.TableSymbol, view, &view.TableSymbol} {
		if globals.GetDatabase(symbol) != database || !globals.IsDatabaseTable(symbol) || !globals.IsDatabaseSymbol(symbol) {
			t.Fatalf("derived table ownership was not recognized for %T", symbol)
		}
	}
	if database.GetAnyTableSymbol("Archive") != external || database.GetAnyTableSymbol("Daily") != view {
		t.Fatal("identity-preserving lookup returned an embedded table")
	}
}

func TestGlobalStateClientSymbolsMirrorUpstreamLookupAndUpdates(t *testing.T) {
	globals := NewGlobalState(nil).WithCache()
	tenant := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	workspace := NewParameterSymbol("workspace", ScalarTypes.Long.Type(), "")

	withClients := globals.WithClientSymbols([]Symbol{tenant, workspace})
	if withClients == globals {
		t.Fatal("WithClientSymbols should clone the global state")
	}
	if withClients.Cache == nil || withClients.Cache == globals.Cache || withClients.Cache.Globals != withClients {
		t.Fatalf("WithClientSymbols should rebind cache, got old:%#v new:%#v", globals.Cache, withClients.Cache)
	}
	if got := globals.GetClientSymbol("tenant"); got != nil {
		t.Fatalf("original GlobalState client lookup mutated = %#v", got)
	}
	if got := withClients.GetClientSymbol("tenant"); got != tenant {
		t.Fatalf("GetClientSymbol tenant = %#v", got)
	}
	if got := withClients.GetClientSymbol("Tenant"); got != nil {
		t.Fatalf("GetClientSymbol should be case-sensitive, got %#v", got)
	}
	if got := withClients.GetClientSymbol("missing"); got != nil {
		t.Fatalf("GetClientSymbol missing = %#v", got)
	}

	replacement := NewParameterSymbol("tenant", ScalarTypes.Bool.Type(), "")
	updated := withClients.AddOrUpdateClientSymbols(replacement)
	if got := updated.GetClientSymbol("tenant"); got != replacement {
		t.Fatalf("updated tenant = %#v", got)
	}
	if got := withClients.GetClientSymbol("tenant"); got != tenant {
		t.Fatalf("previous GlobalState should keep old client symbol, got %#v", got)
	}
	if len(updated.ClientSymbols) != 2 || updated.ClientSymbols[0] != workspace || updated.ClientSymbols[1] != replacement {
		t.Fatalf("AddOrUpdateClientSymbols should replace by name and append replacement, got %#v", updated.ClientSymbols)
	}
	if same := updated.AddOrUpdateClientSymbols(); same != updated {
		t.Fatal("empty AddOrUpdateClientSymbols should return the same GlobalState")
	}
	if same := updated.AddOrUpdateClientSymbols(nil); same != updated {
		t.Fatal("nil AddOrUpdateClientSymbols should return the same GlobalState")
	}

	emptyName := NewParameterSymbol("", ScalarTypes.Bool.Type(), "")
	emptyGlobals := NewGlobalState(nil).WithClientSymbols([]Symbol{emptyName})
	if got := emptyGlobals.GetClientSymbol(""); got != emptyName {
		t.Fatalf("GetClientSymbol empty name = %#v", got)
	}
}

func TestGlobalStateAmbientSymbolsMirrorUpstreamLookupAndUpdates(t *testing.T) {
	globals := NewGlobalState(nil)
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	tenant := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")

	withAmbient := globals.WithAmbientSymbols([]Symbol{table, tenant})
	if got := withAmbient.GetAmbientSymbol("Events"); got != table {
		t.Fatalf("GetAmbientSymbol Events = %#v", got)
	}
	if got := withAmbient.GetAmbientSymbol("events"); got != nil {
		t.Fatalf("GetAmbientSymbol should be case-sensitive, got %#v", got)
	}
	if got := globals.GetAmbientSymbol("Events"); got != nil {
		t.Fatalf("original GlobalState ambient lookup mutated = %#v", got)
	}

	replacement := NewTableSymbol("Events", NewColumnSymbol("Score", ScalarTypes.Long.Type()))
	updated := withAmbient.AddOrUpdateAmbientSymbols(replacement)
	if got := updated.GetAmbientSymbol("Events"); got != replacement {
		t.Fatalf("updated Events = %#v", got)
	}
	if got := withAmbient.GetAmbientSymbol("Events"); got != table {
		t.Fatalf("previous GlobalState should keep old ambient symbol, got %#v", got)
	}
	if len(updated.AmbientSymbols) != 2 || updated.AmbientSymbols[0] != tenant || updated.AmbientSymbols[1] != replacement {
		t.Fatalf("AddOrUpdateAmbientSymbols should replace by name and append replacement, got %#v", updated.AmbientSymbols)
	}
	if same := updated.AddOrUpdateAmbientSymbols(); same != updated {
		t.Fatal("empty AddOrUpdateAmbientSymbols should return the same GlobalState")
	}
	if same := updated.AddOrUpdateAmbientSymbols(nil); same != updated {
		t.Fatal("nil AddOrUpdateAmbientSymbols should return the same GlobalState")
	}

	emptyName := NewParameterSymbol("", ScalarTypes.Bool.Type(), "")
	emptyGlobals := NewGlobalState(nil).WithAmbientSymbols([]Symbol{emptyName})
	if got := emptyGlobals.GetAmbientSymbol(""); got != emptyName {
		t.Fatalf("GetAmbientSymbol empty name = %#v", got)
	}
}

func TestGlobalStateLookupCachesAreConcurrentAndLastEntryWins(t *testing.T) {
	firstAggregate := NewScalarFunctionSymbol("duplicate_aggregate", ScalarTypes.Long.Type())
	lastAggregate := NewScalarFunctionSymbol("duplicate_aggregate", ScalarTypes.String.Type())
	firstPlugin := NewScalarFunctionSymbol("duplicate_plugin", ScalarTypes.Long.Type())
	lastPlugin := NewScalarFunctionSymbol("duplicate_plugin", ScalarTypes.String.Type())
	firstOption := NewOptionSymbolWithType("duplicate_option", "", ScalarTypes.Long, nil)
	lastOption := NewOptionSymbolWithType("duplicate_option", "", ScalarTypes.Bool, nil)
	firstAmbient := NewVariableSymbol("duplicate_ambient", ScalarTypes.Long.Type(), false, nil, nil)
	lastAmbient := NewVariableSymbol("duplicate_ambient", ScalarTypes.String.Type(), false, nil, nil)
	firstClient := NewParameterSymbol("duplicate_client", ScalarTypes.Long.Type(), "")
	lastClient := NewParameterSymbol("duplicate_client", ScalarTypes.String.Type(), "")
	firstAdd := NewOperatorSymbol(OperatorKindAdd, NewSignature(ScalarTypes.Long.Type()))
	lastAdd := NewOperatorSymbol(OperatorKindAdd, NewSignature(ScalarTypes.String.Type()))
	command := EngineCommands.All[0]

	globals := NewGlobalState(nil).
		WithAggregates([]*FunctionSymbol{firstAggregate, lastAggregate}).
		WithPlugIns([]*FunctionSymbol{firstPlugin, lastPlugin}).
		WithOptions([]*OptionSymbol{firstOption, lastOption}).
		WithAmbientSymbols([]Symbol{firstAmbient, lastAmbient}).
		WithClientSymbols([]Symbol{firstClient, lastClient}).
		WithOperators([]*OperatorSymbol{firstAdd, lastAdd})

	const workers = 32
	const iterations = 64
	failures := make(chan string, workers)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for iteration := 0; iteration < iterations; iteration++ {
				switch {
				case globals.GetAggregate("duplicate_aggregate") != lastAggregate:
					failures <- "aggregate lookup did not keep the last catalog entry"
					return
				case globals.GetPlugIn("duplicate_plugin") != lastPlugin:
					failures <- "plug-in lookup did not keep the last catalog entry"
					return
				case globals.GetOption("duplicate_option") != lastOption:
					failures <- "option lookup did not keep the last catalog entry"
					return
				case globals.GetAmbientSymbol("duplicate_ambient") != lastAmbient:
					failures <- "ambient lookup did not keep the last catalog entry"
					return
				case globals.GetClientSymbol("duplicate_client") != lastClient:
					failures <- "client lookup did not keep the last catalog entry"
					return
				case globals.GetOperator(OperatorKindAdd) != lastAdd:
					failures <- "operator lookup did not keep the last catalog entry"
					return
				case globals.GetCommand(command.SymbolName()) != command:
					failures <- "command lookup returned the wrong catalog entry"
					return
				}
			}
		}()
	}
	group.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}

	copy := globals.Copy()
	if copy.lookupCache == nil || copy.lookupCache == globals.lookupCache {
		t.Fatal("GlobalState copy should own an independent lazy lookup cache")
	}
}
