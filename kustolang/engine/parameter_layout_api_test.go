package kustolang

import "testing"

func TestSourceShapedSignatureLayoutMethodsInferExpressionTypes(t *testing.T) {
	optionalText := NewParameter("text", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1))
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), optionalText, limit).
		WithParameterLayout(ParameterLayouts.RepeatingSkipping)
	arguments := []Expression{FakeExpressionCreate(ScalarTypes.Long.Type())}

	parameters := signature.GetArgumentParameters(arguments)
	if len(parameters) != 1 || parameters[0] != limit {
		t.Fatalf("inferred argument parameters = %#v", parameters)
	}
	next := signature.GetNextPossibleParameters(nil)
	if len(next) != 2 || next[0] != optionalText || next[1] != limit {
		t.Fatalf("next possible parameters = %#v", next)
	}
	if got := signature.GetArgumentParametersForTypes([]*TypeSymbol{ScalarTypes.Long.Type()}); len(got) != 1 || got[0] != limit {
		t.Fatalf("typed argument parameters = %#v", got)
	}
}

func TestParameterLayoutDefinitionAppliesWithoutMutatingSignature(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type(), WithParameterOccurrence(1, 2))
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	base := NewSignature(ScalarTypes.Bool.Type(), value, limit).WithLayout(ParameterLayoutFixed)

	applied := ParameterLayouts.Repeating.Apply(base)
	if base.Layout != ParameterLayoutFixed || applied == base || applied.Layout != ParameterLayoutRepeating {
		t.Fatalf("layout application mutated source: base=%p/%s applied=%p/%s", base, base.Layout, applied, applied.Layout)
	}
	if !ParameterLayouts.Repeating.IsValidArgumentCount(base, 2) {
		t.Fatal("repeating definition rejected valid argument count")
	}
}
