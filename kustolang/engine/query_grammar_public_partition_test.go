package kustolang

import "testing"

func TestQueryGrammarPublicPartitionParametersPathsAndScopes(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.partitionOperator,
		`partition hint.concurrency=4 hint.spread=2 hint.materialized=true hint.shufflekey=Tenant hint.strategy=shuffle by Account.Region["Name"] in cluster("help") (summarize Count=count())`).(*PartitionOperator)

	wantParameters := []struct {
		name string
		kind SyntaxKind
	}{
		{name: "hint.concurrency", kind: SyntaxKindLongLiteralExpression},
		{name: "hint.spread", kind: SyntaxKindLongLiteralExpression},
		{name: "hint.materialized", kind: SyntaxKindBooleanLiteralExpression},
		{name: "hint.shufflekey", kind: SyntaxKindNameReference},
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
	}
	if operator.Parameters().Count() != len(wantParameters) {
		t.Fatalf("partition parameters = %d, want %d", operator.Parameters().Count(), len(wantParameters))
	}
	for i, want := range wantParameters {
		parameter := operator.Parameters().At(i)
		if NameDeclarationSimpleName(parameter.Name()) != want.name || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != want.kind {
			t.Fatalf("partition parameter %d = %#v, want %q/%v", i, parameter, want.name, want.kind)
		}
	}
	if operator.ByKeyword().Kind() != SyntaxKindByKeyword || operator.ByExpression().Kind() != SyntaxKindElementExpression {
		t.Fatalf("partition by path = keyword:%v expression:%T/%v", operator.ByKeyword().Kind(), operator.ByExpression(), operator.ByExpression().Kind())
	}
	if operator.Scope() == nil || operator.Scope().InKeyword().Kind() != SyntaxKindInKeyword ||
		operator.Scope().Expression().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("partition function scope = %#v", operator.Scope())
	}
	operand, ok := operator.Operand().(*PartitionSubquery)
	if !ok || operand.Subquery().Kind() != SyntaxKindSummarizeOperator || operand.CloseParen().Width() == 0 {
		t.Fatalf("partition subquery operand = %#v", operator.Operand())
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid partition diagnostics = %#v", diagnostics)
	}

	dynamicScope := parseQueryGrammarProduction(t, grammar.partitionOperator,
		`partition by ["Tenant"] in dynamic(["one", "two"]) (where Enabled)`).(*PartitionOperator)
	if dynamicScope.ByExpression().Kind() != SyntaxKindNameReference || dynamicScope.Scope() == nil ||
		dynamicScope.Scope().Expression().Kind() != SyntaxKindDynamicExpression {
		t.Fatalf("partition bracket path/dynamic scope = by:%#v scope:%#v", dynamicScope.ByExpression(), dynamicScope.Scope())
	}
	if diagnostics := dynamicScope.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid dynamic-scope partition diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicPartitionOperandAlternatives(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	braced := parseQueryGrammarProduction(t, grammar.partitionOperator,
		`partition by Tenant { Source | where Enabled | project Tenant }`).(*PartitionOperator)
	query, ok := braced.Operand().(*PartitionQuery)
	if !ok || query.OpenBrace().Kind() != SyntaxKindOpenBraceToken || query.Query().Kind() != SyntaxKindPipeExpression ||
		query.CloseBrace().Kind() != SyntaxKindCloseBraceToken {
		t.Fatalf("partition braced full expression = %#v", braced.Operand())
	}

	pipeSub := parseQueryGrammarProduction(t, grammar.partitionOperator,
		`partition by Tenant (where Enabled | project Tenant)`).(*PartitionOperator).Operand().(*PartitionSubquery)
	if pipeSub.Subquery().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("partition PipeSubExpression operand = %T/%v", pipeSub.Subquery(), pipeSub.Subquery().Kind())
	}

	fallback := parseQueryGrammarProduction(t, grammar.partitionOperator,
		`partition by Tenant (Source | where Enabled)`).(*PartitionOperator).Operand().(*PartitionSubquery)
	if fallback.Subquery().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("partition general-expression fallback = %T/%v", fallback.Subquery(), fallback.Subquery().Kind())
	}
	for name, operand := range map[string]*PartitionSubquery{"pipe-sub": pipeSub, "fallback": fallback} {
		if diagnostics := operand.ContainedDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%s partition operand diagnostics = %#v", name, diagnostics)
		}
	}
}

