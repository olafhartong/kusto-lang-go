package environment

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"sort"
	"strings"

	kustolang "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

func LoadEnvironments(path string) (KQLEnvironments, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var envs KQLEnvironments
	if err := json.Unmarshal(data, &envs); err != nil {
		return nil, err
	}
	for _, envName := range sortedEnvironmentNames(envs) {
		env := envs[envName]
		if env == nil {
			continue
		}
		if env.FunctionCatalogs == nil && (envName == "sentinel" || envName == "m365_with_sentinel") {
			env.FunctionCatalogs = []string{"sentinel"}
		}
		if err := normalizeEnvironmentDefinition(env); err != nil {
			return nil, fmt.Errorf("environment %q: %w", envName, err)
		}
		if err := validateEntityGroupMembers(env); err != nil {
			return nil, err
		}
	}
	AddM365WithSentinelIfPresent(envs)
	return envs, nil
}

func normalizeEnvironmentDefinition(env *EnvironmentDefinition) error {
	if env == nil {
		return nil
	}
	normalizeFlatEnvironment(env)
	if err := validateFunctionCatalogs(env.FunctionCatalogs); err != nil {
		return err
	}
	if err := normalizeConfiguredFunctions("environment", env.ScalarFunctions, env.TabularFunctions); err != nil {
		return err
	}
	if err := normalizeMaterializedViewQueries(env.MaterializedViews, &env.MaterializedViewQueries); err != nil {
		return err
	}
	if env.Clusters == nil {
		env.Clusters = map[string]ClusterDetails{}
	}
	clusterNames := sortedClusterNames(env.Clusters)
	if len(clusterNames) > 1 && strings.TrimSpace(env.CurrentCluster) == "" {
		return fmt.Errorf("current_cluster is required when more than one cluster is configured")
	}
	seenClusters := map[string]string{}
	for _, clusterName := range clusterNames {
		if !canonicalClusterHostName(clusterName) {
			return fmt.Errorf("cluster name %q must be a canonical host name without a scheme, path, query, fragment, or port", clusterName)
		}
		folded := foldName(clusterName)
		if previous := seenClusters[folded]; previous != "" {
			return fmt.Errorf("cluster names %q and %q differ only by case", previous, clusterName)
		}
		seenClusters[folded] = clusterName
		cluster := env.Clusters[clusterName]
		if cluster.Databases == nil {
			cluster.Databases = map[string]DatabaseDetails{}
		}
		seenDatabases := map[string]string{}
		for _, databaseName := range sortedDatabaseNames(cluster.Databases) {
			foldedDatabase := foldName(databaseName)
			if previous := seenDatabases[foldedDatabase]; previous != "" {
				return fmt.Errorf("cluster %q database names %q and %q differ only by case", clusterName, previous, databaseName)
			}
			seenDatabases[foldedDatabase] = databaseName
			database := cluster.Databases[databaseName]
			normalizeDatabaseDetails(&database)
			if err := normalizeConfiguredFunctions(fmt.Sprintf("cluster %q database %q", clusterName, databaseName), database.ScalarFunctions, database.TabularFunctions); err != nil {
				return err
			}
			if err := normalizeTableOptions(database.TableDetails, &database.TableOptions); err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			if err := normalizeEntityTableOptions("external_table_options", "external table", database.ExternalTables, &database.ExternalTableOptions, true); err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			if err := normalizeEntityTableOptions("materialized_view_options", "materialized view", database.MaterializedViews, &database.MaterializedViewOptions, true); err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			if err := normalizeEntityTableOptions("stored_query_result_options", "stored query result", database.StoredQueryResults, &database.StoredQueryResultOptions, false); err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			if err := normalizeMaterializedViewQueries(database.MaterializedViews, &database.MaterializedViewQueries); err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			if err := validateDatabaseEntityGroupMembers(databaseName, &database); err != nil {
				return fmt.Errorf("cluster %q: %w", clusterName, err)
			}
			cluster.Databases[databaseName] = database
		}
		env.Clusters[clusterName] = cluster
	}
	if len(clusterNames) > 0 {
		clusterName, ok := configuredCurrentClusterName(env)
		if !ok {
			return fmt.Errorf("current_cluster %q is not configured", env.CurrentCluster)
		}
		cluster := env.Clusters[clusterName]
		if strings.TrimSpace(env.CurrentDatabase) == "" && len(cluster.Databases) > 1 {
			return fmt.Errorf("current_database is required when current cluster %q has more than one database", clusterName)
		}
		if strings.TrimSpace(env.CurrentDatabase) != "" && matchingDatabaseDetails(cluster.Databases, env.CurrentDatabase) == nil && !flatEnvironmentHasData(env) {
			return fmt.Errorf("current_database %q is not configured in current cluster %q and no root database schema is present", env.CurrentDatabase, clusterName)
		}
		availableTables := map[string]TableDetails{}
		availableExternalTables := map[string]TableDetails{}
		availableMaterializedViews := map[string]TableDetails{}
		availableStoredQueryResults := map[string]TableDetails{}
		if database := matchingDatabaseDetails(cluster.Databases, configuredCurrentDatabaseName(env, clusterName)); database != nil {
			availableTables = copyTableDetailsMap(database.TableDetails)
			availableExternalTables = copyTableDetailsMap(database.ExternalTables)
			availableMaterializedViews = copyTableDetailsMap(database.MaterializedViews)
			availableStoredQueryResults = copyTableDetailsMap(database.StoredQueryResults)
		}
		for name, table := range env.TableDetails {
			availableTables[name] = table
		}
		for name, table := range env.ExternalTables {
			availableExternalTables[name] = table
		}
		for name, table := range env.MaterializedViews {
			availableMaterializedViews[name] = table
		}
		for name, table := range env.StoredQueryResults {
			availableStoredQueryResults[name] = table
		}
		if err := normalizeTableOptions(availableTables, &env.TableOptions); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("external_table_options", "external table", availableExternalTables, &env.ExternalTableOptions, true); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("materialized_view_options", "materialized view", availableMaterializedViews, &env.MaterializedViewOptions, true); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("stored_query_result_options", "stored query result", availableStoredQueryResults, &env.StoredQueryResultOptions, false); err != nil {
			return err
		}
	} else {
		if err := normalizeTableOptions(env.TableDetails, &env.TableOptions); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("external_table_options", "external table", env.ExternalTables, &env.ExternalTableOptions, true); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("materialized_view_options", "materialized view", env.MaterializedViews, &env.MaterializedViewOptions, true); err != nil {
			return err
		}
		if err := normalizeEntityTableOptions("stored_query_result_options", "stored query result", env.StoredQueryResults, &env.StoredQueryResultOptions, false); err != nil {
			return err
		}
	}
	return nil
}

func canonicalClusterHostName(name string) bool {
	if name == "" || strings.TrimSpace(name) != name || strings.Contains(name, "://") {
		return false
	}
	parsed, err := url.Parse("//" + name)
	return err == nil && parsed.Hostname() == name && parsed.Port() == "" && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}

func matchingDatabaseDetails(databases map[string]DatabaseDetails, name string) *DatabaseDetails {
	for candidate, database := range databases {
		if strings.EqualFold(candidate, strings.TrimSpace(name)) || strings.EqualFold(database.AlternateName, strings.TrimSpace(name)) {
			copy := database
			return &copy
		}
	}
	return nil
}

func flatEnvironmentHasData(env *EnvironmentDefinition) bool {
	return env != nil && (len(env.TableDetails) > 0 || len(env.ExternalTables) > 0 || len(env.MaterializedViews) > 0 ||
		len(env.StoredQueryResults) > 0 || len(env.GraphModels) > 0 || len(env.EntityGroups) > 0 || len(env.MagicFunctions) > 0 ||
		len(env.TabularFunctions) > 0 || len(env.ScalarFunctions) > 0)
}

func normalizeFlatEnvironment(env *EnvironmentDefinition) {
	if env.TableDetails == nil {
		env.TableDetails = map[string]TableDetails{}
	}
	if env.TableOptions == nil {
		env.TableOptions = map[string]TableOptions{}
	}
	if env.ExternalTables == nil {
		env.ExternalTables = map[string]TableDetails{}
	}
	if env.ExternalTableOptions == nil {
		env.ExternalTableOptions = map[string]TableOptions{}
	}
	if env.MaterializedViews == nil {
		env.MaterializedViews = map[string]TableDetails{}
	}
	if env.MaterializedViewOptions == nil {
		env.MaterializedViewOptions = map[string]TableOptions{}
	}
	if env.MaterializedViewQueries == nil {
		env.MaterializedViewQueries = map[string]string{}
	}
	if env.StoredQueryResults == nil {
		env.StoredQueryResults = map[string]TableDetails{}
	}
	if env.StoredQueryResultOptions == nil {
		env.StoredQueryResultOptions = map[string]TableOptions{}
	}
	if env.GraphModels == nil {
		env.GraphModels = map[string]GraphDetails{}
	}
	if env.EntityGroups == nil {
		env.EntityGroups = map[string][]string{}
	}
	if env.MagicFunctions == nil {
		env.MagicFunctions = []string{}
	}
	if env.TabularFunctions == nil {
		env.TabularFunctions = map[string]TabularFunctionDetails{}
	}
	if env.ScalarFunctions == nil {
		env.ScalarFunctions = map[string]ScalarFunctionDetails{}
	}
}

func normalizeDatabaseDetails(database *DatabaseDetails) {
	if database.TableDetails == nil {
		database.TableDetails = map[string]TableDetails{}
	}
	if database.TableOptions == nil {
		database.TableOptions = map[string]TableOptions{}
	}
	if database.ExternalTables == nil {
		database.ExternalTables = map[string]TableDetails{}
	}
	if database.ExternalTableOptions == nil {
		database.ExternalTableOptions = map[string]TableOptions{}
	}
	if database.MaterializedViews == nil {
		database.MaterializedViews = map[string]TableDetails{}
	}
	if database.MaterializedViewOptions == nil {
		database.MaterializedViewOptions = map[string]TableOptions{}
	}
	if database.MaterializedViewQueries == nil {
		database.MaterializedViewQueries = map[string]string{}
	}
	if database.StoredQueryResults == nil {
		database.StoredQueryResults = map[string]TableDetails{}
	}
	if database.StoredQueryResultOptions == nil {
		database.StoredQueryResultOptions = map[string]TableOptions{}
	}
	if database.GraphModels == nil {
		database.GraphModels = map[string]GraphDetails{}
	}
	if database.EntityGroups == nil {
		database.EntityGroups = map[string][]string{}
	}
	if database.MagicFunctions == nil {
		database.MagicFunctions = []string{}
	}
	if database.TabularFunctions == nil {
		database.TabularFunctions = map[string]TabularFunctionDetails{}
	}
	if database.ScalarFunctions == nil {
		database.ScalarFunctions = map[string]ScalarFunctionDetails{}
	}
}

