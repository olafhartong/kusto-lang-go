package kustolang

// Ported from the partition visitors in Binder_TreeBinder.cs and
// Binder_NodeBinder.cs at the revision pinned in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitPartition(node *PartitionOperator) {
	if node == nil || node.IsBound() || node.ByExpression() == nil || node.Operand() == nil {
		return
	}
	var scope Expression
	if node.Scope() != nil {
		scope = node.Scope().Expression()
	}
	t.visitPartitionCommon(node.SyntaxNode, node.Parameters(), node.ByExpression(), scope, node.Operand(), func() *SemanticInfo {
		return t.nodeBinder.bindPartition(node)
	})
}

func (t *treeBinder) visitPartitionBy(node *PartitionByOperator) {
	if node == nil || node.IsBound() || node.Entity() == nil || node.Subquery() == nil {
		return
	}
	t.visitPartitionCommon(node.SyntaxNode, node.Parameters(), node.Entity(), nil, node.Subquery(), func() *SemanticInfo {
		return t.nodeBinder.bindPartitionBy(node)
	})
}

func (t *treeBinder) visitPartitionCommon(node *SyntaxNode, parameters *SyntaxListOf[*NamedParameter], key, scope Expression, operand SyntaxElement, bind func() *SemanticInfo) {
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope, oldScopeKind, oldLocalScope := t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope
	oldInPipe, oldNodeInPipe := t.inPipe, t.nodeBinder.inPipe
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRowScope, oldScopeKind, oldLocalScope
		t.inPipe, t.nodeBinder.inPipe = oldInPipe, oldNodeInPipe
		if !committed {
			t.binder.openTableInferredColumns, t.binder.openTableInferredOrder = oldInferredColumns, oldInferredOrder
			tx.rollback()
		}
	}()

	if parameters != nil {
		for _, parameter := range parameters.Elements() {
			if parameter != nil && parameter.Expression() != nil {
				t.visitElement(parameter.Expression())
				if bound := simpleBinderSyntaxNode(parameter.Expression()); bound == nil || !bound.IsBound() {
					return
				}
			}
		}
	}
	t.visitElement(key)
	if bound := simpleBinderSyntaxNode(key); bound == nil || !bound.IsBound() {
		return
	}
	if scope != nil {
		t.visitElement(scope)
		if bound := simpleBinderSyntaxNode(scope); bound == nil || !bound.IsBound() {
			return
		}
	}

	local := NewLocalScope(oldLocalScope)
	if column, ok := directReferencedColumn(key); ok {
		local.AddSymbol(column)
	}
	t.binder.LocalScope = local
	t.binder.scopeKind = ScopeKindNormal
	t.binder.rowScope = oldRowScope
	_, braced := WrapGeneratedSyntaxElement(operand).(*PartitionQuery)
	if braced {
		t.binder.rowScope = nil
		t.binder.openTableInferredColumns = nil
		t.binder.openTableInferredOrder = nil
		markPartitionQueryTabularSource(operand)
	} else if t.binder.rowScope == nil {
		t.binder.rowScope = EmptyTableSymbol
	}
	t.inPipe, t.nodeBinder.inPipe = !braced, !braced
	t.visitElement(operand)
	if bound := simpleBinderSyntaxNode(operand); bound == nil || !bound.IsBound() {
		return
	}

	t.binder.rowScope = oldRowScope
	t.inPipe, t.nodeBinder.inPipe = oldInPipe, oldInPipe
	info := bind()
	if info == nil {
		return
	}
	node.SetSemanticInfo(*info)
	committed = true
}

func (n *simpleNodeBinder) bindPartitionOperand(expression Expression) *SemanticInfo {
	if table := expressionTableResult(expression); table != nil {
		return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(expression)))
}

