package kustolang

import (
	"strings"
	"testing"
)

func TestAvoidUsingObsoleteFunctionsAnalyzerReportsFunctionObsoleteAndFixesRename(t *testing.T) {
	text := "T | extend Value=parsejson(Payload)"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ParseJsonDeprecated, ScalarTypes.Dynamic.Type()))

	analyzer := NewAvoidUsingObsoleteFunctionsAnalyzer()
	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingObsoleteFunctions ||
		diagnostic.Category != DiagnosticCategoryCorrectness ||
		diagnostic.Severity != DiagnosticSeverityWarning ||
		diagnostic.Start != strings.Index(text, "parsejson") ||
		diagnostic.Length != len("parsejson") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	wantMessage := "The function 'parsejson' is deprecated; use 'parse_json' instead."
	if diagnostic.Message != wantMessage {
		t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
	}

	actions := analyzer.Fixer().GetFixActions(code, diagnostic, []Diagnostic{diagnostic}, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}
	if actions[0].Base().Kind != "Obsolete" ||
		actions[0].Base().Title != "Change to 'parse_json'" ||
		actions[0].Base().Description != "Convert from using deprecated function 'parsejson' to preferred function 'parse_json'" {
		t.Fatalf("action metadata = %#v", actions[0].Base())
	}
	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("fix result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok {
		t.Fatalf("first result action = %#v, want text change", result.Actions[0])
	}
	if change.ChangedText != "T | extend Value=parse_json(Payload)" {
		t.Fatalf("changed text = %q", change.ChangedText)
	}
}

func TestAvoidUsingObsoleteFunctionsAnalyzerReportsSignatureObsoleteAndFixesCompatibleRename(t *testing.T) {
	obsoleteSignature := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.String.Type())).Obsolete("preferred_form")
	obsoleteFunction := NewFunctionSymbol("legacy_form", obsoleteSignature)
	alternativeFunction := NewScalarFunctionSymbol("preferred_form", ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db")).WithFunctions([]*FunctionSymbol{obsoleteFunction, alternativeFunction})
	text := "T | extend Value=legacy_form(Account)"
	code := ParseWithGlobals(text, globals)
	call := firstParsedFunctionCall(t, code)
	call.SetSemanticInfo(NewSemanticInfoForSignature(obsoleteSignature, ScalarTypes.String.Type()))

	analyzer := NewAvoidUsingObsoleteFunctionsAnalyzer()
	diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Start != strings.Index(text, "legacy_form") ||
		diagnostic.Length != len("legacy_form") {
		t.Fatalf("diagnostic location = %#v", diagnostic)
	}
	wantMessage := "This form of the function 'legacy_form' is deprecated; use 'preferred_form' instead."
	if diagnostic.Message != wantMessage {
		t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
	}

	actions := analyzer.Fixer().GetFixActions(code, diagnostic, []Diagnostic{diagnostic}, CodeActionOptionsDefault, nil, CancellationToken{})
	if len(actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions[0])
	}
	result := analyzer.Fixer().ApplyFixAction(code, action, diagnostic.Start, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("fix result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "T | extend Value=preferred_form(Account)" {
		t.Fatalf("change = %#v", result.Actions[0])
	}
}

func TestAvoidUsingObsoleteFunctionsAnalyzerSkipsUnboundAndNonObsoleteFunctions(t *testing.T) {
	unbound := Parse("T | extend Value=parsejson(Payload)")
	if diagnostics := NewAvoidUsingObsoleteFunctionsAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	bound := Parse("T | extend Value=parse_json(Payload)")
	call := firstParsedFunctionCall(t, bound)
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ParseJson, ScalarTypes.Dynamic.Type()))
	if diagnostics := NewAvoidUsingObsoleteFunctionsAnalyzer().Analyze(bound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("non-obsolete diagnostics = %#v, want empty", diagnostics)
	}
}

func TestAvoidUsingObsoleteFunctionsAnalyzerSuppressesUnsafeFixActions(t *testing.T) {
	t.Run("incompatible alternative", func(t *testing.T) {
		obsoleteSignature := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.String.Type())).Obsolete("replacement")
		obsoleteFunction := NewFunctionSymbol("legacy_text", obsoleteSignature)
		alternativeFunction := NewScalarFunctionSymbol("replacement", ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.Long.Type()))
		globals := NewGlobalState(NewDatabaseSymbol("db")).WithFunctions([]*FunctionSymbol{obsoleteFunction, alternativeFunction})
		code := ParseWithGlobals("T | extend Value=legacy_text(Account)", globals)
		call := firstParsedFunctionCall(t, code)
		call.SetSemanticInfo(NewSemanticInfoForSignature(obsoleteSignature, ScalarTypes.String.Type()))

		analyzer := NewAvoidUsingObsoleteFunctionsAnalyzer()
		diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
		if len(diagnostics) != 1 {
			t.Fatalf("diagnostics = %#v, want one", diagnostics)
		}
		actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
		if len(actions) != 0 {
			t.Fatalf("actions = %#v, want none", actions)
		}
	})

	t.Run("row_rank special case", func(t *testing.T) {
		code := Parse("T | extend Rank=row_rank()")
		call := firstParsedFunctionCall(t, code)
		call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.RowRank, ScalarTypes.Long.Type()))

		analyzer := NewAvoidUsingObsoleteFunctionsAnalyzer()
		diagnostics := analyzer.Analyze(code, nil, CancellationToken{})
		if len(diagnostics) != 1 {
			t.Fatalf("diagnostics = %#v, want one", diagnostics)
		}
		actions := analyzer.Fixer().GetFixActions(code, diagnostics[0], diagnostics, CodeActionOptionsDefault, nil, CancellationToken{})
		if len(actions) != 0 {
			t.Fatalf("actions = %#v, want none", actions)
		}
	})
}

func TestAvoidUsingObsoleteFunctionsAnalyzerUsesNameSemanticInfo(t *testing.T) {
	text := "T | extend Value=custom(Payload)"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	call.Name().SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ParseJsonDeprecated, ScalarTypes.Dynamic.Type()))

	diagnostics := NewAvoidUsingObsoleteFunctionsAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message, "'parse_json'") {
		t.Fatalf("diagnostics = %#v, want parse_json warning", diagnostics)
	}
}

func TestAvoidUsingObsoleteFunctionsAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewAvoidUsingObsoleteFunctionsAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeAvoidUsingObsoleteFunctions ||
		diagnostics[0].Category != DiagnosticCategoryCorrectness ||
		diagnostics[0].Severity != DiagnosticSeverityWarning ||
		diagnostics[0].Description != "Avoid using obsolete/deprecated functions." ||
		diagnostics[0].Message != "Avoid using obsolete/deprecated functions." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}
