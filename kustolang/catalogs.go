package kustolang

// Ported from src/Kusto.Language/Options.cs, Properties.cs, KustoCache.cs and
// FunctionBodyFacts.cs. See third_party/kusto/LICENSE for the preserved
// Apache-2.0 license and third_party/kusto/UPSTREAM.json for the pinned source.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// GlobalStateProperty is the non-generic base contract for typed properties.
type GlobalStateProperty = internal.GlobalStateProperty

// TypedGlobalStateProperty mirrors GlobalStateProperty<T>. Go uses package
// functions for the generic GetProperty and WithProperty operations because
// methods cannot introduce their own type parameters.
type TypedGlobalStateProperty[T any] = internal.TypedGlobalStateProperty[T]

// KustoCache is a thread-safe cache keyed by Go type, scoped to a GlobalState.
type KustoCache = internal.KustoCache

// FunctionBodyFacts describes source lookups, dependent parameters and errors
// discovered while binding an invoked function body.
type FunctionBodyFacts = internal.FunctionBodyFacts

// OptionCatalog contains the well-known query options from Options.cs.
type OptionCatalog struct {
	BestEffort                           *OptionSymbol
	DebugPython                          *OptionSymbol
	DeferPartialQueryFailures            *OptionSymbol
	DoNotImpersonate                     *OptionSymbol
	ErrorReportingPlacement              *OptionSymbol
	MaterializedViewShuffleQuery         *OptionSymbol
	MaxEntitiesToUnion                   *OptionSymbol
	MaxMemoryConsumptionPerIterator      *OptionSymbol
	MaxMemoryConsumptionPerQueryPerNode  *OptionSymbol
	MaxOutputColumns                     *OptionSymbol
	NoRequestTimeout                     *OptionSymbol
	NoTruncation                         *OptionSymbol
	ProgressiveProgressReportPeriod      *OptionSymbol
	ProgressiveQueryMinRowCountPerUpdate *OptionSymbol
	PushSelectionThroughAggregation      *OptionSymbol
	QueryBinAutoAt                       *OptionSymbol
	QueryBinAutoSize                     *OptionSymbol
	QueryCursorAfterDefault              *OptionSymbol
	QueryCursorBeforeOrAtDefault         *OptionSymbol
	QueryCursorCurrent                   *OptionSymbol
	QueryCursorDisabled                  *OptionSymbol
	QueryCursorScopedTables              *OptionSymbol
	QueryDataScope                       *OptionSymbol
	QueryDateTimeScopeColumn             *OptionSymbol
	QueryDateTimeScopeFrom               *OptionSymbol
	QueryDateTimeScopeTo                 *OptionSymbol
	QueryDistributionNodesSpanSize       *OptionSymbol
	QueryFanoutNodesPercent              *OptionSymbol
	QueryFanoutThreadsPercent            *OptionSymbol
	QueryForceRowLevelSecurity           *OptionSymbol
	QueryLanguage                        *OptionSymbol
	QueryLogQueryParameters              *OptionSymbol
	QueryNow                             *OptionSymbol
	QueryResultsApplyGetSchema           *OptionSymbol
	QueryResultsCacheForceRefresh        *OptionSymbol
	QueryResultsCacheMaxAge              *OptionSymbol
	QueryResultsCachePerShardEnabled     *OptionSymbol
	QueryWeakConsistencySessionId        *OptionSymbol
	RequestAppName                       *OptionSymbol
	RequestBlockRowLevelSecurity         *OptionSymbol
	RequestCalloutDisabled               *OptionSymbol
	RequestCrossClusterToken             *OptionSymbol
	RequestDescription                   *OptionSymbol
	RequestExternalDataDisabled          *OptionSymbol
	RequestExternalTableDisabled         *OptionSymbol
	RequestReadOnly                      *OptionSymbol
	RequestReadOnlyHardline              *OptionSymbol
	RequestRemoteEntitiesDisabled        *OptionSymbol
	RequestSandboxedExecutionDisabled    *OptionSymbol
	RequestUser                          *OptionSymbol
	ResultsProgressiveEnabled            *OptionSymbol
	ResultsV2NewlinesBetweenFrames       *OptionSymbol
	ServerTimeout                        *OptionSymbol
	TakeMaxRecords                       *OptionSymbol
	TruncationMaxRecords                 *OptionSymbol
	TruncationMaxSize                    *OptionSymbol
	V2FragmentPrimaryTables              *OptionSymbol
	ValidatePermissions                  *OptionSymbol
	all                                  []*OptionSymbol
}

func (c *OptionCatalog) All() []*OptionSymbol {
	if c == nil {
		return nil
	}
	return append([]*OptionSymbol{}, c.all...)
}

// PropertyCatalog contains the predefined GlobalState properties.
type PropertyCatalog struct {
	AllowClientParameters *TypedGlobalStateProperty[bool]
	MaxParseTextSize      *TypedGlobalStateProperty[int]
	MaxAnalysisDepth      *TypedGlobalStateProperty[int]
	MaxCachedExpansions   *TypedGlobalStateProperty[int]
	MaxCachedResultTypes  *TypedGlobalStateProperty[int]
}

var (
	// Options is the well-known query-option catalog from Options.cs.
	Options = newOptionCatalog()

	// Properties is the predefined GlobalState property catalog.
	Properties = &PropertyCatalog{
		AllowClientParameters: internal.Properties.AllowClientParameters,
		MaxParseTextSize:      internal.Properties.MaxParseTextSize,
		MaxAnalysisDepth:      internal.Properties.MaxAnalysisDepth,
		MaxCachedExpansions:   internal.Properties.MaxCachedExpansions,
		MaxCachedResultTypes:  internal.Properties.MaxCachedResultTypes,
	}

	// FunctionBodyFactsDefault is the empty immutable function-facts value.
	FunctionBodyFactsDefault = internal.FunctionBodyFactsDefault
)

