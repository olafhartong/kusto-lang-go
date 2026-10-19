package kustolang

import (
	"fmt"
	"strings"
	"testing"
)

func TestKustoCompleterOffersComputedColumnExamples(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events",
		NewColumnSymbol("Name", ScalarTypes.String.Type()),
		NewColumnSymbol("Flag", ScalarTypes.Bool.Type()))))
	tests := []struct {
		expression string
		want       map[string]string
	}{
		{`iff(Flag, 'yes', 'no')`, map[string]string{"yes": "'yes'", "no": "'no'"}},
		{`case(Name == 'condition-only', 'yes', Flag, 'maybe', 'no')`, map[string]string{"yes": "'yes'", "maybe": "'maybe'", "no": "'no'"}},
		{`coalesce(Name, iff(Flag, 'yes', 'no'), 'fallback')`, map[string]string{"yes": "'yes'", "no": "'no'", "fallback": "'fallback'"}},
		{`iff(Flag, int(1), long(2))`, map[string]string{"1": "int(1)", "2": "long(2)"}},
		{`iff(Flag, h'private', 'public')`, map[string]string{"private": "h'private'", "public": "'public'"}},
		{`iff(Flag, 'a' h'b', 'c' 'd')`, map[string]string{"ab": "h'ab'", "cd": "'cd'"}},
		{`iff(Flag, dynamic([1,2]), dynamic({"x":1}))`, map[string]string{"[1,2]": "dynamic([1,2])", `{"x":1}`: `dynamic({"x":1})`}},
	}
	for _, test := range tests {
		for _, suffix := range []string{" == ", " != ", " =~ ", " in ("} {
			query := "Events | extend Computed = " + test.expression + " | where Computed" + suffix
			t.Run(query, func(t *testing.T) {
				items := NewKustoCodeService(query, globals).GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
				for display, edit := range test.want {
					item := requireCompletionKind(t, items, display, CompletionKindExample)
					if item.BeforeText() != edit {
						t.Fatalf("%s edit = %q, want %q", display, item.BeforeText(), edit)
					}
				}
				requireNoCompletion(t, items, "condition-only")
			})
		}
	}
}

func TestComputedColumnExampleBoundariesAndPrecedence(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events",
		NewColumnSymbol("Flag", ScalarTypes.Bool.Type()))))
	for _, expression := range []string{`'constant-only'`, `strcat('not-a-value-argument', 'other')`} {
		query := "Events | extend Computed = " + expression + " | where Computed == "
		items := NewKustoCodeService(query, globals).GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
		requireNoCompletion(t, items, "constant-only")
		requireNoCompletion(t, items, "not-a-value-argument")
	}
	for _, test := range []struct {
		value   string
		offered bool
	}{
		{strings.Repeat("a", 98), true}, {strings.Repeat("a", 99), false},
		{strings.Repeat("😀", 49), true}, {strings.Repeat("😀", 50), false},
	} {
		query := "Events | extend Computed = iff(Flag, '" + test.value + "', 'short') | where Computed == "
		items := NewKustoCodeService(query, globals).GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
		if test.offered {
			requireCompletionKind(t, items, test.value, CompletionKindExample)
		} else {
			requireNoCompletion(t, items, test.value)
		}
		requireCompletionKind(t, items, "short", CompletionKindExample)
	}
	var expression strings.Builder
	expression.WriteString("case(")
	for i := 0; i < 110; i++ {
		fmt.Fprintf(&expression, "Flag, 'choice%d', ", i)
	}
	expression.WriteString("'final')")
	query := "Events | extend Computed = " + expression.String() + " | where Computed == "
	items := NewKustoCodeService(query, globals).GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "choice98", CompletionKindExample)
	requireNoCompletion(t, items, "choice99")
	requireNoCompletion(t, items, "final")

	code := ParseAndAnalyze("print Computed = iff(true, 'computed', 'other')")
	table, _ := TableSymbolForType(code.ResultType)
	column := table.GetColumn("Computed").WithExamples([]string{"'declared'"})
	declaredGlobals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Declared", column)))
	query = "Declared | where Computed == "
	items = NewKustoCodeService(query, declaredGlobals).GetCompletionItems(len(query), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "declared", CompletionKindExample)
	requireNoCompletion(t, items, "computed")
}
