package kustolang

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

// Source oracle built at 6b5c2f7 using the checked-in Microsoft T4 generators.
// The oracle is development-only; these regressions run entirely in Go.
func TestUpstreamOct07PolicyAndCompletionParity(t *testing.T) {
	type example struct {
		Display string `json:"display"`
		Edit    string `json:"edit"`
	}
	var fixture struct {
		Commit      string                   `json:"commit"`
		Policies    []attestedLanguageResult `json:"policies"`
		Completions []struct {
			Query    string    `json:"query"`
			Examples []example `json:"examples"`
		} `json:"completions"`
	}
	data, err := os.ReadFile("testdata/upstream_6b5c2f7_parity.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != "6b5c2f739a14bd5de52115f9bf9695985f0852dd" || len(fixture.Policies) != 20 || len(fixture.Completions) != 14 {
		t.Fatalf("invalid upstream fixture inventory: %s, %d policies, %d completions", fixture.Commit, len(fixture.Policies), len(fixture.Completions))
	}
	// Known catalog tables isolate command behavior from database discovery.
	database := NewDatabaseSymbol("db", NewTableSymbol("T"), NewTableSymbol("U"), NewTableSymbol("policy"))
	globals := DefaultGlobalState.WithCluster(NewClusterSymbol("test", database)).WithDatabase(database)
	for _, want := range fixture.Policies {
		t.Run(want.Query, func(t *testing.T) {
			if want.Columns == nil {
				want.Columns = []attestedColumn{}
			}
			got := currentAttestedLanguageResult(want.Query, globals)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("native: %#v\nupstream: %#v", got, want)
			}
		})
	}
	for _, want := range fixture.Completions {
		t.Run(want.Query, func(t *testing.T) {
			got := []example{}
			items := NewKustoCodeService(want.Query).GetCompletionItems(byteOffsetToUTF16Position(want.Query, len(want.Query)), CompletionOptionsDefault, CancellationToken{}).Items
			for _, item := range items {
				if item.Kind == CompletionKindExample {
					got = append(got, example{item.DisplayText, item.BeforeText()})
				}
			}
			sort.Slice(got, func(i, j int) bool { return got[i].Display < got[j].Display })
			if !reflect.DeepEqual(got, want.Examples) {
				t.Fatalf("native: %#v\nupstream: %#v", got, want.Examples)
			}
		})
	}
}
