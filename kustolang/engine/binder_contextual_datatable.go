package kustolang

// This file ports the bounded contextual-datatable path from
// Binder_TreeBinder.VisitRowSchema, CreateColumnsFromRowSchema and
// Binder_NodeBinder.VisitContextualDataTableExpression.

func (t *treeBinder) visitContextualDataTable(node *ContextualDataTableExpression) {
	if node == nil || node.IsBound() || !supportsContextualDataTableExpression(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldScanAliases := t.scanAliases
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.scanAliases = oldScanAliases
		t.nodeBinder.inPipe = oldNodeBinderInPipe
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

	// Neither the contextual id nor row-schema declarations see an incoming row.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	if !t.bindContextualDataTableID(node.Id()) {
		return
	}
	for _, separated := range node.Schema().Columns().Elements() {
		declaration := separated.Element()
		if primitive, ok := declaration.Type().(*PrimitiveTypeExpression); ok {
			t.visitElement(primitive)
			if !primitive.IsBound() {
				return
			}
		}
	}

	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) bindContextualDataTableID(id Expression) bool {
	if id == nil {
		return false
	}
	if syntaxElementIsMissing(id) {
		if syntax := simpleBinderSyntaxNode(id); syntax != nil && !syntax.IsBound() {
			syntax.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
		}
		return true
	}
	t.visitElement(id)
	syntax := simpleBinderSyntaxNode(id)
	return syntax != nil && syntax.IsBound()
}

func supportsContextualDataTableExpression(node *ContextualDataTableExpression) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 3 ||
		node.ContextualDataTableKeyword() == nil ||
		node.ContextualDataTableKeyword().Kind() != SyntaxKindContextualDataTableKeyword ||
		node.ContextualDataTableKeyword().Width() == 0 || node.Id() == nil || node.Schema() == nil ||
		!supportsContextualDataTableID(node.Id()) || !supportsDataTableRowSchema(node.Schema()) {
		return false
	}
	return true
}

func supportsContextualDataTableID(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		return expression.SyntaxNode != nil && expression.ChildCount() == 1 && expression.Token() != nil &&
			expression.Token().Width() != 0 && supportsDataTableLiteral(expression)
	case *NameReference:
		return expression.SyntaxNode != nil && expression.ChildCount() == 1 &&
			(syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression))
	case *ParenthesizedExpression:
		return expression.SyntaxNode != nil && expression.ChildCount() == 3 && expression.OpenParen() != nil &&
			expression.OpenParen().Width() != 0 && expression.CloseParen() != nil && expression.CloseParen().Width() != 0 &&
			expression.Expression() != nil &&
			supportsContextualDataTableID(expression.Expression())
	case *BinaryExpression:
		return supportsContextualDataTableBinary(expression)
	default:
		return false
	}
}

func supportsContextualDataTableBinary(expression *BinaryExpression) bool {
	if expression == nil || expression.SyntaxNode == nil || expression.ChildCount() != 3 ||
		expression.Left() == nil || expression.Operator() == nil || expression.Operator().Width() == 0 ||
		expression.Right() == nil || !supportsContextualDataTableID(expression.Left()) ||
		!supportsContextualDataTableID(expression.Right()) {
		return false
	}
	switch expression.Kind() {
	case SyntaxKindLessThanExpression, SyntaxKindLessThanOrEqualExpression,
		SyntaxKindGreaterThanExpression, SyntaxKindGreaterThanOrEqualExpression,
		SyntaxKindEqualExpression, SyntaxKindNotEqualExpression, SyntaxKindEqualTildeExpression,
		SyntaxKindAndExpression, SyntaxKindOrExpression:
		return true
	default:
		return false
	}
}

func (n *simpleNodeBinder) bindContextualDataTable(node *ContextualDataTableExpression) *SemanticInfo {
	if node == nil || node.Id() == nil || node.Schema() == nil {
		return nil
	}
	diagnostics := make([]Diagnostic, 0)
	columns := n.createContextualDataTableColumns(node.Schema(), &diagnostics)
	n.checkContextualDataTableID(node.Id(), &diagnostics)
	table := NewTableSymbol("", columns...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) createContextualDataTableColumns(schema *RowSchema, diagnostics *[]Diagnostic) []*ColumnSymbol {
	columns := make([]*ColumnSymbol, 0, schema.Columns().Count())
	declaredNames := make(map[string]struct{}, schema.Columns().Count())
	for _, separated := range schema.Columns().Elements() {
		declaration := separated.Element()
		primitive, ok := declaration.Type().(*PrimitiveTypeExpression)
		if !ok {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidColumnDeclaration(), declaration))
			continue
		}
		name := NameDeclarationSimpleName(declaration.Name())
		if _, duplicate := declaredNames[name]; duplicate {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(
				DiagnosticDuplicateColumnDeclaration(name), declaration.Name()))
			continue
		}
		declaredNames[name] = struct{}{}
		typ := ScalarTypes.Unknown.Type()
		if symbol, ok := primitive.ReferencedSymbol().(*TypeSymbol); ok {
			typ = symbol
		}
		column := NewColumnSymbolWithSource(name, typ, declaration.Name())
		columns = append(columns, column)
		declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
	return columns
}

func (n *simpleNodeBinder) checkContextualDataTableID(id Expression, diagnostics *[]Diagnostic) {
	if id == nil {
		return
	}
	typ := ExpressionResultType(id)
	if typ == nil || isErrorTypeSymbol(typ) || sameType(typ, ScalarTypes.Unknown.Type()) && syntaxElementIsMissing(id) {
		return
	}
	if !sameType(typ, ScalarTypes.Guid.Type()) && !sameType(typ, ScalarTypes.Unknown.Type()) {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustHaveType(ScalarTypes.Guid.SymbolName()), id))
		return
	}
	if !ExpressionIsLiteral(id) {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeLiteral(), id))
	}
}
