# kusto-lang-go

An embeddable native Go implementation of Microsoft Kusto Query Language
(KQL): lexer, parser, binder, symbols, diagnostics, and editor services.
No .NET runtime or running service is required. This repository provides a Go
library; it does not provide a standalone language server.

## Packages

| Import | Purpose |
| --- | --- |
| `github.com/olafhartong/kusto-lang-go/kustolang` | Public language and editor facade; document diagnostics and editor positions use UTF-16 |
| `github.com/olafhartong/kusto-lang-go/kustolang/environment` | Environment/schema models, catalogs, loaders, and prepared schema snapshots |
| `github.com/olafhartong/kusto-lang-go/kustolang/engine` | Low-level source port, mutable binder and native parsed code; parser and syntax positions use UTF-8 bytes, editor-service positions use UTF-16 |

The engine package declares `package kustolang`; use an `engine` import alias
when importing it together with the facade. The facade and environment package
share the engine's symbol identities. KustoKITT consumes the engine directly
to preserve its existing byte-based analysis and recovery behavior.

The runtime language packages use only the Go standard library. The ASIM
metadata refresh utility uses a Go YAML dependency at development time.
Language packages do not import KustoKITT or embed its operational schemas,
HTTP server, rule scanner, or FalconForce NRT catalogs.

## Install

```sh
go get github.com/olafhartong/kusto-lang-go@v0.1.0
```

`v0.1.0` is the intended first release after this repository is published.
Until then, use a local workspace as described in [Repository split](docs/repository-split.md).

## Build and validate

Use Go 1.27.2 or newer (the module retains Go 1.25 language compatibility).

```sh
make check
make langsmoke
make race
make generate
make provenance
```

[Using the language layer](docs/using-kustolang.md) covers schema loading and
embedding. [Sentinel metadata](docs/sentinel-functions.md) describes the bundled
MIT-licensed function metadata. The [Microsoft upstream changelog](UPSTREAM_CHANGELOG.md)
records affected areas, functions, and observed behavior across source updates.
[Upstream parity](docs/upstream-parity.md)
records source-port coverage and remaining differences.

## Use the language model

Sentinel environments include `_GetWatchlist` and an offline ASIM parser
catalog by default, with parameter signatures and normalized column types for
analysis and editor features. Custom environments can select it with
`"function_catalogs": ["sentinel"]`. See
[Sentinel functions and refreshing ASIM metadata](docs/sentinel-functions.md)
for watchlist schemas, workspace overrides, and the catalog refresh command.

The `kustolang` package exposes the native `KustoCode`, lexer, typed syntax
tree, symbol model, binder scope queries, generated visitors, and complete
`SyntaxKind` catalog. Together with `kustolang/environment` it forms a
self-contained language layer with no dependency on the analyzer packages —
see [Using the kustolang language layer](docs/using-kustolang.md) for
consuming it from other projects.

```go
column := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
table := kustolang.NewTableSymbol("Events", column)
globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))

code := kustolang.ParseAndAnalyze(`Events | project Account`, globals)
for _, diagnostic := range code.GetDiagnostics() {
	fmt.Println(diagnostic.Code, diagnostic.Message)
}

code.Syntax().WalkElements(func(element kustolang.SyntaxElement) {
	reference, ok := kustolang.WrapGeneratedSyntaxElement(element).(*kustolang.NameReference)
	if ok {
		fmt.Println(reference.Text(kustolang.IncludeTriviaMinimal), reference.ReferencedSymbol())
	}
})
```

`ParseAndAnalyzeForEnvironment` builds the same semantic `GlobalState` from
the `kustolang/environment` schema model. Malformed input produces a recoverable typed
tree with missing elements and attached diagnostics. `KustoCode` positions and
diagnostic ranges use UTF-16 code units for compatibility with Kusto.Language;
`KustoCode.TryGetLineAndOffset` also preserves positions between the two UTF-16
code units of an astral character, as `System.String` does;
direct low-level syntax element offsets retain native Go byte indexes through
the unsuffixed and explicit `Byte*` methods. Direct AST consumers can use the
parallel `UTF16*` metrics, bounds, token lookup, and original-tree projection
methods when Kusto.Language-compatible coordinates are required. Node lookup
and ranged token walking likewise have explicit byte and UTF-16 variants. The
byte-sensitive binder, query-parser, and built-in editor/action paths use the
explicit `Byte*` APIs, leaving unsuffixed methods as compatibility aliases.

