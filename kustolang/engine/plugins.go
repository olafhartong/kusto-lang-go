package kustolang

import (
	"fmt"
	"strconv"
	"strings"
)

// Ported from src/Kusto.Language/PlugIns.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type plugInSet struct {
	ActiveUseCounts                   *FunctionSymbol
	ActivityCountsMetrics             *FunctionSymbol
	ActivityEngagement                *FunctionSymbol
	ActivityMetrics                   *FunctionSymbol
	AzureDigitalTwinsQueryRequest     *FunctionSymbol
	AutoCluster                       *FunctionSymbol
	BagUnpack                         *FunctionSymbol
	Basket                            *FunctionSymbol
	CosmosdbSqlRequest                *FunctionSymbol
	CSharp                            *FunctionSymbol
	DCountIntersect                   *FunctionSymbol
	DiffPatterns                      *FunctionSymbol
	DiffPatternsText                  *FunctionSymbol
	EstimateRowsCount                 *FunctionSymbol
	ExecuteShowCommand                *FunctionSymbol
	ExecuteQuery                      *FunctionSymbol
	ExternalDatatable                 *FunctionSymbol
	FunnelSequence                    *FunctionSymbol
	FunnelSequenceCompletion          *FunctionSymbol
	HttpRequest                       *FunctionSymbol
	HttpRequestPost                   *FunctionSymbol
	Identity                          *FunctionSymbol
	IdentityV3                        *FunctionSymbol
	InferStorageSchema                *FunctionSymbol
	InferStorageSchemaWithSuggestions *FunctionSymbol
	GeoPolygonLookup                  *FunctionSymbol
	GeoLineLookup                     *FunctionSymbol
	Ipv4Lookup                        *FunctionSymbol
	Ipv6Lookup                        *FunctionSymbol
	SchemaMerge                       *FunctionSymbol
	Narrow                            *FunctionSymbol
	NewActivityMetrics                *FunctionSymbol
	Pivot                             *FunctionSymbol
	Preview                           *FunctionSymbol
	Python                            *FunctionSymbol
	R                                 *FunctionSymbol
	RollingPercentile                 *FunctionSymbol
	RowsNear                          *FunctionSymbol
	SessionCount                      *FunctionSymbol
	SequenceDetect                    *FunctionSymbol
	SlidingWindowCounts               *FunctionSymbol
	SqlRequest                        *FunctionSymbol
	MySqlRequest                      *FunctionSymbol
	PostgreSqlRequest                 *FunctionSymbol
	DaxRequest                        *FunctionSymbol
	GqlRequest                        *FunctionSymbol
	AIEmbedTextDeprecated             *FunctionSymbol
	AIChatCompletion                  *FunctionSymbol
	AIChatCompletionPrompt            *FunctionSymbol
	AIEmbeddings                      *FunctionSymbol

	autoClusterColumns  []*ColumnSymbol
	basketColumns       []*ColumnSymbol
	diffPatternsColumns []*ColumnSymbol
	httpRequestColumns  []*ColumnSymbol
	narrowColumns       []*ColumnSymbol

	All          []*FunctionSymbol
	KnownCurrent []*FunctionSymbol
}

var PlugIns = newPlugInSet()

func (p *plugInSet) GetPlugIn(name string) *FunctionSymbol {
	if p == nil || name == "" {
		return nil
	}
	for _, function := range p.KnownCurrent {
		if function != nil && function.SymbolName() == name {
			return function
		}
	}
	return nil
}

func PlugInAutoClusterColumns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, PlugIns.autoClusterColumns...)
}

func PlugInBasketColumns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, PlugIns.basketColumns...)
}

func PlugInDiffPatternsColumns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, PlugIns.diffPatternsColumns...)
}

func PlugInHTTPRequestColumns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, PlugIns.httpRequestColumns...)
}

func PlugInNarrowColumns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, PlugIns.narrowColumns...)
}