func AddM365WithSentinelIfPresent(envs KQLEnvironments) {
	m365, okM365 := envs["m365"]
	sentinel, okSentinel := envs["sentinel"]
	if !okM365 || !okSentinel {
		return
	}
	if len(m365.Clusters) > 0 || len(sentinel.Clusters) > 0 {
		return
	}

	merged := &EnvironmentDefinition{
		FunctionCatalogs:             append(append([]string{}, m365.FunctionCatalogs...), sentinel.FunctionCatalogs...),
		TableDetails:                 map[string]TableDetails{},
		TableOptions:                 map[string]TableOptions{},
		ExternalTables:               map[string]TableDetails{},
		ExternalTableOptions:         map[string]TableOptions{},
		MaterializedViews:            map[string]TableDetails{},
		MaterializedViewOptions:      map[string]TableOptions{},
		MaterializedViewQueries:      map[string]string{},
		StoredQueryResults:           map[string]TableDetails{},
		StoredQueryResultOptions:     map[string]TableOptions{},
		GraphModels:                  map[string]GraphDetails{},
		EntityGroups:                 map[string][]string{},
		MagicFunctions:               append([]string{}, m365.MagicFunctions...),
		TabularFunctions:             copyTabularFunctions(m365.TabularFunctions),
		ScalarFunctions:              copyScalarFunctions(m365.ScalarFunctions),
		ExternalTableColumnOrder:     map[string][]string{},
		MaterializedViewColumnOrder:  map[string][]string{},
		StoredQueryResultColumnOrder: map[string][]string{},
	}

	for name, table := range m365.TableDetails {
		merged.TableDetails[name] = copyTableDetails(table)
		if options, ok := tableOptionsForName(m365.TableOptions, name); ok {
			merged.TableOptions[name] = copyTableOptions(options)
		}
	}
	for _, name := range mergeOrderedNames(nil, m365.ExternalTableOrder, m365.ExternalTables) {
		table := m365.ExternalTables[name]
		merged.ExternalTables[name] = copyTableDetails(table)
		merged.ExternalTableOrder = append(merged.ExternalTableOrder, name)
		if merged.ExternalTableColumnOrder == nil {
			merged.ExternalTableColumnOrder = map[string][]string{}
		}
		merged.ExternalTableColumnOrder[name] = append([]string{}, m365.ExternalTableColumnOrder[name]...)
		if options, ok := tableOptionsForName(m365.ExternalTableOptions, name); ok {
			merged.ExternalTableOptions[name] = copyTableOptions(options)
		}
	}
	for _, name := range mergeOrderedNames(nil, m365.MaterializedViewOrder, m365.MaterializedViews) {
		table := m365.MaterializedViews[name]
		merged.MaterializedViews[name] = copyTableDetails(table)
		merged.MaterializedViewOrder = append(merged.MaterializedViewOrder, name)
		if merged.MaterializedViewColumnOrder == nil {
			merged.MaterializedViewColumnOrder = map[string][]string{}
		}
		merged.MaterializedViewColumnOrder[name] = append([]string{}, m365.MaterializedViewColumnOrder[name]...)
		if options, ok := tableOptionsForName(m365.MaterializedViewOptions, name); ok {
			merged.MaterializedViewOptions[name] = copyTableOptions(options)
		}
	}
	for name, query := range m365.MaterializedViewQueries {
		merged.MaterializedViewQueries[name] = query
	}
	for _, name := range mergeOrderedNames(nil, m365.StoredQueryResultOrder, m365.StoredQueryResults) {
		table := m365.StoredQueryResults[name]
		merged.StoredQueryResults[name] = copyTableDetails(table)
		merged.StoredQueryResultOrder = append(merged.StoredQueryResultOrder, name)
		if merged.StoredQueryResultColumnOrder == nil {
			merged.StoredQueryResultColumnOrder = map[string][]string{}
		}
		merged.StoredQueryResultColumnOrder[name] = append([]string{}, m365.StoredQueryResultColumnOrder[name]...)
		if options, ok := tableOptionsForName(m365.StoredQueryResultOptions, name); ok {
			merged.StoredQueryResultOptions[name] = copyTableOptions(options)
		}
	}
	for name, graph := range m365.GraphModels {
		merged.GraphModels[name] = copyGraphDetails(graph)
	}
	for name, members := range m365.EntityGroups {
		merged.EntityGroups[name] = append([]string{}, members...)
	}
	for name, table := range sentinel.TableDetails {
		if _, exists := merged.TableDetails[name]; !exists {
			merged.TableDetails[name] = copyTableDetails(table)
			if options, ok := tableOptionsForName(sentinel.TableOptions, name); ok {
				merged.TableOptions[name] = copyTableOptions(options)
			}
		} else {
			// Shared tables expose both schemas. Keep m365 types for shared
			// columns, but retain Sentinel-only columns such as Type.
			sentinelOptions, _ := tableOptionsForName(sentinel.TableOptions, name)
			options := merged.TableOptions[name]
			for column, typ := range table {
				if _, exists := merged.TableDetails[name][column]; exists {
					continue
				}
				merged.TableDetails[name][column] = normalizeType(typ)
				if details, ok := columnTypeDetailsForName(sentinelOptions.ColumnTypes, column); ok {
					if options.ColumnTypes == nil {
						options.ColumnTypes = map[string]ColumnTypeDetails{}
					}
					options.ColumnTypes[column] = copyColumnTypeDetails(details)
					merged.TableOptions[name] = options
				}
			}
		}
	}
	for _, name := range mergeOrderedNames(nil, sentinel.ExternalTableOrder, sentinel.ExternalTables) {
		table := sentinel.ExternalTables[name]
		if _, exists := merged.ExternalTables[name]; !exists {
			merged.ExternalTables[name] = copyTableDetails(table)
			merged.ExternalTableOrder = append(merged.ExternalTableOrder, name)
			merged.ExternalTableColumnOrder[name] = append([]string{}, sentinel.ExternalTableColumnOrder[name]...)
			if options, ok := tableOptionsForName(sentinel.ExternalTableOptions, name); ok {
				merged.ExternalTableOptions[name] = copyTableOptions(options)
			}
		}
	}
	for _, name := range mergeOrderedNames(nil, sentinel.MaterializedViewOrder, sentinel.MaterializedViews) {
		table := sentinel.MaterializedViews[name]
		if _, exists := merged.MaterializedViews[name]; !exists {
			merged.MaterializedViews[name] = copyTableDetails(table)
			merged.MaterializedViewOrder = append(merged.MaterializedViewOrder, name)
			merged.MaterializedViewColumnOrder[name] = append([]string{}, sentinel.MaterializedViewColumnOrder[name]...)
			if options, ok := tableOptionsForName(sentinel.MaterializedViewOptions, name); ok {
				merged.MaterializedViewOptions[name] = copyTableOptions(options)
			}
			if query, ok := sentinel.MaterializedViewQueries[name]; ok {
				merged.MaterializedViewQueries[name] = query
			}
		}
	}
	for _, name := range mergeOrderedNames(nil, sentinel.StoredQueryResultOrder, sentinel.StoredQueryResults) {
		table := sentinel.StoredQueryResults[name]
		if _, exists := merged.StoredQueryResults[name]; !exists {
			merged.StoredQueryResults[name] = copyTableDetails(table)
			merged.StoredQueryResultOrder = append(merged.StoredQueryResultOrder, name)
			merged.StoredQueryResultColumnOrder[name] = append([]string{}, sentinel.StoredQueryResultColumnOrder[name]...)
			if options, ok := tableOptionsForName(sentinel.StoredQueryResultOptions, name); ok {
				merged.StoredQueryResultOptions[name] = copyTableOptions(options)
			}
		}
	}
	for name, graph := range sentinel.GraphModels {
		if _, exists := merged.GraphModels[name]; !exists {
			merged.GraphModels[name] = copyGraphDetails(graph)
		}
	}
	for name, members := range sentinel.EntityGroups {
		if _, exists := merged.EntityGroups[name]; !exists {
			merged.EntityGroups[name] = append([]string{}, members...)
		}
	}
	for name, function := range sentinel.TabularFunctions {
		if _, exists := merged.TabularFunctions[name]; !exists {
			merged.TabularFunctions[name] = copyTabularFunctionDetails(function)
		}
	}
	for name, function := range sentinel.ScalarFunctions {
		if _, exists := merged.ScalarFunctions[name]; !exists {
			merged.ScalarFunctions[name] = copyScalarFunctionDetails(function)
		}
	}
	envs["m365_with_sentinel"] = merged
}

func validateEntityGroupMembers(env *EnvironmentDefinition) error {
	if env == nil {
		return nil
	}
	for _, groupName := range sortedEntityGroupNames(env.EntityGroups) {
		for _, tableName := range env.EntityGroups[groupName] {
			if _, ok := env.TableDetails[tableName]; !ok {
				return fmt.Errorf("entity group %q references unknown table %q", groupName, tableName)
			}
		}
	}
	return nil
}

func validateDatabaseEntityGroupMembers(databaseName string, database *DatabaseDetails) error {
	if database == nil {
		return nil
	}
	for _, groupName := range sortedEntityGroupNames(database.EntityGroups) {
		for _, tableName := range database.EntityGroups[groupName] {
			if _, ok := database.TableDetails[tableName]; !ok {
				return fmt.Errorf("database %q entity group %q references unknown table %q", databaseName, groupName, tableName)
			}
		}
	}
	return nil
}

func configuredCurrentClusterName(env *EnvironmentDefinition) (string, bool) {
	if env == nil || len(env.Clusters) == 0 {
		return "", false
	}
	name := strings.TrimSpace(env.CurrentCluster)
	if name == "" && len(env.Clusters) == 1 {
		for candidate := range env.Clusters {
			return candidate, true
		}
	}
	for candidate := range env.Clusters {
		if strings.EqualFold(candidate, name) {
			return candidate, true
		}
	}
	return "", false
}

func configuredCurrentDatabaseName(env *EnvironmentDefinition, clusterName string) string {
	if env != nil && strings.TrimSpace(env.CurrentDatabase) != "" {
		requested := strings.TrimSpace(env.CurrentDatabase)
		if cluster, ok := env.Clusters[clusterName]; ok {
			for candidate, database := range cluster.Databases {
				if strings.EqualFold(candidate, requested) || strings.EqualFold(database.AlternateName, requested) {
					return candidate
				}
			}
		}
		return requested
	}
	if env != nil {
		if cluster, ok := env.Clusters[clusterName]; ok && len(cluster.Databases) == 1 {
			for candidate := range cluster.Databases {
				return candidate
			}
		}
	}
	return "db"
}

func databaseDetailsEnvironment(database DatabaseDetails) *EnvironmentDefinition {
	return &EnvironmentDefinition{
		TableDetails:                 copyTableDetailsMap(database.TableDetails),
		TableOptions:                 copyTableOptionsMap(database.TableOptions),
		ExternalTables:               copyTableDetailsMap(database.ExternalTables),
		ExternalTableOptions:         copyTableOptionsMap(database.ExternalTableOptions),
		MaterializedViews:            copyTableDetailsMap(database.MaterializedViews),
		MaterializedViewOptions:      copyTableOptionsMap(database.MaterializedViewOptions),
		MaterializedViewQueries:      copyStringMap(database.MaterializedViewQueries),
		StoredQueryResults:           copyTableDetailsMap(database.StoredQueryResults),
		StoredQueryResultOptions:     copyTableOptionsMap(database.StoredQueryResultOptions),
		GraphModels:                  copyGraphDetailsMap(database.GraphModels),
		EntityGroups:                 copyEntityGroups(database.EntityGroups),
		MagicFunctions:               append([]string{}, database.MagicFunctions...),
		TabularFunctions:             copyTabularFunctions(database.TabularFunctions),
		ScalarFunctions:              copyScalarFunctions(database.ScalarFunctions),
		TableOrder:                   append([]string{}, database.TableOrder...),
		TableColumnOrder:             copyStringSliceMap(database.TableColumnOrder),
		ExternalTableOrder:           append([]string{}, database.ExternalTableOrder...),
		ExternalTableColumnOrder:     copyStringSliceMap(database.ExternalTableColumnOrder),
		MaterializedViewOrder:        append([]string{}, database.MaterializedViewOrder...),
		MaterializedViewColumnOrder:  copyStringSliceMap(database.MaterializedViewColumnOrder),
		StoredQueryResultOrder:       append([]string{}, database.StoredQueryResultOrder...),
		StoredQueryResultColumnOrder: copyStringSliceMap(database.StoredQueryResultColumnOrder),
	}
}

func currentEnvironmentDefinition(env *EnvironmentDefinition) *EnvironmentDefinition {
	if env == nil {
		return nil
	}
	current := &EnvironmentDefinition{}
	if clusterName, ok := configuredCurrentClusterName(env); ok {
		databaseName := configuredCurrentDatabaseName(env, clusterName)
		cluster := env.Clusters[clusterName]
		for candidate, database := range cluster.Databases {
			if strings.EqualFold(candidate, databaseName) {
				current = databaseDetailsEnvironment(database)
				break
			}
		}
	}
	mergeFlatEnvironment(current, env)
	normalizeFlatEnvironment(current)
	// Add server-side aliases only after merging database and root declarations,
	// so explicitly configured time-column types take precedence.
	allowDefenderTimeAliases(current.TableDetails)
	// Catalog selections are validated during normalization. Add defaults only
	// after merging the current database, so its workspace functions win too.
	_ = addFunctionCatalogs(current)
	return current
}

