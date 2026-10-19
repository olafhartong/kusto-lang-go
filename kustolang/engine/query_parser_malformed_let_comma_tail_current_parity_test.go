package kustolang

import "testing"

func TestCurrentMalformedLetCommaTailRetainsTypedPrefix(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
	}{
		{
			name:  "malformed datatable name",
			query: `let G=datatableSource:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-match (a)-[e]->(b) where e.Weight > 0 project Src=a.Source, Weight=e.Weight`,
			diagnostics: []diagnosticLocation{
				{"KS142", 6, 15}, {"KS228", 22, 4}, {"KS142", 22, 4}, {"KS198", 26, 1},
			},
		},
		{
			name:  "missing let equals",
			query: `let Gdatatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-match (a)-[e]->(b) where e.Weight > 0 project Src=a.Source, Weight=e.Weight`,
			diagnostics: []diagnosticLocation{
				{"KS005", 14, 1}, {"KS005", 51, 1}, {"KS005", 51, 1}, {"KS005", 51, 1},
				{"KS141", 52, 1}, {"KS005", 53, 1}, {"KS198", 53, 1},
			},
		},
		{
			name:  "missing tabular parameter colon",
			query: `let Transform=(T(Key:long,Value:string), threshold:long) { T | where Key > threshold | project Value }; datatable(Key:long,Value:string)[1,'a'] | invoke Transform(0)`,
			diagnostics: []diagnosticLocation{
				{"KS211", 15, 1}, {"KS142", 17, 3}, {"KS228", 21, 4}, {"KS142", 21, 4},
				{"KS142", 26, 5}, {"KS228", 32, 6}, {"KS142", 32, 6}, {"KS005", 39, 1}, {"KS198", 39, 1},
			},
		},
		{
			name:  "malformed tabular parameter type",
			query: `let Transform=(T:Key:long,Value:string), threshold:long) { T | where Key > threshold | project Value }; datatable(Key:long,Value:string)[1,'a'] | invoke Transform(0)`,
			diagnostics: []diagnosticLocation{
				{"KS160", 17, 3}, {"KS005", 20, 1}, {"KS006", 20, 0}, {"KS005", 39, 1},
				{"KS005", 39, 1}, {"KS198", 39, 1},
			},
		},
		{
			name:  "missing function declaration paren",
			query: `let Transform=T:(Key:long,Value:string), threshold:long) { T | where Key > threshold | project Value }; datatable(Key:long,Value:string)[1,'a'] | invoke Transform(0)`,
			diagnostics: []diagnosticLocation{
				{"KS142", 14, 1}, {"KS142", 17, 3}, {"KS228", 21, 4}, {"KS142", 21, 4},
				{"KS005", 25, 1}, {"KS198", 25, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.query)
			statements, _ := query.GetChild(1).(*SyntaxList)
			if statements == nil || statements.Count() == 0 {
				t.Fatal("missing retained statement list")
			}
			first := statements.At(0).(*SeparatedElement)
			if _, ok := WrapGeneratedSyntaxElement(first.Element()).(*LetStatement); !ok {
				t.Fatalf("first statement = %T, want LetStatement", WrapGeneratedSyntaxElement(first.Element()))
			}
			if !queryBlockSkippedTokenKindForTest(query, SyntaxKindCommaToken) {
				t.Fatal("skipped tail does not begin with comma")
			}
			assertDiagnosticLocations(t, ParseAndAnalyze(tt.query).GetDiagnostics(), tt.diagnostics...)
		})
	}

	control := ParseQuery(`1 + 2, 3`)
	statements, _ := control.GetChild(1).(*SyntaxList)
	first := statements.At(0).(*SeparatedElement)
	expression := first.Element().GetChild(0)
	if expression == nil || expression.Kind() != SyntaxKindNone {
		t.Fatalf("ordinary expression comma fallback = %#v, want None node", expression)
	}
}

func queryBlockSkippedTokenKindForTest(query *QueryBlock, kind SyntaxKind) bool {
	skipped, ok := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
	return ok && skipped != nil && skipped.Tokens() != nil && skipped.Tokens().Count() > 0 && skipped.Tokens().At(0).Kind() == kind
}
