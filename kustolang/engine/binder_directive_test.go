package kustolang

import "testing"

func TestTryGetDirectiveClusterAndDatabasePreservesAbsentAndEmptyNames(t *testing.T) {
	parse := func(text string) ClientDirective {
		t.Helper()
		directive, ok := TryParseClientDirective(NewEditString(text))
		if !ok {
			t.Fatalf("TryParseClientDirective(%q) failed", text)
		}
		return directive
	}
	assertName := func(label string, got *string, want *string) {
		t.Helper()
		if want == nil {
			if got != nil {
				t.Fatalf("%s = %q, want nil", label, *got)
			}
			return
		}
		if got == nil || *got != *want {
			t.Fatalf("%s = %#v, want %q", label, got, *want)
		}
	}
	ptr := func(value string) *string { return &value }

	for _, test := range []struct {
		name     string
		text     string
		cluster  *string
		database *string
		ok       bool
	}{
		{name: "database only", text: "#database Samples", cluster: nil, database: ptr("Samples"), ok: true},
		{name: "cluster and database", text: "#database help.kusto.windows.net Samples", cluster: ptr("help.kusto.windows.net"), database: ptr("Samples"), ok: true},
		{name: "URI", text: "#database https://help.kusto.windows.net/Samples", cluster: ptr("help.kusto.windows.net"), database: ptr("Samples"), ok: true},
		{name: "explicit empty cluster", text: `#database "" Samples`, cluster: ptr(""), database: ptr("Samples"), ok: true},
		{name: "entity path", text: `#connect cluster("help.kusto.windows.net").database("Samples")`, cluster: ptr("help.kusto.windows.net"), database: ptr("Samples"), ok: true},
		{name: "connection", text: `#connect "Data Source=https://help.kusto.windows.net/Samples"`, cluster: ptr("help.kusto.windows.net"), database: ptr("Samples"), ok: true},
		{name: "named connection", text: "#connect @prod", ok: false},
		{name: "missing database", text: "#database", ok: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			cluster, database, ok := TryGetDirectiveClusterAndDatabase(parse(test.text))
			if ok != test.ok {
				t.Fatalf("ok = %t, want %t", ok, test.ok)
			}
			assertName("cluster", cluster, test.cluster)
			assertName("database", database, test.database)
		})
	}
}

func TestNativeDirectiveBindingSelectsCurrentDatabaseForFollowingSource(t *testing.T) {
	currentTable := NewTableSymbol("Current", NewColumnSymbol("X", ScalarTypes.String.Type()))
	sampleTable := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	current := NewDatabaseSymbol("CurrentDB", currentTable)
	samples := NewDatabaseSymbol("Samples", sampleTable)
	cluster := NewClusterSymbol("current.kusto.windows.net", current, samples)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{cluster}).
		WithCluster(cluster).
		WithDatabase(current)

	query := ParseQuery("#database Samples\nT | project A")
	binder := NewBinder(globals)
	if !binder.TryBind(query) {
		t.Fatal("TryBind returned false")
	}
	if binder.CurrentCluster() != cluster || binder.CurrentDatabase() != samples {
		t.Fatalf("current context = cluster:%#v database:%#v", binder.CurrentCluster(), binder.CurrentDatabase())
	}
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
		t.Fatalf("directive query diagnostics = %#v", diagnostics)
	}

	var source *NameReference
	WalkSyntax(query, func(element SyntaxElement) {
		if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(reference) == "T" {
			source = reference
		}
	})
	if source == nil || source.ReferencedSymbol() != sampleTable {
		t.Fatalf("source symbol = %#v, want Samples.T", source)
	}
	directive := query.Directives().At(0)
	if info, ok := directive.GetSemanticInfo(); !ok || len(info.Diagnostics) != 0 {
		t.Fatalf("directive semantic info = %#v, %t", info, ok)
	}
}

func TestNativeDirectiveDiagnosticsUseUpstreamSeverityAndLocations(t *testing.T) {
	current := NewDatabaseSymbol("CurrentDB", NewTableSymbol("Current"))
	cluster := NewClusterSymbol("current.kusto.windows.net", current)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{cluster}).
		WithCluster(cluster).
		WithDatabase(current)
	text := "#mystery value\n#connect cluster(\"missing\").database(\"db\")\n#database absent\nCurrent"
	query := ParseQuery(text)
	binder := NewBinder(globals)
	binder.TryBind(query)

	unknown := query.Directives().At(0).SemanticDiagnostics()
	if len(unknown) != 1 || unknown[0].Code != "KS250" || unknown[0].Start != 1 || unknown[0].Length != len("mystery") {
		t.Fatalf("unknown directive diagnostics = %#v", unknown)
	}

	connectToken := query.Directives().At(1).Token()
	connect := query.Directives().At(1).SemanticDiagnostics()
	connectStart := len("#mystery value\n")
	if len(connect) != 2 || connect[0].Code != "KS207" || connect[0].Severity != DiagnosticSeverityError ||
		connect[0].Start != connectStart || connect[0].Length != connectToken.Width() ||
		connect[1].Code != "KS208" || connect[1].Start != connectStart || connect[1].Length != connectToken.Width() {
		t.Fatalf("connect directive diagnostics = %#v", connect)
	}

	databaseToken := query.Directives().At(2).Token()
	database := query.Directives().At(2).SemanticDiagnostics()
	databaseStart := connectStart + connectToken.Width() + 1
	if len(database) != 1 || database[0].Code != "KS208" || database[0].Severity != DiagnosticSeverityError ||
		database[0].Start != databaseStart || database[0].Length != databaseToken.Width() {
		t.Fatalf("database directive diagnostics = %#v", database)
	}
	if binder.CurrentCluster() != UnknownClusterSymbol || binder.CurrentDatabase() != UnknownDatabaseSymbol {
		t.Fatalf("unknown context = cluster:%#v database:%#v", binder.CurrentCluster(), binder.CurrentDatabase())
	}
}