func mergeFlatEnvironment(target, overlay *EnvironmentDefinition) {
	if target == nil || overlay == nil {
		return
	}
	target.FunctionCatalogs = append([]string{}, overlay.FunctionCatalogs...)
	resetOverlaidTableOptions(&target.TableOptions, overlay.TableDetails)
	mergeTableDetailsMap(&target.TableDetails, overlay.TableDetails)
	mergeTableOptionsMap(&target.TableOptions, overlay.TableOptions)
	mergeTableOrder(target, overlay)
	resetOverlaidTableOptions(&target.ExternalTableOptions, overlay.ExternalTables)
	mergeTableDetailsMap(&target.ExternalTables, overlay.ExternalTables)
	mergeTableOptionsMap(&target.ExternalTableOptions, overlay.ExternalTableOptions)
	mergeNamedTableOrder(&target.ExternalTableOrder, &target.ExternalTableColumnOrder, overlay.ExternalTableOrder, overlay.ExternalTableColumnOrder, target.ExternalTables)
	resetOverlaidTableOptions(&target.MaterializedViewOptions, overlay.MaterializedViews)
	mergeMaterializedViews(target, overlay)
	mergeTableOptionsMap(&target.MaterializedViewOptions, overlay.MaterializedViewOptions)
	mergeNamedTableOrder(&target.MaterializedViewOrder, &target.MaterializedViewColumnOrder, overlay.MaterializedViewOrder, overlay.MaterializedViewColumnOrder, target.MaterializedViews)
	resetOverlaidTableOptions(&target.StoredQueryResultOptions, overlay.StoredQueryResults)
	mergeTableDetailsMap(&target.StoredQueryResults, overlay.StoredQueryResults)
	mergeTableOptionsMap(&target.StoredQueryResultOptions, overlay.StoredQueryResultOptions)
	mergeNamedTableOrder(&target.StoredQueryResultOrder, &target.StoredQueryResultColumnOrder, overlay.StoredQueryResultOrder, overlay.StoredQueryResultColumnOrder, target.StoredQueryResults)
	if target.GraphModels == nil {
		target.GraphModels = map[string]GraphDetails{}
	}
	for name, graph := range overlay.GraphModels {
		target.GraphModels[name] = copyGraphDetails(graph)
	}
	if target.EntityGroups == nil {
		target.EntityGroups = map[string][]string{}
	}
	for name, members := range overlay.EntityGroups {
		target.EntityGroups[name] = append([]string{}, members...)
	}
	if len(overlay.MagicFunctions) > 0 {
		target.MagicFunctions = append([]string{}, overlay.MagicFunctions...)
	}
	if target.TabularFunctions == nil {
		target.TabularFunctions = map[string]TabularFunctionDetails{}
	}
	for name, function := range overlay.TabularFunctions {
		target.TabularFunctions[name] = copyTabularFunctionDetails(function)
	}
	if target.ScalarFunctions == nil {
		target.ScalarFunctions = map[string]ScalarFunctionDetails{}
	}
	for name, function := range overlay.ScalarFunctions {
		target.ScalarFunctions[name] = copyScalarFunctionDetails(function)
	}
}

func resetOverlaidTableOptions(target *map[string]TableOptions, overlaidTables map[string]TableDetails) {
	if len(*target) == 0 || len(overlaidTables) == 0 {
		return
	}
	for tableName := range overlaidTables {
		for optionName := range *target {
			if strings.EqualFold(optionName, tableName) {
				delete(*target, optionName)
			}
		}
	}
}

func mergeTableDetailsMap(target *map[string]TableDetails, overlay map[string]TableDetails) {
	if *target == nil {
		*target = map[string]TableDetails{}
	}
	for name, table := range overlay {
		(*target)[name] = copyTableDetails(table)
	}
}

func mergeTableOptionsMap(target *map[string]TableOptions, overlay map[string]TableOptions) {
	if *target == nil {
		*target = map[string]TableOptions{}
	}
	for name, options := range overlay {
		(*target)[name] = copyTableOptions(options)
	}
}

func mergeTableOrder(target, overlay *EnvironmentDefinition) {
	if target == nil || overlay == nil {
		return
	}
	target.TableOrder = mergeOrderedNames(target.TableOrder, overlay.TableOrder, target.TableDetails)
	if target.TableColumnOrder == nil {
		target.TableColumnOrder = map[string][]string{}
	}
	for name, order := range overlay.TableColumnOrder {
		target.TableColumnOrder[name] = append([]string{}, order...)
	}
}

func mergeNamedTableOrder(targetOrder *[]string, targetColumns *map[string][]string, overlayOrder []string, overlayColumns map[string][]string, values map[string]TableDetails) {
	*targetOrder = mergeOrderedNames(*targetOrder, overlayOrder, values)
	if *targetColumns == nil {
		*targetColumns = map[string][]string{}
	}
	for name, order := range overlayColumns {
		(*targetColumns)[name] = append([]string{}, order...)
	}
}

func mergeMaterializedViews(target, overlay *EnvironmentDefinition) {
	if target.MaterializedViews == nil {
		target.MaterializedViews = map[string]TableDetails{}
	}
	if target.MaterializedViewQueries == nil {
		target.MaterializedViewQueries = map[string]string{}
	}
	for name, table := range overlay.MaterializedViews {
		target.MaterializedViews[name] = copyTableDetails(table)
		delete(target.MaterializedViewQueries, name)
		if query, ok := overlay.MaterializedViewQueries[name]; ok {
			target.MaterializedViewQueries[name] = query
		}
	}
}

func copyStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for name, value := range in {
		out[name] = value
	}
	return out
}

func copyStringSliceMap(in map[string][]string) map[string][]string {
	out := map[string][]string{}
	for name, values := range in {
		out[name] = append([]string{}, values...)
	}
	return out
}

func normalizeTableOptions(tables map[string]TableDetails, options *map[string]TableOptions) error {
	return normalizeTableOptionsForKind("table_options", "table", tables, options, true)
}

func normalizeEntityTableOptions(optionField, entityKind string, tables map[string]TableDetails, options *map[string]TableOptions, allowState bool) error {
	return normalizeTableOptionsForKind(optionField, entityKind, tables, options, allowState)
}

func normalizeTableOptionsForKind(optionField, entityKind string, tables map[string]TableDetails, options *map[string]TableOptions, allowState bool) error {
	if *options == nil {
		*options = map[string]TableOptions{}
		return nil
	}
	tableNames := map[string][]string{}
	for name := range tables {
		folded := foldName(name)
		tableNames[folded] = append(tableNames[folded], name)
	}
	normalized := make(map[string]TableOptions, len(*options))
	metadataNames := map[string]string{}
	for optionName, value := range *options {
		folded := foldName(optionName)
		matches := tableNames[folded]
		if _, exact := tables[optionName]; exact {
			matches = []string{optionName}
		}
		if len(matches) == 0 {
			return fmt.Errorf("%s entry %q references unknown %s", optionField, optionName, entityKind)
		}
		if len(matches) > 1 {
			sort.Strings(matches)
			return fmt.Errorf("%s entry %q is ambiguous between %ss %q and %q", optionField, optionName, entityKind, matches[0], matches[1])
		}
		if previous := metadataNames[matches[0]]; previous != "" {
			return fmt.Errorf("%s entries %q and %q both reference %s %q", optionField, previous, optionName, entityKind, matches[0])
		}
		if !allowState && (value.IsOpen || value.IsSerialized || value.IsSorted) {
			return fmt.Errorf("%s entry %q cannot set table state; configured %ss do not preserve open, serialized, or sorted state", optionField, optionName, entityKind)
		}
		metadataNames[matches[0]] = optionName
		columnTypes, err := normalizeColumnTypeDetailsMap(fmt.Sprintf("%s entry %q", optionField, matches[0]), tables[matches[0]], value.ColumnTypes)
		if err != nil {
			return err
		}
		value.ColumnTypes = columnTypes
		normalized[matches[0]] = value
	}
	*options = normalized
	return nil
}

func normalizeColumnTypeDetailsMap(owner string, columns TableDetails, details map[string]ColumnTypeDetails) (map[string]ColumnTypeDetails, error) {
	if details == nil {
		return nil, nil
	}
	columnNames := make(map[string][]string, len(columns))
	for name := range columns {
		folded := foldName(name)
		columnNames[folded] = append(columnNames[folded], name)
	}
	normalized := make(map[string]ColumnTypeDetails, len(details))
	metadataNames := make(map[string]string, len(details))
	for _, metadataName := range sortedColumnTypeDetailsNames(details) {
		folded := foldName(metadataName)
		matches := columnNames[folded]
		if _, exact := columns[metadataName]; exact {
			matches = []string{metadataName}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("%s column_types entry %q references unknown column", owner, metadataName)
		}
		if len(matches) > 1 {
			sort.Strings(matches)
			return nil, fmt.Errorf("%s column_types entry %q is ambiguous between columns %q and %q", owner, metadataName, matches[0], matches[1])
		}
		if previous := metadataNames[matches[0]]; previous != "" {
			return nil, fmt.Errorf("%s column_types entries %q and %q both reference column %q", owner, previous, metadataName, matches[0])
		}
		columnName := matches[0]
		if normalizeType(columns[columnName]) != kustolang.ScalarTypes.Dynamic.SymbolName() {
			return nil, fmt.Errorf("%s column_types entry %q requires a dynamic table column", owner, metadataName)
		}
		value, err := normalizeColumnTypeDetails(details[metadataName], fmt.Sprintf("%s column_types entry %q", owner, metadataName))
		if err != nil {
			return nil, err
		}
		metadataNames[columnName] = metadataName
		normalized[columnName] = value
	}
	return normalized, nil
}

const maxColumnTypeDetailsDepth = 64

func normalizeColumnTypeDetails(details ColumnTypeDetails, path string) (ColumnTypeDetails, error) {
	return normalizeColumnTypeDetailsAtDepth(details, path, 0, map[*ColumnTypeDetails]bool{})
}

func normalizeColumnTypeDetailsAtDepth(details ColumnTypeDetails, path string, depth int, stack map[*ColumnTypeDetails]bool) (ColumnTypeDetails, error) {
	if depth > maxColumnTypeDetailsDepth {
		return ColumnTypeDetails{}, fmt.Errorf("%s exceeds maximum nesting depth %d", path, maxColumnTypeDetailsDepth)
	}
	typeName := strings.ToLower(strings.TrimSpace(details.Type))
	switch typeName {
	case "bag":
		if details.Element != nil {
			return ColumnTypeDetails{}, fmt.Errorf("%s bag cannot specify element", path)
		}
		if details.Properties == nil {
			details.Properties = map[string]ColumnTypeDetails{}
		}
		details.Type = "bag"
	case "array":
		if details.Properties != nil {
			return ColumnTypeDetails{}, fmt.Errorf("%s array cannot specify properties", path)
		}
		if len(details.PropertyOrder) > 0 {
			return ColumnTypeDetails{}, fmt.Errorf("%s array cannot specify property order", path)
		}
		if details.Element == nil {
			return ColumnTypeDetails{}, fmt.Errorf("%s array must specify element", path)
		}
		details.Type = "array"
	default:
		if typeName == "" {
			return ColumnTypeDetails{}, fmt.Errorf("%s must specify type", path)
		}
		if details.Properties != nil || details.Element != nil || len(details.PropertyOrder) > 0 {
			return ColumnTypeDetails{}, fmt.Errorf("%s scalar type cannot specify properties or element", path)
		}
		symbol := kustolang.ScalarTypes.Get(typeName)
		if symbol == nil {
			return ColumnTypeDetails{}, fmt.Errorf("%s has unknown scalar type %q", path, typeName)
		}
		if !supportedDynamicShapeScalar(symbol) {
			return ColumnTypeDetails{}, fmt.Errorf("%s scalar type %q cannot be represented exactly inside dynamic", path, symbol.SymbolName())
		}
		details.Type = symbol.SymbolName()
	}
	if details.Type == "bag" {
		normalized := make(map[string]ColumnTypeDetails, len(details.Properties))
		propertyNames, err := normalizeColumnTypeDetailsPropertyOrder(details.Properties, details.PropertyOrder, path)
		if err != nil {
			return ColumnTypeDetails{}, err
		}
		for _, name := range propertyNames {
			value, err := normalizeColumnTypeDetailsAtDepth(details.Properties[name], fmt.Sprintf("%s property %q", path, name), depth+1, stack)
			if err != nil {
				return ColumnTypeDetails{}, err
			}
			normalized[name] = value
		}
		details.Properties = normalized
		details.PropertyOrder = propertyNames
	}
	if details.Type == "array" {
		if stack[details.Element] {
			return ColumnTypeDetails{}, fmt.Errorf("%s contains a recursive element type", path)
		}
		stack[details.Element] = true
		element, err := normalizeColumnTypeDetailsAtDepth(*details.Element, path+" element", depth+1, stack)
		delete(stack, details.Element)
		if err != nil {
			return ColumnTypeDetails{}, err
		}
		details.Element = &element
	}
	return details, nil
}

func normalizeColumnTypeDetailsPropertyOrder(properties map[string]ColumnTypeDetails, order []string, path string) ([]string, error) {
	if len(order) == 0 {
		return sortedColumnTypeDetailsNames(properties), nil
	}
	if len(order) != len(properties) {
		return nil, fmt.Errorf("%s property order must name every property exactly once", path)
	}
	seen := make(map[string]struct{}, len(order))
	for _, name := range order {
		if _, ok := properties[name]; !ok {
			return nil, fmt.Errorf("%s property order references unknown property %q", path, name)
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, fmt.Errorf("%s property order repeats property %q", path, name)
		}
		seen[name] = struct{}{}
	}
	return append([]string(nil), order...), nil
}

func supportedDynamicShapeScalar(symbol *kustolang.ScalarSymbol) bool {
	if symbol == nil {
		return false
	}
	switch symbol.SymbolName() {
	case "dynamic", "bool", "int", "long", "real", "datetime", "timespan", "guid", "string":
		return true
	default:
		return false
	}
}

