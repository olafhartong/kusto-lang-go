package kustolang

// Ported from Binder_TreeBinder.VisitRowSchema and the parse-family visitors in
// Binder_NodeBinder.cs at the revision pinned in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitParseOperator(node QueryOperator) {
	if node == nil {
		return
	}
	syntax := simpleBinderSyntaxNode(node)
	if syntax == nil || syntax.IsBound() {
		return
	}

	tx := newUnionBindingTransaction(syntax)
	committed := false
	oldRow, oldKind, oldLocal := t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope
	oldInPipe, oldNodeInPipe := t.inPipe, t.nodeBinder.inPipe
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope, t.binder.scopeKind, t.binder.LocalScope = oldRow, oldKind, oldLocal
		t.inPipe, t.nodeBinder.inPipe = oldInPipe, oldNodeInPipe
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	switch operator := node.(type) {
	case *ParseOperator:
		t.visitParseParameters(operator.Parameters())
		t.visitElement(operator.Expression())
		t.visitParsePatterns(operator.Patterns())
	case *ParseWhereOperator:
		t.visitParseParameters(operator.Parameters())
		t.visitElement(operator.Expression())
		t.visitParsePatterns(operator.Patterns())
	case *ParseKvOperator:
		t.visitElement(operator.Expression())
		t.visitParseRowSchema(operator.Keys())
		if operator.WithClause() != nil && operator.WithClause().Properties() != nil {
			for _, separated := range operator.WithClause().Properties().Elements() {
				if separated != nil && separated.Element() != nil {
					t.visitElement(separated.Element().Expression())
				}
			}
		}
	default:
		return
	}

	t.bindNode(syntax)
	if !syntax.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) visitParseParameters(parameters *SyntaxListOf[*NamedParameter]) {
	if parameters == nil {
		return
	}
	for _, parameter := range parameters.Elements() {
		if parameter != nil {
			t.visitElement(parameter.Expression())
		}
	}
}

func (t *treeBinder) visitParsePatterns(patterns *SyntaxListOf[SyntaxNodeElement]) {
	if patterns == nil {
		return
	}
	for _, part := range patterns.Elements() {
		switch part := part.(type) {
		case *NameAndTypeDeclaration:
			t.visitElement(part.Type())
		case *NameDeclaration, *StarExpression:
			// These are declarations and pattern markers, not input expressions.
		default:
			t.visitElement(part)
		}
	}
}

func (t *treeBinder) visitParseRowSchema(schema *RowSchema) {
	if schema == nil || schema.Columns() == nil {
		return
	}
	for _, separated := range schema.Columns().Elements() {
		if separated != nil && separated.Element() != nil {
			t.visitElement(separated.Element().Type())
		}
	}
}

func (n *simpleNodeBinder) bindParse(node *ParseOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindParseCommon(node.ParseKeyword(), node.Expression(), node.Patterns(), node.Parameters())
}

func (n *simpleNodeBinder) bindParseWhere(node *ParseWhereOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	return n.bindParseCommon(node.ParseKeyword(), node.Expression(), node.Patterns(), node.Parameters())
}

