package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeDependentTableArgumentRowScopesProfiles(t *testing.T) {
	for _, profile := range kustoCodeFunctionArgumentProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			validText := `Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, Common)`
			valid := profile.parse(validText, kustoCodeFunctionArgumentGlobals())
			if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid ipv4_lookup diagnostics = %#v", diagnostics)
			}
			assertExactRowColumnNames(t, valid.GetColumnsInScope(strings.Index(validText, "SourceIP")), "SourceIP", "Common", "SourceOnly")
			assertExactRowColumnNames(t, valid.GetColumnsInScope(strings.Index(validText, "LookupIP")), "LookupIP", "Common", "LookupOnly")
			assertExactRowColumnNames(t, valid.GetColumnsInScope(strings.LastIndex(validText, "Common")), "Common")

			lookupText := `Source | evaluate ipv4_lookup(Lookup, SourceIP, )`
			lookup := profile.parse(lookupText, kustoCodeFunctionArgumentGlobals())
			if len(lookup.GetDiagnostics()) == 0 {
				t.Fatal("missing lookup key produced no diagnostics")
			}
			lookupPosition := len(lookupText) - 1
			assertExactRowColumnNames(t, lookup.GetColumnsInScope(lookupPosition), "LookupIP", "Common", "LookupOnly")
			lookupItems := NewKustoCodeServiceForCode(lookup).GetCompletionItems(lookupPosition, CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, lookupItems, "LookupIP", "Common", "LookupOnly")
			assertCompletionExcludes(t, lookupItems, "SourceIP", "SourceOnly")

			commonText := `Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, )`
			common := profile.parse(commonText, kustoCodeFunctionArgumentGlobals())
			if len(common.GetDiagnostics()) == 0 {
				t.Fatal("missing common lookup key produced no diagnostics")
			}
			commonPosition := len(commonText) - 1
			assertExactRowColumnNames(t, common.GetColumnsInScope(commonPosition), "Common")
			commonItems := NewKustoCodeServiceForCode(common).GetCompletionItems(commonPosition, CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, commonItems, "Common")
			assertCompletionExcludes(t, commonItems, "SourceIP", "SourceOnly", "LookupIP", "LookupOnly")
		})
	}
}

func TestKustoCodeTupleElementArgumentRowScopesProfiles(t *testing.T) {
	for _, profile := range kustoCodeGraphScopeProfiles()[:1] {
		t.Run(profile.name, func(t *testing.T) {
			validText := graphScopePrefix() + `graph-match (a)-[e]->(b) where all(a, Enabled)`
			valid := profile.parse(validText, kustoCodeFunctionArgumentGraphGlobals())
			if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid graph all diagnostics = %#v", diagnostics)
			}
			assertExactRowColumnNames(t, valid.GetColumnsInScope(strings.Index(validText, "Enabled")), "NodeId", "Label", "Enabled")

			malformedText := graphScopePrefix() + `graph-match (a)-[e]->(b) where all(a, )`
			malformed := profile.parse(malformedText, kustoCodeFunctionArgumentGraphGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("missing graph lambda expression produced no diagnostics")
			}
			position := len(malformedText) - 1
			assertExactRowColumnNames(t, malformed.GetColumnsInScope(position), "NodeId", "Label", "Enabled")
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(position, CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "NodeId", "Label", "Enabled")
			assertCompletionExcludes(t, items, "Source", "Target", "Weight")
		})
	}
}

type kustoCodeFunctionArgumentProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeFunctionArgumentProfiles() []kustoCodeFunctionArgumentProfile {
	return []kustoCodeFunctionArgumentProfile{{
		name: "current",
		parse: func(text string, globals *GlobalState) *KustoCode {
			return ParseAndAnalyze(text, globals)
		},
	}}
}

func kustoCodeFunctionArgumentGlobals() *GlobalState {
	source := NewTableSymbol("Source",
		NewColumnSymbol("SourceIP", ScalarTypes.String.Type()),
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("SourceOnly", ScalarTypes.Long.Type()))
	lookup := NewTableSymbol("Lookup",
		NewColumnSymbol("LookupIP", ScalarTypes.String.Type()),
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("LookupOnly", ScalarTypes.Bool.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", source, lookup))
}

func kustoCodeFunctionArgumentGraphGlobals() *GlobalState {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Long.Type()))
	nodes := NewTableSymbol("Nodes",
		NewColumnSymbol("NodeId", ScalarTypes.String.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()),
		NewColumnSymbol("Enabled", ScalarTypes.Bool.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", edges, nodes))
}

func assertExactRowColumnNames(t *testing.T, table *TableSymbol, names ...string) {
	t.Helper()
	if table == nil {
		t.Fatalf("row scope is nil, want %v", names)
	}
	columns := table.Columns()
	if len(columns) != len(names) {
		t.Fatalf("row columns = %#v, want %v", columns, names)
	}
	for index, name := range names {
		if columns[index] == nil || columns[index].SymbolName() != name {
			t.Fatalf("row columns = %#v, want %v", columns, names)
		}
	}
}
