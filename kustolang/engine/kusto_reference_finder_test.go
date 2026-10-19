package kustolang

import (
	"context"
	"strings"
	"testing"
)

func TestKustoReferenceFinderReportsDirectClusterAndDatabaseCalls(t *testing.T) {
	globals := referenceFinderGlobals(nil)
	text := `cluster("help").database("Samples").T`
	service := NewKustoCodeService(text, globals)

	clusters := service.GetClusterReferences(CancellationToken{})
	if len(clusters) != 1 || clusters[0].Cluster != "help" || clusters[0].Start() != strings.Index(text, `"help"`) || clusters[0].Length() != len(`"help"`) {
		t.Fatalf("cluster references = %#v", clusters)
	}
	databases := service.GetDatabaseReferences(CancellationToken{})
	if len(databases) != 1 || databases[0].Database != "Samples" || databases[0].Cluster != "help.kusto.windows.net" ||
		databases[0].Start() != strings.Index(text, `"Samples"`) || databases[0].Length() != len(`"Samples"`) {
		t.Fatalf("database references = %#v", databases)
	}
}

func TestKustoReferenceFinderUsesCurrentClusterAndConstantArgumentLocations(t *testing.T) {
	globals := referenceFinderGlobals(nil)
	text := `declare query_parameters(target:string = "help"); cluster(target)`
	service := NewKustoCodeService(text, globals)

	clusters := service.GetClusterReferences(CancellationToken{})
	if len(clusters) != 1 || clusters[0].Cluster != "help" || clusters[0].Start() != strings.LastIndex(text, "target") || clusters[0].Length() != len("target") {
		t.Fatalf("constant cluster references = %#v", clusters)
	}
	databaseText := `database("db").Current`
	databases := NewKustoCodeService(databaseText, globals).GetDatabaseReferences(CancellationToken{})
	if len(databases) != 1 || databases[0].Database != "db" || databases[0].Cluster != "current.kusto.windows.net" {
		t.Fatalf("current-cluster database references = %#v", databases)
	}
}

func TestKustoReferenceFinderAttributesCalledFunctionReferencesToCallSite(t *testing.T) {
	remote := NewComputedFunctionSymbol("Remote", `cluster("help").database("Samples").T`, TabularityTabular)
	service := NewKustoCodeService("Remote()", referenceFinderGlobals(remote))

	clusters := service.GetClusterReferences(CancellationToken{})
	if len(clusters) != 1 || clusters[0].Cluster != "help" || clusters[0].Start() != 0 || clusters[0].Length() != len("Remote") {
		t.Fatalf("expanded cluster references = %#v", clusters)
	}
	databases := service.GetDatabaseReferences(CancellationToken{})
	if len(databases) != 1 || databases[0].Database != "Samples" || databases[0].Cluster != "help.kusto.windows.net" || databases[0].Start() != 0 || databases[0].Length() != len("Remote") {
		t.Fatalf("expanded database references = %#v", databases)
	}
}

func TestKustoReferenceFinderPropagatesConstantParametersIntoCalledBodies(t *testing.T) {
	parameter := NewParameter("databaseName", ScalarTypes.String.Type())
	remote := NewComputedFunctionSymbol("Remote", `database(databaseName).Current`, TabularityTabular, parameter)
	service := NewKustoCodeService(`Remote("db")`, referenceFinderGlobals(remote))

	databases := service.GetDatabaseReferences(CancellationToken{})
	if len(databases) != 1 || databases[0].Database != "db" || databases[0].Cluster != "current.kusto.windows.net" || databases[0].Start() != 0 || databases[0].Length() != len("Remote") {
		t.Fatalf("parameterized expanded database references = %#v", databases)
	}
}

func TestKustoReferenceFinderProjectsMacroAlternateRangesToOriginalBody(t *testing.T) {
	text := `macro-expand Targets as Scope (database("Samples").T)`
	refs := NewKustoCodeService(text, macroExpandTestGlobals()).GetDatabaseReferences(CancellationToken{})
	wantStart := strings.Index(text, `"Samples"`)
	if len(refs) != 2 {
		t.Fatalf("macro-expand database references = %#v, want two", refs)
	}
	for _, ref := range refs {
		if ref.Database != "Samples" || ref.Start() != wantStart || ref.Length() != len(`"Samples"`) {
			t.Fatalf("macro-expand database reference = %#v, want Samples at %d+%d", ref, wantStart, len(`"Samples"`))
		}
	}
}

func TestKustoReferenceFinderUsesUTF16PublicRanges(t *testing.T) {
	text := `cluster("😀")`
	refs := NewKustoCodeService(text, referenceFinderGlobals(nil)).GetClusterReferences(CancellationToken{})
	if len(refs) != 1 || refs[0].Cluster != "😀" || refs[0].Start() != 8 || refs[0].Length() != 4 {
		t.Fatalf("UTF-16 cluster references = %#v", refs)
	}
}

