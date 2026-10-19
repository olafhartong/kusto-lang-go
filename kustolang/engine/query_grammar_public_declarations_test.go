package kustolang

import "testing"

func TestQueryGrammarPublicFunctionParametersEmptyAndFull(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.FunctionParameters, "()")
	if empty.Parameters().Count() != 0 || len(empty.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty parameters = count:%d diagnostics:%#v", empty.Parameters().Count(), empty.ContainedDiagnostics())
	}

	full := parseQueryGrammarProduction(t, grammar.FunctionParameters, "(x:string, count:long=1, mode:string=access, T:(A:string, Nested:(B:long)))")
	if full.Parameters().Count() != 4 || len(full.ContainedDiagnostics()) != 0 {
		t.Fatalf("full parameters = count:%d diagnostics:%#v", full.Parameters().Count(), full.ContainedDiagnostics())
	}
	count := full.Parameters().At(1).Element()
	if count.Kind() != SyntaxKindFunctionParameter || count.DefaultValue() == nil || count.DefaultValue().Kind() != SyntaxKindDefaultValueDeclaration ||
		count.DefaultValue().Value().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("literal default AST = %#v", count)
	}
	mode := full.Parameters().At(2).Element().DefaultValue().Value()
	if mode.Kind() != SyntaxKindTokenLiteralExpression || mode.Text(IncludeTriviaMinimal) != "access" {
		t.Fatalf("name-token default = kind:%v text:%q", mode.Kind(), mode.Text(IncludeTriviaMinimal))
	}
	nested, ok := full.Parameters().At(3).Element().NameAndType().Type().(*SchemaTypeExpression)
	if !ok || nested.Columns().Count() != 2 {
		t.Fatalf("nested schema type = %#v", full.Parameters().At(3).Element().NameAndType().Type())
	}
}

func TestQueryGrammarPublicFunctionParametersRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name    string
		text    string
		message string
	}{
		{name: "missing open", text: "x:string)", message: "Expected: ("},
		{name: "missing close", text: "(x:string", message: "Expected: )"},
		{name: "missing name", text: "(:string)", message: "Missing name"},
		{name: "missing type", text: "(x:)", message: "Missing type name"},
		{name: "missing comma", text: "(x:string y:long)", message: "Expected: ,"},
		{name: "trailing comma", text: "(x:string,)", message: "Missing parameter"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameters := parseQueryGrammarProduction(t, grammar.FunctionParameters, test.text)
			if !diagnosticsContainMessage(parameters.ContainedDiagnostics(), "KS005", test.message) &&
				!diagnosticsContainMessage(parameters.ContainedDiagnostics(), "KS006", test.message) {
				t.Fatalf("diagnostics = %#v, want %q", parameters.ContainedDiagnostics(), test.message)
			}
		})
	}
}

func TestQueryGrammarPublicFunctionParameterAnnotationsAndKeywordDefaultBoundary(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	if !queryGrammarParserTreeHasTag(grammar.FunctionParameters, "<function-parameter>", 0) ||
		!queryGrammarParserTreeHasTag(grammar.FunctionParameters, "<literal>", 0) {
		t.Fatal("function parameter parser graph is missing declaration/literal tags")
	}

	parameter := parseQueryGrammarProduction(t, grammar.FunctionParameters, "(x:long=1)").Parameters().At(0).Element()
	if parameter.CompletionHint(1) != CompletionHintSyntax || parameter.DefaultValue().CompletionHint(1) != CompletionHintScalar {
		t.Fatalf("completion hints = parameter:%v default-value:%v", parameter.CompletionHint(1), parameter.DefaultValue().CompletionHint(1))
	}

	tokens := Lex("(x:string=project)")
	result := grammar.FunctionParameters.ParseValue(NewArraySource(tokens), 0)
	if result.Length < 0 || result.Length == len(tokens)-1 || result.Value == nil ||
		!diagnosticsContain(result.Value.ContainedDiagnostics(), "KS006") {
		t.Fatalf("reserved-keyword default parse = %#v diagnostics:%#v", result, result.Value.ContainedDiagnostics())
	}
}
