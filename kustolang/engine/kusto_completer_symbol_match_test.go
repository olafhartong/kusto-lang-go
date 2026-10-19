package kustolang

import (
	"strings"
	"testing"
)

// Ported from KustoCompleter.GetSymbolMatch,
// GetSymbolMatchFromGrammar and GetSymbolMatchFromSyntaxTree in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TestKustoCompleterGeneratedSyntaxSymbolMatches(t *testing.T) {
	globals := completionSymbolMatchGlobals()
	tests := []struct {
		name      string
		text      string
		position  int
		want      SymbolMatch
		dontWant  SymbolMatch
		malformed bool
	}{
		{
			name:     "column grammar slot",
			text:     "Events | project-away Account",
			position: strings.Index("Events | project-away Account", "Account"),
			want:     SymbolMatchColumn | SymbolMatchFunction | SymbolMatchLocal | SymbolMatchScalar,
			dontWant: SymbolMatchTable | SymbolMatchNonScalar,
		},
		{
			name:     "tabular grammar slot",
			text:     "Events | union Archive",
			position: strings.Index("Events | union Archive", "Archive"),
			want: SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchFunction | SymbolMatchLocal |
				SymbolMatchEntityGroupElement | SymbolMatchNonScalar | SymbolMatchMaterializedView | SymbolMatchGraph,
			dontWant: SymbolMatchScalar | SymbolMatchColumn,
		},
		{
			name:      "malformed column list recovery",
			text:      "Events | project-away Account, ",
			position:  len("Events | project-away Account, "),
			want:      SymbolMatchColumn | SymbolMatchFunction | SymbolMatchLocal | SymbolMatchScalar,
			dontWant:  SymbolMatchTable | SymbolMatchNonScalar,
			malformed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.text, globals)
			if malformed := len(code.GetSyntaxDiagnostics()) != 0; malformed != test.malformed {
				t.Fatalf("syntax diagnostics present = %t, want %t: %#v", malformed, test.malformed, code.GetSyntaxDiagnostics())
			}
			match := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{}).getSymbolMatch(test.position)
			if match&test.want != test.want {
				t.Fatalf("symbol match for %q = %v, missing %v", test.text, match, test.want)
			}
			if match&test.dontWant != 0 {
				t.Fatalf("symbol match for %q = %v, unexpectedly contains %v", test.text, match, test.dontWant)
			}
		})
	}
}

func TestKustoCompleterGrammarSymbolMatchPrecedesSyntaxFallback(t *testing.T) {
	text := "restrict access to ("
	code := ParseAndAnalyze(text, completionSymbolMatchGlobals())
	if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatalf("malformed query %q produced no syntax diagnostics", text)
	}
	completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
	want := SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView | SymbolMatchGraphModel
	if grammar := completer.getSymbolMatchFromGrammar(len(text)); grammar != want {
		t.Fatalf("grammar symbol match = %v, want %v", grammar, want)
	}
	if match := completer.getSymbolMatch(len(text)); match != want {
		t.Fatalf("final symbol match = %v, want authoritative grammar match %v", match, want)
	}

	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	externalDisplay := CompletionDisplayText(code.Globals.Database.GetExternalTable("ExternalEvents"))
	for _, display := range []string{"Events", "Archive", externalDisplay, "DailyEvents", "SecurityGraph"} {
		if _, ok := completionItemByDisplayText(items, display); !ok {
			t.Fatalf("restriction completion missing %q: %#v", display, completionDisplayTexts(items))
		}
	}
	for _, display := range []string{"Account", CompletionDisplayText(code.Globals.GetFunction("ScalarFn")), CompletionDisplayText(code.Globals.GetFunction("TabularFn"))} {
		if _, ok := completionItemByDisplayText(items, display); ok {
			t.Fatalf("restriction completion unexpectedly includes %q: %#v", display, completionDisplayTexts(items))
		}
	}
}