func newPlugInSet() *plugInSet {
	pl := &plugInSet{
		autoClusterColumns:  pluginAutoClusterColumns(),
		basketColumns:       pluginBasketColumns(),
		diffPatternsColumns: pluginDiffPatternsColumns(),
		httpRequestColumns: []*ColumnSymbol{
			NewColumnSymbol("ResponseHeaders", ScalarTypes.DynamicBag.Type()),
			NewColumnSymbol("ResponseBody", ScalarTypes.Dynamic.Type()),
		},
		narrowColumns: []*ColumnSymbol{
			NewColumnSymbol("Row", ScalarTypes.Long.Type()),
			NewColumnSymbol("Column", ScalarTypes.String.Type()),
			NewColumnSymbol("Value", ScalarTypes.String.Type()),
		},
	}

	pl.ActiveUseCounts = pluginCustom("active_users_count",
		pluginTimelineDimensionsReturn("TimelineColumn", "Dimension", "dcount"),
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("LookbackWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Period", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("ActivePeriods", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Step", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))

	pl.ActivityCountsMetrics = pluginCustom("activity_counts_metrics",
		pluginTimelineDimensionsReturn("TimelineColumn", "Dimension", "count", "dcount", "new_dcount", "aggregated_dcount"),
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Step", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))

	pl.ActivityEngagement = NewFunctionSymbol("activity_engagement",
		NewCustomSignature(pluginActivityEngagementReturn, TabularityTabular,
			pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
			pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
			pk("InnerActivityWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("OuterActivityWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant))),
		NewCustomSignature(pluginActivityEngagementReturn, TabularityTabular,
			pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
			pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
			pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("InnerActivityWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("OuterActivityWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat))))

	pl.ActivityMetrics = NewFunctionSymbol("activity_metrics",
		NewCustomSignature(pluginActivityMetricsReturn, TabularityTabular,
			pk("IdColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn)),
			pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
			pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("Step", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
			pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat))),
		NewCustomSignature(pluginActivityMetricsReturn, TabularityTabular,
			pk("IdColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn)),
			pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
			pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1)),
			pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1))))

	namIDColumn := pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn))
	namTimelineColumn := pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn))
	namStart := pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant))
	namEnd := pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant))
	namWindow := pk("Window", ParameterTypeScalar)
	namCohort := pk("Cohort", ParameterTypeScalar, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1))
	namDimension := pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat))
	namLookback := pk("lookback", ParameterTypeTabular, WithParameterOccurrence(0, 1))
	pl.NewActivityMetrics = NewFunctionSymbol("new_activity_metrics",
		NewCustomSignature(pluginNewActivityMetricsReturn, TabularityTabular,
			namIDColumn, namTimelineColumn, namStart, namEnd, namWindow, namCohort, namDimension, namLookback).
			WithCustomLayout(newActivityMetricsLayout(namIDColumn, namTimelineColumn, namStart, namEnd, namWindow, namCohort, namDimension, namLookback)))

	pl.AutoCluster = pluginDeclared("autocluster", tableOf(pl.autoClusterColumns...),
		p("SizeWeight", ScalarTypes.Real, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("WeightColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn), WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("NumSeeds", ParameterTypeInteger, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("CustomWildcard", ParameterTypeScalar, WithParameterOccurrence(0, MaxRepeat)))
	pl.AutoCluster.signatures[0].CustomReturnType = pluginPrefixedRowScopeReturn(pl.autoClusterColumns)
	pl.AutoCluster.signatures[0].ReturnKind = ReturnTypeCustom
	pl.AutoCluster.signatures[0].Tabularity = TabularityTabular

	bagUnpackPrefix := p("column_prefix", ScalarTypes.String, WithArgumentKind(ArgumentKindLiteralNotEmpty), WithParameterOccurrence(0, 1))
	pl.BagUnpack = pluginCustom("bag_unpack", pluginBagUnpackReturn,
		pk("column", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn)),
		bagUnpackPrefix)

	pl.Basket = pluginCustom("basket", pluginPrefixedRowScopeReturn(pl.basketColumns),
		p("Threshold", ScalarTypes.Real, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("WeightColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn), WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("MaxDimensions", ParameterTypeInteger, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("CustomWildcard", ParameterTypeScalar, WithParameterOccurrence(0, MaxRepeat)))

	pl.DCountIntersect = pluginCustom("dcount_intersect", pluginDCountIntersectReturn,
		pk("hll", ParameterTypeDynamicArray, WithParameterOccurrence(2, MaxRepeat)))

	pl.DiffPatterns = pluginCustom("diffpatterns", pluginPrefixedRowScopeReturn(pl.diffPatternsColumns),
		pk("SplitColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn)),
		p("SplitValueA", ScalarTypes.String),
		p("SplitValueB", ScalarTypes.String),
		pk("WeightColumn", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn), WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		p("Threshold", ScalarTypes.Real, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("MaxDimensions", ParameterTypeInteger, WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("CustomWildcard", ParameterTypeScalar, WithParameterOccurrence(0, MaxRepeat)))
	// diffpatterns_text is a Kusto service plug-in that upstream
	// Kusto.Language does not model; kept as an intentional extension.
	pl.DiffPatternsText = pluginDeclared("diffpatterns_text", tableOf(
		NewColumnSymbol("Count_of_True", ScalarTypes.Long.Type()),
		NewColumnSymbol("Count_of_False", ScalarTypes.Long.Type()),
		NewColumnSymbol("Percent_of_True", ScalarTypes.Real.Type()),
		NewColumnSymbol("Percent_of_False", ScalarTypes.Real.Type()),
		NewColumnSymbol("Pattern", ScalarTypes.String.Type())),
		p("TextColumn", ScalarTypes.String, WithArgumentKind(ArgumentKindColumn)),
		p("BooleanCondition", ScalarTypes.Bool),
		pk("MinTokens", ParameterTypeInteger, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1)),
		pk("Threshold", ParameterTypeNumber, WithArgumentKind(ArgumentKindConstant), WithParameterDefaultValueIndicator("~"), WithParameterOccurrence(0, 1)),
		pk("MaxTokens", ParameterTypeInteger, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1)))

	pl.EstimateRowsCount = pluginDeclared("estimate_rows_count", tableOf(NewColumnSymbol("EstimatedRowsCount", ScalarTypes.Long.Type())))
	pl.ExecuteShowCommand = pluginOpen("execute_show_command", p("connection_string", ScalarTypes.String), p("command", ScalarTypes.String))
	pl.ExecuteQuery = pluginOpen("execute_query", p("connection_string", ScalarTypes.String), p("query", ScalarTypes.String))
	pl.ExternalDatatable = NewFunctionSymbol("external_datatable", NewCustomSignature(pluginOpenTableReturn, TabularityTabular))

	pl.FunnelSequence = pluginCustom("funnel_sequence", pluginFunnelSequenceReturn,
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("MaxWindowSizeBetweenSteps", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Step", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("StateColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("Sequence", ParameterTypeDynamicArray, WithArgumentKind(ArgumentKindConstant)))
	pl.FunnelSequenceCompletion = pluginCustom("funnel_sequence_completion", pluginFunnelSequenceCompletionReturn,
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("BinSize", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("StateColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("Sequence", ParameterTypeDynamicArray, WithArgumentKind(ArgumentKindConstant)),
		pk("MaxSequencePeriods", ParameterTypeDynamicArray, WithArgumentKind(ArgumentKindConstant)))

	httpRequestTable := tableOf(pl.httpRequestColumns...)
	pl.HttpRequest = pluginDeclared("http_request", httpRequestTable,
		p("Uri", ScalarTypes.String, WithArgumentKind(ArgumentKindConstant)),
		pk("RequestHeaders", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.HttpRequestPost = pluginDeclared("http_request_post", httpRequestTable,
		p("Uri", ScalarTypes.String, WithArgumentKind(ArgumentKindConstant)),
		pk("RequestHeaders", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		p("Content", ScalarTypes.String, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1)))

	pl.AIEmbedTextDeprecated = pluginCustom("ai_embed_text", pluginAIColumnsReturn("Text", "embedding", ScalarTypes.Dynamic.Type()),
		pk("Text", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumnOrLiteral)),
		p("ConnectionString", ScalarTypes.String),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		p("IncludeErrorMessages", ScalarTypes.Bool, WithParameterOccurrence(0, 1))).Obsolete("ai_embeddings")
	pl.AIEmbeddings = pluginCustom("ai_embeddings", pluginAIColumnsReturn("Text", "embeddings", ScalarTypes.Dynamic.Type()),
		pk("Text", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumnOrLiteral)),
		p("ConnectionString", ScalarTypes.String),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		p("IncludeErrorMessages", ScalarTypes.Bool, WithParameterOccurrence(0, 1)))
	pl.AIChatCompletion = pluginCustom("ai_chat_completion", pluginAIColumnsReturn("Prompt", "completion", ScalarTypes.String.Type()),
		pk("Prompt", ParameterTypeDynamicArray, WithArgumentKind(ArgumentKindColumnOrLiteral)),
		p("ConnectionString", ScalarTypes.String),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		p("IncludeErrorMessages", ScalarTypes.Bool, WithParameterOccurrence(0, 1)))
	pl.AIChatCompletionPrompt = pluginCustom("ai_chat_completion_prompt", pluginAIColumnsReturn("Prompt", "completion", ScalarTypes.String.Type()),
		p("Prompt", ScalarTypes.String, WithArgumentKind(ArgumentKindColumnOrLiteral)),
		p("ConnectionString", ScalarTypes.String),
		pk("Options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)),
		p("IncludeErrorMessages", ScalarTypes.Bool, WithParameterOccurrence(0, 1)))

	pl.Identity = NewFunctionSymbol("identity", NewCustomSignature(pluginRowScopeReturn, TabularityTabular))
	pl.IdentityV3 = pluginCustom("identity_v3", pluginRowScopeReturn,
		p("mode", ScalarTypes.String, WithArgumentKind(ArgumentKindConstant)),
		p("exceptionText", ScalarTypes.String, WithArgumentKind(ArgumentKindConstant)))
	pl.InferStorageSchema = pluginDeclared("infer_storage_schema", tableOf(NewColumnSymbol("CslSchema", ScalarTypes.String.Type())), pk("Options", ParameterTypeDynamicBag))
	pl.InferStorageSchemaWithSuggestions = pluginDeclared("infer_storage_schema_with_suggestions", tableOf(NewColumnSymbol("SuggestedTableSchema", ScalarTypes.String.Type())), pk("Options", ParameterTypeDynamicBag))

	geoLookupTable := pk("LookupTable", ParameterTypeTabular)
	geoLookupPolygonKey := pk("LookupPolygonKey", ParameterTypeDynamicBag, WithArgumentKind(ArgumentKindColumnParameter0))
	geoLookupLineKey := pk("LookupLineKey", ParameterTypeDynamicBag, WithArgumentKind(ArgumentKindColumnParameter0))
	geoLookupSourceLongitude := pk("SourceLongitude", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn))
	geoLookupSourceLatitude := pk("SourceLatitude", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn))
	geoLookupPolygonRadius := p("radius", ScalarTypes.Real, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	geoLookupLineRadius := p("radius", ScalarTypes.Real, WithArgumentKind(ArgumentKindLiteral))
	geoLookupReturnUnmatched := p("return_unmatched", ScalarTypes.Bool, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	geoLookupAreaRadius := p("lookup_area_radius", ScalarTypes.Real, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	geoLookupReturnKey := p("return_lookup_key", ScalarTypes.Bool, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	pl.GeoPolygonLookup = NewFunctionSymbol("geo_polygon_lookup",
		NewCustomSignature(pluginGeoLookupReturn(geoLookupPolygonKey), TabularityTabular,
			geoLookupTable, geoLookupPolygonKey, geoLookupSourceLongitude, geoLookupSourceLatitude,
			geoLookupPolygonRadius, geoLookupReturnUnmatched, geoLookupAreaRadius, geoLookupReturnKey).
			WithCustomLayout(geoPolygonLookupLayout(geoLookupTable, geoLookupPolygonKey, geoLookupSourceLongitude, geoLookupSourceLatitude, geoLookupPolygonRadius, geoLookupReturnUnmatched, geoLookupAreaRadius, geoLookupReturnKey)))
	pl.GeoLineLookup = NewFunctionSymbol("geo_line_lookup",
		NewCustomSignature(pluginGeoLookupReturn(geoLookupLineKey), TabularityTabular,
			geoLookupTable, geoLookupLineKey, geoLookupSourceLongitude, geoLookupSourceLatitude,
			geoLookupLineRadius, geoLookupReturnUnmatched, geoLookupAreaRadius, geoLookupReturnKey).
			WithCustomLayout(geoLineLookupLayout(geoLookupTable, geoLookupLineKey, geoLookupSourceLongitude, geoLookupSourceLatitude, geoLookupLineRadius, geoLookupReturnUnmatched, geoLookupAreaRadius, geoLookupReturnKey)))

	ipv4LookupTable := pk("LookupTable", ParameterTypeTabular)
	ipv4SourceKey := pk("SourceIPv4Key", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn))
	ipv4LookupKey := pk("IPv4LookupKey", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumnParameter0))
	ipv4ExtraKey := pk("ExtraKey", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumnParameter0Common), WithParameterOccurrence(0, MaxRepeat))
	ipv4ReturnUnmatched := p("return_unmatched", ScalarTypes.Bool, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	pl.Ipv4Lookup = NewFunctionSymbol("ipv4_lookup",
		NewCustomSignature(pluginIPv4LookupReturn, TabularityTabular,
			ipv4LookupTable, ipv4SourceKey, ipv4LookupKey, ipv4ExtraKey, ipv4ReturnUnmatched).
			WithCustomLayout(ipv4LookupLayout(ipv4LookupTable, ipv4SourceKey, ipv4LookupKey, ipv4ExtraKey, ipv4ReturnUnmatched)))

	ipv6LookupTable := pk("LookupTable", ParameterTypeTabular)
	ipv6SourceKey := pk("SourceIPv6Key", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn))
	ipv6LookupKey := pk("IPv6LookupKey", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumnParameter0))
	ipv6ReturnUnmatched := p("return_unmatched", ScalarTypes.Bool, WithArgumentKind(ArgumentKindLiteral), WithParameterOccurrence(0, 1))
	pl.Ipv6Lookup = NewFunctionSymbol("ipv6_lookup",
		NewCustomSignature(pluginLookupAppendAllReturn, TabularityTabular, ipv6LookupTable, ipv6SourceKey, ipv6LookupKey, ipv6ReturnUnmatched).
			WithCustomLayout(func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
				return fixedCustomParameters(maxInt(len(arguments), len(argumentTypes)), ipv6LookupTable, ipv6SourceKey, ipv6LookupKey, ipv6ReturnUnmatched)
			}))

	pl.SchemaMerge = pluginDeclared("schema_merge", tableOf(
		NewColumnSymbol("ColumnName", ScalarTypes.String.Type()),
		NewColumnSymbol("ColumnOrdinal", ScalarTypes.Int.Type()),
		NewColumnSymbol("DataType", ScalarTypes.String.Type()),
		NewColumnSymbol("ColumnType", ScalarTypes.String.Type())),
		p("PreserveOrder", ScalarTypes.Bool, WithParameterOccurrence(0, 1)))

	pl.Narrow = pluginDeclared("narrow", tableOf(pl.narrowColumns...))
	pl.Pivot = pluginCustom("pivot", pluginPivotReturn,
		pk("pivotColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("aggregateFunction", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate), WithParameterOccurrence(0, 1)),
		pk("columnName", ParameterTypeScalar, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))
	pl.Preview = pluginCustom("preview", func(context CustomReturnTypeContext) Symbol {
		return NewGroupSymbol(rowScopeTable(context), tableOf(NewColumnSymbol("Count", ScalarTypes.Long.Type())))
	}, pk("NumberOfRows", ParameterTypeInteger))
	pl.CSharp = pluginCustom("csharp", pluginOutputSchemaReturn, p("OutputSchema", ScalarTypes.TypeLit), p("Script", ScalarTypes.String), pk("Arguments", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.Python = pluginCustom("python", pluginOutputSchemaReturn, p("OutputSchema", ScalarTypes.TypeLit), p("Script", ScalarTypes.String), pk("Arguments", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)), pk("Artifacts", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.R = pluginCustom("r", pluginOutputSchemaReturn, p("OutputSchema", ScalarTypes.TypeLit), p("Script", ScalarTypes.String), pk("Arguments", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))

	pl.RollingPercentile = pluginCustom("rolling_percentile", pluginRollingPercentileReturn,
		pk("ValueColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Percentile", ParameterTypeNumber, WithArgumentKind(ArgumentKindConstant)),
		pk("IndexColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("BinSize", ParameterTypeSummable),
		pk("BinsPerWindow", ParameterTypeInteger, WithArgumentKind(ArgumentKindConstant)),
		pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))
	pl.RowsNear = pluginCustom("rows_near", pluginRowScopeReturn,
		p("Condition", ScalarTypes.Bool),
		pk("NumRows", ParameterTypeInteger, WithArgumentKind(ArgumentKindConstant)),
		pk("NumRowsAfter", ParameterTypeInteger, WithArgumentKind(ArgumentKindConstant), WithParameterOccurrence(0, 1)))
	pl.SessionCount = pluginCustom("session_count", pluginTimelineDimensionsReturn("TimelineColumn", "Dimension", "count_sessions"),
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Bin", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("LookBackWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))

	sdTimeline := pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn))
	sdMaxStep := pk("MaxSeqeunceStepWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant))
	sdMaxSpan := pk("MaxSequenceSpan", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant))
	sdExpr := p("Expr", ScalarTypes.Bool, WithParameterOccurrence(1, MaxRepeat))
	sdDimension := pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat))
	pl.SequenceDetect = NewFunctionSymbol("sequence_detect",
		NewCustomSignature(pluginSequenceDetectReturn, TabularityTabular, sdTimeline, sdMaxStep, sdMaxSpan, sdExpr, sdDimension).
			WithCustomLayout(sequenceDetectLayout(sdTimeline, sdMaxStep, sdMaxSpan, sdExpr, sdDimension)))

	pl.SlidingWindowCounts = pluginCustom("sliding_window_counts", pluginTimelineDimensionsReturn("TimelineColumn", "Dimension", "Count", "Dcount"),
		pk("IdColumn", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn)),
		pk("TimelineColumn", ParameterTypeSummable, WithArgumentKind(ArgumentKindColumn)),
		pk("Start", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("End", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("LookBackWindow", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Step", ParameterTypeSummable, WithArgumentKind(ArgumentKindConstant)),
		pk("Dimension", ParameterTypeNotDynamic, WithArgumentKind(ArgumentKindColumn), WithParameterOccurrence(0, MaxRepeat)))

	pl.SqlRequest = pluginOpen("sql_request", p("connection_string", ScalarTypes.String), p("sql_query", ScalarTypes.String), pk("sql_parameters", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)), pk("options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.MySqlRequest = pluginOpen("mysql_request", p("connection_string", ScalarTypes.String), p("sql_query", ScalarTypes.String), pk("sql_parameters", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)), pk("options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.PostgreSqlRequest = pluginOpen("postgresql_request", p("connection_string", ScalarTypes.String), p("sql_query", ScalarTypes.String), pk("sql_parameters", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)), pk("options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1))).Hide()
	// Optional parameter bags and options are omitted until supported by the engine.
	pl.DaxRequest = pluginOpen("dax_request", p("connection_string", ScalarTypes.String), p("dax_query", ScalarTypes.String))
	pl.GqlRequest = pluginOpen("gql_request", p("connection_string", ScalarTypes.String), p("gql_query", ScalarTypes.String))
	pl.CosmosdbSqlRequest = pluginOpen("cosmosdb_sql_request", p("connection_string", ScalarTypes.String), p("sql_query", ScalarTypes.String), pk("sql_parameters", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)), pk("options", ParameterTypeDynamicBag, WithParameterOccurrence(0, 1)))
	pl.AzureDigitalTwinsQueryRequest = pluginOpen("azure_digital_twins_query_request", p("endpoint", ScalarTypes.String), p("sql_query", ScalarTypes.String))

	pl.All = []*FunctionSymbol{
		pl.ActiveUseCounts, pl.ActivityCountsMetrics, pl.ActivityEngagement, pl.ActivityMetrics,
		pl.AzureDigitalTwinsQueryRequest, pl.AutoCluster, pl.BagUnpack, pl.Basket,
		pl.CosmosdbSqlRequest, pl.CSharp, pl.DCountIntersect, pl.DiffPatterns, pl.DiffPatternsText,
		pl.EstimateRowsCount, pl.ExecuteShowCommand, pl.ExecuteQuery, pl.ExternalDatatable,
		pl.FunnelSequence, pl.FunnelSequenceCompletion, pl.HttpRequest, pl.HttpRequestPost,
		pl.Identity, pl.IdentityV3, pl.InferStorageSchema, pl.InferStorageSchemaWithSuggestions,
		pl.GeoPolygonLookup, pl.GeoLineLookup, pl.Ipv4Lookup, pl.Ipv6Lookup,
		pl.Narrow, pl.NewActivityMetrics, pl.Pivot, pl.Preview, pl.Python, pl.R,
		pl.RollingPercentile, pl.RowsNear, pl.SessionCount, pl.SequenceDetect,
		pl.SlidingWindowCounts, pl.SqlRequest, pl.MySqlRequest, pl.PostgreSqlRequest,
		pl.DaxRequest, pl.GqlRequest, pl.AIEmbedTextDeprecated, pl.AIChatCompletion, pl.AIChatCompletionPrompt, pl.AIEmbeddings,
	}
	pl.KnownCurrent = []*FunctionSymbol{
		pl.ActiveUseCounts, pl.ActivityCountsMetrics, pl.ActivityEngagement, pl.ActivityMetrics,
		pl.AzureDigitalTwinsQueryRequest, pl.AutoCluster, pl.BagUnpack, pl.Basket,
		pl.CosmosdbSqlRequest, pl.CSharp, pl.DCountIntersect, pl.DiffPatterns,
		pl.EstimateRowsCount, pl.ExecuteShowCommand, pl.ExecuteQuery, pl.ExternalDatatable,
		pl.FunnelSequence, pl.FunnelSequenceCompletion, pl.HttpRequest, pl.HttpRequestPost,
		pl.Identity, pl.IdentityV3, pl.InferStorageSchema, pl.InferStorageSchemaWithSuggestions,
		pl.GeoPolygonLookup, pl.GeoLineLookup, pl.Ipv4Lookup, pl.Ipv6Lookup,
		pl.Narrow, pl.NewActivityMetrics, pl.Pivot, pl.Preview, pl.Python, pl.R,
		pl.RollingPercentile, pl.RowsNear, pl.SessionCount, pl.SequenceDetect,
		pl.SlidingWindowCounts, pl.SqlRequest, pl.MySqlRequest, pl.PostgreSqlRequest,
		pl.DaxRequest, pl.GqlRequest, pl.AIEmbedTextDeprecated, pl.AIChatCompletion, pl.AIChatCompletionPrompt, pl.AIEmbeddings,
	}
	return pl
}

func pluginDeclared(name string, table *TableSymbol, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewSymbolSignature(table, parameters...))
}

func pluginCustom(name string, returnType CustomReturnType, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewCustomSignature(returnType, TabularityTabular, parameters...))
}

func pluginOpen(name string, parameters ...*Parameter) *FunctionSymbol {
	return pluginCustom(name, pluginOpenTableReturn, parameters...)
}

func tableOf(columns ...*ColumnSymbol) *TableSymbol {
	return NewTableSymbol("", columns...)
}

func pluginOpenTableReturn(CustomReturnTypeContext) Symbol {
	return tableOf().WithIsOpen(true)
}

func pluginRowScopeReturn(context CustomReturnTypeContext) Symbol {
	return rowScopeTable(context)
}

func rowScopeTable(context CustomReturnTypeContext) *TableSymbol {
	if context.RowScope != nil {
		return context.RowScope
	}
	return tableOf()
}

func pluginTimelineDimensionsReturn(timelineParameter string, dimensionParameter string, metricNames ...string) CustomReturnType {
	return func(context CustomReturnTypeContext) Symbol {
		columns := []*ColumnSymbol{}
		columns = AddReferencedColumn(columns, context, timelineParameter)
		columns = AddReferencedColumns(columns, context, dimensionParameter)
		for _, name := range metricNames {
			columns = append(columns, NewColumnSymbol(name, ScalarTypes.Long.Type()))
		}
		return tableOf(columns...)
	}
}

func pluginActivityEngagementReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumn(columns, context, "TimelineColumn")
	columns = AddReferencedColumns(columns, context, "Dimension")
	columns = append(columns,
		NewColumnSymbol("dcount_activities_inner", ScalarTypes.Long.Type()),
		NewColumnSymbol("dcount_activities_outer", ScalarTypes.Long.Type()),
		NewColumnSymbol("activity_ratio", ScalarTypes.Real.Type()))
	return tableOf(columns...)
}

func pluginActivityMetricsReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumn(columns, context, "TimelineColumn")
	columns = AddReferencedColumns(columns, context, "Dimension")
	columns = append(columns,
		NewColumnSymbol("dcount_values", ScalarTypes.Long.Type()),
		NewColumnSymbol("dcount_newvalues", ScalarTypes.Long.Type()),
		NewColumnSymbol("retention_rate", ScalarTypes.Real.Type()),
		NewColumnSymbol("churn_rate", ScalarTypes.Real.Type()))
	return tableOf(columns...)
}

func pluginNewActivityMetricsReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	if timeline := context.GetArgument("TimelineColumn"); timeline != nil {
		name := context.GetResultName(timeline)
		typ := argumentType(context, timeline)
		columns = append(columns,
			NewColumnSymbolWithSource(MakeColumnName("from", name), typ, timeline),
			NewColumnSymbolWithSource(MakeColumnName("to", name), typ, timeline))
	}
	columns = AddReferencedColumns(columns, context, "Dimension")
	columns = append(columns,
		NewColumnSymbol("dcount_new_values", ScalarTypes.Long.Type()),
		NewColumnSymbol("dcount_retained_values", ScalarTypes.Long.Type()),
		NewColumnSymbol("dcount_churn_values", ScalarTypes.Long.Type()),
		NewColumnSymbol("retention_rate", ScalarTypes.Real.Type()),
		NewColumnSymbol("churn_rate", ScalarTypes.Real.Type()))
	return tableOf(columns...)
}

func newActivityMetricsLayout(idColumn, timeline, start, end, window, cohort, dimension, lookback *Parameter) ParameterLayoutBuilder {
	return func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
		count := maxInt(len(arguments), len(argumentTypes))
		out := []*Parameter{idColumn, timeline, start, end, window}
		if count > 5 && (isConstantExpression(expressionAt(arguments, 5)) || !isColumnSymbol(symbolAt(argumentTypes, 5))) {
			out = append(out, cohort)
		}
		for len(out) < count {
			index := len(out)
			if symbolAt(argumentTypes, index).SymbolTabularity() == TabularityTabular {
				out = append(out, lookback)
			} else {
				out = append(out, dimension)
			}
		}
		return out
	}
}

func pluginAutoClusterColumns() []*ColumnSymbol {
	return []*ColumnSymbol{
		NewColumnSymbol("SegmentId", ScalarTypes.Long.Type()),
		NewColumnSymbol("Count", ScalarTypes.Long.Type()),
		NewColumnSymbol("Percent", ScalarTypes.Real.Type()),
	}
}

func pluginBasketColumns() []*ColumnSymbol {
	return pluginAutoClusterColumns()
}

func pluginDiffPatternsColumns() []*ColumnSymbol {
	return []*ColumnSymbol{
		NewColumnSymbol("SegmentId", ScalarTypes.Long.Type()),
		NewColumnSymbol("CountA", ScalarTypes.Long.Type()),
		NewColumnSymbol("CountB", ScalarTypes.Long.Type()),
		NewColumnSymbol("PercentA", ScalarTypes.Real.Type()),
		NewColumnSymbol("PercentB", ScalarTypes.Real.Type()),
		NewColumnSymbol("PercentDiffAB", ScalarTypes.Real.Type()),
	}
}

func pluginPrefixedRowScopeReturn(prefixColumns []*ColumnSymbol) CustomReturnType {
	return func(context CustomReturnTypeContext) Symbol {
		columns := append([]*ColumnSymbol{}, prefixColumns...)
		columns = append(columns, rowScopeColumns(context)...)
		return tableOf(columns...)
	}
}

func pluginBagUnpackReturn(context CustomReturnTypeContext) Symbol {
	var excluded Symbol
	if len(context.Arguments) > 0 && len(context.ArgumentSemantics) > 0 {
		excluded = context.ArgumentSemantics[0].ReferencedSymbol()
	}
	excludedName := ""
	if len(context.Arguments) > 0 && context.RowScope != nil {
		if reference, ok := context.Arguments[0].(*NameReference); ok {
			name := NameReferenceSimpleName(reference)
			for _, column := range context.RowScope.Columns() {
				if column != nil && strings.EqualFold(column.SymbolName(), name) {
					excludedName = column.SymbolName()
					if excluded == nil {
						excluded = column
					}
					break
				}
			}
		}
	}
	columns := []*ColumnSymbol{}
	for _, column := range rowScopeColumns(context) {
		if len(context.Arguments) == 0 || (column != excluded && (excludedName == "" || !strings.EqualFold(column.SymbolName(), excludedName))) {
			columns = append(columns, column)
		}
	}
	return NewTableSymbol("", columns...).WithInheritableProperties(context.RowScope).WithIsOpen(true)
}

func pluginDCountIntersectReturn(context CustomReturnTypeContext) Symbol {
	columns := rowScopeColumns(context)
	for i := range context.Arguments {
		columns = append(columns, NewColumnSymbol("s"+strconv.Itoa(i), ScalarTypes.Long.Type()))
	}
	return tableOf(columns...).WithInheritableProperties(context.RowScope)
}

func pluginFunnelSequenceReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumn(columns, context, "TimelineColumn")
	if stateArg := context.GetArgument("StateColumn"); stateArg != nil {
		typ := argumentType(context, stateArg)
		columns = append(columns,
			NewColumnSymbolWithSource("prev", typ, stateArg),
			NewColumnSymbolWithSource("next", typ, stateArg))
	}
	columns = append(columns,
		NewColumnSymbol("dcount", ScalarTypes.Long.Type()),
		NewColumnSymbol("samples", ScalarTypes.DynamicArray.Type()))
	return tableOf(columns...)
}

func pluginFunnelSequenceCompletionReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumn(columns, context, "TimelineColumn")
	if stateArg := context.GetArgument("StateColumn"); stateArg != nil {
		columns = append(columns, NewColumnSymbolWithSource(context.GetResultName(stateArg), ScalarTypes.String.Type(), stateArg))
	}
	columns = append(columns,
		NewColumnSymbol("Period", ScalarTypes.TimeSpan.Type()),
		NewColumnSymbol("dcount", ScalarTypes.Long.Type()))
	return tableOf(columns...)
}

