package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoActors.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	KustoActorNameFunctionInliner      = "InlineDatabaseFunctionActor"
	KustoActorNameExtractExpression    = "ExtractExpressionActor"
	KustoActorNameConvertToMacroExpand = "ConvertToMacroExpandActor"
	KustoActorNameDiagnosticFixer      = "DiagnosticFixActor"
	KustoActorNameRename               = "RenameActor"
)

func KustoActorNamesAll() []string {
	return []string{
		KustoActorNameFunctionInliner,
		KustoActorNameExtractExpression,
		KustoActorNameConvertToMacroExpand,
		KustoActorNameRename,
		KustoActorNameDiagnosticFixer,
	}
}
