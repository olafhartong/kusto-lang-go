package kustolang

// GetStandaloneQueryOperatorSemanticInfo returns the semantic result owned by
// a query operator that the parser recovered as a complete expression
// statement. The upstream tree binder visits such an operator with no row
// scope, so its existing semantic info is the authoritative result.
func GetStandaloneQueryOperatorSemanticInfo(element SyntaxElement) (SemanticInfo, []Diagnostic, bool) {
	operator, ok := standaloneQueryOperator(element)
	if !ok {
		return SemanticInfo{}, []Diagnostic{}, false
	}
	return standaloneQueryOperatorSemanticInfo(operator)
}

// BindStandaloneQueryOperatorSemanticInfo binds the owning query block when
// needed, then returns the standalone operator's semantic result. Binding the
// complete block preserves the parser profile, absolute locations, and lexical
// declarations that the upstream tree binder would see.
func BindStandaloneQueryOperatorSemanticInfo(element SyntaxElement, globals *GlobalState, cancellationTokens ...CancellationToken) (SemanticInfo, []Diagnostic, bool) {
	operator, ok := standaloneQueryOperator(element)
	if !ok {
		return SemanticInfo{}, []Diagnostic{}, false
	}
	if _, bound := operator.GetSemanticInfo(); !bound {
		root := SyntaxElement(operator)
		for {
			parent := root.Parent()
			if isNilSyntaxElement(parent) {
				break
			}
			root = parent
		}
		if root.Kind() != SyntaxKindQueryBlock {
			return SemanticInfo{}, []Diagnostic{}, false
		}
		cancellationToken := binderAPICancellationToken(cancellationTokens)
		if !NewBinder(globals).TryBindWithCancellation(root, cancellationToken) {
			return SemanticInfo{}, []Diagnostic{}, false
		}
		throwBinderAPICancellation(cancellationTokens)
	}
	return standaloneQueryOperatorSemanticInfo(operator)
}

func standaloneQueryOperatorSemanticInfo(operator *SyntaxNode) (SemanticInfo, []Diagnostic, bool) {
	info, ok := operator.GetSemanticInfo()
	if !ok {
		return SemanticInfo{}, []Diagnostic{}, false
	}
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(operator, DiagnosticsIncludeSemantic)
	return info, diagnostics, true
}

func standaloneQueryOperator(element SyntaxElement) (*SyntaxNode, bool) {
	if element == nil {
		return nil, false
	}

	wrapped := WrapGeneratedSyntaxElement(element)
	if statement, ok := wrapped.(*ExpressionStatement); ok {
		element = statement.Expression()
		wrapped = WrapGeneratedSyntaxElement(element)
	}
	operator, ok := wrapped.(QueryOperator)
	if !ok {
		return nil, false
	}
	operatorNode := standaloneSyntaxNode(operator)
	if operatorNode == nil {
		return nil, false
	}

	parent, ok := WrapGeneratedSyntaxElement(operator.Parent()).(*ExpressionStatement)
	if !ok || standaloneSyntaxNode(parent.Expression()) != operatorNode {
		return nil, false
	}
	return operatorNode, true
}

func standaloneSyntaxNode(element SyntaxElement) *SyntaxNode {
	if fragment, ok := element.(syntaxNodeFragment); ok {
		return fragment.syntaxNodeRoot()
	}
	return simpleBinderSyntaxNode(element)
}
