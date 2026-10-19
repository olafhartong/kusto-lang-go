package kustolang

import "sort"

// Ported from src/Kusto.Language/QueryOperatorParameters.cs.

type QueryOperatorParameterValueKind string

const (
	QueryOperatorParameterValueAny               QueryOperatorParameterValueKind = "Any"
	QueryOperatorParameterValueScalarLiteral     QueryOperatorParameterValueKind = "ScalarLiteral"
	QueryOperatorParameterValueIntegerLiteral    QueryOperatorParameterValueKind = "IntegerLiteral"
	QueryOperatorParameterValueNumericLiteral    QueryOperatorParameterValueKind = "NumericLiteral"
	QueryOperatorParameterValueForcedRealLiteral QueryOperatorParameterValueKind = "ForcedRealLiteral"
	QueryOperatorParameterValueStringLiteral     QueryOperatorParameterValueKind = "StringLiteral"
	QueryOperatorParameterValueBoolLiteral       QueryOperatorParameterValueKind = "BoolLiteral"
	QueryOperatorParameterValueSummableLiteral   QueryOperatorParameterValueKind = "SummableLiteral"
	QueryOperatorParameterValueString            QueryOperatorParameterValueKind = "String"
	QueryOperatorParameterValueWord              QueryOperatorParameterValueKind = "Word"
	QueryOperatorParameterValueWordOrNumber      QueryOperatorParameterValueKind = "WordOrNumber"
	QueryOperatorParameterValueNameDeclaration   QueryOperatorParameterValueKind = "NameDeclaration"
	QueryOperatorParameterValueColumn            QueryOperatorParameterValueKind = "Column"
	QueryOperatorParameterValueColumnList        QueryOperatorParameterValueKind = "ColumnList"
)

type QueryOperatorParameter struct {
	BaseSymbol
	valueKind       QueryOperatorParameterValueKind
	isCaseSensitive bool
	values          []string
	isRepeatable    bool
	hasNoEquals     bool
	aliases         []string
	isHidden        bool
}

func NewQueryOperatorParameter(name string, kind QueryOperatorParameterValueKind, options ...QueryOperatorParameterOption) *QueryOperatorParameter {
	parameter := &QueryOperatorParameter{
		BaseSymbol:      NewBaseSymbol(name),
		valueKind:       kind,
		isCaseSensitive: true,
	}
	for _, option := range options {
		option(parameter)
	}
	return parameter
}

type QueryOperatorParameterOption func(*QueryOperatorParameter)

func WithQueryOperatorParameterCaseSensitive(isCaseSensitive bool) QueryOperatorParameterOption {
	return func(p *QueryOperatorParameter) {
		p.isCaseSensitive = isCaseSensitive
	}
}

func WithQueryOperatorParameterValues(values []string) QueryOperatorParameterOption {
	return func(p *QueryOperatorParameter) {
		p.values = append([]string{}, values...)
	}
}

func WithQueryOperatorParameterRepeatable(isRepeatable bool) QueryOperatorParameterOption {
	return func(p *QueryOperatorParameter) {
		p.isRepeatable = isRepeatable
	}
}

func WithQueryOperatorParameterAliases(aliases []string) QueryOperatorParameterOption {
	return func(p *QueryOperatorParameter) {
		p.aliases = append([]string{}, aliases...)
	}
}

func (p *QueryOperatorParameter) ValueKind() QueryOperatorParameterValueKind {
	if p == nil {
		return QueryOperatorParameterValueAny
	}
	return p.valueKind
}

func (p *QueryOperatorParameter) IsCaseSensitive() bool {
	return p != nil && p.isCaseSensitive
}

func (p *QueryOperatorParameter) Values() []string {
	if p == nil {
		return nil
	}
	return append([]string{}, p.values...)
}

func (p *QueryOperatorParameter) IsRepeatable() bool {
	return p != nil && p.isRepeatable
}

func (p *QueryOperatorParameter) HasNoEquals() bool {
	return p != nil && p.hasNoEquals
}

