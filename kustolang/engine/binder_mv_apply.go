package kustolang

// Ported from Binder_TreeBinder.cs and Binder_NodeBinder.cs mv-apply visitors
// at the revision pinned in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitMvApply(node *MvApplyOperator) {
	if node == nil || node.Expressions() == nil || node.Subquery() == nil || node.IsBound() {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldInPipe := t.inPipe
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.inPipe = oldInPipe
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !t.bindMvApplyExpression(separated.Element()) {
			return
		}
	}
	if clause := node.RowLimitClause(); clause != nil {
		t.visitElement(clause.RowLimit())
		if bound := simpleBinderSyntaxNode(clause.RowLimit()); bound == nil || !bound.IsBound() {
			return
		}
	}
	if clause := node.ContextIdClause(); clause != nil {
		t.visitElement(clause.Id())
		if bound := simpleBinderSyntaxNode(clause.Id()); bound == nil || !bound.IsBound() {
			return
		}
	}

	// First bind computes the expanded schema that flows into the subquery.
	t.nodeBinder.inPipe = t.inPipe
	intermediate := t.nodeBinder.bind(node.SyntaxNode)
	if intermediate == nil {
		return
	}
	intermediateTable, _ := intermediate.ReferencedSymbol().(*TableSymbol)
	if intermediateTable == nil {
		return
	}

	t.binder.rowScope = intermediateTable
	t.inPipe = true
	t.visitElement(node.Subquery())
	if !node.Subquery().IsBound() {
		return
	}

	builder := NewProjectionBuilder()
	applied := map[*ColumnSymbol]bool{}
	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element().Expression()
		// An aliased expansion (Item=Items) adds Item while retaining Items.
		// Only an unaliased expansion replaces the original source column.
		if mvExpandNamedExpression(expression) != nil {
			continue
		}
		if column, ok := directMvExpandColumn(mvExpandUnderlyingExpression(expression)); ok {
			applied[column] = true
		}
	}
	if oldRowScope != nil {
		for _, column := range oldRowScope.Columns() {
			if !applied[column] {
				builder.Add(column, ProjectionAddOptions{DoNotRepeat: true})
			}
		}
	}
	if subqueryTable := expressionTableResult(node.Subquery()); subqueryTable != nil {
		for _, column := range subqueryTable.Columns() {
			builder.Add(column, ProjectionAddOptions{Replace: true})
		}
	}
	result := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(oldRowScope)
	node.SetSemanticInfo(NewSemanticInfoForSymbol(result, &result.TypeSymbol, intermediate.Diagnostics...))
	committed = true
}

func (t *treeBinder) bindMvApplyExpression(node *MvApplyExpression) bool {
	if node == nil || !t.stageMvExpansionScalar(node.Expression()) {
		return false
	}
	if clause := node.ToTypeOf(); clause != nil {
		typeExpression := mvExpandDeclaredTypeExpression(clause)
		if typeExpression != nil {
			symbol := declaredSymbolFromPrimitiveTypeExpression(typeExpression)
			typeExpression.SetSemanticInfo(NewSemanticInfoForSymbol(symbol, ScalarTypes.TypeLit.Type()))
		}
		clause.TypeOf().SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.TypeLit.Type()))
	}
	node.SetSemanticInfo(NewSemanticInfoForType(mvApplyExpressionResultType(node)))
	return true
}

func (n *simpleNodeBinder) bindMvApplyExpression(node *MvApplyExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return semanticInfoPtr(NewSemanticInfoForType(mvApplyExpressionResultType(node)))
}

func (n *simpleNodeBinder) bindMvApplySubqueryExpression(node *MvApplySubqueryExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	diagnostics := n.postPipeQueryOperatorDiagnostics(node.Expression(), true)
	if table := expressionTableResult(node.Expression()); table != nil {
		return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(node.Expression()), diagnostics...))
}

func (n *simpleNodeBinder) bindMvApply(node *MvApplyOperator) *SemanticInfo {
	if node == nil || node.Expressions() == nil {
		return nil
	}
	diagnostics := make([]Diagnostic, 0)
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.MvApplyKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.MvApplyParameters)...)
	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		underlying := mvExpandUnderlyingExpression(expression.Expression())
		typ := ExpressionResultType(underlying)
		if topDiagnosticTypeKnown(typ) && !isDynamicType(typ) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(ScalarTypes.Dynamic.SymbolName()), underlying))
		}
	}
	if limit := node.RowLimitClause(); limit != nil {
		diagnostics = append(diagnostics, n.integerDiagnostics(limit.RowLimit())...)
	}
	result := mvApplyIntermediateTable(node, n.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func mvApplyIntermediateTable(node *MvApplyOperator, rowScope *TableSymbol) *TableSymbol {
	builder := NewProjectionBuilder()
	if rowScope != nil {
		builder.AddRange(rowScope.Columns(), ProjectionAddRangeOptions{DoNotRepeat: true})
	}
	defaultSuffix := 1
	if node != nil && node.Expressions() != nil {
		for _, separated := range node.Expressions().Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			expression := separated.Element()
			underlying := mvExpandUnderlyingExpression(expression.Expression())
			name := mvExpandExpressionName(expression.Expression())
			if name == "" {
				name = serializeDefaultColumnName(builder, &defaultSuffix)
			}
			column := NewColumnSymbolWithSource(name, mvApplyExpressionResultType(expression), underlying)
			if source, ok := directMvExpandColumn(underlying); ok {
				column.OriginalColumns = trulyOriginalColumns([]*ColumnSymbol{source})
			}
			builder.Add(column, ProjectionAddOptions{Replace: true})
			if named := mvExpandNamedExpression(expression.Expression()); named != nil {
				named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
			}
		}
	}
	if indexName := mvApplyItemIndexName(node); indexName != "" {
		index := NewColumnSymbol(indexName, ScalarTypes.Long.Type())
		builder.Add(index)
		for _, parameter := range node.Parameters().Elements() {
			if declaration, ok := parameter.Expression().(*NameDeclaration); ok && NameDeclarationSimpleName(declaration) == indexName {
				declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(index))
			}
		}
	}
	return NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(rowScope)
}

func mvApplyExpressionResultType(expression *MvApplyExpression) *TypeSymbol {
	if expression == nil {
		return ScalarTypes.Unknown.Type()
	}
	if primitive := mvExpandDeclaredTypeExpression(expression.ToTypeOf()); primitive != nil {
		return returnTypeFromSymbol(declaredSymbolFromPrimitiveTypeExpression(primitive))
	}
	typ := ExpressionResultType(mvExpandUnderlyingExpression(expression.Expression()))
	if elementType := GetElementType(typ); elementType != nil {
		return elementType
	}
	if isDynamicType(typ) {
		return ScalarTypes.Dynamic.Type()
	}
	if typ == nil || typ == &Error.TypeSymbol {
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func mvApplyItemIndexName(node *MvApplyOperator) string {
	if node == nil || node.Parameters() == nil {
		return ""
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.MvApplyParameters) == QueryOperatorParameters.WithItemIndex {
			if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
				return NameDeclarationSimpleName(declaration)
			}
		}
	}
	return ""
}
