package kustolang

import "strings"

// Ported from Binder_NodeBinder.VisitRenderOperator.
func (n *simpleNodeBinder) bindRender(node *RenderOperator) *SemanticInfo {
	if info, ok := node.GetSemanticInfo(); ok {
		return &info
	}
	diagnostics := n.notFirstDiagnostics(node)

	chartType := node.ChartType()
	if chartType != nil && chartType.Width() != 0 && !renderChartTypeIsKnown(chartType.ValueText()) {
		diagnostics = append(diagnostics,
			n.binder.diagnosticAtSyntaxElement(DiagnosticTermsExpected(KustoFacts.ChartTypes...), chartType))
	}

	diagnostics = append(diagnostics,
		queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.RenderParameters)...)
	if with := node.WithClause(); with != nil {
		diagnostics = append(diagnostics,
			queryOperatorParameterElementDiagnostics(n.binder, separatedNamedParameters(with.Properties()), QueryOperatorParameters.RenderWithProperties)...)
	}

	return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(n.rowScopeResult(), diagnostics...))
}

func renderChartTypeIsKnown(value string) bool {
	for _, chartType := range KustoFacts.ChartTypes {
		if strings.EqualFold(value, chartType) {
			return true
		}
	}
	return false
}
