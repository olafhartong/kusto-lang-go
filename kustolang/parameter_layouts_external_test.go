package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

type oddCountLayout struct {
	parameter *kustolang.Parameter
}

func (l oddCountLayout) Kind() kustolang.ParameterLayout { return kustolang.ParameterLayoutCustom }

func (l oddCountLayout) GetArgumentParameters(_ *kustolang.Signature, arguments []kustolang.Expression, _ ...*kustolang.GlobalState) []*kustolang.Parameter {
	result := make([]*kustolang.Parameter, len(arguments))
	for i := range result {
		result[i] = l.parameter
	}
	return result
}

func (l oddCountLayout) GetArgumentParametersForTypes(_ *kustolang.Signature, argumentTypes []*kustolang.TypeSymbol) []*kustolang.Parameter {
	result := make([]*kustolang.Parameter, len(argumentTypes))
	for i := range result {
		result[i] = l.parameter
	}
	return result
}

func (l oddCountLayout) GetNextPossibleParameters(_ *kustolang.Signature, _ []kustolang.Expression) []*kustolang.Parameter {
	return []*kustolang.Parameter{l.parameter}
}

func (oddCountLayout) IsValidArgumentCount(_ *kustolang.Signature, argumentCount int) bool {
	return argumentCount >= 1 && argumentCount%2 == 1
}

func TestPublicParameterLayoutsCatalogAndOperations(t *testing.T) {
	if kustolang.ParameterLayouts.Fixed.Kind() != kustolang.ParameterLayoutFixed ||
		kustolang.ParameterLayouts.Repeating.Kind() != kustolang.ParameterLayoutRepeating ||
		kustolang.ParameterLayouts.RepeatingSkipping.Kind() != kustolang.ParameterLayoutRepeatingSkip ||
		kustolang.ParameterLayouts.BlockRepeating.Kind() != kustolang.ParameterLayoutBlockRepeating {
		t.Fatal("public ParameterLayouts catalog does not map the upstream singletons")
	}

	optionalText := kustolang.NewParameter("text", kustolang.ScalarTypes.String.Type(), kustolang.WithParameterOccurrence(0, 1))
	limit := kustolang.NewParameter("limit", kustolang.ScalarTypes.Long.Type())
	base := kustolang.NewSignature(kustolang.ScalarTypes.Bool.Type(), optionalText, limit)
	argument := kustolang.CreateFakeExpression(kustolang.ScalarTypes.Long.Type())

	parameters := kustolang.ParameterLayouts.RepeatingSkipping.GetArgumentParameters(base, []kustolang.Expression{argument})
	if len(parameters) != 1 || parameters[0] != limit {
		t.Fatalf("layout argument parameters = %#v", parameters)
	}
	if base.Layout == kustolang.ParameterLayoutRepeatingSkip {
		t.Fatal("layout operation mutated the input signature")
	}
}

func TestPublicCustomParameterLayoutDefinitionBindsValidAndMalformedCalls(t *testing.T) {
	textParameter := kustolang.NewParameter("text", kustolang.ScalarTypes.String.Type())
	numberParameter := kustolang.NewParameter("number", kustolang.ScalarTypes.Long.Type())
	var calls int
	custom := kustolang.ParameterLayouts.Custom(func(_ *kustolang.Signature, arguments []kustolang.Expression, _ []kustolang.Symbol) []*kustolang.Parameter {
		calls++
		result := make([]*kustolang.Parameter, len(arguments))
		if len(result) > 0 {
			result[0] = numberParameter
		}
		if len(result) > 1 {
			result[1] = textParameter
		}
		return result
	})
	signature := kustolang.NewSignature(kustolang.ScalarTypes.Long.Type(), textParameter, numberParameter).
		WithParameterLayout(custom)
	function := kustolang.NewFunctionSymbol("layout_object", signature)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", function))

	valid := kustolang.ParseAndAnalyze(`print value=layout_object(1, "ok")`, globals)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}
	malformed := kustolang.ParseAndAnalyze(`print value=layout_object(1,`, globals)
	if diagnostics := malformed.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS005") {
		t.Fatalf("malformed diagnostics = %#v", diagnostics)
	}
	if signature.Layout != kustolang.ParameterLayoutCustom || calls < 2 {
		t.Fatalf("custom layout metadata/calls = %s/%d", signature.Layout, calls)
	}
}

func TestPublicParameterLayoutStrategyControlsAllOperations(t *testing.T) {
	value := kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type())
	base := kustolang.NewSignature(kustolang.ScalarTypes.Bool.Type(), value)
	signature := base.WithParameterLayout(oddCountLayout{parameter: value})

	if signature == base || !signature.IsValidArgumentCount(3) || signature.IsValidArgumentCount(2) {
		t.Fatalf("strategy count validation = one:%t two:%t three:%t", signature.IsValidArgumentCount(1), signature.IsValidArgumentCount(2), signature.IsValidArgumentCount(3))
	}
	arguments := []kustolang.Expression{
		kustolang.CreateFakeExpression(kustolang.ScalarTypes.String.Type()),
		kustolang.CreateFakeExpression(kustolang.ScalarTypes.Long.Type()),
	}
	if got := signature.GetArgumentParameters(arguments); len(got) != 2 || got[0] != value || got[1] != value {
		t.Fatalf("strategy argument mapping = %#v", got)
	}
	if got := signature.GetNextPossibleParameters(arguments); len(got) != 1 || got[0] != value {
		t.Fatalf("strategy next parameters = %#v", got)
	}
	if _, ok := signature.LayoutStrategy().(oddCountLayout); !ok || !signature.HasCustomParameterLayout() {
		t.Fatalf("strategy identity/custom metadata = %T/%t", signature.LayoutStrategy(), signature.HasCustomParameterLayout())
	}
}
