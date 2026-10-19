package kustolang

import "testing"

func TestNativePartitionFormsWrappersPropertiesAndDownstream(t *testing.T) {
	key := NewColumnSymbol("Key", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", key, value).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	for _, queryText := range []string{
		`T | partition by Key (where Value > 0 | project Key, Value) | project Value`,
		`T | __partitionby Key (where Value > 0 | project Key, Value) | project Value`,
	} {
		query := ParseQuery(queryText)
		NewBinder(globals).TryBind(query)
		result := finalProjectTable(t, query)
		assertOrderedColumns(t, result, "Value")

		var partitionTable *TableSymbol
		if node := firstSyntaxNodeOfKind(query, SyntaxKindPartitionOperator); node != nil {
			partitionTable = semanticTable(t, &PartitionOperator{SyntaxNode: node})
			wrapper := firstBoundNode[*PartitionSubquery](t, query)
			assertOrderedColumns(t, semanticTable(t, wrapper), "Key", "Value")
		} else {
			partitionTable = semanticTable(t, firstBoundNode[*PartitionByOperator](t, query))
		}
		assertOrderedColumns(t, partitionTable, "Key", "Value")
		if !partitionTable.IsOpen() || !partitionTable.IsSerialized() || partitionTable.IsSorted() {
			t.Fatalf("partition properties = open:%t serialized:%t sorted:%t", partitionTable.IsOpen(), partitionTable.IsSerialized(), partitionTable.IsSorted())
		}
	}
}

func TestNativePartitionBracedClearsRowScopeButRetainsExactKeyLocal(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Key", ScalarTypes.String.Type()), NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	braced := ParseQuery(`T | partition by Key { print K=Key, V=Value }`)
	NewBinder(globals).TryBind(braced)
	partition := firstBoundNode[*PartitionOperator](t, braced)
	keyRef := firstNameReferenceWithText(t, partition.Operand(), "Key")
	if _, ok := keyRef.ReferencedSymbol().(*ColumnSymbol); !ok {
		t.Fatalf("braced key symbol = %T, want column", keyRef.ReferencedSymbol())
	}

	parenthesized := ParseQuery(`T | partition by Key (where Value > 0)`)
	NewBinder(globals).TryBind(parenthesized)
	valueRef := firstNameReferenceWithText(t, firstBoundNode[*PartitionOperator](t, parenthesized).Operand(), "Value")
	if _, ok := valueRef.ReferencedSymbol().(*ColumnSymbol); !ok {
		t.Fatalf("parenthesized Value symbol = %T, want column", valueRef.ReferencedSymbol())
	}
}

func TestNativePartitionDiagnosticsParametersAndPostPipeValidation(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	dynamic := ParseQuery(`T | partition by D (count)`)
	NewBinder(globals).TryBind(dynamic)
	assertDiagnosticCodes(t, firstBoundNode[*PartitionOperator](t, dynamic), "KS116")
	printQuery := ParseQuery(`print X=1`)
	NewBinder(globals).TryBind(printQuery)
	if diagnostics := newSimpleNodeBinder(NewBinder(globals)).partitionOperatorDiagnostic(firstBoundNode[*PrintOperator](t, printQuery)); len(diagnostics) != 1 || diagnostics[0].Code != "KS177" {
		t.Fatalf("partition post-pipe diagnostics = %#v, want KS177", diagnostics)
	}

	badScope := ParseQuery(`T | partition by N in dynamic([1]) (count)`)
	NewBinder(globals).TryBind(badScope)
	assertDiagnosticCodes(t, firstBoundNode[*PartitionOperator](t, badScope), "KS111")

	validated := ParseQuery(`T | partition hint.concurrency=1 hint.concurrency=2 by N (count)`)
	NewBinder(globals).TryBind(validated)
	assertDiagnosticCodes(t, firstBoundNode[*PartitionOperator](t, validated), "KS180")

	partitionByParameter := ParseQuery(`T | __partitionby hint.concurrency=1 hint.concurrency=2 N (count)`)
	NewBinder(globals).TryBind(partitionByParameter)
	info, _ := firstBoundNode[*PartitionByOperator](t, partitionByParameter).GetSemanticInfo()
	for _, diagnostic := range info.Diagnostics {
		if diagnostic.Code == "KS180" {
			t.Fatal("PartitionByOperator parameters were validated")
		}
	}
}

func TestNativePartitionRecoveredOperandRestoresStateAndIdempotence(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Key", ScalarTypes.String.Type())).WithIsOpen(true)
	query := ParseQuery(`T | partition by Key`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)
	partition := firstBoundNode[*PartitionOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, partition), "Key")
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("recovered partition leaked binder state")
	}

	valid := ParseQuery(`T | partition by Key (count)`)
	tree.visitElement(valid)
	node := firstBoundNode[*PartitionOperator](t, valid)
	before, _ := node.GetSemanticInfo()
	tree.visitElement(valid)
	after, _ := node.GetSemanticInfo()
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatal("partition rebinding changed semantic information")
	}
}

func firstNameReferenceWithText(t *testing.T, root SyntaxElement, name string) *NameReference {
	t.Helper()
	var found *NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		if found == nil {
			if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(reference) == name {
				found = reference
			}
		}
	})
	if found == nil {
		t.Fatalf("name reference %q not found", name)
	}
	return found
}

func finalProjectTable(t *testing.T, root SyntaxElement) *TableSymbol {
	t.Helper()
	var found *TableSymbol
	WalkSyntax(root, func(element SyntaxElement) {
		if found == nil {
			if pipe, ok := WrapGeneratedSyntaxElement(element).(*PipeExpression); ok {
				if _, project := pipe.Operator().(*ProjectOperator); project && pipe.IsBound() {
					found = expressionTableResult(pipe)
				}
			}
		}
	})
	if found == nil {
		t.Fatal("bound final project pipe not found")
	}
	return found
}
