package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type publicDiagnosticCustomValue struct{}

func (publicDiagnosticCustomValue) ToString() string { return "public source text" }

func TestPublicDiagnosticFactsTypedAdaptations(t *testing.T) {
	kinds := []kustolang.SyntaxKind{kustolang.SyntaxKindOpenParenToken, kustolang.SyntaxKindCloseParenToken}
	types := []*kustolang.TypeSymbol{kustolang.ScalarTypes.Long.Type(), kustolang.ScalarTypes.String.Type()}

	tests := []struct {
		diagnostic kustolang.Diagnostic
		message    string
	}{
		{
			diagnostic: kustolang.DiagnosticFactsGetMissingClauseNamed("partition by"),
			message:    "Missing partition by clause",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetTokenExpectedKinds(kinds...),
			message:    "Expected one of: (, )",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetOperatorNotDefinedTypes("+", types...),
			message:    "The operator '+' is not defined for the operand types long and string.",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetTypeExpectedSymbol(kustolang.NewColumnSymbol("Value", types[0])),
			message:    "A value of type Value expected.",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetTypeExpectedTypes(types),
			message:    "A value of type long or string expected.",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetExpressionMustHaveValueTyped(1, 2),
			message:    "The expression must be one of the values: 1 or 2",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetExpressionMustNotHaveValueTyped([]int{1, 2}),
			message:    "The expression must not be one of the values: 1 or 2",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetExpressionMustHaveTypeTyped(kustolang.ScalarTypes.Long, kustolang.ScalarTypes.String),
			message:    "The expression must have one of the types: long or string.",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetTypeNotAllowedSymbol(kustolang.ScalarTypes.Dynamic),
			message:    "The value of type 'dynamic' is not allowed in this context.",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetFunctionNotDefinedWithMatchingParameterSymbols("f", []*kustolang.ScalarSymbol{kustolang.ScalarTypes.Long, kustolang.ScalarTypes.String}),
			message:    "The function 'f' is not compatible with arguments (long, string)",
		},
		{
			diagnostic: kustolang.DiagnosticFactsGetTypeIsNotIntervalTypeSymbols(kustolang.ScalarTypes.String, kustolang.ScalarTypes.DateTime),
			message:    "The type 'string' is not an appropriate interval type for 'datetime'",
		},
	}

	for _, tt := range tests {
		if tt.diagnostic.Message != tt.message || tt.diagnostic.Description != tt.message {
			t.Errorf("diagnostic = %#v, want message %q", tt.diagnostic, tt.message)
		}
	}
}

func TestPublicDiagnosticFactsGenericValueFormatting(t *testing.T) {
	custom := kustolang.DiagnosticFactsGetExpressionMustHaveValueTyped(publicDiagnosticCustomValue{})
	if custom.Message != "The expression must be the value: public source text" {
		t.Fatalf("custom ToString message = %q", custom.Message)
	}
	null := kustolang.DiagnosticFactsGetExpressionMustHaveValueTyped[any](nil)
	if null.Message != "The expression must be the value: " {
		t.Fatalf("single null interpolation = %q", null.Message)
	}
}

func TestPublicDiagnosticFactsOptionalSourceForms(t *testing.T) {
	item := kustolang.DiagnosticFactsGetNameDoesNotReferToAnyKnownItem("Missing")
	if item.Message != "The name 'Missing' does not refer to any known column, table, variable or function." {
		t.Fatalf("default item diagnostic = %q", item.Message)
	}
	fuzzy := kustolang.DiagnosticFactsGetFuzzyTableNotDefined()
	if fuzzy.Message != "The fuzzy expression does not refer to any known or currently accessible table." {
		t.Fatalf("unnamed fuzzy table diagnostic = %q", fuzzy.Message)
	}
	named := kustolang.DiagnosticFuzzyFunctionNotDefined("Resolve")
	if named.Message != "The fuzzy name 'Resolve' does not refer to any known or currently accessible function." {
		t.Fatalf("named fuzzy function diagnostic = %q", named.Message)
	}
}
