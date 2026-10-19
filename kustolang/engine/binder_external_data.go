package kustolang

// This file ports the bounded ExternalDataExpression path from
// Binder_TreeBinder.VisitExternalDataExpression, CreateColumnsFromRowSchema,
// and Binder_NodeBinder.CheckIsExactType.

func (t *treeBinder) visitExternalData(node *ExternalDataExpression) {
	if node == nil || node.IsBound() || !supportsExternalDataExpression(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe, oldInWildcard, oldScanAliases := t.inPipe, t.inWildcard, t.scanAliases
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRowScope, oldScopeKind, oldLocalScope
		t.inPipe, t.inWildcard, t.scanAliases = oldInPipe, oldInWildcard, oldScanAliases
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

	// Upstream binds child expressions in the current row scope, but does not
	// bind the NamedParameter or with-clause containers themselves.
	for _, parameter := range node.Parameters().Elements() {
		if !t.bindExternalDataChild(parameter.Expression()) {
			return
		}
	}
	for _, separated := range node.Schema().Columns().Elements() {
		if primitive, ok := separated.Element().Type().(*PrimitiveTypeExpression); ok {
			t.visitElement(primitive)
			if !primitive.IsBound() {
				return
			}
		}
	}
	for _, separated := range node.URIs().Elements() {
		if !t.bindExternalDataChild(separated.Element()) {
			return
		}
	}
	if clause := node.WithClause(); clause != nil {
		for _, separated := range clause.Properties().Elements() {
			if !t.bindExternalDataChild(separated.Element().Expression()) {
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

func (t *treeBinder) bindExternalDataChild(expression Expression) bool {
	// Bare property values can parse as declaration-shaped names. Upstream
	// leaves those containers unbound while still committing externaldata.
	if _, ok := expression.(*NameDeclaration); ok {
		return true
	}
	t.visitElement(expression)
	syntax := simpleBinderSyntaxNode(expression)
	return syntax == nil || syntax.IsBound()
}

func supportsExternalDataExpression(node *ExternalDataExpression) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 7 || node.ExternalDataKeyword() == nil ||
		node.ExternalDataKeyword().Kind() != SyntaxKindExternalDataKeyword || node.ExternalDataKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Schema() == nil || node.OpenBracket() == nil || node.URIs() == nil ||
		node.CloseBracket() == nil || !supportsDataTableRowSchema(node.Schema()) {
		return false
	}
	parameters, ok := node.GetChild(1).(*SyntaxList)
	if !ok || parameters.ElementType() != "NamedParameter" {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if !supportsExternalDataParameter(parameter) {
			return false
		}
	}
	uris, ok := node.GetChild(4).(*SyntaxList)
	if !ok || uris.ElementType() != "SeparatedElement<Expression>" {
		return false
	}
	for i := 0; i < uris.Count(); i++ {
		separated, ok := uris.At(i).(*SeparatedElement)
		if !ok || separated == nil || separated.Element() == nil {
			return false
		}
		expression, ok := WrapGeneratedSyntaxElement(separated.Element()).(Expression)
		if !ok || !supportsExternalDataChild(expression) {
			return false
		}
	}
	return supportsExternalDataWithClause(node.WithClause())
}

func supportsExternalDataParameter(parameter *NamedParameter) bool {
	return parameter != nil && parameter.SyntaxNode != nil && parameter.ChildCount() == 3 &&
		parameter.Name() != nil && supportsSchemaNameDeclaration(parameter.Name()) && parameter.EqualToken() != nil &&
		parameter.Expression() != nil && supportsExternalDataChild(parameter.Expression())
}

func supportsExternalDataChild(expression Expression) bool {
	if expression == nil {
		return false
	}
	if simpleBinderSyntaxNode(expression) != nil {
		return true
	}
	switch expression.(type) {
	case *NameDeclaration, *TypeOfLiteralExpression:
		return true
	default:
		return false
	}
}

func supportsExternalDataWithClause(clause *ExternalDataWithClause) bool {
	if clause == nil {
		return true
	}
	if clause.SyntaxNode == nil || clause.ChildCount() != 4 || clause.WithKeyword() == nil || clause.OpenParen() == nil ||
		clause.Properties() == nil || clause.CloseParen() == nil {
		return false
	}
	properties, ok := clause.GetChild(2).(*SyntaxList)
	if !ok || properties.ElementType() != "SeparatedElement<NamedParameter>" {
		return false
	}
	for i := 0; i < properties.Count(); i++ {
		separated, ok := properties.At(i).(*SeparatedElement)
		if !ok || separated == nil || separated.Element() == nil {
			return false
		}
		parameter, ok := WrapGeneratedSyntaxElement(separated.Element()).(*NamedParameter)
		if !ok || !supportsExternalDataParameter(parameter) {
			return false
		}
	}
	return true
}

func (n *simpleNodeBinder) bindExternalData(node *ExternalDataExpression) *SemanticInfo {
	diagnostics := make([]Diagnostic, 0)
	columns := n.createExternalDataColumns(node.Schema(), &diagnostics)
	for _, separated := range node.URIs().Elements() {
		n.checkExternalDataURI(separated.Element(), &diagnostics)
	}
	table := NewTableSymbol("", columns...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) createExternalDataColumns(schema *RowSchema, diagnostics *[]Diagnostic) []*ColumnSymbol {
	columns := make([]*ColumnSymbol, 0, schema.Columns().Count())
	declared := make(map[string]struct{}, schema.Columns().Count())
	for _, separated := range schema.Columns().Elements() {
		declaration := separated.Element()
		primitive, ok := declaration.Type().(*PrimitiveTypeExpression)
		if !ok {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidColumnDeclaration(), declaration))
			continue
		}
		name := NameDeclarationSimpleName(declaration.Name())
		if _, exists := declared[name]; exists {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(name), declaration.Name()))
			continue
		}
		declared[name] = struct{}{}
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

func (n *simpleNodeBinder) checkExternalDataURI(uri Expression, diagnostics *[]Diagnostic) {
	if uri == nil || len(uri.GetContainedSyntaxDiagnostics()) != 0 ||
		len(GetContainedDiagnosticsPreserveDuplicates(uri, DiagnosticsIncludeSemantic)) != 0 {
		return
	}
	typ := ExpressionResultType(uri)
	if typ == nil || isErrorTypeSymbol(typ) || sameType(typ, ScalarTypes.Unknown.Type()) {
		return
	}
	if !SymbolAssignableTo(typ, ScalarTypes.String.Type(), ConversionNone) {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), uri))
	}
}
