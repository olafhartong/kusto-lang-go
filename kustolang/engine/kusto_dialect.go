package kustolang

// Ported from src/Kusto.Language/KustoDialect.cs and ServerKinds.cs.

type KustoDialect int

const (
	KustoDialectUnknown KustoDialect = iota
	KustoDialectClusterManagerCommand
	KustoDialectDataManagerCommand
	KustoDialectEngineCommand
	KustoDialectQuery
)

const (
	ServerKindEngine         = "Engine"
	ServerKindDataManager    = "DataManager"
	ServerKindClusterManager = "ClusterManager"
	ServerKindAriaBridge     = "AriaBridge"
)