The public parser utilities also expose source-shaped `TextFacts`,
`TriviaFacts`, `CommentFacts`, and `ScriptFacts` facades. Their positions and
ranges use UTF-16 code units and retain recovery behavior for line facts,
comment spans, quoted text, and unterminated multiline strings. The existing
free-function parser utilities remain byte-oriented for compatibility. Parser
grammars can be rendered with `Describe`, while the public `SafeList[T]` and
`EmptyReadOnlyList[T]` helpers provide the pinned immutable snapshot and
empty-list semantics without exposing engine state.
Lexical tokens expose explicit `ByteLength` and `UTF16Length` metrics. Native
dictionary, list, validation, hash-set, atomic, and callback-cancellation
helpers cover the corresponding public Kusto.Language utility surface without
adding dependencies. Strict facades also expose the parser's pinned utility
models, including UTF-16 ranged `TextKeyedDictionary` and `StringTable`,
`ConnectionInfo`, `UniqueNameTable`, concurrent dictionaries, object pools,
and most-recently-used caches using only the Go standard library.

`KustoCode` lifecycle methods preserve receiver identity when analysis or
globals are unchanged, and scope, speculative-symbol, and row-scope lookups
propagate cancellation through their syntax traversals. Custom code services
can override combined diagnostics; default combination uses fresh storage and
fallback classification honors custom client-parameter callbacks.

`DefaultGlobalState` matches Kusto.Language's unknown database and cluster
environment. `GetFunction`, `GetAggregate`, and `GetPlugIn` keep the upstream
catalog boundaries and duplicate-selection rules. Go callers that intentionally
need the earlier database-plus-all-catalog lookup can use `GetAnyFunction`;
`NewGlobalState(nil)` retains the closed synthetic `db` compatibility state.
An empty server kind selects no command catalog. `GetAnyTableSymbol` preserves
external-table and materialized-view identity when callers need the concrete
symbol instead of the shared embedded table API.

The public language package also includes the upstream `Options` and
`Properties` catalogs, typed `GlobalState` property helpers, the thread-safe
`KustoCache`, and immutable `FunctionBodyFacts` returned by bound function
calls. Symbol tooling includes upstream-compatible `SchemaDisplay` and
`CompletionDisplay` text, primitive scalar flags, server and materialized-view
kinds, and the public argument, parameter, return, conversion, layout, and
operator enum catalogs. The immutable `QueryOperatorParameter` metadata and
the complete upstream `QueryOperatorParameters` catalog are also available;
catalog list methods return defensive copies. `Operators` exposes the same 55
built-in operator symbols and signature ordering used by semantic binding;
`Aggregates` and `PlugIns` expose the complete pinned public aggregation and
evaluate plug-in catalogs, and `Functions` exposes all 429 built-in function
symbols from the pinned revision. All catalog list methods return defensive
copies and reference the same symbols used by the binder. The public command
catalogs likewise expose all 637 engine commands plus the pinned server-specific
catalogs, including result schemas and grammar constructions. Concrete
`EngineCommandGrammar`, `DataManagerCommandGrammar`,
`ClusterManagerCommandGrammar`, and `AriaBridgeCommandGrammar` types preserve
server-specific construction and dynamic type identity through
`CreateServerCommandGrammar` and `ServerCommandGrammarFrom`. Public
`FunctionHelpers` and tuple constructors support source-compatible custom return
type builders without importing internal packages. `TypeFacts` exposes common
and widest type inference, conversion checks, scalar categories, dynamic-shape
tests, fallback and ignore-dynamic overload behavior, and column
union/intersection helpers. Literal tokens expose typed values for every
upstream literal kind: exact decimals use standard-library `big.Rat`, full-range
timespans use signed 100 ns `TimeSpanValue` ticks with checked `time.Duration`
conversion, datetimes use `time.Time`, and GUIDs use `[16]byte`. Syntax-tree
depth and `WalkSyntaxNodes` count and visit nodes without treating tokens as
nodes. Public `TableState` flags preserve
the upstream serialized/sorted/open transition semantics. Public syntax-tree
construction enforces single-tree root ownership, while `CopyAsFragment`
preserves generated node types, diagnostics, and original-tree positions.

