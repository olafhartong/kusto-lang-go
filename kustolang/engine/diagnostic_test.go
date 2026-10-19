package kustolang

import (
	"encoding/json"
	"testing"
)

func TestDiagnosticDefaultsMatchKustoLanguage(t *testing.T) {
	d := NewDiagnostic("KS211", "The name 'x' does not refer to any known table.")
	if d.Code != "KS211" {
		t.Fatalf("Code = %q", d.Code)
	}
	if d.Category != DiagnosticCategoryGeneral {
		t.Fatalf("Category = %q", d.Category)
	}
	if d.Severity != DiagnosticSeverityError {
		t.Fatalf("Severity = %q", d.Severity)
	}
	if d.Description != d.Message {
		t.Fatalf("Description = %q, Message = %q", d.Description, d.Message)
	}
	if d.HasLocation() {
		t.Fatal("new diagnostic should not have an absolute location")
	}
}

func TestDiagnosticWithLocation(t *testing.T) {
	d := NewDiagnostic("SYN001", "bad syntax").WithLocation(4, 7)
	if !d.HasLocation() {
		t.Fatal("expected absolute location")
	}
	if d.Start != 4 || d.Length != 7 || d.End() != 11 {
		t.Fatalf("location = start %d length %d end %d", d.Start, d.Length, d.End())
	}
}

func TestDiagnosticJSONPreservesZeroLocationCoordinates(t *testing.T) {
	data, err := json.Marshal(NewDiagnostic("KS006", "Missing name").WithLocation(12, 0))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["start"] != float64(12) || got["length"] != float64(0) {
		t.Fatalf("diagnostic JSON = %s", data)
	}
}

func TestDiagnosticFunctionNotDefinedWithMatchingParameters(t *testing.T) {
	d := DiagnosticFunctionNotDefinedWithMatchingParameters("bin", "string", "timespan")
	if d.Code != "KS127" || d.Message != "The function 'bin' is not compatible with arguments (string, timespan)" {
		t.Fatalf("diagnostic = %#v", d)
	}
}

func TestDiagnosticDistinctNewDiagnosticsUsesUpstreamEquality(t *testing.T) {
	base := NewDiagnosticDetailed("KS900", DiagnosticCategoryCorrectness, DiagnosticSeverityError, "description", "message").WithLocation(4, 2)
	existing := []Diagnostic{base}
	additions := []Diagnostic{
		base.WithCategory(DiagnosticCategoryPerformance).WithSeverity(DiagnosticSeverityWarning).WithDescription("other"),
		base.WithLocation(4, 3),
		base.WithMessage("different"),
		base.WithLocationKind(DiagnosticLocationRelative),
		base.WithLocationKind(DiagnosticLocationRelative),
	}

	got := kustoCodeServiceDistinctNewDiagnostics(existing, additions)
	if len(got) != 4 {
		t.Fatalf("distinct diagnostics = %#v", got)
	}
	if got[0].Length != 3 || got[1].Message != "different" ||
		got[2].LocationKind != DiagnosticLocationRelative || got[3].LocationKind != DiagnosticLocationRelative {
		t.Fatalf("distinct diagnostic ordering = %#v", got)
	}
}
