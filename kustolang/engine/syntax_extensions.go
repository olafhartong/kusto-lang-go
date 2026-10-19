package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxExtensions.cs and the
// Name convenience members in SyntaxNode.cs.

// NewNameDeclarationFromToken is the Go spelling of the NameDeclaration
// constructor overload that accepts a SyntaxToken.
func NewNameDeclarationFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) (*NameDeclaration, error) {
	name, err := NewTokenNameTyped(nameToken)
	if err != nil {
		return nil, err
	}
	return NewNameDeclarationTyped(name, diagnostics...)
}

func MustNewNameDeclarationFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) *NameDeclaration {
	declaration, err := NewNameDeclarationFromToken(nameToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return declaration
}

// NewNameReferenceFromName is the Go spelling of the NameReference
// constructor overload whose SymbolMatch argument defaults to Default.
func NewNameReferenceFromName(name Name, diagnostics ...Diagnostic) (*NameReference, error) {
	return NewNameReferenceTyped(name, SymbolMatchDefault, diagnostics...)
}

func MustNewNameReferenceFromName(name Name, diagnostics ...Diagnostic) *NameReference {
	reference, err := NewNameReferenceFromName(name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return reference
}

// NewNameReferenceFromToken is the Go spelling of the NameReference
// constructor overload that accepts a SyntaxToken.
func NewNameReferenceFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) (*NameReference, error) {
	name, err := NewTokenNameTyped(nameToken)
	if err != nil {
		return nil, err
	}
	return NewNameReferenceFromName(name, diagnostics...)
}

func MustNewNameReferenceFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) *NameReference {
	reference, err := NewNameReferenceFromToken(nameToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return reference
}

func NameSimpleName(name Name) string {
	switch n := name.(type) {
	case *TokenName:
		if n.Name() != nil {
			return n.Name().ValueText()
		}
	case *BracketedName:
		if value, ok := ExpressionLiteralValue(n.Name()).(string); ok {
			return value
		}
	case *BracedName:
		if n.Name() != nil {
			return n.Name().ValueText()
		}
	case *WildcardedName:
		if n.Pattern() != nil {
			return n.Pattern().ValueText()
		}
	case *BracketedWildcardedName:
		if n.Pattern() != nil {
			return n.Pattern().ValueText()
		}
	}
	return ""
}

func NameDeclarationSimpleName(declaration *NameDeclaration) string {
	if declaration == nil {
		return ""
	}
	return NameSimpleName(declaration.Name())
}

func (n *NameDeclaration) SimpleName() string {
	return NameDeclarationSimpleName(n)
}

func NameReferenceSimpleName(reference *NameReference) string {
	if reference == nil {
		return ""
	}
	return NameSimpleName(reference.Name())
}

func (n *NameReference) SimpleName() string {
	return NameReferenceSimpleName(n)
}

func GetParameter(list *SyntaxListOf[*NamedParameter], parameter *QueryOperatorParameter) *NamedParameter {
	if list == nil || parameter == nil {
		return nil
	}
	for _, named := range list.Elements() {
		if named == nil {
			continue
		}
		name := NameDeclarationSimpleName(named.Name())
		if name == parameter.SymbolName() || syntaxStringSliceContains(parameter.Aliases(), name) {
			return named
		}
	}
	return nil
}

func GetParameterLiteralValue[T any](list *SyntaxListOf[*NamedParameter], declaration *QueryOperatorParameter, defaultValue T) T {
	parameter := GetParameter(list, declaration)
	if parameter == nil {
		return defaultValue
	}
	if literal, ok := parameter.Expression().(*LiteralExpression); ok {
		if value, ok := ExpressionConstantValue(literal).(T); ok {
			return value
		}
	}
	return defaultValue
}

func GetParameterNameValue(list *SyntaxListOf[*NamedParameter], declaration *QueryOperatorParameter) string {
	parameter := GetParameter(list, declaration)
	if parameter == nil {
		return ""
	}
	switch expression := parameter.Expression().(type) {
	case *NameDeclaration:
		return NameDeclarationSimpleName(expression)
	case *NameReference:
		return NameReferenceSimpleName(expression)
	case *LiteralExpression:
		if expression.Kind() == SyntaxKindStringLiteralExpression || expression.Kind() == SyntaxKindTokenLiteralExpression {
			if info := ExpressionLiteralValueInfo(expression); info != nil {
				return info.ValueText
			}
		}
	case *CompoundStringLiteralExpression:
		if info := ExpressionLiteralValueInfo(expression); info != nil {
			return info.ValueText
		}
	}
	return ""
}

func syntaxStringSliceContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
