package kustolang

// Ported from src/Kusto.Language/Aggregates.cs. See third_party/kusto/LICENSE
// and third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// AggregateCatalog contains the built-in aggregation function symbols used by
// summarize binding. All returns a defensive copy in upstream order.
type AggregateCatalog struct {
	Sum          *FunctionSymbol
	SumIf        *FunctionSymbol
	Cnt          *FunctionSymbol
	Count        *FunctionSymbol
	CountIf      *FunctionSymbol
	DCount       *FunctionSymbol
	DCountIf     *FunctionSymbol
	TDigest      *FunctionSymbol
	TDigestMerge *FunctionSymbol
	MergeTDigest *FunctionSymbol
	Hll          *FunctionSymbol
	HllIf        *FunctionSymbol
	HllMerge     *FunctionSymbol
	Min          *FunctionSymbol
	MinIf        *FunctionSymbol
	Max          *FunctionSymbol
	MaxIf        *FunctionSymbol
	Avg          *FunctionSymbol
	AvgIf        *FunctionSymbol

	MakeListDeprecated *FunctionSymbol
	MakeList           *FunctionSymbol
	MakeListIf         *FunctionSymbol
	MakeListWithNulls  *FunctionSymbol
	MakeSetDeprecated  *FunctionSymbol
	MakeSet            *FunctionSymbol
	MakeSetIf          *FunctionSymbol
	Passthrough        *FunctionSymbol
	MakeDictionary     *FunctionSymbol
	MakeBag            *FunctionSymbol
	MakeBagIf          *FunctionSymbol
	BuildSchema        *FunctionSymbol
	BinaryAllOr        *FunctionSymbol
	BinaryAllAnd       *FunctionSymbol
	BinaryAllXor       *FunctionSymbol
	CountDistinct      *FunctionSymbol
	CountDistinctIf    *FunctionSymbol

	Percentile        *FunctionSymbol
	Percentiles       *FunctionSymbol
	PercentilesArray  *FunctionSymbol
	PercentileW       *FunctionSymbol
	PercentilesW      *FunctionSymbol
	PercentilesWArray *FunctionSymbol
	Stdev             *FunctionSymbol
	StdevIf           *FunctionSymbol
	Stdevp            *FunctionSymbol
	Variance          *FunctionSymbol
	VarianceIf        *FunctionSymbol
	Variancep         *FunctionSymbol
	VariancepIf       *FunctionSymbol
	Covariance        *FunctionSymbol
	CovarianceIf      *FunctionSymbol
	Covariancep       *FunctionSymbol
	CovariancepIf     *FunctionSymbol
	Any               *FunctionSymbol
	TakeAny           *FunctionSymbol
	AnyIf             *FunctionSymbol
	TakeAnyIf         *FunctionSymbol
	ArgMin            *FunctionSymbol
	ArgMax            *FunctionSymbol
	ArgMinDeprecated  *FunctionSymbol
	ArgMaxDeprecated  *FunctionSymbol
	all               []*FunctionSymbol
}

var Aggregates = newAggregateCatalog()

func newAggregateCatalog() *AggregateCatalog {
	source := internal.Aggregates
	return &AggregateCatalog{
		Sum:                source.Sum,
		SumIf:              source.SumIf,
		Cnt:                source.Cnt,
		Count:              source.Count,
		CountIf:            source.CountIf,
		DCount:             source.DCount,
		DCountIf:           source.DCountIf,
		TDigest:            source.TDigest,
		TDigestMerge:       source.TDigestMerge,
		MergeTDigest:       source.MergeTDigest,
		Hll:                source.Hll,
		HllIf:              source.HllIf,
		HllMerge:           source.HllMerge,
		Min:                source.Min,
		MinIf:              source.MinIf,
		Max:                source.Max,
		MaxIf:              source.MaxIf,
		Avg:                source.Avg,
		AvgIf:              source.AvgIf,
		MakeListDeprecated: source.MakeListDeprecated,
		MakeList:           source.MakeList,
		MakeListIf:         source.MakeListIf,
		MakeListWithNulls:  source.MakeListWithNulls,
		MakeSetDeprecated:  source.MakeSetDeprecated,
		MakeSet:            source.MakeSet,
		MakeSetIf:          source.MakeSetIf,
		Passthrough:        source.Passthrough,
		MakeDictionary:     source.MakeDictionary,
		MakeBag:            source.MakeBag,
		MakeBagIf:          source.MakeBagIf,
		BuildSchema:        source.BuildSchema,
		BinaryAllOr:        source.BinaryAllOr,
		BinaryAllAnd:       source.BinaryAllAnd,
		BinaryAllXor:       source.BinaryAllXor,
		CountDistinct:      source.CountDistinct,
		CountDistinctIf:    source.CountDistinctIf,
		Percentile:         source.Percentile,
		Percentiles:        source.Percentiles,
		PercentilesArray:   source.PercentilesArray,
		PercentileW:        source.PercentileW,
		PercentilesW:       source.PercentilesW,
		PercentilesWArray:  source.PercentilesWArray,
		Stdev:              source.Stdev,
		StdevIf:            source.StdevIf,
		Stdevp:             source.Stdevp,
		Variance:           source.Variance,
		VarianceIf:         source.VarianceIf,
		Variancep:          source.Variancep,
		VariancepIf:        source.VariancepIf,
		Covariance:         source.Covariance,
		CovarianceIf:       source.CovarianceIf,
		Covariancep:        source.Covariancep,
		CovariancepIf:      source.CovariancepIf,
		Any:                source.Any,
		TakeAny:            source.TakeAny,
		AnyIf:              source.AnyIf,
		TakeAnyIf:          source.TakeAnyIf,
		ArgMin:             source.ArgMin,
		ArgMax:             source.ArgMax,
		ArgMinDeprecated:   source.ArgMinDeprecated,
		ArgMaxDeprecated:   source.ArgMaxDeprecated,
		all:                append([]*FunctionSymbol{}, source.All...),
	}
}

func (c *AggregateCatalog) All() []*FunctionSymbol {
	if c == nil {
		return nil
	}
	return append([]*FunctionSymbol{}, c.all...)
}

// GetAnyResult mirrors the public custom-return helper used by any/take_any.
func (c *AggregateCatalog) GetAnyResult(context CustomReturnTypeContext, unnamedExpressionPrefix string) Symbol {
	if c == nil {
		return nil
	}
	return internal.GetAnyResult(context, unnamedExpressionPrefix)
}
