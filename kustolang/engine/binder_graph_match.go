package kustolang

// Ported from Binder_TreeBinder.cs graph visitors, Binder_NodeBinder.cs graph
// operators, and Binder_Misc.cs graph pattern declaration helpers.

func (t *treeBinder) visitGraphMatch(node *GraphMatchOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	t.visitGraphPatternOperator(node.SyntaxNode, graph, node.Parameters(), node.Patterns(), node.WhereClause(), node.ProjectClause(),
		func() *SemanticInfo { return t.nodeBinder.bindGraphMatch(node, graph) })
}

func (t *treeBinder) visitGraphShortestPaths(node *GraphShortestPathsOperator, graph *GraphSymbol) {
	if node == nil {
		return
	}
	t.visitGraphPatternOperator(node.SyntaxNode, graph, node.Parameters(), node.Patterns(), node.WhereClause(), node.ProjectClause(),
		func() *SemanticInfo { return t.nodeBinder.bindGraphShortestPaths(node, graph) })
}

func (t *treeBinder) visitGraphPatternOperator(node *SyntaxNode, graph *GraphSymbol, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], where *WhereClause, project *ProjectClause, bind func() *SemanticInfo) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope, oldScopeKind, oldLocalScope := t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRowScope, oldScopeKind, oldLocalScope
		if !committed {
			tx.rollback()
		}
	}()

	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	t.binder.LocalScope = NewLocalScope(oldLocalScope)
	if parameters != nil {
		for _, parameter := range parameters.Elements() {
			if parameter != nil && parameter.Expression() != nil {
				t.visitElement(parameter.Expression())
			}
		}
	}
	t.bindGraphPatternDeclarations(graph, patterns)
	if where != nil {
		t.visitElement(where.Condition())
	}
	if project != nil && project.Expressions() != nil {
		for _, expression := range project.Expressions().Elements() {
			if expression != nil {
				t.visitElement(expression.Element())
			}
		}
	}
	info := bind()
	if info == nil {
		return
	}
	node.SetSemanticInfo(*info)
	committed = true
}

func (t *treeBinder) bindGraphPatternDeclarations(graph *GraphSymbol, patterns *SeparatedSyntaxListOf[*GraphMatchPattern]) {
	if patterns == nil {
		return
	}
	edgeTuple, nodeTuple := NewTupleSymbol(nil), NewTupleSymbol(nil)
	if graph != nil {
		edgeTuple = NewTupleSymbol(t.binder.declaredAndInferredColumns(graph.EdgeShape), graph.EdgeShape)
		if graph.NodeShape != nil {
			nodeTuple = NewTupleSymbol(t.binder.declaredAndInferredColumns(graph.NodeShape), graph.NodeShape)
		}
	}
	for _, separated := range patterns.Elements() {
		if separated == nil || separated.Element() == nil || separated.Element().PatternElements() == nil {
			continue
		}
		for _, notation := range separated.Element().PatternElements().Elements() {
			var declaration *NameDeclaration
			var tuple *TupleSymbol
			switch notation := notation.(type) {
			case *GraphMatchPatternNode:
				declaration, tuple = notation.Name(), nodeTuple
			case *GraphMatchPatternEdge:
				declaration, tuple = notation.Name(), edgeTuple
			}
			if declaration == nil || NameDeclarationSimpleName(declaration) == "" {
				continue
			}
			local := NewVariableSymbol(NameDeclarationSimpleName(declaration), tuple, false, nil, nil)
			declaration.SetSemanticInfo(NewSemanticInfoForSymbol(local, &tuple.TypeSymbol))
			t.binder.LocalScope.AddSymbol(local)
		}
	}
}

func (n *simpleNodeBinder) bindGraphMatch(node *GraphMatchOperator, graph *GraphSymbol) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindGraphPatternOperator(graph, node.GraphMatchKeyword(), node.Patterns(), node.WhereClause(), node.ProjectClause())
}

func (n *simpleNodeBinder) bindGraphShortestPaths(node *GraphShortestPathsOperator, graph *GraphSymbol) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindGraphPatternOperator(graph, node.GraphShortestPathsKeyword(), node.Patterns(), node.WhereClause(), node.ProjectClause())
}

