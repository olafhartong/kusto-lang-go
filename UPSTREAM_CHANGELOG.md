# Microsoft Kusto.Language upstream changelog

Observed changes in Microsoft's [Kusto-Query-Language](https://github.com/microsoft/Kusto-Query-Language)
source that affect this Go port. Dates are upstream commit dates, not NuGet
publication dates or Go release dates. A source synchronization can change
behavior without changing Microsoft's version number. This is a selective
history supported by upstream diffs and existing port regressions.

For each future upstream update, record the revision, area, affected functions
or symbols, and confirmed behavior. If behavior has not been measured, describe
the source update and label it accordingly. Keep the active pin in
[`third_party/kusto/UPSTREAM.json`](third_party/kusto/UPSTREAM.json); document
coverage and intentional differences in [`docs/upstream-parity.md`](docs/upstream-parity.md).
These requirements also apply to future coding agents through
[`AGENTS.md`](AGENTS.md). Local and CI provenance checks require a populated
table in the entry linking to the currently pinned revision.

## 2026-11-08 — graph-model policy commands

Source: [`dbd7ca7`](https://github.com/microsoft/Kusto-Query-Language/commit/dbd7ca7684adfd7a38536f74728670dc54a67acc).
Microsoft's source version remains **13.0.0**. This is a source synchronization,
not a new NuGet or Go package release.

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Graph-model policies | `ShowGraphModelPolicySnapshotIndexing`, `ShowGraphModelStarPolicySnapshotIndexing`, `AlterGraphModelPolicySnapshotIndexing`, `DeleteGraphModelPolicySnapshotIndexing`, and the four corresponding `GraphQuery` commands; Go `EngineCommands` and public `EngineCommandCatalog` | Adds eight commands for `snapshot_indexing` and `graph_query`. For example, `.show graph_model * policy snapshot_indexing` and `.alter graph_model M policy graph_query '{}'` now parse with the five policy-result string columns. Missing or non-string policy values retain syntax diagnostics. Command completion includes both policies. Unknown graph-model name references now use upstream's generic `KS142`; `graph()` lookup keeps its separate graph-model diagnostic. |

The [November source-oracle fixture](kustolang/engine/testdata/upstream_dbd7ca7_behavior.json)
and [native regressions](kustolang/engine/upstream_nov08_test.go) compare 18 valid
and malformed policy commands, result schemas, diagnostics, datetime cases,
and grammar behavior. The two existing corpora (258 queries and 231 commands)
are independently replayed against this exact revision. Historical source
fixtures remain available. Provenance records the current source revision and
hashes of the comparison tools and evidence.

## 2026-11-03 — timespan parsing allocation changes

Source: [`7e658f7`](https://github.com/microsoft/Kusto-Query-Language/commit/7e658f73457320fea8fde24c5f1b5a7e2a87ca59).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Literal parsing | `SyntaxToken.GetTimeSpanValue`, `CultureUtilities.TryParseDoubleInvariant(ReadOnlySpan<char>, out double)` | Source update: modern .NET builds parse number and unit spans without allocating substrings. The Go implementation already slices strings without copying. Numeric parsing and unit behavior are preserved; existing source-oracle literal cases, including rejection of group separators, are replayed under five cultures. No Go runtime change is required for this optimization. |

## 2026-11-03 — cancellation, keyword calls, and command recovery

Source: [`95874ea`](https://github.com/microsoft/Kusto-Query-Language/commit/95874ea5fcd625c391fe62838c60b2aebbf2312b).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Code-service cancellation | `KustoCodeService.TryGetBoundCode`, `GetDiagnostics`, `GetAnalyzerDiagnostics`; Go counterparts | Cancellation propagates as `OperationCanceledError` instead of an empty result or a cached `KS203` analyzer failure. Retrying the same service with a live token produces its diagnostics. Regressions cover binding, diagnostic traversal, analyzer cancellation, concurrent live callers, and absence of cached failure state. |
| Computed function expansion | `Binder.GetFunctionCallExpansion`, `GetDeferredFunctionCallExpansion`, `TryBindCalledFunctionBody`; Go `deferredComputedFunctionExpansion` | A deferred expansion no longer inherits an already-canceled originating bind. Cached `Rows("T")` results can still expand their body after the original token is canceled. Go already avoids storing canceled synchronous expansions; that behavior is retained. Explicit tokens on `TryBindCalledFunctionBody` were already supported. |
| Parse entry point | `KustoCode.ParseAndAnalyze`; Go `ParseAndAnalyzeWithCancellation` | Upstream now forwards the supplied token to creation/binding. Go already forwards it; existing cancellation regressions remain applicable. No Go runtime change is required here. |
| Query grammar | `QueryGrammar.ScanFunctionCall`, `HasAnyExpression`, `HasAllExpression` | Keyword identifiers such as `partition` and `render` now work as user-function call names in the grammar parser. `typeof` and `inline_external_table` keep their dedicated parsing paths. The source oracle also confirms the new `(` completion after keyword identifiers and the changed `fork (count)` / malformed `__partitionby` prefix boundaries; existing expectations now follow those results. Go already emits the correct `HasAnyExpression`/`HasAllExpression` node kinds; dedicated regressions retain this behavior. |
| Command recovery | `CommandGrammar.bestCommandOrPartial`; Go `parseKnownCommand` | A complete command at end of input skips the expensive partial-parser search. Incomplete commands still use recovery. The existing query/command differential corpora validate retained parsing and diagnostics. |
| Source generation | `CodeGenerator.WriteLine(string, bool, bool)` | Source update: adds an unindented output option using `WriteLineNoTabs`. The development oracle builds with the updated generator; the Go runtime has no corresponding text-writer API to change. |

## 2026-11-02 — datetime literal interpretation warnings

Source: [`ebca795`](https://github.com/microsoft/Kusto-Query-Language/commit/ebca7954208c09443825d104aa8ea343a4f6270c).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Datetime analysis | `DateTimeLiteralAnalyzer`, `KustoDateTimeParser.Parse`, `KustoAnalyzerCodes`; Go `NewDateTimeLiteralAnalyzer` and internal `parseKustoDateTime` | Adds correctness warnings `KS517` for unreadable literals and `KS518` for ambiguous numeric literals. `datetime(2026)` warns that it means the year 2026; `datetime(0)` warns that it means seconds from 1970; large tick counts include their exact seven-digit fraction in the message. Explicit dates remain warning-free. The source-oracle cases cover numeric boundaries, fixed-width forms, UTC offsets, fraction rounding, null/now, invalid dates, named months, accepted yearless/time-only shorthand (without false invalid warnings), and UTF-16 locations after an emoji. |
| Literal token values | `SyntaxToken.GetDateTimeValue`; Go `dateTimeLiteralValue` | The new service-literal parser is used by the analyzer, not by token-value conversion. The existing invariant token parser remains unchanged, matching upstream's separation of these APIs. |

## 2026-11-01 — 13.0.0 source synchronization: cache factories and grammar reuse

Source: [`82267c7`](https://github.com/microsoft/Kusto-Query-Language/commit/82267c7628e70898cd8bec54c20fc4e61c0a523a).
Microsoft's source version remains **13.0.0**.

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Parser-source cache | `SourceCache.GetOrCreate<T>(Func<T>)`; Go `SourceCacheGetOrCreateFunc` | A factory now runs once on a cache miss, rather than twice with the first result discarded. A hit reuses the stored object without another factory call. A throwing factory leaves no entry, so a later call can retry. |
| Global-state cache | `KustoCache.GetOrCreate<T>(Func<T>)`; Go `KustoCacheGetOrCreate` | Removes the unused preliminary factory call in C#. Sequential misses now call the factory once and hits reuse the stored object. Go already provided this behavior, so no runtime change is needed here. Go retains its stronger serialized factory execution under concurrency. |
| Query grammar caching | `QueryGrammar.From`, `CreateGrammar`, `CreateGrammars`; Go `QueryGrammarFrom` | Replaces the most-recent grammar plus per-global cache with eight lazy process-wide instances, keyed by `AlwaysProduceEndToken`, `AllowLiteralsWithLineBreaks`, and `AllowNonAdjacentWildcardParts`. Switching options away and back now reuses the same grammar even without a global cache; different schemas, cache settings, and parser kinds share the corresponding instance. Go publishes each variant with `sync.Once` and no longer stores a grammar in a schema's `KustoCache`. |

The [cache/grammar source-oracle fixture](kustolang/engine/testdata/upstream_82267c7_caching.json)
and [native regressions](kustolang/engine/upstream_nov01_cache_test.go) cover
factory invocation counts, retry, all eight variants, reuse, and concurrent
publication. The [current compatibility fixture](kustolang/engine/testdata/upstream_82267c7_compatibility.json)
reruns the 13.0.0 string, literal, scanner, and completion cases. Both oracle
fixtures agree under en-US, fr-FR, tr-TR, da-DK, and cs-CZ. Both existing query
and command corpora (489 cases total) were replayed against this exact source
revision and the Go port; their canonical outputs are unchanged.

## 2026-10-27 — Visual Studio solution format

Source: [`ed2732f`](https://github.com/microsoft/Kusto-Query-Language/commit/ed2732f869244b584be4d6d8d38a70dcc2074524).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Build tooling | `src/Kusto.Language.sln`, `src/Kusto.Language.slnx` | Source update: replaces the legacy Visual Studio solution with an XML solution listing the language and Bridge projects. No language functions or Go runtime behavior change. The source-oracle tool builds its own project and needs no solution-format adaptation. |

## 2026-10-22 — 13.0.0

Source: [`328e05d`](https://github.com/microsoft/Kusto-Query-Language/commit/328e05d10d6c73654e416675fc75af7ed0e05ffd),
upgrading the previously pinned 12.4.1 source.

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Wildcard matching | `KustoFacts.Matches`, database wildcard lookup | Ignoring case now applies to middle segments: `*SAMPLE*` matches `Samples`. Leading, middle, and trailing segments also match supplementary letters and case pairs with different UTF-8 byte lengths. |
| Ordinal equality | `CultureUtilities.Equals`, `SymbolMatch.Matches`, `ValueComparer.AreEquivalent`, `KustoFacts.IsHostName`, connection-part keys | `I` equals `i`; dotless `ı`, long `ſ`, and Kelvin sign `K` remain distinct from ASCII `I`, `S`, and `K`. Deseret and Osage uppercase/lowercase pairs compare equal. |
| Ordering | `CultureUtilities.Compare`, `StringAndNumberComparer`, `Binder.OrderColumns`, `KustoCompleter.GetCompletionItems`, code-action ordering | Completion text uses ordinal UTF-16 ordering, including punctuation and supplementary characters. Natural sorting folds whole supplementary characters. Full ignore-case comparison and bounded UTF-16 segment comparison preserve their distinct upstream ordering semantics. |
| Literal values | `SyntaxToken` real/int/long/decimal/datetime/timespan value helpers | `real(1,5)` and `real(1,,5)` yield `15`; `real(1,.5)` yields `1.5`. Commas after the decimal point or exponent yield the default value. Decimal and timespan components reject group separators. Typed `int(012)` and `long(012)` yield `12`; overflow yields zero. `datetime(01/02/2026)` means January 2. |
| Completion examples | `KustoCompleter.CreateValueExamples`, `CultureUtilities.ToInvariantString` | Values use invariant formatting: `datetime(10/22/2026 13:14:15)`, `timespan(1.02:03:04.0050001)`, `decimal(1.25)`, and canonical GUID text. Invariant lowercasing preserves dotted capital `İ`. |
| Character scanners | `Parsers.Match(char)`, `Parsers.Match(string)`, `TextSource.Matches` | String scanners compare complete Unicode scalars. A `TextSource` restricted to one character no longer matches a two-character token from the underlying string. Partial-scan failure positions are retained. |
| Editor and parser string operations | `EditString`, `CompletionText.Parse`, client directives, formatter, function-body and entity-group wrappers | Source update: culture-sensitive search, prefix/suffix checks, casing, and numeric conversion now use the shared invariant/ordinal helpers. Go's existing exact string operations already provide the corresponding culture-independent search behavior. |

The source-oracle fixture in
[`upstream_328e05d_compatibility.json`](kustolang/engine/testdata/upstream_328e05d_compatibility.json)
records the measured compatibility cases. Both existing differential corpora
(258 queries and 231 function-management commands) were rerun against this
revision; their canonical outputs are unchanged. Literal-value parse failures
continue to return default values without introducing diagnostics.

Microsoft documents remaining host-specific casing differences for a small
set of BMP characters on NLS versus ICU. The checked-in fixture records the
.NET 10 ICU source build used for this port's comparison. Native Go byte
scanner coordinates and explicit UTF-16 scanner coordinates remain separate.

## 2026-10-15 — JavaScript publishing tooling

Source: [`1d44c7b`](https://github.com/microsoft/Kusto-Query-Language/commit/1d44c7bbcb7b1bd7bc1e7463aa3d2408e0df126d).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Build and distribution | Bridge `publish.cmd`: `Build`, `MSbuildSetup` | Source update: locate MSBuild through `vswhere`, report missing tooling, and propagate build failures. No native Go runtime change; this commit does not change tracked language sources. |

## 2026-10-07 — Policy commands and computed-column completions

Source: [`6b5c2f7`](https://github.com/microsoft/Kusto-Query-Language/commit/6b5c2f739a14bd5de52115f9bf9695985f0852dd),
the previous 12.4.1-based pin.

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Change-data-capture policies | `ShowTablePolicyChangeDataCapture`, `AlterTablePolicyChangeDataCapture`, `AlterTablesPolicyChangeDataCapture`, `DeleteTablePolicyChangeDataCapture` | Adds show, alter, bulk-alter, and delete table policy commands with capability lists. Existing source-oracle regressions cover valid and malformed forms. |
| Materialized-view prioritization | `ShowClusterPolicyMaterializedViewPrioritization`, `AlterClusterPolicyMaterializedViewPrioritization`, `AlterMergeClusterPolicyMaterializedViewPrioritization`, `DropClusterPolicyMaterializedViewPrioritization` | Adds four cluster policy commands. Show returns `EffectivePolicy` alongside the ordinary policy fields. |
| Computed-column examples | `KustoCompleter.AddColumnExamples`, `GetPossibleValueExamples`, `CreateComputedColumnExampleCompletion`, `TypeFacts.IsCommonArgumentTypeKind` | Derives literal completions from shared-type arguments of `iff`, `case`, `coalesce`, and related functions. Preserves original numeric syntax and hidden compound strings; limits work and example length to 100, with length measured in UTF-16 units. |

## 2026-10-05 — DAX and GQL plug-ins

Source: [`8aa8e70`](https://github.com/microsoft/Kusto-Query-Language/commit/8aa8e708fe30351395a1b05d9af452a8c7657023).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Evaluate plug-ins | `PlugIns.DaxRequest`, `PlugIns.GqlRequest` | `dax_request` and `gql_request` become visible in completion. Each accepts exactly two required string arguments and returns an open runtime schema. Optional parameter bags and options are removed until supported by the engine. Existing regressions cover visibility, argument count/type errors, and schema state. |

## 2026-09-25 — Database row-ID policies

Source: [`9d09d29`](https://github.com/microsoft/Kusto-Query-Language/commit/9d09d2960fd502885072c38ce6732a44d84fe5f2).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Database policy commands | `ShowDatabasePolicyRowId`, `AlterDatabasePolicyRowId`, `DeleteDatabasePolicyRowId` | Adds database-level show, alter, and delete `rowid` policy commands with the standard policy result schema. Alter accepts `true` or `false`; show also accepts `*`. Existing command regressions cover the added grammar and schemas. |

## 2026-08-30 — Table details and capacity

Source: [`e3d07b0`](https://github.com/microsoft/Kusto-Query-Language/commit/e3d07b079dfe2811a87fb7d72d548951f64a8dbd).

| Area | Impacted functions or symbols | Confirmed behavior or update |
| --- | --- | --- |
| Table metadata | `ShowTablesDetails`, `ShowTableDetails`, shared `ShowTablesDetailsResult` | Adds string columns `RowLevelSecurityPolicy`, `AutoDeletePolicy`, `UpdatePolicy`, and `UpdatePolicyStates`. Existing catalog regressions check the result schemas. |
| Capacity commands | `ShowCapacity` | Accepts the `update-policy-over-external-table` resource. Existing command regressions cover parsing the new resource. |
