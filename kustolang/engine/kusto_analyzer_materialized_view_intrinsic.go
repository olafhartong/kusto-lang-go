package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/PreferUsingMaterializedViewIntrinsicAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var preferUsingMaterializedViewIntrinsicDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Prefer using the `materialized_view()` function where appropriate",
	"Consider using the `materialized_view()` function to query the materialized part only\nSee: https://aka.ms/mvqueries",
)

var materializedViewDownsamplingFunctionNames = map[string]struct{}{
	"bin":          {},
	"bin_at":       {},
	"startofday":   {},
	"startofmonth": {},
	"startofweek":  {},
	"startofyear":  {},
}

func NewPreferUsingMaterializedViewIntrinsicAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "PreferUsingMaterializedViewIntrinsicAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{preferUsingMaterializedViewIntrinsicDiagnostic}
		},
		Analyze: analyzePreferUsingMaterializedViewIntrinsic,
	})
}

func analyzePreferUsingMaterializedViewIntrinsic(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil || code.Kind != CodeKindQuery {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		nameReference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if !ok {
			return
		}
		materializedView, ok := nameReference.ReferencedSymbol().(*MaterializedViewSymbol)
		if !ok {
			return
		}
		if preferMaterializedViewIntrinsicShouldReport(materializedView, code) {
			diagnostics = append(diagnostics, preferUsingMaterializedViewIntrinsicDiagnostic.
				WithLocation(elementTextStart(nameReference, triviaStart), nameReference.Width()))
		}
	})
	return diagnostics
}

func preferMaterializedViewIntrinsicShouldReport(materializedView *MaterializedViewSymbol, context *KustoCode) bool {
	if materializedView == nil {
		return false
	}
	switch materializedView.MaterializedViewKind() {
	case MaterializedViewKindOther:
		return false
	case MaterializedViewKindDownsampling:
		return true
	default:
		if preferMaterializedViewQueryIsDownsampling(materializedView.MaterializedViewQuery(), context.Globals) {
			materializedView.materializedViewKind = MaterializedViewKindDownsampling
			return true
		}
		materializedView.materializedViewKind = MaterializedViewKindOther
		return false
	}
}

func preferMaterializedViewQueryIsDownsampling(query string, globals *GlobalState) bool {
	if query == "" {
		return false
	}
	code := ParseWithGlobals(query, globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		return false
	}
	summarize := preferMaterializedViewSingleSummarizeOperator(preferMaterializedViewLastStatement(code))
	return summarize != nil && preferMaterializedViewSummarizeHasDownsamplingByKey(summarize)
}

func preferMaterializedViewLastStatement(code *KustoCode) Statement {
	if code == nil || code.Query == nil || code.Query.SyntaxNode == nil || code.Query.ChildCount() <= 1 {
		return nil
	}
	list, ok := code.Query.GetChild(1).(*SyntaxList)
	if !ok || list.Count() == 0 {
		return nil
	}
	last := AsSeparatedSyntaxListOf[Statement](list).At(list.Count() - 1)
	if last == nil {
		return nil
	}
	return last.Element()
}

func preferMaterializedViewSingleSummarizeOperator(statement Statement) *SummarizeOperator {
	if statement == nil {
		return nil
	}
	var found *SummarizeOperator
	count := 0
	WalkSyntax(statement, func(element SyntaxElement) {
		if summarize, ok := WrapGeneratedSyntaxElement(element).(*SummarizeOperator); ok {
			found = summarize
			count++
		}
	})
	if count != 1 {
		return nil
	}
	return found
}

func preferMaterializedViewSummarizeHasDownsamplingByKey(summarize *SummarizeOperator) bool {
	if summarize == nil || summarize.ByClause() == nil || summarize.ByClause().Expressions() == nil {
		return false
	}
	expressions := summarize.ByClause().Expressions()
	for i := 0; i < expressions.Count(); i++ {
		element := expressions.At(i)
		if element != nil && preferMaterializedViewExpressionIsDownsamplingKey(element.Element()) {
			return true
		}
	}
	return false
}

func preferMaterializedViewExpressionIsDownsamplingKey(expr Expression) bool {
	switch expression := expr.(type) {
	case *FunctionCallExpression:
		return preferMaterializedViewIsDownsamplingFunctionCall(expression)
	case *SimpleNamedExpression:
		if call, ok := expression.Expression().(*FunctionCallExpression); ok {
			return preferMaterializedViewIsDownsamplingFunctionCall(call)
		}
	}
	return false
}

func preferMaterializedViewIsDownsamplingFunctionCall(call *FunctionCallExpression) bool {
	if call == nil {
		return false
	}
	_, ok := materializedViewDownsamplingFunctionNames[NameReferenceSimpleName(call.Name())]
	return ok
}
