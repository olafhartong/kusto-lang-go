package kustolang

import "testing"

func TestQueryGrammarPublicGraphOperatorsAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	mark := parseQueryGrammarProduction(t, grammar.graphMarkComponentsOperator,
		"graph-mark-components kind=strong with_component_id=ComponentId").(*GraphMarkComponentsOperator)
	if mark.Parameters().Count() != 2 || len(mark.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-mark-components = %#v diagnostics:%#v", mark, mark.ContainedDiagnostics())
	}

	whereNodes := parseQueryGrammarProduction(t, grammar.graphWhereNodesOperator,
		"graph-where-nodes Degree > 1 and Active").(*GraphWhereNodesOperator)
	if whereNodes.Condition().Kind() != SyntaxKindAndExpression || len(whereNodes.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-where-nodes condition = %#v diagnostics:%#v", whereNodes.Condition(), whereNodes.ContainedDiagnostics())
	}
	whereEdges := parseQueryGrammarProduction(t, grammar.graphWhereEdgesOperator,
		"graph-where-edges Weight between (1 .. 10)").(*GraphWhereEdgesOperator)
	if whereEdges.Condition().Kind() != SyntaxKindBetweenExpression || len(whereEdges.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-where-edges condition = %#v diagnostics:%#v", whereEdges.Condition(), whereEdges.ContainedDiagnostics())
	}

	for _, test := range []struct {
		text string
		kind SyntaxKind
	}{
		{"graph-match (source)-->(target)", SyntaxKindGraphMatchOperator},
		{"graph-shortest-paths (source)-->(target)", SyntaxKindGraphShortestPathsOperator},
		{"graph-mark-components", SyntaxKindGraphMarkComponentsOperator},
		{"graph-where-nodes true", SyntaxKindGraphWhereNodesOperator},
		{"graph-where-edges true", SyntaxKindGraphWhereEdgesOperator},
		{"graph-to-table nodes", SyntaxKindGraphToTableOperator},
	} {
		operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
		if operator.Kind() != test.kind {
			t.Fatalf("QueryOperator %q kind = %v, want %v", test.text, operator.Kind(), test.kind)
		}
		following := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, test.text)
		if following.Kind() != test.kind {
			t.Fatalf("FollowingPipeElementExpression %q kind = %v, want %v", test.text, following.Kind(), test.kind)
		}
	}
}

func TestQueryGrammarPublicGraphToTableASTAndRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.graphToTableOperator,
		"graph-to-table nodes as Vertices with_node_id=NodeId, edges as Links with_source_id=SourceId with_target_id=TargetId").(*GraphToTableOperator)
	if operator.OutputClause().Count() != 2 || len(operator.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-to-table output = %#v diagnostics:%#v", operator.OutputClause(), operator.ContainedDiagnostics())
	}
	nodes := operator.OutputClause().At(0).Element()
	edges := operator.OutputClause().At(1).Element()
	if nodes.EntityKeyword().Kind() != SyntaxKindNodesKeyword || NameDeclarationSimpleName(nodes.AsClause().Name()) != "Vertices" || nodes.Parameters().Count() != 1 {
		t.Fatalf("graph-to-table nodes = %#v", nodes)
	}
	if edges.EntityKeyword().Kind() != SyntaxKindGraphEdgesKeyword || NameDeclarationSimpleName(edges.AsClause().Name()) != "Links" || edges.Parameters().Count() != 2 {
		t.Fatalf("graph-to-table edges = %#v", edges)
	}

	missingName := parseQueryGrammarProduction(t, grammar.graphToTableOperator, "graph-to-table nodes as").(*GraphToTableOperator)
	name := missingName.OutputClause().At(0).Element().AsClause().Name()
	if name.Width() != 0 || !diagnosticsContainMessage(name.ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("graph-to-table missing alias = %#v diagnostics:%#v", name, name.ContainedDiagnostics())
	}

	missingOutput := parseQueryGrammarProduction(t, grammar.graphToTableOperator, "graph-to-table").(*GraphToTableOperator)
	output := missingOutput.OutputClause().At(0).Element()
	if output.EntityKeyword().Kind() != SyntaxKindOpenParenToken || output.EntityKeyword().Width() != 0 ||
		output.AsClause().AsKeyword().Kind() != SyntaxKindOpenParenToken || output.Parameters() != nil {
		t.Fatalf("odd graph-to-table missing output = %#v", output)
	}
	if diagnosticsContainMessage(output.ContainedDiagnostics(), "KS230", "requires 1-2 arguments") {
		t.Fatalf("missing output factory unexpectedly owns KS230: %#v", output.ContainedDiagnostics())
	}
	if !diagnosticsContainMessage(output.AsClause().Name().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing output name diagnostics = %#v", output.AsClause().Name().ContainedDiagnostics())
	}

	leadingComma := parseQueryGrammarProduction(t, grammar.graphToTableOperator, "graph-to-table , nodes").(*GraphToTableOperator)
	if leadingComma.OutputClause().Count() != 2 || leadingComma.OutputClause().At(0).Element().Parameters() != nil {
		t.Fatalf("graph-to-table leading comma recovery = %#v", leadingComma.OutputClause())
	}
}

