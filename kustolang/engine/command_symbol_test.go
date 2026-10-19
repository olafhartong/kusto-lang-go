package kustolang

import (
	"sync"
	"testing"
)

func TestCommandSymbolWithResultTypeBuildsSchemaDisplay(t *testing.T) {
	table := NewTableSymbol("",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Count", ScalarTypes.Long.Type()),
	)
	command := NewCommandSymbolWithResultType(".show test", table, ".show test")
	if command.SymbolKind() != SymbolKindCommand {
		t.Fatalf("kind = %s", command.SymbolKind())
	}
	if got, want := command.ResultSchema(), "(Account: string, Count: long)"; got != want {
		t.Fatalf("ResultSchema = %q, want %q", got, want)
	}
	if command.ResultType() != table {
		t.Fatalf("ResultType did not return original table")
	}
	if command.Construction() != ".show test" {
		t.Fatalf("Construction = %q", command.Construction())
	}
}

func TestCommandSymbolWithResultSchemaParsesTable(t *testing.T) {
	command := NewCommandSymbol(".show schema", "(Account: string, ['bad name']: real)")
	table := command.ResultType()
	if table == nil {
		t.Fatalf("ResultType was nil")
	}
	columns := table.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "Account" || columns[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("first column = %#v", columns)
	}
	if columns[1].SymbolName() != "bad name" || columns[1].Type != ScalarTypes.Real.Type() {
		t.Fatalf("second column = %#v", columns[1])
	}
	if got := command.ResultSchema(); got != "(Account: string, ['bad name']: real)" {
		t.Fatalf("ResultSchema = %q", got)
	}
}

func TestCommandSymbolResultTypeKeepsValidMalformedSchemaPrefix(t *testing.T) {
	symbol := NewCommandSymbol("MalformedMetadata", "(TableId: guid, long ShardGroupCount: long, Tail: string)")
	table := symbol.ResultType()
	if table == nil || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "TableId" ||
		table.Columns()[0].Type != ScalarTypes.Guid.Type() {
		t.Fatalf("malformed command result = %#v", table)
	}
}

func TestCommandSymbolSpecialOpenSchemas(t *testing.T) {
	for _, schema := range []string{"()", "(*)"} {
		command := NewCommandSymbol(".show open", schema)
		table := command.ResultType()
		if table == nil || !table.IsOpen() || len(table.Columns()) != 0 {
			t.Fatalf("ResultType(%q) = %#v", schema, table)
		}
	}
}

func TestCommandSymbolLazyResultsAreConcurrentSafe(t *testing.T) {
	fromSchema := NewCommandSymbol(".show schema", "(Account: string, Count: long)")
	fromType := NewCommandSymbolWithResultType(".show type", NewTableSymbol("",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Count", ScalarTypes.Long.Type()),
	))
	var wait sync.WaitGroup
	for i := 0; i < 32; i++ {
		wait.Add(2)
		go func() {
			defer wait.Done()
			if table := fromSchema.ResultType(); table == nil || len(table.Columns()) != 2 {
				t.Errorf("concurrent ResultType = %#v", table)
			}
		}()
		go func() {
			defer wait.Done()
			if schema := fromType.ResultSchema(); schema != "(Account: string, Count: long)" {
				t.Errorf("concurrent ResultSchema = %q", schema)
			}
		}()
	}
	wait.Wait()
}
