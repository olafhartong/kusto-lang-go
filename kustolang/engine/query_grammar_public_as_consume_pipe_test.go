package kustolang

import "testing"

func TestQueryGrammarPublicAsAndConsumeOperatorAST(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	as := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"as hint.materialized=true ['Scoped Rows']").(*AsOperator)
	if as.AsKeyword().Kind() != SyntaxKindAsKeyword {
		t.Fatalf("as keyword kind = %v", as.AsKeyword().Kind())
	}
	if as.Parameters().Count() != 1 {
		t.Fatalf("as parameters = %d, want 1", as.Parameters().Count())
	}
	assertQueryGrammarAsConsumeParameter(t, as.Parameters().At(0), "hint.materialized", SyntaxKindBooleanLiteralExpression)
	if NameDeclarationSimpleName(as.Name()) != "Scoped Rows" || as.Name().Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("as name = %#v, want bracketed Scoped Rows", as.Name())
	}
	if diagnostics := as.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("as diagnostics = %#v", diagnostics)
	}

	consume := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"consume decodeblocks=false").(*ConsumeOperator)
	if consume.ConsumeKeyword().Kind() != SyntaxKindConsumeKeyword {
		t.Fatalf("consume keyword kind = %v", consume.ConsumeKeyword().Kind())
	}
	if consume.Parameters().Count() != 1 {
		t.Fatalf("consume parameters = %d, want 1", consume.Parameters().Count())
	}
	assertQueryGrammarAsConsumeParameter(t, consume.Parameters().At(0), "decodeblocks", SyntaxKindBooleanLiteralExpression)
	if diagnostics := consume.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("consume diagnostics = %#v", diagnostics)
	}

	plainAs := parseQueryGrammarProduction(t, grammar.QueryOperator, "as Alias").(*AsOperator)
	if plainAs.Parameters().Count() != 0 || NameDeclarationSimpleName(plainAs.Name()) != "Alias" {
		t.Fatalf("plain as operator = %#v, want Alias without parameters", plainAs)
	}
	emptyConsume := parseQueryGrammarProduction(t, grammar.QueryOperator, "consume").(*ConsumeOperator)
	if emptyConsume.Parameters().Count() != 0 {
		t.Fatalf("parameterless consume parameters = %d, want 0", emptyConsume.Parameters().Count())
	}
}

func TestQueryGrammarPublicAsAndConsumePipelinePlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"T | as hint.materialized=false Snapshot | consume decodeblocks=true")

	consumePipe, consumeNode := queryGrammarPipeStage(t, expression, SyntaxKindConsumeOperator)
	asPipe, asNode := queryGrammarPipeStage(t, consumePipe.Expression(), SyntaxKindAsOperator)
	if source, ok := asPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("pipe source = %#v, want T", asPipe.Expression())
	}
	if NameDeclarationSimpleName(asNode.(*AsOperator).Name()) != "Snapshot" {
		t.Fatalf("as pipeline name = %#v, want Snapshot", asNode.(*AsOperator).Name())
	}
	if consumeNode.(*ConsumeOperator).Parameters().Count() != 1 {
		t.Fatalf("consume pipeline parameters = %d, want 1", consumeNode.(*ConsumeOperator).Parameters().Count())
	}

	for _, test := range []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "as following pipe element", text: "as Alias", kind: SyntaxKindAsOperator},
		{name: "consume following pipe element", text: "consume", kind: SyntaxKindConsumeOperator},
	} {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.FollowingPipeElementExpression, test.text)
			if operator.Kind() != test.kind {
				t.Fatalf("following pipe operator kind = %v, want %v", operator.Kind(), test.kind)
			}
		})
	}
}

