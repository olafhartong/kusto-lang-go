package kustolang

import "testing"

func TestFunctionSignatureArgumentCounts(t *testing.T) {
	signature := NewSignature(
		ScalarTypes.Bool.Type(),
		NewParameter("required", ScalarTypes.String.Type()),
		NewParameter("optional", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1)),
	)
	if signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 2 {
		t.Fatalf("argument bounds = %d..%d", signature.MinArgumentCount, signature.MaxArgumentCount)
	}
	if !signature.IsValidArgumentCount(1) || !signature.IsValidArgumentCount(2) || signature.IsValidArgumentCount(3) {
		t.Fatalf("unexpected argument count validation")
	}
	if !signature.HasOptionalParameters {
		t.Fatal("expected optional parameter flag")
	}
}

func TestExternalAndMaterializedViewTransformsPreserveConcreteIdentity(t *testing.T) {
	base := NewTableSymbol("Input").WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	extra := NewColumnSymbol("Extra", ScalarTypes.Long.Type())
	external := NewExternalTableSymbol("Archive").
		WithInheritableProperties(base).
		WithDescription("archive").
		WithName("Archive2").
		AddColumns(extra)
	if external == nil || external.SymbolName() != "Archive2" || external.Description() != "archive" ||
		!external.IsOpen() || !external.IsSerialized() || !external.IsSorted() || external.GetColumn("Extra") != extra {
		t.Fatalf("external transform = %#v", external)
	}
	view := NewMaterializedViewSymbol("Daily", nil, "Events | take 1").
		WithInheritableProperties(base).
		WithDescription("daily").
		WithName("Daily2").
		AddColumns(extra)
	if view == nil || view.SymbolName() != "Daily2" || view.Description() != "daily" || view.MaterializedViewQuery() != "Events | take 1" ||
		!view.IsOpen() || !view.IsSerialized() || !view.IsSorted() || view.GetColumn("Extra") != extra {
		t.Fatalf("materialized view transform = %#v", view)
	}
	database := NewDatabaseSymbol("db", external, view)
	if database.GetExternalTable("Archive2") != external || database.GetMaterializedView("Daily2") != view {
		t.Fatalf("transformed entities lost concrete identity: external %#v view %#v", database.ExternalTables(), database.MaterializedViews())
	}
}

func TestTabularityValuesAndSymbolHelpersMirrorUpstream(t *testing.T) {
	values := []Tabularity{
		TabularityNone,
		TabularityScalar,
		TabularityTabular,
		TabularityOther,
		TabularityUnknown,
		TabularityUnspecified,
	}
	want := []string{"None", "Scalar", "Tabular", "Other", "Unknown", "Unspecified"}
	for i, value := range values {
		if string(value) != want[i] {
			t.Fatalf("tabularity[%d] = %q, want %q", i, value, want[i])
		}
	}

	unknown := NewBaseSymbol("unknown")
	if !IsScalar(unknown) || !IsTabular(unknown) {
		t.Fatal("upstream Symbol helpers treat unknown tabularity as both scalar and tabular")
	}
	if IsScalar(NewTableSymbol("Rows")) || !IsTabular(NewTableSymbol("Rows")) {
		t.Fatal("table tabularity helper mismatch")
	}
	if !IsScalar(ScalarTypes.String.Type()) || IsTabular(ScalarTypes.String.Type()) {
		t.Fatal("scalar tabularity helper mismatch")
	}
}

