package environment

import (
	"maps"
	"slices"

	kustolang "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

// PreparedEnvironment owns a normalized, immutable snapshot of an environment.
// Preparation happens at runtime; the external schema remains independently
// deployable. Create a new snapshot to pick up changes to that schema.
type PreparedEnvironment struct {
	catalog *Catalog
}

func PrepareEnvironment(env *EnvironmentDefinition) (*PreparedEnvironment, error) {
	catalog, err := newOwnedCatalog(env, LocalData{})
	if err != nil {
		return nil, err
	}
	return &PreparedEnvironment{catalog: catalog}, nil
}

// NewCatalog creates a request-owned view. Existing schemas and symbols are
// read-only; only maps affected by local declarations and the current database's
// member list are copied. Binders and GlobalState settings belong to the view.
func (p *PreparedEnvironment) NewCatalog(local LocalData) (*Catalog, error) {
	local = cloneLocalData(local)
	if err := normalizeLocalData(&local); err != nil {
		return nil, err
	}
	base := p.catalog
	c := *base
	if len(local.Tables) > 0 {
		c.Tables, c.TableNames = maps.Clone(base.Tables), maps.Clone(base.TableNames)
		c.ImpliedTables = slices.Clone(base.ImpliedTables)
	}
	if len(local.ExternalTables) > 0 {
		c.ExternalTables, c.ExternalTableNames = maps.Clone(base.ExternalTables), maps.Clone(base.ExternalTableNames)
	}
	if len(local.MaterializedViews) > 0 {
		c.MaterializedViews, c.MaterializedNames = maps.Clone(base.MaterializedViews), maps.Clone(base.MaterializedNames)
	}
	if len(local.StoredQueryResults) > 0 {
		c.StoredQueryResults, c.StoredResultNames = maps.Clone(base.StoredQueryResults), maps.Clone(base.StoredResultNames)
	}
	if len(local.GraphModels) > 0 {
		c.GraphModels, c.GraphModelNames = maps.Clone(base.GraphModels), maps.Clone(base.GraphModelNames)
	}
	if len(local.ScalarFunctions) > 0 {
		c.ScalarFunctions = maps.Clone(base.ScalarFunctions)
	}
	if len(local.TabularFunctions) > 0 {
		c.TabularFunctions = maps.Clone(base.TabularFunctions)
	}
	if len(local.Watchlists) > 0 {
		c.MagicFunctions = maps.Clone(base.MagicFunctions)
	}
	c.addLocalData(local)
	// Limit configuration also updates remote catalog views. Copy that small
	// graph (including self references), while sharing its immutable schemas.
	c.DatabaseCatalogs = cloneCatalogViews(base.DatabaseCatalogs, map[*Catalog]*Catalog{base: &c})
	if hasLocalSymbols(local) {
		current := base.Globals.Database
		database := current.WithMembers(appendLocalDatabaseMembers(current.Members(), local))
		cluster := base.Globals.Cluster.UpdateDatabase(current, database)
		c.Globals = base.Globals.AddOrReplaceCluster(cluster).WithCluster(cluster).WithDatabase(database)
		delete(c.DatabaseCatalogs, current)
	}
	c.DatabaseCatalogs[c.Globals.Database] = &c
	c.Binder = kustolang.NewBinder(c.Globals)
	return &c, nil
}

func cloneCatalogViews(catalogs map[*kustolang.DatabaseSymbol]*Catalog, copies map[*Catalog]*Catalog) map[*kustolang.DatabaseSymbol]*Catalog {
	views := make(map[*kustolang.DatabaseSymbol]*Catalog, len(catalogs))
	for database, catalog := range catalogs {
		view, exists := copies[catalog]
		if !exists {
			cloned := *catalog
			view = &cloned
			copies[catalog] = view
			view.DatabaseCatalogs = cloneCatalogViews(catalog.DatabaseCatalogs, copies)
			view.Binder = kustolang.NewBinder(view.Globals)
		}
		views[database] = view
	}
	return views
}

func hasLocalSymbols(local LocalData) bool {
	return len(local.Tables)+len(local.ExternalTables)+len(local.MaterializedViews)+
		len(local.StoredQueryResults)+len(local.GraphModels)+len(local.ScalarFunctions)+
		len(local.TabularFunctions)+len(local.Watchlists) > 0
}
