package kustolang

import "testing"

func TestFormattingOptionsDefaultMatchesUpstream(t *testing.T) {
	options := FormattingOptionsDefault
	if options.IndentationSize != 4 || !options.InsertMissingTokens {
		t.Fatalf("formatting scalar defaults = indentation:%d insertMissing:%t", options.IndentationSize, options.InsertMissingTokens)
	}
	if options.BrackettingStyle != BrackettingStyleVertical ||
		options.SchemaStyle != BrackettingStyleDefault ||
		options.DataTableValueStyle != BrackettingStyleDefault ||
		options.FunctionBodyStyle != BrackettingStyleDefault ||
		options.FunctionParameterStyle != BrackettingStyleDefault ||
		options.FunctionArgumentStyle != BrackettingStyleDiagonal {
		t.Fatalf("formatting bracket defaults = %#v", options)
	}
	if options.PipeOperatorStyle != PlacementStyleSmart ||
		options.ExpressionStyle != PlacementStyleSmart ||
		options.StatementStyle != PlacementStyleNewLine ||
		options.SemicolonStyle != PlacementStyleNone {
		t.Fatalf("formatting placement defaults = %#v", options)
	}
	if options.GeneralSpacing != SpacingStyleOne ||
		options.PrefixOperatorSpacing != SpacingStyleMinimal ||
		options.InfixOperatorSpacing != DualSpacingStyleBoth ||
		options.PipeOperatorSpacing != DualSpacingStyleBoth ||
		options.CommaSpacing != DualSpacingStyleAfter ||
		options.ColonSpacing != DualSpacingStyleAfter ||
		options.AssignmentSpacing != DualSpacingStyleBoth ||
		options.RangeOperatorSpacing != DualSpacingStyleBoth ||
		options.SemicolonSpacing != DualSpacingStyleAfter {
		t.Fatalf("formatting operator spacing defaults = %#v", options)
	}
	if options.ParenthesizedExpressionSpacing != SpacingStyleMinimal ||
		options.ArgumentListSpacing != SpacingStyleMinimal ||
		options.EmptyArgumentListSpacing != SpacingStyleMinimal ||
		options.ParameterListSpacing != SpacingStyleMinimal ||
		options.EmptyParameterListSpacing != SpacingStyleMinimal ||
		options.JsonArraySpacing != SpacingStyleMinimal ||
		options.EmptyJsonArraySpacing != SpacingStyleMinimal ||
		options.JsonObjectSpacing != SpacingStyleMinimal ||
		options.EmptyJsonObjectSpacing != SpacingStyleMinimal ||
		options.DataTableValueSpacing != SpacingStyleMinimal ||
		options.EmptyDataTableValueSpacing != SpacingStyleMinimal ||
		options.FunctionBodySpacing != SpacingStyleOne ||
		options.EmptyFunctionBodySpacing != SpacingStyleOne ||
		options.BeforeFunctionBodySpacing != SpacingStyleOne ||
		options.BeforeParameterListSpacing != SpacingStyleMinimal ||
		options.BeforeArgumentListSpacing != SpacingStyleMinimal ||
		options.BeforeDataTableValueSpacing != SpacingStyleOne {
		t.Fatalf("formatting delimiter spacing defaults = %#v", options)
	}
}

