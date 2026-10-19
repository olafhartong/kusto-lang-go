package kustolang

// This file ports Binder_TreeBinder and Binder_NodeBinder handling for
// ToScalarExpression and ToTableExpression from upstream cd24dcf.

func (t *treeBinder) visitToScalarExpression(node *ToScalarExpression) {
	t.visitToScalarTable(node.SyntaxNode, node.KindParameter(), node.Expression())
}

func (t *treeBinder) visitToTableExpression(node *ToTableExpression) {
	t.visitToScalarTable(node.SyntaxNode, node.KindParameter(), node.Expression())
}

func (t *treeBinder) visitToScalarTable(node *SyntaxNode, kind *NamedParameter, expression Expression) {
	if node == nil || node.IsBound() || !supportsToScalarTableExpression(node, kind, expression) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldScanAliases := t.scanAliases
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.scanAliases = oldScanAliases
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.schemaCannotBeEmpty = oldSchemaCannotBeEmpty
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		t.nodeBinder.lookupRightRowScope = oldLookupRightRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	// Upstream binds the optional kind in the current scope.
	if kind != nil {
		t.visitElement(kind)
		if value := simpleBinderSyntaxNode(kind.Expression()); value != nil && !value.IsBound() {
			return
		}
	}

	// Only the enclosed expression loses the incoming row scope.
	t.binder.rowScope = nil
	if syntaxElementIsMissing(expression) {
		simpleBinderSyntaxNode(expression).SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
	} else if reference, ok := expression.(*NameReference); ok {
		// IsInTabularContext in upstream derives this from generated syntax
		// completion hints. the compatibility implementation marked both wrapper operands Tabular;
		// current Kusto marks only toscalar's operand NonScalar.
		inTabularContext := node.Kind() == SyntaxKindToScalarExpression
		reference.SetSemanticInfo(t.binder.ResolveNameReferenceInRowScope(nil, reference, inTabularContext))
	} else {
		t.visitElement(expression)
	}
	if child := simpleBinderSyntaxNode(expression); child == nil || !child.IsBound() {
		return
	}
	t.binder.rowScope = oldRowScope
	t.bindNode(node)
	if !node.IsBound() {
		return
	}
	committed = true
}

func supportsToScalarTableExpression(node *SyntaxNode, kind *NamedParameter, expression Expression) bool {
	if node.ChildCount() != 5 || expression == nil || simpleBinderSyntaxNode(expression) == nil {
		return false
	}
	keyword, _ := node.GetChild(0).(*SyntaxToken)
	open, _ := node.GetChild(2).(*SyntaxToken)
	close, _ := node.GetChild(4).(*SyntaxToken)
	if keyword == nil || keyword.Width() == 0 || open == nil || open.Kind() != SyntaxKindOpenParenToken ||
		open.Width() == 0 || close == nil || close.Kind() != SyntaxKindCloseParenToken ||
		(close.Width() == 0 && !syntaxElementIsMissing(close)) {
		return false
	}
	switch node.Kind() {
	case SyntaxKindToScalarExpression:
		if keyword.Kind() != SyntaxKindToScalarKeyword {
			return false
		}
	case SyntaxKindToTableExpression:
		if keyword.Kind() != SyntaxKindToTableKeyword {
			return false
		}
	default:
		return false
	}
	if kind == nil {
		return node.GetChild(1) == nil
	}
	return node.GetChild(1) != nil && kind.SyntaxNode != nil && kind.ChildCount() == 3 &&
		kind.Name() != nil && NameDeclarationSimpleName(kind.Name()) == "kind" &&
		kind.EqualToken() != nil && kind.EqualToken().Kind() == SyntaxKindEqualToken && kind.Expression() != nil
}

func (n *simpleNodeBinder) bindToScalarExpression(node *ToScalarExpression) *SemanticInfo {
	diagnostics := n.toScalarTableKindDiagnostics(node.KindParameter(), QueryOperatorParameters.ToScalarKindParameter)
	if table := exactExpressionTable(node.Expression()); table != nil {
		columns := table.Columns()
		if len(columns) == 0 {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTableHasNoColumns(), node.Expression()))
			return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...).WithIsConstant(true))
		}
		return semanticInfoPtr(NewSemanticInfoForType(columns[0].Type, diagnostics...).WithIsConstant(true))
	}
	if info := boundSemanticInfoForExpression(node.Expression()); info != nil && info.ReferencedSymbol() != nil {
		if result := GetResultType(info.ReferencedSymbol()); result == nil || !IsScalar(result) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(n.tableOrScalarExpectedDiagnostic(), node.Expression()))
			return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...).WithIsConstant(true))
		}
	}
	if typ := ExpressionRawResultType(node.Expression()); typ != nil && !isErrorTypeSymbol(typ) && typ.SymbolTabularity() == TabularityScalar {
		return semanticInfoPtr(NewSemanticInfoForType(typ, diagnostics...).WithIsConstant(true))
	}
	diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(n.tableOrScalarExpectedDiagnostic(), node.Expression()))
	return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...).WithIsConstant(true))
}

func (n *simpleNodeBinder) tableOrScalarExpectedDiagnostic() Diagnostic {
	return DiagnosticTableOrScalarExpected()
}

func (n *simpleNodeBinder) bindToTableExpression(node *ToTableExpression) *SemanticInfo {
	diagnostics := n.toScalarTableKindDiagnostics(node.KindParameter(), QueryOperatorParameters.ToTableKindParameter)
	if syntaxElementIsMissing(node.Expression()) {
		return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...))
	}
	if table := exactExpressionTable(node.Expression()); table != nil {
		return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
	}
	if typ := ExpressionRawResultType(node.Expression()); isErrorTypeSymbol(typ) {
		return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...))
	}
	diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), node.Expression()))
	return semanticInfoPtr(NewSemanticInfoForType(&Error.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) toScalarTableKindDiagnostics(parameter *NamedParameter, declaration *QueryOperatorParameter) []Diagnostic {
	if parameter == nil {
		return nil
	}
	parameter.Name().SetSemanticInfo(NewSemanticInfoForSymbol(declaration, ScalarTypes.Unknown.Type()))
	value := ""
	if info := ExpressionLiteralValueInfo(parameter.Expression()); info != nil {
		value = info.ValueText
	}
	values := declaration.Values()
	if !stringInParameterValues(value, values, declaration.IsCaseSensitive()) {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveValue(values...), parameter.Expression())}
	}
	return nil
}

func exactExpressionTable(expression Expression) *TableSymbol {
	if table := expressionTableResult(expression); table != nil {
		return table
	}
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return exactExpressionTable(expression.Expression())
	}
	return nil
}
