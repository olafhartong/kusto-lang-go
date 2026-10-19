package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidStrlenWithDynamicAnalyzerReportsDynamicArrayOrBag(t *testing.T) {
	tests := []struct {
		name string
		typ  *TypeSymbol
	}{
		{name: "dynamic array", typ: ScalarTypes.DynamicArray.Type()},
		{name: "dynamic bag", typ: ScalarTypes.DynamicBag.Type()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "T | extend L=strlen(Payload)"
			code := Parse(text)
			call := firstParsedFunctionCall(t, code)
			call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Strlen, ScalarTypes.Long.Type()))
			setExpressionSemanticType(t, functionCallArgumentAt(call, 0), tt.typ)

			diagnostics := NewAvoidStrlenWithDynamicAnalyzer().Analyze(code, nil, CancellationToken{})
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %#v, want one", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Code != KustoAnalyzerCodeAvoidStrlenWithDynamic ||
				diagnostic.Category != DiagnosticCategoryPerformance ||
				diagnostic.Severity != DiagnosticSeveritySuggestion ||
				diagnostic.Start != strings.Index(text, "strlen") ||
				diagnostic.Length != len("strlen(Payload)") {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestAvoidStrlenWithDynamicAnalyzerSkipsUnboundNonDynamicAndWrongArity(t *testing.T) {
	unbound := Parse("T | extend L=strlen(Payload)")
	if diagnostics := NewAvoidStrlenWithDynamicAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	stringArg := Parse("T | extend L=strlen(Payload)")
	stringCall := firstParsedFunctionCall(t, stringArg)
	stringCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Strlen, ScalarTypes.Long.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(stringCall, 0), ScalarTypes.String.Type())
	if diagnostics := NewAvoidStrlenWithDynamicAnalyzer().Analyze(stringArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("string diagnostics = %#v, want empty", diagnostics)
	}

	wrongArity := Parse("T | extend L=strlen(Payload, Extra)")
	wrongArityCall := firstParsedFunctionCall(t, wrongArity)
	wrongArityCall.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Strlen, ScalarTypes.Long.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(wrongArityCall, 0), ScalarTypes.DynamicArray.Type())
	if diagnostics := NewAvoidStrlenWithDynamicAnalyzer().Analyze(wrongArity, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("wrong-arity diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidStrlenWithDynamicAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidStrlenWithDynamicAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidStrlenWithDynamic ||
		diagnostics[0].Description != "Avoid using the strlen() with dynamic arguments." ||
		diagnostics[0].Message != "Avoid using strlen() with dynamic arguments that may be large bags or arrays." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}
