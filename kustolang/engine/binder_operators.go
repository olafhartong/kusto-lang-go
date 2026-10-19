package kustolang

// Ported from Binder_Operators.cs and the scalar-operator visitors in
// Binder_NodeBinder.cs at the revision pinned in third_party/kusto/UPSTREAM.json.

func (n *simpleNodeBinder) bindBinary(node *BinaryExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.operatorInfo(operatorKindForSyntax(node.Kind()), []Expression{node.Left(), node.Right()}, node.Operator(), true)
}

func (n *simpleNodeBinder) bindPrefixUnary(node *PrefixUnaryExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.operatorInfo(operatorKindForSyntax(node.Kind()), []Expression{node.Expression()}, node.Operator(), true)
}

func (n *simpleNodeBinder) bindIn(node *InExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindListOperator(operatorKindForSyntax(node.Kind()), node.Left(), node.Right(), node.Operator())
}

func (n *simpleNodeBinder) bindHasAny(node *HasAnyExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindListOperator(operatorKindForSyntax(node.Kind()), node.Left(), node.Right(), node.Operator())
}

func (n *simpleNodeBinder) bindHasAll(node *HasAllExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindListOperator(operatorKindForSyntax(node.Kind()), node.Left(), node.Right(), node.Operator())
}

func (n *simpleNodeBinder) bindListOperator(kind OperatorKind, left Expression, right *ExpressionList, location SyntaxElement) *SemanticInfo {
	arguments := []Expression{left}
	if right != nil && right.Expressions() != nil {
		for _, separated := range right.Expressions().Elements() {
			if separated != nil {
				arguments = append(arguments, separated.Element())
			}
		}
	}
	return n.operatorInfo(kind, arguments, location, false)
}

func (n *simpleNodeBinder) bindBetween(node *BetweenExpression) *SemanticInfo {
	if node == nil || node.Right() == nil {
		return nil
	}
	return n.operatorInfo(operatorKindForSyntax(node.Kind()), []Expression{
		node.Left(), node.Right().First(), node.Right().Second(),
	}, node.Operator(), false)
}

func (n *simpleNodeBinder) operatorInfo(kind OperatorKind, arguments []Expression, location SyntaxElement, requireAll bool) *SemanticInfo {
	op := n.operatorForKind(kind)
	if op == nil {
		return nil
	}
	types := make([]*TypeSymbol, len(arguments))
	for i, argument := range arguments {
		types[i] = resultTypeOrError(argument)
	}
	return n.operatorInfoForTypes(op, arguments, types, location, requireAll, true)
}

func (n *simpleNodeBinder) operatorForKind(kind OperatorKind) *OperatorSymbol {
	if n != nil && n.binder != nil && n.binder.Globals != nil {
		return n.binder.Globals.GetOperator(kind)
	}
	return operatorForKind(kind)
}

