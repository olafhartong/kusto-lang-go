package kustolang

// Ported from src/Kusto.Language/QueryOperatorParameters.cs. See
// third_party/kusto/LICENSE and third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type QueryOperatorParameterValueKind = internal.QueryOperatorParameterValueKind

const (
	QueryOperatorParameterValueAny               = internal.QueryOperatorParameterValueAny
	QueryOperatorParameterValueScalarLiteral     = internal.QueryOperatorParameterValueScalarLiteral
	QueryOperatorParameterValueIntegerLiteral    = internal.QueryOperatorParameterValueIntegerLiteral
	QueryOperatorParameterValueNumericLiteral    = internal.QueryOperatorParameterValueNumericLiteral
	QueryOperatorParameterValueForcedRealLiteral = internal.QueryOperatorParameterValueForcedRealLiteral
	QueryOperatorParameterValueStringLiteral     = internal.QueryOperatorParameterValueStringLiteral
	QueryOperatorParameterValueBoolLiteral       = internal.QueryOperatorParameterValueBoolLiteral
	QueryOperatorParameterValueSummableLiteral   = internal.QueryOperatorParameterValueSummableLiteral
	QueryOperatorParameterValueString            = internal.QueryOperatorParameterValueString
	QueryOperatorParameterValueWord              = internal.QueryOperatorParameterValueWord
	QueryOperatorParameterValueWordOrNumber      = internal.QueryOperatorParameterValueWordOrNumber
	QueryOperatorParameterValueNameDeclaration   = internal.QueryOperatorParameterValueNameDeclaration
	QueryOperatorParameterValueColumn            = internal.QueryOperatorParameterValueColumn
	QueryOperatorParameterValueColumnList        = internal.QueryOperatorParameterValueColumnList
)

type QueryOperatorParameter = internal.QueryOperatorParameter
type QueryOperatorParameterOption = internal.QueryOperatorParameterOption

func NewQueryOperatorParameter(name string, kind QueryOperatorParameterValueKind, options ...QueryOperatorParameterOption) *QueryOperatorParameter {
	return internal.NewQueryOperatorParameter(name, kind, options...)
}

func WithQueryOperatorParameterCaseSensitive(value bool) QueryOperatorParameterOption {
	return internal.WithQueryOperatorParameterCaseSensitive(value)
}

func WithQueryOperatorParameterValues(values []string) QueryOperatorParameterOption {
	return internal.WithQueryOperatorParameterValues(values)
}

func WithQueryOperatorParameterRepeatable(value bool) QueryOperatorParameterOption {
	return internal.WithQueryOperatorParameterRepeatable(value)
}

func WithQueryOperatorParameterAliases(aliases []string) QueryOperatorParameterOption {
	return internal.WithQueryOperatorParameterAliases(aliases)
}

// QueryOperatorParameterCatalog contains the predefined operator parameters.
// Its list methods return copies so callers cannot mutate parser metadata.
type QueryOperatorParameterCatalog struct {
	BagExpansion               *QueryOperatorParameter
	BinLegacy                  *QueryOperatorParameter
	CrossCluster               *QueryOperatorParameter
	CrossDB                    *QueryOperatorParameter
	DecodeBlocks               *QueryOperatorParameter
	ExpandOutput               *QueryOperatorParameter
	Flags                      *QueryOperatorParameter
	HintDotConcurrency         *QueryOperatorParameter
	HintDotDistribution        *QueryOperatorParameter
	HintDotMaterialized        *QueryOperatorParameter
	HintDotNumPartitions       *QueryOperatorParameter
	HintDotPassFilters         *QueryOperatorParameter
	HintDotPassFiltersColumn   *QueryOperatorParameter
	HintDotProgressiveTop      *QueryOperatorParameter
	HintDotRemote              *QueryOperatorParameter
	HintDotShuffleKey          *QueryOperatorParameter
	HintDotSpread              *QueryOperatorParameter
	HintDotStrategy            *QueryOperatorParameter
	Id                         *QueryOperatorParameter
	IsFuzzy                    *QueryOperatorParameter
	BestEffort                 *QueryOperatorParameter
	Kind                       *QueryOperatorParameter
	Output                     *QueryOperatorParameter
	WithComponentId            *QueryOperatorParameter
	NoWithSource               *QueryOperatorParameter
	PackedColumn               *QueryOperatorParameter
	SourceColumnIndex          *QueryOperatorParameter
	WithMatchId                *QueryOperatorParameter
	WithItemIndex              *QueryOperatorParameter
	WithSource                 *QueryOperatorParameter
	WithStepName               *QueryOperatorParameter
	ForceRemote                *QueryOperatorParameter
	Cycles                     *QueryOperatorParameter
	ShortestPathsOutputs       *QueryOperatorParameter
	ManagedIdentityAuthEnabled *QueryOperatorParameter
	WithSourceId               *QueryOperatorParameter
	WithTargetId               *QueryOperatorParameter
	WithNodeId                 *QueryOperatorParameter
	RenderKind                 *QueryOperatorParameter
	RenderTitle                *QueryOperatorParameter
	RenderAccumulate           *QueryOperatorParameter
	RenderWithDeprecated       *QueryOperatorParameter
	RenderByDeprecated         *QueryOperatorParameter
	ToScalarKindParameter      *QueryOperatorParameter
	ToTableKindParameter       *QueryOperatorParameter
	GetSchemaKind              *QueryOperatorParameter
	lists                      map[string][]*QueryOperatorParameter
}

