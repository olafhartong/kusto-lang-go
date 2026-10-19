# Using the kustolang language layer in other projects

The `kustolang` and `kustolang/environment` packages form a self-contained,
reusable KQL language layer: a Go port of the Microsoft Kusto parser, binder,
and editor services. They have **no dependency on KustoKITT**, its REST
API, its embedded schema catalogs, or the GPL-licensed FalconForce NRT
catalogs — only the Go standard library at runtime.

Architectural boundaries are enforced by tests
(`internal/guard`) and by an out-of-tree consumer build
(`tools/langsmoke`, run in CI and via `make langsmoke`).

## Packages

| Package | Purpose |
| --- | --- |
| `github.com/olafhartong/kusto-lang-go/kustolang` | Parser, binder, semantic analysis, editor services (completion, hover, code actions, formatting, outlining), script/multi-block support |
| `github.com/olafhartong/kusto-lang-go/kustolang/environment` | JSON-compatible schema model (tables, functions, clusters) and the environments file loader |

## Installation

```sh
go get github.com/olafhartong/kusto-lang-go
```

## Parse and analyze a query

```go
package main

import (
	"fmt"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func main() {
	env := &environment.EnvironmentDefinition{
		TableDetails: map[string]environment.TableDetails{
			"MyLogs": {"Timestamp": "datetime", "Computer": "string", "EventID": "int"},
		},
	}

	code, err := kustolang.ParseAndAnalyzeForEnvironment(
		"MyLogs | where EventID == 4624 | project Timestamp, Computer",
		env, environment.LocalData{})
	if err != nil {
		panic(err)
	}
	for _, d := range code.GetDiagnostics() {
		fmt.Printf("%s [%d..%d): %s\n", d.Code, d.Start, d.End(), d.Message)
	}
}
```

Schema-free parsing is also available: `kustolang.Parse(text)` (syntax only)
and `kustolang.ParseAndAnalyze(text)` (semantics against an empty database).

## Load environments from JSON

The same `environments.json` format used by the analyzer works standalone:

```go
envs, err := environment.LoadEnvironments("environments.json")
if err != nil {
	panic(err)
}
environment.AddM365WithSentinelIfPresent(envs)
code, err := kustolang.ParseAndAnalyzeForEnvironment(query, envs["sentinel"], environment.LocalData{})
```

`environment.LocalData` carries per-request tables, watchlists, and functions
that augment the environment without mutating it.

The `sentinel` and `m365_with_sentinel` environments include `_GetWatchlist`
and ASIM parser metadata by default. For a custom or programmatically created
environment, set `FunctionCatalogs: []string{"sentinel"}`. This catalog is
embedded in the standalone language layer and uses Microsoft's MIT-licensed
parser signature and schema metadata. See
[Sentinel functions and refreshing ASIM metadata](sentinel-functions.md)
for configuration, watchlist schemas, and the update workflow.

## Reuse a GlobalState

Building symbols from an environment has a cost; when analyzing many queries
against the same schema, build the `GlobalState` once:

```go
globals, err := kustolang.NewGlobalStateForEnvironment(env, environment.LocalData{})
if err != nil {
	panic(err)
}
for _, query := range queries {
	code := kustolang.ParseAndAnalyze(query, globals)
	_ = code.GetDiagnostics()
}
```

`GlobalState` is immutable after construction and safe for concurrent use.

## Editor services (completion, hover, code actions)

`CodeService` powers IDE-style features. Positions are UTF-16 code-unit
offsets, matching the upstream C# library and LSP:

```go
service, err := kustolang.NewCodeServiceForEnvironment("MyLogs | where ", env, environment.LocalData{})
if err != nil {
	panic(err)
}

// Completion at the caret.
completion := service.GetCompletionItemsDefault(15, kustolang.CancellationToken{})
for _, item := range completion.Items {
	fmt.Println(item.DisplayText)
}

// Hover info.
info := service.GetQuickInfoDefault(3, kustolang.CancellationToken{})
fmt.Println(info.Text())

// Code actions (fix-its, refactorings) for a position or selection.
actions := service.GetCodeActionsDefault(3, 3, 0)
for _, action := range actions.Actions {
	fmt.Println(action.Base().Title)
}
```

Other services include `GetClassifications` (semantic highlighting),
`GetOutlines`, `GetFormattedText`, `GetRelatedElements`,
`GetClusterReferences`/`GetDatabaseReferences`, and `GetMinimalText`.

## Scripts (multi-statement documents)

`CodeScript` splits a document into blocks (queries, commands, directives)
and manages per-block services — the building block for a language server:

```go
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
// Incremental edits:
script = script.WithText(newText)
```

## Cancellation

Long-running analysis accepts cancellation, either from a `context.Context`
(`kustolang.NewCancellationToken(ctx)` and
`ParseAndAnalyzeForEnvironmentContext`) or a manual token. Cancelled
operations return `kustolang.OperationCanceledError`.

## Compatibility and licensing

- The public `kustolang` API mirrors the upstream
  [Kusto-Query-Language](https://github.com/microsoft/Kusto-Query-Language)
  C# library (Apache-2.0); parity is tracked in
  [upstream-parity.md](upstream-parity.md).
- The language layer contains no GPL-licensed material. The FalconForce NRT
  catalogs (GPL-3.0) are embedded only in the analyzer packages, which the
  language layer never imports.
- The exported surface is pinned by API snapshot tests
  (`internal/guard/testdata/*.golden`); breaking changes are deliberate and
  reviewed.

## Low-level engine integration

`github.com/olafhartong/kusto-lang-go/kustolang/engine` exposes the native
source-port API for consumers such as KustoKITT that manage binding and recovery
directly. Import it as `engine` when using it alongside the facade.
`engine.KustoCode` diagnostics and unsuffixed syntax spans use UTF-8 bytes;
`kustolang.KustoCode` diagnostics and editor-service positions use UTF-16.
Use the explicit UTF-16 syntax helpers at editor boundaries. Symbol types are
shared between the packages; parsed-code wrappers are intentionally distinct.
