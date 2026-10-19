package kustolang

import (
	"strings"
	"testing"
)

func TestPreferUsingMaterializedViewIntrinsicAnalyzerReportsKnownDownsamplingView(t *testing.T) {
	text := "DailySignin | count"
	code := Parse(text)
	materialized := NewMaterializedViewSymbol("DailySignin", []*ColumnSymbol{NewColumnSymbol("Day", ScalarTypes.DateTime.Type())}, "SigninLogs | summarize count() by bin(TimeGenerated, 1d)")
	materialized = materialized.WithMaterializedViewKind(MaterializedViewKindDownsampling)
	name := parsedNameReferenceBySimpleName(t, code, "DailySignin")
	name.SetSemanticInfo(NewSemanticInfoForSymbol(materialized, &materialized.TypeSymbol))

	diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic ||
		diagnostic.Category != DiagnosticCategoryPerformance ||
		diagnostic.Severity != DiagnosticSeveritySuggestion ||
		diagnostic.Start != strings.Index(text, "DailySignin") ||
		diagnostic.Length != len("DailySignin") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestPreferUsingMaterializedViewIntrinsicAnalyzerAnalyzesUnknownViewQueryAndCachesDownsampling(t *testing.T) {
	text := "DailySignin | count"
	code := Parse(text)
	materialized := NewMaterializedViewSymbol(
		"DailySignin",
		[]*ColumnSymbol{NewColumnSymbol("Day", ScalarTypes.DateTime.Type())},
		"SigninLogs | summarize Count=count() by Day=bin(TimeGenerated, 1d)",
	)
	name := parsedNameReferenceBySimpleName(t, code, "DailySignin")
	name.SetSemanticInfo(NewSemanticInfoForSymbol(materialized, &materialized.TypeSymbol))

	diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if materialized.MaterializedViewKind() != MaterializedViewKindDownsampling {
		t.Fatalf("materialized view kind = %s, want downsampling", materialized.MaterializedViewKind())
	}
}

func TestPreferUsingMaterializedViewIntrinsicAnalyzerCachesOtherForNonDownsamplingView(t *testing.T) {
	code := Parse("DailySignin | count")
	materialized := NewMaterializedViewSymbol(
		"DailySignin",
		[]*ColumnSymbol{NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type())},
		"SigninLogs | summarize Count=count() by UserPrincipalName",
	)
	name := parsedNameReferenceBySimpleName(t, code, "DailySignin")
	name.SetSemanticInfo(NewSemanticInfoForSymbol(materialized, &materialized.TypeSymbol))

	diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
	if materialized.MaterializedViewKind() != MaterializedViewKindOther {
		t.Fatalf("materialized view kind = %s, want other", materialized.MaterializedViewKind())
	}
}

func TestPreferUsingMaterializedViewIntrinsicAnalyzerSkipsNonQueryUnboundAndOtherViews(t *testing.T) {
	command := Parse("DailySignin | count")
	command.Kind = CodeKindCommand
	materialized := NewMaterializedViewSymbol("DailySignin", nil, "SigninLogs | summarize count() by bin(TimeGenerated, 1d)").WithMaterializedViewKind(MaterializedViewKindDownsampling)
	name := parsedNameReferenceBySimpleName(t, command, "DailySignin")
	name.SetSemanticInfo(NewSemanticInfoForSymbol(materialized, &materialized.TypeSymbol))
	if diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(command, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("command diagnostics = %#v, want empty", diagnostics)
	}

	unbound := Parse("DailySignin | count")
	if diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	otherCode := Parse("DailySignin | count")
	other := NewMaterializedViewSymbol("DailySignin", nil, "SigninLogs | summarize count() by bin(TimeGenerated, 1d)").WithMaterializedViewKind(MaterializedViewKindOther)
	otherName := parsedNameReferenceBySimpleName(t, otherCode, "DailySignin")
	otherName.SetSemanticInfo(NewSemanticInfoForSymbol(other, &other.TypeSymbol))
	if diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Analyze(otherCode, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("other diagnostics = %#v, want empty", diagnostics)
	}
}

func TestPreferUsingMaterializedViewIntrinsicAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewPreferUsingMaterializedViewIntrinsicAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic ||
		diagnostics[0].Category != DiagnosticCategoryPerformance ||
		diagnostics[0].Severity != DiagnosticSeveritySuggestion ||
		diagnostics[0].Description != "Prefer using the `materialized_view()` function where appropriate" ||
		diagnostics[0].Message != "Consider using the `materialized_view()` function to query the materialized part only\nSee: https://aka.ms/mvqueries" {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func TestPreferMaterializedViewQueryIsDownsamplingRequiresSingleSummarizeWithFunctionByKey(t *testing.T) {
	if !preferMaterializedViewQueryIsDownsampling("SigninLogs | summarize count() by startofday(TimeGenerated)", DefaultGlobalState) {
		t.Fatal("startofday by key should be downsampling")
	}
	if preferMaterializedViewQueryIsDownsampling("SigninLogs | extend Day=startofday(TimeGenerated) | summarize count() by Day", DefaultGlobalState) {
		t.Fatal("extended key should not be detected as direct downsampling")
	}
	if preferMaterializedViewQueryIsDownsampling("SigninLogs | summarize count() by UserPrincipalName", DefaultGlobalState) {
		t.Fatal("plain by key should not be downsampling")
	}
	if preferMaterializedViewQueryIsDownsampling("SigninLogs | summarize count() by startofday(TimeGenerated) | summarize count()", DefaultGlobalState) {
		t.Fatal("multiple summarize operators should not be downsampling")
	}
	if preferMaterializedViewQueryIsDownsampling("SigninLogs | summarize count() by startofday(TimeGenerated); SigninLogs | count", DefaultGlobalState) {
		t.Fatal("only the last query statement should be considered")
	}
}
