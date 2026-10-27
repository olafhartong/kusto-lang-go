package kustolang

import "unicode/utf16"

// Ported from the parser facts in src/Kusto.Language/Parser/KustoFacts.cs that
// are referenced by QueryOperatorParameters.cs and built-in function metadata.

var KustoFacts = struct {
	AgoExamples                                 []string
	ParamTypes                                  []string
	ExtendedParamTypes                          []string
	StorageTypes                                []string
	ChartAxis                                   []string
	ChartKinds                                  []string
	ChartLegends                                []string
	ChartProperties                             []string
	ChartTypes                                  []string
	HiddenChartTypes                            []string
	VisibleChartTypes                           []string
	ChartYSplit                                 []string
	CyclesKinds                                 []string
	DataScopeValues                             []string
	DateDiffParts                               []string
	DateTimeParts                               []string
	Directives                                  []string
	DistinctHintStrategies                      []string
	EvaluateHintDistributions                   []string
	EvaluateHintRemotes                         []string
	GraphMarkComponentsKinds                    []string
	HintConcurrencies                           []string
	HintDistributions                           []string
	HintRemotes                                 []string
	HintSpreads                                 []string
	HintStrategies                              []string
	InlineExternalTableDataFormats              []string
	InlineExternalTableKinds                    []string
	InlineExternalTablePartitionColumnFunctions []string
	JoinHintRemotes                             []string
	JoinHintStrategies                          []string
	JoinKinds                                   []string
	LookupKinds                                 []string
	MakeSeriesKinds                             []string
	MvExpandKinds                               []string
	ParseKinds                                  []string
	PartitionHintConcurrencies                  []string
	PartitionHintSpreads                        []string
	PartitionHintStrategies                     []string
	PartitionedGraphMakeHintStrategies          []string
	ReduceByKinds                               []string
	ScanKinds                                   []string
	ScanStepOutputValues                        []string
	SearchKinds                                 []string
	ShortestPathsOutputs                        []string
	SortHintStrategies                          []string
	SummarizeHintStrategies                     []string
	ToScalarKinds                               []string
	ToTableKinds                                []string
	LimitExamples                               []string
	TopExamples                                 []string
	KnownInternalFunctionNames                  []string
	UnionWithSourceProperties                   []string
	UnionIsFuzzyProperty                        string
	UnionHintConcurrencies                      []string
	UnionHintSpreads                            []string
	UnionKinds                                  []string
	KeywordsAsIdentifiers                       []SyntaxKind
	ExtendedKeywordsAsIdentifiers               []SyntaxKind
	SpecialKeywordsAsIdentifiers                []SyntaxKind
	ForkOperatorKinds                           []SyntaxKind
	PostPipeOperatorKinds                       []SyntaxKind
}{
	AgoExamples:                    []string{"30min", "1h", "1d"},
	ParamTypes:                     []string{"bool", "boolean", "date", "datetime", "decimal", "double", "dynamic", "guid", "int", "int64", "int8", "long", "real", "string", "time", "timespan", "uniqueid"},
	ExtendedParamTypes:             []string{"bool", "boolean", "date", "datetime", "decimal", "double", "dynamic", "float", "guid", "int", "int16", "int32", "int64", "int8", "long", "real", "decimal", "string", "time", "timespan", "uint", "uint16", "uint32", "uint64", "uint8", "ulong", "uniqueid"},
	StorageTypes:                   []string{"bool", "boolean", "date", "datetime", "decimal", "double", "dynamic", "float", "guid", "int", "int16", "int32", "int64", "int8", "long", "real", "decimal", "string", "time", "timespan", "uint", "uint16", "uint32", "uint64", "uint8", "ulong", "uniqueid"},
	ChartAxis:                      []string{"linear", "log"},
	ChartKinds:                     []string{"default", "unstacked", "stacked", "stacked100", "map"},
	ChartLegends:                   []string{"visible", "hidden"},
	ChartProperties:                []string{"title", "xcolumn", "series", "ycolumns", "anomalycolumns", "kind", "xtitle", "ytitle", "xaxis", "yaxis", "legend", "ysplit", "accumulate", "ymin", "ymax", "xmax", "xmin"},
	ChartTypes:                     []string{"table", "list", "barchart", "piechart", "ladderchart", "timechart", "linechart", "anomalychart", "pivotchart", "areachart", "stackedareachart", "scatterchart", "timepivot", "columnchart", "timeline", "3Dchart", "card", "treemap", "plotly", "graph", "sankey"},
	HiddenChartTypes:               []string{"3Dchart", "sankey", "graph"},
	VisibleChartTypes:              []string{"table", "list", "barchart", "piechart", "ladderchart", "timechart", "linechart", "anomalychart", "pivotchart", "areachart", "stackedareachart", "scatterchart", "timepivot", "columnchart", "timeline", "card", "treemap", "plotly"},
	ChartYSplit:                    []string{"none", "axes", "panels"},
	CyclesKinds:                    []string{"none", "all", "unique_edges"},
	DataScopeValues:                []string{"all", "hotcache"},
	DateDiffParts:                  []string{"year", "quarter", "month", "week", "day", "hour", "minute", "second", "millisecond", "microsecond", "nanosecond"},
	DateTimeParts:                  []string{"year", "quarter", "month", "week_of_year", "day", "dayofyear", "hour", "minute", "second", "millisecond", "microsecond", "nanosecond"},
	Directives:                     []string{"automate", "browse", "checkconnectivity", "connect", "connectivitycheck", "crp", "database", "download", "qp", "query", "run", "save", "sqr", "truesight", "upload", "welcome"},
	DistinctHintStrategies:         []string{"auto", "broadcast", "centralized", "shuffle"},
	EvaluateHintDistributions:      []string{"single", "per_node", "per_shard", "default"},
	EvaluateHintRemotes:            []string{"auto", "local"},
	GraphMarkComponentsKinds:       []string{"weak", "strong"},
	HintConcurrencies:              []string{"auto", "left", "local", "right", "unresolved"},
	HintDistributions:              []string{"single", "per_node", "per_shard", "default"},
	HintRemotes:                    []string{"auto", "local"},
	HintSpreads:                    []string{"auto", "local"},
	HintStrategies:                 []string{"auto", "broadcast", "centralized", "shuffle"},
	InlineExternalTableDataFormats: []string{"parquet", "avro", "csv", "tsv", "json", "orc", "txt"},
	InlineExternalTableKinds:       []string{"storage", "delta"},
	InlineExternalTablePartitionColumnFunctions: []string{"hash", "bin", "startofday", "startofweek", "startofmonth", "startofyear"},
	JoinHintRemotes:                    []string{"auto", "left", "local", "right", "unresolved"},
	JoinHintStrategies:                 []string{"auto", "broadcast", "centralized", "shuffle"},
	JoinKinds:                          []string{"inner", "fullouter", "innerunique", "leftanti", "leftantisemi", "anti", "leftouter", "leftsemi", "rightanti", "rightantisemi", "rightouter", "rightsemi"},
	LookupKinds:                        []string{"inner", "leftouter"},
	MakeSeriesKinds:                    []string{"nonempty"},
	MvExpandKinds:                      []string{"bag", "array"},
	ParseKinds:                         []string{"simple", "regex", "relaxed"},
	PartitionHintConcurrencies:         []string{"auto", "left", "local", "right", "unresolved"},
	PartitionHintSpreads:               []string{"auto", "local"},
	PartitionHintStrategies:            []string{"shuffle", "native", "legacy"},
	PartitionedGraphMakeHintStrategies: []string{"shuffle", "native"},
	ReduceByKinds:                      []string{"source"},
	ScanKinds:                          []string{"partial", "full"},
	ScanStepOutputValues:               []string{"all", "last", "none"},
	SearchKinds:                        []string{"default", "case_insensitive", "case_sensitive"},
	ShortestPathsOutputs:               []string{"any", "all"},
	SortHintStrategies:                 []string{"splitBlock", "multipleBlocks"},
	SummarizeHintStrategies:            []string{"shuffle"},
	ToScalarKinds:                      []string{"nooptimization"},
	ToTableKinds:                       []string{"nooptimization"},
	LimitExamples:                      []string{"10", "100", "1000"},
	TopExamples:                        []string{"10", "100", "1000"},
	KnownInternalFunctionNames: []string{
		"__blackbox", "__box", "__cast", "__columnifexists", "__const_cast", "__fetch_contextual_scalar_value",
		"__geo_line_validate", "__geo_polygon_validate", "__getobject", "__getobjectex", "__has_ipv4", "__has_ipv4_prefix",
		"__hash_crc32", "__hash_djb2", "__hash_many_crc32", "__hash_xxxh64", "__invoke", "__is_scalar",
		"__lz4_compress_dynamic_array_to_base64_string", "__make_const", "__null", "__null_as", "__object", "__query_parameter",
		"__regex_complexity", "__row_offset", "__rowstore_ref", "__search_wildcard", "__search_wildcard_explicit_cols",
		"__search_wildcard_explicit_cols_v35", "__shard_record_position", "__sql_add", "__sql_divide", "__sql_modulo",
		"__sql_multiply", "__sql_subtract", "__trace_information", "__var", "__warning", "__get_scalar",
	},
	UnionWithSourceProperties: []string{"withsource", "with_source"},
	UnionIsFuzzyProperty:      "isfuzzy",
	UnionHintConcurrencies:    []string{"auto", "left", "local", "right", "unresolved"},
	UnionHintSpreads:          []string{"auto", "local"},
	UnionKinds:                []string{"inner", "outer"},
	SpecialKeywordsAsIdentifiers: []SyntaxKind{
		SyntaxKindKindKeyword, SyntaxKindWithSourceKeyword, SyntaxKindWith_SourceKeyword,
	},
	ForkOperatorKinds: []SyntaxKind{
		SyntaxKindAsOperator, SyntaxKindCountOperator, SyntaxKindDistinctOperator, SyntaxKindExecuteAndCacheOperator,
		SyntaxKindExtendOperator, SyntaxKindFilterOperator, SyntaxKindInvokeOperator, SyntaxKindMvExpandOperator,
		SyntaxKindParseOperator, SyntaxKindParseWhereOperator, SyntaxKindParseKvOperator, SyntaxKindProjectOperator,
		SyntaxKindProjectAwayOperator, SyntaxKindProjectByNamesOperator, SyntaxKindProjectKeepOperator, SyntaxKindProjectRenameOperator,
		SyntaxKindProjectReorderOperator, SyntaxKindReduceByOperator, SyntaxKindSampleDistinctOperator, SyntaxKindSampleOperator,
		SyntaxKindSortOperator, SyntaxKindSummarizeOperator, SyntaxKindTakeOperator, SyntaxKindTopHittersOperator,
		SyntaxKindTopNestedOperator, SyntaxKindTopOperator,
	},
	PostPipeOperatorKinds: []SyntaxKind{
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
	},
}

