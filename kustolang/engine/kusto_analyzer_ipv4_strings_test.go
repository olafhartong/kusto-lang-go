package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingHasWithIPv4StringsAnalyzerReportsBinaryOperatorsAndFixesFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		operator string
		function string
		wantText string
	}{
		{
			name:     "has",
			text:     "T | where Account has '192.168.1.1'",
			operator: "has",
			function: "has_ipv4",
			wantText: "T | where has_ipv4(Account, '192.168.1.1')",
		},
		{
			name:     "has cs",
			text:     "T | where Account has_cs '192.168.1.1'",
			operator: "has_cs",
			function: "has_ipv4",
			wantText: "T | where has_ipv4(Account, '192.168.1.1')",
		},
		{
			name:     "contains",
			text:     "T | where Account contains '192.168.1.1'",
			operator: "contains",
			function: "has_ipv4",
			wantText: "T | where has_ipv4(Account, '192.168.1.1')",
		},
		{
			name:     "hasprefix",
			text:     "T | where Account hasprefix '192.168.'",
			operator: "hasprefix",
			function: "has_ipv4_prefix",
			wantText: "T | where has_ipv4_prefix(Account, '192.168.')",
		},
		{
			name:     "hasprefix cs",
			text:     "T | where Account hasprefix_cs '192.168.'",
			operator: "hasprefix_cs",
			function: "has_ipv4_prefix",
			wantText: "T | where has_ipv4_prefix(Account, '192.168.')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := containsAnalyzerBoundCode(t, tt.text, true)
			analyzer := NewAvoidUsingHasWithIPv4StringsAnalyzer()

			diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings ||
				diagnostic.Category != DiagnosticCategoryPerformance ||
				diagnostic.Severity != DiagnosticSeveritySuggestion ||
				diagnostic.Start != strings.Index(tt.text, tt.operator) ||
				diagnostic.Length != len(tt.operator) {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
			wantMessage := "Avoid using the '" + tt.operator + "' operator with IPv4 strings. Use the '" + tt.function + "' function instead."
			if diagnostic.Message != wantMessage {
				t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
			}

			actions := analyzer.Fixer().GetFixActions(code, diagnostic, []Diagnostic{diagnostic}, CodeActionOptionsDefault, nil, CancellationToken{})
			if len(actions) != 1 {
				t.Fatalf("actions = %#v, want one", actions)
			}
			action, ok := actions[0].(ApplyAction)
			if !ok {
				t.Fatalf("action = %#v, want apply action", actions[0])
			}
			if actions[0].Base().Title != "Change to '"+tt.function+"'" {
				t.Fatalf("action title = %q", actions[0].Base().Title)
			}
			result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
			if result.FailureReason != "" || len(result.Actions) == 0 {
				t.Fatalf("fix result = %#v", result)
			}
			change, ok := result.Actions[0].(ChangeTextAction)
			if !ok || change.ChangedText != tt.wantText {
				t.Fatalf("change = %#v, want %q", result.Actions[0], tt.wantText)
			}
		})
	}
}

func TestAvoidUsingHasWithIPv4StringsAnalyzerReportsHasAnyAndFixesFunctionCall(t *testing.T) {
	text := "T | where Account has_any ('10.1.1.1', '10.1.1.2')"
	code, _ := ipv4AnalyzerBoundHasAnyCode(t, text, true)
	analyzer := NewAvoidUsingHasWithIPv4StringsAnalyzer()

	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings ||
		diagnostic.Start != strings.Index(text, "has_any") ||
		diagnostic.Length != len("has_any") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	wantMessage := "Avoid using the 'has_any' operator with IPv4 strings. Use the 'has_any_ipv4' function instead."
	if diagnostic.Message != wantMessage {
		t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
	}

	actions := analyzer.Fixer().GetFixActions(code, diagnostic, []Diagnostic{diagnostic}, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}
	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("fix result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "T | where has_any_ipv4(Account, '10.1.1.1', '10.1.1.2')" {
		t.Fatalf("change = %#v", result.Actions[0])
	}
}