func pluginAIColumnsReturn(argumentName string, suffix string, typ *TypeSymbol) CustomReturnType {
	return func(context CustomReturnTypeContext) Symbol {
		columns := rowScopeColumns(context)
		prefix := context.GetResultName(context.GetArgument(argumentName))
		columns = append(columns, NewColumnSymbol(MakeColumnName(prefix, suffix), typ))
		if pluginBoolConstant(context, "IncludeErrorMessages") {
			columns = append(columns, NewColumnSymbol(MakeColumnName(prefix, suffix, "error"), ScalarTypes.String.Type()))
		}
		return tableOf(columns...)
	}
}

func pluginGeoLookupReturn(lookupKeyParameter *Parameter) CustomReturnType {
	return func(context CustomReturnTypeContext) Symbol {
		lookupTable := tableArgumentForParameter(context, "LookupTable")
		if lookupTable == nil {
			return rowScopeTable(context)
		}
		columns := rowScopeColumns(context)
		lookupColumns := lookupTable.Columns()
		if !pluginGeoLookupShouldReturnLookupKey(context, lookupKeyParameter.Name) {
			if key, ok := context.GetArgumentReferencedSymbol(lookupKeyParameter.Name).(*ColumnSymbol); ok {
				filtered := lookupColumns[:0]
				for _, column := range lookupColumns {
					if !strings.EqualFold(column.SymbolName(), key.SymbolName()) {
						filtered = append(filtered, column)
					}
				}
				lookupColumns = filtered
			}
		}
		columns = append(columns, lookupColumns...)
		return tableOf(CombineColumns(CombineKindUniqueNames, columns)...)
	}
}

