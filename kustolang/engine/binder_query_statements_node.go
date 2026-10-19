package kustolang

// Ported from Binder_NodeBinder.cs VisitFunctionParameter and
// VisitSetOptionStatement. See third_party/kusto/LICENSE.

func (n *simpleNodeBinder) bindFunctionParameter(node *FunctionParameter) *SemanticInfo {
	if node == nil || node.NameAndType() == nil {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	if declaration := node.NameAndType(); node.Width() == 0 && declaration.Type() != nil && declaration.Type().Width() == 0 {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics(
			n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidTypeName(""), declaration.Type())))
	}
	if node.DefaultValue() == nil {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	value := node.DefaultValue().Value()
	target := returnTypeFromSymbol(declaredSymbolFromTypeExpression(node.NameAndType().Type()))
	if value == nil || target == nil || isErrorTypeSymbol(target) {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	if !ExpressionIsLiteral(value) || value.Kind() == SyntaxKindTokenLiteralExpression {
		if !isErrorTypeSymbol(resultTypeOrError(value)) {
			return semanticInfoPtr(NewSemanticInfoWithDiagnostics(
				n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeLiteralScalarValue(), value)))
		}
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	valueType := resultTypeOrError(value)
	if !isErrorTypeSymbol(valueType) && !sameType(valueType, ScalarTypes.Unknown.Type()) &&
		!SymbolAssignableTo(valueType, target, ConversionCompatible) {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics(
			n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(target.SymbolName()), value)))
	}
	return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
}

func (n *simpleNodeBinder) bindSetOptionStatement(node *SetOptionStatement) *SemanticInfo {
	if node == nil {
		return nil
	}
	diagnostics := make([]Diagnostic, 0, 2)
	if node.Name() != nil {
		if _, ok := node.Name().Name().(*TokenName); !ok {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticIdentifierNameOnly(), node.Name()))
		}
	}

	var option *OptionSymbol
	if n.binder != nil && n.binder.Globals != nil && node.Name() != nil {
		option = n.binder.Globals.GetOption(NameDeclarationSimpleName(node.Name()))
		if option != nil {
			node.Name().SetSemanticInfo(NewSemanticInfoForSymbol(option, nil))
		}
	}

	if clause := node.ValueClause(); clause != nil && clause.Expression() != nil {
		value := clause.Expression()
		allowedShape := ExpressionIsLiteral(value)
		if _, ok := value.(*NameReference); ok {
			allowedShape = true
		}
		valueType := resultTypeOrError(value)
		if !allowedShape && !isErrorTypeSymbol(valueType) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeConstantOrIdentifier(), value))
		} else if allowedShape && option != nil && len(option.Types()) > 0 &&
			!isErrorTypeSymbol(valueType) && !sameType(valueType, ScalarTypes.Unknown.Type()) {
			types := option.Types()
			targets := make([]Symbol, 0, len(types))
			names := make([]string, 0, len(types))
			for _, typ := range types {
				targets = append(targets, typ.Type())
				names = append(names, typ.SymbolName())
			}
			if !SymbolAssignableToAny(valueType, targets, ConversionCompatible) {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(names...), value))
			}
		}
	}
	if len(diagnostics) == 0 {
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	}
	return semanticInfoPtr(NewSemanticInfoWithDiagnostics(diagnostics...))
}
