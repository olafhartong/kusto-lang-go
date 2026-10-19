package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func TestDefenderTimeAliasesNativeAnalysisAndCompletion(t *testing.T) {
	env := &environment.EnvironmentDefinition{TableDetails: map[string]environment.TableDetails{"DeviceEvents": {"Timestamp": "datetime"}}}
	query := "DeviceEvents | where TimeGenerated > ago(1d) and Timestamp > ago(1d) | project TimeGenerated, Timestamp"
	code, err := kustolang.ParseAndAnalyzeForEnvironment(query, env, environment.LocalData{})
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	table, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatal("missing result schema")
	}
	for _, name := range []string{"Timestamp", "TimeGenerated"} {
		column := table.GetColumn(name)
		if column == nil || column.Type.SymbolName() != "datetime" {
			t.Fatalf("missing datetime column %s", name)
		}
	}
	text := "DeviceEvents | project "
	service, err := kustolang.NewCodeServiceForEnvironment(text, env, environment.LocalData{})
	if err != nil {
		t.Fatal(err)
	}
	completion := service.GetCompletionItems(len(text), kustolang.CompletionOptionsDefault, kustolang.CancellationToken{})
	for _, name := range []string{"Timestamp", "TimeGenerated"} {
		found := false
		for _, item := range completion.Items {
			found = found || item.DisplayText == name
		}
		if !found {
			t.Fatalf("missing completion for %s", name)
		}
	}
}
