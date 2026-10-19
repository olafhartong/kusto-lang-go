package kustolang

// Ported from the MaterializedViewCombineExpression methods in
// Binder_TreeBinder.cs and Binder_NodeBinder.cs.

func (t *treeBinder) visitMaterializedViewCombine(node *MaterializedViewCombineExpression) {
	if t == nil || t.binder == nil || node == nil || node.IsBound() || !supportsMaterializedViewCombine(node) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope = oldRowScope
		t.nodeBinder.inPipe = oldNodeBinderInPipe
	}()

	t.visitElement(node.ViewName())
	t.visitElement(node.BaseClause())
	t.visitElement(node.DeltaClause())

	// Aggregations are evaluated over the delta result, not the ambient row
	// scope. A malformed or non-tabular delta intentionally supplies no scope.
	t.binder.rowScope = materializedViewCombineDeltaScope(node.DeltaClause())
	t.visitElement(node.AggregationsClause())

	t.binder.rowScope = oldRowScope
	t.bindNode(node.SyntaxNode)
}

func supportsMaterializedViewCombine(node *MaterializedViewCombineExpression) bool {
	return node != nil && node.SyntaxNode != nil && node.ChildCount() == 5 &&
		node.MaterializedViewCombineKeyword() != nil && node.ViewName() != nil &&
		supportsMaterializedViewCombineNameClause(node.ViewName()) &&
		supportsMaterializedViewCombineClause(node.BaseClause()) &&
		supportsMaterializedViewCombineClause(node.DeltaClause()) &&
		supportsMaterializedViewCombineClause(node.AggregationsClause())
}

func supportsMaterializedViewCombineNameClause(node *MaterializedViewCombineNameClause) bool {
	return node != nil && node.SyntaxNode != nil && node.ChildCount() == 3 &&
		node.OpenParen() != nil && node.Value() != nil && node.CloseParen() != nil
}

func supportsMaterializedViewCombineClause(node *MaterializedViewCombineClause) bool {
	return node != nil && node.SyntaxNode != nil && node.ChildCount() == 4 &&
		node.Keyword() != nil && node.OpenParen() != nil && node.Expression() != nil && node.CloseParen() != nil
}

func (t *treeBinder) visitMaterializedViewCombineNameClause(node *MaterializedViewCombineNameClause) {
	if !supportsMaterializedViewCombineNameClause(node) || node.IsBound() {
		return
	}
	t.visitElement(node.Value())
	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) visitMaterializedViewCombineClause(node *MaterializedViewCombineClause) {
	if !supportsMaterializedViewCombineClause(node) {
		return
	}
	t.visitElement(node.Expression())
	// Clauses intentionally have no SemanticInfo of their own upstream.
}

func materializedViewCombineDeltaScope(clause *MaterializedViewCombineClause) *TableSymbol {
	if clause == nil {
		return nil
	}
	return expressionTableResult(clause.Expression())
}

func (n *simpleNodeBinder) bindMaterializedViewCombine(node *MaterializedViewCombineExpression) *SemanticInfo {
	if node == nil || node.AggregationsClause() == nil || node.AggregationsClause().Expression() == nil {
		return nil
	}
	aggregations := node.AggregationsClause().Expression()
	if table := expressionTableResult(aggregations); table != nil {
		return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(aggregations)))
}

func (n *simpleNodeBinder) bindMaterializedViewCombineNameClause(node *MaterializedViewCombineNameClause) *SemanticInfo {
	if node == nil || node.Value() == nil {
		return nil
	}

	value := node.Value()
	typ := resultTypeOrError(value)
	if !sameType(typ, ScalarTypes.String.Type()) {
		if !isErrorTypeSymbol(typ) && !sameType(typ, ScalarTypes.Unknown.Type()) {
			return semanticInfoPtr(NewSemanticInfoWithDiagnostics(n.binder.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), value)))
		}
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	if !ExpressionIsLiteral(value) {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics(n.binder.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustBeLiteral(), value)))
	}

	// A valid name clause returns null SemanticInfo upstream.
	return nil
}
