package kustolang

// stageMvExpansionScalar binds a parser-produced scalar used by mv-expand or
// mv-apply without allowing an incomplete expression to leave partial state.
// It is intentionally independent of either operator so mv-expand can adopt it
// when its syntax gate is widened.
func (t *treeBinder) stageMvExpansionScalar(expression Expression) bool {
	if t == nil || t.binder == nil || !supportsMvExpansionScalar(expression) {
		return false
	}

	tx := newUnionBindingTransaction(expression)
	committed := false
	oldRowScope := t.binder.rowScope
	oldPathScope := t.binder.pathScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldDynamicDepth := t.binder.dynamicDepth
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.pathScope = oldPathScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.binder.dynamicDepth = oldDynamicDepth
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	t.visitElement(expression)
	committed = expressionIsBound(expression)
	return committed
}

func supportsMvExpansionScalar(expression Expression) bool {
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
	case *ParenthesizedExpression:
		return mvExpansionTokenIsPresent(expression.OpenParen(), SyntaxKindOpenParenToken) &&
			mvExpansionTokenIsPresentOrParserMissing(expression.CloseParen(), SyntaxKindCloseParenToken) &&
			supportsMvExpansionScalar(expression.Expression())
	case *SimpleNamedExpression:
		return expression.Name() != nil && serializeNameDeclarationIsSupported(expression.Name()) &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			supportsMvExpansionScalar(expression.Expression())
	case *BracketedExpression:
		return completeBracketedExpression(expression) && supportsMvExpansionScalar(expression.Expression())
	case *PathExpression:
		return completePathExpression(expression) && supportsMvExpansionScalar(expression.Expression()) &&
			supportsMvExpansionPathSelector(expression.Selector())
	case *ElementExpression:
		selector, ok := expression.Selector().(*BracketedExpression)
		return ok && completeBracketedExpression(selector) &&
			supportsMvExpansionScalar(expression.Expression()) && supportsMvExpansionScalar(selector.Expression())
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
			if !supportsMvExpansionScalar(argument) {
				return false
			}
		}
		return true
	case *InExpression, *HasAnyExpression, *HasAllExpression, *BetweenExpression:
		node := simpleBinderSyntaxNode(expression)
		return node != nil && supportsScalarOperator(node)
	case *FunctionCallExpression:
		return supportsMvExpansionFunctionCall(expression)
	default:
		return false
	}
}

func supportsMvExpansionPathSelector(selector Expression) bool {
	switch selector := selector.(type) {
	case *NameReference:
		return isCompleteProjectByNamesReference(selector)
	case *BracketedExpression:
		return completeBracketedExpression(selector) && supportsMvExpansionScalar(selector.Expression())
	case *FunctionCallExpression:
		return supportsMvExpansionFunctionCall(selector)
	default:
		return false
	}
}

func supportsMvExpansionFunctionCall(call *FunctionCallExpression) bool {
	if call == nil || call.Name() == nil || !isCompleteProjectByNamesReference(call.Name()) || call.ArgumentList() == nil ||
		!mvExpansionTokenIsPresent(call.ArgumentList().OpenParen(), SyntaxKindOpenParenToken) ||
		call.ArgumentList().Expressions() == nil || call.ArgumentList().CloseParen() == nil ||
		!mvExpansionTokenIsPresentOrParserMissing(call.ArgumentList().CloseParen(), SyntaxKindCloseParenToken) {
		return false
	}
	for _, separated := range call.ArgumentList().Expressions().Elements() {
		if separated == nil || separated.Element() == nil || !supportsMvExpansionScalar(separated.Element()) {
			return false
		}
	}
	return true
}

func mvExpansionTokenIsPresent(token *SyntaxToken, kind SyntaxKind) bool {
	return token != nil && token.Kind() == kind && token.Width() != 0
}

func mvExpansionTokenIsPresentOrParserMissing(token *SyntaxToken, kind SyntaxKind) bool {
	return token != nil && token.Kind() == kind && (token.Width() != 0 || syntaxElementIsMissing(token))
}
