package kustolang

// Ported from the query-parameter and set-option statement binding in
// Binder_TreeBinder.cs and Binder_NodeBinder.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (t *treeBinder) visitQueryParametersStatement(node *QueryParametersStatement) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	if node.Parameters() != nil {
		for _, separated := range node.Parameters().Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			parameter := separated.Element()
			t.visitElement(parameter)
			if !queryParameterChildrenBound(parameter) {
				tx.rollback()
				return
			}
			if separated.Separator() != nil {
				t.visitElement(separated.Separator())
			}
		}
		variables := make([]*VariableSymbol, 0, node.Parameters().Count())
		for _, separated := range node.Parameters().Elements() {
			if separated != nil {
				if variable := t.bindQueryParameterDeclaration(separated.Element()); variable != nil {
					variables = append(variables, variable)
				}
			}
		}
		t.bindNode(node.SyntaxNode)
		if !node.IsBound() {
			tx.rollback()
			return
		}
		if t.binder.LocalScope == nil {
			t.binder.LocalScope = NewLocalScope(nil)
		}
		for _, variable := range variables {
			t.binder.LocalScope.AddSymbol(variable)
		}
		return
	}
	t.bindNode(node.SyntaxNode)
}

func queryParameterChildrenBound(parameter *FunctionParameter) bool {
	if parameter == nil || parameter.NameAndType() == nil {
		return false
	}
	if typ := simpleBinderSyntaxNode(parameter.NameAndType().Type()); typ == nil || !typ.IsBound() {
		return false
	}
	if parameter.DefaultValue() != nil && parameter.DefaultValue().Value() != nil {
		value := simpleBinderSyntaxNode(parameter.DefaultValue().Value())
		return value != nil && value.IsBound()
	}
	return true
}

func (t *treeBinder) bindQueryParameterDeclaration(parameter *FunctionParameter) *VariableSymbol {
	if t == nil || t.binder == nil || parameter == nil || parameter.NameAndType() == nil {
		return nil
	}
	declaration := parameter.NameAndType()
	name := NameDeclarationSimpleName(declaration.Name())
	if name == "" {
		return nil
	}
	typ := declaredSymbolFromTypeExpression(declaration.Type())
	if typ == nil {
		typ = Error
	}
	var source Expression
	if parameter.DefaultValue() != nil {
		source = parameter.DefaultValue().Value()
	}
	variable := NewVariableSymbol(name, typ, source != nil && ExpressionIsConstant(source), ExpressionConstantValueInfo(source), source)
	declaration.Name().SetSemanticInfo(NewSemanticInfoForSymbol(variable, returnTypeFromSymbol(typ)))
	return variable
}

func (t *treeBinder) visitSetOptionStatement(node *SetOptionStatement) {
	if node == nil || node.IsBound() {
		return
	}
	tx := newUnionBindingTransaction(node)
	t.visitElement(node.Name())
	if node.ValueClause() != nil {
		t.visitElement(node.ValueClause())
		if value := node.ValueClause().Expression(); value != nil {
			valueNode := simpleBinderSyntaxNode(value)
			if valueNode == nil || !valueNode.IsBound() {
				tx.rollback()
				return
			}
		}
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

func (t *treeBinder) visitLetStatement(node *LetStatement) {
	if node == nil || node.IsBound() {
		return
	}
	if node.LetKeyword() != nil {
		t.visitElement(node.LetKeyword())
	}
	if node.Name() != nil {
		t.visitElement(node.Name())
	}
	if node.Equal() != nil {
		t.visitElement(node.Equal())
	}
	if declaration, ok := node.Expression().(*FunctionDeclaration); ok {
		t.visitFunctionDeclarationScope(declaration, NameDeclarationSimpleName(node.Name()))
	} else {
		t.visitElement(node.Expression())
	}

	name := NameDeclarationSimpleName(node.Name())
	diagnostics := make([]Diagnostic, 0, 2)
	if t.binder.LocalScope != nil && t.binder.LocalScope.HasSymbol(name) {
		diagnostics = append(diagnostics, t.binder.diagnosticAtSyntaxElement(DiagnosticVariableAlreadyDeclared(name), node.Name()))
	}
	if _, ok := letExpressionResultSymbol(node.Expression()).(*TupleSymbol); ok {
		diagnostics = append(diagnostics, t.binder.diagnosticAtSyntaxElement(DiagnosticMultiValuedExpressionCannotBeAssignedToVariable(), node.Expression()))
	}
	node.SetSemanticInfo(NewSemanticInfoWithDiagnostics(diagnostics...))

	local := letExpressionResultSymbol(node.Expression())
	if function, ok := local.(*FunctionSymbol); ok {
		local = function
	} else if group, ok := local.(*EntityGroupSymbol); ok {
		local = group
	} else {
		local = NewVariableSymbol(name, local, ExpressionIsConstant(node.Expression()), ExpressionConstantValueInfo(node.Expression()), node.Expression())
	}
	if node.Name() != nil {
		node.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(local))
	}
	if t.binder.LocalScope == nil {
		t.binder.LocalScope = NewLocalScope(nil)
	}
	t.binder.LocalScope.AddSymbol(local)
}

func letExpressionResultSymbol(expression Expression) Symbol {
	if expression == nil {
		return Error
	}
	switch expression := expression.(type) {
	case *FunctionDeclaration:
		if function, ok := expression.ReferencedSymbol().(*FunctionSymbol); ok {
			return function
		}
	case *FunctionCallExpression:
		if result, found := expression.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok && symbol != nil {
				return symbol
			}
		}
	}
	if graph := expressionGraphResult(expression); graph != nil {
		return graph
	}
	if table := expressionTableResult(expression); table != nil {
		return table
	}
	if referenced := expressionReferencedSymbol(expression); referenced != nil {
		if group, ok := GetResultType(referenced).(*EntityGroupSymbol); ok {
			return group
		}
	}
	return resultTypeOrError(expression)
}

