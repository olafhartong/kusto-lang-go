package kustolang

// Ported from src/Kusto.Language/PlugIns.cs. See third_party/kusto/LICENSE and
// third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// PlugInCatalog contains the public built-in evaluate plug-ins. Compatibility
// variants used by historical parsing profiles remain internal.
type PlugInCatalog struct {
	ActiveUseCounts                   *FunctionSymbol
	ActivityCountsMetrics             *FunctionSymbol
	ActivityEngagement                *FunctionSymbol
	ActivityMetrics                   *FunctionSymbol
	NewActivityMetrics                *FunctionSymbol
	AutoCluster                       *FunctionSymbol
	BagUnpack                         *FunctionSymbol
	Basket                            *FunctionSymbol
	DCountIntersect                   *FunctionSymbol
	DiffPatterns                      *FunctionSymbol
	EstimateRowsCount                 *FunctionSymbol
	ExecuteShowCommand                *FunctionSymbol
	ExecuteQuery                      *FunctionSymbol
	ExternalDatatable                 *FunctionSymbol
	FunnelSequence                    *FunctionSymbol
	FunnelSequenceCompletion          *FunctionSymbol
	HttpRequest                       *FunctionSymbol
	HttpRequestPost                   *FunctionSymbol
	AIEmbedTextDeprecated             *FunctionSymbol
	AIEmbeddings                      *FunctionSymbol
	AIChatCompletion                  *FunctionSymbol
	AIChatCompletionPrompt            *FunctionSymbol
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
	Pivot                             *FunctionSymbol
	Preview                           *FunctionSymbol
	CSharp                            *FunctionSymbol
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
	CosmosdbSqlRequest                *FunctionSymbol
	DaxRequest                        *FunctionSymbol
	GqlRequest                        *FunctionSymbol
	AzureDigitalTwinsQueryRequest     *FunctionSymbol
	all                               []*FunctionSymbol
}

var PlugIns = newPlugInCatalog()

func newPlugInCatalog() *PlugInCatalog {
	source := internal.PlugIns
	return &PlugInCatalog{
		ActiveUseCounts:                   source.ActiveUseCounts,
		ActivityCountsMetrics:             source.ActivityCountsMetrics,
		ActivityEngagement:                source.ActivityEngagement,
		ActivityMetrics:                   source.ActivityMetrics,
		NewActivityMetrics:                source.NewActivityMetrics,
		AutoCluster:                       source.AutoCluster,
		BagUnpack:                         source.BagUnpack,
		Basket:                            source.Basket,
		DCountIntersect:                   source.DCountIntersect,
		DiffPatterns:                      source.DiffPatterns,
		EstimateRowsCount:                 source.EstimateRowsCount,
		ExecuteShowCommand:                source.ExecuteShowCommand,
		ExecuteQuery:                      source.ExecuteQuery,
		ExternalDatatable:                 source.ExternalDatatable,
		FunnelSequence:                    source.FunnelSequence,
		FunnelSequenceCompletion:          source.FunnelSequenceCompletion,
		HttpRequest:                       source.HttpRequest,
		HttpRequestPost:                   source.HttpRequestPost,
		AIEmbedTextDeprecated:             source.AIEmbedTextDeprecated,
		AIEmbeddings:                      source.AIEmbeddings,
		AIChatCompletion:                  source.AIChatCompletion,
		AIChatCompletionPrompt:            source.AIChatCompletionPrompt,
		Identity:                          source.Identity,
		IdentityV3:                        source.IdentityV3,
		InferStorageSchema:                source.InferStorageSchema,
		InferStorageSchemaWithSuggestions: source.InferStorageSchemaWithSuggestions,
		GeoPolygonLookup:                  source.GeoPolygonLookup,
		GeoLineLookup:                     source.GeoLineLookup,
		Ipv4Lookup:                        source.Ipv4Lookup,
		Ipv6Lookup:                        source.Ipv6Lookup,
		SchemaMerge:                       source.SchemaMerge,
		Narrow:                            source.Narrow,
		Pivot:                             source.Pivot,
		Preview:                           source.Preview,
		CSharp:                            source.CSharp,
		Python:                            source.Python,
		R:                                 source.R,
		RollingPercentile:                 source.RollingPercentile,
		RowsNear:                          source.RowsNear,
		SessionCount:                      source.SessionCount,
		SequenceDetect:                    source.SequenceDetect,
		SlidingWindowCounts:               source.SlidingWindowCounts,
		SqlRequest:                        source.SqlRequest,
		MySqlRequest:                      source.MySqlRequest,
		PostgreSqlRequest:                 source.PostgreSqlRequest,
		CosmosdbSqlRequest:                source.CosmosdbSqlRequest,
		DaxRequest:                        source.DaxRequest,
		GqlRequest:                        source.GqlRequest,
		AzureDigitalTwinsQueryRequest:     source.AzureDigitalTwinsQueryRequest,
		all:                               append([]*FunctionSymbol{}, source.KnownCurrent...),
	}
}

func (c *PlugInCatalog) All() []*FunctionSymbol {
	if c == nil {
		return nil
	}
	return append([]*FunctionSymbol{}, c.all...)
}

func (c *PlugInCatalog) GetPlugIn(name string) *FunctionSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugIns.GetPlugIn(name)
}

// AutoClusterColumns returns the pinned autocluster result prefix schema.
func (c *PlugInCatalog) AutoClusterColumns() []*ColumnSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugInAutoClusterColumns()
}

// BasketColumns returns the pinned basket result prefix schema.
func (c *PlugInCatalog) BasketColumns() []*ColumnSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugInBasketColumns()
}

// DiffPatternsColumns returns the pinned diffpatterns result prefix schema.
func (c *PlugInCatalog) DiffPatternsColumns() []*ColumnSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugInDiffPatternsColumns()
}

// HttpRequestColumns returns the shared http_request result schema.
func (c *PlugInCatalog) HttpRequestColumns() []*ColumnSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugInHTTPRequestColumns()
}

// NarrowColumns returns the pinned narrow result schema.
func (c *PlugInCatalog) NarrowColumns() []*ColumnSymbol {
	if c == nil {
		return nil
	}
	return internal.PlugInNarrowColumns()
}
