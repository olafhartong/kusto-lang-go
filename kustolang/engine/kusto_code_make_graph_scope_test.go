package kustolang

import (
	"testing"
)

func makeGraphScopeTestGlobals() *GlobalState {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Long.Type()))
	nodes := NewTableSymbol("Nodes",
		NewColumnSymbol("NodeId", ScalarTypes.String.Type()),
		NewColumnSymbol("NodeLabel", ScalarTypes.String.Type()))
	other := NewTableSymbol("OtherNodes", NewColumnSymbol("OtherId", ScalarTypes.String.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", edges, nodes, other))
}

func assertCompletionIncludes(t *testing.T, items []CompletionItem, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, ok := completionItemByDisplayText(items, name); !ok {
			t.Fatalf("completion missing %q: %#v", name, completionDisplayTexts(items))
		}
	}
}

func assertCompletionExcludes(t *testing.T, items []CompletionItem, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, ok := completionItemByDisplayText(items, name); ok {
			t.Fatalf("completion unexpectedly includes %q: %#v", name, completionDisplayTexts(items))
		}
	}
}
