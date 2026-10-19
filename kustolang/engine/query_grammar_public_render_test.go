package kustolang

import "testing"

func TestQueryGrammarPublicRenderChartTypesAndCompletionVisibility(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	hidden := map[string]bool{"3Dchart": true, "graph": true, "sankey": true}

	for _, chartType := range KustoFacts.ChartTypes {
		t.Run(chartType, func(t *testing.T) {
			render := parseQueryGrammarProduction(t, grammar.renderOperator, "render "+chartType).(*RenderOperator)
			if render.ChartType().RawText() != chartType || len(render.ContainedDiagnostics()) != 0 {
				t.Fatalf("render chart = %q diagnostics:%#v", render.ChartType().RawText(), render.ContainedDiagnostics())
			}
		})
	}

	unknown := parseQueryGrammarProduction(t, grammar.renderOperator, "render custom_visual").(*RenderOperator)
	if unknown.ChartType().Kind() != SyntaxKindIdentifierToken || unknown.ChartType().RawText() != "custom_visual" || len(unknown.ContainedDiagnostics()) != 0 {
		t.Fatalf("unknown chart token = %#v diagnostics:%#v", unknown.ChartType(), unknown.ContainedDiagnostics())
	}

	items := queryGrammarRenderCompletionItems(grammar.renderOperator, "render ")
	for _, chartType := range KustoFacts.ChartTypes {
		item, ok := completionItemByDisplayText(items, chartType)
		if hidden[chartType] {
			if ok {
				t.Fatalf("hidden chart type %q was offered: %#v", chartType, completionDisplayTexts(items))
			}
			continue
		}
		if !ok || item.Kind != CompletionKindRenderChart {
			t.Fatalf("chart completion %q = %#v, visible:%#v", chartType, item, completionDisplayTexts(items))
		}
	}
}

func TestQueryGrammarPublicRenderMissingChartAndGeneratedShape(t *testing.T) {
	render := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).renderOperator, "render").(*RenderOperator)
	if render.Kind() != SyntaxKindRenderOperator || render.RenderKeyword().Kind() != SyntaxKindRenderKeyword {
		t.Fatalf("render node/keyword kinds = %v/%v", render.Kind(), render.RenderKeyword().Kind())
	}
	if render.ChartType() == nil || render.ChartType().RawText() != "" || render.ChartType().Width() != 0 {
		t.Fatalf("missing chart token = %#v", render.ChartType())
	}
	if !diagnosticsContain(render.ChartType().Diagnostics(), "KS005") || diagnosticsContain(render.ChartType().Diagnostics(), "KS001") {
		t.Fatalf("missing chart diagnostics = %#v", render.ChartType().Diagnostics())
	}
	if render.Parameters() == nil || render.Parameters().Count() != 0 || render.WithClause() != nil {
		t.Fatalf("missing-chart remaining shape = parameters:%#v with:%#v", render.Parameters(), render.WithClause())
	}
}

