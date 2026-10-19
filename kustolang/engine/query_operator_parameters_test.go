package kustolang

import "testing"

func TestQueryOperatorParameterSymbolAndCloneHelpers(t *testing.T) {
	param := NewQueryOperatorParameter(
		"kind",
		QueryOperatorParameterValueWord,
		values([]string{"inner", "outer"}),
		caseSensitive(false),
		aliases("kind_alias"),
	)
	if param.SymbolKind() != SymbolKindQueryOperatorParameter || param.SymbolTabularity() != TabularityUnknown {
		t.Fatalf("query operator parameter metadata = kind:%s tabularity:%s", param.SymbolKind(), param.SymbolTabularity())
	}
	if param.IsCaseSensitive() || len(param.Values()) != 2 || param.Values()[0] != "inner" || len(param.Aliases()) != 1 {
		t.Fatalf("query operator parameter constructor did not preserve metadata")
	}

	hidden := param.Hide()
	if !hidden.IsHidden() || param.IsHidden() {
		t.Fatalf("Hide should clone hidden flag without mutating original")
	}
	revalued := hidden.WithValues([]string{"leftouter"})
	if len(revalued.Values()) != 1 || revalued.Values()[0] != "leftouter" || len(hidden.Values()) != 2 {
		t.Fatalf("WithValues should clone values without mutating original")
	}
	noEquals := revalued.WithHasNoEquals(true)
	if !noEquals.HasNoEquals() || revalued.HasNoEquals() {
		t.Fatalf("WithHasNoEquals should clone flag without mutating original")
	}
}

func TestQueryOperatorParametersCatalog(t *testing.T) {
	if len(QueryOperatorParameters.AllParameters) != 32 {
		t.Fatalf("AllParameters count = %d, want 32", len(QueryOperatorParameters.AllParameters))
	}
	for _, parameter := range QueryOperatorParameters.AllParameters {
		if parameter == nil || !parameter.IsHidden() {
			t.Fatalf("AllParameters entry should be hidden: %#v", parameter)
		}
	}
	if QueryOperatorParameters.BagExpansion.SymbolName() != "bagexpansion" || !QueryOperatorParameters.BagExpansion.IsHidden() {
		t.Fatalf("BagExpansion metadata not preserved")
	}
	if got := QueryOperatorParameters.IsFuzzy.Aliases(); len(got) != 1 || got[0] != "__isFuzzy" {
		t.Fatalf("isfuzzy aliases = %#v", got)
	}
	if got := QueryOperatorParameters.WithSource.Aliases(); len(got) != 1 || got[0] != "with_source" {
		t.Fatalf("withsource aliases = %#v", got)
	}
	if !QueryOperatorParameters.HintDotShuffleKey.IsRepeatable() {
		t.Fatal("hint.shufflekey should be repeatable")
	}
	if QueryOperatorParameters.HintDotPassFilters.ValueKind() != QueryOperatorParameterValueAny {
		t.Fatalf("hint.pass_filters value kind = %s, want Any", QueryOperatorParameters.HintDotPassFilters.ValueKind())
	}
	if QueryOperatorParameters.HintDotPassFiltersColumn.IsRepeatable() {
		t.Fatal("hint.pass_filters_column should not be repeatable")
	}
}

func TestQueryOperatorParameterSets(t *testing.T) {
	join := QueryOperatorParameters.JoinParameters
	if len(join) != 5 {
		t.Fatalf("JoinParameters count = %d, want 5", len(join))
	}
	if join[0].SymbolName() != "kind" || len(join[0].Values()) != len(KustoFacts.JoinKinds) || join[0].Values()[0] != "inner" {
		t.Fatalf("join kind values = %#v", join[0].Values())
	}
	if join[1].SymbolName() != "hint.remote" || join[1].Values()[len(join[1].Values())-1] != "unresolved" {
		t.Fatalf("join hint.remote values = %#v", join[1].Values())
	}
	if len(join[3].Values()) == 0 || join[3].Values()[0] != "auto" {
		t.Fatalf("join hint.strategy values = %#v", join[3].Values())
	}

	distinct := QueryOperatorParameters.DistinctParameters
	if len(distinct) != 3 || len(distinct[1].Values()) == 0 || distinct[1].Values()[0] != "auto" {
		t.Fatalf("DistinctParameters strategy values = %#v", distinct)
	}

	evaluate := QueryOperatorParameters.EvaluateParameters
	if len(evaluate) != 2 {
		t.Fatalf("EvaluateParameters count = %d, want 2", len(evaluate))
	}
	if evaluate[0].SymbolName() != "hint.distribution" || evaluate[1].SymbolName() != "hint.remote" {
		t.Fatalf("evaluate parameters = %#v", evaluate)
	}

	mvExpand := QueryOperatorParameters.MvExpandParameters
	if len(mvExpand) != 3 || mvExpand[0].Values()[0] != "bag" || !mvExpand[1].IsHidden() {
		t.Fatalf("MvExpandParameters = %#v", mvExpand)
	}

	render := QueryOperatorParameters.RenderWithProperties
	if len(render) != 17 {
		t.Fatalf("RenderWithProperties count = %d, want 17", len(render))
	}
	if render[13].SymbolName() != "ymin" || render[13].ValueKind() != QueryOperatorParameterValueForcedRealLiteral {
		t.Fatalf("render ymin metadata = %#v", render[13])
	}

	union := QueryOperatorParameters.UnionParameters
	if len(union) != 6 || union[0].Values()[1] != "outer" || union[1] != QueryOperatorParameters.WithSource {
		t.Fatalf("UnionParameters = %#v", union)
	}

	if len(QueryOperatorParameters.ParseKvWithProperties) != 6 || !QueryOperatorParameters.ParseKvWithProperties[3].IsRepeatable() {
		t.Fatalf("ParseKvWithProperties repeatability not preserved")
	}
	if !QueryOperatorParameters.RenderWithDeprecated.HasNoEquals() || !QueryOperatorParameters.RenderByDeprecated.HasNoEquals() {
		t.Fatalf("deprecated render parameters should have no equals")
	}
}

