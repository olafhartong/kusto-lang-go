package kustolang

import (
	"reflect"
	"testing"
)

func TestKustoFactsStringInventoriesMatchUpstream(t *testing.T) {
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{"ParamTypes", KustoFacts.ParamTypes, []string{"bool", "boolean", "date", "datetime", "decimal", "double", "dynamic", "guid", "int", "int64", "int8", "long", "real", "string", "time", "timespan", "uniqueid"}},
		{"ExtendedParamTypes", KustoFacts.ExtendedParamTypes, []string{"bool", "boolean", "date", "datetime", "decimal", "double", "dynamic", "float", "guid", "int", "int16", "int32", "int64", "int8", "long", "real", "decimal", "string", "time", "timespan", "uint", "uint16", "uint32", "uint64", "uint8", "ulong", "uniqueid"}},
		{"ChartProperties", KustoFacts.ChartProperties, []string{"title", "xcolumn", "series", "ycolumns", "anomalycolumns", "kind", "xtitle", "ytitle", "xaxis", "yaxis", "legend", "ysplit", "accumulate", "ymin", "ymax", "xmax", "xmin"}},
		{"UnionWithSourceProperties", KustoFacts.UnionWithSourceProperties, []string{"withsource", "with_source"}},
		{"LimitExamples", KustoFacts.LimitExamples, []string{"10", "100", "1000"}},
		{"TopExamples", KustoFacts.TopExamples, []string{"10", "100", "1000"}},
		{"KnownInternalFunctionNames", KustoFacts.KnownInternalFunctionNames, []string{
			"__blackbox", "__box", "__cast", "__columnifexists", "__const_cast", "__fetch_contextual_scalar_value",
			"__geo_line_validate", "__geo_polygon_validate", "__getobject", "__getobjectex", "__has_ipv4", "__has_ipv4_prefix",
			"__hash_crc32", "__hash_djb2", "__hash_many_crc32", "__hash_xxxh64", "__invoke", "__is_scalar",
			"__lz4_compress_dynamic_array_to_base64_string", "__make_const", "__null", "__null_as", "__object", "__query_parameter",
			"__regex_complexity", "__row_offset", "__rowstore_ref", "__search_wildcard", "__search_wildcard_explicit_cols",
			"__search_wildcard_explicit_cols_v35", "__shard_record_position", "__sql_add", "__sql_divide", "__sql_modulo",
			"__sql_multiply", "__sql_subtract", "__trace_information", "__var", "__warning", "__get_scalar",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("%s = %#v, want %#v", test.name, test.got, test.want)
			}
		})
	}
	if !reflect.DeepEqual(KustoFacts.StorageTypes, KustoFacts.ExtendedParamTypes) {
		t.Fatalf("StorageTypes = %#v, want ExtendedParamTypes", KustoFacts.StorageTypes)
	}
	if KustoFacts.UnionIsFuzzyProperty != "isfuzzy" {
		t.Fatalf("UnionIsFuzzyProperty = %q", KustoFacts.UnionIsFuzzyProperty)
	}
}

