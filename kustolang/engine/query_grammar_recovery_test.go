package kustolang

import (
	"reflect"
	"testing"
)

func TestQueryGrammarRecoveryFactoriesMatchPinnedShapesAndDiagnosticOwnership(t *testing.T) {
	keywordSource := NewArraySource(Lex("to"))
	keywordMissingExpression := "Missing expression: If the keyword 'to' is intended to be part of an expression it needs to be bracketted as ['to']."

	tests := []struct {
		name       string
		create     func() SyntaxElement
		kind       SyntaxKind
		own        []string
		descendant []string
	}{
		{"named expression", func() SyntaxElement { return QueryGrammarCreateMissingNamedExpression(nil, 0) }, SyntaxKindSimpleNamedExpression, []string{"Missing name"}, nil},
		{"scan assignment", func() SyntaxElement { return QueryGrammarCreateMissingScanAssignment(nil, 0) }, SyntaxKindScanAssignment, []string{"Missing name"}, nil},
		{"join on", func() SyntaxElement { return QueryGrammarCreateMissingJoinOnClause(nil, 0) }, SyntaxKindJoinOnClause, []string{"Missing join on condition clause"}, nil},
		{"function call", func() SyntaxElement { return QueryGrammarCreateMissingFunctionCall(nil, 0) }, SyntaxKindFunctionCallExpression, []string{"Missing function call"}, nil},
		{"function call expression", func() SyntaxElement { return QueryGrammarCreateMissingFunctionCallExpression(nil, 0) }, SyntaxKindFunctionCallExpression, []string{"Missing function call"}, nil},
		{"schema type", func() SyntaxElement { return QueryGrammarCreateMissingSchemaType(nil, 0) }, SyntaxKindSchemaTypeExpression, []string{"Missing schema declaration"}, nil},
		{"row schema", func() SyntaxElement { return QueryGrammarCreateMissingRowSchema(nil, 0) }, SyntaxKindRowSchema, []string{"Missing schema declaration"}, nil},
		{"inline external table kind", func() SyntaxElement { return QueryGrammarCreateMissingInlineExternalTableKindClause(nil, 0) }, SyntaxKindInlineExternalTableKindClause, []string{"External Table Kind definition is expected."}, nil},
		{"inline external table data format", func() SyntaxElement { return QueryGrammarCreateMissingInlineExternalTableDataFormatClause(nil, 0) }, SyntaxKindInlineExternalTableDataFormatClause, []string{"DataFormat definition is expected."}, nil},
		{"inline external table connection strings", func() SyntaxElement {
			return QueryGrammarCreateMissingInlineExternalTableConnectionStringsClause(nil, 0)
		}, SyntaxKindInlineExternalTableConnectionStringsClause, []string{"Connection strings definition is expected."}, nil},
		{"path format tokens", func() SyntaxElement { return QueryGrammarCreateMissingPathFormatTokens(keywordSource, 0) }, SyntaxKindInlineExternalTablePathFormatPartitionColumnReference, []string{"External Table Path Format is empty."}, []string{"Missing expression"}},
		{"partition column", func() SyntaxElement { return QueryGrammarCreateMissingPartitionColumnDeclaration(keywordSource, 0) }, SyntaxKindPartitionColumnDeclaration, []string{"Partition column definition is expected."}, []string{keywordMissingExpression}},
		{"evaluate row schema", func() SyntaxElement { return QueryGrammarCreateMissingEvaluateRowSchema(nil, 0) }, SyntaxKindEvaluateRowSchema, []string{"Missing schema declaration"}, nil},
		{"make series expression", func() SyntaxElement { return QueryGrammarCreateMissingMakeSeriesExpression(keywordSource, 0) }, SyntaxKindMakeSeriesExpression, nil, []string{keywordMissingExpression}},
		{"mv expand expression", func() SyntaxElement { return QueryGrammarCreateMissingMvExpandExpression(keywordSource, 0) }, SyntaxKindMvExpandExpression, nil, []string{keywordMissingExpression}},
		{"mv apply expression", func() SyntaxElement { return QueryGrammarCreateMissingMvApplyExpression(keywordSource, 0) }, SyntaxKindMvApplyExpression, nil, []string{keywordMissingExpression}},
		{"mv apply subquery", func() SyntaxElement { return QueryGrammarCreateMissingMvApplySubqueryExpression(keywordSource, 0) }, SyntaxKindMvApplySubqueryExpression, nil, []string{keywordMissingExpression}},
		{"fork expression", func() SyntaxElement { return QueryGrammarCreateMissingForkExpression(keywordSource, 0) }, SyntaxKindForkExpression, nil, []string{keywordMissingExpression}},
		{"partition operand", func() SyntaxElement { return QueryGrammarCreateMissingPartitionOperand(keywordSource, 0) }, SyntaxKindPartitionSubquery, nil, []string{keywordMissingExpression}},
		{"statement", func() SyntaxElement { return QueryGrammarCreateMissingStatement(nil, 0) }, SyntaxKindExpressionStatement, []string{"Missing statement"}, nil},
		{"statement element", func() SyntaxElement { return QueryGrammarCreateMissingStatementElement(nil, 0) }, SyntaxKindSeparatedElement, nil, []string{"Missing statement"}},
		{"name and type", func() SyntaxElement { return QueryGrammarCreateMissingNameAndTypeDeclaration(nil, 0) }, SyntaxKindNameAndTypeDeclaration, []string{"Missing parameter"}, nil},
		{"function parameter", func() SyntaxElement { return QueryGrammarCreateMissingFunctionParameter(nil, 0) }, SyntaxKindFunctionParameter, []string{"Missing parameter"}, nil},
		{"named parameter", func() SyntaxElement { return QueryGrammarCreateMissingNamedParameter(nil, 0) }, SyntaxKindNamedParameter, []string{"Missing parameter"}, nil},
		{"function declaration", func() SyntaxElement { return QueryGrammarCreateMissingFunctionDeclaration(nil, 0) }, SyntaxKindFunctionDeclaration, []string{"Missing function declaration"}, nil},
		{"materialized base", func() SyntaxElement {
			return QueryGrammarCreateMissingMaterializedViewCombineBaseClause(keywordSource, 0)
		}, SyntaxKindMaterializedViewCombineClause, []string{"Missing base clause"}, []string{keywordMissingExpression}},
		{"materialized delta", func() SyntaxElement {
			return QueryGrammarCreateMissingMaterializedViewCombineDeltaClause(keywordSource, 0)
		}, SyntaxKindMaterializedViewCombineClause, []string{"Missing delta clause"}, []string{keywordMissingExpression}},
		{"materialized aggregates", func() SyntaxElement {
			return QueryGrammarCreateMissingMaterializedViewCombineAggregatesClause(keywordSource, 0)
		}, SyntaxKindMaterializedViewCombineClause, []string{"Missing aggregates clause"}, []string{keywordMissingExpression}},
		{"materialized name", func() SyntaxElement {
			return QueryGrammarCreateMissingMaterializedViewCombineNameClause(keywordSource, 0)
		}, SyntaxKindMaterializedViewCombineExpression, nil, []string{keywordMissingExpression}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := test.create()
			second := test.create()
			if first == nil || second == nil {
				t.Fatalf("factory returned nil: first=%#v second=%#v", first, second)
			}
			if first == second {
				t.Fatal("factory reused a syntax element")
			}
			if first.Kind() != test.kind || first.Width() != 0 {
				t.Fatalf("kind/width = %v/%d, want %v/0", first.Kind(), first.Width(), test.kind)
			}
			if got := queryGrammarRecoveryDiagnosticDescriptions(first.Diagnostics()); !reflect.DeepEqual(got, test.own) {
				t.Fatalf("owned diagnostics = %#v, want %#v", got, test.own)
			}
			if got := queryGrammarRecoveryDescendantDiagnosticDescriptions(first); !reflect.DeepEqual(got, test.descendant) {
				t.Fatalf("descendant diagnostics = %#v, want %#v", got, test.descendant)
			}
		})
	}
}

