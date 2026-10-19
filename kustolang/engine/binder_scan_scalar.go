package kustolang

// supportsScanScalar accepts the parser-produced scalar forms visited by the
// upstream scan binder while keeping malformed injected syntax transactional.
func supportsScanScalar(expression Expression) bool {
	if expression == nil {
		return false
	}
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	case *LiteralExpression:
		return expression.Token() != nil && expression.Token().Width() != 0
	case *TypeOfLiteralExpression:
		return expression.Width() != 0 && !syntaxElementHasMissingChildren(expression)
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *DynamicExpression:
		return supportsDynamicExpression(expression)
	case *ToScalarExpression:
		return supportsToScalarTableExpression(expression.SyntaxNode, expression.KindParameter(), expression.Expression())
	case *OrderedExpression:
		return expression.Expression() != nil && supportsScanScalar(expression.Expression())
	case *ParenthesizedExpression:
		return scanTokenIsPresent(expression.OpenParen(), SyntaxKindOpenParenToken) &&
			scanTokenIsPresentOrParserMissing(expression.CloseParen(), SyntaxKindCloseParenToken) &&
			supportsScanScalar(expression.Expression())
	case *SimpleNamedExpression:
		return expression.Name() != nil && serializeNameDeclarationIsSupported(expression.Name()) &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			supportsScanScalar(expression.Expression())
	case *BracketedExpression:
		return completeBracketedExpression(expression) && supportsScanScalar(expression.Expression())
	case *PathExpression:
		return completePathExpression(expression) && supportsScanScalar(expression.Expression()) &&
			supportsScanPathSelector(expression.Selector())
	case *ElementExpression:
		selector, ok := expression.Selector().(*BracketedExpression)
		return ok && completeBracketedExpression(selector) &&
			supportsScanScalar(expression.Expression()) && supportsScanScalar(selector.Expression())
	case *BinaryExpression, *PrefixUnaryExpression:
		node := simpleBinderSyntaxNode(expression)
		if node == nil || operatorKindForSyntax(node.Kind()) == OperatorKindNone {
			return false
		}
		var operator *SyntaxToken
		switch scalar := expression.(type) {
		case *BinaryExpression:
			operator = scalar.Operator()
		case *PrefixUnaryExpression:
			operator = scalar.Operator()
		}
		if operator == nil || operator.Width() == 0 {
			return false
		}
		for _, argument := range scalarOperatorArguments(expression) {
			if !supportsScanScalar(argument) {
				return false
			}
		}
		return true
	case *FunctionCallExpression:
		return supportsScanFunctionCall(expression)
	default:
		return false
	}
}

func supportsScanPathSelector(selector Expression) bool {
	switch selector := selector.(type) {
	case *NameReference:
		return isCompleteProjectByNamesReference(selector)
	case *BracketedExpression:
		return completeBracketedExpression(selector) && supportsScanScalar(selector.Expression())
	case *FunctionCallExpression:
		return supportsScanFunctionCall(selector)
	default:
		return false
	}
}

func supportsScanFunctionCall(call *FunctionCallExpression) bool {
	if call == nil || call.Name() == nil || !isCompleteProjectByNamesReference(call.Name()) || call.ArgumentList() == nil ||
		!scanTokenIsPresent(call.ArgumentList().OpenParen(), SyntaxKindOpenParenToken) ||
		call.ArgumentList().Expressions() == nil || call.ArgumentList().CloseParen() == nil ||
		!scanTokenIsPresentOrParserMissing(call.ArgumentList().CloseParen(), SyntaxKindCloseParenToken) {
		return false
	}
	for _, separated := range call.ArgumentList().Expressions().Elements() {
		if separated == nil || separated.Element() == nil || !supportsScanScalar(separated.Element()) {
			return false
		}
	}
	return true
}

func scanTokenIsPresent(token *SyntaxToken, kind SyntaxKind) bool {
	return token != nil && token.Kind() == kind && token.Width() != 0
}

func scanTokenIsPresentOrParserMissing(token *SyntaxToken, kind SyntaxKind) bool {
	return token != nil && token.Kind() == kind && (token.Width() != 0 || syntaxElementIsMissing(token))
}