func TestQueryGrammarPublicPartitionByParametersEntityIDAndContextualSubquery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	operator := parseQueryGrammarProduction(t, grammar.partitionByOperator,
		`__partitionby hint.concurrency=4 hint.spread=2 hint.materialized=false hint.shufflekey=Tenant hint.strategy=native Account.Region["Name"] id "slice-1" (Other | where Enabled | summarize Count=count())`).(*PartitionByOperator)

	if operator.PartitionByKeyword().Kind() != SyntaxKindPartitionByKeyword || operator.Parameters().Count() != 5 {
		t.Fatalf("partition-by keyword/parameters = %v/%d", operator.PartitionByKeyword().Kind(), operator.Parameters().Count())
	}
	if operator.Entity().Kind() != SyntaxKindElementExpression {
		t.Fatalf("partition-by entity path = %T/%v", operator.Entity(), operator.Entity().Kind())
	}
	id := operator.IdClause()
	if id == nil || id.IdKeyword().Kind() != SyntaxKindIdKeyword || id.Value().Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("partition-by id clause = %#v", id)
	}
	if operator.OpenParen().Kind() != SyntaxKindOpenParenToken || operator.Subquery().Kind() != SyntaxKindPipeExpression ||
		operator.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("partition-by contextual subquery = %#v", operator)
	}
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid partition-by diagnostics = %#v", diagnostics)
	}

	bracketEntity := parseQueryGrammarProduction(t, grammar.partitionByOperator,
		`__partitionby ["Tenant"] (where Enabled)`).(*PartitionByOperator)
	if bracketEntity.Entity().Kind() != SyntaxKindNameReference || bracketEntity.IdClause() != nil ||
		bracketEntity.Subquery().Kind() != SyntaxKindFilterOperator {
		t.Fatalf("partition-by bracket entity/no-id = %#v", bracketEntity)
	}
}

func TestQueryGrammarPublicPartitionPlacementTagsAndCompletions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private":        grammar.partitionOperator,
		"query-operator": grammar.QueryOperator,
		"following-pipe": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, "partition by Tenant (count)").(*PartitionOperator); !ok {
				t.Fatalf("partition placement produced unexpected node")
			}
		})
	}
	if _, ok := parseQueryGrammarProduction(t, grammar.QueryOperator, "__partitionby Tenant (count)").(*PartitionByOperator); !ok {
		t.Fatal("QueryOperator did not produce PartitionByOperator")
	}
	for name, parser := range map[string]ResultParser[LexicalToken, QueryOperator]{
		"private-partition-by":   grammar.partitionByOperator,
		"following-partition-by": grammar.FollowingPipeElementExpression,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseQueryGrammarProduction(t, parser, "__partitionby Tenant (count)").(*PartitionByOperator); !ok {
				t.Fatal("partition-by placement produced unexpected node")
			}
		})
	}
	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		"Source | partition by Tenant (where Enabled) | take 1")
	takePipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindTakeOperator)
	partitionPipe, node := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindPartitionOperator)
	if source, ok := partitionPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Source" {
		t.Fatalf("partition pipeline source = %#v", partitionPipe.Expression())
	}
	if _, ok := node.(*PartitionOperator); !ok {
		t.Fatalf("partition pipeline operator = %T", node)
	}
	partitionByPipeline := parseQueryGrammarProduction(t, grammar.Expression,
		"Source | __partitionby Tenant (where Enabled) | take 1")
	partitionByTakePipe, _ := queryGrammarPipeStage(t, partitionByPipeline, SyntaxKindTakeOperator)
	partitionByPipe, partitionByNode := queryGrammarPipeStage(t, partitionByTakePipe.Expression(), SyntaxKindPartitionByOperator)
	if source, ok := partitionByPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "Source" {
		t.Fatalf("partition-by pipeline source = %#v", partitionByPipe.Expression())
	}
	if _, ok := partitionByNode.(*PartitionByOperator); !ok {
		t.Fatalf("partition-by pipeline operator = %T", partitionByNode)
	}

	items := []CompletionItem{}
	for _, parser := range FindAnnotatedParsers(NewArraySource(Lex("")), 0, grammar.QueryOperator, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
	}
	item, ok := completionItemByDisplayText(items, "partition")
	if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityLow {
		t.Fatalf("partition completion = %#v, visible:%#v", item, completionDisplayTexts(items))
	}
	if _, ok := completionItemByDisplayText(items, "__partitionby"); ok {
		t.Fatalf("hidden partition-by completion became visible: %#v", completionDisplayTexts(items))
	}
	if grammar.partitionOperator.Tag() != "<partition>" || !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<partition>", 0) {
		t.Fatalf("partition tags = private:%q tree:%v", grammar.partitionOperator.Tag(), queryGrammarParserTreeHasTag(grammar.QueryOperator, "<partition>", 0))
	}
	if grammar.partitionByOperator.Tag() != "" {
		t.Fatalf("hidden partition-by unexpectedly tagged %q", grammar.partitionByOperator.Tag())
	}
}