func TestTableTypePreservesOwningSymbolBehavior(t *testing.T) {
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbol("Events", column)
	typ := &table.TypeSymbol

	if got, ok := TableSymbolForType(typ); !ok || got != table {
		t.Fatalf("TableSymbolForType = %#v, %t, want original table", got, ok)
	}
	if typ.SymbolKind() != SymbolKindTable || typ.SymbolTabularity() != TabularityTabular || IsScalar(typ) || !IsTabular(typ) {
		t.Fatalf("table type behavior = kind:%s tabularity:%s", typ.SymbolKind(), typ.SymbolTabularity())
	}
	if members := typ.Members(); len(members) != 1 || members[0] != column {
		t.Fatalf("table type members = %#v", members)
	}
	if members := FindMembers(typ, "Account", SymbolMatchColumn, false); len(members) != 1 || members[0] != column {
		t.Fatalf("table type member lookup = %#v", members)
	}

	copied := *typ
	if got, ok := TableSymbolForType(&copied); ok || got != nil {
		t.Fatalf("copied type retained table identity = %#v, %t", got, ok)
	}
	if copied.SymbolKind() != SymbolKindType || copied.SymbolTabularity() != TabularityScalar {
		t.Fatalf("detached copied type behavior = kind:%s tabularity:%s", copied.SymbolKind(), copied.SymbolTabularity())
	}
}

func TestTableTypeBackReferenceSurvivesConstructorsAndClones(t *testing.T) {
	column := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	base := NewTableSymbol("Base", column)
	view := NewMaterializedViewSymbol("View", []*ColumnSymbol{column}, "Base | summarize count()")
	external := NewExternalTableSymbol("External", column)
	stored := NewStoredQueryResultSymbol("Stored", column)

	tests := []struct {
		name string
		typ  *TypeSymbol
		want *TableSymbol
		kind SymbolKind
	}{
		{"base", &base.TypeSymbol, base, SymbolKindTable},
		{"materialized view", &view.TypeSymbol, &view.TableSymbol, SymbolKindMaterializedView},
		{"external", &external.TypeSymbol, &external.TableSymbol, SymbolKindTable},
		{"stored query result", &stored.TypeSymbol, &stored.TableSymbol, SymbolKindStoredQueryResult},
		{"renamed", &base.WithName("Renamed").TypeSymbol, nil, SymbolKindTable},
		{"columns clone", &base.WithColumns(nil).TypeSymbol, nil, SymbolKindTable},
		{"view clone", &view.WithMaterializedViewKind(MaterializedViewKindOther).TypeSymbol, nil, SymbolKindMaterializedView},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := TableSymbolForType(test.typ)
			if !ok || got == nil {
				t.Fatalf("TableSymbolForType = %#v, %t", got, ok)
			}
			if test.want != nil && got != test.want {
				t.Fatalf("owner = %#v, want %#v", got, test.want)
			}
			if got != test.typ.table || &got.TypeSymbol != test.typ || test.typ.SymbolKind() != test.kind || test.typ.SymbolTabularity() != TabularityTabular {
				t.Fatalf("owner identity or behavior mismatch: owner=%p type=%p kind=%s tabularity=%s", got, test.typ, test.typ.SymbolKind(), test.typ.SymbolTabularity())
			}
		})
	}
}

func TestAnalyzedKustoCodeTableResultTypeRetainsTable(t *testing.T) {
	code := ParseAndAnalyze("print x=1")
	if code.ResultType == nil {
		t.Fatal("analyzed query has no result type")
	}
	table, ok := TableSymbolForType(code.ResultType)
	if !ok || table == nil || code.ResultType.SymbolTabularity() != TabularityTabular {
		t.Fatalf("result type = %#v, owner=%#v, tabularity=%s", code.ResultType, table, code.ResultType.SymbolTabularity())
	}
	if column := table.GetColumn("x"); column == nil || column.Type != ScalarTypes.Long.Type() {
		t.Fatalf("result columns = %#v", table.Columns())
	}
}

func TestDatabaseTablesFunctionsAndGlobalState(t *testing.T) {
	col := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbol("SecurityEvent", col)
	fn := NewScalarFunctionSymbol("MyScalar", ScalarTypes.Bool.Type(), NewParameter("x", ScalarTypes.String.Type()))
	db := NewDatabaseSymbol("db", table, fn)
	globals := NewGlobalState(db)
	if db.GetTable("SecurityEvent") != table {
		t.Fatal("table lookup failed")
	}
	if db.GetFunction("MyScalar") != fn {
		t.Fatal("function lookup failed")
	}
	if !globals.IsDatabaseTable(table) || !globals.IsDatabaseFunction(fn) {
		t.Fatal("global membership failed")
	}
	if globals.GetTable(col) != table {
		t.Fatal("column reverse table lookup failed")
	}
}