func (n *simpleNodeBinder) bindPartition(node *PartitionOperator) *SemanticInfo {
	operand := partitionOperandExpression(node.Operand())
	_, braced := node.Operand().(*PartitionQuery)
	diagnostics := n.partitionDiagnostics(node.PartitionKeyword(), node.ByExpression(), operand, !braced, false)
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.PartitionParameters)...)
	if scope := node.Scope(); scope != nil && scope.Expression() != nil {
		typ := ExpressionRawResultType(scope.Expression())
		if typ != nil && typ != &Error.TypeSymbol && typ.SymbolTabularity() != TabularityTabular {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), scope.Expression()))
		}
	}
	return n.partitionResult(operand, diagnostics)
}

func (n *simpleNodeBinder) bindPartitionBy(node *PartitionByOperator) *SemanticInfo {
	diagnostics := n.partitionDiagnostics(node.PartitionByKeyword(), node.Entity(), node.Subquery(), true, true)
	table := expressionTableResult(node.Subquery())
	result := NewTableSymbol("", n.binder.declaredAndInferredColumns(table)...).
		WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) partitionDiagnostics(keyword SyntaxElement, key Expression, subquery Expression, restrictedSubquery, allowContextualRoot bool) []Diagnostic {
	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), keyword))
	}
	if typ := ExpressionResultType(key); topDiagnosticTypeKnown(typ) && isDynamicType(typ) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTypeNotAllowed(typ.SymbolName()), key))
	}
	if typ := ExpressionResultType(key); typ != &Error.TypeSymbol {
		if _, ok := directReferencedColumn(key); !ok && key != nil && key.Width() != 0 && len(key.ContainedDiagnostics()) == 0 {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), key))
		}
	}
	if restrictedSubquery {
		diagnostics = append(diagnostics, n.partitionSubqueryDiagnosticsWithContextualRoot(subquery, allowContextualRoot)...)
	}
	return diagnostics
}

func markPartitionQueryTabularSource(operand SyntaxElement) {
	query, ok := WrapGeneratedSyntaxElement(operand).(*PartitionQuery)
	if !ok || query.Query() == nil {
		return
	}
	root := query.Query()
	for {
		pipe, ok := root.(*PipeExpression)
		if !ok {
			break
		}
		root = pipe.Expression()
	}
	if reference, ok := root.(*NameReference); ok && reference.SyntaxNode != nil {
		reference.SyntaxNode.setPayload(recoveredTabularStatementSourcePayload, true)
	}
}

func (n *simpleNodeBinder) partitionResult(subquery Expression, diagnostics []Diagnostic) *SemanticInfo {
	table := expressionTableResult(subquery)
	if table == nil {
		if typ := ExpressionRawResultType(subquery); typ != nil && typ != &Error.TypeSymbol && typ.SymbolTabularity() != TabularityTabular {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), subquery))
		}
		if n.binder.rowScope != nil {
			return semanticInfoPtr(NewSemanticInfoForSymbol(n.binder.rowScope, &n.binder.rowScope.TypeSymbol, diagnostics...))
		}
		return semanticInfoPtr(NewSemanticInfoForSymbol(EmptyTableSymbol, &EmptyTableSymbol.TypeSymbol, diagnostics...))
	}
	result := NewTableSymbol("", n.binder.declaredAndInferredColumns(table)...).
		WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) partitionSubqueryDiagnostics(expression Expression) []Diagnostic {
	return n.partitionSubqueryDiagnosticsWithContextualRoot(expression, true)
}

func (n *simpleNodeBinder) partitionSubqueryDiagnosticsWithContextualRoot(expression Expression, allowContextualRoot bool) []Diagnostic {
	root := expression
	if query, ok := root.(*PartitionQuery); ok {
		root = query.Query()
	} else if subquery, ok := root.(*PartitionSubquery); ok {
		root = subquery.Subquery()
	}
	return n.postPipeQueryOperatorDiagnostics(root, allowContextualRoot)
}

