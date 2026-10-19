package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingContainsAnalyzerReportsContainsOnDatabaseColumn(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		operator string
		hasOp    string
	}{
		{name: "contains", text: "T | where Account contains 'failure'", operator: "contains", hasOp: "has"},
		{name: "contains cs", text: "T | where Account contains_cs 'failure'", operator: "contains_cs", hasOp: "has_cs"},
		{name: "not contains", text: "T | where Account !contains 'failure'", operator: "!contains", hasOp: "!has"},
		{name: "not contains cs", text: "T | where Account !contains_cs 'failure'", operator: "!contains_cs", hasOp: "!has_cs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := containsAnalyzerBoundCode(t, tt.text, true)
			analyzer := NewAvoidUsingContainsAnalyzer()

			diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != KustoAnalyzerCodeAvoidUsingContains ||
				diagnostic.Category != DiagnosticCategoryPerformance ||
				diagnostic.Severity != DiagnosticSeveritySuggestion ||
				diagnostic.Start != strings.Index(tt.text, tt.operator) ||
				diagnostic.Length != len(tt.operator) {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}

			actions := analyzer.Fixer().GetFixActions(code, diagnostic, []Diagnostic{diagnostic}, CodeActionOptionsDefault, nil, CancellationToken{})
			if len(actions) != 1 {
				t.Fatalf("actions = %#v, want one", actions)
			}
			action, ok := actions[0].(ApplyAction)
			if !ok {
				t.Fatalf("action = %#v, want apply action", actions[0])
			}
			if actions[0].Base().Title != "Change to '"+tt.hasOp+"'" {
				t.Fatalf("action title = %q", actions[0].Base().Title)
			}
			result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
			if result.FailureReason != "" || len(result.Actions) == 0 {
				t.Fatalf("fix result = %#v", result)
			}
			change, ok := result.Actions[0].(ChangeTextAction)
			if !ok {
				t.Fatalf("change action = %#v", result.Actions[0])
			}
			wantText := strings.Replace(tt.text, tt.operator, tt.hasOp, 1)
			if change.ChangedText != wantText {
				t.Fatalf("changed text = %q, want %q", change.ChangedText, wantText)
			}
		})
	}
}

func TestAvoidUsingContainsAnalyzerSkipsShortStringsAndNonDatabaseColumns(t *testing.T) {
	short := "T | where Account contains 'ab'"
	shortCode, _ := containsAnalyzerBoundCode(t, short, true)
	if diagnostics := NewAvoidUsingContainsAnalyzer().Analyze(shortCode, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("short-string diagnostics = %#v, want empty", diagnostics)
	}

	unbound := Parse("T | where Account contains 'failure'")
	if diagnostics := NewAvoidUsingContainsAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	local := "T | where Account contains 'failure'"
	localCode, _ := containsAnalyzerBoundCode(t, local, false)
	if diagnostics := NewAvoidUsingContainsAnalyzer().Analyze(localCode, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("local-column diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingContainsAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingContainsAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingContains ||
		diagnostics[0].Description != "Avoid using contains operator" ||
		diagnostics[0].Message != "Avoid using the 'contains' operator as it has a high compute price.\nUse the 'has' operator in cases when full term match is desired (see: https://aka.ms/kusto.stringterms)." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func containsAnalyzerBoundCode(t *testing.T, text string, databaseColumn bool) (*KustoCode, *ColumnSymbol) {
	t.Helper()
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	globals := DefaultGlobalState
	if databaseColumn {
		globals = NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", column)))
	}
	code := ParseWithGlobals(text, globals)
	binary := firstParsedBinaryExpression(t, code)
	left, ok := binary.Left().(interface{ SetSemanticInfo(SemanticInfo) })
	if !ok {
		t.Fatalf("left expression %T cannot carry semantic info", binary.Left())
	}
	left.SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	return code, column
}

func firstParsedBinaryExpression(t *testing.T, code *KustoCode) *BinaryExpression {
	t.Helper()
	var binary *BinaryExpression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if binary != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*BinaryExpression); ok {
			binary = found
		}
	})
	if binary == nil {
		t.Fatalf("binary expression not found in %q", code.Text)
	}
	return binary
}