func (n *simpleNodeBinder) operatorInfoForTypes(op *OperatorSymbol, arguments []Expression, argumentTypes []*TypeSymbol, location SyntaxElement, requireAll, allowDynamicRetry bool) *SemanticInfo {
	matches := GetBestMatchingSignatures(op.Signatures(), arguments, symbolsFromTypeSymbols(argumentTypes), requireAll)
	if len(matches) == 1 {
		signature := matches[0]
		parameters := signature.ArgumentParametersForExpressions(arguments, symbolsFromTypeSymbols(argumentTypes), n.binder.Globals)
		diagnostics := n.functionSignatureDiagnostics(signature, location, arguments, argumentTypes, parameters)
		result := operatorSignatureReturnType(signature, n.binder, arguments, argumentTypes)
		if allowDynamicRetry && sameType(result, ScalarTypes.Dynamic.Type()) && HasDynamicPrimitives(argumentTypes) {
			unwrapped := GetUnwrappedDynamicPrimitives(argumentTypes)
			inner := n.operatorInfoForTypes(op, arguments, unwrapped, location, requireAll, false)
			if inner != nil && IsScalar(inner.ResultType) && !isDynamicType(inner.ResultType) &&
				!sameType(inner.ResultType, ScalarTypes.Unknown.Type()) && !isErrorTypeSymbol(inner.ResultType) {
				result = ScalarTypes.GetDynamic(inner.ResultType)
			}
		}
		info := NewSemanticInfoForSignature(signature, result, diagnostics...)
		info.IsConstant = allOperatorArgumentsConstant(arguments)
		return &info
	}

	if len(matches) == 0 && requireAll {
		matches = GetBestMatchingSignatures(op.Signatures(), arguments, symbolsFromTypeSymbols(argumentTypes), false)
	}
	diagnostics := []Diagnostic{}
	if !argumentTypesHaveErrorsOrUnknown(argumentTypes) {
		names := make([]string, len(argumentTypes))
		for i, typ := range argumentTypes {
			names[i] = typ.SymbolName()
		}
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticOperatorNotDefined(syntaxLocationText(location), names...), location))
	}
	result := GetCommonSignatureReturnType(matches, argumentTypes, func(signature *Signature) *TypeSymbol {
		return operatorSignatureReturnType(signature, n.binder, arguments, argumentTypes)
	})
	return semanticInfoPtr(NewSemanticInfoForSymbol(op, result, diagnostics...))
}

func operatorSignatureReturnType(signature *Signature, binder *Binder, arguments []Expression, argumentTypes []*TypeSymbol) *TypeSymbol {
	if signature != nil && signature.ReturnKind == ReturnTypeCustom && signature.CustomReturnType != nil {
		constants := make([]bool, len(arguments))
		for i, argument := range arguments {
			if info := boundSemanticInfoForExpression(argument); info != nil {
				constants[i] = info.IsConstant
			}
		}
		result := signature.CustomReturnType(CustomReturnTypeContext{
			Arguments:          arguments,
			ArgumentTypes:      argumentTypes,
			ArgumentParameters: signature.ArgumentParametersForExpressions(arguments, symbolsFromTypeSymbols(argumentTypes), binder.Globals),
			ArgumentIsConstant: constants,
			RowScope:           binder.rowScope,
			Signature:          signature,
			Globals:            binder.Globals,
		})
		return returnTypeFromSymbol(result)
	}
	return signature.GetReturnTypeForArguments(binder.Globals, arguments, argumentTypes)
}

func operatorForKind(kind OperatorKind) *OperatorSymbol {
	for _, op := range Operators.All {
		if op != nil && op.OperatorKind == kind {
			return op
		}
	}
	return nil
}

func allOperatorArgumentsConstant(arguments []Expression) bool {
	for _, argument := range arguments {
		info := boundSemanticInfoForExpression(argument)
		if info == nil || !info.IsConstant {
			return false
		}
	}
	return true
}

func syntaxLocationText(location SyntaxElement) string {
	if token, ok := location.(*SyntaxToken); ok && token != nil {
		return token.ValueText()
	}
	return ""
}