func initializeKustoFactsSyntaxKinds() {
	keywords := make([]SyntaxKind, 0)
	for _, kind := range SyntaxKindsWithFixedText() {
		if IsKeywordKind(kind) && SyntaxKindCanBeIdentifier(kind) {
			keywords = append(keywords, kind)
		}
	}
	KustoFacts.KeywordsAsIdentifiers = keywords
	KustoFacts.ExtendedKeywordsAsIdentifiers = append(append([]SyntaxKind{}, keywords...),
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
	)
}

// KustoFactsMatches mirrors KustoFacts.Matches for wildcard patterns where '*'
// means any zero-or-more characters. Empty patterns do not match anything.
func KustoFactsMatches(pattern string, text string, ignoreCase ...bool) bool {
	if pattern == "" {
		return false
	}
	ignore := false
	if len(ignoreCase) > 0 {
		ignore = ignoreCase[0]
	}
	return kustoFactsMatches(utf16.Encode([]rune(pattern)), 0, utf16.Encode([]rune(text)), 0, ignore)
}

func kustoFactsMatches(pattern []uint16, patternSegmentStart int, text []uint16, textPosition int, ignoreCase bool) bool {
	asteriskPosition := -1
	for i, ch := range pattern[patternSegmentStart:] {
		if ch == '*' {
			asteriskPosition = i
			break
		}
	}
	sawAsterisk := asteriskPosition >= 0
	if sawAsterisk {
		asteriskPosition += patternSegmentStart
	}

	patternSegmentEnd := len(pattern)
	if sawAsterisk {
		patternSegmentEnd = asteriskPosition
		for asteriskPosition+1 < len(pattern) && pattern[asteriskPosition+1] == '*' {
			asteriskPosition++
		}
	}

	nextPatternStart := len(pattern)
	if sawAsterisk {
		nextPatternStart = asteriskPosition + 1
	}
	patternSegmentLength := patternSegmentEnd - patternSegmentStart

	switch {
	case patternSegmentLength == 0:
		if patternSegmentStart >= len(pattern) {
			return true
		}
		return kustoFactsMatches(pattern, nextPatternStart, text, textPosition, ignoreCase)
	case patternSegmentStart == 0:
		if !sawAsterisk {
			return len(text) == patternSegmentLength &&
				compareKustoFactSegment(text, 0, pattern, 0, patternSegmentLength, ignoreCase)
		}
		if patternSegmentLength > len(text) ||
			!compareKustoFactSegment(text, 0, pattern, 0, patternSegmentLength, ignoreCase) {
			return false
		}
		return kustoFactsMatches(pattern, nextPatternStart, text, patternSegmentLength, ignoreCase)
	case !sawAsterisk:
		return patternSegmentLength <= len(text)-textPosition &&
			compareKustoFactSegment(text, len(text)-patternSegmentLength, pattern, patternSegmentStart, patternSegmentLength, ignoreCase)
	default:
		matchesPosition := indexOfKustoFactSegment(text, textPosition, pattern, patternSegmentStart, patternSegmentLength, ignoreCase)
		if matchesPosition == -1 {
			return false
		}
		return kustoFactsMatches(pattern, nextPatternStart, text, matchesPosition+patternSegmentLength, ignoreCase)
	}
}

func compareKustoFactSegment(text []uint16, textStart int, pattern []uint16, patternStart int, length int, ignoreCase bool) bool {
	if textStart < 0 || patternStart < 0 || textStart+length > len(text) || patternStart+length > len(pattern) {
		return false
	}
	left := text[textStart : textStart+length]
	right := pattern[patternStart : patternStart+length]
	return ordinalCompareUnits(left, right, ignoreCase) == 0
}

func indexOfKustoFactSegment(text []uint16, textStart int, value []uint16, valueStart int, valueLength int, ignoreCase bool) int {
	if textStart < 0 || textStart > len(text) || valueStart < 0 || valueStart+valueLength > len(value) || valueLength == 0 {
		return -1
	}
	for position := textStart; position+valueLength <= len(text); position++ {
		if compareKustoFactSegment(text, position, value, valueStart, valueLength, ignoreCase) {
			return position
		}
	}
	return -1
}
