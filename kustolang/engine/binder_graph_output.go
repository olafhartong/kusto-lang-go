package kustolang

// Ported from Binder_NodeBinder.cs graph output operators and the
// Binder_TreeBinder.cs GraphToTable visitor at the pinned upstream revision.

func (t *treeBinder) visitGraphMarkComponents(node *GraphMarkComponentsOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	t.visitGraphOutputOperator(node.SyntaxNode, t.binder.rowScope, func() *SemanticInfo {
		return t.nodeBinder.bindGraphMarkComponents(node, graph)
	}, false)
}

func (t *treeBinder) visitGraphWhereNodes(node *GraphWhereNodesOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	var condition Expression
	if node != nil {
		condition = node.Condition()
	}
	t.visitGraphOutputOperator(node.SyntaxNode, t.binder.rowScope, func() *SemanticInfo {
		return t.nodeBinder.bindGraphWhereNodes(node, graph)
	}, false, condition)
}

func (t *treeBinder) visitGraphWhereEdges(node *GraphWhereEdgesOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	t.visitGraphOutputOperator(node.SyntaxNode, t.binder.rowScope, func() *SemanticInfo {
		return t.nodeBinder.bindGraphWhereEdges(node, graph)
	}, false, node.Condition())
}

func (t *treeBinder) visitGraphToTable(node *GraphToTableOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	t.visitGraphOutputOperator(node.SyntaxNode, nil, func() *SemanticInfo {
		return t.nodeBinder.bindGraphToTable(node, graph)
	}, true, graphToTableParameterExpressions(node)...)
}

func (t *treeBinder) visitGraphOutputOperator(node *SyntaxNode, rowScope *TableSymbol, bind func() *SemanticInfo, publishTables bool, children ...SyntaxElement) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope, oldScopeKind, oldNodeInPipe := t.binder.rowScope, t.binder.scopeKind, t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope, t.binder.scopeKind = oldRowScope, oldScopeKind
		t.nodeBinder.inPipe = oldNodeInPipe
		if !committed {
			tx.rollback()
		}
	}()

	t.binder.rowScope, t.binder.scopeKind = rowScope, ScopeKindNormal
	t.nodeBinder.inPipe = t.inPipe
	for _, child := range children {
		if child != nil {
			t.visitElement(child)
		}
	}
	info := bind()
	if info == nil {
		return
	}
	node.SetSemanticInfo(*info)
	if publishTables {
		if group, ok := info.ReferencedSymbol().(*GroupSymbol); ok && t.binder.LocalScope != nil {
			for _, member := range group.Members() {
				if table, ok := member.(*TableSymbol); ok {
					t.binder.LocalScope.AddSymbol(table)
				}
			}
		}
	}
	committed = true
}

func graphToTableParameterExpressions(node *GraphToTableOperator) []SyntaxElement {
	var result []SyntaxElement
	if node == nil || node.OutputClause() == nil {
		return result
	}
	for _, separated := range node.OutputClause().Elements() {
		if separated == nil || separated.Element() == nil {
			continue
		}
		clause := separated.Element()
		if clause.AsClause() != nil && clause.AsClause().Name() != nil {
			result = append(result, clause.AsClause().Name())
		}
		if clause.Parameters() != nil {
			for _, parameter := range clause.Parameters().Elements() {
				if parameter != nil && parameter.Expression() != nil {
					result = append(result, parameter.Expression())
				}
			}
		}
	}
	return result
}

func (n *simpleNodeBinder) bindGraphMarkComponents(node *GraphMarkComponentsOperator, graph *GraphSymbol) *SemanticInfo {
	diagnostics := n.queryOperatorFirstDiagnostics(node.GraphMarkComponentsKeyword())
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.GraphMarkComponentsParameters)...)
	result := copyGraphOrEmpty(graph, n.binder.rowScope)
	name := graphParameterName(node.Parameters(), QueryOperatorParameters.WithComponentId)
	if name == "" {
		name = "ComponentId"
	}
	nodeShape := result.NodeShape
	if nodeShape == nil {
		nodeShape = EmptyTableSymbol
	}
	nodeShape = nodeShape.AddColumns(NewColumnSymbol(name, ScalarTypes.Long.Type()))
	result = NewGraphSymbol(result.SymbolName(), result.EdgeShape, nodeShape)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindGraphWhereNodes(node *GraphWhereNodesOperator, graph *GraphSymbol) *SemanticInfo {
	return n.bindGraphWhere(node.GraphWhereNodesKeyword(), node.Condition(), graph)
}

func (n *simpleNodeBinder) bindGraphWhereEdges(node *GraphWhereEdgesOperator, graph *GraphSymbol) *SemanticInfo {
	return n.bindGraphWhere(node.GraphWhereEdgesKeyword(), node.Condition(), graph)
}