func (t *treeBinder) visitFunctionDeclarationScope(node *FunctionDeclaration, name string) {
	if node == nil {
		return
	}
	oldScope := t.binder.LocalScope
	oldSuffix := t.nodeBinder.defaultColumnNameSuffix
	staticScope := oldScope.Copy()
	t.binder.LocalScope = NewLocalScope(oldScope)
	t.nodeBinder.defaultColumnNameSuffix = 1
	defer func() {
		t.binder.LocalScope = oldScope
		t.nodeBinder.defaultColumnNameSuffix = oldSuffix
	}()
	if node.ViewKeyword() != nil {
		t.visitElement(node.ViewKeyword())
	}
	if parameters := node.Parameters(); parameters != nil {
		t.visitElement(parameters)
		for _, separated := range parameters.Parameters().Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().NameAndType() == nil {
				continue
			}
			declaration := separated.Element().NameAndType()
			parameterName := NameDeclarationSimpleName(declaration.Name())
			if parameterName == "" {
				continue
			}
			typ := declaredSymbolFromTypeExpression(declaration.Type())
			if syntaxElementIsMissing(declaration.Type()) {
				typ = ScalarTypes.Unknown.Type()
			}
			parameter := NewParameterSymbol(parameterName, typ, "")
			declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(parameter))
			t.binder.LocalScope.AddSymbol(parameter)
		}
	}
	if node.Body() != nil {
		t.visitElement(node.Body())
	}

	diagnostics := make([]Diagnostic, 0, 1)
	parameters := NewParametersFromFunctionParameters(node.Parameters())
	canBeTabular := true
	for i, parameter := range parameters {
		if parameter == nil {
			continue
		}
		isTabular := len(parameter.DeclaredSymbols) == 1 && IsTabular(parameter.DeclaredSymbols[0])
		if isTabular && !canBeTabular && node.Parameters() != nil && i < node.Parameters().Parameters().Count() {
			diagnostics = append(diagnostics, t.binder.diagnosticAtSyntaxElement(DiagnosticTabularParametersMustBeDeclaredFirst(), node.Parameters().Parameters().At(i).Element()))
		}
		canBeTabular = isTabular
	}
	function := NewFunctionSymbol(name, NewDeclarationSignature(node.Body(), parameters...)).WithIsView(node.ViewKeyword() != nil)
	node.SetSemanticInfo(NewSemanticInfoForSymbol(function, &function.TypeSymbol, diagnostics...))
	if t.binder.functionStaticScopes == nil {
		t.binder.functionStaticScopes = map[*Signature]*LocalScope{}
	}
	for _, signature := range function.Signatures() {
		t.binder.functionStaticScopes[signature] = staticScope
	}
}
