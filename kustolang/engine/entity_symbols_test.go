package kustolang

import "testing"

func TestClusterSymbolDatabases(t *testing.T) {
	table := NewTableSymbol("SigninLogs")
	primary := NewDatabaseSymbolWithAlternateName("db", "alias", []Symbol{table}, false)
	secondary := NewDatabaseSymbol("secondary")
	cluster := NewClusterSymbol("help", primary)

	if cluster.SymbolKind() != SymbolKindCluster || cluster.SymbolTabularity() != TabularityTabular {
		t.Fatalf("cluster metadata = kind:%s tabularity:%s", cluster.SymbolKind(), cluster.SymbolTabularity())
	}
	if cluster.IsOpen() {
		t.Fatal("cluster should not be open by default")
	}
	if cluster.GetDatabase("db") != primary || cluster.GetDatabase("alias") != primary || cluster.GetDatabase("") != nil {
		t.Fatal("database lookup failed")
	}
	if got := cluster.Databases(); len(got) != 1 || got[0] != primary {
		t.Fatalf("databases = %#v", got)
	}

	withSecond := cluster.AddDatabase(secondary)
	if got := withSecond.Databases(); len(got) != 2 || got[1] != secondary {
		t.Fatalf("after AddDatabase = %#v", got)
	}
	replacement := NewDatabaseSymbol("secondary", table)
	updated := withSecond.AddOrUpdateDatabase(replacement)
	if updated.GetDatabase("secondary") != replacement {
		t.Fatal("AddOrUpdateDatabase should replace same-name database")
	}
	removed := updated.RemoveDatabase(primary)
	if removed.GetDatabase("db") != nil || removed.GetDatabase("secondary") != replacement {
		t.Fatal("RemoveDatabase failed")
	}
	withoutReplacement := updated.RemoveDatabases([]*DatabaseSymbol{replacement})
	if withoutReplacement.GetDatabase("secondary") != nil || withoutReplacement.GetDatabase("db") != primary {
		t.Fatal("RemoveDatabases failed")
	}
	if !UnknownClusterSymbol.IsOpen() || len(UnknownClusterSymbol.Members()) != 0 {
		t.Fatalf("unknown cluster = open:%t members:%d", UnknownClusterSymbol.IsOpen(), len(UnknownClusterSymbol.Members()))
	}
}

func TestClusterSymbolGetDatabaseUsesLastMatchingName(t *testing.T) {
	firstPrimary := NewDatabaseSymbolWithAlternateName("primary-collision", "first", nil, false)
	lastAlternate := NewDatabaseSymbolWithAlternateName("second", "primary-collision", nil, false)
	firstAlternate := NewDatabaseSymbolWithAlternateName("third", "alternate-collision", nil, false)
	lastPrimary := NewDatabaseSymbolWithAlternateName("alternate-collision", "last", nil, false)
	cluster := NewClusterSymbol("help", firstPrimary, lastAlternate, firstAlternate, lastPrimary)

	if got := cluster.GetDatabase("primary-collision"); got != lastAlternate {
		t.Fatalf("primary/alternate collision = %#v, want last database %#v", got, lastAlternate)
	}
	if got := cluster.GetDatabase("alternate-collision"); got != lastPrimary {
		t.Fatalf("alternate/primary collision = %#v, want last database %#v", got, lastPrimary)
	}
}

func TestEntityGroupSymbolAndElement(t *testing.T) {
	if got := EntityGroupBodyFromDefinition(""); got != "entity_group []" {
		t.Fatalf("empty entity group body = %q", got)
	}
	if got := EntityGroupBodyFromDefinition("[cluster('c').database('d')]"); got != "entity_group [cluster('c').database('d')]" {
		t.Fatalf("bracket entity group body = %q", got)
	}
	if got := EntityGroupBodyFromDefinition(`"cluster('c').database('d')"`); got != "entity_group [cluster('c').database('d')]" {
		t.Fatalf("quoted entity group body = %q", got)
	}
	if got := EntityGroupBodyFromDefinition("entity_group [database('d')]"); got != "entity_group [database('d')]" {
		t.Fatalf("existing entity group body = %q", got)
	}

	defined := NewEntityGroupSymbolFromDefinition("prod", "database('d')", "Production")
	if defined.SymbolKind() != SymbolKindEntityGroup || defined.SymbolTabularity() != TabularityNone {
		t.Fatalf("entity group metadata = kind:%s tabularity:%s", defined.SymbolKind(), defined.SymbolTabularity())
	}
	if defined.Definition() != "database('d')" || defined.Description() != "Production" {
		t.Fatalf("entity group text = def:%q desc:%q", defined.Definition(), defined.Description())
	}
	if defined.Signature() == nil || defined.Signature().Symbol != defined || defined.Signature().Body != "entity_group [database('d')]" {
		t.Fatalf("entity group signature = %#v", defined.Signature())
	}

	emptyTable := NewTableSymbol("Empty")
	tableWithMember := NewTableSymbol("Events", NewColumnSymbol("TenantId", ScalarTypes.String.Type()))
	group := NewEntityGroupSymbol("mixed", emptyTable, tableWithMember)
	element := NewEntityGroupElementSymbol("current", group, nil)
	if element.EntityGroup != group || element.UnderlyingSymbol != tableWithMember {
		t.Fatalf("entity group element underlying = group:%#v underlying:%#v", element.EntityGroup, element.UnderlyingSymbol)
	}
	if element.SymbolKind() != SymbolKindEntityGroupElement || element.SymbolTabularity() != TabularityTabular {
		t.Fatalf("entity group element metadata = kind:%s tabularity:%s", element.SymbolKind(), element.SymbolTabularity())
	}
	specials := element.Members()
	if len(specials) != 2 || specials[0].SymbolName() != "$current_database" || specials[1].SymbolName() != "$current_cluster_endpoint" {
		t.Fatalf("special members = %#v", specials)
	}
	if second := element.Members(); len(second) != 2 || second[0] != specials[0] || second[1] != specials[1] {
		t.Fatalf("special members should preserve upstream static symbol identity: first=%#v second=%#v", specials, second)
	}

	orphan := NewEntityGroupElementSymbol("orphan", nil, nil)
	if orphan.EntityGroup != EmptyEntityGroupSymbol || orphan.UnderlyingSymbol != Error || orphan.SymbolTabularity() != TabularityNone {
		t.Fatalf("orphan entity group element = %#v", orphan)
	}
}

