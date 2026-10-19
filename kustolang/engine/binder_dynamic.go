package kustolang

// Ported from the dynamic/JSON methods in Binder_TreeBinder.cs and
// Binder_NodeBinder.cs at upstream cd24dcf.

func (t *treeBinder) visitDynamicExpression(node *DynamicExpression) {
	if t == nil || t.binder == nil || !supportsDynamicExpression(node) || node.IsBound() {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldDepth := t.binder.dynamicDepth
	defer func() {
		t.binder.dynamicDepth = oldDepth
		if !committed {
			tx.rollback()
		}
	}()

	t.binder.dynamicDepth++
	t.visitElement(node.Expression())
	if !expressionIsBound(node.Expression()) {
		return
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func supportsDynamicExpression(node *DynamicExpression) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 4 || node.Dynamic() == nil ||
		node.OpenParen() == nil || node.Expression() == nil || node.CloseParen() == nil {
		return false
	}
	return supportsDynamicJSONValue(node.Expression())
}

func supportsDynamicJSONValue(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		return expression != nil && expression.SyntaxNode != nil && expression.ChildCount() == 1 && expression.Token() != nil
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *PrefixUnaryExpression:
		if expression == nil || expression.Kind() != SyntaxKindUnaryMinusExpression || expression.Operator() == nil || expression.Operator().Width() == 0 {
			return false
		}
		literal, ok := expression.Expression().(*LiteralExpression)
		return ok && literal != nil && (literal.Kind() == SyntaxKindIntLiteralExpression ||
			literal.Kind() == SyntaxKindLongLiteralExpression || literal.Kind() == SyntaxKindRealLiteralExpression)
	case *DynamicExpression:
		return supportsDynamicExpression(expression)
	case *NameReference:
		if syntaxElementIsMissing(expression) {
			return true
		}
		name, ok := expression.Name().(*BracedName)
		return ok && name.OpenBrace() != nil && name.OpenBrace().Width() != 0 && name.Name() != nil &&
			name.Name().Width() != 0 && name.CloseBrace() != nil && name.CloseBrace().Width() != 0
	case *JsonArrayExpression:
		if expression == nil || expression.SyntaxNode == nil || expression.ChildCount() != 3 ||
			expression.OpenBracket() == nil || expression.Values() == nil || expression.CloseBracket() == nil {
			return false
		}
		for _, separated := range expression.Values().Elements() {
			if separated == nil || separated.Element() == nil || !supportsDynamicJSONValue(separated.Element()) {
				return false
			}
		}
		return true
	case *JsonObjectExpression:
		if expression == nil || expression.SyntaxNode == nil || expression.ChildCount() != 3 ||
			expression.OpenBrace() == nil || expression.Pairs() == nil || expression.CloseBrace() == nil {
			return false
		}
		for _, separated := range expression.Pairs().Elements() {
			if separated == nil || !supportsJsonPair(separated.Element()) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func supportsCompoundStringLiteral(expression *CompoundStringLiteralExpression) bool {
	if expression == nil || expression.SyntaxNode == nil || expression.Tokens() == nil || expression.Tokens().Count() < 2 {
		return false
	}
	for _, element := range expression.Tokens().Elements() {
		if element == nil {
			return false
		}
	}
	return true
}

func supportsJsonPair(pair *JsonPair) bool {
	return pair != nil && pair.SyntaxNode != nil && pair.ChildCount() == 3 && pair.Name() != nil &&
		pair.Colon() != nil && pair.Value() != nil && supportsDynamicJSONValue(pair.Value())
}

func expressionIsBound(expression Expression) bool {
	return boundSemanticInfoForExpression(expression) != nil
}

func (n *simpleNodeBinder) bindDynamicExpression(node *DynamicExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	info := boundSemanticInfoForExpression(node.Expression())
	if info == nil || info.ResultType == nil {
		return nil
	}
	result := NewSemanticInfoForType(ScalarTypes.GetDynamic(info.ResultType)).WithIsConstant(true)
	return &result
}

func (n *simpleNodeBinder) bindJsonArrayExpression(node *JsonArrayExpression) *SemanticInfo {
	if node == nil || node.Values() == nil {
		return nil
	}
	types := make([]*TypeSymbol, 0, node.Values().Count())
	for _, separated := range node.Values().Elements() {
		if separated == nil || separated.Element() == nil {
			return nil
		}
		info := boundSemanticInfoForExpression(separated.Element())
		if info == nil || info.ResultType == nil {
			return nil
		}
		types = append(types, info.ResultType)
	}
	result := NewSemanticInfoForType(ScalarTypes.GetDynamicArray(GetCommonResultType(types, ConversionNone)).Type())
	return &result
}

func (n *simpleNodeBinder) bindJsonObjectExpression(node *JsonObjectExpression) *SemanticInfo {
	if node == nil || node.Pairs() == nil {
		return nil
	}
	columns := make([]*ColumnSymbol, 0, node.Pairs().Count())
	for _, separated := range node.Pairs().Elements() {
		if separated == nil || separated.Element() == nil {
			return nil
		}
		pair := separated.Element()
		if pair.Name() == nil || pair.Value() == nil {
			return nil
		}
		typ := ScalarTypes.Unknown.Type()
		if info := boundSemanticInfoForExpression(pair.Value()); info != nil && info.ResultType != nil {
			typ = info.ResultType
		}
		columns = append(columns, NewColumnSymbol(pair.Name().ValueText(), typ))
	}
	result := NewSemanticInfoForType(ScalarTypes.GetDynamicBag(columns...).Type())
	return &result
}