func TestAvoidUsingHasWithIPv4StringsAnalyzerSkipsNonDbColumnsAndInvalidIPv4Constants(t *testing.T) {
	nonDatabaseColumn, _ := containsAnalyzerBoundCode(t, "T | where Account has '192.168.1.1'", false)
	if diagnostics := NewAvoidUsingHasWithIPv4StringsAnalyzer().Analyze(nonDatabaseColumn, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("non-db diagnostics = %#v, want empty", diagnostics)
	}

	invalidIPv4, _ := containsAnalyzerBoundCode(t, "T | where Account has '999.1.1.1'", true)
	if diagnostics := NewAvoidUsingHasWithIPv4StringsAnalyzer().Analyze(invalidIPv4, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("invalid IPv4 diagnostics = %#v, want empty", diagnostics)
	}

	fullIPv4WithPrefixOperator, _ := containsAnalyzerBoundCode(t, "T | where Account hasprefix '192.168.1.1'", true)
	if diagnostics := NewAvoidUsingHasWithIPv4StringsAnalyzer().Analyze(fullIPv4WithPrefixOperator, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("full IPv4 prefix diagnostics = %#v, want empty", diagnostics)
	}

	invalidHasAny, _ := ipv4AnalyzerBoundHasAnyCode(t, "T | where Account has_any ('10.1.1.1', '999.1.1.1')", true)
	if diagnostics := NewAvoidUsingHasWithIPv4StringsAnalyzer().Analyze(invalidHasAny, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("invalid has_any diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingHasWithIPv4StringsAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingHasWithIPv4StringsAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings ||
		diagnostics[0].Category != DiagnosticCategoryPerformance ||
		diagnostics[0].Severity != DiagnosticSeveritySuggestion ||
		diagnostics[0].Description != "Avoid using has operator with IPv4 strings." ||
		diagnostics[0].Message != "Avoid using the 'has' operator with IPv4 strings. Use the 'has_ipv4' function instead." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func TestAnalyzerStringIPv4MatchingPreservesUpstreamNumberShape(t *testing.T) {
	if !analyzerStringIsIPv4("0.1.99.255") || !analyzerStringIsIPv4("01.02.03.04") {
		t.Fatal("valid IPv4 forms did not match")
	}
	if analyzerStringIsIPv4("001.2.3.4") || analyzerStringIsIPv4("256.1.1.1") || analyzerStringIsIPv4("1.2.3") {
		t.Fatal("invalid IPv4 forms matched")
	}
	if !analyzerStringIsIPv4Prefix("10.") || !analyzerStringIsIPv4Prefix("10.20.") || !analyzerStringIsIPv4Prefix("10.20.30.") {
		t.Fatal("valid IPv4 prefix forms did not match")
	}
	if analyzerStringIsIPv4Prefix("10.20") || analyzerStringIsIPv4Prefix("10.20.30.40") || analyzerStringIsIPv4Prefix("001.") {
		t.Fatal("invalid IPv4 prefix forms matched")
	}
}

func ipv4AnalyzerBoundHasAnyCode(t *testing.T, text string, databaseColumn bool) (*KustoCode, *ColumnSymbol) {
	t.Helper()
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	globals := DefaultGlobalState
	if databaseColumn {
		globals = NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", column)))
	}
	code := ParseWithGlobals(text, globals)
	hasAny := firstParsedHasAnyExpression(t, code)
	left, ok := hasAny.Left().(interface{ SetSemanticInfo(SemanticInfo) })
	if !ok {
		t.Fatalf("left expression %T cannot carry semantic info", hasAny.Left())
	}
	left.SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))
	return code, column
}

func firstParsedHasAnyExpression(t *testing.T, code *KustoCode) *HasAnyExpression {
	t.Helper()
	var hasAny *HasAnyExpression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if hasAny != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*HasAnyExpression); ok {
			hasAny = found
		}
	})
	if hasAny == nil {
		t.Fatalf("has_any expression not found in %q", code.Text)
	}
	return hasAny
}