func (n *simpleNodeBinder) bindGraphWhere(keyword SyntaxElement, condition Expression, graph *GraphSymbol) *SemanticInfo {
	diagnostics := n.queryOperatorFirstDiagnostics(keyword)
	typ := ExpressionResultType(condition)
	if typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Bool.Type()) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(ScalarTypes.Bool.Type().SymbolName()), condition))
	}
	result := copyGraphOrEmpty(graph, n.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindGraphToTable(node *GraphToTableOperator, graph *GraphSymbol) *SemanticInfo {
	diagnostics := n.graphInputDiagnostics(node.GraphToTableKeyword(), graph)
	if graph == nil {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics(diagnostics...))
	}
	clauses := graphToTableClauses(node)
	var result Symbol
	switch len(clauses) {
	case 1:
		result, diagnostics = n.graphToTableClause(clauses[0], graph, diagnostics)
	case 2:
		var nodes, edges *GraphToTableOutputClause
		for _, clause := range clauses {
			switch clause.EntityKeyword().Kind() {
			case SyntaxKindNodesKeyword:
				nodes = clause
			case SyntaxKindGraphEdgesKeyword:
				edges = clause
			}
		}
		if nodes == nil || edges == nil {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMissingGraphEntityType(), node.OutputClause()))
			break
		}
		if nodes.AsClause() == nil {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMultipleGraphOutputsRequireAliases(), nodes))
		}
		if edges.AsClause() == nil {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMultipleGraphOutputsRequireAliases(), edges))
		}
		var nodeTable, edgeTable Symbol
		nodeTable, diagnostics = n.graphToTableClause(nodes, graph, diagnostics)
		edgeTable, diagnostics = n.graphToTableClause(edges, graph, diagnostics)
		result = NewGroupSymbol(nodeTable, edgeTable)
	default:
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticIncorrectNumberOfOutputGraphToTableEntities(), node.OutputClause()))
	}
	if result == nil {
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(Error, diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(result, diagnostics...))
}

func (n *simpleNodeBinder) graphToTableClause(clause *GraphToTableOutputClause, graph *GraphSymbol, diagnostics []Diagnostic) (Symbol, []Diagnostic) {
	if clause == nil || clause.EntityKeyword() == nil {
		return Error, diagnostics
	}
	var shape *TableSymbol
	var parameters []*QueryOperatorParameter
	var hashParameters []*QueryOperatorParameter
	switch clause.EntityKeyword().Kind() {
	case SyntaxKindNodesKeyword:
		shape, parameters = graph.NodeShape, QueryOperatorParameters.GraphToTableNodesParameters
		hashParameters = []*QueryOperatorParameter{QueryOperatorParameters.WithNodeId}
	case SyntaxKindGraphEdgesKeyword:
		shape, parameters = graph.EdgeShape, QueryOperatorParameters.GraphToTableEdgesParameters
		hashParameters = []*QueryOperatorParameter{QueryOperatorParameters.WithSourceId, QueryOperatorParameters.WithTargetId}
	default:
		return Error, diagnostics
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, clause.Parameters(), parameters)...)
	if shape == nil {
		shape = EmptyTableSymbol
	}
	columns := make([]*ColumnSymbol, 0, len(shape.Columns())+len(hashParameters))
	for _, parameter := range hashParameters {
		if name := graphParameterName(clause.Parameters(), parameter); name != "" {
			columns = append(columns, NewColumnSymbol(name, ScalarTypes.Long.Type()))
		}
	}
	columns = append(columns, shape.Columns()...)
	name := ""
	if clause.AsClause() != nil {
		name = NameDeclarationSimpleName(clause.AsClause().Name())
	}
	table := NewTableSymbol(name, columns...)
	if clause.AsClause() != nil && clause.AsClause().Name() != nil {
		clause.AsClause().Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(table))
	}
	return table, diagnostics
}

func (n *simpleNodeBinder) graphInputDiagnostics(keyword SyntaxElement, graph *GraphSymbol) []Diagnostic {
	diagnostics := n.queryOperatorFirstDiagnostics(keyword)
	if graph == nil {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorExpectsGraph(), keyword))
	}
	return diagnostics
}

func (n *simpleNodeBinder) queryOperatorFirstDiagnostics(keyword SyntaxElement) []Diagnostic {
	var diagnostics []Diagnostic
	hasPipedInput := n.inPipe
	if keyword != nil {
		if operator, ok := WrapGeneratedSyntaxElement(keyword.Parent()).(QueryOperator); ok {
			hasPipedInput = HasPipedInput(operator)
		}
	}
	if !hasPipedInput {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), keyword))
	}
	return diagnostics
}

func copyGraphOrEmpty(graph *GraphSymbol, row *TableSymbol) *GraphSymbol {
	if graph != nil {
		return NewGraphSymbol(graph.SymbolName(), graph.EdgeShape, graph.NodeShape)
	}
	return NewGraphSymbol("", row, nil)
}

func graphNodeShape(graph *GraphSymbol) *TableSymbol {
	if graph == nil {
		return nil
	}
	return graph.NodeShape
}

func graphEdgeShape(graph *GraphSymbol) *TableSymbol {
	if graph == nil {
		return nil
	}
	return graph.EdgeShape
}

func graphParameterName(parameters *SyntaxListOf[*NamedParameter], wanted *QueryOperatorParameter) string {
	if parameters == nil || wanted == nil {
		return ""
	}
	for _, parameter := range parameters.Elements() {
		if parameter == nil || parameter.Name() == nil || findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), []*QueryOperatorParameter{wanted}) == nil {
			continue
		}
		if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
			return NameDeclarationSimpleName(declaration)
		}
	}
	return ""
}

func graphToTableClauses(node *GraphToTableOperator) []*GraphToTableOutputClause {
	var clauses []*GraphToTableOutputClause
	if node == nil || node.OutputClause() == nil {
		return clauses
	}
	for _, separated := range node.OutputClause().Elements() {
		if separated != nil && separated.Element() != nil {
			clauses = append(clauses, separated.Element())
		}
	}
	return clauses
}