func (n *simpleNodeBinder) postPipeQueryOperatorDiagnostics(expression Expression, allowContextualRoot bool) []Diagnostic {
	if expression == nil {
		return nil
	}
	var diagnostics []Diagnostic
	root := expression
	for {
		pipe, ok := root.(*PipeExpression)
		if !ok {
			break
		}
		diagnostics = append(diagnostics, n.partitionOperatorDiagnostic(pipe.Operator())...)
		root = pipe.Expression()
	}
	if operator, ok := root.(QueryOperator); ok {
		diagnostics = append(diagnostics, n.partitionOperatorDiagnostic(operator)...)
	} else if _, contextual := root.(*ContextualDataTableExpression); !allowContextualRoot || !contextual {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorExpected(), root))
	}
	return diagnostics
}

func partitionOperandExpression(operand PartitionOperand) Expression {
	switch operand := operand.(type) {
	case *PartitionQuery:
		return operand
	case *PartitionSubquery:
		return operand
	default:
		return nil
	}
}

func (n *simpleNodeBinder) partitionOperatorDiagnostic(operator QueryOperator) []Diagnostic {
	if operator == nil || len(operator.ContainedDiagnostics()) != 0 {
		return nil
	}
	if _, ok := partitionSubqueryOperatorKinds[operator.Kind()]; ok {
		return nil
	}
	keyword := firstSyntaxToken(operator)
	name := ""
	if keyword != nil {
		name = keyword.RawText()
	}
	return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorNotAllowedInContext(name), keyword)}
}

var partitionSubqueryOperatorKinds = map[SyntaxKind]struct{}{
	SyntaxKindAsOperator: {}, SyntaxKindAssertSchemaOperator: {}, SyntaxKindConsumeOperator: {}, SyntaxKindCountOperator: {},
	SyntaxKindDistinctOperator: {}, SyntaxKindEvaluateOperator: {}, SyntaxKindExecuteAndCacheOperator: {}, SyntaxKindExtendOperator: {},
	SyntaxKindFacetOperator: {}, SyntaxKindFilterOperator: {}, SyntaxKindForkOperator: {}, SyntaxKindGetSchemaOperator: {},
	SyntaxKindGraphMatchOperator: {}, SyntaxKindGraphMergeOperator: {}, SyntaxKindGraphShortestPathsOperator: {}, SyntaxKindGraphMarkComponentsOperator: {},
	SyntaxKindGraphWhereNodesOperator: {}, SyntaxKindGraphWhereEdgesOperator: {}, SyntaxKindGraphToTableOperator: {},
	SyntaxKindInvokeOperator: {}, SyntaxKindJoinOperator: {}, SyntaxKindLookupOperator: {}, SyntaxKindMakeGraphOperator: {},
	SyntaxKindMakeSeriesOperator: {}, SyntaxKindMvApplyOperator: {}, SyntaxKindMvExpandOperator: {}, SyntaxKindParseOperator: {},
	SyntaxKindParseWhereOperator: {}, SyntaxKindParseKvOperator: {}, SyntaxKindPartitionByOperator: {}, SyntaxKindPartitionOperator: {},
	SyntaxKindProjectOperator: {}, SyntaxKindProjectAwayOperator: {}, SyntaxKindProjectByNamesOperator: {}, SyntaxKindProjectKeepOperator: {},
	SyntaxKindProjectRenameOperator: {}, SyntaxKindProjectReorderOperator: {}, SyntaxKindReduceByOperator: {}, SyntaxKindRenderOperator: {},
	SyntaxKindSampleOperator: {}, SyntaxKindSampleDistinctOperator: {}, SyntaxKindScanOperator: {}, SyntaxKindSearchOperator: {},
	SyntaxKindSerializeOperator: {}, SyntaxKindSortOperator: {}, SyntaxKindSummarizeOperator: {}, SyntaxKindTakeOperator: {},
	SyntaxKindTopHittersOperator: {}, SyntaxKindTopOperator: {}, SyntaxKindTopNestedOperator: {}, SyntaxKindUnionOperator: {},
}
