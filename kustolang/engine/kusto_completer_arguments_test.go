package kustolang

import "testing"

func TestKustoCompleterSuggestsDatabaseFunctionArgumentNamesAndValues(t *testing.T) {
	function := completionArgumentTestFunction()
	globals := NewGlobalState(NewDatabaseSymbol("db", function))

	firstText := "print Result=Choose("
	firstCode := ParseAndAnalyze(firstText, globals)
	if _, ok := newKustoCompleter(firstCode, CompletionOptionsDefault, CancellationToken{}).functionArgumentCompletionContext(len(firstText)); !ok {
		t.Fatalf("no function argument context for incomplete call; diagnostics=%#v", firstCode.GetDiagnostics())
	}
	first := NewKustoCodeService(firstText, globals).GetCompletionItems(len(firstText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, first, "input=", CompletionKindParameter)
	requireCompletionKind(t, first, "mode=", CompletionKindParameter)
	requireCompletionKind(t, first, "enabled=", CompletionKindParameter)

	modeText := "print Result=Choose('value', "
	modeService := NewKustoCodeService(modeText, globals)
	if diagnostics := modeService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete argument query %q produced no diagnostics", modeText)
	}
	mode := modeService.GetCompletionItems(len(modeText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, mode, "'fast'", CompletionKindExample)
	requireCompletionKind(t, mode, "'safe'", CompletionKindExample)
	requireCompletionKind(t, mode, "mode=", CompletionKindParameter)
	requireCompletionKind(t, mode, "enabled=", CompletionKindParameter)
	requireNoCompletion(t, mode, "input=")
	requireNoCompletion(t, mode, "Choose(input, mode, enabled)")

	boolText := "print Result=Choose('value', 'fast', "
	boolean := NewKustoCodeService(boolText, globals).GetCompletionItems(len(boolText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, boolean, "false", CompletionKindExample)
	requireCompletionKind(t, boolean, "true", CompletionKindExample)
}

func TestKustoCompleterUsesNamedArgumentContextAndRequiresNamesAfterReordering(t *testing.T) {
	function := completionArgumentTestFunction()
	globals := NewGlobalState(NewDatabaseSymbol("db", function))

	valueText := "print Result=Choose(mode="
	values := NewKustoCodeService(valueText, globals).GetCompletionItems(len(valueText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, values, "'fast'", CompletionKindExample)
	requireCompletionKind(t, values, "'safe'", CompletionKindExample)

	reorderedText := "print Result=Choose(mode='fast', "
	reordered := NewKustoCodeService(reorderedText, globals).GetCompletionItems(len(reorderedText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, reordered, "input=", CompletionKindParameter)
	requireCompletionKind(t, reordered, "enabled=", CompletionKindParameter)
	requireNoCompletion(t, reordered, "mode=")
	requireNoCompletion(t, reordered, "true")
	requireNoCompletion(t, reordered, "false")

	extraText := "print Result=Choose('value', 'fast', true, "
	extraService := NewKustoCodeService(extraText, globals)
	if diagnostics := extraService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("excess-argument query %q produced no diagnostics", extraText)
	}
	if items := extraService.GetCompletionItems(len(extraText), CompletionOptionsDefault, CancellationToken{}).Items; len(items) != 0 {
		t.Fatalf("excess-argument completions = %#v, want isolated empty result", completionDisplayTexts(items))
	}
}

func TestKustoCompleterDoesNotOfferNamedArgumentsForBuiltIns(t *testing.T) {
	text := "print Result=strcat("
	items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireNoCompletion(t, items, "arg0=")
	requireNoCompletion(t, items, "arg1=")
}

func TestKustoCompleterMissingCloseDoesNotLeakIntoLaterStatement(t *testing.T) {
	function := completionArgumentTestFunction()
	globals := NewGlobalState(NewDatabaseSymbol("db", function))
	text := "print Result=Choose(; print Other="
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("malformed multi-statement query %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireNoCompletion(t, items, "'fast'")
	requireNoCompletion(t, items, "'safe'")
	requireNoCompletion(t, items, "mode=")
}

func completionArgumentTestFunction() *FunctionSymbol {
	return NewScalarFunctionSymbol("Choose", ScalarTypes.String.Type(),
		NewParameter("input", ScalarTypes.String.Type()),
		NewParameter("mode", ScalarTypes.String.Type(),
			WithArgumentKind(ArgumentKindLiteral),
			WithParameterValues("Fast", "Safe")),
		NewParameter("enabled", ScalarTypes.Bool.Type()))
}
