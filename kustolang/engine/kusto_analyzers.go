package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoAnalyzers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type KustoAnalyzersCatalog struct {
	AvoidUsingContains                   KustoAnalyzer
	AvoidUsingNullStringComparison       KustoAnalyzer
	AvoidUsingShortStringComparision     KustoAnalyzer
	AvoidUsingToBoolOnNumerics           KustoAnalyzer
	NullAggregation                      KustoAnalyzer
	AvoidUsingFormatDatetimeInPredicate  KustoAnalyzer
	AvoidUsingLegacyPartition            KustoAnalyzer
	AvoidUsingObsoleteFunctions          KustoAnalyzer
	AvoidJoinWithoutKind                 KustoAnalyzer
	StdevTimespanConversion              KustoAnalyzer
	ColumnHasSameNameAsVariable          KustoAnalyzer
	PreferUsingMaterializedViewIntrinsic KustoAnalyzer
	CalledFunctionHasErrors              KustoAnalyzer
	AvoidUsingHasWithIPv4Strings         KustoAnalyzer
	PreferUsingOptimizedAlternative      KustoAnalyzer
	AvoidStrlenWithDynamic               KustoAnalyzer

	all []KustoAnalyzer
}

var KustoAnalyzers = newKustoAnalyzersCatalog()

func newKustoAnalyzersCatalog() *KustoAnalyzersCatalog {
	catalog := &KustoAnalyzersCatalog{
		AvoidUsingContains:                   NewAvoidUsingContainsAnalyzer(),
		AvoidUsingNullStringComparison:       NewAvoidUsingNullStringComparisonAnalyzer(),
		AvoidUsingShortStringComparision:     NewAvoidUsingShortStringComparisionAnalyzer(),
		AvoidUsingToBoolOnNumerics:           NewAvoidUsingToBoolOnNumericsAnalyzer(),
		NullAggregation:                      NewNullAggregationAnalyzer(),
		AvoidUsingFormatDatetimeInPredicate:  NewAvoidUsingFormatDateTimeInPredicateAnalyzer(),
		AvoidUsingLegacyPartition:            NewAvoidUsingLegacyPartitionAnalyzer(),
		AvoidUsingObsoleteFunctions:          NewAvoidUsingObsoleteFunctionsAnalyzer(),
		AvoidJoinWithoutKind:                 NewAvoidJoinWithoutKindAnalyzer(),
		StdevTimespanConversion:              NewStdevTimespanConversionAnalyzer(),
		ColumnHasSameNameAsVariable:          NewColumnHasSameNameAsVariableAnalyzer(),
		PreferUsingMaterializedViewIntrinsic: NewPreferUsingMaterializedViewIntrinsicAnalyzer(),
		CalledFunctionHasErrors:              NewCalledFunctionHasErrorsAnalyzer(),
		AvoidUsingHasWithIPv4Strings:         NewAvoidUsingHasWithIPv4StringsAnalyzer(),
		PreferUsingOptimizedAlternative:      NewPreferUsingOptimizedAlternativeAnalyzer(),
		AvoidStrlenWithDynamic:               NewAvoidStrlenWithDynamicAnalyzer(),
	}
	catalog.all = ListToReadOnly([]KustoAnalyzer{
		catalog.AvoidUsingContains,
		catalog.AvoidUsingNullStringComparison,
		catalog.AvoidUsingToBoolOnNumerics,
		catalog.NullAggregation,
		catalog.AvoidUsingFormatDatetimeInPredicate,
		catalog.AvoidUsingObsoleteFunctions,
		catalog.AvoidJoinWithoutKind,
		catalog.StdevTimespanConversion,
		catalog.AvoidUsingLegacyPartition,
		catalog.ColumnHasSameNameAsVariable,
		catalog.PreferUsingMaterializedViewIntrinsic,
		catalog.CalledFunctionHasErrors,
		catalog.AvoidUsingHasWithIPv4Strings,
		catalog.PreferUsingOptimizedAlternative,
		catalog.AvoidStrlenWithDynamic,
	})
	return catalog
}

func (c *KustoAnalyzersCatalog) All() []KustoAnalyzer {
	if c == nil {
		return EmptyReadOnlyList[KustoAnalyzer]()
	}
	return ListToReadOnly(c.all)
}