var QueryOperatorParameters = newQueryOperatorParameterCatalog()

func newQueryOperatorParameterCatalog() *QueryOperatorParameterCatalog {
	source := internal.QueryOperatorParameters
	return &QueryOperatorParameterCatalog{
		BagExpansion:               source.BagExpansion,
		BinLegacy:                  source.BinLegacy,
		CrossCluster:               source.CrossCluster,
		CrossDB:                    source.CrossDB,
		DecodeBlocks:               source.DecodeBlocks,
		ExpandOutput:               source.ExpandOutput,
		Flags:                      source.Flags,
		HintDotConcurrency:         source.HintDotConcurrency,
		HintDotDistribution:        source.HintDotDistribution,
		HintDotMaterialized:        source.HintDotMaterialized,
		HintDotNumPartitions:       source.HintDotNumPartitions,
		HintDotPassFilters:         source.HintDotPassFilters,
		HintDotPassFiltersColumn:   source.HintDotPassFiltersColumn,
		HintDotProgressiveTop:      source.HintDotProgressiveTop,
		HintDotRemote:              source.HintDotRemote,
		HintDotShuffleKey:          source.HintDotShuffleKey,
		HintDotSpread:              source.HintDotSpread,
		HintDotStrategy:            source.HintDotStrategy,
		Id:                         source.Id,
		IsFuzzy:                    source.IsFuzzy,
		BestEffort:                 source.BestEffort,
		Kind:                       source.Kind,
		Output:                     source.Output,
		WithComponentId:            source.WithComponentId,
		NoWithSource:               source.NoWithSource,
		PackedColumn:               source.PackedColumn,
		SourceColumnIndex:          source.SourceColumnIndex,
		WithMatchId:                source.WithMatchId,
		WithItemIndex:              source.WithItemIndex,
		WithSource:                 source.WithSource,
		WithStepName:               source.WithStepName,
		ForceRemote:                source.ForceRemote,
		Cycles:                     source.Cycles,
		ShortestPathsOutputs:       source.ShortestPathsOutputs,
		ManagedIdentityAuthEnabled: source.ManagedIdentityAuthEnabled,
		WithSourceId:               source.WithSourceId,
		WithTargetId:               source.WithTargetId,
		WithNodeId:                 source.WithNodeId,
		RenderKind:                 source.RenderKind,
		RenderTitle:                source.RenderTitle,
		RenderAccumulate:           source.RenderAccumulate,
		RenderWithDeprecated:       source.RenderWithDeprecated,
		RenderByDeprecated:         source.RenderByDeprecated,
		ToScalarKindParameter:      source.ToScalarKindParameter,
		ToTableKindParameter:       source.ToTableKindParameter,
		GetSchemaKind:              source.GetSchemaKind,
		lists: map[string][]*QueryOperatorParameter{
			"AllParameters":                    source.AllParameters,
			"AsParameters":                     source.AsParameters,
			"ConsumeParameters":                source.ConsumeParameters,
			"DataTableParameters":              source.DataTableParameters,
			"DistinctParameters":               source.DistinctParameters,
			"EvaluateParameters":               source.EvaluateParameters,
			"ExternalDataWithClauseProperties": source.ExternalDataWithClauseProperties,
			"InlineExternalTableProperties":    source.InlineExternalTableProperties,
			"FilterParameters":                 source.FilterParameters,
			"FindParameters":                   source.FindParameters,
			"GraphMatchParameters":             source.GraphMatchParameters,
			"GraphShortestPathsParameters":     source.GraphShortestPathsParameters,
			"RestrictStatementParameters":      source.RestrictStatementParameters,
			"JoinParameters":                   source.JoinParameters,
			"GraphMarkComponentsParameters":    source.GraphMarkComponentsParameters,
			"LookupParameters":                 source.LookupParameters,
			"MakeSeriesParameters":             source.MakeSeriesParameters,
			"MvApplyParameters":                source.MvApplyParameters,
			"MvExpandParameters":               source.MvExpandParameters,
			"ParseParameters":                  source.ParseParameters,
			"ParseKvWithProperties":            source.ParseKvWithProperties,
			"PartitionByParameters":            source.PartitionByParameters,
			"PartitionParameters":              source.PartitionParameters,
			"ReduceParameters":                 source.ReduceParameters,
			"ReduceWithParameters":             source.ReduceWithParameters,
			"GraphToTableEdgesParameters":      source.GraphToTableEdgesParameters,
			"GraphToTableNodesParameters":      source.GraphToTableNodesParameters,
			"GraphMakeParameters":              source.GraphMakeParameters,
			"RenderParameters":                 source.RenderParameters,
			"RenderWithProperties":             source.RenderWithProperties,
			"SampleParameters":                 source.SampleParameters,
			"SampleDistinctParameters":         source.SampleDistinctParameters,
			"ScanParameters":                   source.ScanParameters,
			"SearchParameters":                 source.SearchParameters,
			"SerializedParameters":             source.SerializedParameters,
			"SortParameters":                   source.SortParameters,
			"SummarizeParameters":              source.SummarizeParameters,
			"TakeParameters":                   source.TakeParameters,
			"TopParameters":                    source.TopParameters,
			"UnionParameters":                  source.UnionParameters,
			"MacroExpandParameters":            source.MacroExpandParameters,
		},
	}
}