func TestQueryGrammarPublicPartitionMissingParts(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingBy := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition Tenant (count)").(*PartitionOperator)
	if missingBy.ByKeyword().Width() != 0 || !diagnosticsContainMessage(missingBy.ByKeyword().Diagnostics(), "KS005", "Expected: by") ||
		NameReferenceSimpleName(missingBy.ByExpression().(*NameReference)) != "Tenant" {
		t.Fatalf("partition missing by recovery = %#v", missingBy)
	}

	missingEntity := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition by (count)").(*PartitionOperator)
	if missingEntity.ByExpression().Width() != 0 || !diagnosticsContainMessage(missingEntity.ByExpression().Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("partition missing entity recovery = %#v diagnostics:%#v", missingEntity.ByExpression(), missingEntity.ByExpression().Diagnostics())
	}

	missingScope := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition by Tenant in (count)").(*PartitionOperator)
	if missingScope.Scope() == nil || missingScope.Scope().Expression().Width() != 0 ||
		!diagnosticsContainMessage(missingScope.Scope().Expression().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("partition missing scope recovery = %#v", missingScope.Scope())
	}

	missingOperand := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition by Tenant").(*PartitionOperator)
	operand := missingOperand.Operand().(*PartitionSubquery)
	if operand.OpenParen().Width() != 0 || operand.Subquery().Width() != 0 || operand.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(operand.Subquery().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("partition missing operand recovery = %#v", operand)
	}

	missingBrace := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition by Tenant { count").(*PartitionOperator).Operand().(*PartitionQuery)
	if missingBrace.CloseBrace().Width() != 0 || !diagnosticsContainMessage(missingBrace.CloseBrace().Diagnostics(), "KS005", "Expected: }") {
		t.Fatalf("partition missing close brace = %#v", missingBrace.CloseBrace())
	}
	missingParen := parseQueryGrammarProduction(t, grammar.partitionOperator, "partition by Tenant (count").(*PartitionOperator).Operand().(*PartitionSubquery)
	if missingParen.CloseParen().Width() != 0 || !diagnosticsContainMessage(missingParen.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("partition missing close paren = %#v", missingParen.CloseParen())
	}
}

func TestQueryGrammarPublicPartitionByMissingParts(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingEntity := parseQueryGrammarProduction(t, grammar.partitionByOperator, "__partitionby (count)").(*PartitionByOperator)
	if missingEntity.Entity().Width() != 0 || !diagnosticsContainMessage(missingEntity.Entity().Diagnostics(), "KS006", "Missing name") {
		t.Fatalf("partition-by missing entity = %#v diagnostics:%#v", missingEntity.Entity(), missingEntity.Entity().Diagnostics())
	}

	missingOperand := parseQueryGrammarProduction(t, grammar.partitionByOperator, "__partitionby Tenant").(*PartitionByOperator)
	if missingOperand.OpenParen().Width() != 0 || missingOperand.Subquery().Width() != 0 || missingOperand.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingOperand.Subquery().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("partition-by missing operand = %#v", missingOperand)
	}

	missingSubquery := parseQueryGrammarProduction(t, grammar.partitionByOperator, "__partitionby Tenant ()").(*PartitionByOperator)
	if missingSubquery.Subquery().Width() != 0 ||
		!diagnosticsContainMessage(missingSubquery.Subquery().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("partition-by missing subquery = %#v diagnostics:%#v", missingSubquery.Subquery(), missingSubquery.Subquery().Diagnostics())
	}

	missingParen := parseQueryGrammarProduction(t, grammar.partitionByOperator, "__partitionby Tenant (count").(*PartitionByOperator)
	if missingParen.CloseParen().Width() != 0 || !diagnosticsContainMessage(missingParen.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("partition-by missing close paren = %#v", missingParen.CloseParen())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.partitionByOperator, "__partitionby Tenant id (count)", 3)
}

func TestQueryGrammarPublicPartitionTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionOperator, "partition by Tenant (count) | where true", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionOperator, "partition by Tenant { count } trailing", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionOperator, "partition by Tenant (count) trailing", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionOperator, "partition by Tenant in dynamic([1]) (count) trailing", 13)

	assertQueryGrammarProductionLeavesTail(t, grammar.partitionByOperator, "__partitionby Tenant (count) | where true", 5)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionByOperator, `__partitionby Tenant id "slice" (count) trailing`, 7)
	assertQueryGrammarProductionLeavesTail(t, grammar.partitionByOperator, "__partitionby Tenant (count) trailing", 5)
}
