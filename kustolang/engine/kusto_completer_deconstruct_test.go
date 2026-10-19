package kustolang

import "testing"

func TestKustoCompleterOffersTupleDeconstructionAfterInsertedOpenParen(t *testing.T) {
	pair := NewFunctionSymbol("pair", NewSymbolSignature(NewTupleSymbolOf(
		NewColumnSymbol("First", ScalarTypes.Long.Type()),
		NewColumnSymbol("Second", ScalarTypes.String.Type()),
	)))
	globals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{pair})
	text := "print (pair()"
	position := len("print (")
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("missing outer close parenthesis %q produced no diagnostics", text)
	}
	info := service.GetCompletionItems(position, CompletionOptionsDefault, CancellationToken{})
	if info.EditStart != position || info.EditLength != len("pair") {
		t.Fatalf("deconstruct edit range = %d/%d, want %d/%d", info.EditStart, info.EditLength, position, len("pair"))
	}
	items := info.Items
	item := requireCompletionKind(t, items, "(First, ...) = ", CompletionKindExample)
	if item.MatchText != "pair" || item.OrderText != "011_pair" || item.priority != CompletionPriorityTop {
		t.Fatalf("deconstruct item = match:%q order:%q priority:%d", item.MatchText, item.OrderText, item.priority)
	}
	want := []CompletionText{
		NewCompletionTextEdit("First"),
		NewCompletionText(", ", false, false),
		NewCompletionTextEdit("Second"),
		NewCompletionText(") = ", false, false),
	}
	if len(item.ApplyTexts) != len(want) {
		t.Fatalf("deconstruct apply texts = %#v", item.ApplyTexts)
	}
	for i := range want {
		if item.ApplyTexts[i] != want[i] {
			t.Fatalf("deconstruct apply text %d = %#v, want %#v", i, item.ApplyTexts[i], want[i])
		}
	}
	if item.BeforeText() != "" || item.AfterText() != "First, Second) = " {
		t.Fatalf("deconstruct edit text = before:%q after:%q", item.BeforeText(), item.AfterText())
	}
	builder := newKustoCompletionBuilder()
	mode := newKustoCompleter(ParseAndAnalyze(text, globals), CompletionOptionsDefault, CancellationToken{}).
		getDeconstructCompletions(position, builder)
	if mode != completionModeCombined || builder.count() != 1 {
		t.Fatalf("deconstruct special case = mode:%d items:%#v", mode, completionDisplayTexts(builder.items()))
	}
}

func TestKustoCompleterDeconstructsNativeSeriesStatsTuple(t *testing.T) {
	text := "print (series_stats(dynamic([1, 2]))"
	position := len("print (")
	items := NewKustoCodeService(text).GetCompletionItems(position, CompletionOptionsDefault, CancellationToken{}).Items
	item, ok := completionItemByMatchText(items, "series_stats")
	if !ok || item.Kind != CompletionKindExample {
		t.Fatalf("series_stats deconstruction missing: %#v", completionDisplayTexts(items))
	}
	if len(item.ApplyTexts) < 4 || item.DisplayText == "" {
		t.Fatalf("series_stats deconstruction = %#v", item)
	}
}

func TestKustoCompleterDoesNotOfferInvalidDeconstruction(t *testing.T) {
	one := NewFunctionSymbol("one", NewSymbolSignature(NewTupleSymbolOf(
		NewColumnSymbol("Only", ScalarTypes.Long.Type()),
	)))
	pair := NewFunctionSymbol("pair", NewSymbolSignature(NewTupleSymbolOf(
		NewColumnSymbol("First", ScalarTypes.Long.Type()),
		NewColumnSymbol("Second", ScalarTypes.String.Type()),
	)))
	globals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{one, pair})
	tests := []struct {
		text     string
		position int
	}{
		{text: "print (pair())", position: len("print (")},
		{text: "print (one()", position: len("print (")},
		{text: "print (1", position: len("print (")},
		{text: "print (Alias=pair()", position: len("print (")},
		{text: "print (pair(); print Result=", position: len("print (pair(); print Result=")},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			items := NewKustoCodeService(test.text, globals).GetCompletionItems(test.position, CompletionOptionsDefault, CancellationToken{}).Items
			if item, ok := completionItemByMatchText(items, "pair"); ok && item.Kind == CompletionKindExample {
				t.Fatalf("invalid deconstruction %q returned pair template: %#v", test.text, completionDisplayTexts(items))
			}
		})
	}
}

func completionItemByMatchText(items []CompletionItem, matchText string) (CompletionItem, bool) {
	for _, item := range items {
		if item.MatchText == matchText {
			return item, true
		}
	}
	return CompletionItem{}, false
}
