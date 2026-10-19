package kustolang

// Ported from the SearchOperator portions of Binder_TreeBinder.cs,
// Binder_NodeBinder.cs, and Binder_TablesAndColumns.cs at the revision pinned
// in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitSearch(node *SearchOperator) {
	if node == nil || node.IsBound() || node.Parameters() == nil || node.Condition() == nil {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRow, oldKind, oldLocal := t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope
	oldPipe, oldWildcard, oldNodePipe := t.inPipe, t.inWildcard, t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRow, oldKind, oldLocal
		t.inPipe, t.inWildcard, t.nodeBinder.inPipe = oldPipe, oldWildcard, oldNodePipe
		if !committed {
			tx.rollback()
		}
	}()

	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil {
			t.visitElement(parameter)
		}
	}
	if node.DataScope() != nil {
		t.visitElement(node.DataScope())
	}

	// Explicit search inputs are tabular expressions, independent of a pipe row.
	t.binder.rowScope = nil
	if clause := node.InClause(); clause != nil && clause.Expressions() != nil {
		for _, separated := range clause.Expressions().Elements() {
			if separated == nil || !t.bindUnionOperand(separated.Element()) {
				return
			}
		}
	}

	t.binder.rowScope = oldRow
	searchScope := t.binder.searchColumnsTable(node)
	t.binder.rowScope = searchScope
	newSearchPredicateBinder(t.binder, t).bind(node.Condition())
	if !simpleBinderSyntaxNode(node.Condition()).IsBound() {
		return
	}

	// NodeBinder observes the incoming row scope when deciding result properties.
	t.binder.rowScope = oldRow
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (b *Binder) searchColumnsTable(node *SearchOperator) *TableSymbol {
	if node != nil && b.rowScope != nil && node.InClause() == nil {
		return b.rowScope
	}
	return b.getTableOfColumnsUnifiedByNameAndType(b.searchTables(node))
}

func (b *Binder) searchTables(node *SearchOperator) []*TableSymbol {
	if node != nil && node.InClause() != nil && node.InClause().Expressions() != nil {
		var tables []*TableSymbol
		for _, separated := range node.InClause().Expressions().Elements() {
			if separated == nil {
				continue
			}
			switch typ := expressionTableResult(separated.Element()); typ {
			case nil:
			default:
				tables = append(tables, typ)
			}
			if info := boundSemanticInfoForExpression(separated.Element()); info != nil {
				if group, ok := info.ReferencedSymbol().(*GroupSymbol); ok {
					for _, member := range group.Members() {
						if table, ok := member.(*TableSymbol); ok {
							tables = append(tables, table)
						}
					}
				}
			}
		}
		return tables
	}
	if b.rowScope != nil {
		return []*TableSymbol{b.rowScope}
	}
	return b.impliedTables()
}

func (n *simpleNodeBinder) bindSearch(node *SearchOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.SearchParameters)
	if condition := node.Condition(); condition != nil {
		typ := ExpressionResultType(condition)
		if topDiagnosticTypeKnown(typ) && !sameType(typ, ScalarTypes.Bool.Type()) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(ScalarTypes.Bool.SymbolName()), condition))
		}
	}
	if clause := node.InClause(); clause != nil {
		if n.inPipe {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorMustBeFirst(), node.SearchKeyword()))
		}
		if clause.Expressions() != nil {
			for _, separated := range clause.Expressions().Elements() {
				if separated != nil && !syntaxElementIsMissing(separated.Element()) && expressionTableResult(separated.Element()) == nil {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), separated.Element()))
				}
			}
		}
	}
	columns := []*ColumnSymbol{NewColumnSymbol("$table", ScalarTypes.String.Type())}
	searchTable := n.binder.searchColumnsTable(node)
	columns = append(columns, n.binder.declaredAndInferredColumns(searchTable)...)
	result := NewTableSymbol("", columns...).WithIsSorted(false)
	if n.binder.rowScope != nil {
		result = result.WithInheritableProperties(n.binder.rowScope)
	} else {
		result = result.WithIsOpen(searchTable.IsOpen())
	}
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}
