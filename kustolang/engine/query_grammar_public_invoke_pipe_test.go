package kustolang

import "testing"

func TestQueryGrammarPublicInvokeSimpleFunctionCallAST(t *testing.T) {
	invoke := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"invoke Transform(1, Value)").(*InvokeOperator)

	if invoke.InvokeKeyword().Kind() != SyntaxKindInvokeKeyword || invoke.ChildCount() != 2 {
		t.Fatalf("invoke keyword/children = %v/%d, want InvokeKeyword/2", invoke.InvokeKeyword().Kind(), invoke.ChildCount())
	}
	call := assertQueryGrammarInvokeCall(t, invoke.Function(), "Transform", 2)
	if call.ArgumentList().Expressions().At(0).Element().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("first argument kind = %v, want LongLiteralExpression", call.ArgumentList().Expressions().At(0).Element().Kind())
	}
	value, ok := call.ArgumentList().Expressions().At(1).Element().(*NameReference)
	if !ok || NameReferenceSimpleName(value) != "Value" {
		t.Fatalf("second argument = %#v, want Value name reference", call.ArgumentList().Expressions().At(1).Element())
	}
	if diagnostics := invoke.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("invoke diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicInvokeDottedCompositeFunctionCallAST(t *testing.T) {
	invoke := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"invoke cluster('help').database('Samples').Transform(1)").(*InvokeOperator)

	outer, ok := invoke.Function().(*PathExpression)
	if !ok || outer.ChildCount() != 3 || outer.Dot().Kind() != SyntaxKindDotToken {
		t.Fatalf("outer invoke path = %#v, want three-child PathExpression", invoke.Function())
	}
	assertQueryGrammarInvokeCall(t, outer.Selector(), "Transform", 1)

	inner, ok := outer.Expression().(*PathExpression)
	if !ok || inner.ChildCount() != 3 || inner.Dot().Kind() != SyntaxKindDotToken {
		t.Fatalf("inner invoke path = %#v, want three-child PathExpression", outer.Expression())
	}
	assertQueryGrammarInvokeCall(t, inner.Expression(), "cluster", 1)
	assertQueryGrammarInvokeCall(t, inner.Selector(), "database", 1)
	if diagnostics := invoke.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("dotted invoke diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicInvokeKnownKeywordAndBracketedFunctionNames(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	known := parseQueryGrammarProduction(t, grammar.QueryOperator, "invoke count()").(*InvokeOperator)
	assertQueryGrammarInvokeCall(t, known.Function(), "count", 0)

	tokens := Lex(`invoke ["Transform"()`)
	result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
	invoke, ok := result.Value.(*InvokeOperator)
	if !ok || result.Length != len(tokens)-1 {
		t.Fatalf("incomplete bracketed invoke = %#v", result)
	}
	call := assertQueryGrammarInvokeCall(t, invoke.Function(), "Transform", 0)
	bracketed, ok := call.Name().Name().(*BracketedName)
	if !ok || bracketed.CloseBracket().Width() != 0 ||
		!diagnosticsContainMessage(bracketed.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("incomplete bracketed function name = %#v diagnostics:%#v", call.Name().Name(), call.Name().ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInvokeNestedPipelinesAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"T | invoke First(1) | invoke database('db').Second(Value)")

	secondPipe, secondNode := queryGrammarPipeStage(t, expression, SyntaxKindInvokeOperator)
	firstPipe, firstNode := queryGrammarPipeStage(t, secondPipe.Expression(), SyntaxKindInvokeOperator)
	if source, ok := firstPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("nested invoke source = %#v, want T", firstPipe.Expression())
	}
	assertQueryGrammarInvokeCall(t, firstNode.(*InvokeOperator).Function(), "First", 1)
	secondPath, ok := secondNode.(*InvokeOperator).Function().(*PathExpression)
	if !ok {
		t.Fatalf("second invoke function = %T, want *PathExpression", secondNode.(*InvokeOperator).Function())
	}
	assertQueryGrammarInvokeCall(t, secondPath.Expression(), "database", 1)
	assertQueryGrammarInvokeCall(t, secondPath.Selector(), "Second", 1)

	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, production.parser, "invoke Keep()").(*InvokeOperator)
			assertQueryGrammarInvokeCall(t, operator.Function(), "Keep", 0)
		})
	}
}