func geoPolygonLookupLayout(table, polygonKey, longitude, latitude, radius, unmatched, areaRadius, returnKey *Parameter) ParameterLayoutBuilder {
	return func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
		count := maxInt(len(arguments), len(argumentTypes))
		out := []*Parameter{table, polygonKey, longitude, latitude}
		for i := 4; i < count; i++ {
			switch strings.ToLower(namedArgumentName(expressionAt(arguments, i))) {
			case "radius":
				out = append(out, radius)
			case "return_unmatched":
				out = append(out, unmatched)
			case "lookup_area_radius":
				out = append(out, areaRadius)
			case "return_lookup_key":
				out = append(out, returnKey)
			default:
				out = append(out, parameterByPosition(i-4, radius, unmatched, areaRadius, returnKey))
			}
		}
		return out
	}
}

func geoLineLookupLayout(table, lineKey, longitude, latitude, radius, unmatched, areaRadius, returnKey *Parameter) ParameterLayoutBuilder {
	return func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
		count := maxInt(len(arguments), len(argumentTypes))
		out := []*Parameter{table, lineKey, longitude, latitude, radius}
		for i := 5; i < count; i++ {
			switch strings.ToLower(namedArgumentName(expressionAt(arguments, i))) {
			case "return_unmatched":
				out = append(out, unmatched)
			case "lookup_area_radius":
				out = append(out, areaRadius)
			case "return_lookup_key":
				out = append(out, returnKey)
			default:
				out = append(out, parameterByPosition(i-4, radius, unmatched, areaRadius, returnKey))
			}
		}
		return out
	}
}