func TestQueryGrammarPublicRenderModernDeclaredPropertiesAndColumnBoundary(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	render := parseQueryGrammarProduction(t, grammar.renderOperator,
		`render timechart with (kind=stacked, title='Title', accumulate=true, xcolumn=['Time Stamp'], ycolumns=A, B, series=C, anomalycolumns=['Odd Name'], access, {tenant}, xtitle='X', ytitle="Y", xaxis=linear, yaxis=log, legend=visible, ysplit=axes, ymin=0, ymax=10.5, xmin=datetime(2024-01-01), xmax=100)`).(*RenderOperator)

	if render.Parameters().Count() != 0 || render.WithClause() == nil {
		t.Fatalf("modern render shape = deprecated:%d with:%#v", render.Parameters().Count(), render.WithClause())
	}
	with := render.WithClause()
	if with.WithKeyword().Kind() != SyntaxKindWithKeyword || with.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		with.LeadingComma() != nil || with.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("modern with delimiters = %#v", with)
	}
	wantNames := []string{"kind", "title", "accumulate", "xcolumn", "ycolumns", "series", "anomalycolumns", "xtitle", "ytitle", "xaxis", "yaxis", "legend", "ysplit", "ymin", "ymax", "xmin", "xmax"}
	if with.Properties().Count() != len(wantNames) {
		t.Fatalf("modern properties = %d, want %d", with.Properties().Count(), len(wantNames))
	}
	for i, want := range wantNames {
		property := with.Properties().At(i).Element()
		if NameDeclarationSimpleName(property.Name()) != want || property.EqualToken().Kind() != SyntaxKindEqualToken || property.EqualToken().Width() == 0 {
			t.Fatalf("modern property %d = %#v, want %q with equals", i, property, want)
		}
	}

	ycolumns := renderPropertyNameList(t, with.Properties().At(4).Element())
	if ycolumns.Names().Count() != 2 || NameReferenceSimpleName(ycolumns.Names().At(0).Element()) != "A" || NameReferenceSimpleName(ycolumns.Names().At(1).Element()) != "B" {
		t.Fatalf("ycolumns boundary list = %#v", ycolumns.Names())
	}
	series := renderPropertyNameList(t, with.Properties().At(5).Element())
	if series.Names().Count() != 1 || NameReferenceSimpleName(series.Names().At(0).Element()) != "C" {
		t.Fatalf("series boundary list = %#v", series.Names())
	}
	anomalies := renderPropertyNameList(t, with.Properties().At(6).Element())
	wantAnomalies := []string{"Odd Name", "access", "tenant"}
	if anomalies.Names().Count() != len(wantAnomalies) {
		t.Fatalf("anomaly name count = %d, want %d", anomalies.Names().Count(), len(wantAnomalies))
	}
	for i, want := range wantAnomalies {
		if got := NameReferenceSimpleName(anomalies.Names().At(i).Element()); got != want {
			t.Fatalf("anomaly name %d = %q, want %q", i, got, want)
		}
	}
	if with.Properties().At(3).Element().ExpressionHint() != CompletionHintColumn ||
		with.Properties().At(4).Element().ExpressionHint() != CompletionHintColumn {
		t.Fatalf("column completion hints = %v/%v", with.Properties().At(3).Element().ExpressionHint(), with.Properties().At(4).Element().ExpressionHint())
	}
	if len(render.ContainedDiagnostics()) != 0 {
		t.Fatalf("valid modern render diagnostics = %#v", render.ContainedDiagnostics())
	}

	known := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (hint.strategy=shuffle)").(*RenderOperator)
	if known.WithClause().Properties().Count() != 1 || NameDeclarationSimpleName(known.WithClause().Properties().At(0).Element().Name()) != "hint.strategy" {
		t.Fatalf("known compatibility property = %#v", known.WithClause().Properties())
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.renderOperator, "render timechart with (unknown=value)", 4)
}

func TestQueryGrammarPublicRenderWithClauseCommaForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	leading := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (, xcolumn=Timestamp)").(*RenderOperator).WithClause()
	if leading.LeadingComma() == nil || leading.LeadingComma().Kind() != SyntaxKindCommaToken || leading.Properties().Count() != 1 || len(leading.ContainedDiagnostics()) != 0 {
		t.Fatalf("leading comma clause = %#v diagnostics:%#v", leading, leading.ContainedDiagnostics())
	}
	empty := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with ()").(*RenderOperator).WithClause()
	if empty.LeadingComma() != nil || empty.Properties().Count() != 0 || len(empty.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty with clause = %#v diagnostics:%#v", empty, empty.ContainedDiagnostics())
	}

	trailing := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (kind=stacked,)").(*RenderOperator).WithClause().Properties()
	if trailing.Count() != 2 || !diagnosticsContainMessage(trailing.At(1).Element().ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("trailing comma properties = %#v", trailing)
	}

	double, length, tokenCount := parseQueryGrammarRenderPrefix(grammar.renderOperator, "render timechart with (kind=stacked,,title='x')")
	properties := double.WithClause().Properties()
	if properties.Count() != 3 || !diagnosticsContainMessage(properties.At(1).Element().ContainedDiagnostics(), "KS006", "Missing parameter") ||
		NameDeclarationSimpleName(properties.At(2).Element().Name()) != "title" {
		t.Fatalf("double comma properties = %#v diagnostics:%#v", properties, double.ContainedDiagnostics())
	}
	if length != tokenCount || double.WithClause().CloseParen().Width() == 0 {
		t.Fatalf("double comma boundary = consumed:%d/%d close:%#v", length, tokenCount, double.WithClause().CloseParen())
	}
}