func (p *QueryOperatorParameter) Aliases() []string {
	if p == nil {
		return nil
	}
	return append([]string{}, p.aliases...)
}

func (p *QueryOperatorParameter) SymbolKind() SymbolKind {
	return SymbolKindQueryOperatorParameter
}

func (p *QueryOperatorParameter) IsHidden() bool {
	return p != nil && p.isHidden
}

func (p *QueryOperatorParameter) WithIsHidden(isHidden bool) *QueryOperatorParameter {
	if p == nil {
		return nil
	}
	if p.isHidden == isHidden {
		return p
	}
	clone := p.clone()
	clone.isHidden = isHidden
	return clone
}

func (p *QueryOperatorParameter) Hide() *QueryOperatorParameter {
	return p.WithIsHidden(true)
}

func (p *QueryOperatorParameter) WithValues(values []string) *QueryOperatorParameter {
	if p == nil {
		return nil
	}
	clone := p.clone()
	clone.values = append([]string{}, values...)
	return clone
}

func (p *QueryOperatorParameter) WithIsCaseSensitive(isCaseSensitive bool) *QueryOperatorParameter {
	if p == nil {
		return nil
	}
	if p.isCaseSensitive == isCaseSensitive {
		return p
	}
	clone := p.clone()
	clone.isCaseSensitive = isCaseSensitive
	return clone
}

func (p *QueryOperatorParameter) WithHasNoEquals(hasNoEquals bool) *QueryOperatorParameter {
	if p == nil {
		return nil
	}
	if p.hasNoEquals == hasNoEquals {
		return p
	}
	clone := p.clone()
	clone.hasNoEquals = hasNoEquals
	return clone
}

func (p *QueryOperatorParameter) clone() *QueryOperatorParameter {
	clone := *p
	clone.values = append([]string{}, p.values...)
	clone.aliases = append([]string{}, p.aliases...)
	return &clone
}

type queryOperatorParameterSet struct {
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

	AllParameters                    []*QueryOperatorParameter
	AsParameters                     []*QueryOperatorParameter
	ConsumeParameters                []*QueryOperatorParameter
	DataTableParameters              []*QueryOperatorParameter
	DistinctParameters               []*QueryOperatorParameter
	EvaluateParameters               []*QueryOperatorParameter
	ExternalDataWithClauseProperties []*QueryOperatorParameter
	InlineExternalTableProperties    []*QueryOperatorParameter
	FilterParameters                 []*QueryOperatorParameter
	FindParameters                   []*QueryOperatorParameter
	GraphMatchParameters             []*QueryOperatorParameter
	GraphShortestPathsParameters     []*QueryOperatorParameter
	RestrictStatementParameters      []*QueryOperatorParameter
	JoinParameters                   []*QueryOperatorParameter
	GraphMarkComponentsParameters    []*QueryOperatorParameter
	LookupParameters                 []*QueryOperatorParameter
	MakeSeriesParameters             []*QueryOperatorParameter
	MvApplyParameters                []*QueryOperatorParameter
	MvExpandParameters               []*QueryOperatorParameter
	ParseParameters                  []*QueryOperatorParameter
	ParseKvWithProperties            []*QueryOperatorParameter
	PartitionByParameters            []*QueryOperatorParameter
	PartitionParameters              []*QueryOperatorParameter
	ReduceParameters                 []*QueryOperatorParameter
	ReduceLeadingParameters          []*QueryOperatorParameter
	ReduceWithParameters             []*QueryOperatorParameter
	GraphToTableEdgesParameters      []*QueryOperatorParameter
	GraphToTableNodesParameters      []*QueryOperatorParameter
	GraphMakeParameters              []*QueryOperatorParameter
	RenderParameters                 []*QueryOperatorParameter
	RenderWithProperties             []*QueryOperatorParameter
	SampleParameters                 []*QueryOperatorParameter
	SampleDistinctParameters         []*QueryOperatorParameter
	ScanParameters                   []*QueryOperatorParameter
	SearchParameters                 []*QueryOperatorParameter
	SerializedParameters             []*QueryOperatorParameter
	SortParameters                   []*QueryOperatorParameter
	SummarizeParameters              []*QueryOperatorParameter
	TakeParameters                   []*QueryOperatorParameter
	TopParameters                    []*QueryOperatorParameter
	UnionParameters                  []*QueryOperatorParameter
	MacroExpandParameters            []*QueryOperatorParameter
}

