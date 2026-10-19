package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCustomParameterLayoutValidAndMalformedBinding(t *testing.T) {
	textParameter := kustolang.NewParameter("text", kustolang.ScalarTypes.String.Type())
	numberParameter := kustolang.NewParameter("number", kustolang.ScalarTypes.Long.Type())
	var calls int
	layout := kustolang.ParameterLayoutBuilder(func(_ *kustolang.Signature, arguments []kustolang.Expression, _ []kustolang.Symbol) []*kustolang.Parameter {
		calls++
		parameters := make([]*kustolang.Parameter, len(arguments))
		if len(parameters) > 0 {
			parameters[0] = numberParameter
		}
		if len(parameters) > 1 {
			parameters[1] = textParameter
		}
		return parameters
	})
	signature := kustolang.NewSignature(kustolang.ScalarTypes.Long.Type(), textParameter, numberParameter).
		WithCustomLayout(layout)
	function := kustolang.NewFunctionSymbol("custom_order", signature)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", function))

	valid := kustolang.ParseAndAnalyze(`print value=custom_order(1, "ok")`, globals)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid custom-layout diagnostics = %#v", diagnostics)
	}
	if signature.Layout != kustolang.ParameterLayoutCustom || signature.CustomLayoutBuilder == nil {
		t.Fatalf("custom layout metadata = %v/%t", signature.Layout, signature.CustomLayoutBuilder != nil)
	}

	malformed := kustolang.ParseAndAnalyze(`print value=custom_order(1,`, globals)
	if diagnostics := malformed.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS005") {
		t.Fatalf("malformed custom-layout diagnostics = %#v", diagnostics)
	}
	if calls < 2 {
		t.Fatalf("custom layout builder calls = %d, want valid and malformed binding", calls)
	}
}

func TestPublicParameterLayoutConstantsMapUpstreamSingletons(t *testing.T) {
	layouts := []kustolang.ParameterLayout{
		kustolang.ParameterLayoutFixed,
		kustolang.ParameterLayoutRepeating,
		kustolang.ParameterLayoutRepeatingSkip,
		kustolang.ParameterLayoutBlockRepeating,
		kustolang.ParameterLayoutCustom,
	}
	seen := map[kustolang.ParameterLayout]bool{}
	for _, layout := range layouts {
		if layout == "" || seen[layout] {
			t.Fatalf("invalid or duplicate parameter layout %q", layout)
		}
		seen[layout] = true
	}
}
