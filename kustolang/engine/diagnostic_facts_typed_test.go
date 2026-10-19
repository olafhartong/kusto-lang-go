package kustolang

import "testing"

type diagnosticCustomValue struct{}

func (diagnosticCustomValue) ToString() string { return "source text" }

func TestDiagnosticFactsTypedOverloadAdaptations(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic Diagnostic
		code       string
		message    string
	}{
		{
			name:       "named missing clause",
			diagnostic: DiagnosticMissingClauseNamed("partition by"),
			code:       "KS006",
			message:    "Missing partition by clause",
		},
		{
			name:       "one syntax kind",
			diagnostic: DiagnosticTokenExpectedKinds(SyntaxKindOpenParenToken),
			code:       "KS005",
			message:    "Expected: (",
		},
		{
			name:       "syntax kind list",
			diagnostic: DiagnosticTokenExpectedKinds(SyntaxKindOpenParenToken, SyntaxKindWhereKeyword),
			code:       "KS005",
			message:    "Expected one of: (, where",
		},
		{
			name:       "one operator type",
			diagnostic: DiagnosticOperatorNotDefinedTypes("+", ScalarTypes.Long.Type()),
			code:       "KS106",
			message:    "The operator '+' is not defined for the operand type long.",
		},
		{
			name:       "operator type list",
			diagnostic: DiagnosticOperatorNotDefinedTypes("+", ScalarTypes.Long.Type(), ScalarTypes.String.Type()),
			code:       "KS106",
			message:    "The operator '+' is not defined for the operand types long and string.",
		},
		{
			name:       "expected symbol",
			diagnostic: DiagnosticTypeExpectedSymbol(NewColumnSymbol("Value", ScalarTypes.Long.Type())),
			code:       "KS107",
			message:    "A value of type Value expected.",
		},
		{
			name:       "expected type list",
			diagnostic: DiagnosticTypeExpectedTypes([]*TypeSymbol{ScalarTypes.Long.Type(), ScalarTypes.String.Type()}),
			code:       "KS107",
			message:    "A value of type long or string expected.",
		},
		{
			name:       "one generic value",
			diagnostic: DiagnosticExpressionMustHaveValueTyped(42),
			code:       "KS140",
			message:    "The expression must be the value: 42",
		},
		{
			name:       "generic value list",
			diagnostic: DiagnosticExpressionMustHaveValueTyped(1, 2, 3),
			code:       "KS140",
			message:    "The expression must be one of the values: 1, 2 or 3",
		},
		{
			name:       "generic excluded value list",
			diagnostic: DiagnosticExpressionMustNotHaveValueTyped([]bool{true, false}),
			code:       "KS140",
			message:    "The expression must not be one of the values: True or False",
		},
		{
			name:       "one expression type",
			diagnostic: DiagnosticExpressionMustHaveTypeTyped(ScalarTypes.Long),
			code:       "KS141",
			message:    "The expression must have the type long.",
		},
		{
			name:       "expression type list",
			diagnostic: DiagnosticExpressionMustHaveTypeTyped(ScalarTypes.Long, ScalarTypes.String),
			code:       "KS141",
			message:    "The expression must have one of the types: long or string.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.diagnostic.Code != tt.code || tt.diagnostic.Message != tt.message || tt.diagnostic.Description != tt.message {
				t.Fatalf("diagnostic = %#v", tt.diagnostic)
			}
		})
	}
}

func TestDiagnosticFactsOtherTypedSignatures(t *testing.T) {
	column := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	tests := []struct {
		diagnostic Diagnostic
		message    string
	}{
		{
			diagnostic: DiagnosticTypeNotAllowedSymbol(column),
			message:    "The value of type 'Value' is not allowed in this context.",
		},
		{
			diagnostic: DiagnosticFunctionNotDefinedWithMatchingParameterSymbols("f", []*ScalarSymbol{ScalarTypes.Long, ScalarTypes.String}),
			message:    "The function 'f' is not compatible with arguments (long, string)",
		},
		{
			diagnostic: DiagnosticTypeIsNotIntervalTypeSymbols(ScalarTypes.String, ScalarTypes.DateTime),
			message:    "The type 'string' is not an appropriate interval type for 'datetime'",
		},
	}
	for _, tt := range tests {
		if tt.diagnostic.Message != tt.message || tt.diagnostic.Description != tt.message {
			t.Errorf("diagnostic = %#v, want message %q", tt.diagnostic, tt.message)
		}
	}
}

func TestDiagnosticFactsStringAPIsRemainAvailable(t *testing.T) {
	var _ func(string, ...string) Diagnostic = DiagnosticOperatorNotDefined
	var _ func(...string) Diagnostic = DiagnosticTypeExpected
	var _ func(...string) Diagnostic = DiagnosticExpressionMustHaveValue
	var _ func(...string) Diagnostic = DiagnosticTokenExpected
}

func TestDiagnosticFactsGenericValueToStringAndNullParity(t *testing.T) {
	custom := DiagnosticExpressionMustHaveValueTyped(diagnosticCustomValue{})
	if custom.Message != "The expression must be the value: source text" {
		t.Fatalf("custom ToString message = %q", custom.Message)
	}

	singleNull := DiagnosticExpressionMustHaveValueTyped[any](nil)
	if singleNull.Message != "The expression must be the value: " {
		t.Fatalf("single null interpolation = %q", singleNull.Message)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("multi-value null should preserve the source ToString null panic")
		}
	}()
	_ = DiagnosticExpressionMustNotHaveValueTyped([]any{"ok", nil})
}
