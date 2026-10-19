package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingNullStringComparisonAnalyzerReportsStringIsNullAndIsNotNull(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		function   *FunctionSymbol
		wantCode   string
		wantCall   string
		wantMsg    string
		wantNewFun string
	}{
		{
			name:       "isnull",
			text:       "T | extend Flag=isnull(Account)",
			function:   Functions.IsNull,
			wantCode:   KustoAnalyzerCodeAvoidUsingIsNullStringComparision,
			wantCall:   "isnull(Account)",
			wantMsg:    "Avoid using isnull on string arguments, use isempty() instead",
			wantNewFun: "isempty",
		},
		{
			name:       "isnotnull",
			text:       "T | extend Flag=isnotnull(Account)",
			function:   Functions.IsNotNull,
			wantCode:   KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison,
			wantCall:   "isnotnull(Account)",
			wantMsg:    "Avoid using isnotnull on string arguments, use isnotempty() instead",
			wantNewFun: "isnotempty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.text)
			call := firstParsedFunctionCall(t, code)
			call.SetSemanticInfo(NewSemanticInfoForSymbol(tt.function, ScalarTypes.Bool.Type()))
			setExpressionSemanticType(t, functionCallArgumentAt(call, 0), ScalarTypes.String.Type())

			analyzer := NewAvoidUsingNullStringComparisonAnalyzer()
			diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != tt.wantCode ||
				diagnostic.Category != DiagnosticCategoryCorrectness ||
				diagnostic.Severity != DiagnosticSeverityWarning ||
				diagnostic.Message != tt.wantMsg ||
				diagnostic.Start != strings.Index(tt.text, tt.function.SymbolName()) ||
				diagnostic.Length != len(tt.wantCall) {
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
			if actions[0].Base().Title != "Change to '"+tt.wantNewFun+"'" {
				t.Fatalf("action title = %q", actions[0].Base().Title)
			}

			result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
			if result.FailureReason != "" || len(result.Actions) == 0 {
				t.Fatalf("fix result = %#v", result)
			}
			change, ok := result.Actions[0].(ChangeTextAction)
			if !ok {
				t.Fatalf("first result action = %#v, want text change", result.Actions[0])
			}
			wantText := strings.Replace(tt.text, tt.function.SymbolName(), tt.wantNewFun, 1)
			if change.ChangedText != wantText {
				t.Fatalf("changed text = %q, want %q", change.ChangedText, wantText)
			}
		})
	}
}

func TestAvoidUsingNullStringComparisonAnalyzerSkipsUnboundNonStringAndNoArguments(t *testing.T) {
	unbound := Parse("T | extend Flag=isnull(Account)")
	if diagnostics := NewAvoidUsingNullStringComparisonAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	longArg := Parse("T | extend Flag=isnull(Account)")
	longCall := firstParsedFunctionCall(t, longArg)
	longCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.IsNull, ScalarTypes.Bool.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(longCall, 0), ScalarTypes.Long.Type())
	if diagnostics := NewAvoidUsingNullStringComparisonAnalyzer().Analyze(longArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("long diagnostics = %#v, want empty", diagnostics)
	}

	noArg := Parse("T | extend Flag=isnull()")
	noArgCall := firstParsedFunctionCall(t, noArg)
	noArgCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.IsNull, ScalarTypes.Bool.Type()))
	if diagnostics := NewAvoidUsingNullStringComparisonAnalyzer().Analyze(noArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("no-arg diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingNullStringComparisonAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingNullStringComparisonAnalyzer().Diagnostics()
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v, want two", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingIsNullStringComparision ||
		diagnostics[0].Description != "Avoid using isnull on string arguments" ||
		diagnostics[0].Message != "Avoid using isnull on string arguments, use isempty() instead" {
		t.Fatalf("first diagnostic metadata = %#v", diagnostics[0])
	}
	if diagnostics[1].Code != KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison ||
		diagnostics[1].Description != "Avoid using isnotnull on string arguments" ||
		diagnostics[1].Message != "Avoid using isnotnull on string arguments, use isnotempty() instead" {
		t.Fatalf("second diagnostic metadata = %#v", diagnostics[1])
	}
}
