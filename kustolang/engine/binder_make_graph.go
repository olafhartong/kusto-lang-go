package kustolang

// Ported from Binder_TreeBinder.cs and Binder_NodeBinder.cs make-graph visitors
// at the revision pinned in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitMakeGraph(node *MakeGraphOperator) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRow, oldKind, oldLocal := t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope
	oldPipe, oldWildcard, oldNodePipe := t.inPipe, t.inWildcard, t.nodeBinder.inPipe
	oldScanAliases := t.scanAliases
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRow, oldKind, oldLocal
		t.inPipe, t.inWildcard, t.nodeBinder.inPipe = oldPipe, oldWildcard, oldNodePipe
		t.scanAliases = oldScanAliases
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

	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil {
			t.visitElement(parameter)
		}
	}
	// Endpoints deliberately resolve against the incoming edge row.
	t.visitElement(node.SourceColumn())
	t.visitElement(node.TargetColumn())

	switch clause := node.WithClause().(type) {
	case *MakeGraphWithTablesAndKeysClause:
		if clause.TablesAndKeys() != nil {
			for _, separated := range clause.TablesAndKeys().Elements() {
				if separated == nil || !t.visitMakeGraphTableAndKey(separated.Element(), oldRow) {
					return
				}
			}
		}
	case *MakeGraphWithImplicitIdClause:
		t.visitElement(clause.Name())
	}

	t.binder.rowScope = oldRow
	t.nodeBinder.inPipe = t.inPipe
	info := t.nodeBinder.bindMakeGraph(node)
	if info == nil {
		return
	}
	node.SetSemanticInfo(*info)

	// Upstream visits this clause after make-graph and uses its subquery result
	// as the operator result while retaining make-graph's diagnostics.
	if clause := node.PartitionedByClause(); clause != nil {
		t.visitElement(clause.Entity())
		t.visitElement(clause.Subquery())
		if clauseInfo := t.nodeBinder.bindMakeGraphPartitionedByClause(clause); clauseInfo != nil {
			clause.SetSemanticInfo(*clauseInfo)
		}
		if subqueryInfo := boundSemanticInfoForExpression(clause.Subquery()); subqueryInfo != nil {
			override := *subqueryInfo
			override.Diagnostics = info.Diagnostics
			node.SetSemanticInfo(override)
		}
	}
	committed = true
}

func (t *treeBinder) visitMakeGraphTableAndKey(clause *MakeGraphTableAndKeyClause, incoming *TableSymbol) bool {
	if clause == nil {
		return true
	}
	t.binder.rowScope = nil
	if !t.bindUnionOperand(clause.Table()) {
		t.visitElement(clause.Table())
	}
	table := expressionTableResult(clause.Table())
	if simpleBinderSyntaxNode(clause.Table()) != nil && !simpleBinderSyntaxNode(clause.Table()).IsBound() {
		return false
	}
	t.binder.rowScope = table
	t.visitElement(clause.Column())
	t.binder.rowScope = incoming
	column := simpleBinderSyntaxNode(clause.Column())
	return column == nil || column.IsBound()
}

func (n *simpleNodeBinder) bindMakeGraph(node *MakeGraphOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.GraphMakeParameters)
	if !HasPipedInput(node) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.MakeGraphKeyword()))
	}
	checkID := func(expression Expression) {
		typ := ExpressionResultType(expression)
		if typ != nil && !IsAnyScalarExceptDynamic(typ) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMakeGraphDynamicNodeIDColumnNotSupported(), expression))
		}
	}
	checkID(node.SourceColumn())
	checkID(node.TargetColumn())

	var nodeShapes []*TableSymbol
	switch clause := node.WithClause().(type) {
	case *MakeGraphWithTablesAndKeysClause:
		if clause.TablesAndKeys() != nil {
			for _, separated := range clause.TablesAndKeys().Elements() {
				if separated == nil || separated.Element() == nil {
					continue
				}
				entry := separated.Element()
				table := expressionTableResult(entry.Table())
				if table == nil && !syntaxElementIsMissing(entry.Table()) {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), entry.Table()))
				} else if table != nil {
					nodeShapes = append(nodeShapes, table)
				}
				checkID(entry.Column())
			}
		}
	case *MakeGraphWithImplicitIdClause:
		name := NameDeclarationSimpleName(clause.Name())
		if name == "" {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMakeGraphImplicitIDShouldNotBeEmpty(), clause.Name()))
		}
		typ := ExpressionResultType(node.SourceColumn())
		if typ == nil {
			typ = ScalarTypes.Unknown.Type()
		}
		table := NewTableSymbol("", NewColumnSymbol(name, typ))
		nodeShapes = append(nodeShapes, table)
		if clause.Name() != nil {
			clause.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(table.Columns()[0]))
		}
	}
	graph := NewGraphSymbolFromNodeShapes("", n.binder.rowScope, nodeShapes)
	return semanticInfoPtr(NewSemanticInfoForSymbol(graph, &graph.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindMakeGraphPartitionedByClause(node *MakeGraphPartitionedByClause) *SemanticInfo {
	if node == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if _, ok := directReferencedColumn(node.Entity()); !ok {
		typ := ExpressionResultType(node.Entity())
		if typ != nil && !isErrorTypeSymbol(typ) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), node.Entity()))
		}
	}
	diagnostics = append(diagnostics, n.partitionSubqueryDiagnostics(node.Subquery())...)

	table := expressionTableResult(node.Subquery())
	if table == nil {
		table = NewTableSymbol("")
	}
	result := NewTableSymbol("", n.binder.declaredAndInferredColumns(table)...).
		WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}
