package kustolang

import (
	"strings"
	"testing"
	"unicode/utf16"
)

func TestKustoCompleterShouldAutoCompleteTokenEdges(t *testing.T) {
	tests := []struct {
		name string
		text string
		key  rune
		want bool
	}{
		{
			name: "identifier edge",
			text: "StormEvents",
			key:  's',
			want: true,
		},
		{
			name: "backspace into identifier",
			text: "StormEvents",
			key:  '\b',
			want: true,
		},
		{
			name: "dot edge",
			text: "StormEvents.",
			key:  '.',
			want: true,
		},
		{
			name: "open paren edge",
			text: "print x=tolower(",
			key:  '(',
			want: true,
		},
		{
			name: "bang edge",
			text: "!",
			key:  '!',
			want: true,
		},
		{
			name: "control key",
			text: "StormEvents",
			key:  '\t',
			want: false,
		},
		{
			name: "newline",
			text: "StormEvents",
			key:  '\n',
			want: false,
		},
		{
			name: "null key",
			text: "StormEvents",
			key:  '\x00',
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKustoCodeService(tt.text).ShouldAutoComplete(len(tt.text), tt.key, CancellationToken{}); got != tt.want {
				t.Fatalf("ShouldAutoComplete(%q, %q) = %v, want %v", tt.text, string(tt.key), got, tt.want)
			}
		})
	}
}

