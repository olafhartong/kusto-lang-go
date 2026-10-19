package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicParameterParsingAndDeclaration(t *testing.T) {
	parameters := kustolang.ParseParameterList(`(threshold:long=long(5), source:(Account:string, Count:long))`)
	if len(parameters) != 2 {
		t.Fatalf("parameters = %#v", parameters)
	}
	threshold := parameters[0]
	if threshold.Name != "threshold" || !threshold.IsOptional() || threshold.DefaultValue == nil ||
		threshold.DeclaredSymbol(0) != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("threshold = %#v", threshold)
	}
	if got := kustolang.ParameterDeclaration(threshold); got != "threshold: long = long(5)" {
		t.Fatalf("threshold declaration = %q", got)
	}
	if got := kustolang.ParameterDeclaration(parameters[1]); got != "source: (Account: string, Count: long)" {
		t.Fatalf("source declaration = %q", got)
	}
	if got := kustolang.ParameterListDeclaration(parameters); got != "(threshold: long = long(5), source: (Account: string, Count: long))" {
		t.Fatalf("parameter list declaration = %q", got)
	}
}

func TestPublicParameterOptionsAndSignatureKinds(t *testing.T) {
	defaultValue := kustolang.ParseExpression(`"fallback"`)
	parameter := kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type(),
		kustolang.WithArgumentKind(kustolang.ArgumentKindLiteral),
		kustolang.WithParameterValues("one", "two"),
		kustolang.WithParameterExamples(`"one"`),
		kustolang.WithParameterCaseSensitive(true),
		kustolang.WithParameterDefaultValueIndicator("default"),
		kustolang.WithParameterDefaultValue(defaultValue),
		kustolang.WithParameterDescription("A value."),
	)
	if !parameter.IsOptional() || !parameter.IsCaseSensitive || parameter.DefaultValueIndicator != "default" ||
		parameter.Description != "A value." || !reflect.DeepEqual(parameter.Values, []any{"one", "two"}) ||
		!reflect.DeepEqual(parameter.Examples, []string{`"one"`}) {
		t.Fatalf("parameter = %#v", parameter)
	}

	returnKind := kustolang.NewReturnKindSignature(kustolang.ReturnTypeParameter0, parameter)
	called := false
	returnKind = returnKind.WithCustomLayout(func(signature *kustolang.Signature, arguments []kustolang.Expression, _ []kustolang.Symbol) []*kustolang.Parameter {
		called = true
		return []*kustolang.Parameter{signature.Parameters[0]}
	})
	if returnKind.ReturnKind != kustolang.ReturnTypeParameter0 || returnKind.Layout != kustolang.ParameterLayoutCustom || returnKind.CustomLayoutBuilder == nil {
		t.Fatalf("return-kind signature = %#v", returnKind)
	}
	got := returnKind.CustomLayoutBuilder(returnKind, []kustolang.Expression{kustolang.ParseExpression(`"x"`)}, []kustolang.Symbol{kustolang.ScalarTypes.String.Type()})
	if !called || len(got) != 1 || got[0] != parameter {
		t.Fatalf("custom layout = called:%t parameters:%#v", called, got)
	}

	table := kustolang.NewTableSymbol("Input", kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.Long.Type()))
	if signature := kustolang.NewSymbolSignature(table); signature.DeclaredReturnSymbol != table || signature.Tabularity != kustolang.TabularityTabular {
		t.Fatalf("symbol signature = %#v", signature)
	}
	if signature := kustolang.NewComputedSignature("print x=1", kustolang.TabularityTabular); signature.Body != "print x=1" || signature.ReturnKind != kustolang.ReturnTypeComputed {
		t.Fatalf("computed signature = %#v", signature)
	}
	body := kustolang.ParseFunctionBody("{ print x=1 }")
	if signature := kustolang.NewDeclarationSignature(body); signature.Declaration != body || signature.ReturnKind != kustolang.ReturnTypeComputed {
		t.Fatalf("declaration signature = %#v", signature)
	}
}

func TestPublicParameterConversionsPreserveSymbolMetadata(t *testing.T) {
	symbol := kustolang.NewParameterSymbol("source", kustolang.NewTableSymbol("Input", kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.Long.Type())), "Input rows.")
	parameter := kustolang.NewParameterFromSymbol(symbol, true, nil)
	if parameter.Name != "source" || !parameter.IsOptional() || parameter.Description != "Input rows." || parameter.DeclaredSymbol(0) != symbol.Type {
		t.Fatalf("parameter = %#v", parameter)
	}
	if got := kustolang.ParameterDeclaration(parameter); got != "source: (A: long)" {
		t.Fatalf("declaration = %q", got)
	}

	declaration := kustolang.ParseFunctionParameters(`(name:string="default")`)
	converted := kustolang.NewParametersFromFunctionParameters(declaration)
	if len(converted) != 1 || converted[0].Name != "name" || !converted[0].IsOptional() || converted[0].DefaultValue == nil {
		t.Fatalf("converted = %#v", converted)
	}
}