func normalizeConfiguredFunctions(owner string, scalarFunctions map[string]ScalarFunctionDetails, tabularFunctions map[string]TabularFunctionDetails) error {
	for _, name := range sortedScalarFunctionNames(scalarFunctions) {
		function := scalarFunctions[name]
		function.OutputType = normalizeType(function.OutputType)
		if function.OutputTypeDetails != nil {
			if function.OutputType != kustolang.ScalarTypes.Dynamic.SymbolName() {
				return fmt.Errorf("%s scalar function %q output_type_details requires output_type dynamic", owner, name)
			}
			normalized, err := normalizeColumnTypeDetails(*function.OutputTypeDetails, fmt.Sprintf("%s scalar function %q output_type_details", owner, name))
			if err != nil {
				return err
			}
			function.OutputTypeDetails = &normalized
		}
		arguments, err := normalizeConfiguredFunctionArguments(owner, name, function.Arguments)
		if err != nil {
			return err
		}
		function.Arguments = arguments
		scalarFunctions[name] = function
	}
	for _, name := range sortedTabularFunctionNames(tabularFunctions) {
		function := tabularFunctions[name]
		function.OutputColumns = copyTableDetails(function.OutputColumns)
		columnTypes, err := normalizeColumnTypeDetailsMap(fmt.Sprintf("%s tabular function %q", owner, name), function.OutputColumns, function.OutputColumnTypes)
		if err != nil {
			return err
		}
		function.OutputColumnTypes = columnTypes
		arguments, err := normalizeConfiguredFunctionArguments(owner, name, function.Arguments)
		if err != nil {
			return err
		}
		function.Arguments = arguments
		tabularFunctions[name] = function
	}
	return nil
}

func normalizeConfiguredFunctionArguments(owner, functionName string, arguments []ArgumentDetails) ([]ArgumentDetails, error) {
	normalized := append([]ArgumentDetails{}, arguments...)
	for i := range normalized {
		argument := normalized[i]
		argument.Type = normalizeType(argument.Type)
		if argument.TypeDetails != nil {
			if argument.Type != kustolang.ScalarTypes.Dynamic.SymbolName() {
				return nil, fmt.Errorf("%s function %q argument %q type_details requires type dynamic", owner, functionName, argument.Name)
			}
			details, err := normalizeColumnTypeDetails(*argument.TypeDetails, fmt.Sprintf("%s function %q argument %q type_details", owner, functionName, argument.Name))
			if err != nil {
				return nil, err
			}
			argument.TypeDetails = &details
		}
		normalized[i] = argument
	}
	return normalized, nil
}

