package kustolang

// Ported from the function-call identity portion of:
//   src/Kusto.Language/Binder/Binder_NodeBinder.cs
//   src/Kusto.Language/Binder/Binder_FunctionCalls.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.
//
// This deliberately attaches only catalog identities and unambiguous
// argument-count signatures. It does not claim complete tree binding.

func AttachKnownAnalyzerSemantics(code *KustoCode) {
	if code == nil || code.Syntax() == nil || code.Globals == nil {
		return
	}
	var expressions []Expression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if expression, ok := WrapGeneratedSyntaxElement(element).(Expression); ok {
			expressions = append(expressions, expression)
		}
	})
	for i := len(expressions) - 1; i >= 0; i-- {
		attachKnownExpressionType(expressions[i])
	}
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, _ int, ancestors []SyntaxElement) {
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call.Name() == nil {
			return
		}
		name := NameReferenceSimpleName(call.Name())
		if name == "" {
			return
		}
		function := knownFunctionForAnalyzerCall(code.Globals, name, ancestors)
		if function == nil {
			return
		}
		info := NewSemanticInfoForSymbol(function, &function.TypeSymbol)
		arguments, argumentTypes := knownFunctionArguments(call)
		if signature := knownFunctionSignature(function, arguments, argumentTypes); signature != nil {
			resultType := signature.GetReturnTypeForArguments(code.Globals, arguments, argumentTypes)
			info = NewSemanticInfoForSignature(signature, resultType)
		}
		call.SetSemanticInfo(info)
		call.Name().SetSemanticInfo(info)
	})
}

func attachKnownExpressionType(expression Expression) {
	if expression == nil || ExpressionRawResultType(expression) != nil {
		return
	}
	var result *TypeSymbol
	switch node := expression.(type) {
	case *LiteralExpression:
		switch node.Kind() {
		case SyntaxKindStringLiteralExpression:
			result = ScalarTypes.String.Type()
		case SyntaxKindLongLiteralExpression:
			result = ScalarTypes.Long.Type()
		case SyntaxKindIntLiteralExpression:
			result = ScalarTypes.Int.Type()
		case SyntaxKindRealLiteralExpression:
			result = ScalarTypes.Real.Type()
		case SyntaxKindDecimalLiteralExpression:
			result = ScalarTypes.Decimal.Type()
		case SyntaxKindBooleanLiteralExpression:
			result = ScalarTypes.Bool.Type()
		case SyntaxKindDateTimeLiteralExpression:
			result = ScalarTypes.DateTime.Type()
		case SyntaxKindTimespanLiteralExpression:
			result = ScalarTypes.TimeSpan.Type()
		case SyntaxKindGuidLiteralExpression:
			result = ScalarTypes.Guid.Type()
		}
	case *CompoundStringLiteralExpression:
		result = ScalarTypes.String.Type()
	case *DynamicExpression:
		switch node.Expression().(type) {
		case *JsonArrayExpression:
			result = ScalarTypes.DynamicArray.Type()
		case *JsonObjectExpression:
			result = ScalarTypes.DynamicBag.Type()
		default:
			result = ScalarTypes.Dynamic.Type()
		}
	case *ParenthesizedExpression:
		result = ExpressionResultType(node.Expression())
	case *BracketedExpression:
		result = ExpressionResultType(node.Expression())
	case *SimpleNamedExpression:
		result = ExpressionResultType(node.Expression())
	}
	if result != nil {
		if accessor, ok := expression.(interface{ SetSemanticInfo(SemanticInfo) }); ok {
			accessor.SetSemanticInfo(NewSemanticInfoForType(result).WithIsConstant(ExpressionIsLiteral(expression)))
		}
	}
}

func knownFunctionArguments(call *FunctionCallExpression) ([]Expression, []*TypeSymbol) {
	count := functionCallArgumentCount(call)
	arguments := make([]Expression, count)
	types := make([]*TypeSymbol, count)
	for i := 0; i < count; i++ {
		arguments[i] = functionCallArgumentAt(call, i)
		types[i] = ExpressionResultType(arguments[i])
		if types[i] == nil {
			types[i] = ScalarTypes.Unknown.Type()
		}
	}
	return arguments, types
}

func knownFunctionSignature(function *FunctionSymbol, arguments []Expression, argumentTypes []*TypeSymbol) *Signature {
	if function == nil {
		return nil
	}
	matches := GetBestMatchingSignatures(function.Signatures(), arguments, symbolsFromTypeSymbols(argumentTypes), false)
	if len(matches) == 1 {
		return matches[0]
	}
	return uniqueArgumentCountSignature(function, len(arguments))
}

func knownFunctionForAnalyzerCall(globals *GlobalState, name string, ancestors []SyntaxElement) *FunctionSymbol {
	if globals == nil || name == "" {
		return nil
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch WrapGeneratedSyntaxElement(ancestors[i]).(type) {
		case *SummarizeOperator, *MakeSeriesOperator:
			if function := globals.GetAggregate(name); function != nil {
				return function
			}
		case *EvaluateOperator:
			if function := globals.GetPlugIn(name); function != nil {
				return function
			}
		}
	}
	if function := globals.GetFunction(name); function != nil {
		return function
	}
	if function := globals.GetAggregate(name); function != nil {
		return function
	}
	return globals.GetPlugIn(name)
}

func uniqueArgumentCountSignature(function *FunctionSymbol, count int) *Signature {
	if function == nil {
		return nil
	}
	var match *Signature
	for _, signature := range function.Signatures() {
		if signature == nil || !signature.IsValidArgumentCount(count) {
			continue
		}
		if match != nil {
			return nil
		}
		match = signature
	}
	return match
}
