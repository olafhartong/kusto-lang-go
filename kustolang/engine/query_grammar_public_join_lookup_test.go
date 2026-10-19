package kustolang

import "testing"

func TestQueryGrammarPublicJoinLookupPlacementAndRootAmbiguity(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	for _, test := range []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "join", text: "join R on Key", kind: SyntaxKindJoinOperator},
		{name: "lookup", text: "lookup R on Key", kind: SyntaxKindLookupOperator},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, production := range []struct {
				name   string
				parser ResultParser[LexicalToken, QueryOperator]
			}{
				{name: "query operator", parser: grammar.QueryOperator},
				{name: "following pipe", parser: grammar.FollowingPipeElementExpression},
			} {
				t.Run(production.name, func(t *testing.T) {
					operator := parseQueryGrammarProduction(t, production.parser, test.text)
					if operator.Kind() != test.kind {
						t.Fatalf("%s kind = %v, want %v", production.name, operator.Kind(), test.kind)
					}
				})
			}

			sub := parseQueryGrammarProduction(t, grammar.PipeSubExpression, test.text)
			if sub.Kind() != test.kind {
				t.Fatalf("pipe-sub kind = %v, want %v", sub.Kind(), test.kind)
			}

			// These keywords can also be names. The ordinary expression root gives
			// the name interpretation precedence and leaves the operator tail.
			tokens := Lex(test.text)
			root := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
			if root.Length != 1 {
				t.Fatalf("ambiguous expression root = %#v, want one-token name prefix", root)
			}
			reference, ok := root.Value.(*NameReference)
			if !ok || NameReferenceSimpleName(reference) != test.name {
				t.Fatalf("ambiguous expression root = %#v, want name %q", root.Value, test.name)
			}
		})
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		"T | join (R | where Active) on Key | lookup L on Key | take 1")
	takePipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindTakeOperator)
	lookupPipe, lookupNode := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindLookupOperator)
	joinPipe, joinNode := queryGrammarPipeStage(t, lookupPipe.Expression(), SyntaxKindJoinOperator)
	if source, ok := joinPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("join pipeline source = %#v, want T", joinPipe.Expression())
	}
	join := joinNode.(*JoinOperator)
	parenthesized, ok := join.Expression().(*ParenthesizedExpression)
	if !ok || parenthesized.Expression().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("piped join RHS = %#v, want parenthesized pipe", join.Expression())
	}
	if _, ok := lookupNode.(*LookupOperator); !ok {
		t.Fatalf("lookup pipeline operator = %T, want *LookupOperator", lookupNode)
	}
	if diagnostics := pipeline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("join/lookup pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicJoinLookupRHSShapes(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, keyword := range []string{"join", "lookup"} {
		t.Run(keyword, func(t *testing.T) {
			bare := queryGrammarJoinLookupRHS(t,
				parseQueryGrammarProduction(t, grammar.QueryOperator, keyword+" R on Key"))
			if reference, ok := bare.(*NameReference); !ok || NameReferenceSimpleName(reference) != "R" {
				t.Fatalf("bare RHS = %#v, want R name", bare)
			}

			parenthesized := queryGrammarJoinLookupRHS(t,
				parseQueryGrammarProduction(t, grammar.QueryOperator, keyword+" (R) on Key"))
			paren, ok := parenthesized.(*ParenthesizedExpression)
			if !ok || paren.Expression().Kind() != SyntaxKindNameReference {
				t.Fatalf("parenthesized RHS = %#v, want parenthesized name", parenthesized)
			}

			piped := queryGrammarJoinLookupRHS(t,
				parseQueryGrammarProduction(t, grammar.QueryOperator, keyword+" (R | where Active | take 2) on Key"))
			paren, ok = piped.(*ParenthesizedExpression)
			if !ok || paren.Expression().Kind() != SyntaxKindPipeExpression {
				t.Fatalf("piped RHS = %#v, want parenthesized pipe", piped)
			}
			if diagnostics := piped.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("piped RHS diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicJoinOnListsEqualityAndFallback(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	join := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"join R on Key, $left.Tenant == $right.Tenant and $left.User == $right.User").(*JoinOperator)
	on := join.ConditionClause().(*JoinOnClause)
	if on.OnKeyword().Kind() != SyntaxKindOnKeyword || on.Expressions().Count() != 2 {
		t.Fatalf("join on clause = keyword:%v expressions:%d", on.OnKeyword().Kind(), on.Expressions().Count())
	}
	if on.Expressions().At(0).Element().Kind() != SyntaxKindNameReference {
		t.Fatalf("first join key kind = %v, want name reference", on.Expressions().At(0).Element().Kind())
	}
	andExpression, ok := on.Expressions().At(1).Element().(*BinaryExpression)
	if !ok || andExpression.Kind() != SyntaxKindAndExpression ||
		andExpression.Left().Kind() != SyntaxKindEqualExpression ||
		andExpression.Right().Kind() != SyntaxKindEqualExpression {
		t.Fatalf("equality-and key = %#v, want And(Equal, Equal)", on.Expressions().At(1).Element())
	}
	if diagnostics := join.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("legal join key diagnostics = %#v", diagnostics)
	}

	for _, test := range []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "not equal", text: "join R on A != B", kind: SyntaxKindNotEqualExpression},
		{name: "or", text: "join R on A == B or C == D", kind: SyntaxKindOrExpression},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*JoinOperator)
			clause := operator.ConditionClause().(*JoinOnClause)
			if got := clause.Expressions().At(0).Element().Kind(); got != test.kind {
				t.Fatalf("fallback expression kind = %v, want %v", got, test.kind)
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("fallback parse diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicJoinOptionalConditionsAndLookupRequiredOn(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	withoutCondition := parseQueryGrammarProduction(t, grammar.QueryOperator, "join R").(*JoinOperator)
	if withoutCondition.ConditionClause() != nil || len(withoutCondition.ContainedDiagnostics()) != 0 {
		t.Fatalf("condition-less join = clause:%#v diagnostics:%#v",
			withoutCondition.ConditionClause(), withoutCondition.ContainedDiagnostics())
	}

	withOn := parseQueryGrammarProduction(t, grammar.QueryOperator, "join R on Key").(*JoinOperator)
	if _, ok := withOn.ConditionClause().(*JoinOnClause); !ok {
		t.Fatalf("join on condition = %T, want *JoinOnClause", withOn.ConditionClause())
	}
	withWhere := parseQueryGrammarProduction(t, grammar.QueryOperator, "join R where A == B").(*JoinOperator)
	where, ok := withWhere.ConditionClause().(*JoinWhereClause)
	if !ok || where.WhereKeyword().Kind() != SyntaxKindWhereKeyword || where.Expression().Kind() != SyntaxKindEqualExpression {
		t.Fatalf("join where condition = %#v", withWhere.ConditionClause())
	}

	lookup := parseQueryGrammarProduction(t, grammar.QueryOperator, "lookup R").(*LookupOperator)
	missing, ok := lookup.LookupClause().(*JoinOnClause)
	if !ok || missing.OnKeyword().Width() != 0 || missing.Expressions().Count() != 0 ||
		!diagnosticsContainMessage(missing.Diagnostics(), "KS006", "Missing join on condition clause") ||
		diagnosticsContain(missing.Diagnostics(), "KS001") {
		t.Fatalf("lookup missing on clause = %#v diagnostics:%#v", lookup.LookupClause(), lookup.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicJoinLookupParametersAndEqualsGuards(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	join := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"join kind=leftouter hint.remote=right hint.shufflekey=Tenant hint.strategy=broadcast hint.num_partitions=8 R on Key").(*JoinOperator)
	assertQueryGrammarJoinLookupParameters(t, join.Parameters(), []queryGrammarJoinLookupParameterExpectation{
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.remote", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.shufflekey", kind: SyntaxKindNameReference},
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.num_partitions", kind: SyntaxKindLongLiteralExpression},
	})
	lookup := parseQueryGrammarProduction(t, grammar.QueryOperator, "lookup kind=leftouter R on Key").(*LookupOperator)
	assertQueryGrammarJoinLookupParameters(t, lookup.Parameters(), []queryGrammarJoinLookupParameterExpectation{
		{name: "kind", kind: SyntaxKindTokenLiteralExpression},
	})
	if len(join.ContainedDiagnostics()) != 0 || len(lookup.ContainedDiagnostics()) != 0 {
		t.Fatalf("valid parameter diagnostics = join:%#v lookup:%#v", join.ContainedDiagnostics(), lookup.ContainedDiagnostics())
	}

	for _, keyword := range []string{"join", "lookup"} {
		t.Run(keyword+" requires equals", func(t *testing.T) {
			text := keyword + " kind inner R on Key"
			result := grammar.QueryOperator.ParseValue(NewArraySource(Lex(text)), 0)
			if result.Length != 2 {
				t.Fatalf("parameter without '=' = %#v, want keyword plus RHS prefix", result)
			}
			operator, ok := result.Value.(QueryOperator)
			if !ok {
				t.Fatalf("parameter without '=' value = %T, want QueryOperator", result.Value)
			}
			var parameters *SyntaxListOf[*NamedParameter]
			var rhs Expression
			switch node := operator.(type) {
			case *JoinOperator:
				parameters, rhs = node.Parameters(), node.Expression()
			case *LookupOperator:
				parameters, rhs = node.Parameters(), node.Expression()
			default:
				t.Fatalf("parameter without '=' operator = %T", operator)
			}
			if parameters.Count() != 0 {
				t.Fatalf("parameter without '=' count = %d, want 0", parameters.Count())
			}
			if reference, ok := rhs.(*NameReference); !ok || NameReferenceSimpleName(reference) != "kind" {
				t.Fatalf("parameter without '=' RHS = %#v, want name kind", rhs)
			}
		})
	}
}

func TestQueryGrammarPublicJoinLookupCompletionInfixAndTags(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			items := queryGrammarJoinLookupCompletionItems(production.parser, "")
			for _, keyword := range []string{"join", "lookup"} {
				item, ok := completionItemByDisplayText(items, keyword)
				if !ok || item.Kind != CompletionKindQueryPrefix || item.priority != CompletionPriorityHigh {
					t.Fatalf("%s completion = %#v, visible items:%#v", keyword, item, completionDisplayTexts(items))
				}
			}
		})
	}

	items := queryGrammarJoinLookupCompletionItems(grammar.QueryOperator, "join R on A ")
	for _, infix := range []string{"==", "and"} {
		item, ok := completionItemByDisplayText(items, infix)
		if !ok || item.Kind != CompletionKindScalarInfix {
			t.Fatalf("join infix %q = %#v, visible items:%#v", infix, item, completionDisplayTexts(items))
		}
	}
	for _, hidden := range []string{"!=", "or"} {
		if item, ok := completionItemByDisplayText(items, hidden); ok {
			t.Fatalf("hidden fallback infix %q became visible: %#v", hidden, item)
		}
	}

	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<join>", 0) {
		t.Fatal("QueryOperator parser tree has no <join> tag")
	}
	if !queryGrammarParserTreeHasTag(grammar.QueryOperator, "<lookup>", 0) {
		t.Fatal("QueryOperator parser tree has no <lookup> tag")
	}
}

