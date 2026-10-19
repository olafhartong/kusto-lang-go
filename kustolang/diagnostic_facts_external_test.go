package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDiagnosticFactsFactories(t *testing.T) {
	if kustolang.DiagnosticFactsFactoryCount != 210 {
		t.Fatalf("factory count = %d", kustolang.DiagnosticFactsFactoryCount)
	}
	if kustolang.DiagnosticFactsUpstreamDeclarationCount != 205 || kustolang.DiagnosticFactsUpstreamFactoryCount != 197 {
		t.Fatalf("upstream inventory = %d declarations, %d factories", kustolang.DiagnosticFactsUpstreamDeclarationCount, kustolang.DiagnosticFactsUpstreamFactoryCount)
	}
	if kustolang.DiagnosticFactsCompatibilityFactoryCount != 13 {
		t.Fatalf("compatibility factory count = %d", kustolang.DiagnosticFactsCompatibilityFactoryCount)
	}
	compatibility := kustolang.DiagnosticFactsCompatibilityFactoryNames()
	if len(compatibility) != 13 {
		t.Fatalf("compatibility factory names = %#v", compatibility)
	}
	compatibility[0] = "changed"
	if kustolang.DiagnosticFactsCompatibilityFactoryNames()[0] == "changed" {
		t.Fatal("compatibility factory names are mutable")
	}
	if kustolang.DiagnosticFactsUpstreamSourceSHA256 != "9ac212ea5eccd1cb4062113736fdda5749928ceeb0e2e254d55829aeac031fa2" {
		t.Fatalf("upstream source hash = %q", kustolang.DiagnosticFactsUpstreamSourceSHA256)
	}
	tests := []struct {
		name        string
		diagnostic  kustolang.Diagnostic
		code        string
		description string
	}{
		{name: "missing text", diagnostic: kustolang.DiagnosticFactsGetMissingText("name"), code: "KS001", description: "Missing: name"},
		{name: "terms expected", diagnostic: kustolang.DiagnosticFactsGetTermsExpected("table", "function"), code: "KS005", description: "Expected one of: table, function"},
		{name: "missing expression", diagnostic: kustolang.DiagnosticFactsGetMissingExpression(), code: "KS006", description: "Missing expression"},
		{name: "argument range", diagnostic: kustolang.DiagnosticFactsGetFunctionExpectsArgumentCountRange("f", 1, 3), code: "KS120", description: "The function 'f' expects between 1 and 3 arguments."},
		{name: "unknown graph snapshot", diagnostic: kustolang.DiagnosticFactsGetNameDoesNotReferToAnyKnownGraphSnapshot("Latest", "Network"), code: "KS261", description: "The name 'Latest' does not refer to any known graph snapshot of graph model 'Network'."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.diagnostic.Code != tt.code || tt.diagnostic.Description != tt.description || tt.diagnostic.Message != tt.description {
				t.Fatalf("diagnostic = %#v", tt.diagnostic)
			}
		})
	}
}
