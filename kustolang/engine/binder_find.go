package kustolang

// Ported from the FindOperator portions of Binder_TreeBinder.cs,
// Binder_NodeBinder.cs, and Binder_TablesAndColumns.cs at the revision pinned
// in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitFind(node *FindOperator) {
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

	if node.DataScope() != nil {
		t.visitElement(node.DataScope())
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil {
			t.visitElement(parameter)
		}
	}

	// Find inputs are entity expressions and never resolve against a pipe row.
	t.binder.rowScope = nil
	if clause := node.InClause(); clause != nil && clause.Expressions() != nil {
		for _, separated := range clause.Expressions().Elements() {
			if separated == nil || !t.bindUnionOperand(separated.Element()) {
				return
			}
		}
	}

	t.binder.rowScope = t.binder.findColumnsTable(node)
	newSearchPredicateBinder(t.binder, t).bind(node.Condition())
	if !simpleBinderSyntaxNode(node.Condition()).IsBound() {
		return
	}
	if node.Project() != nil {
		t.visitElement(node.Project())
	}
	if node.ProjectAway() != nil {
		t.visitElement(node.ProjectAway())
	}

	// Result construction is independent of the temporary predicate scope.
	t.binder.rowScope = oldRow
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (b *Binder) findTables(node *FindOperator) []*TableSymbol {
	if node != nil && node.InClause() != nil && node.InClause().Expressions() != nil {
		var tables []*TableSymbol
		for _, separated := range node.InClause().Expressions().Elements() {
			if separated == nil {
				continue
			}
			if table := expressionTableResult(separated.Element()); table != nil {
				tables = append(tables, table)
			} else if info := boundSemanticInfoForExpression(separated.Element()); info != nil {
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
	return b.impliedTables()
}

func (b *Binder) findColumnsTable(node *FindOperator) *TableSymbol {
	return b.getTableOfColumnsUnifiedByName(b.findTables(node))
}

func (n *simpleNodeBinder) bindFind(node *FindOperator) *SemanticInfo {
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.FindParameters)
	if condition := node.Condition(); condition != nil {
		typ := ExpressionResultType(condition)
		if topDiagnosticTypeKnown(typ) && !sameType(typ, ScalarTypes.Bool.Type()) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(ScalarTypes.Bool.SymbolName()), condition))
		}
	}

	sourceName := findSourceColumnName(node)
	columns := []*ColumnSymbol{NewColumnSymbol(sourceName, ScalarTypes.String.Type())}
	tables := n.binder.findTables(node)
	resultIsOpen := false
	for _, table := range tables {
		resultIsOpen = resultIsOpen || table.IsOpen()
	}
	refColumns := findReferencedColumns(node.Condition())
	refTables := findReferencedTables(node.Condition(), tables, refColumns)
	explicitPack := false

	project := node.Project()
	if project == nil || project.ProjectKeyword() == nil || project.ProjectKeyword().Kind() == SyntaxKindProjectSmartKeyword {
		columns = append(columns, n.binder.getTableOfCommonColumns(refTables).Columns()...)
		for _, column := range refColumns {
			if !containsColumnNameAndType(columns, column) {
				columns = append(columns, column)
			}
		}
		packColumns := findPackColumns(refTables, columns)
		if len(packColumns) > 0 {
			columns = append(columns, NewColumnSymbolFull("pack_", ScalarTypes.Dynamic.Type(), "", packColumns, nil, nil))
		}
		columns = CombineColumns(CombineKindUnifySameNameAndType, columns)
	} else {
		resultIsOpen = false
		items := project.Columns()
		if items != nil {
			for i, separated := range items.Elements() {
				if separated == nil {
					continue
				}
				expression := separated.Element()
				switch value := expression.(type) {
				case *PackExpression:
					explicitPack = true
					if i == items.Count()-1 {
						packColumns := findPackColumns(refTables, columns)
						columns = append(columns, NewColumnSymbolFull("pack_", ScalarTypes.Dynamic.Type(), "", packColumns, nil, nil))
					} else {
						diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticPackMustBeLastItemInList(), value))
					}
				case *TypedColumnReference:
					if column, ok := value.Column().ReferencedSymbol().(*ColumnSymbol); ok {
						typ := typeFromBoundExpression(value.Type())
						columns = append(columns, NewColumnSymbolFull(column.SymbolName(), typ, "", []*ColumnSymbol{column}, nil, nil))
					}
				case *NameReference:
					if column, ok := value.ReferencedSymbol().(*ColumnSymbol); ok {
						columns = append(columns, column)
					}
				}
			}
		}
		columns = collapseExactDuplicateColumns(columns)
	}

	if away := node.ProjectAway(); away != nil && away.Columns() != nil && away.Columns().Count() > 0 {
		first := away.Columns().At(0)
		if first != nil {
			if _, star := first.Element().(*StarExpression); star {
				kept := columns[:0]
				for _, column := range columns {
					if column.SymbolName() == sourceName || explicitPack && column.SymbolName() == "pack_" {
						kept = append(kept, column)
					}
				}
				columns = kept
			} else {
				remove := map[string]bool{}
				for _, separated := range away.Columns().Elements() {
					if separated != nil {
						if column, ok := referencedSymbol(separated.Element()).(*ColumnSymbol); ok {
							remove[column.SymbolName()] = true
						}
					}
				}
				kept := columns[:0]
				for _, column := range columns {
					if !remove[column.SymbolName()] {
						kept = append(kept, column)
					}
				}
				columns = kept
			}
		}
	}

	result := NewTableSymbol("", columns...).WithIsOpen(resultIsOpen).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func findSourceColumnName(node *FindOperator) string {
	if node != nil && node.Parameters() != nil {
		for _, parameter := range node.Parameters().Elements() {
			if parameter != nil && parameter.Name() != nil && findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.FindParameters) == QueryOperatorParameters.WithSource {
				if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
					if name := NameDeclarationSimpleName(declaration); name != "" {
						return name
					}
				}
				if reference, ok := parameter.Expression().(*NameReference); ok {
					if name := NameReferenceSimpleName(reference); name != "" {
						return name
					}
				}
			}
		}
	}
	return "source_"
}

