package environment

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	kustolang "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

//go:embed sentinel_catalog.json
var sentinelCatalogJSON []byte

type sentinelCatalog struct {
	Source    string                           `json:"source"`
	Revision  string                           `json:"revision"`
	Schemas   map[string]TableDetails          `json:"schemas"`
	Functions map[string]sentinelParserDetails `json:"functions"`
}

type sentinelParserDetails struct {
	Schema    string            `json:"schema"`
	Arguments []ArgumentDetails `json:"arguments"`
}

var loadSentinelCatalog = sync.OnceValues(func() (sentinelCatalog, error) {
	var catalog sentinelCatalog
	err := json.Unmarshal(sentinelCatalogJSON, &catalog)
	return catalog, err
})

func validateFunctionCatalogs(names []string) error {
	for _, name := range names {
		if name != "sentinel" {
			return fmt.Errorf("unknown function catalog %q", name)
		}
		if _, err := loadSentinelCatalog(); err != nil {
			return fmt.Errorf("load Sentinel function catalog: %w", err)
		}
	}
	return nil
}

func addFunctionCatalogs(env *EnvironmentDefinition) error {
	if err := validateFunctionCatalogs(env.FunctionCatalogs); err != nil {
		return err
	}
	for range env.FunctionCatalogs {
		catalog, err := loadSentinelCatalog()
		if err != nil {
			return fmt.Errorf("load Sentinel function catalog: %w", err)
		}
		for name, parser := range catalog.Functions {
			// Explicit workspace declarations take precedence over bundled metadata.
			if _, exists := env.TabularFunctions[name]; exists {
				continue
			}
			if _, exists := env.ScalarFunctions[name]; exists || slices.Contains(env.MagicFunctions, name) {
				continue
			}
			env.TabularFunctions[name] = TabularFunctionDetails{
				OutputColumns: copyTableDetails(catalog.Schemas[parser.Schema]),
				Arguments:     copyArgumentDetails(parser.Arguments),
				IsOpen:        true,
			}
		}
		if _, tabular := env.TabularFunctions["_GetWatchlist"]; !tabular {
			if _, scalar := env.ScalarFunctions["_GetWatchlist"]; !scalar && !slices.Contains(env.MagicFunctions, "_GetWatchlist") {
				env.MagicFunctions = append(env.MagicFunctions, "_GetWatchlist")
			}
		}
	}
	return nil
}

func watchlistColumns() TableDetails {
	return TableDetails{
		"_DTItemId": "string", "LastUpdatedTimeUTC": "datetime",
		"SearchKey": "string", "WatchlistItem": "dynamic",
	}
}

func watchlistFunctionSymbol(watchlists map[string]WatchlistDetails) *kustolang.FunctionSymbol {
	watchlists = cloneWatchlists(watchlists)
	return kustolang.NewFunctionSymbol("_GetWatchlist", kustolang.NewCustomSignature(func(context kustolang.CustomReturnTypeContext) kustolang.Symbol {
		columns := watchlistColumns()
		open := true
		if len(context.ArgumentConstantValues) > 0 {
			if alias, ok := context.ArgumentConstantValues[0].(string); ok {
				if declared, exists := watchlists[alias]; exists {
					open = false
					for name, typ := range declared {
						columns[name] = typ
					}
				}
			}
		}
		return tableSymbolFromDetails("", columns).WithIsOpen(open)
	}, kustolang.TabularityTabular,
		kustolang.NewParameter("watchlistAlias", kustolang.ScalarTypes.String.Type()),
	))
}
