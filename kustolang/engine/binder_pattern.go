package kustolang

// Ported from Binder_TreeBinder.cs VisitPatternStatement and
// VisitPatternDeclaration.

func (t *treeBinder) visitPatternStatement(node *PatternStatement) {
	if node == nil {
		return
	}
	var pattern *PatternSymbol
	if declaration := node.Pattern(); declaration != nil {
		t.visitPatternDeclaration(declaration)
		value, _ := declaration.SyntaxNode.Payload(patternDeclarationSymbolPayload)
		pattern, _ = value.(*PatternSymbol)
	}
	if pattern == nil {
		pattern = NewNamedPatternSymbol(NameDeclarationSimpleName(node.Name()))
	}
	name := NameDeclarationSimpleName(node.Name())
	local := NewVariableSymbol(name, pattern, false, nil, nil)
	if node.Name() != nil {
		node.Name().SetSemanticInfo(NewSemanticInfoForSymbol(local, nil))
	}
	if t.binder.LocalScope == nil {
		t.binder.LocalScope = NewLocalScope(nil)
	}
	t.binder.LocalScope.AddSymbol(local)
}

func (t *treeBinder) visitPatternDeclaration(node *PatternDeclaration) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	parameters := t.bindPatternParameterDeclarations(node.Parameters())
	if node.PathParameter() != nil {
		parameters = append(parameters, t.bindPatternParameterDeclaration(node.PathParameter().Parameter()))
	}

	oldLocal := t.binder.LocalScope
	t.binder.LocalScope = NewLocalScope(oldLocal)
	defer func() { t.binder.LocalScope = oldLocal }()
	for _, parameter := range parameters {
		if parameter != nil {
			t.binder.LocalScope.AddSymbol(parameter)
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

func (t *treeBinder) bindPatternParameterDeclarations(parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration]) []*ParameterSymbol {
	if parameters == nil {
		return nil
	}
	symbols := make([]*ParameterSymbol, 0, parameters.Count())
	for _, separated := range parameters.Elements() {
		if separated != nil {
			symbols = append(symbols, t.bindPatternParameterDeclaration(separated.Element()))
		}
	}
	return symbols
}

func (t *treeBinder) bindPatternParameterDeclaration(declaration *NameAndTypeDeclaration) *ParameterSymbol {
	if declaration == nil {
		return nil
	}
	t.visitElement(declaration.Type())
	typ := declaredSymbolFromTypeExpression(declaration.Type())
	if typ == nil {
		typ = Error
	}
	symbol := NewParameterSymbol(NameDeclarationSimpleName(declaration.Name()), typ, "")
	if declaration.Name() != nil {
		declaration.Name().SetSemanticInfo(NewSemanticInfoForSymbol(symbol, returnTypeFromSymbol(typ)))
	}
	return symbol
}
