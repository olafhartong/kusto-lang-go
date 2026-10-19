package kustolang

import "testing"

func TestKustoCompleterOffersBooleanBinaryOperatorExamples(t *testing.T) {
	tests := []string{
		"print Result=true == ",
		"print Result=true != ",
		"print Result=true >= ",
		"print Result=true ==   ",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			service := NewKustoCodeService(text)
			if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
				t.Fatalf("incomplete binary expression %q produced no diagnostics", text)
			}
			items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, "false", CompletionKindExample)
			requireCompletionKind(t, items, "true", CompletionKindExample)
		})
	}
}

func TestKustoCompleterCombinesBinaryExamplesWithExpressionSymbols(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Flag", ScalarTypes.Bool.Type()),
		NewColumnSymbol("Other", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	text := "Events | where Flag == "
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete where expression %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "true", CompletionKindExample)
	requireCompletionKind(t, items, "false", CompletionKindExample)
	requireCompletionKind(t, items, "Other", CompletionKindColumn)
}

func TestKustoCompleterDoesNotInventBinaryExamples(t *testing.T) {
	tests := []string{
		"print Result='x' == ",
		"print Result=true and ",
		"print First=true == ; print Second=",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireNoCompletion(t, items, "false")
			requireNoCompletion(t, items, "true")
		})
	}
}
