package kustolang_test

import (
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicProjectionTabularAndMalformedRecovery(t *testing.T) {
	text := "let T=datatable(A:string)[]; T | project X=T"
	code := kustolang.ParseAndAnalyze(text)
	diagnostics := code.GetDiagnostics()
	start := strings.LastIndex(text, "X=T")
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS108" || diagnostics[0].Start != start || diagnostics[0].Length != len("X=T") {
		t.Fatalf("tabular projection diagnostics = %#v", diagnostics)
	}
	result, ok := kustolang.TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("tabular projection result = %#v", code.ResultType)
	}
	x := result.GetColumn("X")
	if x == nil || x.Type.SymbolTabularity() != kustolang.TabularityTabular {
		t.Fatalf("recovered X = %#v, want tabular column", x)
	}

	malformedText := "datatable(A:string)[] | extend X="
	malformed := kustolang.ParseAndAnalyze(malformedText)
	malformedDiagnostics := malformed.GetDiagnostics()
	if len(malformedDiagnostics) != 1 || malformedDiagnostics[0].Code != "KS006" || malformedDiagnostics[0].Start != len(malformedText) || malformedDiagnostics[0].Length != 0 {
		t.Fatalf("malformed projection diagnostics = %#v", malformedDiagnostics)
	}
	malformedResult, ok := kustolang.TableSymbolForType(malformed.ResultType)
	if !ok || malformedResult.GetColumn("A") == nil {
		t.Fatalf("malformed extend result = %#v", malformed.ResultType)
	}
	if recovered := malformedResult.GetColumn("X"); recovered == nil || recovered.Type != kustolang.ScalarTypes.Unknown.Type() {
		t.Fatalf("malformed extend X = %#v, want unknown", recovered)
	}
}
