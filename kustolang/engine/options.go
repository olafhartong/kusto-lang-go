package kustolang

// Ported from src/Kusto.Language/Options.cs.
// This catalog is generated upstream from Kusto engine query-option tests.

type optionSet struct {
	BestEffort                           *OptionSymbol
	DebugPython                          *OptionSymbol
	DeferPartialQueryFailures            *OptionSymbol
	DoNotImpersonate                     *OptionSymbol
	ErrorReportingPlacement              *OptionSymbol
	MaterializedViewShuffleQuery         *OptionSymbol
	MaxEntitiesToUnion                   *OptionSymbol
	MaxMemoryConsumptionPerIterator      *OptionSymbol
	MaxMemoryConsumptionPerQueryPerNode  *OptionSymbol
	MaxOutputColumns                     *OptionSymbol
	NoRequestTimeout                     *OptionSymbol
	NoTruncation                         *OptionSymbol
	ProgressiveProgressReportPeriod      *OptionSymbol
	ProgressiveQueryMinRowCountPerUpdate *OptionSymbol
	PushSelectionThroughAggregation      *OptionSymbol
	QueryBinAutoAt                       *OptionSymbol
	QueryBinAutoSize                     *OptionSymbol
	QueryCursorAfterDefault              *OptionSymbol
	QueryCursorBeforeOrAtDefault         *OptionSymbol
	QueryCursorCurrent                   *OptionSymbol
	QueryCursorDisabled                  *OptionSymbol
	QueryCursorScopedTables              *OptionSymbol
	QueryDataScope                       *OptionSymbol
	QueryDateTimeScopeColumn             *OptionSymbol
	QueryDateTimeScopeFrom               *OptionSymbol
	QueryDateTimeScopeTo                 *OptionSymbol
	QueryDistributionNodesSpanSize       *OptionSymbol
	QueryFanoutNodesPercent              *OptionSymbol
	QueryFanoutThreadsPercent            *OptionSymbol
	QueryForceRowLevelSecurity           *OptionSymbol
	QueryLanguage                        *OptionSymbol
	QueryLogQueryParameters              *OptionSymbol
	QueryNow                             *OptionSymbol
	QueryResultsApplyGetSchema           *OptionSymbol
	QueryResultsCacheForceRefresh        *OptionSymbol
	QueryResultsCacheMaxAge              *OptionSymbol
	QueryResultsCachePerShardEnabled     *OptionSymbol
	QueryWeakConsistencySessionId        *OptionSymbol
	RequestAppName                       *OptionSymbol
	RequestBlockRowLevelSecurity         *OptionSymbol
	RequestCalloutDisabled               *OptionSymbol
	RequestCrossClusterToken             *OptionSymbol
	RequestDescription                   *OptionSymbol
	RequestExternalDataDisabled          *OptionSymbol
	RequestExternalTableDisabled         *OptionSymbol
	RequestReadOnly                      *OptionSymbol
	RequestReadOnlyHardline              *OptionSymbol
	RequestRemoteEntitiesDisabled        *OptionSymbol
	RequestSandboxedExecutionDisabled    *OptionSymbol
	RequestUser                          *OptionSymbol
	ResultsProgressiveEnabled            *OptionSymbol
	ResultsV2NewlinesBetweenFrames       *OptionSymbol
	ServerTimeout                        *OptionSymbol
	TakeMaxRecords                       *OptionSymbol
	TruncationMaxRecords                 *OptionSymbol
	TruncationMaxSize                    *OptionSymbol
	V2FragmentPrimaryTables              *OptionSymbol
	ValidatePermissions                  *OptionSymbol

	All []*OptionSymbol
}

var Options = newOptionSet()

