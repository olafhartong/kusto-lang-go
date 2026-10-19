package kustolang

// Ported from KustoCompleter.AnyInfixMatches, HasMatchingInfixOperator,
// ReturnTypeMatchesContext and ParameterMatches in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (c *kustoCompleter) anyInfixMatches(left Expression, operatorText string, position int) bool {
	kind, ok := SyntaxKindFromText(operatorText)
	if !ok {
		return false
	}
	operatorKind := SyntaxOperatorKind(kind)
	if operatorKind == OperatorKindNone {
		return false
	}

	for current := left; current != nil; current = parentExpression(current) {
		if position < current.End() {
			break
		}
		resultType := c.completionExpressionResultType(current, position)
		if resultType == nil {
			continue
		}
		hint := c.infixReturnTypeHint(current)
		if completionHasMatchingInfixOperator(c.code.Globals, operatorKind, resultType, hint) {
			return true
		}
	}
	return false
}

func parentExpression(expression Expression) Expression {
	if expression == nil || expression.Parent() == nil {
		return nil
	}
	parent, _ := WrapGeneratedSyntaxElement(expression.Parent()).(Expression)
	return parent
}

func (c *kustoCompleter) infixReturnTypeHint(expression Expression) CompletionHint {
	if expression != nil && expression.Parent() != nil {
		return c.getChildHint(expression.Parent(), expression.IndexInParent(), CompletionHintQuery)
	}
	return CompletionHintQuery
}

func completionParameterHint(signature *Signature, parameter *Parameter, visiting map[*Parameter]struct{}) CompletionHint {
	if parameter == nil {
		return CompletionHintNone
	}
	if _, exists := visiting[parameter]; exists {
		return CompletionHintNone
	}
	visiting[parameter] = struct{}{}
	defer delete(visiting, parameter)

	switch parameter.ArgumentKind {
	case ArgumentKindColumn, ArgumentKindColumnParameter0, ArgumentKindColumnParameter0Common:
		return CompletionHintColumn
	case ArgumentKindStarOnly:
		return CompletionHintNone
	}

	switch parameter.TypeKind {
	case ParameterTypeDeclared:
		if len(parameter.DeclaredTypes) == 0 || !IsScalar(parameter.DeclaredTypes[0]) {
			return CompletionHintNonScalar
		}
		if len(parameter.DeclaredTypes) == 1 && sameType(parameter.DeclaredTypes[0], ScalarTypes.Bool.Type()) {
			return CompletionHintBoolean
		}
		allNumeric := true
		for _, typ := range parameter.DeclaredTypes {
			if !IsNumeric(typ) {
				allNumeric = false
				break
			}
		}
		if allNumeric {
			return CompletionHintNumber
		}
		return CompletionHintScalar
	case ParameterTypeParameter0, ParameterTypeParameter1, ParameterTypeParameter2:
		index := 0
		switch parameter.TypeKind {
		case ParameterTypeParameter1:
			index = 1
		case ParameterTypeParameter2:
			index = 2
		}
		if signature != nil && index >= 0 && index < len(signature.Parameters) {
			return completionParameterHint(signature, signature.Parameters[index], visiting)
		}
		return CompletionHintNone
	case ParameterTypeTabular:
		return CompletionHintNonScalar
	case ParameterTypeDatabase:
		return CompletionHintDatabase
	case ParameterTypeCluster:
		return CompletionHintCluster
	case ParameterTypeNumber, ParameterTypeNumberOrBool:
		return CompletionHintNumber
	default:
		return CompletionHintScalar
	}
}

func completionHasMatchingInfixOperator(globals *GlobalState, kind OperatorKind, argumentType *TypeSymbol, returnTypeHint CompletionHint) bool {
	if globals == nil {
		globals = DefaultGlobalState
	}
	operator := globals.GetOperator(kind)
	if operator == nil {
		return false
	}
	for _, signature := range operator.Signatures() {
		if signature == nil || signature.IsHidden || len(signature.Parameters) == 0 {
			continue
		}
		if completionParameterMatches(signature.Parameters[0], argumentType) &&
			completionReturnTypeMatchesContext(signature, argumentType, returnTypeHint) {
			return true
		}
	}
	return false
}

