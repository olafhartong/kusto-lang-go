# Editor services

For a first completion, hover, or code-action example, see the
[embedding guide](using-kustolang.md#editor-services-completion-hover-code-actions).
For coordinate conventions and package boundaries, see the
[language API reference](language-api.md#schemas-diagnostics-and-source-coordinates).

## Create a code service

The `kustolang` package exposes the native editor pipeline for diagnostics,
completion, quick info, classification, formatting, outlines, related
elements, and code actions. It uses the `kustolang/environment` schema model
(shared with the analyzer via type aliases) and does not load .NET or
`Microsoft.Azure.Kusto.Language` at runtime.

```go
package main

import (
	"fmt"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func main() {
	local := environment.LocalData{
		Tables: map[string]environment.TableDetails{
			"Events": {"Account": "string", "TimeGenerated": "datetime"},
		},
	}
	text := `Events | where Account ==`
	service, err := kustolang.NewCodeServiceForEnvironment(text, nil, local)
	if err != nil {
		panic(err)
	}

	for _, diagnostic := range service.GetDiagnostics(true, kustolang.CancellationToken{}) {
		fmt.Printf("%s: %s\n", diagnostic.Code, diagnostic.Message)
	}
}
```

## Global state, positions, and catalogs

`NewCodeService` uses the default Kusto global state. Use
`NewCodeServiceForEnvironment` when table, function, database, or cluster
metadata is available. Editor positions and ranges follow Kusto.Language's
UTF-16 indexing convention. Source-shaped static catalogs such as
`CodeKinds.Query`, `CodeServiceFeatures.Completion`,
`DiagnosticSeverity.Warning`, `ServerKinds.Engine`, and
`KustoActorNames.DiagnosticFixer` are available alongside the existing
prefixed constants.

## Multi-block documents and service factories

For multi-cell or blank-line-separated documents, `NewCodeScript` and
`NewCodeScriptForEnvironment` expose independently analyzed blocks with
UTF-16-adjusted diagnostics and editor services. `GetCombinedDiagnostics`
combines parser, semantic, and analyzer diagnostics with optional filtering.
Nested composite service factories retain sibling factories when globals are
replaced, and custom block separators remain independently replaceable.