func TestGraphModelAndSnapshotSymbols(t *testing.T) {
	edge := NewComputedSignature("Edges | project Source, Target", TabularityUnknown)
	node := NewComputedSignature("Nodes | project Id", TabularityUnknown)
	snapshot := NewGraphSnapshotSymbol("daily")
	model := NewGraphModelSymbol("model", []*Signature{edge}, []*Signature{node}, []*GraphSnapshotSymbol{snapshot})

	if model.SymbolKind() != SymbolKindGraphModel || model.SymbolTabularity() != TabularityNone {
		t.Fatalf("graph model metadata = kind:%s tabularity:%s", model.SymbolKind(), model.SymbolTabularity())
	}
	if edge.Symbol != model || node.Symbol != model || snapshot.Model != model {
		t.Fatal("graph model should back-reference signatures and snapshots")
	}
	if got, ok := model.TryGetSnapshot("daily"); !ok || got != snapshot {
		t.Fatalf("snapshot lookup = %#v, %t", got, ok)
	}
	if got, ok := model.TryGetSnapshot("missing"); ok || got != nil {
		t.Fatalf("missing snapshot lookup = %#v, %t", got, ok)
	}
	if members := model.Members(); len(members) != 1 || members[0] != snapshot {
		t.Fatalf("graph model members = %#v", members)
	}
	if snapshot.SymbolKind() != SymbolKindGraphSnapshot || snapshot.SymbolTabularity() != TabularityNone {
		t.Fatalf("snapshot metadata = kind:%s tabularity:%s", snapshot.SymbolKind(), snapshot.SymbolTabularity())
	}

	fromNames := NewGraphModelSymbolFromNames("named", []string{"E"}, []string{"N"}, []string{"latest"})
	if len(fromNames.Edges) != 1 || fromNames.Edges[0].Body != "E" || fromNames.Snapshots[0].Model != fromNames {
		t.Fatalf("name-based graph model = %#v", fromNames)
	}
}

func TestGroupAndPatternSymbols(t *testing.T) {
	table := NewTableSymbol("T")
	scalar := ScalarTypes.String.Type()
	group := NewGroupSymbol(table, scalar)
	if group.SymbolName() != "group" || group.SymbolKind() != SymbolKindGroup || group.SymbolTabularity() != TabularityTabular {
		t.Fatalf("group metadata = name:%q kind:%s tabularity:%s", group.SymbolName(), group.SymbolKind(), group.SymbolTabularity())
	}
	if members := group.Members(); len(members) != 2 || members[0] != table || members[1] != scalar {
		t.Fatalf("group members = %#v", members)
	}

	sourceParam := NewParameter("source", ScalarTypes.String.Type())
	pathParam := NewParameter("path", ScalarTypes.String.Type())
	patternSignature := NewPatternSignature([]string{"SecurityEvent"}, "/detections", "T | take 1")
	pattern := NewPatternSymbol("detections", []*Parameter{sourceParam}, pathParam, []*PatternSignature{patternSignature})
	if pattern.SymbolKind() != SymbolKindPattern || pattern.SymbolTabularity() != TabularityTabular {
		t.Fatalf("pattern metadata = kind:%s tabularity:%s", pattern.SymbolKind(), pattern.SymbolTabularity())
	}
	if patternSignature.Symbol != pattern {
		t.Fatal("pattern signature should back-reference pattern")
	}
	signature := patternSignature.Signature()
	if signature == nil || signature.Symbol != pattern || signature.Body != "T | take 1" {
		t.Fatalf("materialized pattern signature = %#v", signature)
	}
	if len(signature.Parameters) != 2 || signature.Parameters[0] != sourceParam || signature.Parameters[1] != pathParam {
		t.Fatalf("pattern signature parameters = %#v", signature.Parameters)
	}
}

