package main

// docexamples compiles the snippets from docs/using-kustolang.md to keep the
// documentation honest.

import (
	"fmt"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func docExamples(env *environment.EnvironmentDefinition, text string, queries []string) {
	// Parse and analyze.
	code, err := kustolang.ParseAndAnalyzeForEnvironment(
		"MyLogs | where EventID == 4624 | project Timestamp, Computer",
		env, environment.LocalData{})
	if err != nil {
		panic(err)
	}
	for _, d := range code.GetDiagnostics() {
		fmt.Printf("%s [%d..%d): %s\n", d.Code, d.Start, d.End(), d.Message)
	}

	// Reuse a GlobalState.
	globals, err := kustolang.NewGlobalStateForEnvironment(env, environment.LocalData{})
	if err != nil {
		panic(err)
	}
	for _, query := range queries {
		code := kustolang.ParseAndAnalyze(query, globals)
		_ = code.GetDiagnostics()
	}

	// Editor services.
	service, err := kustolang.NewCodeServiceForEnvironment("MyLogs | where ", env, environment.LocalData{})
	if err != nil {
		panic(err)
	}
	completion := service.GetCompletionItemsDefault(15, kustolang.CancellationToken{})
	for _, item := range completion.Items {
		fmt.Println(item.DisplayText)
	}
	info := service.GetQuickInfoDefault(3, kustolang.CancellationToken{})
	fmt.Println(info.Text())
	actions := service.GetCodeActionsDefault(3, 3, 0)
	for _, action := range actions.Actions {
		fmt.Println(action.Base().Title)
	}

	// Scripts.
	script, err := kustolang.NewCodeScriptForEnvironment(text, env, environment.LocalData{})
	if err != nil {
		panic(err)
	}
	for _, block := range script.Blocks() {
		service := block.Service()
		if service == nil {
			continue
		}
		_ = service.GetDiagnosticsDefault()
	}
	script = script.WithText(text)
	_ = script
}