func TestColumnSymbolConstructorMatchesUpstreamDefaults(t *testing.T) {
	if got := NewColumnSymbol("MissingType", nil); got.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("nil column type = %#v, want unknown", got.Type)
	}
	if got := NewColumnSymbol("ErrorType", &Error.TypeSymbol); got.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("error column type = %#v, want unknown", got.Type)
	}
	if got := NewColumnSymbol("Account", ScalarTypes.String.Type()); got.Type != ScalarTypes.String.Type() || got.Description != "" || len(got.OriginalColumns) != 0 || len(got.Examples) != 0 {
		t.Fatalf("column defaults = %#v", got)
	}

	original := NewColumnSymbol("Original", ScalarTypes.String.Type())
	derived := NewColumnSymbolFull("Derived", ScalarTypes.Long.Type(), "desc", []*ColumnSymbol{original}, nil, []string{"1"})
	flattened := NewColumnSymbolFull("Flattened", ScalarTypes.Long.Type(), "", []*ColumnSymbol{derived}, nil, nil)
	if len(flattened.OriginalColumns) != 1 || flattened.OriginalColumns[0] != original {
		t.Fatalf("truly original columns = %#v", flattened.OriginalColumns)
	}
}

func TestTableColumnLookupAndMatchingColumns(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	tenant := NewColumnSymbol("TenantId", ScalarTypes.String.Type())
	accountDuplicate := NewColumnSymbol("Account", ScalarTypes.Long.Type())
	resultType := NewColumnSymbol("ResultType", ScalarTypes.Int.Type())
	table := NewTableSymbol("SigninLogs", account, tenant, accountDuplicate, resultType)

	if got, ok := table.TryGetColumn("Account"); !ok || got != account {
		t.Fatalf("TryGetColumn Account = %#v, %t", got, ok)
	}
	if got := table.GetColumn("TenantId"); got != tenant {
		t.Fatalf("GetColumn TenantId = %#v", got)
	}
	if got, ok := table.TryGetColumn("Missing"); ok || got != nil {
		t.Fatalf("TryGetColumn Missing = %#v, %t", got, ok)
	}
	if got := table.GetColumn("Missing"); got != nil {
		t.Fatalf("GetColumn Missing = %#v", got)
	}

	exact := table.GetMatchingColumns("Account")
	if len(exact) != 1 || exact[0] != account {
		t.Fatalf("exact matching columns = %#v", exact)
	}
	wildcard := table.GetMatchingColumns("A**t")
	if len(wildcard) != 2 || wildcard[0] != account || wildcard[1] != accountDuplicate {
		t.Fatalf("wildcard matching columns = %#v", wildcard)
	}
	suffix := table.GetMatchingColumns("*Id")
	if len(suffix) != 1 || suffix[0] != tenant {
		t.Fatalf("suffix matching columns = %#v", suffix)
	}
	if KustoFactsMatches("", "Account") || !KustoFactsMatches("*Type", "ResultType") || KustoFactsMatches("account", "Account") {
		t.Fatal("KustoFactsMatches wildcard behavior diverged")
	}

	members := FindMembers(table, "Account", SymbolMatchColumn, false)
	if len(members) != 1 || members[0] != account {
		t.Fatalf("table member lookup = %#v", members)
	}
	if got := FindMembers(table, "account", SymbolMatchColumn, true); len(got) != 1 || got[0] != account {
		t.Fatalf("case-insensitive table member lookup = %#v", got)
	}
	if got := FindMembersByMatch(table, SymbolMatchColumn); len(got) != 4 {
		t.Fatalf("all table column members = %#v", got)
	}
}

