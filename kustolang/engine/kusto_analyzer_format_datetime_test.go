package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerReportsWhereComparison(t *testing.T) {
	text := "T | where format_datetime(Timestamp,'yyyy') == '2024'"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate ||
		diagnostic.Category != DiagnosticCategoryPerformance ||
		diagnostic.Severity != DiagnosticSeveritySuggestion ||
		diagnostic.Start != strings.Index(text, "format_datetime") ||
		diagnostic.Length != len("format_datetime(Timestamp,'yyyy')") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	wantMessage := "Avoid using the 'format_datetime' function in a filter or predicate.\nInstead, use specific datetime functions like startofday or bin."
	if diagnostic.Message != wantMessage {
		t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerFixesDirectFunctionAlternative(t *testing.T) {
	text := "T | where format_datetime(Timestamp,'yyyy') == '2024'"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	analyzer := NewAvoidUsingFormatDateTimeInPredicateAnalyzer()
	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}
	if actions[0].Base().Kind != "convert_format_datetime" ||
		actions[0].Base().Title != "Change to 'getyear'" ||
		actions[0].Base().Description != "Change use of function 'format_datetime' to function 'getyear'" {
		t.Fatalf("action metadata = %#v", actions[0].Base())
	}

	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostics[0].Start, CodeActionOptionsDefault, CancellationToken{})
	change, ok := singleTextChange(t, result)
	if !ok || change.ChangedText != "T | where getyear(Timestamp) == '2024'" {
		t.Fatalf("change = %#v", result.Actions)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerFixesDatetimePartAlternative(t *testing.T) {
	text := "T | where format_datetime(Timestamp,'fff') == '001'"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	analyzer := NewAvoidUsingFormatDateTimeInPredicateAnalyzer()
	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}
	if actions[0].Base().Title != "Change to 'datetime_part('millisecond',)'" {
		t.Fatalf("action title = %q", actions[0].Base().Title)
	}

	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostics[0].Start, CodeActionOptionsDefault, CancellationToken{})
	change, ok := singleTextChange(t, result)
	if !ok || change.ChangedText != "T | where datetime_part('millisecond', Timestamp) == '001'" {
		t.Fatalf("change = %#v", result.Actions)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerFixRemovesOuterIntCast(t *testing.T) {
	text := "T | where toint(format_datetime(Timestamp,'yyyy')) == 2024"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	analyzer := NewAvoidUsingFormatDateTimeInPredicateAnalyzer()
	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}

	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostics[0].Start, CodeActionOptionsDefault, CancellationToken{})
	change, ok := singleTextChange(t, result)
	if !ok || change.ChangedText != "T | where getyear(Timestamp) == 2024" {
		t.Fatalf("change = %#v", result.Actions)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerSuppressesUnsupportedFixes(t *testing.T) {
	tests := []string{
		"T | where format_datetime(Timestamp,'yyyy','extra') == '2024'",
		"T | where format_datetime(Timestamp, Format) == '2024'",
		"T | where format_datetime(Timestamp,'yyyy-MM') == '2024'",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			code := Parse(text)
			call := parsedFunctionCallByName(t, code, "format_datetime")
			call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

			analyzer := NewAvoidUsingFormatDateTimeInPredicateAnalyzer()
			diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
			if len(actions) != 0 {
				t.Fatalf("actions = %#v, want none", actions)
			}
		})
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerReportsFilterExpression(t *testing.T) {
	text := "T | where format_datetime(Timestamp,'yyyy')"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Start != strings.Index(text, "format_datetime") {
		t.Fatalf("diagnostics = %#v, want one format_datetime diagnostic", diagnostics)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerReportsNonFilterComparison(t *testing.T) {
	text := "T | extend Flag=format_datetime(Timestamp,'yyyy') == '2024'"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "format_datetime")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))

	diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Start != strings.Index(text, "format_datetime") {
		t.Fatalf("diagnostics = %#v, want one format_datetime diagnostic", diagnostics)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerSkipsProjectionAndUnboundCalls(t *testing.T) {
	projection := Parse("T | project Day=format_datetime(Timestamp,'yyyy')")
	projectionCall := parsedFunctionCallByName(t, projection, "format_datetime")
	projectionCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.FormatDatetime, ScalarTypes.String.Type()))
	if diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Analyze(projection, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("projection diagnostics = %#v, want empty", diagnostics)
	}

	unbound := Parse("T | where format_datetime(Timestamp,'yyyy') == '2024'")
	if diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingFormatDateTimeInPredicateAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingFormatDateTimeInPredicateAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate ||
		diagnostics[0].Description != "Avoid using format_datetime() in a filter or predicate" ||
		diagnostics[0].Message != "Avoid using the 'format_datetime' function in a filter or predicate.\nInstead, use specific datetime functions like startofday or bin." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func singleTextChange(t *testing.T, result CodeActionResult) (ChangeTextAction, bool) {
	t.Helper()
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("fix result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	return change, ok
}