func pluginGeoLookupShouldReturnLookupKey(context CustomReturnTypeContext, lookupKeyName string) bool {
	if context.GetArgumentReferencedSymbol(lookupKeyName) == nil {
		return true
	}
	if len(context.Arguments) == 8 && boolLiteralArgument(context, 7) {
		return true
	}
	if namedBoolLiteralArgument(context, "return_lookup_key") {
		return true
	}
	index := firstArgumentIndexForParameter(context, "return_lookup_key")
	return boolLiteralArgument(context, index)
}

func pluginIPv4LookupReturn(context CustomReturnTypeContext) Symbol {
	lookupTable := tableArgumentForParameter(context, "LookupTable")
	if lookupTable == nil {
		return rowScopeTable(context)
	}
	keyNames := map[string]bool{}
	for _, index := range argumentIndexesForParameter(context, "ExtraKey") {
		if column, ok := context.ArgumentSemantics[index].ReferencedSymbol().(*ColumnSymbol); ok {
			keyNames[column.SymbolName()] = true
		}
	}
	columns := rowScopeColumns(context)
	for _, column := range lookupTable.Columns() {
		if !keyNames[column.SymbolName()] {
			columns = append(columns, column)
		}
	}
	return tableOf(CombineColumns(CombineKindUniqueNames, columns)...)
}

