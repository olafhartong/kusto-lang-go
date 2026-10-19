package kustolang

import (
	"strings"
	"testing"
)

func TestDiagnosticFuzzyUnionOperandDoesNotReferToAnyKnownTable(t *testing.T) {
	diagnostic := DiagnosticFuzzyUnionOperandDoesNotReferToAnyKnownTable("ZMissing")
	if diagnostic.Code != "KS205" ||
		diagnostic.Message != "The fuzzy union operand 'ZMissing' does not refer to any known table, tabular variable or function." ||
		diagnostic.Severity != DiagnosticSeverityWarning ||
		diagnostic.LocationKind != DiagnosticLocationRelative ||
		diagnostic.LocationText() != "ZMissing" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestDiagnosticFuzzyTableNotDefined(t *testing.T) {
	diagnostic := DiagnosticFuzzyTableNotDefined("ZMissing")
	if diagnostic.Code != "KS205" ||
		diagnostic.Message != "The fuzzy name 'ZMissing' does not refer to any known or currently accessible table." ||
		diagnostic.Severity != DiagnosticSeverityWarning ||
		diagnostic.LocationKind != DiagnosticLocationRelative ||
		diagnostic.LocationText() != "ZMissing" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestDiagnosticFuzzySourceEntityKinds(t *testing.T) {
	tests := []struct {
		diagnostic Diagnostic
		kind       string
	}{
		{DiagnosticFuzzyClusterNotDefined("C"), "cluster"},
		{DiagnosticFuzzyDatabaseNotDefined("D"), "database"},
		{DiagnosticFuzzyExternalTableNotDefined("E"), "external table"},
		{DiagnosticFuzzyMaterializedViewNotDefined("M"), "materialized view"},
		{DiagnosticFuzzyEntityGroupNotDefined("G"), "entity group"},
		{DiagnosticFuzzyFunctionNotDefined("F"), "function"},
		{DiagnosticFuzzyStoredQueryResultNotDefined("S"), "stored query result"},
	}
	for _, test := range tests {
		if test.diagnostic.Code != "KS205" || test.diagnostic.Severity != DiagnosticSeverityWarning ||
			!strings.Contains(test.diagnostic.Message, "currently accessible "+test.kind) {
			t.Fatalf("%s diagnostic = %#v", test.kind, test.diagnostic)
		}
	}
}
