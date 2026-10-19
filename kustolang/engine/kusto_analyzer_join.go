package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidJoinWithoutKindAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidJoinWithoutKindDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidJoinWithoutKind,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using joins without specifying the join kind.",
	"Avoid using joins without specifying the join kind. The default join behavior may be unexpected.",
)

func NewAvoidJoinWithoutKindAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidJoinWithoutKindAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidJoinWithoutKindDiagnostic}
		},
		Analyze: analyzeAvoidJoinWithoutKind,
	})
}

func analyzeAvoidJoinWithoutKind(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		join, ok := WrapGeneratedSyntaxElement(element).(*JoinOperator)
		if !ok || join.ConditionClause() == nil || joinConditionClauseIsMissing(join.ConditionClause()) || joinHasKindParameter(join) {
			return
		}
		keyword := join.JoinKeyword()
		if keyword == nil {
			return
		}
		diagnostics = append(diagnostics, avoidJoinWithoutKindDiagnostic.WithByteLocation(elementTextStart(keyword, triviaStart), keyword.Width()))
	})
	return diagnostics
}

func joinConditionClauseIsMissing(clause JoinConditionClause) bool {
	if clause == nil {
		return true
	}
	for _, diagnostic := range clause.ContainedDiagnostics() {
		if diagnostic.Code == "KS006" && diagnostic.Message == "Missing join on condition clause" {
			return true
		}
	}
	return false
}

func joinHasKindParameter(join *JoinOperator) bool {
	if join == nil || join.Parameters() == nil {
		return false
	}
	for _, parameter := range join.Parameters().Elements() {
		if NameDeclarationSimpleName(parameter.Name()) == "kind" {
			return true
		}
	}
	return false
}
