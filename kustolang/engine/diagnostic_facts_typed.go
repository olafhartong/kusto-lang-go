package kustolang

import (
	"fmt"
	"reflect"
)

// DiagnosticMissingClauseNamed adapts GetMissingClause(string). The parameterless
// overload remains DiagnosticMissingClause.
func DiagnosticMissingClauseNamed(clauseName string) Diagnostic {
	return DiagnosticMissingMaterializedViewCombineClause(clauseName)
}

// DiagnosticTokenExpectedKinds adapts the SyntaxKind overloads of
// GetTokenExpected. Go variadics represent both the params and list forms.
func DiagnosticTokenExpectedKinds(kinds ...SyntaxKind) Diagnostic {
	texts := make([]string, len(kinds))
	for i, kind := range kinds {
		texts[i] = SyntaxText(kind)
	}
	return DiagnosticTokenExpected(texts...)
}

// DiagnosticOperatorNotDefinedTypes adapts the TypeSymbol overloads of
// GetOperatorNotDefined. Go variadics represent both the params and list forms.
func DiagnosticOperatorNotDefinedTypes(name string, argumentTypes ...*TypeSymbol) Diagnostic {
	return DiagnosticOperatorNotDefined(name, diagnosticSymbolNames(argumentTypes)...)
}

// DiagnosticTypeExpectedSymbol adapts GetTypeExpected(Symbol).
func DiagnosticTypeExpectedSymbol(typ Symbol) Diagnostic {
	return DiagnosticTypeExpected(typ.SymbolName())
}

// DiagnosticTypeExpectedTypes adapts GetTypeExpected(IReadOnlyList<TypeSymbol>).
func DiagnosticTypeExpectedTypes(types []*TypeSymbol) Diagnostic {
	return DiagnosticTypeExpected(diagnosticSymbolNames(types)...)
}

// DiagnosticExpressionMustHaveValueTyped adapts the generic value overloads of
// GetExpressionMustHaveValue.
func DiagnosticExpressionMustHaveValueTyped[T any](values ...T) Diagnostic {
	return DiagnosticExpressionMustHaveValue(diagnosticValueTexts(values, len(values) == 1)...)
}

// DiagnosticExpressionMustNotHaveValueTyped adapts
// GetExpressionMustNotHaveValue<T>(IReadOnlyList<T>).
func DiagnosticExpressionMustNotHaveValueTyped[T any](values []T) Diagnostic {
	return DiagnosticExpressionMustNotHaveValue(diagnosticValueTexts(values, len(values) == 1)...)
}

// DiagnosticExpressionMustHaveTypeTyped adapts the generic Symbol overloads of
// GetExpressionMustHaveType. Go variadics represent both params and list forms.
func DiagnosticExpressionMustHaveTypeTyped[S Symbol](types ...S) Diagnostic {
	return DiagnosticExpressionMustHaveType(diagnosticSymbolNames(types)...)
}

// DiagnosticTypeNotAllowedSymbol preserves the Symbol parameter of
// GetTypeNotAllowed.
func DiagnosticTypeNotAllowedSymbol(typ Symbol) Diagnostic {
	return DiagnosticTypeNotAllowed(typ.SymbolName())
}

// DiagnosticFunctionNotDefinedWithMatchingParameterSymbols preserves the
// IReadOnlyList<Symbol> parameter of GetFunctionNotDefinedWithMatchingParameters.
func DiagnosticFunctionNotDefinedWithMatchingParameterSymbols[S Symbol](name string, argumentTypes []S) Diagnostic {
	return DiagnosticFunctionNotDefinedWithMatchingParameters(name, diagnosticSymbolNames(argumentTypes)...)
}

// DiagnosticTypeIsNotIntervalTypeSymbols preserves the Symbol parameters of
// GetTypeIsNotIntervalType.
func DiagnosticTypeIsNotIntervalTypeSymbols(intervalType, rangeType Symbol) Diagnostic {
	return DiagnosticTypeIsNotIntervalType(intervalType.SymbolName(), rangeType.SymbolName())
}

func diagnosticSymbolNames[S Symbol](symbols []S) []string {
	names := make([]string, len(symbols))
	for i, symbol := range symbols {
		names[i] = symbol.SymbolName()
	}
	return names
}

func diagnosticValueTexts[T any](values []T, interpolation bool) []string {
	texts := make([]string, len(values))
	for i, value := range values {
		texts[i] = diagnosticValueText(any(value), interpolation)
	}
	return texts
}

func diagnosticValueText(value any, interpolation bool) string {
	if diagnosticValueIsNil(value) {
		if interpolation {
			return ""
		}
		panic("DiagnosticFacts generic value ToString called on nil")
	}
	if stringer, ok := value.(interface{ ToString() string }); ok {
		return stringer.ToString()
	}
	if boolean, ok := value.(bool); ok {
		if boolean {
			return "True"
		}
		return "False"
	}
	return fmt.Sprint(value)
}

func diagnosticValueIsNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
