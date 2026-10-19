# Sentinel functions and ASIM parsers

The `sentinel` and derived `m365_with_sentinel` environments load an offline
Sentinel function catalog by default. This applies to both `analyzer` and the
standalone `kustolang/environment` loader, so embedding applications receive the same
function signatures, return types, completion items, and quick information.
No workspace credentials or network connection are required during analysis.

The catalog contains `_GetWatchlist` and ASIM unifying and source-specific
parsers. ASIM metadata comes from Microsoft's
[Azure-Sentinel repository](https://github.com/Azure/Azure-Sentinel):
`Parsers/ASim*/Parsers/*.yaml` supplies names and parameter signatures, and
`ASIM/dev/ASimTester/ASimTester.csv` supplies normalized field types. The bundled
snapshot records the upstream commit and currently includes 474 definitions
with 948 built-in and deployable names across 14 schemas.

Built-in names such as `_Im_Dns` and `_ASim_Dns` are available in Sentinel.
Deployable names such as `imDns` and `vimDnsAzureFirewall` are also recognized;
using those names in a real workspace requires deploying the corresponding
parsers. Recognition is a language model capability, not a check that a parser
or its source tables are deployed. Historical version-specific names absent
from the snapshot can be supplied through workspace function declarations.

## Select the catalog

Custom environment names and programmatically constructed environments can
enable the same defaults explicitly:

```json
{
  "my_workspace": {
    "function_catalogs": ["sentinel"],
    "tables": {}
  }
}
```

In Go, set `EnvironmentDefinition.FunctionCatalogs` to `[]string{"sentinel"}`
before calling `kustolang.NewGlobalStateForEnvironment` or creating an analyzer
session. Other environments, including plain `m365`, do not acquire Sentinel
functions automatically. An explicit `"function_catalogs": []` in a `sentinel`
environment file disables these defaults. Unknown catalog names are rejected.

Explicit `tabular_functions`, `scalar_functions`, or `magic_functions`
declarations take precedence over bundled definitions. For workspace-specific
signatures and columns, use those existing environment configuration fields.
Reload the environment or restart the service after editing its configuration.

ASIM return tables contain known normalized columns and are open to additional
source-specific fields. Custom tabular functions can set `"is_open": true`
when their complete output schema is unavailable; the default remains closed.
Parameter names, types, order, and optionality come from the YAML definitions.
Default values are used to determine optionality, rather than executed.
These are database functions, so the native binder retains its normal
permissive scalar argument conversion rules.

## Watchlist schemas

`_GetWatchlist(watchlistAlias)` accepts one string argument. The language model
knows `_DTItemId`, `LastUpdatedTimeUTC`, `SearchKey`, and `WatchlistItem`. Without
a matching watchlist schema, additional columns remain open because the
watchlist contents are workspace-specific. Supply `local_data.watchlists` to
give them precise types and enable unknown-column diagnostics:

```json
{
  "watchlists": {
    "VIPUsers": {"Account": "string", "Score": "long"}
  }
}
```

The schema is selected when the alias is known at analysis time. It also works
in the native language layer and in prepared analyzer sessions.

## Refresh ASIM metadata

Use a local checkout of Microsoft's repository and regenerate the snapshot:

```sh
git clone --depth 1 --filter=blob:none --sparse \
  https://github.com/Azure/Azure-Sentinel.git /tmp/azure-sentinel
git -C /tmp/azure-sentinel sparse-checkout set --no-cone \
  '/ASIM/dev/ASimTester/ASimTester.csv' '/Parsers/ASim*/Parsers/*.yaml'
go run ./tools/gen-asim-catalog --source-dir /tmp/azure-sentinel
go test ./kustolang ./kustolang/environment ./tools/gen-asim-catalog
```

To refresh an existing checkout, fetch and check out the desired upstream
revision first. The generator reads files locally, records the checkout's
commit, preserves false and null parameter defaults as optional arguments, and
fails on duplicate names, conflicting column types, or unknown schemas. It
does not download files or execute parser bodies. JSON output is deterministic.
Use `--output` to write a snapshot elsewhere for inspection.

Rebuild the consuming application or library after regenerating
`kustolang/environment/sentinel_catalog.json`, then recreate its application
global state or restart its service. The bundled catalog is compiled into the
language layer; ordinary workspace function overrides remain reloadable from
environment JSON without rebuilding. Parser signatures and normalized schemas
describe their public contract; parser bodies and underlying table dependencies
are not evaluated or expanded.