func TestFormattingOptionsWithHelpersReturnModifiedCopies(t *testing.T) {
	original := FormattingOptionsDefault
	changed := original.
		WithIndentationSize(2).
		WithInsertMissingTokens(false).
		WithBrackettingStyle(BrackettingStyleNone).
		WithSchemaStyle(BrackettingStyleVertical).
		WithDataTableValueStyle(BrackettingStyleDiagonal).
		WithFunctionBodyStyle(BrackettingStyleVertical).
		WithFunctionParameterStyle(BrackettingStyleDiagonal).
		WithFunctionArgumentStyle(BrackettingStyleVertical).
		WithPipeOperatorStyle(PlacementStyleNewLine).
		WithExpressionStyle(PlacementStyleNone).
		WithStatementStyle(PlacementStyleSmart).
		WithSemicolonStyle(PlacementStyleNewLine).
		WithGeneralSpacing(SpacingStyleAsIs).
		WithPrefixOperatorSpacing(SpacingStyleOne).
		WithInfixOperatorSpacing(DualSpacingStyleBefore).
		WithPipeOperatorSpacing(DualSpacingStyleAfter).
		WithCommaSpacing(DualSpacingStyleBoth).
		WithColonSpacing(DualSpacingStyleNeither).
		WithAssignmentSpacing(DualSpacingStyleBefore).
		WithRangeOperatorSpacing(DualSpacingStyleAfter).
		WithSemicolonSpacing(DualSpacingStyleNeither).
		WithExpressionParenSpacing(SpacingStyleOne).
		WithArgumentListSpacing(SpacingStyleOne).
		WithEmptyArgumentListSpacing(SpacingStyleOne).
		WithParameterListSpacing(SpacingStyleOne).
		WithEmptyParameterListSpacing(SpacingStyleOne).
		WithJsonArraySpacing(SpacingStyleOne).
		WithEmptyJsonArraySpacing(SpacingStyleOne).
		WithJsonObjectSpacing(SpacingStyleOne).
		WithEmptyJsonObjectSpacing(SpacingStyleOne).
		WithDataTableValueSpacing(SpacingStyleOne).
		WithEmptyDataTableValueSpacing(SpacingStyleOne).
		WithFunctionBodySpacing(SpacingStyleMinimal).
		WithEmptyFunctionBodySpacing(SpacingStyleMinimal).
		WithBeforeFunctionBodySpacing(SpacingStyleMinimal).
		WithBeforeParameterListSpacing(SpacingStyleOne).
		WithBeforeArgumentListSpacing(SpacingStyleOne).
		WithBeforeDataTableValueSpacing(SpacingStyleMinimal)

	if original != FormattingOptionsDefault {
		t.Fatalf("original/default formatting options mutated: %#v", original)
	}
	if changed.IndentationSize != 2 || changed.InsertMissingTokens {
		t.Fatalf("changed scalar options = %#v", changed)
	}
	if changed.BrackettingStyle != BrackettingStyleNone ||
		changed.SchemaStyle != BrackettingStyleVertical ||
		changed.DataTableValueStyle != BrackettingStyleDiagonal ||
		changed.FunctionBodyStyle != BrackettingStyleVertical ||
		changed.FunctionParameterStyle != BrackettingStyleDiagonal ||
		changed.FunctionArgumentStyle != BrackettingStyleVertical {
		t.Fatalf("changed bracket options = %#v", changed)
	}
	if changed.PipeOperatorStyle != PlacementStyleNewLine ||
		changed.ExpressionStyle != PlacementStyleNone ||
		changed.StatementStyle != PlacementStyleSmart ||
		changed.SemicolonStyle != PlacementStyleNewLine {
		t.Fatalf("changed placement options = %#v", changed)
	}
	if changed.GeneralSpacing != SpacingStyleAsIs ||
		changed.PrefixOperatorSpacing != SpacingStyleOne ||
		changed.InfixOperatorSpacing != DualSpacingStyleBefore ||
		changed.PipeOperatorSpacing != DualSpacingStyleAfter ||
		changed.CommaSpacing != DualSpacingStyleBoth ||
		changed.ColonSpacing != DualSpacingStyleNeither ||
		changed.AssignmentSpacing != DualSpacingStyleBefore ||
		changed.RangeOperatorSpacing != DualSpacingStyleAfter ||
		changed.SemicolonSpacing != DualSpacingStyleNeither {
		t.Fatalf("changed operator spacing options = %#v", changed)
	}
	if changed.ParenthesizedExpressionSpacing != SpacingStyleOne ||
		changed.ArgumentListSpacing != SpacingStyleOne ||
		changed.EmptyArgumentListSpacing != SpacingStyleOne ||
		changed.ParameterListSpacing != SpacingStyleOne ||
		changed.EmptyParameterListSpacing != SpacingStyleOne ||
		changed.JsonArraySpacing != SpacingStyleOne ||
		changed.EmptyJsonArraySpacing != SpacingStyleOne ||
		changed.JsonObjectSpacing != SpacingStyleOne ||
		changed.EmptyJsonObjectSpacing != SpacingStyleOne ||
		changed.DataTableValueSpacing != SpacingStyleOne ||
		changed.EmptyDataTableValueSpacing != SpacingStyleOne ||
		changed.FunctionBodySpacing != SpacingStyleMinimal ||
		changed.EmptyFunctionBodySpacing != SpacingStyleMinimal ||
		changed.BeforeFunctionBodySpacing != SpacingStyleMinimal ||
		changed.BeforeParameterListSpacing != SpacingStyleOne ||
		changed.BeforeArgumentListSpacing != SpacingStyleOne ||
		changed.BeforeDataTableValueSpacing != SpacingStyleMinimal {
		t.Fatalf("changed delimiter spacing options = %#v", changed)
	}
}

