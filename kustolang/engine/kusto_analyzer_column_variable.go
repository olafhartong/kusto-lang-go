package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/ColumnHasSameNameAsVariableAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var columnHasSameNameAsVariableDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeColumnHasSameNameAsVariable,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"A variable or parameter is ignored in favor of a column with the same name.",
	"",
)

func NewColumnHasSameNameAsVariableAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "ColumnHasSameNameAsVariableAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{columnHasSameNameAsVariableDiagnostic}
		},
		Analyze: analyzeColumnHasSameNameAsVariable,
	})
}

func analyzeColumnHasSameNameAsVariable(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	localNames := map[string]struct{}{}
	type scopedParameter struct {
		name  string
		owner SyntaxElement
	}
	var scopedParameters []scopedParameter
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		switch node := WrapGeneratedSyntaxElement(element).(type) {
		case *NameDeclaration:
			symbol := node.ReferencedSymbol()
			switch symbol.(type) {
			case *VariableSymbol, *ParameterSymbol:
				if IsScalar(symbol) {
					if name := NameDeclarationSimpleName(node); name != "" {
						localNames[name] = struct{}{}
						if _, ok := symbol.(*ParameterSymbol); ok {
							var owner SyntaxElement
							isFunctionParameter := false
							for i := len(ancestors) - 1; i >= 0; i-- {
								switch WrapGeneratedSyntaxElement(ancestors[i]).(type) {
								case *FunctionParameter:
									isFunctionParameter = true
								case *FunctionDeclaration:
									owner = ancestors[i]
								}
								if owner != nil && isFunctionParameter {
									break
								}
							}
							if owner != nil && isFunctionParameter {
								scopedParameters = append(scopedParameters, scopedParameter{name: name, owner: owner})
							}
						}
					}
				}
			}
		case *NameReference:
			if columnHasSameNameReferenceIsPathSelector(node, ancestors) {
				return
			}
			column, ok := node.ReferencedSymbol().(*ColumnSymbol)
			if !ok {
				return
			}
			if _, ok := localNames[column.SymbolName()]; !ok {
				return
			}
			visibleLocal := code.getAttachedSpeculativeReferencedSymbol(elementTextStart(node, triviaStart), column.SymbolName(), SymbolMatchLocal) != nil
			if !visibleLocal {
				for _, parameter := range scopedParameters {
					if parameter.name != column.SymbolName() {
						continue
					}
					for _, ancestor := range ancestors {
						if ancestor == parameter.owner {
							visibleLocal = true
							break
						}
					}
					if visibleLocal {
						break
					}
				}
			}
			if visibleLocal {
				diagnostics = append(diagnostics, columnHasSameNameAsVariableDiagnostic.
					WithLocation(elementTextStart(node, triviaStart), node.Width()))
			}
		}
	})
	return diagnostics
}

func columnHasSameNameReferenceIsPathSelector(name *NameReference, ancestors []SyntaxElement) bool {
	if name == nil || len(ancestors) == 0 {
		return false
	}
	parent, ok := WrapGeneratedSyntaxElement(ancestors[len(ancestors)-1]).(*PathExpression)
	if !ok {
		return false
	}
	selector, ok := parent.Selector().(*NameReference)
	return ok && selector.SyntaxNode == name.SyntaxNode
}
