package kustolang

import "testing"

func TestKustoCodeIncompleteStatementsDoNotPublishDeclarationsProfiles(t *testing.T) {
	for _, profile := range kustoCodeStatementScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `let Existing=1; let Pending =`
			code := profile.parse(text)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("incomplete let should retain syntax diagnostics")
			}
			assertScopeNames(t,
				code.GetSymbolsInScope(len(text), SymbolMatchLocal, IncludeFunctionKindNone),
				[]string{"Existing"},
			)
			if got := code.GetSpeculativeReferencedSymbol(len(text), "Pending", SymbolMatchLocal); got != nil {
				t.Fatalf("incomplete declaration leaked into speculative scope: %#v", got)
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "Existing")
			assertCompletionExcludes(t, items, "Pending")

			parametersText := `declare query_parameters(p:string`
			parameters := profile.parse(parametersText)
			if diagnostics := parameters.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("incomplete query-parameters statement should retain syntax diagnostics")
			}
			if got := parameters.GetSpeculativeReferencedSymbol(len(parametersText), "p", SymbolMatchLocal); got != nil {
				t.Fatalf("incomplete query parameter leaked into scope: %#v", got)
			}
		})
	}
}

func TestKustoCodeTrailingGrowableStatementWaitsForSeparatorProfiles(t *testing.T) {
	for _, profile := range kustoCodeStatementScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			growableText := `let Existing=1; let Pending=print x=1`
			growable := profile.parse(growableText)
			if diagnostics := growable.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("growable declaration diagnostics = %#v", diagnostics)
			}
			assertScopeNames(t,
				growable.GetSymbolsInScope(len(growableText), SymbolMatchLocal, IncludeFunctionKindNone),
				[]string{"Existing"},
			)

			separatedText := growableText + `; `
			separated := profile.parse(separatedText)
			assertScopeNames(t,
				separated.GetSymbolsInScope(len(separatedText), SymbolMatchLocal, IncludeFunctionKindNone),
				[]string{"Existing", "Pending"},
			)

			fixedText := `let Ready=1`
			fixed := profile.parse(fixedText)
			assertScopeNames(t,
				fixed.GetSymbolsInScope(len(fixedText), SymbolMatchLocal, IncludeFunctionKindNone),
				[]string{"Ready"},
			)
		})
	}
}

type kustoCodeStatementScopeProfile struct {
	name  string
	parse func(string) *KustoCode
}

func kustoCodeStatementScopeProfiles() []kustoCodeStatementScopeProfile {
	return []kustoCodeStatementScopeProfile{{
		name:  "current",
		parse: func(text string) *KustoCode { return ParseAndAnalyze(text) },
	}}
}
