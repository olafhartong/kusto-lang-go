package kustolang

import "testing"

func TestNativeParseAndParseWhereColumnsTypesLineageAndProperties(t *testing.T) {
	message := NewColumnSymbol("Message", ScalarTypes.String.Type())
	table := NewTableSymbol("T", message, NewColumnSymbol("Keep", ScalarTypes.Long.Type())).
		WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	for _, text := range []string{
		`T | parse Message with Prefix:string ":" Count:long | project Prefix, Count, Keep`,
		`T | parse-where Message with Prefix:string ":" Count:long | project Prefix, Count, Keep`,
	} {
		query := ParseQuery(text)
		NewBinder(globals).TryBind(query)
		var result *TableSymbol
		var source Expression
		switch operator := firstParseFamilyNode(t, query).(type) {
		case *ParseOperator:
			assertDiagnosticCodes(t, operator)
			result, source = expressionTableResult(operator), operator.Expression()
		case *ParseWhereOperator:
			assertDiagnosticCodes(t, operator)
			result, source = expressionTableResult(operator), operator.Expression()
		}
		assertOrderedColumns(t, result, "Message", "Keep", "Prefix", "Count")
		if result.GetColumn("Prefix").Type != ScalarTypes.String.Type() || result.GetColumn("Count").Type != ScalarTypes.Long.Type() {
			t.Fatalf("%s: parsed types = %v, %v", text, result.GetColumn("Prefix").Type, result.GetColumn("Count").Type)
		}
		if result.GetColumn("Prefix").Source != source || result.GetColumn("Count").Source != source {
			t.Fatalf("%s: parsed columns did not retain source-expression lineage", text)
		}
		if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
			t.Fatalf("%s: inheritable properties were not preserved", text)
		}
		project := firstBoundNode[*ProjectOperator](t, query)
		assertOrderedColumns(t, expressionTableResult(project), "Prefix", "Count", "Keep")
	}
}

func TestNativeParsePreservesIncomingNameCollisionsAndRejectsLocalDuplicates(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Message", ScalarTypes.String.Type()), NewColumnSymbol("Value", ScalarTypes.Bool.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	collision := ParseQuery(`T | parse Message with Value:long`)
	NewBinder(globals).TryBind(collision)
	operator := firstBoundNode[*ParseOperator](t, collision)
	assertDiagnosticCodes(t, operator)
	columns := expressionTableResult(operator).Columns()
	if len(columns) != 3 || columns[1].SymbolName() != "Value" || columns[2].SymbolName() != "Value" || columns[2].Type != ScalarTypes.Long.Type() {
		t.Fatalf("collision columns = %#v, want incoming and parsed Value columns", columns)
	}

	duplicate := ParseQuery(`T | parse Message with Value ":" Value`)
	NewBinder(globals).TryBind(duplicate)
	duplicateOperator := firstBoundNode[*ParseOperator](t, duplicate)
	assertDiagnosticCodes(t, duplicateOperator, "KS171")
	if got := len(expressionTableResult(duplicateOperator).Columns()); got != 3 {
		t.Fatalf("duplicate declaration produced %d columns, want one parsed column", got)
	}
}

func TestNativeParseKvSchemaPropertiesLineageAndDownstreamScope(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Message", ScalarTypes.String.Type()), NewColumnSymbol("Keep", ScalarTypes.Bool.Type())).WithIsSerialized(true)
	query := ParseQuery(`T | parse-kv Message as (Key:string, Count:long) with (pair_delimiter=";", kv_delimiter="=", greedy=true) | project Key, Count, Keep`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	operator := firstBoundNode[*ParseKvOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := expressionTableResult(operator)
	assertOrderedColumns(t, result, "Message", "Keep", "Key", "Count")
	if result.GetColumn("Key").Type != ScalarTypes.String.Type() || result.GetColumn("Count").Type != ScalarTypes.Long.Type() {
		t.Fatalf("parse-kv types = %#v", result.Columns())
	}
	if result.GetColumn("Key").Source == nil || result.GetColumn("Count").Source == nil || !result.IsSerialized() {
		t.Fatal("parse-kv did not preserve declaration lineage or table properties")
	}
	assertOrderedColumns(t, expressionTableResult(firstBoundNode[*ProjectOperator](t, query)), "Key", "Count", "Keep")

	duplicateProperty := ParseQuery(`T | parse-kv Message as (Key:string) with (greedy=true, greedy=false)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(duplicateProperty)
	assertDiagnosticCodes(t, firstBoundNode[*ParseKvOperator](t, duplicateProperty), "KS180")
}

func TestNativeParseDiagnosticsAndMalformedRecovery(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Message", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		query string
		codes []string
	}{
		{`T | parse Message with * Value`, []string{"KS102", "KS101"}},
		{`T | parse Message with Text:string * ":" Tail`, []string{"KS103"}},
		{`T | parse (T | take 1) with Value`, []string{"KS108"}},
		{`T | parse kind=sideways Message with Value`, []string{"KS005"}},
		{`T | parse-kv (T | take 1) as (Key:string)`, []string{"KS108"}},
		{`T | parse-kv Message as (Key:string, Key:long)`, []string{"KS171"}},
		{`T | parse-kv Message as (Key:(A:string))`, []string{"KS170"}},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			node := firstParseFamilyNode(t, query)
			assertDiagnosticCodes(t, node.(interface{ SemanticDiagnostics() []Diagnostic }), test.codes...)
		})
	}

	for _, text := range []string{`parse Message with Value`, `parse-where Message with Value`, `parse-kv Message as (Key:string)`} {
		operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text)
		NewBinder(globals).TryBind(operator)
		assertDiagnosticCodes(t, operator.(interface{ SemanticDiagnostics() []Diagnostic }), "KS174")
	}
}

func TestNativeParseRestoresStateAndIsIdempotent(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Message", ScalarTypes.String.Type())).WithIsOpen(true)
	query := ParseQuery(`T | parse Message with Value:long`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)
	operator := firstBoundNode[*ParseOperator](t, query)
	before, _ := operator.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := operator.GetSemanticInfo()
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("parse binding leaked row scope, scope kind, or local scope")
	}
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatalf("parse rebinding changed semantic info: before=%#v after=%#v", before, after)
	}
}

func firstParseFamilyNode(t *testing.T, root SyntaxElement) SyntaxElement {
	t.Helper()
	if node := firstOptionalBoundNode[*ParseOperator](root); node != nil {
		return node
	}
	if node := firstOptionalBoundNode[*ParseWhereOperator](root); node != nil {
		return node
	}
	if node := firstOptionalBoundNode[*ParseKvOperator](root); node != nil {
		return node
	}
	t.Fatal("parse-family operator was not bound")
	return nil
}
