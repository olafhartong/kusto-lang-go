package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicFunctionSymbolAuthoringConstructors(t *testing.T) {
	value := kustolang.NewParameter("value", kustolang.ScalarTypes.Long.Type())

	returnKind := kustolang.NewReturnKindFunctionSymbol("identity", kustolang.ReturnTypeParameter0, value)
	returnKindSignature := singleOwnedSignature(t, returnKind)
	if returnKindSignature.ReturnKind != kustolang.ReturnTypeParameter0 || returnKindSignature.Parameters[0] != value {
		t.Fatalf("return-kind signature = %#v", returnKindSignature)
	}

	custom := kustolang.NewCustomFunctionSymbol(
		"custom",
		func(kustolang.CustomReturnTypeContext) kustolang.Symbol { return kustolang.ScalarTypes.String.Type() },
		kustolang.TabularityScalar,
		value,
	)
	customSignature := singleOwnedSignature(t, custom)
	if customSignature.ReturnKind != kustolang.ReturnTypeCustom || customSignature.CustomReturnType == nil || customSignature.Tabularity != kustolang.TabularityScalar {
		t.Fatalf("custom-return signature = %#v", customSignature)
	}

	inferred := kustolang.NewFunctionSymbolFromParameterList("increment", "(value:long)", "value + 1", "Adds one")
	inferredSignature := singleOwnedSignature(t, inferred)
	if inferred.Description() != "Adds one" || len(inferredSignature.Parameters) != 1 || inferredSignature.Parameters[0].Name != "value" {
		t.Fatalf("inferred function = description:%q signature:%#v", inferred.Description(), inferredSignature)
	}
	inferredSignature.ComputeTabularity(kustolang.DefaultGlobalState)
	if inferredSignature.Tabularity != kustolang.TabularityScalar {
		t.Fatalf("inferred tabularity = %s", inferredSignature.Tabularity)
	}

	explicit := kustolang.NewFunctionSymbolFromParameterListWithTabularity(
		"rows",
		"(take:long)",
		"range value from 1 to take step 1",
		kustolang.TabularityTabular,
		"Produces rows",
	)
	explicitSignature := singleOwnedSignature(t, explicit)
	if explicit.Description() != "Produces rows" || explicitSignature.Tabularity != kustolang.TabularityTabular || len(explicitSignature.Parameters) != 1 {
		t.Fatalf("explicit function = description:%q signature:%#v", explicit.Description(), explicitSignature)
	}
}

func TestPublicDeclarationFunctionSymbolBindsThroughGlobalStateFunctions(t *testing.T) {
	declaration := kustolang.ParseFunctionBody("{ value + 1 }")
	parameter := kustolang.NewParameter("value", kustolang.ScalarTypes.Long.Type())
	function := kustolang.NewDeclarationFunctionSymbol("declared_increment", declaration, parameter)
	signature := singleOwnedSignature(t, function)
	if signature.Declaration != declaration {
		t.Fatalf("signature declaration = %#v, want %#v", signature.Declaration, declaration)
	}

	globals := kustolang.DefaultGlobalState.WithFunctions([]*kustolang.FunctionSymbol{function})
	code := kustolang.ParseAndAnalyze("print result=declared_increment(41)", globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("declaration call diagnostics = %#v", diagnostics)
	}
	calls := kustolang.GetDescendants[*kustolang.FunctionCallExpression](code.Syntax())
	if len(calls) != 1 || calls[0].ReferencedSignature() != signature || calls[0].ReferencedSymbol() != function {
		t.Fatalf("bound declaration call = %#v", calls)
	}
	if result := kustolang.ExpressionResultType(calls[0]); result != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("declaration result type = %#v", result)
	}
}

func singleOwnedSignature(t *testing.T, function *kustolang.FunctionSymbol) *kustolang.Signature {
	t.Helper()
	signatures := function.Signatures()
	if len(signatures) != 1 || signatures[0] == nil {
		t.Fatalf("function signatures = %#v", signatures)
	}
	if signatures[0].Symbol != function {
		t.Fatalf("signature owner = %#v, want %#v", signatures[0].Symbol, function)
	}
	return signatures[0]
}