func TestQueryGrammarPublicAsAndConsumeRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingName := parseQueryGrammarProduction(t, grammar.QueryOperator, "as").(*AsOperator)
	if missingName.Name().Width() != 0 || !diagnosticsContainMessage(missingName.Name().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("missing as name = %#v with diagnostics %#v", missingName.Name(), missingName.Name().ContainedDiagnostics())
	}
	keywordResult := grammar.QueryOperator.ParseValue(NewArraySource(Lex("as where")), 0)
	keywordName := keywordResult.Value.(*AsOperator).Name()
	if keywordResult.Length != 1 || !diagnosticsContainMessage(keywordName.ContainedDiagnostics(), "KS006", "Missing name: If the keyword 'where' is intended be used as the name, it needs to be bracketted as ['where'].") {
		t.Fatalf("keyword as name parse = %#v diagnostics:%#v", keywordResult, keywordName.ContainedDiagnostics())
	}

	missingAsValue := parseQueryGrammarProduction(t, grammar.QueryOperator, "as hint.materialized=").(*AsOperator)
	if missingAsValue.Parameters().Count() != 1 {
		t.Fatalf("as missing-value parameters = %d, want 1", missingAsValue.Parameters().Count())
	}
	asParameter := missingAsValue.Parameters().At(0)
	if asParameter.Expression().Kind() != SyntaxKindBooleanLiteralExpression ||
		!diagnosticsContainMessage(asParameter.Expression().ContainedDiagnostics(), "KS006", "Missing boolean") {
		t.Fatalf("as missing parameter value = %#v", asParameter.Expression())
	}
	if !diagnosticsContainMessage(missingAsValue.Name().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("as missing alias diagnostics = %#v", missingAsValue.Name().ContainedDiagnostics())
	}

	missingConsumeValue := parseQueryGrammarProduction(t, grammar.QueryOperator, "consume decodeblocks=").(*ConsumeOperator)
	if missingConsumeValue.Parameters().Count() != 1 {
		t.Fatalf("consume missing-value parameters = %d, want 1", missingConsumeValue.Parameters().Count())
	}
	consumeValue := missingConsumeValue.Parameters().At(0).Expression()
	if consumeValue.Kind() != SyntaxKindBooleanLiteralExpression ||
		!diagnosticsContainMessage(consumeValue.ContainedDiagnostics(), "KS006", "Missing boolean") {
		t.Fatalf("consume missing parameter value = %#v", consumeValue)
	}

	missingConsumeEqual := parseQueryGrammarProduction(t, grammar.QueryOperator, "consume decodeblocks true").(*ConsumeOperator)
	if missingConsumeEqual.Parameters().Count() != 1 {
		t.Fatalf("consume missing-equal parameters = %d, want 1", missingConsumeEqual.Parameters().Count())
	}
	consumeParameter := missingConsumeEqual.Parameters().At(0)
	if consumeParameter.EqualToken().Width() != 0 ||
		!diagnosticsContainMessage(consumeParameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
		t.Fatalf("consume missing equal token/diagnostics = %#v/%#v", consumeParameter.EqualToken(), consumeParameter.EqualToken().Diagnostics())
	}
	if consumeParameter.Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("consume recovered value kind = %v, want boolean literal", consumeParameter.Expression().Kind())
	}
}

func TestQueryGrammarPublicAsAndConsumeRecognizeOtherKnownParameters(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	as := parseQueryGrammarProduction(t, grammar.QueryOperator, "as hint.strategy=shuffle Alias").(*AsOperator)
	if as.Parameters().Count() != 1 || NameDeclarationSimpleName(as.Parameters().At(0).Name()) != "hint.strategy" ||
		as.Parameters().At(0).Expression().Kind() != SyntaxKindTokenLiteralExpression || NameDeclarationSimpleName(as.Name()) != "Alias" {
		t.Fatalf("as known parameter/name = %#v/%#v", as.Parameters(), as.Name())
	}

	consume := parseQueryGrammarProduction(t, grammar.QueryOperator, "consume hint.materialized=true").(*ConsumeOperator)
	if consume.Parameters().Count() != 1 || NameDeclarationSimpleName(consume.Parameters().At(0).Name()) != "hint.materialized" ||
		consume.Parameters().At(0).Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("consume known parameter = %#v", consume.Parameters())
	}
}

func TestQueryGrammarPublicAsAndConsumeUnsupportedTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	// As parameters use equals-needed lookahead, so a missing '=' leaves the
	// would-be parameter and value outside the operator after recovering the name.
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "as hint.materialized true Alias", 1)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "as Alias Extra", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "as Alias, Extra", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "consume decodeblocks=true Extra", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "consume, Extra", 1)
}

func assertQueryGrammarAsConsumeParameter(t *testing.T, parameter *NamedParameter, name string, valueKind SyntaxKind) {
	t.Helper()
	if NameDeclarationSimpleName(parameter.Name()) != name {
		t.Fatalf("parameter name = %q, want %q", NameDeclarationSimpleName(parameter.Name()), name)
	}
	if parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 {
		t.Fatalf("parameter equal token = %#v, want present equal", parameter.EqualToken())
	}
	if parameter.Expression().Kind() != valueKind {
		t.Fatalf("parameter value kind = %v, want %v", parameter.Expression().Kind(), valueKind)
	}
}
