package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingToBoolOnNumericsAnalyzerReportsNumericArguments(t *testing.T) {
	tests := []struct {
		name     string
		function *FunctionSymbol
		text     string
		typ      *TypeSymbol
	}{
		{name: "tobool datetime", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.DateTime.Type()},
		{name: "tobool int", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.Int.Type()},
		{name: "tobool decimal", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.Decimal.Type()},
		{name: "tobool guid", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.Guid.Type()},
		{name: "tobool long", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.Long.Type()},
		{name: "tobool real", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.Real.Type()},
		{name: "tobool timespan", function: Functions.ToBool, text: "T | extend B=tobool(Value)", typ: ScalarTypes.TimeSpan.Type()},
		{name: "toboolean long", function: Functions.ToBoolean, text: "T | extend B=toboolean(Value)", typ: ScalarTypes.Long.Type()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse(tt.text)
			call := firstParsedFunctionCall(t, code)
			call.SetSemanticInfo(NewSemanticInfoForSymbol(tt.function, ScalarTypes.Bool.Type()))
			setExpressionSemanticType(t, functionCallArgumentAt(call, 0), tt.typ)

			diagnostics := NewAvoidUsingToBoolOnNumericsAnalyzer().Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			functionName := tt.function.SymbolName()
			if diagnostic.Code != KustoAnalyzerCodeAvoidUsingToBoolOnNumerics ||
				diagnostic.Category != DiagnosticCategoryCorrectness ||
				diagnostic.Severity != DiagnosticSeverityWarning ||
				diagnostic.Start != strings.Index(tt.text, functionName) ||
				diagnostic.Length != len(functionName+"(Value)") {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestAvoidUsingToBoolOnNumericsAnalyzerSkipsUnboundNonNumericAndNoArguments(t *testing.T) {
	unbound := Parse("T | extend B=tobool(Value)")
	if diagnostics := NewAvoidUsingToBoolOnNumericsAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	stringArg := Parse("T | extend B=tobool(Value)")
	stringCall := firstParsedFunctionCall(t, stringArg)
	stringCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ToBool, ScalarTypes.Bool.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(stringCall, 0), ScalarTypes.String.Type())
	if diagnostics := NewAvoidUsingToBoolOnNumericsAnalyzer().Analyze(stringArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("string diagnostics = %#v, want empty", diagnostics)
	}

	noArg := Parse("T | extend B=tobool()")
	noArgCall := firstParsedFunctionCall(t, noArg)
	noArgCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ToBool, ScalarTypes.Bool.Type()))
	if diagnostics := NewAvoidUsingToBoolOnNumericsAnalyzer().Analyze(noArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("no-arg diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingToBoolOnNumericsAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingToBoolOnNumericsAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingToBoolOnNumerics ||
		diagnostics[0].Description != "Avoid using tobool on numeric arguments" ||
		diagnostics[0].Message != "Avoid using tobool on numeric arguments, use comparison operators instead" {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func setExpressionSemanticType(t *testing.T, expression Expression, typ *TypeSymbol) {
	t.Helper()
	semantic, ok := expression.(interface{ SetSemanticInfo(SemanticInfo) })
	if !ok {
		t.Fatalf("expression %T cannot carry semantic info", expression)
	}
	semantic.SetSemanticInfo(NewSemanticInfoForType(typ))
}