func TestGetResultTypeMirrorsUpstreamSymbolCases(t *testing.T) {
	stringType := ScalarTypes.String.Type()
	longType := ScalarTypes.Long.Type()
	boolType := ScalarTypes.Bool.Type()

	accountColumn := NewColumnSymbol("Account", stringType)
	if got := GetResultType(accountColumn); got != stringType {
		t.Fatalf("column result type = %#v, want %#v", got, stringType)
	}

	variable := NewVariableSymbol("Threshold", longType, false, nil, nil)
	if got := GetResultType(variable); got != longType {
		t.Fatalf("variable result type = %#v, want %#v", got, longType)
	}

	parameter := NewParameterSymbol("enabled", boolType, "")
	if got := GetResultType(parameter); got != boolType {
		t.Fatalf("parameter result type = %#v, want %#v", got, boolType)
	}

	table := NewTableSymbol("Events", NewColumnSymbol("TenantId", stringType))
	entityGroup := NewEntityGroupSymbol("mixed", table)
	element := NewEntityGroupElementSymbol("current", entityGroup, nil)
	if got := GetResultType(element); got != table {
		t.Fatalf("entity group element result type = %#v, want %#v", got, table)
	}

	if got := GetResultType(table); got != table {
		t.Fatalf("table result type = %#v, want %#v", got, table)
	}
	if got := GetResultType(NewGroupSymbol(table)); got != table {
		t.Fatalf("single table group result type = %#v, want %#v", got, table)
	}
	if got := GetResultType(NewGroupSymbol(accountColumn)); got != stringType {
		t.Fatalf("single column group result type = %#v, want %#v", got, stringType)
	}

	scoreColumn := NewColumnSymbol("Score", longType)
	multi := GetResultType(NewGroupSymbol(accountColumn, scoreColumn))
	multiGroup, ok := multi.(*GroupSymbol)
	if !ok {
		t.Fatalf("multi-column group result type = %T, want *GroupSymbol", multi)
	}
	members := multiGroup.Members()
	if len(members) != 2 || members[0] != stringType || members[1] != longType {
		t.Fatalf("multi-column group members = %#v", members)
	}

	option := NewOptionSymbolWithType("query_language", "", ScalarTypes.String, nil)
	if got := GetResultType(NewGroupSymbol(option)); got != nil {
		t.Fatalf("non-type group result type = %#v, want nil", got)
	}
}

func TestAdvancedSymbolsMatchUpstreamFlags(t *testing.T) {
	cluster := NewClusterSymbol("help")
	entityGroup := NewEntityGroupSymbol("prod")
	entityElement := NewEntityGroupElementSymbol("current", entityGroup, nil)
	graph := NewGraphSymbol("g", EmptyTableSymbol, nil)
	model := NewGraphModelSymbolFromNames("model", nil, nil, []string{"daily"})
	snapshot := model.Snapshots[0]
	pattern := NewNamedPatternSymbol("pattern")

	if !SymbolMatches(cluster, SymbolMatchCluster) || SymbolMatches(cluster, SymbolMatchDefault) {
		t.Fatal("cluster matching failed")
	}
	if !SymbolMatches(entityGroup, SymbolMatchEntityGroup) || !SymbolMatches(entityGroup, SymbolMatchDefault) {
		t.Fatal("entity group matching failed")
	}
	if !SymbolMatches(entityElement, SymbolMatchEntityGroupElement) || !SymbolMatches(entityElement, SymbolMatchDefault) {
		t.Fatal("entity group element matching failed")
	}
	if !SymbolMatches(graph, SymbolMatchGraph) || !SymbolMatches(graph, SymbolMatchDefault) {
		t.Fatal("graph matching failed")
	}
	if !SymbolMatches(model, SymbolMatchGraphModel) || !SymbolMatches(model, SymbolMatchAny) || SymbolMatches(model, SymbolMatchDefault) {
		t.Fatal("graph model matching failed")
	}
	if !SymbolMatches(snapshot, SymbolMatchGraphSnapshot) || !SymbolMatches(snapshot, SymbolMatchAny) || SymbolMatches(snapshot, SymbolMatchDefault) {
		t.Fatal("graph snapshot matching failed")
	}
	if !SymbolMatches(pattern, SymbolMatchFunction) || !SymbolMatches(pattern, SymbolMatchFunction|SymbolMatchTabular) {
		t.Fatal("pattern should match function filters")
	}

	entityGroupVariable := NewVariableSymbol("eg", entityGroup, false, nil, nil)
	graphVariable := NewVariableSymbol("gv", graph, false, nil, nil)
	entityElementVariable := NewVariableSymbol("ee", entityElement, false, nil, nil)
	if !SymbolMatches(entityGroupVariable, SymbolMatchEntityGroup) {
		t.Fatal("entity group variable matching failed")
	}
	if !SymbolMatches(graphVariable, SymbolMatchGraph) {
		t.Fatal("graph variable matching failed")
	}
	if !SymbolMatches(entityElementVariable, SymbolMatchEntityGroupElement) {
		t.Fatal("entity group element variable matching failed")
	}
}
