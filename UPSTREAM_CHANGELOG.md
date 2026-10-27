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
