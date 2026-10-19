package kustolang

// Ported from the PatternSymbol path in
// src/Kusto.Language/Binder/Binder_FunctionCalls.cs.

type PatternCallResult struct {
	Pattern            *PatternSymbol
	MatchingSignatures []*PatternSignature
	Result             Symbol
	Diagnostics        []Diagnostic
	// DiagnosticArgumentIndices parallels Diagnostics. A non-negative value
	// identifies the argument that owns the diagnostic; -1 identifies the
	// function name, matching Binder_FunctionCalls.BindPattern.
	DiagnosticArgumentIndices []int
}

func (r PatternCallResult) ResultType() *TypeSymbol {
	if group, ok := r.Result.(*GroupSymbol); ok && group != nil {
		return &group.TypeSymbol
	}
	return returnTypeFromSymbol(r.Result)
}

func BindPatternCall(pattern *PatternSymbol, arguments []Expression, globals *GlobalState) PatternCallResult {
	return BindPatternCallWithArgumentTypes(pattern, arguments, nil, globals)
}

func BindPatternCallWithArgumentTypes(pattern *PatternSymbol, arguments []Expression, argumentTypes []*TypeSymbol, globals *GlobalState) PatternCallResult {
	if pattern == nil {
		return PatternCallResult{
			Result:                    Error,
			Diagnostics:               []Diagnostic{DiagnosticNoPatternMatchesArguments()},
			DiagnosticArgumentIndices: []int{-1},
		}
	}

	diagnostics := []Diagnostic{}
	diagnosticArgumentIndices := []int{}
	parameters := pattern.Parameters()
	if len(parameters) != len(arguments) {
		diagnostics = append(diagnostics, DiagnosticArgumentCountExpected(len(parameters)))
		diagnosticArgumentIndices = append(diagnosticArgumentIndices, -1)
	}

	for i, argument := range arguments {
		if i >= len(parameters) {
			continue
		}
		typ := patternParameterDeclaredType(parameters[i])
		actualType := parameterTypeAt(argumentTypes, i)
		if isErrorTypeSymbol(typ) || sameType(typ, ScalarTypes.Unknown.Type()) ||
			(actualType != nil && (isErrorTypeSymbol(actualType) || sameType(actualType, ScalarTypes.Unknown.Type()))) {
			continue
		}
		if patternArgumentHasExactType(argument, typ, actualType) {
			if !patternArgumentIsLiteral(argument) {
				diagnostics = append(diagnostics, DiagnosticExpressionMustBeLiteral())
				diagnosticArgumentIndices = append(diagnosticArgumentIndices, i)
			}
		} else {
			diagnostics = append(diagnostics, DiagnosticExpressionMustHaveType(typ.SymbolName()))
			diagnosticArgumentIndices = append(diagnosticArgumentIndices, i)
		}
	}

	matching := GetMatchingPatternSignatures(pattern.Signatures(), arguments)
	if len(matching) == 0 {
		diagnostics = append(diagnostics, DiagnosticNoPatternMatchesArguments())
		diagnosticArgumentIndices = append(diagnosticArgumentIndices, -1)
		return PatternCallResult{
			Pattern:                   pattern,
			MatchingSignatures:        matching,
			Result:                    Error,
			Diagnostics:               diagnostics,
			DiagnosticArgumentIndices: diagnosticArgumentIndices,
		}
	}

	return PatternCallResult{
		Pattern:                   pattern,
		MatchingSignatures:        matching,
		Result:                    GetPatternCallReturnSymbol(matching, globals),
		Diagnostics:               diagnostics,
		DiagnosticArgumentIndices: diagnosticArgumentIndices,
	}
}

func parameterTypeAt(types []*TypeSymbol, index int) *TypeSymbol {
	if index < 0 || index >= len(types) {
		return nil
	}
	return types[index]
}

func GetMatchingPatternSignatures(signatures []*PatternSignature, arguments []Expression) []*PatternSignature {
	matching := []*PatternSignature{}
	for _, signature := range signatures {
		if PatternSignatureMatches(signature, arguments, true) {
			matching = append(matching, signature)
		}
	}
	if len(matching) != 0 {
		return matching
	}
	for _, signature := range signatures {
		if PatternSignatureMatches(signature, arguments, false) {
			matching = append(matching, signature)
		}
	}
	return matching
}

func PatternSignatureMatches(signature *PatternSignature, arguments []Expression, exact bool) bool {
	if signature == nil {
		return false
	}
	if exact && len(signature.ArgumentValues) != len(arguments) {
		return false
	}
	for i, argument := range arguments {
		if i >= len(signature.ArgumentValues) {
			continue
		}
		if signature.ArgumentValues[i] != patternArgumentLiteralValue(argument) {
			return false
		}
	}
	return true
}

