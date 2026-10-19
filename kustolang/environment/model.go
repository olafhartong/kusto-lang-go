package environment

type KQLEnvironments map[string]*EnvironmentDefinition

type EnvironmentDefinition struct {
	// FunctionCatalogs selects bundled service function definitions (currently "sentinel").
	FunctionCatalogs             []string                          `json:"function_catalogs"`
	TableDetails                 map[string]TableDetails           `json:"tables"`
	TableOptions                 map[string]TableOptions           `json:"table_options,omitempty"`
	ExternalTables               map[string]TableDetails           `json:"external_tables"`
	ExternalTableOptions         map[string]TableOptions           `json:"external_table_options,omitempty"`
	MaterializedViews            map[string]TableDetails           `json:"materialized_views"`
	MaterializedViewOptions      map[string]TableOptions           `json:"materialized_view_options,omitempty"`
	MaterializedViewQueries      map[string]string                 `json:"materialized_view_queries,omitempty"`
	StoredQueryResults           map[string]TableDetails           `json:"stored_query_results"`
	StoredQueryResultOptions     map[string]TableOptions           `json:"stored_query_result_options,omitempty"`
	GraphModels                  map[string]GraphDetails           `json:"graph_models"`
	EntityGroups                 map[string][]string               `json:"entity_groups"`
	MagicFunctions               []string                          `json:"magic_functions"`
	TabularFunctions             map[string]TabularFunctionDetails `json:"tabular_functions,omitempty"`
	ScalarFunctions              map[string]ScalarFunctionDetails  `json:"scalar_functions,omitempty"`
	Domain                       string                            `json:"domain,omitempty"`
	CurrentCluster               string                            `json:"current_cluster,omitempty"`
	CurrentDatabase              string                            `json:"current_database,omitempty"`
	Clusters                     map[string]ClusterDetails         `json:"clusters,omitempty"`
	TableOrder                   []string                          `json:"-"`
	TableColumnOrder             map[string][]string               `json:"-"`
	ExternalTableOrder           []string                          `json:"-"`
	ExternalTableColumnOrder     map[string][]string               `json:"-"`
	MaterializedViewOrder        []string                          `json:"-"`
	MaterializedViewColumnOrder  map[string][]string               `json:"-"`
	StoredQueryResultOrder       []string                          `json:"-"`
	StoredQueryResultColumnOrder map[string][]string               `json:"-"`
}

type ClusterDetails struct {
	Databases map[string]DatabaseDetails `json:"databases"`
}

type DatabaseDetails struct {
	AlternateName                string                            `json:"alternate_name,omitempty"`
	TableDetails                 map[string]TableDetails           `json:"tables"`
	TableOptions                 map[string]TableOptions           `json:"table_options,omitempty"`
	ExternalTables               map[string]TableDetails           `json:"external_tables"`
	ExternalTableOptions         map[string]TableOptions           `json:"external_table_options,omitempty"`
	MaterializedViews            map[string]TableDetails           `json:"materialized_views"`
	MaterializedViewOptions      map[string]TableOptions           `json:"materialized_view_options,omitempty"`
	MaterializedViewQueries      map[string]string                 `json:"materialized_view_queries,omitempty"`
	StoredQueryResults           map[string]TableDetails           `json:"stored_query_results"`
	StoredQueryResultOptions     map[string]TableOptions           `json:"stored_query_result_options,omitempty"`
	GraphModels                  map[string]GraphDetails           `json:"graph_models"`
	EntityGroups                 map[string][]string               `json:"entity_groups"`
	MagicFunctions               []string                          `json:"magic_functions"`
	TabularFunctions             map[string]TabularFunctionDetails `json:"tabular_functions,omitempty"`
	ScalarFunctions              map[string]ScalarFunctionDetails  `json:"scalar_functions,omitempty"`
	TableOrder                   []string                          `json:"-"`
	TableColumnOrder             map[string][]string               `json:"-"`
	ExternalTableOrder           []string                          `json:"-"`
	ExternalTableColumnOrder     map[string][]string               `json:"-"`
	MaterializedViewOrder        []string                          `json:"-"`
	MaterializedViewColumnOrder  map[string][]string               `json:"-"`
	StoredQueryResultOrder       []string                          `json:"-"`
	StoredQueryResultColumnOrder map[string][]string               `json:"-"`
}

type TableDetails map[string]string
type WatchlistDetails map[string]string

// TableOptions describes table-like symbol properties that cannot be
// represented by the backward-compatible column-name-to-type TableDetails map.
type TableOptions struct {
	IsOpen       bool                         `json:"is_open,omitempty"`
	IsSerialized bool                         `json:"is_serialized,omitempty"`
	IsSorted     bool                         `json:"is_sorted,omitempty"`
	ColumnTypes  map[string]ColumnTypeDetails `json:"column_types,omitempty"`
}

