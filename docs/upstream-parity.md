# Upstream parity

This library ports Microsoft Kusto.Language behavior to native Go from the
Apache-2.0 source revision recorded in `third_party/kusto/UPSTREAM.json`. It
does not disassemble or load `Microsoft.Azure.Kusto.Language`, and no .NET
runtime is used by the library or binary.

## Language and editor coverage

The detailed coverage below was carried forward from the combined project.
Application-level recovery and operational controls are maintained in
[KustoKITT](https://github.com/olafhartong/KustoKITT/blob/main/docs/analyzer-parity.md);
this repository owns the source port, language facade, and editor services.

The practical analyzer milestone covers the supported KQL workflows:

- the current parsing implementation, including recoverable malformed syntax
  and source-compatible diagnostics;
- semantic binding, result schemas, referenced symbols, configured database
  functions, plug-ins, commands, and environment overlays;
- source-equivalent `project`, `project-away`, `project-keep`,
  `project-rename`, `project-reorder`, and `extend` recovery for named scalar,
  tabular, tuple, error-valued, unresolved, and missing expressions, including
  first-failure scalar diagnostics, compound rename diagnostics, declaration
  identity, and deterministic result schemas after malformed input;
- source-equivalent `print` tuple projection, including named first-member
  selection, one-column tuple unwrapping, multi-column tuple expansion, and
  parser-recovered missing closing parentheses;
- source-equivalent `reduce by` exact-string validation and output-column
  source lineage across valid, type-invalid, unresolved, and missing
  expressions, including downstream physical dependency resolution;
- filter/where recovery for missing predicates, incomplete table paths, and
  unterminated `in (` lists, retaining predicate, operator, and downstream
  semantics while preserving exact parser diagnostics;
- facet group identity and generated-column lineage, including exact
  with-clause table reuse, count-column expression/original-column sources,
  dynamic non-column diagnostic short-circuiting, and parenthesized operator
  behavior;
- analyzer rules, fixes, formatting, completion, quick info, classification,
  outlines, related elements, and code scripts;
- public lexer, `KustoCode`, typed syntax tree, grammar productions, parser
  combinators, syntax construction, token scanners, Kusto/Syntax/Text facts,
  option/property catalogs, typed global settings, function-body facts, and
  the type-keyed Kusto cache;
- custom-node descriptor authoring with source-shaped default-shape and
  descriptor factories, including the pinned `WithIsOptional` no-op bug under
  its source name and the behavior-correct existing Go `WithOptional` method;
- the remaining `TokenParser` metadata tail as package functions:
  `ParseTrivia` returns consumed trivia text and `ScanRawGuidLiteral` exposes
  the exact 8-4-4-4-12 scanner already shared with the native lexer;
- immutable query-operator parameter metadata and the complete public
  `QueryOperatorParameters` catalog, with defensive list access and typed AST
  lookup, literal-value, alias, and name-value helpers;
- the complete built-in `Operators` symbol/signature catalog with pinned
  upstream ordering and defensive list access;
- the complete public `Aggregates` and `PlugIns` symbol/signature catalogs,
  including upstream obsolete, hidden, optimized-alternative, and list-membership
  metadata, the five public plug-in result-schema inventories, and the public
  aggregate `GetAnyResult` custom-return helper;
- all 429 public built-in `Functions` symbols from the pinned handwritten and
  generated conversion catalogs, emitted by deterministic standard-library Go
  AST generation and shared directly with the binder;
- all 637 generated engine command symbols and the pinned non-engine server
  command catalogs, preserving result schemas, constructions, ordering, and
  the exact symbols used by command parsing;
- table change-data-capture and cluster materialized-view prioritization policy
  commands, including the effective-policy show result, matching the 2026-10-07
  upstream additions;
- bounded computed-column value completions from shared-type arguments of
  `iff`, `case`, `coalesce`, and similar functions, preserving source literal
  syntax, numeric literal types, and compound hidden strings. A source oracle
  built at the pinned revision supplies dedicated policy and completion
  regressions, including malformed commands and UTF-16 example-length limits;
- visible `dax_request` and `gql_request` plug-ins with exactly two required
  string arguments and an open runtime result schema, matching the 2026-10-05
  upstream source revision;
- the complete `FunctionHelpers` custom-return-type helper surface and public
  tuple construction, including constant extraction, referenced-column
  propagation, summarize grouping, naming, and boolean classification;
- the complete usable `TypeFacts` surface, including common/widest type
  inference, conversion and promotion checks, scalar/dynamic categories, and
  source-compatible fallback, ignore-dynamic, and column union/intersection
  behavior, including the pinned public `IsAnyScalarExceptReadOrBool` typo and
  corrected Go spelling;
- public `CombineKind` and `ResultNameKind` metadata plus source-equivalent
  column/table combine entry points, preserving union/find/unique-name schema
  rules and original-column lineage across valid and recovered malformed KQL;
- public scope/empty-scope lookup, typed fake-expression construction, and a
  UTF-16-positioned generic `SubstringMap`, including diagnostic-free missing
  token identity used by upstream semantic helpers;
- source-shaped syntax-token factories for identifiers, keywords, punctuation,
  operators, literals, other tokens, and missing recovery tokens, including
  category validation, normalized value text, literal prefixes, trivia, and
  multiple diagnostics;
- complete literal-token values, including primitive min/max forms, exact
  standard-library rational decimals, full-range signed 100 ns timespan ticks,
  `time.Time` datetimes, `[16]byte` GUIDs, nulls, and zero-value recovery for
  malformed manually constructed literals;
- public diagnostic description construction, immutable mutation, syntax-node
  location, absolute identity/equality, and source-compatible deduplication;
- the complete adapted public `DiagnosticFacts` catalog: all 205 pinned C#
  declarations and 197 unique current-source names, plus 15 explicitly labeled
  historical/internal/spelling compatibility factories. Generation verifies
  the exact vendored source SHA-256 before emitting native Go forwarders;
- typed `DiagnosticFacts` adaptations for syntax-kind, symbol, type-symbol,
  generic value, generic symbol, and named-clause overload groups, preserving
  C# joining, boolean presentation, diagnostic codes, messages, user-defined
  `ToString()` dispatch, and the source's distinct single-null interpolation
  versus multi-value null failure behavior;
- location-aware syntax diagnostic collection with duplicate retention,
  explicit include masks, postorder expansion/alternate ordering, UTF-16
  public positions, immutable syntax diagnostic replacement/addition, and
  origin-aware conversion that preserves public and JSON-round-tripped
  absolute UTF-16 locations;
- public `DynamicBagSymbol` construction and canonical empty-bag access, with
  immutable property lists, property lookup/update, and source propagation;
- the source-shaped `DynamicSymbol` base hierarchy and canonical
  `DynamicAnySymbol` singleton while retaining existing scalar type identity;
- public function custom-availability callbacks with binder locations, plus
  end-to-end optimized-alternative analyzer behavior for valid and recovered
  malformed function calls;
- public custom parameter-layout builders, fixed/repeating/skipping/block
  constants, and a source-shaped `ParameterLayouts` object catalog. The
  additive `ParameterLayoutStrategy` interface and
  `Signature.WithParameterLayout` support externally defined mapping,
  completion, and count-validation behavior while preserving the comparable
  `Signature.Layout` field and legacy `WithLayout`/`WithCustomLayout` APIs;
- source-shaped signature argument-layout operations infer bound expression
  types, synthesize typed fake expressions for type-only custom layouts,
  preserve surplus custom-builder output, and retain named-argument mapping
  when block layout falls back to a non-repeating parameter set;
- expression- and caret-aware possible-parameter selection distinguishes an
  existing argument, its exact end, following whitespace, and recovered
  missing tails in editor context lookup. Tree binding uses layout-specific
  scope only after a single signature is known and retains upstream broad
  function scope for unresolved overloads;
- user and database function overload ranking applies the pinned implicit
  scalar-coercion policy per function while built-in resolution remains
  strict, including declaration-backed overload sets whose first signature
  identifies the complete set as a local user function;
- practical custom-return analysis derives tabular and dependent argument
  semantics from the layout-mapped parameter list rather than declaration
  indexes, including strategy-backed custom layouts;
- practical aggregate return inference delegates parameter/common/widest
  return kinds to the native `Signature` implementation, preserving named,
  repeating, custom, and block-layout argument-to-parameter mapping;
- public parameter authoring, syntax conversion, parse/render helpers, default
  metadata options, and declared-symbol, computed-body, declaration-backed,
  custom, and return-kind signature construction. Direct function-symbol
  authoring covers return-kind, custom-return, parsed parameter-list,
  explicit/inferred tabularity, and declaration-backed constructor forms;
- public pattern-symbol authoring includes name-only construction plus text and
  syntax function-body signatures, preserving empty text bodies as present and
  nil syntax bodies as absent while retaining declaration identity;
- public `TableState` flags and extension behavior, including the upstream rule
  that making a table unsorted preserves its serialized state;
- strict public `SyntaxTree` root attachment and typed `CopyAsFragment`
  behavior, preserving diagnostics and nested original-tree position mapping,
  with composite/list widths counting interior trivia as upstream does;
- the concrete parser-combinator hierarchy, native constructors, visitors,
  functional primary-parser extension points, parent inspection, safe parsing,
  partial parsing, and custom equality-comparer item/any/sequence matching,
  including non-comparable Go inputs. C# generic-arity overloads use explicit Go names such as
  `MatchValueParser`, `FirstValueParser`, `BestValueParser`, and
  `IfValueParser`; `ResultParserCast` preserves parser tags and byte/UTF-16
  string conversion overloads cover native UTF-8 and C# code-unit inputs.
  List-producing repetition helpers preserve independent result ownership;
- the public character-scanner facade for characters, strings, ASCII letters,
  digits, hexadecimal digits, whitespace, and line breaks, with parser-native
  byte offsets and ranges documented by external tests;
- a parallel native UTF-16 code-unit character-scanner facade matching C#
  `Parser<char>` lengths, ranges, surrogate-pair behavior, failure offsets,
  line breaks, tags, and simple character case mapping;
- source-shaped public `TextFacts`, `TriviaFacts`, `CommentFacts`, and
  `ScriptFacts` static-class facades, using .NET-compatible UTF-16 line starts,
  offsets, trimming ranges, indentation, comment spans, and block ranges across
  CRLF, Unicode line separators, astral text, quoted text, and recovered
  unterminated multiline strings. Existing free-function parser utilities keep
  their byte-coordinate compatibility contract;
- the public parser `Describer.Describe` behavior for tagged, hidden,
  sequenced, alternative, optional, repeated, required, wrapped, and applied
  parser trees, including required-marker suppression;
- strict public `SafeList<T>` immutable snapshots with atomic ownership
  transfer and independent branching, plus the non-nil public
  `EmptyReadOnlyList<T>` shape;
- strict public facades for pinned internal utility models: UTF-16 ranged
  `TextKeyedDictionary<T>` and `StringTable`, `ConnectionInfo`, `Optional<T>`,
  `UniqueNameTable`, comparer-aware concurrent dictionaries, bounded object
  pools, most-recently-used caches, `ValueComparer`, and
  `StringAndNumberComparer`. Collection access is defensive, custom Go key
  equality uses callbacks, and the pinned cache behavior that refreshes but
  does not replace an existing value is preserved;
- public native Common/Utils facades for dictionary add/update, validation,
  hash-set range insertion, atomic compare/exchange, immutable list conversion,
  list search/replacement/distinct operations, callback cancellation, the
  shared `NoDiagnostics` collection, and append-in-place line-start discovery;
- additive dual-coordinate syntax metrics and navigation: explicit `Byte*` and
  `UTF16*` widths, trivia/text bounds, sibling offsets, token and node lookup,
  ranged token walking, and original-tree fragment projection, including
  astral text and zero-width missing tokens while preserving current
  unsuffixed byte behavior. Byte-sensitive binder, query-parser, and built-in
  editor/action paths now call explicit `Byte*` methods; intentional
  `Width() == 0` missing-element checks remain coordinate independent;
- all 46 public `QueryGrammar` missing-node recovery factory names and both
  token-literal overload behaviors, plus the complete 35-entry string-operator
  map. Recovery preserves nested diagnostic ownership, keyword-aware missing
  expressions, fresh node identity, and defensive map copies;
- predefined command function rules reuse the public function body behind an
  explicit forward-parser barrier and construct declarations without the
  query grammar's optional `view` keyword, preserving bounded partial-command
  recovery for generated create/alter function commands. Rejected command
  tails publish no result type, and unresolved command function references use
  the pinned generic `KS142` name diagnostic;
- the complete bounded `CommandGrammar` authoring helper block: custom nodes
  and descriptors, command wrappers, command token/completion and required-token
  recovery, and separated/unseparated list construction with source-equivalent
  leading, doubled, and trailing separator behavior;
- observable `CommandGrammar` runtime helpers for command limiting/start
  detection, bad and missing command recovery, expected-term tag traversal,
  partial-parser failure insertion, and diagnostic-based syntax selection,
  without exposing or mutating grammar caches. Literal-token parser tags retain
  the source's empty-tag behavior so command recovery never leaks numeric Go
  enum values into expected-term diagnostics;
- concrete `EngineCommandGrammar`, `DataManagerCommandGrammar`,
  `ClusterManagerCommandGrammar`, and `AriaBridgeCommandGrammar` types with
  source-shaped constructors and a cached typed factory view over the same
  base grammar state. The existing base-returning factory remains available so
  current Go callers are not broken by the C# inheritance adaptation;
- source-shaped `CommandParserInfo` provider and factory composition for custom
  command grammar sets, preserving parser tags, best/partial selection,
  tied command kinds, expected-term recovery, annotation-driven completion,
  required-token recovery, and unknown/bad fallbacks directly from arbitrary
  parser trees without requiring generated descriptor metadata;
- the complete generic ancestor/descendant traversal family, including
  ancestor-or-self, descendant-or-self, typed filtering, nil-safe lists,
  direct-child/descendant indexes, and parent child-name lookup;
- node-only syntax walking and source-compatible tree depth that exclude token
  leaves and support subtree pruning;
- source-shaped default syntax visitors with generated forwarding for all 213
  syntax-node methods, typed result visitors, and per-node overrides;
- all three upstream parser-visitor forms, including argument-carrying typed
  result dispatch across concrete parser kinds and malformed parse attempts;
- the public lexical-token `IsParentVisitor` singleton, preserving pinned
  parent semantics across composite, repetition, failure, and nil parsers;
- strict public `EditString` and client-directive facades, including immutable
  collective edit tracking, UTF-16 position mapping, typed directive arguments,
  malformed directive recovery, and defensive list access;
- public generic `ParserPath` and `ParserAndIndex` editor models with parent,
  child, and index traversal plus race-safe lazy child caching;
- public completion-display generation and concrete Kusto, unknown, and offset
  code-service facades, plus composable keyed service factories and custom
  `CodeScript` factory routing, existing-code service construction, and
  immutable script factory/globals/separator replacement, including nested
  Kusto-factory discovery and sibling-factory preservation;
- custom Kusto actor dispatch preserves the existing byte-coordinate actor
  contract while adapting the editor service's UTF-16 positions, selections,
  returned edits, and caret positions; actors with explicit UTF-16 methods are
  dispatched directly and opaque action data remains untouched;
- typed concrete Kusto code-service factory recovery from direct and aggregate
  factories, preserving `Globals`, `WithGlobals`, and the existing predicate
  lookup API;
- a public callback-based custom `CodeService` adapter covering every service
  operation, with immutable source text, common-service defaults, offset
  translation, factory composition, and multi-block `CodeScript` routing;
- source-shaped public static value catalogs for `CodeKinds`,
  `CodeServiceFeatures`, `DiagnosticCategory`, `DiagnosticSeverity`,
  `ServerKinds`, and `KustoActorNames`, backed by the same native constants
  consumed by code services, diagnostics, global state, and editor actors;
- a source-shaped `CommandFacts` facade whose line starts, command block
  starts, and command text ranges use the upstream UTF-16 coordinate system;
- virtual-style custom combined-diagnostic dispatch with fresh list storage
  before concatenation/filtering, plus common classification fallback through
  custom client-parameter callbacks rather than a statically bound base;
- source-compatible `KustoCode` identity reuse for unchanged analysis/globals,
  immutable internal text as the facade's coordinate source, and cancellation
  propagation through scope, speculative-symbol, and row-scope syntax walks.
  Its one-based line/offset mapping retains UTF-16 positions between surrogate
  code units instead of collapsing them to a UTF-8 byte boundary;
- race-safe, defensively copied `KustoCode` syntax/all-diagnostic caches,
  cancellation propagation before analyzer execution, and upstream-compatible
  analyzer-failure conversion when an analyzer itself throws cancellation;
- source-shaped default-option and default-scope conveniences for outlines,
  completion, quick info, formatting, symbols-in-scope, and speculative symbol
  lookup, plus the remaining optional `CodeService` call forms for feature
  support, diagnostics, actions, application, and classification, delegating
  to the same explicit pinned defaults as their optioned forms;
- package-owned `OutliningOptions`, `OutlineRange`, and `OutlineInfo` facades
  with pinned defaults, immutable-style option helpers, and defensively owned
  range collections at the public/internal boundary;
- source-shaped syntax-parser EOF controls, default token start, token-list
  append behavior, and the public `KustoCode.GetKind` spelling;
- source-compatible UTF-16 editor limits and caret boundaries: maximum parse
  size and offset-service action clamping count UTF-16 code units, while
  read-only quick-info and element lookups accept positions between surrogate
  halves without allowing edit ranges to split UTF-8 runes;
- quick-info and element caret inspection contain parser or authored-tree
  traversal failures, returning canonical non-nil empty quick-info items and
  the zero text range instead of exposing a panic to editor clients;
- the complete public completion-hint flag enum used by typed syntax
  construction and editor integrations;
- public completion-text creation and marker parsing, option-based completion
  item authoring with rank/priority metadata, deep defensive ownership of item
  and nested apply-text collections, and completion-result construction,
  including upstream post-pipe operator legality and hidden graph-operator
  filtering;
- public disabled-diagnostic construction, builders and capability probes,
  plus source-shaped text-range overlap and replacement/deletion/insertion
  edit factories;
- the complete public code-action construction and result factory surface,
  including fix-all/menu composition, immutable data snapshots, change/move/
  rename sequencing, UTF-16 `EditString` overloads, canonical empty results,
  source-compatible failure defaults, nullable failure-state distinction, and
  nominal embeddable bases for external code/apply/result action types. Service
  boundaries recursively isolate built-in action/result graphs without changing
  constructor object-identity behavior;
- computed-function tabularity inference equivalent to upstream
  `Signature.ComputeTabularity(GlobalState)`, using syntax classification then
  native binding for scalar, tabular, parameterized, database-function,
  braced-body, and malformed-body cases;
- public editor result construction for formatted text, analyzer information,
  client parameters, syntax/entity references, outlines, classifications,
  quick info, and related elements, including canonical empty values and
  defensive list accessors. Native caches and public/custom service boundaries
  snapshot outer and nested collections, client-parameter indexes, analyzer
  diagnostics, quick-info parts, and completion apply text;
- exact `MinimalTextKind` to `IncludeTrivia` mapping for Kusto and common code
  services, including rejection of unknown enum values as upstream does;
- UTF-16-budgeted outline collapsed text with valid UTF-8 rune boundaries and
  source-compatible service fallback when invalid options make outlining fail;
- programmatic construction of open and alternate-name database/cluster
  schemas, options, entity groups and elements, graph shapes/models/snapshots,
  groups, table-backed commands, schema-derived tuples/graphs, and canonical
  unknown/empty/error/void/parameter symbol instances;
- public symbol member lookup, assignability, and table equivalence helpers,
  including upstream-compatible case-insensitive column lookup and the exact
  pinned stateful lazy-cache behavior of `DatabaseSymbol.Contains`, plus
  last-declared database precedence for primary and alternate cluster names;
- source-compatible `GlobalState` catalog boundaries and default environment:
  `GetFunction` searches ordinary functions with last-name precedence,
  aggregate and plug-in lookups remain separate, built-in ownership follows
  the selected symbol, and the default uses unknown database/cluster symbols;
  `GetAnyFunction` and `NewGlobalState(nil)` explicitly preserve the earlier Go
  all-source lookup and closed synthetic `db` conveniences. Empty server kinds
  select no command catalog, and derived table ownership recognizes external
  and materialized-view wrapper and embedded identities; `GetAnyTableSymbol`
  preserves the concrete wrapper identity where Go embedding differs from C#
  inheritance;
- source-ordered defensive `ScalarTypes.All`, strict ordinal `GetSymbol`, the
  upstream `ScalarTypes.Type` spelling with `TypeLit` compatibility, a named
  package-owned `ScalarTypesAPI` facade, dynamic bag/tuple schema construction,
  and symbol-wide common-type inference for regular and table-derived symbols;
- source-shaped instance navigation on the public `SyntaxElement` interface,
  including missing-child checks, child/descendant indexes, depth, ancestry,
  parent names, token snapshots, and token/element walks;
- source-shaped instance semantics on the public `Expression` interface,
  including result/raw-result types, constant and literal classification, and
  typed `ValueInfo` access. `VariableSymbol` constant metadata uses the same
  typed representation while retaining a raw-value compatibility constructor;
- direct simple-name access and token/name construction conveniences for
  public `NameDeclaration` and `NameReference` nodes;
- a named package-owned `SyntaxFactsAPI` facade over the complete native kind,
  category, text, operator, identifier, keyword, and fixed-text inventories;
- the pinned `invoke` completion exceptions for `cluster()` and `database()`,
  which remain candidates even when the incoming row schema does not satisfy
  ordinary first-parameter filtering;
- source-shaped scalar-infix completion metadata and filtering for logical,
  membership, list, and range operators, including specialized insertion text,
  nearest-expression context, UTF-16 positions, malformed-statement recovery,
  and the pinned post-pipe operator inventory;
- the complete public `QueryParser` text/token/start/options overload families
  and a named package-owned `QueryParserAPI` facade for queries, expressions,
  literals, entity paths/groups, row schemas, function parameters, individual
  parameters, and function bodies, with recoverable malformed input;
- explicit lexical-token `ByteLength` and `UTF16Length` accessors, including
  trivia and astral characters, while retaining `Length` as the existing byte
  compatibility spelling;
- public symbol presentation and construction metadata, including schema and
  completion display, primitive scalar flags, server/materialized-view kinds,
  and argument/parameter/return/layout/conversion/operator enums;
- an explicit comparable `BlockSeparatorHandle` adaptation for C# delegate
  identity, allowing `CodeScript.WithSeparatorHandle` no-op reuse without
  unsafe Go runtime-layout assumptions while retaining the function API;


The checked-in corpus intentionally includes both valid and invalid KQL. A
cross-surface batch covering functions, local declarations, union, join,
lookup, parse-kv, make-series, partition, fork, facet, scan, top-nested, and
malformed recovery was also compared directly with the pinned Microsoft
library; representative cases are permanent native tests and require no .NET
runtime. The full repository tests, race checks for the language packages, `go
vet`, generated-code determinism, and local provenance
validation are the required gates for changes to these surfaces.

A separate development-only mutation batch removes delimiters, truncates at
structural boundaries, and appends incomplete tails to ten valid queries that
exercise local scalar/tabular functions, union, join, parse-kv, make-series,
partition, scan, mv-apply, and range. The resulting 416 current-profile queries
are compared against the pinned C# source build for ordered diagnostics and
result columns. The latest stable run matches all 416 queries exactly after
normalizing JSON object-key order; the ordered diagnostics and sorted result
columns are otherwise byte-for-byte equal. The source oracle is built from the
pinned public repository revision, including its checked-in T4 generation
templates. It is not linked into, loaded by, or distributed with the Go library
or binary, and representative recovery cases remain as permanent native tests.

A second development-only batch applies the same structural mutations to
facet, fork, top-nested, distinct, sample-distinct, parse, parse-where,
serialize/project-rename, search, and find queries. Its 304 current-profile
queries also match the pinned C# oracle exactly for ordered diagnostic code,
message, UTF-16 span, and sorted result columns. This batch specifically keeps
standalone missing-pipe recovery, empty bracket selectors, and synthesized
trailing clauses under differential coverage.

A third development-only batch contains 474 valid and malformed current-profile
queries covering graph-match, graph-shortest-paths, make-graph followed by
graph-to-table, lookup, evaluate bag_unpack, invoke, mv-expand, and mv-apply.
The current native result matches all 474 cases exactly for ordered diagnostic
code, message, UTF-16 span, and sorted result columns. Permanent regressions
cover incomplete make-graph endpoints, missing mv-expand type names,
dynamic-object comma recovery, graph-to-table trailing commas, standalone and
piped invoke implicit arguments, qualified graph projection names, malformed
graph edges, detached graph operators, skipped compound graph punctuation, and
typed malformed-let prefix ownership. The oracle remains development-only and
is neither linked into nor distributed with the Go library or binary.

A fourth development-only batch contains 298 valid and malformed
current-profile queries covering partition, make-series, parse, fork,
top-nested, and scan. It matches all 298 cases exactly for ordered diagnostic
code, message, UTF-16 span, and sorted result columns. Permanent regressions
cover adjacent compound string literals in datatable values, malformed
serialize projections, incomplete datetime literal tails, and make-series
range type diagnostics. The fixes retain parsed source prefixes and skipped
tail ownership instead of discarding them through opaque fallback syntax.

A fifth development-only batch contains 345 valid and structurally malformed
current-profile queries covering project-away, project-keep, project-rename,
project-reorder, extend, distinct, sample, sample-distinct, summarize, sort,
take, top, parse-kv, serialize, and count. It matches all 345 cases exactly for
ordered diagnostic code, message, UTF-16 span, and sorted result columns. The
batch verifies recovered projection schemas, summarize grouping names, nested
aggregate context diagnostics, deleted delimiters, truncations, and incomplete
operator tails. Across all five batches, 1,837 queries now match the pinned C#
oracle exactly.

A sixth development-only batch contains 300 valid and structurally malformed
current-profile queries covering declare-query-parameters, restrict, pattern,
assert-schema, getschema, render, execute-and-cache, consume, macro-expand,
externaldata, and inline external tables. It matches all 300 cases exactly for
ordered diagnostic code, message, UTF-16 span, sorted result columns, open-table
state, and result type. The fixes preserve top-level skipped-token recovery,
publish declaration-less patterns, attach macro alternates to full statement
lists, and reject malformed inline table syntax without typed-nil panics. Across
all six batches, 2,137 queries match the pinned C# oracle exactly, including
the result-state fields now replayed against the first five batches.

A seventh development-only batch contains 258 deterministic valid and
structurally malformed current-profile queries covering `as`, `filter`,
`project-by-names`, `reduce by`, `top-hitters`, graph operators, alias and set
statements, and `__partitionby`. The standard-library generator is checked in
at `tools/gen-kusto-parity-corpus`; it deletes structural delimiters, truncates
at structural boundaries, and adds explicit recovery cases. All 258 queries
match the pinned C# oracle exactly for diagnostic code and UTF-16 span, sorted
result columns, open-table state, and result type. Permanent regressions cover
standalone recovered `as`, no-row-scope `project-by-names` projection,
malformed `__partitionby` result publication, and valid and recovered print
tuple expansion. Across all seven batches, 2,395 queries match exactly.

An eighth development-only batch contains 231 deterministic valid and
structurally malformed function-management commands. Its standard-library
generator is checked in at `tools/gen-kusto-command-parity-corpus` and covers
show, create, create-or-alter, alter, drop, principal-role, and access-control
productions with truncations, deleted delimiters, broken property values,
principal lists, and incomplete function bodies. All 231 commands match the
pinned C# oracle exactly for ordered diagnostic code, message, and UTF-16 span,
sorted result columns, open-table state, and result type. Permanent regressions
cover typed-nil optional parser output, property-value expected terms, rejected
`view` tails, partial access-control commands, and unresolved altered-function
references. Across all eight batches, 2,626 queries match exactly.

The native port deliberately does not reproduce three pinned C# exceptions on
malformed input. Invalid `getschema kind=1` returns `KS141` and the default
schema instead of throwing `InvalidCastException`. Invalid inline external
table partition/path forms return recovery diagnostics instead of the pinned
`InvalidCastException` or `IndexOutOfRangeException`. These are bounded safety
differences: valid syntax and non-throwing malformed cases remain differential
parity targets.

## Remaining full-library work

Language coverage is not the same as a one-for-one Go representation of
every public and internal C# API. Full upstream parity still requires:

- continued representation and behavior work for internal parser, binder, and
  editor implementation classes that are not part of the public analyzer API;
- continued expansion of the exact current-profile structural mutation batches
  beyond their present 416-query, 304-query, 474-query, 298-query, 345-query,
  300-query, 258-query, and 231-query suites, especially when upstream changes
  grammar recovery, diagnostic ordering, result-shape behavior, or adds new
  statements and operators;
- culture-specific .NET formatting for arbitrary generic diagnostic values
  that do not provide a source-shaped `ToString()` method. The pinned code uses
  thread-local `CultureInfo.CurrentCulture`, which has no direct Go standard-
  library equivalent; null, booleans, and custom `ToString()` behavior are
  represented explicitly without a process-global locale shim;
- current-culture `.NET Convert.ChangeType` and string comparison behavior in
  `ValueComparer`. Native null, reference/value identity, NaN, numeric
  precedence, and conversion-failure behavior are represented, but Go's
  standard library has no per-thread `CultureInfo` equivalent;
- source-compatible UTF-16 widths and coordinates for direct public AST
  consumers under the existing unsuffixed method names. Analyzer/editor
  diagnostics and explicit `UTF16*` AST methods are UTF-16, but the current
  unsuffixed syntax methods remain UTF-8 byte based for backward compatibility;
- remaining inheritance and compile-time read-only collection details that do
  not map exactly through Go composition and aliases. Runtime cache/service
  ownership and defensive accessors are represented, but compatibility fields
  remain writable until their public aliases can be replaced safely;
- replacement of public Go type aliases with strict facade types where
  the alias exposes implementation fields that the C# type does not expose;
- continued API and behavior comparison whenever the pinned Microsoft source
  revision changes, followed by expansion of the valid and malformed parity
  corpus for new grammar, diagnostics, symbols, and editor behavior.

`go run ./cmd/kusto-upstream-check` reports whether Microsoft has advanced the
tracked revision and classifies changed grammar, function/operator, command,
diagnostic, and editor files. Upstream advancement is informational when the
complete comparison contains no changes within the configured `source_paths`.
Changes within those paths fail the check even when they have no surface
category, including files renamed into or out of a tracked path. Missing or
invalid comparison data, comparisons reaching GitHub's 300-file cap, and
behind or diverged histories still fail the check.
`go run ./cmd/kusto-upstream-check
-verify-provenance -local-only` validates the pinned commits, source paths,
license hash, third-party notice, and checked-in differential artifacts without
network access. Every attestation names the exact pinned commit and hashes its
corpus, generator, C# oracle output, and native output. A pin change therefore
fails local and CI provenance until a fresh differential batch is recorded.