func pluginLookupAppendAllReturn(context CustomReturnTypeContext) Symbol {
	lookupTable := tableArgumentForParameter(context, "LookupTable")
	if lookupTable == nil {
		return rowScopeTable(context)
	}
	columns := append(rowScopeColumns(context), lookupTable.Columns()...)
	return tableOf(CombineColumns(CombineKindUniqueNames, columns)...)
}

func ipv4LookupLayout(table, sourceKey, lookupKey, extraKey, returnUnmatched *Parameter) ParameterLayoutBuilder {
	return func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
		count := maxInt(len(arguments), len(argumentTypes))
		out := []*Parameter{table, sourceKey, lookupKey}
		for i := 3; i < count; i++ {
			argument := expressionAt(arguments, i)
			name := namedArgumentName(argument)
			if i == count-1 && ((name != "" && name == returnUnmatched.Name) ||
				(name == "" && boolLiteralExpression(argument))) {
				out = append(out, returnUnmatched)
			} else {
				out = append(out, extraKey)
			}
		}
		return out
	}
}

func boolLiteralExpression(expression Expression) bool {
	literal, ok := expression.(*LiteralExpression)
	return ok && literal.Kind() == SyntaxKindBooleanLiteralExpression
}

func pluginPivotReturn(context CustomReturnTypeContext) Symbol {
	var pivotColumn *ColumnSymbol
	if len(context.Arguments) > 0 {
		pivotColumn, _ = expressionReferencedSymbol(context.Arguments[0]).(*ColumnSymbol)
	}
	var aggregateColumn *ColumnSymbol
	if len(context.Arguments) > 1 {
		if call, ok := context.Arguments[1].(*FunctionCallExpression); ok && call.ArgumentList() != nil && call.ArgumentList().Expressions().Count() > 0 {
			aggregateColumn, _ = expressionReferencedSymbol(call.ArgumentList().Expressions().At(0).Element()).(*ColumnSymbol)
		}
	}
	columns := []*ColumnSymbol{}
	for _, argument := range context.Arguments[2:] {
		if column, ok := expressionReferencedSymbol(argument).(*ColumnSymbol); ok {
			columns = append(columns, column)
		}
	}
	if len(columns) == 0 {
		for _, column := range rowScopeColumns(context) {
			if column != pivotColumn && column != aggregateColumn {
				columns = append(columns, column)
			}
		}
	}
	return tableOf(columns...).WithIsOpen(true)
}

