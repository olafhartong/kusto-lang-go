package kustolang

import "testing"

func TestNativeJoinLookupQualifiedConditionParity(t *testing.T) {
	leftKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	leftTenant := NewColumnSymbol("Tenant", ScalarTypes.String.Type())
	leftDynamic := NewColumnSymbol("DynamicKey", ScalarTypes.Dynamic.Type())
	left := NewTableSymbol("Left",
		leftKey,
		leftTenant,
		leftDynamic,
		NewColumnSymbol("LOnly", ScalarTypes.String.Type()),
	)
	rightOtherKey := NewColumnSymbol("OtherKey", ScalarTypes.Long.Type())
	rightTenant := NewColumnSymbol("Tenant", ScalarTypes.String.Type())
	right := NewTableSymbol("Right",
		rightOtherKey,
		rightTenant,
		NewColumnSymbol("DynamicKey", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("TextKey", ScalarTypes.String.Type()),
		NewColumnSymbol("ROnly", ScalarTypes.Bool.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	t.Run("qualified equality", func(t *testing.T) {
		query := ParseQuery("Left | join (Right) on $left.Key == $right.OtherKey")
		NewBinder(globals).TryBind(query)
		join := firstBoundNode[*JoinOperator](t, query)
		assertDiagnosticCodes(t, join)
		assertOrderedColumns(t, semanticTable(t, join),
			"Key", "Tenant", "DynamicKey", "LOnly", "OtherKey", "Tenant1", "DynamicKey1", "TextKey", "ROnly")
	})

	t.Run("recursive and and parentheses", func(t *testing.T) {
		query := ParseQuery("Left | join Right on ($left.Key == $right.OtherKey and ($left.Tenant == $right.Tenant and $left.Key == $right.OtherKey)) | count")
		NewBinder(globals).TryBind(query)
		join := firstBoundNode[*JoinOperator](t, query)
		assertDiagnosticCodes(t, join)
		count := firstBoundNode[*CountOperator](t, query)
		assertOrderedColumns(t, semanticTable(t, count), "Count")
	})

	t.Run("lookup substitutes and removes qualified pairs", func(t *testing.T) {
		query := ParseQuery("Left | lookup Right on $left.Key == $right.OtherKey and $left.Tenant == $right.Tenant | project Key, Tenant, LOnly, ROnly")
		NewBinder(globals).TryBind(query)
		lookup := firstBoundNode[*LookupOperator](t, query)
		assertDiagnosticCodes(t, lookup)
		assertOrderedColumns(t, semanticTable(t, lookup), "Key", "Tenant", "DynamicKey", "LOnly", "DynamicKey1", "TextKey", "ROnly")
		if key := semanticTable(t, lookup).GetColumn("Key"); len(key.OriginalColumns) != 2 || key.OriginalColumns[0] != leftKey || key.OriginalColumns[1] != rightOtherKey {
			t.Fatalf("qualified lookup key provenance = %#v", key)
		}
		if tenant := semanticTable(t, lookup).GetColumn("Tenant"); len(tenant.OriginalColumns) != 2 || tenant.OriginalColumns[0] != leftTenant || tenant.OriginalColumns[1] != rightTenant {
			t.Fatalf("qualified lookup tenant provenance = %#v", tenant)
		}
		project := firstBoundNode[*ProjectOperator](t, query)
		assertOrderedColumns(t, semanticTable(t, project), "Key", "Tenant", "LOnly", "ROnly")
	})

	for _, test := range []struct {
		name  string
		query string
		codes []string
	}{
		{name: "invalid operator", query: "Left | join Right on $left.Key != $right.OtherKey", codes: []string{"KS182"}},
		{name: "invalid literal", query: "Left | join Right on true", codes: []string{"KS182"}},
		{name: "invalid left operand", query: "Left | join Right on Key == $right.OtherKey", codes: []string{"KS183"}},
		{name: "invalid right operand", query: "Left | lookup Right on $left.Key == OtherKey", codes: []string{"KS183"}},
		{name: "dynamic key", query: "Left | join Right on $left.DynamicKey == $right.DynamicKey", codes: []string{"KS241"}},
		{name: "inconsistent types", query: "Left | join Right on $left.Key == $right.TextKey", codes: []string{"KS242"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			joinOrLookup := firstJoinOrLookupForParityTest(t, query)
			assertDiagnosticCodes(t, joinOrLookup, test.codes...)
			if !joinOrLookup.IsBound() {
				t.Fatal("diagnosed join/lookup condition was not bound")
			}
		})
	}

	t.Run("parser recovered missing right operand", func(t *testing.T) {
		query := ParseQuery("Left | join Right on $left.Key == | count")
		NewBinder(globals).TryBind(query)
		join := firstBoundNode[*JoinOperator](t, query)
		if !diagnosticsContain(join.ContainedDiagnostics(), "KS006") || !diagnosticsContain(join.SemanticDiagnostics(), "KS183") {
			t.Fatalf("missing RHS diagnostics = %#v", join.ContainedDiagnostics())
		}
		assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")
	})
}

type boundJoinOrLookup interface {
	SyntaxElement
	IsBound() bool
	SemanticDiagnostics() []Diagnostic
}

func firstJoinOrLookupForParityTest(t *testing.T, root SyntaxElement) boundJoinOrLookup {
	t.Helper()
	if joins := boundNodes[*JoinOperator](root); len(joins) != 0 {
		return joins[0]
	}
	if lookups := boundNodes[*LookupOperator](root); len(lookups) != 0 {
		return lookups[0]
	}
	t.Fatal("query has no join or lookup operator")
	return nil
}