func TestKustoFactsAliasedInventoriesMatchUpstream(t *testing.T) {
	for _, test := range []struct {
		name string
		got  []string
		want []string
	}{
		{"DistinctHintStrategies", KustoFacts.DistinctHintStrategies, KustoFacts.HintStrategies},
		{"EvaluateHintDistributions", KustoFacts.EvaluateHintDistributions, KustoFacts.HintDistributions},
		{"EvaluateHintRemotes", KustoFacts.EvaluateHintRemotes, KustoFacts.HintRemotes},
		{"JoinHintStrategies", KustoFacts.JoinHintStrategies, KustoFacts.HintStrategies},
		{"PartitionHintConcurrencies", KustoFacts.PartitionHintConcurrencies, KustoFacts.HintConcurrencies},
		{"PartitionHintSpreads", KustoFacts.PartitionHintSpreads, KustoFacts.HintSpreads},
		{"UnionHintConcurrencies", KustoFacts.UnionHintConcurrencies, KustoFacts.HintConcurrencies},
		{"UnionHintSpreads", KustoFacts.UnionHintSpreads, KustoFacts.HintSpreads},
	} {
		if !reflect.DeepEqual(test.got, test.want) {
			t.Fatalf("%s = %#v, want %#v", test.name, test.got, test.want)
		}
	}

	var visible []string
	for _, chartType := range KustoFacts.ChartTypes {
		if !containsKustoFactString(KustoFacts.HiddenChartTypes, chartType) {
			visible = append(visible, chartType)
		}
	}
	if !reflect.DeepEqual(KustoFacts.VisibleChartTypes, visible) {
		t.Fatalf("VisibleChartTypes = %#v, want %#v", KustoFacts.VisibleChartTypes, visible)
	}
}

