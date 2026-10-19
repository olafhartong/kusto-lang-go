package kustolang_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func TestSentinelDefaultFunctionCatalog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "environments.json")
	if err := os.WriteFile(path, []byte(`{"sentinel":{},"m365":{},"other":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	envs, err := environment.LoadEnvironments(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sentinel", "m365_with_sentinel"} {
		globals, err := kustolang.NewGlobalStateForEnvironment(envs[name], environment.LocalData{})
		if err != nil {
			t.Fatal(err)
		}
		for _, function := range []string{"_GetWatchlist", "_Im_Dns", "_ASim_Dns", "_Im_Authentication", "_Im_Dns_AzureFirewall", "imDns"} {
			if globals.Database.GetFunction(function) == nil {
				t.Errorf("%s missing %s", name, function)
			}
		}
	}
	for _, name := range []string{"m365", "other"} {
		if len(envs[name].TabularFunctions) != 0 || len(envs[name].MagicFunctions) != 0 {
			t.Fatalf("Sentinel functions leaked into %s", name)
		}
	}
}

func TestSentinelNativeSchemaSignaturesAndEditor(t *testing.T) {
	env := &environment.EnvironmentDefinition{FunctionCatalogs: []string{"sentinel"}}
	queries := []string{
		`_Im_Dns() | project DnsQuery, SrcIpAddr, TimeGenerated, DnsQueryType`,
		`_Im_Dns(starttime=ago(1h), domain_has_any=dynamic(["example.org"]), pack=true) | project DnsQuery`,
		`_ASim_Dns() | project DnsQuery`,
		`_Im_Dns_AzureFirewall() | project DnsQuery`,
		`_Im_Authentication() | project TargetUsername, EventResult`,
		`_GetWatchlist("unknown") | project SearchKey, CustomColumn`,
	}
	for _, query := range queries {
		code, err := kustolang.ParseAndAnalyzeForEnvironment(query, env, environment.LocalData{})
		if err != nil {
			t.Fatal(err)
		}
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Errorf("%s: %#v", query, diagnostics)
		}
	}
	for _, query := range []string{`_GetWatchlist()`, `_GetWatchlist("a", "b")`, `_Im_Dns(no_such_parameter=true)`} {
		code, err := kustolang.ParseAndAnalyzeForEnvironment(query, env, environment.LocalData{})
		if err != nil {
			t.Fatal(err)
		}
		if len(code.GetDiagnostics()) == 0 {
			t.Errorf("missing argument diagnostics for %s", query)
		}
	}
	text := `_Im_Dns() | project Dns`
	service, err := kustolang.NewCodeServiceForEnvironment(text, env, environment.LocalData{})
	if err != nil {
		t.Fatal(err)
	}
	completion := service.GetCompletionItems(len(text), kustolang.CompletionOptionsDefault, kustolang.CancellationToken{})
	found := false
	for _, item := range completion.Items {
		found = found || item.DisplayText == "DnsQuery"
	}
	if !found {
		t.Fatal("DnsQuery missing from ASIM column completions")
	}
	quick := service.GetQuickInfo(2, kustolang.QuickInfoOptionsDefault, kustolang.CancellationToken{})
	if len(quick.Items) == 0 || !strings.Contains(quick.Items[0].Text(), "_Im_Dns") {
		t.Fatalf("ASIM quick info = %#v", quick)
	}
}

func TestSentinelWatchlistNativeSchema(t *testing.T) {
	env := &environment.EnvironmentDefinition{FunctionCatalogs: []string{"sentinel"}}
	local := environment.LocalData{Watchlists: map[string]environment.WatchlistDetails{"VIP": {"Account": "string", "Score": "long"}}}
	code, err := kustolang.ParseAndAnalyzeForEnvironment(`_GetWatchlist("VIP") | project SearchKey, Account, Score`, env, local)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	table, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok || table.GetColumn("Score") == nil || table.GetColumn("Score").Type.SymbolName() != "long" {
		t.Fatalf("watchlist return type = %#v", table)
	}
	invalid, err := kustolang.ParseAndAnalyzeForEnvironment(`_GetWatchlist("VIP") | project Missing`, env, local)
	if err != nil || len(invalid.GetDiagnostics()) == 0 {
		t.Fatalf("declared watchlist should reject unknown column: %v", err)
	}
}

func TestSentinelCatalogOptOutAndOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "environments.json")
	if err := os.WriteFile(path, []byte(`{"sentinel":{"function_catalogs":[]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	envs, err := environment.LoadEnvironments(path)
	if err != nil || len(envs["sentinel"].TabularFunctions) != 0 {
		t.Fatalf("opt out failed: %v", err)
	}
	serialized, err := json.Marshal(envs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, serialized, 0o600); err != nil {
		t.Fatal(err)
	}
	envs, err = environment.LoadEnvironments(path)
	if err != nil || len(envs["sentinel"].FunctionCatalogs) != 0 {
		t.Fatalf("opt out did not survive JSON round trip: %v", err)
	}
	env := &environment.EnvironmentDefinition{
		FunctionCatalogs: []string{"sentinel"},
		TabularFunctions: map[string]environment.TabularFunctionDetails{"_Im_Dns": {OutputColumns: environment.TableDetails{"Custom": "long"}}},
	}
	code, err := kustolang.ParseAndAnalyzeForEnvironment(`_Im_Dns() | project Custom`, env, environment.LocalData{})
	if err != nil || len(code.GetDiagnostics()) != 0 {
		t.Fatalf("workspace override failed: %v", err)
	}
	if _, err := kustolang.NewGlobalStateForEnvironment(&environment.EnvironmentDefinition{FunctionCatalogs: []string{"typo"}}, environment.LocalData{}); err == nil {
		t.Fatal("unknown catalog should be rejected")
	}
	local := environment.LocalData{TabularFunctions: map[string]environment.TabularFunctionDetails{"_Im_Dns": {OutputColumns: environment.TableDetails{"Local": "string"}}}}
	code, err = kustolang.ParseAndAnalyzeForEnvironment(`_Im_Dns() | project Local`, &environment.EnvironmentDefinition{FunctionCatalogs: []string{"sentinel"}}, local)
	if err != nil || len(code.GetDiagnostics()) != 0 {
		t.Fatalf("local override failed: %v, %#v", err, code.GetDiagnostics())
	}
}

func TestSentinelCurrentDatabaseOverride(t *testing.T) {
	env := &environment.EnvironmentDefinition{
		FunctionCatalogs: []string{"sentinel"},
		CurrentCluster:   "workspace",
		CurrentDatabase:  "db",
		Clusters: map[string]environment.ClusterDetails{"workspace": {Databases: map[string]environment.DatabaseDetails{
			"db": {TabularFunctions: map[string]environment.TabularFunctionDetails{"_Im_Dns": {OutputColumns: environment.TableDetails{"DatabaseColumn": "long"}}}},
		}}},
	}
	code, err := kustolang.ParseAndAnalyzeForEnvironment(`_Im_Dns() | project DatabaseColumn`, env, environment.LocalData{})
	if err != nil || len(code.GetDiagnostics()) != 0 {
		t.Fatalf("current database override failed: %v", err)
	}
	table, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok || table.GetColumn("DatabaseColumn").Type.SymbolName() != "long" {
		t.Fatalf("current database return schema = %#v", table)
	}
}