func GetPatternCallReturnSymbol(signatures []*PatternSignature, globals *GlobalState) Symbol {
	if len(signatures) == 1 && !signatures[0].HasPathValue() {
		return patternSignatureReturnSymbol(signatures[0], globals)
	}

	paths := []Symbol{}
	types := []Symbol{}
	for _, signature := range signatures {
		if signature == nil {
			continue
		}
		typ := patternSignatureReturnSymbol(signature, globals)
		if signature.HasPathValue() {
			paths = append(paths, NewVariableSymbol(signature.PathValue, typ, false, nil, nil))
			continue
		}
		if !symbolSliceContains(types, typ) {
			types = append(types, typ)
		}
	}

	switch {
	case len(paths) > 0 && len(types) > 0:
		return NewGroupSymbol(append(paths, types...)...)
	case len(paths) > 0:
		return NewGroupSymbol(paths...)
	case len(types) == 1:
		return types[0]
	default:
		return NewGroupSymbol(types...)
	}
}

func patternSignatureReturnSymbol(signature *PatternSignature, globals *GlobalState) Symbol {
	if signature == nil || signature.Signature() == nil {
		return ScalarTypes.Unknown.Type()
	}
	return signature.Signature().GetReturnType(globals)
}

func symbolSliceContains(symbols []Symbol, target Symbol) bool {
	for _, symbol := range symbols {
		if symbol == target {
			return true
		}
	}
	return false
}

func patternParameterDeclaredType(parameter *Parameter) *TypeSymbol {
	if parameter == nil || len(parameter.DeclaredTypes) == 0 || parameter.DeclaredTypes[0] == nil {
		return ScalarTypes.Unknown.Type()
	}
	return parameter.DeclaredTypes[0]
}

func patternArgumentHasExactType(argument Expression, typ *TypeSymbol, boundType *TypeSymbol) bool {
	if boundType != nil {
		return sameType(boundType, typ)
	}
	actual, ok := patternArgumentLiteralType(argument)
	if !ok {
		return false
	}
	return sameType(actual, typ)
}

func patternArgumentLiteralType(argument Expression) (*TypeSymbol, bool) {
	switch value := argument.(type) {
	case *ParenthesizedExpression:
		return patternArgumentLiteralType(value.Expression())
	case *SimpleNamedExpression:
		return patternArgumentLiteralType(value.Expression())
	case *CompoundNamedExpression:
		return patternArgumentLiteralType(value.Expression())
	case *PrefixUnaryExpression:
		return patternArgumentLiteralType(value.Expression())
	case *CompoundStringLiteralExpression:
		return ScalarTypes.String.Type(), true
	case *TypeOfLiteralExpression:
		return ScalarTypes.TypeLit.Type(), true
	case *DynamicExpression:
		return ScalarTypes.Dynamic.Type(), true
	case *LiteralExpression:
		switch value.Kind() {
		case SyntaxKindStringLiteralExpression:
			return ScalarTypes.String.Type(), true
		case SyntaxKindBooleanLiteralExpression:
			return ScalarTypes.Bool.Type(), true
		case SyntaxKindLongLiteralExpression:
			return ScalarTypes.Long.Type(), true
		case SyntaxKindRealLiteralExpression:
			return ScalarTypes.Real.Type(), true
		case SyntaxKindDecimalLiteralExpression:
			return ScalarTypes.Decimal.Type(), true
		case SyntaxKindIntLiteralExpression:
			return ScalarTypes.Int.Type(), true
		case SyntaxKindGuidLiteralExpression:
			return ScalarTypes.Guid.Type(), true
		case SyntaxKindDateTimeLiteralExpression:
			return ScalarTypes.DateTime.Type(), true
		case SyntaxKindTimespanLiteralExpression:
			return ScalarTypes.TimeSpan.Type(), true
		case SyntaxKindNullLiteralExpression:
			return ScalarTypes.Null.Type(), true
		}
	}
	return ScalarTypes.Unknown.Type(), false
}

func patternArgumentIsLiteral(argument Expression) bool {
	switch value := argument.(type) {
	case *ParenthesizedExpression:
		return patternArgumentIsLiteral(value.Expression())
	case *SimpleNamedExpression:
		return patternArgumentIsLiteral(value.Expression())
	case *CompoundNamedExpression:
		return patternArgumentIsLiteral(value.Expression())
	case *PrefixUnaryExpression:
		return patternArgumentIsLiteral(value.Expression())
	case *LiteralExpression, *CompoundStringLiteralExpression, *TypeOfLiteralExpression, *DynamicExpression:
		return true
	default:
		return false
	}
}

func patternArgumentLiteralValue(argument Expression) string {
	switch value := argument.(type) {
	case *ParenthesizedExpression:
		return patternArgumentLiteralValue(value.Expression())
	case *SimpleNamedExpression:
		return patternArgumentLiteralValue(value.Expression())
	case *CompoundNamedExpression:
		return patternArgumentLiteralValue(value.Expression())
	case *PrefixUnaryExpression:
		if value.Operator() == nil {
			return patternArgumentLiteralValue(value.Expression())
		}
		return value.Operator().Text(IncludeTriviaMinimal) + patternArgumentLiteralValue(value.Expression())
	case *LiteralExpression:
		if value.Token() == nil {
			return ""
		}
		return value.Token().ValueText()
	case *CompoundStringLiteralExpression:
		if value.Tokens() == nil {
			return ""
		}
		text := ""
		for i := 0; i < value.Tokens().Count(); i++ {
			token := value.Tokens().At(i)
			if token != nil {
				text += token.ValueText()
			}
		}
		return text
	default:
		return ""
	}
}