func TestTableCloneHelpersMirrorUpstreamShape(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbolWithDescription("SigninLogs", []*ColumnSymbol{account}, "sign-ins").
		WithIsSerialized(true).
		WithIsOpen(true)

	if table.WithName("SigninLogs") != table || table.WithDescription("sign-ins") != table || table.WithColumns(table.Columns()) != table || table.AddColumns() != table {
		t.Fatal("unchanged table helpers should preserve identity")
	}

	renamed := table.WithName("AuditLogs")
	if renamed == table || renamed.SymbolName() != "AuditLogs" || table.SymbolName() != "SigninLogs" || !renamed.IsSerialized() || !renamed.IsOpen() {
		t.Fatalf("WithName clone = %#v original=%q", renamed, table.SymbolName())
	}
	described := table.WithDescription("audit")
	if described == table || described.Description() != "audit" || table.Description() != "sign-ins" {
		t.Fatalf("WithDescription clone = %#v original desc=%q", described, table.Description())
	}

	tenant := NewColumnSymbol("TenantId", ScalarTypes.String.Type())
	replaced := table.WithColumns([]*ColumnSymbol{tenant})
	if replaced == table || len(replaced.Columns()) != 1 || replaced.Columns()[0] != tenant || len(table.Columns()) != 1 || table.Columns()[0] != account {
		t.Fatalf("WithColumns clone = %#v original=%#v", replaced.Columns(), table.Columns())
	}
	added := table.AddColumns(tenant)
	if added == table || len(added.Columns()) != 2 || added.Columns()[0] != account || added.Columns()[1] != tenant || len(table.Columns()) != 1 {
		t.Fatalf("AddColumns clone = %#v original=%#v", added.Columns(), table.Columns())
	}

	source := ParseExpression("Account")
	sourced := table.WithSource(source)
	if sourced == table || len(sourced.Columns()) != 1 || sourced.Columns()[0].Source != source || table.Columns()[0].Source != nil {
		t.Fatalf("WithSource clone = %#v original=%#v", sourced.Columns(), table.Columns())
	}
	if sourced.WithSource(source) != sourced {
		t.Fatal("WithSource should preserve identity when columns already reference the source")
	}

	stateSource := NewTableSymbol("State").WithIsSorted(true)
	inherited := table.WithInheritableProperties(stateSource)
	if inherited == table || inherited.IsOpen() || inherited.IsSerialized() || !inherited.IsSorted() {
		t.Fatalf("WithInheritableProperties = open:%t serialized:%t sorted:%t", inherited.IsOpen(), inherited.IsSerialized(), inherited.IsSorted())
	}
	if table.WithInheritableProperties(table) != table {
		t.Fatal("unchanged inheritable state should preserve identity")
	}
}

