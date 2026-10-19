package kustolang

import "testing"

func TestNativeCurrentMvApplySubqueryQueryOperatorContract(t *testing.T) {
	contextual := `__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (X:long)`
	tests := []struct {
		name          string
		root          string
		want          *Diagnostic
		wantMalformed bool
	}{
		{name: "scalar root", root: `1 + 2`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), 1, 5)},
		{name: "source pipe root", root: `Other | count`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), 1, 5)},
		{name: "valid post-pipe operators", root: `where true | count`},
		{name: "contextual root", root: contextual},
		{name: "contextual pipe root", root: contextual + ` | count`},
		{name: "disallowed operator", root: `print x=1`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorNotAllowedInContext("print"), 1, 5)},
		{name: "malformed disallowed operator", root: `print`, wantMalformed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := cloneSyntaxElementPreservingGeneratedWrapper(parsedQueryExpression(t, test.root), true).(Expression)
			node := MustNewMvApplySubqueryExpressionTyped(subqueryToken(SyntaxKindOpenParenToken), root, subqueryToken(SyntaxKindCloseParenToken))
			if !NewBinder(NewGlobalState(nil)).TryBind(node) {
				t.Fatal("mv-apply subquery did not bind")
			}
			assertExactSubqueryDiagnostic(t, node.SemanticDiagnostics(), test.want)
			if test.wantMalformed && !diagnosticsContain(root.ContainedDiagnostics(), "KS006") {
				t.Fatalf("malformed root diagnostics = %#v, want KS006", root.ContainedDiagnostics())
			}
		})
	}
}

func TestNativeCurrentPartitionSubqueryQueryOperatorContract(t *testing.T) {
	contextual := `__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (X:long)`
	tests := []struct {
		name          string
		root          string
		want          *Diagnostic
		wantMalformed bool
	}{
		{name: "scalar root", root: `1 + 2`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), 1, 5)},
		{name: "source pipe root", root: `Other | count`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), 1, 5)},
		{name: "valid post-pipe operators", root: `where true | count`},
		{name: "contextual root is not allowed", root: contextual, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), 1, len(contextual))},
		{name: "disallowed operator", root: `print x=1`, want: exactSubqueryDiagnostic(DiagnosticQueryOperatorNotAllowedInContext("print"), 1, 5)},
		{name: "malformed disallowed operator", root: `print`, wantMalformed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := cloneSyntaxElementPreservingGeneratedWrapper(parsedQueryExpression(t, test.root), true).(Expression)
			node := MustNewPartitionSubqueryTyped(subqueryToken(SyntaxKindOpenParenToken), root, subqueryToken(SyntaxKindCloseParenToken))
			binder := NewBinder(NewGlobalState(nil))
			if !binder.TryBind(node) {
				t.Fatal("partition subquery did not bind")
			}
			diagnostics := newSimpleNodeBinder(binder).partitionSubqueryDiagnosticsWithContextualRoot(node, false)
			assertExactSubqueryDiagnostic(t, diagnostics, test.want)
			if test.wantMalformed && !diagnosticsContain(root.ContainedDiagnostics(), "KS006") {
				t.Fatalf("malformed root diagnostics = %#v, want KS006", root.ContainedDiagnostics())
			}
		})
	}
}

func TestNativeCurrentPartitionBindersApplyContextualRootPolicy(t *testing.T) {
	const contextual = `__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (X:long)`
	table := NewTableSymbol("T", NewColumnSymbol("Key", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	legacy := ParseQuery(`T | partition by Key (count)`)
	legacyOperator := &PartitionOperator{SyntaxNode: firstSyntaxNodeOfKind(legacy, SyntaxKindPartitionOperator)}
	legacyRoot := cloneSyntaxElementPreservingGeneratedWrapper(parsedQueryExpression(t, contextual), true).(Expression)
	legacyOperator.Operand().(*PartitionSubquery).SyntaxNode.children[1] = syntaxElementFromTyped(legacyRoot)
	if !NewBinder(globals).TryBind(legacy) {
		t.Fatal("legacy partition did not bind")
	}
	assertExactSubqueryDiagnostic(t, legacyOperator.SemanticDiagnostics(), exactSubqueryDiagnostic(DiagnosticQueryOperatorExpected(), len(`T | partition by Key (`), len(contextual)))

	contextualPartition := ParseQuery(`T | __partitionby Key (count)`)
	contextualOperator := &PartitionByOperator{SyntaxNode: firstSyntaxNodeOfKind(contextualPartition, SyntaxKindPartitionByOperator)}
	contextualRoot := cloneSyntaxElementPreservingGeneratedWrapper(parsedQueryExpression(t, contextual), true).(Expression)
	contextualOperator.SyntaxNode.children[5] = syntaxElementFromTyped(contextualRoot)
	if !NewBinder(globals).TryBind(contextualPartition) {
		t.Fatal("contextual partition did not bind")
	}
	assertExactSubqueryDiagnostic(t, contextualOperator.SemanticDiagnostics(), nil)
}

func subqueryToken(kind SyntaxKind) *SyntaxToken {
	return SyntaxTokenFromLexical(NewLexicalToken(kind, "", SyntaxText(kind)))
}

func exactSubqueryDiagnostic(diagnostic Diagnostic, start, length int) *Diagnostic {
	diagnostic = diagnostic.WithLocation(start, length)
	return &diagnostic
}

func assertExactSubqueryDiagnostic(t *testing.T, got []Diagnostic, want *Diagnostic) {
	t.Helper()
	if want == nil {
		if len(got) != 0 {
			t.Fatalf("semantic diagnostics = %#v, want none", got)
		}
		return
	}
	if len(got) != 1 || got[0] != *want {
		t.Fatalf("semantic diagnostics = %#v, want %#v", got, []Diagnostic{*want})
	}
}
