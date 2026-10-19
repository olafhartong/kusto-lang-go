package kustolang

// Ported from the PatternDeclaration path in
// src/Kusto.Language/Binder/Binder_NodeBinder.cs.

const patternDeclarationSymbolPayload = "kusto.pattern-declaration-symbol"

func BindPatternStatement(statement *PatternStatement) SemanticInfo {
	if statement == nil {
		return EmptySemanticInfo
	}
	name := ""
	if statement.Name() != nil {
		name = parameterNameFromDeclaration(statement.Name())
	}
	return BindPatternDeclaration(name, statement.Pattern())
}

func BindPatternDeclaration(name string, declaration *PatternDeclaration) SemanticInfo {
	symbol, diagnostics := buildPatternDeclaration(name, declaration, nil)
	return NewSemanticInfoForSymbol(symbol, &symbol.TypeSymbol, diagnostics...)
}

type patternDiagnosticLocator func(Diagnostic, SyntaxElement) Diagnostic

func buildPatternDeclaration(name string, declaration *PatternDeclaration, locate patternDiagnosticLocator) (*PatternSymbol, []Diagnostic) {
	if declaration == nil {
		return NewPatternSymbol(name, nil, nil, nil), []Diagnostic{}
	}

	diagnostics := []Diagnostic{}
	parameters := bindPatternParameters(declaration.Parameters())
	pathParameter := bindPatternPathParameter(declaration.PathParameter())
	signatures := []*PatternSignature{}

	if declaration.Patterns() != nil {
		for i := 0; i < declaration.Patterns().Count(); i++ {
			pattern := declaration.Patterns().At(i)
			if pattern == nil {
				continue
			}
			signatures = append(signatures, bindPatternMatch(pattern, declaration.Parameters(), declaration.PathParameter(), &diagnostics, locate))
		}
	}

	symbol := NewPatternSymbol(name, parameters, pathParameter, signatures)
	return symbol, diagnostics
}

func bindPatternParameters(parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration]) []*Parameter {
	if parameters == nil {
		return []*Parameter{}
	}
	out := make([]*Parameter, 0, parameters.Count())
	for _, element := range parameters.Elements() {
		if element == nil || element.Element() == nil {
			continue
		}
		out = append(out, bindPatternParameter(element.Element()))
	}
	return out
}

func bindPatternParameter(declaration *NameAndTypeDeclaration) *Parameter {
	if declaration == nil {
		return NewParameter("", ScalarTypes.Unknown.Type())
	}
	name := parameterNameFromDeclaration(declaration.Name())
	symbol := declaredSymbolFromTypeExpression(declaration.Type())
	return NewParameterSymbolType(name, symbol)
}

func bindPatternPathParameter(path *PatternPathParameter) *Parameter {
	if path == nil {
		return nil
	}
	return bindPatternParameter(path.Parameter())
}

func bindPatternMatch(pattern *PatternMatch, parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration], pathParameter *PatternPathParameter, diagnostics *[]Diagnostic, locate patternDiagnosticLocator) *PatternSignature {
	values := []string{}
	if pattern == nil {
		return NewPatternSignature(values, "", "")
	}

	parameterCount := 0
	if parameters != nil {
		parameterCount = parameters.Count()
	}

	if pattern.ParameterValues() != nil && pattern.ParameterValues().Expressions() != nil {
		expressions := pattern.ParameterValues().Expressions()
		if expressions.Count() != parameterCount {
			addPatternDiagnostic(diagnostics, DiagnosticValueCountMustEqualParameterCount(), pattern.ParameterValues(), locate)
		}
		for i := 0; i < expressions.Count(); i++ {
			value := expressions.At(i)
			if value == nil || value.Element() == nil {
				values = append(values, "")
				continue
			}
			text, ok := patternLiteralStringValue(value.Element(), value.Element(), diagnostics, locate)
			if !ok {
				text = ""
			}
			values = append(values, text)
		}
	} else if parameterCount != 0 {
		addPatternDiagnostic(diagnostics, DiagnosticValueCountMustEqualParameterCount(), pattern.ParameterValues(), locate)
	}

	pathValue := ""
	switch {
	case pattern.PathValue() == nil && pathParameter != nil:
		addPatternDiagnostic(diagnostics, DiagnosticPathValueExpected(), pattern.EqualToken(), locate)
	case pattern.PathValue() != nil && pathParameter == nil:
		addPatternDiagnostic(diagnostics, DiagnosticPathValueWithNoPathParameter(), pattern.PathValue(), locate)
	case pattern.PathValue() != nil && pathParameter != nil:
		if text, ok := patternLiteralStringValue(pattern.PathValue().Value(), pattern.PathValue().Value(), diagnostics, locate); ok {
			pathValue = text
		}
	}

	body := ""
	var declaration *FunctionBody
	if pattern.Body() != nil {
		body = pattern.Body().Text(IncludeTriviaAll)
		declaration = pattern.Body().CloneTyped(true)
	}
	signature := NewPatternSignature(values, pathValue, body, pattern.PathValue() != nil && pathParameter != nil)
	signature.declaration = declaration
	return signature
}

func addPatternDiagnostic(diagnostics *[]Diagnostic, diagnostic Diagnostic, location SyntaxElement, locate patternDiagnosticLocator) {
	if locate != nil && location != nil {
		diagnostic = locate(diagnostic, location)
	}
	*diagnostics = append(*diagnostics, diagnostic)
}

func patternLiteralStringValue(expr Expression, location SyntaxElement, diagnostics *[]Diagnostic, locate patternDiagnosticLocator) (string, bool) {
	if expr == nil || syntaxElementIsMissing(expr) {
		return "", false
	}
	if typ := ExpressionRawResultType(expr); isErrorTypeSymbol(typ) || sameType(typ, ScalarTypes.Unknown.Type()) {
		return "", false
	}
	switch value := expr.(type) {
	case *ParenthesizedExpression:
		return patternLiteralStringValue(value.Expression(), location, diagnostics, locate)
	case *LiteralExpression:
		if value.Kind() != SyntaxKindStringLiteralExpression {
			addPatternDiagnostic(diagnostics, DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), location, locate)
			return "", false
		}
		if value.Token() == nil {
			return "", true
		}
		return value.Token().ValueText(), true
	case *CompoundStringLiteralExpression:
		if value.Tokens() == nil {
			return "", true
		}
		text := ""
		for i := 0; i < value.Tokens().Count(); i++ {
			token := value.Tokens().At(i)
			if token != nil {
				text += token.ValueText()
			}
		}
		return text, true
	default:
		addPatternDiagnostic(diagnostics, DiagnosticExpressionMustBeLiteral(), location, locate)
		return "", false
	}
}

func (n *simpleNodeBinder) bindPatternDeclaration(node *PatternDeclaration) *SemanticInfo {
	name := ""
	if node != nil && node.Parent() != nil {
		if statement, ok := WrapGeneratedSyntaxElement(node.Parent()).(*PatternStatement); ok {
			name = NameDeclarationSimpleName(statement.Name())
		}
	}
	symbol, diagnostics := buildPatternDeclaration(name, node, n.binder.diagnosticAtSyntaxElement)
	if node != nil && node.SyntaxNode != nil {
		node.SyntaxNode.setPayload(patternDeclarationSymbolPayload, symbol)
	}
	return semanticInfoPtr(NewSemanticInfoForType(&symbol.TypeSymbol, diagnostics...))
}
