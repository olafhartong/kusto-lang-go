package kustolang

import "testing"

func TestQueryGrammarPublicRemainingOperatorsAcceptGlobalParameters(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name       string
		text       string
		parameters func(QueryOperator) *SyntaxListOf[*NamedParameter]
	}{
		{name: "where", text: "where hint.strategy=shuffle Enabled", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "filter", text: "filter hint.strategy=shuffle Enabled", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "serialize", text: "serialize hint.strategy=shuffle Copy=Value", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SerializeOperator).Parameters()
		}},
		{name: "sample", text: "sample hint.strategy=shuffle 10", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleOperator).Parameters()
		}},
		{name: "sample-distinct", text: "sample-distinct hint.strategy=shuffle 3 of User", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleDistinctOperator).Parameters()
		}},
		{name: "take", text: "take hint.strategy=shuffle 5", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
		{name: "limit", text: "limit hint.strategy=shuffle 6", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			assertQueryGrammarSharedParameters(t, test.parameters(operator), []queryGrammarSharedParameterExpectation{
				{name: "hint.strategy", kind: SyntaxKindTokenLiteralExpression},
			})
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("operator diagnostics = %#v, want none", diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicGetSchemaDeclaredParameter(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "getschema kind=csl").(*GetSchemaOperator)
	parameter := operator.KindParameter()
	if parameter == nil {
		t.Fatal("getschema kind parameter = nil")
	}
	assertQueryGrammarRemainingParameter(t, parameter, "kind", SyntaxKindTokenLiteralExpression)
	if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("getschema diagnostics = %#v, want none", diagnostics)
	}
}

func TestQueryGrammarPublicRemainingOperatorsEqualsNeeded(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name       string
		text       string
		parameters func(QueryOperator) *SyntaxListOf[*NamedParameter]
		wantLength int
	}{
		{name: "where", text: "where hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "filter", text: "filter hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "serialize", text: "serialize hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SerializeOperator).Parameters()
		}},
		{name: "sample", text: "sample hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleOperator).Parameters()
		}},
		{name: "sample-distinct", text: "sample-distinct hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleDistinctOperator).Parameters()
		}},
		{name: "take", text: "take hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
		{name: "limit", text: "limit hint.strategy shuffle", wantLength: 1, parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := Lex(test.text)
			result := grammar.QueryOperator.ParseValue(NewArraySource(tokens), 0)
			if result.Length != test.wantLength || result.Length == len(tokens)-1 {
				t.Fatalf("operator prefix parse = %#v, want length %d with a tail", result, test.wantLength)
			}
			if count := test.parameters(result.Value).Count(); count != 0 {
				t.Fatalf("parameters = %d, want 0 without '='", count)
			}
		})
	}

	getSchema := parseQueryGrammarProduction(t, grammar.QueryOperator, "getschema kind csl").(*GetSchemaOperator)
	parameter := getSchema.KindParameter()
	if parameter == nil || parameter.EqualToken().Width() != 0 ||
		!diagnosticsContainMessage(parameter.EqualToken().Diagnostics(), "KS005", "Expected: =") {
		t.Fatalf("getschema missing equal recovery = %#v", parameter)
	}
	if parameter.Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("getschema recovered value kind = %v, want token literal", parameter.Expression().Kind())
	}
}

