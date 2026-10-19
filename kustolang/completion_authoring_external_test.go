package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCompletionTextParsingAndItemDefaults(t *testing.T) {
	parsed := kustolang.ParseCompletionText(`func([[argument]])`)
	want := []kustolang.CompletionText{
		kustolang.NewCompletionText("func(", false, false),
		kustolang.NewCompletionTextEdit("argument"),
		kustolang.NewCompletionText(")", false, false),
	}
	if !reflect.DeepEqual(parsed, want) {
		t.Fatalf("parsed completion text = %#v, want %#v", parsed, want)
	}
	cursor := kustolang.ParseCompletionText("before|after")
	if len(cursor) != 2 || cursor[0].Text != "before" || cursor[0].Caret || cursor[1].Text != "after" || !cursor[1].Caret || cursor[1].Select {
		t.Fatalf("cursor completion text = %#v", cursor)
	}

	item := kustolang.NewSyntaxCompletionItem("where")
	if item.Kind != kustolang.CompletionKindSyntax || item.DisplayText != "where" || item.MatchText != "where" || item.OrderText != "where" ||
		item.BeforeText() != "where" || item.AfterText() != "" || item.Rank() != kustolang.CompletionRankDefault ||
		item.Priority() != kustolang.CompletionPriorityNormal || item.Retrigger {
		t.Fatalf("default completion item = %#v rank:%d priority:%d", item, item.Rank(), item.Priority())
	}
}

func TestPublicCompletionItemOptionsAndDefensiveCopies(t *testing.T) {
	apply := []kustolang.CompletionText{
		kustolang.NewCompletionText("fn(", false, false),
		kustolang.NewCompletionTextEdit("value"),
		kustolang.NewCompletionText(")", false, false),
	}
	item := kustolang.NewCompletionItem(
		kustolang.CompletionKindBuiltInFunction,
		"fn",
		kustolang.WithCompletionMatchText("function"),
		kustolang.WithCompletionOrderText("001_fn"),
		kustolang.WithCompletionApplyTexts(apply...),
		kustolang.WithCompletionRank(kustolang.CompletionRankFunction),
		kustolang.WithCompletionPriority(kustolang.CompletionPriorityHigh),
		kustolang.WithCompletionRetrigger(true),
	)
	apply[0].Text = "mutated"
	if item.MatchText != "function" || item.OrderText != "001_fn" || item.BeforeText() != "fn(" || item.AfterText() != "value)" ||
		item.Rank() != kustolang.CompletionRankFunction || item.Priority() != kustolang.CompletionPriorityHigh || !item.Retrigger {
		t.Fatalf("completion item = %#v rank:%d priority:%d", item, item.Rank(), item.Priority())
	}

	items := []kustolang.CompletionItem{item}
	info := kustolang.NewCompletionInfo(items, 7, 2)
	items[0] = kustolang.NewSyntaxCompletionItem("changed")
	if len(info.Items) != 1 || info.Items[0].DisplayText != "fn" || info.EditStart != 7 || info.EditLength != 2 {
		t.Fatalf("completion info = %#v", info)
	}
	empty := kustolang.CompletionInfoEmpty
	if len(empty.Items) != 0 || empty.EditStart != 0 || empty.EditLength != 0 {
		t.Fatalf("empty completion info = %#v", empty)
	}
}

func TestPublicCompletionBeforeAfterAndMarkerOptions(t *testing.T) {
	item := kustolang.NewCompletionItem(
		kustolang.CompletionKindSyntax,
		"display",
		kustolang.WithCompletionBeforeText("before"),
		kustolang.WithCompletionAfterText("after"),
	)
	if item.BeforeText() != "before" || item.AfterText() != "after" || !item.HasCaret() {
		t.Fatalf("before/after item = %#v", item)
	}
	marked := kustolang.NewSyntaxCompletionItem("call", kustolang.WithCompletionApplyTextMarkers(`call([[arg]])`))
	if marked.BeforeText() != "call(" || marked.AfterText() != "arg)" || marked.EditText() != "call(" {
		t.Fatalf("marked item = %#v", marked)
	}
}