func TestQueryGrammarPublicJoinLookupRecoveryAndTailBoundaries(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	joinMissingRHS := parseQueryGrammarProduction(t, grammar.QueryOperator, "join").(*JoinOperator)
	assertQueryGrammarJoinLookupMissingExpression(t, joinMissingRHS.Expression(), "join RHS")
	lookupMissingRHS := parseQueryGrammarProduction(t, grammar.QueryOperator, "lookup").(*LookupOperator)
	assertQueryGrammarJoinLookupMissingExpression(t, lookupMissingRHS.Expression(), "lookup RHS")
	if on, ok := lookupMissingRHS.LookupClause().(*JoinOnClause); !ok || on.Expressions().Count() != 0 ||
		!diagnosticsContain(on.Diagnostics(), "KS006") {
		t.Fatalf("lookup missing clause = %#v", lookupMissingRHS.LookupClause())
	}

	leadingOn := grammar.QueryOperator.ParseValue(NewArraySource(Lex("lookup on Key")), 0)
	lookupLeadingOn, ok := leadingOn.Value.(*LookupOperator)
	if !ok || leadingOn.Length != 2 {
		t.Fatalf("lookup leading on ambiguity = %#v, want lookup/on prefix", leadingOn)
	}
	if reference, ok := lookupLeadingOn.Expression().(*NameReference); !ok || NameReferenceSimpleName(reference) != "on" {
		t.Fatalf("lookup leading on RHS = %#v, want name on", lookupLeadingOn.Expression())
	}

	for _, text := range []string{"join R on", "lookup R on"} {
		t.Run(text, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, text)
			var clause *JoinOnClause
			switch node := operator.(type) {
			case *JoinOperator:
				clause = node.ConditionClause().(*JoinOnClause)
			case *LookupOperator:
				clause = node.LookupClause().(*JoinOnClause)
			}
			if clause.Expressions().Count() != 1 {
				t.Fatalf("missing on expression count = %d, want 1", clause.Expressions().Count())
			}
			assertQueryGrammarJoinLookupMissingExpression(t, clause.Expressions().At(0).Element(), "on expression")
		})
	}

	whereMissing := parseQueryGrammarProduction(t, grammar.QueryOperator, "join R where").(*JoinOperator)
	where := whereMissing.ConditionClause().(*JoinWhereClause)
	assertQueryGrammarJoinLookupMissingExpression(t, where.Expression(), "where expression")

	for _, test := range []struct {
		name         string
		text         string
		count        int
		missingIndex int
	}{
		{name: "leading comma", text: "join R on , A", count: 2, missingIndex: 0},
		{name: "middle comma", text: "join R on A,, B", count: 3, missingIndex: 1},
		{name: "trailing comma", text: "join R on A,", count: 2, missingIndex: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			join := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*JoinOperator)
			on := join.ConditionClause().(*JoinOnClause)
			if on.Expressions().Count() != test.count {
				t.Fatalf("malformed key count = %d, want %d", on.Expressions().Count(), test.count)
			}
			assertQueryGrammarJoinLookupMissingExpression(t, on.Expressions().At(test.missingIndex).Element(), "comma key")
		})
	}

	for _, test := range []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "missing equality RHS", text: "join R on A ==", kind: SyntaxKindEqualExpression},
		{name: "missing and RHS", text: "join R on A and", kind: SyntaxKindAndExpression},
	} {
		t.Run(test.name, func(t *testing.T) {
			join := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*JoinOperator)
			binary := join.ConditionClause().(*JoinOnClause).Expressions().At(0).Element().(*BinaryExpression)
			if binary.Kind() != test.kind {
				t.Fatalf("malformed binary kind = %v, want %v", binary.Kind(), test.kind)
			}
			assertQueryGrammarJoinLookupMissingExpression(t, binary.Right(), test.name)
		})
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "join R on Key | where Active", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "lookup R on Key | take 1", 4)
	missingComma := parseQueryGrammarProduction(t, grammar.QueryOperator, "join R on Key Extra").(*JoinOperator)
	missingCommaKeys := missingComma.ConditionClause().(*JoinOnClause).Expressions()
	if missingCommaKeys.Count() != 2 || missingCommaKeys.At(0).Separator() == nil ||
		!diagnosticsContainMessage(missingCommaKeys.At(0).Separator().Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("join missing-comma keys = %#v", missingCommaKeys)
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "lookup (R) on Key )", 6)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "lookup R where Active", 2)
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "R on Key")
}