func TestTableEquivalenceMirrorsUpstreamRules(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	base := NewTableSymbolWithDescription("Events", []*ColumnSymbol{account, score}, "events").
		WithIsSerialized(true).
		WithIsOpen(true)
	same := NewTableSymbolWithDescription("Events", []*ColumnSymbol{
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	}, "events").WithIsSerialized(true).WithIsOpen(true)

	if !AreTablesEquivalent(base, same) || !AreTableResultsEquivalent(base, same) || !AreTableColumnsEquivalent(base, same) {
		t.Fatal("matching table shape should be fully equivalent")
	}
	if !AreTablesEquivalent(base, base) {
		t.Fatal("same table pointer should be equivalent")
	}

	renamed := same.WithName("OtherEvents")
	if AreTablesEquivalent(base, renamed) || !AreTableResultsEquivalent(base, renamed) {
		t.Fatal("table result equivalence should ignore table name, full equivalence should not")
	}
	redescribed := same.WithDescription("other")
	if AreTablesEquivalent(base, redescribed) || !AreTableResultsEquivalent(base, redescribed) {
		t.Fatal("table result equivalence should ignore description, full equivalence should not")
	}

	unsorted := same.WithIsSerialized(false)
	if AreTableResultsEquivalent(base, unsorted) || !AreTableColumnsEquivalent(base, unsorted) {
		t.Fatal("result equivalence should include serialized/sorted/open state, column equivalence should not")
	}
	reordered := same.WithColumns([]*ColumnSymbol{score, account})
	if AreTableColumnsEquivalent(base, reordered) {
		t.Fatal("column equivalence should be ordered")
	}
	wrongType := same.WithColumns([]*ColumnSymbol{
		NewColumnSymbol("Account", ScalarTypes.Long.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	})
	if AreTableColumnsEquivalent(base, wrongType) {
		t.Fatal("column equivalence should compare type identity")
	}
	wrongName := same.WithColumns([]*ColumnSymbol{
		NewColumnSymbol("TenantId", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	})
	if AreTableColumnsEquivalent(base, wrongName) {
		t.Fatal("column equivalence should compare column names")
	}

	external := NewExternalTableSymbolWithDescription("Events", []*ColumnSymbol{account, score}, "events")
	external.TableSymbol = *external.TableSymbol.WithIsSerialized(true).WithIsOpen(true)
	if AreTablesEquivalent(base, external) || !AreTableResultsEquivalent(base, external) {
		t.Fatal("full equivalence should include table classification, result equivalence should not")
	}
	materialized := NewMaterializedViewSymbol("Events", []*ColumnSymbol{account, score}, "Events | take 1", "events")
	materialized.TableSymbol = *materialized.TableSymbol.WithIsSerialized(true).WithIsOpen(true)
	if AreTablesEquivalent(base, materialized) || !AreTableResultsEquivalent(base, materialized) {
		t.Fatal("materialized view classification should only affect full equivalence")
	}
}

func TestTableSubclassesAndDatabaseViews(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	normal := NewTableSymbol("SigninLogs", account)
	external := NewExternalTableSymbolWithDescription("Archive", []*ColumnSymbol{account}, "archived sign-ins")
	flagExternal := NewTableSymbol("LegacyArchive", account).WithIsExternal(true)
	materialized := NewMaterializedViewSymbol("DailySignin", []*ColumnSymbol{account}, "SigninLogs | summarize count() by Account", "daily sign-ins")
	flagMaterialized := NewTableSymbol("LegacyDailySignin", account).WithIsMaterializedView(true)
	stored := NewStoredQueryResultSymbol("RecentSignin", account)
	flagStored := NewTableSymbol("LegacyRecentSignin", account).WithIsStoredQueryResult(true)
	db := NewDatabaseSymbol("db", normal, external, flagExternal, materialized, flagMaterialized, stored, flagStored)

	if external.SymbolKind() != SymbolKindTable || external.SymbolTabularity() != TabularityTabular || !external.IsExternal() || external.Description() != "archived sign-ins" {
		t.Fatalf("external table metadata = kind:%s tabularity:%s external:%t desc:%q", external.SymbolKind(), external.SymbolTabularity(), external.IsExternal(), external.Description())
	}
	if materialized.SymbolKind() != SymbolKindMaterializedView || materialized.SymbolTabularity() != TabularityTabular || !materialized.IsMaterializedView() || materialized.MaterializedViewQuery() == "" || materialized.Description() != "daily sign-ins" {
		t.Fatalf("materialized view metadata = kind:%s tabularity:%s materialized:%t query:%q desc:%q", materialized.SymbolKind(), materialized.SymbolTabularity(), materialized.IsMaterializedView(), materialized.MaterializedViewQuery(), materialized.Description())
	}
	if materialized.MaterializedViewKind() != MaterializedViewKindUnknown {
		t.Fatalf("materialized view default kind = %s", materialized.MaterializedViewKind())
	}
	downsampling := materialized.WithMaterializedViewKind(MaterializedViewKindDownsampling)
	if downsampling == materialized || downsampling.MaterializedViewKind() != MaterializedViewKindDownsampling || materialized.MaterializedViewKind() != MaterializedViewKindUnknown {
		t.Fatalf("materialized view kind clone = clone:%s original:%s same:%t", downsampling.MaterializedViewKind(), materialized.MaterializedViewKind(), downsampling == materialized)
	}
	if downsampling.MaterializedViewQuery() != materialized.MaterializedViewQuery() || downsampling.Description() != materialized.Description() || downsampling.Columns()[0] != account {
		t.Fatalf("materialized view kind clone did not preserve metadata")
	}
	if downsampling.WithMaterializedViewKind(MaterializedViewKindDownsampling) != downsampling {
		t.Fatalf("unchanged materialized view kind should preserve identity")
	}
	if stored.SymbolKind() != SymbolKindStoredQueryResult || stored.SymbolTabularity() != TabularityTabular || !stored.IsStoredQueryResult() {
		t.Fatalf("stored query result metadata = kind:%s tabularity:%s stored:%t", stored.SymbolKind(), stored.SymbolTabularity(), stored.IsStoredQueryResult())
	}
	if got := stored.Columns(); len(got) != 1 || got[0] != account {
		t.Fatalf("stored query result columns = %#v", got)
	}
	if flagExternal.SymbolKind() != SymbolKindTable || flagStored.SymbolKind() != SymbolKindStoredQueryResult || flagMaterialized.SymbolKind() != SymbolKindMaterializedView {
		t.Fatalf("flagged table kinds = external:%s stored:%s materialized:%s", flagExternal.SymbolKind(), flagStored.SymbolKind(), flagMaterialized.SymbolKind())
	}
	if SchemaDisplayText(external) != "(Account: string)" || SchemaDisplayText(materialized) != "(Account: string)" || SchemaDisplayText(stored) != "(Account: string)" {
		t.Fatalf("subclass schema display = external:%q materialized:%q stored:%q", SchemaDisplayText(external), SchemaDisplayText(materialized), SchemaDisplayText(stored))
	}

	tables := db.Tables()
	if len(tables) != 1 || tables[0] != normal {
		t.Fatalf("database tables = %#v", tables)
	}
	externalTables := db.ExternalTables()
	if len(externalTables) != 1 || externalTables[0] != external {
		t.Fatalf("database external tables = %#v", externalTables)
	}
	materializedViews := db.MaterializedViews()
	if len(materializedViews) != 1 || materializedViews[0] != materialized {
		t.Fatalf("database materialized views = %#v", materializedViews)
	}
	storedResults := db.StoredQueryResults()
	if len(storedResults) != 1 || storedResults[0] != stored {
		t.Fatalf("database stored query results = %#v", storedResults)
	}
	if db.GetExternalTable("Archive") != external || db.GetExternalTable("LegacyArchive") != nil {
		t.Fatalf("external table lookup failed")
	}
	if db.GetMaterializedView("DailySignin") != materialized || db.GetMaterializedView("LegacyDailySignin") != nil {
		t.Fatalf("materialized view lookup failed")
	}
	if db.GetStoredQueryResult("RecentSignin") != stored || db.GetStoredQueryResult("LegacyRecentSignin") != nil {
		t.Fatalf("stored query result lookup failed")
	}
	if db.GetTable("SigninLogs") != normal || db.GetTable("Archive") != nil || db.GetTable("LegacyArchive") != nil || db.GetTable("DailySignin") != nil || db.GetTable("LegacyDailySignin") != nil || db.GetTable("RecentSignin") != nil || db.GetTable("LegacyRecentSignin") != nil {
		t.Fatalf("table lookup should only return ordinary tables")
	}
	if db.GetAnyTable("SigninLogs") != normal || db.GetAnyTable("Archive") != &external.TableSymbol || db.GetAnyTable("DailySignin") != &materialized.TableSymbol || db.GetAnyTable("RecentSignin") != nil {
		t.Fatalf("any-table lookup failed")
	}
	if db.GetAnyTableSymbol("SigninLogs") != normal || db.GetAnyTableSymbol("Archive") != external || db.GetAnyTableSymbol("DailySignin") != materialized || db.GetAnyTableSymbol("RecentSignin") != nil {
		t.Fatalf("identity-preserving any-table lookup failed")
	}
}

func TestTableSchemaConstructorsMirrorUpstream(t *testing.T) {
	table, err := NewTableSymbolFromSchema("Events", "Account:string, ['bad name']: long", "events")
	if err != nil {
		t.Fatalf("NewTableSymbolFromSchema returned error: %v", err)
	}
	if table.SymbolName() != "Events" || table.Description() != "events" || SchemaDisplayText(table) != "(Account: string, ['bad name']: long)" {
		t.Fatalf("table schema constructor = name:%q desc:%q schema:%q", table.SymbolName(), table.Description(), SchemaDisplayText(table))
	}

	external, err := NewExternalTableSymbolFromSchema("Archive", "Account:string, ['bad name']: long", "archive")
	if err != nil {
		t.Fatalf("NewExternalTableSymbolFromSchema returned error: %v", err)
	}
	if external.SymbolKind() != SymbolKindTable || !external.IsExternal() || external.Description() != "archive" || SchemaDisplayText(external) != SchemaDisplayText(table) {
		t.Fatalf("external schema constructor = kind:%s external:%t desc:%q schema:%q", external.SymbolKind(), external.IsExternal(), external.Description(), SchemaDisplayText(external))
	}

	materialized, err := NewMaterializedViewSymbolFromSchema("DailySignin", "(Account:string)", "SigninLogs | summarize count() by Account", "daily")
	if err != nil {
		t.Fatalf("NewMaterializedViewSymbolFromSchema returned error: %v", err)
	}
	if materialized.SymbolKind() != SymbolKindMaterializedView || !materialized.IsMaterializedView() || materialized.MaterializedViewQuery() == "" || materialized.Description() != "daily" || SchemaDisplayText(materialized) != "(Account: string)" {
		t.Fatalf("materialized schema constructor = kind:%s materialized:%t query:%q desc:%q schema:%q", materialized.SymbolKind(), materialized.IsMaterializedView(), materialized.MaterializedViewQuery(), materialized.Description(), SchemaDisplayText(materialized))
	}

	aliased, err := NewExternalTableSymbolFromSchema("Aliased", "Account:float")
	if err != nil || len(aliased.Columns()) != 1 || aliased.Columns()[0].Type != ScalarTypes.Real.Type() {
		t.Fatalf("schema constructor scalar alias = %#v, %v", aliased, err)
	}
}

func TestDatabaseTypedMemberViewsAndCloneHelpers(t *testing.T) {
	column := NewColumnSymbol("TenantId", ScalarTypes.String.Type())
	external := NewExternalTableSymbol("Shared", column)
	function := NewScalarFunctionSymbol("Shared", ScalarTypes.Bool.Type())
	table := NewTableSymbol("Shared", column)
	entityGroup := NewEntityGroupSymbol("Prod", table)
	graphModel := NewGraphModelSymbolFromNames("GraphModel", []string{"Edges"}, []string{"Nodes"}, []string{"daily"})
	db := NewDatabaseSymbolWithAlternateName("db", "alias", []Symbol{external, function, table, entityGroup, graphModel}, false)

	if db.GetMember("Shared") != external {
		t.Fatalf("raw member lookup should preserve member order")
	}
	if db.GetTable("Shared") != table || db.GetFunction("Shared") != function || db.GetExternalTable("Shared") != external {
		t.Fatalf("typed lookups should scan typed database views")
	}
	if db.GetAnyTable("Shared") != table {
		t.Fatalf("GetAnyTable should prefer ordinary tables")
	}
	if got := db.EntityGroups(); len(got) != 1 || got[0] != entityGroup || db.GetEntityGroup("Prod") != entityGroup {
		t.Fatalf("entity group view/lookup failed: %#v", got)
	}
	if got := db.GraphModels(); len(got) != 1 || got[0] != graphModel || db.GetGraphModel("GraphModel") != graphModel {
		t.Fatalf("graph model view/lookup failed: %#v", got)
	}

	if same := db.WithAlternateName("alias"); same != db {
		t.Fatal("WithAlternateName should return the same instance for unchanged alternate name")
	}
	renamed := db.WithAlternateName("renamed")
	if renamed == db || renamed.AlternateName() != "renamed" || renamed.GetTable("Shared") != table {
		t.Fatalf("WithAlternateName clone failed")
	}
	replaced := db.WithMembers([]Symbol{function})
	if replaced.GetFunction("Shared") != function || replaced.GetTable("Shared") != nil || db.GetTable("Shared") != table {
		t.Fatalf("WithMembers should replace members without mutating the original")
	}
	added := replaced.AddMembers(table)
	if added.GetTable("Shared") != table || replaced.GetTable("Shared") != nil {
		t.Fatalf("AddMembers should append members without mutating the original")
	}
	if !UnknownDatabaseSymbol.IsOpen() || len(UnknownDatabaseSymbol.Members()) != 0 {
		t.Fatalf("unknown database = open:%t members:%d", UnknownDatabaseSymbol.IsOpen(), len(UnknownDatabaseSymbol.Members()))
	}
}

func TestLocalAndOptionSymbols(t *testing.T) {
	scalarVariable := NewVariableSymbol("threshold", ScalarTypes.Long.Type(), true, int64(10), nil)
	if scalarVariable.SymbolKind() != SymbolKindVariable || scalarVariable.SymbolTabularity() != TabularityScalar {
		t.Fatalf("variable symbol metadata = kind:%s tabularity:%s", scalarVariable.SymbolKind(), scalarVariable.SymbolTabularity())
	}
	if scalarVariable.ConstantValue() != int64(10) {
		t.Fatalf("variable constant value = %#v", scalarVariable.ConstantValue())
	}

	rowsTable := NewTableSymbol("T")
	tabularVariable := NewVariableSymbol("rows", rowsTable, false, nil, nil)
	if tabularVariable.SymbolTabularity() != TabularityTabular {
		t.Fatalf("tabular variable tabularity = %s", tabularVariable.SymbolTabularity())
	}

	parameter := NewParameterSymbol("tenantId", ScalarTypes.String.Type(), "Tenant identifier")
	if parameter.SymbolKind() != SymbolKindParameter || parameter.Description() != "Tenant identifier" {
		t.Fatalf("parameter symbol = kind:%s desc:%q", parameter.SymbolKind(), parameter.Description())
	}

	option := NewOptionSymbolWithType("query_now", "Overrides now().", ScalarTypes.DateTime, []string{"datetime(2024-01-01)"})
	if option.SymbolKind() != SymbolKindOption || option.SymbolTabularity() != TabularityScalar {
		t.Fatalf("option symbol = kind:%s tabularity:%s", option.SymbolKind(), option.SymbolTabularity())
	}
	if got := option.Types(); len(got) != 1 || got[0] != ScalarTypes.DateTime {
		t.Fatalf("option types = %#v", got)
	}
	examples := option.Examples()
	examples[0] = "mutated"
	if option.Examples()[0] != "datetime(2024-01-01)" {
		t.Fatal("option examples accessor did not return a copy")
	}

	globals := NewGlobalState(nil).WithOptions([]*OptionSymbol{option})
	if globals.GetOption("query_now") != option || globals.GetOption("missing") != nil {
		t.Fatalf("global option lookup failed")
	}
	copy := globals.WithOptions(nil)
	if copy.GetOption("query_now") != nil || globals.GetOption("query_now") != option {
		t.Fatalf("WithOptions should copy state without mutating original")
	}
}
