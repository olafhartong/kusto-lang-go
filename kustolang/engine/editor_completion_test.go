package kustolang

import "testing"

func TestEditorEnumNumericOrderMatchesUpstream(t *testing.T) {
	if int(BrackettingStyleDefault) != 0 || int(BrackettingStyleDiagonal) != 3 {
		t.Fatalf("bracketting style order changed: default=%d diagonal=%d", BrackettingStyleDefault, BrackettingStyleDiagonal)
	}
	if int(PlacementStyleNone) != 0 || int(PlacementStyleSmart) != 2 {
		t.Fatalf("placement style order changed: none=%d smart=%d", PlacementStyleNone, PlacementStyleSmart)
	}
	if int(SpacingStyleAsIs) != 0 || int(SpacingStyleOne) != 2 {
		t.Fatalf("spacing style order changed: as-is=%d one=%d", SpacingStyleAsIs, SpacingStyleOne)
	}
	if int(DualSpacingStyleAsIs) != 0 || int(DualSpacingStyleBoth) != 4 {
		t.Fatalf("dual spacing style order changed: as-is=%d both=%d", DualSpacingStyleAsIs, DualSpacingStyleBoth)
	}
	if int(PositionBiasLeft) != 0 || int(PositionBiasRight) != 1 {
		t.Fatalf("position bias order changed: left=%d right=%d", PositionBiasLeft, PositionBiasRight)
	}
	if int(MinimalTextKindRemoveLeadingWhitespaceAndComments) != 0 || int(MinimalTextKindSingleLine) != 2 {
		t.Fatalf("minimal text kind order changed: leading=%d single-line=%d", MinimalTextKindRemoveLeadingWhitespaceAndComments, MinimalTextKindSingleLine)
	}
}

func TestCompletionKindAndPriorityNumericValuesMatchUpstream(t *testing.T) {
	cases := []struct {
		kind  CompletionKind
		value int
	}{
		{CompletionKindUnknown, 0},
		{CompletionKindKeyword, 1},
		{CompletionKindPunctuation, 2},
		{CompletionKindSyntax, 3},
		{CompletionKindColumn, 13},
		{CompletionKindCluster, 22},
		{CompletionKindMaterialiedView, 23},
		{CompletionKindStoredQueryResult, 28},
	}
	for _, test := range cases {
		if got := int(test.kind); got != test.value {
			t.Fatalf("completion kind numeric value = %d, want %d", got, test.value)
		}
	}
	if int(CompletionRankLiteral) != 1 || int(CompletionRankDefault) != 12 {
		t.Fatalf("completion rank order changed: literal=%d default=%d", CompletionRankLiteral, CompletionRankDefault)
	}
	if int(CompletionPriorityTop) != 1 || int(CompletionPriorityDefault) != 3 || int(CompletionPriorityLow) != 4 {
		t.Fatalf("completion priority order changed: top=%d default=%d low=%d", CompletionPriorityTop, CompletionPriorityDefault, CompletionPriorityLow)
	}
}

func TestCompletionTextParseCursorMarkerTakesPrecedence(t *testing.T) {
	texts := ParseCompletionText("prefix|suffix[[ignored]]")
	if len(texts) != 2 {
		t.Fatalf("texts length = %d, want 2: %#v", len(texts), texts)
	}
	if texts[0] != NewCompletionText("prefix", false, false) {
		t.Fatalf("before cursor = %#v", texts[0])
	}
	if texts[1] != NewCompletionText("suffix[[ignored]]", true, false) {
		t.Fatalf("after cursor = %#v", texts[1])
	}
}

func TestCompletionTextParseEditableSegments(t *testing.T) {
	texts := ParseCompletionText("let [[name]] = value")
	if len(texts) != 3 {
		t.Fatalf("texts length = %d, want 3: %#v", len(texts), texts)
	}
	if texts[0].Text != "let " || texts[0].Caret || texts[0].Select {
		t.Fatalf("prefix segment = %#v", texts[0])
	}
	if texts[1] != NewCompletionTextEdit("name") {
		t.Fatalf("edit segment = %#v", texts[1])
	}
	if texts[2].Text != " = value" || texts[2].Caret || texts[2].Select {
		t.Fatalf("suffix segment = %#v", texts[2])
	}

	unclosed := ParseCompletionText("let [[name")
	if len(unclosed) != 1 || unclosed[0].Text != "let [[name" {
		t.Fatalf("unclosed edit marker = %#v", unclosed)
	}

	empty := ParseCompletionText("")
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty parse = %#v, want non-nil empty", empty)
	}
}

