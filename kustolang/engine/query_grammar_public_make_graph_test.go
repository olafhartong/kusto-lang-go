package kustolang

import "testing"

func TestQueryGrammarPublicMakeGraphFullAST(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).makeGraphOperator,
		`make-graph hint.strategy=shuffle Source --> Target with_node_id=NodeId partitioned-by Tenant (where Active | count)`).(*MakeGraphOperator)

	if operator.MakeGraphKeyword().Kind() != SyntaxKindMakeGraphKeyword ||
		NameReferenceSimpleName(operator.SourceColumn()) != "Source" ||
		operator.DirectionToken().Kind() != SyntaxKindDashDashGreaterThanToken ||
		NameReferenceSimpleName(operator.TargetColumn()) != "Target" {
		t.Fatalf("make-graph fixed fields = %#v", operator)
	}
	if operator.Parameters().Count() != 1 || NameDeclarationSimpleName(operator.Parameters().At(0).Name()) != "hint.strategy" {
		t.Fatalf("make-graph parameters = %#v", operator.Parameters())
	}
	implicit, ok := operator.WithClause().(*MakeGraphWithImplicitIdClause)
	if !ok || implicit.WithNodeIdKeyword().Kind() != SyntaxKindWithNodeIdKeyword ||
		implicit.EqualToken().Kind() != SyntaxKindEqualToken || NameDeclarationSimpleName(implicit.Name()) != "NodeId" {
		t.Fatalf("make-graph implicit id = %#v", operator.WithClause())
	}
	partitioned := operator.PartitionedByClause()
	if partitioned == nil || NameReferenceSimpleName(partitioned.Entity()) != "Tenant" ||
		partitioned.OpenParen().Kind() != SyntaxKindOpenParenToken || partitioned.CloseParen().Kind() != SyntaxKindCloseParenToken ||
		partitioned.Subquery().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("make-graph partition = %#v", partitioned)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid make-graph diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicMakeGraphTablesAndDirections(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source -- Target with Nodes() on NodeId, database("db").table("People") on PersonId`).(*MakeGraphOperator)
	with, ok := operator.WithClause().(*MakeGraphWithTablesAndKeysClause)
	if !ok || with.WithKeyword().Kind() != SyntaxKindWithKeyword || with.TablesAndKeys().Count() != 2 {
		t.Fatalf("make-graph tables clause = %#v", operator.WithClause())
	}
	if operator.DirectionToken().Kind() != SyntaxKindDashDashToken {
		t.Fatalf("make-graph undirected token = %v", operator.DirectionToken().Kind())
	}
	first := with.TablesAndKeys().At(0).Element()
	second := with.TablesAndKeys().At(1).Element()
	if first.Table().Kind() != SyntaxKindFunctionCallExpression || NameReferenceSimpleName(first.Column()) != "NodeId" ||
		second.Table().Kind() != SyntaxKindPathExpression || NameReferenceSimpleName(second.Column()) != "PersonId" {
		t.Fatalf("make-graph table/key entries = %#v / %#v", first, second)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid make-graph table diagnostics = %#v", diagnostics)
	}

	for _, text := range []string{
		`make-graph Source --> Target with +Nodes() on NodeId`,
		`make-graph Source --> Target with -Nodes() on NodeId`,
	} {
		signed := parseQueryGrammarProduction(t, grammar.makeGraphOperator, text).(*MakeGraphOperator)
		entry := signed.WithClause().(*MakeGraphWithTablesAndKeysClause).TablesAndKeys().At(0).Element()
		if entry.Table().Kind() != SyntaxKindUnaryPlusExpression && entry.Table().Kind() != SyntaxKindUnaryMinusExpression {
			t.Fatalf("signed invocation %q kind = %v", text, entry.Table().Kind())
		}
	}
}

func TestQueryGrammarPublicMakeGraphRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	missing := parseQueryGrammarProduction(t, grammar.makeGraphOperator, `make-graph`).(*MakeGraphOperator)
	for name, element := range map[string]SyntaxElement{
		"source":    missing.SourceColumn(),
		"direction": missing.DirectionToken(),
		"target":    missing.TargetColumn(),
	} {
		if element == nil || element.Width() != 0 || len(element.ContainedDiagnostics()) == 0 {
			t.Fatalf("missing make-graph %s = %#v diagnostics:%#v", name, element, element.ContainedDiagnostics())
		}
	}

	missingDirection := parseQueryGrammarProduction(t, grammar.makeGraphOperator, `make-graph Source Target`).(*MakeGraphOperator)
	if missingDirection.DirectionToken().Width() != 0 || NameReferenceSimpleName(missingDirection.TargetColumn()) != "Target" {
		t.Fatalf("missing direction recovery = %#v", missingDirection)
	}

	implicit := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source --> Target with_node_id NodeId`).(*MakeGraphOperator).WithClause().(*MakeGraphWithImplicitIdClause)
	if implicit.EqualToken().Width() != 0 || NameDeclarationSimpleName(implicit.Name()) != "NodeId" {
		t.Fatalf("implicit id equals recovery = %#v", implicit)
	}
	missingID := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source --> Target with_node_id=`).(*MakeGraphOperator).WithClause().(*MakeGraphWithImplicitIdClause)
	if missingID.Name().Width() != 0 || len(missingID.Name().ContainedDiagnostics()) == 0 {
		t.Fatalf("implicit id name recovery = %#v", missingID.Name())
	}

	tables := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source --> Target with Nodes()`).(*MakeGraphOperator).WithClause().(*MakeGraphWithTablesAndKeysClause)
	entry := tables.TablesAndKeys().At(0).Element()
	if entry.OnKeyword().Width() != 0 || entry.Column().Width() != 0 {
		t.Fatalf("table/key recovery = %#v", entry)
	}

	partitioned := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source --> Target partitioned-by`).(*MakeGraphOperator).PartitionedByClause()
	if partitioned.Entity().Width() != 0 || partitioned.OpenParen().Width() != 0 ||
		partitioned.Subquery().Width() != 0 || partitioned.CloseParen().Width() != 0 {
		t.Fatalf("partitioned-by recovery = %#v", partitioned)
	}

	validPartition := parseQueryGrammarProduction(t, grammar.makeGraphOperator,
		`make-graph Source --> Target partitioned-by Tenant (where Active | count)`).(*MakeGraphOperator).PartitionedByClause()
	if NameReferenceSimpleName(validPartition.Entity()) != "Tenant" || validPartition.Entity().Width() == 0 {
		t.Fatalf("valid partition entity = %#v", validPartition.Entity())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.makeGraphOperator,
		`make-graph Source --> Target partitioned-by Tenant.Region (where Active | count)`, 6)
}

func TestQueryGrammarPublicMakeGraphPartitionBoundary(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		text      string
		wantCount int
	}{
		{`make-graph Source --> Target with partitioned-by Tenant (count)`, 1},
		{`make-graph Source --> Target with Nodes() on NodeId, partitioned-by Tenant (count)`, 2},
	} {
		operator := parseQueryGrammarProduction(t, grammar.makeGraphOperator, test.text).(*MakeGraphOperator)
		with := operator.WithClause().(*MakeGraphWithTablesAndKeysClause)
		if with.TablesAndKeys().Count() != test.wantCount || operator.PartitionedByClause() == nil ||
			NameReferenceSimpleName(operator.PartitionedByClause().Entity()) != "Tenant" {
			t.Fatalf("partition boundary %q = tables:%d partition:%#v", test.text, with.TablesAndKeys().Count(), operator.PartitionedByClause())
		}
		missingEntry := with.TablesAndKeys().At(test.wantCount - 1).Element()
		if test.wantCount == 1 || missingEntry.Table().Width() == 0 {
			if missingEntry.Table().Width() != 0 || len(missingEntry.ContainedDiagnostics()) == 0 {
				t.Fatalf("partition boundary missing entry = %#v", missingEntry)
			}
		}
	}
}

func TestQueryGrammarPublicMakeGraphPlacementTagsAndCompletion(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := `make-graph Source --> Target`
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private":        grammar.makeGraphOperator,
		"query operator": grammar.QueryOperator,
		"following pipe": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, text).(*MakeGraphOperator); !ok {
				t.Fatal("make-graph placement produced unexpected node")
			}
		})
	}
	pipeline := parseQueryGrammarProduction(t, grammar.Expression, `Edges | make-graph Source --> Target`)
	_, stage := queryGrammarPipeStage(t, pipeline, SyntaxKindMakeGraphOperator)
	if _, ok := stage.(*MakeGraphOperator); !ok {
		t.Fatalf("make-graph pipeline stage = %T", stage)
	}

	for _, tag := range []string{
		"<make-graph>",
		"<table-and-key-clause>",
		"<make-graph-with-tables-and-keys-clause>",
		"<make-graph-with-implicit-node-id-clause>",
		"<make-graph-partitioned-by-clause>",
	} {
		if !queryGrammarParserTreeHasTag(grammar.QueryOperator, tag, 0) {
			t.Fatalf("QueryOperator parser tree is missing %s tag", tag)
		}
	}

	items := queryGrammarMakeGraphCompletionItems(grammar.QueryOperator, "")
	makeGraph, ok := completionItemByDisplayText(items, "make-graph")
	if !ok || makeGraph.Kind != CompletionKindQueryPrefix || makeGraph.BeforeText() != "make-graph" {
		t.Fatalf("make-graph completion = %#v, visible:%#v", makeGraph, completionDisplayTexts(items))
	}
	lookupIndex := completionDisplayTextIndex(items, "lookup")
	makeGraphIndex := completionDisplayTextIndex(items, "make-graph")
	makeSeriesIndex := completionDisplayTextIndex(items, "make-series")
	if lookupIndex < 0 || makeGraphIndex <= lookupIndex || makeSeriesIndex <= makeGraphIndex {
		t.Fatalf("post-pipe completion order lookup:%d make-graph:%d make-series:%d", lookupIndex, makeGraphIndex, makeSeriesIndex)
	}

	items = queryGrammarMakeGraphCompletionItems(grammar.makeGraphOperator, `make-graph Source --> Target `)
	implicit, ok := completionItemByDisplayText(items, "with_node_id")
	if !ok || implicit.BeforeText() != "with_node_id=" || implicit.AfterText() != "" {
		t.Fatalf("with_node_id completion = %#v, visible:%#v", implicit, completionDisplayTexts(items))
	}
	items = queryGrammarMakeGraphCompletionItems(grammar.makeGraphOperator, `make-graph Source `)
	if direction, ok := completionItemByDisplayText(items, "-->"); !ok || direction.BeforeText() != "-->" {
		t.Fatalf("directed edge completion = %#v, visible:%#v", direction, completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "--"); ok {
		t.Fatalf("hidden undirected edge completion exposed: %#v", completionDisplayTexts(items))
	}
}

func queryGrammarMakeGraphCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}

func completionDisplayTextIndex(items []CompletionItem, display string) int {
	for i, item := range items {
		if item.DisplayText == display {
			return i
		}
	}
	return -1
}
