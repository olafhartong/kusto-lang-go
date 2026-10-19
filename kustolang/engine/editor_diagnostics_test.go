package kustolang

import "testing"

func TestDisabledDiagnosticsParsesSortsAndFiltersDiagnostics(t *testing.T) {
	filter := ParseDisabledDiagnostics("KS200, sev:Warning, KS100, sev:Suggestion, sev:Information")
	if got, want := filter.String(), "sev:Information,sev:Suggestion,sev:Warning,KS100,KS200"; got != want {
		t.Fatalf("filter text = %q, want %q", got, want)
	}

	if !filter.IsServerityDisabled(DiagnosticSeverityWarning) || !filter.IsSeverityDisabled(DiagnosticSeveritySuggestion) {
		t.Fatalf("expected warning/suggestion severities disabled: %s", filter.String())
	}
	if filter.IsServerityDisabled(DiagnosticSeverityError) || !filter.IsSeverityEnabled(DiagnosticSeverityError) {
		t.Fatalf("error severity should remain enabled")
	}
	if !filter.IsCodeDisabled("KS100") || filter.IsCodeEnabled("KS100") || !filter.IsCodeEnabled("KS999") {
		t.Fatalf("code enablement mismatch")
	}
	if !filter.IsDiagnosticDisabled(NewDiagnostic("KS999", "warning").WithSeverity(DiagnosticSeverityWarning)) {
		t.Fatalf("warning diagnostic should be disabled by severity")
	}
	if !filter.IsDiagnosticDisabled(NewDiagnostic("KS100", "code").WithSeverity(DiagnosticSeverityError)) {
		t.Fatalf("KS100 diagnostic should be disabled by code")
	}
	if !filter.IsDiagnosticEnabled(NewDiagnostic("KS999", "error").WithSeverity(DiagnosticSeverityError)) {
		t.Fatalf("unfiltered error diagnostic should be enabled")
	}
}

func TestDisabledDiagnosticsBuilderAndCanDisableRules(t *testing.T) {
	if !CanDisableSeverity(DiagnosticSeverityWarning) ||
		!CanDisableSeverity(DiagnosticSeveritySuggestion) ||
		!CanDisableSeverity(DiagnosticSeverityInformation) ||
		CanDisableSeverity(DiagnosticSeverityError) ||
		CanDisableSeverity(DiagnosticSeverityHidden) {
		t.Fatalf("severity disable rules changed")
	}
	if !CanDisableCode(KustoAnalyzerCodeAvoidUsingContains) ||
		!CanDisableCode(KustoAnalyzerCodeAvoidStrlenWithDynamic) ||
		CanDisableCode(KustoAnalyzerCodeAvoidUsingShortStringComparision) ||
		CanDisableCode("KS001") {
		t.Fatalf("code disable rules changed")
	}

	filter := ParseDisabledDiagnostics("sev:Warning,KS100")
	updated := filter.
		WithSeverityEnabled(DiagnosticSeverityWarning, true).
		WithSeverityEnabled(DiagnosticSeverityInformation, false).
		WithSeverityEnabled(DiagnosticSeverityError, false).
		WithCodeEnabled(KustoAnalyzerCodeAvoidUsingContains, false).
		WithCodeEnabled("KS001", false).
		WithCodeEnabled("KS100", true)

	if got, want := updated.String(), "sev:Information,KS500"; got != want {
		t.Fatalf("updated filter text = %q, want %q", got, want)
	}
	if filter.String() != "sev:Warning,KS100" {
		t.Fatalf("source filter mutated: %q", filter.String())
	}

	builder := NewDisabledDiagnosticsBuilder()
	builder.SetCodeEnabled(KustoAnalyzerCodeAvoidStrlenWithDynamic, false)
	builder.SetSeverityEnabled(DiagnosticSeveritySuggestion, false)
	if got, want := builder.String(), "sev:Suggestion,KS516"; got != want {
		t.Fatalf("builder text = %q, want %q", got, want)
	}
}

func TestDisabledDiagnosticsDefaultAndWhitespaceCodeBehavior(t *testing.T) {
	if DisabledDiagnosticsDefault.String() != "" ||
		DisabledDiagnosticsDefault.IsCodeDisabled("KS500") ||
		DisabledDiagnosticsDefault.IsServerityDisabled(DiagnosticSeverityWarning) {
		t.Fatalf("default disabled diagnostics should not filter anything")
	}

	blankCode := ParseDisabledDiagnostics(" ")
	if got := blankCode.String(); got != "" {
		t.Fatalf("blank-code text = %q, want empty joined string", got)
	}
	if !blankCode.IsCodeDisabled("") {
		t.Fatalf("whitespace item should match upstream trim-to-empty code behavior")
	}
}

func TestQuickInfoOptionsDefaultsAndCloneHelpers(t *testing.T) {
	options := QuickInfoOptionsDefault
	if !options.ShowDescriptions || !options.ShowDiagnostics || options.DiagnosticFilter.String() != "" {
		t.Fatalf("quick-info defaults = %#v", options)
	}

	filter := ParseDisabledDiagnostics("sev:Warning")
	changed := options.
		WithShowDescriptions(false).
		WithShowDiagnostics(false).
		WithDiagnosticFilter(&filter)
	if changed.ShowDescriptions || changed.ShowDiagnostics || changed.DiagnosticFilter.String() != "sev:Warning" {
		t.Fatalf("changed quick-info options = %#v", changed)
	}
	if !options.ShowDescriptions || !options.ShowDiagnostics || options.DiagnosticFilter.String() != "" {
		t.Fatalf("default options mutated = %#v", options)
	}

	reset := changed.WithDiagnosticFilter(nil)
	if reset.DiagnosticFilter.String() != "" {
		t.Fatalf("nil diagnostic filter should reset to default, got %q", reset.DiagnosticFilter.String())
	}
}
