package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingLegacyPartitionAnalyzerReportsImplicitAndExplicitLegacy(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "implicit legacy",
			text: "T | partition by User (summarize Count=count() by App)",
		},
		{
			name: "explicit legacy",
			text: "T | partition hint.strategy=legacy by User (summarize Count=count() by App)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := NewAvoidUsingLegacyPartitionAnalyzer().Analyze(Parse(tt.text), nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != KustoAnalyzerCodeAvoidUsingLegacyPartition ||
				diagnostic.Category != DiagnosticCategoryGeneral ||
				diagnostic.Severity != DiagnosticSeveritySuggestion ||
				diagnostic.Start != strings.Index(tt.text, "partition") ||
				diagnostic.Length != len(tt.text)-strings.Index(tt.text, "partition") {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestAvoidUsingLegacyPartitionAnalyzerSkipsNativeShuffleKeyAndBraceQuery(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "native",
			text: "T | partition hint.strategy=native by User (summarize Count=count() by App)",
		},
		{
			name: "shuffle",
			text: "T | partition hint.strategy=shuffle by User (summarize Count=count() by App)",
		},
		{
			name: "shuffle key",
			text: "T | partition hint.shufflekey=User by User (summarize Count=count() by App)",
		},
		{
			name: "brace query",
			text: "T | partition by User { summarize Count=count() by App }",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diagnostics := NewAvoidUsingLegacyPartitionAnalyzer().Analyze(Parse(tt.text), nil, CancellationToken{}); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want empty", diagnostics)
			}
		})
	}
}

func TestAvoidUsingLegacyPartitionAnalyzerSkipsInvalidNonLiteralStrategy(t *testing.T) {
	text := "T | partition hint.strategy=(Strategy) by User (summarize Count=count() by App)"
	if diagnostics := NewAvoidUsingLegacyPartitionAnalyzer().Analyze(Parse(text), nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingLegacyPartitionAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingLegacyPartitionAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	want := "It is recommended to use the `native` or `shuffle` strategy of the partition operator. The legacy mode is limited to 64 partitions and is less efficient than the `native` or `shuffle` strategies."
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingLegacyPartition ||
		diagnostics[0].Description != want ||
		diagnostics[0].Message != want {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}