func TestQueryGrammarPublicInvokeLowPriorityCompletion(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, production := range []struct {
		name   string
		parser ResultParser[LexicalToken, QueryOperator]
	}{
		{name: "query operator", parser: grammar.QueryOperator},
		{name: "following pipe element", parser: grammar.FollowingPipeElementExpression},
	} {
		t.Run(production.name, func(t *testing.T) {
			parsers := FindAnnotatedParsers(NewArraySource(Lex("")), 0, production.parser, 0)
			items := []CompletionItem{}
			for _, parser := range parsers {
				items = append(items, syntaxCompletionItemsFromAnnotations(parser.Annotations())...)
			}
			invoke, ok := completionItemByDisplayText(items, "invoke")
			if !ok || invoke.Kind != CompletionKindQueryPrefix || invoke.priority != CompletionPriorityLow {
				t.Fatalf("invoke completion = %#v, visible items:%#v", invoke, completionDisplayTexts(items))
			}
		})
	}
}

func TestQueryGrammarPublicInvokeRecoveryRejectionAndTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missing := parseQueryGrammarProduction(t, grammar.QueryOperator, "invoke").(*InvokeOperator)
	if missing.Function().Kind() != SyntaxKindNameReference || missing.Function().Width() != 0 ||
		!diagnosticsContainMessage(missing.Function().Diagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing invoke function = %#v diagnostics:%#v", missing.Function(), missing.Function().Diagnostics())
	}

	for _, text := range []string{"invoke Transform", "invoke Transform.Member"} {
		t.Run(text, func(t *testing.T) {
			tokens := Lex(text)
			result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
			invoke, ok := result.Value.(*InvokeOperator)
			if result.Length != 1 || !ok {
				t.Fatalf("non-call recovery %q = %#v, want invoke-only prefix", text, result)
			}
			if invoke.Function().Width() != 0 || !diagnosticsContainMessage(invoke.Function().Diagnostics(), "KS006", "Missing expression") {
				t.Fatalf("non-call recovery function %q = %#v diagnostics:%#v", text, invoke.Function(), invoke.Function().Diagnostics())
			}
		})
	}

	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "Transform()")
	badResult := grammar.FollowingPipeElementExpression.ParseValue(NewArraySource(Lex("Transform()")), 0)
	if badResult.Length != 1 {
		t.Fatalf("following-pipe non-call length = %d, want identifier-only prefix", badResult.Length)
	}
	bad, ok := badResult.Value.(*BadQueryOperator)
	if !ok || !diagnosticsContain(bad.Diagnostics(), "KS176") {
		t.Fatalf("following-pipe non-call = %#v, want diagnosed *BadQueryOperator", badResult.Value)
	}
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "invoke Keep() Extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "invoke Keep(), Extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "invoke Keep().Member", 4)
}

func assertQueryGrammarInvokeCall(t *testing.T, expression Expression, name string, argumentCount int) *FunctionCallExpression {
	t.Helper()
	call, ok := expression.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("invoke function component = %T, want *FunctionCallExpression", expression)
	}
	if call.ChildCount() != 2 || NameReferenceSimpleName(call.Name()) != name {
		t.Fatalf("invoke call children/name = %d/%q, want 2/%q", call.ChildCount(), NameReferenceSimpleName(call.Name()), name)
	}
	arguments := call.ArgumentList()
	if arguments.ChildCount() != 3 || arguments.OpenParen().Kind() != SyntaxKindOpenParenToken ||
		arguments.CloseParen().Kind() != SyntaxKindCloseParenToken || arguments.Expressions().Count() != argumentCount {
		t.Fatalf("invoke call %q argument list = %#v, want three children and %d arguments", name, arguments, argumentCount)
	}
	return call
}