func TestQueryGrammarCreateMissingTokenLiteralOverloads(t *testing.T) {
	tokens := []string{"first", "last"}
	fromList := QueryGrammarCreateMissingTokenLiteralFromList(tokens)
	variadic := QueryGrammarCreateMissingTokenLiteral("first", "last")
	tokens[0] = "changed after factory creation"

	for name, factory := range map[string]func(Source[LexicalToken], int) Expression{
		"list": fromList, "variadic": variadic,
	} {
		t.Run(name, func(t *testing.T) {
			first := factory(nil, 0)
			second := factory(nil, 0)
			if first == second {
				t.Fatal("factory closure reused a syntax node")
			}
			if first.Kind() != SyntaxKindTokenLiteralExpression || first.Width() != 0 {
				t.Fatalf("kind/width = %v/%d", first.Kind(), first.Width())
			}
			if got := queryGrammarRecoveryDiagnosticDescriptions(first.Diagnostics()); !reflect.DeepEqual(got, []string{"Expected one of: first, last"}) {
				t.Fatalf("diagnostics = %#v", got)
			}
			if got := queryGrammarRecoveryDescendantDiagnosticDescriptions(first); len(got) != 0 {
				t.Fatalf("descendant diagnostics = %#v", got)
			}
		})
	}
}

func queryGrammarRecoveryDiagnosticDescriptions(diagnostics []Diagnostic) []string {
	if len(diagnostics) == 0 {
		return nil
	}
	descriptions := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		descriptions[i] = diagnostic.Description
	}
	return descriptions
}

func queryGrammarRecoveryDescendantDiagnosticDescriptions(element SyntaxElement) []string {
	var descriptions []string
	for i := 0; i < element.ChildCount(); i++ {
		child := element.GetChild(i)
		if child == nil {
			continue
		}
		descriptions = append(descriptions, queryGrammarRecoveryDiagnosticDescriptions(child.Diagnostics())...)
		descriptions = append(descriptions, queryGrammarRecoveryDescendantDiagnosticDescriptions(child)...)
	}
	return descriptions
}