func TestOutliningOptionsDefaultsAndCloneHelpers(t *testing.T) {
	options := OutliningOptionsDefault
	if options.MaxCollapsedTextLength != 120 || !options.Queries || options.Statements {
		t.Fatalf("outlining defaults = %#v", options)
	}
	changed := options.WithMaxCollapesedTextlength(60).WithQueries(false).WithStatements(true)
	if changed.MaxCollapsedTextLength != 60 || changed.Queries || !changed.Statements {
		t.Fatalf("changed outlining options = %#v", changed)
	}
	if options != OutliningOptionsDefault {
		t.Fatalf("outlining default mutated: %#v", options)
	}
}

func TestQuickInfoTextAndItems(t *testing.T) {
	textOnly := NewQuickInfoText("plain")
	if got := textOnly.Text(); got != "plain" {
		t.Fatalf("text-only quick info = %q", got)
	}
	if len(textOnly.Items) != 1 || textOnly.Items[0].Kind != QuickInfoKindText || textOnly.Items[0].Text() != "plain" {
		t.Fatalf("text-only quick info items = %#v", textOnly.Items)
	}

	info := NewQuickInfo(
		NewQuickInfoItemText(QuickInfoKindTable, "Events"),
		NewQuickInfoItem(QuickInfoKindBuiltInFunction,
			NewClassifiedTextKind(ClassificationKindFunction, "ago"),
			NewClassifiedText("("),
			NewClassifiedText("timespan"),
			NewClassifiedText(")")),
		NewQuickInfoItemText(QuickInfoKindOption, "query_now"),
	)
	if got, want := info.Text(), "(table) Events\n\n(function) ago(timespan)\n\nquery_now"; got != want {
		t.Fatalf("quick info text = %q, want %q", got, want)
	}

	if int(QuickInfoKindText) != 0 || int(QuickInfoKindColumn) != 4 || int(QuickInfoKindGraph) != 20 {
		t.Fatalf("quick info enum order changed: text=%d column=%d graph=%d", QuickInfoKindText, QuickInfoKindColumn, QuickInfoKindGraph)
	}
	if QuickInfoEmpty.Items == nil || len(QuickInfoEmpty.Items) != 0 || QuickInfoEmpty.Text() != "" {
		t.Fatalf("empty quick info = items:%#v text:%q", QuickInfoEmpty.Items, QuickInfoEmpty.Text())
	}
}

func TestRelatedInfoNavigationAndRangeHelpers(t *testing.T) {
	elements := []RelatedElement{
		NewRelatedElement(2, 4, RelatedElementKindDeclaration),
		NewRelatedElementWithCursors(10, 3, RelatedElementKindReference, 11, 12),
	}
	info := NewRelatedInfo(elements, 1)
	elements[0] = NewRelatedElement(0, 1, RelatedElementKindOther)
	if len(info.Elements) != 2 || info.Elements[0].Start() != 2 || info.Elements[0].Length() != 4 || info.Elements[0].End() != 6 {
		t.Fatalf("related info should copy range elements, got %#v", info.Elements)
	}
	if info.CurrentIndex != 1 || info.GetNextIndex(0) != 1 || info.GetNextIndex(1) != 0 || info.GetPreviousIndex(1) != 0 || info.GetPreviousIndex(0) != 1 {
		t.Fatalf("related info navigation failed: %#v next0=%d next1=%d prev1=%d prev0=%d", info, info.GetNextIndex(0), info.GetNextIndex(1), info.GetPreviousIndex(1), info.GetPreviousIndex(0))
	}
	if info.Elements[1].CursorLeft != 11 || info.Elements[1].CursorRight != 12 || info.Elements[1].Kind != RelatedElementKindReference {
		t.Fatalf("related element cursors/kind = %#v", info.Elements[1])
	}
	if int(RelatedElementKindSyntax) != 0 || int(RelatedElementKindOther) != 3 {
		t.Fatalf("related element kind order changed: syntax=%d other=%d", RelatedElementKindSyntax, RelatedElementKindOther)
	}
	if RelatedInfoEmpty.Elements == nil || len(RelatedInfoEmpty.Elements) != 0 || RelatedInfoEmpty.GetPreviousIndex(0) != -1 {
		t.Fatalf("empty related info = %#v prev0=%d", RelatedInfoEmpty, RelatedInfoEmpty.GetPreviousIndex(0))
	}
}

func TestFindRelatedOptionsFlags(t *testing.T) {
	if FindRelatedOptionsNone != 0 || FindRelatedOptionsSeeThroughVariables != 1 {
		t.Fatalf("find related options values = none:%d see-through:%d", FindRelatedOptionsNone, FindRelatedOptionsSeeThroughVariables)
	}
	if FindRelatedOptionsSeeThroughVariables&FindRelatedOptionsSeeThroughVariables == 0 {
		t.Fatalf("find related flag should be set")
	}
}
