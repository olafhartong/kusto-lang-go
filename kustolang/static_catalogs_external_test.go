package kustolang_test

// External-package parity tests for pinned static string catalogs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"slices"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCodeKindsAndServiceFeaturesCatalogs(t *testing.T) {
	wantKinds := kustolang.CodeKindsAPI{
		Query:     "Query",
		Command:   "Command",
		Directive: "Directive",
		Unknown:   "Unknown",
	}
	if !reflect.DeepEqual(kustolang.CodeKinds, wantKinds) {
		t.Fatalf("CodeKinds = %#v, want %#v", kustolang.CodeKinds, wantKinds)
	}

	wantFeatures := kustolang.CodeServiceFeaturesAPI{
		Diagnostics:        "Diagnostics",
		Classification:     "Classification",
		Completion:         "Completion",
		Outlining:          "Outlining",
		QuickInfo:          "QuickInfo",
		ClusterReferences:  "ClusterReferences",
		DatabaseReferences: "DatabaseReferences",
		MinimalText:        "MinimalText",
		Formatting:         "Formatting",
		ClientParameters:   "ClientParameters",
	}
	if !reflect.DeepEqual(kustolang.CodeServiceFeatures, wantFeatures) {
		t.Fatalf("CodeServiceFeatures = %#v, want %#v", kustolang.CodeServiceFeatures, wantFeatures)
	}

	service := kustolang.NewKustoCodeService("print x=1")
	if service.Kind() != kustolang.CodeKinds.Query ||
		!service.IsFeatureSupported(kustolang.CodeServiceFeatures.Completion, 0) ||
		!service.IsFeatureSupported(kustolang.CodeServiceFeatures.Diagnostics, 0) {
		t.Fatalf("service catalog integration = kind:%q completion:%t diagnostics:%t",
			service.Kind(),
			service.IsFeatureSupported(kustolang.CodeServiceFeatures.Completion, 0),
			service.IsFeatureSupported(kustolang.CodeServiceFeatures.Diagnostics, 0))
	}
}

func TestPublicDiagnosticCatalogsIntegrateWithDiagnostics(t *testing.T) {
	wantCategories := kustolang.DiagnosticCategoryAPI{
		General:     "General",
		Correctness: "Correctness",
		Performance: "Performance",
	}
	if !reflect.DeepEqual(kustolang.DiagnosticCategory, wantCategories) {
		t.Fatalf("DiagnosticCategory = %#v, want %#v", kustolang.DiagnosticCategory, wantCategories)
	}
	wantSeverities := kustolang.DiagnosticSeverityAPI{
		Error:       "Error",
		Warning:     "Warning",
		Suggestion:  "Suggestion",
		Information: "Information",
		Hidden:      "Hidden",
	}
	if !reflect.DeepEqual(kustolang.DiagnosticSeverity, wantSeverities) {
		t.Fatalf("DiagnosticSeverity = %#v, want %#v", kustolang.DiagnosticSeverity, wantSeverities)
	}

	diagnostic := kustolang.NewDiagnosticDetailed(
		"APP900",
		kustolang.DiagnosticCategory.Correctness,
		kustolang.DiagnosticSeverity.Warning,
		"description",
		"message",
	)
	if diagnostic.Category != "Correctness" || diagnostic.Severity != "Warning" {
		t.Fatalf("catalog diagnostic = %#v", diagnostic)
	}
}

func TestPublicServerKindsAndActorNamesCatalogs(t *testing.T) {
	wantServerKinds := kustolang.ServerKindsAPI{
		Engine:         "Engine",
		DataManager:    "DataManager",
		ClusterManager: "ClusterManager",
		AriaBridge:     "AriaBridge",
	}
	if !reflect.DeepEqual(kustolang.ServerKinds, wantServerKinds) {
		t.Fatalf("ServerKinds = %#v, want %#v", kustolang.ServerKinds, wantServerKinds)
	}
	if globals := kustolang.DefaultGlobalState.WithServerKind(kustolang.ServerKinds.DataManager); globals.ServerKind != "DataManager" {
		t.Fatalf("server-kind catalog integration = %q", globals.ServerKind)
	}

	wantActors := kustolang.KustoActorNamesAPI{
		FunctionInliner:      "InlineDatabaseFunctionActor",
		ExtractExpression:    "ExtractExpressionActor",
		ConvertToMacroExpand: "ConvertToMacroExpandActor",
		DiagnosticFixer:      "DiagnosticFixActor",
		Rename:               "RenameActor",
	}
	if !reflect.DeepEqual(kustolang.KustoActorNames, wantActors) {
		t.Fatalf("KustoActorNames = %#v, want %#v", kustolang.KustoActorNames, wantActors)
	}
	all := kustolang.KustoActorNamesAll()
	for _, actor := range []string{
		kustolang.KustoActorNames.FunctionInliner,
		kustolang.KustoActorNames.ExtractExpression,
		kustolang.KustoActorNames.ConvertToMacroExpand,
		kustolang.KustoActorNames.DiagnosticFixer,
		kustolang.KustoActorNames.Rename,
	} {
		if !slices.Contains(all, actor) {
			t.Fatalf("actor catalog value %q missing from %#v", actor, all)
		}
	}
}

func TestPublicStaticCatalogCopiesDoNotMutateGlobals(t *testing.T) {
	kinds := kustolang.CodeKinds
	kinds.Query = "Changed"
	features := kustolang.CodeServiceFeatures
	features.Completion = "Changed"
	if kustolang.CodeKinds.Query != "Query" || kustolang.CodeServiceFeatures.Completion != "Completion" {
		t.Fatal("copy mutation changed a package catalog")
	}
}