func TestCompletionItemDefaultsAndCaretText(t *testing.T) {
	item := NewCompletionItem(CompletionKindKeyword, "where", WithCompletionAfterText(" "))
	if item.Kind != CompletionKindKeyword || item.DisplayText != "where" || item.MatchText != "where" || item.OrderText != "where" {
		t.Fatalf("completion item metadata = %#v", item)
	}
	if item.rank != CompletionRankDefault || item.priority != CompletionPriorityNormal || item.Retrigger {
		t.Fatalf("completion item ordering/retrigger = rank:%d priority:%d retrigger:%t", item.rank, item.priority, item.Retrigger)
	}
	if got := item.BeforeText(); got != "where" {
		t.Fatalf("BeforeText = %q, want where", got)
	}
	if got := item.AfterText(); got != " " {
		t.Fatalf("AfterText = %q, want space", got)
	}
	if !item.HasCaret() || item.CaretIndex() != 1 {
		t.Fatalf("caret = has:%t index:%d", item.HasCaret(), item.CaretIndex())
	}

	item.ApplyTexts[0].Text = "mutated"
	if got := item.BeforeText(); got != "mutated" {
		t.Fatalf("local item mutation should be visible to receiver copy, got %q", got)
	}
}

func TestCompletionItemCloneHelpersMirrorApplyTextRules(t *testing.T) {
	item := NewSyntaxCompletionItem("call", WithCompletionApplyTextMarkers("call([[arg]])"))
	if got, want := item.BeforeText(), "call("; got != want {
		t.Fatalf("BeforeText = %q, want %q", got, want)
	}
	if got, want := item.AfterText(), "arg)"; got != want {
		t.Fatalf("AfterText = %q, want %q", got, want)
	}

	withBefore := item.WithBeforeText("invoke(")
	if got, want := withBefore.BeforeText(), "invoke("; got != want {
		t.Fatalf("WithBeforeText BeforeText = %q, want %q", got, want)
	}
	if got, want := withBefore.AfterText(), "arg)"; got != want {
		t.Fatalf("WithBeforeText AfterText = %q, want %q", got, want)
	}

	withAfter := item.WithAfterText(")")
	if got, want := withAfter.BeforeText(), "call("; got != want {
		t.Fatalf("WithAfterText BeforeText = %q, want %q", got, want)
	}
	if got, want := withAfter.AfterText(), ")"; got != want {
		t.Fatalf("WithAfterText AfterText = %q, want %q", got, want)
	}

	noCaret := NewCompletionItem(CompletionKindIdentifier, "Account").WithAfterText("Suffix")
	if got, want := noCaret.BeforeText(), "Account"; got != want {
		t.Fatalf("no-caret BeforeText = %q, want %q", got, want)
	}
	if got, want := noCaret.AfterText(), "Suffix"; got != want {
		t.Fatalf("no-caret AfterText = %q, want %q", got, want)
	}

	clone := item.
		WithKind(CompletionKindBuiltInFunction).
		WithDisplayText("display").
		WithMatchText("match").
		WithOrderText("order").
		WithRank(CompletionRankFunction).
		WithPriority(CompletionPriorityHigh).
		WithRetrigger(true)
	if clone.Kind != CompletionKindBuiltInFunction || clone.DisplayText != "display" || clone.MatchText != "match" || clone.OrderText != "order" {
		t.Fatalf("clone metadata = %#v", clone)
	}
	if clone.rank != CompletionRankFunction || clone.priority != CompletionPriorityHigh || !clone.Retrigger {
		t.Fatalf("clone ordering/retrigger = rank:%d priority:%d retrigger:%t", clone.rank, clone.priority, clone.Retrigger)
	}
	if item.DisplayText != "call" || item.Retrigger {
		t.Fatalf("original item mutated: %#v", item)
	}
}

func TestCompletionItemApplyTextInputsAreDefensivelyCopied(t *testing.T) {
	applyTexts := []CompletionText{
		NewCompletionText("call(", false, false),
		NewCompletionText("arg)", true, false),
	}
	option := WithCompletionApplyTexts(applyTexts...)
	applyTexts[0].Text = "changed before construction"

	first := NewSyntaxCompletionItem("call", option)
	second := NewSyntaxCompletionItem("call", option)
	if got, want := first.BeforeText(), "call("; got != want {
		t.Fatalf("first BeforeText = %q, want %q", got, want)
	}
	if got, want := second.BeforeText(), "call("; got != want {
		t.Fatalf("second BeforeText = %q, want %q", got, want)
	}

	first.ApplyTexts[0].Text = "changed first"
	if got, want := second.BeforeText(), "call("; got != want {
		t.Fatalf("second BeforeText after first mutation = %q, want %q", got, want)
	}

	replacement := []CompletionText{
		NewCompletionText("invoke(", false, false),
		NewCompletionText(")", true, false),
	}
	withApplyTexts := second.WithApplyTexts(replacement...)
	replacement[0].Text = "changed replacement"
	if got, want := withApplyTexts.BeforeText(), "invoke("; got != want {
		t.Fatalf("WithApplyTexts BeforeText = %q, want %q", got, want)
	}
}