var QueryOperatorParameters = newQueryOperatorParameterSet()

func newQueryOperatorParameterSet() *queryOperatorParameterSet {
	q := &queryOperatorParameterSet{}
	q.BagExpansion = qop("bagexpansion", QueryOperatorParameterValueWord, values(KustoFacts.MvExpandKinds)).Hide()
	q.BinLegacy = qop("bin_legacy", QueryOperatorParameterValueAny)
	q.CrossCluster = qop("__crossCluster", QueryOperatorParameterValueAny)
	q.CrossDB = qop("__crossDB", QueryOperatorParameterValueAny)
	q.DecodeBlocks = qop("decodeblocks", QueryOperatorParameterValueBoolLiteral, repeatable(false))
	q.ExpandOutput = qop("expandoutput", QueryOperatorParameterValueAny)
	q.Flags = qop("flags", QueryOperatorParameterValueWord)
	q.HintDotConcurrency = qop("hint.concurrency", QueryOperatorParameterValueWordOrNumber, values(KustoFacts.HintConcurrencies), caseSensitive(false))
	q.HintDotDistribution = qop("hint.distribution", QueryOperatorParameterValueWord, values(KustoFacts.HintDistributions), caseSensitive(false))
	q.HintDotMaterialized = qop("hint.materialized", QueryOperatorParameterValueBoolLiteral)
	q.HintDotNumPartitions = qop("hint.num_partitions", QueryOperatorParameterValueIntegerLiteral, repeatable(false))
	q.HintDotPassFilters = qop("hint.pass_filters", QueryOperatorParameterValueAny)
	q.HintDotPassFiltersColumn = qop("hint.pass_filters_column", QueryOperatorParameterValueColumn)
	q.HintDotProgressiveTop = qop("hint.progressive_top", QueryOperatorParameterValueBoolLiteral)
	q.HintDotRemote = qop("hint.remote", QueryOperatorParameterValueWord, values(KustoFacts.HintRemotes), caseSensitive(false))
	q.HintDotShuffleKey = qop("hint.shufflekey", QueryOperatorParameterValueColumn, repeatable(true))
	q.HintDotSpread = qop("hint.spread", QueryOperatorParameterValueWordOrNumber, values(KustoFacts.HintSpreads), caseSensitive(false))
	q.HintDotStrategy = qop("hint.strategy", QueryOperatorParameterValueWord, values(KustoFacts.HintStrategies), caseSensitive(false))
	q.Id = qop("__id", QueryOperatorParameterValueAny)
	q.IsFuzzy = qop("isfuzzy", QueryOperatorParameterValueBoolLiteral, aliases("__isFuzzy"))
	q.BestEffort = qop("best_effort", QueryOperatorParameterValueBoolLiteral)
	q.Kind = qop("kind", QueryOperatorParameterValueWord)
	q.Output = qop("output", QueryOperatorParameterValueWord)
	q.WithComponentId = qop("with_component_id", QueryOperatorParameterValueNameDeclaration)
	q.NoWithSource = qop("__noWithSource", QueryOperatorParameterValueAny)
	q.PackedColumn = qop("__packedColumn", QueryOperatorParameterValueColumn)
	q.SourceColumnIndex = qop("__sourceColumnIndex", QueryOperatorParameterValueAny)
	q.WithMatchId = qop("with_match_id", QueryOperatorParameterValueNameDeclaration)
	q.WithItemIndex = qop("with_itemindex", QueryOperatorParameterValueNameDeclaration)
	q.WithSource = qop("withsource", QueryOperatorParameterValueNameDeclaration, aliases("with_source"))
	q.WithStepName = qop("with_step_name", QueryOperatorParameterValueNameDeclaration)
	q.WithSourceId = qop("with_source_id", QueryOperatorParameterValueNameDeclaration)
	q.WithTargetId = qop("with_target_id", QueryOperatorParameterValueNameDeclaration)
	q.WithNodeId = qop("with_node_id", QueryOperatorParameterValueNameDeclaration)
	q.ForceRemote = qop("force_remote", QueryOperatorParameterValueBoolLiteral)
	q.Cycles = qop("cycles", QueryOperatorParameterValueWord, values(KustoFacts.CyclesKinds))
	q.ShortestPathsOutputs = qop("output", QueryOperatorParameterValueWord, values(KustoFacts.ShortestPathsOutputs))
	q.ManagedIdentityAuthEnabled = qop("AllowManagedIdentityAuthentication", QueryOperatorParameterValueBoolLiteral, repeatable(false))

	q.AllParameters = []*QueryOperatorParameter{
		q.BagExpansion.Hide(),
		q.BinLegacy.Hide(),
		q.CrossCluster.Hide(),
		q.CrossDB.Hide(),
		q.DecodeBlocks.Hide(),
		q.ExpandOutput.Hide(),
		q.HintDotConcurrency.Hide(),
		q.HintDotDistribution.Hide(),
		q.HintDotMaterialized.Hide(),
		q.HintDotNumPartitions.Hide(),
		q.HintDotPassFilters.Hide(),
		q.HintDotPassFiltersColumn.Hide(),
		q.HintDotProgressiveTop.Hide(),
		q.HintDotRemote.Hide(),
		q.HintDotShuffleKey.Hide(),
		q.HintDotSpread.Hide(),
		q.HintDotStrategy.Hide(),
		q.Id.Hide(),
		q.IsFuzzy.Hide(),
		q.Kind.Hide(),
		q.NoWithSource.Hide(),
		q.PackedColumn.Hide(),
		q.SourceColumnIndex.Hide(),
		q.Cycles.Hide(),
		q.WithMatchId.Hide(),
		q.WithItemIndex.Hide(),
		q.WithSource.Hide(),
		q.WithStepName.Hide(),
		q.WithSourceId.Hide(),
		q.WithTargetId.Hide(),
		q.WithNodeId.Hide(),
		q.ForceRemote.Hide(),
	}

	q.AsParameters = []*QueryOperatorParameter{q.HintDotMaterialized}
	q.ConsumeParameters = []*QueryOperatorParameter{q.DecodeBlocks}
	q.DataTableParameters = nil
	q.DistinctParameters = []*QueryOperatorParameter{
		q.HintDotShuffleKey,
		q.HintDotStrategy.WithValues(KustoFacts.DistinctHintStrategies),
		q.HintDotNumPartitions,
	}
	q.EvaluateParameters = []*QueryOperatorParameter{
		q.HintDotDistribution.WithValues(KustoFacts.EvaluateHintDistributions),
		q.HintDotRemote.WithValues(KustoFacts.EvaluateHintRemotes),
	}
	q.ExternalDataWithClauseProperties = nil
	q.InlineExternalTableProperties = nil
	q.FilterParameters = nil
	q.FindParameters = []*QueryOperatorParameter{q.WithSource}
	q.GraphMatchParameters = []*QueryOperatorParameter{q.Cycles.WithValues(KustoFacts.CyclesKinds)}
	q.GraphShortestPathsParameters = []*QueryOperatorParameter{q.Output.WithValues(KustoFacts.ShortestPathsOutputs)}
	q.RestrictStatementParameters = []*QueryOperatorParameter{q.ManagedIdentityAuthEnabled}
	q.JoinParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.JoinKinds),
		q.HintDotRemote.WithValues(KustoFacts.JoinHintRemotes),
		q.HintDotShuffleKey,
		q.HintDotStrategy.WithValues(KustoFacts.JoinHintStrategies),
		q.HintDotNumPartitions,
	}
	q.GraphMarkComponentsParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.GraphMarkComponentsKinds),
		q.WithComponentId,
	}
	q.LookupParameters = []*QueryOperatorParameter{q.Kind.WithValues(KustoFacts.LookupKinds)}
	q.MakeSeriesParameters = []*QueryOperatorParameter{
		q.HintDotShuffleKey,
		q.Kind.WithValues(KustoFacts.MakeSeriesKinds),
	}
	q.MvApplyParameters = []*QueryOperatorParameter{q.WithItemIndex}
	q.MvExpandParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.MvExpandKinds),
		q.BagExpansion.WithValues(KustoFacts.MvExpandKinds).Hide(),
		q.WithItemIndex,
	}
	q.ParseParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.ParseKinds),
		q.Flags,
	}
	q.ParseKvWithProperties = []*QueryOperatorParameter{
		qop("pair_delimiter", QueryOperatorParameterValueStringLiteral),
		qop("kv_delimiter", QueryOperatorParameterValueStringLiteral),
		qop("regex", QueryOperatorParameterValueStringLiteral),
		qop("quote", QueryOperatorParameterValueStringLiteral, repeatable(true)),
		qop("escape", QueryOperatorParameterValueStringLiteral, repeatable(true)),
		qop("greedy", QueryOperatorParameterValueBoolLiteral),
	}
	q.PartitionByParameters = []*QueryOperatorParameter{
		q.HintDotConcurrency.WithValues(KustoFacts.PartitionHintConcurrencies),
		q.HintDotSpread.WithValues(KustoFacts.PartitionHintSpreads),
		q.HintDotMaterialized,
		q.HintDotShuffleKey,
		q.HintDotStrategy.WithValues(KustoFacts.PartitionHintStrategies),
	}
	q.PartitionParameters = []*QueryOperatorParameter{
		q.HintDotConcurrency.WithValues(KustoFacts.PartitionHintConcurrencies),
		q.HintDotSpread.WithValues(KustoFacts.PartitionHintSpreads),
		q.HintDotMaterialized,
		q.HintDotShuffleKey,
		q.HintDotStrategy.WithValues(KustoFacts.PartitionHintStrategies),
	}
	q.ReduceParameters = []*QueryOperatorParameter{q.Kind.WithValues(KustoFacts.ReduceByKinds)}
	q.ReduceWithParameters = []*QueryOperatorParameter{
		qop("threshold", QueryOperatorParameterValueNumericLiteral),
		qop("characters", QueryOperatorParameterValueStringLiteral),
	}
	q.ReduceLeadingParameters = append(append([]*QueryOperatorParameter{}, q.ReduceParameters...), q.ReduceWithParameters...)
	q.GraphToTableEdgesParameters = []*QueryOperatorParameter{q.WithSourceId, q.WithTargetId}
	q.GraphToTableNodesParameters = []*QueryOperatorParameter{q.WithNodeId}
	q.GraphMakeParameters = []*QueryOperatorParameter{
		q.HintDotStrategy.WithValues(KustoFacts.PartitionedGraphMakeHintStrategies),
	}
	q.RenderKind = qop("kind", QueryOperatorParameterValueWord, values(KustoFacts.ChartKinds))
	q.RenderTitle = qop("title", QueryOperatorParameterValueString)
	q.RenderAccumulate = qop("accumulate", QueryOperatorParameterValueBoolLiteral)
	q.RenderWithDeprecated = qop("with", QueryOperatorParameterValueStringLiteral).WithHasNoEquals(true)
	q.RenderByDeprecated = qop("by", QueryOperatorParameterValueColumnList).WithHasNoEquals(true)
	q.RenderParameters = []*QueryOperatorParameter{
		q.RenderKind.Hide(),
		q.RenderTitle.Hide(),
		q.RenderAccumulate.Hide(),
		q.RenderWithDeprecated.Hide(),
		q.RenderByDeprecated.Hide(),
	}
	q.RenderWithProperties = []*QueryOperatorParameter{
		q.RenderKind,
		q.RenderTitle,
		q.RenderAccumulate,
		qop("xcolumn", QueryOperatorParameterValueColumn),
		qop("ycolumns", QueryOperatorParameterValueColumnList),
		qop("anomalycolumns", QueryOperatorParameterValueColumnList),
		qop("series", QueryOperatorParameterValueColumnList),
		qop("xtitle", QueryOperatorParameterValueString),
		qop("ytitle", QueryOperatorParameterValueString),
		qop("xaxis", QueryOperatorParameterValueWord, values(KustoFacts.ChartAxis)),
		qop("yaxis", QueryOperatorParameterValueWord, values(KustoFacts.ChartAxis)),
		qop("legend", QueryOperatorParameterValueWord, values(KustoFacts.ChartLegends)),
		qop("ysplit", QueryOperatorParameterValueWord, values(KustoFacts.ChartYSplit)),
		qop("ymin", QueryOperatorParameterValueForcedRealLiteral),
		qop("ymax", QueryOperatorParameterValueForcedRealLiteral),
		qop("xmin", QueryOperatorParameterValueScalarLiteral),
		qop("xmax", QueryOperatorParameterValueScalarLiteral),
	}
	q.SampleParameters = nil
	q.SampleDistinctParameters = nil
	q.ScanParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.ScanKinds).Hide(),
		q.WithMatchId,
	}
	q.SearchParameters = []*QueryOperatorParameter{q.Kind.WithValues(KustoFacts.SearchKinds)}
	q.SerializedParameters = nil
	q.SortParameters = []*QueryOperatorParameter{q.HintDotStrategy.WithValues(KustoFacts.SortHintStrategies)}
	q.SummarizeParameters = []*QueryOperatorParameter{
		q.HintDotShuffleKey,
		q.HintDotStrategy.WithValues(KustoFacts.SummarizeHintStrategies),
		q.HintDotNumPartitions,
	}
	q.TakeParameters = nil
	q.TopParameters = []*QueryOperatorParameter{q.HintDotProgressiveTop}
	q.ToScalarKindParameter = q.Kind.WithValues(KustoFacts.ToScalarKinds)
	q.ToTableKindParameter = q.Kind.WithValues(KustoFacts.ToTableKinds)
	q.UnionParameters = []*QueryOperatorParameter{
		q.Kind.WithValues(KustoFacts.UnionKinds),
		q.WithSource,
		q.IsFuzzy,
		q.BestEffort,
		q.HintDotConcurrency.WithValues(KustoFacts.UnionHintConcurrencies),
		q.HintDotSpread.WithValues(KustoFacts.UnionHintSpreads),
	}
	q.MacroExpandParameters = []*QueryOperatorParameter{
		q.ForceRemote,
		q.IsFuzzy,
		q.BestEffort,
	}
	q.GetSchemaKind = qop("kind", QueryOperatorParameterValueWord, values([]string{"csl"}))
	return q
}

func KnownQueryOperatorParameterNames() []string {
	names := []string{}
	for _, parameter := range QueryOperatorParameters.AllParameters {
		if parameter == nil {
			continue
		}
		names = append(names, parameter.SymbolName())
		names = append(names, parameter.Aliases()...)
	}
	sort.Strings(names)
	return names
}

func qop(name string, kind QueryOperatorParameterValueKind, options ...QueryOperatorParameterOption) *QueryOperatorParameter {
	return NewQueryOperatorParameter(name, kind, options...)
}

func values(values []string) QueryOperatorParameterOption {
	return WithQueryOperatorParameterValues(values)
}

func caseSensitive(isCaseSensitive bool) QueryOperatorParameterOption {
	return WithQueryOperatorParameterCaseSensitive(isCaseSensitive)
}

func repeatable(isRepeatable bool) QueryOperatorParameterOption {
	return WithQueryOperatorParameterRepeatable(isRepeatable)
}

func aliases(aliases ...string) QueryOperatorParameterOption {
	return WithQueryOperatorParameterAliases(aliases)
}