func operatorKindForSyntax(kind SyntaxKind) OperatorKind {
	switch kind {
	case SyntaxKindUnaryMinusExpression:
		return OperatorKindUnaryMinus
	case SyntaxKindUnaryPlusExpression:
		return OperatorKindUnaryPlus
	case SyntaxKindAddExpression:
		return OperatorKindAdd
	case SyntaxKindSubtractExpression:
		return OperatorKindSubtract
	case SyntaxKindMultiplyExpression:
		return OperatorKindMultiply
	case SyntaxKindDivideExpression:
		return OperatorKindDivide
	case SyntaxKindModuloExpression:
		return OperatorKindModulo
	case SyntaxKindLessThanExpression:
		return OperatorKindLessThan
	case SyntaxKindLessThanOrEqualExpression:
		return OperatorKindLessThanOrEqual
	case SyntaxKindGreaterThanExpression:
		return OperatorKindGreaterThan
	case SyntaxKindGreaterThanOrEqualExpression:
		return OperatorKindGreaterThanOrEqual
	case SyntaxKindEqualExpression:
		return OperatorKindEqual
	case SyntaxKindNotEqualExpression:
		return OperatorKindNotEqual
	case SyntaxKindAndExpression:
		return OperatorKindAnd
	case SyntaxKindOrExpression:
		return OperatorKindOr
	case SyntaxKindInExpression:
		return OperatorKindIn
	case SyntaxKindInCsExpression:
		return OperatorKindInCs
	case SyntaxKindNotInExpression:
		return OperatorKindNotIn
	case SyntaxKindNotInCsExpression:
		return OperatorKindNotInCs
	case SyntaxKindBetweenExpression:
		return OperatorKindBetween
	case SyntaxKindNotBetweenExpression:
		return OperatorKindNotBetween
	case SyntaxKindEqualTildeExpression:
		return OperatorKindEqualTilde
	case SyntaxKindBangTildeExpression:
		return OperatorKindBangTilde
	case SyntaxKindHasExpression:
		return OperatorKindHas
	case SyntaxKindHasCsExpression:
		return OperatorKindHasCs
	case SyntaxKindNotHasExpression:
		return OperatorKindNotHas
	case SyntaxKindNotHasCsExpression:
		return OperatorKindNotHasCs
	case SyntaxKindHasPrefixExpression:
		return OperatorKindHasPrefix
	case SyntaxKindHasPrefixCsExpression:
		return OperatorKindHasPrefixCs
	case SyntaxKindNotHasPrefixExpression:
		return OperatorKindNotHasPrefix
	case SyntaxKindNotHasPrefixCsExpression:
		return OperatorKindNotHasPrefixCs
	case SyntaxKindHasSuffixExpression:
		return OperatorKindHasSuffix
	case SyntaxKindHasSuffixCsExpression:
		return OperatorKindHasSuffixCs
	case SyntaxKindNotHasSuffixExpression:
		return OperatorKindNotHasSuffix
	case SyntaxKindNotHasSuffixCsExpression:
		return OperatorKindNotHasSuffixCs
	case SyntaxKindLikeExpression:
		return OperatorKindLike
	case SyntaxKindLikeCsExpression:
		return OperatorKindLikeCs
	case SyntaxKindNotLikeExpression:
		return OperatorKindNotLike
	case SyntaxKindNotLikeCsExpression:
		return OperatorKindNotLikeCs
	case SyntaxKindContainsExpression:
		return OperatorKindContains
	case SyntaxKindContainsCsExpression:
		return OperatorKindContainsCs
	case SyntaxKindNotContainsExpression:
		return OperatorKindNotContains
	case SyntaxKindNotContainsCsExpression:
		return OperatorKindNotContainsCs
	case SyntaxKindStartsWithExpression:
		return OperatorKindStartsWith
	case SyntaxKindStartsWithCsExpression:
		return OperatorKindStartsWithCs
	case SyntaxKindNotStartsWithExpression:
		return OperatorKindNotStartsWith
	case SyntaxKindNotStartsWithCsExpression:
		return OperatorKindNotStartsWithCs
	case SyntaxKindEndsWithExpression:
		return OperatorKindEndsWith
	case SyntaxKindEndsWithCsExpression:
		return OperatorKindEndsWithCs
	case SyntaxKindNotEndsWithExpression:
		return OperatorKindNotEndsWith
	case SyntaxKindNotEndsWithCsExpression:
		return OperatorKindNotEndsWithCs
	case SyntaxKindMatchesRegexExpression:
		return OperatorKindMatchRegex
	case SyntaxKindSearchExpression:
		return OperatorKindSearch
	case SyntaxKindHasAnyExpression:
		return OperatorKindHasAny
	case SyntaxKindHasAllExpression:
		return OperatorKindHasAll
	default:
		return OperatorKindNone
	}
}
