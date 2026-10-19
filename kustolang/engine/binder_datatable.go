package kustolang

// This file ports the bounded datatable path from Binder_TreeBinder.VisitRowSchema
// and Binder_NodeBinder.VisitDataTableExpression.

func (t *treeBinder) visitDataTable(node *DataTableExpression) {
	if node == nil || node.IsBound() || !supportsDataTableExpression(node) {
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

	// Datatable declarations and values do not resolve against an incoming pipe row.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && parameter.Expression() != nil {
			t.visitElement(parameter.Expression())
			if syntax := simpleBinderSyntaxNode(parameter.Expression()); syntax != nil &&
				!syntax.IsBound() && dataTableExpressionRequiresBinding(parameter.Expression()) {
				return
			}
		}
	}
	for _, separated := range node.Schema().Columns().Elements() {
		declaration := separated.Element()
		t.visitElement(declaration.Type())
		if primitive, ok := declaration.Type().(*PrimitiveTypeExpression); ok && !primitive.IsBound() {
			return
		}
	}
	for _, separated := range node.Values().Elements() {
		if !t.bindDataTableValue(separated.Element()) {
			return
		}
	}

	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func supportsDataTableExpression(node *DataTableExpression) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 7 ||
		node.DataTableKeyword() == nil || node.DataTableKeyword().Kind() != SyntaxKindDataTableKeyword ||
		node.DataTableKeyword().Width() == 0 || node.Parameters() == nil || node.Schema() == nil ||
		node.OpenBracket() == nil || node.LeadingComma() != nil && node.LeadingComma().Kind() != SyntaxKindCommaToken ||
		node.Values() == nil || node.CloseBracket() == nil {
		return false
	}
	parameters, ok := node.GetChild(1).(*SyntaxList)
	if !ok || parameters.ElementType() != "NamedParameter" {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if !supportsDataTableParameter(parameter) {
			return false
		}
	}
	if !supportsDataTableRowSchema(node.Schema()) {
		return false
	}
	values, ok := node.GetChild(5).(*SyntaxList)
	if !ok || values.ElementType() != "SeparatedElement<Expression>" {
		return false
	}
	for i := 0; i < values.Count(); i++ {
		separated, ok := values.At(i).(*SeparatedElement)
		if !ok || separated == nil || separated.Element() == nil {
			return false
		}
		expression, ok := WrapGeneratedSyntaxElement(separated.Element()).(Expression)
		if !ok || !supportsDataTableValue(expression) {
			return false
		}
	}
	return true
}

func supportsDataTableParameter(parameter *NamedParameter) bool {
	if parameter == nil || parameter.SyntaxNode == nil || parameter.ChildCount() != 3 ||
		parameter.Name() == nil || parameter.Name().Name() == nil || parameter.EqualToken() == nil ||
		parameter.Expression() == nil || !supportsSchemaNameDeclaration(parameter.Name()) {
		return false
	}
	switch expression := parameter.Expression().(type) {
	case *LiteralExpression:
		return supportsDataTableLiteral(expression)
	case *NameDeclaration:
		return supportsSchemaNameDeclaration(expression)
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	default:
		return false
	}
}

func supportsDataTableRowSchema(schema *RowSchema) bool {
	if schema == nil || schema.SyntaxNode == nil || schema.ChildCount() != 4 || schema.OpenParen() == nil ||
		schema.LeadingComma() != nil && schema.LeadingComma().Kind() != SyntaxKindCommaToken ||
		schema.Columns() == nil || schema.CloseParen() == nil {
		return false
	}
	columns, ok := schema.GetChild(2).(*SyntaxList)
	if !ok || columns.ElementType() != "SeparatedElement<NameAndTypeDeclaration>" {
		return false
	}
	for i := 0; i < columns.Count(); i++ {
		separated, ok := columns.At(i).(*SeparatedElement)
		if !ok || separated == nil || separated.Element() == nil {
			return false
		}
		declaration, ok := WrapGeneratedSyntaxElement(separated.Element()).(*NameAndTypeDeclaration)
		if !ok || declaration.SyntaxNode == nil || declaration.ChildCount() != 3 || declaration.Name() == nil ||
			declaration.Colon() == nil || declaration.Type() == nil || !supportsSchemaNameDeclaration(declaration.Name()) {
			return false
		}
		switch typ := declaration.Type().(type) {
		case *PrimitiveTypeExpression:
			if typ.SyntaxNode == nil || typ.ChildCount() != 1 || typ.Type() == nil {
				return false
			}
		case *SchemaTypeExpression:
			if !supportsSchemaTypeExpression(typ) {
				return false
			}
		default:
			if simpleBinderSyntaxNode(typ) == nil {
				return false
			}
		}
	}
	return true
}