func (n *simpleNodeBinder) bindGraphPatternOperator(graph *GraphSymbol, keyword SyntaxElement, patterns *SeparatedSyntaxListOf[*GraphMatchPattern], where *WhereClause, project *ProjectClause) *SemanticInfo {
	diagnostics := n.queryOperatorFirstDiagnostics(keyword)
	if graph == nil {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorExpectsGraph(), keyword))
	}
	for _, item := range checkGraphMatchPatternsLocated(patterns) {
		location := item.location
		if location == nil {
			location = patterns
		}
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(item.diagnostic, location))
	}
	if where != nil {
		typ := ExpressionResultType(where.Condition())
		if typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Unknown.Type()) && !sameType(typ, ScalarTypes.Bool.Type()) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(ScalarTypes.Bool.Type().SymbolName()), where.Condition()))
		}
	}
	if project == nil {
		if graph == nil {
			return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(Error, diagnostics...))
		}
		return semanticInfoPtr(NewSemanticInfoForSymbol(graph, &graph.TypeSymbol, diagnostics...))
	}

	builder := NewProjectionBuilder()
	rangedEdges := graphRangedEdgeNames(patterns)
	if project.Expressions() != nil {
		for _, separated := range project.Expressions().Elements() {
			if separated != nil {
				n.addGraphProjectionExpression(separated.Element(), rangedEdges, builder, &diagnostics)
			}
		}
	}
	table := NewTableSymbol("", builder.GetProjection()...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) addGraphProjectionExpression(expression Expression, rangedEdges map[string]struct{}, builder *ProjectionBuilder, diagnostics *[]Diagnostic) {
	if expression == nil {
		return
	}
	name := graphProjectionName(expression)
	if name == "" {
		return
	}
	typ := ExpressionResultType(graphProjectionValue(expression))
	if typ == nil || typ == &Error.TypeSymbol {
		typ = ScalarTypes.Unknown.Type()
	}
	if graphExpressionReferencesAny(expression, rangedEdges) {
		typ = ScalarTypes.GetDynamicArray(typ).Type()
	}
	column := NewColumnSymbolWithSource(name, typ, graphProjectionValue(expression))
	builder.Declare(column, diagnostics, expression)
	if named, ok := expression.(*SimpleNamedExpression); ok && named.Name() != nil {
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
}

func graphProjectionValue(expression Expression) Expression {
	if named, ok := expression.(*SimpleNamedExpression); ok {
		return named.Expression()
	}
	return expression
}

func graphProjectionName(expression Expression) string {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return NameDeclarationSimpleName(expression.Name())
	case *PathExpression:
		return graphSelectorProjectionName(expression.Expression(), expression.Selector())
	case *ElementExpression:
		return graphSelectorProjectionName(expression.Expression(), expression.Selector())
	default:
		return expressionResultName(expression, "", nil, nil)
	}
}

func graphSelectorProjectionName(left, selector Expression) string {
	leftName := ""
	switch left.(type) {
	case *PathExpression, *ElementExpression:
		leftName = graphProjectionName(left)
	default:
		leftName = expressionResultName(left, "", nil, nil)
	}
	rightName := expressionResultName(selector, "", nil, nil)
	if leftName != "" {
		return leftName + "_" + rightName
	}
	return rightName
}

func graphRangedEdgeNames(patterns *SeparatedSyntaxListOf[*GraphMatchPattern]) map[string]struct{} {
	names := map[string]struct{}{}
	if patterns == nil {
		return names
	}
	for _, separated := range patterns.Elements() {
		if separated == nil || separated.Element() == nil || separated.Element().PatternElements() == nil {
			continue
		}
		for _, notation := range separated.Element().PatternElements().Elements() {
			if edge, ok := notation.(*GraphMatchPatternEdge); ok && edge.Range() != nil && edge.Name() != nil {
				names[NameDeclarationSimpleName(edge.Name())] = struct{}{}
			}
		}
	}
	return names
}

func graphExpressionReferencesAny(expression Expression, names map[string]struct{}) bool {
	found := false
	var walk func(SyntaxElement)
	walk = func(element SyntaxElement) {
		if element == nil || found {
			return
		}
		if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok {
			if _, exists := names[NameReferenceSimpleName(reference)]; exists {
				found = true
				return
			}
		}
		for i := 0; i < element.ChildCount(); i++ {
			walk(element.GetChild(i))
		}
	}
	walk(expression)
	return found
}

func expressionGraphResult(expression Expression) *GraphSymbol {
	if expression == nil {
		return nil
	}
	accessor, ok := any(expression).(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	})
	if !ok {
		return nil
	}
	info, ok := accessor.GetSemanticInfo()
	if !ok {
		return nil
	}
	switch symbol := info.ReferencedSymbol().(type) {
	case *GraphSymbol:
		return symbol
	case *VariableSymbol:
		graph, _ := symbol.Type.(*GraphSymbol)
		return graph
	default:
		return nil
	}
}