func TestQueryGrammarPublicRenderColumnListCommaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (ycolumns=)").(*RenderOperator)
	emptyNames := renderPropertyNameList(t, empty.WithClause().Properties().At(0).Element())
	if emptyNames.Names().Count() != 1 || !diagnosticsContainMessage(emptyNames.Names().At(0).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("empty ycolumns = %#v diagnostics:%#v", emptyNames.Names(), empty.ContainedDiagnostics())
	}

	trailing := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (ycolumns=A,)").(*RenderOperator)
	trailingProperties := trailing.WithClause().Properties()
	trailingNames := renderPropertyNameList(t, trailingProperties.At(0).Element()).Names()
	if trailingProperties.Count() != 1 || trailingNames.Count() != 2 ||
		!diagnosticsContainMessage(trailingNames.At(1).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("trailing ycolumns comma = %#v diagnostics:%#v", trailingProperties, trailing.ContainedDiagnostics())
	}

	double, length, tokenCount := parseQueryGrammarRenderPrefix(grammar.renderOperator, "render timechart with (ycolumns=A,,B)")
	names := renderPropertyNameList(t, double.WithClause().Properties().At(0).Element()).Names()
	if names.Count() != 3 || !diagnosticsContainMessage(names.At(1).Element().ContainedDiagnostics(), "KS006", "Missing name") ||
		NameReferenceSimpleName(names.At(2).Element()) != "B" {
		t.Fatalf("double ycolumns comma = %#v diagnostics:%#v", names, double.ContainedDiagnostics())
	}
	if length != tokenCount || double.WithClause().CloseParen().Width() == 0 {
		t.Fatalf("double ycolumns boundary = consumed:%d/%d close:%#v", length, tokenCount, double.WithClause().CloseParen())
	}
}

