package kustolang

import "fmt"

// Ported from Binder_TreeBinder.cs VisitFacetOperator. The optional with
// expression is a fork-style pipeline rooted in the table entering facet.
func (t *treeBinder) visitFacet(node *FacetOperator) {
	if node == nil || node.IsBound() {
		return
	}
	oldRowScope, oldInPipe, oldNodeInPipe := t.binder.rowScope, t.inPipe, t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope = oldRowScope
		t.inPipe, t.nodeBinder.inPipe = oldInPipe, oldNodeInPipe
	}()

	if expressions := node.Expressions(); expressions != nil {
		for _, separated := range expressions.Elements() {
			if separated != nil {
				t.visitElement(separated.Element())
			}
		}
	}

	t.binder.rowScope = oldRowScope
	if t.binder.rowScope == nil {
		// A facet with-clause is a fork-style pipeline. Upstream binds a
		// standalone facet's nested operator against RowScopeOrEmpty, which
		// preserves unknown-name diagnostics without inventing input columns.
		t.binder.rowScope = NewTableSymbol("")
	}
	t.inPipe, t.nodeBinder.inPipe = true, true
	switch clause := node.WithClause().(type) {
	case *FacetWithExpressionClause:
		t.visitElement(clause.Expression())
	case *FacetWithOperatorClause:
		t.visitElement(clause.Operator())
	}

	t.binder.rowScope = oldRowScope
	t.inPipe, t.nodeBinder.inPipe = oldInPipe, oldNodeInPipe
	t.bindNode(node.SyntaxNode)
}

// Ported from Binder_NodeBinder.cs VisitFacetOperator.
func (n *simpleNodeBinder) bindFacet(node *FacetOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if !n.hasPipedInput(node) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticQueryOperatorCannotBeFirst(), node.FacetKeyword()))
	}

	members := make([]Symbol, 0)
	if with := facetWithExpression(node.WithClause()); with != nil {
		if _, operatorClause := node.WithClause().(*FacetWithOperatorClause); operatorClause {
			diagnostics = append(diagnostics, n.facetWithDiagnostics(with)...)
		}
		if table := expressionTableResult(with); table != nil {
			// Binder_NodeBinder adds the exact with-clause result table to the
			// group; its identity, name, columns, and table state are preserved.
			members = append(members, table)
		} else if typ := ExpressionRawResultType(with); typ != nil && typ != &Error.TypeSymbol && typ.SymbolTabularity() != TabularityTabular {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), with))
		}
	}

	if expressions := node.Expressions(); expressions != nil {
		facetNumber := 0
		for _, separated := range expressions.Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			expression := separated.Element()
			column := GetResultColumn(expression)
			isColumn := column != nil
			if !isColumn && !isErrorTypeSymbol(resultTypeOrError(expression)) {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), expression))
			}
			typ := ExpressionResultType(expression)
			if isColumn && topDiagnosticTypeKnown(typ) && isDynamicType(typ) {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTypeNotAllowed(typ.SymbolName()), expression))
			}

			name := expressionResultName(expression, "", n.binder.rowScope, n.binder.Globals)
			column = getOrDeclareFacetColumn(expression, name, typ)
			countColumn := getOrDeclareFacetColumn(expression, "count_"+name, ScalarTypes.Long.Type())
			facetNumber++
			tableName := "Facet"
			if facetNumber > 1 {
				tableName = fmt.Sprintf("Facet_%d", facetNumber)
			}
			table := NewTableSymbol(tableName, column, countColumn).
				WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
			members = append(members, table)
		}
	}

	group := NewGroupSymbol(members...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(group, &group.TypeSymbol, diagnostics...))
}

// Mirrors Binder.GetOrDeclareColumnForExpression for facet outputs. Renamed
// columns retain both the original column and the facet expression so semantic
// consumers can recover their physical lineage.
func getOrDeclareFacetColumn(expression Expression, name string, typ *TypeSymbol) *ColumnSymbol {
	if column := GetResultColumn(expression); column != nil {
		if name == "" || column.SymbolName() == name {
			return column
		}
		if typ == nil {
			typ = column.Type
		}
		return NewColumnSymbolFull(name, typ, "", []*ColumnSymbol{column}, expression, nil)
	}
	if typ == nil {
		typ = resultTypeOrError(expression)
	}
	return NewColumnSymbolWithSource(name, typ, expression)
}

func facetWithExpression(clause FacetWithClause) Expression {
	switch clause := clause.(type) {
	case *FacetWithExpressionClause:
		return clause.Expression()
	case *FacetWithOperatorClause:
		expression, _ := clause.Operator().(Expression)
		return expression
	default:
		return nil
	}
}

func (n *simpleNodeBinder) facetWithDiagnostics(expression Expression) []Diagnostic {
	var diagnostics []Diagnostic
	root := expression
	for {
		pipe, ok := root.(*PipeExpression)
		if !ok {
			break
		}
		diagnostics = append(diagnostics, n.forkQueryOperatorDiagnostic(pipe.Operator())...)
		root = pipe.Expression()
	}
	if operator, ok := root.(QueryOperator); ok {
		diagnostics = append(diagnostics, n.forkQueryOperatorDiagnostic(operator)...)
	}
	return diagnostics
}
