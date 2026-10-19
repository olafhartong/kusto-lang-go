package kustolang

import (
	"slices"
	"sort"
	"testing"
)

// Ported from KustoCompleter.AnyInfixMatches and its signature/type helpers in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TestKustoCompleterFiltersScalarInfixByTypeAndReturnContext(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Flag", ScalarTypes.Bool.Type()),
		NewColumnSymbol("Count", ScalarTypes.Long.Type()),
		NewColumnSymbol("Text", ScalarTypes.String.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	options := CompletionOptionsDefault.WithIncludeSymbols(false)
	tests := []struct {
		name      string
		text      string
		position  int
		want      []string
		dontWant  []string
		malformed bool
	}{
		{
			name:     "numeric in boolean context",
			text:     "Events | where Count ",
			want:     []string{"==", "<", ">="},
			dontWant: []string{"+", "*"},
		},
		{
			name:     "numeric in scalar context",
			text:     "Events | project Count ",
			want:     []string{"+", "*", "/"},
			dontWant: []string{"==", "<"},
		},
		{
			name:     "string equality",
			text:     "Events | where Text ",
			want:     []string{"==", "!="},
			dontWant: []string{"<", "+"},
		},
		{
			name:     "hidden boolean equality",
			text:     "Events | where Flag ",
			dontWant: []string{"==", "!=", "+"},
		},
		{
			name:      "malformed unexpected close",
			text:      "Events | where Count )",
			position:  len("Events | where Count "),
			want:      []string{"==", "<", ">="},
			dontWant:  []string{"+", "*"},
			malformed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			position := test.position
			if position == 0 {
				position = len(test.text)
			}
			code := ParseAndAnalyze(test.text, globals)
			if malformed := len(code.GetSyntaxDiagnostics()) != 0; malformed != test.malformed {
				t.Fatalf("syntax diagnostics present = %t, want %t: %#v", malformed, test.malformed, code.GetSyntaxDiagnostics())
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(position, options, CancellationToken{}).Items
			for _, display := range test.want {
				item, ok := completionItemByDisplayText(items, display)
				if !ok {
					t.Fatalf("completion items missing %q: %#v", display, completionDisplayTexts(items))
				}
				if item.Kind != CompletionKindScalarInfix {
					t.Fatalf("completion %q kind = %d, want scalar infix", display, item.Kind)
				}
			}
			for _, display := range test.dontWant {
				if _, ok := completionItemByDisplayText(items, display); ok {
					t.Fatalf("completion items unexpectedly include %q: %#v", display, completionDisplayTexts(items))
				}
			}
		})
	}
}

func TestKustoCompleterOrdinarySyntaxDoesNotLeakScalarInfix(t *testing.T) {
	tests := []struct {
		text string
		want []string
	}{
		{text: "print 1 ", want: []string{"%", "*", "+", ",", "-", ".", "/", "[", "|"}},
		{text: "let ", want: []string{".", "[", "|"}},
		{text: "query_parameters ", want: []string{".", ";", "[", "|"}},
		{text: "print \"😀\"; let ", want: []string{".", "[", "|"}},
		{text: "// 😀\nlet ", want: []string{".", "[", "|"}},
		{text: "let x = ; ", want: []string{}},
	}
	options := CompletionOptionsDefault.
		WithIncludeSymbols(false).
		WithIncludeExtendedSyntax(false)

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			items := NewKustoCodeService(test.text).
				GetCompletionItems(utf16TextLength(test.text), options, CancellationToken{}).
				Items
			got := completionDisplayTexts(items)
			sort.Strings(got)
			if !slices.Equal(got, test.want) {
				t.Fatalf("ordinary syntax completions = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestKustoCompleterInfixSignatureVisibilityAndReturnHints(t *testing.T) {
	longType := ScalarTypes.Long.Type()
	if !completionHasMatchingInfixOperator(DefaultGlobalState, OperatorKindAdd, longType, CompletionHintScalar) {
		t.Fatal("numeric addition should match a scalar return context")
	}
	if completionHasMatchingInfixOperator(DefaultGlobalState, OperatorKindAdd, longType, CompletionHintBoolean) {
		t.Fatal("numeric addition should not match a boolean return context")
	}
	if !completionHasMatchingInfixOperator(DefaultGlobalState, OperatorKindEqual, longType, CompletionHintBoolean) {
		t.Fatal("numeric equality should match a boolean return context")
	}
	if completionHasMatchingInfixOperator(DefaultGlobalState, OperatorKindEqual, longType, CompletionHintScalar) {
		t.Fatal("equality should not match a non-boolean scalar return context")
	}
	if completionHasMatchingInfixOperator(DefaultGlobalState, OperatorKindEqual, ScalarTypes.Bool.Type(), CompletionHintBoolean) {
		t.Fatal("hidden boolean equality signature should not produce completion")
	}
	if completionParameterMatches(NewParameterKind("value", ParameterTypeInteger), ScalarTypes.Real.Type()) {
		t.Fatal("real should not match an integer parameter")
	}
	if !completionParameterMatches(NewParameter("value", ScalarTypes.Real.Type()), longType) {
		t.Fatal("long should match a promotable declared real parameter")
	}
}