func newOptionSet() *optionSet {
	options := &optionSet{}
	options.BestEffort = optionWithType(
		"best_effort",
		`If set, allows fuzzy resolution and connectivity issues of union legs. The set of union sources is reduced to the set of table references that exist and are accessible at the time.
If at least one such table is found, any failure will yield a warning in the query status results, but will not prevent the query execution.`,
		ScalarTypes.Bool,
	)
	options.DebugPython = optionWithTypes(
		"query_python_debug",
		"If set, generate python debug query for the enumerated python node (default first).",
		[]*ScalarSymbol{ScalarTypes.Bool, ScalarTypes.Int},
	)
	options.DeferPartialQueryFailures = optionWithType(
		"deferpartialqueryfailures",
		"If true, disables reporting partial query failures as part of the result set.",
		ScalarTypes.Bool,
	)
	options.DoNotImpersonate = optionWithType(
		"request_impersonation_disabled",
		"If specified, indicates that the service should not impersonate the caller's identity.",
		ScalarTypes.Bool,
	)
	options.ErrorReportingPlacement = optionWithType(
		"results_error_reporting_placement",
		"Decides the placement of errors in the result set.",
		ScalarTypes.String,
		"'in_data'",
		"'end_of_table'",
		"'end_of_dataset'",
	)
	options.MaterializedViewShuffleQuery = optionWithType(
		"materialized_view_shuffle",
		`A hint to use shuffle strategy for materialized views that are referenced in the query.
The property is an array of materialized views names and the shuffle keys to use.
examples: 'dynamic(`,
		ScalarTypes.Dynamic,
	)
	options.MaxEntitiesToUnion = optionWithType(
		"query_max_entities_in_union",
		"Overrides the default maximum number of columns a query is allowed to produce.",
		ScalarTypes.Long,
	)
	options.MaxMemoryConsumptionPerIterator = optionWithType(
		"maxmemoryconsumptionperiterator",
		"Overrides the default maximum amount of memory a query operator may allocate.",
		ScalarTypes.Long,
	)
	options.MaxMemoryConsumptionPerQueryPerNode = optionWithType(
		"max_memory_consumption_per_query_per_node",
		"Overrides the default maximum amount of memory a whole query may allocate per node.",
		ScalarTypes.Long,
	)
	options.MaxOutputColumns = optionWithType(
		"maxoutputcolumns",
		"Overrides the default maximum number of columns a query is allowed to produce.",
		ScalarTypes.Long,
	)
	options.NoRequestTimeout = optionWithType(
		"norequesttimeout",
		"Enables setting the request timeout to its maximum value. This option cannot be set via a set-statement.",
		ScalarTypes.Bool,
	)
	options.NoTruncation = optionWithType(
		"notruncation",
		"Enables suppressing truncation of the query results returned to the caller.",
		ScalarTypes.Bool,
	)
	options.ProgressiveProgressReportPeriod = option(
		"query_results_progressive_update_period",
		"Hint for Kusto as to how often to send progress frames (takes effect only if OptionResultsProgressiveEnabled is set)",
	)
	options.ProgressiveQueryMinRowCountPerUpdate = option(
		"query_results_progressive_row_count",
		"Hint for Kusto as to how many records to send in each update (takes effect only if OptionResultsProgressiveEnabled is set)",
	)
	options.PushSelectionThroughAggregation = optionWithType(
		"push_selection_through_aggregation",
		"If true, push simple selection through aggregation",
		ScalarTypes.Bool,
	)
	options.QueryBinAutoAt = option(
		"query_bin_auto_at",
		"When evaluating the bin_auto() function, the start value to use.",
	)
	options.QueryBinAutoSize = option(
		"query_bin_auto_size",
		"When evaluating the bin_auto() function, the bin size value to use.",
	)
	options.QueryCursorAfterDefault = optionWithType(
		"query_cursor_after_default",
		"The default parameter value of the cursor_after() function when called without parameters.",
		ScalarTypes.String,
	)
	options.QueryCursorBeforeOrAtDefault = optionWithType(
		"query_cursor_before_or_at_default",
		"The default parameter value of the cursor_before_or_at() function when called without parameters.",
		ScalarTypes.String,
	)
	options.QueryCursorCurrent = optionWithType(
		"query_cursor_current",
		"Overrides the cursor value returned by the cursor_current() or current_cursor() functions.",
		ScalarTypes.String,
	)
	options.QueryCursorDisabled = optionWithType(
		"query_cursor_disabled",
		"Disables usage of cursor functions in the context of the query.",
		ScalarTypes.Bool,
	)
	options.QueryCursorScopedTables = optionWithType(
		"query_cursor_scoped_tables",
		"List of table names that should be scoped to cursor_after_default .. cursor_before_or_at_default (upper bound is optional).",
		ScalarTypes.Dynamic,
	)
	options.QueryDataScope = optionWithType(
		"query_datascope",
		"Controls the query's datascope -- whether the query applies to all data or just part of it.",
		ScalarTypes.String,
		"'default'",
		"'all'",
		"'hotcache'",
	)
	options.QueryDateTimeScopeColumn = optionWithType(
		"query_datetimescope_column",
		"Controls the column name for the query's datetime scope (query_datetimescope_to / query_datetimescope_from).",
		ScalarTypes.String,
	)
	options.QueryDateTimeScopeFrom = optionWithType(
		"query_datetimescope_from",
		"Controls the query's datetime scope (earliest) -- used as auto-applied filter on query_datetimescope_column only (if defined).",
		ScalarTypes.DateTime,
	)
	options.QueryDateTimeScopeTo = optionWithType(
		"query_datetimescope_to",
		"Controls the query's datetime scope (latest) -- used as auto-applied filter on query_datetimescope_column only (if defined).",
		ScalarTypes.DateTime,
	)
	options.QueryDistributionNodesSpanSize = optionWithType(
		"query_distribution_nodes_span",
		`If set, controls the way the subquery merge behaves: the executing node will introduce an additional level
in the query hierarchy for each subgroup of nodes; the size of the subgroup is set by this option.`,
		ScalarTypes.Int,
	)
	options.QueryFanoutNodesPercent = optionWithType(
		"query_fanout_nodes_percent",
		"The percentage of nodes to fan out execution to.",
		ScalarTypes.Int,
	)
	options.QueryFanoutThreadsPercent = optionWithType(
		"query_fanout_threads_percent",
		"The percentage of threads to fan out execution to.",
		ScalarTypes.Int,
	)
	options.QueryForceRowLevelSecurity = optionWithType(
		"query_force_row_level_security",
		"If specified, forces Row Level Security rules, even if row_level_security policy is disabled",
		ScalarTypes.Bool,
	)
	options.QueryLanguage = optionWithType(
		"query_language",
		"Controls how the query text is to be interpreted.",
		ScalarTypes.String,
		"'csl'",
		"'kql'",
		"'sql'",
	)
	options.QueryLogQueryParameters = optionWithType(
		"query_log_query_parameters",
		"Enables logging of the query parameters, so that they can be viewed later in the `.show queries` journal.",
		ScalarTypes.Bool,
	)
	options.QueryNow = optionWithType(
		"query_now",
		"Overrides the datetime value returned by the now(0s) function.",
		ScalarTypes.DateTime,
	)
	options.QueryResultsApplyGetSchema = optionWithType(
		"query_results_apply_getschema",
		"If set, retrieves the schema of each tabular data in the results of the query instead of the data itself.",
		ScalarTypes.Bool,
	)
	options.QueryResultsCacheForceRefresh = optionWithType(
		"query_results_cache_force_refresh",
		"If set, forces query results cache refresh for a specific query. Must be used in combination with 'query_results_cache_max_age', and sent via ClientRequestProperties object (not as 'set' statement).",
		ScalarTypes.Bool,
	)
	options.QueryResultsCacheMaxAge = optionWithType(
		"query_results_cache_max_age",
		"If positive, controls the maximum age of the cached query results the service is allowed to return",
		ScalarTypes.TimeSpan,
	)
	options.QueryResultsCachePerShardEnabled = optionWithType(
		"query_results_cache_per_shard",
		"If set, enables per-shard query cache.",
		ScalarTypes.Bool,
	)
	options.QueryWeakConsistencySessionId = optionWithType(
		"query_weakconsistency_session_id",
		"Sets the query weak consistency session id. Takes effect when 'queryconsistency' mode is set to 'weakconsistency_by_session_id'.",
		ScalarTypes.String,
	)
	options.RequestAppName = optionWithType(
		"request_app_name",
		"Request application name to be used in the reporting (e.g. show queries).",
		ScalarTypes.String,
	)
	options.RequestBlockRowLevelSecurity = optionWithType(
		"request_block_row_level_security",
		"If specified, blocks access to tables for which row_level_security policy is enabled",
		ScalarTypes.Bool,
	)
	options.RequestCalloutDisabled = optionWithType(
		"request_callout_disabled",
		"If specified, indicates that the request can't call-out to a user-provided service.",
		ScalarTypes.Bool,
	)
	options.RequestCrossClusterToken = optionWithType(
		"request_cross_cluster_token",
		"Cryptographically-strong indication the query is being sent from one cluster to another, as a cross-cluster query",
		ScalarTypes.String,
	)
	options.RequestDescription = optionWithType(
		"request_description",
		"Arbitrary text that the author of the request wants to include as the request description.",
		ScalarTypes.String,
	)
	options.RequestExternalDataDisabled = optionWithType(
		"request_external_data_disabled",
		"If specified, indicates that the request can't access external data (using externaldata operator) or external tables.",
		ScalarTypes.Bool,
	)
	options.RequestExternalTableDisabled = optionWithType(
		"request_external_table_disabled",
		"If specified, indicates that the request can't access external tables.",
		ScalarTypes.Bool,
	)
	options.RequestReadOnly = optionWithType(
		"request_readonly",
		"If specified, indicates that the request can't write anything.",
		ScalarTypes.Bool,
	)
	options.RequestReadOnlyHardline = optionWithType(
		"request_readonly_hardline",
		"If specified, indicates that the request must not be able to write anything, and any functionality (such as plugins) in which this flag can't be honored will be disabled.",
		ScalarTypes.Bool,
	)
	options.RequestRemoteEntitiesDisabled = optionWithType(
		"request_remote_entities_disabled",
		"If specified, indicates that the request can't access remote databases and clusters.",
		ScalarTypes.Bool,
	)
	options.RequestSandboxedExecutionDisabled = optionWithType(
		"request_sandboxed_execution_disabled",
		"If specified, indicates that the request can't invoke code in the sandbox.",
		ScalarTypes.Bool,
	)
	options.RequestUser = optionWithType(
		"request_user",
		"Request user to be used in the reporting (e.g. show queries).",
		ScalarTypes.String,
	)
	options.ResultsProgressiveEnabled = option(
		"results_progressive_enabled",
		"If set, enables the progressive query stream",
	)
	options.ResultsV2NewlinesBetweenFrames = optionWithType(
		"results_v2_newlines_between_frames",
		"Adds new lines between frames in the results, in order to make it easier to parse them.",
		ScalarTypes.Bool,
	)
	options.ServerTimeout = optionWithType(
		"servertimeout",
		"Overrides the default request timeout. This option cannot be set via a set-statement.",
		ScalarTypes.TimeSpan,
	)
	options.TakeMaxRecords = optionWithType(
		"query_take_max_records",
		"Enables limiting query results to this number of records.",
		ScalarTypes.Long,
	)
	options.TruncationMaxRecords = optionWithType(
		"truncationmaxrecords",
		"Overrides the default maximum number of records a query is allowed to return to the caller (truncation).",
		ScalarTypes.Long,
	)
	options.TruncationMaxSize = optionWithType(
		"truncationmaxsize",
		"Overrides the default maximum data size a query is allowed to return to the caller (truncation).",
		ScalarTypes.Long,
	)
	options.V2FragmentPrimaryTables = optionWithType(
		"results_v2_fragment_primary_tables",
		"Causes primary tables to be sent in multiple fragments, each containing a subset of the rows.",
		ScalarTypes.Bool,
	)
	options.ValidatePermissions = optionWithType(
		"validate_permissions",
		"Validates user's permissions to perform the query and doesn't run the query itself.",
		ScalarTypes.Bool,
	)
	options.All = []*OptionSymbol{
		options.BestEffort,
		options.DebugPython,
		options.DeferPartialQueryFailures,
		options.DoNotImpersonate,
		options.ErrorReportingPlacement,
		options.MaterializedViewShuffleQuery,
		options.MaxEntitiesToUnion,
		options.MaxMemoryConsumptionPerIterator,
		options.MaxMemoryConsumptionPerQueryPerNode,
		options.MaxOutputColumns,
		options.NoRequestTimeout,
		options.NoTruncation,
		options.ProgressiveProgressReportPeriod,
		options.ProgressiveQueryMinRowCountPerUpdate,
		options.PushSelectionThroughAggregation,
		options.QueryBinAutoAt,
		options.QueryBinAutoSize,
		options.QueryCursorAfterDefault,
		options.QueryCursorBeforeOrAtDefault,
		options.QueryCursorCurrent,
		options.QueryCursorDisabled,
		options.QueryCursorScopedTables,
		options.QueryDataScope,
		options.QueryDateTimeScopeColumn,
		options.QueryDateTimeScopeFrom,
		options.QueryDateTimeScopeTo,
		options.QueryDistributionNodesSpanSize,
		options.QueryFanoutNodesPercent,
		options.QueryFanoutThreadsPercent,
		options.QueryForceRowLevelSecurity,
		options.QueryLanguage,
		options.QueryLogQueryParameters,
		options.QueryNow,
		options.QueryResultsApplyGetSchema,
		options.QueryResultsCacheForceRefresh,
		options.QueryResultsCacheMaxAge,
		options.QueryResultsCachePerShardEnabled,
		options.QueryWeakConsistencySessionId,
		options.RequestAppName,
		options.RequestBlockRowLevelSecurity,
		options.RequestCalloutDisabled,
		options.RequestCrossClusterToken,
		options.RequestDescription,
		options.RequestExternalDataDisabled,
		options.RequestExternalTableDisabled,
		options.RequestReadOnly,
		options.RequestReadOnlyHardline,
		options.RequestRemoteEntitiesDisabled,
		options.RequestSandboxedExecutionDisabled,
		options.RequestUser,
		options.ResultsProgressiveEnabled,
		options.ResultsV2NewlinesBetweenFrames,
		options.ServerTimeout,
		options.TakeMaxRecords,
		options.TruncationMaxRecords,
		options.TruncationMaxSize,
		options.V2FragmentPrimaryTables,
		options.ValidatePermissions,
	}
	return options
}

func option(name, description string, examples ...string) *OptionSymbol {
	return NewOptionSymbol(name, description, nil, examples)
}

func optionWithType(name, description string, typ *ScalarSymbol, examples ...string) *OptionSymbol {
	return NewOptionSymbolWithType(name, description, typ, examples)
}

func optionWithTypes(name, description string, types []*ScalarSymbol, examples ...string) *OptionSymbol {
	return NewOptionSymbol(name, description, types, examples)
}