func TestKustoFactsForQueryOperatorParameters(t *testing.T) {
	tests := map[string][]string{
		"HintStrategies":          KustoFacts.HintStrategies,
		"JoinKinds":               KustoFacts.JoinKinds,
		"PartitionHintStrategies": KustoFacts.PartitionHintStrategies,
		"ChartKinds":              KustoFacts.ChartKinds,
		"ChartTypes":              KustoFacts.ChartTypes,
		"DataScopeValues":         KustoFacts.DataScopeValues,
	}
	for name, values := range tests {
		if len(values) == 0 {
			t.Fatalf("%s should not be empty", name)
		}
	}
	if KustoFacts.DistinctHintStrategies[0] != KustoFacts.HintStrategies[0] {
		t.Fatalf("DistinctHintStrategies should mirror HintStrategies")
	}
	if KustoFacts.UnionHintSpreads[1] != "local" {
		t.Fatalf("UnionHintSpreads = %#v", KustoFacts.UnionHintSpreads)
	}
}

func TestKnownQueryOperatorParameterNames(t *testing.T) {
	names := KnownQueryOperatorParameterNames()
	if len(names) != 34 {
		t.Fatalf("known query operator parameter names count = %d, want 34", len(names))
	}
	if names[0] != "__crossCluster" {
		t.Fatalf("known names should be sorted, first = %q", names[0])
	}
	seenIsFuzzyAlias := false
	seenWithSourceAlias := false
	seenWithNodeID := false
	seenWithSourceID := false
	seenWithTargetID := false
	for _, name := range names {
		if name == "__isFuzzy" {
			seenIsFuzzyAlias = true
		}
		if name == "with_source" {
			seenWithSourceAlias = true
		}
		if name == "with_node_id" {
			seenWithNodeID = true
		}
		if name == "with_source_id" {
			seenWithSourceID = true
		}
		if name == "with_target_id" {
			seenWithTargetID = true
		}
	}
	if !seenIsFuzzyAlias || !seenWithSourceAlias || !seenWithNodeID || !seenWithSourceID || !seenWithTargetID {
		t.Fatalf("known names missing aliases: %v", names)
	}
}

func TestGraphToTableIdentifiersRemainPracticalCompatibilityGlobals(t *testing.T) {
	for _, name := range []string{"with_source_id", "with_target_id", "with_node_id"} {
		if parameter := findSpecificQueryOperatorParameter(name, QueryOperatorParameters.AllParameters); parameter == nil {
			t.Fatalf("practical compatibility parameter %q is not global", name)
		}
	}
	if findSpecificQueryOperatorParameter("with_source_id", QueryOperatorParameters.GraphToTableEdgesParameters) == nil ||
		findSpecificQueryOperatorParameter("with_target_id", QueryOperatorParameters.GraphToTableEdgesParameters) == nil ||
		findSpecificQueryOperatorParameter("with_node_id", QueryOperatorParameters.GraphToTableNodesParameters) == nil {
		t.Fatal("graph-to-table parameter sets are missing graph identifier parameters")
	}
}

func TestQueryOperatorParameterNameMatchingIsCaseSensitive(t *testing.T) {
	if got := findSpecificQueryOperatorParameter("__isFuzzy", QueryOperatorParameters.AllParameters); got == nil || got.SymbolName() != "isfuzzy" {
		t.Fatalf("exact __isFuzzy alias match = %#v", got)
	}
	if got := findSpecificQueryOperatorParameter("__isfuzzy", QueryOperatorParameters.AllParameters); got != nil {
		t.Fatalf("lowercase __isfuzzy should not match exact alias, got %#v", got)
	}
	if got := findSpecificQueryOperatorParameter("with_source", QueryOperatorParameters.AllParameters); got == nil || got.SymbolName() != "withsource" {
		t.Fatalf("exact with_source alias match = %#v", got)
	}
	if got := findSpecificQueryOperatorParameter("WITH_SOURCE", QueryOperatorParameters.AllParameters); got != nil {
		t.Fatalf("uppercase WITH_SOURCE should not match exact alias, got %#v", got)
	}
	if got := findSpecificQueryOperatorParameter("AllowManagedIdentityAuthentication", QueryOperatorParameters.RestrictStatementParameters); got != QueryOperatorParameters.ManagedIdentityAuthEnabled {
		t.Fatalf("exact managed identity property match = %#v", got)
	}
	if got := findSpecificQueryOperatorParameter("allowmanagedidentityauthentication", QueryOperatorParameters.RestrictStatementParameters); got != nil {
		t.Fatalf("lowercase managed identity property should not match exact name, got %#v", got)
	}
}