func findReferencedColumns(root SyntaxElement) []*ColumnSymbol {
	var columns []*ColumnSymbol
	var walk func(SyntaxElement)
	walk = func(element SyntaxElement) {
		if element == nil {
			return
		}
		if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok {
			if column, ok := reference.ReferencedSymbol().(*ColumnSymbol); ok && !containsColumnIdentity(columns, column) {
				columns = append(columns, column)
			}
		}
		for i := 0; i < element.ChildCount(); i++ {
			walk(element.GetChild(i))
		}
	}
	walk(root)
	return columns
}

func findReferencedTables(condition Expression, tables []*TableSymbol, columns []*ColumnSymbol) []*TableSymbol {
	if len(tables) == 1 || findReferencesAllTables(condition) {
		return append([]*TableSymbol(nil), tables...)
	}
	var result []*TableSymbol
	for _, table := range tables {
		for _, column := range columns {
			if table.IsOpen() || table.GetColumn(column.SymbolName()) != nil {
				result = append(result, table)
				break
			}
		}
	}
	return result
}

func findReferencesAllTables(root SyntaxElement) bool {
	found := false
	var walk func(SyntaxElement)
	walk = func(element SyntaxElement) {
		if element == nil || found {
			return
		}
		switch expression := WrapGeneratedSyntaxElement(element).(type) {
		case *StarExpression:
			found = true
		case Expression:
			if info := boundSemanticInfoForExpression(expression); info != nil && info.IsConstant && sameType(info.ResultType, ScalarTypes.Bool.Type()) {
				if value := ExpressionLiteralValueInfo(expression); value != nil {
					_, isString := value.Value.(string)
					if isString {
						found = true
					}
				}
			}
		}
		for i := 0; i < element.ChildCount(); i++ {
			walk(element.GetChild(i))
		}
	}
	walk(root)
	return found
}

func findCommonColumns(tables []*TableSymbol) []*ColumnSymbol {
	if len(tables) == 0 {
		return nil
	}
	var result []*ColumnSymbol
	for _, candidate := range tables[0].Columns() {
		group := []*ColumnSymbol{candidate}
		common := true
		for _, table := range tables[1:] {
			column := table.GetColumn(candidate.SymbolName())
			if column == nil || !sameType(column.Type, candidate.Type) {
				common = false
				break
			}
			group = append(group, column)
		}
		if common {
			if len(group) == 1 {
				result = append(result, candidate)
			} else {
				result = append(result, NewColumnSymbolFull(candidate.SymbolName(), candidate.Type, "", group, nil, nil))
			}
		}
	}
	return result
}

func findPackColumns(tables []*TableSymbol, projected []*ColumnSymbol) []*ColumnSymbol {
	names := map[string]bool{}
	for _, column := range projected {
		names[column.SymbolName()] = true
	}
	var packed []*ColumnSymbol
	for _, table := range tables {
		if table.IsOpen() {
			continue
		}
		for _, column := range table.Columns() {
			if !names[column.SymbolName()] {
				packed = append(packed, column)
				break
			}
		}
	}
	return packed
}

func containsColumnIdentity(columns []*ColumnSymbol, target *ColumnSymbol) bool {
	for _, column := range columns {
		if column == target {
			return true
		}
	}
	return false
}

func containsColumnNameAndType(columns []*ColumnSymbol, target *ColumnSymbol) bool {
	for _, column := range columns {
		if column.SymbolName() == target.SymbolName() && sameType(column.Type, target.Type) {
			return true
		}
	}
	return false
}

func collapseExactDuplicateColumns(columns []*ColumnSymbol) []*ColumnSymbol {
	seen := map[string]map[*TypeSymbol]bool{}
	result := make([]*ColumnSymbol, 0, len(columns))
	for _, column := range columns {
		byType := seen[column.SymbolName()]
		if byType == nil {
			byType = map[*TypeSymbol]bool{}
			seen[column.SymbolName()] = byType
		}
		if byType[column.Type] {
			continue
		}
		byType[column.Type] = true
		result = append(result, column)
	}
	return result
}

func referencedSymbol(expression Expression) Symbol {
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		return accessor.ReferencedSymbol()
	}
	return nil
}