func TestCompletionItemWithMethodsIsolateApplyTexts(t *testing.T) {
	tests := []struct {
		name string
		with func(CompletionItem) CompletionItem
	}{
		{"kind", func(item CompletionItem) CompletionItem { return item.WithKind(CompletionKindKeyword) }},
		{"display text", func(item CompletionItem) CompletionItem { return item.WithDisplayText("display") }},
		{"match text", func(item CompletionItem) CompletionItem { return item.WithMatchText("match") }},
		{"order text", func(item CompletionItem) CompletionItem { return item.WithOrderText("order") }},
		{"apply texts", func(item CompletionItem) CompletionItem {
			return item.WithApplyTexts(
				NewCompletionText("apply", false, false),
				NewCompletionText(" tail", true, false),
			)
		}},
		{"apply text markers", func(item CompletionItem) CompletionItem { return item.WithApplyTextMarkers("marked| tail") }},
		{"before text", func(item CompletionItem) CompletionItem { return item.WithBeforeText("before") }},
		{"edit text", func(item CompletionItem) CompletionItem { return item.WithEditText("edit") }},
		{"after text", func(item CompletionItem) CompletionItem { return item.WithAfterText("after") }},
		{"rank", func(item CompletionItem) CompletionItem { return item.WithRank(CompletionRankKeyword) }},
		{"priority", func(item CompletionItem) CompletionItem { return item.WithPriority(CompletionPriorityHigh) }},
		{"retrigger", func(item CompletionItem) CompletionItem { return item.WithRetrigger(true) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := NewSyntaxCompletionItem("source", WithCompletionApplyTextMarkers("source| tail"))
			result := test.with(source)
			resultTail := result.ApplyTexts[1].Text

			result.ApplyTexts[0].Text = "changed result"
			if got, want := source.ApplyTexts[0].Text, "source"; got != want {
				t.Fatalf("source apply text after result mutation = %q, want %q", got, want)
			}

			source.ApplyTexts[1].Text = "changed source"
			if got := result.ApplyTexts[1].Text; got != resultTail {
				t.Fatalf("result apply text after source mutation = %q, want %q", got, resultTail)
			}
		})
	}
}

func TestCompletionInfoAndOptionsDefaults(t *testing.T) {
	items := []CompletionItem{NewSyntaxCompletionItem("where", WithCompletionApplyTextMarkers("where| "))}
	info := NewCompletionInfo(items, 3, 5)
	items[0].ApplyTexts[0].Text = "mutated source text"
	if got, want := info.Items[0].ApplyTexts[0].Text, "where"; got != want {
		t.Fatalf("completion info nested source mutation = %q, want %q", got, want)
	}
	info.Items[0].ApplyTexts[1].Text = "mutated result text"
	if got, want := items[0].ApplyTexts[1].Text, " "; got != want {
		t.Fatalf("source nested result mutation = %q, want %q", got, want)
	}
	items[0] = NewSyntaxCompletionItem("project")
	if len(info.Items) != 1 || info.Items[0].DisplayText != "where" || info.EditStart != 3 || info.EditLength != 5 {
		t.Fatalf("completion info = %#v", info)
	}
	if CompletionInfoEmpty.Items == nil || len(CompletionInfoEmpty.Items) != 0 {
		t.Fatalf("empty completion info items = %#v, want non-nil empty", CompletionInfoEmpty.Items)
	}

	options := CompletionOptionsDefault
	if !options.IncludeSymbols() || !options.IncludeSyntax() || !options.IncludePunctuationOnlySyntax() || !options.AutoAppendWhitespace() {
		t.Fatalf("default completion options missing expected flags: %#v", options)
	}
	if options.EnableParameterInjection() || options.IncludeExtendedSyntax() || options.IncludeFunctions != IncludeFunctionKindAll {
		t.Fatalf("default completion options unexpected flags: %#v", options)
	}

	changed := options.
		WithIncludeSymbols(false).
		WithIncludeSyntax(false).
		WithIncludePunctuationOnlySyntax(false).
		WithAutoAppendWhitespace(false).
		WithEnableParameterInjection(true).
		WithIncludeExtendedSyntax(true).
		WithIncludeFunctionsEnabled(false)
	if changed.IncludeSymbols() || changed.IncludeSyntax() || changed.IncludePunctuationOnlySyntax() || changed.AutoAppendWhitespace() {
		t.Fatalf("changed completion options should have default flags disabled: %#v", changed)
	}
	if !changed.EnableParameterInjection() || !changed.IncludeExtendedSyntax() || changed.IncludeFunctions != IncludeFunctionKindNone {
		t.Fatalf("changed completion options missing enabled flags: %#v", changed)
	}
	if !options.IncludeSymbols() || options.IncludeFunctions != IncludeFunctionKindAll {
		t.Fatalf("default options mutated: %#v", options)
	}
}