// ColumnTypeDetails describes the known shape of a dynamic column. Properties
// represent a property bag and Element represents an array element type.
// PropertyOrder is optional for Go callers; JSON object declaration order is
// captured automatically.
type ColumnTypeDetails struct {
	Type          string                       `json:"type"`
	Properties    map[string]ColumnTypeDetails `json:"properties,omitempty"`
	PropertyOrder []string                     `json:"-"`
	Element       *ColumnTypeDetails           `json:"element,omitempty"`
}

type GraphDetails struct {
	Edges     map[string]TableDetails `json:"edges"`
	Nodes     map[string]TableDetails `json:"nodes"`
	Snapshots []string                `json:"snapshots"`
}

type LocalData struct {
	Watchlists                   map[string]WatchlistDetails       `json:"watchlists"`
	Tables                       map[string]TableDetails           `json:"tables"`
	TableOptions                 map[string]TableOptions           `json:"table_options,omitempty"`
	ExternalTables               map[string]TableDetails           `json:"external_tables"`
	ExternalTableOptions         map[string]TableOptions           `json:"external_table_options,omitempty"`
	MaterializedViews            map[string]TableDetails           `json:"materialized_views"`
	MaterializedViewOptions      map[string]TableOptions           `json:"materialized_view_options,omitempty"`
	MaterializedViewQueries      map[string]string                 `json:"materialized_view_queries,omitempty"`
	StoredQueryResults           map[string]TableDetails           `json:"stored_query_results"`
	StoredQueryResultOptions     map[string]TableOptions           `json:"stored_query_result_options,omitempty"`
	GraphModels                  map[string]GraphDetails           `json:"graph_models"`
	TabularFunctions             map[string]TabularFunctionDetails `json:"tabular_functions"`
	ScalarFunctions              map[string]ScalarFunctionDetails  `json:"scalar_functions"`
	TableOrder                   []string                          `json:"-"`
	TableColumnOrder             map[string][]string               `json:"-"`
	ExternalTableOrder           []string                          `json:"-"`
	ExternalTableColumnOrder     map[string][]string               `json:"-"`
	MaterializedViewOrder        []string                          `json:"-"`
	MaterializedViewColumnOrder  map[string][]string               `json:"-"`
	StoredQueryResultOrder       []string                          `json:"-"`
	StoredQueryResultColumnOrder map[string][]string               `json:"-"`
}

type ArgumentDetails struct {
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	TypeDetails *ColumnTypeDetails `json:"type_details,omitempty"`
	Optional    bool               `json:"optional"`
	MissingType bool               `json:"-"`
}

type ScalarFunctionDetails struct {
	OutputType        string             `json:"output_type"`
	OutputTypeDetails *ColumnTypeDetails `json:"output_type_details,omitempty"`
	Arguments         []ArgumentDetails  `json:"arguments"`
}

type TabularFunctionDetails struct {
	IsOpen            bool                         `json:"is_open,omitempty"`
	OutputColumns     TableDetails                 `json:"output_columns"`
	OutputColumnTypes map[string]ColumnTypeDetails `json:"output_column_types,omitempty"`
	Arguments         []ArgumentDetails            `json:"arguments"`
}

type ResultColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ResultTable struct {
	Name    string         `json:"name"`
	Columns []ResultColumn `json:"columns"`
	IsOpen  bool           `json:"is_open"`
}

func (l *LocalData) normalize() {
	if l.Watchlists == nil {
		l.Watchlists = map[string]WatchlistDetails{}
	}
	if l.Tables == nil {
		l.Tables = map[string]TableDetails{}
	}
	if l.TableOptions == nil {
		l.TableOptions = map[string]TableOptions{}
	}
	if l.ExternalTables == nil {
		l.ExternalTables = map[string]TableDetails{}
	}
	if l.ExternalTableOptions == nil {
		l.ExternalTableOptions = map[string]TableOptions{}
	}
	if l.MaterializedViews == nil {
		l.MaterializedViews = map[string]TableDetails{}
	}
	if l.MaterializedViewOptions == nil {
		l.MaterializedViewOptions = map[string]TableOptions{}
	}
	if l.MaterializedViewQueries == nil {
		l.MaterializedViewQueries = map[string]string{}
	}
	if l.StoredQueryResults == nil {
		l.StoredQueryResults = map[string]TableDetails{}
	}
	if l.StoredQueryResultOptions == nil {
		l.StoredQueryResultOptions = map[string]TableOptions{}
	}
	if l.GraphModels == nil {
		l.GraphModels = map[string]GraphDetails{}
	}
	if l.TabularFunctions == nil {
		l.TabularFunctions = map[string]TabularFunctionDetails{}
	}
	if l.ScalarFunctions == nil {
		l.ScalarFunctions = map[string]ScalarFunctionDetails{}
	}
}
