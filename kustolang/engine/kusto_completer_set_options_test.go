package kustolang

import "testing"

func TestKustoCompleterOffersIsolatedSetOptionNames(t *testing.T) {
	first := NewOptionSymbolWithType("first_option", "", ScalarTypes.String, nil)
	second := NewOptionSymbolWithType("second option", "", ScalarTypes.Bool, nil)
	globals := NewGlobalState(nil).WithOptions([]*OptionSymbol{first, second})
	for _, test := range []struct {
		text     string
		position int
	}{
		{text: "set ", position: len("set ")},
		{text: "set first_op", position: len("set first_op")},
		{text: "set  = true", position: len("set ")},
	} {
		t.Run(test.text, func(t *testing.T) {
			items := NewKustoCodeService(test.text, globals).GetCompletionItems(test.position, CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, "first_option", CompletionKindOption)
			secondItem := requireCompletionKind(t, items, "second option", CompletionKindOption)
			if secondItem.BeforeText() != "['second option']" {
				t.Fatalf("bracketed option insertion = %q", secondItem.BeforeText())
			}
			if len(items) != 2 {
				t.Fatalf("isolated option-name completions = %#v", completionDisplayTexts(items))
			}
		})
	}
}

func TestKustoCompleterOffersIsolatedSetOptionExamplesForMissingValue(t *testing.T) {
	text := "set query_language =   "
	service := NewKustoCodeService(text)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete set-option statement %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	for _, example := range []struct {
		display string
		match   string
	}{
		{display: "'csl'", match: "csl"},
		{display: "'kql'", match: "kql"},
		{display: "'sql'", match: "sql"},
	} {
		item := requireCompletionKind(t, items, example.display, CompletionKindExample)
		if item.BeforeText() != example.display || item.MatchText != example.match {
			t.Fatalf("set-option example %q = before:%q match:%q", example.display, item.BeforeText(), item.MatchText)
		}
	}
	if len(items) != 3 {
		t.Fatalf("isolated set-option completions = %#v, want only three examples", completionDisplayTexts(items))
	}
}

func TestKustoCompleterSetOptionExamplesPreserveEscapedQuotedValues(t *testing.T) {
	option := NewOptionSymbolWithType("custom_mode", "", ScalarTypes.String, []string{`'can\'t'`, `"quoted"`})
	globals := NewGlobalState(nil).WithOptions([]*OptionSymbol{option})
	text := "set custom_mode = "
	items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items

	single := requireCompletionKind(t, items, `'can\'t'`, CompletionKindExample)
	if single.BeforeText() != `'can\'t'` || single.MatchText != "can't" {
		t.Fatalf("single-quoted option example = before:%q match:%q", single.BeforeText(), single.MatchText)
	}
	double := requireCompletionKind(t, items, `"quoted"`, CompletionKindExample)
	if double.BeforeText() != `"quoted"` || double.MatchText != "quoted" {
		t.Fatalf("double-quoted option example = before:%q match:%q", double.BeforeText(), double.MatchText)
	}
}

func TestKustoCompleterDoesNotLeakSetOptionExamples(t *testing.T) {
	tests := []string{
		"set missing = ",
		"set query_language ",
		"set query_language = 'kql' ",
		"set query_language = ; print Result=",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			for _, example := range []string{"'csl'", "'kql'", "'sql'"} {
				requireNoCompletion(t, items, example)
			}
		})
	}
}

func TestKustoCompleterSetOptionValueContextBeforeExistingTokenAndAcrossStatements(t *testing.T) {
	beforeValue := "set query_language =   'kql'"
	position := len("set query_language =  ")
	items := NewKustoCodeService(beforeValue).GetCompletionItems(position, CompletionOptionsDefault, CancellationToken{}).Items
	for _, example := range []string{"'csl'", "'kql'", "'sql'"} {
		requireCompletionKind(t, items, example, CompletionKindExample)
	}

	datascope := "set query_language = ; set query_datascope = "
	items = NewKustoCodeService(datascope).GetCompletionItems(len(datascope), CompletionOptionsDefault, CancellationToken{}).Items
	for _, example := range []string{"'default'", "'all'", "'hotcache'"} {
		requireCompletionKind(t, items, example, CompletionKindExample)
	}
	for _, stale := range []string{"'csl'", "'kql'", "'sql'"} {
		requireNoCompletion(t, items, stale)
	}
}

func TestKustoCompleterSuppressesSetOptionExamplesAtLiteralEdges(t *testing.T) {
	for _, text := range []string{
		"set query_language = '",
		"set query_language = 'kql'",
		"set query_language = 'kql' ",
	} {
		t.Run(text, func(t *testing.T) {
			items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			for _, example := range []string{"'csl'", "'kql'", "'sql'"} {
				requireNoCompletion(t, items, example)
			}
		})
	}
}

func TestKustoCompleterKnownOptionWithoutExamplesIsIsolated(t *testing.T) {
	text := "set query_now = "
	service := NewKustoCodeService(text)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete set-option statement %q produced no diagnostics", text)
	}
	if items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items; len(items) != 0 {
		t.Fatalf("known option without examples should isolate an empty list, got %#v", completionDisplayTexts(items))
	}
}
