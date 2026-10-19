package kustolang

func (t *treeBinder) visitEvaluate(node *EvaluateOperator) {
	old := t.binder.scopeKind
	t.binder.scopeKind = ScopeKindPlugIn
	defer func() { t.binder.scopeKind = old }()
	for _, p := range node.Parameters().Elements() {
		if p != nil {
			t.visitElement(p.Expression())
			t.visitElement(p)
		}
	}
	t.visitElement(node.FunctionCall())
	if clause := node.Schema(); clause != nil {
		t.visitElement(clause)
	}
	t.bindNode(node.SyntaxNode)
}

func (n *simpleNodeBinder) bindEvaluate(node *EvaluateOperator) *SemanticInfo {
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.EvaluateParameters)
	result := resultTypeOrError(node.FunctionCall())
	if clause := node.Schema(); clause != nil && clause.Schema() != nil {
		schema := clause.Schema()
		columns := []*ColumnSymbol{}
		explicit := map[string]bool{}
		for _, separated := range schema.Columns().Elements() {
			if separated != nil {
				explicit[NameDeclarationSimpleName(separated.Element().Name())] = true
			}
		}
		if schema.AsteriskToken() != nil && schema.AsteriskToken().Width() > 0 {
			resultSymbol, _ := node.FunctionCall().SyntaxNode.Payload(functionCallResultSymbolPayload)
			if table, ok := resultSymbol.(*TableSymbol); ok {
				for _, col := range table.Columns() {
					if !explicit[col.SymbolName()] {
						columns = append(columns, col)
					}
				}
			}
		}
		declared := map[string]bool{}
		for _, separated := range schema.Columns().Elements() {
			if separated == nil {
				continue
			}
			declaration := separated.Element()
			name := NameDeclarationSimpleName(declaration.Name())
			if declared[name] {
				continue
			}
			declared[name] = true
			column := n.evaluateDeclaredColumn(declaration, &diagnostics)
			columns = append(columns, column)
			declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		}
		table := NewTableSymbol("", columns...)
		return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForType(result, diagnostics...))
}

func (n *simpleNodeBinder) evaluateDeclaredColumn(declaration *NameAndTypeDeclaration, diagnostics *[]Diagnostic) *ColumnSymbol {
	if declaration != nil {
		if primitive, ok := declaration.Type().(*PrimitiveTypeExpression); ok && primitive.Type() != nil && primitive.Type().Width() == 0 {
			typ := ScalarTypes.Unknown.Type()
			return NewColumnSymbolWithSource(NameDeclarationSimpleName(declaration.Name()), typ, declaration.Name())
		}
	}
	return scanDeclaredColumn(declaration)
}