func TestKustoCompleterGetCompletionItemsEditRange(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		position   int
		wantStart  int
		wantLength int
	}{
		{
			name:       "identifier prefix at end",
			text:       "StormEvents | where Sta",
			position:   len("StormEvents | where Sta"),
			wantStart:  strings.Index("StormEvents | where Sta", "Sta"),
			wantLength: len("Sta"),
		},
		{
			name:       "inside identifier",
			text:       "StormEvents | where State",
			position:   strings.Index("StormEvents | where State", "State") + len("St"),
			wantStart:  strings.Index("StormEvents | where State", "State"),
			wantLength: len("State"),
		},
		{
			name:       "whitespace between tokens",
			text:       "StormEvents | ",
			position:   len("StormEvents | "),
			wantStart:  len("StormEvents | "),
			wantLength: 0,
		},
		{
			name:       "dash joined known prefix",
			text:       "StormEvents | project-ke",
			position:   len("StormEvents | project-ke"),
			wantStart:  strings.Index("StormEvents | project-ke", "project"),
			wantLength: len("project-ke"),
		},
		{
			name:       "dash joined unknown prefix",
			text:       "StormEvents | foo-ba",
			position:   len("StormEvents | foo-ba"),
			wantStart:  strings.Index("StormEvents | foo-ba", "foo"),
			wantLength: len("foo"),
		},
		{
			name:       "literal range with suppressed items",
			text:       `print x="abc"`,
			position:   strings.Index(`print x="abc"`, "b"),
			wantStart:  strings.Index(`print x="abc"`, `"`),
			wantLength: len(`"abc"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := CompletionOptionsDefault.WithIncludeSyntax(false).WithIncludeSymbols(false)
			info := NewKustoCodeService(tt.text).GetCompletionItems(tt.position, options, CancellationToken{})
			if len(info.Items) != 0 {
				t.Fatalf("GetCompletionItems(%q) items = %#v, want none with syntax disabled", tt.text, info.Items)
			}
			if info.EditStart != tt.wantStart || info.EditLength != tt.wantLength {
				t.Fatalf("GetCompletionItems(%q) edit = %d/%d, want %d/%d", tt.text, info.EditStart, info.EditLength, tt.wantStart, tt.wantLength)
			}
		})
	}
}

func TestKustoCodeServiceCompletionUsesUTF16EditRange(t *testing.T) {
	text := "// 😀\nStormEvents | where Sta"
	position := len(utf16.Encode([]rune(text)))
	identifierStart := strings.LastIndex(text, "Sta")
	wantStart := len(utf16.Encode([]rune(text[:identifierStart])))

	options := CompletionOptionsDefault.WithIncludeSyntax(false).WithIncludeSymbols(false)
	info := NewKustoCodeService(text).GetCompletionItems(position, options, CancellationToken{})
	if info.EditStart != wantStart || info.EditLength != 3 {
		t.Fatalf("UTF-16 completion edit = %d/%d, want %d/3", info.EditStart, info.EditLength, wantStart)
	}
}

func TestKustoCompleterQueryOperatorSyntaxCompletions(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantItems  []string
		wantAbsent []string
	}{
		{
			name:       "piped operator position",
			text:       "StormEvents | ",
			wantItems:  []string{"where", "project", "summarize", "union"},
			wantAbsent: []string{"print", "find", "macro-expand", "range", "graph-where-nodes", "graph-where-edges"},
		},
		{
			name:       "piped partial filters",
			text:       "StormEvents | pro",
			wantItems:  []string{"project", "project-away", "project-keep"},
			wantAbsent: []string{"where", "summarize"},
		},
		{
			name:       "piped dash partial filters",
			text:       "StormEvents | mv-",
			wantItems:  []string{"mv-apply", "mv-expand"},
			wantAbsent: []string{"mvapply", "mvexpand", "where"},
		},
		{
			name:       "operator body suppresses operator syntax",
			text:       "StormEvents | where ",
			wantAbsent: []string{"where", "project", "summarize"},
		},
		{
			name:       "root partial filters to root operators",
			text:       "pri",
			wantItems:  []string{"print"},
			wantAbsent: []string{"project", "where"},
		},
		{
			name:      "nested malformed pipe keeps query context",
			text:      "print x=(T | ",
			wantItems: []string{"where", "project", "summarize"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewKustoCodeService(tt.text).GetCompletionItems(len(tt.text), CompletionOptionsDefault, CancellationToken{})
			for _, want := range tt.wantItems {
				item, ok := completionItemByDisplayText(info.Items, want)
				if !ok {
					t.Fatalf("completion items for %q missing %q: %#v", tt.text, want, completionDisplayTexts(info.Items))
				}
				if item.Kind != CompletionKindQueryPrefix {
					t.Fatalf("completion %q kind = %v, want query prefix", want, item.Kind)
				}
				if !strings.HasSuffix(item.BeforeText(), " ") || !item.Retrigger {
					t.Fatalf("completion %q apply/retrigger = before:%q retrigger:%t", want, item.BeforeText(), item.Retrigger)
				}
			}
			for _, absent := range tt.wantAbsent {
				if _, ok := completionItemByDisplayText(info.Items, absent); ok {
					t.Fatalf("completion items for %q unexpectedly included %q: %#v", tt.text, absent, completionDisplayTexts(info.Items))
				}
			}
		})
	}
}

func TestKustoCompleterCommandSyntaxCompletions(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantItems  []string
		wantAbsent []string
	}{
		{
			name:      "partial first command token",
			text:      ".sh",
			wantItems: []string{"show"},
		},
		{
			name:       "root command hides hidden aliases",
			text:       ".",
			wantItems:  []string{"attach", "detach"},
			wantAbsent: []string{"load"},
		},
		{
			name:      "after show command token",
			text:      ".show ",
			wantItems: []string{"database", "databases", "version"},
		},
		{
			name:       "malformed partial command keeps suggesting matching syntax",
			text:       ".show dat",
			wantItems:  []string{"database", "databases"},
			wantAbsent: []string{"version"},
		},
		{
			name:       "pipe tail suppresses command syntax",
			text:       ".show databases | ",
			wantAbsent: []string{"database", "version"},
		},
		{
			name:       "attach database hides hidden selector alternatives",
			text:       ".attach database ",
			wantAbsent: []string{"all", "metadata"},
		},
		{
			name:       "external table kind hides hidden storage aliases",
			text:       ".create external table Ext (Raw:string) kind = ",
			wantItems:  []string{"storage"},
			wantAbsent: []string{"blob", "adl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewKustoCodeService(tt.text).GetCompletionItems(len(tt.text), CompletionOptionsDefault, CancellationToken{})
			for _, want := range tt.wantItems {
				item, ok := completionItemByDisplayText(info.Items, want)
				if !ok {
					t.Fatalf("completion items for %q missing %q: %#v", tt.text, want, completionDisplayTexts(info.Items))
				}
				if item.Kind != CompletionKindCommandPrefix {
					t.Fatalf("completion %q kind = %v, want command prefix", want, item.Kind)
				}
				if !strings.HasSuffix(item.BeforeText(), " ") || !item.Retrigger {
					t.Fatalf("completion %q apply/retrigger = before:%q retrigger:%t", want, item.BeforeText(), item.Retrigger)
				}
			}
			for _, absent := range tt.wantAbsent {
				if _, ok := completionItemByDisplayText(info.Items, absent); ok {
					t.Fatalf("completion items for %q unexpectedly included %q: %#v", tt.text, absent, completionDisplayTexts(info.Items))
				}
			}
		})
	}
}

func TestKustoCompleterCommandSyntaxRespectsOptions(t *testing.T) {
	options := CompletionOptionsDefault.WithIncludeSyntax(false)
	info := NewKustoCodeService(".show ").GetCompletionItems(len(".show "), options, CancellationToken{})
	if len(info.Items) != 0 {
		t.Fatalf("completion items with syntax disabled = %#v, want none", info.Items)
	}

	options = CompletionOptionsDefault.WithAutoAppendWhitespace(false)
	info = NewKustoCodeService(".sh").GetCompletionItems(len(".sh"), options, CancellationToken{})
	show, ok := completionItemByDisplayText(info.Items, "show")
	if !ok {
		t.Fatalf("completion items missing show: %#v", completionDisplayTexts(info.Items))
	}
	if show.BeforeText() != "show" || show.Retrigger {
		t.Fatalf("show completion with auto whitespace disabled = before:%q retrigger:%t", show.BeforeText(), show.Retrigger)
	}
}

func TestKustoCompleterQuerySyntaxRespectsOptions(t *testing.T) {
	options := CompletionOptionsDefault.WithIncludeSyntax(false)
	info := NewKustoCodeService("StormEvents | ").GetCompletionItems(len("StormEvents | "), options, CancellationToken{})
	if len(info.Items) != 0 {
		t.Fatalf("query completion items with syntax disabled = %#v, want none", info.Items)
	}

	options = CompletionOptionsDefault.WithAutoAppendWhitespace(false)
	info = NewKustoCodeService("StormEvents | wh").GetCompletionItems(len("StormEvents | wh"), options, CancellationToken{})
	where, ok := completionItemByDisplayText(info.Items, "where")
	if !ok {
		t.Fatalf("completion items missing where: %#v", completionDisplayTexts(info.Items))
	}
	if where.BeforeText() != "where" || where.Retrigger {
		t.Fatalf("where completion with auto whitespace disabled = before:%q retrigger:%t", where.BeforeText(), where.Retrigger)
	}
}

func TestKustoCompleterCommandTableSymbolCompletions(t *testing.T) {
	signinLogs := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	securityEvent := NewTableSymbol("SecurityEvent", NewColumnSymbol("Computer", ScalarTypes.String.Type()))
	archive := NewExternalTableSymbol("ArchiveEvents", NewColumnSymbol("Raw", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", signinLogs, securityEvent, archive))

	info := NewKustoCodeService(".show table ", globals).GetCompletionItems(len(".show table "), CompletionOptionsDefault, CancellationToken{})
	for _, want := range []string{"SigninLogs", "SecurityEvent"} {
		item, ok := completionItemByDisplayText(info.Items, want)
		if !ok {
			t.Fatalf(".show table completions missing %s: %#v", want, completionDisplayTexts(info.Items))
		}
		if item.Kind != CompletionKindTable {
			t.Fatalf("%s completion kind = %v, want table", want, item.Kind)
		}
	}
	if _, ok := completionItemByDisplayText(info.Items, "ArchiveEvents"); ok {
		t.Fatalf(".show table completions unexpectedly included external table: %#v", completionDisplayTexts(info.Items))
	}

	partial := NewKustoCodeService(".show table Sign", globals).GetCompletionItems(len(".show table Sign"), CompletionOptionsDefault, CancellationToken{})
	if partial.EditStart != len(".show table ") || partial.EditLength != len("Sign") {
		t.Fatalf("partial edit range = %d/%d", partial.EditStart, partial.EditLength)
	}
	if _, ok := completionItemByDisplayText(partial.Items, "SigninLogs"); !ok {
		t.Fatalf("partial table completions missing SigninLogs: %#v", completionDisplayTexts(partial.Items))
	}
	if _, ok := completionItemByDisplayText(partial.Items, "SecurityEvent"); ok {
		t.Fatalf("partial table completions unexpectedly included SecurityEvent: %#v", completionDisplayTexts(partial.Items))
	}
}

func TestKustoCompleterCommandExternalTableSymbolCompletions(t *testing.T) {
	signinLogs := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	archive := NewExternalTableSymbol("ArchiveEvents", NewColumnSymbol("Raw", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", signinLogs, archive))

	info := NewKustoCodeService(".show external table ", globals).GetCompletionItems(len(".show external table "), CompletionOptionsDefault, CancellationToken{})
	item, ok := completionItemByDisplayText(info.Items, "ArchiveEvents")
	if !ok {
		t.Fatalf(".show external table completions missing ArchiveEvents: %#v", completionDisplayTexts(info.Items))
	}
	if item.Kind != CompletionKindTable {
		t.Fatalf("ArchiveEvents completion kind = %v, want table", item.Kind)
	}
	if _, ok := completionItemByDisplayText(info.Items, "SigninLogs"); ok {
		t.Fatalf(".show external table completions unexpectedly included normal table: %#v", completionDisplayTexts(info.Items))
	}
}

func TestKustoCompleterCommandMaterializedViewSymbolCompletions(t *testing.T) {
	signinLogs := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	daily := NewMaterializedViewSymbol("DailySignin", []*ColumnSymbol{NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type())}, "SigninLogs | summarize count()")
	globals := NewGlobalState(NewDatabaseSymbol("db", signinLogs, daily))

	info := NewKustoCodeService(".show materialized-view ", globals).GetCompletionItems(len(".show materialized-view "), CompletionOptionsDefault, CancellationToken{})
	item, ok := completionItemByDisplayText(info.Items, "DailySignin")
	if !ok {
		t.Fatalf(".show materialized-view completions missing DailySignin: %#v", completionDisplayTexts(info.Items))
	}
	if item.Kind != CompletionKindMaterialiedView {
		t.Fatalf("DailySignin completion kind = %v, want materialized view", item.Kind)
	}
	if _, ok := completionItemByDisplayText(info.Items, "SigninLogs"); ok {
		t.Fatalf(".show materialized-view completions unexpectedly included normal table: %#v", completionDisplayTexts(info.Items))
	}
}

func TestKustoCompleterCommandSymbolCompletionsRespectOptions(t *testing.T) {
	signinLogs := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", signinLogs))

	syntaxDisabled := CompletionOptionsDefault.WithIncludeSyntax(false)
	info := NewKustoCodeService(".show table ", globals).GetCompletionItems(len(".show table "), syntaxDisabled, CancellationToken{})
	if _, ok := completionItemByDisplayText(info.Items, "SigninLogs"); !ok {
		t.Fatalf("syntax-disabled command symbol completions missing SigninLogs: %#v", completionDisplayTexts(info.Items))
	}

	symbolsDisabled := CompletionOptionsDefault.WithIncludeSymbols(false)
	info = NewKustoCodeService(".show table ", globals).GetCompletionItems(len(".show table "), symbolsDisabled, CancellationToken{})
	if _, ok := completionItemByDisplayText(info.Items, "SigninLogs"); ok {
		t.Fatalf("symbol-disabled command completions included SigninLogs: %#v", completionDisplayTexts(info.Items))
	}
}

func TestKustoCompletionBuilderDeduplicatesAndPromotesQueryPrefix(t *testing.T) {
	builder := newKustoCompletionBuilder()
	keyword := NewCompletionItem(CompletionKindKeyword, "where", WithCompletionOrderText("keyword"))
	queryPrefix := NewCompletionItem(CompletionKindQueryPrefix, "where", WithCompletionOrderText("query"))
	column := NewCompletionItem(CompletionKindColumn, "Account")

	builder.add(keyword)
	builder.add(column)
	builder.add(queryPrefix)
	builder.add(NewCompletionItem(CompletionKindKeyword, "Account"))

	items := builder.items()
	if builder.count() != 2 || len(items) != 2 {
		t.Fatalf("builder count/items = %d/%#v, want two unique display texts", builder.count(), items)
	}
	if items[0].Kind != CompletionKindQueryPrefix || items[0].OrderText != "query" {
		t.Fatalf("where item = %#v, want promoted query prefix", items[0])
	}
	if items[1].Kind != CompletionKindColumn || items[1].DisplayText != "Account" {
		t.Fatalf("Account item = %#v, want original column", items[1])
	}
}

func completionItemByDisplayText(items []CompletionItem, displayText string) (CompletionItem, bool) {
	for _, item := range items {
		if item.DisplayText == displayText {
			return item, true
		}
	}
	return CompletionItem{}, false
}

func completionDisplayTexts(items []CompletionItem) []string {
	texts := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, item.DisplayText)
	}
	return texts
}

func TestKustoCompleterOrdersCompletionItemsAndRetriggers(t *testing.T) {
	completer := newKustoCompleter(Parse("StormEvents | "), CompletionOptionsDefault, CancellationToken{})
	items := []CompletionItem{
		NewCompletionItem(CompletionKindKeyword, "where", WithCompletionOrderText("Where")),
		NewCompletionItem(CompletionKindColumn, "Account", WithCompletionPriority(CompletionPriorityHigh)),
		NewCompletionItem(CompletionKindTable, "StormEvents"),
		NewCompletionItem(CompletionKindBuiltInFunction, "count", WithCompletionBeforeText("count(")),
	}

	ordered := completer.orderCompletionItems(items)
	if len(ordered) != 4 {
		t.Fatalf("ordered length = %d", len(ordered))
	}
	if ordered[0].Kind != CompletionKindColumn || ordered[0].OrderText != "032_account" {
		t.Fatalf("first ordered item = %#v, want high-priority column", ordered[0])
	}
	if ordered[1].Kind != CompletionKindTable || ordered[1].OrderText != "043_stormevents" {
		t.Fatalf("second ordered item = %#v, want table", ordered[1])
	}
	if ordered[2].Kind != CompletionKindBuiltInFunction || ordered[2].OrderText != "073_count" || !ordered[2].Retrigger {
		t.Fatalf("third ordered item = %#v, want retriggered function", ordered[2])
	}
	if ordered[3].Kind != CompletionKindKeyword || ordered[3].OrderText != "083_where" {
		t.Fatalf("fourth ordered item = %#v, want keyword", ordered[3])
	}
}

func TestKustoCompleterOrderingRankMatchesUpstreamCases(t *testing.T) {
	completer := newKustoCompleter(Parse("StormEvents"), CompletionOptionsDefault, CancellationToken{})
	tests := []struct {
		item CompletionItem
		want CompletionRank
	}{
		{NewCompletionItem(CompletionKindExample, "1"), CompletionRankLiteral},
		{NewCompletionItem(CompletionKindQueryPrefix, "where"), CompletionRankKeyword},
		{NewCompletionItem(CompletionKindAggregateFunction, "count"), CompletionRankAggregate},
		{NewCompletionItem(CompletionKindColumn, "Account"), CompletionRankColumn},
		{NewCompletionItem(CompletionKindTable, "StormEvents"), CompletionRankTable},
		{NewCompletionItem(CompletionKindMaterialiedView, "View"), CompletionRankEntity},
		{NewCompletionItem(CompletionKindVariable, "$left"), CompletionRankLiteral},
		{NewCompletionItem(CompletionKindParameter, "p"), CompletionRankVariable},
		{NewCompletionItem(CompletionKindBuiltInFunction, "strlen"), CompletionRankFunction},
		{NewCompletionItem(CompletionKindScalarInfix, "has"), CompletionRankStringOperator},
		{NewCompletionItem(CompletionKindScalarInfix, "+"), CompletionRankMathOperator},
		{NewCompletionItem(CompletionKindSyntax, "("), CompletionRankOther},
		{NewCompletionItem(CompletionKindColumn, "Forced", WithCompletionRank(CompletionRankLiteral)), CompletionRankLiteral},
	}

	for _, tt := range tests {
		if got := completer.getOrderingRank(tt.item); got != tt.want {
			t.Fatalf("rank for %#v = %v, want %v", tt.item, got, tt.want)
		}
	}
}

func TestKustoCompleterShouldAutoCompleteAfterWhitespace(t *testing.T) {
	tests := []string{
		"StormEvents | ",
		"StormEvents | where ",
		"print x=(1) ",
		"print x= ",
	}

	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			if got := NewKustoCodeService(text).ShouldAutoComplete(len(text), ' ', CancellationToken{}); !got {
				t.Fatalf("ShouldAutoComplete(%q, space) = false, want true", text)
			}
		})
	}
}

func TestKustoCompleterShouldAutoCompleteRejectsLiteralDirectiveAndComment(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		position int
		key      rune
	}{
		{
			name:     "inside string literal",
			text:     `print x="abc"`,
			position: strings.Index(`print x="abc"`, "b"),
			key:      'b',
		},
		{
			name:     "unterminated string literal",
			text:     `print x="abc`,
			position: strings.Index(`print x="abc`, "b"),
			key:      'b',
		},
		{
			name:     "inside directive",
			text:     "#set querytrace=true\nStormEvents",
			position: strings.Index("#set querytrace=true\nStormEvents", "querytrace"),
			key:      'q',
		},
		{
			name:     "inside line comment",
			text:     "// comment\nStormEvents",
			position: strings.Index("// comment\nStormEvents", "comment"),
			key:      'c',
		},
		{
			name:     "inside comment before malformed query",
			text:     "// comment\nStormEvents | where == 1",
			position: strings.Index("// comment\nStormEvents | where == 1", "comment"),
			key:      'c',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.position < 0 {
				t.Fatalf("test position not found in %q", tt.text)
			}
			if got := NewKustoCodeService(tt.text).ShouldAutoComplete(tt.position, tt.key, CancellationToken{}); got {
				t.Fatalf("ShouldAutoComplete(%q at %d) = true, want false", tt.text, tt.position)
			}
		})
	}
}

func TestKustoCompleterAllowsCompletionAtCommentBoundaries(t *testing.T) {
	trivia := "// comment\n"
	if KustoCompleterIsInsideComment(trivia, 0) {
		t.Fatal("comment start reported as inside comment")
	}
	if !KustoCompleterIsInsideComment(trivia, 4) {
		t.Fatal("comment body reported as outside comment")
	}
	if KustoCompleterIsInsideComment(trivia, len(trivia)) {
		t.Fatal("position after line break reported as inside comment")
	}
}