func TestQueryGrammarPublicRenderParenthesizedColumnListMissingNameRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	render, length, tokenCount := parseQueryGrammarRenderPrefix(
		grammar.renderOperator,
		"render timechart with (ycolumns=(Count, Missing))",
	)
	property := render.WithClause().Properties().At(0).Element()
	names := renderPropertyNameList(t, property).Names()
	if names.Count() != 1 || !diagnosticsContainMessage(names.At(0).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("parenthesized ycolumns recovery = %#v", names)
	}
	if length >= tokenCount || render.WithClause().CloseParen().Width() != 0 {
		t.Fatalf("parenthesized ycolumns boundary = consumed:%d/%d close:%#v", length, tokenCount, render.WithClause().CloseParen())
	}

	control := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (ycolumns=Count)").(*RenderOperator)
	controlNames := renderPropertyNameList(t, control.WithClause().Properties().At(0).Element()).Names()
	if controlNames.Count() != 1 || NameReferenceSimpleName(controlNames.At(0).Element()) != "Count" || len(control.ContainedDiagnostics()) != 0 {
		t.Fatalf("ordinary ycolumns control = %#v diagnostics:%#v", controlNames, control.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicRenderDeprecatedFormsAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	render := parseQueryGrammarProduction(t, grammar.renderOperator,
		`render barchart kind=stacked title='Legacy' accumulate=true with 'legacy-options' by A, ['Display Name'], access, {tenant}`).(*RenderOperator)
	if render.WithClause() != nil || render.Parameters().Count() != 5 {
		t.Fatalf("deprecated render shape = parameters:%d with:%#v", render.Parameters().Count(), render.WithClause())
	}
	want := []string{"kind", "title", "accumulate", "with", "by"}
	for i, name := range want {
		parameter := render.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != name {
			t.Fatalf("deprecated parameter %d = %q, want %q", i, NameDeclarationSimpleName(parameter.Name()), name)
		}
		if parameter.EqualToken().Width() != map[bool]int{true: 1, false: 0}[i < 3] {
			t.Fatalf("deprecated parameter %q equals width = %d", name, parameter.EqualToken().Width())
		}
	}
	by := renderPropertyNameList(t, render.Parameters().At(4))
	wantBy := []string{"A", "Display Name", "access", "tenant"}
	if by.Names().Count() != len(wantBy) {
		t.Fatalf("deprecated by names = %d, want %d", by.Names().Count(), len(wantBy))
	}
	for i, wantName := range wantBy {
		if got := NameReferenceSimpleName(by.Names().At(i).Element()); got != wantName {
			t.Fatalf("deprecated by name %d = %q, want %q", i, got, wantName)
		}
	}
	if len(render.ContainedDiagnostics()) != 0 {
		t.Fatalf("valid deprecated render diagnostics = %#v", render.ContainedDiagnostics())
	}

	boundary := parseQueryGrammarProduction(t, grammar.renderOperator,
		`render barchart by A, B kind=stacked title='T' accumulate=false with 'old'`).(*RenderOperator)
	if boundary.Parameters().Count() != 5 || renderPropertyNameList(t, boundary.Parameters().At(0)).Names().Count() != 2 {
		t.Fatalf("deprecated property boundary = %#v", boundary.Parameters())
	}
	modernBoundary := parseQueryGrammarProduction(t, grammar.renderOperator, "render barchart with (title='modern')").(*RenderOperator)
	if modernBoundary.Parameters().Count() != 0 || modernBoundary.WithClause() == nil {
		t.Fatalf("modern/deprecated with boundary = %#v", modernBoundary)
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.renderOperator, "render barchart with (title='modern') kind=stacked", 8)
	assertQueryGrammarProductionLeavesTail(t, grammar.renderOperator, "render barchart , kind=stacked", 2)
}

func TestQueryGrammarPublicRenderPlacementTerminalTailsTagsAndCompletions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private":        grammar.renderOperator,
		"query":          grammar.QueryOperator,
		"following-pipe": grammar.FollowingPipeElementExpression,
		"fork-pipe":      grammar.forkPipeOperator,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, "render timechart").(*RenderOperator); !ok {
				t.Fatal("render production returned a non-render operator")
			}
		})
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression, "T | render timechart with (xcolumn=Timestamp) | take 1")
	takePipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindTakeOperator)
	renderPipe, renderNode := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindRenderOperator)
	if source, ok := renderPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("render pipeline source = %#v", renderPipe.Expression())
	}
	if renderNode.(*RenderOperator).WithClause().Properties().Count() != 1 {
		t.Fatalf("render pipeline node = %#v", renderNode)
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.renderOperator, "render timechart with (xcolumn=Timestamp) Extra", 8)
	assertQueryGrammarProductionLeavesTail(t, grammar.renderOperator, "render timechart with (xcolumn=Timestamp), Extra", 8)

	if grammar.renderOperator.Tag() != "<render>" ||
		!queryGrammarParserTreeHasTag(grammar.QueryOperator, "<render>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.FollowingPipeElementExpression, "<render>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.forkPipeOperator, "<render>", 0) {
		t.Fatalf("render tags = direct:%q query:%t following:%t fork:%t", grammar.renderOperator.Tag(),
			queryGrammarParserTreeHasTag(grammar.QueryOperator, "<render>", 0),
			queryGrammarParserTreeHasTag(grammar.FollowingPipeElementExpression, "<render>", 0),
			queryGrammarParserTreeHasTag(grammar.forkPipeOperator, "<render>", 0))
	}
	for _, parser := range []Parser[LexicalToken]{grammar.renderOperator, grammar.QueryOperator, grammar.FollowingPipeElementExpression} {
		item, ok := completionItemByDisplayText(queryGrammarRenderCompletionItems(parser, ""), "render")
		if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityHigh {
			t.Fatalf("render completion through %q = %#v", parser.Tag(), item)
		}
	}
	if _, ok := completionItemByDisplayText(queryGrammarRenderCompletionItems(grammar.forkPipeOperator, ""), "render"); ok {
		t.Fatal("render completion became visible through the hidden fork QueryOperator fallback")
	}

	root := NewKustoCodeService("").GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{})
	if _, ok := completionItemByDisplayText(root.Items, "render"); ok {
		t.Fatalf("post-pipe render appeared at query root: %#v", completionDisplayTexts(root.Items))
	}
	pipedText := "T | "
	piped := NewKustoCodeService(pipedText).GetCompletionItems(len(pipedText), CompletionOptionsDefault, CancellationToken{})
	if item, ok := completionItemByDisplayText(piped.Items, "render"); !ok || item.Kind != CompletionKindQueryPrefix {
		t.Fatalf("piped render completion = %#v, visible:%#v", item, completionDisplayTexts(piped.Items))
	}
}