func (c *QueryOperatorParameterCatalog) list(name string) []*QueryOperatorParameter {
	if c == nil {
		return nil
	}
	return append([]*QueryOperatorParameter{}, c.lists[name]...)
}

func (c *QueryOperatorParameterCatalog) AllParameters() []*QueryOperatorParameter {
	return c.list("AllParameters")
}
func (c *QueryOperatorParameterCatalog) AsParameters() []*QueryOperatorParameter {
	return c.list("AsParameters")
}
func (c *QueryOperatorParameterCatalog) ConsumeParameters() []*QueryOperatorParameter {
	return c.list("ConsumeParameters")
}
func (c *QueryOperatorParameterCatalog) DataTableParameters() []*QueryOperatorParameter {
	return c.list("DataTableParameters")
}
func (c *QueryOperatorParameterCatalog) DistinctParameters() []*QueryOperatorParameter {
	return c.list("DistinctParameters")
}
func (c *QueryOperatorParameterCatalog) EvaluateParameters() []*QueryOperatorParameter {
	return c.list("EvaluateParameters")
}
func (c *QueryOperatorParameterCatalog) ExternalDataWithClauseProperties() []*QueryOperatorParameter {
	return c.list("ExternalDataWithClauseProperties")
}
func (c *QueryOperatorParameterCatalog) InlineExternalTableProperties() []*QueryOperatorParameter {
	return c.list("InlineExternalTableProperties")
}
func (c *QueryOperatorParameterCatalog) FilterParameters() []*QueryOperatorParameter {
	return c.list("FilterParameters")
}
func (c *QueryOperatorParameterCatalog) FindParameters() []*QueryOperatorParameter {
	return c.list("FindParameters")
}
func (c *QueryOperatorParameterCatalog) GraphMatchParameters() []*QueryOperatorParameter {
	return c.list("GraphMatchParameters")
}
func (c *QueryOperatorParameterCatalog) GraphShortestPathsParameters() []*QueryOperatorParameter {
	return c.list("GraphShortestPathsParameters")
}
func (c *QueryOperatorParameterCatalog) RestrictStatementParameters() []*QueryOperatorParameter {
	return c.list("RestrictStatementParameters")
}
func (c *QueryOperatorParameterCatalog) JoinParameters() []*QueryOperatorParameter {
	return c.list("JoinParameters")
}
func (c *QueryOperatorParameterCatalog) GraphMarkComponentsParameters() []*QueryOperatorParameter {
	return c.list("GraphMarkComponentsParameters")
}
func (c *QueryOperatorParameterCatalog) LookupParameters() []*QueryOperatorParameter {
	return c.list("LookupParameters")
}
func (c *QueryOperatorParameterCatalog) MakeSeriesParameters() []*QueryOperatorParameter {
	return c.list("MakeSeriesParameters")
}
func (c *QueryOperatorParameterCatalog) MvApplyParameters() []*QueryOperatorParameter {
	return c.list("MvApplyParameters")
}
func (c *QueryOperatorParameterCatalog) MvExpandParameters() []*QueryOperatorParameter {
	return c.list("MvExpandParameters")
}
func (c *QueryOperatorParameterCatalog) ParseParameters() []*QueryOperatorParameter {
	return c.list("ParseParameters")
}
func (c *QueryOperatorParameterCatalog) ParseKvWithProperties() []*QueryOperatorParameter {
	return c.list("ParseKvWithProperties")
}
func (c *QueryOperatorParameterCatalog) PartitionByParameters() []*QueryOperatorParameter {
	return c.list("PartitionByParameters")
}
func (c *QueryOperatorParameterCatalog) PartitionParameters() []*QueryOperatorParameter {
	return c.list("PartitionParameters")
}
func (c *QueryOperatorParameterCatalog) ReduceParameters() []*QueryOperatorParameter {
	return c.list("ReduceParameters")
}
func (c *QueryOperatorParameterCatalog) ReduceWithParameters() []*QueryOperatorParameter {
	return c.list("ReduceWithParameters")
}
func (c *QueryOperatorParameterCatalog) GraphToTableEdgesParameters() []*QueryOperatorParameter {
	return c.list("GraphToTableEdgesParameters")
}
func (c *QueryOperatorParameterCatalog) GraphToTableNodesParameters() []*QueryOperatorParameter {
	return c.list("GraphToTableNodesParameters")
}
func (c *QueryOperatorParameterCatalog) GraphMakeParameters() []*QueryOperatorParameter {
	return c.list("GraphMakeParameters")
}
func (c *QueryOperatorParameterCatalog) RenderParameters() []*QueryOperatorParameter {
	return c.list("RenderParameters")
}
func (c *QueryOperatorParameterCatalog) RenderWithProperties() []*QueryOperatorParameter {
	return c.list("RenderWithProperties")
}
func (c *QueryOperatorParameterCatalog) SampleParameters() []*QueryOperatorParameter {
	return c.list("SampleParameters")
}
func (c *QueryOperatorParameterCatalog) SampleDistinctParameters() []*QueryOperatorParameter {
	return c.list("SampleDistinctParameters")
}
func (c *QueryOperatorParameterCatalog) ScanParameters() []*QueryOperatorParameter {
	return c.list("ScanParameters")
}
func (c *QueryOperatorParameterCatalog) SearchParameters() []*QueryOperatorParameter {
	return c.list("SearchParameters")
}
func (c *QueryOperatorParameterCatalog) SerializedParameters() []*QueryOperatorParameter {
	return c.list("SerializedParameters")
}
func (c *QueryOperatorParameterCatalog) SortParameters() []*QueryOperatorParameter {
	return c.list("SortParameters")
}
func (c *QueryOperatorParameterCatalog) SummarizeParameters() []*QueryOperatorParameter {
	return c.list("SummarizeParameters")
}
func (c *QueryOperatorParameterCatalog) TakeParameters() []*QueryOperatorParameter {
	return c.list("TakeParameters")
}
func (c *QueryOperatorParameterCatalog) TopParameters() []*QueryOperatorParameter {
	return c.list("TopParameters")
}
func (c *QueryOperatorParameterCatalog) UnionParameters() []*QueryOperatorParameter {
	return c.list("UnionParameters")
}
func (c *QueryOperatorParameterCatalog) MacroExpandParameters() []*QueryOperatorParameter {
	return c.list("MacroExpandParameters")
}
