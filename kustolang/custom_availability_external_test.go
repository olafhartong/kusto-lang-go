package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCustomAvailabilityReceivesBinderLocation(t *testing.T) {
	var locations []kustolang.SyntaxElement
	function := kustolang.NewScalarFunctionSymbol("available_here", kustolang.ScalarTypes.Long.Type()).
		WithCustomAvailability(func(context kustolang.CustomAvailabilityContext) bool {
			locations = append(locations, context.Location)
			return context.Location != nil
		})
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", function))

	valid := kustolang.ParseAndAnalyze(`print value=available_here()`, globals)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid custom function diagnostics = %#v", diagnostics)
	}
	malformed := kustolang.ParseAndAnalyze(`print value=available_here(`, globals)
	if diagnostics := malformed.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS005") {
		t.Fatalf("malformed custom function diagnostics = %#v", diagnostics)
	}
	if len(locations) < 2 {
		t.Fatalf("availability callback locations = %#v", locations)
	}
	for index, location := range locations {
		if location == nil {
			t.Fatalf("availability callback location[%d] is nil", index)
		}
	}
	if function.CustomAvailability() == nil || !function.IsAvailable(kustolang.CustomAvailabilityContext{Location: valid.Syntax()}) {
		t.Fatal("public custom availability metadata is not callable")
	}
}

func TestPublicOptimizedAlternativeAnalyzerValidAndMalformed(t *testing.T) {
	validText := `datatable(x:long)[1] | summarize count_distinct(x)`
	valid := kustolang.NewKustoCodeService(validText)
	diagnostics := valid.GetAnalyzerDiagnostics(true, kustolang.CancellationToken{})
	if !hasDiagnostic(diagnostics, "KS515") {
		t.Fatalf("optimized-alternative diagnostics = %#v, want KS515", diagnostics)
	}

	malformedText := `datatable(x:long)[1] | summarize count_distinct(`
	malformed := kustolang.NewKustoCodeService(malformedText)
	if diagnostics := malformed.GetDiagnostics(true, kustolang.CancellationToken{}); !hasDiagnostic(diagnostics, "KS005") {
		t.Fatalf("malformed optimized call diagnostics = %#v", diagnostics)
	}
	if diagnostics := malformed.GetAnalyzerDiagnostics(true, kustolang.CancellationToken{}); !hasDiagnostic(diagnostics, "KS515") {
		t.Fatalf("recovered malformed call lost optimization diagnostic = %#v", diagnostics)
	}
}