func TestQueryGrammarPublicRenderMalformedDiagnostics(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingOpen, length, tokenCount := parseQueryGrammarRenderPrefix(grammar.renderOperator, "render timechart with title='x'")
	if length != tokenCount || missingOpen.WithClause() != nil || missingOpen.Parameters().Count() != 2 ||
		NameDeclarationSimpleName(missingOpen.Parameters().At(0).Name()) != "with" ||
		!diagnosticsContainMessage(missingOpen.Parameters().At(0).Expression().Diagnostics(), "KS006", "Missing value") ||
		NameDeclarationSimpleName(missingOpen.Parameters().At(1).Name()) != "title" {
		t.Fatalf("deprecated with recovery = consumed:%d/%d node:%#v diagnostics:%#v", length, tokenCount, missingOpen, missingOpen.ContainedDiagnostics())
	}
	missingClose := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (title='x'").(*RenderOperator)
	if !diagnosticsContainMessage(missingClose.WithClause().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing with close paren diagnostics = %#v", missingClose.ContainedDiagnostics())
	}

	missingEquals := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with (accumulate maybe, xcolumn Timestamp)").(*RenderOperator)
	if missingEquals.WithClause().Properties().Count() != 2 {
		t.Fatalf("missing-equals properties = %#v", missingEquals.WithClause().Properties())
	}
	for i := 0; i < 2; i++ {
		property := missingEquals.WithClause().Properties().At(i).Element()
		if property.EqualToken().Width() != 0 || !diagnosticsContain(property.EqualToken().Diagnostics(), "KS005") {
			t.Fatalf("missing equals property %d = %#v diagnostics:%#v", i, property, property.ContainedDiagnostics())
		}
	}

	missingValues := parseQueryGrammarProduction(t, grammar.renderOperator,
		"render timechart with (accumulate=, ymin=, yaxis=)").(*RenderOperator).WithClause().Properties()
	wantKinds := []SyntaxKind{SyntaxKindBooleanLiteralExpression, SyntaxKindRealLiteralExpression, SyntaxKindTokenLiteralExpression}
	for i, wantKind := range wantKinds {
		value := missingValues.At(i).Element().Expression()
		if value.Kind() != wantKind || len(value.ContainedDiagnostics()) == 0 {
			t.Fatalf("missing value %d = kind:%v diagnostics:%#v", i, value.Kind(), value.ContainedDiagnostics())
		}
	}

	deprecatedWith := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart with").(*RenderOperator)
	if deprecatedWith.Parameters().Count() != 1 || !diagnosticsContainMessage(deprecatedWith.Parameters().At(0).Expression().ContainedDiagnostics(), "KS006", "Missing value") {
		t.Fatalf("missing deprecated with value = %#v diagnostics:%#v", deprecatedWith.Parameters(), deprecatedWith.ContainedDiagnostics())
	}
	deprecatedBy := parseQueryGrammarProduction(t, grammar.renderOperator, "render timechart by").(*RenderOperator)
	if deprecatedBy.Parameters().Count() != 1 {
		t.Fatalf("missing deprecated by list = %#v", deprecatedBy.Parameters())
	}
	byNames := renderPropertyNameList(t, deprecatedBy.Parameters().At(0)).Names()
	if byNames.Count() != 1 || !diagnosticsContainMessage(byNames.At(0).Element().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing deprecated by name = %#v diagnostics:%#v", byNames, deprecatedBy.ContainedDiagnostics())
	}
}

func renderPropertyNameList(t *testing.T, parameter *NamedParameter) *NameReferenceList {
	t.Helper()
	list, ok := parameter.Expression().(*NameReferenceList)
	if !ok {
		t.Fatalf("parameter %q expression = %T, want *NameReferenceList", NameDeclarationSimpleName(parameter.Name()), parameter.Expression())
	}
	return list
}

func parseQueryGrammarRenderPrefix(parser ResultParser[LexicalToken, QueryOperator], text string) (*RenderOperator, int, int) {
	tokens := Lex(text)
	result := parser.ParseValue(NewArraySource(tokens), 0)
	render, _ := result.Value.(*RenderOperator)
	return render, result.Length, len(tokens) - 1
}

func queryGrammarRenderCompletionItems(parser Parser[LexicalToken], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