func pluginOutputSchemaReturn(context CustomReturnTypeContext) Symbol {
	if table := tableArgumentAt(context, 0); table != nil {
		return table
	}
	return tableOf().WithIsOpen(true)
}

func pluginRollingPercentileReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumn(columns, context, "IndexColumn")
	columns = AddReferencedColumns(columns, context, "Dimension")
	bins := pluginConstantString(context, "BinsPerWindow", "0")
	percentile := pluginConstantString(context, "Percentile", "0")
	valueName := "value"
	valueType := ScalarTypes.Long.Type()
	if valueColumn, ok := context.GetArgumentReferencedSymbol("ValueColumn").(*ColumnSymbol); ok {
		valueName = valueColumn.SymbolName()
		valueType = valueColumn.Type
	}
	columns = append(columns, NewColumnSymbol("rolling_"+bins+"_percentile_"+valueName+"_"+percentile, valueType))
	return tableOf(columns...)
}

func pluginSequenceDetectReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	columns = AddReferencedColumns(columns, context, "Dimension")
	if timelineArg := context.GetArgument("TimelineColumn"); timelineArg != nil {
		timelineName := context.GetResultName(timelineArg)
		timelineType := argumentType(context, timelineArg)
		for _, arg := range context.GetArguments("Expr") {
			columns = append(columns, NewColumnSymbolWithSource(MakeColumnName(context.GetResultName(arg), timelineName), timelineType, arg))
		}
	}
	columns = append(columns, NewColumnSymbol("Duration", ScalarTypes.TimeSpan.Type()))
	return tableOf(columns...)
}

