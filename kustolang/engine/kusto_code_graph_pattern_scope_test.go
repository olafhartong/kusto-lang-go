package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeGraphPatternAliasScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeGraphScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, operator := range profile.operators {
				t.Run(operator, func(t *testing.T) {
					text := graphScopePrefix() + operator + ` (a)-[e]->(b) where a.Label != "" project Src=a.NodeId, Weight=e.Weight | project Src`
					code := profile.parse(text, kustoCodeGraphScopeGlobals())
					if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
						t.Fatalf("valid %s diagnostics = %#v", operator, diagnostics)
					}

					before := strings.Index(text, "(a)") + 1
					if symbols := code.GetSymbolsInScope(before, SymbolMatchAny, IncludeFunctionKindAll); containsScopeSymbolName(symbols, "a") || containsScopeSymbolName(symbols, "e") || containsScopeSymbolName(symbols, "b") {
						t.Fatalf("%s aliases visible before clause: %#v", operator, symbols)
					}
					where := strings.Index(text, "where")
					assertGraphPatternAliasSymbols(t, code, where)

					project := strings.Index(text, "project Src")
					assertGraphPatternAliasSymbols(t, code, project)
					if symbols := code.GetSymbolsInScope(strings.LastIndex(text, "project Src")+len("project "), SymbolMatchAny, IncludeFunctionKindAll); containsScopeSymbolName(symbols, "a") || containsScopeSymbolName(symbols, "e") || containsScopeSymbolName(symbols, "b") {
						t.Fatalf("%s aliases leaked downstream: %#v", operator, symbols)
					}
				})
			}
		})
	}
}

func TestKustoCompleterGraphPatternAliasMembersProfiles(t *testing.T) {
	for _, profile := range kustoCodeGraphScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, operator := range profile.operators {
				t.Run(operator, func(t *testing.T) {
					root := graphScopePrefix() + operator + ` (a)-[e]->(b) where `
					code := profile.parse(root, kustoCodeGraphScopeGlobals())
					if len(code.GetDiagnostics()) == 0 {
						t.Fatalf("malformed %s missing predicate produced no diagnostics", operator)
					}
					assertGraphPatternAliasSymbols(t, code, len(root))
					items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(root), CompletionOptionsDefault, CancellationToken{}).Items
					assertCompletionIncludes(t, items, "a", "e", "b")

					nodePath := root + "a."
					nodeCode := profile.parse(nodePath, kustoCodeGraphScopeGlobals())
					if len(nodeCode.GetDiagnostics()) == 0 {
						t.Fatalf("malformed %s node path produced no diagnostics", operator)
					}
					nodeItems := NewKustoCodeServiceForCode(nodeCode).GetCompletionItems(len(nodePath), CompletionOptionsDefault, CancellationToken{}).Items
					assertCompletionIncludes(t, nodeItems, "NodeId", "Label")
					assertCompletionExcludes(t, nodeItems, "Weight", "Source", "Target", "a", "e", "b")

					edgePath := root + "e."
					edgeCode := profile.parse(edgePath, kustoCodeGraphScopeGlobals())
					edgeItems := NewKustoCodeServiceForCode(edgeCode).GetCompletionItems(len(edgePath), CompletionOptionsDefault, CancellationToken{}).Items
					assertCompletionIncludes(t, edgeItems, "Source", "Target", "Weight")
					assertCompletionExcludes(t, edgeItems, "NodeId", "Label", "a", "e", "b")
				})
			}
		})
	}
}

func TestKustoCodeGraphPatternProjectOnlyScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeGraphScopeProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := graphScopePrefix() + `graph-match (a)-[e]->(b) project Src=a.NodeId`
			code := profile.parse(text, kustoCodeGraphScopeGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("project-only graph-match diagnostics = %#v", diagnostics)
			}
			assertGraphPatternAliasSymbols(t, code, strings.Index(text, "project"))
		})
	}
}

type kustoCodeGraphScopeProfile struct {
	name      string
	operators []string
	parse     func(string, *GlobalState) *KustoCode
}

func kustoCodeGraphScopeProfiles() []kustoCodeGraphScopeProfile {
	return []kustoCodeGraphScopeProfile{{
		name:      "current",
		operators: []string{"graph-match", "graph-shortest-paths"},
		parse: func(text string, globals *GlobalState) *KustoCode {
			return ParseAndAnalyze(text, globals)
		},
	}}
}

func graphScopePrefix() string {
	return `Edges | make-graph Source --> Target with Nodes on NodeId | `
}

func kustoCodeGraphScopeGlobals() *GlobalState {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Long.Type()))
	nodes := NewTableSymbol("Nodes",
		NewColumnSymbol("NodeId", ScalarTypes.String.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", edges, nodes))
}

func assertGraphPatternAliasSymbols(t *testing.T, code *KustoCode, position int) {
	t.Helper()
	symbols := code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll)
	for _, name := range []string{"a", "e", "b"} {
		if !containsScopeSymbolName(symbols, name) {
			t.Fatalf("graph scope missing %q: %#v", name, symbols)
		}
	}
	a, ok := code.GetSpeculativeReferencedSymbol(position, "a", SymbolMatchAny).(*VariableSymbol)
	if !ok {
		t.Fatalf("node alias a is not a VariableSymbol")
	}
	if tuple, ok := a.Type.(*TupleSymbol); !ok || FindFirstMemberMatch(tuple, "NodeId", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Label", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Weight", SymbolMatchColumn, false) != nil {
		t.Fatalf("node alias tuple = %#v", a.Type)
	}
	e, ok := code.GetSpeculativeReferencedSymbol(position, "e", SymbolMatchAny).(*VariableSymbol)
	if !ok {
		t.Fatalf("edge alias e is not a VariableSymbol")
	}
	if tuple, ok := e.Type.(*TupleSymbol); !ok || FindFirstMemberMatch(tuple, "Source", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Target", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Weight", SymbolMatchColumn, false) == nil || FindFirstMemberMatch(tuple, "Label", SymbolMatchColumn, false) != nil {
		t.Fatalf("edge alias tuple = %#v", e.Type)
	}
}