func TestKustoReferenceFinderReportsClientDirectiveReferences(t *testing.T) {
	globals := referenceFinderGlobals(nil)
	tests := []struct {
		name            string
		text            string
		wantCluster     string
		wantDatabase    string
		wantDBCluster   string
		wantLine        string
		wantClusterRefs int
		wantDatabaseRef int
	}{
		{
			name:            "database with explicit cluster",
			text:            "#database \"help.kusto.windows.net\" \"Samples\"\nT",
			wantCluster:     "help.kusto.windows.net",
			wantDatabase:    "Samples",
			wantDBCluster:   "help.kusto.windows.net",
			wantLine:        `#database "help.kusto.windows.net" "Samples"`,
			wantClusterRefs: 1,
			wantDatabaseRef: 1,
		},
		{
			name:            "preceding cluster supplies database context",
			text:            "#cluster \"help\"\n#database \"Samples\"\nT",
			wantDatabase:    "Samples",
			wantDBCluster:   "help",
			wantLine:        `#database "Samples"`,
			wantDatabaseRef: 1,
		},
		{
			name:            "database URI",
			text:            "#database \"https://help.kusto.windows.net/Samples\"\nT",
			wantCluster:     "help.kusto.windows.net",
			wantDatabase:    "Samples",
			wantDBCluster:   "help.kusto.windows.net",
			wantLine:        `#database "https://help.kusto.windows.net/Samples"`,
			wantClusterRefs: 1,
			wantDatabaseRef: 1,
		},
		{
			name:            "connect entity path",
			text:            "#connect cluster(\"help\").database(\"Samples\")\nT",
			wantCluster:     "help",
			wantDatabase:    "Samples",
			wantDBCluster:   "help",
			wantLine:        `#connect cluster("help").database("Samples")`,
			wantClusterRefs: 1,
			wantDatabaseRef: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewKustoCodeService(test.text, globals)
			clusters := service.GetClusterReferences(CancellationToken{})
			if len(clusters) != test.wantClusterRefs {
				t.Fatalf("cluster references = %#v", clusters)
			}
			if len(clusters) == 1 && clusters[0].Cluster != test.wantCluster {
				t.Fatalf("cluster reference = %#v", clusters[0])
			}
			databases := service.GetDatabaseReferences(CancellationToken{})
			if len(databases) != test.wantDatabaseRef {
				t.Fatalf("database references = %#v", databases)
			}
			if len(databases) == 1 {
				start := byteOffsetToUTF16Position(test.text, strings.Index(test.text, test.wantLine))
				length := byteOffsetToUTF16Position(test.wantLine, len(test.wantLine))
				if databases[0].Database != test.wantDatabase || databases[0].Cluster != test.wantDBCluster || databases[0].Start() != start || databases[0].Length() != length {
					t.Fatalf("database reference = %#v, want %s/%s at %d+%d", databases[0], test.wantDBCluster, test.wantDatabase, start, length)
				}
				if len(clusters) == 1 && (clusters[0].Start() != start || clusters[0].Length() != length) {
					t.Fatalf("cluster directive range = %#v, want %d+%d", clusters[0], start, length)
				}
			}
		})
	}
}

func TestKustoReferenceFinderDirectiveControlsAndUTF16Ranges(t *testing.T) {
	globals := referenceFinderGlobals(nil)
	for _, text := range []string{"#database\nT", "#connect @named\nT", "#unknown value\nT"} {
		service := NewKustoCodeService(text, globals)
		if clusters := service.GetClusterReferences(CancellationToken{}); len(clusters) != 0 {
			t.Fatalf("cluster references for %q = %#v", text, clusters)
		}
		if databases := service.GetDatabaseReferences(CancellationToken{}); len(databases) != 0 {
			t.Fatalf("database references for %q = %#v", text, databases)
		}
	}

	text := "#database \"help.kusto.windows.net\" \"😀db\"\nT"
	refs := NewKustoCodeService(text, globals).GetDatabaseReferences(CancellationToken{})
	wantLength := byteOffsetToUTF16Position(strings.Split(text, "\n")[0], len(strings.Split(text, "\n")[0]))
	if len(refs) != 1 || refs[0].Database != "😀db" || refs[0].Length() != wantLength {
		t.Fatalf("UTF-16 directive references = %#v, want length %d", refs, wantLength)
	}
}

func TestKustoReferenceFinderSkipsNonconstantAndSurvivesMalformedOrCanceledCode(t *testing.T) {
	globals := referenceFinderGlobals(nil)
	nonconstant := NewKustoCodeService(`T | project C=cluster(A)`, globals)
	if refs := nonconstant.GetClusterReferences(CancellationToken{}); len(refs) != 0 {
		t.Fatalf("nonconstant cluster references = %#v", refs)
	}

	malformed := NewKustoCodeService(`cluster("help"`, globals)
	refs := malformed.GetClusterReferences(CancellationToken{})
	if len(refs) != 1 || refs[0].Cluster != "help" {
		t.Fatalf("malformed cluster references = %#v", refs)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if refs := NewKustoCodeService(`cluster("help")`, globals).GetClusterReferences(NewCancellationToken(canceled)); len(refs) != 0 {
		t.Fatalf("canceled cluster references = %#v", refs)
	}
}

func referenceFinderGlobals(function *FunctionSymbol) *GlobalState {
	currentTable := NewTableSymbol("Current", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	currentMembers := []Symbol{currentTable}
	if function != nil {
		currentMembers = append(currentMembers, function)
	}
	current := NewDatabaseSymbol("db", currentMembers...)
	samples := NewDatabaseSymbol("Samples", NewTableSymbol("T", NewColumnSymbol("B", ScalarTypes.String.Type())))
	currentCluster := NewClusterSymbol("current.kusto.windows.net", current)
	helpCluster := NewClusterSymbol("help.kusto.windows.net", samples)
	return NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{currentCluster, helpCluster}).
		WithCluster(currentCluster).
		WithDatabase(current)
}
