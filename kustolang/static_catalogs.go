package kustolang

// Ported from:
//   src/Kusto.Language/Diagnostics/Diagnostic.cs
//   src/Kusto.Language/Editor/CodeKinds.cs
//   src/Kusto.Language/Editor/CodeServiceFeatures.cs
//   src/Kusto.Language/Editor/Kusto/KustoActors.cs
//   src/Kusto.Language/ServerKinds.cs
// at c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// CodeKindsAPI mirrors the upstream static CodeKinds class.
type CodeKindsAPI struct {
	Query     string
	Command   string
	Directive string
	Unknown   string
}

// CodeKinds exposes the known editor code-kind names.
var CodeKinds = CodeKindsAPI{
	Query:     CodeKindsQuery,
	Command:   CodeKindsCommand,
	Directive: CodeKindsDirective,
	Unknown:   CodeKindsUnknown,
}

// CodeServiceFeaturesAPI mirrors the upstream static CodeServiceFeatures
// class.
type CodeServiceFeaturesAPI struct {
	Diagnostics        string
	Classification     string
	Completion         string
	Outlining          string
	QuickInfo          string
	ClusterReferences  string
	DatabaseReferences string
	MinimalText        string
	Formatting         string
	ClientParameters   string
}

// CodeServiceFeatures exposes editor feature names accepted by CodeService.
var CodeServiceFeatures = CodeServiceFeaturesAPI{
	Diagnostics:        CodeServiceFeatureDiagnostics,
	Classification:     CodeServiceFeatureClassification,
	Completion:         CodeServiceFeatureCompletion,
	Outlining:          CodeServiceFeatureOutlining,
	QuickInfo:          CodeServiceFeatureQuickInfo,
	ClusterReferences:  CodeServiceFeatureClusterReferences,
	DatabaseReferences: CodeServiceFeatureDatabaseReferences,
	MinimalText:        CodeServiceFeatureMinimalText,
	Formatting:         CodeServiceFeatureFormatting,
	ClientParameters:   CodeServiceFeatureClientParameters,
}

// DiagnosticCategoryAPI mirrors the upstream static DiagnosticCategory class.
type DiagnosticCategoryAPI struct {
	General     string
	Correctness string
	Performance string
}

// DiagnosticCategory exposes the standard diagnostic categories.
var DiagnosticCategory = DiagnosticCategoryAPI{
	General:     DiagnosticCategoryGeneral,
	Correctness: DiagnosticCategoryCorrectness,
	Performance: DiagnosticCategoryPerformance,
}

// DiagnosticSeverityAPI mirrors the upstream static DiagnosticSeverity class.
type DiagnosticSeverityAPI struct {
	Error       string
	Warning     string
	Suggestion  string
	Information string
	Hidden      string
}

// DiagnosticSeverity exposes the standard diagnostic severities.
var DiagnosticSeverity = DiagnosticSeverityAPI{
	Error:       DiagnosticSeverityError,
	Warning:     DiagnosticSeverityWarning,
	Suggestion:  DiagnosticSeveritySuggestion,
	Information: DiagnosticSeverityInformation,
	Hidden:      DiagnosticSeverityHidden,
}

// ServerKindsAPI mirrors the upstream static ServerKinds class.
type ServerKindsAPI struct {
	Engine         string
	DataManager    string
	ClusterManager string
	AriaBridge     string
}

// ServerKinds exposes the command server-kind names used by GlobalState.
var ServerKinds = ServerKindsAPI{
	Engine:         ServerKindEngine,
	DataManager:    ServerKindDataManager,
	ClusterManager: ServerKindClusterManager,
	AriaBridge:     ServerKindAriaBridge,
}

// KustoActorNamesAPI mirrors the upstream static KustoActorNames class.
type KustoActorNamesAPI struct {
	FunctionInliner      string
	ExtractExpression    string
	ConvertToMacroExpand string
	DiagnosticFixer      string
	Rename               string
}

// KustoActorNames exposes editor actor names accepted by action APIs.
var KustoActorNames = KustoActorNamesAPI{
	FunctionInliner:      KustoActorNameFunctionInliner,
	ExtractExpression:    KustoActorNameExtractExpression,
	ConvertToMacroExpand: KustoActorNameConvertToMacroExpand,
	DiagnosticFixer:      KustoActorNameDiagnosticFixer,
	Rename:               KustoActorNameRename,
}
