package kustolang

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func DiagnosticFactsGetMissingClauseNamed(clauseName string) Diagnostic {
	return internal.DiagnosticMissingClauseNamed(clauseName)
}

func DiagnosticFactsGetTokenExpectedKinds(kinds ...SyntaxKind) Diagnostic {
	return internal.DiagnosticTokenExpectedKinds(kinds...)
}

func DiagnosticFactsGetOperatorNotDefinedTypes(name string, argumentTypes ...*TypeSymbol) Diagnostic {
	return internal.DiagnosticOperatorNotDefinedTypes(name, argumentTypes...)
}

func DiagnosticFactsGetTypeExpectedSymbol(typ Symbol) Diagnostic {
	return internal.DiagnosticTypeExpectedSymbol(typ)
}

func DiagnosticFactsGetTypeExpectedTypes(types []*TypeSymbol) Diagnostic {
	return internal.DiagnosticTypeExpectedTypes(types)
}

func DiagnosticFactsGetExpressionMustHaveValueTyped[T any](values ...T) Diagnostic {
	return internal.DiagnosticExpressionMustHaveValueTyped(values...)
}

func DiagnosticFactsGetExpressionMustNotHaveValueTyped[T any](values []T) Diagnostic {
	return internal.DiagnosticExpressionMustNotHaveValueTyped(values)
}

func DiagnosticFactsGetExpressionMustHaveTypeTyped[S Symbol](types ...S) Diagnostic {
	return internal.DiagnosticExpressionMustHaveTypeTyped(types...)
}

func DiagnosticFactsGetTypeNotAllowedSymbol(typ Symbol) Diagnostic {
	return internal.DiagnosticTypeNotAllowedSymbol(typ)
}

func DiagnosticFactsGetFunctionNotDefinedWithMatchingParameterSymbols[S Symbol](name string, argumentTypes []S) Diagnostic {
	return internal.DiagnosticFunctionNotDefinedWithMatchingParameterSymbols(name, argumentTypes)
}

func DiagnosticFactsGetTypeIsNotIntervalTypeSymbols(intervalType, rangeType Symbol) Diagnostic {
	return internal.DiagnosticTypeIsNotIntervalTypeSymbols(intervalType, rangeType)
}
