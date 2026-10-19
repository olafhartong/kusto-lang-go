package kustolang

import "testing"

// Ported from KustoCompleter.GetGrammarAnnotations/GetAnnotatedParsers and
// GetSyntaxCompletions in src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TestKustoCompleterQueryGrammarAnnotationCompletions(t *testing.T) {
	options := CompletionOptionsDefault.WithIncludeSymbols(false)
	tests := []struct {
		name      string
		text      string
		want      []string
		malformed bool
	}{
		{name: "summarize optional by", text: "T | summarize count() ", want: []string{"by"}},
		{name: "top required by", text: "T | top 5 ", want: []string{"by"}, malformed: true},
		{name: "make series clauses", text: "T | make-series count() ", want: []string{"default", "on"}, malformed: true},
		{name: "render properties", text: "T | render timechart ", want: []string{"with"}},
		{name: "join condition", text: "T | join T2 ", want: []string{"on"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := Parse(test.text)
			if malformed := len(code.GetSyntaxDiagnostics()) != 0; malformed != test.malformed {
				t.Fatalf("syntax diagnostics present = %t, want %t: %#v", malformed, test.malformed, code.GetSyntaxDiagnostics())
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(test.text), options, CancellationToken{}).Items
			for _, display := range test.want {
				if _, ok := completionItemByDisplayText(items, display); !ok {
					t.Fatalf("completion items missing %q: %#v", display, completionDisplayTexts(items))
				}
			}
		})
	}

	text := "T | render timechart "
	disabled := options.WithIncludeSyntax(false)
	items := NewKustoCodeService(text).GetCompletionItems(len(text), disabled, CancellationToken{}).Items
	if _, ok := completionItemByDisplayText(items, "with"); ok {
		t.Fatal("syntax-disabled completion unexpectedly included grammar annotation")
	}
}

func TestKustoCompleterDataScopeTabularSuffixAnnotation(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	options := CompletionOptionsDefault.WithIncludeSymbols(false)

	validText := "T "
	validService := NewKustoCodeService(validText, globals)
	if diagnostics := validService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("valid datascope completion context diagnostics = %#v", diagnostics)
	}
	validItems := validService.GetCompletionItems(len(validText), options, CancellationToken{}).Items
	datascope, ok := completionItemByDisplayText(validItems, "datascope")
	if !ok {
		t.Fatalf("tabular suffix completions missing datascope: %#v", completionDisplayTexts(validItems))
	}
	if datascope.Kind != CompletionKindTabularSuffix {
		t.Fatalf("datascope completion kind = %v, want %v", datascope.Kind, CompletionKindTabularSuffix)
	}
	scalarText := "let X=1; X "
	scalarItems := NewKustoCodeService(scalarText, globals).
		GetCompletionItems(len(scalarText), options, CancellationToken{}).Items
	requireNoCompletion(t, scalarItems, "datascope")
	unresolvedText := "Missing "
	unresolvedItems := NewKustoCodeService(unresolvedText, globals).
		GetCompletionItems(len(unresolvedText), options, CancellationToken{}).Items
	requireNoCompletion(t, unresolvedItems, "datascope")

	malformedText := "T datascope="
	malformedService := NewKustoCodeService(malformedText, globals)
	if diagnostics := malformedService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("missing datascope value produced no diagnostics")
	}
	malformedItems := malformedService.GetCompletionItems(len(malformedText), options, CancellationToken{}).Items
	requireCompletionKind(t, malformedItems, "all", CompletionKindSyntax)
	requireCompletionKind(t, malformedItems, "hotcache", CompletionKindKeyword)

	disabledItems := validService.GetCompletionItems(
		len(validText),
		options.WithIncludeSyntax(false),
		CancellationToken{},
	).Items
	requireNoCompletion(t, disabledItems, "datascope")
}
