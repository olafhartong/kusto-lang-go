package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDisabledDiagnosticConstructionAndCapabilities(t *testing.T) {
	disabled := kustolang.NewDisabledDiagnostics("sev:Warning,KS500")
	if !disabled.IsSeverityDisabled(kustolang.DiagnosticSeverityWarning) || !disabled.IsCodeDisabled(kustolang.KustoAnalyzerCodeAvoidUsingContains) {
		t.Fatalf("disabled diagnostics = %q", disabled.String())
	}
	if !kustolang.CanDisableSeverity(kustolang.DiagnosticSeverityWarning) || kustolang.CanDisableSeverity(kustolang.DiagnosticSeverityError) ||
		!kustolang.CanDisableCode(kustolang.KustoAnalyzerCodeAvoidUsingContains) || kustolang.CanDisableCode("KS001") {
		t.Fatal("disabled diagnostic capability probes do not match analyzer policy")
	}
	builder := kustolang.NewDisabledDiagnosticsBuilder(disabled)
	builder.SetSeverityEnabled(kustolang.DiagnosticSeverityWarning, true)
	builder.SetCodeEnabled(kustolang.KustoAnalyzerCodeAvoidUsingContains, true)
	if got := builder.ToDisabledDiagnostics(); got.String() != "" {
		t.Fatalf("builder result = %q", got.String())
	}
}

func TestPublicTextRangeAndEditFactories(t *testing.T) {
	rangeValue := kustolang.TextRangeFromBounds(4, 9)
	if rangeValue.Start != 4 || rangeValue.Length != 5 || rangeValue.End() != 9 || !rangeValue.Overlaps(kustolang.NewTextRange(9, 2)) ||
		!kustolang.TextRangesOverlap(4, 5, 9, 2) {
		t.Fatalf("text range = %#v", rangeValue)
	}
	replacement := kustolang.NewTextReplacement(4, 5, "value")
	deletion := kustolang.NewTextDeletion(4, 5)
	insertion := kustolang.NewTextInsertion(4, "value")
	if replacement.Start != 4 || replacement.DeleteLength != 5 || replacement.InsertText != "value" ||
		deletion.Start != 4 || deletion.DeleteLength != 5 || deletion.InsertText != "" ||
		insertion.Start != 4 || insertion.DeleteLength != 0 || insertion.InsertText != "value" {
		t.Fatalf("edits = replacement:%#v deletion:%#v insertion:%#v", replacement, deletion, insertion)
	}
}