type queryGrammarJoinLookupParameterExpectation struct {
	name string
	kind SyntaxKind
}

func assertQueryGrammarJoinLookupParameters(t *testing.T, parameters *SyntaxListOf[*NamedParameter], want []queryGrammarJoinLookupParameterExpectation) {
	t.Helper()
	if parameters.Count() != len(want) {
		t.Fatalf("parameter count = %d, want %d", parameters.Count(), len(want))
	}
	for i, expected := range want {
		parameter := parameters.At(i)
		if NameDeclarationSimpleName(parameter.Name()) != expected.name ||
			parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 ||
			parameter.Expression().Kind() != expected.kind {
			t.Fatalf("parameter %d = %#v, want %q/%v", i, parameter, expected.name, expected.kind)
		}
	}
}

func queryGrammarJoinLookupRHS(t *testing.T, operator QueryOperator) Expression {
	t.Helper()
	switch node := operator.(type) {
	case *JoinOperator:
		return node.Expression()
	case *LookupOperator:
		return node.Expression()
	default:
		t.Fatalf("join/lookup operator = %T", operator)
		return nil
	}
}

func queryGrammarJoinLookupCompletionItems(parser ResultParser[LexicalToken, QueryOperator], text string) []CompletionItem {
	tokens := Lex(text)
	items := []CompletionItem{}
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, parser, 0) {
		items = append(items, syntaxCompletionItemsFromAnnotations(annotated.Annotations())...)
	}
	return items
}

func assertQueryGrammarJoinLookupMissingExpression(t *testing.T, expression Expression, context string) {
	t.Helper()
	if expression == nil || expression.Width() != 0 ||
		!diagnosticsContainMessage(expression.Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing %s = %#v diagnostics:%#v", context, expression, expression.ContainedDiagnostics())
	}
}