func TestKustoFactsSyntaxKindInventoriesMatchUpstream(t *testing.T) {
	var keywords []SyntaxKind
	for _, kind := range SyntaxKindsWithFixedText() {
		if IsKeywordKind(kind) && SyntaxKindCanBeIdentifier(kind) {
			keywords = append(keywords, kind)
		}
	}
	if !reflect.DeepEqual(KustoFacts.KeywordsAsIdentifiers, keywords) {
		t.Fatalf("KeywordsAsIdentifiers = %#v, want %#v", KustoFacts.KeywordsAsIdentifiers, keywords)
	}

	extendedSuffix := []SyntaxKind{
		SyntaxKindAccumulateKeyword, SyntaxKindAsKeyword, SyntaxKindByKeyword, SyntaxKindContainsKeyword,
		SyntaxKindConsumeKeyword, SyntaxKindCountKeyword, SyntaxKindDataTableKeyword, SyntaxKindDistinctKeyword,
		SyntaxKindEarliestKeyword, SyntaxKindExtendKeyword, SyntaxKindExternalDataKeyword, SyntaxKindInlineExternalTableKeyword,
		SyntaxKindDataFormatKeyword, SyntaxKindDateTimePatternKeyword, SyntaxKindFindKeyword, SyntaxKindFilterKeyword,
		SyntaxKindHasKeyword, SyntaxKindInKeyword, SyntaxKindInvokeKeyword, SyntaxKindLatestKeyword,
		SyntaxKindLimitKeyword, SyntaxKindMaterializeKeyword, SyntaxKindMdmKeyword, SyntaxKindOfKeyword,
		SyntaxKindParseKeyword, SyntaxKindPlotlyKeyword, SyntaxKindPrintKeyword, SyntaxKindSampleKeyword,
		SyntaxKindSampleDistinctKeyword, SyntaxKindScanKeyword, SyntaxKindSearchKeyword, SyntaxKindSerializeKeyword,
		SyntaxKindSetKeyword, SyntaxKindSortKeyword, SyntaxKindSqlKeyword, SyntaxKindSummarizeKeyword,
		SyntaxKindTakeKeyword, SyntaxKindTitleKeyword, SyntaxKindToKeyword, SyntaxKindTopKeyword,
		SyntaxKindToScalarKeyword, SyntaxKindToTableKeyword, SyntaxKindTopNestedKeyword, SyntaxKindTopHittersKeyword,
		SyntaxKindVerboseKeyword, SyntaxKindViewersKeyword, SyntaxKindWhereKeyword,
	}
	wantExtended := append(append([]SyntaxKind{}, keywords...), extendedSuffix...)
	if !reflect.DeepEqual(KustoFacts.ExtendedKeywordsAsIdentifiers, wantExtended) {
		t.Fatalf("ExtendedKeywordsAsIdentifiers = %#v, want %#v", KustoFacts.ExtendedKeywordsAsIdentifiers, wantExtended)
	}

	wantSpecial := []SyntaxKind{SyntaxKindKindKeyword, SyntaxKindWithSourceKeyword, SyntaxKindWith_SourceKeyword}
	if !reflect.DeepEqual(KustoFacts.SpecialKeywordsAsIdentifiers, wantSpecial) {
		t.Fatalf("SpecialKeywordsAsIdentifiers = %#v", KustoFacts.SpecialKeywordsAsIdentifiers)
	}

	wantFork := []SyntaxKind{
		SyntaxKindAsOperator, SyntaxKindCountOperator, SyntaxKindDistinctOperator, SyntaxKindExecuteAndCacheOperator,
		SyntaxKindExtendOperator, SyntaxKindFilterOperator, SyntaxKindInvokeOperator, SyntaxKindMvExpandOperator,
		SyntaxKindParseOperator, SyntaxKindParseWhereOperator, SyntaxKindParseKvOperator, SyntaxKindProjectOperator,
		SyntaxKindProjectAwayOperator, SyntaxKindProjectByNamesOperator, SyntaxKindProjectKeepOperator, SyntaxKindProjectRenameOperator,
		SyntaxKindProjectReorderOperator, SyntaxKindReduceByOperator, SyntaxKindSampleDistinctOperator, SyntaxKindSampleOperator,
		SyntaxKindSortOperator, SyntaxKindSummarizeOperator, SyntaxKindTakeOperator, SyntaxKindTopHittersOperator,
		SyntaxKindTopNestedOperator, SyntaxKindTopOperator,
	}
	if !reflect.DeepEqual(KustoFacts.ForkOperatorKinds, wantFork) {
		t.Fatalf("ForkOperatorKinds = %#v, want %#v", KustoFacts.ForkOperatorKinds, wantFork)
	}
	wantPostPipe := []SyntaxKind{
		SyntaxKindAsOperator, SyntaxKindConsumeOperator, SyntaxKindCountOperator, SyntaxKindDistinctOperator,
		SyntaxKindEvaluateOperator, SyntaxKindExecuteAndCacheOperator, SyntaxKindExtendOperator, SyntaxKindFacetOperator,
		SyntaxKindFilterOperator, SyntaxKindFindOperator, SyntaxKindForkOperator, SyntaxKindGetSchemaOperator,
		SyntaxKindGraphMatchOperator, SyntaxKindGraphToTableOperator, SyntaxKindInvokeOperator, SyntaxKindJoinOperator,
		SyntaxKindLookupOperator, SyntaxKindMakeSeriesOperator, SyntaxKindMakeGraphOperator, SyntaxKindMvApplyOperator,
		SyntaxKindMvExpandOperator, SyntaxKindParseOperator, SyntaxKindParseWhereOperator, SyntaxKindParseKvOperator,
		SyntaxKindPartitionOperator, SyntaxKindPartitionByOperator, SyntaxKindProjectAwayOperator, SyntaxKindProjectByNamesOperator,
		SyntaxKindProjectKeepOperator, SyntaxKindProjectOperator, SyntaxKindProjectRenameOperator, SyntaxKindProjectReorderOperator,
		SyntaxKindReduceByOperator, SyntaxKindRenderOperator, SyntaxKindSampleDistinctOperator, SyntaxKindSampleOperator,
		SyntaxKindScanOperator, SyntaxKindSearchOperator, SyntaxKindSerializeOperator, SyntaxKindSummarizeOperator,
		SyntaxKindTakeOperator, SyntaxKindSortOperator, SyntaxKindTopHittersOperator, SyntaxKindTopOperator,
		SyntaxKindTopNestedOperator, SyntaxKindUnionOperator,
	}
	if !reflect.DeepEqual(KustoFacts.PostPipeOperatorKinds, wantPostPipe) {
		t.Fatalf("PostPipeOperatorKinds = %#v, want %#v", KustoFacts.PostPipeOperatorKinds, wantPostPipe)
	}
}

func containsKustoFactString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
