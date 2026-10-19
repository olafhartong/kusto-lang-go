package kustolang

import "testing"

func TestCheckGraphMatchPatternDiagnostics(t *testing.T) {
	tests := []struct {
		name string
		text string
		code string
		msg  string
	}{
		{
			name: "starts with edge",
			text: `T | graph-match -->(a)`,
			code: "KS239",
			msg:  "graph-match element expected type is 'node', but received 'edge' type.",
		},
		{
			name: "adjacent nodes",
			text: `T | graph-match (a)(b)`,
			code: "KS239",
			msg:  "graph-match element expected type is 'edge', but received 'node' type.",
		},
		{
			name: "ends with edge",
			text: `T | graph-match (a)-->`,
			code: "KS238",
			msg:  "Missing graph-match pattern element.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := parseFirstGraphMatchOperator(t, tt.text)
			diagnostics := CheckGraphMatchPatterns(operator.Patterns())
			if !diagnosticsContainMessage(diagnostics, tt.code, tt.msg) {
				t.Fatalf("expected %s %q, got %#v", tt.code, tt.msg, diagnostics)
			}
		})
	}
}

func TestCheckGraphMatchPatternValid(t *testing.T) {
	operator := parseFirstGraphMatchOperator(t, `T | graph-match (a)-[e*1..3]->(b), (b)-->(c)`)
	if diagnostics := CheckGraphMatchPatterns(operator.Patterns()); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}

func parseFirstGraphMatchOperator(t *testing.T, text string) *GraphMatchOperator {
	t.Helper()
	expr := ParseExpression(text)
	pipe, ok := expr.(*PipeExpression)
	if !ok {
		t.Fatalf("expression = %#v, want pipe", expr)
	}
	operator, ok := pipe.Operator().(*GraphMatchOperator)
	if !ok {
		t.Fatalf("operator = %#v, want graph-match", pipe.Operator())
	}
	return operator
}