func newOptionCatalog() *OptionCatalog {
	source := internal.Options
	return &OptionCatalog{
		BestEffort:                           source.BestEffort,
		DebugPython:                          source.DebugPython,
		DeferPartialQueryFailures:            source.DeferPartialQueryFailures,
		DoNotImpersonate:                     source.DoNotImpersonate,
		ErrorReportingPlacement:              source.ErrorReportingPlacement,
		MaterializedViewShuffleQuery:         source.MaterializedViewShuffleQuery,
		MaxEntitiesToUnion:                   source.MaxEntitiesToUnion,
		MaxMemoryConsumptionPerIterator:      source.MaxMemoryConsumptionPerIterator,
		MaxMemoryConsumptionPerQueryPerNode:  source.MaxMemoryConsumptionPerQueryPerNode,
		MaxOutputColumns:                     source.MaxOutputColumns,
		NoRequestTimeout:                     source.NoRequestTimeout,
		NoTruncation:                         source.NoTruncation,
		ProgressiveProgressReportPeriod:      source.ProgressiveProgressReportPeriod,
		ProgressiveQueryMinRowCountPerUpdate: source.ProgressiveQueryMinRowCountPerUpdate,
		PushSelectionThroughAggregation:      source.PushSelectionThroughAggregation,
		QueryBinAutoAt:                       source.QueryBinAutoAt,
		QueryBinAutoSize:                     source.QueryBinAutoSize,
		QueryCursorAfterDefault:              source.QueryCursorAfterDefault,
		QueryCursorBeforeOrAtDefault:         source.QueryCursorBeforeOrAtDefault,
		QueryCursorCurrent:                   source.QueryCursorCurrent,
		QueryCursorDisabled:                  source.QueryCursorDisabled,
		QueryCursorScopedTables:              source.QueryCursorScopedTables,
		QueryDataScope:                       source.QueryDataScope,
		QueryDateTimeScopeColumn:             source.QueryDateTimeScopeColumn,
		QueryDateTimeScopeFrom:               source.QueryDateTimeScopeFrom,
		QueryDateTimeScopeTo:                 source.QueryDateTimeScopeTo,
		QueryDistributionNodesSpanSize:       source.QueryDistributionNodesSpanSize,
		QueryFanoutNodesPercent:              source.QueryFanoutNodesPercent,
		QueryFanoutThreadsPercent:            source.QueryFanoutThreadsPercent,
		QueryForceRowLevelSecurity:           source.QueryForceRowLevelSecurity,
		QueryLanguage:                        source.QueryLanguage,
		QueryLogQueryParameters:              source.QueryLogQueryParameters,
		QueryNow:                             source.QueryNow,
		QueryResultsApplyGetSchema:           source.QueryResultsApplyGetSchema,
		QueryResultsCacheForceRefresh:        source.QueryResultsCacheForceRefresh,
		QueryResultsCacheMaxAge:              source.QueryResultsCacheMaxAge,
		QueryResultsCachePerShardEnabled:     source.QueryResultsCachePerShardEnabled,
		QueryWeakConsistencySessionId:        source.QueryWeakConsistencySessionId,
		RequestAppName:                       source.RequestAppName,
		RequestBlockRowLevelSecurity:         source.RequestBlockRowLevelSecurity,
		RequestCalloutDisabled:               source.RequestCalloutDisabled,
		RequestCrossClusterToken:             source.RequestCrossClusterToken,
		RequestDescription:                   source.RequestDescription,
		RequestExternalDataDisabled:          source.RequestExternalDataDisabled,
		RequestExternalTableDisabled:         source.RequestExternalTableDisabled,
		RequestReadOnly:                      source.RequestReadOnly,
		RequestReadOnlyHardline:              source.RequestReadOnlyHardline,
		RequestRemoteEntitiesDisabled:        source.RequestRemoteEntitiesDisabled,
		RequestSandboxedExecutionDisabled:    source.RequestSandboxedExecutionDisabled,
		RequestUser:                          source.RequestUser,
		ResultsProgressiveEnabled:            source.ResultsProgressiveEnabled,
		ResultsV2NewlinesBetweenFrames:       source.ResultsV2NewlinesBetweenFrames,
		ServerTimeout:                        source.ServerTimeout,
		TakeMaxRecords:                       source.TakeMaxRecords,
		TruncationMaxRecords:                 source.TruncationMaxRecords,
		TruncationMaxSize:                    source.TruncationMaxSize,
		V2FragmentPrimaryTables:              source.V2FragmentPrimaryTables,
		ValidatePermissions:                  source.ValidatePermissions,
		all:                                  append([]*OptionSymbol{}, source.All...),
	}
}

func NewGlobalStateProperty[T any](name string, defaultValue ...T) *TypedGlobalStateProperty[T] {
	return internal.NewGlobalStateProperty(name, defaultValue...)
}

func GlobalStateGetProperty[T any](globals *GlobalState, property *TypedGlobalStateProperty[T]) T {
	return internal.GlobalStateGetProperty(globals, property)
}

func GlobalStateWithProperty[T any](globals *GlobalState, property *TypedGlobalStateProperty[T], value T) *GlobalState {
	return internal.GlobalStateWithProperty(globals, property, value)
}

func NewKustoCache(globals *GlobalState) *KustoCache {
	return internal.NewKustoCache(globals)
}

func KustoCacheGetOrCreate[T any](cache *KustoCache, creator func() T) T {
	return internal.KustoCacheGetOrCreate(cache, creator)
}

func KustoCacheTryGetValue[T any](cache *KustoCache) (T, bool) {
	return internal.KustoCacheTryGetValue[T](cache)
}
