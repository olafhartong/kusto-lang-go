package kustolang

// Ported from the compatibility implementation Binder_TreeBinder and Binder_NodeBinder
// GraphMergeOperator visitors. The operator was removed from current Kusto.

func (t *treeBinder) visitGraphMerge(node *GraphMergeOperator, left *GraphSymbol) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		if !committed {
			tx.rollback()
		}
	}()

	// The right graph is an invocation expression and cannot see the incoming
	// graph's edge columns.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	t.binder.LocalScope = oldLocalScope
	t.nodeBinder.joinRightRowScope = nil
	t.visitElement(node.Graph())
	if !expressionIsBound(node.Graph()) {
		return
	}
	right := expressionGraphResult(node.Graph())

	// The optional on clause uses the same left/right edge scopes as join.
	t.binder.LocalScope = oldLocalScope
	if node.OnClause() != nil {
		t.binder.rowScope = graphEdgeShape(left)
		t.nodeBinder.joinRightRowScope = graphEdgeShape(right)
		t.visitElement(node.OnClause())
	}

	t.nodeBinder.inPipe = t.inPipe
	info := t.nodeBinder.bindGraphMerge(node, left, right)
	if info == nil {
		return
	}
	node.SetSemanticInfo(*info)
	committed = true
}

func (n *simpleNodeBinder) bindGraphMerge(node *GraphMergeOperator, left, right *GraphSymbol) *SemanticInfo {
	if node == nil {
		return nil
	}
	diagnostics := n.queryOperatorFirstDiagnostics(node.GraphMergeKeyword())
	if left == nil {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticQueryOperatorExpectsGraph(), node.GraphMergeKeyword()))
	}
	if right == nil {
		if typ := resultTypeOrError(node.Graph()); typ != &Error.TypeSymbol {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
				DiagnosticGraphExpected(), node.Graph()))
		}
	}

	merged := MergeGraphSymbols(left, right)
	if merged == nil {
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(Error, diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForSymbol(merged, &merged.TypeSymbol, diagnostics...))
}
