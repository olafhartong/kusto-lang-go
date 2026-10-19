package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingShortStringComparisionAnalyzerReportsFirstShortDatabaseComparisonOperand(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		literal string
	}{
		{name: "equal", text: `T | where Account == "ab"`, literal: `"ab"`},
		{name: "not equal left constant", text: `T | where "ab" != Account`, literal: `"ab"`},
		{name: "has", text: `T | where Account has "ab"`, literal: `"ab"`},
		{name: "not has", text: `T | where Account !has "ab"`, literal: `"ab"`},
		{name: "has cs", text: `T | where Account has_cs "ab"`, literal: `"ab"`},
		{name: "not has cs", text: `T | where Account !has_cs "ab"`, literal: `"ab"`},
		{name: "starts with", text: `T | where Account startswith "ab"`, literal: `"ab"`},
		{name: "not starts with", text: `T | where Account !startswith "ab"`, literal: `"ab"`},
		{name: "has any list", text: `T | where Account has_any ("long", "ab", "x")`, literal: `"ab"`},
		{name: "has all list", text: `T | where Account has_all ("long", "ab", "x")`, literal: `"ab"`},
		{name: "UTF-16 length three", text: `T | where Account == "😀a"`, literal: `"😀a"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := shortStringAnalyzerBoundCode(t, tt.text, true)
			diagnostics := NewAvoidUsingShortStringComparisionAnalyzer().Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != KustoAnalyzerCodeAvoidUsingShortStringComparision ||
				diagnostic.Category != DiagnosticCategoryPerformance ||
				diagnostic.Severity != DiagnosticSeveritySuggestion ||
				diagnostic.Start != strings.Index(tt.text, tt.literal) ||
				diagnostic.Length != len(tt.literal) {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestAvoidUsingShortStringComparisionAnalyzerNegativeControls(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		databaseColumn bool
	}{
		{name: "empty", text: `T | where Account == ""`, databaseColumn: true},
		{name: "four ASCII units", text: `T | where Account == "abcd"`, databaseColumn: true},
		{name: "four UTF-16 units", text: `T | where Account == "😀ab"`, databaseColumn: true},
		{name: "contains unsupported", text: `T | where Account contains "ab"`, databaseColumn: true},
		{name: "starts with cs unsupported", text: `T | where Account startswith_cs "ab"`, databaseColumn: true},
		{name: "non database column", text: `T | where Account == "ab"`, databaseColumn: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := shortStringAnalyzerBoundCode(t, tt.text, tt.databaseColumn)
			if diagnostics := NewAvoidUsingShortStringComparisionAnalyzer().Analyze(code, nil, CancellationToken{}); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want empty", diagnostics)
			}
		})
	}
}

func TestAvoidUsingShortStringComparisionAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingShortStringComparisionAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingShortStringComparision ||
		diagnostic.Category != DiagnosticCategoryPerformance ||
		diagnostic.Severity != DiagnosticSeveritySuggestion ||
		diagnostic.Description != "Avoid short string comparisons" ||
		diagnostic.Message != "Avoid using short strings (less than 4 characters) for string comparison operations (see: https://aka.ms/kusto.stringterms)." {
		t.Fatalf("diagnostic metadata = %#v", diagnostic)
	}
}

func shortStringAnalyzerBoundCode(t *testing.T, text string, databaseColumn bool) *KustoCode {
	t.Helper()
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	globals := DefaultGlobalState
	if databaseColumn {
		globals = NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", column)))
	}
	code := ParseWithGlobals(text, globals)
	reference := firstNameReferenceInElement(t, code.Syntax(), "Account")
	reference.SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	return code
}
