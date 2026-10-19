package kustolang

import "testing"

func TestParsePartitionBracedBodyUsesFullExpressionProduction(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		wantBodyKind    SyntaxKind
		wantDiagnostics []string
	}{
		{
			name:         "explicit source pipeline",
			text:         `Source | partition by Tenant { Other | where Enabled | project Tenant }`,
			wantBodyKind: SyntaxKindPipeExpression,
		},
		{
			name:            "empty body",
			text:            `Source | partition by Tenant {}`,
			wantBodyKind:    SyntaxKindNameReference,
			wantDiagnostics: []string{"KS006"},
		},
		{
			name:            "missing close brace",
			text:            `Source | partition by Tenant { Other | where Enabled`,
			wantBodyKind:    SyntaxKindPipeExpression,
			wantDiagnostics: []string{"KS005"},
		},
		{
			name:            "missing operator after pipe",
			text:            `Source | partition by Tenant { Other | }`,
			wantBodyKind:    SyntaxKindPipeExpression,
			wantDiagnostics: []string{"KS176"},
		},
		{
			name:            "unexpected unpiped tail",
			text:            `Source | partition by Tenant { Other Tail }`,
			wantBodyKind:    SyntaxKindNameReference,
			wantDiagnostics: []string{"KS005", "KS005", "KS198"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := ParseQuery(tt.text)
			partition := parsedPartitionOperator(t, block)
			operand, ok := partition.Operand().(*PartitionQuery)
			if !ok {
				t.Fatalf("partition operand = %#v, want *PartitionQuery", partition.Operand())
			}
			if got := operand.Query().Kind(); got != tt.wantBodyKind {
				t.Errorf("body kind = %v, want %v", got, tt.wantBodyKind)
			}
			assertPartitionDiagnosticCodes(t, block, tt.wantDiagnostics)
		})
	}
}

func parsedPartitionOperator(t *testing.T, block *QueryBlock) *PartitionOperator {
	t.Helper()
	var partition *PartitionOperator
	WalkSyntax(block, func(element SyntaxElement) {
		if partition != nil {
			return
		}
		if node, ok := WrapGeneratedSyntaxElement(element).(*PartitionOperator); ok {
			partition = node
		}
	})
	if partition == nil {
		t.Fatalf("partition operator not found in %#v", block)
	}
	return partition
}

func assertPartitionDiagnosticCodes(t *testing.T, block *QueryBlock, want []string) {
	t.Helper()
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(block, DiagnosticsIncludeSyntactic)
	got := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		got[i] = diagnostic.Code
	}
	if len(got) != len(want) {
		t.Fatalf("diagnostic codes = %#v, want %#v; diagnostics = %#v", got, want, diagnostics)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic codes = %#v, want %#v; diagnostics = %#v", got, want, diagnostics)
		}
	}
}
