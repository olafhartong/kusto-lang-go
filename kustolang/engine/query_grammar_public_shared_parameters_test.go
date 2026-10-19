package kustolang

import "testing"

func TestQueryGrammarPublicSharedParametersDeclaredAndKnown(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	distinct := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"distinct hint.shufflekey=Tenant hint.strategy=shuffle hint.num_partitions=4 hint.materialized=true Value").(*DistinctOperator)
	assertQueryGrammarSharedParameters(t, distinct.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.shufflekey", kind: SyntaxKindNameReference},
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
		{name: "hint.num_partitions", kind: SyntaxKindLongLiteralExpression},
		{name: "hint.materialized", kind: SyntaxKindBooleanLiteralExpression},
	})

	sort := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"sort hint.strategy=splitBlock decodeblocks=false by Score").(*SortOperator)
	assertQueryGrammarSharedParameters(t, sort.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
		{name: "decodeblocks", kind: SyntaxKindBooleanLiteralExpression},
	})

	top := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"top hint.progressive_top=true hint.strategy=shuffle 5 by Score").(*TopOperator)
	assertQueryGrammarSharedParameters(t, top.Parameters(), []queryGrammarSharedParameterExpectation{
		{name: "hint.progressive_top", kind: SyntaxKindBooleanLiteralExpression},
		{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
	})
}

func TestQueryGrammarPublicSharedParametersEqualsNeeded(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	// Distinct and top require '=' during parameter lookahead. Without it, the
	// parameter-like token is left for the operand parser and remains unconsumed.
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "distinct hint.strategy shuffle", 1)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "top hint.progressive_top true 5 by Score", 1)

	// Sort uses equalsNeeded=false, so it recognizes the name and recovers a
	// missing '=' while retaining the following value.
	sort := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"sort hint.strategy splitBlock by Score").(*SortOperator)
	parameter := sort.Parameters().At(0)
	if parameter.EqualToken().Width() != 0 ||
		!diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
		t.Fatalf("sort missing equal token/diagnostics = %#v/%#v", parameter.EqualToken(), parameter.EqualToken().Diagnostics())
	}
	if parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("sort recovered parameter value kind = %v, want token literal", parameter.Expression().Kind())
	}
}

func TestQueryGrammarPublicSharedParametersMissingValues(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name        string
		text        string
		parameterAt func(QueryOperator) *NamedParameter
		kind        SyntaxKind
		code        string
		message     string
	}{
		{
			name: "distinct integer",
			text: "distinct hint.num_partitions=",
			parameterAt: func(operator QueryOperator) *NamedParameter {
				return operator.(*DistinctOperator).Parameters().At(0)
			},
			kind: SyntaxKindLongLiteralExpression, code: "KS006", message: "Missing number",
		},
		{
			name: "sort word",
			text: "sort hint.strategy=",
			parameterAt: func(operator QueryOperator) *NamedParameter {
				return operator.(*SortOperator).Parameters().At(0)
			},
			kind: SyntaxKindTokenLiteralExpression, code: "KS005", message: "Expected one of: splitBlock, multipleBlocks",
		},
		{
			name: "top boolean",
			text: "top hint.progressive_top=",
			parameterAt: func(operator QueryOperator) *NamedParameter {
				return operator.(*TopOperator).Parameters().At(0)
			},
			kind: SyntaxKindBooleanLiteralExpression, code: "KS006", message: "Missing boolean",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			value := test.parameterAt(operator).Expression()
			if value.Kind() != test.kind || value.Width() != 0 ||
				!diagnosticsContainMessage(value.ContainedDiagnostics(), test.code, test.message) {
				t.Fatalf("missing parameter value = %#v with diagnostics %#v", value, value.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicSharedParametersInvalidDeclaredValuesFallBack(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name      string
		text      string
		parameter func(QueryOperator) *NamedParameter
	}{
		{
			name: "distinct non-integer",
			text: "distinct hint.num_partitions=many Value",
			parameter: func(operator QueryOperator) *NamedParameter {
				return operator.(*DistinctOperator).Parameters().At(0)
			},
		},
		{
			name: "sort unknown strategy",
			text: "sort hint.strategy=bogus by Score",
			parameter: func(operator QueryOperator) *NamedParameter {
				return operator.(*SortOperator).Parameters().At(0)
			},
		},
		{
			name: "top non-boolean",
			text: "top hint.progressive_top=perhaps 5 by Score",
			parameter: func(operator QueryOperator) *NamedParameter {
				return operator.(*TopOperator).Parameters().At(0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameter := test.parameter(parseQueryGrammarProduction(t, grammar.QueryOperator, test.text))
			if parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
				t.Fatalf("fallback value kind = %v, want token literal", parameter.Expression().Kind())
			}
			if diagnostics := parameter.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("fallback parameter diagnostics = %#v, want none", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicSharedParametersRepeatParsing(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name       string
		text       string
		parameters func(QueryOperator) *SyntaxListOf[*NamedParameter]
	}{
		{
			name: "distinct",
			text: "distinct hint.shufflekey=Tenant hint.shufflekey=Account Value",
			parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
				return operator.(*DistinctOperator).Parameters()
			},
		},
		{
			name: "sort",
			text: "sort hint.strategy=splitBlock hint.strategy=multipleBlocks by Score",
			parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
				return operator.(*SortOperator).Parameters()
			},
		},
		{
			name: "top",
			text: "top hint.progressive_top=true hint.progressive_top=false 5 by Score",
			parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
				return operator.(*TopOperator).Parameters()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameters := test.parameters(parseQueryGrammarProduction(t, grammar.QueryOperator, test.text))
			if parameters.Count() != 2 {
				t.Fatalf("repeated parameters = %d, want 2", parameters.Count())
			}
			if NameDeclarationSimpleName(parameters.At(0).Name()) != NameDeclarationSimpleName(parameters.At(1).Name()) {
				t.Fatalf("repeated parameter names = %q/%q", NameDeclarationSimpleName(parameters.At(0).Name()), NameDeclarationSimpleName(parameters.At(1).Name()))
			}
		})
	}
}

func TestQueryGrammarPublicSharedParametersUnsupportedTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "distinct Value | where X", 2)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "sort by Score | where X", 3)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "top 5 by Score | where X", 4)
	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "distinct unknown=1 | where X", 4)
}

type queryGrammarSharedParameterExpectation struct {
	name string
	kind SyntaxKind
}

func assertQueryGrammarSharedParameters(t *testing.T, parameters *SyntaxListOf[*NamedParameter], want []queryGrammarSharedParameterExpectation) {
	t.Helper()
	if parameters.Count() != len(want) {
		t.Fatalf("parameters = %d, want %d", parameters.Count(), len(want))
	}
	for i, expectation := range want {
		parameter := parameters.At(i)
		if name := NameDeclarationSimpleName(parameter.Name()); name != expectation.name {
			t.Fatalf("parameter %d name = %q, want %q", i, name, expectation.name)
		}
		if parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 {
			t.Fatalf("parameter %d equal token = %#v, want present equal", i, parameter.EqualToken())
		}
		if kind := parameter.Expression().Kind(); kind != expectation.kind {
			t.Fatalf("parameter %d value kind = %v, want %v", i, kind, expectation.kind)
		}
	}
}