func sortedColumnTypeDetailsNames(details map[string]ColumnTypeDetails) []string {
	names := make([]string, 0, len(details))
	for name := range details {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func orderedColumnTypeDetailsNames(details ColumnTypeDetails) []string {
	return mergeOrderedNames(nil, details.PropertyOrder, details.Properties)
}

func normalizeMaterializedViewQueries(views map[string]TableDetails, queries *map[string]string) error {
	if *queries == nil {
		*queries = map[string]string{}
		return nil
	}
	viewNames := map[string][]string{}
	for name := range views {
		folded := foldName(name)
		viewNames[folded] = append(viewNames[folded], name)
	}
	normalized := make(map[string]string, len(*queries))
	metadataNames := map[string]string{}
	for _, metadataName := range sortedStringMapNames(*queries) {
		folded := foldName(metadataName)
		matches := viewNames[folded]
		if _, exact := views[metadataName]; exact {
			matches = []string{metadataName}
		}
		if len(matches) == 0 {
			return fmt.Errorf("materialized_view_queries entry %q references unknown materialized view", metadataName)
		}
		if len(matches) > 1 {
			sort.Strings(matches)
			return fmt.Errorf("materialized_view_queries entry %q is ambiguous between materialized views %q and %q", metadataName, matches[0], matches[1])
		}
		if previous := metadataNames[matches[0]]; previous != "" {
			return fmt.Errorf("materialized_view_queries entries %q and %q both reference materialized view %q", previous, metadataName, matches[0])
		}
		metadataNames[matches[0]] = metadataName
		normalized[matches[0]] = (*queries)[metadataName]
	}
	*queries = normalized
	return nil
}

func sortedStringMapNames(values map[string]string) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func copyTableDetailsMap(in map[string]TableDetails) map[string]TableDetails {
	out := map[string]TableDetails{}
	for name, table := range in {
		out[name] = copyTableDetails(table)
	}
	return out
}

func copyTableOptionsMap(in map[string]TableOptions) map[string]TableOptions {
	out := map[string]TableOptions{}
	for name, options := range in {
		out[name] = copyTableOptions(options)
	}
	return out
}

func copyTableOptions(options TableOptions) TableOptions {
	options.ColumnTypes = copyColumnTypeDetailsMap(options.ColumnTypes)
	return options
}

func copyColumnTypeDetailsMap(in map[string]ColumnTypeDetails) map[string]ColumnTypeDetails {
	if in == nil {
		return nil
	}
	out := make(map[string]ColumnTypeDetails, len(in))
	for name, details := range in {
		out[name] = copyColumnTypeDetails(details)
	}
	return out
}

func copyColumnTypeDetails(in ColumnTypeDetails) ColumnTypeDetails {
	return copyColumnTypeDetailsSeen(in, map[*ColumnTypeDetails]*ColumnTypeDetails{})
}

func copyColumnTypeDetailsSeen(in ColumnTypeDetails, seen map[*ColumnTypeDetails]*ColumnTypeDetails) ColumnTypeDetails {
	out := in
	out.PropertyOrder = append([]string(nil), in.PropertyOrder...)
	if in.Properties != nil {
		out.Properties = make(map[string]ColumnTypeDetails, len(in.Properties))
		for name, property := range in.Properties {
			out.Properties[name] = copyColumnTypeDetailsSeen(property, seen)
		}
	}
	if in.Element != nil {
		if existing := seen[in.Element]; existing != nil {
			out.Element = existing
		} else {
			element := &ColumnTypeDetails{}
			seen[in.Element] = element
			*element = copyColumnTypeDetailsSeen(*in.Element, seen)
			out.Element = element
		}
	}
	return out
}

func tableOptionsForName(options map[string]TableOptions, name string) (TableOptions, bool) {
	if value, ok := options[name]; ok {
		return value, true
	}
	for candidate, value := range options {
		if strings.EqualFold(candidate, name) {
			return value, true
		}
	}
	return TableOptions{}, false
}

func tableIsOpen(options map[string]TableOptions, name string) bool {
	value, ok := tableOptionsForName(options, name)
	return ok && value.IsOpen
}

func copyGraphDetailsMap(in map[string]GraphDetails) map[string]GraphDetails {
	out := map[string]GraphDetails{}
	for name, graph := range in {
		out[name] = copyGraphDetails(graph)
	}
	return out
}

func copyEntityGroups(in map[string][]string) map[string][]string {
	out := map[string][]string{}
	for name, members := range in {
		out[name] = append([]string{}, members...)
	}
	return out
}

func copyTabularFunctions(in map[string]TabularFunctionDetails) map[string]TabularFunctionDetails {
	out := map[string]TabularFunctionDetails{}
	for name, function := range in {
		out[name] = copyTabularFunctionDetails(function)
	}
	return out
}

func copyScalarFunctions(in map[string]ScalarFunctionDetails) map[string]ScalarFunctionDetails {
	out := map[string]ScalarFunctionDetails{}
	for name, function := range in {
		out[name] = copyScalarFunctionDetails(function)
	}
	return out
}

func copyTabularFunctionDetails(function TabularFunctionDetails) TabularFunctionDetails {
	function.OutputColumns = copyTableDetails(function.OutputColumns)
	function.OutputColumnTypes = copyColumnTypeDetailsMap(function.OutputColumnTypes)
	function.Arguments = copyArgumentDetails(function.Arguments)
	return function
}

func copyScalarFunctionDetails(function ScalarFunctionDetails) ScalarFunctionDetails {
	if function.OutputTypeDetails != nil {
		details := copyColumnTypeDetails(*function.OutputTypeDetails)
		function.OutputTypeDetails = &details
	}
	function.Arguments = copyArgumentDetails(function.Arguments)
	return function
}

func copyArgumentDetails(arguments []ArgumentDetails) []ArgumentDetails {
	out := append([]ArgumentDetails{}, arguments...)
	for i := range out {
		if out[i].TypeDetails != nil {
			details := copyColumnTypeDetails(*out[i].TypeDetails)
			out[i].TypeDetails = &details
		}
	}
	return out
}

func copyTableDetails(in TableDetails) TableDetails {
	out := TableDetails{}
	for k, v := range in {
		out[k] = normalizeType(v)
	}
	return out
}

func copyGraphDetails(in GraphDetails) GraphDetails {
	out := GraphDetails{
		Edges:     map[string]TableDetails{},
		Nodes:     map[string]TableDetails{},
		Snapshots: append([]string{}, in.Snapshots...),
	}
	for name, table := range in.Edges {
		out.Edges[name] = copyTableDetails(table)
	}
	for name, table := range in.Nodes {
		out.Nodes[name] = copyTableDetails(table)
	}
	return out
}

type Catalog struct {
	Tables             map[string]TableSchema
	TableNames         map[string]string
	ImpliedTables      []TableSchema
	ExternalTables     map[string]TableSchema
	ExternalTableNames map[string]string
	MaterializedViews  map[string]TableSchema
	MaterializedNames  map[string]string
	StoredQueryResults map[string]TableSchema
	StoredResultNames  map[string]string
	GraphModels        map[string]GraphShapeDetails
	GraphModelNames    map[string]string
	EntityGroups       map[string][]TableSchema
	EntityGroupMembers map[string][]string
	EntityGroupNames   map[string]string

	ScalarFunctions  map[string]ScalarFunctionDetails
	TabularFunctions map[string]TabularFunctionDetails
	MagicFunctions   map[string]bool
	Watchlists       map[string]WatchlistDetails
	Globals          *kustolang.GlobalState
	Binder           *kustolang.Binder
	DatabaseCatalogs map[*kustolang.DatabaseSymbol]*Catalog
}

type GraphShapeDetails struct {
	Edges      TableSchema
	Nodes      TableSchema
	Snapshots  map[string]string
	ModelNames []string
}

func NewCatalog(env *EnvironmentDefinition, local LocalData) (*Catalog, error) {
	if err := normalizeLocalData(&local); err != nil {
		return nil, err
	}
	if err := normalizeEnvironmentDefinition(env); err != nil {
		return nil, err
	}
	currentEnv := currentEnvironmentDefinition(env)
	c := &Catalog{
		Tables:             map[string]TableSchema{},
		TableNames:         map[string]string{},
		ExternalTables:     map[string]TableSchema{},
		ExternalTableNames: map[string]string{},
		MaterializedViews:  map[string]TableSchema{},
		MaterializedNames:  map[string]string{},
		StoredQueryResults: map[string]TableSchema{},
		StoredResultNames:  map[string]string{},
		GraphModels:        map[string]GraphShapeDetails{},
		GraphModelNames:    map[string]string{},
		EntityGroups:       map[string][]TableSchema{},
		EntityGroupMembers: map[string][]string{},
		EntityGroupNames:   map[string]string{},
		ScalarFunctions:    map[string]ScalarFunctionDetails{},
		TabularFunctions:   map[string]TabularFunctionDetails{},
		MagicFunctions:     map[string]bool{},
		Watchlists:         local.Watchlists,
		DatabaseCatalogs:   map[*kustolang.DatabaseSymbol]*Catalog{},
	}

	if env != nil {
		if err := validateEntityGroupMembers(env); err != nil {
			return c, err
		}
	}
	if currentEnv != nil {
		if err := validateEntityGroupMembers(currentEnv); err != nil {
			return c, err
		}
		for _, name := range mergeOrderedNames(nil, currentEnv.TableOrder, currentEnv.TableDetails) {
			options, _ := tableOptionsForName(currentEnv.TableOptions, name)
			c.addTableOrdered(name, currentEnv.TableDetails[name], currentEnv.TableColumnOrder[name], options, true)
			c.ImpliedTables = append(c.ImpliedTables, c.Tables[name])
		}
		for _, name := range mergeOrderedNames(nil, currentEnv.ExternalTableOrder, currentEnv.ExternalTables) {
			table := currentEnv.ExternalTables[name]
			options, _ := tableOptionsForName(currentEnv.ExternalTableOptions, name)
			c.addExternalTable(name, table, currentEnv.ExternalTableColumnOrder[name], options, true)
		}
		for _, name := range mergeOrderedNames(nil, currentEnv.MaterializedViewOrder, currentEnv.MaterializedViews) {
			table := currentEnv.MaterializedViews[name]
			options, _ := tableOptionsForName(currentEnv.MaterializedViewOptions, name)
			c.addMaterializedView(name, table, currentEnv.MaterializedViewColumnOrder[name], options, true)
		}
		for _, name := range mergeOrderedNames(nil, currentEnv.StoredQueryResultOrder, currentEnv.StoredQueryResults) {
			table := currentEnv.StoredQueryResults[name]
			options, _ := tableOptionsForName(currentEnv.StoredQueryResultOptions, name)
			c.addStoredQueryResult(name, table, currentEnv.StoredQueryResultColumnOrder[name], options, true)
		}
		for name, graph := range currentEnv.GraphModels {
			c.addGraphModel(name, graph)
		}
		for _, name := range currentEnv.MagicFunctions {
			c.MagicFunctions[name] = true
		}
		for name, fn := range currentEnv.ScalarFunctions {
			c.ScalarFunctions[name] = copyScalarFunctionDetails(fn)
		}
		for name, fn := range currentEnv.TabularFunctions {
			c.TabularFunctions[name] = copyTabularFunctionDetails(fn)
		}
		for _, groupName := range sortedEntityGroupNames(currentEnv.EntityGroups) {
			memberNames := append([]string{}, currentEnv.EntityGroups[groupName]...)
			members := make([]TableSchema, 0, len(memberNames))
			for _, tableName := range memberNames {
				table, _ := c.LookupTable(tableName)
				members = append(members, table)
			}
			c.EntityGroups[groupName] = members
			c.EntityGroupMembers[groupName] = memberNames
			c.EntityGroupNames[foldName(groupName)] = groupName
		}
	}
	c.addLocalData(local)
	c.Globals = BuildGlobalState(env, local)
	c.Binder = kustolang.NewBinder(c.Globals)
	if err := c.initializeDatabaseCatalogs(env); err != nil {
		return c, err
	}

	return c, nil
}

// newOwnedCatalog isolates public analyzer operations from NewCatalog's
// historical in-place normalization contract. This permits concurrent analysis
// of shared configuration without changing direct internal catalog behavior.
func newOwnedCatalog(env *EnvironmentDefinition, local LocalData) (*Catalog, error) {
	return NewCatalog(cloneEnvironmentDefinition(env), cloneLocalData(local))
}

func cloneEnvironmentDefinition(env *EnvironmentDefinition) *EnvironmentDefinition {
	if env == nil {
		return nil
	}
	return &EnvironmentDefinition{
		FunctionCatalogs:             slices.Clone(env.FunctionCatalogs),
		TableDetails:                 cloneTableDetailsMap(env.TableDetails),
		TableOptions:                 cloneTableOptionsMap(env.TableOptions),
		ExternalTables:               cloneTableDetailsMap(env.ExternalTables),
		ExternalTableOptions:         cloneTableOptionsMap(env.ExternalTableOptions),
		MaterializedViews:            cloneTableDetailsMap(env.MaterializedViews),
		MaterializedViewOptions:      cloneTableOptionsMap(env.MaterializedViewOptions),
		MaterializedViewQueries:      cloneStringMap(env.MaterializedViewQueries),
		StoredQueryResults:           cloneTableDetailsMap(env.StoredQueryResults),
		StoredQueryResultOptions:     cloneTableOptionsMap(env.StoredQueryResultOptions),
		GraphModels:                  cloneGraphDetailsMap(env.GraphModels),
		EntityGroups:                 cloneEntityGroups(env.EntityGroups),
		MagicFunctions:               append([]string(nil), env.MagicFunctions...),
		TabularFunctions:             cloneTabularFunctions(env.TabularFunctions),
		ScalarFunctions:              cloneScalarFunctions(env.ScalarFunctions),
		Domain:                       env.Domain,
		CurrentCluster:               env.CurrentCluster,
		CurrentDatabase:              env.CurrentDatabase,
		Clusters:                     cloneClusters(env.Clusters),
		TableOrder:                   append([]string(nil), env.TableOrder...),
		TableColumnOrder:             cloneStringSliceMap(env.TableColumnOrder),
		ExternalTableOrder:           append([]string(nil), env.ExternalTableOrder...),
		ExternalTableColumnOrder:     cloneStringSliceMap(env.ExternalTableColumnOrder),
		MaterializedViewOrder:        append([]string(nil), env.MaterializedViewOrder...),
		MaterializedViewColumnOrder:  cloneStringSliceMap(env.MaterializedViewColumnOrder),
		StoredQueryResultOrder:       append([]string(nil), env.StoredQueryResultOrder...),
		StoredQueryResultColumnOrder: cloneStringSliceMap(env.StoredQueryResultColumnOrder),
	}
}

func cloneLocalData(local LocalData) LocalData {
	return LocalData{
		Watchlists:                   cloneWatchlists(local.Watchlists),
		Tables:                       cloneTableDetailsMap(local.Tables),
		TableOptions:                 cloneTableOptionsMap(local.TableOptions),
		ExternalTables:               cloneTableDetailsMap(local.ExternalTables),
		ExternalTableOptions:         cloneTableOptionsMap(local.ExternalTableOptions),
		MaterializedViews:            cloneTableDetailsMap(local.MaterializedViews),
		MaterializedViewOptions:      cloneTableOptionsMap(local.MaterializedViewOptions),
		MaterializedViewQueries:      cloneStringMap(local.MaterializedViewQueries),
		StoredQueryResults:           cloneTableDetailsMap(local.StoredQueryResults),
		StoredQueryResultOptions:     cloneTableOptionsMap(local.StoredQueryResultOptions),
		GraphModels:                  cloneGraphDetailsMap(local.GraphModels),
		TabularFunctions:             cloneTabularFunctions(local.TabularFunctions),
		ScalarFunctions:              cloneScalarFunctions(local.ScalarFunctions),
		TableOrder:                   append([]string(nil), local.TableOrder...),
		TableColumnOrder:             cloneStringSliceMap(local.TableColumnOrder),
		ExternalTableOrder:           append([]string(nil), local.ExternalTableOrder...),
		ExternalTableColumnOrder:     cloneStringSliceMap(local.ExternalTableColumnOrder),
		MaterializedViewOrder:        append([]string(nil), local.MaterializedViewOrder...),
		MaterializedViewColumnOrder:  cloneStringSliceMap(local.MaterializedViewColumnOrder),
		StoredQueryResultOrder:       append([]string(nil), local.StoredQueryResultOrder...),
		StoredQueryResultColumnOrder: cloneStringSliceMap(local.StoredQueryResultColumnOrder),
	}
}

func cloneClusters(clusters map[string]ClusterDetails) map[string]ClusterDetails {
	if clusters == nil {
		return nil
	}
	result := make(map[string]ClusterDetails, len(clusters))
	for name, cluster := range clusters {
		result[name] = ClusterDetails{Databases: cloneDatabases(cluster.Databases)}
	}
	return result
}

func cloneDatabases(databases map[string]DatabaseDetails) map[string]DatabaseDetails {
	if databases == nil {
		return nil
	}
	result := make(map[string]DatabaseDetails, len(databases))
	for name, database := range databases {
		result[name] = cloneDatabaseDetails(database)
	}
	return result
}

func cloneDatabaseDetails(database DatabaseDetails) DatabaseDetails {
	return DatabaseDetails{
		AlternateName:                database.AlternateName,
		TableDetails:                 cloneTableDetailsMap(database.TableDetails),
		TableOptions:                 cloneTableOptionsMap(database.TableOptions),
		ExternalTables:               cloneTableDetailsMap(database.ExternalTables),
		ExternalTableOptions:         cloneTableOptionsMap(database.ExternalTableOptions),
		MaterializedViews:            cloneTableDetailsMap(database.MaterializedViews),
		MaterializedViewOptions:      cloneTableOptionsMap(database.MaterializedViewOptions),
		MaterializedViewQueries:      cloneStringMap(database.MaterializedViewQueries),
		StoredQueryResults:           cloneTableDetailsMap(database.StoredQueryResults),
		StoredQueryResultOptions:     cloneTableOptionsMap(database.StoredQueryResultOptions),
		GraphModels:                  cloneGraphDetailsMap(database.GraphModels),
		EntityGroups:                 cloneEntityGroups(database.EntityGroups),
		MagicFunctions:               append([]string(nil), database.MagicFunctions...),
		TabularFunctions:             cloneTabularFunctions(database.TabularFunctions),
		ScalarFunctions:              cloneScalarFunctions(database.ScalarFunctions),
		TableOrder:                   append([]string(nil), database.TableOrder...),
		TableColumnOrder:             cloneStringSliceMap(database.TableColumnOrder),
		ExternalTableOrder:           append([]string(nil), database.ExternalTableOrder...),
		ExternalTableColumnOrder:     cloneStringSliceMap(database.ExternalTableColumnOrder),
		MaterializedViewOrder:        append([]string(nil), database.MaterializedViewOrder...),
		MaterializedViewColumnOrder:  cloneStringSliceMap(database.MaterializedViewColumnOrder),
		StoredQueryResultOrder:       append([]string(nil), database.StoredQueryResultOrder...),
		StoredQueryResultColumnOrder: cloneStringSliceMap(database.StoredQueryResultColumnOrder),
	}
}

func cloneTableDetailsMap(tables map[string]TableDetails) map[string]TableDetails {
	if tables == nil {
		return nil
	}
	result := make(map[string]TableDetails, len(tables))
	for name, table := range tables {
		result[name] = cloneTableDetails(table)
	}
	return result
}

func cloneTableDetails(table TableDetails) TableDetails {
	if table == nil {
		return nil
	}
	result := make(TableDetails, len(table))
	for name, typ := range table {
		result[name] = typ
	}
	return result
}

func cloneTableOptionsMap(options map[string]TableOptions) map[string]TableOptions {
	if options == nil {
		return nil
	}
	result := make(map[string]TableOptions, len(options))
	for name, option := range options {
		result[name] = copyTableOptions(option)
	}
	return result
}

func cloneGraphDetailsMap(graphs map[string]GraphDetails) map[string]GraphDetails {
	if graphs == nil {
		return nil
	}
	result := make(map[string]GraphDetails, len(graphs))
	for name, graph := range graphs {
		result[name] = GraphDetails{
			Edges:     cloneTableDetailsMap(graph.Edges),
			Nodes:     cloneTableDetailsMap(graph.Nodes),
			Snapshots: append([]string(nil), graph.Snapshots...),
		}
	}
	return result
}

func cloneEntityGroups(groups map[string][]string) map[string][]string {
	if groups == nil {
		return nil
	}
	result := make(map[string][]string, len(groups))
	for name, members := range groups {
		result[name] = append([]string(nil), members...)
	}
	return result
}

func cloneWatchlists(watchlists map[string]WatchlistDetails) map[string]WatchlistDetails {
	if watchlists == nil {
		return nil
	}
	result := make(map[string]WatchlistDetails, len(watchlists))
	for name, columns := range watchlists {
		copy := make(WatchlistDetails, len(columns))
		for column, typ := range columns {
			copy[column] = typ
		}
		result[name] = copy
	}
	return result
}

func cloneTabularFunctions(functions map[string]TabularFunctionDetails) map[string]TabularFunctionDetails {
	if functions == nil {
		return nil
	}
	result := make(map[string]TabularFunctionDetails, len(functions))
	for name, function := range functions {
		function.OutputColumns = cloneTableDetails(function.OutputColumns)
		function.OutputColumnTypes = copyColumnTypeDetailsMap(function.OutputColumnTypes)
		function.Arguments = copyArgumentDetails(function.Arguments)
		result[name] = function
	}
	return result
}

func cloneScalarFunctions(functions map[string]ScalarFunctionDetails) map[string]ScalarFunctionDetails {
	if functions == nil {
		return nil
	}
	result := make(map[string]ScalarFunctionDetails, len(functions))
	for name, function := range functions {
		result[name] = copyScalarFunctionDetails(function)
	}
	return result
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	return copyStringMap(values)
}

func cloneStringSliceMap(values map[string][]string) map[string][]string {
	if values == nil {
		return nil
	}
	return copyStringSliceMap(values)
}

func (c *Catalog) initializeDatabaseCatalogs(env *EnvironmentDefinition) error {
	if c == nil || c.Globals == nil {
		return nil
	}
	if c.DatabaseCatalogs == nil {
		c.DatabaseCatalogs = map[*kustolang.DatabaseSymbol]*Catalog{}
	}
	if c.Globals.Database != nil {
		c.DatabaseCatalogs[c.Globals.Database] = c
	}
	if env == nil {
		return nil
	}
	for _, clusterName := range sortedClusterNames(env.Clusters) {
		clusterSymbol := c.Globals.GetCluster(clusterName)
		if clusterSymbol == nil {
			continue
		}
		clusterDetails := env.Clusters[clusterName]
		for _, databaseName := range sortedDatabaseNames(clusterDetails.Databases) {
			databaseSymbol := matchingDatabaseSymbol(clusterSymbol, databaseName)
			if databaseSymbol == nil || databaseSymbol == c.Globals.Database {
				continue
			}
			child, err := NewCatalog(databaseDetailsEnvironment(clusterDetails.Databases[databaseName]), LocalData{})
			if err != nil {
				return fmt.Errorf("cluster %q database %q: %w", clusterName, databaseName, err)
			}
			namespaceCatalogReferences(child, clusterName, databaseName)
			c.DatabaseCatalogs[databaseSymbol] = child
		}
	}
	return nil
}

func (c *Catalog) CatalogForDatabase(database *kustolang.DatabaseSymbol) *Catalog {
	if c == nil || database == nil {
		return nil
	}
	if catalog := c.DatabaseCatalogs[database]; catalog != nil {
		return catalog
	}
	return nil
}

func matchingDatabaseSymbol(cluster *kustolang.ClusterSymbol, name string) *kustolang.DatabaseSymbol {
	if cluster == nil {
		return nil
	}
	for _, database := range cluster.Databases() {
		if database != nil && (strings.EqualFold(database.SymbolName(), name) || strings.EqualFold(database.AlternateName(), name)) {
			return database
		}
	}
	return nil
}

func namespaceCatalogReferences(catalog *Catalog, clusterName, databaseName string) {
	if catalog == nil {
		return
	}
	prefix := "cluster\x00" + clusterName + "\x00database\x00" + databaseName + "\x00"
	for _, tables := range []map[string]TableSchema{
		catalog.Tables,
		catalog.ExternalTables,
		catalog.MaterializedViews,
		catalog.StoredQueryResults,
	} {
		for name, table := range tables {
			namespaceTableSchemaReferences(&table, prefix)
			tables[name] = table
		}
	}
	for name, graph := range catalog.GraphModels {
		namespaceTableSchemaReferences(&graph.Edges, prefix)
		namespaceTableSchemaReferences(&graph.Nodes, prefix)
		catalog.GraphModels[name] = graph
	}
	for name, members := range catalog.EntityGroups {
		for index := range members {
			namespaceTableSchemaReferences(&members[index], prefix)
		}
		catalog.EntityGroups[name] = members
	}
}

func namespaceTableSchemaReferences(table *TableSchema, prefix string) {
	if table == nil {
		return
	}
	for name, column := range table.Columns {
		if column.ReferenceID != "" {
			column.ReferenceID = prefix + column.ReferenceID
			table.Columns[name] = column
		}
	}
	for name, candidates := range table.ColumnCandidates {
		for index := range candidates {
			if candidates[index].ReferenceID != "" {
				candidates[index].ReferenceID = prefix + candidates[index].ReferenceID
			}
		}
		table.ColumnCandidates[name] = candidates
	}
}

func (c *Catalog) addTableOrdered(name string, details TableDetails, order []string, options TableOptions, origin bool) {
	c.addNamedTableOrdered(c.Tables, c.TableNames, "table", name, details, order, origin)
	table := c.Tables[name]
	table.Open = options.IsOpen
	table.Serialized = options.IsSerialized
	table.Sorted = options.IsSorted
	applyColumnTypeDetailsToSchema(&table, options.ColumnTypes)
	c.Tables[name] = table
}

func (c *Catalog) addExternalTable(name string, details TableDetails, order []string, options TableOptions, origin bool) {
	c.addNamedTableWithOptions(c.ExternalTables, c.ExternalTableNames, "external-table", name, details, order, options, origin)
}

func (c *Catalog) addMaterializedView(name string, details TableDetails, order []string, options TableOptions, origin bool) {
	c.addNamedTableWithOptions(c.MaterializedViews, c.MaterializedNames, "materialized-view", name, details, order, options, origin)
}

func (c *Catalog) addStoredQueryResult(name string, details TableDetails, order []string, options TableOptions, origin bool) {
	c.addNamedTableWithOptions(c.StoredQueryResults, c.StoredResultNames, "stored-query-result", name, details, order, options, origin)
}

func (c *Catalog) addGraphModel(name string, details GraphDetails) {
	c.GraphModels[name] = graphShapeFromDetails(details)
	c.GraphModelNames[foldName(name)] = name
}

func (c *Catalog) addNamedTableWithOptions(tables map[string]TableSchema, names map[string]string, kind string, name string, details TableDetails, order []string, options TableOptions, origin bool) {
	c.addNamedTableOrdered(tables, names, kind, name, details, order, origin)
	table := tables[name]
	table.Open = options.IsOpen
	table.Serialized = options.IsSerialized
	table.Sorted = options.IsSorted
	applyColumnTypeDetailsToSchema(&table, options.ColumnTypes)
	tables[name] = table
}

func (c *Catalog) addNamedTableOrdered(tables map[string]TableSchema, names map[string]string, kind string, name string, details TableDetails, order []string, origin bool) {
	schema := NewTableSchema()
	keys := mergeOrderedNames(nil, order, details)
	for _, k := range keys {
		schema.Set(k, normalizeType(details[k]), origin)
		if origin {
			column := schema.Columns[k]
			column.ReferenceID = kind + "\x00" + name + "\x00" + k
			schema.Columns[k] = column
		}
	}
	tables[name] = schema
	names[foldName(name)] = name
}

func BuildGlobalState(env *EnvironmentDefinition, local LocalData) *kustolang.GlobalState {
	local.normalize()
	currentEnv := currentEnvironmentDefinition(env)
	currentDatabaseName := "db"
	currentClusterName := ""
	if env != nil {
		if clusterName, ok := configuredCurrentClusterName(env); ok {
			currentClusterName = clusterName
			currentDatabaseName = configuredCurrentDatabaseName(env, clusterName)
		} else if strings.TrimSpace(env.CurrentDatabase) != "" {
			currentDatabaseName = strings.TrimSpace(env.CurrentDatabase)
		}
	}
	currentAlternateName := ""
	if env != nil && currentClusterName != "" {
		cluster := env.Clusters[currentClusterName]
		for name, database := range cluster.Databases {
			if strings.EqualFold(name, currentDatabaseName) {
				currentAlternateName = database.AlternateName
				break
			}
		}
	}
	currentDatabase := databaseSymbolFromEnvironment(currentDatabaseName, currentAlternateName, currentEnv, local)
	if env == nil || len(env.Clusters) == 0 {
		globals := kustolang.NewGlobalState(currentDatabase)
		if env != nil && strings.TrimSpace(env.CurrentCluster) != "" {
			cluster := kustolang.NewClusterSymbol(strings.TrimSpace(env.CurrentCluster), currentDatabase)
			globals = globals.WithClusterList([]*kustolang.ClusterSymbol{cluster}).WithCluster(cluster).WithDatabase(currentDatabase)
		}
		if env != nil && strings.TrimSpace(env.Domain) != "" {
			globals = globals.WithDomain(env.Domain)
		}
		return globals
	}

	clusters := make([]*kustolang.ClusterSymbol, 0, len(env.Clusters))
	var currentCluster *kustolang.ClusterSymbol
	for _, clusterName := range sortedClusterNames(env.Clusters) {
		clusterDetails := env.Clusters[clusterName]
		databases := make([]*kustolang.DatabaseSymbol, 0, len(clusterDetails.Databases)+1)
		currentIncluded := false
		for _, databaseName := range sortedDatabaseNames(clusterDetails.Databases) {
			databaseDetails := clusterDetails.Databases[databaseName]
			if strings.EqualFold(clusterName, currentClusterName) && strings.EqualFold(databaseName, currentDatabaseName) {
				databases = append(databases, currentDatabase)
				currentIncluded = true
				continue
			}
			databases = append(databases, databaseSymbolFromEnvironment(databaseName, databaseDetails.AlternateName,
				databaseDetailsEnvironment(databaseDetails), LocalData{}))
		}
		if strings.EqualFold(clusterName, currentClusterName) && !currentIncluded {
			databases = append(databases, currentDatabase)
		}
		cluster := kustolang.NewClusterSymbolWithDatabases(clusterName, databases, false)
		clusters = append(clusters, cluster)
		if strings.EqualFold(clusterName, currentClusterName) {
			currentCluster = cluster
		}
	}
	if currentCluster == nil {
		currentCluster = kustolang.NewClusterSymbol(currentDatabaseName+":cluster", currentDatabase)
		clusters = append(clusters, currentCluster)
	}
	globals := kustolang.NewGlobalState(currentDatabase).
		WithClusterList(clusters).
		WithCluster(currentCluster).
		WithDatabase(currentDatabase)
	if strings.TrimSpace(env.Domain) != "" {
		globals = globals.WithDomain(env.Domain)
	}
	return globals
}

func databaseSymbolFromEnvironment(name, alternateName string, env *EnvironmentDefinition, local LocalData) *kustolang.DatabaseSymbol {
	var members []kustolang.Symbol
	tableSymbols := map[string]*kustolang.TableSymbol{}
	if env != nil {
		// Non-current databases reach this path without the root overlay above.
		allowDefenderTimeAliases(env.TableDetails)
		for _, name := range mergeOrderedNames(nil, env.TableOrder, env.TableDetails) {
			options, _ := tableOptionsForName(env.TableOptions, name)
			table := tableSymbolFromDetailsOrderedWithOptions(name, env.TableDetails[name], env.TableColumnOrder[name], options)
			tableSymbols[name] = table
			members = append(members, table)
		}
		for _, name := range mergeOrderedNames(nil, env.ExternalTableOrder, env.ExternalTables) {
			options, _ := tableOptionsForName(env.ExternalTableOptions, name)
			members = append(members, externalTableSymbolFromDetails(name, env.ExternalTables[name], env.ExternalTableColumnOrder[name], options))
		}
		for _, name := range mergeOrderedNames(nil, env.MaterializedViewOrder, env.MaterializedViews) {
			options, _ := tableOptionsForName(env.MaterializedViewOptions, name)
			members = append(members, materializedViewSymbolFromDetails(name, env.MaterializedViews[name], env.MaterializedViewColumnOrder[name], options, env.MaterializedViewQueries[name]))
		}
		for _, name := range mergeOrderedNames(nil, env.StoredQueryResultOrder, env.StoredQueryResults) {
			options, _ := tableOptionsForName(env.StoredQueryResultOptions, name)
			members = append(members, storedQueryResultSymbolFromDetails(name, env.StoredQueryResults[name], env.StoredQueryResultColumnOrder[name], options))
		}
		for _, name := range sortedGraphModelNames(env.GraphModels) {
			members = append(members, graphModelSymbolFromDetails(name, env.GraphModels[name]))
		}
		for _, name := range env.MagicFunctions {
			if name == "_GetWatchlist" {
				members = append(members, watchlistFunctionSymbol(local.Watchlists))
			} else {
				members = append(members, magicFunctionSymbol(name))
			}
		}
		for _, name := range sortedScalarFunctionNames(env.ScalarFunctions) {
			fn := env.ScalarFunctions[name]
			members = append(members, kustolang.NewScalarFunctionSymbol(name, scalarFunctionTypeSymbol(fn), parametersFromArguments(fn.Arguments)...))
		}
		for _, name := range sortedTabularFunctionNames(env.TabularFunctions) {
			fn := env.TabularFunctions[name]
			members = append(members, tabularFunctionSymbolFromDetails(name, fn))
		}
		for _, groupName := range sortedEntityGroupNames(env.EntityGroups) {
			groupMembers := make([]kustolang.Symbol, 0, len(env.EntityGroups[groupName]))
			for _, tableName := range env.EntityGroups[groupName] {
				if table := tableSymbols[tableName]; table != nil {
					groupMembers = append(groupMembers, table)
				}
			}
			members = append(members, kustolang.NewEntityGroupSymbol(groupName, groupMembers...))
		}
	}
	members = appendLocalDatabaseMembers(members, local)
	return kustolang.NewDatabaseSymbolWithAlternateName(name, alternateName, members, false)
}

func replaceOrdinaryTableMember(members []kustolang.Symbol, replacement *kustolang.TableSymbol) []kustolang.Symbol {
	if replacement == nil {
		return members
	}
	for i, member := range members {
		table, ok := member.(*kustolang.TableSymbol)
		if ok && table.SymbolName() == replacement.SymbolName() {
			members[i] = replacement
			return members
		}
	}
	return append(members, replacement)
}

func replaceSpecialEntityMember(members []kustolang.Symbol, replacement kustolang.Symbol) []kustolang.Symbol {
	if replacement == nil {
		return members
	}
	for i, member := range members {
		if member != nil && member.SymbolName() == replacement.SymbolName() && sameSpecialEntityKind(member, replacement) {
			members[i] = replacement
			return members
		}
	}
	return append(members, replacement)
}

func sameSpecialEntityKind(left, right kustolang.Symbol) bool {
	switch left.(type) {
	case *kustolang.ExternalTableSymbol:
		_, ok := right.(*kustolang.ExternalTableSymbol)
		return ok
	case *kustolang.MaterializedViewSymbol:
		_, ok := right.(*kustolang.MaterializedViewSymbol)
		return ok
	case *kustolang.StoredQueryResultSymbol:
		_, ok := right.(*kustolang.StoredQueryResultSymbol)
		return ok
	default:
		return false
	}
}

func tableSymbolFromDetails(name string, details TableDetails) *kustolang.TableSymbol {
	return kustolang.NewTableSymbol(name, columnSymbolsFromDetails(details)...)
}

func tableSymbolFromDetailsOrderedWithOptions(name string, details TableDetails, order []string, options TableOptions) *kustolang.TableSymbol {
	table := kustolang.NewTableSymbol(name, columnSymbolsFromDetailsOrderedWithTypes(details, order, options.ColumnTypes)...)
	table = table.WithIsOpen(options.IsOpen).
		WithIsSerialized(options.IsSerialized).
		WithIsSorted(options.IsSorted)
	return table
}

func scalarFunctionTypeSymbol(details ScalarFunctionDetails) *kustolang.TypeSymbol {
	if details.OutputTypeDetails != nil {
		return typeSymbolFromColumnTypeDetails(*details.OutputTypeDetails, true)
	}
	return scalarType(details.OutputType)
}

func scalarFunctionPublicType(details ScalarFunctionDetails) string {
	return publicTypeName(scalarFunctionTypeSymbol(details))
}

func argumentTypeSymbol(details ArgumentDetails) *kustolang.TypeSymbol {
	if details.TypeDetails != nil {
		return typeSymbolFromColumnTypeDetails(*details.TypeDetails, true)
	}
	return scalarType(details.Type)
}

func tabularFunctionSchema(details TabularFunctionDetails, origin bool) TableSchema {
	table := schemaFromTableDetails(details.OutputColumns, origin)
	table.Open = details.IsOpen
	applyColumnTypeDetailsToSchema(&table, details.OutputColumnTypes)
	return table
}

func tabularFunctionTableSymbol(name string, details TabularFunctionDetails) *kustolang.TableSymbol {
	return kustolang.NewTableSymbol(name, columnSymbolsFromDetailsOrderedWithTypes(details.OutputColumns, nil, details.OutputColumnTypes)...).WithIsOpen(details.IsOpen)
}

func tabularFunctionSymbolFromDetails(name string, details TabularFunctionDetails) *kustolang.FunctionSymbol {
	details = copyTabularFunctionDetails(details)
	return kustolang.NewFunctionSymbol(name, kustolang.NewCustomSignature(func(context kustolang.CustomReturnTypeContext) kustolang.Symbol {
		table := tabularFunctionTableSymbol("", details)
		if context.RowScope != nil {
			table = table.WithInheritableProperties(context.RowScope)
		}
		return table
	}, kustolang.TabularityTabular, parametersFromArguments(details.Arguments)...))
}

func externalTableSymbolFromDetails(name string, details TableDetails, order []string, options TableOptions) *kustolang.ExternalTableSymbol {
	return kustolang.NewExternalTableSymbol(name, columnSymbolsFromDetailsOrderedWithTypes(details, order, options.ColumnTypes)...).
		WithIsOpen(options.IsOpen).
		WithIsSerialized(options.IsSerialized).
		WithIsSorted(options.IsSorted)
}

func materializedViewSymbolFromDetails(name string, details TableDetails, order []string, options TableOptions, query string) *kustolang.MaterializedViewSymbol {
	return kustolang.NewMaterializedViewSymbol(name, columnSymbolsFromDetailsOrderedWithTypes(details, order, options.ColumnTypes), query).
		WithIsOpen(options.IsOpen).
		WithIsSerialized(options.IsSerialized).
		WithIsSorted(options.IsSorted)
}

func storedQueryResultSymbolFromDetails(name string, details TableDetails, order []string, options TableOptions) *kustolang.StoredQueryResultSymbol {
	return kustolang.NewStoredQueryResultSymbol(name, columnSymbolsFromDetailsOrderedWithTypes(details, order, options.ColumnTypes)...)
}

func graphModelSymbolFromDetails(name string, details GraphDetails) *kustolang.GraphModelSymbol {
	edges := graphSignaturesFromDetails(details.Edges)
	nodes := graphSignaturesFromDetails(details.Nodes)
	snapshots := make([]*kustolang.GraphSnapshotSymbol, 0, len(details.Snapshots))
	for _, snapshot := range details.Snapshots {
		snapshots = append(snapshots, kustolang.NewGraphSnapshotSymbol(snapshot))
	}
	return kustolang.NewGraphModelSymbol(name, edges, nodes, snapshots)
}

func graphSignaturesFromDetails(details map[string]TableDetails) []*kustolang.Signature {
	names := sortedTableNames(details)
	signatures := make([]*kustolang.Signature, 0, len(names))
	for _, name := range names {
		signatures = append(signatures, kustolang.NewSymbolSignature(tableSymbolFromDetails(name, details[name])))
	}
	return signatures
}

func columnSymbolsFromDetails(details TableDetails) []*kustolang.ColumnSymbol {
	return columnSymbolsFromDetailsOrdered(details, nil)
}

func columnSymbolsFromDetailsOrdered(details TableDetails, order []string) []*kustolang.ColumnSymbol {
	return columnSymbolsFromDetailsOrderedWithTypes(details, order, nil)
}

func columnSymbolsFromDetailsOrderedWithTypes(details TableDetails, order []string, columnTypes map[string]ColumnTypeDetails) []*kustolang.ColumnSymbol {
	var columns []*kustolang.ColumnSymbol
	keys := mergeOrderedNames(nil, order, details)
	for _, key := range keys {
		typ := scalarType(details[key])
		if value, ok := columnTypeDetailsForName(columnTypes, key); ok {
			typ = typeSymbolFromColumnTypeDetails(value, true)
		}
		columns = append(columns, kustolang.NewColumnSymbol(key, typ))
	}
	return columns
}

func applyColumnTypeDetailsToSchema(schema *TableSchema, columnTypes map[string]ColumnTypeDetails) {
	if schema == nil {
		return
	}
	for _, name := range sortedColumnTypeDetailsNames(columnTypes) {
		if existing, ok := schema.Columns[name]; ok {
			column := columnFromTypeSymbol(typeSymbolFromColumnTypeDetails(columnTypes[name], true), existing.Origin)
			column.ReferenceID = existing.ReferenceID
			schema.SetColumn(name, column)
		}
	}
}

func columnTypeDetailsForName(details map[string]ColumnTypeDetails, name string) (ColumnTypeDetails, bool) {
	if value, ok := details[name]; ok {
		return value, true
	}
	for candidate, value := range details {
		if strings.EqualFold(candidate, name) {
			return value, true
		}
	}
	return ColumnTypeDetails{}, false
}

func typeSymbolFromColumnTypeDetails(details ColumnTypeDetails, dynamicRoot bool) *kustolang.TypeSymbol {
	if details.Type == "bag" {
		properties := make([]*kustolang.ColumnSymbol, 0, len(details.Properties))
		for _, name := range orderedColumnTypeDetailsNames(details) {
			properties = append(properties, kustolang.NewColumnSymbol(name, typeSymbolFromColumnTypeDetails(details.Properties[name], false)))
		}
		return kustolang.ScalarTypes.GetDynamicBag(properties...).Type()
	}
	if details.Type == "array" {
		return kustolang.ScalarTypes.GetDynamicArray(typeSymbolFromColumnTypeDetails(*details.Element, false)).Type()
	}
	typ := scalarType(details.Type)
	if dynamicRoot {
		return kustolang.ScalarTypes.GetDynamic(typ)
	}
	return typ
}

func mergeOrderedNames[T any](existing, preferred []string, values map[string]T) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	appendKnown := func(names []string) {
		for _, name := range names {
			if _, ok := values[name]; !ok {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			result = append(result, name)
		}
	}
	appendKnown(existing)
	appendKnown(preferred)
	remaining := make([]string, 0, len(values)-len(result))
	for name := range values {
		if _, ok := seen[name]; !ok {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	return append(result, remaining...)
}

func parametersFromArguments(args []ArgumentDetails) []*kustolang.Parameter {
	params := make([]*kustolang.Parameter, 0, len(args))
	for _, arg := range args {
		min := 1
		if arg.Optional {
			min = 0
		}
		params = append(params, kustolang.NewParameter(arg.Name, argumentTypeSymbol(arg), kustolang.WithParameterOccurrence(min, 1)))
	}
	return params
}

func scalarType(name string) *kustolang.TypeSymbol {
	if strings.EqualFold(strings.TrimSpace(name), kustolang.Error.SymbolName()) {
		return &kustolang.Error.TypeSymbol
	}
	if symbol := kustolang.ScalarTypes.Get(name); symbol != nil {
		return symbol.Type()
	}
	return kustolang.ScalarTypes.Dynamic.Type()
}

func magicFunctionSymbol(name string) *kustolang.FunctionSymbol {
	switch name {
	case "FileProfile":
		return magicInvokeFunctionSymbol(name, TableDetails{
			"SHA1": "string", "SHA256": "string", "MD5": "string", "FileSize": "long",
			"GlobalPrevalence": "long", "GlobalFirstSeen": "datetime", "GlobalLastSeen": "datetime",
			"Signer": "string", "Issuer": "string", "SignerHash": "string", "IsCertificateValid": "bool",
			"IsRootSignerMicrosoft": "bool", "SignatureState": "string", "IsExecutable": "bool",
			"ThreatName": "string", "Publisher": "string", "SoftwareName": "string", "ProfileAvailability": "string",
		})
	case "DeviceFromIP":
		return magicInvokeFunctionSymbol(name, TableDetails{"IP": "string", "DeviceId": "string"})
	case "_GetWatchlist":
		return watchlistFunctionSymbol(nil)
	default:
		return kustolang.NewComputedFunctionSymbol(name, "", kustolang.TabularityUnknown)
	}
}

func magicInvokeFunctionSymbol(name string, output TableDetails) *kustolang.FunctionSymbol {
	return kustolang.NewFunctionSymbol(name, kustolang.NewCustomSignature(func(context kustolang.CustomReturnTypeContext) kustolang.Symbol {
		builder := kustolang.NewProjectionBuilder()
		if len(context.ArgumentResultSymbols) > 0 {
			if input, ok := context.ArgumentResultSymbols[0].(*kustolang.TableSymbol); ok && input != nil {
				builder.AddRange(input.Columns(), kustolang.ProjectionAddRangeOptions{DoNotRepeat: true})
			}
		}
		for _, column := range columnSymbolsFromDetailsOrderedWithTypes(output, nil, nil) {
			builder.Add(column)
		}
		return kustolang.NewTableSymbol("", builder.GetProjection()...)
	}, kustolang.TabularityTabular,
		kustolang.NewParameterKind("T", kustolang.ParameterTypeTabular),
		kustolang.NewParameter("x", kustolang.ScalarTypes.String.Type(), kustolang.WithParameterOccurrence(0, 1)),
		kustolang.NewParameter("y", kustolang.ScalarTypes.String.Type(), kustolang.WithParameterOccurrence(0, 1)),
	))
}

func sortedTableNames(tables map[string]TableDetails) []string {
	names := make([]string, 0, len(tables))
	for name := range tables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedScalarFunctionNames(functions map[string]ScalarFunctionDetails) []string {
	names := make([]string, 0, len(functions))
	for name := range functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedTabularFunctionNames(functions map[string]TabularFunctionDetails) []string {
	names := make([]string, 0, len(functions))
	for name := range functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedEntityGroupNames(groups map[string][]string) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedEnvironmentNames(envs KQLEnvironments) []string {
	names := make([]string, 0, len(envs))
	for name := range envs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedClusterNames(clusters map[string]ClusterDetails) []string {
	names := make([]string, 0, len(clusters))
	for name := range clusters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedDatabaseNames(databases map[string]DatabaseDetails) []string {
	names := make([]string, 0, len(databases))
	for name := range databases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Catalog) LookupTable(name string) (TableSchema, bool) {
	if table, ok := lookupNamedTable(c.Tables, c.TableNames, name); ok {
		return table, true
	}
	if c.Globals != nil && c.Globals.Database != nil {
		if table := c.Globals.Database.GetTable(name); table != nil {
			return tableSchemaFromSymbol(table), true
		}
	}
	return TableSchema{}, false
}

func (c *Catalog) LookupExternalTable(name string) (TableSchema, bool) {
	if table, ok := lookupNamedTable(c.ExternalTables, c.ExternalTableNames, name); ok {
		return table, true
	}
	if c.Globals != nil && c.Globals.Database != nil {
		if table := c.Globals.Database.GetExternalTable(name); table != nil {
			return tableSchemaFromSymbol(&table.TableSymbol), true
		}
	}
	return TableSchema{}, false
}

func (c *Catalog) LookupMaterializedView(name string) (TableSchema, bool) {
	if table, ok := lookupNamedTable(c.MaterializedViews, c.MaterializedNames, name); ok {
		return table, true
	}
	if c.Globals != nil && c.Globals.Database != nil {
		if table := c.Globals.Database.GetMaterializedView(name); table != nil {
			return tableSchemaFromSymbol(&table.TableSymbol), true
		}
	}
	return TableSchema{}, false
}

func (c *Catalog) LookupStoredQueryResult(name string) (TableSchema, bool) {
	if table, ok := lookupNamedTable(c.StoredQueryResults, c.StoredResultNames, name); ok {
		return table, true
	}
	if c.Globals != nil && c.Globals.Database != nil {
		if stored := c.Globals.Database.GetStoredQueryResult(name); stored != nil {
			return tableSchemaFromSymbol(&stored.TableSymbol), true
		}
	}
	return TableSchema{}, false
}

func (c *Catalog) LookupGraphModel(name string) (GraphShapeDetails, bool) {
	if graph, ok := lookupNamedGraph(c.GraphModels, c.GraphModelNames, name); ok {
		return graph, true
	}
	if c.Globals != nil && c.Globals.Database != nil {
		if model := c.Globals.Database.GetGraphModel(name); model != nil {
			return graphShapeFromSymbol(model), true
		}
	}
	return GraphShapeDetails{}, false
}

func (c *Catalog) LookupEntityGroup(name string) (resolvedName string, members []TableSchema, memberNames []string, ok bool) {
	if c == nil {
		return "", nil, nil, false
	}
	resolvedName, ok = c.EntityGroupNames[foldName(name)]
	if !ok {
		return "", nil, nil, false
	}
	stored := c.EntityGroups[resolvedName]
	members = make([]TableSchema, 0, len(stored))
	for _, member := range stored {
		members = append(members, member.Clone())
	}
	return resolvedName, members, append([]string{}, c.EntityGroupMembers[resolvedName]...), true
}

func lookupNamedTable(tables map[string]TableSchema, names map[string]string, name string) (TableSchema, bool) {
	if table, ok := tables[name]; ok {
		return table.Clone(), true
	}
	return TableSchema{}, false
}

func lookupNamedGraph(graphs map[string]GraphShapeDetails, names map[string]string, name string) (GraphShapeDetails, bool) {
	if graph, ok := graphs[name]; ok {
		return graph.Clone(), true
	}
	return GraphShapeDetails{}, false
}

func (c *Catalog) HasFunction(name string) bool {
	_, okScalar := c.ScalarFunctions[name]
	_, okTabular := c.TabularFunctions[name]
	if okScalar || okTabular || c.MagicFunctions[name] {
		return true
	}
	return c.Globals != nil && c.Globals.Database != nil && c.Globals.Database.GetFunction(name) != nil
}

func graphShapeFromDetails(details GraphDetails) GraphShapeDetails {
	snapshots := map[string]string{}
	for _, snapshot := range details.Snapshots {
		snapshots[snapshot] = snapshot
	}
	modelNames := make([]string, 0, len(details.Edges)+len(details.Nodes))
	for name := range details.Edges {
		modelNames = append(modelNames, name)
	}
	for name := range details.Nodes {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames)
	return GraphShapeDetails{
		Edges:      mergedTableDetails(details.Edges),
		Nodes:      mergedTableDetails(details.Nodes),
		Snapshots:  snapshots,
		ModelNames: modelNames,
	}
}

func graphShapeFromSymbol(model *kustolang.GraphModelSymbol) GraphShapeDetails {
	shape := GraphShapeDetails{
		Edges:     NewTableSchema(),
		Nodes:     NewTableSchema(),
		Snapshots: map[string]string{},
	}
	if model == nil {
		return shape
	}
	for _, edge := range model.Edges {
		shape.Edges.Merge(tableSchemaFromSignature(edge))
	}
	for _, node := range model.Nodes {
		shape.Nodes.Merge(tableSchemaFromSignature(node))
	}
	for _, snapshot := range model.Snapshots {
		if snapshot != nil {
			shape.Snapshots[snapshot.SymbolName()] = snapshot.SymbolName()
		}
	}
	return shape
}

func mergedTableDetails(details map[string]TableDetails) TableSchema {
	out := NewTableSchema()
	for _, name := range sortedTableNames(details) {
		out.Merge(schemaFromTableDetails(details[name], true))
	}
	return out
}

func tableSchemaFromSignature(signature *kustolang.Signature) TableSchema {
	if signature == nil || signature.DeclaredReturnSymbol == nil {
		return NewTableSchema()
	}
	switch symbol := signature.DeclaredReturnSymbol.(type) {
	case *kustolang.TableSymbol:
		return tableSchemaFromSymbol(symbol)
	case *kustolang.ExternalTableSymbol:
		return tableSchemaFromSymbol(&symbol.TableSymbol)
	case *kustolang.MaterializedViewSymbol:
		return tableSchemaFromSymbol(&symbol.TableSymbol)
	case *kustolang.StoredQueryResultSymbol:
		return tableSchemaFromSymbol(&symbol.TableSymbol)
	default:
		return NewTableSchema()
	}
}

func (g GraphShapeDetails) Clone() GraphShapeDetails {
	snapshots := map[string]string{}
	for key, value := range g.Snapshots {
		snapshots[key] = value
	}
	modelNames := append([]string{}, g.ModelNames...)
	return GraphShapeDetails{
		Edges:      g.Edges.Clone(),
		Nodes:      g.Nodes.Clone(),
		Snapshots:  snapshots,
		ModelNames: modelNames,
	}
}

func sortedGraphModelNames(graphs map[string]GraphDetails) []string {
	names := make([]string, 0, len(graphs))
	for name := range graphs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Catalog) ValidateTypes() error {
	for tableName, table := range c.Tables {
		for colName, col := range table.Columns {
			if col.Type == "" {
				return fmt.Errorf("unknown type for column %s in table %s", colName, tableName)
			}
		}
	}
	return nil
}

func tableSchemaFromSymbol(table *kustolang.TableSymbol) TableSchema {
	schema := NewTableSchema()
	if table == nil {
		return schema
	}
	schema.Open = table.IsOpen()
	schema.Serialized = table.IsSerialized()
	schema.Sorted = table.IsSorted()
	for _, column := range table.Columns() {
		converted := columnFromTypeSymbol(column.Type, true)
		if column.SymbolName() == "" {
			schema.SetUnnamedColumn(converted)
		} else {
			schema.SetColumn(column.SymbolName(), converted)
		}
	}
	return schema
}

func normalizeLocalData(local *LocalData) error {
	local.normalize()
	allowDefenderTimeAliases(local.Tables)
	if err := normalizeConfiguredFunctions("local_data", local.ScalarFunctions, local.TabularFunctions); err != nil {
		return err
	}
	if err := normalizeTableOptions(local.Tables, &local.TableOptions); err != nil {
		return fmt.Errorf("local_data: %w", err)
	}
	if err := normalizeEntityTableOptions("external_table_options", "external table", local.ExternalTables, &local.ExternalTableOptions, true); err != nil {
		return fmt.Errorf("local_data: %w", err)
	}
	if err := normalizeEntityTableOptions("materialized_view_options", "materialized view", local.MaterializedViews, &local.MaterializedViewOptions, true); err != nil {
		return fmt.Errorf("local_data: %w", err)
	}
	if err := normalizeEntityTableOptions("stored_query_result_options", "stored query result", local.StoredQueryResults, &local.StoredQueryResultOptions, false); err != nil {
		return fmt.Errorf("local_data: %w", err)
	}
	if err := normalizeMaterializedViewQueries(local.MaterializedViews, &local.MaterializedViewQueries); err != nil {
		return fmt.Errorf("local_data: %w", err)
	}
	return nil
}

func (c *Catalog) addLocalData(local LocalData) {
	c.Watchlists = local.Watchlists
	for _, name := range mergeOrderedNames(nil, local.TableOrder, local.Tables) {
		options, _ := tableOptionsForName(local.TableOptions, name)
		c.addTableOrdered(name, local.Tables[name], local.TableColumnOrder[name], options, true)
		c.ImpliedTables = append(c.ImpliedTables, c.Tables[name])
	}
	for _, name := range mergeOrderedNames(nil, local.ExternalTableOrder, local.ExternalTables) {
		table := local.ExternalTables[name]
		options, _ := tableOptionsForName(local.ExternalTableOptions, name)
		c.addExternalTable(name, table, local.ExternalTableColumnOrder[name], options, true)
	}
	for _, name := range mergeOrderedNames(nil, local.MaterializedViewOrder, local.MaterializedViews) {
		table := local.MaterializedViews[name]
		options, _ := tableOptionsForName(local.MaterializedViewOptions, name)
		c.addMaterializedView(name, table, local.MaterializedViewColumnOrder[name], options, true)
	}
	for _, name := range mergeOrderedNames(nil, local.StoredQueryResultOrder, local.StoredQueryResults) {
		table := local.StoredQueryResults[name]
		options, _ := tableOptionsForName(local.StoredQueryResultOptions, name)
		c.addStoredQueryResult(name, table, local.StoredQueryResultColumnOrder[name], options, true)
	}
	for name, graph := range local.GraphModels {
		c.addGraphModel(name, graph)
	}
	for name, fn := range local.ScalarFunctions {
		c.ScalarFunctions[name] = copyScalarFunctionDetails(fn)
	}
	for name, fn := range local.TabularFunctions {
		c.TabularFunctions[name] = copyTabularFunctionDetails(fn)
	}
	if len(local.Watchlists) > 0 {
		c.MagicFunctions["_GetWatchlist"] = true
	}
}

func appendLocalDatabaseMembers(members []kustolang.Symbol, local LocalData) []kustolang.Symbol {
	for _, name := range mergeOrderedNames(nil, local.TableOrder, local.Tables) {
		options, _ := tableOptionsForName(local.TableOptions, name)
		table := tableSymbolFromDetailsOrderedWithOptions(name, local.Tables[name], local.TableColumnOrder[name], options)
		if _, configured := tableOptionsForName(local.TableOptions, name); configured {
			members = replaceOrdinaryTableMember(members, table)
		} else {
			members = append(members, table)
		}
	}
	for _, name := range mergeOrderedNames(nil, local.ExternalTableOrder, local.ExternalTables) {
		options, _ := tableOptionsForName(local.ExternalTableOptions, name)
		members = replaceSpecialEntityMember(members, externalTableSymbolFromDetails(name, local.ExternalTables[name], local.ExternalTableColumnOrder[name], options))
	}
	for _, name := range mergeOrderedNames(nil, local.MaterializedViewOrder, local.MaterializedViews) {
		options, _ := tableOptionsForName(local.MaterializedViewOptions, name)
		members = replaceSpecialEntityMember(members, materializedViewSymbolFromDetails(name, local.MaterializedViews[name], local.MaterializedViewColumnOrder[name], options, local.MaterializedViewQueries[name]))
	}
	for _, name := range mergeOrderedNames(nil, local.StoredQueryResultOrder, local.StoredQueryResults) {
		options, _ := tableOptionsForName(local.StoredQueryResultOptions, name)
		members = replaceSpecialEntityMember(members, storedQueryResultSymbolFromDetails(name, local.StoredQueryResults[name], local.StoredQueryResultColumnOrder[name], options))
	}
	for _, name := range sortedGraphModelNames(local.GraphModels) {
		members = append(members, graphModelSymbolFromDetails(name, local.GraphModels[name]))
	}
	for _, name := range sortedScalarFunctionNames(local.ScalarFunctions) {
		fn := local.ScalarFunctions[name]
		members = replaceFunctionMember(members, kustolang.NewScalarFunctionSymbol(name, scalarFunctionTypeSymbol(fn), parametersFromArguments(fn.Arguments)...))
	}
	for _, name := range sortedTabularFunctionNames(local.TabularFunctions) {
		fn := local.TabularFunctions[name]
		members = replaceFunctionMember(members, tabularFunctionSymbolFromDetails(name, fn))
	}
	if len(local.Watchlists) > 0 {
		members = replaceFunctionMember(members, watchlistFunctionSymbol(local.Watchlists))
	}
	return members
}

func replaceFunctionMember(members []kustolang.Symbol, replacement *kustolang.FunctionSymbol) []kustolang.Symbol {
	for i, member := range members {
		if function, ok := member.(*kustolang.FunctionSymbol); ok && function.SymbolName() == replacement.SymbolName() {
			members[i] = replacement
			return members
		}
	}
	return append(members, replacement)
}