func TestQueryGrammarPublicRemainingOperatorsMissingParameterValues(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name      string
		text      string
		parameter func(QueryOperator) *NamedParameter
	}{
		{name: "where", text: "where hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*FilterOperator).Parameters().At(0)
		}},
		{name: "filter", text: "filter hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*FilterOperator).Parameters().At(0)
		}},
		{name: "serialize", text: "serialize hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*SerializeOperator).Parameters().At(0)
		}},
		{name: "sample", text: "sample hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*SampleOperator).Parameters().At(0)
		}},
		{name: "sample-distinct", text: "sample-distinct hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*SampleDistinctOperator).Parameters().At(0)
		}},
		{name: "take", text: "take hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*TakeOperator).Parameters().At(0)
		}},
		{name: "limit", text: "limit hint.strategy=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*TakeOperator).Parameters().At(0)
		}},
		{name: "getschema", text: "getschema kind=", parameter: func(operator QueryOperator) *NamedParameter {
			return operator.(*GetSchemaOperator).KindParameter()
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameter := test.parameter(parseQueryGrammarProduction(t, grammar.QueryOperator, test.text))
			value := parameter.Expression()
			if value.Kind() != SyntaxKindTokenLiteralExpression || value.Width() != 0 ||
				!diagnosticsContain(value.ContainedDiagnostics(), "KS005") {
				t.Fatalf("missing parameter value = %#v with diagnostics %#v", value, value.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicRemainingOperatorsParameterFreeControls(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name       string
		text       string
		parameters func(QueryOperator) *SyntaxListOf[*NamedParameter]
	}{
		{name: "where", text: "where Enabled", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "filter", text: "filter Enabled", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*FilterOperator).Parameters()
		}},
		{name: "serialize", text: "serialize Copy=Value", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SerializeOperator).Parameters()
		}},
		{name: "sample", text: "sample 10", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleOperator).Parameters()
		}},
		{name: "sample-distinct", text: "sample-distinct 3 of User", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*SampleDistinctOperator).Parameters()
		}},
		{name: "take", text: "take 5", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
		{name: "limit", text: "limit 6", parameters: func(operator QueryOperator) *SyntaxListOf[*NamedParameter] {
			return operator.(*TakeOperator).Parameters()
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if count := test.parameters(operator).Count(); count != 0 {
				t.Fatalf("parameter-free operator parameters = %d, want 0", count)
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("parameter-free operator diagnostics = %#v", diagnostics)
			}
		})
	}

	getSchema := parseQueryGrammarProduction(t, grammar.QueryOperator, "getschema").(*GetSchemaOperator)
	if getSchema.KindParameter() != nil || len(getSchema.ContainedDiagnostics()) != 0 {
		t.Fatalf("parameter-free getschema = %#v diagnostics:%#v", getSchema.KindParameter(), getSchema.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicRemainingOperatorsLeaveTails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		text       string
		wantLength int
	}{
		{text: "where Enabled | count", wantLength: 2},
		{text: "filter Enabled | count", wantLength: 2},
		{text: "serialize Copy=Value | count", wantLength: 4},
		{text: "sample 10 | count", wantLength: 2},
		{text: "sample-distinct 3 of User | count", wantLength: 4},
		{text: "take 5 | count", wantLength: 2},
		{text: "limit 6 | count", wantLength: 2},
		{text: "getschema kind=csl | count", wantLength: 4},
	} {
		t.Run(test.text, func(t *testing.T) {
			assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, test.text, test.wantLength)
		})
	}
}

func TestQueryGrammarPublicFiltersPracticalGraphCompatibilityGlobals(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tokens := Lex("with_node_id=Node")

	global := queryGrammarParameterList(grammar, nil, true).ParseValue(NewArraySource(tokens), 0)
	if global.Length != 0 || global.Value.Count() != 0 {
		t.Fatalf("pinned global parameter list parsed graph-only name: %#v", global)
	}

	graphSpecific := queryGrammarParameterList(grammar, QueryOperatorParameters.GraphToTableNodesParameters, true).
		ParseValue(NewArraySource(tokens), 0)
	if graphSpecific.Length != len(tokens)-1 || graphSpecific.Value.Count() != 1 ||
		NameDeclarationSimpleName(graphSpecific.Value.At(0).Name()) != "with_node_id" {
		t.Fatalf("graph-specific parameter list did not parse node id: %#v", graphSpecific)
	}
}

func assertQueryGrammarRemainingParameter(t *testing.T, parameter *NamedParameter, name string, valueKind SyntaxKind) {
	t.Helper()
	if got := NameDeclarationSimpleName(parameter.Name()); got != name {
		t.Fatalf("parameter name = %q, want %q", got, name)
	}
	if parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 {
		t.Fatalf("parameter equal token = %#v, want present equal", parameter.EqualToken())
	}
	if got := parameter.Expression().Kind(); got != valueKind {
		t.Fatalf("parameter value kind = %v, want %v", got, valueKind)
	}
}
