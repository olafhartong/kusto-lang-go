package kustolang

import "testing"

func TestParseCurrentTerminalJoinRecoveryReferenceParity(t *testing.T) {
	const queryText = `datatable(k:long,a:string)[1,'x'] | join`

	query := ParseQuery(queryText)
	diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeSyntactic)
	want := []Diagnostic{
		DiagnosticMissingJoinOnClause().WithLocation(36, 4),
		DiagnosticMissingExpression().WithLocation(40, 0),
	}
	if !diagnosticsEqualLocationAndMessage(diagnostics, want) {
		t.Fatalf("diagnostics = %#v, want %#v", diagnostics, want)
	}

	expression := firstParsedQueryExpression(t, queryText)
	pipe, ok := expression.(*PipeExpression)
	if !ok {
		t.Fatalf("expression = %T, want *PipeExpression", expression)
	}
	join, ok := pipe.Operator().(*JoinOperator)
	if !ok {
		t.Fatalf("operator = %T, want *JoinOperator", pipe.Operator())
	}
	if !syntaxElementIsMissing(join.Expression()) || join.ConditionClause() != nil {
		t.Fatalf("join recovery = expression:%#v condition:%#v", join.Expression(), join.ConditionClause())
	}
}

func diagnosticsEqualLocationAndMessage(got, want []Diagnostic) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Code != want[i].Code || got[i].Message != want[i].Message ||
			got[i].Start != want[i].Start || got[i].Length != want[i].Length {
			return false
		}
	}
	return true
}

func diagnosticCountByMessage(diagnostics []Diagnostic, code, message string) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Message == message {
			count++
		}
	}
	return count
}