func supportsDataTableValue(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		return supportsDataTableLiteral(expression)
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *DynamicExpression:
		return expression.SyntaxNode != nil && expression.ChildCount() == 4 && expression.Dynamic() != nil &&
			expression.Dynamic().Kind() == SyntaxKindDynamicKeyword && expression.OpenParen() != nil &&
			expression.Expression() != nil && expression.CloseParen() != nil &&
			supportsDataTableJSONValue(expression.Expression())
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	case *RangeOperator:
		return supportsRangeOperator(expression)
	case *DataTableExpression:
		return supportsDataTableExpression(expression)
	default:
		return false
	}
}

func supportsDataTableJSONValue(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		return supportsDataTableLiteral(expression)
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *DynamicExpression:
		return supportsDataTableValue(expression)
	case *JsonObjectExpression:
		if expression.SyntaxNode == nil || expression.ChildCount() != 3 || expression.OpenBrace() == nil ||
			expression.Pairs() == nil || expression.CloseBrace() == nil {
			return false
		}
		list, ok := expression.GetChild(1).(*SyntaxList)
		if !ok || list.ElementType() != "SeparatedElement<JsonPair>" {
			return false
		}
		for _, separated := range expression.Pairs().Elements() {
			if separated == nil || separated.Element() == nil || !supportsDataTableJSONPair(separated.Element()) {
				return false
			}
		}
		return true
	case *JsonArrayExpression:
		if expression.SyntaxNode == nil || expression.ChildCount() != 3 || expression.OpenBracket() == nil ||
			expression.Values() == nil || expression.CloseBracket() == nil {
			return false
		}
		list, ok := expression.GetChild(1).(*SyntaxList)
		if !ok || list.ElementType() != "SeparatedElement<Expression>" {
			return false
		}
		for _, separated := range expression.Values().Elements() {
			if separated == nil || separated.Element() == nil || !supportsDataTableJSONValue(separated.Element()) {
				return false
			}
		}
		return true
	case *PrefixUnaryExpression:
		return expression.SyntaxNode != nil && expression.ChildCount() == 2 && expression.Operator() != nil &&
			expression.Operator().Kind() == SyntaxKindMinusToken && expression.Expression() != nil &&
			supportsDataTableJSONValue(expression.Expression())
	default:
		return false
	}
}

func supportsDataTableJSONPair(pair *JsonPair) bool {
	return pair != nil && pair.SyntaxNode != nil && pair.ChildCount() == 3 && pair.Name() != nil &&
		pair.Colon() != nil && pair.Value() != nil && supportsDataTableJSONValue(pair.Value())
}

func supportsDataTableLiteral(expression *LiteralExpression) bool {
	if expression == nil || expression.Token() == nil {
		return false
	}
	switch expression.Kind() {
	case SyntaxKindBooleanLiteralExpression, SyntaxKindIntLiteralExpression, SyntaxKindLongLiteralExpression,
		SyntaxKindRealLiteralExpression, SyntaxKindDecimalLiteralExpression, SyntaxKindDateTimeLiteralExpression,
		SyntaxKindTimespanLiteralExpression, SyntaxKindGuidLiteralExpression, SyntaxKindStringLiteralExpression,
		SyntaxKindNullLiteralExpression, SyntaxKindTokenLiteralExpression:
		return true
	default:
		return false
	}
}

func dataTableExpressionRequiresBinding(expression Expression) bool {
	switch expression.(type) {
	case *NameDeclaration:
		return false
	default:
		return true
	}
}

func (t *treeBinder) bindDataTableValue(expression Expression) bool {
	switch expression := expression.(type) {
	case *DynamicExpression:
		t.visitElement(expression.Expression())
		expression.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Dynamic.Type()).WithIsConstant(true))
		return true
	default:
		t.visitElement(expression)
		return simpleBinderSyntaxNode(expression) != nil && simpleBinderSyntaxNode(expression).IsBound()
	}
}

func (n *simpleNodeBinder) bindDataTable(node *DataTableExpression) *SemanticInfo {
	if node == nil || node.Schema() == nil || node.Values() == nil {
		return nil
	}
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.DataTableParameters)
	columns := n.createDataTableColumns(node.Schema(), &diagnostics)
	if len(columns) > 0 && node.Values().Count()%len(columns) != 0 {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticIncorrectNumberOfDataValues(len(columns)), node.Values().SyntaxList))
	}
	for _, separated := range node.Values().Elements() {
		expression := separated.Element()
		typ := ExpressionResultType(expression)
		if typ != nil && typ != &Error.TypeSymbol && !expressionResultIsScalar(expression) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
		}
	}
	table := NewTableSymbol("", columns...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) createDataTableColumns(schema *RowSchema, diagnostics *[]Diagnostic) []*ColumnSymbol {
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
		typ := &Error.TypeSymbol
		if symbol, ok := primitive.ReferencedSymbol().(*TypeSymbol); ok {
			typ = symbol
		}
		column := NewColumnSymbol(name, typ)
		columns = append(columns, column)
		declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
	return columns
}