func TestQueryGrammarPublicGraphMatchPatternsAndClauses(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.graphMatchOperator,
		"graph-match cycles=unique_edges (a)-->(b), (b)<--(c), (c)--(d), (d)<-[edge*+lower()..-upper()]-(e) where a.score > 0 project a, endNode=e").(*GraphMatchOperator)
	if operator.Parameters().Count() != 1 || operator.Patterns().Count() != 4 || operator.WhereClause() == nil || operator.ProjectClause() == nil {
		t.Fatalf("graph-match clauses = %#v", operator)
	}
	if operator.ProjectClause().Expressions().Count() != 2 || len(operator.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-match project/diagnostics = %#v/%#v", operator.ProjectClause(), operator.ContainedDiagnostics())
	}

	plainKinds := []SyntaxKind{SyntaxKindDashDashGreaterThanToken, SyntaxKindLessThanDashDashToken, SyntaxKindDashDashToken}
	for i, kind := range plainKinds {
		edge := operator.Patterns().At(i).Element().PatternElements().At(1).(*GraphMatchPatternEdge)
		if edge.FirstToken().Kind() != kind || edge.Name() != nil || edge.Range() != nil || edge.LastToken() != nil {
			t.Fatalf("plain graph edge %d = %#v", i, edge)
		}
	}
	bracketed := operator.Patterns().At(3).Element().PatternElements().At(1).(*GraphMatchPatternEdge)
	if bracketed.FirstToken().Kind() != SyntaxKindLessThanDashBracketToken || bracketed.LastToken().Kind() != SyntaxKindBracketDashToken ||
		NameDeclarationSimpleName(bracketed.Name()) != "edge" || bracketed.Range() == nil {
		t.Fatalf("bracketed graph edge = %#v", bracketed)
	}
	if bracketed.Range().RangeStart().Kind() != SyntaxKindUnaryPlusExpression || bracketed.Range().RangeEnd().Kind() != SyntaxKindUnaryMinusExpression {
		t.Fatalf("graph edge invocation range = %#v", bracketed.Range())
	}

	shortest := parseQueryGrammarProduction(t, grammar.graphShortestPathsOperator,
		"graph-shortest-paths output=all (source)-[edge*1..3]->(target) where edge.enabled project source, target").(*GraphShortestPathsOperator)
	if shortest.Parameters().Count() != 1 || shortest.Patterns().Count() != 1 || shortest.WhereClause() == nil || shortest.ProjectClause() == nil || len(shortest.ContainedDiagnostics()) != 0 {
		t.Fatalf("graph-shortest-paths = %#v diagnostics:%#v", shortest, shortest.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicGraphMalformedOwnershipAndBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingFilter := parseQueryGrammarProduction(t, grammar.graphWhereNodesOperator, "graph-where-nodes").(*GraphWhereNodesOperator)
	if missingFilter.Condition().Width() != 0 || !diagnosticsContainMessage(missingFilter.Condition().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing graph filter = %#v diagnostics:%#v", missingFilter.Condition(), missingFilter.Condition().ContainedDiagnostics())
	}

	rangeChild := parseQueryGrammarProduction(t, grammar.graphMatchPatternEdgeRange, "* ..")
	if rangeChild.RangeStart().Width() != 0 || rangeChild.RangeEnd().Width() != 0 || rangeChild.DotDotToken().Width() == 0 {
		t.Fatalf("missing graph range endpoints = %#v", rangeChild)
	}
	if !diagnosticsContainMessage(rangeChild.RangeStart().ContainedDiagnostics(), "KS006", "Missing expression") ||
		!diagnosticsContainMessage(rangeChild.RangeEnd().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("graph range endpoint diagnostics = %#v/%#v", rangeChild.RangeStart().ContainedDiagnostics(), rangeChild.RangeEnd().ContainedDiagnostics())
	}
	missingDots := parseQueryGrammarProduction(t, grammar.graphMatchPatternEdgeRange, "*1 3")
	if missingDots.DotDotToken().Width() != 0 || !diagnosticsContainMessage(missingDots.DotDotToken().Diagnostics(), "KS005", "Expected: ..") {
		t.Fatalf("graph range missing dots = %#v diagnostics:%#v", missingDots.DotDotToken(), missingDots.DotDotToken().Diagnostics())
	}

	missingPattern := parseQueryGrammarProduction(t, grammar.graphMatchOperator, "graph-match where true").(*GraphMatchOperator)
	pattern := missingPattern.Patterns().At(0).Element()
	if !diagnosticsContainMessage(pattern.PatternElements().Diagnostics(), "KS217", "Missing graph-match pattern.") {
		t.Fatalf("missing pattern list diagnostics = %#v", pattern.PatternElements().Diagnostics())
	}
	notation := pattern.PatternElements().At(0)
	if !diagnosticsContainMessage(notation.ContainedDiagnostics(), "KS238", "Missing graph-match pattern element.") {
		t.Fatalf("missing notation diagnostics = %#v", notation.ContainedDiagnostics())
	}
	if len(pattern.Diagnostics()) != 0 {
		t.Fatalf("missing pattern node unexpectedly owns diagnostics: %#v", pattern.Diagnostics())
	}

	assertQueryGrammarProductionRejects(t, grammar.graphMatchPatternNode, "(node")
	missingEdgeClose := parseQueryGrammarProduction(t, grammar.graphMatchPatternEdge, "-[edge").(*GraphMatchPatternEdge)
	if missingEdgeClose.LastToken().Width() != 0 ||
		!diagnosticsContainMessage(missingEdgeClose.LastToken().Diagnostics(), "KS005", "Expected one of: ]->, ]-") {
		t.Fatalf("missing graph edge closer = %#v diagnostics:%#v", missingEdgeClose.LastToken(), missingEdgeClose.LastToken().Diagnostics())
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.graphMatchOperator,
		"graph-match (a)-->(b) trailing", len(Lex("graph-match (a)-->(b)"))-1)
	assertQueryGrammarProductionLeavesTail(t, grammar.graphToTableOperator,
		"graph-to-table nodes trailing", len(Lex("graph-to-table nodes"))-1)
}

func TestQueryGrammarPublicGraphTagsCompletionsAndOrder(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, tag := range []string{
		"<graph-mark-components>",
		"<graph-where-nodes>",
		"<graph-where-edges>",
		"<graph-to-table-as-clause>",
		"graph-to-table-output-nodes-clause",
		"graph-to-table-output-edges-clause",
		"<graph-to-table-output-clause>",
		"<graph-to-table>",
		"<graph-match-pattern-clause>",
		"<graph-match-operator>",
		"<graph-shortest-paths-operator>",
	} {
		if !queryGrammarParserTreeHasTag(grammar.QueryOperator, tag, 0) {
			t.Fatalf("QueryOperator parser tree is missing %s", tag)
		}
	}

	patternItems := syntaxCompletionItemsFromAnnotations(grammar.graphMatchPattern.Annotations())
	for _, snippet := range []string{"(n1)-[e]->(n2)", "(n1)-[e1]->(n2)-[e2]->(n3)", "(n1)-[e*1..3]->(n2)"} {
		item, ok := completionItemByDisplayText(patternItems, snippet)
		if !ok || item.Kind != CompletionKindSyntax {
			t.Fatalf("graph pattern completion %q = %#v, visible:%#v", snippet, item, completionDisplayTexts(patternItems))
		}
	}

	items := queryGrammarGraphCompletionItems(grammar.QueryOperator, "")
	ordered := []string{"getschema", "graph-match", "graph-shortest-paths", "graph-mark-components", "graph-to-table", "invoke", "join", "lookup", "make-graph"}
	last := -1
	for _, display := range ordered {
		index := completionDisplayTextIndex(items, display)
		if index <= last {
			t.Fatalf("graph operator completion order %q index=%d after=%d; visible:%#v", display, index, last, completionDisplayTexts(items))
		}
		last = index
	}
	for _, hidden := range []string{"graph-where-nodes", "graph-where-edges"} {
		if _, ok := completionItemByDisplayText(items, hidden); ok {
			t.Fatalf("hidden graph filter completion exposed: %q in %#v", hidden, completionDisplayTexts(items))
		}
	}

	asItems := queryGrammarGraphCompletionItems(grammar.graphToTableOperator, "graph-to-table nodes ")
	asItem, ok := completionItemByDisplayText(asItems, "as")
	if !ok || asItem.Kind != CompletionKindKeyword || asItem.priority != CompletionPriorityLow {
		t.Fatalf("graph-to-table as completion = %#v, visible:%#v", asItem, completionDisplayTexts(asItems))
	}
}

func queryGrammarGraphCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}