func sequenceDetectLayout(timeline, maxStep, maxSpan, expr, dimension *Parameter) ParameterLayoutBuilder {
	return func(_ *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter {
		count := maxInt(len(arguments), len(argumentTypes))
		out := []*Parameter{timeline, maxStep, maxSpan, expr}
		for len(out) < count && arraySortArgumentIsBool(expressionAt(arguments, len(out)), symbolAt(argumentTypes, len(out))) {
			out = append(out, expr)
		}
		for len(out) < count {
			out = append(out, dimension)
		}
		return out
	}
}

func tableArgumentForParameter(context CustomReturnTypeContext, parameterName string) *TableSymbol {
	if index := firstArgumentIndexForParameter(context, parameterName); index >= 0 {
		return tableArgumentAt(context, index)
	}
	return nil
}

func tableArgumentAt(context CustomReturnTypeContext, index int) *TableSymbol {
	if table, ok := argumentResultSymbolAt(context, index).(*TableSymbol); ok {
		return table
	}
	if index >= 0 && index < len(context.ArgumentTypes) && context.ArgumentTypes[index] != nil && context.ArgumentTypes[index].SymbolTabularity() == TabularityTabular {
		if table, ok := TableSymbolForType(context.ArgumentTypes[index]); ok {
			return table
		}
		return NewTableSymbol(context.ArgumentTypes[index].SymbolName())
	}
	return nil
}

func pluginBoolConstant(context CustomReturnTypeContext, parameterName string) bool {
	index := firstArgumentIndexForParameter(context, parameterName)
	if index < 0 || index >= len(context.ArgumentConstantValues) {
		return false
	}
	value, ok := context.ArgumentConstantValues[index].(bool)
	return ok && value
}

func pluginConstantString(context CustomReturnTypeContext, parameterName string, fallback string) string {
	index := firstArgumentIndexForParameter(context, parameterName)
	if index >= 0 && index < len(context.ArgumentConstantValues) && context.ArgumentConstantValues[index] != nil {
		return fmt.Sprint(context.ArgumentConstantValues[index])
	}
	return fallback
}

func boolLiteralArgument(context CustomReturnTypeContext, index int) bool {
	if index < 0 || index >= len(context.ArgumentConstantValues) {
		return false
	}
	value, ok := context.ArgumentConstantValues[index].(bool)
	return ok && value
}

func namedBoolLiteralArgument(context CustomReturnTypeContext, parameterName string) bool {
	for _, argument := range context.Arguments {
		named, ok := argument.(*SimpleNamedExpression)
		if !ok || !strings.EqualFold(NameDeclarationSimpleName(named.Name()), parameterName) {
			continue
		}
		value, ok := boolLiteralExpressionValue(named.Expression())
		if ok && value {
			return true
		}
	}
	return false
}

func boolLiteralExpressionValue(expression Expression) (bool, bool) {
	literal, ok := expression.(*LiteralExpression)
	if !ok || literal.Kind() != SyntaxKindBooleanLiteralExpression || literal.Token() == nil {
		return false, false
	}
	value, err := strconv.ParseBool(strings.ToLower(literal.Token().ValueText()))
	if err != nil {
		return false, false
	}
	return value, true
}

func isConstantExpression(argument Expression) bool {
	if argument == nil {
		return false
	}
	switch argument.Kind() {
	case SyntaxKindBooleanLiteralExpression, SyntaxKindIntLiteralExpression, SyntaxKindLongLiteralExpression,
		SyntaxKindRealLiteralExpression, SyntaxKindDecimalLiteralExpression, SyntaxKindDateTimeLiteralExpression,
		SyntaxKindTimespanLiteralExpression, SyntaxKindGuidLiteralExpression, SyntaxKindStringLiteralExpression,
		SyntaxKindNullLiteralExpression, SyntaxKindTokenLiteralExpression:
		return true
	default:
		return false
	}
}

func isBoolExpression(argument Expression) bool {
	return argument != nil && argument.Kind() == SyntaxKindBooleanLiteralExpression
}

func isColumnSymbol(symbol Symbol) bool {
	_, ok := symbol.(*ColumnSymbol)
	return ok
}

func namedArgumentName(argument Expression) string {
	if named, ok := argument.(*SimpleNamedExpression); ok {
		return parameterNameFromDeclaration(named.Name())
	}
	return ""
}

func parameterByPosition(index int, parameters ...*Parameter) *Parameter {
	if index < 0 || index >= len(parameters) {
		return UnknownParameter
	}
	return parameters[index]
}

func fixedCustomParameters(count int, parameters ...*Parameter) []*Parameter {
	out := make([]*Parameter, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, parameterByPosition(i, parameters...))
	}
	return out
}