func completionReturnTypeMatchesContext(signature *Signature, argumentType *TypeSymbol, hint CompletionHint) bool {
	switch {
	case hint&CompletionHintBoolean != 0:
		return completionSignatureReturnsBoolean(signature, argumentType)
	case hint&(CompletionHintScalar|CompletionHintAggregate) != 0:
		return completionSignatureReturnsNonBooleanScalar(signature, argumentType)
	case hint&CompletionHintNumber != 0:
		return completionSignatureReturnsNumeric(signature, argumentType)
	case hint&CompletionHintColumn != 0:
		return false
	default:
		return true
	}
}

func completionSignatureReturnsBoolean(signature *Signature, argumentType *TypeSymbol) bool {
	switch signature.ReturnKind {
	case ReturnTypeCommon, ReturnTypeCommonNonDynamic, ReturnTypeWidest, ReturnTypeParameter0:
		return scalarFromType(argumentType) != nil && sameType(argumentType, ScalarTypes.Bool.Type())
	case ReturnTypeDeclared:
		return sameType(signature.DeclaredReturnType, ScalarTypes.Bool.Type())
	default:
		return false
	}
}

func completionSignatureReturnsNonBooleanScalar(signature *Signature, argumentType *TypeSymbol) bool {
	switch signature.ReturnKind {
	case ReturnTypeCommon, ReturnTypeCommonNonDynamic, ReturnTypeWidest, ReturnTypeParameter0:
		return scalarFromType(argumentType) != nil && !sameType(argumentType, ScalarTypes.Bool.Type())
	case ReturnTypeDeclared:
		return signature.DeclaredReturnType != nil && IsScalar(signature.DeclaredReturnType) &&
			!sameType(signature.DeclaredReturnType, ScalarTypes.Bool.Type())
	default:
		return false
	}
}

func completionSignatureReturnsNumeric(signature *Signature, argumentType *TypeSymbol) bool {
	switch signature.ReturnKind {
	case ReturnTypeCommon, ReturnTypeCommonNonDynamic, ReturnTypeWidest, ReturnTypeParameter0:
		return scalarFromType(argumentType) != nil && IsNumeric(argumentType)
	case ReturnTypeDeclared:
		return signature.DeclaredReturnType != nil && IsNumeric(signature.DeclaredReturnType)
	default:
		return false
	}
}

func completionParameterMatches(parameter *Parameter, typ *TypeSymbol) bool {
	if parameter == nil || typ == nil {
		return false
	}
	switch parameter.TypeKind {
	case ParameterTypeAny:
		return true
	case ParameterTypeDeclared:
		for _, declared := range parameter.DeclaredTypes {
			if SymbolAssignableTo(typ, declared) {
				return true
			}
		}
		return len(parameter.DeclaredTypes) > 0 && IsPromotableTo(typ, parameter.DeclaredTypes[0])
	case ParameterTypeScalar:
		return IsScalar(typ)
	case ParameterTypeInteger:
		return IsInteger(typ)
	case ParameterTypeRealOrDecimal:
		return IsRealOrDecimal(typ)
	case ParameterTypeStringOrDynamic:
		return IsStringOrDynamic(typ)
	case ParameterTypeStringOrArray:
		return IsStringOrArray(typ)
	case ParameterTypeIntegerOrArray:
		return IsIntegerOrArray(typ)
	case ParameterTypeNumber:
		return IsNumeric(typ)
	case ParameterTypeNumberOrBool:
		return IsNumericOrBool(typ)
	case ParameterTypeSummable:
		return IsSummable(typ)
	case ParameterTypeOrderable:
		return IsOrderable(typ)
	case ParameterTypeTabular:
		return IsTabular(typ)
	case ParameterTypeDatabase:
		return typ.SymbolKind() == SymbolKindDatabase
	case ParameterTypeCluster:
		return typ.SymbolKind() == SymbolKindCluster
	case ParameterTypeNotBool:
		return IsAnyScalarExceptBool(typ)
	case ParameterTypeNotRealOrBool:
		return IsAnyScalarExceptRealOrBool(typ)
	case ParameterTypeNotDynamic:
		return IsAnyScalarExceptDynamic(typ)
	case ParameterTypeParameter0, ParameterTypeParameter1, ParameterTypeParameter2,
		ParameterTypeCommonScalar, ParameterTypeCommonNumber, ParameterTypeCommonSummable,
		ParameterTypeCommonOrderable, ParameterTypeCommonScalarOrDynamic:
		return scalarFromType(typ) != nil
	default:
		return false
	}
}
