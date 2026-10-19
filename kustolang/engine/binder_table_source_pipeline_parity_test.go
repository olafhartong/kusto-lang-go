package kustolang

import "testing"

func TestNativeTableSourcePipelinesUnionSummarizeAndJoinFunctions(t *testing.T) {
	left := NewTableSymbol("Left",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("Value", ScalarTypes.String.Type()),
	)
	right := NewTableSymbol("Right",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	unionQuery := ParseQuery("union Left, (Right | summarize Total=sum(Score) by Key) | project Key, Value, Total")
	NewBinder(globals).TryBind(unionQuery)
	union := firstBoundNode[*UnionOperator](t, unionQuery)
	assertDiagnosticCodes(t, union)
	assertOrderedColumns(t, semanticTable(t, union), "Key", "Value", "Total")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, unionQuery)), "Key", "Value", "Total")

	joinQuery := ParseQuery("Left | join (Right | where strlen(Label) > 0 | summarize Total=sum(Score) by Key) on Key | project Value, Total")
	NewBinder(globals).TryBind(joinQuery)
	join := firstBoundNode[*JoinOperator](t, joinQuery)
	assertDiagnosticCodes(t, join)
	assertOrderedColumns(t, semanticTable(t, join), "Key", "Value", "Key1", "Total")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, joinQuery)), "Value", "Total")
}

func TestNativeTableSourcePipelinesParenthesizedDataTableAndTableCall(t *testing.T) {
	right := NewTableSymbol("Right", NewColumnSymbol("FromTable", ScalarTypes.String.Type()))
	query := ParseQuery(`union (datatable(FromData:long)[1]), (table("Right")) | project FromData, FromTable`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", right))).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertDiagnosticCodes(t, union)
	assertOrderedColumns(t, semanticTable(t, union), "FromData", "FromTable")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "FromData", "FromTable")
}

func TestNativeTableSourcePipelinesLookupDistinctExtendAndJoinMvExpand(t *testing.T) {
	left := NewTableSymbol("Left",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("Value", ScalarTypes.String.Type()),
	)
	right := NewTableSymbol("Right",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()),
		NewColumnSymbol("Items", ScalarTypes.Dynamic.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	lookupQuery := ParseQuery("Left | lookup (Right | extend LabelLength=strlen(Label) | distinct Key, Label, LabelLength) on Key | project Key, Value, Label, LabelLength")
	NewBinder(globals).TryBind(lookupQuery)
	lookup := firstBoundNode[*LookupOperator](t, lookupQuery)
	assertDiagnosticCodes(t, lookup)
	assertOrderedColumns(t, semanticTable(t, lookup), "Key", "Value", "Label", "LabelLength")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, lookupQuery)), "Key", "Value", "Label", "LabelLength")

	joinQuery := ParseQuery("Left | join (Right | mv-expand Item=Items to typeof(long) | summarize MaxItem=max(Item) by Key) on Key | project Value, MaxItem")
	NewBinder(globals).TryBind(joinQuery)
	join := firstBoundNode[*JoinOperator](t, joinQuery)
	assertDiagnosticCodes(t, join)
	assertOrderedColumns(t, semanticTable(t, join), "Key", "Value", "Key1", "MaxItem")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, joinQuery)), "Value", "MaxItem")
}

func TestNativeTableSourcePipelinesKeepChildDiagnosticsAndDiagnoseParenthesizedScalars(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	childError := ParseQuery("Left | join (Right | where Missing > 0 | project Key) on Key | count")
	NewBinder(globals).TryBind(childError)
	join := firstBoundNode[*JoinOperator](t, childError)
	if !join.IsBound() || !diagnosticsContain(GetContainedDiagnosticsPreserveDuplicates(join, DiagnosticsIncludeDefault), "KS142") {
		t.Fatalf("join child diagnostics = bound:%t diagnostics:%#v", join.IsBound(), join.ContainedDiagnostics())
	}
	if !firstBoundNode[*CountOperator](t, childError).IsBound() {
		t.Fatal("child semantic error prevented downstream binding")
	}

	for _, text := range []string{
		"union (1) | count",
		"Left | join (1) on Key | count",
		"Left | lookup (1) on Key | count",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(globals).TryBind(query)
			if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault); !diagnosticsContain(diagnostics, "KS111") {
				t.Fatalf("diagnostics = %#v, want KS111", diagnostics)
			}
			if !firstBoundNode[*CountOperator](t, query).IsBound() {
				t.Fatal("parenthesized scalar diagnostic prevented downstream binding")
			}
		})
	}
}

func TestNativeTableSourcePipelinesParserRecoveredMissingCloseBinds(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right",
		NewColumnSymbol("Key", ScalarTypes.Long.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))
	for _, text := range []string{
		"union (Right | summarize Total=sum(Score) by Key",
		"Left | join (Right | summarize Total=sum(Score) by Key on Key | count",
		"Left | lookup (Right | extend Total=Score + 1 on Key | count",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(globals).TryBind(query)
			parenthesized := firstBoundNode[*ParenthesizedExpression](t, query)
			if parenthesized.CloseParen().Width() != 0 ||
				!diagnosticsContainMessage(parenthesized.CloseParen().Diagnostics(), "KS005", "Expected: )") ||
				!parenthesized.IsBound() {
				t.Fatalf("missing close recovery = bound:%t diagnostics:%#v", parenthesized.IsBound(), parenthesized.CloseParen().Diagnostics())
			}
		})
	}
}

func TestNativeTableSourcePipelinesInjectedMissingCloseRollsBack(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))
	for _, text := range []string{
		"union (Right | project Key)",
		"Left | join (Right | project Key) on Key",
		"Left | lookup (Right | project Key) on Key",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			parenthesized := firstBoundNode[*ParenthesizedExpression](t, query)
			parenthesized.SyntaxNode.children[2] = NewMissingToken(SyntaxKindCloseParenToken, Diagnostic{})
			binder := NewBinder(globals)
			outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
			binder.rowScope = outer
			binder.scopeKind = ScopeKindAggregate
			local := binder.LocalScope
			tree := newTreeBinder(binder)
			tree.inPipe = true
			tree.inWildcard = true
			tree.bind(query)

			if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
				!tree.inPipe || !tree.inWildcard {
				t.Fatal("malformed operand changed binder state")
			}
			for _, node := range boundNodes[*SyntaxNode](query) {
				if node.IsBound() {
					t.Fatalf("malformed operand committed node semantics: %v", node.Kind())
				}
			}
		})
	}
}

func TestNativeTableSourcePipelineRollbackRestoresOpenTableInference(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type())).WithIsOpen(true)
	query := ParseQuery("Left | join (Right | where Future > 0 | summarize C=count()) on Key")
	summarize := firstBoundNode[*SummarizeOperator](t, query)
	summarize.SyntaxNode.children[0] = nil

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	existing := binder.getOpenTableInferredColumn("Existing", right)
	newTreeBinder(binder).bind(query)
	if firstBoundNode[*JoinOperator](t, query).IsBound() {
		t.Fatal("join with injected malformed inner operator was bound")
	}
	columns := binder.openTableInferredColumns[right]
	order := binder.openTableInferredOrder[right]
	if len(columns) != 1 || columns["Existing"] != existing || len(order) != 1 || order[0] != existing {
		t.Fatalf("open-table inference after rollback = columns:%#v order:%#v", columns, order)
	}
}
