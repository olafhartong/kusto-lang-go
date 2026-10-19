package kustolang

import "testing"

func TestKustoAnalyzersCatalogAllOrderAndInstances(t *testing.T) {
	all := KustoAnalyzers.All()
	want := []struct {
		analyzer KustoAnalyzer
		name     string
		code     string
	}{
		{KustoAnalyzers.AvoidUsingContains, "AvoidUsingContainsAnalyzer", KustoAnalyzerCodeAvoidUsingContains},
		{KustoAnalyzers.AvoidUsingNullStringComparison, "AvoidUsingNullStringComparisonAnalyzer", KustoAnalyzerCodeAvoidUsingIsNullStringComparision},
		{KustoAnalyzers.AvoidUsingToBoolOnNumerics, "AvoidUsingToBoolOnNumericsAnalyzer", KustoAnalyzerCodeAvoidUsingToBoolOnNumerics},
		{KustoAnalyzers.NullAggregation, "NullAggregationAnalyzer", KustoAnalyzerCodeNullAggregation},
		{KustoAnalyzers.AvoidUsingFormatDatetimeInPredicate, "AvoidUsingFormatDateTimeInPredicateAnalyzer", KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate},
		{KustoAnalyzers.AvoidUsingObsoleteFunctions, "AvoidUsingObsoleteFunctionsAnalyzer", KustoAnalyzerCodeAvoidUsingObsoleteFunctions},
		{KustoAnalyzers.AvoidJoinWithoutKind, "AvoidJoinWithoutKindAnalyzer", KustoAnalyzerCodeAvoidJoinWithoutKind},
		{KustoAnalyzers.StdevTimespanConversion, "StdevTimespanConversionAnalyzer", KustoAnalyzerCodeStdevTimespanConversion},
		{KustoAnalyzers.AvoidUsingLegacyPartition, "AvoidUsingLegacyPartitionAnalyzer", KustoAnalyzerCodeAvoidUsingLegacyPartition},
		{KustoAnalyzers.ColumnHasSameNameAsVariable, "ColumnHasSameNameAsVariableAnalyzer", KustoAnalyzerCodeColumnHasSameNameAsVariable},
		{KustoAnalyzers.PreferUsingMaterializedViewIntrinsic, "PreferUsingMaterializedViewIntrinsicAnalyzer", KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic},
		{KustoAnalyzers.CalledFunctionHasErrors, "CalledFunctionHasErrorsAnalyzer", KustoAnalyzerCodeCalledFunctionHasErrors},
		{KustoAnalyzers.AvoidUsingHasWithIPv4Strings, "AvoidUsingHasWithIPv4StringsAnalyzer", KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings},
		{KustoAnalyzers.PreferUsingOptimizedAlternative, "PreferUsingOptimizedAlternative", KustoAnalyzerCodePreferUsingOptimizedAlternative},
		{KustoAnalyzers.AvoidStrlenWithDynamic, "AvoidStrlenWithDynamicAnalyzer", KustoAnalyzerCodeAvoidStrlenWithDynamic},
	}
	if len(all) != len(want) {
		t.Fatalf("analyzer count = %d, want %d", len(all), len(want))
	}
	for i, expected := range want {
		if all[i] != expected.analyzer {
			t.Fatalf("analyzer[%d] = %#v, want catalog field %#v", i, all[i], expected.analyzer)
		}
		if all[i].Name() != expected.name {
			t.Fatalf("analyzer[%d] name = %q, want %q", i, all[i].Name(), expected.name)
		}
		diagnostics := all[i].Diagnostics()
		if len(diagnostics) == 0 || diagnostics[0].Code != expected.code {
			t.Fatalf("analyzer[%d] diagnostics = %#v, want first code %s", i, diagnostics, expected.code)
		}
	}
}

func TestKustoAnalyzersAllReturnsCopy(t *testing.T) {
	all := KustoAnalyzers.All()
	if len(all) == 0 {
		t.Fatal("expected analyzers")
	}
	original := all[0]
	all[0] = nil
	if KustoAnalyzers.All()[0] != original {
		t.Fatal("All did not return a defensive copy")
	}
	if got := (*KustoAnalyzersCatalog)(nil).All(); got == nil || len(got) != 0 {
		t.Fatalf("nil catalog All = %#v, want non-nil empty", got)
	}
}
