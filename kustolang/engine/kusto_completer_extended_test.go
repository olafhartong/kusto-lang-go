package kustolang

import "testing"

func TestKustoCompleterExtendedSyntaxUsesMalformedQueryLookahead(t *testing.T) {
	query := "StormEvents | join "
	if diagnostics := Parse(query).GetSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("incomplete join query should retain syntax diagnostics")
	}

	service := NewKustoCodeService(query, DefaultGlobalState)
	ordinary := service.GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
	if _, ok := completionItemByDisplayAndRank(ordinary, "dynamic ()", CompletionRankOther); ok {
		t.Fatal("default completion unexpectedly included extended syntax")
	}

	options := CompletionOptionsDefault.WithIncludeExtendedSyntax(true)
	extended := service.GetCompletionItems(len(query), options, CancellationToken{}).Items
	item, ok := completionItemByDisplayAndRank(extended, "dynamic ()", CompletionRankOther)
	if !ok {
		t.Fatalf("extended completion missing dynamic lookahead: %#v", completionDisplayTexts(extended))
	}
	if item.priority != CompletionPriorityLow || item.MatchText != "dynamic()" || item.BeforeText() != "dynamic (" || item.AfterText() != ")" || !item.HasCaret() {
		t.Fatalf("extended item = display:%q match:%q before:%q after:%q rank:%d priority:%d", item.DisplayText, item.MatchText, item.BeforeText(), item.AfterText(), item.rank, item.priority)
	}

	valid := "print x=1"
	if diagnostics := Parse(valid).GetSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid control query diagnostics = %#v", diagnostics)
	}
	_ = NewKustoCodeService(valid, DefaultGlobalState).GetCompletionItems(len(valid), options, CancellationToken{})
}

func TestCreateExtendedSyntaxCompletionCombinesApplyText(t *testing.T) {
	parts := []CompletionItem{
		NewCompletionItem(CompletionKindQueryPrefix, "render"),
		NewCompletionItem(CompletionKindKeyword, "with"),
		NewCompletionItem(CompletionKindPunctuation, "(", WithCompletionAfterText(")")),
	}
	item := createExtendedSyntaxCompletion(parts)
	if item.DisplayText != "render with (" || item.MatchText != "renderwith(" || item.BeforeText() != "render with (" || item.AfterText() != ")" || !item.HasCaret() {
		t.Fatalf("combined item = display:%q match:%q before:%q after:%q caret:%t", item.DisplayText, item.MatchText, item.BeforeText(), item.AfterText(), item.HasCaret())
	}
	if item.rank != CompletionRankOther || item.priority != CompletionPriorityLow {
		t.Fatalf("combined rank/priority = %d/%d", item.rank, item.priority)
	}
}

func TestKustoCompleterExtendedSyntaxUsesMalformedCommandLookahead(t *testing.T) {
	command := ".show "
	if diagnostics := Parse(command).GetSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("incomplete show command should retain syntax diagnostics")
	}

	service := NewKustoCodeService(command)
	ordinary := service.GetCompletionItems(len(command), CompletionOptionsDefault, CancellationToken{}).Items
	if _, ok := completionItemByDisplayAndRank(ordinary, "database identity", CompletionRankOther); ok {
		t.Fatal("default command completion unexpectedly included extended syntax")
	}

	options := CompletionOptionsDefault.WithIncludeExtendedSyntax(true)
	items := service.GetCompletionItems(len(command), options, CancellationToken{}).Items
	identity, ok := completionItemByDisplayAndRank(items, "database identity", CompletionRankOther)
	if !ok {
		t.Fatalf("extended command completion missing database identity: %#v", completionDisplayTexts(items))
	}
	if identity.Kind != CompletionKindCommandPrefix || identity.priority != CompletionPriorityLow || identity.MatchText != "databaseidentity" || identity.BeforeText() != "database identity" || identity.AfterText() != "" {
		t.Fatalf("identity item = kind:%d match:%q before:%q after:%q rank:%d priority:%d", identity.Kind, identity.MatchText, identity.BeforeText(), identity.AfterText(), identity.rank, identity.priority)
	}

	create := NewKustoCodeService(".create ").GetCompletionItems(len(".create "), options, CancellationToken{}).Items
	table, ok := completionItemByDisplayAndRank(create, "table ()", CompletionRankOther)
	if !ok || !table.HasCaret() {
		t.Fatalf("extended create completion missing table placeholder: %#v", completionDisplayTexts(create))
	}
	if function, ok := completionItemByDisplayAndRank(create, "function () { }", CompletionRankOther); !ok || !function.HasCaret() {
		t.Fatalf("extended create completion missing function placeholder: %#v", completionDisplayTexts(create))
	}

	valid := ".show version"
	if diagnostics := Parse(valid).GetSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid command diagnostics = %#v", diagnostics)
	}
	_ = NewKustoCodeService(valid).GetCompletionItems(len(valid), options, CancellationToken{})
}

func completionItemByDisplayAndRank(items []CompletionItem, display string, rank CompletionRank) (CompletionItem, bool) {
	for _, item := range items {
		if item.DisplayText == display && item.rank == rank {
			return item, true
		}
	}
	return CompletionItem{}, false
}