func TestTextEditConstructorsAndValidation(t *testing.T) {
	replacement := NewTextReplacement(2, 3, "abc")
	if replacement.Start != 2 || replacement.DeleteLength != 3 || replacement.InsertText != "abc" {
		t.Fatalf("replacement = %#v", replacement)
	}
	deletion := NewTextDeletion(4, 5)
	if deletion.Start != 4 || deletion.DeleteLength != 5 || deletion.InsertText != "" {
		t.Fatalf("deletion = %#v", deletion)
	}
	insertion := NewTextInsertion(6, "x")
	if insertion.Start != 6 || insertion.DeleteLength != 0 || insertion.InsertText != "x" {
		t.Fatalf("insertion = %#v", insertion)
	}

	err := recoverPanic(t, func() {
		NewTextEdit(-1, 0, "")
	})
	if outOfRange, ok := err.(ArgumentOutOfRangeError); !ok || outOfRange.ParameterName != "start" {
		t.Fatalf("negative start panic = %#v, want ArgumentOutOfRangeError(start)", err)
	}

	err = recoverPanic(t, func() {
		NewTextEdit(0, -1, "")
	})
	if outOfRange, ok := err.(ArgumentOutOfRangeError); !ok || outOfRange.ParameterName != "deleteLength" {
		t.Fatalf("negative delete panic = %#v, want ArgumentOutOfRangeError(deleteLength)", err)
	}
}

func TestCompletionDisplayTextForPortedSymbolShapes(t *testing.T) {
	table := NewTableSymbol("Events")
	if got := CompletionDisplayText(table); got != "Events" {
		t.Fatalf("table display = %q", got)
	}
	external := NewExternalTableSymbol("Archive")
	if got := CompletionDisplayText(external); got != "external_table('Archive')" {
		t.Fatalf("external table display = %q", got)
	}
	if got := CompletionDisplayText(external, true); got != "Archive" {
		t.Fatalf("external table name-only display = %q", got)
	}

	required := NewParameter("left", ScalarTypes.Long.Type())
	optional := NewParameter("right", ScalarTypes.Long.Type(), WithParameterOccurrence(0, 1))
	repeat := NewParameter("more", ScalarTypes.Long.Type(), WithParameterOccurrence(0, MaxRepeat))
	fn := NewFunctionSymbol("sum2", NewSignature(ScalarTypes.Long.Type(), required, optional, repeat))
	if got := CompletionDisplayText(fn); got != "sum2(left, [right], [more], ...)" {
		t.Fatalf("function display = %q", got)
	}
	if got := CompletionDisplayText(fn, true); got != "sum2" {
		t.Fatalf("function name-only display = %q", got)
	}

	pattern := NewPatternSymbol("detections", []*Parameter{NewParameter("source", ScalarTypes.String.Type())}, nil, nil)
	if got := CompletionDisplayText(pattern); got != "detections(source)" {
		t.Fatalf("pattern display = %q", got)
	}
	callableVariable := NewVariableSymbol("callable", fn, false, nil, nil)
	if got := CompletionDisplayText(callableVariable); got != "sum2(left, [right], [more], ...)" {
		t.Fatalf("callable variable display = %q", got)
	}
	callableParameter := NewParameterSymbol("callback", fn, "")
	if got := CompletionDisplayText(callableParameter, true); got != "sum2" {
		t.Fatalf("callable parameter display = %q", got)
	}
	option := NewOptionSymbolWithType("query_now", "", ScalarTypes.DateTime, nil)
	if got := CompletionDisplayText(option); got != "query_now" {
		t.Fatalf("option display = %q", got)
	}
}