func TestKustoCompleterSymbolMatchFallsBackToSyntaxTree(t *testing.T) {
	tests := []struct {
		name string
		text string
		want SymbolMatch
	}{
		{
			name: "missing scalar expression",
			text: "print x=",
			want: SymbolMatchColumn | SymbolMatchFunction | SymbolMatchLocal | SymbolMatchScalar,
		},
		{
			name: "missing non-scalar join operand",
			text: "Events | join (",
			want: SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchFunction | SymbolMatchLocal |
				SymbolMatchEntityGroupElement | SymbolMatchNonScalar | SymbolMatchMaterializedView | SymbolMatchGraph,
		},
		{
			name: "missing render column parameter",
			text: "Events | render timechart with (xcolumn=",
			want: SymbolMatchColumn,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.text, completionSymbolMatchGlobals())
			if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) == 0 {
				t.Fatalf("malformed query %q produced no syntax diagnostics", test.text)
			}
			completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
			if grammar := completer.getSymbolMatchFromGrammar(len(test.text)); grammar != SymbolMatchNone {
				t.Fatalf("grammar symbol match for %q = %v, want none", test.text, grammar)
			}
			if syntax := completer.getSymbolMatchFromSyntaxTree(len(test.text)); syntax != test.want {
				t.Fatalf("syntax symbol match for %q = %v, want %v", test.text, syntax, test.want)
			}
			if match := completer.getSymbolMatch(len(test.text)); match != test.want {
				t.Fatalf("final symbol match for %q = %v, want syntax fallback %v", test.text, match, test.want)
			}
		})
	}
}

func TestKustoCompleterFiltersSymbolsByGrammarMatch(t *testing.T) {
	globals := completionSymbolMatchGlobals()
	scalarDisplay := CompletionDisplayText(globals.GetFunction("ScalarFn"))
	tabularDisplay := CompletionDisplayText(globals.GetFunction("TabularFn"))
	externalDisplay := CompletionDisplayText(globals.Database.GetExternalTable("ExternalEvents"))
	tests := []struct {
		name      string
		text      string
		position  int
		want      []string
		dontWant  []string
		malformed bool
	}{
		{
			name:     "valid project-away column",
			text:     "Events | project-away Account",
			position: strings.Index("Events | project-away Account", "Account"),
			want:     []string{"Account", "Flag", scalarDisplay},
			dontWant: []string{tabularDisplay, "Archive"},
		},
		{
			name:      "malformed project-away trailing comma",
			text:      "Events | project-away Account, ",
			position:  len("Events | project-away Account, "),
			want:      []string{"Account", "Flag", scalarDisplay},
			dontWant:  []string{tabularDisplay, "Archive"},
			malformed: true,
		},
		{
			name:     "valid union filters scalar local",
			text:     "let ScalarLocal=1; let TableLocal=Events; Events | union Archive",
			position: strings.Index("let ScalarLocal=1; let TableLocal=Events; Events | union Archive", "Archive"),
			want:     []string{"Events", "Archive", externalDisplay, "DailyEvents", "TableLocal", tabularDisplay},
			dontWant: []string{"Account", "Flag", "ScalarLocal", scalarDisplay},
		},
		{
			name:     "valid where filters tabular local",
			text:     "let ScalarLocal=1; let TableLocal=Events; Events | where Flag",
			position: strings.Index("let ScalarLocal=1; let TableLocal=Events; Events | where Flag", "Flag"),
			want:     []string{"Account", "Flag", "ScalarLocal", scalarDisplay},
			dontWant: []string{"Archive", "TableLocal", tabularDisplay},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.text, globals)
			if malformed := len(code.GetSyntaxDiagnostics()) != 0; malformed != test.malformed {
				t.Fatalf("syntax diagnostics present = %t, want %t: %#v", malformed, test.malformed, code.GetSyntaxDiagnostics())
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(test.position, CompletionOptionsDefault, CancellationToken{}).Items
			for _, display := range test.want {
				if _, ok := completionItemByDisplayText(items, display); !ok {
					t.Fatalf("completion items for %q missing %q: %#v", test.text, display, completionDisplayTexts(items))
				}
			}
			for _, display := range test.dontWant {
				if _, ok := completionItemByDisplayText(items, display); ok {
					t.Fatalf("completion items for %q unexpectedly include %q: %#v", test.text, display, completionDisplayTexts(items))
				}
			}
		})
	}
}

func completionSymbolMatchGlobals() *GlobalState {
	events := NewTableSymbol("Events",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Flag", ScalarTypes.Bool.Type()),
	)
	archive := NewTableSymbol("Archive", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("ExternalEvents", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	view := NewMaterializedViewSymbol("DailyEvents", []*ColumnSymbol{NewColumnSymbol("Account", ScalarTypes.String.Type())}, "Events")
	graph := NewGraphModelSymbolFromNames("SecurityGraph", []string{"Edges"}, []string{"Nodes"}, nil)
	scalar := NewScalarFunctionSymbol("ScalarFn", ScalarTypes.Bool.Type())
	tabular := NewComputedFunctionSymbol("TabularFn", "Events", TabularityTabular)
	return NewGlobalState(NewDatabaseSymbol("db", events, archive, external, view, graph)).WithFunctions([]*FunctionSymbol{scalar, tabular})
}
