# Language API and compatibility details

For installation and a complete first example, start with
[Using the language layer](using-kustolang.md). This reference covers the
package boundaries, language model, coordinate conventions, and lower-level APIs.
See [Editor services](editor-services.md) for document and IDE integrations.
This repository provides a Go library; it does not include a standalone
language server.

## Packages and architecture

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

## Sentinel metadata

Sentinel environments include `_GetWatchlist` and an offline ASIM parser
catalog by default, with parameter signatures and normalized column types for
analysis and editor features. Custom environments can select it with
`"function_catalogs": ["sentinel"]`. See
[Sentinel functions and refreshing ASIM metadata](sentinel-functions.md)
for watchlist schemas, workspace overrides, and the catalog refresh command.

## Parse, analyze, and inspect syntax

The `kustolang` package exposes the native `KustoCode`, lexer, typed syntax
tree, symbol model, binder scope queries, generated visitors, and complete
`SyntaxKind` catalog. Together with `kustolang/environment` it forms a
self-contained language layer with no dependency on the analyzer packages —
see [Using the kustolang language layer](using-kustolang.md) for
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

## Schemas, diagnostics, and source coordinates

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

## Text facts and utility collections

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

## Lifecycle and cancellation

`KustoCode` lifecycle methods preserve receiver identity when analysis or
globals are unchanged, and scope, speculative-symbol, and row-scope lookups
propagate cancellation through their syntax traversals. Custom code services
can override combined diagnostics; default combination uses fresh storage and
fallback classification honors custom client-parameter callbacks.

## Global state and catalog lookup

`DefaultGlobalState` matches Kusto.Language's unknown database and cluster
environment. `GetFunction`, `GetAggregate`, and `GetPlugIn` keep the upstream
catalog boundaries and duplicate-selection rules. Go callers that intentionally
need the earlier database-plus-all-catalog lookup can use `GetAnyFunction`;
`NewGlobalState(nil)` retains the closed synthetic `db` compatibility state.
An empty server kind selects no command catalog. `GetAnyTableSymbol` preserves
external-table and materialized-view identity when callers need the concrete
symbol instead of the shared embedded table API.

## Symbols, catalogs, and type helpers

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
union/intersection helpers.

## Literal values and syntax ownership

Literal tokens expose typed values for every
upstream literal kind: exact decimals use standard-library `big.Rat`, full-range
timespans use signed 100 ns `TimeSpanValue` ticks with checked `time.Duration`
conversion, datetimes use `time.Time`, and GUIDs use `[16]byte`. Syntax-tree
depth and `WalkSyntaxNodes` count and visit nodes without treating tokens as
nodes. Public `TableState` flags preserve
the upstream serialized/sorted/open transition semantics. Public syntax-tree
construction enforces single-tree root ownership, while `CopyAsFragment`
preserves generated node types, diagnostics, and original-tree positions.

## Direct grammar parsing and parser combinators

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
