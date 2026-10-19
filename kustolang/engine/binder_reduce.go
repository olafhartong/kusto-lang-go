package kustolang

// Ported from Binder_NodeBinder.VisitReduceByOperator at the revision pinned
// in third_party/kusto/UPSTREAM.json. See third_party/kusto/LICENSE.
func (n *simpleNodeBinder) bindReduceBy(node *ReduceByOperator) *SemanticInfo {
	diagnostics := n.notFirstDiagnostics(node)
	diagnostics = append(diagnostics,
		queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.ReduceParameters)...)

	expression := node.Expression()
	if typ := ExpressionResultType(expression); topDiagnosticTypeKnown(typ) &&
		!sameType(typ, ScalarTypes.String.Type()) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), expression))
	}

	if with := node.With(); with != nil {
		diagnostics = append(diagnostics, queryOperatorParameterElementDiagnostics(
			n.binder, separatedNamedParameters(with.Parameters()), QueryOperatorParameters.ReduceWithParameters)...)
	}

	columns := make([]*ColumnSymbol, 0, 3)
	if reduceKindIsSource(node.Parameters()) && n.binder.rowScope != nil {
		columns = append(columns, n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
	}
	columns = append(columns,
		NewColumnSymbolWithSource("Pattern", ScalarTypes.String.Type(), expression),
		NewColumnSymbolWithSource("Count", ScalarTypes.Long.Type(), expression),
		NewColumnSymbolWithSource("Representative", ScalarTypes.String.Type(), expression))

	return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(NewTableSymbol("", columns...), diagnostics...))
}

func reduceKindIsSource(parameters *SyntaxListOf[*NamedParameter]) bool {
	if parameters == nil {
		return false
	}
	for _, parameter := range parameters.Elements() {
		if parameter == nil || NameDeclarationSimpleName(parameter.Name()) != "kind" {
			continue
		}
		if value := ExpressionLiteralValueInfo(parameter.Expression()); value != nil && value.ValueText == "source" {
			return true
		}
	}
	return false
}
