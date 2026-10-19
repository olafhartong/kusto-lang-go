package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidJoinWithoutKindAnalyzerReportsJoinWithCondition(t *testing.T) {
	text := "T | join (U) on Key"
	analyzer := NewAvoidJoinWithoutKindAnalyzer()
	diagnostics := analyzer.Analyze(Parse(text), nil, CancellationToken{})

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeAvoidJoinWithoutKind ||
		diagnostic.Category != DiagnosticCategoryCorrectness ||
		diagnostic.Severity != DiagnosticSeverityWarning ||
		diagnostic.Start != strings.Index(text, "join") ||
		diagnostic.Length != len("join") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestAvoidJoinWithoutKindAnalyzerSkipsKindParameter(t *testing.T) {
	text := "T | join kind=inner (U) on Key"
	diagnostics := NewAvoidJoinWithoutKindAnalyzer().Analyze(Parse(text), nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidJoinWithoutKindAnalyzerSkipsJoinWithoutConditionClause(t *testing.T) {
	text := "T | join (U)"
	diagnostics := NewAvoidJoinWithoutKindAnalyzer().Analyze(Parse(text), nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidJoinWithoutKindAnalyzerSkipsJoinWithMissingOnKeyword(t *testing.T) {
	text := "T | join (U) Key"
	diagnostics := NewAvoidJoinWithoutKindAnalyzer().Analyze(Parse(text), nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidJoinWithoutKindAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidJoinWithoutKindAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidJoinWithoutKind ||
		diagnostics[0].Description != "Avoid using joins without specifying the join kind." ||
		diagnostics[0].Message != "Avoid using joins without specifying the join kind. The default join behavior may be unexpected." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}