The same package also exposes `QueryGrammar`, `CommandGrammar`, native parser
combinators, and syntax-token/list construction. Parser matching supports
custom equality functions for item, set, and sequence matching, including
non-comparable Go values. Result parsers can be cast with tag preservation,
and conversion helpers support both native UTF-8 bytes and UTF-16 code units.
Grammar productions can be
used directly without invoking the higher-level analyzer:

```go
grammar := kustolang.QueryGrammarFrom(kustolang.DefaultGlobalState)
query := kustolang.ParseFirstSyntax(
	grammar.QueryBlock,
	`StormEvents | where State ==`,
)
for _, diagnostic := range query.GetContainedSyntaxDiagnostics() {
	fmt.Println(diagnostic.Code, diagnostic.Message)
}
```

Direct grammar parsing uses the same recovery productions as `KustoCode`, so
incorrect syntax returns a typed tree containing missing elements and syntax
diagnostics instead of being accepted as valid input.

## Use the editor services

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

`NewCodeService` uses the default Kusto global state. Use
`NewCodeServiceForEnvironment` when table, function, database, or cluster
metadata is available. Editor positions and ranges follow Kusto.Language's
UTF-16 indexing convention. Source-shaped static catalogs such as
`CodeKinds.Query`, `CodeServiceFeatures.Completion`,
`DiagnosticSeverity.Warning`, `ServerKinds.Engine`, and
`KustoActorNames.DiagnosticFixer` are available alongside the existing
prefixed constants.

For multi-cell or blank-line-separated documents, `NewCodeScript` and
`NewCodeScriptForEnvironment` expose independently analyzed blocks with
UTF-16-adjusted diagnostics and editor services. `GetCombinedDiagnostics`
combines parser, semantic, and analyzer diagnostics with optional filtering.
Nested composite service factories retain sibling factories when globals are
replaced, and custom block separators remain independently replaceable.

## Repository layout

- `kustolang/` — public facade and editor API
- `kustolang/engine/` — native source port and its tests
- `kustolang/environment/` — schema model, environment support, and metadata
- `tools/` — generators and independent consumer smoke module
- `cmd/kusto-upstream-check/` — development-time upstream/provenance checker
- `parity-artifacts/` — commit-bound differential evidence
- `third_party/` — upstream licenses and source snapshots

## Licensing and provenance

Copyright 2026 Olaf Hartong. The existing GPL-3.0 project license is retained
in [LICENSE](LICENSE); this extraction does not relicense the project.
Ported Microsoft code retains its Apache-2.0 notices and preserved license.
Sentinel metadata retains its MIT attribution. See
[Third-party notices](THIRD_PARTY_NOTICES.md). FalconForce NRT catalogs remain
in KustoKITT and are not distributed by this repository.

## Upstream tracking

The pinned Microsoft revision is checked weekly in CI. The scheduled check also
classifies changes to grammar/parser syntax, functions/operators/plug-ins,
commands, diagnostics, and editor sources, and compares vendored generator
source snapshots with the exact files at the pinned GitHub commit. Run the same
standard-library-only revision check locally with:

```bash
go run ./cmd/kusto-upstream-check
```

The command resolves upstream `HEAD` and uses GitHub's commit-comparison API
against the pinned revision. It exits zero only when the commits are identical;
ahead, behind, diverged, missing-pin, malformed-response, and API-error states
are reported distinctly and fail closed.

Verify the recorded revisions, license hash, notice, manifest paths, and
commit-bound differential artifacts without network access with:

```bash
go run ./cmd/kusto-upstream-check -verify-provenance -local-only
```

Changing the pinned revision requires a fresh differential attestation for
that exact commit. The checker verifies each checked-in corpus, generator,
C#-oracle result, and native result hash and rejects stale or missing evidence.