func (n *simpleNodeBinder) bindParseCommon(keyword SyntaxElement, source Expression, patterns *SyntaxListOf[SyntaxNodeElement], parameters *SyntaxListOf[*NamedParameter]) *SemanticInfo {
	diagnostics := queryOperatorParameterDiagnostics(n.binder, parameters, QueryOperatorParameters.ParseParameters)
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), keyword))
	}
	n.checkParseScalar(source, &diagnostics)
	columns := append([]*ColumnSymbol(nil), n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
	declared := map[string]struct{}{}

	if patterns != nil {
		parts := patterns.Elements()
		for i, part := range parts {
			n.checkParsePatternPart(parts, i, part, &diagnostics)
			switch declaration := part.(type) {
			case *NameDeclaration:
				name := NameDeclarationSimpleName(declaration)
				if n.declareParseColumn(name, declaration, declared, &diagnostics) {
					column := NewColumnSymbolWithSource(name, ScalarTypes.String.Type(), source)
					columns = append(columns, column)
					declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				}
			case *NameAndTypeDeclaration:
				primitive, ok := declaration.Type().(*PrimitiveTypeExpression)
				if !ok {
					continue
				}
				name := NameDeclarationSimpleName(declaration.Name())
				if n.declareParseColumn(name, declaration.Name(), declared, &diagnostics) {
					typ := parsePrimitiveType(primitive)
					column := NewColumnSymbolWithSource(name, typ, source)
					columns = append(columns, column)
					declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				}
			}
		}
	}

	table := NewTableSymbol("", columns...).WithInheritableProperties(n.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindParseKv(node *ParseKvOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.ParseKvKeyword()))
	}
	n.checkParseScalar(node.Expression(), &diagnostics)
	columns := append([]*ColumnSymbol(nil), n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
	declared := map[string]struct{}{}
	if node.Keys() != nil && node.Keys().Columns() != nil {
		for _, separated := range node.Keys().Columns().Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			declaration := separated.Element()
			primitive, ok := declaration.Type().(*PrimitiveTypeExpression)
			if !ok {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidColumnDeclaration(), declaration))
				continue
			}
			name := NameDeclarationSimpleName(declaration.Name())
			if n.declareParseColumn(name, declaration.Name(), declared, &diagnostics) {
				column := NewColumnSymbolWithSource(name, parsePrimitiveType(primitive), declaration.Name())
				columns = append(columns, column)
				declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
			}
		}
	}
	if node.WithClause() != nil {
		diagnostics = append(diagnostics, queryOperatorParameterElementDiagnostics(n.binder, separatedNamedParameters(node.WithClause().Properties()), QueryOperatorParameters.ParseKvWithProperties)...)
	}
	table := NewTableSymbol("", columns...).WithInheritableProperties(n.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func separatedNamedParameters(parameters *SeparatedSyntaxListOf[*NamedParameter]) []*NamedParameter {
	if parameters == nil {
		return nil
	}
	elements := make([]*NamedParameter, 0, parameters.Count())
	for _, separated := range parameters.Elements() {
		if separated != nil && separated.Element() != nil {
			elements = append(elements, separated.Element())
		}
	}
	return elements
}

func (n *simpleNodeBinder) checkParseScalar(expression Expression, diagnostics *[]Diagnostic) {
	if parseExpressionIsTabular(expression) {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
		return
	}
	typ := ExpressionResultType(expression)
	if typ != nil && !isErrorTypeSymbol(typ) && !sameType(typ, ScalarTypes.Unknown.Type()) && !IsScalar(typ) {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
	}
}

func parseExpressionIsTabular(expression SyntaxElement) bool {
	if expression == nil {
		return false
	}
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		if symbol := accessor.ReferencedSymbol(); symbol != nil && symbol.SymbolTabularity() == TabularityTabular {
			return true
		}
	}
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return parseExpressionIsTabular(expression.Expression())
	case *PipeExpression:
		return parseExpressionIsTabular(expression.Operator())
	}
	return false
}

func (n *simpleNodeBinder) declareParseColumn(name string, location SyntaxElement, declared map[string]struct{}, diagnostics *[]Diagnostic) bool {
	if _, exists := declared[name]; exists {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(name), location))
		return false
	}
	declared[name] = struct{}{}
	return true
}

func parsePrimitiveType(expression *PrimitiveTypeExpression) *TypeSymbol {
	if symbol, ok := expression.ReferencedSymbol().(*TypeSymbol); ok && !isErrorTypeSymbol(symbol) {
		return symbol
	}
	if symbol, ok := schemaPrimitiveTypeSymbol(expression).(*TypeSymbol); ok && !isErrorTypeSymbol(symbol) {
		return symbol
	}
	return ScalarTypes.Unknown.Type()
}

func (n *simpleNodeBinder) checkParsePatternPart(parts []SyntaxNodeElement, index int, part SyntaxNodeElement, diagnostics *[]Diagnostic) {
	location := SyntaxElement(part)
	switch part.(type) {
	case *LiteralExpression, *CompoundStringLiteralExpression:
		return
	case *StarExpression:
		if index+1 < len(parts) && !isParseStringLiteral(parts[index+1]) {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticParsePatternStringLiteralMustFollowStar(), location))
		}
		if index > 0 && isParseStringDeclaration(parts[index-1]) {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticParsePatternUsingStarAfterStringColumnIsAmbiguous(), location))
		}
	case *NameDeclaration, *NameAndTypeDeclaration:
		if index > 0 && !isParseStringLiteral(parts[index-1]) {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticParsePatternNameDoesNotFollowStringLiteral(), location))
		}
	default:
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidPatternPart(), location))
	}
}

func isParseStringLiteral(part SyntaxNodeElement) bool {
	switch part.(type) {
	case *LiteralExpression, *CompoundStringLiteralExpression:
		return true
	default:
		return false
	}
}

func isParseStringDeclaration(part SyntaxNodeElement) bool {
	switch part := part.(type) {
	case *NameDeclaration, *BracketedExpression:
		return true
	case *NameAndTypeDeclaration:
		primitive, ok := part.Type().(*PrimitiveTypeExpression)
		return ok && sameType(parsePrimitiveType(primitive), ScalarTypes.String.Type())
	default:
		return false
	}
}
