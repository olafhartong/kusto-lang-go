package kustolang

import "testing"

func TestDebugDisplayCoreSymbols(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	count := NewColumnSymbol("Count", ScalarTypes.Long.Type())
	table := NewTableSymbol("Events", account, count)
	database := NewDatabaseSymbolWithAlternateName("db", "pretty", []Symbol{table}, false)
	cluster := NewClusterSymbol("help")
	group := NewGroupSymbol(account, ScalarTypes.Bool.Type())

	tests := []struct {
		name   string
		symbol Symbol
		want   string
	}{
		{"nil", nil, ""},
		{"column", account, "Account: string"},
		{"table", table, "Events: (Account: string, Count: long)"},
		{"database", database, "'pretty'"},
		{"cluster", cluster, "'help'"},
		{"group", group, "[Account: string, bool]"},
		{"stored", NewStoredQueryResultSymbol("Recent", account), "Recent: (Account: string)"},
		{"external", NewExternalTableSymbol("Archive", account), "Archive: (Account: string)"},
		{"materialized", NewMaterializedViewSymbol("Daily", []*ColumnSymbol{count}, "Events | summarize count()"), "Daily: (Count: long)"},
	}
	for _, tt := range tests {
		if got := DebugDisplayText(tt.symbol); got != tt.want {
			t.Fatalf("%s debug display = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDebugDisplayDynamicTupleGraphAndEntityGroupSymbols(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Real.Type())
	edge := NewTableSymbol("Edges", NewColumnSymbol("Source", ScalarTypes.String.Type()))
	node := NewTableSymbol("Nodes", NewColumnSymbol("Id", ScalarTypes.String.Type()))

	tests := []struct {
		name   string
		symbol Symbol
		want   string
	}{
		{"dynamic", ScalarTypes.Dynamic, "dynamic"},
		{"dynamic primitive", ScalarTypes.DynamicLong, "dynamic(long)"},
		{"dynamic array any", ScalarTypes.DynamicArray, "dynamic([])"},
		{"dynamic array typed", ScalarTypes.DynamicArrayOfString, "dynamic([string])"},
		{"dynamic bag empty", ScalarTypes.DynamicBag, "dynamic({})"},
		{"dynamic bag typed", NewDynamicBagSymbol(account, score), "dynamic({Account: string, Score: real})"},
		{"tuple", NewTupleSymbolOf(account, score), "[Account: string, Score: real]"},
		{"graph edge", NewGraphSymbol("G", edge, nil), "G: [EdgeEdges: (Source: string)]"},
		{"graph node", NewGraphSymbol("G", edge, node), "G: [EdgeEdges: (Source: string), NodeNodes: (Id: string)]"},
		{"entity group definition", NewEntityGroupSymbolFromDefinition("prod", "database('d')", ""), "prod: database('d')"},
		{"entity group members", NewEntityGroupSymbol("mixed", edge, node), "mixed: Edges: (Source: string), Nodes: (Id: string)"},
	}
	for _, tt := range tests {
		if got := DebugDisplayText(tt.symbol); got != tt.want {
			t.Fatalf("%s debug display = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDebugDisplayFunctionsSignaturesAndParameters(t *testing.T) {
	required := NewParameter("value", ScalarTypes.String.Type())
	optional := NewParameter("limit", ScalarTypes.Long.Type(), WithParameterOccurrence(0, 1))
	repeatable := NewParameterKind("terms", ParameterTypeStringOrDynamic, WithParameterOccurrence(1, MaxRepeat))
	fn := NewFunctionSymbol("lookup", NewSignature(ScalarTypes.Bool.Type(), required, optional, repeatable))

	if got, want := DebugDisplayText(fn), "lookup = (value, [limit], terms, ...)"; got != want {
		t.Fatalf("function debug display = %q, want %q", got, want)
	}
	if got, want := DebugDisplayParameterText(NewParameterTypes("choice", []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()})), "choice: string|long"; got != want {
		t.Fatalf("declared parameter debug display = %q, want %q", got, want)
	}
	if got, want := DebugDisplayParameterText(NewParameterKind("rows", ParameterTypeTabular)), "rows: <Tabular>"; got != want {
		t.Fatalf("kind parameter debug display = %q, want %q", got, want)
	}
	signature := fn.Signatures()[0]
	if got, want := DebugDisplaySignatureText(signature, true, true), "lookup(value: string, [limit: long], terms: string|dynamic, ...)"; got != want {
		t.Fatalf("verbose signature debug display = %q, want %q", got, want)
	}
	linked := NewSignature(
		ScalarTypes.String.Type(),
		NewParameter("first", ScalarTypes.Long.Type()),
		NewParameterKind("same", ParameterTypeParameter0),
		NewParameterKind("fallback", ParameterTypeAny),
	)
	if got, want := DebugDisplaySignatureText(linked, false, true), "(first: long, same: long, fallback: scalar)"; got != want {
		t.Fatalf("parameter-kind signature debug display = %q, want %q", got, want)
	}
}
